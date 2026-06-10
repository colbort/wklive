import { ref } from 'vue'

const isLanguagePanelOpen = ref(false)

export function useLanguagePanel() {
  function openLanguagePanel() {
    isLanguagePanelOpen.value = true
  }

  function closeLanguagePanel() {
    isLanguagePanelOpen.value = false
  }

  return {
    isLanguagePanelOpen,
    openLanguagePanel,
    closeLanguagePanel,
  }
}
