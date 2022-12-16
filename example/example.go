package main

import (
	"fmt"
	"os"
	"time"

	"github.com/GGXXLL/rule/client"
	"github.com/GGXXLL/rule/driver"
	"github.com/GGXXLL/rule/dto"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	clientv3 "go.etcd.io/etcd/client/v3"
)

/*
样例配置文件
style: advanced
rule:
  - if: IsToday(date)
    then:
      name: foo
  - if: true
    then:
      name: baz

*/

type config struct {
	Name string
}

func main() {
	logger := log.NewJSONLogger(os.Stdout)
	cli, err := clientv3.New(clientv3.Config{Endpoints: []string{"127.0.0.1:2379"}})
	if err != nil {
		_ = level.Error(logger).Log("msg", "connect etcd", "err", err)
		return
	}
	etcdDrv := driver.NewEtcdDriver(cli, driver.WithPrefix("/example"))

	engine, clean, err := client.DefaultRuleEngine(etcdDrv, logger)
	if err != nil {
		_ = level.Error(logger).Log("msg", "init rule engine", "err", err)
		return
	}
	defer clean()

	r, err := engine.Of("/example/foo").Payload(dto.Payload{"date": time.Now().Format("2006-01-02")})
	if err != nil {
		_ = level.Error(logger).Log("msg", "init rule engine", "err", err)
		return
	}
	fmt.Println(r.String("name"))

	var c config
	err = r.Unmarshal("", &c)
	if err != nil {
		_ = level.Error(logger).Log("msg", "unmarshal result to struct", "err", err)
		return
	}
	fmt.Println(c)
}
