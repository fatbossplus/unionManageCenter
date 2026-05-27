<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import KpiCard from '@/components/common/KpiCard.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import Pagination from '@/components/common/Pagination.vue'
import { getMessageList, markRead as apiMarkRead, markAllRead as apiMarkAllRead } from '@/api/message'

const breadcrumbs = [{ label: '首页' }, { label: '系统' }, { label: '消息通知' }]
const pageStats = reactive({ total: 0, unread: 0, system: 0, order: 0, security: 0 })

const typeMap: Record<string, { label: string; color: string; icon: string }> = {
  system:   { label: '系统通知', color: '#3b82f6', icon: 'bell' },
  order:    { label: '订单通知', color: '#f59e0b', icon: 'order' },
  finance:  { label: '财务通知', color: '#10b981', icon: 'money' },
  security: { label: '安全告警', color: '#ef4444', icon: 'shield' },
}

interface MsgRow { id: string; title: string; content: string; type: string; read: boolean; createdAt: string }

const allList = ref<MsgRow[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const activeFilter = ref('all')

const filteredList = computed(() => {
  if (activeFilter.value === 'all') return allList.value
  if (activeFilter.value === 'unread') return allList.value.filter(m => !m.read)
  return allList.value.filter(m => m.type === activeFilter.value)
})

function normalize(raw: any): MsgRow {
  return {
    id: String(raw.id),
    title: raw.title,
    content: raw.content,
    type: raw.type,
    read: raw.is_read === 1,
    createdAt: raw.created_at?.slice(0, 16) || '',
  }
}

async function loadList() {
  loading.value = true
  try {
    const res: any = await getMessageList({ page: page.value, pageSize: pageSize.value })
    allList.value = (res.list || []).map(normalize)
    total.value = res.total ?? 0
    pageStats.total    = res.total ?? 0
    pageStats.unread   = allList.value.filter(m => !m.read).length
    pageStats.system   = allList.value.filter(m => m.type === 'system').length
    pageStats.order    = allList.value.filter(m => m.type === 'order').length
    pageStats.security = allList.value.filter(m => m.type === 'security').length
  } catch (e: any) {
    uni.showToast({ title: e?.message || '加载失败', icon: 'none' })
  } finally {
    loading.value = false
  }
}

async function markRead(id: string) {
  try {
    await apiMarkRead([id])
    const item = allList.value.find(m => m.id === id)
    if (item) item.read = true
    pageStats.unread = allList.value.filter(m => !m.read).length
  } catch {}
}

async function doMarkAllRead() {
  try {
    await apiMarkAllRead()
    allList.value.forEach(m => { m.read = true })
    pageStats.unread = 0
    uni.showToast({ title: '全部标记已读', icon: 'success' })
  } catch (e: any) {
    uni.showToast({ title: e?.message || '操作失败', icon: 'none' })
  }
}

function onPageChange(p: number) { page.value = p; loadList() }
function onPageSizeChange(s: number) { pageSize.value = s; page.value = 1; loadList() }

onMounted(loadList)
</script>

<template>
  <AppLayout :breadcrumbs="breadcrumbs">
    <view class="stats-row">
      <KpiCard icon="message" label="消息总数"   :value="pageStats.total"   :trend="{ dir:'up', text:'今日+5' }"   icon-bg="#eff6ff" />
      <KpiCard icon="warning" label="未读消息"   :value="pageStats.unread"  :trend="{ dir:'up', text:'待处理' }"   icon-bg="#fef2f2" />
      <KpiCard icon="bell" label="系统通知"   :value="pageStats.system"  :trend="{ dir:'down', text:'无新增' }" icon-bg="#eff6ff" />
      <KpiCard icon="order" label="订单通知"   :value="pageStats.order"   :trend="{ dir:'up', text:'+12' }"     icon-bg="#fffbeb" />
      <KpiCard icon="shield" label="安全告警"   :value="pageStats.security":trend="{ dir:'up', text:'需关注' }"  icon-bg="#fef2f2" />
    </view>

    <view class="card msg-panel">
      <view class="msg-toolbar">
        <view class="filter-tabs">
          <text v-for="f in [['all','全部'],['unread','未读'],['system','系统'],['order','订单'],['security','安全']]"
            :key="f[0]" class="ftab" :class="{ active: activeFilter === f[0] }"
            @click="activeFilter = f[0]">{{ f[1] }}</text>
        </view>
        <view class="t-btn t-btn-outline" @click="doMarkAllRead"><SvgIcon name="check" /> 全部已读</view>
      </view>

      <view v-if="!filteredList.length && !loading" class="empty-tip">暂无消息</view>
      <view v-for="msg in filteredList" :key="msg.id" class="msg-item" :class="{ unread: !msg.read }" @click="markRead(msg.id)">
        <view class="msg-icon" :style="{ background: typeMap[msg.type]?.color + '20', color: typeMap[msg.type]?.color }">
          <SvgIcon :name="typeMap[msg.type]?.icon || 'bell'" />
        </view>
        <view class="msg-body">
          <view class="msg-head">
            <text class="msg-title">{{ msg.title }}</text>
            <view v-if="!msg.read" class="unread-dot" />
            <text class="msg-time">{{ msg.createdAt }}</text>
          </view>
          <text class="msg-content">{{ msg.content }}</text>
          <StatusBadge :status="msg.type === 'security' ? 'danger' : msg.type === 'order' ? 'warning' : 'info'"
            :label="typeMap[msg.type]?.label || msg.type" />
        </view>
      </view>

      <Pagination :total="total" :page="page" :page-size="pageSize"
        @page-change="onPageChange" @page-size-change="onPageSizeChange" />
    </view>
  </AppLayout>
</template>

<style lang="scss" scoped>
.stats-row { display:grid; grid-template-columns:repeat(5,1fr); gap:12px; margin-bottom:16px; }
.msg-panel { overflow:hidden; }
.msg-toolbar { display:flex; align-items:center; justify-content:space-between; padding:16px 20px; border-bottom:1px solid var(--color-border-light); }
.filter-tabs { display:flex; gap:4px; }
.ftab { font-size:13px; padding:5px 14px; border-radius:7px; cursor:pointer; color:var(--color-text-secondary); &.active { background:var(--color-primary-light); color:var(--color-primary); font-weight:600; } }
.t-btn { height:32px; border-radius:7px; font-size:12px; font-weight:500; display:flex; align-items:center; padding:0 14px; cursor:pointer; }
.t-btn-outline { background:var(--color-card-bg); color:var(--color-text-secondary); border:1px solid var(--color-border); }
.msg-item {
  display:flex; gap:14px; padding:16px 20px; border-bottom:1px solid var(--color-border-light);
  cursor:pointer; transition:background 0.1s;
  &:hover { background:var(--color-border-light); }
  &.unread { background:var(--color-primary-light); }
}
.msg-icon { width:40px; height:40px; border-radius:10px; flex-shrink:0; display:flex; align-items:center; justify-content:center; font-size:18px; }
.msg-body { flex:1; min-width:0; }
.msg-head { display:flex; align-items:center; gap:8px; margin-bottom:6px; }
.msg-title { font-size:14px; font-weight:600; color:var(--color-text-primary); flex:1; }
.unread-dot { width:8px; height:8px; border-radius:50%; background:#ef4444; flex-shrink:0; }
.msg-time { font-size:12px; color:var(--color-text-muted); flex-shrink:0; }
.msg-content { font-size:13px; color:var(--color-text-secondary); margin-bottom:8px; line-height:1.5; }
.empty-tip { text-align:center; padding:40px; color:var(--color-text-muted); font-size:13px; }
</style>
