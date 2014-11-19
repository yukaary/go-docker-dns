package apiwatch

import (
	"encoding/json"
	"github.com/golang/glog"
	"net"
	"net/http"
	"strconv"
)

type AR struct {
	records map[string]net.IP
}

type processevent func(string, string)

func (self *AR) initialize() {
	self.records = make(map[string]net.IP)
}

func NewAR() *AR {
	ar := &AR{}
	ar.initialize()
	return ar
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

func ReadStream(endpoint string, fn processevent) {
	resp, err := http.Get(endpoint)
	if err == nil {
		func() {
			defer resp.Body.Close()
			for {
				line := []byte{}
				for {
					b := []byte{0}
					_, err := resp.Body.Read(b)
					if err != nil {
						glog.Errorf("Error reading stuff: %v", err)
					}
					line = append(line, b[0])
					if b[0] == '}' {
						break
					}
				}
				// Debugging
				//glog.Infof("%s", string(line[:]))

				var data interface{}
				json.Unmarshal(line, &data)
				event := JsonToMap(data)
				status, _ := event["status"].(string)
				containerid, _ := event["id"].(string)
				fn(containerid, status)
			}
		}()
	} else {
		glog.Error("Error in stream: %v", err)
	}
}

func GetContent(endpoint string) (data interface{}, result bool) {

	resp, err := http.Get(endpoint)
	if err == nil {
		defer resp.Body.Close()
		var data interface{}
		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&data); err != nil {
			glog.Error("Error in decoding: ", err)
			return nil, false
		}
		return data, true
	} else {
		glog.Error("Error in stream: ", err)
	}

	return nil, false
}

func JsonToMap(f interface{}) map[string]interface{} {
	m := map[string]interface{}{}

	switch vf := f.(type) {
	case map[string]interface{}:
		for k, v := range vf {
			switch vv := v.(type) {
			case string, float64:
				m[k] = vv
			default:
				m[k] = JsonToMap(v)
			}
		}
	case []interface{}:
		for k, v := range vf {
			switch vv := v.(type) {
			case string, float64:
				m[strconv.Itoa(k)] = vv
			default:
				m[strconv.Itoa(k)] = JsonToMap(v)
			}
		}
	}

	return m
}
