import { createRouter, createWebHistory } from 'vue-router'
import { staticRoutes } from './staticRoutes'
import { useAuthStore } from '@/stores/auth'
import { useTenantStore } from '@/stores/tenant'

export const router = createRouter({
  history: createWebHistory(),
  routes: staticRoutes,
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  const tenant = useTenantStore()

  if ((to.meta as any)?.public) return true

  if (!auth.token) return { path: '/login', query: { redirect: to.fullPath } }

  await tenant.ensureLoaded()

  if (!auth.isProfileLoaded) {
    await auth.fetchProfile()
  }

  return true
})
