# Admin UI - 开发者指南

## 📖 项目概述

Admin UI 是一个现代化的 Vue 3 + TypeScript + Element Plus 后台管理系统前端框架。
经过优化后，提供了高度组织化的代码结构、丰富的工具库和最佳实践。

## 🚀 快速上手

### 1. 安装依赖
```bash
npm install
```

### 2. 启动开发服务器
```bash
npm run dev
```

打开 http://localhost:5173，应该会自动跳转到登录页面。

### 3. 构建生产包
```bash
npm run build
```

## 📚 核心概念

### API 和服务层

所有与后端通信都应该通过 `services` 目录下的专用服务类：

```typescript
// 不推荐：直接在组件中调用 API
import { get } from '@/utils/request'
const user = await get('/admin/users/1')

// ✅ 推荐：使用服务类
import { userService } from '@/services/UserService'
const user = await userService.getUser(1)
```

**好处**：
- 集中管理 API 逻辑
- 便于维护和测试
- 代码重用性更高
- 类型安全

### 可组合式函数 (Composables)

使用 Composables 来复用逻辑，而不是 mixins：

```typescript
// 列表页面示例
import { usePagination, useLoading } from '@/composables'
import { userService } from '@/services/UserService'

export default {
  setup() {
    const { pagination, updateTotal } = usePagination(20)
    const { loading, withLoading } = useLoading()
    const users = ref([])

    const fetchUsers = () =>
      withLoading(async () => {
        const res = await userService.list(pagination)
        users.value = res.data?.list || []
        updateTotal(res.data?.total || 0)
      })

    onMounted(() => fetchUsers())

    return { users, pagination, loading, fetchUsers }
  }
}
```

### 错误处理

统一使用 `errorHandler` 处理异常：

```typescript
import { errorHandler } from '@/utils/error'

try {
  await userService.deleteUser(id)
} catch (error) {
  // 自动显示 ElMessage.error，并记录日志
  errorHandler.handle(error)
}
```

### 日志记录

使用 `logger` 进行调试和监控：

```typescript
import { logger } from '@/utils/logger'

logger.debug('User object:', user)
logger.info('User created successfully')
logger.warn('User status is inactive')
logger.error('Failed to fetch user', error)
```

## 🎨 代码规范

### 命名规范

| 类型 | 规范 | 示例 |
|------|------|------|
| 组件 | PascalCase | `UserForm.vue`, `DataTable.vue` |
| 文件 | kebab-case | `user-form.vue`, `data-table.vue` |
| 函数/变量 | camelCase | `getUserData()`, `userName` |
| 常量 | UPPER_SNAKE_CASE | `API_BASE_URL`, `MAX_RETRY_COUNT` |
| 类 | PascalCase | `UserService`, `BaseService` |
| 接口 | PascalCase + I 前缀或省略 | `UserQuery`, `User` |

### 文件组织

```
按功能模块组织，而非按类型组织

❌ 不推荐：
views/
├── users/
│   ├── Index.vue
│   ├── Add.vue
│   └── Edit.vue
api/
├── user.ts
components/
└── UserForm.vue

✅ 推荐：
features/users/
├── views/
│   ├── Index.vue
│   ├── Add.vue
│   └── Edit.vue
├── services/
│   └── UserService.ts
├── types/
│   └── User.ts
├── components/
│   └── UserForm.vue
└── stores/
    └── userStore.ts
```

### TypeScript 最佳实践

```typescript
// ✅ 使用具体类型，避免 any
interface User {
  id: number
  username: string
  email: string
  status: 'active' | 'inactive'
}

// ✅ 为函数参数和返回值标注类型
function getUser(id: number): Promise<User> {
  return userService.getUser(id)
}

// ✅ 使用 const assertion 避免过度模糊
const config = {
  timeout: 5000,
  retries: 3,
} as const

// ❌ 避免使用 any
function handleResponse(data: any) { }

// ✅ 即使不确定，也使用 unknown
function handleResponse(data: unknown) {
  if (typeof data === 'object' && data !== null) {
    // ...
  }
}
```

## 🧪 开发工作流

### 创建新页面的步骤

1. **创建路由**

```typescript
// router/index.ts
{
  path: '/system/users',
  component: () => import('@/views/system/UserList.vue'),
  meta: {
    title: '用户管理',
    icon: 'User',
  }
}
```

2. **创建服务**

```typescript
// services/UserService.ts
export class UserService extends BaseService {
  constructor() {
    super('/admin/users')
  }
  // 实现业务方法
}
```

3. **创建类型**

```typescript
// types/User.ts
export interface User {
  id: number
  username: string
  email: string
  // ...
}
```

4. **创建视图组件**

```vue
<!-- views/system/UserList.vue -->
<template>
  <div class="user-list">
    <!-- 列表内容 -->
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { usePagination, useLoading } from '@/composables'
import { userService } from '@/services/UserService'

// 组件逻辑
</script>

<style scoped>
/* 样式 */
</style>
```

### 代码审查要点

- [ ] 是否使用了正确的服务类
- [ ] 类型标注是否完整
- [ ] 是否正确处理了错误
- [ ] 是否遵循命名规范
- [ ] 是否有必要的注释
- [ ] 是否考虑了性能影响
- [ ] 是否需要权限检查

## 🔍 调试技巧

### 查看网络请求

在浏览器开发者工具的 **Network** 标签中查看所有 HTTP 请求。
所有请求都会自动记录到控制台（开发环境）。

### 查看日志

```typescript
// 启用日志记录
import { logger } from '@/utils/logger'

logger.debug('当前用户:', authStore.user)
```

### Vue DevTools

安装 Vue DevTools 浏览器扩展：
- 检查组件状态
- 查看 Pinia store 变化
- 追踪事件发射

### 时间旅行调试

使用 Pinia DevTools 中的操作历史：
```typescript
// 在 DevTools 中回溯状态变化
```

## 📦 依赖管理

### 添加新包

```bash
npm install package-name

# 开发依赖
npm install --save-dev package-name

# 指定版本
npm install package-name@latest
```

### 更新依赖

```bash
# 检查过时的包
npm outdated

# 更新所有依赖
npm update

# 安全审计
npm audit
npm audit fix
```

## 🌍 环境变量

### 使用环境变量

```typescript
import { ENV } from '@/config/environment'

// 自动适配当前环境
const apiUrl = ENV.API_BASE_URL
const isDev = ENV.IS_DEV
```

### 添加新环境变量

1. 在 `.env`, `.env.development`, `.env.production` 中定义

```env
VITE_CUSTOM_VAR=value
```

2. 在 `config/environment.ts` 中映射

```typescript
export const ENV = {
  CUSTOM_VAR: import.meta.env.VITE_CUSTOM_VAR || 'default',
}
```

3. 在代码中使用

```typescript
import { ENV } from '@/config/environment'
console.log(ENV.CUSTOM_VAR)
```

## 🚨 常见问题

### Q: 如何处理异步数据加载
A: 使用 `useAsync` composable

```typescript
const { data, loading, error } = useAsync(
  () => userService.getUser(id),
  { immediate: true }
)
```

### Q: 如何保存用户偏好设置
A: 使用 `useLocalStorage` composable

```typescript
const { data: theme } = useLocalStorage('theme', 'light')
```

### Q: 如何实现权限控制
A: 使用 `v-perm` 指令或在路由中检查

```vue
<!-- 使用指令隐藏元素 -->
<el-button v-perm="'user:delete'">Delete</el-button>

<!-- 检查权限与否 -->
<el-button v-if="authStore.hasPerm('user:delete')">Delete</el-button>
```

### Q: 如何处理表单验证
A: 使用 `useForm` composable 和 Element Plus 的验证

```typescript
const { form, formRef, submit } = useForm({
  initialData: { username: '', email: '' }
})

const handleSubmit = async () => {
  const data = await submit()
  if (data) {
    await userService.createUser(data)
  }
}
```

## 📖 相关资源

- [Vue 3 官方文档](https://vuejs.org)
- [TypeScript 官方文档](https://www.typescriptlang.org)
- [Element Plus 文档](https://element-plus.org)
- [Vite 官方文档](https://vitejs.dev)
- [Pinia 文档](https://pinia.vuejs.org)
- [Vue Router 文档](https://router.vuejs.org)

---

**最后更新**: 2026-03-12
