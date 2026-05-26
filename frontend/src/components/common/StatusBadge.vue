<script setup lang="ts">
const props = defineProps<{
  status: 'success' | 'warning' | 'danger' | 'info' | 'default'
  label: string
}>()

const colorMap = {
  success: { bg: '#f0fdf4', color: '#16a34a' },
  warning: { bg: '#fffbeb', color: '#d97706' },
  danger:  { bg: '#fef2f2', color: '#dc2626' },
  info:    { bg: '#eff6ff', color: '#1e40af' },
  default: { bg: '#f9fafb', color: '#6b7280' },
}
</script>

<template>
  <view class="badge" :style="{ background: colorMap[status].bg, color: colorMap[status].color }">
    <view class="dot" :style="{ background: colorMap[status].color }" />
    <text>{{ label }}</text>
  </view>
</template>

<style lang="scss" scoped>
.badge {
  display: inline-flex; align-items: center; gap: 4px;
  padding: 3px 9px; border-radius: 5px;
  font-size: 11px; font-weight: 500;
}
.dot {
  width: 5px; height: 5px; border-radius: 50%; flex-shrink: 0;
}
</style>
