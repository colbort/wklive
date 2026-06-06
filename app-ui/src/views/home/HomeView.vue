<script setup lang="ts">
import { RouterLink } from 'vue-router'

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
</script>

<template>
  <section class="ave-home">
    <section class="ave-hero">
      <div class="ave-hero__copy">
        <h1>
          <span>{{ t('home.heroPrefix') }}</span>
          {{ t('home.heroSuffix') }}
        </h1>
        <p>{{ t('home.heroSub') }}</p>
        <RouterLink to="/markets" class="ave-home__cta ave-hero__cta">
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
      <RouterLink to="/markets" class="ave-home__cta">
        {{ t('home.startTrading') }}
      </RouterLink>
      <img class="ave-reward__image" :src="rewardImage" alt="">
    </section>
  </section>
</template>

<style scoped>
.ave-home {
  --home-green: #00b90f;
  --home-bg: #0b0e18;
  --home-muted: rgba(255, 255, 255, 0.55);
  width: min(100%, 414px);
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
  color: #fff;
}

.ave-home span {
  color: var(--home-green);
}

.ave-hero,
.ave-section,
.ave-reward {
  position: relative;
  padding-inline: 20px;
}

.ave-hero {
  min-height: 682px;
  padding-top: 46px;
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
  color: #fff;
  font-weight: 800;
  letter-spacing: 0;
}

.ave-hero h1 {
  display: grid;
  gap: 2px;
  font-size: clamp(38px, 10vw, 42px);
  line-height: 1.1;
}

.ave-hero p {
  margin: 22px 0 0;
  color: #fff;
  font-size: clamp(20px, 5.4vw, 23px);
  font-weight: 700;
}

.ave-home__cta {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 154px;
  height: 50px;
  margin-top: 24px;
  border-radius: 999px;
  background: var(--home-green);
  color: #fff;
  font-size: 24px;
  font-weight: 700;
  line-height: 1;
}

.ave-hero__cta {
  width: 154px;
  height: 50px;
  max-width: 68vw;
  margin-top: 24px;
  font-size: 24px;
}

.ave-hero__portal {
  position: absolute;
  left: 50%;
  width: 414px;
  max-width: none;
  transform: translateX(-50%);
}

.ave-section {
  padding-top: 34px;
  text-align: center;
}

.ave-section h2,
.ave-reward h2 {
  font-size: clamp(28px, 7.8vw, 32px);
  line-height: 1.22;
}

.ave-section > p,
.ave-reward > p {
  margin: 12px 0 0;
  color: var(--home-muted);
  font-size: clamp(16px, 4.3vw, 18px);
  font-weight: 700;
  line-height: 1.45;
}

.ave-section__phones {
  display: block;
  width: 414px;
  max-width: none;
  margin: 32px 0 20px 50%;
  transform: translateX(-50%);
}

.ave-feature-list,
.ave-flat-list,
.ave-trust-card,
.ave-toggle-card {
  border: 2px solid rgba(143, 146, 160, 0.5);
  background: linear-gradient(135deg, rgba(40, 43, 55, 0.8), rgba(21, 24, 34, 0.72));
  box-shadow: inset 0 0 80px rgba(255, 255, 255, 0.02);
}

.ave-feature-list {
  display: grid;
  gap: 18px;
  margin-top: 18px;
  border: 0;
  background: transparent;
  box-shadow: none;
}

.ave-feature-card {
  display: flex;
  align-items: center;
  gap: 16px;
  min-height: 94px;
  padding: 14px 22px;
  text-align: left;
  border: 2px solid rgba(143, 146, 160, 0.5);
  border-radius: 20px;
  background:
    linear-gradient(90deg, rgba(50, 53, 66, 0.78), rgba(24, 27, 38, 0.72)),
    url('../../../assets/home/boxbg.png') center / cover no-repeat;
}

.ave-feature-card img {
  flex: 0 0 auto;
  width: 46px;
  height: 46px;
}

.ave-feature-card h3,
.ave-flat-item h3,
.ave-toggle-item h3,
.ave-trust-item h3 {
  margin: 0 0 8px;
  color: #fff;
  font-size: clamp(20px, 5.2vw, 22px);
  font-weight: 800;
  line-height: 1.18;
}

.ave-feature-card p,
.ave-flat-item p,
.ave-toggle-item p,
.ave-trust-item p {
  margin: 0;
  color: var(--home-muted);
  font-size: clamp(14px, 3.9vw, 16px);
  font-weight: 700;
  line-height: 1.42;
}

.ave-section--strategy {
  padding-top: 56px;
}

.ave-strategy-scene {
  position: relative;
  min-height: 388px;
  margin-top: 24px;
}

.ave-strategy-scene__orders {
  position: absolute;
  left: 50%;
  top: 0;
  z-index: 2;
  width: 414px;
  max-width: none;
  transform: translateX(-50%);
}

.ave-strategy-scene__chart {
  position: absolute;
  left: 7%;
  bottom: -30px;
  z-index: 1;
  width: 236px;
}

.ave-flat-list {
  display: grid;
  gap: 0;
  margin-top: 50px;
  border: 0;
  background: transparent;
}

.ave-flat-item,
.ave-trust-item {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 22px 4px;
  text-align: left;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
}

.ave-flat-item:first-child {
  border-top: 0;
}

.ave-flat-item img,
.ave-trust-item img {
  flex: 0 0 auto;
  width: 58px;
  height: 58px;
}

.ave-section--contract {
  padding-top: 52px;
}

.ave-contract__phones {
  display: block;
  width: 414px;
  max-width: none;
  margin: 30px 0 24px 50%;
  transform: translateX(-50%);
}

.ave-toggle-card {
  display: grid;
  gap: 0;
  padding: 16px 20px;
  border-radius: 20px;
  text-align: left;
}

.ave-toggle-item {
  display: grid;
  grid-template-columns: 34px minmax(0, 1fr);
  gap: 14px;
  align-items: start;
  padding: 18px 0;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
}

.ave-toggle-item:first-child {
  border-top: 0;
}

.ave-toggle-item__switch {
  position: relative;
  width: 28px;
  height: 18px;
  margin-top: 4px;
  border-radius: 999px;
  background: var(--home-green);
}

.ave-toggle-item__switch::after {
  content: '';
  position: absolute;
  top: 0;
  right: 0;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: #fff;
}

.ave-section--trust {
  padding-top: 56px;
}

.ave-trust-card {
  display: grid;
  gap: 0;
  margin-top: 32px;
  padding: 14px 20px;
  border-radius: 20px;
}

.ave-trust-item {
  padding: 22px 0;
}

.ave-trust-item:first-child {
  border-top: 0;
}

.ave-reward {
  padding-top: 58px;
  padding-bottom: calc(120px + env(safe-area-inset-bottom));
  text-align: center;
}

.ave-reward h2 {
  max-width: 11ch;
  margin-inline: auto;
  font-size: clamp(34px, 9.4vw, 39px);
}

.ave-reward > p {
  color: #fff;
  font-size: 20px;
}

.ave-reward__image {
  display: block;
  width: 414px;
  max-width: none;
  margin: 40px 0 0 50%;
  transform: translateX(-50%);
}

@media (min-width: 900px) {
  .ave-home {
    border-inline: 1px solid rgba(255, 255, 255, 0.04);
  }

  .ave-hero {
    min-height: 682px;
  }
}
</style>
