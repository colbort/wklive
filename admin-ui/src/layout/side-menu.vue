<script setup lang="ts">
import { computed } from 'vue'
import { useAuthStore, type MenuNode } from '@/stores/auth'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

const auth = useAuthStore()
const router = useRouter()
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
</script>

<template>
  <el-menu router :default-active="$route.path">
    <!-- ✅ 首页固定显示 -->
    <el-menu-item index="/" @click="go('/')">
      <span>{{ t('route.home') }}</span>
    </el-menu-item>

    <template v-for="m in menuTree" :key="m.id">
      <el-sub-menu v-if="m.menuType === 1" :index="String(m.id)">
        <template #title>
          <span>{{ labelById(m.id, m.name) }}</span>
        </template>

        <el-menu-item
          v-for="c in childrenMenus(m)"
          :key="c.id"
          :index="c.path"
          @click="go(c.path)"
        >
          <span>{{ labelById(c.id, c.name) }}</span>
        </el-menu-item>
      </el-sub-menu>

      <el-menu-item
        v-else-if="m.menuType === 2"
        :index="m.path"
        @click="go(m.path)"
      >
        <span>{{ labelById(m.id, m.name) }}</span>
      </el-menu-item>
    </template>
  </el-menu>
</template>
