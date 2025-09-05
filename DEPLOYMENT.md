# Telegram 图床部署指南

本文档提供了 Telegram 图床项目的部署和调试指南。

## 本地开发环境部署

### 前置条件

1. 安装 Go 1.16+
2. 安装 Node.js 14+
3. 安装 npm 或 yarn
4. 创建 Telegram Bot 并获取 Bot Token
5. 创建 GitHub OAuth 应用并获取 Client ID 和 Client Secret

### 后端部署

1. 进入后端目录：

```bash
cd backend
```

2. 修改 `config.yaml` 文件，填入正确的配置：

```yaml
# 数据库配置
database:
  driver: sqlite3
  connection: telegram_photo.db

# 服务器配置
server:
  port: 8080
  jwt_secret: your_jwt_secret_key_change_this_in_production
  jwt_expire: 168h  # 7天

# Telegram配置
telegram:
  bot_token: your_telegram_bot_token
  api_url: https://api.telegram.org

# GitHub OAuth配置
github:
  client_id: your_github_client_id
  client_secret: your_github_client_secret
  redirect_uri: http://localhost:8080/api/v1/auth/github/callback
  # 前端回调地址
  frontend_callback: http://localhost:5173/auth/callback

# 管理员配置
admin:
  user_ids:
    - github_user_id_1
    - github_user_id_2
```

3. 安装依赖并运行：

```bash
go mod tidy
go run main.go
```

### 前端部署

1. 进入前端目录：

```bash
cd frontend
```

2. 修改 `.env` 文件，填入管理员 GitHub 用户 ID：

```
VITE_ADMIN_IDS=github_user_id_1,github_user_id_2
```

3. 安装依赖并运行：

```bash
npm install
npm run dev
```

4. 访问 http://localhost:5173 即可使用应用

## 生产环境部署

### 后端部署

1. 编译后端：

```bash
cd backend
go build -o telegram-photo-server
```

2. 配置 `config.yaml` 文件，注意修改以下内容：
   - 设置更安全的 `jwt_secret`
   - 更新 GitHub OAuth 的 `redirect_uri` 和 `frontend_callback` 为生产环境 URL

3. 运行服务：

```bash
./telegram-photo-server
```

### 前端部署

1. 编译前端：

```bash
cd frontend
npm run build
```

2. 将 `dist` 目录下的文件部署到 Web 服务器（如 Nginx）

3. 配置 Nginx：

```nginx
server {
    listen 80;
    server_name your-domain.com;

    root /path/to/dist;
    index index.html;

    # 前端路由处理
    location / {
        try_files $uri $uri/ /index.html;
    }

    # API 代理
    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # 图片代理
    location /proxy/ {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 调试指南

### 后端调试

1. 检查日志输出，查看是否有错误信息

2. 确认 Telegram Bot Token 是否有效：
   - 可以通过访问 `https://api.telegram.org/bot{your_bot_token}/getMe` 验证

3. 确认 GitHub OAuth 配置是否正确：
   - 检查 GitHub 开发者设置中的 OAuth 应用配置
   - 确认 redirect_uri 是否与配置文件一致

4. 检查数据库连接：
   - 确认 SQLite 数据库文件是否存在且可写

### 前端调试

1. 使用浏览器开发者工具检查网络请求和控制台错误

2. 确认 API 请求地址是否正确：
   - 开发环境中应该是 `http://localhost:8080`
   - 生产环境中应该是相对路径 `/api`

3. 检查 JWT 认证：
   - 确认 localStorage 中是否存在 token
   - 确认请求头中是否包含 `Authorization: Bearer {token}`

## 常见问题

### 1. GitHub 登录失败

- 检查 GitHub OAuth 应用的 Client ID 和 Client Secret 是否正确
- 确认 redirect_uri 是否与 GitHub OAuth 应用配置一致
- 检查前端回调地址是否正确

### 2. 图片上传失败

- 确认 Telegram Bot Token 是否有效
- 检查图片大小是否超过限制（20MB）
- 检查网络连接是否正常

### 3. 管理员功能无法访问

- 确认当前登录的 GitHub 用户 ID 是否在管理员列表中
- 检查前端 `.env` 文件中的 `VITE_ADMIN_IDS` 配置是否正确
- 检查后端 `config.yaml` 文件中的 `admin.user_ids` 配置是否正确

### 4. 代理访问图片失败

- 确认图片 ID 是否存在于数据库中
- 检查 Telegram API 是否可以正常访问
- 检查网络连接是否正常