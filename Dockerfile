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
COPY --from=frontend-builder /app/frontend/dist ./dist
COPY backend/config.yaml ./

# 创建上传目录
RUN mkdir -p uploads

EXPOSE 8080
CMD ["./telegram-photo-server"]