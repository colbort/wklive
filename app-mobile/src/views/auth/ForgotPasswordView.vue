<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'

import AppIcon from '@/components/common/AppIcon.vue'
import BottomDrawer from '@/components/common/BottomDrawer.vue'
import { useI18n } from '@/i18n'

const router = useRouter()
const { t } = useI18n()
const account = ref('')
const password = ref('')
const confirmPassword = ref('')
const showPassword = ref(false)
const showConfirmPassword = ref(false)
const showVerifySheet = ref(false)
const errorMessage = ref('')

const strengthLevel = computed(() => {
  let level = 0
  if (password.value.length >= 8) level += 1
  if (/[A-Z]/.test(password.value) && /[a-z]/.test(password.value)) level += 1
  if (/\d/.test(password.value) || /[^A-Za-z0-9]/.test(password.value)) level += 1
  if (password.value.length >= 12) level += 1
  return level
})

function goBack() {
  if (window.history.length > 1) {
    router.back()
    return
  }
  router.push('/login')
}

function goLanguageSelect() {
  router.push('/language')
}

function submitReset() {
  errorMessage.value = ''
  if (!account.value.trim()) {
    errorMessage.value = t('auth.inputAccount')
    return
  }
  if (password.value.length < 8) {
    errorMessage.value = t('auth.passwordMin8')
    return
  }
  if (password.value !== confirmPassword.value) {
    errorMessage.value = t('security.passwordMismatch')
    return
  }
  showVerifySheet.value = true
}
</script>

<template>
  <section class="auth-page">
    <header class="auth-topbar">
      <button
        type="button"
        class="icon-button"
        :aria-label="t('common.back')"
        @click="goBack"
      >
        <AppIcon name="back" class="back-icon-svg" />
      </button>
      <div class="auth-topbar__right">
        <button
          type="button"
          class="icon-button"
          :aria-label="t('userMenu.customerService')"
          @click="showVerifySheet = true"
        >
          <AppIcon name="headset" class="top-icon-svg" />
        </button>
        <button
          type="button"
          class="icon-button"
          :aria-label="t('common.language')"
          @click="goLanguageSelect"
        >
          <AppIcon name="globe" class="top-icon-svg" />
        </button>
      </div>
    </header>

    <main class="auth-content">
      <h1>{{ t('auth.forgotTitle') }}</h1>

      <form class="auth-form" @submit.prevent="submitReset">
        <label class="auth-field">
          <input
            v-model="account"
            :placeholder="t('auth.accountPlaceholder')"
            autocomplete="username"
          >
        </label>

        <label class="auth-field">
          <input
            v-model="password"
            :type="showPassword ? 'text' : 'password'"
            :placeholder="t('auth.passwordMin8')"
            autocomplete="new-password"
          >
          <button
            type="button"
            class="field-action"
            :aria-label="t('security.togglePassword')"
            @click="showPassword = !showPassword"
          >
            <AppIcon :name="showPassword ? 'eye' : 'eye-off'" class="field-action-svg" />
          </button>
        </label>

        <div class="strength-bars" aria-hidden="true">
          <span v-for="index in 4" :key="index" :class="{ active: strengthLevel >= index }" />
        </div>

        <label class="auth-field">
          <input
            v-model="confirmPassword"
            :type="showConfirmPassword ? 'text' : 'password'"
            :placeholder="t('auth.confirmNewPassword')"
            autocomplete="new-password"
          >
          <button
            type="button"
            class="field-action"
            :aria-label="t('security.togglePassword')"
            @click="showConfirmPassword = !showConfirmPassword"
          >
            <AppIcon :name="showConfirmPassword ? 'eye' : 'eye-off'" class="field-action-svg" />
          </button>
        </label>

        <p v-if="errorMessage" class="auth-error">
          {{ errorMessage }}
        </p>

        <button type="submit" class="primary-button">
          {{ t('auth.retrievePassword') }}
        </button>
        <RouterLink to="/login" class="login-link">
          {{ t('auth.goLogin') }}
        </RouterLink>
      </form>
    </main>

    <BottomDrawer
      v-model="showVerifySheet"
      :title="t('auth.verifyMethod')"
      :aria-label="t('auth.verifyMethod')"
      :close-label="t('common.close')"
      max-height="88dvh"
      :z-index="30"
    >
      <div class="verify-sheet">
        <button type="button" class="service-button">
          {{ t('auth.contactService') }}
        </button>
      </div>
    </BottomDrawer>
  </section>
</template>

<style scoped>
.auth-page {
  width: min(100%, var(--app-width, 414px));
  max-width: var(--app-width, 414px);
  height: 100dvh;
  min-height: 100dvh;
  overflow-x: hidden;
  overflow-y: auto;
  overscroll-behavior-x: none;
  -webkit-overflow-scrolling: touch;
  margin: 0 auto;
  padding: 24px 28px 42px;
  background: #0d0e17;
  color: var(--text);
}

.auth-topbar {
  position: sticky;
  top: 0;
  z-index: 30;
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: -24px -28px 0;
  padding: 24px 28px 10px;
  background: #0d0e17;
}

.auth-topbar__right {
  display: inline-flex;
  gap: 16px;
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
  color: var(--text);
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

.headset-icon {
  position: relative;
  width: 30px;
  height: 30px;
  border: 4px solid currentColor;
  border-bottom-color: transparent;
  border-radius: 50% 50% 8px 8px;
}

.headset-icon::before,
.headset-icon::after {
  content: '';
  position: absolute;
  top: 10px;
  width: 8px;
  height: 14px;
  border: 3px solid currentColor;
  border-radius: 6px;
}

.headset-icon::before {
  left: -8px;
}

.headset-icon::after {
  right: -8px;
}

.auth-content {
  width: 100%;
  min-width: 0;
  padding-top: 76px;
}

.auth-content h1 {
  margin: 0 0 78px;
  font-size: 42px;
  line-height: 1;
  font-weight: 900;
  letter-spacing: 0;
}

.auth-form {
  display: grid;
  gap: 36px;
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
  color: var(--text);
  font-size: 24px;
  font-weight: 800;
}

.auth-field input::placeholder {
  color: #8f9098;
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

.strength-bars {
  display: grid;
  width: 212px;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
  margin: -22px 0 -6px 14px;
}

.strength-bars span {
  height: 5px;
  border-radius: 5px;
  background: #1e202b;
}

.strength-bars span.active {
  background: #00c313;
}

.auth-error {
  margin: -14px 0 0;
  color: #ff6666;
  font-size: 16px;
  font-weight: 700;
}

.primary-button {
  min-height: 102px;
  margin-top: 34px;
  border: 0;
  border-radius: 50px;
  background: #00c313;
  color: var(--text);
  font-size: 30px;
  font-weight: 900;
}

.login-link {
  justify-self: center;
  color: #00c313;
  font-size: 26px;
  font-weight: 900;
  text-decoration: none;
}

.verify-sheet {
  min-height: 390px;
  padding-top: 46px;
}

.service-button {
  width: 100%;
  min-height: 96px;
  border: 0;
  border-radius: 48px;
  background: #464750;
  color: var(--text);
  font-size: 28px;
  font-weight: 900;
}

@media (max-width: 520px) {
  .auth-page {
    padding: 20px 26px 36px;
  }

  .auth-content {
    padding-top: 70px;
  }

  .auth-content h1 {
    font-size: 38px;
  }

  .auth-field,
  .primary-button {
    min-height: 84px;
  }

  .auth-field input {
    font-size: 21px;
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

.headset-icon {
  width: 25px;
  height: 25px;
  border-width: 3px;
}

.auth-content {
  padding-top: 54px;
}

.auth-content h1 {
  margin-bottom: 52px;
  font-size: 34px;
}

.auth-form {
  gap: 24px;
}

.auth-field {
  min-height: 74px;
  border-radius: 22px;
  padding: 0 18px;
}

.auth-field input {
  font-size: 19px;
}

.strength-bars {
  width: 180px;
  margin: -14px 0 -4px 12px;
}

.primary-button {
  min-height: 76px;
  margin-top: 22px;
  border-radius: 38px;
  font-size: 24px;
}

.login-link {
  font-size: 20px;
}

.verify-sheet {
  min-height: 310px;
  padding-top: 30px;
}

.service-button {
  min-height: 74px;
  border-radius: 37px;
  font-size: 22px;
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
    padding-top: 46px;
  }

  .auth-content h1 {
    margin-bottom: 42px;
    font-size: 30px;
  }

  .auth-form {
    gap: 20px;
  }

  .auth-field {
    min-height: 66px;
    border-radius: 18px;
    padding: 0 14px;
  }

  .auth-field input {
    font-size: 17px;
  }

  .primary-button {
    min-height: 68px;
    font-size: 21px;
  }
}

.top-icon-svg,
.field-action-svg {
  width: 28px;
  height: 28px;
}

.back-icon-svg {
  width: 24px;
  height: 24px;
  transform: translateX(-1px);
}

@media (min-width: 0) {
  .auth-page {
    padding: 16px 22px 28px;
    background: var(--page-bg);
  }

  .auth-topbar {
    margin: -16px -22px 0;
    padding: 16px 22px 6px;
    background: var(--page-bg);
  }

  .auth-topbar__right {
    gap: 12px;
  }

  .icon-button {
    width: 38px;
    height: 38px;
  }

  .chevron-left {
    width: 13px;
    height: 13px;
    border-left-width: 3px;
    border-bottom-width: 3px;
  }

  .top-icon-svg {
    width: 22px;
    height: 22px;
  }

  .back-icon-svg {
    width: 23px;
    height: 23px;
  }

  .auth-content {
    padding-top: 44px;
  }

  .auth-content h1 {
    margin-bottom: 74px;
    font-size: 25px;
    font-weight: 800;
    line-height: 1;
  }

  .auth-form {
    gap: 24px;
  }

  .auth-field {
    min-height: 58px;
    border-radius: 18px;
    background: #1f212c;
    padding: 0 14px;
  }

  .auth-field input {
    font-size: 17px;
    font-weight: 500;
  }

  .field-action {
    width: 28px;
    height: 28px;
    padding: 0;
  }

  .field-action-svg {
    width: 21px;
    height: 21px;
  }

  .strength-bars {
    width: 128px;
    gap: 5px;
    margin: -20px 0 -2px 8px;
  }

  .strength-bars span {
    height: 4px;
    border-radius: 4px;
  }

  .primary-button {
    min-height: 60px;
    margin-top: 34px;
    border-radius: 999px;
    font-size: 20px;
    font-weight: 800;
  }

  .login-link {
    margin-top: -4px;
    font-size: 18px;
    font-weight: 700;
  }
}
</style>
