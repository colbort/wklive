<script setup lang="ts">
import boxBg from '../../assets/home/boxbg.png'
import icon1 from '../../assets/home/icon1.png'
import icon2 from '../../assets/home/icon2.png'
import icon3 from '../../assets/home/icon3.png'
import icon4 from '../../assets/home/icon4.png'
import icon5 from '../../assets/home/icon5.png'
import heroImg from '../../assets/home/img1.png'
import toolsImg from '../../assets/home/img2.png'
import tradeImg from '../../assets/home/img3.png'
import posterImg from '../../assets/home/img4.png'
import trustImg from '../../assets/home/img5.png'
import rewardImg from '../../assets/home/img6.png'

const smartCards = [
  {
    title: '深度流动性',
    text: '享受极低滑点，轻松完成大额交易',
    icon: icon1,
  },
  {
    title: '专业级图表工具',
    text: '多维数据分析，精准把握市场趋势',
    icon: icon2,
    featured: true,
  },
  {
    title: '超低费用结构',
    text: '节省成本，最大化您的收益',
    icon: icon3,
  },
]

const strategyItems = [
  {
    title: '智能订单类型',
    text: '满足限价、止损、条件触发等复杂需求',
    icon: icon4,
  },
  {
    title: '高级API交易',
    text: '毫秒级响应，机构级执行效率',
    icon: icon5,
  },
]

const futuresFeatures = [
  ['无到期日，灵活持仓', '摆脱交割压力，避免移仓损耗'],
  ['7x24小时全球交易', '全天候机会，无开盘跳空'],
  ['多空双向交易', '熊市也能赚钱，对冲现货风险'],
  ['深度滑动型和低滑点', '主流币种买卖价差小，大单对价格冲击小'],
]

const trustCards = [
  ['极简交易体验', '全平台支持 - 桌面端与移动端，随时随地轻松买卖加密货币'],
  ['资产安全 全力守护', '3亿USDT用户保护基金 -- 行业领先保障，为您的交易保驾护航'],
  ['全天候专业支持', '24/7 客服 - 屡获殊荣的团队，随时解答您的疑问'],
]

const footerProducts = ['股票', '外汇', '大宗商品', '加密货币', '期权合约']
const footerOther = ['订单中心', '资产中心', '用户中心']
const socials = ['Twitter', 'LinkedIn', 'Facebook', 'Telegram', 'YouTube']
</script>

<template>
  <div class="home-page">
    <section class="hero-section grid-bg">
      <div class="hero-copy">
        <h1><span>从股票到加密</span>，<br>平台全搞定</h1>
        <p>安心托付，交易致远</p>
        <RouterLink to="/trades" class="primary-cta">
          开始交易
        </RouterLink>
      </div>
      <img class="hero-art" :src="heroImg" alt="">
    </section>

    <section class="smart-section section-pad">
      <h2>更明智的<span>交易</span>，从这里开始</h2>
      <div class="smart-cards">
        <article
          v-for="card in smartCards"
          :key="card.title"
          class="smart-card"
          :class="{ 'smart-card--featured': card.featured }"
        >
          <img :src="card.icon" alt="">
          <h3>{{ card.title }}</h3>
          <p>{{ card.text }}</p>
        </article>
      </div>
    </section>

    <section class="strategy-section section-pad">
      <h2>高级功能支持<span>多样化</span>策略</h2>
      <div class="strategy-list">
        <article v-for="item in strategyItems" :key="item.title" class="strategy-item">
          <img :src="item.icon" alt="">
          <div>
            <h3>{{ item.title }}</h3>
            <p>{{ item.text }}</p>
          </div>
        </article>
      </div>
      <img class="tools-art" :src="toolsImg" alt="">
    </section>

    <section class="market-flow section-pad">
      <img class="trade-card trade-card--left" :src="tradeImg" alt="">
      <img class="poster-card" :src="posterImg" alt="">
      <h2><span>永续合约</span>，无到期日的期货<br>交易</h2>
    </section>

    <section class="futures-section section-pad">
      <div class="orbit-line" />
      <div class="futures-copy">
        <h2><span>永续合约</span>，无到期日的期货<br>交易</h2>
        <ul>
          <li v-for="[title, text] in futuresFeatures" :key="title">
            <span />
            <div>
              <strong>{{ title }}</strong>
              <p>{{ text }}</p>
            </div>
          </li>
        </ul>
      </div>
    </section>

    <section class="trust-section section-pad">
      <h2>专业、值得信赖的<span>综合交易所</span></h2>
      <p>全球数百万用户的共同选择</p>
      <div class="trust-cards">
        <article v-for="(card, index) in trustCards" :key="card[0]" class="trust-card">
          <img :src="index === 1 ? trustImg : boxBg" alt="">
          <h3>{{ card[0] }}</h3>
          <p>{{ card[1] }}</p>
        </article>
      </div>
    </section>

    <section class="reward-section section-pad">
      <div class="reward-copy">
        <h2>智享交易，安享<br>财富</h2>
        <p>新用户可得丰富奖励</p>
        <RouterLink to="/login" class="primary-cta">
          立即注册
        </RouterLink>
      </div>
      <img :src="rewardImg" alt="">
    </section>

    <footer class="site-footer">
      <div class="footer-line" />
      <div class="footer-inner">
        <div class="footer-brand">
          <div class="footer-logo">
            <span class="mini-mark"><span /></span>
            <strong>AVE</strong>
          </div>
          <button type="button" class="lang-button">
            <span>◎</span>
            中文简体
            <i />
          </button>
          <div class="download-card">
            <h3>APP下载</h3>
            <p>扫码下载手机端，随时随地交易</p>
            <div class="qr-row">
              <div>
                <span class="qr-code" />
                <em>IOS</em>
              </div>
              <div>
                <span class="qr-code" />
                <em>Android</em>
              </div>
            </div>
          </div>
        </div>
        <div class="footer-links">
          <div>
            <h3>产品</h3>
            <a v-for="item in footerProducts" :key="item" href="#">{{ item }}</a>
          </div>
          <div>
            <h3>其他</h3>
            <a v-for="item in footerOther" :key="item" href="#">{{ item }}</a>
          </div>
          <div>
            <h3>社交媒体&社区</h3>
            <a v-for="item in socials" :key="item" href="#">
              <span>{{ item.slice(0, 1) }}</span>{{ item }}
            </a>
          </div>
        </div>
      </div>
      <p class="copyright">
        © 2025 AVE交易所. 保留所有权利 Copyright © 2022-2026
      </p>
    </footer>

    <button class="support-button" type="button" aria-label="在线客服">
      <span />
    </button>
  </div>
</template>

<style scoped>
.home-page {
  position: relative;
  overflow: hidden;
  background:
    radial-gradient(circle at 9% 62%, rgb(77 105 110 / 16%), transparent 24%),
    linear-gradient(180deg, #0b0d16 0%, #0a0c15 100%);
  color: var(--text);
}

.section-pad {
  position: relative;
  padding: 96px 64px;
}

.grid-bg {
  background-image:
    linear-gradient(rgb(255 255 255 / 5%) 1px, transparent 1px),
    linear-gradient(90deg, rgb(255 255 255 / 5%) 1px, transparent 1px);
  background-position: -2px -2px;
  background-size: 205px 205px;
}

.hero-section {
  position: relative;
  display: flex;
  min-height: 930px;
  align-items: center;
  padding: 92px 64px 120px;
}

.hero-section::after {
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  height: 300px;
  background: linear-gradient(180deg, transparent, #0b0d16);
  content: '';
}

.hero-copy {
  position: relative;
  z-index: 2;
  width: 58%;
  padding-top: 30px;
}

h1,
h2,
h3,
p {
  margin: 0;
}

.hero-copy h1 {
  max-width: 1030px;
  font-size: 102px;
  font-weight: 950;
  line-height: 1.18;
}

.hero-copy h1 span,
h2 span {
  color: var(--accent);
}

.hero-copy p {
  margin-top: 48px;
  font-size: 46px;
  line-height: 1.3;
}

.primary-cta {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 290px;
  height: 104px;
  margin-top: 110px;
  border-radius: 999px;
  background: var(--accent);
  color: #fff;
  font-size: 40px;
  font-weight: 900;
}

.hero-art {
  position: absolute;
  z-index: 1;
  right: -18px;
  top: 155px;
  width: min(43vw, 750px);
  min-width: 610px;
}

.smart-section {
  min-height: 850px;
  padding-top: 44px;
}

.smart-section h2,
.strategy-section h2,
.trust-section h2 {
  text-align: center;
  font-size: 64px;
  font-weight: 950;
  line-height: 1.18;
}

.smart-cards {
  position: relative;
  display: grid;
  max-width: 1900px;
  margin: 370px auto 0;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  align-items: end;
  border: 1px solid rgb(255 255 255 / 24%);
  border-radius: 24px;
  background: linear-gradient(180deg, rgb(255 255 255 / 7%), rgb(255 255 255 / 2%));
}

.smart-card {
  min-height: 390px;
  padding: 64px 56px 48px;
}

.smart-card img {
  width: 112px;
  height: 112px;
  object-fit: contain;
}

.smart-card h3 {
  margin-top: 78px;
  font-size: 40px;
  line-height: 1.2;
}

.smart-card p {
  margin-top: 34px;
  color: var(--text-soft);
  font-size: 26px;
  font-weight: 800;
  line-height: 1.45;
}

.smart-card--featured {
  min-height: 675px;
  margin-top: -286px;
  border-radius: 24px 24px 0 0;
  background: var(--accent);
}

.smart-card--featured img {
  width: 132px;
  height: 132px;
  filter: brightness(0) invert(1);
}

.smart-card--featured h3,
.smart-card--featured p {
  color: #fff;
}

.strategy-section {
  min-height: 960px;
  padding-top: 164px;
}

.strategy-list {
  display: grid;
  max-width: 1480px;
  margin: 120px auto 0;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 160px;
}

.strategy-item {
  display: flex;
  align-items: center;
  gap: 28px;
}

.strategy-item img {
  width: 110px;
  height: 110px;
}

.strategy-item h3 {
  font-size: 40px;
  line-height: 1.2;
}

.strategy-item p {
  margin-top: 28px;
  color: var(--text-soft);
  font-size: 26px;
  font-weight: 800;
}

.tools-art {
  display: block;
  width: 730px;
  margin: 130px auto -260px;
  opacity: 0.86;
}

.market-flow {
  min-height: 900px;
}

.market-flow::before,
.futures-section .orbit-line {
  position: absolute;
  width: 1650px;
  height: 570px;
  border: 4px solid rgb(0 166 9 / 72%);
  border-color: rgb(0 166 9 / 72%) transparent transparent transparent;
  border-radius: 50%;
  content: '';
  pointer-events: none;
}

.market-flow::before {
  top: 82px;
  left: 50%;
  transform: translateX(-48%) rotate(4deg);
}

.trade-card,
.poster-card {
  position: absolute;
  opacity: 0.82;
}

.trade-card--left {
  left: 150px;
  top: 90px;
  width: 500px;
}

.poster-card {
  right: 290px;
  top: 170px;
  width: 540px;
}

.market-flow h2 {
  position: absolute;
  right: 124px;
  bottom: 60px;
  font-size: 58px;
  font-weight: 950;
  line-height: 1.35;
}

.futures-section {
  min-height: 930px;
}

.futures-section .orbit-line {
  top: 48px;
  left: -370px;
  height: 690px;
  transform: rotate(-18deg);
}

.futures-copy {
  width: 780px;
  margin-left: auto;
  margin-right: 300px;
}

.futures-copy h2 {
  font-size: 64px;
  font-weight: 950;
  line-height: 1.34;
}

.futures-copy ul {
  display: grid;
  gap: 52px;
  margin: 86px 0 0;
  padding: 0;
  list-style: none;
}

.futures-copy li {
  display: flex;
  gap: 16px;
}

.futures-copy li > span {
  flex: 0 0 auto;
  width: 32px;
  height: 20px;
  margin-top: 8px;
  border-radius: 999px;
  background: linear-gradient(90deg, var(--accent) 0 48%, #fff 48% 100%);
}

.futures-copy strong {
  display: block;
  font-size: 28px;
  line-height: 1.2;
}

.futures-copy p {
  margin-top: 16px;
  color: var(--text-soft);
  font-size: 22px;
  font-weight: 800;
}

.trust-section {
  min-height: 900px;
  padding-top: 180px;
}

.trust-section > p {
  margin-top: 26px;
  text-align: center;
  color: var(--text-soft);
  font-size: 28px;
  font-weight: 700;
}

.trust-cards {
  display: grid;
  max-width: 1920px;
  margin: 130px auto 0;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 38px;
}

.trust-card {
  min-height: 420px;
  padding: 76px 52px 44px;
  border: 1px solid rgb(255 255 255 / 20%);
  border-radius: 16px;
  background: linear-gradient(180deg, rgb(255 255 255 / 7%), rgb(255 255 255 / 2%));
  text-align: center;
}

.trust-card img {
  width: 122px;
  height: 122px;
  object-fit: contain;
}

.trust-card h3 {
  margin-top: 62px;
  font-size: 38px;
  line-height: 1.2;
}

.trust-card p {
  margin-top: 38px;
  color: var(--text-soft);
  font-size: 24px;
  font-weight: 800;
  line-height: 1.42;
}

.reward-section {
  display: flex;
  min-height: 760px;
  align-items: center;
  justify-content: space-between;
  gap: 60px;
  padding-top: 80px;
}

.reward-copy h2 {
  font-size: 78px;
  font-weight: 950;
  line-height: 1.35;
}

.reward-copy p {
  margin-top: 74px;
  font-size: 34px;
  font-weight: 800;
}

.reward-section img {
  width: min(48vw, 745px);
}

.site-footer {
  position: relative;
  padding: 0 64px 0;
  background: #0a0c15;
}

.footer-line {
  height: 6px;
  border-radius: 999px;
  background: linear-gradient(90deg, #c6c9df, #6075a8 88%, #c7c9da);
}

.footer-inner {
  display: grid;
  max-width: 1900px;
  margin: 0 auto;
  padding: 64px 0 48px;
  grid-template-columns: 560px 1fr;
  gap: 200px;
}

.footer-logo {
  display: flex;
  align-items: center;
  gap: 14px;
  font-size: 46px;
  font-weight: 950;
}

.mini-mark {
  position: relative;
  display: grid;
  width: 64px;
  height: 64px;
  border-radius: 14px;
  background: linear-gradient(135deg, #2ed8c6, #31bd86 48%, #8155e7);
}

.mini-mark::before {
  position: absolute;
  inset: 10px 13px;
  border-radius: 7px;
  background: #0a0c15;
  clip-path: polygon(50% 0, 100% 84%, 50% 57%, 0 84%);
  content: '';
}

.lang-button {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  margin-top: 72px;
  border: 0;
  background: transparent;
  color: var(--text);
  font-size: 20px;
}

.lang-button span {
  display: grid;
  width: 42px;
  height: 42px;
  place-items: center;
  border-radius: 50%;
  background: rgb(255 255 255 / 8%);
}

.lang-button i {
  width: 0;
  height: 0;
  border-top: 7px solid currentColor;
  border-right: 6px solid transparent;
  border-left: 6px solid transparent;
}

.download-card {
  width: 456px;
  margin-top: 68px;
  padding: 38px 36px;
  border-radius: 24px;
  background: rgb(29 32 44);
}

.download-card h3 {
  font-size: 28px;
}

.download-card p {
  margin-top: 28px;
  color: var(--text-soft);
  font-size: 22px;
  font-weight: 800;
}

.qr-row {
  display: grid;
  margin-top: 40px;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
}

.qr-row > div {
  display: grid;
  justify-items: center;
  gap: 18px;
}

.qr-code {
  display: block;
  width: 184px;
  height: 184px;
  border: 16px solid #fff;
  border-radius: 18px;
  background:
    linear-gradient(90deg, #000 12px, transparent 12px 22px, #000 22px 34px, transparent 34px) 0 0 /
      46px 46px,
    linear-gradient(#000 12px, transparent 12px 22px, #000 22px 34px, transparent 34px) 0 0 / 46px
      46px,
    #fff;
}

.qr-row em {
  color: var(--text);
  font-size: 22px;
  font-style: normal;
}

.footer-links {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 86px;
  padding-top: 96px;
}

.footer-links div {
  display: grid;
  align-content: start;
  gap: 32px;
}

.footer-links h3 {
  margin-bottom: 14px;
  font-size: 30px;
}

.footer-links a {
  display: inline-flex;
  align-items: center;
  gap: 14px;
  color: var(--text);
  font-size: 24px;
  font-weight: 700;
}

.footer-links a span {
  display: grid;
  width: 30px;
  height: 30px;
  place-items: center;
  border-radius: 50%;
  background: #0b8fea;
  font-size: 15px;
  font-weight: 950;
}

.copyright {
  margin: 0 -64px;
  padding: 28px 64px;
  border-top: 1px solid rgb(255 255 255 / 7%);
  color: rgb(255 255 255 / 45%);
  text-align: center;
  font-size: 18px;
  font-weight: 700;
}

.support-button {
  position: fixed;
  z-index: 30;
  right: 34px;
  bottom: 34px;
  display: grid;
  width: 130px;
  height: 130px;
  place-items: center;
  border: 0;
  border-radius: 28px;
  background: #9cb993;
  box-shadow: 0 18px 60px rgb(0 0 0 / 26%);
}

.support-button span {
  position: relative;
  width: 62px;
  height: 62px;
  border-radius: 50% 50% 45% 45%;
  background:
    radial-gradient(circle at 36% 39%, #111 0 4px, transparent 5px),
    radial-gradient(circle at 64% 39%, #111 0 4px, transparent 5px),
    linear-gradient(#ffd9c9, #ffd9c9);
}

.support-button span::before {
  position: absolute;
  inset: -8px 5px auto;
  height: 22px;
  border-radius: 18px 18px 6px 6px;
  background: #2d1f1d;
  content: '';
}

.support-button span::after {
  position: absolute;
  right: 8px;
  bottom: 12px;
  left: 8px;
  height: 5px;
  border-radius: 999px;
  background: #2b7e45;
  content: '';
}

@media (max-width: 1500px) {
  .hero-copy h1 {
    font-size: 82px;
  }

  .hero-copy p {
    font-size: 38px;
  }

  .hero-art {
    min-width: 540px;
  }

  .smart-section h2,
  .strategy-section h2,
  .trust-section h2,
  .futures-copy h2 {
    font-size: 54px;
  }
}
</style>
