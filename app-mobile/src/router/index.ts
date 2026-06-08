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
import LanguageSelectView from '@/views/common/LanguageSelectView.vue'
import HomeView from '@/views/home/HomeView.vue'
import MarketsView from '@/views/markets/MarketsView.vue'
import BindAccountView from '@/views/profile/BindAccountView.vue'
import ChangeSecurityPasswordView from '@/views/profile/ChangeSecurityPasswordView.vue'
import ProfileView from '@/views/profile/ProfileView.vue'
import SecuritySettingsView from '@/views/profile/SecuritySettingsView.vue'
import TradesView from '@/views/trades/TradesView.vue'
import Test1 from '@/views/dev/Test1.vue'
import Test2 from '@/views/dev/Test2.vue'
import Test3 from '@/views/dev/Test3.vue'
import Test4 from '@/views/dev/Test4.vue'

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
      path: '/language',
      name: 'language-select',
      component: LanguageSelectView,
      meta: { title: '语言选择', hideBottomNav: true },
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
          meta: { title: '充值', hideBottomNav: true },
        },
        {
          path: 'assets/withdraw',
          name: 'asset-withdraw',
          component: AssetWithdrawView,
          meta: { title: '提现', hideBottomNav: true },
        },
        {
          path: 'assets/transfer',
          name: 'asset-transfer',
          component: AssetTransferView,
          meta: { title: '划转', hideBottomNav: true },
        },
        {
          path: 'assets/flows',
          name: 'asset-flows',
          component: AssetFundRecordsView,
          meta: { title: '资金记录', hideBottomNav: true },
        },
        {
          path: 'assets/recharge/detail/:orderNo',
          name: 'asset-recharge-detail',
          component: AssetRechargeDetailView,
          meta: { title: '充值详情', hideBottomNav: true },
        },
        { path: 'profile', name: 'profile', component: ProfileView, meta: { title: '我的' } },
        {
          path: 'profile/security',
          name: 'profile-security',
          component: SecuritySettingsView,
          meta: { title: '安全设置', hideBottomNav: true },
        },
        {
          path: 'profile/security/login-password',
          name: 'security-login-password',
          component: ChangeSecurityPasswordView,
          meta: { title: '修改登录密码', hideBottomNav: true },
        },
        {
          path: 'profile/security/pay-password',
          name: 'security-pay-password',
          component: ChangeSecurityPasswordView,
          meta: { title: '修改交易密码', hideBottomNav: true },
        },
        {
          path: 'profile/security/bind-phone',
          name: 'security-bind-phone',
          component: BindAccountView,
          meta: { title: '手机绑定', hideBottomNav: true },
        },
        {
          path: 'profile/security/bind-email',
          name: 'security-bind-email',
          component: BindAccountView,
          meta: { title: '邮箱绑定', hideBottomNav: true },
        },
        {
          path: 'test1',
          name: 'test1',
          component: Test1,
          meta: { title: '测试页面1', hideBottomNav: true },
        },
        {
          path: 'test2',
          name: 'test2',
          component: Test2,
          meta: { title: '测试页面2', hideBottomNav: true },
        },
        {
          path: 'test3',
          name: 'test3',
          component: Test3,
          meta: { title: '测试页面3', hideBottomNav: true },
        },
        {
          path: 'test4',
          name: 'test4',
          component: Test4,
          meta: { title: '测试页面4', hideBottomNav: true },
        },
      ],
    },
  ],
})

router.afterEach((to) => {
  document.title = `AVE - ${String(to.meta.title || '首页')}`
})
