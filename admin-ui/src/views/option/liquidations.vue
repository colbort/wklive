<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="search" @reset="resetQuery">
      <el-form-item v-if="authStore.isSystemAdmin" :label="t('option.tenantId')">
        <TenantSelect v-model="query.tenantId" />
      </el-form-item>
      <el-form-item :label="t('option.userId')">
        <UserSelect v-model="query.userId" :tenant-id="query.tenantId" />
      </el-form-item>
      <el-form-item :label="t('option.accountId')">
        <el-input-number v-model="query.accountId" :min="0" :precision="0" />
      </el-form-item>
      <el-form-item :label="t('option.contractId')">
        <ContractSelect v-model="query.contractId" :tenant-id="query.tenantId" />
      </el-form-item>
    </CrudQueryCard>

    <el-card v-loading="loading" shadow="never" class="table-card">
      <el-table :data="items" stripe>
        <el-table-column
          v-if="authStore.isSystemAdmin"
          prop="tenantId"
          :label="t('option.tenantId')"
          width="110"
        />
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
              @click="retry(row)"
            >
              {{ t('option.retry') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      <CursorPagination
        v-model:limit="pagination.pagination.limit"
        :total="pagination.pagination.total"
        :has-prev="pagination.pagination.hasPrev"
        :has-next="pagination.pagination.hasNext"
        @prev="pagination.prevAndLoad(loadData)"
        @next="pagination.nextAndLoad(loadData)"
        @limit-change="pagination.resetAndLoad(loadData)"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { optionService, type OptionLiquidation } from '@/services'
import { usePagination } from '@/composables'
import { useAuthStore } from '@/stores/auth'
import TenantSelect from '@/components/TenantSelect.vue'
import UserSelect from '@/components/UserSelect.vue'
import ContractSelect from '@/components/ContractSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()
const authStore = useAuthStore()
const loading = ref(false)
const items = ref<OptionLiquidation[]>([])
const pagination = usePagination<number>(20)
const query = reactive({
  tenantId: (authStore.isTenantUser ? authStore.profileTenantId : undefined) as number | undefined,
  userId: undefined as number | undefined,
  accountId: undefined as number | undefined,
  contractId: undefined as number | undefined,
})

async function loadData() {
  loading.value = true
  try {
    const response = await optionService.listLiquidations({
      ...query,
      cursor: pagination.pagination.cursor,
      limit: pagination.pagination.limit,
    })
    items.value = response.data || []
    pagination.updateFromResponse(response)
  } finally {
    loading.value = false
  }
}

function search() {
  pagination.resetAndLoad(loadData)
}

function resetQuery() {
  query.tenantId = authStore.isTenantUser ? authStore.profileTenantId || undefined : undefined
  query.userId = undefined
  query.accountId = undefined
  query.contractId = undefined
  pagination.resetAndLoad(loadData)
}

async function retry(row: OptionLiquidation) {
  await ElMessageBox.confirm(t('option.retryLiquidationConfirm'), t('option.retryLiquidation'), {
    type: 'warning',
  })
  await optionService.retryLiquidation({ tenantId: row.tenantId, liquidationId: row.id })
  ElMessage.success(t('common.success'))
  await loadData()
}

onMounted(loadData)
</script>
