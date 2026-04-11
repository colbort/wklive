<template>
  <div class="itick-tenant-categories">
    <div class="page-header">
      <h2>租户分类配置</h2>
      <div class="header-actions">
        <el-button type="primary" :disabled="!queryParams.tenantId" @click="handleAdd">
          <el-icon><Plus /></el-icon>
          新增
        </el-button>
        <el-button :disabled="!queryParams.tenantId" @click="openBatchDialog">
          <el-icon><EditPen /></el-icon>
          批量配置
        </el-button>
        <el-button @click="refreshCurrentPage">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <el-card class="query-card" shadow="never">
      <el-form :model="queryParams" inline label-width="90px">
        <el-form-item label="租户ID">
          <el-input-number
            v-model="queryParams.tenantId"
            :min="1"
            :precision="0"
            controls-position="right"
            style="width: 180px"
          />
        </el-form-item>

        <el-form-item label="分类类型">
          <el-input-number
            v-model="queryParams.categoryType"
            :min="0"
            :precision="0"
            controls-position="right"
            style="width: 180px"
          />
        </el-form-item>

        <el-form-item label="启用状态">
          <el-select v-model="queryParams.status" clearable style="width: 180px">
            <el-option label="全部" :value="0" />
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="2" />
          </el-select>
        </el-form-item>

        <el-form-item label="显示状态">
          <el-select v-model="queryParams.visibleStatus" clearable style="width: 180px">
            <el-option label="全部" :value="0" />
            <el-option label="显示" :value="1" />
            <el-option label="隐藏" :value="2" />
          </el-select>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleQuery">
            搜索
          </el-button>
          <el-button @click="resetQuery">
            重置
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="table-card" shadow="never">
      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="tenantId" label="租户ID" width="100" />
        <el-table-column prop="categoryId" label="分类ID" width="100" />
        <el-table-column prop="categoryType" label="分类类型" width="100" />
        <el-table-column prop="categoryName" label="分类名称" min-width="180" />

        <el-table-column label="启用状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.enabled === 1 ? 'success' : 'info'">
              {{ row.enabled === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="APP显示" width="100">
          <template #default="{ row }">
            <el-tag :type="row.appVisible === 1 ? 'success' : 'warning'">
              {{ row.appVisible === 1 ? '显示' : '隐藏' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="sort" label="排序" width="90" />
        <el-table-column
          prop="icon"
          label="图标"
          min-width="160"
          show-overflow-tooltip
        />
        <el-table-column
          prop="remark"
          label="备注"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column label="创建时间" min-width="170">
          <template #default="{ row }">
            {{ formatDate(row.createTimes) }}
          </template>
        </el-table-column>
        <el-table-column label="更新时间" min-width="170">
          <template #default="{ row }">
            {{ formatDate(row.updateTimes) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleDetail(row)">
              详情
            </el-button>
            <el-button link type="primary" @click="handleEdit(row)">
              编辑
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-bar">
        <span>共 {{ pagination.total }} 条</span>
        <el-button :disabled="!pagination.hasPrev" @click="handlePrevPage">
          上一页
        </el-button>
        <el-button :disabled="!pagination.hasNext" type="primary" @click="handleNextPage">
          下一页
        </el-button>
        <el-select v-model="pagination.limit" style="width: 100px" @change="handleLimitChange">
          <el-option :value="10" label="10条/页" />
          <el-option :value="20" label="20条/页" />
          <el-option :value="50" label="50条/页" />
          <el-option :value="100" label="100条/页" />
        </el-select>
      </div>
    </el-card>

    <el-dialog
      v-model="formDialogVisible"
      :title="formMode === 'add' ? '新增租户分类' : '编辑租户分类'"
      width="620px"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="110px"
      >
        <el-form-item label="租户ID" prop="tenantId">
          <el-input-number
            v-model="form.tenantId"
            :min="1"
            :precision="0"
            controls-position="right"
            style="width: 100%"
            :disabled="formMode === 'edit'"
          />
        </el-form-item>

        <el-form-item v-if="formMode === 'add'" label="分类ID" prop="categoryId">
          <el-input-number
            v-model="form.categoryId"
            :min="1"
            :precision="0"
            controls-position="right"
            style="width: 100%"
          />
        </el-form-item>

        <el-form-item label="启用状态" prop="enabled">
          <el-radio-group v-model="form.enabled">
            <el-radio :value="1">
              启用
            </el-radio>
            <el-radio :value="2">
              禁用
            </el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="APP显示" prop="appVisible">
          <el-radio-group v-model="form.appVisible">
            <el-radio :value="1">
              显示
            </el-radio>
            <el-radio :value="2">
              隐藏
            </el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="排序" prop="sort">
          <el-input-number
            v-model="form.sort"
            :min="0"
            :precision="0"
            controls-position="right"
            style="width: 100%"
          />
        </el-form-item>

        <el-form-item label="备注" prop="remark">
          <el-input
            v-model="form.remark"
            type="textarea"
            :rows="4"
            maxlength="200"
            show-word-limit
            placeholder="请输入备注"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="formDialogVisible = false">
          取消
        </el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitForm">
          确定
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailDialogVisible" title="租户分类详情" width="700px">
      <el-descriptions v-loading="detailLoading" :column="2" border>
        <el-descriptions-item label="ID">
          {{ detail.id ?? '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="租户ID">
          {{ detail.tenantId ?? '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="分类ID">
          {{ detail.categoryId ?? '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="分类类型">
          {{
            detail.categoryType ?? '-'
          }}
        </el-descriptions-item>
        <el-descriptions-item label="分类名称">
          {{
            detail.categoryName || '-'
          }}
        </el-descriptions-item>
        <el-descriptions-item label="图标">
          {{ detail.icon || '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="启用状态">
          {{ detail.enabled === 1 ? '启用' : detail.enabled === 2 ? '禁用' : '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="APP显示">
          {{ detail.appVisible === 1 ? '显示' : detail.appVisible === 2 ? '隐藏' : '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="排序">
          {{ detail.sort ?? '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="备注">
          {{ detail.remark || '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">
          {{ formatDate(detail.createTimes??0) }}
        </el-descriptions-item>
        <el-descriptions-item label="更新时间">
          {{ formatDate(detail.updateTimes??0) }}
        </el-descriptions-item>
      </el-descriptions>

      <template #footer>
        <el-button type="primary" @click="detailDialogVisible = false">
          关闭
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="batchDialogVisible" title="批量配置租户分类" width="920px">
      <div class="batch-toolbar">
        <div class="batch-tip">
          未提交的记录会被视为删除，请谨慎保存。
        </div>
        <div class="batch-actions">
          <el-button @click="appendBatchRow">
            新增一行
          </el-button>
          <el-button type="primary" :loading="batchSubmitting" @click="submitBatch">
            保存批量配置
          </el-button>
        </div>
      </div>

      <el-table :data="batchRows" border>
        <el-table-column label="分类ID" min-width="120">
          <template #default="{ row }">
            <el-input-number
              v-model="row.categoryId"
              :min="1"
              :precision="0"
              controls-position="right"
              style="width: 100%"
            />
          </template>
        </el-table-column>
        <el-table-column label="启用状态" min-width="130">
          <template #default="{ row }">
            <el-select v-model="row.enabled" style="width: 100%">
              <el-option label="启用" :value="1" />
              <el-option label="禁用" :value="2" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column label="APP显示" min-width="130">
          <template #default="{ row }">
            <el-select v-model="row.appVisible" style="width: 100%">
              <el-option label="显示" :value="1" />
              <el-option label="隐藏" :value="2" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column label="排序" min-width="120">
          <template #default="{ row }">
            <el-input-number
              v-model="row.sort"
              :min="0"
              :precision="0"
              controls-position="right"
              style="width: 100%"
            />
          </template>
        </el-table-column>
        <el-table-column label="备注" min-width="220">
          <template #default="{ row }">
            <el-input v-model="row.remark" maxlength="200" show-word-limit />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="90" fixed="right">
          <template #default="{ $index }">
            <el-button link type="danger" @click="removeBatchRow($index)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue'
import { EditPen, Plus, Refresh } from '@element-plus/icons-vue'
import { ElMessage, type FormRules } from 'element-plus'
import { usePagination } from '@/composables/usePagination'
import { useLoading } from '@/composables/useLoading'
import { useForm } from '@/composables/useForm'
import {
  tenantCategoriesService,
  type ItickTenantCategory,
  type ListTenantCategoriesReq,
  type TenantCategoryItem,
} from '@/services/itick/TenantCategoriesService'
import { formatDate } from '@/utils'

type FormData = {
  id?: number
  tenantId?: number
  categoryId?: number
  enabled: number
  appVisible: number
  sort: number
  remark: string
}

const { pagination, updatePagination, reset: resetPagination } = usePagination(20)
const { loading, withLoading } = useLoading()

const { form: queryParams, reset: resetQueryParams } = useForm<ListTenantCategoriesReq>({
  initialData: {
    tenantId: 0,
    categoryType: 0,
    status: 0,
    visibleStatus: 0,
    cursor: null,
    limit: 20,
  },
})

const { form, formRef, reset: resetForm } = useForm<FormData>({
  initialData: {
    id: undefined,
    tenantId: undefined,
    categoryId: undefined,
    enabled: 1,
    appVisible: 1,
    sort: 0,
    remark: '',
  },
})

const list = ref<ItickTenantCategory[]>([])
const detail = ref<Partial<ItickTenantCategory>>({})
const detailLoading = ref(false)
const submitLoading = ref(false)
const batchSubmitting = ref(false)
const formDialogVisible = ref(false)
const detailDialogVisible = ref(false)
const batchDialogVisible = ref(false)
const formMode = ref<'add' | 'edit'>('add')
const batchRows = ref<TenantCategoryItem[]>([])

const rules: FormRules<FormData> = {
  tenantId: [{ required: true, message: '请输入租户ID', trigger: 'blur' }],
  categoryId: [{ required: true, message: '请输入分类ID', trigger: 'blur' }],
  enabled: [{ required: true, message: '请选择启用状态', trigger: 'change' }],
  appVisible: [{ required: true, message: '请选择显示状态', trigger: 'change' }],
}

const cleanedQueryParams = computed<ListTenantCategoriesReq | null>(() => {
  if (!queryParams.tenantId) {
    return null
  }

  const params: ListTenantCategoriesReq = {
    tenantId: Number(queryParams.tenantId),
    cursor: pagination.cursor,
    limit: pagination.limit,
  }

  if (queryParams.categoryType) {
    params.categoryType = Number(queryParams.categoryType)
  }
  if (queryParams.status) {
    params.status = queryParams.status
  }
  if (queryParams.visibleStatus) {
    params.visibleStatus = queryParams.visibleStatus
  }

  return params
})

const getList = async () => {
  if (!cleanedQueryParams.value) {
    list.value = []
    updatePagination(0, false, false, null, null)
    return
  }

  await withLoading(async () => {
    const res = await tenantCategoriesService.getList(cleanedQueryParams.value as ListTenantCategoriesReq)
    list.value = res?.data || []
    updatePagination(
      res?.total || 0,
      !!res?.hasNext,
      !!res?.hasPrev,
      res?.nextCursor || null,
      res?.prevCursor || null,
    )
  }).catch(() => {
    ElMessage.error('加载失败')
  })
}

const handleQuery = () => {
  pagination.cursor = null
  getList()
}

const resetQuery = () => {
  resetQueryParams()
  list.value = []
  resetPagination()
}

const handleLimitChange = () => {
  pagination.cursor = null
  getList()
}

const refreshCurrentPage = () => {
  getList()
}

const handlePrevPage = () => {
  if (pagination.hasPrev && pagination.prevCursor) {
    pagination.cursor = pagination.prevCursor
    getList()
  }
}

const handleNextPage = () => {
  if (pagination.hasNext && pagination.nextCursor) {
    pagination.cursor = pagination.nextCursor
    getList()
  }
}

const handleAdd = async () => {
  formMode.value = 'add'
  resetForm()
  form.tenantId = Number(queryParams.tenantId) || undefined
  formDialogVisible.value = true
  await nextTick()
  formRef.value?.clearValidate()
}

const handleEdit = async (row: ItickTenantCategory) => {
  formMode.value = 'edit'
  resetForm()
  try {
    const res = await tenantCategoriesService.detail(row.id, row.tenantId)
    const data = res?.data || row
    Object.assign(form, {
      id: data.id,
      tenantId: data.tenantId,
      categoryId: data.categoryId,
      enabled: data.enabled,
      appVisible: data.appVisible,
      sort: data.sort || 0,
      remark: data.remark || '',
    })
    formDialogVisible.value = true
    await nextTick()
    formRef.value?.clearValidate()
  } catch {
    ElMessage.error('加载详情失败')
  }
}

const handleDetail = async (row: ItickTenantCategory) => {
  detailDialogVisible.value = true
  detailLoading.value = true
  detail.value = {}
  try {
    const res = await tenantCategoriesService.detail(row.id, row.tenantId)
    detail.value = res?.data || {}
  } catch {
    ElMessage.error('加载详情失败')
  } finally {
    detailLoading.value = false
  }
}

const submitForm = async () => {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    if (formMode.value === 'add') {
      await tenantCategoriesService.create({
        tenantId: Number(form.tenantId),
        categoryId: Number(form.categoryId),
        enabled: form.enabled,
        appVisible: form.appVisible,
        sort: form.sort,
        remark: form.remark,
      })
      ElMessage.success('创建成功')
    } else {
      await tenantCategoriesService.update(form.id as number, {
        tenantId: Number(form.tenantId),
        enabled: form.enabled,
        appVisible: form.appVisible,
        sort: form.sort,
        remark: form.remark,
      })
      ElMessage.success('更新成功')
    }

    formDialogVisible.value = false
    getList()
  } catch {
    ElMessage.error(formMode.value === 'add' ? '创建失败' : '更新失败')
  } finally {
    submitLoading.value = false
  }
}

const openBatchDialog = () => {
  batchRows.value = list.value.map((item) => ({
    id: item.id,
    categoryId: item.categoryId,
    enabled: item.enabled,
    appVisible: item.appVisible,
    sort: item.sort,
    remark: item.remark || '',
  }))
  batchDialogVisible.value = true
}

const appendBatchRow = () => {
  batchRows.value.push({
    categoryId: 0,
    enabled: 1,
    appVisible: 1,
    sort: 0,
    remark: '',
  })
}

const removeBatchRow = (index: number) => {
  batchRows.value.splice(index, 1)
}

const submitBatch = async () => {
  const tenantId = Number(queryParams.tenantId)
  if (!tenantId) {
    ElMessage.warning('请先输入租户ID')
    return
  }

  const cleaned = batchRows.value
    .filter((item) => item.categoryId && Number(item.categoryId) > 0)
    .map((item) => ({
      id: item.id,
      categoryId: Number(item.categoryId),
      enabled: item.enabled,
      appVisible: item.appVisible,
      sort: Number(item.sort || 0),
      remark: item.remark || '',
    }))

  batchSubmitting.value = true
  try {
    await tenantCategoriesService.batchUpsert({
      tenantId,
      data: cleaned,
    })
    ElMessage.success('批量保存成功')
    batchDialogVisible.value = false
    getList()
  } catch {
    ElMessage.error('批量保存失败')
  } finally {
    batchSubmitting.value = false
  }
}

onMounted(() => {
  if (queryParams.tenantId) {
    getList()
  }
})
</script>

<style scoped>
.itick-tenant-categories {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.page-header h2 {
  margin: 0;
  color: #333;
}

.header-actions {
  display: flex;
  gap: 10px;
  align-items: center;
}

.query-card,
.table-card {
  margin-bottom: 16px;
}

.pagination-bar {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  align-items: center;
  margin-top: 16px;
}

.batch-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.batch-tip {
  color: #909399;
  font-size: 13px;
}

.batch-actions {
  display: flex;
  gap: 8px;
}
</style>
