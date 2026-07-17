<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { apiExchangeGuestTransfer } from '@wklive/api/api/userPublic'

const transferStatus = ref('')

onMounted(async () => {
  const code = new URLSearchParams(window.location.hash.slice(1)).get('code') || ''
  if (!code) return

  window.history.replaceState(null, '', window.location.pathname + window.location.search)
  transferStatus.value = '正在恢复游客身份…'
  try {
    await apiExchangeGuestTransfer({ code })
    transferStatus.value = '游客身份已恢复'
  } catch {
    transferStatus.value = '游客身份恢复失败，请重新获取迁移链接'
  }
})
</script>

<template>
  <section class="page-card profile-page">
    <h2>用户中心</h2>
    <p v-if="transferStatus">{{ transferStatus }}</p>
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
  border-radius: var(--px-12);
  background: var(--surface-soft);
  color: var(--text);
  font-weight: 800;
}
</style>
