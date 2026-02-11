<script setup lang="ts">
import { computed, reactive, ref, onMounted, nextTick } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type TreeInstance } from 'element-plus'
import { useI18n } from 'vue-i18n'
import type { SysRole } from '../../types/system/role'
import type { MenuNode, PermItem } from '../../types/system/menus'

// ===== API =====
import { apiRoleList, apiRoleUpdate, apiRoleDelete, apiRoleGrant, apiRoleCreate, apiRoleGrantDetail } from '@/api/system/roles'
import { apiMenuTree, apiPermList } from '@/api/system/menus'

// ===== i18n =====
const { t } = useI18n()

// ===== helpers =====
function isSuperRole(r: SysRole | null | undefined) {
  if (!r) return false
  return r.isSuper === true || r.code === 'super_admin' || r.id === 1
}

// ===== state =====
const loading = ref(false)
const tableData = ref<SysRole[]>([])
const total = ref(0)

const query = reactive({
  keyword: '',
  status: 0 as 0 | 1 | 2, // 0=全部, 1=启用, 2=禁用
  page: 1,
  pageSize: 20,
})

// ===== list =====
async function fetchList() {
  loading.value = true
  try {
    // ✅ 兼容：后端不接受 status=0（全部）时，不传 status
    const q: any = { ...query }
    if (q.status === 0) delete q.status

    const resp = await apiRoleList(q)
    tableData.value = resp.list || []
    total.value = resp.total || 0
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.failed'))
  } finally {
    loading.value = false
  }
}

function unwrapList(resp: any): any[] {
  // 兼容 data / list / rows / result 之类
  if (!resp) return []
  if (Array.isArray(resp)) return resp
  return resp.data || resp.list || resp.rows || resp.result || []
}

function unwrapData(resp: any): any {
  if (!resp) return null
  // detail 这种通常在 data 里
  return resp.data ?? resp
}

// 把扁平菜单转成树（并过滤掉 menuType=3 的按钮项，树里只放目录/菜单）
function buildMenuTree(flat: any[]): any[] {
  const list = (flat || []).filter((x) => x && x.menuType !== 3)

  const map = new Map<number, any>()
  list.forEach((n) => map.set(n.id, { ...n, children: [] }))

  const roots: any[] = []
  map.forEach((node) => {
    const pid = node.parentId
    if (pid && map.has(pid)) {
      map.get(pid).children.push(node)
    } else {
      roots.push(node)
    }
  })

  // 可选：按 sort 排序
  const sortRec = (arr: any[]) => {
    arr.sort((a, b) => (a.sort ?? 0) - (b.sort ?? 0))
    arr.forEach((x) => x.children?.length && sortRec(x.children))
  }
  sortRec(roots)

  return roots
}


function onSearch() {
  query.page = 1
  fetchList()
}

function onReset() {
  query.keyword = ''
  query.status = 0
  onSearch()
}

// ===== create/update dialog =====
const editVisible = ref(false)
const editFormRef = ref<FormInstance>()
const editForm = reactive({
  id: 0,
  name: '',
  code: '',
  remark: '',
  status: 1 as 1 | 2, // 1启用 2禁用
})
const editIsUpdate = computed(() => editForm.id > 0)

function openCreate() {
  Object.assign(editForm, { id: 0, name: '', code: '', remark: '', status: 1 })
  editVisible.value = true
}
function openUpdate(row: SysRole) {
  if (isSuperRole(row)) return
  Object.assign(editForm, {
    id: row.id,
    name: row.name,
    code: row.code,
    remark: row.remark || '',
    status: row.status === 2 ? 2 : 1,
  })
  editVisible.value = true
}

async function submitEdit() {
  await editFormRef.value?.validate?.()
  try {
    const payload = { ...editForm }
    const resp = editIsUpdate.value ? await apiRoleUpdate(payload) : await apiRoleCreate(payload)
    if (resp.code === 200) {
      ElMessage.success(resp.msg || t('common.success'))
      editVisible.value = false
      fetchList()
    } else {
      ElMessage.error(resp.msg || t('common.failed'))
    }
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.failed'))
  }
}

async function onDelete(row: SysRole) {
  if (isSuperRole(row)) return
  await ElMessageBox.confirm(t('common.confirmDelete'), t('common.tip'), { type: 'warning' })
  try {
    const resp = await apiRoleDelete(row.id)
    if (resp.code === 200) {
      ElMessage.success(resp.msg || t('common.success'))
      fetchList()
    } else {
      ElMessage.error(resp.msg || t('common.failed'))
    }
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.failed'))
  }
}

// ===== grant dialog =====
const grantVisible = ref(false)
const currentRole = ref<SysRole | null>(null)

const grantLoading = ref(false)
const menuTree = ref<MenuNode[]>([])
const permList = ref<PermItem[]>([])

const menuTreeRef = ref<TreeInstance>()
const checkedPermKeys = ref<string[]>([])

const grantReadonly = computed(() => isSuperRole(currentRole.value))

function openGrant(row: SysRole) {
  currentRole.value = row
  grantVisible.value = true
  initGrant(row.id)
}

async function initGrant(roleId: number) {
  grantLoading.value = true
  try {
    // ✅ 每次打开先清理（避免切角色残留）
    menuTree.value = []
    permList.value = []
    checkedPermKeys.value = []
    await nextTick()
    menuTreeRef.value?.setCheckedKeys?.([])

    const [menusResp, permsResp, detailResp] = await Promise.all([
      apiMenuTree(),
      apiPermList(),
      apiRoleGrantDetail(roleId),
    ])

    const menusFlat = unwrapList(menusResp)
    const perms = unwrapList(permsResp)
    const detail = unwrapData(detailResp) || {}

    menuTree.value = buildMenuTree(menusFlat)
    permList.value = perms

    await nextTick()
    menuTreeRef.value?.setCheckedKeys((detail.menuIds || []) as any)
    checkedPermKeys.value = (detail.permKeys || []) as any
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.failed'))
  } finally {
    grantLoading.value = false
  }
}

function collectCheckedMenuIds(): number[] {
  const tree = menuTreeRef.value
  if (!tree) return []
  // Element Plus tree：getCheckedKeys + getHalfCheckedKeys
  const full = (tree.getCheckedKeys?.() || []) as number[]
  const half = (tree.getHalfCheckedKeys?.() || []) as number[]
  return Array.from(new Set([...full, ...half]))
}

async function submitGrant() {
  if (!currentRole.value) return
  if (grantReadonly.value) {
    ElMessage.warning(t('system.superAdminNoGrant'))
    return
  }

  try {
    const payload = {
      roleId: currentRole.value.id,
      menuIds: collectCheckedMenuIds(),
      permKeys: checkedPermKeys.value,
    }
    const resp = await apiRoleGrant(payload)
    if (resp.code === 200) {
      ElMessage.success(resp.msg || t('common.success'))
      grantVisible.value = false
      // ✅ 可选：保存后刷新列表（比如你要展示更新时间/状态）
      fetchList()
    } else {
      ElMessage.error(resp.msg || t('common.failed'))
    }
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.failed'))
  }
}

function onGrantClosed() {
  currentRole.value = null
  grantLoading.value = false
  menuTree.value = []
  permList.value = []
  checkedPermKeys.value = []
  menuTreeRef.value?.setCheckedKeys?.([])
}

// ===== init =====
onMounted(fetchList)
</script>

<template>
  <el-card>
    <template #header>
      <div style="display:flex; align-items:center; justify-content:space-between; gap:12px;">
        <div>{{ t('system.roles') }}</div>
        <div style="display:flex; gap: 8px; flex-wrap: wrap;">
          <el-button type="primary" v-perm="'sys:role:add'" @click="openCreate">
            {{ t('perms.sys:role:add') }}
          </el-button>
        </div>
      </div>
    </template>

    <!-- 查询区 -->
    <div style="display:flex; gap:8px; align-items:center; margin-bottom:12px; flex-wrap: wrap;">
      <el-input v-model="query.keyword" :placeholder="t('common.keyword')" clearable style="max-width:260px;" />

      <el-select v-model="query.status" style="width:140px;" :placeholder="t('common.status')" @change="onSearch">
        <el-option :label="t('common.all')" :value="0" />
        <el-option :label="t('common.enabled')" :value="1" />
        <el-option :label="t('common.disabled')" :value="2" />
      </el-select>

      <el-button @click="onSearch">{{ t('common.search') }}</el-button>
      <el-button @click="onReset">{{ t('common.reset') }}</el-button>
    </div>

    <!-- 表格 -->
    <el-table :data="tableData" v-loading="loading" style="width:100%;">
      <el-table-column prop="id" label="ID" width="90" />
      <el-table-column prop="name" :label="t('system.roleName')" min-width="160" />
      <el-table-column prop="code" :label="t('system.roleCode')" min-width="160" />

      <!-- 状态 -->
      <el-table-column :label="t('common.status')" width="110">
        <template #default="{ row }">
          <el-tag v-if="(row as any).status === 1" type="success">{{ t('common.enabled') }}</el-tag>
          <el-tag v-else type="info">{{ t('common.disabled') }}</el-tag>
        </template>
      </el-table-column>

      <el-table-column prop="remark" :label="t('common.remark')" min-width="200" />

      <el-table-column :label="t('common.actions')" width="320" fixed="right">
        <template #default="{ row }">
          <el-button
            size="small"
            v-perm="'sys:role:update'"
            :disabled="isSuperRole(row)"
            @click="openUpdate(row)"
          >
            {{ t('common.edit') }}
          </el-button>

          <el-button
            size="small"
            v-perm="'sys:role:grant'"
            :disabled="isSuperRole(row)"
            @click="openGrant(row)"
          >
            {{ t('system.grant') }}
          </el-button>

          <el-button
            size="small"
            type="danger"
            v-perm="'sys:role:delete'"
            :disabled="isSuperRole(row)"
            @click="onDelete(row)"
          >
            {{ t('common.delete') }}
          </el-button>

          <el-tag v-if="isSuperRole(row)" type="warning" style="margin-left:8px;">
            {{ t('system.superAdmin') }}
          </el-tag>
        </template>
      </el-table-column>
    </el-table>

    <div style="display:flex; justify-content:flex-end; margin-top:12px;">
      <el-pagination
        v-model:current-page="query.page"
        v-model:page-size="query.pageSize"
        :total="total"
        background
        layout="total, prev, pager, next, sizes"
        @current-change="fetchList"
        @size-change="() => { query.page = 1; fetchList() }"
      />
    </div>
  </el-card>

  <!-- 新增/编辑弹窗 -->
  <el-dialog v-model="editVisible" :title="editIsUpdate ? t('system.roleEdit') : t('system.roleAdd')" width="520px">
    <el-form ref="editFormRef" :model="editForm" label-width="110px">
      <el-form-item :label="t('system.roleName')" prop="name" :rules="[{ required: true, message: t('common.required') }]">
        <el-input v-model="editForm.name" />
      </el-form-item>

      <el-form-item :label="t('system.roleCode')" prop="code" :rules="[{ required: true, message: t('common.required') }]">
        <el-input v-model="editForm.code" :disabled="editIsUpdate" />
      </el-form-item>

      <el-form-item :label="t('common.status')" prop="status" :rules="[{ required: true, message: t('common.required') }]">
        <el-radio-group v-model="editForm.status">
          <el-radio :label="1">{{ t('common.enabled') }}</el-radio>
          <el-radio :label="2">{{ t('common.disabled') }}</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item :label="t('common.remark')" prop="remark">
        <el-input v-model="editForm.remark" type="textarea" :rows="3" />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="editVisible = false">{{ t('common.cancel') }}</el-button>
      <el-button type="primary" @click="submitEdit">{{ t('common.confirm') }}</el-button>
    </template>
  </el-dialog>

  <!-- 授权弹窗：菜单 + 按钮权限 -->
  <el-dialog
    v-model="grantVisible"
    :title="t('system.grantTitle', { role: currentRole?.name || '' })"
    width="900px"
    @closed="onGrantClosed"
  >
    <el-alert
      v-if="grantReadonly"
      type="warning"
      :closable="false"
      :title="t('system.superAdminAllPerms')"
      style="margin-bottom:12px;"
    />

    <el-tabs v-loading="grantLoading">
      <el-tab-pane :label="t('system.grantMenu')">
        <div v-if="!menuTree || menuTree.length === 0" style="padding:24px; text-align:center; color:#999;">
          {{ t('common.noData') }}
        </div>
        <el-tree
          v-else
          ref="menuTreeRef"
          :data="menuTree"
          node-key="id"
          show-checkbox
          :props="{ label: 'name', children: 'children' }"
          :check-strictly="false"
          :disabled="grantReadonly"
          default-expand-all
          style="max-height:520px; overflow:auto;"
        />
      </el-tab-pane>

      <el-tab-pane :label="t('system.grantPerms')">
        <div v-if="!permList || permList.length === 0" style="padding:24px; text-align:center; color:#999;">
          {{ t('common.noData') }}
        </div>

        <el-checkbox-group v-else v-model="checkedPermKeys" :disabled="grantReadonly" style="display:block;">
          <div style="display:flex; flex-wrap:wrap; gap:10px;">
            <el-checkbox
              v-for="p in permList"
              :key="p.key"
              :label="p.key"
              border
            >
              {{ p.name }} ({{ p.key }})
            </el-checkbox>
          </div>
        </el-checkbox-group>
      </el-tab-pane>
    </el-tabs>

    <template #footer>
      <el-button @click="grantVisible = false">{{ t('common.cancel') }}</el-button>
      <el-button type="primary" :disabled="grantReadonly" @click="submitGrant">
        {{ t('common.save') }}
      </el-button>
    </template>
  </el-dialog>
</template>
