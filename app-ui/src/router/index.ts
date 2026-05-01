import { createRouter, createWebHistory } from 'vue-router'

import AppShell from '@/layout/AppShell.vue'
import HomeView from '@/views/HomeView.vue'
import MarketsView from '@/views/MarketsView.vue'
import TradesView from '@/views/TradesView.vue'
import AssetsView from '@/views/AssetsView.vue'
import AssetRechargeView from '@/views/AssetRechargeView.vue'
import AssetTransferView from '@/views/AssetTransferView.vue'
import AssetWithdrawView from '@/views/AssetWithdrawView.vue'
import AuthForgotPasswordView from '@/views/AuthForgotPasswordView.vue'
import AuthLoginView from '@/views/AuthLoginView.vue'
import AuthRegisterView from '@/views/AuthRegisterView.vue'
import ProfileView from '@/views/ProfileView.vue'

export const router = createRouter({
  history: createWebHistory(import.meta.env.VITE_ROUTER_BASE || '/'),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: AuthLoginView,
      meta: { title: '登录' },
    },
    {
      path: '/forgot-password',
      name: 'forgot-password',
      component: AuthForgotPasswordView,
      meta: { title: '忘记密码' },
    },
    {
      path: '/register',
      name: 'register',
      component: AuthRegisterView,
      meta: { title: '注册' },
    },
    {
      path: '/',
      component: AppShell,
      children: [
        { path: '', redirect: '/home' },
        { path: 'home', name: 'home', component: HomeView, meta: { title: '首页' } },
        { path: 'markets', name: 'markets', component: MarketsView, meta: { title: '市场' } },
        { path: 'trades', name: 'trades', component: TradesView, meta: { title: '交易' } },
        { path: 'assets', name: 'assets', component: AssetsView, meta: { title: '资产' } },
        {
          path: 'assets/recharge',
          name: 'asset-recharge',
          component: AssetRechargeView,
          meta: { title: '充值', hideTabbar: true },
        },
        {
          path: 'assets/withdraw',
          name: 'asset-withdraw',
          component: AssetWithdrawView,
          meta: { title: '提现', hideTabbar: true },
        },
        {
          path: 'assets/transfer',
          name: 'asset-transfer',
          component: AssetTransferView,
          meta: { title: '划转', hideTabbar: true },
        },
        { path: 'profile', name: 'profile', component: ProfileView, meta: { title: '我的' } },
      ],
    },
  ],
})

router.afterEach((to) => {
  document.title = `AVE - ${String(to.meta.title || '首页')}`
})
