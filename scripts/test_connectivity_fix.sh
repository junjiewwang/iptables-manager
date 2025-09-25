#!/bin/bash

# 隧道接口连通性修复功能测试脚本
# 用于验证一键修复功能是否与手动脚本效果一致

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 配置变量
TUNNEL_INTERFACE="tun0"
DOCKER_BRIDGE=""
API_BASE_URL="http://localhost:8080/api"

# 检查必要的工具
check_prerequisites() {
    log_info "检查必要的工具..."
    
    if ! command -v iptables &> /dev/null; then
        log_error "iptables 未安装"
        exit 1
    fi
    
    if ! command -v curl &> /dev/null; then
        log_error "curl 未安装"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        log_warning "jq 未安装，JSON解析可能受限"
    fi
    
    log_success "工具检查完成"
}

# 获取Docker网桥列表
get_docker_bridges() {
    log_info "获取Docker网桥列表..."
    
    local response=$(curl -s "${API_BASE_URL}/tunnel/docker-bridges" || echo "")
    if [ -z "$response" ]; then
        log_error "无法获取Docker网桥列表，请确保后端服务正在运行"
        exit 1
    fi
    
    # 尝试解析JSON并获取第一个bridge类型的网桥
    if command -v jq &> /dev/null; then
        DOCKER_BRIDGE=$(echo "$response" | jq -r '.docker_bridges[] | select(.driver == "bridge") | .name' | head -1)
    else
        # 简单的文本解析
        DOCKER_BRIDGE=$(echo "$response" | grep -o '"name":"[^"]*"' | head -1 | cut -d'"' -f4)
    fi
    
    if [ -z "$DOCKER_BRIDGE" ] || [ "$DOCKER_BRIDGE" = "null" ]; then
        log_error "未找到可用的Docker网桥"
        exit 1
    fi
    
    log_success "找到Docker网桥: $DOCKER_BRIDGE"
}

# 检查接口是否存在
check_interfaces() {
    log_info "检查网络接口..."
    
    if ! ip link show "$TUNNEL_INTERFACE" &> /dev/null; then
        log_error "隧道接口 $TUNNEL_INTERFACE 不存在"
        exit 1
    fi
    
    if ! ip link show "$DOCKER_BRIDGE" &> /dev/null; then
        log_error "Docker网桥 $DOCKER_BRIDGE 不存在"
        exit 1
    fi
    
    log_success "网络接口检查完成"
}

# 清理现有规则
cleanup_rules() {
    log_info "清理现有的相关规则..."
    
    # 清理FORWARD链中的相关规则
    iptables -L FORWARD -n --line-numbers | grep -E "(${TUNNEL_INTERFACE}|${DOCKER_BRIDGE})" | tac | while read line; do
        line_num=$(echo "$line" | awk '{print $1}')
        if [[ "$line_num" =~ ^[0-9]+$ ]]; then
            iptables -D FORWARD "$line_num" 2>/dev/null || true
        fi
    done
    
    # 清理DOCKER-ISOLATION-STAGE-2链中的相关规则
    if iptables -L DOCKER-ISOLATION-STAGE-2 -n &> /dev/null; then
        iptables -L DOCKER-ISOLATION-STAGE-2 -n --line-numbers | grep -E "(${TUNNEL_INTERFACE}|${DOCKER_BRIDGE})" | tac | while read line; do
            line_num=$(echo "$line" | awk '{print $1}')
            if [[ "$line_num" =~ ^[0-9]+$ ]]; then
                iptables -D DOCKER-ISOLATION-STAGE-2 "$line_num" 2>/dev/null || true
            fi
        done
    fi
    
    log_success "规则清理完成"
}

# 记录修复前的规则状态
record_before_state() {
    log_info "记录修复前的规则状态..."
    
    echo "=== FORWARD链规则 (修复前) ===" > /tmp/rules_before.txt
    iptables -L FORWARD -n --line-numbers >> /tmp/rules_before.txt
    
    echo "" >> /tmp/rules_before.txt
    echo "=== DOCKER-ISOLATION-STAGE-2链规则 (修复前) ===" >> /tmp/rules_before.txt
    iptables -L DOCKER-ISOLATION-STAGE-2 -n --line-numbers >> /tmp/rules_before.txt 2>/dev/null || echo "链不存在" >> /tmp/rules_before.txt
    
    log_success "修复前状态已记录到 /tmp/rules_before.txt"
}

# 执行一键修复
execute_fix() {
    log_info "执行一键修复功能..."
    
    local payload="{\"tunnel_interface\":\"${TUNNEL_INTERFACE}\",\"docker_bridge\":\"${DOCKER_BRIDGE}\"}"
    local response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$payload" \
        "${API_BASE_URL}/tunnel/fix-connectivity" || echo "")
    
    if [ -z "$response" ]; then
        log_error "一键修复请求失败"
        return 1
    fi
    
    # 解析响应
    if command -v jq &> /dev/null; then
        local success=$(echo "$response" | jq -r '.fix_result.success // false')
        local fixed_issues=$(echo "$response" | jq -r '.fix_result.fixed_issues[]? // empty')
        local applied_rules=$(echo "$response" | jq -r '.fix_result.applied_rules[]? // empty')
        
        if [ "$success" = "true" ]; then
            log_success "一键修复执行成功"
            
            if [ -n "$fixed_issues" ]; then
                log_info "修复的问题:"
                echo "$fixed_issues" | while read issue; do
                    echo "  - $issue"
                done
            fi
            
            if [ -n "$applied_rules" ]; then
                log_info "应用的规则:"
                echo "$applied_rules" | while read rule; do
                    echo "  - $rule"
                done
            fi
        else
            log_error "一键修复执行失败"
            return 1
        fi
    else
        log_info "一键修复请求已发送，响应: $response"
    fi
    
    return 0
}

# 记录修复后的规则状态
record_after_state() {
    log_info "记录修复后的规则状态..."
    
    echo "=== FORWARD链规则 (修复后) ===" > /tmp/rules_after.txt
    iptables -L FORWARD -n --line-numbers >> /tmp/rules_after.txt
    
    echo "" >> /tmp/rules_after.txt
    echo "=== DOCKER-ISOLATION-STAGE-2链规则 (修复后) ===" >> /tmp/rules_after.txt
    iptables -L DOCKER-ISOLATION-STAGE-2 -n --line-numbers >> /tmp/rules_after.txt 2>/dev/null || echo "链不存在" >> /tmp/rules_after.txt
    
    log_success "修复后状态已记录到 /tmp/rules_after.txt"
}

# 验证规则是否正确添加
verify_rules() {
    log_info "验证规则是否正确添加..."
    
    local errors=0
    
    # 检查FORWARD链中的规则
    log_info "检查FORWARD链规则..."
    
    # 检查tun0 -> bridge规则
    if iptables -L FORWARD -n | grep -q "ACCEPT.*${TUNNEL_INTERFACE}.*${DOCKER_BRIDGE}"; then
        log_success "✓ 找到 ${TUNNEL_INTERFACE} -> ${DOCKER_BRIDGE} FORWARD规则"
    else
        log_error "✗ 未找到 ${TUNNEL_INTERFACE} -> ${DOCKER_BRIDGE} FORWARD规则"
        ((errors++))
    fi
    
    # 检查conntrack规则
    if iptables -L FORWARD -v | grep -q "ctstate RELATED,ESTABLISHED.*${DOCKER_BRIDGE}.*${TUNNEL_INTERFACE}"; then
        log_success "✓ 找到 ${DOCKER_BRIDGE} -> ${TUNNEL_INTERFACE} conntrack规则"
    else
        log_warning "! 未找到标准conntrack规则，检查是否使用state模块"
        if iptables -L FORWARD -v | grep -q "state RELATED,ESTABLISHED.*${DOCKER_BRIDGE}.*${TUNNEL_INTERFACE}"; then
            log_success "✓ 找到 ${DOCKER_BRIDGE} -> ${TUNNEL_INTERFACE} state规则"
        else
            log_error "✗ 未找到任何状态跟踪规则"
            ((errors++))
        fi
    fi
    
    # 检查Docker隔离规则
    if iptables -L DOCKER-ISOLATION-STAGE-2 -n &> /dev/null; then
        log_info "检查Docker隔离规则..."
        if iptables -L DOCKER-ISOLATION-STAGE-2 -n | grep -q "RETURN.*${TUNNEL_INTERFACE}.*${DOCKER_BRIDGE}"; then
            log_success "✓ 找到 ${TUNNEL_INTERFACE} -> ${DOCKER_BRIDGE} 隔离RETURN规则"
        else
            log_error "✗ 未找到 ${TUNNEL_INTERFACE} -> ${DOCKER_BRIDGE} 隔离RETURN规则"
            ((errors++))
        fi
    else
        log_warning "! DOCKER-ISOLATION-STAGE-2链不存在，跳过隔离规则检查"
    fi
    
    # 检查规则位置
    log_info "检查规则位置..."
    local forward_rules=$(iptables -L FORWARD -n --line-numbers | head -5)
    if echo "$forward_rules" | grep -q "${TUNNEL_INTERFACE}.*${DOCKER_BRIDGE}"; then
        log_success "✓ ${TUNNEL_INTERFACE} -> ${DOCKER_BRIDGE} 规则位于FORWARD链前部"
    else
        log_warning "! ${TUNNEL_INTERFACE} -> ${DOCKER_BRIDGE} 规则可能不在FORWARD链前部"
    fi
    
    if [ $errors -eq 0 ]; then
        log_success "规则验证通过"
        return 0
    else
        log_error "规则验证失败，发现 $errors 个问题"
        return 1
    fi
}

# 生成对比报告
generate_report() {
    log_info "生成对比报告..."
    
    cat > /tmp/fix_test_report.txt << EOF
# 隧道接口连通性修复测试报告

## 测试配置
- 隧道接口: $TUNNEL_INTERFACE
- Docker网桥: $DOCKER_BRIDGE
- 测试时间: $(date)

## 修复前规则状态
$(cat /tmp/rules_before.txt)

## 修复后规则状态
$(cat /tmp/rules_after.txt)

## 规则变化对比
$(diff /tmp/rules_before.txt /tmp/rules_after.txt || true)

EOF
    
    log_success "测试报告已生成: /tmp/fix_test_report.txt"
}

# 主函数
main() {
    log_info "开始隧道接口连通性修复功能测试"
    
    check_prerequisites
    get_docker_bridges
    check_interfaces
    
    log_info "使用配置: 隧道接口=$TUNNEL_INTERFACE, Docker网桥=$DOCKER_BRIDGE"
    
    cleanup_rules
    record_before_state
    
    if execute_fix; then
        sleep 2  # 等待规则生效
        record_after_state
        
        if verify_rules; then
            log_success "🎉 一键修复功能测试通过！"
        else
            log_error "❌ 一键修复功能测试失败！"
        fi
    else
        log_error "❌ 一键修复执行失败！"
    fi
    
    generate_report
    
    log_info "测试完成，详细信息请查看:"
    log_info "  - 修复前状态: /tmp/rules_before.txt"
    log_info "  - 修复后状态: /tmp/rules_after.txt"
    log_info "  - 完整报告: /tmp/fix_test_report.txt"
}

# 执行主函数
main "$@"