<script setup lang="ts">
const rows = [
  ['BTC/USDT', '63,850.22', '+3.15%', '1.42B'],
  ['ETH/USDT', '1,688.90', '+2.08%', '823.4M'],
  ['CNY/USDT', '0.1380', '-0.21%', '64.2M'],
  ['XAU/USD', '2,331.42', '+0.42%', '112.8M'],
]
</script>

<template>
  <section class="page-card market-page">
    <header>
      <div>
        <h2>市场行情</h2>
        <p>外汇、加密货币、股票、期货、指数、基金</p>
      </div>
      <input type="search" placeholder="搜索产品">
    </header>

    <table>
      <thead>
        <tr>
          <th>产品</th>
          <th>最新价</th>
          <th>涨跌幅</th>
          <th>成交额</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="row in rows" :key="row[0]">
          <td>{{ row[0] }}</td>
          <td>{{ row[1] }}</td>
          <td :class="{ down: row[2].startsWith('-') }">
            {{ row[2] }}
          </td>
          <td>{{ row[3] }}</td>
        </tr>
      </tbody>
    </table>
  </section>
</template>

<style scoped>
.market-page {
  padding: 26px;
}

header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 22px;
}

h2,
p {
  margin: 0;
}

p {
  margin-top: 6px;
  color: var(--text-muted);
}

input {
  width: 320px;
  height: 42px;
  padding: 0 14px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--surface-soft);
  color: var(--text);
}

table {
  width: 100%;
  border-collapse: collapse;
}

th,
td {
  padding: 18px 12px;
  border-bottom: 1px solid var(--border);
  text-align: left;
}

th {
  color: var(--text-muted);
  font-size: 13px;
}

td:nth-child(2),
td:nth-child(3) {
  color: var(--accent);
  font-weight: 800;
}

.down {
  color: var(--danger) !important;
}
</style>
