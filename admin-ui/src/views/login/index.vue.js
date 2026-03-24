import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores'
import { useI18n } from 'vue-i18n'
import { useForm, useLoading } from '@/composables'
const { t } = useI18n()
const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const { loading, withLoading } = useLoading()
// 初始化表单
const { form } = useForm({
  initialData: {
    username: '',
    password: '',
    googleCode: '',
  },
})
async function submit() {
  await withLoading(async () => {
    await auth.login({
      username: form.username,
      password: form.password,
      googleCode: form.googleCode || undefined,
    })
    await auth.fetchProfile()
    const redirect = route.query.redirect || '/home'
    router.replace(redirect)
  })
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
  ...{ class: 'wrap' },
})
const __VLS_0 = {}.ElCard
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ // @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(
  __VLS_0,
  new __VLS_0({
    ...{ class: 'card' },
  }),
)
const __VLS_2 = __VLS_1(
  {
    ...{ class: 'card' },
  },
  ...__VLS_functionalComponentArgsRest(__VLS_1),
)
__VLS_3.slots.default
{
  const { header: __VLS_thisSlot } = __VLS_3.slots
  __VLS_ctx.t('route.login')
}
const __VLS_4 = {}.ElForm
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ // @ts-ignore
const __VLS_5 = __VLS_asFunctionalComponent(
  __VLS_4,
  new __VLS_4({
    labelPosition: 'top',
  }),
)
const __VLS_6 = __VLS_5(
  {
    labelPosition: 'top',
  },
  ...__VLS_functionalComponentArgsRest(__VLS_5),
)
__VLS_7.slots.default
const __VLS_8 = {}.ElFormItem
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ // @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(
  __VLS_8,
  new __VLS_8({
    label: __VLS_ctx.t('auth.username'),
  }),
)
const __VLS_10 = __VLS_9(
  {
    label: __VLS_ctx.t('auth.username'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_9),
)
__VLS_11.slots.default
const __VLS_12 = {}.ElInput
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ // @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(
  __VLS_12,
  new __VLS_12({
    modelValue: __VLS_ctx.form.username,
    autocomplete: 'username',
  }),
)
const __VLS_14 = __VLS_13(
  {
    modelValue: __VLS_ctx.form.username,
    autocomplete: 'username',
  },
  ...__VLS_functionalComponentArgsRest(__VLS_13),
)
var __VLS_11
const __VLS_16 = {}.ElFormItem
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ // @ts-ignore
const __VLS_17 = __VLS_asFunctionalComponent(
  __VLS_16,
  new __VLS_16({
    label: __VLS_ctx.t('auth.password'),
  }),
)
const __VLS_18 = __VLS_17(
  {
    label: __VLS_ctx.t('auth.password'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_17),
)
__VLS_19.slots.default
const __VLS_20 = {}.ElInput
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ // @ts-ignore
const __VLS_21 = __VLS_asFunctionalComponent(
  __VLS_20,
  new __VLS_20({
    modelValue: __VLS_ctx.form.password,
    type: 'password',
    autocomplete: 'current-password',
    showPassword: true,
  }),
)
const __VLS_22 = __VLS_21(
  {
    modelValue: __VLS_ctx.form.password,
    type: 'password',
    autocomplete: 'current-password',
    showPassword: true,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_21),
)
var __VLS_19
const __VLS_24 = {}.ElFormItem
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ // @ts-ignore
const __VLS_25 = __VLS_asFunctionalComponent(
  __VLS_24,
  new __VLS_24({
    label: __VLS_ctx.t('auth.googleCode'),
  }),
)
const __VLS_26 = __VLS_25(
  {
    label: __VLS_ctx.t('auth.googleCode'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_25),
)
__VLS_27.slots.default
const __VLS_28 = {}.ElInput
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ // @ts-ignore
const __VLS_29 = __VLS_asFunctionalComponent(
  __VLS_28,
  new __VLS_28({
    modelValue: __VLS_ctx.form.googleCode,
  }),
)
const __VLS_30 = __VLS_29(
  {
    modelValue: __VLS_ctx.form.googleCode,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_29),
)
var __VLS_27
const __VLS_32 = {}.ElButton
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ // @ts-ignore
const __VLS_33 = __VLS_asFunctionalComponent(
  __VLS_32,
  new __VLS_32({
    ...{ onClick: {} },
    type: 'primary',
    loading: __VLS_ctx.loading,
    ...{ style: {} },
  }),
)
const __VLS_34 = __VLS_33(
  {
    ...{ onClick: {} },
    type: 'primary',
    loading: __VLS_ctx.loading,
    ...{ style: {} },
  },
  ...__VLS_functionalComponentArgsRest(__VLS_33),
)
let __VLS_36
let __VLS_37
let __VLS_38
const __VLS_39 = {
  onClick: __VLS_ctx.submit,
}
__VLS_35.slots.default
__VLS_ctx.t('auth.submit')
var __VLS_35
var __VLS_7
var __VLS_3
/** @type {__VLS_StyleScopedClasses['wrap']} */ /** @type {__VLS_StyleScopedClasses['card']} */ var __VLS_dollars
const __VLS_self = (await import('vue')).defineComponent({
  setup() {
    return {
      t: t,
      loading: loading,
      form: form,
      submit: submit,
    }
  },
})
export default (await import('vue')).defineComponent({
  setup() {
    return {}
  },
}) /* PartiallyEnd: #4569/main.vue */
