<script setup lang="ts">
import { computed, nextTick, ref } from 'vue'
import { useRouter } from 'vue-router'

import { getTenantCode } from '@/api/http'
import { apiLogin } from '@/api/userPublic'
import { useI18n } from '@/i18n'

const LOGIN_TYPE_PHONE = 2
const LOGIN_TYPE_EMAIL = 3
const { t, toggleLocale } = useI18n()

type EthereumProvider = {
  providers?: EthereumProvider[]
  isMetaMask?: boolean
  isBraveWallet?: boolean
  isCoinbaseWallet?: boolean
  isOkxWallet?: boolean
  isTokenPocket?: boolean
  isTrust?: boolean
  isBitKeep?: boolean
  isBitgetWallet?: boolean
  isImToken?: boolean
  isMathWallet?: boolean
  isSafePal?: boolean
  isRabby?: boolean
  isPhantom?: boolean
  isOneKey?: boolean
  isGateWallet?: boolean
  isCoin98?: boolean
  isExodus?: boolean
  isOpera?: boolean
  isFrame?: boolean
  isTally?: boolean
  isZerion?: boolean
  isRainbow?: boolean
  isBybit?: boolean
  isKuCoinWallet?: boolean
  isHaloWallet?: boolean
  isSubWallet?: boolean
  isXDEFI?: boolean
  isKeplr?: boolean
  isCosmostation?: boolean
  isWalletConnect?: boolean
  selectedProvider?: EthereumProvider
  ethereum?: EthereumProvider
  request: <T = unknown>(args: { method: string; params?: unknown[] }) => Promise<T>
  [key: string]: unknown
}

type TronLinkProvider = {
  request?: <T = unknown>(args: { method: string; params?: unknown[] }) => Promise<T>
  ready?: boolean
}

type TronWebProvider = {
  defaultAddress?: {
    base58?: string
    hex?: string
  }
  toHex?: (message: string) => string
  trx?: {
    sign?: (message: string) => Promise<string>
    signMessageV2?: (message: string) => Promise<string>
  }
}

type WalletProvider = EthereumProvider | TronLinkProvider

type WalletOption = {
  key: string
  name: string
  provider: WalletProvider | null
  installed: boolean
  installUrl: string
  chainType: 'evm' | 'tron'
}

type WalletDefinition = {
  key: string
  name: string
  installUrl: string
  chainType: 'evm' | 'tron'
  detect: (provider: EthereumProvider) => boolean
}

declare global {
  interface Window {
    ethereum?: EthereumProvider
    okxwallet?: EthereumProvider
    trustwallet?: EthereumProvider
    BinanceChain?: EthereumProvider
    bitkeep?: { ethereum?: EthereumProvider }
    bitgetWallet?: EthereumProvider
    imToken?: EthereumProvider
    mathwallet?: EthereumProvider
    safepalProvider?: EthereumProvider
    rabby?: EthereumProvider
    phantom?: { ethereum?: EthereumProvider }
    oneKeyWallet?: EthereumProvider
    gatewallet?: EthereumProvider
    coin98?: { provider?: EthereumProvider }
    exodus?: { ethereum?: EthereumProvider }
    tally?: EthereumProvider
    zerionWallet?: EthereumProvider
    rainbow?: EthereumProvider
    bybitWallet?: EthereumProvider
    kucoinWallet?: EthereumProvider
    haloWallet?: EthereumProvider
    subwallet?: EthereumProvider
    xfi?: { ethereum?: EthereumProvider }
    keplr?: EthereumProvider
    cosmostation?: { ethereum?: EthereumProvider }
    tronLink?: TronLinkProvider
    tronWeb?: TronWebProvider
  }
}

const router = useRouter()
const activeTab = ref<'email' | 'phone'>('email')
const account = ref('')
const password = ref('')
const remember = ref(false)
const showPassword = ref(false)
const googleCode = ref<string[]>(Array(6).fill(''))
const googleCodeInputs = ref<HTMLInputElement[]>([])
const submitting = ref(false)
const walletConnecting = ref(false)
const walletAddress = ref('')
const walletError = ref('')
const walletSheetOpen = ref(false)
const pendingWallet = ref<WalletOption | null>(null)
const walletRequestId = ref(0)
const errorMessage = ref('')

const accountPlaceholder = computed(() =>
  activeTab.value === 'email' ? t('auth.yourEmail') : t('auth.phonePlaceholder'),
)
const loginType = computed(() => (activeTab.value === 'email' ? LOGIN_TYPE_EMAIL : LOGIN_TYPE_PHONE))
const googleCodeValue = computed(() => googleCode.value.join(''))
const shortWalletAddress = computed(() => {
  if (!walletAddress.value) return ''
  return `${walletAddress.value.slice(0, 6)}...${walletAddress.value.slice(-4)}`
})
const walletOptions = computed<WalletOption[]>(() => {
  const providers = getWalletProviders()
  const usedProviders = new Set<EthereumProvider>()
  const options = walletDefinitions.map((wallet) => {
    const provider = wallet.chainType === 'tron'
      ? getTronProvider()
      : providers.find((item) => wallet.detect(item)) || null
    if (provider && wallet.chainType === 'evm') usedProviders.add(provider as EthereumProvider)
    return {
      key: wallet.key,
      name: wallet.name,
      provider,
      installed: Boolean(provider),
      installUrl: wallet.installUrl,
      chainType: wallet.chainType,
    }
  })
  const fallback = providers.find((provider) => !usedProviders.has(provider)) || null

  const wallets = [
    ...options,
    {
      key: 'browser',
      name: 'Browser Wallet',
      provider: fallback,
      installed: Boolean(fallback),
      installUrl: 'https://metamask.io/download/',
      chainType: 'evm' as const,
    },
  ]
  return wallets.sort((left, right) => Number(right.installed) - Number(left.installed))
})

const walletDefinitions: WalletDefinition[] = [
  { key: 'tronlink', name: 'TronLink', installUrl: 'https://www.tronlink.org/', chainType: 'tron', detect: () => false },
  { key: 'brave', name: 'Brave Wallet', installUrl: 'https://brave.com/wallet/', chainType: 'evm', detect: (provider) => Boolean(provider.isBraveWallet) },
  { key: 'metamask', name: 'MetaMask', installUrl: 'https://metamask.io/download/', chainType: 'evm', detect: (provider) => Boolean(provider.isMetaMask && !provider.isBraveWallet) },
  { key: 'okx', name: 'OKX Wallet', installUrl: 'https://www.okx.com/web3', chainType: 'evm', detect: (provider) => Boolean(provider.isOkxWallet || provider === window.okxwallet) },
  { key: 'tokenpocket', name: 'TokenPocket', installUrl: 'https://www.tokenpocket.pro/', chainType: 'evm', detect: (provider) => Boolean(provider.isTokenPocket) },
  { key: 'trust', name: 'Trust Wallet', installUrl: 'https://trustwallet.com/download', chainType: 'evm', detect: (provider) => Boolean(provider.isTrust || provider === window.trustwallet) },
  { key: 'bitget', name: 'Bitget Wallet', installUrl: 'https://web3.bitget.com/', chainType: 'evm', detect: (provider) => Boolean(provider.isBitKeep || provider.isBitgetWallet || provider === window.bitgetWallet || provider === window.bitkeep?.ethereum) },
  { key: 'coinbase', name: 'Coinbase Wallet', installUrl: 'https://www.coinbase.com/wallet/downloads', chainType: 'evm', detect: (provider) => Boolean(provider.isCoinbaseWallet) },
  { key: 'imtoken', name: 'imToken', installUrl: 'https://token.im/', chainType: 'evm', detect: (provider) => Boolean(provider.isImToken || provider === window.imToken) },
  { key: 'math', name: 'MathWallet', installUrl: 'https://mathwallet.org/', chainType: 'evm', detect: (provider) => Boolean(provider.isMathWallet || provider === window.mathwallet) },
  { key: 'safepal', name: 'SafePal', installUrl: 'https://www.safepal.com/download', chainType: 'evm', detect: (provider) => Boolean(provider.isSafePal || provider === window.safepalProvider) },
  { key: 'rabby', name: 'Rabby Wallet', installUrl: 'https://rabby.io/', chainType: 'evm', detect: (provider) => Boolean(provider.isRabby || provider === window.rabby) },
  { key: 'phantom', name: 'Phantom', installUrl: 'https://phantom.app/download', chainType: 'evm', detect: (provider) => Boolean(provider.isPhantom || provider === window.phantom?.ethereum) },
  { key: 'binance', name: 'Binance Wallet', installUrl: 'https://www.binance.com/web3wallet', chainType: 'evm', detect: (provider) => Boolean(provider === window.BinanceChain) },
  { key: 'onekey', name: 'OneKey Wallet', installUrl: 'https://onekey.so/download/', chainType: 'evm', detect: (provider) => Boolean(provider.isOneKey || provider === window.oneKeyWallet) },
  { key: 'gate', name: 'Gate Wallet', installUrl: 'https://www.gate.io/web3', chainType: 'evm', detect: (provider) => Boolean(provider.isGateWallet || provider === window.gatewallet) },
  { key: 'coin98', name: 'Coin98 Wallet', installUrl: 'https://coin98.com/wallet', chainType: 'evm', detect: (provider) => Boolean(provider.isCoin98 || provider === window.coin98?.provider) },
  { key: 'exodus', name: 'Exodus', installUrl: 'https://www.exodus.com/download/', chainType: 'evm', detect: (provider) => Boolean(provider.isExodus || provider === window.exodus?.ethereum) },
  { key: 'opera', name: 'Opera Wallet', installUrl: 'https://www.opera.com/crypto/next', chainType: 'evm', detect: (provider) => Boolean(provider.isOpera) },
  { key: 'frame', name: 'Frame', installUrl: 'https://frame.sh/', chainType: 'evm', detect: (provider) => Boolean(provider.isFrame) },
  { key: 'tally', name: 'Taho/Tally', installUrl: 'https://taho.xyz/', chainType: 'evm', detect: (provider) => Boolean(provider.isTally || provider === window.tally) },
  { key: 'zerion', name: 'Zerion Wallet', installUrl: 'https://zerion.io/wallet', chainType: 'evm', detect: (provider) => Boolean(provider.isZerion || provider === window.zerionWallet) },
  { key: 'rainbow', name: 'Rainbow', installUrl: 'https://rainbow.me/', chainType: 'evm', detect: (provider) => Boolean(provider.isRainbow || provider === window.rainbow) },
  { key: 'bybit', name: 'Bybit Wallet', installUrl: 'https://www.bybit.com/web3', chainType: 'evm', detect: (provider) => Boolean(provider.isBybit || provider === window.bybitWallet) },
  { key: 'kucoin', name: 'KuCoin Wallet', installUrl: 'https://www.kucoin.com/web3', chainType: 'evm', detect: (provider) => Boolean(provider.isKuCoinWallet || provider === window.kucoinWallet) },
  { key: 'halo', name: 'Halo Wallet', installUrl: 'https://halo.social/wallet', chainType: 'evm', detect: (provider) => Boolean(provider.isHaloWallet || provider === window.haloWallet) },
  { key: 'subwallet', name: 'SubWallet', installUrl: 'https://www.subwallet.app/download.html', chainType: 'evm', detect: (provider) => Boolean(provider.isSubWallet || provider === window.subwallet) },
  { key: 'xdefi', name: 'XDEFI Wallet', installUrl: 'https://www.xdefi.io/', chainType: 'evm', detect: (provider) => Boolean(provider.isXDEFI || provider === window.xfi?.ethereum) },
  { key: 'keplr', name: 'Keplr', installUrl: 'https://www.keplr.app/download', chainType: 'evm', detect: (provider) => Boolean(provider.isKeplr || provider === window.keplr) },
  { key: 'cosmostation', name: 'Cosmostation Wallet', installUrl: 'https://cosmostation.io/products/cosmostation_wallet', chainType: 'evm', detect: (provider) => Boolean(provider.isCosmostation || provider === window.cosmostation?.ethereum) },
  { key: 'walletconnect', name: 'WalletConnect', installUrl: 'https://walletconnect.network/', chainType: 'evm', detect: (provider) => Boolean(provider.isWalletConnect) },
]

function goBack() {
  if (window.history.length > 1) {
    router.back()
    return
  }
  router.push('/profile')
}

function setGoogleCodeInputRef(element: unknown, index: number) {
  if (element instanceof HTMLInputElement) googleCodeInputs.value[index] = element
}

function focusGoogleCodeInput(index: number) {
  const target = googleCodeInputs.value[Math.max(0, Math.min(index, googleCode.value.length - 1))]
  if (!target) return
  nextTick(() => {
    target.focus()
    target.select()
  })
}

function applyGoogleCodeDigits(index: number, value: string) {
  const digits = value.replace(/\D/g, '').slice(0, googleCode.value.length - index)
  if (!digits) {
    googleCode.value[index] = ''
    return
  }

  digits.split('').forEach((digit, offset) => {
    googleCode.value[index + offset] = digit
  })

  const nextIndex = Math.min(index + digits.length, googleCode.value.length - 1)
  focusGoogleCodeInput(nextIndex)
}

function handleGoogleCodeInput(index: number, event: Event) {
  applyGoogleCodeDigits(index, (event.target as HTMLInputElement).value)
}

function selectGoogleCodeInput(event: FocusEvent) {
  const target = event.target as HTMLInputElement
  target.select()
}

function handleGoogleCodeKeydown(index: number, event: KeyboardEvent) {
  if (event.key === 'Backspace') {
    event.preventDefault()
    if (googleCode.value[index]) {
      googleCode.value[index] = ''
      return
    }
    if (index > 0) {
      googleCode.value[index - 1] = ''
      focusGoogleCodeInput(index - 1)
    }
    return
  }

  if (event.key === 'Delete') {
    event.preventDefault()
    googleCode.value[index] = ''
    return
  }

  if (event.key === 'ArrowLeft' && index > 0) {
    event.preventDefault()
    focusGoogleCodeInput(index - 1)
    return
  }

  if (event.key === 'ArrowRight' && index < googleCode.value.length - 1) {
    event.preventDefault()
    focusGoogleCodeInput(index + 1)
  }
}

function handleGoogleCodePaste(index: number, event: ClipboardEvent) {
  event.preventDefault()
  applyGoogleCodeDigits(index, event.clipboardData?.getData('text') || '')
}

async function submitLogin() {
  if (submitting.value) return

  errorMessage.value = ''
  const tenantCode = getTenantCode()
  if (!tenantCode) {
    errorMessage.value = t('profile.tenantMissing')
    return
  }
  if (!account.value.trim() || !password.value) {
    errorMessage.value = t('auth.inputAccountPassword')
    return
  }
  if (googleCodeValue.value.length > 0 && googleCodeValue.value.length < googleCode.value.length) {
    errorMessage.value = t('auth.inputFullGoogleCode')
    focusGoogleCodeInput(googleCode.value.findIndex((digit) => !digit))
    return
  }

  submitting.value = true
  try {
    const res = await apiLogin({
      tenantCode,
      loginType: loginType.value,
      account: account.value.trim(),
      password: password.value,
      googleCode: googleCodeValue.value || undefined,
    })
    if (res.code !== 200) {
      errorMessage.value = res.msg || t('profile.loginFailed')
      if (res.code === 2057) focusGoogleCodeInput(0)
      return
    }
    router.replace('/profile')
  } catch (error) {
    console.warn('login failed', error)
    errorMessage.value = t('profile.loginFailed')
  } finally {
    submitting.value = false
  }
}

function getWalletProviders() {
  const providers: EthereumProvider[] = []
  const seen = new Set<EthereumProvider>()
  const addProvider = (provider?: EthereumProvider | null) => {
    if (!provider || typeof provider.request !== 'function' || seen.has(provider)) return
    seen.add(provider)
    providers.push(provider)
  }

  const ethereum = window.ethereum
  if (Array.isArray(ethereum?.providers) && ethereum.providers.length) {
    ethereum.providers.forEach(addProvider)
  }
  addProvider(ethereum?.selectedProvider)
  addProvider(ethereum)
  addProvider(window.okxwallet)
  addProvider(window.trustwallet)
  addProvider(window.BinanceChain)
  addProvider(window.bitkeep?.ethereum)
  addProvider(window.bitgetWallet)
  addProvider(window.imToken)
  addProvider(window.mathwallet)
  addProvider(window.safepalProvider)
  addProvider(window.rabby)
  addProvider(window.phantom?.ethereum)
  addProvider(window.oneKeyWallet)
  addProvider(window.gatewallet)
  addProvider(window.coin98?.provider)
  addProvider(window.exodus?.ethereum)
  addProvider(window.tally)
  addProvider(window.zerionWallet)
  addProvider(window.rainbow)
  addProvider(window.bybitWallet)
  addProvider(window.kucoinWallet)
  addProvider(window.haloWallet)
  addProvider(window.subwallet)
  addProvider(window.xfi?.ethereum)
  addProvider(window.keplr)
  addProvider(window.cosmostation?.ethereum)

  return providers
}

function getTronProvider() {
  if (window.tronLink || window.tronWeb?.defaultAddress?.base58) {
    return window.tronLink || ({ ready: true } as TronLinkProvider)
  }
  return null
}

function openWalletSheet() {
  walletError.value = ''
  errorMessage.value = ''
  pendingWallet.value = null
  walletSheetOpen.value = true
}

function closeWalletSheet() {
  walletRequestId.value += 1
  walletSheetOpen.value = false
  pendingWallet.value = null
}

function installWallet(wallet: WalletOption) {
  walletError.value = ''
  window.open(wallet.installUrl, '_blank', 'noopener,noreferrer')
}

function handleWalletSelected(wallet: WalletOption) {
  if (wallet.installed) {
    pendingWallet.value = wallet
    if (wallet.chainType === 'tron') {
      connectTronWallet(wallet.provider as TronLinkProvider)
      return
    }
    connectWallet(wallet.provider as EthereumProvider)
    return
  }
  installWallet(wallet)
}

function backToWalletList() {
  walletRequestId.value += 1
  walletConnecting.value = false
  walletError.value = ''
  pendingWallet.value = null
}

function retryPendingWallet() {
  if (!pendingWallet.value) return
  walletRequestId.value += 1
  walletConnecting.value = false
  handleWalletSelected(pendingWallet.value)
}

function isUserRejectedWalletError(error: any) {
  const message = String(error?.message || '').toLowerCase()
  return error?.code === 4001 || error?.code === '4001' || message.includes('user rejected') || message.includes('user denied') || message.includes('cancel')
}

function isDefiniteWalletError(error: any) {
  const message = String(error?.message || '').toLowerCase()
  if (isUserRejectedWalletError(error)) return true
  return (
    error?.code === -32002 ||
    error?.code === '-32002' ||
    message.includes('already pending') ||
    message.includes('invalid transaction') ||
    message.includes('unsupported') ||
    message.includes('not support')
  )
}

function wait(ms: number) {
  return new Promise((resolve) => window.setTimeout(resolve, ms))
}

async function waitForTronAddress() {
  const startedAt = Date.now()
  const timeout = 90000
  const requestId = walletRequestId.value
  while (walletSheetOpen.value && pendingWallet.value?.key === 'tronlink' && requestId === walletRequestId.value) {
    const address = window.tronWeb?.defaultAddress?.base58
    if (address) return address
    if (Date.now() - startedAt >= timeout) return ''
    await wait(800)
  }
  return ''
}

async function connectWallet(selectedProvider?: EthereumProvider | null) {
  if (walletConnecting.value) return

  const requestId = walletRequestId.value + 1
  walletRequestId.value = requestId
  walletError.value = ''
  errorMessage.value = ''

  const provider = selectedProvider || (walletOptions.value.find((wallet) => wallet.installed && wallet.chainType === 'evm')?.provider as EthereumProvider | null) || null
  if (!provider) {
    walletError.value = t('auth.walletNotDetected')
    return
  }

  walletConnecting.value = true
  try {
    const accounts = await provider.request<string[]>({ method: 'eth_requestAccounts' })
    if (requestId !== walletRequestId.value) return
    const address = accounts?.[0]
    if (!address) {
      walletError.value = t('auth.walletAddressMissing')
      return
    }

    walletAddress.value = address
    const message = `Wklive wallet login\nAddress: ${address}\nTime: ${Date.now()}`
    await provider.request({
      method: 'personal_sign',
      params: [message, address],
    })
    if (requestId !== walletRequestId.value) return
    walletSheetOpen.value = false
    pendingWallet.value = null
  } catch (error: any) {
    console.warn('connect wallet failed', error)
    walletError.value = isUserRejectedWalletError(error)
      ? t('auth.walletCanceled')
      : t('auth.walletConnectFailed')
  } finally {
    if (requestId === walletRequestId.value) walletConnecting.value = false
  }
}

async function connectTronWallet(selectedProvider?: TronLinkProvider | null) {
  if (walletConnecting.value) return

  const requestId = walletRequestId.value + 1
  walletRequestId.value = requestId
  walletError.value = ''
  errorMessage.value = ''
  walletConnecting.value = true
  try {
    if (selectedProvider?.request) {
      await selectedProvider.request({ method: 'tron_requestAccounts' }).catch((error: any) => {
        if (isDefiniteWalletError(error)) throw error
        console.warn('tronlink request pending', error)
      })
    }

    const address = await waitForTronAddress()
    if (requestId !== walletRequestId.value) return
    if (!address) {
      walletError.value = walletSheetOpen.value ? t('auth.walletTimeout') : ''
      return
    }

    walletAddress.value = address
    const message = `Wklive wallet login\nAddress: ${address}\nTime: ${Date.now()}`
    if (window.tronWeb?.trx?.signMessageV2) {
      await window.tronWeb.trx.signMessageV2(message)
    } else if (window.tronWeb?.trx?.sign && window.tronWeb?.toHex) {
      await window.tronWeb.trx.sign(window.tronWeb.toHex(message))
    }
    if (requestId !== walletRequestId.value) return
    walletSheetOpen.value = false
    pendingWallet.value = null
  } catch (error: any) {
    console.warn('connect tronlink failed', error)
    walletError.value = isUserRejectedWalletError(error)
      ? t('auth.walletCanceled')
      : t('auth.tronConnectFailed')
  } finally {
    if (requestId === walletRequestId.value) walletConnecting.value = false
  }
}
</script>

<template>
  <section class="auth-page">
    <header class="auth-topbar">
      <button type="button" class="icon-button" :aria-label="t('common.back')" @click="goBack">
        <span class="chevron-left" />
      </button>
      <button type="button" class="icon-button" :aria-label="t('common.language')" @click="toggleLocale">
        <span class="globe-icon" />
      </button>
    </header>

    <main class="auth-content">
      <h1>{{ t('auth.loginTitle') }}</h1>

      <div class="auth-tabs" role="tablist" :aria-label="t('auth.loginMethod')">
        <button
          type="button"
          :class="{ active: activeTab === 'email' }"
          role="tab"
          :aria-selected="activeTab === 'email'"
          @click="activeTab = 'email'"
        >
          {{ t('auth.email') }}
        </button>
        <button
          type="button"
          :class="{ active: activeTab === 'phone' }"
          role="tab"
          :aria-selected="activeTab === 'phone'"
          @click="activeTab = 'phone'"
        >
          {{ t('auth.phone') }}
        </button>
      </div>

      <form class="auth-form" @submit.prevent="submitLogin">
        <label class="auth-field">
          <span v-if="activeTab === 'email'" class="field-icon mail-icon" />
          <span v-else class="phone-prefix">+1 <i /></span>
          <input v-model="account" :placeholder="accountPlaceholder" autocomplete="username" />
        </label>

        <label class="auth-field">
          <span class="field-icon lock-icon" />
          <input
            v-model="password"
            :type="showPassword ? 'text' : 'password'"
            :placeholder="t('auth.passwordPlaceholder')"
            autocomplete="current-password"
          />
          <button
            type="button"
            class="field-action"
            :aria-label="t('security.togglePassword')"
            @click="showPassword = !showPassword"
          >
            <span class="eye-off-icon" />
          </button>
        </label>

        <div class="google-code-field">
          <span>{{ t('auth.googleCode') }}</span>
          <div class="google-code-boxes" :aria-label="t('auth.googleCode')">
            <input
              v-for="(_, index) in googleCode"
              :key="index"
              :ref="(element) => setGoogleCodeInputRef(element, index)"
              :value="googleCode[index]"
              inputmode="numeric"
              autocomplete="one-time-code"
              maxlength="1"
              :aria-label="t('auth.googleCode')"
              @focus="selectGoogleCodeInput"
              @input="handleGoogleCodeInput(index, $event)"
              @keydown="handleGoogleCodeKeydown(index, $event)"
              @paste="handleGoogleCodePaste(index, $event)"
            />
          </div>
        </div>

        <div class="auth-row">
          <label class="remember-control">
            <input v-model="remember" type="checkbox" />
            <span />
            <em>{{ t('auth.remember') }}</em>
          </label>
          <RouterLink to="/forgot-password">{{ t('auth.forgotPassword') }}</RouterLink>
        </div>

        <p v-if="errorMessage" class="auth-error">{{ errorMessage }}</p>

        <button type="submit" class="primary-button" :disabled="submitting">
          {{ submitting ? t('auth.loggingIn') : t('auth.loginTitle') }}
        </button>
      </form>

      <div class="quick-login">
        <div class="quick-login__divider">
          <span />
          <strong>{{ t('auth.quickLogin') }}</strong>
          <span />
        </div>
        <button type="button" class="wallet-button" :disabled="walletConnecting" @click="openWalletSheet">
          <span class="wallet-icon" />
        </button>
        <strong>{{ walletConnecting ? t('auth.connecting') : walletAddress ? shortWalletAddress : t('auth.connectWallet') }}</strong>
        <p v-if="walletError" class="wallet-error">{{ walletError }}</p>
      </div>

      <p class="auth-switch">
        {{ t('auth.noAccount') }}
        <button type="button" @click="router.push('/register')">{{ t('auth.goRegister') }}</button>
      </p>
    </main>

    <div
      v-if="walletSheetOpen"
      class="wallet-sheet-layer"
      @click.self="closeWalletSheet"
    >
      <div class="wallet-sheet" role="dialog" aria-modal="true" aria-label="Connect Wallet">
        <header class="wallet-sheet__header">
          <button
            v-if="pendingWallet"
            type="button"
            class="wallet-sheet__back"
            :aria-label="t('common.back')"
            @click="backToWalletList"
          >
            ‹
          </button>
          <button v-else type="button" class="wallet-sheet__help" :aria-label="t('auth.help')">?</button>
          <strong>{{ pendingWallet ? pendingWallet.name : 'Connect Wallet' }}</strong>
          <button
            type="button"
            class="wallet-sheet__close"
            :aria-label="t('common.close')"
            @click="closeWalletSheet"
          >
            ×
          </button>
        </header>

        <div v-if="pendingWallet" class="wallet-pending">
          <span class="wallet-pending__icon" :class="`wallet-row__icon--${pendingWallet.key}`">
            {{ pendingWallet.name.slice(0, 1) }}
          </span>
          <strong>{{ t('auth.continueInWallet', { wallet: pendingWallet.name }) }}</strong>
          <p>{{ t('auth.acceptWalletRequest') }}</p>
          <button type="button" @click="retryPendingWallet">
            <span>↻</span>
            {{ t('auth.tryAgain') }}
          </button>
        </div>

        <div v-else class="wallet-list">
          <button
            v-for="wallet in walletOptions"
            :key="wallet.key"
            type="button"
            class="wallet-row"
            :disabled="walletConnecting"
            @click="handleWalletSelected(wallet)"
          >
            <span class="wallet-row__icon" :class="`wallet-row__icon--${wallet.key}`">
              {{ wallet.name.slice(0, 1) }}
            </span>
            <span>{{ wallet.name }}</span>
            <em :class="{ install: !wallet.installed }">{{ wallet.installed ? t('auth.installed') : t('auth.install') }}</em>
          </button>

          <button type="button" class="wallet-search" disabled>
            <span class="wallet-search__icon" />
            <span>{{ t('auth.searchWallet') }}</span>
            <em>{{ t('auth.local') }}</em>
          </button>
        </div>

        <p v-if="walletError" class="wallet-sheet__error">{{ walletError }}</p>
        <p class="wallet-sheet__hint">{{ t('auth.walletHint') }}</p>
      </div>
    </div>
  </section>
</template>

<style scoped>
.auth-page {
  width: 100%;
  max-width: 100%;
  height: 100dvh;
  min-height: 100dvh;
  overflow-x: hidden;
  overflow-y: auto;
  overscroll-behavior-x: none;
  -webkit-overflow-scrolling: touch;
  margin: 0 auto;
  padding: 24px 28px 42px;
  background: #0d0e17;
  color: #fff;
}

.auth-topbar {
  position: sticky;
  top: 0;
  z-index: 30;
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-width: 0;
  margin: -24px -28px 0;
  padding: 24px 28px 10px;
  background: #0d0e17;
}

.icon-button {
  display: inline-flex;
  width: 54px;
  height: 54px;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 50%;
  background: #252733;
  color: #fff;
}

.chevron-left {
  width: 18px;
  height: 18px;
  border-left: 4px solid currentColor;
  border-bottom: 4px solid currentColor;
  transform: rotate(45deg);
}

.globe-icon {
  position: relative;
  width: 28px;
  height: 28px;
  border: 4px solid currentColor;
  border-radius: 50%;
}

.globe-icon::before,
.globe-icon::after {
  content: '';
  position: absolute;
  inset: 3px 8px;
  border-left: 3px solid currentColor;
  border-right: 3px solid currentColor;
  border-radius: 50%;
}

.globe-icon::after {
  inset: 10px -3px auto;
  height: 3px;
  border: 0;
  background: currentColor;
}

.auth-content {
  width: 100%;
  min-width: 0;
  padding-top: 106px;
}

.auth-content h1 {
  margin: 0 0 76px;
  font-size: 48px;
  line-height: 1;
  font-weight: 900;
  letter-spacing: 0;
}

.auth-tabs {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  border-bottom: 1px solid #20222d;
}

.auth-tabs button {
  position: relative;
  height: 54px;
  border: 0;
  background: transparent;
  color: #8e9099;
  font-size: 26px;
  font-weight: 800;
}

.auth-tabs button.active {
  color: #00c313;
}

.auth-tabs button.active::after {
  content: '';
  position: absolute;
  right: 0;
  bottom: -1px;
  left: 0;
  height: 5px;
  background: #00c313;
}

.auth-form {
  display: grid;
  gap: 30px;
  margin-top: 64px;
}

.auth-field {
  display: flex;
  min-height: 102px;
  align-items: center;
  gap: 14px;
  border-radius: 28px;
  background: #20212b;
  padding: 0 22px;
}

.auth-field input {
  min-width: 0;
  flex: 1;
  border: 0;
  outline: 0;
  background: transparent;
  color: #fff;
  font-size: 26px;
  font-weight: 800;
}

.auth-field input::placeholder {
  color: #8f9098;
}

.google-code-field {
  display: grid;
  gap: 16px;
}

.google-code-field > span {
  color: #8f9098;
  font-size: 20px;
  font-weight: 800;
}

.google-code-boxes {
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  gap: 10px;
}

.google-code-boxes input {
  width: 100%;
  aspect-ratio: 1;
  min-width: 0;
  border: 0;
  border-radius: 18px;
  outline: 2px solid transparent;
  background: #20212b;
  color: #fff;
  font-size: 28px;
  font-weight: 900;
  text-align: center;
  transition: outline-color 0.18s ease, background 0.18s ease;
}

.google-code-boxes input:focus {
  outline-color: #00c313;
  background: #252733;
}

.field-icon {
  position: relative;
  width: 56px;
  height: 56px;
  flex: 0 0 auto;
  border-radius: 50%;
  background: #4a4c58;
}

.mail-icon::before {
  content: '';
  position: absolute;
  inset: 16px 13px 14px;
  border: 3px solid #fff;
}

.mail-icon::after {
  content: '';
  position: absolute;
  top: 17px;
  left: 15px;
  width: 26px;
  height: 18px;
  border-left: 3px solid #fff;
  border-bottom: 3px solid #fff;
  transform: rotate(-45deg) skew(12deg, 12deg);
}

.lock-icon::before {
  content: '';
  position: absolute;
  left: 15px;
  right: 15px;
  bottom: 14px;
  height: 22px;
  border: 3px solid #fff;
}

.lock-icon::after {
  content: '';
  position: absolute;
  top: 13px;
  left: 18px;
  width: 20px;
  height: 22px;
  border: 3px solid #fff;
  border-bottom: 0;
  border-radius: 12px 12px 0 0;
}

.phone-prefix {
  display: inline-flex;
  align-items: center;
  gap: 14px;
  color: #fff;
  font-size: 24px;
  font-weight: 800;
}

.phone-prefix i {
  width: 10px;
  height: 10px;
  border-right: 2px solid #a5a7af;
  border-bottom: 2px solid #a5a7af;
  transform: rotate(45deg);
}

.field-action {
  display: inline-flex;
  width: 42px;
  height: 42px;
  align-items: center;
  justify-content: center;
  border: 0;
  background: transparent;
  color: #9b9ca4;
}

.eye-off-icon {
  position: relative;
  width: 28px;
  height: 16px;
  border: 3px solid currentColor;
  border-radius: 50%;
}

.eye-off-icon::before {
  content: '';
  position: absolute;
  top: 4px;
  left: 8px;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
}

.eye-off-icon::after {
  content: '';
  position: absolute;
  top: -8px;
  left: 12px;
  width: 3px;
  height: 32px;
  background: currentColor;
  transform: rotate(-45deg);
}

.auth-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  min-width: 0;
  margin-top: -2px;
}

.remember-control {
  display: inline-flex;
  align-items: center;
  gap: 12px;
}

.remember-control input {
  position: absolute;
  opacity: 0;
}

.remember-control span {
  width: 30px;
  height: 30px;
  border: 2px solid #6d707d;
  border-radius: 8px;
}

.remember-control input:checked + span {
  border-color: #00c313;
  background: #00c313;
}

.remember-control em,
.auth-row a {
  color: #fff;
  font-size: 24px;
  font-style: normal;
  font-weight: 800;
  text-decoration: none;
}

.auth-error {
  margin: -12px 0 0;
  color: #ff6666;
  font-size: 16px;
  font-weight: 700;
}

.primary-button {
  min-height: 102px;
  margin-top: 64px;
  border: 0;
  border-radius: 50px;
  background: #00c313;
  color: #fff;
  font-size: 30px;
  font-weight: 900;
}

.primary-button:disabled {
  opacity: 0.7;
}

.quick-login {
  display: grid;
  justify-items: center;
  gap: 20px;
  margin-top: 60px;
  color: #fff;
}

.quick-login__divider {
  display: grid;
  width: 100%;
  grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
  align-items: center;
  gap: 18px;
}

.quick-login__divider span {
  height: 1px;
  background: #22242e;
}

.quick-login__divider strong,
.quick-login > strong {
  font-size: 26px;
  font-weight: 900;
}

.wallet-button {
  display: inline-flex;
  width: 108px;
  height: 108px;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 50%;
  background: #282a36;
  color: #00c313;
}

.wallet-button:disabled {
  opacity: 0.7;
}

.wallet-error {
  max-width: 100%;
  margin: -8px 0 0;
  color: #ff6b5f;
  font-size: 16px;
  font-weight: 700;
  text-align: center;
}

.wallet-icon {
  position: relative;
  width: 38px;
  height: 30px;
  border-radius: 4px;
  background: currentColor;
}

.wallet-icon::before {
  content: '';
  position: absolute;
  top: -9px;
  left: 9px;
  width: 28px;
  height: 18px;
  border-radius: 3px;
  background: currentColor;
  transform: rotate(-25deg);
}

.wallet-icon::after {
  content: '';
  position: absolute;
  top: 12px;
  right: 5px;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #282a36;
}

.auth-switch {
  margin: 46px 0 0;
  text-align: center;
  color: #8f9098;
  font-size: 24px;
  font-weight: 800;
}

.auth-switch button {
  border: 0;
  background: transparent;
  color: #00c313;
  font: inherit;
}

.wallet-sheet-layer {
  position: fixed;
  inset: 0;
  z-index: 80;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  background: rgba(0, 0, 0, 0.68);
  backdrop-filter: blur(8px);
}

.wallet-sheet {
  width: min(100%, 720px);
  max-height: min(78dvh, 560px);
  overflow-y: auto;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 28px 28px 0 0;
  background: #202020;
  padding: 28px 28px 24px;
  color: #fff;
  box-shadow: 0 -18px 40px rgba(0, 0, 0, 0.35);
}

.wallet-sheet__header {
  display: grid;
  grid-template-columns: 42px minmax(0, 1fr) 42px;
  align-items: center;
  margin-bottom: 26px;
}

.wallet-sheet__header strong {
  text-align: center;
  font-size: 24px;
  font-weight: 800;
}

.wallet-sheet__help,
.wallet-sheet__back,
.wallet-sheet__close {
  display: inline-flex;
  width: 36px;
  height: 36px;
  align-items: center;
  justify-content: center;
  border: 0;
  background: transparent;
  color: #fff;
  font-size: 28px;
  line-height: 1;
}

.wallet-sheet__back {
  font-size: 34px;
}

.wallet-sheet__back:disabled {
  opacity: 0.45;
}

.wallet-sheet__help {
  width: 28px;
  height: 28px;
  border: 2px solid currentColor;
  border-radius: 50%;
  font-size: 16px;
  font-weight: 800;
}

.wallet-list {
  display: grid;
  gap: 12px;
}

.wallet-pending {
  display: grid;
  justify-items: center;
  padding: 54px 16px 24px;
  text-align: center;
}

.wallet-pending__icon {
  position: relative;
  display: inline-flex;
  width: 92px;
  height: 92px;
  align-items: center;
  justify-content: center;
  border-radius: 24px;
  color: #fff;
  font-size: 38px;
  font-weight: 900;
}

.wallet-pending__icon::after {
  content: '';
  position: absolute;
  inset: -18px;
  border: 5px solid transparent;
  border-right-color: #2095ff;
  border-bottom-color: #2095ff;
  border-radius: 50%;
  animation: wallet-pending-spin 1.1s linear infinite;
}

@keyframes wallet-pending-spin {
  to {
    transform: rotate(360deg);
  }
}

.wallet-pending strong {
  margin-top: 54px;
  color: #fff;
  font-size: 26px;
  font-weight: 800;
}

.wallet-pending p {
  margin: 16px 0 38px;
  color: #9d9d9d;
  font-size: 22px;
  font-weight: 700;
}

.wallet-pending button {
  display: inline-flex;
  min-height: 58px;
  align-items: center;
  justify-content: center;
  gap: 14px;
  border: 1px solid rgba(255, 255, 255, 0.25);
  border-radius: 18px;
  background: transparent;
  padding: 0 30px;
  color: #fff;
  font-size: 22px;
  font-weight: 700;
}

.wallet-pending button:disabled {
  opacity: 0.55;
}

.wallet-pending button span {
  font-size: 22px;
}

.wallet-row,
.wallet-search {
  display: grid;
  width: 100%;
  min-width: 0;
  grid-template-columns: 54px minmax(0, 1fr) auto;
  align-items: center;
  gap: 18px;
  border: 0;
  border-radius: 16px;
  background: transparent;
  padding: 14px 18px;
  color: #fff;
  text-align: left;
}

.wallet-row:disabled {
  opacity: 0.62;
}

.wallet-row:not(:disabled):active {
  background: rgba(255, 255, 255, 0.05);
}

.wallet-row span:nth-child(2),
.wallet-search span:nth-child(2) {
  overflow: hidden;
  min-width: 0;
  font-size: 24px;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.wallet-row em,
.wallet-search em {
  border-radius: 8px;
  background: rgba(38, 169, 106, 0.18);
  padding: 5px 8px;
  color: #35c686;
  font-size: 16px;
  font-style: normal;
  font-weight: 700;
}

.wallet-row em.install {
  background: rgba(255, 255, 255, 0.08);
  color: #b8b8b8;
}

.wallet-row__icon,
.wallet-search__icon {
  display: inline-flex;
  width: 48px;
  height: 48px;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
  background: #2b2b2b;
  color: #fff;
  font-size: 22px;
  font-weight: 900;
}

.wallet-row__icon--brave {
  background: #f15a24;
}

.wallet-row__icon--metamask {
  background: #f6851b;
}

.wallet-row__icon--okx,
.wallet-row__icon--browser {
  background: #111;
}

.wallet-row__icon--coinbase {
  background: #2457ff;
}

.wallet-row__icon--tokenpocket {
  background: #2f7df6;
}

.wallet-row__icon--trust {
  background: #3375bb;
}

.wallet-row__icon--tronlink {
  background: #ff1f3d;
}

.wallet-row__icon--bitget {
  background: #00d1ff;
  color: #061018;
}

.wallet-row__icon--imtoken {
  background: #1f6fff;
}

.wallet-row__icon--math {
  background: #111b3a;
}

.wallet-row__icon--safepal {
  background: #5b6cff;
}

.wallet-row__icon--rabby {
  background: #7b5cff;
}

.wallet-row__icon--phantom {
  background: #ab9ff2;
  color: #171724;
}

.wallet-row__icon--binance {
  background: #f3ba2f;
  color: #171300;
}

.wallet-row__icon--onekey {
  background: #00b578;
}

.wallet-row__icon--gate {
  background: #2454ff;
}

.wallet-row__icon--coin98 {
  background: #d8a600;
}

.wallet-row__icon--exodus {
  background: #6c4cff;
}

.wallet-row__icon--opera {
  background: #ff1b2d;
}

.wallet-row__icon--frame,
.wallet-row__icon--tally {
  background: #f4f4f4;
  color: #111;
}

.wallet-row__icon--zerion {
  background: #2962ff;
}

.wallet-row__icon--rainbow {
  background: linear-gradient(135deg, #ff4d6d, #ffd166 45%, #22c55e 70%, #38bdf8);
}

.wallet-row__icon--bybit {
  background: #f7a600;
  color: #111;
}

.wallet-row__icon--kucoin {
  background: #24ae8f;
}

.wallet-row__icon--halo {
  background: #15d46f;
  color: #07140c;
}

.wallet-row__icon--subwallet {
  background: #004bff;
}

.wallet-row__icon--xdefi {
  background: #34d399;
  color: #06120d;
}

.wallet-row__icon--keplr,
.wallet-row__icon--cosmostation {
  background: #121826;
}

.wallet-row__icon--walletconnect {
  background: #3b99fc;
}

.wallet-search {
  margin-top: 6px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.04);
  opacity: 0.72;
}

.wallet-search__icon {
  position: relative;
  background: rgba(255, 255, 255, 0.04);
}

.wallet-search__icon::before {
  content: '';
  width: 14px;
  height: 14px;
  border: 3px solid #aaa;
  border-radius: 50%;
}

.wallet-search__icon::after {
  content: '';
  position: absolute;
  right: 13px;
  bottom: 13px;
  width: 11px;
  height: 3px;
  border-radius: 3px;
  background: #aaa;
  transform: rotate(45deg);
}

.wallet-sheet__error {
  margin: 16px 0 0;
  color: #ff6b5f;
  font-size: 15px;
  font-weight: 700;
  text-align: center;
}

.wallet-sheet__hint {
  margin: 18px 0 0;
  color: #9b9b9b;
  font-size: 14px;
  font-weight: 600;
  text-align: center;
}

@media (max-width: 520px) {
  .auth-page {
    padding: 20px 26px 36px;
  }

  .auth-content {
    padding-top: 92px;
  }

  .auth-content h1 {
    margin-bottom: 66px;
    font-size: 42px;
  }

  .auth-field,
  .primary-button {
    min-height: 84px;
  }

  .auth-field input,
  .auth-tabs button {
    font-size: 22px;
  }

  .remember-control em,
  .auth-row a,
  .auth-switch,
  .phone-prefix {
    font-size: 20px;
  }
}

.auth-page {
  padding: 18px 22px 30px;
}

.auth-topbar {
  margin: -18px -22px 0;
  padding: 18px 22px 10px;
}

.icon-button {
  width: 46px;
  height: 46px;
}

.chevron-left {
  width: 16px;
  height: 16px;
  border-left-width: 3px;
  border-bottom-width: 3px;
}

.globe-icon {
  width: 24px;
  height: 24px;
  border-width: 3px;
}

.auth-content {
  padding-top: 68px;
}

.auth-content h1 {
  margin-bottom: 48px;
  font-size: 38px;
}

.auth-tabs button {
  height: 46px;
  font-size: 21px;
}

.auth-form {
  gap: 22px;
  margin-top: 42px;
}

.auth-field {
  min-height: 74px;
  border-radius: 22px;
  padding: 0 18px;
}

.auth-field input {
  font-size: 20px;
}

.field-icon {
  width: 44px;
  height: 44px;
}

.phone-prefix,
.remember-control em,
.auth-row a,
.auth-switch {
  font-size: 18px;
}

.remember-control span {
  width: 26px;
  height: 26px;
}

.primary-button {
  min-height: 76px;
  margin-top: 36px;
  border-radius: 38px;
  font-size: 24px;
}

.quick-login {
  gap: 14px;
  margin-top: 38px;
}

.quick-login__divider strong,
.quick-login > strong {
  font-size: 20px;
}

.wallet-button {
  width: 84px;
  height: 84px;
}

.auth-switch {
  margin-top: 30px;
}

@media (max-width: 390px) {
  .auth-page {
    padding: 16px 18px 28px;
  }

  .auth-topbar {
    margin: -16px -18px 0;
    padding: 16px 18px 8px;
  }

  .auth-content {
    padding-top: 52px;
  }

  .auth-content h1 {
    margin-bottom: 38px;
    font-size: 32px;
  }

  .auth-form {
    gap: 18px;
    margin-top: 32px;
  }

  .auth-field {
    min-height: 66px;
    border-radius: 18px;
    padding: 0 14px;
  }

  .auth-field input,
  .auth-tabs button {
    font-size: 18px;
  }

  .remember-control em,
  .auth-row a,
  .auth-switch,
  .phone-prefix {
    font-size: 16px;
  }

  .primary-button {
    min-height: 68px;
    margin-top: 26px;
    font-size: 21px;
  }

  .quick-login {
    margin-top: 30px;
  }
}

@media (max-width: 959px) {
  .auth-page {
    padding: 14px 24px 28px;
  }

  .auth-topbar {
    margin: -14px -24px 0;
    padding: 14px 24px 8px;
  }

  .icon-button {
    width: 42px;
    height: 42px;
  }

  .chevron-left {
    width: 15px;
    height: 15px;
    border-left-width: 3px;
    border-bottom-width: 3px;
  }

  .globe-icon {
    width: 22px;
    height: 22px;
    border-width: 3px;
  }

  .auth-content {
    padding-top: 42px;
  }

  .auth-content h1 {
    margin-bottom: 42px;
    font-size: 34px;
  }

  .auth-tabs button {
    height: 40px;
    font-size: 19px;
  }

  .auth-tabs button.active::after {
    height: 3px;
  }

  .auth-form {
    gap: 18px;
    margin-top: 34px;
  }

  .auth-field {
    min-height: 62px;
    border-radius: 18px;
    padding: 0 14px;
  }

  .auth-field input {
    font-size: 18px;
  }

  .field-icon {
    width: 38px;
    height: 38px;
  }

  .field-action {
    width: 34px;
    height: 34px;
  }

  .eye-off-icon {
    width: 24px;
    height: 14px;
    border-width: 3px;
  }

  .phone-prefix,
  .remember-control em,
  .auth-row a,
  .auth-switch {
    font-size: 15px;
  }

  .remember-control span {
    width: 24px;
    height: 24px;
    border-radius: 7px;
  }

  .primary-button {
    min-height: 66px;
    margin-top: 28px;
    border-radius: 33px;
    font-size: 22px;
  }

  .quick-login {
    gap: 12px;
    margin-top: 34px;
  }

  .quick-login__divider strong,
  .quick-login > strong {
    font-size: 18px;
  }

  .wallet-button {
    width: 72px;
    height: 72px;
  }

  .wallet-sheet {
    max-height: 70dvh;
    border-radius: 24px 24px 0 0;
    padding: 22px 20px 20px;
  }

  .wallet-sheet__header {
    grid-template-columns: 34px minmax(0, 1fr) 34px;
    margin-bottom: 18px;
  }

  .wallet-sheet__header strong {
    font-size: 20px;
  }

  .wallet-pending {
    padding: 42px 10px 18px;
  }

  .wallet-pending__icon {
    width: 78px;
    height: 78px;
    border-radius: 20px;
    font-size: 30px;
  }

  .wallet-pending strong {
    margin-top: 44px;
    font-size: 21px;
  }

  .wallet-pending p {
    margin: 12px 0 30px;
    font-size: 17px;
  }

  .wallet-pending button {
    min-height: 52px;
    border-radius: 16px;
    font-size: 18px;
  }

  .wallet-row,
  .wallet-search {
    grid-template-columns: 44px minmax(0, 1fr) auto;
    gap: 12px;
    padding: 11px 8px;
  }

  .wallet-row__icon,
  .wallet-search__icon {
    width: 42px;
    height: 42px;
    border-radius: 10px;
    font-size: 18px;
  }

  .wallet-row span:nth-child(2),
  .wallet-search span:nth-child(2) {
    font-size: 18px;
  }

  .wallet-row em,
  .wallet-search em {
    font-size: 12px;
  }
}
</style>
