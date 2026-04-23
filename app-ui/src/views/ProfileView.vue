<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import { getAccessToken } from '@/api/http'
import { apiGetProfile, apiLogout } from '@/api/userPrivate'
import { apiGuestLogin } from '@/api/userPublic'
import { useTenantStore } from '@/stores/tenant'
import type { UserProfile } from '@/types/user'

const guestMenuItems = [
  { key: 'language', label: '语言', value: '中文简体', icon: '◎', flag: '🇨🇳' },
  { key: 'service', label: '客服', icon: '♧' },
  { key: 'download', label: 'APP下载', icon: '⇩' },
  { key: 'whitepaper', label: '白皮书', icon: 'i' },
  { key: 'company', label: '公司资质', icon: 'i' },
  { key: 'regulation', label: '监管文件', icon: 'i' },
]

const userMenuItems = [
  { key: 'language', label: '语言', value: '中文简体', icon: '◎', flag: '🇨🇳' },
  { key: 'bank', label: '收款账号', icon: '▭' },
  { key: 'security', label: '安全', icon: '盾' },
  { key: 'service', label: '客服', icon: '♧' },
  { key: 'download', label: 'APP下载', icon: '⇩' },
  { key: 'whitepaper', label: '白皮书', icon: 'i' },
  { key: 'company', label: '公司资质', icon: 'i' },
  { key: 'regulation', label: '监管文件', icon: 'i' },
  { key: 'logout', label: '退出登录', icon: '↪' },
]

const profile = ref<UserProfile | null>(null)
const isLoggedIn = ref(Boolean(getAccessToken()))
const loadingProfile = ref(false)
const loggingGuest = ref(false)
const guestLoginError = ref('')
const tenantStore = useTenantStore()

const menuItems = computed(() => (isLoggedIn.value ? userMenuItems : guestMenuItems))
const userBase = computed(() => profile.value?.base ?? null)
const displayName = computed(() => userBase.value?.nickname || userBase.value?.username || 'GUEST-6721')
const displayId = computed(() => userBase.value?.userNo || (userBase.value?.id ? String(userBase.value.id) : '42110599'))
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
    guestLoginError.value = '租户信息缺失'
    return
  }

  loggingGuest.value = true
  try {
    const res = await apiGuestLogin({ tenantCode: tenantStore.tenantCode })
    if (res.code !== 0 && res.code !== 200) {
      guestLoginError.value = res.msg || '登录失败'
      return
    }
    isLoggedIn.value = true
    await loadProfile()
  } catch (error) {
    console.warn('guest login failed', error)
    guestLoginError.value = '登录失败'
  } finally {
    loggingGuest.value = false
  }
}

async function handleMenuClick(key: string) {
  if (key !== 'logout') return

  try {
    await apiLogout()
  } finally {
    profile.value = null
    isLoggedIn.value = false
  }
}
</script>

<template>
  <section class="profile-page">
    <header class="profile-header">
      <h1>用户中心</h1>
    </header>

    <section v-if="isLoggedIn" class="profile-user" :aria-busy="loadingProfile">
      <div class="profile-avatar" :class="{ 'profile-avatar--image': avatarUrl }">
        <img v-if="avatarUrl" :src="avatarUrl" alt="" />
        <span v-else />
      </div>
      <div class="profile-user__info">
        <h2>{{ displayName }}</h2>
        <p>ID:{{ displayId }}</p>
      </div>
    </section>

    <section v-else class="profile-hero" aria-label="账户">
      <div class="profile-hero__intro">
        <div>
          <h2>欢迎使用本平台</h2>
          <p>安全可靠，极速体验</p>
        </div>
        <button
          type="button"
          class="profile-demo"
          :disabled="loggingGuest"
          :aria-busy="loggingGuest"
          @click="handleGuestLogin"
        >
          <span>{{ loggingGuest ? '登录中' : '模拟用户登录' }}</span>
          <i />
        </button>
      </div>
      <p v-if="guestLoginError" class="profile-hero__error">{{ guestLoginError }}</p>

      <div class="profile-actions">
        <button type="button" class="profile-actions__login">登录</button>
        <button type="button" class="profile-actions__register">注册</button>
      </div>
    </section>

    <nav class="profile-menu" aria-label="用户中心菜单">
      <button
        v-for="item in menuItems"
        :key="item.key"
        type="button"
        class="profile-menu__item"
        @click="handleMenuClick(item.key)"
      >
        <span
          class="profile-menu__icon"
          :class="{
            'profile-menu__icon--info': item.icon === 'i',
            'profile-menu__icon--shield': item.icon === '盾',
            'profile-menu__icon--bank': item.icon === '▭',
          }"
        >
          {{ item.icon }}
        </span>
        <span class="profile-menu__label">{{ item.label }}</span>
        <span v-if="item.flag" class="profile-menu__flag">{{ item.flag }}</span>
        <span v-if="item.value" class="profile-menu__value">{{ item.value }}</span>
        <i class="profile-menu__arrow" />
      </button>
    </nav>
  </section>
</template>

<style scoped>
.profile-page {
  width: 100%;
  max-width: 760px;
  min-height: 100%;
  margin: 0 auto;
  padding: 22px 28px calc(118px + env(safe-area-inset-bottom));
  overflow-x: hidden;
  background: #0b0c15;
  color: #f7f8fb;
}

.profile-header {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 54px;
  margin-bottom: 42px;
}

.profile-header h1 {
  margin: 0;
  font-size: 26px;
  font-weight: 500;
  letter-spacing: 0;
}

.profile-hero {
  display: grid;
  gap: 34px;
  margin-bottom: 60px;
}

.profile-user {
  display: flex;
  align-items: center;
  gap: 24px;
  min-height: 96px;
  margin: 22px 0 54px;
}

.profile-avatar {
  display: grid;
  place-items: center;
  flex: none;
  width: 74px;
  height: 74px;
  overflow: hidden;
  border-radius: 999px;
  background:
    radial-gradient(circle at 50% 34%, #f8ecef 0 25%, transparent 26%),
    radial-gradient(circle at 34% 20%, #14151d 0 16%, transparent 17%),
    radial-gradient(circle at 64% 20%, #14151d 0 16%, transparent 17%),
    radial-gradient(circle at 50% 78%, #43c35d 0 34%, transparent 35%),
    #9ac291;
  border: 6px solid #7ea97b;
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
  margin: 0 0 12px;
  overflow: hidden;
  color: #fff;
  font-size: 26px;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.profile-user__info p {
  margin: 0;
  color: #9396a1;
  font-size: 19px;
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
  font-size: 28px;
  font-weight: 700;
  line-height: 1.15;
}

.profile-hero p {
  margin: 0;
  color: #9b9daa;
  font-size: 20px;
  font-weight: 700;
}

.profile-hero__error {
  margin: -18px 0 0;
  color: #ff6b6b;
  font-size: 16px;
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
  gap: 14px;
  flex: none;
  color: #04c704;
  font-size: 19px;
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

.profile-actions {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: 26px;
}

.profile-actions button {
  min-height: 78px;
  border-radius: 999px;
  font-size: 25px;
  font-weight: 700;
}

.profile-actions__login {
  background: #242631;
}

.profile-actions__register {
  background: #02b904;
}

.profile-menu {
  display: grid;
  gap: 20px;
}

.profile-menu__item {
  display: grid;
  grid-template-columns: 48px minmax(0, 1fr) auto auto 18px;
  align-items: center;
  gap: 16px;
  min-height: 86px;
  padding: 0 28px;
  border-radius: 28px;
  background: #1b1d27;
  text-align: left;
}

.profile-menu__icon {
  display: grid;
  place-items: center;
  width: 34px;
  height: 34px;
  color: #fff;
  font-size: 28px;
  line-height: 1;
}

.profile-menu__icon--info {
  border: 3px solid currentColor;
  border-radius: 999px;
  font-size: 22px;
  font-weight: 700;
  font-family: Georgia, serif;
}

.profile-menu__icon--bank {
  position: relative;
  color: transparent;
}

.profile-menu__icon--bank::before,
.profile-menu__icon--bank::after {
  position: absolute;
  content: '';
}

.profile-menu__icon--bank::before {
  inset: 5px 2px 7px;
  border: 2px solid #fff;
  border-radius: 2px;
}

.profile-menu__icon--bank::after {
  left: 7px;
  right: 7px;
  bottom: 12px;
  height: 2px;
  background: #fff;
  box-shadow: 0 7px 0 #fff;
}

.profile-menu__icon--shield {
  width: 34px;
  height: 34px;
  color: transparent;
  background:
    linear-gradient(135deg, transparent 42%, #fff 43% 52%, transparent 53%) center / 18px 18px no-repeat;
  clip-path: polygon(50% 4%, 86% 18%, 86% 52%, 50% 96%, 14% 52%, 14% 18%);
  outline: 3px solid #fff;
  outline-offset: -6px;
}

.profile-menu__label {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 23px;
  font-weight: 500;
}

.profile-menu__flag {
  font-size: 30px;
  line-height: 1;
}

.profile-menu__value {
  color: #f7f8fb;
  font-size: 21px;
  white-space: nowrap;
}

@media (max-width: 959px) {
  .profile-page {
    max-width: none;
  }
}

@media (max-width: 520px) {
  .profile-page {
    padding: 18px 16px calc(112px + env(safe-area-inset-bottom));
  }

  .profile-header {
    margin-bottom: 38px;
  }

  .profile-hero {
    gap: 30px;
    margin-bottom: 52px;
  }

  .profile-user {
    gap: 18px;
    margin: 16px 0 46px;
  }

  .profile-avatar {
    width: 64px;
    height: 64px;
    border-width: 5px;
  }

  .profile-user__info h2 {
    margin-bottom: 8px;
    font-size: 22px;
  }

  .profile-user__info p {
    font-size: 17px;
  }

  .profile-hero__intro {
    align-items: flex-start;
  }

  .profile-hero h2 {
    font-size: 24px;
  }

  .profile-hero p {
    font-size: 18px;
  }

  .profile-demo {
    gap: 10px;
    font-size: 17px;
  }

  .profile-actions {
    gap: 16px;
  }

  .profile-actions button {
    min-height: 60px;
    font-size: 21px;
  }

  .profile-menu {
    gap: 14px;
  }

  .profile-menu__item {
    grid-template-columns: 38px minmax(0, 1fr) auto auto 14px;
    gap: 10px;
    min-height: 72px;
    padding: 0 18px;
    border-radius: 22px;
  }

  .profile-menu__icon {
    width: 30px;
    height: 30px;
    font-size: 24px;
  }

  .profile-menu__icon--info {
    border-width: 2px;
    font-size: 19px;
  }

  .profile-menu__icon--shield {
    width: 30px;
    height: 30px;
  }

  .profile-menu__label {
    font-size: 20px;
  }

  .profile-menu__flag {
    font-size: 24px;
  }

  .profile-menu__value {
    font-size: 18px;
  }
}
</style>
