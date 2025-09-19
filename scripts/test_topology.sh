#!/bin/bash

# IPTables 拓扑图测试脚本
# 用于测试拓扑图功能和API接口

echo "=== IPTables 拓扑图测试脚本 ==="
echo "测试时间: $(date)"
echo

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 测试函数
test_api() {
    local url=$1
    local description=$2
    
    echo -e "${BLUE}测试: ${description}${NC}"
    echo "URL: $url"
    
    response=$(curl -s -w "\n%{http_code}" "$url" \
        -H "Authorization: Bearer your-token-here" \
        -H "Content-Type: application/json")
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "200" ]; then
        echo -e "${GREEN}✓ 成功 (HTTP $http_code)${NC}"
        echo "响应数据长度: $(echo "$body" | wc -c) 字符"
        
        # 尝试解析JSON并显示基本信息
        if command -v jq >/dev/null 2>&1; then
            nodes_count=$(echo "$body" | jq -r '.data.nodes | length' 2>/dev/null || echo "N/A")
            links_count=$(echo "$body" | jq -r '.data.links | length' 2>/dev/null || echo "N/A")
            flows_count=$(echo "$body" | jq -r '.data.flow | length' 2>/dev/null || echo "N/A")
            
            echo "  - 节点数量: $nodes_count"
            echo "  - 连接数量: $links_count"
            echo "  - 流路径数量: $flows_count"
        fi
    else
        echo -e "${RED}✗ 失败 (HTTP $http_code)${NC}"
        echo "错误响应: $body"
    fi
    echo
}

# 检查依赖
echo -e "${YELLOW}检查系统依赖...${NC}"

if ! command -v curl >/dev/null 2>&1; then
    echo -e "${RED}错误: curl 未安装${NC}"
    exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
    echo -e "${YELLOW}警告: jq 未安装，将无法解析JSON响应${NC}"
fi

echo -e "${GREEN}✓ 依赖检查完成${NC}"
echo

# 检查后端服务
echo -e "${YELLOW}检查后端服务状态...${NC}"
if curl -s http://localhost:8080/health >/dev/null 2>&1; then
    echo -e "${GREEN}✓ 后端服务运行正常${NC}"
else
    echo -e "${RED}✗ 后端服务未运行或无法访问${NC}"
    echo "请确保后端服务在 http://localhost:8080 运行"
    exit 1
fi
echo

# 测试基础API
echo -e "${YELLOW}测试基础API接口...${NC}"
test_api "http://localhost:8080/api/tables" "获取所有表信息"

# 测试拓扑图API
echo -e "${YELLOW}测试拓扑图API接口...${NC}"
test_api "http://localhost:8080/api/topology" "获取拓扑图数据"

# 测试iptables命令
echo -e "${YELLOW}测试iptables命令...${NC}"

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

# 测试特殊链
echo -e "${YELLOW}测试特殊链...${NC}"

special_chains=("FORWARD" "POSTROUTING" "DOCKER-ISOLATION-STAGE-2")
for chain in "${special_chains[@]}"; do
    echo -e "${BLUE}测试链: $chain${NC}"
    
    # 尝试不同的表
    found=false
    for table in "${tables[@]}"; do
        if iptables -t "$table" -L "$chain" -v >/dev/null 2>&1; then
            echo -e "${GREEN}✓ 链 $chain 在表 $table 中找到${NC}"
            
            # 获取规则数量
            rule_count=$(iptables -t "$table" -L "$chain" -v 2>/dev/null | grep -E "^[[:space:]]*[0-9]+" | wc -l)
            echo "  - 规则数量: $rule_count"
            found=true
            break
        fi
    done
    
    if [ "$found" = false ]; then
        echo -e "${YELLOW}⚠ 链 $chain 未找到或无法访问${NC}"
    fi
    echo
done

# 生成测试报告
echo -e "${YELLOW}生成测试报告...${NC}"

report_file="topology_test_report_$(date +%Y%m%d_%H%M%S).txt"
{
    echo "IPTables 拓扑图测试报告"
    echo "========================"
    echo "测试时间: $(date)"
    echo
    
    echo "系统信息:"
    echo "- 操作系统: $(uname -s)"
    echo "- 内核版本: $(uname -r)"
    echo "- IPTables版本: $(iptables --version 2>/dev/null || echo '未知')"
    echo
    
    echo "表统计:"
    for table in "${tables[@]}"; do
        if iptables -t "$table" -L -n >/dev/null 2>&1; then
            chain_count=$(iptables -t "$table" -L -n 2>/dev/null | grep "^Chain " | wc -l)
            rule_count=$(iptables -t "$table" -L -n --line-numbers 2>/dev/null | grep -E "^[0-9]+" | wc -l)
            echo "- 表 $table: $chain_count 链, $rule_count 规则"
        else
            echo "- 表 $table: 无法访问"
        fi
    done
    echo
    
    echo "API测试结果:"
    echo "- 健康检查: $(curl -s http://localhost:8080/health >/dev/null 2>&1 && echo '通过' || echo '失败')"
    echo "- 表API: $(curl -s http://localhost:8080/api/tables >/dev/null 2>&1 && echo '通过' || echo '失败')"
    echo "- 拓扑API: $(curl -s http://localhost:8080/api/topology >/dev/null 2>&1 && echo '通过' || echo '失败')"
    
} > "$report_file"

echo -e "${GREEN}✓ 测试报告已保存到: $report_file${NC}"
echo

# 提供使用建议
echo -e "${YELLOW}使用建议:${NC}"
echo "1. 确保以root权限运行后端服务，以便访问iptables"
echo "2. 在Docker容器中运行时，确保容器有足够的权限"
echo "3. 如果某些表或链无法访问，检查系统的iptables配置"
echo "4. 前端访问地址: http://localhost:8080"
echo "5. 拓扑图页面: http://localhost:8080/topology"
echo

echo -e "${GREEN}测试完成！${NC}"