<script setup lang="ts">
import Sidebar from './Sidebar.vue'
import Topbar from './Topbar.vue'

defineProps<{
  breadcrumbs: { label: string; path?: string }[]
}>()
</script>

<template>
  <view class="app-layout">
    <Sidebar />
    <view class="layout-main">
      <Topbar :breadcrumbs="breadcrumbs" />
      <view class="layout-content">
        <slot />
      </view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.app-layout {
  display: flex;
  min-height: 100vh;
  background: var(--color-bg);
}
.layout-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  overflow: hidden;
}
.layout-content {
  flex: 1;
  overflow-y: auto;
  padding: 20px 24px;
}
</style>
