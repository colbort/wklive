<template>
  <div class="module-page">
    <div class="page-header">
      <h2>{{ t('option.trades') }}</h2>
      <div class="header-actions">
        <el-button @click="loadCurrent">{{ t('common.refresh') }}</el-button>
      </div>
    </div>

    <el-card shadow="never" class="query-card">
      <el-form :model="query" inline label-width="90px">
        <el-form-item :label="t('option.tenantId')">
          <el-input-number v-model="query.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('option.contractId')">
          <el-input-number v-model="query.contractId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('option.tradeNo')">
          <el-input v-model="query.tradeNo" clearable />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadCurrent">{{ t('common.search') }}</el-button>
          <el-button @click="resetCurrent">{{ t('common.reset') }}</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column :label="t('option.tradeNo')" prop="tradeNo" min-width="180" show-overflow-tooltip />
        <el-table-column :label="t('option.contractId')" prop="contractId" width="100" />
        <el-table-column :label="t('option.price')" prop="price" min-width="120" show-overflow-tooltip />
        <el-table-column :label="t('option.qty')" prop="qty" min-width="120" show-overflow-tooltip />
        <el-table-column :label="t('option.tradeTime')" prop="tradeTime" min-width="160" show-overflow-tooltip />
        <el-table-column :label="t('common.actions')" width="100" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">{{ t('option.detail') }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="detailVisible" :title="t('option.detail')" width="760px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { optionService } from '@/services'

const { t } = useI18n()

const loading = ref(false)
const rows = ref<Record<string, any>[]>([])
const detailVisible = ref(false)
const detailData = ref<Record<string, any>>({})
const query = reactive({
  tenantId: undefined as number | undefined,
  contractId: undefined as number | undefined,
  tradeNo: '',
  limit: 100,
})

const pickList = (res: any) => res?.data || res?.list || []

const loadCurrent = async () => {
  loading.value = true
  try {
    rows.value = pickList(await optionService.listTrades(query))
  } finally {
    loading.value = false
  }
}

const resetCurrent = () => {
  query.tenantId = undefined
  query.contractId = undefined
  query.tradeNo = ''
  query.limit = 100
  loadCurrent()
}

const showDetail = async (row: Record<string, any>) => {
  detailData.value = (await optionService.getTrade({ tenantId: row.tenantId, id: row.id, tradeNo: row.tradeNo })).data || row
  detailVisible.value = true
}

onMounted(loadCurrent)
</script>

<style scoped></style>
