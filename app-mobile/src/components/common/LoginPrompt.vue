<script setup lang="ts">
import { useI18n } from '@/i18n'

defineProps<{
  actionText?: string
  compact?: boolean
}>()

const { t } = useI18n()
</script>

<template>
  <section
    class="login-prompt"
    :class="{
      'login-prompt--compact': compact,
    }"
  >
    <p>{{ t('assets.loginFirst') }}</p>
    <div class="login-prompt__actions">
      <RouterLink class="login-prompt__link" to="/login">
        {{ t('common.login') }}
      </RouterLink>
      <i>/</i>
      <RouterLink class="login-prompt__link" to="/register">
        {{ t('common.register') }}
      </RouterLink>
      <strong v-if="actionText">{{ actionText }}</strong>
    </div>
  </section>
</template>

<style scoped>
.login-prompt {
  display: grid;
  align-content: start;
  min-height: 210px;
  color: #8f929d;
  text-align: center;
}

.login-prompt--compact {
  min-height: 160px;
}

.login-prompt p {
  display: grid;
  min-height: 92px;
  place-items: center;
  margin: 0;
  border-bottom: 1px solid #242633;
  color: #8f929d;
  font-size: 14px;
  font-weight: 600;
}

.login-prompt__actions {
  display: grid;
  justify-items: center;
  gap: 22px;
  padding-top: 22px;
  font-size: 14px;
  font-weight: 700;
}

.login-prompt__link {
  color: #02b904;
  text-decoration: none;
}

.login-prompt__actions i {
  color: #8f929d;
  font-style: normal;
}

.login-prompt__actions strong {
  color: #8f929d;
  font-size: 14px;
  font-weight: 700;
}
</style>
