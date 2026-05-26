# 架构设计

## 整体架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        客户端 (H5 / 小程序)                      │
│                    UniApp Vue3 + Pinia + SVG                    │
└──────────────────────────────┬──────────────────────────────────┘
                               │ HTTP / REST
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                      统一网关 (app/gateway)                      │
│                        Go + Gin v1.12                           │
│                                                                 │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌───────────────┐  │
│  │  CORS    │  │  JWT     │  │  Logger  │  │  Recovery     │  │
│  │ 中间件   │  │ 鉴权     │  │  日志    │  │  错误恢复     │  │
│  └──────────┘  └──────────┘  └──────────┘  └───────────────┘  │
│                                                                 │
│  Handler 层                                                     │
│  auth │ user │ role │ permission │ org │ order │ finance       │
│  message │ dashboard │ report                                  │
└──────────────────────────────┬──────────────────────────────────┘
                               │ GORM v2
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                         MySQL 8.x                               │
│                      数据库：union_manage                        │
│                                                                 │
│  users  roles  permissions  user_roles  role_permissions        │
│  orgs  org_members  orders  finance_accounts                    │
│  finance_settlements  messages  dict_types  dict_items          │
│  operation_logs                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## 分层设计

### 后端分层

```
cmd/server/main.go          # 入口：初始化 DB、启动 HTTP
    └── router/router.go    # 路由注册，公开路由 + JWT 保护路由
        └── handler/*.go    # 业务逻辑，直接操作 GORM
            └── model/*.go  # GORM 数据模型
pkg/
    ├── auth/               # JWT 生成/解析，独立于业务
    ├── database/           # DB 单例，连接池配置
    ├── middleware/         # 可复用中间件（跨服务）
    └── response/           # 统一 JSON 响应结构
```

### 前端分层

```
pages/          # 页面层：组合 API 调用 + 本地状态 + 模板渲染
api/            # 接口层：封装 uni.request，数据归一化（snake_case → camelCase）
stores/         # 全局状态：user（登录信息）、theme（当前主题）
components/     # UI 组件：KpiCard、FilterPanel、Pagination、StatusBadge
utils/          # 工具函数：auth（token 存取）、format（时间格式化）
styles/themes/  # CSS 变量主题，通过 body class 切换
```

## RBAC 权限模型

```
User ──m:n── Role ──m:n── Permission
                              │
                    type: menu | button | api
                              │
                    ┌─────────┴────────┐
                  menu              api/button
               (前端菜单控制)      (后端接口控制)
```

权限粒度：

| 类型 | 示例 | 作用 |
|------|------|------|
| `menu` | `/dashboard` | 控制侧边栏菜单显示 |
| `button` | `user:create` | 控制页面按钮是否可见 |
| `api` | `GET:/users` | 后端接口级鉴权（可扩展） |

## 数据流（以大盘为例）

```
dashboard/index.vue
    │
    ├── getDashboardStats()   →  GET /dashboard/stats
    │       └── handler/dashboard.go
    │               └── SELECT COUNT(*) FROM users ...
    │
    ├── getTrendData(period) →  GET /dashboard/trend?period=month
    │       └── GROUP BY DATE(created_at) 聚合
    │
    ├── getOrgTypeDistrib()  →  GET /dashboard/org-type
    │
    └── getRealtimeEvents()  →  GET /dashboard/events
            └── 查最近 10 条 operation_logs
```

## 图表架构

统计图全部使用**浏览器原生 SVG**，无需第三方图表库，实现零依赖、高性能的数据可视化：

```
SVG viewBox 坐标系 (固定逻辑尺寸)
    ├── Y 轴：niceMax 算法取整 → 5 档刻度 + 水平网格线 + 标签
    ├── X 轴：自适应间距 → 日期标签 + 刻度线
    ├── 曲线：三次贝塞尔平滑（C 命令）+ 渐变面积
    └── Hover：
            mousemove 换算 SVG 坐标 → 找最近数据点 index
            → 绘制竖向虚线 + 系列圆点
            → 绝对定位 HTML div 看板（随鼠标自动左右切边）
```

## 扩展方向

- **微服务拆分**：`app/` 下已预留各子服务骨架（user/org/order…），可在流量增长后独立部署，网关通过服务注册（consul/etcd）路由
- **缓存层**：热点数据（在线用户数、大盘统计）接入 Redis，减少 DB 查询
- **消息队列**：订单状态变更、财务结算触发改为 MQ 异步处理（Kafka/NATS）
- **链路追踪**：接入 OpenTelemetry，配合 Jaeger 实现分布式追踪
