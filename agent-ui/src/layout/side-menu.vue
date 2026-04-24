<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { Coin, CreditCard, DataBoard, House, User } from '@element-plus/icons-vue'

const props = defineProps<{
  collapsed: boolean
}>()

const route = useRoute()

const menuTree = computed(() => [
  { path: '/home', title: '工作台', icon: House },
  { path: '/categories', title: '分类管理', icon: DataBoard },
  { path: '/products', title: '产品管理', icon: Coin },
  { path: '/users', title: '用户管理', icon: User },
  { path: '/recharge-orders', title: '充值订单', icon: CreditCard },
  { path: '/withdraw-orders', title: '提现订单', icon: CreditCard },
])
</script>

<template>
  <el-menu
    class="aside-menu"
    router
    :default-active="route.path"
    :collapse="props.collapsed"
    :collapse-transition="false"
  >
    <el-menu-item v-for="item in menuTree" :key="item.path" :index="item.path">
      <el-icon>
        <component :is="item.icon" />
      </el-icon>
      <template #title>{{ item.title }}</template>
    </el-menu-item>
  </el-menu>
</template>

<style scoped>
.aside-menu {
  border-right: none;
  overflow-x: hidden;
}

:deep(.el-menu--collapse) {
  width: 100%;
}
</style>
