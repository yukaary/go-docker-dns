package dns

import (
	"github.com/golang/glog"
	"github.com/miekg/dns"
	"net"
	"strconv"
)

const port = "53"

func HandleContainerRequest(w dns.ResponseWriter, r *dns.Msg) {
	var (
		v4  bool
		rr  dns.RR
		str string
		a   net.IP
	)
	m := new(dns.Msg)
	m.SetReply(r)

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
	ipaddr := Arecord.Find(hostname)

	//  If hostname is not registered or a container was stopped,
	//  Do not reply anything.
	if ipaddr == nil {
		glog.Info("no answer.")
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

func Serve(net, name, secret string) {
	switch name {
	case "":
		server := &dns.Server{Addr: ":" + port, Net: net, TsigSecret: nil}
		err := server.ListenAndServe()
		if err != nil {
			glog.Errorf("Failed to setup the "+net+" server: %s\n", err.Error())
		}
	default:
		server := &dns.Server{Addr: ":" + port, Net: net, TsigSecret: map[string]string{name: secret}}
		err := server.ListenAndServe()
		if err != nil {
			glog.Errorf("Failed to setup the "+net+" server: %s\n", err.Error())
		}
	}
}
