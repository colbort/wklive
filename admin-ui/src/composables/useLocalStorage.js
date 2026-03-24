/**
 * 本地存储 Hook
 */
import { ref, watch } from 'vue'
export function useLocalStorage(key, initialValue, parse = true) {
  const data = ref(initialValue)
  // 初始化：从本地存储读取
  const storedValue = localStorage.getItem(key)
  if (storedValue) {
    try {
      data.value = parse ? JSON.parse(storedValue) : storedValue
    } catch (error) {
      console.error(`Failed to parse localStorage[${key}]`, error)
    }
  }
  // 监听变化，自动写入本地存储
  watch(
    () => data.value,
    (newValue) => {
      if (newValue === undefined) {
        localStorage.removeItem(key)
      } else {
        localStorage.setItem(key, parse ? JSON.stringify(newValue) : newValue)
      }
    },
    { deep: true },
  )
  const remove = () => {
    data.value = undefined
    localStorage.removeItem(key)
  }
  const clear = () => {
    data.value = initialValue
  }
  return {
    data,
    remove,
    clear,
  }
}
