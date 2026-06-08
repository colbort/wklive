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
    <p class="login-prompt__line">
      <RouterLink class="login-prompt__link" to="/login">
        {{ t('common.login') }}
      </RouterLink>
      <span>{{ t('common.or') }}</span>
      <RouterLink class="login-prompt__link" to="/register">
        {{ t('common.register') }}
      </RouterLink>
      <strong v-if="actionText">{{ actionText }}</strong>
    </p>
  </section>
</template>

<style scoped>
.login-prompt {
  display: flex;
  min-height: 180px;
  align-items: center;
  justify-content: center;
  color: var(--muted);
  text-align: center;
}

.login-prompt--compact {
  min-height: 140px;
}

.login-prompt__line {
  display: inline-flex;
  align-items: baseline;
  justify-content: center;
  min-width: 0;
  margin: 0;
  color: var(--muted);
  font-size: .75rem;
  font-weight: 700;
  line-height: 1.25;
  white-space: nowrap;
}

.login-prompt__link {
  color: var(--accent);
  text-decoration: none;
}

.login-prompt__line span {
  color: var(--muted);
}

.login-prompt__line strong {
  color: var(--muted);
  font: inherit;
}
</style>
