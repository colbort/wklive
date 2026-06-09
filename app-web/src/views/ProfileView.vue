<template>
  <section class="page-card profile-page">
    <h2>用户中心</h2>
    <div class="profile-grid">
      <RouterLink to="/login">
        登录 / 注册
      </RouterLink>
      <button type="button">
        实名认证
      </button>
      <button type="button">
        安全设置
      </button>
      <button type="button">
        收款账户
      </button>
    </div>
  </section>
</template>

<style scoped>
.profile-page {
  padding: 26px;
}

h2 {
  margin: 0 0 20px;
}

.profile-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
}

a,
button {
  display: grid;
  min-height: 120px;
  place-items: center;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  background: var(--surface-soft);
  color: var(--text);
  font-weight: 800;
}
</style>
