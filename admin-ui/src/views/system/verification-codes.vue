<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Message, View } from '@element-plus/icons-vue'
import CursorPagination from '@/components/common/CursorPagination.vue'
import TenantSelect from '@/components/TenantSelect.vue'
import { useLoading } from '@/composables/useLoading'
import { useOptions } from '@/composables/useOptions'
import { usePagination } from '@/composables/usePagination'
import { getCoreOptions } from '@/stores/core'
import { formatDate } from '@/utils'
import { verificationCodeService } from '@/services'
import type { OptionGroup, VerificationCodeRecordItem } from '@/services'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()
const { loading, withLoading } = useLoading()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

const optionGroups = ref<OptionGroup[]>([])
const { optionItems, optionLabel } = useOptions(optionGroups)

const channelOptions = optionItems('verificationCodeChannel')
const sceneOptions = optionItems('verificationCodeScene')
const statusOptions = optionItems('verificationCodeStatus')
const effectiveChannelOptions = computed(() =>
  channelOptions.value.filter((item) => item.value > 0),
)
const effectiveSceneOptions = computed(() => sceneOptions.value.filter((item) => item.value > 0))

const queryForm = reactive({
  tenantId: undefined as number | undefined,
  channel: undefined as number | undefined,
  target: '',
  scene: undefined as number | undefined,
  status: undefined as number | undefined,
})

const records = ref<VerificationCodeRecordItem[]>([])
const detailVisible = ref(false)
const detailLoading = ref(false)
const detailData = ref<VerificationCodeRecordItem | null>(null)

const testVisible = ref(false)
const testLoading = ref(false)
const testFormRef = ref<FormInstance>()
const testForm = reactive({
  tenantId: 0,
  channel: 1,
  email: '',
  phone: '',
  scene: 100,
})

const testRules: FormRules = {
  tenantId: [{ required: true, message: t('system.pleaseSelectTenant'), trigger: 'change' }],
  channel: [{ required: true, message: t('system.pleaseSelectChannel'), trigger: 'change' }],
  scene: [{ required: true, message: t('system.pleaseSelectScene'), trigger: 'change' }],
}

function channelTagType(channel: number) {
  return channel === 1 ? 'primary' : 'success'
}

function statusTagType(status: number) {
  if (status === 1) return 'success'
  if (status === 2) return 'danger'
  return 'info'
}

async function loadOptions() {
  try {
    const res = await getCoreOptions()
    if (res.code === 200) {
      optionGroups.value = res.data || []
    }
  } catch (error) {
    console.warn('load system options failed', error)
  }
}

async function fetchList() {
  await withLoading(async () => {
    try {
      const res = await verificationCodeService.getList({
        tenantId: queryForm.tenantId,
        channel: queryForm.channel === 0 ? undefined : queryForm.channel,
        target: queryForm.target || undefined,
        scene: queryForm.scene === 0 ? undefined : queryForm.scene,
        status: queryForm.status === 0 ? undefined : queryForm.status,
        cursor: pagination.cursor,
        limit: pagination.limit,
      })
      if (res.code !== 200) throw new Error(res.msg)
      records.value = res.data || []
      updateFromResponse(res)
    } catch (error) {
      ElMessage.error(error instanceof Error ? error.message : t('common.loadFailed'))
    }
  })
}

function loadList() {
  resetAndLoad(fetchList)
}

function resetQuery() {
  queryForm.tenantId = undefined
  queryForm.channel = undefined
  queryForm.target = ''
  queryForm.scene = undefined
  queryForm.status = undefined
  resetAndLoad(fetchList)
}

function nextPage() {
  nextAndLoad(fetchList)
}

function prevPage() {
  prevAndLoad(fetchList)
}

async function showDetail(row: VerificationCodeRecordItem) {
  detailVisible.value = true
  detailLoading.value = true
  detailData.value = row
  try {
    const res = await verificationCodeService.getDetail(row.id)
    if (res.code !== 200) throw new Error(res.msg)
    detailData.value = res.data || row
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t('common.loadFailed'))
  } finally {
    detailLoading.value = false
  }
}

function openTestDialog() {
  testForm.tenantId = 0
  testForm.channel = 1
  testForm.email = ''
  testForm.phone = ''
  testForm.scene = 100
  testVisible.value = true
}

async function submitTest() {
  if (!testFormRef.value) return
  const valid = await testFormRef.value.validate().catch(() => false)
  if (!valid) return

  if (testForm.channel === 1 && !testForm.email.trim()) {
    ElMessage.warning(t('system.pleaseInputEmail'))
    return
  }
  if (testForm.channel === 2 && !testForm.phone.trim()) {
    ElMessage.warning(t('system.pleaseInputPhone'))
    return
  }

  testLoading.value = true
  try {
    const res = await verificationCodeService.testSend({
      tenantId: testForm.tenantId,
      channel: testForm.channel,
      email: testForm.channel === 1 ? testForm.email.trim() : undefined,
      phone: testForm.channel === 2 ? testForm.phone.trim() : undefined,
      scene: testForm.scene,
    })
    if (res.code !== 200) throw new Error(res.msg)
    ElMessage.success(t('system.verificationCodeSent'))
    testVisible.value = false
    resetAndLoad(fetchList)
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t('system.verificationCodeSendFailed'))
  } finally {
    testLoading.value = false
  }
}

onMounted(() => {
  loadOptions()
  fetchList()
})
</script>

<template>
  <div class="module-page verification-code-page">
    <CrudQueryCard :model="queryForm" @search="loadList" @reset="resetQuery">
      <el-form-item :label="t('common.tenantId')">
        <TenantSelect v-model="queryForm.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('system.channel')">
        <el-select v-model="queryForm.channel" clearable style="width: 150px">
          <el-option
            v-for="item in channelOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('system.scene')">
        <el-select v-model="queryForm.scene" clearable style="width: 170px">
          <el-option
            v-for="item in sceneOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('common.status')">
        <el-select v-model="queryForm.status" clearable style="width: 150px">
          <el-option
            v-for="item in statusOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('system.target')">
        <el-input
          v-model="queryForm.target"
          :placeholder="t('system.emailOrPhone')"
          clearable
          style="width: 220px"
        />
      </el-form-item>
      <template #actions>
        <el-button v-perm="'sys:verification-code:test'" type="primary" @click="openTestDialog">
          <el-icon><Message /></el-icon>
          {{ t('system.testSendVerificationCode') }}
        </el-button>
      </template>
    </CrudQueryCard>

    <el-card class="table-card" shadow="never">
      <el-table
        v-loading="loading"
        :data="records"
        row-key="id"
        stripe
      >
        <el-table-column
          prop="id"
          :label="t('common.id')"
          width="80"
          align="center"
        />
        <el-table-column
          prop="tenantId"
          :label="t('common.tenantId')"
          width="100"
          align="center"
        />
        <el-table-column :label="t('system.channel')" width="120" align="center">
          <template #default="{ row }">
            <el-tag :type="channelTagType(row.channel)">
              {{ optionLabel('verificationCodeChannel', row.channel) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="target" :label="t('system.target')" min-width="170" />
        <el-table-column :label="t('system.scene')" width="150" align="center">
          <template #default="{ row }">
            {{ optionLabel('verificationCodeScene', row.scene) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="code"
          :label="t('system.verificationCode')"
          width="110"
          align="center"
        />
        <el-table-column :label="t('common.status')" width="120" align="center">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">
              {{ optionLabel('verificationCodeStatus', row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="provider" :label="t('system.provider')" width="120" />
        <el-table-column
          prop="errorMessage"
          :label="t('system.errorMessage')"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column prop="createTimes" :label="t('common.createTimes')" width="170">
          <template #default="{ row }">
            {{ formatDate(row.createTimes) }}
          </template>
        </el-table-column>
        <el-table-column
          :label="t('common.actions')"
          width="110"
          align="center"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              v-perm="'sys:verification-code:detail'"
              type="primary"
              size="small"
              @click="showDetail(row)"
            >
              <el-icon><View /></el-icon>
              {{ t('common.detail') }}
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
        @limit-change="
          () => {
            resetAndLoad(fetchList)
          }
        "
      />
    </el-card>

    <el-drawer v-model="detailVisible" :title="t('system.verificationCodeDetail')" size="520px">
      <el-skeleton v-if="detailLoading" :rows="8" animated />
      <el-descriptions v-else-if="detailData" :column="1" border>
        <el-descriptions-item :label="t('common.id')">
          {{ detailData.id }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.tenantId')">
          {{ detailData.tenantId }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('system.channel')">
          {{ optionLabel('verificationCodeChannel', detailData.channel) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('system.target')">
          {{ detailData.target }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('system.scene')">
          {{ optionLabel('verificationCodeScene', detailData.scene) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('system.verificationCode')">
          {{ detailData.code }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.status')">
          {{ optionLabel('verificationCodeStatus', detailData.status) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('system.provider')">
          {{ detailData.provider || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('system.errorMessage')">
          <span class="break-text">{{ detailData.errorMessage || '-' }}</span>
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.createTimes')">
          {{ formatDate(detailData.createTimes) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.updateTimes')">
          {{ formatDate(detailData.updateTimes) }}
        </el-descriptions-item>
      </el-descriptions>
    </el-drawer>

    <el-dialog
      v-model="testVisible"
      :title="t('system.testSendVerificationCode')"
      width="520px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="testFormRef"
        :model="testForm"
        :rules="testRules"
        label-width="120px"
      >
        <el-form-item :label="t('common.tenantId')" prop="tenantId">
          <TenantSelect v-model="testForm.tenantId" include-system />
        </el-form-item>
        <el-form-item :label="t('system.channel')" prop="channel">
          <el-select v-model="testForm.channel" style="width: 100%">
            <el-option
              v-for="item in effectiveChannelOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item v-if="testForm.channel === 1" :label="t('system.email')" prop="email">
          <el-input v-model="testForm.email" :placeholder="t('system.pleaseInputEmail')" />
        </el-form-item>
        <el-form-item v-else :label="t('system.phone')" prop="phone">
          <el-input v-model="testForm.phone" :placeholder="t('system.pleaseInputPhone')" />
        </el-form-item>
        <el-form-item :label="t('system.scene')" prop="scene">
          <el-select v-model="testForm.scene" style="width: 100%">
            <el-option
              v-for="item in effectiveSceneOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="testVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button
          v-perm="'sys:verification-code:test'"
          type="primary"
          :loading="testLoading"
          @click="submitTest"
        >
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>
