<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="search" @reset="resetQuery">
      <el-form-item v-if="authStore.isSystemAdmin" :label="t('option.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <template #actions>
        <el-button
          v-perm="'option:trading-halt:create'"
          type="danger"
          @click="createVisible = true"
        >
          {{ t('option.haltTrading') }}
        </el-button>
      </template>
    </CrudQueryCard>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column
          v-if="authStore.isSystemAdmin"
          prop="tenantId"
          :label="t('option.tenantId')"
          width="110"
        />
        <el-table-column prop="haltNo" :label="t('option.haltNo')" min-width="180" />
        <el-table-column prop="contractId" :label="t('option.contractId')" width="110" />
        <el-table-column prop="source" :label="t('option.haltSource')" width="100" />
        <el-table-column prop="reason" :label="t('option.reason')" min-width="180" />
        <el-table-column :label="t('option.cancelResult')" min-width="150">
          <template #default="{ row }">
            {{ row.cancelSuccess }}/{{ row.cancelTotal }}
            <span v-if="row.cancelFailed" class="danger">({{ row.cancelFailed }})</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('option.startedAt')" min-width="170">
          <template #default="{ row }">{{ formatTime(row.startedAt) }}</template>
        </el-table-column>
        <el-table-column :label="t('common.status')" width="100">
          <template #default="{ row }">
            {{ row.status === 1 ? t('option.haltActive') : t('option.haltLifted') }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="100" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 1"
              v-perm="'option:trading-halt:resume'"
              link
              type="success"
              @click="resume(row)"
            >
              {{ t('option.resumeTrading') }}
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

    <el-dialog v-model="createVisible" :title="t('option.haltTrading')" width="520px">
      <el-form :model="form" label-width="120px">
        <el-form-item :label="t('option.contractId')">
          <el-input-number v-model="form.contractId" :min="1" />
        </el-form-item>
        <el-form-item :label="t('option.reason')">
          <el-input v-model="form.reason" type="textarea" />
        </el-form-item>
        <el-form-item :label="t('option.evidenceRef')">
          <el-input v-model="form.evidenceRef" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="danger" :loading="saving" @click="haltTrading">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'
import TenantSelect from '@/components/TenantSelect.vue'
import { usePagination } from '@/composables'
import { optionService, type OptionTradingHalt } from '@/services'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const authStore = useAuthStore()
const loading = ref(false)
const saving = ref(false)
const createVisible = ref(false)
const rows = ref<OptionTradingHalt[]>([])
const pagination = usePagination<number>(20)
const page = pagination.pagination
const query = reactive({
  tenantId: (authStore.isTenantUser ? authStore.profileTenantId : undefined) as number | undefined,
})
const form = reactive({ contractId: 1, reason: '', evidenceRef: '' })

async function loadRows() {
  loading.value = true
  try {
    const response = await optionService.listTradingHalts({
      tenantId: query.tenantId,
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
  pagination.reset()
  void loadRows()
}

async function haltTrading() {
  if (!query.tenantId) return
  saving.value = true
  try {
    await optionService.haltContractTrading({ tenantId: query.tenantId, ...form })
    ElMessage.success(t('common.success'))
    createVisible.value = false
    await loadRows()
  } finally {
    saving.value = false
  }
}

async function resume(row: OptionTradingHalt) {
  const { value } = await ElMessageBox.prompt(t('option.resumeReason'), t('option.resumeTrading'), {
    inputValidator: (text) => Boolean(text?.trim()),
  })
  await optionService.resumeContractTrading({
    tenantId: row.tenantId,
    haltId: row.id,
    reason: value,
  })
  ElMessage.success(t('common.success'))
  await loadRows()
}

const formatTime = (value?: number) => (value ? new Date(value * 1000).toLocaleString() : '-')

onMounted(loadRows)
</script>

<style scoped>
.danger {
  color: var(--el-color-danger);
}
</style>
