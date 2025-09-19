#!/bin/bash

# 测试系统原生网络命令功能
echo "=== 测试系统原生网络命令功能 ==="

# 设置API基础URL
API_BASE="http://localhost:8080/api"
TOKEN="test-token"

echo "1. 测试网络接口获取..."
curl -s -H "Authorization: Bearer $TOKEN" "$API_BASE/interfaces" | jq '.[0:3]' || echo "接口获取失败"

echo -e "\n2. 测试Docker网桥获取（使用系统原生命令）..."
curl -s -H "Authorization: Bearer $TOKEN" "$API_BASE/docker/bridges" | jq '.[0:2]' || echo "网桥获取失败"

echo -e "\n3. 测试网络连接获取..."
curl -s -H "Authorization: Bearer $TOKEN" "$API_BASE/network/connections" | jq '.[0:5]' || echo "网络连接获取失败"

echo -e "\n4. 测试路由表获取..."
curl -s -H "Authorization: Bearer $TOKEN" "$API_BASE/network/routes" | jq '.[0:5]' || echo "路由表获取失败"

echo -e "\n5. 测试网桥规则获取..."
curl -s -H "Authorization: Bearer $TOKEN" "$API_BASE/bridges/docker0/rules" | jq '.[0:3]' || echo "网桥规则获取失败"

echo -e "\n=== 直接测试系统命令 ==="

echo -e "\n6. 测试ip命令..."
echo "网络接口:"
ip addr show | head -20

echo -e "\n网络路由:"
ip route show | head -10

echo -e "\n7. 测试netstat命令..."
echo "网络连接:"
netstat -tuln | head -10

echo -e "\n8. 测试iptables命令..."
echo "iptables规则:"
iptables -t filter -L -n --line-numbers | head -20

echo -e "\n=== 测试完成 ==="