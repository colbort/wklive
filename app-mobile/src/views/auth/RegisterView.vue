<script setup lang="ts">
import QRCode from 'qrcode'
import { computed, nextTick, onUnmounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'

import { getTenantCode } from '@/api/http'
import {
  apiEnableGoogle2FA,
  apiInitGoogle2FA,
  apiSetPayPassword,
  apiSubmitIdentity,
} from '@/api/userPrivate'
import { apiRegister, apiSendVerificationCode } from '@/api/userPublic'
import RotateCaptcha from '@/components/auth/RotateCaptcha.vue'
import AppIcon from '@/components/common/AppIcon.vue'
import CountryDialCodePicker from '@/components/common/CountryDialCodePicker.vue'
import { countryDialCodes, type CountryDialCode } from '@/constants/countryDialCodes'
import { useI18n } from '@/i18n'
import {
  VERIFICATION_CODE_CHANNEL_EMAIL,
  VERIFICATION_CODE_CHANNEL_PHONE,
  VERIFICATION_CODE_SCENE_REGISTER,
} from '@/types/auth'

const REGISTER_TYPE_PHONE = 2
const REGISTER_TYPE_EMAIL = 3
type IdentityFileKey = 'front' | 'back' | 'handheld'
type CodeInputKind = 'email' | 'google'
const router = useRouter()
const { t } = useI18n()
const step = ref(1)
const accountMode = ref<'email' | 'phone'>('email')
const account = ref('')
const password = ref('')
const confirmPassword = ref('')
const inviteCode = ref('')
const agreed = ref(true)
const payPassword = ref('')
const emailCode = ref<string[]>(Array(6).fill(''))
const emailCodeInputs = ref<HTMLInputElement[]>([])
const identityName = ref('')
const identityNo = ref('')
const identityFiles = ref({
  front: '',
  back: '',
  handheld: '',
})
const googleSecret = ref('')
const googleQr = ref('')
const googleCode = ref<string[]>(Array(6).fill(''))
const googleCodeInputs = ref<HTMLInputElement[]>([])
const showPassword = ref(false)
const showConfirmPassword = ref(false)
const showPayPassword = ref(false)
const showCaptcha = ref(true)
const captchaPassed = ref(false)
const selectedCountry = ref<CountryDialCode>(
  countryDialCodes.find((item) => item.dialCode === '+1') || countryDialCodes[0],
)
const emailCodeCountdown = ref(0)
let emailCodeTimer: number | undefined
const submitting = ref(false)
const sendingCode = ref(false)
const errorMessage = ref('')

const steps = [
  { index: 1, labelKey: 'auth.createAccount' },
  { index: 2, labelKey: 'auth.payPassword' },
  { index: 3, labelKey: 'auth.verifyCode' },
  { index: 4, labelKey: 'auth.identityVerify' },
  { index: 5, labelKey: 'auth.googleAuthenticator' },
]

const passwordStrength = computed(() => {
  let level = 0
  if (password.value.length >= 8) level += 1
  if (/[A-Z]/.test(password.value) && /[a-z]/.test(password.value)) level += 1
  if (/\d/.test(password.value) || /[^A-Za-z0-9]/.test(password.value)) level += 1
  if (password.value.length >= 12) level += 1
  return level
})
const payPasswordStrength = computed(() =>
  payPassword.value.length >= 6 ? 4 : payPassword.value.length ? 1 : 0,
)
const accountPlaceholder = computed(() =>
  accountMode.value === 'email' ? t('auth.yourEmail') : t('auth.phonePlaceholder'),
)
const codeValue = computed(() => emailCode.value.join(''))
const googleCodeValue = computed(() => googleCode.value.join(''))

watch(step, (value) => {
  if (value === 3) {
    sendVerificationCode()
  }
  if (value === 5) {
    loadGoogle2FA()
  }
})

onUnmounted(() => {
  stopEmailCodeCountdown()
})

function goBack() {
  if (showCaptcha.value) {
    router.push('/login')
    return
  }
  if (step.value > 1) {
    step.value -= 1
    return
  }
  router.push('/login')
}

function skipStep() {
  if (step.value < 5) {
    step.value += 1
    return
  }
  router.push('/profile')
}

function getCodeState(kind: CodeInputKind) {
  return kind === 'email'
    ? { code: emailCode.value, inputs: emailCodeInputs.value }
    : { code: googleCode.value, inputs: googleCodeInputs.value }
}

function setCodeInputRef(kind: CodeInputKind, element: unknown, index: number) {
  if (!(element instanceof HTMLInputElement)) return
  getCodeState(kind).inputs[index] = element
}

function focusCodeInput(kind: CodeInputKind, index: number) {
  const { code, inputs } = getCodeState(kind)
  const target = inputs[Math.max(0, Math.min(index, code.length - 1))]
  if (!target) return
  nextTick(() => {
    target.focus()
    target.select()
  })
}

function applyCodeDigits(kind: CodeInputKind, index: number, value: string) {
  const { code } = getCodeState(kind)
  const digits = value.replace(/\D/g, '').slice(0, code.length - index)
  if (!digits) {
    code[index] = ''
    return
  }

  digits.split('').forEach((digit, offset) => {
    code[index + offset] = digit
  })

  focusCodeInput(kind, Math.min(index + digits.length, code.length - 1))
}

function handleCodeInput(kind: CodeInputKind, index: number, event: Event) {
  applyCodeDigits(kind, index, (event.target as HTMLInputElement).value)
}

function selectCodeInput(event: FocusEvent) {
  const target = event.target as HTMLInputElement
  target.select()
}

function handleCodeKeydown(kind: CodeInputKind, index: number, event: KeyboardEvent) {
  const { code } = getCodeState(kind)
  if (event.key === 'Backspace') {
    event.preventDefault()
    if (code[index]) {
      code[index] = ''
      return
    }
    if (index > 0) {
      code[index - 1] = ''
      focusCodeInput(kind, index - 1)
    }
    return
  }

  if (event.key === 'Delete') {
    event.preventDefault()
    code[index] = ''
    return
  }

  if (event.key === 'ArrowLeft' && index > 0) {
    event.preventDefault()
    focusCodeInput(kind, index - 1)
    return
  }

  if (event.key === 'ArrowRight' && index < code.length - 1) {
    event.preventDefault()
    focusCodeInput(kind, index + 1)
  }
}

function handleCodePaste(kind: CodeInputKind, index: number, event: ClipboardEvent) {
  event.preventDefault()
  applyCodeDigits(kind, index, event.clipboardData?.getData('text') || '')
}

function stopEmailCodeCountdown() {
  if (!emailCodeTimer) return
  window.clearInterval(emailCodeTimer)
  emailCodeTimer = undefined
}

function startEmailCodeCountdown(seconds = 115) {
  stopEmailCodeCountdown()
  emailCodeCountdown.value = seconds
  emailCodeTimer = window.setInterval(() => {
    emailCodeCountdown.value -= 1
    if (emailCodeCountdown.value <= 0) {
      emailCodeCountdown.value = 0
      stopEmailCodeCountdown()
    }
  }, 1000)
}

function normalizedAccount() {
  const value = account.value.trim()
  if (accountMode.value === 'email') return value
  return `${selectedCountry.value.dialCode}${value.replace(/^\+/, '')}`
}

async function sendVerificationCode() {
  if (emailCodeCountdown.value > 0 || sendingCode.value) return
  const target = normalizedAccount()
  if (!target) {
    errorMessage.value =
      accountMode.value === 'email' ? t('security.inputEmail') : t('security.inputPhone')
    return
  }

  sendingCode.value = true
  errorMessage.value = ''
  emailCode.value = Array(6).fill('')
  try {
    const res = await apiSendVerificationCode({
      channel:
        accountMode.value === 'email'
          ? VERIFICATION_CODE_CHANNEL_EMAIL
          : VERIFICATION_CODE_CHANNEL_PHONE,
      email: accountMode.value === 'email' ? target : undefined,
      phone: accountMode.value === 'phone' ? target : undefined,
      scene: VERIFICATION_CODE_SCENE_REGISTER,
    })
    if (res.code !== 200) {
      errorMessage.value = res.msg || t('auth.verificationCodeSendFailed')
      return
    }
    startEmailCodeCountdown()
  } catch (error) {
    console.warn('send verification code failed', error)
    errorMessage.value = t('auth.verificationCodeSendFailed')
  } finally {
    sendingCode.value = false
  }
}

async function continueStep() {
  errorMessage.value = ''
  if (step.value === 1) {
    await submitRegister()
  } else if (step.value === 2) {
    await submitPayPassword()
  } else if (step.value === 3) {
    if (codeValue.value.length !== 6) {
      errorMessage.value = t('auth.inputSixDigitCode')
      focusCodeInput(
        'email',
        emailCode.value.findIndex((digit) => !digit),
      )
      return
    }
    step.value = 4
  } else if (step.value === 4) {
    await submitIdentity()
  } else {
    await submitGoogle2FA()
  }
}

async function submitRegister() {
  const tenantCode = getTenantCode()
  if (!tenantCode) {
    errorMessage.value = t('profile.tenantMissing')
    return
  }
  if (!account.value.trim()) {
    errorMessage.value =
      accountMode.value === 'email' ? t('security.inputEmail') : t('security.inputPhone')
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
  if (!agreed.value) {
    errorMessage.value = t('auth.agreeTermsRequired')
    return
  }

  submitting.value = true
  try {
    const payload = {
      tenantCode,
      registerType: accountMode.value === 'email' ? REGISTER_TYPE_EMAIL : REGISTER_TYPE_PHONE,
      password: password.value,
      confirmPassword: confirmPassword.value,
      inviteCode: inviteCode.value.trim() || undefined,
      email: accountMode.value === 'email' ? normalizedAccount() : undefined,
      phone: accountMode.value === 'phone' ? normalizedAccount() : undefined,
    }
    const res = await apiRegister(payload)
    if (res.code !== 200) {
      errorMessage.value = res.msg || t('auth.registerFailed')
      return
    }
    step.value = 2
  } catch (error) {
    console.warn('register failed', error)
    errorMessage.value = t('auth.registerFailed')
  } finally {
    submitting.value = false
  }
}

async function submitPayPassword() {
  if (payPassword.value.length < 6) {
    errorMessage.value = t('auth.inputPayPassword')
    return
  }
  submitting.value = true
  try {
    const res = await apiSetPayPassword({
      password: payPassword.value,
      confirmPassword: payPassword.value,
    })
    if (res.code !== 200) {
      errorMessage.value = res.msg || t('auth.setFailed')
      return
    }
    step.value = 3
  } catch (error) {
    console.warn('set pay password failed', error)
    errorMessage.value = t('auth.setFailed')
  } finally {
    submitting.value = false
  }
}

async function submitIdentity() {
  if (!identityName.value.trim() || !identityNo.value.trim()) {
    errorMessage.value = t('auth.inputIdentityInfo')
    return
  }
  submitting.value = true
  try {
    const res = await apiSubmitIdentity({
      realName: identityName.value.trim(),
      idType: 1,
      idNo: identityNo.value.trim(),
      frontImage: identityFiles.value.front,
      backImage: identityFiles.value.back,
      handheldImage: identityFiles.value.handheld,
      kycLevel: 1,
    })
    if (res.code !== 200) {
      errorMessage.value = res.msg || t('common.failed')
      return
    }
    step.value = 5
  } catch (error) {
    console.warn('submit identity failed', error)
    errorMessage.value = t('common.failed')
  } finally {
    submitting.value = false
  }
}

async function loadGoogle2FA() {
  if (googleSecret.value || googleQr.value) return
  try {
    const res = await apiInitGoogle2FA()
    if (res.code !== 200) return
    googleSecret.value = res.data?.secret || ''
    googleQr.value = res.data?.qrCodeUrl || ''
    if (!googleQr.value && googleSecret.value) {
      googleQr.value = await QRCode.toDataURL(googleSecret.value, {
        errorCorrectionLevel: 'M',
        margin: 1,
        width: 220,
        color: { dark: '#000000', light: '#ffffff' },
      })
    }
  } catch (error) {
    console.warn('init google 2fa failed', error)
  }
}

async function submitGoogle2FA() {
  if (googleCodeValue.value.length !== 6) {
    errorMessage.value = t('auth.googleCode')
    focusCodeInput(
      'google',
      googleCode.value.findIndex((digit) => !digit),
    )
    return
  }
  submitting.value = true
  try {
    const res = await apiEnableGoogle2FA({ googleCode: googleCodeValue.value })
    if (res.code !== 200) {
      errorMessage.value = res.msg || t('auth.bindFailed')
      return
    }
    router.replace('/profile')
  } catch (error) {
    console.warn('enable google 2fa failed', error)
    errorMessage.value = t('auth.bindFailed')
  } finally {
    submitting.value = false
  }
}

function completeCaptcha() {
  captchaPassed.value = true
  showCaptcha.value = false
  errorMessage.value = ''
}

function markUpload(type: IdentityFileKey) {
  identityFiles.value[type] = `${type}-uploaded`
}
</script>

<template>
  <section class="register-page" :class="{ 'register-page--captcha': showCaptcha }">
    <header class="register-topbar">
      <button
        type="button"
        class="icon-button"
        :aria-label="t('common.back')"
        @click="goBack"
      >
        <AppIcon name="back" class="back-icon-svg" />
      </button>
      <button
        v-if="showCaptcha"
        type="button"
        class="icon-button"
        :aria-label="t('common.language')"
        @click="router.push('/language')"
      >
        <AppIcon name="globe" class="top-icon-svg" />
      </button>
      <button
        v-else-if="step > 1"
        type="button"
        class="skip-button"
        @click="skipStep"
      >
        {{ t('common.skip') }}
      </button>
    </header>

    <RotateCaptcha v-if="showCaptcha" @success="completeCaptcha" />

    <main v-else class="register-content">
      <nav class="register-steps" :aria-label="t('auth.registerSteps')">
        <div
          v-for="item in steps"
          :key="item.index"
          class="step-item"
          :class="{ done: item.index < step, active: item.index === step }"
        >
          <span>
            <svg
              v-if="item.index < step"
              class="step-status-svg step-status-svg--done"
              viewBox="0 0 24 24"
              aria-hidden="true"
            >
              <polyline points="6 12.5 10.2 16.5 18 7.5" />
            </svg>
            <svg
              v-else-if="item.index === step"
              class="step-status-svg step-status-svg--active"
              viewBox="0 0 24 24"
              aria-hidden="true"
            >
              <circle cx="12" cy="12" r="10.8" />
              <circle cx="12" cy="12" r="1.8" />
            </svg>
            <template v-else>{{ item.index }}</template>
          </span>
          <em>{{ t(item.labelKey) }}</em>
        </div>
      </nav>

      <section v-if="step === 1" class="step-panel">
        <h1>{{ t('auth.createYourAccount') }}</h1>
        <div class="auth-tabs" role="tablist" :aria-label="t('auth.registerMethod')">
          <button
            type="button"
            :class="{ active: accountMode === 'email' }"
            @click="accountMode = 'email'"
          >
            {{ t('auth.email') }}
          </button>
          <button
            type="button"
            :class="{ active: accountMode === 'phone' }"
            @click="accountMode = 'phone'"
          >
            {{ t('auth.phone') }}
          </button>
        </div>

        <label class="auth-field">
          <CountryDialCodePicker v-if="accountMode === 'phone'" v-model="selectedCountry" />
          <input v-model="account" :placeholder="accountPlaceholder" autocomplete="username">
          <button
            type="button"
            class="field-action"
            :class="{ 'field-action--hidden': !account }"
            :aria-label="t('common.clear')"
            :tabindex="account ? 0 : -1"
            @click="account = ''"
          >
            <AppIcon name="close-circle" class="field-action-svg" />
          </button>
        </label>
        <label class="auth-field">
          <input
            v-model="password"
            :type="showPassword ? 'text' : 'password'"
            :placeholder="t('auth.passwordMin8')"
            autocomplete="new-password"
          >
          <button type="button" class="field-action" @click="showPassword = !showPassword">
            <AppIcon :name="showPassword ? 'eye' : 'eye-off'" class="field-action-svg" />
          </button>
        </label>
        <div class="strength-bars">
          <span v-for="index in 4" :key="index" :class="{ active: passwordStrength >= index }" />
        </div>
        <label class="auth-field">
          <input
            v-model="confirmPassword"
            :type="showConfirmPassword ? 'text' : 'password'"
            :placeholder="t('security.confirmNewPassword')"
            autocomplete="new-password"
          >
          <button
            type="button"
            class="field-action"
            @click="showConfirmPassword = !showConfirmPassword"
          >
            <AppIcon :name="showConfirmPassword ? 'eye' : 'eye-off'" class="field-action-svg" />
          </button>
        </label>
        <label class="auth-field">
          <input v-model="inviteCode" :placeholder="t('auth.inviteCode')">
        </label>
        <label class="agree-control">
          <input v-model="agreed" type="checkbox">
          <span>
            <svg viewBox="0 0 16 16" aria-hidden="true">
              <path d="M3.25 8.1 6.45 11.2 12.8 4.8" />
            </svg>
          </span>
          <em>{{ t('auth.agreeTerms') }}<b>{{ t('auth.privacyPolicy') }}</b>{{ t('common.and') }}<b>{{ t('auth.userTerms') }}</b></em>
        </label>
      </section>

      <section v-else-if="step === 2" class="step-panel step-panel--loose">
        <h1>{{ t('auth.payPasswordSetting') }}</h1>
        <label class="auth-field">
          <input
            v-model="payPassword"
            :type="showPayPassword ? 'text' : 'password'"
            :placeholder="t('auth.payPassword')"
          >
          <button type="button" class="field-action" @click="showPayPassword = !showPayPassword">
            <AppIcon :name="showPayPassword ? 'eye' : 'eye-off'" class="field-action-svg" />
          </button>
        </label>
        <div class="strength-bars">
          <span v-for="index in 4" :key="index" :class="{ active: payPasswordStrength >= index }" />
        </div>
      </section>

      <section v-else-if="step === 3" class="step-panel step-panel--loose verify-panel">
        <h1>{{ t('auth.emailVerifyTitle') }}</h1>
        <p>{{ t('auth.emailVerifyHint') }}</p>
        <div class="code-head">
          <strong>{{ t('auth.inputSixDigitCode') }}</strong>
          <button
            type="button"
            class="code-send-button"
            :disabled="sendingCode || emailCodeCountdown > 0"
            @click="sendVerificationCode"
          >
            {{
              sendingCode
                ? t('auth.sendingVerificationCode')
                : emailCodeCountdown > 0
                  ? `${emailCodeCountdown}s`
                  : t('auth.resendVerificationCode')
            }}
          </button>
        </div>
        <div class="code-boxes">
          <input
            v-for="(_, index) in emailCode"
            :key="index"
            :ref="(element) => setCodeInputRef('email', element, index)"
            :value="emailCode[index]"
            inputmode="numeric"
            autocomplete="one-time-code"
            maxlength="1"
            @focus="selectCodeInput"
            @input="handleCodeInput('email', index, $event)"
            @keydown="handleCodeKeydown('email', index, $event)"
            @paste="handleCodePaste('email', index, $event)"
          >
        </div>
      </section>

      <section v-else-if="step === 4" class="step-panel identity-panel">
        <h1>{{ t('auth.identityInfo') }}</h1>
        <label class="auth-field required-field">
          <span>*</span>
          <input v-model="identityName" :placeholder="t('auth.legalName')">
        </label>
        <label class="auth-field required-field">
          <span>*</span>
          <input v-model="identityNo" :placeholder="t('auth.idNumber')">
        </label>

        <h2>{{ t('auth.idUpload') }}</h2>
        <p>{{ t('auth.idUploadHint') }}</p>
        <div class="upload-grid">
          <button type="button" @click="markUpload('front')">
            <span class="upload-card-art upload-card-art--front" aria-hidden="true">
              <i />
              <i />
              <i />
              <b />
            </span>
            <i class="camera-dot" />
            <strong>{{ t('auth.idFront') }}</strong>
          </button>
          <button type="button" @click="markUpload('back')">
            <span class="upload-card-art upload-card-art--back" aria-hidden="true">
              <i />
              <i />
              <b />
            </span>
            <i class="camera-dot" />
            <strong>{{ t('auth.idBack') }}</strong>
          </button>
          <button type="button" @click="markUpload('handheld')">
            <span class="upload-card-art upload-card-art--handheld" aria-hidden="true">
              <i />
              <i />
              <b />
            </span>
            <i class="camera-dot" />
            <strong>{{ t('auth.idHandheld') }}</strong>
          </button>
        </div>

        <h2>{{ t('auth.uploadRequirements') }}</h2>
        <p>{{ t('auth.uploadRequirementsHint') }}</p>
        <div class="require-box">
          <span>
            <i class="require-thumb require-thumb--ok" aria-hidden="true" />
            <b>✓</b>{{ t('auth.standardShot') }}
          </span>
          <span>
            <i class="require-thumb require-thumb--cut" aria-hidden="true" />
            <b>×</b>{{ t('auth.incompleteShot') }}
          </span>
          <span>
            <i class="require-thumb require-thumb--blur" aria-hidden="true" />
            <b>×</b>{{ t('auth.blurryShot') }}
          </span>
          <span>
            <i class="require-thumb require-thumb--flash" aria-hidden="true" />
            <b>×</b>{{ t('auth.overexposedShot') }}
          </span>
        </div>
      </section>

      <section v-else class="step-panel google-panel">
        <h1>{{ t('auth.bindGoogleAuthenticator') }}</h1>
        <p>{{ t('auth.backupSecretHint') }}</p>
        <div class="qr-card">
          <img v-if="googleQr" :src="googleQr" :alt="t('auth.googleQrAlt')">
        </div>
        <div class="secret-card">
          <strong>{{ googleSecret || '' }}</strong>
          <button type="button">
            {{ t('common.copy') }}
          </button>
        </div>
        <h2>{{ t('auth.googleCode') }}</h2>
        <div class="code-boxes">
          <input
            v-for="(_, index) in googleCode"
            :key="index"
            :ref="(element) => setCodeInputRef('google', element, index)"
            :value="googleCode[index]"
            inputmode="numeric"
            autocomplete="one-time-code"
            maxlength="1"
            @focus="selectCodeInput"
            @input="handleCodeInput('google', index, $event)"
            @keydown="handleCodeKeydown('google', index, $event)"
            @paste="handleCodePaste('google', index, $event)"
          >
        </div>
      </section>

      <p v-if="errorMessage" class="auth-error">
        {{ errorMessage }}
      </p>
      <button
        type="button"
        class="primary-button"
        :class="{ 'primary-button--identity': step === 4 }"
        :disabled="submitting"
        @click="continueStep"
      >
        {{
          step === 5 ? t('auth.bind') : submitting ? t('common.submitting') : t('common.continue')
        }}
      </button>
      <p v-if="step === 1" class="auth-switch">
        {{ t('auth.haveAccount') }}
        <button type="button" @click="router.push('/login')">
          {{ t('auth.goLogin') }}
        </button>
      </p>
    </main>
  </section>
</template>

<style scoped>
.register-page {
  width: 100%;
  max-width: 100%;
  height: 100dvh;
  min-height: 100dvh;
  overflow-x: hidden;
  overflow-y: auto;
  overscroll-behavior-x: none;
  -webkit-overflow-scrolling: touch;
  margin: 0 auto;
  padding: 24px 28px 34px;
  background: #0d0e17;
  color: #fff;
}

.register-topbar {
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

.top-icon-svg {
  width: 28px;
  height: 28px;
}

.back-icon-svg {
  width: 24px;
  height: 24px;
  transform: translateX(-1px);
}

.skip-button {
  border: 0;
  background: transparent;
  color: #fff;
  font-size: 26px;
  font-weight: 900;
}

.register-content {
  width: 100%;
  min-width: 0;
  padding-top: 62px;
}

.register-steps {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 8px;
  margin-bottom: 76px;
}

.step-item {
  position: relative;
  display: grid;
  min-width: 0;
  justify-items: center;
  gap: 12px;
  color: #fff;
  font-weight: 900;
}

.step-item:not(:last-child)::after {
  content: '';
  position: absolute;
  top: 18px;
  left: calc(50% + 24px);
  width: calc(100% - 28px);
  height: 3px;
  background: #2a2c36;
}

.step-item span {
  position: relative;
  z-index: 1;
  display: inline-flex;
  width: 36px;
  height: 36px;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: #666973;
  color: #fff;
  font-size: 20px;
}

.step-item.active span {
  background: #f2fff2;
  color: transparent;
}

.step-item.done span {
  background: #00c313;
  font-size: 0;
}

.step-status-svg {
  display: block;
  width: 18px;
  height: 18px;
}

.step-status-svg--done {
  fill: none;
  stroke: #fff;
  stroke-width: 2.4;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.step-status-svg--active {
  fill: #00c313;
}

.step-status-svg--active circle:last-child {
  fill: #fff;
}

.step-item em {
  font-style: normal;
  font-size: 18px;
  line-height: 1.15;
  text-align: center;
  word-break: keep-all;
}

.step-panel {
  display: grid;
  width: 100%;
  min-width: 0;
  gap: 28px;
}

.step-panel--loose {
  gap: 34px;
}

.step-panel h1 {
  margin: 0;
  font-size: 38px;
  line-height: 1.15;
  font-weight: 900;
  letter-spacing: 0;
}

.step-panel p {
  margin: -16px 0 14px;
  color: #8f9098;
  font-size: 22px;
  font-weight: 800;
}

.auth-tabs {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  border-bottom: 1px solid #20222d;
}

.auth-tabs button {
  position: relative;
  height: 60px;
  border: 0;
  background: transparent;
  color: #8e9099;
  font-size: 26px;
  font-weight: 900;
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

.auth-field {
  display: flex;
  min-height: 102px;
  align-items: center;
  gap: 14px;
  border-radius: 28px;
  background: #20212b;
  padding: 0 28px;
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

.field-action--hidden {
  pointer-events: none;
  opacity: 0;
}

.field-action-svg {
  width: 28px;
  height: 28px;
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
  margin: -18px 0 -8px 30px;
}

.strength-bars span {
  height: 5px;
  border-radius: 5px;
  background: #1e202b;
}

.strength-bars span.active {
  background: #00c313;
}

.agree-control {
  display: inline-flex;
  align-items: center;
  gap: 14px;
  color: #8f9098;
  font-size: 24px;
  font-weight: 800;
}

.agree-control input {
  position: absolute;
  opacity: 0;
}

.agree-control span {
  display: grid;
  width: 30px;
  height: 30px;
  place-items: center;
  border: 2px solid #00c313;
  border-radius: 8px;
}

.agree-control span svg {
  width: 70%;
  height: 70%;
  opacity: 0;
}

.agree-control span svg path {
  fill: none;
  stroke: #00c313;
  stroke-linecap: round;
  stroke-linejoin: round;
  stroke-width: 2.6;
}

.agree-control input:checked + span svg {
  opacity: 1;
}

.agree-control em {
  font-style: normal;
}

.agree-control b,
.auth-switch button {
  color: #00c313;
}

.primary-button {
  width: 100%;
  min-height: 102px;
  margin-top: 34px;
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

.auth-error {
  margin: 14px 0 -10px;
  color: #ff6666;
  font-size: 16px;
  font-weight: 700;
}

.auth-switch {
  margin: 28px 0 0;
  text-align: center;
  color: #8f9098;
  font-size: 20px;
  font-weight: 800;
}

.auth-switch button {
  border: 0;
  background: transparent;
  font: inherit;
}

.code-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  margin-top: 36px;
}

.code-head strong,
.google-panel h2,
.identity-panel h2 {
  font-size: 28px;
  font-weight: 900;
}

.code-send-button {
  min-width: 188px;
  border: 0;
  border-radius: 42px;
  background: #20212b;
  padding: 26px;
  color: #00c313;
  text-align: center;
  font-size: 26px;
  font-weight: 900;
}

.code-send-button:disabled {
  opacity: 0.75;
}

.code-boxes {
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  gap: 16px;
}

.code-boxes input {
  width: 100%;
  aspect-ratio: 0.78;
  border: 2px solid #3a3c47;
  border-radius: 28px;
  outline: 0;
  background: #2a2b35;
  color: #fff;
  text-align: center;
  font-size: 32px;
  font-weight: 900;
}

.code-boxes input:focus {
  border-color: #00c313;
}

.required-field > span {
  color: #ff4b43;
  font-size: 24px;
  font-weight: 900;
}

.identity-panel {
  gap: 18px;
}

.identity-panel h1 {
  font-size: 26px;
  font-weight: 900;
}

.identity-panel h2 {
  margin: 14px 0 -8px;
  font-size: 24px;
  line-height: 1.2;
}

.identity-panel p {
  margin: -8px 0 0;
  font-size: 16px;
  line-height: 1.35;
  font-weight: 800;
}

.identity-panel .auth-field {
  min-height: 74px;
  border-radius: 22px;
  padding: 0 22px;
}

.identity-panel .auth-field input {
  font-size: 20px;
}

.identity-panel .required-field > span {
  font-size: 18px;
}

.primary-button--identity {
  min-height: 76px;
  margin-top: 24px;
  border-radius: 38px;
  font-size: 24px;
}

.upload-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18px;
}

.upload-grid button {
  position: relative;
  min-height: 146px;
  border: 0;
  border-radius: 22px;
  background: #2a2b35;
  color: #fff;
  overflow: hidden;
}

.upload-grid button:nth-child(3) {
  grid-column: span 1;
}

.camera-dot {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 38px;
  height: 38px;
  border-radius: 50%;
  background: #00c313;
  transform: translate(-50%, -50%);
}

.camera-dot::before {
  content: '';
  position: absolute;
  inset: 10px;
  border: 3px solid #fff;
  border-radius: 50%;
}

.upload-grid strong {
  position: absolute;
  right: 0;
  bottom: 24px;
  left: 0;
  font-size: 24px;
}

.upload-card-art {
  position: absolute;
  inset: 18px 18px 30px;
  opacity: 0.42;
}

.upload-card-art i,
.upload-card-art b {
  position: absolute;
  display: block;
  border-radius: 6px;
  background: #7c7d76;
}

.upload-card-art--front i:nth-child(1) {
  top: 18px;
  left: 8px;
  width: 10px;
  height: 10px;
}

.upload-card-art--front i:nth-child(2) {
  top: 18px;
  left: 24px;
  width: 10px;
  height: 10px;
}

.upload-card-art--front i:nth-child(3) {
  top: 18px;
  left: 40px;
  width: 10px;
  height: 10px;
}

.upload-card-art--front b {
  right: 10px;
  bottom: 8px;
  width: 42px;
  height: 52px;
  border-radius: 50% 50% 22px 22px;
}

.upload-card-art--front::before,
.upload-card-art--front::after,
.upload-card-art--back::before,
.upload-card-art--back::after {
  content: '';
  position: absolute;
  left: 8px;
  height: 8px;
  border-radius: 8px;
  background: #7c7d76;
}

.upload-card-art--front::before {
  top: 42px;
  width: 76px;
}

.upload-card-art--front::after {
  bottom: 10px;
  width: 128px;
}

.upload-card-art--back i:nth-child(1) {
  top: 28px;
  left: 48px;
  width: 52px;
  height: 52px;
  border-radius: 50%;
}

.upload-card-art--back i:nth-child(2),
.upload-card-art--back b {
  right: 8px;
  height: 8px;
  border-radius: 8px;
}

.upload-card-art--back i:nth-child(2) {
  top: 26px;
  width: 68px;
}

.upload-card-art--back b {
  bottom: 14px;
  width: 128px;
}

.upload-card-art--back::before {
  bottom: 34px;
  width: 118px;
}

.upload-card-art--back::after {
  display: none;
}

.upload-card-art--handheld {
  inset: 8px 20px 18px;
}

.upload-card-art--handheld i:nth-child(1) {
  top: 2px;
  left: 50%;
  width: 58px;
  height: 58px;
  border-radius: 50%;
  transform: translateX(-50%);
}

.upload-card-art--handheld i:nth-child(2) {
  right: 12px;
  bottom: 2px;
  left: 12px;
  height: 76px;
  border-radius: 50% 50% 12px 12px;
}

.upload-card-art--handheld b {
  bottom: 18px;
  left: 50%;
  width: 58px;
  height: 42px;
  border-radius: 6px;
  transform: translateX(-50%);
}

.require-box {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
  border-radius: 24px;
  background: #20212b;
  padding: 24px;
}

.require-box span {
  display: grid;
  min-width: 0;
  justify-items: center;
  gap: 8px;
  color: #8f9098;
  font-size: 17px;
  font-weight: 900;
  line-height: 1.1;
  text-align: center;
}

.require-box b {
  margin-right: 8px;
  color: #00c313;
}

.require-box span:not(:first-child) b {
  color: #ff5c4d;
}

.require-thumb {
  position: relative;
  display: block;
  width: 76px;
  height: 46px;
  border-radius: 12px;
  background: #454650;
  overflow: hidden;
}

.require-thumb::before,
.require-thumb::after {
  content: '';
  position: absolute;
  border-radius: 4px;
  background: #8b8c8e;
}

.require-thumb::before {
  top: 13px;
  left: 12px;
  width: 32px;
  height: 6px;
}

.require-thumb::after {
  right: 11px;
  bottom: 10px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
}

.require-thumb--cut {
  transform: scaleX(0.78);
  transform-origin: left center;
}

.require-thumb--blur {
  filter: blur(1.5px);
  opacity: 0.72;
}

.require-thumb--flash::after {
  inset: 0;
  width: auto;
  height: auto;
  border-radius: 12px;
  background: linear-gradient(100deg, transparent 18%, #fff 54%, transparent 80%);
  opacity: 0.78;
}

.qr-card {
  display: flex;
  width: min(326px, 100%);
  aspect-ratio: 1;
  height: auto;
  align-items: center;
  justify-content: center;
  border-radius: 24px;
  background: #fff;
}

.qr-card img {
  width: 286px;
  height: 286px;
  object-fit: contain;
}

.secret-card {
  display: flex;
  min-width: 0;
  min-height: 90px;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  border-radius: 22px;
  background: #2a2b35;
  padding-left: 28px;
}

.secret-card strong {
  min-width: 0;
  overflow: hidden;
  color: #fff;
  font-size: 25px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.secret-card button {
  align-self: stretch;
  min-width: 146px;
  border: 0;
  border-radius: 22px;
  background: #00c313;
  color: #fff;
  font-size: 26px;
  font-weight: 900;
}

@media (max-width: 520px) {
  .register-page {
    padding: 20px 26px 32px;
  }

  .register-steps {
    gap: 4px;
  }

  .step-item em {
    font-size: 13px;
  }

  .auth-field,
  .primary-button {
    min-height: 84px;
  }

  .auth-field input,
  .auth-tabs button {
    font-size: 21px;
  }

  .code-boxes {
    gap: 10px;
  }

  .code-boxes input {
    border-radius: 20px;
  }
}

.register-page {
  padding: 18px 22px 28px;
}

.register-topbar {
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

.skip-button {
  font-size: 20px;
  font-weight: 700;
}

.register-content {
  padding-top: 42px;
}

.register-steps {
  gap: 4px;
  margin-bottom: 48px;
}

.step-item {
  gap: 8px;
}

.step-item:not(:last-child)::after {
  top: 14px;
  left: calc(50% + 18px);
  width: calc(100% - 24px);
}

.step-item span {
  width: 28px;
  height: 28px;
  font-size: 16px;
}

.step-item.active span {
  border: 0;
}

.step-item.done span {
  font-size: 0;
}

.step-status-svg {
  width: 15px;
  height: 15px;
}

.step-status-svg--done {
  stroke-width: 2.1;
}

.step-item em {
  font-size: 13px;
}

.step-panel {
  gap: 20px;
}

.step-panel--loose {
  gap: 24px;
}

.step-panel h1 {
  font-size: 30px;
}

.step-panel p {
  font-size: 17px;
}

.auth-tabs button {
  height: 46px;
  font-size: 21px;
}

.auth-field {
  min-height: 74px;
  border-radius: 22px;
  padding: 0 20px;
}

.auth-field input {
  font-size: 19px;
}

.agree-control {
  font-size: 17px;
}

.strength-bars {
  width: 180px;
  margin: -14px 0 -5px 24px;
}

.primary-button {
  min-height: 76px;
  margin-top: 22px;
  border-radius: 38px;
  font-size: 24px;
}

.code-head {
  margin-top: 24px;
}

.code-head strong,
.google-panel h2,
.identity-panel h2 {
  font-size: 22px;
}

.code-send-button {
  min-width: 140px;
  border-radius: 32px;
  padding: 18px;
  font-size: 21px;
}

.code-boxes {
  gap: 10px;
}

.code-boxes input {
  border-radius: 20px;
  font-size: 25px;
}

.upload-grid {
  gap: 14px;
}

.upload-grid button {
  min-height: 112px;
  border-radius: 18px;
}

.identity-panel {
  gap: 16px;
}

.identity-panel h1 {
  font-size: 24px;
}

.identity-panel h2 {
  margin: 12px 0 -8px;
  font-size: 22px;
}

.identity-panel p {
  margin: -8px 0 0;
  font-size: 15px;
}

.identity-panel .auth-field {
  min-height: 58px;
  border-radius: 18px;
  padding: 0 18px;
}

.identity-panel .auth-field input {
  font-size: 17px;
}

.identity-panel .upload-grid {
  gap: 14px;
}

.identity-panel .upload-grid button {
  min-height: 112px;
  border-radius: 18px;
}

.identity-panel .upload-grid button:nth-child(3) {
  grid-column: 1 / 2;
}

.upload-grid strong {
  bottom: 16px;
  font-size: 18px;
}

.require-box {
  border-radius: 20px;
  padding: 16px;
}

.identity-panel .require-box {
  gap: 10px;
  border-radius: 18px;
  padding: 12px 14px 11px;
}

.require-box span {
  font-size: 13px;
}

.identity-panel .require-box span {
  gap: 6px;
  font-size: 12px;
}

.identity-panel .require-thumb {
  width: 62px;
  height: 38px;
  border-radius: 10px;
}

.identity-panel .require-thumb--flash::after {
  border-radius: 10px;
}

.primary-button--identity {
  min-height: 60px;
  margin-top: 18px;
  border-radius: 30px;
  font-size: 20px;
}

.qr-card {
  width: 260px;
  height: 260px;
}

.qr-card img {
  width: 230px;
  height: 230px;
}

.secret-card {
  min-height: 74px;
  border-radius: 18px;
}

.secret-card strong {
  font-size: 20px;
}

.secret-card button {
  min-width: 112px;
  border-radius: 18px;
  font-size: 21px;
}

@media (max-width: 390px) {
  .register-page {
    padding: 16px 18px 28px;
  }

  .register-topbar {
    margin: -16px -18px 0;
    padding: 16px 18px 8px;
  }

  .register-content {
    padding-top: 34px;
  }

  .register-steps {
    margin-bottom: 36px;
  }

  .step-item span {
    width: 24px;
    height: 24px;
    font-size: 14px;
  }

  .step-item:not(:last-child)::after {
    top: 12px;
    left: calc(50% + 16px);
  }

  .step-item.active span {
    border: 0;
  }

  .step-item em {
    font-size: 11px;
  }

  .step-panel h1 {
    font-size: 26px;
  }

  .auth-field {
    min-height: 66px;
    border-radius: 18px;
    padding: 0 14px;
  }

  .auth-field input,
  .auth-tabs button {
    font-size: 17px;
  }

  .agree-control,
  .auth-switch {
    font-size: 15px;
  }

  .primary-button {
    min-height: 68px;
    font-size: 21px;
  }

  .code-head {
    align-items: flex-start;
    flex-direction: column;
  }

  .code-send-button {
    min-width: 128px;
    padding: 14px 20px;
  }

  .code-boxes {
    gap: 8px;
  }

  .code-boxes input {
    border-radius: 16px;
  }

  .require-box {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .identity-panel .require-box {
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 6px;
    padding: 10px;
  }

  .identity-panel .require-box span {
    font-size: 11px;
  }

  .identity-panel .require-thumb {
    width: 52px;
    height: 32px;
  }

  .qr-card {
    width: min(230px, 100%);
  }

  .qr-card img {
    width: 202px;
    height: 202px;
  }

  .secret-card {
    padding-left: 16px;
  }

  .secret-card button {
    min-width: 92px;
  }
}

@media (min-width: 0) {
  .register-page {
    padding: 14px 22px 28px;
  }

  .register-topbar {
    margin: -14px -22px 0;
    padding: 14px 22px 8px;
  }

  .icon-button {
    width: 38px;
    height: 38px;
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

  .register-content {
    padding-top: 34px;
  }

  .register-steps {
    align-items: start;
    gap: 3px;
    margin-bottom: 42px;
  }

  .step-item {
    gap: 6px;
  }

  .step-item:not(:last-child)::after {
    top: 12px;
    left: calc(50% + 16px);
    width: calc(100% - 20px);
    height: 2px;
  }

  .step-item span {
    width: 24px;
    height: 24px;
    font-size: 13px;
  }

  .step-item.active span {
    border: 0;
  }

  .step-item.done span {
    font-size: 0;
  }

  .step-status-svg {
    width: 14px;
    height: 14px;
  }

  .step-status-svg--done {
    stroke-width: 2;
  }

  .step-item em {
    font-size: 14px;
    line-height: 1.1;
    white-space: normal;
  }

  .step-panel {
    gap: 22px;
  }

  .step-panel h1 {
    font-size: 25px;
    font-weight: 800;
  }

  .verify-panel p {
    margin: -8px 0 30px;
    font-size: 15px;
    line-height: 1.35;
    font-weight: 700;
  }

  .verify-panel .code-head {
    margin-top: 0;
  }

  .verify-panel .code-head strong {
    font-size: 18px;
    line-height: 1.2;
    font-weight: 800;
  }

  .verify-panel .code-send-button {
    min-width: 94px;
    border-radius: 24px;
    padding: 14px 18px;
    font-size: 18px;
    font-weight: 800;
  }

  .auth-tabs button {
    height: 52px;
    font-size: 20px;
    font-weight: 700;
  }

  .auth-tabs button.active::after {
    height: 3px;
  }

  .auth-field {
    min-height: 58px;
    border-radius: 18px;
    padding: 0 16px;
  }

  .auth-field input {
    font-size: 17px;
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
    width: 168px;
    margin: -10px 0 -4px 16px;
  }

  .agree-control,
  .auth-switch {
    font-size: 15px;
  }

  .agree-control span {
    width: 18px;
    height: 18px;
    border-radius: 5px;
  }

  .primary-button {
    min-height: 60px;
    margin-top: 12px;
    border-radius: 33px;
    font-size: 20px;
    font-weight: 800;
  }

  .register-page--captcha {
    padding: 14px 22px 28px;
  }

  .register-page--captcha .register-topbar {
    margin: -14px -22px 0;
    padding: 14px 22px 6px;
  }

  .register-page--captcha .icon-button {
    width: 38px;
    height: 38px;
  }

  .register-page--captcha .top-icon-svg {
    width: 22px;
    height: 22px;
  }

  .register-page--captcha .back-icon-svg {
    width: 23px;
    height: 23px;
  }

  .identity-panel {
    gap: 16px;
  }

  .identity-panel h1 {
    font-size: 24px;
    font-weight: 800;
  }

  .identity-panel h2 {
    margin: 12px 0 -8px;
    font-size: 22px;
    font-weight: 900;
  }

  .identity-panel p {
    margin: -8px 0 0;
    font-size: 15px;
    line-height: 1.35;
  }

  .identity-panel .auth-field {
    min-height: 58px;
    border-radius: 18px;
    padding: 0 18px;
  }

  .identity-panel .auth-field input {
    font-size: 17px;
  }

  .identity-panel .upload-grid {
    gap: 14px;
  }

  .identity-panel .upload-grid button {
    min-height: 112px;
    border-radius: 18px;
  }

  .identity-panel .upload-grid button:nth-child(3) {
    grid-column: 1 / 2;
  }

  .identity-panel .upload-grid strong {
    bottom: 16px;
    font-size: 18px;
  }

  .identity-panel .require-box {
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 10px;
    border-radius: 18px;
    padding: 12px 14px 11px;
  }

  .identity-panel .require-box span {
    gap: 6px;
    font-size: 12px;
  }

  .identity-panel .require-thumb {
    width: 62px;
    height: 38px;
    border-radius: 10px;
  }

  .primary-button--identity {
    min-height: 60px;
    margin-top: 18px;
    border-radius: 30px;
    font-size: 20px;
  }
}
</style>
