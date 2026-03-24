import { useAuthStore } from '@/stores'
import { useI18n } from 'vue-i18n'
const auth = useAuthStore()
const { t } = useI18n()
debugger /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {}
let __VLS_components
let __VLS_directives
const __VLS_0 = {}.ElCard
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ // @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({}))
const __VLS_2 = __VLS_1({}, ...__VLS_functionalComponentArgsRest(__VLS_1))
var __VLS_4 = {}
__VLS_3.slots.default
{
  const { header: __VLS_thisSlot } = __VLS_3.slots
  __VLS_ctx.t('route.home')
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({})
__VLS_ctx.t('home.welcome', {
  name: __VLS_ctx.auth.user?.nickname || __VLS_ctx.auth.user?.username || '-',
})
__VLS_asFunctionalElement(
  __VLS_intrinsicElements.div,
  __VLS_intrinsicElements.div,
)({
  ...{ style: {} },
})
__VLS_ctx.t('home.multiLangDesc')
var __VLS_3
var __VLS_dollars
const __VLS_self = (await import('vue')).defineComponent({
  setup() {
    return {
      auth: auth,
      t: t,
    }
  },
})
export default (await import('vue')).defineComponent({
  setup() {
    return {}
  },
}) /* PartiallyEnd: #4569/main.vue */
