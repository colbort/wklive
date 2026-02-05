import { createRouter, createWebHistory } from 'vue-router'
import { staticRoutes } from './staticRoutes'
import { useAuthStore } from '@/stores/auth'
import { buildRoutesFromMenus } from './dynamic'

export const router = createRouter({
  history: createWebHistory(),
  routes: staticRoutes,
})

let dynamicAdded = false

router.beforeEach(async (to) => {
  const auth = useAuthStore()

  if ((to.meta as any)?.public) return true

  if (!auth.token) return { path: '/login', query: { redirect: to.fullPath } }

  if (!auth.isProfileLoaded) {
    await auth.fetchProfile()
  }

  if (!dynamicAdded) {
    const dyn = buildRoutesFromMenus(auth.menus)
    dyn.forEach((r) => router.addRoute('Layout', r))
    router.addRoute({
      path: '/:pathMatch(.*)*',
      name: 'NotFound',
      component: () => import('@/views/error/404.vue'),
      meta: { public: true, titleKey: 'route.notFound' },
    })
    dynamicAdded = true
    return { ...to, replace: true }
  }

  return true
})
