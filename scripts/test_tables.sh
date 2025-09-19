#!/bin/bash

# IPTables 表管理测试脚本
echo "=== IPTables 表管理系统测试 ==="

# 检查iptables命令是否可用
if ! command -v iptables &> /dev/null; then
    echo "❌ iptables 命令未找到，请确保已安装 iptables"
    exit 1
fi

echo "✅ iptables 命令可用"

# 测试各个表的命令
echo ""
echo "=== 测试 iptables 命令 ==="

echo "1. 测试 raw 表："
sudo iptables -t raw -L -n --line-numbers | head -10

echo ""
echo "2. 测试 mangle 表："
sudo iptables -t mangle -L -n --line-numbers | head -10

echo ""
echo "3. 测试 nat 表："
sudo iptables -t nat -L -n --line-numbers | head -10

echo ""
echo "4. 测试 filter 表："
sudo iptables -t filter -L -n --line-numbers | head -10

echo ""
echo "5. 测试 FORWARD 链详细信息："
iptables -L FORWARD -v | head -10

echo ""
echo "6. 测试 NAT POSTROUTING 链详细信息："
iptables -t nat -L POSTROUTING -v | head -10

echo ""
echo "7. 测试 DOCKER 相关链（如果存在）："
if iptables -L DOCKER-ISOLATION-STAGE-2 -v &> /dev/null; then
    echo "✅ DOCKER-ISOLATION-STAGE-2 链存在"
    iptables -L DOCKER-ISOLATION-STAGE-2 -v | head -5
else
    echo "ℹ️  DOCKER-ISOLATION-STAGE-2 链不存在（正常，如果未安装Docker）"
fi

echo ""
echo "=== 测试后端API（需要后端服务运行） ==="

# 检查后端服务是否运行
if curl -s http://localhost:8080/api/tables &> /dev/null; then
    echo "✅ 后端服务正在运行"
    
    echo "测试获取所有表信息："
    curl -s -H "Authorization: Bearer test-token" http://localhost:8080/api/tables | jq '.[0].table_name' 2>/dev/null || echo "需要登录获取token"
    
else
    echo "❌ 后端服务未运行，请先启动后端服务"
fi

echo ""
echo "=== 测试完成 ==="
echo "如果所有测试通过，表管理功能应该可以正常工作"