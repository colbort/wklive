# app-mobile

`app-mobile` 是面向移动端的 Web/App 项目，设计稿基准宽度为 `414px`。

PC 端不在本项目中实现。PC 项目在：

```text
../app-web
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
- Capacitor
- 本地共享包 `@wklive/api`

## 项目边界

- `app-mobile`：只做移动端页面和移动端交互。
- `app-web`：只做 PC 端页面和 PC 端交互。
- `app-packages`：共享 API、类型、通用工具。
- 移动端不要再写 desktop 判断，也不要放 PC 页面。
- PC 端不要复用移动端页面壳和移动端样式。

## API 使用规则

接口来自本地共享包：

```ts
import { apiGetProfile } from '@/api/userPrivate'
```

`app-mobile/src/api/*` 只是转发 `@wklive/api`。

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
http://localhost:5174
```

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

Capacitor 打包环境：

```text
.env.capacitor
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

Web 构建：

```bash
npm run build
```

移动端构建：

```bash
npm run build:mobile
```

格式化：

```bash
make fmt
```

## 移动端打包

项目使用 Capacitor 作为 Android/iOS 原生壳。

首次打包前建议使用 Node 20。

1. 准备移动端环境变量：

```bash
cp .env.capacitor.example .env.capacitor
```

将 `.env.capacitor` 中的 `VITE_API_BASE_URL` 改成真实 HTTPS API 域名：

```env
VITE_APP_TARGET=capacitor
VITE_API_BASE_URL=https://api.example.com
VITE_API_BASE_PATH=/app
VITE_ROUTER_BASE=/
VITE_TENANT_CODE=6TCYR
```

2. 构建 Web 资源：

```bash
npm run build:mobile
```

3. 首次生成原生工程：

```bash
npm run cap:add:android
npm run cap:add:ios
```

4. 同步资源并打开原生工程：

```bash
npm run cap:sync
npm run cap:open:android
npm run cap:open:ios
```

Android 在 Android Studio 中打 APK/AAB。

iOS 在 Xcode 中打包，需要 Apple Developer 账号。

## 相关项目

启动 PC 端：

```bash
cd ../app-web
npm install
npm run dev
```

更新共享 API 后，通常需要在使用方项目重新跑：

```bash
npm run type-check
```
