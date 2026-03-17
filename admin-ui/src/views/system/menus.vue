<template>
  <div class="app-container">
    <el-card shadow="never" class="mb-16">
      <el-form :model="queryForm" inline label-width="80px">
        <el-form-item :label="t('common.keyword')">
          <el-input
            v-model="queryForm.keyword"
            :placeholder="t('system.pleaseInputMenuName')"
            clearable
            style="width: 220px"
            @keyup.enter="handleSearch"
          />
        </el-form-item>

        <el-form-item :label="t('system.menuType')">
          <el-select
            v-model="queryForm.menuType"
            clearable
            :placeholder="t('common.all')"
            style="width: 140px"
          >
            <el-option :label="t('system.directory')" :value="1" />
            <el-option :label="t('system.menu')" :value="2" />
            <el-option :label="t('system.button')" :value="3" />
          </el-select>
        </el-form-item>

        <el-form-item :label="t('common.status')">
          <el-select
            v-model="queryForm.status"
            clearable
            :placeholder="t('common.all')"
            style="width: 140px"
          >
            <el-option :label="t('common.enabled')" :value="1" />
            <el-option :label="t('common.disabled')" :value="0" />
          </el-select>
        </el-form-item>

        <el-form-item :label="t('common.visible')">
          <el-select
            v-model="queryForm.visible"
            clearable
            :placeholder="t('common.all')"
            style="width: 140px"
          >
            <el-option :label="t('common.visible')" :value="1" />
            <el-option :label="t('common.hidden')" :value="0" />
          </el-select>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleSearch">
            {{ t('common.search') }}
          </el-button>
          <el-button @click="handleReset">
            {{ t('common.reset') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never">
      <template #header>
        <div class="toolbar">
          <div class="toolbar-left">
            <span class="card-title">{{ t('system.menus') }}</span>
          </div>
          <div class="toolbar-right">
            <el-button type="primary" @click="handleAdd(0)">
              {{ t('system.addMenu') }}
            </el-button>
            <el-button @click="getList">
              {{ t('common.refresh') }}
            </el-button>
          </div>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="tableData"
        row-key="id"
        border
        :tree-props="{ children: 'children' }"
      >
        <el-table-column :label="t('system.name')" prop="name" min-width="180">
          <template #default="{ row }">
            {{ getMenuTitle(row) }}
          </template>
        </el-table-column>

        <el-table-column :label="t('system.menuType')" width="100" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.menuType === 1" type="warning">{{ t('system.directory') }}</el-tag>
            <el-tag v-else-if="row.menuType === 2" type="success">{{ t('system.menu') }}</el-tag>
            <el-tag v-else type="info">{{ t('system.button') }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column :label="t('system.path')" prop="path" min-width="150" show-overflow-tooltip />
        <el-table-column :label="t('system.component')" prop="component" min-width="180" show-overflow-tooltip />

        <el-table-column :label="t('system.icon')" width="160">
          <template #default="{ row }">
            <div v-if="row.icon" class="menu-icon-cell">
              <el-icon v-if="resolveIconComponent(row.icon)" class="menu-icon-preview">
                <component :is="resolveIconComponent(row.icon)" />
              </el-icon>
              <span class="menu-icon-text">{{ row.icon }}</span>
            </div>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>

        <el-table-column :label="t('system.perms')" prop="perms" min-width="180" show-overflow-tooltip />
        <el-table-column :label="t('system.sort')" prop="sort" width="80" align="center" />

        <el-table-column :label="t('common.visible')" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.visible === 1 ? 'success' : 'info'">
              {{ row.visible === 1 ? t('common.visible') : t('common.hidden') }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column :label="t('common.status')" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? t('common.enabled') : t('common.disabled') }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column :label="t('common.actions')" width="180" fixed="right" align="center">
          <template #default="{ row }" align="right">
            <el-button
              v-if="row.menuType !== 3"
              link
              type="primary"
              @click="handleAdd(row.id)"
            >
              {{ t('system.addChild') }}
            </el-button>
            <el-button link type="primary" @click="handleEdit(row)">
              {{ t('common.edit') }}
            </el-button>
            <el-button link type="danger" @click="handleDelete(row)">
              {{ t('common.delete') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogType === 'add' ? t('system.addMenu') : t('system.editMenu')"
      width="760px"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item :label="t('system.parentMenu')" prop="parentId">
          <el-tree-select
            v-model="formData.parentId"
            :data="parentTreeOptions"
            node-key="id"
            check-strictly
            :render-after-expand="false"
            :props="{ label: 'name', children: 'children', value: 'id' }"
            :placeholder="t('system.pleaseSelectParentMenu')"
            style="width: 100%"
          />
        </el-form-item>

        <el-form-item :label="t('system.menuName')" prop="name">
          <el-input
            v-model="formData.name"
            :placeholder="t('system.pleaseInputMenuName')"
          />
        </el-form-item>

        <el-form-item :label="t('system.menuType')" prop="menuType">
          <el-radio-group v-model="formData.menuType" @change="handleMenuTypeChange">
            <el-radio :label="1">{{ t('system.directory') }}</el-radio>
            <el-radio :label="2">{{ t('system.menu') }}</el-radio>
            <el-radio :label="3">{{ t('system.button') }}</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-row :gutter="16" v-if="formData.menuType !== 3">
          <el-col :span="12">
            <el-form-item :label="t('system.path')" prop="path">
              <el-input
                v-model="formData.path"
                :placeholder="t('system.pleaseInputPath')"
              />
            </el-form-item>
          </el-col>

          <el-col :span="12" v-if="formData.menuType === 2">
            <el-form-item :label="t('system.component')" prop="component">
              <el-input
                v-model="formData.component"
                :placeholder="t('system.pleaseInputComponent')"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="16" v-if="formData.menuType !== 3">
          <el-col :span="14">
            <el-form-item :label="t('system.icon')" prop="icon">
              <div class="icon-picker-box">
                <el-input
                  v-model="formData.icon"
                  :placeholder="t('system.pleaseInputIcon')"
                  clearable
                >
                  <template #prepend>
                    <el-icon v-if="currentIconComponent">
                      <component :is="currentIconComponent" />
                    </el-icon>
                    <span v-else class="text-muted">-</span>
                  </template>
                </el-input>

                <el-popover
                  placement="bottom-start"
                  :width="520"
                  trigger="click"
                >
                  <template #reference>
                    <el-button>{{ t('system.selectIcon') }}</el-button>
                  </template>

                  <div class="icon-panel">
                    <div
                      v-for="iconName in iconNames"
                      :key="iconName"
                      class="icon-item"
                      @click="selectIcon(iconName)"
                    >
                      <el-icon class="icon-item-preview">
                        <component :is="resolveIconComponent(iconName)" />
                      </el-icon>
                      <span class="icon-item-text">{{ iconName }}</span>
                    </div>
                  </div>
                </el-popover>

                <el-button @click="clearIcon">
                  {{ t('common.clear') }}
                </el-button>
              </div>
            </el-form-item>
          </el-col>

          <el-col :span="10">
            <el-form-item :label="t('system.iconPreview')">
              <div class="icon-preview-box">
                <el-icon v-if="currentIconComponent" class="icon-preview-large">
                  <component :is="currentIconComponent" />
                </el-icon>
                <span v-else class="text-muted">{{ t('system.noIcon') }}</span>
              </div>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="16" v-if="formData.menuType !== 3">
          <el-col :span="12">
            <el-form-item :label="t('system.sort')" prop="sort">
              <el-input-number
                v-model="formData.sort"
                :min="0"
                :max="9999"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="16" v-if="formData.menuType === 3">
          <el-col :span="12">
            <el-form-item :label="t('system.sort')" prop="sort">
              <el-input-number
                v-model="formData.sort"
                :min="0"
                :max="9999"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item v-if="formData.menuType === 3" :label="t('system.perms')" prop="perms">
          <el-input
            v-model="formData.perms"
            :placeholder="t('system.pleaseInputPerms')"
          />
        </el-form-item>

        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item :label="t('common.visible')" prop="visible">
              <el-radio-group v-model="formData.visible">
                <el-radio :label="1">{{ t('common.visible') }}</el-radio>
                <el-radio :label="0">{{ t('common.hidden') }}</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>

          <el-col :span="12">
            <el-form-item :label="t('common.status')" prop="status">
              <el-radio-group v-model="formData.status">
                <el-radio :label="1">{{ t('common.enabled') }}</el-radio>
                <el-radio :label="0">{{ t('common.disabled') }}</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import { menuService } from '@/services'
import { useLoading } from '@/composables/useLoading'
import { useForm } from '@/composables/useForm'
import { useConfirm } from '@/composables/useConfirm'
import type { SysMenuCreateReq, SysMenuItem, SysMenuTreeItem, SysMenuUpdateReq } from '@/services/system/MenuService'

const { t, te } = useI18n()

type DialogType = 'add' | 'edit'

type MenuFormData = {
  id: number | undefined
  parentId: number
  name: string
  menuType: number
  path: string
  component: string
  icon: string
  sort: number
  visible: number
  status: number
  perms: string
}

type QueryFormData = {
  keyword: string
  menuType: number | undefined
  status: number | undefined
  visible: number | undefined
}

type RespBase = {
  code?: number
  msg?: string
}

type ApiResp<T = any> = {
  code?: number
  msg?: string
  base?: RespBase
  data?: T
}

const iconMap = ElementPlusIconsVue as Record<string, any>
const iconNames = Object.keys(iconMap).sort()

// Composables
const { loading, withLoading: withMainLoading } = useLoading()
const { loading: submitLoading, withLoading: withSubmitLoading } = useLoading()
const { form: queryForm } = useForm<QueryFormData>({
  initialData: {
    keyword: '',
    menuType: undefined,
    status: undefined,
    visible: undefined,
  }
})
const { confirm } = useConfirm()

// Dialog state
const dialogVisible = ref(false)
const dialogType = ref<DialogType>('add')
const formRef = ref<FormInstance>()

// Query pagination
const queryPage = { page: 1, size: 1000 }

// Menu tree data
const rawList = ref<SysMenuItem[]>([])
const tableData = ref<SysMenuTreeItem[]>([])

const createDefaultForm = (): MenuFormData => ({
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
})

const { form: formData } = useForm<MenuFormData>({
  initialData: createDefaultForm()
})

const currentEditId = computed(() => formData.id ?? 0)

const currentIconComponent = computed(() => {
  return resolveIconComponent(formData.icon)
})

const childrenIdMap = computed(() => {
  const map = new Map<number, number[]>()

  const dfs = (node: SysMenuTreeItem): number[] => {
    const ids: number[] = []
    for (const child of node.children || []) {
      ids.push(child.id)
      ids.push(...dfs(child))
    }
    map.set(node.id, ids)
    return ids
  }

  for (const node of tableData.value) {
    dfs(node)
  }

  return map
})

const parentTreeOptions = computed(() => {
  const excludeIds = new Set<number>()

  if (dialogType.value === 'edit' && currentEditId.value) {
    excludeIds.add(currentEditId.value)
    const childIds = childrenIdMap.value.get(currentEditId.value) || []
    childIds.forEach(id => excludeIds.add(id))
  }

  const filterNodes = (nodes: SysMenuTreeItem[]): SysMenuTreeItem[] => {
    return nodes
      .filter(node => node.menuType !== 3 && !excludeIds.has(node.id))
      .map(node => ({
        ...node,
        children: filterNodes(node.children || []),
      }))
  }

  return [
    {
      id: 0,
      name: t('system.topMenu'),
      children: filterNodes(tableData.value),
    },
  ]
})

const rules = computed<FormRules>(() => ({
  parentId: [
    { required: true, message: t('system.pleaseSelectParentMenu'), trigger: 'change' },
  ],
  name: [
    { required: true, message: t('system.pleaseInputMenuName'), trigger: 'blur' },
  ],
  menuType: [
    { required: true, message: t('system.pleaseSelectMenuType'), trigger: 'change' },
  ],
  path: [
    {
      validator: (_rule, value, callback) => {
        if (formData.menuType !== 3 && !String(value || '').trim()) {
          callback(new Error(t('system.pleaseInputPath')))
          return
        }
        callback()
      },
      trigger: 'blur',
    },
  ],
  component: [
    {
      validator: (_rule, value, callback) => {
        if (formData.menuType === 2 && !String(value || '').trim()) {
          callback(new Error(t('system.pleaseInputComponent')))
          return
        }
        callback()
      },
      trigger: 'blur',
    },
  ],
  perms: [
    {
      validator: (_rule, value, callback) => {
        if (formData.menuType === 3 && !String(value || '').trim()) {
          callback(new Error(t('system.pleaseInputPerms')))
          return
        }
        callback()
      },
      trigger: 'blur',
    },
  ],
  sort: [
    { required: true, message: t('system.pleaseInputSort'), trigger: 'blur' },
  ],
}))

function resolveIconComponent(iconName?: string) {
  if (!iconName) return null
  return iconMap[iconName] || null
}

function selectIcon(iconName: string) {
  formData.icon = iconName
}

function clearIcon() {
  formData.icon = ''
}

function getMenuTitle(row: SysMenuItem) {
  const key = `menu.${row.id}`
  if (te(key)) return t(key)
  return row.name
}

function buildTree(list: SysMenuItem[]): SysMenuTreeItem[] {
  const map = new Map<number, SysMenuTreeItem>()
  const roots: SysMenuTreeItem[] = []

  list.forEach(item => {
    map.set(item.id, {
      ...item,
      children: [],
    })
  })

  map.forEach(item => {
    if (item.parentId === 0) {
      roots.push(item)
      return
    }

    const parent = map.get(item.parentId)
    if (parent) {
      parent.children ||= []
      parent.children.push(item)
    } else {
      roots.push(item)
    }
  })

  const sortTree = (nodes: SysMenuTreeItem[]) => {
    nodes.sort((a, b) => a.sort - b.sort)
    nodes.forEach(node => {
      if (node.children?.length) {
        sortTree(node.children)
      }
    })
  }

  sortTree(roots)
  return roots
}

function resetForm() {
  Object.assign(formData, createDefaultForm())
  nextTick(() => {
    formRef.value?.clearValidate()
  })
}

function normalizeFormByType() {
  if (formData.menuType === 1) {
    formData.component = ''
    formData.perms = ''
  } else if (formData.menuType === 2) {
    formData.perms = ''
  } else if (formData.menuType === 3) {
    formData.path = ''
    formData.component = ''
    formData.icon = ''
  }
}

function handleMenuTypeChange() {
  normalizeFormByType()
  nextTick(() => {
    formRef.value?.clearValidate(['path', 'component', 'perms'])
  })
}

function isSuccessCode(code?: number) {
  return code === 0 || code === 200
}

function getRespCode(res: ApiResp<any>) {
  return res?.code ?? res?.base?.code
}

function getRespMsg(res: ApiResp<any>) {
  return res?.msg || res?.base?.msg || t('common.failed')
}

function assertApiSuccess<T = any>(res: ApiResp<T>, defaultMsg?: string): T {
  const code = getRespCode(res)
  if (!isSuccessCode(code)) {
    throw new Error(getRespMsg(res) || defaultMsg || t('common.failed'))
  }
  return res?.data as T
}

function showError(error: unknown) {
  const msg =
    error instanceof Error
      ? error.message
      : typeof error === 'string'
        ? error
        : t('common.failed')

  ElMessage.error(msg || t('common.failed'))
}

async function getList() {
  await withMainLoading(async () => {
    try {
      const res = await menuService.getList({
        page: queryPage.page,
        size: queryPage.size,
        keyword: queryForm.keyword || '',
        menuType: queryForm.menuType ?? 0,
        status: queryForm.status ?? 0,
        visible: queryForm.visible ?? 0,
      }) as ApiResp<SysMenuItem[]>

      const list = assertApiSuccess<SysMenuItem[]>(res, t('common.failed'))
      rawList.value = Array.isArray(list) ? list : []
      tableData.value = buildTree(rawList.value)
    } catch (error) {
      rawList.value = []
      tableData.value = []
      showError(error)
    }
  })
}

function handleSearch() {
  queryPage.page = 1
  getList()
}

function handleReset() {
  queryForm.keyword = ''
  queryForm.menuType = undefined
  queryForm.status = undefined
  queryForm.visible = undefined
  queryPage.page = 1
  queryPage.size = 20
  getList()
}

function resetFormData() {
  Object.assign(formData, createDefaultForm())
  nextTick(() => {
    formRef.value?.clearValidate()
  })
}

function handleAdd(parentId = 0) {
  resetFormData()
  dialogType.value = 'add'
  formData.parentId = parentId
  dialogVisible.value = true
}

function handleEdit(row: SysMenuItem) {
  resetFormData()
  dialogType.value = 'edit'

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
  })

  dialogVisible.value = true

  nextTick(() => {
    formRef.value?.clearValidate()
  })
}

async function handleDelete(row: SysMenuItem) {
  try {
    await confirm(
      t('system.confirmDeleteMenu', { name: getMenuTitle(row) }),
      { type: 'warning' }
    )

    const res = await menuService.delete(row.id)
    assertApiSuccess(res, t('common.failed'))

    ElMessage.success(t('common.success'))
    await getList()
  } catch (error: any) {
    if (error === 'cancel') {
      return
    }
    showError(error)
  }
}

async function handleSubmit() {
  normalizeFormByType()
  await formRef.value?.validate()

  await withSubmitLoading(async () => {
    try {
      if (dialogType.value === 'add') {
        const payload: SysMenuCreateReq = {
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
        }

        const res = await menuService.create(payload)
        assertApiSuccess(res, t('common.failed'))
      } else {
        const payload: SysMenuUpdateReq = {
          id: formData.id as number,
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
        }

        const res = await menuService.update(formData.id!, payload)
        assertApiSuccess(res, t('common.failed'))
      }

      ElMessage.success(t('common.success'))
      dialogVisible.value = false
      await getList()
    } catch (error) {
      showError(error)
    }
  })
}

onMounted(() => {
  getList()
})
</script>

<style scoped>
.app-container {
  padding: 16px;
}

.mb-16 {
  margin-bottom: 16px;
}

.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.toolbar-left {
  display: flex;
  align-items: center;
}

.toolbar-right {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
}

.menu-icon-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.menu-icon-preview {
  font-size: 16px;
}

.menu-icon-text {
  font-size: 13px;
  color: var(--el-text-color-regular);
}

.icon-picker-box {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
}

.icon-picker-box :deep(.el-input) {
  flex: 1;
}

.icon-panel {
  max-height: 320px;
  overflow-y: auto;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
}

.icon-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  border: 1px solid var(--el-border-color);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.icon-item:hover {
  border-color: var(--el-color-primary);
  background: var(--el-fill-color-light);
}

.icon-item-preview {
  font-size: 18px;
}

.icon-item-text {
  font-size: 12px;
  line-height: 1.2;
  word-break: break-all;
}

.icon-preview-box {
  width: 100%;
  min-height: 84px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px dashed var(--el-border-color);
  border-radius: 8px;
  background: var(--el-fill-color-lighter);
}

.icon-preview-large {
  font-size: 32px;
}

.text-muted {
  color: var(--el-text-color-placeholder);
}
</style>