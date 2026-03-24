import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
const props = defineProps()
const emit = defineEmits()
const activeTab = ref('aliyun')
const form = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})
function handleTabClick(_tab) {
  // 仅切换视图选项卡，不修改 oss_type
}
debugger /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {}
let __VLS_components
let __VLS_directives
/** @type {__VLS_StyleScopedClasses['config-tabs']} */ /** @type {__VLS_StyleScopedClasses['config-tabs']} */ /** @type {__VLS_StyleScopedClasses['config-tabs']} */ /** @type {__VLS_StyleScopedClasses['config-tabs']} */ /** @type {__VLS_StyleScopedClasses['el-tabs__item']} */ /** @type {__VLS_StyleScopedClasses['config-tabs']} */ /** @type {__VLS_StyleScopedClasses['el-tabs__item']} */ /** @type {__VLS_StyleScopedClasses['config-tabs']} */ // CSS variable injection
// CSS variable injection end
__VLS_asFunctionalElement(
  __VLS_intrinsicElements.div,
  __VLS_intrinsicElements.div,
)({
  ...{ class: 'object-storage-config' },
})
const __VLS_0 = {}.ElTabs
/** @type {[typeof __VLS_components.ElTabs, typeof __VLS_components.elTabs, typeof __VLS_components.ElTabs, typeof __VLS_components.elTabs, ]} */ // @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(
  __VLS_0,
  new __VLS_0({
    ...{ onTabClick: {} },
    modelValue: __VLS_ctx.activeTab,
    ...{ class: 'config-tabs' },
  }),
)
const __VLS_2 = __VLS_1(
  {
    ...{ onTabClick: {} },
    modelValue: __VLS_ctx.activeTab,
    ...{ class: 'config-tabs' },
  },
  ...__VLS_functionalComponentArgsRest(__VLS_1),
)
let __VLS_4
let __VLS_5
let __VLS_6
const __VLS_7 = {
  onTabClick: __VLS_ctx.handleTabClick,
}
__VLS_3.slots.default
const __VLS_8 = {}.ElTabPane
/** @type {[typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, ]} */ // @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(
  __VLS_8,
  new __VLS_8({
    label: 'Aliyun OSS',
    name: 'aliyun',
  }),
)
const __VLS_10 = __VLS_9(
  {
    label: 'Aliyun OSS',
    name: 'aliyun',
  },
  ...__VLS_functionalComponentArgsRest(__VLS_9),
)
__VLS_11.slots.default
const __VLS_12 = {}.ElCard
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ // @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(
  __VLS_12,
  new __VLS_12({
    shadow: 'never',
    ...{ class: 'config-card' },
  }),
)
const __VLS_14 = __VLS_13(
  {
    shadow: 'never',
    ...{ class: 'config-card' },
  },
  ...__VLS_functionalComponentArgsRest(__VLS_13),
)
__VLS_15.slots.default
const __VLS_16 = {}.ElRow
/** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ // @ts-ignore
const __VLS_17 = __VLS_asFunctionalComponent(
  __VLS_16,
  new __VLS_16({
    gutter: 20,
  }),
)
const __VLS_18 = __VLS_17(
  {
    gutter: 20,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_17),
)
__VLS_19.slots.default
const __VLS_20 = {}.ElCol
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ // @ts-ignore
const __VLS_21 = __VLS_asFunctionalComponent(
  __VLS_20,
  new __VLS_20({
    span: 12,
  }),
)
const __VLS_22 = __VLS_21(
  {
    span: 12,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_21),
)
__VLS_23.slots.default
const __VLS_24 = {}.ElFormItem
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ // @ts-ignore
const __VLS_25 = __VLS_asFunctionalComponent(
  __VLS_24,
  new __VLS_24({
    label: __VLS_ctx.t('system.endpoint'),
  }),
)
const __VLS_26 = __VLS_25(
  {
    label: __VLS_ctx.t('system.endpoint'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_25),
)
__VLS_27.slots.default
const __VLS_28 = {}.ElInput
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ // @ts-ignore
const __VLS_29 = __VLS_asFunctionalComponent(
  __VLS_28,
  new __VLS_28({
    modelValue: __VLS_ctx.form.aliyun_oss.endpoint,
    placeholder: __VLS_ctx.t('system.endpointPlaceholder'),
  }),
)
const __VLS_30 = __VLS_29(
  {
    modelValue: __VLS_ctx.form.aliyun_oss.endpoint,
    placeholder: __VLS_ctx.t('system.endpointPlaceholder'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_29),
)
var __VLS_27
var __VLS_23
const __VLS_32 = {}.ElCol
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ // @ts-ignore
const __VLS_33 = __VLS_asFunctionalComponent(
  __VLS_32,
  new __VLS_32({
    span: 12,
  }),
)
const __VLS_34 = __VLS_33(
  {
    span: 12,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_33),
)
__VLS_35.slots.default
const __VLS_36 = {}.ElFormItem
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ // @ts-ignore
const __VLS_37 = __VLS_asFunctionalComponent(
  __VLS_36,
  new __VLS_36({
    label: __VLS_ctx.t('system.accessKeyId'),
  }),
)
const __VLS_38 = __VLS_37(
  {
    label: __VLS_ctx.t('system.accessKeyId'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_37),
)
__VLS_39.slots.default
const __VLS_40 = {}.ElInput
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ // @ts-ignore
const __VLS_41 = __VLS_asFunctionalComponent(
  __VLS_40,
  new __VLS_40({
    modelValue: __VLS_ctx.form.aliyun_oss.access_key_id,
    placeholder: __VLS_ctx.t('system.accessKeyIdPlaceholder'),
  }),
)
const __VLS_42 = __VLS_41(
  {
    modelValue: __VLS_ctx.form.aliyun_oss.access_key_id,
    placeholder: __VLS_ctx.t('system.accessKeyIdPlaceholder'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_41),
)
var __VLS_39
var __VLS_35
var __VLS_19
const __VLS_44 = {}.ElRow
/** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ // @ts-ignore
const __VLS_45 = __VLS_asFunctionalComponent(
  __VLS_44,
  new __VLS_44({
    gutter: 20,
  }),
)
const __VLS_46 = __VLS_45(
  {
    gutter: 20,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_45),
)
__VLS_47.slots.default
const __VLS_48 = {}.ElCol
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ // @ts-ignore
const __VLS_49 = __VLS_asFunctionalComponent(
  __VLS_48,
  new __VLS_48({
    span: 12,
  }),
)
const __VLS_50 = __VLS_49(
  {
    span: 12,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_49),
)
__VLS_51.slots.default
const __VLS_52 = {}.ElFormItem
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ // @ts-ignore
const __VLS_53 = __VLS_asFunctionalComponent(
  __VLS_52,
  new __VLS_52({
    label: __VLS_ctx.t('system.accessKeySecret'),
  }),
)
const __VLS_54 = __VLS_53(
  {
    label: __VLS_ctx.t('system.accessKeySecret'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_53),
)
__VLS_55.slots.default
const __VLS_56 = {}.ElInput
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ // @ts-ignore
const __VLS_57 = __VLS_asFunctionalComponent(
  __VLS_56,
  new __VLS_56({
    modelValue: __VLS_ctx.form.aliyun_oss.access_key_secret,
    placeholder: __VLS_ctx.t('system.accessKeySecretPlaceholder'),
  }),
)
const __VLS_58 = __VLS_57(
  {
    modelValue: __VLS_ctx.form.aliyun_oss.access_key_secret,
    placeholder: __VLS_ctx.t('system.accessKeySecretPlaceholder'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_57),
)
var __VLS_55
var __VLS_51
const __VLS_60 = {}.ElCol
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ // @ts-ignore
const __VLS_61 = __VLS_asFunctionalComponent(
  __VLS_60,
  new __VLS_60({
    span: 12,
  }),
)
const __VLS_62 = __VLS_61(
  {
    span: 12,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_61),
)
__VLS_63.slots.default
const __VLS_64 = {}.ElFormItem
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ // @ts-ignore
const __VLS_65 = __VLS_asFunctionalComponent(
  __VLS_64,
  new __VLS_64({
    label: __VLS_ctx.t('system.bucketName'),
  }),
)
const __VLS_66 = __VLS_65(
  {
    label: __VLS_ctx.t('system.bucketName'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_65),
)
__VLS_67.slots.default
const __VLS_68 = {}.ElInput
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ // @ts-ignore
const __VLS_69 = __VLS_asFunctionalComponent(
  __VLS_68,
  new __VLS_68({
    modelValue: __VLS_ctx.form.aliyun_oss.bucket_name,
    placeholder: __VLS_ctx.t('system.bucketNamePlaceholder'),
  }),
)
const __VLS_70 = __VLS_69(
  {
    modelValue: __VLS_ctx.form.aliyun_oss.bucket_name,
    placeholder: __VLS_ctx.t('system.bucketNamePlaceholder'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_69),
)
var __VLS_67
var __VLS_63
var __VLS_47
const __VLS_72 = {}.ElRow
/** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ // @ts-ignore
const __VLS_73 = __VLS_asFunctionalComponent(
  __VLS_72,
  new __VLS_72({
    gutter: 20,
  }),
)
const __VLS_74 = __VLS_73(
  {
    gutter: 20,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_73),
)
__VLS_75.slots.default
const __VLS_76 = {}.ElCol
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ // @ts-ignore
const __VLS_77 = __VLS_asFunctionalComponent(
  __VLS_76,
  new __VLS_76({
    span: 12,
  }),
)
const __VLS_78 = __VLS_77(
  {
    span: 12,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_77),
)
__VLS_79.slots.default
const __VLS_80 = {}.ElFormItem
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ // @ts-ignore
const __VLS_81 = __VLS_asFunctionalComponent(
  __VLS_80,
  new __VLS_80({
    label: __VLS_ctx.t('system.bucketUrl'),
  }),
)
const __VLS_82 = __VLS_81(
  {
    label: __VLS_ctx.t('system.bucketUrl'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_81),
)
__VLS_83.slots.default
const __VLS_84 = {}.ElInput
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ // @ts-ignore
const __VLS_85 = __VLS_asFunctionalComponent(
  __VLS_84,
  new __VLS_84({
    modelValue: __VLS_ctx.form.aliyun_oss.bucket_url,
    placeholder: __VLS_ctx.t('system.bucketUrlPlaceholder'),
  }),
)
const __VLS_86 = __VLS_85(
  {
    modelValue: __VLS_ctx.form.aliyun_oss.bucket_url,
    placeholder: __VLS_ctx.t('system.bucketUrlPlaceholder'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_85),
)
var __VLS_83
var __VLS_79
var __VLS_75
var __VLS_15
var __VLS_11
const __VLS_88 = {}.ElTabPane
/** @type {[typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, ]} */ // @ts-ignore
const __VLS_89 = __VLS_asFunctionalComponent(
  __VLS_88,
  new __VLS_88({
    label: 'Tencent COS',
    name: 'tencent',
  }),
)
const __VLS_90 = __VLS_89(
  {
    label: 'Tencent COS',
    name: 'tencent',
  },
  ...__VLS_functionalComponentArgsRest(__VLS_89),
)
__VLS_91.slots.default
const __VLS_92 = {}.ElCard
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ // @ts-ignore
const __VLS_93 = __VLS_asFunctionalComponent(
  __VLS_92,
  new __VLS_92({
    shadow: 'never',
    ...{ class: 'config-card' },
  }),
)
const __VLS_94 = __VLS_93(
  {
    shadow: 'never',
    ...{ class: 'config-card' },
  },
  ...__VLS_functionalComponentArgsRest(__VLS_93),
)
__VLS_95.slots.default
const __VLS_96 = {}.ElRow
/** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ // @ts-ignore
const __VLS_97 = __VLS_asFunctionalComponent(
  __VLS_96,
  new __VLS_96({
    gutter: 20,
  }),
)
const __VLS_98 = __VLS_97(
  {
    gutter: 20,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_97),
)
__VLS_99.slots.default
const __VLS_100 = {}.ElCol
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ // @ts-ignore
const __VLS_101 = __VLS_asFunctionalComponent(
  __VLS_100,
  new __VLS_100({
    span: 12,
  }),
)
const __VLS_102 = __VLS_101(
  {
    span: 12,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_101),
)
__VLS_103.slots.default
const __VLS_104 = {}.ElFormItem
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ // @ts-ignore
const __VLS_105 = __VLS_asFunctionalComponent(
  __VLS_104,
  new __VLS_104({
    label: __VLS_ctx.t('system.region'),
  }),
)
const __VLS_106 = __VLS_105(
  {
    label: __VLS_ctx.t('system.region'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_105),
)
__VLS_107.slots.default
const __VLS_108 = {}.ElInput
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ // @ts-ignore
const __VLS_109 = __VLS_asFunctionalComponent(
  __VLS_108,
  new __VLS_108({
    modelValue: __VLS_ctx.form.tencent_cos.region,
    placeholder: __VLS_ctx.t('system.regionPlaceholder'),
  }),
)
const __VLS_110 = __VLS_109(
  {
    modelValue: __VLS_ctx.form.tencent_cos.region,
    placeholder: __VLS_ctx.t('system.regionPlaceholder'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_109),
)
var __VLS_107
var __VLS_103
const __VLS_112 = {}.ElCol
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ // @ts-ignore
const __VLS_113 = __VLS_asFunctionalComponent(
  __VLS_112,
  new __VLS_112({
    span: 12,
  }),
)
const __VLS_114 = __VLS_113(
  {
    span: 12,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_113),
)
__VLS_115.slots.default
const __VLS_116 = {}.ElFormItem
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ // @ts-ignore
const __VLS_117 = __VLS_asFunctionalComponent(
  __VLS_116,
  new __VLS_116({
    label: __VLS_ctx.t('system.secretId'),
  }),
)
const __VLS_118 = __VLS_117(
  {
    label: __VLS_ctx.t('system.secretId'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_117),
)
__VLS_119.slots.default
const __VLS_120 = {}.ElInput
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ // @ts-ignore
const __VLS_121 = __VLS_asFunctionalComponent(
  __VLS_120,
  new __VLS_120({
    modelValue: __VLS_ctx.form.tencent_cos.secret_id,
    placeholder: __VLS_ctx.t('system.secretIdPlaceholder'),
  }),
)
const __VLS_122 = __VLS_121(
  {
    modelValue: __VLS_ctx.form.tencent_cos.secret_id,
    placeholder: __VLS_ctx.t('system.secretIdPlaceholder'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_121),
)
var __VLS_119
var __VLS_115
var __VLS_99
const __VLS_124 = {}.ElRow
/** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ // @ts-ignore
const __VLS_125 = __VLS_asFunctionalComponent(
  __VLS_124,
  new __VLS_124({
    gutter: 20,
  }),
)
const __VLS_126 = __VLS_125(
  {
    gutter: 20,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_125),
)
__VLS_127.slots.default
const __VLS_128 = {}.ElCol
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ // @ts-ignore
const __VLS_129 = __VLS_asFunctionalComponent(
  __VLS_128,
  new __VLS_128({
    span: 12,
  }),
)
const __VLS_130 = __VLS_129(
  {
    span: 12,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_129),
)
__VLS_131.slots.default
const __VLS_132 = {}.ElFormItem
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ // @ts-ignore
const __VLS_133 = __VLS_asFunctionalComponent(
  __VLS_132,
  new __VLS_132({
    label: __VLS_ctx.t('system.secretKey'),
  }),
)
const __VLS_134 = __VLS_133(
  {
    label: __VLS_ctx.t('system.secretKey'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_133),
)
__VLS_135.slots.default
const __VLS_136 = {}.ElInput
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ // @ts-ignore
const __VLS_137 = __VLS_asFunctionalComponent(
  __VLS_136,
  new __VLS_136({
    modelValue: __VLS_ctx.form.tencent_cos.secret_key,
    placeholder: __VLS_ctx.t('system.secretKeyPlaceholder'),
  }),
)
const __VLS_138 = __VLS_137(
  {
    modelValue: __VLS_ctx.form.tencent_cos.secret_key,
    placeholder: __VLS_ctx.t('system.secretKeyPlaceholder'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_137),
)
var __VLS_135
var __VLS_131
const __VLS_140 = {}.ElCol
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ // @ts-ignore
const __VLS_141 = __VLS_asFunctionalComponent(
  __VLS_140,
  new __VLS_140({
    span: 12,
  }),
)
const __VLS_142 = __VLS_141(
  {
    span: 12,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_141),
)
__VLS_143.slots.default
const __VLS_144 = {}.ElFormItem
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ // @ts-ignore
const __VLS_145 = __VLS_asFunctionalComponent(
  __VLS_144,
  new __VLS_144({
    label: __VLS_ctx.t('system.bucketName'),
  }),
)
const __VLS_146 = __VLS_145(
  {
    label: __VLS_ctx.t('system.bucketName'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_145),
)
__VLS_147.slots.default
const __VLS_148 = {}.ElInput
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ // @ts-ignore
const __VLS_149 = __VLS_asFunctionalComponent(
  __VLS_148,
  new __VLS_148({
    modelValue: __VLS_ctx.form.tencent_cos.bucket_name,
    placeholder: __VLS_ctx.t('system.bucketNamePlaceholder'),
  }),
)
const __VLS_150 = __VLS_149(
  {
    modelValue: __VLS_ctx.form.tencent_cos.bucket_name,
    placeholder: __VLS_ctx.t('system.bucketNamePlaceholder'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_149),
)
var __VLS_147
var __VLS_143
var __VLS_127
const __VLS_152 = {}.ElRow
/** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ // @ts-ignore
const __VLS_153 = __VLS_asFunctionalComponent(
  __VLS_152,
  new __VLS_152({
    gutter: 20,
  }),
)
const __VLS_154 = __VLS_153(
  {
    gutter: 20,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_153),
)
__VLS_155.slots.default
const __VLS_156 = {}.ElCol
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ // @ts-ignore
const __VLS_157 = __VLS_asFunctionalComponent(
  __VLS_156,
  new __VLS_156({
    span: 12,
  }),
)
const __VLS_158 = __VLS_157(
  {
    span: 12,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_157),
)
__VLS_159.slots.default
const __VLS_160 = {}.ElFormItem
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ // @ts-ignore
const __VLS_161 = __VLS_asFunctionalComponent(
  __VLS_160,
  new __VLS_160({
    label: __VLS_ctx.t('system.bucketUrl'),
  }),
)
const __VLS_162 = __VLS_161(
  {
    label: __VLS_ctx.t('system.bucketUrl'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_161),
)
__VLS_163.slots.default
const __VLS_164 = {}.ElInput
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ // @ts-ignore
const __VLS_165 = __VLS_asFunctionalComponent(
  __VLS_164,
  new __VLS_164({
    modelValue: __VLS_ctx.form.tencent_cos.bucket_url,
    placeholder: __VLS_ctx.t('system.bucketUrlPlaceholder'),
  }),
)
const __VLS_166 = __VLS_165(
  {
    modelValue: __VLS_ctx.form.tencent_cos.bucket_url,
    placeholder: __VLS_ctx.t('system.bucketUrlPlaceholder'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_165),
)
var __VLS_163
var __VLS_159
var __VLS_155
var __VLS_95
var __VLS_91
const __VLS_168 = {}.ElTabPane
/** @type {[typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, ]} */ // @ts-ignore
const __VLS_169 = __VLS_asFunctionalComponent(
  __VLS_168,
  new __VLS_168({
    label: 'MinIO',
    name: 'minio',
  }),
)
const __VLS_170 = __VLS_169(
  {
    label: 'MinIO',
    name: 'minio',
  },
  ...__VLS_functionalComponentArgsRest(__VLS_169),
)
__VLS_171.slots.default
const __VLS_172 = {}.ElCard
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ // @ts-ignore
const __VLS_173 = __VLS_asFunctionalComponent(
  __VLS_172,
  new __VLS_172({
    shadow: 'never',
    ...{ class: 'config-card' },
  }),
)
const __VLS_174 = __VLS_173(
  {
    shadow: 'never',
    ...{ class: 'config-card' },
  },
  ...__VLS_functionalComponentArgsRest(__VLS_173),
)
__VLS_175.slots.default
const __VLS_176 = {}.ElRow
/** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ // @ts-ignore
const __VLS_177 = __VLS_asFunctionalComponent(
  __VLS_176,
  new __VLS_176({
    gutter: 20,
  }),
)
const __VLS_178 = __VLS_177(
  {
    gutter: 20,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_177),
)
__VLS_179.slots.default
const __VLS_180 = {}.ElCol
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ // @ts-ignore
const __VLS_181 = __VLS_asFunctionalComponent(
  __VLS_180,
  new __VLS_180({
    span: 12,
  }),
)
const __VLS_182 = __VLS_181(
  {
    span: 12,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_181),
)
__VLS_183.slots.default
const __VLS_184 = {}.ElFormItem
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ // @ts-ignore
const __VLS_185 = __VLS_asFunctionalComponent(
  __VLS_184,
  new __VLS_184({
    label: __VLS_ctx.t('system.endpoint'),
  }),
)
const __VLS_186 = __VLS_185(
  {
    label: __VLS_ctx.t('system.endpoint'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_185),
)
__VLS_187.slots.default
const __VLS_188 = {}.ElInput
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ // @ts-ignore
const __VLS_189 = __VLS_asFunctionalComponent(
  __VLS_188,
  new __VLS_188({
    modelValue: __VLS_ctx.form.minio.endpoint,
    placeholder: __VLS_ctx.t('system.endpointPlaceholder'),
  }),
)
const __VLS_190 = __VLS_189(
  {
    modelValue: __VLS_ctx.form.minio.endpoint,
    placeholder: __VLS_ctx.t('system.endpointPlaceholder'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_189),
)
var __VLS_187
var __VLS_183
const __VLS_192 = {}.ElCol
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ // @ts-ignore
const __VLS_193 = __VLS_asFunctionalComponent(
  __VLS_192,
  new __VLS_192({
    span: 12,
  }),
)
const __VLS_194 = __VLS_193(
  {
    span: 12,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_193),
)
__VLS_195.slots.default
const __VLS_196 = {}.ElFormItem
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ // @ts-ignore
const __VLS_197 = __VLS_asFunctionalComponent(
  __VLS_196,
  new __VLS_196({
    label: __VLS_ctx.t('system.accessKeyId'),
  }),
)
const __VLS_198 = __VLS_197(
  {
    label: __VLS_ctx.t('system.accessKeyId'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_197),
)
__VLS_199.slots.default
const __VLS_200 = {}.ElInput
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ // @ts-ignore
const __VLS_201 = __VLS_asFunctionalComponent(
  __VLS_200,
  new __VLS_200({
    modelValue: __VLS_ctx.form.minio.access_key_id,
    placeholder: __VLS_ctx.t('system.accessKeyIdPlaceholder'),
  }),
)
const __VLS_202 = __VLS_201(
  {
    modelValue: __VLS_ctx.form.minio.access_key_id,
    placeholder: __VLS_ctx.t('system.accessKeyIdPlaceholder'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_201),
)
var __VLS_199
var __VLS_195
var __VLS_179
const __VLS_204 = {}.ElRow
/** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ // @ts-ignore
const __VLS_205 = __VLS_asFunctionalComponent(
  __VLS_204,
  new __VLS_204({
    gutter: 20,
  }),
)
const __VLS_206 = __VLS_205(
  {
    gutter: 20,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_205),
)
__VLS_207.slots.default
const __VLS_208 = {}.ElCol
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ // @ts-ignore
const __VLS_209 = __VLS_asFunctionalComponent(
  __VLS_208,
  new __VLS_208({
    span: 12,
  }),
)
const __VLS_210 = __VLS_209(
  {
    span: 12,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_209),
)
__VLS_211.slots.default
const __VLS_212 = {}.ElFormItem
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ // @ts-ignore
const __VLS_213 = __VLS_asFunctionalComponent(
  __VLS_212,
  new __VLS_212({
    label: __VLS_ctx.t('system.accessKeySecret'),
  }),
)
const __VLS_214 = __VLS_213(
  {
    label: __VLS_ctx.t('system.accessKeySecret'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_213),
)
__VLS_215.slots.default
const __VLS_216 = {}.ElInput
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ // @ts-ignore
const __VLS_217 = __VLS_asFunctionalComponent(
  __VLS_216,
  new __VLS_216({
    modelValue: __VLS_ctx.form.minio.access_key_secret,
    placeholder: __VLS_ctx.t('system.accessKeySecretPlaceholder'),
  }),
)
const __VLS_218 = __VLS_217(
  {
    modelValue: __VLS_ctx.form.minio.access_key_secret,
    placeholder: __VLS_ctx.t('system.accessKeySecretPlaceholder'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_217),
)
var __VLS_215
var __VLS_211
const __VLS_220 = {}.ElCol
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ // @ts-ignore
const __VLS_221 = __VLS_asFunctionalComponent(
  __VLS_220,
  new __VLS_220({
    span: 12,
  }),
)
const __VLS_222 = __VLS_221(
  {
    span: 12,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_221),
)
__VLS_223.slots.default
const __VLS_224 = {}.ElFormItem
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ // @ts-ignore
const __VLS_225 = __VLS_asFunctionalComponent(
  __VLS_224,
  new __VLS_224({
    label: __VLS_ctx.t('system.bucketName'),
  }),
)
const __VLS_226 = __VLS_225(
  {
    label: __VLS_ctx.t('system.bucketName'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_225),
)
__VLS_227.slots.default
const __VLS_228 = {}.ElInput
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ // @ts-ignore
const __VLS_229 = __VLS_asFunctionalComponent(
  __VLS_228,
  new __VLS_228({
    modelValue: __VLS_ctx.form.minio.bucket_name,
    placeholder: __VLS_ctx.t('system.bucketNamePlaceholder'),
  }),
)
const __VLS_230 = __VLS_229(
  {
    modelValue: __VLS_ctx.form.minio.bucket_name,
    placeholder: __VLS_ctx.t('system.bucketNamePlaceholder'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_229),
)
var __VLS_227
var __VLS_223
var __VLS_207
const __VLS_232 = {}.ElRow
/** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ // @ts-ignore
const __VLS_233 = __VLS_asFunctionalComponent(
  __VLS_232,
  new __VLS_232({
    gutter: 20,
  }),
)
const __VLS_234 = __VLS_233(
  {
    gutter: 20,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_233),
)
__VLS_235.slots.default
const __VLS_236 = {}.ElCol
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ // @ts-ignore
const __VLS_237 = __VLS_asFunctionalComponent(
  __VLS_236,
  new __VLS_236({
    span: 12,
  }),
)
const __VLS_238 = __VLS_237(
  {
    span: 12,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_237),
)
__VLS_239.slots.default
const __VLS_240 = {}.ElFormItem
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ // @ts-ignore
const __VLS_241 = __VLS_asFunctionalComponent(
  __VLS_240,
  new __VLS_240({
    label: __VLS_ctx.t('system.bucketUrl'),
  }),
)
const __VLS_242 = __VLS_241(
  {
    label: __VLS_ctx.t('system.bucketUrl'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_241),
)
__VLS_243.slots.default
const __VLS_244 = {}.ElInput
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ // @ts-ignore
const __VLS_245 = __VLS_asFunctionalComponent(
  __VLS_244,
  new __VLS_244({
    modelValue: __VLS_ctx.form.minio.bucket_url,
    placeholder: __VLS_ctx.t('system.bucketUrlPlaceholder'),
  }),
)
const __VLS_246 = __VLS_245(
  {
    modelValue: __VLS_ctx.form.minio.bucket_url,
    placeholder: __VLS_ctx.t('system.bucketUrlPlaceholder'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_245),
)
var __VLS_243
var __VLS_239
var __VLS_235
var __VLS_175
var __VLS_171
var __VLS_3
const __VLS_248 = {}.ElFormItem
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ // @ts-ignore
const __VLS_249 = __VLS_asFunctionalComponent(
  __VLS_248,
  new __VLS_248({
    label: __VLS_ctx.t('system.ossType'),
    prop: 'oss_type',
  }),
)
const __VLS_250 = __VLS_249(
  {
    label: __VLS_ctx.t('system.ossType'),
    prop: 'oss_type',
  },
  ...__VLS_functionalComponentArgsRest(__VLS_249),
)
__VLS_251.slots.default
const __VLS_252 = {}.ElSelect
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ // @ts-ignore
const __VLS_253 = __VLS_asFunctionalComponent(
  __VLS_252,
  new __VLS_252({
    modelValue: __VLS_ctx.form.oss_type,
    placeholder: __VLS_ctx.t('system.ossTypePlaceholder'),
    filterable: false,
  }),
)
const __VLS_254 = __VLS_253(
  {
    modelValue: __VLS_ctx.form.oss_type,
    placeholder: __VLS_ctx.t('system.ossTypePlaceholder'),
    filterable: false,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_253),
)
__VLS_255.slots.default
const __VLS_256 = {}.ElOption
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ // @ts-ignore
const __VLS_257 = __VLS_asFunctionalComponent(
  __VLS_256,
  new __VLS_256({
    label: 'Aliyun OSS',
    value: 1,
  }),
)
const __VLS_258 = __VLS_257(
  {
    label: 'Aliyun OSS',
    value: 1,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_257),
)
const __VLS_260 = {}.ElOption
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ // @ts-ignore
const __VLS_261 = __VLS_asFunctionalComponent(
  __VLS_260,
  new __VLS_260({
    label: 'Tencent COS',
    value: 2,
  }),
)
const __VLS_262 = __VLS_261(
  {
    label: 'Tencent COS',
    value: 2,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_261),
)
const __VLS_264 = {}.ElOption
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ // @ts-ignore
const __VLS_265 = __VLS_asFunctionalComponent(
  __VLS_264,
  new __VLS_264({
    label: 'MinIO',
    value: 3,
  }),
)
const __VLS_266 = __VLS_265(
  {
    label: 'MinIO',
    value: 3,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_265),
)
var __VLS_255
var __VLS_251
const __VLS_268 = {}.ElFormItem
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ // @ts-ignore
const __VLS_269 = __VLS_asFunctionalComponent(
  __VLS_268,
  new __VLS_268({
    label: __VLS_ctx.t('system.ossDomain'),
    prop: 'oss_domain',
  }),
)
const __VLS_270 = __VLS_269(
  {
    label: __VLS_ctx.t('system.ossDomain'),
    prop: 'oss_domain',
  },
  ...__VLS_functionalComponentArgsRest(__VLS_269),
)
__VLS_271.slots.default
const __VLS_272 = {}.ElInput
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ // @ts-ignore
const __VLS_273 = __VLS_asFunctionalComponent(
  __VLS_272,
  new __VLS_272({
    modelValue: __VLS_ctx.form.oss_domain,
    placeholder: __VLS_ctx.t('common.pleaseEnter'),
  }),
)
const __VLS_274 = __VLS_273(
  {
    modelValue: __VLS_ctx.form.oss_domain,
    placeholder: __VLS_ctx.t('common.pleaseEnter'),
  },
  ...__VLS_functionalComponentArgsRest(__VLS_273),
)
var __VLS_271
/** @type {__VLS_StyleScopedClasses['object-storage-config']} */ /** @type {__VLS_StyleScopedClasses['config-tabs']} */ /** @type {__VLS_StyleScopedClasses['config-card']} */ /** @type {__VLS_StyleScopedClasses['config-card']} */ /** @type {__VLS_StyleScopedClasses['config-card']} */ var __VLS_dollars
const __VLS_self = (await import('vue')).defineComponent({
  setup() {
    return {
      t: t,
      activeTab: activeTab,
      form: form,
      handleTabClick: handleTabClick,
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
