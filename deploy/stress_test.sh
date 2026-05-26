#!/usr/bin/env bash
# ============================================================
# 联盟管理中心 — 压力测试脚本（Apache Bench）
# Usage: bash deploy/stress_test.sh
# ============================================================
BASE="http://localhost:8080/api/v1"
BOLD='\033[1m'; GREEN='\033[0;32m'; RED='\033[0;31m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; NC='\033[0m'

echo -e "${BOLD}=== 联盟管理中心 压力测试 ===${NC}"
echo "工具: Apache Bench (ab)"
echo "时间: $(date)"
echo ""

# ── 获取 Token ──────────────────────────────────────────────
echo -e "${YELLOW}→ 获取认证 Token...${NC}"
TOKEN=$(curl -s -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | \
  python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")
if [ -z "$TOKEN" ]; then echo -e "${RED}Token 获取失败，退出${NC}"; exit 1; fi
echo -e "${GREEN}Token 获取成功${NC}"

# 压测参数
N_LIGHT=500    # 轻量测试请求数
N_HEAVY=200    # 复杂查询请求数
N_LOGIN=300    # 登录并发
C_LOW=10       # 低并发
C_MED=50       # 中并发
C_HIGH=100     # 高并发

section() {
  echo ""
  echo -e "${BOLD}${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
  echo -e "${BOLD}${CYAN} $1${NC}"
  echo -e "${BOLD}${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

run_ab() {
  local name="$1" n="$2" c="$3" url="$4" method="${5:-GET}" body="${6:-}"
  echo -e "\n${YELLOW}▶ $name${NC}  n=$n c=$c"

  if [ "$method" = "POST" ] && [ -n "$body" ]; then
    echo "$body" > /tmp/ab_body.json
    RESULT=$(ab -n "$n" -c "$c" -T "application/json" \
      -H "Authorization: Bearer $TOKEN" \
      -p /tmp/ab_body.json \
      "$url" 2>&1)
  else
    RESULT=$(ab -n "$n" -c "$c" \
      -H "Authorization: Bearer $TOKEN" \
      "$url" 2>&1)
  fi

  RPS=$(echo "$RESULT" | grep "Requests per second" | awk '{print $4}')
  P50=$(echo "$RESULT" | grep "50%" | awk '{print $2}')
  P99=$(echo "$RESULT" | grep "99%" | awk '{print $2}')
  FAIL=$(echo "$RESULT" | grep "Failed requests" | awk '{print $3}')
  NON2XX=$(echo "$RESULT" | grep "Non-2xx responses" | awk '{print $3}')

  echo    "  QPS    : ${RPS} req/s"
  echo    "  P50延迟: ${P50} ms"
  echo    "  P99延迟: ${P99} ms"

  ERR=0
  [ -n "$FAIL" ] && [ "$FAIL" != "0" ] && { echo -e "  ${RED}✗ 失败请求: $FAIL${NC}"; ERR=1; }
  [ -n "$NON2XX" ] && { echo -e "  ${RED}✗ 非2xx响应: $NON2XX${NC}"; ERR=1; }
  [ "$ERR" = "0" ] && echo -e "  ${GREEN}✓ 零失败${NC}"
}

# ── 1. 登录接口压测 ──────────────────────────────────────────
section "1. 认证接口（无需 Token）"
run_ab "POST /auth/login (低并发 c=10)"    $N_LOGIN $C_LOW  "$BASE/auth/login" POST '{"username":"admin","password":"admin123"}'
run_ab "POST /auth/login (中并发 c=50)"    $N_LOGIN $C_MED  "$BASE/auth/login" POST '{"username":"admin","password":"admin123"}'
run_ab "POST /auth/login (高并发 c=100)"   $N_LOGIN $C_HIGH "$BASE/auth/login" POST '{"username":"admin","password":"admin123"}'

# ── 2. 用户列表（带分页/缓存友好）──────────────────────────────
section "2. 用户模块"
run_ab "GET /users?page=1&page_size=20 (c=10)"  $N_LIGHT $C_LOW  "$BASE/users?page=1&page_size=20"
run_ab "GET /users?page=1&page_size=20 (c=50)"  $N_LIGHT $C_MED  "$BASE/users?page=1&page_size=20"
run_ab "GET /users?page=1&page_size=20 (c=100)" $N_HEAVY $C_HIGH "$BASE/users?page=1&page_size=20"
run_ab "GET /users/me (高频个人信息)"            $N_LIGHT $C_MED  "$BASE/users/me"

# ── 3. 联盟 ─────────────────────────────────────────────────
section "3. 联盟模块"
run_ab "GET /orgs (c=50)"    $N_LIGHT $C_MED  "$BASE/orgs"
run_ab "GET /orgs (c=100)"   $N_HEAVY $C_HIGH "$BASE/orgs"

# ── 4. 订单 ─────────────────────────────────────────────────
section "4. 订单模块"
run_ab "GET /orders (c=50)"           $N_LIGHT $C_MED  "$BASE/orders"
run_ab "GET /orders?status=2 (c=100)" $N_HEAVY $C_HIGH "$BASE/orders?status=2"

# ── 5. 大屏统计（聚合查询）──────────────────────────────────
section "5. 大屏统计（数据库聚合）"
run_ab "GET /dashboard/stats (c=10)"     $N_HEAVY $C_LOW  "$BASE/dashboard/stats"
run_ab "GET /dashboard/stats (c=50)"     $N_HEAVY $C_MED  "$BASE/dashboard/stats"
run_ab "GET /dashboard/trend (c=50)"     $N_HEAVY $C_MED  "$BASE/dashboard/trend?period=month"
run_ab "GET /dashboard/org-rank (c=50)"  $N_HEAVY $C_MED  "$BASE/dashboard/org-rank"

# ── 6. 报表（最重SQL）──────────────────────────────────────
section "6. 报表接口（最重聚合 SQL）"
run_ab "GET /reports/summary (c=10)"   $N_HEAVY $C_LOW "$BASE/reports/summary"
run_ab "GET /reports/daily?days=30 (c=10)" $N_HEAVY $C_LOW "$BASE/reports/daily?days=30"
run_ab "GET /reports/roles (c=20)"     $N_HEAVY 20     "$BASE/reports/roles"

# ── 7. 消息 ─────────────────────────────────────────────────
section "7. 消息接口"
run_ab "GET /messages (c=50)" $N_LIGHT $C_MED "$BASE/messages"

echo ""
echo -e "${BOLD}${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BOLD}压力测试完成 — $(date)${NC}"
echo -e "${BOLD}${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
