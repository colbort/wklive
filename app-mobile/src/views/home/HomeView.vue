<script setup lang="ts">
import { RouterLink, useRouter } from 'vue-router'

import AppIcon from '@/components/common/AppIcon.vue'
import CommonPage from '@/components/common/CommonPage.vue'
import boxIconSimple from '../../../assets/home/icon1.png'
import boxIconShield from '../../../assets/home/icon2.png'
import boxIconSupport from '../../../assets/home/icon3.png'
import boxIconOrders from '../../../assets/home/icon4.png'
import boxIconApi from '../../../assets/home/icon5.png'
import heroPortalImage from '../../../assets/home/img1.png'
import smartPhonesImage from '../../../assets/home/img2.png'
import strategyOrdersImage from '../../../assets/home/img3.png'
import strategyChartImage from '../../../assets/home/img4.png'
import contractPhonesImage from '../../../assets/home/img5.png'
import rewardImage from '../../../assets/home/img6.png'
import featureIconLiquidity from '../../../assets/home/svg1.svg'
import featureIconCharts from '../../../assets/home/svg2.svg'
import featureIconFees from '../../../assets/home/svg3.svg'
import { useI18n } from '@/i18n'

const { t } = useI18n()
const router = useRouter()

const featureCards = [
  {
    icon: featureIconLiquidity,
    titleKey: 'home.deepLiquidity',
    descKey: 'home.deepLiquidityDesc',
  },
  {
    icon: featureIconCharts,
    titleKey: 'home.proCharts',
    descKey: 'home.proChartsDesc',
  },
  {
    icon: featureIconFees,
    titleKey: 'home.lowFees',
    descKey: 'home.lowFeesDesc',
  },
]

const strategyFeatures = [
  {
    icon: boxIconOrders,
    titleKey: 'home.smartOrders',
    descKey: 'home.smartOrdersDesc',
  },
  {
    icon: boxIconApi,
    titleKey: 'home.apiTrading',
    descKey: 'home.apiTradingDesc',
  },
]

const contractFeatures = [
  {
    titleKey: 'home.noExpiry',
    descKey: 'home.noExpiryDesc',
  },
  {
    titleKey: 'home.globalTrading',
    descKey: 'home.globalTradingDesc',
  },
  {
    titleKey: 'home.twoWayTrading',
    descKey: 'home.twoWayTradingDesc',
  },
  {
    titleKey: 'home.lowSlippage',
    descKey: 'home.lowSlippageDesc',
  },
]

const trustFeatures = [
  {
    icon: boxIconSimple,
    titleKey: 'home.simpleExperience',
    descKey: 'home.simpleExperienceDesc',
  },
  {
    icon: boxIconShield,
    titleKey: 'home.assetSecurity',
    descKey: 'home.assetSecurityDesc',
  },
  {
    icon: boxIconSupport,
    titleKey: 'home.proSupport',
    descKey: 'home.proSupportDesc',
  },
]

function changeLanguage() {
  router.push('/language')
}
</script>

<template>
  <CommonPage :show-back="false" :nav-height="72">
    <template #tabbar>
      <header class="ave-home-tabbar">
        <RouterLink to="/home" class="ave-home-brand">
          <span class="ave-home-brand__mark" />
          <strong>AVE</strong>
        </RouterLink>

        <div class="ave-home-tabbar__actions">
          <button class="ave-home-action" type="button" :aria-label="t('common.search')">
            <AppIcon name="search" class="ave-home-action__icon" />
          </button>
          <button
            class="ave-home-action"
            type="button"
            :aria-label="t('common.language')"
            @click="changeLanguage()"
          >
            <AppIcon name="globe" class="ave-home-action__icon" />
          </button>
        </div>
      </header>
    </template>

    <template #default>
      <section class="ave-home">
        <section class="ave-hero">
          <div class="ave-hero__copy">
            <h1>
              <span>{{ t('home.heroPrefix') }}</span>
              {{ t('home.heroSuffix') }}
            </h1>
            <p>{{ t('home.heroSub') }}</p>
            <RouterLink to="/trades" class="ave-home__cta ave-hero__cta">
              {{ t('home.startTrading') }}
            </RouterLink>
          </div>

          <img class="ave-hero__portal" :src="heroPortalImage" alt="">
        </section>

        <section class="ave-section ave-section--smart">
          <h2>
            {{ t('home.smartTradingPrefix') }}<span>{{ t('home.smartTradingEmphasis') }}</span>{{ t('home.smartTradingSuffix') }}
          </h2>
          <p>{{ t('home.alwaysTrading') }}</p>
          <img class="ave-section__phones" :src="smartPhonesImage" alt="">

          <div class="ave-feature-list">
            <article v-for="item in featureCards" :key="item.titleKey" class="ave-feature-card">
              <img :src="item.icon" alt="">
              <div>
                <h3>{{ t(item.titleKey) }}</h3>
                <p>{{ t(item.descKey) }}</p>
              </div>
            </article>
          </div>
        </section>

        <section class="ave-section ave-section--strategy">
          <h2>
            {{ t('home.strategyPrefix') }}<span>{{ t('home.strategyEmphasis') }}</span>{{ t('home.strategySuffix') }}
          </h2>
          <p>{{ t('home.strategySub') }}</p>

          <div class="ave-strategy-scene">
            <img class="ave-strategy-scene__orders" :src="strategyOrdersImage" alt="">
            <img class="ave-strategy-scene__chart" :src="strategyChartImage" alt="">
          </div>

          <div class="ave-flat-list">
            <article v-for="item in strategyFeatures" :key="item.titleKey" class="ave-flat-item">
              <img :src="item.icon" alt="">
              <div>
                <h3>{{ t(item.titleKey) }}</h3>
                <p>{{ t(item.descKey) }}</p>
              </div>
            </article>
          </div>
        </section>

        <section class="ave-section ave-section--contract">
          <h2>{{ t('home.contractTitle') }}</h2>
          <img class="ave-contract__phones" :src="contractPhonesImage" alt="">

          <div class="ave-toggle-card">
            <article v-for="item in contractFeatures" :key="item.titleKey" class="ave-toggle-item">
              <span class="ave-toggle-item__switch" />
              <div>
                <h3>{{ t(item.titleKey) }}</h3>
                <p>{{ t(item.descKey) }}</p>
              </div>
            </article>
          </div>
        </section>

        <section class="ave-section ave-section--trust">
          <h2>
            {{ t('home.trustPrefix') }}<span>{{ t('home.trustEmphasis') }}</span>
          </h2>
          <p>{{ t('home.trustSub') }}</p>

          <div class="ave-trust-card">
            <article v-for="item in trustFeatures" :key="item.titleKey" class="ave-trust-item">
              <img :src="item.icon" alt="">
              <div>
                <h3>{{ t(item.titleKey) }}</h3>
                <p>{{ t(item.descKey) }}</p>
              </div>
            </article>
          </div>
        </section>

        <section class="ave-reward">
          <h2>{{ t('home.rewardTitle') }}</h2>
          <p>{{ t('home.rewardSub') }}</p>
          <RouterLink to="/trades" class="ave-home__cta">
            {{ t('home.startTrading') }}
          </RouterLink>
          <img class="ave-reward__image" :src="rewardImage" alt="">
        </section>
      </section>
    </template>
  </CommonPage>
</template>

<style scoped>
.ave-home-tabbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  height: 72px;
  padding: 8px 18px;
  background:
    linear-gradient(rgba(124, 146, 170, 0.08) 2px, transparent 2px),
    linear-gradient(90deg, rgba(124, 146, 170, 0.08) 2px, transparent 2px), rgba(10, 13, 22, 0.96);
  background-size:
    104px 104px,
    104px 104px,
    auto;
  box-sizing: border-box;
}

.ave-home-brand {
  display: inline-flex;
  align-items: center;
  min-width: 0;
  gap: 8px;
}

.ave-home-brand__mark {
  flex: 0 0 auto;
  width: 42px;
  height: 42px;
  background: url('../../../assets/favicon.png') center / contain no-repeat;
}

.ave-home-brand strong {
  color: var(--text);
  font-size: 1.35rem;
  font-weight: 800;
  line-height: 1;
}

.ave-home-tabbar__actions {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  flex: none;
}

.ave-home-action {
  display: grid;
  width: 38px;
  height: 38px;
  place-items: center;
  border: 0;
  border-radius: 50%;
  background: rgba(29, 32, 44, 0.9);
  color: var(--text);
  font: inherit;
  font-size: 0.95rem;
}

.ave-home-action__icon {
  width: 22px;
  height: 22px;
}

.ave-home {
  --home-green: #00b90f;
  --home-bg: #0b0e18;
  --home-muted: rgba(255, 255, 255, 0.55);
  width: min(100%, var(--app-width, 414px));
  margin: 0 auto;
  overflow: hidden;
  background:
    radial-gradient(circle at 18% 2%, rgba(78, 92, 78, 0.28), transparent 24rem),
    linear-gradient(rgba(124, 146, 170, 0.08) 2px, transparent 2px),
    linear-gradient(90deg, rgba(124, 146, 170, 0.08) 2px, transparent 2px), var(--home-bg);
  background-size:
    auto,
    206px 206px,
    206px 206px,
    auto;
  color: var(--text);
}

.ave-home span {
  color: var(--home-green);
}

.ave-hero,
.ave-section,
.ave-reward {
  position: relative;
  padding-inline: 18px;
}

.ave-hero {
  min-height: 650px;
  padding-top: 38px;
  text-align: center;
}

.ave-hero__copy {
  position: relative;
  z-index: 2;
}

.ave-hero h1,
.ave-section h2,
.ave-reward h2 {
  margin: 0;
  color: var(--text);
  font-weight: 800;
  letter-spacing: 0;
}

.ave-hero h1 {
  display: grid;
  gap: 1px;
  font-size: clamp(1.7rem, 9vw, 1.9rem);
  line-height: 1.1;
}

.ave-hero p {
  margin: 18px 0 0;
  color: var(--text);
  font-size: clamp(0.9rem, 4.8vw, 1.02rem);
  font-weight: 700;
}

.ave-home__cta {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 140px;
  height: 44px;
  margin-top: 20px;
  border-radius: 999px;
  background: var(--home-green);
  color: var(--text);
  font-size: 1.05rem;
  font-weight: 700;
  line-height: 1;
}

.ave-hero__cta {
  width: 140px;
  height: 44px;
  max-width: 68vw;
  margin-top: 20px;
  font-size: 1.05rem;
}

.ave-hero__portal {
  position: absolute;
  left: 50%;
  width: min(100%, var(--app-width, 414px));
  transform: translateX(-50%);
}

.ave-section {
  padding-top: 28px;
  text-align: center;
}

.ave-section h2,
.ave-reward h2 {
  font-size: clamp(1.25rem, 7vw, 1.42rem);
  line-height: 1.22;
}

.ave-section > p,
.ave-reward > p {
  margin: 10px 0 0;
  color: var(--home-muted);
  font-size: clamp(0.72rem, 3.9vw, 0.82rem);
  font-weight: 700;
  line-height: 1.45;
}

.ave-section__phones {
  display: block;
  width: min(100%, var(--app-width, 414px));
  margin: 28px 0 16px 50%;
  transform: translateX(-50%);
}

.ave-feature-list,
.ave-flat-list,
.ave-trust-card,
.ave-toggle-card {
  border: 1px solid rgba(143, 146, 160, 0.5);
  background: linear-gradient(135deg, rgba(40, 43, 55, 0.8), rgba(21, 24, 34, 0.72));
  box-shadow: inset 0 0 80px rgba(255, 255, 255, 0.02);
}

.ave-feature-list {
  display: grid;
  gap: 14px;
  margin-top: 14px;
  border: 0;
  background: transparent;
  box-shadow: none;
}

.ave-feature-card {
  display: flex;
  align-items: center;
  gap: 14px;
  min-height: 84px;
  padding: 12px 18px;
  text-align: left;
  border: 1px solid rgba(143, 146, 160, 0.5);
  border-radius: 16px;
  background:
    linear-gradient(90deg, rgba(50, 53, 66, 0.78), rgba(24, 27, 38, 0.72)),
    url('../../../assets/home/boxbg.png') center / cover no-repeat;
}

.ave-feature-card img {
  flex: 0 0 auto;
  width: 42px;
  height: 42px;
}

.ave-feature-card h3,
.ave-flat-item h3,
.ave-toggle-item h3,
.ave-trust-item h3 {
  margin: 0 0 6px;
  color: var(--text);
  font-size: clamp(0.9rem, 4.7vw, 1rem);
  font-weight: 800;
  line-height: 1.18;
}

.ave-feature-card p,
.ave-flat-item p,
.ave-toggle-item p,
.ave-trust-item p {
  margin: 0;
  color: var(--home-muted);
  font-size: clamp(0.62rem, 3.5vw, 0.72rem);
  font-weight: 700;
  line-height: 1.42;
}

.ave-section--strategy {
  padding-top: 48px;
}

.ave-strategy-scene {
  position: relative;
  min-height: 360px;
  margin-top: 20px;
}

.ave-strategy-scene__orders {
  position: absolute;
  left: 50%;
  top: 0;
  z-index: 2;
  width: min(100%, var(--app-width, 414px));
  transform: translateX(-50%);
}

.ave-strategy-scene__chart {
  position: absolute;
  left: 7%;
  bottom: -8px;
  z-index: 1;
  width: 218px;
}

.ave-flat-list {
  display: grid;
  gap: 0;
  margin-top: 42px;
  border: 0;
  background: transparent;
}

.ave-flat-item,
.ave-trust-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 18px 2px;
  text-align: left;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
}

.ave-flat-item:first-child {
  border-top: 0;
}

.ave-flat-item img,
.ave-trust-item img {
  flex: 0 0 auto;
  width: 52px;
  height: 52px;
}

.ave-section--contract {
  padding-top: 44px;
}

.ave-contract__phones {
  display: block;
  width: min(100%, var(--app-width, 414px));
  margin: 26px 0 20px 50%;
  transform: translateX(-50%);
}

.ave-toggle-card {
  display: grid;
  gap: 0;
  padding: 14px 18px;
  border-radius: 16px;
  text-align: left;
}

.ave-toggle-item {
  display: grid;
  grid-template-columns: 30px minmax(0, 1fr);
  gap: 12px;
  align-items: start;
  padding: 15px 0;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
}

.ave-toggle-item:first-child {
  border-top: 0;
}

.ave-toggle-item__switch {
  position: relative;
  width: 25px;
  height: 16px;
  margin-top: 3px;
  border-radius: 999px;
  background: var(--home-green);
}

.ave-toggle-item__switch::after {
  content: '';
  position: absolute;
  top: 0;
  right: 0;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: #fff;
}

.ave-section--trust {
  padding-top: 48px;
}

.ave-trust-card {
  display: grid;
  gap: 0;
  margin-top: 26px;
  padding: 12px 18px;
  border-radius: 16px;
}

.ave-trust-item {
  padding: 18px 0;
}

.ave-trust-item:first-child {
  border-top: 0;
}

.ave-reward {
  padding-top: 50px;
  padding-bottom: calc(108px + env(safe-area-inset-bottom));
  text-align: center;
}

.ave-reward h2 {
  max-width: 11ch;
  margin-inline: auto;
  font-size: clamp(1.5rem, 8.4vw, 1.72rem);
}

.ave-reward > p {
  color: var(--text);
  font-size: 0.9rem;
}

.ave-reward__image {
  display: block;
  width: min(100%, var(--app-width, 414px));
  margin: 34px 0 0 50%;
  transform: translateX(-50%);
}
</style>
