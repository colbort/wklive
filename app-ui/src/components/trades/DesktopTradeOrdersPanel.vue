<template>
  <section class="desktop-bottom-panel">
    <header class="desktop-bottom-panel__tabs">
      <button type="button" class="active">当前委托</button>
      <button type="button">订单历史</button>
    </header>

    <div class="desktop-bottom-panel__table">
      <div class="desktop-bottom-panel__head">
        <span>合约名称</span>
        <span>盈亏</span>
        <span>价格</span>
        <span>持仓数量</span>
        <span>成交价/平仓价格</span>
        <span>止盈/止损</span>
        <span>订单价值</span>
        <span>详情</span>
      </div>

      <div class="desktop-bottom-panel__empty">
        <div class="desktop-bottom-panel__empty-icon">▤</div>
        <p>暂无数据</p>
      </div>
    </div>
  </section>
</template>

<style scoped>
.desktop-bottom-panel {
  display: grid;
  grid-template-rows: auto minmax(0, 1fr);
  min-height: 0;
  border-top: 1px solid #242633;
  background: #0b0c15;
}

.desktop-bottom-panel__tabs {
  display: flex;
  gap: 28px;
  padding: 14px 18px 0;
}

.desktop-bottom-panel__tabs button {
  position: relative;
  border: 0;
  background: transparent;
  color: #8f929d;
  font: inherit;
  font-size: 15px;
  font-weight: 500;
  padding-bottom: 16px;
}

.desktop-bottom-panel__tabs button.active {
  color: #fff;
}

.desktop-bottom-panel__tabs button.active::after {
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  height: 4px;
  border-radius: 999px;
  background: #fff;
  content: '';
}

.desktop-bottom-panel__table {
  display: grid;
  grid-template-rows: auto minmax(0, 1fr);
  min-height: 0;
}

.desktop-bottom-panel__head {
  display: grid;
  grid-template-columns: 1.3fr 0.55fr 0.5fr 0.75fr 1fr 0.8fr 0.8fr 0.45fr;
  gap: 18px;
  padding: 16px 18px;
  border-top: 1px solid #242633;
  color: #8f929d;
  font-size: 13px;
}

.desktop-bottom-panel__empty {
  display: grid;
  place-items: center;
  align-content: center;
  gap: 12px;
  color: #7d828e;
}

.desktop-bottom-panel__empty-icon {
  width: 72px;
  height: 72px;
  display: grid;
  place-items: center;
  border-radius: 999px;
  background: radial-gradient(circle at 50% 35%, rgba(255, 255, 255, 0.16), rgba(255, 255, 255, 0.02));
  color: #cfd3dc;
  font-size: 28px;
}
</style>
