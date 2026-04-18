<template>
  <div class="module-page">
    <div class="page-header">
      <h2>{{ t('staking.products') }}</h2>
      <div class="header-actions">
        <el-button @click="loadProducts"> {{ t('common.refresh') }} </el-button>
        <el-button type="primary" @click="openProductDialog()"> {{ t('staking.addProduct') }} </el-button>
      </div>
    </div>

    <el-card shadow="never" class="query-card">
      <el-form :model="query" inline label-width="90px">
        <el-form-item :label="t('staking.tenantId')">
          <el-input-number v-model="query.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('staking.productNo')">
          <el-input v-model="query.productNo" clearable />
        </el-form-item>
        <el-form-item :label="t('staking.productName')">
          <el-input v-model="query.productName" clearable />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadProducts"> {{ t('common.search') }} </el-button>
          <el-button @click="resetQuery"> {{ t('common.reset') }} </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column :label="t('staking.productNo')" prop="productNo" min-width="180" show-overflow-tooltip />
        <el-table-column
          prop="productName"
          :label="t('staking.productName')"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column :label="t('staking.coinSymbol')" prop="coinSymbol" width="120" />
        <el-table-column prop="apr" label="APR" min-width="120" show-overflow-tooltip />
        <el-table-column :label="t('staking.lockDays')" prop="lockDays" width="120" />
        <el-table-column :label="t('common.status')" prop="status" width="100" />
        <el-table-column :label="t('common.actions')" width="220" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)"> {{ t('common.detail') }} </el-button>
            <el-button link type="primary" @click="openProductDialog(row)"> {{ t('common.edit') }} </el-button>
            <el-button link type="warning" @click="changeStatus(row)"> {{ t('staking.statusAction') }} </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog
      v-model="productVisible"
      :title="productForm.id ? t('staking.editProduct') : t('staking.addProduct')"
      width="760px"
    >
      <el-form label-width="110px">
        <el-form-item :label="t('staking.tenantId')">
          <el-input-number v-model="productForm.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('staking.productNo')">
          <el-input v-model="productForm.productNo" />
        </el-form-item>
        <el-form-item :label="t('staking.productName')">
          <el-input v-model="productForm.productName" />
        </el-form-item>
        <el-form-item :label="t('staking.productType')">
          <el-select v-model="productForm.productType" style="width: 100%">
            <el-option
              v-for="item in productTypeOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('staking.coinName')">
          <el-input v-model="productForm.coinName" />
        </el-form-item>
        <el-form-item :label="t('staking.coinSymbol')">
          <el-input v-model="productForm.coinSymbol" />
        </el-form-item>
        <el-form-item :label="t('staking.rewardCoinName')">
          <el-input v-model="productForm.rewardCoinName" />
        </el-form-item>
        <el-form-item :label="t('staking.rewardCoinSymbol')">
          <el-input v-model="productForm.rewardCoinSymbol" />
        </el-form-item>
        <el-form-item label="APR">
          <el-input v-model="productForm.apr" />
        </el-form-item>
        <el-form-item :label="t('staking.lockDays')">
          <el-input-number v-model="productForm.lockDays" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('staking.minAmount')">
          <el-input v-model="productForm.minAmount" />
        </el-form-item>
        <el-form-item :label="t('staking.maxAmount')">
          <el-input v-model="productForm.maxAmount" />
        </el-form-item>
        <el-form-item :label="t('staking.stepAmount')">
          <el-input v-model="productForm.stepAmount" />
        </el-form-item>
        <el-form-item :label="t('staking.totalAmount')">
          <el-input v-model="productForm.totalAmount" />
        </el-form-item>
        <el-form-item :label="t('staking.userLimitAmount')">
          <el-input v-model="productForm.userLimitAmount" />
        </el-form-item>
        <el-form-item :label="t('staking.interestMode')">
          <el-select v-model="productForm.interestMode" style="width: 100%">
            <el-option
              v-for="item in interestModeOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('staking.rewardMode')">
          <el-select v-model="productForm.rewardMode" style="width: 100%">
            <el-option
              v-for="item in rewardModeOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('staking.allowEarlyRedeem')">
          <el-select v-model="productForm.allowEarlyRedeem" style="width: 100%">
            <el-option
              v-for="item in yesNoOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('staking.earlyRedeemRate')">
          <el-input v-model="productForm.earlyRedeemRate" />
        </el-form-item>
        <el-form-item :label="t('common.status')">
          <el-select v-model="productForm.status" style="width: 100%">
            <el-option
              v-for="item in productStatusOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('common.sort')">
          <el-input-number v-model="productForm.sort" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('staking.operatorUid')">
          <el-input-number v-model="productForm.operatorUid" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('common.remark')">
          <el-input v-model="productForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="productVisible = false"> {{ t('common.cancel') }} </el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitProduct"> {{ t('common.confirm') }} </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" :title="t('itick.detail')" width="760px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { AdminProductUpdateReq, stakingService, type OptionGroup } from '@/services'
import { findOptionGroup, getOptionLabel } from '@/utils/options'

const { t } = useI18n()

const optionGroups = ref<OptionGroup[]>([])
const productTypeOptions = computed(() => findOptionGroup(optionGroups.value, 'productType'))
const productStatusOptions = computed(() => findOptionGroup(optionGroups.value, 'productStatus'))
const interestModeOptions = computed(() => findOptionGroup(optionGroups.value, 'interestMode'))
const rewardModeOptions = computed(() => findOptionGroup(optionGroups.value, 'rewardMode'))
const yesNoOptions = computed(() => findOptionGroup(optionGroups.value, 'yesNo'))
const loading = ref(false)
const submitLoading = ref(false)
const rows = ref<Record<string, any>[]>([])
const detailVisible = ref(false)
const detailData = ref<Record<string, any>>({})
const productVisible = ref(false)

const query = reactive({
  tenantId: undefined as number | undefined,
  productNo: '',
  productName: '',
  coinSymbol: '',
  limit: 100,
})

const productForm = reactive<AdminProductUpdateReq>({
  id: 0,
  tenantId: 0,
  productNo: '',
  productName: '',
  productType: 1,
  coinName: '',
  coinSymbol: '',
  rewardCoinName: '',
  rewardCoinSymbol: '',
  apr: '',
  lockDays: 0,
  minAmount: '',
  maxAmount: '',
  stepAmount: '',
  totalAmount: '',
  userLimitAmount: '',
  interestMode: 1,
  rewardMode: 1,
  allowEarlyRedeem: 0,
  earlyRedeemRate: '',
  status: 1,
  sort: 0,
  operatorUid: 0,
  remark: '',
})

const pickList = (res: any) => res?.data || res?.list || []

const loadOptions = async () => {
  const res = await stakingService.getOptions()
  optionGroups.value = res.data || []
}

const loadProducts = async () => {
  loading.value = true
  try {
    rows.value = pickList(await stakingService.listProducts(query))
  } finally {
    loading.value = false
  }
}

const resetQuery = () => {
  query.tenantId = undefined
  query.productNo = ''
  query.productName = ''
  query.coinSymbol = ''
  query.limit = 100
  loadProducts()
}

const showDetail = async (row: Record<string, any>) => {
  detailData.value =
    (await stakingService.getProduct({ tenantId: row.tenantId, id: row.id })).data || row
  detailVisible.value = true
}

const openProductDialog = (row?: Record<string, any>) => {
  Object.assign(
    productForm,
    {
      id: 0,
      tenantId: 0,
      productNo: '',
      productName: '',
      productType: 1,
      coinName: '',
      coinSymbol: '',
      rewardCoinName: '',
      rewardCoinSymbol: '',
      apr: '',
      lockDays: 0,
      minAmount: '',
      maxAmount: '',
      stepAmount: '',
      totalAmount: '',
      userLimitAmount: '',
      interestMode: 1,
      rewardMode: 1,
      allowEarlyRedeem: 0,
      earlyRedeemRate: '',
      status: 1,
      sort: 0,
      operatorUid: 0,
      remark: '',
    },
    row || {},
  )
  productVisible.value = true
}

const submitProduct = async () => {
  submitLoading.value = true
  try {
    if (productForm.id) await stakingService.updateProduct(productForm)
    else await stakingService.createProduct(productForm)
    ElMessage.success(t('staking.saveSuccess'))
    productVisible.value = false
    loadProducts()
  } finally {
    submitLoading.value = false
  }
}

const changeStatus = async (row: Record<string, any>) => {
  const { value } = await ElMessageBox.prompt(t('staking.pleaseInputNewStatus'), t('staking.changeStatus'), {
    inputValue: String(row.status || 0),
  })
  await stakingService.changeProductStatus({
    tenantId: row.tenantId,
    id: row.id,
    status: Number(value),
    operatorUid: row.updateUserId || 0,
  })
  ElMessage.success(t('staking.statusUpdated'))
  loadProducts()
}

onMounted(async () => {
  await Promise.all([loadProducts(), loadOptions()])
})
</script>

<style scoped></style>
