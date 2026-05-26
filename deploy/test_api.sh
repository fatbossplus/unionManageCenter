#!/usr/bin/env bash
# ============================================================
# 联盟管理中心 — 全接口功能测试脚本
# Usage: bash deploy/test_api.sh
# ============================================================
BASE="http://localhost:8080/api/v1"
PASS=0; FAIL=0; TOTAL=0

GREEN='\033[0;32m'; RED='\033[0;31m'; YELLOW='\033[1;33m'; NC='\033[0m'; BOLD='\033[1m'

check() {
  local name="$1" method="$2" url="$3" body="$4" expect_code="${5:-0}"
  TOTAL=$((TOTAL+1))
  if [ -n "$body" ]; then
    resp=$(curl -s -X "$method" "$url" -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d "$body")
  else
    resp=$(curl -s -X "$method" "$url" -H "Authorization: Bearer $TOKEN")
  fi
  code=$(echo "$resp" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('code',999))" 2>/dev/null)
  if [ "$code" = "$expect_code" ]; then
    echo -e "  ${GREEN}✓${NC} $name"
    PASS=$((PASS+1))
  else
    echo -e "  ${RED}✗${NC} $name  (code=$code expect=$expect_code)"
    echo "    resp: $(echo "$resp" | head -c 200)"
    FAIL=$((FAIL+1))
  fi
}

check_no_auth() {
  local name="$1" method="$2" url="$3" body="$4" expect_code="${5:-0}"
  TOTAL=$((TOTAL+1))
  if [ -n "$body" ]; then
    resp=$(curl -s -X "$method" "$url" -H "Content-Type: application/json" -d "$body")
  else
    resp=$(curl -s -X "$method" "$url")
  fi
  code=$(echo "$resp" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('code',999))" 2>/dev/null)
  if [ "$code" = "$expect_code" ]; then
    echo -e "  ${GREEN}✓${NC} $name"
    PASS=$((PASS+1))
    echo "$resp"
  else
    echo -e "  ${RED}✗${NC} $name  (code=$code expect=$expect_code)"
    echo "    resp: $(echo "$resp" | head -c 300)"
    FAIL=$((FAIL+1))
  fi
}

section() { echo -e "\n${BOLD}${YELLOW}▶ $1${NC}"; }

echo -e "${BOLD}=== 联盟管理中心 接口功能测试 ===${NC}"
echo "BASE: $BASE"

# ─── 1. 认证 ────────────────────────────────────────────────
section "1. 认证模块"

echo -e "  ${YELLOW}→ 登录 (正确凭证)${NC}"
LOGIN_RESP=$(curl -s -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}')
LOGIN_CODE=$(echo "$LOGIN_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['code'])" 2>/dev/null)
TOKEN=$(echo "$LOGIN_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])" 2>/dev/null)
if [ "$LOGIN_CODE" = "0" ] && [ -n "$TOKEN" ]; then
  echo -e "  ${GREEN}✓${NC} POST /auth/login — token 获取成功"
  PASS=$((PASS+1)); TOTAL=$((TOTAL+1))
else
  echo -e "  ${RED}✗${NC} POST /auth/login — 失败: $LOGIN_RESP"
  FAIL=$((FAIL+1)); TOTAL=$((TOTAL+1))
  exit 1
fi

check  "POST /auth/login 错误密码 → 400"      POST "$BASE/auth/login" '{"username":"admin","password":"wrong"}' 400
check  "GET  /auth/menus (带Token)"            GET  "$BASE/auth/menus"
check  "POST /auth/logout"                     POST "$BASE/auth/logout" '{}'

# 无 Token 应返回 401
TOTAL=$((TOTAL+1))
NO_AUTH=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/users")
if [ "$NO_AUTH" = "401" ]; then
  echo -e "  ${GREEN}✓${NC} GET /users 无Token → HTTP 401"
  PASS=$((PASS+1))
else
  echo -e "  ${RED}✗${NC} GET /users 无Token → HTTP $NO_AUTH (expect 401)"
  FAIL=$((FAIL+1))
fi

# 重新登录（logout 后 token 依然有效，JWT 无状态）
TOKEN=$(echo "$LOGIN_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")

# ─── 2. 用户模块 ─────────────────────────────────────────────
section "2. 用户模块"
check "GET  /users (列表)"          GET  "$BASE/users"
check "GET  /users?page=1&page_size=5" GET "$BASE/users?page=1&page_size=5"
check "GET  /users/me"             GET  "$BASE/users/me"
check "GET  /users/1"              GET  "$BASE/users/1"

# 创建用户
TS=$(date +%s)
CREATE_USER=$(curl -s -X POST "$BASE/users" \
  -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
  -d "{\"username\":\"test_user_${TS}\",\"password\":\"test123456\",\"email\":\"test${TS}@test.com\",\"real_name\":\"测试用户\",\"status\":1}")
CREATE_CODE=$(echo "$CREATE_USER" | python3 -c "import sys,json; print(json.load(sys.stdin)['code'])" 2>/dev/null)
NEW_UID=$(echo "$CREATE_USER" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null)
TOTAL=$((TOTAL+1))
if [ "$CREATE_CODE" = "0" ] && [ -n "$NEW_UID" ]; then
  echo -e "  ${GREEN}✓${NC} POST /users — 创建成功 id=$NEW_UID"
  PASS=$((PASS+1))
else
  echo -e "  ${RED}✗${NC} POST /users — 创建失败: $CREATE_USER"
  FAIL=$((FAIL+1)); NEW_UID=1
fi

check "PUT  /users/$NEW_UID (更新)"   PUT "$BASE/users/$NEW_UID" '{"real_name":"已更新姓名"}'
check "POST /users/batch-enable"     POST "$BASE/users/batch-enable" "{\"ids\":[$NEW_UID]}"
check "POST /users/batch-disable"    POST "$BASE/users/batch-disable" "{\"ids\":[$NEW_UID]}"
check "DELETE /users/$NEW_UID"       DELETE "$BASE/users/$NEW_UID"

# ─── 3. 角色模块 ─────────────────────────────────────────────
section "3. 角色模块"
check "GET  /roles"                 GET  "$BASE/roles"

CREATE_ROLE=$(curl -s -X POST "$BASE/roles" \
  -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
  -d "{\"name\":\"测试角色${TS}\",\"code\":\"test_role_${TS}\",\"description\":\"自动测试角色\",\"status\":1}")
ROLE_CODE=$(echo "$CREATE_ROLE" | python3 -c "import sys,json; print(json.load(sys.stdin)['code'])" 2>/dev/null)
NEW_RID=$(echo "$CREATE_ROLE" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null)
TOTAL=$((TOTAL+1))
if [ "$ROLE_CODE" = "0" ]; then
  echo -e "  ${GREEN}✓${NC} POST /roles — 创建成功 id=$NEW_RID"
  PASS=$((PASS+1))
else
  echo -e "  ${RED}✗${NC} POST /roles — $CREATE_ROLE"
  FAIL=$((FAIL+1)); NEW_RID=1
fi
check "PUT  /roles/$NEW_RID"         PUT  "$BASE/roles/$NEW_RID" '{"description":"已更新"}'
check "PUT  /roles/$NEW_RID/permissions" PUT "$BASE/roles/$NEW_RID/permissions" '{"permission_ids":[1,2,3]}'
check "DELETE /roles/$NEW_RID"       DELETE "$BASE/roles/$NEW_RID"

# ─── 4. 权限模块 ─────────────────────────────────────────────
section "4. 权限模块"
check "GET  /permissions (树形)"    GET  "$BASE/permissions"

# ─── 5. 联盟模块 ─────────────────────────────────────────────
section "5. 联盟模块"
check "GET  /orgs"                  GET  "$BASE/orgs"
check "GET  /orgs?type=ec&status=1" GET  "$BASE/orgs?type=ec&status=1"

CREATE_ORG=$(curl -s -X POST "$BASE/orgs" \
  -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
  -d "{\"name\":\"测试联盟${TS}\",\"type\":\"ec\",\"description\":\"自动测试\",\"region\":\"测试地区\"}")
ORG_CODE=$(echo "$CREATE_ORG" | python3 -c "import sys,json; print(json.load(sys.stdin)['code'])" 2>/dev/null)
NEW_OID=$(echo "$CREATE_ORG" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null)
TOTAL=$((TOTAL+1))
if [ "$ORG_CODE" = "0" ]; then
  echo -e "  ${GREEN}✓${NC} POST /orgs — 创建成功 id=$NEW_OID"
  PASS=$((PASS+1))
else
  echo -e "  ${RED}✗${NC} POST /orgs — $CREATE_ORG"
  FAIL=$((FAIL+1)); NEW_OID=1
fi
check "GET  /orgs/$NEW_OID"         GET  "$BASE/orgs/$NEW_OID"
check "PUT  /orgs/$NEW_OID"         PUT  "$BASE/orgs/$NEW_OID" '{"status":1}'
check "GET  /orgs/$NEW_OID/members" GET  "$BASE/orgs/$NEW_OID/members"
check "POST /orgs/$NEW_OID/members" POST "$BASE/orgs/$NEW_OID/members" '{"user_id":1,"role":"admin"}'
check "DELETE /orgs/$NEW_OID/members/1" DELETE "$BASE/orgs/$NEW_OID/members/1"
check "DELETE /orgs/$NEW_OID"       DELETE "$BASE/orgs/$NEW_OID"

# ─── 6. 订单模块 ─────────────────────────────────────────────
section "6. 订单模块"
check "GET  /orders"                GET  "$BASE/orders"
check "GET  /orders?status=1"       GET  "$BASE/orders?status=1"
check "GET  /orders?page=1&page_size=10" GET "$BASE/orders?page=1&page_size=10"

# ─── 7. 财务模块 ─────────────────────────────────────────────
section "7. 财务模块"
check "GET  /finance"               GET  "$BASE/finance"
check "GET  /finance/accounts"      GET  "$BASE/finance/accounts"

# ─── 8. 消息模块 ─────────────────────────────────────────────
section "8. 消息模块"
check "GET  /messages"              GET  "$BASE/messages"
check "GET  /messages?is_read=0"    GET  "$BASE/messages?is_read=0"
check "POST /messages/read-all"     POST "$BASE/messages/read-all" '{}'

# ─── 9. 大屏统计 ─────────────────────────────────────────────
section "9. 大屏统计"
check "GET  /dashboard/stats"       GET  "$BASE/dashboard/stats"
check "GET  /dashboard/trend?period=month"   GET "$BASE/dashboard/trend?period=month"
check "GET  /dashboard/trend?period=quarter" GET "$BASE/dashboard/trend?period=quarter"
check "GET  /dashboard/trend?period=year"    GET "$BASE/dashboard/trend?period=year"
check "GET  /dashboard/org-types"   GET  "$BASE/dashboard/org-types"
check "GET  /dashboard/org-rank"    GET  "$BASE/dashboard/org-rank"
check "GET  /dashboard/events"      GET  "$BASE/dashboard/events"

# ─── 10. 报表模块 ────────────────────────────────────────────
section "10. 报表模块"
check "GET  /reports/summary"       GET  "$BASE/reports/summary"
check "GET  /reports/daily?days=7"  GET  "$BASE/reports/daily?days=7"
check "GET  /reports/daily?days=30" GET  "$BASE/reports/daily?days=30"
check "GET  /reports/daily?days=90" GET  "$BASE/reports/daily?days=90"
check "GET  /reports/roles"         GET  "$BASE/reports/roles"

# ─── 汇总 ────────────────────────────────────────────────────
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${BOLD}测试结果: 总计 $TOTAL | ${GREEN}通过 $PASS${NC} | ${RED}失败 $FAIL${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
[ "$FAIL" = "0" ] && echo -e "${GREEN}🎉 全部通过！${NC}" || echo -e "${RED}⚠️  存在失败项，请检查上方日志${NC}"

# 清理测试数据
docker exec mysql mysql -uroot -p123456 union_manage \
  -e "DELETE FROM users WHERE username LIKE 'test_user_%'; DELETE FROM roles WHERE code LIKE 'test_role_%'; DELETE FROM orgs WHERE name LIKE '测试联盟%'; UPDATE users SET status=1,deleted_at=NULL WHERE username='admin';" 2>/dev/null
echo -e "${YELLOW}(测试数据已清理)${NC}"
