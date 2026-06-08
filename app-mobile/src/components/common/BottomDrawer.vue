<script setup lang="ts">
import AppIcon from './AppIcon.vue'

withDefaults(
  defineProps<{
    modelValue: boolean
    title?: string
    ariaLabel?: string
    showClose?: boolean
    showHandle?: boolean
    closeOnBackdrop?: boolean
    closeLabel?: string
    maxHeight?: string
    zIndex?: number
  }>(),
  {
    title: '',
    ariaLabel: '',
    showClose: true,
    showHandle: true,
    closeOnBackdrop: true,
    closeLabel: '关闭',
    maxHeight: '72dvh',
    zIndex: 1000,
  },
)

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
          <span v-if="showHandle" class="bottom-drawer__handle" aria-hidden="true" />

          <slot v-if="$slots.header" name="header" :close="close" />
          <header v-else-if="title || showClose" class="bottom-drawer__header">
            <h2 v-if="title">
              {{ title }}
            </h2>
            <button
              v-if="showClose"
              type="button"
              class="bottom-drawer__close"
              :aria-label="closeLabel"
              @click="close"
            >
              <AppIcon name="close" class="bottom-drawer__close-icon" />
            </button>
          </header>

          <div class="bottom-drawer__body">
            <slot :close="close" />
          </div>

          <footer v-if="$slots.footer" class="bottom-drawer__footer">
            <slot name="footer" :close="close" />
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
  background: var(--overlay-bg);
  backdrop-filter: blur(8px);
}

.bottom-drawer__panel {
  position: relative;
  display: flex;
  flex-direction: column;
  width: var(--app-width, 100vw);
  max-height: var(--bottom-drawer-max-height);
  padding: var(--page-padding-y) var(--page-padding-x) calc(var(--page-padding-y) + env(safe-area-inset-bottom));
  overflow: hidden;
  border-radius: var(--radius-xl) var(--radius-xl) 0 0;
  background: var(--sheet-bg);
  color: var(--text);
  box-shadow: var(--shadow-drawer);
  box-sizing: border-box;
  touch-action: pan-y;
}

.bottom-drawer__handle {
  display: block;
  flex: 0 0 auto;
  width: clamp(42px, 2.7rem, 62px);
  height: clamp(4px, .25rem, 6px);
  margin: 0 auto var(--space-lg);
  border-radius: 999px;
  background: var(--handle-bg);
}

.bottom-drawer__header {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  min-height: var(--icon-button-size);
  margin-bottom: var(--space-sm);
}

.bottom-drawer__header h2 {
  margin: 0;
  color: var(--text);
  font-size: 1.1rem;
  font-weight: 700;
  line-height: 1.2;
  text-align: center;
}

.bottom-drawer__close {
  position: absolute;
  top: 50%;
  right: 2px;
  width: var(--icon-button-size);
  height: var(--icon-button-size);
  transform: translateY(-50%);
  border: 0;
  background: transparent;
  color: var(--text);
  font: inherit;
  font-size: 1.55rem;
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
