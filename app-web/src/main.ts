import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { configureApiClient } from '@wklive/api/api/http'

import App from './App.vue'
import router from './router'
import './styles/global.css'

const env = import.meta.env

configureApiClient({
  apiBaseUrl: env.VITE_API_BASE_URL,
  apiBasePath: env.VITE_API_BASE_PATH,
  appTarget: env.VITE_APP_TARGET,
  tenantCode: env.VITE_TENANT_CODE,
  tenantId: env.VITE_TENANT_ID,
})

createApp(App).use(createPinia()).use(router).mount('#app')
