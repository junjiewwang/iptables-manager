# 构建阶段 - 前端
FROM node:18-alpine AS frontend-builder

WORKDIR /app

# 复制前端文件
COPY package*.json ./
COPY tsconfig*.json ./
COPY vite.config.ts ./
COPY index.html ./
COPY src/ ./src/

# 安装依赖并构建
RUN npm install
RUN npm run build

# 构建阶段 - 后端
FROM golang:1.21-alpine AS backend-builder

WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git

# 复制Go模块文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY *.go ./

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# 运行阶段
FROM alpine:latest

# 安装必要的包
RUN apk --no-cache add ca-certificates iptables

WORKDIR /root/

# 从构建阶段复制文件
COPY --from=backend-builder /app/main .
COPY --from=frontend-builder /app/dist ./dist
COPY .env .

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./main"]