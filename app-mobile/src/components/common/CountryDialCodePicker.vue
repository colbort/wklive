<script setup lang="ts">
import { computed, nextTick, ref } from 'vue'

import AppIcon from '@/components/common/AppIcon.vue'
import BottomDrawer from '@/components/common/BottomDrawer.vue'
import { countryDialCodes, type CountryDialCode } from '@/constants/countryDialCodes'
import { useI18n } from '@/i18n'

const props = defineProps<{
  modelValue: CountryDialCode
}>()

const emit = defineEmits<{
  (event: 'update:modelValue', value: CountryDialCode): void
}>()

const { t } = useI18n()
const open = ref(false)
const search = ref('')
const listRef = ref<HTMLDivElement | null>(null)

const filteredCountries = computed(() => {
  const keyword = search.value.trim().toLowerCase()
  if (!keyword) return countryDialCodes
  return countryDialCodes.filter((item) =>
    `${item.name} ${item.dialCode}`.toLowerCase().includes(keyword),
  )
})

function getCountryKey(country: CountryDialCode) {
  return `${country.name}-${country.dialCode}`
}

function isSelected(country: CountryDialCode) {
  return getCountryKey(country) === getCountryKey(props.modelValue)
}

function scrollSelectedCountry() {
  nextTick(() => {
    const selectedRow = listRef.value?.querySelector<HTMLButtonElement>('.country-row--selected')
    selectedRow?.scrollIntoView({ block: 'center' })
  })
}

function openSheet() {
  search.value = ''
  open.value = true
  scrollSelectedCountry()
}

function closeSheet() {
  open.value = false
}

function selectCountry(country: CountryDialCode) {
  emit('update:modelValue', country)
  open.value = false
}
</script>

<template>
  <button type="button" class="phone-prefix" @click="openSheet">
    <span>{{ modelValue.dialCode }}</span>
    <AppIcon name="chevron-down" />
  </button>

  <BottomDrawer
    v-model="open"
    title="区号选择"
    aria-label="区号选择"
    :close-label="t('common.close')"
    max-height="82dvh"
    :z-index="90"
    @close="closeSheet"
  >
    <div class="country-sheet-body">
      <label class="country-search">
        <AppIcon name="search" />
        <input v-model="search" placeholder="输入区号" inputmode="search">
      </label>
      <div ref="listRef" class="country-list">
        <button
          v-for="country in filteredCountries"
          :key="getCountryKey(country)"
          type="button"
          class="country-row"
          :class="{ 'country-row--selected': isSelected(country) }"
          @click="selectCountry(country)"
        >
          <span>{{ country.name }} {{ country.dialCode }}</span>
          <b v-if="isSelected(country)">✓</b>
        </button>
      </div>
    </div>
  </BottomDrawer>
</template>

<style scoped>
.phone-prefix {
  display: inline-flex;
  flex: none;
  align-items: center;
  gap: 8px;
  border: 0;
  background: transparent;
  color: var(--text);
  font-size: 0.9rem;
  font-weight: 600;
}

.phone-prefix svg {
  width: 16px;
  height: 16px;
  color: var(--muted);
}

.country-sheet-body {
  display: grid;
  grid-template-rows: auto minmax(0, 1fr);
  min-height: 0;
}

.country-search {
  display: grid;
  grid-template-columns: 24px minmax(0, 1fr);
  align-items: center;
  gap: 10px;
  min-height: 46px;
  border-radius: 16px;
  background: var(--panel-bg);
  padding: 0 16px;
}

.country-search svg {
  width: 22px;
  height: 22px;
}

.country-search input {
  min-width: 0;
  border: 0;
  outline: 0;
  background: transparent;
  color: var(--text);
  font-size: 0.8rem;
  font-weight: 600;
}

.country-search input::placeholder {
  color: var(--text-subtle);
}

.country-list {
  min-height: 0;
  max-height: calc(82dvh - 170px);
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
  margin-top: 26px;
  padding: 0 14px;
}

.country-row {
  display: flex;
  width: 100%;
  min-height: 56px;
  align-items: center;
  justify-content: space-between;
  border: 0;
  border-bottom: 1px solid var(--border-subtle);
  background: transparent;
  color: var(--text);
  font-size: 0.8rem;
  font-weight: 500;
  text-align: left;
}

.country-row--selected {
  color: var(--success-strong);
}

.country-row b {
  color: var(--success-strong);
  font-size: 0.9rem;
  font-weight: 900;
}

@media (min-width: 0) {
  .phone-prefix svg {
    width: 14px;
    height: 14px;
  }
}
</style>
