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
