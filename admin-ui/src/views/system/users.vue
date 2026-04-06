<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { userService, roleService } from '@/services'
import { ArrowDown } from '@element-plus/icons-vue'
import type { SysUserItem, CreateUserRequest, UpdateUserRequest } from '@/services'
import { SysRole } from '@/services/system/RoleService'
import { usePagination } from '@/composables/usePagination'
import { useLoading } from '@/composables/useLoading'
import { useForm } from '@/composables/useForm'
import { useConfirm } from '@/composables/useConfirm'
import { formatDate } from '@/utils'

const { t } = useI18n()

// Pagination and main list
const {
  pagination,
  updatePagination,
  nextPage: paginationNextPage,
  prevPage: paginationPrevPage,
} = usePagination(10)
const list = ref<SysUserItem[]>([])
const { loading, withLoading: withMainLoading } = useLoading()

// Query form
const { form: queryForm } = useForm({
  initialData: {
    keyword: '',
    status: undefined as number | undefined,
  },
})

const statusOptions = [
  { label: t('common.enabled'), value: 1 },
  { label: t('common.disabled'), value: 2 },
]

async function fetchList() {
  await withMainLoading(async () => {
    try {
      const res = await userService.getList({
        keyword: queryForm.keyword || undefined,
        status: queryForm.status,
        cursor: pagination.cursor,
        limit: pagination.limit,
      })
      // 兼容 code=0 / 200
      if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || 'list failed')
      list.value = res.data || []
      updatePagination(
        res.total || 0,
        res.hasNext || false,
        res.hasPrev || false,
        res.nextCursor || null,
        res.prevCursor || null,
      )
    } catch (e: any) {
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
  queryForm.keyword = ''
  queryForm.status = undefined
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

// ---------- 角色缓存（分配角色用） ----------
const { loading: roleLoading, withLoading: withRoleLoading } = useLoading()
const roles = ref<SysRole[]>([])
async function fetchRoles() {
  await withRoleLoading(async () => {
    try {
      const res = await roleService.getList({ cursor: null, limit: 9999, status: 1 })
      if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || 'role list failed')
      roles.value = res.data || []
    } catch (e: any) {
      ElMessage.error(e?.message || t('common.loadFailed'))
    }
  })
}

// ---------- 新增/编辑 ----------
const editVisible = ref(false)
const editMode = ref<'create' | 'update'>('create')
const { form: editForm } = useForm({
  initialData: {
    id: 0,
    username: '',
    password: '',
    nickname: '',
    status: 1,
    roleIds: [] as number[],
  },
})
const { loading: editFormLoading, withLoading: withEditLoading } = useLoading()

function openCreate() {
  editMode.value = 'create'
  editForm.id = 0
  editForm.username = ''
  editForm.password = ''
  editForm.nickname = ''
  editForm.status = 1
  editForm.roleIds = []
  editVisible.value = true
}

function openEdit(row: SysUserItem) {
  editMode.value = 'update'
  editForm.id = row.id
  editForm.username = row.username
  editForm.password = ''
  editForm.nickname = row.nickname || ''
  editForm.status = row.status
  editForm.roleIds = (row.roleIds || []).slice()
  editVisible.value = true
}

async function submitEdit() {
  await withEditLoading(async () => {
    try {
      if (editMode.value === 'create') {
        if (!editForm.username || !editForm.password) {
          ElMessage.warning(t('common.pleaseInputAccountAndPassword'))
          return
        }
        const res = await userService.create({
          username: editForm.username,
          password: editForm.password,
          nickname: editForm.nickname || undefined,
          status: editForm.status,
          roleIds: editForm.roleIds,
        })
        if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || 'create failed')
        ElMessage.success(t('common.success'))
      } else {
        const res = await userService.update(editForm.id, {
          nickname: editForm.nickname || undefined,
          status: editForm.status,
          roleIds: editForm.roleIds,
        })
        if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || 'update failed')
        ElMessage.success(t('common.success'))
      }
      editVisible.value = false
      fetchList()
    } catch (e: any) {
      ElMessage.error(e?.message || t('common.failed'))
    }
  })
}

// ---------- 删除 ----------
const { confirm } = useConfirm()

async function onDelete(row: SysUserItem) {
  try {
    await confirm(t('common.confirmDeleteUser', { username: row.username }), { type: 'warning' })
    const res = await userService.delete(row.id)
    if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || 'delete failed')
    ElMessage.success(t('common.success'))
    fetchList()
  } catch (e: any) {
    if (e === 'cancel') return
    ElMessage.error(e?.message || t('common.failed'))
  }
}

// ---------- 启用/禁用 ----------
async function onToggleStatus(row: SysUserItem) {
  try {
    const next = row.status === 1 ? 0 : 1
    const res = await userService.updateUserStatus(row.id, next)
    if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || 'status failed')
    ElMessage.success(t('common.success'))
    fetchList()
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.failed'))
  }
}

// ---------- 重置密码 ----------
const pwdVisible = ref(false)
const { form: pwdForm } = useForm({
  initialData: { id: 0, username: '', password: '' },
})
const { loading: pwdSubmitLoading, withLoading: withPwdLoading } = useLoading()

function openResetPwd(row: SysUserItem) {
  pwdForm.id = row.id
  pwdForm.username = row.username
  pwdForm.password = ''
  pwdVisible.value = true
}

async function submitResetPwd() {
  await withPwdLoading(async () => {
    try {
      if (!pwdForm.password) {
        ElMessage.warning(t('common.pleaseInputNewPassword'))
        return
      }
      const res = await userService.resetPassword(pwdForm.id, pwdForm.password)
      if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || 'reset pwd failed')
      ElMessage.success(t('common.success'))
      pwdVisible.value = false
    } catch (e: any) {
      ElMessage.error(e?.message || t('common.failed'))
    }
  })
}

// ---------- 分配角色 ----------
const roleVisible = ref(false)
const { form: roleForm } = useForm({
  initialData: { userId: 0, username: '', roleIds: [] as number[] },
})
const { loading: roleAssignLoading, withLoading: withRoleAssignLoading } = useLoading()

function openAssignRoles(row: SysUserItem) {
  roleForm.userId = row.id
  roleForm.username = row.username
  roleForm.roleIds = (row.roleIds || []).slice()
  roleVisible.value = true
}

async function submitAssignRoles() {
  await withRoleAssignLoading(async () => {
    try {
      const res = await userService.assignUserRoles(roleForm.userId, roleForm.roleIds)
      if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || 'assign roles failed')
      ElMessage.success(t('common.success'))
      roleVisible.value = false
      fetchList()
    } catch (e: any) {
      ElMessage.error(e?.message || t('common.failed'))
    }
  })
}

// ---------- Google 2FA ----------
const g2Visible = ref(false)
const { form: g2User } = useForm({
  initialData: { userId: 0, username: '' },
})
const { form: g2Init } = useForm({
  initialData: { secret: '', otpauthUrl: '', qrCode: '' },
})
const { form: g2Form } = useForm({
  initialData: { code: '' },
})
const { loading: g2InitLoading, withLoading: withG2InitLoading } = useLoading()
const { loading: g2EnableLoading, withLoading: withG2EnableLoading } = useLoading()
const { loading: g2DisableLoading, withLoading: withG2DisableLoading } = useLoading()

function openGoogle2fa(row: SysUserItem) {
  g2User.userId = row.id
  g2User.username = row.username
  g2Init.secret = ''
  g2Init.otpauthUrl = ''
  g2Init.qrCode = ''
  g2Form.code = ''
  g2Visible.value = true
}

async function doG2Init() {
  await withG2InitLoading(async () => {
    try {
      const res = await userService.initGoogle2FA(g2User.userId)
      if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || 'init failed')
      g2Init.secret = res.data?.secret || ''
      g2Init.otpauthUrl = res.data?.otpauthUrl || ''
      g2Init.qrCode = res.data?.qrCode || ''
      ElMessage.success(t('common.success'))
    } catch (e: any) {
      ElMessage.error(e?.message || t('common.failed'))
    }
  })
}

async function doG2Bind() {
  try {
    if (!g2Form.code) {
      ElMessage.warning(t('common.pleaseInputCode'))
      return
    }
    const res = await userService.bindGoogle2FA(g2User.userId, g2Init.secret, g2Form.code)
    if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || 'bind failed')
    ElMessage.success(t('common.success'))
    fetchList()
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.failed'))
  }
}

async function copySecret() {
  if (!g2Init.secret) {
    ElMessage.warning(t('common.noData'))
    return
  }
  try {
    await navigator.clipboard.writeText(g2Init.secret)
    ElMessage.success(t('common.copied'))
  } catch (e) {
    ElMessage.error(t('common.copyFailed'))
  }
}

async function copyOtpauthUrl() {
  if (!g2Init.otpauthUrl) {
    ElMessage.warning(t('common.noData'))
    return
  }
  try {
    await navigator.clipboard.writeText(g2Init.otpauthUrl)
    ElMessage.success(t('common.copied'))
  } catch (e) {
    ElMessage.error(t('common.copyFailed'))
  }
}

async function doG2Enable() {
  await withG2EnableLoading(async () => {
    try {
      if (!g2Form.code) {
        ElMessage.warning(t('common.pleaseInputCode'))
        return
      }
      const res = await userService.enableGoogle2FA(g2User.userId, g2Form.code)
      if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || 'enable failed')
      ElMessage.success(t('common.success'))
      fetchList()
    } catch (e: any) {
      ElMessage.error(e?.message || t('common.failed'))
    }
  })
}

async function doG2Disable() {
  await withG2DisableLoading(async () => {
    try {
      const res = await userService.disableGoogle2FA(g2User.userId, g2Form.code || undefined)
      if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || 'disable failed')
      ElMessage.success(t('common.success'))
      fetchList()
    } catch (e: any) {
      ElMessage.error(e?.message || t('common.failed'))
    }
  })
}

async function doG2Reset() {
  try {
    await confirm(t('common.confirmReset2fa'), { type: 'warning' })
    const res = await userService.resetGoogle2FA(g2User.userId)
    if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || 'reset failed')
    ElMessage.success(t('common.success'))
    fetchList()
  } catch (e: any) {
    if (e === 'cancel') return
    ElMessage.error(e?.message || t('common.failed'))
  }
}

const roleNameMap = computed(() => {
  const m = new Map<number, string>()
  roles.value.forEach((r) => m.set(r.id, r.name))
  return m
})

onMounted(async () => {
  await Promise.all([fetchRoles(), fetchList()])
})
</script>

<template>
  <el-card>
    <template #header>
      <div style="display: flex; align-items: center; justify-content: space-between; gap: 12px">
        <div>{{ t('system.users') }}</div>

        <div style="display: flex; gap: 8px">
          <el-input
            v-model="queryForm.keyword"
            style="width: 220px"
            clearable
            :placeholder="t('common.accountNicknameKeyword')"
            @keyup.enter="onSearch"
          />
          <el-select
            v-model="queryForm.status"
            style="width: 140px"
            clearable
            :placeholder="t('common.status')"
          >
            <el-option
              v-for="o in statusOptions"
              :key="o.value"
              :label="o.label"
              :value="o.value"
            />
          </el-select>

          <el-button @click="onSearch">
            {{ t('common.search') }}
          </el-button>
          <el-button @click="onReset">
            {{ t('common.reset') }}
          </el-button>

          <el-button v-perm="'sys:user:add'" type="primary" @click="openCreate">
            {{ t('perms.sys:user:add') }}
          </el-button>
        </div>
      </div>
    </template>

    <el-table v-loading="loading" :data="list" row-key="id">
      <el-table-column prop="id" :label="t('common.id')" width="80" />
      <el-table-column prop="username" :label="t('common.username')" min-width="140" />
      <el-table-column prop="nickname" :label="t('common.nickname')" min-width="160" />
      <el-table-column :label="t('common.role')" min-width="180">
        <template #default="{ row }">
          <el-tag v-for="rid in row.roleIds || []" :key="rid" style="margin-right: 6px">
            {{ roleNameMap.get(rid) || '#' + rid }}
          </el-tag>
          <span v-if="!(row.roleIds && row.roleIds.length)" style="color: #999">-</span>
        </template>
      </el-table-column>

      <el-table-column :label="t('common.google2fa')" width="110">
        <template #default="{ row }">
          <el-tag :type="row.google2faEnabled === 1 ? 'success' : 'info'">
            {{ row.google2faEnabled === 1 ? t('common.enabled') : t('common.disabled') }}
          </el-tag>
        </template>
      </el-table-column>

      <el-table-column :label="t('common.status')" width="110">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'">
            {{ row.status === 1 ? t('common.enabled') : t('common.disabled') }}
          </el-tag>
        </template>
      </el-table-column>

      <el-table-column :label="t('common.createTimes')" min-width="170">
        <template #default="{ row }">
          <span style="color: #666">{{ formatDate(row.createTimes) }}</span>
        </template>
      </el-table-column>

      <el-table-column :label="t('common.actions')" width="140" fixed="right">
        <template #default="{ row }">
          <el-dropdown trigger="click">
            <el-button size="small">
              {{ t('common.actions') }}
              <el-icon class="el-icon--right">
                <ArrowDown />
              </el-icon>
            </el-button>

            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item v-perm="'sys:user:update'" @click="openEdit(row)">
                  {{ t('perms.sys:user:update') }}
                </el-dropdown-item>

                <el-dropdown-item v-perm="'sys:user:resetpwd'" @click="openResetPwd(row)">
                  {{ t('perms.sys:user:resetpwd') }}
                </el-dropdown-item>

                <el-dropdown-item v-perm="'sys:user:assignrole'" @click="openAssignRoles(row)">
                  {{ t('perms.sys:user:assignrole') }}
                </el-dropdown-item>

                <el-dropdown-item v-perm="'sys:user:google2fa'" @click="openGoogle2fa(row)">
                  {{ t('perms.sys:user:google2fa') }}
                </el-dropdown-item>

                <el-dropdown-item v-perm="'sys:user:update'" divided @click="onToggleStatus(row)">
                  {{ row.status === 1 ? t('common.disable') : t('common.enable') }}
                </el-dropdown-item>

                <el-dropdown-item v-perm="'sys:user:delete'" @click="onDelete(row)">
                  <span style="color: var(--el-color-danger)">
                    {{ t('perms.sys:user:delete') }}
                  </span>
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
      </el-table-column>
    </el-table>

    <div
      style="
        display: flex;
        justify-content: flex-end;
        gap: 10px;
        align-items: center;
        margin-top: 12px;
      "
    >
      <span>{{ t('common.totalItems', { count: pagination.total }) }}</span>
      <el-button :disabled="!pagination.hasPrev" @click="prevPage">
        {{ t('common.prevPage') }}
      </el-button>
      <el-button :disabled="!pagination.hasNext" @click="nextPage">
        {{ t('common.nextPage') }}
      </el-button>
      <el-select
        v-model="pagination.limit"
        style="width: 100px"
        @change="
          () => {
            pagination.cursor = null
            pagination.hasPrev = false
            fetchList()
          }
        "
      >
        <el-option label="10" :value="10" />
        <el-option label="20" :value="20" />
        <el-option label="50" :value="50" />
      </el-select>
    </div>
  </el-card>

  <!-- 新增/编辑 -->
  <el-dialog
    v-model="editVisible"
    :title="editMode === 'create' ? t('common.addUser') : t('common.editUser')"
    width="520px"
  >
    <el-form label-width="90px">
      <el-form-item v-if="editMode === 'create'" :label="t('common.username')">
        <el-input v-model="editForm.username" />
      </el-form-item>
      <el-form-item v-if="editMode === 'create'" :label="t('common.password')">
        <el-input v-model="editForm.password" type="password" show-password />
      </el-form-item>
      <el-form-item :label="t('common.nickname')">
        <el-input v-model="editForm.nickname" />
      </el-form-item>
      <el-form-item :label="t('common.status')">
        <el-switch v-model="editForm.status" :active-value="1" :inactive-value="0" />
      </el-form-item>
      <el-form-item :label="t('common.role')">
        <el-select
          v-model="editForm.roleIds"
          multiple
          filterable
          style="width: 100%"
          :loading="roleLoading"
        >
          <el-option v-for="r in roles" :key="r.id" :label="r.name" :value="r.id" />
        </el-select>
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="editVisible = false">
        {{ t('common.cancel') }}
      </el-button>
      <el-button type="primary" :loading="editFormLoading" @click="submitEdit">
        {{ t('common.confirm') }}
      </el-button>
    </template>
  </el-dialog>

  <!-- 重置密码 -->
  <el-dialog v-model="pwdVisible" :title="t('common.resetPassword')" width="420px">
    <el-form label-width="90px">
      <el-form-item :label="t('common.username')">
        <el-input :model-value="pwdForm.username" disabled />
      </el-form-item>
      <el-form-item :label="t('common.newPassword')">
        <el-input v-model="pwdForm.password" type="password" show-password />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="pwdVisible = false">
        {{ t('common.cancel') }}
      </el-button>
      <el-button type="primary" :loading="pwdSubmitLoading" @click="submitResetPwd">
        {{ t('common.confirm') }}
      </el-button>
    </template>
  </el-dialog>

  <!-- 分配角色 -->
  <el-dialog v-model="roleVisible" :title="t('common.assignRoles')" width="520px">
    <el-form label-width="90px">
      <el-form-item :label="t('common.username')">
        <el-input :model-value="roleForm.username" disabled />
      </el-form-item>
      <el-form-item :label="t('common.role')">
        <el-select
          v-model="roleForm.roleIds"
          multiple
          filterable
          style="width: 100%"
          :loading="roleLoading"
        >
          <el-option v-for="r in roles" :key="r.id" :label="r.name" :value="r.id" />
        </el-select>
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="roleVisible = false">
        {{ t('common.cancel') }}
      </el-button>
      <el-button type="primary" :loading="roleAssignLoading" @click="submitAssignRoles">
        {{ t('common.confirm') }}
      </el-button>
    </template>
  </el-dialog>

  <!-- Google 2FA -->
  <el-dialog v-model="g2Visible" :title="t('common.google2faManage')" width="680px">
    <div style="display: flex; gap: 16px">
      <div style="flex: 1">
        <div style="margin-bottom: 8px; color: #666">
          {{ t('common.user') }}：{{ g2User.username }}（ID: {{ g2User.userId }}）
        </div>

        <div style="display: flex; gap: 8px; flex-wrap: wrap; margin-bottom: 12px">
          <el-button v-perm="'sys:user:2fa:init'" :loading="g2InitLoading" @click="doG2Init">
            {{ t('perms.sys:user:2fa:init') }}
          </el-button>
          <el-button
            v-perm="'sys:user:2fa:enable'"
            type="success"
            :loading="g2EnableLoading"
            @click="doG2Enable"
          >
            {{ t('perms.sys:user:2fa:enable') }}
          </el-button>
          <el-button
            v-perm="'sys:user:2fa:disable'"
            type="warning"
            :loading="g2DisableLoading"
            @click="doG2Disable"
          >
            {{ t('perms.sys:user:2fa:disable') }}
          </el-button>
          <el-button v-perm="'sys:user:2fa:reset'" type="danger" @click="doG2Reset">
            {{ t('perms.sys:user:2fa:reset') }}
          </el-button>
        </div>

        <el-form label-width="100px">
          <el-form-item :label="t('common.code')">
            <div style="display: flex; gap: 8px">
              <el-input
                v-model="g2Form.code"
                :placeholder="t('common.enterGoogleCode')"
                style="flex: 1"
              />
              <el-button @click="doG2Bind">
                {{ t('perms.sys:user:2fa:bind') }}
              </el-button>
            </div>
          </el-form-item>

          <el-form-item :label="t('common.secret')">
            <div style="display: flex; gap: 8px">
              <el-input :model-value="g2Init.secret" readonly style="flex: 1" />
              <el-button :disabled="!g2Init.secret" @click="copySecret">
                {{ t('common.copy') }}
              </el-button>
            </div>
          </el-form-item>

          <el-form-item :label="t('common.otpauthUrl')">
            <div style="display: flex; gap: 8px">
              <el-input :model-value="g2Init.otpauthUrl" readonly style="flex: 1" />
              <el-button :disabled="!g2Init.otpauthUrl" @click="copyOtpauthUrl">
                {{ t('common.copy') }}
              </el-button>
            </div>
          </el-form-item>
        </el-form>
      </div>

      <div style="width: 260px">
        <div style="margin-bottom: 8px; color: #666">
          {{ t('common.qrCode') }}
        </div>
        <div
          style="
            background: #f7f8fa;
            border: 1px solid #eee;
            border-radius: 8px;
            padding: 12px;
            min-height: 240px;
          "
        >
          <img v-if="g2Init.qrCode" :src="g2Init.qrCode" style="width: 100%; height: auto" />
          <div v-else style="color: #999">
            {{ t('common.click2faBindGenerateQrCode') }}
          </div>
        </div>
        <div v-if="g2Init.qrCode" style="margin-top: 8px; font-size: 12px; color: #666">
          {{ t('common.scanQrCodeWithGoogleAuthenticator') }}
        </div>
      </div>
    </div>

    <template #footer>
      <el-button @click="g2Visible = false">
        {{ t('common.cancel') }}
      </el-button>
    </template>
  </el-dialog>
</template>
