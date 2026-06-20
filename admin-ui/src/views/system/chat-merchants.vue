<template>
  <div class="sys-chat-merchants module-page">
    <CrudQueryCard :model="queryForm" @search="loadList" @reset="resetQuery">
      <el-form-item :label="t('system.chatMerchantCode')">
        <el-input
          v-model="queryForm.merchantCode"
          :placeholder="t('system.chatMerchantCodePlaceholder')"
          clearable
          @keyup.enter="loadList"
        />
      </el-form-item>
      <el-form-item :label="t('system.chatMerchantName')">
        <el-input
          v-model="queryForm.merchantName"
          :placeholder="t('system.chatMerchantNamePlaceholder')"
          clearable
          @keyup.enter="loadList"
        />
      </el-form-item>
      <el-form-item :label="t('system.enabled')">
        <el-select
          v-model="queryForm.enabled"
          :placeholder="t('system.pleaseSelectStatus')"
          clearable
          @change="loadList"
        >
          <el-option
            v-for="item in enabledSelectOptions"
            :key="item.value"
            :label="enabledOptionLabel(item.value, item.code)"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <template #actions>
        <el-button v-perm="'sys:chat-merchant:add'" type="primary" @click="handleCreate">
          {{ t('common.add') }}
        </el-button>
      </template>
    </CrudQueryCard>

    <el-card class="table-card" shadow="never">
      <el-table v-loading="loading" :data="list" :empty-text="t('common.noData')" stripe>
        <el-table-column prop="id" :label="t('common.id')" width="80" align="center" />
        <el-table-column
          prop="merchantCode"
          :label="t('system.chatMerchantCode')"
          min-width="150"
        />
        <el-table-column
          prop="merchantName"
          :label="t('system.chatMerchantName')"
          min-width="160"
        />
        <el-table-column prop="contactName" :label="t('system.contactName')" min-width="120" />
        <el-table-column prop="contactPhone" :label="t('system.contactPhone')" min-width="130" />
        <el-table-column prop="contactEmail" :label="t('system.contactEmail')" min-width="170" />
        <el-table-column :label="t('system.enabled')" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.enabled === 1 ? 'success' : 'info'">
              {{ enabledLabel(row.enabled) }}
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
        <el-table-column :label="t('common.actions')" width="220" align="center" fixed="right">
          <template #default="{ row }">
            <el-button type="info" size="small" @click="handleDetail(row)">
              {{ t('common.detail') }}
            </el-button>
            <el-button
              v-perm="'sys:chat-merchant:update'"
              type="primary"
              size="small"
              @click="handleEdit(row)"
            >
              {{ t('common.edit') }}
            </el-button>
            <el-button
              v-perm="'sys:chat-merchant:delete'"
              type="danger"
              size="small"
              @click="handleDelete(row)"
            >
              {{ t('common.delete') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <CursorPagination
        v-model:limit="pagination.limit"
        :total="pagination.total"
        :has-prev="pagination.hasPrev"
        :has-next="pagination.hasNext"
        @prev="prevPage"
        @next="nextPage"
        @limit-change="handleSizeChange"
      />
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? t('system.editChatMerchant') : t('system.addChatMerchant')"
      width="720px"
      :close-on-click-modal="false"
    >
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="130px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="t('system.chatMerchantCode')" prop="merchantCode">
              <el-input
                v-model="formData.merchantCode"
                :placeholder="t('system.pleaseInputChatMerchantCode')"
                maxlength="64"
                show-word-limit
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('system.chatMerchantName')" prop="merchantName">
              <el-input
                v-model="formData.merchantName"
                :placeholder="t('system.pleaseInputChatMerchantName')"
                maxlength="128"
                show-word-limit
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('common.password')" prop="password">
              <el-input
                v-model="formData.password"
                type="password"
                show-password
                :placeholder="
                  isEdit
                    ? t('system.chatMerchantPasswordKeepPlaceholder')
                    : t('common.pleaseInputNewPassword')
                "
                maxlength="64"
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
                maxlength="64"
                show-word-limit
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('system.contactPhone')" prop="contactPhone">
              <el-input
                v-model="formData.contactPhone"
                :placeholder="t('system.pleaseInputContactPhone')"
                maxlength="32"
                show-word-limit
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="t('system.contactEmail')" prop="contactEmail">
              <el-input
                v-model="formData.contactEmail"
                :placeholder="t('system.pleaseInputContactEmail')"
                maxlength="128"
                show-word-limit
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('system.enabled')" prop="enabled">
              <el-select v-model="formData.enabled" style="width: 100%">
                <el-option
                  v-for="item in enabledFormOptions"
                  :key="item.value"
                  :label="enabledOptionLabel(item.value, item.code)"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
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
            maxlength="255"
            show-word-limit
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button
          v-perm="isEdit ? 'sys:chat-merchant:update' : 'sys:chat-merchant:add'"
          type="primary"
          :loading="submitLoading"
          @click="handleSubmit"
        >
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="detailVisible" :title="t('system.chatMerchantDetail')" size="720px">
      <el-descriptions v-if="detailData" :column="2" border>
        <el-descriptions-item :label="t('common.id')">
          {{ detailData.id }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('system.enabled')">
          <el-tag :type="detailData.enabled === 1 ? 'success' : 'info'">
            {{ enabledLabel(detailData.enabled) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="t('system.chatMerchantCode')">
          {{ detailData.merchantCode }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('system.chatMerchantName')">
          {{ detailData.merchantName }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('system.contactName')">
          {{ detailData.contactName || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('system.contactPhone')">
          {{ detailData.contactPhone || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('system.contactEmail')">
          {{ detailData.contactEmail || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('system.expireTime')">
          {{ formatDate(detailData.expireTime) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.createTimes')">
          {{ formatDate(detailData.createTimes) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.updateTimes')">
          {{ formatDate(detailData.updateTimes) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.remark')" :span="2">
          {{ detailData.remark || '-' }}
        </el-descriptions-item>
      </el-descriptions>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { chatMerchantsService } from '@/services/system/ChatMerchantsService'
import type { OptionGroup } from '@/services'
import type {
  SysChatMerchantCreateReq,
  SysChatMerchantItem,
} from '@/services/system/ChatMerchantsService'
import { useForm } from '@/composables/useForm'
import { useLoading } from '@/composables/useLoading'
import { usePagination } from '@/composables/usePagination'
import { formatDate } from '@/utils'
import {
  findFormOptionGroup,
  findOptionGroup,
  getOptionLabel,
  getOptionValueLabel,
} from '@/utils/options'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()
const optionGroups = ref<OptionGroup[]>([])
const enabledOptions = computed(() => findOptionGroup(optionGroups.value, 'enabled'))
const enabledSelectOptions = computed(() => {
  const options = enabledOptions.value
  return options.length
    ? options
    : [
        { value: 0, code: 'COMMON_STATUS_UNKNOWN' },
        { value: 1, code: 'COMMON_STATUS_ENABLED' },
        { value: 2, code: 'COMMON_STATUS_DISABLED' },
      ]
})
const enabledFormOptions = computed(() => {
  const options = findFormOptionGroup(optionGroups.value, 'enabled')
  return options.length
    ? options
    : [
        { value: 1, code: 'COMMON_STATUS_ENABLED' },
        { value: 2, code: 'COMMON_STATUS_DISABLED' },
      ]
})

const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)
const list = ref<SysChatMerchantItem[]>([])
const { loading, withLoading } = useLoading()

const { form: queryForm } = useForm({
  initialData: {
    merchantCode: '',
    merchantName: '',
    contactName: '',
    enabled: 0,
  },
})

const dialogVisible = ref(false)
const detailVisible = ref(false)
const detailData = ref<SysChatMerchantItem>()
const isEdit = ref(false)
const submitLoading = ref(false)
const formRef = ref()

const { form: formData, reset: resetForm } = useForm({
  initialData: {
    id: 0,
    merchantCode: '',
    merchantName: '',
    enabled: 1,
    expireTime: Date.now() + 365 * 24 * 60 * 60 * 1000,
    contactName: '',
    contactPhone: '',
    contactEmail: '',
    password: '',
    remark: '',
  },
})

const formRules = {
  merchantCode: [
    { required: true, message: t('system.pleaseInputChatMerchantCode'), trigger: 'blur' },
  ],
  merchantName: [
    { required: true, message: t('system.pleaseInputChatMerchantName'), trigger: 'blur' },
  ],
  password: [
    {
      validator: (_rule: unknown, value: string, callback: (error?: Error) => void) => {
        if (!isEdit.value && !value) {
          callback(new Error(t('common.pleaseInputNewPassword')))
          return
        }
        callback()
      },
      trigger: 'blur',
    },
  ],
  enabled: [{ required: true, message: t('system.pleaseSelectStatus'), trigger: 'change' }],
  expireTime: [{ required: true, message: t('validation.required'), trigger: 'change' }],
}

function enabledOptionLabel(value: number | string | undefined, code?: string) {
  if (Number(value) === 1) return t('common.enabled')
  if (Number(value) === 2) return t('common.disabled')
  if (Number(value) === 0) return t('common.all')
  return getOptionLabel(t, code, Number(value))
}

function enabledLabel(value: number | undefined) {
  const label = getOptionValueLabel(optionGroups.value, 'enabled', value, t)
  if (label && label !== value) return label
  return enabledOptionLabel(value)
}

async function loadList() {
  await withLoading(async () => {
    try {
      const res = await chatMerchantsService.getList({
        merchantCode: queryForm.merchantCode || undefined,
        merchantName: queryForm.merchantName || undefined,
        contactName: queryForm.contactName || undefined,
        enabled: queryForm.enabled === 0 ? undefined : queryForm.enabled,
        cursor: pagination.cursor,
        limit: pagination.limit,
      })
      if (res.code !== 200) throw new Error(res.msg || 'list failed')
      list.value = res.data || []
      updateFromResponse(res)
    } catch (error: unknown) {
      ElMessage.error(error instanceof Error ? error.message : t('common.loadFailed'))
    }
  })
}

async function fetchOptions() {
  try {
    const res = await chatMerchantsService.getOptions()
    if (res.code !== 200) throw new Error(res.msg || 'options failed')
    optionGroups.value = res.data || []
  } catch (error: unknown) {
    ElMessage.error(error instanceof Error ? error.message : t('common.loadFailed'))
  }
}

function handleSizeChange(size: number) {
  pagination.limit = size
  resetAndLoad(loadList)
}

function resetQuery() {
  queryForm.merchantCode = ''
  queryForm.merchantName = ''
  queryForm.contactName = ''
  queryForm.enabled = 0
  resetAndLoad(loadList)
}

function nextPage() {
  nextAndLoad(loadList)
}

function prevPage() {
  prevAndLoad(loadList)
}

function handleCreate() {
  isEdit.value = false
  resetForm()
  dialogVisible.value = true
}

function handleEdit(row: SysChatMerchantItem) {
  isEdit.value = true
  resetForm()
  Object.assign(formData, {
    id: row.id,
    merchantCode: row.merchantCode,
    merchantName: row.merchantName,
    enabled: row.enabled,
    expireTime: row.expireTime,
    contactName: row.contactName,
    contactPhone: row.contactPhone,
    contactEmail: row.contactEmail,
    password: '',
    remark: row.remark || '',
  })
  dialogVisible.value = true
}

async function handleDetail(row: SysChatMerchantItem) {
  try {
    const res = await chatMerchantsService.detail({ id: row.id })
    if (res.code !== 200) throw new Error(res.msg || t('common.loadFailed'))
    detailData.value = res.data || row
    detailVisible.value = true
  } catch (error: unknown) {
    ElMessage.error(error instanceof Error ? error.message : t('common.loadFailed'))
  }
}

async function handleDelete(row: SysChatMerchantItem) {
  try {
    await ElMessageBox.confirm(t('common.confirmDelete'), t('common.warning'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning',
    })

    const res = await chatMerchantsService.delete(row.id)
    if (res.code !== 200) throw new Error(res.msg || 'delete failed')

    ElMessage.success(t('common.deleteSuccess'))
    loadList()
  } catch (error: unknown) {
    if ((error instanceof Error ? error.message : '') !== 'cancel') {
      ElMessage.error(error instanceof Error ? error.message : t('common.deleteFailed'))
    }
  }
}

async function handleSubmit() {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    submitLoading.value = true

    if (isEdit.value) {
      const res = await chatMerchantsService.update(formData.id, {
        merchantCode: formData.merchantCode,
        merchantName: formData.merchantName,
        enabled: formData.enabled,
        expireTime: formData.expireTime,
        contactName: formData.contactName,
        contactPhone: formData.contactPhone,
        contactEmail: formData.contactEmail,
        password: formData.password || undefined,
        remark: formData.remark || '',
      })
      if (res.code !== 200) throw new Error(res.msg || t('common.updateFailed'))
      ElMessage.success(t('common.updateSuccess'))
    } else {
      const data: SysChatMerchantCreateReq = {
        merchantCode: formData.merchantCode,
        merchantName: formData.merchantName,
        enabled: formData.enabled,
        expireTime: formData.expireTime,
        contactName: formData.contactName,
        contactPhone: formData.contactPhone,
        contactEmail: formData.contactEmail,
        password: formData.password,
        remark: formData.remark || '',
      }
      const res = await chatMerchantsService.create(data)
      if (res.code !== 200) throw new Error(res.msg || t('common.createFailed'))
      ElMessage.success(t('common.createSuccess'))
    }

    dialogVisible.value = false
    loadList()
  } catch (error: unknown) {
    ElMessage.error(error instanceof Error ? error.message : t('common.operationFailed'))
  } finally {
    submitLoading.value = false
  }
}

onMounted(() => {
  fetchOptions()
  loadList()
})
</script>
