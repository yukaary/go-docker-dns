package main

import (
	yukaarydns "github.com/yukaary/go-docker-dns/dns"
	"strings"
	"testing"
)

func TestReferArecord(t *testing.T) {
	var arecord *yukaarydns.AR
	arecord = yukaarydns.NewAR()
	if arecord == nil {
		t.Errorf("Can't create Arecord.")
	}
}

func TestGlobalArecord(t *testing.T) {
	var arecord *yukaarydns.AR
	arecord = yukaarydns.NewAR()
	t.Logf("compare struct %v", arecord == yukaarydns.Arecord)
	if arecord != yukaarydns.Arecord {
		t.Errorf("arecord != yukaarydns.Arecord.")
	}
}

func TestSplitUrls(t *testing.T) {
	var urlchain = "192.168.1.105:2375,192.168.1.106:2375"
	urls := strings.Split(urlchain, ",")
	for _, url := range urls {
		t.Logf("url:%s", url)
	}
	if len(urls) != 2 {
		t.Errorf("Can't split URL.")
	}
}

func TestSplitUrlsWithSpace(t *testing.T) {
	var urlchain = "192.168.1.105:2375, 192.168.1.106:2375"
	urls := strings.Split(urlchain, ",")
	for _, url := range urls {
		url = strings.Replace(url, " ", "", -1)
		t.Logf("url:%s", url)
	}
	if len(urls) != 2 {
		t.Errorf("Can't split URL.")
	}
}
