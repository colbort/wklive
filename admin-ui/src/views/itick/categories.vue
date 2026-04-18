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
        <el-form-item :label="t('itick.categoryType')">
          <el-select
            v-model="queryParams.categoryType"
            :placeholder="t('common.pleaseSelect')"
            clearable
            style="width: 180px"
          >
            <el-option
              v-for="item in categoryTypeOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>

        <el-form-item :label="t('itick.enabledStatus')">
          <el-select
            v-model="queryParams.enabled"
            :placeholder="t('common.pleaseSelect')"
            clearable
            style="width: 180px"
          >
            <el-option
              v-for="item in statusOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>

        <el-form-item :label="t('itick.appVisible')">
          <el-select
            v-model="queryParams.appVisible"
            :placeholder="t('common.pleaseSelect')"
            clearable
            style="width: 180px"
          >
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
        <el-table-column :label="t('itick.categoryType')" width="120">
          <template #default="{ row }">
            {{ getOptionValueLabel(optionGroups, 'categoryType', row.categoryType, t) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('itick.categoryCode')" prop="categoryCode" min-width="140" />
        <el-table-column :label="t('itick.categoryName')" prop="categoryName" min-width="160" />

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

        <el-table-column :label="t('common.sort')" prop="sort" width="90" />
        <el-table-column :label="t('common.icon')" min-width="180">
          <template #default="{ row }">
            <div v-if="row.icon" class="icon-cell">
              <el-image
                :src="resolveAssetUrl(row.icon)"
                class="icon-preview"
                :preview-teleported="true"
              />
            </div>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.remark')" prop="remark" min-width="180" show-overflow-tooltip />

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
            <el-button link type="primary" @click="handleDetail(row)"> {{ t('itick.detail') }} </el-button>
            <el-button link type="primary" @click="handleEdit(row)"> {{ t('common.edit') }} </el-button>
            <el-button link type="warning" @click="handleSync(row)"> {{ t('itick.syncProducts') }} </el-button>
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
      :title="formMode === 'add' ? t('itick.addCategory') : t('itick.editCategory')"
      width="620px"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item v-if="formMode === 'add'" :label="t('itick.categoryType')" prop="categoryType">
          <el-select v-model="form.categoryType" :placeholder="t('common.pleaseSelect')" style="width: 100%">
            <el-option
              v-for="item in categoryTypeOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>

        <el-form-item :label="t('itick.categoryName')" prop="categoryName">
          <el-input
            v-model="form.categoryName"
            :placeholder="t('itick.pleaseInputCategoryName')"
            maxlength="50"
            show-word-limit
          />
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

        <el-form-item :label="t('common.icon')" prop="icon">
          <div class="icon-upload-field">
            <div v-if="form.icon" class="icon-upload-preview">
              <el-image
                :src="resolveAssetUrl(form.icon)"
                class="icon-preview-large"
                :preview-teleported="true"
              />
              <div class="icon-url">{{ form.icon }}</div>
            </div>
            <el-upload
              action="#"
              :auto-upload="false"
              :show-file-list="false"
              :on-change="handleIconSelect"
              accept="image/*"
            >
              <el-button type="primary" :loading="submitLoading"> {{ t('itick.uploadImage') }} </el-button>
            </el-upload>
          </div>
        </el-form-item>

        <el-form-item :label="t('common.remark')" prop="remark">
          <el-input
            v-model="form.remark"
            type="textarea"
            :rows="4"
            :placeholder="t('common.remark')"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="formDialogVisible = false"> {{ t('common.cancel') }} </el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitForm"> {{ t('common.confirm') }} </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailDialogVisible" :title="t('itick.categoryDetail')" width="700px">
      <el-descriptions v-loading="detailLoading" :column="2" border>
        <el-descriptions-item label="ID">
          {{ detail.id ?? '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('itick.categoryType')">
          {{ getOptionValueLabel(optionGroups, 'categoryType', detail.categoryType, t) || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('itick.categoryCode')">
          {{ detail.categoryCode || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('itick.categoryName')">
          {{ detail.categoryName || '-' }}
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
        <el-descriptions-item :label="t('common.remark')" :span="2">
          {{ detail.remark || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.createTimes')">
          {{ formatDate(detail?.createTimes ?? 0) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('itick.updateTimes')">
          {{ formatDate(detail?.updateTimes ?? 0) }}
        </el-descriptions-item>
      </el-descriptions>

      <template #footer>
        <el-button type="primary" @click="detailDialogVisible = false"> {{ t('common.close') }} </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue'
import { ElMessage, type FormRules, type UploadFile } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables/usePagination'
import { useLoading } from '@/composables/useLoading'
import { useForm } from '@/composables/useForm'
import { buildSystemAssetUrl, useSystemCore } from '@/composables/useSystemCore'
import type { OptionGroup } from '@/services'
import { apiUploadFile } from '@/api/system/upload'
import {
  categoriesService,
  type ItickCategory,
  type ListCategoriesReq,
} from '@/services/itick/CategoriesService'
import { formatDate } from '@/utils'
import { findOptionGroup, getOptionLabel, getOptionValueLabel } from '@/utils/options'

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
const { systemCore, loadSystemCore } = useSystemCore()
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
const optionGroups = ref<OptionGroup[]>([])
const formDialogVisible = ref(false)
const detailDialogVisible = ref(false)
const formMode = ref<'add' | 'edit'>('add')
const categoryTypeOptions = computed(() => findOptionGroup(optionGroups.value, 'categoryType'))
const statusOptions = computed(() => findOptionGroup(optionGroups.value, 'status'))
const visibleOptions = computed(() => findOptionGroup(optionGroups.value, 'visible'))
const resolveAssetUrl = (url?: string) => buildSystemAssetUrl(systemCore.value.assetUrl, url)

const rules: FormRules<FormData> = {
  categoryType: [{ required: true, message: t('itick.pleaseInputCategoryType'), trigger: 'blur' }],
  categoryName: [{ required: true, message: t('itick.pleaseInputCategoryName'), trigger: 'blur' }],
  enabled: [{ required: true, message: t('itick.pleaseSelectEnabledStatus'), trigger: 'change' }],
  appVisible: [{ required: true, message: t('itick.pleaseSelectAppVisible'), trigger: 'change' }],
  sort: [{ required: true, message: t('itick.pleaseInputSort'), trigger: 'blur' }],
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

const loadOptions = async () => {
  try {
    const res = await categoriesService.getOptions()
    optionGroups.value = res.data || []
  } catch {
    ElMessage.error(t('common.loadFailed'))
  }
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

const handleIconSelect = async (uploadFile: UploadFile) => {
  if (!uploadFile.raw) return

  if (!uploadFile.raw.type.startsWith('image/')) {
    ElMessage.error(t('app.pleaseSelectImageFile'))
    return
  }

  if (uploadFile.raw.size > 5 * 1024 * 1024) {
    ElMessage.error(t('app.avatarSizeLimit'))
    return
  }

  submitLoading.value = true
  try {
    const res = await apiUploadFile(uploadFile.raw)
    if (res.code === 0 || res.code === 200) {
      form.icon = res.data?.url || ''
      ElMessage.success(t('common.uploadSuccess'))
      return
    }
    throw new Error(res.msg || t('common.uploadFailed'))
  } catch (error: unknown) {
    ElMessage.error(error instanceof Error ? error.message : t('common.uploadFailed'))
  } finally {
    submitLoading.value = false
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
    ElMessage.success(
      t('itick.syncTaskSubmittedWithTaskNo', { taskNo: res?.data?.taskNo || '-' }),
    )
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
  loadSystemCore()
  loadOptions()
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

.icon-cell,
.icon-upload-field,
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
