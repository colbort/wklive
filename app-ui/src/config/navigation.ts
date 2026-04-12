export type AppNavItem = {
  key: string
  label: string
  path: string
  icon: string
}

export const appNavigation: AppNavItem[] = [
  { key: 'home', label: '首页', path: '/home', icon: 'Ho' },
  { key: 'markets', label: '市场', path: '/markets', icon: 'Mk' },
  { key: 'assets', label: '交易', path: '/assets', icon: 'Tr' },
  { key: 'profile', label: '我的', path: '/profile', icon: 'Me' },
]
