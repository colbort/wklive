import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { ElMessage } from 'element-plus';
import { cronJobService } from '@/services';
import { usePagination } from '@/composables/usePagination';
import { useLoading } from '@/composables/useLoading';
import { useForm } from '@/composables/useForm';
import { formatDate } from '@/utils';
const { t } = useI18n();
// Pagination and list
const { pagination, updatePagination, nextPage: paginationNextPage, prevPage: paginationPrevPage, } = usePagination(10);
const { loading, withLoading } = useLoading();
// Query form
const { form: queryForm } = useForm({
    initialData: {
        jobName: '',
        invokeTarget: '',
        status: undefined,
    },
});
const list_ref = ref([]);
// Detail drawer
const detailDrawerVisible = ref(false);
const detailData = ref(null);
async function fetchList() {
    await withLoading(async () => {
        try {
            const res = await cronJobService.getLogList({
                jobName: queryForm.jobName || undefined,
                invokeTarget: queryForm.invokeTarget || undefined,
                status: queryForm.status,
                cursor: pagination.cursor,
                limit: pagination.limit,
            });
            if (res.code !== 0 && res.code !== 200)
                throw new Error(res.msg);
            list_ref.value = res.data || [];
            updatePagination(res.total || 0, res.hasNext || false, res.hasPrev || false, res.nextCursor || null, res.prevCursor || null);
        }
        catch (e) {
            ElMessage.error(e?.message || t('common.loadFailed'));
        }
    });
}
function onSearch() {
    pagination.cursor = null;
    pagination.hasPrev = false;
    fetchList();
}
function onReset() {
    queryForm.jobName = '';
    queryForm.invokeTarget = '';
    queryForm.status = undefined;
    pagination.cursor = null;
    pagination.hasPrev = false;
    fetchList();
}
function nextPage() {
    paginationNextPage();
    fetchList();
}
function prevPage() {
    paginationPrevPage();
    fetchList();
}
function showDetail(row) {
    detailData.value = row;
    detailDrawerVisible.value = true;
}
// Format timestamp to date string
const formatTimestamp = (timestamp) => {
    if (!timestamp || timestamp === 0)
        return '-';
    return formatDate(new Date(timestamp * 1000).getTime());
};
// Calculate duration
const calculateDuration = (row) => {
    if (!row.endTime || !row.startTime)
        return '-';
    return `${row.endTime - row.startTime}ms`;
};
onMounted(() => {
    fetchList();
});
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
const __VLS_0 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({}));
const __VLS_2 = __VLS_1({}, ...__VLS_functionalComponentArgsRest(__VLS_1));
var __VLS_4 = {};
__VLS_3.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_3.slots;
    (__VLS_ctx.t('system.cronJobLog'));
}
const __VLS_5 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_6 = __VLS_asFunctionalComponent(__VLS_5, new __VLS_5({
    model: (__VLS_ctx.queryForm),
    inline: true,
    ...{ style: {} },
}));
const __VLS_7 = __VLS_6({
    model: (__VLS_ctx.queryForm),
    inline: true,
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_6));
__VLS_8.slots.default;
const __VLS_9 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_10 = __VLS_asFunctionalComponent(__VLS_9, new __VLS_9({
    label: (__VLS_ctx.t('system.jobName')),
}));
const __VLS_11 = __VLS_10({
    label: (__VLS_ctx.t('system.jobName')),
}, ...__VLS_functionalComponentArgsRest(__VLS_10));
__VLS_12.slots.default;
const __VLS_13 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_14 = __VLS_asFunctionalComponent(__VLS_13, new __VLS_13({
    modelValue: (__VLS_ctx.queryForm.jobName),
    placeholder: (__VLS_ctx.t('common.pleaseEnter')),
    clearable: true,
    ...{ style: {} },
}));
const __VLS_15 = __VLS_14({
    modelValue: (__VLS_ctx.queryForm.jobName),
    placeholder: (__VLS_ctx.t('common.pleaseEnter')),
    clearable: true,
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_14));
var __VLS_12;
const __VLS_17 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_18 = __VLS_asFunctionalComponent(__VLS_17, new __VLS_17({
    label: (__VLS_ctx.t('system.invokeTarget')),
}));
const __VLS_19 = __VLS_18({
    label: (__VLS_ctx.t('system.invokeTarget')),
}, ...__VLS_functionalComponentArgsRest(__VLS_18));
__VLS_20.slots.default;
const __VLS_21 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_22 = __VLS_asFunctionalComponent(__VLS_21, new __VLS_21({
    modelValue: (__VLS_ctx.queryForm.invokeTarget),
    placeholder: (__VLS_ctx.t('common.pleaseEnter')),
    clearable: true,
    ...{ style: {} },
}));
const __VLS_23 = __VLS_22({
    modelValue: (__VLS_ctx.queryForm.invokeTarget),
    placeholder: (__VLS_ctx.t('common.pleaseEnter')),
    clearable: true,
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_22));
var __VLS_20;
const __VLS_25 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_26 = __VLS_asFunctionalComponent(__VLS_25, new __VLS_25({
    label: (__VLS_ctx.t('common.status')),
}));
const __VLS_27 = __VLS_26({
    label: (__VLS_ctx.t('common.status')),
}, ...__VLS_functionalComponentArgsRest(__VLS_26));
__VLS_28.slots.default;
const __VLS_29 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_30 = __VLS_asFunctionalComponent(__VLS_29, new __VLS_29({
    modelValue: (__VLS_ctx.queryForm.status),
    placeholder: (__VLS_ctx.t('common.pleaseSelect')),
    clearable: true,
    ...{ style: {} },
}));
const __VLS_31 = __VLS_30({
    modelValue: (__VLS_ctx.queryForm.status),
    placeholder: (__VLS_ctx.t('common.pleaseSelect')),
    clearable: true,
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_30));
__VLS_32.slots.default;
const __VLS_33 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_34 = __VLS_asFunctionalComponent(__VLS_33, new __VLS_33({
    label: (__VLS_ctx.t('common.success')),
    value: (1),
}));
const __VLS_35 = __VLS_34({
    label: (__VLS_ctx.t('common.success')),
    value: (1),
}, ...__VLS_functionalComponentArgsRest(__VLS_34));
const __VLS_37 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_38 = __VLS_asFunctionalComponent(__VLS_37, new __VLS_37({
    label: (__VLS_ctx.t('common.failed')),
    value: (0),
}));
const __VLS_39 = __VLS_38({
    label: (__VLS_ctx.t('common.failed')),
    value: (0),
}, ...__VLS_functionalComponentArgsRest(__VLS_38));
var __VLS_32;
var __VLS_28;
const __VLS_41 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_42 = __VLS_asFunctionalComponent(__VLS_41, new __VLS_41({}));
const __VLS_43 = __VLS_42({}, ...__VLS_functionalComponentArgsRest(__VLS_42));
__VLS_44.slots.default;
const __VLS_45 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_46 = __VLS_asFunctionalComponent(__VLS_45, new __VLS_45({
    ...{ 'onClick': {} },
    type: "primary",
}));
const __VLS_47 = __VLS_46({
    ...{ 'onClick': {} },
    type: "primary",
}, ...__VLS_functionalComponentArgsRest(__VLS_46));
let __VLS_49;
let __VLS_50;
let __VLS_51;
const __VLS_52 = {
    onClick: (__VLS_ctx.onSearch)
};
__VLS_48.slots.default;
(__VLS_ctx.t('common.search'));
var __VLS_48;
const __VLS_53 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_54 = __VLS_asFunctionalComponent(__VLS_53, new __VLS_53({
    ...{ 'onClick': {} },
}));
const __VLS_55 = __VLS_54({
    ...{ 'onClick': {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_54));
let __VLS_57;
let __VLS_58;
let __VLS_59;
const __VLS_60 = {
    onClick: (__VLS_ctx.onReset)
};
__VLS_56.slots.default;
(__VLS_ctx.t('common.reset'));
var __VLS_56;
var __VLS_44;
var __VLS_8;
const __VLS_61 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_62 = __VLS_asFunctionalComponent(__VLS_61, new __VLS_61({
    data: (__VLS_ctx.list_ref),
    rowKey: "id",
    ...{ style: {} },
}));
const __VLS_63 = __VLS_62({
    data: (__VLS_ctx.list_ref),
    rowKey: "id",
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_62));
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.loading) }, null, null);
__VLS_64.slots.default;
const __VLS_65 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_66 = __VLS_asFunctionalComponent(__VLS_65, new __VLS_65({
    prop: "id",
    label: (__VLS_ctx.t('common.id')),
    width: "70",
}));
const __VLS_67 = __VLS_66({
    prop: "id",
    label: (__VLS_ctx.t('common.id')),
    width: "70",
}, ...__VLS_functionalComponentArgsRest(__VLS_66));
const __VLS_69 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_70 = __VLS_asFunctionalComponent(__VLS_69, new __VLS_69({
    prop: "jobId",
    label: (__VLS_ctx.t('system.jobId')),
    width: "80",
}));
const __VLS_71 = __VLS_70({
    prop: "jobId",
    label: (__VLS_ctx.t('system.jobId')),
    width: "80",
}, ...__VLS_functionalComponentArgsRest(__VLS_70));
const __VLS_73 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_74 = __VLS_asFunctionalComponent(__VLS_73, new __VLS_73({
    prop: "jobName",
    label: (__VLS_ctx.t('system.jobName')),
    minWidth: "120",
}));
const __VLS_75 = __VLS_74({
    prop: "jobName",
    label: (__VLS_ctx.t('system.jobName')),
    minWidth: "120",
}, ...__VLS_functionalComponentArgsRest(__VLS_74));
const __VLS_77 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_78 = __VLS_asFunctionalComponent(__VLS_77, new __VLS_77({
    prop: "invokeTarget",
    label: (__VLS_ctx.t('system.invokeTarget')),
    minWidth: "140",
    showOverflowTooltip: true,
}));
const __VLS_79 = __VLS_78({
    prop: "invokeTarget",
    label: (__VLS_ctx.t('system.invokeTarget')),
    minWidth: "140",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_78));
const __VLS_81 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_82 = __VLS_asFunctionalComponent(__VLS_81, new __VLS_81({
    prop: "status",
    label: (__VLS_ctx.t('common.status')),
    width: "100",
}));
const __VLS_83 = __VLS_82({
    prop: "status",
    label: (__VLS_ctx.t('common.status')),
    width: "100",
}, ...__VLS_functionalComponentArgsRest(__VLS_82));
__VLS_84.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_84.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_85 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_86 = __VLS_asFunctionalComponent(__VLS_85, new __VLS_85({
        type: (row.status === 1 ? 'success' : 'danger'),
    }));
    const __VLS_87 = __VLS_86({
        type: (row.status === 1 ? 'success' : 'danger'),
    }, ...__VLS_functionalComponentArgsRest(__VLS_86));
    __VLS_88.slots.default;
    (row.status === 1 ? __VLS_ctx.t('common.success') : __VLS_ctx.t('common.failed'));
    var __VLS_88;
}
var __VLS_84;
const __VLS_89 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_90 = __VLS_asFunctionalComponent(__VLS_89, new __VLS_89({
    prop: "startTime",
    label: (__VLS_ctx.t('system.startTime')),
    width: "170",
}));
const __VLS_91 = __VLS_90({
    prop: "startTime",
    label: (__VLS_ctx.t('system.startTime')),
    width: "170",
}, ...__VLS_functionalComponentArgsRest(__VLS_90));
__VLS_92.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_92.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (__VLS_ctx.formatTimestamp(row.startTime));
}
var __VLS_92;
const __VLS_93 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_94 = __VLS_asFunctionalComponent(__VLS_93, new __VLS_93({
    prop: "endTime",
    label: (__VLS_ctx.t('system.endTime')),
    width: "170",
}));
const __VLS_95 = __VLS_94({
    prop: "endTime",
    label: (__VLS_ctx.t('system.endTime')),
    width: "170",
}, ...__VLS_functionalComponentArgsRest(__VLS_94));
__VLS_96.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_96.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (__VLS_ctx.formatTimestamp(row.endTime));
}
var __VLS_96;
const __VLS_97 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_98 = __VLS_asFunctionalComponent(__VLS_97, new __VLS_97({
    label: (__VLS_ctx.t('system.duration')),
    width: "100",
}));
const __VLS_99 = __VLS_98({
    label: (__VLS_ctx.t('system.duration')),
    width: "100",
}, ...__VLS_functionalComponentArgsRest(__VLS_98));
__VLS_100.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_100.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (__VLS_ctx.calculateDuration(row));
}
var __VLS_100;
const __VLS_101 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_102 = __VLS_asFunctionalComponent(__VLS_101, new __VLS_101({
    prop: "message",
    label: (__VLS_ctx.t('system.message')),
    minWidth: "140",
    showOverflowTooltip: true,
}));
const __VLS_103 = __VLS_102({
    prop: "message",
    label: (__VLS_ctx.t('system.message')),
    minWidth: "140",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_102));
const __VLS_105 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_106 = __VLS_asFunctionalComponent(__VLS_105, new __VLS_105({
    label: (__VLS_ctx.t('common.actions')),
    width: "120",
    fixed: "right",
}));
const __VLS_107 = __VLS_106({
    label: (__VLS_ctx.t('common.actions')),
    width: "120",
    fixed: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_106));
__VLS_108.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_108.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_109 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_110 = __VLS_asFunctionalComponent(__VLS_109, new __VLS_109({
        ...{ 'onClick': {} },
        type: "primary",
        size: "small",
    }));
    const __VLS_111 = __VLS_110({
        ...{ 'onClick': {} },
        type: "primary",
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_110));
    let __VLS_113;
    let __VLS_114;
    let __VLS_115;
    const __VLS_116 = {
        onClick: (...[$event]) => {
            __VLS_ctx.showDetail(row);
        }
    };
    __VLS_112.slots.default;
    (__VLS_ctx.t('system.viewDetail'));
    var __VLS_112;
}
var __VLS_108;
var __VLS_64;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ style: {} },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
(__VLS_ctx.t('common.totalItems', { count: __VLS_ctx.pagination.total }));
const __VLS_117 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_118 = __VLS_asFunctionalComponent(__VLS_117, new __VLS_117({
    ...{ 'onClick': {} },
    disabled: (!__VLS_ctx.pagination.hasPrev),
}));
const __VLS_119 = __VLS_118({
    ...{ 'onClick': {} },
    disabled: (!__VLS_ctx.pagination.hasPrev),
}, ...__VLS_functionalComponentArgsRest(__VLS_118));
let __VLS_121;
let __VLS_122;
let __VLS_123;
const __VLS_124 = {
    onClick: (__VLS_ctx.prevPage)
};
__VLS_120.slots.default;
(__VLS_ctx.t('common.prevPage'));
var __VLS_120;
const __VLS_125 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_126 = __VLS_asFunctionalComponent(__VLS_125, new __VLS_125({
    ...{ 'onClick': {} },
    disabled: (!__VLS_ctx.pagination.hasNext),
}));
const __VLS_127 = __VLS_126({
    ...{ 'onClick': {} },
    disabled: (!__VLS_ctx.pagination.hasNext),
}, ...__VLS_functionalComponentArgsRest(__VLS_126));
let __VLS_129;
let __VLS_130;
let __VLS_131;
const __VLS_132 = {
    onClick: (__VLS_ctx.nextPage)
};
__VLS_128.slots.default;
(__VLS_ctx.t('common.nextPage'));
var __VLS_128;
const __VLS_133 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_134 = __VLS_asFunctionalComponent(__VLS_133, new __VLS_133({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.pagination.limit),
    ...{ style: {} },
}));
const __VLS_135 = __VLS_134({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.pagination.limit),
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_134));
let __VLS_137;
let __VLS_138;
let __VLS_139;
const __VLS_140 = {
    onChange: (() => {
        __VLS_ctx.pagination.cursor = null;
        __VLS_ctx.pagination.hasPrev = false;
        __VLS_ctx.fetchList();
    })
};
__VLS_136.slots.default;
const __VLS_141 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_142 = __VLS_asFunctionalComponent(__VLS_141, new __VLS_141({
    label: "10",
    value: (10),
}));
const __VLS_143 = __VLS_142({
    label: "10",
    value: (10),
}, ...__VLS_functionalComponentArgsRest(__VLS_142));
const __VLS_145 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_146 = __VLS_asFunctionalComponent(__VLS_145, new __VLS_145({
    label: "20",
    value: (20),
}));
const __VLS_147 = __VLS_146({
    label: "20",
    value: (20),
}, ...__VLS_functionalComponentArgsRest(__VLS_146));
const __VLS_149 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_150 = __VLS_asFunctionalComponent(__VLS_149, new __VLS_149({
    label: "50",
    value: (50),
}));
const __VLS_151 = __VLS_150({
    label: "50",
    value: (50),
}, ...__VLS_functionalComponentArgsRest(__VLS_150));
var __VLS_136;
const __VLS_153 = {}.ElDrawer;
/** @type {[typeof __VLS_components.ElDrawer, typeof __VLS_components.elDrawer, typeof __VLS_components.ElDrawer, typeof __VLS_components.elDrawer, ]} */ ;
// @ts-ignore
const __VLS_154 = __VLS_asFunctionalComponent(__VLS_153, new __VLS_153({
    modelValue: (__VLS_ctx.detailDrawerVisible),
    title: (__VLS_ctx.t('system.logDetail')),
    size: "50%",
}));
const __VLS_155 = __VLS_154({
    modelValue: (__VLS_ctx.detailDrawerVisible),
    title: (__VLS_ctx.t('system.logDetail')),
    size: "50%",
}, ...__VLS_functionalComponentArgsRest(__VLS_154));
__VLS_156.slots.default;
if (__VLS_ctx.detailData) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    const __VLS_157 = {}.ElDescriptions;
    /** @type {[typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, ]} */ ;
    // @ts-ignore
    const __VLS_158 = __VLS_asFunctionalComponent(__VLS_157, new __VLS_157({
        column: (1),
        border: true,
    }));
    const __VLS_159 = __VLS_158({
        column: (1),
        border: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_158));
    __VLS_160.slots.default;
    const __VLS_161 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_162 = __VLS_asFunctionalComponent(__VLS_161, new __VLS_161({
        label: (__VLS_ctx.t('common.id')),
    }));
    const __VLS_163 = __VLS_162({
        label: (__VLS_ctx.t('common.id')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_162));
    __VLS_164.slots.default;
    (__VLS_ctx.detailData.id);
    var __VLS_164;
    const __VLS_165 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_166 = __VLS_asFunctionalComponent(__VLS_165, new __VLS_165({
        label: (__VLS_ctx.t('system.jobId')),
    }));
    const __VLS_167 = __VLS_166({
        label: (__VLS_ctx.t('system.jobId')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_166));
    __VLS_168.slots.default;
    (__VLS_ctx.detailData.jobId);
    var __VLS_168;
    const __VLS_169 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_170 = __VLS_asFunctionalComponent(__VLS_169, new __VLS_169({
        label: (__VLS_ctx.t('system.jobName')),
    }));
    const __VLS_171 = __VLS_170({
        label: (__VLS_ctx.t('system.jobName')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_170));
    __VLS_172.slots.default;
    (__VLS_ctx.detailData.jobName);
    var __VLS_172;
    const __VLS_173 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_174 = __VLS_asFunctionalComponent(__VLS_173, new __VLS_173({
        label: (__VLS_ctx.t('system.invokeTarget')),
    }));
    const __VLS_175 = __VLS_174({
        label: (__VLS_ctx.t('system.invokeTarget')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_174));
    __VLS_176.slots.default;
    (__VLS_ctx.detailData.invokeTarget);
    var __VLS_176;
    const __VLS_177 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_178 = __VLS_asFunctionalComponent(__VLS_177, new __VLS_177({
        label: (__VLS_ctx.t('system.cronExpression')),
    }));
    const __VLS_179 = __VLS_178({
        label: (__VLS_ctx.t('system.cronExpression')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_178));
    __VLS_180.slots.default;
    (__VLS_ctx.detailData.cronExpression);
    var __VLS_180;
    const __VLS_181 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_182 = __VLS_asFunctionalComponent(__VLS_181, new __VLS_181({
        label: (__VLS_ctx.t('common.status')),
    }));
    const __VLS_183 = __VLS_182({
        label: (__VLS_ctx.t('common.status')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_182));
    __VLS_184.slots.default;
    const __VLS_185 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_186 = __VLS_asFunctionalComponent(__VLS_185, new __VLS_185({
        type: (__VLS_ctx.detailData.status === 1 ? 'success' : 'danger'),
    }));
    const __VLS_187 = __VLS_186({
        type: (__VLS_ctx.detailData.status === 1 ? 'success' : 'danger'),
    }, ...__VLS_functionalComponentArgsRest(__VLS_186));
    __VLS_188.slots.default;
    (__VLS_ctx.detailData.status === 1 ? __VLS_ctx.t('common.success') : __VLS_ctx.t('common.failed'));
    var __VLS_188;
    var __VLS_184;
    const __VLS_189 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_190 = __VLS_asFunctionalComponent(__VLS_189, new __VLS_189({
        label: (__VLS_ctx.t('system.startTime')),
    }));
    const __VLS_191 = __VLS_190({
        label: (__VLS_ctx.t('system.startTime')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_190));
    __VLS_192.slots.default;
    (__VLS_ctx.formatTimestamp(__VLS_ctx.detailData.startTime));
    var __VLS_192;
    const __VLS_193 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_194 = __VLS_asFunctionalComponent(__VLS_193, new __VLS_193({
        label: (__VLS_ctx.t('system.endTime')),
    }));
    const __VLS_195 = __VLS_194({
        label: (__VLS_ctx.t('system.endTime')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_194));
    __VLS_196.slots.default;
    (__VLS_ctx.formatTimestamp(__VLS_ctx.detailData.endTime));
    var __VLS_196;
    const __VLS_197 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_198 = __VLS_asFunctionalComponent(__VLS_197, new __VLS_197({
        label: (__VLS_ctx.t('system.duration')),
    }));
    const __VLS_199 = __VLS_198({
        label: (__VLS_ctx.t('system.duration')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_198));
    __VLS_200.slots.default;
    (__VLS_ctx.calculateDuration(__VLS_ctx.detailData));
    var __VLS_200;
    const __VLS_201 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_202 = __VLS_asFunctionalComponent(__VLS_201, new __VLS_201({
        label: (__VLS_ctx.t('system.message')),
    }));
    const __VLS_203 = __VLS_202({
        label: (__VLS_ctx.t('system.message')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_202));
    __VLS_204.slots.default;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    (__VLS_ctx.detailData.message || '-');
    var __VLS_204;
    const __VLS_205 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_206 = __VLS_asFunctionalComponent(__VLS_205, new __VLS_205({
        label: (__VLS_ctx.t('system.exceptionInfo')),
    }));
    const __VLS_207 = __VLS_206({
        label: (__VLS_ctx.t('system.exceptionInfo')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_206));
    __VLS_208.slots.default;
    if (__VLS_ctx.detailData.exceptionInfo) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ style: {} },
        });
        (__VLS_ctx.detailData.exceptionInfo);
    }
    else {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ style: {} },
        });
    }
    var __VLS_208;
    const __VLS_209 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_210 = __VLS_asFunctionalComponent(__VLS_209, new __VLS_209({
        label: (__VLS_ctx.t('common.createdAt')),
    }));
    const __VLS_211 = __VLS_210({
        label: (__VLS_ctx.t('common.createdAt')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_210));
    __VLS_212.slots.default;
    (__VLS_ctx.formatTimestamp(__VLS_ctx.detailData.createTime));
    var __VLS_212;
    var __VLS_160;
}
var __VLS_156;
var __VLS_3;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            t: t,
            pagination: pagination,
            loading: loading,
            queryForm: queryForm,
            list_ref: list_ref,
            detailDrawerVisible: detailDrawerVisible,
            detailData: detailData,
            fetchList: fetchList,
            onSearch: onSearch,
            onReset: onReset,
            nextPage: nextPage,
            prevPage: prevPage,
            showDetail: showDetail,
            formatTimestamp: formatTimestamp,
            calculateDuration: calculateDuration,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
