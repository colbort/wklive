<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="search" @reset="resetQuery">
      <el-form-item v-if="authStore.isSystemAdmin" :label="t('option.tenantId')">
        <TenantSelect v-model="query.tenantId" />
      </el-form-item>
      <el-form-item :label="t('option.userId')">
        <UserSelect v-model="query.userId" :tenant-id="query.tenantId" />
      </el-form-item>
      <el-form-item :label="t('option.contractId')">
        <ContractSelect v-model="query.contractId" :tenant-id="query.tenantId" />
      </el-form-item>
    </CrudQueryCard>

    <el-card v-loading="loading" shadow="never" class="table-card">
      <el-descriptions
        v-if="userControl"
        :column="4"
        border
        class="control-summary"
      >
        <el-descriptions-item :label="t('option.userId')">
          {{ userControl.userId }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('option.killSwitch')">
          <el-tag :type="userControl.killSwitch === 1 ? 'danger' : 'success'">
            {{ userControl.killSwitch === 1 ? t('common.enabled') : t('common.disabled') }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="t('option.controlReason')">
          {{ userControl.reason || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.actions')">
          <el-button
            v-if="userControl.killSwitch === 1"
            v-perm="'option:trading-control:release'"
            link
            type="warning"
            @click="releaseKillSwitch"
          >
            {{ t('option.releaseKillSwitch') }}
          </el-button>
        </el-descriptions-item>
      </el-descriptions>

      <el-table :data="events" stripe>
        <el-table-column prop="id" label="ID" width="90" />
        <el-table-column
          v-if="authStore.isSystemAdmin"
          prop="tenantId"
          :label="t('option.tenantId')"
          width="110"
        />
        <el-table-column prop="eventType" :label="t('option.controlEventType')" min-width="180" />
        <el-table-column prop="reason" :label="t('option.controlReason')" min-width="180" />
        <el-table-column prop="userId" :label="t('option.userId')" width="100" />
        <el-table-column prop="contractId" :label="t('option.contractId')" width="110" />
        <el-table-column prop="orderId" :label="t('option.orderId')" width="110" />
        <el-table-column
          prop="detail"
          :label="t('option.controlDetail')"
          min-width="280"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.createTimes')" width="180">
          <template #default="{ row }">
            {{ formatTime(row.createTimes) }}
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
import {
  optionService,
  type OptionTradingControlEvent,
  type OptionUserTradingControl,
} from '@/services'
import { usePagination } from '@/composables'
import { useAuthStore } from '@/stores/auth'
import TenantSelect from '@/components/TenantSelect.vue'
import UserSelect from '@/components/UserSelect.vue'
import ContractSelect from '@/components/ContractSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()
const authStore = useAuthStore()
const loading = ref(false)
const events = ref<OptionTradingControlEvent[]>([])
const userControl = ref<OptionUserTradingControl | null>(null)
const pagination = usePagination<number>(20)
const query = reactive({
  tenantId: (authStore.isTenantUser ? authStore.profileTenantId : undefined) as number | undefined,
  userId: undefined as number | undefined,
  contractId: undefined as number | undefined,
})

async function loadData() {
  loading.value = true
  userControl.value = null
  try {
    const response = await optionService.listTradingControlEvents({
      ...query,
      cursor: pagination.pagination.cursor,
      limit: pagination.pagination.limit,
    })
    events.value = response.data || []
    pagination.updateFromResponse(response)
    if (query.tenantId && query.userId) {
      const detail = await optionService.getUserTradingControl({
        tenantId: query.tenantId,
        userId: query.userId,
      })
      userControl.value = detail.data || null
    }
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
  query.contractId = undefined
  pagination.resetAndLoad(loadData)
}

async function releaseKillSwitch() {
  if (!userControl.value) return
  const { value } = await ElMessageBox.prompt(
    t('option.releaseKillSwitchReason'),
    t('option.releaseKillSwitch'),
    {
      inputValidator: (input) => Boolean(input?.trim()) || t('option.releaseReasonRequired'),
      type: 'warning',
    },
  )
  await optionService.releaseUserKillSwitch({
    tenantId: userControl.value.tenantId,
    userId: userControl.value.userId,
    reason: value.trim(),
  })
  ElMessage.success(t('common.success'))
  await loadData()
}

function formatTime(timestamp: number) {
  return timestamp ? new Date(timestamp * 1000).toLocaleString() : '-'
}

onMounted(loadData)
</script>

<style scoped>
.control-summary {
  margin-bottom: 16px;
}
</style>
