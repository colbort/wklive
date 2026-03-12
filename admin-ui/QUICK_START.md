# 快速开始 🚀

## 一、验证优化完成

```bash
cd /home/momo/local/go/src/wklive/admin-ui
bash check-optimization.sh
```

期望看到 ✅ **所有优化已完成！**

## 二、安装依赖

```bash
npm install
```

## 三、启动开发服务器

```bash
npm run dev
```

打开浏览器访问 http://localhost:5173

## 四、质量检查

```bash
# 代码检查 + 自动修复
npm run lint

# 代码格式化
npm run format

# TypeScript 类型检查
npm run type-check
```

## 五、查看文档

| 文档 | 内容 |
|------|------|
| [OPTIMIZATION.md](./OPTIMIZATION.md) | 📋 优化详情和代码示例 |
| [OPTIMIZATION_REPORT.md](./OPTIMIZATION_REPORT.md) | 📊 优化总结报告 |
| [DEVELOPER_GUIDE.md](./DEVELOPER_GUIDE.md) | 📖 完整开发者指南 |

## 六、核心改进一览

### ✨ 环境变量配置
```typescript
import { ENV } from '@/config/environment'
console.log(ENV.API_BASE_URL)  // 自动适配环境
```

### ✨ 增强的请求工具
```typescript
import { get, post } from '@/utils/request-enhanced'

// 自动拦截器、日志、错误处理
const data = await get('/api/users', { retry: 3 })
```

### ✨ 统一错误处理
```typescript
import { errorHandler } from '@/utils/error'

try {
  await fetch()
} catch (error) {
  errorHandler.handle(error)  // 自动显示提示
}
```

### ✨ 服务层架构
```typescript
import { userService } from '@/services/UserService'

// 所有 CRUD 都在服务类中
const users = await userService.list({ page: 1 })
const user = await userService.getUser(1)
```

### ✨ Composables (Vue 3 Hooks)
```typescript
import { usePagination, useLoading, useForm } from '@/composables'

// 拿来即用的业务逻辑
const { pagination, nextPage } = usePagination()
const { loading, withLoading } = useLoading()
const { form, submit } = useForm({ initialData: {} })
```

## 七、下一步

1. **阅读** [DEVELOPER_GUIDE.md](./DEVELOPER_GUIDE.md)
2. **查看示例** `src/services/UserService.ts`
3. **创建第一个页面** 按照文档中的步骤
4. **运行代码检查** `npm run lint`

---

**所有文件**都已创建并检查通过！开始开发吧 💪
