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
        <el-select v-model="query.status" clearable style="width: 150px">
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
        <el-button v-perm="'option:trading-halt:create'" type="danger" @click="haltVisible = true">
          {{ t('option.haltTrading') }}
        </el-button>
      </template>
    </CrudQueryCard>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="calendars" stripe>
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
          <template #default="{ row }">
            {{ calendarStatus(row.status) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('option.effectiveFrom')" min-width="170">
          <template #default="{ row }">
            {{ formatTime(row.effectiveFrom) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('option.weeklySessions')" min-width="260">
          <template #default="{ row }">
            {{ sessionSummary(row.sessions) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('option.exceptionWindows')" min-width="220">
          <template #default="{ row }">
            {{ exceptionSummary(row.exceptions) }}
          </template>
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
        v-model:limit="calendarPagination.pagination.limit"
        :total="calendarPagination.pagination.total"
        :has-prev="calendarPagination.pagination.hasPrev"
        :has-next="calendarPagination.pagination.hasNext"
        @prev="calendarPagination.prevAndLoad(loadCalendars)"
        @next="calendarPagination.nextAndLoad(loadCalendars)"
        @limit-change="calendarPagination.resetAndLoad(loadCalendars)"
      />
    </el-card>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="halts" stripe>
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
          <template #default="{ row }">
            {{ formatTime(row.startedAt) }}
          </template>
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
        v-model:limit="haltPagination.pagination.limit"
        :total="haltPagination.pagination.total"
        :has-prev="haltPagination.pagination.hasPrev"
        :has-next="haltPagination.pagination.hasNext"
        @prev="haltPagination.prevAndLoad(loadHalts)"
        @next="haltPagination.nextAndLoad(loadHalts)"
        @limit-change="haltPagination.resetAndLoad(loadHalts)"
      />
    </el-card>

    <el-dialog v-model="createVisible" :title="t('option.createCalendarVersion')" width="680px">
      <el-form :model="calendarForm" label-width="150px">
        <el-form-item :label="t('option.calendarCode')">
          <el-input v-model="calendarForm.calendarCode" placeholder="CONTINUOUS_24_7" />
        </el-form-item>
        <el-form-item :label="t('option.timezone')">
          <el-input v-model="calendarForm.timezone" placeholder="Asia/Hong_Kong" />
        </el-form-item>
        <el-form-item :label="t('option.effectiveFrom')">
          <el-date-picker v-model="calendarForm.effectiveFrom" type="datetime" />
        </el-form-item>
        <el-form-item :label="t('option.weeklySessions')">
          <el-input
            v-model="calendarForm.sessions"
            type="textarea"
            :rows="8"
            :placeholder="t('option.sessionFormatHint')"
          />
        </el-form-item>
        <el-form-item :label="t('option.exceptionWindows')">
          <el-input
            v-model="calendarForm.exceptions"
            type="textarea"
            :rows="5"
            :placeholder="t('option.exceptionFormatHint')"
          />
        </el-form-item>
        <el-form-item :label="t('option.calendarChangeReason')">
          <el-input v-model="calendarForm.changeReason" type="textarea" />
        </el-form-item>
        <el-form-item :label="t('option.evidenceRef')">
          <el-input v-model="calendarForm.evidenceRef" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button type="primary" :loading="saving" @click="createCalendar">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="haltVisible" :title="t('option.haltTrading')" width="520px">
      <el-form :model="haltForm" label-width="120px">
        <el-form-item :label="t('option.contractId')">
          <el-input-number v-model="haltForm.contractId" :min="1" />
        </el-form-item>
        <el-form-item :label="t('option.reason')">
          <el-input v-model="haltForm.reason" type="textarea" />
        </el-form-item>
        <el-form-item :label="t('option.evidenceRef')">
          <el-input v-model="haltForm.evidenceRef" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="haltVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button type="danger" :loading="saving" @click="haltTrading">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  optionService,
  type OptionTradingCalendar,
  type OptionTradingCalendarSession,
  type OptionTradingCalendarException,
  type OptionTradingHalt,
  type TradingCalendarExceptionInput,
  type TradingCalendarSessionInput,
} from '@/services'
import TenantSelect from '@/components/TenantSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'
import { useAuthStore } from '@/stores/auth'
import { usePagination } from '@/composables'

const { t } = useI18n()
const authStore = useAuthStore()
const loading = ref(false)
const saving = ref(false)
const createVisible = ref(false)
const haltVisible = ref(false)
const calendars = ref<OptionTradingCalendar[]>([])
const halts = ref<OptionTradingHalt[]>([])
const calendarPagination = usePagination<number>(20)
const haltPagination = usePagination<number>(20)
const query = reactive({
  tenantId: (authStore.isTenantUser ? authStore.profileTenantId : undefined) as number | undefined,
  calendarCode: '',
  status: undefined as number | undefined,
})
const calendarForm = reactive({
  calendarCode: '',
  timezone: 'UTC',
  effectiveFrom: undefined as Date | undefined,
  sessions:
    '0,00:00,24:00\n1,00:00,24:00\n2,00:00,24:00\n3,00:00,24:00\n4,00:00,24:00\n5,00:00,24:00\n6,00:00,24:00',
  exceptions: '',
  changeReason: '',
  evidenceRef: '',
})
const haltForm = reactive({ contractId: 1, reason: '', evidenceRef: '' })

function resetQuery() {
  query.tenantId = authStore.isTenantUser ? authStore.profileTenantId || undefined : undefined
  query.calendarCode = ''
  query.status = undefined
  calendarPagination.reset()
  haltPagination.reset()
  void refresh()
}

function search() {
  calendarPagination.reset()
  haltPagination.reset()
  void refresh()
}

async function loadCalendars() {
  const response = await optionService.listTradingCalendars({
    tenantId: query.tenantId,
    calendarCode: query.calendarCode || undefined,
    status: query.status,
    cursor: calendarPagination.pagination.cursor,
    limit: calendarPagination.pagination.limit,
  })
  calendars.value = response.data || []
  calendarPagination.updateFromResponse(response)
}

async function loadHalts() {
  const response = await optionService.listTradingHalts({
    tenantId: query.tenantId,
    cursor: haltPagination.pagination.cursor,
    limit: haltPagination.pagination.limit,
  })
  halts.value = response.data || []
  haltPagination.updateFromResponse(response)
}

async function refresh() {
  loading.value = true
  try {
    await Promise.all([loadCalendars(), loadHalts()])
  } finally {
    loading.value = false
  }
}

function openCreate() {
  calendarForm.effectiveFrom = new Date(Date.now() + 10 * 60 * 1000)
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
  return calendarForm.sessions
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
  return calendarForm.exceptions
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
      calendarCode: calendarForm.calendarCode,
      timezone: calendarForm.timezone,
      effectiveFrom: calendarForm.effectiveFrom
        ? Math.floor(calendarForm.effectiveFrom.getTime() / 1000)
        : undefined,
      changeReason: calendarForm.changeReason,
      evidenceRef: calendarForm.evidenceRef,
      sessions: parseSessions(),
      exceptions: parseExceptions(),
    })
    ElMessage.success(t('common.success'))
    createVisible.value = false
    await refresh()
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
    {
      inputValidator: (text) => Boolean(text?.trim()),
    },
  )
  await optionService.reviewTradingCalendar({
    tenantId: row.tenantId,
    calendarId: row.id,
    approve,
    reason: value,
  })
  ElMessage.success(t('common.success'))
  await refresh()
}

async function haltTrading() {
  if (!query.tenantId) return
  saving.value = true
  try {
    await optionService.haltContractTrading({ tenantId: query.tenantId, ...haltForm })
    ElMessage.success(t('common.success'))
    haltVisible.value = false
    await refresh()
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
  await refresh()
}

function calendarStatus(status: number) {
  return [
    '',
    t('option.calendarDraft'),
    t('option.calendarApproved'),
    t('option.calendarRejected'),
    t('option.calendarSuperseded'),
  ][status]
}

function formatTime(value?: number) {
  return value ? new Date(value * 1000).toLocaleString() : '-'
}

function clock(value: number) {
  const hour = Math.floor(value / 3600)
  const minute = Math.floor((value % 3600) / 60)
  return `${String(hour).padStart(2, '0')}:${String(minute).padStart(2, '0')}`
}

function sessionSummary(sessions: OptionTradingCalendarSession[]) {
  return (sessions || [])
    .map((item) => `${item.weekday} ${clock(item.openSecond)}-${clock(item.closeSecond)}`)
    .join('; ')
}

function exceptionSummary(exceptions: OptionTradingCalendarException[]) {
  return (exceptions || [])
    .map(
      (item) =>
        `${item.exceptionType === 1 ? 'CLOSED' : 'OPEN'} ${formatTime(item.startTime)}-${formatTime(item.endTime)}`,
    )
    .join('; ')
}

onMounted(refresh)
</script>

<style scoped>
.table-card {
  margin-top: 16px;
}
.danger {
  color: var(--el-color-danger);
}
</style>
