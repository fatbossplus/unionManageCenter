<script setup lang="ts">
import { ref, reactive } from 'vue'
import SvgIcon from '@/components/common/SvgIcon.vue'

export interface FilterField {
  key: string
  label: string
  type: 'input' | 'select' | 'daterange'
  placeholder?: string
  options?: { label: string; value: string }[]
}

export interface QuickTag {
  key: string
  label: string
  color: string
  params: Record<string, unknown>
}

const props = defineProps<{
  fields: FilterField[]
  quickTags?: QuickTag[]
}>()

const emit = defineEmits<{
  search: [params: Record<string, unknown>]
  reset: []
  export: []
}>()

const form = reactive<Record<string, unknown>>({})
const activeTag = ref<string | null>(null)
const expanded = ref(true)

function selectTag(tag: QuickTag) {
  if (activeTag.value === tag.key) {
    activeTag.value = null
    Object.keys(tag.params).forEach(k => delete form[k])
  } else {
    activeTag.value = tag.key
    Object.assign(form, tag.params)
  }
}

function handleSearch() {
  const params: Record<string, unknown> = {}
  Object.keys(form).forEach(k => { if (form[k] !== '' && form[k] !== undefined) params[k] = form[k] })
  emit('search', params)
}

function handleReset() {
  Object.keys(form).forEach(k => delete form[k])
  activeTag.value = null
  emit('reset')
}

function getSelectedLabel(field: FilterField): string {
  if (!form[field.key]) return field.placeholder || '请选择'
  const opt = (field.options || []).find(o => o.value === form[field.key])
  return opt?.label || String(form[field.key])
}
</script>

<template>
  <view class="filter-panel card">
    <view class="fp-header">
      <view class="fp-title">
        <SvgIcon name="search" style="font-size:14px;color:var(--color-primary)" />
        <text>筛选条件</text>
        <view
          v-if="Object.values(form).filter(v => v !== '' && v !== undefined).length"
          class="active-count"
        >{{ Object.values(form).filter(v => v !== '' && v !== undefined).length }}</view>
      </view>
      <view class="fp-toggle" @click="expanded = !expanded">
        <SvgIcon :name="expanded ? 'chevron-up' : 'chevron-down'" />
        <text>{{ expanded ? '收起' : '展开' }}</text>
      </view>
    </view>

    <view v-if="expanded">
      <view class="fp-fields">
        <view v-for="field in fields" :key="field.key" class="fp-group">
          <text class="fp-label">{{ field.label }}</text>
          <view class="fp-control">
            <input
              v-if="field.type === 'input'"
              v-model="form[field.key] as string"
              :placeholder="field.placeholder || '请输入'"
              class="fp-input"
            />
            <picker
              v-else-if="field.type === 'select'"
              :range="field.options || []"
              range-key="label"
              @change="(e: any) => form[field.key] = (field.options || [])[e.detail.value]?.value"
            >
              <view class="fp-select">
                <text class="fp-select-text">{{ getSelectedLabel(field) }}</text>
                <SvgIcon name="chevron-down" class="fp-arrow" />
              </view>
            </picker>
          </view>
        </view>
      </view>

      <view v-if="quickTags?.length" class="fp-tags">
        <text class="tag-label">快捷筛选：</text>
        <view
          v-for="tag in quickTags"
          :key="tag.key"
          class="qf-tag"
          :class="{ active: activeTag === tag.key }"
          @click="selectTag(tag)"
        >
          <view class="qf-dot" :style="{ background: tag.color }" />
          <text>{{ tag.label }}</text>
        </view>
      </view>

      <view class="fp-actions">
        <view class="fp-btn fp-btn-primary" @click="handleSearch">
          <SvgIcon name="search" /> 查询
        </view>
        <view class="fp-btn fp-btn-outline" @click="handleReset">
          <SvgIcon name="refresh" /> 重置
        </view>
        <view class="fp-actions-right">
          <view class="fp-btn fp-btn-outline" @click="emit('export')">
            <SvgIcon name="export" /> 导出
          </view>
          <slot name="extra-actions" />
        </view>
      </view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.filter-panel { padding: 18px 20px; margin-bottom: 16px; }
.fp-header {
  display: flex; align-items: center; justify-content: space-between; margin-bottom: 14px;
}
.fp-title { display: flex; align-items: center; gap: 8px; font-size: 13px; font-weight: 600; color: var(--color-text-primary); }
.active-count {
  background: #ef4444; color: #fff;
  font-size: 10px; padding: 1px 5px; border-radius: 4px; font-weight: 400;
}
.fp-toggle {
  font-size: 12px; color: var(--color-primary); cursor: pointer;
  display: flex; align-items: center; gap: 3px;
}
.fp-fields {
  display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; margin-bottom: 12px;
}
.fp-group { display: flex; flex-direction: column; gap: 5px; }
.fp-label { font-size: 11px; color: var(--color-text-secondary); font-weight: 500; }
.fp-control {
  height: 34px; border: 1px solid var(--color-border); border-radius: 7px;
  background: var(--color-card-bg); overflow: hidden;
  &:focus-within { border-color: var(--color-primary); }
}
.fp-input {
  width: 100%; height: 32px; border: none; outline: none;
  background: transparent; padding: 0 10px;
  font-size: 13px; color: var(--color-text-primary);
}
.fp-select {
  height: 34px; display: flex; align-items: center;
  padding: 0 10px; justify-content: space-between;
}
.fp-select-text { font-size: 13px; color: var(--color-text-secondary); }
.fp-arrow { font-size: 11px; color: var(--color-text-muted); }
.fp-tags {
  display: flex; align-items: center; gap: 8px; flex-wrap: wrap; margin-bottom: 14px;
}
.tag-label { font-size: 11px; color: var(--color-text-muted); }
.qf-tag {
  height: 26px; padding: 0 10px; border-radius: 5px;
  border: 1px solid var(--color-border); font-size: 11px;
  color: var(--color-text-secondary); cursor: pointer;
  display: flex; align-items: center; gap: 4px;
  background: var(--color-card-bg);
  &.active {
    background: var(--color-primary-light);
    border-color: var(--color-primary);
    color: var(--color-primary); font-weight: 600;
  }
}
.qf-dot { width: 6px; height: 6px; border-radius: 50%; }
.fp-actions { display: flex; align-items: center; gap: 8px; }
.fp-actions-right { margin-left: auto; display: flex; gap: 8px; }
.fp-btn {
  height: 34px; border-radius: 7px; border: none; cursor: pointer;
  font-size: 13px; font-weight: 500;
  display: flex; align-items: center; gap: 6px; padding: 0 16px;
}
.fp-btn-primary { background: var(--color-primary); color: #fff; }
.fp-btn-outline {
  background: var(--color-card-bg); color: var(--color-text-secondary);
  border: 1px solid var(--color-border);
}
</style>
