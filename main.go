// A simple DNS server for docker container.
// Based on Miek Gieben's Go DNS library and sample code exdns/reflect/reflect.go
//
package main

import (
	"flag"
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"github.com/golang/glog"
	"github.com/miekg/dns"
	"github.com/yukaary/go-docker-dns/apiwatch"
	yukaarydns "github.com/yukaary/go-docker-dns/dns"
	etcdutil "github.com/yukaary/go-docker-dns/etcd"
	yukaaryutils "github.com/yukaary/go-docker-dns/utils"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime/pprof"
	"strings"
	"syscall"
)

var (
	tsig                  *string
	machine_discover_etcd *string
	discover_key          *string
)

func main() {
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	tsig = flag.String("tsig", "", "use MD5 hmac tsig: keyname:base64")
	machine_discover_etcd = flag.String("machine-discover-etcd", "http://127.0.0.1:4001", "etcd url for machine discovery")
	discover_key = flag.String("discover-key", "machines", "machine discover key")

	// initialize A record
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
	docker_endpoints := etcdutil.GetClusterEndpoints(*machine_discover_etcd, *discover_key, 2375)
	// NEXT, this should support machine scaling
	etcd_endpoints := etcdutil.GetClusterEndpoints(*machine_discover_etcd, *discover_key, 4001)
	client := etcd.NewClient(etcd_endpoints)

	etcdcli := etcdutil.NewEtcdClient(client)
	etcdcli.AddDirIfNotExist("services")

	for _, url := range docker_endpoints {
		glog.Infof("url %s", url)
		go watch(url, etcdcli)
	}

	//go watch(*docker)

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
func watch(url string, etcdcli *etcdutil.EtcdClient) {

	// test given host provides a remote api.
	testUrl := url + "/images/json"
	if _, ret := apiwatch.GetContent(testUrl); ret == false {
		glog.Errorf("cloud not access test endpoint %s. It might not provide a docker remote api.", testUrl)
		os.Exit(1)
	}

	// watch http streaming on /events.
	eventUrl := url + "/events"
	glog.Infof("start watching docker api: %s", eventUrl)

	apiwatch.ReadStream(eventUrl, func(id string, status string) {
		inspectUrl := url + "/containers/" + id + "/json"

		switch status {
		case "start":
			glog.Infof("inspect: %s\n", inspectUrl)
			data, _ := apiwatch.GetContent(inspectUrl)
			containerInfo := apiwatch.JsonToMap(data)
			config, _ := containerInfo["Config"].(map[string]interface{})

			networkSettings, _ := containerInfo["NetworkSettings"].(map[string]interface{})
			registerIp(config["Hostname"].(string), networkSettings["IPAddress"].(string), etcdcli)
		case "stop":
			glog.Infof("inspect: %s\n", inspectUrl)
			data, _ := apiwatch.GetContent(inspectUrl)
			containerInfo := apiwatch.JsonToMap(data)
			config, _ := containerInfo["Config"].(map[string]interface{})

			unregisterIp(config["Hostname"].(string), etcdcli)
		default:
		}
	})
}

func registerIp(id string, ipaddr string, etcdcli *etcdutil.EtcdClient) {
	glog.Infof("register %s, %s", id, ipaddr)
	yukaarydns.Arecord.Add(id+".", net.ParseIP(ipaddr))

	service, _ := yukaaryutils.SplitScaledHostname(id)
	etcdcli.AddDirIfNotExist("services", service)
	etcdcli.Set(ipaddr, "services", service, id)
}

func unregisterIp(id string, etcdcli *etcdutil.EtcdClient) {
	glog.Infof("unregister %s", id)
	yukaarydns.Arecord.Add(id+".", nil)

	service, _ := yukaaryutils.SplitScaledHostname(id)
	etcdcli.DeleteAll("services", service, id)
}
