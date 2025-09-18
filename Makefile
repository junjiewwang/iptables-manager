# IPTables 管理系统 Makefile

.PHONY: help dev build up down clean logs frontend backend database

# 默认目标
help:
	@echo "IPTables 管理系统 - 可用命令:"
	@echo ""
	@echo "开发环境:"
	@echo "  make dev          - 启动开发环境"
	@echo "  make frontend     - 启动前端开发服务器"
	@echo "  make backend      - 启动后端开发服务器"
	@echo "  make database     - 启动开发数据库"
	@echo ""
	@echo "生产环境:"
	@echo "  make build        - 构建所有镜像"
	@echo "  make up           - 启动所有服务"
	@echo "  make down         - 停止所有服务"
	@echo "  make restart      - 重启所有服务"
	@echo ""
	@echo "维护:"
	@echo "  make logs         - 查看服务日志"
	@echo "  make clean        - 清理Docker资源"
	@echo "  make reset        - 重置整个环境"

# 开发环境
dev:
	@echo "启动开发环境..."
	@chmod +x scripts/dev.sh
	@./scripts/dev.sh

# 前端开发
frontend:
	@echo "启动前端开发服务器..."
	@cd frontend && npm install && npm run dev

# 后端开发
backend:
	@echo "启动后端开发服务器..."
	@cd backend && go mod tidy && go run main.go

# 开发数据库
database:
	@echo "启动开发数据库..."
	@docker run -d \
		--name iptables-mysql-dev \
		-e MYSQL_ROOT_PASSWORD=root123456 \
		-e MYSQL_DATABASE=iptables_management \
		-e MYSQL_USER=iptables_user \
		-e MYSQL_PASSWORD=iptables_pass \
		-p 3306:3306 \
		-v $(PWD)/sql/init.sql:/docker-entrypoint-initdb.d/init.sql \
		mysql:8.0

# 生产环境
build:
	@echo "构建Docker镜像..."
	@docker-compose build

up:
	@echo "启动生产环境..."
	@chmod +x scripts/deploy.sh
	@./scripts/deploy.sh prod

down:
	@echo "停止所有服务..."
	@docker-compose down

restart: down up

# 维护命令
logs:
	@echo "查看服务日志..."
	@docker-compose logs -f

clean:
	@echo "清理Docker资源..."
	@docker-compose down -v
	@docker system prune -f
	@docker volume prune -f

reset: clean
	@echo "重置整个环境..."
	@docker-compose down -v --remove-orphans
	@docker system prune -af
	@docker volume prune -f

# 快速部署
deploy:
	@echo "快速部署..."
	@chmod +x scripts/deploy.sh
	@./scripts/deploy.sh

# 检查服务状态
status:
	@echo "服务状态:"
	@docker-compose ps

# 进入容器
shell-frontend:
	@docker-compose exec frontend sh

shell-backend:
	@docker-compose exec backend sh

shell-mysql:
	@docker-compose exec mysql mysql -u root -p

# 备份数据库
backup-db:
	@echo "备份数据库..."
	@docker-compose exec mysql mysqldump -u root -p iptables_management > backup_$(shell date +%Y%m%d_%H%M%S).sql

# 安装依赖
install:
	@echo "安装前端依赖..."
	@cd frontend && npm install
	@echo "安装后端依赖..."
	@cd backend && go mod tidy