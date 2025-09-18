# IPTables 管理系统 - 项目结构

## 📁 项目目录结构

```
iptables-management/
├── 📁 frontend/                    # Vue3 前端项目
│   ├── 📁 src/
│   │   ├── 📁 api/                 # API 接口定义
│   │   │   └── index.ts            # 统一的 API 服务
│   │   ├── 📁 stores/              # Pinia 状态管理
│   │   │   └── user.ts             # 用户状态管理
│   │   ├── 📁 views/               # 页面组件
│   │   │   ├── Login.vue           # 登录页面
│   │   │   ├── Dashboard.vue       # 仪表盘页面
│   │   │   ├── Rules.vue           # 规则管理页面
│   │   │   ├── Topology.vue        # 网络拓扑页面
│   │   │   └── Logs.vue            # 操作日志页面
│   │   ├── App.vue                 # 根组件
│   │   └── main.ts                 # 应用入口
│   ├── index.html                  # HTML 模板
│   ├── package.json                # 前端依赖配置
│   ├── vite.config.ts              # Vite 构建配置
│   ├── tsconfig.json               # TypeScript 配置
│   ├── tsconfig.node.json          # Node.js TypeScript 配置
│   ├── Dockerfile                  # 前端 Docker 配置
│   └── nginx.conf                  # Nginx 配置文件
│
├── 📁 backend/                     # Go 后端项目
│   ├── 📁 config/                  # 配置文件
│   │   └── database.go             # 数据库配置和连接
│   ├── 📁 models/                  # 数据模型
│   │   └── models.go               # GORM 数据模型定义
│   ├── 📁 services/                # 业务逻辑层
│   │   ├── auth_service.go         # 认证服务
│   │   ├── rule_service.go         # 规则管理服务
│   │   └── log_service.go          # 日志服务
│   ├── 📁 handlers/                # HTTP 处理器
│   │   ├── auth_handler.go         # 认证处理器
│   │   ├── rule_handler.go         # 规则处理器
│   │   └── log_handler.go          # 日志处理器
│   ├── 📁 middleware/              # 中间件
│   │   └── auth.go                 # JWT 认证中间件
│   ├── main.go                     # 应用入口
│   ├── go.mod                      # Go 模块配置
│   ├── .env                        # 环境变量配置
│   └── Dockerfile                  # 后端 Docker 配置
│
├── 📁 sql/                         # 数据库脚本
│   └── init.sql                    # 数据库初始化脚本
│
├── 📁 scripts/                     # 部署和开发脚本
│   ├── deploy.sh                   # 生产环境部署脚本
│   └── dev.sh                      # 开发环境启动脚本
│
├── compose.yaml                    # Docker Compose 配置
├── Makefile                        # 项目管理命令
├── README.md                       # 项目说明文档
└── PROJECT_STRUCTURE.md            # 项目结构说明（本文件）
```

## 🏗️ 架构设计

### 前端架构 (Vue3 + TypeScript)
- **框架**: Vue 3 + Composition API
- **UI库**: Element Plus
- **状态管理**: Pinia
- **路由**: Vue Router
- **HTTP客户端**: Axios
- **构建工具**: Vite
- **类型检查**: TypeScript

### 后端架构 (Go + Gin + GORM)
- **Web框架**: Gin
- **ORM**: GORM
- **数据库**: MySQL 8.0
- **认证**: JWT
- **架构模式**: 分层架构 (Handler -> Service -> Model)

### 数据库设计
- **users**: 用户表
- **iptables_rules**: IPTables规则表
- **operation_logs**: 操作日志表

## 🚀 部署架构

### Docker 容器化
- **frontend**: Nginx + Vue3 构建产物
- **backend**: Go 二进制文件 + Alpine Linux
- **mysql**: MySQL 8.0 官方镜像

### 网络架构
```
Internet
    ↓
[Nginx:80] → [Frontend Container]
    ↓
[Backend:8080] → [Backend Container]
    ↓
[MySQL:3306] → [Database Container]
```

## 📋 功能模块

### 1. 用户认证模块
- **登录/登出**: JWT 令牌认证
- **权限控制**: 管理员/普通用户角色
- **会话管理**: 自动令牌刷新

### 2. 规则管理模块
- **CRUD操作**: 创建、读取、更新、删除规则
- **规则验证**: 表单验证和规则语法检查
- **批量操作**: 批量导入/导出规则

### 3. 监控模块
- **实时统计**: 规则数量、链状态统计
- **可视化图表**: ECharts 图表展示
- **系统状态**: 服务健康状态监控

### 4. 日志模块
- **操作审计**: 详细的用户操作记录
- **日志查询**: 按用户、时间、操作类型筛选
- **日志导出**: CSV 格式导出功能

### 5. 拓扑图模块
- **网络可视化**: 基于 ECharts 的网络拓扑图
- **交互式操作**: 节点点击、缩放、拖拽
- **规则关系**: 可视化规则之间的关系

## 🔧 开发工作流

### 本地开发
1. **前端开发**: `make frontend` 或 `cd frontend && npm run dev`
2. **后端开发**: `make backend` 或 `cd backend && go run main.go`
3. **数据库**: `make database` 启动开发数据库

### 集成开发
1. **一键启动**: `make dev` 启动完整开发环境
2. **代码热重载**: 前后端支持代码热重载
3. **API代理**: Vite 开发服务器代理后端API

### 生产部署
1. **构建镜像**: `make build`
2. **启动服务**: `make up`
3. **查看状态**: `make status`
4. **查看日志**: `make logs`

## 🔒 安全特性

### 认证安全
- JWT 令牌认证
- 密码 bcrypt 加密
- 令牌过期机制

### API安全
- CORS 跨域配置
- 请求参数验证
- SQL注入防护 (GORM)

### 部署安全
- 容器化隔离
- 环境变量配置
- 网络访问控制

## 📊 性能优化

### 前端优化
- 代码分割和懒加载
- 静态资源缓存
- Gzip 压缩

### 后端优化
- 数据库连接池
- 查询优化
- 响应缓存

### 部署优化
- 多阶段 Docker 构建
- 镜像体积优化
- 健康检查机制

## 🧪 测试策略

### 单元测试
- Go 后端单元测试
- Vue 组件测试

### 集成测试
- API 接口测试
- 数据库集成测试

### 端到端测试
- 用户流程测试
- 浏览器兼容性测试

## 📈 监控和日志

### 应用监控
- 健康检查端点
- 性能指标收集
- 错误日志记录

### 基础设施监控
- Docker 容器状态
- 数据库连接状态
- 系统资源使用

## 🔄 CI/CD 流程

### 持续集成
1. 代码提交触发构建
2. 运行测试套件
3. 构建 Docker 镜像
4. 推送到镜像仓库

### 持续部署
1. 从镜像仓库拉取
2. 更新 Docker Compose 配置
3. 滚动更新服务
4. 健康检查验证

## 📚 扩展性设计

### 水平扩展
- 无状态后端设计
- 数据库读写分离
- 负载均衡支持

### 功能扩展
- 插件化架构
- API 版本控制
- 微服务拆分准备

---

**注意**: 这是一个学习和演示项目，在生产环境使用前请进行充分的安全评估和性能测试。