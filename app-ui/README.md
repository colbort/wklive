# app-ui

`app-ui` 是面向 C 端的 Web 项目，采用一个项目同时支持手机端和 PC 端。

## 技术栈

- Vue 3
- Vite
- TypeScript
- Vue Router
- Pinia

## 设计原则

- 一个项目，两端适配
- 共享 API、状态、业务逻辑
- 页面壳按设备区分：移动端底部导航，PC 端侧栏布局
- 业务页面优先复用，差异较大的页面再拆分双实现

## 启动

```bash
npm install
npm run dev
```

默认开发地址：

- `http://localhost:5174`

默认代理到：

- `http://localhost:5555/app`

## 移动端打包

项目使用 Capacitor 作为 Android/iOS 原生壳。首次打包前需要使用 Node 18+。

1. 安装依赖：

```bash
npm install
```

2. 准备移动端环境变量：

```bash
cp .env.capacitor.example .env.capacitor
```

将 `.env.capacitor` 中的 `VITE_API_BASE_URL` 改成真实 HTTPS API 域名，例如：

```env
VITE_APP_TARGET=capacitor
VITE_API_BASE_URL=https://api.example.com
VITE_API_BASE_PATH=/app
```

3. 构建 Web 资源：

```bash
npm run build:mobile
```

4. 首次生成原生工程：

```bash
npm run cap:add:android
npm run cap:add:ios
```

5. 同步资源并打开原生工程：

```bash
npm run cap:sync
npm run cap:open:android
npm run cap:open:ios
```

Android 在 Android Studio 中打 APK/AAB；iOS 在 Xcode 中打包，需要 Apple Developer 账号。
