<template>
  <div class="sys-config">
    <div class="page-header">
      <h2>{{ t('system.config') }}</h2>
      <el-button type="primary" @click="handleCreate">
        <el-icon><Plus /></el-icon>
        {{ t('common.add') }}
      </el-button>
    </div>

    <!-- 查询表单 -->
    <el-card class="query-card" shadow="never">
      <el-form :model="queryForm" inline>
        <el-form-item :label="t('system.configKey')">
          <el-input
            v-model="queryForm.keyword"
            :placeholder="t('common.pleaseEnter')"
            clearable
            @keyup.enter="fetchList"
          />
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
        :data="list"
        v-loading="loading"
        :empty-text="t('common.noData')"
        stripe
      >
        <el-table-column
          prop="id"
          :label="t('common.id')"
          width="80"
          align="center"
        />
        <el-table-column
          prop="configKey"
          :label="t('system.configKey')"
          min-width="150"
        />
        <el-table-column
          prop="configValue"
          :label="t('system.configValue')"
          min-width="200"
        >
          <template #default="{ row }">
            <el-tooltip :content="row.configValue" placement="top">
              <span class="config-value">{{ row.configValue }}</span>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column
          prop="remark"
          :label="t('common.remark')"
          min-width="150"
        />
        <el-table-column
          prop="createdAt"
          :label="t('common.createdAt')"
          width="160"
          align="center"
        >
          <template #default="{ row }">
            {{ formatDate(row.createdAt) }}
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
              type="primary"
              size="small"
              @click="handleEdit(row)"
            >
              {{ t('common.edit') }}
            </el-button>
            <el-button
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
      <div class="pagination">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
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
        <el-form-item
          :label="t('system.configKey')"
          prop="configKey"
        >
          <el-input
            v-model="formData.configKey"
            :placeholder="t('common.pleaseEnter')"
            :disabled="isEdit"
          />
        </el-form-item>
        <el-form-item
          :label="t('system.configValue')"
          prop="configValue"
        >
          <el-input
            v-model="formData.configValue"
            type="textarea"
            :rows="4"
            :placeholder="t('common.pleaseEnter')"
          />
        </el-form-item>
        <el-form-item
          :label="t('common.remark')"
          prop="remark"
        >
          <el-input
            v-model="formData.remark"
            :placeholder="t('common.pleaseEnter')"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button
          type="primary"
          :loading="submitLoading"
          @click="handleSubmit"
        >
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
import type { SysConfigItem, SysConfigCreateReq, SysConfigUpdateReq } from '@/services'
import { usePagination } from '@/composables/usePagination'
import { useLoading } from '@/composables/useLoading'
import { useForm } from '@/composables/useForm'
import { useConfirm } from '@/composables/useConfirm'
import { formatDate } from '@/utils'

const { t } = useI18n()

// Pagination and main list
const { pagination, updateTotal } = usePagination(10)
const list = ref<SysConfigItem[]>([])
const { loading, withLoading } = useLoading()

// Query form
const { form: queryForm } = useForm({
  initialData: {
    keyword: ''
  }
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
    remark: ''
  }
})

// Form validation rules
const formRules = {
  configKey: [
    { required: true, message: t('validation.required'), trigger: 'blur' }
  ],
  configValue: [
    { required: true, message: t('validation.required'), trigger: 'blur' }
  ]
}

// Fetch list
async function fetchList() {
  await withLoading(async () => {
    try {
      const res = await configService.getList({
        keyword: queryForm.keyword || undefined,
        page: pagination.page,
        size: pagination.pageSize,
      })
      if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || 'list failed')
      list.value = res.data || []
      updateTotal(res.total || 0)
    } catch (e: any) {
      ElMessage.error(e?.message || t('common.loadFailed'))
    }
  })
}

// Handle pagination
function handleSizeChange(size: number) {
  pagination.pageSize = size
  pagination.page = 1
  fetchList()
}

function handleCurrentChange(page: number) {
  pagination.page = page
  fetchList()
}

// Handle reset
function handleReset() {
  queryForm.keyword = ''
  pagination.page = 1
  fetchList()
}

// Handle create
function handleCreate() {
  isEdit.value = false
  resetForm()
  dialogVisible.value = true
}

// Handle edit
function handleEdit(row: SysConfigItem) {
  isEdit.value = true
  resetForm()
  Object.assign(formData, {
    id: row.id,
    configKey: row.configKey,
    configValue: row.configValue,
    remark: row.remark || ''
  })
  dialogVisible.value = true
}

// Handle delete
async function handleDelete(row: SysConfigItem) {
  try {
    await ElMessageBox.confirm(
      t('common.confirmDelete'),
      t('common.warning'),
      {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        type: 'warning',
      }
    )

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

    if (isEdit.value) {
      const { id, ...updateData } = formData
      const res = await configService.update(id, updateData)
      if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || 'update failed')
      ElMessage.success(t('common.updateSuccess'))
    } else {
      const data: SysConfigCreateReq = {
        configKey: formData.configKey,
        configValue: formData.configValue,
        remark: formData.remark || undefined
      }
      const res = await configService.create(data)
      if (res.code !== 0 && res.code !== 200) throw new Error(res.msg || 'create failed')
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