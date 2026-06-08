<script setup lang="ts">
import { computed, ref } from 'vue'

import { useI18n } from '@/i18n'

const captchaImages = [
  new URL('../../../assets/checks/check1.webp', import.meta.url).href,
  new URL('../../../assets/checks/check2.webp', import.meta.url).href,
  new URL('../../../assets/checks/check3.webp', import.meta.url).href,
  new URL('../../../assets/checks/check4.webp', import.meta.url).href,
  new URL('../../../assets/checks/check5.webp', import.meta.url).href,
  new URL('../../../assets/checks/check6.webp', import.meta.url).href,
]

const emit = defineEmits<{
  success: []
}>()

const { t } = useI18n()
const progress = ref(0)
const passed = ref(false)
const failed = ref(false)
const targetAngle = ref(randomTargetAngle())
const captchaImage = ref(randomCaptchaImage())

const rotation = computed(() => normalizeAngle(targetAngle.value - progress.value * 3.6))
const targetProgress = computed(() => targetAngle.value / 3.6)
const statusText = computed(() => {
  if (passed.value) return t('captcha.passed')
  if (failed.value) return t('captcha.failed')
  return t('captcha.pending')
})

function randomTargetAngle() {
  return 75 + Math.round(Math.random() * 210)
}

function randomCaptchaImage() {
  return captchaImages[Math.floor(Math.random() * captchaImages.length)]
}

function normalizeAngle(value: number) {
  const normalized = ((value % 360) + 360) % 360
  return normalized > 180 ? normalized - 360 : normalized
}

function resetChallenge() {
  progress.value = 0
  passed.value = false
  failed.value = false
  targetAngle.value = randomTargetAngle()
  captchaImage.value = randomCaptchaImage()
}

function complete() {
  progress.value = targetProgress.value
  passed.value = true
  failed.value = false
  emit('success')
}

function handleInput(event: Event) {
  const value = Number((event.target as HTMLInputElement).value)
  progress.value = Number.isFinite(value) ? value : 0
  failed.value = false
}

function validateOnRelease() {
  if (passed.value) return

  if (Math.abs(rotation.value) <= 8) {
    complete()
    return
  }

  failed.value = true
  window.setTimeout(resetChallenge, 420)
}
</script>

<template>
  <main class="rotate-captcha">
    <h1>{{ t('captcha.title') }}</h1>
    <p>{{ t('captcha.hint') }}</p>
    <div class="rotate-captcha__image">
      <img :src="captchaImage" alt="" :style="{ transform: `rotate(${rotation}deg)` }" />
    </div>
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
  color: var(--text);
  font-size: 1.5rem;
  line-height: 1.15;
  font-weight: 900;
  letter-spacing: 0;
}

.rotate-captcha p {
  margin: 22px 0 0;
  color: #8f9098;
  font-size: 0.85rem;
  font-weight: 800;
}

.rotate-captcha__image {
  width: min(292px, 78vw);
  aspect-ratio: 1;
  margin: 38px 0 50px;
  overflow: hidden;
  border: 14px solid #191b26;
  border-radius: 50%;
  background: #191b26;
}

.rotate-captcha__image img {
  width: 100%;
  height: 100%;
  display: block;
  object-fit: cover;
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
  font-size: 1.7rem;
  transition: left 0.08s ease-out;
}

.rotate-captcha strong {
  margin-top: 38px;
  color: #8f9098;
  font-size: 1.1rem;
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
    font-size: 1.3rem;
  }

  .rotate-captcha p {
    font-size: 0.75rem;
  }

  .rotate-captcha__image {
    width: min(260px, 82vw);
    margin: 34px 0 46px;
  }
}

@media (min-width: 0) {
  .rotate-captcha {
    padding-top: 28px;
  }

  .rotate-captcha h1 {
    font-size: 1.5rem;
    font-weight: 800;
  }

  .rotate-captcha p {
    margin-top: 18px;
    font-size: 0.85rem;
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
    font-size: 1.4rem;
  }

  .rotate-captcha strong {
    margin-top: 36px;
    font-size: 0.9rem;
    font-weight: 700;
  }
}
</style>
