<script setup lang="ts">
/**
 * 租户工作台：展示当前固定 tenantCode 对应的租户信息和快捷入口。
 */
import { computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useTenantStore } from '@/stores/tenant'

const auth = useAuthStore()
const tenant = useTenantStore()

const shortcuts = computed(() => [
  { title: '分类管理', desc: '管理当前租户可展示的分类', path: '/categories' },
  { title: '产品管理', desc: '管理当前租户可展示的产品', path: '/products' },
  { title: '用户管理', desc: '查看和维护当前租户用户', path: '/users' },
  { title: '充值订单', desc: '处理充值订单和回调重试', path: '/recharge-orders' },
  { title: '提现订单', desc: '审核提现订单', path: '/withdraw-orders' },
])

onMounted(() => {
  tenant.ensureLoaded()
})
</script>

<template>
  <div class="dashboard">
    <el-card shadow="never">
      <template #header>租户工作台</template>
      <div class="summary-grid">
        <div class="summary-item">
          <div class="summary-label">当前租户</div>
          <div class="summary-value">{{ tenant.tenantName || '-' }}</div>
        </div>
        <div class="summary-item">
          <div class="summary-label">tenantCode</div>
          <div class="summary-value">{{ tenant.tenantCode || '-' }}</div>
        </div>
        <div class="summary-item">
          <div class="summary-label">租户ID</div>
          <div class="summary-value">{{ tenant.tenantId || '-' }}</div>
        </div>
        <div class="summary-item">
          <div class="summary-label">当前登录</div>
          <div class="summary-value">{{ auth.user?.nickname || auth.user?.username || '-' }}</div>
        </div>
      </div>
    </el-card>

    <div class="shortcut-grid">
      <el-card v-for="item in shortcuts" :key="item.path" shadow="hover" class="shortcut-card">
        <div class="shortcut-title">{{ item.title }}</div>
        <div class="shortcut-desc">{{ item.desc }}</div>
        <router-link class="shortcut-link" :to="item.path">进入</router-link>
      </el-card>
    </div>
  </div>
</template>

<style scoped>
.dashboard {
  display: grid;
  gap: 16px;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.summary-item {
  padding: 16px;
  border-radius: 12px;
  background: #f7f8fa;
}

.summary-label {
  color: #909399;
  font-size: 13px;
  margin-bottom: 8px;
}

.summary-value {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
}

.shortcut-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
}

.shortcut-card {
  min-height: 150px;
}

.shortcut-title {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 10px;
}

.shortcut-desc {
  color: #606266;
  line-height: 1.7;
  margin-bottom: 16px;
}

.shortcut-link {
  color: #409eff;
  text-decoration: none;
}
</style>
