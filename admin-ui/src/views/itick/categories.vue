<template>
  <div class="itick-categories">
    <div class="page-header">
      <h2>{{ t('itick.categories') }}</h2>
      <div class="header-actions">
        <el-button type="primary" @click="handleAdd">
          <el-icon><Plus /></el-icon>
          {{ t('common.add') }}
        </el-button>
        <el-button @click="refreshCurrentPage">
          <el-icon><Refresh /></el-icon>
          {{ t('common.refresh') }}
        </el-button>
      </div>
    </div>

    <el-card class="query-card" shadow="never">
      <el-form :model="queryParams" inline label-width="90px">
        <el-form-item label="分类类型">
          <el-input
            v-model="queryParams.categoryType"
            placeholder="请输入分类类型"
            clearable
            style="width: 180px"
            @keyup.enter="handleQuery"
          />
        </el-form-item>

        <el-form-item label="启用状态">
          <el-select
            v-model="queryParams.enabled"
            placeholder="请选择"
            clearable
            style="width: 180px"
          >
            <el-option label="全部" :value="0" />
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="2" />
          </el-select>
        </el-form-item>

        <el-form-item label="APP显示">
          <el-select
            v-model="queryParams.appVisible"
            placeholder="请选择"
            clearable
            style="width: 180px"
          >
            <el-option label="全部" :value="0" />
            <el-option label="显示" :value="1" />
            <el-option label="隐藏" :value="2" />
          </el-select>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleQuery">搜索</el-button>
          <el-button @click="resetQuery">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="table-card" shadow="never">
      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="categoryType" label="分类类型" width="100" />
        <el-table-column prop="categoryCode" label="分类编码" min-width="140" />
        <el-table-column prop="categoryName" label="分类名称" min-width="160" />

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
          min-width="180"
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
            <el-button link type="primary" @click="handleDetail(row)">详情</el-button>
            <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="warning" @click="handleSync(row)">同步产品</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-bar">
        <span>{{ t('common.totalItems', { count: pagination.total }) }}</span>
        <el-button :disabled="!pagination.hasPrev" @click="handlePrevPage">
          {{ t('common.prevPage') }}
        </el-button>
        <el-button :disabled="!pagination.hasNext" type="primary" @click="handleNextPage">
          {{ t('common.nextPage') }}
        </el-button>
        <el-select v-model="pagination.limit" style="width: 100px" @change="handleLimitChange">
          <el-option :value="10" :label="t('common.perPage10')" />
          <el-option :value="20" :label="t('common.perPage20')" />
          <el-option :value="50" :label="t('common.perPage50')" />
          <el-option :value="100" :label="t('common.perPage100')" />
        </el-select>
      </div>
    </el-card>

    <el-dialog
      v-model="formDialogVisible"
      :title="formMode === 'add' ? '新增分类' : '编辑分类'"
      width="620px"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item v-if="formMode === 'add'" label="分类类型" prop="categoryType">
          <el-input-number
            v-model="form.categoryType"
            :min="1"
            :precision="0"
            controls-position="right"
            style="width: 100%"
          />
        </el-form-item>

        <el-form-item label="分类名称" prop="categoryName">
          <el-input
            v-model="form.categoryName"
            placeholder="请输入分类名称"
            maxlength="50"
            show-word-limit
          />
        </el-form-item>

        <el-form-item label="启用状态" prop="enabled">
          <el-radio-group v-model="form.enabled">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="2">禁用</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="APP显示" prop="appVisible">
          <el-radio-group v-model="form.appVisible">
            <el-radio :value="1">显示</el-radio>
            <el-radio :value="2">隐藏</el-radio>
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

        <el-form-item label="图标" prop="icon">
          <el-input v-model="form.icon" placeholder="请输入图标" />
        </el-form-item>

        <el-form-item label="备注" prop="remark">
          <el-input
            v-model="form.remark"
            type="textarea"
            :rows="4"
            placeholder="请输入备注"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="formDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitForm"> 确定 </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailDialogVisible" title="分类详情" width="700px">
      <el-descriptions :column="2" border v-loading="detailLoading">
        <el-descriptions-item label="ID">{{ detail.id ?? '-' }}</el-descriptions-item>
        <el-descriptions-item label="分类类型">{{
          detail.categoryType ?? '-'
        }}</el-descriptions-item>
        <el-descriptions-item label="分类编码">{{
          detail.categoryCode || '-'
        }}</el-descriptions-item>
        <el-descriptions-item label="分类名称">{{
          detail.categoryName || '-'
        }}</el-descriptions-item>
        <el-descriptions-item label="启用状态">
          {{ detail.enabled === 1 ? '启用' : detail.enabled === 2 ? '禁用' : '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="APP显示">
          {{ detail.appVisible === 1 ? '显示' : detail.appVisible === 2 ? '隐藏' : '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="排序">{{ detail.sort ?? '-' }}</el-descriptions-item>
        <el-descriptions-item label="图标">{{ detail.icon || '-' }}</el-descriptions-item>
        <el-descriptions-item label="备注" :span="2">{{
          detail.remark || '-'
        }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">
          {{ formatDate(detail?.createTimes??0) }}
        </el-descriptions-item>
        <el-descriptions-item label="更新时间">
          {{ formatDate(detail?.updateTimes??0) }}
        </el-descriptions-item>
      </el-descriptions>

      <template #footer>
        <el-button type="primary" @click="detailDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, onMounted } from 'vue'
import { ElMessage, type FormRules } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables/usePagination'
import { useLoading } from '@/composables/useLoading'
import { useForm } from '@/composables/useForm'
import {
  categoriesService,
  type ItickCategory,
  type ListCategoriesReq,
} from '@/services/itick/CategoriesService'
import { formatDate } from '@/utils'

type FormData = {
  id?: number
  categoryType?: number
  categoryName: string
  enabled: number
  appVisible: number
  sort: number
  icon: string
  remark: string
}

const { t } = useI18n()
const { pagination, updatePagination, reset: resetPagination } = usePagination(20)
const { loading, withLoading } = useLoading()

const { form: queryParams, reset: resetQueryParams } = useForm<ListCategoriesReq>({
  initialData: {
    categoryType: undefined,
    enabled: 0,
    appVisible: 0,
    cursor: null,
    limit: 20,
  },
})

const {
  form: form,
  formRef,
  reset: resetForm,
} = useForm<FormData>({
  initialData: {
    id: undefined,
    categoryType: undefined,
    categoryName: '',
    enabled: 1,
    appVisible: 1,
    sort: 0,
    icon: '',
    remark: '',
  },
})

const submitLoading = ref(false)
const detailLoading = ref(false)
const list = ref<ItickCategory[]>([])
const detail = ref<Partial<ItickCategory>>({})
const formDialogVisible = ref(false)
const detailDialogVisible = ref(false)
const formMode = ref<'add' | 'edit'>('add')

const rules: FormRules<FormData> = {
  categoryType: [{ required: true, message: '请输入分类类型', trigger: 'blur' }],
  categoryName: [{ required: true, message: '请输入分类名称', trigger: 'blur' }],
  enabled: [{ required: true, message: '请选择启用状态', trigger: 'change' }],
  appVisible: [{ required: true, message: '请选择APP显示状态', trigger: 'change' }],
  sort: [{ required: true, message: '请输入排序', trigger: 'blur' }],
}

const cleanedQueryParams = computed<ListCategoriesReq>(() => {
  const params: ListCategoriesReq = {
    cursor: queryParams.cursor,
    limit: queryParams.limit,
  }

  if (queryParams.categoryType && queryParams.categoryType !== 0) {
    params.categoryType = Number(queryParams.categoryType)
  }
  if (queryParams.enabled && queryParams.enabled !== 0) {
    params.enabled = queryParams.enabled
  }
  if (queryParams.appVisible && queryParams.appVisible !== 0) {
    params.appVisible = queryParams.appVisible
  }

  return params
})

const getList = async () => {
  await withLoading(async () => {
    try {
      const res = await categoriesService.getList({
        ...cleanedQueryParams.value,
        cursor: pagination.cursor,
      })
      list.value = res?.data || []
      updatePagination(
        res?.total || 0,
        !!res?.hasNext,
        !!res?.hasPrev,
        res?.nextCursor || null,
        res?.prevCursor || null,
      )
    } catch (error) {
      ElMessage.error(t('common.loadFailed'))
    }
  })
}

const handleQuery = () => {
  pagination.cursor = null
  getList()
}

const resetQuery = () => {
  resetQueryParams()
  resetPagination()
  getList()
}

const handleLimitChange = () => {
  pagination.cursor = null
  getList()
}

const refreshCurrentPage = () => {
  getList()
}

const handleAdd = async () => {
  formMode.value = 'add'
  resetForm()
  formDialogVisible.value = true
  await nextTick()
  formRef.value?.clearValidate()
}

const handleEdit = async (row: ItickCategory) => {
  formMode.value = 'edit'
  resetForm()

  try {
    const res = await categoriesService.detail(row.id)
    const data = res?.data || row

    Object.assign(form, {
      id: data.id,
      categoryType: data.categoryType,
      categoryName: data.categoryName || '',
      enabled: data.enabled,
      appVisible: data.appVisible,
      sort: data.sort || 0,
      icon: data.icon || '',
      remark: data.remark || '',
    })

    formDialogVisible.value = true
    await nextTick()
    formRef.value?.clearValidate()
  } catch (error) {
    ElMessage.error(t('common.loadFailed'))
  }
}

const handleDetail = async (row: ItickCategory) => {
  detailDialogVisible.value = true
  detailLoading.value = true
  detail.value = {}

  try {
    const res = await categoriesService.detail(row.id)
    detail.value = res?.data || {}
  } catch (error) {
    ElMessage.error(t('common.loadFailed'))
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
      await categoriesService.create({
        categoryType: Number(form.categoryType),
        categoryName: form.categoryName,
        enabled: form.enabled,
        appVisible: form.appVisible,
        sort: form.sort,
        icon: form.icon,
        remark: form.remark,
      })
      ElMessage.success(t('common.createSuccess'))
    } else {
      await categoriesService.update(form.id as number, {
        categoryName: form.categoryName,
        enabled: form.enabled,
        appVisible: form.appVisible,
        sort: form.sort,
        icon: form.icon,
        remark: form.remark,
      })
      ElMessage.success(t('common.updateSuccess'))
    }

    formDialogVisible.value = false
    getList()
  } catch (error) {
    ElMessage.error(formMode.value === 'add' ? t('common.createFailed') : t('common.updateFailed'))
  } finally {
    submitLoading.value = false
  }
}

const handleSync = async (row: ItickCategory) => {
  try {
    const res = await categoriesService.syncProducts({ id: row.id })
    ElMessage.success(`同步任务已提交，任务号：${res?.data?.taskNo || '-'}`)
  } catch {
    ElMessage.error(t('common.operationFailed'))
  }
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

onMounted(() => {
  getList()
})
</script>

<style scoped>
.itick-categories {
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
  border-radius: 12px;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.cursor-pagination {
  margin-top: 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}

.pagination-bar {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  align-items: center;
  margin-top: 12px;
}
</style>