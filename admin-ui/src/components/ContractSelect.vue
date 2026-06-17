<template>
  <el-popover
    v-model:visible="visible"
    placement="bottom-start"
    trigger="click"
    width="700"
    popper-class="contract-select-popover"
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

    <div class="contract-select-panel">
      <el-form inline class="contract-select-filter" @submit.prevent>
        <el-form-item :label="t('option.contractCode')">
          <el-input
            v-model="queryContractCode"
            clearable
            style="width: 200px"
            @keyup.enter="searchContracts"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="searchContracts">
            {{ t('common.search') }}
          </el-button>
          <el-button @click="resetQuery">
            {{ t('common.reset') }}
          </el-button>
        </el-form-item>
      </el-form>

      <el-table
        v-loading="loading"
        :data="contracts"
        height="320"
        highlight-current-row
        @row-click="selectContract"
      >
        <el-table-column prop="id" :label="t('option.contractId')" width="90" />
        <el-table-column prop="contractCode" :label="t('option.contractCode')" min-width="180" />
        <el-table-column prop="underlyingSymbol" :label="t('option.underlying')" min-width="130" />
        <el-table-column prop="strikePrice" :label="t('option.strikePrice')" min-width="110" />
        <el-table-column prop="tenantId" :label="t('option.tenantId')" width="100" />
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
import { optionService, type OptionContract } from '@/services'

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
  selected: [value: OptionContract | null]
}>()

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

const visible = ref(false)
const loading = ref(false)
const contracts = ref<OptionContract[]>([])
const selectedContract = ref<OptionContract | null>(null)
const queryContractId = ref<number | undefined>(undefined)
const queryContractCode = ref('')

const displayValue = computed(() => {
  if (!props.modelValue) return ''
  const contract = selectedContract.value
  if (contract && contract.id === props.modelValue) {
    return `${contract.id} - ${contract.contractCode}`
  }
  return String(props.modelValue)
})

watch(
  () => props.modelValue,
  (value) => {
    queryContractId.value = value
    if (!value || selectedContract.value?.id !== value) {
      selectedContract.value = null
    }
  },
)

watch(
  () => props.tenantId,
  () => {
    contracts.value = []
  },
)

async function loadContracts() {
  loading.value = true
  try {
    if (queryContractId.value) {
      const detail = await optionService.getContract({
        tenantId: props.tenantId || undefined,
        id: queryContractId.value,
      })
      const contract = detail.data?.contract
      contracts.value = contract ? [contract] : []
      selectedContract.value =
        contract && contract.id === props.modelValue ? contract : selectedContract.value
      updateFromResponse({ total: contracts.value.length })
      return
    }

    const res = await optionService.listContracts({
      cursor: pagination.cursor,
      limit: pagination.limit,
      tenantId: props.tenantId || undefined,
      contractCode: queryContractCode.value || undefined,
    })
    contracts.value = (res.data || []).map((item) => item.contract).filter(Boolean)
    updateFromResponse(res)
    syncSelectedFromList()
  } finally {
    loading.value = false
  }
}

function syncSelectedFromList() {
  if (!props.modelValue) return
  const current = contracts.value.find((item) => item.id === props.modelValue)
  if (current) selectedContract.value = current
}

function handleShow() {
  queryContractId.value = props.modelValue
  resetAndLoad(loadContracts)
}

function searchContracts() {
  resetAndLoad(loadContracts)
}

function resetQuery() {
  queryContractId.value = undefined
  queryContractCode.value = ''
  resetAndLoad(loadContracts)
}

function handleLimitChange() {
  resetAndLoad(loadContracts)
}

function handlePrevPage() {
  prevAndLoad(loadContracts)
}

function handleNextPage() {
  nextAndLoad(loadContracts)
}

function selectContract(row: OptionContract) {
  selectedContract.value = row
  emit('update:modelValue', row.id)
  emit('change', row.id)
  emit('selected', row)
  visible.value = false
}

function clearValue() {
  selectedContract.value = null
  emit('update:modelValue', undefined)
  emit('change', undefined)
  emit('selected', null)
}
</script>

<style scoped>
.contract-select-panel {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.contract-select-filter {
  display: flex;
  align-items: center;
  gap: 10px;
}

.contract-select-filter :deep(.el-form-item) {
  margin-right: 0;
  margin-bottom: 0;
}

</style>
