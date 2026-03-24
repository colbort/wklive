import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { usePagination } from '@/composables/usePagination'
import { useLoading } from '@/composables/useLoading'
import { useForm } from '@/composables/useForm'
import { logService } from '@/services'
const { t } = useI18n()
// Pagination and list
const {
  pagination,
  updatePagination,
  nextPage: paginationNextPage,
  prevPage: paginationPrevPage,
} = usePagination(10)
const { loading, withLoading } = useLoading()
// Query form
const { form: queryForm } = useForm({
  initialData: {
    username: '',
    success: undefined,
  },
})
const list_ref = ref([])
async function fetchList() {
  await withLoading(async () => {
    try {
      const res = await logService.getLoginLogs({
        username: queryForm.username || undefined,
        success: queryForm.success,
        cursor: pagination.cursor,
        limit: pagination.limit,
      })
      if (res.code !== 0 && res.code !== 200) throw new Error(res.msg)
      list_ref.value = res.data || []
      updatePagination(
        res.total || 0,
        res.hasNext || false,
        res.hasPrev || false,
        res.nextCursor || null,
        res.prevCursor || null,
      )
      console.log('res.hasPrev', res.hasPrev)
      console.log('res.hasNext', res.hasNext)
      console.log('pagination.hasPrev', pagination.hasPrev)
      console.log('pagination.hasNext', pagination.hasNext)
    } catch (e) {
      ElMessage.error(e?.message || t('common.loadFailed'))
    }
  })
}
function onSearch() {
  pagination.cursor = null
  pagination.hasPrev = false
  fetchList()
}
function onReset() {
  queryForm.username = ''
  queryForm.success = undefined
  pagination.cursor = null
  pagination.hasPrev = false
  fetchList()
}
function nextPage() {
  paginationNextPage()
  fetchList()
}
function prevPage() {
  paginationPrevPage()
  fetchList()
}
onMounted(() => {
  fetchList()
})
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
  __VLS_ctx.t('system.loginLog')
}
const __VLS_5 = {}.ElForm
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ // @ts-ignore
const __VLS_6 = __VLS_asFunctionalComponent(
  __VLS_5,
  new __VLS_5({
    model: __VLS_ctx.queryForm,
    inline: true,
    ...{ style: {} },
  }),
)
const __VLS_7 = __VLS_6(
  {
    model: __VLS_ctx.queryForm,
    inline: true,
    ...{ style: {} },
  },
  ...__VLS_functionalComponentArgsRest(__VLS_6),
)
__VLS_8.slots.default
const __VLS_9 = {}.ElFormItem
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ // @ts-ignore
const __VLS_10 = __VLS_asFunctionalComponent(
  __VLS_9,
  new __VLS_9({
    label: __VLS_ctx.t('common.username'),
  }),
)
const __VLS_11 = __VLS_10(
  {
    label: __VLS_ctx.t('common.username'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_10),
)
__VLS_12.slots.default
const __VLS_13 = {}.ElInput
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ // @ts-ignore
const __VLS_14 = __VLS_asFunctionalComponent(
  __VLS_13,
  new __VLS_13({
    modelValue: __VLS_ctx.queryForm.username,
    placeholder: __VLS_ctx.t('common.pleaseInputUsername'),
    clearable: true,
    ...{ style: {} },
  }),
)
const __VLS_15 = __VLS_14(
  {
    modelValue: __VLS_ctx.queryForm.username,
    placeholder: __VLS_ctx.t('common.pleaseInputUsername'),
    clearable: true,
    ...{ style: {} },
  },
  ...__VLS_functionalComponentArgsRest(__VLS_14),
)
var __VLS_12
const __VLS_17 = {}.ElFormItem
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ // @ts-ignore
const __VLS_18 = __VLS_asFunctionalComponent(
  __VLS_17,
  new __VLS_17({
    label: __VLS_ctx.t('common.result'),
  }),
)
const __VLS_19 = __VLS_18(
  {
    label: __VLS_ctx.t('common.result'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_18),
)
__VLS_20.slots.default
const __VLS_21 = {}.ElSelect
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ // @ts-ignore
const __VLS_22 = __VLS_asFunctionalComponent(
  __VLS_21,
  new __VLS_21({
    modelValue: __VLS_ctx.queryForm.success,
    placeholder: __VLS_ctx.t('common.pleaseSelectResult'),
    clearable: true,
    ...{ style: {} },
  }),
)
const __VLS_23 = __VLS_22(
  {
    modelValue: __VLS_ctx.queryForm.success,
    placeholder: __VLS_ctx.t('common.pleaseSelectResult'),
    clearable: true,
    ...{ style: {} },
  },
  ...__VLS_functionalComponentArgsRest(__VLS_22),
)
__VLS_24.slots.default
const __VLS_25 = {}.ElOption
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ // @ts-ignore
const __VLS_26 = __VLS_asFunctionalComponent(
  __VLS_25,
  new __VLS_25({
    label: __VLS_ctx.t('common.success'),
    value: 1,
  }),
)
const __VLS_27 = __VLS_26(
  {
    label: __VLS_ctx.t('common.success'),
    value: 1,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_26),
)
const __VLS_29 = {}.ElOption
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ // @ts-ignore
const __VLS_30 = __VLS_asFunctionalComponent(
  __VLS_29,
  new __VLS_29({
    label: __VLS_ctx.t('common.failed'),
    value: 0,
  }),
)
const __VLS_31 = __VLS_30(
  {
    label: __VLS_ctx.t('common.failed'),
    value: 0,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_30),
)
var __VLS_24
var __VLS_20
const __VLS_33 = {}.ElFormItem
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ // @ts-ignore
const __VLS_34 = __VLS_asFunctionalComponent(__VLS_33, new __VLS_33({}))
const __VLS_35 = __VLS_34({}, ...__VLS_functionalComponentArgsRest(__VLS_34))
__VLS_36.slots.default
const __VLS_37 = {}.ElButton
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ // @ts-ignore
const __VLS_38 = __VLS_asFunctionalComponent(
  __VLS_37,
  new __VLS_37({
    ...{ onClick: {} },
    type: 'primary',
  }),
)
const __VLS_39 = __VLS_38(
  {
    ...{ onClick: {} },
    type: 'primary',
  },
  ...__VLS_functionalComponentArgsRest(__VLS_38),
)
let __VLS_41
let __VLS_42
let __VLS_43
const __VLS_44 = {
  onClick: __VLS_ctx.onSearch,
}
__VLS_40.slots.default
__VLS_ctx.t('common.search')
var __VLS_40
const __VLS_45 = {}.ElButton
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ // @ts-ignore
const __VLS_46 = __VLS_asFunctionalComponent(
  __VLS_45,
  new __VLS_45({
    ...{ onClick: {} },
  }),
)
const __VLS_47 = __VLS_46(
  {
    ...{ onClick: {} },
  },
  ...__VLS_functionalComponentArgsRest(__VLS_46),
)
let __VLS_49
let __VLS_50
let __VLS_51
const __VLS_52 = {
  onClick: __VLS_ctx.onReset,
}
__VLS_48.slots.default
__VLS_ctx.t('common.reset')
var __VLS_48
var __VLS_36
var __VLS_8
const __VLS_53 = {}.ElTable
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ // @ts-ignore
const __VLS_54 = __VLS_asFunctionalComponent(
  __VLS_53,
  new __VLS_53({
    data: __VLS_ctx.list_ref,
    rowKey: 'id',
    ...{ style: {} },
  }),
)
const __VLS_55 = __VLS_54(
  {
    data: __VLS_ctx.list_ref,
    rowKey: 'id',
    ...{ style: {} },
  },
  ...__VLS_functionalComponentArgsRest(__VLS_54),
)
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(
  null,
  { ...__VLS_directiveBindingRestFields, value: __VLS_ctx.loading },
  null,
  null,
)
__VLS_56.slots.default
const __VLS_57 = {}.ElTableColumn
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ // @ts-ignore
const __VLS_58 = __VLS_asFunctionalComponent(
  __VLS_57,
  new __VLS_57({
    prop: 'id',
    label: __VLS_ctx.t('common.id'),
    width: '70',
  }),
)
const __VLS_59 = __VLS_58(
  {
    prop: 'id',
    label: __VLS_ctx.t('common.id'),
    width: '70',
  },
  ...__VLS_functionalComponentArgsRest(__VLS_58),
)
const __VLS_61 = {}.ElTableColumn
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ // @ts-ignore
const __VLS_62 = __VLS_asFunctionalComponent(
  __VLS_61,
  new __VLS_61({
    prop: 'username',
    label: __VLS_ctx.t('common.username'),
    minWidth: '120',
  }),
)
const __VLS_63 = __VLS_62(
  {
    prop: 'username',
    label: __VLS_ctx.t('common.username'),
    minWidth: '120',
  },
  ...__VLS_functionalComponentArgsRest(__VLS_62),
)
const __VLS_65 = {}.ElTableColumn
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ // @ts-ignore
const __VLS_66 = __VLS_asFunctionalComponent(
  __VLS_65,
  new __VLS_65({
    prop: 'ip',
    label: __VLS_ctx.t('common.loginIP'),
    minWidth: '130',
  }),
)
const __VLS_67 = __VLS_66(
  {
    prop: 'ip',
    label: __VLS_ctx.t('common.loginIP'),
    minWidth: '130',
  },
  ...__VLS_functionalComponentArgsRest(__VLS_66),
)
const __VLS_69 = {}.ElTableColumn
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ // @ts-ignore
const __VLS_70 = __VLS_asFunctionalComponent(
  __VLS_69,
  new __VLS_69({
    prop: 'ua',
    label: __VLS_ctx.t('common.userAgent'),
    minWidth: '180',
    showOverflowTooltip: true,
  }),
)
const __VLS_71 = __VLS_70(
  {
    prop: 'ua',
    label: __VLS_ctx.t('common.userAgent'),
    minWidth: '180',
    showOverflowTooltip: true,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_70),
)
const __VLS_73 = {}.ElTableColumn
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ // @ts-ignore
const __VLS_74 = __VLS_asFunctionalComponent(
  __VLS_73,
  new __VLS_73({
    prop: 'success',
    label: __VLS_ctx.t('common.result'),
    width: '100',
  }),
)
const __VLS_75 = __VLS_74(
  {
    prop: 'success',
    label: __VLS_ctx.t('common.result'),
    width: '100',
  },
  ...__VLS_functionalComponentArgsRest(__VLS_74),
)
__VLS_76.slots.default
{
  const { default: __VLS_thisSlot } = __VLS_76.slots
  const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot)
  const __VLS_77 = {}.ElTag
  /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ // @ts-ignore
  const __VLS_78 = __VLS_asFunctionalComponent(
    __VLS_77,
    new __VLS_77({
      type: row.success === 1 ? 'success' : 'danger',
    }),
  )
  const __VLS_79 = __VLS_78(
    {
      type: row.success === 1 ? 'success' : 'danger',
    },
    ...__VLS_functionalComponentArgsRest(__VLS_78),
  )
  __VLS_80.slots.default
  row.success === 1 ? __VLS_ctx.t('common.loginSuccess') : __VLS_ctx.t('common.loginFailed')
  var __VLS_80
}
var __VLS_76
const __VLS_81 = {}.ElTableColumn
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ // @ts-ignore
const __VLS_82 = __VLS_asFunctionalComponent(
  __VLS_81,
  new __VLS_81({
    prop: 'msg',
    label: __VLS_ctx.t('common.failureReason'),
    minWidth: '180',
    showOverflowTooltip: true,
  }),
)
const __VLS_83 = __VLS_82(
  {
    prop: 'msg',
    label: __VLS_ctx.t('common.failureReason'),
    minWidth: '180',
    showOverflowTooltip: true,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_82),
)
__VLS_84.slots.default
{
  const { default: __VLS_thisSlot } = __VLS_84.slots
  const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot)
  if (row.success !== 1) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({})
    row.msg
  } else {
    __VLS_asFunctionalElement(
      __VLS_intrinsicElements.span,
      __VLS_intrinsicElements.span,
    )({
      ...{ style: {} },
    })
  }
}
var __VLS_84
const __VLS_85 = {}.ElTableColumn
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ // @ts-ignore
const __VLS_86 = __VLS_asFunctionalComponent(
  __VLS_85,
  new __VLS_85({
    prop: 'loginAt',
    label: __VLS_ctx.t('common.loginTime'),
    minWidth: '170',
  }),
)
const __VLS_87 = __VLS_86(
  {
    prop: 'loginAt',
    label: __VLS_ctx.t('common.loginTime'),
    minWidth: '170',
  },
  ...__VLS_functionalComponentArgsRest(__VLS_86),
)
__VLS_88.slots.default
{
  const { default: __VLS_thisSlot } = __VLS_88.slots
  const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot)
  __VLS_asFunctionalElement(
    __VLS_intrinsicElements.span,
    __VLS_intrinsicElements.span,
  )({
    ...{ style: {} },
  })
  row.loginAt ? new Date(row.loginAt * 1000).toLocaleString() : '-'
}
var __VLS_88
var __VLS_56
__VLS_asFunctionalElement(
  __VLS_intrinsicElements.div,
  __VLS_intrinsicElements.div,
)({
  ...{ style: {} },
})
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({})
__VLS_ctx.t('common.totalItems', { count: __VLS_ctx.pagination.total })
const __VLS_89 = {}.ElButton
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ // @ts-ignore
const __VLS_90 = __VLS_asFunctionalComponent(
  __VLS_89,
  new __VLS_89({
    ...{ onClick: {} },
    disabled: !__VLS_ctx.pagination.hasPrev,
  }),
)
const __VLS_91 = __VLS_90(
  {
    ...{ onClick: {} },
    disabled: !__VLS_ctx.pagination.hasPrev,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_90),
)
let __VLS_93
let __VLS_94
let __VLS_95
const __VLS_96 = {
  onClick: __VLS_ctx.prevPage,
}
__VLS_92.slots.default
__VLS_ctx.t('common.prevPage')
var __VLS_92
const __VLS_97 = {}.ElButton
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ // @ts-ignore
const __VLS_98 = __VLS_asFunctionalComponent(
  __VLS_97,
  new __VLS_97({
    ...{ onClick: {} },
    disabled: !__VLS_ctx.pagination.hasNext,
  }),
)
const __VLS_99 = __VLS_98(
  {
    ...{ onClick: {} },
    disabled: !__VLS_ctx.pagination.hasNext,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_98),
)
let __VLS_101
let __VLS_102
let __VLS_103
const __VLS_104 = {
  onClick: __VLS_ctx.nextPage,
}
__VLS_100.slots.default
__VLS_ctx.t('common.nextPage')
var __VLS_100
const __VLS_105 = {}.ElSelect
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ // @ts-ignore
const __VLS_106 = __VLS_asFunctionalComponent(
  __VLS_105,
  new __VLS_105({
    ...{ onChange: {} },
    modelValue: __VLS_ctx.pagination.limit,
    ...{ style: {} },
  }),
)
const __VLS_107 = __VLS_106(
  {
    ...{ onChange: {} },
    modelValue: __VLS_ctx.pagination.limit,
    ...{ style: {} },
  },
  ...__VLS_functionalComponentArgsRest(__VLS_106),
)
let __VLS_109
let __VLS_110
let __VLS_111
const __VLS_112 = {
  onChange: () => {
    __VLS_ctx.pagination.cursor = null
    __VLS_ctx.pagination.hasPrev = false
    __VLS_ctx.pagination.nextCursor = null
    __VLS_ctx.pagination.prevCursor = null
    __VLS_ctx.fetchList()
  },
}
__VLS_108.slots.default
const __VLS_113 = {}.ElOption
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ // @ts-ignore
const __VLS_114 = __VLS_asFunctionalComponent(
  __VLS_113,
  new __VLS_113({
    label: '10',
    value: 10,
  }),
)
const __VLS_115 = __VLS_114(
  {
    label: '10',
    value: 10,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_114),
)
const __VLS_117 = {}.ElOption
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ // @ts-ignore
const __VLS_118 = __VLS_asFunctionalComponent(
  __VLS_117,
  new __VLS_117({
    label: '20',
    value: 20,
  }),
)
const __VLS_119 = __VLS_118(
  {
    label: '20',
    value: 20,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_118),
)
const __VLS_121 = {}.ElOption
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ // @ts-ignore
const __VLS_122 = __VLS_asFunctionalComponent(
  __VLS_121,
  new __VLS_121({
    label: '50',
    value: 50,
  }),
)
const __VLS_123 = __VLS_122(
  {
    label: '50',
    value: 50,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_122),
)
var __VLS_108
var __VLS_3
var __VLS_dollars
const __VLS_self = (await import('vue')).defineComponent({
  setup() {
    return {
      t: t,
      pagination: pagination,
      loading: loading,
      queryForm: queryForm,
      list_ref: list_ref,
      fetchList: fetchList,
      onSearch: onSearch,
      onReset: onReset,
      nextPage: nextPage,
      prevPage: prevPage,
    }
  },
})
export default (await import('vue')).defineComponent({
  setup() {
    return {}
  },
}) /* PartiallyEnd: #4569/main.vue */
