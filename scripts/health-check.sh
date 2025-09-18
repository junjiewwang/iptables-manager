#!/bin/bash

# IPTables 管理系统健康检查脚本

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

# 检查服务状态
check_service() {
    local service_name=$1
    local url=$2
    local expected_status=${3:-200}
    
    log_info "检查 $service_name 服务..."
    
    if curl -s -o /dev/null -w "%{http_code}" "$url" | grep -q "$expected_status"; then
        log_success "$service_name 服务正常"
        return 0
    else
        log_error "$service_name 服务异常"
        return 1
    fi
}

# 检查 Docker 容器状态
check_containers() {
    log_info "检查 Docker 容器状态..."
    
    local containers=("iptables-frontend" "iptables-backend" "iptables-mysql")
    local all_healthy=true
    
    for container in "${containers[@]}"; do
        if docker ps --format "table {{.Names}}\t{{.Status}}" | grep -q "$container.*Up"; then
            log_success "$container 容器运行正常"
        else
            log_error "$container 容器未运行或异常"
            all_healthy=false
        fi
    done
    
    if [ "$all_healthy" = true ]; then
        return 0
    else
        return 1
    fi
}

# 检查数据库连接
check_database() {
    log_info "检查数据库连接..."
    
    if docker-compose exec -T mysql mysqladmin ping -h localhost --silent 2>/dev/null; then
        log_success "数据库连接正常"
        return 0
    else
        log_error "数据库连接失败"
        return 1
    fi
}

# 检查 API 端点
check_api_endpoints() {
    log_info "检查 API 端点..."
    
    local endpoints=(
        "http://localhost:8080/health"
        "http://localhost:8080/api/login"
    )
    
    local all_healthy=true
    
    for endpoint in "${endpoints[@]}"; do
        if curl -s -f "$endpoint" >/dev/null 2>&1 || [ $? -eq 22 ]; then
            log_success "API 端点 $endpoint 可访问"
        else
            log_error "API 端点 $endpoint 不可访问"
            all_healthy=false
        fi
    done
    
    if [ "$all_healthy" = true ]; then
        return 0
    else
        return 1
    fi
}

# 检查前端页面
check_frontend() {
    log_info "检查前端页面..."
    
    if curl -s -f "http://localhost" >/dev/null 2>&1; then
        log_success "前端页面可访问"
        return 0
    else
        log_error "前端页面不可访问"
        return 1
    fi
}

# 生成健康报告
generate_report() {
    local overall_status=$1
    
    echo
    echo "=================================="
    echo "🏥 IPTables 管理系统健康检查报告"
    echo "=================================="
    echo "检查时间: $(date)"
    echo
    
    if [ "$overall_status" = "healthy" ]; then
        echo -e "总体状态: ${GREEN}健康${NC} ✅"
        echo
        echo "📊 服务状态:"
        echo "  - 前端服务: 正常"
        echo "  - 后端服务: 正常"
        echo "  - 数据库服务: 正常"
        echo
        echo "🌐 访问地址:"
        echo "  - 前端: http://localhost"
        echo "  - 后端API: http://localhost:8080"
        echo "  - 健康检查: http://localhost:8080/health"
        echo
        echo "👤 默认账户:"
        echo "  - 管理员: admin / admin123"
        echo "  - 普通用户: user1 / user123"
    else
        echo -e "总体状态: ${RED}异常${NC} ❌"
        echo
        echo "🔧 故障排除建议:"
        echo "  1. 检查 Docker 服务是否运行: docker ps"
        echo "  2. 查看服务日志: docker-compose logs"
        echo "  3. 重启服务: docker-compose restart"
        echo "  4. 检查端口占用: netstat -tulpn | grep -E ':(80|8080|3306)'"
    fi
    
    echo "=================================="
}

# 主函数
main() {
    log_info "开始 IPTables 管理系统健康检查"
    echo
    
    local checks_passed=0
    local total_checks=5
    
    # 执行各项检查
    if check_containers; then
        ((checks_passed++))
    fi
    
    if check_database; then
        ((checks_passed++))
    fi
    
    if check_api_endpoints; then
        ((checks_passed++))
    fi
    
    if check_frontend; then
        ((checks_passed++))
    fi
    
    # 额外的服务检查
    if check_service "后端健康检查" "http://localhost:8080/health"; then
        ((checks_passed++))
    fi
    
    # 生成报告
    if [ $checks_passed -eq $total_checks ]; then
        generate_report "healthy"
        exit 0
    else
        generate_report "unhealthy"
        log_error "健康检查失败: $checks_passed/$total_checks 项检查通过"
        exit 1
    fi
}

# 脚本入口
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi