import { onMounted, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { useI18n } from 'vue-i18n';
import { Plus, Search, Refresh } from '@element-plus/icons-vue';
import { configService } from '@/services';
import { usePagination } from '@/composables/usePagination';
import { useLoading } from '@/composables/useLoading';
import { useForm } from '@/composables/useForm';
import { formatDate } from '@/utils';
import SystemCoreConfig from './components/SystemCoreConfig.vue';
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
// Config type special forms
const systemCoreForm = ref({
    site_name: '',
    site_logo: '',
});
const activeTab = ref('aliyun');
const objectStorageForm = ref({
    aliyun_oss: {
        endpoint: '',
        access_key_id: '',
        access_key_secret: '',
        bucket_name: '',
        bucket_url: '',
    },
    tencent_cos: {
        region: '',
        secret_id: '',
        secret_key: '',
        bucket_name: '',
        bucket_url: '',
    },
    minio: {
        endpoint: '',
        access_key_id: '',
        access_key_secret: '',
        bucket_name: '',
        bucket_url: '',
    },
    oss_type: 1,
    oss_domain: '',
});
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
function resetTypeForms() {
    systemCoreForm.value = {
        site_name: '',
        site_logo: '',
    };
    objectStorageForm.value = {
        aliyun_oss: {
            endpoint: '',
            access_key_id: '',
            access_key_secret: '',
            bucket_name: '',
            bucket_url: '',
        },
        tencent_cos: {
            region: '',
            secret_id: '',
            secret_key: '',
            bucket_name: '',
            bucket_url: '',
        },
        minio: {
            endpoint: '',
            access_key_id: '',
            access_key_secret: '',
            bucket_name: '',
            bucket_url: '',
        },
        oss_type: 1,
        oss_domain: '',
    };
}
function handleConfigKeyChange(value) {
    if (value === 'SYSTEM_CORE') {
        systemCoreForm.value = {
            site_name: '',
            site_logo: '',
        };
        formData.configValue = '';
    }
    else if (value === 'OBJECT_STORAGE') {
        objectStorageForm.value = {
            aliyun_oss: {
                endpoint: '',
                access_key_id: '',
                access_key_secret: '',
                bucket_name: '',
                bucket_url: '',
            },
            tencent_cos: {
                region: '',
                secret_id: '',
                secret_key: '',
                bucket_name: '',
                bucket_url: '',
            },
            minio: {
                endpoint: '',
                access_key_id: '',
                access_key_secret: '',
                bucket_name: '',
                bucket_url: '',
            },
            oss_type: 1,
            oss_domain: '',
        };
        activeTab.value = 'aliyun';
        formData.configValue = '';
    }
}
function handleTabClick(_tab) {
    // 仅切换视图选项卡，不修改 oss_type
}
function handleOssTypeChange(_value) {
    // 仅修改 oss_type，不切换选项卡
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
    resetTypeForms();
    loadKeys();
    dialogVisible.value = true;
}
// Handle edit
function handleEdit(row) {
    isEdit.value = true;
    resetForm();
    resetTypeForms();
    Object.assign(formData, {
        id: row.id,
        configKey: row.configKey,
        remark: row.remark || ''
    });
    if (row.configKey === 'SYSTEM_CORE') {
        try {
            const parsed = JSON.parse(row.configValue || '{}');
            systemCoreForm.value = {
                site_name: parsed.site_name || '',
                site_logo: parsed.site_logo || '',
            };
        }
        catch {
            systemCoreForm.value = { site_name: '', site_logo: '' };
        }
    }
    else if (row.configKey === 'OBJECT_STORAGE') {
        try {
            const parsed = JSON.parse(row.configValue || '{}');
            objectStorageForm.value = {
                aliyun_oss: parsed.aliyun_oss || objectStorageForm.value.aliyun_oss,
                tencent_cos: parsed.tencent_cos || objectStorageForm.value.tencent_cos,
                minio: parsed.minio || objectStorageForm.value.minio,
                oss_type: parsed.oss_type || 1,
                oss_domain: parsed.oss_domain || '',
            };
            if (parsed.oss_type === 1)
                activeTab.value = 'aliyun';
            else if (parsed.oss_type === 2)
                activeTab.value = 'tencent';
            else
                activeTab.value = 'minio';
        }
        catch {
            resetTypeForms();
        }
    }
    else {
        formData.configValue = row.configValue;
    }
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
        if (formData.configKey === 'SYSTEM_CORE') {
            if (!systemCoreForm.value.site_name) {
                throw new Error(t('validation.required'));
            }
            formData.configValue = JSON.stringify(systemCoreForm.value);
        }
        else if (formData.configKey === 'OBJECT_STORAGE') {
            if (!objectStorageForm.value.oss_domain) {
                throw new Error(t('validation.required'));
            }
            formData.configValue = JSON.stringify(objectStorageForm.value);
        }
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
/** @type {__VLS_StyleScopedClasses['config-tabs']} */ ;
/** @type {__VLS_StyleScopedClasses['config-tabs']} */ ;
/** @type {__VLS_StyleScopedClasses['config-tabs']} */ ;
/** @type {__VLS_StyleScopedClasses['config-tabs']} */ ;
/** @type {__VLS_StyleScopedClasses['el-tabs__item']} */ ;
/** @type {__VLS_StyleScopedClasses['config-tabs']} */ ;
/** @type {__VLS_StyleScopedClasses['el-tabs__item']} */ ;
/** @type {__VLS_StyleScopedClasses['config-tabs']} */ ;
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
__VLS_asFunctionalDirective(__VLS_directives.vPerm)(null, { ...__VLS_directiveBindingRestFields, value: ('sys:config:add') }, null, null);
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
const __VLS_28 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.queryForm.keyword),
    placeholder: (__VLS_ctx.t('system.pleaseSelect')),
    filterable: true,
    clearable: true,
}));
const __VLS_30 = __VLS_29({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.queryForm.keyword),
    placeholder: (__VLS_ctx.t('system.pleaseSelect')),
    filterable: true,
    clearable: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_29));
let __VLS_32;
let __VLS_33;
let __VLS_34;
const __VLS_35 = {
    onChange: (__VLS_ctx.fetchList)
};
__VLS_31.slots.default;
for (const [key] of __VLS_getVForSourceType((__VLS_ctx.keys))) {
    const __VLS_36 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({
        key: (key),
        label: (__VLS_ctx.t('system.' + key) || key),
        value: (key),
    }));
    const __VLS_38 = __VLS_37({
        key: (key),
        label: (__VLS_ctx.t('system.' + key) || key),
        value: (key),
    }, ...__VLS_functionalComponentArgsRest(__VLS_37));
}
var __VLS_31;
var __VLS_27;
const __VLS_40 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_41 = __VLS_asFunctionalComponent(__VLS_40, new __VLS_40({}));
const __VLS_42 = __VLS_41({}, ...__VLS_functionalComponentArgsRest(__VLS_41));
__VLS_43.slots.default;
const __VLS_44 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_45 = __VLS_asFunctionalComponent(__VLS_44, new __VLS_44({
    ...{ 'onClick': {} },
    type: "primary",
}));
const __VLS_46 = __VLS_45({
    ...{ 'onClick': {} },
    type: "primary",
}, ...__VLS_functionalComponentArgsRest(__VLS_45));
let __VLS_48;
let __VLS_49;
let __VLS_50;
const __VLS_51 = {
    onClick: (__VLS_ctx.fetchList)
};
__VLS_47.slots.default;
const __VLS_52 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_53 = __VLS_asFunctionalComponent(__VLS_52, new __VLS_52({}));
const __VLS_54 = __VLS_53({}, ...__VLS_functionalComponentArgsRest(__VLS_53));
__VLS_55.slots.default;
const __VLS_56 = {}.Search;
/** @type {[typeof __VLS_components.Search, ]} */ ;
// @ts-ignore
const __VLS_57 = __VLS_asFunctionalComponent(__VLS_56, new __VLS_56({}));
const __VLS_58 = __VLS_57({}, ...__VLS_functionalComponentArgsRest(__VLS_57));
var __VLS_55;
(__VLS_ctx.t('common.search'));
var __VLS_47;
const __VLS_60 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_61 = __VLS_asFunctionalComponent(__VLS_60, new __VLS_60({
    ...{ 'onClick': {} },
}));
const __VLS_62 = __VLS_61({
    ...{ 'onClick': {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_61));
let __VLS_64;
let __VLS_65;
let __VLS_66;
const __VLS_67 = {
    onClick: (__VLS_ctx.handleReset)
};
__VLS_63.slots.default;
const __VLS_68 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_69 = __VLS_asFunctionalComponent(__VLS_68, new __VLS_68({}));
const __VLS_70 = __VLS_69({}, ...__VLS_functionalComponentArgsRest(__VLS_69));
__VLS_71.slots.default;
const __VLS_72 = {}.Refresh;
/** @type {[typeof __VLS_components.Refresh, ]} */ ;
// @ts-ignore
const __VLS_73 = __VLS_asFunctionalComponent(__VLS_72, new __VLS_72({}));
const __VLS_74 = __VLS_73({}, ...__VLS_functionalComponentArgsRest(__VLS_73));
var __VLS_71;
(__VLS_ctx.t('common.reset'));
var __VLS_63;
var __VLS_43;
var __VLS_23;
var __VLS_19;
const __VLS_76 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_77 = __VLS_asFunctionalComponent(__VLS_76, new __VLS_76({
    ...{ class: "table-card" },
    shadow: "never",
}));
const __VLS_78 = __VLS_77({
    ...{ class: "table-card" },
    shadow: "never",
}, ...__VLS_functionalComponentArgsRest(__VLS_77));
__VLS_79.slots.default;
const __VLS_80 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_81 = __VLS_asFunctionalComponent(__VLS_80, new __VLS_80({
    data: (__VLS_ctx.list),
    emptyText: (__VLS_ctx.t('common.noData')),
    stripe: true,
}));
const __VLS_82 = __VLS_81({
    data: (__VLS_ctx.list),
    emptyText: (__VLS_ctx.t('common.noData')),
    stripe: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_81));
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.loading) }, null, null);
__VLS_83.slots.default;
const __VLS_84 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_85 = __VLS_asFunctionalComponent(__VLS_84, new __VLS_84({
    prop: "id",
    label: (__VLS_ctx.t('common.id')),
    width: "80",
    align: "center",
}));
const __VLS_86 = __VLS_85({
    prop: "id",
    label: (__VLS_ctx.t('common.id')),
    width: "80",
    align: "center",
}, ...__VLS_functionalComponentArgsRest(__VLS_85));
const __VLS_88 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_89 = __VLS_asFunctionalComponent(__VLS_88, new __VLS_88({
    prop: "configKey",
    label: (__VLS_ctx.t('system.configKey')),
    minWidth: "150",
}));
const __VLS_90 = __VLS_89({
    prop: "configKey",
    label: (__VLS_ctx.t('system.configKey')),
    minWidth: "150",
}, ...__VLS_functionalComponentArgsRest(__VLS_89));
const __VLS_92 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_93 = __VLS_asFunctionalComponent(__VLS_92, new __VLS_92({
    prop: "configValue",
    label: (__VLS_ctx.t('system.configValue')),
    minWidth: "200",
}));
const __VLS_94 = __VLS_93({
    prop: "configValue",
    label: (__VLS_ctx.t('system.configValue')),
    minWidth: "200",
}, ...__VLS_functionalComponentArgsRest(__VLS_93));
__VLS_95.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_95.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_96 = {}.ElTooltip;
    /** @type {[typeof __VLS_components.ElTooltip, typeof __VLS_components.elTooltip, typeof __VLS_components.ElTooltip, typeof __VLS_components.elTooltip, ]} */ ;
    // @ts-ignore
    const __VLS_97 = __VLS_asFunctionalComponent(__VLS_96, new __VLS_96({
        content: (row.configValue),
        placement: "top",
    }));
    const __VLS_98 = __VLS_97({
        content: (row.configValue),
        placement: "top",
    }, ...__VLS_functionalComponentArgsRest(__VLS_97));
    __VLS_99.slots.default;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "config-value" },
    });
    (row.configValue);
    var __VLS_99;
}
var __VLS_95;
const __VLS_100 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_101 = __VLS_asFunctionalComponent(__VLS_100, new __VLS_100({
    prop: "remark",
    label: (__VLS_ctx.t('common.remark')),
    minWidth: "150",
}));
const __VLS_102 = __VLS_101({
    prop: "remark",
    label: (__VLS_ctx.t('common.remark')),
    minWidth: "150",
}, ...__VLS_functionalComponentArgsRest(__VLS_101));
const __VLS_104 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_105 = __VLS_asFunctionalComponent(__VLS_104, new __VLS_104({
    prop: "createdAt",
    label: (__VLS_ctx.t('common.createdAt')),
    width: "160",
    align: "center",
}));
const __VLS_106 = __VLS_105({
    prop: "createdAt",
    label: (__VLS_ctx.t('common.createdAt')),
    width: "160",
    align: "center",
}, ...__VLS_functionalComponentArgsRest(__VLS_105));
__VLS_107.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_107.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (__VLS_ctx.formatDate(row.createdAt));
}
var __VLS_107;
const __VLS_108 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_109 = __VLS_asFunctionalComponent(__VLS_108, new __VLS_108({
    label: (__VLS_ctx.t('common.actions')),
    width: "150",
    align: "center",
    fixed: "right",
}));
const __VLS_110 = __VLS_109({
    label: (__VLS_ctx.t('common.actions')),
    width: "150",
    align: "center",
    fixed: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_109));
__VLS_111.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_111.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_112 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_113 = __VLS_asFunctionalComponent(__VLS_112, new __VLS_112({
        ...{ 'onClick': {} },
        type: "primary",
        size: "small",
    }));
    const __VLS_114 = __VLS_113({
        ...{ 'onClick': {} },
        type: "primary",
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_113));
    let __VLS_116;
    let __VLS_117;
    let __VLS_118;
    const __VLS_119 = {
        onClick: (...[$event]) => {
            __VLS_ctx.handleEdit(row);
        }
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPerm)(null, { ...__VLS_directiveBindingRestFields, value: ('sys:config:update') }, null, null);
    __VLS_115.slots.default;
    (__VLS_ctx.t('common.edit'));
    var __VLS_115;
    const __VLS_120 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_121 = __VLS_asFunctionalComponent(__VLS_120, new __VLS_120({
        ...{ 'onClick': {} },
        type: "danger",
        size: "small",
    }));
    const __VLS_122 = __VLS_121({
        ...{ 'onClick': {} },
        type: "danger",
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_121));
    let __VLS_124;
    let __VLS_125;
    let __VLS_126;
    const __VLS_127 = {
        onClick: (...[$event]) => {
            __VLS_ctx.handleDelete(row);
        }
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPerm)(null, { ...__VLS_directiveBindingRestFields, value: ('sys:config:delete') }, null, null);
    __VLS_123.slots.default;
    (__VLS_ctx.t('common.delete'));
    var __VLS_123;
}
var __VLS_111;
var __VLS_83;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ style: {} },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
(__VLS_ctx.t('common.totalItems', { count: __VLS_ctx.pagination.total }));
const __VLS_128 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_129 = __VLS_asFunctionalComponent(__VLS_128, new __VLS_128({
    ...{ 'onClick': {} },
    disabled: (!__VLS_ctx.pagination.hasPrev),
}));
const __VLS_130 = __VLS_129({
    ...{ 'onClick': {} },
    disabled: (!__VLS_ctx.pagination.hasPrev),
}, ...__VLS_functionalComponentArgsRest(__VLS_129));
let __VLS_132;
let __VLS_133;
let __VLS_134;
const __VLS_135 = {
    onClick: (__VLS_ctx.prevPage)
};
__VLS_131.slots.default;
(__VLS_ctx.t('common.prevPage'));
var __VLS_131;
const __VLS_136 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_137 = __VLS_asFunctionalComponent(__VLS_136, new __VLS_136({
    ...{ 'onClick': {} },
    disabled: (!__VLS_ctx.pagination.hasNext),
}));
const __VLS_138 = __VLS_137({
    ...{ 'onClick': {} },
    disabled: (!__VLS_ctx.pagination.hasNext),
}, ...__VLS_functionalComponentArgsRest(__VLS_137));
let __VLS_140;
let __VLS_141;
let __VLS_142;
const __VLS_143 = {
    onClick: (__VLS_ctx.nextPage)
};
__VLS_139.slots.default;
(__VLS_ctx.t('common.nextPage'));
var __VLS_139;
const __VLS_144 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_145 = __VLS_asFunctionalComponent(__VLS_144, new __VLS_144({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.pagination.limit),
    ...{ style: {} },
}));
const __VLS_146 = __VLS_145({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.pagination.limit),
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_145));
let __VLS_148;
let __VLS_149;
let __VLS_150;
const __VLS_151 = {
    onChange: (() => { __VLS_ctx.pagination.cursor = null; __VLS_ctx.pagination.hasPrev = false; __VLS_ctx.fetchList(); })
};
__VLS_147.slots.default;
const __VLS_152 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_153 = __VLS_asFunctionalComponent(__VLS_152, new __VLS_152({
    label: "10",
    value: (10),
}));
const __VLS_154 = __VLS_153({
    label: "10",
    value: (10),
}, ...__VLS_functionalComponentArgsRest(__VLS_153));
const __VLS_156 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_157 = __VLS_asFunctionalComponent(__VLS_156, new __VLS_156({
    label: "20",
    value: (20),
}));
const __VLS_158 = __VLS_157({
    label: "20",
    value: (20),
}, ...__VLS_functionalComponentArgsRest(__VLS_157));
const __VLS_160 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_161 = __VLS_asFunctionalComponent(__VLS_160, new __VLS_160({
    label: "50",
    value: (50),
}));
const __VLS_162 = __VLS_161({
    label: "50",
    value: (50),
}, ...__VLS_functionalComponentArgsRest(__VLS_161));
var __VLS_147;
var __VLS_79;
const __VLS_164 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_165 = __VLS_asFunctionalComponent(__VLS_164, new __VLS_164({
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.isEdit ? __VLS_ctx.t('common.edit') : __VLS_ctx.t('common.add')),
    width: "600px",
    closeOnClickModal: (false),
}));
const __VLS_166 = __VLS_165({
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.isEdit ? __VLS_ctx.t('common.edit') : __VLS_ctx.t('common.add')),
    width: "600px",
    closeOnClickModal: (false),
}, ...__VLS_functionalComponentArgsRest(__VLS_165));
__VLS_167.slots.default;
const __VLS_168 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_169 = __VLS_asFunctionalComponent(__VLS_168, new __VLS_168({
    ref: "formRef",
    model: (__VLS_ctx.formData),
    rules: (__VLS_ctx.formRules),
    labelWidth: "100px",
}));
const __VLS_170 = __VLS_169({
    ref: "formRef",
    model: (__VLS_ctx.formData),
    rules: (__VLS_ctx.formRules),
    labelWidth: "100px",
}, ...__VLS_functionalComponentArgsRest(__VLS_169));
/** @type {typeof __VLS_ctx.formRef} */ ;
var __VLS_172 = {};
__VLS_171.slots.default;
const __VLS_174 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_175 = __VLS_asFunctionalComponent(__VLS_174, new __VLS_174({
    label: (__VLS_ctx.t('system.configKey')),
    prop: "configKey",
}));
const __VLS_176 = __VLS_175({
    label: (__VLS_ctx.t('system.configKey')),
    prop: "configKey",
}, ...__VLS_functionalComponentArgsRest(__VLS_175));
__VLS_177.slots.default;
if (!__VLS_ctx.isEdit) {
    const __VLS_178 = {}.ElSelect;
    /** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
    // @ts-ignore
    const __VLS_179 = __VLS_asFunctionalComponent(__VLS_178, new __VLS_178({
        ...{ 'onChange': {} },
        modelValue: (__VLS_ctx.formData.configKey),
        placeholder: (__VLS_ctx.t('system.pleaseSelect')),
        filterable: true,
        clearable: true,
    }));
    const __VLS_180 = __VLS_179({
        ...{ 'onChange': {} },
        modelValue: (__VLS_ctx.formData.configKey),
        placeholder: (__VLS_ctx.t('system.pleaseSelect')),
        filterable: true,
        clearable: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_179));
    let __VLS_182;
    let __VLS_183;
    let __VLS_184;
    const __VLS_185 = {
        onChange: (__VLS_ctx.handleConfigKeyChange)
    };
    __VLS_181.slots.default;
    for (const [key] of __VLS_getVForSourceType((__VLS_ctx.keys))) {
        const __VLS_186 = {}.ElOption;
        /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
        // @ts-ignore
        const __VLS_187 = __VLS_asFunctionalComponent(__VLS_186, new __VLS_186({
            key: (key),
            label: (__VLS_ctx.t('system.' + key) || key),
            value: (key),
        }));
        const __VLS_188 = __VLS_187({
            key: (key),
            label: (__VLS_ctx.t('system.' + key) || key),
            value: (key),
        }, ...__VLS_functionalComponentArgsRest(__VLS_187));
    }
    var __VLS_181;
}
else {
    const __VLS_190 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_191 = __VLS_asFunctionalComponent(__VLS_190, new __VLS_190({
        modelValue: (__VLS_ctx.formData.configKey),
        placeholder: (__VLS_ctx.t('common.pleaseEnter')),
        disabled: true,
    }));
    const __VLS_192 = __VLS_191({
        modelValue: (__VLS_ctx.formData.configKey),
        placeholder: (__VLS_ctx.t('common.pleaseEnter')),
        disabled: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_191));
}
var __VLS_177;
if (__VLS_ctx.formData.configKey === 'SYSTEM_CORE') {
    /** @type {[typeof SystemCoreConfig, ]} */ ;
    // @ts-ignore
    const __VLS_194 = __VLS_asFunctionalComponent(SystemCoreConfig, new SystemCoreConfig({
        modelValue: (__VLS_ctx.systemCoreForm),
    }));
    const __VLS_195 = __VLS_194({
        modelValue: (__VLS_ctx.systemCoreForm),
    }, ...__VLS_functionalComponentArgsRest(__VLS_194));
}
else if (__VLS_ctx.formData.configKey === 'OBJECT_STORAGE') {
    /** @type {[typeof ObjectStorageConfig, ]} */ ;
    // @ts-ignore
    const __VLS_197 = __VLS_asFunctionalComponent(ObjectStorageConfig, new ObjectStorageConfig({
        modelValue: (__VLS_ctx.objectStorageForm),
    }));
    const __VLS_198 = __VLS_197({
        modelValue: (__VLS_ctx.objectStorageForm),
    }, ...__VLS_functionalComponentArgsRest(__VLS_197));
}
else {
    const __VLS_200 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_201 = __VLS_asFunctionalComponent(__VLS_200, new __VLS_200({
        label: (__VLS_ctx.t('system.configValue')),
        prop: "configValue",
    }));
    const __VLS_202 = __VLS_201({
        label: (__VLS_ctx.t('system.configValue')),
        prop: "configValue",
    }, ...__VLS_functionalComponentArgsRest(__VLS_201));
    __VLS_203.slots.default;
    const __VLS_204 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_205 = __VLS_asFunctionalComponent(__VLS_204, new __VLS_204({
        modelValue: (__VLS_ctx.formData.configValue),
        type: "textarea",
        rows: (4),
        placeholder: (__VLS_ctx.t('common.pleaseEnter')),
    }));
    const __VLS_206 = __VLS_205({
        modelValue: (__VLS_ctx.formData.configValue),
        type: "textarea",
        rows: (4),
        placeholder: (__VLS_ctx.t('common.pleaseEnter')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_205));
    var __VLS_203;
}
var __VLS_171;
{
    const { footer: __VLS_thisSlot } = __VLS_167.slots;
    const __VLS_208 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_209 = __VLS_asFunctionalComponent(__VLS_208, new __VLS_208({
        ...{ 'onClick': {} },
    }));
    const __VLS_210 = __VLS_209({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_209));
    let __VLS_212;
    let __VLS_213;
    let __VLS_214;
    const __VLS_215 = {
        onClick: (...[$event]) => {
            __VLS_ctx.dialogVisible = false;
        }
    };
    __VLS_211.slots.default;
    (__VLS_ctx.t('common.cancel'));
    var __VLS_211;
    const __VLS_216 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_217 = __VLS_asFunctionalComponent(__VLS_216, new __VLS_216({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.submitLoading),
    }));
    const __VLS_218 = __VLS_217({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.submitLoading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_217));
    let __VLS_220;
    let __VLS_221;
    let __VLS_222;
    const __VLS_223 = {
        onClick: (__VLS_ctx.handleSubmit)
    };
    __VLS_219.slots.default;
    (__VLS_ctx.t('common.confirm'));
    var __VLS_219;
}
var __VLS_167;
/** @type {__VLS_StyleScopedClasses['sys-config']} */ ;
/** @type {__VLS_StyleScopedClasses['page-header']} */ ;
/** @type {__VLS_StyleScopedClasses['query-card']} */ ;
/** @type {__VLS_StyleScopedClasses['table-card']} */ ;
/** @type {__VLS_StyleScopedClasses['config-value']} */ ;
// @ts-ignore
var __VLS_173 = __VLS_172;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            Plus: Plus,
            Search: Search,
            Refresh: Refresh,
            formatDate: formatDate,
            SystemCoreConfig: SystemCoreConfig,
            ObjectStorageConfig: ObjectStorageConfig,
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
            systemCoreForm: systemCoreForm,
            objectStorageForm: objectStorageForm,
            formRules: formRules,
            fetchList: fetchList,
            handleReset: handleReset,
            handleConfigKeyChange: handleConfigKeyChange,
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
