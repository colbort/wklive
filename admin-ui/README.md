# admin-ui (Vue3 + Vite + Element-Plus + Pinia + vue-i18n)

## Features
- admin-api 固定端口：`http://localhost:8888`
- 首页固定展示（不依赖后端菜单）
- 菜单、按钮权限基于 `/admin/auth/profile` 的 `menus + perms`
- 内置多语言：`zh-CN` / `en-US`（后续新增语言仅需增加 locale 文件）

## Run
```bash
npm i
npm run dev
```

## Notes
- 后端菜单 `component` 字段需匹配：`/src/views/${component}.vue`
  - 例如：`system/users` -> `src/views/system/users.vue`
