# logstash-conf

- 自动生成和配置logstash.conf

***

## 使用

- 安装依赖

```bash
go mod tidy
```

- 编译

```bash
go build -o logstash-conf
```

- help

```bash
logstash-conf -h

Usage:
  logstash-conf [flags]

Examples:
logstash-conf [--confDir] [--topic] [--group]

Flags:
  -c, --conf string    *specify logstash.conf file path
  -g, --group string   specify kafka group id (default "donet")
  -h, --help           help for logstash-conf
  -t, --topic string   *kafka topic name, example: test.abc
```

- 配置

```bash
logstash-conf --conf=logstash.conf --topic=test.abc
# test.haha add done
```

## logstash.conf

```
input {
   kafka {
    codec => "json"
    topics => ["test.xxx"]
    bootstrap_servers => "10.32.12.12:9092"
    auto_offset_reset => "latest"
    group_id => "donet"
    decorate_events => true
    type => "testxxx"
  }kafka {
    codec => "json"
    topics => ["test.service"]
    bootstrap_servers => "10.32.12.12:9092"
    auto_offset_reset => "latest"
    group_id => "donet"
    decorate_events => true
    type => "testservice"
  }
}
output {
  if[type] == "testwcx"{
   elasticsearch {
     hosts => [ "10.32.12.11:9200" ]
     index => "testwcx-%{+YYYY.MM.dd}"
  }
 }if[type] == "testcrmr"{
   elasticsearch {
     hosts => [ "10.32.12.11:9200" ]
     index => "testcrmr-%{+YYYY.MM.dd}"
  }
 }
}
```
