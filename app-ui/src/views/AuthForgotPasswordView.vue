<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
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

function submitReset() {
  errorMessage.value = ''
  if (!account.value.trim()) {
    errorMessage.value = '请输入邮箱或手机号'
    return
  }
  if (password.value.length < 8) {
    errorMessage.value = '密码最少8个字符'
    return
  }
  if (password.value !== confirmPassword.value) {
    errorMessage.value = '两次输入的密码不一致'
    return
  }
  showVerifySheet.value = true
}
</script>

<template>
  <section class="auth-page">
    <header class="auth-topbar">
      <button type="button" class="icon-button" aria-label="返回" @click="goBack">
        <span class="chevron-left" />
      </button>
      <div class="auth-topbar__right">
        <button type="button" class="icon-button" aria-label="客服" @click="showVerifySheet = true">
          <span class="headset-icon" />
        </button>
        <button type="button" class="icon-button" aria-label="语言">
          <span class="globe-icon" />
        </button>
      </div>
    </header>

    <main class="auth-content">
      <h1>忘记密码</h1>

      <form class="auth-form" @submit.prevent="submitReset">
        <label class="auth-field">
          <input v-model="account" placeholder="邮箱/手机号" autocomplete="username" />
        </label>

        <label class="auth-field">
          <input
            v-model="password"
            :type="showPassword ? 'text' : 'password'"
            placeholder="密码最少8个字符"
            autocomplete="new-password"
          />
          <button
            type="button"
            class="field-action"
            aria-label="切换密码显示"
            @click="showPassword = !showPassword"
          >
            <span class="eye-off-icon" />
          </button>
        </label>

        <div class="strength-bars" aria-hidden="true">
          <span v-for="index in 4" :key="index" :class="{ active: strengthLevel >= index }" />
        </div>

        <label class="auth-field">
          <input
            v-model="confirmPassword"
            :type="showConfirmPassword ? 'text' : 'password'"
            placeholder="请再次输入新密码"
            autocomplete="new-password"
          />
          <button
            type="button"
            class="field-action"
            aria-label="切换密码显示"
            @click="showConfirmPassword = !showConfirmPassword"
          >
            <span class="eye-off-icon" />
          </button>
        </label>

        <p v-if="errorMessage" class="auth-error">{{ errorMessage }}</p>

        <button type="submit" class="primary-button">找回密码</button>
        <RouterLink to="/login" class="login-link">去登录</RouterLink>
      </form>
    </main>

    <Transition name="sheet">
      <div v-if="showVerifySheet" class="sheet-layer" @click.self="showVerifySheet = false">
        <section class="verify-sheet" role="dialog" aria-modal="true" aria-label="验证方式">
          <i class="sheet-handle" />
          <button type="button" class="sheet-close" aria-label="关闭" @click="showVerifySheet = false">
            <span />
          </button>
          <h2>验证方式</h2>
          <button type="button" class="service-button">联系客服</button>
        </section>
      </div>
    </Transition>
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
  color: #fff;
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
  color: #fff;
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

.sheet-layer {
  position: fixed;
  inset: 0;
  z-index: 30;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  background: rgba(0, 0, 0, 0.62);
  backdrop-filter: blur(10px);
}

.verify-sheet {
  position: relative;
  width: min(100%, 760px);
  max-height: 88dvh;
  overflow-x: hidden;
  overflow-y: auto;
  min-height: 390px;
  border-radius: 32px 32px 0 0;
  background: #25262f;
  padding: 76px 28px 42px;
}

.sheet-handle {
  position: absolute;
  top: 18px;
  left: 50%;
  width: 58px;
  height: 7px;
  border-radius: 8px;
  background: #a8a9ae;
  transform: translateX(-50%);
}

.sheet-close {
  position: absolute;
  top: 42px;
  right: 30px;
  width: 42px;
  height: 42px;
  border: 0;
  background: transparent;
  color: #fff;
}

.sheet-close span::before,
.sheet-close span::after {
  content: '';
  position: absolute;
  top: 20px;
  left: 6px;
  width: 30px;
  height: 3px;
  background: currentColor;
}

.sheet-close span::before {
  transform: rotate(45deg);
}

.sheet-close span::after {
  transform: rotate(-45deg);
}

.verify-sheet h2 {
  margin: 0 0 70px;
  text-align: center;
  font-size: 30px;
  line-height: 1;
  font-weight: 900;
}

.service-button {
  width: 100%;
  min-height: 96px;
  border: 0;
  border-radius: 48px;
  background: #464750;
  color: #fff;
  font-size: 28px;
  font-weight: 900;
}

.sheet-enter-active,
.sheet-leave-active {
  transition: opacity 0.18s ease;
}

.sheet-enter-active .verify-sheet,
.sheet-leave-active .verify-sheet {
  transition: transform 0.18s ease;
}

.sheet-enter-from,
.sheet-leave-to {
  opacity: 0;
}

.sheet-enter-from .verify-sheet,
.sheet-leave-to .verify-sheet {
  transform: translateY(100%);
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
  padding: 58px 22px 32px;
}

.verify-sheet h2 {
  margin-bottom: 46px;
  font-size: 24px;
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
</style>
