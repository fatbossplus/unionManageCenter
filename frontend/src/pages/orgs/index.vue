<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import KpiCard from '@/components/common/KpiCard.vue'
import FilterPanel from '@/components/common/FilterPanel.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import Pagination from '@/components/common/Pagination.vue'
import { getOrgList, createOrg, updateOrg, deleteOrg, type OrgItem } from '@/api/org'
import { get, post, put, del } from '@/api/request'
import type { FilterField, QuickTag } from '@/components/common/FilterPanel.vue'
import { useUserStore } from '@/stores/user'
import IconBtn from '@/components/common/IconBtn.vue'

const userStore = useUserStore()

const breadcrumbs = [{ label: '首页' }, { label: '核心业务' }, { label: '联盟管理' }]
const pageStats = reactive({ total: 0, active: 0, pending: 0, frozen: 0, todayNew: 0 })

const filterFields: FilterField[] = [
  { key: 'keyword', label: '关键词', type: 'input', placeholder: '联盟名称 / 负责人' },
  { key: 'type',    label: '联盟类型', type: 'select', options: [
    { label:'全部类型',value:'' },{ label:'电商联盟',value:'ec' },
    { label:'服务联盟',value:'service' },{ label:'内容联盟',value:'content' },{ label:'其他',value:'other' },
  ]},
  { key: 'status',  label: '审核状态', type: 'select', options: [
    { label:'全部',value:'' },{ label:'正常',value:'1' },{ label:'待审核',value:'2' },{ label:'已冻结',value:'3' },
  ]},
  { key: 'startDate', label: '成立时间（起）', type: 'input', placeholder: 'YYYY-MM-DD' },
  { key: 'endDate',   label: '成立时间（止）', type: 'input', placeholder: 'YYYY-MM-DD' },
]
const quickTags: QuickTag[] = [
  { key: 'pending', label: '待审核',   color: '#f59e0b', params: { status: '2' } },
  { key: 'active',  label: '正常运营', color: '#10b981', params: { status: '1' } },
  { key: 'frozen',  label: '已冻结',   color: '#ef4444', params: { status: '3' } },
]

const list   = ref<(OrgItem & { _raw: any })[]>([])
const total  = ref(0)
const page   = ref(1)
const pageSize = ref(20)
const loading  = ref(false)
const filterParams = ref<Record<string, unknown>>({})

const typeCodeMap: Record<string, string> = { ec: '电商联盟', service: '服务联盟', content: '内容联盟', other: '其他' }
const typeOptions = [{ label: '电商联盟', value: 'ec' }, { label: '服务联盟', value: 'service' }, { label: '内容联盟', value: 'content' }, { label: '其他', value: 'other' }]

function normalizeOrg(raw: any): OrgItem & { _raw: any } {
  const statusMap: Record<number, OrgItem['status']> = { 1: 'active', 2: 'pending', 3: 'frozen' }
  return {
    id: String(raw.id), name: raw.name,
    type: typeCodeMap[raw.type] || raw.type,
    status: statusMap[raw.status] || 'active',
    region: raw.region || '',
    memberCount: raw.member_count ?? 0,
    leader: raw.leader?.real_name || raw.leader?.username || '',
    createdAt: raw.created_at?.slice(0, 10) || '',
    _raw: raw,
  }
}

async function loadList() {
  loading.value = true
  try {
    const res: any = await get('/orgs', { ...filterParams.value, page: page.value, page_size: pageSize.value })
    list.value  = (res.list || []).map(normalizeOrg)
    total.value = res.total ?? 0
    pageStats.total   = res.total ?? 0
    pageStats.active  = list.value.filter(o => o.status === 'active').length
    pageStats.pending = list.value.filter(o => o.status === 'pending').length
    pageStats.frozen  = list.value.filter(o => o.status === 'frozen').length
  } catch (e: any) {
    uni.showToast({ title: e?.message || '加载失败', icon: 'none' })
  } finally { loading.value = false }
}

function onSearch(p: Record<string, unknown>) { filterParams.value = p; page.value = 1; loadList() }
function onPageChange(p: number) { page.value = p; loadList() }
function onPageSizeChange(s: number) { pageSize.value = s; page.value = 1; loadList() }

// ══════════════════════════════════════════
// 新增 / 编辑 Modal
// ══════════════════════════════════════════
const showFormModal = ref(false)
const formMode  = ref<'create' | 'edit'>('create')
const formSaving = ref(false)
const editingId  = ref('')
const form = reactive({ name: '', type: 'ec', region: '', description: '', status: 1 })

function openCreate() {
  formMode.value = 'create'; editingId.value = ''
  Object.assign(form, { name: '', type: 'ec', region: '', description: '', status: 1 })
  showFormModal.value = true
}
function openEdit(row: any) {
  formMode.value = 'edit'; editingId.value = row.id
  Object.assign(form, {
    name: row.name,
    type: row._raw?.type || 'ec',
    region: row.region,
    description: row._raw?.description || '',
    status: row._raw?.status || 1,
  })
  showFormModal.value = true
}
async function saveForm() {
  if (!form.name.trim()) { uni.showToast({ title: '联盟名称不能为空', icon: 'none' }); return }
  formSaving.value = true
  try {
    const payload = { name: form.name, type: form.type, region: form.region, description: form.description, status: form.status }
    if (formMode.value === 'create') {
      await createOrg(payload as any)
      uni.showToast({ title: '联盟创建成功', icon: 'success' })
    } else {
      await updateOrg(editingId.value, payload as any)
      uni.showToast({ title: '更新成功', icon: 'success' })
    }
    showFormModal.value = false; loadList()
  } catch (e: any) {
    uni.showToast({ title: e?.message || '保存失败', icon: 'none' })
  } finally { formSaving.value = false }
}

// ══════════════════════════════════════════
// 成员管理 Modal
// ══════════════════════════════════════════
const showMembersModal = ref(false)
const memberOrg        = ref<any>(null)
const members          = ref<any[]>([])
const memberLoading    = ref(false)
const addMemberUserId  = ref('')
const addMemberRole    = ref('member')

async function openMembers(row: any) {
  memberOrg.value = row; showMembersModal.value = true
  memberLoading.value = true
  try {
    const res: any = await get(`/orgs/${row.id}/members`)
    members.value = res.list || res || []
  } catch { members.value = [] }
  finally { memberLoading.value = false }
}
async function addMember() {
  if (!addMemberUserId.value) { uni.showToast({ title: '请输入用户ID', icon: 'none' }); return }
  try {
    await post(`/orgs/${memberOrg.value.id}/members`, { user_id: Number(addMemberUserId.value), role: addMemberRole.value })
    uni.showToast({ title: '成员添加成功', icon: 'success' })
    addMemberUserId.value = ''; openMembers(memberOrg.value)
  } catch (e: any) { uni.showToast({ title: e?.message || '添加失败', icon: 'none' }) }
}
async function removeMember(userId: number) {
  try {
    await del(`/orgs/${memberOrg.value.id}/members/${userId}`)
    uni.showToast({ title: '已移除成员', icon: 'success' })
    openMembers(memberOrg.value)
  } catch (e: any) { uni.showToast({ title: e?.message || '移除失败', icon: 'none' }) }
}

// ══════════════════════════════════════════
// 更多菜单 (···)
// ══════════════════════════════════════════
const moreMenuRow   = ref<any>(null)
const moreMenuStyle = ref('')
function openMoreMenu(e: MouseEvent, row: any) {
  moreMenuRow.value = moreMenuRow.value?.id === row.id ? null : row
  const rect = (e.currentTarget as HTMLElement).getBoundingClientRect()
  moreMenuStyle.value = `top:${rect.bottom + 4}px;right:${window.innerWidth - rect.right}px`
}
function closeMoreMenu() { moreMenuRow.value = null }

// ══════════════════════════════════════════
// 权限班子 Modal
// ══════════════════════════════════════════
const showTeamModal   = ref(false)
const teamOrg         = ref<any>(null)
const team            = ref<any[]>([])
const teamLoading     = ref(false)
const allAdmins       = ref<any[]>([])
const allRoles        = ref<any[]>([])
const addTeamAdminId  = ref<number|null>(null)
const addTeamRoleId   = ref<number>(4)
const addTeamRemark   = ref('')
const teamSaving      = ref(false)

async function openTeam(row: any) {
  teamOrg.value = row
  showTeamModal.value = true
  teamLoading.value = true
  try {
    const [t, admins, roles]: any[] = await Promise.all([
      get(`/orgs/${row.id}/team`),
      get('/admins', { page: 1, page_size: 100 }),
      get('/roles'),
    ])
    team.value      = Array.isArray(t) ? t : (t.list || [])
    allAdmins.value = admins?.list || []
    allRoles.value  = Array.isArray(roles) ? roles : (roles?.list || roles || [])
    addTeamAdminId.value = null
    addTeamRoleId.value  = allRoles.value.find((r: any) => r.code === 'operator')?.id || 4
    addTeamRemark.value  = ''
  } catch { team.value = [] }
  finally { teamLoading.value = false }
}

async function addTeamMember() {
  if (!addTeamAdminId.value) { uni.showToast({ title: '请选择管理员', icon: 'none' }); return }
  teamSaving.value = true
  try {
    await post(`/orgs/${teamOrg.value.id}/team`, {
      admin_id: addTeamAdminId.value, role_id: addTeamRoleId.value, remark: addTeamRemark.value
    })
    uni.showToast({ title: '已加入权限班子', icon: 'success' })
    openTeam(teamOrg.value)
  } catch (e: any) { uni.showToast({ title: e?.message || '添加失败', icon: 'none' }) }
  finally { teamSaving.value = false }
}

async function updateTeamRole(m: any, newRoleId: number) {
  try {
    await put(`/orgs/${teamOrg.value.id}/team/${m.admin_id}`, { role_id: newRoleId })
    m.role_id = newRoleId
    m.role = allRoles.value.find((r: any) => r.id === newRoleId)
    uni.showToast({ title: '角色已更新', icon: 'success' })
  } catch (e: any) { uni.showToast({ title: e?.message || '更新失败', icon: 'none' }) }
}

async function removeTeamMember(m: any) {
  uni.showModal({ title: '确认移除', content: `确认将「${m.admin?.username}」从权限班子中移除？`,
    success: async (res) => {
      if (!res.confirm) return
      try {
        await del(`/orgs/${teamOrg.value.id}/team/${m.admin_id}`)
        uni.showToast({ title: '已移除', icon: 'success' })
        openTeam(teamOrg.value)
      } catch (e: any) { uni.showToast({ title: e?.message || '移除失败', icon: 'none' }) }
    }
  })
}

const roleColors: Record<string, string> = {
  superadmin: '#ef4444', org_admin: '#3b82f6', finance: '#10b981', operator: '#f59e0b'
}

async function handleDelete(row: any) {
  closeMoreMenu()
  uni.showModal({ title: '确认删除', content: `确认删除联盟「${row.name}」？此操作不可逆。`, success: async (res) => {
    if (!res.confirm) return
    try { await deleteOrg(row.id); uni.showToast({ title: '已删除', icon: 'success' }); loadList() }
    catch (e: any) { uni.showToast({ title: e?.message || '删除失败', icon: 'none' }) }
  }})
}

onMounted(loadList)
</script>

<template>
  <AppLayout :breadcrumbs="breadcrumbs">
    <!-- 统计栏 -->
    <view class="stats-row">
      <KpiCard icon="🏢" label="联盟总数"  :value="pageStats.total"    :trend="{ dir:'up', text:'5.1%' }"    icon-bg="#eff6ff" />
      <KpiCard icon="✅" label="正常运营"  :value="pageStats.active"   :trend="{ dir:'up', text:'本月+6' }"  icon-bg="#f0fdf4" />
      <KpiCard icon="⏳" label="待审核"    :value="pageStats.pending"  :trend="{ dir:'up', text:'今日+3' }"  icon-bg="#fffbeb" />
      <KpiCard icon="❄️" label="已冻结"    :value="pageStats.frozen"   :trend="{ dir:'down', text:'无变化' }" icon-bg="#fef2f2" />
      <KpiCard icon="🆕" label="今日新增"  :value="pageStats.todayNew" :trend="{ dir:'up', text:'+3' }"      icon-bg="#faf5ff" />
    </view>

    <FilterPanel :fields="filterFields" :quick-tags="quickTags"
      @search="onSearch" @reset="() => { filterParams.value = {}; loadList() }" @export="() => {}">
      <template #extra-actions>
        <view v-if="userStore.hasPermission('org:create')" class="fp-btn fp-btn-primary" @click="openCreate">＋ 新增联盟</view>
      </template>
    </FilterPanel>

    <view class="card table-card" @click="closeMoreMenu">
      <view class="table-toolbar">
        <text class="sel-info">共 <text class="em">{{ total }}</text> 条联盟</text>
        <view class="icon-btn" @click="loadList">🔄</view>
      </view>
      <view class="t-head">
        <text class="th" style="flex:2">联盟名称</text>
        <text class="th" style="flex:0.8">类型</text>
        <text class="th" style="flex:0.8">状态</text>
        <text class="th" style="flex:0.9">地区</text>
        <text class="th" style="flex:0.6">成员数</text>
        <text class="th" style="flex:0.8">负责人</text>
        <text class="th" style="flex:1">成立时间</text>
        <text class="th" style="flex:1.3">操作</text>
      </view>
      <view v-if="loading" class="table-empty">加载中...</view>
      <view v-else>
        <view v-for="row in list" :key="row.id" class="t-row">
          <text class="td org-name" style="flex:2">{{ row.name }}</text>
          <view class="td" style="flex:0.8"><text class="org-tag">{{ row.type }}</text></view>
          <view class="td" style="flex:0.8">
            <StatusBadge :status="row.status==='active'?'success':row.status==='pending'?'warning':'danger'"
              :label="row.status==='active'?'正常':row.status==='pending'?'待审核':'已冻结'" />
          </view>
          <text class="td t-muted" style="flex:0.9">{{ row.region || '—' }}</text>
          <text class="td" style="flex:0.6">{{ row.memberCount }}</text>
          <text class="td t-muted" style="flex:0.8">{{ row.leader || '—' }}</text>
          <text class="td t-muted" style="flex:1">{{ row.createdAt.slice(0,10) }}</text>
          <view class="td action-btns" style="flex:1.3">
            <IconBtn icon="👥" tip="成员管理" type="view"    @click="openMembers(row)" />
            <IconBtn icon="🔐" tip="权限班子" type="team"    @click="openTeam(row)" />
            <IconBtn v-if="userStore.hasPermission('org:update')" icon="✏️" tip="编辑联盟" type="edit" @click="openEdit(row)" />
            <IconBtn icon="⋯"  tip="更多操作" type="default" @click="(e:any) => openMoreMenu(e as MouseEvent, row)" />
          </view>
        </view>
        <view v-if="!list.length && !loading" class="table-empty">暂无数据</view>
      </view>
      <Pagination :total="total" :page="page" :page-size="pageSize"
        @page-change="onPageChange" @page-size-change="onPageSizeChange" />
    </view>

    <!-- 更多菜单 -->
    <!-- #ifdef H5 -->
    <view v-if="moreMenuRow" class="more-menu" :style="moreMenuStyle" @click.stop>
      <view v-if="userStore.hasPermission('org:update')" class="mm-item" @click="openEdit(moreMenuRow); closeMoreMenu()">✏️ 编辑联盟</view>
      <view class="mm-item" @click="openMembers(moreMenuRow); closeMoreMenu()">👥 成员管理</view>
      <view class="mm-item" @click="openTeam(moreMenuRow); closeMoreMenu()">🔐 权限班子</view>
      <view class="mm-divider"/>
      <view v-if="userStore.hasPermission('org:delete')" class="mm-item text-danger" @click="handleDelete(moreMenuRow)">🗑 删除联盟</view>
    </view>

    <!-- 新增/编辑 Modal -->
    <view v-if="showFormModal" class="modal-mask" @click.self="showFormModal = false">
      <view class="modal-box">
        <view class="modal-header">
          <text class="modal-title">{{ formMode==='create' ? '新增联盟' : '编辑联盟' }}</text>
          <text class="modal-close" @click="showFormModal = false">✕</text>
        </view>
        <view class="modal-body">
          <view class="form-row">
            <text class="form-label">联盟名称 <text class="req">*</text></text>
            <input class="form-input" v-model="form.name" placeholder="请输入联盟名称"/>
          </view>
          <view class="form-row">
            <text class="form-label">联盟类型</text>
            <view class="form-radio-group">
              <view v-for="opt in typeOptions" :key="opt.value"
                class="form-radio" :class="{ active: form.type === opt.value }"
                @click="form.type = opt.value">{{ opt.label }}</view>
            </view>
          </view>
          <view class="form-row">
            <text class="form-label">所在地区</text>
            <input class="form-input" v-model="form.region" placeholder="省/市，如：北京市"/>
          </view>
          <view class="form-row">
            <text class="form-label">描述</text>
            <textarea class="form-textarea" v-model="form.description" placeholder="联盟简介（选填）"/>
          </view>
          <view class="form-row">
            <text class="form-label">状态</text>
            <view class="form-radio-group">
              <view class="form-radio" :class="{ active: form.status===1 }" @click="form.status=1">✓ 正常</view>
              <view class="form-radio" :class="{ active: form.status===2 }" @click="form.status=2">⏳ 待审核</view>
              <view class="form-radio" :class="{ active: form.status===3 }" @click="form.status=3">❄️ 冻结</view>
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

    <!-- 成员管理 Modal -->
    <view v-if="showMembersModal" class="modal-mask" @click.self="showMembersModal = false">
      <view class="modal-box modal-lg">
        <view class="modal-header">
          <text class="modal-title">成员管理 · {{ memberOrg?.name }}</text>
          <text class="modal-close" @click="showMembersModal = false">✕</text>
        </view>
        <view class="modal-body">
          <!-- 添加成员 -->
          <view class="add-member-bar">
            <input class="form-input" style="flex:1" v-model="addMemberUserId" placeholder="输入用户 ID"/>
            <select class="form-select" v-model="addMemberRole">
              <option value="member">普通成员</option>
              <option value="admin">管理员</option>
              <option value="finance">财务</option>
            </select>
            <view class="m-btn m-btn-primary" style="white-space:nowrap" @click="addMember">＋ 添加</view>
          </view>
          <!-- 成员列表 -->
          <view v-if="memberLoading" class="table-empty">加载中...</view>
          <view v-else>
            <view class="mem-head">
              <text style="flex:0.5">ID</text>
              <text style="flex:1.5">用户名</text>
              <text style="flex:1">角色</text>
              <text style="flex:1">加入时间</text>
              <text style="flex:0.8">操作</text>
            </view>
            <view v-for="m in members" :key="m.user_id || m.id" class="mem-row">
              <text class="t-muted" style="flex:0.5">{{ m.user_id || m.id }}</text>
              <text style="flex:1.5">{{ m.user?.username || m.username || '—' }}</text>
              <text class="t-muted" style="flex:1">{{ m.role || '成员' }}</text>
              <text class="t-muted" style="flex:1">{{ m.created_at?.slice(0,10) || '—' }}</text>
              <view style="flex:0.8">
                <IconBtn icon="✕" tip="移除成员" type="danger" size="sm" @click="removeMember(m.user_id || m.id)" />
              </view>
            </view>
            <view v-if="!members.length" class="table-empty" style="padding:20px">暂无成员</view>
          </view>
        </view>
        <view class="modal-footer">
          <view class="m-btn m-btn-cancel" @click="showMembersModal = false">关闭</view>
        </view>
      </view>
    </view>
    <!-- 权限班子 Modal -->
    <view v-if="showTeamModal" class="modal-mask" @click.self="showTeamModal = false">
      <view class="modal-box modal-xl">
        <view class="modal-header">
          <view>
            <text class="modal-title">🔐 权限班子 · {{ teamOrg?.name }}</text>
            <text class="modal-sub">为该联盟指定专属管理员，班子成员可对本联盟进行操作管理</text>
          </view>
          <text class="modal-close" @click="showTeamModal = false">✕</text>
        </view>
        <view class="modal-body">

          <!-- 添加成员栏 -->
          <view v-if="userStore.hasPermission('org:update')" class="team-add-bar">
            <text class="team-add-title">添加成员</text>
            <view class="team-add-row">
              <select class="form-select flex1" v-model="addTeamAdminId">
                <option :value="null">-- 选择管理员 --</option>
                <option v-for="a in allAdmins" :key="a.id" :value="a.id">
                  {{ a.username }}{{ a.real_name ? ' (' + a.real_name + ')' : '' }}
                </option>
              </select>
              <select class="form-select" v-model="addTeamRoleId">
                <option v-for="r in allRoles.filter((r:any) => r.code !== 'superadmin')" :key="r.id" :value="r.id">
                  {{ r.name }}
                </option>
              </select>
              <input class="form-input" style="flex:1.5" v-model="addTeamRemark" placeholder="备注（选填）"/>
              <view class="m-btn m-btn-primary" :class="{loading: teamSaving}" @click="addTeamMember">
                {{ teamSaving ? '添加中...' : '＋ 加入班子' }}
              </view>
            </view>
          </view>

          <!-- 成员列表 -->
          <view v-if="teamLoading" class="table-empty">加载中...</view>
          <view v-else>
            <view class="team-empty" v-if="!team.length">
              <text class="team-empty-icon">👥</text>
              <text class="team-empty-text">该联盟暂无权限班子，请添加管理员</text>
            </view>
            <view v-else>
              <view class="team-head">
                <text style="flex:1.5">管理员账号</text>
                <text style="flex:1">当前角色</text>
                <text style="flex:1">调整角色</text>
                <text style="flex:1.5">备注</text>
                <text style="flex:0.8">操作</text>
              </view>
              <view v-for="m in team" :key="m.id" class="team-row">
                <!-- 头像 + 账号 -->
                <view style="flex:1.5;display:flex;align-items:center;gap:8px">
                  <view class="t-avatar"
                    :style="{ background: roleColors[m.admin?.role_id] || '#6366f1' }">
                    {{ m.admin?.username?.charAt(0)?.toUpperCase() || '?' }}
                  </view>
                  <view>
                    <text class="t-name">{{ m.admin?.username || '—' }}</text>
                    <text class="t-email">{{ m.admin?.real_name || m.admin?.email || '' }}</text>
                  </view>
                </view>
                <!-- 当前角色 -->
                <view style="flex:1">
                  <view class="role-badge"
                    :style="{ background: (roleColors[m.role?.code]||'#6366f1')+'22', color: roleColors[m.role?.code]||'#6366f1' }">
                    {{ m.role?.name || '—' }}
                  </view>
                </view>
                <!-- 调整角色 -->
                <view style="flex:1">
                  <select class="form-select-sm"
                    :value="m.role_id"
                    @change="(e:any) => updateTeamRole(m, Number(e.target.value))">
                    <option v-for="r in allRoles.filter((r:any) => r.code !== 'superadmin')" :key="r.id" :value="r.id">
                      {{ r.name }}
                    </option>
                  </select>
                </view>
                <!-- 备注 -->
                <text class="t-muted" style="flex:1.5">{{ m.remark || '—' }}</text>
                <!-- 操作 -->
                <view style="flex:0.8">
                  <IconBtn v-if="userStore.hasPermission('org:update')"
                    icon="✕" tip="移除班子成员" type="danger" size="sm" @click="removeTeamMember(m)" />
                </view>
              </view>
            </view>
          </view>

          <!-- 权限说明 -->
          <view class="team-tips">
            <text class="tips-title">📌 权限班子说明</text>
            <text class="tips-item">· 权限班子成员可对本联盟进行编辑、成员管理等操作，无需全局权限</text>
            <text class="tips-item">· 角色决定成员在本联盟内的权限范围（与全局RBAC角色定义一致）</text>
            <text class="tips-item">· superadmin 无需加入班子，默认拥有所有联盟的管理权限</text>
          </view>
        </view>
        <view class="modal-footer">
          <text class="footer-info">共 {{ team.length }} 位班子成员</text>
          <view class="m-btn m-btn-cancel" @click="showTeamModal = false">关闭</view>
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
.sel-info { font-size: 13px; color: var(--color-text-secondary); }
.em { color: var(--color-primary); font-weight: 600; }
.fp-btn { height: 34px; border-radius: 7px; border: none; cursor: pointer; font-size: 13px; font-weight: 500; display: flex; align-items: center; padding: 0 16px; }
.fp-btn-primary { background: var(--color-primary); color: #fff; }
.icon-btn { width: 32px; height: 32px; border-radius: 7px; border: 1px solid var(--color-border); background: var(--color-card-bg); display: flex; align-items: center; justify-content: center; font-size: 14px; cursor: pointer; }
.t-head { display: flex; padding: 10px 16px; background: var(--color-border-light); border-bottom: 1px solid var(--color-border); }
.th { font-size: 12px; font-weight: 600; color: var(--color-text-secondary); padding-right: 8px; }
.t-row { display: flex; align-items: center; padding: 11px 16px; border-bottom: 1px solid var(--color-border-light); &:hover { background: var(--color-border-light); } &:last-child { border: none; } }
.td { font-size: 13px; color: var(--color-text-primary); padding-right: 8px; }
.t-muted { font-size: 12px; color: var(--color-text-muted); }
.org-name { font-weight: 500; }
.org-tag { background: var(--color-border-light); color: var(--color-text-primary); padding: 2px 8px; border-radius: 4px; font-size: 11px; }
.action-btns { display: flex; gap: 6px; }
.act-btn { font-size: 12px; padding: 4px 10px; border-radius: 5px; cursor: pointer; font-weight: 500; }
.act-view { background: var(--color-primary-light); color: var(--color-primary); }
.act-edit { background: #f0fdf4; color: #16a34a; }
.act-more { background: var(--color-border-light); color: var(--color-text-secondary); letter-spacing: 2px; }
.act-danger { background: #fef2f2; color: #ef4444; }
.table-empty { padding: 40px; text-align: center; color: var(--color-text-muted); font-size: 13px; }
.more-menu { position: fixed; z-index: 200; background: var(--color-card-bg); border: 1px solid var(--color-border); border-radius: 10px; box-shadow: 0 8px 24px rgba(0,0,0,0.12); padding: 6px 0; min-width: 140px; }
.mm-item { padding: 9px 16px; font-size: 13px; cursor: pointer; color: var(--color-text-primary); &:hover { background: var(--color-border-light); } }
.mm-divider { height: 1px; background: var(--color-border-light); margin: 4px 0; }
.text-danger { color: #ef4444; }
.modal-mask { position: fixed; inset: 0; background: rgba(0,0,0,0.45); z-index: 300; display: flex; align-items: center; justify-content: center; }
.modal-box { background: var(--color-card-bg); border-radius: 16px; width: 500px; max-width: 92vw; box-shadow: 0 20px 60px rgba(0,0,0,0.2); overflow: hidden; }
.modal-lg { width: 640px; }
.modal-header { display: flex; align-items: center; justify-content: space-between; padding: 18px 24px; border-bottom: 1px solid var(--color-border-light); }
.modal-title { font-size: 16px; font-weight: 700; color: var(--color-text-primary); }
.modal-close { font-size: 18px; color: var(--color-text-muted); cursor: pointer; &:hover { color: var(--color-text-primary); } }
.modal-body { padding: 20px 24px; max-height: 65vh; overflow-y: auto; }
.modal-footer { display: flex; justify-content: flex-end; gap: 10px; padding: 16px 24px; border-top: 1px solid var(--color-border-light); }
.form-row { margin-bottom: 16px; }
.form-label { font-size: 13px; font-weight: 500; color: var(--color-text-secondary); display: block; margin-bottom: 6px; }
.req { color: #ef4444; }
.form-input { width: 100%; height: 38px; border: 1.5px solid var(--color-border); border-radius: 8px; padding: 0 12px; font-size: 13px; color: var(--color-text-primary); background: var(--color-card-bg); box-sizing: border-box; &:focus { border-color: var(--color-primary); outline: none; } }
.form-textarea { width: 100%; height: 80px; border: 1.5px solid var(--color-border); border-radius: 8px; padding: 10px 12px; font-size: 13px; color: var(--color-text-primary); background: var(--color-card-bg); box-sizing: border-box; resize: none; }
.form-select { height: 38px; border: 1.5px solid var(--color-border); border-radius: 8px; padding: 0 10px; font-size: 13px; background: var(--color-card-bg); color: var(--color-text-primary); }
.form-radio-group { display: flex; flex-wrap: wrap; gap: 8px; }
.form-radio { padding: 6px 14px; border-radius: 7px; font-size: 12px; cursor: pointer; border: 1.5px solid var(--color-border); color: var(--color-text-secondary); &.active { border-color: var(--color-primary); background: var(--color-primary-light); color: var(--color-primary); font-weight: 600; } }
.m-btn { height: 36px; border-radius: 8px; font-size: 13px; font-weight: 500; display: flex; align-items: center; padding: 0 20px; cursor: pointer; }
.m-btn-cancel  { background: var(--color-border-light); color: var(--color-text-secondary); }
.m-btn-primary { background: var(--color-primary); color: #fff; &.loading { opacity: 0.7; pointer-events: none; } }
.add-member-bar { display: flex; gap: 10px; align-items: center; margin-bottom: 16px; padding-bottom: 16px; border-bottom: 1px solid var(--color-border-light); }
.mem-head { display: flex; padding: 8px 0; border-bottom: 1px solid var(--color-border-light); font-size: 12px; font-weight: 600; color: var(--color-text-secondary); }
.mem-row { display: flex; align-items: center; padding: 10px 0; border-bottom: 1px solid var(--color-border-light); font-size: 13px; color: var(--color-text-primary); &:last-child { border: none; } }
/* 权限班子 */
.modal-xl { width: 760px; }
.modal-sub { display: block; font-size: 12px; color: var(--color-text-muted); margin-top: 3px; }
.act-team { background: #f5f3ff; color: #7c3aed; }
.team-add-bar { background: var(--color-border-light); border-radius: 10px; padding: 14px 16px; margin-bottom: 20px; }
.team-add-title { font-size: 13px; font-weight: 600; color: var(--color-text-secondary); display: block; margin-bottom: 10px; }
.team-add-row { display: flex; gap: 10px; align-items: center; }
.flex1 { flex: 1; }
.team-head { display: flex; padding: 9px 4px; border-bottom: 1px solid var(--color-border-light); font-size: 12px; font-weight: 600; color: var(--color-text-secondary); }
.team-row { display: flex; align-items: center; padding: 12px 4px; border-bottom: 1px solid var(--color-border-light); font-size: 13px; &:last-child { border: none; } }
.t-avatar { width: 32px; height: 32px; border-radius: 50%; display: flex; align-items: center; justify-content: center; color: #fff; font-size: 13px; font-weight: bold; flex-shrink: 0; }
.t-name { font-size: 13px; font-weight: 600; display: block; }
.t-email { font-size: 11px; color: var(--color-text-muted); display: block; }
.role-badge { display: inline-block; padding: 3px 10px; border-radius: 5px; font-size: 11px; font-weight: 600; }
.form-select-sm { height: 32px; border: 1.5px solid var(--color-border); border-radius: 6px; padding: 0 8px; font-size: 12px; background: var(--color-card-bg); color: var(--color-text-primary); width: 100%; }
.team-empty { padding: 40px; text-align: center; }
.team-empty-icon { font-size: 32px; display: block; margin-bottom: 8px; }
.team-empty-text { font-size: 13px; color: var(--color-text-muted); display: block; }
.team-tips { margin-top: 20px; background: #eff6ff; border-radius: 8px; padding: 12px 16px; border-left: 3px solid #3b82f6; }
.tips-title { font-size: 13px; font-weight: 600; color: #1e40af; display: block; margin-bottom: 6px; }
.tips-item { font-size: 12px; color: #3b82f6; display: block; line-height: 1.8; }
.footer-info { font-size: 13px; color: var(--color-text-muted); flex: 1; }
</style>
