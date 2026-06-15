<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { cronJobService } from '@/services'
import type { SysCronJobLogItem } from '@/services/system/CronJobService'
import { usePagination } from '@/composables/usePagination'
import { useLoading } from '@/composables/useLoading'
import { useForm } from '@/composables/useForm'
import { formatDate } from '@/utils'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()

// Pagination and list
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)
const { loading, withLoading } = useLoading()

// Query form
const { form: queryForm } = useForm({
  initialData: {
    jobName: '',
    invokeTarget: '',
    status: undefined as number | undefined,
  },
})

const list_ref = ref<SysCronJobLogItem[]>([])

// Detail drawer
const detailDrawerVisible = ref(false)
const detailData = ref<SysCronJobLogItem | null>(null)

async function fetchList() {
  await withLoading(async () => {
    try {
      const res = await cronJobService.getLogList({
        jobName: queryForm.jobName || undefined,
        invokeTarget: queryForm.invokeTarget || undefined,
        status: queryForm.status,
        cursor: pagination.cursor,
        limit: pagination.limit,
      })
      if (res.code !== 200) throw new Error(res.msg)
      list_ref.value = res.data || []
      updateFromResponse(res)
    } catch (error: unknown) {
      ElMessage.error(error instanceof Error ? error.message : t('common.loadFailed'))
    }
  })
}

function loadList() {
  resetAndLoad(fetchList)
}

function resetQuery() {
  queryForm.jobName = ''
  queryForm.invokeTarget = ''
  queryForm.status = undefined
  resetAndLoad(fetchList)
}

function nextPage() {
  nextAndLoad(fetchList)
}

function prevPage() {
  prevAndLoad(fetchList)
}

function showDetail(row: SysCronJobLogItem) {
  detailData.value = row
  detailDrawerVisible.value = true
}

// Calculate duration
const calculateDuration = (row: SysCronJobLogItem) => {
  if (!row.endTime || !row.startTime) return '-'
  return `${row.endTime - row.startTime}ms`
}

onMounted(() => {
  fetchList()
})
</script>

<template>
  <div class="module-page">
    <CrudQueryCard
      :model="queryForm"
      label-width="auto"
      @search="loadList"
      @reset="resetQuery"
    >
      <el-form-item :label="t('system.jobName')">
        <el-input v-model="queryForm.jobName" :placeholder="t('common.pleaseEnter')" clearable />
      </el-form-item>

      <el-form-item :label="t('system.invokeTarget')">
        <el-input
          v-model="queryForm.invokeTarget"
          :placeholder="t('common.pleaseEnter')"
          clearable
        />
      </el-form-item>

      <el-form-item :label="t('common.status')">
        <el-select v-model="queryForm.status" :placeholder="t('common.pleaseSelect')" clearable>
          <el-option :label="t('common.success')" :value="1" />
          <el-option :label="t('common.failed')" :value="0" />
        </el-select>
      </el-form-item>
    </CrudQueryCard>

    <el-card class="table-card">
      <!-- Table -->
      <el-table
        v-loading="loading"
        :data="list_ref"
        row-key="id"
        style="margin-bottom: 16px"
      >
        <el-table-column prop="id" :label="t('common.id')" width="70" />
        <el-table-column prop="jobId" :label="t('system.jobId')" width="80" />
        <el-table-column prop="jobName" :label="t('system.jobName')" min-width="120" />
        <el-table-column
          prop="invokeTarget"
          :label="t('system.invokeTarget')"
          min-width="140"
          show-overflow-tooltip
        />
        <el-table-column prop="status" :label="t('common.status')" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? t('common.success') : t('common.failed') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="startTime" :label="t('system.startTime')" width="170">
          <template #default="{ row }">
            <span>{{ formatDate(row.startTime) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="endTime" :label="t('system.endTime')" width="170">
          <template #default="{ row }">
            <span>{{ formatDate(row.endTime) }}</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('system.duration')" width="100">
          <template #default="{ row }">
            <span>{{ calculateDuration(row) }}</span>
          </template>
        </el-table-column>
        <el-table-column
          prop="message"
          :label="t('system.message')"
          min-width="140"
          show-overflow-tooltip
        />
        <el-table-column
          :label="t('common.actions')"
          align="center"
          width="110"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="showDetail(row)">
              {{ t('system.viewDetail') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- Pagination -->
      <CursorPagination
        v-model:limit="pagination.limit"
        :total="pagination.total"
        :has-prev="pagination.hasPrev"
        :has-next="pagination.hasNext"
        @prev="prevPage"
        @next="nextPage"
        @limit-change="
          () => {
            resetAndLoad(fetchList)
          }
        "
      />

      <!-- Detail Drawer -->
      <el-drawer v-model="detailDrawerVisible" :title="t('system.logDetail')" size="50%">
        <div v-if="detailData" style="padding: 0 20px">
          <el-descriptions :column="1" border>
            <el-descriptions-item :label="t('common.id')">
              {{ detailData.id }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('system.jobId')">
              {{ detailData.jobId }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('system.jobName')">
              {{ detailData.jobName }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('system.invokeTarget')">
              {{ detailData.invokeTarget }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('system.cronExpression')">
              {{ detailData.cronExpression }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('common.status')">
              <el-tag :type="detailData.status === 1 ? 'success' : 'danger'">
                {{ detailData.status === 1 ? t('common.success') : t('common.failed') }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item :label="t('system.startTime')">
              {{ formatDate(detailData.startTime) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('system.endTime')">
              {{ formatDate(detailData.endTime) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('system.duration')">
              {{ calculateDuration(detailData) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('system.message')">
              <div
                style="
                  word-break: break-all;
                  white-space: pre-wrap;
                  max-height: 200px;
                  overflow-y: auto;
                "
              >
                {{ detailData.message || '-' }}
              </div>
            </el-descriptions-item>
            <el-descriptions-item :label="t('system.exceptionInfo')">
              <div
                v-if="detailData.exceptionInfo"
                style="
                  word-break: break-all;
                  white-space: pre-wrap;
                  max-height: 300px;
                  overflow-y: auto;
                  background-color: #f5f5f5;
                  padding: 8px;
                  border-radius: 4px;
                  color: #d32f2f;
                "
              >
                {{ detailData.exceptionInfo }}
              </div>
              <div v-else style="color: #999">
                -
              </div>
            </el-descriptions-item>
            <el-descriptions-item :label="t('common.createTimes')">
              {{ formatDate(detailData.createTimes) }}
            </el-descriptions-item>
          </el-descriptions>
        </div>
      </el-drawer>
    </el-card>
  </div>
</template>
