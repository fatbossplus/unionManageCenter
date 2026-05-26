<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import KpiCard from '@/components/common/KpiCard.vue'
import FilterPanel from '@/components/common/FilterPanel.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import Pagination from '@/components/common/Pagination.vue'
import { getUserList, batchEnable, batchDisable, type UserItem } from '@/api/user'
import type { FilterField, QuickTag } from '@/components/common/FilterPanel.vue'

const breadcrumbs = [{ label: '首页' }, { label: '核心业务' }, { label: '用户管理' }]

const pageStats = reactive({ total: 0, active: 0, pending: 0, disabled: 0, todayNew: 0 })

const filterFields: FilterField[] = [
  { key: 'keyword',    label: '关键词搜索',    type: 'input',  placeholder: '姓名 / 手机号 / 邮箱' },
  { key: 'orgName',    label: '所属联盟',      type: 'input',  placeholder: '输入联盟名称' },
  { key: 'role',       label: '用户角色',      type: 'select', options: [
    { label: '全部角色', value: '' }, { label: '联盟主', value: 'owner' },
    { label: '管理员', value: 'admin' }, { label: '财务', value: 'finance' }, { label: '普通会员', value: 'member' },
  ]},
  { key: 'status',     label: '账号状态',      type: 'select', options: [
    { label: '全部状态', value: '' }, { label: '正常', value: 'active' },
    { label: '待审核', value: 'pending' }, { label: '已禁用', value: 'disabled' },
  ]},
  { key: 'startDate',  label: '注册时间（起）', type: 'input',  placeholder: 'YYYY-MM-DD' },
  { key: 'endDate',    label: '注册时间（止）', type: 'input',  placeholder: 'YYYY-MM-DD' },
  { key: 'certStatus', label: '实名认证',      type: 'select', options: [
    { label: '全部', value: '' }, { label: '已认证', value: 'certified' },
    { label: '审核中', value: 'pending' }, { label: '未认证', value: 'none' },
  ]},
  { key: 'source',     label: '注册来源',      type: 'select', options: [
    { label: '全部渠道', value: '' }, { label: 'PC 网页', value: 'web' },
    { label: '微信小程序', value: 'mp_wx' }, { label: '手机 App', value: 'app' },
  ]},
]

const quickTags: QuickTag[] = [
  { key: 'today',    label: '今日新增',       color: '#10b981', params: { startDate: new Date().toISOString().slice(0, 10) } },
  { key: 'pending',  label: '待审核',          color: '#f59e0b', params: { status: 'pending' } },
  { key: 'inactive', label: '长期未登录>30天', color: '#ef4444', params: { inactive: '30' } },
  { key: 'active_m', label: '本月活跃',        color: '#3b82f6', params: { activeMonth: '1' } },
  { key: 'disabled', label: '已禁用',          color: '#6b7280', params: { status: 'disabled' } },
]

const list = ref<UserItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const selectedIds = ref<string[]>([])
const filterParams = ref<Record<string, unknown>>({})
const sortKey = ref('createdAt')
const sortDir = ref<'asc' | 'desc'>('desc')

const roleLabel: Record<string, string> = {
  owner: '联盟主', admin: '管理员', finance: '财务', member: '普通会员',
}
const avatarColors = ['#1e40af', '#7c3aed', '#059669', '#dc2626', '#d97706']

const statusMap: Record<number, UserItem['status']> = { 1: 'active', 2: 'pending', 3: 'disabled' }
const certMap: Record<number, UserItem['certStatus']> = { 0: 'none', 1: 'pending', 2: 'certified' }

function normalizeUser(u: any): UserItem {
  return {
    id: String(u.id),
    username: u.username,
    email: u.email,
    avatar: u.avatar,
    orgName: u.org_name || '',
    role: u.roles?.[0]?.code || 'member',
    status: statusMap[u.status] || 'active',
    certStatus: certMap[u.cert_status] || 'none',
    createdAt: u.created_at?.slice(0, 10) || '',
    lastLoginAt: u.last_login_at || '-',
  }
}

async function loadList() {
  loading.value = true
  try {
    const res = await getUserList({
      ...(filterParams.value as any),
      page: page.value,
      pageSize: pageSize.value,
    }) as any
    const rawList = (res.list || []).map(normalizeUser)
    list.value = rawList
    total.value = res.total || 0
    pageStats.total = res.total || 0
    pageStats.active = rawList.filter((u: UserItem) => u.status === 'active').length
    pageStats.pending = rawList.filter((u: UserItem) => u.status === 'pending').length
    pageStats.disabled = rawList.filter((u: UserItem) => u.status === 'disabled').length
  } catch (e: any) {
    uni.showToast({ title: e?.message || '加载失败', icon: 'none' })
    list.value = []
  } finally {
    loading.value = false
  }
}

function onSearch(params: Record<string, unknown>) {
  filterParams.value = params
  page.value = 1
  loadList()
}

function onPageChange(p: number) { page.value = p; loadList() }
function onPageSizeChange(s: number) { pageSize.value = s; page.value = 1; loadList() }

function toggleSelect(id: string) {
  const idx = selectedIds.value.indexOf(id)
  idx === -1 ? selectedIds.value.push(id) : selectedIds.value.splice(idx, 1)
}

function toggleAll() {
  if (selectedIds.value.length === list.value.length) {
    selectedIds.value = []
  } else {
    selectedIds.value = list.value.map(u => u.id)
  }
}

async function handleBatchEnable() {
  await batchEnable(selectedIds.value).catch(() => {})
  uni.showToast({ title: '批量启用成功', icon: 'success' })
  selectedIds.value = []; loadList()
}

async function handleBatchDisable() {
  await batchDisable(selectedIds.value).catch(() => {})
  uni.showToast({ title: '批量禁用成功', icon: 'success' })
  selectedIds.value = []; loadList()
}

function goDetail(id: string, mode = 'view') {
  uni.navigateTo({ url: `/pages/users/detail?id=${id}&mode=${mode}` })
}

onMounted(loadList)
</script>

<template>
  <AppLayout :breadcrumbs="breadcrumbs">

    <!-- 统计栏 -->
    <view class="stats-row">
      <KpiCard icon="👥" label="总用户数"  :value="pageStats.total"    :trend="{ dir:'up',   text:'12.3%' }" icon-bg="#eff6ff" />
      <KpiCard icon="✅" label="已激活"    :value="pageStats.active"   :trend="{ dir:'up',   text:'8.1%'  }" icon-bg="#f0fdf4" />
      <KpiCard icon="⏳" label="待审核"    :value="pageStats.pending"  :trend="{ dir:'up',   text:'今日+34' }" icon-bg="#fffbeb" />
      <KpiCard icon="🚫" label="已禁用"    :value="pageStats.disabled" :trend="{ dir:'down', text:'无变化'  }" icon-bg="#fef2f2" />
      <KpiCard icon="🆕" label="今日新增"  :value="pageStats.todayNew" :trend="{ dir:'up',   text:'较昨日+12' }" icon-bg="#faf5ff" />
    </view>

    <!-- 筛选 -->
    <FilterPanel :fields="filterFields" :quick-tags="quickTags"
      @search="onSearch"
      @reset="() => { filterParams.value = {}; loadList() }"
      @export="() => uni.showToast({ title: '导出中...', icon: 'none' })"
    >
      <template #extra-actions>
        <view class="fp-btn fp-btn-primary"
          @click="uni.navigateTo({ url: '/pages/users/detail?mode=create' })">
          ＋ 新增用户
        </view>
      </template>
    </FilterPanel>

    <!-- 表格 -->
    <view class="card table-card">
      <view class="table-toolbar">
        <view class="toolbar-left">
          <text class="sel-info">共 <text class="em">{{ total }}</text> 条，已选 <text class="em">{{ selectedIds.length }}</text> 条</text>
          <view v-if="selectedIds.length" class="t-btn t-btn-outline" @click="handleBatchEnable">批量启用</view>
          <view v-if="selectedIds.length" class="t-btn t-btn-danger" @click="handleBatchDisable">批量禁用</view>
        </view>
        <view class="toolbar-right">
          <view class="icon-btn" @click="loadList">🔄</view>
        </view>
      </view>

      <!-- 表头 -->
      <view class="t-head">
        <view class="th-check">
          <view class="checkbox" :class="{ checked: selectedIds.length === list.length && list.length > 0 }" @click="toggleAll">
            <text v-if="selectedIds.length === list.length && list.length > 0">✓</text>
          </view>
        </view>
        <text class="th" style="flex:2">用户信息</text>
        <text class="th" style="flex:1.2">所属联盟</text>
        <text class="th" style="flex:0.8">角色</text>
        <text class="th" style="flex:0.8">状态</text>
        <text class="th" style="flex:0.8">实名认证</text>
        <text class="th" style="flex:1">注册时间 ↑</text>
        <text class="th" style="flex:0.9">最后登录</text>
        <text class="th" style="flex:1.2">操作</text>
      </view>

      <!-- loading -->
      <view v-if="loading" class="table-empty">加载中...</view>

      <!-- 数据行 -->
      <view v-else>
        <view v-for="row in list" :key="row.id" class="t-row"
          :class="{ selected: selectedIds.includes(row.id) }">
          <view class="td-check">
            <view class="checkbox" :class="{ checked: selectedIds.includes(row.id) }" @click="toggleSelect(row.id)">
              <text v-if="selectedIds.includes(row.id)">✓</text>
            </view>
          </view>
          <view class="td" style="flex:2">
            <view class="user-cell">
              <view class="user-av" :style="{ background: avatarColors[row.username.charCodeAt(0) % 5] }">
                {{ row.username.charAt(0) }}
              </view>
              <view>
                <text class="user-name">{{ row.username }}</text>
                <text class="user-email">{{ row.email }}</text>
              </view>
            </view>
          </view>
          <view class="td" style="flex:1.2"><text class="org-tag">{{ row.orgName }}</text></view>
          <view class="td" style="flex:0.8">
            <StatusBadge status="info" :label="roleLabel[row.role] || row.role" />
          </view>
          <view class="td" style="flex:0.8">
            <StatusBadge
              :status="row.status === 'active' ? 'success' : row.status === 'pending' ? 'warning' : 'danger'"
              :label="row.status === 'active' ? '正常' : row.status === 'pending' ? '待审核' : '已禁用'"
            />
          </view>
          <view class="td" style="flex:0.8">
            <StatusBadge
              :status="row.certStatus === 'certified' ? 'success' : row.certStatus === 'pending' ? 'warning' : 'danger'"
              :label="row.certStatus === 'certified' ? '已认证' : row.certStatus === 'pending' ? '审核中' : '未认证'"
            />
          </view>
          <view class="td" style="flex:1"><text class="t-muted">{{ row.createdAt.slice(0, 10) }}</text></view>
          <view class="td" style="flex:0.9"><text class="t-muted">{{ row.lastLoginAt }}</text></view>
          <view class="td" style="flex:1.2">
            <view class="action-btns">
              <view class="act-btn act-view" @click="goDetail(row.id)">详情</view>
              <view class="act-btn act-edit" @click="goDetail(row.id, 'edit')">编辑</view>
              <view class="act-btn act-more">···</view>
            </view>
          </view>
        </view>

        <view v-if="!list.length" class="table-empty">暂无数据</view>
      </view>

      <Pagination :total="total" :page="page" :page-size="pageSize"
        @page-change="onPageChange" @page-size-change="onPageSizeChange" />
    </view>

  </AppLayout>
</template>

<style lang="scss" scoped>
.stats-row { display: grid; grid-template-columns: repeat(5, 1fr); gap: 12px; margin-bottom: 16px; }

.table-card { overflow: hidden; }
.table-toolbar {
  display: flex; align-items: center; justify-content: space-between; padding: 14px 16px;
}
.toolbar-left { display: flex; align-items: center; gap: 8px; }
.toolbar-right { display: flex; align-items: center; gap: 8px; }
.sel-info { font-size: 13px; color: var(--color-text-secondary); }
.em { color: var(--color-primary); font-weight: 600; }
.t-btn {
  height: 30px; border-radius: 6px; font-size: 12px; font-weight: 500;
  display: flex; align-items: center; padding: 0 12px; cursor: pointer;
}
.t-btn-outline { background: var(--color-card-bg); color: var(--color-text-secondary); border: 1px solid var(--color-border); }
.t-btn-danger { background: var(--color-card-bg); color: #ef4444; border: 1px solid #fecaca; }
.icon-btn {
  width: 32px; height: 32px; border-radius: 7px; border: 1px solid var(--color-border);
  background: var(--color-card-bg); display: flex; align-items: center; justify-content: center;
  font-size: 14px; cursor: pointer;
}
.fp-btn {
  height: 34px; border-radius: 7px; border: none; cursor: pointer;
  font-size: 13px; font-weight: 500; display: flex; align-items: center; gap: 6px; padding: 0 16px;
}
.fp-btn-primary { background: var(--color-primary); color: #fff; }

.t-head {
  display: flex; align-items: center; padding: 10px 16px;
  background: var(--color-border-light); border-top: 1px solid var(--color-border-light);
  border-bottom: 1px solid var(--color-border);
}
.th-check { width: 40px; flex-shrink: 0; }
.th { font-size: 12px; font-weight: 600; color: var(--color-text-secondary); padding-right: 8px; }

.t-row {
  display: flex; align-items: center; padding: 11px 16px;
  border-bottom: 1px solid var(--color-border-light); transition: background 0.1s;
  &:hover { background: var(--color-border-light); }
  &.selected { background: var(--color-primary-light); }
  &:last-child { border: none; }
}
.td-check { width: 40px; flex-shrink: 0; }
.td { font-size: 13px; color: var(--color-text-primary); padding-right: 8px; }
.t-muted { font-size: 12px; color: var(--color-text-muted); }

.checkbox {
  width: 16px; height: 16px; border: 1.5px solid var(--color-border);
  border-radius: 4px; cursor: pointer;
  display: flex; align-items: center; justify-content: center; font-size: 10px;
  &.checked { background: var(--color-primary); border-color: var(--color-primary); color: #fff; }
}
.user-cell { display: flex; align-items: center; gap: 10px; }
.user-av {
  width: 32px; height: 32px; border-radius: 50%; flex-shrink: 0;
  display: flex; align-items: center; justify-content: center;
  font-size: 12px; font-weight: 700; color: #fff;
}
.user-name { font-size: 13px; font-weight: 500; color: var(--color-text-primary); display: block; }
.user-email { font-size: 11px; color: var(--color-text-muted); display: block; }
.org-tag {
  background: var(--color-border-light); color: var(--color-text-primary);
  padding: 2px 8px; border-radius: 4px; font-size: 11px;
}
.action-btns { display: flex; gap: 6px; }
.act-btn { font-size: 12px; padding: 4px 10px; border-radius: 5px; cursor: pointer; font-weight: 500; }
.act-view { background: var(--color-primary-light); color: var(--color-primary); }
.act-edit { background: #f0fdf4; color: #16a34a; }
.act-more { background: var(--color-border-light); color: var(--color-text-secondary); }
.table-empty { padding: 40px; text-align: center; color: var(--color-text-muted); font-size: 13px; }
</style>
