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
- SQLite 数据库
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
database:
  driver: sqlite3
  connection: telegram_photo.db

server:
  port: 8080
  jwt_secret: your_jwt_secret

telegram:
  bot_token: your_telegram_bot_token

github:
  client_id: your_github_client_id
  client_secret: your_github_client_secret
  redirect_uri: http://localhost:8080/api/v1/auth/github/callback

admin:
  user_ids:
    - github_user_id_1
    - github_user_id_2
```

2. 运行后端：

```bash
cd backend
go mod tidy
go run main.go
```

### 前端配置

1. 创建 `.env` 文件：

```
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

## 许可证

MIT