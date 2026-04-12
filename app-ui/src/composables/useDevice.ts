import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

const DESKTOP_BREAKPOINT = 960

export function useDevice() {
  const width = ref(typeof window === 'undefined' ? 1280 : window.innerWidth)

  const updateWidth = () => {
    width.value = window.innerWidth
  }

  onMounted(() => {
    updateWidth()
    window.addEventListener('resize', updateWidth)
  })

  onBeforeUnmount(() => {
    window.removeEventListener('resize', updateWidth)
  })

  const isDesktop = computed(() => width.value >= DESKTOP_BREAKPOINT)
  const isMobile = computed(() => !isDesktop.value)

  return {
    width,
    isDesktop,
    isMobile,
  }
}
