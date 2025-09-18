# IPTables 管理系统 - 部署指南

## 🚀 快速开始

### 方式一：使用 Docker Compose（推荐）

1. **克隆项目**
```bash
git clone <repository-url>
cd iptables-management
```

2. **一键部署**
```bash
# 使用 Makefile
make deploy

# 或直接使用脚本
chmod +x scripts/deploy.sh
./scripts/deploy.sh
```

3. **访问应用**
- 前端地址: http://localhost
- 后端API: http://localhost:8080
- 默认账户: admin / admin123

### 方式二：开发环境

```bash
# 启动开发环境
make dev

# 或使用脚本
chmod +x scripts/dev.sh
./scripts/dev.sh
```

## 📋 环境要求

### 生产环境
- Docker 20.10+
- Docker Compose 2.0+
- 2GB+ RAM
- 10GB+ 磁盘空间

### 开发环境
- Node.js 18+
- Go 1.21+
- MySQL 8.0+
- Docker（用于数据库）

## 🔧 配置说明

### 数据库配置
项目已配置使用云数据库：
- 主机: 11.142.154.110:3306
- 数据库: 9wqn1hsc
- 用户名: with_ddfttbicvjvxvtly
- 密码: nsK9lQ!iiL4)di

### 环境变量
后端环境变量在 `/backend/.env` 文件中：
```env
MYSQL_HOST=11.142.154.110
MYSQL_PORT=3306
MYSQL_DATABASE_NAME=9wqn1hsc
MYSQL_USERNAME=with_ddfttbicvjvxvtly
MYSQL_PASSWORD=nsK9lQ!iiL4)di
PORT=8080
JWT_SECRET=iptables-management-secret-key-2024
```

## 🏗️ 项目架构

### 目录结构
```
iptables-management/
├── frontend/          # Vue3 前端
├── backend/           # Go 后端
├── sql/              # 数据库脚本
├── scripts/          # 部署脚本
├── compose.yaml      # Docker Compose 配置
└── Makefile         # 项目管理命令
```

### 技术栈
- **前端**: Vue3 + Element Plus + TypeScript + Vite
- **后端**: Go + Gin + GORM + JWT
- **数据库**: MySQL 8.0
- **部署**: Docker + Docker Compose + Nginx

## 📊 功能特性

### 核心功能
- ✅ 用户认证（JWT）
- ✅ 规则管理（CRUD）
- ✅ 实时统计
- ✅ 操作日志
- ✅ 网络拓扑图
- ✅ 响应式设计

### 安全特性
- 🔐 JWT 令牌认证
- 🔒 密码加密存储
- 🛡️ SQL 注入防护
- 🚫 CORS 跨域保护

## 🎯 使用指南

### 1. 登录系统
- 访问 http://localhost
- 使用默认账户登录：
  - 管理员: admin / admin123
  - 普通用户: user1 / user123

### 2. 管理规则
- 点击"规则管理"菜单
- 添加、编辑、删除 IPTables 规则
- 支持多种协议和目标动作

### 3. 查看统计
- 仪表盘显示实时统计信息
- 图表展示规则分布和操作趋势

### 4. 查看日志
- "操作日志"菜单查看所有操作记录
- 支持按用户、操作类型筛选

### 5. 网络拓扑
- "拓扑图"菜单查看网络结构
- 交互式图表展示规则关系

## 🔧 常用命令

### Makefile 命令
```bash
make help          # 查看所有可用命令
make dev           # 启动开发环境
make build         # 构建 Docker 镜像
make up            # 启动生产环境
make down          # 停止所有服务
make logs          # 查看服务日志
make clean         # 清理 Docker 资源
make status        # 查看服务状态
```

### Docker Compose 命令
```bash
docker-compose up -d        # 后台启动服务
docker-compose down         # 停止服务
docker-compose ps           # 查看服务状态
docker-compose logs -f      # 查看实时日志
docker-compose restart      # 重启服务
```

## 🐛 故障排除

### 常见问题

1. **端口冲突**
```bash
# 检查端口占用
netstat -tulpn | grep :80
netstat -tulpn | grep :8080

# 修改端口（在 compose.yaml 中）
ports:
  - "8080:80"  # 改为其他端口
```

2. **数据库连接失败**
```bash
# 检查数据库连接
docker-compose logs backend

# 验证数据库配置
cat backend/.env
```

3. **前端无法访问后端**
```bash
# 检查后端服务状态
curl http://localhost:8080/health

# 检查 CORS 配置
docker-compose logs backend | grep CORS
```

4. **Docker 构建失败**
```bash
# 清理 Docker 缓存
docker system prune -f

# 重新构建
docker-compose build --no-cache
```

### 日志查看
```bash
# 查看所有服务日志
docker-compose logs

# 查看特定服务日志
docker-compose logs frontend
docker-compose logs backend

# 实时查看日志
docker-compose logs -f --tail=100
```

## 🔄 更新部署

### 更新代码
```bash
# 拉取最新代码
git pull origin main

# 重新构建和部署
make build
make up
```

### 数据备份
```bash
# 备份数据库
make backup-db

# 或手动备份
docker-compose exec mysql mysqldump -u root -p iptables_management > backup.sql
```

## 🚀 生产环境优化

### 性能优化
1. **启用 Gzip 压缩**（已在 nginx.conf 中配置）
2. **静态资源缓存**（已配置）
3. **数据库连接池**（已在后端配置）

### 安全加固
1. **修改默认密码**
```bash
# 修改 JWT 密钥
vim backend/.env
# 修改 JWT_SECRET
```

2. **配置 HTTPS**
```bash
# 在 nginx.conf 中添加 SSL 配置
# 或使用反向代理（如 Traefik、Nginx Proxy Manager）
```

3. **网络隔离**
```yaml
# 在 compose.yaml 中配置内部网络
networks:
  internal:
    driver: bridge
    internal: true
```

## 📞 技术支持

### 获取帮助
1. 查看项目文档
2. 检查 GitHub Issues
3. 查看日志文件
4. 联系技术支持

### 贡献代码
1. Fork 项目
2. 创建功能分支
3. 提交 Pull Request
4. 代码审查

---

**注意**: 这是一个演示项目，在生产环境使用前请进行充分的安全评估和性能测试。