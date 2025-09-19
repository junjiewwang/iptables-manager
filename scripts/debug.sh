#!/bin/bash

echo "=== IPTables管理系统调试脚本 ==="
echo "当前时间: $(date)"
echo ""

echo "1. 检查Docker容器状态..."
docker-compose ps

echo ""
echo "2. 检查后端日志..."
echo "最近的后端日志:"
docker-compose logs --tail=20 backend

echo ""
echo "3. 检查数据库文件..."
if [ -f "./data/iptables.db" ]; then
    echo "数据库文件存在: ./data/iptables.db"
    echo "文件大小: $(ls -lh ./data/iptables.db | awk '{print $5}')"
else
    echo "数据库文件不存在: ./data/iptables.db"
fi

echo ""
echo "4. 测试API连接..."
echo "测试健康检查API:"
curl -s http://localhost:8080/health || echo "API连接失败"

echo ""
echo "5. 检查实际的iptables规则..."
echo "当前系统iptables规则:"
iptables -L -n --line-numbers 2>/dev/null || echo "无法获取iptables规则 (需要root权限)"

echo ""
echo "=== 调试脚本完成 ==="