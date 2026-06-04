<script setup lang="ts">
import { computed, ref } from 'vue'

import { useI18n } from '@/i18n'

const emit = defineEmits<{
  success: []
}>()

const { t } = useI18n()
const progress = ref(0)
const passed = ref(false)
const failed = ref(false)

const rotation = computed(() => 180 - progress.value * 1.8)
const statusText = computed(() => {
  if (passed.value) return t('captcha.passed')
  if (failed.value) return t('captcha.failed')
  return t('captcha.pending')
})

function complete() {
  progress.value = 100
  passed.value = true
  failed.value = false
  emit('success')
}

function handleInput(event: Event) {
  const value = Number((event.target as HTMLInputElement).value)
  progress.value = Number.isFinite(value) ? value : 0
  failed.value = false
  if (progress.value >= 98) {
    complete()
  }
}

function validateOnRelease() {
  if (passed.value) return

  if (progress.value >= 98) {
    complete()
    return
  }

  failed.value = true
  progress.value = 0
}
</script>

<template>
  <main class="rotate-captcha">
    <h1>{{ t('captcha.title') }}</h1>
    <p>{{ t('captcha.hint') }}</p>
    <div class="rotate-captcha__image" :style="{ transform: `rotate(${rotation}deg)` }" />
    <label
      class="rotate-captcha__slider"
      :class="{ 'is-failed': failed, 'is-passed': passed }"
      :style="{ '--captcha-progress': String(progress / 100) }"
    >
      <input
        v-model.number="progress"
        type="range"
        min="0"
        max="100"
        step="1"
        :aria-label="t('captcha.slider')"
        @input="handleInput"
        @change="validateOnRelease"
      />
      <span>→</span>
    </label>
    <strong :class="{ 'is-failed': failed, 'is-passed': passed }">{{ statusText }}</strong>
  </main>
</template>

<style scoped>
.rotate-captcha {
  display: grid;
  width: 100%;
  min-width: 0;
  justify-items: center;
  padding-top: 68px;
}

.rotate-captcha h1 {
  margin: 0;
  color: #fff;
  font-size: 30px;
  line-height: 1.15;
  font-weight: 900;
  letter-spacing: 0;
}

.rotate-captcha p {
  margin: 22px 0 0;
  color: #8f9098;
  font-size: 17px;
  font-weight: 800;
}

.rotate-captcha__image {
  width: min(292px, 78vw);
  aspect-ratio: 1;
  margin: 38px 0 50px;
  border: 14px solid #191b26;
  border-radius: 50%;
  background:
    linear-gradient(20deg, rgba(255, 255, 255, 0.28), transparent 34%),
    linear-gradient(145deg, #9ec8d4 0 24%, #5f7f8f 25% 36%, #d7d6c8 37% 48%, #436448 49% 68%, #d89cc0 69% 100%);
  transition: transform 0.12s ease-out;
}

.rotate-captcha__slider {
  position: relative;
  display: block;
  width: 100%;
  height: 66px;
  border-radius: 33px;
  background: #1b1d27;
  touch-action: pan-x;
  overflow: hidden;
}

.rotate-captcha__slider::before {
  content: '';
  position: absolute;
  inset: 0 auto 0 0;
  width: calc((100% - 66px) * var(--captcha-progress, 0) + 66px);
  border-radius: inherit;
  background: #00c313;
  opacity: 0.95;
}

.rotate-captcha__slider.is-failed::before {
  width: 66px;
  background: #ff473c;
}

.rotate-captcha__slider.is-passed::before {
  width: 100%;
}

.rotate-captcha__slider input {
  position: absolute;
  inset: 0;
  z-index: 2;
  width: 100%;
  height: 100%;
  opacity: 0;
  cursor: pointer;
}

.rotate-captcha__slider span {
  position: absolute;
  top: 0;
  left: calc((100% - 66px) * var(--captcha-progress, 0));
  z-index: 1;
  display: inline-flex;
  width: 66px;
  height: 66px;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: #fff;
  color: #111;
  font-size: 34px;
  transition: left 0.08s ease-out;
}

.rotate-captcha strong {
  margin-top: 38px;
  color: #8f9098;
  font-size: 22px;
  font-weight: 900;
}

.rotate-captcha strong.is-failed {
  color: #ff473c;
}

.rotate-captcha strong.is-passed {
  color: #00c313;
}

@media (max-width: 390px) {
  .rotate-captcha {
    padding-top: 58px;
  }

  .rotate-captcha h1 {
    font-size: 26px;
  }

  .rotate-captcha p {
    font-size: 15px;
  }

  .rotate-captcha__image {
    width: min(260px, 82vw);
    margin: 34px 0 46px;
  }
}

@media (max-width: 959px) {
  .rotate-captcha {
    padding-top: 28px;
  }

  .rotate-captcha h1 {
    font-size: 30px;
    font-weight: 800;
  }

  .rotate-captcha p {
    margin-top: 18px;
    font-size: 17px;
    font-weight: 700;
  }

  .rotate-captcha__image {
    width: min(230px, 62vw);
    margin: 38px 0 62px;
    border-width: 10px;
  }

  .rotate-captcha__slider {
    height: 48px;
    border-radius: 24px;
  }

  .rotate-captcha__slider::before {
    width: calc((100% - 48px) * var(--captcha-progress, 0) + 48px);
  }

  .rotate-captcha__slider.is-failed::before {
    width: 48px;
  }

  .rotate-captcha__slider span {
    left: calc((100% - 48px) * var(--captcha-progress, 0));
    width: 48px;
    height: 48px;
    font-size: 28px;
  }

  .rotate-captcha strong {
    margin-top: 36px;
    font-size: 18px;
    font-weight: 700;
  }
}
</style>
