# 前端部署指南

本文档提供了如何打包前端代码并将其集成到后端服务的详细说明。项目已提供多种部署方式，包括Windows批处理脚本和Docker容器化部署。

## 1. 打包前端代码

### 前提条件

- Node.js 环境 (推荐 v18+)
- npm 包管理器

### 打包步骤

1. 进入前端项目目录：

```bash
cd frontend
```

2. 安装依赖：

```bash
npm install
```

3. 构建生产环境代码：

```bash
npm run build
```

构建完成后，会在 `frontend/dist` 目录下生成打包后的静态文件。

## 2. 集成到后端服务

有两种方式可以将前端代码集成到后端服务：

### 方式一：使用后端服务静态文件功能

1. 将前端构建产物复制到后端静态资源目录：

```bash
# 在项目根目录执行
mkdir -p backend/static
cp -r frontend/dist/* backend/static/
```

2. 修改后端代码，添加静态文件服务：

```go
// 在 main.go 中添加

// 设置静态文件服务
r.Static("/static", "./static")

// 添加前端入口路由
r.NoRoute(func(c *gin.Context) {
    c.File("./static/index.html")
})
```

### 方式二：使用 Docker 多阶段构建

项目已提供 `Dockerfile` 和 `docker-compose.yml` 文件，可以直接使用 Docker 进行部署。

#### Dockerfile

```dockerfile
# 第一阶段：构建前端
FROM node:18 as frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# 第二阶段：构建后端
FROM golang:1.20 as backend-builder
WORKDIR /app
COPY backend/ ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o telegram-photo-server .

# 第三阶段：最终镜像
FROM alpine:latest
WORKDIR /app

# 安装运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 复制构建产物
COPY --from=backend-builder /app/telegram-photo-server .
COPY --from=frontend-builder /app/frontend/dist ./static
COPY backend/config.yaml ./

# 创建上传目录
RUN mkdir -p uploads

EXPOSE 8080
CMD ["./telegram-photo-server"]
```

#### docker-compose.yml

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

#### 使用 Docker 部署

1. 使用 Dockerfile 构建并运行：

```bash
# 构建镜像
docker build -t telegram-photo .

# 运行容器
docker run -p 8080:8080 -v $(pwd)/uploads:/app/uploads -v $(pwd)/config.yaml:/app/config.yaml telegram-photo
```

2. 使用 docker-compose 部署（推荐）：

```bash
# 构建并启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

## 3. 使用提供的部署工具

项目已提供多种部署工具，可以根据需要选择使用。

### 3.1 Windows批处理脚本（deploy.bat）

项目根目录下提供了 `deploy.bat` 批处理脚本，可以一键完成前端构建和部署：

```batch
@echo off
echo 正在构建前端应用...
cd frontend
call npm run build

echo 创建后端静态文件目录...
cd ..
mkdir backend\static 2>nul

echo 复制前端构建产物到后端静态目录...
xcopy /E /Y frontend\dist\* backend\static\

echo 部署完成！
echo 您可以使用 main_with_frontend.go 替换 main.go 来启动集成了前端的后端服务
echo 或者直接运行: cd backend && go run main_with_frontend.go
```

使用方法：

1. 在项目根目录下双击运行 `deploy.bat`
2. 脚本会自动构建前端代码并复制到后端静态目录
3. 完成后，可以使用 `main_with_frontend.go` 启动服务

### 3.2 集成前端的后端代码（main_with_frontend.go）

项目提供了 `backend/main_with_frontend.go` 文件，已经包含了前端静态文件服务的配置：

```go
package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/telegram-photo/api/v1"
	"github.com/telegram-photo/config"
	"github.com/telegram-photo/middleware"
	"github.com/telegram-photo/model"
)

func main() {
	// 加载配置
	if err := config.Init(); err != nil {
		log.Fatalf("配置初始化失败: %v", err)
	}

	// 初始化数据库
	if err := model.Init(); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	// 创建Gin引擎
	r := gin.Default()

	// 注册中间件
	r.Use(middleware.Cors())

	// 注册API路由
	v1.RegisterRoutes(r)

	// 静态文件服务 - 前端构建产物
	r.Static("/assets", "./static/assets")
	r.StaticFile("/favicon.ico", "./static/favicon.ico")

	// 处理前端路由 - 将非API请求转发到前端入口文件
	r.NoRoute(func(c *gin.Context) {
		// 如果是API请求，返回404
		if strings.HasPrefix(c.Request.URL.Path, "/api/") || strings.HasPrefix(c.Request.URL.Path, "/proxy/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "API not found"})
			return
		}

		// 否则返回前端入口文件
		c.File("./static/index.html")
	})

	// 启动服务器
	port := viper.GetString("server.port")
	if port == "" {
		port = "8080"
	}

	log.Printf("服务器启动在 http://localhost:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
```

使用方法：

```bash
cd backend
go run main_with_frontend.go
```

## 4. 生产环境配置

### 环境变量

在生产环境中，需要设置正确的环境变量：

1. 创建 `.env.production` 文件在前端目录：

```
VITE_API_BASE=/api
```

2. 重新构建前端代码：

```bash
cd frontend
npm run build
```

### 反向代理配置

如果使用 Nginx 作为反向代理，可以使用以下配置：

```nginx
server {
    listen 80;
    server_name your-domain.com;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 5. GitHub Actions 自动打包部署

项目已配置 GitHub Actions 工作流，可以实现代码推送后自动构建和部署。

### 5.1 构建工作流 (build.yml)

```yaml
name: Build and Deploy

on:
  push:
    branches: [ main, master ]
  pull_request:
    branches: [ main, master ]

jobs:
  build:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v3

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.20'

    - name: Set up Node.js
      uses: actions/setup-node@v3
      with:
        node-version: '18'
        cache: 'npm'
        cache-dependency-path: frontend/package-lock.json

    - name: Install frontend dependencies
      run: cd frontend && npm ci

    - name: Build frontend
       run: cd frontend && npm run build

     - name: Create dist directory
       run: mkdir -p backend/dist

     - name: Copy frontend build to dist directory
       run: cp -r frontend/dist/* backend/dist/

     - name: Build backend
       run: cd backend && go build -o telegram-photo-server .

    - name: Upload artifact
      uses: actions/upload-artifact@v3
      with:
        name: telegram-photo
        path: |
          backend/telegram-photo-server
          backend/static
          backend/config.yaml
```

### 5.2 Docker 构建和发布工作流 (docker.yml)

```yaml
name: Docker Build and Push

on:
  push:
    branches: [ main, master ]
    tags: [ 'v*' ]
  pull_request:
    branches: [ main, master ]

jobs:
  docker:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v3

      - name: Set up QEMU
        uses: docker/setup-qemu-action@v2

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v2

      - name: Login to Docker Hub
        if: github.event_name != 'pull_request'
        uses: docker/login-action@v2
        with:
          username: ${{ secrets.DOCKERHUB_USERNAME }}
          password: ${{ secrets.DOCKERHUB_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v4
        with:
          images: ${{ secrets.DOCKERHUB_USERNAME }}/telegram-photo
          tags: |
            type=ref,event=branch
            type=ref,event=pr
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=sha

      - name: Build and push
        uses: docker/build-push-action@v4
        with:
          context: .
          push: ${{ github.event_name != 'pull_request' }}
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

### 5.3 配置 GitHub Secrets

要使用 Docker Hub 发布功能，需要在 GitHub 仓库设置中添加以下 Secrets：

1. `DOCKERHUB_USERNAME`: Docker Hub 用户名
2. `DOCKERHUB_TOKEN`: Docker Hub 访问令牌

### 5.4 使用方法

1. 将代码推送到 GitHub 仓库的 main 或 master 分支
2. GitHub Actions 会自动触发构建流程
3. 构建完成后，可以在 Actions 页面下载构建产物
4. 如果配置了 Docker Hub 凭据，还会自动构建并推送 Docker 镜像

## 6. 总结

通过以上步骤，您可以将前端代码打包并集成到后端服务中，实现单一服务部署。这种方式简化了部署流程，减少了维护成本。

项目提供了多种部署方式：

1. Windows 批处理脚本（deploy.bat）
2. Docker 容器化部署（Dockerfile 和 docker-compose.yml）
3. GitHub Actions 自动构建和部署

您可以根据实际需求选择合适的部署方式。对于更复杂的部署需求，可以考虑使用 Kubernetes 进行容器编排。