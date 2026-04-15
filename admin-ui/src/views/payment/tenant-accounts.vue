<template>
  <div class="payment-page">
    <div class="page-header">
      <h2>租户支付账号</h2>
      <div>
        <el-button type="primary" @click="openDialog()">
          新增账号
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
        <el-form-item label="关键字">
          <el-input v-model="query.keyword" clearable />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadList">
            查询
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never">
      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="tenantId" label="租户ID" width="100" />
        <el-table-column prop="platformId" label="平台ID" width="100" />
        <el-table-column prop="accountCode" label="账号编码" min-width="140" />
        <el-table-column prop="accountName" label="账号名称" min-width="160" />
        <el-table-column prop="merchantId" label="商户号" min-width="140" />
        <el-table-column prop="status" label="状态" width="90" />
        <el-table-column prop="isDefault" label="默认" width="80" />
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

    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑账号' : '新增账号'" width="760px">
      <el-form label-width="120px">
        <el-form-item label="租户ID">
          <el-input-number
            v-model="form.tenantId"
            :min="1"
            :precision="0"
            :disabled="!!form.id"
          />
        </el-form-item>
        <el-form-item v-if="!form.id" label="开通平台ID">
          <el-input-number v-model="form.tenantPayPlatformId" :min="1" :precision="0" />
        </el-form-item>
        <el-form-item v-if="!form.id" label="平台ID">
          <el-input-number v-model="form.platformId" :min="1" :precision="0" />
        </el-form-item>
        <el-form-item v-if="!form.id" label="账号编码">
          <el-input v-model="form.accountCode" />
        </el-form-item>
        <el-form-item label="账号名称">
          <el-input v-model="form.accountName" />
        </el-form-item>
        <el-form-item label="APP ID">
          <el-input v-model="form.appId" />
        </el-form-item>
        <el-form-item label="商户号">
          <el-input v-model="form.merchantId" />
        </el-form-item>
        <el-form-item label="商户名">
          <el-input v-model="form.merchantName" />
        </el-form-item>
        <el-form-item label="API Key密文">
          <el-input v-model="form.apiKeyCipher" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="API Secret密文">
          <el-input v-model="form.apiSecretCipher" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="私钥密文">
          <el-input v-model="form.privateKeyCipher" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="公钥">
          <el-input v-model="form.publicKey" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="证书密文">
          <el-input v-model="form.certCipher" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="扩展配置">
          <el-input v-model="form.extConfig" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="状态">
          <el-input-number
            v-model="form.status"
            :min="1"
            :max="2"
            :precision="0"
          />
        </el-form-item>
        <el-form-item label="默认">
          <el-input-number
            v-model="form.isDefault"
            :min="0"
            :max="1"
            :precision="0"
          />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">
          取消
        </el-button>
        <el-button type="primary" @click="submit">
          确定
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" title="详情" width="760px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { tenantService, type TenantPayAccount } from '@/services'

const loading = ref(false)
const list = ref<TenantPayAccount[]>([])
const dialogVisible = ref(false)
const detailVisible = ref(false)
const detailData = ref<Record<string, any>>({})

const query = reactive({
  tenantId: 0,
  platformId: 0,
  keyword: '',
})

const createEmptyForm = () => ({
  id: 0,
  tenantId: 0,
  tenantPayPlatformId: 0,
  platformId: 0,
  accountCode: '',
  accountName: '',
  appId: '',
  merchantId: '',
  merchantName: '',
  apiKeyCipher: '',
  apiSecretCipher: '',
  privateKeyCipher: '',
  publicKey: '',
  certCipher: '',
  extConfig: '',
  status: 1,
  isDefault: 1,
  remark: '',
})

const form = reactive(createEmptyForm())

const loadList = async () => {
  loading.value = true
  try {
    const res = await tenantService.getTenantAccountList({
      ...query,
      tenantId: query.tenantId || undefined,
      platformId: query.platformId || undefined,
      keyword: query.keyword || undefined,
      limit: 100,
    })
    list.value = res.data || []
  } finally {
    loading.value = false
  }
}

const openDialog = (row?: TenantPayAccount) => {
  Object.assign(form, createEmptyForm(), row || {})
  dialogVisible.value = true
}

const submit = async () => {
  if (form.id) {
    await tenantService.updateTenantAccount({ ...form })
  } else {
    await tenantService.createTenantAccount({ ...form })
  }
  ElMessage.success('操作成功')
  dialogVisible.value = false
  loadList()
}

const showDetail = async (row: TenantPayAccount) => {
  const res = await tenantService.getTenantAccountDetail(row.id, row.tenantId)
  detailData.value = res.data || row
  detailVisible.value = true
}

onMounted(loadList)
</script>

<style scoped></style>
