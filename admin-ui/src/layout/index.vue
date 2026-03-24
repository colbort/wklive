<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import SideMenu from './side-menu.vue'
import TopBar from './top-bar.vue'
import { getSystemCore } from '@/stores/core'
import { http } from '@/utils/request'

type SystemCoreModel = {
  siteName: string
  siteLogo: string
}

const collapsed = ref(false)
const systemCore = ref<SystemCoreModel>({
  siteName: 'Admin UI',
  siteLogo: '',
})

onMounted(async () => {
  try {
    const res = await getSystemCore()
    if (res && res.code === 200 && res.data) {
      if (res.data.siteName) systemCore.value.siteName = res.data.siteName
      if (res.data.siteLogo) systemCore.value.siteLogo = res.data.siteLogo
    }
  } catch (err) {
    console.error('getSystemCore failed', err)
  }
})

// 展开态宽度
const asideWidth = ref(240)
const MIN_W = 200
const MAX_W = 360
const COLLAPSED_W = 64

const realAsideWidth = computed(() => (collapsed.value ? COLLAPSED_W : asideWidth.value))

let dragging = false
let startX = 0
let startW = 0

function toggleSider() {
  collapsed.value = !collapsed.value
}

function onDragStart(e: MouseEvent) {
  if (collapsed.value) return
  dragging = true
  startX = e.clientX
  startW = asideWidth.value
  document.body.style.userSelect = 'none'
  document.addEventListener('mousemove', onDragMove)
  document.addEventListener('mouseup', onDragEnd)
}

function onDragMove(e: MouseEvent) {
  if (!dragging) return
  const dx = e.clientX - startX
  const next = Math.min(MAX_W, Math.max(MIN_W, startW + dx))
  asideWidth.value = next
}

function onDragEnd() {
  dragging = false
  document.body.style.userSelect = ''
  document.removeEventListener('mousemove', onDragMove)
  document.removeEventListener('mouseup', onDragEnd)
}

onBeforeUnmount(() => {
  document.removeEventListener('mousemove', onDragMove)
  document.removeEventListener('mouseup', onDragEnd)
})

function formatUrl(url: string | undefined) {
  console.log('formatUrl', url)
  if (!url) return ''
  const fullUrl = url.startsWith('http') ? url : `${http.defaults.baseURL}${url}`
  return `${fullUrl}${fullUrl.includes('?') ? '&' : '?'}t=${Date.now()}`
}
</script>

<template>
  <div class="layout">
    <aside class="sider" :style="{ width: realAsideWidth + 'px' }">
      <div class="brand">
        <img
          v-if="systemCore.siteLogo"
          :src="formatUrl(systemCore.siteLogo)"
          alt="logo"
          class="brand-logo"
        />
        <span class="brand-text" v-if="!collapsed">{{ systemCore.siteName }}</span>
        <span class="brand-text" v-else>{{
          systemCore.siteName ? systemCore.siteName.slice(0, 2).toUpperCase() : 'AI'
        }}</span>
      </div>

      <!-- 关键：把 collapsed 传给侧边菜单，让 el-menu collapse -->
      <SideMenu :collapsed="collapsed" />

      <!-- 拖拽条：仅展开态显示 -->
      <div v-if="!collapsed" class="resizer" @mousedown="onDragStart" />
    </aside>

    <main class="main">
      <!-- TopBar 放折叠按钮 -->
      <TopBar :collapsed="collapsed" @toggle-sider="toggleSider" />

      <div class="content">
        <router-view />
      </div>
    </main>
  </div>
</template>

<style scoped>
.layout {
  display: flex;
  height: 100vh;
}

/* ✅ 禁止左右滑动的关键：overflow-x hidden */
.sider {
  border-right: 1px solid #eee;
  display: flex;
  flex-direction: column;
  background: #fff;
  overflow-x: hidden;
  overflow-y: auto;
  position: relative;
}

/* brand 不要撑出横向滚动 */
.brand {
  height: 56px;
  display: flex;
  align-items: center;
  padding: 0 16px;
  font-weight: 700;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.brand-logo {
  height: 32px;
  width: 32px;
  margin-right: 8px;
  border-radius: 4px;
  object-fit: contain;
}

.main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}
.content {
  padding: 16px;
  overflow: auto;
  flex: 1;
  background: #f7f8fa;
  min-width: 0;
}

/* ✅ 拖拽条 */
.resizer {
  position: absolute;
  top: 0;
  right: 0;
  width: 6px;
  height: 100%;
  cursor: col-resize;
}
.resizer:hover {
  background: rgba(0, 0, 0, 0.06);
}
</style>
