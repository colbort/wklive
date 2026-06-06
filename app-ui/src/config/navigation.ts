export type AppNavIcon = 'nav-home' | 'nav-market' | 'nav-trade' | 'nav-assets' | 'nav-profile'

export type AppNavItem = {
  key: string
  labelKey: string
  path: string
  icon: AppNavIcon
}

export const appNavigation: AppNavItem[] = [
  { key: 'home', labelKey: 'nav.home', path: '/home', icon: 'nav-home' },
  { key: 'markets', labelKey: 'nav.markets', path: '/markets', icon: 'nav-market' },
  { key: 'trade', labelKey: 'nav.trade', path: '/trades', icon: 'nav-trade' },
  { key: 'wallet', labelKey: 'nav.wallet', path: '/assets', icon: 'nav-assets' },
  { key: 'profile', labelKey: 'nav.profile', path: '/profile', icon: 'nav-profile' },
]
