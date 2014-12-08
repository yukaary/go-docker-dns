package dns

import (
	"net"
)

type AR struct {
	records map[string]net.IP
}

/*
 * global host-IP records.
 */
var Arecord *AR

func NewAR() *AR {
	Arecord = &AR{}
	Arecord.initialize()
	return Arecord
}

func (self *AR) initialize() {
	self.records = make(map[string]net.IP)
}

func (self *AR) Add(name string, ip net.IP) {
	self.records[name] = ip
}

func (self *AR) Find(name string) (ip net.IP) {
	return self.records[name]
}

func (self *AR) Records() map[string]net.IP {
	return self.records
}
