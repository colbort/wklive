export type AppNavItem = {
  key: string
  label: string
  path: string
  icon: string
}

export const appNavigation: AppNavItem[] = [
  { key: 'home', label: '首页', path: '/home', icon: '⌂' },
  { key: 'markets', label: '市场', path: '/markets', icon: '⌁' },
  { key: 'trade', label: '交易', path: '/assets', icon: '⇄' },
  { key: 'wallet', label: '资产', path: '/assets', icon: '◇' },
  { key: 'profile', label: '用户', path: '/profile', icon: '◉' },
]
