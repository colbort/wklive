<script setup lang="ts">
import { RouterLink } from 'vue-router'

const featureCards = [
  {
    icon: '⟆',
    title: '深度流动性',
    desc: '享受极低滑点，轻松完成大额交易',
  },
  {
    icon: '↗',
    title: '专业级图表工具',
    desc: '多维数据分析，精准把握市场趋势',
  },
  {
    icon: '$',
    title: '超低费用结构',
    desc: '节省成本，最大化您的收益',
  },
]

const strategyFeatures = [
  {
    icon: '✎',
    title: '智能订单类型',
    desc: '满足限价、止损、条件触发等复杂需求',
  },
  {
    icon: '⇆',
    title: '高级API交易',
    desc: '毫秒级响应，机构级执行效率',
  },
]

const footerColumns = [
  {
    title: '产品',
    links: ['股票', '外汇', '大宗商品', '加密货币', '期权合约'],
  },
  {
    title: '其他',
    links: ['订单中心', '资产中心', '用户中心'],
  },
  {
    title: '社交媒体&社区',
    links: ['Twitter', 'LinkedIn', 'Facebook', 'Telegram', 'YouTube'],
  },
]
</script>

<template>
  <section class="ave-home">
    <section class="ave-hero">
      <div class="ave-hero__copy">
        <h1>
          <span class="ave-home__green">从股票到加密</span>，
          <br />
          一个平台全搞定
        </h1>
        <p>安心托付，交易致远</p>
        <RouterLink to="/markets" class="ave-home__cta">开始交易</RouterLink>
      </div>

      <div class="ave-hero__visual">
        <div class="ave-portal">
          <div class="ave-portal__ring" />
          <div class="ave-portal__inner" />
          <span class="ave-coin ave-coin--btc">₿</span>
          <span class="ave-coin ave-coin--usdt">₮</span>
          <span class="ave-coin ave-coin--spiral">◎</span>
          <div class="ave-portal__base ave-portal__base--top" />
          <div class="ave-portal__base ave-portal__base--bottom" />
        </div>
      </div>
    </section>

    <section class="ave-section ave-section--feature">
      <h2>
        更明智的<span class="ave-home__green">交易</span>，从这里开始
      </h2>
      <p class="ave-section__sub">24小时交易，涨跌都能赚</p>

      <div class="ave-feature-grid">
        <article v-for="item in featureCards" :key="item.title" class="ave-feature-card">
          <div class="ave-feature-card__icon">{{ item.icon }}</div>
          <div>
            <h3>{{ item.title }}</h3>
            <p>{{ item.desc }}</p>
          </div>
        </article>
      </div>
    </section>

    <section class="ave-section ave-section--strategy">
      <h2>
        高级功能支持<span class="ave-home__green">多样化</span>策略
      </h2>
      <p class="ave-section__sub">从美股到比特币，投资从未如此简单</p>

      <div class="ave-strategy__top">
        <article
          v-for="item in strategyFeatures"
          :key="item.title"
          class="ave-strategy__info"
        >
          <div class="ave-strategy__icon">{{ item.icon }}</div>
          <div>
            <h3>{{ item.title }}</h3>
            <p>{{ item.desc }}</p>
          </div>
        </article>
      </div>

      <div class="ave-strategy__scene">
        <div class="ave-strategy__line" />
        <div class="ave-phone ave-phone--left">
          <div class="ave-phone__screen">
            <div class="ave-phone__title">Transact</div>
            <div class="ave-phone__chart" />
            <div class="ave-phone__rows">
              <span />
              <span />
              <span />
              <span />
            </div>
            <button>Buy</button>
          </div>
        </div>

        <div class="ave-phone ave-phone--right">
          <div class="ave-phone__screen">
            <div class="ave-phone__title">BTC/USDT</div>
            <div class="ave-phone__tabs">
              <span class="active">15min</span>
              <span>30min</span>
              <span>60min</span>
            </div>
            <div class="ave-phone__trade-card">
              <div class="ave-phone__field" />
              <div class="ave-phone__field short" />
            </div>
            <div class="ave-phone__actions">
              <button class="green">Buy the rise</button>
              <button class="red">Buy the fall</button>
            </div>
          </div>
        </div>

        <div class="ave-badge ave-badge--btc">₿</div>
        <div class="ave-badge ave-badge--usdt">₮</div>
        <div class="ave-strategy__panel">
          <div class="ave-strategy__panel-image" />
        </div>
      </div>

      <div class="ave-strategy__bottom">
        <article
          v-for="item in strategyFeatures"
          :key="`${item.title}-mobile`"
          class="ave-strategy__list-item"
        >
          <div class="ave-strategy__icon">{{ item.icon }}</div>
          <div>
            <h3>{{ item.title }}</h3>
            <p>{{ item.desc }}</p>
          </div>
        </article>
      </div>
    </section>

    <section class="ave-footer">
      <div class="ave-footer__divider" />

      <div class="ave-footer__head">
        <div class="site-brand site-brand--footer">
          <span class="site-brand__mark">A</span>
          <div>
            <strong>AVE</strong>
          </div>
        </div>

        <button class="ave-lang-switch">🌐 中文简体</button>
      </div>

      <div class="ave-footer__grid">
        <article class="ave-download-card">
          <h3>APP下载</h3>
          <p>扫码下载手机客户端，随时随地交易</p>
          <div class="ave-download-card__codes">
            <div>
              <div class="ave-qr" />
              <span>IOS</span>
            </div>
            <div>
              <div class="ave-qr" />
              <span>Android</span>
            </div>
          </div>
        </article>

        <div class="ave-footer__links">
          <article v-for="column in footerColumns" :key="column.title" class="ave-footer__column">
            <h4>{{ column.title }}</h4>
            <RouterLink v-for="item in column.links" :key="item" to="/home">
              {{ item }}
            </RouterLink>
          </article>
        </div>
      </div>

      <p class="ave-footer__copyright">
        © 2025 AVE交易所. 保留所有权利 | Copyright © 2022-2026
      </p>
    </section>
  </section>
</template>

<style scoped>
.ave-home {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.ave-home__green {
  color: #08c200;
}

.ave-hero,
.ave-section,
.ave-footer {
  position: relative;
  overflow: hidden;
  border-radius: 0 0 28px 28px;
  background:
    linear-gradient(90deg, rgba(47, 57, 52, 0.56) 0%, rgba(18, 21, 31, 0.12) 24%, rgba(14, 16, 27, 0) 55%),
    linear-gradient(rgba(120, 167, 216, 0.08) 1px, transparent 1px),
    linear-gradient(90deg, rgba(120, 167, 216, 0.08) 1px, transparent 1px),
    linear-gradient(180deg, #0d0f1a 0%, #10131c 100%);
  background-size:
    auto,
    100% 206px,
    206px 100%,
    auto;
}

.ave-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(420px, 0.92fr);
  align-items: center;
  min-height: 980px;
  padding: 180px 72px 110px;
}

.ave-hero__copy {
  position: relative;
  z-index: 2;
}

.ave-hero__copy h1 {
  margin: 0;
  font-size: clamp(74px, 6.2vw, 108px);
  line-height: 1.08;
  font-weight: 800;
  letter-spacing: -0.04em;
}

.ave-hero__copy p {
  margin: 38px 0 0;
  color: #f5f7fb;
  font-size: clamp(28px, 2vw, 42px);
}

.ave-home__cta {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 292px;
  height: 102px;
  margin-top: 76px;
  border-radius: 999px;
  background: #0ab300;
  color: #fff;
  font-size: 34px;
  font-weight: 700;
}

.ave-hero__visual {
  display: flex;
  justify-content: center;
  align-items: center;
}

.ave-portal {
  position: relative;
  width: min(100%, 760px);
  aspect-ratio: 1 / 1;
}

.ave-portal__ring {
  position: absolute;
  inset: 12% 8% 10% 14%;
  border-radius: 50%;
  border: 24px solid #f0f0ef;
  box-shadow:
    inset 0 0 0 10px #d8ece2,
    0 0 0 2px rgba(255, 255, 255, 0.18);
}

.ave-portal__inner {
  position: absolute;
  inset: 24% 20% 18% 26%;
  border-radius: 50%;
  background:
    radial-gradient(circle at 68% 66%, rgba(33, 255, 127, 0.95), rgba(17, 106, 52, 0.15) 18%, transparent 28%),
    radial-gradient(circle at 36% 38%, rgba(55, 255, 151, 0.34), transparent 18%),
    radial-gradient(circle at 55% 60%, rgba(16, 255, 117, 0.12), transparent 55%),
    linear-gradient(135deg, #08140d, #0a3a1e 48%, #18c85e 100%);
  box-shadow: inset 0 0 100px rgba(0, 0, 0, 0.45);
}

.ave-portal::before,
.ave-portal::after {
  content: '';
  position: absolute;
  border-radius: 50%;
  border: 2px solid rgba(15, 161, 62, 0.12);
}

.ave-portal::before {
  inset: 2% 0 0 10%;
}

.ave-portal::after {
  inset: 10% -6% -8% 6%;
}

.ave-coin,
.ave-badge {
  position: absolute;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: #f1f2f3;
  color: #12b93b;
  font-weight: 800;
  box-shadow: 0 10px 28px rgba(0, 0, 0, 0.28);
}

.ave-coin--btc {
  right: 16%;
  top: 44%;
  width: 118px;
  height: 118px;
  font-size: 66px;
}

.ave-coin--usdt {
  left: 18%;
  top: 54%;
  width: 86px;
  height: 86px;
  font-size: 44px;
}

.ave-coin--spiral {
  left: 2%;
  bottom: 18%;
  width: 104px;
  height: 104px;
  font-size: 54px;
}

.ave-portal__base {
  position: absolute;
  background: #f3f3f2;
  border: 2px solid rgba(0, 0, 0, 0.16);
}

.ave-portal__base--top {
  right: 14%;
  bottom: 9%;
  width: 240px;
  height: 66px;
  transform: skew(-26deg);
}

.ave-portal__base--bottom {
  right: 26%;
  bottom: 2%;
  width: 230px;
  height: 38px;
  transform: skew(-26deg);
}

.ave-section {
  padding: 110px 56px 80px;
}

.ave-section h2 {
  margin: 0;
  text-align: center;
  font-size: clamp(56px, 4vw, 84px);
  font-weight: 800;
  line-height: 1.18;
}

.ave-section__sub {
  margin: 26px 0 0;
  text-align: center;
  color: #8f919b;
  font-size: clamp(22px, 1.7vw, 30px);
  font-weight: 600;
}

.ave-feature-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 26px;
  margin-top: 86px;
}

.ave-feature-card {
  display: flex;
  gap: 26px;
  align-items: center;
  min-height: 230px;
  padding: 36px 32px;
  border-radius: 34px;
  border: 2px solid rgba(121, 124, 140, 0.5);
  background: linear-gradient(180deg, rgba(38, 40, 51, 0.82), rgba(28, 30, 41, 0.76));
}

.ave-feature-card__icon,
.ave-strategy__icon {
  flex: none;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  color: #08c200;
}

.ave-feature-card__icon {
  width: 104px;
  height: 104px;
  font-size: 66px;
}

.ave-feature-card h3,
.ave-strategy__info h3,
.ave-strategy__list-item h3 {
  margin: 0 0 16px;
  font-size: clamp(28px, 2vw, 40px);
  font-weight: 700;
}

.ave-feature-card p,
.ave-strategy__info p,
.ave-strategy__list-item p {
  margin: 0;
  color: #9b9da6;
  font-size: clamp(18px, 1.4vw, 28px);
  line-height: 1.6;
  font-weight: 600;
}

.ave-section--strategy {
  padding-bottom: 110px;
}

.ave-strategy__top {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 42px;
  margin-top: 88px;
}

.ave-strategy__info {
  display: flex;
  align-items: center;
  gap: 26px;
  padding: 0 12px;
}

.ave-strategy__icon {
  width: 112px;
  height: 112px;
  background: rgba(38, 40, 51, 0.92);
  border: 2px solid rgba(121, 124, 140, 0.34);
  font-size: 58px;
}

.ave-strategy__scene {
  position: relative;
  min-height: 900px;
  margin-top: 44px;
}

.ave-strategy__line {
  position: absolute;
  left: -8%;
  right: -8%;
  top: 36%;
  height: 420px;
  border-top: 4px solid rgba(8, 194, 0, 0.9);
  border-bottom: 4px solid transparent;
  border-radius: 50%;
  transform: rotate(-8deg);
  opacity: 0.82;
}

.ave-phone {
  position: absolute;
  width: 290px;
  height: 592px;
  padding: 18px;
  border-radius: 48px;
  background: linear-gradient(180deg, #23252d, #0d1018);
  box-shadow: 0 24px 60px rgba(0, 0, 0, 0.38);
}

.ave-phone--left {
  left: 8%;
  bottom: 90px;
  transform: rotate(-14deg);
}

.ave-phone--right {
  left: 34%;
  top: 50px;
  transform: rotate(7deg);
}

.ave-phone__screen {
  height: 100%;
  padding: 18px;
  border-radius: 34px;
  background: linear-gradient(180deg, rgba(18, 22, 32, 0.96), rgba(14, 17, 24, 0.98));
}

.ave-phone__title {
  color: #f5f7fb;
  font-size: 28px;
  font-weight: 700;
}

.ave-phone__chart {
  height: 150px;
  margin-top: 22px;
  border-radius: 22px;
  background:
    linear-gradient(180deg, transparent 0 56%, rgba(31, 255, 109, 0.3) 56% 58%, transparent 58% 100%),
    linear-gradient(135deg, rgba(31, 255, 109, 0.08), rgba(255, 255, 255, 0.02));
}

.ave-phone__rows {
  display: grid;
  gap: 14px;
  margin-top: 20px;
}

.ave-phone__rows span,
.ave-phone__field {
  height: 38px;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.08);
}

.ave-phone button {
  width: 100%;
  height: 54px;
  margin-top: 24px;
  border: 0;
  border-radius: 16px;
  background: #16d45a;
  color: #fff;
  font-size: 22px;
  font-weight: 700;
}

.ave-phone__tabs {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
  margin-top: 18px;
}

.ave-phone__tabs span {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 40px;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.06);
  color: #a4a7b1;
  font-size: 14px;
}

.ave-phone__tabs span.active {
  background: rgba(8, 194, 0, 0.18);
  color: #18dc55;
}

.ave-phone__trade-card {
  display: grid;
  gap: 12px;
  margin-top: 22px;
}

.ave-phone__field.short {
  width: 50%;
}

.ave-phone__actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
  margin-top: 22px;
}

.ave-phone__actions button {
  margin-top: 0;
  font-size: 16px;
}

.ave-phone__actions .green {
  background: #16d45a;
}

.ave-phone__actions .red {
  background: #ff614e;
}

.ave-badge {
  width: 100px;
  height: 100px;
  background: #08b800;
  color: #041108;
  font-size: 62px;
}

.ave-badge--btc {
  left: 14%;
  top: 58%;
}

.ave-badge--usdt {
  left: 42%;
  top: 47%;
}

.ave-strategy__panel {
  position: absolute;
  left: 8%;
  bottom: 130px;
  width: 440px;
  padding: 18px;
  border-radius: 40px;
  border: 2px solid rgba(121, 124, 140, 0.34);
  background: linear-gradient(180deg, rgba(34, 37, 47, 0.82), rgba(27, 29, 39, 0.84));
}

.ave-strategy__panel-image {
  aspect-ratio: 1.55 / 1;
  border-radius: 28px;
  background:
    radial-gradient(circle at 68% 42%, rgba(7, 255, 115, 0.3), transparent 18%),
    radial-gradient(circle at 35% 38%, rgba(255, 255, 255, 0.18), transparent 10%),
    linear-gradient(180deg, rgba(6, 9, 13, 0.34), rgba(7, 9, 14, 0.14)),
    linear-gradient(135deg, #040608, #10361e 60%, #0b0d12 100%);
}

.ave-strategy__bottom {
  display: none;
}

.ave-footer {
  padding: 0 56px 28px;
}

.ave-footer__divider {
  height: 8px;
  margin: 0 -56px 54px;
  background: linear-gradient(90deg, #5f6ea6 0 52%, #b4b7de 52% 100%);
}

.ave-footer__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}

.site-brand--footer {
  transform: scale(1.06);
  transform-origin: left center;
}

.ave-lang-switch {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  height: 56px;
  padding: 0 18px;
  border: 0;
  border-radius: 28px;
  background: rgba(30, 33, 44, 0.92);
  color: #f5f7fb;
  font-size: 18px;
}

.ave-footer__grid {
  display: grid;
  grid-template-columns: 460px minmax(0, 1fr);
  gap: 70px;
  margin-top: 54px;
}

.ave-download-card {
  padding: 36px;
  border-radius: 28px;
  background: rgba(30, 33, 44, 0.92);
}

.ave-download-card h3 {
  margin: 0;
  font-size: 28px;
}

.ave-download-card p {
  margin: 28px 0 0;
  color: #dfe1e7;
  font-size: 20px;
  line-height: 1.7;
}

.ave-download-card__codes {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18px;
  margin-top: 32px;
  text-align: center;
}

.ave-download-card__codes span {
  display: block;
  margin-top: 14px;
  font-size: 20px;
  font-weight: 600;
}

.ave-qr {
  aspect-ratio: 1;
  border-radius: 18px;
  background:
    linear-gradient(90deg, #000 12%, transparent 12% 16%, #000 16% 20%, transparent 20% 28%, #000 28% 34%, transparent 34% 100%),
    linear-gradient(#000 12%, transparent 12% 16%, #000 16% 20%, transparent 20% 28%, #000 28% 34%, transparent 34% 100%),
    #fff;
}

.ave-footer__links {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 34px;
  padding-top: 8px;
}

.ave-footer__column {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.ave-footer__column h4 {
  margin: 0 0 18px;
  font-size: 28px;
}

.ave-footer__column a {
  color: #e8eaf0;
  font-size: 20px;
}

.ave-footer__copyright {
  margin: 54px 0 0;
  padding-top: 18px;
  border-top: 1px solid rgba(121, 124, 140, 0.18);
  text-align: center;
  color: #888b97;
  font-size: 16px;
}

@media (max-width: 959px) {
  .ave-hero,
  .ave-section,
  .ave-footer {
    border-radius: 0;
  }

  .ave-hero {
    grid-template-columns: 1fr;
    min-height: auto;
    padding: 132px 22px 48px;
  }

  .ave-hero__copy {
    order: 1;
    text-align: center;
  }

  .ave-hero__copy h1 {
    font-size: clamp(42px, 9.4vw, 72px);
    line-height: 1.1;
  }

  .ave-hero__copy p {
    margin-top: 26px;
    font-size: 26px;
  }

  .ave-home__cta {
    min-width: 0;
    width: 282px;
    height: 92px;
    margin-top: 34px;
    font-size: 28px;
  }

  .ave-hero__visual {
    order: 2;
    margin-top: 40px;
  }

  .ave-portal {
    width: 100%;
    max-width: 560px;
    aspect-ratio: 0.96 / 1;
  }

  .ave-portal__ring {
    inset: 10% 4% 14% 13%;
    border-width: 18px;
  }

  .ave-portal__inner {
    inset: 22% 14% 22% 22%;
  }

  .ave-coin--btc {
    width: 86px;
    height: 86px;
    font-size: 46px;
  }

  .ave-coin--usdt,
  .ave-coin--spiral {
    width: 68px;
    height: 68px;
    font-size: 34px;
  }

  .ave-portal__base--top {
    right: 15%;
    width: 180px;
    height: 48px;
  }

  .ave-portal__base--bottom {
    right: 32%;
    width: 152px;
    height: 26px;
  }

  .ave-section {
    padding: 62px 22px 32px;
  }

  .ave-section h2 {
    font-size: clamp(30px, 7.8vw, 52px);
  }

  .ave-section__sub {
    margin-top: 18px;
    font-size: 18px;
  }

  .ave-feature-grid {
    grid-template-columns: 1fr;
    gap: 18px;
    margin-top: 40px;
  }

  .ave-feature-card {
    min-height: 0;
    padding: 24px 20px;
    border-radius: 22px;
    gap: 16px;
  }

  .ave-feature-card__icon,
  .ave-strategy__icon {
    width: 72px;
    height: 72px;
    font-size: 38px;
  }

  .ave-feature-card h3,
  .ave-strategy__info h3,
  .ave-strategy__list-item h3 {
    margin-bottom: 10px;
    font-size: 24px;
  }

  .ave-feature-card p,
  .ave-strategy__info p,
  .ave-strategy__list-item p {
    font-size: 16px;
  }

  .ave-strategy__top {
    display: none;
  }

  .ave-section--strategy {
    padding-bottom: 44px;
  }

  .ave-strategy__scene {
    min-height: 760px;
    margin-top: 34px;
  }

  .ave-strategy__line {
    top: 32%;
    left: -42%;
    right: -42%;
    height: 300px;
    transform: rotate(2deg);
  }

  .ave-phone {
    width: 180px;
    height: 372px;
    padding: 12px;
    border-radius: 28px;
  }

  .ave-phone--left {
    left: 6%;
    top: 120px;
    bottom: auto;
  }

  .ave-phone--right {
    left: 44%;
    top: 36px;
  }

  .ave-phone__screen {
    padding: 12px;
    border-radius: 22px;
  }

  .ave-phone__title {
    font-size: 18px;
  }

  .ave-phone__chart {
    height: 88px;
    margin-top: 14px;
  }

  .ave-phone__rows {
    gap: 9px;
    margin-top: 14px;
  }

  .ave-phone__rows span,
  .ave-phone__field {
    height: 22px;
  }

  .ave-phone button {
    height: 34px;
    margin-top: 14px;
    border-radius: 10px;
    font-size: 14px;
  }

  .ave-phone__tabs {
    gap: 8px;
    margin-top: 12px;
  }

  .ave-phone__tabs span {
    height: 26px;
    border-radius: 10px;
    font-size: 10px;
  }

  .ave-phone__actions {
    gap: 8px;
    margin-top: 14px;
  }

  .ave-phone__actions button {
    font-size: 10px;
  }

  .ave-badge {
    width: 70px;
    height: 70px;
    font-size: 42px;
  }

  .ave-badge--btc {
    left: 64%;
    top: 56%;
  }

  .ave-badge--usdt {
    left: 80%;
    top: 24%;
  }

  .ave-strategy__panel {
    left: 10%;
    right: 10%;
    bottom: 136px;
    width: auto;
    padding: 14px;
    border-radius: 26px;
  }

  .ave-strategy__bottom {
    display: grid;
    gap: 0;
    margin-top: 34px;
  }

  .ave-strategy__list-item {
    display: flex;
    gap: 16px;
    align-items: center;
    padding: 24px 0;
    border-top: 1px solid rgba(121, 124, 140, 0.18);
  }

  .ave-strategy__list-item:first-child {
    border-top: 0;
  }

  .ave-footer {
    padding: 0 22px 22px;
  }

  .ave-footer__divider {
    margin: 0 -22px 28px;
  }

  .ave-footer__head {
    align-items: flex-start;
    flex-direction: column;
  }

  .ave-footer__grid {
    grid-template-columns: 1fr;
    gap: 30px;
    margin-top: 28px;
  }

  .ave-download-card {
    padding: 22px;
  }

  .ave-download-card h3,
  .ave-footer__column h4 {
    font-size: 24px;
  }

  .ave-download-card p,
  .ave-footer__column a,
  .ave-download-card__codes span {
    font-size: 16px;
  }

  .ave-footer__links {
    grid-template-columns: 1fr;
    gap: 24px;
  }

  .ave-footer__column {
    gap: 14px;
  }

  .ave-footer__copyright {
    margin-top: 30px;
    font-size: 13px;
  }
}
</style>
