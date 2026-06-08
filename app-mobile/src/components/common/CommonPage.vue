<template>
  <div
    class="page-414"
    :style="pageStyle"
  >
    <!--
      滚动头部模式：
      有二级菜单 或 有 custom 时使用
      tabbar 跟随内容滚动，二级菜单 sticky 吸顶
    -->
    <template v-if="scrollHeader">
      <div class="scroll-view full-scroll">
        <!-- 自定义 tabbar -->
        <div
          v-if="$slots.tabbar"
          class="tabbar-wrap"
        >
          <slot
            name="tabbar"
            :title="title"
            :show-back="showBack"
            :right-text="rightText"
            :right-icon="rightIcon"
            :on-back="onBack"
            :on-right-click="onRightClick"
          />
        </div>

        <!-- 默认 tabbar -->
        <header
          v-else
          class="header-bar"
        >
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
            >

            <span v-else class="right-text">
              {{ rightText }}
            </span>
          </button>
        </header>

        <!-- 自定义区域：有才显示 -->
        <div
          v-if="$slots.custom"
          class="custom-area"
        >
          <slot name="custom" />
        </div>

        <!-- 二级菜单：sticky 吸顶 -->
        <div
          v-if="menus.length"
          class="sub-menu sticky-sub-menu"
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

    <!--
      只有 tabbar：
      tabbar 固定顶部，content 滚动
    -->
    <template v-else>
      <!-- 自定义 tabbar 固定顶部 -->
      <div
        v-if="$slots.tabbar"
        class="fixed-header custom-tabbar-wrap"
      >
        <slot
          name="tabbar"
          :title="title"
          :show-back="showBack"
          :right-text="rightText"
          :right-icon="rightIcon"
          :on-back="onBack"
          :on-right-click="onRightClick"
        />
      </div>

      <!-- 默认 tabbar 固定顶部 -->
      <header
        v-else
        class="header-bar fixed-header"
      >
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
          >

          <span v-else class="right-text">
            {{ rightText }}
          </span>
        </button>
      </header>

      <div class="scroll-view fixed-body-scroll">
        <main class="content">
          <slot :active-key="currentKey" />
        </main>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, useSlots, watch, type PropType } from 'vue'
import AppIcon from '@/components/common/AppIcon.vue'

type MenuValue = string | number

type MenuItem = {
  label: string
  value: MenuValue
}

type TabbarSlotProps = {
  title: string
  showBack: boolean
  rightText: string
  rightIcon: string
  onBack: () => void
  onRightClick: () => void
}

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
    type: Array as PropType<MenuItem[]>,
    default: () => [],
  },

  modelValue: {
    type: [String, Number] as PropType<MenuValue>,
    default: '',
  },

  navHeight: {
    type: Number,
    default: 58,
  },
})

defineSlots<{
  tabbar?: (props: TabbarSlotProps) => unknown
  right?: () => unknown
  custom?: () => unknown
  default?: (props: { activeKey: MenuValue }) => unknown
}>()

const emit = defineEmits<{
  'update:modelValue': [value: MenuValue]
  back: []
  'right-click': []
  'menu-change': [value: MenuValue]
}>()

const slots = useSlots()

const pageStyle = computed(() => {
  return {
    '--nav-height': `${props.navHeight}px`,
  }
})

const hasRight = computed(() => {
  return !!slots.right || !!props.rightIcon || !!props.rightText
})

/**
 * 关键逻辑：
 * 有二级菜单：tabbar 跟内容一起滚动，二级菜单 sticky 吸顶
 * 有 custom：tabbar + custom 一起滚动
 * 没菜单没 custom：只有 tabbar，tabbar 固定
 */
const scrollHeader = computed(() => {
  return props.menus.length > 0 || !!slots.custom
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
      emit('update:modelValue', value[0].value)
      emit('menu-change', value[0].value)
    }
  },
  {
    deep: true,
  },
)

function changeMenu(value: MenuValue) {
  if (value === currentKey.value) return

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
  width: min(100%, var(--app-width, 414px));
  height: 100%;
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

/* 通用滚动容器 */
.scroll-view {
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

/* 有二级菜单 / 有 custom 时，整个页面滚动 */
.full-scroll {
  height: 100%;
}

/* 只有 tabbar 时，content 从 tabbar 下方开始滚动 */
.fixed-body-scroll {
  position: absolute;
  left: 0;
  right: 0;
  top: var(--nav-height);
  bottom: 0;
}

/* 默认 tabbar */
.header-bar {
  width: 100%;
  height: var(--nav-height);
  background: #0b0d16;
  position: relative;
  z-index: 40;
  box-sizing: border-box;
  flex-shrink: 0;
}

/* 自定义 tabbar 外层 */
.tabbar-wrap {
  width: 100%;
  height: var(--nav-height);
  background: #0b0d16;
  position: relative;
  z-index: 40;
  box-sizing: border-box;
  flex-shrink: 0;
}

/* 只有 tabbar 时固定顶部 */
.fixed-header {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: var(--nav-height);
  background: #0b0d16;
  transform: none;
  z-index: 90;
}

/* 自定义 tabbar 固定外层 */
.custom-tabbar-wrap {
  box-sizing: border-box;
}

/* 左侧返回按钮 */
.header-left {
  position: absolute;
  left: 24px;
  top: 0;
  width: 44px;
  height: var(--nav-height);
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
  height: var(--nav-height);
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
  height: var(--nav-height);
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
  height: var(--nav-height);
  background: #1b1e29;
  display: flex;
  align-items: center;
  box-sizing: border-box;
}

/* 二级菜单吸顶 */
.sticky-sub-menu {
  position: sticky;
  top: 0;
  z-index: 80;
}

.sub-menu-item {
  flex: 1;
  height: var(--nav-height);
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
  min-height: calc(100vh - var(--nav-height));
  background: #0b0d16;
  box-sizing: border-box;
  padding: 0;
}
</style>
