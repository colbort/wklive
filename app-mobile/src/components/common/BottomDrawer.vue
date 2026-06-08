<script setup lang="ts">
withDefaults(defineProps<{
  modelValue: boolean
  title?: string
  ariaLabel?: string
  showClose?: boolean
  showHandle?: boolean
  closeOnBackdrop?: boolean
  closeLabel?: string
  maxHeight?: string
  zIndex?: number
}>(), {
  title: '',
  ariaLabel: '',
  showClose: true,
  showHandle: true,
  closeOnBackdrop: true,
  closeLabel: '关闭',
  maxHeight: '72dvh',
  zIndex: 1000,
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  close: []
}>()

defineSlots<{
  header?: (props: { close: () => void }) => unknown
  default?: (props: { close: () => void }) => unknown
  footer?: (props: { close: () => void }) => unknown
}>()

function close() {
  emit('update:modelValue', false)
  emit('close')
}

function closeFromBackdrop(enabled: boolean) {
  if (enabled) close()
}
</script>

<template>
  <Teleport to="body">
    <Transition name="bottom-drawer">
      <div
        v-if="modelValue"
        class="bottom-drawer"
        :style="{
          '--bottom-drawer-max-height': maxHeight,
          '--bottom-drawer-z-index': zIndex,
        }"
        role="presentation"
        @click.self="closeFromBackdrop(closeOnBackdrop)"
      >
        <section
          class="bottom-drawer__panel"
          role="dialog"
          aria-modal="true"
          :aria-label="ariaLabel || title"
        >
          <span
            v-if="showHandle"
            class="bottom-drawer__handle"
            aria-hidden="true"
          />

          <slot
            v-if="$slots.header"
            name="header"
            :close="close"
          />
          <header
            v-else-if="title || showClose"
            class="bottom-drawer__header"
          >
            <h2 v-if="title">{{ title }}</h2>
            <button
              v-if="showClose"
              type="button"
              class="bottom-drawer__close"
              :aria-label="closeLabel"
              @click="close"
            >
              ×
            </button>
          </header>

          <div class="bottom-drawer__body">
            <slot :close="close" />
          </div>

          <footer
            v-if="$slots.footer"
            class="bottom-drawer__footer"
          >
            <slot
              name="footer"
              :close="close"
            />
          </footer>
        </section>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.bottom-drawer {
  position: fixed;
  inset: 0;
  z-index: var(--bottom-drawer-z-index);
  display: flex;
  align-items: flex-end;
  justify-content: center;
  overflow: hidden;
  background: rgba(3, 4, 10, 0.08);
  backdrop-filter: blur(8px);
}

.bottom-drawer__panel {
  position: relative;
  display: flex;
  flex-direction: column;
  width: min(100%, 414px);
  max-height: var(--bottom-drawer-max-height);
  padding: 22px 22px calc(18px + env(safe-area-inset-bottom));
  overflow: hidden;
  border-radius: 28px 28px 0 0;
  background: #24252d;
  color: #f6f7fb;
  box-shadow: 0 -24px 70px rgba(0, 0, 0, 0.42);
  box-sizing: border-box;
  touch-action: pan-y;
}

.bottom-drawer__handle {
  display: block;
  flex: 0 0 auto;
  width: 54px;
  height: 5px;
  margin: 0 auto 22px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.52);
}

.bottom-drawer__header {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  min-height: 42px;
  margin-bottom: 10px;
}

.bottom-drawer__header h2 {
  margin: 0;
  color: #ffffff;
  font-size: 22px;
  font-weight: 700;
  line-height: 1.2;
  text-align: center;
}

.bottom-drawer__close {
  position: absolute;
  top: 50%;
  right: 2px;
  width: 40px;
  height: 40px;
  transform: translateY(-50%);
  border: 0;
  background: transparent;
  color: #ffffff;
  font: inherit;
  font-size: 31px;
  line-height: 1;
  cursor: pointer;
}

.bottom-drawer__body {
  flex: 1 1 auto;
  min-height: 0;
  overflow: hidden auto;
  overscroll-behavior: contain;
  -webkit-overflow-scrolling: touch;
}

.bottom-drawer__body::-webkit-scrollbar {
  width: 0;
  height: 0;
}

.bottom-drawer__footer {
  flex: 0 0 auto;
}

.bottom-drawer-enter-active,
.bottom-drawer-leave-active {
  transition: opacity 0.18s ease;
}

.bottom-drawer-enter-active .bottom-drawer__panel,
.bottom-drawer-leave-active .bottom-drawer__panel {
  transition: transform 0.18s ease;
}

.bottom-drawer-enter-from,
.bottom-drawer-leave-to {
  opacity: 0;
}

.bottom-drawer-enter-from .bottom-drawer__panel,
.bottom-drawer-leave-to .bottom-drawer__panel {
  transform: translateY(18px);
}
</style>
