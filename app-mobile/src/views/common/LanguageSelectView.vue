<script setup lang="ts">
import { useRouter } from 'vue-router'

import AppIcon from '@/components/common/AppIcon.vue'
import { useI18n, type AppLocale } from '@/i18n'

const router = useRouter()
const { locale, setLocale, t } = useI18n()

const languageOptions: Array<{ code: string; label: string; flag: string; locale?: AppLocale }> = [
  { code: 'en-US', label: 'English', flag: '🇺🇸', locale: 'en-US' },
  { code: 'es-ES', label: 'Español', flag: '🇪🇸' },
  { code: 'ja-JP', label: '日本語', flag: '🇯🇵' },
  { code: 'ko-KR', label: '한국어', flag: '🇰🇷' },
  { code: 'ru-RU', label: 'Русский', flag: '🇷🇺' },
  { code: 'fr-FR', label: 'Français', flag: '🇫🇷' },
  { code: 'pt-BR', label: 'Português', flag: '🇧🇷' },
  { code: 'ms-MY', label: 'Malaysia', flag: '🇲🇾' },
  { code: 'zh-HK', label: '中文繁體', flag: '🇭🇰' },
  { code: 'zh-CN', label: '中文简体', flag: '🇨🇳', locale: 'zh-CN' },
]

function goBack() {
  if (window.history.length > 1) {
    router.back()
    return
  }
  router.replace('/profile')
}

function selectLanguage(option: { locale?: AppLocale }) {
  if (!option.locale) return
  setLocale(option.locale)
  goBack()
}
</script>

<template>
  <section class="language-page">
    <header class="language-header">
      <button
        type="button"
        class="icon-button"
        :aria-label="t('common.back')"
        @click="goBack"
      >
        <AppIcon name="back" class="back-icon-svg" />
      </button>
      <h1>语言选择</h1>
      <span />
    </header>

    <div class="language-list">
      <button
        v-for="item in languageOptions"
        :key="item.code"
        type="button"
        class="language-row"
        :class="{ 'language-row--active': item.locale === locale }"
        @click="selectLanguage(item)"
      >
        <span class="language-flag">{{ item.flag }}</span>
        <strong>{{ item.label }}</strong>
        <i />
      </button>
    </div>
  </section>
</template>

<style scoped>
.language-page {
  width: 100%;
  max-width: 100%;
  height: 100vh;
  height: 100dvh;
  min-height: 100dvh;
  margin: 0 auto;
  overflow-x: hidden;
  overflow-y: auto;
  overscroll-behavior-y: contain;
  -webkit-overflow-scrolling: touch;
  background: var(--page-bg);
  padding: 22px 22px 48px;
  color: var(--text);
}

.language-header {
  display: grid;
  grid-template-columns: 44px minmax(0, 1fr) 44px;
  align-items: center;
  min-width: 0;
}

.language-header h1 {
  overflow: hidden;
  margin: 0;
  font-size: 28px;
  font-weight: 800;
  line-height: 1.2;
  text-align: center;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.icon-button {
  display: inline-flex;
  width: 44px;
  height: 44px;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 50%;
  background: #252733;
  color: var(--text);
}

.chevron-left {
  width: 15px;
  height: 15px;
  border-left: 3px solid currentColor;
  border-bottom: 3px solid currentColor;
  transform: rotate(45deg);
}

.back-icon-svg {
  width: 24px;
  height: 24px;
  transform: translateX(-1px);
}

.language-list {
  display: grid;
  gap: 14px;
  margin-top: 72px;
}

.language-row {
  display: grid;
  min-height: 86px;
  grid-template-columns: 54px minmax(0, 1fr) 34px;
  align-items: center;
  gap: 14px;
  border: 0;
  border-radius: 20px;
  background: #1d1f2a;
  padding: 0 20px 0 24px;
  color: var(--text);
  text-align: left;
}

.language-flag {
  display: grid;
  width: 42px;
  height: 42px;
  place-items: center;
  overflow: hidden;
  border-radius: 50%;
  font-size: 38px;
  line-height: 1;
}

.language-row strong {
  min-width: 0;
  overflow: hidden;
  font-size: 22px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.language-row i {
  position: relative;
  display: block;
  width: 28px;
  height: 28px;
  border: 2px solid #3a3d49;
  border-radius: 50%;
}

.language-row--active i {
  border: 2px solid #00c313;
  box-shadow: none;
}

.language-row--active i::after {
  content: '';
  position: absolute;
  inset: 5px;
  border-radius: 50%;
  background: #00c313;
}

@media (min-width: 0) {
  .language-page {
    padding: 14px 22px 48px;
  }

  .language-header {
    grid-template-columns: 38px minmax(0, 1fr) 38px;
  }

  .language-header h1 {
    font-size: 25px;
    font-weight: 800;
  }

  .icon-button {
    width: 38px;
    height: 38px;
  }

  .chevron-left {
    width: 13px;
    height: 13px;
    border-left-width: 3px;
    border-bottom-width: 3px;
  }

  .back-icon-svg {
    width: 23px;
    height: 23px;
  }

  .language-list {
    gap: 11px;
    margin-top: 40px;
  }

  .language-row {
    min-height: 56px;
    grid-template-columns: 38px minmax(0, 1fr) 20px;
    gap: 10px;
    border-radius: 16px;
    padding: 0 18px;
  }

  .language-flag {
    width: 32px;
    height: 32px;
    font-size: 30px;
  }

  .language-row strong {
    font-size: 18px;
    font-weight: 500;
  }

  .language-row i {
    width: 20px;
    height: 20px;
    border-width: 1.5px;
  }

  .language-row--active i {
    border-width: 2px;
    box-shadow: none;
  }

  .language-row--active i::after {
    inset: 4px;
  }
}

@media (max-width: 390px) {
  .language-page {
    padding: 14px 18px 42px;
  }

  .language-list {
    gap: 10px;
    margin-top: 72px;
  }

  .language-row {
    min-height: 54px;
    border-radius: 15px;
    padding: 0 16px;
  }
}
</style>
