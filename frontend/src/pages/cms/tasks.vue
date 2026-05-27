<script setup lang="ts">
import { ref, onMounted } from 'vue'
import SvgIcon from '@/components/common/SvgIcon.vue'
import IconBtn from '@/components/common/IconBtn.vue'

const PLATFORM_LABELS: Record<string, string> = {
  wechat: '微信公众号', rednote: '小红书', douyin: '抖音', csdn: 'CSDN',
}
const TASK_TYPE_LABELS: Record<string, string> = {
  search_title:  '关键词搜索',
  follow_author: '关注作者',
}
const CRON_OPTIONS = [
  { value: '', label: '手动执行' },
  { value: '@hourly', label: '每小时' },
  { value: '@daily', label: '每天' },
  { value: '@weekly', label: '每周' },
]

const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 10
const loading = ref(false)
const saving = ref(false)
const running = ref<Record<number, boolean>>({})
const showModal = ref(false)
const editMode = ref(false)
const curTask = ref<any>(null)

const form = ref({
  task_name: '', platform: 'wechat', task_type: 'search_title',
  target_param: '', target_platform: 'wechat', account_id: 0,
  cron_expr: '', fetch_limit: 5, status: 1,
})

const filter = ref({ platform: '' })
const accounts = ref<any[]>([])

async function load() {
  loading.value = true
  const token = uni.getStorageSync('token')
  const params = new URLSearchParams({ page: String(page.value), page_size: String(pageSize),
    ...(filter.value.platform ? { platform: filter.value.platform } : {}) })
  const [, r] = await uni.request({
    url: `http://localhost:8080/api/v1/cms/tasks?${params}`,
    header: { Authorization: `Bearer ${token}` },
  }) as any
  loading.value = false
  list.value  = r?.data?.data?.list  || []
  total.value = r?.data?.data?.total || 0
}

async function loadAccounts() {
  const token = uni.getStorageSync('token')
  const [, r] = await uni.request({
    url: 'http://localhost:8080/api/v1/cms/accounts?page_size=100',
    header: { Authorization: `Bearer ${token}` },
  }) as any
  accounts.value = r?.data?.data?.list || []
}

function openCreate() {
  editMode.value = false
  curTask.value = null
  form.value = { task_name: '', platform: 'wechat', task_type: 'search_title',
    target_param: '', target_platform: 'wechat', account_id: 0, cron_expr: '', fetch_limit: 5, status: 1 }
  showModal.value = true
}

function openEdit(task: any) {
  editMode.value = true
  curTask.value = task
  form.value = { ...task }
  showModal.value = true
}

async function save() {
  if (!form.value.task_name) return uni.showToast({ title: '请填写任务名称', icon: 'none' })
  if (!form.value.target_param) return uni.showToast({ title: '请填写搜索词或作者ID', icon: 'none' })
  saving.value = true
  const token = uni.getStorageSync('token')
  const url = editMode.value
    ? `http://localhost:8080/api/v1/cms/tasks/${curTask.value.id}`
    : 'http://localhost:8080/api/v1/cms/tasks'
  const [err, res] = await uni.request({
    url, method: editMode.value ? 'PUT' : 'POST',
    header: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
    data: form.value,
  }) as any
  saving.value = false
  if (!err && res.data.code === 0) {
    uni.showToast({ title: '保存成功', icon: 'success' })
    showModal.value = false; load()
  } else uni.showToast({ title: res?.data?.message || '失败', icon: 'none' })
}

async function runTask(task: any) {
  running.value[task.id] = true
  const token = uni.getStorageSync('token')
  await uni.request({
    url: `http://localhost:8080/api/v1/cms/tasks/${task.id}/run`,
    method: 'POST',
    header: { Authorization: `Bearer ${token}` },
  })
  running.value[task.id] = false
  uni.showToast({ title: '任务已触发，请查看内容列表', icon: 'none', duration: 2000 })
  setTimeout(load, 2000)
}

async function toggleStatus(task: any) {
  const token = uni.getStorageSync('token')
  await uni.request({
    url: `http://localhost:8080/api/v1/cms/tasks/${task.id}`,
    method: 'PUT',
    header: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
    data: { status: task.status === 1 ? 0 : 1 },
  })
  load()
}

async function del(task: any) {
  const ok = await new Promise(r => uni.showModal({ title: '确认删除', content: `删除任务「${task.task_name}」？`, success: (res) => r(res.confirm) }))
  if (!ok) return
  const token = uni.getStorageSync('token')
  await uni.request({ url: `http://localhost:8080/api/v1/cms/tasks/${task.id}`, method: 'DELETE', header: { Authorization: `Bearer ${token}` } })
  load()
}

onMounted(() => { load(); loadAccounts() })
</script>

<template>
  <view class="page">
    <view class="page-header">
      <view class="page-title"><SvgIcon name="calendar" /> 采集任务</view>
      <view class="header-actions">
        <select class="filter-select" v-model="filter.platform" @change="() => { page=1; load() }">
          <option value="">全部平台</option>
          <option v-for="(v,k) in PLATFORM_LABELS" :key="k" :value="k">{{ v }}</option>
        </select>
        <button class="btn btn-primary" @click="openCreate"><SvgIcon name="add" /> 新建任务</button>
      </view>
    </view>

    <view class="table-wrap">
      <view v-if="loading" class="loading-tip">加载中...</view>
      <view v-else-if="!list.length" class="empty-tip">暂无采集任务</view>
      <table v-else class="data-table">
        <thead>
          <tr><th>任务名称</th><th>采集平台</th><th>类型</th><th>目标参数</th>
              <th>采集数</th><th>执行周期</th><th>状态</th><th>上次执行</th><th>操作</th></tr>
        </thead>
        <tbody>
          <tr v-for="t in list" :key="t.id">
            <td>{{ t.task_name }}</td>
            <td><span class="platform-tag">{{ PLATFORM_LABELS[t.platform] || t.platform }}</span></td>
            <td>{{ TASK_TYPE_LABELS[t.task_type] }}</td>
            <td class="target-param">{{ t.target_param }}</td>
            <td>{{ t.fetch_limit }}</td>
            <td>{{ t.cron_expr || '手动' }}</td>
            <td>
              <span class="status-pill" :class="t.status === 1 ? 'on' : 'off'">
                {{ t.status === 1 ? '运行中' : '已停用' }}
              </span>
            </td>
            <td>{{ t.last_run_at ? t.last_run_at.substring(0,19) : '-' }}</td>
            <td class="action-cell">
              <IconBtn icon="refresh" :tip="running[t.id] ? '执行中...' : '立即执行'"
                @click="runTask(t)" :disabled="running[t.id]" />
              <IconBtn :icon="t.status===1 ? 'ban' : 'check'"
                :tip="t.status===1 ? '停用' : '启用'" @click="toggleStatus(t)" />
              <IconBtn icon="edit" tip="编辑" @click="openEdit(t)" />
              <IconBtn icon="delete" tip="删除" danger @click="del(t)" />
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

    <!-- 新增/编辑弹窗 -->
    <view v-if="showModal" class="modal-mask" @click.self="showModal=false">
      <view class="modal-box">
        <view class="modal-header">
          <text>{{ editMode ? '编辑任务' : '新建采集任务' }}</text>
          <view @click="showModal=false"><SvgIcon name="close" /></view>
        </view>
        <view class="modal-body">
          <view class="form-row">
            <label>任务名称 <span class="required">*</span></label>
            <input v-model="form.task_name" placeholder="如：科技新闻每日采集" class="form-input" />
          </view>
          <view class="form-row-2">
            <view class="form-row">
              <label>采集平台 <span class="required">*</span></label>
              <select v-model="form.platform" class="form-select">
                <option v-for="(v,k) in PLATFORM_LABELS" :key="k" :value="k">{{ v }}</option>
              </select>
            </view>
            <view class="form-row">
              <label>任务类型 <span class="required">*</span></label>
              <select v-model="form.task_type" class="form-select">
                <option v-for="(v,k) in TASK_TYPE_LABELS" :key="k" :value="k">{{ v }}</option>
              </select>
            </view>
          </view>
          <view class="form-row">
            <label>
              {{ form.task_type === 'search_title' ? '搜索关键词' : '作者ID/用户名' }}
              <span class="required">*</span>
            </label>
            <input v-model="form.target_param"
              :placeholder="form.task_type==='search_title' ? '如：人工智能最新进展' : '如：zhangsan（CSDN用户名）'"
              class="form-input" />
          </view>
          <view class="form-row-2">
            <view class="form-row">
              <label>目标发布平台</label>
              <select v-model="form.target_platform" class="form-select">
                <option v-for="(v,k) in PLATFORM_LABELS" :key="k" :value="k">{{ v }}</option>
              </select>
            </view>
            <view class="form-row">
              <label>每次采集条数</label>
              <input v-model.number="form.fetch_limit" type="number" min="1" max="50" class="form-input" />
            </view>
          </view>
          <view class="form-row-2">
            <view class="form-row">
              <label>执行周期</label>
              <select v-model="form.cron_expr" class="form-select">
                <option v-for="opt in CRON_OPTIONS" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
              </select>
            </view>
            <view class="form-row">
              <label>发布账号</label>
              <select v-model.number="form.account_id" class="form-select">
                <option :value="0">不指定</option>
                <option v-for="acc in accounts.filter(a=>a.platform===form.target_platform)" :key="acc.id" :value="acc.id">
                  {{ acc.account_name }}
                </option>
              </select>
            </view>
          </view>
        </view>
        <view class="modal-footer">
          <button class="btn" @click="showModal=false">取消</button>
          <button class="btn btn-primary" :disabled="saving" @click="save">{{ saving ? '保存中...' : '保存' }}</button>
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
.data-table th { background: var(--color-bg); padding: 10px 14px; text-align: left; color: var(--color-text-muted); white-space: nowrap; }
.data-table td { padding: 10px 14px; border-top: 1px solid var(--color-border); color: var(--color-text); }
.data-table tr:hover td { background: var(--color-bg); }
.platform-tag { padding: 2px 8px; border-radius: 12px; background: var(--color-primary); color: #fff; font-size: 12px; }
.target-param { max-width: 180px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.status-pill { padding: 2px 10px; border-radius: 10px; font-size: 12px; }
.status-pill.on  { background: #e8f7ef; color: #27ae60; }
.status-pill.off { background: #f5f5f5; color: #999; }
.action-cell { display: flex; gap: 6px; }
.loading-tip, .empty-tip { padding: 40px; text-align: center; color: var(--color-text-muted); }
.pagination { display: flex; justify-content: center; align-items: center; gap: 12px; margin-top: 16px; }
.page-btn { height: 30px; padding: 0 14px; border-radius: 6px; border: 1px solid var(--color-border); background: var(--color-surface); cursor: pointer; color: var(--color-text); font-size: 13px; }
.page-info { color: var(--color-text-muted); font-size: 13px; }
.modal-mask { position: fixed; inset: 0; background: rgba(0,0,0,.45); display: flex; align-items: center; justify-content: center; z-index: 1000; }
.modal-box { background: var(--color-surface); border-radius: 10px; width: 560px; max-height: 85vh; overflow-y: auto; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--color-border); font-weight: 600; color: var(--color-text); }
.modal-body { padding: 20px; }
.modal-footer { display: flex; justify-content: flex-end; gap: 10px; padding: 14px 20px; border-top: 1px solid var(--color-border); }
.form-row { margin-bottom: 14px; }
.form-row > label { display: block; font-size: 13px; color: var(--color-text-muted); margin-bottom: 6px; }
.form-row-2 { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
.required { color: #f56c6c; }
.form-input, .form-select { width: 100%; height: 34px; border: 1px solid var(--color-border); border-radius: 6px; padding: 0 10px; background: var(--color-bg); color: var(--color-text); font-size: 13px; box-sizing: border-box; }
</style>
