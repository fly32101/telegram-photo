# GitHub Actions 自动打包部署指南

本项目已配置 GitHub Actions 工作流，可以实现代码推送后自动构建和部署。本文档详细介绍如何使用和配置这些自动化工作流。

## 1. 可用的工作流

### 1.1 构建工作流 (build.yml)

此工作流用于构建前端和后端代码，并将构建产物作为 GitHub Artifacts 上传。

**触发条件**：
- 推送到 main 或 master 分支
- 对 main 或 master 分支发起 Pull Request

**执行步骤**：
1. 检出代码
2. 设置 Go 环境 (1.20)
3. 设置 Node.js 环境 (18)
4. 安装前端依赖
5. 构建前端代码
6. 创建后端dist目录
7. 复制前端构建产物到后端dist目录
8. 构建后端代码
9. 上传构建产物

**构建产物**：
- backend/telegram-photo-server (后端可执行文件)
- backend/dist (前端静态文件)
- backend/config.yaml (配置文件)

### 1.2 Docker 构建和发布工作流 (docker.yml)

此工作流用于构建 Docker 镜像并推送到 Docker Hub。

**触发条件**：
- 推送到 main 或 master 分支
- 推送标签 (以 v 开头)
- 对 main 或 master 分支发起 Pull Request

**执行步骤**：
1. 检出代码
2. 设置 QEMU (支持多架构构建)
3. 设置 Docker Buildx
4. 登录到 Docker Hub (仅在非 PR 事件时)
5. 提取元数据 (用于标记镜像)
6. 构建并推送 Docker 镜像 (仅在非 PR 事件时推送)

## 2. 配置 GitHub Secrets

要使用 Docker Hub 发布功能，需要在 GitHub 仓库设置中添加以下 Secrets：

1. `DOCKERHUB_USERNAME`: Docker Hub 用户名
2. `DOCKERHUB_TOKEN`: Docker Hub 访问令牌

### 2.1 获取 Docker Hub 访问令牌

1. 登录 [Docker Hub](https://hub.docker.com/)
2. 点击右上角的用户头像，选择 "Account Settings"
3. 在左侧菜单中选择 "Security"
4. 点击 "New Access Token"
5. 输入令牌名称和描述，选择适当的权限
6. 点击 "Generate" 生成令牌
7. 复制生成的令牌 (注意：令牌只会显示一次)

### 2.2 添加 GitHub Secrets

1. 在 GitHub 仓库页面，点击 "Settings"
2. 在左侧菜单中选择 "Secrets and variables" -> "Actions"
3. 点击 "New repository secret"
4. 添加 `DOCKERHUB_USERNAME` 和 `DOCKERHUB_TOKEN` 两个 Secrets

## 3. 使用方法

### 3.1 下载构建产物

1. 在 GitHub 仓库页面，点击 "Actions"
2. 选择 "Build and Deploy" 工作流
3. 点击最新的成功运行记录
4. 在 "Artifacts" 部分，点击 "telegram-photo" 下载构建产物

### 3.2 使用 Docker 镜像

如果配置了 Docker Hub 凭据，可以直接使用推送的 Docker 镜像：

```bash
# 拉取最新镜像
docker pull <DOCKERHUB_USERNAME>/telegram-photo:latest

# 运行容器
docker run -d -p 8080:8080 \
  -v /path/to/uploads:/app/uploads \
  -v /path/to/config.yaml:/app/config.yaml \
  <DOCKERHUB_USERNAME>/telegram-photo:latest
```

## 4. 自定义工作流

如果需要自定义工作流，可以编辑 `.github/workflows/` 目录下的 YAML 文件。

### 4.1 修改构建参数

可以修改 `build.yml` 文件中的以下参数：

- Go 版本：修改 `go-version` 字段
- Node.js 版本：修改 `node-version` 字段
- 构建产物路径：修改 `path` 字段

### 4.2 修改 Docker 构建参数

可以修改 `docker.yml` 文件中的以下参数：

- 镜像名称：修改 `images` 字段
- 标签规则：修改 `tags` 字段
- 构建上下文：修改 `context` 字段

## 5. 故障排除

### 5.1 构建失败

如果构建失败，可以查看工作流运行日志：

1. 在 GitHub 仓库页面，点击 "Actions"
2. 选择失败的工作流运行记录
3. 查看详细日志，找出失败原因

### 5.2 Docker 推送失败

如果 Docker 镜像推送失败，请检查：

1. GitHub Secrets 是否正确配置
2. Docker Hub 访问令牌是否有效
3. Docker Hub 用户名是否正确
4. Docker Hub 仓库是否存在

## 6. 最佳实践

1. **版本标签**：使用语义化版本标签 (如 v1.0.0) 来标记重要版本
2. **分支保护**：为 main/master 分支启用分支保护，要求 PR 审核
3. **测试集成**：考虑在工作流中添加测试步骤
4. **环境变量**：使用 GitHub Secrets 存储敏感信息
5. **缓存优化**：工作流已配置 npm 和 Docker 缓存，可加快构建速度