# 部署指南

## 环境要求

| 组件 | 最低版本 | 推荐版本 |
|------|----------|----------|
| Go | 1.21 | 1.25 |
| Node.js | 18.x | 20.x |
| MySQL | 5.7 | 8.x |
| 操作系统 | Linux x86_64 | Ubuntu 22.04 / CentOS 8 |

---

## 一、开发环境

### 1. 克隆项目

```bash
git clone https://github.com/fatbossplus/unionManageCenter.git
cd unionManageCenter
```

### 2. 启动 MySQL（Docker）

```bash
docker run -d --name union-mysql \
  -e MYSQL_ROOT_PASSWORD=123456 \
  -p 3306:3306 \
  --restart unless-stopped \
  mysql:8

# 等待 MySQL 就绪（约 30 秒）
sleep 30

# 初始化库表 + 初始数据
docker exec -i union-mysql mysql -uroot -p123456 < deploy/sql/init.sql
```

### 3. 配置后端

编辑 `app/gateway/configs/config.yaml`：

```yaml
server:
  port: 8080
  mode: debug        # 生产环境改为 release

database:
  host: 127.0.0.1
  port: 3306
  user: root
  password: "123456"  # 修改为实际密码
  dbname: union_manage

jwt:
  secret: "your-strong-secret-key-here"  # 生产环境必须修改
  expire: 24         # Token 有效期（小时）
```

### 4. 启动后端

```bash
cd app/gateway
go run cmd/server/main.go

# 或编译后运行
go build -o ../../bin/gateway cmd/server/main.go
../../bin/gateway
```

### 5. 启动前端

```bash
cd frontend
npm install
npm run dev:h5      # 访问 http://localhost:5173
```

---

## 二、生产环境部署

### 后端编译

```bash
# 在项目根目录
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags="-s -w" \
  -o bin/gateway \
  ./app/gateway/cmd/server/main.go
```

### 前端构建

```bash
cd frontend
npm run build:h5
# 产物在 frontend/dist/build/h5/
```

### Systemd 服务（Linux）

创建 `/etc/systemd/system/union-gateway.service`：

```ini
[Unit]
Description=Union Manage Center Gateway
After=network.target mysql.service

[Service]
Type=simple
User=www
WorkingDirectory=/opt/unionManageCenter/app/gateway
ExecStart=/opt/unionManageCenter/bin/gateway
Restart=on-failure
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
```

```bash
systemctl daemon-reload
systemctl enable union-gateway
systemctl start union-gateway
systemctl status union-gateway
```

### Nginx 配置

```nginx
# 前端 H5 静态文件
server {
    listen 80;
    server_name your-domain.com;

    root /opt/unionManageCenter/frontend/dist/build/h5;
    index index.html;

    # SPA history 模式
    location / {
        try_files $uri $uri/ /index.html;
    }

    # 代理后端 API
    location /api/ {
        rewrite ^/api/(.*) /$1 break;
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

---

## 三、Docker Compose 一键部署

创建 `docker-compose.yml`（可选，按需使用）：

```yaml
version: '3.8'

services:
  mysql:
    image: mysql:8
    environment:
      MYSQL_ROOT_PASSWORD: "StrongPassword!2026"
      MYSQL_DATABASE: union_manage
    volumes:
      - mysql_data:/var/lib/mysql
      - ./deploy/sql/init.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "3306:3306"
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5

  gateway:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      DB_HOST: mysql
      DB_PASSWORD: "StrongPassword!2026"
      JWT_SECRET: "your-production-secret"
    depends_on:
      mysql:
        condition: service_healthy

volumes:
  mysql_data:
```

```bash
docker compose up -d
```

---

## 四、环境变量（可选覆盖配置）

后端支持通过环境变量覆盖 `config.yaml`：

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| `SERVER_PORT` | 8080 | 监听端口 |
| `DB_HOST` | 127.0.0.1 | 数据库主机 |
| `DB_PORT` | 3306 | 数据库端口 |
| `DB_USER` | root | 数据库用户 |
| `DB_PASSWORD` | 123456 | 数据库密码 |
| `DB_NAME` | union_manage | 数据库名 |
| `JWT_SECRET` | — | JWT 密钥（生产必填） |

---

## 五、安全加固（生产必做）

1. **修改 JWT Secret**：`jwt.secret` 至少 32 位随机字符串
2. **修改数据库密码**：避免使用默认密码
3. **关闭 Debug 模式**：`server.mode: release`
4. **配置防火墙**：只开放 80/443，后端 8080 端口仅内网访问
5. **启用 HTTPS**：Nginx 配置 SSL 证书（推荐 Let's Encrypt）
6. **定期备份**：`mysqldump union_manage > backup_$(date +%Y%m%d).sql`

---

## 六、常见问题

**Q: 后端启动报 `Error 1049: Unknown database`**

```bash
# 手动创建数据库后重新执行 init.sql
mysql -uroot -p -e "CREATE DATABASE IF NOT EXISTS union_manage CHARACTER SET utf8mb4;"
mysql -uroot -p union_manage < deploy/sql/init.sql
```

**Q: 前端登录报网络错误**

检查 `frontend/src/api/request.ts` 中的 `BASE_URL` 是否指向正确后端地址。

**Q: 端口被占用**

```bash
# 查找占用端口的进程
lsof -ti:8080 | xargs kill -9
```

**Q: 压力测试提示 ab 未找到**

```bash
# macOS
brew install httpd

# Ubuntu / Debian
apt install apache2-utils
```
