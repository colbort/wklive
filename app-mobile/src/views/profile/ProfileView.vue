<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { getAccessToken } from '@/api/http'
import { apiGetProfile, apiLogout } from '@/api/userPrivate'
import { apiGuestLogin } from '@/api/userPublic'
import AppIcon from '@/components/common/AppIcon.vue'
import { useI18n } from '@/i18n'
import { useTenantStore } from '@/stores/tenant'
import type { UserProfile } from '@/types/user'

type ProfileMenuIcon =
  | 'globe'
  | 'id-card'
  | 'credit-card'
  | 'shield-check'
  | 'headset'
  | 'cloud-download'
  | 'info-circle'
  | 'user-plus'
  | 'logout'

interface ProfileMenuItem {
  key: string
  labelKey: string
  icon: ProfileMenuIcon
}

const guestMenuItems: ProfileMenuItem[] = [
  { key: 'language', labelKey: 'common.language', icon: 'globe' },
  { key: 'service', labelKey: 'userMenu.customerService', icon: 'headset' },
  { key: 'download', labelKey: 'userMenu.appDownload', icon: 'cloud-download' },
  { key: 'whitepaper', labelKey: 'nav.whitepaper', icon: 'info-circle' },
  { key: 'company', labelKey: 'nav.license', icon: 'info-circle' },
  { key: 'regulation', labelKey: 'nav.compliance', icon: 'info-circle' },
]

const userMenuItems: ProfileMenuItem[] = [
  { key: 'language', labelKey: 'common.language', icon: 'globe' },
  { key: 'identity', labelKey: 'userMenu.identity', icon: 'id-card' },
  { key: 'bank', labelKey: 'userMenu.paymentAccount', icon: 'credit-card' },
  { key: 'security', labelKey: 'userMenu.security', icon: 'shield-check' },
  { key: 'invite', labelKey: 'userMenu.invite', icon: 'user-plus' },
  { key: 'service', labelKey: 'userMenu.customerService', icon: 'headset' },
  { key: 'download', labelKey: 'userMenu.appDownload', icon: 'cloud-download' },
  { key: 'whitepaper', labelKey: 'nav.whitepaper', icon: 'info-circle' },
  { key: 'company', labelKey: 'nav.license', icon: 'info-circle' },
  { key: 'regulation', labelKey: 'nav.compliance', icon: 'info-circle' },
  { key: 'logout', labelKey: 'common.logout', icon: 'logout' },
  { key: 'test1', labelKey: 'userMenu.test1', icon: 'info-circle' },
  { key: 'test2', labelKey: 'userMenu.test2', icon: 'info-circle' },
  { key: 'test3', labelKey: 'userMenu.test3', icon: 'info-circle' },
  { key: 'test4', labelKey: 'userMenu.test4', icon: 'info-circle' },
]

const profile = ref<UserProfile | null>(null)
const isLoggedIn = ref(Boolean(getAccessToken()))
const loadingProfile = ref(false)
const loggingGuest = ref(false)
const guestLoginError = ref('')
const tenantStore = useTenantStore()
const router = useRouter()
const { locale, t } = useI18n()

const userBase = computed(() => profile.value?.user ?? null)
const menuItems = computed<ProfileMenuItem[]>(() => {
  if (!isLoggedIn.value) return guestMenuItems
  return userMenuItems
})
const languageName = computed(() =>
  locale.value === 'zh-CN' ? t('common.zhCN') : t('common.enUS'),
)
const languageFlag = computed(() => (locale.value === 'zh-CN' ? '🇨🇳' : '🇺🇸'))
const displayName = computed(
  () => userBase.value?.nickname || userBase.value?.username || 'GUEST-6721',
)
const displayId = computed(
  () => userBase.value?.userNo || (userBase.value?.id ? String(userBase.value.id) : '42110599'),
)
const avatarUrl = computed(() => userBase.value?.avatar || '')

onMounted(() => {
  if (isLoggedIn.value) {
    loadProfile()
  }
})

async function loadProfile() {
  loadingProfile.value = true
  try {
    const res = await apiGetProfile()
    profile.value = res.data
  } catch (error) {
    console.warn('load profile failed', error)
  } finally {
    loadingProfile.value = false
  }
}

async function handleGuestLogin() {
  if (loggingGuest.value) return

  guestLoginError.value = ''
  tenantStore.hydrateFromEnv()
  if (!tenantStore.tenantCode) {
    guestLoginError.value = t('profile.tenantMissing')
    return
  }

  loggingGuest.value = true
  try {
    const res = await apiGuestLogin({ tenantCode: tenantStore.tenantCode })
    if (res.code !== 200) {
      guestLoginError.value = res.msg || t('profile.loginFailed')
      return
    }
    isLoggedIn.value = true
    await loadProfile()
  } catch (error) {
    console.warn('guest login failed', error)
    guestLoginError.value = t('profile.loginFailed')
  } finally {
    loggingGuest.value = false
  }
}

async function handleMenuClick(key: string) {
  if (key === 'security') {
    router.push('/profile/security')
    return
  }

  if (key === 'language') {
    router.push('/language')
    return
  }

  if (key === 'test1') {
    router.push('/test1')
    return
  }

  if (key === 'test2') {
    router.push('/test2')
    return
  }

  if (key === 'test3') {
    router.push('/test3')
    return
  }

  if (key === 'test4') {
    router.push('/test4')
    return
  }

  if (key !== 'logout') return

  try {
    await apiLogout()
  } finally {
    profile.value = null
    isLoggedIn.value = false
  }
}

function goLogin() {
  router.push('/login')
}
</script>

<template>
  <section class="profile-page">
    <header class="profile-header">
      <h1>{{ t('profile.title') }}</h1>
    </header>

    <section v-if="isLoggedIn" class="profile-user" :aria-busy="loadingProfile">
      <div class="profile-avatar" :class="{ 'profile-avatar--image': avatarUrl }">
        <img v-if="avatarUrl" :src="avatarUrl" alt="">
        <span v-else />
      </div>
      <div class="profile-user__info">
        <h2>{{ displayName }}</h2>
        <p>ID:{{ displayId }}</p>
      </div>
    </section>

    <section v-else class="profile-hero" :aria-label="t('profile.account')">
      <div class="profile-hero__intro">
        <div>
          <h2>{{ t('profile.welcome') }}</h2>
          <p>{{ t('profile.intro') }}</p>
        </div>
        <button
          type="button"
          class="profile-demo"
          :disabled="loggingGuest"
          :aria-busy="loggingGuest"
          @click="handleGuestLogin"
        >
          <span>{{ loggingGuest ? t('profile.loggingIn') : t('profile.guestLogin') }}</span>
          <i />
        </button>
      </div>
      <p v-if="guestLoginError" class="profile-hero__error">
        {{ guestLoginError }}
      </p>

      <div class="profile-actions">
        <button type="button" class="profile-actions__login" @click="goLogin">
          {{ t('common.login') }}
        </button>
        <button type="button" class="profile-actions__register" @click="router.push('/register')">
          {{ t('common.register') }}
        </button>
      </div>
    </section>

    <nav class="profile-menu" :aria-label="t('profile.menuLabel')">
      <button
        v-for="item in menuItems"
        :key="item.key"
        type="button"
        class="profile-menu__item"
        @click="handleMenuClick(item.key)"
      >
        <span class="profile-menu__icon">
          <AppIcon :name="item.icon" class="profile-menu__icon-svg" />
        </span>
        <span class="profile-menu__label">{{ t(item.labelKey) }}</span>
        <span v-if="item.key === 'language'" class="profile-menu__flag">{{ languageFlag }}</span>
        <span v-if="item.key === 'language'" class="profile-menu__value">{{ languageName }}</span>
        <i class="profile-menu__arrow" />
      </button>
    </nav>
  </section>
</template>

<style scoped>
.profile-page {
  width: 100%;
  max-width: 100%;
  min-height: 100%;
  margin: 0 auto;
  padding: 10px 24px calc(88px + env(safe-area-inset-bottom));
  overflow-x: hidden;
  background: var(--page-bg);
  color: #f7f8fb;
}

.profile-header {
  position: sticky;
  top: 0;
  z-index: 25;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 36px;
  margin: -10px -24px 20px;
  padding: 10px 24px 6px;
  background: var(--page-bg);
}

.profile-header h1 {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 500;
  letter-spacing: 0;
}

.profile-hero {
  display: grid;
  gap: 24px;
  margin-bottom: 38px;
}

.profile-user {
  display: flex;
  align-items: center;
  gap: 16px;
  min-height: 56px;
  margin: 2px 6px 26px;
}

.profile-avatar {
  display: grid;
  place-items: center;
  flex: none;
  width: 54px;
  height: 54px;
  overflow: hidden;
  border-radius: 999px;
  background:
    radial-gradient(circle at 50% 34%, #f8ecef 0 25%, transparent 26%),
    radial-gradient(circle at 34% 20%, #14151d 0 16%, transparent 17%),
    radial-gradient(circle at 64% 20%, #14151d 0 16%, transparent 17%),
    radial-gradient(circle at 50% 78%, #43c35d 0 34%, transparent 35%), #9ac291;
  border: 4px solid #7ea97b;
}

.profile-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.profile-avatar--image {
  background: #1b1d27;
  border-color: rgba(255, 255, 255, 0.1);
}

.profile-user__info {
  min-width: 0;
}

.profile-user__info h2 {
  margin: 0 0 5px;
  overflow: hidden;
  color: var(--text);
  font-size: 0.9rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.profile-user__info p {
  margin: 0;
  color: #9396a1;
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.04em;
}

.profile-hero__intro {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
}

.profile-hero h2 {
  margin: 0 0 12px;
  font-size: 1.2rem;
  font-weight: 700;
  line-height: 1.15;
}

.profile-hero p {
  margin: 0;
  color: #9b9daa;
  font-size: 0.85rem;
  font-weight: 700;
}

.profile-hero__error {
  margin: -18px 0 0;
  color: var(--danger);
  font-size: 0.8rem;
  font-weight: 600;
}

.profile-demo,
.profile-actions button,
.profile-menu__item {
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
}

.profile-demo {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  flex: none;
  color: #04c704;
  font-size: 0.8rem;
  font-weight: 500;
}

.profile-demo:disabled {
  cursor: wait;
  opacity: 0.72;
}

.profile-demo i,
.profile-menu__arrow {
  width: 14px;
  height: 14px;
  transform: rotate(-45deg);
  border-right: 2px solid #8b8e99;
  border-bottom: 2px solid #8b8e99;
}

.profile-menu__arrow {
  grid-column: 5;
  justify-self: end;
}

.profile-actions {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: 12px;
}

.profile-actions > button {
  min-height: 52px;
  border-radius: 999px;
  color: #fff;
  font-size: 0.9rem;
  font-weight: 600;
}

.profile-actions > .profile-actions__login {
  background: #272832;
}

.profile-actions > .profile-actions__register {
  background: #04b900;
}

.profile-menu {
  display: grid;
  gap: 10px;
}

.profile-menu__item {
  display: grid;
  grid-template-columns: 30px minmax(0, 1fr) 24px minmax(0, 76px) 14px;
  align-items: center;
  gap: 10px;
  min-height: 54px;
  padding: 0 18px;
  border-radius: 17px;
  background: #1b1d27;
  text-align: left;
}

.profile-menu__icon {
  grid-column: 1;
  display: inline-flex;
  width: 28px;
  height: 28px;
  align-items: center;
  justify-content: center;
  color: var(--text);
  line-height: 1;
}

.profile-menu__icon-svg {
  width: 25px;
  height: 25px;
}

.profile-menu__label {
  grid-column: 2;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.85rem;
  font-weight: 500;
}

.profile-menu__flag {
  grid-column: 3;
  display: inline-flex;
  justify-content: center;
  font-size: 1.1rem;
  line-height: 1;
}

.profile-menu__value {
  grid-column: 4;
  overflow: hidden;
  color: #f7f8fb;
  font-size: 0.8rem;
  text-align: right;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 520px) {
  .profile-page {
    padding: 10px 16px calc(88px + env(safe-area-inset-bottom));
  }

  .profile-header {
    margin: -10px -16px 18px;
    padding: 10px 16px 6px;
  }

  .profile-header h1 {
    font-size: 1.05rem;
  }

  .profile-hero {
    gap: 24px;
    margin-bottom: 38px;
  }

  .profile-user {
    gap: 14px;
    margin: 2px 6px 24px;
  }

  .profile-avatar {
    width: 50px;
    height: 50px;
    border-width: 4px;
  }

  .profile-user__info h2 {
    margin-bottom: 5px;
    font-size: 0.85rem;
  }

  .profile-user__info p {
    font-size: 0.65rem;
  }

  .profile-hero__intro {
    align-items: flex-start;
  }

  .profile-hero h2 {
    font-size: 1.1rem;
  }

  .profile-hero p {
    font-size: 0.8rem;
  }

  .profile-demo {
    gap: 8px;
    font-size: 0.75rem;
  }

  .profile-actions {
    gap: 12px;
  }

  .profile-menu {
    gap: 10px;
  }

  .profile-menu__item {
    grid-template-columns: 28px minmax(0, 1fr) 22px minmax(0, 72px) 14px;
    gap: 9px;
    min-height: 52px;
    padding: 0 16px;
    border-radius: 16px;
  }

  .profile-menu__icon {
    width: 23px;
    height: 23px;
  }

  .profile-menu__icon-svg {
    width: 23px;
    height: 23px;
  }

  .profile-menu__label {
    font-size: 0.8rem;
  }

  .profile-menu__flag {
    font-size: 1.2rem;
  }

  .profile-menu__value {
    font-size: 0.9rem;
  }
}
</style>
