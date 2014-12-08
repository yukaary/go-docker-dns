package main

import (
	yukaarydns "github.com/yukaary/go-docker-dns/dns"
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
