<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="loadList" @reset="resetQuery">
      <el-form-item :label="t('staking.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('staking.productNo')">
        <el-input v-model="query.productNo" clearable />
      </el-form-item>
      <el-form-item :label="t('staking.productName')">
        <el-input v-model="query.productName" clearable />
      </el-form-item>
      <el-form-item :label="t('staking.productType')">
        <el-select v-model="query.productType" clearable>
          <el-option
            v-for="item in productTypeFormOptions"
            :key="item.value"
            :label="getOptionLabel(t, item.code, item.value)"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('common.status')">
        <el-select v-model="query.status" clearable>
          <el-option
            v-for="item in productStatusFormOptions"
            :key="item.value"
            :label="getOptionLabel(t, item.code, item.value)"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <template #actions>
        <el-button v-perm="'staking:product:add'" type="primary" @click="openProductDialog()">
          {{ t('common.add') }}
        </el-button>
      </template>
    </CrudQueryCard>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column
          :label="t('staking.productNo')"
          prop="productNo"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column
          prop="productName"
          :label="t('staking.productName')"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column :label="t('staking.coinSymbol')" prop="coinSymbol" width="120" />
        <el-table-column :label="t('staking.productType')" prop="productType" width="110">
          <template #default="{ row }">
            {{ optionLabel('stakingProductType', row.productType) }}
          </template>
        </el-table-column>
        <el-table-column prop="apr" label="APR" min-width="120" show-overflow-tooltip />
        <el-table-column :label="t('staking.lockDays')" prop="lockDays" width="120" />
        <el-table-column :label="t('common.status')" prop="status" width="100">
          <template #default="{ row }">{{ optionLabel('productStatus', row.status) }}</template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" align="center" width="220" fixed="right">
          <template #default="{ row }">
            <el-button
              v-perm="'staking:product:detail'"
              link
              type="primary"
              @click="showDetail(row)"
            >
              {{ t('common.detail') }}
            </el-button>
            <el-button
              v-perm="'staking:product:update'"
              link
              type="primary"
              @click="openProductDialog(row)"
            >
              {{ t('common.edit') }}
            </el-button>
            <el-button
              v-perm="'staking:product:update:status'"
              link
              type="warning"
              @click="changeStatus(row)"
            >
              {{ t('staking.statusAction') }}
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
      v-model="productVisible"
      :title="productForm.id ? t('staking.editProduct') : t('staking.addProduct')"
      width="760px"
    >
      <el-form label-width="110px">
        <el-form-item :label="t('staking.tenantId')">
          <TenantSelect
            v-model="productForm.tenantId"
            include-system
            :disabled="productForm.id > 0"
          />
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
              v-for="item in productTypeFormOptions"
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
              v-for="item in interestModeFormOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('staking.rewardMode')">
          <el-select v-model="productForm.rewardMode" style="width: 100%">
            <el-option
              v-for="item in rewardModeFormOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('staking.allowEarlyRedeem')">
          <el-select v-model="productForm.allowEarlyRedeem" style="width: 100%">
            <el-option
              v-for="item in yesNoFormOptions"
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
              v-for="item in productStatusFormOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('common.sort')">
          <el-input-number v-model="productForm.sort" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('common.remark')">
          <el-input v-model="productForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="productVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button
          v-perm="productForm.id ? 'staking:product:update' : 'staking:product:add'"
          type="primary"
          :loading="submitLoading"
          @click="submitProduct"
        >
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="statusVisible" :title="t('staking.changeStatus')" width="420px">
      <el-form label-width="90px">
        <el-form-item :label="t('common.status')">
          <el-select v-model="statusForm.status" style="width: 100%">
            <el-option
              v-for="item in productStatusFormOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="statusVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitStatusChange">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" :title="t('market.detail')" width="760px">
      <el-descriptions v-if="detailData" :column="2" border>
        <el-descriptions-item :label="t('staking.productNo')">{{
          detailData.productNo
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.productName')">{{
          detailData.productName
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.tenantId')">{{
          detailData.tenantId
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.productType')">{{
          optionLabel('stakingProductType', detailData.productType)
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.coinSymbol')">{{
          detailData.coinSymbol
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.rewardCoinSymbol')">{{
          detailData.rewardCoinSymbol
        }}</el-descriptions-item>
        <el-descriptions-item label="APR">{{ detailData.apr }}%</el-descriptions-item>
        <el-descriptions-item :label="t('staking.lockDays')">{{
          detailData.lockDays
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.interestMode')">{{
          optionLabel('interestMode', detailData.interestMode)
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.rewardMode')">{{
          optionLabel('rewardMode', detailData.rewardMode)
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.allowEarlyRedeem')">{{
          optionLabel('yesNo', detailData.allowEarlyRedeem)
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.minAmount')">{{
          detailData.minAmount
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.maxAmount')">{{
          detailData.maxAmount
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.totalAmount')">{{
          detailData.totalAmount
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.stakedAmount')">{{
          detailData.stakedAmount
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('common.status')">{{
          optionLabel('productStatus', detailData.status)
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('common.updateTimes')">{{
          formatTime(detailData.updateTimes)
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('common.remark')" :span="2">{{
          detailData.remark || '-'
        }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables'
import { ElMessage } from 'element-plus'
import { ProductUpdateReq, stakingService, type OptionGroup, type StakeProduct } from '@/services'
import { findFormOptionGroup, getOptionLabel, getOptionValueLabel } from '@/utils/options'
import TenantSelect from '@/components/TenantSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)
const formatTime = (value: number) => (value > 0 ? new Date(value).toLocaleString() : '-')

const optionGroups = ref<OptionGroup[]>([])
const productTypeFormOptions = computed(() =>
  findFormOptionGroup(optionGroups.value, 'stakingProductType'),
)
const productStatusFormOptions = computed(() =>
  findFormOptionGroup(optionGroups.value, 'productStatus'),
)
const interestModeFormOptions = computed(() =>
  findFormOptionGroup(optionGroups.value, 'interestMode'),
)
const rewardModeFormOptions = computed(() => findFormOptionGroup(optionGroups.value, 'rewardMode'))
const yesNoFormOptions = computed(() => findFormOptionGroup(optionGroups.value, 'yesNo'))
const loading = ref(false)
const submitLoading = ref(false)
const rows = ref<StakeProduct[]>([])
const detailVisible = ref(false)
const detailData = ref<StakeProduct | null>(null)
const productVisible = ref(false)
const statusVisible = ref(false)

const query = reactive({
  tenantId: undefined as number | undefined,
  productNo: '',
  productName: '',
  coinSymbol: '',
  productType: undefined as number | undefined,
  status: undefined as number | undefined,
  limit: 20,
})

const statusForm = reactive({ tenantId: 0, id: 0, status: 1 })
const optionLabel = (key: string, value: number) =>
  getOptionValueLabel(optionGroups.value, key, value, t)

const productForm = reactive<ProductUpdateReq>({
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
  allowEarlyRedeem: 2,
  earlyRedeemRate: '',
  status: 1,
  sort: 0,
  remark: '',
})

const loadOptions = async () => {
  const res = await stakingService.getOptions()
  optionGroups.value = res.data || []
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await stakingService.listProducts({
      ...query,
      cursor: pagination.cursor,
      limit: pagination.limit,
    })
    rows.value = res?.data || []
    updateFromResponse(res)
  } finally {
    loading.value = false
  }
}

const resetQuery = () => {
  query.tenantId = undefined
  query.productNo = ''
  query.productName = ''
  query.coinSymbol = ''
  query.productType = undefined
  query.status = undefined
  query.limit = 20
  loadList()
}

const showDetail = async (row: StakeProduct) => {
  detailData.value =
    (await stakingService.getProduct({ tenantId: row.tenantId, id: row.id })).data || row
  detailVisible.value = true
}

const openProductDialog = (row?: StakeProduct) => {
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
      allowEarlyRedeem: 2,
      earlyRedeemRate: '',
      status: 1,
      sort: 0,
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
    loadList()
  } finally {
    submitLoading.value = false
  }
}

const changeStatus = async (row: StakeProduct) => {
  Object.assign(statusForm, { tenantId: row.tenantId, id: row.id, status: row.status })
  statusVisible.value = true
}

const submitStatusChange = async () => {
  submitLoading.value = true
  try {
    await stakingService.changeProductStatus({
      tenantId: statusForm.tenantId,
      id: statusForm.id,
      status: statusForm.status,
    })
    ElMessage.success(t('staking.statusUpdated'))
    statusVisible.value = false
    loadList()
  } finally {
    submitLoading.value = false
  }
}

function handleLimitChange() {
  resetAndLoad(loadList)
}

function handlePrevPage() {
  prevAndLoad(loadList)
}

function handleNextPage() {
  nextAndLoad(loadList)
}

onMounted(async () => {
  await Promise.all([loadList(), loadOptions()])
})
</script>
