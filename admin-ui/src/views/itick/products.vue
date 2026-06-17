<template>
  <div class="itick-products module-page">
    <CrudQueryCard :model="queryParams" @search="loadList" @reset="resetQuery">
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

      <el-form-item :label="t('itick.market')">
        <el-input
          v-model="queryParams.market"
          :placeholder="t('itick.pleaseInputMarket')"
          clearable
          style="width: 180px"
          @keyup.enter="loadList"
        />
      </el-form-item>

      <el-form-item :label="t('itick.symbol')">
        <el-input
          v-model="queryParams.symbol"
          :placeholder="t('itick.pleaseInputSymbol')"
          clearable
          style="width: 180px"
          @keyup.enter="loadList"
        />
      </el-form-item>

      <el-form-item :label="t('itick.keyword')">
        <el-input
          v-model="queryParams.keyword"
          :placeholder="t('common.keyword')"
          clearable
          style="width: 180px"
          @keyup.enter="loadList"
        />
      </el-form-item>

      <el-form-item :label="t('itick.enabledStatus')">
        <el-select
          v-model="queryParams.enabled"
          :placeholder="t('common.pleaseSelect')"
          clearable
          style="width: 180px"
        >
          <el-option
            v-for="item in enabledOptions"
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

      <template #actions>
        <el-button v-perm="'itick:product:add'" type="primary" @click="handleAdd">
          <el-icon><Plus /></el-icon>
          {{ t('common.add') }}
        </el-button>
      </template>
    </CrudQueryCard>

    <el-card class="table-card" shadow="never">
      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" :label="t('common.id')" width="80" />
        <el-table-column :label="t('itick.categoryType')" width="120">
          <template #default="{ row }">
            {{ getOptionValueLabel(optionGroups, 'categoryType', row.categoryType, t) }}
          </template>
        </el-table-column>
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
              {{ getOptionValueLabel(optionGroups, 'enabled', row.enabled, t) }}
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

        <el-table-column :label="t('itick.syncPriority')" width="110">
          <template #default="{ row }">
            <el-tag :type="syncPriorityTagType(row.syncPriority)">
              {{ getSyncPriorityLabel(row.syncPriority) }}
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
            </div>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column
          prop="remark"
          :label="t('common.remark')"
          min-width="180"
          show-overflow-tooltip
        />

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

        <el-table-column
          :label="t('common.actions')"
          align="center"
          width="120"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              v-perm="'itick:product:detail'"
              link
              type="primary"
              @click="handleDetail(row)"
            >
              {{ t('itick.detail') }}
            </el-button>
            <el-button
              v-perm="'itick:product:update'"
              link
              type="primary"
              @click="handleEdit(row)"
            >
              {{ t('common.edit') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <CursorPagination
        v-model:limit="pagination.limit"
        :total="pagination.total"
        :has-prev="pagination.hasPrev"
        :has-next="pagination.hasNext"
        @prev="handlePrevPage"
        @next="handleNextPage"
        @limit-change="handleLimitChange"
      />
    </el-card>

    <el-dialog
      v-model="formDialogVisible"
      :title="formMode === 'add' ? t('itick.addProduct') : t('itick.editProduct')"
      width="700px"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="120px"
      >
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item
              v-if="formMode === 'add'"
              :label="t('itick.categoryType')"
              prop="categoryType"
            >
              <el-select
                v-model="form.categoryType"
                :placeholder="t('common.pleaseSelect')"
                style="width: 100%"
              >
                <el-option
                  v-for="item in categoryTypeFormOptions"
                  :key="item.value"
                  :label="getOptionLabel(t, item.code, item.value)"
                  :value="item.value"
                />
              </el-select>
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
              <el-select v-model="form.enabled" style="width: 100%">
                <el-option
                  v-for="item in enabledFormOptions"
                  :key="item.value"
                  :label="getOptionLabel(t, item.code, item.value)"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('itick.appVisible')" prop="appVisible">
              <el-select v-model="form.appVisible" style="width: 100%">
                <el-option
                  v-for="item in visibleFormOptions"
                  :key="item.value"
                  :label="getOptionLabel(t, item.code, item.value)"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="t('itick.syncPriority')" prop="syncPriority">
              <el-select v-model="form.syncPriority" style="width: 100%">
                <el-option
                  v-for="item in syncPriorityFormOptions"
                  :key="item.value"
                  :label="getOptionLabel(t, item.code, item.value)"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
          </el-col>
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
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="t('common.icon')" prop="icon">
              <div class="icon-upload-field">
                <div v-if="form.icon" class="icon-upload-preview">
                  <el-image
                    :src="resolveAssetUrl(form.icon)"
                    class="icon-preview-large"
                    :preview-teleported="true"
                  />
                  <div class="icon-url">
                    {{ form.icon }}
                  </div>
                </div>
                <el-upload
                  action="#"
                  :auto-upload="false"
                  :show-file-list="false"
                  :on-change="handleIconSelect"
                  accept="image/*"
                >
                  <el-button type="primary" :loading="submitLoading">
                    {{ t('itick.uploadImage') }}
                  </el-button>
                </el-upload>
              </div>
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
        <el-button @click="formDialogVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button
          v-perm="formMode === 'add' ? 'itick:product:add' : 'itick:product:update'"
          type="primary"
          :loading="submitLoading"
          @click="submitForm"
        >
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailDialogVisible" :title="t('itick.productDetail')" width="800px">
      <el-descriptions v-loading="detailLoading" :column="2" border>
        <el-descriptions-item :label="t('common.id')">
          {{ detail.id ?? '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('itick.categoryType')">
          {{ getOptionValueLabel(optionGroups, 'categoryType', detail.categoryType, t) || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('itick.categoryName')">
          {{ detail.categoryName || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('itick.categoryCode')">
          {{ detail.categoryCode || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('itick.market')">
          {{ detail.market || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('itick.symbol')">
          {{ detail.symbol || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('itick.code')">
          {{ detail.code || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('itick.name')">
          {{ detail.name || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('itick.displayName')">
          {{ detail.displayName || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('itick.baseCoin')">
          {{ detail.baseCoin || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('itick.quoteCoin')">
          {{ detail.quoteCoin || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('itick.enabledStatus')">
          {{ getOptionValueLabel(optionGroups, 'enabled', detail.enabled, t) || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('itick.appVisible')">
          {{ getOptionValueLabel(optionGroups, 'visible', detail.appVisible, t) || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('itick.syncPriority')">
          {{ getSyncPriorityLabel(detail.syncPriority) }}
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
            <div class="icon-url">
              {{ detail.icon }}
            </div>
          </div>
          <span v-else>-</span>
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.remark')" :span="2">
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
        <el-button type="primary" @click="detailDialogVisible = false">
          {{ t('common.close') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, onMounted } from 'vue'
import { ElMessage, type FormRules, type UploadFile } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables/usePagination'
import { useLoading } from '@/composables/useLoading'
import { useForm } from '@/composables/useForm'
import { buildSystemAssetUrl, useSystemCore } from '@/composables/useSystemCore'
import type { OptionGroup } from '@/services'
import { apiUploadFile } from '@/api/system/upload'
import {
  productsService,
  type ItickProduct,
  type ListProductsReq,
} from '@/services/itick/ProductsService'
import { formatDate } from '@/utils'
import {
  findFormOptionGroup,
  findOptionGroup,
  getOptionLabel,
  getOptionValueLabel,
} from '@/utils/options'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

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
  syncPriority: number
  sort: number
  icon: string
  remark: string
}

const { t } = useI18n()
const { systemCore, loadSystemCore } = useSystemCore()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)
const { loading, withLoading } = useLoading()

const { form: queryParams, reset: resetQueryParams } = useForm<ListProductsReq>({
  initialData: {
    categoryType: undefined,
    market: '',
    symbol: '',
    keyword: '',
    enabled: 0,
    appVisible: 0,
    cursor: undefined,
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
    syncPriority: 2,
    sort: 0,
    icon: '',
    remark: '',
  },
})

const submitLoading = ref(false)
const detailLoading = ref(false)
const list = ref<ItickProduct[]>([])
const detail = ref<Partial<ItickProduct>>({})
const optionGroups = ref<OptionGroup[]>([])
const formDialogVisible = ref(false)
const detailDialogVisible = ref(false)
const formMode = ref<'add' | 'edit'>('add')
const categoryTypeOptions = computed(() => findOptionGroup(optionGroups.value, 'categoryType'))
const enabledOptions = computed(() => findOptionGroup(optionGroups.value, 'enabled'))
const visibleOptions = computed(() => findOptionGroup(optionGroups.value, 'visible'))
const categoryTypeFormOptions = computed(() =>
  findFormOptionGroup(optionGroups.value, 'categoryType'),
)
const enabledFormOptions = computed(() => findFormOptionGroup(optionGroups.value, 'enabled'))
const visibleFormOptions = computed(() => findFormOptionGroup(optionGroups.value, 'visible'))
const syncPriorityFormOptions = computed(() =>
  findFormOptionGroup(optionGroups.value, 'syncPriority'),
)
const resolveAssetUrl = (url?: string) => buildSystemAssetUrl(systemCore.value.assetUrl, url)
const getSyncPriorityLabel = (value?: number) =>
  getOptionValueLabel(optionGroups.value, 'syncPriority', Number(value), t) || '-'
const syncPriorityTagType = (value?: number) => {
  if (Number(value) === 1) return 'danger'
  if (Number(value) === 3) return 'info'
  return 'success'
}

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
  syncPriority: [
    { required: true, message: t('itick.pleaseSelectSyncPriority'), trigger: 'change' },
  ],
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
  if (queryParams.symbol && queryParams.symbol.trim()) {
    params.symbol = queryParams.symbol.trim()
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
      updateFromResponse(res)
    } catch (_) {
      ElMessage.error(t('common.loadFailed'))
    }
  })
}

const loadOptions = async () => {
  try {
    const res = await productsService.getOptions()
    optionGroups.value = res.data || []
  } catch {
    ElMessage.error(t('common.loadFailed'))
  }
}

const loadList = () => {
  resetAndLoad(getList)
}

const resetQuery = () => {
  resetQueryParams()
  resetAndLoad(getList)
}

const handleLimitChange = () => {
  resetAndLoad(getList)
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
      syncPriority: data.syncPriority || 2,
      sort: data.sort || 0,
      icon: data.icon || '',
      remark: data.remark || '',
    })

    formDialogVisible.value = true
    await nextTick()
    formRef.value?.clearValidate()
  } catch (_) {
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
  } catch (_) {
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
    if (res.code === 200) {
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
        syncPriority: form.syncPriority,
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
        syncPriority: form.syncPriority,
        sort: form.sort,
        icon: form.icon,
        remark: form.remark,
      })
      ElMessage.success(t('common.updateSuccess'))
    }

    formDialogVisible.value = false
    getList()
  } catch (_) {
    ElMessage.error(formMode.value === 'add' ? t('common.createFailed') : t('common.updateFailed'))
  } finally {
    submitLoading.value = false
  }
}

const handlePrevPage = () => {
  prevAndLoad(getList)
}

const handleNextPage = () => {
  nextAndLoad(getList)
}

onMounted(() => {
  loadSystemCore()
  loadOptions()
  getList()
})
</script>

<style scoped>
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
