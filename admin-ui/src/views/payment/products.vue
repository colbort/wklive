<template>
  <div class="payment-page module-page">
    <div class="page-header">
      <h2>{{ t('payment.products') }}</h2>
      <div class="header-actions">
        <el-button @click="loadProducts">
          <el-icon><Refresh /></el-icon>
          {{ t('common.refresh') }}
        </el-button>
      </div>
    </div>

    <el-card shadow="never" class="query-card">
      <el-form :model="productQuery" inline label-width="90px">
        <el-form-item :label="t('payment.platformId')">
          <el-input-number v-model="productQuery.platformId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('payment.productCode')">
          <el-input v-model="productQuery.productCode" clearable />
        </el-form-item>
        <el-form-item :label="t('common.keyword')">
          <el-input v-model="productQuery.keyword" clearable />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadProducts">
            {{ t('common.search') }}
          </el-button>
          <el-button @click="resetProductQuery">
            {{ t('common.reset') }}
          </el-button>
          <el-button type="primary" @click="openProductDialog()">
            {{ t('payment.addProduct') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="productLoading" :data="products" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="platformId" :label="t('payment.platformId')" width="100" />
        <el-table-column prop="productCode" :label="t('payment.productCode')" min-width="140" />
        <el-table-column prop="productName" :label="t('payment.productName')" min-width="160" />
        <el-table-column :label="t('payment.sceneType')" width="120">
          <template #default="{ row }">
            {{ getOptionValueLabel(optionGroups, 'sceneType', row.sceneType, t) }}
          </template>
        </el-table-column>
        <el-table-column prop="currency" :label="t('payment.currency')" width="100" />
        <el-table-column :label="t('common.status')" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">
              {{ getOptionValueLabel(optionGroups, 'status', row.status, t) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="remark"
          :label="t('common.remark')"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.actions')" width="160" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showProductDetail(row)">
              {{ t('common.detail') }}
            </el-button>
            <el-button link type="primary" @click="openProductDialog(row)">
              {{ t('common.edit') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog
      v-model="productDialogVisible"
      :title="productForm.id ? t('payment.editProduct') : t('payment.addProduct')"
      width="640px"
    >
      <el-form label-width="100px">
        <el-form-item v-if="!productForm.id" :label="t('payment.platformId')">
          <div class="platform-verify-row">
            <el-input-number
              v-model="productForm.platformId"
              :min="1"
              :precision="0"
              @change="handlePlatformIdChange"
            />
            <el-button :loading="platformChecking" @click="checkPlatform">
              {{ t('payment.verifyPlatform') }}
            </el-button>
            <span v-if="platformVerified" class="platform-verified-text">
              {{ t('payment.verified') }}
            </span>
          </div>
        </el-form-item>
        <el-form-item v-if="!productForm.id" :label="t('payment.productCode')">
          <el-input v-model="productForm.productCode" />
        </el-form-item>
        <el-form-item :label="t('payment.productName')">
          <el-input v-model="productForm.productName" />
        </el-form-item>
        <el-form-item :label="t('payment.sceneType')">
          <el-select v-model="productForm.sceneType" style="width: 100%">
            <el-option
              v-for="item in sceneTypeOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('payment.currency')">
          <el-input v-model="productForm.currency" />
        </el-form-item>
        <el-form-item :label="t('common.status')">
          <el-select v-model="productForm.status" style="width: 100%">
            <el-option
              v-for="item in statusOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('common.remark')">
          <el-input v-model="productForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="productDialogVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button
          type="primary"
          :loading="submitLoading"
          :disabled="isSubmitDisabled"
          @click="submitProduct"
        >
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" :title="detailTitle" width="720px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Refresh } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { catalogService, type OptionGroup, type PayProduct } from '@/services'
import { findOptionGroup, getOptionLabel, getOptionValueLabel } from '@/utils/options'

const { t } = useI18n()

const submitLoading = ref(false)
const productLoading = ref(false)
const products = ref<PayProduct[]>([])
const detailVisible = ref(false)
const detailTitle = ref('')
const detailData = ref<PayProduct | null>(null)
const productDialogVisible = ref(false)
const platformChecking = ref(false)
const platformVerified = ref(false)
const verifiedPlatformId = ref(0)
const optionGroups = ref<OptionGroup[]>([])
const statusOptions = computed(() => findOptionGroup(optionGroups.value, 'status'))
const sceneTypeOptions = computed(() => findOptionGroup(optionGroups.value, 'sceneType'))
const isSubmitDisabled = computed(
  () =>
    !productForm.id &&
    (!platformVerified.value || verifiedPlatformId.value !== productForm.platformId),
)

const productQuery = reactive({ platformId: 0, productCode: '', keyword: '' })

const productForm = reactive({
  id: 0,
  platformId: 0,
  productCode: '',
  productName: '',
  sceneType: 1,
  currency: '',
  status: 1,
  remark: '',
})

const loadProducts = async () => {
  productLoading.value = true
  try {
    const res = await catalogService.getProductList({
      ...productQuery,
      platformId: productQuery.platformId || undefined,
      limit: 100,
    })
    products.value = res.data || []
  } finally {
    productLoading.value = false
  }
}

const loadOptions = async () => {
  const res = await catalogService.getOptions()
  optionGroups.value = res.data || []
}

const resetProductQuery = () => {
  Object.assign(productQuery, { platformId: 0, productCode: '', keyword: '' })
  loadProducts()
}

const openProductDialog = (row?: PayProduct) => {
  Object.assign(
    productForm,
    row || {
      id: 0,
      platformId: 0,
      productCode: '',
      productName: '',
      sceneType: 1,
      currency: '',
      status: 1,
      remark: '',
    },
  )
  platformVerified.value = !!row?.id
  verifiedPlatformId.value = row?.id ? row.platformId : 0
  productDialogVisible.value = true
}

const handlePlatformIdChange = () => {
  platformVerified.value = false
  verifiedPlatformId.value = 0
}

const validatePlatformExists = async (platformId: number) => {
  if (!platformId) {
    ElMessage.warning(t('payment.pleaseInputPlatformId'))
    return false
  }

  platformChecking.value = true
  try {
    const res = await catalogService.getPlatformDetail(platformId)
    if (!res.data?.id) {
      ElMessage.error(t('payment.platformNotFound'))
      return false
    }
    return true
  } catch {
    ElMessage.error(t('payment.platformNotFound'))
    return false
  } finally {
    platformChecking.value = false
  }
}

const checkPlatform = async () => {
  const platformExists = await validatePlatformExists(productForm.platformId)
  platformVerified.value = platformExists
  verifiedPlatformId.value = platformExists ? productForm.platformId : 0
  if (platformExists) {
    ElMessage.success(t('payment.platformVerifiedSuccess'))
  }
}

const submitProduct = async () => {
  if (!productForm.id) {
    if (!platformVerified.value || verifiedPlatformId.value !== productForm.platformId) {
      ElMessage.warning(t('payment.pleaseVerifyPlatformFirst'))
      return
    }
  }

  submitLoading.value = true
  try {
    if (productForm.id) {
      await catalogService.updateProduct(productForm.id, { ...productForm })
    } else {
      await catalogService.createProduct({ ...productForm })
    }
    ElMessage.success(t('common.operationSuccess'))
    productDialogVisible.value = false
    loadProducts()
  } finally {
    submitLoading.value = false
  }
}

const showProductDetail = async (row: PayProduct) => {
  const res = await catalogService.getProductDetail(row.id)
  detailTitle.value = t('payment.productDetail')
  detailData.value = res.data || row
  detailVisible.value = true
}

onMounted(async () => {
  await Promise.all([loadProducts(), loadOptions()])
})
</script>

<style scoped>
.query-card :deep(.el-form-item) {
  margin-bottom: 12px;
}

.platform-verify-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.platform-verified-text {
  color: var(--el-color-success);
  font-size: 14px;
}
</style>
