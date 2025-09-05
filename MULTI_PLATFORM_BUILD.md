# 多系统构建指南

本项目配置了GitHub Actions多系统构建工作流，可以在发布标签时自动构建Linux、Windows和macOS版本，并创建GitHub Release。本文档详细介绍如何使用和配置这个自动化工作流。

## 1. 多系统构建工作流

### 1.1 工作流文件

多系统构建工作流定义在 `.github/workflows/multi-platform-build.yml` 文件中。

### 1.2 触发条件

此工作流仅在推送标签（以 `v` 开头）时触发，例如：
- v1.0.0
- v2.1.3
- v0.5.0-beta

### 1.3 构建矩阵

工作流使用GitHub Actions的矩阵策略，同时为以下操作系统构建应用：
- Ubuntu Linux (ubuntu-latest)
- Windows (windows-latest)
- macOS (macos-latest)

### 1.4 构建产物

每个平台的构建产物包括：
- 平台特定的后端可执行文件（使用main_with_frontend.go构建，已集成前端静态文件支持）
- 前端静态文件（dist目录）
- 配置文件（config.yaml）

这些文件会被打包成ZIP文件并上传到GitHub Release。

### 1.5 构建优化

后端构建使用了以下优化：
- 使用main_with_frontend.go作为入口文件，确保后端可以正确加载前端静态文件
- 使用-ldflags="-s -w"减小可执行文件大小（去除调试信息和符号表）

### 1.6 工作流最佳实践

工作流配置采用了以下最佳实践：
- 使用最新的GitHub Actions组件版本（如actions/upload-artifact@v4和actions/download-artifact@v4）
- 使用矩阵构建策略，提高构建效率
- 针对不同操作系统使用条件执行（if条件）
- 使用标准Go版本格式（1.21而非1.21.0）以确保兼容性

## 2. 使用方法

### 2.1 创建发布标签

要触发多系统构建，需要创建并推送一个标签：

```bash
# 创建标签
git tag v1.0.0

# 推送标签到GitHub
git push origin v1.0.0
```

### 2.2 查看构建进度

1. 在GitHub仓库页面，点击 "Actions"
2. 在左侧菜单中选择 "Multi-Platform Build and Release"
3. 点击最新的运行记录查看详细进度

### 2.3 下载构建产物

构建完成后，会自动创建一个GitHub Release：

1. 在GitHub仓库页面，点击 "Releases"
2. 找到对应版本的Release
3. 下载需要的平台版本（ZIP文件）

## 3. 自定义配置

### 3.1 修改构建参数

如果需要自定义构建参数，可以编辑 `.github/workflows/multi-platform-build.yml` 文件：

- 修改Go版本：更改 `go-version` 字段
- 修改Node.js版本：更改 `node-version` 字段
- 添加/移除构建平台：修改 `matrix.os` 数组
- 修改输出文件名：更改 `matrix.include` 中的 `output_name` 和 `asset_name`

### 3.2 添加更多构建步骤

如果需要添加更多构建步骤（如运行测试、代码检查等），可以在 `steps` 部分添加相应的操作。

## 4. 故障排除

### 4.1 构建失败

如果构建失败，请检查：

1. 查看GitHub Actions日志，找出失败原因
2. 确认代码可以在本地正常构建
3. 检查工作流文件语法是否正确

### 4.2 Release创建失败

如果Release创建失败，请检查：

1. 确认GitHub Token权限是否足够
2. 检查是否已存在同名Release

## 5. 最佳实践

1. **语义化版本**：使用语义化版本标签（如v1.0.0），遵循[语义化版本规范](https://semver.org/)
2. **发布说明**：在创建标签前，先准备好Release Notes
3. **预发布**：对于不稳定版本，使用预发布标签（如v1.0.0-beta.1）
4. **测试验证**：在创建正式发布标签前，确保代码已经过充分测试
5. **版本更新**：每次发布新版本时，记得更新应用内的版本号

## 6. 示例：完整发布流程

```bash
# 1. 确保代码已合并到主分支
git checkout main
git pull

# 2. 更新版本号（在相关文件中）
# 编辑版本文件...

# 3. 提交版本更新
git add .
git commit -m "chore: bump version to 1.0.0"
git push

# 4. 创建标签
git tag v1.0.0

# 5. 推送标签
git push origin v1.0.0

# 6. 等待GitHub Actions完成构建
# 7. 在GitHub Releases页面编辑发布说明
```

通过以上步骤，您可以利用GitHub Actions自动构建多系统版本，简化发布流程。