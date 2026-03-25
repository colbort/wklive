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
const { pagination, updatePagination, nextPage: paginationNextPage, prevPage: paginationPrevPage, } = usePagination(10);
const list = ref([]);
const { loading, withLoading: withMainLoading } = useLoading();
// Query form
const { form: queryForm } = useForm({
    initialData: {
        keyword: '',
        status: undefined,
    },
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
                cursor: pagination.cursor,
                limit: pagination.limit,
            });
            // 兼容 code=0 / 200
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
function onSearch() {
    pagination.cursor = null;
    pagination.hasPrev = false;
    fetchList();
}
function onReset() {
    queryForm.keyword = '';
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
// ---------- 角色缓存（分配角色用） ----------
const { loading: roleLoading, withLoading: withRoleLoading } = useLoading();
const roles = ref([]);
async function fetchRoles() {
    await withRoleLoading(async () => {
        try {
            const res = await roleService.getList({ cursor: null, limit: 9999, status: 1 });
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
    },
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
    initialData: { id: 0, username: '', password: '' },
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
    initialData: { userId: 0, username: '', roleIds: [] },
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
    initialData: { userId: 0, username: '' },
});
const { form: g2Init } = useForm({
    initialData: { secret: '', otpauthUrl: '', qrCode: '' },
});
const { form: g2Form } = useForm({
    initialData: { code: '' },
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
            if (res.code !== 0 && res.code !== 200)
                throw new Error(res.msg || 'init failed');
            g2Init.secret = res.data?.secret || '';
            g2Init.otpauthUrl = res.data?.otpauthUrl || '';
            g2Init.qrCode = res.data?.qrCode || '';
            console.log('QR Code data:', g2Init.qrCode); // 调试信息
            ElMessage.success(t('common.success'));
        }
        catch (e) {
            ElMessage.error(e?.message || t('common.failed'));
        }
    });
}
async function doG2Bind() {
    try {
        if (!g2Form.code) {
            ElMessage.warning(t('common.pleaseInputCode'));
            return;
        }
        const res = await userService.bindGoogle2FA(g2User.userId, g2Init.secret, g2Form.code);
        if (res.code !== 0 && res.code !== 200)
            throw new Error(res.msg || 'bind failed');
        ElMessage.success(t('common.success'));
        fetchList();
    }
    catch (e) {
        ElMessage.error(e?.message || t('common.failed'));
    }
}
async function copySecret() {
    if (!g2Init.secret) {
        ElMessage.warning(t('common.noData'));
        return;
    }
    try {
        await navigator.clipboard.writeText(g2Init.secret);
        ElMessage.success(t('common.copied'));
    }
    catch (e) {
        ElMessage.error(t('common.copyFailed'));
    }
}
async function copyOtpauthUrl() {
    if (!g2Init.otpauthUrl) {
        ElMessage.warning(t('common.noData'));
        return;
    }
    try {
        await navigator.clipboard.writeText(g2Init.otpauthUrl);
        ElMessage.success(t('common.copied'));
    }
    catch (e) {
        ElMessage.error(t('common.copyFailed'));
    }
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
    for (const [rid] of __VLS_getVForSourceType((row.roleIds || []))) {
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
        (__VLS_ctx.roleNameMap.get(rid) || '#' + rid);
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
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
(__VLS_ctx.t('common.totalItems', { count: __VLS_ctx.pagination.total }));
const __VLS_160 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_161 = __VLS_asFunctionalComponent(__VLS_160, new __VLS_160({
    ...{ 'onClick': {} },
    disabled: (!__VLS_ctx.pagination.hasPrev),
}));
const __VLS_162 = __VLS_161({
    ...{ 'onClick': {} },
    disabled: (!__VLS_ctx.pagination.hasPrev),
}, ...__VLS_functionalComponentArgsRest(__VLS_161));
let __VLS_164;
let __VLS_165;
let __VLS_166;
const __VLS_167 = {
    onClick: (__VLS_ctx.prevPage)
};
__VLS_163.slots.default;
(__VLS_ctx.t('common.prevPage'));
var __VLS_163;
const __VLS_168 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_169 = __VLS_asFunctionalComponent(__VLS_168, new __VLS_168({
    ...{ 'onClick': {} },
    disabled: (!__VLS_ctx.pagination.hasNext),
}));
const __VLS_170 = __VLS_169({
    ...{ 'onClick': {} },
    disabled: (!__VLS_ctx.pagination.hasNext),
}, ...__VLS_functionalComponentArgsRest(__VLS_169));
let __VLS_172;
let __VLS_173;
let __VLS_174;
const __VLS_175 = {
    onClick: (__VLS_ctx.nextPage)
};
__VLS_171.slots.default;
(__VLS_ctx.t('common.nextPage'));
var __VLS_171;
const __VLS_176 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_177 = __VLS_asFunctionalComponent(__VLS_176, new __VLS_176({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.pagination.limit),
    ...{ style: {} },
}));
const __VLS_178 = __VLS_177({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.pagination.limit),
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_177));
let __VLS_180;
let __VLS_181;
let __VLS_182;
const __VLS_183 = {
    onChange: (() => {
        __VLS_ctx.pagination.cursor = null;
        __VLS_ctx.pagination.hasPrev = false;
        __VLS_ctx.fetchList();
    })
};
__VLS_179.slots.default;
const __VLS_184 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_185 = __VLS_asFunctionalComponent(__VLS_184, new __VLS_184({
    label: "10",
    value: (10),
}));
const __VLS_186 = __VLS_185({
    label: "10",
    value: (10),
}, ...__VLS_functionalComponentArgsRest(__VLS_185));
const __VLS_188 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_189 = __VLS_asFunctionalComponent(__VLS_188, new __VLS_188({
    label: "20",
    value: (20),
}));
const __VLS_190 = __VLS_189({
    label: "20",
    value: (20),
}, ...__VLS_functionalComponentArgsRest(__VLS_189));
const __VLS_192 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_193 = __VLS_asFunctionalComponent(__VLS_192, new __VLS_192({
    label: "50",
    value: (50),
}));
const __VLS_194 = __VLS_193({
    label: "50",
    value: (50),
}, ...__VLS_functionalComponentArgsRest(__VLS_193));
var __VLS_179;
var __VLS_3;
const __VLS_196 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_197 = __VLS_asFunctionalComponent(__VLS_196, new __VLS_196({
    modelValue: (__VLS_ctx.editVisible),
    title: (__VLS_ctx.editMode === 'create' ? __VLS_ctx.t('common.addUser') : __VLS_ctx.t('common.editUser')),
    width: "520px",
}));
const __VLS_198 = __VLS_197({
    modelValue: (__VLS_ctx.editVisible),
    title: (__VLS_ctx.editMode === 'create' ? __VLS_ctx.t('common.addUser') : __VLS_ctx.t('common.editUser')),
    width: "520px",
}, ...__VLS_functionalComponentArgsRest(__VLS_197));
__VLS_199.slots.default;
const __VLS_200 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_201 = __VLS_asFunctionalComponent(__VLS_200, new __VLS_200({
    labelWidth: "90px",
}));
const __VLS_202 = __VLS_201({
    labelWidth: "90px",
}, ...__VLS_functionalComponentArgsRest(__VLS_201));
__VLS_203.slots.default;
if (__VLS_ctx.editMode === 'create') {
    const __VLS_204 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_205 = __VLS_asFunctionalComponent(__VLS_204, new __VLS_204({
        label: (__VLS_ctx.t('common.username')),
    }));
    const __VLS_206 = __VLS_205({
        label: (__VLS_ctx.t('common.username')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_205));
    __VLS_207.slots.default;
    const __VLS_208 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_209 = __VLS_asFunctionalComponent(__VLS_208, new __VLS_208({
        modelValue: (__VLS_ctx.editForm.username),
    }));
    const __VLS_210 = __VLS_209({
        modelValue: (__VLS_ctx.editForm.username),
    }, ...__VLS_functionalComponentArgsRest(__VLS_209));
    var __VLS_207;
}
if (__VLS_ctx.editMode === 'create') {
    const __VLS_212 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_213 = __VLS_asFunctionalComponent(__VLS_212, new __VLS_212({
        label: (__VLS_ctx.t('common.password')),
    }));
    const __VLS_214 = __VLS_213({
        label: (__VLS_ctx.t('common.password')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_213));
    __VLS_215.slots.default;
    const __VLS_216 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_217 = __VLS_asFunctionalComponent(__VLS_216, new __VLS_216({
        modelValue: (__VLS_ctx.editForm.password),
        type: "password",
        showPassword: true,
    }));
    const __VLS_218 = __VLS_217({
        modelValue: (__VLS_ctx.editForm.password),
        type: "password",
        showPassword: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_217));
    var __VLS_215;
}
const __VLS_220 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_221 = __VLS_asFunctionalComponent(__VLS_220, new __VLS_220({
    label: (__VLS_ctx.t('common.nickname')),
}));
const __VLS_222 = __VLS_221({
    label: (__VLS_ctx.t('common.nickname')),
}, ...__VLS_functionalComponentArgsRest(__VLS_221));
__VLS_223.slots.default;
const __VLS_224 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_225 = __VLS_asFunctionalComponent(__VLS_224, new __VLS_224({
    modelValue: (__VLS_ctx.editForm.nickname),
}));
const __VLS_226 = __VLS_225({
    modelValue: (__VLS_ctx.editForm.nickname),
}, ...__VLS_functionalComponentArgsRest(__VLS_225));
var __VLS_223;
const __VLS_228 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_229 = __VLS_asFunctionalComponent(__VLS_228, new __VLS_228({
    label: (__VLS_ctx.t('common.status')),
}));
const __VLS_230 = __VLS_229({
    label: (__VLS_ctx.t('common.status')),
}, ...__VLS_functionalComponentArgsRest(__VLS_229));
__VLS_231.slots.default;
const __VLS_232 = {}.ElSwitch;
/** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
// @ts-ignore
const __VLS_233 = __VLS_asFunctionalComponent(__VLS_232, new __VLS_232({
    modelValue: (__VLS_ctx.editForm.status),
    activeValue: (1),
    inactiveValue: (0),
}));
const __VLS_234 = __VLS_233({
    modelValue: (__VLS_ctx.editForm.status),
    activeValue: (1),
    inactiveValue: (0),
}, ...__VLS_functionalComponentArgsRest(__VLS_233));
var __VLS_231;
const __VLS_236 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_237 = __VLS_asFunctionalComponent(__VLS_236, new __VLS_236({
    label: (__VLS_ctx.t('common.role')),
}));
const __VLS_238 = __VLS_237({
    label: (__VLS_ctx.t('common.role')),
}, ...__VLS_functionalComponentArgsRest(__VLS_237));
__VLS_239.slots.default;
const __VLS_240 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_241 = __VLS_asFunctionalComponent(__VLS_240, new __VLS_240({
    modelValue: (__VLS_ctx.editForm.roleIds),
    multiple: true,
    filterable: true,
    ...{ style: {} },
    loading: (__VLS_ctx.roleLoading),
}));
const __VLS_242 = __VLS_241({
    modelValue: (__VLS_ctx.editForm.roleIds),
    multiple: true,
    filterable: true,
    ...{ style: {} },
    loading: (__VLS_ctx.roleLoading),
}, ...__VLS_functionalComponentArgsRest(__VLS_241));
__VLS_243.slots.default;
for (const [r] of __VLS_getVForSourceType((__VLS_ctx.roles))) {
    const __VLS_244 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_245 = __VLS_asFunctionalComponent(__VLS_244, new __VLS_244({
        key: (r.id),
        label: (r.name),
        value: (r.id),
    }));
    const __VLS_246 = __VLS_245({
        key: (r.id),
        label: (r.name),
        value: (r.id),
    }, ...__VLS_functionalComponentArgsRest(__VLS_245));
}
var __VLS_243;
var __VLS_239;
var __VLS_203;
{
    const { footer: __VLS_thisSlot } = __VLS_199.slots;
    const __VLS_248 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_249 = __VLS_asFunctionalComponent(__VLS_248, new __VLS_248({
        ...{ 'onClick': {} },
    }));
    const __VLS_250 = __VLS_249({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_249));
    let __VLS_252;
    let __VLS_253;
    let __VLS_254;
    const __VLS_255 = {
        onClick: (...[$event]) => {
            __VLS_ctx.editVisible = false;
        }
    };
    __VLS_251.slots.default;
    (__VLS_ctx.t('common.cancel'));
    var __VLS_251;
    const __VLS_256 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_257 = __VLS_asFunctionalComponent(__VLS_256, new __VLS_256({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.editFormLoading),
    }));
    const __VLS_258 = __VLS_257({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.editFormLoading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_257));
    let __VLS_260;
    let __VLS_261;
    let __VLS_262;
    const __VLS_263 = {
        onClick: (__VLS_ctx.submitEdit)
    };
    __VLS_259.slots.default;
    (__VLS_ctx.t('common.confirm'));
    var __VLS_259;
}
var __VLS_199;
const __VLS_264 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_265 = __VLS_asFunctionalComponent(__VLS_264, new __VLS_264({
    modelValue: (__VLS_ctx.pwdVisible),
    title: (__VLS_ctx.t('common.resetPassword')),
    width: "420px",
}));
const __VLS_266 = __VLS_265({
    modelValue: (__VLS_ctx.pwdVisible),
    title: (__VLS_ctx.t('common.resetPassword')),
    width: "420px",
}, ...__VLS_functionalComponentArgsRest(__VLS_265));
__VLS_267.slots.default;
const __VLS_268 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_269 = __VLS_asFunctionalComponent(__VLS_268, new __VLS_268({
    labelWidth: "90px",
}));
const __VLS_270 = __VLS_269({
    labelWidth: "90px",
}, ...__VLS_functionalComponentArgsRest(__VLS_269));
__VLS_271.slots.default;
const __VLS_272 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_273 = __VLS_asFunctionalComponent(__VLS_272, new __VLS_272({
    label: (__VLS_ctx.t('common.username')),
}));
const __VLS_274 = __VLS_273({
    label: (__VLS_ctx.t('common.username')),
}, ...__VLS_functionalComponentArgsRest(__VLS_273));
__VLS_275.slots.default;
const __VLS_276 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_277 = __VLS_asFunctionalComponent(__VLS_276, new __VLS_276({
    modelValue: (__VLS_ctx.pwdForm.username),
    disabled: true,
}));
const __VLS_278 = __VLS_277({
    modelValue: (__VLS_ctx.pwdForm.username),
    disabled: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_277));
var __VLS_275;
const __VLS_280 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_281 = __VLS_asFunctionalComponent(__VLS_280, new __VLS_280({
    label: (__VLS_ctx.t('common.newPassword')),
}));
const __VLS_282 = __VLS_281({
    label: (__VLS_ctx.t('common.newPassword')),
}, ...__VLS_functionalComponentArgsRest(__VLS_281));
__VLS_283.slots.default;
const __VLS_284 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_285 = __VLS_asFunctionalComponent(__VLS_284, new __VLS_284({
    modelValue: (__VLS_ctx.pwdForm.password),
    type: "password",
    showPassword: true,
}));
const __VLS_286 = __VLS_285({
    modelValue: (__VLS_ctx.pwdForm.password),
    type: "password",
    showPassword: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_285));
var __VLS_283;
var __VLS_271;
{
    const { footer: __VLS_thisSlot } = __VLS_267.slots;
    const __VLS_288 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_289 = __VLS_asFunctionalComponent(__VLS_288, new __VLS_288({
        ...{ 'onClick': {} },
    }));
    const __VLS_290 = __VLS_289({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_289));
    let __VLS_292;
    let __VLS_293;
    let __VLS_294;
    const __VLS_295 = {
        onClick: (...[$event]) => {
            __VLS_ctx.pwdVisible = false;
        }
    };
    __VLS_291.slots.default;
    (__VLS_ctx.t('common.cancel'));
    var __VLS_291;
    const __VLS_296 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_297 = __VLS_asFunctionalComponent(__VLS_296, new __VLS_296({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.pwdSubmitLoading),
    }));
    const __VLS_298 = __VLS_297({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.pwdSubmitLoading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_297));
    let __VLS_300;
    let __VLS_301;
    let __VLS_302;
    const __VLS_303 = {
        onClick: (__VLS_ctx.submitResetPwd)
    };
    __VLS_299.slots.default;
    (__VLS_ctx.t('common.confirm'));
    var __VLS_299;
}
var __VLS_267;
const __VLS_304 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_305 = __VLS_asFunctionalComponent(__VLS_304, new __VLS_304({
    modelValue: (__VLS_ctx.roleVisible),
    title: (__VLS_ctx.t('common.assignRoles')),
    width: "520px",
}));
const __VLS_306 = __VLS_305({
    modelValue: (__VLS_ctx.roleVisible),
    title: (__VLS_ctx.t('common.assignRoles')),
    width: "520px",
}, ...__VLS_functionalComponentArgsRest(__VLS_305));
__VLS_307.slots.default;
const __VLS_308 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_309 = __VLS_asFunctionalComponent(__VLS_308, new __VLS_308({
    labelWidth: "90px",
}));
const __VLS_310 = __VLS_309({
    labelWidth: "90px",
}, ...__VLS_functionalComponentArgsRest(__VLS_309));
__VLS_311.slots.default;
const __VLS_312 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_313 = __VLS_asFunctionalComponent(__VLS_312, new __VLS_312({
    label: (__VLS_ctx.t('common.username')),
}));
const __VLS_314 = __VLS_313({
    label: (__VLS_ctx.t('common.username')),
}, ...__VLS_functionalComponentArgsRest(__VLS_313));
__VLS_315.slots.default;
const __VLS_316 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_317 = __VLS_asFunctionalComponent(__VLS_316, new __VLS_316({
    modelValue: (__VLS_ctx.roleForm.username),
    disabled: true,
}));
const __VLS_318 = __VLS_317({
    modelValue: (__VLS_ctx.roleForm.username),
    disabled: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_317));
var __VLS_315;
const __VLS_320 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_321 = __VLS_asFunctionalComponent(__VLS_320, new __VLS_320({
    label: (__VLS_ctx.t('common.role')),
}));
const __VLS_322 = __VLS_321({
    label: (__VLS_ctx.t('common.role')),
}, ...__VLS_functionalComponentArgsRest(__VLS_321));
__VLS_323.slots.default;
const __VLS_324 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_325 = __VLS_asFunctionalComponent(__VLS_324, new __VLS_324({
    modelValue: (__VLS_ctx.roleForm.roleIds),
    multiple: true,
    filterable: true,
    ...{ style: {} },
    loading: (__VLS_ctx.roleLoading),
}));
const __VLS_326 = __VLS_325({
    modelValue: (__VLS_ctx.roleForm.roleIds),
    multiple: true,
    filterable: true,
    ...{ style: {} },
    loading: (__VLS_ctx.roleLoading),
}, ...__VLS_functionalComponentArgsRest(__VLS_325));
__VLS_327.slots.default;
for (const [r] of __VLS_getVForSourceType((__VLS_ctx.roles))) {
    const __VLS_328 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_329 = __VLS_asFunctionalComponent(__VLS_328, new __VLS_328({
        key: (r.id),
        label: (r.name),
        value: (r.id),
    }));
    const __VLS_330 = __VLS_329({
        key: (r.id),
        label: (r.name),
        value: (r.id),
    }, ...__VLS_functionalComponentArgsRest(__VLS_329));
}
var __VLS_327;
var __VLS_323;
var __VLS_311;
{
    const { footer: __VLS_thisSlot } = __VLS_307.slots;
    const __VLS_332 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_333 = __VLS_asFunctionalComponent(__VLS_332, new __VLS_332({
        ...{ 'onClick': {} },
    }));
    const __VLS_334 = __VLS_333({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_333));
    let __VLS_336;
    let __VLS_337;
    let __VLS_338;
    const __VLS_339 = {
        onClick: (...[$event]) => {
            __VLS_ctx.roleVisible = false;
        }
    };
    __VLS_335.slots.default;
    (__VLS_ctx.t('common.cancel'));
    var __VLS_335;
    const __VLS_340 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_341 = __VLS_asFunctionalComponent(__VLS_340, new __VLS_340({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.roleAssignLoading),
    }));
    const __VLS_342 = __VLS_341({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.roleAssignLoading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_341));
    let __VLS_344;
    let __VLS_345;
    let __VLS_346;
    const __VLS_347 = {
        onClick: (__VLS_ctx.submitAssignRoles)
    };
    __VLS_343.slots.default;
    (__VLS_ctx.t('common.confirm'));
    var __VLS_343;
}
var __VLS_307;
const __VLS_348 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_349 = __VLS_asFunctionalComponent(__VLS_348, new __VLS_348({
    modelValue: (__VLS_ctx.g2Visible),
    title: (__VLS_ctx.t('common.google2faManage')),
    width: "680px",
}));
const __VLS_350 = __VLS_349({
    modelValue: (__VLS_ctx.g2Visible),
    title: (__VLS_ctx.t('common.google2faManage')),
    width: "680px",
}, ...__VLS_functionalComponentArgsRest(__VLS_349));
__VLS_351.slots.default;
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
const __VLS_352 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_353 = __VLS_asFunctionalComponent(__VLS_352, new __VLS_352({
    ...{ 'onClick': {} },
    loading: (__VLS_ctx.g2InitLoading),
}));
const __VLS_354 = __VLS_353({
    ...{ 'onClick': {} },
    loading: (__VLS_ctx.g2InitLoading),
}, ...__VLS_functionalComponentArgsRest(__VLS_353));
let __VLS_356;
let __VLS_357;
let __VLS_358;
const __VLS_359 = {
    onClick: (__VLS_ctx.doG2Init)
};
__VLS_asFunctionalDirective(__VLS_directives.vPerm)(null, { ...__VLS_directiveBindingRestFields, value: ('sys:user:2fa:init') }, null, null);
__VLS_355.slots.default;
(__VLS_ctx.t('perms.sys:user:2fa:init'));
var __VLS_355;
const __VLS_360 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_361 = __VLS_asFunctionalComponent(__VLS_360, new __VLS_360({
    ...{ 'onClick': {} },
    type: "success",
    loading: (__VLS_ctx.g2EnableLoading),
}));
const __VLS_362 = __VLS_361({
    ...{ 'onClick': {} },
    type: "success",
    loading: (__VLS_ctx.g2EnableLoading),
}, ...__VLS_functionalComponentArgsRest(__VLS_361));
let __VLS_364;
let __VLS_365;
let __VLS_366;
const __VLS_367 = {
    onClick: (__VLS_ctx.doG2Enable)
};
__VLS_asFunctionalDirective(__VLS_directives.vPerm)(null, { ...__VLS_directiveBindingRestFields, value: ('sys:user:2fa:enable') }, null, null);
__VLS_363.slots.default;
(__VLS_ctx.t('perms.sys:user:2fa:enable'));
var __VLS_363;
const __VLS_368 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_369 = __VLS_asFunctionalComponent(__VLS_368, new __VLS_368({
    ...{ 'onClick': {} },
    type: "warning",
    loading: (__VLS_ctx.g2DisableLoading),
}));
const __VLS_370 = __VLS_369({
    ...{ 'onClick': {} },
    type: "warning",
    loading: (__VLS_ctx.g2DisableLoading),
}, ...__VLS_functionalComponentArgsRest(__VLS_369));
let __VLS_372;
let __VLS_373;
let __VLS_374;
const __VLS_375 = {
    onClick: (__VLS_ctx.doG2Disable)
};
__VLS_asFunctionalDirective(__VLS_directives.vPerm)(null, { ...__VLS_directiveBindingRestFields, value: ('sys:user:2fa:disable') }, null, null);
__VLS_371.slots.default;
(__VLS_ctx.t('perms.sys:user:2fa:disable'));
var __VLS_371;
const __VLS_376 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_377 = __VLS_asFunctionalComponent(__VLS_376, new __VLS_376({
    ...{ 'onClick': {} },
    type: "danger",
}));
const __VLS_378 = __VLS_377({
    ...{ 'onClick': {} },
    type: "danger",
}, ...__VLS_functionalComponentArgsRest(__VLS_377));
let __VLS_380;
let __VLS_381;
let __VLS_382;
const __VLS_383 = {
    onClick: (__VLS_ctx.doG2Reset)
};
__VLS_asFunctionalDirective(__VLS_directives.vPerm)(null, { ...__VLS_directiveBindingRestFields, value: ('sys:user:2fa:reset') }, null, null);
__VLS_379.slots.default;
(__VLS_ctx.t('perms.sys:user:2fa:reset'));
var __VLS_379;
const __VLS_384 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_385 = __VLS_asFunctionalComponent(__VLS_384, new __VLS_384({
    labelWidth: "100px",
}));
const __VLS_386 = __VLS_385({
    labelWidth: "100px",
}, ...__VLS_functionalComponentArgsRest(__VLS_385));
__VLS_387.slots.default;
const __VLS_388 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_389 = __VLS_asFunctionalComponent(__VLS_388, new __VLS_388({
    label: (__VLS_ctx.t('common.code')),
}));
const __VLS_390 = __VLS_389({
    label: (__VLS_ctx.t('common.code')),
}, ...__VLS_functionalComponentArgsRest(__VLS_389));
__VLS_391.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ style: {} },
});
const __VLS_392 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_393 = __VLS_asFunctionalComponent(__VLS_392, new __VLS_392({
    modelValue: (__VLS_ctx.g2Form.code),
    placeholder: (__VLS_ctx.t('common.enterGoogleCode')),
    ...{ style: {} },
}));
const __VLS_394 = __VLS_393({
    modelValue: (__VLS_ctx.g2Form.code),
    placeholder: (__VLS_ctx.t('common.enterGoogleCode')),
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_393));
const __VLS_396 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_397 = __VLS_asFunctionalComponent(__VLS_396, new __VLS_396({
    ...{ 'onClick': {} },
}));
const __VLS_398 = __VLS_397({
    ...{ 'onClick': {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_397));
let __VLS_400;
let __VLS_401;
let __VLS_402;
const __VLS_403 = {
    onClick: (__VLS_ctx.doG2Bind)
};
__VLS_399.slots.default;
(__VLS_ctx.t('perms.sys:user:2fa:bind'));
var __VLS_399;
var __VLS_391;
const __VLS_404 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_405 = __VLS_asFunctionalComponent(__VLS_404, new __VLS_404({
    label: (__VLS_ctx.t('common.secret')),
}));
const __VLS_406 = __VLS_405({
    label: (__VLS_ctx.t('common.secret')),
}, ...__VLS_functionalComponentArgsRest(__VLS_405));
__VLS_407.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ style: {} },
});
const __VLS_408 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_409 = __VLS_asFunctionalComponent(__VLS_408, new __VLS_408({
    modelValue: (__VLS_ctx.g2Init.secret),
    readonly: true,
    ...{ style: {} },
}));
const __VLS_410 = __VLS_409({
    modelValue: (__VLS_ctx.g2Init.secret),
    readonly: true,
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_409));
const __VLS_412 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_413 = __VLS_asFunctionalComponent(__VLS_412, new __VLS_412({
    ...{ 'onClick': {} },
    disabled: (!__VLS_ctx.g2Init.secret),
}));
const __VLS_414 = __VLS_413({
    ...{ 'onClick': {} },
    disabled: (!__VLS_ctx.g2Init.secret),
}, ...__VLS_functionalComponentArgsRest(__VLS_413));
let __VLS_416;
let __VLS_417;
let __VLS_418;
const __VLS_419 = {
    onClick: (__VLS_ctx.copySecret)
};
__VLS_415.slots.default;
(__VLS_ctx.t('common.copy'));
var __VLS_415;
var __VLS_407;
const __VLS_420 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_421 = __VLS_asFunctionalComponent(__VLS_420, new __VLS_420({
    label: (__VLS_ctx.t('common.otpauthUrl')),
}));
const __VLS_422 = __VLS_421({
    label: (__VLS_ctx.t('common.otpauthUrl')),
}, ...__VLS_functionalComponentArgsRest(__VLS_421));
__VLS_423.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ style: {} },
});
const __VLS_424 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_425 = __VLS_asFunctionalComponent(__VLS_424, new __VLS_424({
    modelValue: (__VLS_ctx.g2Init.otpauthUrl),
    readonly: true,
    ...{ style: {} },
}));
const __VLS_426 = __VLS_425({
    modelValue: (__VLS_ctx.g2Init.otpauthUrl),
    readonly: true,
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_425));
const __VLS_428 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_429 = __VLS_asFunctionalComponent(__VLS_428, new __VLS_428({
    ...{ 'onClick': {} },
    disabled: (!__VLS_ctx.g2Init.otpauthUrl),
}));
const __VLS_430 = __VLS_429({
    ...{ 'onClick': {} },
    disabled: (!__VLS_ctx.g2Init.otpauthUrl),
}, ...__VLS_functionalComponentArgsRest(__VLS_429));
let __VLS_432;
let __VLS_433;
let __VLS_434;
const __VLS_435 = {
    onClick: (__VLS_ctx.copyOtpauthUrl)
};
__VLS_431.slots.default;
(__VLS_ctx.t('common.copy'));
var __VLS_431;
var __VLS_423;
var __VLS_387;
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
if (__VLS_ctx.g2Init.qrCode) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    (__VLS_ctx.t('common.scanQrCodeWithGoogleAuthenticator'));
}
{
    const { footer: __VLS_thisSlot } = __VLS_351.slots;
    const __VLS_436 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_437 = __VLS_asFunctionalComponent(__VLS_436, new __VLS_436({
        ...{ 'onClick': {} },
    }));
    const __VLS_438 = __VLS_437({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_437));
    let __VLS_440;
    let __VLS_441;
    let __VLS_442;
    const __VLS_443 = {
        onClick: (...[$event]) => {
            __VLS_ctx.g2Visible = false;
        }
    };
    __VLS_439.slots.default;
    (__VLS_ctx.t('common.cancel'));
    var __VLS_439;
}
var __VLS_351;
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
            nextPage: nextPage,
            prevPage: prevPage,
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
            doG2Bind: doG2Bind,
            copySecret: copySecret,
            copyOtpauthUrl: copyOtpauthUrl,
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
