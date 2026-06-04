<script setup lang="ts">
import { computed, nextTick, ref } from 'vue'

import AppIcon from '@/components/common/AppIcon.vue'
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

  <Teleport to="body">
    <Transition name="country-sheet">
      <div v-if="open" class="country-sheet-layer" @click.self="closeSheet">
        <section class="country-sheet" role="dialog" aria-modal="true" aria-label="区号选择">
          <i class="country-sheet__handle" />
          <button
            type="button"
            class="country-sheet__close"
            :aria-label="t('common.close')"
            @click="closeSheet"
          >
            <AppIcon name="close" />
          </button>
          <h2>区号选择</h2>
          <label class="country-search">
            <AppIcon name="search" />
            <input v-model="search" placeholder="输入区号" inputmode="search" />
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
        </section>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.phone-prefix {
  display: inline-flex;
  flex: none;
  align-items: center;
  gap: 8px;
  border: 0;
  background: transparent;
  color: #fff;
  font-size: 18px;
  font-weight: 600;
}

.phone-prefix svg {
  width: 16px;
  height: 16px;
  color: #9a9ca6;
}

.country-sheet-layer {
  position: fixed;
  inset: 0;
  z-index: 90;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  background: rgba(0, 0, 0, 0.58);
  backdrop-filter: blur(12px);
}

.country-sheet {
  position: relative;
  display: grid;
  grid-template-rows: auto auto minmax(0, 1fr);
  width: min(100%, 414px);
  max-height: min(82dvh, 720px);
  overflow: hidden;
  border-radius: 26px 26px 0 0;
  background: #262731;
  padding: 36px 28px 28px;
  color: #fff;
}

.country-sheet__handle {
  position: absolute;
  top: 13px;
  left: 50%;
  width: 38px;
  height: 4px;
  border-radius: 999px;
  background: #a2a3a9;
  transform: translateX(-50%);
}

.country-sheet__close {
  position: absolute;
  top: 20px;
  right: 22px;
  display: inline-flex;
  width: 34px;
  height: 34px;
  align-items: center;
  justify-content: center;
  border: 0;
  background: transparent;
  color: #fff;
}

.country-sheet__close svg {
  width: 26px;
  height: 26px;
}

.country-sheet h2 {
  margin: 0 0 22px;
  text-align: center;
  font-size: 21px;
  font-weight: 700;
}

.country-search {
  display: grid;
  grid-template-columns: 24px minmax(0, 1fr);
  align-items: center;
  gap: 10px;
  min-height: 46px;
  border-radius: 16px;
  background: #191b25;
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
  color: #fff;
  font-size: 16px;
  font-weight: 600;
}

.country-search input::placeholder {
  color: #8e9098;
}

.country-list {
  min-height: 0;
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
  border-bottom: 1px solid #1d1f28;
  background: transparent;
  color: #fff;
  font-size: 16px;
  font-weight: 500;
  text-align: left;
}

.country-row--selected {
  color: #00c819;
}

.country-row b {
  color: #00c819;
  font-size: 18px;
  font-weight: 900;
}

.country-sheet-enter-active,
.country-sheet-leave-active {
  transition: opacity 0.18s ease;
}

.country-sheet-enter-active .country-sheet,
.country-sheet-leave-active .country-sheet {
  transition: transform 0.18s ease;
}

.country-sheet-enter-from,
.country-sheet-leave-to {
  opacity: 0;
}

.country-sheet-enter-from .country-sheet,
.country-sheet-leave-to .country-sheet {
  transform: translateY(100%);
}

@media (max-width: 959px) {
  .phone-prefix svg {
    width: 14px;
    height: 14px;
  }
}
</style>
