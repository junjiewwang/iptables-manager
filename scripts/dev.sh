#!/bin/bash

# IPTables 管理系统开发环境脚本

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

# 检查依赖
check_dependencies() {
    log_info "检查开发环境依赖..."
    
    # 检查 Node.js
    if ! command -v node &> /dev/null; then
        log_error "Node.js 未安装，请先安装 Node.js 18+"
        exit 1
    fi
    
    # 检查 npm
    if ! command -v npm &> /dev/null; then
        log_error "npm 未安装"
        exit 1
    fi
    
    # 检查 Go
    if ! command -v go &> /dev/null; then
        log_error "Go 未安装，请先安装 Go 1.21+"
        exit 1
    fi
    
    log_success "依赖检查通过"
}

# 启动数据库
start_database() {
    log_info "启动开发数据库..."
    
    # 检查是否已有数据库容器运行
    if docker ps | grep -q iptables-mysql-dev; then
        log_info "数据库已在运行"
        return
    fi
    
    # 启动 MySQL 容器
    docker run -d \
        --name iptables-mysql-dev \
        -e MYSQL_ROOT_PASSWORD=root123456 \
        -e MYSQL_DATABASE=iptables_management \
        -e MYSQL_USER=iptables_user \
        -e MYSQL_PASSWORD=iptables_pass \
        -p 3306:3306 \
        -v "$(pwd)/sql/init.sql:/docker-entrypoint-initdb.d/init.sql" \
        mysql:8.0
    
    # 等待数据库启动
    log_info "等待数据库启动..."
    timeout=60
    while [ $timeout -gt 0 ]; do
        if docker exec iptables-mysql-dev mysqladmin ping -h localhost --silent 2>/dev/null; then
            log_success "数据库已启动"
            break
        fi
        sleep 2
        timeout=$((timeout-2))
    done
    
    if [ $timeout -le 0 ]; then
        log_error "数据库启动超时"
        exit 1
    fi
}

# 安装前端依赖
install_frontend_deps() {
    log_info "安装前端依赖..."
    cd frontend
    npm install
    cd ..
    log_success "前端依赖安装完成"
}

# 安装后端依赖
install_backend_deps() {
    log_info "安装后端依赖..."
    cd backend
    go mod tidy
    cd ..
    log_success "后端依赖安装完成"
}

# 启动前端开发服务器
start_frontend() {
    log_info "启动前端开发服务器..."
    cd frontend
    npm run dev &
    FRONTEND_PID=$!
    cd ..
    log_success "前端开发服务器已启动 (PID: $FRONTEND_PID)"
}

# 启动后端开发服务器
start_backend() {
    log_info "启动后端开发服务器..."
    cd backend
    
    # 设置开发环境变量
    export MYSQL_HOST=localhost
    export MYSQL_PORT=3306
    export MYSQL_DATABASE_NAME=iptables_management
    export MYSQL_USERNAME=iptables_user
    export MYSQL_PASSWORD=iptables_pass
    export PORT=8080
    export JWT_SECRET=iptables-management-secret-key-2024
    
    go run main.go &
    BACKEND_PID=$!
    cd ..
    log_success "后端开发服务器已启动 (PID: $BACKEND_PID)"
}

# 显示开发信息
show_dev_info() {
    log_success "开发环境启动完成！"
    echo
    echo "==================================="
    echo "🛠️  IPTables 管理系统开发环境"
    echo "==================================="
    echo "📱 前端开发服务器: http://localhost:3000"
    echo "🔧 后端开发服务器: http://localhost:8080"
    echo "🗄️  开发数据库: localhost:3306"
    echo
    echo "👤 默认账户:"
    echo "   管理员: admin / admin123"
    echo "   普通用户: user1 / user123"
    echo
    echo "🔧 开发命令:"
    echo "   停止开发环境: Ctrl+C"
    echo "   重启前端: cd frontend && npm run dev"
    echo "   重启后端: cd backend && go run main.go"
    echo "   查看数据库: docker exec -it iptables-mysql-dev mysql -u root -p"
    echo "==================================="
}

# 清理函数
cleanup() {
    log_info "清理开发环境..."
    
    # 停止前端服务器
    if [ ! -z "$FRONTEND_PID" ]; then
        kill $FRONTEND_PID 2>/dev/null || true
    fi
    
    # 停止后端服务器
    if [ ! -z "$BACKEND_PID" ]; then
        kill $BACKEND_PID 2>/dev/null || true
    fi
    
    # 停止数据库容器
    docker stop iptables-mysql-dev 2>/dev/null || true
    docker rm iptables-mysql-dev 2>/dev/null || true
    
    log_success "开发环境已清理"
}

# 等待服务就绪
wait_for_services() {
    log_info "等待服务就绪..."
    
    # 等待后端
    timeout=30
    while [ $timeout -gt 0 ]; do
        if curl -f http://localhost:8080/health >/dev/null 2>&1; then
            log_success "后端服务已就绪"
            break
        fi
        sleep 2
        timeout=$((timeout-2))
    done
    
    # 等待前端
    timeout=30
    while [ $timeout -gt 0 ]; do
        if curl -f http://localhost:3000 >/dev/null 2>&1; then
            log_success "前端服务已就绪"
            break
        fi
        sleep 2
        timeout=$((timeout-2))
    done
}

# 主函数
main() {
    log_info "启动 IPTables 管理系统开发环境"
    
    # 设置信号处理
    trap cleanup EXIT INT TERM
    
    check_dependencies
    start_database
    install_frontend_deps
    install_backend_deps
    start_backend
    start_frontend
    wait_for_services
    show_dev_info
    
    # 等待用户中断
    log_info "开发环境运行中，按 Ctrl+C 停止..."
    wait
}

# 脚本入口
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi