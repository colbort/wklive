<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { setLocale, type Locale } from '@/i18n'
import { useAuthStore } from '@/stores'
import { useRouter } from 'vue-router'
import { Expand, Fold, User, Setting, Lock } from '@element-plus/icons-vue'
import { ElMessageBox } from 'element-plus'

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

function changePassword() {
  // TODO: 实现修改密码逻辑，可以打开一个对话框
  ElMessageBox.prompt(t('app.newPasswordPrompt'), t('app.changePassword'), {
    confirmButtonText: t('common.confirm'),
    cancelButtonText: t('common.cancel'),
    inputPattern: /^.{6,}$/,
    inputErrorMessage: t('app.passwordMinLength'),
  }).then(({ value }) => {
    // 调用API修改密码
    console.log('新密码:', value)
    // auth.changePassword(value)
  }).catch(() => {
    console.log('取消修改密码')
  })
}

function openSettings() {
  // TODO: 实现设置逻辑，可以打开一个对话框修改昵称等
  ElMessageBox.prompt(t('app.newNicknamePrompt'), t('app.settings'), {
    confirmButtonText: t('common.confirm'),
    cancelButtonText: t('common.cancel'),
    inputValue: auth.user?.nickname || '',
  }).then(({ value }) => {
    // 调用API修改昵称
    console.log('新昵称:', value)
    // auth.updateProfile({ nickname: value })
  }).catch(() => {
    console.log('取消设置')
  })
}

// Avatar upload handling
function onAvatarClick() {
  // trigger hidden file input
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = 'image/*'
  input.onchange = async (e: Event) => {
    const file = (e.target as HTMLInputElement).files?.[0]
    if (!file) return
    // TODO: upload file to server and get URL
    const reader = new FileReader()
    reader.onload = () => {
      const url = reader.result as string
      console.log('selected avatar base64', url)
      // directly mutate store user so avatar updates immediately
      if (auth.user) {
        auth.user.avatar = url
      }
      // TODO: upload file to server and replace with real URL if necessary
    }
    reader.readAsDataURL(file)
  }
  input.click()
}

function handleCommand(command: string) {
  switch (command) {
    case 'changePassword':
      changePassword()
      break
    case 'settings':
      openSettings()
      break
    case 'logout':
      logout()
      break
  }
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

      <el-dropdown trigger="contextmenu" @command="handleCommand">
        <div class="avatar-container" :title="t('app.uploadAvatar')" @click.stop.prevent="onAvatarClick">
          <el-avatar :size="32" :src="auth.user?.avatar" :alt="auth.user?.nickname || auth.user?.username">
            <el-icon><User /></el-icon>
          </el-avatar>
        </div>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="changePassword">
              <el-icon><Lock /></el-icon>
              {{ t('app.changePassword') }}
            </el-dropdown-item>
            <el-dropdown-item command="settings">
              <el-icon><Setting /></el-icon>
              {{ t('app.settings') }}
            </el-dropdown-item>
            <el-dropdown-item command="logout" divided>
              {{ t('app.logout') }}
            </el-dropdown-item>
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
.avatar-container { cursor: pointer; }
.collapse-btn { padding: 6px; }
</style>
