<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { Expand, Fold, OfficeBuilding, SwitchButton, User } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { useTenantStore } from '@/stores/tenant'

const props = defineProps<{
  collapsed: boolean
}>()

const emit = defineEmits<{
  (e: 'toggle-sider'): void
}>()

const auth = useAuthStore()
const tenant = useTenantStore()
const router = useRouter()

const displayName = computed(() => auth.user?.nickname || auth.user?.username || 'Agent')

function logout() {
  auth.logout()
  router.replace('/login')
}
</script>

<template>
  <header class="topbar">
    <div class="topbar__left">
      <el-button text @click="emit('toggle-sider')">
        <el-icon>
          <component :is="props.collapsed ? Expand : Fold" />
        </el-icon>
      </el-button>
      <div class="tenant-pill">
        <el-icon><OfficeBuilding /></el-icon>
        <span>{{ tenant.tenantName || tenant.tenantCode || 'Tenant' }}</span>
      </div>
    </div>

    <div class="topbar__right">
      <div class="user-meta">
        <el-icon><User /></el-icon>
        <span>{{ displayName }}</span>
      </div>
      <el-button text @click="logout">
        <el-icon><SwitchButton /></el-icon>
        <span>退出</span>
      </el-button>
    </div>
  </header>
</template>

<style scoped>
.topbar {
  height: 56px;
  border-bottom: 1px solid #eee;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  background: #fff;
  gap: 12px;
}

.topbar__left,
.topbar__right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.tenant-pill,
.user-meta {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 0 10px;
  height: 34px;
  border-radius: 17px;
  background: #f4f6f8;
  color: #303133;
}

span {
  line-height: 1;
}
</style>
