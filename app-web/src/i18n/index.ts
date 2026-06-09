import { computed, ref } from 'vue'
import enUS from './locales/en-US'
import zhCN from './locales/zh-CN'

const messages = {
  'zh-CN': zhCN,
  'en-US': enUS,
} as const

export type Locale = keyof typeof messages
type MessageKey = keyof typeof zhCN

const LOCALE_STORAGE_KEY = 'app_locale'

function resolveInitialLocale(): Locale {
  if (typeof localStorage === 'undefined') return 'zh-CN'

  const savedLocale = localStorage.getItem(LOCALE_STORAGE_KEY)
  if (savedLocale === 'zh-CN' || savedLocale === 'en-US') return savedLocale

  return navigator.language.toLowerCase().startsWith('en') ? 'en-US' : 'zh-CN'
}

export const locale = ref<Locale>(resolveInitialLocale())

export const currentLanguageLabel = computed(() => messages[locale.value]['locale.short'])

export function setLocale(nextLocale: Locale) {
  locale.value = nextLocale
  localStorage.setItem(LOCALE_STORAGE_KEY, nextLocale)
  document.documentElement.lang = nextLocale
}

export function toggleLocale() {
  setLocale(locale.value === 'zh-CN' ? 'en-US' : 'zh-CN')
}

export function t(key: MessageKey) {
  return messages[locale.value][key]
}

setLocale(locale.value)
