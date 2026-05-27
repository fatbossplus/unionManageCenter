<script setup lang="ts">
import { ref } from 'vue'
import { useUserStore } from '@/stores/user'
import { post } from '@/api/request'

const form = ref({ username: '', password: '' })
const loading = ref(false)
const userStore = useUserStore()

async function handleLogin() {
  if (!form.value.username || !form.value.password) {
    uni.showToast({ title: '请填写账号和密码', icon: 'none' })
    return
  }
  loading.value = true
  try {
    const res = await post<{ token: string; user: any }>('/auth/login', {
      username: form.value.username,
      password: form.value.password,
    })
    userStore.setToken(res.token)
    userStore.setInfo({
      id: String(res.user.id),
      username: res.user.username,
      role: res.user.role,
      email: res.user.email,
    })
    // 登录后立即拉取权限码，写入 store 供各页面使用
    await userStore.loadPermissions()
    uni.reLaunch({ url: '/pages/dashboard/index' })
  } catch (e: any) {
    uni.showToast({ title: e?.message || '登录失败', icon: 'none' })
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <view class="login-page">
    <view class="login-bg" />
    <view class="login-card">
      <view class="login-logo">
        <view class="logo-icon">联</view>
        <text class="logo-title">联盟管理中心</text>
        <text class="logo-sub">Union Manage Center</text>
      </view>

      <view class="form-group">
        <text class="form-label">账号</text>
        <input
          v-model="form.username"
          class="form-input"
          placeholder="请输入账号"
          @confirm="handleLogin"
        />
      </view>

      <view class="form-group">
        <text class="form-label">密码</text>
        <input
          v-model="form.password"
          class="form-input"
          :password="true"
          placeholder="请输入密码"
          @confirm="handleLogin"
        />
      </view>

      <view class="login-btn" :class="{ loading }" @click="handleLogin">
        {{ loading ? '登录中...' : '登 录' }}
      </view>

      <text class="login-tip">默认账号：admin / admin123</text>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.login-page {
  min-height: 100vh;
  background: var(--color-bg);
  display: flex; align-items: center; justify-content: center;
  position: relative; overflow: hidden;
}
.login-bg {
  position: absolute; inset: 0;
  background: linear-gradient(135deg, var(--color-primary-dark) 0%, var(--color-primary) 50%, #60a5fa 100%);
  opacity: 0.08;
}
.login-card {
  width: 400px;
  background: var(--color-card-bg);
  border-radius: 20px; padding: 44px 40px;
  box-shadow: 0 20px 60px rgba(0,0,0,0.12);
  position: relative; z-index: 1;
}
.login-logo { text-align: center; margin-bottom: 36px; }
.logo-icon {
  width: 60px; height: 60px; border-radius: 16px; margin: 0 auto 14px;
  background: linear-gradient(135deg, var(--color-primary-dark), var(--color-primary));
  display: flex; align-items: center; justify-content: center;
  color: #fff; font-size: 26px; font-weight: bold;
  box-shadow: 0 8px 24px rgba(30, 64, 175, 0.3);
}
.logo-title { font-size: 22px; font-weight: 700; color: var(--color-text-primary); display: block; }
.logo-sub { font-size: 12px; color: var(--color-text-muted); display: block; margin-top: 4px; }
.form-group { margin-bottom: 20px; }
.form-label { font-size: 13px; color: var(--color-text-secondary); display: block; margin-bottom: 6px; font-weight: 500; }
.form-input {
  width: 100%; height: 44px;
  border: 1.5px solid var(--color-border); border-radius: 10px;
  padding: 0 14px; font-size: 14px;
  color: var(--color-text-primary); background: var(--color-card-bg);
  transition: border-color 0.15s;
  &:focus { border-color: var(--color-primary); }
}
.login-btn {
  width: 100%; height: 46px;
  background: linear-gradient(135deg, var(--color-primary-dark), var(--color-primary));
  color: #fff; border-radius: 10px; font-size: 15px; font-weight: 600;
  display: flex; align-items: center; justify-content: center;
  cursor: pointer; margin-top: 8px;
  box-shadow: 0 4px 16px rgba(30, 64, 175, 0.3);
  transition: opacity 0.15s;
  &.loading { opacity: 0.7; cursor: not-allowed; }
  &:hover:not(.loading) { opacity: 0.9; }
}
.login-tip { display: block; text-align: center; margin-top: 16px; font-size: 11px; color: var(--color-text-muted); }
</style>
