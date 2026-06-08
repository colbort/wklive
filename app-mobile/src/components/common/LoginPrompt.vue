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
  color: var(--muted);
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
  border-bottom: 1px solid var(--border-soft);
  color: var(--muted);
  font-size: 0.7rem;
  font-weight: 600;
}

.login-prompt__actions {
  display: grid;
  justify-items: center;
  gap: 22px;
  padding-top: 22px;
  font-size: 0.7rem;
  font-weight: 700;
}

.login-prompt__link {
  color: var(--accent);
  text-decoration: none;
}

.login-prompt__actions i {
  color: var(--muted);
  font-style: normal;
}

.login-prompt__actions strong {
  color: var(--muted);
  font-size: 0.7rem;
  font-weight: 700;
}
</style>
