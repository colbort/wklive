<script setup lang="ts">
import { computed } from 'vue'
import { useSupportPanel } from '@/composables/useSupportPanel'

import languageIcon from '../../assets/home/link_4.svg'
import closeIcon from '../../assets/home/down_icon.svg'

const { isSupportPanelOpen, closeSupportPanel } = useSupportPanel()

const supportMessageTime = computed(() => {
  const now = new Date()
  const date = [
    now.getFullYear(),
    String(now.getMonth() + 1).padStart(2, '0'),
    String(now.getDate()).padStart(2, '0'),
  ].join('-')
  const time = [
    String(now.getHours()).padStart(2, '0'),
    String(now.getMinutes()).padStart(2, '0'),
    String(now.getSeconds()).padStart(2, '0'),
  ].join(':')

  return `${date} ${time}`
})
</script>

<template>
  <section v-if="isSupportPanelOpen" class="support-panel" aria-label="在线客服窗口">
    <header class="support-panel__header">
      <div class="support-panel__title">
        <span class="support-panel__badge">
          <svg viewBox="0 0 100 100" aria-hidden="true">
            <path d="M 46.193 15.213 c -19.605 1.933 -35.57 17.58 -37.888 37.145 c -1.41 11.919 2.161 22.959 8.84 31.424 l -2.407 12.422 c -0.256 1.312 1.007 2.407 2.27 1.973 l 11.918 -4.104 c 7.765 4.587 17.08 6.837 27.015 5.584 c 19.526 -2.467 35.055 -18.51 36.87 -38.103 c 2.477 -26.747 -19.85 -48.976 -46.618 -46.342 Z m 29.846 51.856 a 2.14 2.14 0 0 1 -2.141 2.14 h -7.943 a 2.14 2.14 0 0 1 -2.14 -2.14 V 55.516 c 0 -1.076 0.798 -1.924 1.825 -2.082 c -1.332 -6.502 -7.104 -11.415 -13.991 -11.415 c -6.897 0 -12.659 4.903 -13.99 11.415 c 1.025 0.158 1.824 1.006 1.824 2.082 v 11.553 a 2.14 2.14 0 0 1 -2.14 2.14 H 36 a 9.97 9.97 0 0 0 8.505 9.255 a 1.702 1.702 0 0 1 1.589 -1.085 h 6.58 c 0.948 0 1.718 0.77 1.718 1.717 v 3.581 c 0 0.947 -0.77 1.717 -1.717 1.717 h -6.61 c -0.79 0 -1.421 -0.543 -1.629 -1.273 c -7.153 -0.73 -12.767 -6.64 -13.073 -13.901 H 29.37 a 2.14 2.14 0 0 1 -2.14 -2.141 V 55.516 c 0 -1.184 0.956 -2.141 2.14 -2.141 h 0.188 c 1.45 -10.912 10.784 -19.367 22.081 -19.367 c 11.297 0 20.63 8.445 22.081 19.357 h 0.188 c 1.184 0 2.14 0.957 2.14 2.14 V 67.07 h -0.01 Z" />
          </svg>
        </span>
        <h2>在线客服</h2>
      </div>
      <div class="support-panel__actions">
        <button class="support-panel__round" type="button" aria-label="语言">
          <img :src="languageIcon" alt="">
        </button>
        <button
          class="support-panel__round support-panel__round--close"
          type="button"
          aria-label="收起客服"
          @click="closeSupportPanel"
        >
          <img :src="closeIcon" alt="Close">
        </button>
      </div>
    </header>
    <div class="support-panel__body">
      <time>{{ supportMessageTime }}</time>
      <div class="support-message">
        <span class="support-message__avatar">
          <svg viewBox="0 0 100 100" aria-hidden="true">
            <path d="M 46.193 15.213 c -19.605 1.933 -35.57 17.58 -37.888 37.145 c -1.41 11.919 2.161 22.959 8.84 31.424 l -2.407 12.422 c -0.256 1.312 1.007 2.407 2.27 1.973 l 11.918 -4.104 c 7.765 4.587 17.08 6.837 27.015 5.584 c 19.526 -2.467 35.055 -18.51 36.87 -38.103 c 2.477 -26.747 -19.85 -48.976 -46.618 -46.342 Z m 29.846 51.856 a 2.14 2.14 0 0 1 -2.141 2.14 h -7.943 a 2.14 2.14 0 0 1 -2.14 -2.14 V 55.516 c 0 -1.076 0.798 -1.924 1.825 -2.082 c -1.332 -6.502 -7.104 -11.415 -13.991 -11.415 c -6.897 0 -12.659 4.903 -13.99 11.415 c 1.025 0.158 1.824 1.006 1.824 2.082 v 11.553 a 2.14 2.14 0 0 1 -2.14 2.14 H 36 a 9.97 9.97 0 0 0 8.505 9.255 a 1.702 1.702 0 0 1 1.589 -1.085 h 6.58 c 0.948 0 1.718 0.77 1.718 1.717 v 3.581 c 0 0.947 -0.77 1.717 -1.717 1.717 h -6.61 c -0.79 0 -1.421 -0.543 -1.629 -1.273 c -7.153 -0.73 -12.767 -6.64 -13.073 -13.901 H 29.37 a 2.14 2.14 0 0 1 -2.14 -2.141 V 55.516 c 0 -1.184 0.956 -2.141 2.14 -2.141 h 0.188 c 1.45 -10.912 10.784 -19.367 22.081 -19.367 c 11.297 0 20.63 8.445 22.081 19.357 h 0.188 c 1.184 0 2.14 0.957 2.14 2.14 V 67.07 h -0.01 Z" />
          </svg>
        </span>
        <p>Hello, I'm your dedicated customer service representative. How can I help you?</p>
      </div>
    </div>
    <footer class="support-panel__footer">
      <button class="support-panel__attach" type="button" aria-label="上传附件">
        <svg viewBox="0 0 20 20" aria-hidden="true">
          <path d="m 16.364 13.536 l -1.414 -1.414 l 1.414 -1.415 a 5 5 0 0 0 -7.071 -7.07 L 7.878 5.05 L 6.464 3.636 l 1.414 -1.414 a 7 7 0 0 1 9.9 9.9 l -1.414 1.414 Z m -2.829 2.828 l -1.414 1.415 a 7 7 0 0 1 -9.9 -9.9 l 1.415 -1.414 L 5.05 7.879 L 3.636 9.293 a 5 5 0 0 0 7.07 7.071 l 1.415 -1.414 l 1.414 1.414 Z m -0.707 -10.606 l 1.415 1.414 l -7.072 7.07 l -1.414 -1.413 l 7.071 -7.071 Z" />
        </svg>
      </button>
      <input type="text" aria-label="输入消息">
      <button class="support-panel__send" type="button" aria-label="发送">
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path d="m 21.727 2.957 l -5.454 19.086 c -0.15 0.528 -0.475 0.552 -0.717 0.07 L 11 13 L 1.923 9.369 c -0.51 -0.204 -0.503 -0.51 0.034 -0.688 l 19.086 -6.362 c 0.529 -0.177 0.832 0.12 0.684 0.638 Z m -2.692 2.14 L 6.812 9.17 l 5.637 2.255 l 3.04 6.081 l 3.546 -12.41 Z" />
        </svg>
      </button>
    </footer>
  </section>
</template>

<style scoped>
.support-panel {
  position: fixed;
  z-index: 40;
  right: var(--px-20);
  bottom: var(--px-20);
  display: grid;
  width: min(var(--px-1200), calc(100vw - var(--px-84)));
  height: calc(100vh - var(--px-50));
  grid-template-rows: var(--px-110) 1fr var(--px-92);
  border-radius: var(--px-28);
  background: #0c0e17;
  box-shadow: 0 var(--px-28) var(--px-100) rgb(0 0 0 / 42%);
  overflow: hidden;
}

.support-panel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--px-30);
  background: rgb(29 32 44);
}

.support-panel__title {
  display: flex;
  align-items: center;
  gap: var(--px-22);
}

.support-panel__badge,
.support-message__avatar {
  display: grid;
  flex: 0 0 auto;
  place-items: center;
  border-radius: 50%;
}

.support-panel__badge {
  width: var(--px-56);
  height: var(--px-56);
}

.support-panel__badge svg,
.support-message__avatar svg {
  display: block;
  width: 100%;
  height: 100%;
  fill: var(--accent);
  pointer-events: none;
}

.support-panel__title h2 {
  margin: 0;
  font-size: var(--font-size-28);
  font-weight: var(--font-weight-800);
}

.support-panel__actions {
  display: flex;
  align-items: center;
  gap: var(--px-14);
}

.support-panel__round {
  position: relative;
  width: var(--px-50);
  height: var(--px-50);
  border: 0;
  border-radius: 50%;
  background: rgb(255 255 255 / 9%);
}

.support-panel__round img {
  display: block;
  width: var(--px-28);
  height: var(--px-28);
  object-fit: contain;
}

.support-panel__body {
  padding: var(--px-38) var(--px-58);
  overflow: auto;
}

.support-panel__body time {
  display: block;
  color: rgb(255 255 255 / 48%);
  text-align: center;
  font-size: var(--font-size-24);
  font-weight: var(--font-weight-800);
}

.support-message {
  display: flex;
  align-items: flex-start;
  gap: var(--px-20);
  margin-top: var(--px-28);
}

.support-message__avatar {
  width: var(--px-44);
  height: var(--px-44);
}

.support-message p {
  position: relative;
  max-width: var(--px-560);
  margin: 0;
  padding: var(--px-28) var(--px-32);
  border-radius: var(--px-14);
  background: var(--white);
  color: #050505;
  font-size: var(--font-size-24);
  line-height: 1.35;
}

.support-message p::before {
  position: absolute;
  top: var(--px-28);
  left: calc(var(--px-10) * -1);
  width: 0;
  height: 0;
  border-top: var(--px-10) solid transparent;
  border-right: var(--px-12) solid var(--white);
  border-bottom: var(--px-10) solid transparent;
  content: '';
}

.support-panel__footer {
  display: grid;
  align-items: center;
  grid-template-columns: var(--px-64) 1fr var(--px-64);
  gap: var(--px-18);
  padding: 0 var(--px-40);
  border-top: 1px solid rgb(255 255 255 / 8%);
  background: rgb(29 32 44);
}

.support-panel__footer input {
  width: 100%;
  height: var(--px-52);
  border: 0;
  border-radius: var(--px-20);
  outline: 0;
  background: rgb(255 255 255 / 6%);
  color: var(--text);
  font-size: var(--font-size-18);
}

.support-panel__attach,
.support-panel__send {
  position: relative;
  width: var(--px-44);
  height: var(--px-44);
  border: 0;
  background: transparent;
}

.support-panel__attach svg,
.support-panel__send svg {
  display: block;
  width: var(--px-34);
  height: var(--px-34);
  fill: var(--accent);
}
</style>
