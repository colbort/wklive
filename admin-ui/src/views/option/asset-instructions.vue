<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="search" @reset="resetQuery">
      <el-form-item v-if="authStore.isSystemAdmin" :label="t('option.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('option.bizNo')">
        <el-input v-model="query.bizNo" clearable />
      </el-form-item>
      <el-form-item :label="t('option.instructionStatus')">
        <el-select v-model="query.status" clearable>
          <el-option :label="t('option.pending')" :value="1" />
          <el-option :label="t('option.failed')" :value="4" />
          <el-option :label="t('option.manualReview')" :value="5" />
        </el-select>
      </el-form-item>
    </CrudQueryCard>

    <el-card v-loading="loading" shadow="never" class="table-card">
      <el-table :data="rows" stripe>
        <el-table-column
          v-if="authStore.isSystemAdmin"
          prop="tenantId"
          :label="t('option.tenantId')"
          width="110"
        />
        <el-table-column prop="instructionNo" :label="t('option.instructionNo')" min-width="220" />
        <el-table-column prop="bizNo" :label="t('option.bizNo')" min-width="180" />
        <el-table-column prop="userId" :label="t('option.userId')" width="100" />
        <el-table-column prop="coin" :label="t('option.coin')" width="90" />
        <el-table-column prop="amount" :label="t('option.amount')" min-width="130" />
        <el-table-column prop="stepNo" :label="t('option.stepNo')" width="80" />
        <el-table-column prop="status" :label="t('common.status')" width="90" />
        <el-table-column prop="retryCount" :label="t('option.retryCount')" width="90" />
        <el-table-column
          prop="lastErrorMsg"
          :label="t('option.lastErrorMsg')"
          min-width="220"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.actions')" width="100" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="[4, 5].includes(row.status) && !row.deliveryUnitId"
              v-perm="'option:operations:asset-retry'"
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
        v-model:limit="page.limit"
        :total="page.total"
        :has-prev="page.hasPrev"
        :has-next="page.hasNext"
        @prev="pagination.prevAndLoad(loadRows)"
        @next="pagination.nextAndLoad(loadRows)"
        @limit-change="pagination.resetAndLoad(loadRows)"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'
import TenantSelect from '@/components/TenantSelect.vue'
import { usePagination } from '@/composables'
import { optionService, type OptionAssetInstruction } from '@/services'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const authStore = useAuthStore()
const loading = ref(false)
const rows = ref<OptionAssetInstruction[]>([])
const pagination = usePagination<number>(20)
const page = pagination.pagination
const query = reactive({
  tenantId: (authStore.isTenantUser ? authStore.profileTenantId : undefined) as number | undefined,
  bizNo: '',
  status: undefined as number | undefined,
})

async function loadRows() {
  loading.value = true
  try {
    const response = await optionService.listAssetInstructions({
      tenantId: query.tenantId,
      bizNo: query.bizNo || undefined,
      status: query.status,
      cursor: page.cursor,
      limit: page.limit,
    })
    rows.value = response.data || []
    pagination.updateFromResponse(response)
  } finally {
    loading.value = false
  }
}

function search() {
  pagination.reset()
  void loadRows()
}

function resetQuery() {
  query.tenantId = authStore.isTenantUser ? authStore.profileTenantId || undefined : undefined
  query.bizNo = ''
  query.status = undefined
  pagination.reset()
  void loadRows()
}

async function retry(row: OptionAssetInstruction) {
  const { value } = await ElMessageBox.prompt(t('option.retryReason'), t('option.retry'), {
    inputType: 'textarea',
    inputValidator: (input) => {
      const reason = input?.trim() || ''
      if (!reason) return t('option.retryReasonRequired')
      return Array.from(reason).length <= 64 || t('option.retryReasonTooLong')
    },
  })
  await optionService.retryAssetInstruction({
    tenantId: row.tenantId,
    instructionId: row.id,
    reason: value.trim(),
  })
  ElMessage.success(t('common.success'))
  await loadRows()
}

onMounted(loadRows)
</script>
