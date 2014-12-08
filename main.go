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
	yukaarydns "github.com/yukaary/go-docker-dns/dns"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime/pprof"
	"strings"
	"syscall"
)

var (
	tsig   *string
	docker *string
)

func main() {
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	tsig = flag.String("tsig", "", "use MD5 hmac tsig: keyname:base64")
	docker = flag.String("url", "192.168.59.103:2375", "docker url")

	yukaarydns.NewAR()

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

	//dns.HandleFunc("yukaary.net", handleYukaaryRequest)
	dns.HandleFunc(".", yukaarydns.HandleContainerRequest)

	go yukaarydns.Serve("tcp", name, secret)
	go yukaarydns.Serve("udp", name, secret)

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

/**
 * Watch docker api "/events" to monitor container start/stop.
 */
func watch(url string) {

	// test given host provides a remote api.
	testUrl := "http://" + url + "/images/json"
	if _, ret := apiwatch.GetContent(testUrl); ret == false {
		glog.Errorf("cloud not access test endpoint %s. It might not provide a docker remote api.", testUrl)
		os.Exit(1)
	}

	// watch http streaming on /events.
	eventUrl := "http://" + url + "/events"
	glog.Infof("start watching docker api: %s", eventUrl)

	apiwatch.ReadStream(eventUrl, func(id string, status string) {
		inspectUrl := "http://" + url + "/containers/" + id + "/json"

		switch status {
		case "start":
			glog.Infof("inspect: %s\n", inspectUrl)
			data, _ := apiwatch.GetContent(inspectUrl)
			containerInfo := apiwatch.JsonToMap(data)
			config, _ := containerInfo["Config"].(map[string]interface{})

			networkSettings, _ := containerInfo["NetworkSettings"].(map[string]interface{})
			registerIp(config["Hostname"].(string), networkSettings["IPAddress"].(string))
		case "stop":
			glog.Infof("inspect: %s\n", inspectUrl)
			data, _ := apiwatch.GetContent(inspectUrl)
			containerInfo := apiwatch.JsonToMap(data)
			config, _ := containerInfo["Config"].(map[string]interface{})

			unregisterIp(config["Hostname"].(string))
		default:
		}
	})
}

func registerIp(id string, ipaddr string) {
	glog.Infof("register %s, %s", id, ipaddr)
	yukaarydns.Arecord.Add(id+".", net.ParseIP(ipaddr))
}

func unregisterIp(id string) {
	glog.Infof("unregister %s", id)
	yukaarydns.Arecord.Add(id+".", nil)
}
