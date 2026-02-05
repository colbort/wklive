<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { setLocale, type Locale } from '@/i18n'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'
import { Expand, Fold } from '@element-plus/icons-vue'

const props = defineProps<{
  collapsed: boolean
}>()

const emit = defineEmits<{
  (e: 'toggle-sider'): void
}>()

const { t, locale } = useI18n()
const auth = useAuthStore()
const router = useRouter()

const current = computed(() => locale.value as Locale)

function change(val: Locale) {
  setLocale(val)
}

function logout() {
  auth.logout()
  router.push('/login')
}
</script>

<template>
  <div class="topbar">
    <div class="left">
      <el-button text class="collapse-btn" @click="emit('toggle-sider')">
        <el-icon>
          <component :is="props.collapsed ? Expand : Fold" />
        </el-icon>
      </el-button>

      <span class="title">{{ t('app.title') }}</span>
    </div>

    <div class="right">
      <el-select style="width: 140px" :model-value="current" @update:model-value="change">
        <el-option label="中文" value="zh-CN" />
        <el-option label="English" value="en-US" />
      </el-select>

      <el-dropdown>
        <span class="user">
          {{ auth.user?.nickname || auth.user?.username || '-' }}
        </span>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item @click="logout">{{ t('app.logout') }}</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </div>
</template>

<style scoped>
.topbar { height: 56px; background: #fff; border-bottom: 1px solid #eee; display:flex; align-items:center; justify-content: space-between; padding: 0 16px; }
.left { display:flex; align-items:center; gap: 8px; min-width: 0; }
.title { font-weight: 700; }
.right { display:flex; gap: 12px; align-items:center; }
.user { cursor: pointer; }
.collapse-btn { padding: 6px; }
</style>
