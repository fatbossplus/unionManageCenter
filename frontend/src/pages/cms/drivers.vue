<script setup lang="ts">
import { ref, onMounted } from 'vue'
import SvgIcon from '@/components/common/SvgIcon.vue'

const KEY_LABELS: Record<string, string> = {
  'wechat.scraper':    '微信公众号 · 采集',
  'rednote.scraper':   '小红书 · 采集',
  'douyin.scraper':    '抖音 · 采集',
  'csdn.scraper':      'CSDN · 采集',
  'ai.rewriter':       'AI · 改写引擎',
  'compliance.local':  '合规 · 本地检查',
  'compliance.api':    '合规 · 云API检查',
  'proxy.pool':        '代理 · IP池',
  'wechat.publisher':  '微信公众号 · 发布',
  'rednote.publisher': '小红书 · 发布',
  'douyin.publisher':  '抖音 · 发布',
  'csdn.publisher':    'CSDN · 发布',
}

const rows = ref<any[]>([])
const meta = ref<Record<string, any[]>>({})
const loading = ref(false)
const saving = ref(false)

// 展开编辑某一行
const editing = ref<Record<string, boolean>>({})
const editForms = ref<Record<string, any>>({})
const configInputs = ref<Record<string, Record<string, string>>>({})

async function load() {
  loading.value = true
  const token = uni.getStorageSync('token')
  const [, r1] = await uni.request({ url: 'http://localhost:8080/api/v1/cms/drivers', header: { Authorization: `Bearer ${token}` } }) as any
  const [, r2] = await uni.request({ url: 'http://localhost:8080/api/v1/cms/drivers/meta', header: { Authorization: `Bearer ${token}` } }) as any
  loading.value = false
  rows.value = r1?.data?.data || []
  meta.value  = r2?.data?.data || {}
}

function toggleEdit(row: any) {
  const key = row.config?.config_key
  if (!key) return
  editing.value[key] = !editing.value[key]
  if (editing.value[key]) {
    editForms.value[key] = {
      config_key:  key,
      org_id:      row.config?.org_id || 0,
      driver_name: row.config?.driver_name,
      enabled:     row.config?.enabled ?? 1,
    }
    // 初始化参数输入框
    configInputs.value[key] = {}
    try {
      const cfg = JSON.parse(row.config?.config_json || '{}')
      configInputs.value[key] = cfg
    } catch {}
  }
}

function getAvailable(row: any) {
  const key = row.config?.config_key
  return meta.value[key] || row.available || []
}

function onDriverChange(key: string, driverName: string) {
  editForms.value[key].driver_name = driverName
  configInputs.value[key] = {} // 清空参数
}

function getConfigSchema(key: string): any[] {
  const driverName = editForms.value[key]?.driver_name
  const available = getAvailableByKey(key)
  const driver = available.find((d: any) => d.name === driverName)
  return driver?.config_schema || []
}

function getAvailableByKey(key: string) {
  const row = rows.value.find(r => r.config?.config_key === key)
  return meta.value[key] || row?.available || []
}

async function saveDriver(key: string) {
  const form = editForms.value[key]
  saving.value = true
  const token = uni.getStorageSync('token')

  // 收集 config_json
  const cfgObj = configInputs.value[key] || {}
  const payload = {
    ...form,
    config_json: Object.keys(cfgObj).length ? JSON.stringify(cfgObj) : '',
  }

  const [err, res] = await uni.request({
    url: 'http://localhost:8080/api/v1/cms/drivers',
    method: 'PUT',
    header: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
    data: payload,
  }) as any
  saving.value = false
  if (!err && res.data.code === 0) {
    uni.showToast({ title: '保存成功', icon: 'success' })
    editing.value[key] = false
    load()
  } else {
    uni.showToast({ title: res?.data?.message || '保存失败', icon: 'none' })
  }
}

onMounted(load)
</script>

<template>
  <view class="page">
    <view class="page-header">
      <view class="page-title"><SvgIcon name="setting" /> 驱动配置</view>
      <text class="page-desc">配置各平台的采集、AI改写、合规检查、发布驱动，可在免费/付费之间切换</text>
    </view>

    <view v-if="loading" class="loading-tip">加载中...</view>
    <view v-else class="driver-grid">
      <view v-for="row in rows" :key="row.config?.config_key" class="driver-card">
        <view class="card-head">
          <view>
            <text class="card-title">{{ KEY_LABELS[row.config?.config_key] || row.config?.config_key }}</text>
            <view class="driver-badges">
              <span class="badge" :class="row.config?.driver_type === 'free' ? 'badge-free' : 'badge-paid'">
                {{ row.config?.driver_type === 'free' ? '免费驱动' : '付费驱动' }}
              </span>
              <span class="badge" :class="row.config?.enabled ? 'badge-on' : 'badge-off'">
                {{ row.config?.enabled ? '已启用' : '已禁用' }}
              </span>
            </view>
          </view>
          <view class="card-actions">
            <text class="current-driver">{{ row.config?.driver_name }}</text>
            <button class="btn-edit" @click="toggleEdit(row)">
              {{ editing[row.config?.config_key] ? '收起' : '修改' }}
            </button>
          </view>
        </view>

        <!-- 编辑区 -->
        <view v-if="editing[row.config?.config_key]" class="edit-area">
          <view class="form-row">
            <label>切换驱动</label>
            <view class="driver-options">
              <view v-for="drv in getAvailable(row)" :key="drv.name"
                class="driver-option"
                :class="{ active: editForms[row.config?.config_key]?.driver_name === drv.name }"
                @click="onDriverChange(row.config?.config_key, drv.name)">
                <view class="drv-name">{{ drv.display_name }}</view>
                <span class="badge" :class="drv.type === 'free' ? 'badge-free' : 'badge-paid'">
                  {{ drv.type === 'free' ? '免费' : '付费' }}
                </span>
                <view class="drv-cost">{{ drv.cost_desc }}</view>
                <view class="drv-stars">
                  <SvgIcon v-for="i in drv.stability" :key="i" name="star" style="color:#f5a623;font-size:12px;" />
                  <SvgIcon v-for="i in (5-drv.stability)" :key="'e'+i" name="star" style="color:#ddd;font-size:12px;" />
                </view>
                <view class="drv-desc">{{ drv.desc }}</view>
              </view>
            </view>
          </view>

          <!-- 驱动参数 -->
          <view v-if="getConfigSchema(row.config?.config_key).length" class="form-row">
            <label>驱动参数</label>
            <view v-for="field in getConfigSchema(row.config?.config_key)" :key="field.key" class="config-field">
              <label class="field-label">{{ field.label }}<span v-if="field.required" class="required">*</span></label>
              <input
                v-model="configInputs[row.config?.config_key][field.key]"
                :type="field.type === 'password' ? 'password' : 'text'"
                :placeholder="field.placeholder"
                class="form-input"
              />
            </view>
          </view>

          <view class="form-row">
            <label>启用状态</label>
            <view class="toggle-wrap">
              <input type="checkbox" v-model="editForms[row.config?.config_key].enabled" true-value="1" false-value="0" />
              <text>{{ editForms[row.config?.config_key]?.enabled ? '启用' : '禁用' }}</text>
            </view>
          </view>

          <view class="edit-footer">
            <button class="btn" @click="editing[row.config?.config_key]=false">取消</button>
            <button class="btn btn-primary" :disabled="saving" @click="saveDriver(row.config?.config_key)">
              {{ saving ? '保存中...' : '保存配置' }}
            </button>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<style scoped>
.page { padding: 24px; background: var(--color-bg); min-height: 100vh; }
.page-header { margin-bottom: 20px; }
.page-title { font-size: 18px; font-weight: 600; color: var(--color-text); display: flex; align-items: center; gap: 8px; margin-bottom: 4px; }
.page-desc { font-size: 13px; color: var(--color-text-muted); }
.loading-tip { padding: 40px; text-align: center; color: var(--color-text-muted); }
.driver-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(460px, 1fr)); gap: 16px; }
.driver-card { background: var(--color-surface); border-radius: 8px; border: 1px solid var(--color-border); overflow: hidden; }
.card-head { display: flex; justify-content: space-between; align-items: flex-start; padding: 14px 16px; }
.card-title { font-size: 14px; font-weight: 600; color: var(--color-text); }
.driver-badges { display: flex; gap: 6px; margin-top: 6px; }
.badge { padding: 1px 8px; border-radius: 10px; font-size: 11px; }
.badge-free { background: #e8f7ef; color: #2ecc71; }
.badge-paid { background: #fff3e0; color: #f39c12; }
.badge-on  { background: #e8f4ff; color: var(--color-primary); }
.badge-off { background: #f5f5f5; color: #999; }
.card-actions { display: flex; align-items: center; gap: 10px; }
.current-driver { font-size: 12px; color: var(--color-text-muted); background: var(--color-bg); padding: 2px 8px; border-radius: 4px; }
.btn-edit { height: 28px; padding: 0 12px; border-radius: 6px; border: 1px solid var(--color-primary); background: transparent; color: var(--color-primary); cursor: pointer; font-size: 12px; }
.edit-area { border-top: 1px solid var(--color-border); padding: 16px; background: var(--color-bg); }
.form-row { margin-bottom: 14px; }
.form-row > label { display: block; font-size: 12px; color: var(--color-text-muted); margin-bottom: 8px; }
.driver-options { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 10px; }
.driver-option { border: 2px solid var(--color-border); border-radius: 8px; padding: 10px 12px; cursor: pointer; background: var(--color-surface); transition: all .2s; }
.driver-option.active { border-color: var(--color-primary); background: var(--color-primary-light, #f0f7ff); }
.drv-name { font-size: 13px; font-weight: 600; color: var(--color-text); margin-bottom: 4px; }
.drv-cost { font-size: 12px; color: var(--color-text-muted); margin: 3px 0; }
.drv-stars { display: flex; gap: 2px; margin: 3px 0; }
.drv-desc { font-size: 11px; color: var(--color-text-muted); line-height: 1.5; margin-top: 4px; }
.config-field { margin-bottom: 10px; }
.field-label { font-size: 12px; color: var(--color-text-muted); display: block; margin-bottom: 4px; }
.required { color: #f56c6c; }
.form-input { width: 100%; height: 32px; border: 1px solid var(--color-border); border-radius: 6px; padding: 0 8px; background: var(--color-surface); color: var(--color-text); font-size: 13px; box-sizing: border-box; }
.toggle-wrap { display: flex; align-items: center; gap: 8px; font-size: 13px; color: var(--color-text); }
.edit-footer { display: flex; justify-content: flex-end; gap: 10px; margin-top: 14px; }
.btn { height: 32px; padding: 0 16px; border-radius: 6px; border: 1px solid var(--color-border); background: var(--color-surface); color: var(--color-text); cursor: pointer; font-size: 13px; }
.btn-primary { background: var(--color-primary); color: #fff; border-color: var(--color-primary); }
</style>
