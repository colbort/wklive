<script setup lang="ts">
import { computed } from 'vue'
import { useAuthStore, type MenuNode } from '@/stores/auth'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'

import {
  House,
  Setting,
  Menu as MenuIcon,
  User,
  List,
  Operation,
  Tickets,
} from '@element-plus/icons-vue'

const props = defineProps<{
  collapsed: boolean
}>()

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()
const { t, te } = useI18n()

const menuTree = computed(() => (auth.menus || []).slice().sort((a, b) => (a.sort ?? 0) - (b.sort ?? 0)))

function labelById(id: number, fallback: string) {
  const key = `menu.${id}`
  return te(key) ? t(key) : fallback
}

function go(path?: string) {
  if (path) router.push(path)
}

function childrenMenus(n: MenuNode) {
  return (n.children || [])
    .filter((x) => x.menuType === 2 && x.visible !== 0 && x.status !== 0)
    .slice()
    .sort((a, b) => (a.sort ?? 0) - (b.sort ?? 0))
}

// ✅ icon 映射：后端 sys_menu.icon -> ElementPlus Icon
function iconComp(icon?: string) {
  switch (icon) {
    case 'Setting':
      return Setting
    case 'User':
      return User
    case 'List':
      return List
    case 'Operation':
      return Operation
    case 'Tickets':
      return Tickets
    default:
      return MenuIcon
  }
}
</script>

<template>
  <el-menu
    class="aside-menu"
    router
    :default-active="route.path"
    :collapse="props.collapsed"
    :collapse-transition="false"
  >
    <!-- ✅ 首页固定显示 -->
    <el-menu-item index="/" @click="go('/')">
      <el-icon><House /></el-icon>
      <template #title>{{ t('route.home') }}</template>
    </el-menu-item>

    <template v-for="m in menuTree" :key="m.id">
      <el-sub-menu v-if="m.menuType === 1" :index="String(m.id)">
        <template #title>
          <el-icon><component :is="iconComp(m.icon)" /></el-icon>
          <span>{{ labelById(m.id, m.name) }}</span>
        </template>

        <el-menu-item
          v-for="c in childrenMenus(m)"
          :key="c.id"
          :index="c.path"
          @click="go(c.path)"
        >
          <el-icon><component :is="iconComp(c.icon)" /></el-icon>
          <template #title>{{ labelById(c.id, c.name) }}</template>
        </el-menu-item>
      </el-sub-menu>

      <el-menu-item
        v-else-if="m.menuType === 2"
        :index="m.path"
        @click="go(m.path)"
      >
        <el-icon><component :is="iconComp(m.icon)" /></el-icon>
        <template #title>{{ labelById(m.id, m.name) }}</template>
      </el-menu-item>
    </template>
  </el-menu>
</template>

<style scoped>
/* ✅ 不要横向滚动 */
.aside-menu {
  border-right: none;
  overflow-x: hidden;
}

/* ✅ 修复折叠后图标居中更舒服 */
:deep(.el-menu--collapse) {
  width: 100%;
}
</style>
