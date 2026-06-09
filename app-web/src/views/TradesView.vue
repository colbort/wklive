<template>
  <section class="trade-layout">
    <div class="page-card chart-panel">
      <header>
        <strong>BTCUSDT</strong>
        <span class="muted">63,850.7200 +3.15%</span>
      </header>
      <div class="chart-placeholder">
        PC K 线区域
      </div>
    </div>

    <aside class="page-card order-panel">
      <h2>下单</h2>
      <div class="switch">
        <button type="button">
          限价
        </button>
        <button type="button" class="active">
          市价
        </button>
      </div>
      <input placeholder="数量(BTC)">
      <button type="button" class="buy">
        买入开多
      </button>
      <button type="button" class="sell">
        卖出开空
      </button>
    </aside>
  </section>
</template>

<style scoped>
.trade-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 360px;
  gap: 18px;
}

.chart-panel,
.order-panel {
  padding: 22px;
}

.chart-panel header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 16px;
  font-size: 22px;
}

.chart-placeholder {
  display: grid;
  min-height: 620px;
  place-items: center;
  border: 1px dashed var(--border-strong);
  border-radius: var(--px-12);
  color: var(--text-muted);
}

.order-panel {
  display: grid;
  align-content: start;
  gap: 16px;
}

.order-panel h2 {
  margin: 0;
}

.switch {
  display: grid;
  grid-template-columns: 1fr 1fr;
  padding: 4px;
  border-radius: 999px;
  background: var(--surface-soft);
}

button,
input {
  height: 48px;
  border: 0;
  border-radius: var(--px-8);
}

button {
  background: transparent;
  color: var(--text-muted);
  font-weight: 800;
}

button.active,
.buy {
  background: var(--accent);
  color: rgb(3 10 8);
}

input {
  padding: 0 14px;
  background: var(--surface-soft);
  color: var(--text);
}

.sell {
  background: var(--danger);
  color: var(--text);
}
</style>
