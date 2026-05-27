<script setup lang="ts">
defineProps<{
  icon: string
  tip: string
  type?: 'view' | 'edit' | 'danger' | 'warn' | 'team' | 'key' | 'money' | 'default'
  disabled?: boolean
  size?: 'sm' | 'md'
}>()
defineEmits<{ (e: 'click'): void }>()
</script>

<template>
  <!-- #ifdef H5 -->
  <view
    class="ib-wrap"
    :class="[`ib-${type || 'default'}`, size === 'sm' ? 'ib-sm' : '', { 'ib-disabled': disabled }]"
    @click.stop="!disabled && $emit('click')"
  >
    <text class="ib-icon">{{ icon }}</text>
    <view class="ib-tooltip">{{ tip }}</view>
  </view>
  <!-- #endif -->
  <!-- #ifndef H5 -->
  <view class="ib-wrap" :class="[`ib-${type || 'default'}`, { 'ib-disabled': disabled }]"
    @click.stop="!disabled && $emit('click')">
    <text class="ib-icon">{{ icon }}</text>
  </view>
  <!-- #endif -->
</template>

<style lang="scss" scoped>
.ib-wrap {
  position: relative;
  width: 30px; height: 30px;
  border-radius: 7px;
  display: inline-flex; align-items: center; justify-content: center;
  cursor: pointer; flex-shrink: 0;
  transition: all 0.15s;
  &:hover { filter: brightness(0.93); transform: translateY(-1px); box-shadow: 0 2px 6px rgba(0,0,0,0.1); }
  &:active { transform: translateY(0); }
  &.ib-sm { width: 26px; height: 26px; border-radius: 5px; }
  &.ib-disabled { opacity: 0.4; cursor: not-allowed; pointer-events: none; }
}
.ib-icon { font-size: 14px; line-height: 1; user-select: none; }

/* 颜色主题 */
.ib-view    { background: var(--color-primary-light, #eff6ff); color: var(--color-primary, #3b82f6); }
.ib-edit    { background: #f0fdf4; color: #16a34a; }
.ib-danger  { background: #fef2f2; color: #ef4444; }
.ib-warn    { background: #fffbeb; color: #f59e0b; }
.ib-team    { background: #f5f3ff; color: #7c3aed; }
.ib-key     { background: #eff6ff; color: #3b82f6; }
.ib-money   { background: #f0fdf4; color: #059669; }
.ib-default { background: var(--color-border-light, #f1f5f9); color: var(--color-text-secondary, #64748b); }

/* Tooltip */
.ib-tooltip {
  position: absolute;
  bottom: calc(100% + 7px);
  left: 50%; transform: translateX(-50%);
  background: #1e293b; color: #fff;
  padding: 4px 9px; border-radius: 5px;
  font-size: 11px; white-space: nowrap;
  pointer-events: none;
  opacity: 0; transition: opacity 0.12s;
  z-index: 9999;
  box-shadow: 0 2px 8px rgba(0,0,0,0.2);
  &::after {
    content: '';
    position: absolute; top: 100%; left: 50%; transform: translateX(-50%);
    border: 4px solid transparent;
    border-top-color: #1e293b;
  }
}
.ib-wrap:hover .ib-tooltip { opacity: 1; }
</style>
