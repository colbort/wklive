<template>
  <div class="module-page">
    <CrudQueryCard :model="queryForm" @search="loadList" @reset="resetQuery">
      <el-form-item :label="t('system.tenantName')" required>
        <TenantSelect v-model="queryForm.tenantId" @change="handleTenantChange" />
      </el-form-item>
      <template #actions>
        <el-button v-perm="'sys:tenant-domain:add'" type="primary" @click="openCreate">
          {{ t('common.add') }}
        </el-button>
      </template>
    </CrudQueryCard>

    <el-card shadow="never">
      <el-alert
        :title="t('system.tenantDomainTip')"
        type="info"
        :closable="false"
        show-icon
        class="domain-tip"
      />
      <el-table
        v-loading="loading"
        :data="list"
        :empty-text="t('common.noData')"
        stripe
      >
        <el-table-column
          prop="id"
          :label="t('common.id')"
          width="80"
          align="center"
        />
        <el-table-column prop="origin" :label="t('system.domainOrigin')" min-width="260" />
        <el-table-column :label="t('system.status')" width="120" align="center">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">
              {{ optionLabel('tenantDomainStatus', row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="priority"
          :label="t('system.domainPriority')"
          width="110"
          align="center"
        />
        <el-table-column
          v-if="canViewMigrationStats"
          :label="t('system.notMigratedGuests')"
          align="center"
        >
          <el-table-column :label="t('system.notMigratedTotal')" width="100" align="center">
            <template #default="{ row }">
              {{ migrationStat(row, 'notMigratedCount') }}
            </template>
          </el-table-column>
          <el-table-column :label="t('system.activeLast7Days')" width="90" align="center">
            <template #default="{ row }">
              {{ migrationStat(row, 'activeLast7DaysCount') }}
            </template>
          </el-table-column>
          <el-table-column :label="t('system.active8To30Days')" width="100" align="center">
            <template #default="{ row }">
              {{ migrationStat(row, 'active8To30DaysCount') }}
            </template>
          </el-table-column>
          <el-table-column :label="t('system.active31To90Days')" width="110" align="center">
            <template #default="{ row }">
              {{ migrationStat(row, 'active31To90DaysCount') }}
            </template>
          </el-table-column>
          <el-table-column :label="t('system.inactiveOver90Days')" width="110" align="center">
            <template #default="{ row }">
              {{ migrationStat(row, 'inactiveOver90DaysCount') }}
            </template>
          </el-table-column>
        </el-table-column>
        <el-table-column :label="t('common.updateTimes')" width="180" align="center">
          <template #default="{ row }">
            {{ formatDate(row.updateTimes) }}
          </template>
        </el-table-column>
        <el-table-column
          :label="t('common.actions')"
          width="170"
          align="center"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              v-perm="'sys:tenant-domain:update'"
              type="primary"
              size="small"
              @click="openEdit(row)"
            >
              {{ t('common.edit') }}
            </el-button>
            <el-button
              v-perm="'sys:tenant-domain:delete'"
              type="danger"
              size="small"
              @click="removeDomain(row)"
            >
              {{ t('common.delete') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? t('system.editTenantDomain') : t('system.addTenantDomain')"
      width="560px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="110px"
      >
        <el-form-item :label="t('system.domainOrigin')" prop="origin">
          <el-input v-model="form.origin" placeholder="https://example.com" clearable />
        </el-form-item>
        <el-form-item :label="t('system.status')" prop="status">
          <el-select v-model="form.status" style="width: 100%">
            <el-option
              v-for="item in domainStatusOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('system.domainPriority')" prop="priority">
          <el-input-number
            v-model="form.priority"
            :min="0"
            :max="2147483647"
            style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button
          v-perm="isEdit ? 'sys:tenant-domain:update' : 'sys:tenant-domain:add'"
          type="primary"
          :loading="submitting"
          @click="submit"
        >
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import type { OptionGroup } from '@/services'
import {
  tenantDomainsService,
  type SysTenantDomainCreateReq,
  type SysTenantDomainGuestMigrationStats,
  type SysTenantDomainItem,
} from '@/services'
import { useLoading } from '@/composables/useLoading'
import { useOptions } from '@/composables/useOptions'
import { formatDate } from '@/utils'
import { useAuthStore } from '@/stores/auth'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'
import TenantSelect from '@/components/TenantSelect.vue'

const { t } = useI18n()
const auth = useAuthStore()
const canViewMigrationStats = computed(() =>
  auth.hasPerm('sys:tenant-domain:guest-migration-stats'),
)
const optionGroups = ref<OptionGroup[]>([])
const { formOptionItems, optionLabel } = useOptions(optionGroups)
const domainStatusOptions = formOptionItems('tenantDomainStatus')
const { loading, withLoading } = useLoading()
const list = ref<SysTenantDomainItem[]>([])
const queryForm = reactive<{ tenantId?: number }>({ tenantId: undefined })
const dialogVisible = ref(false)
const submitting = ref(false)
const formRef = ref()
const form = reactive({ id: 0, origin: '', status: 1, priority: 0 })
const isEdit = computed(() => form.id > 0)

const rules = {
  origin: [{ required: true, message: t('system.pleaseInputDomainOrigin'), trigger: 'blur' }],
  status: [{ required: true, message: t('system.pleaseSelectStatus'), trigger: 'change' }],
}

function statusTagType(status: number) {
  if (status === 1) return 'success'
  if (status === 2) return 'warning'
  return 'info'
}

function migrationStat(row: SysTenantDomainItem, field: keyof SysTenantDomainGuestMigrationStats) {
  if (row.status !== 2) return '--'
  return row.migrationStats?.[field] ?? '--'
}

async function loadList() {
  if (!queryForm.tenantId) {
    list.value = []
    return
  }
  await withLoading(async () => {
    try {
      const res = await tenantDomainsService.getList({ tenantId: queryForm.tenantId! })
      if (res.code !== 200) throw new Error(res.msg)
      const domains = res.data || []
      await Promise.all(
        domains
          .filter((domain) => canViewMigrationStats.value && domain.status === 2)
          .map(async (domain) => {
            try {
              const statsRes = await tenantDomainsService.getGuestMigrationStats(
                queryForm.tenantId!,
                domain.origin,
              )
              if (statsRes.code === 200) domain.migrationStats = statsRes.data
            } catch {
              // 单个域名统计失败不影响域名列表展示
            }
          }),
      )
      list.value = domains
    } catch (error) {
      ElMessage.error(error instanceof Error ? error.message : t('common.loadFailed'))
    }
  })
}

function handleTenantChange() {
  loadList()
}

function resetQuery() {
  queryForm.tenantId = undefined
  list.value = []
}

function resetForm() {
  Object.assign(form, { id: 0, origin: '', status: 1, priority: 0 })
}

function openCreate() {
  if (!queryForm.tenantId) {
    ElMessage.warning(t('system.pleaseSelectTenant'))
    return
  }
  resetForm()
  dialogVisible.value = true
}

function openEdit(row: SysTenantDomainItem) {
  Object.assign(form, row)
  dialogVisible.value = true
}

async function submit() {
  if (!formRef.value || !queryForm.tenantId) return
  try {
    await formRef.value.validate()
    submitting.value = true
    const payload = { origin: form.origin, status: form.status, priority: form.priority }
    const res = isEdit.value
      ? await tenantDomainsService.update(form.id, payload)
      : await tenantDomainsService.create({
          tenantId: queryForm.tenantId,
          ...payload,
        } as SysTenantDomainCreateReq)
    if (res.code !== 200) throw new Error(res.msg)
    ElMessage.success(isEdit.value ? t('common.updateSuccess') : t('common.createSuccess'))
    dialogVisible.value = false
    await loadList()
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t('common.operationFailed'))
  } finally {
    submitting.value = false
  }
}

async function removeDomain(row: SysTenantDomainItem) {
  try {
    await ElMessageBox.confirm(t('common.confirmDelete'), t('common.warning'), { type: 'warning' })
    const res = await tenantDomainsService.delete(row.id)
    if (res.code !== 200) throw new Error(res.msg)
    ElMessage.success(t('common.deleteSuccess'))
    await loadList()
  } catch (error) {
    if (error !== 'cancel')
      ElMessage.error(error instanceof Error ? error.message : t('common.deleteFailed'))
  }
}

onMounted(async () => {
  const res = await tenantDomainsService.getOptions()
  optionGroups.value = res.data || []
})
</script>

<style scoped>
.domain-tip {
  margin-bottom: 16px;
}
</style>
