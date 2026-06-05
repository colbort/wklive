<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { apiGetProfile } from '@/api/userPrivate'
import AppIcon from '@/components/common/AppIcon.vue'
import { useI18n } from '@/i18n'
import type { UserProfile } from '@/types/user'

const router = useRouter()
const { t } = useI18n()
const profile = ref<UserProfile | null>(null)

const phoneLabel = computed(() => {
  const phone = profile.value?.identity?.phone || ''
  if (!phone) return t('security.phoneBind')
  return `${t('security.phoneBind')} ${phone.slice(0, 3)}****${phone.slice(-4)}`
})

const emailLabel = computed(() => {
  const email = profile.value?.identity?.email || ''
  if (!email) return t('security.emailBind')
  const [name, domain] = email.split('@')
  if (!name || !domain) return t('security.editEmail')
  const visibleName = name.length <= 2 ? `${name[0] || ''}*` : `${name.slice(0, 2)}***`
  return `${t('security.editEmail')} ${visibleName}@${domain}`
})

const rows = computed(() => [
  { key: 'login-password', label: t('security.editLoginPassword') },
  { key: 'pay-password', label: t('security.editPayPassword') },
  { key: 'bind-phone', label: phoneLabel.value },
  { key: 'bind-email', label: emailLabel.value },
])

onMounted(loadProfile)

async function loadProfile() {
  try {
    const res = await apiGetProfile()
    if (res.code === 200) profile.value = res.data
  } catch (error) {
    console.warn('load profile failed', error)
  }
}

function handleRowClick(key: string) {
  if (key === 'login-password') router.push('/profile/security/login-password')
  if (key === 'pay-password') router.push('/profile/security/pay-password')
  if (key === 'bind-phone') router.push('/profile/security/bind-phone')
  if (key === 'bind-email') router.push('/profile/security/bind-email')
}
</script>

<template>
  <section class="security-page">
    <header class="security-header">
      <button
        type="button"
        class="back-button"
        :aria-label="t('common.back')"
        @click="router.back()"
      >
        <AppIcon name="back" class="back-icon-svg" />
      </button>
      <h1>{{ t('security.title') }}</h1>
    </header>

    <nav class="security-list" :aria-label="t('security.title')">
      <button
        v-for="row in rows"
        :key="row.key"
        type="button"
        class="security-row"
        @click="handleRowClick(row.key)"
      >
        <span>{{ row.label }}</span>
        <i />
      </button>
    </nav>
  </section>
</template>

<style scoped>
.security-page {
  width: 100%;
  max-width: 680px;
  min-height: 100dvh;
  margin: 0 auto;
  padding: 18px 36px;
  overflow-x: hidden;
  background: #0b0c15;
  color: #fff;
}

.security-header {
  position: relative;
  display: grid;
  grid-template-columns: 48px minmax(0, 1fr) 48px;
  align-items: center;
  min-height: 48px;
  margin-bottom: 52px;
}

.security-header h1 {
  margin: 0;
  text-align: center;
  font-size: 24px;
  font-weight: 800;
  letter-spacing: 0;
}

.back-button {
  display: inline-flex;
  width: 44px;
  height: 44px;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 999px;
  background: #242631;
  color: #fff;
}

.back-icon-svg {
  width: 24px;
  height: 24px;
  transform: translateX(-1px);
}

.security-list {
  display: grid;
  gap: 14px;
}

.security-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 16px;
  align-items: center;
  min-height: 70px;
  padding: 0 24px;
  border: 0;
  border-radius: 22px;
  background: #1b1d27;
  color: #fff;
  text-align: left;
}

.security-row span {
  min-width: 0;
  overflow: hidden;
  font-size: 19px;
  font-weight: 800;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.security-row i {
  width: 12px;
  height: 12px;
  border-right: 2px solid currentColor;
  border-bottom: 2px solid currentColor;
  opacity: 0.9;
  transform: rotate(-45deg);
}

@media (max-width: 520px) {
  .security-page {
    padding: 16px 28px;
  }

  .security-header {
    grid-template-columns: 42px minmax(0, 1fr) 42px;
    min-height: 42px;
    margin-bottom: 48px;
  }

  .security-header h1 {
    font-size: 22px;
  }

  .back-button {
    width: 40px;
    height: 40px;
  }

  .security-list {
    gap: 14px;
  }

  .security-row {
    min-height: 64px;
    padding: 0 22px;
    border-radius: 20px;
  }

  .security-row span {
    font-size: 17px;
  }
}
</style>
