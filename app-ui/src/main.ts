import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from '@/App.vue'
import { router } from '@/router'
import { useSystemStore } from '@/stores/system'
import { useTenantStore } from '@/stores/tenant'
import '@/styles/global.css'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
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
