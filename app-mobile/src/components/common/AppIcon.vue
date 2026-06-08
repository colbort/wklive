<script setup lang="ts">
defineProps<{
  name:
    | 'back'
    | 'globe'
    | 'headset'
    | 'mail'
    | 'lock'
    | 'eye'
    | 'eye-off'
    | 'close-circle'
    | 'search'
    | 'pause'
    | 'star'
    | 'chevron-up'
    | 'close'
    | 'chevron-down'
    | 'id-card'
    | 'credit-card'
    | 'shield-check'
    | 'user-plus'
    | 'cloud-download'
    | 'info-circle'
    | 'logout'
    | 'nav-home'
    | 'nav-market'
    | 'nav-trade'
    | 'nav-assets'
    | 'nav-profile'
}>()
</script>

<template>
  <svg :class="`app-icon--${name}`" viewBox="0 0 24 24" aria-hidden="true">
    <template v-if="name === 'back'">
      <path d="m14.5 5-7 7 7 7" />
    </template>

    <template v-else-if="name === 'globe'">
      <path d="M12 2.75a9.25 9.25 0 1 0 0 18.5 9.25 9.25 0 0 0 0-18.5Z" />
      <path
        d="M3.25 12h17.5M12 2.75c2.4 2.53 3.61 5.61 3.61 9.25s-1.21 6.72-3.61 9.25M12 2.75C9.6 5.28 8.39 8.36 8.39 12s1.21 6.72 3.61 9.25"
      />
    </template>

    <template v-else-if="name === 'headset'">
      <path d="M4.25 13v-1.35a7.75 7.75 0 0 1 15.5 0V13" />
      <path d="M4.25 13h2.3v5.25h-2.3a1.9 1.9 0 0 1-1.9-1.9V14.9a1.9 1.9 0 0 1 1.9-1.9Z" />
      <path d="M19.75 13h-2.3v5.25h2.3a1.9 1.9 0 0 0 1.9-1.9V14.9a1.9 1.9 0 0 0-1.9-1.9Z" />
      <path d="M7.25 18.25c.55 1.55 1.9 2.32 4.05 2.32h1.2" />
    </template>

    <template v-else-if="name === 'mail'">
      <path d="M4.5 6.5h15v11h-15z" />
      <path d="m5.25 7.25 6.75 5.5 6.75-5.5" />
    </template>

    <template v-else-if="name === 'lock'">
      <path d="M6.5 10.5h13v9h-13z" />
      <path d="M9 10.5V8a4 4 0 0 1 8 0v2.5" />
    </template>

    <template v-else-if="name === 'eye'">
      <path d="M2.75 12s3.5-6 9.25-6 9.25 6 9.25 6-3.5 6-9.25 6-9.25-6-9.25-6Z" />
      <path d="M9 12a3 3 0 1 0 6 0 3 3 0 0 0-6 0Z" />
    </template>

    <template v-else-if="name === 'eye-off'">
      <path d="M2.75 12s3.5-6 9.25-6 9.25 6 9.25 6-3.5 6-9.25 6-9.25-6-9.25-6Z" />
      <path d="M9.75 9.75a3.18 3.18 0 0 1 4.5 4.5M4 4l16 16" />
    </template>

    <template v-else-if="name === 'search'">
      <path d="M10.75 18.5a7.25 7.25 0 1 0 0-14.5 7.25 7.25 0 0 0 0 14.5ZM16 16l4 4" />
    </template>

    <template v-else-if="name === 'pause'">
      <path class="app-icon__solid" d="M7 5h3.5v14H7V5Zm6.5 0H17v14h-3.5V5Z" />
    </template>

    <template v-else-if="name === 'star'">
      <path
        d="m12 3.5 2.55 5.16 5.7.83-4.12 4.02.97 5.68L12 16.5l-5.1 2.69.97-5.68-4.12-4.02 5.7-.83L12 3.5Z"
      />
    </template>

    <template v-else-if="name === 'chevron-up'">
      <path d="m6.5 15 5.5-5.5 5.5 5.5" />
    </template>

    <template v-else-if="name === 'close'">
      <path d="M5.5 5.5 18.5 18.5M18.5 5.5 5.5 18.5" />
    </template>

    <template v-else-if="name === 'chevron-down'">
      <path d="m7 9 5 5 5-5" />
    </template>

    <template v-else-if="name === 'id-card'">
      <path
        d="M4.5 6h13.6a1.4 1.4 0 0 1 1.4 1.4v8.35a1.4 1.4 0 0 1-1.4 1.4H4.5a1.4 1.4 0 0 1-1.4-1.4V7.4A1.4 1.4 0 0 1 4.5 6Z"
      />
      <path
        d="M8 11.65a1.78 1.78 0 1 0 0-3.56 1.78 1.78 0 0 0 0 3.56ZM5.75 14.95c.43-1.17 1.18-1.75 2.25-1.75s1.82.58 2.25 1.75M12.75 9.5h3.2M12.75 13.25h2.35"
      />
      <path d="M16.9 16.3a3.25 3.25 0 1 0 0-6.5 3.25 3.25 0 0 0 0 6.5Z" />
      <path d="m15.45 13.05.9.9 1.9-2.05" />
    </template>

    <template v-else-if="name === 'credit-card'">
      <path
        d="M3.75 6.5h16.5a1.5 1.5 0 0 1 1.5 1.5v8.75a1.5 1.5 0 0 1-1.5 1.5H3.75a1.5 1.5 0 0 1-1.5-1.5V8a1.5 1.5 0 0 1 1.5-1.5Z"
      />
      <path d="M2.75 10.4h18.5M6.25 15.25h1.8M10.5 15.25h1.8" />
    </template>

    <template v-else-if="name === 'shield-check'">
      <path d="M12 3.25 19 6v5.25c0 4.2-2.47 6.98-7 8.5-4.53-1.52-7-4.3-7-8.5V6l7-2.75Z" />
      <path d="m8.75 12.25 2.1 2.1 4.4-4.7" />
    </template>

    <template v-else-if="name === 'user-plus'">
      <path d="M3.75 19.25c.58-3.25 2.55-4.88 5.9-4.88 1.8 0 3.23.48 4.28 1.43" />
      <path d="M9.65 11.75a3.7 3.7 0 1 0 0-7.4 3.7 3.7 0 0 0 0 7.4Z" />
      <path d="M15.25 15.75h5.3M18.45 12.8l2.95 2.95-2.95 2.95" />
    </template>

    <template v-else-if="name === 'cloud-download'">
      <path
        d="M7.25 18.25H6.8a4.05 4.05 0 0 1-.58-8.06 5.75 5.75 0 0 1 10.95-1.42 4.53 4.53 0 0 1 .25 9.04h-.67"
      />
      <path d="M12 12.25v7M8.75 16.25 12 19.75l3.25-3.5" />
    </template>

    <template v-else-if="name === 'info-circle'">
      <path d="M12 3.25a8.75 8.75 0 1 0 0 17.5 8.75 8.75 0 0 0 0-17.5Z" />
      <path d="M12 10.5v5.25M12 7.5h.01" />
    </template>

    <template v-else-if="name === 'logout'">
      <path d="M13.25 5.25H6.5a1.5 1.5 0 0 0-1.5 1.5v10.5a1.5 1.5 0 0 0 1.5 1.5h6.75" />
      <path d="M14.5 9.25 18.75 12l-4.25 2.75M9.75 12h8.5" />
    </template>

    <template v-else-if="name === 'nav-home'">
      <path
        class="app-icon__nav-home-body"
        d="M24.5 22.333c0 .645-.522 1.167-1.167 1.167H4.667A1.167 1.167 0 0 1 3.5 22.333V10.071c0-.36.166-.7.45-.921l9.334-7.26a1.166 1.166 0 0 1 1.432 0l9.333 7.26c.285.22.451.56.451.92v12.263Zm-2.333-1.166V10.64L14 4.29l-8.167 6.352v10.525h16.334Zm-14-4.667h11.666v2.333H8.167V16.5Z"
      />
      <path
        class="app-icon__nav-active-shape"
        d="M24.5 22.333c0 .645-.522 1.167-1.167 1.167H4.667A1.167 1.167 0 0 1 3.5 22.333V10.071c0-.36.166-.7.45-.921l9.334-7.26a1.166 1.166 0 0 1 1.432 0l9.333 7.26c.285.22.451.56.451.92v12.263Z"
      />
      <path
        class="app-icon__nav-active-shape app-icon__nav-active-cut"
        d="M8.167 16.5h11.666v2.333H8.167V16.5Z"
      />
    </template>

    <template v-else-if="name === 'nav-market'">
      <path
        class="app-icon__nav-frame"
        fill-rule="evenodd"
        clip-rule="evenodd"
        d="M19.33 4.2H8A3.8 3.8 0 0 0 4.2 8v11.33a3.8 3.8 0 0 0 3.8 3.8h11.33a3.8 3.8 0 0 0 3.8-3.8V8a3.8 3.8 0 0 0-3.8-3.8ZM8 2a6 6 0 0 0-6 6v11.33a6 6 0 0 0 6 6h11.33a6 6 0 0 0 6-6V8a6 6 0 0 0-6-6H8Z"
      />
      <path
        class="app-icon__nav-symbol"
        d="m8.25 16.15 4-4 2.1 2.1 4.8-4.85 1.65 1.65-6.45 6.5-2.1-2.1-2.4 2.4-1.6-1.7Z"
      />
      <path
        class="app-icon__nav-active-shape"
        d="M19.33 2a6 6 0 0 1 6 6v11.33a6 6 0 0 1-6 6H8a6 6 0 0 1-6-6V8a6 6 0 0 1 6-6h11.33Zm-4.54 11.387-2.58-2.58L7 16.018l1.823 1.825 3.387-3.387 2.58 2.58L20 11.822 18.176 10l-3.387 3.387Z"
      />
    </template>

    <template v-else-if="name === 'nav-trade'">
      <path
        class="app-icon__nav-frame"
        fill-rule="evenodd"
        clip-rule="evenodd"
        d="M19.33 4.2H8A3.8 3.8 0 0 0 4.2 8v11.33a3.8 3.8 0 0 0 3.8 3.8h11.33a3.8 3.8 0 0 0 3.8-3.8V8a3.8 3.8 0 0 0-3.8-3.8ZM8 2a6 6 0 0 0-6 6v11.33a6 6 0 0 0 6 6h11.33a6 6 0 0 0 6-6V8a6 6 0 0 0-6-6H8Z"
      />
      <path
        class="app-icon__nav-active-shape"
        d="M19.33 2a6 6 0 0 1 6 6v11.33a6 6 0 0 1-6 6H8a6 6 0 0 1-6-6V8a6 6 0 0 1 6-6h11.33ZM8 15.167h10.5V17.5h-4.667V21L8 15.167Zm5.833-4.667V7l5.834 5.833h-10.5V10.5h4.666Z"
      />
      <path
        class="app-icon__nav-symbol"
        d="M8 15.167h10.5V17.5h-4.667V21L8 15.167Zm5.833-4.667V7l5.834 5.833h-10.5V10.5h4.666Z"
      />
    </template>

    <template v-else-if="name === 'nav-assets'">
      <path
        class="app-icon__nav-frame"
        fill-rule="evenodd"
        clip-rule="evenodd"
        d="M19.33 4.2H8A3.8 3.8 0 0 0 4.2 8v11.33a3.8 3.8 0 0 0 3.8 3.8h11.33a3.8 3.8 0 0 0 3.8-3.8V8a3.8 3.8 0 0 0-3.8-3.8ZM8 2a6 6 0 0 0-6 6v11.33a6 6 0 0 0 6 6h11.33a6 6 0 0 0 6-6V8a6 6 0 0 0-6-6H8Z"
      />
      <path
        class="app-icon__nav-active-shape"
        d="M19.33 2a6 6 0 0 1 6 6v11.33a6 6 0 0 1-6 6H8a6 6 0 0 1-6-6V8a6 6 0 0 1 6-6h11.33Zm-5.555 6 5.774 5.775-5.774 5.774L8 13.775 13.775 8Zm0 3.3L11.3 13.775l2.475 2.475 2.475-2.475-2.475-2.475Z"
      />
      <path
        class="app-icon__nav-symbol"
        d="m13.775 8 5.774 5.775-5.774 5.774L8 13.775 13.775 8Zm0 3.3L11.3 13.775l2.475 2.475 2.475-2.475-2.475-2.475Z"
      />
    </template>

    <template v-else-if="name === 'nav-profile'">
      <path
        class="app-icon__nav-frame"
        fill-rule="evenodd"
        clip-rule="evenodd"
        d="M19.33 4.2H8A3.8 3.8 0 0 0 4.2 8v11.33a3.8 3.8 0 0 0 3.8 3.8h11.33a3.8 3.8 0 0 0 3.8-3.8V8a3.8 3.8 0 0 0-3.8-3.8ZM8 2a6 6 0 0 0-6 6v11.33a6 6 0 0 0 6 6h11.33a6 6 0 0 0 6-6V8a6 6 0 0 0-6-6H8Z"
      />
      <path
        class="app-icon__nav-active-shape"
        d="M19.33 2a6 6 0 0 1 6 6v11.33a6 6 0 0 1-6 6H8a6 6 0 0 1-6-6V8a6 6 0 0 1 6-6h11.33ZM9.043 21.21A11.64 11.64 0 0 1 7 20.07a8.16 8.16 0 0 1 6.862-3.737c2.799 0 5.27 1.409 6.74 3.555-.625.463-1.299.864-2.012 1.196a5.825 5.825 0 0 0-4.728-2.417 5.827 5.827 0 0 0-4.819 2.544Zm4.7-6.043a4.083 4.083 0 1 1 0-8.167 4.083 4.083 0 0 1 0 8.167Zm0-2.334a1.75 1.75 0 1 0 0-3.5 1.75 1.75 0 0 0 0 3.5Z"
      />
      <path
        class="app-icon__nav-symbol"
        d="M9.043 21.21A11.64 11.64 0 0 1 7 20.07a8.16 8.16 0 0 1 6.862-3.737c2.799 0 5.27 1.409 6.74 3.555-.625.463-1.299.864-2.012 1.196a5.825 5.825 0 0 0-4.728-2.417 5.827 5.827 0 0 0-4.819 2.544Zm4.7-6.043a4.083 4.083 0 1 1 0-8.167 4.083 4.083 0 0 1 0 8.167Zm0-2.334a1.75 1.75 0 1 0 0-3.5 1.75 1.75 0 0 0 0 3.5Z"
      />
    </template>

    <template v-else>
      <path class="app-icon__fill" d="M12 3.25a8.75 8.75 0 1 0 0 17.5 8.75 8.75 0 0 0 0-17.5Z" />
      <path class="app-icon__cut" d="m9 9 6 6M15 9l-6 6" />
    </template>
  </svg>
</template>

<style scoped>
svg {
  display: block;
}

path {
  fill: none;
  stroke: currentColor;
  stroke-linecap: round;
  stroke-linejoin: round;
  stroke-width: 2;
}

.app-icon--back path {
  stroke-width: 1.4;
}

.app-icon--nav-home path,
.app-icon--nav-market path,
.app-icon--nav-trade path,
.app-icon--nav-assets path,
.app-icon--nav-profile path {
  fill: currentColor;
  stroke: none;
}

.app-icon--nav-home,
.app-icon--nav-market,
.app-icon--nav-trade,
.app-icon--nav-assets,
.app-icon--nav-profile {
  overflow: visible;
}

.app-icon__nav-active-shape {
  display: none;
}

.app-icon--nav-home .app-icon__nav-active-cut {
  fill: var(--nav-symbol-fill);
}

.app-icon__fill {
  fill: var(--text-icon-muted);
  stroke: var(--text-icon-muted);
}

.app-icon__solid {
  fill: currentColor;
  stroke: none;
}

.app-icon__cut {
  stroke: var(--control-bg-soft);
  stroke-width: 1.4;
}
</style>
