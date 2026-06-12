<template>
  <div class="payment-page module-page">
    <div class="page-header">
      <h2>{{ t('payment.platforms') }}</h2>
      <div class="header-actions">
        <el-button @click="loadPlatforms">
          <el-icon><Refresh /></el-icon>
          {{ t('common.refresh') }}
        </el-button>
      </div>
    </div>

    <CrudQueryCard :model="platformQuery" label-width="90px" :show-actions="false">
      <el-form-item :label="t('payment.platformCode')">
        <el-input v-model="platformQuery.platformCode" clearable />
      </el-form-item>
      <el-form-item :label="t('common.keyword')">
        <el-input v-model="platformQuery.keyword" clearable />
      </el-form-item>
      <el-form-item :label="t('common.enabled')">
        <el-select v-model="platformQuery.enabled" clearable style="width: 160px">
          <el-option :label="t('payment.all')" :value="0" />
          <el-option
            v-for="item in enabledOptions"
            :key="item.value"
            :label="getOptionLabel(t, item.code, item.value)"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="loadPlatforms">
          {{ t('common.search') }}
        </el-button>
        <el-button @click="resetPlatformQuery">
          {{ t('common.reset') }}
        </el-button>
        <el-button v-perm="'payment:platform:add'" type="primary" @click="openPlatformDialog()">
          {{ t('payment.addPlatform') }}
        </el-button>
      </el-form-item>
    </CrudQueryCard>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="platformLoading" :data="platforms" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="platformCode" :label="t('payment.platformCode')" min-width="140" />
        <el-table-column prop="platformName" :label="t('payment.platformName')" min-width="160" />
        <el-table-column :label="t('payment.platformType')" width="120">
          <template #default="{ row }">
            {{ getPlatformTypeLabel(row.platformType) }}
          </template>
        </el-table-column>
        <!-- <el-table-column prop="icon" :label="t('common.icon')" width="100" align="center">
          <template #default="{ row }">
            <el-image
              v-if="row.icon"
              :src="resolveAssetUrl(row.icon)"
              class="platform-icon-preview"
              :preview-teleported="true"
            />
            <span v-else>-</span>
          </template>
        </el-table-column> -->
        <el-table-column
          prop="notifyUrl"
          :label="t('payment.notifyUrl')"
          min-width="220"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.enabled')" width="100">
          <template #default="{ row }">
            <el-tag :type="row.enabled === 1 ? 'success' : 'info'">
              {{ getOptionValueLabel(optionGroups, 'enabled', row.enabled, t) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="remark"
          :label="t('common.remark')"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.actions')" width="160" fixed="right">
          <template #default="{ row }">
            <el-button
              v-perm="'payment:platform:detail'"
              link
              type="primary"
              @click="showPlatformDetail(row)"
            >
              {{ t('common.detail') }}
            </el-button>
            <el-button
              v-perm="'payment:platform:update'"
              link
              type="primary"
              @click="openPlatformDialog(row)"
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
      v-model="platformDialogVisible"
      :title="platformForm.id ? t('payment.editPlatform') : t('payment.addPlatform')"
      width="640px"
    >
      <el-form label-width="100px">
        <el-form-item v-if="!platformForm.id" :label="t('payment.systemPlatform')">
          <el-select
            v-model="selectedPlatformCode"
            filterable
            :placeholder="t('payment.pleaseSelectSystemPlatform')"
            style="width: 100%"
            @change="handlePlatformChange"
          >
            <el-option
              v-for="item in supportedPlatforms"
              :key="item.platformCode"
              :label="`${item.platformName} (${item.platformCode})`"
              :value="item.platformCode"
            />
          </el-select>
        </el-form-item>
        <el-form-item v-if="!platformForm.id" :label="t('payment.platformCode')">
          <el-input v-model="platformForm.platformCode" disabled />
        </el-form-item>
        <el-form-item :label="t('payment.platformName')">
          <el-input v-model="platformForm.platformName" :disabled="!platformForm.id" />
        </el-form-item>
        <el-form-item :label="t('payment.platformType')">
          <el-select v-model="platformForm.platformType" style="width: 100%">
            <el-option
              v-for="item in platformTypeOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('payment.notifyUrl')">
          <el-input v-model="platformForm.notifyUrl" />
        </el-form-item>
        <el-form-item :label="t('payment.returnUrl')">
          <el-input v-model="platformForm.returnUrl" />
        </el-form-item>
        <el-form-item :label="t('common.icon')">
          <div class="platform-icon-upload">
            <el-image
              v-if="platformForm.icon"
              :src="resolveAssetUrl(platformForm.icon)"
              class="platform-icon-preview"
              :preview-teleported="true"
            />
            <el-upload
              action="#"
              :auto-upload="false"
              :show-file-list="false"
              :on-change="handlePlatformIconSelect"
              accept="image/*"
            >
              <el-button type="primary">
                {{ t('payment.uploadImage') }}
              </el-button>
            </el-upload>
          </div>
        </el-form-item>
        <el-form-item :label="t('common.enabled')">
          <el-select v-model="platformForm.enabled" style="width: 100%">
            <el-option
              v-for="item in enabledOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('common.remark')">
          <el-input v-model="platformForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="platformDialogVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button
          v-perm="platformForm.id ? 'payment:platform:update' : 'payment:platform:add'"
          type="primary"
          :loading="submitLoading"
          @click="submitPlatform"
        >
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-drawer
      v-model="detailVisible"
      :title="detailTitle"
      direction="rtl"
      size="520px"
    >
      <div v-if="detailData" class="platform-detail">
        <div class="detail-icon-panel">
          <el-image
            v-if="detailData.icon"
            :src="resolveAssetUrl(detailData.icon)"
            class="detail-icon-preview"
            fit="contain"
            :preview-src-list="[resolveAssetUrl(detailData.icon)]"
            :preview-teleported="true"
          />
          <div v-else class="detail-icon-empty">
            -
          </div>
          <div class="detail-icon-meta">
            <div class="detail-icon-title">
              {{ t('common.icon') }}
            </div>
            <div class="detail-icon-url">
              {{ detailData.icon || '-' }}
            </div>
          </div>
        </div>

        <PaymentDetailDescriptions
          :data="detailDisplayData"
          :option-groups="optionGroups"
          :columns="1"
        />
      </div>
      <el-empty v-else :description="t('common.noData')" />
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables'
import { Refresh } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import type { UploadFile } from 'element-plus'
import {
  catalogService,
  type OptionGroup,
  type PayPlatform,
  type PayPlatformItem,
} from '@/services'
import PaymentDetailDescriptions from '@/components/payment/PaymentDetailDescriptions.vue'
import { apiUploadFile } from '@/api/system/upload'
import { buildSystemAssetUrl, useSystemCore } from '@/composables/useSystemCore'
import { findOptionGroup, getOptionLabel, getOptionValueLabel } from '@/utils/options'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)
const { systemCore, loadSystemCore } = useSystemCore()

const submitLoading = ref(false)
const platformLoading = ref(false)
const platforms = ref<PayPlatform[]>([])
const supportedPlatforms = ref<PayPlatformItem[]>([])
const detailVisible = ref(false)
const detailTitle = ref('')
const detailData = ref<PayPlatform | null>(null)
const platformDialogVisible = ref(false)
const selectedPlatformCode = ref('')
const optionGroups = ref<OptionGroup[]>([])
const platformTypeOptions = computed(() => findOptionGroup(optionGroups.value, 'platformType'))
const enabledOptions = computed(() => findOptionGroup(optionGroups.value, 'enabled'))
const detailDisplayData = computed(() => {
  if (!detailData.value) return null

  const { icon: _icon, ...rest } = detailData.value
  return rest
})

const platformQuery = reactive({ platformCode: '', keyword: '', enabled: 0 })

const platformForm = reactive({
  id: 0,
  platformCode: '',
  platformName: '',
  platformType: 1,
  notifyUrl: '',
  returnUrl: '',
  icon: '',
  enabled: 1,
  remark: '',
})

const loadPlatforms = async () => {
  platformLoading.value = true
  try {
    const res = await catalogService.getPlatformList({
      ...platformQuery,
      cursor: pagination.cursor,
      limit: pagination.limit,
    })
    platforms.value = res.data || []
    updateFromResponse(res)
  } finally {
    platformLoading.value = false
  }
}

const loadSupportedPlatforms = async () => {
  const res = await catalogService.getPayPlatforms()
  supportedPlatforms.value = res.data || []
}

const loadOptions = async () => {
  const res = await catalogService.getOptions()
  optionGroups.value = res.data || []
}

const resetPlatformQuery = () => {
  Object.assign(platformQuery, { platformCode: '', keyword: '', enabled: 0 })
  loadPlatforms()
}

const openPlatformDialog = (row?: PayPlatform) => {
  selectedPlatformCode.value = ''
  Object.assign(
    platformForm,
    row || {
      id: 0,
      platformCode: '',
      platformName: '',
      platformType: 1,
      notifyUrl: '',
      returnUrl: '',
      icon: '',
      enabled: 1,
      remark: '',
    },
  )
  platformDialogVisible.value = true
}

const handlePlatformChange = (platformCode: string) => {
  const matched = supportedPlatforms.value.find(
    (item: { platformCode: string }) => item.platformCode === platformCode,
  )
  if (!matched) return
  platformForm.platformCode = matched.platformCode
  platformForm.platformName = matched.platformName
}

const getPlatformTypeLabel = (value: number) => {
  return getOptionValueLabel(optionGroups.value, 'platformType', value, t)
}

const resolveAssetUrl = (url?: string) => buildSystemAssetUrl(systemCore.value.assetUrl, url)

const handlePlatformIconSelect = async (uploadFile: UploadFile) => {
  if (!uploadFile.raw) return

  if (!uploadFile.raw.type.startsWith('image/')) {
    ElMessage.error(t('app.pleaseSelectImageFile'))
    return
  }

  if (uploadFile.raw.size > 5 * 1024 * 1024) {
    ElMessage.error(t('app.avatarSizeLimit'))
    return
  }

  submitLoading.value = true
  try {
    const res = await apiUploadFile(uploadFile.raw)
    if (res.code === 200) {
      platformForm.icon = res.data?.url || ''
      ElMessage.success(t('common.uploadSuccess'))
      return
    }
    throw new Error(res.msg || t('common.uploadFailed'))
  } catch (error: unknown) {
    ElMessage.error(error instanceof Error ? error.message : t('common.uploadFailed'))
  } finally {
    submitLoading.value = false
  }
}

const submitPlatform = async () => {
  if (!platformForm.id && !platformForm.platformCode) {
    ElMessage.warning(t('payment.pleaseSelectSystemPlatform'))
    return
  }

  submitLoading.value = true
  try {
    if (platformForm.id) {
      await catalogService.updatePlatform(platformForm.id, { ...platformForm })
    } else {
      await catalogService.createPlatform({ ...platformForm })
    }
    ElMessage.success(t('common.operationSuccess'))
    platformDialogVisible.value = false
    loadPlatforms()
  } finally {
    submitLoading.value = false
  }
}

const showPlatformDetail = async (row: PayPlatform) => {
  const res = await catalogService.getPlatformDetail(row.id)
  detailTitle.value = t('payment.platformDetail')
  detailData.value = res.data || row
  detailVisible.value = true
}

function handleLimitChange() {
  resetAndLoad(loadPlatforms)
}

function handlePrevPage() {
  prevAndLoad(loadPlatforms)
}

function handleNextPage() {
  nextAndLoad(loadPlatforms)
}

onMounted(async () => {
  await Promise.all([loadSystemCore(), loadPlatforms(), loadSupportedPlatforms(), loadOptions()])
})
</script>

<style scoped>
.query-card :deep(.el-form-item) {
  margin-bottom: 12px;
}

.platform-icon-upload {
  display: flex;
  align-items: center;
  gap: 12px;
}

.platform-icon-preview {
  width: 72px;
  height: 72px;
  border: 1px solid var(--el-border-color);
  border-radius: 6px;
  object-fit: contain;
  background: var(--el-fill-color-light);
}

.platform-detail {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.detail-icon-panel {
  display: flex;
  gap: 14px;
  align-items: center;
  padding: 16px;
  border: 1px solid var(--el-border-color-light);
  border-radius: 8px;
  background: var(--el-fill-color-extra-light);
}

.detail-icon-preview,
.detail-icon-empty {
  width: 96px;
  height: 96px;
  flex: 0 0 auto;
  border: 1px solid var(--el-border-color);
  border-radius: 8px;
  background: #fff;
}

.detail-icon-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--el-text-color-placeholder);
}

.detail-icon-meta {
  min-width: 0;
}

.detail-icon-title {
  margin-bottom: 8px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.detail-icon-url {
  color: var(--el-text-color-secondary);
  word-break: break-all;
  line-height: 1.5;
}
</style>
