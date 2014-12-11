package etcdutil_test

import (
	"github.com/coreos/go-etcd/etcd"
	etcdcli "github.com/yukaary/go-docker-dns/etcd"
	"testing"
)

/**
 * These tests assume local etcd server(http://127.0.0.1:4001) has launched.
 */

func TestHasNoDir(t *testing.T) {
	client := etcd.NewClient([]string{"http://127.0.0.1:4001"})
	etcdcli := etcdcli.NewEtcdClient(client)

	exist := etcdcli.HasDir("hello")
	if exist == true {
		t.Errorf("Threre must not exist directory hello")
	}
}

func TestHasDir(t *testing.T) {
	client := etcd.NewClient([]string{"http://127.0.0.1:4001"})
	etcdcli := etcdcli.NewEtcdClient(client)

	exist := etcdcli.HasDir("hello")
	if exist == true {
		t.Errorf("Threre must not exist directory hello")
	}
}

func TestAddDir(t *testing.T) {
	client := etcd.NewClient([]string{"http://127.0.0.1:4001"})
	etcdcli := etcdcli.NewEtcdClient(client)
	ret := etcdcli.AddDirIfNotExist("yukaary")
	if !ret {
		t.Errorf("Can't add directory.")
	}
	etcdcli.DeleteAll("yukaary")
}

func TestAddSubDir(t *testing.T) {
	client := etcd.NewClient([]string{"http://127.0.0.1:4001"})
	etcdcli := etcdcli.NewEtcdClient(client)
	ret := etcdcli.AddDirIfNotExist("yukaary", "crafter")
	if !ret {
		t.Errorf("Can't add directory.")
	}
	etcdcli.DeleteAll("yukaary")
}

func TestAddKeyValue(t *testing.T) {
	client := etcd.NewClient([]string{"http://127.0.0.1:4001"})
	etcdcli := etcdcli.NewEtcdClient(client)
	ret := etcdcli.AddDirIfNotExist("yukaary", "crafter")
	if !ret {
		t.Errorf("Can't add directory.")
	}
	etcdcli.Set("world", "yukaary", "crafter", "hello")
	etcdcli.DeleteAll("yukaary")
}
