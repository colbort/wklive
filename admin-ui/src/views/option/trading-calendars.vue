<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="search" @reset="resetQuery">
      <el-form-item v-if="authStore.isSystemAdmin" :label="t('option.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('option.calendarCode')">
        <el-input v-model="query.calendarCode" clearable />
      </el-form-item>
      <el-form-item :label="t('common.status')">
        <el-select v-model="query.status" clearable>
          <el-option :label="t('option.calendarDraft')" :value="1" />
          <el-option :label="t('option.calendarApproved')" :value="2" />
          <el-option :label="t('option.calendarRejected')" :value="3" />
          <el-option :label="t('option.calendarSuperseded')" :value="4" />
        </el-select>
      </el-form-item>
      <template #actions>
        <el-button v-perm="'option:trading-calendar:create'" type="primary" @click="openCreate">
          {{ t('option.createCalendarVersion') }}
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
        <el-table-column prop="calendarCode" :label="t('option.calendarCode')" min-width="180" />
        <el-table-column prop="version" :label="t('option.version')" width="80" />
        <el-table-column prop="timezone" :label="t('option.timezone')" min-width="150" />
        <el-table-column :label="t('common.status')" width="110">
          <template #default="{ row }">{{ calendarStatus(row.status) }}</template>
        </el-table-column>
        <el-table-column :label="t('option.effectiveFrom')" min-width="170">
          <template #default="{ row }">{{ formatTime(row.effectiveFrom) }}</template>
        </el-table-column>
        <el-table-column :label="t('option.weeklySessions')" min-width="260">
          <template #default="{ row }">{{ sessionSummary(row.sessions) }}</template>
        </el-table-column>
        <el-table-column :label="t('option.exceptionWindows')" min-width="220">
          <template #default="{ row }">{{ exceptionSummary(row.exceptions) }}</template>
        </el-table-column>
        <el-table-column
          prop="changeReason"
          :label="t('option.calendarChangeReason')"
          min-width="180"
        />
        <el-table-column prop="evidenceRef" :label="t('option.evidenceRef')" min-width="180" />
        <el-table-column :label="t('common.actions')" width="160" fixed="right">
          <template #default="{ row }">
            <template v-if="row.status === 1">
              <el-button
                v-perm="'option:trading-calendar:review'"
                link
                type="success"
                @click="review(row, true)"
              >
                {{ t('option.approve') }}
              </el-button>
              <el-button
                v-perm="'option:trading-calendar:review'"
                link
                type="danger"
                @click="review(row, false)"
              >
                {{ t('option.reject') }}
              </el-button>
            </template>
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

    <el-dialog v-model="createVisible" :title="t('option.createCalendarVersion')" width="680px">
      <el-form :model="form" label-width="150px">
        <el-form-item :label="t('option.calendarCode')">
          <el-input v-model="form.calendarCode" placeholder="CONTINUOUS_24_7" />
        </el-form-item>
        <el-form-item :label="t('option.timezone')">
          <el-input v-model="form.timezone" placeholder="Asia/Hong_Kong" />
        </el-form-item>
        <el-form-item :label="t('option.effectiveFrom')">
          <el-date-picker v-model="form.effectiveFrom" type="datetime" />
        </el-form-item>
        <el-form-item :label="t('option.weeklySessions')">
          <el-input
            v-model="form.sessions"
            type="textarea"
            :rows="8"
            :placeholder="t('option.sessionFormatHint')"
          />
        </el-form-item>
        <el-form-item :label="t('option.exceptionWindows')">
          <el-input
            v-model="form.exceptions"
            type="textarea"
            :rows="5"
            :placeholder="t('option.exceptionFormatHint')"
          />
        </el-form-item>
        <el-form-item :label="t('option.calendarChangeReason')">
          <el-input v-model="form.changeReason" type="textarea" />
        </el-form-item>
        <el-form-item :label="t('option.evidenceRef')">
          <el-input v-model="form.evidenceRef" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="saving" @click="createCalendar">
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
import {
  optionService,
  type OptionTradingCalendar,
  type OptionTradingCalendarException,
  type OptionTradingCalendarSession,
  type TradingCalendarExceptionInput,
  type TradingCalendarSessionInput,
} from '@/services'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const authStore = useAuthStore()
const loading = ref(false)
const saving = ref(false)
const createVisible = ref(false)
const rows = ref<OptionTradingCalendar[]>([])
const pagination = usePagination<number>(20)
const page = pagination.pagination
const query = reactive({
  tenantId: (authStore.isTenantUser ? authStore.profileTenantId : undefined) as number | undefined,
  calendarCode: '',
  status: undefined as number | undefined,
})
const form = reactive({
  calendarCode: '',
  timezone: 'UTC',
  effectiveFrom: undefined as Date | undefined,
  sessions:
    '0,00:00,24:00\n1,00:00,24:00\n2,00:00,24:00\n3,00:00,24:00\n4,00:00,24:00\n5,00:00,24:00\n6,00:00,24:00',
  exceptions: '',
  changeReason: '',
  evidenceRef: '',
})

async function loadRows() {
  loading.value = true
  try {
    const response = await optionService.listTradingCalendars({
      tenantId: query.tenantId,
      calendarCode: query.calendarCode || undefined,
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
  query.calendarCode = ''
  query.status = undefined
  pagination.reset()
  void loadRows()
}

function openCreate() {
  form.effectiveFrom = new Date(Date.now() + 10 * 60 * 1000)
  createVisible.value = true
}

function parseClock(value: string): number {
  const match = /^(\d{1,2}):([0-5]\d)$/.exec(value.trim())
  if (!match) throw new Error(t('option.invalidSessionFormat'))
  const hour = Number(match[1])
  if (hour > 48 || (hour === 48 && match[2] !== '00')) {
    throw new Error(t('option.invalidSessionFormat'))
  }
  return hour * 3600 + Number(match[2]) * 60
}

function parseSessions(): TradingCalendarSessionInput[] {
  return form.sessions
    .split('\n')
    .filter((line) => line.trim())
    .map((line) => {
      const parts = line.split(',')
      if (parts.length !== 3) throw new Error(t('option.invalidSessionFormat'))
      return {
        weekday: Number(parts[0]),
        openSecond: parseClock(parts[1]),
        closeSecond: parseClock(parts[2]),
      }
    })
}

function parseExceptions(): TradingCalendarExceptionInput[] {
  return form.exceptions
    .split('\n')
    .filter((line) => line.trim())
    .map((line) => {
      const parts = line.split(',').map((part) => part.trim())
      if (parts.length < 4) throw new Error(t('option.invalidExceptionFormat'))
      const startTime = Math.floor(Date.parse(parts[1]) / 1000)
      const endTime = Math.floor(Date.parse(parts[2]) / 1000)
      const exceptionType =
        parts[0].toUpperCase() === 'CLOSED' ? 1 : parts[0].toUpperCase() === 'OPEN' ? 2 : 0
      if (!exceptionType || !Number.isFinite(startTime) || !Number.isFinite(endTime)) {
        throw new Error(t('option.invalidExceptionFormat'))
      }
      return {
        exceptionType,
        startTime,
        endTime,
        reason: parts[3],
        announcementRef: parts.slice(4).join(','),
      }
    })
}

async function createCalendar() {
  if (!query.tenantId) return
  saving.value = true
  try {
    await optionService.createTradingCalendar({
      tenantId: query.tenantId,
      calendarCode: form.calendarCode,
      timezone: form.timezone,
      effectiveFrom: form.effectiveFrom
        ? Math.floor(form.effectiveFrom.getTime() / 1000)
        : undefined,
      changeReason: form.changeReason,
      evidenceRef: form.evidenceRef,
      sessions: parseSessions(),
      exceptions: parseExceptions(),
    })
    ElMessage.success(t('common.success'))
    createVisible.value = false
    await loadRows()
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : String(error))
  } finally {
    saving.value = false
  }
}

async function review(row: OptionTradingCalendar, approve: boolean) {
  const { value } = await ElMessageBox.prompt(
    t('option.reviewReason'),
    t('option.calendarReview'),
    { inputValidator: (text) => Boolean(text?.trim()) },
  )
  await optionService.reviewTradingCalendar({
    tenantId: row.tenantId,
    calendarId: row.id,
    approve,
    reason: value,
  })
  ElMessage.success(t('common.success'))
  await loadRows()
}

const calendarStatus = (status: number) =>
  [
    '',
    t('option.calendarDraft'),
    t('option.calendarApproved'),
    t('option.calendarRejected'),
    t('option.calendarSuperseded'),
  ][status]
const formatTime = (value?: number) => (value ? new Date(value * 1000).toLocaleString() : '-')
const clock = (value: number) => {
  const hour = Math.floor(value / 3600)
  const minute = Math.floor((value % 3600) / 60)
  return `${String(hour).padStart(2, '0')}:${String(minute).padStart(2, '0')}`
}
const sessionSummary = (sessions: OptionTradingCalendarSession[]) =>
  (sessions || [])
    .map((item) => `${item.weekday} ${clock(item.openSecond)}-${clock(item.closeSecond)}`)
    .join('; ')
const exceptionSummary = (exceptions: OptionTradingCalendarException[]) =>
  (exceptions || [])
    .map(
      (item) =>
        `${item.exceptionType === 1 ? 'CLOSED' : 'OPEN'} ${formatTime(item.startTime)}-${formatTime(item.endTime)}`,
    )
    .join('; ')

onMounted(loadRows)
</script>
