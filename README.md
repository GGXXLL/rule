## 简介

基于 `yaml` 语法与条件表达式，快捷配置参数，减少重复、配置表的设计和操作。

## 格式

### basic

```yaml
style: basic
rule:
  name: foo
  age: 18
  gender: 1
```

### advanced

基于多个 `if` 分支，返回不同的数据

```yaml
style: advanced
rule:
  - if: age >= 10
    then:
      name: foo
  - if: true
    then:
      name: baz
```

这个语法下可以进行嵌套：

```yaml
style: advanced
rule:
  - if: age >= 10
    child:
      style: advanced
      rule:
        - if: age < 20
          then:
            name: foo
        - if: age < 30
          then:
            name: bar
  - if: true
    then:
      foo: baz
```

### switch

基于某个字段进行等值判断时，可以写为：

```yaml
style: switch
by: name
rule:
  - case: foo
    style: basic
    rule:
      i: 1
  - case: bar
    style: basic
    rule:
      i: 2
default:
  style: basic
  rule:
    i: 3
```

### 函数

基于 `dto.Payload` 默认提供了以下函数方法：

- `Now() time.Time`
- `Date(s string) time.Time`
- `DaysAgo(s string) int`
- `HoursAgo(s string) int`
- `MinutesAgo(s string) int`
- `DateTime(s string) time.Time`
- `IsBefore(s string) bool`
- `IsAfter(s string) bool`
- `IsBetween(begin string, end string) bool`
- `IsWeekday(day int) bool`
- `IsWeekend() bool`
- `IsToday(s string) bool`
- `IsHourRange(begin int, end int) bool`
- `ToString(str interface{}) string`
- `ToInt(int interface{}) int`

```yaml
style: advanced
rule:
  - if: IsToday(date)
    then:
      name: foo
  - if: true
    then:
      name: baz
```

## 客户端

以 `etcd` 作为存储工具, 并准备路径为 `/example/foo` 的规则配置：

```yaml
style: advanced
rule:
  - if: IsToday(date)
    then:
      name: foo
  - if: true
    then:
      name: baz
```

读取配置：

```go
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

type config struct {
	Name string
}

func main() {
	logger := log.NewJSONLogger(os.Stdout)
	// 创建 etcd 客户端
	cli, err := clientv3.New(clientv3.Config{Endpoints: []string{"127.0.0.1:2379"}})
	if err != nil {
		_ = level.Error(logger).Log("msg", "connect etcd", "err", err)
		return
	}
	// 初始化 driver， 用户可以替换实现
	etcdDrv := driver.NewEtcdDriver(cli, driver.WithPrefix("/example"))

	engine, clean, err := client.DefaultRuleEngine(etcdDrv, logger)
	if err != nil {
		_ = level.Error(logger).Log("msg", "init rule engine", "err", err)
		return
	}
	defer clean()

	// 指定配置名称，并传入条件参数
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

```