package etcdutil

import (
	"github.com/coreos/go-etcd/etcd"
	"github.com/golang/glog"
	"net/url"
	"strconv"
	"strings"
)

type EtcdClient struct {
	client *etcd.Client
}

func NewEtcdClient(etcd_endpoints []string) *EtcdClient {
	client := etcd.NewClient(etcd_endpoints)
	etcdcli := &EtcdClient{client: client}
	return etcdcli
}

func (self *EtcdClient) Set(value string, keychain ...string) bool {
	key := strings.Join(keychain, "/")
	res, err := self.client.Set(key, value, 0)
	if err != nil {
		glog.Errorf("Can't set key %s value %s", key, value)
		glog.Errorf("> %s", res.Action)
		return false
	}
	return true
}

func (self *EtcdClient) GetDir(key string) etcd.Nodes {
	// get child dir/value recursively
	res, err := self.client.Get(key, true, true)
	if err == nil && res.Node.Dir == true {
		return res.Node.Nodes
	}
	return nil
}

func (self *EtcdClient) AddDirIfNotExist(keychain ...string) bool {

	// check given dir has already exist.
	key := strings.Join(keychain, "/")
	exist := self.HasDir(key)
	if exist {
		return true
	}

	// add dir
	_, err := self.client.CreateDir(key, 0)
	if err == nil {
		return true
	}
	return false
}

func (self *EtcdClient) DeleteAll(keychain ...string) {
	key := strings.Join(keychain, "/")
	self.client.Delete(key, true)
}

func (self *EtcdClient) HasDir(path string) bool {
	res, err := self.client.Get(path, true, true)
	if err == nil && res.Node.Dir == true {
		return true
	}
	return false
}

func convertPort(url *url.URL, port int) string {
	array := strings.Split(url.Host, ":")
	return url.Scheme + "://" + array[0] + ":" + strconv.Itoa(port)
}

func setKv(machine, key, value string) {
	machines := []string{machine}
	client := etcd.NewClient(machines)
	client.Set(key, value, 0)
}

func getDir(machine, key string) etcd.Nodes {
	machines := []string{machine}
	client := etcd.NewClient(machines)

	res, err := client.Get(key, true, true)
	if err == nil && res.Node.Dir == true {
		return res.Node.Nodes
	}
	return nil
}

func GetClusterEndpoints(machine_discover_etcd, key string, port int) []string {
	// Discover cluster machines.
	nodes := getDir(machine_discover_etcd, key)

	// Paese URLs
	etcd_endpoints := []string{}
	for _, node := range nodes {
		endpoint, _ := url.Parse(node.Value)
		etcd_endpoints = append(etcd_endpoints, convertPort(endpoint, port))
	}
	return etcd_endpoints
}
