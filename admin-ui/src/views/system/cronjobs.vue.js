import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { ElMessage } from 'element-plus';
import { Plus, Edit, Delete, VideoPlay, CircleCheck, CircleCloseFilled } from '@element-plus/icons-vue';
import { cronJobService } from '@/services';
import { usePagination } from '@/composables/usePagination';
import { useLoading } from '@/composables/useLoading';
import { useForm } from '@/composables/useForm';
import { useConfirm } from '@/composables/useConfirm';
import { formatDate } from '@/utils';
const { t } = useI18n();
const { confirm } = useConfirm();
// Pagination and main list
const { pagination, updatePagination, nextPage: paginationNextPage, prevPage: paginationPrevPage, } = usePagination(10);
const list = ref([]);
const { loading, withLoading } = useLoading();
// Query form
const { form: queryForm } = useForm({
    initialData: {
        jobName: '',
        jobGroup: '',
        status: undefined,
    },
});
// Dialog and form
const dialogVisible = ref(false);
const isEdit = ref(false);
const submitLoading = ref(false);
const formRef = ref();
const { form: formData, reset: resetForm } = useForm({
    initialData: {
        id: 0,
        jobName: '',
        jobGroup: 'DEFAULT',
        invokeTarget: '',
        cronExpression: '',
        status: 0,
        remark: '',
    },
});
// Handlers for dropdown
const handlers = ref([]);
// Form rules
const formRules = computed(() => ({
    jobName: [
        { required: true, message: t('common.required'), trigger: 'blur' },
        { min: 1, max: 100, message: 'Length should be 1-100', trigger: 'blur' },
    ],
    jobGroup: [{ required: true, message: t('common.required'), trigger: 'blur' }],
    invokeTarget: [{ required: true, message: t('common.required'), trigger: 'blur' }],
    cronExpression: [{ required: true, message: t('common.required'), trigger: 'blur' }],
}));
// Fetch list
async function fetchList() {
    await withLoading(async () => {
        try {
            const res = await cronJobService.getList({
                jobName: queryForm.jobName || undefined,
                jobGroup: queryForm.jobGroup || undefined,
                status: queryForm.status,
                cursor: pagination.cursor,
                limit: pagination.limit,
            });
            if (res.code !== 0 && res.code !== 200)
                throw new Error(res.msg);
            list.value = res.data || [];
            updatePagination(res.total || 0, res.hasNext || false, res.hasPrev || false, res.nextCursor || null, res.prevCursor || null);
        }
        catch (e) {
            ElMessage.error(e?.message || t('common.loadFailed'));
        }
    });
}
// Fetch handlers
async function fetchHandlers() {
    try {
        const res = await cronJobService.handlers();
        if (res.code !== 0 && res.code !== 200)
            throw new Error(res.msg);
        handlers.value = res.data || [];
    }
    catch (e) {
        ElMessage.error(e?.message || t('common.loadFailed'));
    }
}
// Search and reset
function onSearch() {
    pagination.cursor = null;
    pagination.hasPrev = false;
    fetchList();
}
function onReset() {
    queryForm.jobName = '';
    queryForm.jobGroup = '';
    queryForm.status = undefined;
    pagination.cursor = null;
    pagination.hasPrev = false;
    fetchList();
}
// Pagination
function nextPage() {
    paginationNextPage();
    fetchList();
}
function prevPage() {
    paginationPrevPage();
    fetchList();
}
// Dialog operations
function handleCreate() {
    isEdit.value = false;
    resetForm();
    dialogVisible.value = true;
}
function handleEdit(row) {
    isEdit.value = true;
    formData.id = row.id;
    formData.jobName = row.jobName;
    formData.jobGroup = row.jobGroup;
    formData.invokeTarget = row.invokeTarget;
    formData.cronExpression = row.cronExpression;
    formData.status = row.status;
    formData.remark = row.remark || '';
    dialogVisible.value = true;
}
// Submit
async function handleSubmit() {
    if (!formRef.value)
        return;
    try {
        await formRef.value.validate();
        submitLoading.value = true;
        try {
            let res;
            if (isEdit.value) {
                res = await cronJobService.update(formData.id, formData);
            }
            else {
                res = await cronJobService.create(formData);
            }
            if (res.code !== 0 && res.code !== 200)
                throw new Error(res.msg);
            ElMessage.success(isEdit.value ? t('common.updateSuccess') : t('common.createSuccess'));
            dialogVisible.value = false;
            pagination.cursor = null;
            pagination.hasPrev = false;
            fetchList();
        }
        finally {
            submitLoading.value = false;
        }
    }
    catch (e) {
        ElMessage.error(e?.message || 'Error');
    }
}
// Delete
async function handleDelete(row) {
    try {
        await confirm(`${t('common.confirmDelete')} - ${row.jobName}?`);
        const res = await cronJobService.delete(row.id);
        if (res.code !== 0 && res.code !== 200)
            throw new Error(res.msg);
        ElMessage.success(t('common.success'));
        fetchList();
    }
    catch (e) {
        if (e.message !== 'Cancel') {
            ElMessage.error(e?.message || t('common.failed'));
        }
    }
}
// Run task
async function handleRun(row) {
    try {
        await confirm(`${t('system.runTask')} - ${row.jobName}?`);
        const res = await cronJobService.run(row.id);
        if (res.code !== 0 && res.code !== 200)
            throw new Error(res.msg);
        ElMessage.success(t('common.success'));
    }
    catch (e) {
        if (e.message !== 'Cancel') {
            ElMessage.error(e?.message || t('common.failed'));
        }
    }
}
// Start task
async function handleStart(row) {
    try {
        const res = await cronJobService.start(row.id);
        if (res.code !== 0 && res.code !== 200)
            throw new Error(res.msg);
        ElMessage.success(t('common.success'));
        fetchList();
    }
    catch (e) {
        ElMessage.error(e?.message || t('common.failed'));
    }
}
// Stop task
async function handleStop(row) {
    try {
        const res = await cronJobService.stop(row.id);
        if (res.code !== 0 && res.code !== 200)
            throw new Error(res.msg);
        ElMessage.success(t('common.success'));
        fetchList();
    }
    catch (e) {
        ElMessage.error(e?.message || t('common.failed'));
    }
}
// Format date
const formatDateFn = (date) => {
    if (!date)
        return '-';
    return formatDate(new Date(date * 1000).getTime());
};
// Load on mount
onMounted(() => {
    fetchHandlers();
    fetchList();
});
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
// CSS variable injection 
// CSS variable injection end 
const __VLS_0 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({}));
const __VLS_2 = __VLS_1({}, ...__VLS_functionalComponentArgsRest(__VLS_1));
var __VLS_4 = {};
__VLS_3.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_3.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (__VLS_ctx.t('system.cronJobs'));
    const __VLS_5 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_6 = __VLS_asFunctionalComponent(__VLS_5, new __VLS_5({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_7 = __VLS_6({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_6));
    let __VLS_9;
    let __VLS_10;
    let __VLS_11;
    const __VLS_12 = {
        onClick: (__VLS_ctx.handleCreate)
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPerm)(null, { ...__VLS_directiveBindingRestFields, value: ('sys:job:add') }, null, null);
    __VLS_8.slots.default;
    const __VLS_13 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_14 = __VLS_asFunctionalComponent(__VLS_13, new __VLS_13({}));
    const __VLS_15 = __VLS_14({}, ...__VLS_functionalComponentArgsRest(__VLS_14));
    __VLS_16.slots.default;
    const __VLS_17 = {}.Plus;
    /** @type {[typeof __VLS_components.Plus, ]} */ ;
    // @ts-ignore
    const __VLS_18 = __VLS_asFunctionalComponent(__VLS_17, new __VLS_17({}));
    const __VLS_19 = __VLS_18({}, ...__VLS_functionalComponentArgsRest(__VLS_18));
    var __VLS_16;
    (__VLS_ctx.t('common.add'));
    var __VLS_8;
}
const __VLS_21 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_22 = __VLS_asFunctionalComponent(__VLS_21, new __VLS_21({
    model: (__VLS_ctx.queryForm),
    inline: true,
    ...{ style: {} },
}));
const __VLS_23 = __VLS_22({
    model: (__VLS_ctx.queryForm),
    inline: true,
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_22));
__VLS_24.slots.default;
const __VLS_25 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_26 = __VLS_asFunctionalComponent(__VLS_25, new __VLS_25({
    label: (__VLS_ctx.t('system.jobName')),
}));
const __VLS_27 = __VLS_26({
    label: (__VLS_ctx.t('system.jobName')),
}, ...__VLS_functionalComponentArgsRest(__VLS_26));
__VLS_28.slots.default;
const __VLS_29 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_30 = __VLS_asFunctionalComponent(__VLS_29, new __VLS_29({
    modelValue: (__VLS_ctx.queryForm.jobName),
    placeholder: (__VLS_ctx.t('common.pleaseEnter')),
    clearable: true,
    ...{ style: {} },
}));
const __VLS_31 = __VLS_30({
    modelValue: (__VLS_ctx.queryForm.jobName),
    placeholder: (__VLS_ctx.t('common.pleaseEnter')),
    clearable: true,
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_30));
var __VLS_28;
const __VLS_33 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_34 = __VLS_asFunctionalComponent(__VLS_33, new __VLS_33({
    label: (__VLS_ctx.t('system.jobGroup')),
}));
const __VLS_35 = __VLS_34({
    label: (__VLS_ctx.t('system.jobGroup')),
}, ...__VLS_functionalComponentArgsRest(__VLS_34));
__VLS_36.slots.default;
const __VLS_37 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_38 = __VLS_asFunctionalComponent(__VLS_37, new __VLS_37({
    modelValue: (__VLS_ctx.queryForm.jobGroup),
    placeholder: (__VLS_ctx.t('common.pleaseEnter')),
    clearable: true,
    ...{ style: {} },
}));
const __VLS_39 = __VLS_38({
    modelValue: (__VLS_ctx.queryForm.jobGroup),
    placeholder: (__VLS_ctx.t('common.pleaseEnter')),
    clearable: true,
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_38));
var __VLS_36;
const __VLS_41 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_42 = __VLS_asFunctionalComponent(__VLS_41, new __VLS_41({
    label: (__VLS_ctx.t('common.status')),
}));
const __VLS_43 = __VLS_42({
    label: (__VLS_ctx.t('common.status')),
}, ...__VLS_functionalComponentArgsRest(__VLS_42));
__VLS_44.slots.default;
const __VLS_45 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_46 = __VLS_asFunctionalComponent(__VLS_45, new __VLS_45({
    modelValue: (__VLS_ctx.queryForm.status),
    placeholder: (__VLS_ctx.t('common.pleaseSelect')),
    clearable: true,
    ...{ style: {} },
}));
const __VLS_47 = __VLS_46({
    modelValue: (__VLS_ctx.queryForm.status),
    placeholder: (__VLS_ctx.t('common.pleaseSelect')),
    clearable: true,
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_46));
__VLS_48.slots.default;
const __VLS_49 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_50 = __VLS_asFunctionalComponent(__VLS_49, new __VLS_49({
    label: (__VLS_ctx.t('common.enabled')),
    value: (1),
}));
const __VLS_51 = __VLS_50({
    label: (__VLS_ctx.t('common.enabled')),
    value: (1),
}, ...__VLS_functionalComponentArgsRest(__VLS_50));
const __VLS_53 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_54 = __VLS_asFunctionalComponent(__VLS_53, new __VLS_53({
    label: (__VLS_ctx.t('common.disabled')),
    value: (0),
}));
const __VLS_55 = __VLS_54({
    label: (__VLS_ctx.t('common.disabled')),
    value: (0),
}, ...__VLS_functionalComponentArgsRest(__VLS_54));
var __VLS_48;
var __VLS_44;
const __VLS_57 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_58 = __VLS_asFunctionalComponent(__VLS_57, new __VLS_57({}));
const __VLS_59 = __VLS_58({}, ...__VLS_functionalComponentArgsRest(__VLS_58));
__VLS_60.slots.default;
const __VLS_61 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_62 = __VLS_asFunctionalComponent(__VLS_61, new __VLS_61({
    ...{ 'onClick': {} },
    type: "primary",
}));
const __VLS_63 = __VLS_62({
    ...{ 'onClick': {} },
    type: "primary",
}, ...__VLS_functionalComponentArgsRest(__VLS_62));
let __VLS_65;
let __VLS_66;
let __VLS_67;
const __VLS_68 = {
    onClick: (__VLS_ctx.onSearch)
};
__VLS_64.slots.default;
(__VLS_ctx.t('common.search'));
var __VLS_64;
const __VLS_69 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_70 = __VLS_asFunctionalComponent(__VLS_69, new __VLS_69({
    ...{ 'onClick': {} },
}));
const __VLS_71 = __VLS_70({
    ...{ 'onClick': {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_70));
let __VLS_73;
let __VLS_74;
let __VLS_75;
const __VLS_76 = {
    onClick: (__VLS_ctx.onReset)
};
__VLS_72.slots.default;
(__VLS_ctx.t('common.reset'));
var __VLS_72;
var __VLS_60;
var __VLS_24;
const __VLS_77 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_78 = __VLS_asFunctionalComponent(__VLS_77, new __VLS_77({
    data: (__VLS_ctx.list),
    rowKey: "id",
    ...{ style: {} },
}));
const __VLS_79 = __VLS_78({
    data: (__VLS_ctx.list),
    rowKey: "id",
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_78));
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.loading) }, null, null);
__VLS_80.slots.default;
const __VLS_81 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_82 = __VLS_asFunctionalComponent(__VLS_81, new __VLS_81({
    prop: "id",
    label: (__VLS_ctx.t('common.id')),
    width: "70",
}));
const __VLS_83 = __VLS_82({
    prop: "id",
    label: (__VLS_ctx.t('common.id')),
    width: "70",
}, ...__VLS_functionalComponentArgsRest(__VLS_82));
const __VLS_85 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_86 = __VLS_asFunctionalComponent(__VLS_85, new __VLS_85({
    prop: "jobName",
    label: (__VLS_ctx.t('system.jobName')),
    minWidth: "120",
}));
const __VLS_87 = __VLS_86({
    prop: "jobName",
    label: (__VLS_ctx.t('system.jobName')),
    minWidth: "120",
}, ...__VLS_functionalComponentArgsRest(__VLS_86));
const __VLS_89 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_90 = __VLS_asFunctionalComponent(__VLS_89, new __VLS_89({
    prop: "jobGroup",
    label: (__VLS_ctx.t('system.jobGroup')),
    width: "100",
}));
const __VLS_91 = __VLS_90({
    prop: "jobGroup",
    label: (__VLS_ctx.t('system.jobGroup')),
    width: "100",
}, ...__VLS_functionalComponentArgsRest(__VLS_90));
const __VLS_93 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_94 = __VLS_asFunctionalComponent(__VLS_93, new __VLS_93({
    prop: "invokeTarget",
    label: (__VLS_ctx.t('system.invokeTarget')),
    minWidth: "180",
}));
const __VLS_95 = __VLS_94({
    prop: "invokeTarget",
    label: (__VLS_ctx.t('system.invokeTarget')),
    minWidth: "180",
}, ...__VLS_functionalComponentArgsRest(__VLS_94));
const __VLS_97 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_98 = __VLS_asFunctionalComponent(__VLS_97, new __VLS_97({
    prop: "cronExpression",
    label: (__VLS_ctx.t('system.cronExpression')),
    width: "140",
}));
const __VLS_99 = __VLS_98({
    prop: "cronExpression",
    label: (__VLS_ctx.t('system.cronExpression')),
    width: "140",
}, ...__VLS_functionalComponentArgsRest(__VLS_98));
const __VLS_101 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_102 = __VLS_asFunctionalComponent(__VLS_101, new __VLS_101({
    prop: "status",
    label: (__VLS_ctx.t('common.status')),
    width: "100",
}));
const __VLS_103 = __VLS_102({
    prop: "status",
    label: (__VLS_ctx.t('common.status')),
    width: "100",
}, ...__VLS_functionalComponentArgsRest(__VLS_102));
__VLS_104.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_104.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_105 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_106 = __VLS_asFunctionalComponent(__VLS_105, new __VLS_105({
        type: (row.status === 1 ? 'success' : 'info'),
    }));
    const __VLS_107 = __VLS_106({
        type: (row.status === 1 ? 'success' : 'info'),
    }, ...__VLS_functionalComponentArgsRest(__VLS_106));
    __VLS_108.slots.default;
    (row.status === 1 ? __VLS_ctx.t('common.enabled') : __VLS_ctx.t('common.disabled'));
    var __VLS_108;
}
var __VLS_104;
const __VLS_109 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_110 = __VLS_asFunctionalComponent(__VLS_109, new __VLS_109({
    prop: "createBy",
    label: (__VLS_ctx.t('system.createBy')),
    width: "100",
}));
const __VLS_111 = __VLS_110({
    prop: "createBy",
    label: (__VLS_ctx.t('system.createBy')),
    width: "100",
}, ...__VLS_functionalComponentArgsRest(__VLS_110));
const __VLS_113 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_114 = __VLS_asFunctionalComponent(__VLS_113, new __VLS_113({
    prop: "createTime",
    label: (__VLS_ctx.t('common.createdAt')),
    width: "170",
}));
const __VLS_115 = __VLS_114({
    prop: "createTime",
    label: (__VLS_ctx.t('common.createdAt')),
    width: "170",
}, ...__VLS_functionalComponentArgsRest(__VLS_114));
__VLS_116.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_116.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (__VLS_ctx.formatDateFn(row.createTime));
}
var __VLS_116;
const __VLS_117 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_118 = __VLS_asFunctionalComponent(__VLS_117, new __VLS_117({
    label: (__VLS_ctx.t('common.actions')),
    width: "280",
    fixed: "right",
}));
const __VLS_119 = __VLS_118({
    label: (__VLS_ctx.t('common.actions')),
    width: "280",
    fixed: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_118));
__VLS_120.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_120.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_121 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_122 = __VLS_asFunctionalComponent(__VLS_121, new __VLS_121({
        ...{ 'onClick': {} },
        type: "primary",
        size: "small",
    }));
    const __VLS_123 = __VLS_122({
        ...{ 'onClick': {} },
        type: "primary",
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_122));
    let __VLS_125;
    let __VLS_126;
    let __VLS_127;
    const __VLS_128 = {
        onClick: (...[$event]) => {
            __VLS_ctx.handleRun(row);
        }
    };
    __VLS_124.slots.default;
    const __VLS_129 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_130 = __VLS_asFunctionalComponent(__VLS_129, new __VLS_129({}));
    const __VLS_131 = __VLS_130({}, ...__VLS_functionalComponentArgsRest(__VLS_130));
    __VLS_132.slots.default;
    const __VLS_133 = {}.VideoPlay;
    /** @type {[typeof __VLS_components.VideoPlay, ]} */ ;
    // @ts-ignore
    const __VLS_134 = __VLS_asFunctionalComponent(__VLS_133, new __VLS_133({}));
    const __VLS_135 = __VLS_134({}, ...__VLS_functionalComponentArgsRest(__VLS_134));
    var __VLS_132;
    (__VLS_ctx.t('system.run'));
    var __VLS_124;
    if (row.status === 0) {
        const __VLS_137 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_138 = __VLS_asFunctionalComponent(__VLS_137, new __VLS_137({
            ...{ 'onClick': {} },
            type: "success",
            size: "small",
        }));
        const __VLS_139 = __VLS_138({
            ...{ 'onClick': {} },
            type: "success",
            size: "small",
        }, ...__VLS_functionalComponentArgsRest(__VLS_138));
        let __VLS_141;
        let __VLS_142;
        let __VLS_143;
        const __VLS_144 = {
            onClick: (...[$event]) => {
                if (!(row.status === 0))
                    return;
                __VLS_ctx.handleStart(row);
            }
        };
        __VLS_140.slots.default;
        const __VLS_145 = {}.ElIcon;
        /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
        // @ts-ignore
        const __VLS_146 = __VLS_asFunctionalComponent(__VLS_145, new __VLS_145({}));
        const __VLS_147 = __VLS_146({}, ...__VLS_functionalComponentArgsRest(__VLS_146));
        __VLS_148.slots.default;
        const __VLS_149 = {}.CircleCheck;
        /** @type {[typeof __VLS_components.CircleCheck, ]} */ ;
        // @ts-ignore
        const __VLS_150 = __VLS_asFunctionalComponent(__VLS_149, new __VLS_149({}));
        const __VLS_151 = __VLS_150({}, ...__VLS_functionalComponentArgsRest(__VLS_150));
        var __VLS_148;
        (__VLS_ctx.t('system.start'));
        var __VLS_140;
    }
    if (row.status === 1) {
        const __VLS_153 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_154 = __VLS_asFunctionalComponent(__VLS_153, new __VLS_153({
            ...{ 'onClick': {} },
            type: "warning",
            size: "small",
        }));
        const __VLS_155 = __VLS_154({
            ...{ 'onClick': {} },
            type: "warning",
            size: "small",
        }, ...__VLS_functionalComponentArgsRest(__VLS_154));
        let __VLS_157;
        let __VLS_158;
        let __VLS_159;
        const __VLS_160 = {
            onClick: (...[$event]) => {
                if (!(row.status === 1))
                    return;
                __VLS_ctx.handleStop(row);
            }
        };
        __VLS_156.slots.default;
        const __VLS_161 = {}.ElIcon;
        /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
        // @ts-ignore
        const __VLS_162 = __VLS_asFunctionalComponent(__VLS_161, new __VLS_161({}));
        const __VLS_163 = __VLS_162({}, ...__VLS_functionalComponentArgsRest(__VLS_162));
        __VLS_164.slots.default;
        const __VLS_165 = {}.CircleCloseFilled;
        /** @type {[typeof __VLS_components.CircleCloseFilled, ]} */ ;
        // @ts-ignore
        const __VLS_166 = __VLS_asFunctionalComponent(__VLS_165, new __VLS_165({}));
        const __VLS_167 = __VLS_166({}, ...__VLS_functionalComponentArgsRest(__VLS_166));
        var __VLS_164;
        (__VLS_ctx.t('system.stop'));
        var __VLS_156;
    }
    const __VLS_169 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_170 = __VLS_asFunctionalComponent(__VLS_169, new __VLS_169({
        ...{ 'onClick': {} },
        type: "primary",
        size: "small",
    }));
    const __VLS_171 = __VLS_170({
        ...{ 'onClick': {} },
        type: "primary",
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_170));
    let __VLS_173;
    let __VLS_174;
    let __VLS_175;
    const __VLS_176 = {
        onClick: (...[$event]) => {
            __VLS_ctx.handleEdit(row);
        }
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPerm)(null, { ...__VLS_directiveBindingRestFields, value: ('sys:job:update') }, null, null);
    __VLS_172.slots.default;
    const __VLS_177 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_178 = __VLS_asFunctionalComponent(__VLS_177, new __VLS_177({}));
    const __VLS_179 = __VLS_178({}, ...__VLS_functionalComponentArgsRest(__VLS_178));
    __VLS_180.slots.default;
    const __VLS_181 = {}.Edit;
    /** @type {[typeof __VLS_components.Edit, ]} */ ;
    // @ts-ignore
    const __VLS_182 = __VLS_asFunctionalComponent(__VLS_181, new __VLS_181({}));
    const __VLS_183 = __VLS_182({}, ...__VLS_functionalComponentArgsRest(__VLS_182));
    var __VLS_180;
    (__VLS_ctx.t('common.edit'));
    var __VLS_172;
    const __VLS_185 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_186 = __VLS_asFunctionalComponent(__VLS_185, new __VLS_185({
        ...{ 'onClick': {} },
        type: "danger",
        size: "small",
    }));
    const __VLS_187 = __VLS_186({
        ...{ 'onClick': {} },
        type: "danger",
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_186));
    let __VLS_189;
    let __VLS_190;
    let __VLS_191;
    const __VLS_192 = {
        onClick: (...[$event]) => {
            __VLS_ctx.handleDelete(row);
        }
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPerm)(null, { ...__VLS_directiveBindingRestFields, value: ('sys:job:delete') }, null, null);
    __VLS_188.slots.default;
    const __VLS_193 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_194 = __VLS_asFunctionalComponent(__VLS_193, new __VLS_193({}));
    const __VLS_195 = __VLS_194({}, ...__VLS_functionalComponentArgsRest(__VLS_194));
    __VLS_196.slots.default;
    const __VLS_197 = {}.Delete;
    /** @type {[typeof __VLS_components.Delete, ]} */ ;
    // @ts-ignore
    const __VLS_198 = __VLS_asFunctionalComponent(__VLS_197, new __VLS_197({}));
    const __VLS_199 = __VLS_198({}, ...__VLS_functionalComponentArgsRest(__VLS_198));
    var __VLS_196;
    (__VLS_ctx.t('common.delete'));
    var __VLS_188;
}
var __VLS_120;
var __VLS_80;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ style: {} },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
(__VLS_ctx.t('common.totalItems', { count: __VLS_ctx.pagination.total }));
const __VLS_201 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_202 = __VLS_asFunctionalComponent(__VLS_201, new __VLS_201({
    ...{ 'onClick': {} },
    disabled: (!__VLS_ctx.pagination.hasPrev),
}));
const __VLS_203 = __VLS_202({
    ...{ 'onClick': {} },
    disabled: (!__VLS_ctx.pagination.hasPrev),
}, ...__VLS_functionalComponentArgsRest(__VLS_202));
let __VLS_205;
let __VLS_206;
let __VLS_207;
const __VLS_208 = {
    onClick: (__VLS_ctx.prevPage)
};
__VLS_204.slots.default;
(__VLS_ctx.t('common.prevPage'));
var __VLS_204;
const __VLS_209 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_210 = __VLS_asFunctionalComponent(__VLS_209, new __VLS_209({
    ...{ 'onClick': {} },
    disabled: (!__VLS_ctx.pagination.hasNext),
}));
const __VLS_211 = __VLS_210({
    ...{ 'onClick': {} },
    disabled: (!__VLS_ctx.pagination.hasNext),
}, ...__VLS_functionalComponentArgsRest(__VLS_210));
let __VLS_213;
let __VLS_214;
let __VLS_215;
const __VLS_216 = {
    onClick: (__VLS_ctx.nextPage)
};
__VLS_212.slots.default;
(__VLS_ctx.t('common.nextPage'));
var __VLS_212;
const __VLS_217 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_218 = __VLS_asFunctionalComponent(__VLS_217, new __VLS_217({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.pagination.limit),
    ...{ style: {} },
}));
const __VLS_219 = __VLS_218({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.pagination.limit),
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_218));
let __VLS_221;
let __VLS_222;
let __VLS_223;
const __VLS_224 = {
    onChange: (() => {
        __VLS_ctx.pagination.cursor = null;
        __VLS_ctx.pagination.hasPrev = false;
        __VLS_ctx.fetchList();
    })
};
__VLS_220.slots.default;
const __VLS_225 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_226 = __VLS_asFunctionalComponent(__VLS_225, new __VLS_225({
    label: "10",
    value: (10),
}));
const __VLS_227 = __VLS_226({
    label: "10",
    value: (10),
}, ...__VLS_functionalComponentArgsRest(__VLS_226));
const __VLS_229 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_230 = __VLS_asFunctionalComponent(__VLS_229, new __VLS_229({
    label: "20",
    value: (20),
}));
const __VLS_231 = __VLS_230({
    label: "20",
    value: (20),
}, ...__VLS_functionalComponentArgsRest(__VLS_230));
const __VLS_233 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_234 = __VLS_asFunctionalComponent(__VLS_233, new __VLS_233({
    label: "50",
    value: (50),
}));
const __VLS_235 = __VLS_234({
    label: "50",
    value: (50),
}, ...__VLS_functionalComponentArgsRest(__VLS_234));
var __VLS_220;
const __VLS_237 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_238 = __VLS_asFunctionalComponent(__VLS_237, new __VLS_237({
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.isEdit ? __VLS_ctx.t('common.edit') : __VLS_ctx.t('common.add')),
    width: "600px",
    closeOnClickModal: (false),
}));
const __VLS_239 = __VLS_238({
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.isEdit ? __VLS_ctx.t('common.edit') : __VLS_ctx.t('common.add')),
    width: "600px",
    closeOnClickModal: (false),
}, ...__VLS_functionalComponentArgsRest(__VLS_238));
__VLS_240.slots.default;
const __VLS_241 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_242 = __VLS_asFunctionalComponent(__VLS_241, new __VLS_241({
    ref: "formRef",
    model: (__VLS_ctx.formData),
    rules: (__VLS_ctx.formRules),
    labelWidth: "140px",
}));
const __VLS_243 = __VLS_242({
    ref: "formRef",
    model: (__VLS_ctx.formData),
    rules: (__VLS_ctx.formRules),
    labelWidth: "140px",
}, ...__VLS_functionalComponentArgsRest(__VLS_242));
/** @type {typeof __VLS_ctx.formRef} */ ;
var __VLS_245 = {};
__VLS_244.slots.default;
const __VLS_247 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_248 = __VLS_asFunctionalComponent(__VLS_247, new __VLS_247({
    label: (__VLS_ctx.t('system.jobName')),
    prop: "jobName",
}));
const __VLS_249 = __VLS_248({
    label: (__VLS_ctx.t('system.jobName')),
    prop: "jobName",
}, ...__VLS_functionalComponentArgsRest(__VLS_248));
__VLS_250.slots.default;
const __VLS_251 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_252 = __VLS_asFunctionalComponent(__VLS_251, new __VLS_251({
    modelValue: (__VLS_ctx.formData.jobName),
    placeholder: (__VLS_ctx.t('common.pleaseEnter')),
    maxlength: "100",
}));
const __VLS_253 = __VLS_252({
    modelValue: (__VLS_ctx.formData.jobName),
    placeholder: (__VLS_ctx.t('common.pleaseEnter')),
    maxlength: "100",
}, ...__VLS_functionalComponentArgsRest(__VLS_252));
var __VLS_250;
const __VLS_255 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_256 = __VLS_asFunctionalComponent(__VLS_255, new __VLS_255({
    label: (__VLS_ctx.t('system.jobGroup')),
    prop: "jobGroup",
}));
const __VLS_257 = __VLS_256({
    label: (__VLS_ctx.t('system.jobGroup')),
    prop: "jobGroup",
}, ...__VLS_functionalComponentArgsRest(__VLS_256));
__VLS_258.slots.default;
const __VLS_259 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_260 = __VLS_asFunctionalComponent(__VLS_259, new __VLS_259({
    modelValue: (__VLS_ctx.formData.jobGroup),
    placeholder: (__VLS_ctx.t('common.pleaseEnter')),
    maxlength: "50",
}));
const __VLS_261 = __VLS_260({
    modelValue: (__VLS_ctx.formData.jobGroup),
    placeholder: (__VLS_ctx.t('common.pleaseEnter')),
    maxlength: "50",
}, ...__VLS_functionalComponentArgsRest(__VLS_260));
var __VLS_258;
const __VLS_263 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_264 = __VLS_asFunctionalComponent(__VLS_263, new __VLS_263({
    label: (__VLS_ctx.t('system.invokeTarget')),
    prop: "invokeTarget",
}));
const __VLS_265 = __VLS_264({
    label: (__VLS_ctx.t('system.invokeTarget')),
    prop: "invokeTarget",
}, ...__VLS_functionalComponentArgsRest(__VLS_264));
__VLS_266.slots.default;
const __VLS_267 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_268 = __VLS_asFunctionalComponent(__VLS_267, new __VLS_267({
    modelValue: (__VLS_ctx.formData.invokeTarget),
    placeholder: (__VLS_ctx.t('common.pleaseSelect')),
    filterable: true,
    clearable: true,
}));
const __VLS_269 = __VLS_268({
    modelValue: (__VLS_ctx.formData.invokeTarget),
    placeholder: (__VLS_ctx.t('common.pleaseSelect')),
    filterable: true,
    clearable: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_268));
__VLS_270.slots.default;
for (const [handler] of __VLS_getVForSourceType((__VLS_ctx.handlers))) {
    const __VLS_271 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_272 = __VLS_asFunctionalComponent(__VLS_271, new __VLS_271({
        key: (handler.invokeTarget),
        label: (`${handler.jobName} (${handler.invokeTarget})`),
        value: (handler.invokeTarget),
    }));
    const __VLS_273 = __VLS_272({
        key: (handler.invokeTarget),
        label: (`${handler.jobName} (${handler.invokeTarget})`),
        value: (handler.invokeTarget),
    }, ...__VLS_functionalComponentArgsRest(__VLS_272));
}
var __VLS_270;
var __VLS_266;
const __VLS_275 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_276 = __VLS_asFunctionalComponent(__VLS_275, new __VLS_275({
    label: (__VLS_ctx.t('system.cronExpression')),
    prop: "cronExpression",
}));
const __VLS_277 = __VLS_276({
    label: (__VLS_ctx.t('system.cronExpression')),
    prop: "cronExpression",
}, ...__VLS_functionalComponentArgsRest(__VLS_276));
__VLS_278.slots.default;
const __VLS_279 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_280 = __VLS_asFunctionalComponent(__VLS_279, new __VLS_279({
    modelValue: (__VLS_ctx.formData.cronExpression),
    placeholder: (__VLS_ctx.t('system.cronExpressionPlaceholder')),
    maxlength: "100",
}));
const __VLS_281 = __VLS_280({
    modelValue: (__VLS_ctx.formData.cronExpression),
    placeholder: (__VLS_ctx.t('system.cronExpressionPlaceholder')),
    maxlength: "100",
}, ...__VLS_functionalComponentArgsRest(__VLS_280));
var __VLS_278;
const __VLS_283 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_284 = __VLS_asFunctionalComponent(__VLS_283, new __VLS_283({
    label: (__VLS_ctx.t('common.status')),
}));
const __VLS_285 = __VLS_284({
    label: (__VLS_ctx.t('common.status')),
}, ...__VLS_functionalComponentArgsRest(__VLS_284));
__VLS_286.slots.default;
const __VLS_287 = {}.ElRadioGroup;
/** @type {[typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, ]} */ ;
// @ts-ignore
const __VLS_288 = __VLS_asFunctionalComponent(__VLS_287, new __VLS_287({
    modelValue: (__VLS_ctx.formData.status),
}));
const __VLS_289 = __VLS_288({
    modelValue: (__VLS_ctx.formData.status),
}, ...__VLS_functionalComponentArgsRest(__VLS_288));
__VLS_290.slots.default;
const __VLS_291 = {}.ElRadio;
/** @type {[typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, ]} */ ;
// @ts-ignore
const __VLS_292 = __VLS_asFunctionalComponent(__VLS_291, new __VLS_291({
    label: (1),
}));
const __VLS_293 = __VLS_292({
    label: (1),
}, ...__VLS_functionalComponentArgsRest(__VLS_292));
__VLS_294.slots.default;
(__VLS_ctx.t('common.enabled'));
var __VLS_294;
const __VLS_295 = {}.ElRadio;
/** @type {[typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, ]} */ ;
// @ts-ignore
const __VLS_296 = __VLS_asFunctionalComponent(__VLS_295, new __VLS_295({
    label: (0),
}));
const __VLS_297 = __VLS_296({
    label: (0),
}, ...__VLS_functionalComponentArgsRest(__VLS_296));
__VLS_298.slots.default;
(__VLS_ctx.t('common.disabled'));
var __VLS_298;
var __VLS_290;
var __VLS_286;
const __VLS_299 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_300 = __VLS_asFunctionalComponent(__VLS_299, new __VLS_299({
    label: (__VLS_ctx.t('common.remark')),
}));
const __VLS_301 = __VLS_300({
    label: (__VLS_ctx.t('common.remark')),
}, ...__VLS_functionalComponentArgsRest(__VLS_300));
__VLS_302.slots.default;
const __VLS_303 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_304 = __VLS_asFunctionalComponent(__VLS_303, new __VLS_303({
    modelValue: (__VLS_ctx.formData.remark),
    placeholder: (__VLS_ctx.t('common.pleaseEnter')),
    type: "textarea",
    rows: (3),
    maxlength: "500",
}));
const __VLS_305 = __VLS_304({
    modelValue: (__VLS_ctx.formData.remark),
    placeholder: (__VLS_ctx.t('common.pleaseEnter')),
    type: "textarea",
    rows: (3),
    maxlength: "500",
}, ...__VLS_functionalComponentArgsRest(__VLS_304));
var __VLS_302;
var __VLS_244;
{
    const { footer: __VLS_thisSlot } = __VLS_240.slots;
    const __VLS_307 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_308 = __VLS_asFunctionalComponent(__VLS_307, new __VLS_307({
        ...{ 'onClick': {} },
    }));
    const __VLS_309 = __VLS_308({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_308));
    let __VLS_311;
    let __VLS_312;
    let __VLS_313;
    const __VLS_314 = {
        onClick: (...[$event]) => {
            __VLS_ctx.dialogVisible = false;
        }
    };
    __VLS_310.slots.default;
    (__VLS_ctx.t('common.cancel'));
    var __VLS_310;
    const __VLS_315 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_316 = __VLS_asFunctionalComponent(__VLS_315, new __VLS_315({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.submitLoading),
    }));
    const __VLS_317 = __VLS_316({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.submitLoading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_316));
    let __VLS_319;
    let __VLS_320;
    let __VLS_321;
    const __VLS_322 = {
        onClick: (__VLS_ctx.handleSubmit)
    };
    __VLS_318.slots.default;
    (__VLS_ctx.t('common.confirm'));
    var __VLS_318;
}
var __VLS_240;
var __VLS_3;
// @ts-ignore
var __VLS_246 = __VLS_245;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            Plus: Plus,
            Edit: Edit,
            Delete: Delete,
            VideoPlay: VideoPlay,
            CircleCheck: CircleCheck,
            CircleCloseFilled: CircleCloseFilled,
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
            handlers: handlers,
            formRules: formRules,
            fetchList: fetchList,
            onSearch: onSearch,
            onReset: onReset,
            nextPage: nextPage,
            prevPage: prevPage,
            handleCreate: handleCreate,
            handleEdit: handleEdit,
            handleSubmit: handleSubmit,
            handleDelete: handleDelete,
            handleRun: handleRun,
            handleStart: handleStart,
            handleStop: handleStop,
            formatDateFn: formatDateFn,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
