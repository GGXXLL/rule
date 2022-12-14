package driver

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var etcdClient *clientv3.Client
var prefix = "rule/test/etcd"

func TestMain(m *testing.M) {
	os.Setenv("ETCD_ADDR", "127.0.0.1:2379")
	if os.Getenv("ETCD_ADDR") == "" {
		os.Exit(0)
	}
	cli, err := clientv3.New(clientv3.Config{Endpoints: strings.Split(os.Getenv("ETCD_ADDR"), ",")})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer cli.Close()
	ctx := context.Background()
	_, err = cli.Put(ctx, prefix+"/a", "foo")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	etcdClient = cli
	m.Run()
}

func TestNewEtcdDriver_One(t *testing.T) {
	d := NewEtcdDriver(etcdClient, WithPrefix(prefix))
	v, err := d.One(context.Background(), prefix+"/a")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(v, []byte("foo")) {
		t.Fatalf("should got foo, but got %s", v)
	}
}

func TestNewEtcdDriver_All(t *testing.T) {
	d := NewEtcdDriver(etcdClient, WithPrefix(prefix))
	kvs, err := d.All(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(kvs) == 0 {
		t.Fatal("kvs length should be 1")
	}
	if !bytes.Equal(kvs[0].Value, []byte("foo")) {
		t.Fatalf("should got foo, but got %s", kvs[0].Value)
	}
}

func TestNewEtcdDriver_Watch(t *testing.T) {
	d := NewEtcdDriver(etcdClient, WithPrefix(prefix))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kvChan := d.Watch(ctx)

	time.Sleep(100 * time.Millisecond)
	kb := prefix + "/b"
	go func() {
		etcdClient.Put(ctx, kb, "foo")
	}()

	kv := <-kvChan
	if kv.Key != kb {
		t.Fatalf("want %s, got %s", kb, kv.Key)
	}
	if !bytes.Equal(kv.Value, []byte("foo")) {
		t.Fatalf("should got foo, but got %s", kv.Value)
	}
	cancel()
}
