<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import KpiCard from '@/components/common/KpiCard.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { get, post, put, del } from '@/api/request'

const breadcrumbs = [{ label: '首页' }, { label: '核心业务' }, { label: '权限配置' }]
const activeTab = ref('roles')

interface RoleRow { id: number; name: string; code: string; description: string; user_count: number; perm_count: number; status: number }
const pageStats  = reactive({ roleCount: 0, permCount: 0, userCount: 0 })
const roles      = ref<RoleRow[]>([])
const permTree   = ref<any[]>([])
const loading    = ref(false)
const maxPerm    = ref(1)

async function loadData() {
  loading.value = true
  try {
    const [statsRaw, treeRaw]: any[] = await Promise.all([get('/reports/roles'), get('/permissions')])
    roles.value = statsRaw.roles || []
    pageStats.roleCount = roles.value.length
    pageStats.userCount = statsRaw.total_users || 0
    pageStats.permCount = statsRaw.total_perms || 0
    maxPerm.value = Math.max(...roles.value.map((r: RoleRow) => r.perm_count), 1)
    permTree.value = treeRaw || []
  } catch (e: any) {
    uni.showToast({ title: e?.message || '加载失败', icon: 'none' })
  } finally { loading.value = false }
}

// ══ 新增 / 编辑角色 Modal ══
const showRoleModal = ref(false)
const roleFormMode  = ref<'create' | 'edit'>('create')
const roleSaving    = ref(false)
const editingRoleId = ref(0)
const roleForm = reactive({ name: '', code: '', description: '', status: 1 })

function openCreateRole() {
  roleFormMode.value = 'create'; editingRoleId.value = 0
  Object.assign(roleForm, { name: '', code: '', description: '', status: 1 })
  showRoleModal.value = true
}
function openEditRole(role: RoleRow) {
  roleFormMode.value = 'edit'; editingRoleId.value = role.id
  Object.assign(roleForm, { name: role.name, code: role.code, description: role.description || '', status: role.status })
  showRoleModal.value = true
}
async function saveRole() {
  if (!roleForm.name.trim()) { uni.showToast({ title: '角色名称不能为空', icon: 'none' }); return }
  if (roleFormMode.value === 'create' && !roleForm.code.trim()) { uni.showToast({ title: '角色编码不能为空', icon: 'none' }); return }
  roleSaving.value = true
  try {
    if (roleFormMode.value === 'create') {
      await post('/roles', { name: roleForm.name, code: roleForm.code, description: roleForm.description, status: roleForm.status })
      uni.showToast({ title: '角色创建成功', icon: 'success' })
    } else {
      await put(`/roles/${editingRoleId.value}`, { name: roleForm.name, description: roleForm.description, status: roleForm.status })
      uni.showToast({ title: '更新成功', icon: 'success' })
    }
    showRoleModal.value = false; loadData()
  } catch (e: any) {
    uni.showToast({ title: e?.message || '保存失败', icon: 'none' })
  } finally { roleSaving.value = false }
}

async function deleteRole(role: RoleRow) {
  uni.showModal({ title: '确认删除', content: `确认删除角色「${role.name}」？`, success: async (res) => {
    if (!res.confirm) return
    try { await del(`/roles/${role.id}`); uni.showToast({ title: '已删除', icon: 'success' }); loadData() }
    catch (e: any) { uni.showToast({ title: e?.message || '删除失败', icon: 'none' }) }
  }})
}

// ══ 配置权限 Modal ══
const showPermModal    = ref(false)
const permModalRole    = ref<RoleRow | null>(null)
const allPerms         = ref<any[]>([])
const checkedPermIds   = ref<number[]>([])
const permSaving       = ref(false)

// 扁平化权限树
const flatPerms = computed(() => {
  const list: any[] = []
  for (const node of permTree.value) {
    list.push({ ...node, indent: 0 })
    for (const child of node.children || []) list.push({ ...child, indent: 1 })
  }
  return list
})

async function openPermConfig(role: RoleRow) {
  permModalRole.value = role; showPermModal.value = true; checkedPermIds.value = []
  try {
    // 获取该角色当前拥有的权限
    const roleDetail: any = await get(`/roles/${role.id}`)
    const owned: number[] = (roleDetail?.permissions || []).map((p: any) => Number(p.id))
    checkedPermIds.value = owned
  } catch { checkedPermIds.value = [] }
}
function togglePerm(id: number) {
  const idx = checkedPermIds.value.indexOf(id)
  idx === -1 ? checkedPermIds.value.push(id) : checkedPermIds.value.splice(idx, 1)
}
function toggleAll() {
  if (checkedPermIds.value.length === flatPerms.value.length) {
    checkedPermIds.value = []
  } else {
    checkedPermIds.value = flatPerms.value.map(p => Number(p.id))
  }
}
async function savePermConfig() {
  if (!permModalRole.value) return
  permSaving.value = true
  try {
    await post(`/roles/${permModalRole.value.id}/permissions`, { permission_ids: checkedPermIds.value })
    uni.showToast({ title: '权限配置保存成功', icon: 'success' })
    showPermModal.value = false; loadData()
  } catch (e: any) {
    uni.showToast({ title: e?.message || '保存失败', icon: 'none' })
  } finally { permSaving.value = false }
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
          <view class="t-btn t-btn-primary" @click="openCreateRole">＋ 新增角色</view>
        </view>
        <view class="t-head">
          <text class="th" style="flex:1.5">角色名称</text>
          <text class="th" style="flex:1.2">编码</text>
          <text class="th" style="flex:0.8">用户数</text>
          <text class="th" style="flex:1.2">权限覆盖</text>
          <text class="th" style="flex:0.7">状态</text>
          <text class="th" style="flex:1.5">操作</text>
        </view>
        <view v-if="!roles.length && !loading" class="empty-tip">暂无角色数据</view>
        <view v-for="role in roles" :key="role.code" class="t-row">
          <view class="td role-name-cell" style="flex:1.5">
            <view class="role-av">{{ role.name.charAt(0) }}</view>
            <view>
              <text style="font-size:13px;font-weight:500">{{ role.name }}</text>
              <text class="t-muted" style="font-size:11px;display:block">{{ role.description || '—' }}</text>
            </view>
          </view>
          <text class="td t-muted mono" style="flex:1.2">{{ role.code }}</text>
          <text class="td" style="flex:0.8">{{ role.user_count.toLocaleString() }}</text>
          <view class="td" style="flex:1.2">
            <view class="perm-bar-wrap">
              <view class="perm-bar">
                <view class="perm-fill" :style="{ width: Math.min(role.perm_count / maxPerm * 100, 100) + '%' }" />
              </view>
              <text class="perm-num">{{ role.perm_count }}</text>
            </view>
          </view>
          <view class="td" style="flex:0.7">
            <StatusBadge :status="role.status===1?'success':'danger'" :label="role.status===1?'启用':'禁用'" />
          </view>
          <view class="td action-btns" style="flex:1.5">
            <view class="act-btn act-view" @click="openPermConfig(role)">配置权限</view>
            <view class="act-btn act-edit" @click="openEditRole(role)">编辑</view>
            <view class="act-btn act-danger" @click="deleteRole(role)">删除</view>
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

    <!-- 新增/编辑角色 Modal -->
    <!-- #ifdef H5 -->
    <view v-if="showRoleModal" class="modal-mask" @click.self="showRoleModal = false">
      <view class="modal-box">
        <view class="modal-header">
          <text class="modal-title">{{ roleFormMode==='create' ? '新增角色' : '编辑角色' }}</text>
          <text class="modal-close" @click="showRoleModal = false">✕</text>
        </view>
        <view class="modal-body">
          <view class="form-row">
            <text class="form-label">角色名称 <text class="req">*</text></text>
            <input class="form-input" v-model="roleForm.name" placeholder="如：运营专员"/>
          </view>
          <view class="form-row">
            <text class="form-label">角色编码 <text class="req" v-if="roleFormMode==='create'">*</text></text>
            <input class="form-input" v-model="roleForm.code" placeholder="如：operator（英文，创建后不可修改）" :disabled="roleFormMode==='edit'"/>
          </view>
          <view class="form-row">
            <text class="form-label">描述</text>
            <input class="form-input" v-model="roleForm.description" placeholder="角色说明（选填）"/>
          </view>
          <view class="form-row">
            <text class="form-label">状态</text>
            <view class="form-radio-group">
              <view class="form-radio" :class="{ active: roleForm.status===1 }" @click="roleForm.status=1">✓ 启用</view>
              <view class="form-radio" :class="{ active: roleForm.status===0 }" @click="roleForm.status=0">✗ 禁用</view>
            </view>
          </view>
        </view>
        <view class="modal-footer">
          <view class="m-btn m-btn-cancel" @click="showRoleModal = false">取消</view>
          <view class="m-btn m-btn-primary" :class="{ loading: roleSaving }" @click="saveRole">
            {{ roleSaving ? '保存中...' : '确认保存' }}
          </view>
        </view>
      </view>
    </view>

    <!-- 配置权限 Modal -->
    <view v-if="showPermModal" class="modal-mask" @click.self="showPermModal = false">
      <view class="modal-box modal-lg">
        <view class="modal-header">
          <text class="modal-title">配置权限 · {{ permModalRole?.name }}</text>
          <text class="modal-close" @click="showPermModal = false">✕</text>
        </view>
        <view class="modal-body">
          <view class="perm-select-bar">
            <text class="perm-select-info">已选 <text class="em">{{ checkedPermIds.length }}</text> / 共 {{ flatPerms.length }} 项</text>
            <view class="t-btn-sm" @click="toggleAll">{{ checkedPermIds.length === flatPerms.length ? '取消全选' : '全选' }}</view>
          </view>
          <view v-for="perm in flatPerms" :key="perm.id"
            class="perm-check-row" :class="{ indent: perm.indent > 0, checked: checkedPermIds.includes(Number(perm.id)) }"
            @click="togglePerm(Number(perm.id))">
            <view class="perm-checkbox" :class="{ checked: checkedPermIds.includes(Number(perm.id)) }">
              <text v-if="checkedPermIds.includes(Number(perm.id))">✓</text>
            </view>
            <text class="perm-icon-sm">{{ perm.type===1?'📁':perm.type===2?'🔘':'🔗' }}</text>
            <text class="perm-check-name">{{ perm.name }}</text>
            <text class="perm-check-code t-muted">{{ perm.code }}</text>
          </view>
        </view>
        <view class="modal-footer">
          <view class="m-btn m-btn-cancel" @click="showPermModal = false">取消</view>
          <view class="m-btn m-btn-primary" :class="{ loading: permSaving }" @click="savePermConfig">
            {{ permSaving ? '保存中...' : '保存权限配置' }}
          </view>
        </view>
      </view>
    </view>
    <!-- #endif -->
  </AppLayout>
</template>

<style lang="scss" scoped>
.stats-row { display:grid; grid-template-columns:repeat(3,1fr); gap:12px; margin-bottom:16px; }
.perm-panel { overflow:hidden; }
.perm-tabs { display:flex; border-bottom:1px solid var(--color-border-light); }
.p-tab { padding:14px 20px; font-size:13px; cursor:pointer; color:var(--color-text-secondary); border-bottom:2px solid transparent; &.active { color:var(--color-primary); border-bottom-color:var(--color-primary); font-weight:600; } }
.toolbar { display:flex; justify-content:flex-end; align-items:center; padding:14px 16px; gap:8px; }
.t-btn { height:34px; border-radius:7px; font-size:13px; font-weight:500; display:flex; align-items:center; padding:0 16px; cursor:pointer; }
.t-btn-primary { background:var(--color-primary); color:#fff; }
.t-btn-sm { height:28px; border-radius:6px; font-size:12px; padding:0 12px; cursor:pointer; background:var(--color-border-light); color:var(--color-text-secondary); display:flex; align-items:center; }
.t-head { display:flex; padding:10px 16px; background:var(--color-border-light); border-bottom:1px solid var(--color-border); }
.th { font-size:12px; font-weight:600; color:var(--color-text-secondary); padding-right:8px; }
.t-row { display:flex; align-items:center; padding:11px 16px; border-bottom:1px solid var(--color-border-light); &:hover{background:var(--color-border-light);} }
.td { font-size:13px; color:var(--color-text-primary); padding-right:8px; }
.t-muted { font-size:12px; color:var(--color-text-muted); }
.mono { font-family:monospace; }
.role-name-cell { display:flex; align-items:center; gap:10px; }
.role-av { width:32px; height:32px; border-radius:8px; background:var(--color-primary-light); color:var(--color-primary); display:flex; align-items:center; justify-content:center; font-weight:700; font-size:13px; flex-shrink:0; }
.perm-bar-wrap { display:flex; align-items:center; gap:6px; }
.perm-bar { flex:1; height:6px; background:var(--color-border-light); border-radius:3px; overflow:hidden; }
.perm-fill { height:100%; background:var(--color-primary); border-radius:3px; transition:width 0.3s; }
.perm-num { font-size:12px; color:var(--color-text-muted); min-width:24px; }
.action-btns { display:flex; gap:6px; flex-wrap:wrap; }
.act-btn { font-size:12px; padding:4px 10px; border-radius:5px; cursor:pointer; font-weight:500; }
.act-view { background:var(--color-primary-light); color:var(--color-primary); }
.act-edit { background:#f0fdf4; color:#16a34a; }
.act-danger { background:#fef2f2; color:#ef4444; }
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

/* Modal 通用 */
.modal-mask { position:fixed; inset:0; background:rgba(0,0,0,0.45); z-index:300; display:flex; align-items:center; justify-content:center; }
.modal-box { background:var(--color-card-bg); border-radius:16px; width:480px; max-width:92vw; box-shadow:0 20px 60px rgba(0,0,0,0.2); overflow:hidden; }
.modal-lg { width:640px; }
.modal-header { display:flex; align-items:center; justify-content:space-between; padding:18px 24px; border-bottom:1px solid var(--color-border-light); }
.modal-title { font-size:16px; font-weight:700; color:var(--color-text-primary); }
.modal-close { font-size:18px; color:var(--color-text-muted); cursor:pointer; }
.modal-body { padding:20px 24px; max-height:65vh; overflow-y:auto; }
.modal-footer { display:flex; justify-content:flex-end; gap:10px; padding:16px 24px; border-top:1px solid var(--color-border-light); }
.form-row { margin-bottom:16px; }
.form-label { font-size:13px; font-weight:500; color:var(--color-text-secondary); display:block; margin-bottom:6px; }
.req { color:#ef4444; }
.form-input { width:100%; height:38px; border:1.5px solid var(--color-border); border-radius:8px; padding:0 12px; font-size:13px; color:var(--color-text-primary); background:var(--color-card-bg); box-sizing:border-box; &:focus{border-color:var(--color-primary);outline:none;} &:disabled{background:var(--color-border-light);color:var(--color-text-muted);} }
.form-radio-group { display:flex; gap:10px; }
.form-radio { padding:6px 14px; border-radius:7px; font-size:12px; cursor:pointer; border:1.5px solid var(--color-border); color:var(--color-text-secondary); &.active{border-color:var(--color-primary);background:var(--color-primary-light);color:var(--color-primary);font-weight:600;} }
.m-btn { height:36px; border-radius:8px; font-size:13px; font-weight:500; display:flex; align-items:center; padding:0 20px; cursor:pointer; }
.m-btn-cancel  { background:var(--color-border-light); color:var(--color-text-secondary); }
.m-btn-primary { background:var(--color-primary); color:#fff; &.loading{opacity:0.7;pointer-events:none;} }

/* 权限配置 Modal */
.perm-select-bar { display:flex; align-items:center; justify-content:space-between; margin-bottom:12px; padding-bottom:12px; border-bottom:1px solid var(--color-border-light); }
.perm-select-info { font-size:13px; color:var(--color-text-secondary); }
.em { color:var(--color-primary); font-weight:600; }
.perm-check-row { display:flex; align-items:center; gap:10px; padding:9px 10px; border-radius:8px; cursor:pointer; transition:background 0.1s; &:hover{background:var(--color-border-light);} &.indent{padding-left:28px;} &.checked{background:var(--color-primary-light);} }
.perm-checkbox { width:16px; height:16px; border:1.5px solid var(--color-border); border-radius:4px; display:flex; align-items:center; justify-content:center; font-size:10px; flex-shrink:0; &.checked{background:var(--color-primary);border-color:var(--color-primary);color:#fff;} }
.perm-icon-sm { width:18px; text-align:center; font-size:13px; }
.perm-check-name { font-size:13px; color:var(--color-text-primary); min-width:100px; flex:1; }
.perm-check-code { font-size:11px; font-family:monospace; }
</style>
