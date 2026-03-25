import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { staticRoutes } from './staticRoutes'
import { useAuthStore } from '@/stores'
import { buildRoutesFromMenus } from './dynamic'

export const router = createRouter({
  history: createWebHistory(),
  routes: staticRoutes,
})

const dynamicRouteNames = new Set<string>()
let dynamicAdded = false

function ensureNotFoundRoute() {
  if (router.hasRoute('NotFound')) return

  router.addRoute({
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/error/404.vue'),
    meta: { public: true, titleKey: 'route.notFound' },
  })
}

function addDynamicRoutes(routes: RouteRecordRaw[]) {
  routes.forEach((route) => {
    router.addRoute('Layout', route)
    if (route.name) {
      dynamicRouteNames.add(String(route.name))
    }
  })
  ensureNotFoundRoute()
  dynamicAdded = true
}

export function resetDynamicRoutes() {
  dynamicRouteNames.forEach((name) => {
    if (router.hasRoute(name)) {
      router.removeRoute(name)
    }
  })
  dynamicRouteNames.clear()

  if (router.hasRoute('NotFound')) {
    router.removeRoute('NotFound')
  }

  dynamicAdded = false
}

router.beforeEach(async (to) => {
  const auth = useAuthStore()

  if ((to.meta as any)?.public) return true

  if (!auth.token) return { path: '/login', query: { redirect: to.fullPath } }

  if (!auth.isProfileLoaded) {
    await auth.fetchProfile()
  }

  if (!dynamicAdded) {
    addDynamicRoutes(buildRoutesFromMenus(auth.menus))
    return { ...to, replace: true }
  }

  return true
})
