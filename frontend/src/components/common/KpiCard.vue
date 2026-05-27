<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { countUp } from '@/utils/countup'
import SvgIcon from '@/components/common/SvgIcon.vue'

const props = defineProps<{
  icon: string
  label: string
  value: number
  unit?: string
  trend?: { dir: 'up' | 'down'; text: string }
  iconBg?: string
  sparkline?: number[]
}>()

const displayValue = ref(0)

onMounted(() => {
  countUp(0, props.value, 800, (v) => {
    displayValue.value = v
  })
})
</script>

<template>
  <view class="kpi-card card">
    <view class="kpi-top">
      <text class="kpi-label">{{ label }}</text>
      <view class="kpi-icon" :style="{ background: iconBg || 'var(--color-primary-light)' }">
        <SvgIcon :name="icon" />
      </view>
    </view>
    <text class="kpi-num">{{ unit }}{{ displayValue.toLocaleString() }}</text>
    <view v-if="trend" class="kpi-trend">
      <text :class="trend.dir === 'up' ? 'trend-up' : 'trend-down'">
        <SvgIcon :name="trend.dir === 'up' ? 'arrow-up' : 'arrow-down'" />
        {{ trend.text }}
      </text>
    </view>
    <view v-if="sparkline?.length" class="sparkline">
      <view
        v-for="(v, i) in sparkline"
        :key="i"
        class="spark-bar"
        :style="{ height: (v / Math.max(...sparkline) * 100) + '%' }"
      />
    </view>
  </view>
</template>

<style lang="scss" scoped>
.kpi-card {
  padding: 18px 20px;
  display: flex; flex-direction: column; gap: 10px;
}
.kpi-top {
  display: flex; align-items: flex-start; justify-content: space-between;
}
.kpi-label { font-size: 12px; color: var(--color-text-muted); }
.kpi-icon {
  width: 38px; height: 38px; border-radius: 10px;
  display: flex; align-items: center; justify-content: center; font-size: 20px;
}
.kpi-num {
  font-size: 26px; font-weight: 700;
  color: var(--color-text-primary); line-height: 1;
}
.kpi-trend { display: flex; align-items: center; gap: 6px; font-size: 12px; }
.trend-up { color: #10b981; }
.trend-down { color: #ef4444; }
.sparkline {
  height: 28px; display: flex; align-items: flex-end; gap: 2px;
}
.spark-bar {
  flex: 1; border-radius: 2px 2px 0 0;
  background: linear-gradient(to top, var(--color-primary-dark), var(--color-primary));
  opacity: 0.7;
  transition: height 0.3s ease;
}
</style>
