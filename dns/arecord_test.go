package dns_test

import (
	yukaarydns "github.com/yukaary/go-docker-dns/dns"
	"net"
	"testing"
)

func TestMakeAR(t *testing.T) {
	arecords := yukaarydns.NewAR()
	ip := arecords.Find("some.where")
	t.Logf("ip:%v", ip)
	if ip != nil {
		t.Errorf("IP addr should be <nil>")
	}
}

func TestAddFindAR(t *testing.T) {
	arecords := yukaarydns.NewAR()
	arecords.Add("localhost", net.ParseIP("127.0.0.1"))
	ip := arecords.Find("localhost")
	t.Logf("ip:%v", ip.String())
	if ip.String() != "127.0.0.1" {
		t.Errorf("can't find correct IP addr.")
	}

}
