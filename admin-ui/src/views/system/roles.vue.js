import { computed, ref, onMounted, nextTick } from 'vue';
import { ElMessage } from 'element-plus';
import { useI18n } from 'vue-i18n';
import { usePagination, useLoading, useConfirm, useForm } from '@/composables';
import { roleService, menuService } from '@/services';
// ===== i18n =====
const { t } = useI18n();
// ===== helpers =====
function isSuperRole(r) {
    if (!r)
        return false;
    return r.isSuper === true || r.code === 'super_admin' || r.id === 1;
}
// ===== state =====
const { pagination, updateTotal } = usePagination(20);
const { loading, withLoading } = useLoading();
const { confirm } = useConfirm();
const { form: queryForm } = useForm({
    initialData: { keyword: '', status: 0 },
});
const tableData = ref([]);
// ===== list =====
async function fetchList() {
    await withLoading(async () => {
        try {
            const q = {
                keyword: queryForm.keyword,
                status: queryForm.status,
                page: pagination.page,
                size: pagination.pageSize,
            };
            if (q.status === 0)
                delete q.status;
            const resp = await roleService.getList(q);
            tableData.value = resp.data || [];
            updateTotal(resp.total || 0);
        }
        catch (e) {
            ElMessage.error(e?.message || t('common.failed'));
        }
    });
}
function unwrapList(resp) {
    // 兼容 data / list / rows / result 之类
    if (!resp)
        return [];
    if (Array.isArray(resp))
        return resp;
    return resp.data || resp.list || resp.rows || resp.result || [];
}
function unwrapData(resp) {
    if (!resp)
        return null;
    // detail 这种通常在 data 里
    return resp.data ?? resp;
}
// 把扁平菜单转成树（并过滤掉 menuType=3 的按钮项，树里只放目录/菜单）
function buildMenuTree(flat) {
    const list = (flat || []).filter((x) => x && x.menuType !== 3);
    const map = new Map();
    list.forEach((n) => map.set(n.id, { ...n, children: [] }));
    const roots = [];
    map.forEach((node) => {
        const pid = node.parentId;
        if (pid && map.has(pid)) {
            map.get(pid).children.push(node);
        }
        else {
            roots.push(node);
        }
    });
    // 可选：按 sort 排序
    const sortRec = (arr) => {
        arr.sort((a, b) => (a.sort ?? 0) - (b.sort ?? 0));
        arr.forEach((x) => x.children?.length && sortRec(x.children));
    };
    sortRec(roots);
    return roots;
}
function onSearch() {
    pagination.page = 1;
    fetchList();
}
function onReset() {
    queryForm.keyword = '';
    queryForm.status = 0;
    onSearch();
}
// ===== create/update dialog =====
const editVisible = ref(false);
const editFormRef = ref();
const { form: editForm } = useForm({
    initialData: {
        id: 0,
        name: '',
        code: '',
        remark: '',
        status: 1,
    },
});
const editIsUpdate = computed(() => editForm.id > 0);
const { loading: editLoading, withLoading: withEditLoading } = useLoading();
function openCreate() {
    editForm.id = 0;
    editForm.name = '';
    editForm.code = '';
    editForm.remark = '';
    editForm.status = 1;
    editVisible.value = true;
}
function openUpdate(row) {
    if (isSuperRole(row))
        return;
    editForm.id = row.id;
    editForm.name = row.name;
    editForm.code = row.code;
    editForm.remark = row.remark || '';
    editForm.status = row.status === 2 ? 2 : 1;
    editVisible.value = true;
}
async function submitEdit() {
    await editFormRef.value?.validate?.();
    await withEditLoading(async () => {
        try {
            const payload = { ...editForm };
            const resp = editIsUpdate.value
                ? await roleService.update(editForm.id, payload)
                : await roleService.create(payload);
            if (resp.code === 200) {
                ElMessage.success(resp.msg || t('common.success'));
                editVisible.value = false;
                fetchList();
            }
            else {
                ElMessage.error(resp.msg || t('common.failed'));
            }
        }
        catch (e) {
            ElMessage.error(e?.message || t('common.failed'));
        }
    });
}
async function onDelete(row) {
    if (isSuperRole(row))
        return;
    try {
        await confirm(t('common.confirmDelete'), { type: 'warning' });
        const resp = await roleService.delete(row.id);
        if (resp.code === 200) {
            ElMessage.success(resp.msg || t('common.success'));
            fetchList();
        }
        else {
            ElMessage.error(resp.msg || t('common.failed'));
        }
    }
    catch (e) {
        if (e === 'cancel')
            return;
        ElMessage.error(e?.message || t('common.failed'));
    }
}
// ===== grant dialog =====
const grantVisible = ref(false);
const currentRole = ref(null);
const { loading: grantLoading, withLoading: withGrantLoading } = useLoading();
const menuTree = ref([]);
const permList = ref([]);
const menuTreeRef = ref();
const checkedPermKeys = ref([]);
const grantReadonly = computed(() => isSuperRole(currentRole.value));
function openGrant(row) {
    currentRole.value = row;
    grantVisible.value = true;
    initGrant(row.id);
}
async function initGrant(roleId) {
    await withGrantLoading(async () => {
        try {
            // ✅ 每次打开先清理（避免切角色残留）
            menuTree.value = [];
            permList.value = [];
            checkedPermKeys.value = [];
            await nextTick();
            menuTreeRef.value?.setCheckedKeys?.([]);
            const [menusResp, permsResp, detailResp] = await Promise.all([
                menuService.getMenuTree(),
                menuService.getPermissionList(),
                roleService.getRoleGrantDetail(roleId),
            ]);
            const menusFlat = unwrapList(menusResp);
            const perms = unwrapList(permsResp);
            const detail = unwrapData(detailResp) || {};
            menuTree.value = buildMenuTree(menusFlat);
            permList.value = perms;
            await nextTick();
            menuTreeRef.value?.setCheckedKeys((detail.menuIds || []));
            checkedPermKeys.value = (detail.permKeys || []);
        }
        catch (e) {
            ElMessage.error(e?.message || t('common.failed'));
        }
    });
}
function collectCheckedMenuIds() {
    const tree = menuTreeRef.value;
    if (!tree)
        return [];
    // Element Plus tree：getCheckedKeys + getHalfCheckedKeys
    const full = (tree.getCheckedKeys?.() || []);
    const half = (tree.getHalfCheckedKeys?.() || []);
    return Array.from(new Set([...full, ...half]));
}
async function submitGrant() {
    if (!currentRole.value)
        return;
    if (grantReadonly.value) {
        ElMessage.warning(t('system.superAdminNoGrant'));
        return;
    }
    try {
        const payload = {
            roleId: currentRole.value.id,
            menuIds: collectCheckedMenuIds(),
            permKeys: checkedPermKeys.value,
        };
        const resp = await roleService.grantRole(payload);
        if (resp.code === 200) {
            ElMessage.success(resp.msg || t('common.success'));
            grantVisible.value = false;
            // ✅ 可选：保存后刷新列表（比如你要展示更新时间/状态）
            fetchList();
        }
        else {
            ElMessage.error(resp.msg || t('common.failed'));
        }
    }
    catch (e) {
        ElMessage.error(e?.message || t('common.failed'));
    }
}
function onGrantClosed() {
    currentRole.value = null;
    menuTree.value = [];
    permList.value = [];
    checkedPermKeys.value = [];
    menuTreeRef.value?.setCheckedKeys?.([]);
}
// ===== init =====
onMounted(fetchList);
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
const __VLS_0 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({}));
const __VLS_2 = __VLS_1({}, ...__VLS_functionalComponentArgsRest(__VLS_1));
__VLS_3.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_3.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
    (__VLS_ctx.t('system.roles'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    const __VLS_4 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_5 = __VLS_asFunctionalComponent(__VLS_4, new __VLS_4({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_6 = __VLS_5({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_5));
    let __VLS_8;
    let __VLS_9;
    let __VLS_10;
    const __VLS_11 = {
        onClick: (__VLS_ctx.openCreate)
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPerm)(null, { ...__VLS_directiveBindingRestFields, value: ('sys:role:add') }, null, null);
    __VLS_7.slots.default;
    (__VLS_ctx.t('perms.sys:role:add'));
    var __VLS_7;
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ style: {} },
});
const __VLS_12 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
    modelValue: (__VLS_ctx.queryForm.keyword),
    placeholder: (__VLS_ctx.t('common.keyword')),
    clearable: true,
    ...{ style: {} },
}));
const __VLS_14 = __VLS_13({
    modelValue: (__VLS_ctx.queryForm.keyword),
    placeholder: (__VLS_ctx.t('common.keyword')),
    clearable: true,
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_13));
const __VLS_16 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.queryForm.status),
    ...{ style: {} },
    placeholder: (__VLS_ctx.t('common.status')),
}));
const __VLS_18 = __VLS_17({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.queryForm.status),
    ...{ style: {} },
    placeholder: (__VLS_ctx.t('common.status')),
}, ...__VLS_functionalComponentArgsRest(__VLS_17));
let __VLS_20;
let __VLS_21;
let __VLS_22;
const __VLS_23 = {
    onChange: (__VLS_ctx.onSearch)
};
__VLS_19.slots.default;
const __VLS_24 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
    label: (__VLS_ctx.t('common.all')),
    value: (0),
}));
const __VLS_26 = __VLS_25({
    label: (__VLS_ctx.t('common.all')),
    value: (0),
}, ...__VLS_functionalComponentArgsRest(__VLS_25));
const __VLS_28 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({
    label: (__VLS_ctx.t('common.enabled')),
    value: (1),
}));
const __VLS_30 = __VLS_29({
    label: (__VLS_ctx.t('common.enabled')),
    value: (1),
}, ...__VLS_functionalComponentArgsRest(__VLS_29));
const __VLS_32 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
    label: (__VLS_ctx.t('common.disabled')),
    value: (2),
}));
const __VLS_34 = __VLS_33({
    label: (__VLS_ctx.t('common.disabled')),
    value: (2),
}, ...__VLS_functionalComponentArgsRest(__VLS_33));
var __VLS_19;
const __VLS_36 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({
    ...{ 'onClick': {} },
}));
const __VLS_38 = __VLS_37({
    ...{ 'onClick': {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_37));
let __VLS_40;
let __VLS_41;
let __VLS_42;
const __VLS_43 = {
    onClick: (__VLS_ctx.onSearch)
};
__VLS_39.slots.default;
(__VLS_ctx.t('common.search'));
var __VLS_39;
const __VLS_44 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_45 = __VLS_asFunctionalComponent(__VLS_44, new __VLS_44({
    ...{ 'onClick': {} },
}));
const __VLS_46 = __VLS_45({
    ...{ 'onClick': {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_45));
let __VLS_48;
let __VLS_49;
let __VLS_50;
const __VLS_51 = {
    onClick: (__VLS_ctx.onReset)
};
__VLS_47.slots.default;
(__VLS_ctx.t('common.reset'));
var __VLS_47;
const __VLS_52 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_53 = __VLS_asFunctionalComponent(__VLS_52, new __VLS_52({
    data: (__VLS_ctx.tableData),
    ...{ style: {} },
}));
const __VLS_54 = __VLS_53({
    data: (__VLS_ctx.tableData),
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_53));
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.loading) }, null, null);
__VLS_55.slots.default;
const __VLS_56 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_57 = __VLS_asFunctionalComponent(__VLS_56, new __VLS_56({
    prop: "id",
    label: (__VLS_ctx.t('common.id')),
    width: "90",
}));
const __VLS_58 = __VLS_57({
    prop: "id",
    label: (__VLS_ctx.t('common.id')),
    width: "90",
}, ...__VLS_functionalComponentArgsRest(__VLS_57));
const __VLS_60 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_61 = __VLS_asFunctionalComponent(__VLS_60, new __VLS_60({
    prop: "name",
    label: (__VLS_ctx.t('system.roleName')),
    minWidth: "160",
}));
const __VLS_62 = __VLS_61({
    prop: "name",
    label: (__VLS_ctx.t('system.roleName')),
    minWidth: "160",
}, ...__VLS_functionalComponentArgsRest(__VLS_61));
const __VLS_64 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_65 = __VLS_asFunctionalComponent(__VLS_64, new __VLS_64({
    prop: "code",
    label: (__VLS_ctx.t('system.roleCode')),
    minWidth: "160",
}));
const __VLS_66 = __VLS_65({
    prop: "code",
    label: (__VLS_ctx.t('system.roleCode')),
    minWidth: "160",
}, ...__VLS_functionalComponentArgsRest(__VLS_65));
const __VLS_68 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_69 = __VLS_asFunctionalComponent(__VLS_68, new __VLS_68({
    label: (__VLS_ctx.t('common.status')),
    width: "110",
}));
const __VLS_70 = __VLS_69({
    label: (__VLS_ctx.t('common.status')),
    width: "110",
}, ...__VLS_functionalComponentArgsRest(__VLS_69));
__VLS_71.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_71.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    if (row.status === 1) {
        const __VLS_72 = {}.ElTag;
        /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
        // @ts-ignore
        const __VLS_73 = __VLS_asFunctionalComponent(__VLS_72, new __VLS_72({
            type: "success",
        }));
        const __VLS_74 = __VLS_73({
            type: "success",
        }, ...__VLS_functionalComponentArgsRest(__VLS_73));
        __VLS_75.slots.default;
        (__VLS_ctx.t('common.enabled'));
        var __VLS_75;
    }
    else {
        const __VLS_76 = {}.ElTag;
        /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
        // @ts-ignore
        const __VLS_77 = __VLS_asFunctionalComponent(__VLS_76, new __VLS_76({
            type: "info",
        }));
        const __VLS_78 = __VLS_77({
            type: "info",
        }, ...__VLS_functionalComponentArgsRest(__VLS_77));
        __VLS_79.slots.default;
        (__VLS_ctx.t('common.disabled'));
        var __VLS_79;
    }
}
var __VLS_71;
const __VLS_80 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_81 = __VLS_asFunctionalComponent(__VLS_80, new __VLS_80({
    prop: "remark",
    label: (__VLS_ctx.t('common.remark')),
    minWidth: "200",
}));
const __VLS_82 = __VLS_81({
    prop: "remark",
    label: (__VLS_ctx.t('common.remark')),
    minWidth: "200",
}, ...__VLS_functionalComponentArgsRest(__VLS_81));
const __VLS_84 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_85 = __VLS_asFunctionalComponent(__VLS_84, new __VLS_84({
    label: (__VLS_ctx.t('common.actions')),
    width: "320",
    fixed: "right",
}));
const __VLS_86 = __VLS_85({
    label: (__VLS_ctx.t('common.actions')),
    width: "320",
    fixed: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_85));
__VLS_87.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_87.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_88 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_89 = __VLS_asFunctionalComponent(__VLS_88, new __VLS_88({
        ...{ 'onClick': {} },
        size: "small",
        disabled: (__VLS_ctx.isSuperRole(row)),
    }));
    const __VLS_90 = __VLS_89({
        ...{ 'onClick': {} },
        size: "small",
        disabled: (__VLS_ctx.isSuperRole(row)),
    }, ...__VLS_functionalComponentArgsRest(__VLS_89));
    let __VLS_92;
    let __VLS_93;
    let __VLS_94;
    const __VLS_95 = {
        onClick: (...[$event]) => {
            __VLS_ctx.openUpdate(row);
        }
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPerm)(null, { ...__VLS_directiveBindingRestFields, value: ('sys:role:update') }, null, null);
    __VLS_91.slots.default;
    (__VLS_ctx.t('common.edit'));
    var __VLS_91;
    const __VLS_96 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_97 = __VLS_asFunctionalComponent(__VLS_96, new __VLS_96({
        ...{ 'onClick': {} },
        size: "small",
        disabled: (__VLS_ctx.isSuperRole(row)),
    }));
    const __VLS_98 = __VLS_97({
        ...{ 'onClick': {} },
        size: "small",
        disabled: (__VLS_ctx.isSuperRole(row)),
    }, ...__VLS_functionalComponentArgsRest(__VLS_97));
    let __VLS_100;
    let __VLS_101;
    let __VLS_102;
    const __VLS_103 = {
        onClick: (...[$event]) => {
            __VLS_ctx.openGrant(row);
        }
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPerm)(null, { ...__VLS_directiveBindingRestFields, value: ('sys:role:grant') }, null, null);
    __VLS_99.slots.default;
    (__VLS_ctx.t('system.grant'));
    var __VLS_99;
    const __VLS_104 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_105 = __VLS_asFunctionalComponent(__VLS_104, new __VLS_104({
        ...{ 'onClick': {} },
        size: "small",
        type: "danger",
        disabled: (__VLS_ctx.isSuperRole(row)),
    }));
    const __VLS_106 = __VLS_105({
        ...{ 'onClick': {} },
        size: "small",
        type: "danger",
        disabled: (__VLS_ctx.isSuperRole(row)),
    }, ...__VLS_functionalComponentArgsRest(__VLS_105));
    let __VLS_108;
    let __VLS_109;
    let __VLS_110;
    const __VLS_111 = {
        onClick: (...[$event]) => {
            __VLS_ctx.onDelete(row);
        }
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPerm)(null, { ...__VLS_directiveBindingRestFields, value: ('sys:role:delete') }, null, null);
    __VLS_107.slots.default;
    (__VLS_ctx.t('common.delete'));
    var __VLS_107;
    if (__VLS_ctx.isSuperRole(row)) {
        const __VLS_112 = {}.ElTag;
        /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
        // @ts-ignore
        const __VLS_113 = __VLS_asFunctionalComponent(__VLS_112, new __VLS_112({
            type: "warning",
            ...{ style: {} },
        }));
        const __VLS_114 = __VLS_113({
            type: "warning",
            ...{ style: {} },
        }, ...__VLS_functionalComponentArgsRest(__VLS_113));
        __VLS_115.slots.default;
        (__VLS_ctx.t('system.superAdmin'));
        var __VLS_115;
    }
}
var __VLS_87;
var __VLS_55;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ style: {} },
});
const __VLS_116 = {}.ElPagination;
/** @type {[typeof __VLS_components.ElPagination, typeof __VLS_components.elPagination, ]} */ ;
// @ts-ignore
const __VLS_117 = __VLS_asFunctionalComponent(__VLS_116, new __VLS_116({
    ...{ 'onCurrentChange': {} },
    ...{ 'onSizeChange': {} },
    currentPage: (__VLS_ctx.pagination.page),
    pageSize: (__VLS_ctx.pagination.pageSize),
    total: (__VLS_ctx.pagination.total),
    background: true,
    layout: "total, prev, pager, next, sizes",
}));
const __VLS_118 = __VLS_117({
    ...{ 'onCurrentChange': {} },
    ...{ 'onSizeChange': {} },
    currentPage: (__VLS_ctx.pagination.page),
    pageSize: (__VLS_ctx.pagination.pageSize),
    total: (__VLS_ctx.pagination.total),
    background: true,
    layout: "total, prev, pager, next, sizes",
}, ...__VLS_functionalComponentArgsRest(__VLS_117));
let __VLS_120;
let __VLS_121;
let __VLS_122;
const __VLS_123 = {
    onCurrentChange: (__VLS_ctx.fetchList)
};
const __VLS_124 = {
    onSizeChange: (() => { __VLS_ctx.pagination.page = 1; __VLS_ctx.fetchList(); })
};
var __VLS_119;
var __VLS_3;
const __VLS_125 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_126 = __VLS_asFunctionalComponent(__VLS_125, new __VLS_125({
    modelValue: (__VLS_ctx.editVisible),
    title: (__VLS_ctx.editIsUpdate ? __VLS_ctx.t('system.roleEdit') : __VLS_ctx.t('system.roleAdd')),
    width: "520px",
}));
const __VLS_127 = __VLS_126({
    modelValue: (__VLS_ctx.editVisible),
    title: (__VLS_ctx.editIsUpdate ? __VLS_ctx.t('system.roleEdit') : __VLS_ctx.t('system.roleAdd')),
    width: "520px",
}, ...__VLS_functionalComponentArgsRest(__VLS_126));
__VLS_128.slots.default;
const __VLS_129 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_130 = __VLS_asFunctionalComponent(__VLS_129, new __VLS_129({
    ref: "editFormRef",
    model: (__VLS_ctx.editForm),
    labelWidth: "110px",
}));
const __VLS_131 = __VLS_130({
    ref: "editFormRef",
    model: (__VLS_ctx.editForm),
    labelWidth: "110px",
}, ...__VLS_functionalComponentArgsRest(__VLS_130));
/** @type {typeof __VLS_ctx.editFormRef} */ ;
var __VLS_133 = {};
__VLS_132.slots.default;
const __VLS_135 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_136 = __VLS_asFunctionalComponent(__VLS_135, new __VLS_135({
    label: (__VLS_ctx.t('system.roleName')),
    prop: "name",
    rules: ([{ required: true, message: __VLS_ctx.t('common.required') }]),
}));
const __VLS_137 = __VLS_136({
    label: (__VLS_ctx.t('system.roleName')),
    prop: "name",
    rules: ([{ required: true, message: __VLS_ctx.t('common.required') }]),
}, ...__VLS_functionalComponentArgsRest(__VLS_136));
__VLS_138.slots.default;
const __VLS_139 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_140 = __VLS_asFunctionalComponent(__VLS_139, new __VLS_139({
    modelValue: (__VLS_ctx.editForm.name),
}));
const __VLS_141 = __VLS_140({
    modelValue: (__VLS_ctx.editForm.name),
}, ...__VLS_functionalComponentArgsRest(__VLS_140));
var __VLS_138;
const __VLS_143 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_144 = __VLS_asFunctionalComponent(__VLS_143, new __VLS_143({
    label: (__VLS_ctx.t('system.roleCode')),
    prop: "code",
    rules: ([{ required: true, message: __VLS_ctx.t('common.required') }]),
}));
const __VLS_145 = __VLS_144({
    label: (__VLS_ctx.t('system.roleCode')),
    prop: "code",
    rules: ([{ required: true, message: __VLS_ctx.t('common.required') }]),
}, ...__VLS_functionalComponentArgsRest(__VLS_144));
__VLS_146.slots.default;
const __VLS_147 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_148 = __VLS_asFunctionalComponent(__VLS_147, new __VLS_147({
    modelValue: (__VLS_ctx.editForm.code),
    disabled: (__VLS_ctx.editIsUpdate),
}));
const __VLS_149 = __VLS_148({
    modelValue: (__VLS_ctx.editForm.code),
    disabled: (__VLS_ctx.editIsUpdate),
}, ...__VLS_functionalComponentArgsRest(__VLS_148));
var __VLS_146;
const __VLS_151 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_152 = __VLS_asFunctionalComponent(__VLS_151, new __VLS_151({
    label: (__VLS_ctx.t('common.status')),
    prop: "status",
    rules: ([{ required: true, message: __VLS_ctx.t('common.required') }]),
}));
const __VLS_153 = __VLS_152({
    label: (__VLS_ctx.t('common.status')),
    prop: "status",
    rules: ([{ required: true, message: __VLS_ctx.t('common.required') }]),
}, ...__VLS_functionalComponentArgsRest(__VLS_152));
__VLS_154.slots.default;
const __VLS_155 = {}.ElRadioGroup;
/** @type {[typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, ]} */ ;
// @ts-ignore
const __VLS_156 = __VLS_asFunctionalComponent(__VLS_155, new __VLS_155({
    modelValue: (__VLS_ctx.editForm.status),
}));
const __VLS_157 = __VLS_156({
    modelValue: (__VLS_ctx.editForm.status),
}, ...__VLS_functionalComponentArgsRest(__VLS_156));
__VLS_158.slots.default;
const __VLS_159 = {}.ElRadio;
/** @type {[typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, ]} */ ;
// @ts-ignore
const __VLS_160 = __VLS_asFunctionalComponent(__VLS_159, new __VLS_159({
    label: (1),
}));
const __VLS_161 = __VLS_160({
    label: (1),
}, ...__VLS_functionalComponentArgsRest(__VLS_160));
__VLS_162.slots.default;
(__VLS_ctx.t('common.enabled'));
var __VLS_162;
const __VLS_163 = {}.ElRadio;
/** @type {[typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, ]} */ ;
// @ts-ignore
const __VLS_164 = __VLS_asFunctionalComponent(__VLS_163, new __VLS_163({
    label: (2),
}));
const __VLS_165 = __VLS_164({
    label: (2),
}, ...__VLS_functionalComponentArgsRest(__VLS_164));
__VLS_166.slots.default;
(__VLS_ctx.t('common.disabled'));
var __VLS_166;
var __VLS_158;
var __VLS_154;
const __VLS_167 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_168 = __VLS_asFunctionalComponent(__VLS_167, new __VLS_167({
    label: (__VLS_ctx.t('common.remark')),
    prop: "remark",
}));
const __VLS_169 = __VLS_168({
    label: (__VLS_ctx.t('common.remark')),
    prop: "remark",
}, ...__VLS_functionalComponentArgsRest(__VLS_168));
__VLS_170.slots.default;
const __VLS_171 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_172 = __VLS_asFunctionalComponent(__VLS_171, new __VLS_171({
    modelValue: (__VLS_ctx.editForm.remark),
    type: "textarea",
    rows: (3),
}));
const __VLS_173 = __VLS_172({
    modelValue: (__VLS_ctx.editForm.remark),
    type: "textarea",
    rows: (3),
}, ...__VLS_functionalComponentArgsRest(__VLS_172));
var __VLS_170;
var __VLS_132;
{
    const { footer: __VLS_thisSlot } = __VLS_128.slots;
    const __VLS_175 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_176 = __VLS_asFunctionalComponent(__VLS_175, new __VLS_175({
        ...{ 'onClick': {} },
    }));
    const __VLS_177 = __VLS_176({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_176));
    let __VLS_179;
    let __VLS_180;
    let __VLS_181;
    const __VLS_182 = {
        onClick: (...[$event]) => {
            __VLS_ctx.editVisible = false;
        }
    };
    __VLS_178.slots.default;
    (__VLS_ctx.t('common.cancel'));
    var __VLS_178;
    const __VLS_183 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_184 = __VLS_asFunctionalComponent(__VLS_183, new __VLS_183({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.editLoading),
    }));
    const __VLS_185 = __VLS_184({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.editLoading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_184));
    let __VLS_187;
    let __VLS_188;
    let __VLS_189;
    const __VLS_190 = {
        onClick: (__VLS_ctx.submitEdit)
    };
    __VLS_186.slots.default;
    (__VLS_ctx.t('common.confirm'));
    var __VLS_186;
}
var __VLS_128;
const __VLS_191 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_192 = __VLS_asFunctionalComponent(__VLS_191, new __VLS_191({
    ...{ 'onClosed': {} },
    modelValue: (__VLS_ctx.grantVisible),
    title: (__VLS_ctx.t('system.grantTitle', { role: __VLS_ctx.currentRole?.name || '' })),
    width: "900px",
}));
const __VLS_193 = __VLS_192({
    ...{ 'onClosed': {} },
    modelValue: (__VLS_ctx.grantVisible),
    title: (__VLS_ctx.t('system.grantTitle', { role: __VLS_ctx.currentRole?.name || '' })),
    width: "900px",
}, ...__VLS_functionalComponentArgsRest(__VLS_192));
let __VLS_195;
let __VLS_196;
let __VLS_197;
const __VLS_198 = {
    onClosed: (__VLS_ctx.onGrantClosed)
};
__VLS_194.slots.default;
if (__VLS_ctx.grantReadonly) {
    const __VLS_199 = {}.ElAlert;
    /** @type {[typeof __VLS_components.ElAlert, typeof __VLS_components.elAlert, ]} */ ;
    // @ts-ignore
    const __VLS_200 = __VLS_asFunctionalComponent(__VLS_199, new __VLS_199({
        type: "warning",
        closable: (false),
        title: (__VLS_ctx.t('system.superAdminAllPerms')),
        ...{ style: {} },
    }));
    const __VLS_201 = __VLS_200({
        type: "warning",
        closable: (false),
        title: (__VLS_ctx.t('system.superAdminAllPerms')),
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_200));
}
const __VLS_203 = {}.ElTabs;
/** @type {[typeof __VLS_components.ElTabs, typeof __VLS_components.elTabs, typeof __VLS_components.ElTabs, typeof __VLS_components.elTabs, ]} */ ;
// @ts-ignore
const __VLS_204 = __VLS_asFunctionalComponent(__VLS_203, new __VLS_203({}));
const __VLS_205 = __VLS_204({}, ...__VLS_functionalComponentArgsRest(__VLS_204));
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.grantLoading) }, null, null);
__VLS_206.slots.default;
const __VLS_207 = {}.ElTabPane;
/** @type {[typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, ]} */ ;
// @ts-ignore
const __VLS_208 = __VLS_asFunctionalComponent(__VLS_207, new __VLS_207({
    label: (__VLS_ctx.t('system.grantMenu')),
}));
const __VLS_209 = __VLS_208({
    label: (__VLS_ctx.t('system.grantMenu')),
}, ...__VLS_functionalComponentArgsRest(__VLS_208));
__VLS_210.slots.default;
if (!__VLS_ctx.menuTree || __VLS_ctx.menuTree.length === 0) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    (__VLS_ctx.t('common.noData'));
}
else {
    const __VLS_211 = {}.ElTree;
    /** @type {[typeof __VLS_components.ElTree, typeof __VLS_components.elTree, ]} */ ;
    // @ts-ignore
    const __VLS_212 = __VLS_asFunctionalComponent(__VLS_211, new __VLS_211({
        ref: "menuTreeRef",
        data: (__VLS_ctx.menuTree),
        nodeKey: "id",
        showCheckbox: true,
        props: ({ label: 'name', children: 'children' }),
        checkStrictly: (false),
        disabled: (__VLS_ctx.grantReadonly),
        defaultExpandAll: true,
        ...{ style: {} },
    }));
    const __VLS_213 = __VLS_212({
        ref: "menuTreeRef",
        data: (__VLS_ctx.menuTree),
        nodeKey: "id",
        showCheckbox: true,
        props: ({ label: 'name', children: 'children' }),
        checkStrictly: (false),
        disabled: (__VLS_ctx.grantReadonly),
        defaultExpandAll: true,
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_212));
    /** @type {typeof __VLS_ctx.menuTreeRef} */ ;
    var __VLS_215 = {};
    var __VLS_214;
}
var __VLS_210;
const __VLS_217 = {}.ElTabPane;
/** @type {[typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, ]} */ ;
// @ts-ignore
const __VLS_218 = __VLS_asFunctionalComponent(__VLS_217, new __VLS_217({
    label: (__VLS_ctx.t('system.grantPerms')),
}));
const __VLS_219 = __VLS_218({
    label: (__VLS_ctx.t('system.grantPerms')),
}, ...__VLS_functionalComponentArgsRest(__VLS_218));
__VLS_220.slots.default;
if (!__VLS_ctx.permList || __VLS_ctx.permList.length === 0) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    (__VLS_ctx.t('common.noData'));
}
else {
    const __VLS_221 = {}.ElCheckboxGroup;
    /** @type {[typeof __VLS_components.ElCheckboxGroup, typeof __VLS_components.elCheckboxGroup, typeof __VLS_components.ElCheckboxGroup, typeof __VLS_components.elCheckboxGroup, ]} */ ;
    // @ts-ignore
    const __VLS_222 = __VLS_asFunctionalComponent(__VLS_221, new __VLS_221({
        modelValue: (__VLS_ctx.checkedPermKeys),
        disabled: (__VLS_ctx.grantReadonly),
        ...{ style: {} },
    }));
    const __VLS_223 = __VLS_222({
        modelValue: (__VLS_ctx.checkedPermKeys),
        disabled: (__VLS_ctx.grantReadonly),
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_222));
    __VLS_224.slots.default;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    for (const [p] of __VLS_getVForSourceType((__VLS_ctx.permList))) {
        const __VLS_225 = {}.ElCheckbox;
        /** @type {[typeof __VLS_components.ElCheckbox, typeof __VLS_components.elCheckbox, typeof __VLS_components.ElCheckbox, typeof __VLS_components.elCheckbox, ]} */ ;
        // @ts-ignore
        const __VLS_226 = __VLS_asFunctionalComponent(__VLS_225, new __VLS_225({
            key: (p.key),
            label: (p.key),
            border: true,
        }));
        const __VLS_227 = __VLS_226({
            key: (p.key),
            label: (p.key),
            border: true,
        }, ...__VLS_functionalComponentArgsRest(__VLS_226));
        __VLS_228.slots.default;
        (p.name);
        (p.key);
        var __VLS_228;
    }
    var __VLS_224;
}
var __VLS_220;
var __VLS_206;
{
    const { footer: __VLS_thisSlot } = __VLS_194.slots;
    const __VLS_229 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_230 = __VLS_asFunctionalComponent(__VLS_229, new __VLS_229({
        ...{ 'onClick': {} },
    }));
    const __VLS_231 = __VLS_230({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_230));
    let __VLS_233;
    let __VLS_234;
    let __VLS_235;
    const __VLS_236 = {
        onClick: (...[$event]) => {
            __VLS_ctx.grantVisible = false;
        }
    };
    __VLS_232.slots.default;
    (__VLS_ctx.t('common.cancel'));
    var __VLS_232;
    const __VLS_237 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_238 = __VLS_asFunctionalComponent(__VLS_237, new __VLS_237({
        ...{ 'onClick': {} },
        type: "primary",
        disabled: (__VLS_ctx.grantReadonly),
    }));
    const __VLS_239 = __VLS_238({
        ...{ 'onClick': {} },
        type: "primary",
        disabled: (__VLS_ctx.grantReadonly),
    }, ...__VLS_functionalComponentArgsRest(__VLS_238));
    let __VLS_241;
    let __VLS_242;
    let __VLS_243;
    const __VLS_244 = {
        onClick: (__VLS_ctx.submitGrant)
    };
    __VLS_240.slots.default;
    (__VLS_ctx.t('common.save'));
    var __VLS_240;
}
var __VLS_194;
// @ts-ignore
var __VLS_134 = __VLS_133, __VLS_216 = __VLS_215;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            t: t,
            isSuperRole: isSuperRole,
            pagination: pagination,
            loading: loading,
            queryForm: queryForm,
            tableData: tableData,
            fetchList: fetchList,
            onSearch: onSearch,
            onReset: onReset,
            editVisible: editVisible,
            editFormRef: editFormRef,
            editForm: editForm,
            editIsUpdate: editIsUpdate,
            editLoading: editLoading,
            openCreate: openCreate,
            openUpdate: openUpdate,
            submitEdit: submitEdit,
            onDelete: onDelete,
            grantVisible: grantVisible,
            currentRole: currentRole,
            grantLoading: grantLoading,
            menuTree: menuTree,
            permList: permList,
            menuTreeRef: menuTreeRef,
            checkedPermKeys: checkedPermKeys,
            grantReadonly: grantReadonly,
            openGrant: openGrant,
            submitGrant: submitGrant,
            onGrantClosed: onGrantClosed,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
