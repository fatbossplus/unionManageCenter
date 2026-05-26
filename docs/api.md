# 接口文档

- **Base URL**：`http://localhost:8080`
- **Content-Type**：`application/json`
- **认证**：除登录接口外，所有接口需在请求头携带 `Authorization: Bearer <token>`

## 统一响应格式

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

| code | 说明 |
|------|------|
| 0 | 成功 |
| 400 | 参数错误 |
| 401 | 未登录 / Token 无效 |
| 403 | 无权限 |
| 500 | 服务器内部错误 |

---

## 认证模块 `/auth`

### 登录

```
POST /auth/login
```

**请求体**

```json
{ "username": "admin", "password": "admin123" }
```

**响应**

```json
{
  "code": 0,
  "data": {
    "token": "eyJhbGci...",
    "user": {
      "id": 1,
      "username": "admin",
      "nickname": "超级管理员",
      "role_code": "superadmin"
    }
  }
}
```

### 获取当前用户信息

```
GET /users/me
```

---

## 用户管理 `/users`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/users` | 用户列表（支持分页 + 筛选） |
| POST | `/users` | 创建用户 |
| GET | `/users/:id` | 用户详情 |
| PUT | `/users/:id` | 更新用户 |
| DELETE | `/users/:id` | 删除用户 |
| POST | `/users/:id/enable` | 启用用户 |
| POST | `/users/:id/disable` | 停用用户 |
| POST | `/users/:id/roles` | 分配角色 |

**列表查询参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| page | int | 页码，默认 1 |
| page_size | int | 每页数，默认 20 |
| keyword | string | 用户名/昵称模糊搜索 |
| status | int | 状态：1 正常，2 停用 |
| role_code | string | 按角色筛选 |

---

## 角色管理 `/roles`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/roles` | 角色列表 |
| POST | `/roles` | 创建角色 |
| PUT | `/roles/:id` | 更新角色 |
| DELETE | `/roles/:id` | 删除角色 |
| POST | `/roles/:id/permissions` | 为角色分配权限 |

**创建角色请求体**

```json
{
  "name": "运营专员",
  "code": "operator",
  "description": "负责日常运营管理"
}
```

---

## 权限管理 `/permissions`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/permissions` | 权限树列表 |
| POST | `/permissions` | 创建权限 |
| PUT | `/permissions/:id` | 更新权限 |
| DELETE | `/permissions/:id` | 删除权限 |

**权限类型**

| type | 说明 |
|------|------|
| menu | 菜单权限 |
| button | 按钮权限 |
| api | 接口权限 |

---

## 联盟管理 `/orgs`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/orgs` | 联盟列表 |
| POST | `/orgs` | 创建联盟 |
| GET | `/orgs/:id` | 联盟详情 |
| PUT | `/orgs/:id` | 更新联盟 |
| DELETE | `/orgs/:id` | 删除联盟 |
| GET | `/orgs/:id/members` | 联盟成员列表 |
| POST | `/orgs/:id/members` | 添加联盟成员 |
| DELETE | `/orgs/:id/members/:uid` | 移除成员 |

---

## 订单管理 `/orders`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/orders` | 订单列表 |
| POST | `/orders` | 创建订单 |
| GET | `/orders/:id` | 订单详情 |
| POST | `/orders/:id/refund` | 订单退款 |

**订单状态**

| status | 说明 |
|--------|------|
| 1 | 待支付 |
| 2 | 已支付 |
| 3 | 已退款 |
| 4 | 已取消 |

---

## 财务结算 `/finance`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/finance/accounts` | 账户列表 |
| GET | `/finance/settlements` | 结算单列表 |
| POST | `/finance/settlements/:id/settle` | 手动触发结算 |

---

## 消息通知 `/messages`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/messages` | 消息列表 |
| POST | `/messages` | 发送消息 |
| POST | `/messages/:id/read` | 标记已读 |
| POST | `/messages/read-all` | 全部已读 |

---

## 数据大盘 `/dashboard`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/dashboard/stats` | 核心 KPI 指标 |
| GET | `/dashboard/trend` | 趋势数据（折线图） |
| GET | `/dashboard/org-type` | 联盟类型分布（圆环图） |
| GET | `/dashboard/org-rank` | 联盟流水排行 TOP5 |
| GET | `/dashboard/events` | 实时动态列表 |

**trend 查询参数**

| 参数 | 值 | 说明 |
|------|-----|------|
| period | month | 近 30 天（默认） |
| period | quarter | 近 90 天 |
| period | year | 近 365 天 |

**stats 响应示例**

```json
{
  "total_users": 1280,
  "active_orgs": 56,
  "monthly_revenue": 328600,
  "pending_orders": 12,
  "today_revenue": 8900,
  "today_new_users": 23,
  "online_users": 47
}
```

---

## 数据报表 `/reports`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/reports/summary` | 月度汇总 + 环比趋势 |
| GET | `/reports/daily` | 日粒度明细数据 |
| GET | `/reports/roles` | 角色分布统计 |

**daily 查询参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| days | int | 查询天数，默认 30 |

**daily 响应示例**

```json
[
  {
    "date": "2026-05-26",
    "new_users": 23,
    "active_users": 189,
    "revenue": 12400,
    "orders": 47,
    "trend": "↑ 12.3%"
  }
]
```

**summary 响应示例**

```json
{
  "month_new_users": 580,
  "month_revenue": 328600,
  "active_orgs": 56,
  "month_orders": 1204,
  "trends": {
    "users": "+15.2%",
    "revenue": "+8.7%",
    "orders": "+11.4%"
  }
}
```
