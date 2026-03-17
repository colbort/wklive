import { computed, onMounted, ref } from 'vue';
import { ElMessage } from 'element-plus';
import { useI18n } from 'vue-i18n';
import { userService, roleService } from '@/services';
import { ArrowDown } from '@element-plus/icons-vue';
import { usePagination } from '@/composables/usePagination';
import { useLoading } from '@/composables/useLoading';
import { useForm } from '@/composables/useForm';
import { useConfirm } from '@/composables/useConfirm';
const { t } = useI18n();
// Pagination and main list
const { pagination, updateTotal } = usePagination(10);
const list = ref([]);
const { loading, withLoading: withMainLoading } = useLoading();
// Query form
const { form: queryForm } = useForm({
    initialData: {
        keyword: '',
        status: undefined
    }
});
const statusOptions = [
    { label: t('common.enabled'), value: 1 },
    { label: t('common.disabled'), value: 2 },
];
async function fetchList() {
    await withMainLoading(async () => {
        try {
            const res = await userService.getList({
                keyword: queryForm.keyword || undefined,
                status: queryForm.status,
                page: pagination.page,
                size: pagination.pageSize,
            });
            // 兼容 code=0 / 200
            if (res.code !== 0 && res.code !== 200)
                throw new Error(res.msg || 'list failed');
            list.value = res.data || [];
            updateTotal(res.total || 0);
        }
        catch (e) {
            ElMessage.error(e?.message || t('common.loadFailed'));
        }
    });
}
function onSearch() {
    pagination.page = 1;
    fetchList();
}
function onReset() {
    queryForm.keyword = '';
    queryForm.status = undefined;
    pagination.page = 1;
    fetchList();
}
// ---------- 角色缓存（分配角色用） ----------
const { loading: roleLoading, withLoading: withRoleLoading } = useLoading();
const roles = ref([]);
async function fetchRoles() {
    await withRoleLoading(async () => {
        try {
            const res = await roleService.getList({ page: 1, size: 9999, status: 1 });
            if (res.code !== 0 && res.code !== 200)
                throw new Error(res.msg || 'role list failed');
            roles.value = res.data || [];
        }
        catch (e) {
            ElMessage.error(e?.message || t('common.loadFailed'));
        }
    });
}
// ---------- 新增/编辑 ----------
const editVisible = ref(false);
const editMode = ref('create');
const { form: editForm } = useForm({
    initialData: {
        id: 0,
        username: '',
        password: '',
        nickname: '',
        status: 1,
        roleIds: [],
    }
});
const { loading: editFormLoading, withLoading: withEditLoading } = useLoading();
function openCreate() {
    editMode.value = 'create';
    editForm.id = 0;
    editForm.username = '';
    editForm.password = '';
    editForm.nickname = '';
    editForm.status = 1;
    editForm.roleIds = [];
    editVisible.value = true;
}
function openEdit(row) {
    editMode.value = 'update';
    editForm.id = row.id;
    editForm.username = row.username;
    editForm.password = '';
    editForm.nickname = row.nickname || '';
    editForm.status = row.status;
    editForm.roleIds = (row.roleIds || []).slice();
    editVisible.value = true;
}
async function submitEdit() {
    await withEditLoading(async () => {
        try {
            if (editMode.value === 'create') {
                if (!editForm.username || !editForm.password) {
                    ElMessage.warning(t('common.pleaseInputAccountAndPassword'));
                    return;
                }
                const res = await userService.create({
                    username: editForm.username,
                    password: editForm.password,
                    nickname: editForm.nickname || undefined,
                    status: editForm.status,
                    roleIds: editForm.roleIds,
                });
                if (res.code !== 0 && res.code !== 200)
                    throw new Error(res.msg || 'create failed');
                ElMessage.success(t('common.success'));
            }
            else {
                const res = await userService.update(editForm.id, {
                    nickname: editForm.nickname || undefined,
                    status: editForm.status,
                    roleIds: editForm.roleIds,
                });
                if (res.code !== 0 && res.code !== 200)
                    throw new Error(res.msg || 'update failed');
                ElMessage.success(t('common.success'));
            }
            editVisible.value = false;
            fetchList();
        }
        catch (e) {
            ElMessage.error(e?.message || t('common.failed'));
        }
    });
}
// ---------- 删除 ----------
const { confirm } = useConfirm();
async function onDelete(row) {
    try {
        await confirm(t('common.confirmDeleteUser', { username: row.username }), { type: 'warning' });
        const res = await userService.delete(row.id);
        if (res.code !== 0 && res.code !== 200)
            throw new Error(res.msg || 'delete failed');
        ElMessage.success(t('common.success'));
        fetchList();
    }
    catch (e) {
        if (e === 'cancel')
            return;
        ElMessage.error(e?.message || t('common.failed'));
    }
}
// ---------- 启用/禁用 ----------
async function onToggleStatus(row) {
    try {
        const next = row.status === 1 ? 0 : 1;
        const res = await userService.updateUserStatus(row.id, next);
        if (res.code !== 0 && res.code !== 200)
            throw new Error(res.msg || 'status failed');
        ElMessage.success(t('common.success'));
        fetchList();
    }
    catch (e) {
        ElMessage.error(e?.message || t('common.failed'));
    }
}
// ---------- 重置密码 ----------
const pwdVisible = ref(false);
const { form: pwdForm } = useForm({
    initialData: { id: 0, username: '', password: '' }
});
const { loading: pwdSubmitLoading, withLoading: withPwdLoading } = useLoading();
function openResetPwd(row) {
    pwdForm.id = row.id;
    pwdForm.username = row.username;
    pwdForm.password = '';
    pwdVisible.value = true;
}
async function submitResetPwd() {
    await withPwdLoading(async () => {
        try {
            if (!pwdForm.password) {
                ElMessage.warning(t('common.pleaseInputNewPassword'));
                return;
            }
            const res = await userService.resetPassword(pwdForm.id, pwdForm.password);
            if (res.code !== 0 && res.code !== 200)
                throw new Error(res.msg || 'reset pwd failed');
            ElMessage.success(t('common.success'));
            pwdVisible.value = false;
        }
        catch (e) {
            ElMessage.error(e?.message || t('common.failed'));
        }
    });
}
// ---------- 分配角色 ----------
const roleVisible = ref(false);
const { form: roleForm } = useForm({
    initialData: { userId: 0, username: '', roleIds: [] }
});
const { loading: roleAssignLoading, withLoading: withRoleAssignLoading } = useLoading();
function openAssignRoles(row) {
    roleForm.userId = row.id;
    roleForm.username = row.username;
    roleForm.roleIds = (row.roleIds || []).slice();
    roleVisible.value = true;
}
async function submitAssignRoles() {
    await withRoleAssignLoading(async () => {
        try {
            const res = await userService.assignUserRoles(roleForm.userId, roleForm.roleIds);
            if (res.code !== 0 && res.code !== 200)
                throw new Error(res.msg || 'assign roles failed');
            ElMessage.success(t('common.success'));
            roleVisible.value = false;
            fetchList();
        }
        catch (e) {
            ElMessage.error(e?.message || t('common.failed'));
        }
    });
}
// ---------- Google 2FA ----------
const g2Visible = ref(false);
const { form: g2User } = useForm({
    initialData: { userId: 0, username: '' }
});
const { form: g2Init } = useForm({
    initialData: { secret: '', otpauthUrl: '', qrCode: '' }
});
const { form: g2Form } = useForm({
    initialData: { code: '' }
});
const { loading: g2InitLoading, withLoading: withG2InitLoading } = useLoading();
const { loading: g2EnableLoading, withLoading: withG2EnableLoading } = useLoading();
const { loading: g2DisableLoading, withLoading: withG2DisableLoading } = useLoading();
function openGoogle2fa(row) {
    g2User.userId = row.id;
    g2User.username = row.username;
    g2Init.secret = '';
    g2Init.otpauthUrl = '';
    g2Init.qrCode = '';
    g2Form.code = '';
    g2Visible.value = true;
}
async function doG2Init() {
    await withG2InitLoading(async () => {
        try {
            const res = await userService.initGoogle2FA(g2User.userId);
            if (res.code !== 200)
                throw new Error(res.msg || 'init failed');
            g2Init.secret = res.data?.secret || '';
            g2Init.otpauthUrl = res.data?.otpauthUrl || '';
            g2Init.qrCode = res.data?.qrCode || '';
            ElMessage.success(t('common.success'));
        }
        catch (e) {
            ElMessage.error(e?.message || t('common.failed'));
        }
    });
}
async function doG2Enable() {
    await withG2EnableLoading(async () => {
        try {
            if (!g2Form.code) {
                ElMessage.warning(t('common.pleaseInputCode'));
                return;
            }
            const res = await userService.enableGoogle2FA(g2User.userId, g2Form.code);
            if (res.code !== 0 && res.code !== 200)
                throw new Error(res.msg || 'enable failed');
            ElMessage.success(t('common.success'));
            fetchList();
        }
        catch (e) {
            ElMessage.error(e?.message || t('common.failed'));
        }
    });
}
async function doG2Disable() {
    await withG2DisableLoading(async () => {
        try {
            const res = await userService.disableGoogle2FA(g2User.userId, g2Form.code || undefined);
            if (res.code !== 0 && res.code !== 200)
                throw new Error(res.msg || 'disable failed');
            ElMessage.success(t('common.success'));
            fetchList();
        }
        catch (e) {
            ElMessage.error(e?.message || t('common.failed'));
        }
    });
}
async function doG2Reset() {
    try {
        await confirm(t('common.confirmReset2fa'), { type: 'warning' });
        const res = await userService.resetGoogle2FA(g2User.userId);
        if (res.code !== 0 && res.code !== 200)
            throw new Error(res.msg || 'reset failed');
        ElMessage.success(t('common.success'));
        fetchList();
    }
    catch (e) {
        if (e === 'cancel')
            return;
        ElMessage.error(e?.message || t('common.failed'));
    }
}
const roleNameMap = computed(() => {
    const m = new Map();
    roles.value.forEach((r) => m.set(r.id, r.name));
    return m;
});
onMounted(async () => {
    await Promise.all([fetchRoles(), fetchList()]);
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
__VLS_3.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_3.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
    (__VLS_ctx.t('system.users'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    const __VLS_4 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_5 = __VLS_asFunctionalComponent(__VLS_4, new __VLS_4({
        ...{ 'onKeyup': {} },
        modelValue: (__VLS_ctx.queryForm.keyword),
        ...{ style: {} },
        clearable: true,
        placeholder: (__VLS_ctx.t('common.accountNicknameKeyword')),
    }));
    const __VLS_6 = __VLS_5({
        ...{ 'onKeyup': {} },
        modelValue: (__VLS_ctx.queryForm.keyword),
        ...{ style: {} },
        clearable: true,
        placeholder: (__VLS_ctx.t('common.accountNicknameKeyword')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_5));
    let __VLS_8;
    let __VLS_9;
    let __VLS_10;
    const __VLS_11 = {
        onKeyup: (__VLS_ctx.onSearch)
    };
    var __VLS_7;
    const __VLS_12 = {}.ElSelect;
    /** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
    // @ts-ignore
    const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
        modelValue: (__VLS_ctx.queryForm.status),
        ...{ style: {} },
        clearable: true,
        placeholder: (__VLS_ctx.t('common.status')),
    }));
    const __VLS_14 = __VLS_13({
        modelValue: (__VLS_ctx.queryForm.status),
        ...{ style: {} },
        clearable: true,
        placeholder: (__VLS_ctx.t('common.status')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_13));
    __VLS_15.slots.default;
    for (const [o] of __VLS_getVForSourceType((__VLS_ctx.statusOptions))) {
        const __VLS_16 = {}.ElOption;
        /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
        // @ts-ignore
        const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
            key: (o.value),
            label: (o.label),
            value: (o.value),
        }));
        const __VLS_18 = __VLS_17({
            key: (o.value),
            label: (o.label),
            value: (o.value),
        }, ...__VLS_functionalComponentArgsRest(__VLS_17));
    }
    var __VLS_15;
    const __VLS_20 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
        ...{ 'onClick': {} },
    }));
    const __VLS_22 = __VLS_21({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_21));
    let __VLS_24;
    let __VLS_25;
    let __VLS_26;
    const __VLS_27 = {
        onClick: (__VLS_ctx.onSearch)
    };
    __VLS_23.slots.default;
    (__VLS_ctx.t('common.search'));
    var __VLS_23;
    const __VLS_28 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({
        ...{ 'onClick': {} },
    }));
    const __VLS_30 = __VLS_29({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_29));
    let __VLS_32;
    let __VLS_33;
    let __VLS_34;
    const __VLS_35 = {
        onClick: (__VLS_ctx.onReset)
    };
    __VLS_31.slots.default;
    (__VLS_ctx.t('common.reset'));
    var __VLS_31;
    const __VLS_36 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_38 = __VLS_37({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_37));
    let __VLS_40;
    let __VLS_41;
    let __VLS_42;
    const __VLS_43 = {
        onClick: (__VLS_ctx.openCreate)
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPerm)(null, { ...__VLS_directiveBindingRestFields, value: ('sys:user:add') }, null, null);
    __VLS_39.slots.default;
    (__VLS_ctx.t('perms.sys:user:add'));
    var __VLS_39;
}
const __VLS_44 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_45 = __VLS_asFunctionalComponent(__VLS_44, new __VLS_44({
    data: (__VLS_ctx.list),
    rowKey: "id",
}));
const __VLS_46 = __VLS_45({
    data: (__VLS_ctx.list),
    rowKey: "id",
}, ...__VLS_functionalComponentArgsRest(__VLS_45));
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.loading) }, null, null);
__VLS_47.slots.default;
const __VLS_48 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_49 = __VLS_asFunctionalComponent(__VLS_48, new __VLS_48({
    prop: "id",
    label: (__VLS_ctx.t('common.id')),
    width: "80",
}));
const __VLS_50 = __VLS_49({
    prop: "id",
    label: (__VLS_ctx.t('common.id')),
    width: "80",
}, ...__VLS_functionalComponentArgsRest(__VLS_49));
const __VLS_52 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_53 = __VLS_asFunctionalComponent(__VLS_52, new __VLS_52({
    prop: "username",
    label: (__VLS_ctx.t('common.username')),
    minWidth: "140",
}));
const __VLS_54 = __VLS_53({
    prop: "username",
    label: (__VLS_ctx.t('common.username')),
    minWidth: "140",
}, ...__VLS_functionalComponentArgsRest(__VLS_53));
const __VLS_56 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_57 = __VLS_asFunctionalComponent(__VLS_56, new __VLS_56({
    prop: "nickname",
    label: (__VLS_ctx.t('common.nickname')),
    minWidth: "160",
}));
const __VLS_58 = __VLS_57({
    prop: "nickname",
    label: (__VLS_ctx.t('common.nickname')),
    minWidth: "160",
}, ...__VLS_functionalComponentArgsRest(__VLS_57));
const __VLS_60 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_61 = __VLS_asFunctionalComponent(__VLS_60, new __VLS_60({
    label: (__VLS_ctx.t('common.role')),
    minWidth: "180",
}));
const __VLS_62 = __VLS_61({
    label: (__VLS_ctx.t('common.role')),
    minWidth: "180",
}, ...__VLS_functionalComponentArgsRest(__VLS_61));
__VLS_63.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_63.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    for (const [rid] of __VLS_getVForSourceType(((row.roleIds || [])))) {
        const __VLS_64 = {}.ElTag;
        /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
        // @ts-ignore
        const __VLS_65 = __VLS_asFunctionalComponent(__VLS_64, new __VLS_64({
            key: (rid),
            ...{ style: {} },
        }));
        const __VLS_66 = __VLS_65({
            key: (rid),
            ...{ style: {} },
        }, ...__VLS_functionalComponentArgsRest(__VLS_65));
        __VLS_67.slots.default;
        (__VLS_ctx.roleNameMap.get(rid) || ('#' + rid));
        var __VLS_67;
    }
    if (!(row.roleIds && row.roleIds.length)) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ style: {} },
        });
    }
}
var __VLS_63;
const __VLS_68 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_69 = __VLS_asFunctionalComponent(__VLS_68, new __VLS_68({
    label: (__VLS_ctx.t('common.google2fa')),
    width: "110",
}));
const __VLS_70 = __VLS_69({
    label: (__VLS_ctx.t('common.google2fa')),
    width: "110",
}, ...__VLS_functionalComponentArgsRest(__VLS_69));
__VLS_71.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_71.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_72 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_73 = __VLS_asFunctionalComponent(__VLS_72, new __VLS_72({
        type: (row.google2faEnabled === 1 ? 'success' : 'info'),
    }));
    const __VLS_74 = __VLS_73({
        type: (row.google2faEnabled === 1 ? 'success' : 'info'),
    }, ...__VLS_functionalComponentArgsRest(__VLS_73));
    __VLS_75.slots.default;
    (row.google2faEnabled === 1 ? __VLS_ctx.t('common.enabled') : __VLS_ctx.t('common.disabled'));
    var __VLS_75;
}
var __VLS_71;
const __VLS_76 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_77 = __VLS_asFunctionalComponent(__VLS_76, new __VLS_76({
    label: (__VLS_ctx.t('common.status')),
    width: "110",
}));
const __VLS_78 = __VLS_77({
    label: (__VLS_ctx.t('common.status')),
    width: "110",
}, ...__VLS_functionalComponentArgsRest(__VLS_77));
__VLS_79.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_79.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_80 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_81 = __VLS_asFunctionalComponent(__VLS_80, new __VLS_80({
        type: (row.status === 1 ? 'success' : 'danger'),
    }));
    const __VLS_82 = __VLS_81({
        type: (row.status === 1 ? 'success' : 'danger'),
    }, ...__VLS_functionalComponentArgsRest(__VLS_81));
    __VLS_83.slots.default;
    (row.status === 1 ? __VLS_ctx.t('common.enabled') : __VLS_ctx.t('common.disabled'));
    var __VLS_83;
}
var __VLS_79;
const __VLS_84 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_85 = __VLS_asFunctionalComponent(__VLS_84, new __VLS_84({
    label: (__VLS_ctx.t('common.createdAt')),
    minWidth: "170",
}));
const __VLS_86 = __VLS_85({
    label: (__VLS_ctx.t('common.createdAt')),
    minWidth: "170",
}, ...__VLS_functionalComponentArgsRest(__VLS_85));
__VLS_87.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_87.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ style: {} },
    });
    (row.createdAt ? new Date(row.createdAt * 1000).toLocaleString() : '-');
}
var __VLS_87;
const __VLS_88 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_89 = __VLS_asFunctionalComponent(__VLS_88, new __VLS_88({
    label: (__VLS_ctx.t('common.actions')),
    width: "140",
    fixed: "right",
}));
const __VLS_90 = __VLS_89({
    label: (__VLS_ctx.t('common.actions')),
    width: "140",
    fixed: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_89));
__VLS_91.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_91.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_92 = {}.ElDropdown;
    /** @type {[typeof __VLS_components.ElDropdown, typeof __VLS_components.elDropdown, typeof __VLS_components.ElDropdown, typeof __VLS_components.elDropdown, ]} */ ;
    // @ts-ignore
    const __VLS_93 = __VLS_asFunctionalComponent(__VLS_92, new __VLS_92({
        trigger: "click",
    }));
    const __VLS_94 = __VLS_93({
        trigger: "click",
    }, ...__VLS_functionalComponentArgsRest(__VLS_93));
    __VLS_95.slots.default;
    const __VLS_96 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_97 = __VLS_asFunctionalComponent(__VLS_96, new __VLS_96({
        size: "small",
    }));
    const __VLS_98 = __VLS_97({
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_97));
    __VLS_99.slots.default;
    (__VLS_ctx.t('common.actions'));
    const __VLS_100 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_101 = __VLS_asFunctionalComponent(__VLS_100, new __VLS_100({
        ...{ class: "el-icon--right" },
    }));
    const __VLS_102 = __VLS_101({
        ...{ class: "el-icon--right" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_101));
    __VLS_103.slots.default;
    const __VLS_104 = {}.ArrowDown;
    /** @type {[typeof __VLS_components.ArrowDown, ]} */ ;
    // @ts-ignore
    const __VLS_105 = __VLS_asFunctionalComponent(__VLS_104, new __VLS_104({}));
    const __VLS_106 = __VLS_105({}, ...__VLS_functionalComponentArgsRest(__VLS_105));
    var __VLS_103;
    var __VLS_99;
    {
        const { dropdown: __VLS_thisSlot } = __VLS_95.slots;
        const __VLS_108 = {}.ElDropdownMenu;
        /** @type {[typeof __VLS_components.ElDropdownMenu, typeof __VLS_components.elDropdownMenu, typeof __VLS_components.ElDropdownMenu, typeof __VLS_components.elDropdownMenu, ]} */ ;
        // @ts-ignore
        const __VLS_109 = __VLS_asFunctionalComponent(__VLS_108, new __VLS_108({}));
        const __VLS_110 = __VLS_109({}, ...__VLS_functionalComponentArgsRest(__VLS_109));
        __VLS_111.slots.default;
        const __VLS_112 = {}.ElDropdownItem;
        /** @type {[typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, ]} */ ;
        // @ts-ignore
        const __VLS_113 = __VLS_asFunctionalComponent(__VLS_112, new __VLS_112({
            ...{ 'onClick': {} },
        }));
        const __VLS_114 = __VLS_113({
            ...{ 'onClick': {} },
        }, ...__VLS_functionalComponentArgsRest(__VLS_113));
        let __VLS_116;
        let __VLS_117;
        let __VLS_118;
        const __VLS_119 = {
            onClick: (...[$event]) => {
                __VLS_ctx.openEdit(row);
            }
        };
        __VLS_asFunctionalDirective(__VLS_directives.vPerm)(null, { ...__VLS_directiveBindingRestFields, value: ('sys:user:update') }, null, null);
        __VLS_115.slots.default;
        (__VLS_ctx.t('perms.sys:user:update'));
        var __VLS_115;
        const __VLS_120 = {}.ElDropdownItem;
        /** @type {[typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, ]} */ ;
        // @ts-ignore
        const __VLS_121 = __VLS_asFunctionalComponent(__VLS_120, new __VLS_120({
            ...{ 'onClick': {} },
        }));
        const __VLS_122 = __VLS_121({
            ...{ 'onClick': {} },
        }, ...__VLS_functionalComponentArgsRest(__VLS_121));
        let __VLS_124;
        let __VLS_125;
        let __VLS_126;
        const __VLS_127 = {
            onClick: (...[$event]) => {
                __VLS_ctx.openResetPwd(row);
            }
        };
        __VLS_asFunctionalDirective(__VLS_directives.vPerm)(null, { ...__VLS_directiveBindingRestFields, value: ('sys:user:resetpwd') }, null, null);
        __VLS_123.slots.default;
        (__VLS_ctx.t('perms.sys:user:resetpwd'));
        var __VLS_123;
        const __VLS_128 = {}.ElDropdownItem;
        /** @type {[typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, ]} */ ;
        // @ts-ignore
        const __VLS_129 = __VLS_asFunctionalComponent(__VLS_128, new __VLS_128({
            ...{ 'onClick': {} },
        }));
        const __VLS_130 = __VLS_129({
            ...{ 'onClick': {} },
        }, ...__VLS_functionalComponentArgsRest(__VLS_129));
        let __VLS_132;
        let __VLS_133;
        let __VLS_134;
        const __VLS_135 = {
            onClick: (...[$event]) => {
                __VLS_ctx.openAssignRoles(row);
            }
        };
        __VLS_asFunctionalDirective(__VLS_directives.vPerm)(null, { ...__VLS_directiveBindingRestFields, value: ('sys:user:assignrole') }, null, null);
        __VLS_131.slots.default;
        (__VLS_ctx.t('perms.sys:user:assignrole'));
        var __VLS_131;
        const __VLS_136 = {}.ElDropdownItem;
        /** @type {[typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, ]} */ ;
        // @ts-ignore
        const __VLS_137 = __VLS_asFunctionalComponent(__VLS_136, new __VLS_136({
            ...{ 'onClick': {} },
        }));
        const __VLS_138 = __VLS_137({
            ...{ 'onClick': {} },
        }, ...__VLS_functionalComponentArgsRest(__VLS_137));
        let __VLS_140;
        let __VLS_141;
        let __VLS_142;
        const __VLS_143 = {
            onClick: (...[$event]) => {
                __VLS_ctx.openGoogle2fa(row);
            }
        };
        __VLS_asFunctionalDirective(__VLS_directives.vPerm)(null, { ...__VLS_directiveBindingRestFields, value: ('sys:user:google2fa') }, null, null);
        __VLS_139.slots.default;
        (__VLS_ctx.t('perms.sys:user:google2fa'));
        var __VLS_139;
        const __VLS_144 = {}.ElDropdownItem;
        /** @type {[typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, ]} */ ;
        // @ts-ignore
        const __VLS_145 = __VLS_asFunctionalComponent(__VLS_144, new __VLS_144({
            ...{ 'onClick': {} },
            divided: true,
        }));
        const __VLS_146 = __VLS_145({
            ...{ 'onClick': {} },
            divided: true,
        }, ...__VLS_functionalComponentArgsRest(__VLS_145));
        let __VLS_148;
        let __VLS_149;
        let __VLS_150;
        const __VLS_151 = {
            onClick: (...[$event]) => {
                __VLS_ctx.onToggleStatus(row);
            }
        };
        __VLS_asFunctionalDirective(__VLS_directives.vPerm)(null, { ...__VLS_directiveBindingRestFields, value: ('sys:user:update') }, null, null);
        __VLS_147.slots.default;
        (row.status === 1 ? __VLS_ctx.t('common.disable') : __VLS_ctx.t('common.enable'));
        var __VLS_147;
        const __VLS_152 = {}.ElDropdownItem;
        /** @type {[typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, ]} */ ;
        // @ts-ignore
        const __VLS_153 = __VLS_asFunctionalComponent(__VLS_152, new __VLS_152({
            ...{ 'onClick': {} },
        }));
        const __VLS_154 = __VLS_153({
            ...{ 'onClick': {} },
        }, ...__VLS_functionalComponentArgsRest(__VLS_153));
        let __VLS_156;
        let __VLS_157;
        let __VLS_158;
        const __VLS_159 = {
            onClick: (...[$event]) => {
                __VLS_ctx.onDelete(row);
            }
        };
        __VLS_asFunctionalDirective(__VLS_directives.vPerm)(null, { ...__VLS_directiveBindingRestFields, value: ('sys:user:delete') }, null, null);
        __VLS_155.slots.default;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ style: {} },
        });
        (__VLS_ctx.t('perms.sys:user:delete'));
        var __VLS_155;
        var __VLS_111;
    }
    var __VLS_95;
}
var __VLS_91;
var __VLS_47;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ style: {} },
});
const __VLS_160 = {}.ElPagination;
/** @type {[typeof __VLS_components.ElPagination, typeof __VLS_components.elPagination, ]} */ ;
// @ts-ignore
const __VLS_161 = __VLS_asFunctionalComponent(__VLS_160, new __VLS_160({
    ...{ 'onUpdate:currentPage': {} },
    ...{ 'onUpdate:pageSize': {} },
    background: true,
    layout: "total, prev, pager, next, sizes",
    total: (__VLS_ctx.pagination.total),
    pageSize: (__VLS_ctx.pagination.pageSize),
    currentPage: (__VLS_ctx.pagination.page),
}));
const __VLS_162 = __VLS_161({
    ...{ 'onUpdate:currentPage': {} },
    ...{ 'onUpdate:pageSize': {} },
    background: true,
    layout: "total, prev, pager, next, sizes",
    total: (__VLS_ctx.pagination.total),
    pageSize: (__VLS_ctx.pagination.pageSize),
    currentPage: (__VLS_ctx.pagination.page),
}, ...__VLS_functionalComponentArgsRest(__VLS_161));
let __VLS_164;
let __VLS_165;
let __VLS_166;
const __VLS_167 = {
    'onUpdate:currentPage': ((p) => { __VLS_ctx.pagination.page = p; __VLS_ctx.fetchList(); })
};
const __VLS_168 = {
    'onUpdate:pageSize': ((s) => { __VLS_ctx.pagination.pageSize = s; __VLS_ctx.pagination.page = 1; __VLS_ctx.fetchList(); })
};
var __VLS_163;
var __VLS_3;
const __VLS_169 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_170 = __VLS_asFunctionalComponent(__VLS_169, new __VLS_169({
    modelValue: (__VLS_ctx.editVisible),
    title: (__VLS_ctx.editMode === 'create' ? __VLS_ctx.t('common.addUser') : __VLS_ctx.t('common.editUser')),
    width: "520px",
}));
const __VLS_171 = __VLS_170({
    modelValue: (__VLS_ctx.editVisible),
    title: (__VLS_ctx.editMode === 'create' ? __VLS_ctx.t('common.addUser') : __VLS_ctx.t('common.editUser')),
    width: "520px",
}, ...__VLS_functionalComponentArgsRest(__VLS_170));
__VLS_172.slots.default;
const __VLS_173 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_174 = __VLS_asFunctionalComponent(__VLS_173, new __VLS_173({
    labelWidth: "90px",
}));
const __VLS_175 = __VLS_174({
    labelWidth: "90px",
}, ...__VLS_functionalComponentArgsRest(__VLS_174));
__VLS_176.slots.default;
if (__VLS_ctx.editMode === 'create') {
    const __VLS_177 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_178 = __VLS_asFunctionalComponent(__VLS_177, new __VLS_177({
        label: (__VLS_ctx.t('common.username')),
    }));
    const __VLS_179 = __VLS_178({
        label: (__VLS_ctx.t('common.username')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_178));
    __VLS_180.slots.default;
    const __VLS_181 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_182 = __VLS_asFunctionalComponent(__VLS_181, new __VLS_181({
        modelValue: (__VLS_ctx.editForm.username),
    }));
    const __VLS_183 = __VLS_182({
        modelValue: (__VLS_ctx.editForm.username),
    }, ...__VLS_functionalComponentArgsRest(__VLS_182));
    var __VLS_180;
}
if (__VLS_ctx.editMode === 'create') {
    const __VLS_185 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_186 = __VLS_asFunctionalComponent(__VLS_185, new __VLS_185({
        label: (__VLS_ctx.t('common.password')),
    }));
    const __VLS_187 = __VLS_186({
        label: (__VLS_ctx.t('common.password')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_186));
    __VLS_188.slots.default;
    const __VLS_189 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_190 = __VLS_asFunctionalComponent(__VLS_189, new __VLS_189({
        modelValue: (__VLS_ctx.editForm.password),
        type: "password",
        showPassword: true,
    }));
    const __VLS_191 = __VLS_190({
        modelValue: (__VLS_ctx.editForm.password),
        type: "password",
        showPassword: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_190));
    var __VLS_188;
}
const __VLS_193 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_194 = __VLS_asFunctionalComponent(__VLS_193, new __VLS_193({
    label: (__VLS_ctx.t('common.nickname')),
}));
const __VLS_195 = __VLS_194({
    label: (__VLS_ctx.t('common.nickname')),
}, ...__VLS_functionalComponentArgsRest(__VLS_194));
__VLS_196.slots.default;
const __VLS_197 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_198 = __VLS_asFunctionalComponent(__VLS_197, new __VLS_197({
    modelValue: (__VLS_ctx.editForm.nickname),
}));
const __VLS_199 = __VLS_198({
    modelValue: (__VLS_ctx.editForm.nickname),
}, ...__VLS_functionalComponentArgsRest(__VLS_198));
var __VLS_196;
const __VLS_201 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_202 = __VLS_asFunctionalComponent(__VLS_201, new __VLS_201({
    label: (__VLS_ctx.t('common.status')),
}));
const __VLS_203 = __VLS_202({
    label: (__VLS_ctx.t('common.status')),
}, ...__VLS_functionalComponentArgsRest(__VLS_202));
__VLS_204.slots.default;
const __VLS_205 = {}.ElSwitch;
/** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
// @ts-ignore
const __VLS_206 = __VLS_asFunctionalComponent(__VLS_205, new __VLS_205({
    modelValue: (__VLS_ctx.editForm.status),
    activeValue: (1),
    inactiveValue: (0),
}));
const __VLS_207 = __VLS_206({
    modelValue: (__VLS_ctx.editForm.status),
    activeValue: (1),
    inactiveValue: (0),
}, ...__VLS_functionalComponentArgsRest(__VLS_206));
var __VLS_204;
const __VLS_209 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_210 = __VLS_asFunctionalComponent(__VLS_209, new __VLS_209({
    label: (__VLS_ctx.t('common.role')),
}));
const __VLS_211 = __VLS_210({
    label: (__VLS_ctx.t('common.role')),
}, ...__VLS_functionalComponentArgsRest(__VLS_210));
__VLS_212.slots.default;
const __VLS_213 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_214 = __VLS_asFunctionalComponent(__VLS_213, new __VLS_213({
    modelValue: (__VLS_ctx.editForm.roleIds),
    multiple: true,
    filterable: true,
    ...{ style: {} },
    loading: (__VLS_ctx.roleLoading),
}));
const __VLS_215 = __VLS_214({
    modelValue: (__VLS_ctx.editForm.roleIds),
    multiple: true,
    filterable: true,
    ...{ style: {} },
    loading: (__VLS_ctx.roleLoading),
}, ...__VLS_functionalComponentArgsRest(__VLS_214));
__VLS_216.slots.default;
for (const [r] of __VLS_getVForSourceType((__VLS_ctx.roles))) {
    const __VLS_217 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_218 = __VLS_asFunctionalComponent(__VLS_217, new __VLS_217({
        key: (r.id),
        label: (r.name),
        value: (r.id),
    }));
    const __VLS_219 = __VLS_218({
        key: (r.id),
        label: (r.name),
        value: (r.id),
    }, ...__VLS_functionalComponentArgsRest(__VLS_218));
}
var __VLS_216;
var __VLS_212;
var __VLS_176;
{
    const { footer: __VLS_thisSlot } = __VLS_172.slots;
    const __VLS_221 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_222 = __VLS_asFunctionalComponent(__VLS_221, new __VLS_221({
        ...{ 'onClick': {} },
    }));
    const __VLS_223 = __VLS_222({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_222));
    let __VLS_225;
    let __VLS_226;
    let __VLS_227;
    const __VLS_228 = {
        onClick: (...[$event]) => {
            __VLS_ctx.editVisible = false;
        }
    };
    __VLS_224.slots.default;
    (__VLS_ctx.t('common.cancel'));
    var __VLS_224;
    const __VLS_229 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_230 = __VLS_asFunctionalComponent(__VLS_229, new __VLS_229({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.editFormLoading),
    }));
    const __VLS_231 = __VLS_230({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.editFormLoading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_230));
    let __VLS_233;
    let __VLS_234;
    let __VLS_235;
    const __VLS_236 = {
        onClick: (__VLS_ctx.submitEdit)
    };
    __VLS_232.slots.default;
    (__VLS_ctx.t('common.confirm'));
    var __VLS_232;
}
var __VLS_172;
const __VLS_237 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_238 = __VLS_asFunctionalComponent(__VLS_237, new __VLS_237({
    modelValue: (__VLS_ctx.pwdVisible),
    title: (__VLS_ctx.t('common.resetPassword')),
    width: "420px",
}));
const __VLS_239 = __VLS_238({
    modelValue: (__VLS_ctx.pwdVisible),
    title: (__VLS_ctx.t('common.resetPassword')),
    width: "420px",
}, ...__VLS_functionalComponentArgsRest(__VLS_238));
__VLS_240.slots.default;
const __VLS_241 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_242 = __VLS_asFunctionalComponent(__VLS_241, new __VLS_241({
    labelWidth: "90px",
}));
const __VLS_243 = __VLS_242({
    labelWidth: "90px",
}, ...__VLS_functionalComponentArgsRest(__VLS_242));
__VLS_244.slots.default;
const __VLS_245 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_246 = __VLS_asFunctionalComponent(__VLS_245, new __VLS_245({
    label: (__VLS_ctx.t('common.username')),
}));
const __VLS_247 = __VLS_246({
    label: (__VLS_ctx.t('common.username')),
}, ...__VLS_functionalComponentArgsRest(__VLS_246));
__VLS_248.slots.default;
const __VLS_249 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_250 = __VLS_asFunctionalComponent(__VLS_249, new __VLS_249({
    modelValue: (__VLS_ctx.pwdForm.username),
    disabled: true,
}));
const __VLS_251 = __VLS_250({
    modelValue: (__VLS_ctx.pwdForm.username),
    disabled: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_250));
var __VLS_248;
const __VLS_253 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_254 = __VLS_asFunctionalComponent(__VLS_253, new __VLS_253({
    label: (__VLS_ctx.t('common.newPassword')),
}));
const __VLS_255 = __VLS_254({
    label: (__VLS_ctx.t('common.newPassword')),
}, ...__VLS_functionalComponentArgsRest(__VLS_254));
__VLS_256.slots.default;
const __VLS_257 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_258 = __VLS_asFunctionalComponent(__VLS_257, new __VLS_257({
    modelValue: (__VLS_ctx.pwdForm.password),
    type: "password",
    showPassword: true,
}));
const __VLS_259 = __VLS_258({
    modelValue: (__VLS_ctx.pwdForm.password),
    type: "password",
    showPassword: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_258));
var __VLS_256;
var __VLS_244;
{
    const { footer: __VLS_thisSlot } = __VLS_240.slots;
    const __VLS_261 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_262 = __VLS_asFunctionalComponent(__VLS_261, new __VLS_261({
        ...{ 'onClick': {} },
    }));
    const __VLS_263 = __VLS_262({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_262));
    let __VLS_265;
    let __VLS_266;
    let __VLS_267;
    const __VLS_268 = {
        onClick: (...[$event]) => {
            __VLS_ctx.pwdVisible = false;
        }
    };
    __VLS_264.slots.default;
    (__VLS_ctx.t('common.cancel'));
    var __VLS_264;
    const __VLS_269 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_270 = __VLS_asFunctionalComponent(__VLS_269, new __VLS_269({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.pwdSubmitLoading),
    }));
    const __VLS_271 = __VLS_270({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.pwdSubmitLoading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_270));
    let __VLS_273;
    let __VLS_274;
    let __VLS_275;
    const __VLS_276 = {
        onClick: (__VLS_ctx.submitResetPwd)
    };
    __VLS_272.slots.default;
    (__VLS_ctx.t('common.confirm'));
    var __VLS_272;
}
var __VLS_240;
const __VLS_277 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_278 = __VLS_asFunctionalComponent(__VLS_277, new __VLS_277({
    modelValue: (__VLS_ctx.roleVisible),
    title: (__VLS_ctx.t('common.assignRoles')),
    width: "520px",
}));
const __VLS_279 = __VLS_278({
    modelValue: (__VLS_ctx.roleVisible),
    title: (__VLS_ctx.t('common.assignRoles')),
    width: "520px",
}, ...__VLS_functionalComponentArgsRest(__VLS_278));
__VLS_280.slots.default;
const __VLS_281 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_282 = __VLS_asFunctionalComponent(__VLS_281, new __VLS_281({
    labelWidth: "90px",
}));
const __VLS_283 = __VLS_282({
    labelWidth: "90px",
}, ...__VLS_functionalComponentArgsRest(__VLS_282));
__VLS_284.slots.default;
const __VLS_285 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_286 = __VLS_asFunctionalComponent(__VLS_285, new __VLS_285({
    label: (__VLS_ctx.t('common.username')),
}));
const __VLS_287 = __VLS_286({
    label: (__VLS_ctx.t('common.username')),
}, ...__VLS_functionalComponentArgsRest(__VLS_286));
__VLS_288.slots.default;
const __VLS_289 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_290 = __VLS_asFunctionalComponent(__VLS_289, new __VLS_289({
    modelValue: (__VLS_ctx.roleForm.username),
    disabled: true,
}));
const __VLS_291 = __VLS_290({
    modelValue: (__VLS_ctx.roleForm.username),
    disabled: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_290));
var __VLS_288;
const __VLS_293 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_294 = __VLS_asFunctionalComponent(__VLS_293, new __VLS_293({
    label: (__VLS_ctx.t('common.role')),
}));
const __VLS_295 = __VLS_294({
    label: (__VLS_ctx.t('common.role')),
}, ...__VLS_functionalComponentArgsRest(__VLS_294));
__VLS_296.slots.default;
const __VLS_297 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_298 = __VLS_asFunctionalComponent(__VLS_297, new __VLS_297({
    modelValue: (__VLS_ctx.roleForm.roleIds),
    multiple: true,
    filterable: true,
    ...{ style: {} },
    loading: (__VLS_ctx.roleLoading),
}));
const __VLS_299 = __VLS_298({
    modelValue: (__VLS_ctx.roleForm.roleIds),
    multiple: true,
    filterable: true,
    ...{ style: {} },
    loading: (__VLS_ctx.roleLoading),
}, ...__VLS_functionalComponentArgsRest(__VLS_298));
__VLS_300.slots.default;
for (const [r] of __VLS_getVForSourceType((__VLS_ctx.roles))) {
    const __VLS_301 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_302 = __VLS_asFunctionalComponent(__VLS_301, new __VLS_301({
        key: (r.id),
        label: (r.name),
        value: (r.id),
    }));
    const __VLS_303 = __VLS_302({
        key: (r.id),
        label: (r.name),
        value: (r.id),
    }, ...__VLS_functionalComponentArgsRest(__VLS_302));
}
var __VLS_300;
var __VLS_296;
var __VLS_284;
{
    const { footer: __VLS_thisSlot } = __VLS_280.slots;
    const __VLS_305 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_306 = __VLS_asFunctionalComponent(__VLS_305, new __VLS_305({
        ...{ 'onClick': {} },
    }));
    const __VLS_307 = __VLS_306({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_306));
    let __VLS_309;
    let __VLS_310;
    let __VLS_311;
    const __VLS_312 = {
        onClick: (...[$event]) => {
            __VLS_ctx.roleVisible = false;
        }
    };
    __VLS_308.slots.default;
    (__VLS_ctx.t('common.cancel'));
    var __VLS_308;
    const __VLS_313 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_314 = __VLS_asFunctionalComponent(__VLS_313, new __VLS_313({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.roleAssignLoading),
    }));
    const __VLS_315 = __VLS_314({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.roleAssignLoading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_314));
    let __VLS_317;
    let __VLS_318;
    let __VLS_319;
    const __VLS_320 = {
        onClick: (__VLS_ctx.submitAssignRoles)
    };
    __VLS_316.slots.default;
    (__VLS_ctx.t('common.confirm'));
    var __VLS_316;
}
var __VLS_280;
const __VLS_321 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_322 = __VLS_asFunctionalComponent(__VLS_321, new __VLS_321({
    modelValue: (__VLS_ctx.g2Visible),
    title: (__VLS_ctx.t('common.google2faManage')),
    width: "680px",
}));
const __VLS_323 = __VLS_322({
    modelValue: (__VLS_ctx.g2Visible),
    title: (__VLS_ctx.t('common.google2faManage')),
    width: "680px",
}, ...__VLS_functionalComponentArgsRest(__VLS_322));
__VLS_324.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ style: {} },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ style: {} },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ style: {} },
});
(__VLS_ctx.t('common.user'));
(__VLS_ctx.g2User.username);
(__VLS_ctx.g2User.userId);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ style: {} },
});
const __VLS_325 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_326 = __VLS_asFunctionalComponent(__VLS_325, new __VLS_325({
    ...{ 'onClick': {} },
    loading: (__VLS_ctx.g2InitLoading),
}));
const __VLS_327 = __VLS_326({
    ...{ 'onClick': {} },
    loading: (__VLS_ctx.g2InitLoading),
}, ...__VLS_functionalComponentArgsRest(__VLS_326));
let __VLS_329;
let __VLS_330;
let __VLS_331;
const __VLS_332 = {
    onClick: (__VLS_ctx.doG2Init)
};
__VLS_asFunctionalDirective(__VLS_directives.vPerm)(null, { ...__VLS_directiveBindingRestFields, value: ('sys:user:2fa:init') }, null, null);
__VLS_328.slots.default;
(__VLS_ctx.t('perms.sys:user:2fa:init'));
var __VLS_328;
const __VLS_333 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_334 = __VLS_asFunctionalComponent(__VLS_333, new __VLS_333({
    ...{ 'onClick': {} },
    type: "success",
    loading: (__VLS_ctx.g2EnableLoading),
}));
const __VLS_335 = __VLS_334({
    ...{ 'onClick': {} },
    type: "success",
    loading: (__VLS_ctx.g2EnableLoading),
}, ...__VLS_functionalComponentArgsRest(__VLS_334));
let __VLS_337;
let __VLS_338;
let __VLS_339;
const __VLS_340 = {
    onClick: (__VLS_ctx.doG2Enable)
};
__VLS_asFunctionalDirective(__VLS_directives.vPerm)(null, { ...__VLS_directiveBindingRestFields, value: ('sys:user:2fa:enable') }, null, null);
__VLS_336.slots.default;
(__VLS_ctx.t('perms.sys:user:2fa:enable'));
var __VLS_336;
const __VLS_341 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_342 = __VLS_asFunctionalComponent(__VLS_341, new __VLS_341({
    ...{ 'onClick': {} },
    type: "warning",
    loading: (__VLS_ctx.g2DisableLoading),
}));
const __VLS_343 = __VLS_342({
    ...{ 'onClick': {} },
    type: "warning",
    loading: (__VLS_ctx.g2DisableLoading),
}, ...__VLS_functionalComponentArgsRest(__VLS_342));
let __VLS_345;
let __VLS_346;
let __VLS_347;
const __VLS_348 = {
    onClick: (__VLS_ctx.doG2Disable)
};
__VLS_asFunctionalDirective(__VLS_directives.vPerm)(null, { ...__VLS_directiveBindingRestFields, value: ('sys:user:2fa:disable') }, null, null);
__VLS_344.slots.default;
(__VLS_ctx.t('perms.sys:user:2fa:disable'));
var __VLS_344;
const __VLS_349 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_350 = __VLS_asFunctionalComponent(__VLS_349, new __VLS_349({
    ...{ 'onClick': {} },
    type: "danger",
}));
const __VLS_351 = __VLS_350({
    ...{ 'onClick': {} },
    type: "danger",
}, ...__VLS_functionalComponentArgsRest(__VLS_350));
let __VLS_353;
let __VLS_354;
let __VLS_355;
const __VLS_356 = {
    onClick: (__VLS_ctx.doG2Reset)
};
__VLS_asFunctionalDirective(__VLS_directives.vPerm)(null, { ...__VLS_directiveBindingRestFields, value: ('sys:user:2fa:reset') }, null, null);
__VLS_352.slots.default;
(__VLS_ctx.t('perms.sys:user:2fa:reset'));
var __VLS_352;
const __VLS_357 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_358 = __VLS_asFunctionalComponent(__VLS_357, new __VLS_357({
    labelWidth: "100px",
}));
const __VLS_359 = __VLS_358({
    labelWidth: "100px",
}, ...__VLS_functionalComponentArgsRest(__VLS_358));
__VLS_360.slots.default;
const __VLS_361 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_362 = __VLS_asFunctionalComponent(__VLS_361, new __VLS_361({
    label: (__VLS_ctx.t('common.code')),
}));
const __VLS_363 = __VLS_362({
    label: (__VLS_ctx.t('common.code')),
}, ...__VLS_functionalComponentArgsRest(__VLS_362));
__VLS_364.slots.default;
const __VLS_365 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_366 = __VLS_asFunctionalComponent(__VLS_365, new __VLS_365({
    modelValue: (__VLS_ctx.g2Form.code),
    placeholder: (__VLS_ctx.t('common.enterGoogleCode')),
}));
const __VLS_367 = __VLS_366({
    modelValue: (__VLS_ctx.g2Form.code),
    placeholder: (__VLS_ctx.t('common.enterGoogleCode')),
}, ...__VLS_functionalComponentArgsRest(__VLS_366));
var __VLS_364;
const __VLS_369 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_370 = __VLS_asFunctionalComponent(__VLS_369, new __VLS_369({
    label: "secret",
}));
const __VLS_371 = __VLS_370({
    label: "secret",
}, ...__VLS_functionalComponentArgsRest(__VLS_370));
__VLS_372.slots.default;
const __VLS_373 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_374 = __VLS_asFunctionalComponent(__VLS_373, new __VLS_373({
    modelValue: (__VLS_ctx.g2Init.secret),
    readonly: true,
}));
const __VLS_375 = __VLS_374({
    modelValue: (__VLS_ctx.g2Init.secret),
    readonly: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_374));
var __VLS_372;
const __VLS_377 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_378 = __VLS_asFunctionalComponent(__VLS_377, new __VLS_377({
    label: "otpauthUrl",
}));
const __VLS_379 = __VLS_378({
    label: "otpauthUrl",
}, ...__VLS_functionalComponentArgsRest(__VLS_378));
__VLS_380.slots.default;
const __VLS_381 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_382 = __VLS_asFunctionalComponent(__VLS_381, new __VLS_381({
    modelValue: (__VLS_ctx.g2Init.otpauthUrl),
    readonly: true,
}));
const __VLS_383 = __VLS_382({
    modelValue: (__VLS_ctx.g2Init.otpauthUrl),
    readonly: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_382));
var __VLS_380;
var __VLS_360;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ style: {} },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ style: {} },
});
(__VLS_ctx.t('common.qrCode'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ style: {} },
});
if (__VLS_ctx.g2Init.qrCode) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.img)({
        src: (__VLS_ctx.g2Init.qrCode),
        ...{ style: {} },
    });
}
else {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    (__VLS_ctx.t('common.click2faBindGenerateQrCode'));
}
{
    const { footer: __VLS_thisSlot } = __VLS_324.slots;
    const __VLS_385 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_386 = __VLS_asFunctionalComponent(__VLS_385, new __VLS_385({
        ...{ 'onClick': {} },
    }));
    const __VLS_387 = __VLS_386({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_386));
    let __VLS_389;
    let __VLS_390;
    let __VLS_391;
    const __VLS_392 = {
        onClick: (...[$event]) => {
            __VLS_ctx.g2Visible = false;
        }
    };
    __VLS_388.slots.default;
    (__VLS_ctx.t('common.cancel'));
    var __VLS_388;
}
var __VLS_324;
/** @type {__VLS_StyleScopedClasses['el-icon--right']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            ArrowDown: ArrowDown,
            t: t,
            pagination: pagination,
            list: list,
            loading: loading,
            queryForm: queryForm,
            statusOptions: statusOptions,
            fetchList: fetchList,
            onSearch: onSearch,
            onReset: onReset,
            roleLoading: roleLoading,
            roles: roles,
            editVisible: editVisible,
            editMode: editMode,
            editForm: editForm,
            editFormLoading: editFormLoading,
            openCreate: openCreate,
            openEdit: openEdit,
            submitEdit: submitEdit,
            onDelete: onDelete,
            onToggleStatus: onToggleStatus,
            pwdVisible: pwdVisible,
            pwdForm: pwdForm,
            pwdSubmitLoading: pwdSubmitLoading,
            openResetPwd: openResetPwd,
            submitResetPwd: submitResetPwd,
            roleVisible: roleVisible,
            roleForm: roleForm,
            roleAssignLoading: roleAssignLoading,
            openAssignRoles: openAssignRoles,
            submitAssignRoles: submitAssignRoles,
            g2Visible: g2Visible,
            g2User: g2User,
            g2Init: g2Init,
            g2Form: g2Form,
            g2InitLoading: g2InitLoading,
            g2EnableLoading: g2EnableLoading,
            g2DisableLoading: g2DisableLoading,
            openGoogle2fa: openGoogle2fa,
            doG2Init: doG2Init,
            doG2Enable: doG2Enable,
            doG2Disable: doG2Disable,
            doG2Reset: doG2Reset,
            roleNameMap: roleNameMap,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
