import { createApp } from 'vue';
import { createPinia } from 'pinia';
import { watch } from 'vue';
import ElementPlus from 'element-plus';
import 'element-plus/dist/index.css';
import App from './App.vue';
import { router } from './router';
import { i18n, elLocaleMap } from './i18n';
import { setupPermDirective } from './directives/perm';
const app = createApp(App);
// Get initial locale from i18n
const initialLocale = i18n.global.locale.value;
const epLocale = elLocaleMap[initialLocale] || elLocaleMap['zh-CN'];
app.use(createPinia());
app.use(i18n);
app.use(router);
app.use(ElementPlus, { locale: epLocale });
setupPermDirective(app);
// Watch for locale changes and update Element Plus
watch(() => i18n.global.locale.value, (newLocale) => {
    const elLocale = elLocaleMap[newLocale] || elLocaleMap['zh-CN'];
    app.config.globalProperties.$ELEMENT = { locale: elLocale };
});
app.mount('#app');
