<script setup lang="ts">
import { locale, setLocale } from '@/i18n'
import type { Locale } from '@/i18n'
import { useLanguagePanel } from '@/composables/useLanguagePanel'

const { isLanguagePanelOpen, closeLanguagePanel } = useLanguagePanel()

const languageOptions: Array<{
  label: string
  flag: string
  locale?: Locale
}> = [
  { label: 'English', flag: '🇺🇸', locale: 'en-US' },
  { label: 'Español', flag: '🇪🇸' },
  { label: '日本語', flag: '🇯🇵' },
  { label: '한국어', flag: '🇰🇷' },
  { label: 'Русский', flag: '🇷🇺' },
  { label: 'Français', flag: '🇫🇷' },
  { label: 'Português', flag: '🇧🇷' },
  { label: 'Malaysia', flag: '🇲🇾' },
  { label: '中文繁體', flag: '🇨🇳' },
  { label: '中文简体', flag: '🇨🇳', locale: 'zh-CN' },
]

function selectLanguage(nextLocale?: Locale) {
  if (nextLocale) {
    setLocale(nextLocale)
    closeLanguagePanel()
  }
}
</script>

<template>
  <div
    v-if="isLanguagePanelOpen"
    class="language-dialog"
    role="dialog"
    aria-modal="true"
    aria-labelledby="language-dialog-title"
    @click.self="closeLanguagePanel"
  >
    <section class="language-dialog__panel">
      <button
        class="language-dialog__close"
        type="button"
        aria-label="关闭"
        @click="closeLanguagePanel"
      />
      <h2 id="language-dialog-title">
        语言选择
      </h2>
      <label class="language-search">
        <input type="text" placeholder="请输入语言名称">
        <span />
      </label>
      <div class="language-list">
        <button
          v-for="item in languageOptions"
          :key="item.label"
          class="language-option"
          :class="{ 'language-option--active': item.locale === locale }"
          type="button"
          @click="selectLanguage(item.locale)"
        >
          <span class="language-option__flag">{{ item.flag }}</span>
          <span>{{ item.label }}</span>
          <i />
        </button>
      </div>
    </section>
  </div>
</template>

<style scoped>
.language-dialog {
  position: fixed;
  z-index: 80;
  inset: 0;
  display: grid;
  place-items: center;
  padding: var(--px-32);
  background: rgb(0 0 0 / 42%);
}

.language-dialog__panel {
  position: relative;
  width: min(var(--px-420), calc(100vw - var(--px-64)));
  max-height: calc(100vh - var(--px-64));
  padding: var(--px-26) var(--px-26) var(--px-28);
  border-radius: var(--px-28);
  background: rgb(35 36 44);
  box-shadow: 0 var(--px-24) var(--px-80) rgb(0 0 0 / 45%);
  font-size: 16px;
  overflow-y: auto;
}

.language-dialog__panel h2 {
  margin: 0;
  text-align: center;
  font-size: 22px;
  font-weight: var(--font-weight-600);
  line-height: 1.25;
}

.language-dialog__close {
  position: absolute;
  top: var(--px-30);
  right: var(--px-30);
  width: var(--px-24);
  height: var(--px-24);
  border: 0;
  background: transparent;
}

.language-dialog__close::before,
.language-dialog__close::after {
  position: absolute;
  top: 50%;
  left: 50%;
  width: var(--px-22);
  height: var(--px-2);
  border-radius: var(--px-999);
  background: rgb(255 255 255 / 52%);
  content: '';
}

.language-dialog__close::before {
  transform: translate(-50%, -50%) rotate(45deg);
}

.language-dialog__close::after {
  transform: translate(-50%, -50%) rotate(-45deg);
}

.language-search {
  position: relative;
  display: block;
  margin-top: var(--px-46);
}

.language-search input {
  width: 100%;
  height: var(--px-52);
  padding: 0 var(--px-58) 0 var(--px-24);
  border: 1px solid rgb(255 255 255 / 10%);
  border-radius: var(--px-999);
  outline: 0;
  background: rgb(255 255 255 / 6%);
  color: var(--text);
  font-size: 18px;
  font-weight: var(--font-weight-600);
}

.language-search input::placeholder {
  color: rgb(255 255 255 / 34%);
}

.language-search span {
  position: absolute;
  top: 50%;
  right: var(--px-26);
  width: var(--px-20);
  height: var(--px-20);
  border: var(--px-2) solid var(--white);
  border-radius: 50%;
  transform: translateY(-50%);
}

.language-search span::after {
  position: absolute;
  right: calc(var(--px-6) * -1);
  bottom: calc(var(--px-6) * -1);
  width: var(--px-10);
  height: var(--px-2);
  border-radius: var(--px-999);
  background: var(--white);
  content: '';
  transform: rotate(45deg);
  transform-origin: center;
}

.language-list {
  display: grid;
  gap: var(--px-8);
  margin-top: var(--px-24);
}

.language-option {
  display: grid;
  min-height: var(--px-40);
  align-items: center;
  grid-template-columns: var(--px-34) 1fr var(--px-24);
  gap: var(--px-14);
  padding: 0 var(--px-18) 0 var(--px-16);
  border: 0;
  border-radius: var(--px-14);
  background: transparent;
  color: var(--text);
  text-align: left;
  font-size: 18px;
  font-weight: var(--font-weight-600);
  transition:
    background 0.18s ease,
    color 0.18s ease;
}

.language-option:hover {
  background: rgb(16 19 30 / 72%);
}

.language-option__flag {
  display: block;
  width: var(--px-32);
  height: var(--px-32);
  border-radius: 50%;
  font-size: 26px;
  line-height: var(--px-32);
  overflow: hidden;
}

.language-option i {
  display: block;
  width: var(--px-24);
  height: var(--px-24);
  border: var(--px-2) solid rgb(255 255 255 / 8%);
  border-radius: 50%;
}

.language-option--active i {
  border-color: var(--accent);
  box-shadow: inset 0 0 0 var(--px-5) rgb(35 36 44);
  background: var(--accent);
}
</style>
