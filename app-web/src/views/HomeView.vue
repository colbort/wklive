<script setup lang="ts">
const metrics = [
  { label: '今日成交额', value: '$128.42M', trend: '+12.8%' },
  { label: '活跃市场', value: '238', trend: '+18' },
  { label: '在线用户', value: '42,905', trend: '+6.4%' },
  { label: '系统延迟', value: '18ms', trend: '稳定' },
]

const quickActions = ['现货交易', '合约交易', '资金划转', '充值提现']
</script>

<template>
  <section class="home-grid">
    <div class="hero page-card">
      <div>
        <p>从股票到加密，一个平台全搞定</p>
        <h1>专业、稳定的 PC 综合交易台</h1>
        <span>实时行情、深度订单簿、多资产账户管理，面向桌面端高频操作。</span>
      </div>
      <RouterLink to="/trades">开始交易</RouterLink>
    </div>

    <div class="metrics">
      <article v-for="item in metrics" :key="item.label" class="metric page-card">
        <span>{{ item.label }}</span>
        <strong>{{ item.value }}</strong>
        <em>{{ item.trend }}</em>
      </article>
    </div>

    <div class="page-card panel">
      <header>
        <strong>快捷入口</strong>
        <span class="muted">常用操作</span>
      </header>
      <div class="actions">
        <button v-for="action in quickActions" :key="action" type="button">{{ action }}</button>
      </div>
    </div>

    <div class="page-card panel">
      <header>
        <strong>行情提醒</strong>
        <span class="muted">WebSocket 实时更新</span>
      </header>
      <div class="placeholder">等待接入 PC 行情组件</div>
    </div>
  </section>
</template>

<style scoped>
.home-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 18px;
}

.hero {
  display: flex;
  grid-column: 1 / -1;
  align-items: flex-end;
  justify-content: space-between;
  min-height: 280px;
  padding: 34px;
  overflow: hidden;
  background:
    radial-gradient(circle at 78% 24%, rgb(6 203 30 / 26%), transparent 28%),
    linear-gradient(135deg, rgb(22 28 42), rgb(11 14 24));
}

.hero p {
  margin: 0 0 12px;
  color: var(--accent);
  font-weight: 800;
}

.hero h1 {
  max-width: 680px;
  margin: 0 0 16px;
  font-size: 48px;
  line-height: 1.08;
}

.hero span {
  color: var(--text-soft);
}

.hero a {
  flex: 0 0 auto;
  padding: 14px 26px;
  border-radius: 999px;
  background: var(--accent);
  color: rgb(3 10 8);
  font-weight: 900;
}

.metrics {
  display: grid;
  grid-column: 1 / -1;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 18px;
}

.metric {
  display: grid;
  gap: 10px;
  padding: 22px;
}

.metric span,
.metric em {
  color: var(--text-muted);
  font-style: normal;
}

.metric strong {
  font-size: 30px;
}

.panel {
  grid-column: span 2;
  padding: 22px;
}

.panel header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 18px;
}

.actions {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.actions button {
  min-height: 64px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--surface-soft);
  color: var(--text);
  font-weight: 700;
}

.placeholder {
  display: grid;
  min-height: 110px;
  place-items: center;
  color: var(--text-muted);
}
</style>
