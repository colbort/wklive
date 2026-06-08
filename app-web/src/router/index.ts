import { createRouter, createWebHistory } from 'vue-router'

import AppShell from '@/layout/AppShell.vue'
import AssetsView from '@/views/AssetsView.vue'
import HomeView from '@/views/HomeView.vue'
import LoginView from '@/views/LoginView.vue'
import MarketsView from '@/views/MarketsView.vue'
import ProfileView from '@/views/ProfileView.vue'
import TradesView from '@/views/TradesView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.VITE_ROUTER_BASE || '/'),
  routes: [
    {
      path: '/',
      component: AppShell,
      children: [
        { path: '', redirect: '/home' },
        { path: 'home', name: 'home', component: HomeView },
        { path: 'markets', name: 'markets', component: MarketsView },
        { path: 'trades', name: 'trades', component: TradesView },
        { path: 'assets', name: 'assets', component: AssetsView },
        { path: 'profile', name: 'profile', component: ProfileView },
      ],
    },
    { path: '/login', name: 'login', component: LoginView },
  ],
})

export default router
