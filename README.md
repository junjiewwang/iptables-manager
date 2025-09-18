# IPTables 管理系统

一个现代化的 IPTables 防火墙规则管理系统，采用前后端分离架构，提供直观的 Web 界面来管理 Linux 防火墙规则。

## 🏗️ 项目架构

```
iptables-management/
├── frontend/                 # Vue3 前端项目
│   ├── src/
│   │   ├── api/             # API 接口
│   │   ├── components/      # 组件
│   │   ├── stores/          # 状态管理
│   │   └── views/           # 页面视图
│   ├── Dockerfile           # 前端 Docker 配置
│   └── package.json         # 前端依赖配置
├── backend/                 # Go 后端项目
│   ├── config/              # 配置文件
│   ├── handlers/            # 请求处理器
│   ├── middleware/          # 中间件
│   ├── models/              # 数据模型
│   ├── services/            # 业务逻辑
│   ├── Dockerfile           # 后端 Docker 配置
│   └── go.mod               # Go 模块配置
├── sql/                     # 数据库脚本
│   └── init.sql             # 数据库初始化脚本
├── compose.yaml             # Docker Compose 配置
└── README.md                # 项目文档
```

## 🚀 技术栈

### 前端
- **Vue 3** - 渐进式 JavaScript 框架
- **Element Plus** - Vue 3 UI 组件库
- **TypeScript** - 类型安全的 JavaScript
- **Vite** - 现代化构建工具
- **Pinia** - Vue 状态管理
- **Vue Router** - 路由管理
- **Axios** - HTTP 客户端

### 后端
- **Go 1.21** - 高性能编程语言
- **Gin** - 轻量级 Web 框架
- **GORM** - Go ORM 库
- **JWT** - 身份认证
- **MySQL** - 关系型数据库

### 部署
- **Docker** - 容器化部署
- **Docker Compose** - 多容器编排
- **Nginx** - 反向代理和静态文件服务

## 📋 功能特性

- 🔐 **用户认证** - JWT 令牌认证，支持管理员和普通用户角色
- 📊 **仪表盘** - 实时统计信息和可视化图表
- 🛡️ **规则管理** - 创建、编辑、删除 IPTables 规则
- 🌐 **网络拓扑** - 可视化网络拓扑图
- 📝 **操作日志** - 详细的操作审计日志
- 🔄 **实时监控** - 系统状态实时监控
- 📱 **响应式设计** - 支持桌面和移动设备

## 🛠️ 快速开始

### 环境要求

- Docker 20.10+
- Docker Compose 2.0+

### 一键部署

1. **克隆项目**
```bash
git clone <repository-url>
cd iptables-management
```

2. **启动服务**
```bash
docker-compose up -d
```

3. **访问应用**
- 前端界面: http://localhost
- 后端API: http://localhost:8080
- 数据库: localhost:3306

### 默认账户

- **管理员**: admin / admin123
- **普通用户**: user1 / user123

## 🔧 开发环境

### 前端开发

```bash
cd frontend
npm install
npm run dev
```

### 后端开发

```bash
cd backend
go mod tidy
go run main.go
```

### 数据库

```bash
# 连接到 MySQL
mysql -h localhost -P 3306 -u iptables_user -p
# 密码: iptables_pass
```

## 📚 API 文档

### 认证接口

- `POST /api/login` - 用户登录

### 规则管理

- `GET /api/rules` - 获取所有规则
- `POST /api/rules` - 创建新规则
- `PUT /api/rules/:id` - 更新规则
- `DELETE /api/rules/:id` - 删除规则

### 统计信息

- `GET /api/statistics` - 获取统计数据

### 操作日志

- `GET /api/logs` - 获取操作日志

## 🐳 Docker 部署

### 构建镜像

```bash
# 构建前端镜像
docker build -t iptables-frontend ./frontend

# 构建后端镜像
docker build -t iptables-backend ./backend
```

### 使用 Docker Compose

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

## 🔒 安全配置

### 环境变量

在生产环境中，请修改以下环境变量：

```bash
# 数据库配置
MYSQL_ROOT_PASSWORD=your_root_password
MYSQL_PASSWORD=your_db_password

# JWT 密钥
JWT_SECRET=your_jwt_secret_key
```

### 防火墙规则

确保以下端口在防火墙中正确配置：

- 80 (HTTP)
- 443 (HTTPS)
- 3306 (MySQL，仅内部访问)
- 8080 (后端API，仅内部访问)

## 📊 监控和日志

### 健康检查

- 前端: http://localhost/
- 后端: http://localhost:8080/health
- 数据库: 通过 Docker 健康检查

### 日志查看

```bash
# 查看所有服务日志
docker-compose logs

# 查看特定服务日志
docker-compose logs frontend
docker-compose logs backend
docker-compose logs mysql
```

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🆘 故障排除

### 常见问题

1. **数据库连接失败**
   - 检查 MySQL 服务是否启动
   - 验证数据库连接参数
   - 确认网络连接正常

2. **前端无法访问后端**
   - 检查后端服务状态
   - 验证 API 端点配置
   - 查看 CORS 设置

3. **Docker 构建失败**
   - 清理 Docker 缓存: `docker system prune`
   - 检查 Dockerfile 语法
   - 验证依赖文件存在

### 获取帮助

如果遇到问题，请：

1. 查看项目文档
2. 检查 GitHub Issues
3. 提交新的 Issue

---

**注意**: 本系统仅用于学习和测试目的，在生产环境中使用前请进行充分的安全评估。