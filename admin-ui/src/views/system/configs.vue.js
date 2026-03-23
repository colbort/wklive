import { onMounted, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { useI18n } from 'vue-i18n';
import { Plus, Search, Refresh } from '@element-plus/icons-vue';
import { configService } from '@/services';
import { usePagination } from '@/composables/usePagination';
import { useLoading } from '@/composables/useLoading';
import { useForm } from '@/composables/useForm';
import { formatDate } from '@/utils';
const { t } = useI18n();
// Pagination and main list
const { pagination, updatePagination, nextPage: paginationNextPage, prevPage: paginationPrevPage } = usePagination(10);
const list = ref([]);
const { loading, withLoading } = useLoading();
// Query form
const { form: queryForm } = useForm({
    initialData: {
        keyword: ''
    }
});
// Dialog and form
const dialogVisible = ref(false);
const isEdit = ref(false);
const submitLoading = ref(false);
const formRef = ref();
const { form: formData, reset: resetForm } = useForm({
    initialData: {
        id: 0,
        configKey: '',
        configValue: '',
        remark: ''
    }
});
// Keys for configKey selection
const keys = ref([]);
// Form validation rules
const formRules = {
    configKey: [
        { required: true, message: t('validation.required'), trigger: 'blur' }
    ],
    configValue: [
        { required: true, message: t('validation.required'), trigger: 'blur' }
    ]
};
// Load available keys
async function loadKeys() {
    try {
        const res = await configService.getKeys();
        if (res.code !== 0 && res.code !== 200)
            throw new Error(res.msg || 'Failed to load keys');
        keys.value = res.data || [];
    }
    catch (e) {
        ElMessage.error(e?.message || 'Failed to load keys');
    }
}
// Fetch list
async function fetchList() {
    await withLoading(async () => {
        try {
            const res = await configService.getList({
                keyword: queryForm.keyword || undefined,
                cursor: pagination.cursor,
                limit: pagination.limit,
            });
            if (res.code !== 0 && res.code !== 200)
                throw new Error(res.msg || 'list failed');
            list.value = res.data || [];
            updatePagination(res.total || 0, res.hasNext || false, res.hasPrev || false, res.nextCursor || null, res.prevCursor || null);
        }
        catch (e) {
            ElMessage.error(e?.message || t('common.loadFailed'));
        }
    });
}
// Handle pagination
function handleSizeChange(size) {
    pagination.limit = size;
    pagination.cursor = null;
    pagination.hasPrev = false;
    fetchList();
}
// Handle reset
function handleReset() {
    queryForm.keyword = '';
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
// Handle create
function handleCreate() {
    isEdit.value = false;
    resetForm();
    loadKeys();
    dialogVisible.value = true;
}
// Handle edit
function handleEdit(row) {
    isEdit.value = true;
    resetForm();
    Object.assign(formData, {
        id: row.id,
        configKey: row.configKey,
        configValue: row.configValue,
        remark: row.remark || ''
    });
    dialogVisible.value = true;
}
// Handle delete
async function handleDelete(row) {
    try {
        await ElMessageBox.confirm(t('common.confirmDelete'), t('common.warning'), {
            confirmButtonText: t('common.confirm'),
            cancelButtonText: t('common.cancel'),
            type: 'warning',
        });
        const res = await configService.delete(row.id);
        if (res.code !== 0 && res.code !== 200)
            throw new Error(res.msg || 'delete failed');
        ElMessage.success(t('common.deleteSuccess'));
        fetchList();
    }
    catch (e) {
        if (e !== 'cancel') {
            ElMessage.error(e?.message || t('common.deleteFailed'));
        }
    }
}
// Handle submit
async function handleSubmit() {
    if (!formRef.value)
        return;
    try {
        await formRef.value.validate();
        submitLoading.value = true;
        if (isEdit.value) {
            const { id, ...updateData } = formData;
            const res = await configService.update(id, updateData);
            if (res.code !== 0 && res.code !== 200)
                throw new Error(res.msg || t('common.updateFailed'));
            ElMessage.success(t('common.updateSuccess'));
        }
        else {
            const data = {
                configKey: formData.configKey,
                configValue: formData.configValue,
                remark: formData.remark || undefined
            };
            const res = await configService.create(data);
            if (res.code !== 0 && res.code !== 200)
                throw new Error(res.msg || t('common.createFailed'));
            ElMessage.success(t('common.createSuccess'));
        }
        dialogVisible.value = false;
        fetchList();
    }
    catch (e) {
        ElMessage.error(e?.message || t('common.operationFailed'));
    }
    finally {
        submitLoading.value = false;
    }
}
// Initialize
onMounted(() => {
    loadKeys();
    fetchList();
});
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['page-header']} */ ;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "sys-config" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-header" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h2, __VLS_intrinsicElements.h2)({});
(__VLS_ctx.t('system.config'));
const __VLS_0 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    ...{ 'onClick': {} },
    type: "primary",
}));
const __VLS_2 = __VLS_1({
    ...{ 'onClick': {} },
    type: "primary",
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
let __VLS_4;
let __VLS_5;
let __VLS_6;
const __VLS_7 = {
    onClick: (__VLS_ctx.handleCreate)
};
__VLS_asFunctionalDirective(__VLS_directives.vPerm)(null, { ...__VLS_directiveBindingRestFields, value: ('system:config:add') }, null, null);
__VLS_3.slots.default;
const __VLS_8 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({}));
const __VLS_10 = __VLS_9({}, ...__VLS_functionalComponentArgsRest(__VLS_9));
__VLS_11.slots.default;
const __VLS_12 = {}.Plus;
/** @type {[typeof __VLS_components.Plus, ]} */ ;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({}));
const __VLS_14 = __VLS_13({}, ...__VLS_functionalComponentArgsRest(__VLS_13));
var __VLS_11;
(__VLS_ctx.t('common.add'));
var __VLS_3;
const __VLS_16 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
    ...{ class: "query-card" },
    shadow: "never",
}));
const __VLS_18 = __VLS_17({
    ...{ class: "query-card" },
    shadow: "never",
}, ...__VLS_functionalComponentArgsRest(__VLS_17));
__VLS_19.slots.default;
const __VLS_20 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
    model: (__VLS_ctx.queryForm),
    inline: true,
}));
const __VLS_22 = __VLS_21({
    model: (__VLS_ctx.queryForm),
    inline: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_21));
__VLS_23.slots.default;
const __VLS_24 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
    label: (__VLS_ctx.t('system.configKey')),
}));
const __VLS_26 = __VLS_25({
    label: (__VLS_ctx.t('system.configKey')),
}, ...__VLS_functionalComponentArgsRest(__VLS_25));
__VLS_27.slots.default;
const __VLS_28 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({
    ...{ 'onKeyup': {} },
    modelValue: (__VLS_ctx.queryForm.keyword),
    placeholder: (__VLS_ctx.t('common.pleaseEnter')),
    clearable: true,
}));
const __VLS_30 = __VLS_29({
    ...{ 'onKeyup': {} },
    modelValue: (__VLS_ctx.queryForm.keyword),
    placeholder: (__VLS_ctx.t('common.pleaseEnter')),
    clearable: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_29));
let __VLS_32;
let __VLS_33;
let __VLS_34;
const __VLS_35 = {
    onKeyup: (__VLS_ctx.fetchList)
};
var __VLS_31;
var __VLS_27;
const __VLS_36 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({}));
const __VLS_38 = __VLS_37({}, ...__VLS_functionalComponentArgsRest(__VLS_37));
__VLS_39.slots.default;
const __VLS_40 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_41 = __VLS_asFunctionalComponent(__VLS_40, new __VLS_40({
    ...{ 'onClick': {} },
    type: "primary",
}));
const __VLS_42 = __VLS_41({
    ...{ 'onClick': {} },
    type: "primary",
}, ...__VLS_functionalComponentArgsRest(__VLS_41));
let __VLS_44;
let __VLS_45;
let __VLS_46;
const __VLS_47 = {
    onClick: (__VLS_ctx.fetchList)
};
__VLS_43.slots.default;
const __VLS_48 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_49 = __VLS_asFunctionalComponent(__VLS_48, new __VLS_48({}));
const __VLS_50 = __VLS_49({}, ...__VLS_functionalComponentArgsRest(__VLS_49));
__VLS_51.slots.default;
const __VLS_52 = {}.Search;
/** @type {[typeof __VLS_components.Search, ]} */ ;
// @ts-ignore
const __VLS_53 = __VLS_asFunctionalComponent(__VLS_52, new __VLS_52({}));
const __VLS_54 = __VLS_53({}, ...__VLS_functionalComponentArgsRest(__VLS_53));
var __VLS_51;
(__VLS_ctx.t('common.search'));
var __VLS_43;
const __VLS_56 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_57 = __VLS_asFunctionalComponent(__VLS_56, new __VLS_56({
    ...{ 'onClick': {} },
}));
const __VLS_58 = __VLS_57({
    ...{ 'onClick': {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_57));
let __VLS_60;
let __VLS_61;
let __VLS_62;
const __VLS_63 = {
    onClick: (__VLS_ctx.handleReset)
};
__VLS_59.slots.default;
const __VLS_64 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_65 = __VLS_asFunctionalComponent(__VLS_64, new __VLS_64({}));
const __VLS_66 = __VLS_65({}, ...__VLS_functionalComponentArgsRest(__VLS_65));
__VLS_67.slots.default;
const __VLS_68 = {}.Refresh;
/** @type {[typeof __VLS_components.Refresh, ]} */ ;
// @ts-ignore
const __VLS_69 = __VLS_asFunctionalComponent(__VLS_68, new __VLS_68({}));
const __VLS_70 = __VLS_69({}, ...__VLS_functionalComponentArgsRest(__VLS_69));
var __VLS_67;
(__VLS_ctx.t('common.reset'));
var __VLS_59;
var __VLS_39;
var __VLS_23;
var __VLS_19;
const __VLS_72 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_73 = __VLS_asFunctionalComponent(__VLS_72, new __VLS_72({
    ...{ class: "table-card" },
    shadow: "never",
}));
const __VLS_74 = __VLS_73({
    ...{ class: "table-card" },
    shadow: "never",
}, ...__VLS_functionalComponentArgsRest(__VLS_73));
__VLS_75.slots.default;
const __VLS_76 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_77 = __VLS_asFunctionalComponent(__VLS_76, new __VLS_76({
    data: (__VLS_ctx.list),
    emptyText: (__VLS_ctx.t('common.noData')),
    stripe: true,
}));
const __VLS_78 = __VLS_77({
    data: (__VLS_ctx.list),
    emptyText: (__VLS_ctx.t('common.noData')),
    stripe: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_77));
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.loading) }, null, null);
__VLS_79.slots.default;
const __VLS_80 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_81 = __VLS_asFunctionalComponent(__VLS_80, new __VLS_80({
    prop: "id",
    label: (__VLS_ctx.t('common.id')),
    width: "80",
    align: "center",
}));
const __VLS_82 = __VLS_81({
    prop: "id",
    label: (__VLS_ctx.t('common.id')),
    width: "80",
    align: "center",
}, ...__VLS_functionalComponentArgsRest(__VLS_81));
const __VLS_84 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_85 = __VLS_asFunctionalComponent(__VLS_84, new __VLS_84({
    prop: "configKey",
    label: (__VLS_ctx.t('system.configKey')),
    minWidth: "150",
}));
const __VLS_86 = __VLS_85({
    prop: "configKey",
    label: (__VLS_ctx.t('system.configKey')),
    minWidth: "150",
}, ...__VLS_functionalComponentArgsRest(__VLS_85));
const __VLS_88 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_89 = __VLS_asFunctionalComponent(__VLS_88, new __VLS_88({
    prop: "configValue",
    label: (__VLS_ctx.t('system.configValue')),
    minWidth: "200",
}));
const __VLS_90 = __VLS_89({
    prop: "configValue",
    label: (__VLS_ctx.t('system.configValue')),
    minWidth: "200",
}, ...__VLS_functionalComponentArgsRest(__VLS_89));
__VLS_91.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_91.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_92 = {}.ElTooltip;
    /** @type {[typeof __VLS_components.ElTooltip, typeof __VLS_components.elTooltip, typeof __VLS_components.ElTooltip, typeof __VLS_components.elTooltip, ]} */ ;
    // @ts-ignore
    const __VLS_93 = __VLS_asFunctionalComponent(__VLS_92, new __VLS_92({
        content: (row.configValue),
        placement: "top",
    }));
    const __VLS_94 = __VLS_93({
        content: (row.configValue),
        placement: "top",
    }, ...__VLS_functionalComponentArgsRest(__VLS_93));
    __VLS_95.slots.default;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "config-value" },
    });
    (row.configValue);
    var __VLS_95;
}
var __VLS_91;
const __VLS_96 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_97 = __VLS_asFunctionalComponent(__VLS_96, new __VLS_96({
    prop: "remark",
    label: (__VLS_ctx.t('common.remark')),
    minWidth: "150",
}));
const __VLS_98 = __VLS_97({
    prop: "remark",
    label: (__VLS_ctx.t('common.remark')),
    minWidth: "150",
}, ...__VLS_functionalComponentArgsRest(__VLS_97));
const __VLS_100 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_101 = __VLS_asFunctionalComponent(__VLS_100, new __VLS_100({
    prop: "createdAt",
    label: (__VLS_ctx.t('common.createdAt')),
    width: "160",
    align: "center",
}));
const __VLS_102 = __VLS_101({
    prop: "createdAt",
    label: (__VLS_ctx.t('common.createdAt')),
    width: "160",
    align: "center",
}, ...__VLS_functionalComponentArgsRest(__VLS_101));
__VLS_103.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_103.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (__VLS_ctx.formatDate(row.createdAt));
}
var __VLS_103;
const __VLS_104 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_105 = __VLS_asFunctionalComponent(__VLS_104, new __VLS_104({
    label: (__VLS_ctx.t('common.actions')),
    width: "150",
    align: "center",
    fixed: "right",
}));
const __VLS_106 = __VLS_105({
    label: (__VLS_ctx.t('common.actions')),
    width: "150",
    align: "center",
    fixed: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_105));
__VLS_107.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_107.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_108 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_109 = __VLS_asFunctionalComponent(__VLS_108, new __VLS_108({
        ...{ 'onClick': {} },
        type: "primary",
        size: "small",
    }));
    const __VLS_110 = __VLS_109({
        ...{ 'onClick': {} },
        type: "primary",
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_109));
    let __VLS_112;
    let __VLS_113;
    let __VLS_114;
    const __VLS_115 = {
        onClick: (...[$event]) => {
            __VLS_ctx.handleEdit(row);
        }
    };
    __VLS_111.slots.default;
    (__VLS_ctx.t('common.edit'));
    var __VLS_111;
    const __VLS_116 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_117 = __VLS_asFunctionalComponent(__VLS_116, new __VLS_116({
        ...{ 'onClick': {} },
        type: "danger",
        size: "small",
    }));
    const __VLS_118 = __VLS_117({
        ...{ 'onClick': {} },
        type: "danger",
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_117));
    let __VLS_120;
    let __VLS_121;
    let __VLS_122;
    const __VLS_123 = {
        onClick: (...[$event]) => {
            __VLS_ctx.handleDelete(row);
        }
    };
    __VLS_119.slots.default;
    (__VLS_ctx.t('common.delete'));
    var __VLS_119;
}
var __VLS_107;
var __VLS_79;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ style: {} },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
(__VLS_ctx.pagination.total);
const __VLS_124 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_125 = __VLS_asFunctionalComponent(__VLS_124, new __VLS_124({
    ...{ 'onClick': {} },
    disabled: (!__VLS_ctx.pagination.hasPrev),
}));
const __VLS_126 = __VLS_125({
    ...{ 'onClick': {} },
    disabled: (!__VLS_ctx.pagination.hasPrev),
}, ...__VLS_functionalComponentArgsRest(__VLS_125));
let __VLS_128;
let __VLS_129;
let __VLS_130;
const __VLS_131 = {
    onClick: (__VLS_ctx.prevPage)
};
__VLS_127.slots.default;
var __VLS_127;
const __VLS_132 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_133 = __VLS_asFunctionalComponent(__VLS_132, new __VLS_132({
    ...{ 'onClick': {} },
    disabled: (!__VLS_ctx.pagination.hasNext),
}));
const __VLS_134 = __VLS_133({
    ...{ 'onClick': {} },
    disabled: (!__VLS_ctx.pagination.hasNext),
}, ...__VLS_functionalComponentArgsRest(__VLS_133));
let __VLS_136;
let __VLS_137;
let __VLS_138;
const __VLS_139 = {
    onClick: (__VLS_ctx.nextPage)
};
__VLS_135.slots.default;
var __VLS_135;
const __VLS_140 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_141 = __VLS_asFunctionalComponent(__VLS_140, new __VLS_140({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.pagination.limit),
    ...{ style: {} },
}));
const __VLS_142 = __VLS_141({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.pagination.limit),
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_141));
let __VLS_144;
let __VLS_145;
let __VLS_146;
const __VLS_147 = {
    onChange: (() => { __VLS_ctx.pagination.cursor = null; __VLS_ctx.pagination.hasPrev = false; __VLS_ctx.fetchList(); })
};
__VLS_143.slots.default;
const __VLS_148 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_149 = __VLS_asFunctionalComponent(__VLS_148, new __VLS_148({
    label: "10",
    value: (10),
}));
const __VLS_150 = __VLS_149({
    label: "10",
    value: (10),
}, ...__VLS_functionalComponentArgsRest(__VLS_149));
const __VLS_152 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_153 = __VLS_asFunctionalComponent(__VLS_152, new __VLS_152({
    label: "20",
    value: (20),
}));
const __VLS_154 = __VLS_153({
    label: "20",
    value: (20),
}, ...__VLS_functionalComponentArgsRest(__VLS_153));
const __VLS_156 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_157 = __VLS_asFunctionalComponent(__VLS_156, new __VLS_156({
    label: "50",
    value: (50),
}));
const __VLS_158 = __VLS_157({
    label: "50",
    value: (50),
}, ...__VLS_functionalComponentArgsRest(__VLS_157));
var __VLS_143;
var __VLS_75;
const __VLS_160 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_161 = __VLS_asFunctionalComponent(__VLS_160, new __VLS_160({
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.isEdit ? __VLS_ctx.t('common.edit') : __VLS_ctx.t('common.add')),
    width: "600px",
    closeOnClickModal: (false),
}));
const __VLS_162 = __VLS_161({
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.isEdit ? __VLS_ctx.t('common.edit') : __VLS_ctx.t('common.add')),
    width: "600px",
    closeOnClickModal: (false),
}, ...__VLS_functionalComponentArgsRest(__VLS_161));
__VLS_163.slots.default;
const __VLS_164 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_165 = __VLS_asFunctionalComponent(__VLS_164, new __VLS_164({
    ref: "formRef",
    model: (__VLS_ctx.formData),
    rules: (__VLS_ctx.formRules),
    labelWidth: "100px",
}));
const __VLS_166 = __VLS_165({
    ref: "formRef",
    model: (__VLS_ctx.formData),
    rules: (__VLS_ctx.formRules),
    labelWidth: "100px",
}, ...__VLS_functionalComponentArgsRest(__VLS_165));
/** @type {typeof __VLS_ctx.formRef} */ ;
var __VLS_168 = {};
__VLS_167.slots.default;
const __VLS_170 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_171 = __VLS_asFunctionalComponent(__VLS_170, new __VLS_170({
    label: (__VLS_ctx.t('system.configKey')),
    prop: "configKey",
}));
const __VLS_172 = __VLS_171({
    label: (__VLS_ctx.t('system.configKey')),
    prop: "configKey",
}, ...__VLS_functionalComponentArgsRest(__VLS_171));
__VLS_173.slots.default;
if (!__VLS_ctx.isEdit) {
    const __VLS_174 = {}.ElSelect;
    /** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
    // @ts-ignore
    const __VLS_175 = __VLS_asFunctionalComponent(__VLS_174, new __VLS_174({
        modelValue: (__VLS_ctx.formData.configKey),
        placeholder: (__VLS_ctx.t('system.pleaseSelect')),
        filterable: true,
        clearable: true,
    }));
    const __VLS_176 = __VLS_175({
        modelValue: (__VLS_ctx.formData.configKey),
        placeholder: (__VLS_ctx.t('system.pleaseSelect')),
        filterable: true,
        clearable: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_175));
    __VLS_177.slots.default;
    for (const [key] of __VLS_getVForSourceType((__VLS_ctx.keys))) {
        const __VLS_178 = {}.ElOption;
        /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
        // @ts-ignore
        const __VLS_179 = __VLS_asFunctionalComponent(__VLS_178, new __VLS_178({
            key: (key),
            label: (__VLS_ctx.t('system.' + key) || key),
            value: (key),
        }));
        const __VLS_180 = __VLS_179({
            key: (key),
            label: (__VLS_ctx.t('system.' + key) || key),
            value: (key),
        }, ...__VLS_functionalComponentArgsRest(__VLS_179));
    }
    var __VLS_177;
}
else {
    const __VLS_182 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_183 = __VLS_asFunctionalComponent(__VLS_182, new __VLS_182({
        modelValue: (__VLS_ctx.formData.configKey),
        placeholder: (__VLS_ctx.t('common.pleaseEnter')),
        disabled: true,
    }));
    const __VLS_184 = __VLS_183({
        modelValue: (__VLS_ctx.formData.configKey),
        placeholder: (__VLS_ctx.t('common.pleaseEnter')),
        disabled: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_183));
}
var __VLS_173;
const __VLS_186 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_187 = __VLS_asFunctionalComponent(__VLS_186, new __VLS_186({
    label: (__VLS_ctx.t('system.configValue')),
    prop: "configValue",
}));
const __VLS_188 = __VLS_187({
    label: (__VLS_ctx.t('system.configValue')),
    prop: "configValue",
}, ...__VLS_functionalComponentArgsRest(__VLS_187));
__VLS_189.slots.default;
const __VLS_190 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_191 = __VLS_asFunctionalComponent(__VLS_190, new __VLS_190({
    modelValue: (__VLS_ctx.formData.configValue),
    type: "textarea",
    rows: (4),
    placeholder: (__VLS_ctx.t('common.pleaseEnter')),
}));
const __VLS_192 = __VLS_191({
    modelValue: (__VLS_ctx.formData.configValue),
    type: "textarea",
    rows: (4),
    placeholder: (__VLS_ctx.t('common.pleaseEnter')),
}, ...__VLS_functionalComponentArgsRest(__VLS_191));
var __VLS_189;
const __VLS_194 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_195 = __VLS_asFunctionalComponent(__VLS_194, new __VLS_194({
    label: (__VLS_ctx.t('common.remark')),
    prop: "remark",
}));
const __VLS_196 = __VLS_195({
    label: (__VLS_ctx.t('common.remark')),
    prop: "remark",
}, ...__VLS_functionalComponentArgsRest(__VLS_195));
__VLS_197.slots.default;
const __VLS_198 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_199 = __VLS_asFunctionalComponent(__VLS_198, new __VLS_198({
    modelValue: (__VLS_ctx.formData.remark),
    placeholder: (__VLS_ctx.t('common.pleaseEnter')),
}));
const __VLS_200 = __VLS_199({
    modelValue: (__VLS_ctx.formData.remark),
    placeholder: (__VLS_ctx.t('common.pleaseEnter')),
}, ...__VLS_functionalComponentArgsRest(__VLS_199));
var __VLS_197;
var __VLS_167;
{
    const { footer: __VLS_thisSlot } = __VLS_163.slots;
    const __VLS_202 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_203 = __VLS_asFunctionalComponent(__VLS_202, new __VLS_202({
        ...{ 'onClick': {} },
    }));
    const __VLS_204 = __VLS_203({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_203));
    let __VLS_206;
    let __VLS_207;
    let __VLS_208;
    const __VLS_209 = {
        onClick: (...[$event]) => {
            __VLS_ctx.dialogVisible = false;
        }
    };
    __VLS_205.slots.default;
    (__VLS_ctx.t('common.cancel'));
    var __VLS_205;
    const __VLS_210 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_211 = __VLS_asFunctionalComponent(__VLS_210, new __VLS_210({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.submitLoading),
    }));
    const __VLS_212 = __VLS_211({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.submitLoading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_211));
    let __VLS_214;
    let __VLS_215;
    let __VLS_216;
    const __VLS_217 = {
        onClick: (__VLS_ctx.handleSubmit)
    };
    __VLS_213.slots.default;
    (__VLS_ctx.t('common.confirm'));
    var __VLS_213;
}
var __VLS_163;
/** @type {__VLS_StyleScopedClasses['sys-config']} */ ;
/** @type {__VLS_StyleScopedClasses['page-header']} */ ;
/** @type {__VLS_StyleScopedClasses['query-card']} */ ;
/** @type {__VLS_StyleScopedClasses['table-card']} */ ;
/** @type {__VLS_StyleScopedClasses['config-value']} */ ;
// @ts-ignore
var __VLS_169 = __VLS_168;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            Plus: Plus,
            Search: Search,
            Refresh: Refresh,
            formatDate: formatDate,
            t: t,
            pagination: pagination,
            list: list,
            loading: loading,
            queryForm: queryForm,
            dialogVisible: dialogVisible,
            isEdit: isEdit,
            submitLoading: submitLoading,
            formRef: formRef,
            formData: formData,
            keys: keys,
            formRules: formRules,
            fetchList: fetchList,
            handleReset: handleReset,
            nextPage: nextPage,
            prevPage: prevPage,
            handleCreate: handleCreate,
            handleEdit: handleEdit,
            handleDelete: handleDelete,
            handleSubmit: handleSubmit,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
