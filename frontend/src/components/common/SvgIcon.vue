<script setup lang="ts">
defineProps<{ name: string; size?: string | number }>()
</script>

<template>
  <!-- #ifdef H5 -->
  <svg
    class="svg-icon"
    :style="size ? `width:${size}px;height:${size}px` : ''"
    aria-hidden="true"
  >
    <use :xlink:href="`#icon-${name}`" />
  </svg>
  <!-- #endif -->
  <!-- #ifndef H5 -->
  <text class="svg-icon-fallback">·</text>
  <!-- #endif -->
</template>

<style scoped>
.svg-icon {
  display: inline-block;
  width: 1em;
  height: 1em;
  fill: currentColor;
  overflow: hidden;
  vertical-align: -0.15em;
  flex-shrink: 0;
}
.svg-icon-fallback {
  font-size: 0.8em;
}
</style>
