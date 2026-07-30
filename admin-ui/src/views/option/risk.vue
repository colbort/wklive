<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="loadAll" @reset="resetQuery">
      <el-form-item :label="t('option.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('option.userId')">
        <UserSelect v-model="query.userId" :tenant-id="query.tenantId || undefined" />
      </el-form-item>
      <el-form-item :label="t('option.contractId')">
        <ContractSelect v-model="query.contractId" :tenant-id="query.tenantId || undefined" />
      </el-form-item>
    </CrudQueryCard>

    <el-card v-loading="loading" shadow="never" class="table-card">
      <template #header>{{ t('option.riskAccounts') }}</template>
      <el-table :data="riskAccounts" stripe>
        <el-table-column prop="userId" :label="t('option.userId')" width="100" />
        <el-table-column prop="accountId" :label="t('option.accountId')" width="110" />
        <el-table-column prop="settleCoin" :label="t('option.settleCoin')" width="100" />
        <el-table-column prop="equity" :label="t('option.equity')" min-width="130" />
        <el-table-column
          prop="maintenanceMargin"
          :label="t('option.maintenanceMargin')"
          min-width="150"
        />
        <el-table-column
          prop="portfolioScenarioLoss"
          :label="t('option.portfolioScenarioLoss')"
          min-width="160"
        />
        <el-table-column
          prop="portfolioShortFloor"
          :label="t('option.portfolioShortFloor')"
          min-width="160"
        />
        <el-table-column prop="riskRate" :label="t('option.riskRate')" min-width="120" />
        <el-table-column prop="status" :label="t('option.status')" width="90" />
      </el-table>
    </el-card>

    <el-card v-loading="loading" shadow="never" class="table-card">
      <template #header>{{ t('option.liquidations') }}</template>
      <el-table :data="liquidations" stripe>
        <el-table-column prop="liquidationNo" :label="t('option.liquidationNo')" min-width="220" />
        <el-table-column prop="userId" :label="t('option.userId')" width="100" />
        <el-table-column prop="contractId" :label="t('option.contractId')" width="100" />
        <el-table-column prop="quantity" :label="t('option.quantity')" min-width="110" />
        <el-table-column prop="collateralAmount" :label="t('option.collateral')" min-width="130" />
        <el-table-column
          prop="insuranceFundAmount"
          :label="t('option.insuranceFund')"
          min-width="140"
        />
        <el-table-column
          prop="backstopAmount"
          :label="t('option.platformBackstop')"
          min-width="130"
        />
        <el-table-column
          prop="deficitResolution"
          :label="t('option.deficitResolution')"
          min-width="130"
        />
        <el-table-column prop="remainingDeficit" :label="t('option.deficit')" min-width="120" />
        <el-table-column prop="status" :label="t('option.status')" width="90" />
        <el-table-column :label="t('common.actions')" width="110" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="[4, 6].includes(row.status)"
              v-perm="'option:liquidation:retry'"
              link
              type="warning"
              @click="retryLiquidation(row)"
            >
              {{ t('option.retry') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import {
  optionService,
  type OptionLiquidation,
  type OptionRiskAccount,
} from '@/services'
import TenantSelect from '@/components/TenantSelect.vue'
import UserSelect from '@/components/UserSelect.vue'
import ContractSelect from '@/components/ContractSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()
const loading = ref(false)
const riskAccounts = ref<OptionRiskAccount[]>([])
const liquidations = ref<OptionLiquidation[]>([])
const query = reactive({
  tenantId: undefined as number | undefined,
  userId: undefined as number | undefined,
  contractId: undefined as number | undefined,
})

async function loadAll() {
  loading.value = true
  try {
    const [riskResp, liquidationResp] = await Promise.all([
      optionService.listRiskAccounts({ ...query, limit: 100 }),
      optionService.listLiquidations({ ...query, limit: 100 }),
    ])
    riskAccounts.value = riskResp.data || []
    liquidations.value = liquidationResp.data || []
  } finally {
    loading.value = false
  }
}

function resetQuery() {
  query.tenantId = undefined
  query.userId = undefined
  query.contractId = undefined
  void loadAll()
}

async function retryLiquidation(row: OptionLiquidation) {
  await ElMessageBox.confirm(
    t('option.retryLiquidationConfirm'),
    t('option.retryLiquidation'),
    { type: 'warning' },
  )
  await optionService.retryLiquidation({
    tenantId: row.tenantId,
    liquidationId: row.id,
  })
  ElMessage.success(t('common.success'))
  await loadAll()
}

onMounted(loadAll)
</script>

<style scoped>
.table-card + .table-card {
  margin-top: 16px;
}
</style>
