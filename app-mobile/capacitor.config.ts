import type { CapacitorConfig } from '@capacitor/cli'

const config: CapacitorConfig = {
  appId: 'com.wklive.app',
  appName: 'WkLive',
  webDir: 'dist',
  server: {
    androidScheme: 'https',
  },
}

export default config
