<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="loadAll" @reset="resetQuery">
      <el-form-item :label="t('option.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('option.userId')">
        <UserSelect v-model="query.userId" :tenant-id="query.tenantId || undefined" />
      </el-form-item>
      <el-form-item :label="t('option.contractId')">
        <ContractSelect v-model="query.contractId" :tenant-id="query.tenantId || undefined" />
      </el-form-item>
      <el-form-item :label="t('option.settleCoin')">
        <el-input v-model="query.settleCoin" clearable style="width: 140px" />
      </el-form-item>
    </CrudQueryCard>

    <el-card v-loading="loading" shadow="never" class="table-card">
      <template #header>
        {{ t('option.riskAccounts') }}
      </template>
      <el-table :data="riskAccounts" stripe>
        <el-table-column prop="userId" :label="t('option.userId')" width="100" />
        <el-table-column prop="accountId" :label="t('option.accountId')" width="110" />
        <el-table-column prop="settleCoin" :label="t('option.settleCoin')" width="100" />
        <el-table-column prop="equity" :label="t('option.equity')" min-width="130" />
        <el-table-column
          prop="netOptionValue"
          :label="t('option.netOptionValue')"
          min-width="140"
        />
        <el-table-column
          prop="maintenanceMargin"
          :label="t('option.maintenanceMargin')"
          min-width="150"
        />
        <el-table-column
          prop="portfolioScenarioLoss"
          :label="t('option.portfolioScenarioLoss')"
          min-width="160"
        />
        <el-table-column
          prop="portfolioShortFloor"
          :label="t('option.portfolioShortFloor')"
          min-width="160"
        />
        <el-table-column
          prop="portfolioConcentrationAddon"
          :label="t('option.portfolioConcentrationAddon')"
          min-width="170"
        />
        <el-table-column
          prop="portfolioLiquidityAddon"
          :label="t('option.portfolioLiquidityAddon')"
          min-width="160"
        />
        <el-table-column
          prop="portfolioRiskConfigVersion"
          :label="t('option.portfolioConfigVersion')"
          width="130"
        />
        <el-table-column prop="riskRate" :label="t('option.riskRate')" min-width="120" />
        <el-table-column prop="status" :label="t('option.status')" width="90" />
      </el-table>
    </el-card>

    <el-card v-loading="loading" shadow="never" class="table-card">
      <template #header>
        <div class="card-header">
          <span>{{ t('option.portfolioRiskGovernance') }}</span>
          <el-button
            v-if="query.tenantId"
            v-perm="'option:portfolio-risk:create'"
            type="primary"
            @click="openPortfolioConfigDialog()"
          >
            {{ t('option.createPortfolioConfig') }}
          </el-button>
        </div>
      </template>
      <el-alert
        :title="t('option.portfolioGovernanceWarning')"
        type="warning"
        :closable="false"
        class="dialog-alert"
      />
      <el-alert
        v-if="!query.tenantId"
        :title="t('option.selectTenantForControlEvents')"
        type="info"
        :closable="false"
      />
      <el-table v-else :data="portfolioConfigs" stripe>
        <el-table-column prop="settleCoin" :label="t('option.settleCoin')" width="100" />
        <el-table-column prop="version" :label="t('option.portfolioConfigVersion')" width="100" />
        <el-table-column :label="t('option.status')" width="110">
          <template #default="{ row }">
            <el-tag :type="portfolioStatusType(row.status)">
              {{ portfolioStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="initialShockRate"
          :label="t('option.initialShockRate')"
          width="140"
        />
        <el-table-column
          prop="maintenanceShockRate"
          :label="t('option.maintenanceShockRate')"
          width="150"
        />
        <el-table-column
          prop="scenarioShocks"
          :label="t('option.scenarioShocks')"
          min-width="220"
          show-overflow-tooltip
        />
        <el-table-column
          prop="concentrationThreshold"
          :label="t('option.concentrationThreshold')"
          min-width="170"
        />
        <el-table-column :label="t('option.portfolioAddonRates')" min-width="150">
          <template #default="{ row }">
            {{ row.concentrationAddonRate }} / {{ row.liquidityAddonRate }}
          </template>
        </el-table-column>
        <el-table-column :label="t('option.effectivePeriod')" min-width="220">
          <template #default="{ row }">
            {{ formatControlTime(row.effectiveFrom) }} –
            {{ row.effectiveUntil ? formatControlTime(row.effectiveUntil) : '∞' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="evidenceRef"
          :label="t('option.evidenceRef')"
          min-width="200"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.actions')" width="210" fixed="right">
          <template #default="{ row }">
            <template v-if="row.status === 1">
              <el-button
                v-perm="'option:portfolio-risk:review'"
                link
                type="success"
                @click="reviewPortfolioConfig(row, true)"
              >
                {{ t('option.approve') }}
              </el-button>
              <el-button
                v-perm="'option:portfolio-risk:review'"
                link
                type="danger"
                @click="reviewPortfolioConfig(row, false)"
              >
                {{ t('option.reject') }}
              </el-button>
            </template>
            <el-button
              v-if="[2, 4].includes(row.status)"
              v-perm="'option:portfolio-risk:create'"
              link
              type="warning"
              @click="openPortfolioConfigDialog(row)"
            >
              {{ t('option.createRollbackVersion') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card v-loading="loading" shadow="never" class="table-card">
      <template #header>
        <div class="card-header">
          <span>{{ t('option.mmpWorkbench') }}</span>
          <el-button
            v-if="query.tenantId && query.userId && query.contractId"
            v-perm="'option:mmp:config'"
            type="primary"
            @click="openMMPDialog()"
          >
            {{ t('option.configureMMP') }}
          </el-button>
        </div>
      </template>
      <el-alert
        v-if="!query.tenantId"
        :title="t('option.selectTenantForControlEvents')"
        type="info"
        :closable="false"
      />
      <el-table v-else :data="mmpConfigs" stripe>
        <el-table-column prop="userId" :label="t('option.userId')" width="100" />
        <el-table-column prop="contractId" :label="t('option.contractId')" width="110" />
        <el-table-column prop="groupCode" :label="t('option.mmpGroup')" min-width="130" />
        <el-table-column :label="t('option.status')" width="110">
          <template #default="{ row }">
            <el-tag :type="mmpStatusType(row.status)">
              {{ mmpStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="qtyThreshold" :label="t('option.mmpQtyThreshold')" min-width="140" />
        <el-table-column
          prop="tradeCountThreshold"
          :label="t('option.mmpTradeThreshold')"
          min-width="140"
        />
        <el-table-column
          prop="lossThreshold"
          :label="t('option.mmpLossThreshold')"
          min-width="150"
        />
        <el-table-column :label="t('option.mmpWindowState')" min-width="210">
          <template #default="{ row }">
            {{ row.accumulatedQty }} / {{ row.tradeCount }} / {{ row.accumulatedLoss }}
          </template>
        </el-table-column>
        <el-table-column
          prop="triggerReason"
          :label="t('option.triggerReason')"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column
          prop="lastErrorMsg"
          :label="t('option.lastError')"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.actions')" width="150" fixed="right">
          <template #default="{ row }">
            <el-button
              v-perm="'option:mmp:config'"
              link
              type="primary"
              @click="openMMPDialog(row)"
            >
              {{ t('common.edit') }}
            </el-button>
            <el-button
              v-if="row.status === 2"
              v-perm="'option:mmp:reset'"
              link
              type="warning"
              @click="resetMMP(row)"
            >
              {{ t('option.resetMMP') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card v-loading="loading" shadow="never" class="table-card">
      <template #header>
        {{ t('option.tradingControls') }}
      </template>
      <el-alert
        v-if="!query.tenantId"
        :title="t('option.selectTenantForControlEvents')"
        type="info"
        :closable="false"
      />
      <template v-else>
        <el-descriptions
          v-if="tradingControl"
          :column="4"
          border
          class="control-summary"
        >
          <el-descriptions-item :label="t('option.userId')">
            {{ tradingControl.userId }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.killSwitch')">
            <el-tag :type="tradingControl.killSwitch === 1 ? 'danger' : 'success'">
              {{ tradingControl.killSwitch === 1 ? t('common.enabled') : t('common.disabled') }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.controlReason')">
            {{ tradingControl.reason || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('common.actions')">
            <el-button
              v-if="tradingControl.killSwitch === 1"
              v-perm="'option:trading-control:release'"
              link
              type="warning"
              @click="releaseKillSwitch"
            >
              {{ t('option.releaseKillSwitch') }}
            </el-button>
          </el-descriptions-item>
        </el-descriptions>
        <el-table :data="controlEvents" stripe>
          <el-table-column prop="id" label="ID" width="90" />
          <el-table-column prop="eventType" :label="t('option.controlEventType')" min-width="180" />
          <el-table-column prop="reason" :label="t('option.controlReason')" min-width="180" />
          <el-table-column prop="userId" :label="t('option.userId')" width="100" />
          <el-table-column prop="contractId" :label="t('option.contractId')" width="110" />
          <el-table-column prop="orderId" :label="t('option.orderId')" width="110" />
          <el-table-column
            prop="detail"
            :label="t('option.controlDetail')"
            min-width="280"
            show-overflow-tooltip
          />
          <el-table-column :label="t('common.createTimes')" width="180">
            <template #default="{ row }">
              {{ formatControlTime(row.createTimes) }}
            </template>
          </el-table-column>
        </el-table>
      </template>
    </el-card>

    <el-card v-loading="loading" shadow="never" class="table-card">
      <template #header>
        <div class="card-header">
          <span>{{ t('option.tradeCorrections') }}</span>
          <el-button
            v-if="query.tenantId"
            v-perm="'option:trade-correction:create'"
            type="danger"
            @click="openCorrectionDialog"
          >
            {{ t('option.createTradeCorrection') }}
          </el-button>
        </div>
      </template>
      <el-alert
        v-if="!query.tenantId"
        :title="t('option.selectTenantForControlEvents')"
        type="info"
        :closable="false"
      />
      <el-table v-else :data="tradeCorrections" stripe>
        <el-table-column type="expand">
          <template #default="{ row }">
            <el-table :data="row.legs" size="small" border>
              <el-table-column prop="legNo" :label="t('option.legNo')" width="80" />
              <el-table-column prop="userId" :label="t('option.userId')" width="110" />
              <el-table-column prop="accountId" :label="t('option.accountId')" width="120" />
              <el-table-column prop="coin" :label="t('option.coin')" width="90" />
              <el-table-column :label="t('option.correctionDirection')" width="110">
                <template #default="{ row: leg }">
                  {{ leg.direction === 1 ? t('option.debit') : t('option.credit') }}
                </template>
              </el-table-column>
              <el-table-column prop="amount" :label="t('option.amount')" min-width="130" />
              <el-table-column
                prop="instructionNo"
                :label="t('option.instructionNo')"
                min-width="240"
              />
            </el-table>
          </template>
        </el-table-column>
        <el-table-column prop="caseNo" :label="t('option.caseNo')" min-width="220" />
        <el-table-column prop="tradeId" :label="t('option.tradeId')" width="110" />
        <el-table-column prop="contractId" :label="t('option.contractId')" width="110" />
        <el-table-column :label="t('option.status')" width="130">
          <template #default="{ row }">
            {{ correctionStatusLabel(row.status) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="reason"
          :label="t('option.correctionReason')"
          min-width="220"
          show-overflow-tooltip
        />
        <el-table-column
          prop="evidenceRef"
          :label="t('option.evidenceRef')"
          min-width="200"
          show-overflow-tooltip
        />
        <el-table-column
          prop="lastErrorMsg"
          :label="t('option.lastError')"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.actions')" width="170" fixed="right">
          <template #default="{ row }">
            <template v-if="row.status === 1">
              <el-button
                v-perm="'option:trade-correction:review'"
                link
                type="success"
                @click="reviewCorrection(row, true)"
              >
                {{ t('option.approve') }}
              </el-button>
              <el-button
                v-perm="'option:trade-correction:review'"
                link
                type="danger"
                @click="reviewCorrection(row, false)"
              >
                {{ t('option.reject') }}
              </el-button>
            </template>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card v-loading="loading" shadow="never" class="table-card">
      <template #header>
        {{ t('option.liquidations') }}
      </template>
      <el-table :data="liquidations" stripe>
        <el-table-column prop="liquidationNo" :label="t('option.liquidationNo')" min-width="220" />
        <el-table-column prop="userId" :label="t('option.userId')" width="100" />
        <el-table-column prop="contractId" :label="t('option.contractId')" width="100" />
        <el-table-column prop="quantity" :label="t('option.quantity')" min-width="110" />
        <el-table-column prop="collateralAmount" :label="t('option.collateral')" min-width="130" />
        <el-table-column
          prop="insuranceFundAmount"
          :label="t('option.insuranceFund')"
          min-width="140"
        />
        <el-table-column
          prop="backstopAmount"
          :label="t('option.platformBackstop')"
          min-width="130"
        />
        <el-table-column
          prop="deficitResolution"
          :label="t('option.deficitResolution')"
          min-width="130"
        />
        <el-table-column prop="remainingDeficit" :label="t('option.deficit')" min-width="120" />
        <el-table-column prop="status" :label="t('option.status')" width="90" />
        <el-table-column :label="t('common.actions')" width="110" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="[4, 6].includes(row.status)"
              v-perm="'option:liquidation:retry'"
              link
              type="warning"
              @click="retryLiquidation(row)"
            >
              {{ t('option.retry') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog
      v-model="portfolioDialogVisible"
      :title="
        portfolioForm.sourceConfigId
          ? t('option.createRollbackVersion')
          : t('option.createPortfolioConfig')
      "
      width="720px"
      destroy-on-close
    >
      <el-alert
        :title="t('option.portfolioApprovalWarning')"
        type="warning"
        :closable="false"
        class="dialog-alert"
      />
      <el-form :model="portfolioForm" label-width="210px">
        <el-form-item :label="t('option.settleCoin')" required>
          <el-input v-model="portfolioForm.settleCoin" maxlength="16" />
        </el-form-item>
        <el-form-item :label="t('option.initialShockRate')" required>
          <el-input
            v-model="portfolioForm.initialShockRate"
            :disabled="Boolean(portfolioForm.sourceConfigId)"
          />
        </el-form-item>
        <el-form-item :label="t('option.maintenanceShockRate')" required>
          <el-input
            v-model="portfolioForm.maintenanceShockRate"
            :disabled="Boolean(portfolioForm.sourceConfigId)"
          />
        </el-form-item>
        <el-form-item :label="t('option.scenarioShocks')" required>
          <el-input
            v-model="portfolioForm.scenarioShocks"
            :disabled="Boolean(portfolioForm.sourceConfigId)"
          />
        </el-form-item>
        <el-form-item :label="t('option.concentrationThreshold')" required>
          <el-input
            v-model="portfolioForm.concentrationThreshold"
            :disabled="Boolean(portfolioForm.sourceConfigId)"
          />
        </el-form-item>
        <el-form-item :label="t('option.concentrationAddonRate')" required>
          <el-input
            v-model="portfolioForm.concentrationAddonRate"
            :disabled="Boolean(portfolioForm.sourceConfigId)"
          />
        </el-form-item>
        <el-form-item :label="t('option.liquidityAddonRate')" required>
          <el-input
            v-model="portfolioForm.liquidityAddonRate"
            :disabled="Boolean(portfolioForm.sourceConfigId)"
          />
        </el-form-item>
        <el-form-item :label="t('option.effectiveFromUnix')" required>
          <el-input-number v-model="portfolioForm.effectiveFrom" :min="1" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('option.changeReason')" required>
          <el-input
            v-model="portfolioForm.changeReason"
            type="textarea"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
        <el-form-item :label="t('option.evidenceRef')" required>
          <el-input v-model="portfolioForm.evidenceRef" maxlength="500" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="portfolioDialogVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button type="primary" :loading="submittingPortfolio" @click="submitPortfolioConfig">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="mmpDialogVisible"
      :title="t('option.configureMMP')"
      width="680px"
      destroy-on-close
    >
      <el-alert
        :title="t('option.mmpConfigWarning')"
        type="warning"
        :closable="false"
        class="dialog-alert"
      />
      <el-form :model="mmpForm" label-width="180px">
        <el-form-item :label="t('option.mmpGroup')" required>
          <el-input v-model="mmpForm.groupCode" maxlength="32" :disabled="mmpEditing" />
        </el-form-item>
        <el-form-item :label="t('common.status')" required>
          <el-switch
            v-model="mmpForm.enabled"
            :active-value="1"
            :inactive-value="2"
            :active-text="t('common.enabled')"
            :inactive-text="t('common.disabled')"
          />
        </el-form-item>
        <el-form-item :label="t('option.mmpQtyThreshold')" required>
          <el-input v-model="mmpForm.qtyThreshold" />
        </el-form-item>
        <el-form-item :label="t('option.mmpTradeThreshold')" required>
          <el-input-number v-model="mmpForm.tradeCountThreshold" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('option.mmpLossThreshold')" required>
          <el-input v-model="mmpForm.lossThreshold" />
        </el-form-item>
        <el-form-item :label="t('option.mmpWindowSeconds')" required>
          <el-input-number
            v-model="mmpForm.windowSeconds"
            :min="1"
            :max="3600"
            :precision="0"
          />
        </el-form-item>
        <el-form-item :label="t('option.mmpCooldownSeconds')" required>
          <el-input-number
            v-model="mmpForm.cooldownSeconds"
            :min="0"
            :max="86400"
            :precision="0"
          />
        </el-form-item>
        <el-form-item :label="t('option.controlReason')" required>
          <el-input
            v-model="mmpForm.reason"
            type="textarea"
            :rows="2"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="mmpDialogVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button type="primary" :loading="submittingMMP" @click="submitMMP">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="correctionDialogVisible"
      :title="t('option.createTradeCorrection')"
      width="900px"
      destroy-on-close
    >
      <el-alert
        :title="t('option.tradeCorrectionWarning')"
        type="error"
        :closable="false"
        class="dialog-alert"
      />
      <el-form :model="correctionForm" label-width="130px">
        <el-form-item :label="t('option.tradeId')" required>
          <el-input-number v-model="correctionForm.tradeId" :min="1" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('option.correctionReason')" required>
          <el-input
            v-model="correctionForm.reason"
            type="textarea"
            :rows="2"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
        <el-form-item :label="t('option.evidenceRef')" required>
          <el-input v-model="correctionForm.evidenceRef" maxlength="500" />
        </el-form-item>
        <el-form-item :label="t('option.correctionLegs')" required>
          <div class="legs-editor">
            <div v-for="(leg, index) in correctionForm.legs" :key="index" class="leg-row">
              <el-input-number
                v-model="leg.userId"
                :min="1"
                :precision="0"
                :placeholder="t('option.userId')"
              />
              <el-input-number
                v-model="leg.accountId"
                :min="0"
                :precision="0"
                :placeholder="t('option.accountId')"
              />
              <el-input v-model="leg.coin" :placeholder="t('option.coin')" />
              <el-select v-model="leg.direction">
                <el-option :label="t('option.debit')" :value="1" />
                <el-option :label="t('option.credit')" :value="2" />
              </el-select>
              <el-input v-model="leg.amount" :placeholder="t('option.amount')" />
              <el-button
                :disabled="correctionForm.legs.length <= 2"
                type="danger"
                link
                @click="removeCorrectionLeg(index)"
              >
                {{ t('common.delete') }}
              </el-button>
            </div>
            <el-button link type="primary" @click="addCorrectionLeg">
              {{ t('option.addCorrectionLeg') }}
            </el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="correctionDialogVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button type="danger" :loading="submittingCorrection" @click="submitCorrection">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import {
  optionService,
  type OptionLiquidation,
  type OptionRiskAccount,
  type OptionTradingControlEvent,
  type OptionTradeCorrection,
  type TradeCorrectionLegInput,
  type OptionUserTradingControl,
  type OptionMMPConfig,
  type OptionPortfolioRiskConfig,
} from '@/services'
import TenantSelect from '@/components/TenantSelect.vue'
import UserSelect from '@/components/UserSelect.vue'
import ContractSelect from '@/components/ContractSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()
const loading = ref(false)
const riskAccounts = ref<OptionRiskAccount[]>([])
const liquidations = ref<OptionLiquidation[]>([])
const tradingControl = ref<OptionUserTradingControl | null>(null)
const controlEvents = ref<OptionTradingControlEvent[]>([])
const tradeCorrections = ref<OptionTradeCorrection[]>([])
const mmpConfigs = ref<OptionMMPConfig[]>([])
const portfolioConfigs = ref<OptionPortfolioRiskConfig[]>([])
const portfolioDialogVisible = ref(false)
const submittingPortfolio = ref(false)
const portfolioForm = reactive({
  settleCoin: 'USDT',
  modelMethod: 1,
  initialShockRate: '0.3',
  maintenanceShockRate: '0.2',
  scenarioShocks: '-1,-0.5,-0.3,-0.2,0.2,0.3,0.5,1,4',
  concentrationThreshold: '100000',
  concentrationAddonRate: '0.1',
  liquidityAddonRate: '0.02',
  effectiveFrom: Math.floor(Date.now() / 1000),
  changeReason: '',
  evidenceRef: '',
  sourceConfigId: 0,
})
const mmpDialogVisible = ref(false)
const submittingMMP = ref(false)
const mmpEditing = ref(false)
const mmpForm = reactive({
  groupCode: '',
  enabled: 1,
  qtyThreshold: '0',
  tradeCountThreshold: 0,
  lossThreshold: '0',
  windowSeconds: 5,
  cooldownSeconds: 30,
  reason: '',
})
const correctionDialogVisible = ref(false)
const submittingCorrection = ref(false)
const correctionForm = reactive({
  tradeId: 0,
  reason: '',
  evidenceRef: '',
  legs: [] as TradeCorrectionLegInput[],
})
const query = reactive({
  tenantId: undefined as number | undefined,
  userId: undefined as number | undefined,
  contractId: undefined as number | undefined,
  settleCoin: '',
})

async function loadAll() {
  loading.value = true
  try {
    const [riskResp, liquidationResp] = await Promise.all([
      optionService.listRiskAccounts({ ...query, limit: 100 }),
      optionService.listLiquidations({ ...query, limit: 100 }),
    ])
    riskAccounts.value = riskResp.data || []
    liquidations.value = liquidationResp.data || []
    tradingControl.value = null
    controlEvents.value = []
    tradeCorrections.value = []
    mmpConfigs.value = []
    portfolioConfigs.value = []
    if (query.tenantId) {
      const [eventsResp, correctionsResp, mmpResp, portfolioResp] = await Promise.all([
        optionService.listTradingControlEvents({
          tenantId: query.tenantId,
          userId: query.userId,
          contractId: query.contractId,
          limit: 100,
        }),
        optionService.listTradeCorrections({
          tenantId: query.tenantId,
          contractId: query.contractId,
          limit: 100,
        }),
        optionService.listMMPConfigs({
          tenantId: query.tenantId,
          userId: query.userId,
          contractId: query.contractId,
          limit: 100,
        }),
        optionService.listPortfolioRiskConfigs({
          tenantId: query.tenantId,
          settleCoin: query.settleCoin || undefined,
          limit: 100,
        }),
      ])
      controlEvents.value = eventsResp.data || []
      tradeCorrections.value = correctionsResp.data || []
      mmpConfigs.value = mmpResp.data || []
      portfolioConfigs.value = portfolioResp.data || []
      if (query.userId) {
        const controlResp = await optionService.getUserTradingControl({
          tenantId: query.tenantId,
          userId: query.userId,
        })
        tradingControl.value = controlResp.data || null
      }
    }
  } finally {
    loading.value = false
  }
}

function openPortfolioConfigDialog(source?: OptionPortfolioRiskConfig) {
  if (!query.tenantId) {
    ElMessage.error(t('option.selectTenantForControlEvents'))
    return
  }
  portfolioForm.settleCoin = source?.settleCoin || query.settleCoin || 'USDT'
  portfolioForm.modelMethod = 1
  portfolioForm.initialShockRate = source?.initialShockRate || '0.3'
  portfolioForm.maintenanceShockRate = source?.maintenanceShockRate || '0.2'
  portfolioForm.scenarioShocks = source?.scenarioShocks || '-1,-0.5,-0.3,-0.2,0.2,0.3,0.5,1,4'
  portfolioForm.concentrationThreshold = source?.concentrationThreshold || '100000'
  portfolioForm.concentrationAddonRate = source?.concentrationAddonRate || '0.1'
  portfolioForm.liquidityAddonRate = source?.liquidityAddonRate || '0.02'
  portfolioForm.effectiveFrom = Math.floor(Date.now() / 1000) + 60
  portfolioForm.changeReason = ''
  portfolioForm.evidenceRef = ''
  portfolioForm.sourceConfigId = source?.id || 0
  portfolioDialogVisible.value = true
}

async function submitPortfolioConfig() {
  if (
    !query.tenantId ||
    !portfolioForm.settleCoin.trim() ||
    !portfolioForm.changeReason.trim() ||
    !portfolioForm.evidenceRef.trim() ||
    portfolioForm.effectiveFrom <= 0
  ) {
    ElMessage.error(t('option.completePortfolioConfig'))
    return
  }
  submittingPortfolio.value = true
  try {
    await optionService.createPortfolioRiskConfig({
      tenantId: query.tenantId,
      ...portfolioForm,
      settleCoin: portfolioForm.settleCoin.trim().toUpperCase(),
      changeReason: portfolioForm.changeReason.trim(),
      evidenceRef: portfolioForm.evidenceRef.trim(),
      sourceConfigId: portfolioForm.sourceConfigId || undefined,
    })
    portfolioDialogVisible.value = false
    ElMessage.success(t('common.success'))
    await loadAll()
  } finally {
    submittingPortfolio.value = false
  }
}

async function reviewPortfolioConfig(row: OptionPortfolioRiskConfig, approve: boolean) {
  const { value } = await ElMessageBox.prompt(
    approve ? t('option.approvePortfolioReason') : t('option.rejectPortfolioReason'),
    approve ? t('option.approve') : t('option.reject'),
    { inputValidator: (input) => Boolean(input?.trim()) || t('option.reviewReasonRequired') },
  )
  await optionService.reviewPortfolioRiskConfig({
    tenantId: row.tenantId,
    configId: row.id,
    approve,
    reason: value.trim(),
  })
  ElMessage.success(t('common.success'))
  await loadAll()
}

function portfolioStatusLabel(status: number) {
  return (
    {
      1: t('option.pendingReview'),
      2: t('option.approved'),
      3: t('option.rejected'),
      4: t('option.superseded'),
    }[status] || String(status)
  )
}

function portfolioStatusType(status: number) {
  return status === 2 ? 'success' : status === 1 ? 'warning' : status === 3 ? 'danger' : 'info'
}

function openMMPDialog(row?: OptionMMPConfig) {
  if (!query.tenantId || !query.userId || !query.contractId) {
    ElMessage.error(t('option.selectMMPContext'))
    return
  }
  mmpEditing.value = Boolean(row)
  mmpForm.groupCode = row?.groupCode || ''
  mmpForm.enabled = row?.enabled || 1
  mmpForm.qtyThreshold = row?.qtyThreshold || '0'
  mmpForm.tradeCountThreshold = row?.tradeCountThreshold || 0
  mmpForm.lossThreshold = row?.lossThreshold || '0'
  mmpForm.windowSeconds = row?.windowSeconds || 5
  mmpForm.cooldownSeconds = row?.cooldownSeconds ?? 30
  mmpForm.reason = ''
  mmpDialogVisible.value = true
}

async function submitMMP() {
  if (
    !query.tenantId ||
    !query.userId ||
    !query.contractId ||
    !/^[A-Za-z0-9_-]{1,32}$/.test(mmpForm.groupCode) ||
    !mmpForm.reason.trim()
  ) {
    ElMessage.error(t('option.completeMMPForm'))
    return
  }
  if (
    mmpForm.enabled === 1 &&
    Number(mmpForm.qtyThreshold) <= 0 &&
    mmpForm.tradeCountThreshold <= 0 &&
    Number(mmpForm.lossThreshold) <= 0
  ) {
    ElMessage.error(t('option.mmpThresholdRequired'))
    return
  }
  submittingMMP.value = true
  try {
    await optionService.upsertMMPConfig({
      tenantId: query.tenantId,
      userId: query.userId,
      contractId: query.contractId,
      ...mmpForm,
      reason: mmpForm.reason.trim(),
    })
    mmpDialogVisible.value = false
    ElMessage.success(t('common.success'))
    await loadAll()
  } finally {
    submittingMMP.value = false
  }
}

async function resetMMP(row: OptionMMPConfig) {
  const { value } = await ElMessageBox.prompt(t('option.resetMMPReason'), t('option.resetMMP'), {
    inputValidator: (input) => Boolean(input?.trim()) || t('option.reviewReasonRequired'),
  })
  await optionService.resetMMPConfig({
    tenantId: row.tenantId,
    userId: row.userId,
    contractId: row.contractId,
    groupCode: row.groupCode,
    reason: value.trim(),
  })
  ElMessage.success(t('common.success'))
  await loadAll()
}

function mmpStatusLabel(status: number) {
  return (
    {
      1: t('option.mmpActive'),
      2: t('option.mmpTriggered'),
      3: t('option.mmpDisabled'),
    }[status] || String(status)
  )
}

function mmpStatusType(status: number) {
  return status === 1 ? 'success' : status === 2 ? 'danger' : 'info'
}

function resetQuery() {
  query.tenantId = undefined
  query.userId = undefined
  query.contractId = undefined
  query.settleCoin = ''
  void loadAll()
}

async function retryLiquidation(row: OptionLiquidation) {
  await ElMessageBox.confirm(t('option.retryLiquidationConfirm'), t('option.retryLiquidation'), {
    type: 'warning',
  })
  await optionService.retryLiquidation({
    tenantId: row.tenantId,
    liquidationId: row.id,
  })
  ElMessage.success(t('common.success'))
  await loadAll()
}

async function releaseKillSwitch() {
  if (!tradingControl.value) return
  const { value } = await ElMessageBox.prompt(
    t('option.releaseKillSwitchReason'),
    t('option.releaseKillSwitch'),
    {
      inputValidator: (input) => Boolean(input?.trim()) || t('option.releaseReasonRequired'),
      type: 'warning',
    },
  )
  await optionService.releaseUserKillSwitch({
    tenantId: tradingControl.value.tenantId,
    userId: tradingControl.value.userId,
    reason: value.trim(),
  })
  ElMessage.success(t('common.success'))
  await loadAll()
}

function formatControlTime(timestamp: number) {
  return timestamp ? new Date(timestamp * 1000).toLocaleString() : '-'
}

function newCorrectionLeg(direction: number): TradeCorrectionLegInput {
  return { userId: 0, accountId: 0, coin: '', direction, amount: '' }
}

function openCorrectionDialog() {
  correctionForm.tradeId = 0
  correctionForm.reason = ''
  correctionForm.evidenceRef = ''
  correctionForm.legs = [newCorrectionLeg(1), newCorrectionLeg(2)]
  correctionDialogVisible.value = true
}

function addCorrectionLeg() {
  correctionForm.legs.push(newCorrectionLeg(2))
}

function removeCorrectionLeg(index: number) {
  correctionForm.legs.splice(index, 1)
}

async function submitCorrection() {
  if (
    !query.tenantId ||
    correctionForm.tradeId <= 0 ||
    !correctionForm.reason.trim() ||
    !correctionForm.evidenceRef.trim() ||
    correctionForm.legs.length < 2 ||
    correctionForm.legs.some((leg) => leg.userId <= 0 || !leg.coin.trim() || !leg.amount)
  ) {
    ElMessage.error(t('option.completeCorrectionForm'))
    return
  }
  submittingCorrection.value = true
  try {
    await optionService.createTradeCorrection({
      tenantId: query.tenantId,
      tradeId: correctionForm.tradeId,
      action: 1,
      reason: correctionForm.reason.trim(),
      evidenceRef: correctionForm.evidenceRef.trim(),
      legs: correctionForm.legs.map((leg) => ({ ...leg, coin: leg.coin.trim() })),
    })
    correctionDialogVisible.value = false
    ElMessage.success(t('common.success'))
    await loadAll()
  } finally {
    submittingCorrection.value = false
  }
}

async function reviewCorrection(row: OptionTradeCorrection, approve: boolean) {
  const { value } = await ElMessageBox.prompt(
    approve ? t('option.approveCorrectionReason') : t('option.rejectCorrectionReason'),
    approve ? t('option.approve') : t('option.reject'),
    { inputValidator: (input) => Boolean(input?.trim()) || t('option.reviewReasonRequired') },
  )
  await optionService.reviewTradeCorrection({
    tenantId: row.tenantId,
    correctionId: row.id,
    approve,
    reason: value.trim(),
  })
  ElMessage.success(t('common.success'))
  await loadAll()
}

function correctionStatusLabel(status: number) {
  const key = {
    1: 'pendingReview',
    2: 'rejected',
    3: 'executing',
    4: 'completed',
    5: 'manualReview',
  }[status]
  return key ? t(`option.${key}`) : String(status)
}

onMounted(loadAll)
</script>

<style scoped>
.table-card + .table-card {
  margin-top: 16px;
}

.control-summary {
  margin-bottom: 16px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.dialog-alert {
  margin-bottom: 16px;
}

.legs-editor {
  width: 100%;
}

.leg-row {
  display: grid;
  grid-template-columns: 150px 150px 100px 120px 140px 70px;
  gap: 8px;
  margin-bottom: 8px;
}
</style>
