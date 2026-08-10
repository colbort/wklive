<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="loadList" @reset="resetQuery">
      <el-form-item :label="t('trade.tenantId')">
        <div class="query-field">
          <TenantSelect v-model="query.tenantId" class="tenant-select-filter" include-system />
        </div>
      </el-form-item>

      <el-form-item :label="t('trade.productType')">
        <el-select v-model="query.productType" clearable class="query-field">
          <el-option
            v-for="item in productTypeOptions"
            :key="item.value"
            :label="optionItemLabel(item)"
            :value="item.value"
          />
        </el-select>
      </el-form-item>

      <el-form-item :label="t('market.categoryType')">
        <el-select v-model="query.categoryType" clearable class="query-field">
          <el-option
            v-for="item in categoryTypeOptions"
            :key="item.value"
            :label="optionItemLabel(item)"
            :value="item.value"
          />
        </el-select>
      </el-form-item>

      <el-form-item :label="t('trade.status')">
        <el-select v-model="query.status" clearable class="query-field">
          <el-option
            v-for="item in symbolStatusOptions"
            :key="item.value"
            :label="optionItemLabel(item)"
            :value="item.value"
          />
        </el-select>
      </el-form-item>

      <el-form-item :label="t('common.keyword')">
        <el-input v-model="query.keyword" clearable class="query-keyword" />
      </el-form-item>

      <template #actions>
        <el-button v-perm="'trade:symbol:add'" type="primary" @click="openSymbolDialog()">
          <el-icon><Plus /></el-icon>
          {{ t('common.add') }}
        </el-button>
      </template>
    </CrudQueryCard>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column prop="tenantId" :label="t('trade.tenantId')" width="100" />

        <el-table-column min-width="190" show-overflow-tooltip>
          <template #header>
            {{ t('trade.symbol') }} / {{ t('trade.displaySymbol') }}
          </template>
          <template #default="{ row }">
            <div class="symbol-cell">
              <span class="symbol-code">{{ row.symbol || '-' }}/{{ row.displaySymbol || '-' }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column :label="t('market.categoryType')" width="120">
          <template #default="{ row }">
            <el-tag v-if="row.categoryType" size="small" effect="light">
              {{ optionLabel('categoryType', row.categoryType) }}
            </el-tag>
            <span v-else class="muted">-</span>
          </template>
        </el-table-column>

        <el-table-column :label="t('trade.productType')" min-width="130">
          <template #default="{ row }">
            <el-tag size="small" effect="light">
              {{ optionLabel('productType', row.productType) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column :label="t('trade.contractType')" min-width="130">
          <template #default="{ row }">
            {{ optionLabel('contractType', row.contractType) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('trade.contractValueType')" min-width="150">
          <template #default="{ row }">
            {{ optionLabel('contractValueType', row.contractValueType) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('trade.status')" width="120">
          <template #default="{ row }">
            <el-tag size="small" :type="symbolStatusTagType(row.status)" effect="light">
              {{ optionLabel('symbolStatus', row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column min-width="230" show-overflow-tooltip>
          <template #header>
            {{ t('trade.baseAsset') }} / {{ t('trade.quoteAsset') }} /
            {{ t('trade.settleAsset') }}
          </template>
          <template #default="{ row }">
            <div class="asset-pair">
              <el-tag size="small">
                {{ row.baseAsset || '-' }}
              </el-tag>
              <span>/</span>
              <el-tag size="small" type="info">
                {{ row.quoteAsset || '-' }}
              </el-tag>
              <span>/</span>
              <span v-if="row.settleAsset" class="muted">{{ row.settleAsset }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column :label="t('trade.priceTick')" min-width="120">
          <template #default="{ row }">
            {{ row.priceTick || '-' }}
          </template>
        </el-table-column>
        <el-table-column :label="t('trade.qtyStep')" min-width="120">
          <template #default="{ row }">
            {{ row.qtyStep || '-' }}
          </template>
        </el-table-column>
        <el-table-column :label="t('trade.maxNotional')" min-width="130">
          <template #default="{ row }">
            {{ row.maxNotional || '-' }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.sort')" width="90">
          <template #default="{ row }">
            {{ row.sort || 0 }}
          </template>
        </el-table-column>
        <el-table-column
          :label="t('common.actions')"
          align="center"
          width="260"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              v-perm="'trade:symbol:detail'"
              link
              type="primary"
              @click="showDetail(row)"
            >
              {{ t('option.detail') }}
            </el-button>
            <el-button
              v-perm="'trade:symbol:update'"
              link
              type="primary"
              @click="openSymbolDialog(row)"
            >
              {{ t('common.edit') }}
            </el-button>
            <el-button
              v-if="isSpotMarket(row)"
              v-perm="'trade:symbol:spot-config'"
              link
              type="primary"
              @click="openSpotDialog(row)"
            >
              {{ t('trade.spotConfig') }}
            </el-button>
            <el-button
              v-if="isContractMarket(row)"
              v-perm="'trade:symbol:contract-config'"
              link
              type="primary"
              @click="openContractDialog(row)"
            >
              {{ t('trade.contractConfig') }}
            </el-button>
            <el-button
              v-if="row.productType === 3"
              v-perm="'trade:symbol:seconds-config'"
              link
              type="primary"
              @click="openSecondsDialog(row)"
            >
              {{ t('trade.secondsConfig') }}
            </el-button>
            <el-button
              v-if="isContractMarket(row)"
              v-perm="'trade:symbol:leverage-config:update'"
              link
              type="primary"
              @click="openLeverageDialog(row)"
            >
              {{ t('trade.leverageConfig') }}
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
      v-model="symbolVisible"
      :title="symbolForm.id ? t('trade.editSymbol') : t('trade.addSymbol')"
      width="920px"
    >
      <el-form label-width="108px" class="dialog-form">
        <div class="form-grid">
          <el-form-item :label="t('trade.tenantId')">
            <TenantSelect v-model="symbolForm.tenantId" include-system />
          </el-form-item>

          <el-form-item :label="t('trade.symbol')">
            <el-input v-model="symbolForm.symbol" :disabled="Boolean(symbolForm.id)" />
          </el-form-item>

          <el-form-item :label="t('trade.displaySymbol')">
            <el-input v-model="symbolForm.displaySymbol" />
          </el-form-item>

          <el-form-item :label="t('market.categoryType')">
            <el-select
              v-model="symbolForm.categoryType"
              class="full-width"
            >
              <el-option
                v-for="item in categoryTypeFormOptions"
                :key="item.value"
                :label="optionItemLabel(item)"
                :value="item.value"
              />
            </el-select>
          </el-form-item>

          <el-form-item :label="t('trade.productType')">
            <el-select
              v-if="isCryptoCategory"
              v-model="symbolForm.productType"
              class="full-width"
              :disabled="Boolean(symbolForm.id)"
            >
              <el-option
                v-for="item in productTypeFormOptions"
                :key="item.value"
                :label="optionItemLabel(item)"
                :value="item.value"
              />
            </el-select>
            <el-input
              v-else
              :model-value="optionLabel('productType', fixedProductTypeValue)"
              disabled
              class="full-width"
            />
          </el-form-item>

          <el-form-item v-if="isDerivativeProduct" :label="t('trade.contractType')">
            <el-select
              v-model="symbolForm.contractType"
              class="full-width"
              :disabled="Boolean(symbolForm.id)"
            >
              <el-option
                v-for="item in contractTypeFormOptions"
                :key="item.value"
                :label="optionItemLabel(item)"
                :value="item.value"
              />
            </el-select>
          </el-form-item>

          <el-form-item v-if="isDerivativeProduct" :label="t('trade.contractValueType')">
            <el-select
              v-model="symbolForm.contractValueType"
              class="full-width"
              :disabled="Boolean(symbolForm.id) || !isDerivativeProduct"
            >
              <el-option
                v-for="item in contractValueTypeFormOptions"
                :key="item.value"
                :label="optionItemLabel(item)"
                :value="item.value"
              />
            </el-select>
          </el-form-item>

          <el-form-item :label="t('trade.status')">
            <el-select v-model="symbolForm.status" class="full-width">
              <el-option
                v-for="item in symbolStatusFormOptions"
                :key="item.value"
                :label="optionItemLabel(item)"
                :value="item.value"
              />
            </el-select>
          </el-form-item>

          <el-form-item :label="t('trade.baseAsset')">
            <el-input v-model="symbolForm.baseAsset" :disabled="Boolean(symbolForm.id)" />
          </el-form-item>

          <el-form-item :label="t('trade.quoteAsset')">
            <el-input v-model="symbolForm.quoteAsset" :disabled="Boolean(symbolForm.id)" />
          </el-form-item>

          <el-form-item :label="t('trade.settleAsset')">
            <el-input v-model="symbolForm.settleAsset" :disabled="Boolean(symbolForm.id)" />
          </el-form-item>

          <el-form-item v-if="isDerivativeProduct" :label="t('trade.marginAsset')">
            <el-input v-model="symbolForm.marginAsset" :disabled="Boolean(symbolForm.id)" />
          </el-form-item>

          <el-form-item :label="t('trade.priceScale')">
            <el-input-number
              v-model="symbolForm.priceScale"
              :min="0"
              :precision="0"
              class="full-width"
            />
          </el-form-item>

          <el-form-item :label="t('trade.qtyScale')">
            <el-input-number
              v-model="symbolForm.qtyScale"
              :min="0"
              :precision="0"
              class="full-width"
            />
          </el-form-item>

          <el-form-item :label="t('trade.maxNotional')">
            <el-input v-model="symbolForm.maxNotional" />
          </el-form-item>

          <el-form-item :label="t('trade.minPrice')">
            <el-input v-model="symbolForm.minPrice" />
          </el-form-item>

          <el-form-item :label="t('trade.maxPrice')">
            <el-input v-model="symbolForm.maxPrice" />
          </el-form-item>

          <el-form-item :label="t('trade.priceTick')">
            <el-input v-model="symbolForm.priceTick" />
          </el-form-item>

          <el-form-item :label="t('trade.minQty')">
            <el-input v-model="symbolForm.minQty" />
          </el-form-item>

          <el-form-item :label="t('trade.maxQty')">
            <el-input v-model="symbolForm.maxQty" />
          </el-form-item>

          <el-form-item :label="t('trade.qtyStep')">
            <el-input v-model="symbolForm.qtyStep" />
          </el-form-item>

          <el-form-item :label="t('trade.minNotional')">
            <el-input v-model="symbolForm.minNotional" />
          </el-form-item>

          <el-form-item :label="t('trade.listingTime')">
            <el-date-picker
              v-model="symbolListingTime"
              type="datetime"
              clearable
              class="full-width"
            />
          </el-form-item>

          <el-form-item :label="t('trade.tradingStartTime')">
            <el-date-picker
              v-model="symbolOpenTime"
              type="datetime"
              clearable
              :disabled-date="disableTradingStartDate"
              class="full-width"
            />
          </el-form-item>

          <el-form-item :label="t('trade.tradingEndTime')">
            <el-date-picker
              v-model="symbolCloseTime"
              type="datetime"
              clearable
              :disabled-date="disableTradingEndDate"
              class="full-width"
            />
          </el-form-item>

          <el-form-item :label="t('common.sort')">
            <el-input-number
              v-model="symbolForm.sort"
              :min="0"
              :precision="0"
              class="full-width"
            />
          </el-form-item>

          <el-form-item :label="t('common.remark')" class="wide">
            <el-input v-model="symbolForm.remark" type="textarea" :rows="3" />
          </el-form-item>
        </div>
      </el-form>

      <template #footer>
        <el-button @click="symbolVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button
          v-perm="symbolForm.id ? 'trade:symbol:update' : 'trade:symbol:add'"
          type="primary"
          :loading="submitLoading"
          @click="submitSymbol"
        >
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="secondsVisible"
      :title="t('trade.secondsConfig')"
      width="900px"
      class="seconds-config-dialog"
    >
      <el-form label-width="148px" class="dialog-form">
        <div class="form-grid two">
          <el-form-item :label="t('trade.tenantId')">
            <TenantSelect v-model="secondsForm.tenantId" include-system disabled />
          </el-form-item>
          <el-form-item :label="t('trade.symbolId')">
            <el-input-number v-model="secondsForm.symbolId" disabled class="full-width" />
          </el-form-item>
          <el-form-item :label="t('trade.durationSeconds')">
            <el-input-number
              v-model="secondsForm.durationSeconds"
              :min="1"
              :disabled="editingSecondsConfig"
              class="full-width"
            />
          </el-form-item>
          <el-form-item :label="t('trade.payoutRate')">
            <el-input v-model="secondsForm.payoutRate" />
          </el-form-item>
          <el-form-item :label="t('trade.secondsFeeRate')">
            <el-input v-model="secondsForm.feeRate" />
          </el-form-item>
          <el-form-item :label="t('trade.drawRule')">
            <el-input-number
              v-model="secondsForm.drawRule"
              :min="1"
              :max="2"
              class="full-width"
            />
          </el-form-item>
          <el-form-item :label="t('trade.quoteValidityMs')">
            <el-input-number v-model="secondsForm.quoteValidityMs" :min="1" class="full-width" />
          </el-form-item>
          <el-form-item :label="t('trade.settlementWindowMs')">
            <el-input-number v-model="secondsForm.settlementWindowMs" :min="1" class="full-width" />
          </el-form-item>
          <el-form-item :label="t('trade.startPriceSource')">
            <el-input
              v-model="secondsForm.startPriceSource"
              :placeholder="t('trade.priceSourcePlaceholder')"
            />
          </el-form-item>
          <el-form-item :label="t('trade.settlementPriceSource')">
            <el-input
              v-model="secondsForm.settlementPriceSource"
              :placeholder="t('trade.priceSourcePlaceholder')"
            />
          </el-form-item>
          <el-form-item :label="t('trade.settlementPriceAlgorithm')">
            <el-input v-model="secondsForm.settlementPriceAlgorithm" />
          </el-form-item>
          <el-form-item :label="t('trade.drawTolerance')">
            <el-input v-model="secondsForm.drawTolerance" />
          </el-form-item>
          <el-form-item :label="t('trade.maxExposureAmount')">
            <el-input v-model="secondsForm.maxExposureAmount" />
          </el-form-item>
          <el-form-item :label="t('trade.minStake')">
            <el-input v-model="secondsForm.minStake" />
          </el-form-item>
          <el-form-item :label="t('trade.maxStake')">
            <el-input v-model="secondsForm.maxStake" />
          </el-form-item>
          <el-form-item :label="t('trade.upEnabled')">
            <el-select v-model="secondsForm.upEnabled" class="full-width">
              <el-option
                v-for="item in enableStatusOptions"
                :key="item.value"
                :label="optionItemLabel(item)"
                :value="item.value"
              />
            </el-select>
          </el-form-item>
          <el-form-item :label="t('trade.downEnabled')">
            <el-select v-model="secondsForm.downEnabled" class="full-width">
              <el-option
                v-for="item in enableStatusOptions"
                :key="item.value"
                :label="optionItemLabel(item)"
                :value="item.value"
              />
            </el-select>
          </el-form-item>
        </div>
      </el-form>

      <div class="dialog-subheader">
        <strong>{{ secondsSymbol?.displaySymbol || secondsSymbol?.symbol || '-' }}</strong>
        <el-button v-perm="'trade:symbol:seconds-config'" size="small" @click="newSecondsConfig">
          <el-icon><Plus /></el-icon>
          {{ t('common.add') }}
        </el-button>
      </div>

      <el-table :data="secondsRows" size="small" border>
        <el-table-column prop="durationSeconds" :label="t('trade.durationSeconds')" width="150" />
        <el-table-column prop="payoutRate" :label="t('trade.payoutRate')" width="110" />
        <el-table-column prop="feeRate" :label="t('trade.secondsFeeRate')" width="110" />
        <el-table-column :label="t('trade.stakeRange')" min-width="150">
          <template #default="{ row }">
            {{ row.minStake }} - {{ row.maxStake }}
          </template>
        </el-table-column>
        <el-table-column :label="t('trade.directionStatus')" min-width="150">
          <template #default="{ row }">
            {{ t('trade.upEnabled') }}: {{ optionLabel('enableStatus', row.upEnabled) }} /
            {{ t('trade.downEnabled') }}: {{ optionLabel('enableStatus', row.downEnabled) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" align="center" width="90">
          <template #default="{ row }">
            <el-button
              v-perm="'trade:symbol:seconds-config'"
              link
              type="primary"
              @click="editSecondsConfig(row)"
            >
              {{ t('common.edit') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="secondsVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button
          v-perm="'trade:symbol:seconds-config'"
          type="primary"
          :loading="submitLoading"
          @click="submitSecondsConfig"
        >
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="spotVisible" :title="t('trade.spotConfig')" width="620px">
      <el-form label-width="116px" class="dialog-form">
        <div class="form-grid two">
          <el-form-item :label="t('trade.tenantId')">
            <TenantSelect v-model="spotForm.tenantId" include-system disabled />
          </el-form-item>

          <el-form-item :label="t('trade.symbolId')">
            <el-input-number
              v-model="spotForm.symbolId"
              :min="0"
              :precision="0"
              disabled
              class="full-width"
            />
          </el-form-item>

          <el-form-item :label="t('trade.makerFeeRate')">
            <el-input v-model="spotForm.makerFeeRate" />
          </el-form-item>

          <el-form-item :label="t('trade.takerFeeRate')">
            <el-input v-model="spotForm.takerFeeRate" />
          </el-form-item>

          <el-form-item :label="t('trade.buyEnabled')">
            <el-select v-model="spotForm.buyEnabled" class="full-width">
              <el-option
                v-for="item in enableStatusOptions"
                :key="item.value"
                :label="optionItemLabel(item)"
                :value="item.value"
              />
            </el-select>
          </el-form-item>

          <el-form-item :label="t('trade.sellEnabled')">
            <el-select v-model="spotForm.sellEnabled" class="full-width">
              <el-option
                v-for="item in enableStatusOptions"
                :key="item.value"
                :label="optionItemLabel(item)"
                :value="item.value"
              />
            </el-select>
          </el-form-item>
        </div>
      </el-form>

      <template #footer>
        <el-button @click="spotVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button
          v-perm="'trade:symbol:spot-config'"
          type="primary"
          :loading="submitLoading"
          @click="submitSpotConfig"
        >
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="contractVisible"
      :title="t('trade.contractConfig')"
      width="900px"
      class="contract-config-dialog"
    >
      <el-alert
        :title="contractConfigSummary"
        type="info"
        :closable="false"
        show-icon
      />
      <el-form label-width="126px" class="dialog-form">
        <div class="form-grid two">
          <el-form-item :label="t('trade.tenantId')">
            <TenantSelect v-model="contractForm.tenantId" include-system disabled />
          </el-form-item>

          <el-form-item :label="t('trade.symbolId')">
            <el-input-number
              v-model="contractForm.symbolId"
              :min="0"
              :precision="0"
              disabled
              class="full-width"
            />
          </el-form-item>

          <el-form-item :label="t('trade.contractSize')">
            <el-input v-model="contractForm.contractSize">
              <template #append>
                {{ contractSizeUnit }}
              </template>
            </el-input>
          </el-form-item>

          <el-form-item :label="t('option.multiplier')">
            <el-input v-model="contractForm.multiplier" />
          </el-form-item>

          <el-form-item :label="t('trade.maintenanceMarginRate')">
            <el-input v-model="contractForm.maintenanceMarginRate" />
          </el-form-item>

          <el-form-item :label="t('trade.initialMarginRate')">
            <el-input v-model="contractForm.initialMarginRate" />
          </el-form-item>

          <el-form-item :label="t('trade.makerFeeRate')">
            <el-input v-model="contractForm.makerFeeRate" />
          </el-form-item>

          <el-form-item :label="t('trade.takerFeeRate')">
            <el-input v-model="contractForm.takerFeeRate" />
          </el-form-item>

          <el-form-item v-if="isPerpetualContract" :label="t('trade.fundingIntervalMinutes')">
            <el-input-number
              v-model="contractForm.fundingIntervalMinutes"
              :min="0"
              :precision="0"
              class="full-width"
            />
          </el-form-item>

          <el-form-item v-if="isPerpetualContract" :label="t('trade.fundingRateCap')">
            <el-input v-model="contractForm.fundingRateCap" />
          </el-form-item>

          <el-form-item v-if="isPerpetualContract" :label="t('trade.fundingRateFloor')">
            <el-input v-model="contractForm.fundingRateFloor" />
          </el-form-item>

          <el-form-item v-if="isPerpetualContract" :label="t('trade.fundingRateSource')">
            <el-input
              v-model="contractForm.fundingRateSource"
              :placeholder="t('trade.fundingRateSourcePlaceholder')"
            />
          </el-form-item>

          <el-form-item :label="t('trade.indexSymbol')">
            <el-input v-model="contractForm.indexSymbol" />
          </el-form-item>

          <el-form-item :label="t('trade.markPriceSource')">
            <el-input
              v-model="contractForm.markPriceSource"
              :placeholder="t('trade.priceSourcePlaceholder')"
            />
          </el-form-item>

          <el-form-item v-if="isDeliveryContract" :label="t('trade.settlementPriceSource')">
            <el-input
              v-model="contractForm.settlementPriceSource"
              :placeholder="t('trade.priceSourcePlaceholder')"
            />
          </el-form-item>

          <el-form-item v-if="isDeliveryContract" :label="t('option.deliverTime')">
            <el-date-picker
              v-model="contractDeliveryTime"
              type="datetime"
              clearable
              :disabled-date="disableDeliveryDate"
              class="full-width"
            />
          </el-form-item>

          <el-form-item v-if="isDeliveryContract" :label="t('trade.openCutoffTime')">
            <el-date-picker
              v-model="contractOpenCutoffTime"
              type="datetime"
              clearable
              :disabled-date="disableOpenCutoffDate"
              class="full-width"
            />
          </el-form-item>

          <el-form-item v-if="isDeliveryContract" :label="t('trade.matchingStopTime')">
            <el-date-picker
              v-model="contractMatchingStopTime"
              type="datetime"
              clearable
              :disabled-date="disableMatchingStopDate"
              class="full-width"
            />
          </el-form-item>

          <el-form-item v-if="isDeliveryContract" :label="t('trade.settlementWindowSeconds')">
            <el-input-number
              v-model="contractForm.settlementWindowSeconds"
              :min="0"
              :precision="0"
              class="full-width"
            />
          </el-form-item>

          <el-form-item v-if="isDeliveryContract" :label="t('trade.settlementPriceAlgorithm')">
            <el-input v-model="contractForm.settlementPriceAlgorithm" />
          </el-form-item>

          <el-form-item v-if="isDeliveryContract" :label="t('trade.deliveryFeeRate')">
            <el-input v-model="contractForm.deliveryFeeRate" />
          </el-form-item>

          <el-form-item :label="t('trade.liquidationFeeRate')">
            <el-input v-model="contractForm.liquidationFeeRate" />
          </el-form-item>

          <el-form-item :label="t('trade.supportCross')">
            <el-select v-model="contractForm.supportCross" class="full-width">
              <el-option :label="t('trade.supported')" :value="1" />
              <el-option :label="t('trade.notSupported')" :value="0" />
            </el-select>
          </el-form-item>

          <el-form-item :label="t('trade.supportIsolated')">
            <el-select v-model="contractForm.supportIsolated" class="full-width">
              <el-option :label="t('trade.supported')" :value="1" />
              <el-option :label="t('trade.notSupported')" :value="0" />
            </el-select>
          </el-form-item>

          <el-form-item :label="t('trade.openLongEnabled')">
            <el-select v-model="contractForm.openLongEnabled" class="full-width">
              <el-option
                v-for="item in enableStatusOptions"
                :key="item.value"
                :label="optionItemLabel(item)"
                :value="item.value"
              />
            </el-select>
          </el-form-item>

          <el-form-item :label="t('trade.openShortEnabled')">
            <el-select v-model="contractForm.openShortEnabled" class="full-width">
              <el-option
                v-for="item in enableStatusOptions"
                :key="item.value"
                :label="optionItemLabel(item)"
                :value="item.value"
              />
            </el-select>
          </el-form-item>

          <el-form-item :label="t('trade.closeLongEnabled')">
            <el-select v-model="contractForm.closeLongEnabled" class="full-width">
              <el-option
                v-for="item in enableStatusOptions"
                :key="item.value"
                :label="optionItemLabel(item)"
                :value="item.value"
              />
            </el-select>
          </el-form-item>

          <el-form-item :label="t('trade.closeShortEnabled')">
            <el-select v-model="contractForm.closeShortEnabled" class="full-width">
              <el-option
                v-for="item in enableStatusOptions"
                :key="item.value"
                :label="optionItemLabel(item)"
                :value="item.value"
              />
            </el-select>
          </el-form-item>
        </div>
      </el-form>

      <template #footer>
        <el-button @click="contractVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button
          v-perm="'trade:symbol:contract-config'"
          type="primary"
          :loading="submitLoading"
          @click="submitContractConfig"
        >
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="leverageVisible" :title="t('trade.leverageConfig')" width="820px">
      <el-form label-width="126px" class="dialog-form">
        <div class="form-grid two">
          <el-form-item :label="t('trade.tenantId')">
            <TenantSelect v-model="leverageForm.tenantId" include-system disabled />
          </el-form-item>

          <el-form-item :label="t('trade.symbolId')">
            <el-input-number
              v-model="leverageForm.symbolId"
              :min="0"
              :precision="0"
              disabled
              class="full-width"
            />
          </el-form-item>

          <el-form-item :label="t('trade.marginMode')">
            <el-select v-model="leverageForm.marginMode" class="full-width">
              <el-option
                v-for="item in availableLeverageMarginModeOptions"
                :key="item.value"
                :label="optionItemLabel(item)"
                :value="item.value"
              />
            </el-select>
          </el-form-item>

          <el-form-item :label="t('trade.defaultLeverage')">
            <el-select v-model="leverageForm.defaultLeverage" class="full-width">
              <el-option
                v-for="item in defaultLeverageOptions"
                :key="item.value"
                :label="optionItemLabel(item)"
                :value="item.value"
              />
            </el-select>
          </el-form-item>

          <el-form-item :label="t('trade.status')">
            <el-select v-model="leverageForm.enabled" class="full-width">
              <el-option
                v-for="item in enableStatusOptions"
                :key="item.value"
                :label="optionItemLabel(item)"
                :value="item.value"
              />
            </el-select>
          </el-form-item>

          <el-form-item :label="t('common.sort')">
            <el-input-number
              v-model="leverageForm.sort"
              :min="0"
              :precision="0"
              class="full-width"
            />
          </el-form-item>

          <el-form-item :label="t('trade.leverageValues')" class="wide">
            <el-select
              v-model="leverageForm.leverageValues"
              multiple
              filterable
              class="full-width"
              @change="handleLeverageValuesChange"
            >
              <el-option
                v-for="item in leverageValueFormOptions"
                :key="item.value"
                :label="optionItemLabel(item)"
                :value="item.value"
              />
            </el-select>
          </el-form-item>

          <el-form-item :label="t('common.remark')" class="wide">
            <el-input v-model="leverageForm.remark" type="textarea" :rows="2" />
          </el-form-item>
        </div>
      </el-form>

      <div class="dialog-subheader">
        <strong>{{ leverageSymbol?.displaySymbol || leverageSymbol?.symbol || '-' }}</strong>
        <el-button
          v-perm="'trade:symbol:leverage-config:update'"
          size="small"
          :disabled="!canAddLeverageConfig"
          @click="newLeverageConfig()"
        >
          <el-icon><Plus /></el-icon>
          {{ t('common.add') }}
        </el-button>
      </div>

      <el-table
        :data="leverageGroups"
        size="small"
        border
        class="leverage-table"
      >
        <el-table-column :label="t('trade.marginMode')" width="130">
          <template #default="{ row }">
            {{ optionLabel('marginMode', row.marginMode) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('trade.leverageValues')" min-width="190">
          <template #default="{ row }">
            {{ (row.leverageValues || []).join(', ') || '-' }}
          </template>
        </el-table-column>
        <el-table-column :label="t('trade.defaultLeverage')" width="130">
          <template #default="{ row }">
            {{ row.defaultLeverage }}X
          </template>
        </el-table-column>
        <el-table-column :label="t('trade.maxLeverage')" width="120">
          <template #default="{ row }">
            {{ row.leverageValues[row.leverageValues.length - 1] || 1 }}X
          </template>
        </el-table-column>
        <el-table-column :label="t('trade.status')" width="110">
          <template #default="{ row }">
            <el-tag size="small" :type="row.enabled === 1 ? 'success' : 'info'" effect="light">
              {{ optionLabel('enableStatus', row.enabled) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" align="center" width="90">
          <template #default="{ row }">
            <el-button
              v-perm="'trade:symbol:leverage-config:update'"
              link
              type="primary"
              @click="editLeverageConfig(row)"
            >
              {{ t('common.edit') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <template #footer>
        <el-button @click="leverageVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button
          v-perm="'trade:symbol:leverage-config:update'"
          type="primary"
          :loading="submitLoading"
          @click="submitLeverageConfig"
        >
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" :title="t('option.detail')" width="860px">
      <el-descriptions v-if="detailData" :column="3" border>
        <el-descriptions-item label="ID">
          {{ detailData.id }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.tenantId')">
          {{ detailData.tenantId }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.status')">
          <el-tag size="small" :type="symbolStatusTagType(detailData.status)" effect="light">
            {{ optionLabel('symbolStatus', detailData.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.symbol')">
          {{ detailData.symbol || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.displaySymbol')">
          {{ detailData.displaySymbol || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.productType')">
          {{ optionLabel('productType', detailData.productType) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.baseAsset')">
          {{ detailData.baseAsset || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.quoteAsset')">
          {{ detailData.quoteAsset || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.settleAsset')">
          {{ detailData.settleAsset || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.contractType')">
          {{ optionLabel('contractType', detailData.contractType) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.contractValueType')">
          {{ optionLabel('contractValueType', detailData.contractValueType) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.marginAsset')">
          {{ detailData.marginAsset || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.priceScale')">
          {{ detailData.priceScale }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.qtyScale')">
          {{ detailData.qtyScale }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.minPrice')">
          {{ detailData.minPrice || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.maxPrice')">
          {{ detailData.maxPrice || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.priceTick')">
          {{ detailData.priceTick || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.minQty')">
          {{ detailData.minQty || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.maxQty')">
          {{ detailData.maxQty || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.qtyStep')">
          {{ detailData.qtyStep || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.minNotional')">
          {{ detailData.minNotional || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.maxNotional')">
          {{ detailData.maxNotional || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.sort')">
          {{ detailData.sort || 0 }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.listingTime')">
          {{ formatDate(detailData.listingTime || 0) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.tradingStartTime')">
          {{ formatDate(detailData.tradingStartTime || 0) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.tradingEndTime')">
          {{ formatDate(detailData.tradingEndTime || 0) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.createTimes')">
          {{ formatDate(detailData.createTimes || 0) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.updateTimes')">
          {{ formatDate(detailData.updateTimes || 0) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.remark')" :span="2">
          {{ detailData.remark || '-' }}
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables'
import {
  tradeService,
  type OptionGroup,
  type OptionItem,
  type TradeSymbol,
  type TradeSymbolContract,
  type TradeSymbolLeverageConfig,
  type TradeSymbolSeconds,
} from '@/services'
import { formatDate } from '@/utils'
import { findOptionGroup, getOptionLabel } from '@/utils/options'
import TenantSelect from '@/components/TenantSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

type TagType = '' | 'success' | 'warning' | 'info' | 'danger'
type DatePickerValue = Date | string | number | null | undefined

interface SymbolQuery {
  tenantId: number | undefined
  categoryType: number | undefined
  productType: number | undefined
  keyword: string
  status: number | undefined
}

interface SymbolForm {
  id: number
  tenantId: number
  categoryType: number
  symbol: string
  displaySymbol: string
  productType: number
  baseAsset: string
  quoteAsset: string
  settleAsset: string
  contractType: number
  contractValueType: number
  marginAsset: string
  status: number
  priceScale: number
  qtyScale: number
  minPrice: string
  maxPrice: string
  priceTick: string
  minQty: string
  maxQty: string
  qtyStep: string
  minNotional: string
  maxNotional: string
  listingTime: number
  tradingStartTime: number
  tradingEndTime: number
  sort: number
  remark: string
}

interface SpotForm {
  tenantId: number
  symbolId: number
  makerFeeRate: string
  takerFeeRate: string
  buyEnabled: number
  sellEnabled: number
}

interface ContractForm {
  tenantId: number
  symbolId: number
  contractSize: string
  multiplier: string
  maintenanceMarginRate: string
  initialMarginRate: string
  makerFeeRate: string
  takerFeeRate: string
  fundingIntervalMinutes: number
  deliveryTime: number
  supportCross: number
  supportIsolated: number
  fundingRateCap: string
  fundingRateFloor: string
  fundingRateSource: string
  indexSymbol: string
  markPriceSource: string
  settlementPriceSource: string
  openCutoffTime: number
  matchingStopTime: number
  settlementWindowSeconds: number
  settlementPriceAlgorithm: string
  deliveryFeeRate: string
  liquidationFeeRate: string
  openLongEnabled: number
  openShortEnabled: number
  closeLongEnabled: number
  closeShortEnabled: number
}

interface SecondsForm {
  tenantId: number
  symbolId: number
  durationSeconds: number
  payoutRate: string
  feeRate: string
  drawRule: number
  startPriceSource: string
  settlementPriceSource: string
  quoteValidityMs: number
  settlementWindowMs: number
  settlementPriceAlgorithm: string
  drawTolerance: string
  maxExposureAmount: string
  minStake: string
  maxStake: string
  upEnabled: number
  downEnabled: number
}

interface LeverageForm {
  tenantId: number
  symbolId: number
  marginMode: number
  leverageValues: number[]
  defaultLeverage: number
  enabled: number
  sort: number
  remark: string
}

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

const loading = ref(false)
const submitLoading = ref(false)
const rows = ref<TradeSymbol[]>([])
const detailVisible = ref(false)
const detailData = ref<TradeSymbol | null>(null)
const symbolVisible = ref(false)
const spotVisible = ref(false)
const contractVisible = ref(false)
const secondsVisible = ref(false)
const leverageVisible = ref(false)
const optionGroups = ref<OptionGroup[]>([])
const leverageSymbol = ref<TradeSymbol | null>(null)
const leverageContractConfig = ref<TradeSymbolContract | null>(null)
const contractSymbol = ref<TradeSymbol | null>(null)
const secondsSymbol = ref<TradeSymbol | null>(null)
const secondsRows = ref<TradeSymbolSeconds[]>([])
const editingSecondsConfig = ref(false)
const leverageRows = ref<TradeSymbolLeverageConfig[]>([])

const productTypeFallbackOptions: OptionItem[] = [
  { value: 0, code: 'PRODUCT_TYPE_UNKNOWN' },
  { value: 1, code: 'PRODUCT_TYPE_SPOT' },
  { value: 2, code: 'PRODUCT_TYPE_DERIVATIVE' },
  { value: 3, code: 'PRODUCT_TYPE_SECONDS' },
]

const contractTypeFallbackOptions: OptionItem[] = [
  { value: 0, code: 'CONTRACT_TYPE_NOT_APPLICABLE' },
  { value: 1, code: 'CONTRACT_TYPE_PERPETUAL' },
  { value: 2, code: 'CONTRACT_TYPE_DELIVERY' },
]

const contractValueTypeFallbackOptions: OptionItem[] = [
  { value: 0, code: 'CONTRACT_VALUE_TYPE_NOT_APPLICABLE' },
  { value: 1, code: 'CONTRACT_VALUE_TYPE_LINEAR' },
  { value: 2, code: 'CONTRACT_VALUE_TYPE_INVERSE' },
]

const symbolStatusFallbackOptions: OptionItem[] = [
  { value: 0, code: 'SYMBOL_STATUS_UNKNOWN' },
  { value: 1, code: 'SYMBOL_STATUS_ENABLED' },
  { value: 2, code: 'SYMBOL_STATUS_DISABLED' },
  { value: 3, code: 'SYMBOL_STATUS_CLOSE_ONLY' },
]

const enableStatusFallbackOptions: OptionItem[] = [
  { value: 0, code: 'ENABLE_STATUS_DISABLED' },
  { value: 1, code: 'ENABLE_STATUS_ENABLED' },
]

const marginModeFallbackOptions: OptionItem[] = [
  { value: 0, code: 'MARGIN_MODE_UNKNOWN' },
  { value: 1, code: 'MARGIN_MODE_CROSS' },
  { value: 2, code: 'MARGIN_MODE_ISOLATED' },
]

const leverageValueFallbackOptions: OptionItem[] = [
  { value: 1, code: 'LEVERAGE_VALUE_1X' },
  { value: 2, code: 'LEVERAGE_VALUE_2X' },
  { value: 5, code: 'LEVERAGE_VALUE_5X' },
  { value: 10, code: 'LEVERAGE_VALUE_10X' },
  { value: 20, code: 'LEVERAGE_VALUE_20X' },
  { value: 50, code: 'LEVERAGE_VALUE_50X' },
  { value: 75, code: 'LEVERAGE_VALUE_75X' },
  { value: 100, code: 'LEVERAGE_VALUE_100X' },
  { value: 125, code: 'LEVERAGE_VALUE_125X' },
]

const query = reactive<SymbolQuery>({
  tenantId: undefined,
  categoryType: undefined,
  productType: undefined,
  keyword: '',
  status: undefined,
})

const getDefaultSymbolForm = (): SymbolForm => ({
  id: 0,
  tenantId: 0,
  categoryType: 2,
  symbol: '',
  displaySymbol: '',
  productType: 1,
  baseAsset: '',
  quoteAsset: '',
  settleAsset: '',
  contractType: 0,
  contractValueType: 0,
  marginAsset: '',
  status: 1,
  priceScale: 2,
  qtyScale: 4,
  minPrice: '',
  maxPrice: '',
  priceTick: '',
  minQty: '',
  maxQty: '',
  qtyStep: '',
  minNotional: '',
  maxNotional: '',
  listingTime: 0,
  tradingStartTime: 0,
  tradingEndTime: 0,
  sort: 0,
  remark: '',
})

const getDefaultSpotForm = (): SpotForm => ({
  tenantId: 0,
  symbolId: 0,
  makerFeeRate: '',
  takerFeeRate: '',
  buyEnabled: 1,
  sellEnabled: 1,
})

const getDefaultContractForm = (row?: TradeSymbol | null): ContractForm => ({
  tenantId: 0,
  symbolId: 0,
  contractSize: row?.contractValueType === 2 ? '100' : '1',
  multiplier: '1',
  maintenanceMarginRate: '0.005',
  initialMarginRate: '0.01',
  makerFeeRate: '0.0002',
  takerFeeRate: '0.0005',
  fundingIntervalMinutes: row?.contractType === 1 ? 480 : 0,
  deliveryTime: 0,
  supportCross: 0,
  supportIsolated: 1,
  fundingRateCap: row?.contractType === 1 ? '0.003' : '0',
  fundingRateFloor: row?.contractType === 1 ? '-0.003' : '0',
  fundingRateSource: row?.contractType === 1 ? 'funding-v1' : '',
  indexSymbol: row?.symbol || '',
  markPriceSource: row?.symbol || '',
  settlementPriceSource: row?.contractType === 2 ? row.symbol : '',
  openCutoffTime: 0,
  matchingStopTime: 0,
  settlementWindowSeconds: row?.contractType === 2 ? 60 : 0,
  settlementPriceAlgorithm: 'last-v1',
  deliveryFeeRate: row?.contractType === 2 ? '0.0005' : '0',
  liquidationFeeRate: '0.005',
  openLongEnabled: 1,
  openShortEnabled: 1,
  closeLongEnabled: 1,
  closeShortEnabled: 1,
})

const getDefaultSecondsForm = (): SecondsForm => ({
  tenantId: 0,
  symbolId: 0,
  durationSeconds: 60,
  payoutRate: '',
  feeRate: '0',
  drawRule: 1,
  startPriceSource: '',
  settlementPriceSource: '',
  quoteValidityMs: 3000,
  settlementWindowMs: 1000,
  settlementPriceAlgorithm: 'last-v1',
  drawTolerance: '0',
  maxExposureAmount: '0',
  minStake: '',
  maxStake: '',
  upEnabled: 1,
  downEnabled: 1,
})

const defaultLeverageValues = (
  maxLeverage: number,
  options: OptionItem[] = leverageValueFallbackOptions,
) => {
  const maxValue = Math.max(1, Number(maxLeverage || 1))
  const values = options.map((item) => Number(item.value)).filter((value) => value <= maxValue)
  return values.length ? values : [1]
}

const getDefaultLeverageForm = (
  row?: TradeSymbol | null,
  marginMode = 1,
  leverageOptions: OptionItem[] = leverageValueFallbackOptions,
): LeverageForm => {
  const values = defaultLeverageValues(125, leverageOptions)
  return {
    tenantId: row?.tenantId || 0,
    symbolId: row?.id || 0,
    marginMode,
    leverageValues: values,
    defaultLeverage: values[0] || 1,
    enabled: 1,
    sort: marginMode,
    remark: '',
  }
}

const symbolForm = reactive<SymbolForm>(getDefaultSymbolForm())
const spotForm = reactive<SpotForm>(getDefaultSpotForm())
const contractForm = reactive<ContractForm>(getDefaultContractForm())
const secondsForm = reactive<SecondsForm>(getDefaultSecondsForm())
const leverageForm = reactive<LeverageForm>(getDefaultLeverageForm())

const optionGroupWithFallback = (key: string, fallback: OptionItem[]) =>
  computed(() => {
    const options = findOptionGroup(optionGroups.value, key)
    if (!options.length) return fallback

    // Keep the server-provided labels, but retain supported enum values when
    // an older API response returns only a partial option group.
    const merged = new Map(fallback.map((item) => [Number(item.value), item]))
    for (const item of options) {
      merged.set(Number(item.value), item)
    }
    return Array.from(merged.values()).sort((left, right) => left.value - right.value)
  })

const withoutUnknown = (options: OptionItem[]) => options.filter((item) => item.value !== 0)

const productTypeOptions = optionGroupWithFallback('productType', productTypeFallbackOptions)
const categoryTypeOptions = computed(() =>
  findOptionGroup(optionGroups.value, 'categoryType').filter((item) => Number(item.value) > 0),
)
const contractTypeOptions = optionGroupWithFallback('contractType', contractTypeFallbackOptions)
const contractValueTypeOptions = optionGroupWithFallback(
  'contractValueType',
  contractValueTypeFallbackOptions,
)
const symbolStatusOptions = optionGroupWithFallback('symbolStatus', symbolStatusFallbackOptions)
const enableStatusOptions = optionGroupWithFallback('enableStatus', enableStatusFallbackOptions)
const marginModeOptions = optionGroupWithFallback('marginMode', marginModeFallbackOptions)
const leverageValueOptions = optionGroupWithFallback('leverageValue', leverageValueFallbackOptions)
const productTypeFormOptions = computed(() => withoutUnknown(productTypeOptions.value))
const categoryTypeFormOptions = categoryTypeOptions
const isCryptoCategory = computed(() => symbolForm.categoryType === 2)
const isFutureCategory = computed(() => symbolForm.categoryType === 4)
const isDerivativeProduct = computed(() => symbolForm.productType === 2)
const contractTypeFormOptions = computed(() =>
  isDerivativeProduct.value
    ? withoutUnknown(contractTypeOptions.value)
    : contractTypeOptions.value.filter((item) => item.value === 0),
)
const contractValueTypeFormOptions = computed(() =>
  isDerivativeProduct.value
    ? withoutUnknown(contractValueTypeOptions.value)
    : contractValueTypeOptions.value.filter((item) => item.value === 0),
)
const symbolStatusFormOptions = computed(() => withoutUnknown(symbolStatusOptions.value))
const marginModeFormOptions = computed(() => withoutUnknown(marginModeOptions.value))
const leverageValueFormOptions = computed(() => {
  return leverageValueOptions.value.filter((item) => Number(item.value) > 0)
})
const defaultLeverageOptions = computed(() => {
  const selected = new Set(leverageForm.leverageValues.map(Number))
  return leverageValueFormOptions.value.filter((item) => selected.has(Number(item.value)))
})

const optionValueByCode = (key: string, code: string, fallback: number) => {
  const option = findOptionGroup(optionGroups.value, key).find((item) => item.code === code)
  return Number(option?.value || fallback)
}
const spotMarketValue = computed(() => optionValueByCode('productType', 'PRODUCT_TYPE_SPOT', 1))
const derivativeMarketValue = computed(() =>
  optionValueByCode('productType', 'PRODUCT_TYPE_DERIVATIVE', 2),
)
const perpetualContractTypeValue = computed(() =>
  optionValueByCode('contractType', 'CONTRACT_TYPE_PERPETUAL', 1),
)
const deliveryContractTypeValue = computed(() =>
  optionValueByCode('contractType', 'CONTRACT_TYPE_DELIVERY', 2),
)
const linearContractValueType = computed(() =>
  optionValueByCode('contractValueType', 'CONTRACT_VALUE_TYPE_LINEAR', 1),
)
const isPerpetualContract = computed(
  () => contractSymbol.value?.contractType === perpetualContractTypeValue.value,
)
const isDeliveryContract = computed(
  () => contractSymbol.value?.contractType === deliveryContractTypeValue.value,
)
const contractSizeUnit = computed(() =>
  contractSymbol.value?.contractValueType === linearContractValueType.value
    ? `${contractSymbol.value?.baseAsset || '-'}/${t('trade.contractQuantityUnit')}`
    : `${contractSymbol.value?.quoteAsset || 'USD'}/${t('trade.contractQuantityUnit')}`,
)
const contractConfigSummary = computed(() => {
  const row = contractSymbol.value
  if (!row) return '-'
  return `${row.displaySymbol || row.symbol} · ${optionLabel('contractType', row.contractType)} · ${optionLabel('contractValueType', row.contractValueType)} · ${t('trade.marginAsset')}: ${row.marginAsset || '-'} · ${t('trade.settleAsset')}: ${row.settleAsset || '-'}`
})
const contractMarketValues = computed(() => {
  const codes = new Set(['PRODUCT_TYPE_DERIVATIVE'])
  const values = findOptionGroup(optionGroups.value, 'productType')
    .filter((item) => codes.has(item.code))
    .map((item) => Number(item.value))
  return values.length ? values : [2]
})
const crossMarginModeValue = computed(() => optionValueByCode('marginMode', 'MARGIN_MODE_CROSS', 1))
const isolatedMarginModeValue = computed(() =>
  optionValueByCode('marginMode', 'MARGIN_MODE_ISOLATED', 2),
)
const enabledStatusValue = computed(() =>
  optionValueByCode('enableStatus', 'ENABLE_STATUS_ENABLED', 1),
)

const fixedProductTypeValue = computed(() =>
  isFutureCategory.value ? derivativeMarketValue.value : spotMarketValue.value,
)

watch(
  () => symbolForm.categoryType,
  (categoryType) => {
    if (categoryType === 4) {
      symbolForm.productType = derivativeMarketValue.value
      if (symbolForm.contractType === 0) symbolForm.contractType = deliveryContractTypeValue.value
      if (symbolForm.contractValueType === 0) {
        symbolForm.contractValueType = linearContractValueType.value
      }
      return
    }
    if (categoryType !== 2) {
      symbolForm.productType = spotMarketValue.value
      symbolForm.contractType = 0
      symbolForm.contractValueType = 0
      symbolForm.marginAsset = ''
    }
  },
)

watch(
  () => symbolForm.productType,
  (productType) => {
    if (productType !== 2) {
      symbolForm.contractType = 0
      symbolForm.contractValueType = 0
      symbolForm.marginAsset = ''
      return
    }
    if (symbolForm.contractType === 0) symbolForm.contractType = 1
    if (symbolForm.contractValueType === 0) symbolForm.contractValueType = 1
  },
)

const timestampToDate = (timestamp?: number) => {
  if (!timestamp) return null
  return new Date(timestamp < 1e12 ? timestamp * 1000 : timestamp)
}

const dateToTimestamp = (value: DatePickerValue) => {
  if (!value) return 0
  const time =
    typeof value === 'number'
      ? value < 1e12
        ? value * 1000
        : value
      : value instanceof Date
        ? value.getTime()
        : new Date(value).getTime()
  return Number.isNaN(time) ? 0 : Math.floor(time)
}

const MINUTE_MILLISECONDS = 60 * 1000
const DEFAULT_START_DELAY = 10 * MINUTE_MILLISECONDS
const DEFAULT_TRADING_DURATION = 30 * MINUTE_MILLISECONDS
const DEFAULT_OPEN_CUTOFF_ADVANCE = 30 * MINUTE_MILLISECONDS
const DEFAULT_DELIVERY_DELAY = 10 * MINUTE_MILLISECONDS

const dayStartMillis = (timestamp: number) => {
  const date = timestampToDate(timestamp) || new Date(0)
  date.setHours(0, 0, 0, 0)
  return date.getTime()
}

const disabledBefore = (date: Date, timestamp: number) =>
  timestamp > 0 && date.getTime() < dayStartMillis(timestamp)

const disabledAfter = (date: Date, timestamp: number) => {
  if (timestamp <= 0) return false
  const end = timestampToDate(timestamp) || new Date(0)
  end.setHours(23, 59, 59, 999)
  return date.getTime() > end.getTime()
}

const disableTradingStartDate = (date: Date) => disabledBefore(date, symbolForm.listingTime)
const disableTradingEndDate = (date: Date) => disabledBefore(date, symbolForm.tradingStartTime)
const disableOpenCutoffDate = (date: Date) =>
  disabledBefore(date, contractSymbol.value?.tradingStartTime || 0) ||
  disabledAfter(date, contractForm.matchingStopTime || contractForm.deliveryTime)
const disableMatchingStopDate = (date: Date) =>
  disabledBefore(
    date,
    contractForm.openCutoffTime || contractSymbol.value?.tradingStartTime || 0,
  ) || disabledAfter(date, contractForm.deliveryTime)
const disableDeliveryDate = (date: Date) =>
  disabledBefore(date, contractForm.matchingStopTime || contractSymbol.value?.tradingEndTime || 0)

const symbolOpenTime = computed({
  get: () => timestampToDate(symbolForm.tradingStartTime),
  set: (value: DatePickerValue) => {
    const selected = dateToTimestamp(value)
    if (!selected) {
      symbolForm.tradingStartTime = 0
      return
    }
    symbolForm.tradingStartTime = Math.max(
      selected,
      symbolForm.listingTime ? symbolForm.listingTime + MINUTE_MILLISECONDS : selected,
    )
    if (symbolForm.tradingEndTime <= symbolForm.tradingStartTime) {
      symbolForm.tradingEndTime = symbolForm.tradingStartTime + DEFAULT_TRADING_DURATION
    }
  },
})

const symbolListingTime = computed({
  get: () => timestampToDate(symbolForm.listingTime),
  set: (value: DatePickerValue) => {
    symbolForm.listingTime = dateToTimestamp(value)
    if (!symbolForm.listingTime) return
    if (symbolForm.tradingStartTime <= symbolForm.listingTime) {
      symbolForm.tradingStartTime = symbolForm.listingTime + DEFAULT_START_DELAY
    }
    if (symbolForm.tradingEndTime <= symbolForm.tradingStartTime) {
      symbolForm.tradingEndTime = symbolForm.tradingStartTime + DEFAULT_TRADING_DURATION
    }
  },
})

const symbolCloseTime = computed({
  get: () => timestampToDate(symbolForm.tradingEndTime),
  set: (value: DatePickerValue) => {
    const selected = dateToTimestamp(value)
    if (!selected) {
      symbolForm.tradingEndTime = 0
      return
    }
    symbolForm.tradingEndTime = Math.max(
      selected,
      symbolForm.tradingStartTime ? symbolForm.tradingStartTime + MINUTE_MILLISECONDS : selected,
    )
  },
})

const contractDeliveryTime = computed({
  get: () => timestampToDate(contractForm.deliveryTime),
  set: (value: DatePickerValue) => {
    const selected = dateToTimestamp(value)
    contractForm.deliveryTime = selected
      ? Math.max(
          selected,
          contractForm.matchingStopTime
            ? contractForm.matchingStopTime + MINUTE_MILLISECONDS
            : selected,
        )
      : 0
  },
})

const contractOpenCutoffTime = computed({
  get: () => timestampToDate(contractForm.openCutoffTime),
  set: (value: DatePickerValue) => {
    const selected = dateToTimestamp(value)
    const tradingStart = contractSymbol.value?.tradingStartTime || 0
    contractForm.openCutoffTime = selected
      ? Math.max(selected, tradingStart ? tradingStart + MINUTE_MILLISECONDS : selected)
      : 0
    if (
      contractForm.openCutoffTime &&
      contractForm.matchingStopTime <= contractForm.openCutoffTime
    ) {
      contractForm.matchingStopTime = contractForm.openCutoffTime + DEFAULT_DELIVERY_DELAY
    }
    if (
      contractForm.matchingStopTime &&
      contractForm.deliveryTime <= contractForm.matchingStopTime
    ) {
      contractForm.deliveryTime = contractForm.matchingStopTime + DEFAULT_DELIVERY_DELAY
    }
  },
})

const contractMatchingStopTime = computed({
  get: () => timestampToDate(contractForm.matchingStopTime),
  set: (value: DatePickerValue) => {
    const selected = dateToTimestamp(value)
    const minimum = contractForm.openCutoffTime || contractSymbol.value?.tradingStartTime || 0
    contractForm.matchingStopTime = selected
      ? Math.max(selected, minimum ? minimum + MINUTE_MILLISECONDS : selected)
      : 0
    if (
      contractForm.matchingStopTime &&
      contractForm.deliveryTime <= contractForm.matchingStopTime
    ) {
      contractForm.deliveryTime = contractForm.matchingStopTime + DEFAULT_DELIVERY_DELAY
    }
  },
})

const applyDefaultDeliveryTimeline = () => {
  const row = contractSymbol.value
  if (!row?.tradingEndTime) return
  const matchingStopTime = row.tradingEndTime
  const openCutoffTime = Math.max(
    row.tradingStartTime ? row.tradingStartTime + MINUTE_MILLISECONDS : 0,
    matchingStopTime - DEFAULT_OPEN_CUTOFF_ADVANCE,
  )
  if (!contractForm.matchingStopTime) contractForm.matchingStopTime = matchingStopTime
  if (!contractForm.openCutoffTime) contractForm.openCutoffTime = openCutoffTime
  if (!contractForm.deliveryTime) {
    contractForm.deliveryTime = matchingStopTime + DEFAULT_DELIVERY_DELAY
  }
}

const optionItemLabel = (item: OptionItem) => getOptionLabel(t, item.code, item.value)

const optionLabel = (key: string, value?: number | string) => {
  const fallbackMap: Record<string, OptionItem[]> = {
    productType: productTypeFallbackOptions,
    contractType: contractTypeFallbackOptions,
    contractValueType: contractValueTypeFallbackOptions,
    symbolStatus: symbolStatusFallbackOptions,
    enableStatus: enableStatusFallbackOptions,
    marginMode: marginModeFallbackOptions,
    leverageValue: leverageValueFallbackOptions,
  }
  const option =
    findOptionGroup(optionGroups.value, key).find((item) => String(item.value) === String(value)) ||
    fallbackMap[key]?.find((item) => String(item.value) === String(value))
  return option ? optionItemLabel(option) : '-'
}

const symbolStatusTagType = (status?: number): TagType => {
  switch (status) {
    case 1:
      return 'success'
    case 2:
      return 'danger'
    case 3:
      return 'warning'
    default:
      return 'info'
  }
}

const isSpotMarket = (row: TradeSymbol) => row.productType === spotMarketValue.value

const isContractMarket = (row: TradeSymbol) => contractMarketValues.value.includes(row.productType)

const applyLeverageForm = (config: LeverageForm | TradeSymbolLeverageConfig) => {
  Object.assign(leverageForm, {
    tenantId: config.tenantId,
    symbolId: config.symbolId,
    marginMode: config.marginMode,
    leverageValues: [...config.leverageValues],
    defaultLeverage: config.defaultLeverage,
    enabled: config.enabled,
    sort: config.sort,
    remark: config.remark || '',
  })
}

const loadOptions = async () => {
  try {
    optionGroups.value = (await tradeService.getOptions()).data || []
  } catch (error) {
    console.error('load trade options failed', error)
  }
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await tradeService.listSymbols({
      ...query,
      cursor: pagination.cursor,
      limit: pagination.limit,
    })
    rows.value = res?.data || []
    updateFromResponse(res)
  } finally {
    loading.value = false
  }
}

const resetQuery = () => {
  query.tenantId = undefined
  query.categoryType = undefined
  query.productType = undefined
  query.keyword = ''
  query.status = undefined
  resetAndLoad(loadList)
}

const showDetail = async (row: TradeSymbol) => {
  detailData.value =
    (await tradeService.getSymbol({ tenantId: row.tenantId, id: row.id })).data?.symbol || row
  detailVisible.value = true
}

const openSymbolDialog = (row?: TradeSymbol) => {
  Object.assign(symbolForm, getDefaultSymbolForm(), row || {})
  symbolVisible.value = true
}

const submitSymbol = async () => {
  if (!categoryTypeFormOptions.value.some((item) => item.value === symbolForm.categoryType)) {
    ElMessage.warning(t('market.pleaseInputCategoryType'))
    return
  }
  if (
    symbolForm.listingTime > 0 &&
    symbolForm.tradingStartTime > 0 &&
    symbolForm.tradingStartTime <= symbolForm.listingTime
  ) {
    ElMessage.warning('开始交易时间必须晚于上线时间')
    return
  }
  if (
    symbolForm.tradingStartTime > 0 &&
    symbolForm.tradingEndTime > 0 &&
    symbolForm.tradingEndTime <= symbolForm.tradingStartTime
  ) {
    ElMessage.warning('停止交易时间必须晚于开始交易时间')
    return
  }
  if (
    symbolForm.productType === 2 &&
    symbolForm.contractType === deliveryContractTypeValue.value &&
    (!symbolForm.listingTime || !symbolForm.tradingStartTime || !symbolForm.tradingEndTime)
  ) {
    ElMessage.warning('交割合约必须配置上线、开始交易和停止交易时间')
    return
  }
  submitLoading.value = true
  try {
    if (symbolForm.id) {
      await tradeService.updateSymbol(symbolForm)
    } else {
      await tradeService.createSymbol(symbolForm)
    }
    ElMessage.success(t('trade.saveSuccess'))
    symbolVisible.value = false
    loadList()
  } finally {
    submitLoading.value = false
  }
}

const openSpotDialog = async (row: TradeSymbol) => {
  Object.assign(spotForm, getDefaultSpotForm(), {
    tenantId: row.tenantId || 0,
    symbolId: row.id || 0,
  })
  const detail = await tradeService.getSymbol({ tenantId: row.tenantId, id: row.id })
  const spot = detail.data?.spot
  if (spot?.symbolId) {
    Object.assign(spotForm, {
      tenantId: spot.tenantId,
      symbolId: spot.symbolId,
      makerFeeRate: spot.makerFeeRate,
      takerFeeRate: spot.takerFeeRate,
      buyEnabled: spot.buyEnabled,
      sellEnabled: spot.sellEnabled,
    })
  }
  spotVisible.value = true
}

const submitSpotConfig = async () => {
  submitLoading.value = true
  try {
    await tradeService.setSpotConfig(spotForm)
    ElMessage.success(t('trade.saveSuccessSpotConfig'))
    spotVisible.value = false
  } finally {
    submitLoading.value = false
  }
}

const openContractDialog = async (row: TradeSymbol) => {
  contractSymbol.value = row
  Object.assign(contractForm, getDefaultContractForm(row), {
    tenantId: row.tenantId || 0,
    symbolId: row.id || 0,
  })
  const detail = await tradeService.getSymbol({ tenantId: row.tenantId, id: row.id })
  const contract = detail.data?.contract
  if (contract?.symbolId) {
    Object.assign(contractForm, {
      tenantId: contract.tenantId,
      symbolId: contract.symbolId,
      contractSize: contract.contractSize,
      multiplier: contract.multiplier,
      maintenanceMarginRate: contract.maintenanceMarginRate,
      initialMarginRate: contract.initialMarginRate,
      makerFeeRate: contract.makerFeeRate,
      takerFeeRate: contract.takerFeeRate,
      fundingIntervalMinutes: contract.fundingIntervalMinutes,
      deliveryTime: contract.deliveryTime,
      supportCross: contract.supportCross,
      supportIsolated: contract.supportIsolated,
      fundingRateCap: contract.fundingRateCap,
      fundingRateFloor: contract.fundingRateFloor,
      fundingRateSource: contract.fundingRateSource,
      indexSymbol: contract.indexSymbol,
      markPriceSource: contract.markPriceSource,
      settlementPriceSource: contract.settlementPriceSource,
      openCutoffTime: contract.openCutoffTime,
      matchingStopTime: contract.matchingStopTime,
      settlementWindowSeconds: contract.settlementWindowSeconds,
      settlementPriceAlgorithm: contract.settlementPriceAlgorithm,
      deliveryFeeRate: contract.deliveryFeeRate,
      liquidationFeeRate: contract.liquidationFeeRate,
      openLongEnabled: contract.openLongEnabled,
      openShortEnabled: contract.openShortEnabled,
      closeLongEnabled: contract.closeLongEnabled,
      closeShortEnabled: contract.closeShortEnabled,
    })
  }
  if (isDeliveryContract.value) applyDefaultDeliveryTimeline()
  contractVisible.value = true
}

const submitContractConfig = async () => {
  if (contractForm.supportCross !== 1 && contractForm.supportIsolated !== 1) {
    ElMessage.warning(t('trade.marginModeSupportRequired'))
    return
  }
  if (isPerpetualContract.value) {
    Object.assign(contractForm, {
      deliveryTime: 0,
      openCutoffTime: 0,
      matchingStopTime: 0,
      settlementWindowSeconds: 0,
      settlementPriceAlgorithm: '',
      deliveryFeeRate: '0',
    })
  } else if (isDeliveryContract.value) {
    if (
      !contractForm.openCutoffTime ||
      !contractForm.matchingStopTime ||
      !contractForm.deliveryTime
    ) {
      ElMessage.warning('交割合约必须配置停止开仓、停止撮合和交割时间')
      return
    }
    if (
      contractForm.openCutoffTime >= contractForm.matchingStopTime ||
      contractForm.matchingStopTime >= contractForm.deliveryTime
    ) {
      ElMessage.warning('时间顺序必须是：停止开仓 < 停止撮合 < 交割')
      return
    }
    if (
      contractSymbol.value?.tradingStartTime &&
      contractForm.openCutoffTime <= contractSymbol.value.tradingStartTime
    ) {
      ElMessage.warning('停止开仓时间不能早于交易对开始交易时间')
      return
    }
    if (
      contractSymbol.value?.tradingEndTime &&
      contractForm.matchingStopTime > contractSymbol.value.tradingEndTime
    ) {
      ElMessage.warning('停止撮合时间不能晚于交易对停止交易时间')
      return
    }
    Object.assign(contractForm, {
      fundingIntervalMinutes: 0,
      fundingRateCap: '0',
      fundingRateFloor: '0',
      fundingRateSource: '',
    })
  }
  submitLoading.value = true
  try {
    await tradeService.setContractConfig(contractForm)
    ElMessage.success(t('trade.saveSuccessContractConfig'))
    contractVisible.value = false
  } finally {
    submitLoading.value = false
  }
}

const loadSecondsConfigs = async () => {
  const row = secondsSymbol.value
  if (!row) return
  const detail = await tradeService.getSymbol({ tenantId: row.tenantId, id: row.id })
  secondsRows.value = (detail.data?.secondsConfigs || [])
    .slice()
    .sort((left, right) => left.durationSeconds - right.durationSeconds)
}

const newSecondsConfig = () => {
  const row = secondsSymbol.value
  Object.assign(secondsForm, getDefaultSecondsForm(), {
    tenantId: row?.tenantId || 0,
    symbolId: row?.id || 0,
  })
  editingSecondsConfig.value = false
}

const editSecondsConfig = (config: TradeSymbolSeconds) => {
  Object.assign(secondsForm, getDefaultSecondsForm(), config)
  editingSecondsConfig.value = true
}

const openSecondsDialog = async (row: TradeSymbol) => {
  secondsSymbol.value = row
  secondsRows.value = []
  newSecondsConfig()
  secondsVisible.value = true
  await loadSecondsConfigs()
  if (secondsRows.value.length) editSecondsConfig(secondsRows.value[0])
}

const submitSecondsConfig = async () => {
  submitLoading.value = true
  try {
    await tradeService.setSecondsConfig(secondsForm)
    ElMessage.success(t('trade.saveSuccess'))
    const savedDuration = secondsForm.durationSeconds
    await loadSecondsConfigs()
    const saved = secondsRows.value.find((item) => item.durationSeconds === savedDuration)
    if (saved) editSecondsConfig(saved)
  } finally {
    submitLoading.value = false
  }
}

const loadLeverageConfigs = async () => {
  const row = leverageSymbol.value
  if (!row) return
  const res = await tradeService.listSymbolLeverageConfigs({
    tenantId: row.tenantId,
    symbolId: row.id,
    limit: 20,
  })
  leverageRows.value = res.data || []
}

const availableLeverageMarginModeOptions = computed(() =>
  marginModeFormOptions.value.filter((item) => {
    const value = Number(item.value)
    if (value === crossMarginModeValue.value) {
      return leverageContractConfig.value?.supportCross === 1
    }
    if (value === isolatedMarginModeValue.value) {
      return leverageContractConfig.value?.supportIsolated === 1
    }
    return false
  }),
)
const availableLeverageMarginModes = computed(() =>
  availableLeverageMarginModeOptions.value.map((item) => Number(item.value)),
)
const configuredLeverageMarginModes = computed(
  () => new Set(leverageRows.value.map((item) => Number(item.marginMode))),
)
const leverageGroups = computed<LeverageForm[]>(() => {
  const supported = new Set(availableLeverageMarginModes.value)
  return leverageRows.value
    .filter((item) => supported.has(Number(item.marginMode)))
    .map((item) => ({
      tenantId: item.tenantId,
      symbolId: item.symbolId,
      marginMode: item.marginMode,
      leverageValues: normalizeLeverageValues(item.leverageValues || []),
      defaultLeverage: item.defaultLeverage || item.leverageValues?.[0] || 1,
      enabled: item.enabled,
      sort: item.sort,
      remark: item.remark || '',
    }))
})
const unusedLeverageMarginModes = computed(() =>
  availableLeverageMarginModes.value.filter(
    (value) => !configuredLeverageMarginModes.value.has(value),
  ),
)
const canAddLeverageConfig = computed(() => unusedLeverageMarginModes.value.length > 0)

const nextLeverageMarginMode = () =>
  unusedLeverageMarginModes.value[0] ||
  availableLeverageMarginModes.value[0] ||
  isolatedMarginModeValue.value

const newLeverageConfig = (marginMode = nextLeverageMarginMode()) => {
  applyLeverageForm({
    ...getDefaultLeverageForm(leverageSymbol.value, marginMode, leverageValueOptions.value),
    enabled: enabledStatusValue.value,
  })
}

const normalizeLeverageValues = (values: number[]) =>
  Array.from(new Set(values.map(Number)))
    .filter((value) => Number.isFinite(value) && value > 0)
    .sort((left, right) => left - right)

const handleLeverageValuesChange = () => {
  const values = normalizeLeverageValues(leverageForm.leverageValues)
  leverageForm.leverageValues = values
  if (!values.length) return

  if (!values.includes(Number(leverageForm.defaultLeverage || 0))) {
    leverageForm.defaultLeverage = values[0]
  }
}

const editLeverageConfig = (row: LeverageForm | TradeSymbolLeverageConfig) => {
  applyLeverageForm(row)
}

const openLeverageDialog = async (row: TradeSymbol) => {
  leverageSymbol.value = row
  leverageContractConfig.value = null
  leverageRows.value = []
  leverageVisible.value = true
  const [detail] = await Promise.all([
    tradeService.getSymbol({ tenantId: row.tenantId, id: row.id }),
    loadLeverageConfigs(),
  ])
  leverageContractConfig.value = detail.data?.contract || null
  if (leverageGroups.value.length) {
    editLeverageConfig(leverageGroups.value[0])
  } else {
    newLeverageConfig()
  }
}

const submitLeverageConfig = async () => {
  if (!availableLeverageMarginModes.value.includes(Number(leverageForm.marginMode))) {
    ElMessage.warning('该合约不支持所选保证金模式，不能配置对应杠杆')
    return
  }
  const values = normalizeLeverageValues(leverageForm.leverageValues)
  if (!values.length) {
    ElMessage.warning(t('trade.leverageValuesRequired'))
    return
  }
  const finalValues = values
  const defaultLeverage = finalValues.includes(leverageForm.defaultLeverage)
    ? leverageForm.defaultLeverage
    : finalValues[0]

  Object.assign(leverageForm, {
    leverageValues: finalValues,
    defaultLeverage,
  })

  submitLoading.value = true
  try {
    await tradeService.setSymbolLeverageConfig({
      tenantId: leverageForm.tenantId,
      symbolId: leverageForm.symbolId,
      marginMode: leverageForm.marginMode,
      leverageValues: finalValues,
      defaultLeverage,
      enabled: leverageForm.enabled,
      sort: leverageForm.sort,
      remark: leverageForm.remark,
    })
    ElMessage.success(t('trade.saveSuccessSymbolLeverageConfig'))
    await loadLeverageConfigs()
  } finally {
    submitLoading.value = false
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

onMounted(() => {
  loadOptions()
  loadList()
})
</script>

<style scoped>
.query-form :deep(.el-form-item) {
  margin-bottom: 12px;
}

.query-field {
  width: 220px;
}

.query-keyword {
  width: 220px;
}

.symbol-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
  line-height: 1.35;
}

.symbol-code {
  font-weight: 600;
}

.muted {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.asset-pair {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}

.dialog-form :deep(.el-form-item) {
  margin-bottom: 14px;
}

:deep(.seconds-config-dialog .el-dialog__body) {
  max-height: calc(100vh - 220px);
  overflow-y: auto;
}

:deep(.contract-config-dialog .el-dialog__body) {
  max-height: calc(100vh - 220px);
  overflow-y: auto;
}

:deep(.contract-config-dialog .el-alert) {
  margin-bottom: 18px;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  column-gap: 16px;
}

.form-grid.two {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.wide {
  grid-column: 1 / -1;
}

.full-width {
  width: 100%;
}

.dialog-subheader {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin: 2px 0 12px;
}

.leverage-table {
  margin-bottom: 4px;
}

@media (max-width: 768px) {
  .query-field,
  .query-keyword {
    width: 100%;
  }

  .form-grid,
  .form-grid.two {
    grid-template-columns: 1fr;
  }
}
</style>
