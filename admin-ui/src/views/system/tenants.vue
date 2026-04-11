<template>
  <div class="sys-tenants">
    <div class="page-header">
      <h2>{{ t('system.tenants') }}</h2>
      <el-button v-perm="'sys:tenant:add'" type="primary" @click="handleCreate">
        <el-icon><Plus /></el-icon>
        {{ t('common.add') }}
      </el-button>
    </div>

    <!-- 查询表单 -->
    <el-card class="query-card" shadow="never">
      <el-form :model="queryForm" inline>
        <el-form-item :label="t('system.tenantCode')">
          <el-input
            v-model="queryForm.tenantCode"
            :placeholder="t('system.tenantCodePlaceholder')"
            clearable
            @keyup.enter="fetchList"
          />
        </el-form-item>
        <el-form-item :label="t('system.tenantName')">
          <el-input
            v-model="queryForm.tenantName"
            :placeholder="t('system.tenantNamePlaceholder')"
            clearable
            @keyup.enter="fetchList"
          />
        </el-form-item>
        <el-form-item :label="t('system.contactName')">
          <el-input
            v-model="queryForm.contactName"
            :placeholder="t('system.contactNamePlaceholder')"
            clearable
            @keyup.enter="fetchList"
          />
        </el-form-item>
        <el-form-item :label="t('system.status')">
          <el-select
            v-model="queryForm.status"
            :placeholder="t('system.pleaseSelectStatus')"
            clearable
            @change="fetchList"
          >
            <el-option :label="t('common.all')" :value="0" />
            <el-option :label="t('common.enabled')" :value="1" />
            <el-option :label="t('common.disabled')" :value="2" />
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
        <el-table-column prop="tenantCode" :label="t('system.tenantCode')" min-width="150" />
        <el-table-column prop="tenantName" :label="t('system.tenantName')" min-width="150" />
        <el-table-column prop="contactName" :label="t('system.contactName')" min-width="120" />
        <el-table-column prop="contactPhone" :label="t('system.contactPhone')" min-width="130" />
        <el-table-column :label="t('system.status')" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">
              {{ row.status === 1 ? t('common.enabled') : t('common.disabled') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="expireTime"
          :label="t('system.expireTime')"
          min-width="160"
          align="center"
        >
          <template #default="{ row }">
            {{ formatDate(row.expireTime) }}
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
          width="180"
          align="center"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              v-perm="'sys:tenant:update'"
              type="primary"
              size="small"
              @click="handleEdit(row)"
            >
              {{ t('common.edit') }}
            </el-button>
            <el-button
              v-perm="'sys:tenant:delete'"
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
        <el-select v-model="pagination.limit" style="width: 100px" @change="handleSizeChange">
          <el-option :value="10" :label="t('common.perPage10')" />
          <el-option :value="20" :label="t('common.perPage20')" />
          <el-option :value="50" :label="t('common.perPage50')" />
        </el-select>
      </div>
    </el-card>

    <!-- 新增/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? t('system.editTenant') : t('system.addTenant')"
      width="700px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="120px"
      >
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="t('system.tenantCode')" prop="tenantCode">
              <el-input
                v-model="formData.tenantCode"
                :placeholder="t('system.pleaseInputTenantCode')"
                maxlength="50"
                show-word-limit
                :disabled="isEdit"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('system.tenantName')" prop="tenantName">
              <el-input
                v-model="formData.tenantName"
                :placeholder="t('system.pleaseInputTenantName')"
                maxlength="100"
                show-word-limit
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="t('system.contactName')" prop="contactName">
              <el-input
                v-model="formData.contactName"
                :placeholder="t('system.pleaseInputContactName')"
                maxlength="50"
                show-word-limit
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('system.contactPhone')" prop="contactPhone">
              <el-input
                v-model="formData.contactPhone"
                :placeholder="t('system.pleaseInputContactPhone')"
                maxlength="20"
                show-word-limit
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="t('system.status')" prop="status">
              <el-radio-group v-model="formData.status">
                <el-radio :value="1">
                  {{ t('common.enabled') }}
                </el-radio>
                <el-radio :value="2">
                  {{ t('common.disabled') }}
                </el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('system.expireTime')" prop="expireTime">
              <el-date-picker
                v-model="formData.expireTime"
                type="datetime"
                :placeholder="t('common.pleaseSelect')"
                format="YYYY-MM-DD HH:mm:ss"
                value-format="x"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item :label="t('common.remark')">
          <el-input
            v-model="formData.remark"
            type="textarea"
            :rows="4"
            :placeholder="t('common.remark')"
            maxlength="200"
            show-word-limit
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
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { Plus, Search, Refresh } from '@element-plus/icons-vue'
import { tenantsService } from '@/services/system/TenantsService'
import type {
  SysTenantItem,
  SysTenantCreateReq,
  SysTenantUpdateReq,
} from '@/services/system/TenantsService'
import { usePagination } from '@/composables/usePagination'
import { useLoading } from '@/composables/useLoading'
import { useForm } from '@/composables/useForm'
import { formatDate } from '@/utils'

const { t } = useI18n()

// Pagination and main list
const {
  pagination,
  updatePagination,
  nextPage: paginationNextPage,
  prevPage: paginationPrevPage,
} = usePagination(10)
const list = ref<SysTenantItem[]>([])
const { loading, withLoading } = useLoading()

// Query form
const { form: queryForm } = useForm({
  initialData: {
    tenantCode: '',
    tenantName: '',
    contactName: '',
    status: 0,
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
    tenantCode: '',
    tenantName: '',
    status: 1,
    expireTime: Date.now() + 365 * 24 * 60 * 60 * 1000, // 默认一年后过期
    contactName: '',
    contactPhone: '',
    remark: '',
  },
})

// Form validation rules
const formRules = {
  tenantCode: [{ required: true, message: t('system.pleaseInputTenantCode'), trigger: 'blur' }],
  tenantName: [{ required: true, message: t('system.pleaseInputTenantName'), trigger: 'blur' }],
  status: [{ required: true, message: t('system.pleaseSelectStatus'), trigger: 'change' }],
  expireTime: [{ required: true, message: t('validation.required'), trigger: 'change' }],
  contactName: [{ required: true, message: t('system.pleaseInputContactName'), trigger: 'blur' }],
  contactPhone: [{ required: true, message: t('system.pleaseInputContactPhone'), trigger: 'blur' }],
}

// Fetch list
async function fetchList() {
  await withLoading(async () => {
    try {
      const params = {
        tenantCode: queryForm.tenantCode || undefined,
        tenantName: queryForm.tenantName || undefined,
        contactName: queryForm.contactName || undefined,
        status: queryForm.status === 0 ? undefined : queryForm.status,
        cursor: pagination.cursor,
        limit: pagination.limit,
      }
      const res = await tenantsService.getList(params)
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
  queryForm.tenantCode = ''
  queryForm.tenantName = ''
  queryForm.contactName = ''
  queryForm.status = 0
  pagination.cursor = null
  pagination.hasPrev = false
  fetchList()
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
  dialogVisible.value = true
}

// Handle edit
function handleEdit(row: SysTenantItem) {
  isEdit.value = true
  resetForm()
  Object.assign(formData, {
    id: row.id,
    tenantCode: row.tenantCode,
    tenantName: row.tenantName,
    status: row.status,
    expireTime: row.expireTime,
    contactName: row.contactName,
    contactPhone: row.contactPhone,
    remark: row.remark || '',
  })
  dialogVisible.value = true
}

// Handle delete
async function handleDelete(row: SysTenantItem) {
  try {
    await ElMessageBox.confirm(t('common.confirmDelete'), t('common.warning'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning',
    })

    const res = await tenantsService.delete(row.id)
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

    if (isEdit.value) {
      const { id, ...updateData } = formData
      const res = await tenantsService.update(id, updateData)
      if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || t('common.updateFailed'))
      ElMessage.success(t('common.updateSuccess'))
    } else {
      const data: SysTenantCreateReq = {
        tenantCode: formData.tenantCode,
        tenantName: formData.tenantName,
        status: formData.status,
        expireTime: formData.expireTime,
        contactName: formData.contactName,
        contactPhone: formData.contactPhone,
        remark: formData.remark || '',
      }
      const res = await tenantsService.create(data)
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
  fetchList()
})
</script>

<style scoped>
.sys-tenants {
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
</style>
