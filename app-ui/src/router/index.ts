import { createRouter, createWebHistory } from 'vue-router'

import AppShell from '@/layout/AppShell.vue'
import AssetFundRecordsView from '@/views/assets/FundRecordsView.vue'
import AssetRechargeDetailView from '@/views/assets/RechargeDetailView.vue'
import AssetRechargeView from '@/views/assets/RechargeView.vue'
import AssetsView from '@/views/assets/AssetsView.vue'
import AssetTransferView from '@/views/assets/TransferView.vue'
import AssetWithdrawView from '@/views/assets/WithdrawView.vue'
import AuthForgotPasswordView from '@/views/auth/ForgotPasswordView.vue'
import AuthLoginView from '@/views/auth/LoginView.vue'
import AuthRegisterView from '@/views/auth/RegisterView.vue'
import HomeView from '@/views/home/HomeView.vue'
import MarketsView from '@/views/markets/MarketsView.vue'
import BindAccountView from '@/views/profile/BindAccountView.vue'
import ChangeSecurityPasswordView from '@/views/profile/ChangeSecurityPasswordView.vue'
import ProfileView from '@/views/profile/ProfileView.vue'
import SecuritySettingsView from '@/views/profile/SecuritySettingsView.vue'
import TradesView from '@/views/trades/TradesView.vue'

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
        {
          path: 'assets/flows',
          name: 'asset-flows',
          component: AssetFundRecordsView,
          meta: { title: '资金记录', hideTabbar: true },
        },
        {
          path: 'assets/recharge/detail/:orderNo',
          name: 'asset-recharge-detail',
          component: AssetRechargeDetailView,
          meta: { title: '充值详情', hideTabbar: true },
        },
        { path: 'profile', name: 'profile', component: ProfileView, meta: { title: '我的' } },
        {
          path: 'profile/security',
          name: 'profile-security',
          component: SecuritySettingsView,
          meta: { title: '安全设置', hideTabbar: true },
        },
        {
          path: 'profile/security/login-password',
          name: 'security-login-password',
          component: ChangeSecurityPasswordView,
          meta: { title: '修改登录密码', hideTabbar: true },
        },
        {
          path: 'profile/security/pay-password',
          name: 'security-pay-password',
          component: ChangeSecurityPasswordView,
          meta: { title: '修改交易密码', hideTabbar: true },
        },
        {
          path: 'profile/security/bind-phone',
          name: 'security-bind-phone',
          component: BindAccountView,
          meta: { title: '手机绑定', hideTabbar: true },
        },
        {
          path: 'profile/security/bind-email',
          name: 'security-bind-email',
          component: BindAccountView,
          meta: { title: '邮箱绑定', hideTabbar: true },
        },
      ],
    },
  ],
})

router.afterEach((to) => {
  document.title = `AVE - ${String(to.meta.title || '首页')}`
})
