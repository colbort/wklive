<script setup lang="ts">
import { computed, reactive, ref, onMounted, nextTick } from 'vue'
import { ElMessage, type FormInstance, type TreeInstance } from 'element-plus'
import { useI18n } from 'vue-i18n'
import type { SysRole } from '@/services/system/RoleService'
import type { MenuNode, PermItem } from '@/services/system/MenuService'
import { usePagination, useLoading, useConfirm, useForm } from '@/composables'

import { roleService, menuService } from '@/services'

// ===== i18n =====
const { t } = useI18n()

// ===== helpers =====
function isSuperRole(r: SysRole | null | undefined) {
  if (!r) return false
  return r.isSuper === true || r.code === 'super_admin' || r.id === 1
}

// ===== state =====
const { pagination, updatePagination, nextPage: paginationNextPage, prevPage: paginationPrevPage } = usePagination(20)
const { loading, withLoading } = useLoading()
const { confirm } = useConfirm()
const { form: queryForm } = useForm({
  initialData: { keyword: '', status: 0 as 0 | 1 | 2 },
})

const tableData = ref<SysRole[]>([])

// ===== list =====
async function fetchList() {
  await withLoading(async () => {
    try {
      const q: any = {
        keyword: queryForm.keyword,
        status: queryForm.status,
        cursor: pagination.cursor,
        limit: pagination.limit,
      }
      if (q.status === 0) delete q.status

      const resp = await roleService.getList(q)
      tableData.value = resp.data || []
      updatePagination(resp.total || 0, resp.hasNext || false, resp.hasPrev || false, resp.nextCursor || null, resp.prevCursor || null)
    } catch (e: any) {
      ElMessage.error(e?.message || t('common.failed'))
    }
  })
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

// 把扁平菜单转成树（包含按钮权限 menuType=3，以树状方式展示）
function buildMenuTree(flat: any[]): any[] {
  const list = (flat || []).filter((x) => x)

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
  pagination.cursor = null
  pagination.hasPrev = false
  fetchList()
}

function onReset() {
  queryForm.keyword = ''
  queryForm.status = 0
  onSearch()
}

function nextPage() {
  paginationNextPage()
  fetchList()
}

function prevPage() {
  paginationPrevPage()
  fetchList()
}

// ===== create/update dialog =====
const editVisible = ref(false)
const editFormRef = ref<FormInstance>()
const { form: editForm } = useForm({
  initialData: {
    id: 0,
    name: '',
    code: '',
    remark: '',
    status: 1 as 1 | 2,
  },
})
const editIsUpdate = computed(() => editForm.id > 0)
const { loading: editLoading, withLoading: withEditLoading } = useLoading()

function openCreate() {
  editForm.id = 0
  editForm.name = ''
  editForm.code = ''
  editForm.remark = ''
  editForm.status = 1
  editVisible.value = true
}
function openUpdate(row: SysRole) {
  if (isSuperRole(row)) return
  editForm.id = row.id
  editForm.name = row.name
  editForm.code = row.code
  editForm.remark = row.remark || ''
  editForm.status = row.status === 2 ? 2 : 1
  editVisible.value = true
}

async function submitEdit() {
  await editFormRef.value?.validate?.()
  await withEditLoading(async () => {
    try {
      const payload = { ...editForm }
      const resp = editIsUpdate.value
        ? await roleService.update(editForm.id, payload)
        : await roleService.create(payload)
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
  })
}

async function onDelete(row: SysRole) {
  if (isSuperRole(row)) return
  try {
    await confirm(t('common.confirmDelete'), { type: 'warning' })
    const resp = await roleService.delete(row.id)
    if (resp.code === 200) {
      ElMessage.success(resp.msg || t('common.success'))
      fetchList()
    } else {
      ElMessage.error(resp.msg || t('common.failed'))
    }
  } catch (e: any) {
    if (e === 'cancel') return
    ElMessage.error(e?.message || t('common.failed'))
  }
}

// ===== grant dialog =====
const grantVisible = ref(false)
const currentRole = ref<SysRole | null>(null)

const { loading: grantLoading, withLoading: withGrantLoading } = useLoading()
const menuTree = ref<MenuNode[]>([])
const menuNodeMap = ref<Map<number, any>>(new Map())
const permList = ref<PermItem[]>([])

const menuTreeRef = ref<TreeInstance>()
const checkedPermKeys = ref<string[]>([])

const grantReadonly = computed(() => isSuperRole(currentRole.value))

function flattenMenuTree(nodes: any[], result: any[] = []): any[] {
  for (const node of nodes || []) {
    result.push(node)
    if (node.children?.length) {
      flattenMenuTree(node.children, result)
    }
  }
  return result
}

function updateMenuNodeMap(nodes: any[]) {
  const map = new Map<number, any>()
  flattenMenuTree(nodes).forEach((node) => {
    if (node && node.id != null) {
      map.set(node.id, node)
    }
  })
  menuNodeMap.value = map
}

function getCheckedButtonPermKeys(): string[] {
  const checkedNodes = (menuTreeRef.value?.getCheckedNodes?.() || []) as any[]
  return checkedNodes
    .filter((node) => node.menuType === 3 && node.perms)
    .map((node) => node.perms as string)
    .filter(Boolean)
}

function onMenuTreeCheck() {
  checkedPermKeys.value = getCheckedButtonPermKeys()
}

function openGrant(row: SysRole) {
  currentRole.value = row
  grantVisible.value = true
  initGrant(row.id)
}

async function initGrant(roleId: number) {
  await withGrantLoading(async () => {
    try {
      // ✅ 每次打开先清理（避免切角色残留）
      menuTree.value = []
      permList.value = []
      checkedPermKeys.value = []
      await nextTick()
      menuTreeRef.value?.setCheckedKeys?.([])

      const [menusResp, permsResp, detailResp] = await Promise.all([
        menuService.getMenuTree(),
        menuService.getPermissionList(),
        roleService.getRoleGrantDetail(roleId),
      ])

      const menusFlat = unwrapList(menusResp)
      const perms = unwrapList(permsResp)
      const detail = unwrapData(detailResp) || {}

      menuTree.value = buildMenuTree(menusFlat)
      permList.value = perms
      updateMenuNodeMap(menuTree.value)

      // 角色原有菜单 + 按钮权限需要转成对应 node id
      const menuIds = Array.isArray(detail.menuIds) ? detail.menuIds : []
      const permKeys = Array.isArray(detail.permKeys) ? detail.permKeys : []
      const permKeyToId = new Map<string, number>()
      flattenMenuTree(menuTree.value).forEach((node) => {
        if (node.menuType === 3 && node.perms) {
          permKeyToId.set(node.perms, node.id)
        }
      })
      const buttonIds = permKeys
        .map((k: string) => permKeyToId.get(k))
        .filter((id: number | undefined): id is number => id != null)

      await nextTick()
      menuTreeRef.value?.setCheckedKeys([...menuIds, ...buttonIds] as any)
      checkedPermKeys.value = permKeys as string[]
    } catch (e: any) {
      ElMessage.error(e?.message || t('common.failed'))
    }
  })
}

function collectCheckedMenuIds(): number[] {
  const tree = menuTreeRef.value
  if (!tree) return []
  // Element Plus tree：getCheckedKeys + getHalfCheckedKeys
  const full = (tree.getCheckedKeys?.() || []) as number[]
  const half = (tree.getHalfCheckedKeys?.() || []) as number[]
  const allIds = Array.from(new Set([...full, ...half]))

  return allIds.filter((id) => {
    const node = menuNodeMap.value.get(Number(id))
    return node && node.menuType !== 3
  })
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
    const resp = await roleService.grantRole(payload)
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
      <el-input v-model="queryForm.keyword" :placeholder="t('common.keyword')" clearable style="max-width:260px;" />

      <el-select v-model="queryForm.status" style="width:140px;" :placeholder="t('common.status')" @change="onSearch">
        <el-option :label="t('common.all')" :value="0" />
        <el-option :label="t('common.enabled')" :value="1" />
        <el-option :label="t('common.disabled')" :value="2" />
      </el-select>

      <el-button @click="onSearch">{{ t('common.search') }}</el-button>
      <el-button @click="onReset">{{ t('common.reset') }}</el-button>
    </div>

    <!-- 表格 -->
    <el-table :data="tableData" v-loading="loading" style="width:100%;">
      <el-table-column prop="id" :label="t('common.id')" width="90" />
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

    <div style="display:flex; justify-content:flex-end; gap: 10px; align-items: center; margin-top:12px;">
      <span>{{ t('common.totalItems', { count: pagination.total }) }}</span>
      <el-button @click="prevPage" :disabled="!pagination.hasPrev">{{ t('common.prevPage') }}</el-button>
      <el-button @click="nextPage" :disabled="!pagination.hasNext">{{ t('common.nextPage') }}</el-button>
      <el-select v-model="pagination.limit" style="width: 100px" @change="() => { pagination.cursor = null; pagination.hasPrev = false; fetchList() }">
        <el-option label="10" :value="10" />
        <el-option label="20" :value="20" />
        <el-option label="50" :value="50" />
      </el-select>
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
      <el-button type="primary" :loading="editLoading" @click="submitEdit">{{ t('common.confirm') }}</el-button>
    </template>
  </el-dialog>

  <!-- 授权弹窗：菜单 + 按钮权限 -->
  <el-dialog
    v-model="grantVisible"
    :title="t('system.grantTitle', { role: currentRole?.name || '' })"
    width="400px"
    :style="{ maxWidth: '460px' }"
    @closed="onGrantClosed"
  >
    <el-alert
      v-if="grantReadonly"
      type="warning"
      :closable="false"
      :title="t('system.superAdminAllPerms')"
      style="margin-bottom:12px;"
    />

    <div v-loading="grantLoading">
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
        :default-expand-all="false"
        @check="onMenuTreeCheck"
        style="max-height:420px; overflow:auto;"
      />
    </div>

    <template #footer>
      <el-button @click="grantVisible = false">{{ t('common.cancel') }}</el-button>
      <el-button type="primary" :disabled="grantReadonly" @click="submitGrant">
        {{ t('common.save') }}
      </el-button>
    </template>
  </el-dialog>
</template>
