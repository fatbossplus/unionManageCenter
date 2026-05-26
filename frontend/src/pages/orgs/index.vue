<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import KpiCard from '@/components/common/KpiCard.vue'
import FilterPanel from '@/components/common/FilterPanel.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import Pagination from '@/components/common/Pagination.vue'
import { getOrgList, type OrgItem } from '@/api/org'
import type { FilterField, QuickTag } from '@/components/common/FilterPanel.vue'

const breadcrumbs = [{ label: '首页' }, { label: '核心业务' }, { label: '联盟管理' }]
const pageStats = reactive({ total: 0, active: 0, pending: 0, frozen: 0, todayNew: 0 })

const filterFields: FilterField[] = [
  { key: 'keyword',   label: '关键词',    type: 'input',  placeholder: '联盟名称 / 负责人' },
  { key: 'type',      label: '联盟类型',  type: 'select', options: [
    { label: '全部类型', value: '' }, { label: '电商联盟', value: 'ec' },
    { label: '服务联盟', value: 'service' }, { label: '内容联盟', value: 'content' }, { label: '其他', value: 'other' },
  ]},
  { key: 'status',    label: '审核状态',  type: 'select', options: [
    { label: '全部', value: '' }, { label: '正常', value: 'active' },
    { label: '待审核', value: 'pending' }, { label: '已冻结', value: 'frozen' },
  ]},
  { key: 'region',    label: '所在地区',  type: 'input',  placeholder: '省/市' },
  { key: 'startDate', label: '成立时间（起）', type: 'input', placeholder: 'YYYY-MM-DD' },
  { key: 'endDate',   label: '成立时间（止）', type: 'input', placeholder: 'YYYY-MM-DD' },
]

const quickTags: QuickTag[] = [
  { key: 'pending', label: '待审核', color: '#f59e0b', params: { status: 'pending' } },
  { key: 'active',  label: '正常运营', color: '#10b981', params: { status: 'active' } },
  { key: 'frozen',  label: '已冻结',   color: '#ef4444', params: { status: 'frozen' } },
  { key: 'new',     label: '本月新增', color: '#3b82f6', params: { startDate: new Date().toISOString().slice(0, 7) + '-01' } },
]

const list = ref<OrgItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const filterParams = ref<Record<string, unknown>>({})

async function loadList() {
  loading.value = true
  try {
    const res = await getOrgList({ ...filterParams.value as any, page: page.value, pageSize: pageSize.value })
    list.value = res.list
    total.value = res.total
    pageStats.total = res.total
    pageStats.active  = res.list.filter(o => o.status === 'active').length
    pageStats.pending = res.list.filter(o => o.status === 'pending').length
    pageStats.frozen  = res.list.filter(o => o.status === 'frozen').length
  } catch (e: any) {
    uni.showToast({ title: e?.message || '加载失败', icon: 'none' })
  } finally {
    loading.value = false
  }
}

function onSearch(params: Record<string, unknown>) {
  filterParams.value = params; page.value = 1; loadList()
}
function onPageChange(p: number) { page.value = p; loadList() }
function onPageSizeChange(s: number) { pageSize.value = s; page.value = 1; loadList() }

onMounted(loadList)
</script>

<template>
  <AppLayout :breadcrumbs="breadcrumbs">
    <view class="stats-row">
      <KpiCard icon="🏢" label="联盟总数"   :value="pageStats.total"   :trend="{ dir:'up', text:'5.1%' }"  icon-bg="#eff6ff" />
      <KpiCard icon="✅" label="正常运营"   :value="pageStats.active"  :trend="{ dir:'up', text:'本月+6' }" icon-bg="#f0fdf4" />
      <KpiCard icon="⏳" label="待审核"     :value="pageStats.pending" :trend="{ dir:'up', text:'今日+3' }" icon-bg="#fffbeb" />
      <KpiCard icon="❄️" label="已冻结"     :value="pageStats.frozen"  :trend="{ dir:'down', text:'无变化' }" icon-bg="#fef2f2" />
      <KpiCard icon="🆕" label="今日新增"   :value="pageStats.todayNew" :trend="{ dir:'up', text:'+3' }"   icon-bg="#faf5ff" />
    </view>
    <FilterPanel :fields="filterFields" :quick-tags="quickTags"
      @search="onSearch" @reset="() => { filterParams = {}; loadList() }" @export="() => {}" />
    <view class="card table-card">
      <view class="table-toolbar">
        <text class="sel-info">共 <text class="em">{{ total }}</text> 条联盟</text>
        <view class="t-btn t-btn-primary" @click="uni.navigateTo({ url: '/pages/orgs/detail?mode=create' })">＋ 新增联盟</view>
      </view>
      <view class="t-head">
        <text class="th" style="flex:2">联盟名称</text>
        <text class="th" style="flex:1">类型</text>
        <text class="th" style="flex:1">状态</text>
        <text class="th" style="flex:1">地区</text>
        <text class="th" style="flex:0.8">成员数</text>
        <text class="th" style="flex:0.9">负责人</text>
        <text class="th" style="flex:1">成立时间</text>
        <text class="th" style="flex:1.2">操作</text>
      </view>
      <view v-for="row in list" :key="row.id" class="t-row">
        <text class="td org-name" style="flex:2">{{ row.name }}</text>
        <text class="td" style="flex:1"><text class="org-tag">{{ row.type }}</text></text>
        <view class="td" style="flex:1">
          <StatusBadge :status="row.status === 'active' ? 'success' : row.status === 'pending' ? 'warning' : 'danger'"
            :label="row.status === 'active' ? '正常' : row.status === 'pending' ? '待审核' : '已冻结'" />
        </view>
        <text class="td t-muted" style="flex:1">{{ row.region }}</text>
        <text class="td" style="flex:0.8">{{ row.memberCount }}</text>
        <text class="td t-muted" style="flex:0.9">{{ row.leader }}</text>
        <text class="td t-muted" style="flex:1">{{ row.createdAt.slice(0,10) }}</text>
        <view class="td action-btns" style="flex:1.2">
          <view class="act-btn act-view" @click="uni.navigateTo({ url: `/pages/orgs/detail?id=${row.id}` })">详情</view>
          <view class="act-btn act-edit">编辑</view>
          <view class="act-btn act-more">···</view>
        </view>
      </view>
      <Pagination :total="total" :page="page" :page-size="pageSize"
        @page-change="onPageChange" @page-size-change="onPageSizeChange" />
    </view>
  </AppLayout>
</template>

<style lang="scss" scoped>
.stats-row { display: grid; grid-template-columns: repeat(5, 1fr); gap: 12px; margin-bottom: 16px; }
.table-card { overflow: hidden; }
.table-toolbar { display: flex; align-items: center; justify-content: space-between; padding: 14px 16px; }
.sel-info { font-size: 13px; color: var(--color-text-secondary); }
.em { color: var(--color-primary); font-weight: 600; }
.t-btn { height: 34px; border-radius: 7px; font-size: 13px; font-weight: 500; display: flex; align-items: center; padding: 0 16px; cursor: pointer; }
.t-btn-primary { background: var(--color-primary); color: #fff; }
.t-head { display: flex; padding: 10px 16px; background: var(--color-border-light); border-top: 1px solid var(--color-border-light); border-bottom: 1px solid var(--color-border); }
.th { font-size: 12px; font-weight: 600; color: var(--color-text-secondary); padding-right: 8px; }
.t-row { display: flex; align-items: center; padding: 11px 16px; border-bottom: 1px solid var(--color-border-light); transition: background 0.1s; &:hover { background: var(--color-border-light); } }
.td { font-size: 13px; color: var(--color-text-primary); padding-right: 8px; }
.t-muted { font-size: 12px; color: var(--color-text-muted); }
.org-name { font-weight: 500; }
.org-tag { background: var(--color-border-light); color: var(--color-text-primary); padding: 2px 8px; border-radius: 4px; font-size: 11px; }
.action-btns { display: flex; gap: 6px; }
.act-btn { font-size: 12px; padding: 4px 10px; border-radius: 5px; cursor: pointer; font-weight: 500; }
.act-view { background: var(--color-primary-light); color: var(--color-primary); }
.act-edit { background: #f0fdf4; color: #16a34a; }
.act-more { background: var(--color-border-light); color: var(--color-text-secondary); }
</style>
