import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import elZhCN from 'element-plus/es/locale/lang/zh-cn'
import elEnUS from 'element-plus/es/locale/lang/en'
import 'element-plus/dist/index.css'

import App from './App.vue'
import { router } from './router'
import { i18n, elLocaleMap } from './i18n'
import { setupPermDirective } from './directives/perm'

const app = createApp(App)

// Get initial locale from i18n
const initialLocale = i18n.global.locale.value as keyof typeof elLocaleMap
const epLocale = elLocaleMap[initialLocale] || elLocaleMap['zh-CN']

app.use(createPinia())
app.use(i18n)
app.use(router)
app.use(ElementPlus, { locale: epLocale })

setupPermDirective(app)

app.mount('#app')
