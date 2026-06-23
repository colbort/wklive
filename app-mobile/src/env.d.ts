/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_APP_TARGET?: 'web' | 'capacitor'
  readonly VITE_API_BASE_URL?: string
  readonly VITE_DEV_PROXY_TARGET?: string
  readonly VITE_API_BASE_PATH?: string
  readonly VITE_ROUTER_BASE?: string
  readonly VITE_TENANT_CODE?: string
  readonly VITE_CHAT_UI_URL?: string
  readonly VITE_CHAT_WS_URL?: string
  readonly VITE_CHAT_TOKEN?: string
  readonly VITE_CHAT_API_KEY?: string
  readonly VITE_CHAT_API_SECRET?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

declare module '*.vue' {
  import type { DefineComponent } from 'vue'

  const component: DefineComponent<Record<string, never>, Record<string, never>, any>
  export default component
}
