<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="load" @reset="reset">
      <el-form-item :label="t('trade.tenantId')">
        <TenantSelect v-model="query.tenantId" />
      </el-form-item><el-form-item :label="t('trade.settleAsset')">
        <el-input v-model="query.settleAsset" clearable />
      </el-form-item><el-button v-perm="'trade:insurance-fund-account:update'" type="primary" @click="edit()">
        {{ t('common.add') }}
      </el-button><el-button v-perm="'trade:insurance-fund-account:update'" @click="openPlatformAccount">
        {{ t('trade.platformAccount') }}
      </el-button>
    </CrudQueryCard><el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column prop="symbolId" :label="t('trade.symbolId')" /><el-table-column
          prop="settleAsset"
          :label="t('trade.settleAsset')"
        /><el-table-column prop="adlEnabled" :label="t('trade.adlEnabled')" /><el-table-column
          prop="status"
          :label="t('trade.status')"
        /><el-table-column :label="t('common.actions')">
          <template #default="{ row }">
            <el-button link @click="edit(row)">
              {{ t('common.edit') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table><CursorPagination
        v-model:limit="pagination.limit"
        :total="pagination.total"
        :has-prev="pagination.hasPrev"
        :has-next="pagination.hasNext"
        @prev="prevAndLoad(load)"
        @next="nextAndLoad(load)"
        @limit-change="resetAndLoad(load)"
      />
      >
    </el-card><el-dialog v-model="visible" :title="t('trade.insuranceFundAccounts')" width="520">
      <el-form :model="form" label-width="150">
        <el-form-item :label="t('trade.tenantId')">
          <TenantSelect v-model="form.tenantId" />
        </el-form-item><el-form-item :label="t('trade.symbolId')">
          <el-input-number v-model="form.symbolId" :min="0" />
        </el-form-item><el-form-item :label="t('trade.settleAsset')">
          <el-input v-model="form.settleAsset" />
        </el-form-item><el-form-item :label="t('trade.adlEnabled')">
          <el-switch v-model="adl" />
        </el-form-item>
      </el-form><template #footer>
        <el-button @click="visible = false">
          {{ t('common.cancel') }}
        </el-button><el-button type="primary" @click="save">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog><el-dialog v-model="platformVisible" :title="t('trade.platformAccount')" width="560">
      <el-form :model="platformForm" label-width="150">
        <el-form-item :label="t('trade.tenantId')">
          <TenantSelect v-model="platformForm.tenantId" />
        </el-form-item><el-form-item :label="t('trade.accountType')">
          <el-select v-model="platformForm.accountType" style="width: 100%">
            <el-option label="INSURANCE_FUND" value="INSURANCE_FUND" />
            <el-option label="OPTION_BACKSTOP" value="OPTION_BACKSTOP" />
          </el-select>
        </el-form-item><el-form-item :label="t('trade.settleAsset')">
          <el-input v-model="platformForm.coin" />
        </el-form-item><el-form-item :label="t('trade.availableAmount')">
          <el-input
            :model-value="platformAccount?.availableAmount || '0'"
            disabled
          />
        </el-form-item><el-form-item :label="t('trade.adjustDirection')">
          <el-radio-group v-model="platformForm.direction">
            <el-radio-button :value="1">
              {{ t('trade.increase') }}
            </el-radio-button><el-radio-button v-if="platformForm.accountType === 'INSURANCE_FUND'" :value="2">
              {{ t('trade.decrease') }}
            </el-radio-button>
          </el-radio-group>
        </el-form-item><el-form-item :label="t('trade.adjustAmount')">
          <el-input v-model="platformForm.amount" />
        </el-form-item><el-form-item :label="t('trade.requestNo')">
          <el-input v-model="platformForm.requestNo" />
        </el-form-item><el-form-item :label="t('common.remark')">
          <el-input v-model="platformForm.remark" />
        </el-form-item>
      </el-form><template #footer>
        <el-button @click="loadPlatformAccount">
          {{ t('common.search') }}
        </el-button><el-button type="primary" @click="createPlatformAccount">
          {{ t('trade.createAccount') }}
        </el-button><el-button type="success" @click="adjustPlatformAccount">
          {{ t('trade.adjustBalance') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>
<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'
import TenantSelect from '@/components/TenantSelect.vue'
import { usePagination } from '@/composables'
import { apiTradeListInsuranceFundAccounts, apiTradeSetInsuranceFundAccount } from '@/api/trade'
import { apiAdjustPlatformAccount, apiGetPlatformAccount, apiSetPlatformAccount } from '@/api/asset'
import type { InsuranceFundAccount, PlatformAccount } from '@/services'
const { t } = useI18n(),
  loading = ref(false),
  visible = ref(false),
  platformVisible = ref(false),
  platformAccount = ref<PlatformAccount>(),
  rows = ref<InsuranceFundAccount[]>([]),
  query = reactive({ tenantId: undefined as number | undefined, settleAsset: '' }),
  form = reactive({
    id: 0,
    tenantId: 0,
    symbolId: 0,
    settleAsset: 'USDT',
    adlEnabled: 2,
    status: 1,
    version: 0,
  }),
  platformForm = reactive({
    tenantId: 0,
    accountType: 'INSURANCE_FUND',
    coin: 'USDT',
    direction: 1,
    amount: '',
    requestNo: '',
    remark: '',
  })
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)
const adl = computed({
  get: () => form.adlEnabled === 1,
  set: (v) => (form.adlEnabled = v ? 1 : 2),
})
watch(
  () => platformForm.accountType,
  (accountType) => {
    if (accountType === 'OPTION_BACKSTOP') platformForm.direction = 1
    platformAccount.value = undefined
  },
)
async function load() {
  loading.value = true
  try {
    const r = await apiTradeListInsuranceFundAccounts({
      ...query,
      cursor: pagination.cursor,
      limit: pagination.limit,
    })
    rows.value = r.data || []
    updateFromResponse(r)
  } finally {
    loading.value = false
  }
}
function reset() {
  query.tenantId = undefined
  query.settleAsset = ''
  resetAndLoad(load)
}
function edit(r?: InsuranceFundAccount) {
  Object.assign(
    form,
    r || {
      id: 0,
      tenantId: query.tenantId || 0,
      symbolId: 0,
      settleAsset: 'USDT',
      adlEnabled: 2,
      status: 1,
      version: 0,
    },
  )
  visible.value = true
}
async function save() {
  await apiTradeSetInsuranceFundAccount(form)
  ElMessage.success(t('common.success'))
  visible.value = false
  load()
}
function openPlatformAccount() {
  platformForm.tenantId = query.tenantId || 0
  platformForm.coin = query.settleAsset || 'USDT'
  platformVisible.value = true
  if (platformForm.tenantId > 0) loadPlatformAccount()
}
async function loadPlatformAccount() {
  const r = await apiGetPlatformAccount({
    tenantId: platformForm.tenantId,
    accountType: platformForm.accountType,
    coin: platformForm.coin,
  })
  platformAccount.value = r.data
}
async function createPlatformAccount() {
  const r = await apiSetPlatformAccount({
    tenantId: platformForm.tenantId,
    accountType: platformForm.accountType,
    coin: platformForm.coin,
    status: 1,
    version: platformAccount.value?.version || 0,
  })
  platformAccount.value = r.data
  ElMessage.success(t('common.success'))
}
async function adjustPlatformAccount() {
  const r = await apiAdjustPlatformAccount(platformForm)
  platformAccount.value = r.data
  ElMessage.success(t('common.success'))
}
onMounted(load)
</script>
