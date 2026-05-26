<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  total: number
  page: number
  pageSize: number
}>()

const emit = defineEmits<{
  pageChange: [page: number]
  pageSizeChange: [size: number]
}>()

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))

const visiblePages = computed(() => {
  const pages: (number | '...')[] = []
  const p = props.page
  const t = totalPages.value
  if (t <= 7) {
    for (let i = 1; i <= t; i++) pages.push(i)
  } else {
    pages.push(1)
    if (p > 3) pages.push('...')
    for (let i = Math.max(2, p - 1); i <= Math.min(t - 1, p + 1); i++) pages.push(i)
    if (p < t - 2) pages.push('...')
    pages.push(t)
  }
  return pages
})

const pageSizeOptions = [10, 20, 50, 100]
</script>

<template>
  <view class="pagination-wrap">
    <text class="page-info">
      共 <text class="em">{{ total }}</text> 条，第 <text class="em">{{ page }}</text> / <text class="em">{{ totalPages }}</text> 页
    </text>

    <view class="page-btns">
      <view
        class="page-btn"
        :class="{ disabled: page <= 1 }"
        @click="page > 1 && emit('pageChange', page - 1)"
      >‹</view>
      <view
        v-for="p in visiblePages"
        :key="String(p)"
        class="page-btn"
        :class="{ active: p === page, ellipsis: p === '...' }"
        @click="typeof p === 'number' && emit('pageChange', p)"
      >{{ p }}</view>
      <view
        class="page-btn"
        :class="{ disabled: page >= totalPages }"
        @click="page < totalPages && emit('pageChange', page + 1)"
      >›</view>
    </view>

    <picker
      :range="pageSizeOptions"
      :value="pageSizeOptions.indexOf(pageSize)"
      @change="(e: any) => emit('pageSizeChange', pageSizeOptions[e.detail.value])"
    >
      <view class="page-size-select">每页 {{ pageSize }} 条 ▾</view>
    </picker>
  </view>
</template>

<style lang="scss" scoped>
.pagination-wrap {
  display: flex; align-items: center; justify-content: space-between;
  padding: 14px 16px; border-top: 1px solid var(--color-border-light);
  background: var(--color-border-light);
}
.page-info { font-size: 12px; color: var(--color-text-muted); }
.em { color: var(--color-text-primary); font-weight: 600; }
.page-btns { display: flex; align-items: center; gap: 6px; }
.page-btn {
  min-width: 32px; height: 32px; border-radius: 7px;
  border: 1px solid var(--color-border); background: var(--color-card-bg);
  font-size: 13px; color: var(--color-text-primary);
  display: flex; align-items: center; justify-content: center;
  cursor: pointer; padding: 0 8px;
  &:hover { border-color: var(--color-primary); color: var(--color-primary); }
  &.active { background: var(--color-primary); color: #fff; border-color: var(--color-primary); }
  &.disabled { opacity: 0.4; cursor: not-allowed; &:hover { border-color: var(--color-border); color: var(--color-text-primary); } }
  &.ellipsis { border: none; background: transparent; cursor: default; &:hover { color: var(--color-text-primary); } }
}
.page-size-select {
  font-size: 12px; color: var(--color-text-secondary);
  border: 1px solid var(--color-border); padding: 5px 10px;
  border-radius: 6px; background: var(--color-card-bg);
}
</style>
