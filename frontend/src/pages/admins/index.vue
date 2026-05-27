<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import IconBtn from '@/components/common/IconBtn.vue'
import { get, post, put, del } from '@/api/request'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const breadcrumbs = [{ label: '系统设置' }, { label: '管理员管理' }]

// ── 列表 ──
interface AdminItem {
  id: number; username: string; email: string|null; phone: string
  real_name: string; avatar: string; role_id: number; role_name: string
  role_code: string; status: number; last_login_at: string|null
}
const list      = ref<AdminItem[]>([])
const total     = ref(0)
const page      = ref(1)
const pageSize  = ref(20)
const loading   = ref(false)
const keyword   = ref('')

async function loadList() {
  loading.value = true
  try {
    const res: any = await get('/admins', { page: page.value, page_size: pageSize.value, keyword: keyword.value })
    list.value  = res.list || []
    total.value = res.total || 0
  } catch (e: any) {
    uni.showToast({ title: e?.message || '加载失败', icon: 'none' })
  } finally { loading.value = false }
}

// ── 角色列表（用于下拉）──
const roles = ref<{ id: number; name: string; code: string }[]>([])
async function loadRoles() {
  try { roles.value = (await get('/roles') as any) || [] }
  catch { /* ignore */ }
}

// ── 新增/编辑 Modal ──
const showFormModal = ref(false)
const formMode = ref<'create'|'edit'>('create')
const formSaving = ref(false)
const editingId  = ref(0)
const form = reactive({ username: '', password: '', email: '', phone: '', real_name: '', role_id: 4, status: 1 })

function openCreate() {
  formMode.value = 'create'; editingId.value = 0
  Object.assign(form, { username: '', password: '', email: '', phone: '', real_name: '', role_id: 4, status: 1 })
  showFormModal.value = true
}
function openEdit(row: AdminItem) {
  formMode.value = 'edit'; editingId.value = row.id
  Object.assign(form, { username: row.username, password: '', email: row.email || '',
    phone: row.phone, real_name: row.real_name, role_id: row.role_id, status: row.status })
  showFormModal.value = true
}
async function saveForm() {
  if (!form.username) { uni.showToast({ title: '用户名不能为空', icon: 'none' }); return }
  if (formMode.value === 'create' && !form.password) { uni.showToast({ title: '密码不能为空', icon: 'none' }); return }
  formSaving.value = true
  try {
    const payload: any = { username: form.username, email: form.email || undefined,
      phone: form.phone, real_name: form.real_name, role_id: form.role_id, status: form.status }
    if (form.password) payload.password = form.password
    if (formMode.value === 'create') {
      await post('/admins', payload)
      uni.showToast({ title: '管理员已创建', icon: 'success' })
    } else {
      await put(`/admins/${editingId.value}`, payload)
      uni.showToast({ title: '更新成功', icon: 'success' })
    }
    showFormModal.value = false; loadList()
  } catch (e: any) {
    uni.showToast({ title: e?.message || '保存失败', icon: 'none' })
  } finally { formSaving.value = false }
}

// ── 重置密码 Modal ──
const showResetModal = ref(false)
const resetAdminId  = ref(0)
const resetAdminName = ref('')
const resetPwd      = ref('')
const resetSaving   = ref(false)
function openReset(row: AdminItem) {
  resetAdminId.value = row.id; resetAdminName.value = row.username
  resetPwd.value = ''; showResetModal.value = true
}
async function doReset() {
  if (!resetPwd.value || resetPwd.value.length < 6) {
    uni.showToast({ title: '密码至少6位', icon: 'none' }); return
  }
  resetSaving.value = true
  try {
    await post(`/admins/${resetAdminId.value}/reset-password`, { password: resetPwd.value })
    uni.showToast({ title: '密码已重置', icon: 'success' })
    showResetModal.value = false
  } catch (e: any) {
    uni.showToast({ title: e?.message || '重置失败', icon: 'none' })
  } finally { resetSaving.value = false }
}

// ── 删除 ──
function handleDelete(row: AdminItem) {
  if (row.id === Number(userStore.info?.id)) {
    uni.showToast({ title: '不能删除自己', icon: 'none' }); return
  }
  uni.showModal({ title: '确认删除', content: `确认删除管理员「${row.username}」？`,
    success: async (res) => {
      if (!res.confirm) return
      try { await del(`/admins/${row.id}`); uni.showToast({ title: '已删除', icon: 'success' }); loadList() }
      catch (e: any) { uni.showToast({ title: e?.message || '删除失败', icon: 'none' }) }
    }
  })
}

// ── 更多菜单 ──
const moreMenuRow   = ref<AdminItem|null>(null)
const moreMenuStyle = ref('')
function openMoreMenu(e: MouseEvent, row: AdminItem) {
  moreMenuRow.value = moreMenuRow.value?.id === row.id ? null : row
  const el = e.currentTarget as HTMLElement
  const rect = el.getBoundingClientRect()
  moreMenuStyle.value = `top:${rect.bottom+4}px;right:${window.innerWidth-rect.right}px`
}
function closeMoreMenu() { moreMenuRow.value = null }

const roleColors: Record<string, string> = {
  superadmin: '#ef4444', org_admin: '#3b82f6',
  finance: '#10b981', operator: '#f59e0b'
}

onMounted(() => { loadList(); loadRoles() })
</script>

<template>
  <AppLayout :breadcrumbs="breadcrumbs">

    <!-- 工具栏 -->
    <view class="card toolbar-card">
      <view class="tb-left">
        <input class="tb-search" v-model="keyword" placeholder="搜索用户名/姓名" @confirm="loadList"/>
        <view class="tb-btn" @click="loadList">🔍 查询</view>
      </view>
      <view class="tb-right">
        <view v-if="userStore.hasPermission('admin:create')" class="tb-btn tb-btn-primary" @click="openCreate">＋ 新增管理员</view>
      </view>
    </view>

    <!-- 表格 -->
    <view class="card table-card" @click="closeMoreMenu">
      <view class="table-wrap">
        <!-- 表头 -->
        <view class="tr th-row">
          <view class="th" style="flex:0.4">ID</view>
          <view class="th" style="flex:1.2">账号</view>
          <view class="th" style="flex:1">真实姓名</view>
          <view class="th" style="flex:0.8">角色</view>
          <view class="th" style="flex:0.8">状态</view>
          <view class="th" style="flex:1.2">最后登录</view>
          <view class="th" style="flex:1">操作</view>
        </view>

        <view v-for="row in list" :key="row.id" class="tr">
          <view class="td" style="flex:0.4"><text class="t-muted">{{ row.id }}</text></view>
          <view class="td" style="flex:1.2">
            <view class="admin-avatar" :style="{ background: roleColors[row.role_code] || '#6366f1' }">
              {{ row.username.charAt(0).toUpperCase() }}
            </view>
            <view>
              <text class="admin-name">{{ row.username }}</text>
              <text class="admin-email">{{ row.email || '—' }}</text>
            </view>
          </view>
          <view class="td" style="flex:1"><text>{{ row.real_name || '—' }}</text></view>
          <view class="td" style="flex:0.8">
            <view class="role-tag" :style="{ background: (roleColors[row.role_code]||'#6366f1')+'22', color: roleColors[row.role_code]||'#6366f1' }">
              {{ row.role_name }}
            </view>
          </view>
          <view class="td" style="flex:0.8">
            <view :class="row.status===1 ? 'badge-ok' : 'badge-off'">
              {{ row.status===1 ? '正常' : '禁用' }}
            </view>
          </view>
          <view class="td" style="flex:1.2">
            <text class="t-muted">{{ row.last_login_at ? row.last_login_at.slice(0,16) : '从未登录' }}</text>
          </view>
          <view class="td" style="flex:1">
            <view class="action-btns">
              <IconBtn v-if="userStore.hasPermission('admin:update')" icon="✏️" tip="编辑管理员" type="edit" @click="openEdit(row)" />
              <IconBtn v-if="userStore.hasPermission('admin:update')" icon="🔑" tip="重置密码" type="key" @click="openReset(row)" />
              <IconBtn icon="⋯" tip="更多操作" type="default" @click="(e:any)=>openMoreMenu(e as MouseEvent, row)" />
            </view>
          </view>
        </view>
        <view v-if="!list.length" class="table-empty">暂无管理员数据</view>
      </view>

      <!-- 分页 -->
      <view class="pagination-bar">
        <text class="pg-info">共 {{ total }} 条</text>
        <view class="pg-btn" :class="{disabled: page<=1}" @click="page>1&&(page--,loadList())">上一页</view>
        <text class="pg-cur">{{ page }}</text>
        <view class="pg-btn" :class="{disabled: page*pageSize>=total}" @click="page*pageSize<total&&(page++,loadList())">下一页</view>
      </view>
    </view>

    <!-- 更多菜单 -->
    <!-- #ifdef H5 -->
    <view v-if="moreMenuRow" class="more-menu" :style="moreMenuStyle" @click.stop>
      <view v-if="userStore.hasPermission('admin:update')" class="mm-item" @click="openEdit(moreMenuRow!); closeMoreMenu()">✏️ 编辑</view>
      <view v-if="userStore.hasPermission('admin:update')" class="mm-item" @click="openReset(moreMenuRow!); closeMoreMenu()">🔑 重置密码</view>
      <view class="mm-divider"/>
      <view v-if="userStore.hasPermission('admin:delete') && moreMenuRow!.id !== Number(userStore.info?.id)"
            class="mm-item text-danger" @click="handleDelete(moreMenuRow!); closeMoreMenu()">🗑 删除</view>
    </view>
    <!-- #endif -->

    <!-- 新增/编辑 Modal -->
    <!-- #ifdef H5 -->
    <view v-if="showFormModal" class="modal-mask" @click.self="showFormModal=false">
      <view class="modal-box">
        <view class="modal-header">
          <text class="modal-title">{{ formMode==='create'?'新增管理员':'编辑管理员' }}</text>
          <text class="modal-close" @click="showFormModal=false">✕</text>
        </view>
        <view class="modal-body">
          <view class="form-row">
            <text class="form-label">用户名 <text class="req">*</text></text>
            <input class="form-input" v-model="form.username" placeholder="登录用户名" :disabled="formMode==='edit'"/>
          </view>
          <view class="form-row">
            <text class="form-label">{{ formMode==='create'?'密码 *':'新密码（留空不修改）' }}</text>
            <input class="form-input" type="password" v-model="form.password" placeholder="至少6位"/>
          </view>
          <view class="form-row">
            <text class="form-label">真实姓名</text>
            <input class="form-input" v-model="form.real_name" placeholder="请输入真实姓名"/>
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
            <text class="form-label">角色 <text class="req">*</text></text>
            <view class="form-radio-group">
              <view v-for="r in roles" :key="r.id"
                    class="form-radio" :class="{active: form.role_id===r.id}"
                    :style="form.role_id===r.id?{borderColor: roleColors[r.code]||'#6366f1', background:(roleColors[r.code]||'#6366f1')+'22', color:roleColors[r.code]||'#6366f1'}:{}"
                    @click="form.role_id=r.id">
                {{ r.name }}
              </view>
            </view>
          </view>
          <view class="form-row">
            <text class="form-label">状态</text>
            <view class="form-radio-group">
              <view class="form-radio" :class="{active: form.status===1}" @click="form.status=1">✓ 正常</view>
              <view class="form-radio" :class="{active: form.status===0}" @click="form.status=0">🚫 禁用</view>
            </view>
          </view>
        </view>
        <view class="modal-footer">
          <view class="m-btn m-btn-cancel" @click="showFormModal=false">取消</view>
          <view class="m-btn m-btn-primary" :class="{loading:formSaving}" @click="saveForm">
            {{ formSaving?'保存中...':'确认保存' }}
          </view>
        </view>
      </view>
    </view>
    <!-- #endif -->

    <!-- 重置密码 Modal -->
    <!-- #ifdef H5 -->
    <view v-if="showResetModal" class="modal-mask" @click.self="showResetModal=false">
      <view class="modal-box" style="width:360px">
        <view class="modal-header">
          <text class="modal-title">重置密码 — {{ resetAdminName }}</text>
          <text class="modal-close" @click="showResetModal=false">✕</text>
        </view>
        <view class="modal-body">
          <view class="form-row">
            <text class="form-label">新密码 <text class="req">*</text></text>
            <input class="form-input" type="password" v-model="resetPwd" placeholder="至少6位"/>
          </view>
        </view>
        <view class="modal-footer">
          <view class="m-btn m-btn-cancel" @click="showResetModal=false">取消</view>
          <view class="m-btn m-btn-primary" :class="{loading:resetSaving}" @click="doReset">
            {{ resetSaving?'重置中...':'确认重置' }}
          </view>
        </view>
      </view>
    </view>
    <!-- #endif -->

  </AppLayout>
</template>

<style lang="scss" scoped>
.toolbar-card { display: flex; justify-content: space-between; align-items: center; padding: 14px 20px; margin-bottom: 16px; }
.tb-left, .tb-right { display: flex; gap: 10px; align-items: center; }
.tb-search { height: 36px; border: 1.5px solid var(--color-border); border-radius: 8px; padding: 0 12px; font-size: 13px; width: 220px; background: var(--color-card-bg); color: var(--color-text-primary); }
.tb-btn { height: 36px; padding: 0 16px; border-radius: 8px; font-size: 13px; cursor: pointer; display: flex; align-items: center; background: var(--color-border-light); color: var(--color-text-secondary); }
.tb-btn-primary { background: var(--color-primary); color: #fff; }
.table-card { padding: 0; overflow: visible; position: relative; }
.table-wrap { overflow-x: auto; }
.tr { display: flex; align-items: center; border-bottom: 1px solid var(--color-border-light); padding: 0 20px; }
.th-row { background: var(--color-border-light); font-weight: 600; }
.th, .td { padding: 12px 8px; font-size: 13px; color: var(--color-text-primary); overflow: visible; }
.th { font-size: 12px; color: var(--color-text-secondary); }
.t-muted { color: var(--color-text-muted); font-size: 12px; }
.admin-avatar { width: 32px; height: 32px; border-radius: 50%; display: flex; align-items: center; justify-content: center; color: #fff; font-size: 13px; font-weight: bold; flex-shrink: 0; margin-right: 8px; }
.td { display: flex; align-items: center; }
.admin-name { font-size: 13px; font-weight: 600; display: block; }
.admin-email { font-size: 11px; color: var(--color-text-muted); display: block; }
.role-tag { padding: 3px 10px; border-radius: 5px; font-size: 11px; font-weight: 600; }
.badge-ok { color: #16a34a; background: #f0fdf4; padding: 3px 10px; border-radius: 5px; font-size: 11px; font-weight: 600; }
.badge-off { color: #ef4444; background: #fef2f2; padding: 3px 10px; border-radius: 5px; font-size: 11px; font-weight: 600; }
.action-btns { display: flex; gap: 6px; }
.act-btn { font-size: 12px; padding: 4px 10px; border-radius: 5px; cursor: pointer; font-weight: 500; }
.act-edit  { background: #f0fdf4; color: #16a34a; }
.act-reset { background: #eff6ff; color: #3b82f6; }
.act-more  { background: var(--color-border-light); color: var(--color-text-secondary); letter-spacing: 2px; }
.table-empty { padding: 40px; text-align: center; color: var(--color-text-muted); font-size: 13px; }
.pagination-bar { display: flex; align-items: center; gap: 12px; padding: 14px 20px; border-top: 1px solid var(--color-border-light); }
.pg-info { font-size: 13px; color: var(--color-text-muted); }
.pg-btn { padding: 5px 14px; border: 1px solid var(--color-border); border-radius: 6px; font-size: 12px; cursor: pointer; &.disabled { opacity: 0.4; pointer-events: none; } }
.pg-cur { font-size: 13px; font-weight: 600; min-width: 24px; text-align: center; }

/* 更多菜单 */
.more-menu { position: fixed; z-index: 200; background: var(--color-card-bg); border: 1px solid var(--color-border); border-radius: 10px; box-shadow: 0 8px 24px rgba(0,0,0,0.12); padding: 6px 0; min-width: 140px; }
.mm-item { padding: 9px 16px; font-size: 13px; cursor: pointer; &:hover { background: var(--color-border-light); } }
.mm-divider { height: 1px; background: var(--color-border-light); margin: 4px 0; }
.text-danger { color: #ef4444; }

/* Modal */
.modal-mask { position: fixed; inset: 0; background: rgba(0,0,0,0.45); z-index: 300; display: flex; align-items: center; justify-content: center; }
.modal-box { background: var(--color-card-bg); border-radius: 16px; width: 500px; max-width: 92vw; box-shadow: 0 20px 60px rgba(0,0,0,0.2); overflow: hidden; }
.modal-header { display: flex; align-items: center; justify-content: space-between; padding: 18px 24px; border-bottom: 1px solid var(--color-border-light); }
.modal-title { font-size: 16px; font-weight: 700; }
.modal-close { font-size: 18px; color: var(--color-text-muted); cursor: pointer; }
.modal-body { padding: 20px 24px; max-height: 60vh; overflow-y: auto; }
.modal-footer { display: flex; justify-content: flex-end; gap: 10px; padding: 16px 24px; border-top: 1px solid var(--color-border-light); }
.form-row { margin-bottom: 16px; }
.form-label { font-size: 13px; font-weight: 500; color: var(--color-text-secondary); display: block; margin-bottom: 6px; }
.req { color: #ef4444; }
.form-input { width: 100%; height: 38px; border: 1.5px solid var(--color-border); border-radius: 8px; padding: 0 12px; font-size: 13px; background: var(--color-card-bg); color: var(--color-text-primary); box-sizing: border-box; &:disabled { background: var(--color-border-light); } }
.form-radio-group { display: flex; gap: 8px; flex-wrap: wrap; }
.form-radio { padding: 6px 14px; border-radius: 7px; font-size: 12px; cursor: pointer; border: 1.5px solid var(--color-border); color: var(--color-text-secondary); &.active { border-color: var(--color-primary); background: var(--color-primary-light); color: var(--color-primary); font-weight: 600; } }
.m-btn { height: 36px; border-radius: 8px; font-size: 13px; font-weight: 500; display: flex; align-items: center; padding: 0 20px; cursor: pointer; }
.m-btn-cancel  { background: var(--color-border-light); color: var(--color-text-secondary); }
.m-btn-primary { background: var(--color-primary); color: #fff; &.loading { opacity: 0.7; pointer-events: none; } }
</style>
