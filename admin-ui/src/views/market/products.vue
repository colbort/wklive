<template>
  <div class="market-products module-page">
    <CrudQueryCard :model="queryParams" @search="loadList" @reset="resetQuery">
      <el-form-item :label="t('market.categoryType')">
        <el-select
          v-model="queryParams.categoryType"
          :placeholder="t('common.pleaseSelect')"
          clearable
          style="width: 180px"
        >
          <el-option
            v-for="item in categoryTypeOptions"
            :key="item.value"
            :label="getOptionLabel(t, item.code, item.value)"
            :value="item.value"
          />
        </el-select>
      </el-form-item>

      <el-form-item :label="t('market.market')">
        <el-input
          v-model="queryParams.market"
          :placeholder="t('market.pleaseInputMarket')"
          clearable
          style="width: 180px"
          @keyup.enter="loadList"
        />
      </el-form-item>

      <el-form-item :label="t('market.symbol')">
        <el-input
          v-model="queryParams.symbol"
          :placeholder="t('market.pleaseInputSymbol')"
          clearable
          style="width: 180px"
          @keyup.enter="loadList"
        />
      </el-form-item>

      <el-form-item :label="t('market.keyword')">
        <el-input
          v-model="queryParams.keyword"
          :placeholder="t('common.keyword')"
          clearable
          style="width: 180px"
          @keyup.enter="loadList"
        />
      </el-form-item>

      <el-form-item :label="t('market.enabledStatus')">
        <el-select
          v-model="queryParams.enabled"
          :placeholder="t('common.pleaseSelect')"
          clearable
          style="width: 180px"
        >
          <el-option
            v-for="item in enabledOptions"
            :key="item.value"
            :label="getOptionLabel(t, item.code, item.value)"
            :value="item.value"
          />
        </el-select>
      </el-form-item>

      <el-form-item :label="t('market.appVisible')">
        <el-select
          v-model="queryParams.appVisible"
          :placeholder="t('common.pleaseSelect')"
          clearable
          style="width: 180px"
        >
          <el-option
            v-for="item in visibleOptions"
            :key="item.value"
            :label="getOptionLabel(t, item.code, item.value)"
            :value="item.value"
          />
        </el-select>
      </el-form-item>

      <template #actions>
        <el-button v-perm="'market:product:add'" type="primary" @click="handleAdd">
          <el-icon><Plus /></el-icon>
          {{ t('common.add') }}
        </el-button>
      </template>
    </CrudQueryCard>

    <el-card class="table-card" shadow="never">
      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" :label="t('common.id')" width="80" />
        <el-table-column :label="t('market.categoryType')" width="120">
          <template #default="{ row }">
            {{ getOptionValueLabel(optionGroups, 'categoryType', row.categoryType, t) }}
          </template>
        </el-table-column>
        <el-table-column prop="categoryName" :label="t('market.categoryName')" min-width="140" />
        <el-table-column prop="categoryCode" :label="t('market.categoryCode')" min-width="140" />
        <el-table-column prop="market" :label="t('market.market')" width="100" />
        <el-table-column prop="symbol" :label="t('market.symbol')" min-width="120" />
        <el-table-column prop="code" :label="t('market.code')" min-width="120" />
        <el-table-column prop="name" :label="t('market.name')" min-width="140" />
        <el-table-column prop="displayName" :label="t('market.displayName')" min-width="140" />
        <el-table-column prop="baseCoin" :label="t('market.baseCoin')" width="100" />
        <el-table-column prop="quoteCoin" :label="t('market.quoteCoin')" width="100" />

        <el-table-column :label="t('market.enabledStatus')" width="100">
          <template #default="{ row }">
            <el-tag :type="row.enabled === 1 ? 'success' : 'info'">
              {{ getOptionValueLabel(optionGroups, 'enabled', row.enabled, t) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column :label="t('market.appVisible')" width="100">
          <template #default="{ row }">
            <el-tag :type="row.appVisible === 1 ? 'success' : 'warning'">
              {{ getOptionValueLabel(optionGroups, 'visible', row.appVisible, t) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column :label="t('market.syncPriority')" width="110">
          <template #default="{ row }">
            <el-tag :type="syncPriorityTagType(row.syncPriority)">
              {{ getSyncPriorityLabel(row.syncPriority) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="sort" :label="t('common.sort')" width="90" />
        <el-table-column :label="t('common.icon')" min-width="180">
          <template #default="{ row }">
            <div v-if="row.icon" class="icon-cell">
              <el-image
                :src="resolveAssetUrl(row.icon)"
                class="icon-preview"
                :preview-teleported="true"
              />
            </div>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column
          prop="remark"
          :label="t('common.remark')"
          min-width="180"
          show-overflow-tooltip
        />

        <el-table-column :label="t('common.createTimes')" min-width="170">
          <template #default="{ row }">
            {{ formatDate(row.createTimes) }}
          </template>
        </el-table-column>

        <el-table-column :label="t('market.updateTimes')" min-width="170">
          <template #default="{ row }">
            {{ formatDate(row.updateTimes) }}
          </template>
        </el-table-column>

        <el-table-column
          :label="t('common.actions')"
          align="center"
          width="210"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              v-perm="'market:product:detail'"
              link
              type="primary"
              @click="handleDetail(row)"
            >
              {{ t('market.detail') }}
            </el-button>
            <el-button
              v-perm="'market:kline:view'"
              link
              type="primary"
              @click="handleKline(row)"
            >
              {{ t('market.klineView') }}
            </el-button>
            <el-button
              v-perm="'market:product:update'"
              link
              type="primary"
              @click="handleEdit(row)"
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
        :show-total="false"
        @prev="handlePrevPage"
        @next="handleNextPage"
        @limit-change="handleLimitChange"
      />
    </el-card>

    <el-dialog
      v-model="formDialogVisible"
      :title="formMode === 'add' ? t('market.addProduct') : t('market.editProduct')"
      width="700px"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="120px"
      >
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item
              v-if="formMode === 'add'"
              :label="t('market.categoryType')"
              prop="categoryType"
            >
              <el-select
                v-model="form.categoryType"
                :placeholder="t('common.pleaseSelect')"
                style="width: 100%"
              >
                <el-option
                  v-for="item in categoryTypeFormOptions"
                  :key="item.value"
                  :label="getOptionLabel(t, item.code, item.value)"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('market.categoryName')" prop="categoryName">
              <el-input
                v-model="form.categoryName"
                :placeholder="t('market.pleaseInputCategoryName')"
                maxlength="100"
                show-word-limit
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="t('market.categoryCode')" prop="categoryCode">
              <el-input
                v-model="form.categoryCode"
                :placeholder="t('market.categoryCode')"
                maxlength="50"
                show-word-limit
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('market.market')" prop="market">
              <el-input
                v-model="form.market"
                :placeholder="t('market.pleaseInputMarket')"
                maxlength="50"
                show-word-limit
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="t('market.symbol')" prop="symbol">
              <el-input
                v-model="form.symbol"
                :placeholder="t('market.pleaseInputSymbol')"
                maxlength="50"
                show-word-limit
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('market.code')" prop="code">
              <el-input
                v-model="form.code"
                :placeholder="t('market.pleaseInputCode')"
                maxlength="50"
                show-word-limit
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="t('market.name')" prop="name">
              <el-input
                v-model="form.name"
                :placeholder="t('market.pleaseInputName')"
                maxlength="100"
                show-word-limit
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('market.displayName')" prop="displayName">
              <el-input
                v-model="form.displayName"
                :placeholder="t('market.pleaseInputDisplayName')"
                maxlength="100"
                show-word-limit
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="t('market.baseCoin')" prop="baseCoin">
              <el-input
                v-model="form.baseCoin"
                :placeholder="t('market.pleaseInputBaseCoin')"
                maxlength="20"
                show-word-limit
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('market.quoteCoin')" prop="quoteCoin">
              <el-input
                v-model="form.quoteCoin"
                :placeholder="t('market.pleaseInputQuoteCoin')"
                maxlength="20"
                show-word-limit
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="t('market.enabledStatus')" prop="enabled">
              <el-select v-model="form.enabled" style="width: 100%">
                <el-option
                  v-for="item in enabledFormOptions"
                  :key="item.value"
                  :label="getOptionLabel(t, item.code, item.value)"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('market.appVisible')" prop="appVisible">
              <el-select v-model="form.appVisible" style="width: 100%">
                <el-option
                  v-for="item in visibleFormOptions"
                  :key="item.value"
                  :label="getOptionLabel(t, item.code, item.value)"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="t('market.syncPriority')" prop="syncPriority">
              <el-select v-model="form.syncPriority" style="width: 100%">
                <el-option
                  v-for="item in syncPriorityFormOptions"
                  :key="item.value"
                  :label="getOptionLabel(t, item.code, item.value)"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('common.sort')" prop="sort">
              <el-input-number
                v-model="form.sort"
                :min="0"
                :precision="0"
                controls-position="right"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="t('common.icon')" prop="icon">
              <div class="icon-upload-field">
                <div v-if="form.icon" class="icon-upload-preview">
                  <el-image
                    :src="resolveAssetUrl(form.icon)"
                    class="icon-preview-large"
                    :preview-teleported="true"
                  />
                  <div class="icon-url">
                    {{ form.icon }}
                  </div>
                </div>
                <el-upload
                  action="#"
                  :auto-upload="false"
                  :show-file-list="false"
                  :on-change="handleIconSelect"
                  accept="image/*"
                >
                  <el-button type="primary" :loading="submitLoading">
                    {{ t('market.uploadImage') }}
                  </el-button>
                </el-upload>
              </div>
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item :label="t('common.remark')" prop="remark">
          <el-input
            v-model="form.remark"
            type="textarea"
            :rows="4"
            :placeholder="t('common.remark')"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="formDialogVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button
          v-perm="formMode === 'add' ? 'market:product:add' : 'market:product:update'"
          type="primary"
          :loading="submitLoading"
          @click="submitForm"
        >
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="klineDialogVisible"
      class="kline-history-dialog"
      :title="`${t('market.klineHistory')} - ${klineProduct?.symbol || ''}`"
      top="5vh"
      width="1100px"
    >
      <div class="kline-history-toolbar">
        <el-form :inline="true" :model="klineQuery">
          <el-form-item :label="t('market.klineType')">
            <el-select v-model="klineQuery.kType" style="width: 160px">
              <el-option
                v-for="item in klineTypeOptions"
                :key="item.value"
                :label="getOptionLabel(t, item.code, item.value)"
                :value="item.value"
              />
            </el-select>
          </el-form-item>
          <el-form-item :label="t('market.endTime')">
            <el-date-picker
              v-model="klineQuery.endTs"
              type="datetime"
              value-format="x"
              :placeholder="t('market.latestData')"
              clearable
            />
          </el-form-item>
          <el-form-item :label="t('market.klineLimit')">
            <el-input-number
              v-model="klineQuery.limit"
              :min="1"
              :max="5000"
              :step="100"
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="klineLoading" @click="loadKlines">
              {{ t('common.search') }}
            </el-button>
            <el-button
              v-perm="'market:kline:syncHistory'"
              type="warning"
              :loading="klineSyncLoading"
              @click="syncKlineHistory"
            >
              {{ t('market.syncKlineHistory') }}
            </el-button>
          </el-form-item>
        </el-form>

        <el-alert
          :title="t('market.syncKlineHistoryTip')"
          type="info"
          :closable="false"
          show-icon
        />
      </div>
      <el-table
        v-loading="klineLoading"
        :data="klineList"
        stripe
        height="100%"
      >
        <el-table-column :label="t('market.klineTime')" min-width="180">
          <template #default="{ row }">
            {{ formatDate(row.ts) }}
          </template>
        </el-table-column>
        <el-table-column prop="open" :label="t('market.open')" min-width="120" />
        <el-table-column prop="high" :label="t('market.high')" min-width="120" />
        <el-table-column prop="low" :label="t('market.low')" min-width="120" />
        <el-table-column prop="close" :label="t('market.close')" min-width="120" />
        <el-table-column prop="volume" :label="t('market.volume')" min-width="130" />
        <el-table-column prop="turnover" :label="t('market.turnover')" min-width="140" />
      </el-table>
    </el-dialog>

    <el-dialog v-model="detailDialogVisible" :title="t('market.productDetail')" width="800px">
      <el-descriptions v-loading="detailLoading" :column="2" border>
        <el-descriptions-item :label="t('common.id')">
          {{ detail.id ?? '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('market.categoryType')">
          {{ getOptionValueLabel(optionGroups, 'categoryType', detail.categoryType, t) || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('market.categoryName')">
          {{ detail.categoryName || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('market.categoryCode')">
          {{ detail.categoryCode || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('market.market')">
          {{ detail.market || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('market.symbol')">
          {{ detail.symbol || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('market.code')">
          {{ detail.code || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('market.name')">
          {{ detail.name || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('market.displayName')">
          {{ detail.displayName || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('market.baseCoin')">
          {{ detail.baseCoin || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('market.quoteCoin')">
          {{ detail.quoteCoin || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('market.enabledStatus')">
          {{ getOptionValueLabel(optionGroups, 'enabled', detail.enabled, t) || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('market.appVisible')">
          {{ getOptionValueLabel(optionGroups, 'visible', detail.appVisible, t) || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('market.syncPriority')">
          {{ getSyncPriorityLabel(detail.syncPriority) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.sort')">
          {{ detail.sort ?? '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.icon')">
          <div v-if="detail.icon" class="icon-detail">
            <el-image
              :src="resolveAssetUrl(detail.icon)"
              class="icon-preview-large"
              :preview-teleported="true"
            />
            <div class="icon-url">
              {{ detail.icon }}
            </div>
          </div>
          <span v-else>-</span>
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.remark')" :span="2">
          {{ detail.remark || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.createTimes')">
          {{ formatDate(detail.createTimes ?? 0) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('market.updateTimes')">
          {{ formatDate(detail.updateTimes ?? 0) }}
        </el-descriptions-item>
      </el-descriptions>

      <el-divider content-position="left">
        {{ t('market.tradingHours') }}
      </el-divider>
      <el-empty
        v-if="!detail.tradingCalendar?.id"
        :description="t('market.noTradingCalendar')"
        :image-size="64"
      />
      <el-descriptions v-else :column="2" border>
        <el-descriptions-item :label="t('market.calendarScope')">
          {{
            detail.tradingCalendar.productSpecific
              ? t('market.productSpecificCalendar')
              : t('market.marketDefaultCalendar')
          }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('market.timezone')">
          {{ detail.tradingCalendar.timezone || 'UTC' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('market.market')">
          {{ detail.tradingCalendar.market || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('market.exchange')">
          {{ detail.tradingCalendar.exchange || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('market.weeklySessions')" :span="2">
          <div v-if="detail.tradingCalendar.sessions?.length" class="trading-session-list">
            <div
              v-for="session in detail.tradingCalendar.sessions"
              :key="session.id"
              class="trading-session-item"
            >
              <el-tag size="small" effect="plain">
                {{ formatWeekdayMask(session.weekdayMask) }}
              </el-tag>
              <span>{{ session.startTime }}–{{ session.endTime }}</span>
              <el-tag
                v-if="session.crossDay"
                size="small"
                type="warning"
                effect="plain"
              >
                {{ t('market.crossDay') }}
              </el-tag>
              <span class="session-type">{{ session.sessionType }}</span>
            </div>
          </div>
          <span v-else>-</span>
        </el-descriptions-item>
      </el-descriptions>

      <template #footer>
        <el-button type="primary" @click="detailDialogVisible = false">
          {{ t('common.close') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, onMounted } from 'vue'
import { ElMessage, type FormRules, type UploadFile } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables/usePagination'
import { useLoading } from '@/composables/useLoading'
import { useForm } from '@/composables/useForm'
import { buildSystemAssetUrl, useSystemCore } from '@/composables/useSystemCore'
import type { OptionGroup } from '@/services'
import { apiUploadFile } from '@/api/system/upload'
import {
  productsService,
  type MarketProduct,
  type ListProductsReq,
  type Kline,
} from '@/services/market/ProductsService'
import { formatDate } from '@/utils'
import {
  findFormOptionGroup,
  findOptionGroup,
  getOptionLabel,
  getOptionValueLabel,
} from '@/utils/options'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

type FormData = {
  id?: number
  categoryType?: number
  categoryName: string
  categoryCode: string
  market: string
  symbol: string
  code: string
  name: string
  displayName: string
  baseCoin: string
  quoteCoin: string
  enabled: number
  appVisible: number
  syncPriority: number
  sort: number
  icon: string
  remark: string
}

const { t } = useI18n()
const { systemCore, loadSystemCore } = useSystemCore()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)
const { loading, withLoading } = useLoading()

const { form: queryParams, reset: resetQueryParams } = useForm<ListProductsReq>({
  initialData: {
    categoryType: undefined,
    market: '',
    symbol: '',
    keyword: '',
    enabled: 0,
    appVisible: 0,
    cursor: undefined,
    limit: 20,
  },
})

const {
  form: form,
  formRef,
  reset: resetForm,
} = useForm<FormData>({
  initialData: {
    id: undefined,
    categoryType: undefined,
    categoryName: '',
    categoryCode: '',
    market: '',
    symbol: '',
    code: '',
    name: '',
    displayName: '',
    baseCoin: '',
    quoteCoin: '',
    enabled: 1,
    appVisible: 1,
    syncPriority: 2,
    sort: 0,
    icon: '',
    remark: '',
  },
})

const submitLoading = ref(false)
const detailLoading = ref(false)
const list = ref<MarketProduct[]>([])
const detail = ref<Partial<MarketProduct>>({})
const optionGroups = ref<OptionGroup[]>([])
const formDialogVisible = ref(false)
const detailDialogVisible = ref(false)
const klineDialogVisible = ref(false)
const klineLoading = ref(false)
const klineSyncLoading = ref(false)
const klineProduct = ref<MarketProduct>()
const klineList = ref<Kline[]>([])
const klineQuery = ref<{
  kType: number
  endTs: number | string | null
  limit: number
}>({
  kType: 1,
  endTs: null,
  limit: 500,
})
const formMode = ref<'add' | 'edit'>('add')
const categoryTypeOptions = computed(() => findOptionGroup(optionGroups.value, 'categoryType'))
const klineTypeOptions = computed(() => findOptionGroup(optionGroups.value, 'klineType'))
const enabledOptions = computed(() => findOptionGroup(optionGroups.value, 'enabled'))
const visibleOptions = computed(() => findOptionGroup(optionGroups.value, 'visible'))
const categoryTypeFormOptions = computed(() =>
  findFormOptionGroup(optionGroups.value, 'categoryType'),
)
const enabledFormOptions = computed(() => findFormOptionGroup(optionGroups.value, 'enabled'))
const visibleFormOptions = computed(() => findFormOptionGroup(optionGroups.value, 'visible'))
const syncPriorityFormOptions = computed(() =>
  findFormOptionGroup(optionGroups.value, 'syncPriority'),
)
const resolveAssetUrl = (url?: string) => buildSystemAssetUrl(systemCore.value.assetUrl, url)
const getSyncPriorityLabel = (value?: number) =>
  getOptionValueLabel(optionGroups.value, 'syncPriority', Number(value), t) || '-'
const syncPriorityTagType = (value?: number) => {
  if (Number(value) === 1) return 'danger'
  if (Number(value) === 3) return 'info'
  return 'success'
}

const rules: FormRules<FormData> = {
  categoryType: [
    {
      required: true,
      message: t('market.pleaseInputCategoryType'),
      trigger: 'blur',
    },
  ],
  categoryName: [
    {
      required: true,
      message: t('market.pleaseInputCategoryName'),
      trigger: 'blur',
    },
  ],
  categoryCode: [{ required: true, message: t('market.categoryCode'), trigger: 'blur' }],
  market: [{ required: true, message: t('market.pleaseInputMarket'), trigger: 'blur' }],
  symbol: [{ required: true, message: t('market.pleaseInputSymbol'), trigger: 'blur' }],
  code: [{ required: true, message: t('market.pleaseInputCode'), trigger: 'blur' }],
  name: [{ required: true, message: t('market.pleaseInputName'), trigger: 'blur' }],
  displayName: [
    {
      required: true,
      message: t('market.pleaseInputDisplayName'),
      trigger: 'blur',
    },
  ],
  baseCoin: [
    {
      required: true,
      message: t('market.pleaseInputBaseCoin'),
      trigger: 'blur',
    },
  ],
  quoteCoin: [
    {
      required: true,
      message: t('market.pleaseInputQuoteCoin'),
      trigger: 'blur',
    },
  ],
  enabled: [
    {
      required: true,
      message: t('market.pleaseSelectEnabledStatus'),
      trigger: 'change',
    },
  ],
  appVisible: [
    {
      required: true,
      message: t('market.pleaseSelectAppVisible'),
      trigger: 'change',
    },
  ],
  syncPriority: [
    {
      required: true,
      message: t('market.pleaseSelectSyncPriority'),
      trigger: 'change',
    },
  ],
  sort: [{ required: true, message: t('market.pleaseInputSort'), trigger: 'blur' }],
}

const cleanedQueryParams = computed<ListProductsReq>(() => {
  const params: ListProductsReq = {
    cursor: queryParams.cursor,
    limit: queryParams.limit,
  }

  if (queryParams.categoryType && queryParams.categoryType !== 0) {
    params.categoryType = Number(queryParams.categoryType)
  }
  if (queryParams.market && queryParams.market.trim()) {
    params.market = queryParams.market.trim()
  }
  if (queryParams.symbol && queryParams.symbol.trim()) {
    params.symbol = queryParams.symbol.trim()
  }
  if (queryParams.keyword && queryParams.keyword.trim()) {
    params.keyword = queryParams.keyword.trim()
  }
  if (queryParams.enabled && queryParams.enabled !== 0) {
    params.enabled = queryParams.enabled
  }
  if (queryParams.appVisible && queryParams.appVisible !== 0) {
    params.appVisible = queryParams.appVisible
  }

  return params
})

const getList = async () => {
  await withLoading(async () => {
    try {
      const res = await productsService.getList({
        ...cleanedQueryParams.value,
        cursor: pagination.cursor,
      })
      list.value = res?.data || []
      updateFromResponse(res)
    } catch (_) {
      ElMessage.error(t('common.loadFailed'))
    }
  })
}

const loadOptions = async () => {
  try {
    const res = await productsService.getOptions()
    optionGroups.value = res.data || []
  } catch {
    ElMessage.error(t('common.loadFailed'))
  }
}

const loadList = () => {
  resetAndLoad(getList)
}

const resetQuery = () => {
  resetQueryParams()
  resetAndLoad(getList)
}

const handleLimitChange = () => {
  resetAndLoad(getList)
}

const handleAdd = async () => {
  formMode.value = 'add'
  resetForm()
  formDialogVisible.value = true
  await nextTick()
  formRef.value?.clearValidate()
}

const handleEdit = async (row: MarketProduct) => {
  formMode.value = 'edit'
  resetForm()

  try {
    const res = await productsService.detail(row.id)
    const data = res?.data || row

    Object.assign(form, {
      id: data.id,
      categoryType: data.categoryType,
      categoryName: data.categoryName || '',
      categoryCode: data.categoryCode || '',
      market: data.market || '',
      symbol: data.symbol || '',
      code: data.code || '',
      name: data.name || '',
      displayName: data.displayName || '',
      baseCoin: data.baseCoin || '',
      quoteCoin: data.quoteCoin || '',
      enabled: data.enabled,
      appVisible: data.appVisible,
      syncPriority: data.syncPriority || 2,
      sort: data.sort || 0,
      icon: data.icon || '',
      remark: data.remark || '',
    })

    formDialogVisible.value = true
    await nextTick()
    formRef.value?.clearValidate()
  } catch (_) {
    ElMessage.error(t('common.loadFailed'))
  }
}

const handleDetail = async (row: MarketProduct) => {
  detailDialogVisible.value = true
  detailLoading.value = true
  detail.value = {}

  try {
    const res = await productsService.detail(row.id)
    detail.value = res?.data || {}
  } catch (_) {
    ElMessage.error(t('common.loadFailed'))
  } finally {
    detailLoading.value = false
  }
}

const formatWeekdayMask = (mask: number) => {
  const names = [
    t('market.weekdaySunday'),
    t('market.weekdayMonday'),
    t('market.weekdayTuesday'),
    t('market.weekdayWednesday'),
    t('market.weekdayThursday'),
    t('market.weekdayFriday'),
    t('market.weekdaySaturday'),
  ]
  const selected = names.filter((_, index) => (Number(mask) & (1 << index)) !== 0)
  return selected.length ? selected.join('、') : '-'
}

const buildKlineRequest = () => {
  const product = klineProduct.value
  if (!product) return null
  return {
    categoryCode: product.categoryCode,
    market: product.market,
    symbol: product.symbol,
    kType: Number(klineQuery.value.kType),
    endTs: Number(klineQuery.value.endTs || 0),
    limit: Number(klineQuery.value.limit),
  }
}

const loadKlines = async () => {
  const params = buildKlineRequest()
  if (!params) return
  klineLoading.value = true
  try {
    const res = await productsService.kline(params)
    klineList.value = res.data || []
  } catch {
    ElMessage.error(t('common.loadFailed'))
  } finally {
    klineLoading.value = false
  }
}

const handleKline = (row: MarketProduct) => {
  klineProduct.value = row
  klineList.value = []
  klineQuery.value.endTs = null
  klineDialogVisible.value = true
  loadKlines()
}

const syncKlineHistory = async () => {
  const query = buildKlineRequest()
  if (!query) return
  const { limit: _queryLimit, ...params } = query
  klineSyncLoading.value = true
  try {
    const res = await productsService.syncKlineHistory(params)
    ElMessage.success(t('market.syncKlineHistorySuccess', { count: res.syncedCount || 0 }))
    await loadKlines()
  } catch {
    ElMessage.error(t('market.syncKlineHistoryFailed'))
  } finally {
    klineSyncLoading.value = false
  }
}

const handleIconSelect = async (uploadFile: UploadFile) => {
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
      form.icon = res.data?.url || ''
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

const submitForm = async () => {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    if (formMode.value === 'add') {
      await productsService.create({
        categoryType: Number(form.categoryType),
        categoryName: form.categoryName,
        categoryCode: form.categoryCode,
        market: form.market,
        symbol: form.symbol,
        code: form.code,
        name: form.name,
        displayName: form.displayName,
        baseCoin: form.baseCoin,
        quoteCoin: form.quoteCoin,
        enabled: form.enabled,
        appVisible: form.appVisible,
        syncPriority: form.syncPriority,
        sort: form.sort,
        icon: form.icon,
        remark: form.remark,
      })
      ElMessage.success(t('common.createSuccess'))
    } else {
      await productsService.update(form.id as number, {
        name: form.name,
        displayName: form.displayName,
        baseCoin: form.baseCoin,
        quoteCoin: form.quoteCoin,
        enabled: form.enabled,
        appVisible: form.appVisible,
        syncPriority: form.syncPriority,
        sort: form.sort,
        icon: form.icon,
        remark: form.remark,
      })
      ElMessage.success(t('common.updateSuccess'))
    }

    formDialogVisible.value = false
    getList()
  } catch (_) {
    ElMessage.error(formMode.value === 'add' ? t('common.createFailed') : t('common.updateFailed'))
  } finally {
    submitLoading.value = false
  }
}

const handlePrevPage = () => {
  prevAndLoad(getList)
}

const handleNextPage = () => {
  nextAndLoad(getList)
}

onMounted(() => {
  loadSystemCore()
  loadOptions()
  getList()
})
</script>

<style scoped>
.icon-cell,
.icon-upload-field,
.icon-detail {
  display: flex;
  align-items: center;
  gap: 12px;
}

.icon-preview {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  flex-shrink: 0;
}

.icon-preview-large {
  width: 64px;
  height: 64px;
  border-radius: 12px;
  flex-shrink: 0;
}

.icon-url {
  color: #606266;
  word-break: break-all;
}

:deep(.kline-history-dialog) {
  display: flex;
  height: 90vh;
  margin-bottom: 0;
  flex-direction: column;
}

:deep(.kline-history-dialog .el-dialog__body) {
  display: flex;
  min-height: 0;
  flex: 1;
  flex-direction: column;
  overflow: hidden;
}

.kline-history-toolbar {
  flex-shrink: 0;
  padding-bottom: 16px;
}

.trading-session-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.trading-session-item {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.session-type {
  color: var(--el-text-color-secondary);
}
</style>
