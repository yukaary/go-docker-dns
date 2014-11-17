package main

import (
	"encoding/json"
	"flag"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

var (
	docker *string = flag.String("url", "192.168.59.103:2375", "docker url")
)

type processevent func(string, string)

func main() {
	flag.Parse()

	glog.Info("url:" + *docker)

	// print image list.
	/*
		imagesUrl := "http://" + *docker + "/images/json"
		data := getContent(imagesUrl)
		//whatIsThisJson(data)

		var images map[string]interface{}
		images = mapJson(data)

		glog.Infof("%s", images)
	*/

	// stream docker events.
	eventUrl := "http://" + *docker + "/events"

	readStream(eventUrl, func(id string, status string) {
		inspectUrl := "http://" + *docker + "/containers/" + id + "/json"

		switch status {
		case "start":
			glog.Infof("inspect: %s\n", inspectUrl)
			data := getContent(inspectUrl)
			containerInfo := mapJson(data)

			config, _ := containerInfo["Config"].(map[string]interface{})
			glog.Infof("hostname: %s\n", config["Hostname"])

			networkSettings, _ := containerInfo["NetworkSettings"].(map[string]interface{})
			glog.Infof("IPAddress: %s\n", networkSettings["IPAddress"])
			registerIp(id)
		case "stop":
			unregisterIp(id)
		default:
		}
	})
}

func processEvent(id string, status string) {
	switch status {
	case "start":
		registerIp(id)
	case "stop":
		unregisterIp(id)
	default:
	}
}

func registerIp(id string) {
	glog.Infof("register an IP addr of container: %s\n", id)
}

func unregisterIp(id string) {
	glog.Infof("unregister an IP addr of container: %s\n", id)
}

func readHttpGet(endpoint string) {
	resp, err := http.Get(endpoint)
	if err != nil {
		glog.Error("%s", err)
		os.Exit(1)
	} else {
		defer resp.Body.Close()
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			glog.Error("%s", err)
			os.Exit(1)
		}
		glog.Info("%s", string(contents))
	}
}

func getContent(endpoint string) (data interface{}) {

	resp, err := http.Get(endpoint)
	if err == nil {
		defer resp.Body.Close()
		//var data interface{}
		var data interface{}
		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&data); err != nil {
			glog.Error("Error in decoding: ", err)
			os.Exit(1)
		}
		glog.Info("OK. get content.")
		return data
	} else {
		glog.Error("Error in stream: ", err)
		os.Exit(1)
	}

	return
}

func mapJson(f interface{}) map[string]interface{} {
	m := map[string]interface{}{}

	switch vf := f.(type) {
	case map[string]interface{}:
		//glog.Info("found map.")
		for k, v := range vf {
			switch vv := v.(type) {
			case string, float64:
				m[k] = vv
			default:
				m[k] = mapJson(v)
			}
		}
	case []interface{}:
		//glog.Info("found array.")
		for k, v := range vf {
			switch vv := v.(type) {
			case string, float64:
				m[strconv.Itoa(k)] = vv
			default:
				m[strconv.Itoa(k)] = mapJson(v)
			}
		}
	}

	return m
}

func whatIsThisJson(f interface{}) {
	switch vf := f.(type) {
	case map[string]interface{}:
		glog.Info("This is a map.")
		for k, v := range vf {
			//glog.Infof("(k, v) = (%s, %s)", k, v)
			switch vv := v.(type) {
			case string:
				glog.Infof("(k, v) = (%s, %s)", k, vv)
				//glog.Infof("%v: is string - %q", k, vv)
			case float64:
				glog.Infof("(k, v) = (%s, %s)", k, vv)
				//glog.Infof("%v: is int - %q", k, vv)
			default:
				//glog.Infof("key = %s in %s", k, vv)
				whatIsThisJson(v)
			}
		}
	case []interface{}:
		glog.Info("This is an array.")
		for k, v := range vf {
			switch vv := v.(type) {
			case string:
				glog.Infof("(k, v) = (%s, %s)", k, vv)
				//glog.Infof("%v: is string - %q", k, vv)
			case float64:
				glog.Infof("(k, v) = (%s, %s)", k, vv)
				//glog.Infof("%v: is int - %q", k, vv)
			default:
				//glog.Infof("key = %s in %s", k, vv)
				whatIsThisJson(v)
			}
		}
	default:
		//glog.Info("It looks a nil.", vf)
	}
}

func readStream(endpoint string, fn processevent) {
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
						glog.Error("Error reading stuff: %v", err)
					}
					line = append(line, b[0])
					if b[0] == '}' {
						break
					}
				}
				//s := string(line[:])
				//glog.Info(s)
				var data interface{}
				json.Unmarshal(line, &data)
				event := mapJson(data)
				status, _ := event["status"].(string)
				containerid, _ := event["id"].(string)
				fn(containerid, status)
			}
		}()
	} else {
		glog.Error("Error in stream: %v", err)
	}
}
