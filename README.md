## 项目目录说明

```
/wklive
├─ admin-api          # 管理后台 API 网关（go-zero）
│  ├─ api             # REST API 定义
│  ├─ etc             # 服务配置
│  └─ internal        # handler、logic、middleware、svc 等实现
├─ admin-ui           # 管理后台 Web（Vue 3 + Element Plus + TypeScript）
│  ├─ public          # 静态资源
│  └─ src             # API、组件、路由、状态、页面等
├─ app-api            # 客户端 API 网关（go-zero）
│  ├─ api
│  ├─ etc
│  └─ internal
├─ app-web            # 客户端 Web（Vue 3 + Vite）
│  ├─ public
│  ├─ assets
│  └─ src
├─ app-mobile         # 移动客户端（Ionic/Capacitor + Vue）
│  ├─ android         # Android 工程
│  ├─ ios             # iOS 工程
│  └─ src
├─ app-packages       # app-web/app-mobile 共享的 TypeScript 包
│  └─ src             # API、类型和工具
├─ chat-api           # 客服业务 API 网关
│  ├─ api
│  ├─ etc
│  └─ internal
├─ chat-ui            # 客服聊天客户端
│  └─ src
├─ chat-admin-api     # 客服管理 API 网关
│  ├─ api
│  ├─ etc
│  └─ internal
├─ chat-admin-ui      # 客服管理后台
│  └─ src
├─ services           # gRPC 业务微服务
│  ├─ asset          # 资产服务
│  ├─ chat           # 客服服务
│  ├─ market          # 行情与产品数据服务
│  ├─ option         # 期权服务
│  ├─ payment        # 支付服务
│  ├─ staking        # 质押服务
│  ├─ system         # 系统与租户服务
│  ├─ trade          # 交易服务
│  └─ user           # 用户服务
├─ proto              # gRPC/Protobuf 协议与生成代码
│  ├─ asset
│  ├─ chat
│  ├─ common
│  ├─ market
│  ├─ option
│  ├─ payment
│  ├─ staking
│  ├─ system
│  ├─ trade
│  └─ user
├─ common             # Go 公共库
│  ├─ bus             # Redis 消息总线
│  ├─ etcd / nacos    # 配置与服务发现
│  ├─ i18n            # 国际化与错误响应
│  ├─ middleware      # 通用中间件
│  ├─ storage         # 对象存储封装
│  └─ utils           # 通用工具
├─ .github            # GitHub Actions 与项目辅助文档
├─ init.sql           # 数据库初始化数据
├─ etcdwp.json        # Etcd 配置示例/备份
├─ wklive.code-workspace
└─ README.md
```

## 修复 go 文件中的 import
```
// 安装 goimports
go install golang.org/x/tools/cmd/goimports@latest
find . -name "*.go" -print0 | xargs -0 goimports -w
```

## 格式化 proto
```
// 安装 brew install clang-format
find . -name "*.proto" -exec clang-format -i {} \;
```

## VSCODE 插件
- Protobuf VSC 
- goctl

## ETCD 配置如下

### 公共配置 /wklive/common/config

```
Log:
  Mode: console
  Encoding: plain
  Stat: false

Mysql:
  DataSource: root:123456@tcp(127.0.0.1:3306)/wklive?charset=utf8mb4&parseTime=true&loc=Local

CacheRedis:
  - Host: 127.0.0.1:6379
    Type: node
    Pass:

Redis:
  Key: test123
  Host: 127.0.0.1:6379
  Type: node
  Pass:

BusRedis:
  - Host: 127.0.0.1:6379
    Type: node
    Pass:

LockRedis:
  - Host: 127.0.0.1:6379
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
    - "127.0.0.1:2379"
    Key: system.rpc


UserRpc:
  rpcType: zrpc
  Etcd:
    Hosts:
    - "127.0.0.1:2379"
    Key: user.rpc

PaymentRpc:
  rpcType: zrpc
  Etcd:
    Hosts:
    - "127.0.0.1:2379"
    Key: payment.rpc

MarketRpc:
  rpcType: zrpc
  Etcd:
    Hosts:
    - "127.0.0.1:2379"
    Key: market.rpc

AssetRpc:
  rpcType: zrpc
  Etcd:
    Hosts:
    - "127.0.0.1:2379"
    Key: asset.rpc   

OptionRpc:
  rpcType: zrpc
  Etcd:
    Hosts:
    - "127.0.0.1:2379"
    Key: option.rpc   

StakingRpc:
  rpcType: zrpc
  Etcd:
    Hosts:
    - "127.0.0.1:2379"
    Key: staking.rpc   

TradeRpc:
  rpcType: zrpc
  Etcd:
    Hosts:
    - "127.0.0.1:2379"
    Key: trade.rpc   
```

### app-api 网关配置 /wklive/app-api/config

```
Name: app
Host: 0.0.0.0
Port: 6666

SystemRpc:
  rpcType: zrpc
  Etcd:
    Hosts:
    - "127.0.0.1:2379"
    Key: system.rpc

UserRpc:
  rpcType: zrpc
  Etcd:
    Hosts:
    - "127.0.0.1:2379"
    Key: user.rpc

PaymentRpc:
  rpcType: zrpc
  Etcd:
    Hosts:
    - "127.0.0.1:2379"
    Key: payment.rpc

MarketRpc:
  rpcType: zrpc
  Etcd:
    Hosts:
    - "127.0.0.1:2379"
    Key: market.rpc

AssetRpc:
  rpcType: zrpc
  Etcd:
    Hosts:
    - "127.0.0.1:2379"
    Key: asset.rpc   

OptionRpc:
  rpcType: zrpc
  Etcd:
    Hosts:
    - "127.0.0.1:2379"
    Key: option.rpc   

StakingRpc:
  rpcType: zrpc
  Etcd:
    Hosts:
    - "127.0.0.1:2379"
    Key: staking.rpc   

TradeRpc:
  rpcType: zrpc
  Etcd:
    Hosts:
    - "127.0.0.1:2379"
    Key: trade.rpc   
```

### system-rpc 微服务配置 /wklive/system-rpc/config

```
Name: system.rpc
ListenOn: 0.0.0.0:8080
Mode: dev
Etcd:
  Hosts:
  - 127.0.0.1:2379
  Key: system.rpc
MarketRpc:
  rpcType: zrpc
  Etcd:
    Hosts:
    - "127.0.0.1:2379"
    Key: market.rpc
```

### user-rpc 微服务配置 /wklive/user-rpc/config

```
Name: user.rpc
ListenOn: 0.0.0.0:8081
Mode: dev
Etcd:
  Hosts:
  - 127.0.0.1:2379
  Key: user.rpc

SystemRpc:
  rpcType: zrpc
  Etcd:
    Hosts:
    - "127.0.0.1:2379"
    Key: system.rpc   
```

### market-rpc 微服务配置 /wklive/market-rpc/config

```
Name: market.rpc
ListenOn: 0.0.0.0:8082
Mode: dev
Etcd:
  Hosts:
  - 127.0.0.1:2379
  Key: market.rpc

SystemRpc:
  rpcType: zrpc
  Etcd:
    Hosts:
    - "127.0.0.1:2379"
    Key: system.rpc

Itick:
  ApiUrl: https://api.itick.org
  WSUrl: wss://api.itick.org
  Token: 5093272afb5241dfa3fd5505937289804447d9d6941547b2ab45929024c0fd4b

Mongo:
  Url: "mongodb://root:openIM123@127.0.0.1:27017/stock?authSource=admin"
  Db: "wklive"

KlineWriter:
  QueueSize: 20000
  BatchSize: 300
  FlushIntervalMs: 1000
  WriteTimeoutMs: 5000
```

### payment-rpc 微服务配置 /wklive/payment-rpc/config
```
Name: payment.rpc
ListenOn: 0.0.0.0:8083
Mode: dev
Etcd:
  Hosts:
  - 127.0.0.1:2379
  Key: payment.rpc
```

### asset-rpc 微服务配置 /wklive/asset-rpc/config
```
Name: asset.rpc
ListenOn: 0.0.0.0:8084
Mode: dev
Etcd:
  Hosts:
  - 127.0.0.1:2379
  Key: asset.rpc
```

### option-rpc 微服务配置 /wklive/option-rpc/config
```
Name: option.rpc
ListenOn: 0.0.0.0:8085
Mode: dev
Etcd:
  Hosts:
  - 127.0.0.1:2379
  Key: option.rpc

AssetRpc:
  rpcType: zrpc
  Etcd:
    Hosts:
    - "127.0.0.1:2379"
    Key: asset.rpc   
```

### staking-rpc 微服务配置 /wklive/staking-rpc/config
```
Name: staking.rpc
ListenOn: 0.0.0.0:8086
Mode: dev
Etcd:
  Hosts:
  - 127.0.0.1:2379
  Key: staking.rpc

AssetRpc:
  rpcType: zrpc
  Etcd:
    Hosts:
    - "127.0.0.1:2379"
    Key: asset.rpc   
```

### trade-rpc 微服务配置 /wklive/trade-rpc/config
```
Name: trade.rpc
ListenOn: 0.0.0.0:8087
Mode: dev
Etcd:
  Hosts:
  - 127.0.0.1:2379
  Key: trade.rpc

AssetRpc:
  rpcType: zrpc
  Etcd:
    Hosts:
    - "127.0.0.1:2379"
    Key: asset.rpc   
```
