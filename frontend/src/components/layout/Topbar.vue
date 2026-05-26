<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useThemeStore } from '@/stores/theme'
import type { ThemeKey } from '@/stores/theme'

defineProps<{
  breadcrumbs: { label: string; path?: string }[]
}>()

const themeStore = useThemeStore()
const timeStr = ref('')
let timer: ReturnType<typeof setInterval>

function updateTime() {
  const now = new Date()
  const pad = (n: number) => String(n).padStart(2, '0')
  timeStr.value = `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())} ${pad(now.getHours())}:${pad(now.getMinutes())}:${pad(now.getSeconds())}`
}

onMounted(() => {
  updateTime()
  timer = setInterval(updateTime, 1000)
})
onUnmounted(() => clearInterval(timer))

const themes: { key: ThemeKey; bg: string }[] = [
  { key: 'a', bg: '#0d1117' },
  { key: 'b', bg: '#1e40af' },
  { key: 'c', bg: 'linear-gradient(135deg,#7c3aed,#f59e0b)' },
]
</script>

<template>
  <view class="topbar">
    <view class="breadcrumb">
      <template v-for="(crumb, idx) in breadcrumbs" :key="idx">
        <text v-if="idx > 0" class="sep">›</text>
        <text :class="idx === breadcrumbs.length - 1 ? 'cur' : 'parent'">{{ crumb.label }}</text>
      </template>
    </view>

    <view class="topbar-right">
      <text class="time-display">{{ timeStr }}</text>

      <view class="icon-btn" @click="uni.navigateTo({ url: '/pages/messages/index' })">
        💬
        <view class="notif-dot" />
      </view>

      <view class="theme-switcher">
        <text class="t-label">主题</text>
        <view
          v-for="t in themes"
          :key="t.key"
          class="t-btn"
          :class="{ active: themeStore.current === t.key }"
          :style="{ background: t.bg }"
          @click="themeStore.apply(t.key)"
        >{{ t.key.toUpperCase() }}</view>
      </view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.topbar {
  height: var(--topbar-height);
  background: var(--color-topbar-bg);
  border-bottom: 1px solid var(--color-border-light);
  display: flex; align-items: center; padding: 0 24px; gap: 16px;
  position: sticky; top: 0; z-index: 10;
  box-shadow: 0 1px 4px rgba(0,0,0,0.04);
}
.breadcrumb { display: flex; align-items: center; gap: 6px; }
.sep { color: var(--color-border); font-size: 13px; }
.parent { font-size: 13px; color: var(--color-text-muted); }
.cur { font-size: 13px; color: var(--color-text-primary); font-weight: 600; }
.topbar-right { margin-left: auto; display: flex; align-items: center; gap: 12px; }
.time-display {
  font-size: 12px; color: var(--color-text-secondary);
  background: var(--color-border-light); padding: 5px 12px;
  border-radius: 8px; border: 1px solid var(--color-border);
}
.icon-btn {
  width: 34px; height: 34px; border-radius: 8px;
  background: var(--color-border-light); border: 1px solid var(--color-border);
  display: flex; align-items: center; justify-content: center;
  font-size: 15px; cursor: pointer; position: relative;
}
.notif-dot {
  position: absolute; top: 7px; right: 7px;
  width: 7px; height: 7px; background: #ef4444;
  border-radius: 50%; border: 1.5px solid var(--color-topbar-bg);
}
.theme-switcher {
  display: flex; align-items: center; gap: 6px;
  background: var(--color-card-bg); border-radius: 50px;
  padding: 5px 10px; box-shadow: var(--color-card-shadow);
  border: 1px solid var(--color-border);
}
.t-label { font-size: 11px; color: var(--color-text-muted); }
.t-btn {
  width: 24px; height: 24px; border-radius: 50%;
  border: 2px solid transparent; cursor: pointer;
  display: flex; align-items: center; justify-content: center;
  font-size: 9px; font-weight: 700; color: #fff;
  transition: all 0.2s;
  &.active { border-color: var(--color-text-primary); transform: scale(1.2); }
}
</style>
