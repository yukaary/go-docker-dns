// A simple DNS server for docker container.
// Based on Miek Gieben's Go DNS library and sample code exdns/reflect/reflect.go
//
package main

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/miekg/dns"
	"github.com/yukaary/go-docker-dns/apiwatch"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime/pprof"
	"strconv"
	"strings"
	"syscall"
)

var (
	printf   *bool
	compress *bool
	tsig     *string
	docker   *string
	arecords *apiwatch.AR
)

const port = "53"

func handleContainerRequest(w dns.ResponseWriter, r *dns.Msg) {
	var (
		v4  bool
		rr  dns.RR
		str string
		a   net.IP
	)
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = *compress

	if ip, ok := w.RemoteAddr().(*net.UDPAddr); ok {
		str = "Port: " + strconv.Itoa(ip.Port) + " (udp)"
		a = ip.IP
		v4 = a.To4() != nil
	}
	if ip, ok := w.RemoteAddr().(*net.TCPAddr); ok {
		str = "Port: " + strconv.Itoa(ip.Port) + " (tcp)"
		a = ip.IP
		v4 = a.To4() != nil
	}

	glog.Infof("question: %v", m.Question[0])
	glog.Infof("str: %s", str)
	glog.Infof("a: %s", a.String())

	// Find IP address from a hostname.
	hostname := m.Question[0].Name
	ipaddr := arecords.Find(hostname)
	glog.Infof("ipaddr: %s", ipaddr.String())

	//  If hostname is not registered or a container was stopped,
	//  Do not reply anything.
	if ipaddr == nil {
		w.WriteMsg(m)
		return
	}

	if v4 {
		rr = new(dns.A)
		rr.(*dns.A).Hdr = dns.RR_Header{Name: hostname, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 0}
		rr.(*dns.A).A = ipaddr.To4()
	} else {
		rr = new(dns.AAAA)
		rr.(*dns.AAAA).Hdr = dns.RR_Header{Name: hostname, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: 0}
		rr.(*dns.AAAA).AAAA = ipaddr
	}

	t := new(dns.TXT)
	t.Hdr = dns.RR_Header{Name: hostname, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0}
	t.Txt = []string{str}

	switch r.Question[0].Qtype {
	case dns.TypeTXT:
		m.Answer = append(m.Answer, t)
		m.Extra = append(m.Extra, rr)
	default:
		fallthrough
	case dns.TypeAAAA, dns.TypeA:
		m.Answer = append(m.Answer, rr)
		m.Extra = append(m.Extra, t)
	case dns.TypeAXFR, dns.TypeIXFR:
		c := make(chan *dns.Envelope)
		tr := new(dns.Transfer)
		defer close(c)
		err := tr.Out(w, r, c)
		if err != nil {
			return
		}
		// Statical zone information
		soa, _ := dns.NewRR(`10.in-addr.arpa. IN SOA docker.net. root.docker.net. 2014111813 21600 7200 604800 3600`)
		c <- &dns.Envelope{RR: []dns.RR{soa, t, rr, soa}}
		w.Hijack()
		// w.Close() // Client closes connection
		return
	}

	glog.Infof("m:%v\n", m.String())
	w.WriteMsg(m)
}

func serve(net, name, secret string) {
	switch name {
	case "":
		server := &dns.Server{Addr: ":" + port, Net: net, TsigSecret: nil}
		err := server.ListenAndServe()
		if err != nil {
			fmt.Printf("Failed to setup the "+net+" server: %s\n", err.Error())
		}
	default:
		server := &dns.Server{Addr: ":" + port, Net: net, TsigSecret: map[string]string{name: secret}}
		err := server.ListenAndServe()
		if err != nil {
			fmt.Printf("Failed to setup the "+net+" server: %s\n", err.Error())
		}
	}
}

func main() {
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	printf = flag.Bool("print", false, "print replies")
	compress = flag.Bool("compress", false, "compress replies")
	tsig = flag.String("tsig", "", "use MD5 hmac tsig: keyname:base64")
	docker = flag.String("url", "192.168.59.103:2375", "docker url")

	arecords = apiwatch.NewAR()

	var name, secret string
	flag.Usage = func() {
		flag.PrintDefaults()
	}
	flag.Parse()
	if *tsig != "" {
		a := strings.SplitN(*tsig, ":", 2)
		name, secret = dns.Fqdn(a[0]), a[1] // fqdn the name, which everybody forgets...
	}
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	//dns.HandleFunc("kogasan.net", handleReflect)
	dns.HandleFunc(".", handleContainerRequest)

	go serve("tcp", name, secret)
	go serve("udp", name, secret)

	// run docker api watcher
	go watch(*docker)

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

forever:
	for {
		select {
		case s := <-sig:
			fmt.Printf("Signal (%d) received, stopping\n", s)
			break forever
		}
	}
}

func watch(url string) {
	eventUrl := "http://" + url + "/events"

	glog.Infof("start watching docker api: %s", eventUrl)

	apiwatch.ReadStream(eventUrl, func(id string, status string) {
		inspectUrl := "http://" + url + "/containers/" + id + "/json"

		switch status {
		case "start":
			glog.Infof("inspect: %s\n", inspectUrl)
			data := apiwatch.GetContent(inspectUrl)
			containerInfo := apiwatch.JsonToMap(data)
			config, _ := containerInfo["Config"].(map[string]interface{})

			networkSettings, _ := containerInfo["NetworkSettings"].(map[string]interface{})
			registerIp(config["Hostname"].(string), networkSettings["IPAddress"].(string))
		case "stop":
			glog.Infof("inspect: %s\n", inspectUrl)
			data := apiwatch.GetContent(inspectUrl)
			containerInfo := apiwatch.JsonToMap(data)
			config, _ := containerInfo["Config"].(map[string]interface{})

			unregisterIp(config["Hostname"].(string))
		default:
		}
	})
}

func registerIp(id string, ipaddr string) {
	glog.Infof("register %s, %s", id, ipaddr)
	arecords.Add(id+".", net.ParseIP(ipaddr))
}

func unregisterIp(id string) {
	glog.Infof("unregister %s", id)
	arecords.Add(id+".", nil)
}
