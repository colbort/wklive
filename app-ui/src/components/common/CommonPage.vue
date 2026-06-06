<template>
  <div class="page-414">
    <!-- 有 custom：tabbar + custom 一起滚动 -->
    <template v-if="hasCustom">
      <div class="scroll-view">
        <header class="header-bar">
          <button
            v-if="showBack"
            type="button"
            class="header-left"
            @click="onBack"
          >
            <AppIcon name="back" class="back-icon-svg" />
          </button>

          <div class="header-title">
            {{ title }}
          </div>

          <button
            v-if="hasRight"
            type="button"
            class="header-right"
            @click="onRightClick"
          >
            <slot v-if="$slots.right" name="right" />

            <img
              v-else-if="rightIcon"
              class="right-icon"
              :src="rightIcon"
              alt=""
            />

            <span v-else class="right-text">
              {{ rightText }}
            </span>
          </button>
        </header>

        <div class="custom-area">
          <slot name="custom" />
        </div>

        <div
          v-if="menus.length"
          class="sub-menu"
          :style="{ top: '0px' }"
        >
          <div
            v-for="item in menus"
            :key="item.value"
            class="sub-menu-item"
            :class="{ active: item.value === currentKey }"
            @click="changeMenu(item.value)"
          >
            {{ item.label }}
          </div>
        </div>

        <main class="content">
          <slot :active-key="currentKey" />
        </main>
      </div>
    </template>

    <!-- 没有 custom：tabbar 固定顶部 -->
    <template v-else>
      <header class="header-bar fixed-header">
        <button
          v-if="showBack"
          type="button"
          class="header-left"
          @click="onBack"
        >
          <AppIcon name="back" class="back-icon-svg" />
        </button>

        <div class="header-title">
          {{ title }}
        </div>

        <button
          v-if="hasRight"
          type="button"
          class="header-right"
          @click="onRightClick"
        >
          <slot v-if="$slots.right" name="right" />

          <img
            v-else-if="rightIcon"
            class="right-icon"
            :src="rightIcon"
            alt=""
          />

          <span v-else class="right-text">
            {{ rightText }}
          </span>
        </button>
      </header>

      <div class="scroll-view no-custom">
        <div
          v-if="menus.length"
          class="sub-menu"
          :style="{ top: `${navHeight}px` }"
        >
          <div
            v-for="item in menus"
            :key="item.value"
            class="sub-menu-item"
            :class="{ active: item.value === currentKey }"
            @click="changeMenu(item.value)"
          >
            {{ item.label }}
          </div>
        </div>

        <main class="content">
          <slot :active-key="currentKey" />
        </main>
      </div>
    </template>
  </div>
</template>

<script setup>
import { computed, ref, useSlots, watch } from 'vue'
import AppIcon from '@/components/common/AppIcon.vue'

const props = defineProps({
  title: {
    type: String,
    default: '',
  },

  showBack: {
    type: Boolean,
    default: true,
  },

  rightText: {
    type: String,
    default: '',
  },

  rightIcon: {
    type: String,
    default: '',
  },

  menus: {
    type: Array,
    default: () => [],
  },

  modelValue: {
    type: [String, Number],
    default: '',
  },

  navHeight: {
    type: Number,
    default: 58,
  },
})

const emit = defineEmits([
  'update:modelValue',
  'back',
  'right-click',
  'menu-change',
])

const slots = useSlots()

const hasCustom = computed(() => !!slots.custom)

const hasRight = computed(() => {
  return !!slots.right || !!props.rightIcon || !!props.rightText
})

const currentKey = ref(
  props.modelValue || props.menus?.[0]?.value || '',
)

watch(
  () => props.modelValue,
  value => {
    if (value !== currentKey.value) {
      currentKey.value = value
    }
  },
)

watch(
  () => props.menus,
  value => {
    if (!currentKey.value && value?.length) {
      currentKey.value = value[0].value
    }
  },
  {
    deep: true,
  },
)

function changeMenu(value) {
  currentKey.value = value
  emit('update:modelValue', value)
  emit('menu-change', value)
}

function onBack() {
  emit('back')
}

function onRightClick() {
  emit('right-click')
}
</script>

<style scoped>
.page-414 {
  width: min(100%, 414px);
  height: 100vh;
  height: 100dvh;
  margin: 0 auto;
  background: #0b0d16;
  color: #ffffff;
  position: relative;
  overflow: hidden;
  box-sizing: border-box;
  font-family:
    -apple-system,
    BlinkMacSystemFont,
    "PingFang SC",
    "Microsoft YaHei",
    Arial,
    sans-serif;
}

/* 滚动容器 */
.scroll-view {
  height: 100%;
  overflow-y: auto;
  overflow-x: hidden;
  box-sizing: border-box;
  background: #0b0d16;
  -webkit-overflow-scrolling: touch;
}

.scroll-view::-webkit-scrollbar {
  width: 0;
  height: 0;
}

/* 没有 custom 时，内容避开顶部 tabbar */
.scroll-view.no-custom {
  padding-top: 58px;
}

/* 顶部导航 */
.header-bar {
  width: 100%;
  height: 58px;
  background: #0b0d16;
  position: relative;
  z-index: 40;
  box-sizing: border-box;
  flex-shrink: 0;
}

/* 没有 custom 时，固定在 page-414 内部顶部 */
.fixed-header {
  position: absolute;
  top: 0;
  left: 0;
  transform: none;
}

/* 左侧返回按钮 */
.header-left {
  position: absolute;
  left: 24px;
  top: 0;
  width: 44px;
  height: 58px;
  border: none;
  outline: none;
  padding: 0;
  margin: 0;
  background: transparent;
  color: #ffffff;
  display: flex;
  align-items: center;
  justify-content: flex-start;
  box-sizing: border-box;
  cursor: pointer;
  user-select: none;
  appearance: none;
  -webkit-appearance: none;
}

/* 返回图标 */
.back-icon-svg {
  width: 24px;
  height: 24px;
  display: block;
  color: #ffffff;
}

/* 标题永远居中 */
.header-title {
  position: absolute;
  left: 100px;
  right: 100px;
  top: 0;
  height: 58px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
  font-size: 22px;
  font-weight: 800;
  text-align: center;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  pointer-events: none;
}

/* 右侧按钮，可没有 */
.header-right {
  position: absolute;
  right: 24px;
  top: 0;
  min-width: 44px;
  max-width: 120px;
  height: 58px;
  border: none;
  outline: none;
  padding: 0;
  margin: 0;
  background: transparent;
  color: #21ff00;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  box-sizing: border-box;
  font-size: 15px;
  font-weight: 800;
  cursor: pointer;
  user-select: none;
  appearance: none;
  -webkit-appearance: none;
}

.right-text {
  max-width: 108px;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.right-icon {
  width: 22px;
  height: 22px;
  object-fit: contain;
  display: block;
}

/* 自定义区域 */
.custom-area {
  background: #0b0d16;
  box-sizing: border-box;
}

/* 二级菜单 */
.sub-menu {
  width: 100%;
  height: 58px;
  background: #1b1e29;
  display: flex;
  align-items: center;
  position: sticky;
  z-index: 30;
  box-sizing: border-box;
}

/* 二级菜单 item */
.sub-menu-item {
  flex: 1;
  height: 58px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 15px;
  color: #9ca1ad;
  font-weight: 600;
  position: relative;
  cursor: pointer;
  user-select: none;
}

.sub-menu-item.active {
  color: #21ff00;
  font-weight: 700;
}

.sub-menu-item.active::after {
  content: '';
  position: absolute;
  bottom: 8px;
  width: 24px;
  height: 3px;
  border-radius: 3px;
  background: #21ff00;
}

/* 内容区域 */
.content {
  min-height: calc(100vh - 58px);
  background: #0b0d16;
  box-sizing: border-box;
  padding: 0;
}
</style>
