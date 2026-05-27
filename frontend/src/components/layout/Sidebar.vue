<script setup lang="ts">
import { ref, computed } from 'vue'
import { useUserStore } from '@/stores/user'
import SvgIcon from '@/components/common/SvgIcon.vue'

const collapsed  = ref(false)
const userStore  = useUserStore()

const menus = computed(() => {
  const base = [
    {
      section: '概览',
      items: [
        { icon: 'home',     label: '首页大屏',   path: '/pages/dashboard/index', perm: '' },
      ],
    },
    {
      section: '核心业务',
      items: [
        { icon: 'group',    label: '用户管理',   path: '/pages/users/index',       perm: 'user' },
        { icon: 'building', label: '联盟管理',   path: '/pages/orgs/index',        perm: 'org' },
        { icon: 'lock',     label: '权限配置',   path: '/pages/permissions/index', perm: 'permission' },
        { icon: 'order',    label: '订单中心',   path: '/pages/orders/index',      perm: 'order' },
      ],
    },
    {
      section: '财务 & 报表',
      items: [
        { icon: 'money',   label: '财务结算', path: '/pages/finance/index', perm: 'finance' },
        { icon: 'chart',   label: '数据报表', path: '/pages/reports/index', perm: 'report' },
      ],
    },
    {
      section: '系统',
      items: [
        { icon: 'message', label: '消息通知',   path: '/pages/messages/index', perm: 'message' },
        { icon: 'setting', label: '系统设置',   path: '/pages/settings/index', perm: '' },
        ...(userStore.hasPermission('admin:list')
          ? [{ icon: 'user', label: '管理员管理', path: '/pages/admins/index', perm: 'admin:list' }]
          : []),
      ],
    },
  ]
  return base.map(group => ({
    ...group,
    items: group.items.filter(item => !item.perm || userStore.hasPermission(item.perm)),
  })).filter(group => group.items.length > 0)
})

// #ifdef H5
const currentPath = ref(window.location.pathname.replace('/pages', '').replace('/index.html', '') || '/pages/dashboard/index')
// #endif
// #ifndef H5
const currentPath = ref('/pages/dashboard/index')
// #endif

function navigate(path: string) {
  currentPath.value = path
  uni.navigateTo({ url: path })
}
</script>

<template>
  <view class="sidebar" :class="{ collapsed }">
    <view class="sidebar-logo">
      <view class="logo-icon">联</view>
      <view v-if="!collapsed" class="logo-text-wrap">
        <text class="logo-title">联盟管理中心</text>
        <text class="logo-sub">Union Manage Center</text>
      </view>
    </view>

    <view class="collapse-btn" @click="collapsed = !collapsed">
      <SvgIcon :name="collapsed ? 'chevron-right' : 'chevron-left'" />
    </view>

    <scroll-view class="sidebar-nav" scroll-y>
      <template v-for="group in menus" :key="group.section">
        <text v-if="!collapsed" class="nav-section">{{ group.section }}</text>
        <view
          v-for="item in group.items"
          :key="item.path"
          class="nav-item"
          :class="{ active: currentPath.includes(item.path.replace('/pages','').replace('/index','')) }"
          @click="navigate(item.path)"
        >
          <view class="nav-icon">
            <SvgIcon :name="item.icon" />
          </view>
          <text v-if="!collapsed" class="nav-label">{{ item.label }}</text>
          <text v-if="!collapsed && item.badge" class="nav-badge">{{ item.badge }}</text>
        </view>
      </template>
    </scroll-view>

    <view class="sidebar-bottom">
      <view class="user-av">{{ userStore.info?.username?.charAt(0).toUpperCase() || '管' }}</view>
      <view v-if="!collapsed" class="user-info">
        <text class="user-name">{{ userStore.info?.username || '管理员' }}</text>
        <text class="user-role">{{ userStore.info?.email || userStore.info?.role || '' }}</text>
      </view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.sidebar {
  width: var(--sidebar-width);
  min-height: 100vh;
  background: var(--color-sidebar-bg);
  box-shadow: 2px 0 8px rgba(0,0,0,0.06);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  transition: width 0.25s ease;
  overflow: hidden;
  &.collapsed { width: var(--sidebar-collapsed-width); }
}
.sidebar-logo {
  height: var(--topbar-height);
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 16px;
  border-bottom: 1px solid var(--color-sidebar-border);
  white-space: nowrap;
  overflow: hidden;
}
.logo-icon {
  width: 32px; height: 32px; border-radius: 8px; flex-shrink: 0;
  background: linear-gradient(135deg, var(--color-primary-dark), var(--color-primary));
  display: flex; align-items: center; justify-content: center;
  color: #fff; font-size: 14px; font-weight: bold;
}
.logo-title { font-size: 14px; font-weight: 700; color: var(--color-text-primary); display: block; }
.logo-sub { font-size: 10px; color: var(--color-text-muted); display: block; }
.collapse-btn {
  padding: 8px 0;
  text-align: center;
  font-size: 14px;
  color: var(--color-text-muted);
  cursor: pointer;
  border-bottom: 1px solid var(--color-border-light);
  display: flex; align-items: center; justify-content: center;
  &:hover { color: var(--color-primary); }
}
.sidebar-nav { flex: 1; }
.nav-section {
  display: block;
  padding: 10px 16px 4px;
  font-size: 10px;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.08em;
  white-space: nowrap;
}
.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 16px;
  font-size: 13px;
  color: var(--color-sidebar-text);
  cursor: pointer;
  position: relative;
  transition: all 0.15s;
  white-space: nowrap;
  &:hover { background: var(--color-border-light); }
  &.active {
    background: var(--color-sidebar-active-bg);
    color: var(--color-sidebar-active-text);
    font-weight: 600;
    &::before {
      content: '';
      position: absolute; left: 0; top: 50%;
      transform: translateY(-50%);
      width: 3px; height: 20px;
      background: var(--color-primary);
      border-radius: 0 3px 3px 0;
    }
  }
}
.nav-icon {
  font-size: 18px; flex-shrink: 0;
  display: flex; align-items: center; justify-content: center;
  width: 20px; height: 20px;
}
.nav-label { flex: 1; }
.nav-badge {
  background: #ef4444; color: #fff;
  font-size: 10px; padding: 1px 6px; border-radius: 10px;
}
.sidebar-bottom {
  padding: 14px 16px;
  border-top: 1px solid var(--color-sidebar-border);
  display: flex; align-items: center; gap: 10px;
  overflow: hidden;
}
.user-av {
  width: 32px; height: 32px; border-radius: 50%; flex-shrink: 0;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  display: flex; align-items: center; justify-content: center;
  color: #fff; font-size: 12px; font-weight: bold;
}
.user-name { font-size: 12px; font-weight: 600; color: var(--color-text-primary); display: block; white-space: nowrap; }
.user-role { font-size: 10px; color: var(--color-text-muted); display: block; white-space: nowrap; }
</style>
