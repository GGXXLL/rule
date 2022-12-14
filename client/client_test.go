package client

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/GGXXLL/rule/driver"
	"github.com/GGXXLL/rule/dto"
	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestDefaultRuleEngine(t *testing.T) {
	addr := os.Getenv("ETCD_ADDR")
	if addr == "" {
		t.Skipf("set ETCD_ADDR to run TestDefaultRuleEngine")
	}
	cli, err := clientv3.New(clientv3.Config{Endpoints: strings.Split(addr, ",")})
	if err != nil {
		t.Fatal(err)
	}
	_, err = cli.Put(context.Background(), "/rule/test/foo", `
style: advanced
rule:
  - if: name == "a"
    then:
      age: 1`)
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Delete(context.Background(), "/rule/test/foo")

	time.Sleep(time.Second)

	engine, clean, err := DefaultRuleEngine(driver.NewEtcdDriver(cli, driver.WithPrefix("/rule/test")))
	if err != nil {
		t.Fatal(err)
	}
	defer clean()

	r, err := engine.Of("/rule/test/foo").Payload(dto.Payload{"name": "a"})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, r.Int("age"))
}
