<script setup lang="ts">
/**
 * 租户分类管理：固定当前 tenantCode，对分类展示状态、排序和备注做管理。
 */
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import {
  categoriesService,
  tenantCategoriesService,
  type ItickCategory,
  type ItickTenantCategory,
  type OptionGroup,
} from '@/services'
import { useTenantStore } from '@/stores/tenant'
import { formatDate } from '@/utils'
import { buildAssetUrl } from '@/utils/file-url'
import { findOptionGroup, getOptionLabel, getOptionValueLabel } from '@/utils/options'

const tenant = useTenantStore()
const { t } = useI18n()
const loading = ref(false)
const submitLoading = ref(false)
const list = ref<ItickTenantCategory[]>([])
const categoryOptions = ref<ItickCategory[]>([])
const optionGroups = ref<OptionGroup[]>([])
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)

const query = reactive({
  categoryType: undefined as number | undefined,
  status: undefined as number | undefined,
  visibleStatus: undefined as number | undefined,
})

const form = reactive({
  categoryId: undefined as number | undefined,
  enabled: 1,
  appVisible: 1,
  sort: 0,
  remark: '',
})

const categoryTypeOptions = computed(() => findOptionGroup(optionGroups.value, 'categoryType'))
const statusOptions = computed(() => findOptionGroup(optionGroups.value, 'status'))
const visibleOptions = computed(() => findOptionGroup(optionGroups.value, 'visible'))

async function loadOptions() {
  const [tenantOptionRes, categoryRes] = await Promise.all([
    tenantCategoriesService.getOptions(),
    categoriesService.getList({ limit: 200 }),
  ])
  optionGroups.value = tenantOptionRes.data || []
  categoryOptions.value = categoryRes.data || []
}

async function loadList() {
  await tenant.ensureLoaded()
  loading.value = true
  try {
    const res = await tenantCategoriesService.getList({
      tenantId: tenant.tenantId,
      categoryType: query.categoryType,
      status: query.status,
      visibleStatus: query.visibleStatus,
      limit: 200,
    })
    list.value = res.data || []
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = null
  Object.assign(form, {
    categoryId: undefined,
    enabled: 1,
    appVisible: 1,
    sort: 0,
    remark: '',
  })
  dialogVisible.value = true
}

function openEdit(row: ItickTenantCategory) {
  editingId.value = row.id
  Object.assign(form, {
    categoryId: row.categoryId,
    enabled: row.enabled,
    appVisible: row.appVisible,
    sort: row.sort,
    remark: row.remark || '',
  })
  dialogVisible.value = true
}

async function submitForm() {
  await tenant.ensureLoaded()
  submitLoading.value = true
  try {
    if (editingId.value) {
      await tenantCategoriesService.update(editingId.value, {
        id: editingId.value,
        tenantId: tenant.tenantId,
        enabled: form.enabled,
        appVisible: form.appVisible,
        sort: form.sort,
        remark: form.remark,
      })
      ElMessage.success('分类已更新')
    } else {
      if (!form.categoryId) {
        ElMessage.warning('请选择分类')
        return
      }
      await tenantCategoriesService.create({
        tenantId: tenant.tenantId,
        categoryId: form.categoryId,
        enabled: form.enabled,
        appVisible: form.appVisible,
        sort: form.sort,
        remark: form.remark,
      })
      ElMessage.success('分类已添加')
    }
    dialogVisible.value = false
    await loadList()
  } finally {
    submitLoading.value = false
  }
}

onMounted(async () => {
  await tenant.ensureLoaded()
  await Promise.all([loadOptions(), loadList()])
})
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h2>分类管理</h2>
        <p>当前租户：{{ tenant.tenantName || tenant.tenantCode }}</p>
      </div>
      <el-button type="primary" @click="openCreate">新增分类</el-button>
    </div>

    <el-card shadow="never" class="query-card">
      <el-form :model="query" inline>
        <el-form-item label="分类类型">
          <el-select v-model="query.categoryType" clearable style="width: 180px">
            <el-option
              v-for="item in categoryTypeOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="启用状态">
          <el-select v-model="query.status" clearable style="width: 160px">
            <el-option
              v-for="item in statusOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="前台显示">
          <el-select v-model="query.visibleStatus" clearable style="width: 160px">
            <el-option
              v-for="item in visibleOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadList">查询</el-button>
          <el-button @click="Object.assign(query, { categoryType: undefined, status: undefined, visibleStatus: undefined }); loadList()">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never">
      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="categoryName" label="分类名称" min-width="180" />
        <el-table-column label="分类类型" width="120">
          <template #default="{ row }">
            {{ getOptionValueLabel(optionGroups, 'categoryType', row.categoryType, t) }}
          </template>
        </el-table-column>
        <el-table-column label="图标" width="90">
          <template #default="{ row }">
            <el-image v-if="row.icon" :src="buildAssetUrl(row.icon)" style="width: 28px; height: 28px" />
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column label="启用" width="90">
          <template #default="{ row }">
            <el-tag :type="row.enabled === 1 ? 'success' : 'info'">{{ row.enabled === 1 ? '启用' : '停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="显示" width="90">
          <template #default="{ row }">
            <el-tag :type="row.appVisible === 1 ? 'success' : 'warning'">{{ row.appVisible === 1 ? '显示' : '隐藏' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="180" show-overflow-tooltip />
        <el-table-column label="更新时间" min-width="170">
          <template #default="{ row }">{{ formatDate(row.updateTimes) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑分类' : '新增分类'" width="560px">
      <el-form label-width="100px">
        <el-form-item label="租户">
          <el-input :model-value="tenant.tenantName || tenant.tenantCode" disabled />
        </el-form-item>
        <el-form-item v-if="!editingId" label="平台分类">
          <el-select v-model="form.categoryId" filterable clearable style="width: 100%">
            <el-option
              v-for="item in categoryOptions"
              :key="item.id"
              :label="`${item.id} - ${item.categoryName}`"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="启用状态">
          <el-radio-group v-model="form.enabled">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="2">停用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="前台显示">
          <el-radio-group v-model="form.appVisible">
            <el-radio :value="1">显示</el-radio>
            <el-radio :value="2">隐藏</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" :precision="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="3" maxlength="200" show-word-limit />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitForm">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page { display: grid; gap: 16px; }
.page-header { display: flex; align-items: center; justify-content: space-between; }
.page-header h2 { margin: 0 0 6px; }
.page-header p { margin: 0; color: #909399; }
</style>
