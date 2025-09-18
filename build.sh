#!/bin/bash

echo "开始构建IPTables管理系统..."

# 检查Node.js是否安装
if ! command -v node &> /dev/null; then
    echo "错误: Node.js 未安装，请先安装Node.js"
    exit 1
fi

# 检查Go是否安装
if ! command -v go &> /dev/null; then
    echo "错误: Go 未安装，请先安装Go"
    exit 1
fi

echo "1. 安装前端依赖..."
npm install

if [ $? -ne 0 ]; then
    echo "错误: 前端依赖安装失败"
    exit 1
fi

echo "2. 构建前端..."
npm run build

if [ $? -ne 0 ]; then
    echo "错误: 前端构建失败"
    exit 1
fi

echo "3. 下载Go依赖..."
go mod tidy

if [ $? -ne 0 ]; then
    echo "错误: Go依赖下载失败"
    exit 1
fi

echo "4. 构建Go后端..."
go build -o iptables-management .

if [ $? -ne 0 ]; then
    echo "错误: Go后端构建失败"
    exit 1
fi

echo "5. 设置执行权限..."
chmod +x iptables-management

echo "构建完成！"
echo ""
echo "运行方式："
echo "1. 直接运行: ./iptables-management"
echo "2. Docker构建: docker build -f Dockerfile.go -t iptables-management ."
echo "3. Docker运行: docker run -p 8080:8080 iptables-management"
echo ""
echo "访问地址: http://localhost:8080"