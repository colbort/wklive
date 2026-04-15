<template>
  <div class="sys-config">
    <div class="page-header">
      <h2>{{ t('system.config') }}</h2>
      <el-button v-perm="'sys:config:add'" type="primary" @click="handleCreate">
        <el-icon><Plus /></el-icon>
        {{ t('common.add') }}
      </el-button>
    </div>

    <!-- 查询表单 -->
    <el-card class="query-card" shadow="never">
      <el-form :model="queryForm" inline>
        <el-form-item :label="t('system.configKey')">
          <el-select
            v-model="queryForm.keyword"
            :placeholder="t('system.pleaseSelect')"
            filterable
            clearable
            @change="fetchList"
          >
            <el-option
              v-for="key in keys"
              :key="key.value"
              :label="getOptionLabel(key.code, key.value)"
              :value="key.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchList">
            <el-icon><Search /></el-icon>
            {{ t('common.search') }}
          </el-button>
          <el-button @click="handleReset">
            <el-icon><Refresh /></el-icon>
            {{ t('common.reset') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 数据表格 -->
    <el-card class="table-card" shadow="never">
      <el-table
        v-loading="loading"
        :data="list"
        :empty-text="t('common.noData')"
        stripe
      >
        <el-table-column
          prop="id"
          :label="t('common.id')"
          width="80"
          align="center"
        />
        <el-table-column prop="configKey" :label="t('system.configKey')" min-width="150" />
        <el-table-column prop="configValue" :label="t('system.configValue')" min-width="200">
          <template #default="{ row }">
            <el-tooltip :content="row.configValue" placement="top" popper-class="config-tip">
              <span class="config-value">{{ row.configValue }}</span>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column prop="remark" :label="t('common.remark')" min-width="150" />
        <el-table-column
          prop="createTimes"
          :label="t('common.createTimes')"
          width="160"
          align="center"
        >
          <template #default="{ row }">
            {{ formatDate(row.createTimes) }}
          </template>
        </el-table-column>
        <el-table-column
          :label="t('common.actions')"
          width="150"
          align="center"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              v-perm="'sys:config:update'"
              type="primary"
              size="small"
              @click="handleEdit(row)"
            >
              {{ t('common.edit') }}
            </el-button>
            <el-button
              v-perm="'sys:config:delete'"
              type="danger"
              size="small"
              @click="handleDelete(row)"
            >
              {{ t('common.delete') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div
        style="
          display: flex;
          justify-content: flex-end;
          gap: 10px;
          align-items: center;
          margin-top: 12px;
        "
      >
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
    </el-card>

    <!-- 新增/编辑对话框 -->
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
        label-width="100px"
      >
        <el-form-item :label="t('system.configKey')" prop="configKey">
          <el-select
            v-if="!isEdit"
            v-model="formData.configKey"
            :placeholder="t('system.pleaseSelect')"
            filterable
            clearable
            @change="handleConfigKeyChange"
          >
            <el-option
              v-for="key in keys"
              :key="key.value"
              :label="getOptionLabel(key.code, key.value)"
              :value="key.value"
            />
          </el-select>
          <el-input
            v-else
            v-model="formData.configKey"
            :placeholder="t('common.pleaseEnter')"
            disabled
          />
        </el-form-item>
        <template v-if="formData.configKey === 'SYSTEM_CORE'">
          <SystemCoreConfig v-model="systemCoreForm" />
        </template>

        <template v-else-if="formData.configKey === 'OBJECT_STORAGE'">
          <ObjectStorageConfigComponent v-model="objectStorageForm" />
        </template>

        <template v-else>
          <el-form-item :label="t('system.configValue')" prop="configValue">
            <el-input
              v-model="formData.configValue"
              type="textarea"
              :rows="4"
              :placeholder="t('common.pleaseEnter')"
            />
          </el-form-item>
        </template>
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
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { Plus, Search, Refresh } from '@element-plus/icons-vue'
import { configService } from '@/services'
import type {
  SysConfigItem,
  SysConfigCreateReq,
  SysConfigUpdateReq,
  SysOptionItem,
} from '@/services'
import type { SystemCore, ObjectStorageConfig } from '@/services/system/ConfigService'
import { usePagination } from '@/composables/usePagination'
import { useLoading } from '@/composables/useLoading'
import { useForm } from '@/composables/useForm'
import { useConfirm } from '@/composables/useConfirm'
import { formatDate } from '@/utils'
import SystemCoreConfig from './components/SystemCoreConfig.vue'
import ObjectStorageConfigComponent from './components/ObjectStorageConfig.vue'

const { t } = useI18n()

function getOptionLabel(code?: string, value?: number) {
  if (!code) return value || ''
  const key = `options.${code}`
  const translated = t(key)
  console.log(translated)
  console.log(key)
  return translated === key ? (value || code) : translated
}

// Pagination and main list
const {
  pagination,
  updatePagination,
  nextPage: paginationNextPage,
  prevPage: paginationPrevPage,
} = usePagination(10)
const list = ref<SysConfigItem[]>([])
const { loading, withLoading } = useLoading()

// Query form
const { form: queryForm } = useForm({
  initialData: {
    keyword: '',
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
    configKey: '',
    configValue: '',
    remark: '',
  },
})

// Keys for configKey selection
const keys = ref<SysOptionItem[]>([])

// Config type special forms
const systemCoreForm = ref<SystemCore>({
  site_name: '',
  site_logo: '',
})

const activeTab = ref('aliyun')

const objectStorageForm = ref<ObjectStorageConfig>({
  aliyun_oss: {
    endpoint: '',
    access_key_id: '',
    access_key_secret: '',
    bucket_name: '',
    bucket_url: '',
  },
  tencent_cos: {
    market: '',
    secret_id: '',
    secret_key: '',
    bucket_name: '',
    bucket_url: '',
  },
  minio: {
    endpoint: '',
    access_key_id: '',
    access_key_secret: '',
    bucket_name: '',
    bucket_url: '',
  },
  oss_type: 1,
  oss_domain: '',
})

// Form validation rules
const formRules = {
  configKey: [{ required: true, message: t('validation.required'), trigger: 'blur' }],
  configValue: [{ required: true, message: t('validation.required'), trigger: 'blur' }],
}

// Load available keys
async function loadKeys() {
  try {
    const res = await configService.getKeys()
    if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || 'Failed to load keys')
    keys.value = res.data || []
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load keys')
  }
}

// Fetch list
async function fetchList() {
  await withLoading(async () => {
    try {
      const res = await configService.getList({
        keyword: queryForm.keyword || undefined,
        cursor: pagination.cursor,
        limit: pagination.limit,
      })
      if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || 'list failed')
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

// Handle pagination
function handleSizeChange(size: number) {
  pagination.limit = size
  pagination.cursor = null
  pagination.hasPrev = false
  fetchList()
}

// Handle reset
function handleReset() {
  queryForm.keyword = ''
  pagination.cursor = null
  pagination.hasPrev = false
  fetchList()
}

function resetTypeForms() {
  systemCoreForm.value = {
    site_name: '',
    site_logo: '',
  }
  objectStorageForm.value = {
    aliyun_oss: {
      endpoint: '',
      access_key_id: '',
      access_key_secret: '',
      bucket_name: '',
      bucket_url: '',
    },
    tencent_cos: {
      market: '',
      secret_id: '',
      secret_key: '',
      bucket_name: '',
      bucket_url: '',
    },
    minio: {
      endpoint: '',
      access_key_id: '',
      access_key_secret: '',
      bucket_name: '',
      bucket_url: '',
    },
    oss_type: 1,
    oss_domain: '',
  }
}

function handleConfigKeyChange(value: string) {
  if (value === 'SYSTEM_CORE') {
    systemCoreForm.value = {
      site_name: '',
      site_logo: '',
    }
    formData.configValue = ''
  } else if (value === 'OBJECT_STORAGE') {
    objectStorageForm.value = {
      aliyun_oss: {
        endpoint: '',
        access_key_id: '',
        access_key_secret: '',
        bucket_name: '',
        bucket_url: '',
      },
      tencent_cos: {
        market: '',
        secret_id: '',
        secret_key: '',
        bucket_name: '',
        bucket_url: '',
      },
      minio: {
        endpoint: '',
        access_key_id: '',
        access_key_secret: '',
        bucket_name: '',
        bucket_url: '',
      },
      oss_type: 1,
      oss_domain: '',
    }
    formData.configValue = ''
  }
}

function handleTabClick(_tab: any) {
  // 仅切换视图选项卡，不修改 oss_type
}

function handleOssTypeChange(_value: number) {
  // 仅修改 oss_type，不切换选项卡
}

function nextPage() {
  paginationNextPage()
  fetchList()
}

function prevPage() {
  paginationPrevPage()
  fetchList()
}

// Handle create
function handleCreate() {
  isEdit.value = false
  resetForm()
  resetTypeForms()
  loadKeys()
  dialogVisible.value = true
}

// Handle edit
function handleEdit(row: SysConfigItem) {
  isEdit.value = true
  resetForm()
  resetTypeForms()
  Object.assign(formData, {
    id: row.id,
    configKey: row.configKey,
    remark: row.remark || '',
  })

  if (row.configKey === 'SYSTEM_CORE') {
    try {
      const parsed = JSON.parse(row.configValue || '{}')
      systemCoreForm.value = {
        site_name: parsed.site_name || '',
        site_logo: parsed.site_logo || '',
      }
    } catch {
      systemCoreForm.value = { site_name: '', site_logo: '' }
    }
  } else if (row.configKey === 'OBJECT_STORAGE') {
    try {
      const parsed = JSON.parse(row.configValue || '{}')
      objectStorageForm.value = {
        aliyun_oss: parsed.aliyun_oss || objectStorageForm.value.aliyun_oss,
        tencent_cos: parsed.tencent_cos || objectStorageForm.value.tencent_cos,
        minio: parsed.minio || objectStorageForm.value.minio,
        oss_type: parsed.oss_type || 1,
        oss_domain: parsed.oss_domain || '',
      }
      if (parsed.oss_type === 1) activeTab.value = 'aliyun'
      else if (parsed.oss_type === 2) activeTab.value = 'tencent'
      else activeTab.value = 'minio'
    } catch {
      resetTypeForms()
    }
  } else {
    formData.configValue = row.configValue
  }

  dialogVisible.value = true
}

// Handle delete
async function handleDelete(row: SysConfigItem) {
  try {
    await ElMessageBox.confirm(t('common.confirmDelete'), t('common.warning'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning',
    })

    const res = await configService.delete(row.id)
    if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || 'delete failed')

    ElMessage.success(t('common.deleteSuccess'))
    fetchList()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || t('common.deleteFailed'))
    }
  }
}

// Handle submit
async function handleSubmit() {
  if (!formRef.value) return

  try {
    await formRef.value.validate()

    submitLoading.value = true

    if (formData.configKey === 'SYSTEM_CORE') {
      if (!systemCoreForm.value.site_name) {
        throw new Error(t('validation.required'))
      }
      formData.configValue = JSON.stringify(systemCoreForm.value)
    } else if (formData.configKey === 'OBJECT_STORAGE') {
      if (!objectStorageForm.value.oss_domain) {
        throw new Error(t('validation.required'))
      }
      formData.configValue = JSON.stringify(objectStorageForm.value)
    }

    if (isEdit.value) {
      const { id, ...updateData } = formData
      const res = await configService.update(id, updateData)
      if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || t('common.updateFailed'))
      ElMessage.success(t('common.updateSuccess'))
    } else {
      const data: SysConfigCreateReq = {
        configKey: formData.configKey,
        configValue: formData.configValue,
        remark: formData.remark || undefined,
      }
      const res = await configService.create(data)
      if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || t('common.createFailed'))
      ElMessage.success(t('common.createSuccess'))
    }

    dialogVisible.value = false
    fetchList()
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.operationFailed'))
  } finally {
    submitLoading.value = false
  }
}

// Initialize
onMounted(() => {
  loadKeys()
  fetchList()
})
</script>

<style scoped>
.sys-config {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  color: #333;
}

.query-card {
  margin-bottom: 20px;
}

.table-card {
  margin-bottom: 20px;
}

.pagination {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}

.config-value {
  display: inline-block;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>

<style>
.config-tip {
  max-width: 520px !important;
  white-space: normal !important;
  word-break: break-all !important;
  overflow-wrap: anywhere !important;
  line-height: 1.5;
}
</style>
