<template>
  <div class="module-page snapshot-outbox-page">
    <CrudQueryCard :model="query" @search="load" @reset="reset">
      <el-form-item class="operation-query-item" :label="t('itick.snapshotId')">
        <el-input v-model="query.snapshotId" clearable class="snapshot-id-control" />
      </el-form-item>
      <el-form-item class="operation-query-item" :label="t('common.status')">
        <el-select v-model="query.status" clearable class="status-control">
          <el-option
            v-for="item in statuses"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
    </CrudQueryCard>
    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column prop="id" :label="t('common.id')" width="80" />
        <el-table-column
          prop="snapshotId"
          :label="t('itick.snapshotId')"
          min-width="240"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.status')" width="120">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)">
              {{ statusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="retryCount" :label="t('itick.retryCount')" width="100" />
        <el-table-column :label="t('itick.redisPublishedAt')" min-width="180">
          <template #default="{ row }">
            {{ formatTime(row.redisPublishedAt) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('itick.optionPublishedAt')" min-width="180">
          <template #default="{ row }">
            {{ formatTime(row.optionPublishedAt) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="lastErrorMsg"
          :label="t('itick.lastErrorMsg')"
          min-width="220"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.updateTimes')" min-width="180">
          <template #default="{ row }">
            {{ formatTime(row.updateTimes) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="190" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 4 || row.status === 5"
              v-perm="'itick:snapshot-outbox:retry'"
              link
              type="warning"
              @click="retry(row)"
            >
              {{ t('itick.retry') }}
            </el-button>
            <el-button
              v-perm="'itick:authoritative-snapshot:revoke'"
              link
              type="danger"
              @click="openRevoke(row)"
            >
              {{ t('itick.revoke') }}
            </el-button>
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
    <el-dialog v-model="revokeVisible" :title="t('itick.revokeSnapshot')" width="560px">
      <el-form :model="revokeForm" label-width="180px">
        <el-form-item :label="t('itick.snapshotId')">
          <el-input v-model="revokeForm.snapshotId" disabled />
        </el-form-item>
        <el-form-item :label="t('itick.replacementSnapshotId')">
          <el-input v-model="revokeForm.replacementSnapshotId" />
        </el-form-item>
        <el-form-item :label="t('itick.reason')" required>
          <el-input v-model="revokeForm.reason" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="revokeVisible = false"> {{ t('common.cancel') }} </el-button
        ><el-button type="danger" :loading="revoking" @click="revoke">
          {{ t('itick.revoke') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'
import { usePagination } from '@/composables'
import {
  apiListSnapshotOutbox,
  apiRetrySnapshotOutbox,
  apiRevokeAuthoritativeSnapshot,
  type SnapshotOutbox,
} from '@/api/itick/price-engine'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)
const loading = ref(false),
  revokeVisible = ref(false),
  revoking = ref(false),
  rows = ref<SnapshotOutbox[]>([])
const query = reactive({ status: undefined as number | undefined, snapshotId: '' })
const revokeForm = reactive({ snapshotId: '', replacementSnapshotId: '', reason: '' })
const statuses = computed(() =>
  [1, 2, 3, 4, 5].map((value) => ({ value, label: t(`itick.outboxStatus${value}`) })),
)
function statusLabel(status: number) {
  return statuses.value.find((item) => item.value === status)?.label || String(status)
}
function statusType(status: number) {
  return status === 3
    ? 'success'
    : status === 4 || status === 5
      ? 'danger'
      : status === 2
        ? 'warning'
        : 'info'
}
function formatTime(value: number) {
  return value > 0 ? new Date(value).toLocaleString() : '-'
}
async function load() {
  loading.value = true
  try {
    const res = await apiListSnapshotOutbox({
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
function reset() {
  query.status = undefined
  query.snapshotId = ''
  resetAndLoad(load)
}
async function retry(row: SnapshotOutbox) {
  await ElMessageBox.confirm(t('itick.retryOutboxConfirm'))
  await apiRetrySnapshotOutbox(row.id)
  ElMessage.success(t('common.success'))
  load()
}
function openRevoke(row: SnapshotOutbox) {
  Object.assign(revokeForm, { snapshotId: row.snapshotId, replacementSnapshotId: '', reason: '' })
  revokeVisible.value = true
}
async function revoke() {
  if (!revokeForm.reason.trim()) {
    ElMessage.warning(t('itick.revokeReasonRequired'))
    return
  }
  revoking.value = true
  try {
    await apiRevokeAuthoritativeSnapshot(revokeForm)
    ElMessage.success(t('common.success'))
    revokeVisible.value = false
    load()
  } finally {
    revoking.value = false
  }
}
onMounted(load)
</script>
