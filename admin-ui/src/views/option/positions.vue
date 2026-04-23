<template>
  <div class="module-page">
    <div class="page-header">
      <h2>{{ t('option.positions') }}</h2>
      <div class="header-actions">
        <el-button @click="loadCurrent">
          {{ t('common.refresh') }}
        </el-button>
      </div>
    </div>

    <el-card shadow="never" class="query-card">
      <el-form :model="query" inline label-width="90px">
        <el-form-item :label="t('option.tenantId')">
          <el-input-number v-model="query.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('option.userId')">
          <el-input-number v-model="query.uid" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('option.contractId')">
          <el-input-number v-model="query.contractId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadCurrent">
            {{ t('common.search') }}
          </el-button>
          <el-button @click="resetCurrent">
            {{ t('common.reset') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column :label="t('option.userId')" prop="uid" width="100" />
        <el-table-column :label="t('option.contractId')" prop="contractId" width="100" />
        <el-table-column :label="t('option.side')" prop="side" width="100" />
        <el-table-column
          :label="t('option.positionQty')"
          prop="positionQty"
          min-width="120"
          show-overflow-tooltip
        />
        <el-table-column
          :label="t('option.unrealizedPnl')"
          prop="unrealizedPnl"
          min-width="120"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.actions')" width="100" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">
              {{ t('option.detail') }}
            </el-button>
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
import { optionService, type OptionPosition, type OptionPositionDetail } from '@/services'

const { t } = useI18n()

const loading = ref(false)
const rows = ref<OptionPosition[]>([])
const detailVisible = ref(false)
const detailData = ref<OptionPositionDetail | OptionPosition | null>(null)
const query = reactive({
  tenantId: undefined as number | undefined,
  uid: undefined as number | undefined,
  contractId: undefined as number | undefined,
  limit: 100,
})

const pickList = (res: any) => res?.data || res?.list || []

const loadCurrent = async () => {
  loading.value = true
  try {
    rows.value = pickList(await optionService.listPositions(query))
  } finally {
    loading.value = false
  }
}

const resetCurrent = () => {
  query.tenantId = undefined
  query.uid = undefined
  query.contractId = undefined
  query.limit = 100
  loadCurrent()
}

const showDetail = async (row: OptionPosition) => {
  detailData.value =
    (await optionService.getPosition({ tenantId: row.tenantId, id: row.id })).data || row
  detailVisible.value = true
}

onMounted(loadCurrent)
</script>

<style scoped></style>
