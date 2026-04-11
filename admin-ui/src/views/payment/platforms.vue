<template>
  <div class="payment-page module-page">
    <div class="page-header">
      <h2>支付基础配置</h2>
      <div class="header-actions">
        <el-button @click="refreshCurrentPage">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <el-tabs v-model="activeTab">
      <el-tab-pane label="支付平台" name="platforms">
        <el-card shadow="never" class="query-card">
          <el-form :model="platformQuery" inline label-width="90px">
            <el-form-item label="平台编码">
              <el-input v-model="platformQuery.platformCode" clearable />
            </el-form-item>
            <el-form-item label="关键字">
              <el-input v-model="platformQuery.keyword" clearable />
            </el-form-item>
            <el-form-item label="状态">
              <el-select v-model="platformQuery.status" clearable style="width: 160px">
                <el-option label="全部" :value="0" />
                <el-option label="启用" :value="1" />
                <el-option label="禁用" :value="2" />
              </el-select>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="loadPlatforms">
                查询
              </el-button>
              <el-button @click="resetPlatformQuery">
                重置
              </el-button>
              <el-button type="primary" @click="openPlatformDialog()">
                新增平台
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <el-card shadow="never" class="table-card">
          <el-table v-loading="platformLoading" :data="platforms" stripe>
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column prop="platformCode" label="平台编码" min-width="140" />
            <el-table-column prop="platformName" label="平台名称" min-width="160" />
            <el-table-column prop="platformType" label="平台类型" width="100" />
            <el-table-column
              prop="notifyUrl"
              label="回调地址"
              min-width="220"
              show-overflow-tooltip
            />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 1 ? 'success' : 'info'">
                  {{ row.status === 1 ? '启用' : '禁用' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column
              prop="remark"
              label="备注"
              min-width="180"
              show-overflow-tooltip
            />
            <el-table-column label="操作" width="160" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="showPlatformDetail(row)">
                  详情
                </el-button>
                <el-button link type="primary" @click="openPlatformDialog(row)">
                  编辑
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="支付产品" name="products">
        <el-card shadow="never" class="query-card">
          <el-form :model="productQuery" inline label-width="90px">
            <el-form-item label="平台ID">
              <el-input-number v-model="productQuery.platformId" :min="0" :precision="0" />
            </el-form-item>
            <el-form-item label="产品编码">
              <el-input v-model="productQuery.productCode" clearable />
            </el-form-item>
            <el-form-item label="关键字">
              <el-input v-model="productQuery.keyword" clearable />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="loadProducts">
                查询
              </el-button>
              <el-button @click="resetProductQuery">
                重置
              </el-button>
              <el-button type="primary" @click="openProductDialog()">
                新增产品
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <el-card shadow="never" class="table-card">
          <el-table v-loading="productLoading" :data="products" stripe>
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column prop="platformId" label="平台ID" width="100" />
            <el-table-column prop="productCode" label="产品编码" min-width="140" />
            <el-table-column prop="productName" label="产品名称" min-width="160" />
            <el-table-column prop="sceneType" label="场景类型" width="100" />
            <el-table-column prop="currency" label="币种" width="100" />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 1 ? 'success' : 'info'">
                  {{ row.status === 1 ? '启用' : '禁用' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column
              prop="remark"
              label="备注"
              min-width="180"
              show-overflow-tooltip
            />
            <el-table-column label="操作" width="160" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="showProductDetail(row)">
                  详情
                </el-button>
                <el-button link type="primary" @click="openProductDialog(row)">
                  编辑
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="platformDialogVisible" :title="platformForm.id ? '编辑平台' : '新增平台'" width="640px">
      <el-form label-width="100px">
        <el-form-item v-if="!platformForm.id" label="平台编码">
          <el-input v-model="platformForm.platformCode" />
        </el-form-item>
        <el-form-item label="平台名称">
          <el-input v-model="platformForm.platformName" />
        </el-form-item>
        <el-form-item label="平台类型">
          <el-input-number v-model="platformForm.platformType" :min="1" :precision="0" />
        </el-form-item>
        <el-form-item label="通知地址">
          <el-input v-model="platformForm.notifyUrl" />
        </el-form-item>
        <el-form-item label="返回地址">
          <el-input v-model="platformForm.returnUrl" />
        </el-form-item>
        <el-form-item label="图标">
          <el-input v-model="platformForm.icon" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="platformForm.status">
            <el-radio :value="1">
              启用
            </el-radio>
            <el-radio :value="2">
              禁用
            </el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="platformForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="platformDialogVisible = false">
          取消
        </el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitPlatform">
          确定
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="productDialogVisible" :title="productForm.id ? '编辑产品' : '新增产品'" width="640px">
      <el-form label-width="100px">
        <el-form-item v-if="!productForm.id" label="平台ID">
          <el-input-number v-model="productForm.platformId" :min="1" :precision="0" />
        </el-form-item>
        <el-form-item v-if="!productForm.id" label="产品编码">
          <el-input v-model="productForm.productCode" />
        </el-form-item>
        <el-form-item label="产品名称">
          <el-input v-model="productForm.productName" />
        </el-form-item>
        <el-form-item label="场景类型">
          <el-input-number v-model="productForm.sceneType" :min="1" :precision="0" />
        </el-form-item>
        <el-form-item label="币种">
          <el-input v-model="productForm.currency" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="productForm.status">
            <el-radio :value="1">
              启用
            </el-radio>
            <el-radio :value="2">
              禁用
            </el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="productForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="productDialogVisible = false">
          取消
        </el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitProduct">
          确定
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" :title="detailTitle" width="720px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { catalogService, type PayPlatform, type PayProduct } from '@/services'

const activeTab = ref('platforms')
const submitLoading = ref(false)
const platformLoading = ref(false)
const productLoading = ref(false)
const platforms = ref<PayPlatform[]>([])
const products = ref<PayProduct[]>([])
const detailVisible = ref(false)
const detailTitle = ref('详情')
const detailData = ref<Record<string, any>>({})
const platformDialogVisible = ref(false)
const productDialogVisible = ref(false)

const platformQuery = reactive({ platformCode: '', keyword: '', status: 0 })
const productQuery = reactive({ platformId: 0, productCode: '', keyword: '' })

const platformForm = reactive({
  id: 0,
  platformCode: '',
  platformName: '',
  platformType: 1,
  notifyUrl: '',
  returnUrl: '',
  icon: '',
  status: 1,
  remark: '',
})

const productForm = reactive({
  id: 0,
  platformId: 0,
  productCode: '',
  productName: '',
  sceneType: 1,
  currency: '',
  status: 1,
  remark: '',
})

const loadPlatforms = async () => {
  platformLoading.value = true
  try {
    const res = await catalogService.getPlatformList({ ...platformQuery, limit: 100 })
    platforms.value = res.data || []
  } finally {
    platformLoading.value = false
  }
}

const loadProducts = async () => {
  productLoading.value = true
  try {
    const res = await catalogService.getProductList({
      ...productQuery,
      platformId: productQuery.platformId || undefined,
      limit: 100,
    })
    products.value = res.data || []
  } finally {
    productLoading.value = false
  }
}

const refreshCurrentPage = () => {
  loadPlatforms()
  loadProducts()
}

const resetPlatformQuery = () => {
  Object.assign(platformQuery, { platformCode: '', keyword: '', status: 0 })
  loadPlatforms()
}

const resetProductQuery = () => {
  Object.assign(productQuery, { platformId: 0, productCode: '', keyword: '' })
  loadProducts()
}

const openPlatformDialog = (row?: PayPlatform) => {
  Object.assign(platformForm, row || {
    id: 0,
    platformCode: '',
    platformName: '',
    platformType: 1,
    notifyUrl: '',
    returnUrl: '',
    icon: '',
    status: 1,
    remark: '',
  })
  platformDialogVisible.value = true
}

const openProductDialog = (row?: PayProduct) => {
  Object.assign(productForm, row || {
    id: 0,
    platformId: 0,
    productCode: '',
    productName: '',
    sceneType: 1,
    currency: '',
    status: 1,
    remark: '',
  })
  productDialogVisible.value = true
}

const submitPlatform = async () => {
  submitLoading.value = true
  try {
    if (platformForm.id) {
      await catalogService.updatePlatform(platformForm.id, { ...platformForm })
    } else {
      await catalogService.createPlatform({ ...platformForm })
    }
    ElMessage.success('操作成功')
    platformDialogVisible.value = false
    loadPlatforms()
  } finally {
    submitLoading.value = false
  }
}

const submitProduct = async () => {
  submitLoading.value = true
  try {
    if (productForm.id) {
      await catalogService.updateProduct(productForm.id, { ...productForm })
    } else {
      await catalogService.createProduct({ ...productForm })
    }
    ElMessage.success('操作成功')
    productDialogVisible.value = false
    loadProducts()
  } finally {
    submitLoading.value = false
  }
}

const showPlatformDetail = async (row: PayPlatform) => {
  const res = await catalogService.getPlatformDetail(row.id)
  detailTitle.value = '平台详情'
  detailData.value = res.data || row
  detailVisible.value = true
}

const showProductDetail = async (row: PayProduct) => {
  const res = await catalogService.getProductDetail(row.id)
  detailTitle.value = '产品详情'
  detailData.value = res.data || row
  detailVisible.value = true
}

onMounted(refreshCurrentPage)
</script>

<style scoped>
.query-card :deep(.el-form-item) {
  margin-bottom: 12px;
}
</style>
