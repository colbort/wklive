<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { apiGetProfile, apiUpdateIdentity } from '@/api/userPrivate'
import AppIcon from '@/components/common/AppIcon.vue'
import { useI18n } from '@/i18n'
import type { UserIdentity } from '@/types/user'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const account = ref('')
const identity = ref<UserIdentity | null>(null)
const submitting = ref(false)
const message = ref('')
const errorMessage = ref('')

const isEmailMode = computed(() => route.name === 'security-bind-email')
const accountTypeName = computed(() =>
  isEmailMode.value ? t('security.email') : t('security.phone'),
)
const pageTitle = computed(() => {
  const current = isEmailMode.value ? identity.value?.email : identity.value?.phone
  if (isEmailMode.value) return current ? t('security.editEmail') : t('security.bindEmail')
  return current ? t('security.editPhone') : t('security.bindPhone')
})
const inputMode = computed(() => (isEmailMode.value ? 'email' : 'tel'))
const autocomplete = computed(() => (isEmailMode.value ? 'email' : 'tel'))
const placeholder = computed(() =>
  isEmailMode.value ? t('security.inputEmail') : t('security.inputPhone'),
)
const submitText = computed(() => {
  if (submitting.value) return t('common.submitting')
  return identity.value && (isEmailMode.value ? identity.value.email : identity.value.phone)
    ? t('security.confirmEdit')
    : t('security.confirmBind')
})

onMounted(loadIdentity)

async function loadIdentity() {
  try {
    const res = await apiGetProfile()
    if (res.code === 200) {
      identity.value = res.data?.identity || null
      account.value = isEmailMode.value ? identity.value?.email || '' : identity.value?.phone || ''
    }
  } catch (error) {
    console.warn('load profile failed', error)
  }
}

function validateAccount(value: string) {
  if (!value) return isEmailMode.value ? t('security.inputEmail') : t('security.inputPhone')
  if (isEmailMode.value) {
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) return t('security.invalidEmail')
    return ''
  }
  if (!/^\+?\d{6,20}$/.test(value)) return t('security.invalidPhone')
  return ''
}

async function submitBindAccount() {
  if (submitting.value) return

  message.value = ''
  errorMessage.value = ''
  const normalizedAccount = account.value.trim()
  const validateMessage = validateAccount(normalizedAccount)
  if (validateMessage) {
    errorMessage.value = validateMessage
    return
  }

  submitting.value = true
  try {
    const res = await apiUpdateIdentity(
      isEmailMode.value ? { email: normalizedAccount } : { phone: normalizedAccount },
    )
    if (res.code !== 200) {
      errorMessage.value = res.msg || t('security.submitFailed')
      return
    }
    message.value = t('security.submitSuccess')
    window.setTimeout(() => router.replace('/profile/security'), 600)
  } catch (error) {
    console.warn('bind account failed', error)
    errorMessage.value = t('security.submitFailed')
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <section class="bind-account-page">
    <header class="bind-account-header">
      <button
        type="button"
        class="back-button"
        :aria-label="t('common.back')"
        @click="router.back()"
      >
        <AppIcon name="back" class="back-icon-svg" />
      </button>
      <h1>{{ pageTitle }}</h1>
    </header>

    <form class="bind-account-form" @submit.prevent="submitBindAccount">
      <label class="bind-account-field">
        <span>{{ accountTypeName }}</span>
        <input
          v-model="account"
          :inputmode="inputMode"
          :autocomplete="autocomplete"
          :placeholder="placeholder"
        >
      </label>

      <p v-if="errorMessage" class="form-message form-message--error">
        {{ errorMessage }}
      </p>
      <p v-if="message" class="form-message">
        {{ message }}
      </p>

      <button type="submit" class="submit-button" :disabled="submitting">
        {{ submitText }}
      </button>
    </form>
  </section>
</template>

<style scoped>
.bind-account-page {
  width: 100%;
  max-width: 100%;
  min-height: 100dvh;
  margin: 0 auto;
  padding: 18px 36px;
  overflow-x: hidden;
  background: #0b0c15;
  color: #fff;
}

.bind-account-header {
  display: grid;
  grid-template-columns: 48px minmax(0, 1fr) 48px;
  align-items: center;
  min-height: 48px;
  margin-bottom: 52px;
}

.bind-account-header h1 {
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

.bind-account-form {
  display: grid;
  gap: 18px;
}

.bind-account-field {
  display: grid;
  gap: 10px;
}

.bind-account-field span {
  color: #a0a3ad;
  font-size: 15px;
  font-weight: 700;
}

.bind-account-field input {
  width: 100%;
  min-height: 70px;
  border: 0;
  border-radius: 22px;
  outline: 2px solid transparent;
  background: #1b1d27;
  color: #fff;
  padding: 0 22px;
  font-size: 19px;
  font-weight: 700;
}

.bind-account-field input:focus {
  outline-color: #00c313;
}

.bind-account-field input::placeholder {
  color: #777a86;
}

.form-message {
  margin: -4px 0 0;
  color: #00c313;
  font-size: 14px;
  font-weight: 700;
}

.form-message--error {
  color: #ff6868;
}

.submit-button {
  min-height: 70px;
  margin-top: 12px;
  border: 0;
  border-radius: 999px;
  background: #00c313;
  color: #fff;
  font-size: 22px;
  font-weight: 900;
}

.submit-button:disabled {
  opacity: 0.72;
}

@media (max-width: 520px) {
  .bind-account-page {
    padding: 16px 24px;
  }

  .bind-account-header {
    grid-template-columns: 42px minmax(0, 1fr) 42px;
    min-height: 42px;
    margin-bottom: 46px;
  }

  .bind-account-header h1 {
    font-size: 22px;
  }

  .back-button {
    width: 40px;
    height: 40px;
  }

  .bind-account-field input,
  .submit-button {
    min-height: 64px;
    border-radius: 20px;
    font-size: 18px;
  }
}
</style>
