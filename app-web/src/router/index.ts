import { createRouter, createWebHistory } from 'vue-router'

import AppShell from '@/layout/AppShell.vue'
import AssetsView from '@/views/AssetsView.vue'
import CommoditiesView from '@/views/CommoditiesView.vue'
import CompanyCredentialsView from '@/views/CompanyCredentialsView.vue'
import CryptoContractsView from '@/views/CryptoContractsView.vue'
import ForexView from '@/views/ForexView.vue'
import HomeView from '@/views/HomeView.vue'
import LoginView from '@/views/LoginView.vue'
import MarketsView from '@/views/MarketsView.vue'
import OptionsContractsView from '@/views/OptionsContractsView.vue'
import ProfileView from '@/views/ProfileView.vue'
import RegulatoryFilesView from '@/views/RegulatoryFilesView.vue'
import StocksView from '@/views/StocksView.vue'
import TradesView from '@/views/TradesView.vue'
import WhitepaperView from '@/views/WhitepaperView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.VITE_ROUTER_BASE || '/'),
  routes: [
    {
      path: '/',
      component: AppShell,
      children: [
        { path: '', redirect: '/home' },
        { path: 'home', name: 'home', component: HomeView },
        { path: 'crypto-contracts', name: 'crypto-contracts', component: CryptoContractsView },
        { path: 'stocks', name: 'stocks', component: StocksView },
        { path: 'forex', name: 'forex', component: ForexView },
        { path: 'commodities', name: 'commodities', component: CommoditiesView },
        { path: 'options-contracts', name: 'options-contracts', component: OptionsContractsView },
        {
          path: 'company-credentials',
          name: 'company-credentials',
          component: CompanyCredentialsView,
        },
        { path: 'whitepaper', name: 'whitepaper', component: WhitepaperView },
        { path: 'regulatory-files', name: 'regulatory-files', component: RegulatoryFilesView },
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
