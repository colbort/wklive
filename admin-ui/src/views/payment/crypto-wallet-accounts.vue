<template>
  <div class="payment-page module-page">
    <CrudQueryCard
      :model="query"
      label-width="auto"
      @search="loadList"
      @reset="resetQuery"
    >
      <el-form-item :label="t('common.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('common.keyword')">
        <el-input v-model="query.keyword" clearable />
      </el-form-item>
      <el-form-item :label="t('payment.provider')">
        <el-input v-model="query.provider" clearable />
      </el-form-item>
      <el-form-item :label="t('common.enabled')">
        <el-select v-model="query.enabled" clearable style="width: 140px">
          <el-option :label="t('common.enabled')" :value="1" /><el-option
            :label="t('common.disabled')"
            :value="2"
          />
        </el-select>
      </el-form-item>
      <template #actions>
        <el-button
          v-perm="'payment:crypto-wallet-account:add'"
          type="primary"
          @click="openDialog()"
        >
          {{ t('common.add') }}
        </el-button>
      </template>
    </CrudQueryCard>
    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" :label="t('common.id')" width="80" /><el-table-column
          prop="tenantId"
          :label="t('common.tenantId')"
          width="90"
        />
        <el-table-column
          prop="accountCode"
          :label="t('payment.accountCode')"
          min-width="150"
        /><el-table-column
          prop="accountName"
          :label="t('payment.accountName')"
          min-width="150"
        /><el-table-column
          prop="provider"
          :label="t('payment.provider')"
          width="120"
        /><el-table-column
          prop="isDefault"
          :label="t('common.default')"
          width="80"
        /><el-table-column prop="enabled" :label="t('common.enabled')" width="80" />
        <el-table-column
          :label="t('common.actions')"
          align="center"
          width="140"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              v-perm="'payment:crypto-wallet-account:detail'"
              link
              type="primary"
              @click="showDetail(row)"
            >
              {{ t('common.detail') }}
            </el-button>
            <el-button
              v-perm="'payment:crypto-wallet-account:update'"
              link
              type="primary"
              @click="openDialog(row)"
            >
              {{ t('common.edit') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <CursorPagination
        v-model:limit="pagination.limit"
        :total="pagination.total"
        :has-prev="pagination.hasPrev"
        :has-next="pagination.hasNext"
        @prev="handlePrevPage"
        @next="handleNextPage"
        @limit-change="handleLimitChange"
      />
    </el-card>
    <el-dialog
      v-model="dialogVisible"
      :title="form.id ? t('payment.editCryptoWalletAccount') : t('payment.addCryptoWalletAccount')"
      width="680px"
    >
      <el-form label-width="130px">
        <el-form-item :label="t('common.tenantId')">
          <TenantSelect v-model="form.tenantId" include-system />
        </el-form-item>
        <el-form-item :label="t('payment.accountCode')">
          <el-input v-model="form.accountCode" :disabled="Boolean(form.id)" />
        </el-form-item>
        <el-form-item :label="t('payment.accountName')">
          <el-input v-model="form.accountName" />
        </el-form-item>
        <el-form-item :label="t('payment.provider')">
          <el-input v-model="form.provider" />
        </el-form-item>
        <el-form-item label="API Key">
          <el-input v-model="form.apiKeyCipher" type="textarea" />
        </el-form-item>
        <el-form-item label="API Secret">
          <el-input v-model="form.apiSecretCipher" type="textarea" />
        </el-form-item>
        <el-form-item :label="t('payment.callbackSecret')">
          <el-input v-model="form.callbackSecretCipher" type="textarea" />
        </el-form-item>
        <el-form-item :label="t('payment.extConfig')">
          <el-input v-model="form.extConfig" type="textarea" />
        </el-form-item>
        <el-form-item :label="t('common.enabled')">
          <el-select v-model="form.enabled" style="width: 100%">
            <el-option :label="t('common.enabled')" :value="1" /><el-option
              :label="t('common.disabled')"
              :value="2"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('common.default')">
          <el-switch v-model="form.isDefault" :active-value="1" :inactive-value="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button
          v-perm="
            form.id ? 'payment:crypto-wallet-account:update' : 'payment:crypto-wallet-account:add'
          "
          type="primary"
          @click="submit"
        >
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>
    <el-dialog
      v-model="detailVisible"
      :title="t('payment.cryptoWalletAccountDetail')"
      width="760px"
    >
      <PaymentDetailDescriptions :data="detailData" />
    </el-dialog>
  </div>
</template>
<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables'
import { ElMessage } from 'element-plus'
import { cryptoService, type CryptoWalletAccount } from '@/services'
import TenantSelect from '@/components/TenantSelect.vue'
import PaymentDetailDescriptions from '@/components/payment/PaymentDetailDescriptions.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'
const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)
const loading = ref(false),
  dialogVisible = ref(false),
  detailVisible = ref(false)
const list = ref<CryptoWalletAccount[]>([]),
  detailData = ref<CryptoWalletAccount | null>(null)
const query = reactive({
  tenantId: 0,
  keyword: '',
  provider: '',
  enabled: undefined as number | undefined,
})
const form = reactive({
  id: 0,
  tenantId: 0,
  accountCode: '',
  accountName: '',
  provider: 'self',
  apiKeyCipher: '',
  apiSecretCipher: '',
  callbackSecretCipher: '',
  extConfig: '',
  enabled: 1,
  isDefault: 2,
  createTimes: 0,
  updateTimes: 0,
})
function params() {
  return Object.fromEntries(
    Object.entries(query).filter(([, v]) => v !== '' && v !== 0 && v !== undefined),
  )
}
async function loadList() {
  loading.value = true
  try {
    const res = await cryptoService.listWalletAccounts({
      ...params(),
      cursor: pagination.cursor,
      limit: pagination.limit,
    })
    list.value = res.data || []
    updateFromResponse(res)
  } finally {
    loading.value = false
  }
}
function resetQuery() {
  Object.assign(query, { tenantId: 0, keyword: '', provider: '', enabled: undefined })
  void loadList()
}
function openDialog(row?: CryptoWalletAccount) {
  Object.assign(
    form,
    row || {
      id: 0,
      tenantId: 0,
      accountCode: '',
      accountName: '',
      provider: 'self',
      apiKeyCipher: '',
      apiSecretCipher: '',
      callbackSecretCipher: '',
      extConfig: '',
      enabled: 1,
      isDefault: 2,
      createTimes: 0,
      updateTimes: 0,
    },
  )
  dialogVisible.value = true
}
function showDetail(row: CryptoWalletAccount) {
  detailData.value = row
  detailVisible.value = true
}
async function submit() {
  const res = form.id
    ? await cryptoService.updateWalletAccount(form)
    : await cryptoService.createWalletAccount(form)
  if (res.code === 200) {
    ElMessage.success(t('common.success'))
    dialogVisible.value = false
    await loadList()
  }
}
function handleLimitChange() {
  resetAndLoad(loadList)
}

function handlePrevPage() {
  prevAndLoad(loadList)
}

function handleNextPage() {
  nextAndLoad(loadList)
}

onMounted(loadList)
</script>
