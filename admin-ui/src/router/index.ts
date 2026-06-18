import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { staticRoutes } from './staticRoutes'
import { useAuthStore } from '@/stores'
import { getSystemCore } from '@/stores/core'
import { buildRoutesFromMenus } from './dynamic'

export const router = createRouter({
  history: createWebHistory(),
  routes: staticRoutes,
})

const dynamicRouteNames = new Set<string>()
let dynamicAdded = false
let mustGoogleF2a: number | null = null

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

async function isGoogle2faRequired() {
  if (mustGoogleF2a === null) {
    try {
      const res = await getSystemCore()
      mustGoogleF2a = Number(res.data?.mustGoogleF2a || 0)
    } catch {
      mustGoogleF2a = 0
    }
  }

  return mustGoogleF2a === 1
}

function isGoogle2faEnabled(value: unknown) {
  return value === 1 || value === '1' || value === 'ENABLE_ENABLED'
}

router.beforeEach(async (to) => {
  const auth = useAuthStore()

  if ((to.meta as any)?.public) return true

  if (!auth.token) return { path: '/login', query: { redirect: to.fullPath } }

  if (!auth.isProfileLoaded) {
    await auth.fetchProfile()
  }

  const shouldBindGoogle2fa =
    (await isGoogle2faRequired()) && !isGoogle2faEnabled(auth.user?.google2FaEnabled)
  const isGoogle2faBindPage = to.name === 'Google2faBind'

  if (shouldBindGoogle2fa && !isGoogle2faBindPage) {
    return {
      path: '/google2fa-bind',
      query: { redirect: to.fullPath },
      replace: true,
    }
  }

  if (!shouldBindGoogle2fa && isGoogle2faBindPage) {
    return { path: '/home', replace: true }
  }

  if (isGoogle2faBindPage) return true

  if (!dynamicAdded) {
    addDynamicRoutes(buildRoutesFromMenus(auth.menus))
    return { ...to, replace: true }
  }

  return true
})
