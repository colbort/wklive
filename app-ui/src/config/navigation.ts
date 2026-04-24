export type AppNavItem = {
  key: string
  label: string
  path: string
  icon: string
}

export type DesktopNavChild = {
  key: string
  symbol: string
  price: string
  change: string
  direction: 'up' | 'down' | 'flat'
  badge: string
}

export type DesktopNavItem = {
  key: string
  label: string
  path: string
  children?: DesktopNavChild[]
}

export const appNavigation: AppNavItem[] = [
  { key: 'home', label: '首页', path: '/home', icon: '⌂' },
  { key: 'markets', label: '市场', path: '/markets', icon: '⌁' },
  { key: 'trade', label: '交易', path: '/trades', icon: '⇄' },
  { key: 'wallet', label: '资产', path: '/assets', icon: '◇' },
  { key: 'profile', label: '用户', path: '/profile', icon: '◉' },
]

export const desktopNavigation: DesktopNavItem[] = [
  {
    key: 'crypto',
    label: '加密货币合约',
    path: '/markets',
    children: [
      { key: 'btc', symbol: 'BTC/USDT', price: '77677.65', change: '-2.07%', direction: 'down', badge: '₿' },
      { key: 'eth', symbol: 'ETH/USDT', price: '2331', change: '-3.11%', direction: 'down', badge: 'Ξ' },
      { key: 'bch', symbol: 'BCH/USDT', price: '455.03', change: '-2.16%', direction: 'down', badge: 'Ƀ' },
      { key: 'xrp', symbol: 'XRP/USDT', price: '1.41617', change: '-2.91%', direction: 'down', badge: '✕' },
      { key: 'ltc', symbol: 'LTC/USDT', price: '55.17', change: '-2.33%', direction: 'down', badge: 'Ł' },
      { key: 'doge', symbol: 'DOGE/USDT', price: '0.095726', change: '-2.28%', direction: 'down', badge: 'Ð' },
    ],
  },
  {
    key: 'stock',
    label: '股票',
    path: '/markets',
    children: [
      { key: 'aapl', symbol: 'AAPL', price: '213.48', change: '+0.62%', direction: 'up', badge: 'A' },
      { key: 'nvda', symbol: 'NVDA', price: '107.24', change: '+1.18%', direction: 'up', badge: 'N' },
      { key: 'tsla', symbol: 'TSLA', price: '171.95', change: '-0.44%', direction: 'down', badge: 'T' },
      { key: 'msft', symbol: 'MSFT', price: '421.86', change: '+0.27%', direction: 'up', badge: 'M' },
    ],
  },
  {
    key: 'forex',
    label: '外汇',
    path: '/markets',
    children: [
      { key: 'eurusd', symbol: 'EUR/USD', price: '1.0726', change: '+0.14%', direction: 'up', badge: '€' },
      { key: 'gbpusd', symbol: 'GBP/USD', price: '1.2544', change: '-0.09%', direction: 'down', badge: '£' },
      { key: 'usdjpy', symbol: 'USD/JPY', price: '154.33', change: '+0.22%', direction: 'up', badge: '$' },
      { key: 'usdchf', symbol: 'USD/CHF', price: '0.9086', change: '-0.11%', direction: 'down', badge: '₣' },
    ],
  },
  {
    key: 'commodity',
    label: '大宗商品',
    path: '/markets',
    children: [
      { key: 'gold', symbol: 'XAU/USD', price: '2338.40', change: '+0.38%', direction: 'up', badge: 'Au' },
      { key: 'silver', symbol: 'XAG/USD', price: '27.41', change: '+0.21%', direction: 'up', badge: 'Ag' },
      { key: 'wti', symbol: 'WTI', price: '82.12', change: '-0.54%', direction: 'down', badge: 'O' },
      { key: 'ng', symbol: 'NATGAS', price: '2.18', change: '-1.02%', direction: 'down', badge: 'G' },
    ],
  },
  {
    key: 'option',
    label: '期权合约',
    path: '/markets',
    children: [
      { key: 'btc-call', symbol: 'BTC 90K CALL', price: '1280', change: '+4.16%', direction: 'up', badge: 'C' },
      { key: 'btc-put', symbol: 'BTC 80K PUT', price: '940', change: '-1.24%', direction: 'down', badge: 'P' },
      { key: 'eth-call', symbol: 'ETH 3K CALL', price: '146', change: '+2.87%', direction: 'up', badge: 'C' },
      { key: 'eth-put', symbol: 'ETH 2.2K PUT', price: '88', change: '-0.65%', direction: 'down', badge: 'P' },
    ],
  },
  { key: 'license', label: '公司资质', path: '/home' },
  { key: 'whitepaper', label: '白皮书', path: '/home' },
  { key: 'compliance', label: '监管文件', path: '/home' },
]
