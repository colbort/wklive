<script setup lang="ts">
const markets = [
  { symbol: 'EUR/USD', price: '1.0862', change: '+0.31%' },
  { symbol: 'BTC/USDT', price: '84,562.11', change: '+3.82%' },
  { symbol: 'XAU/USD', price: '2,364.80', change: '+1.16%' },
  { symbol: 'WTI', price: '81.47', change: '-0.54%' },
]
</script>

<template>
  <section class="page-grid landing-subpage">
    <article class="panel-card panel-card--wide">
      <div class="panel-card__header">
        <h4>市场</h4>
        <span>外汇 · 加密货币 · 商品</span>
      </div>
      <p>统一查看不同市场分类和热门品种，保持接近交易应用的浏览体验。</p>
    </article>

    <article class="panel-card">
      <div class="panel-card__header">
        <h4>外汇</h4>
      </div>
      <p>主流与次主流货币对。</p>
    </article>

    <article class="panel-card">
      <div class="panel-card__header">
        <h4>加密货币</h4>
      </div>
      <p>数字资产合约与热门币对。</p>
    </article>

    <article class="panel-card">
      <div class="panel-card__header">
        <h4>商品</h4>
      </div>
      <p>黄金、原油等商品市场。</p>
    </article>

    <article class="panel-card panel-card--wide">
      <div class="panel-card__header">
        <h4>热门品种</h4>
        <span>市场行情</span>
      </div>
      <div class="market-list">
        <div v-for="item in markets" :key="item.symbol" class="market-item">
          <div>
            <strong>{{ item.symbol }}</strong>
            <p>实时价格</p>
          </div>
          <div class="market-item__right">
            <strong>{{ item.price }}</strong>
            <span :class="item.change.startsWith('+') ? 'up' : 'down'">{{ item.change }}</span>
          </div>
        </div>
      </div>
    </article>
  </section>
</template>
