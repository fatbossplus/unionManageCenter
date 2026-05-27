<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import { useThemeStore } from '@/stores/theme'
import { useUserStore } from '@/stores/user'
import { get, put, post } from '@/api/request'

const breadcrumbs = [{ label: '首页' }, { label: '系统' }, { label: '系统设置' }]
const themeStore = useThemeStore()
const userStore  = useUserStore()
const activeSection = ref('basic')

const sections = [
  { key: 'basic',    label: '基础信息' },
  { key: 'profile',  label: '个人资料' },
  { key: 'security', label: '修改密码' },
  { key: 'theme',    label: '主题外观' },
]

const themes = [
  { key: 'a' as const, label: '深色科技',  desc: '深色底 + 蓝色强调，适合低光环境',  preview: '#0d1117' },
  { key: 'b' as const, label: '纯净浅蓝',  desc: '白底 + 深蓝侧边，简约清爽（默认）', preview: '#1e40af' },
  { key: 'c' as const, label: '渐变紫金',  desc: '白底 + 紫渐变，现代活泼',           preview: 'linear-gradient(135deg,#7c3aed,#f59e0b)' },
]

// 基础信息（系统配置，仅展示）
const sysForm = reactive({ siteName: '联盟管理中心', domain: 'admin.union.example.com', email: 'admin@union.com', icp: '' })
const sysSaving = ref(false)
function saveSysConfig() {
  sysSaving.value = true
  setTimeout(() => { sysSaving.value = false; uni.showToast({ title: '系统配置已保存', icon: 'success' }) }, 600)
}

// 个人资料
const profileForm = reactive({ username: '', email: '', phone: '', realName: '' })
const profileSaving = ref(false)

async function loadProfile() {
  try {
    const me: any = await get('/users/me')  // 由 AdminHandler.Me 处理，返回当前管理员信息
    Object.assign(profileForm, { username: me.username || '', email: me.email || '', phone: me.phone || '', realName: me.real_name || '' })
  } catch {}
}
async function saveProfile() {
  profileSaving.value = true
  try {
    const userId = userStore.info?.id
    if (!userId) { uni.showToast({ title: '未获取到用户信息', icon: 'none' }); return }
    await put(`/admins/${userId}`, { email: profileForm.email, phone: profileForm.phone, real_name: profileForm.realName })
    // 同步更新 store 中的 email 字段
    userStore.setInfo({ ...userStore.info!, email: profileForm.email })
    uni.showToast({ title: '个人资料已更新', icon: 'success' })
  } catch (e: any) {
    uni.showToast({ title: e?.message || '保存失败', icon: 'none' })
  } finally { profileSaving.value = false }
}

// 修改密码
const pwdForm = reactive({ oldPwd: '', newPwd: '', confirmPwd: '' })
const pwdSaving = ref(false)
async function changePassword() {
  if (!pwdForm.oldPwd || !pwdForm.newPwd) { uni.showToast({ title: '请填写完整密码信息', icon: 'none' }); return }
  if (pwdForm.newPwd !== pwdForm.confirmPwd) { uni.showToast({ title: '两次新密码不一致', icon: 'none' }); return }
  if (pwdForm.newPwd.length < 6) { uni.showToast({ title: '新密码至少 6 位', icon: 'none' }); return }
  pwdSaving.value = true
  try {
    const userId = userStore.info?.id
    await put(`/admins/${userId}`, { password: pwdForm.newPwd })
    uni.showToast({ title: '密码修改成功，请重新登录', icon: 'success' })
    Object.assign(pwdForm, { oldPwd: '', newPwd: '', confirmPwd: '' })
  } catch (e: any) {
    uni.showToast({ title: e?.message || '修改失败', icon: 'none' })
  } finally { pwdSaving.value = false }
}

onMounted(loadProfile)
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
          <text class="section-sub">系统基础参数配置，修改后立即生效</text>
          <view class="form-row">
            <view class="form-item">
              <text class="form-label">系统名称</text>
              <input class="form-input" v-model="sysForm.siteName"/>
            </view>
            <view class="form-item">
              <text class="form-label">系统域名</text>
              <input class="form-input" v-model="sysForm.domain"/>
            </view>
          </view>
          <view class="form-row">
            <view class="form-item">
              <text class="form-label">联系邮箱</text>
              <input class="form-input" v-model="sysForm.email"/>
            </view>
            <view class="form-item">
              <text class="form-label">备案号</text>
              <input class="form-input" v-model="sysForm.icp" placeholder="粤ICP备XXXXXXXX号"/>
            </view>
          </view>
          <view class="form-actions">
            <view class="s-btn s-btn-primary" :class="{ saving: sysSaving }" @click="saveSysConfig">
              {{ sysSaving ? '保存中...' : '保存设置' }}
            </view>
          </view>
        </view>

        <!-- 个人资料 -->
        <view v-else-if="activeSection === 'profile'">
          <text class="section-title">个人资料</text>
          <text class="section-sub">修改您的账号基础信息</text>
          <view class="form-row">
            <view class="form-item">
              <text class="form-label">用户名（不可修改）</text>
              <input class="form-input" :value="profileForm.username" disabled/>
            </view>
            <view class="form-item">
              <text class="form-label">真实姓名</text>
              <input class="form-input" v-model="profileForm.realName" placeholder="请输入真实姓名"/>
            </view>
          </view>
          <view class="form-row">
            <view class="form-item">
              <text class="form-label">邮箱</text>
              <input class="form-input" v-model="profileForm.email" placeholder="请输入邮箱"/>
            </view>
            <view class="form-item">
              <text class="form-label">手机号</text>
              <input class="form-input" v-model="profileForm.phone" placeholder="请输入手机号"/>
            </view>
          </view>
          <view class="form-actions">
            <view class="s-btn s-btn-primary" :class="{ saving: profileSaving }" @click="saveProfile">
              {{ profileSaving ? '保存中...' : '保存资料' }}
            </view>
          </view>
        </view>

        <!-- 修改密码 -->
        <view v-else-if="activeSection === 'security'">
          <text class="section-title">修改密码</text>
          <text class="section-sub">建议定期更换密码以保障账号安全</text>
          <view class="form-row single">
            <view class="form-item">
              <text class="form-label">当前密码</text>
              <input class="form-input" type="password" v-model="pwdForm.oldPwd" placeholder="请输入当前密码"/>
            </view>
          </view>
          <view class="form-row single">
            <view class="form-item">
              <text class="form-label">新密码</text>
              <input class="form-input" type="password" v-model="pwdForm.newPwd" placeholder="请输入新密码（至少6位）"/>
            </view>
          </view>
          <view class="form-row single">
            <view class="form-item">
              <text class="form-label">确认新密码</text>
              <input class="form-input" type="password" v-model="pwdForm.confirmPwd" placeholder="再次输入新密码"/>
            </view>
          </view>
          <view class="form-actions">
            <view class="s-btn s-btn-primary" :class="{ saving: pwdSaving }" @click="changePassword">
              {{ pwdSaving ? '修改中...' : '确认修改密码' }}
            </view>
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
              <view class="theme-preview" :style="{ background: t.preview }"/>
              <view class="theme-info">
                <text class="theme-name">{{ t.label }}</text>
                <text class="theme-desc">{{ t.desc }}</text>
              </view>
              <view v-if="themeStore.current === t.key" class="theme-check">✓</view>
            </view>
          </view>
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
.form-row.single { grid-template-columns:1fr; max-width:360px; }
.form-actions { margin-top:24px; }
.s-btn { height:38px; border-radius:8px; font-size:13px; font-weight:500; display:inline-flex; align-items:center; padding:0 20px; cursor:pointer; transition:opacity 0.15s; }
.s-btn-primary { background:var(--color-primary); color:#fff; &.saving{opacity:0.7;pointer-events:none;} }
.theme-options { display:flex; flex-direction:column; gap:12px; margin-top:16px; }
.theme-option { display:flex; align-items:center; gap:16px; padding:16px; border:2px solid var(--color-border); border-radius:12px; cursor:pointer; transition:all 0.15s; &:hover{border-color:var(--color-primary);} &.active{border-color:var(--color-primary);background:var(--color-primary-light);} }
.theme-preview { width:48px; height:48px; border-radius:10px; flex-shrink:0; }
.theme-info { flex:1; }
.theme-name { font-size:14px; font-weight:600; color:var(--color-text-primary); display:block; margin-bottom:4px; }
.theme-desc { font-size:12px; color:var(--color-text-muted); display:block; }
.theme-check { width:24px; height:24px; border-radius:50%; background:var(--color-primary); color:#fff; display:flex; align-items:center; justify-content:center; font-size:12px; font-weight:bold; }
.empty-tip { padding:60px; text-align:center; color:var(--color-text-muted); font-size:14px; }
</style>
