<template>
  <section class="nav-page">
    <p>Stocks</p>
    <h1>股票</h1>
    <span>连接全球股票市场，提供清晰行情、低延迟报价与便捷交易入口。</span>
  </section>
</template>

<style scoped>
.nav-page {
  min-height: calc(100vh - 88px);
  padding: 120px 64px;
  background: #0b0d16;
}

p,
h1,
span {
  margin: 0;
}

p {
  color: var(--accent);
  font-size: 20px;
  font-weight: 900;
}

h1 {
  margin-top: 18px;
  font-size: 72px;
  line-height: 1.15;
}

span {
  display: block;
  max-width: 760px;
  margin-top: 28px;
  color: var(--text-soft);
  font-size: 28px;
  line-height: 1.5;
}
</style>
