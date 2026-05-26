<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import KpiCard from '@/components/common/KpiCard.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { get, put } from '@/api/request'

const breadcrumbs = [{ label: '首页' }, { label: '核心业务' }, { label: '权限配置' }]
const activeTab = ref('roles')

interface RoleRow { name: string; code: string; user_count: number; perm_count: number; status: number }
const pageStats = reactive({ roleCount: 0, permCount: 0, userCount: 0 })
const roles = ref<RoleRow[]>([])
const permTree = ref<any[]>([])
const loading = ref(false)
const maxPerm = ref(1)

async function loadData() {
  loading.value = true
  try {
    const [statsRaw, treeRaw]: any[] = await Promise.all([
      get('/reports/roles'),
      get('/permissions'),
    ])
    roles.value = statsRaw.roles || []
    pageStats.roleCount = roles.value.length
    pageStats.userCount = statsRaw.total_users || 0
    pageStats.permCount = statsRaw.total_perms || 0
    maxPerm.value = Math.max(...roles.value.map((r: RoleRow) => r.perm_count), 1)
    permTree.value = treeRaw || []
  } catch (e: any) {
    uni.showToast({ title: e?.message || '加载失败', icon: 'none' })
  } finally {
    loading.value = false
  }
}

onMounted(loadData)
</script>

<template>
  <AppLayout :breadcrumbs="breadcrumbs">
    <view class="stats-row">
      <KpiCard icon="🔐" label="角色总数"   :value="pageStats.roleCount" :trend="{ dir:'up', text:'系统配置' }" icon-bg="#eff6ff" />
      <KpiCard icon="📋" label="权限项总数" :value="pageStats.permCount" :trend="{ dir:'up', text:'系统内置' }" icon-bg="#f0fdf4" />
      <KpiCard icon="👥" label="注册用户数" :value="pageStats.userCount" :trend="{ dir:'up', text:'实时统计' }" icon-bg="#fff7ed" />
    </view>

    <view class="card perm-panel">
      <view class="perm-tabs">
        <view class="p-tab" :class="{ active: activeTab==='roles' }" @click="activeTab='roles'">角色管理</view>
        <view class="p-tab" :class="{ active: activeTab==='perms' }" @click="activeTab='perms'">权限树</view>
      </view>

      <!-- 角色管理 -->
      <view v-if="activeTab==='roles'">
        <view class="toolbar">
          <view v-if="loading" class="loading-tip">加载中…</view>
          <view class="t-btn t-btn-primary" @click="uni.navigateTo({ url: '/pages/permissions/role-edit' })">＋ 新增角色</view>
        </view>
        <view class="t-head">
          <text class="th" style="flex:1.5">角色名称</text>
          <text class="th" style="flex:2">描述</text>
          <text class="th" style="flex:0.8">用户数</text>
          <text class="th" style="flex:1.2">权限覆盖</text>
          <text class="th" style="flex:0.8">状态</text>
          <text class="th" style="flex:1.2">操作</text>
        </view>
        <view v-if="!roles.length && !loading" class="empty-tip">暂无角色数据</view>
        <view v-for="role in roles" :key="role.code" class="t-row">
          <view class="td role-name-cell" style="flex:1.5">
            <view class="role-av">{{ role.name.charAt(0) }}</view>
            <text>{{ role.name }}</text>
          </view>
          <text class="td t-muted" style="flex:2">{{ role.code }}</text>
          <text class="td" style="flex:0.8">{{ role.user_count.toLocaleString() }}</text>
          <view class="td" style="flex:1.2">
            <view class="perm-bar">
              <view class="perm-fill" :style="{ width: Math.min(role.perm_count / maxPerm * 100, 100) + '%' }" />
            </view>
            <text class="perm-num">{{ role.perm_count }}</text>
          </view>
          <view class="td" style="flex:0.8">
            <StatusBadge :status="role.status===1?'success':'danger'" :label="role.status===1?'启用':'禁用'" />
          </view>
          <view class="td action-btns" style="flex:1.2">
            <view class="act-btn act-view">配置权限</view>
            <view class="act-btn act-edit">编辑</view>
          </view>
        </view>
      </view>

      <!-- 权限树 -->
      <view v-else-if="activeTab==='perms'">
        <view class="toolbar"><view v-if="loading" class="loading-tip">加载中…</view></view>
        <view v-if="!permTree.length && !loading" class="empty-tip">暂无权限数据</view>
        <view v-for="node in permTree" :key="node.id" class="perm-node">
          <view class="perm-node-row perm-parent">
            <text class="perm-icon">{{ node.type===1?'📁':node.type===2?'🔘':'🔗' }}</text>
            <text class="perm-name">{{ node.name }}</text>
            <text class="perm-code t-muted">{{ node.code }}</text>
            <text class="perm-path t-muted">{{ node.path }}</text>
          </view>
          <view v-for="child in node.children" :key="child.id" class="perm-node-row perm-child">
            <text class="perm-icon">{{ child.type===1?'📄':child.type===2?'▪️':'🔌' }}</text>
            <text class="perm-name">{{ child.name }}</text>
            <text class="perm-code t-muted">{{ child.code }}</text>
            <text class="perm-path t-muted">{{ child.path }}</text>
          </view>
        </view>
      </view>
    </view>
  </AppLayout>
</template>

<style lang="scss" scoped>
.stats-row { display:grid; grid-template-columns:repeat(3,1fr); gap:12px; margin-bottom:16px; }
.perm-panel { overflow:hidden; }
.perm-tabs { display:flex; border-bottom:1px solid var(--color-border-light); }
.p-tab { padding:14px 20px; font-size:13px; cursor:pointer; color:var(--color-text-secondary); border-bottom:2px solid transparent; &.active { color:var(--color-primary); border-bottom-color:var(--color-primary); font-weight:600; } }
.toolbar { display:flex; justify-content:space-between; align-items:center; padding:14px 16px; }
.t-btn { height:34px; border-radius:7px; font-size:13px; font-weight:500; display:flex; align-items:center; padding:0 16px; cursor:pointer; }
.t-btn-primary { background:var(--color-primary); color:#fff; }
.t-head { display:flex; padding:10px 16px; background:var(--color-border-light); border-bottom:1px solid var(--color-border); }
.th { font-size:12px; font-weight:600; color:var(--color-text-secondary); padding-right:8px; }
.t-row { display:flex; align-items:center; padding:11px 16px; border-bottom:1px solid var(--color-border-light); &:hover{background:var(--color-border-light);} }
.td { font-size:13px; color:var(--color-text-primary); padding-right:8px; }
.t-muted { font-size:12px; color:var(--color-text-muted); }
.role-name-cell { display:flex; align-items:center; gap:10px; }
.role-av { width:32px; height:32px; border-radius:8px; background:var(--color-primary-light); color:var(--color-primary); display:flex; align-items:center; justify-content:center; font-weight:700; font-size:13px; flex-shrink:0; }
.perm-bar { flex:1; height:6px; background:var(--color-border-light); border-radius:3px; overflow:hidden; margin-right:6px; }
.perm-fill { height:100%; background:var(--color-primary); border-radius:3px; transition:width 0.3s; }
.perm-num { font-size:12px; color:var(--color-text-muted); min-width:24px; }
.action-btns { display:flex; gap:6px; }
.act-btn { font-size:12px; padding:4px 10px; border-radius:5px; cursor:pointer; font-weight:500; }
.act-view { background:var(--color-primary-light); color:var(--color-primary); }
.act-edit { background:#f0fdf4; color:#16a34a; }
.empty-tip { text-align:center; padding:40px; color:var(--color-text-muted); font-size:13px; }
.loading-tip { font-size:12px; color:var(--color-text-muted); }
.perm-node { border-bottom:1px solid var(--color-border-light); }
.perm-node-row { display:flex; align-items:center; gap:12px; padding:10px 20px; }
.perm-parent { background:var(--color-border-light); font-weight:600; }
.perm-child { padding-left:44px; &:hover{background:var(--color-border-light);} }
.perm-icon { width:20px; text-align:center; }
.perm-name { font-size:13px; color:var(--color-text-primary); min-width:100px; }
.perm-code { font-size:11px; font-family:monospace; min-width:160px; }
.perm-path { font-size:11px; font-family:monospace; flex:1; }
</style>
