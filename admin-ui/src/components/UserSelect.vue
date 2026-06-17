<template>
  <el-popover
    v-model:visible="visible"
    placement="bottom-start"
    trigger="click"
    width="460"
    popper-class="user-select-popover"
    @show="handleShow"
  >
    <template #reference>
      <el-input
        :model-value="displayValue"
        :placeholder="placeholder || t('common.pleaseSelect')"
        :disabled="disabled"
        readonly
        clearable
        @clear.stop="clearValue"
      />
    </template>

    <div class="user-select-panel">
      <el-form inline class="user-select-filter" @submit.prevent>
        <el-form-item :label="t('users.userId')">
          <el-input-number
            v-model="queryUserId"
            :min="0"
            :precision="0"
            controls-position="right"
            @keyup.enter="searchUsers"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="searchUsers">
            {{ t('common.search') }}
          </el-button>
          <el-button @click="resetQuery">
            {{ t('common.reset') }}
          </el-button>
        </el-form-item>
      </el-form>

      <el-table
        v-loading="loading"
        :data="users"
        height="280"
        highlight-current-row
        @row-click="selectUser"
      >
        <el-table-column prop="id" :label="t('users.userId')" width="90" />
        <el-table-column prop="nickname" :label="t('users.nickname')" min-width="180" />
        <el-table-column :label="t('users.isGuest')" width="100">
          <template #default="{ row }">
            <span :class="getGuestTagClass(row.isGuest)">
              {{ getGuestLabel(row.isGuest) }}
            </span>
          </template>
        </el-table-column>
      </el-table>

      <CursorPagination
        v-model:limit="pagination.limit"
        :total="pagination.total"
        :has-prev="pagination.hasPrev"
        :has-next="pagination.hasNext"
        :disabled="loading"
        :select-teleported="false"
        @prev="handlePrevPage"
        @next="handleNextPage"
        @limit-change="handleLimitChange"
      />
    </div>
  </el-popover>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import CursorPagination from '@/components/common/CursorPagination.vue'
import { usePagination } from '@/composables/usePagination'
import { memberUserService, type MemberUserItem } from '@/services'

const props = withDefaults(
  defineProps<{
    modelValue?: number
    disabled?: boolean
    placeholder?: string
    tenantId?: number
  }>(),
  {
    modelValue: undefined,
    disabled: false,
    placeholder: '',
    tenantId: undefined,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: number | undefined]
  change: [value: number | undefined]
  selected: [value: MemberUserItem | null]
}>()

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

const visible = ref(false)
const loading = ref(false)
const users = ref<MemberUserItem[]>([])
const queryUserId = ref<number | undefined>(undefined)

const displayValue = computed(() => (props.modelValue ? String(props.modelValue) : ''))

watch(
  () => props.modelValue,
  (value) => {
    queryUserId.value = value
  },
)

async function loadUsers() {
  loading.value = true
  try {
    const res = await memberUserService.getList({
      cursor: pagination.cursor,
      limit: pagination.limit,
      tenantId: props.tenantId || undefined,
      userId: queryUserId.value || undefined,
    })
    users.value = res.data || []
    updateFromResponse(res)
  } finally {
    loading.value = false
  }
}

function handleShow() {
  queryUserId.value = props.modelValue
  resetAndLoad(loadUsers)
}

function searchUsers() {
  resetAndLoad(loadUsers)
}

function resetQuery() {
  queryUserId.value = undefined
  resetAndLoad(loadUsers)
}

function handleLimitChange() {
  resetAndLoad(loadUsers)
}

function handlePrevPage() {
  prevAndLoad(loadUsers)
}

function handleNextPage() {
  nextAndLoad(loadUsers)
}

function selectUser(row: MemberUserItem) {
  emit('update:modelValue', row.id)
  emit('change', row.id)
  emit('selected', row)
  visible.value = false
}

function clearValue() {
  emit('update:modelValue', undefined)
  emit('change', undefined)
  emit('selected', null)
}

function getGuestLabel(value?: number) {
  return Number(value) === 2 ? t('users.yes') : t('users.no')
}

function getGuestTagClass(value?: number) {
  return Number(value) === 2 ? 'option-tag option-tag--green' : 'option-tag option-tag--red'
}
</script>

<style scoped>
.user-select-panel {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.user-select-filter {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.user-select-filter :deep(.el-form-item) {
  margin: 0;
}

.user-select-filter :deep(.el-input-number) {
  width: 180px;
}

:global(.user-select-popover .pagination-bar) {
  justify-content: flex-start;
  gap: 8px;
  margin-top: 0;
  padding-top: 10px;
  border-top: 1px solid var(--el-border-color-lighter);
}

:global(.user-select-popover .pagination-bar .el-button) {
  min-width: 72px;
  padding: 8px 14px;
}

:global(.user-select-popover .pagination-bar .el-select) {
  width: 96px !important;
}

.option-tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 42px;
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
  line-height: 20px;
}

.option-tag--green {
  color: #059669;
  background: #dcfce7;
}

.option-tag--red {
  color: #dc2626;
  background: #fee2e2;
}
</style>
