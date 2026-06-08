<script setup lang="ts">
import { ref } from 'vue'

const account = ref('')
const password = ref('')
</script>

<template>
  <main class="login-page">
    <section class="page-card login-card">
      <RouterLink to="/home" class="back">返回首页</RouterLink>
      <h1>登录</h1>
      <input v-model="account" placeholder="邮箱 / 手机" />
      <input v-model="password" placeholder="密码" type="password" />
      <button type="button">登录</button>
    </section>
  </main>
</template>

<style scoped>
.login-page {
  display: grid;
  min-height: 100vh;
  place-items: center;
  padding: 40px;
}

.login-card {
  display: grid;
  width: 420px;
  gap: 16px;
  padding: 30px;
}

.back {
  color: var(--text-muted);
}

h1 {
  margin: 0 0 12px;
}

input,
button {
  height: 48px;
  border: 0;
  border-radius: var(--radius-sm);
}

input {
  padding: 0 14px;
  background: var(--surface-soft);
  color: var(--text);
}

button {
  background: var(--accent);
  color: rgb(3 10 8);
  font-weight: 900;
}
</style>
