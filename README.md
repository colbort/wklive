
## ETCD 配置如下

### 公共配置  /wklive/common/config
```
Log:
  Mode: console
  Encoding: plain
  Stat: false

Mysql:
  DataSource: root:123456@tcp(192.168.10.116:3306)/wklive?charset=utf8mb4&parseTime=true&loc=Local

CacheRedis:
  - Host: 192.168.10.116:6379
    Type: node
    Pass:

Jwt:
  AccessSecret: "your_access_secret"
  AccessExpire: 3600
```


### admin-api 网关配置 /wklive/admin-api/config
```
Name: admin
Host: 0.0.0.0
Port: 8888

SystemRpc:
  rpcType: zrpc
  Etcd:
    Hosts: 
    - "192.168.10.116:2379"
    Key: system.rpc
```


### system-rpc 微服务配置 /wklive/system-rpc/config
```
Name: system.rpc
ListenOn: 0.0.0.0:8080
Mode: dev
Etcd:
  Hosts:
  - 192.168.10.116:2379
  Key: system.rpc
```