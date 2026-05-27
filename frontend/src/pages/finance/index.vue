<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import KpiCard from '@/components/common/KpiCard.vue'
import FilterPanel from '@/components/common/FilterPanel.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import Pagination from '@/components/common/Pagination.vue'
import { getFinanceList, settleFinance } from '@/api/finance'
import type { FilterField, QuickTag } from '@/components/common/FilterPanel.vue'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()

const breadcrumbs = [{ label: '首页' }, { label: '财务 & 报表' }, { label: '财务结算' }]
const pageStats = reactive({ total: 0, done: 0, pending: 0, processing: 0, totalAmount: 0 })

const filterFields: FilterField[] = [
  { key: 'keyword',     label: '联盟名称',        type: 'input' },
  { key: 'status',      label: '结算状态',        type: 'select', options: [{ label:'全部',value:'' },{ label:'待结算',value:'1' },{ label:'结算中',value:'2' },{ label:'已结算',value:'3' }] },
  { key: 'period',      label: '结算周期',        type: 'select', options: [{ label:'全部',value:'' },{ label:'日结',value:'daily' },{ label:'周结',value:'weekly' },{ label:'月结',value:'monthly' }] },
  { key: 'minAmount',   label: '最小金额（元）',  type: 'input', placeholder: '0' },
  { key: 'maxAmount',   label: '最大金额（元）',  type: 'input', placeholder: '999999' },
  { key: 'startDate',   label: '结算时间（起）',  type: 'input', placeholder: 'YYYY-MM-DD' },
  { key: 'endDate',     label: '结算时间（止）',  type: 'input', placeholder: 'YYYY-MM-DD' },
]
const quickTags: QuickTag[] = [
  { key: 'pending',    label: '待结算',   color: '#f59e0b', params: { status: '1' } },
  { key: 'processing', label: '结算中',   color: '#3b82f6', params: { status: '2' } },
  { key: 'this_month', label: '本月结算', color: '#10b981', params: { startDate: new Date().toISOString().slice(0,7) + '-01' } },
]

const statusCfg = {
  1: { label: '待结算', s: 'warning'  as const },
  2: { label: '结算中', s: 'info'     as const },
  3: { label: '已结算', s: 'success'  as const },
  4: { label: '失败',   s: 'danger'   as const },
}
const periodLabel: Record<string, string> = { daily: '日结', weekly: '周结', monthly: '月结' }

interface FinanceRow {
  id: string; orgName: string; amount: number; _status: number
  period: string; settledAt: string; createdAt: string
}

const list = ref<FinanceRow[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const filterParams = ref<Record<string, unknown>>({})

function normalize(raw: any): FinanceRow {
  return {
    id: String(raw.id),
    orgName: raw.org?.name || '',
    amount: raw.amount ?? 0,
    _status: raw.status,
    period: periodLabel[raw.period] || raw.period || '',
    settledAt: raw.settled_at?.slice(0, 10) || '',
    createdAt: raw.created_at?.slice(0, 10) || '',
  }
}

async function loadList() {
  loading.value = true
  try {
    const res: any = await getFinanceList({ ...filterParams.value as any, page: page.value, pageSize: pageSize.value })
    list.value = (res.list || []).map(normalize)
    total.value = res.total ?? 0
    pageStats.total       = res.total ?? 0
    pageStats.done        = list.value.filter(r => r._status === 3).length
    pageStats.pending     = list.value.filter(r => r._status === 1).length
    pageStats.processing  = list.value.filter(r => r._status === 2).length
    pageStats.totalAmount = list.value.reduce((s, r) => s + r.amount, 0)
  } catch (e: any) {
    uni.showToast({ title: e?.message || '加载失败', icon: 'none' })
  } finally {
    loading.value = false
  }
}

async function handleSettle(id: string) {
  try {
    await settleFinance(id)
    uni.showToast({ title: '结算成功', icon: 'success' })
    loadList()
  } catch (e: any) {
    uni.showToast({ title: e?.message || '结算失败', icon: 'none' })
  }
}

function onSearch(params: Record<string, unknown>) { filterParams.value = params; page.value = 1; loadList() }
function onPageChange(p: number) { page.value = p; loadList() }
function onPageSizeChange(s: number) { pageSize.value = s; page.value = 1; loadList() }

// ══ 结算详情 Modal ══
const showDetailModal = ref(false)
const detailRow = ref<FinanceRow | null>(null)
function openDetail(row: FinanceRow) { detailRow.value = row; showDetailModal.value = true }

onMounted(loadList)
</script>

<template>
  <AppLayout :breadcrumbs="breadcrumbs">
    <view class="stats-row">
      <KpiCard icon="📋" label="结算总笔数"   :value="pageStats.total"       :trend="{ dir:'up', text:'12%' }"    icon-bg="#eff6ff" />
      <KpiCard icon="✅" label="已结算"       :value="pageStats.done"        :trend="{ dir:'up', text:'88.1%' }"  icon-bg="#f0fdf4" />
      <KpiCard icon="⏳" label="待结算"       :value="pageStats.pending"     :trend="{ dir:'up', text:'+23今日' }" icon-bg="#fffbeb" />
      <KpiCard icon="🔄" label="结算中"       :value="pageStats.processing"  :trend="{ dir:'down', text:'处理中' }" icon-bg="#eff6ff" />
      <KpiCard icon="💰" label="总结算金额(元)" :value="pageStats.totalAmount":trend="{ dir:'up', text:'本月' }"   icon-bg="#f0fdf4" />
    </view>
    <FilterPanel :fields="filterFields" :quick-tags="quickTags"
      @search="onSearch" @reset="() => { filterParams = {}; loadList() }" @export="()=>{}" />
    <view class="card table-card">
      <view class="table-toolbar">
        <text class="sel-info">共 <text class="em">{{ total }}</text> 条结算记录</text>
        <view v-if="loading" class="loading-tip">加载中…</view>
      </view>
      <view class="t-head">
        <text class="th" style="flex:2">联盟名称</text>
        <text class="th" style="flex:1">结算金额</text>
        <text class="th" style="flex:0.9">状态</text>
        <text class="th" style="flex:0.8">周期</text>
        <text class="th" style="flex:1.1">结算时间</text>
        <text class="th" style="flex:1">操作</text>
      </view>
      <view v-if="!list.length && !loading" class="empty-tip">暂无结算记录</view>
      <view v-for="row in list" :key="row.id" class="t-row">
        <text class="td" style="flex:2;font-weight:500">{{ row.orgName }}</text>
        <text class="td amount" style="flex:1">¥{{ row.amount.toLocaleString() }}</text>
        <view class="td" style="flex:0.9">
          <StatusBadge
            :status="statusCfg[row._status as keyof typeof statusCfg]?.s || 'info'"
            :label="statusCfg[row._status as keyof typeof statusCfg]?.label || '未知'" />
        </view>
        <text class="td t-muted" style="flex:0.8">{{ row.period }}</text>
        <text class="td t-muted" style="flex:1.1">{{ row.settledAt || '—' }}</text>
        <view class="td action-btns" style="flex:1">
          <view class="act-btn act-view" @click="openDetail(row)">详情</view>
          <view v-if="row._status === 1 && userStore.hasPermission('finance:settle')" class="act-btn act-edit" @click="handleSettle(row.id)">结算</view>
        </view>
      </view>
      <Pagination :total="total" :page="page" :page-size="pageSize"
        @page-change="onPageChange" @page-size-change="onPageSizeChange" />
    </view>

    <!-- 结算详情 Modal -->
    <!-- #ifdef H5 -->
    <view v-if="showDetailModal && detailRow" class="modal-mask" @click.self="showDetailModal = false">
      <view class="modal-box">
        <view class="modal-header">
          <text class="modal-title">结算详情</text>
          <text class="modal-close" @click="showDetailModal = false">✕</text>
        </view>
        <view class="modal-body">
          <view class="settle-amount-card">
            <text class="sa-label">结算金额</text>
            <text class="sa-amount">¥{{ detailRow.amount.toLocaleString() }}</text>
          </view>
          <view class="detail-grid">
            <view class="dg-item"><text class="dg-label">结算ID</text><text class="dg-val">{{ detailRow.id }}</text></view>
            <view class="dg-item"><text class="dg-label">所属联盟</text><text class="dg-val">{{ detailRow.orgName || '—' }}</text></view>
            <view class="dg-item">
              <text class="dg-label">结算状态</text>
              <StatusBadge :status="statusCfg[detailRow._status as keyof typeof statusCfg]?.s || 'info'"
                :label="statusCfg[detailRow._status as keyof typeof statusCfg]?.label || '未知'" />
            </view>
            <view class="dg-item"><text class="dg-label">结算周期</text><text class="dg-val">{{ detailRow.period }}</text></view>
            <view class="dg-item"><text class="dg-label">创建时间</text><text class="dg-val">{{ detailRow.createdAt }}</text></view>
            <view class="dg-item"><text class="dg-label">结算时间</text><text class="dg-val">{{ detailRow.settledAt || '—' }}</text></view>
          </view>
        </view>
        <view class="modal-footer">
          <view class="m-btn m-btn-cancel" @click="showDetailModal = false">关闭</view>
          <view v-if="detailRow._status === 1 && userStore.hasPermission('finance:settle')" class="m-btn m-btn-primary"
            @click="() => { showDetailModal = false; handleSettle(detailRow!.id) }">立即结算</view>
        </view>
      </view>
    </view>
    <!-- #endif -->
  </AppLayout>
</template>

<style lang="scss" scoped>
.stats-row { display:grid; grid-template-columns:repeat(5,1fr); gap:12px; margin-bottom:16px; }
.table-card { overflow:hidden; }
.table-toolbar { display:flex; padding:14px 16px; }
.sel-info { font-size:13px; color:var(--color-text-secondary); }
.em { color:var(--color-primary); font-weight:600; }
.t-head { display:flex; padding:10px 16px; background:var(--color-border-light); border-bottom:1px solid var(--color-border); }
.th { font-size:12px; font-weight:600; color:var(--color-text-secondary); padding-right:8px; }
.t-row { display:flex; align-items:center; padding:11px 16px; border-bottom:1px solid var(--color-border-light); &:hover{background:var(--color-border-light);} }
.td { font-size:13px; color:var(--color-text-primary); padding-right:8px; }
.t-muted { font-size:12px; color:var(--color-text-muted); }
.amount { font-weight:700; color:#16a34a; font-size:14px; }
.action-btns { display:flex; gap:6px; }
.act-btn { font-size:12px; padding:4px 10px; border-radius:5px; cursor:pointer; font-weight:500; }
.act-view { background:var(--color-primary-light); color:var(--color-primary); }
.act-edit { background:#f0fdf4; color:#16a34a; }
.empty-tip { text-align:center; padding:40px; color:var(--color-text-muted); font-size:13px; }
.loading-tip { font-size:12px; color:var(--color-text-muted); }
.modal-mask { position:fixed; inset:0; background:rgba(0,0,0,0.45); z-index:300; display:flex; align-items:center; justify-content:center; }
.modal-box { background:var(--color-card-bg); border-radius:16px; width:460px; max-width:92vw; box-shadow:0 20px 60px rgba(0,0,0,0.2); overflow:hidden; }
.modal-header { display:flex; align-items:center; justify-content:space-between; padding:18px 24px; border-bottom:1px solid var(--color-border-light); }
.modal-title { font-size:16px; font-weight:700; color:var(--color-text-primary); }
.modal-close { font-size:18px; color:var(--color-text-muted); cursor:pointer; }
.modal-body { padding:20px 24px; }
.modal-footer { display:flex; justify-content:flex-end; gap:10px; padding:16px 24px; border-top:1px solid var(--color-border-light); }
.m-btn { height:36px; border-radius:8px; font-size:13px; font-weight:500; display:flex; align-items:center; padding:0 20px; cursor:pointer; }
.m-btn-cancel  { background:var(--color-border-light); color:var(--color-text-secondary); }
.m-btn-primary { background:var(--color-primary); color:#fff; }
.settle-amount-card { background:linear-gradient(135deg,var(--color-primary-dark),var(--color-primary)); border-radius:12px; padding:20px; text-align:center; margin-bottom:20px; }
.sa-label  { font-size:12px; color:rgba(255,255,255,0.8); display:block; margin-bottom:4px; }
.sa-amount { font-size:28px; font-weight:800; color:#fff; display:block; }
.detail-grid { display:grid; grid-template-columns:1fr 1fr; gap:14px; }
.dg-item { display:flex; flex-direction:column; gap:4px; }
.dg-label { font-size:11px; color:var(--color-text-muted); }
.dg-val { font-size:13px; font-weight:500; color:var(--color-text-primary); }
</style>
