import { ref } from 'vue'

const isSupportPanelOpen = ref(false)

export function useSupportPanel() {
  function openSupportPanel() {
    isSupportPanelOpen.value = true
  }

  function closeSupportPanel() {
    isSupportPanelOpen.value = false
  }

  return {
    isSupportPanelOpen,
    openSupportPanel,
    closeSupportPanel,
  }
}
