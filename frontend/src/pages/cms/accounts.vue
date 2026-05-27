<script setup lang="ts">
import { ref, onMounted } from 'vue'
import SvgIcon from '@/components/common/SvgIcon.vue'
import IconBtn from '@/components/common/IconBtn.vue'

const PLATFORM_LABELS: Record<string, { label: string; color: string }> = {
  wechat:  { label: '微信公众号', color: '#07c160' },
  rednote: { label: '小红书',    color: '#ff2442' },
  douyin:  { label: '抖音',      color: '#161823' },
  csdn:    { label: 'CSDN',      color: '#fc5531' },
}
const STATUS_MAP: Record<number, { label: string; color: string }> = {
  0: { label: '禁用',     color: '#999' },
  1: { label: '正常',     color: '#67c23a' },
  2: { label: '凭证失效', color: '#e6a23c' },
  3: { label: '已封禁',   color: '#f56c6c' },
}

const list     = ref<any[]>([])
const total    = ref(0)
const loading  = ref(false)
const page     = ref(1)
const pageSize = 10

const filter = ref({ platform: '', status: '' })
const showModal  = ref(false)
const showAudit  = ref(false)
const editMode   = ref(false)
const curAccount = ref<any>(null)
const auditLogs  = ref<any[]>([])
const saving     = ref(false)

const form = ref({
  platform: 'wechat',
  account_name: '',
  account_uid: '',
  cred_json: '',
  expires_at: '',
  remark: '',
})

const credPlaceholders: Record<string, string> = {
  wechat:  '{"app_id":"xxx","app_secret":"xxx","author_name":"xxx"}',
  rednote: '{"phone":"xxx","cookie":"..."}',
  douyin:  '{"client_key":"xxx","access_token":"..."}',
  csdn:    '{"cookie":"..."}',
}

async function load() {
  loading.value = true
  const params = new URLSearchParams({
    page: String(page.value), page_size: String(pageSize),
    ...(filter.value.platform ? { platform: filter.value.platform } : {}),
    ...(filter.value.status   ? { status:   filter.value.status }   : {}),
  })
  const token = uni.getStorageSync('token')
  const [err, res] = await uni.request({
    url: `http://localhost:8080/api/v1/cms/accounts?${params}`,
    header: { Authorization: `Bearer ${token}` },
  }) as any
  loading.value = false
  if (!err && res.data.code === 0) {
    list.value  = res.data.data.list  || []
    total.value = res.data.data.total || 0
  }
}

function openCreate() {
  editMode.value = false
  curAccount.value = null
  form.value = { platform: 'wechat', account_name: '', account_uid: '', cred_json: '', expires_at: '', remark: '' }
  showModal.value = true
}

function openEdit(acc: any) {
  editMode.value = true
  curAccount.value = acc
  form.value = {
    platform: acc.platform,
    account_name: acc.account_name,
    account_uid: acc.account_uid || '',
    cred_json: '',
    expires_at: acc.expires_at ? acc.expires_at.substring(0, 10) : '',
    remark: acc.remark || '',
  }
  showModal.value = true
}

async function openAudit(acc: any) {
  curAccount.value = acc
  const token = uni.getStorageSync('token')
  const [, res] = await uni.request({
    url: `http://localhost:8080/api/v1/cms/accounts/${acc.id}/audit-logs`,
    header: { Authorization: `Bearer ${token}` },
  }) as any
  auditLogs.value = res?.data?.data?.list || []
  showAudit.value = true
}

async function save() {
  if (!form.value.account_name) return uni.showToast({ title: '请填写账号名称', icon: 'none' })
  if (!editMode.value && !form.value.cred_json) return uni.showToast({ title: '请填写凭证JSON', icon: 'none' })

  saving.value = true
  const token = uni.getStorageSync('token')
  const payload: any = { ...form.value }
  if (editMode.value && !payload.cred_json) delete payload.cred_json

  const url = editMode.value
    ? `http://localhost:8080/api/v1/cms/accounts/${curAccount.value.id}`
    : 'http://localhost:8080/api/v1/cms/accounts'

  const [err, res] = await uni.request({
    url,
    method: editMode.value ? 'PUT' : 'POST',
    header: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
    data: payload,
  }) as any
  saving.value = false

  if (!err && res.data.code === 0) {
    uni.showToast({ title: editMode.value ? '更新成功' : '添加成功', icon: 'success' })
    showModal.value = false
    load()
  } else {
    uni.showToast({ title: res?.data?.message || '操作失败', icon: 'none' })
  }
}

async function del(acc: any) {
  const confirmed = await new Promise(r =>
    uni.showModal({ title: '确认删除', content: `删除账号「${acc.account_name}」？此操作不可恢复。`, success: (res) => r(res.confirm) })
  )
  if (!confirmed) return
  const token = uni.getStorageSync('token')
  await uni.request({
    url: `http://localhost:8080/api/v1/cms/accounts/${acc.id}`,
    method: 'DELETE',
    header: { Authorization: `Bearer ${token}` },
  })
  uni.showToast({ title: '已删除', icon: 'success' })
  load()
}

onMounted(load)
</script>

<template>
  <view class="page">
    <view class="page-header">
      <view class="page-title"><SvgIcon name="link" /> 平台账号管理</view>
      <view class="header-actions">
        <select class="filter-select" v-model="filter.platform" @change="load">
          <option value="">全部平台</option>
          <option v-for="(v,k) in PLATFORM_LABELS" :key="k" :value="k">{{ v.label }}</option>
        </select>
        <select class="filter-select" v-model="filter.status" @change="load">
          <option value="">全部状态</option>
          <option value="1">正常</option>
          <option value="2">凭证失效</option>
          <option value="0">禁用</option>
        </select>
        <button class="btn btn-primary" @click="openCreate"><SvgIcon name="add" /> 绑定账号</button>
      </view>
    </view>

    <view class="table-wrap">
      <view v-if="loading" class="loading-tip">加载中...</view>
      <view v-else-if="!list.length" class="empty-tip">暂无数据</view>
      <table v-else class="data-table">
        <thead>
          <tr>
            <th>平台</th><th>账号名称</th><th>平台UID</th>
            <th>凭证版本</th><th>状态</th><th>最后使用</th><th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="acc in list" :key="acc.id">
            <td>
              <span class="platform-tag" :style="{ background: PLATFORM_LABELS[acc.platform]?.color || '#888' }">
                {{ PLATFORM_LABELS[acc.platform]?.label || acc.platform }}
              </span>
            </td>
            <td>{{ acc.account_name }}</td>
            <td>{{ acc.account_uid || '-' }}</td>
            <td>v{{ acc.cred_version }}</td>
            <td>
              <span class="status-dot" :style="{ color: STATUS_MAP[acc.status]?.color }">
                ● {{ STATUS_MAP[acc.status]?.label || acc.status }}
              </span>
            </td>
            <td>{{ acc.last_used_at ? acc.last_used_at.substring(0,10) : '-' }}</td>
            <td class="action-cell">
              <IconBtn icon="edit"   tip="编辑"    @click="openEdit(acc)" />
              <IconBtn icon="shield" tip="审计日志" @click="openAudit(acc)" />
              <IconBtn icon="delete" tip="删除"    danger @click="del(acc)" />
            </td>
          </tr>
        </tbody>
      </table>
    </view>

    <!-- 分页 -->
    <view class="pagination" v-if="total > pageSize">
      <button class="page-btn" :disabled="page <= 1" @click="page--; load()">上一页</button>
      <text class="page-info">{{ page }} / {{ Math.ceil(total / pageSize) }}</text>
      <button class="page-btn" :disabled="page * pageSize >= total" @click="page++; load()">下一页</button>
    </view>

    <!-- 新增/编辑弹窗 -->
    <view v-if="showModal" class="modal-mask" @click.self="showModal=false">
      <view class="modal-box">
        <view class="modal-header">
          <text>{{ editMode ? '编辑账号' : '绑定新账号' }}</text>
          <view @click="showModal=false"><SvgIcon name="close" /></view>
        </view>
        <view class="modal-body">
          <view class="form-row">
            <label>平台</label>
            <select v-model="form.platform" :disabled="editMode" class="form-select">
              <option v-for="(v,k) in PLATFORM_LABELS" :key="k" :value="k">{{ v.label }}</option>
            </select>
          </view>
          <view class="form-row">
            <label>账号名称 <span class="required">*</span></label>
            <input v-model="form.account_name" placeholder="公众号/账号显示名称" class="form-input" />
          </view>
          <view class="form-row">
            <label>平台UID</label>
            <input v-model="form.account_uid" placeholder="平台唯一ID（可选）" class="form-input" />
          </view>
          <view class="form-row">
            <label>凭证 JSON {{ editMode ? '（留空不更新）' : '' }} <span class="required" v-if="!editMode">*</span></label>
            <textarea v-model="form.cred_json" class="form-textarea"
              :placeholder="credPlaceholders[form.platform]" rows="4" />
            <text class="form-hint">凭证将使用 AES-256-GCM 加密存储，请勿在备注中填写明文</text>
          </view>
          <view class="form-row">
            <label>过期时间</label>
            <input v-model="form.expires_at" type="date" class="form-input" />
          </view>
          <view class="form-row">
            <label>备注</label>
            <input v-model="form.remark" placeholder="备注（可选）" class="form-input" />
          </view>
        </view>
        <view class="modal-footer">
          <button class="btn" @click="showModal=false">取消</button>
          <button class="btn btn-primary" :disabled="saving" @click="save">
            {{ saving ? '保存中...' : '保存' }}
          </button>
        </view>
      </view>
    </view>

    <!-- 审计日志弹窗 -->
    <view v-if="showAudit" class="modal-mask" @click.self="showAudit=false">
      <view class="modal-box modal-lg">
        <view class="modal-header">
          <text>凭证审计日志 — {{ curAccount?.account_name }}</text>
          <view @click="showAudit=false"><SvgIcon name="close" /></view>
        </view>
        <view class="modal-body">
          <view v-if="!auditLogs.length" class="empty-tip">暂无审计记录</view>
          <table v-else class="data-table">
            <thead><tr><th>时间</th><th>操作</th><th>原因</th><th>IP</th></tr></thead>
            <tbody>
              <tr v-for="log in auditLogs" :key="log.id">
                <td>{{ log.created_at?.substring(0,19) }}</td>
                <td><span class="action-tag">{{ log.action }}</span></td>
                <td>{{ log.reason || '-' }}</td>
                <td>{{ log.ip || '-' }}</td>
              </tr>
            </tbody>
          </table>
        </view>
      </view>
    </view>
  </view>
</template>

<style scoped>
.page { padding: 24px; background: var(--color-bg); min-height: 100vh; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.page-title { font-size: 18px; font-weight: 600; color: var(--color-text); display: flex; align-items: center; gap: 8px; }
.header-actions { display: flex; gap: 10px; align-items: center; }
.filter-select { height: 34px; border: 1px solid var(--color-border); border-radius: 6px; padding: 0 10px; background: var(--color-surface); color: var(--color-text); font-size: 13px; }
.btn { height: 34px; padding: 0 16px; border-radius: 6px; border: 1px solid var(--color-border); background: var(--color-surface); color: var(--color-text); cursor: pointer; font-size: 13px; display: inline-flex; align-items: center; gap: 6px; }
.btn-primary { background: var(--color-primary); color: #fff; border-color: var(--color-primary); }
.table-wrap { background: var(--color-surface); border-radius: 8px; overflow: auto; }
.data-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.data-table th { background: var(--color-bg); padding: 10px 14px; text-align: left; color: var(--color-text-muted); font-weight: 500; white-space: nowrap; }
.data-table td { padding: 10px 14px; border-top: 1px solid var(--color-border); color: var(--color-text); }
.data-table tr:hover td { background: var(--color-bg); }
.platform-tag { padding: 2px 8px; border-radius: 12px; color: #fff; font-size: 12px; white-space: nowrap; }
.status-dot { font-size: 13px; }
.action-cell { display: flex; gap: 6px; }
.action-tag { background: var(--color-primary-light, #e8f4ff); color: var(--color-primary); padding: 1px 8px; border-radius: 4px; font-size: 12px; }
.loading-tip, .empty-tip { padding: 40px; text-align: center; color: var(--color-text-muted); }
.pagination { display: flex; justify-content: center; align-items: center; gap: 12px; margin-top: 16px; }
.page-btn { height: 30px; padding: 0 14px; border-radius: 6px; border: 1px solid var(--color-border); background: var(--color-surface); cursor: pointer; color: var(--color-text); font-size: 13px; }
.page-info { color: var(--color-text-muted); font-size: 13px; }
.modal-mask { position: fixed; inset: 0; background: rgba(0,0,0,.45); display: flex; align-items: center; justify-content: center; z-index: 1000; }
.modal-box { background: var(--color-surface); border-radius: 10px; width: 520px; max-height: 80vh; overflow-y: auto; }
.modal-lg { width: 680px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--color-border); font-weight: 600; color: var(--color-text); }
.modal-body { padding: 20px; }
.modal-footer { display: flex; justify-content: flex-end; gap: 10px; padding: 14px 20px; border-top: 1px solid var(--color-border); }
.form-row { margin-bottom: 16px; }
.form-row label { display: block; font-size: 13px; color: var(--color-text-muted); margin-bottom: 6px; }
.required { color: #f56c6c; }
.form-input, .form-select, .form-textarea { width: 100%; height: 36px; border: 1px solid var(--color-border); border-radius: 6px; padding: 0 10px; background: var(--color-bg); color: var(--color-text); font-size: 13px; box-sizing: border-box; }
.form-textarea { height: auto; padding: 8px 10px; resize: vertical; }
.form-hint { font-size: 12px; color: var(--color-text-muted); margin-top: 4px; display: block; }
</style>
