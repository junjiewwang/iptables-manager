#!/bin/bash

# tun0 Destination流量测试脚本
# 用于验证tun0接口是否正确转发到自定义网桥
# 基于接口信息: tun0 -> destination 192.168.252.2

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置参数
TUN_INTERFACE="tun0"
TUN_LOCAL_IP="192.168.252.1"
TUN_DESTINATION="192.168.252.2"
TUN_NETMASK="255.255.255.255"

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

# 检查权限
check_permissions() {
    if [[ $EUID -ne 0 ]]; then
        log_error "此脚本需要root权限运行"
        echo "请使用: sudo $0"
        exit 1
    fi
}

# 检查依赖工具
check_dependencies() {
    local deps=("ip" "iptables" "tcpdump" "hping3" "netcat" "ping")
    local missing=()
    
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            missing+=("$dep")
        fi
    done
    
    if [[ ${#missing[@]} -gt 0 ]]; then
        log_error "缺少依赖工具: ${missing[*]}"
        log_info "请安装缺少的工具:"
        log_info "Ubuntu/Debian: apt-get install ${missing[*]}"
        log_info "CentOS/RHEL: yum install ${missing[*]}"
        exit 1
    fi
}

# 检查tun0接口状态
check_tun0_interface() {
    log_info "检查tun0接口状态..."
    
    if ! ip link show "$TUN_INTERFACE" &> /dev/null; then
        log_error "tun0接口不存在"
        return 1
    fi
    
    # 获取接口详细信息
    local interface_info=$(ip addr show "$TUN_INTERFACE")
    echo "$interface_info"
    
    # 验证IP配置
    if echo "$interface_info" | grep -q "$TUN_LOCAL_IP"; then
        log_success "tun0本地IP配置正确: $TUN_LOCAL_IP"
    else
        log_warning "tun0本地IP配置可能不匹配"
    fi
    
    # 检查接口状态
    if echo "$interface_info" | grep -q "state UP"; then
        log_success "tun0接口状态: UP"
    else
        log_warning "tun0接口状态异常"
    fi
}

# 获取Docker自定义网桥
get_docker_bridges() {
    log_info "获取Docker自定义网桥..."
    
    local bridges=($(ip link show | grep -E "br-[a-f0-9]{12}" | awk -F': ' '{print $2}' | awk '{print $1}'))
    
    if [[ ${#bridges[@]} -eq 0 ]]; then
        log_warning "未找到Docker自定义网桥"
        return 1
    fi
    
    echo "找到的Docker自定义网桥:"
    for bridge in "${bridges[@]}"; do
        local bridge_info=$(ip addr show "$bridge" | grep "inet " | awk '{print $2}')
        echo "  - $bridge: $bridge_info"
    done
    
    # 返回第一个网桥作为测试目标
    echo "${bridges[0]}"
}

# 设置iptables规则监控
setup_iptables_monitoring() {
    local bridge="$1"
    log_info "设置iptables规则监控 (tun0 -> $bridge)..."
    
    # 清理可能存在的测试规则
    iptables -D FORWARD -i "$TUN_INTERFACE" -o "$bridge" -j LOG --log-prefix "TUN0_TO_BRIDGE: " 2>/dev/null || true
    iptables -D FORWARD -i "$bridge" -o "$TUN_INTERFACE" -j LOG --log-prefix "BRIDGE_TO_TUN0: " 2>/dev/null || true
    
    # 添加日志规则
    iptables -I FORWARD 1 -i "$TUN_INTERFACE" -o "$bridge" -j LOG --log-prefix "TUN0_TO_BRIDGE: " --log-level 4
    iptables -I FORWARD 1 -i "$bridge" -o "$TUN_INTERFACE" -j LOG --log-prefix "BRIDGE_TO_TUN0: " --log-level 4
    
    log_success "iptables监控规则已设置"
}

# 清理iptables监控规则
cleanup_iptables_monitoring() {
    local bridge="$1"
    log_info "清理iptables监控规则..."
    
    iptables -D FORWARD -i "$TUN_INTERFACE" -o "$bridge" -j LOG --log-prefix "TUN0_TO_BRIDGE: " 2>/dev/null || true
    iptables -D FORWARD -i "$bridge" -o "$TUN_INTERFACE" -j LOG --log-prefix "BRIDGE_TO_TUN0: " 2>/dev/null || true
    
    log_success "iptables监控规则已清理"
}

# 启动数据包捕获
start_packet_capture() {
    local bridge="$1"
    local capture_file="/tmp/tun0_bridge_capture_$(date +%s).pcap"
    
    log_info "启动数据包捕获..."
    log_info "捕获文件: $capture_file"
    
    # 在后台启动tcpdump
    tcpdump -i "$TUN_INTERFACE" -w "${capture_file}_tun0" -s 0 &
    local tcpdump_tun_pid=$!
    
    tcpdump -i "$bridge" -w "${capture_file}_bridge" -s 0 &
    local tcpdump_bridge_pid=$!
    
    echo "$tcpdump_tun_pid $tcpdump_bridge_pid $capture_file"
}

# 停止数据包捕获
stop_packet_capture() {
    local pids="$1"
    local capture_file="$2"
    
    log_info "停止数据包捕获..."
    
    # 停止tcpdump进程
    for pid in $pids; do
        kill "$pid" 2>/dev/null || true
    done
    
    sleep 2
    
    # 显示捕获统计
    if [[ -f "${capture_file}_tun0" ]]; then
        local tun0_packets=$(tcpdump -r "${capture_file}_tun0" 2>/dev/null | wc -l)
        log_info "tun0捕获的数据包: $tun0_packets"
    fi
    
    if [[ -f "${capture_file}_bridge" ]]; then
        local bridge_packets=$(tcpdump -r "${capture_file}_bridge" 2>/dev/null | wc -l)
        log_info "网桥捕获的数据包: $bridge_packets"
    fi
}

# 模拟destination流量 - 方法1: 使用hping3
simulate_destination_traffic_hping3() {
    local bridge="$1"
    local bridge_ip="$2"
    
    log_info "方法1: 使用hping3模拟destination流量..."
    
    # 模拟从tun0 destination发送到网桥的流量
    log_info "发送ICMP包到网桥IP: $bridge_ip"
    hping3 -c 5 -I "$TUN_INTERFACE" -a "$TUN_DESTINATION" "$bridge_ip" &
    local hping_pid=$!
    
    sleep 3
    kill "$hping_pid" 2>/dev/null || true
    
    log_info "hping3测试完成"
}

# 模拟destination流量 - 方法2: 使用netcat
simulate_destination_traffic_netcat() {
    local bridge="$1"
    local bridge_ip="$2"
    
    log_info "方法2: 使用netcat模拟TCP连接..."
    
    # 在网桥上启动临时服务器
    local test_port=12345
    timeout 10 nc -l -s "$bridge_ip" -p "$test_port" &
    local nc_server_pid=$!
    
    sleep 1
    
    # 从tun0尝试连接
    echo "test data from tun0 destination" | timeout 5 nc -s "$TUN_DESTINATION" "$bridge_ip" "$test_port" &
    local nc_client_pid=$!
    
    sleep 3
    
    # 清理进程
    kill "$nc_server_pid" "$nc_client_pid" 2>/dev/null || true
    
    log_info "netcat测试完成"
}

# 模拟destination流量 - 方法3: 使用自定义路由
simulate_destination_traffic_route() {
    local bridge="$1"
    local bridge_ip="$2"
    
    log_info "方法3: 使用路由表操作模拟流量..."
    
    # 备份原始路由
    local original_route=$(ip route get "$bridge_ip" 2>/dev/null || echo "")
    
    # 添加临时路由，强制通过tun0
    ip route add "$bridge_ip/32" dev "$TUN_INTERFACE" src "$TUN_DESTINATION" 2>/dev/null || true
    
    sleep 1
    
    # 发送ping测试
    ping -c 3 -I "$TUN_INTERFACE" "$bridge_ip" &
    local ping_pid=$!
    
    sleep 5
    kill "$ping_pid" 2>/dev/null || true
    
    # 恢复路由
    ip route del "$bridge_ip/32" dev "$TUN_INTERFACE" 2>/dev/null || true
    
    log_info "路由测试完成"
}

# 检查转发规则
check_forwarding_rules() {
    local bridge="$1"
    
    log_info "检查当前转发规则..."
    
    echo "=== FORWARD链规则 ==="
    iptables -L FORWARD -n -v --line-numbers | grep -E "(tun0|$bridge)" || echo "未找到相关规则"
    
    echo ""
    echo "=== NAT表规则 ==="
    iptables -t nat -L -n -v --line-numbers | grep -E "(tun0|$bridge)" || echo "未找到相关规则"
    
    echo ""
    echo "=== 内核转发状态 ==="
    local ip_forward=$(cat /proc/sys/net/ipv4/ip_forward)
    if [[ "$ip_forward" == "1" ]]; then
        log_success "IP转发已启用"
    else
        log_warning "IP转发未启用，可能影响转发功能"
        log_info "启用IP转发: echo 1 > /proc/sys/net/ipv4/ip_forward"
    fi
}

# 分析测试结果
analyze_test_results() {
    local bridge="$1"
    
    log_info "分析测试结果..."
    
    # 检查内核日志中的转发记录
    echo "=== 内核日志中的转发记录 ==="
    dmesg | tail -50 | grep -E "(TUN0_TO_BRIDGE|BRIDGE_TO_TUN0)" || echo "未找到转发日志"
    
    echo ""
    echo "=== 接口统计信息 ==="
    echo "tun0统计:"
    cat /proc/net/dev | grep "$TUN_INTERFACE" || echo "未找到tun0统计"
    
    echo ""
    echo "$bridge统计:"
    cat /proc/net/dev | grep "$bridge" || echo "未找到网桥统计"
    
    # 检查连接跟踪
    echo ""
    echo "=== 连接跟踪信息 ==="
    if command -v conntrack &> /dev/null; then
        conntrack -L | grep -E "($TUN_DESTINATION|$TUN_LOCAL_IP)" | head -10 || echo "未找到相关连接"
    else
        echo "conntrack工具未安装"
    fi
}

# 生成测试报告
generate_test_report() {
    local bridge="$1"
    local bridge_ip="$2"
    local start_time="$3"
    local end_time="$4"
    
    local report_file="/tmp/tun0_destination_test_report_$(date +%s).txt"
    
    log_info "生成测试报告: $report_file"
    
    cat > "$report_file" << EOF
tun0 Destination流量转发测试报告
=====================================

测试时间: $(date)
测试持续时间: $((end_time - start_time))秒

接口配置:
- tun0接口: $TUN_INTERFACE
- 本地IP: $TUN_LOCAL_IP
- Destination IP: $TUN_DESTINATION
- 目标网桥: $bridge
- 网桥IP: $bridge_ip

测试方法:
1. hping3模拟ICMP流量
2. netcat模拟TCP连接
3. 路由表操作测试

转发规则检查:
$(iptables -L FORWARD -n -v | grep -E "(tun0|$bridge)" || echo "未找到FORWARD规则")

NAT规则检查:
$(iptables -t nat -L -n -v | grep -E "(tun0|$bridge)" || echo "未找到NAT规则")

内核转发状态:
IP转发: $(cat /proc/sys/net/ipv4/ip_forward)

接口统计:
tun0: $(cat /proc/net/dev | grep tun0 || echo "未找到")
$bridge: $(cat /proc/net/dev | grep $bridge || echo "未找到")

建议:
1. 如果未看到转发日志，检查iptables FORWARD规则
2. 确保IP转发已启用
3. 检查网桥配置是否正确
4. 验证路由表配置

EOF

    log_success "测试报告已生成: $report_file"
}

# 主测试函数
main() {
    local start_time=$(date +%s)
    
    log_info "开始tun0 destination流量转发测试..."
    
    # 检查环境
    check_permissions
    check_dependencies
    check_tun0_interface
    
    # 获取目标网桥
    local bridge=$(get_docker_bridges)
    if [[ -z "$bridge" ]]; then
        log_error "未找到Docker自定义网桥，无法进行测试"
        exit 1
    fi
    
    # 获取网桥IP
    local bridge_ip=$(ip addr show "$bridge" | grep "inet " | awk '{print $2}' | cut -d'/' -f1)
    if [[ -z "$bridge_ip" ]]; then
        log_error "无法获取网桥IP地址"
        exit 1
    fi
    
    log_info "测试目标: $bridge ($bridge_ip)"
    
    # 设置监控
    setup_iptables_monitoring "$bridge"
    
    # 启动数据包捕获
    local capture_info=$(start_packet_capture "$bridge")
    local capture_pids=$(echo "$capture_info" | awk '{print $1 " " $2}')
    local capture_file=$(echo "$capture_info" | awk '{print $3}')
    
    # 检查当前规则
    check_forwarding_rules "$bridge"
    
    echo ""
    log_info "开始流量模拟测试..."
    
    # 执行各种测试方法
    simulate_destination_traffic_hping3 "$bridge" "$bridge_ip"
    sleep 2
    
    simulate_destination_traffic_netcat "$bridge" "$bridge_ip"
    sleep 2
    
    simulate_destination_traffic_route "$bridge" "$bridge_ip"
    sleep 2
    
    # 停止捕获
    stop_packet_capture "$capture_pids" "$capture_file"
    
    # 分析结果
    analyze_test_results "$bridge"
    
    # 清理监控规则
    cleanup_iptables_monitoring "$bridge"
    
    local end_time=$(date +%s)
    
    # 生成报告
    generate_test_report "$bridge" "$bridge_ip" "$start_time" "$end_time"
    
    log_success "tun0 destination流量转发测试完成!"
    log_info "请检查生成的报告和捕获文件以分析转发情况"
}

# 清理函数
cleanup() {
    log_info "执行清理操作..."
    
    # 清理可能残留的iptables规则
    local bridges=($(ip link show | grep -E "br-[a-f0-9]{12}" | awk -F': ' '{print $2}' | awk '{print $1}'))
    for bridge in "${bridges[@]}"; do
        cleanup_iptables_monitoring "$bridge" 2>/dev/null || true
    done
    
    # 清理可能残留的进程
    pkill -f "tcpdump.*tun0" 2>/dev/null || true
    pkill -f "hping3" 2>/dev/null || true
    pkill -f "nc.*12345" 2>/dev/null || true
}

# 设置信号处理
trap cleanup EXIT INT TERM

# 显示帮助信息
show_help() {
    cat << EOF
tun0 Destination流量转发测试脚本

用法: $0 [选项]

选项:
  -h, --help     显示此帮助信息
  -v, --verbose  详细输出模式
  -d, --debug    调试模式

描述:
  此脚本用于测试tun0接口的destination流量(192.168.252.2)
  是否正确转发到Docker自定义网桥。

测试方法:
  1. 使用hping3模拟ICMP流量
  2. 使用netcat模拟TCP连接
  3. 使用路由表操作测试转发

注意:
  - 需要root权限运行
  - 会临时添加iptables日志规则
  - 会启动数据包捕获
  - 测试完成后会自动清理

EOF
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -v|--verbose)
            set -x
            shift
            ;;
        -d|--debug)
            set -x
            DEBUG=1
            shift
            ;;
        *)
            log_error "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
done

# 运行主函数
main "$@"