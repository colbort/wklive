<template>
  <div>
    <CrudQueryCard :model="query" @search="load" @reset="reset"
      ><el-form-item :label="t('trade.tenantId')"
        ><TenantSelect v-model="query.tenantId" /></el-form-item
      ><el-form-item :label="t('trade.symbolId')"
        ><el-input-number v-model="query.symbolId" :min="0" /></el-form-item
      ><el-form-item :label="t('trade.snapshotKind')"
        ><el-input v-model="query.snapshotKind" clearable /></el-form-item></CrudQueryCard
    ><el-card shadow="never"
      ><el-table v-loading="loading" :data="rows" stripe
        ><el-table-column
          prop="snapshotId"
          :label="t('trade.snapshotId')"
          min-width="220"
          show-overflow-tooltip /><el-table-column
          prop="snapshotKind"
          :label="t('trade.snapshotKind')" /><el-table-column
          prop="symbolId"
          :label="t('trade.symbolId')" /><el-table-column
          prop="price"
          :label="t('trade.price')" /><el-table-column
          prop="markPrice"
          :label="t('trade.markPrice')" /><el-table-column
          prop="indexPrice"
          :label="t('trade.indexPrice')" /><el-table-column
          prop="fundingRate"
          :label="t('trade.fundingRate')" /><el-table-column
          prop="sourceTimestamp"
          :label="t('trade.sourceTimestamp')"
          min-width="150" /><el-table-column
          prop="revision"
          :label="t('trade.revision')" /><el-table-column
          prop="confirmed"
          :label="t('trade.confirmed')" /></el-table
    ></el-card>
  </div>
</template>
<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { apiTradeListMarketSnapshots } from '@/api/trade'
import type { TradeMarketSnapshot } from '@/services'
const { t } = useI18n(),
  loading = ref(false),
  rows = ref<TradeMarketSnapshot[]>([]),
  query = reactive({
    tenantId: undefined as number | undefined,
    symbolId: undefined as number | undefined,
    snapshotKind: '',
  })
async function load() {
  loading.value = true
  try {
    const r = await apiTradeListMarketSnapshots({ ...query, limit: 100 })
    rows.value = r.data || []
  } finally {
    loading.value = false
  }
}
function reset() {
  query.tenantId = undefined
  query.symbolId = undefined
  query.snapshotKind = ''
  load()
}
onMounted(load)
</script>
