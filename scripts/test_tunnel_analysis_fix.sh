#!/bin/bash

# 测试隧道分析功能修复效果
echo "=== 隧道分析功能修复测试 ==="

# 设置API基础URL
API_BASE="http://localhost:8080/api"
TOKEN="test-token"

echo "1. 测试获取隧道接口列表..."
curl -s -H "Authorization: Bearer $TOKEN" "$API_BASE/tunnel/interfaces" | jq '.' || echo "隧道接口获取失败"

echo -e "\n2. 测试获取Docker网桥列表..."
curl -s -H "Authorization: Bearer $TOKEN" "$API_BASE/tunnel/docker-bridges" | jq '.' || echo "Docker网桥获取失败"

echo -e "\n3. 测试隧道接口信息获取（假设tun0存在）..."
curl -s -H "Authorization: Bearer $TOKEN" "$API_BASE/tunnel/tun0/info" | jq '.' || echo "隧道接口信息获取失败"

echo -e "\n4. 测试隧道接口规则获取..."
curl -s -H "Authorization: Bearer $TOKEN" "$API_BASE/tunnel/tun0/rules" | jq '.' || echo "隧道接口规则获取失败"

echo -e "\n5. 测试通信路径分析（tun0 -> docker0）..."
curl -s -H "Authorization: Bearer $TOKEN" "$API_BASE/tunnel/analyze-communication?tunnel_interface=tun0&docker_bridge=docker0" | jq '.' || echo "通信路径分析失败"

echo -e "\n6. 测试通信路径分析（tun0 -> br-xxx）..."
# 首先获取可用的网桥
BRIDGES=$(curl -s -H "Authorization: Bearer $TOKEN" "$API_BASE/tunnel/docker-bridges" | jq -r '.docker_bridges[]?.name' | grep "^br-" | head -1)
if [ -n "$BRIDGES" ]; then
    echo "使用网桥: $BRIDGES"
    curl -s -H "Authorization: Bearer $TOKEN" "$API_BASE/tunnel/analyze-communication?tunnel_interface=tun0&docker_bridge=$BRIDGES" | jq '.'
else
    echo "未找到br-前缀的网桥，跳过测试"
fi

echo -e "\n7. 测试规则生成..."
curl -s -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d '{
        "tunnel_interface": "tun0",
        "docker_bridge": "docker0",
        "direction": "bidirectional",
        "protocol": "all",
        "action": "ACCEPT",
        "enable_nat": true,
        "enable_logging": false
    }' \
    "$API_BASE/tunnel/generate-rules" | jq '.' || echo "规则生成失败"

echo -e "\n=== 直接测试网络连通性 ==="

echo -e "\n8. 检查tun0接口是否存在..."
ip link show tun0 2>/dev/null && echo "tun0接口存在" || echo "tun0接口不存在"

echo -e "\n9. 检查docker0网桥是否存在..."
ip link show docker0 2>/dev/null && echo "docker0网桥存在" || echo "docker0网桥不存在"

echo -e "\n10. 检查相关iptables规则..."
echo "FORWARD链规则（包含tun0）:"
iptables -t filter -L FORWARD -n -v --line-numbers | grep -i tun || echo "未找到tun相关的FORWARD规则"

echo -e "\nNAT POSTROUTING规则（包含tun0）:"
iptables -t nat -L POSTROUTING -n -v --line-numbers | grep -i tun || echo "未找到tun相关的NAT规则"

echo -e "\n11. 测试实际连通性（如果接口存在）..."
if ip link show tun0 >/dev/null 2>&1 && ip link show docker0 >/dev/null 2>&1; then
    TUN_IP=$(ip addr show tun0 | grep 'inet ' | awk '{print $2}' | cut -d'/' -f1 | head -1)
    DOCKER_IP=$(ip addr show docker0 | grep 'inet ' | awk '{print $2}' | cut -d'/' -f1 | head -1)
    
    if [ -n "$TUN_IP" ] && [ -n "$DOCKER_IP" ]; then
        echo "尝试从tun0($TUN_IP)ping docker0($DOCKER_IP)..."
        ping -c 1 -W 2 -I tun0 $DOCKER_IP && echo "连通性测试成功" || echo "连通性测试失败"
    else
        echo "无法获取接口IP地址"
    fi
else
    echo "接口不存在，跳过连通性测试"
fi

echo -e "\n=== 测试完成 ==="
echo "请检查以上输出，验证修复效果："
echo "1. API调用是否返回正确的数据结构"
echo "2. 统计信息是否不再全为0"
echo "3. 规则筛选是否更加精确"
echo "4. 优化建议是否基于实际情况"
echo "5. 通信路径分析是否反映真实状态"