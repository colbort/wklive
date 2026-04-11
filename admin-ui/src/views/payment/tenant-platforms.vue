<template>
  <div class="payment-page module-page">
    <div class="page-header">
      <h2>租户支付平台</h2>
      <div class="header-actions">
        <el-button type="primary" @click="openDialog()">
          新增开通
        </el-button>
        <el-button @click="loadList">
          刷新
        </el-button>
      </div>
    </div>
    <el-card shadow="never" class="query-card">
      <el-form :model="query" inline label-width="90px">
        <el-form-item label="租户ID">
          <el-input-number v-model="query.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="平台ID">
          <el-input-number v-model="query.platformId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="query.status" clearable style="width: 160px">
            <el-option label="全部" :value="0" /><el-option label="启用" :value="1" /><el-option label="禁用" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadList">
            查询
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="tenantId" label="租户ID" width="100" />
        <el-table-column prop="platformId" label="平台ID" width="100" />
        <el-table-column prop="status" label="状态" width="90" />
        <el-table-column prop="openStatus" label="开通状态" width="100" />
        <el-table-column prop="remark" label="备注" min-width="180" />
        <el-table-column label="操作" width="160">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">
              详情
            </el-button>
            <el-button link type="primary" @click="openDialog(row)">
              编辑
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑租户平台' : '开通租户平台'" width="620px">
      <el-form label-width="100px">
        <el-form-item label="租户ID">
          <el-input-number
            v-model="form.tenantId"
            :min="1"
            :precision="0"
            :disabled="!!form.id"
          />
        </el-form-item>
        <el-form-item v-if="!form.id" label="平台ID">
          <el-input-number v-model="form.platformId" :min="1" :precision="0" />
        </el-form-item>
        <el-form-item label="状态">
          <el-input-number v-model="form.status" :min="1" :precision="0" />
        </el-form-item>
        <el-form-item label="开通状态">
          <el-input-number v-model="form.openStatus" :min="1" :precision="0" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible=false">
          取消
        </el-button><el-button type="primary" @click="submit">
          确定
        </el-button>
      </template>
    </el-dialog>
    <el-dialog v-model="detailVisible" title="详情" width="680px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>
<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { tenantService, type TenantPayPlatform } from '@/services'
const loading = ref(false)
const list = ref<TenantPayPlatform[]>([])
const dialogVisible = ref(false)
const detailVisible = ref(false)
const detailData = ref<Record<string, any>>({})
const query = reactive({ tenantId: 0, platformId: 0, status: 0 })
const form = reactive({ id: 0, tenantId: 0, platformId: 0, status: 1, openStatus: 1, remark: '' })
const loadList = async () => { loading.value = true; try { const res = await tenantService.getTenantPlatformList({ ...query, tenantId: query.tenantId || undefined, platformId: query.platformId || undefined, limit: 100 }); list.value = res.data || [] } finally { loading.value = false } }
const openDialog = (row?: TenantPayPlatform) => { Object.assign(form, row || { id: 0, tenantId: 0, platformId: 0, status: 1, openStatus: 1, remark: '' }); dialogVisible.value = true }
const submit = async () => { if (form.id) { await tenantService.updateTenantPlatform({ ...form }) } else { await tenantService.openTenantPlatform({ tenantId: form.tenantId, platformId: form.platformId, status: form.status, openStatus: form.openStatus, remark: form.remark }) } ElMessage.success('操作成功'); dialogVisible.value = false; loadList() }
const showDetail = async (row: TenantPayPlatform) => { const res = await tenantService.getTenantPlatformDetail(row.id, row.tenantId); detailData.value = res.data || row; detailVisible.value = true }
onMounted(loadList)
</script>
<style scoped></style>
