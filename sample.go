package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var (
	docker   *string = flag.String("url", "192.168.59.103:2375", "docker url")
	logLevel *string = flag.String("logging", "debug", "which log level: [debug,info,warn,error,fetal]")
	logger   *log.Logger
)

func main() {
	flag.Parse()

	logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
	logger.Print("url:" + *docker)
	logger.Print("loglevel:" + *logLevel)

	imagesUrl := "http://" + *docker + "/images/json"
	readHttpGet(imagesUrl)

	eventUrl := "http://" + *docker + "/events"
	readStream(eventUrl)
}

func readHttpGet(endpoint string) {
	resp, err := http.Get(endpoint)
	if err != nil {
		logger.Print("%s", err)
		os.Exit(1)
	} else {
		defer resp.Body.Close()
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Print("%s", err)
			os.Exit(1)
		}
		logger.Print("%s", string(contents))
	}
}

func readStream(endpoint string) {
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
						logger.Print("Error reading stuff: %v", err)
					}
					line = append(line, b[0])
					if b[0] == '}' {
						break
					}
				}
				s := string(line[:])
				logger.Print(s)
			}
		}()
	} else {
		logger.Print("Error in stream: %v", err)
	}
}
