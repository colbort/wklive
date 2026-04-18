<template>
  <div class="module-page">
    <div class="page-header">
      <h2>{{ t('trade.marginAccounts') }}</h2>
      <div class="header-actions">
        <el-button @click="loadCurrent">
          {{ t('common.refresh') }}
        </el-button>
      </div>
    </div>

    <el-card shadow="never" class="query-card">
      <el-form
        :model="currentQuery"
        inline
        label-width="90px"
      >
        <el-form-item
          v-for="field in currentFields"
          :key="field.key"
          :label="field.label"
        >
          <el-input
            v-if="field.type !== 'number'"
            v-model="currentQuery[field.key]"
            clearable
          />

          <el-input-number
            v-else
            v-model="currentQuery[field.key]"
            :min="0"
            :precision="0"
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            @click="loadCurrent"
          >
            {{ t('common.search') }}
          </el-button>
          <el-button @click="resetCurrent">
            {{ t('common.reset') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" class="table-card">
      <el-table
        v-loading="loading"
        :data="rows"
        stripe
      >
        <el-table-column
          v-for="column in currentColumns"
          :key="column.prop"
          :prop="column.prop"
          :label="column.label"
          :min-width="column.width || 140"
          show-overflow-tooltip
        />

        <el-table-column
          :label="t('common.actions')"
          width="100"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              link
              type="primary"
              @click="showDetail(row)"
            >
              {{ t('option.detail') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog
      v-model="detailVisible"
      :title="t('option.detail')"
      width="760px"
    >
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { tradeService } from '@/services'

const { t } = useI18n()

interface CurrentQuery {
  tenantId: number | undefined
  userId: number | undefined
  marketType: number | undefined
  limit: number
}

interface CurrentField {
  key: 'tenantId' | 'userId' | 'marketType'
  label: string
  type: 'number'
}

interface CurrentColumn {
  prop: string
  label: string
  width?: number
}

const loading = ref(false)
const rows = ref<Record<string, any>[]>([])
const detailVisible = ref(false)
const detailData = ref<Record<string, any>>({})

const currentQuery = reactive<CurrentQuery>({
  tenantId: undefined,
  userId: undefined,
  marketType: undefined,
  limit: 100,
})

const currentFields: CurrentField[] = [
  { key: 'tenantId', label: t('trade.tenantId'), type: 'number' },
  { key: 'userId', label: t('trade.userId'), type: 'number' },
  { key: 'marketType', label: t('trade.marketType'), type: 'number' },
]

const currentColumns: CurrentColumn[] = [
  { prop: 'userId', label: t('trade.userId'), width: 100 },
  { prop: 'marketType', label: t('trade.marketType'), width: 100 },
  { prop: 'marginAsset', label: t('trade.marginAsset'), width: 120 },
  { prop: 'balance', label: t('trade.balance') },
  { prop: 'availableBalance', label: t('trade.availableBalance') },
]

const pickList = (res: any) => res?.data || res?.list || []

const loadCurrent = async () => {
  loading.value = true
  try {
    rows.value = pickList(await tradeService.listMarginAccounts(currentQuery))
  } finally {
    loading.value = false
  }
}

const resetCurrent = () => {
  currentQuery.tenantId = undefined
  currentQuery.userId = undefined
  currentQuery.marketType = undefined
  currentQuery.limit = 100
  loadCurrent()
}

const showDetail = (row: Record<string, any>) => {
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
