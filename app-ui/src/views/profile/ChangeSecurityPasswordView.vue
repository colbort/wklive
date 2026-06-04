<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { apiChangeLoginPassword, apiChangePayPassword, apiSetPayPassword } from '@/api/userPrivate'
import { useI18n } from '@/i18n'

type FieldKey = 'oldPassword' | 'newPassword' | 'confirmPassword'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const submitting = ref(false)
const errorMessage = ref('')
const message = ref('')
const visible = ref<Record<FieldKey, boolean>>({
  oldPassword: false,
  newPassword: false,
  confirmPassword: false,
})

const isPayPassword = computed(() => route.name === 'security-pay-password')
const title = computed(() =>
  isPayPassword.value ? t('security.editTradePassword') : t('security.editLoginPassword'),
)
const fields = computed(() => {
  if (isPayPassword.value) {
    return [
      { key: 'newPassword' as const, placeholder: t('security.newTradePassword'), model: newPassword },
      { key: 'confirmPassword' as const, placeholder: t('security.confirmNewTradePassword'), model: confirmPassword },
    ]
  }

  return [
    { key: 'oldPassword' as const, placeholder: t('security.oldPassword'), model: oldPassword },
    { key: 'newPassword' as const, placeholder: t('security.newPassword'), model: newPassword },
    { key: 'confirmPassword' as const, placeholder: t('security.confirmNewPassword'), model: confirmPassword },
  ]
})
const passwordStrength = computed(() => {
  let level = 0
  if (newPassword.value.length >= 6) level += 1
  if (newPassword.value.length >= 8) level += 1
  if (/[A-Za-z]/.test(newPassword.value) && /\d/.test(newPassword.value)) level += 1
  if (/[^A-Za-z0-9]/.test(newPassword.value)) level += 1
  return level
})

function fieldValue(key: FieldKey) {
  if (key === 'oldPassword') return oldPassword.value
  if (key === 'newPassword') return newPassword.value
  return confirmPassword.value
}

function updateField(key: FieldKey, value: string) {
  if (key === 'oldPassword') oldPassword.value = value
  else if (key === 'newPassword') newPassword.value = value
  else confirmPassword.value = value
}

async function submitPassword() {
  if (submitting.value) return

  errorMessage.value = ''
  message.value = ''
  if (!isPayPassword.value && !oldPassword.value) {
    errorMessage.value = t('security.inputOldPassword')
    return
  }
  if (!newPassword.value || !confirmPassword.value) {
    errorMessage.value = t('security.inputNewPassword')
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    errorMessage.value = t('security.passwordMismatch')
    return
  }

  submitting.value = true
  try {
    const res = isPayPassword.value
      ? await apiSetPayPassword({
          password: newPassword.value,
          confirmPassword: confirmPassword.value,
        }).catch(() =>
          apiChangePayPassword({
            oldPassword: oldPassword.value,
            newPassword: newPassword.value,
            confirmPassword: confirmPassword.value,
          }),
        )
      : await apiChangeLoginPassword({
          oldPassword: oldPassword.value,
          newPassword: newPassword.value,
          confirmPassword: confirmPassword.value,
        })

    if (res.code !== 200) {
      errorMessage.value = res.msg || t('security.editFailed')
      return
    }
    message.value = t('security.editSuccess')
    window.setTimeout(() => router.replace('/profile/security'), 600)
  } catch (error) {
    console.warn('change password failed', error)
    errorMessage.value = t('security.editFailed')
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <section class="password-page">
    <header class="password-header">
      <button type="button" class="back-button" :aria-label="t('common.back')" @click="router.back()">
        <span />
      </button>
      <h1>{{ title }}</h1>
    </header>

    <form class="password-form" @submit.prevent="submitPassword">
      <template v-for="field in fields" :key="field.key">
        <label class="password-field">
          <input
            :value="fieldValue(field.key)"
            :type="visible[field.key] ? 'text' : 'password'"
            :placeholder="field.placeholder"
            autocomplete="new-password"
            @input="updateField(field.key, ($event.target as HTMLInputElement).value)"
          />
          <button
            type="button"
            class="eye-button"
            :aria-label="t('security.togglePassword')"
            @click="visible[field.key] = !visible[field.key]"
          >
            <span />
          </button>
        </label>
        <div v-if="field.key === 'newPassword'" class="strength-bars">
          <span v-for="index in 4" :key="index" :class="{ active: passwordStrength >= index }" />
        </div>
      </template>

      <p v-if="errorMessage" class="form-message form-message--error">{{ errorMessage }}</p>
      <p v-if="message" class="form-message">{{ message }}</p>

      <button type="submit" class="submit-button" :disabled="submitting">
        {{ submitting ? t('security.editing') : t('security.edit') }}
      </button>
    </form>
  </section>
</template>

<style scoped>
.password-page {
  width: 100%;
  max-width: 680px;
  min-height: 100dvh;
  margin: 0 auto;
  padding: 18px 36px;
  overflow-x: hidden;
  background: #0b0c15;
  color: #fff;
}

.password-header {
  display: grid;
  grid-template-columns: 48px minmax(0, 1fr) 48px;
  align-items: center;
  min-height: 48px;
  margin-bottom: 54px;
}

.password-header h1 {
  margin: 0;
  text-align: center;
  font-size: 24px;
  font-weight: 900;
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

.back-button span {
  width: 15px;
  height: 15px;
  border-left: 3px solid currentColor;
  border-bottom: 3px solid currentColor;
  transform: rotate(45deg);
}

.password-form {
  display: grid;
  gap: 20px;
}

.password-field {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 34px;
  align-items: center;
  min-height: 76px;
  padding: 0 24px;
  border-radius: 22px;
  background: #292a33;
}

.password-field input {
  min-width: 0;
  border: 0;
  outline: 0;
  background: transparent;
  color: #fff;
  font-size: 20px;
  font-weight: 800;
}

.password-field input::placeholder {
  color: #8d8e95;
}

.eye-button {
  display: inline-flex;
  width: 34px;
  height: 34px;
  align-items: center;
  justify-content: center;
  border: 0;
  background: transparent;
  color: #8d8e95;
}

.eye-button span {
  position: relative;
  width: 22px;
  height: 13px;
  border: 2px solid currentColor;
  border-radius: 999px;
}

.eye-button span::before {
  position: absolute;
  top: 3px;
  left: 8px;
  width: 4px;
  height: 4px;
  border-radius: 999px;
  background: currentColor;
  content: '';
}

.eye-button span::after {
  position: absolute;
  top: -7px;
  left: 10px;
  width: 2px;
  height: 27px;
  background: currentColor;
  content: '';
  transform: rotate(-45deg);
}

.strength-bars {
  display: grid;
  grid-template-columns: repeat(4, 40px);
  gap: 7px;
  margin: -10px 0 0 28px;
}

.strength-bars span {
  height: 4px;
  border-radius: 999px;
  background: #171923;
}

.strength-bars span.active {
  background: #00c313;
}

.form-message {
  margin: -6px 0 0;
  color: #00c313;
  font-size: 14px;
  font-weight: 700;
}

.form-message--error {
  color: #ff6868;
}

.submit-button {
  min-height: 76px;
  margin-top: 12px;
  border: 0;
  border-radius: 999px;
  background: #00bd09;
  color: #fff;
  font-size: 24px;
  font-weight: 900;
}

.submit-button:disabled {
  opacity: 0.72;
}

@media (max-width: 520px) {
  .password-page {
    padding: 16px 24px;
  }

  .password-header {
    grid-template-columns: 42px minmax(0, 1fr) 42px;
    min-height: 42px;
    margin-bottom: 48px;
  }

  .password-header h1 {
    font-size: 22px;
  }

  .back-button {
    width: 40px;
    height: 40px;
  }

  .password-form {
    gap: 18px;
  }

  .password-field {
    min-height: 68px;
    padding: 0 20px;
    border-radius: 20px;
  }

  .password-field input {
    font-size: 18px;
  }

  .strength-bars {
    grid-template-columns: repeat(4, 36px);
    margin-left: 24px;
  }

  .submit-button {
    min-height: 68px;
    font-size: 22px;
  }
}
</style>
