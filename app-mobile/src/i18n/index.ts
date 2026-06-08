import { computed, inject, reactive, type App } from 'vue'

import enUS from './locales/en-US'
import zhCN from './locales/zh-CN'

export const locales = {
  'zh-CN': zhCN,
  'en-US': enUS,
} as const

export type AppLocale = keyof typeof locales
export type MessageSchema = typeof zhCN

const STORAGE_KEY = 'app-mobile-locale'
const I18N_KEY = Symbol('app-i18n')

type Params = Record<string, string | number>
type I18nApi = ReturnType<typeof createI18nApi>

const state = reactive({
  locale: getInitialLocale(),
})

function getInitialLocale(): AppLocale {
  const stored = localStorage.getItem(STORAGE_KEY) as AppLocale | null
  if (stored && stored in locales) return stored

  const language = navigator.language.toLowerCase()
  return language.startsWith('zh') ? 'zh-CN' : 'en-US'
}

function resolveMessage(path: string, locale: AppLocale) {
  const parts = path.split('.').filter(Boolean)
  let current: unknown = locales[locale]

  for (const part of parts) {
    if (!current || typeof current !== 'object' || !(part in current)) {
      return undefined
    }
    current = (current as Record<string, unknown>)[part]
  }

  return typeof current === 'string' ? current : undefined
}

export function t(path: string, params?: Params) {
  const message = resolveMessage(path, state.locale) || resolveMessage(path, 'zh-CN') || path
  if (!params) return message

  return message.replace(/\{(\w+)\}/g, (_, key: string) => String(params[key] ?? `{${key}}`))
}

export function translateApiError(code: number | string | null | undefined, fallback = '') {
  const numericCode = Number(code)
  const path = Number.isFinite(numericCode) ? `apiErrors.${numericCode}` : ''
  return (path && (resolveMessage(path, state.locale) || resolveMessage(path, 'zh-CN'))) || fallback
}

export function setLocale(locale: AppLocale) {
  state.locale = locale
  localStorage.setItem(STORAGE_KEY, locale)
  document.documentElement.lang = locale
}

export function getLocale() {
  return state.locale
}

export function toggleLocale() {
  setLocale(state.locale === 'zh-CN' ? 'en-US' : 'zh-CN')
}

function createI18nApi() {
  return {
    locale: computed(() => state.locale),
    t,
    setLocale,
    toggleLocale,
    translateApiError,
  }
}

export function createI18n() {
  const api = createI18nApi()
  return {
    install(app: App) {
      document.documentElement.lang = state.locale
      app.provide(I18N_KEY, api)
      app.config.globalProperties.$t = t
    },
  }
}

export function useI18n() {
  return inject<I18nApi>(I18N_KEY) || createI18nApi()
}
