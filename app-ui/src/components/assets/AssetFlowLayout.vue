<script setup lang='ts'>
import { useRouter } from 'vue-router'

import AppIcon from '@/components/common/AppIcon.vue'
import { useI18n } from '@/i18n'

defineProps<{
  title: string
  rightText?: string
  rightTo?: string | Record<string, unknown>
  narrow?: boolean
}>()

const router = useRouter()
const { t } = useI18n()
</script>

<template>
  <section class="asset-flow-page" :class="{ 'asset-flow-page--narrow': narrow }">
    <header class="asset-flow-header" :class="{ 'asset-flow-header--plain-right': !rightText }">
      <button type="button" class="asset-flow-back" :aria-label="t('common.back')" @click="router.back()">
        <AppIcon name="back" class="asset-flow-back__icon" />
      </button>
      <h1>{{ title }}</h1>
      <button
        v-if="rightText"
        type="button"
        class="asset-flow-link"
        @click="rightTo ? router.push(rightTo) : undefined"
      >
        {{ rightText }}
      </button>
      <span v-else />
    </header>

    <main class="asset-flow-body">
      <slot />
    </main>
  </section>
</template>

<style scoped>
.asset-flow-page {
  width: 100%;
  max-width: 100%;
  min-height: 100dvh;
  overflow-x: hidden;
  padding: 10px 18px 28px;
  background: #0b0c15;
  color: #f8f8fb;
}

button {
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
}

.asset-flow-header {
  position: sticky;
  top: 0;
  z-index: 30;
  display: grid;
  grid-template-columns: 36px minmax(0, 1fr) minmax(0, 80px);
  align-items: center;
  min-height: 36px;
  min-width: 0;
  margin: -10px -18px 0;
  padding: 10px 18px 8px;
  background: #0b0c15;
}

.asset-flow-header--plain-right {
  grid-template-columns: 36px minmax(0, 1fr) 36px;
}

.asset-flow-back {
  display: grid;
  width: 36px;
  height: 36px;
  place-items: center;
  border-radius: 50%;
  background: #242633;
}

.asset-flow-back__icon {
  width: 23px;
  height: 23px;
  transform: translateX(-1px);
}

.asset-flow-header h1 {
  min-width: 0;
  overflow: hidden;
  margin: 0;
  font-size: 18px;
  text-align: center;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.asset-flow-link {
  color: #fff;
  font-size: 14px;
  font-weight: 700;
  text-align: right;
}

.asset-flow-body {
  width: 100%;
  min-width: 0;
  padding-top: 20px;
}

@media (max-width: 390px) {
  .asset-flow-page {
    padding: 8px 14px 24px;
  }

  .asset-flow-header {
    grid-template-columns: 34px minmax(0, 1fr) minmax(0, 70px);
    margin: -8px -14px 0;
    padding: 8px 14px 8px;
  }

  .asset-flow-header--plain-right {
    grid-template-columns: 34px minmax(0, 1fr) 34px;
  }

  .asset-flow-back {
    width: 34px;
    height: 34px;
  }

  .asset-flow-back__icon {
    width: 22px;
    height: 22px;
  }

  .asset-flow-header h1 {
    font-size: 17px;
  }

  .asset-flow-link {
    font-size: 13px;
  }
}

@media (min-width: 768px) {
  .asset-flow-page--narrow {
    max-width: 430px;
    min-height: calc(100dvh - 76px);
    margin: 0 auto;
  }
}
</style>
