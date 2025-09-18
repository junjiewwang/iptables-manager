#!/bin/bash

# IPTables 管理系统部署脚本
# 使用方法: ./scripts/deploy.sh [dev|prod]

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

# 检查 Docker 和 Docker Compose
check_dependencies() {
    log_info "检查依赖..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    fi
    
    log_success "依赖检查通过"
}

# 环境配置
setup_environment() {
    local env=${1:-dev}
    log_info "设置 $env 环境..."
    
    if [ "$env" = "prod" ]; then
        log_warning "生产环境部署，请确保已修改默认密码和密钥"
        read -p "是否继续? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "部署已取消"
            exit 0
        fi
    fi
}

# 构建镜像
build_images() {
    log_info "构建 Docker 镜像..."
    
    # 构建前端镜像
    log_info "构建前端镜像..."
    docker build -t iptables-frontend ./frontend
    
    # 构建后端镜像
    log_info "构建后端镜像..."
    docker build -t iptables-backend ./backend
    
    log_success "镜像构建完成"
}

# 启动服务
start_services() {
    log_info "启动服务..."
    
    # 停止现有服务
    docker-compose down 2>/dev/null || true
    
    # 启动服务
    docker-compose up -d
    
    log_success "服务启动完成"
}

# 等待服务就绪
wait_for_services() {
    log_info "等待服务就绪..."
    
    # 等待数据库
    log_info "等待数据库启动..."
    timeout=60
    while [ $timeout -gt 0 ]; do
        if docker-compose exec -T mysql mysqladmin ping -h localhost --silent; then
            log_success "数据库已就绪"
            break
        fi
        sleep 2
        timeout=$((timeout-2))
    done
    
    if [ $timeout -le 0 ]; then
        log_error "数据库启动超时"
        exit 1
    fi
    
    # 等待后端
    log_info "等待后端服务启动..."
    timeout=60
    while [ $timeout -gt 0 ]; do
        if curl -f http://localhost:8080/health >/dev/null 2>&1; then
            log_success "后端服务已就绪"
            break
        fi
        sleep 2
        timeout=$((timeout-2))
    done
    
    if [ $timeout -le 0 ]; then
        log_error "后端服务启动超时"
        exit 1
    fi
    
    # 等待前端
    log_info "等待前端服务启动..."
    timeout=30
    while [ $timeout -gt 0 ]; do
        if curl -f http://localhost >/dev/null 2>&1; then
            log_success "前端服务已就绪"
            break
        fi
        sleep 2
        timeout=$((timeout-2))
    done
    
    if [ $timeout -le 0 ]; then
        log_error "前端服务启动超时"
        exit 1
    fi
}

# 显示部署信息
show_deployment_info() {
    log_success "部署完成！"
    echo
    echo "==================================="
    echo "🚀 IPTables 管理系统已启动"
    echo "==================================="
    echo "📱 前端地址: http://localhost"
    echo "🔧 后端API: http://localhost:8080"
    echo "🗄️  数据库: localhost:3306"
    echo
    echo "👤 默认账户:"
    echo "   管理员: admin / admin123"
    echo "   普通用户: user1 / user123"
    echo
    echo "📊 服务状态:"
    docker-compose ps
    echo
    echo "📝 查看日志: docker-compose logs -f"
    echo "🛑 停止服务: docker-compose down"
    echo "==================================="
}

# 清理函数
cleanup() {
    log_info "清理资源..."
    docker-compose down 2>/dev/null || true
}

# 主函数
main() {
    local env=${1:-dev}
    
    log_info "开始部署 IPTables 管理系统 ($env 环境)"
    
    # 设置错误处理
    trap cleanup ERR
    
    check_dependencies
    setup_environment "$env"
    build_images
    start_services
    wait_for_services
    show_deployment_info
    
    log_success "部署脚本执行完成"
}

# 脚本入口
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi