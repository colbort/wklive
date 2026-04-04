<template>
  <div class="itick-products">
    <div class="page-header">
      <h2>{{ t('itick.products') }}</h2>
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
          <el-input
            v-model="queryParams.categoryType"
            :placeholder="t('itick.pleaseInputCategoryType')"
            clearable
            style="width: 180px"
            @keyup.enter="handleQuery"
          />
        </el-form-item>

        <el-form-item :label="t('itick.market')">
          <el-input
            v-model="queryParams.market"
            :placeholder="t('itick.pleaseInputMarket')"
            clearable
            style="width: 180px"
            @keyup.enter="handleQuery"
          />
        </el-form-item>

        <el-form-item :label="t('itick.keyword')">
          <el-input
            v-model="queryParams.keyword"
            :placeholder="t('common.keyword')"
            clearable
            style="width: 180px"
            @keyup.enter="handleQuery"
          />
        </el-form-item>

        <el-form-item :label="t('itick.enabledStatus')">
          <el-select
            v-model="queryParams.enabled"
            :placeholder="t('common.pleaseSelect')"
            clearable
            style="width: 180px"
          >
            <el-option :label="t('common.all')" :value="0" />
            <el-option :label="t('common.enabled')" :value="1" />
            <el-option :label="t('common.disabled')" :value="2" />
          </el-select>
        </el-form-item>

        <el-form-item :label="t('itick.appVisible')">
          <el-select
            v-model="queryParams.appVisible"
            :placeholder="t('common.pleaseSelect')"
            clearable
            style="width: 180px"
          >
            <el-option :label="t('common.all')" :value="0" />
            <el-option :label="t('itick.show')" :value="1" />
            <el-option :label="t('itick.hide')" :value="2" />
          </el-select>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleQuery">{{ t('common.search') }}</el-button>
          <el-button @click="resetQuery">{{ t('common.reset') }}</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="table-card" shadow="never">
      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" :label="t('common.id')" width="80" />
        <el-table-column prop="categoryType" :label="t('itick.categoryType')" width="100" />
        <el-table-column prop="categoryName" :label="t('itick.categoryName')" min-width="140" />
        <el-table-column prop="categoryCode" :label="t('itick.categoryCode')" min-width="140" />
        <el-table-column prop="market" :label="t('itick.market')" width="100" />
        <el-table-column prop="symbol" :label="t('itick.symbol')" min-width="120" />
        <el-table-column prop="code" :label="t('itick.code')" min-width="120" />
        <el-table-column prop="name" :label="t('itick.name')" min-width="140" />
        <el-table-column prop="displayName" :label="t('itick.displayName')" min-width="140" />
        <el-table-column prop="baseCoin" :label="t('itick.baseCoin')" width="100" />
        <el-table-column prop="quoteCoin" :label="t('itick.quoteCoin')" width="100" />

        <el-table-column :label="t('itick.enabledStatus')" width="100">
          <template #default="{ row }">
            <el-tag :type="row.enabled === 1 ? 'success' : 'info'">
              {{ row.enabled === 1 ? t('common.enabled') : t('common.disabled') }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column :label="t('itick.appVisible')" width="100">
          <template #default="{ row }">
            <el-tag :type="row.appVisible === 1 ? 'success' : 'warning'">
              {{ row.appVisible === 1 ? t('itick.show') : t('itick.hide') }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="sort" :label="t('common.sort')" width="90" />
        <el-table-column
          prop="icon"
          :label="t('common.icon')"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column
          prop="remark"
          :label="t('common.remark')"
          min-width="180"
          show-overflow-tooltip
        />

        <el-table-column :label="t('common.createTimes')" min-width="170">
          <template #default="{ row }">
            {{ formatTime(row.createTimes) }}
          </template>
        </el-table-column>

        <el-table-column :label="t('itick.updateTimes')" min-width="170">
          <template #default="{ row }">
            {{ formatTime(row.updateTimes) }}
          </template>
        </el-table-column>

        <el-table-column :label="t('common.actions')" width="180" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleDetail(row)">{{
              t('itick.detail')
            }}</el-button>
            <el-button link type="primary" @click="handleEdit(row)">{{
              t('common.edit')
            }}</el-button>
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
      :title="formMode === 'add' ? t('itick.addProduct') : t('itick.editProduct')"
      width="700px"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="120px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item
              v-if="formMode === 'add'"
              :label="t('itick.categoryType')"
              prop="categoryType"
            >
              <el-input-number
                v-model="form.categoryType"
                :min="1"
                :precision="0"
                controls-position="right"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('itick.categoryName')" prop="categoryName">
              <el-input
                v-model="form.categoryName"
                :placeholder="t('itick.pleaseInputCategoryName')"
                maxlength="100"
                show-word-limit
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="t('itick.categoryCode')" prop="categoryCode">
              <el-input
                v-model="form.categoryCode"
                :placeholder="t('itick.categoryCode')"
                maxlength="50"
                show-word-limit
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('itick.market')" prop="market">
              <el-input
                v-model="form.market"
                :placeholder="t('itick.pleaseInputMarket')"
                maxlength="50"
                show-word-limit
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="t('itick.symbol')" prop="symbol">
              <el-input
                v-model="form.symbol"
                :placeholder="t('itick.pleaseInputSymbol')"
                maxlength="50"
                show-word-limit
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('itick.code')" prop="code">
              <el-input
                v-model="form.code"
                :placeholder="t('itick.pleaseInputCode')"
                maxlength="50"
                show-word-limit
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="t('itick.name')" prop="name">
              <el-input
                v-model="form.name"
                :placeholder="t('itick.pleaseInputName')"
                maxlength="100"
                show-word-limit
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('itick.displayName')" prop="displayName">
              <el-input
                v-model="form.displayName"
                :placeholder="t('itick.pleaseInputDisplayName')"
                maxlength="100"
                show-word-limit
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="t('itick.baseCoin')" prop="baseCoin">
              <el-input
                v-model="form.baseCoin"
                :placeholder="t('itick.pleaseInputBaseCoin')"
                maxlength="20"
                show-word-limit
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('itick.quoteCoin')" prop="quoteCoin">
              <el-input
                v-model="form.quoteCoin"
                :placeholder="t('itick.pleaseInputQuoteCoin')"
                maxlength="20"
                show-word-limit
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="t('itick.enabledStatus')" prop="enabled">
              <el-radio-group v-model="form.enabled">
                <el-radio :value="1">{{ t('common.enabled') }}</el-radio>
                <el-radio :value="2">{{ t('common.disabled') }}</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('itick.appVisible')" prop="appVisible">
              <el-radio-group v-model="form.appVisible">
                <el-radio :value="1">{{ t('itick.show') }}</el-radio>
                <el-radio :value="2">{{ t('itick.hide') }}</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="t('common.sort')" prop="sort">
              <el-input-number
                v-model="form.sort"
                :min="0"
                :precision="0"
                controls-position="right"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('common.icon')" prop="icon">
              <el-input v-model="form.icon" :placeholder="t('common.icon')" />
            </el-form-item>
          </el-col>
        </el-row>

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
        <el-button @click="formDialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitForm">{{
          t('common.confirm')
        }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailDialogVisible" :title="t('itick.productDetail')" width="800px">
      <el-descriptions :column="2" border v-loading="detailLoading">
        <el-descriptions-item :label="t('common.id')">{{ detail.id ?? '-' }}</el-descriptions-item>
        <el-descriptions-item :label="t('itick.categoryType')">{{
          detail.categoryType ?? '-'
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('itick.categoryName')">{{
          detail.categoryName || '-'
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('itick.categoryCode')">{{
          detail.categoryCode || '-'
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('itick.market')">{{
          detail.market || '-'
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('itick.symbol')">{{
          detail.symbol || '-'
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('itick.code')">{{
          detail.code || '-'
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('itick.name')">{{
          detail.name || '-'
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('itick.displayName')">{{
          detail.displayName || '-'
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('itick.baseCoin')">{{
          detail.baseCoin || '-'
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('itick.quoteCoin')">{{
          detail.quoteCoin || '-'
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('itick.enabledStatus')">
          {{
            detail.enabled === 1
              ? t('common.enabled')
              : detail.enabled === 2
              ? t('common.disabled')
              : '-'
          }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('itick.appVisible')">
          {{
            detail.appVisible === 1
              ? t('itick.show')
              : detail.appVisible === 2
              ? t('itick.hide')
              : '-'
          }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.sort')">{{
          detail.sort ?? '-'
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('common.icon')">{{
          detail.icon || '-'
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('common.remark')" :span="2">{{
          detail.remark || '-'
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('common.createTimes')">
          {{ formatTime(detail.createTimes) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('itick.updateTimes')">
          {{ formatTime(detail.updateTimes) }}
        </el-descriptions-item>
      </el-descriptions>

      <template #footer>
        <el-button type="primary" @click="detailDialogVisible = false">{{
          t('common.close')
        }}</el-button>
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
  productsService,
  type ItickProduct,
  type ListProductsReq,
} from '@/services/itick/ProductsService'

type FormData = {
  id?: number
  categoryType?: number
  categoryName: string
  categoryCode: string
  market: string
  symbol: string
  code: string
  name: string
  displayName: string
  baseCoin: string
  quoteCoin: string
  enabled: number
  appVisible: number
  sort: number
  icon: string
  remark: string
}

const { t } = useI18n()
const { pagination, updatePagination, reset: resetPagination } = usePagination(20)
const { loading, withLoading } = useLoading()

const { form: queryParams, reset: resetQueryParams } = useForm<ListProductsReq>({
  initialData: {
    categoryType: undefined,
    market: '',
    keyword: '',
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
    categoryCode: '',
    market: '',
    symbol: '',
    code: '',
    name: '',
    displayName: '',
    baseCoin: '',
    quoteCoin: '',
    enabled: 1,
    appVisible: 1,
    sort: 0,
    icon: '',
    remark: '',
  },
})

const submitLoading = ref(false)
const detailLoading = ref(false)
const list = ref<ItickProduct[]>([])
const detail = ref<Partial<ItickProduct>>({})
const formDialogVisible = ref(false)
const detailDialogVisible = ref(false)
const formMode = ref<'add' | 'edit'>('add')

const rules: FormRules<FormData> = {
  categoryType: [{ required: true, message: t('itick.pleaseInputCategoryType'), trigger: 'blur' }],
  categoryName: [{ required: true, message: t('itick.pleaseInputCategoryName'), trigger: 'blur' }],
  categoryCode: [{ required: true, message: t('itick.categoryCode'), trigger: 'blur' }],
  market: [{ required: true, message: t('itick.pleaseInputMarket'), trigger: 'blur' }],
  symbol: [{ required: true, message: t('itick.pleaseInputSymbol'), trigger: 'blur' }],
  code: [{ required: true, message: t('itick.pleaseInputCode'), trigger: 'blur' }],
  name: [{ required: true, message: t('itick.pleaseInputName'), trigger: 'blur' }],
  displayName: [{ required: true, message: t('itick.pleaseInputDisplayName'), trigger: 'blur' }],
  baseCoin: [{ required: true, message: t('itick.pleaseInputBaseCoin'), trigger: 'blur' }],
  quoteCoin: [{ required: true, message: t('itick.pleaseInputQuoteCoin'), trigger: 'blur' }],
  enabled: [{ required: true, message: t('itick.pleaseSelectEnabledStatus'), trigger: 'change' }],
  appVisible: [{ required: true, message: t('itick.pleaseSelectAppVisible'), trigger: 'change' }],
  sort: [{ required: true, message: t('itick.pleaseInputSort'), trigger: 'blur' }],
}

const cleanedQueryParams = computed<ListProductsReq>(() => {
  const params: ListProductsReq = {
    cursor: queryParams.cursor,
    limit: queryParams.limit,
  }

  if (queryParams.categoryType && queryParams.categoryType !== 0) {
    params.categoryType = Number(queryParams.categoryType)
  }
  if (queryParams.market && queryParams.market.trim()) {
    params.market = queryParams.market.trim()
  }
  if (queryParams.keyword && queryParams.keyword.trim()) {
    params.keyword = queryParams.keyword.trim()
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
      const res = await productsService.getList({
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

const handleEdit = async (row: ItickProduct) => {
  formMode.value = 'edit'
  resetForm()

  try {
    const res = await productsService.detail(row.id)
    const data = res?.data || row

    Object.assign(form, {
      id: data.id,
      categoryType: data.categoryType,
      categoryName: data.categoryName || '',
      categoryCode: data.categoryCode || '',
      market: data.market || '',
      symbol: data.symbol || '',
      code: data.code || '',
      name: data.name || '',
      displayName: data.displayName || '',
      baseCoin: data.baseCoin || '',
      quoteCoin: data.quoteCoin || '',
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

const handleDetail = async (row: ItickProduct) => {
  detailDialogVisible.value = true
  detailLoading.value = true
  detail.value = {}

  try {
    const res = await productsService.detail(row.id)
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
      await productsService.create({
        categoryType: Number(form.categoryType),
        categoryName: form.categoryName,
        categoryCode: form.categoryCode,
        market: form.market,
        symbol: form.symbol,
        code: form.code,
        name: form.name,
        displayName: form.displayName,
        baseCoin: form.baseCoin,
        quoteCoin: form.quoteCoin,
        enabled: form.enabled,
        appVisible: form.appVisible,
        sort: form.sort,
        icon: form.icon,
        remark: form.remark,
      })
      ElMessage.success(t('common.createSuccess'))
    } else {
      await productsService.update(form.id as number, {
        name: form.name,
        displayName: form.displayName,
        baseCoin: form.baseCoin,
        quoteCoin: form.quoteCoin,
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

const formatTime = (timestamp?: number) => {
  if (!timestamp) return '-'
  let time = Number(timestamp)
  if (String(time).length === 10) time = time * 1000
  const date = new Date(time)
  const YYYY = date.getFullYear()
  const MM = String(date.getMonth() + 1).padStart(2, '0')
  const DD = String(date.getDate()).padStart(2, '0')
  const hh = String(date.getHours()).padStart(2, '0')
  const mm = String(date.getMinutes()).padStart(2, '0')
  const ss = String(date.getSeconds()).padStart(2, '0')
  return `${YYYY}-${MM}-${DD} ${hh}:${mm}:${ss}`
}

onMounted(() => {
  getList()
})
</script>

<style scoped>
.itick-products {
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

.query-card {
  margin-bottom: 16px;
}

.table-card {
  margin-bottom: 16px;
}

.pagination-bar {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 16px;
  margin-top: 16px;
}
</style>
