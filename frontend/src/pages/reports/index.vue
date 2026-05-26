<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import KpiCard from '@/components/common/KpiCard.vue'
import FilterPanel from '@/components/common/FilterPanel.vue'
import { get } from '@/api/request'
import type { FilterField } from '@/components/common/FilterPanel.vue'

const breadcrumbs = [{ label: '首页' }, { label: '财务 & 报表' }, { label: '数据报表' }]
const activeTab = ref('user')
const loadingStats = ref(false)
const loadingDaily  = ref(false)
const days = ref(30)

const filterFields: FilterField[] = [
  { key: 'type',      label: '报表类型', type: 'select', options: [
    { label:'用户增长',value:'user' }, { label:'流水趋势',value:'revenue' },
    { label:'联盟分析',value:'org'  }, { label:'订单统计',value:'order' },
  ]},
  { key: 'dimension', label: '时间范围', type: 'select', options: [
    { label:'近7天',value:'7' }, { label:'近30天',value:'30' }, { label:'近90天',value:'90' },
  ]},
]

const tabs = [
  { key: 'user',    label: '👥 用户增长', color: '#3b82f6', color2: '#8b5cf6' },
  { key: 'revenue', label: '💰 流水趋势', color: '#10b981', color2: '' },
  { key: 'org',     label: '🏢 活跃用户', color: '#f59e0b', color2: '' },
  { key: 'order',   label: '📦 订单统计', color: '#ef4444', color2: '' },
]

interface SummaryStat { icon: string; label: string; value: number; trend: { dir: 'up'|'down'; text: string }; bg: string }
const summaryStats = ref<SummaryStat[]>([])

interface DailyRow { date: string; new_users: number; active_users: number; revenue: number; orders: number; trend: string }
const dailyRows = ref<DailyRow[]>([])

async function loadSummary() {
  loadingStats.value = true
  try {
    const res: any = await get('/reports/summary')
    summaryStats.value = [
      { icon:'👥', label:'本月新增用户', value: res.month_new_users ?? 0,
        trend:{ dir:(res.trends?.users||'').startsWith('+')?'up':'down', text:res.trends?.users||'0%' }, bg:'#eff6ff' },
      { icon:'💰', label:'本月总流水', value: res.month_revenue ?? 0,
        trend:{ dir:(res.trends?.revenue||'').startsWith('+')?'up':'down', text:res.trends?.revenue||'0%' }, bg:'#f0fdf4' },
      { icon:'🏢', label:'活跃联盟数', value: res.active_orgs ?? 0,
        trend:{ dir:'up', text:'实时统计' }, bg:'#fff7ed' },
      { icon:'📦', label:'本月订单数', value: res.month_orders ?? 0,
        trend:{ dir:(res.trends?.orders||'').startsWith('+')?'up':'down', text:res.trends?.orders||'0%' }, bg:'#faf5ff' },
    ]
  } catch (e: any) {
    uni.showToast({ title: e?.message || '统计数据加载失败', icon: 'none' })
  } finally { loadingStats.value = false }
}

async function loadDaily(d = 30) {
  loadingDaily.value = true
  try {
    const res: any[] = await get('/reports/daily', { days: d }) as any[]
    dailyRows.value = res || []
  } catch (e: any) {
    uni.showToast({ title: e?.message || '明细数据加载失败', icon: 'none' })
  } finally { loadingDaily.value = false }
}

function onSearch(params: Record<string, unknown>) {
  if (params.type) activeTab.value = params.type as string
  const d = parseInt(params.dimension as string) || 30
  days.value = d; loadDaily(d)
}

onMounted(() => { loadSummary(); loadDaily() })

// ═══════════════════════════════════════════════
// 报表 SVG 折线图：坐标轴 + Hover 看板
// ═══════════════════════════════════════════════
const RPL = 58, RPR = 16, RPT = 18, RPB = 36
const RSVG_W = 720, RSVG_H = 240
const RCW = RSVG_W - RPL - RPR
const RCH = RSVG_H - RPT - RPB

// 根据 tab 取对应的系列数据
const series1 = computed<number[]>(() => {
  if (!dailyRows.value.length) return []
  switch (activeTab.value) {
    case 'user':    return dailyRows.value.map(r => r.new_users)
    case 'revenue': return dailyRows.value.map(r => r.revenue)
    case 'org':     return dailyRows.value.map(r => r.active_users)
    case 'order':   return dailyRows.value.map(r => r.orders)
    default:        return []
  }
})
const series2 = computed<number[]>(() => {
  if (activeTab.value !== 'user') return []
  return dailyRows.value.map(r => r.active_users)
})

const activeTabCfg = computed(() => tabs.find(t => t.key === activeTab.value) || tabs[0])

function niceMax(v: number) {
  if (v === 0) return 10
  const mag = Math.pow(10, Math.floor(Math.log10(v)))
  return Math.ceil(v / mag) * mag
}
const rYNiceMax = computed(() => niceMax(Math.max(...series1.value, ...series2.value, 1)))
const rYTicks   = computed(() => {
  const n = 5
  return Array.from({ length: n + 1 }, (_, i) => Math.round(rYNiceMax.value / n * i))
})
const rXLabels  = computed(() => {
  const d = dailyRows.value
  if (!d.length) return []
  const step = Math.max(1, Math.floor(d.length / 9))
  return d.map((r, i) => ({ label: r.date, i })).filter((_, i) => i % step === 0)
})

function rpx(i: number) { return RPL + (i / Math.max(dailyRows.value.length - 1, 1)) * RCW }
function rpy(v: number)  { return RPT + RCH - (v / rYNiceMax.value) * RCH }

function rSmoothPath(vals: number[]): string {
  if (vals.length < 2) return ''
  const pts = vals.map((v, i) => ({ x: rpx(i), y: rpy(v) }))
  let d = `M ${pts[0].x},${pts[0].y}`
  for (let i = 1; i < pts.length; i++) {
    const cpx = (pts[i-1].x + pts[i].x) / 2
    d += ` C ${cpx},${pts[i-1].y} ${cpx},${pts[i].y} ${pts[i].x},${pts[i].y}`
  }
  return d
}
function rAreaPath(vals: number[]): string {
  const p = rSmoothPath(vals)
  if (!p) return ''
  return p + ` L${rpx(vals.length-1)},${RPT+RCH} L${RPL},${RPT+RCH} Z`
}

function fmtY(v: number) {
  if (v >= 10000) return (v / 10000).toFixed(1) + 'w'
  if (v >= 1000)  return (v / 1000).toFixed(1) + 'k'
  return String(v)
}

// Hover
const rHoverIdx    = ref(-1)
const rHoverPx     = ref(0)
const rHoverPy1    = ref(0)
const rHoverPy2    = ref(0)
const rTooltipStyle = ref('')

function onReportMouseMove(e: MouseEvent) {
  const svg = e.currentTarget as SVGSVGElement
  const rect = svg.getBoundingClientRect()
  const mx = (e.clientX - rect.left) * (RSVG_W / rect.width)
  if (mx < RPL - 4 || mx > RSVG_W - RPR || !dailyRows.value.length) { rHoverIdx.value = -1; return }
  const idx = Math.round(Math.max(0, Math.min(dailyRows.value.length - 1, (mx - RPL) / RCW * (dailyRows.value.length - 1))))
  rHoverIdx.value = idx
  rHoverPx.value  = rpx(idx)
  rHoverPy1.value = rpy(series1.value[idx] ?? 0)
  rHoverPy2.value = series2.value.length ? rpy(series2.value[idx] ?? 0) : -999
  const wrapX = (rHoverPx.value / RSVG_W) * rect.width
  const side  = wrapX > rect.width * 0.65 ? 'right' : 'left'
  rTooltipStyle.value = side === 'left'
    ? `left:${wrapX + 12}px;top:8px`
    : `right:${rect.width - wrapX + 12}px;top:8px`
}
function onReportMouseLeave() { rHoverIdx.value = -1 }
</script>

<template>
  <AppLayout :breadcrumbs="breadcrumbs">
    <view class="stats-row">
      <KpiCard v-for="s in summaryStats" :key="s.label"
        :icon="s.icon" :label="s.label" :value="s.value" :trend="s.trend" :icon-bg="s.bg" />
      <view v-if="loadingStats && !summaryStats.length" class="stat-loading">统计加载中…</view>
    </view>

    <FilterPanel :fields="filterFields" @search="onSearch" @reset="() => { loadSummary(); loadDaily() }" @export="()=>{}" />

    <view class="card report-panel">
      <view class="report-tabs">
        <view v-for="t in tabs" :key="t.key" class="r-tab"
          :class="{ active: activeTab === t.key }" @click="activeTab = t.key">{{ t.label }}</view>
      </view>
      <!-- ══ 报表折线图 ══ -->
      <!-- #ifdef H5 -->
      <view class="report-chart-wrap" v-if="dailyRows.length">
        <svg
          :viewBox="`0 0 ${RSVG_W} ${RSVG_H}`"
          preserveAspectRatio="xMidYMid meet"
          style="width:100%;height:auto;display:block;cursor:crosshair"
          @mousemove="onReportMouseMove"
          @mouseleave="onReportMouseLeave"
        >
          <defs>
            <linearGradient id="rFill1" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" :stop-color="activeTabCfg.color" stop-opacity="0.25"/>
              <stop offset="100%" :stop-color="activeTabCfg.color" stop-opacity="0"/>
            </linearGradient>
            <linearGradient id="rFill2" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stop-color="#8b5cf6" stop-opacity="0.2"/>
              <stop offset="100%" stop-color="#8b5cf6" stop-opacity="0"/>
            </linearGradient>
            <clipPath id="rChartClip">
              <rect :x="RPL" :y="RPT" :width="RCW" :height="RCH"/>
            </clipPath>
          </defs>

          <!-- Y 轴网格 + 标签 -->
          <g v-for="tick in rYTicks" :key="tick">
            <line :x1="RPL" :y1="rpy(tick)" :x2="RPL+RCW" :y2="rpy(tick)"
              stroke="currentColor" stroke-opacity="0.07" stroke-width="1"/>
            <text :x="RPL-8" :y="rpy(tick)+4" text-anchor="end"
              font-size="11" fill="currentColor" fill-opacity="0.45">{{ fmtY(tick) }}</text>
          </g>
          <!-- Y 轴线 -->
          <line :x1="RPL" :y1="RPT" :x2="RPL" :y2="RPT+RCH" stroke="currentColor" stroke-opacity="0.15"/>
          <!-- Y 轴标题 -->
          <text :x="12" :y="RPT+RCH/2" transform-origin="12 120"
            :transform="`rotate(-90,12,${RPT+RCH/2})`"
            font-size="11" fill="currentColor" fill-opacity="0.4" text-anchor="middle">
            {{ activeTab==='revenue' ? '金额(元)' : activeTab==='user' ? '人数' : '数量' }}
          </text>

          <!-- X 轴标签 -->
          <g v-for="lb in rXLabels" :key="lb.i">
            <text :x="rpx(lb.i)" :y="RPT+RCH+22" text-anchor="middle"
              font-size="10" fill="currentColor" fill-opacity="0.45">{{ lb.label }}</text>
            <line :x1="rpx(lb.i)" :y1="RPT+RCH" :x2="rpx(lb.i)" :y2="RPT+RCH+5"
              stroke="currentColor" stroke-opacity="0.2"/>
          </g>
          <!-- X 轴线 -->
          <line :x1="RPL" :y1="RPT+RCH" :x2="RPL+RCW" :y2="RPT+RCH" stroke="currentColor" stroke-opacity="0.15"/>
          <!-- X 轴标题 -->
          <text :x="RPL+RCW/2" :y="RSVG_H-2" text-anchor="middle"
            font-size="11" fill="currentColor" fill-opacity="0.4">日期</text>

          <!-- 面积 + 折线 -->
          <g clip-path="url(#rChartClip)">
            <path v-if="series2.length" :d="rAreaPath(series2)" fill="url(#rFill2)"/>
            <path :d="rAreaPath(series1)" fill="url(#rFill1)"/>
            <path v-if="series2.length" :d="rSmoothPath(series2)" fill="none" stroke="#8b5cf6" stroke-width="1.8"/>
            <path :d="rSmoothPath(series1)" fill="none" :stroke="activeTabCfg.color" stroke-width="2.5"/>
          </g>

          <!-- Hover -->
          <g v-if="rHoverIdx >= 0" style="pointer-events:none">
            <line :x1="rHoverPx" :y1="RPT" :x2="rHoverPx" :y2="RPT+RCH"
              stroke="#94a3b8" stroke-width="1" stroke-dasharray="4 3"/>
            <circle :cx="rHoverPx" :cy="rHoverPy1" r="5" :fill="activeTabCfg.color" stroke="#fff" stroke-width="2"/>
            <circle v-if="series2.length" :cx="rHoverPx" :cy="rHoverPy2" r="5" fill="#8b5cf6" stroke="#fff" stroke-width="2"/>
          </g>
        </svg>

        <!-- Hover 看板 -->
        <view v-if="rHoverIdx >= 0 && dailyRows[rHoverIdx]" class="r-tooltip" :style="rTooltipStyle">
          <view class="tt-date">📅 {{ dailyRows[rHoverIdx].date }}</view>
          <template v-if="activeTab === 'user'">
            <view class="tt-row">
              <view class="tt-dot" :style="{background:activeTabCfg.color}"/>
              <text class="tt-lbl">新增用户</text>
              <text class="tt-val">{{ dailyRows[rHoverIdx].new_users.toLocaleString() }}</text>
            </view>
            <view class="tt-row">
              <view class="tt-dot" style="background:#8b5cf6"/>
              <text class="tt-lbl">活跃用户</text>
              <text class="tt-val">{{ dailyRows[rHoverIdx].active_users.toLocaleString() }}</text>
            </view>
          </template>
          <template v-else-if="activeTab === 'revenue'">
            <view class="tt-row">
              <view class="tt-dot" :style="{background:activeTabCfg.color}"/>
              <text class="tt-lbl">流水</text>
              <text class="tt-val">¥{{ dailyRows[rHoverIdx].revenue.toLocaleString() }}</text>
            </view>
          </template>
          <template v-else-if="activeTab === 'org'">
            <view class="tt-row">
              <view class="tt-dot" :style="{background:activeTabCfg.color}"/>
              <text class="tt-lbl">活跃用户</text>
              <text class="tt-val">{{ dailyRows[rHoverIdx].active_users.toLocaleString() }}</text>
            </view>
          </template>
          <template v-else>
            <view class="tt-row">
              <view class="tt-dot" :style="{background:activeTabCfg.color}"/>
              <text class="tt-lbl">订单数</text>
              <text class="tt-val">{{ dailyRows[rHoverIdx].orders.toLocaleString() }}</text>
            </view>
          </template>
          <view class="tt-row tt-divider">
            <text class="tt-lbl">环比</text>
            <text class="tt-val" :class="dailyRows[rHoverIdx].trend.startsWith('↑')?'up':'down'">
              {{ dailyRows[rHoverIdx].trend }}
            </text>
          </view>
        </view>

        <!-- 图例 -->
        <view class="r-legend">
          <view class="legend-item">
            <view class="legend-line" :style="{background:activeTabCfg.color}"/>
            <text>{{ activeTab==='user'?'新增用户': activeTab==='revenue'?'流水': activeTab==='org'?'活跃用户':'订单数' }}</text>
          </view>
          <view v-if="activeTab==='user'" class="legend-item">
            <view class="legend-line" style="background:#8b5cf6"/>
            <text>活跃用户</text>
          </view>
          <text class="legend-hint">← 鼠标划过查看数据</text>
        </view>
      </view>
      <view v-else-if="loadingDaily" class="chart-loading">图表数据加载中…</view>
      <view v-else class="chart-empty">暂无数据，请先录入业务数据</view>
      <!-- #endif -->
      <!-- #ifndef H5 -->
      <view class="chart-placeholder">
        <view class="chart-ph-inner">
          <text class="chart-ph-icon">📊</text>
          <text class="chart-ph-text">{{ tabs.find(t=>t.key===activeTab)?.label }} 趋势图</text>
          <text class="chart-ph-sub">图表仅支持 H5 平台</text>
        </view>
      </view>
      <!-- #endif -->

      <view class="report-table-title">
        近 {{ days }} 天明细
        <view v-if="loadingDaily" class="loading-tip">加载中…</view>
      </view>
      <view class="t-head">
        <text class="th" style="flex:1.5">日期</text>
        <text class="th" style="flex:1">新增用户</text>
        <text class="th" style="flex:1">活跃用户</text>
        <text class="th" style="flex:1.2">流水（元）</text>
        <text class="th" style="flex:1">订单数</text>
        <text class="th" style="flex:0.9">环比</text>
      </view>
      <view v-if="!dailyRows.length && !loadingDaily" class="empty-tip">暂无数据</view>
      <view v-for="row in dailyRows" :key="row.date" class="t-row">
        <text class="td t-muted" style="flex:1.5">{{ row.date }}</text>
        <text class="td" style="flex:1">{{ row.new_users.toLocaleString() }}</text>
        <text class="td" style="flex:1">{{ row.active_users.toLocaleString() }}</text>
        <text class="td amount" style="flex:1.2">¥{{ row.revenue.toLocaleString() }}</text>
        <text class="td" style="flex:1">{{ row.orders.toLocaleString() }}</text>
        <text class="td" :class="row.trend.startsWith('↑')?'trend-up':'trend-down'" style="flex:0.9">{{ row.trend }}</text>
      </view>
    </view>
  </AppLayout>
</template>

<style lang="scss" scoped>
.stats-row { display:grid; grid-template-columns:repeat(4,1fr); gap:12px; margin-bottom:16px; position:relative; }
.stat-loading { position:absolute; inset:0; display:flex; align-items:center; justify-content:center; font-size:12px; color:var(--color-text-muted); }
.report-panel { overflow:hidden; }
.report-tabs { display:flex; gap:0; border-bottom:1px solid var(--color-border-light); }
.r-tab { padding:14px 20px; font-size:13px; cursor:pointer; color:var(--color-text-secondary); border-bottom:2px solid transparent; &.active { color:var(--color-primary); border-bottom-color:var(--color-primary); font-weight:600; } }

/* 报表图 */
.report-chart-wrap {
  position: relative; margin: 16px; padding-bottom: 4px;
  svg { color: var(--color-text-primary); overflow: visible; }
}
.r-legend { display:flex; align-items:center; gap:20px; margin-top:4px; padding-left:4px; }
.legend-item { display:flex; align-items:center; gap:5px; font-size:11px; color:var(--color-text-secondary); }
.legend-line { width:14px; height:2.5px; border-radius:2px; }
.legend-hint { font-size:10px; color:var(--color-text-muted); margin-left:auto; }

/* Hover 看板 */
.r-tooltip {
  position: absolute; pointer-events: none; z-index: 30;
  background: var(--color-card-bg);
  border: 1px solid var(--color-border);
  border-radius: 10px; padding: 10px 14px;
  box-shadow: 0 8px 24px rgba(0,0,0,0.13);
  min-width: 160px;
}
.tt-date { font-size:11px; font-weight:600; color:var(--color-text-secondary); margin-bottom:8px; }
.tt-row { display:flex; align-items:center; gap:7px; margin-bottom:4px; &:last-child{ margin:0; } }
.tt-divider { padding-top:6px; border-top:1px solid var(--color-border-light); margin-top:4px; }
.tt-dot { width:8px; height:8px; border-radius:50%; flex-shrink:0; }
.tt-lbl { font-size:11px; color:var(--color-text-muted); flex:1; }
.tt-val { font-size:12px; font-weight:700; color:var(--color-text-primary); &.up{color:#16a34a;} &.down{color:#ef4444;} }

.chart-loading { height:200px; display:flex; align-items:center; justify-content:center; color:var(--color-text-muted); font-size:13px; margin:16px; }
.chart-empty   { height:180px; display:flex; align-items:center; justify-content:center; background:var(--color-border-light); border-radius:10px; margin:16px; color:var(--color-text-muted); font-size:13px; }
.chart-placeholder { height:200px; display:flex; align-items:center; justify-content:center; background:var(--color-border-light); margin:16px; border-radius:10px; }
.chart-ph-inner { text-align:center; }
.chart-ph-icon { font-size:36px; display:block; margin-bottom:10px; }
.chart-ph-text { font-size:15px; font-weight:600; color:var(--color-text-primary); display:block; }
.chart-ph-sub  { font-size:12px; color:var(--color-text-muted); display:block; margin-top:4px; }

.report-table-title { padding:14px 16px 8px; font-size:13px; font-weight:600; color:var(--color-text-secondary); display:flex; align-items:center; gap:12px; }
.t-head { display:flex; padding:10px 16px; background:var(--color-border-light); border-bottom:1px solid var(--color-border); }
.th { font-size:12px; font-weight:600; color:var(--color-text-secondary); padding-right:8px; }
.t-row { display:flex; align-items:center; padding:11px 16px; border-bottom:1px solid var(--color-border-light); &:hover{background:var(--color-border-light);} }
.td { font-size:13px; color:var(--color-text-primary); padding-right:8px; }
.t-muted { font-size:12px; color:var(--color-text-muted); }
.amount { font-weight:600; color:#16a34a; }
.trend-up { color:#16a34a; font-size:12px; }
.trend-down { color:#ef4444; font-size:12px; }
.empty-tip { text-align:center; padding:40px; color:var(--color-text-muted); font-size:13px; }
.loading-tip { font-size:12px; color:var(--color-text-muted); font-weight:400; }
</style>
