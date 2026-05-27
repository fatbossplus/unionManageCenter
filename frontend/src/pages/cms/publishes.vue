<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import SvgIcon from '@/components/common/SvgIcon.vue'
import IconBtn from '@/components/common/IconBtn.vue'

const PLATFORM_LABELS: Record<string, string> = {
  wechat: '微信公众号', rednote: '小红书', douyin: '抖音', csdn: 'CSDN',
}
const STATUS_MAP: Record<string, { label: string; color: string }> = {
  draft:     { label: '草稿',   color: '#909399' },
  reviewing: { label: '审核中', color: '#e6a23c' },
  approved:  { label: '已审核', color: '#67c23a' },
  scheduled: { label: '定时发布', color: '#409eff' },
  published: { label: '已发布', color: '#27ae60' },
  failed:    { label: '发布失败', color: '#f56c6c' },
}

const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 10
const loading = ref(false)
const publishing = ref<Record<number, boolean>>({})

const filterPlatform = ref('')
const filterStatus   = ref('')

const showDetail = ref(false)
const cur = ref<any>(null)
const accounts = ref<any[]>([])
const rejectReason = ref('')
const showRejectModal = ref(false)
const rejectTarget = ref<any>(null)

// 统计数据
const stats = ref<any[]>([])
const statsByStatus = computed(() => {
  const m: Record<string, number> = {}
  stats.value.forEach(s => { m[s.status] = (m[s.status] || 0) + s.count })
  return m
})

async function loadStats() {
  const token = uni.getStorageSync('token')
  const [, r] = await uni.request({ url: 'http://localhost:8080/api/v1/cms/publishes/stats', header: { Authorization: `Bearer ${token}` } }) as any
  stats.value = r?.data?.data || []
}

async function load() {
  loading.value = true
  const token = uni.getStorageSync('token')
  const params = new URLSearchParams({
    page: String(page.value), page_size: String(pageSize),
    ...(filterPlatform.value ? { platform: filterPlatform.value } : {}),
    ...(filterStatus.value   ? { status:   filterStatus.value }   : {}),
  })
  const [, r] = await uni.request({ url: `http://localhost:8080/api/v1/cms/publishes?${params}`, header: { Authorization: `Bearer ${token}` } }) as any
  loading.value = false
  list.value  = r?.data?.data?.list  || []
  total.value = r?.data?.data?.total || 0
}

async function loadAccounts() {
  const token = uni.getStorageSync('token')
  const [, r] = await uni.request({ url: 'http://localhost:8080/api/v1/cms/accounts?page_size=100', header: { Authorization: `Bearer ${token}` } }) as any
  accounts.value = r?.data?.data?.list || []
}

async function openDetail(item: any) {
  const token = uni.getStorageSync('token')
  const [, r] = await uni.request({ url: `http://localhost:8080/api/v1/cms/publishes/${item.id}`, header: { Authorization: `Bearer ${token}` } }) as any
  cur.value = r?.data?.data || item
  showDetail.value = true
}

async function approve(item: any) {
  const token = uni.getStorageSync('token')
  const [, r] = await uni.request({
    url: `http://localhost:8080/api/v1/cms/publishes/${item.id}/approve`,
    method: 'POST', header: { Authorization: `Bearer ${token}` },
  }) as any
  if (r?.data?.code === 0) { uni.showToast({ title: '审核通过', icon: 'success' }); load(); loadStats() }
}

function openReject(item: any) {
  rejectTarget.value = item; rejectReason.value = ''; showRejectModal.value = true
}

async function confirmReject() {
  if (!rejectReason.value) return uni.showToast({ title: '请填写拒绝原因', icon: 'none' })
  const token = uni.getStorageSync('token')
  await uni.request({
    url: `http://localhost:8080/api/v1/cms/publishes/${rejectTarget.value.id}/reject`,
    method: 'POST', header: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
    data: { reason: rejectReason.value },
  })
  showRejectModal.value = false; load()
}

async function publish(item: any) {
  if (!item.account_id) return uni.showToast({ title: '请先关联发布账号', icon: 'none' })
  publishing.value[item.id] = true
  const token = uni.getStorageSync('token')
  const [, r] = await uni.request({
    url: `http://localhost:8080/api/v1/cms/publishes/${item.id}/publish`,
    method: 'POST', header: { Authorization: `Bearer ${token}` },
  }) as any
  publishing.value[item.id] = false
  uni.showToast({ title: r?.data?.data?.message || '已提交发布', icon: 'none', duration: 2500 })
  setTimeout(() => { load(); loadStats() }, 2000)
}

async function del(item: any) {
  if (item.status !== 'draft') return uni.showToast({ title: '仅草稿可删除', icon: 'none' })
  const ok = await new Promise(r => uni.showModal({ title: '确认删除', content: '删除此草稿？', success: res => r(res.confirm) }))
  if (!ok) return
  const token = uni.getStorageSync('token')
  await uni.request({ url: `http://localhost:8080/api/v1/cms/publishes/${item.id}`, method: 'DELETE', header: { Authorization: `Bearer ${token}` } })
  load()
}

onMounted(() => { load(); loadStats(); loadAccounts() })
</script>

<template>
  <view class="page">
    <!-- 统计卡片 -->
    <view class="stat-row">
      <view v-for="(v, k) in STATUS_MAP" :key="k" class="stat-card">
        <view class="stat-num" :style="{ color: v.color }">{{ statsByStatus[k] || 0 }}</view>
        <view class="stat-label">{{ v.label }}</view>
      </view>
    </view>

    <view class="page-header">
      <view class="page-title"><SvgIcon name="file" /> 发布管理</view>
      <view class="header-actions">
        <select class="filter-select" v-model="filterPlatform" @change="() => { page=1; load() }">
          <option value="">全部平台</option>
          <option v-for="(v,k) in PLATFORM_LABELS" :key="k" :value="k">{{ v }}</option>
        </select>
        <select class="filter-select" v-model="filterStatus" @change="() => { page=1; load() }">
          <option value="">全部状态</option>
          <option v-for="(v,k) in STATUS_MAP" :key="k" :value="k">{{ v.label }}</option>
        </select>
        <button class="btn" @click="load"><SvgIcon name="refresh" /></button>
      </view>
    </view>

    <view class="table-wrap">
      <view v-if="loading" class="loading-tip">加载中...</view>
      <view v-else-if="!list.length" class="empty-tip">暂无发布任务</view>
      <table v-else class="data-table">
        <thead>
          <tr><th>标题</th><th>目标平台</th><th>状态</th>
              <th>创建时间</th><th>定时发布</th><th>操作</th></tr>
        </thead>
        <tbody>
          <tr v-for="item in list" :key="item.id">
            <td class="title-cell">{{ item.final_title || '（无标题）' }}</td>
            <td><span class="platform-tag">{{ PLATFORM_LABELS[item.target_platform] || item.target_platform }}</span></td>
            <td>
              <span class="status-pill"
                :style="{ color: STATUS_MAP[item.status]?.color, background: STATUS_MAP[item.status]?.color + '20' }">
                {{ STATUS_MAP[item.status]?.label || item.status }}
              </span>
            </td>
            <td>{{ item.created_at?.substring(0,16) }}</td>
            <td>{{ item.scheduled_at ? item.scheduled_at.substring(0,16) : '-' }}</td>
            <td class="action-cell">
              <IconBtn icon="eye"    tip="查看详情" @click="openDetail(item)" />
              <IconBtn v-if="item.status==='draft' || item.status==='failed'"
                icon="check-circle" tip="审核通过" @click="approve(item)" />
              <IconBtn v-if="item.status==='draft' || item.status==='approved'"
                icon="close" tip="审核拒绝" @click="openReject(item)" />
              <IconBtn v-if="item.status==='approved'"
                icon="export"   tip="立即发布"
                :disabled="publishing[item.id]" @click="publish(item)" />
              <IconBtn v-if="item.status==='draft'"
                icon="delete" tip="删除" danger @click="del(item)" />
            </td>
          </tr>
        </tbody>
      </table>
    </view>

    <view class="pagination" v-if="total > pageSize">
      <button class="page-btn" :disabled="page<=1" @click="page--; load()">上一页</button>
      <text class="page-info">{{ page }} / {{ Math.ceil(total/pageSize) }}</text>
      <button class="page-btn" :disabled="page*pageSize>=total" @click="page++; load()">下一页</button>
    </view>

    <!-- 详情弹窗 -->
    <view v-if="showDetail && cur" class="modal-mask" @click.self="showDetail=false">
      <view class="modal-box modal-lg">
        <view class="modal-header">
          <text>发布详情</text>
          <view @click="showDetail=false"><SvgIcon name="close" /></view>
        </view>
        <view class="modal-body">
          <view class="detail-field"><label>标题</label><text>{{ cur.final_title }}</text></view>
          <view class="detail-field"><label>平台</label>
            <span class="platform-tag">{{ PLATFORM_LABELS[cur.target_platform] }}</span>
          </view>
          <view class="detail-field"><label>状态</label>
            <span class="status-pill" :style="{color: STATUS_MAP[cur.status]?.color}">
              {{ STATUS_MAP[cur.status]?.label }}
            </span>
          </view>
          <view v-if="cur.failure_reason" class="detail-field error">
            <label>失败原因</label><text>{{ cur.failure_reason }}</text>
          </view>
          <view class="detail-field"><label>正文预览</label></view>
          <view class="content-preview">{{ (cur.final_text || '').substring(0, 500) }}{{ cur.final_text?.length > 500 ? '...' : '' }}</view>
          <view v-if="cur.final_tags?.length" class="detail-field">
            <label>标签</label>
            <view class="tags-wrap"><span v-for="tag in cur.final_tags" :key="tag" class="tag">{{ tag }}</span></view>
          </view>
          <view class="detail-field"><label>发布账号</label>
            <select v-model.number="cur.account_id" class="form-select-sm">
              <option :value="0">未指定</option>
              <option v-for="acc in accounts.filter(a=>a.platform===cur.target_platform)" :key="acc.id" :value="acc.id">
                {{ acc.account_name }}
              </option>
            </select>
          </view>
        </view>
        <view class="modal-footer">
          <button v-if="cur.status==='approved'" class="btn btn-primary"
            :disabled="publishing[cur.id]" @click="publish(cur); showDetail=false">
            立即发布
          </button>
          <button class="btn" @click="showDetail=false">关闭</button>
        </view>
      </view>
    </view>

    <!-- 拒绝弹窗 -->
    <view v-if="showRejectModal" class="modal-mask" @click.self="showRejectModal=false">
      <view class="modal-box" style="width:400px">
        <view class="modal-header">
          <text>审核拒绝</text>
          <view @click="showRejectModal=false"><SvgIcon name="close" /></view>
        </view>
        <view class="modal-body">
          <view class="form-row">
            <label>拒绝原因 <span class="required">*</span></label>
            <textarea v-model="rejectReason" class="form-textarea" placeholder="请填写具体原因，帮助内容改进..." rows="3" />
          </view>
        </view>
        <view class="modal-footer">
          <button class="btn" @click="showRejectModal=false">取消</button>
          <button class="btn btn-danger" @click="confirmReject">确认拒绝</button>
        </view>
      </view>
    </view>
  </view>
</template>

<style scoped>
.page { padding: 24px; background: var(--color-bg); min-height: 100vh; }
.stat-row { display: flex; gap: 12px; margin-bottom: 20px; flex-wrap: wrap; }
.stat-card { background: var(--color-surface); border-radius: 8px; padding: 14px 20px; min-width: 100px; text-align: center; border: 1px solid var(--color-border); }
.stat-num { font-size: 24px; font-weight: 700; }
.stat-label { font-size: 12px; color: var(--color-text-muted); margin-top: 4px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-title { font-size: 18px; font-weight: 600; color: var(--color-text); display: flex; align-items: center; gap: 8px; }
.header-actions { display: flex; gap: 10px; align-items: center; }
.filter-select { height: 34px; border: 1px solid var(--color-border); border-radius: 6px; padding: 0 10px; background: var(--color-surface); color: var(--color-text); font-size: 13px; }
.btn { height: 34px; padding: 0 16px; border-radius: 6px; border: 1px solid var(--color-border); background: var(--color-surface); color: var(--color-text); cursor: pointer; font-size: 13px; display: inline-flex; align-items: center; gap: 6px; }
.btn-primary { background: var(--color-primary); color: #fff; border-color: var(--color-primary); }
.btn-danger  { background: #f56c6c; color: #fff; border-color: #f56c6c; }
.table-wrap { background: var(--color-surface); border-radius: 8px; overflow: auto; }
.data-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.data-table th { background: var(--color-bg); padding: 10px 14px; text-align: left; color: var(--color-text-muted); white-space: nowrap; }
.data-table td { padding: 10px 14px; border-top: 1px solid var(--color-border); color: var(--color-text); }
.data-table tr:hover td { background: var(--color-bg); }
.title-cell { max-width: 280px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.platform-tag { padding: 2px 8px; border-radius: 12px; background: var(--color-primary); color: #fff; font-size: 12px; }
.status-pill { padding: 2px 10px; border-radius: 10px; font-size: 12px; }
.action-cell { display: flex; gap: 6px; }
.loading-tip, .empty-tip { padding: 40px; text-align: center; color: var(--color-text-muted); }
.pagination { display: flex; justify-content: center; align-items: center; gap: 12px; margin-top: 16px; }
.page-btn { height: 30px; padding: 0 14px; border-radius: 6px; border: 1px solid var(--color-border); background: var(--color-surface); cursor: pointer; color: var(--color-text); font-size: 13px; }
.page-info { color: var(--color-text-muted); font-size: 13px; }
.modal-mask { position: fixed; inset: 0; background: rgba(0,0,0,.45); display: flex; align-items: center; justify-content: center; z-index: 1000; }
.modal-box { background: var(--color-surface); border-radius: 10px; width: 520px; max-height: 85vh; overflow-y: auto; }
.modal-lg { width: 680px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--color-border); font-weight: 600; color: var(--color-text); }
.modal-body { padding: 20px; }
.modal-footer { display: flex; justify-content: flex-end; gap: 10px; padding: 14px 20px; border-top: 1px solid var(--color-border); }
.detail-field { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; font-size: 13px; }
.detail-field label { width: 80px; color: var(--color-text-muted); flex-shrink: 0; }
.detail-field.error text { color: #f56c6c; }
.content-preview { background: var(--color-bg); border-radius: 6px; padding: 12px; font-size: 13px; line-height: 1.7; color: var(--color-text); margin-bottom: 14px; max-height: 200px; overflow-y: auto; white-space: pre-wrap; }
.tags-wrap { display: flex; flex-wrap: wrap; gap: 6px; }
.tag { padding: 2px 10px; border-radius: 12px; background: var(--color-primary-light, #e8f4ff); color: var(--color-primary); font-size: 12px; }
.form-select-sm { height: 30px; border: 1px solid var(--color-border); border-radius: 6px; padding: 0 8px; background: var(--color-bg); color: var(--color-text); font-size: 13px; }
.form-row { margin-bottom: 14px; }
.form-row > label { display: block; font-size: 13px; color: var(--color-text-muted); margin-bottom: 6px; }
.required { color: #f56c6c; }
.form-textarea { width: 100%; border: 1px solid var(--color-border); border-radius: 6px; padding: 8px 10px; background: var(--color-bg); color: var(--color-text); font-size: 13px; resize: vertical; box-sizing: border-box; }
</style>
