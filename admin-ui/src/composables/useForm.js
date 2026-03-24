/**
 * 表单 Hook
 */
import { ref, reactive } from 'vue'
import { logger } from '@/utils/logger'
export function useForm(options) {
  const { initialData, validate } = options
  const form = reactive({ ...initialData })
  const formRef = ref()
  const errors = reactive({})
  const reset = () => {
    Object.keys(form).forEach((key) => {
      form[key] = initialData[key]
    })
    // Clear errors
    Object.keys(errors).forEach((key) => {
      delete errors[key]
    })
  }
  const getFormData = () => {
    return { ...form }
  }
  const setFormData = (data) => {
    Object.assign(form, data)
  }
  const clear = () => {
    Object.keys(form).forEach((key) => {
      form[key] = undefined
    })
  }
  const submit = async () => {
    try {
      // 使用 Element Plus 的表单验证
      if (formRef.value) {
        await formRef.value.validate()
      }
      // 自定义验证
      if (validate) {
        const isValid = await validate(getFormData())
        if (!isValid) {
          logger.warn('Form validation failed')
          return null
        }
      }
      return getFormData()
    } catch (error) {
      logger.error('Form submit error:', error)
      return null
    }
  }
  return {
    form: form,
    formRef,
    errors,
    reset,
    getFormData,
    setFormData,
    clear,
    submit,
  }
}
