<template>
  <div class="module-page">
    <div class="page-header">
      <h2>{{ t('trade.cancelLogs') }}</h2>
      <div class="header-actions">
        <el-button @click="loadCurrent">
          {{ t('common.refresh') }}
        </el-button>
      </div>
    </div>

    <el-card shadow="never" class="query-card">
      <el-form :model="currentQuery" inline label-width="90px">
        <el-form-item v-for="field in currentFields" :key="field.key" :label="field.label">
          <el-input v-if="field.type !== 'number'" v-model="currentQuery[field.key]" clearable />

          <el-input-number v-else v-model="currentQuery[field.key]" :min="0" :precision="0" />
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
        <el-table-column
          v-for="column in currentColumns"
          :key="column.prop"
          :prop="column.prop"
          :label="column.label"
          :min-width="column.width || 140"
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
import { tradeService, type TradeCancelLog } from '@/services'

const { t } = useI18n()

interface CurrentQuery {
  tenantId: number | undefined
  userId: number | undefined
  orderNo: string
  limit: number
}

type CurrentField =
  | {
      key: 'tenantId' | 'userId'
      label: string
      type: 'number'
    }
  | {
      key: 'orderNo'
      label: string
      type?: 'text'
    }

interface CurrentColumn {
  prop: string
  label: string
  width?: number
}

const loading = ref(false)
const rows = ref<TradeCancelLog[]>([])
const detailVisible = ref(false)
const detailData = ref<TradeCancelLog | null>(null)

const currentQuery = reactive<CurrentQuery>({
  tenantId: undefined,
  userId: undefined,
  orderNo: '',
  limit: 100,
})

const currentFields: CurrentField[] = [
  { key: 'tenantId', label: t('trade.tenantId'), type: 'number' },
  { key: 'userId', label: t('trade.userId'), type: 'number' },
  { key: 'orderNo', label: t('trade.orderNo'), type: 'text' },
]

const currentColumns: CurrentColumn[] = [
  { prop: 'orderNo', label: t('trade.orderNo'), width: 180 },
  { prop: 'userId', label: t('trade.userId'), width: 100 },
  { prop: 'cancelSource', label: t('trade.cancelSource'), width: 100 },
  { prop: 'cancelReason', label: t('trade.cancelReason'), width: 200 },
]

const pickList = (res: any) => res?.data || res?.list || []

const loadCurrent = async () => {
  loading.value = true
  try {
    rows.value = pickList(await tradeService.listCancelLogs(currentQuery))
  } finally {
    loading.value = false
  }
}

const resetCurrent = () => {
  currentQuery.tenantId = undefined
  currentQuery.userId = undefined
  currentQuery.orderNo = ''
  currentQuery.limit = 100
  loadCurrent()
}

const showDetail = (row: TradeCancelLog) => {
  detailData.value = row
  detailVisible.value = true
}

onMounted(loadCurrent)
</script>

<style scoped>
.detail-pre {
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
