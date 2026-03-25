import { computed, nextTick, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { ElMessage } from 'element-plus';
import * as ElementPlusIconsVue from '@element-plus/icons-vue';
import { menuService } from '@/services';
import { useLoading } from '@/composables/useLoading';
import { useForm } from '@/composables/useForm';
import { useConfirm } from '@/composables/useConfirm';
const { t, te } = useI18n();
const iconMap = ElementPlusIconsVue;
const iconNames = Object.keys(iconMap).sort();
// Composables
const { loading, withLoading: withMainLoading } = useLoading();
const { loading: submitLoading, withLoading: withSubmitLoading } = useLoading();
const { form: queryForm } = useForm({
    initialData: {
        keyword: '',
        menuType: undefined,
        status: undefined,
        visible: undefined,
    },
});
const { confirm } = useConfirm();
// Dialog state
const dialogVisible = ref(false);
const dialogType = ref('add');
const formRef = ref();
// Query pagination
const queryPage = { cursor: null, limit: 1000 };
// Menu tree data
const rawList = ref([]);
const tableData = ref([]);
const createDefaultForm = () => ({
    id: undefined,
    parentId: 0,
    name: '',
    menuType: 1,
    path: '',
    component: '',
    icon: '',
    sort: 0,
    visible: 1,
    status: 1,
    perms: '',
});
const { form: formData } = useForm({
    initialData: createDefaultForm(),
});
const currentEditId = computed(() => formData.id ?? 0);
const currentIconComponent = computed(() => {
    return resolveIconComponent(formData.icon);
});
const childrenIdMap = computed(() => {
    const map = new Map();
    const dfs = (node) => {
        const ids = [];
        for (const child of node.children || []) {
            ids.push(child.id);
            ids.push(...dfs(child));
        }
        map.set(node.id, ids);
        return ids;
    };
    for (const node of tableData.value) {
        dfs(node);
    }
    return map;
});
const parentTreeOptions = computed(() => {
    const excludeIds = new Set();
    if (dialogType.value === 'edit' && currentEditId.value) {
        excludeIds.add(currentEditId.value);
        const childIds = childrenIdMap.value.get(currentEditId.value) || [];
        childIds.forEach((id) => excludeIds.add(id));
    }
    const filterNodes = (nodes) => {
        return nodes
            .filter((node) => node.menuType !== 3 && !excludeIds.has(node.id))
            .map((node) => ({
            ...node,
            children: filterNodes(node.children || []),
        }));
    };
    return [
        {
            id: 0,
            name: t('system.topMenu'),
            children: filterNodes(tableData.value),
        },
    ];
});
const rules = computed(() => ({
    parentId: [{ required: true, message: t('system.pleaseSelectParentMenu'), trigger: 'change' }],
    name: [{ required: true, message: t('system.pleaseInputMenuName'), trigger: 'blur' }],
    menuType: [{ required: true, message: t('system.pleaseSelectMenuType'), trigger: 'change' }],
    path: [
        {
            validator: (_rule, value, callback) => {
                if (formData.menuType !== 3 && !String(value || '').trim()) {
                    callback(new Error(t('system.pleaseInputPath')));
                    return;
                }
                callback();
            },
            trigger: 'blur',
        },
    ],
    component: [
        {
            validator: (_rule, value, callback) => {
                if (formData.menuType === 2 && !String(value || '').trim()) {
                    callback(new Error(t('system.pleaseInputComponent')));
                    return;
                }
                callback();
            },
            trigger: 'blur',
        },
    ],
    perms: [
        {
            validator: (_rule, value, callback) => {
                if (formData.menuType === 3 && !String(value || '').trim()) {
                    callback(new Error(t('system.pleaseInputPerms')));
                    return;
                }
                callback();
            },
            trigger: 'blur',
        },
    ],
    sort: [{ required: true, message: t('system.pleaseInputSort'), trigger: 'blur' }],
}));
function resolveIconComponent(iconName) {
    if (!iconName)
        return null;
    return iconMap[iconName] || null;
}
function selectIcon(iconName) {
    formData.icon = iconName;
}
function clearIcon() {
    formData.icon = '';
}
function getMenuTitle(row) {
    const key = `menu.${row.id}`;
    if (te(key))
        return t(key);
    return row.name;
}
function buildTree(list) {
    const map = new Map();
    const roots = [];
    list.forEach((item) => {
        map.set(item.id, {
            ...item,
            children: [],
        });
    });
    map.forEach((item) => {
        if (item.parentId === 0) {
            roots.push(item);
            return;
        }
        const parent = map.get(item.parentId);
        if (parent) {
            parent.children ||= [];
            parent.children.push(item);
        }
        else {
            roots.push(item);
        }
    });
    const sortTree = (nodes) => {
        nodes.sort((a, b) => a.sort - b.sort);
        nodes.forEach((node) => {
            if (node.children?.length) {
                sortTree(node.children);
            }
        });
    };
    sortTree(roots);
    return roots;
}
function resetForm() {
    Object.assign(formData, createDefaultForm());
    nextTick(() => {
        formRef.value?.clearValidate();
    });
}
function normalizeFormByType() {
    if (formData.menuType === 1) {
        formData.component = '';
        formData.perms = '';
    }
    else if (formData.menuType === 2) {
        formData.perms = '';
    }
    else if (formData.menuType === 3) {
        formData.path = '';
        formData.component = '';
        formData.icon = '';
    }
}
function handleMenuTypeChange() {
    normalizeFormByType();
    nextTick(() => {
        formRef.value?.clearValidate(['path', 'component', 'perms']);
    });
}
function isSuccessCode(code) {
    return code === 0 || code === 200;
}
function getRespCode(res) {
    return res?.code ?? res?.base?.code;
}
function getRespMsg(res) {
    return res?.msg || res?.base?.msg || t('common.failed');
}
function assertApiSuccess(res, defaultMsg) {
    const code = getRespCode(res);
    if (!isSuccessCode(code)) {
        throw new Error(getRespMsg(res) || defaultMsg || t('common.failed'));
    }
    return res?.data;
}
function showError(error) {
    const msg = error instanceof Error ? error.message : typeof error === 'string' ? error : t('common.failed');
    ElMessage.error(msg || t('common.failed'));
}
async function getList() {
    await withMainLoading(async () => {
        try {
            const res = (await menuService.getList({
                cursor: queryPage.cursor,
                limit: queryPage.limit,
                keyword: queryForm.keyword || '',
                menuType: queryForm.menuType ?? 0,
                status: queryForm.status ?? 0,
                visible: queryForm.visible ?? 0,
            }));
            const list = assertApiSuccess(res, t('common.failed'));
            rawList.value = Array.isArray(list) ? list : [];
            tableData.value = buildTree(rawList.value);
        }
        catch (error) {
            rawList.value = [];
            tableData.value = [];
            showError(error);
        }
    });
}
function handleSearch() {
    queryPage.cursor = null;
    getList();
}
function handleReset() {
    queryForm.keyword = '';
    queryForm.menuType = undefined;
    queryForm.status = undefined;
    queryForm.visible = undefined;
    queryPage.cursor = null;
    queryPage.limit = 20;
    getList();
}
function resetFormData() {
    Object.assign(formData, createDefaultForm());
    nextTick(() => {
        formRef.value?.clearValidate();
    });
}
function handleAdd(parentId = 0) {
    resetFormData();
    dialogType.value = 'add';
    formData.parentId = parentId;
    dialogVisible.value = true;
}
function handleEdit(row) {
    resetFormData();
    dialogType.value = 'edit';
    Object.assign(formData, {
        id: row.id,
        parentId: row.parentId,
        name: row.name,
        menuType: row.menuType,
        path: row.path,
        component: row.component,
        icon: row.icon,
        sort: row.sort,
        visible: row.visible,
        status: row.status,
        perms: row.perms,
    });
    dialogVisible.value = true;
    nextTick(() => {
        formRef.value?.clearValidate();
    });
}
async function handleDelete(row) {
    try {
        await confirm(t('system.confirmDeleteMenu', { name: getMenuTitle(row) }), { type: 'warning' });
        const res = await menuService.delete(row.id);
        assertApiSuccess(res, t('common.failed'));
        ElMessage.success(t('common.success'));
        await getList();
    }
    catch (error) {
        if (error === 'cancel') {
            return;
        }
        showError(error);
    }
}
async function handleSubmit() {
    normalizeFormByType();
    await formRef.value?.validate();
    await withSubmitLoading(async () => {
        try {
            if (dialogType.value === 'add') {
                const payload = {
                    parentId: formData.parentId,
                    name: formData.name.trim(),
                    menuType: formData.menuType,
                    path: formData.path.trim(),
                    component: formData.component.trim(),
                    icon: formData.icon.trim(),
                    sort: formData.sort,
                    visible: formData.visible,
                    status: formData.status,
                    perms: formData.perms.trim(),
                };
                const res = await menuService.create(payload);
                assertApiSuccess(res, t('common.failed'));
            }
            else {
                const payload = {
                    id: formData.id,
                    parentId: formData.parentId,
                    name: formData.name.trim(),
                    menuType: formData.menuType,
                    path: formData.path.trim(),
                    component: formData.component.trim(),
                    icon: formData.icon.trim(),
                    sort: formData.sort,
                    visible: formData.visible,
                    status: formData.status,
                    perms: formData.perms.trim(),
                };
                const res = await menuService.update(formData.id, payload);
                assertApiSuccess(res, t('common.failed'));
            }
            ElMessage.success(t('common.success'));
            dialogVisible.value = false;
            await getList();
        }
        catch (error) {
            showError(error);
        }
    });
}
onMounted(() => {
    getList();
});
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['icon-picker-box']} */ ;
/** @type {__VLS_StyleScopedClasses['icon-item']} */ ;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "app-container" },
});
const __VLS_0 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    shadow: "never",
    ...{ class: "mb-16" },
}));
const __VLS_2 = __VLS_1({
    shadow: "never",
    ...{ class: "mb-16" },
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
__VLS_3.slots.default;
const __VLS_4 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_5 = __VLS_asFunctionalComponent(__VLS_4, new __VLS_4({
    model: (__VLS_ctx.queryForm),
    inline: true,
    labelWidth: "80px",
}));
const __VLS_6 = __VLS_5({
    model: (__VLS_ctx.queryForm),
    inline: true,
    labelWidth: "80px",
}, ...__VLS_functionalComponentArgsRest(__VLS_5));
__VLS_7.slots.default;
const __VLS_8 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
    label: (__VLS_ctx.t('common.keyword')),
}));
const __VLS_10 = __VLS_9({
    label: (__VLS_ctx.t('common.keyword')),
}, ...__VLS_functionalComponentArgsRest(__VLS_9));
__VLS_11.slots.default;
const __VLS_12 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
    ...{ 'onKeyup': {} },
    modelValue: (__VLS_ctx.queryForm.keyword),
    placeholder: (__VLS_ctx.t('system.pleaseInputMenuName')),
    clearable: true,
    ...{ style: {} },
}));
const __VLS_14 = __VLS_13({
    ...{ 'onKeyup': {} },
    modelValue: (__VLS_ctx.queryForm.keyword),
    placeholder: (__VLS_ctx.t('system.pleaseInputMenuName')),
    clearable: true,
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_13));
let __VLS_16;
let __VLS_17;
let __VLS_18;
const __VLS_19 = {
    onKeyup: (__VLS_ctx.handleSearch)
};
var __VLS_15;
var __VLS_11;
const __VLS_20 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
    label: (__VLS_ctx.t('system.menuType')),
}));
const __VLS_22 = __VLS_21({
    label: (__VLS_ctx.t('system.menuType')),
}, ...__VLS_functionalComponentArgsRest(__VLS_21));
__VLS_23.slots.default;
const __VLS_24 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
    modelValue: (__VLS_ctx.queryForm.menuType),
    clearable: true,
    placeholder: (__VLS_ctx.t('common.all')),
    ...{ style: {} },
}));
const __VLS_26 = __VLS_25({
    modelValue: (__VLS_ctx.queryForm.menuType),
    clearable: true,
    placeholder: (__VLS_ctx.t('common.all')),
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_25));
__VLS_27.slots.default;
const __VLS_28 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({
    label: (__VLS_ctx.t('system.directory')),
    value: (1),
}));
const __VLS_30 = __VLS_29({
    label: (__VLS_ctx.t('system.directory')),
    value: (1),
}, ...__VLS_functionalComponentArgsRest(__VLS_29));
const __VLS_32 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
    label: (__VLS_ctx.t('system.menu')),
    value: (2),
}));
const __VLS_34 = __VLS_33({
    label: (__VLS_ctx.t('system.menu')),
    value: (2),
}, ...__VLS_functionalComponentArgsRest(__VLS_33));
const __VLS_36 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({
    label: (__VLS_ctx.t('system.button')),
    value: (3),
}));
const __VLS_38 = __VLS_37({
    label: (__VLS_ctx.t('system.button')),
    value: (3),
}, ...__VLS_functionalComponentArgsRest(__VLS_37));
var __VLS_27;
var __VLS_23;
const __VLS_40 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_41 = __VLS_asFunctionalComponent(__VLS_40, new __VLS_40({
    label: (__VLS_ctx.t('common.status')),
}));
const __VLS_42 = __VLS_41({
    label: (__VLS_ctx.t('common.status')),
}, ...__VLS_functionalComponentArgsRest(__VLS_41));
__VLS_43.slots.default;
const __VLS_44 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_45 = __VLS_asFunctionalComponent(__VLS_44, new __VLS_44({
    modelValue: (__VLS_ctx.queryForm.status),
    clearable: true,
    placeholder: (__VLS_ctx.t('common.all')),
    ...{ style: {} },
}));
const __VLS_46 = __VLS_45({
    modelValue: (__VLS_ctx.queryForm.status),
    clearable: true,
    placeholder: (__VLS_ctx.t('common.all')),
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_45));
__VLS_47.slots.default;
const __VLS_48 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_49 = __VLS_asFunctionalComponent(__VLS_48, new __VLS_48({
    label: (__VLS_ctx.t('common.enabled')),
    value: (1),
}));
const __VLS_50 = __VLS_49({
    label: (__VLS_ctx.t('common.enabled')),
    value: (1),
}, ...__VLS_functionalComponentArgsRest(__VLS_49));
const __VLS_52 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_53 = __VLS_asFunctionalComponent(__VLS_52, new __VLS_52({
    label: (__VLS_ctx.t('common.disabled')),
    value: (0),
}));
const __VLS_54 = __VLS_53({
    label: (__VLS_ctx.t('common.disabled')),
    value: (0),
}, ...__VLS_functionalComponentArgsRest(__VLS_53));
var __VLS_47;
var __VLS_43;
const __VLS_56 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_57 = __VLS_asFunctionalComponent(__VLS_56, new __VLS_56({
    label: (__VLS_ctx.t('common.visible')),
}));
const __VLS_58 = __VLS_57({
    label: (__VLS_ctx.t('common.visible')),
}, ...__VLS_functionalComponentArgsRest(__VLS_57));
__VLS_59.slots.default;
const __VLS_60 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_61 = __VLS_asFunctionalComponent(__VLS_60, new __VLS_60({
    modelValue: (__VLS_ctx.queryForm.visible),
    clearable: true,
    placeholder: (__VLS_ctx.t('common.all')),
    ...{ style: {} },
}));
const __VLS_62 = __VLS_61({
    modelValue: (__VLS_ctx.queryForm.visible),
    clearable: true,
    placeholder: (__VLS_ctx.t('common.all')),
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_61));
__VLS_63.slots.default;
const __VLS_64 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_65 = __VLS_asFunctionalComponent(__VLS_64, new __VLS_64({
    label: (__VLS_ctx.t('common.visible')),
    value: (1),
}));
const __VLS_66 = __VLS_65({
    label: (__VLS_ctx.t('common.visible')),
    value: (1),
}, ...__VLS_functionalComponentArgsRest(__VLS_65));
const __VLS_68 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_69 = __VLS_asFunctionalComponent(__VLS_68, new __VLS_68({
    label: (__VLS_ctx.t('common.hidden')),
    value: (0),
}));
const __VLS_70 = __VLS_69({
    label: (__VLS_ctx.t('common.hidden')),
    value: (0),
}, ...__VLS_functionalComponentArgsRest(__VLS_69));
var __VLS_63;
var __VLS_59;
const __VLS_72 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_73 = __VLS_asFunctionalComponent(__VLS_72, new __VLS_72({}));
const __VLS_74 = __VLS_73({}, ...__VLS_functionalComponentArgsRest(__VLS_73));
__VLS_75.slots.default;
const __VLS_76 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_77 = __VLS_asFunctionalComponent(__VLS_76, new __VLS_76({
    ...{ 'onClick': {} },
    type: "primary",
}));
const __VLS_78 = __VLS_77({
    ...{ 'onClick': {} },
    type: "primary",
}, ...__VLS_functionalComponentArgsRest(__VLS_77));
let __VLS_80;
let __VLS_81;
let __VLS_82;
const __VLS_83 = {
    onClick: (__VLS_ctx.handleSearch)
};
__VLS_79.slots.default;
(__VLS_ctx.t('common.search'));
var __VLS_79;
const __VLS_84 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_85 = __VLS_asFunctionalComponent(__VLS_84, new __VLS_84({
    ...{ 'onClick': {} },
}));
const __VLS_86 = __VLS_85({
    ...{ 'onClick': {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_85));
let __VLS_88;
let __VLS_89;
let __VLS_90;
const __VLS_91 = {
    onClick: (__VLS_ctx.handleReset)
};
__VLS_87.slots.default;
(__VLS_ctx.t('common.reset'));
var __VLS_87;
var __VLS_75;
var __VLS_7;
var __VLS_3;
const __VLS_92 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_93 = __VLS_asFunctionalComponent(__VLS_92, new __VLS_92({
    shadow: "never",
}));
const __VLS_94 = __VLS_93({
    shadow: "never",
}, ...__VLS_functionalComponentArgsRest(__VLS_93));
__VLS_95.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_95.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "toolbar" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "toolbar-left" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "card-title" },
    });
    (__VLS_ctx.t('system.menus'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "toolbar-right" },
    });
    const __VLS_96 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_97 = __VLS_asFunctionalComponent(__VLS_96, new __VLS_96({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_98 = __VLS_97({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_97));
    let __VLS_100;
    let __VLS_101;
    let __VLS_102;
    const __VLS_103 = {
        onClick: (...[$event]) => {
            __VLS_ctx.handleAdd(0);
        }
    };
    __VLS_99.slots.default;
    (__VLS_ctx.t('system.addMenu'));
    var __VLS_99;
    const __VLS_104 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_105 = __VLS_asFunctionalComponent(__VLS_104, new __VLS_104({
        ...{ 'onClick': {} },
    }));
    const __VLS_106 = __VLS_105({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_105));
    let __VLS_108;
    let __VLS_109;
    let __VLS_110;
    const __VLS_111 = {
        onClick: (__VLS_ctx.getList)
    };
    __VLS_107.slots.default;
    (__VLS_ctx.t('common.refresh'));
    var __VLS_107;
}
const __VLS_112 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_113 = __VLS_asFunctionalComponent(__VLS_112, new __VLS_112({
    data: (__VLS_ctx.tableData),
    rowKey: "id",
    border: true,
    treeProps: ({ children: 'children' }),
}));
const __VLS_114 = __VLS_113({
    data: (__VLS_ctx.tableData),
    rowKey: "id",
    border: true,
    treeProps: ({ children: 'children' }),
}, ...__VLS_functionalComponentArgsRest(__VLS_113));
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.loading) }, null, null);
__VLS_115.slots.default;
const __VLS_116 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_117 = __VLS_asFunctionalComponent(__VLS_116, new __VLS_116({
    label: (__VLS_ctx.t('system.name')),
    prop: "name",
    minWidth: "180",
}));
const __VLS_118 = __VLS_117({
    label: (__VLS_ctx.t('system.name')),
    prop: "name",
    minWidth: "180",
}, ...__VLS_functionalComponentArgsRest(__VLS_117));
__VLS_119.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_119.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (__VLS_ctx.getMenuTitle(row));
}
var __VLS_119;
const __VLS_120 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_121 = __VLS_asFunctionalComponent(__VLS_120, new __VLS_120({
    label: (__VLS_ctx.t('system.menuType')),
    width: "100",
    align: "center",
}));
const __VLS_122 = __VLS_121({
    label: (__VLS_ctx.t('system.menuType')),
    width: "100",
    align: "center",
}, ...__VLS_functionalComponentArgsRest(__VLS_121));
__VLS_123.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_123.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    if (row.menuType === 1) {
        const __VLS_124 = {}.ElTag;
        /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
        // @ts-ignore
        const __VLS_125 = __VLS_asFunctionalComponent(__VLS_124, new __VLS_124({
            type: "warning",
        }));
        const __VLS_126 = __VLS_125({
            type: "warning",
        }, ...__VLS_functionalComponentArgsRest(__VLS_125));
        __VLS_127.slots.default;
        (__VLS_ctx.t('system.directory'));
        var __VLS_127;
    }
    else if (row.menuType === 2) {
        const __VLS_128 = {}.ElTag;
        /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
        // @ts-ignore
        const __VLS_129 = __VLS_asFunctionalComponent(__VLS_128, new __VLS_128({
            type: "success",
        }));
        const __VLS_130 = __VLS_129({
            type: "success",
        }, ...__VLS_functionalComponentArgsRest(__VLS_129));
        __VLS_131.slots.default;
        (__VLS_ctx.t('system.menu'));
        var __VLS_131;
    }
    else {
        const __VLS_132 = {}.ElTag;
        /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
        // @ts-ignore
        const __VLS_133 = __VLS_asFunctionalComponent(__VLS_132, new __VLS_132({
            type: "info",
        }));
        const __VLS_134 = __VLS_133({
            type: "info",
        }, ...__VLS_functionalComponentArgsRest(__VLS_133));
        __VLS_135.slots.default;
        (__VLS_ctx.t('system.button'));
        var __VLS_135;
    }
}
var __VLS_123;
const __VLS_136 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_137 = __VLS_asFunctionalComponent(__VLS_136, new __VLS_136({
    label: (__VLS_ctx.t('system.path')),
    prop: "path",
    minWidth: "150",
    showOverflowTooltip: true,
}));
const __VLS_138 = __VLS_137({
    label: (__VLS_ctx.t('system.path')),
    prop: "path",
    minWidth: "150",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_137));
const __VLS_140 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_141 = __VLS_asFunctionalComponent(__VLS_140, new __VLS_140({
    label: (__VLS_ctx.t('system.component')),
    prop: "component",
    minWidth: "180",
    showOverflowTooltip: true,
}));
const __VLS_142 = __VLS_141({
    label: (__VLS_ctx.t('system.component')),
    prop: "component",
    minWidth: "180",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_141));
const __VLS_144 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_145 = __VLS_asFunctionalComponent(__VLS_144, new __VLS_144({
    label: (__VLS_ctx.t('system.icon')),
    width: "160",
}));
const __VLS_146 = __VLS_145({
    label: (__VLS_ctx.t('system.icon')),
    width: "160",
}, ...__VLS_functionalComponentArgsRest(__VLS_145));
__VLS_147.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_147.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    if (row.icon) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "menu-icon-cell" },
        });
        if (__VLS_ctx.resolveIconComponent(row.icon)) {
            const __VLS_148 = {}.ElIcon;
            /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
            // @ts-ignore
            const __VLS_149 = __VLS_asFunctionalComponent(__VLS_148, new __VLS_148({
                ...{ class: "menu-icon-preview" },
            }));
            const __VLS_150 = __VLS_149({
                ...{ class: "menu-icon-preview" },
            }, ...__VLS_functionalComponentArgsRest(__VLS_149));
            __VLS_151.slots.default;
            const __VLS_152 = ((__VLS_ctx.resolveIconComponent(row.icon)));
            // @ts-ignore
            const __VLS_153 = __VLS_asFunctionalComponent(__VLS_152, new __VLS_152({}));
            const __VLS_154 = __VLS_153({}, ...__VLS_functionalComponentArgsRest(__VLS_153));
            var __VLS_151;
        }
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "menu-icon-text" },
        });
        (row.icon);
    }
    else {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "text-muted" },
        });
    }
}
var __VLS_147;
const __VLS_156 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_157 = __VLS_asFunctionalComponent(__VLS_156, new __VLS_156({
    label: (__VLS_ctx.t('system.perms')),
    prop: "perms",
    minWidth: "180",
    showOverflowTooltip: true,
}));
const __VLS_158 = __VLS_157({
    label: (__VLS_ctx.t('system.perms')),
    prop: "perms",
    minWidth: "180",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_157));
const __VLS_160 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_161 = __VLS_asFunctionalComponent(__VLS_160, new __VLS_160({
    label: (__VLS_ctx.t('system.sort')),
    prop: "sort",
    width: "80",
    align: "center",
}));
const __VLS_162 = __VLS_161({
    label: (__VLS_ctx.t('system.sort')),
    prop: "sort",
    width: "80",
    align: "center",
}, ...__VLS_functionalComponentArgsRest(__VLS_161));
const __VLS_164 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_165 = __VLS_asFunctionalComponent(__VLS_164, new __VLS_164({
    label: (__VLS_ctx.t('common.visible')),
    width: "90",
    align: "center",
}));
const __VLS_166 = __VLS_165({
    label: (__VLS_ctx.t('common.visible')),
    width: "90",
    align: "center",
}, ...__VLS_functionalComponentArgsRest(__VLS_165));
__VLS_167.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_167.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_168 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_169 = __VLS_asFunctionalComponent(__VLS_168, new __VLS_168({
        type: (row.visible === 1 ? 'success' : 'info'),
    }));
    const __VLS_170 = __VLS_169({
        type: (row.visible === 1 ? 'success' : 'info'),
    }, ...__VLS_functionalComponentArgsRest(__VLS_169));
    __VLS_171.slots.default;
    (row.visible === 1 ? __VLS_ctx.t('common.visible') : __VLS_ctx.t('common.hidden'));
    var __VLS_171;
}
var __VLS_167;
const __VLS_172 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_173 = __VLS_asFunctionalComponent(__VLS_172, new __VLS_172({
    label: (__VLS_ctx.t('common.status')),
    width: "90",
    align: "center",
}));
const __VLS_174 = __VLS_173({
    label: (__VLS_ctx.t('common.status')),
    width: "90",
    align: "center",
}, ...__VLS_functionalComponentArgsRest(__VLS_173));
__VLS_175.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_175.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_176 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_177 = __VLS_asFunctionalComponent(__VLS_176, new __VLS_176({
        type: (row.status === 1 ? 'success' : 'danger'),
    }));
    const __VLS_178 = __VLS_177({
        type: (row.status === 1 ? 'success' : 'danger'),
    }, ...__VLS_functionalComponentArgsRest(__VLS_177));
    __VLS_179.slots.default;
    (row.status === 1 ? __VLS_ctx.t('common.enabled') : __VLS_ctx.t('common.disabled'));
    var __VLS_179;
}
var __VLS_175;
const __VLS_180 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_181 = __VLS_asFunctionalComponent(__VLS_180, new __VLS_180({
    label: (__VLS_ctx.t('common.actions')),
    width: "180",
    fixed: "right",
    align: "center",
}));
const __VLS_182 = __VLS_181({
    label: (__VLS_ctx.t('common.actions')),
    width: "180",
    fixed: "right",
    align: "center",
}, ...__VLS_functionalComponentArgsRest(__VLS_181));
__VLS_183.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_183.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    if (row.menuType !== 3) {
        const __VLS_184 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_185 = __VLS_asFunctionalComponent(__VLS_184, new __VLS_184({
            ...{ 'onClick': {} },
            link: true,
            type: "primary",
        }));
        const __VLS_186 = __VLS_185({
            ...{ 'onClick': {} },
            link: true,
            type: "primary",
        }, ...__VLS_functionalComponentArgsRest(__VLS_185));
        let __VLS_188;
        let __VLS_189;
        let __VLS_190;
        const __VLS_191 = {
            onClick: (...[$event]) => {
                if (!(row.menuType !== 3))
                    return;
                __VLS_ctx.handleAdd(row.id);
            }
        };
        __VLS_187.slots.default;
        (__VLS_ctx.t('system.addChild'));
        var __VLS_187;
    }
    const __VLS_192 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_193 = __VLS_asFunctionalComponent(__VLS_192, new __VLS_192({
        ...{ 'onClick': {} },
        link: true,
        type: "primary",
    }));
    const __VLS_194 = __VLS_193({
        ...{ 'onClick': {} },
        link: true,
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_193));
    let __VLS_196;
    let __VLS_197;
    let __VLS_198;
    const __VLS_199 = {
        onClick: (...[$event]) => {
            __VLS_ctx.handleEdit(row);
        }
    };
    __VLS_195.slots.default;
    (__VLS_ctx.t('common.edit'));
    var __VLS_195;
    const __VLS_200 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_201 = __VLS_asFunctionalComponent(__VLS_200, new __VLS_200({
        ...{ 'onClick': {} },
        link: true,
        type: "danger",
    }));
    const __VLS_202 = __VLS_201({
        ...{ 'onClick': {} },
        link: true,
        type: "danger",
    }, ...__VLS_functionalComponentArgsRest(__VLS_201));
    let __VLS_204;
    let __VLS_205;
    let __VLS_206;
    const __VLS_207 = {
        onClick: (...[$event]) => {
            __VLS_ctx.handleDelete(row);
        }
    };
    __VLS_203.slots.default;
    (__VLS_ctx.t('common.delete'));
    var __VLS_203;
}
var __VLS_183;
var __VLS_115;
var __VLS_95;
const __VLS_208 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_209 = __VLS_asFunctionalComponent(__VLS_208, new __VLS_208({
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.dialogType === 'add' ? __VLS_ctx.t('system.addMenu') : __VLS_ctx.t('system.editMenu')),
    width: "760px",
    destroyOnClose: true,
}));
const __VLS_210 = __VLS_209({
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.dialogType === 'add' ? __VLS_ctx.t('system.addMenu') : __VLS_ctx.t('system.editMenu')),
    width: "760px",
    destroyOnClose: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_209));
__VLS_211.slots.default;
const __VLS_212 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_213 = __VLS_asFunctionalComponent(__VLS_212, new __VLS_212({
    ref: "formRef",
    model: (__VLS_ctx.formData),
    rules: (__VLS_ctx.rules),
    labelWidth: "100px",
}));
const __VLS_214 = __VLS_213({
    ref: "formRef",
    model: (__VLS_ctx.formData),
    rules: (__VLS_ctx.rules),
    labelWidth: "100px",
}, ...__VLS_functionalComponentArgsRest(__VLS_213));
/** @type {typeof __VLS_ctx.formRef} */ ;
var __VLS_216 = {};
__VLS_215.slots.default;
const __VLS_218 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_219 = __VLS_asFunctionalComponent(__VLS_218, new __VLS_218({
    label: (__VLS_ctx.t('system.parentMenu')),
    prop: "parentId",
}));
const __VLS_220 = __VLS_219({
    label: (__VLS_ctx.t('system.parentMenu')),
    prop: "parentId",
}, ...__VLS_functionalComponentArgsRest(__VLS_219));
__VLS_221.slots.default;
const __VLS_222 = {}.ElTreeSelect;
/** @type {[typeof __VLS_components.ElTreeSelect, typeof __VLS_components.elTreeSelect, ]} */ ;
// @ts-ignore
const __VLS_223 = __VLS_asFunctionalComponent(__VLS_222, new __VLS_222({
    modelValue: (__VLS_ctx.formData.parentId),
    data: (__VLS_ctx.parentTreeOptions),
    nodeKey: "id",
    checkStrictly: true,
    renderAfterExpand: (false),
    props: ({ label: 'name', children: 'children', value: 'id' }),
    placeholder: (__VLS_ctx.t('system.pleaseSelectParentMenu')),
    ...{ style: {} },
}));
const __VLS_224 = __VLS_223({
    modelValue: (__VLS_ctx.formData.parentId),
    data: (__VLS_ctx.parentTreeOptions),
    nodeKey: "id",
    checkStrictly: true,
    renderAfterExpand: (false),
    props: ({ label: 'name', children: 'children', value: 'id' }),
    placeholder: (__VLS_ctx.t('system.pleaseSelectParentMenu')),
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_223));
var __VLS_221;
const __VLS_226 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_227 = __VLS_asFunctionalComponent(__VLS_226, new __VLS_226({
    label: (__VLS_ctx.t('system.menuName')),
    prop: "name",
}));
const __VLS_228 = __VLS_227({
    label: (__VLS_ctx.t('system.menuName')),
    prop: "name",
}, ...__VLS_functionalComponentArgsRest(__VLS_227));
__VLS_229.slots.default;
const __VLS_230 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_231 = __VLS_asFunctionalComponent(__VLS_230, new __VLS_230({
    modelValue: (__VLS_ctx.formData.name),
    placeholder: (__VLS_ctx.t('system.pleaseInputMenuName')),
}));
const __VLS_232 = __VLS_231({
    modelValue: (__VLS_ctx.formData.name),
    placeholder: (__VLS_ctx.t('system.pleaseInputMenuName')),
}, ...__VLS_functionalComponentArgsRest(__VLS_231));
var __VLS_229;
const __VLS_234 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_235 = __VLS_asFunctionalComponent(__VLS_234, new __VLS_234({
    label: (__VLS_ctx.t('system.menuType')),
    prop: "menuType",
}));
const __VLS_236 = __VLS_235({
    label: (__VLS_ctx.t('system.menuType')),
    prop: "menuType",
}, ...__VLS_functionalComponentArgsRest(__VLS_235));
__VLS_237.slots.default;
const __VLS_238 = {}.ElRadioGroup;
/** @type {[typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, ]} */ ;
// @ts-ignore
const __VLS_239 = __VLS_asFunctionalComponent(__VLS_238, new __VLS_238({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.formData.menuType),
}));
const __VLS_240 = __VLS_239({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.formData.menuType),
}, ...__VLS_functionalComponentArgsRest(__VLS_239));
let __VLS_242;
let __VLS_243;
let __VLS_244;
const __VLS_245 = {
    onChange: (__VLS_ctx.handleMenuTypeChange)
};
__VLS_241.slots.default;
const __VLS_246 = {}.ElRadio;
/** @type {[typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, ]} */ ;
// @ts-ignore
const __VLS_247 = __VLS_asFunctionalComponent(__VLS_246, new __VLS_246({
    label: (1),
}));
const __VLS_248 = __VLS_247({
    label: (1),
}, ...__VLS_functionalComponentArgsRest(__VLS_247));
__VLS_249.slots.default;
(__VLS_ctx.t('system.directory'));
var __VLS_249;
const __VLS_250 = {}.ElRadio;
/** @type {[typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, ]} */ ;
// @ts-ignore
const __VLS_251 = __VLS_asFunctionalComponent(__VLS_250, new __VLS_250({
    label: (2),
}));
const __VLS_252 = __VLS_251({
    label: (2),
}, ...__VLS_functionalComponentArgsRest(__VLS_251));
__VLS_253.slots.default;
(__VLS_ctx.t('system.menu'));
var __VLS_253;
const __VLS_254 = {}.ElRadio;
/** @type {[typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, ]} */ ;
// @ts-ignore
const __VLS_255 = __VLS_asFunctionalComponent(__VLS_254, new __VLS_254({
    label: (3),
}));
const __VLS_256 = __VLS_255({
    label: (3),
}, ...__VLS_functionalComponentArgsRest(__VLS_255));
__VLS_257.slots.default;
(__VLS_ctx.t('system.button'));
var __VLS_257;
var __VLS_241;
var __VLS_237;
if (__VLS_ctx.formData.menuType !== 3) {
    const __VLS_258 = {}.ElRow;
    /** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ ;
    // @ts-ignore
    const __VLS_259 = __VLS_asFunctionalComponent(__VLS_258, new __VLS_258({
        gutter: (16),
    }));
    const __VLS_260 = __VLS_259({
        gutter: (16),
    }, ...__VLS_functionalComponentArgsRest(__VLS_259));
    __VLS_261.slots.default;
    const __VLS_262 = {}.ElCol;
    /** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
    // @ts-ignore
    const __VLS_263 = __VLS_asFunctionalComponent(__VLS_262, new __VLS_262({
        span: (12),
    }));
    const __VLS_264 = __VLS_263({
        span: (12),
    }, ...__VLS_functionalComponentArgsRest(__VLS_263));
    __VLS_265.slots.default;
    const __VLS_266 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_267 = __VLS_asFunctionalComponent(__VLS_266, new __VLS_266({
        label: (__VLS_ctx.t('system.path')),
        prop: "path",
    }));
    const __VLS_268 = __VLS_267({
        label: (__VLS_ctx.t('system.path')),
        prop: "path",
    }, ...__VLS_functionalComponentArgsRest(__VLS_267));
    __VLS_269.slots.default;
    const __VLS_270 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_271 = __VLS_asFunctionalComponent(__VLS_270, new __VLS_270({
        modelValue: (__VLS_ctx.formData.path),
        placeholder: (__VLS_ctx.t('system.pleaseInputPath')),
    }));
    const __VLS_272 = __VLS_271({
        modelValue: (__VLS_ctx.formData.path),
        placeholder: (__VLS_ctx.t('system.pleaseInputPath')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_271));
    var __VLS_269;
    var __VLS_265;
    if (__VLS_ctx.formData.menuType === 2) {
        const __VLS_274 = {}.ElCol;
        /** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
        // @ts-ignore
        const __VLS_275 = __VLS_asFunctionalComponent(__VLS_274, new __VLS_274({
            span: (12),
        }));
        const __VLS_276 = __VLS_275({
            span: (12),
        }, ...__VLS_functionalComponentArgsRest(__VLS_275));
        __VLS_277.slots.default;
        const __VLS_278 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_279 = __VLS_asFunctionalComponent(__VLS_278, new __VLS_278({
            label: (__VLS_ctx.t('system.component')),
            prop: "component",
        }));
        const __VLS_280 = __VLS_279({
            label: (__VLS_ctx.t('system.component')),
            prop: "component",
        }, ...__VLS_functionalComponentArgsRest(__VLS_279));
        __VLS_281.slots.default;
        const __VLS_282 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_283 = __VLS_asFunctionalComponent(__VLS_282, new __VLS_282({
            modelValue: (__VLS_ctx.formData.component),
            placeholder: (__VLS_ctx.t('system.pleaseInputComponent')),
        }));
        const __VLS_284 = __VLS_283({
            modelValue: (__VLS_ctx.formData.component),
            placeholder: (__VLS_ctx.t('system.pleaseInputComponent')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_283));
        var __VLS_281;
        var __VLS_277;
    }
    var __VLS_261;
}
if (__VLS_ctx.formData.menuType !== 3) {
    const __VLS_286 = {}.ElRow;
    /** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ ;
    // @ts-ignore
    const __VLS_287 = __VLS_asFunctionalComponent(__VLS_286, new __VLS_286({
        gutter: (16),
    }));
    const __VLS_288 = __VLS_287({
        gutter: (16),
    }, ...__VLS_functionalComponentArgsRest(__VLS_287));
    __VLS_289.slots.default;
    const __VLS_290 = {}.ElCol;
    /** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
    // @ts-ignore
    const __VLS_291 = __VLS_asFunctionalComponent(__VLS_290, new __VLS_290({
        span: (14),
    }));
    const __VLS_292 = __VLS_291({
        span: (14),
    }, ...__VLS_functionalComponentArgsRest(__VLS_291));
    __VLS_293.slots.default;
    const __VLS_294 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_295 = __VLS_asFunctionalComponent(__VLS_294, new __VLS_294({
        label: (__VLS_ctx.t('system.icon')),
        prop: "icon",
    }));
    const __VLS_296 = __VLS_295({
        label: (__VLS_ctx.t('system.icon')),
        prop: "icon",
    }, ...__VLS_functionalComponentArgsRest(__VLS_295));
    __VLS_297.slots.default;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "icon-picker-box" },
    });
    const __VLS_298 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_299 = __VLS_asFunctionalComponent(__VLS_298, new __VLS_298({
        modelValue: (__VLS_ctx.formData.icon),
        placeholder: (__VLS_ctx.t('system.pleaseInputIcon')),
        clearable: true,
    }));
    const __VLS_300 = __VLS_299({
        modelValue: (__VLS_ctx.formData.icon),
        placeholder: (__VLS_ctx.t('system.pleaseInputIcon')),
        clearable: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_299));
    __VLS_301.slots.default;
    {
        const { prepend: __VLS_thisSlot } = __VLS_301.slots;
        if (__VLS_ctx.currentIconComponent) {
            const __VLS_302 = {}.ElIcon;
            /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
            // @ts-ignore
            const __VLS_303 = __VLS_asFunctionalComponent(__VLS_302, new __VLS_302({}));
            const __VLS_304 = __VLS_303({}, ...__VLS_functionalComponentArgsRest(__VLS_303));
            __VLS_305.slots.default;
            const __VLS_306 = ((__VLS_ctx.currentIconComponent));
            // @ts-ignore
            const __VLS_307 = __VLS_asFunctionalComponent(__VLS_306, new __VLS_306({}));
            const __VLS_308 = __VLS_307({}, ...__VLS_functionalComponentArgsRest(__VLS_307));
            var __VLS_305;
        }
        else {
            __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
                ...{ class: "text-muted" },
            });
        }
    }
    var __VLS_301;
    const __VLS_310 = {}.ElPopover;
    /** @type {[typeof __VLS_components.ElPopover, typeof __VLS_components.elPopover, typeof __VLS_components.ElPopover, typeof __VLS_components.elPopover, ]} */ ;
    // @ts-ignore
    const __VLS_311 = __VLS_asFunctionalComponent(__VLS_310, new __VLS_310({
        placement: "bottom-start",
        width: (520),
        trigger: "click",
    }));
    const __VLS_312 = __VLS_311({
        placement: "bottom-start",
        width: (520),
        trigger: "click",
    }, ...__VLS_functionalComponentArgsRest(__VLS_311));
    __VLS_313.slots.default;
    {
        const { reference: __VLS_thisSlot } = __VLS_313.slots;
        const __VLS_314 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_315 = __VLS_asFunctionalComponent(__VLS_314, new __VLS_314({}));
        const __VLS_316 = __VLS_315({}, ...__VLS_functionalComponentArgsRest(__VLS_315));
        __VLS_317.slots.default;
        (__VLS_ctx.t('system.selectIcon'));
        var __VLS_317;
    }
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "icon-panel" },
    });
    for (const [iconName] of __VLS_getVForSourceType((__VLS_ctx.iconNames))) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ onClick: (...[$event]) => {
                    if (!(__VLS_ctx.formData.menuType !== 3))
                        return;
                    __VLS_ctx.selectIcon(iconName);
                } },
            key: (iconName),
            ...{ class: "icon-item" },
        });
        const __VLS_318 = {}.ElIcon;
        /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
        // @ts-ignore
        const __VLS_319 = __VLS_asFunctionalComponent(__VLS_318, new __VLS_318({
            ...{ class: "icon-item-preview" },
        }));
        const __VLS_320 = __VLS_319({
            ...{ class: "icon-item-preview" },
        }, ...__VLS_functionalComponentArgsRest(__VLS_319));
        __VLS_321.slots.default;
        const __VLS_322 = ((__VLS_ctx.resolveIconComponent(iconName)));
        // @ts-ignore
        const __VLS_323 = __VLS_asFunctionalComponent(__VLS_322, new __VLS_322({}));
        const __VLS_324 = __VLS_323({}, ...__VLS_functionalComponentArgsRest(__VLS_323));
        var __VLS_321;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "icon-item-text" },
        });
        (iconName);
    }
    var __VLS_313;
    const __VLS_326 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_327 = __VLS_asFunctionalComponent(__VLS_326, new __VLS_326({
        ...{ 'onClick': {} },
    }));
    const __VLS_328 = __VLS_327({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_327));
    let __VLS_330;
    let __VLS_331;
    let __VLS_332;
    const __VLS_333 = {
        onClick: (__VLS_ctx.clearIcon)
    };
    __VLS_329.slots.default;
    (__VLS_ctx.t('common.clear'));
    var __VLS_329;
    var __VLS_297;
    var __VLS_293;
    const __VLS_334 = {}.ElCol;
    /** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
    // @ts-ignore
    const __VLS_335 = __VLS_asFunctionalComponent(__VLS_334, new __VLS_334({
        span: (10),
    }));
    const __VLS_336 = __VLS_335({
        span: (10),
    }, ...__VLS_functionalComponentArgsRest(__VLS_335));
    __VLS_337.slots.default;
    const __VLS_338 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_339 = __VLS_asFunctionalComponent(__VLS_338, new __VLS_338({
        label: (__VLS_ctx.t('system.iconPreview')),
    }));
    const __VLS_340 = __VLS_339({
        label: (__VLS_ctx.t('system.iconPreview')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_339));
    __VLS_341.slots.default;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "icon-preview-box" },
    });
    if (__VLS_ctx.currentIconComponent) {
        const __VLS_342 = {}.ElIcon;
        /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
        // @ts-ignore
        const __VLS_343 = __VLS_asFunctionalComponent(__VLS_342, new __VLS_342({
            ...{ class: "icon-preview-large" },
        }));
        const __VLS_344 = __VLS_343({
            ...{ class: "icon-preview-large" },
        }, ...__VLS_functionalComponentArgsRest(__VLS_343));
        __VLS_345.slots.default;
        const __VLS_346 = ((__VLS_ctx.currentIconComponent));
        // @ts-ignore
        const __VLS_347 = __VLS_asFunctionalComponent(__VLS_346, new __VLS_346({}));
        const __VLS_348 = __VLS_347({}, ...__VLS_functionalComponentArgsRest(__VLS_347));
        var __VLS_345;
    }
    else {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "text-muted" },
        });
        (__VLS_ctx.t('system.noIcon'));
    }
    var __VLS_341;
    var __VLS_337;
    var __VLS_289;
}
if (__VLS_ctx.formData.menuType !== 3) {
    const __VLS_350 = {}.ElRow;
    /** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ ;
    // @ts-ignore
    const __VLS_351 = __VLS_asFunctionalComponent(__VLS_350, new __VLS_350({
        gutter: (16),
    }));
    const __VLS_352 = __VLS_351({
        gutter: (16),
    }, ...__VLS_functionalComponentArgsRest(__VLS_351));
    __VLS_353.slots.default;
    const __VLS_354 = {}.ElCol;
    /** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
    // @ts-ignore
    const __VLS_355 = __VLS_asFunctionalComponent(__VLS_354, new __VLS_354({
        span: (12),
    }));
    const __VLS_356 = __VLS_355({
        span: (12),
    }, ...__VLS_functionalComponentArgsRest(__VLS_355));
    __VLS_357.slots.default;
    const __VLS_358 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_359 = __VLS_asFunctionalComponent(__VLS_358, new __VLS_358({
        label: (__VLS_ctx.t('system.sort')),
        prop: "sort",
    }));
    const __VLS_360 = __VLS_359({
        label: (__VLS_ctx.t('system.sort')),
        prop: "sort",
    }, ...__VLS_functionalComponentArgsRest(__VLS_359));
    __VLS_361.slots.default;
    const __VLS_362 = {}.ElInputNumber;
    /** @type {[typeof __VLS_components.ElInputNumber, typeof __VLS_components.elInputNumber, ]} */ ;
    // @ts-ignore
    const __VLS_363 = __VLS_asFunctionalComponent(__VLS_362, new __VLS_362({
        modelValue: (__VLS_ctx.formData.sort),
        min: (0),
        max: (9999),
        ...{ style: {} },
    }));
    const __VLS_364 = __VLS_363({
        modelValue: (__VLS_ctx.formData.sort),
        min: (0),
        max: (9999),
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_363));
    var __VLS_361;
    var __VLS_357;
    var __VLS_353;
}
if (__VLS_ctx.formData.menuType === 3) {
    const __VLS_366 = {}.ElRow;
    /** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ ;
    // @ts-ignore
    const __VLS_367 = __VLS_asFunctionalComponent(__VLS_366, new __VLS_366({
        gutter: (16),
    }));
    const __VLS_368 = __VLS_367({
        gutter: (16),
    }, ...__VLS_functionalComponentArgsRest(__VLS_367));
    __VLS_369.slots.default;
    const __VLS_370 = {}.ElCol;
    /** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
    // @ts-ignore
    const __VLS_371 = __VLS_asFunctionalComponent(__VLS_370, new __VLS_370({
        span: (12),
    }));
    const __VLS_372 = __VLS_371({
        span: (12),
    }, ...__VLS_functionalComponentArgsRest(__VLS_371));
    __VLS_373.slots.default;
    const __VLS_374 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_375 = __VLS_asFunctionalComponent(__VLS_374, new __VLS_374({
        label: (__VLS_ctx.t('system.sort')),
        prop: "sort",
    }));
    const __VLS_376 = __VLS_375({
        label: (__VLS_ctx.t('system.sort')),
        prop: "sort",
    }, ...__VLS_functionalComponentArgsRest(__VLS_375));
    __VLS_377.slots.default;
    const __VLS_378 = {}.ElInputNumber;
    /** @type {[typeof __VLS_components.ElInputNumber, typeof __VLS_components.elInputNumber, ]} */ ;
    // @ts-ignore
    const __VLS_379 = __VLS_asFunctionalComponent(__VLS_378, new __VLS_378({
        modelValue: (__VLS_ctx.formData.sort),
        min: (0),
        max: (9999),
        ...{ style: {} },
    }));
    const __VLS_380 = __VLS_379({
        modelValue: (__VLS_ctx.formData.sort),
        min: (0),
        max: (9999),
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_379));
    var __VLS_377;
    var __VLS_373;
    var __VLS_369;
}
if (__VLS_ctx.formData.menuType === 3) {
    const __VLS_382 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_383 = __VLS_asFunctionalComponent(__VLS_382, new __VLS_382({
        label: (__VLS_ctx.t('system.perms')),
        prop: "perms",
    }));
    const __VLS_384 = __VLS_383({
        label: (__VLS_ctx.t('system.perms')),
        prop: "perms",
    }, ...__VLS_functionalComponentArgsRest(__VLS_383));
    __VLS_385.slots.default;
    const __VLS_386 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_387 = __VLS_asFunctionalComponent(__VLS_386, new __VLS_386({
        modelValue: (__VLS_ctx.formData.perms),
        placeholder: (__VLS_ctx.t('system.pleaseInputPerms')),
    }));
    const __VLS_388 = __VLS_387({
        modelValue: (__VLS_ctx.formData.perms),
        placeholder: (__VLS_ctx.t('system.pleaseInputPerms')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_387));
    var __VLS_385;
}
const __VLS_390 = {}.ElRow;
/** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ ;
// @ts-ignore
const __VLS_391 = __VLS_asFunctionalComponent(__VLS_390, new __VLS_390({
    gutter: (16),
}));
const __VLS_392 = __VLS_391({
    gutter: (16),
}, ...__VLS_functionalComponentArgsRest(__VLS_391));
__VLS_393.slots.default;
const __VLS_394 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_395 = __VLS_asFunctionalComponent(__VLS_394, new __VLS_394({
    span: (12),
}));
const __VLS_396 = __VLS_395({
    span: (12),
}, ...__VLS_functionalComponentArgsRest(__VLS_395));
__VLS_397.slots.default;
const __VLS_398 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_399 = __VLS_asFunctionalComponent(__VLS_398, new __VLS_398({
    label: (__VLS_ctx.t('common.visible')),
    prop: "visible",
}));
const __VLS_400 = __VLS_399({
    label: (__VLS_ctx.t('common.visible')),
    prop: "visible",
}, ...__VLS_functionalComponentArgsRest(__VLS_399));
__VLS_401.slots.default;
const __VLS_402 = {}.ElRadioGroup;
/** @type {[typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, ]} */ ;
// @ts-ignore
const __VLS_403 = __VLS_asFunctionalComponent(__VLS_402, new __VLS_402({
    modelValue: (__VLS_ctx.formData.visible),
}));
const __VLS_404 = __VLS_403({
    modelValue: (__VLS_ctx.formData.visible),
}, ...__VLS_functionalComponentArgsRest(__VLS_403));
__VLS_405.slots.default;
const __VLS_406 = {}.ElRadio;
/** @type {[typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, ]} */ ;
// @ts-ignore
const __VLS_407 = __VLS_asFunctionalComponent(__VLS_406, new __VLS_406({
    label: (1),
}));
const __VLS_408 = __VLS_407({
    label: (1),
}, ...__VLS_functionalComponentArgsRest(__VLS_407));
__VLS_409.slots.default;
(__VLS_ctx.t('common.visible'));
var __VLS_409;
const __VLS_410 = {}.ElRadio;
/** @type {[typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, ]} */ ;
// @ts-ignore
const __VLS_411 = __VLS_asFunctionalComponent(__VLS_410, new __VLS_410({
    label: (0),
}));
const __VLS_412 = __VLS_411({
    label: (0),
}, ...__VLS_functionalComponentArgsRest(__VLS_411));
__VLS_413.slots.default;
(__VLS_ctx.t('common.hidden'));
var __VLS_413;
var __VLS_405;
var __VLS_401;
var __VLS_397;
const __VLS_414 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_415 = __VLS_asFunctionalComponent(__VLS_414, new __VLS_414({
    span: (12),
}));
const __VLS_416 = __VLS_415({
    span: (12),
}, ...__VLS_functionalComponentArgsRest(__VLS_415));
__VLS_417.slots.default;
const __VLS_418 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_419 = __VLS_asFunctionalComponent(__VLS_418, new __VLS_418({
    label: (__VLS_ctx.t('common.status')),
    prop: "status",
}));
const __VLS_420 = __VLS_419({
    label: (__VLS_ctx.t('common.status')),
    prop: "status",
}, ...__VLS_functionalComponentArgsRest(__VLS_419));
__VLS_421.slots.default;
const __VLS_422 = {}.ElRadioGroup;
/** @type {[typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, ]} */ ;
// @ts-ignore
const __VLS_423 = __VLS_asFunctionalComponent(__VLS_422, new __VLS_422({
    modelValue: (__VLS_ctx.formData.status),
}));
const __VLS_424 = __VLS_423({
    modelValue: (__VLS_ctx.formData.status),
}, ...__VLS_functionalComponentArgsRest(__VLS_423));
__VLS_425.slots.default;
const __VLS_426 = {}.ElRadio;
/** @type {[typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, ]} */ ;
// @ts-ignore
const __VLS_427 = __VLS_asFunctionalComponent(__VLS_426, new __VLS_426({
    label: (1),
}));
const __VLS_428 = __VLS_427({
    label: (1),
}, ...__VLS_functionalComponentArgsRest(__VLS_427));
__VLS_429.slots.default;
(__VLS_ctx.t('common.enabled'));
var __VLS_429;
const __VLS_430 = {}.ElRadio;
/** @type {[typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, ]} */ ;
// @ts-ignore
const __VLS_431 = __VLS_asFunctionalComponent(__VLS_430, new __VLS_430({
    label: (0),
}));
const __VLS_432 = __VLS_431({
    label: (0),
}, ...__VLS_functionalComponentArgsRest(__VLS_431));
__VLS_433.slots.default;
(__VLS_ctx.t('common.disabled'));
var __VLS_433;
var __VLS_425;
var __VLS_421;
var __VLS_417;
var __VLS_393;
var __VLS_215;
{
    const { footer: __VLS_thisSlot } = __VLS_211.slots;
    const __VLS_434 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_435 = __VLS_asFunctionalComponent(__VLS_434, new __VLS_434({
        ...{ 'onClick': {} },
    }));
    const __VLS_436 = __VLS_435({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_435));
    let __VLS_438;
    let __VLS_439;
    let __VLS_440;
    const __VLS_441 = {
        onClick: (...[$event]) => {
            __VLS_ctx.dialogVisible = false;
        }
    };
    __VLS_437.slots.default;
    (__VLS_ctx.t('common.cancel'));
    var __VLS_437;
    const __VLS_442 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_443 = __VLS_asFunctionalComponent(__VLS_442, new __VLS_442({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.submitLoading),
    }));
    const __VLS_444 = __VLS_443({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.submitLoading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_443));
    let __VLS_446;
    let __VLS_447;
    let __VLS_448;
    const __VLS_449 = {
        onClick: (__VLS_ctx.handleSubmit)
    };
    __VLS_445.slots.default;
    (__VLS_ctx.t('common.confirm'));
    var __VLS_445;
}
var __VLS_211;
/** @type {__VLS_StyleScopedClasses['app-container']} */ ;
/** @type {__VLS_StyleScopedClasses['mb-16']} */ ;
/** @type {__VLS_StyleScopedClasses['toolbar']} */ ;
/** @type {__VLS_StyleScopedClasses['toolbar-left']} */ ;
/** @type {__VLS_StyleScopedClasses['card-title']} */ ;
/** @type {__VLS_StyleScopedClasses['toolbar-right']} */ ;
/** @type {__VLS_StyleScopedClasses['menu-icon-cell']} */ ;
/** @type {__VLS_StyleScopedClasses['menu-icon-preview']} */ ;
/** @type {__VLS_StyleScopedClasses['menu-icon-text']} */ ;
/** @type {__VLS_StyleScopedClasses['text-muted']} */ ;
/** @type {__VLS_StyleScopedClasses['icon-picker-box']} */ ;
/** @type {__VLS_StyleScopedClasses['text-muted']} */ ;
/** @type {__VLS_StyleScopedClasses['icon-panel']} */ ;
/** @type {__VLS_StyleScopedClasses['icon-item']} */ ;
/** @type {__VLS_StyleScopedClasses['icon-item-preview']} */ ;
/** @type {__VLS_StyleScopedClasses['icon-item-text']} */ ;
/** @type {__VLS_StyleScopedClasses['icon-preview-box']} */ ;
/** @type {__VLS_StyleScopedClasses['icon-preview-large']} */ ;
/** @type {__VLS_StyleScopedClasses['text-muted']} */ ;
// @ts-ignore
var __VLS_217 = __VLS_216;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            t: t,
            iconNames: iconNames,
            loading: loading,
            submitLoading: submitLoading,
            queryForm: queryForm,
            dialogVisible: dialogVisible,
            dialogType: dialogType,
            formRef: formRef,
            tableData: tableData,
            formData: formData,
            currentIconComponent: currentIconComponent,
            parentTreeOptions: parentTreeOptions,
            rules: rules,
            resolveIconComponent: resolveIconComponent,
            selectIcon: selectIcon,
            clearIcon: clearIcon,
            getMenuTitle: getMenuTitle,
            handleMenuTypeChange: handleMenuTypeChange,
            getList: getList,
            handleSearch: handleSearch,
            handleReset: handleReset,
            handleAdd: handleAdd,
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
