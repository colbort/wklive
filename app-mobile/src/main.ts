import { createApp } from 'vue'
import { createPinia } from 'pinia'

import { configureApiClient } from '@/api/http'
import App from '@/App.vue'
import { createI18n, translateApiError } from '@/i18n'
import { router } from '@/router'
import { useSystemStore } from '@/stores/system'
import { useTenantStore } from '@/stores/tenant'
import '@/styles/global.css'

const app = createApp(App)
const pinia = createPinia()

configureApiClient({
  apiBaseUrl:
    import.meta.env.VITE_APP_TARGET === 'capacitor'
      ? import.meta.env.VITE_API_BASE_URL
      : '',
  translateApiError,
})

app.use(pinia)
app.use(createI18n())
app.use(router)

const systemStore = useSystemStore(pinia)
const tenantStore = useTenantStore(pinia)

tenantStore.hydrateFromEnv()

systemStore
  .loadSystemCore()
  .catch((error: unknown) => {
    console.warn('Failed to load system core config', error)
  })
  .finally(() => {
    app.mount('#app')
  })
