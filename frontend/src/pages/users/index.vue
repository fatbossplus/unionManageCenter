<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import KpiCard from '@/components/common/KpiCard.vue'
import FilterPanel from '@/components/common/FilterPanel.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import Pagination from '@/components/common/Pagination.vue'
import { getUserList, getUserDetail, createUser, updateUser, deleteUser, batchEnable, batchDisable, type UserItem } from '@/api/user'
import { get, post } from '@/api/request'
import type { FilterField, QuickTag } from '@/components/common/FilterPanel.vue'

const breadcrumbs = [{ label: '首页' }, { label: '核心业务' }, { label: '用户管理' }]
const pageStats = reactive({ total: 0, active: 0, pending: 0, disabled: 0, todayNew: 0 })

const filterFields: FilterField[] = [
  { key: 'keyword',    label: '关键词搜索',    type: 'input',  placeholder: '姓名 / 手机号 / 邮箱' },
  { key: 'status',     label: '账号状态',      type: 'select', options: [
    { label: '全部状态', value: '' }, { label: '正常', value: '1' },
    { label: '待审核', value: '2'  }, { label: '已禁用', value: '3' },
  ]},
  { key: 'startDate',  label: '注册时间（起）', type: 'input',  placeholder: 'YYYY-MM-DD' },
  { key: 'endDate',    label: '注册时间（止）', type: 'input',  placeholder: 'YYYY-MM-DD' },
]
const quickTags: QuickTag[] = [
  { key: 'today',    label: '今日新增', color: '#10b981', params: { startDate: new Date().toISOString().slice(0, 10) } },
  { key: 'pending',  label: '待审核',   color: '#f59e0b', params: { status: '2' } },
  { key: 'disabled', label: '已禁用',   color: '#6b7280', params: { status: '3' } },
]

const list    = ref<(UserItem & { _raw: any })[]>([])
const total   = ref(0)
const page    = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const selectedIds = ref<string[]>([])
const filterParams = ref<Record<string, unknown>>({})

const statusMap: Record<number, UserItem['status']> = { 1: 'active', 2: 'pending', 3: 'disabled' }
const certMap: Record<number, UserItem['certStatus']> = { 0: 'none', 1: 'pending', 2: 'certified' }
const avatarColors = ['#1e40af', '#7c3aed', '#059669', '#dc2626', '#d97706']

function normalizeUser(u: any): UserItem & { _raw: any } {
  return {
    id: String(u.id), username: u.username, email: u.email || '',
    avatar: u.avatar, orgName: u.org_name || '',
    role: u.roles?.[0]?.code || 'member',
    status: statusMap[u.status] || 'active',
    certStatus: certMap[u.cert_status] || 'none',
    createdAt: u.created_at?.slice(0, 10) || '',
    lastLoginAt: u.last_login_at || '-',
    _raw: u,
  }
}

async function loadList() {
  loading.value = true
  try {
    const res: any = await getUserList({ ...filterParams.value as any, page: page.value, pageSize: pageSize.value })
    const rawList  = (res.list || []).map(normalizeUser)
    list.value = rawList; total.value = res.total || 0
    pageStats.total    = res.total || 0
    pageStats.active   = rawList.filter((u: UserItem) => u.status === 'active').length
    pageStats.pending  = rawList.filter((u: UserItem) => u.status === 'pending').length
    pageStats.disabled = rawList.filter((u: UserItem) => u.status === 'disabled').length
    pageStats.todayNew = rawList.filter((u: any) => u.createdAt === new Date().toISOString().slice(0, 10)).length
  } catch (e: any) {
    uni.showToast({ title: e?.message || '加载失败', icon: 'none' })
  } finally { loading.value = false }
}

function onSearch(p: Record<string, unknown>) { filterParams.value = p; page.value = 1; loadList() }
function onPageChange(p: number) { page.value = p; loadList() }
function onPageSizeChange(s: number) { pageSize.value = s; page.value = 1; loadList() }
function toggleSelect(id: string) { const i = selectedIds.value.indexOf(id); i === -1 ? selectedIds.value.push(id) : selectedIds.value.splice(i, 1) }
function toggleAll() { selectedIds.value = selectedIds.value.length === list.value.length ? [] : list.value.map(u => u.id) }

async function handleBatchEnable() {
  if (!selectedIds.value.length) return
  try { await batchEnable(selectedIds.value); uni.showToast({ title: '批量启用成功', icon: 'success' }) }
  catch { uni.showToast({ title: '操作失败', icon: 'none' }) }
  selectedIds.value = []; loadList()
}
async function handleBatchDisable() {
  if (!selectedIds.value.length) return
  try { await batchDisable(selectedIds.value); uni.showToast({ title: '批量禁用成功', icon: 'success' }) }
  catch { uni.showToast({ title: '操作失败', icon: 'none' }) }
  selectedIds.value = []; loadList()
}

// ══════════════════════════════════════════
// 新增 / 编辑 Modal
// ══════════════════════════════════════════
const showFormModal = ref(false)
const formMode = ref<'create' | 'edit'>('create')
const formSaving = ref(false)
const editingId   = ref('')
const form = reactive({ username: '', email: '', password: '', phone: '', realName: '', status: 1 })

function openCreate() {
  formMode.value = 'create'; editingId.value = ''
  Object.assign(form, { username: '', email: '', password: '', phone: '', realName: '', status: 1 })
  showFormModal.value = true
}
function openEdit(row: any) {
  formMode.value = 'edit'; editingId.value = row.id
  Object.assign(form, {
    username: row.username, email: row.email,
    password: '', phone: row._raw?.phone || '',
    realName: row._raw?.real_name || '', status: row._raw?.status || 1,
  })
  showFormModal.value = true
}
async function saveForm() {
  if (!form.username) { uni.showToast({ title: '用户名不能为空', icon: 'none' }); return }
  if (formMode.value === 'create' && !form.password) { uni.showToast({ title: '密码不能为空', icon: 'none' }); return }
  formSaving.value = true
  try {
    const payload: any = { username: form.username, email: form.email, phone: form.phone, real_name: form.realName, status: form.status }
    if (form.password) payload.password = form.password
    if (formMode.value === 'create') {
      await createUser(payload)
      uni.showToast({ title: '用户创建成功', icon: 'success' })
    } else {
      await updateUser(editingId.value, payload)
      uni.showToast({ title: '更新成功', icon: 'success' })
    }
    showFormModal.value = false; loadList()
  } catch (e: any) {
    uni.showToast({ title: e?.message || '保存失败', icon: 'none' })
  } finally { formSaving.value = false }
}

// ══════════════════════════════════════════
// 详情 Modal
// ══════════════════════════════════════════
const showDetailModal = ref(false)
const detailUser = ref<any>(null)
async function openDetail(id: string) {
  try {
    const raw: any = await getUserDetail(id)
    detailUser.value = { ...raw, ...normalizeUser(raw) }
    showDetailModal.value = true
  } catch { uni.showToast({ title: '加载详情失败', icon: 'none' }) }
}

// ══════════════════════════════════════════
// 更多菜单 (···)
// ══════════════════════════════════════════
const moreMenuRow  = ref<any>(null)
const moreMenuStyle = ref('')
function openMoreMenu(e: MouseEvent, row: any) {
  moreMenuRow.value = moreMenuRow.value?.id === row.id ? null : row
  const el = e.currentTarget as HTMLElement
  const rect = el.getBoundingClientRect()
  moreMenuStyle.value = `top:${rect.bottom + 4}px;right:${window.innerWidth - rect.right}px`
}
function closeMoreMenu() { moreMenuRow.value = null }

async function singleEnable(row: any) {
  closeMoreMenu()
  try { await batchEnable([row.id]); uni.showToast({ title: '已启用', icon: 'success' }); loadList() }
  catch { uni.showToast({ title: '操作失败', icon: 'none' }) }
}
async function singleDisable(row: any) {
  closeMoreMenu()
  try { await batchDisable([row.id]); uni.showToast({ title: '已禁用', icon: 'success' }); loadList() }
  catch { uni.showToast({ title: '操作失败', icon: 'none' }) }
}
async function handleDelete(row: any) {
  closeMoreMenu()
  uni.showModal({ title: '确认删除', content: `确认删除用户「${row.username}」？此操作不可逆。`, success: async (res) => {
    if (!res.confirm) return
    try { await deleteUser(row.id); uni.showToast({ title: '已删除', icon: 'success' }); loadList() }
    catch (e: any) { uni.showToast({ title: e?.message || '删除失败', icon: 'none' }) }
  }})
}

// 角色分配
async function assignRole(userId: string, roleId: number) {
  try { await post(`/users/${userId}/roles`, { role_ids: [roleId] }); uni.showToast({ title: '角色已分配', icon: 'success' }) }
  catch (e: any) { uni.showToast({ title: e?.message || '分配失败', icon: 'none' }) }
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
      @export="() => uni.showToast({ title: '导出中...', icon: 'none' })">
      <template #extra-actions>
        <view class="fp-btn fp-btn-primary" @click="openCreate">＋ 新增用户</view>
      </template>
    </FilterPanel>

    <!-- 表格 -->
    <view class="card table-card" @click="closeMoreMenu">
      <view class="table-toolbar">
        <view class="toolbar-left">
          <text class="sel-info">共 <text class="em">{{ total }}</text> 条，已选 <text class="em">{{ selectedIds.length }}</text> 条</text>
          <view v-if="selectedIds.length" class="t-btn t-btn-outline" @click.stop="handleBatchEnable">批量启用</view>
          <view v-if="selectedIds.length" class="t-btn t-btn-danger"  @click.stop="handleBatchDisable">批量禁用</view>
        </view>
        <view class="toolbar-right">
          <view class="icon-btn" @click="loadList">🔄</view>
        </view>
      </view>

      <view class="t-head">
        <view class="th-check">
          <view class="checkbox" :class="{ checked: selectedIds.length === list.length && list.length > 0 }" @click="toggleAll">
            <text v-if="selectedIds.length === list.length && list.length > 0">✓</text>
          </view>
        </view>
        <text class="th" style="flex:2">用户信息</text>
        <text class="th" style="flex:1">所属联盟</text>
        <text class="th" style="flex:0.8">状态</text>
        <text class="th" style="flex:0.8">实名认证</text>
        <text class="th" style="flex:1">注册时间</text>
        <text class="th" style="flex:0.9">最后登录</text>
        <text class="th" style="flex:1.2">操作</text>
      </view>

      <view v-if="loading" class="table-empty">加载中...</view>
      <view v-else>
        <view v-for="row in list" :key="row.id" class="t-row" :class="{ selected: selectedIds.includes(row.id) }">
          <view class="td-check">
            <view class="checkbox" :class="{ checked: selectedIds.includes(row.id) }" @click.stop="toggleSelect(row.id)">
              <text v-if="selectedIds.includes(row.id)">✓</text>
            </view>
          </view>
          <view class="td" style="flex:2">
            <view class="user-cell">
              <view class="user-av" :style="{ background: avatarColors[row.username.charCodeAt(0) % 5] }">
                {{ row.username.charAt(0).toUpperCase() }}
              </view>
              <view>
                <text class="user-name">{{ row.username }}</text>
                <text class="user-email">{{ row.email || '—' }}</text>
              </view>
            </view>
          </view>
          <view class="td" style="flex:1"><text class="org-tag">{{ row.orgName || '—' }}</text></view>
          <view class="td" style="flex:0.8">
            <StatusBadge :status="row.status==='active'?'success':row.status==='pending'?'warning':'danger'"
              :label="row.status==='active'?'正常':row.status==='pending'?'待审核':'已禁用'" />
          </view>
          <view class="td" style="flex:0.8">
            <StatusBadge :status="row.certStatus==='certified'?'success':row.certStatus==='pending'?'warning':'danger'"
              :label="row.certStatus==='certified'?'已认证':row.certStatus==='pending'?'审核中':'未认证'" />
          </view>
          <view class="td" style="flex:1"><text class="t-muted">{{ row.createdAt }}</text></view>
          <view class="td" style="flex:0.9"><text class="t-muted">{{ row.lastLoginAt !== '-' ? row.lastLoginAt.slice(0,10) : '—' }}</text></view>
          <view class="td" style="flex:1.2">
            <view class="action-btns">
              <view class="act-btn act-view" @click.stop="openDetail(row.id)">详情</view>
              <view class="act-btn act-edit" @click.stop="openEdit(row)">编辑</view>
              <view class="act-btn act-more" @click.stop="(e) => openMoreMenu(e as MouseEvent, row)">···</view>
            </view>
          </view>
        </view>
        <view v-if="!list.length" class="table-empty">暂无数据</view>
      </view>

      <Pagination :total="total" :page="page" :page-size="pageSize"
        @page-change="onPageChange" @page-size-change="onPageSizeChange" />
    </view>

    <!-- ··· 更多菜单 -->
    <!-- #ifdef H5 -->
    <view v-if="moreMenuRow" class="more-menu" :style="moreMenuStyle" @click.stop>
      <view class="mm-item" @click="openDetail(moreMenuRow.id); closeMoreMenu()">👁 查看详情</view>
      <view class="mm-item" @click="openEdit(moreMenuRow); closeMoreMenu()">✏️ 编辑信息</view>
      <view v-if="moreMenuRow.status !== 'active'" class="mm-item" @click="singleEnable(moreMenuRow)">✅ 启用账号</view>
      <view v-if="moreMenuRow.status === 'active'" class="mm-item text-warn" @click="singleDisable(moreMenuRow)">🚫 禁用账号</view>
      <view class="mm-divider"/>
      <view class="mm-item text-danger" @click="handleDelete(moreMenuRow)">🗑 删除用户</view>
    </view>
    <!-- #endif -->

    <!-- ══ 新增/编辑 Modal ══ -->
    <!-- #ifdef H5 -->
    <view v-if="showFormModal" class="modal-mask" @click.self="showFormModal = false">
      <view class="modal-box">
        <view class="modal-header">
          <text class="modal-title">{{ formMode === 'create' ? '新增用户' : '编辑用户' }}</text>
          <text class="modal-close" @click="showFormModal = false">✕</text>
        </view>
        <view class="modal-body">
          <view class="form-row">
            <text class="form-label">用户名 <text class="req">*</text></text>
            <input class="form-input" v-model="form.username" placeholder="请输入用户名" :disabled="formMode==='edit'"/>
          </view>
          <view class="form-row">
            <text class="form-label">{{ formMode === 'create' ? '密码 *' : '新密码（不修改留空）' }}</text>
            <input class="form-input" type="password" v-model="form.password" placeholder="请输入密码"/>
          </view>
          <view class="form-row">
            <text class="form-label">邮箱</text>
            <input class="form-input" v-model="form.email" placeholder="请输入邮箱"/>
          </view>
          <view class="form-row">
            <text class="form-label">手机号</text>
            <input class="form-input" v-model="form.phone" placeholder="请输入手机号"/>
          </view>
          <view class="form-row">
            <text class="form-label">真实姓名</text>
            <input class="form-input" v-model="form.realName" placeholder="请输入真实姓名"/>
          </view>
          <view class="form-row">
            <text class="form-label">状态</text>
            <view class="form-radio-group">
              <view class="form-radio" :class="{ active: form.status === 1 }" @click="form.status = 1">✓ 正常</view>
              <view class="form-radio" :class="{ active: form.status === 2 }" @click="form.status = 2">⏳ 待审核</view>
              <view class="form-radio" :class="{ active: form.status === 3 }" @click="form.status = 3">🚫 禁用</view>
            </view>
          </view>
        </view>
        <view class="modal-footer">
          <view class="m-btn m-btn-cancel" @click="showFormModal = false">取消</view>
          <view class="m-btn m-btn-primary" :class="{ loading: formSaving }" @click="saveForm">
            {{ formSaving ? '保存中...' : '确认保存' }}
          </view>
        </view>
      </view>
    </view>
    <!-- #endif -->

    <!-- ══ 详情 Modal ══ -->
    <!-- #ifdef H5 -->
    <view v-if="showDetailModal && detailUser" class="modal-mask" @click.self="showDetailModal = false">
      <view class="modal-box">
        <view class="modal-header">
          <text class="modal-title">用户详情</text>
          <text class="modal-close" @click="showDetailModal = false">✕</text>
        </view>
        <view class="modal-body">
          <view class="detail-avatar-row">
            <view class="detail-av" :style="{ background: avatarColors[detailUser.username?.charCodeAt(0) % 5] }">
              {{ detailUser.username?.charAt(0).toUpperCase() }}
            </view>
            <view>
              <text class="detail-name">{{ detailUser.username }}</text>
              <StatusBadge :status="detailUser.status==='active'?'success':detailUser.status==='pending'?'warning':'danger'"
                :label="detailUser.status==='active'?'正常':detailUser.status==='pending'?'待审核':'已禁用'" />
            </view>
          </view>
          <view class="detail-grid">
            <view class="dg-item"><text class="dg-label">用户ID</text><text class="dg-val">{{ detailUser.id }}</text></view>
            <view class="dg-item"><text class="dg-label">邮箱</text><text class="dg-val">{{ detailUser.email || '—' }}</text></view>
            <view class="dg-item"><text class="dg-label">手机号</text><text class="dg-val">{{ detailUser._raw?.phone || '—' }}</text></view>
            <view class="dg-item"><text class="dg-label">真实姓名</text><text class="dg-val">{{ detailUser._raw?.real_name || '—' }}</text></view>
            <view class="dg-item"><text class="dg-label">所属联盟</text><text class="dg-val">{{ detailUser.orgName || '—' }}</text></view>
            <view class="dg-item"><text class="dg-label">实名认证</text><text class="dg-val">{{ detailUser.certStatus==='certified'?'已认证':detailUser.certStatus==='pending'?'审核中':'未认证' }}</text></view>
            <view class="dg-item"><text class="dg-label">注册时间</text><text class="dg-val">{{ detailUser.createdAt }}</text></view>
            <view class="dg-item"><text class="dg-label">最后登录</text><text class="dg-val">{{ detailUser.lastLoginAt !== '-' ? detailUser.lastLoginAt?.slice(0,16) : '—' }}</text></view>
          </view>
        </view>
        <view class="modal-footer">
          <view class="m-btn m-btn-cancel" @click="showDetailModal = false">关闭</view>
          <view class="m-btn m-btn-primary" @click="() => { showDetailModal = false; openEdit(detailUser) }">编辑信息</view>
        </view>
      </view>
    </view>
    <!-- #endif -->

  </AppLayout>
</template>

<style lang="scss" scoped>
.stats-row { display: grid; grid-template-columns: repeat(5, 1fr); gap: 12px; margin-bottom: 16px; }
.table-card { overflow: visible; position: relative; }
.table-toolbar { display: flex; align-items: center; justify-content: space-between; padding: 14px 16px; }
.toolbar-left { display: flex; align-items: center; gap: 8px; }
.toolbar-right { display: flex; align-items: center; gap: 8px; }
.sel-info { font-size: 13px; color: var(--color-text-secondary); }
.em { color: var(--color-primary); font-weight: 600; }
.t-btn { height: 30px; border-radius: 6px; font-size: 12px; font-weight: 500; display: flex; align-items: center; padding: 0 12px; cursor: pointer; }
.t-btn-outline { background: var(--color-card-bg); color: var(--color-text-secondary); border: 1px solid var(--color-border); }
.t-btn-danger  { background: var(--color-card-bg); color: #ef4444; border: 1px solid #fecaca; }
.icon-btn { width: 32px; height: 32px; border-radius: 7px; border: 1px solid var(--color-border); background: var(--color-card-bg); display: flex; align-items: center; justify-content: center; font-size: 14px; cursor: pointer; }
.fp-btn { height: 34px; border-radius: 7px; border: none; cursor: pointer; font-size: 13px; font-weight: 500; display: flex; align-items: center; gap: 6px; padding: 0 16px; }
.fp-btn-primary { background: var(--color-primary); color: #fff; }
.t-head { display: flex; align-items: center; padding: 10px 16px; background: var(--color-border-light); border-top: 1px solid var(--color-border-light); border-bottom: 1px solid var(--color-border); }
.th-check { width: 40px; flex-shrink: 0; }
.th { font-size: 12px; font-weight: 600; color: var(--color-text-secondary); padding-right: 8px; }
.t-row { display: flex; align-items: center; padding: 11px 16px; border-bottom: 1px solid var(--color-border-light); transition: background 0.1s; &:hover { background: var(--color-border-light); } &.selected { background: var(--color-primary-light); } &:last-child { border: none; } }
.td-check { width: 40px; flex-shrink: 0; }
.td { font-size: 13px; color: var(--color-text-primary); padding-right: 8px; }
.t-muted { font-size: 12px; color: var(--color-text-muted); }
.checkbox { width: 16px; height: 16px; border: 1.5px solid var(--color-border); border-radius: 4px; cursor: pointer; display: flex; align-items: center; justify-content: center; font-size: 10px; &.checked { background: var(--color-primary); border-color: var(--color-primary); color: #fff; } }
.user-cell { display: flex; align-items: center; gap: 10px; }
.user-av { width: 32px; height: 32px; border-radius: 50%; flex-shrink: 0; display: flex; align-items: center; justify-content: center; font-size: 12px; font-weight: 700; color: #fff; }
.user-name { font-size: 13px; font-weight: 500; color: var(--color-text-primary); display: block; }
.user-email { font-size: 11px; color: var(--color-text-muted); display: block; }
.org-tag { background: var(--color-border-light); color: var(--color-text-primary); padding: 2px 8px; border-radius: 4px; font-size: 11px; }
.action-btns { display: flex; gap: 6px; }
.act-btn { font-size: 12px; padding: 4px 10px; border-radius: 5px; cursor: pointer; font-weight: 500; user-select: none; }
.act-view { background: var(--color-primary-light); color: var(--color-primary); }
.act-edit { background: #f0fdf4; color: #16a34a; }
.act-more { background: var(--color-border-light); color: var(--color-text-secondary); letter-spacing: 2px; }
.table-empty { padding: 40px; text-align: center; color: var(--color-text-muted); font-size: 13px; }

/* 更多菜单 */
.more-menu {
  position: fixed; z-index: 200; background: var(--color-card-bg);
  border: 1px solid var(--color-border); border-radius: 10px;
  box-shadow: 0 8px 24px rgba(0,0,0,0.12); padding: 6px 0; min-width: 140px;
}
.mm-item { padding: 9px 16px; font-size: 13px; cursor: pointer; color: var(--color-text-primary); &:hover { background: var(--color-border-light); } }
.mm-divider { height: 1px; background: var(--color-border-light); margin: 4px 0; }
.text-warn { color: #f59e0b; }
.text-danger { color: #ef4444; }

/* Modal 公共样式 */
.modal-mask { position: fixed; inset: 0; background: rgba(0,0,0,0.45); z-index: 300; display: flex; align-items: center; justify-content: center; }
.modal-box { background: var(--color-card-bg); border-radius: 16px; width: 480px; max-width: 92vw; box-shadow: 0 20px 60px rgba(0,0,0,0.2); overflow: hidden; }
.modal-header { display: flex; align-items: center; justify-content: space-between; padding: 18px 24px; border-bottom: 1px solid var(--color-border-light); }
.modal-title { font-size: 16px; font-weight: 700; color: var(--color-text-primary); }
.modal-close { font-size: 18px; color: var(--color-text-muted); cursor: pointer; line-height: 1; &:hover { color: var(--color-text-primary); } }
.modal-body { padding: 20px 24px; max-height: 60vh; overflow-y: auto; }
.modal-footer { display: flex; justify-content: flex-end; gap: 10px; padding: 16px 24px; border-top: 1px solid var(--color-border-light); }
.form-row { margin-bottom: 16px; }
.form-label { font-size: 13px; font-weight: 500; color: var(--color-text-secondary); display: block; margin-bottom: 6px; }
.req { color: #ef4444; }
.form-input { width: 100%; height: 38px; border: 1.5px solid var(--color-border); border-radius: 8px; padding: 0 12px; font-size: 13px; color: var(--color-text-primary); background: var(--color-card-bg); box-sizing: border-box; &:focus { border-color: var(--color-primary); outline: none; } &:disabled { background: var(--color-border-light); color: var(--color-text-muted); } }
.form-radio-group { display: flex; gap: 10px; }
.form-radio { padding: 6px 14px; border-radius: 7px; font-size: 12px; cursor: pointer; border: 1.5px solid var(--color-border); color: var(--color-text-secondary); &.active { border-color: var(--color-primary); background: var(--color-primary-light); color: var(--color-primary); font-weight: 600; } }
.m-btn { height: 36px; border-radius: 8px; font-size: 13px; font-weight: 500; display: flex; align-items: center; padding: 0 20px; cursor: pointer; }
.m-btn-cancel  { background: var(--color-border-light); color: var(--color-text-secondary); }
.m-btn-primary { background: var(--color-primary); color: #fff; &.loading { opacity: 0.7; pointer-events: none; } }

/* 详情 Modal */
.detail-avatar-row { display: flex; align-items: center; gap: 16px; margin-bottom: 20px; padding-bottom: 16px; border-bottom: 1px solid var(--color-border-light); }
.detail-av { width: 52px; height: 52px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 20px; font-weight: 700; color: #fff; flex-shrink: 0; }
.detail-name { font-size: 18px; font-weight: 700; color: var(--color-text-primary); display: block; margin-bottom: 6px; }
.detail-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
.dg-item { display: flex; flex-direction: column; gap: 3px; }
.dg-label { font-size: 11px; color: var(--color-text-muted); }
.dg-val { font-size: 13px; font-weight: 500; color: var(--color-text-primary); }
</style>
