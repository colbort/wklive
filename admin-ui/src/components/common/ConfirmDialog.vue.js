import { ref, watch } from 'vue'
const props = withDefaults(defineProps(), {
  title: 'Dialog',
  confirmText: 'Confirm',
  width: '500px',
  loading: false,
})
const emit = defineEmits()
const visible = ref(props.modelValue)
watch(
  () => props.modelValue,
  (val) => {
    visible.value = val
  },
)
watch(visible, (val) => {
  emit('update:modelValue', val)
})
const handleConfirm = () => {
  emit('confirm')
}
debugger /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_withDefaultsArg = (function (t) {
  return t
})({
  title: 'Dialog',
  confirmText: 'Confirm',
  width: '500px',
  loading: false,
})
const __VLS_ctx = {}
let __VLS_components
let __VLS_directives
const __VLS_0 = {}.ElDialog
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ // @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(
  __VLS_0,
  new __VLS_0({
    ...{ onClose: {} },
    modelValue: __VLS_ctx.visible,
    title: __VLS_ctx.title,
    width: __VLS_ctx.width,
    closeOnClickModal: false,
    closeOnPressEscape: false,
  }),
)
const __VLS_2 = __VLS_1(
  {
    ...{ onClose: {} },
    modelValue: __VLS_ctx.visible,
    title: __VLS_ctx.title,
    width: __VLS_ctx.width,
    closeOnClickModal: false,
    closeOnPressEscape: false,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_1),
)
let __VLS_4
let __VLS_5
let __VLS_6
const __VLS_7 = {
  onClose: (...[$event]) => {
    __VLS_ctx.$emit('close')
  },
}
var __VLS_8 = {}
__VLS_3.slots.default
var __VLS_9 = {}
{
  const { footer: __VLS_thisSlot } = __VLS_3.slots
  const __VLS_11 = {}.ElButton
  /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ // @ts-ignore
  const __VLS_12 = __VLS_asFunctionalComponent(
    __VLS_11,
    new __VLS_11({
      ...{ onClick: {} },
    }),
  )
  const __VLS_13 = __VLS_12(
    {
      ...{ onClick: {} },
    },
    ...__VLS_functionalComponentArgsRest(__VLS_12),
  )
  let __VLS_15
  let __VLS_16
  let __VLS_17
  const __VLS_18 = {
    onClick: (...[$event]) => {
      __VLS_ctx.visible = false
    },
  }
  __VLS_14.slots.default
  var __VLS_14
  const __VLS_19 = {}.ElButton
  /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ // @ts-ignore
  const __VLS_20 = __VLS_asFunctionalComponent(
    __VLS_19,
    new __VLS_19({
      ...{ onClick: {} },
      type: 'primary',
      loading: __VLS_ctx.loading,
    }),
  )
  const __VLS_21 = __VLS_20(
    {
      ...{ onClick: {} },
      type: 'primary',
      loading: __VLS_ctx.loading,
    },
    ...__VLS_functionalComponentArgsRest(__VLS_20),
  )
  let __VLS_23
  let __VLS_24
  let __VLS_25
  const __VLS_26 = {
    onClick: __VLS_ctx.handleConfirm,
  }
  __VLS_22.slots.default
  __VLS_ctx.confirmText
  var __VLS_22
}
var __VLS_3
// @ts-ignore
var __VLS_10 = __VLS_9
var __VLS_dollars
const __VLS_self = (await import('vue')).defineComponent({
  setup() {
    return {
      visible: visible,
      handleConfirm: handleConfirm,
    }
  },
  __typeEmits: {},
  __typeProps: {},
  props: {},
})
const __VLS_component = (await import('vue')).defineComponent({
  setup() {
    return {}
  },
  __typeEmits: {},
  __typeProps: {},
  props: {},
})
export default {} /* PartiallyEnd: #4569/main.vue */
