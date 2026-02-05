<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import {
  apiUserList,
  apiUserCreate,
  apiUserUpdate,
  apiUserDelete,
  apiChangeUserStatus,
  apiResetUserPwd,
  apiAssignUserRoles,
  apiGoogle2faInit,
  apiGoogle2faEnable,
  apiGoogle2faDisable,
  apiGoogle2faReset,
  type SysUserItem,
} from '@/api/system/users'
import { apiRoleList, type RoleItem } from '@/api/system/roles'
import { ArrowDown } from '@element-plus/icons-vue'

const { t } = useI18n()

const loading = ref(false)
const list = ref<SysUserItem[]>([])
const total = ref(0)

const query = reactive({
  keyword: '',
  status: undefined as number | undefined,
  page: 1,
  size: 10,
})

const statusOptions = [
  { label: t('common.enabled'), value: 1 },
  { label: t('common.disabled'), value: 2 },
]

async function fetchList() {
  loading.value = true
  try {
    const res = await apiUserList({
      keyword: query.keyword || undefined,
      status: query.status,
      page: query.page,
      size: query.size,
    })
    // 兼容 code=0 / 200
    if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || 'list failed')
    list.value = res.list || []
    total.value = res.total || 0
  } catch (e: any) {
    ElMessage.error(e?.message || '加载失败')
  } finally {
    loading.value = false
  }
}

function onSearch() {
  query.page = 1
  fetchList()
}
function onReset() {
  query.keyword = ''
  query.status = undefined
  query.page = 1
  fetchList()
}

// ---------- 角色缓存（分配角色用） ----------
const roleLoading = ref(false)
const roles = ref<RoleItem[]>([])
async function fetchRoles() {
  roleLoading.value = true
  try {
    const res = await apiRoleList({ page: 1, size: 9999, status: 1 })
    if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || 'role list failed')
    roles.value = res.list || []
  } catch (e: any) {
    ElMessage.error(e?.message || '角色加载失败')
  } finally {
    roleLoading.value = false
  }
}

// ---------- 新增/编辑 ----------
const editVisible = ref(false)
const editMode = ref<'create' | 'update'>('create')
const editForm = reactive({
  id: 0,
  username: '',
  password: '',
  nickname: '',
  status: 1,
  roleIds: [] as number[],
})

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
  try {
    if (editMode.value === 'create') {
      if (!editForm.username || !editForm.password) {
        ElMessage.warning('请输入账号和密码')
        return
      }
      const res = await apiUserCreate({
        username: editForm.username,
        password: editForm.password,
        nickname: editForm.nickname || undefined,
        status: editForm.status,
        roleIds: editForm.roleIds,
      })
      if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || 'create failed')
      ElMessage.success('创建成功')
    } else {
      const res = await apiUserUpdate({
        id: editForm.id,
        nickname: editForm.nickname || undefined,
        status: editForm.status,
        roleIds: editForm.roleIds,
      })
      if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || 'update failed')
      ElMessage.success('更新成功')
    }
    editVisible.value = false
    fetchList()
  } catch (e: any) {
    ElMessage.error(e?.message || '提交失败')
  }
}

// ---------- 删除 ----------
async function onDelete(row: SysUserItem) {
  try {
    await ElMessageBox.confirm(`确定删除用户「${row.username}」？`, '提示', { type: 'warning' })
    const res = await apiUserDelete(row.id)
    if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || 'delete failed')
    ElMessage.success('删除成功')
    fetchList()
  } catch (e: any) {
    if (e === 'cancel') return
    ElMessage.error(e?.message || '删除失败')
  }
}

// ---------- 启用/禁用 ----------
async function onToggleStatus(row: SysUserItem) {
  try {
    const next = row.status === 1 ? 0 : 1
    const res = await apiChangeUserStatus({ id: row.id, status: next })
    if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || 'status failed')
    ElMessage.success('操作成功')
    fetchList()
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}

// ---------- 重置密码 ----------
const pwdVisible = ref(false)
const pwdForm = reactive({ id: 0, username: '', password: '' })

function openResetPwd(row: SysUserItem) {
  pwdForm.id = row.id
  pwdForm.username = row.username
  pwdForm.password = ''
  pwdVisible.value = true
}
async function submitResetPwd() {
  try {
    if (!pwdForm.password) {
      ElMessage.warning('请输入新密码')
      return
    }
    const res = await apiResetUserPwd({ id: pwdForm.id, password: pwdForm.password })
    if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || 'reset pwd failed')
    ElMessage.success('密码已重置')
    pwdVisible.value = false
  } catch (e: any) {
    ElMessage.error(e?.message || '重置失败')
  }
}

// ---------- 分配角色 ----------
const roleVisible = ref(false)
const roleForm = reactive({ userId: 0, username: '', roleIds: [] as number[] })

function openAssignRoles(row: SysUserItem) {
  roleForm.userId = row.id
  roleForm.username = row.username
  roleForm.roleIds = (row.roleIds || []).slice()
  roleVisible.value = true
}
async function submitAssignRoles() {
  try {
    const res = await apiAssignUserRoles({ userId: roleForm.userId, roleIds: roleForm.roleIds })
    if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || 'assign roles failed')
    ElMessage.success('角色已更新')
    roleVisible.value = false
    fetchList()
  } catch (e: any) {
    ElMessage.error(e?.message || '更新失败')
  }
}

// ---------- Google 2FA ----------
const g2Visible = ref(false)
const g2User = reactive({ userId: 0, username: '' })
const g2Init = reactive({ secret: '', otpauthUrl: '', qrCode: '' })
const g2Code = ref('')

function openGoogle2fa(row: SysUserItem) {
  g2User.userId = row.id
  g2User.username = row.username
  g2Init.secret = ''
  g2Init.otpauthUrl = ''
  g2Init.qrCode = ''
  g2Code.value = ''
  g2Visible.value = true
}

async function doG2Init() {
  try {
    const res = await apiGoogle2faInit({ userId: g2User.userId })
    if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || 'init failed')
    g2Init.secret = res.secret || ''
    g2Init.otpauthUrl = res.otpauthUrl || ''
    g2Init.qrCode = res.qrCode || ''
    ElMessage.success('已生成绑定信息')
  } catch (e: any) {
    ElMessage.error(e?.message || '初始化失败')
  }
}

async function doG2Enable() {
  try {
    if (!g2Code.value) {
      ElMessage.warning('请输入验证码')
      return
    }
    const res = await apiGoogle2faEnable({ userId: g2User.userId, code: g2Code.value })
    if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || 'enable failed')
    ElMessage.success('已启用 2FA')
    fetchList()
  } catch (e: any) {
    ElMessage.error(e?.message || '启用失败')
  }
}

async function doG2Disable() {
  try {
    // 你接口里 code optional，按你后端规则：如果强制需要就要求输入
    const res = await apiGoogle2faDisable({ userId: g2User.userId, code: g2Code.value || undefined })
    if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || 'disable failed')
    ElMessage.success('已禁用 2FA')
    fetchList()
  } catch (e: any) {
    ElMessage.error(e?.message || '禁用失败')
  }
}

async function doG2Reset() {
  try {
    await ElMessageBox.confirm('确定重置该用户的 2FA？重置后需要重新绑定。', '提示', { type: 'warning' })
    const res = await apiGoogle2faReset({ userId: g2User.userId })
    if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || 'reset failed')
    ElMessage.success('已重置 2FA')
    fetchList()
  } catch (e: any) {
    if (e === 'cancel') return
    ElMessage.error(e?.message || '重置失败')
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
      <div style="display:flex; align-items:center; justify-content:space-between; gap:12px;">
        <div>{{ t('system.users') }}</div>

        <div style="display:flex; gap:8px;">
          <el-input
            v-model="query.keyword"
            style="width: 220px"
            clearable
            placeholder="账号/昵称关键字"
            @keyup.enter="onSearch"
          />
          <el-select v-model="query.status" style="width: 140px" clearable placeholder="状态">
            <el-option v-for="o in statusOptions" :key="o.value" :label="o.label" :value="o.value" />
          </el-select>

          <el-button @click="onSearch">{{ t('common.search') }}</el-button>
          <el-button @click="onReset">{{ t('common.reset') }}</el-button>

          <el-button type="primary" v-perm="'sys:user:add'" @click="openCreate">
            {{ t('perms.sys:user:add') }}
          </el-button>
        </div>
      </div>
    </template>

    <el-table :data="list" v-loading="loading" row-key="id">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="username" label="账号" min-width="140" />
      <el-table-column prop="nickname" label="昵称" min-width="160" />
      <el-table-column label="角色" min-width="180">
        <template #default="{ row }">
          <el-tag v-for="rid in (row.roleIds || [])" :key="rid" style="margin-right:6px;">
            {{ roleNameMap.get(rid) || ('#' + rid) }}
          </el-tag>
          <span v-if="!(row.roleIds && row.roleIds.length)" style="color:#999;">-</span>
        </template>
      </el-table-column>

      <el-table-column label="2FA" width="110">
        <template #default="{ row }">
          <el-tag :type="row.google2faEnabled === 1 ? 'success' : 'info'">
            {{ row.google2faEnabled === 1 ? '已启用' : '未启用' }}
          </el-tag>
        </template>
      </el-table-column>

      <el-table-column label="状态" width="110">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'">
            {{ row.status === 1 ? t('common.enabled') : t('common.disabled') }}
          </el-tag>
        </template>
      </el-table-column>

      <el-table-column label="创建时间" min-width="170">
        <template #default="{ row }">
          <span style="color:#666;">{{ row.createdAt ? new Date(row.createdAt * 1000).toLocaleString() : '-' }}</span>
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

          <el-dropdown-item divided v-perm="'sys:user:update'" @click="onToggleStatus(row)">
            {{ row.status === 1 ? '禁用' : '启用' }}
          </el-dropdown-item>

          <el-dropdown-item v-perm="'sys:user:delete'" @click="onDelete(row)">
            <span style="color: var(--el-color-danger);">
              {{ t('perms.sys:user:delete') }}
            </span>
          </el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>
  </template>
</el-table-column>

    </el-table>

    <div style="display:flex; justify-content:flex-end; margin-top: 12px;">
      <el-pagination
        background
        layout="total, prev, pager, next, sizes"
        :total="total"
        :page-size="query.size"
        :current-page="query.page"
        @update:current-page="(p:number)=>{query.page=p; fetchList()}"
        @update:page-size="(s:number)=>{query.size=s; query.page=1; fetchList()}"
      />
    </div>
  </el-card>

  <!-- 新增/编辑 -->
  <el-dialog v-model="editVisible" :title="editMode==='create' ? '新增用户' : '编辑用户'" width="520px">
    <el-form label-width="90px">
      <el-form-item label="账号" v-if="editMode==='create'">
        <el-input v-model="editForm.username" />
      </el-form-item>
      <el-form-item label="密码" v-if="editMode==='create'">
        <el-input v-model="editForm.password" type="password" show-password />
      </el-form-item>
      <el-form-item label="昵称">
        <el-input v-model="editForm.nickname" />
      </el-form-item>
      <el-form-item label="状态">
        <el-switch v-model="editForm.status" :active-value="1" :inactive-value="0" />
      </el-form-item>
      <el-form-item label="角色">
        <el-select v-model="editForm.roleIds" multiple filterable style="width: 100%;" :loading="roleLoading">
          <el-option v-for="r in roles" :key="r.id" :label="r.name" :value="r.id" />
        </el-select>
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="editVisible=false">{{ t('common.cancel') }}</el-button>
      <el-button type="primary" @click="submitEdit">{{ t('common.confirm') }}</el-button>
    </template>
  </el-dialog>

  <!-- 重置密码 -->
  <el-dialog v-model="pwdVisible" title="重置密码" width="420px">
    <el-form label-width="90px">
      <el-form-item label="账号">
        <el-input :model-value="pwdForm.username" disabled />
      </el-form-item>
      <el-form-item label="新密码">
        <el-input v-model="pwdForm.password" type="password" show-password />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="pwdVisible=false">{{ t('common.cancel') }}</el-button>
      <el-button type="primary" @click="submitResetPwd">{{ t('common.confirm') }}</el-button>
    </template>
  </el-dialog>

  <!-- 分配角色 -->
  <el-dialog v-model="roleVisible" title="分配角色" width="520px">
    <el-form label-width="90px">
      <el-form-item label="账号">
        <el-input :model-value="roleForm.username" disabled />
      </el-form-item>
      <el-form-item label="角色">
        <el-select v-model="roleForm.roleIds" multiple filterable style="width: 100%;" :loading="roleLoading">
          <el-option v-for="r in roles" :key="r.id" :label="r.name" :value="r.id" />
        </el-select>
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="roleVisible=false">{{ t('common.cancel') }}</el-button>
      <el-button type="primary" @click="submitAssignRoles">{{ t('common.confirm') }}</el-button>
    </template>
  </el-dialog>

  <!-- Google 2FA -->
  <el-dialog v-model="g2Visible" title="Google 2FA 管理" width="680px">
    <div style="display:flex; gap: 16px;">
      <div style="flex: 1;">
        <div style="margin-bottom: 8px; color:#666;">用户：{{ g2User.username }}（ID: {{ g2User.userId }}）</div>

        <div style="display:flex; gap: 8px; flex-wrap: wrap; margin-bottom: 12px;">
          <el-button v-perm="'sys:user:2fa:init'" @click="doG2Init">{{ t('perms.sys:user:2fa:init') }}</el-button>
          <el-button type="success" v-perm="'sys:user:2fa:enable'" @click="doG2Enable">{{ t('perms.sys:user:2fa:enable') }}</el-button>
          <el-button type="warning" v-perm="'sys:user:2fa:disable'" @click="doG2Disable">{{ t('perms.sys:user:2fa:disable') }}</el-button>
          <el-button type="danger" v-perm="'sys:user:2fa:reset'" @click="doG2Reset">{{ t('perms.sys:user:2fa:reset') }}</el-button>
        </div>

        <el-form label-width="100px">
          <el-form-item label="验证码">
            <el-input v-model="g2Code" placeholder="需要时输入 Google Authenticator 6 位码" />
          </el-form-item>

          <el-form-item label="secret">
            <el-input :model-value="g2Init.secret" readonly />
          </el-form-item>

          <el-form-item label="otpauthUrl">
            <el-input :model-value="g2Init.otpauthUrl" readonly />
          </el-form-item>
        </el-form>
      </div>

      <div style="width: 260px;">
        <div style="margin-bottom: 8px; color:#666;">二维码</div>
        <div style="background:#f7f8fa; border:1px solid #eee; border-radius:8px; padding:12px; min-height: 240px;">
          <img v-if="g2Init.qrCode" :src="g2Init.qrCode" style="width:100%; height:auto;" />
          <div v-else style="color:#999;">点击“2FA绑定”生成二维码</div>
        </div>
      </div>
    </div>

    <template #footer>
      <el-button @click="g2Visible=false">{{ t('common.cancel') }}</el-button>
    </template>
  </el-dialog>
</template>

