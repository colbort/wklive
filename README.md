
## 项目目录说明

```
/wklive
├─ admin-api      # 管理后台 API 网关服务（go-zero），路由、鉴权、业务接口聚合
│  ├─ api         # API 接口定义文件
│  ├─ avatars     # 头像上传相关
│  ├─ curl        # API 测试脚本
│  ├─ etc         # 配置文件
│  ├─ internal    # 内部实现
│  │  ├─ config   # 配置管理
│  │  ├─ handler  # HTTP 处理器
│  │  ├─ logic    # 业务逻辑
│  │  ├─ middleware # 中间件
│  │  └─ svc      # 服务上下文
│  ├─ admin.go    # 主入口文件
│  ├─ go.mod      # Go 模块文件
│  └─ Makefile    # 构建脚本
├─ admin-ui       # 管理后台前端（Vue 3 + Element Plus + TypeScript）
│  ├─ src
│  │  ├─ api      # API 调用封装
│  │  ├─ components # 公共组件
│  │  ├─ composables # Vue 组合式函数
│  │  ├─ config   # 配置文件
│  │  ├─ directives # Vue 指令
│  │  ├─ i18n     # 国际化
│  │  ├─ layout   # 页面布局组件
│  │  ├─ router   # 路由配置
│  │  ├─ services # 服务层
│  │  ├─ stores   # Pinia 状态管理
│  │  ├─ utils    # 工具函数
│  │  ├─ views    # 页面组件
│  │  ├─ App.vue  # 根组件
│  │  └─ main.ts  # 应用入口
│  ├─ public       # 静态资源
│  ├─ dist         # 构建输出
│  ├─ package.json # 项目配置
│  ├─ vite.config.ts # Vite 配置
│  ├─ tsconfig.json # TypeScript 配置
│  └─ index.html   # HTML 模板
├─ services       # 业务微服务模块
│  ├─ system      # 系统管理微服务
│  └─ user        # 用户管理微服务
├─ common         # 公共工具库（配置加载、存储、Nacos/Etcd、JWT 认证等）
│  ├─ etcd        # Etcd 客户端封装
│  ├─ nacos       # Nacos 客户端封装
│  ├─ storage     # 存储服务（OSS、MinIO、COS）
│  ├─ utils       # 通用工具函数
│  └─ go.mod      # Go 模块文件
├─ proto          # gRPC/Protobuf 定义
│  ├─ system      # 系统服务协议定义
│  └─ user        # 用户服务协议定义
├─ app-api        # 应用 API 接口层（预留扩展）
│  ├─ api         # API 接口定义
│  ├─ etc         # 配置文件
│  ├─ internal    # 内部实现
│  ├─ app.go      # 主入口文件
│  └─ go.mod      # Go 模块文件
├─ .github        # GitHub Actions 工作流
├─ wklive.code-workspace # VS Code 工作区配置
└─ README.md      # 项目说明文档
```

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


UserRpc:
  rpcType: zrpc
  Etcd:
    Hosts: 
    - "192.168.10.116:2379"
    Key: user.rpc

PaymentRpc:
  rpcType: zrpc
  Etcd:
    Hosts: 
    - "192.168.10.116:2379"
    Key: payment.rpc

ItickRpc:
  rpcType: zrpc
  Etcd:
    Hosts: 
    - "192.168.10.116:2379"
    Key: itick.rpc
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
ItickRpc:
  rpcType: zrpc
  Etcd:
    Hosts: 
    - "192.168.10.116:2379"
    Key: itick.rpc
```


### user-rpc 微服务配置 /wklive/user-rpc/config
```
Name: user.rpc
ListenOn: 0.0.0.0:8081
Mode: dev
Etcd:
  Hosts:
  - 192.168.10.116:2379
  Key: user.rpc
```


### itick-rpc 微服务配置 /wklive/itick-rpc/config
```
Name: itick.rpc
ListenOn: 0.0.0.0:8082
Mode: dev
Etcd:
  Hosts:
  - 192.168.10.116:2379
  Key: itick.rpc
Itick:
  ApiUrl: https://api.itick.org
  WSUrl: wss://api.itick.org
  Token: 5093272afb5241dfa3fd5505937289804447d9d6941547b2ab45929024c0fd4b
```


### payment-rpc 微服务配置 /wklive/payment-rpc/config
```
Name: payment.rpc
ListenOn: 0.0.0.0:8083
Mode: dev
Etcd:
  Hosts:
  - 192.168.10.116:2379
  Key: payment.rpc
```