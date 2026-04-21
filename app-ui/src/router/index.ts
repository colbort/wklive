import { createRouter, createWebHistory } from 'vue-router'

import AppShell from '@/layout/AppShell.vue'
import HomeView from '@/views/HomeView.vue'
import MarketsView from '@/views/MarketsView.vue'
import TradesView from '@/views/TradesView.vue'
import AssetsView from '@/views/AssetsView.vue'
import ProfileView from '@/views/ProfileView.vue'

export const router = createRouter({
  history: createWebHistory(import.meta.env.VITE_ROUTER_BASE || '/'),
  routes: [
    {
      path: '/',
      component: AppShell,
      children: [
        { path: '', redirect: '/home' },
        { path: 'home', name: 'home', component: HomeView, meta: { title: '首页' } },
        { path: 'markets', name: 'markets', component: MarketsView, meta: { title: '市场' } },
        { path: 'trades', name: 'trades', component: TradesView, meta: { title: '交易' } },
        { path: 'assets', name: 'assets', component: AssetsView, meta: { title: '资产' } },
        { path: 'profile', name: 'profile', component: ProfileView, meta: { title: '我的' } },
      ],
    },
  ],
})

router.afterEach((to) => {
  document.title = `AVE - ${String(to.meta.title || '首页')}`
})
