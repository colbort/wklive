<template>
  <div class="itick-tenant-categories">
    <div class="page-header">
      <h2>{{ t('itick.tenantCategories') }}</h2>
      <div class="header-actions">
        <el-button type="primary" :disabled="!queryParams.tenantId" @click="handleAdd">
          <el-icon><Plus /></el-icon>
          {{ t('common.add') }}
        </el-button>
        <el-button :disabled="!queryParams.tenantId" @click="openBatchDialog">
          <el-icon><EditPen /></el-icon>
          {{ t('itick.batchTenantCategories') }}
        </el-button>
        <el-button @click="refreshCurrentPage">
          <el-icon><Refresh /></el-icon>
          {{ t('common.refresh') }}
        </el-button>
      </div>
    </div>

    <el-card class="query-card" shadow="never">
      <el-form :model="queryParams" inline label-width="90px">
        <el-form-item :label="t('common.tenantId')">
          <el-input-number
            v-model="queryParams.tenantId"
            :min="1"
            :precision="0"
            controls-position="right"
            style="width: 180px"
          />
        </el-form-item>

        <el-form-item :label="t('itick.categoryType')">
          <el-select v-model="queryParams.categoryType" clearable style="width: 180px">
            <el-option
              v-for="item in categoryTypeOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>

        <el-form-item :label="t('itick.enabledStatus')">
          <el-select v-model="queryParams.status" clearable style="width: 180px">
            <el-option
              v-for="item in statusOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>

        <el-form-item :label="t('itick.appVisible')">
          <el-select v-model="queryParams.visibleStatus" clearable style="width: 180px">
            <el-option
              v-for="item in visibleOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleQuery"> {{ t('common.search') }} </el-button>
          <el-button @click="resetQuery"> {{ t('common.reset') }} </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="table-card" shadow="never">
      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="tenantId" :label="t('common.tenantId')" width="100" />
        <el-table-column prop="categoryId" :label="t('itick.category')" width="100" />
        <el-table-column :label="t('itick.categoryType')" width="120">
          <template #default="{ row }">
            {{ getOptionValueLabel(optionGroups, 'categoryType', row.categoryType, t) }}
          </template>
        </el-table-column>
        <el-table-column prop="categoryName" :label="t('itick.categoryName')" min-width="180" />

        <el-table-column :label="t('itick.enabledStatus')" width="100">
          <template #default="{ row }">
            <el-tag :type="row.enabled === 1 ? 'success' : 'info'">
              {{ getOptionValueLabel(optionGroups, 'status', row.enabled, t) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column :label="t('itick.appVisible')" width="100">
          <template #default="{ row }">
            <el-tag :type="row.appVisible === 1 ? 'success' : 'warning'">
              {{ getOptionValueLabel(optionGroups, 'visible', row.appVisible, t) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="sort" :label="t('common.sort')" width="90" />
        <el-table-column :label="t('common.icon')" min-width="180">
          <template #default="{ row }">
            <div v-if="row.icon" class="icon-cell">
              <el-image
                :src="resolveAssetUrl(row.icon)"
                class="icon-preview"
                :preview-teleported="true"
              />
              <span class="icon-url">{{ row.icon }}</span>
            </div>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="remark" :label="t('common.remark')" min-width="180" show-overflow-tooltip />
        <el-table-column :label="t('common.createTimes')" min-width="170">
          <template #default="{ row }">
            {{ formatDate(row.createTimes) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('itick.updateTimes')" min-width="170">
          <template #default="{ row }">
            {{ formatDate(row.updateTimes) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="180" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleDetail(row)"> {{ t('common.detail') }} </el-button>
            <el-button link type="primary" @click="handleEdit(row)"> {{ t('common.edit') }} </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-bar">
        <span>{{ t('common.totalItems', { count: pagination.total }) }}</span>
        <el-button :disabled="!pagination.hasPrev" @click="handlePrevPage"> {{ t('common.prevPage') }} </el-button>
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
      :title="formMode === 'add' ? t('itick.addTenantCategory') : t('itick.editTenantCategory')"
      width="620px"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
        <el-form-item :label="t('common.tenantId')" prop="tenantId">
          <el-input-number
            v-model="form.tenantId"
            :min="1"
            :precision="0"
            controls-position="right"
            style="width: 100%"
            :disabled="formMode === 'edit'"
          />
        </el-form-item>

        <el-form-item v-if="formMode === 'add'" :label="t('itick.category')" prop="categoryId">
          <el-select
            v-model="form.categoryId"
            filterable
            clearable
            :placeholder="t('itick.pleaseSelectCategory')"
            style="width: 100%"
          >
            <el-option
              v-for="item in categoryOptions"
              :key="item.id"
              :label="`${item.id} - ${item.categoryName}`"
              :value="item.id"
            />
          </el-select>
        </el-form-item>

        <el-form-item :label="t('itick.enabledStatus')" prop="enabled">
          <el-radio-group v-model="form.enabled">
            <el-radio v-for="item in statusOptions" :key="item.value" :value="item.value">
              {{ getOptionLabel(t, item.code, item.value) }}
            </el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item :label="t('itick.appVisible')" prop="appVisible">
          <el-radio-group v-model="form.appVisible">
            <el-radio v-for="item in visibleOptions" :key="item.value" :value="item.value">
              {{ getOptionLabel(t, item.code, item.value) }}
            </el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item :label="t('common.sort')" prop="sort">
          <el-input-number
            v-model="form.sort"
            :min="0"
            :precision="0"
            controls-position="right"
            style="width: 100%"
          />
        </el-form-item>

        <el-form-item :label="t('common.remark')" prop="remark">
          <el-input
            v-model="form.remark"
            type="textarea"
            :rows="4"
            maxlength="200"
            show-word-limit
            :placeholder="t('common.remark')"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="formDialogVisible = false"> {{ t('common.cancel') }} </el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitForm"> {{ t('common.confirm') }} </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailDialogVisible" :title="t('itick.tenantCategoryDetail')" width="700px">
      <el-descriptions v-loading="detailLoading" :column="2" border>
        <el-descriptions-item label="ID">
          {{ detail.id ?? '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.tenantId')">
          {{ detail.tenantId ?? '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('itick.category')">
          {{ detail.categoryId ?? '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('itick.categoryType')">
          {{ getOptionValueLabel(optionGroups, 'categoryType', detail.categoryType, t) || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('itick.categoryName')">
          {{ detail.categoryName || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.icon')">
          <div v-if="detail.icon" class="icon-detail">
            <el-image
              :src="resolveAssetUrl(detail.icon)"
              class="icon-preview-large"
              :preview-teleported="true"
            />
            <div class="icon-url">{{ detail.icon }}</div>
          </div>
          <span v-else>-</span>
        </el-descriptions-item>
        <el-descriptions-item :label="t('itick.enabledStatus')">
          {{ getOptionValueLabel(optionGroups, 'status', detail.enabled, t) || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('itick.appVisible')">
          {{ getOptionValueLabel(optionGroups, 'visible', detail.appVisible, t) || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.sort')">
          {{ detail.sort ?? '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.remark')">
          {{ detail.remark || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.createTimes')">
          {{ formatDate(detail.createTimes ?? 0) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('itick.updateTimes')">
          {{ formatDate(detail.updateTimes ?? 0) }}
        </el-descriptions-item>
      </el-descriptions>

      <template #footer>
        <el-button type="primary" @click="detailDialogVisible = false"> {{ t('common.close') }} </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="batchDialogVisible" :title="t('itick.batchTenantCategories')" width="920px">
      <div class="batch-toolbar">
        <div class="batch-tip">{{ t('itick.batchSaveTip') }}</div>
        <div class="batch-actions">
          <el-button @click="appendBatchRow"> {{ t('common.add') }} </el-button>
          <el-button type="primary" :loading="batchSubmitting" @click="submitBatch">
            {{ t('common.save') }}
          </el-button>
        </div>
      </div>

      <el-table :data="batchRows" border>
        <el-table-column :label="t('itick.category')" min-width="220">
          <template #default="{ row }">
            <el-select
              v-model="row.categoryId"
              filterable
              clearable
              :placeholder="t('itick.pleaseSelectCategory')"
              style="width: 100%"
            >
              <el-option
                v-for="item in categoryOptions"
                :key="item.id"
                :label="`${item.id} - ${item.categoryName}`"
                :value="item.id"
              />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column :label="t('itick.enabledStatus')" min-width="130">
          <template #default="{ row }">
            <el-select v-model="row.enabled" style="width: 100%">
              <el-option
                v-for="item in statusOptions"
                :key="item.value"
                :label="getOptionLabel(t, item.code, item.value)"
                :value="item.value"
              />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column :label="t('itick.appVisible')" min-width="130">
          <template #default="{ row }">
            <el-select v-model="row.appVisible" style="width: 100%">
              <el-option
                v-for="item in visibleOptions"
                :key="item.value"
                :label="getOptionLabel(t, item.code, item.value)"
                :value="item.value"
              />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.sort')" min-width="120">
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
        <el-table-column :label="t('common.remark')" min-width="220">
          <template #default="{ row }">
            <el-input v-model="row.remark" maxlength="200" show-word-limit />
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="90" fixed="right">
          <template #default="{ $index }">
            <el-button link type="danger" @click="removeBatchRow($index)"> {{ t('common.delete') }} </el-button>
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
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables/usePagination'
import { useLoading } from '@/composables/useLoading'
import { useForm } from '@/composables/useForm'
import { buildSystemAssetUrl, useSystemCore } from '@/composables/useSystemCore'
import type { OptionGroup } from '@/services'
import { categoriesService, type ItickCategory } from '@/services/itick/CategoriesService'
import {
  tenantCategoriesService,
  type ItickTenantCategory,
  type ListTenantCategoriesReq,
  type TenantCategoryItem,
} from '@/services/itick/TenantCategoriesService'
import { formatDate } from '@/utils'
import { findOptionGroup, getOptionLabel, getOptionValueLabel } from '@/utils/options'

type FormData = {
  id?: number
  tenantId?: number
  categoryId?: number
  enabled: number
  appVisible: number
  sort: number
  remark: string
}

const { t } = useI18n()
const { systemCore, loadSystemCore } = useSystemCore()
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

const {
  form,
  formRef,
  reset: resetForm,
} = useForm<FormData>({
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
const categoryOptions = ref<ItickCategory[]>([])
const optionGroups = ref<OptionGroup[]>([])
const detailLoading = ref(false)
const submitLoading = ref(false)
const batchSubmitting = ref(false)
const formDialogVisible = ref(false)
const detailDialogVisible = ref(false)
const batchDialogVisible = ref(false)
const formMode = ref<'add' | 'edit'>('add')
const batchRows = ref<TenantCategoryItem[]>([])
const categoryTypeOptions = computed(() => findOptionGroup(optionGroups.value, 'categoryType'))
const statusOptions = computed(() => findOptionGroup(optionGroups.value, 'status'))
const visibleOptions = computed(() => findOptionGroup(optionGroups.value, 'visible'))
const resolveAssetUrl = (url?: string) => buildSystemAssetUrl(systemCore.value.assetUrl, url)

const rules: FormRules<FormData> = {
  tenantId: [{ required: true, message: t('itick.pleaseInputTenantId'), trigger: 'blur' }],
  categoryId: [{ required: true, message: t('itick.pleaseSelectCategory'), trigger: 'change' }],
  enabled: [{ required: true, message: t('itick.pleaseSelectEnabledStatus'), trigger: 'change' }],
  appVisible: [{ required: true, message: t('itick.pleaseSelectAppVisible'), trigger: 'change' }],
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
    const res = await tenantCategoriesService.getList(
      cleanedQueryParams.value as ListTenantCategoriesReq,
    )
    list.value = res?.data || []
    updatePagination(
      res?.total || 0,
      !!res?.hasNext,
      !!res?.hasPrev,
      res?.nextCursor || null,
      res?.prevCursor || null,
    )
  }).catch(() => {
    ElMessage.error(t('itick.loadFailed'))
  })
}

const loadOptions = async () => {
  try {
    const [optionsRes, categoriesRes] = await Promise.all([
      tenantCategoriesService.getOptions(),
      categoriesService.getList({ limit: 100 }),
    ])
    optionGroups.value = optionsRes.data || []
    categoryOptions.value = categoriesRes.data || []
  } catch {
    ElMessage.error(t('itick.loadOptionsFailed'))
  }
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
    ElMessage.error(t('itick.loadDetailFailed'))
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
    ElMessage.error(t('itick.loadDetailFailed'))
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
      ElMessage.success(t('itick.createdSuccess'))
    } else {
      await tenantCategoriesService.update(form.id as number, {
        tenantId: Number(form.tenantId),
        enabled: form.enabled,
        appVisible: form.appVisible,
        sort: form.sort,
        remark: form.remark,
      })
      ElMessage.success(t('itick.updatedSuccess'))
    }

    formDialogVisible.value = false
    getList()
  } catch {
    ElMessage.error(t(formMode.value === 'add' ? 'itick.createdFailed' : 'itick.updatedFailed'))
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
    ElMessage.warning(t('itick.pleaseInputTenantFirst'))
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
    ElMessage.success(t('itick.batchSavedSuccess'))
    batchDialogVisible.value = false
    getList()
  } catch {
    ElMessage.error(t('itick.batchSavedFailed'))
  } finally {
    batchSubmitting.value = false
  }
}

onMounted(() => {
  loadSystemCore()
  loadOptions()
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

.icon-cell,
.icon-detail {
  display: flex;
  align-items: center;
  gap: 12px;
}

.icon-preview {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  flex-shrink: 0;
}

.icon-preview-large {
  width: 64px;
  height: 64px;
  border-radius: 12px;
  flex-shrink: 0;
}

.icon-url {
  color: #606266;
  word-break: break-all;
}
</style>
