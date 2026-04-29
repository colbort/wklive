<template>
  <div class="payment-page module-page">
    <div class="page-header">
      <h2>链上充值地址</h2>
      <div class="header-actions">
        <el-button @click="loadList">刷新</el-button>
        <el-button type="primary" @click="openDialog()">新增地址</el-button>
      </div>
    </div>

    <el-card shadow="never" class="query-card">
      <el-form :model="query" inline label-width="90px">
        <el-form-item label="租户ID"><el-input-number v-model="query.tenantId" :min="0" :precision="0" /></el-form-item>
        <el-form-item label="用户ID"><el-input-number v-model="query.userId" :min="0" :precision="0" /></el-form-item>
        <el-form-item label="账户类型"><el-input-number v-model="query.walletType" :min="0" :precision="0" /></el-form-item>
        <el-form-item label="币种"><el-input v-model="query.coin" clearable /></el-form-item>
        <el-form-item label="链">
          <el-select v-model="query.chainCode" clearable style="width: 150px">
            <el-option
              v-for="item in chainCodeOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="String(item.value)"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="地址"><el-input v-model="query.address" clearable /></el-form-item>
        <el-form-item label="状态">
          <el-select v-model="query.status" clearable style="width: 150px">
            <el-option label="禁用" :value="1" />
            <el-option label="可用" :value="2" />
            <el-option label="冻结" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadList">搜索</el-button>
          <el-button @click="resetQuery">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never">
      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="tenantId" label="租户" width="90" />
        <el-table-column prop="userId" label="用户" width="100" />
        <el-table-column prop="walletType" label="账户" width="80" />
        <el-table-column prop="coin" label="币种" width="90" />
        <el-table-column label="链" width="110">
          <template #default="{ row }">{{ formatChainCode(row.chainCode) }}</template>
        </el-table-column>
        <el-table-column prop="address" label="地址" min-width="260" show-overflow-tooltip />
        <el-table-column prop="memo" label="Memo" min-width="120" show-overflow-tooltip />
        <el-table-column label="类型" width="130">
          <template #default="{ row }">{{ row.addressType === 2 ? '公共地址+memo' : '用户独享' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }"><el-tag>{{ statusText(row.status) }}</el-tag></template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">详情</el-button>
            <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑地址' : '新增地址'" width="680px">
      <el-form label-width="110px">
        <el-form-item label="租户ID"><el-input-number v-model="form.tenantId" :min="0" :precision="0" /></el-form-item>
        <template v-if="!form.id">
          <el-form-item label="用户ID"><el-input-number v-model="form.userId" :min="0" :precision="0" /></el-form-item>
          <el-form-item label="账户类型"><el-input-number v-model="form.walletType" :min="1" :precision="0" /></el-form-item>
          <el-form-item label="币种"><el-input v-model="form.coin" /></el-form-item>
          <el-form-item label="链">
            <el-select v-model="form.chainCode" style="width: 100%">
              <el-option
                v-for="item in chainCodeOptions"
                :key="item.value"
                :label="getOptionLabel(t, item.code, item.value)"
                :value="String(item.value)"
              />
            </el-select>
          </el-form-item>
        </template>
        <el-form-item label="地址"><el-input v-model="form.address" /></el-form-item>
        <el-form-item label="Memo"><el-input v-model="form.memo" /></el-form-item>
        <el-form-item label="来源">
          <el-select v-model="form.addressSource" style="width: 100%"><el-option label="系统生成" :value="1" /><el-option label="第三方分配" :value="2" /><el-option label="手工导入" :value="3" /></el-select>
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="form.addressType" style="width: 100%"><el-option label="用户独享" :value="1" /><el-option label="公共地址+memo" :value="2" /></el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status" style="width: 100%"><el-option label="禁用" :value="1" /><el-option label="可用" :value="2" /><el-option label="冻结" :value="3" /></el-select>
        </el-form-item>
      </el-form>
      <template #footer><el-button @click="dialogVisible = false">取消</el-button><el-button type="primary" @click="submit">确认</el-button></template>
    </el-dialog>
    <el-dialog v-model="detailVisible" title="地址详情" width="760px"><pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre></el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { catalogService, cryptoService, type CryptoRechargeAddress, type OptionGroup } from '@/services'
import { findOptionGroup, getOptionLabel } from '@/utils/options'

const { t } = useI18n()

const loading = ref(false)
const list = ref<CryptoRechargeAddress[]>([])
const dialogVisible = ref(false)
const detailVisible = ref(false)
const detailData = ref<CryptoRechargeAddress | null>(null)
const optionGroups = ref<OptionGroup[]>([])
const chainCodeOptions = computed(() => findOptionGroup(optionGroups.value, 'chainCode'))
const query = reactive({ tenantId: 0, userId: 0, walletType: 0, coin: '', chainCode: '', address: '', status: undefined as number | undefined })
const form = reactive({ id: 0, tenantId: 0, userId: 0, walletType: 1, coin: 'USDT', chainCode: '20', address: '', memo: '', addressSource: 3, addressType: 1, status: 2 })

function params() {
  return Object.fromEntries(Object.entries(query).filter(([, v]) => v !== '' && v !== 0 && v !== undefined))
}
async function loadList() {
  loading.value = true
  try { list.value = (await cryptoService.listRechargeAddresses({ ...params(), limit: 100 })).data || [] } finally { loading.value = false }
}
async function loadOptions() { optionGroups.value = (await catalogService.getOptions()).data || [] }
function resetQuery() { Object.assign(query, { tenantId: 0, userId: 0, walletType: 0, coin: '', chainCode: '', address: '', status: undefined }); void loadList() }
function openDialog(row?: CryptoRechargeAddress) { Object.assign(form, row ? { ...row, chainCode: String(row.chainCode || '') } : { id: 0, tenantId: 0, userId: 0, walletType: 1, coin: 'USDT', chainCode: '20', address: '', memo: '', addressSource: 3, addressType: 1, status: 2 }); dialogVisible.value = true }
function showDetail(row: CryptoRechargeAddress) { detailData.value = row; detailVisible.value = true }
function statusText(status: number) { return status === 1 ? '禁用' : status === 3 ? '冻结' : '可用' }
function formatChainCode(value: string) {
  const item = chainCodeOptions.value.find((option) => String(option.value) === String(value))
  return item ? getOptionLabel(t, item.code, item.value) : value
}
async function submit() {
  const res = form.id ? await cryptoService.updateRechargeAddress(form) : await cryptoService.createRechargeAddress(form)
  if (res.code === 200 || res.code === 0) { ElMessage.success('操作成功'); dialogVisible.value = false; await loadList() }
}
onMounted(() => { void loadOptions(); void loadList() })
</script>
