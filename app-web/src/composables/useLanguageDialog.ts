import { ref } from 'vue'

const isLanguageDialogOpen = ref(false)

export function useLanguageDialog() {
  function openLanguageDialog() {
    isLanguageDialogOpen.value = true
  }

  function closeLanguageDialog() {
    isLanguageDialogOpen.value = false
  }

  return {
    isLanguageDialogOpen,
    openLanguageDialog,
    closeLanguageDialog,
  }
}
