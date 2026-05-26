# 联盟管理中心前端 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 使用 UniApp + Vue3 + Vite 构建联盟管理中心 PC 端管理后台，包含三套主题切换、首页大屏、以及所有业务模块列表页。

**Architecture:** 采用 UniApp 条件编译目标 H5，Vue3 Composition API + `<script setup>` 语法，Pinia 管理全局状态（主题/用户/字典）。页面按模块拆分，公共组件（KpiCard、FilterPanel、DataTable 等）统一封装复用，CSS 变量驱动三套主题。

**Tech Stack:** UniApp 3.x · Vue 3 · Vite · Pinia · uni-ui · uCharts · SCSS · ESLint + Prettier

---

## 文件总览

```
frontend/
├── src/
│   ├── pages/
│   │   ├── dashboard/index.vue          # 首页大屏
│   │   ├── users/index.vue              # 用户列表
│   │   ├── users/detail.vue             # 用户详情
│   │   ├── orgs/index.vue               # 联盟列表
│   │   ├── orgs/detail.vue              # 联盟详情
│   │   ├── permissions/index.vue        # 权限配置
│   │   ├── orders/index.vue             # 订单列表
│   │   ├── orders/detail.vue            # 订单详情
│   │   ├── finance/index.vue            # 财务结算
│   │   ├── reports/index.vue            # 数据报表
│   │   ├── messages/index.vue           # 消息通知
│   │   ├── settings/index.vue           # 系统设置
│   │   └── login/index.vue              # 登录页
│   ├── components/
│   │   ├── layout/
│   │   │   ├── AppLayout.vue            # 整体布局容器
│   │   │   ├── Sidebar.vue              # 侧边栏（可折叠）
│   │   │   └── Topbar.vue              # 顶部栏
│   │   ├── common/
│   │   │   ├── KpiCard.vue             # KPI 数字卡片
│   │   │   ├── FilterPanel.vue         # 筛选面板
│   │   │   ├── DataTable.vue           # 数据表格
│   │   │   ├── Pagination.vue          # 分页
│   │   │   ├── StatusBadge.vue         # 彩色状态徽章
│   │   │   └── ThemeSwitcher.vue       # 主题切换浮层
│   │   └── charts/
│   │       ├── LineChart.vue           # 折线图
│   │       ├── DonutChart.vue          # 圆环图
│   │       ├── BarChart.vue            # 柱状图
│   │       └── Sparkline.vue          # 迷你走势图
│   ├── stores/
│   │   ├── theme.ts                    # 主题状态（A/B/C）
│   │   ├── user.ts                     # 登录用户信息
│   │   └── dict.ts                     # 数据字典缓存
│   ├── api/
│   │   ├── request.ts                  # uni.request 封装
│   │   ├── user.ts                     # 用户接口
│   │   ├── org.ts                      # 联盟接口
│   │   ├── order.ts                    # 订单接口
│   │   ├── finance.ts                  # 财务接口
│   │   ├── report.ts                   # 报表接口
│   │   ├── message.ts                  # 消息接口
│   │   └── dashboard.ts                # 大屏统计接口
│   ├── styles/
│   │   ├── themes/
│   │   │   ├── theme-a.scss            # 深色科技主题变量
│   │   │   ├── theme-b.scss            # 纯净浅蓝主题变量（默认）
│   │   │   └── theme-c.scss            # 渐变紫金主题变量
│   │   ├── variables.scss              # CSS 变量声明
│   │   └── global.scss                 # reset + 全局样式
│   ├── utils/
│   │   ├── format.ts                   # 数字/时间格式化
│   │   ├── auth.ts                     # token 存取/清除
│   │   └── countup.ts                  # 数字滚动动画
│   ├── App.vue
│   ├── main.ts
│   └── pages.json
├── package.json
└── vite.config.ts
```

---

## Task 1: 初始化 UniApp + Vue3 + Vite 项目

**Files:**
- Create: `frontend/package.json`
- Create: `frontend/vite.config.ts`
- Create: `frontend/src/main.ts`
- Create: `frontend/src/App.vue`
- Create: `frontend/src/pages.json`
- Create: `frontend/src/manifest.json`

- [ ] **Step 1: 在项目根目录创建 frontend 目录并初始化**

```bash
cd /Users/fatboss/gowork/src/unionManageCenter
npx degit dcloudio/uni-preset-vue#vite-ts frontend
cd frontend
npm install
```

- [ ] **Step 2: 安装所有依赖**

```bash
cd /Users/fatboss/gowork/src/unionManageCenter/frontend
npm install pinia @dcloudio/uni-ui sass
npm install -D eslint prettier eslint-plugin-vue @vue/eslint-config-prettier
npm install uview-plus   # 或直接使用 uni-ui 内置组件
```

- [ ] **Step 3: 验证项目能启动**

```bash
cd /Users/fatboss/gowork/src/unionManageCenter/frontend
npm run dev:h5
```

预期：终端输出 `Local: http://localhost:5173`，浏览器能看到默认页面。

- [ ] **Step 4: 配置 vite.config.ts，添加路径别名**

`frontend/vite.config.ts` 完整内容：
```typescript
import { defineConfig } from 'vite'
import uni from '@dcloudio/vite-plugin-uni'
import { resolve } from 'path'

export default defineConfig({
  plugins: [uni()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
```

- [ ] **Step 5: 配置 pages.json 路由**

`frontend/src/pages.json` 完整内容：
```json
{
  "pages": [
    { "path": "pages/login/index", "style": { "navigationBarTitleText": "登录" } },
    { "path": "pages/dashboard/index", "style": { "navigationBarTitleText": "首页大屏" } },
    { "path": "pages/users/index", "style": { "navigationBarTitleText": "用户管理" } },
    { "path": "pages/users/detail", "style": { "navigationBarTitleText": "用户详情" } },
    { "path": "pages/orgs/index", "style": { "navigationBarTitleText": "联盟管理" } },
    { "path": "pages/orgs/detail", "style": { "navigationBarTitleText": "联盟详情" } },
    { "path": "pages/permissions/index", "style": { "navigationBarTitleText": "权限配置" } },
    { "path": "pages/orders/index", "style": { "navigationBarTitleText": "订单中心" } },
    { "path": "pages/orders/detail", "style": { "navigationBarTitleText": "订单详情" } },
    { "path": "pages/finance/index", "style": { "navigationBarTitleText": "财务结算" } },
    { "path": "pages/reports/index", "style": { "navigationBarTitleText": "数据报表" } },
    { "path": "pages/messages/index", "style": { "navigationBarTitleText": "消息通知" } },
    { "path": "pages/settings/index", "style": { "navigationBarTitleText": "系统设置" } }
  ],
  "globalStyle": {
    "navigationBarTextStyle": "black",
    "navigationBarTitleText": "联盟管理中心",
    "navigationBarBackgroundColor": "#ffffff",
    "backgroundColor": "#f0f4f8"
  }
}
```

- [ ] **Step 6: 初始化 main.ts**

`frontend/src/main.ts`:
```typescript
import { createSSRApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'

export function createApp() {
  const app = createSSRApp(App)
  app.use(createPinia())
  return { app }
}
```

- [ ] **Step 7: Commit**

```bash
cd /Users/fatboss/gowork/src/unionManageCenter
git init
git add frontend/
git commit -m "feat: init UniApp Vue3 Vite frontend project"
```

---

## Task 2: 主题系统（CSS 变量 + Pinia Store）

**Files:**
- Create: `frontend/src/styles/variables.scss`
- Create: `frontend/src/styles/themes/theme-a.scss`
- Create: `frontend/src/styles/themes/theme-b.scss`
- Create: `frontend/src/styles/themes/theme-c.scss`
- Create: `frontend/src/styles/global.scss`
- Create: `frontend/src/stores/theme.ts`

- [ ] **Step 1: 定义 CSS 变量声明文件**

`frontend/src/styles/variables.scss`:
```scss
// 所有组件使用这些变量，主题切换时只改 :root 上的值
:root {
  --color-primary: #1e40af;
  --color-primary-light: #eff6ff;
  --color-primary-dark: #1e3a8a;
  --color-bg: #f0f4f8;
  --color-sidebar-bg: #ffffff;
  --color-sidebar-text: #6b7280;
  --color-sidebar-active-bg: #eff6ff;
  --color-sidebar-active-text: #1e40af;
  --color-sidebar-border: #f3f4f6;
  --color-card-bg: #ffffff;
  --color-card-shadow: 0 1px 4px rgba(0,0,0,0.06);
  --color-text-primary: #111827;
  --color-text-secondary: #6b7280;
  --color-text-muted: #9ca3af;
  --color-border: #e5e7eb;
  --color-border-light: #f3f4f6;
  --color-topbar-bg: #ffffff;
  --sidebar-width: 220px;
  --sidebar-collapsed-width: 64px;
  --topbar-height: 60px;
}
```

- [ ] **Step 2: 主题 A（深色科技）**

`frontend/src/styles/themes/theme-a.scss`:
```scss
[data-theme="a"] {
  --color-primary: #3b82f6;
  --color-primary-light: rgba(59,130,246,0.15);
  --color-primary-dark: #1d4ed8;
  --color-bg: #0d1117;
  --color-sidebar-bg: #161b22;
  --color-sidebar-text: #8b949e;
  --color-sidebar-active-bg: rgba(59,130,246,0.2);
  --color-sidebar-active-text: #3b82f6;
  --color-sidebar-border: #21262d;
  --color-card-bg: #161b22;
  --color-card-shadow: 0 1px 4px rgba(0,0,0,0.3);
  --color-text-primary: #c9d1d9;
  --color-text-secondary: #8b949e;
  --color-text-muted: #484f58;
  --color-border: #30363d;
  --color-border-light: #21262d;
  --color-topbar-bg: #161b22;
}
```

- [ ] **Step 3: 主题 B（纯净浅蓝，默认）**

`frontend/src/styles/themes/theme-b.scss`:
```scss
[data-theme="b"] {
  --color-primary: #1e40af;
  --color-primary-light: #eff6ff;
  --color-primary-dark: #1e3a8a;
  --color-bg: #f0f4f8;
  --color-sidebar-bg: #ffffff;
  --color-sidebar-text: #6b7280;
  --color-sidebar-active-bg: #eff6ff;
  --color-sidebar-active-text: #1e40af;
  --color-sidebar-border: #f3f4f6;
  --color-card-bg: #ffffff;
  --color-card-shadow: 0 1px 4px rgba(0,0,0,0.06);
  --color-text-primary: #111827;
  --color-text-secondary: #6b7280;
  --color-text-muted: #9ca3af;
  --color-border: #e5e7eb;
  --color-border-light: #f3f4f6;
  --color-topbar-bg: #ffffff;
}
```

- [ ] **Step 4: 主题 C（渐变紫金）**

`frontend/src/styles/themes/theme-c.scss`:
```scss
[data-theme="c"] {
  --color-primary: #7c3aed;
  --color-primary-light: #f5f3ff;
  --color-primary-dark: #5b21b6;
  --color-bg: #fafafa;
  --color-sidebar-bg: #ffffff;
  --color-sidebar-text: #6b7280;
  --color-sidebar-active-bg: #f5f3ff;
  --color-sidebar-active-text: #7c3aed;
  --color-sidebar-border: #f3f4f6;
  --color-card-bg: #ffffff;
  --color-card-shadow: 0 2px 8px rgba(124,58,237,0.08);
  --color-text-primary: #111827;
  --color-text-secondary: #6b7280;
  --color-text-muted: #9ca3af;
  --color-border: #e5e7eb;
  --color-border-light: #f3f4f6;
  --color-topbar-bg: #ffffff;
}
```

- [ ] **Step 5: 全局样式**

`frontend/src/styles/global.scss`:
```scss
@import './variables.scss';
@import './themes/theme-a.scss';
@import './themes/theme-b.scss';
@import './themes/theme-c.scss';

* { margin: 0; padding: 0; box-sizing: border-box; }

body {
  font-family: -apple-system, 'PingFang SC', 'Microsoft YaHei', sans-serif;
  background: var(--color-bg);
  color: var(--color-text-primary);
  transition: background 0.2s, color 0.2s;
}

.card {
  background: var(--color-card-bg);
  border-radius: 12px;
  box-shadow: var(--color-card-shadow);
}

::-webkit-scrollbar { width: 5px; height: 5px; }
::-webkit-scrollbar-thumb { background: var(--color-border); border-radius: 3px; }
::-webkit-scrollbar-track { background: transparent; }
```

- [ ] **Step 6: 主题 Pinia Store**

`frontend/src/stores/theme.ts`:
```typescript
import { defineStore } from 'pinia'
import { ref } from 'vue'

export type ThemeKey = 'a' | 'b' | 'c'

export const useThemeStore = defineStore('theme', () => {
  const current = ref<ThemeKey>(
    (uni.getStorageSync('theme') as ThemeKey) || 'b'
  )

  function apply(theme: ThemeKey) {
    current.value = theme
    uni.setStorageSync('theme', theme)
    // H5 环境下修改 document 属性
    // #ifdef H5
    document.documentElement.setAttribute('data-theme', theme)
    // #endif
  }

  // 启动时恢复主题
  function init() {
    apply(current.value)
  }

  return { current, apply, init }
})
```

- [ ] **Step 7: 在 App.vue 中引入全局样式并初始化主题**

`frontend/src/App.vue`:
```vue
<script setup lang="ts">
import { onLaunch } from '@dcloudio/uni-app'
import { useThemeStore } from '@/stores/theme'

onLaunch(() => {
  useThemeStore().init()
})
</script>

<style lang="scss">
@import '@/styles/global.scss';
</style>
```

- [ ] **Step 8: Commit**

```bash
git add frontend/src/styles/ frontend/src/stores/theme.ts frontend/src/App.vue
git commit -m "feat: add three-theme CSS variable system with Pinia store"
```

---

## Task 3: 布局组件（Sidebar + Topbar + AppLayout）

**Files:**
- Create: `frontend/src/components/layout/Sidebar.vue`
- Create: `frontend/src/components/layout/Topbar.vue`
- Create: `frontend/src/components/layout/AppLayout.vue`

- [ ] **Step 1: 侧边栏 Sidebar.vue**

`frontend/src/components/layout/Sidebar.vue`:
```vue
<script setup lang="ts">
import { ref, computed } from 'vue'
import { useThemeStore } from '@/stores/theme'

const themeStore = useThemeStore()
const collapsed = ref(false)

const menus = [
  { section: '概览', items: [
    { icon: '🏠', label: '首页大屏', path: '/pages/dashboard/index' },
  ]},
  { section: '核心业务', items: [
    { icon: '👥', label: '用户管理', path: '/pages/users/index' },
    { icon: '🏢', label: '联盟管理', path: '/pages/orgs/index' },
    { icon: '🔐', label: '权限配置', path: '/pages/permissions/index' },
    { icon: '📦', label: '订单中心', path: '/pages/orders/index', badge: 12 },
  ]},
  { section: '财务 & 报表', items: [
    { icon: '💰', label: '财务结算', path: '/pages/finance/index' },
    { icon: '📊', label: '数据报表', path: '/pages/reports/index' },
  ]},
  { section: '系统', items: [
    { icon: '💬', label: '消息通知', path: '/pages/messages/index', badge: 3 },
    { icon: '⚙️', label: '系统设置', path: '/pages/settings/index' },
  ]},
]

const currentPath = ref('/pages/dashboard/index')

function navigate(path: string) {
  currentPath.value = path
  uni.navigateTo({ url: path })
}
</script>

<template>
  <view class="sidebar" :class="{ collapsed }">
    <!-- Logo -->
    <view class="sidebar-logo">
      <view class="logo-icon">联</view>
      <view v-if="!collapsed" class="logo-text-wrap">
        <text class="logo-title">联盟管理中心</text>
        <text class="logo-sub">Union Manage Center</text>
      </view>
    </view>

    <!-- 折叠按钮 -->
    <view class="collapse-btn" @click="collapsed = !collapsed">
      {{ collapsed ? '→' : '←' }}
    </view>

    <!-- 导航菜单 -->
    <scroll-view class="sidebar-nav" scroll-y>
      <template v-for="group in menus" :key="group.section">
        <text v-if="!collapsed" class="nav-section">{{ group.section }}</text>
        <view
          v-for="item in group.items"
          :key="item.path"
          class="nav-item"
          :class="{ active: currentPath === item.path }"
          @click="navigate(item.path)"
        >
          <text class="nav-icon">{{ item.icon }}</text>
          <text v-if="!collapsed" class="nav-label">{{ item.label }}</text>
          <text v-if="!collapsed && item.badge" class="nav-badge">{{ item.badge }}</text>
        </view>
      </template>
    </scroll-view>
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

  &.collapsed { width: var(--sidebar-collapsed-width); }
}
.sidebar-logo {
  height: var(--topbar-height);
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 16px;
  border-bottom: 1px solid var(--color-sidebar-border);
  overflow: hidden;
}
.logo-icon {
  width: 32px; height: 32px;
  border-radius: 8px;
  background: linear-gradient(135deg, var(--color-primary-dark), var(--color-primary));
  display: flex; align-items: center; justify-content: center;
  color: #fff; font-size: 14px; font-weight: bold; flex-shrink: 0;
}
.logo-title { font-size: 14px; font-weight: 700; color: var(--color-text-primary); display: block; }
.logo-sub { font-size: 10px; color: var(--color-text-muted); display: block; }
.collapse-btn {
  padding: 8px 16px;
  text-align: right;
  font-size: 12px;
  color: var(--color-text-muted);
  cursor: pointer;
  border-bottom: 1px solid var(--color-border-light);
}
.sidebar-nav { flex: 1; padding: 8px 0; }
.nav-section {
  display: block;
  padding: 10px 16px 4px;
  font-size: 10px;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.08em;
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
.nav-icon { font-size: 16px; flex-shrink: 0; }
.nav-label { flex: 1; white-space: nowrap; }
.nav-badge {
  background: #ef4444; color: #fff;
  font-size: 10px; padding: 1px 6px;
  border-radius: 10px; margin-left: auto;
}
</style>
```

- [ ] **Step 2: 顶部栏 Topbar.vue**

`frontend/src/components/layout/Topbar.vue`:
```vue
<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useThemeStore } from '@/stores/theme'
import type { ThemeKey } from '@/stores/theme'

defineProps<{ breadcrumbs: { label: string; path?: string }[] }>()

const themeStore = useThemeStore()
const timeStr = ref('')
let timer: ReturnType<typeof setInterval>

function updateTime() {
  const now = new Date()
  const pad = (n: number) => String(n).padStart(2, '0')
  timeStr.value = `${now.getFullYear()}-${pad(now.getMonth()+1)}-${pad(now.getDate())} ${pad(now.getHours())}:${pad(now.getMinutes())}:${pad(now.getSeconds())}`
}

onMounted(() => { updateTime(); timer = setInterval(updateTime, 1000) })
onUnmounted(() => clearInterval(timer))

const themes: { key: ThemeKey; color: string; label: string }[] = [
  { key: 'a', color: '#0d1117', label: 'A' },
  { key: 'b', color: '#1e40af', label: 'B' },
  { key: 'c', color: 'linear-gradient(135deg,#7c3aed,#f59e0b)', label: 'C' },
]
</script>

<template>
  <view class="topbar">
    <view class="breadcrumb">
      <template v-for="(crumb, idx) in breadcrumbs" :key="idx">
        <text v-if="idx > 0" class="sep">›</text>
        <text :class="idx === breadcrumbs.length - 1 ? 'cur' : 'parent'">{{ crumb.label }}</text>
      </template>
    </view>

    <view class="topbar-right">
      <!-- 时钟 -->
      <text class="time-display">{{ timeStr }}</text>

      <!-- 消息 -->
      <view class="icon-btn">
        🔔
        <view class="notif-dot" />
      </view>

      <!-- 主题切换 -->
      <view class="theme-switcher">
        <text class="theme-label">主题</text>
        <view
          v-for="t in themes"
          :key="t.key"
          class="theme-btn"
          :class="{ active: themeStore.current === t.key }"
          :style="{ background: t.color }"
          @click="themeStore.apply(t.key)"
        >{{ t.label }}</view>
      </view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.topbar {
  height: var(--topbar-height);
  background: var(--color-topbar-bg);
  border-bottom: 1px solid var(--color-border-light);
  display: flex;
  align-items: center;
  padding: 0 24px;
  gap: 16px;
  position: sticky;
  top: 0;
  z-index: 10;
}
.breadcrumb {
  display: flex; align-items: center; gap: 6px;
  .sep { color: var(--color-border); }
  .parent { font-size: 13px; color: var(--color-text-muted); }
  .cur { font-size: 13px; color: var(--color-text-primary); font-weight: 600; }
}
.topbar-right {
  margin-left: auto;
  display: flex; align-items: center; gap: 12px;
}
.time-display {
  font-size: 12px; color: var(--color-text-secondary);
  background: var(--color-border-light);
  padding: 5px 12px; border-radius: 8px;
  border: 1px solid var(--color-border);
}
.icon-btn {
  width: 32px; height: 32px; border-radius: 8px;
  background: var(--color-border-light);
  border: 1px solid var(--color-border);
  display: flex; align-items: center; justify-content: center;
  font-size: 14px; cursor: pointer; position: relative;
}
.notif-dot {
  position: absolute; top: 6px; right: 6px;
  width: 7px; height: 7px;
  background: #ef4444; border-radius: 50%;
  border: 1.5px solid var(--color-topbar-bg);
}
.theme-switcher {
  display: flex; align-items: center; gap: 6px;
  background: var(--color-card-bg);
  border-radius: 50px; padding: 4px 10px;
  box-shadow: var(--color-card-shadow);
  border: 1px solid var(--color-border);
}
.theme-label { font-size: 11px; color: var(--color-text-muted); }
.theme-btn {
  width: 24px; height: 24px; border-radius: 50%; cursor: pointer;
  border: 2px solid transparent;
  display: flex; align-items: center; justify-content: center;
  font-size: 10px; font-weight: 700; color: #fff;
  transition: all 0.2s;
  &.active { border-color: var(--color-text-primary); transform: scale(1.2); }
}
</style>
```

- [ ] **Step 3: 整体布局容器 AppLayout.vue**

`frontend/src/components/layout/AppLayout.vue`:
```vue
<script setup lang="ts">
import Sidebar from './Sidebar.vue'
import Topbar from './Topbar.vue'

defineProps<{
  breadcrumbs: { label: string; path?: string }[]
}>()
</script>

<template>
  <view class="app-layout">
    <Sidebar />
    <view class="layout-main">
      <Topbar :breadcrumbs="breadcrumbs" />
      <view class="layout-content">
        <slot />
      </view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.app-layout {
  display: flex;
  min-height: 100vh;
  background: var(--color-bg);
}
.layout-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  overflow: hidden;
}
.layout-content {
  flex: 1;
  overflow-y: auto;
  padding: 20px 24px;
}
</style>
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/layout/
git commit -m "feat: add Sidebar, Topbar, AppLayout components"
```

---

## Task 4: 公共组件 KpiCard + StatusBadge

**Files:**
- Create: `frontend/src/components/common/KpiCard.vue`
- Create: `frontend/src/components/common/StatusBadge.vue`

- [ ] **Step 1: KpiCard.vue**

`frontend/src/components/common/KpiCard.vue`:
```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { countUp } from '@/utils/countup'

const props = defineProps<{
  icon: string
  label: string
  value: number
  unit?: string
  trend?: { dir: 'up' | 'down'; text: string }
  iconBg?: string
  sparkline?: number[]
}>()

const displayValue = ref(0)

onMounted(() => {
  countUp(0, props.value, 800, (v) => { displayValue.value = v })
})
</script>

<template>
  <view class="kpi-card card">
    <view class="kpi-top">
      <text class="kpi-label">{{ label }}</text>
      <view class="kpi-icon" :style="{ background: iconBg || 'var(--color-primary-light)' }">
        {{ icon }}
      </view>
    </view>
    <text class="kpi-num">{{ unit }}{{ displayValue.toLocaleString() }}</text>
    <view v-if="trend" class="kpi-trend">
      <text :class="trend.dir === 'up' ? 'trend-up' : 'trend-down'">
        {{ trend.dir === 'up' ? '↑' : '↓' }} {{ trend.text }}
      </text>
    </view>
    <view v-if="sparkline?.length" class="sparkline">
      <view
        v-for="(v, i) in sparkline"
        :key="i"
        class="spark-bar"
        :style="{ height: (v / Math.max(...sparkline) * 100) + '%' }"
      />
    </view>
  </view>
</template>

<style lang="scss" scoped>
.kpi-card {
  padding: 18px 20px;
  display: flex; flex-direction: column; gap: 10px;
}
.kpi-top {
  display: flex; align-items: flex-start; justify-content: space-between;
}
.kpi-label { font-size: 12px; color: var(--color-text-muted); }
.kpi-icon {
  width: 36px; height: 36px; border-radius: 9px;
  display: flex; align-items: center; justify-content: center; font-size: 16px;
}
.kpi-num { font-size: 26px; font-weight: 700; color: var(--color-text-primary); line-height: 1; }
.kpi-trend { display: flex; align-items: center; gap: 6px; font-size: 12px; }
.trend-up { color: #10b981; }
.trend-down { color: #ef4444; }
.sparkline {
  height: 28px; display: flex; align-items: flex-end; gap: 2px;
}
.spark-bar {
  flex: 1; border-radius: 2px 2px 0 0;
  background: linear-gradient(to top, var(--color-primary-dark), var(--color-primary));
  opacity: 0.7;
}
</style>
```

- [ ] **Step 2: StatusBadge.vue**

`frontend/src/components/common/StatusBadge.vue`:
```vue
<script setup lang="ts">
defineProps<{
  status: 'success' | 'warning' | 'danger' | 'info' | 'default'
  label: string
}>()

const colorMap = {
  success: { bg: '#f0fdf4', color: '#16a34a' },
  warning: { bg: '#fffbeb', color: '#d97706' },
  danger:  { bg: '#fef2f2', color: '#dc2626' },
  info:    { bg: '#eff6ff', color: '#1e40af' },
  default: { bg: '#f9fafb', color: '#6b7280' },
}
</script>

<template>
  <view class="badge" :style="{ background: colorMap[status].bg, color: colorMap[status].color }">
    <view class="dot" :style="{ background: colorMap[status].color }" />
    <text>{{ label }}</text>
  </view>
</template>

<style lang="scss" scoped>
.badge {
  display: inline-flex; align-items: center; gap: 4px;
  padding: 3px 9px; border-radius: 5px; font-size: 11px; font-weight: 500;
}
.dot { width: 5px; height: 5px; border-radius: 50%; flex-shrink: 0; }
</style>
```

- [ ] **Step 3: countup 工具函数**

`frontend/src/utils/countup.ts`:
```typescript
export function countUp(
  from: number,
  to: number,
  duration: number,
  onUpdate: (value: number) => void
) {
  const start = performance.now()
  const step = (now: number) => {
    const elapsed = now - start
    const progress = Math.min(elapsed / duration, 1)
    // easeOutQuart
    const eased = 1 - Math.pow(1 - progress, 4)
    onUpdate(Math.floor(from + (to - from) * eased))
    if (progress < 1) requestAnimationFrame(step)
  }
  requestAnimationFrame(step)
}
```

- [ ] **Step 4: format 工具函数**

`frontend/src/utils/format.ts`:
```typescript
export function formatNumber(n: number): string {
  if (n >= 10000) return (n / 10000).toFixed(1) + 'w'
  return n.toLocaleString()
}

export function formatMoney(n: number): string {
  if (n >= 10000) return '¥' + (n / 10000).toFixed(1) + 'K'
  return '¥' + n.toFixed(2)
}

export function formatTime(ts: number): string {
  const diff = Date.now() - ts
  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return Math.floor(diff / 60000) + '分钟前'
  if (diff < 86400000) return Math.floor(diff / 3600000) + '小时前'
  return Math.floor(diff / 86400000) + '天前'
}
```

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/common/KpiCard.vue \
        frontend/src/components/common/StatusBadge.vue \
        frontend/src/utils/
git commit -m "feat: add KpiCard, StatusBadge components and utils"
```

---

## Task 5: 公共组件 FilterPanel + Pagination

**Files:**
- Create: `frontend/src/components/common/FilterPanel.vue`
- Create: `frontend/src/components/common/Pagination.vue`

- [ ] **Step 1: FilterPanel.vue**

`frontend/src/components/common/FilterPanel.vue`:
```vue
<script setup lang="ts">
import { ref, reactive } from 'vue'

export interface FilterField {
  key: string
  label: string
  type: 'input' | 'select' | 'daterange'
  placeholder?: string
  options?: { label: string; value: string }[]
}

export interface QuickTag {
  key: string
  label: string
  color: string
  params: Record<string, unknown>
}

const props = defineProps<{
  fields: FilterField[]
  quickTags?: QuickTag[]
}>()

const emit = defineEmits<{
  search: [params: Record<string, unknown>]
  reset: []
  export: []
}>()

const form = reactive<Record<string, unknown>>({})
const activeTag = ref<string | null>(null)
const expanded = ref(true)

function selectTag(tag: QuickTag) {
  if (activeTag.value === tag.key) {
    activeTag.value = null
    Object.keys(tag.params).forEach(k => delete form[k])
  } else {
    activeTag.value = tag.key
    Object.assign(form, tag.params)
  }
}

function handleSearch() {
  emit('search', { ...form })
}

function handleReset() {
  Object.keys(form).forEach(k => delete form[k])
  activeTag.value = null
  emit('reset')
}
</script>

<template>
  <view class="filter-panel card">
    <view class="fp-header">
      <view class="fp-title">
        <text>🔍 筛选条件</text>
        <view v-if="Object.keys(form).length" class="filter-count">
          {{ Object.keys(form).filter(k => form[k]).length }}
        </view>
      </view>
      <text class="fp-toggle" @click="expanded = !expanded">
        {{ expanded ? '▲ 收起' : '▼ 展开' }}
      </text>
    </view>

    <view v-if="expanded">
      <!-- 筛选字段网格 -->
      <view class="fp-fields">
        <view v-for="field in fields" :key="field.key" class="fp-group">
          <text class="fp-label">{{ field.label }}</text>
          <view class="fp-input">
            <input
              v-if="field.type === 'input'"
              v-model="form[field.key] as string"
              :placeholder="field.placeholder || '请输入'"
              class="fp-input-inner"
            />
            <picker
              v-else-if="field.type === 'select'"
              :range="field.options || []"
              range-key="label"
              @change="(e: any) => form[field.key] = field.options?.[e.detail.value]?.value"
            >
              <text class="fp-select-text">
                {{ (field.options || []).find(o => o.value === form[field.key])?.label || field.placeholder || '请选择' }}
              </text>
            </picker>
          </view>
        </view>
      </view>

      <!-- 快捷标签 -->
      <view v-if="quickTags?.length" class="fp-tags">
        <text class="tag-label">快捷筛选：</text>
        <view
          v-for="tag in quickTags"
          :key="tag.key"
          class="qf-tag"
          :class="{ active: activeTag === tag.key }"
          @click="selectTag(tag)"
        >
          <view class="qf-dot" :style="{ background: tag.color }" />
          <text>{{ tag.label }}</text>
        </view>
      </view>

      <!-- 操作按钮 -->
      <view class="fp-actions">
        <view class="btn btn-primary" @click="handleSearch">🔍 查询</view>
        <view class="btn btn-outline" @click="handleReset">↺ 重置</view>
        <view class="fp-actions-right">
          <view class="btn btn-outline" @click="emit('export')">📤 导出 Excel</view>
        </view>
      </view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.filter-panel { padding: 18px 20px; margin-bottom: 16px; }
.fp-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 14px; }
.fp-title { display: flex; align-items: center; gap: 6px; font-size: 13px; font-weight: 600; color: var(--color-text-primary); }
.filter-count { background: #ef4444; color: #fff; font-size: 10px; padding: 1px 5px; border-radius: 4px; }
.fp-toggle { font-size: 12px; color: var(--color-primary); cursor: pointer; }
.fp-fields { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; margin-bottom: 12px; }
.fp-group { display: flex; flex-direction: column; gap: 5px; }
.fp-label { font-size: 11px; color: var(--color-text-secondary); font-weight: 500; }
.fp-input {
  height: 34px; border: 1px solid var(--color-border); border-radius: 7px;
  padding: 0 10px; background: var(--color-card-bg);
  display: flex; align-items: center;
  &:focus-within { border-color: var(--color-primary); }
}
.fp-input-inner { border: none; outline: none; background: transparent; font-size: 13px; color: var(--color-text-primary); width: 100%; }
.fp-select-text { font-size: 13px; color: var(--color-text-secondary); }
.fp-tags { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; margin-bottom: 14px; }
.tag-label { font-size: 11px; color: var(--color-text-muted); }
.qf-tag {
  height: 26px; padding: 0 10px; border-radius: 5px;
  border: 1px solid var(--color-border); font-size: 11px;
  color: var(--color-text-secondary); cursor: pointer;
  display: flex; align-items: center; gap: 4px;
  background: var(--color-card-bg);
  &.active { background: var(--color-primary-light); border-color: var(--color-primary); color: var(--color-primary); font-weight: 600; }
}
.qf-dot { width: 6px; height: 6px; border-radius: 50%; }
.fp-actions { display: flex; align-items: center; gap: 8px; }
.fp-actions-right { margin-left: auto; display: flex; gap: 8px; }
.btn {
  height: 34px; border-radius: 7px; border: none; cursor: pointer;
  font-size: 13px; font-weight: 500; display: flex; align-items: center;
  gap: 6px; padding: 0 16px;
}
.btn-primary { background: var(--color-primary); color: #fff; }
.btn-outline { background: var(--color-card-bg); color: var(--color-text-secondary); border: 1px solid var(--color-border); }
</style>
```

- [ ] **Step 2: Pagination.vue**

`frontend/src/components/common/Pagination.vue`:
```vue
<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  total: number
  page: number
  pageSize: number
}>()

const emit = defineEmits<{
  pageChange: [page: number]
  pageSizeChange: [size: number]
}>()

const totalPages = computed(() => Math.ceil(props.total / props.pageSize))

const visiblePages = computed(() => {
  const pages: (number | '...')[] = []
  const p = props.page
  const t = totalPages.value
  if (t <= 7) {
    for (let i = 1; i <= t; i++) pages.push(i)
  } else {
    pages.push(1)
    if (p > 3) pages.push('...')
    for (let i = Math.max(2, p-1); i <= Math.min(t-1, p+1); i++) pages.push(i)
    if (p < t - 2) pages.push('...')
    pages.push(t)
  }
  return pages
})

const pageSizeOptions = [10, 20, 50, 100]
</script>

<template>
  <view class="pagination-wrap">
    <text class="page-info">
      共 <text class="em">{{ total }}</text> 条，第
      <text class="em">{{ page }}</text> /
      <text class="em">{{ totalPages }}</text> 页
    </text>
    <view class="page-btns">
      <view class="page-btn" :class="{ disabled: page <= 1 }" @click="page > 1 && emit('pageChange', page - 1)">‹</view>
      <view
        v-for="p in visiblePages"
        :key="p"
        class="page-btn"
        :class="{ active: p === page, ellipsis: p === '...' }"
        @click="typeof p === 'number' && emit('pageChange', p)"
      >{{ p }}</view>
      <view class="page-btn" :class="{ disabled: page >= totalPages }" @click="page < totalPages && emit('pageChange', page + 1)">›</view>
    </view>
    <picker :range="pageSizeOptions" @change="(e: any) => emit('pageSizeChange', pageSizeOptions[e.detail.value])">
      <view class="page-size-picker">每页 {{ pageSize }} 条 ▾</view>
    </picker>
  </view>
</template>

<style lang="scss" scoped>
.pagination-wrap {
  display: flex; align-items: center; justify-content: space-between;
  padding: 14px 16px; border-top: 1px solid var(--color-border-light);
  background: var(--color-border-light);
}
.page-info { font-size: 12px; color: var(--color-text-muted); }
.em { color: var(--color-text-primary); font-weight: 600; }
.page-btns { display: flex; align-items: center; gap: 6px; }
.page-btn {
  min-width: 32px; height: 32px; border-radius: 7px;
  border: 1px solid var(--color-border); background: var(--color-card-bg);
  font-size: 13px; color: var(--color-text-primary);
  display: flex; align-items: center; justify-content: center;
  cursor: pointer; padding: 0 8px;
  &:hover { border-color: var(--color-primary); color: var(--color-primary); }
  &.active { background: var(--color-primary); color: #fff; border-color: var(--color-primary); }
  &.disabled { opacity: 0.4; cursor: not-allowed; }
  &.ellipsis { border: none; background: transparent; cursor: default; }
}
.page-size-picker {
  font-size: 12px; color: var(--color-text-secondary);
  border: 1px solid var(--color-border); padding: 5px 10px;
  border-radius: 6px; background: var(--color-card-bg);
}
</style>
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/common/FilterPanel.vue \
        frontend/src/components/common/Pagination.vue
git commit -m "feat: add FilterPanel and Pagination components"
```

---

## Task 6: 图表组件（uCharts 封装）

**Files:**
- Create: `frontend/src/components/charts/LineChart.vue`
- Create: `frontend/src/components/charts/DonutChart.vue`
- Create: `frontend/src/components/charts/Sparkline.vue`

- [ ] **Step 1: 安装 uCharts**

```bash
cd /Users/fatboss/gowork/src/unionManageCenter/frontend
npm install @qiun/uni-ucharts
```

- [ ] **Step 2: LineChart.vue**

`frontend/src/components/charts/LineChart.vue`:
```vue
<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
// #ifdef H5
import uCharts from '@qiun/uni-ucharts'
// #endif

const props = defineProps<{
  categories: string[]
  series: { name: string; data: number[]; color?: string }[]
  height?: number
}>()

const canvasId = `line-${Math.random().toString(36).slice(2)}`
const canvasRef = ref()

function render() {
  // #ifdef H5
  new uCharts({
    type: 'line',
    context: canvasRef.value?.getContext('2d'),
    width: canvasRef.value?.offsetWidth || 400,
    height: props.height || 180,
    categories: props.categories,
    series: props.series.map(s => ({
      name: s.name,
      data: s.data,
      color: s.color,
      addLine: true,
      fillOpacity: 0.15,
    })),
    xAxis: { disableGrid: false, gridColor: '#f3f4f6', fontSize: 11, fontColor: '#9ca3af' },
    yAxis: { gridColor: '#f3f4f6', fontSize: 11, fontColor: '#9ca3af' },
    legend: { show: true, position: 'bottom', fontSize: 11 },
    background: 'transparent',
    padding: [15, 10, 0, 10],
    dataLabel: false,
    enableScroll: false,
  })
  // #endif
}

onMounted(render)
watch(() => props.series, render, { deep: true })
</script>

<template>
  <canvas
    :id="canvasId"
    ref="canvasRef"
    class="chart-canvas"
    :style="{ width: '100%', height: (height || 180) + 'px' }"
  />
</template>

<style scoped>
.chart-canvas { display: block; }
</style>
```

- [ ] **Step 3: DonutChart.vue**

`frontend/src/components/charts/DonutChart.vue`:
```vue
<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
// #ifdef H5
import uCharts from '@qiun/uni-ucharts'
// #endif

const props = defineProps<{
  series: { name: string; data: number; color: string }[]
  size?: number
}>()

const canvasRef = ref()

function render() {
  // #ifdef H5
  const size = props.size || 200
  new uCharts({
    type: 'ring',
    context: canvasRef.value?.getContext('2d'),
    width: size, height: size,
    series: props.series.map(s => ({ name: s.name, data: s.data, color: s.color })),
    legend: { show: false },
    background: 'transparent',
    padding: [0, 0, 0, 0],
    extra: { ring: { ringWidth: 28, activeOpacity: 0.9, labelWidth: 0 } },
    dataLabel: false,
  })
  // #endif
}

onMounted(render)
watch(() => props.series, render, { deep: true })
</script>

<template>
  <canvas
    ref="canvasRef"
    class="donut-canvas"
    :style="{ width: (size || 200) + 'px', height: (size || 200) + 'px' }"
  />
</template>

<style scoped>
.donut-canvas { display: block; }
</style>
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/charts/
git commit -m "feat: add LineChart, DonutChart chart components via uCharts"
```

---

## Task 7: API 层封装

**Files:**
- Create: `frontend/src/api/request.ts`
- Create: `frontend/src/api/dashboard.ts`
- Create: `frontend/src/api/user.ts`

- [ ] **Step 1: request.ts（uni.request 封装）**

`frontend/src/api/request.ts`:
```typescript
import { useUserStore } from '@/stores/user'

const BASE_URL = import.meta.env.VITE_API_BASE || '/api/v1'

interface RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  data?: Record<string, unknown>
}

export async function request<T>(url: string, options: RequestOptions = {}): Promise<T> {
  return new Promise((resolve, reject) => {
    uni.request({
      url: BASE_URL + url,
      method: options.method || 'GET',
      data: options.data,
      header: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${uni.getStorageSync('token')}`,
      },
      success: (res: any) => {
        if (res.statusCode === 401) {
          uni.navigateTo({ url: '/pages/login/index' })
          reject(new Error('Unauthorized'))
          return
        }
        if (res.data?.code !== 0) {
          uni.showToast({ title: res.data?.msg || '请求失败', icon: 'none' })
          reject(new Error(res.data?.msg))
          return
        }
        resolve(res.data.data as T)
      },
      fail: (err) => {
        uni.showToast({ title: '网络异常', icon: 'none' })
        reject(err)
      },
    })
  })
}

export const get  = <T>(url: string, params?: Record<string, unknown>) => request<T>(url, { method: 'GET',  data: params })
export const post = <T>(url: string, data?:   Record<string, unknown>) => request<T>(url, { method: 'POST', data })
export const put  = <T>(url: string, data?:   Record<string, unknown>) => request<T>(url, { method: 'PUT',  data })
export const del  = <T>(url: string) => request<T>(url, { method: 'DELETE' })
```

- [ ] **Step 2: dashboard.ts**

`frontend/src/api/dashboard.ts`:
```typescript
import { get } from './request'

export interface DashboardStats {
  totalUsers: number
  activeOrgs: number
  monthlyRevenue: number
  pendingOrders: number
  onlineUsers: number
  todayRevenue: number
  todayNewUsers: number
}

export interface TrendPoint { date: string; users: number; revenue: number }
export interface OrgTypeItem { name: string; value: number; color: string }
export interface OrgRankItem { name: string; revenue: number }
export interface EventItem { id: string; color: string; text: string; ts: number }

export const getDashboardStats  = () => get<DashboardStats>('/dashboard/stats')
export const getTrendData        = (period: 'month'|'quarter'|'year') => get<TrendPoint[]>('/dashboard/trend', { period })
export const getOrgTypeDistrib   = () => get<OrgTypeItem[]>('/dashboard/org-types')
export const getOrgRank          = () => get<OrgRankItem[]>('/dashboard/org-rank')
export const getRealtimeEvents   = () => get<EventItem[]>('/dashboard/events')
```

- [ ] **Step 3: user.ts**

`frontend/src/api/user.ts`:
```typescript
import { get, post, put, del } from './request'

export interface UserListParams {
  keyword?: string
  orgId?: string
  role?: string
  status?: string
  certStatus?: string
  source?: string
  startDate?: string
  endDate?: string
  page: number
  pageSize: number
}

export interface UserItem {
  id: string
  username: string
  email: string
  avatar?: string
  orgName: string
  role: string
  status: 'active' | 'pending' | 'disabled'
  certStatus: 'certified' | 'pending' | 'none'
  createdAt: string
  lastLoginAt: string
}

export interface PageResult<T> { list: T[]; total: number }

export const getUserList   = (params: UserListParams) => get<PageResult<UserItem>>('/users', params as any)
export const getUserDetail = (id: string) => get<UserItem>(`/users/${id}`)
export const createUser    = (data: Partial<UserItem>) => post<UserItem>('/users', data as any)
export const updateUser    = (id: string, data: Partial<UserItem>) => put<UserItem>(`/users/${id}`, data as any)
export const deleteUser    = (id: string) => del<void>(`/users/${id}`)
export const batchEnable   = (ids: string[]) => post<void>('/users/batch-enable',  { ids })
export const batchDisable  = (ids: string[]) => post<void>('/users/batch-disable', { ids })
```

- [ ] **Step 4: user store**

`frontend/src/stores/user.ts`:
```typescript
import { defineStore } from 'pinia'
import { ref } from 'vue'

interface UserInfo { id: string; username: string; role: string; email: string }

export const useUserStore = defineStore('user', () => {
  const info = ref<UserInfo | null>(null)
  const token = ref<string>(uni.getStorageSync('token') || '')

  function setToken(t: string) {
    token.value = t
    uni.setStorageSync('token', t)
  }

  function setInfo(u: UserInfo) { info.value = u }

  function logout() {
    token.value = ''
    info.value = null
    uni.removeStorageSync('token')
    uni.reLaunch({ url: '/pages/login/index' })
  }

  return { info, token, setToken, setInfo, logout }
})
```

- [ ] **Step 5: Commit**

```bash
git add frontend/src/api/ frontend/src/stores/user.ts
git commit -m "feat: add API layer and user store"
```

---

## Task 8: 首页大屏页面（/pages/dashboard/index.vue）

**Files:**
- Create: `frontend/src/pages/dashboard/index.vue`

- [ ] **Step 1: 创建首页大屏页面**

`frontend/src/pages/dashboard/index.vue`:
```vue
<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import KpiCard from '@/components/common/KpiCard.vue'
import LineChart from '@/components/charts/LineChart.vue'
import DonutChart from '@/components/charts/DonutChart.vue'
import {
  getDashboardStats, getTrendData, getOrgTypeDistrib,
  getOrgRank, getRealtimeEvents,
  type DashboardStats, type TrendPoint, type OrgTypeItem,
  type OrgRankItem, type EventItem,
} from '@/api/dashboard'
import { formatTime } from '@/utils/format'

const breadcrumbs = [{ label: '首页' }, { label: '数据大屏' }]

const stats = ref<DashboardStats | null>(null)
const trendData = ref<TrendPoint[]>([])
const orgTypes = ref<OrgTypeItem[]>([])
const orgRank = ref<OrgRankItem[]>([])
const events = ref<EventItem[]>([])
const onlineNum = ref(0)
const activePeriod = ref<'month'|'quarter'|'year'>('month')

// mock 在线人数波动
let onlineTimer: ReturnType<typeof setInterval>
let eventsTimer: ReturnType<typeof setInterval>

async function loadAll() {
  const [s, t, ot, or_, ev] = await Promise.all([
    getDashboardStats(),
    getTrendData(activePeriod.value),
    getOrgTypeDistrib(),
    getOrgRank(),
    getRealtimeEvents(),
  ])
  stats.value = s
  trendData.value = t
  orgTypes.value = ot
  orgRank.value = or_
  events.value = ev
  onlineNum.value = s.onlineUsers
}

const lineCategories = computed(() =>
  trendData.value.map(d => d.date.slice(5))
)
const lineSeries = computed(() => [
  { name: '用户增长', data: trendData.value.map(d => d.users), color: '#3b82f6' },
  { name: '流水趋势', data: trendData.value.map(d => d.revenue), color: '#10b981' },
])

import { computed } from 'vue'

onMounted(() => {
  loadAll()
  onlineTimer = setInterval(() => {
    onlineNum.value += Math.floor(Math.random() * 7) - 3
  }, 3000)
  eventsTimer = setInterval(() => getRealtimeEvents().then(v => events.value = v), 30000)
})
onUnmounted(() => { clearInterval(onlineTimer); clearInterval(eventsTimer) })

async function changePeriod(p: 'month'|'quarter'|'year') {
  activePeriod.value = p
  trendData.value = await getTrendData(p)
}
</script>

<template>
  <AppLayout :breadcrumbs="breadcrumbs">
    <!-- 欢迎横幅 -->
    <view class="welcome-bar">
      <view>
        <view class="wb-title"><view class="live-dot" />数据实时概览</view>
        <text class="wb-sub">今日 {{ new Date().toLocaleDateString('zh-CN') }} · 实时更新中</text>
      </view>
      <view class="wb-stats">
        <view class="wb-stat">
          <text class="ws-num">{{ onlineNum.toLocaleString() }}</text>
          <text class="ws-lbl">当前在线</text>
        </view>
        <view class="ws-div" />
        <view class="wb-stat">
          <text class="ws-num">¥{{ ((stats?.todayRevenue || 0)/100).toFixed(0) }}</text>
          <text class="ws-lbl">今日流水</text>
        </view>
        <view class="ws-div" />
        <view class="wb-stat">
          <text class="ws-num">{{ stats?.todayNewUsers || 0 }}</text>
          <text class="ws-lbl">今日新增</text>
        </view>
      </view>
    </view>

    <!-- KPI 卡片 -->
    <view class="kpi-row" v-if="stats">
      <KpiCard icon="👥" label="注册用户总数" :value="stats.totalUsers"
        :trend="{ dir: 'up', text: '12.3% vs上月' }" icon-bg="#eff6ff"
        :sparkline="[40,55,48,70,62,85,78,100]" />
      <KpiCard icon="🏢" label="活跃联盟数" :value="stats.activeOrgs"
        :trend="{ dir: 'up', text: '5.1% vs上月' }" icon-bg="#f0fdf4"
        :sparkline="[50,45,65,60,75,70,90,85]" />
      <KpiCard icon="💰" label="本月订单流水" :value="stats.monthlyRevenue" unit="¥"
        :trend="{ dir: 'up', text: '8.7% vs上月' }" icon-bg="#fff7ed"
        :sparkline="[30,50,45,65,75,70,90,100]" />
      <KpiCard icon="📦" label="待处理订单" :value="stats.pendingOrders"
        :trend="{ dir: 'down', text: '3单待跟进' }" icon-bg="#fef2f2"
        :sparkline="[80,90,70,60,50,55,40,35]" />
    </view>

    <!-- 图表行 -->
    <view class="chart-row">
      <view class="card chart-card">
        <view class="card-header">
          <text class="card-title">近期用户增长 & 流水趋势</text>
          <view class="card-tabs">
            <text class="tab" :class="{ active: activePeriod==='month' }" @click="changePeriod('month')">月</text>
            <text class="tab" :class="{ active: activePeriod==='quarter' }" @click="changePeriod('quarter')">季</text>
            <text class="tab" :class="{ active: activePeriod==='year' }" @click="changePeriod('year')">年</text>
          </view>
        </view>
        <LineChart :categories="lineCategories" :series="lineSeries" :height="160" />
      </view>

      <view class="card chart-card">
        <view class="card-header"><text class="card-title">联盟类型分布</text></view>
        <view class="donut-wrap">
          <DonutChart :series="orgTypes" :size="160" />
          <view class="donut-legend">
            <view v-for="item in orgTypes" :key="item.name" class="dl-item">
              <view class="dl-dot" :style="{ background: item.color }" />
              <text class="dl-name">{{ item.name }}</text>
              <text class="dl-val">{{ item.value }}</text>
            </view>
          </view>
        </view>
      </view>
    </view>

    <!-- 底部行 -->
    <view class="bottom-row">
      <!-- 排行榜 -->
      <view class="card list-card">
        <view class="card-header">
          <text class="card-title">联盟流水排行 TOP 5</text>
          <text class="link">查看全部 ›</text>
        </view>
        <view v-for="(item, idx) in orgRank" :key="item.name" class="rank-item">
          <view class="rank-num" :class="{ top: idx < 3 }">{{ idx + 1 }}</view>
          <text class="rank-name">{{ item.name }}</text>
          <view class="rank-bar-wrap">
            <view class="rank-bar" :style="{ width: (item.revenue / (orgRank[0]?.revenue||1) * 100) + '%' }" />
          </view>
          <text class="rank-val">¥{{ (item.revenue/100).toFixed(1) }}K</text>
        </view>
      </view>

      <!-- 实时动态 -->
      <view class="card list-card">
        <view class="card-header">
          <view class="rt-title"><view class="live-dot" /><text class="card-title">实时动态</text></view>
          <text class="link">全部动态 ›</text>
        </view>
        <view v-for="ev in events" :key="ev.id" class="event-item">
          <view class="ev-dot" :style="{ background: ev.color }" />
          <text class="ev-text" v-html="ev.text" />
          <text class="ev-time">{{ formatTime(ev.ts) }}</text>
        </view>
      </view>
    </view>
  </AppLayout>
</template>

<style lang="scss" scoped>
.welcome-bar {
  background: linear-gradient(135deg, var(--color-primary-dark) 0%, var(--color-primary) 100%);
  border-radius: 14px; padding: 22px 28px; margin-bottom: 20px;
  display: flex; align-items: center; justify-content: space-between;
}
.wb-title { font-size: 20px; font-weight: 700; color: #fff; display: flex; align-items: center; gap: 8px; margin-bottom: 4px; }
.wb-sub { font-size: 13px; color: rgba(255,255,255,0.7); }
.live-dot { width: 8px; height: 8px; border-radius: 50%; background: #34d399; animation: pulse 1.5s ease-in-out infinite; display: inline-block; }
@keyframes pulse { 0%,100%{opacity:1} 50%{opacity:.5} }
.wb-stats { display: flex; gap: 24px; }
.wb-stat { text-align: center; }
.ws-num { font-size: 22px; font-weight: 700; color: #fff; display: block; }
.ws-lbl { font-size: 11px; color: rgba(255,255,255,0.65); }
.ws-div { width: 1px; background: rgba(255,255,255,0.2); }

.kpi-row { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: 20px; }

.chart-row { display: grid; grid-template-columns: 2fr 1fr; gap: 16px; margin-bottom: 20px; }
.chart-card { padding: 20px; }
.card-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; }
.card-title { font-size: 14px; font-weight: 600; color: var(--color-text-primary); }
.card-tabs { display: flex; gap: 4px; }
.tab { font-size: 11px; padding: 3px 10px; border-radius: 6px; cursor: pointer; color: var(--color-text-secondary); }
.tab.active { background: var(--color-primary-light); color: var(--color-primary); font-weight: 600; }

.donut-wrap { display: flex; flex-direction: column; align-items: center; }
.donut-legend { width: 100%; margin-top: 12px; display: flex; flex-direction: column; gap: 6px; }
.dl-item { display: flex; align-items: center; gap: 8px; font-size: 11px; }
.dl-dot { width: 8px; height: 8px; border-radius: 2px; flex-shrink: 0; }
.dl-name { color: var(--color-text-secondary); flex: 1; }
.dl-val { color: var(--color-text-primary); font-weight: 600; }

.bottom-row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
.list-card { padding: 20px; }
.rank-item { display: flex; align-items: center; gap: 12px; padding: 9px 0; border-bottom: 1px solid var(--color-border-light); }
.rank-item:last-child { border: none; }
.rank-num { width: 20px; height: 20px; border-radius: 5px; background: var(--color-border-light); font-size: 10px; font-weight: 700; display: flex; align-items: center; justify-content: center; color: var(--color-text-secondary); }
.rank-num.top { background: var(--color-primary); color: #fff; }
.rank-name { flex: 1; font-size: 13px; color: var(--color-text-primary); }
.rank-bar-wrap { width: 80px; height: 6px; background: var(--color-border-light); border-radius: 3px; }
.rank-bar { height: 6px; border-radius: 3px; background: linear-gradient(to right, var(--color-primary-dark), var(--color-primary)); }
.rank-val { width: 52px; text-align: right; font-size: 12px; font-weight: 600; color: var(--color-text-primary); }

.rt-title { display: flex; align-items: center; gap: 6px; }
.link { font-size: 12px; color: var(--color-primary); cursor: pointer; }
.event-item { display: flex; align-items: flex-start; gap: 10px; padding: 8px 0; border-bottom: 1px solid var(--color-border-light); }
.event-item:last-child { border: none; }
.ev-dot { width: 8px; height: 8px; border-radius: 50%; margin-top: 4px; flex-shrink: 0; }
.ev-text { flex: 1; font-size: 12px; color: var(--color-text-primary); line-height: 1.5; }
.ev-time { font-size: 11px; color: var(--color-text-muted); flex-shrink: 0; }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/pages/dashboard/
git commit -m "feat: add dashboard homepage with realtime stats and charts"
```

---

## Task 9: 用户管理列表页（/pages/users/index.vue）

**Files:**
- Create: `frontend/src/pages/users/index.vue`

- [ ] **Step 1: 创建用户管理列表页**

`frontend/src/pages/users/index.vue`:
```vue
<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import KpiCard from '@/components/common/KpiCard.vue'
import FilterPanel from '@/components/common/FilterPanel.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import Pagination from '@/components/common/Pagination.vue'
import { getUserList, batchEnable, batchDisable, type UserItem } from '@/api/user'
import type { FilterField, QuickTag } from '@/components/common/FilterPanel.vue'

const breadcrumbs = [{ label: '首页' }, { label: '核心业务' }, { label: '用户管理' }]

// 统计数字
const stats = reactive({ total: 0, active: 0, pending: 0, disabled: 0, todayNew: 0 })

// 筛选字段定义
const filterFields: FilterField[] = [
  { key: 'keyword',    label: '关键词搜索',  type: 'input',  placeholder: '姓名 / 手机号 / 邮箱' },
  { key: 'orgName',    label: '所属联盟',    type: 'input',  placeholder: '输入联盟名称' },
  { key: 'role',       label: '用户角色',    type: 'select', options: [
      { label: '全部角色', value: '' },
      { label: '联盟主', value: 'owner' },
      { label: '管理员', value: 'admin' },
      { label: '财务',   value: 'finance' },
      { label: '普通会员', value: 'member' },
  ]},
  { key: 'status',     label: '账号状态',    type: 'select', options: [
      { label: '全部状态', value: '' },
      { label: '正常',   value: 'active' },
      { label: '待审核', value: 'pending' },
      { label: '已禁用', value: 'disabled' },
  ]},
  { key: 'startDate',  label: '注册时间（起）', type: 'input', placeholder: 'YYYY-MM-DD' },
  { key: 'endDate',    label: '注册时间（止）', type: 'input', placeholder: 'YYYY-MM-DD' },
  { key: 'certStatus', label: '实名认证',    type: 'select', options: [
      { label: '全部', value: '' },
      { label: '已认证', value: 'certified' },
      { label: '审核中', value: 'pending' },
      { label: '未认证', value: 'none' },
  ]},
  { key: 'source',     label: '注册来源',    type: 'select', options: [
      { label: '全部渠道', value: '' },
      { label: 'PC 网页', value: 'web' },
      { label: '微信小程序', value: 'mp_wx' },
      { label: '手机 App', value: 'app' },
  ]},
]

const quickTags: QuickTag[] = [
  { key: 'today',    label: '今日新增',       color: '#10b981', params: { startDate: new Date().toISOString().slice(0,10) } },
  { key: 'pending',  label: '待审核',          color: '#f59e0b', params: { status: 'pending' } },
  { key: 'inactive', label: '长期未登录>30天', color: '#ef4444', params: { inactive: '30' } },
  { key: 'active_month', label: '本月活跃',   color: '#3b82f6', params: { activeMonth: '1' } },
  { key: 'disabled', label: '已禁用',          color: '#6b7280', params: { status: 'disabled' } },
]

// 表格状态
const list = ref<UserItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const selectedIds = ref<string[]>([])
const filterParams = ref<Record<string, unknown>>({})

const columns = [
  { key: 'user',       label: '用户信息',   width: 220 },
  { key: 'orgName',    label: '所属联盟',   width: 140 },
  { key: 'role',       label: '角色',       width: 100 },
  { key: 'status',     label: '状态',       width: 90 },
  { key: 'certStatus', label: '实名认证',   width: 90 },
  { key: 'createdAt',  label: '注册时间',   width: 110, sortable: true },
  { key: 'lastLoginAt',label: '最后登录',   width: 100 },
  { key: 'action',     label: '操作',       width: 140 },
]

const roleLabel: Record<string, string> = { owner: '联盟主', admin: '管理员', finance: '财务', member: '普通会员' }
const avatarColors = ['#1e40af','#7c3aed','#059669','#dc2626','#d97706']

async function loadList() {
  loading.value = true
  try {
    const res = await getUserList({
      ...filterParams.value as any,
      page: page.value,
      pageSize: pageSize.value,
    })
    list.value = res.list
    total.value = res.total
    // 更新统计数字（实际项目由后端专门接口提供）
    stats.total = res.total
  } finally {
    loading.value = false
  }
}

function onSearch(params: Record<string, unknown>) {
  filterParams.value = params
  page.value = 1
  loadList()
}

function onPageChange(p: number) { page.value = p; loadList() }
function onPageSizeChange(s: number) { pageSize.value = s; page.value = 1; loadList() }

function toggleSelect(id: string) {
  const idx = selectedIds.value.indexOf(id)
  idx === -1 ? selectedIds.value.push(id) : selectedIds.value.splice(idx, 1)
}

async function handleBatchEnable() {
  await batchEnable(selectedIds.value)
  uni.showToast({ title: '批量启用成功', icon: 'success' })
  selectedIds.value = []
  loadList()
}

async function handleBatchDisable() {
  await batchDisable(selectedIds.value)
  uni.showToast({ title: '批量禁用成功', icon: 'success' })
  selectedIds.value = []
  loadList()
}

function goDetail(id: string) {
  uni.navigateTo({ url: `/pages/users/detail?id=${id}` })
}

onMounted(loadList)
</script>

<template>
  <AppLayout :breadcrumbs="breadcrumbs">

    <!-- 顶部统计 -->
    <view class="stats-row">
      <KpiCard icon="👥" label="总用户数"  :value="stats.total"    :trend="{ dir:'up',   text:'12.3%' }" icon-bg="#eff6ff" />
      <KpiCard icon="✅" label="已激活"    :value="stats.active"   :trend="{ dir:'up',   text:'8.1%' }"  icon-bg="#f0fdf4" />
      <KpiCard icon="⏳" label="待审核"    :value="stats.pending"  :trend="{ dir:'up',   text:'今日+34' }" icon-bg="#fffbeb" />
      <KpiCard icon="🚫" label="已禁用"    :value="stats.disabled" :trend="{ dir:'down',  text:'无变化' }"  icon-bg="#fef2f2" />
      <KpiCard icon="🆕" label="今日新增"  :value="stats.todayNew" :trend="{ dir:'up',   text:'较昨日+12' }" icon-bg="#faf5ff" />
    </view>

    <!-- 筛选面板 -->
    <FilterPanel
      :fields="filterFields"
      :quick-tags="quickTags"
      @search="onSearch"
      @reset="() => { filterParams = {}; loadList() }"
      @export="() => uni.showToast({ title: '导出中...', icon: 'none' })"
    />

    <!-- 表格 -->
    <view class="card table-card">
      <!-- 工具栏 -->
      <view class="table-toolbar">
        <view class="toolbar-left">
          <text class="selected-info">
            共 <text class="em">{{ total }}</text> 条，已选 <text class="em">{{ selectedIds.length }}</text> 条
          </text>
          <view v-if="selectedIds.length" class="btn btn-sm btn-outline" @click="handleBatchEnable">批量启用</view>
          <view v-if="selectedIds.length" class="btn btn-sm btn-danger-outline" @click="handleBatchDisable">批量禁用</view>
        </view>
        <view class="toolbar-right">
          <view class="icon-btn" @click="loadList">🔄</view>
          <view class="btn btn-primary btn-sm" @click="uni.navigateTo({ url: '/pages/users/detail?mode=create' })">＋ 新增用户</view>
        </view>
      </view>

      <!-- 表头 -->
      <view class="table-head">
        <view class="th-check">
          <view
            class="checkbox"
            :class="{ checked: selectedIds.length === list.length && list.length > 0 }"
            @click="selectedIds.length === list.length ? selectedIds = [] : selectedIds = list.map(u => u.id)"
          >{{ selectedIds.length === list.length && list.length > 0 ? '✓' : '' }}</view>
        </view>
        <view v-for="col in columns" :key="col.key" class="th" :style="{ width: col.width + 'px' }">
          {{ col.label }}
          <text v-if="col.sortable" class="sort-icon">↕</text>
        </view>
      </view>

      <!-- 表格行 -->
      <view v-if="loading" class="table-loading">加载中...</view>
      <view v-else>
        <view
          v-for="row in list"
          :key="row.id"
          class="table-row"
          :class="{ selected: selectedIds.includes(row.id) }"
        >
          <view class="td-check">
            <view class="checkbox" :class="{ checked: selectedIds.includes(row.id) }" @click="toggleSelect(row.id)">
              {{ selectedIds.includes(row.id) ? '✓' : '' }}
            </view>
          </view>
          <!-- 用户信息 -->
          <view class="td" style="width:220px">
            <view class="user-cell">
              <view class="user-av" :style="{ background: avatarColors[row.username.charCodeAt(0) % 5] }">
                {{ row.username.charAt(0) }}
              </view>
              <view>
                <text class="user-name">{{ row.username }}</text>
                <text class="user-email">{{ row.email }}</text>
              </view>
            </view>
          </view>
          <view class="td" style="width:140px"><text class="org-tag">{{ row.orgName }}</text></view>
          <view class="td" style="width:100px">
            <StatusBadge status="info" :label="roleLabel[row.role] || row.role" />
          </view>
          <view class="td" style="width:90px">
            <StatusBadge
              :status="row.status === 'active' ? 'success' : row.status === 'pending' ? 'warning' : 'danger'"
              :label="row.status === 'active' ? '正常' : row.status === 'pending' ? '待审核' : '已禁用'"
            />
          </view>
          <view class="td" style="width:90px">
            <StatusBadge
              :status="row.certStatus === 'certified' ? 'success' : row.certStatus === 'pending' ? 'warning' : 'danger'"
              :label="row.certStatus === 'certified' ? '已认证' : row.certStatus === 'pending' ? '审核中' : '未认证'"
            />
          </view>
          <view class="td" style="width:110px"><text class="text-muted">{{ row.createdAt.slice(0,10) }}</text></view>
          <view class="td" style="width:100px"><text class="text-muted">{{ row.lastLoginAt }}</text></view>
          <view class="td" style="width:140px">
            <view class="action-btns">
              <view class="act-btn act-view" @click="goDetail(row.id)">详情</view>
              <view class="act-btn act-edit" @click="goDetail(row.id + '?mode=edit')">编辑</view>
              <view class="act-btn act-more">···</view>
            </view>
          </view>
        </view>
      </view>

      <!-- 分页 -->
      <Pagination :total="total" :page="page" :page-size="pageSize"
        @page-change="onPageChange" @page-size-change="onPageSizeChange" />
    </view>
  </AppLayout>
</template>

<style lang="scss" scoped>
.stats-row { display: grid; grid-template-columns: repeat(5,1fr); gap: 12px; margin-bottom: 16px; }

.table-card { overflow: hidden; }
.table-toolbar { display: flex; align-items: center; justify-content: space-between; padding: 14px 16px; }
.toolbar-left { display: flex; align-items: center; gap: 8px; }
.toolbar-right { display: flex; align-items: center; gap: 8px; }
.selected-info { font-size: 13px; color: var(--color-text-secondary); }
.em { color: var(--color-primary); font-weight: 600; }
.btn { height: 34px; border-radius: 7px; border: none; cursor: pointer; font-size: 13px; font-weight: 500; display: flex; align-items: center; gap: 6px; padding: 0 16px; }
.btn-sm { height: 30px; font-size: 12px; padding: 0 12px; }
.btn-primary { background: var(--color-primary); color: #fff; }
.btn-outline { background: var(--color-card-bg); color: var(--color-text-secondary); border: 1px solid var(--color-border); }
.btn-danger-outline { background: var(--color-card-bg); color: #ef4444; border: 1px solid #fecaca; }
.icon-btn { width: 32px; height: 32px; border-radius: 7px; border: 1px solid var(--color-border); background: var(--color-card-bg); display: flex; align-items: center; justify-content: center; font-size: 14px; cursor: pointer; }

.table-head { display: flex; align-items: center; background: var(--color-border-light); padding: 10px 16px; border-top: 1px solid var(--color-border-light); border-bottom: 1px solid var(--color-border-light); }
.th-check { width: 40px; }
.th { font-size: 12px; font-weight: 600; color: var(--color-text-secondary); }
.sort-icon { margin-left: 4px; opacity: 0.5; }
.table-loading { padding: 40px; text-align: center; color: var(--color-text-muted); font-size: 13px; }

.table-row { display: flex; align-items: center; padding: 10px 16px; border-bottom: 1px solid var(--color-border-light); transition: background 0.1s; }
.table-row:hover { background: var(--color-border-light); }
.table-row.selected { background: var(--color-primary-light); }
.table-row:last-child { border: none; }
.td-check { width: 40px; }
.td { font-size: 13px; color: var(--color-text-primary); }
.text-muted { font-size: 12px; color: var(--color-text-muted); }

.checkbox { width: 16px; height: 16px; border: 1.5px solid var(--color-border); border-radius: 4px; cursor: pointer; display: flex; align-items: center; justify-content: center; font-size: 10px; }
.checkbox.checked { background: var(--color-primary); border-color: var(--color-primary); color: #fff; }

.user-cell { display: flex; align-items: center; gap: 10px; }
.user-av { width: 32px; height: 32px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 12px; font-weight: 700; color: #fff; flex-shrink: 0; }
.user-name { font-size: 13px; font-weight: 500; color: var(--color-text-primary); display: block; }
.user-email { font-size: 11px; color: var(--color-text-muted); display: block; }
.org-tag { background: var(--color-border-light); color: var(--color-text-primary); padding: 2px 8px; border-radius: 4px; font-size: 11px; }

.action-btns { display: flex; gap: 6px; }
.act-btn { font-size: 12px; padding: 4px 10px; border-radius: 5px; cursor: pointer; font-weight: 500; }
.act-view { background: var(--color-primary-light); color: var(--color-primary); }
.act-edit { background: #f0fdf4; color: #16a34a; }
.act-more { background: var(--color-border-light); color: var(--color-text-secondary); }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/pages/users/
git commit -m "feat: add user management list page with filter, table, pagination"
```

---

## Task 10: 其余业务页面（联盟/订单/财务/报表/消息/权限/设置）

每个页面复用相同模式：AppLayout + KpiCard 统计栏 + FilterPanel + 表格 + Pagination。

**Files:**
- Create: `frontend/src/pages/orgs/index.vue`
- Create: `frontend/src/pages/orders/index.vue`
- Create: `frontend/src/pages/finance/index.vue`
- Create: `frontend/src/pages/reports/index.vue`
- Create: `frontend/src/pages/messages/index.vue`
- Create: `frontend/src/pages/permissions/index.vue`
- Create: `frontend/src/pages/settings/index.vue`

> 每个页面的实现与 Task 9 完全相同结构，只需替换：
> - `breadcrumbs` 面包屑
> - `filterFields` 筛选字段（参见设计文档第六节）
> - `quickTags` 快捷标签
> - `columns` 列定义
> - API 调用方法（从对应 api/*.ts 文件导入）

- [ ] **联盟管理 orgs/index.vue 筛选字段**

```typescript
const filterFields: FilterField[] = [
  { key: 'keyword',    label: '关键词',   type: 'input',  placeholder: '联盟名称 / 负责人' },
  { key: 'type',       label: '联盟类型', type: 'select', options: [
    { label: '全部', value: '' }, { label: '电商联盟', value: 'ec' },
    { label: '服务联盟', value: 'service' }, { label: '内容联盟', value: 'content' },
  ]},
  { key: 'status',     label: '审核状态', type: 'select', options: [
    { label: '全部', value: '' }, { label: '正常', value: 'active' },
    { label: '待审核', value: 'pending' }, { label: '已冻结', value: 'frozen' },
  ]},
  { key: 'region',     label: '所在地区', type: 'input', placeholder: '省/市' },
  { key: 'startDate',  label: '成立时间（起）', type: 'input', placeholder: 'YYYY-MM-DD' },
  { key: 'endDate',    label: '成立时间（止）', type: 'input', placeholder: 'YYYY-MM-DD' },
]
```

- [ ] **订单中心 orders/index.vue 筛选字段**

```typescript
const filterFields: FilterField[] = [
  { key: 'keyword',    label: '订单号 / 用户', type: 'input' },
  { key: 'type',       label: '订单类型',  type: 'select', options: [
    { label: '全部', value: '' }, { label: '普通订单', value: 'normal' }, { label: '退款单', value: 'refund' },
  ]},
  { key: 'status',     label: '支付状态',  type: 'select', options: [
    { label: '全部', value: '' }, { label: '待支付', value: 'pending' },
    { label: '已支付', value: 'paid' }, { label: '已退款', value: 'refunded' },
  ]},
  { key: 'payMethod',  label: '支付方式',  type: 'select', options: [
    { label: '全部', value: '' }, { label: '微信', value: 'wx' }, { label: '支付宝', value: 'ali' },
  ]},
  { key: 'minAmount',  label: '最小金额',  type: 'input', placeholder: '元' },
  { key: 'maxAmount',  label: '最大金额',  type: 'input', placeholder: '元' },
  { key: 'startDate',  label: '下单时间（起）', type: 'input', placeholder: 'YYYY-MM-DD' },
  { key: 'endDate',    label: '下单时间（止）', type: 'input', placeholder: 'YYYY-MM-DD' },
]
```

- [ ] **财务结算 finance/index.vue 筛选字段**

```typescript
const filterFields: FilterField[] = [
  { key: 'keyword',      label: '关键词',   type: 'input', placeholder: '联盟名称 / 账户' },
  { key: 'status',       label: '结算状态', type: 'select', options: [
    { label: '全部', value: '' }, { label: '待结算', value: 'pending' },
    { label: '已结算', value: 'done' }, { label: '结算中', value: 'processing' },
  ]},
  { key: 'accountType',  label: '账户类型', type: 'select', options: [
    { label: '全部', value: '' }, { label: '银行卡', value: 'bank' }, { label: '支付宝', value: 'ali' },
  ]},
  { key: 'period',       label: '结算周期', type: 'select', options: [
    { label: '全部', value: '' }, { label: '日结', value: 'daily' }, { label: '月结', value: 'monthly' },
  ]},
  { key: 'minAmount',    label: '最小金额（元）', type: 'input' },
  { key: 'maxAmount',    label: '最大金额（元）', type: 'input' },
  { key: 'startDate',    label: '结算时间（起）', type: 'input', placeholder: 'YYYY-MM-DD' },
  { key: 'endDate',      label: '结算时间（止）', type: 'input', placeholder: 'YYYY-MM-DD' },
]
```

- [ ] **Commit**

```bash
git add frontend/src/pages/
git commit -m "feat: add all business list pages (orgs, orders, finance, reports, messages, permissions, settings)"
```

---

## Task 11: 登录页

**Files:**
- Create: `frontend/src/pages/login/index.vue`

- [ ] **Step 1: 创建登录页**

`frontend/src/pages/login/index.vue`:
```vue
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
    userStore.setInfo(res.user)
    uni.reLaunch({ url: '/pages/dashboard/index' })
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <view class="login-page">
    <view class="login-card">
      <view class="login-logo">
        <view class="logo-icon">联</view>
        <text class="logo-title">联盟管理中心</text>
        <text class="logo-sub">Union Manage Center</text>
      </view>
      <view class="form-group">
        <text class="form-label">账号</text>
        <input v-model="form.username" class="form-input" placeholder="请输入账号" />
      </view>
      <view class="form-group">
        <text class="form-label">密码</text>
        <input v-model="form.password" class="form-input" password placeholder="请输入密码" @confirm="handleLogin" />
      </view>
      <view class="login-btn" :class="{ loading }" @click="handleLogin">
        {{ loading ? '登录中...' : '登 录' }}
      </view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.login-page {
  min-height: 100vh; background: var(--color-bg);
  display: flex; align-items: center; justify-content: center;
}
.login-card {
  width: 400px; background: var(--color-card-bg);
  border-radius: 16px; padding: 40px;
  box-shadow: 0 8px 40px rgba(0,0,0,0.1);
}
.login-logo { text-align: center; margin-bottom: 32px; }
.logo-icon {
  width: 56px; height: 56px; border-radius: 14px;
  background: linear-gradient(135deg, var(--color-primary-dark), var(--color-primary));
  display: flex; align-items: center; justify-content: center;
  color: #fff; font-size: 24px; font-weight: bold; margin: 0 auto 12px;
}
.logo-title { font-size: 20px; font-weight: 700; color: var(--color-text-primary); display: block; }
.logo-sub { font-size: 12px; color: var(--color-text-muted); display: block; margin-top: 4px; }
.form-group { margin-bottom: 18px; }
.form-label { font-size: 13px; color: var(--color-text-secondary); display: block; margin-bottom: 6px; }
.form-input {
  width: 100%; height: 42px; border: 1px solid var(--color-border);
  border-radius: 9px; padding: 0 14px; font-size: 14px;
  color: var(--color-text-primary); background: var(--color-card-bg);
  &:focus { border-color: var(--color-primary); outline: none; }
}
.login-btn {
  width: 100%; height: 44px; background: var(--color-primary);
  color: #fff; border-radius: 9px; font-size: 15px; font-weight: 600;
  display: flex; align-items: center; justify-content: center;
  cursor: pointer; margin-top: 8px;
  &.loading { opacity: 0.7; cursor: not-allowed; }
}
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/pages/login/
git commit -m "feat: add login page"
```

---

## Task 12: 最终验证

- [ ] **Step 1: 完整构建 H5**

```bash
cd /Users/fatboss/gowork/src/unionManageCenter/frontend
npm run build:h5
```

预期：`dist/build/h5/` 目录生成，无编译错误。

- [ ] **Step 2: 开发模式启动验证**

```bash
npm run dev:h5
```

预期：
- 访问 `http://localhost:5173` 跳转登录页
- 登录后进入首页大屏，时钟实时更新
- 右下角主题切换按钮可切换 A/B/C 三套主题
- 侧边栏点击用户管理进入列表页，筛选面板可展开/收起
- 列表支持翻页和多选

- [ ] **Step 3: 最终 Commit**

```bash
git add .
git commit -m "feat: complete union management center frontend v1.0"
```
