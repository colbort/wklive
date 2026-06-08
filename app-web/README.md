# app-web

`app-web` 是面向 PC 端的 Web 项目。

移动端不在本项目中实现。移动端项目在：

```text
../app-mobile
```

通用 API、类型和工具放在：

```text
../app-packages
```

## 技术栈

- Vue 3
- Vite
- TypeScript
- Vue Router
- Pinia
- 本地共享包 `@wklive/api`

## 项目边界

- `app-web`：只做 PC 端页面和 PC 端交互。
- `app-mobile`：只做移动端页面和移动端交互。
- `app-packages`：共享 API、类型、通用工具。
- PC 端不要复用移动端页面壳、底部 Tabbar、移动端 `rem` 适配样式。
- 移动端不要在 `app-web` 中写兼容逻辑。

## API 使用规则

接口来自本地共享包：

```ts
import { apiGetProfile } from '@/api'
```

也可以按共享包路径直接引入：

```ts
import { apiGetProfile } from '@wklive/api/api/userPrivate'
```

共享包中按 `app-api/api/*.api` 的后端定义区分接口权限：

- 后端 `.api` 中没有 `jwt: Jwt`：使用 `http`
- 后端 `.api` 中有 `jwt: Jwt`：使用 `authHttp`

未登录时调用 `authHttp` 接口会在前端直接拦截，避免无意义的 401 请求。

## 启动

安装依赖：

```bash
npm install
```

启动开发服务：

```bash
npm run dev
```

默认开发地址：

```text
http://localhost:5175
```

如果端口被占用，Vite 会自动使用下一个可用端口。

默认代理到：

```text
http://localhost:5555/app
```

## 环境变量

开发环境：

```text
.env.development
```

生产环境：

```text
.env.production
```

常用配置：

```env
VITE_API_BASE_URL=http://127.0.0.1:5555
VITE_API_BASE_PATH=/app
VITE_ROUTER_BASE=/
VITE_TENANT_CODE=6TCYR
```

## 常用命令

类型检查：

```bash
npm run type-check
```

构建：

```bash
npm run build
```

预览构建产物：

```bash
npm run preview
```

格式化：

```bash
make fmt
```

## 目录说明

```text
src/
  api/        本项目 API 转发入口
  layout/     PC 页面壳
  router/     路由配置
  styles/     PC 全局样式
  views/      PC 页面
```

## 相关项目

启动移动端：

```bash
cd ../app-mobile
npm install
npm run dev
```

更新共享 API 后，通常需要在使用方项目重新跑：

```bash
npm run type-check
```
