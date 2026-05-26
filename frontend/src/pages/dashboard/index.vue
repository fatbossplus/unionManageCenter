<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import KpiCard from '@/components/common/KpiCard.vue'
import {
  getDashboardStats, getTrendData, getOrgTypeDistrib,
  getOrgRank, getRealtimeEvents,
  type DashboardStats, type TrendPoint,
  type OrgTypeItem, type OrgRankItem, type EventItem,
} from '@/api/dashboard'
import { formatTime } from '@/utils/format'

const breadcrumbs = [{ label: '首页' }, { label: '数据大屏' }]

const stats    = ref<DashboardStats | null>(null)
const trendData = ref<TrendPoint[]>([])
const orgTypes  = ref<OrgTypeItem[]>([])
const orgRank   = ref<OrgRankItem[]>([])
const events    = ref<EventItem[]>([])
const onlineNum = ref(0)
const activePeriod = ref<'month' | 'quarter' | 'year'>('month')

let eventsTimer: ReturnType<typeof setInterval>

async function loadAll() {
  const [s, t, ot, or_, ev] = await Promise.allSettled([
    getDashboardStats(), getTrendData(activePeriod.value),
    getOrgTypeDistrib(), getOrgRank(), getRealtimeEvents(),
  ])
  if (s.status === 'fulfilled') { stats.value = s.value; onlineNum.value = s.value.onlineUsers }
  else uni.showToast({ title: '统计数据加载失败', icon: 'none' })
  if (t.status === 'fulfilled') trendData.value = t.value
  if (ot.status === 'fulfilled') orgTypes.value = ot.value
  if (or_.status === 'fulfilled') orgRank.value = or_.value
  if (ev.status === 'fulfilled') events.value = ev.value
}

async function changePeriod(p: 'month' | 'quarter' | 'year') {
  activePeriod.value = p
  try { trendData.value = await getTrendData(p) }
  catch (e: any) { uni.showToast({ title: e?.message || '加载失败', icon: 'none' }) }
}

onMounted(() => {
  loadAll()
  eventsTimer = setInterval(() => getRealtimeEvents().then(v => { events.value = v }).catch(() => {}), 30000)
})
onUnmounted(() => clearInterval(eventsTimer))

// ═══════════════════════════════════════════════
// 折线图 — 坐标轴 + Hover 看板
// ═══════════════════════════════════════════════
const PL = 62, PR = 16, PT = 16, PB = 38  // padding: left, right, top, bottom
const SVG_W = 640, SVG_H = 230
const CW = SVG_W - PL - PR   // chart area width
const CH = SVG_H - PT - PB   // chart area height

// 最大值（两系列共用同一 Y 轴比例，各自独立线但共享格）
const userMax  = computed(() => Math.max(...trendData.value.map(d => d.users),  1))
const revMax   = computed(() => Math.max(...trendData.value.map(d => d.revenue), 1))
const yMax     = computed(() => Math.max(userMax.value, revMax.value))

// Y 轴刻度：5 档，取整美观
function niceMax(v: number): number {
  if (v === 0) return 10
  const mag = Math.pow(10, Math.floor(Math.log10(v)))
  return Math.ceil(v / mag) * mag
}
const yNiceMax = computed(() => niceMax(yMax.value))
const yTicks   = computed(() => {
  const n = 5
  return Array.from({ length: n + 1 }, (_, i) => Math.round(yNiceMax.value / n * i))
})

// X 轴刻度：每隔 step 取一个日期标签
const xLabels = computed(() => {
  const d = trendData.value
  if (!d.length) return []
  const maxLabels = 8
  const step = Math.max(1, Math.floor(d.length / maxLabels))
  return d.map((p, i) => ({ label: p.date, i })).filter((_, i) => i % step === 0)
})

// 坐标计算
function px(i: number, total: number) { return PL + (i / Math.max(total - 1, 1)) * CW }
function py(v: number)                { return PT + CH - (v / yNiceMax.value) * CH }

// 折线平滑 path（三次贝塞尔）
function smoothPath(vals: number[], total: number): string {
  if (vals.length < 2) return ''
  const pts = vals.map((v, i) => ({ x: px(i, total), y: py(v) }))
  let d = `M ${pts[0].x},${pts[0].y}`
  for (let i = 1; i < pts.length; i++) {
    const cpx = (pts[i - 1].x + pts[i].x) / 2
    d += ` C ${cpx},${pts[i - 1].y} ${cpx},${pts[i].y} ${pts[i].x},${pts[i].y}`
  }
  return d
}
function areaPath(vals: number[], total: number): string {
  const p = smoothPath(vals, total)
  if (!p) return ''
  const lastX = px(vals.length - 1, total)
  return p + ` L${lastX},${PT + CH} L${PL},${PT + CH} Z`
}

// Hover 状态
const hoverIdx   = ref(-1)
const hoverPx    = ref(0)
const hoverPy1   = ref(0)  // users 点
const hoverPy2   = ref(0)  // revenue 点
const tooltipStyle = ref('')

function onLineMouseMove(e: MouseEvent) {
  const svg = e.currentTarget as SVGSVGElement
  const rect = svg.getBoundingClientRect()
  const mx = (e.clientX - rect.left) * (SVG_W / rect.width)
  if (mx < PL - 4 || mx > SVG_W - PR || !trendData.value.length) { hoverIdx.value = -1; return }
  const idx = Math.round(Math.max(0, Math.min(trendData.value.length - 1, (mx - PL) / CW * (trendData.value.length - 1))))
  hoverIdx.value = idx
  hoverPx.value  = px(idx, trendData.value.length)
  hoverPy1.value = py(trendData.value[idx].users)
  hoverPy2.value = py(trendData.value[idx].revenue)
  // 计算 tooltip 相对 .line-chart-wrap 的位置
  const wrapX = (hoverPx.value / SVG_W) * rect.width
  const side  = wrapX > rect.width * 0.65 ? 'right' : 'left'
  tooltipStyle.value = side === 'left'
    ? `left:${wrapX + 12}px;top:8px`
    : `right:${rect.width - wrapX + 12}px;top:8px`
}
function onLineMouseLeave() { hoverIdx.value = -1 }

// 格式化 Y 轴标签
function fmtY(v: number) {
  if (v >= 10000) return (v / 10000).toFixed(1) + 'w'
  if (v >= 1000)  return (v / 1000).toFixed(1) + 'k'
  return String(v)
}

// ═══════════════════════════════════════════════
// 圆环图 — SVG 弧形 + Hover 看板
// ═══════════════════════════════════════════════
const CX = 100, CY = 100, R = 80, IR = 54  // 圆心、外径、内径
const donutHoverIdx = ref(-1)
const donutTooltip  = ref<{ name: string; value: number; pct: string } | null>(null)
const donutTipStyle = ref('')

const donutTotal = computed(() => orgTypes.value.reduce((s, t) => s + t.value, 0))

function polarXY(r: number, deg: number) {
  const rad = (deg - 90) * Math.PI / 180
  return { x: CX + r * Math.cos(rad), y: CY + r * Math.sin(rad) }
}
function arcPath(startDeg: number, endDeg: number, expand = false): string {
  const r  = expand ? R + 5 : R
  const ir = expand ? IR - 3 : IR
  const s  = polarXY(r, startDeg), e = polarXY(r, endDeg)
  const si = polarXY(ir, endDeg),  ei = polarXY(ir, startDeg)
  const large = endDeg - startDeg > 180 ? 1 : 0
  return `M ${s.x} ${s.y} A ${r} ${r} 0 ${large} 1 ${e.x} ${e.y} L ${si.x} ${si.y} A ${ir} ${ir} 0 ${large} 0 ${ei.x} ${ei.y} Z`
}
const donutSegments = computed(() => {
  let offset = 0
  return orgTypes.value.map((t, i) => {
    const deg   = donutTotal.value > 0 ? (t.value / donutTotal.value) * 360 : 0
    const start = offset, end = offset + deg
    offset += deg
    return { ...t, start, end, i }
  })
})

function onDonutEnter(e: MouseEvent, idx: number) {
  donutHoverIdx.value = idx
  const t = orgTypes.value[idx]
  donutTooltip.value = {
    name: t.name, value: t.value,
    pct: donutTotal.value > 0 ? (t.value / donutTotal.value * 100).toFixed(1) + '%' : '0%',
  }
  const el = (e.currentTarget as HTMLElement).closest('.donut-chart-wrap') as HTMLElement
  const rect = el?.getBoundingClientRect()
  const mx = e.clientX - (rect?.left ?? 0)
  const my = e.clientY - (rect?.top  ?? 0)
  donutTipStyle.value = `left:${mx + 10}px;top:${my - 10}px`
}
function onDonutMove(e: MouseEvent) {
  if (donutHoverIdx.value < 0) return
  const el = (e.currentTarget as HTMLElement).closest('.donut-chart-wrap') as HTMLElement
  const rect = el?.getBoundingClientRect()
  donutTipStyle.value = `left:${e.clientX - (rect?.left ?? 0) + 10}px;top:${e.clientY - (rect?.top ?? 0) - 10}px`
}
function onDonutLeave() { donutHoverIdx.value = -1; donutTooltip.value = null }

// ═══════════════════════════════════════════════
// 排行榜 Hover
// ═══════════════════════════════════════════════
const rankHoverIdx = ref(-1)
</script>

<template>
  <AppLayout :breadcrumbs="breadcrumbs">

    <!-- 欢迎横幅 -->
    <view class="welcome-bar">
      <view>
        <view class="wb-title">
          <view class="live-dot" />
          <text>数据实时概览</text>
        </view>
        <text class="wb-sub">今日 {{ new Date().toLocaleDateString('zh-CN') }} · 实时更新中</text>
      </view>
      <view class="wb-stats">
        <view class="wb-stat">
          <text class="ws-num">{{ onlineNum.toLocaleString() }}</text>
          <text class="ws-lbl">当前在线</text>
        </view>
        <view class="ws-div" />
        <view class="wb-stat">
          <text class="ws-num">¥{{ (stats?.todayRevenue || 0).toLocaleString() }}</text>
          <text class="ws-lbl">今日流水</text>
        </view>
        <view class="ws-div" />
        <view class="wb-stat">
          <text class="ws-num">{{ stats?.todayNewUsers || 0 }}</text>
          <text class="ws-lbl">今日新增</text>
        </view>
      </view>
    </view>

    <!-- KPI 卡片 -->
    <view class="kpi-row" v-if="stats">
      <KpiCard icon="👥" label="注册用户总数" :value="stats.totalUsers"
        :trend="{ dir: 'up', text: '12.3% vs上月' }" icon-bg="#eff6ff"
        :sparkline="[40,55,48,70,62,85,78,100]" />
      <KpiCard icon="🏢" label="活跃联盟数" :value="stats.activeOrgs"
        :trend="{ dir: 'up', text: '5.1% vs上月' }" icon-bg="#f0fdf4"
        :sparkline="[50,45,65,60,75,70,90,85]" />
      <KpiCard icon="💰" label="本月流水(元)" :value="stats.monthlyRevenue"
        :trend="{ dir: 'up', text: '8.7% vs上月' }" icon-bg="#fff7ed"
        :sparkline="[30,50,45,65,75,70,90,100]" />
      <KpiCard icon="📦" label="待处理订单" :value="stats.pendingOrders"
        :trend="{ dir: 'down', text: '3单待跟进' }" icon-bg="#fef2f2"
        :sparkline="[80,90,70,60,50,55,40,35]" />
    </view>

    <!-- 图表行 -->
    <view class="chart-row">

      <!-- ══ 折线图：坐标轴 + Hover 数据看板 ══ -->
      <view class="card chart-card">
        <view class="card-header">
          <text class="card-title">近期趋势（用户增长 & 流水）</text>
          <view class="card-tabs">
            <text v-for="p in ['month','quarter','year']" :key="p"
              class="tab" :class="{ active: activePeriod === p }"
              @click="changePeriod(p as any)">
              {{ p === 'month' ? '近30天' : p === 'quarter' ? '近90天' : '近365天' }}
            </text>
          </view>
        </view>

        <!-- #ifdef H5 -->
        <view class="line-chart-wrap">
          <svg
            :viewBox="`0 0 ${SVG_W} ${SVG_H}`"
            preserveAspectRatio="xMidYMid meet"
            style="width:100%;height:auto;display:block;cursor:crosshair"
            @mousemove="onLineMouseMove"
            @mouseleave="onLineMouseLeave"
          >
            <defs>
              <linearGradient id="fillBlue" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stop-color="#3b82f6" stop-opacity="0.22"/>
                <stop offset="100%" stop-color="#3b82f6" stop-opacity="0"/>
              </linearGradient>
              <linearGradient id="fillGreen" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stop-color="#10b981" stop-opacity="0.18"/>
                <stop offset="100%" stop-color="#10b981" stop-opacity="0"/>
              </linearGradient>
              <clipPath id="chartClip">
                <rect :x="PL" :y="PT" :width="CW" :height="CH"/>
              </clipPath>
            </defs>

            <!-- Y 轴网格线 + 刻度标签 -->
            <g v-for="tick in yTicks" :key="tick">
              <line
                :x1="PL" :y1="py(tick)" :x2="PL + CW" :y2="py(tick)"
                stroke="currentColor" stroke-opacity="0.07" stroke-width="1"
              />
              <text
                :x="PL - 8" :y="py(tick) + 4"
                text-anchor="end" font-size="11" fill="currentColor" fill-opacity="0.45"
              >{{ fmtY(tick) }}</text>
            </g>
            <!-- Y 轴竖线 -->
            <line :x1="PL" :y1="PT" :x2="PL" :y2="PT + CH" stroke="currentColor" stroke-opacity="0.15" stroke-width="1"/>

            <!-- X 轴标签 -->
            <g v-for="lb in xLabels" :key="lb.i">
              <text
                :x="px(lb.i, trendData.length)" :y="PT + CH + 20"
                text-anchor="middle" font-size="10" fill="currentColor" fill-opacity="0.45"
              >{{ lb.label }}</text>
              <line
                :x1="px(lb.i, trendData.length)" :y1="PT + CH"
                :x2="px(lb.i, trendData.length)" :y2="PT + CH + 5"
                stroke="currentColor" stroke-opacity="0.2" stroke-width="1"
              />
            </g>
            <!-- X 轴横线 -->
            <line :x1="PL" :y1="PT + CH" :x2="PL + CW" :y2="PT + CH" stroke="currentColor" stroke-opacity="0.15" stroke-width="1"/>

            <!-- 面积图 -->
            <g clip-path="url(#chartClip)">
              <path :d="areaPath(trendData.map(d=>d.revenue), trendData.length)" fill="url(#fillGreen)"/>
              <path :d="areaPath(trendData.map(d=>d.users),   trendData.length)" fill="url(#fillBlue)"/>
              <!-- 折线 -->
              <path :d="smoothPath(trendData.map(d=>d.revenue), trendData.length)" fill="none" stroke="#10b981" stroke-width="2"/>
              <path :d="smoothPath(trendData.map(d=>d.users),   trendData.length)" fill="none" stroke="#3b82f6" stroke-width="2.5"/>
            </g>

            <!-- Hover 十字线 + 数据点 -->
            <g v-if="hoverIdx >= 0" style="pointer-events:none">
              <line
                :x1="hoverPx" :y1="PT" :x2="hoverPx" :y2="PT + CH"
                stroke="#94a3b8" stroke-width="1" stroke-dasharray="4 3"
              />
              <!-- users 点 -->
              <circle :cx="hoverPx" :cy="hoverPy1" r="5" fill="#3b82f6" stroke="#fff" stroke-width="2"/>
              <!-- revenue 点 -->
              <circle :cx="hoverPx" :cy="hoverPy2" r="5" fill="#10b981" stroke="#fff" stroke-width="2"/>
              <!-- X 轴十字 -->
              <line
                :x1="PL" :y1="hoverPy1" :x2="hoverPx" :y2="hoverPy1"
                stroke="#3b82f6" stroke-width="0.8" stroke-dasharray="3 3" opacity="0.4"
              />
              <line
                :x1="PL" :y1="hoverPy2" :x2="hoverPx" :y2="hoverPy2"
                stroke="#10b981" stroke-width="0.8" stroke-dasharray="3 3" opacity="0.4"
              />
            </g>
          </svg>

          <!-- Hover 浮动数据看板 -->
          <view
            v-if="hoverIdx >= 0 && trendData[hoverIdx]"
            class="line-tooltip"
            :style="tooltipStyle"
          >
            <view class="tt-date">📅 {{ trendData[hoverIdx].date }}</view>
            <view class="tt-row">
              <view class="tt-dot" style="background:#3b82f6"/>
              <text class="tt-lbl">用户增长</text>
              <text class="tt-val">{{ trendData[hoverIdx].users.toLocaleString() }}</text>
            </view>
            <view class="tt-row">
              <view class="tt-dot" style="background:#10b981"/>
              <text class="tt-lbl">流水</text>
              <text class="tt-val">¥{{ trendData[hoverIdx].revenue.toLocaleString() }}</text>
            </view>
          </view>
        </view>
        <!-- #endif -->

        <!-- #ifndef H5 -->
        <view style="height:200px;display:flex;align-items:center;justify-content:center;">
          <text style="color:var(--color-text-muted);font-size:12px;">图表仅支持 H5</text>
        </view>
        <!-- #endif -->

        <view class="chart-legend">
          <view class="legend-item"><view class="legend-line" style="background:#3b82f6"/><text>用户增长</text></view>
          <view class="legend-item"><view class="legend-line" style="background:#10b981"/><text>流水趋势</text></view>
          <text class="legend-hint">← 鼠标划过查看数据</text>
        </view>
      </view>

      <!-- ══ 圆环图：SVG 弧形 + Hover 看板 ══ -->
      <view class="card chart-card">
        <view class="card-header"><text class="card-title">联盟类型分布</text></view>

        <!-- #ifdef H5 -->
        <view class="donut-chart-wrap" @mousemove="onDonutMove">
          <svg viewBox="0 0 200 200" style="width:170px;height:170px;display:block;margin:0 auto;overflow:visible">
            <g v-for="seg in donutSegments" :key="seg.i"
               style="cursor:pointer;transition:opacity .15s"
               :style="{ opacity: donutHoverIdx >= 0 && donutHoverIdx !== seg.i ? 0.55 : 1 }"
               @mouseenter="(e) => onDonutEnter(e, seg.i)"
               @mouseleave="onDonutLeave">
              <path
                :d="arcPath(seg.start, seg.end, donutHoverIdx === seg.i)"
                :fill="seg.color"
                :filter="donutHoverIdx === seg.i ? 'drop-shadow(0 2px 6px rgba(0,0,0,0.25))' : ''"
              />
            </g>
            <!-- 中心文字 -->
            <text :x="CX" :y="CY - 8" text-anchor="middle" font-size="20" font-weight="700" fill="currentColor">
              {{ donutHoverIdx >= 0 ? orgTypes[donutHoverIdx]?.value : donutTotal }}
            </text>
            <text :x="CX" :y="CY + 10" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.5">
              {{ donutHoverIdx >= 0 ? orgTypes[donutHoverIdx]?.name : '联盟总数' }}
            </text>
          </svg>

          <!-- Hover 浮动看板 -->
          <view v-if="donutTooltip" class="donut-tooltip" :style="donutTipStyle">
            <view class="tt-date">{{ donutTooltip.name }}</view>
            <view class="tt-row">
              <text class="tt-lbl">数量</text>
              <text class="tt-val">{{ donutTooltip.value }}</text>
            </view>
            <view class="tt-row">
              <text class="tt-lbl">占比</text>
              <text class="tt-val">{{ donutTooltip.pct }}</text>
            </view>
          </view>

          <view class="donut-legend">
            <view
              v-for="(item, i) in orgTypes" :key="item.name"
              class="dl-item"
              :class="{ 'dl-active': donutHoverIdx === i }"
              @mouseenter="donutHoverIdx = i"
              @mouseleave="donutHoverIdx = -1"
            >
              <view class="dl-dot" :style="{ background: item.color }"/>
              <text class="dl-name">{{ item.name }}</text>
              <text class="dl-val">{{ item.value }}</text>
              <text class="dl-pct">{{ donutTotal > 0 ? (item.value / donutTotal * 100).toFixed(0) : 0 }}%</text>
            </view>
          </view>
        </view>
        <!-- #endif -->
      </view>
    </view>

    <!-- 底部行 -->
    <view class="bottom-row">
      <view class="card list-card">
        <view class="card-header">
          <text class="card-title">联盟流水排行 TOP 5</text>
          <text class="link" @click="uni.navigateTo({url:'/pages/orgs/index'})">查看全部 ›</text>
        </view>
        <view
          v-for="(item, idx) in orgRank" :key="item.name"
          class="rank-item"
          :class="{ 'rank-hovered': rankHoverIdx === idx }"
          @mouseenter="rankHoverIdx = idx"
          @mouseleave="rankHoverIdx = -1"
        >
          <view class="rank-num" :class="{ top: idx < 3 }">{{ idx + 1 }}</view>
          <text class="rank-name">{{ item.name }}</text>
          <view class="rank-bar-wrap">
            <view class="rank-bar" :style="{ width: (item.revenue / (orgRank[0]?.revenue || 1) * 100) + '%' }"/>
          </view>
          <!-- hover 时显示精确数值，否则显示 K 单位 -->
          <view class="rank-val-wrap">
            <text class="rank-val">
              {{ rankHoverIdx === idx ? '¥' + item.revenue.toLocaleString() : '¥' + (item.revenue >= 1000 ? (item.revenue/1000).toFixed(1)+'K' : item.revenue) }}
            </text>
          </view>
        </view>
      </view>

      <view class="card list-card">
        <view class="card-header">
          <view style="display:flex;align-items:center;gap:6px;">
            <view class="live-dot" />
            <text class="card-title">实时动态</text>
          </view>
          <text class="link" @click="uni.navigateTo({url:'/pages/messages/index'})">全部动态 ›</text>
        </view>
        <view v-for="ev in events" :key="ev.id" class="event-item">
          <view class="ev-dot" :style="{ background: ev.color }" />
          <!-- #ifdef H5 -->
          <text class="ev-text" v-html="ev.text" />
          <!-- #endif -->
          <!-- #ifndef H5 -->
          <text class="ev-text">{{ ev.text.replace(/<[^>]+>/g, '') }}</text>
          <!-- #endif -->
          <text class="ev-time">{{ formatTime(ev.ts) }}</text>
        </view>
      </view>
    </view>

  </AppLayout>
</template>

<style lang="scss" scoped>
.welcome-bar {
  background: linear-gradient(135deg, var(--color-primary-dark) 0%, var(--color-primary) 100%);
  border-radius: 14px; padding: 22px 28px; margin-bottom: 20px;
  display: flex; align-items: center; justify-content: space-between;
  overflow: hidden; position: relative;
  &::after {
    content: ''; position: absolute; right: -40px; top: -40px;
    width: 200px; height: 200px; border-radius: 50%; background: rgba(255,255,255,0.05);
  }
}
.wb-title {
  font-size: 20px; font-weight: 700; color: #fff;
  display: flex; align-items: center; gap: 8px; margin-bottom: 4px;
}
.wb-sub { font-size: 13px; color: rgba(255,255,255,0.7); }
.live-dot {
  width: 8px; height: 8px; border-radius: 50%; background: #34d399;
  flex-shrink: 0;
  animation: pulse 1.5s ease-in-out infinite;
}
@keyframes pulse { 0%,100%{opacity:1} 50%{opacity:.5} }
.wb-stats { display: flex; gap: 24px; z-index: 1; }
.wb-stat { text-align: center; }
.ws-num { font-size: 22px; font-weight: 700; color: #fff; display: block; }
.ws-lbl { font-size: 11px; color: rgba(255,255,255,0.65); display: block; margin-top: 2px; }
.ws-div { width: 1px; background: rgba(255,255,255,0.2); }

.kpi-row { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: 20px; }

.chart-row { display: grid; grid-template-columns: 2fr 1fr; gap: 16px; margin-bottom: 20px; }
.chart-card { padding: 20px; }
.card-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; }
.card-title { font-size: 14px; font-weight: 600; color: var(--color-text-primary); }
.card-tabs { display: flex; gap: 4px; }
.tab {
  font-size: 11px; padding: 3px 10px; border-radius: 6px; cursor: pointer;
  color: var(--color-text-secondary);
  &.active { background: var(--color-primary-light); color: var(--color-primary); font-weight: 600; }
}

/* ── 折线图 ── */
.line-chart-wrap {
  width: 100%; position: relative; overflow: visible;
  border-radius: 6px;
  svg { color: var(--color-text-primary); }
}
.chart-legend { display: flex; align-items: center; gap: 20px; margin-top: 6px; }
.legend-item { display: flex; align-items: center; gap: 5px; font-size: 11px; color: var(--color-text-secondary); }
.legend-line { width: 14px; height: 2.5px; border-radius: 2px; }
.legend-hint { font-size: 10px; color: var(--color-text-muted); margin-left: auto; }

/* 折线图 Tooltip */
.line-tooltip {
  position: absolute; pointer-events: none; z-index: 20;
  background: var(--color-card-bg);
  border: 1px solid var(--color-border);
  border-radius: 10px; padding: 10px 14px;
  box-shadow: 0 8px 24px rgba(0,0,0,0.12);
  min-width: 150px;
}
.tt-date { font-size: 11px; font-weight: 600; color: var(--color-text-secondary); margin-bottom: 8px; }
.tt-row { display: flex; align-items: center; gap: 7px; margin-bottom: 4px; &:last-child { margin: 0; } }
.tt-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.tt-lbl { font-size: 11px; color: var(--color-text-muted); flex: 1; }
.tt-val { font-size: 12px; font-weight: 700; color: var(--color-text-primary); }

/* ── 圆环图 ── */
.donut-chart-wrap {
  position: relative; display: flex; flex-direction: column; align-items: center;
  svg { overflow: visible; }
}
.donut-tooltip {
  position: absolute; pointer-events: none; z-index: 20;
  background: var(--color-card-bg);
  border: 1px solid var(--color-border);
  border-radius: 10px; padding: 10px 14px;
  box-shadow: 0 8px 24px rgba(0,0,0,0.14);
  min-width: 110px;
}
.donut-legend { width: 100%; margin-top: 10px; display: flex; flex-direction: column; gap: 5px; }
.dl-item {
  display: flex; align-items: center; gap: 8px; font-size: 11px;
  padding: 4px 6px; border-radius: 6px; cursor: pointer;
  transition: background 0.15s;
  &:hover, &.dl-active { background: var(--color-border-light); }
}
.dl-dot { width: 8px; height: 8px; border-radius: 2px; flex-shrink: 0; }
.dl-name { flex: 1; color: var(--color-text-secondary); }
.dl-val { font-weight: 600; color: var(--color-text-primary); min-width: 20px; text-align: right; }
.dl-pct { color: var(--color-text-muted); min-width: 30px; text-align: right; }

/* ── 排行榜 ── */
.bottom-row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
.list-card { padding: 20px; }
.link { font-size: 12px; color: var(--color-primary); cursor: pointer; }
.rank-item {
  display: flex; align-items: center; gap: 12px; padding: 9px 0;
  border-bottom: 1px solid var(--color-border-light); border-radius: 6px;
  padding-left: 4px; padding-right: 4px; transition: background 0.1s;
  &:last-child { border: none; }
  &.rank-hovered { background: var(--color-border-light); }
}
.rank-num {
  width: 20px; height: 20px; border-radius: 5px;
  background: var(--color-border-light); font-size: 10px; font-weight: 700;
  display: flex; align-items: center; justify-content: center;
  color: var(--color-text-secondary); flex-shrink: 0;
  &.top { background: var(--color-primary); color: #fff; }
}
.rank-name { flex: 1; font-size: 13px; color: var(--color-text-primary); }
.rank-bar-wrap { width: 72px; height: 6px; background: var(--color-border-light); border-radius: 3px; flex-shrink: 0; }
.rank-bar { height: 6px; border-radius: 3px; background: linear-gradient(to right, var(--color-primary-dark), var(--color-primary)); transition: width 0.3s; }
.rank-val-wrap { width: 72px; text-align: right; flex-shrink: 0; }
.rank-val { font-size: 12px; font-weight: 600; color: var(--color-text-primary); }

/* ── 实时动态 ── */
.event-item {
  display: flex; align-items: flex-start; gap: 10px; padding: 8px 0;
  border-bottom: 1px solid var(--color-border-light);
  &:last-child { border: none; }
}
.ev-dot { width: 8px; height: 8px; border-radius: 50%; margin-top: 4px; flex-shrink: 0; }
.ev-text { flex: 1; font-size: 12px; color: var(--color-text-primary); line-height: 1.5; }
.ev-time { font-size: 11px; color: var(--color-text-muted); flex-shrink: 0; white-space: nowrap; }
</style>
