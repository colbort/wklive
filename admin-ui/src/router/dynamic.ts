import type { RouteRecordRaw } from 'vue-router'
import type { MenuNode } from '@/stores/auth'

const viewModules = import.meta.glob('@/views/**/*.vue')

function resolveView(component?: string) {
  if (!component) return null
  const key = `/src/views/${component}.vue`
  return viewModules[key] ? viewModules[key] : null
}

function sortChildren(nodes: MenuNode[]) {
  return [...(nodes || [])].sort((a, b) => (a.sort ?? 0) - (b.sort ?? 0))
}

// export function buildRoutesFromMenus(menus: MenuNode[]): RouteRecordRaw[] {
//   const build = (nodes: MenuNode[], parentPath = ''): RouteRecordRaw[] => {
//     const res: RouteRecordRaw[] = []

//     for (const n of sortChildren(nodes)) {
//       if (n.status === 0 || n.visible === 0) continue

//       // 👉 拼接完整 path
//       const fullPath = resolvePath(parentPath, n.path)

//       // 👉 目录（menuType=1）
//       if (n.menuType === 1) {
//         const children = build(n.children || [], fullPath)

//         // ⚠️ 没子节点可以直接跳过
//         if (!children.length) continue

//         res.push({
//           path: fullPath || `/_dir_${n.id}`, // 防止空 path
//           name: `Dir_${n.id}`,
//           component: () => import('@/views/error/missing-view.vue'), // 你的 Layout
//           meta: {
//             title: n.name,
//             icon: n.icon,
//             menuId: n.id,
//           },
//           children,
//         })
//       }

//       // 👉 菜单（menuType=2）
//       else if (n.menuType === 2 && n.path) {
//         const comp = resolveView(n.component)

//         res.push({
//           path: fullPath,
//           name: `Menu_${n.id}`,
//           component:
//             (comp as any) ||
//             (() => import('@/views/error/missing-view.vue')),
//           meta: {
//             title: n.name,
//             titleKey: `menu.${n.id}`,
//             icon: n.icon,
//             menuId: n.id,
//           },
//         })
//       }

//       // 👉 按钮（menuType=3）不参与路由
//     }

//     return res
//   }

//   return build(menus)
// }

// function resolvePath(parent: string, path?: string) {
//   if (!path) return parent
//   if (path.startsWith('/')) return path
//   return `${parent}/${path}`.replace(/\/+/g, '/')
// }

export function buildRoutesFromMenus(menus: MenuNode[]): RouteRecordRaw[] {
  const routes: RouteRecordRaw[] = []

  const dfs = (nodes: MenuNode[]) => {
    for (const n of sortChildren(nodes)) {
      if (n.status === 0 || n.visible === 0) continue
      if (n.menuType === 2 && n.path) {
        const comp = resolveView(n.component)
        routes.push({
          path: n.path,
          name: `Menu_${n.id}`,
          component: (comp as any) || (() => import('@/views/error/missing-view.vue')),
          meta: {
            title: n.name,
            titleKey: `menu.${n.id}`, // ✅ 方便后期做 i18n：有 key 用 key，否则回退 name
            icon: n.icon,
            menuId: n.id,
          },
        })
      }
      if (n.children?.length) dfs(n.children)
    }
  }

  dfs(menus || [])
  return routes
}
