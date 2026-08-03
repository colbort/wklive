<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="load" @reset="resetQuery">
      <el-form-item :label="t('trade.tenantId')">
        <TenantSelect v-model="query.tenantId" />
      </el-form-item>
      <el-form-item :label="t('trade.symbolId')">
        <SymbolSelect
          v-model="query.symbolId"
          :tenant-id="query.tenantId"
        />
      </el-form-item>
      <el-form-item :label="t('trade.enabled')">
        <el-select v-model="query.enabled" clearable>
          <el-option :label="t('common.enabled')" :value="1" />
          <el-option :label="t('common.disabled')" :value="2" />
        </el-select>
      </el-form-item>
      <template #actions>
        <el-button v-perm="'trade:risk-tier:update'" type="primary" @click="openCreate">
          {{
            t('trade.addRiskTier')
          }}
        </el-button>
      </template>
    </CrudQueryCard>
    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column prop="id" :label="t('trade.id')" min-width="90" />
        <el-table-column prop="symbolId" :label="t('trade.symbolId')" min-width="110" />
        <el-table-column prop="tierNo" :label="t('trade.tierNo')" min-width="100" />
        <el-table-column prop="notionalFloor" :label="t('trade.notionalFloor')" min-width="140" />
        <el-table-column prop="notionalCap" :label="t('trade.notionalCap')" min-width="140" />
        <el-table-column prop="maxLeverage" :label="t('trade.maxLeverage')" min-width="130" />
        <el-table-column
          prop="initialMarginRate"
          :label="t('trade.initialMarginRate')"
          min-width="150"
        />
        <el-table-column
          prop="maintenanceMarginRate"
          :label="t('trade.maintenanceMarginRate')"
          min-width="180"
        />
        <el-table-column
          prop="enabled"
          :label="t('trade.enabled')"
          min-width="100"
        >
          <template #default="{ row }">
            <el-tag :type="row.enabled === 1 ? 'success' : 'info'">
              {{
                row.enabled === 1 ? t('common.enabled') : t('common.disabled')
              }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="updateTimes"
          :label="t('trade.updateTimes')"
          min-width="190"
        >
          <template #default="{ row }">
            {{
              formatTime(row.updateTimes)
            }}
          </template>
        </el-table-column>
      </el-table>
      <CursorPagination
        v-model:limit="pagination.limit"
        :total="pagination.total"
        :has-prev="pagination.hasPrev"
        :has-next="pagination.hasNext"
        @prev="prevAndLoad(load)"
        @next="nextAndLoad(load)"
        @limit-change="resetAndLoad(load)"
      />
    </el-card>

    <el-dialog v-model="visible" :title="t('trade.addRiskTier')" width="620px">
      <el-form :model="form" label-width="150px">
        <el-form-item :label="t('trade.tenantId')">
          <TenantSelect v-model="form.tenantId" />
        </el-form-item>
        <el-form-item :label="t('trade.symbolId')">
          <SymbolSelect
            v-model="form.symbolId"
            :tenant-id="form.tenantId"
          />
        </el-form-item>
        <el-form-item :label="t('trade.tierNo')">
          <el-input-number
            v-model="form.tierNo"
            :min="1"
          />
        </el-form-item>
        <el-form-item :label="t('trade.notionalFloor')">
          <el-input v-model="form.notionalFloor" />
        </el-form-item>
        <el-form-item :label="t('trade.notionalCap')">
          <el-input v-model="form.notionalCap" />
        </el-form-item>
        <el-form-item :label="t('trade.maxLeverage')">
          <el-input-number
            v-model="form.maxLeverage"
            :min="1"
          />
        </el-form-item>
        <el-form-item :label="t('trade.initialMarginRate')">
          <el-input v-model="form.initialMarginRate" />
        </el-form-item>
        <el-form-item :label="t('trade.maintenanceMarginRate')">
          <el-input v-model="form.maintenanceMarginRate" />
        </el-form-item>
        <el-form-item :label="t('trade.maintenanceAmount')">
          <el-input v-model="form.maintenanceAmount" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="visible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button type="primary" :loading="saving" @click="save">
          {{
            t('common.confirm')
          }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>
<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'
import SymbolSelect from '@/components/SymbolSelect.vue'
import TenantSelect from '@/components/TenantSelect.vue'
import { usePagination } from '@/composables'
import { formatDate } from '@/utils'
import { apiTradeListRiskTiers, apiTradeSetRiskTier } from '@/api/trade'
import type { ContractRiskLimitTier } from '@/services'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)
const query = reactive<{ tenantId?: number; symbolId?: number; enabled?: number }>({})
const rows = ref<ContractRiskLimitTier[]>([])
const loading = ref(false)
const visible = ref(false)
const saving = ref(false)
const form = reactive({
  tenantId: undefined as number | undefined,
  symbolId: 0,
  tierNo: 1,
  notionalFloor: '0',
  notionalCap: '0',
  maxLeverage: 1,
  initialMarginRate: '0',
  maintenanceMarginRate: '0',
  maintenanceAmount: '0',
  enabled: 1,
})
async function load() {
  loading.value = true
  try {
    const res = await apiTradeListRiskTiers({
      ...query,
      cursor: pagination.cursor,
      limit: pagination.limit,
    })
    rows.value = res.data || []
    updateFromResponse(res)
  } finally {
    loading.value = false
  }
}
function resetQuery() {
  Object.assign(query, { tenantId: undefined, symbolId: undefined, enabled: undefined })
  resetAndLoad(load)
}
function openCreate() {
  Object.assign(form, {
    tenantId: query.tenantId,
    symbolId: query.symbolId || 0,
    tierNo: 1,
    notionalFloor: '0',
    notionalCap: '0',
    maxLeverage: 1,
    initialMarginRate: '0',
    maintenanceMarginRate: '0',
    maintenanceAmount: '0',
    enabled: 1,
  })
  visible.value = true
}
async function save() {
  saving.value = true
  try {
    await apiTradeSetRiskTier(form)
    ElMessage.success(t('common.success'))
    visible.value = false
    await resetAndLoad(load)
  } finally {
    saving.value = false
  }
}
function formatTime(value: number) {
  return value > 0 ? formatDate(value) : '-'
}
onMounted(load)
</script>
