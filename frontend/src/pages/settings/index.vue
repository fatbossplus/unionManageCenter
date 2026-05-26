<script setup lang="ts">
import { ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import { useThemeStore } from '@/stores/theme'

const breadcrumbs = [{ label: '首页' }, { label: '系统' }, { label: '系统设置' }]
const themeStore = useThemeStore()
const activeSection = ref('basic')

const sections = [
  { key: 'basic',   label: '基础信息' },
  { key: 'security',label: '安全配置' },
  { key: 'theme',   label: '主题外观' },
  { key: 'notify',  label: '通知设置' },
]

const themes = [
  { key: 'a' as const, label: '深色科技', desc: '深色底 + 蓝色强调，适合低光环境', preview: '#0d1117' },
  { key: 'b' as const, label: '纯净浅蓝', desc: '白底 + 深蓝侧边，简约清爽（默认）', preview: '#1e40af' },
  { key: 'c' as const, label: '渐变紫金', desc: '白底 + 紫渐变，现代活泼', preview: 'linear-gradient(135deg,#7c3aed,#f59e0b)' },
]
</script>

<template>
  <AppLayout :breadcrumbs="breadcrumbs">
    <view class="settings-layout">
      <view class="settings-nav card">
        <view v-for="s in sections" :key="s.key" class="s-nav-item"
          :class="{ active: activeSection === s.key }" @click="activeSection = s.key">
          {{ s.label }}
        </view>
      </view>

      <view class="settings-body card">
        <!-- 基础信息 -->
        <view v-if="activeSection === 'basic'">
          <text class="section-title">基础信息配置</text>
          <view class="form-row">
            <view class="form-item">
              <text class="form-label">系统名称</text>
              <input class="form-input" value="联盟管理中心" />
            </view>
            <view class="form-item">
              <text class="form-label">系统域名</text>
              <input class="form-input" value="admin.union.example.com" />
            </view>
          </view>
          <view class="form-row">
            <view class="form-item">
              <text class="form-label">联系邮箱</text>
              <input class="form-input" value="admin@union.com" />
            </view>
            <view class="form-item">
              <text class="form-label">备案号</text>
              <input class="form-input" placeholder="粤ICP备XXXXXXXX号" />
            </view>
          </view>
          <view class="form-actions">
            <view class="s-btn s-btn-primary">保存设置</view>
          </view>
        </view>

        <!-- 主题外观 -->
        <view v-else-if="activeSection === 'theme'">
          <text class="section-title">主题外观</text>
          <text class="section-sub">选择一套主题，整个管理后台将同步切换配色方案</text>
          <view class="theme-options">
            <view v-for="t in themes" :key="t.key" class="theme-option"
              :class="{ active: themeStore.current === t.key }"
              @click="themeStore.apply(t.key)">
              <view class="theme-preview" :style="{ background: t.preview }" />
              <view class="theme-info">
                <text class="theme-name">{{ t.label }}</text>
                <text class="theme-desc">{{ t.desc }}</text>
              </view>
              <view v-if="themeStore.current === t.key" class="theme-check">✓</view>
            </view>
          </view>
        </view>

        <!-- 其他 -->
        <view v-else class="empty-tip">
          <text>{{ sections.find(s=>s.key===activeSection)?.label }} 配置模块开发中...</text>
        </view>
      </view>
    </view>
  </AppLayout>
</template>

<style lang="scss" scoped>
.settings-layout { display:flex; gap:16px; align-items:flex-start; }
.settings-nav { width:160px; flex-shrink:0; padding:8px 0; }
.s-nav-item { padding:10px 20px; font-size:13px; color:var(--color-text-secondary); cursor:pointer; &:hover{background:var(--color-border-light);} &.active{background:var(--color-primary-light);color:var(--color-primary);font-weight:600;} }
.settings-body { flex:1; padding:24px; }
.section-title { font-size:16px; font-weight:700; color:var(--color-text-primary); display:block; margin-bottom:6px; }
.section-sub { font-size:13px; color:var(--color-text-muted); display:block; margin-bottom:20px; }
.form-row { display:grid; grid-template-columns:1fr 1fr; gap:16px; margin-bottom:16px; }
.form-item { display:flex; flex-direction:column; gap:6px; }
.form-label { font-size:12px; color:var(--color-text-secondary); font-weight:500; }
.form-input { height:38px; border:1px solid var(--color-border); border-radius:8px; padding:0 12px; font-size:13px; color:var(--color-text-primary); background:var(--color-card-bg); }
.form-actions { margin-top:24px; }
.s-btn { height:38px; border-radius:8px; font-size:13px; font-weight:500; display:inline-flex; align-items:center; padding:0 20px; cursor:pointer; }
.s-btn-primary { background:var(--color-primary); color:#fff; }
.theme-options { display:flex; flex-direction:column; gap:12px; margin-top:16px; }
.theme-option { display:flex; align-items:center; gap:16px; padding:16px; border:2px solid var(--color-border); border-radius:12px; cursor:pointer; transition:all 0.15s; &:hover{border-color:var(--color-primary);} &.active{border-color:var(--color-primary);background:var(--color-primary-light);} }
.theme-preview { width:48px; height:48px; border-radius:10px; flex-shrink:0; }
.theme-info { flex:1; }
.theme-name { font-size:14px; font-weight:600; color:var(--color-text-primary); display:block; margin-bottom:4px; }
.theme-desc { font-size:12px; color:var(--color-text-muted); display:block; }
.theme-check { width:24px; height:24px; border-radius:50%; background:var(--color-primary); color:#fff; display:flex; align-items:center; justify-content:center; font-size:12px; font-weight:bold; }
.empty-tip { padding:60px; text-align:center; color:var(--color-text-muted); font-size:14px; }
</style>
