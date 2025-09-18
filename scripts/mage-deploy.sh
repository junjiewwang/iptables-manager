#!/bin/bash

# IPTables 管理系统 - Mage 部署脚本
# 使用 Mage 构建工具进行前后端统一构建和部署

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

# 检查必要的工具
check_requirements() {
    log_info "检查必要的工具..."
    
    # 检查 Go
    if ! command -v go &> /dev/null; then
        log_error "Go 未安装，请先安装 Go"
        exit 1
    fi
    
    # 检查 Node.js
    if ! command -v node &> /dev/null; then
        log_error "Node.js 未安装，请先安装 Node.js"
        exit 1
    fi
    
    # 检查 Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi
    
    # 检查 Docker Compose
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        log_error "Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    fi
    
    log_success "所有必要工具检查完成"
}

# 安装 Mage
install_mage() {
    log_info "检查并安装 Mage..."
    
    if ! command -v mage &> /dev/null; then
        log_info "Mage 未安装，正在安装..."
        go install github.com/magefile/mage@latest
        
        # 检查安装是否成功
        if ! command -v mage &> /dev/null; then
            log_error "Mage 安装失败，请检查 GOPATH 和 PATH 环境变量"
            exit 1
        fi
        log_success "Mage 安装成功"
    else
        log_success "Mage 已安装"
    fi
}

# 清理旧的构建产物
cleanup() {
    log_info "清理旧的构建产物..."
    
    # 使用 Mage 清理
    mage clean
    
    # 停止并删除旧的容器
    if docker ps -a | grep -q "iptables-management"; then
        log_info "停止并删除旧的容器..."
        docker-compose -f docker-compose.unified.yml down --remove-orphans
    fi
    
    # 删除旧的镜像
    if docker images | grep -q "iptables-management"; then
        log_info "删除旧的镜像..."
        docker rmi $(docker images | grep "iptables-management" | awk '{print $3}') 2>/dev/null || true
    fi
    
    log_success "清理完成"
}

# 构建应用
build_app() {
    log_info "使用 Mage 构建应用..."
    
    # 安装依赖
    log_info "安装依赖..."
    mage install
    
    # 构建前后端
    log_info "构建前后端应用..."
    mage build
    
    log_success "应用构建完成"
}

# 构建 Docker 镜像
build_docker() {
    log_info "构建 Docker 镜像..."
    
    # 构建镜像
    docker build -f Dockerfile.unified -t iptables-management:latest .
    
    log_success "Docker 镜像构建完成"
}

# 部署应用
deploy_app() {
    log_info "部署应用..."
    
    # 启动服务
    docker-compose -f docker-compose.unified.yml up -d
    
    # 等待服务启动
    log_info "等待服务启动..."
    sleep 10
    
    # 检查服务状态
    if docker-compose -f docker-compose.unified.yml ps | grep -q "Up"; then
        log_success "服务启动成功"
        
        # 显示服务信息
        echo ""
        log_info "服务信息："
        echo "  - 应用地址: http://localhost:8080"
        echo "  - 数据库地址: localhost:3306"
        echo ""
        
        # 显示容器状态
        docker-compose -f docker-compose.unified.yml ps
        
    else
        log_error "服务启动失败"
        log_info "查看日志："
        docker-compose -f docker-compose.unified.yml logs
        exit 1
    fi
}

# 健康检查
health_check() {
    log_info "执行健康检查..."
    
    # 等待应用完全启动
    sleep 5
    
    # 检查应用健康状态
    if curl -f http://localhost:8080/health > /dev/null 2>&1; then
        log_success "应用健康检查通过"
    else
        log_warning "应用健康检查失败，请检查日志"
        docker-compose -f docker-compose.unified.yml logs app
    fi
}

# 显示帮助信息
show_help() {
    echo "IPTables 管理系统 - Mage 部署脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  build     仅构建应用"
    echo "  deploy    构建并部署应用"
    echo "  clean     清理构建产物和容器"
    echo "  logs      查看应用日志"
    echo "  status    查看服务状态"
    echo "  stop      停止服务"
    echo "  restart   重启服务"
    echo "  help      显示此帮助信息"
    echo ""
}

# 主函数
main() {
    case "${1:-deploy}" in
        "build")
            check_requirements
            install_mage
            build_app
            build_docker
            ;;
        "deploy")
            check_requirements
            install_mage
            cleanup
            build_app
            build_docker
            deploy_app
            health_check
            ;;
        "clean")
            cleanup
            ;;
        "logs")
            docker-compose -f docker-compose.unified.yml logs -f
            ;;
        "status")
            docker-compose -f docker-compose.unified.yml ps
            ;;
        "stop")
            docker-compose -f docker-compose.unified.yml down
            ;;
        "restart")
            docker-compose -f docker-compose.unified.yml restart
            ;;
        "help")
            show_help
            ;;
        *)
            log_error "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"