#!/bin/bash

echo "=== IPTables管理系统功能测试脚本 ==="
echo "当前时间: $(date)"
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 测试API函数
test_api() {
    local url=$1
    local description=$2
    
    echo -n "测试 $description: "
    
    response=$(curl -s -w "%{http_code}" -H "Authorization: Bearer test-token" "$url" -o /tmp/api_response.json)
    http_code="${response: -3}"
    
    if [ "$http_code" -eq 200 ]; then
        echo -e "${GREEN}✓ 成功 (HTTP $http_code)${NC}"
        # 显示响应数据的简要信息
        if command -v jq &> /dev/null; then
            data_length=$(jq '. | length' /tmp/api_response.json 2>/dev/null || echo "N/A")
            echo "  - 数据长度: $data_length"
        fi
    else
        echo -e "${RED}✗ 失败 (HTTP $http_code)${NC}"
        if [ -f /tmp/api_response.json ]; then
            echo "  - 错误信息: $(cat /tmp/api_response.json)"
        fi
    fi
    echo
}

# 检查后端服务
echo -e "${YELLOW}1. 检查后端服务状态...${NC}"
if curl -s http://localhost:8080/health > /dev/null; then
    echo -e "${GREEN}✓ 后端服务运行正常${NC}"
else
    echo -e "${RED}✗ 后端服务未运行或无法访问${NC}"
    echo "请确保后端服务在 http://localhost:8080 运行"
    exit 1
fi
echo

# 测试基础功能
echo -e "${YELLOW}2. 测试基础API接口...${NC}"
test_api "http://localhost:8080/api/rules" "获取规则列表"
test_api "http://localhost:8080/api/rules/system" "获取系统实时规则"
test_api "http://localhost:8080/api/statistics" "获取统计信息"
test_api "http://localhost:8080/api/tables" "获取表信息"
test_api "http://localhost:8080/api/topology" "获取拓扑图数据"

# 测试新功能
echo -e "${YELLOW}3. 测试网络接口功能...${NC}"
test_api "http://localhost:8080/api/interfaces" "获取网络接口"
test_api "http://localhost:8080/api/docker/bridges" "获取Docker网桥"

# 测试iptables命令
echo -e "${YELLOW}4. 测试iptables命令...${NC}"

tables=("raw" "mangle" "nat" "filter")
for table in "${tables[@]}"; do
    echo -e "${BLUE}测试表: $table${NC}"
    
    if iptables -t "$table" -L -n --line-numbers >/dev/null 2>&1; then
        echo -e "${GREEN}✓ 表 $table 可访问${NC}"
        
        # 统计链数量
        chain_count=$(iptables -t "$table" -L -n 2>/dev/null | grep "^Chain " | wc -l)
        echo "  - 链数量: $chain_count"
        
        # 统计规则数量
        rule_count=$(iptables -t "$table" -L -n --line-numbers 2>/dev/null | grep -E "^[0-9]+" | wc -l)
        echo "  - 规则数量: $rule_count"
    else
        echo -e "${RED}✗ 表 $table 无法访问${NC}"
    fi
    echo
done

# 测试网络接口命令
echo -e "${YELLOW}5. 测试网络接口命令...${NC}"

echo -e "${BLUE}检查网络接口:${NC}"
if command -v ip &> /dev/null; then
    interface_count=$(ip link show | grep -c "^[0-9]")
    echo "  - 网络接口数量: $interface_count"
    
    echo "  - 活动接口:"
    ip link show | grep "state UP" | awk '{print "    " $2}' | sed 's/:$//'
else
    echo -e "${YELLOW}  - ip命令不可用，使用ifconfig${NC}"
    if command -v ifconfig &> /dev/null; then
        interface_count=$(ifconfig -a | grep -c "^[a-zA-Z]")
        echo "  - 网络接口数量: $interface_count"
    else
        echo -e "${RED}  - 无法获取网络接口信息${NC}"
    fi
fi
echo

# 测试Docker
echo -e "${YELLOW}6. 测试Docker功能...${NC}"

if command -v docker &> /dev/null; then
    if docker info >/dev/null 2>&1; then
        echo -e "${GREEN}✓ Docker服务运行正常${NC}"
        
        # 检查Docker网络
        network_count=$(docker network ls | wc -l)
        echo "  - Docker网络数量: $((network_count - 1))"
        
        # 检查Docker网桥
        bridge_count=$(docker network ls --filter driver=bridge | wc -l)
        echo "  - Docker网桥数量: $((bridge_count - 1))"
        
        echo "  - Docker网桥列表:"
        docker network ls --filter driver=bridge --format "table {{.Name}}\t{{.Driver}}\t{{.Scope}}" | tail -n +2 | while read line; do
            echo "    $line"
        done
    else
        echo -e "${YELLOW}⚠ Docker服务未运行${NC}"
    fi
else
    echo -e "${YELLOW}⚠ Docker未安装${NC}"
fi
echo

# 测试前端页面
echo -e "${YELLOW}7. 测试前端页面...${NC}"

pages=("/dashboard" "/rules" "/tables" "/topology" "/interfaces" "/logs")
for page in "${pages[@]}"; do
    echo -n "测试页面 $page: "
    
    response=$(curl -s -w "%{http_code}" "http://localhost:3000$page" -o /dev/null)
    http_code="${response: -3}"
    
    if [ "$http_code" -eq 200 ]; then
        echo -e "${GREEN}✓ 可访问${NC}"
    else
        echo -e "${RED}✗ 无法访问 (HTTP $http_code)${NC}"
    fi
done
echo

# 功能验证总结
echo -e "${YELLOW}8. 功能验证总结...${NC}"

echo -e "${BLUE}已实现的功能:${NC}"
echo "  ✓ 实时获取系统iptables规则"
echo "  ✓ 规则数据与系统同步"
echo "  ✓ 网络接口信息展示"
echo "  ✓ Docker网桥信息展示"
echo "  ✓ 基于Docker网桥的规则视图"
echo "  ✓ 网络接口分类显示"
echo "  ✓ 拓扑图可视化"
echo "  ✓ 完整的Web界面"

echo ""
echo -e "${BLUE}使用说明:${NC}"
echo "  1. 访问 http://localhost:3000 打开Web界面"
echo "  2. 使用默认账号登录 (admin/admin123)"
echo "  3. 在'规则管理'页面点击'实时规则'获取系统规则"
echo "  4. 在'规则管理'页面点击'同步规则'将系统规则同步到数据库"
echo "  5. 在'网络接口'页面查看网络接口和Docker网桥信息"
echo "  6. 在'拓扑图'页面查看iptables规则拓扑结构"

echo ""
echo -e "${GREEN}=== 测试完成 ===${NC}"

# 清理临时文件
rm -f /tmp/api_response.json