# 联盟管理中心 · UnionManageCenter

> 面向联盟业务的全栈管理平台，后端基于 Go + Gin + GORM，前端基于 UniApp (Vue3)，支持 H5 / 小程序多端输出。

![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)
![Vue](https://img.shields.io/badge/Vue-3.x-4FC08D?logo=vue.js)
![UniApp](https://img.shields.io/badge/UniApp-3.x-2B9939?logo=data:image/png;base64,)
![MySQL](https://img.shields.io/badge/MySQL-8.x-4479A1?logo=mysql)
![License](https://img.shields.io/badge/license-MIT-green)

---

## 目录

- [功能概览](#功能概览)
- [技术栈](#技术栈)
- [项目结构](#项目结构)
- [快速启动](#快速启动)
- [前端说明](#前端说明)
- [接口文档](#接口文档)
- [测试](#测试)
- [部署](#部署)

---

## 功能概览

| 模块 | 功能 |
|------|------|
| **用户管理** | 注册/查询/启停/角色分配，支持多状态筛选 |
| **角色权限** | RBAC 三层权限树（菜单/按钮/API），角色灵活授权 |
| **联盟管理** | 联盟 CRUD、成员管理、类型分类，流水排行 |
| **订单中心** | 订单列表、状态流转、退款操作，多维筛选 |
| **财务结算** | 结算账单、金额统计、手动结算触发 |
| **消息通知** | 系统通知、已读/未读标记、全部已读 |
| **数据报表** | 月度 KPI、日粒度明细、环比趋势 |
| **数据大屏** | 实时在线、今日流水、趋势折线、类型圆环、流水排行 |

### 图表特性

- SVG 原生渲染，无需第三方图表库
- **坐标轴**：Y 轴自动取整刻度 + 网格线，X 轴日期标签自适应间距
- **Hover 数据看板**：鼠标划过展示十字准线、数据圆点、浮动数据卡片
- 三套主题（蓝/绿/深色）可全局切换

---

## 技术栈

### 后端

| 技术 | 版本 | 用途 |
|------|------|------|
| Go | 1.25 | 开发语言 |
| Gin | v1.12 | HTTP 框架 |
| GORM v2 | v1.31 | ORM |
| MySQL | 8.x | 关系型数据库 |
| golang-jwt | v5 | JWT 认证 |
| bcrypt | stdlib | 密码哈希 |
| go.work | — | 多模块工作区 |

### 前端

| 技术 | 版本 | 用途 |
|------|------|------|
| UniApp | 3.x | 多端框架 |
| Vue | 3.4 | UI 框架 |
| Vite | 5.2 | 构建工具 |
| Pinia | 3.x | 状态管理 |
| TypeScript | 4.9 | 类型安全 |
| SCSS | — | 样式 + 主题变量 |
| SVG | — | 图表渲染 |

---

## 项目结构

```
unionManageCenter/
├── app/
│   ├── gateway/                   # 统一网关服务（主后端）
│   │   ├── cmd/server/main.go     # 入口
│   │   ├── configs/config.yaml    # 服务配置
│   │   └── internal/
│   │       ├── handler/           # 业务 Handler
│   │       │   ├── auth.go        # 登录/注册
│   │       │   ├── user.go        # 用户管理
│   │       │   ├── role.go        # 角色管理
│   │       │   ├── permission.go  # 权限管理
│   │       │   ├── org.go         # 联盟管理
│   │       │   ├── order.go       # 订单管理
│   │       │   ├── finance.go     # 财务结算
│   │       │   ├── message.go     # 消息通知
│   │       │   ├── dashboard.go   # 数据大屏
│   │       │   └── report.go      # 数据报表
│   │       ├── model/             # GORM 数据模型
│   │       └── router/router.go   # 路由注册
│   └── user/ org/ order/ ...      # 其他微服务模块（预留扩展）
│
├── pkg/                           # 公共基础包
│   ├── auth/jwt.go                # JWT 生成/解析
│   ├── database/database.go       # GORM 初始化
│   ├── middleware/
│   │   ├── auth.go                # JWT 鉴权中间件
│   │   └── cors.go                # CORS 中间件
│   └── response/response.go       # 统一响应结构
│
├── frontend/                      # UniApp 前端
│   └── src/
│       ├── api/                   # 接口封装 + 数据归一化
│       ├── components/
│       │   ├── common/            # KpiCard / FilterPanel / Pagination / StatusBadge
│       │   └── layout/            # AppLayout / Sidebar / Topbar
│       ├── pages/                 # 各业务页面
│       │   ├── dashboard/         # 数据大屏
│       │   ├── users/             # 用户管理
│       │   ├── orgs/              # 联盟管理
│       │   ├── orders/            # 订单中心
│       │   ├── finance/           # 财务结算
│       │   ├── messages/          # 消息通知
│       │   ├── permissions/       # 权限配置
│       │   └── reports/           # 数据报表
│       ├── stores/                # Pinia 状态（user / theme）
│       ├── styles/
│       │   └── themes/            # theme-a/b/c 三套主题
│       └── utils/                 # auth / format / countup
│
├── deploy/
│   ├── sql/init.sql               # 建库建表 + 初始数据
│   ├── test_api.sh                # 全接口功能测试脚本
│   └── stress_test.sh             # 压力测试脚本（Apache Bench）
│
├── docs/                          # 项目文档
│   ├── architecture.md            # 架构设计
│   ├── api.md                     # 接口文档
│   └── deployment.md              # 部署指南
│
├── go.work                        # Go 多模块工作区
└── README.md
```

---

## 快速启动

### 前置依赖

- Go 1.21+
- Node.js 18+
- MySQL 8.x（或 Docker）

### 1. 启动数据库

```bash
# Docker 方式（推荐）
docker run -d --name union-mysql \
  -e MYSQL_ROOT_PASSWORD=123456 \
  -p 3306:3306 \
  mysql:8

# 初始化库表
docker exec -i union-mysql mysql -uroot -p123456 < deploy/sql/init.sql
```

### 2. 启动后端

```bash
cd app/gateway
go run cmd/server/main.go
# 默认监听 http://localhost:8080
```

> 配置文件：`app/gateway/configs/config.yaml`，可按需修改数据库连接和 JWT secret。

### 3. 启动前端

```bash
cd frontend
npm install
npm run dev:h5
# 默认访问 http://localhost:5173
```

### 默认账号

| 账号 | 密码 | 角色 |
|------|------|------|
| admin | admin123 | 超级管理员 |

---

## 前端说明

### 主题切换

前端内置三套 CSS 变量主题，在右上角设置面板中可实时切换：

| 主题 | 描述 |
|------|------|
| 主题 A | 科技蓝（默认） |
| 主题 B | 清新绿 |
| 主题 C | 深色模式 |

### 图表交互

所有统计图均使用原生 SVG 实现（H5 平台）：

- **折线图**：平滑贝塞尔曲线 + 渐变面积，鼠标移入展示十字准线和数据浮窗
- **圆环图**：各扇区独立 hover 高亮，中心数值联动切换
- **柱状/排行**：hover 展开精确数值

---

## 接口文档

详见 [docs/api.md](docs/api.md)

基础信息：

- Base URL：`http://localhost:8080`
- 认证方式：`Authorization: Bearer <token>`
- 统一响应格式：

```json
{
  "code": 0,
  "message": "ok",
  "data": { ... }
}
```

| code | 含义 |
|------|------|
| 0 | 成功 |
| 400 | 参数错误 |
| 401 | 未授权 / Token 过期 |
| 500 | 服务器错误 |

---

## 测试

### 功能测试（49 个用例）

```bash
chmod +x deploy/test_api.sh
./deploy/test_api.sh
```

覆盖范围：Auth → 用户 CRUD → 角色权限 → 联盟成员 → 订单 → 财务 → 消息 → 大盘 → 报表

### 压力测试

```bash
# 依赖 Apache Bench（macOS：brew install httpd）
chmod +x deploy/stress_test.sh
./deploy/stress_test.sh
```

---

## 部署

详见 [docs/deployment.md](docs/deployment.md)

---

## License

MIT © 2026 fatbossplus
