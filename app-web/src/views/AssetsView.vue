<template>
  <section class="page-card assets-page">
    <header>
      <h2>资产总览</h2>
      <div>
        <button type="button">
          充值
        </button>
        <button type="button">
          提现
        </button>
        <button type="button">
          划转
        </button>
      </div>
    </header>
    <div class="empty">
      登录后查看资产数据
    </div>
  </section>
</template>

<style scoped>
.assets-page {
  min-height: 640px;
  padding: 26px;
}

header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

h2 {
  margin: 0;
}

header div {
  display: flex;
  gap: 10px;
}

button {
  height: 40px;
  padding: 0 18px;
  border: 0;
  border-radius: 999px;
  background: var(--surface-soft);
  color: var(--text);
  font-weight: 700;
}

button:first-child {
  background: var(--accent);
  color: rgb(3 10 8);
}

.empty {
  display: grid;
  min-height: 520px;
  place-items: center;
  color: var(--text-muted);
}
</style>
