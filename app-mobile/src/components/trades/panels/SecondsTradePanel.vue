<script setup lang="ts">
import { computed } from 'vue'

import { useI18n } from '@/i18n'
import type { TradeSymbolSeconds } from '@/types/trade'

type SubmitSide = 'buy' | 'sell'

const props = defineProps<{
  secondsConfigs: TradeSymbolSeconds[]
  secondsDuration: number
  tradeQty: string
  settleAsset: string
  availableBalance: string
  tradeSymbolLoading: boolean
  isLoggedIn: boolean
  tradeAvailable: boolean
  tradeMessage: string
  tradeError: string
  submittingSide: SubmitSide | null
}>()

const emit = defineEmits<{
  (e: 'update:seconds-duration', value: number): void
  (e: 'update:tradeQty', value: string): void
  (e: 'submit-order', side: SubmitSide): void
}>()

const { t } = useI18n()
const sortedConfigs = computed(() =>
  [...props.secondsConfigs].sort((left, right) => left.durationSeconds - right.durationSeconds),
)
const selectedConfig = computed(
  () =>
    sortedConfigs.value.find((item) => item.durationSeconds === props.secondsDuration) ||
    sortedConfigs.value[0] ||
    null,
)
const submitDisabled = computed(
  () =>
    !props.isLoggedIn ||
    !props.tradeAvailable ||
    props.tradeSymbolLoading ||
    Boolean(props.submittingSide),
)
const unavailableText = computed(() => {
  if (!props.isLoggedIn) return ''
  if (props.tradeSymbolLoading) return t('trade.configLoading')
  if (!props.tradeAvailable) return t('trade.unavailable')
  return ''
})

function inputValue(event: Event) {
  return (event.target as HTMLInputElement).value
}

function payoutRate(value?: string) {
  const rate = Number(value)
  if (!Number.isFinite(rate)) return '--'
  return `${rate <= 1 ? rate * 100 : rate}%`
}
</script>

<template>
  <section class="seconds-trade-panel">
    <h3>{{ t('trade.expiryDuration') }}</h3>
    <div class="duration-grid">
      <button
        v-for="config in sortedConfigs"
        :key="config.id"
        type="button"
        :class="{ active: config.durationSeconds === secondsDuration }"
        @click="emit('update:seconds-duration', config.durationSeconds)"
      >
        <strong>{{ config.durationSeconds }}S</strong>
        <span>{{ t('trade.payoutRate') }} {{ payoutRate(config.payoutRate) }}</span>
      </button>
    </div>

    <h3>{{ t('trade.investmentAmount') }}</h3>
    <label class="trade-input">
      <input
        :value="tradeQty"
        inputmode="decimal"
        :placeholder="`${selectedConfig?.minStake || '--'} - ${selectedConfig?.maxStake || '--'} ${settleAsset}`"
        @input="emit('update:tradeQty', inputValue($event))"
      >
    </label>

    <div class="seconds-summary">
      <span>{{ t('trade.available') }}</span><strong>{{ availableBalance }} {{ settleAsset }}</strong>
      <span>{{ t('trade.expiryDuration') }}</span><strong>{{ secondsDuration || '--' }}S</strong> <span>{{ t('trade.payoutRate') }}</span><strong>{{ payoutRate(selectedConfig?.payoutRate) }}</strong>
    </div>

    <div class="seconds-actions">
      <button
        type="button"
        class="wide-action wide-action--buy"
        :disabled="submitDisabled || selectedConfig?.upEnabled !== 1"
        @click="emit('submit-order', 'buy')"
      >
        {{ submittingSide === 'buy' ? t('common.submitting') : t('trade.buyUp') }}
      </button>
      <button
        type="button"
        class="wide-action wide-action--sell"
        :disabled="submitDisabled || selectedConfig?.downEnabled !== 1"
        @click="emit('submit-order', 'sell')"
      >
        {{ submittingSide === 'sell' ? t('common.submitting') : t('trade.buyDown') }}
      </button>
    </div>

    <p v-if="tradeError || unavailableText" class="order-message order-message--error">
      {{ tradeError || unavailableText }}
    </p>
    <p v-else-if="tradeMessage" class="order-message">
      {{ tradeMessage }}
    </p>
  </section>
</template>

<style scoped>
.seconds-trade-panel h3 {
  margin: 18px 0 12px;
  font-size: 0.78rem;
  font-weight: 500;
}

button,
input {
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
}

.duration-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
}

.duration-grid button {
  display: flex;
  min-height: 58px;
  padding: 7px 4px;
  border: 1px solid var(--divider-visible);
  border-radius: 10px;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.duration-grid button.active {
  border-color: var(--border-strong);
  background: var(--selection-bg);
}

.duration-grid span {
  font-size: 0.52rem;
}
.duration-grid strong {
  font-size: 0.82rem;
}

.trade-input {
  display: flex;
  min-height: 48px;
  padding: 0 12px;
  border-radius: 12px;
  background: var(--control-bg);
  align-items: center;
}

.trade-input input {
  width: 100%;
  outline: 0;
}

.seconds-summary {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 8px 12px;
  margin: 16px 0;
  color: var(--muted);
  font-size: 0.68rem;
}

.seconds-summary strong {
  color: var(--text);
  font-weight: 500;
}

.seconds-actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.wide-action {
  min-height: 46px;
  margin-bottom: 18px;
  border-radius: 12px;
  color: var(--text);
  font-size: 0.75rem;
  font-weight: 700;
}

.wide-action--buy {
  background: var(--success);
}
.wide-action--sell {
  background: var(--danger-strong);
}
.wide-action:disabled {
  opacity: 0.52;
}

.order-message {
  color: var(--success);
  font-size: 0.65rem;
}
.order-message--error {
  color: var(--danger);
}
</style>
