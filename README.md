# Telegram 图床

基于 Telegram Bot API 的图床服务，支持图片上传、管理和代理访问。

## 功能特点

- 使用 GitHub OAuth 进行用户认证
- 图片上传到 Telegram 服务器，无需本地存储
- 图片代理访问，支持缓存
- 用户图片管理（查看、删除）
- 管理员统计和图片管理功能

## 技术栈

### 后端

- Go
- Gin 框架
- GORM
- MYSQL 数据库
- JWT 认证

### 前端

- Vue 3
- Vite
- Pinia 状态管理
- Vue Router
- Tailwind CSS

## 项目结构

```
├── backend/                # 后端代码
│   ├── api/                # API 处理函数
│   │   └── v1/             # API 版本 1
│   ├── config/             # 配置管理
│   ├── middleware/         # 中间件
│   ├── model/              # 数据模型
│   ├── service/            # 外部服务
│   └── main.go             # 入口文件
└── frontend/              # 前端代码
    ├── public/             # 静态资源
    └── src/                # 源代码
        ├── assets/         # 资源文件
        ├── components/     # 组件
        ├── router/         # 路由
        ├── store/          # 状态管理
        ├── views/          # 页面
        └── App.vue         # 主组件
```

## 安装和运行

### 环境要求

- Go 1.16+
- Node.js 14+
- npm 或 yarn

### 后端配置

1. 创建 `config.yaml` 文件：

```yaml
# 数据库配置
database:
  type: mysql  # 数据库类型，支持mysql、sqlite3
  host: localhost  # 数据库主机地址
  port: 3306  # 数据库端口
  user: root  # 数据库用户名
  password: password  # 数据库密码
  name: telegram_photo  # 数据库名称

# 服务器配置
server:
  port: 8080  # 服务器端口
  jwt_secret: your_jwt_secret_key_change_this_in_production  # JWT密钥
  jwt_expire: 168h  # JWT过期时间，7天

# Telegram配置
telegram:
  bot_token: your_telegram_bot_token  # Telegram Bot Token
  api_url: https://api.telegram.org  # Telegram API地址
  chat_id: your_telegram_chat_id  # Telegram聊天ID

# GitHub OAuth配置
github:
  client_id: your_github_client_id  # GitHub OAuth应用Client ID
  client_secret: your_github_client_secret  # GitHub OAuth应用Client Secret
  redirect_uri: http://localhost:8080/api/v1/auth/github/callback  # GitHub OAuth回调地址
  # 前端回调地址
  frontend_callback: http://localhost:8080/auth/callback  # 前端回调地址

# 管理员配置
admin:
  user_ids:  # 管理员GitHub用户ID列表
    - github_user_id_1
    - github_user_id_2
```

2. 运行后端：

```bash
cd backend
go mod tidy
go run main.go
```

### 前后端集成部署

项目提供了简便的部署脚本，可以将前端构建产物集成到后端服务中：

1. 使用部署脚本：

```bash
# 在项目根目录执行
.\deploy.bat
```

2. 运行集成后的应用：

```bash
cd backend
go run main_with_frontend.go
```

这将启动一个包含前端静态文件的后端服务，可以通过 http://localhost:8080 访问完整应用。

### Docker部署

项目支持使用Docker进行部署：

1. 使用Dockerfile构建镜像：

```bash
# 在项目根目录执行
docker build -t telegram-photo .

# 运行容器
docker run -p 8080:8080 -v $(pwd)/uploads:/app/uploads -v $(pwd)/config.yaml:/app/config.yaml telegram-photo
```

2. 使用docker-compose部署（推荐）：

```bash
# 在项目根目录执行
docker-compose up -d
```

docker-compose.yml文件配置如下：

```yaml
version: '3'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    volumes:
      - ./uploads:/app/uploads
      - ./config.yaml:/app/config.yaml
    restart: unless-stopped
    environment:
      - TZ=Asia/Shanghai
```

### 前端配置

1. 创建 `.env` 文件：

```
# API基础路径，开发环境指向后端服务
VITE_API_BASE=http://localhost:8080/api

# 代理路径，用于访问图片
VITE_PROXY_BASE=http://localhost:8080/proxy

# 管理员GitHub用户ID列表，与后端config.yaml中的admin.user_ids保持一致
VITE_ADMIN_IDS=github_user_id_1,github_user_id_2
```

2. 安装依赖并运行：

```bash
cd frontend
npm install
npm run dev
```

## API 接口

### 认证相关

- `GET /api/v1/auth/github` - 重定向到 GitHub 授权页面
- `GET /api/v1/auth/github/callback` - GitHub 授权回调

### 图片相关（需要认证）

- `POST /api/v1/image/upload` - 上传图片
- `GET /api/v1/image/list` - 获取用户图片列表
- `DELETE /api/v1/image/:id` - 删除图片

### 管理员接口（需要管理员权限）

- `GET /api/v1/admin/images` - 获取所有图片
- `GET /api/v1/admin/stats` - 获取统计信息

### 代理访问

- `GET /proxy/image/:file_id` - 代理访问图片

## GitHub Actions自动部署

项目已配置GitHub Actions工作流，可以实现代码推送后自动构建和部署：

1. 构建工作流（build.yml）：在代码推送到main/master分支或创建PR时自动构建前后端代码
2. Docker构建工作流（docker.yml）：自动构建Docker镜像并推送到Docker Hub

详细配置请参考项目根目录下的GITHUB_ACTIONS.md文件。

## 更多文档

- DEPLOYMENT.md：详细的部署指南
- FRONTEND_DEPLOYMENT.md：前端部署详细说明
- GITHUB_ACTIONS.md：GitHub Actions配置说明
- APIREADME.md：API接口详细文档

## 许可证

MIT