<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus,
  Edit,
  Delete,
  VideoPlay,
  CircleCheck,
  CircleCloseFilled,
} from '@element-plus/icons-vue'
import { cronJobService, type OptionGroup } from '@/services'
import type {
  SysCronJobItem,
  SysCronJobCreateReq,
  SysCronJobUpdateReq,
} from '@/services/system/CronJobService'
import { usePagination } from '@/composables/usePagination'
import { useLoading } from '@/composables/useLoading'
import { useForm } from '@/composables/useForm'
import { useConfirm } from '@/composables/useConfirm'
import { formatDate } from '@/utils'
import { findOptionGroup, getOptionLabel, getOptionValueLabel } from '@/utils/options'

const { t } = useI18n()
const { confirm } = useConfirm()
const optionGroups = ref<OptionGroup[]>([])
const jobStatusOptions = computed(() => findOptionGroup(optionGroups.value, 'jobStatus'))

// Pagination and main list
const {
  pagination,
  updatePagination,
  nextPage: paginationNextPage,
  prevPage: paginationPrevPage,
} = usePagination(10)
const list = ref<SysCronJobItem[]>([])
const { loading, withLoading } = useLoading()

// Query form
const { form: queryForm } = useForm({
  initialData: {
    jobName: '',
    jobGroup: '',
    status: undefined as number | undefined,
  },
})

// Dialog and form
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitLoading = ref(false)
const formRef = ref()

const { form: formData, reset: resetForm } = useForm({
  initialData: {
    id: 0,
    jobName: '',
    jobGroup: 'DEFAULT',
    invokeTarget: '',
    cronExpression: '',
    status: 0,
    remark: '',
  },
})

// Handlers for dropdown
const handlers = ref<Array<{ invokeTarget: string; jobName: string }>>([])

// Form rules
const formRules = computed(() => ({
  jobName: [
    { required: true, message: t('common.required'), trigger: 'blur' },
    { min: 1, max: 100, message: 'Length should be 1-100', trigger: 'blur' },
  ],
  jobGroup: [{ required: true, message: t('common.required'), trigger: 'blur' }],
  invokeTarget: [{ required: true, message: t('common.required'), trigger: 'blur' }],
  cronExpression: [{ required: true, message: t('common.required'), trigger: 'blur' }],
}))

// Fetch list
async function fetchList() {
  await withLoading(async () => {
    try {
      const res = await cronJobService.getList({
        jobName: queryForm.jobName || undefined,
        jobGroup: queryForm.jobGroup || undefined,
        status: queryForm.status,
        cursor: pagination.cursor,
        limit: pagination.limit,
      })
      if (res.code !== 0 && res.code !== 200) throw new Error(res.msg)
      list.value = res.data || []
      updatePagination(
        res.total || 0,
        res.hasNext || false,
        res.hasPrev || false,
        res.nextCursor || null,
        res.prevCursor || null,
      )
    } catch (e: any) {
      ElMessage.error(e?.message || t('common.loadFailed'))
    }
  })
}

// Fetch handlers
async function fetchHandlers() {
  try {
    const res = await cronJobService.handlers()
    if (res.code !== 0 && res.code !== 200) throw new Error(res.msg)
    handlers.value = res.data || []
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.loadFailed'))
  }
}

async function fetchOptions() {
  try {
    const res = await cronJobService.getOptions()
    if (res.code !== 0 && res.code !== 200) throw new Error(res.msg)
    optionGroups.value = res.data || []
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.loadFailed'))
  }
}

// Search and reset
function onSearch() {
  pagination.cursor = null
  pagination.hasPrev = false
  fetchList()
}

function onReset() {
  queryForm.jobName = ''
  queryForm.jobGroup = ''
  queryForm.status = undefined
  pagination.cursor = null
  pagination.hasPrev = false
  fetchList()
}

// Pagination
function nextPage() {
  paginationNextPage()
  fetchList()
}

function prevPage() {
  paginationPrevPage()
  fetchList()
}

// Dialog operations
function handleCreate() {
  isEdit.value = false
  resetForm()
  dialogVisible.value = true
}

function handleEdit(row: SysCronJobItem) {
  isEdit.value = true
  formData.id = row.id
  formData.jobName = row.jobName
  formData.jobGroup = row.jobGroup
  formData.invokeTarget = row.invokeTarget
  formData.cronExpression = row.cronExpression
  formData.status = row.status
  formData.remark = row.remark || ''
  dialogVisible.value = true
}

// Submit
async function handleSubmit() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
    submitLoading.value = true

    try {
      let res
      if (isEdit.value) {
        res = await cronJobService.update(formData.id, formData as SysCronJobUpdateReq)
      } else {
        res = await cronJobService.create(formData as SysCronJobCreateReq)
      }

      if (res.code !== 0 && res.code !== 200) throw new Error(res.msg)
      ElMessage.success(isEdit.value ? t('common.updateSuccess') : t('common.createSuccess'))
      dialogVisible.value = false
      pagination.cursor = null
      pagination.hasPrev = false
      fetchList()
    } finally {
      submitLoading.value = false
    }
  } catch (e: any) {
    ElMessage.error(e?.message || 'Error')
  }
}

// Delete
async function handleDelete(row: SysCronJobItem) {
  try {
    await confirm(`${t('common.confirmDelete')} - ${row.jobName}?`)
    const res = await cronJobService.delete(row.id)
    if (res.code !== 0 && res.code !== 200) throw new Error(res.msg)
    ElMessage.success(t('common.success'))
    fetchList()
  } catch (e: any) {
    if (e.message !== 'Cancel') {
      ElMessage.error(e?.message || t('common.failed'))
    }
  }
}

// Run task
async function handleRun(row: SysCronJobItem) {
  try {
    await confirm(`${t('system.runTask')} - ${row.jobName}?`)
    const res = await cronJobService.run(row.id)
    if (res.code !== 0 && res.code !== 200) throw new Error(res.msg)
    ElMessage.success(t('common.success'))
  } catch (e: any) {
    if (e.message !== 'Cancel') {
      ElMessage.error(e?.message || t('common.failed'))
    }
  }
}

// Start task
async function handleStart(row: SysCronJobItem) {
  try {
    const res = await cronJobService.start(row.id)
    if (res.code !== 0 && res.code !== 200) throw new Error(res.msg)
    ElMessage.success(t('common.success'))
    fetchList()
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.failed'))
  }
}

// Stop task
async function handleStop(row: SysCronJobItem) {
  try {
    const res = await cronJobService.stop(row.id)
    if (res.code !== 0 && res.code !== 200) throw new Error(res.msg)
    ElMessage.success(t('common.success'))
    fetchList()
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.failed'))
  }
}

// Load on mount
onMounted(() => {
  fetchOptions()
  fetchHandlers()
  fetchList()
})
</script>

<template>
  <el-card>
    <template #header>
      <div style="display: flex; justify-content: space-between; align-items: center">
        <span>{{ t('system.cronJobs') }}</span>
        <el-button v-perm="'sys:job:add'" type="primary" @click="handleCreate">
          <el-icon><Plus /></el-icon>
          {{ t('common.add') }}
        </el-button>
      </div>
    </template>

    <!-- Query Form -->
    <el-form :model="queryForm" inline style="margin-bottom: 16px">
      <el-form-item :label="t('system.jobName')">
        <el-input
          v-model="queryForm.jobName"
          :placeholder="t('common.pleaseEnter')"
          clearable
          style="width: 180px"
        />
      </el-form-item>

      <el-form-item :label="t('system.jobGroup')">
        <el-input
          v-model="queryForm.jobGroup"
          :placeholder="t('common.pleaseEnter')"
          clearable
          style="width: 180px"
        />
      </el-form-item>

      <el-form-item :label="t('common.status')">
        <el-select
          v-model="queryForm.status"
          :placeholder="t('common.pleaseSelect')"
          clearable
          style="width: 140px"
        >
          <el-option
            v-for="item in jobStatusOptions"
            :key="item.value"
            :label="getOptionLabel(t, item.code, item.value)"
            :value="item.value"
          />
        </el-select>
      </el-form-item>

      <el-form-item>
        <el-button type="primary" @click="onSearch">
          {{ t('common.search') }}
        </el-button>
        <el-button @click="onReset">
          {{ t('common.reset') }}
        </el-button>
      </el-form-item>
    </el-form>

    <!-- Table -->
    <el-table
      v-loading="loading"
      :data="list"
      row-key="id"
      style="margin-bottom: 16px"
    >
      <el-table-column prop="id" :label="t('common.id')" width="70" />
      <el-table-column prop="jobName" :label="t('system.jobName')" min-width="120" />
      <el-table-column prop="jobGroup" :label="t('system.jobGroup')" width="100" />
      <el-table-column prop="invokeTarget" :label="t('system.invokeTarget')" min-width="180" />
      <el-table-column prop="cronExpression" :label="t('system.cronExpression')" width="140" />
      <el-table-column prop="status" :label="t('common.status')" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'">
            {{ getOptionValueLabel(optionGroups, 'jobStatus', row.status, t) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="createBy" :label="t('system.createBy')" width="100" />
      <el-table-column prop="createTimes" :label="t('common.createTimes')" width="170">
        <template #default="{ row }">
          <span>{{ formatDate(row.createTimes) }}</span>
        </template>
      </el-table-column>
      <el-table-column :label="t('common.actions')" width="280" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" size="small" @click="handleRun(row)">
            <el-icon><VideoPlay /></el-icon>
            {{ t('system.run') }}
          </el-button>
          <el-button
            v-if="row.status === 0"
            type="success"
            size="small"
            @click="handleStart(row)"
          >
            <el-icon><CircleCheck /></el-icon>
            {{ t('system.start') }}
          </el-button>
          <el-button
            v-if="row.status === 1"
            type="warning"
            size="small"
            @click="handleStop(row)"
          >
            <el-icon><CircleCloseFilled /></el-icon>
            {{ t('system.stop') }}
          </el-button>
          <el-button
            v-perm="'sys:job:update'"
            type="primary"
            size="small"
            @click="handleEdit(row)"
          >
            <el-icon><Edit /></el-icon>
            {{ t('common.edit') }}
          </el-button>
          <el-button
            v-perm="'sys:job:delete'"
            type="danger"
            size="small"
            @click="handleDelete(row)"
          >
            <el-icon><Delete /></el-icon>
            {{ t('common.delete') }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- Pagination -->
    <div style="display: flex; justify-content: flex-end; gap: 10px; align-items: center">
      <span>{{ t('common.totalItems', { count: pagination.total }) }}</span>
      <el-button :disabled="!pagination.hasPrev" @click="prevPage">
        {{ t('common.prevPage') }}
      </el-button>
      <el-button :disabled="!pagination.hasNext" @click="nextPage">
        {{ t('common.nextPage') }}
      </el-button>
      <el-select
        v-model="pagination.limit"
        style="width: 100px"
        @change="
          () => {
            pagination.cursor = null
            pagination.hasPrev = false
            fetchList()
          }
        "
      >
        <el-option label="10" :value="10" />
        <el-option label="20" :value="20" />
        <el-option label="50" :value="50" />
      </el-select>
    </div>

    <!-- Create/Edit Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? t('common.edit') : t('common.add')"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="140px"
      >
        <el-form-item :label="t('system.jobName')" prop="jobName">
          <el-input
            v-model="formData.jobName"
            :placeholder="t('common.pleaseEnter')"
            maxlength="100"
          />
        </el-form-item>

        <el-form-item :label="t('system.jobGroup')" prop="jobGroup">
          <el-input
            v-model="formData.jobGroup"
            :placeholder="t('common.pleaseEnter')"
            maxlength="50"
          />
        </el-form-item>

        <el-form-item :label="t('system.invokeTarget')" prop="invokeTarget">
          <el-select
            v-model="formData.invokeTarget"
            :placeholder="t('common.pleaseSelect')"
            filterable
            clearable
          >
            <el-option
              v-for="handler in handlers"
              :key="handler.invokeTarget"
              :label="`${handler.jobName} (${handler.invokeTarget})`"
              :value="handler.invokeTarget"
            />
          </el-select>
        </el-form-item>

        <el-form-item :label="t('system.cronExpression')" prop="cronExpression">
          <el-input
            v-model="formData.cronExpression"
            :placeholder="t('system.cronExpressionPlaceholder')"
            maxlength="100"
          />
        </el-form-item>

        <el-form-item :label="t('common.status')">
          <el-radio-group v-model="formData.status">
            <el-radio
              v-for="item in jobStatusOptions"
              :key="item.value"
              :label="item.value"
            >
              {{ getOptionLabel(t, item.code, item.value) }}
            </el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item :label="t('common.remark')">
          <el-input
            v-model="formData.remark"
            :placeholder="t('common.pleaseEnter')"
            type="textarea"
            :rows="3"
            maxlength="500"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<style scoped>
:deep(.el-card__header) {
  padding: 18px 20px;
}
</style>
