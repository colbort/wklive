<script setup lang="ts">
defineProps<{
  name: 'back' | 'globe' | 'headset' | 'mail' | 'lock' | 'eye' | 'eye-off' | 'close-circle'
    | 'search'
    | 'close'
    | 'chevron-down'
}>()
</script>

<template>
  <svg :class="`app-icon--${name}`" viewBox="0 0 24 24" aria-hidden="true">
    <template v-if="name === 'back'">
      <path d="m14.5 5-7 7 7 7" />
    </template>

    <template v-else-if="name === 'globe'">
      <path d="M12 2.75a9.25 9.25 0 1 0 0 18.5 9.25 9.25 0 0 0 0-18.5Z" />
      <path d="M3.25 12h17.5M12 2.75c2.4 2.53 3.61 5.61 3.61 9.25s-1.21 6.72-3.61 9.25M12 2.75C9.6 5.28 8.39 8.36 8.39 12s1.21 6.72 3.61 9.25" />
    </template>

    <template v-else-if="name === 'headset'">
      <path d="M5 12.5v-1.25a7 7 0 0 1 14 0v1.25" />
      <path d="M5 12.5h2.5v6H5a2 2 0 0 1-2-2v-2a2 2 0 0 1 2-2ZM19 12.5h-2.5v6H19a2 2 0 0 0 2-2v-2a2 2 0 0 0-2-2Z" />
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

    <template v-else-if="name === 'close'">
      <path d="M5.5 5.5 18.5 18.5M18.5 5.5 5.5 18.5" />
    </template>

    <template v-else-if="name === 'chevron-down'">
      <path d="m7 9 5 5 5-5" />
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
  stroke-width: 2.4;
}

.app-icon__fill {
  fill: #c5c8d0;
  stroke: #c5c8d0;
}

.app-icon__cut {
  stroke: #252733;
  stroke-width: 2.4;
}
</style>
