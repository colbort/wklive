<template>
  <el-popover
    v-model:visible="visible"
    placement="bottom-start"
    trigger="click"
    width="620"
    popper-class="symbol-select-popover"
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

    <div class="symbol-select-panel">
      <el-form inline class="symbol-select-filter" @submit.prevent>
        <el-form-item :label="t('common.keyword')">
          <el-input
            v-model="queryKeyword"
            clearable
            style="width: 180px"
            @keyup.enter="searchSymbols"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="searchSymbols">
            {{ t('common.search') }}
          </el-button>
          <el-button @click="resetQuery">
            {{ t('common.reset') }}
          </el-button>
        </el-form-item>
      </el-form>

      <el-table
        v-loading="loading"
        :data="symbols"
        height="300"
        highlight-current-row
        @row-click="selectSymbol"
      >
        <el-table-column prop="id" :label="t('trade.symbolId')" width="90" />
        <el-table-column prop="symbol" :label="t('trade.symbol')" min-width="140" />
        <el-table-column prop="displaySymbol" :label="t('trade.displaySymbol')" min-width="150" />
        <el-table-column prop="tenantId" :label="t('trade.tenantId')" width="100" />
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
import { tradeService, type TradeSymbol } from '@/services'

const props = withDefaults(
  defineProps<{
    modelValue?: number
    disabled?: boolean
    placeholder?: string
    tenantId?: number
    marketType?: number
  }>(),
  {
    modelValue: undefined,
    disabled: false,
    placeholder: '',
    tenantId: undefined,
    marketType: undefined,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: number | undefined]
  change: [value: number | undefined]
  selected: [value: TradeSymbol | null]
}>()

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

const visible = ref(false)
const loading = ref(false)
const symbols = ref<TradeSymbol[]>([])
const selectedSymbol = ref<TradeSymbol | null>(null)
const querySymbolId = ref<number | undefined>(undefined)
const queryKeyword = ref('')

const displayValue = computed(() => {
  if (!props.modelValue) return ''
  const symbol = selectedSymbol.value
  if (symbol && symbol.id === props.modelValue) {
    return `${symbol.id} - ${symbol.displaySymbol || symbol.symbol}`
  }
  return String(props.modelValue)
})

watch(
  () => props.modelValue,
  (value) => {
    querySymbolId.value = value
    if (!value || selectedSymbol.value?.id !== value) {
      selectedSymbol.value = null
    }
  },
)

watch(
  () => props.tenantId,
  () => {
    symbols.value = []
  },
)

async function loadSymbols() {
  loading.value = true
  try {
    if (querySymbolId.value) {
      const detail = await tradeService.getSymbol({
        tenantId: props.tenantId || undefined,
        id: querySymbolId.value,
      })
      const symbol = detail.data?.symbol
      symbols.value = symbol ? [symbol] : []
      selectedSymbol.value = symbol && symbol.id === props.modelValue ? symbol : selectedSymbol.value
      updateFromResponse({ total: symbols.value.length })
      return
    }

    const res = await tradeService.listSymbols({
      cursor: pagination.cursor,
      limit: pagination.limit,
      tenantId: props.tenantId || undefined,
      marketType: props.marketType || undefined,
      keyword: queryKeyword.value || undefined,
    })
    symbols.value = res.data || []
    updateFromResponse(res)
    syncSelectedFromList()
  } finally {
    loading.value = false
  }
}

function syncSelectedFromList() {
  if (!props.modelValue) return
  const current = symbols.value.find((item) => item.id === props.modelValue)
  if (current) selectedSymbol.value = current
}

function handleShow() {
  querySymbolId.value = props.modelValue
  resetAndLoad(loadSymbols)
}

function searchSymbols() {
  resetAndLoad(loadSymbols)
}

function resetQuery() {
  querySymbolId.value = undefined
  queryKeyword.value = ''
  resetAndLoad(loadSymbols)
}

function handleLimitChange() {
  resetAndLoad(loadSymbols)
}

function handlePrevPage() {
  prevAndLoad(loadSymbols)
}

function handleNextPage() {
  nextAndLoad(loadSymbols)
}

function selectSymbol(row: TradeSymbol) {
  selectedSymbol.value = row
  emit('update:modelValue', row.id)
  emit('change', row.id)
  emit('selected', row)
  visible.value = false
}

function clearValue() {
  selectedSymbol.value = null
  emit('update:modelValue', undefined)
  emit('change', undefined)
  emit('selected', null)
}
</script>

<style scoped>
.symbol-select-panel {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.symbol-select-filter {
  display: flex;
  align-items: center;
  gap: 10px;
}

.symbol-select-filter :deep(.el-form-item) {
  margin-right: 0;
  margin-bottom: 0;
}

</style>
