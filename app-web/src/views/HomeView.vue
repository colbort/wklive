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
        <div>
          <div class="footer-logo">
            <span class="mini-mark" />
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
  width: 100%;
  min-width: 100%;
  overflow: hidden;
  background:
    radial-gradient(circle at 9% 62%, rgb(77 105 110 / 16%), transparent 24%),
    linear-gradient(180deg, #0b0d16 0%, #0a0c15 100%);
  color: var(--text);
}

.section-pad {
  position: relative;
  padding: var(--px-96) var(--px-64);
}

.grid-bg {
  background-image:
    linear-gradient(rgb(255 255 255 / 5%) 1px, transparent 1px),
    linear-gradient(90deg, rgb(255 255 255 / 5%) 1px, transparent 1px);
  background-position: -2px -2px;
  background-size: var(--px-205) var(--px-205);
}

.hero-section {
  position: relative;
  display: flex;
  width: 100%;
  min-height: var(--px-930);
  align-items: center;
  padding: var(--px-92) var(--px-64) var(--px-120);
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
  padding-top: var(--px-30);
}

h1,
h2,
h3,
p {
  margin: 0;
}

.hero-copy h1 {
  max-width: var(--px-1030);
  font-size: var(--font-size-102);
  font-weight: var(--font-weight-950);
  line-height: 1.18;
}

.hero-copy h1 span,
h2 span {
  color: var(--accent);
}

.hero-copy p {
  margin-top: var(--px-48);
  font-size: var(--font-size-46);
  line-height: 1.3;
}

.primary-cta {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: var(--px-290);
  height: var(--px-104);
  margin-top: var(--px-110);
  border-radius: var(--px-999);
  background: var(--accent);
  color: var(--white);
  font-size: var(--font-size-40);
  font-weight: var(--font-weight-900);
}

.hero-art {
  position: absolute;
  z-index: 1;
  right: calc(var(--px-18) * -1);
  top: var(--px-155);
  width: min(43vw, var(--px-750));
  min-width: var(--px-610);
}

.smart-section {
  min-height: var(--px-850);
  padding-top: var(--px-44);
}

.smart-section h2,
.strategy-section h2,
.trust-section h2 {
  text-align: center;
  font-size: var(--font-size-64);
  font-weight: var(--font-weight-950);
  line-height: 1.18;
}

.smart-cards {
  position: relative;
  display: grid;
  max-width: var(--px-1900);
  margin: var(--px-370) auto 0;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  align-items: end;
  border: 1px solid rgb(255 255 255 / 24%);
  border-radius: var(--px-24);
  background: linear-gradient(180deg, rgb(255 255 255 / 7%), rgb(255 255 255 / 2%));
}

.smart-card {
  min-height: var(--px-390);
  padding: var(--px-64) var(--px-56) var(--px-48);
}

.smart-card img {
  width: var(--px-112);
  height: var(--px-112);
  object-fit: contain;
}

.smart-card h3 {
  margin-top: var(--px-78);
  font-size: var(--font-size-40);
  line-height: 1.2;
}

.smart-card p {
  margin-top: var(--px-34);
  color: var(--text-soft);
  font-size: var(--font-size-26);
  font-weight: var(--font-weight-800);
  line-height: 1.45;
}

.smart-card--featured {
  min-height: var(--px-675);
  margin-top: calc(var(--px-286) * -1);
  border-radius: var(--px-24) var(--px-24) 0 0;
  background: var(--accent);
}

.smart-card--featured img {
  width: var(--px-132);
  height: var(--px-132);
  filter: brightness(0) invert(1);
}

.smart-card--featured h3,
.smart-card--featured p {
  color: var(--white);
}

.strategy-section {
  min-height: var(--px-960);
  padding-top: var(--px-164);
}

.strategy-list {
  display: grid;
  max-width: var(--px-1480);
  margin: var(--px-120) auto 0;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--px-160);
}

.strategy-item {
  display: flex;
  align-items: center;
  gap: var(--px-28);
}

.strategy-item img {
  width: var(--px-110);
  height: var(--px-110);
}

.strategy-item h3 {
  font-size: var(--font-size-40);
  line-height: 1.2;
}

.strategy-item p {
  margin-top: var(--px-28);
  color: var(--text-soft);
  font-size: var(--font-size-26);
  font-weight: var(--font-weight-800);
}

.tools-art {
  display: block;
  width: var(--px-730);
  margin: var(--px-130) auto calc(var(--px-260) * -1);
  opacity: 0.86;
}

.market-flow {
  min-height: var(--px-900);
}

.market-flow::before,
.futures-section .orbit-line {
  position: absolute;
  width: var(--px-1650);
  height: var(--px-570);
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
  left: var(--px-150);
  top: var(--px-90);
  width: var(--px-500);
}

.poster-card {
  right: var(--px-290);
  top: var(--px-170);
  width: var(--px-540);
}

.market-flow h2 {
  position: absolute;
  right: 124px;
  bottom: 60px;
  font-size: var(--font-size-58);
  font-weight: var(--font-weight-950);
  line-height: 1.35;
}

.futures-section {
  min-height: var(--px-930);
}

.futures-section .orbit-line {
  top: 48px;
  left: -370px;
  height: 690px;
  transform: rotate(-18deg);
}

.futures-copy {
  width: var(--px-780);
  margin-left: auto;
  margin-right: var(--px-300);
}

.futures-copy h2 {
  font-size: var(--font-size-64);
  font-weight: var(--font-weight-950);
  line-height: 1.34;
}

.futures-copy ul {
  display: grid;
  gap: var(--px-52);
  margin: var(--px-86) 0 0;
  padding: 0;
  list-style: none;
}

.futures-copy li {
  display: flex;
  gap: 16px;
}

.futures-copy li > span {
  flex: 0 0 auto;
  width: var(--px-32);
  height: var(--px-20);
  margin-top: 8px;
  border-radius: var(--px-999);
  background: linear-gradient(90deg, var(--accent) 0 48%, var(--white) 48% 100%);
}

.futures-copy strong {
  display: block;
  font-size: var(--font-size-28);
  line-height: 1.2;
}

.futures-copy p {
  margin-top: 16px;
  color: var(--text-soft);
  font-size: var(--font-size-22);
  font-weight: var(--font-weight-800);
}

.trust-section {
  min-height: var(--px-900);
  padding-top: var(--px-180);
}

.trust-section > p {
  margin-top: 26px;
  text-align: center;
  color: var(--text-soft);
  font-size: var(--font-size-28);
  font-weight: var(--font-weight-800);
}

.trust-cards {
  display: grid;
  max-width: var(--px-1920);
  margin: var(--px-130) auto 0;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--px-38);
}

.trust-card {
  min-height: var(--px-420);
  padding: var(--px-76) var(--px-52) var(--px-44);
  border: 1px solid rgb(255 255 255 / 20%);
  border-radius: var(--px-16);
  background: linear-gradient(180deg, rgb(255 255 255 / 7%), rgb(255 255 255 / 2%));
  text-align: center;
}

.trust-card img {
  width: var(--px-122);
  height: var(--px-122);
  object-fit: contain;
}

.trust-card h3 {
  margin-top: 62px;
  font-size: var(--font-size-38);
  line-height: 1.2;
}

.trust-card p {
  margin-top: 38px;
  color: var(--text-soft);
  font-size: var(--font-size-24);
  font-weight: var(--font-weight-800);
  line-height: 1.42;
}

.reward-section {
  display: flex;
  min-height: var(--px-760);
  align-items: center;
  justify-content: space-between;
  gap: var(--px-60);
  padding-top: var(--px-80);
}

.reward-copy h2 {
  font-size: var(--font-size-78);
  font-weight: var(--font-weight-950);
  line-height: 1.35;
}

.reward-copy p {
  margin-top: var(--px-74);
  font-size: var(--font-size-34);
  font-weight: var(--font-weight-800);
}

.reward-section img {
  width: min(48vw, var(--px-745));
}

.site-footer {
  position: relative;
  padding: 0 var(--px-64) 0;
  background: #0a0c15;
}

.footer-line {
  height: var(--px-6);
  border-radius: var(--px-999);
  background: linear-gradient(90deg, #c6c9df, #6075a8 88%, #c7c9da);
}

.footer-inner {
  display: grid;
  max-width: var(--px-1900);
  margin: 0 auto;
  padding: var(--px-64) 0 var(--px-48);
  grid-template-columns: var(--px-560) 1fr;
  gap: var(--px-200);
}

.footer-logo {
  display: flex;
  align-items: center;
  gap: 14px;
  font-size: var(--font-size-46);
  font-weight: var(--font-weight-950);
}

.mini-mark {
  position: relative;
  display: grid;
  width: var(--px-64);
  height: var(--px-64);
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
  font-size: var(--font-size-18);
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
  width: var(--px-456);
  margin-top: 68px;
  padding: var(--px-38) var(--px-36);
  border-radius: var(--px-24);
  background: rgb(29 32 44);
}

.download-card h3 {
  font-size: var(--font-size-28);
}

.download-card p {
  margin-top: 28px;
  color: var(--text-soft);
  font-size: var(--font-size-22);
  font-weight: var(--font-weight-800);
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
  width: var(--px-184);
  height: var(--px-184);
  border: var(--px-16) solid var(--white);
  border-radius: 18px;
  background:
    linear-gradient(90deg, #000 12px, transparent 12px 22px, #000 22px 34px, transparent 34px) 0 0 /
      46px 46px,
    linear-gradient(#000 12px, transparent 12px 22px, #000 22px 34px, transparent 34px) 0 0 / 46px
      46px,
    var(--white);
}

.qr-row em {
  color: var(--text);
  font-size: var(--font-size-22);
  font-style: normal;
}

.footer-links {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--px-86);
  padding-top: 96px;
}

.footer-links div {
  display: grid;
  align-content: start;
  gap: 32px;
}

.footer-links h3 {
  margin-bottom: 14px;
  font-size: var(--font-size-30);
}

.footer-links a {
  display: inline-flex;
  align-items: center;
  gap: 14px;
  color: var(--text);
  font-size: var(--font-size-24);
  font-weight: var(--font-weight-800);
}

.footer-links a span {
  display: grid;
  width: var(--px-30);
  height: var(--px-30);
  place-items: center;
  border-radius: 50%;
  background: #0b8fea;
  font-size: var(--font-size-15);
  font-weight: var(--font-weight-950);
}

.copyright {
  margin: 0 calc(var(--px-64) * -1);
  padding: 28px var(--px-64);
  border-top: 1px solid rgb(255 255 255 / 7%);
  color: rgb(255 255 255 / 45%);
  text-align: center;
  font-size: var(--font-size-18);
  font-weight: var(--font-weight-800);
}

.support-button {
  position: fixed;
  z-index: 30;
  right: var(--px-34);
  bottom: var(--px-34);
  display: grid;
  width: var(--px-130);
  height: var(--px-130);
  place-items: center;
  border: 0;
  border-radius: var(--px-28);
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
  .section-pad {
    padding-right: var(--px-48);
    padding-left: var(--px-48);
  }

  .hero-section {
    padding-right: var(--px-48);
    padding-left: var(--px-48);
  }

  .hero-copy h1 {
    font-size: var(--font-size-82);
  }

  .hero-copy p {
    font-size: var(--font-size-38);
  }

  .primary-cta {
    min-width: var(--px-240);
    height: var(--px-86);
    font-size: var(--font-size-32);
  }

  .hero-art {
    min-width: var(--px-540);
  }

  .smart-section h2,
  .strategy-section h2,
  .trust-section h2,
  .futures-copy h2 {
    font-size: var(--font-size-54);
  }

  .smart-card h3,
  .strategy-item h3 {
    font-size: var(--font-size-34);
  }

  .smart-card p,
  .strategy-item p {
    font-size: var(--font-size-22);
  }

  .strategy-list {
    gap: var(--px-96);
  }

  .tools-art {
    width: var(--px-620);
  }

  .market-flow h2 {
    font-size: var(--font-size-50);
  }

  .futures-copy {
    margin-right: var(--px-160);
  }

  .reward-copy h2 {
    font-size: var(--font-size-64);
  }

  .footer-inner {
    grid-template-columns: var(--px-480) 1fr;
    gap: var(--px-120);
  }

  .download-card {
    width: var(--px-420);
  }
}

@media (max-width: 1280px) {
  .section-pad {
    padding-right: var(--px-36);
    padding-left: var(--px-36);
  }

  .hero-section {
    padding-right: var(--px-36);
    padding-left: var(--px-36);
  }

  .hero-copy h1 {
    font-size: var(--font-size-70);
  }

  .hero-copy p {
    font-size: var(--font-size-32);
  }

  .hero-art {
    min-width: var(--px-480);
  }

  .smart-section h2,
  .strategy-section h2,
  .trust-section h2,
  .futures-copy h2 {
    font-size: var(--font-size-46);
  }

  .smart-cards {
    margin-top: var(--px-300);
  }

  .smart-card {
    padding: var(--px-48) var(--px-36) var(--px-40);
  }

  .smart-card--featured {
    min-height: var(--px-600);
    margin-top: calc(var(--px-240) * -1);
  }

  .strategy-list {
    gap: var(--px-64);
  }

  .futures-copy {
    margin-right: var(--px-72);
  }

  .footer-inner {
    grid-template-columns: var(--px-420) 1fr;
    gap: var(--px-80);
  }

  .support-button {
    right: var(--px-24);
    bottom: var(--px-24);
    width: var(--px-112);
    height: var(--px-112);
  }
}
</style>
