<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import KpiCard from '@/components/common/KpiCard.vue'
import FilterPanel from '@/components/common/FilterPanel.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import Pagination from '@/components/common/Pagination.vue'
import { getOrderList, refundOrder, type OrderItem } from '@/api/order'
import type { FilterField, QuickTag } from '@/components/common/FilterPanel.vue'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()

const breadcrumbs = [{ label: '首页' }, { label: '核心业务' }, { label: '订单中心' }]
const pageStats = reactive({ total: 0, paid: 0, pending: 0, refunded: 0, todayNew: 0 })

const filterFields: FilterField[] = [
  { key: 'keyword',   label: '订单号 / 用户',  type: 'input' },
  { key: 'type',      label: '订单类型',       type: 'select', options: [{ label:'全部',value:'' },{ label:'普通订单',value:'normal' },{ label:'退款单',value:'refund' }] },
  { key: 'status',    label: '支付状态',       type: 'select', options: [{ label:'全部',value:'' },{ label:'待支付',value:'1' },{ label:'已支付',value:'2' },{ label:'已退款',value:'3' }] },
  { key: 'payMethod', label: '支付方式',       type: 'select', options: [{ label:'全部',value:'' },{ label:'微信支付',value:'wx' },{ label:'支付宝',value:'ali' },{ label:'银行卡',value:'bank' }] },
  { key: 'minAmount', label: '最小金额（元）', type: 'input', placeholder: '0' },
  { key: 'maxAmount', label: '最大金额（元）', type: 'input', placeholder: '99999' },
  { key: 'startDate', label: '下单时间（起）', type: 'input', placeholder: 'YYYY-MM-DD' },
  { key: 'endDate',   label: '下单时间（止）', type: 'input', placeholder: 'YYYY-MM-DD' },
]
const quickTags: QuickTag[] = [
  { key: 'pending',  label: '待支付',  color: '#f59e0b', params: { status: '1' } },
  { key: 'today',    label: '今日订单', color: '#3b82f6', params: { startDate: new Date().toISOString().slice(0,10) } },
  { key: 'refunded', label: '退款订单', color: '#ef4444', params: { status: '3' } },
]

const statusCfg = {
  1: { label: '待支付', status: 'warning' as const },
  2: { label: '已支付', status: 'success' as const },
  3: { label: '已退款', status: 'danger'  as const },
  4: { label: '已取消', status: 'info'    as const },
}
const payMethodLabel: Record<string, string> = { wx: '微信支付', ali: '支付宝', bank: '银行卡' }
const typeLabel: Record<string, string> = { normal: '普通订单', refund: '退款单' }

function normalizeOrder(raw: any): OrderItem & { _status: number } {
  return {
    id: String(raw.id),
    orderNo: raw.order_no,
    type: typeLabel[raw.type] || raw.type,
    status: raw.status === 2 ? 'paid' : raw.status === 3 ? 'refunded' : raw.status === 4 ? 'cancelled' : 'pending',
    payMethod: payMethodLabel[raw.pay_method] || raw.pay_method || '',
    amount: raw.amount ?? 0,
    userName: raw.user?.username || '',
    orgName: raw.org?.name || '',
    createdAt: raw.created_at?.slice(0, 16) || '',
    _status: raw.status,
  }
}

const list = ref<(OrderItem & { _status: number })[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const filterParams = ref<Record<string, unknown>>({})

async function loadList() {
  loading.value = true
  try {
    const res: any = await getOrderList({ ...filterParams.value as any, page: page.value, pageSize: pageSize.value })
    list.value = (res.list || []).map(normalizeOrder)
    total.value = res.total ?? 0
    pageStats.total    = res.total ?? 0
    pageStats.paid     = list.value.filter(o => o._status === 2).length
    pageStats.pending  = list.value.filter(o => o._status === 1).length
    pageStats.refunded = list.value.filter(o => o._status === 3).length
    pageStats.todayNew = list.value.filter(o => o.createdAt.startsWith(new Date().toISOString().slice(0,10))).length
  } catch (e: any) {
    uni.showToast({ title: e?.message || '加载失败', icon: 'none' })
  } finally {
    loading.value = false
  }
}

async function handleRefund(id: string) {
  uni.showModal({ title: '确认退款', content: '确认对该订单发起退款？此操作不可逆。', success: async (res) => {
    if (!res.confirm) return
    try {
      await refundOrder(id, '管理员手动退款')
      uni.showToast({ title: '退款成功', icon: 'success' })
      loadList()
    } catch (e: any) {
      uni.showToast({ title: e?.message || '退款失败', icon: 'none' })
    }
  }})
}

function onSearch(params: Record<string, unknown>) { filterParams.value = params; page.value = 1; loadList() }
function onPageChange(p: number) { page.value = p; loadList() }
function onPageSizeChange(s: number) { pageSize.value = s; page.value = 1; loadList() }

// ══ 订单详情 Modal ══
const showDetailModal = ref(false)
const detailRow = ref<any>(null)
const statusLabels: Record<number, string> = { 1: '待支付', 2: '已支付', 3: '已退款', 4: '已取消' }
const statusColors: Record<number, string> = { 1: '#f59e0b', 2: '#10b981', 3: '#ef4444', 4: '#94a3b8' }

function openDetail(row: any) { detailRow.value = row; showDetailModal.value = true }

onMounted(loadList)
</script>

<template>
  <AppLayout :breadcrumbs="breadcrumbs">
    <view class="stats-row">
      <KpiCard icon="📦" label="订单总数"   :value="pageStats.total"   :trend="{ dir:'up', text:'8.7%' }"   icon-bg="#eff6ff" />
      <KpiCard icon="✅" label="已支付"     :value="pageStats.paid"    :trend="{ dir:'up', text:'本月' }"    icon-bg="#f0fdf4" />
      <KpiCard icon="⏳" label="待支付"     :value="pageStats.pending" :trend="{ dir:'up', text:'今日+12' }" icon-bg="#fffbeb" />
      <KpiCard icon="↩️" label="已退款"     :value="pageStats.refunded":trend="{ dir:'down', text:'较昨' }"  icon-bg="#fef2f2" />
      <KpiCard icon="🆕" label="今日新增"   :value="pageStats.todayNew":trend="{ dir:'up', text:'+86' }"    icon-bg="#faf5ff" />
    </view>
    <FilterPanel :fields="filterFields" :quick-tags="quickTags"
      @search="onSearch" @reset="() => { filterParams = {}; loadList() }" @export="()=>{}" />
    <view class="card table-card">
      <view class="table-toolbar">
        <text class="sel-info">共 <text class="em">{{ total }}</text> 条订单</text>
        <view v-if="loading" class="loading-tip">加载中…</view>
      </view>
      <view class="t-head">
        <text class="th" style="flex:2">订单号</text>
        <text class="th" style="flex:0.8">类型</text>
        <text class="th" style="flex:0.9">状态</text>
        <text class="th" style="flex:0.8">支付方式</text>
        <text class="th" style="flex:0.8">金额</text>
        <text class="th" style="flex:1">用户</text>
        <text class="th" style="flex:1.2">所属联盟</text>
        <text class="th" style="flex:1.2">下单时间</text>
        <text class="th" style="flex:1">操作</text>
      </view>
      <view v-if="!list.length && !loading" class="empty-tip">暂无订单数据</view>
      <view v-for="row in list" :key="row.id" class="t-row">
        <text class="td order-no" style="flex:2">{{ row.orderNo }}</text>
        <text class="td t-muted" style="flex:0.8">{{ row.type }}</text>
        <view class="td" style="flex:0.9">
          <StatusBadge
            :status="statusCfg[row._status as keyof typeof statusCfg]?.status || 'info'"
            :label="statusCfg[row._status as keyof typeof statusCfg]?.label || '未知'" />
        </view>
        <text class="td t-muted" style="flex:0.8">{{ row.payMethod }}</text>
        <text class="td amount" style="flex:0.8">¥{{ row.amount.toLocaleString() }}</text>
        <text class="td t-muted" style="flex:1">{{ row.userName }}</text>
        <text class="td" style="flex:1.2"><text class="org-tag">{{ row.orgName }}</text></text>
        <text class="td t-muted" style="flex:1.2">{{ row.createdAt }}</text>
        <view class="td action-btns" style="flex:1">
          <view class="act-btn act-view" @click="openDetail(row)">详情</view>
          <view v-if="row._status === 2 && userStore.hasPermission('order:refund')" class="act-btn act-edit" @click="handleRefund(row.id)">退款</view>
        </view>
      </view>
      <Pagination :total="total" :page="page" :page-size="pageSize"
        @page-change="onPageChange" @page-size-change="onPageSizeChange" />
    </view>

    <!-- 订单详情 Modal -->
    <!-- #ifdef H5 -->
    <view v-if="showDetailModal && detailRow" class="modal-mask" @click.self="showDetailModal = false">
      <view class="modal-box">
        <view class="modal-header">
          <text class="modal-title">订单详情</text>
          <text class="modal-close" @click="showDetailModal = false">✕</text>
        </view>
        <view class="modal-body">
          <view class="order-status-bar" :style="{ borderLeftColor: statusColors[detailRow._status] }">
            <text class="os-label">{{ statusLabels[detailRow._status] || '未知' }}</text>
            <text class="os-time">{{ detailRow.createdAt }}</text>
          </view>
          <view class="detail-grid">
            <view class="dg-item"><text class="dg-label">订单号</text><text class="dg-val mono">{{ detailRow.orderNo }}</text></view>
            <view class="dg-item"><text class="dg-label">订单类型</text><text class="dg-val">{{ detailRow.type }}</text></view>
            <view class="dg-item"><text class="dg-label">支付方式</text><text class="dg-val">{{ detailRow.payMethod || '—' }}</text></view>
            <view class="dg-item"><text class="dg-label">订单金额</text><text class="dg-val amount">¥{{ detailRow.amount.toLocaleString() }}</text></view>
            <view class="dg-item"><text class="dg-label">用户</text><text class="dg-val">{{ detailRow.userName || '—' }}</text></view>
            <view class="dg-item"><text class="dg-label">所属联盟</text><text class="dg-val">{{ detailRow.orgName || '—' }}</text></view>
            <view class="dg-item"><text class="dg-label">下单时间</text><text class="dg-val">{{ detailRow.createdAt }}</text></view>
          </view>
        </view>
        <view class="modal-footer">
          <view class="m-btn m-btn-cancel" @click="showDetailModal = false">关闭</view>
          <view v-if="detailRow._status === 2" class="m-btn m-btn-danger" @click="() => { showDetailModal = false; handleRefund(detailRow.id) }">发起退款</view>
        </view>
      </view>
    </view>
    <!-- #endif -->
  </AppLayout>
</template>

<style lang="scss" scoped>
.stats-row { display: grid; grid-template-columns: repeat(5,1fr); gap:12px; margin-bottom:16px; }
.table-card { overflow: hidden; }
.table-toolbar { display:flex; align-items:center; justify-content:space-between; padding:14px 16px; }
.sel-info { font-size:13px; color:var(--color-text-secondary); }
.em { color:var(--color-primary); font-weight:600; }
.t-head { display:flex; padding:10px 16px; background:var(--color-border-light); border-bottom:1px solid var(--color-border); }
.th { font-size:12px; font-weight:600; color:var(--color-text-secondary); padding-right:8px; }
.t-row { display:flex; align-items:center; padding:11px 16px; border-bottom:1px solid var(--color-border-light); &:hover{background:var(--color-border-light);} }
.td { font-size:13px; color:var(--color-text-primary); padding-right:8px; }
.t-muted { font-size:12px; color:var(--color-text-muted); }
.order-no { font-family: monospace; font-size:12px; color:var(--color-primary); }
.amount { font-weight:600; color:#16a34a; }
.org-tag { background:var(--color-border-light); padding:2px 8px; border-radius:4px; font-size:11px; }
.action-btns { display:flex; gap:6px; }
.act-btn { font-size:12px; padding:4px 10px; border-radius:5px; cursor:pointer; font-weight:500; }
.act-view { background:var(--color-primary-light); color:var(--color-primary); }
.act-edit { background:#fef2f2; color:#dc2626; }
.empty-tip { text-align:center; padding:40px; color:var(--color-text-muted); font-size:13px; }
.loading-tip { font-size:12px; color:var(--color-text-muted); }
/* Modal */
.modal-mask { position:fixed; inset:0; background:rgba(0,0,0,0.45); z-index:300; display:flex; align-items:center; justify-content:center; }
.modal-box { background:var(--color-card-bg); border-radius:16px; width:480px; max-width:92vw; box-shadow:0 20px 60px rgba(0,0,0,0.2); overflow:hidden; }
.modal-header { display:flex; align-items:center; justify-content:space-between; padding:18px 24px; border-bottom:1px solid var(--color-border-light); }
.modal-title { font-size:16px; font-weight:700; color:var(--color-text-primary); }
.modal-close { font-size:18px; color:var(--color-text-muted); cursor:pointer; }
.modal-body { padding:20px 24px; max-height:60vh; overflow-y:auto; }
.modal-footer { display:flex; justify-content:flex-end; gap:10px; padding:16px 24px; border-top:1px solid var(--color-border-light); }
.m-btn { height:36px; border-radius:8px; font-size:13px; font-weight:500; display:flex; align-items:center; padding:0 20px; cursor:pointer; }
.m-btn-cancel { background:var(--color-border-light); color:var(--color-text-secondary); }
.m-btn-danger  { background:#fef2f2; color:#dc2626; border:1px solid #fecaca; }
.order-status-bar { border-left:3px solid #94a3b8; padding:8px 14px; background:var(--color-border-light); border-radius:0 8px 8px 0; margin-bottom:20px; display:flex; align-items:center; justify-content:space-between; }
.os-label { font-size:14px; font-weight:700; color:var(--color-text-primary); }
.os-time  { font-size:12px; color:var(--color-text-muted); }
.detail-grid { display:grid; grid-template-columns:1fr 1fr; gap:14px; }
.dg-item { display:flex; flex-direction:column; gap:4px; }
.dg-label { font-size:11px; color:var(--color-text-muted); }
.dg-val { font-size:13px; font-weight:500; color:var(--color-text-primary); }
.dg-val.mono { font-family:monospace; font-size:12px; color:var(--color-primary); }
.dg-val.amount { color:#16a34a; font-weight:700; }
</style>
