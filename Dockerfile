# 多阶段构建 Dockerfile
# 第一阶段：前端构建
FROM node:18-alpine AS frontend-builder

WORKDIR /app/frontend

# 复制前端源码中不频繁变动的文件（如package.json）
COPY frontend/package*.json ./
RUN npm install

# 复制其余前端源码并构建
COPY frontend/ ./
RUN npm run build

# 第二阶段：后端构建
FROM golang:1.21-bullseye AS backend-builder

# 安装构建依赖并清理缓存
RUN apt-get update && apt-get install -y gcc libc6-dev libsqlite3-dev && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# 复制后端源码
COPY backend/ ./backend/

# 构建后端应用
WORKDIR /app/backend
RUN go mod tidy && CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o /app/iptables-backend .

# 第四阶段：运行阶段
FROM debian:bullseye-slim

# 安装运行时依赖并清理缓存
RUN apt-get update && apt-get install -y ca-certificates iptables sqlite3 iproute2 net-tools iputils-ping hping3 && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# 从后端构建阶段复制二进制文件
COPY --from=backend-builder /app/iptables-backend ./main

# 从前端构建阶段复制构建产物
COPY --from=frontend-builder /app/frontend/dist/ ./dist/

# 复制scripts目录
COPY scripts/ ./scripts/

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./main"]