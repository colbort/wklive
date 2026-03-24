import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { apiUploadAvatar } from '@/api/system/upload'
const { t } = useI18n()
const props = defineProps()
const emit = defineEmits()
const uploading = ref(false)
const form = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})
async function handleLogoChange(uploadFile) {
  if (!uploadFile.raw) return
  try {
    uploading.value = true
    const res = await apiUploadAvatar(uploadFile.raw)
    if (res.code === 0 || res.code === 200) {
      form.value.site_logo = res.data?.url || ''
      ElMessage.success(t('common.uploadSuccess') || 'Upload successful')
    } else {
      throw new Error(res.msg || 'Upload failed')
    }
  } catch (e) {
    ElMessage.error(e?.message || t('common.uploadFailed') || 'Upload failed')
  } finally {
    uploading.value = false
  }
}
debugger /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {}
let __VLS_components
let __VLS_directives
// CSS variable injection
// CSS variable injection end
__VLS_asFunctionalElement(
  __VLS_intrinsicElements.div,
  __VLS_intrinsicElements.div,
)({
  ...{ class: 'system-core-config' },
})
const __VLS_0 = {}.ElFormItem
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ // @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(
  __VLS_0,
  new __VLS_0({
    label: __VLS_ctx.t('system.configValue'),
    prop: 'site_name',
  }),
)
const __VLS_2 = __VLS_1(
  {
    label: __VLS_ctx.t('system.configValue'),
    prop: 'site_name',
  },
  ...__VLS_functionalComponentArgsRest(__VLS_1),
)
__VLS_3.slots.default
const __VLS_4 = {}.ElInput
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ // @ts-ignore
const __VLS_5 = __VLS_asFunctionalComponent(
  __VLS_4,
  new __VLS_4({
    modelValue: __VLS_ctx.form.site_name,
    placeholder: __VLS_ctx.t('system.siteName') || __VLS_ctx.t('common.pleaseEnter'),
  }),
)
const __VLS_6 = __VLS_5(
  {
    modelValue: __VLS_ctx.form.site_name,
    placeholder: __VLS_ctx.t('system.siteName') || __VLS_ctx.t('common.pleaseEnter'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_5),
)
var __VLS_3
const __VLS_8 = {}.ElFormItem
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ // @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(
  __VLS_8,
  new __VLS_8({
    label: __VLS_ctx.t('system.siteLogo'),
    prop: 'site_logo',
  }),
)
const __VLS_10 = __VLS_9(
  {
    label: __VLS_ctx.t('system.siteLogo'),
    prop: 'site_logo',
  },
  ...__VLS_functionalComponentArgsRest(__VLS_9),
)
__VLS_11.slots.default
__VLS_asFunctionalElement(
  __VLS_intrinsicElements.div,
  __VLS_intrinsicElements.div,
)({
  ...{ class: 'logo-upload-container' },
})
if (__VLS_ctx.form.site_logo) {
  const __VLS_12 = {}.ElImage
  /** @type {[typeof __VLS_components.ElImage, typeof __VLS_components.elImage, ]} */ // @ts-ignore
  const __VLS_13 = __VLS_asFunctionalComponent(
    __VLS_12,
    new __VLS_12({
      src: __VLS_ctx.form.site_logo,
      ...{ style: {} },
      previewTeleported: true,
    }),
  )
  const __VLS_14 = __VLS_13(
    {
      src: __VLS_ctx.form.site_logo,
      ...{ style: {} },
      previewTeleported: true,
    },
    ...__VLS_functionalComponentArgsRest(__VLS_13),
  )
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({})
const __VLS_16 = {}.ElUpload
/** @type {[typeof __VLS_components.ElUpload, typeof __VLS_components.elUpload, typeof __VLS_components.ElUpload, typeof __VLS_components.elUpload, ]} */ // @ts-ignore
const __VLS_17 = __VLS_asFunctionalComponent(
  __VLS_16,
  new __VLS_16({
    action: '#',
    autoUpload: false,
    onChange: __VLS_ctx.handleLogoChange,
    accept: 'image/*',
  }),
)
const __VLS_18 = __VLS_17(
  {
    action: '#',
    autoUpload: false,
    onChange: __VLS_ctx.handleLogoChange,
    accept: 'image/*',
  },
  ...__VLS_functionalComponentArgsRest(__VLS_17),
)
__VLS_19.slots.default
const __VLS_20 = {}.ElButton
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ // @ts-ignore
const __VLS_21 = __VLS_asFunctionalComponent(
  __VLS_20,
  new __VLS_20({
    type: 'primary',
    loading: __VLS_ctx.uploading,
  }),
)
const __VLS_22 = __VLS_21(
  {
    type: 'primary',
    loading: __VLS_ctx.uploading,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_21),
)
__VLS_23.slots.default
__VLS_ctx.uploading ? __VLS_ctx.t('common.uploading') : __VLS_ctx.t('common.upload')
var __VLS_23
var __VLS_19
__VLS_asFunctionalElement(
  __VLS_intrinsicElements.p,
  __VLS_intrinsicElements.p,
)({
  ...{ style: {} },
})
__VLS_ctx.t('common.uploadImageTip')
var __VLS_11
/** @type {__VLS_StyleScopedClasses['system-core-config']} */ /** @type {__VLS_StyleScopedClasses['logo-upload-container']} */ var __VLS_dollars
const __VLS_self = (await import('vue')).defineComponent({
  setup() {
    return {
      t: t,
      uploading: uploading,
      form: form,
      handleLogoChange: handleLogoChange,
    }
  },
  __typeEmits: {},
  __typeProps: {},
})
export default (await import('vue')).defineComponent({
  setup() {
    return {}
  },
  __typeEmits: {},
  __typeProps: {},
}) /* PartiallyEnd: #4569/main.vue */
