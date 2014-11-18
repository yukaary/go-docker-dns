package apiwatch

import (
	"."
	"net"
	"testing"
)

func TestMakeAR(t *testing.T) {
	arecords := apiwatch.NewAR()
	ip := arecords.Find("some.where")
	t.Logf("ip:%v", ip)
	if ip != nil {
		t.Errorf("IP addr should be <nil>")
	}
}

func TestAddFindAR(t *testing.T) {
	arecords := apiwatch.NewAR()
	arecords.Add("localhost", net.ParseIP("127.0.0.1"))
	ip := arecords.Find("localhost")
	t.Logf("ip:%v", ip.String())
	if ip.String() != "127.0.0.1" {
		t.Errorf("can't find correct IP addr.")
	}

}
