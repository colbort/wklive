import { useI18n } from 'vue-i18n'
const { t } = useI18n()
debugger /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {}
let __VLS_components
let __VLS_directives
const __VLS_0 = {}.ElResult
/** @type {[typeof __VLS_components.ElResult, typeof __VLS_components.elResult, typeof __VLS_components.ElResult, typeof __VLS_components.elResult, ]} */ // @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(
  __VLS_0,
  new __VLS_0({
    icon: 'warning',
    title: '404',
    subTitle: __VLS_ctx.t('error.notFound'),
  }),
)
const __VLS_2 = __VLS_1(
  {
    icon: 'warning',
    title: '404',
    subTitle: __VLS_ctx.t('error.notFound'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_1),
)
var __VLS_4 = {}
__VLS_3.slots.default
{
  const { extra: __VLS_thisSlot } = __VLS_3.slots
  const __VLS_5 = {}.ElButton
  /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ // @ts-ignore
  const __VLS_6 = __VLS_asFunctionalComponent(
    __VLS_5,
    new __VLS_5({
      ...{ onClick: {} },
      type: 'primary',
    }),
  )
  const __VLS_7 = __VLS_6(
    {
      ...{ onClick: {} },
      type: 'primary',
    },
    ...__VLS_functionalComponentArgsRest(__VLS_6),
  )
  let __VLS_9
  let __VLS_10
  let __VLS_11
  const __VLS_12 = {
    onClick: (...[$event]) => {
      __VLS_ctx.$router.push('/')
    },
  }
  __VLS_8.slots.default
  __VLS_ctx.t('common.goHome')
  var __VLS_8
}
var __VLS_3
var __VLS_dollars
const __VLS_self = (await import('vue')).defineComponent({
  setup() {
    return {
      t: t,
    }
  },
})
export default (await import('vue')).defineComponent({
  setup() {
    return {}
  },
}) /* PartiallyEnd: #4569/main.vue */
