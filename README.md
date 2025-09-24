# IPTables 管理系统

一个现代化的 IPTables 防火墙规则管理系统，采用前后端分离架构，提供直观的 Web 界面来管理 Linux 防火墙规则。支持网络拓扑可视化、隧道接口分析、Docker网桥通信等高级功能。

## 🏗️ 项目架构

```
iptables-manager/
├── frontend/                 # Vue3 前端项目
│   ├── src/
│   │   ├── api/             # API 接口层
│   │   │   ├── index.ts     # 主要API接口
│   │   │   └── chainTable.ts # 链表API
│   │   ├── components/      # 可复用组件
│   │   │   └── TunnelAnalysis.vue # 隧道分析组件
│   │   ├── stores/          # Pinia 状态管理
│   │   │   └── user.ts      # 用户状态
│   │   └── views/           # 页面视图
│   │       ├── Login.vue    # 登录页面
│   │       ├── ChainTableView.vue # 链表视图
│   │       ├── Rules.vue    # 规则管理
│   │       ├── Tables.vue   # 表管理
│   │       ├── Interfaces.vue # 网络接口
│   │       ├── Topology.vue # 网络拓扑
│   │       └── Logs.vue     # 操作日志
│   ├── Dockerfile           # 前端容器配置
│   ├── nginx.conf           # Nginx配置
│   └── package.json         # 前端依赖配置
├── backend/                 # Go 后端项目
│   ├── config/              # 配置管理
│   │   └── database.go      # 数据库配置
│   ├── controllers/         # 控制器层
│   │   └── tunnel_controller.go # 隧道分析控制器
│   ├── handlers/            # 请求处理器
│   │   ├── auth_handler.go  # 认证处理
│   │   ├── rule_handler.go  # 规则处理
│   │   ├── table_handler.go # 表处理
│   │   ├── network_handler.go # 网络处理
│   │   ├── topology_handler.go # 拓扑处理
│   │   ├── chain_table_handler.go # 链表处理
│   │   └── log_handler.go   # 日志处理
│   ├── middleware/          # 中间件
│   │   └── auth.go          # JWT认证中间件
│   ├── models/              # 数据模型
│   │   └── models.go        # 所有数据模型
│   ├── services/            # 业务逻辑层
│   │   ├── auth_service.go  # 认证服务
│   │   ├── rule_service.go  # 规则服务
│   │   ├── table_service.go # 表服务
│   │   ├── network_service.go # 网络服务
│   │   ├── topology_service.go # 拓扑服务
│   │   └── log_service.go   # 日志服务
│   ├── data/                # 数据存储目录
│   ├── Dockerfile           # 后端容器配置
│   ├── main.go              # 应用入口
│   └── go.mod               # Go模块配置
├── docs/                    # 项目文档
│   ├── TUNNEL_DOCKER_INTEGRATION.md # 隧道Docker集成文档
│   ├── TOPOLOGY.md          # 网络拓扑文档
│   ├── ARCHITECTURE_IMPROVEMENTS.md # 架构改进文档
│   └── ...                  # 其他技术文档
├── scripts/                 # 脚本工具
│   ├── debug.sh             # 调试脚本
│   ├── test_system.sh       # 系统测试
│   ├── test_topology.sh     # 拓扑测试
│   └── ...                  # 其他测试脚本
├── sql/                     # 数据库脚本
│   └── init.sql             # 数据库初始化
├── Dockerfile               # 统一容器配置
├── docker-compose.yml       # Docker Compose配置
├── Makefile                 # Make构建配置
├── magefile.go              # Mage构建配置
└── README.md                # 项目文档
```

## 🚀 技术栈

### 前端技术
- **Vue 3.4** - 渐进式 JavaScript 框架
- **Element Plus 2.4** - Vue 3 企业级UI组件库
- **TypeScript 5.2** - 类型安全的 JavaScript 超集
- **Vite 5.0** - 现代化前端构建工具
- **Pinia 2.1** - Vue 3 官方状态管理库
- **Vue Router 4.2** - Vue.js 官方路由管理器
- **Axios 1.6** - Promise 基于的 HTTP 客户端
- **Vue Flow** - 流程图和网络拓扑可视化
- **ECharts 5.4** - 数据可视化图表库

### 后端技术
- **Go 1.21** - 高性能系统编程语言
- **Gin 1.9** - 轻量级高性能 Web 框架
- **GORM 1.25** - Go 语言 ORM 库
- **JWT v4** - JSON Web Token 身份认证
- **SQLite** - 轻量级嵌入式数据库
- **CORS** - 跨域资源共享支持
- **Crypto** - 密码学加密支持

### 开发工具
- **Mage** - Go 语言构建工具
- **Docker** - 容器化部署平台
- **Docker Compose** - 多容器应用编排
- **Nginx** - 高性能 Web 服务器和反向代理
- **Make** - 传统构建工具支持

## 📋 功能特性

### 🔐 核心功能
- **用户认证** - JWT 令牌认证，支持安全的用户会话管理
- **规则管理** - 完整的 IPTables 规则 CRUD 操作
- **表管理** - filter、nat、mangle、raw 表的统一管理
- **链管理** - INPUT、OUTPUT、FORWARD 等链的可视化管理
- **操作日志** - 详细的操作审计和历史记录

### 🌐 网络分析
- **网络拓扑** - 交互式网络拓扑图，支持节点筛选和关系分析
- **接口管理** - 网络接口状态监控和分类显示
- **隧道分析** - tun/tap 隧道接口与 Docker 网桥通信分析
- **Docker集成** - Docker 网桥和容器网络的深度分析
- **通信路径** - 数据包流向的完整路径跟踪

### 📊 可视化功能
- **链表视图** - 规则链的表格化展示和管理
- **统计图表** - 基于 ECharts 的数据可视化
- **实时监控** - 网络接口和规则的实时状态监控
- **性能分析** - 包转发统计和性能指标展示

### 🛠️ 高级工具
- **规则生成器** - 智能生成常用的 IPTables 规则
- **批量操作** - 支持规则的批量导入导出
- **配置备份** - 系统配置的备份和恢复功能
- **调试工具** - 内置的网络调试和故障排查工具

### 📱 用户体验
- **响应式设计** - 完美适配桌面、平板和移动设备
- **主题支持** - 支持浅色和深色主题切换
- **国际化** - 多语言支持（中文/英文）
- **快捷操作** - 键盘快捷键和批量操作支持

## 🛠️ 快速开始

### 环境要求

- **Docker** 20.10+ - 容器化运行环境
- **Docker Compose** 2.0+ - 多容器编排工具
- **Linux系统** - 支持 iptables 的 Linux 发行版
- **网络权限** - 需要 NET_ADMIN 权限管理 iptables

### 一键部署

1. **克隆项目**
```bash
git clone <repository-url>
cd iptables-manager
```

2. **启动服务**
```bash
# 使用 Docker Compose
docker-compose up -d

# 或使用 Make 命令
make up
```

3. **访问应用**
- **Web界面**: http://localhost:8888
- **API接口**: http://localhost:8888/api
- **健康检查**: http://localhost:8888/health
- **数据库**: SQLite 文件存储在 `/app/data/iptables.db`

### 默认账户

- **管理员**: admin / [请在首次部署后修改默认密码]
- **普通用户**: user1 / [请在首次部署后修改默认密码]

> ⚠️ **安全提醒**: 首次登录后请立即修改默认密码，确保系统安全。

## 🔧 开发环境

### 使用 Mage 构建系统

本项目使用 [Mage](https://magefile.org/) 作为主要构建工具，同时保持对传统 Make 的兼容。

```bash
# 安装 Mage（如果未安装）
go install github.com/magefile/mage@latest

# 查看所有可用命令
mage -l

# 或使用 Make 查看帮助
make help
```

### 开发环境设置

#### 1. 安装依赖
```bash
# 使用 Mage
mage install

# 或使用 Make
make install
```

#### 2. 启动开发环境
```bash
# 启动完整开发环境（前端+后端）
mage dev
# 或
make dev

# 单独启动前端开发服务器
cd frontend
npm install
npm run dev

# 单独启动后端开发服务器
cd backend
go mod tidy
go run main.go
```

#### 3. 构建项目
```bash
# 构建整个项目
mage build
# 或
make build

# 构建 Docker 镜像
mage docker:build
# 或
make docker-build
```

### 数据库

项目使用 SQLite 作为默认数据库，数据文件位于：
```bash
# 开发环境
backend/data/iptables.db

# 生产环境（Docker）
/app/data/iptables.db
```

### 开发工具

#### 代码检查和测试
```bash
# 运行代码检查
mage lint
# 或
make lint

# 运行测试
mage test
# 或
make test
```


## 📚 API 文档

### 🔐 认证接口
- `POST /api/login` - 用户登录
- `POST /api/logout` - 用户登出
- `GET /api/user/profile` - 获取用户信息

### 🛡️ 规则管理
- `GET /api/rules` - 获取所有规则
- `POST /api/rules` - 创建新规则
- `PUT /api/rules/:id` - 更新规则
- `DELETE /api/rules/:id` - 删除规则
- `POST /api/rules/batch` - 批量操作规则

### 📊 表和链管理
- `GET /api/tables` - 获取所有表信息
- `GET /api/tables/:table/chains` - 获取指定表的链信息
- `GET /api/chain-table` - 获取链表视图数据
- `POST /api/chain-table/filter` - 筛选链表数据

### 🌐 网络管理
- `GET /api/network/interfaces` - 获取网络接口列表
- `GET /api/network/interfaces/:name` - 获取指定接口详情
- `GET /api/network/statistics` - 获取网络统计信息

### 🗺️ 网络拓扑
- `GET /api/topology` - 获取网络拓扑数据
- `GET /api/topology/nodes` - 获取拓扑节点
- `GET /api/topology/edges` - 获取拓扑边关系
- `POST /api/topology/filter` - 筛选拓扑数据

### 🔍 隧道分析
- `GET /api/tunnel/interfaces` - 获取隧道接口列表
- `GET /api/tunnel/docker-bridges` - 获取Docker网桥列表
- `GET /api/tunnel/:interface/rules` - 获取接口相关规则
- `GET /api/tunnel/:interface/info` - 获取接口详细信息
- `GET /api/tunnel/analyze-communication` - 分析通信路径
- `POST /api/tunnel/generate-rules` - 生成通信规则

### 📝 日志管理
- `GET /api/logs` - 获取操作日志
- `GET /api/logs/search` - 搜索日志
- `DELETE /api/logs` - 清理日志

### 🔧 系统管理
- `GET /api/health` - 健康检查
- `GET /api/system/info` - 获取系统信息
- `GET /api/system/status` - 获取系统状态

## 🐳 Docker 部署

### 统一容器架构

本项目采用统一容器架构，前端和后端打包在同一个容器中，简化部署和管理。

### 构建镜像

```bash
# 构建统一应用镜像
docker build -t iptables-manager .

# 或使用 Mage/Make 构建
mage docker:build
make docker-build
```

### 使用 Docker Compose

```bash
# 启动服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down

# 重启服务
docker-compose restart
```

### 容器配置

#### 环境变量
```yaml
environment:
  - PORT=8888                    # 应用端口
  - DB_PATH=/app/data/iptables.db # 数据库路径
  - JWT_SECRET=your-jwt-secret   # JWT密钥
  - GIN_MODE=release             # Gin运行模式
```

#### 网络配置
```yaml
network_mode: host               # 使用主机网络模式
cap_add:
  - NET_ADMIN                    # 网络管理权限
```

#### 数据持久化
```yaml
volumes:
  - ./data:/app/data             # 数据目录挂载
```

### 生产环境部署

#### 1. 环境准备
```bash
# 创建数据目录
mkdir -p ./data

# 设置权限
chmod 755 ./data
```

#### 2. 配置文件
```bash
# 复制并修改配置
cp docker-compose.yml docker-compose.prod.yml

# 修改生产环境配置
vim docker-compose.prod.yml
```

#### 3. 启动生产服务
```bash
# 使用生产配置启动
docker-compose -f docker-compose.prod.yml up -d

# 检查服务状态
docker-compose -f docker-compose.prod.yml ps
```

### 容器管理命令

```bash
# 进入容器
docker-compose exec app sh

# 查看容器资源使用
docker stats iptables-management-app

# 备份数据
docker-compose exec app cp /app/data/iptables.db /tmp/backup.db

# 更新应用
docker-compose pull
docker-compose up -d
```

## 🔒 安全配置

### 环境变量配置

在生产环境中，请务必修改以下环境变量：

```bash
# 应用配置
PORT=8888                                    # 应用服务端口
GIN_MODE=release                            # 生产模式

# 数据库配置
DB_PATH=/app/data/iptables.db               # SQLite数据库路径

# 安全配置
JWT_SECRET=<请生成32位以上随机密钥>          # JWT签名密钥
```

#### 生产环境配置示例
```bash
# 创建环境变量文件
cat > .env << EOF
PORT=8888
DB_PATH=/app/data/iptables.db
JWT_SECRET=$(openssl rand -base64 32)
GIN_MODE=release
EOF
```

> 🔐 **安全建议**:
> - JWT密钥使用32位以上的随机字符串
> - 定期轮换JWT密钥
> - 使用环境变量或密钥管理服务存储敏感信息
> - 生产环境中启用HTTPS
> - 定期备份SQLite数据库文件

### 网络和防火墙配置

#### 端口配置
确保以下端口在防火墙中正确配置：

- **8888** (HTTP) - 主应用端口
- **443** (HTTPS) - 生产环境HTTPS端口（推荐）
- **22** (SSH) - 远程管理端口（限制访问）

#### 防火墙规则示例
```bash
# 允许应用端口
sudo ufw allow 8888/tcp

# 允许HTTPS（生产环境）
sudo ufw allow 443/tcp

# 限制SSH访问（可选）
sudo ufw limit ssh

# 启用防火墙
sudo ufw enable
```

#### 反向代理配置（推荐）
```nginx
# /etc/nginx/sites-available/iptables-manager
server {
    listen 80;
    server_name your-domain.com;
    
    # 重定向到HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;
    
    # SSL配置
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    # 代理到应用
    location / {
        proxy_pass http://localhost:8888;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

## 📊 监控和日志

### 健康检查

#### 应用健康检查
```bash
# 主应用健康检查
curl -f http://localhost:8888/health

# 详细系统信息
curl http://localhost:8888/api/system/info

# 系统状态检查
curl http://localhost:8888/api/system/status
```

#### 自动化健康检查
```bash
# 使用Make命令
make health

# 或直接使用脚本
./scripts/test_system.sh
```

### 日志管理

#### 容器日志
```bash
# 查看应用日志
docker-compose logs app

# 实时跟踪日志
docker-compose logs -f app

# 查看最近100行日志
docker-compose logs --tail=100 app

# 查看特定时间段日志
docker-compose logs --since="2024-01-01T00:00:00" app
```

#### 应用内日志
```bash
# 查看操作日志（通过API）
curl -H "Authorization: Bearer <token>" \
     http://localhost:8888/api/logs

# 搜索日志
curl -H "Authorization: Bearer <token>" \
     "http://localhost:8888/api/logs/search?keyword=error"
```

#### 系统日志
```bash
# 查看系统iptables日志
sudo journalctl -u iptables

# 查看网络接口日志
sudo dmesg | grep -i network

# 查看Docker日志
sudo journalctl -u docker
```

### 性能监控

#### 容器资源监控
```bash
# 查看容器资源使用
docker stats iptables-management-app

# 查看容器详细信息
docker inspect iptables-management-app
```

#### 系统监控
```bash
# 网络接口统计
cat /proc/net/dev

# iptables规则统计
sudo iptables -L -n -v

# 系统负载
top -p $(docker inspect -f '{{.State.Pid}}' iptables-management-app)
```

### 日志轮转配置

#### Docker日志轮转
```yaml
# docker-compose.yml
services:
  app:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

#### 系统日志轮转
```bash
# /etc/logrotate.d/iptables-manager
/var/log/iptables-manager/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 644 root root
}
```

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🔐 安全最佳实践

### 部署前安全检查清单

- [ ] 修改所有默认密码
- [ ] 配置强密码策略
- [ ] 设置防火墙规则
- [ ] 启用HTTPS（生产环境）
- [ ] 配置访问日志和监控
- [ ] 定期备份数据库
- [ ] 更新系统和依赖包



## 🆘 故障排除

### 常见问题及解决方案

#### 1. 应用启动问题

**问题**: 容器启动失败
```bash
# 检查容器状态
docker-compose ps

# 查看启动日志
docker-compose logs app

# 检查端口占用
sudo netstat -tlnp | grep 8888
```

**解决方案**:
- 确保端口8888未被占用
- 检查数据目录权限: `chmod 755 ./data`
- 验证Docker有NET_ADMIN权限

#### 2. 权限问题

**问题**: iptables命令执行失败
```bash
# 检查容器权限
docker inspect iptables-management-app | grep -i cap

# 测试iptables权限
docker-compose exec app iptables -L
```

**解决方案**:
- 确保容器有NET_ADMIN权限
- 检查SELinux设置: `getenforce`
- 验证用户在docker组中: `groups $USER`

#### 3. 网络连接问题

**问题**: 无法访问Web界面
```bash
# 检查服务监听
docker-compose exec app netstat -tlnp

# 测试本地连接
curl -I http://localhost:8888/health

# 检查防火墙
sudo ufw status
```

**解决方案**:
- 确认防火墙允许8888端口
- 检查网络模式配置
- 验证反向代理配置（如果使用）

#### 4. 数据库问题

**问题**: SQLite数据库错误
```bash
# 检查数据库文件
ls -la ./data/iptables.db

# 测试数据库连接
docker-compose exec app sqlite3 /app/data/iptables.db ".tables"

# 检查数据库权限
docker-compose exec app ls -la /app/data/
```

**解决方案**:
- 确保数据目录可写
- 检查磁盘空间: `df -h`
- 重新初始化数据库（如果需要）

#### 5. 认证问题

**问题**: JWT认证失败
```bash
# 检查JWT配置
docker-compose exec app env | grep JWT

# 查看认证日志
docker-compose logs app | grep -i auth
```

**解决方案**:
- 验证JWT_SECRET环境变量
- 检查系统时间同步
- 清除浏览器缓存和Cookie

### 调试工具

#### 手动调试命令
```bash
# 进入容器调试
docker-compose exec app sh

# 检查iptables规则
docker-compose exec app iptables -L -n -v

# 查看网络接口
docker-compose exec app ip addr show

# 检查进程状态
docker-compose exec app ps aux
```


### 获取技术支持

如果遇到问题，请按以下步骤获取帮助：

1. **收集信息**
   ```bash
   # 生成诊断报告
   ./scripts/debug.sh > debug_report.txt
   
   # 收集系统信息
   docker-compose logs > app_logs.txt
   ```

2. **查看文档**
   - 检查 `docs/` 目录下的技术文档
   - 查看相关的故障排除指南
   - 搜索已知问题和解决方案

3. **社区支持**
   - 搜索 GitHub Issues
   - 提交详细的问题报告
   - 包含系统信息和错误日志

4. **问题报告模板**
   ```markdown
   ## 问题描述
   [详细描述遇到的问题]
   
   ## 环境信息
   - 操作系统: [Linux发行版和版本]
   - Docker版本: [docker --version]
   - 应用版本: [git commit hash]
   
   ## 重现步骤
   1. [步骤1]
   2. [步骤2]
   3. [步骤3]
   
   ## 错误日志
   [粘贴相关错误日志]
   
   ## 已尝试的解决方案
   [列出已经尝试的解决方法]
   ```

## 🌟 高级功能特色

### 🔍 隧道接口分析
本系统独有的隧道接口分析功能，支持：
- **tun/tap接口识别**: 自动发现和分析VPN隧道接口
- **Docker网桥集成**: 深度分析Docker网桥与隧道的通信路径
- **数据流跟踪**: 完整跟踪数据包从隧道到容器的传输过程
- **智能规则生成**: 根据网络拓扑自动生成优化的iptables规则

### 🗺️ 交互式网络拓扑
- **实时拓扑图**: 基于Vue Flow的动态网络拓扑可视化
- **节点筛选**: 支持按接口类型、状态等条件筛选显示
- **关系分析**: 直观展示网络接口间的连接关系
- **性能监控**: 实时显示接口流量和规则匹配统计

### 📊 链表可视化管理
- **多维度展示**: 支持按表、链、规则等多个维度组织显示
- **实时筛选**: 强大的筛选和搜索功能
- **批量操作**: 支持规则的批量编辑和管理
- **统计分析**: 详细的规则匹配和性能统计

### 🛠️ 现代化构建系统
- **Mage集成**: 使用现代化的Mage构建工具
- **Make兼容**: 保持对传统Make命令的完全兼容
- **统一容器**: 前后端统一打包，简化部署流程
- **开发友好**: 完整的开发、测试、调试工具链

## 📖 项目文档

本项目包含完整的技术文档，位于 `docs/` 目录：

- **[隧道Docker集成](docs/TUNNEL_DOCKER_INTEGRATION.md)** - 隧道接口与Docker网桥通信分析
- **[网络拓扑](docs/TOPOLOGY.md)** - 网络拓扑可视化功能说明
- **[架构改进](docs/ARCHITECTURE_IMPROVEMENTS.md)** - 系统架构设计和改进
- **[链表可视化](docs/CHAIN_TABLE_VISUALIZATION.md)** - 链表视图功能详解
- **[数据流优化](docs/DATAFLOW_OPTIMIZATION_COMPLETE.md)** - 数据流处理优化
- **[调试指南](docs/DEBUG.md)** - 系统调试和故障排除

## 🎯 使用场景

### 网络管理员
- **防火墙管理**: 直观的Web界面管理复杂的iptables规则
- **网络监控**: 实时监控网络接口状态和流量统计
- **故障排查**: 强大的网络拓扑和数据流分析工具

### DevOps工程师
- **容器网络**: Docker网桥和容器网络的深度分析
- **自动化部署**: 完整的容器化部署方案
- **监控集成**: 丰富的API接口支持监控系统集成

### 安全工程师
- **规则审计**: 完整的操作日志和规则变更记录
- **安全分析**: 网络通信路径的安全性分析
- **合规检查**: 支持安全合规性检查和报告生成

### 开发者
- **API丰富**: 完整的RESTful API支持二次开发
- **现代技术栈**: Vue3 + Go + TypeScript 现代化技术栈
- **扩展友好**: 模块化设计，易于功能扩展

---

**⚠️ 重要安全声明**: 
- 本系统包含网络安全管理功能，请确保在授权环境中使用
- 生产环境部署前必须进行完整的安全评估和渗透测试
- 定期更新系统组件，关注安全漏洞公告
- 遵循最小权限原则，仅授予必要的访问权限