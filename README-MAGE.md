# IPTables 管理系统 - Mage 构建指南

本项目使用 [Mage](https://magefile.org/) 作为构建工具来管理前后端的构建和部署流程。

## 什么是 Mage？

Mage 是一个使用 Go 编写的构建工具，类似于 Make，但使用 Go 语言编写构建脚本，提供了更好的跨平台支持和类型安全。

## 安装要求

### 必需工具
- Go 1.21+
- Node.js 18+
- Docker
- Docker Compose

### 安装 Mage

```bash
# 安装 Mage
go install github.com/magefile/mage@latest

# 验证安装
mage -version
```

## 项目结构

```
.
├── magefile.go              # Mage 构建脚本
├── go.mod                   # Go 模块文件（用于 Mage）
├── Dockerfile.unified       # 统一的 Docker 构建文件
├── docker-compose.unified.yml # 统一的 Docker Compose 配置
├── scripts/
│   └── mage-deploy.sh      # Mage 部署脚本
├── frontend/               # 前端项目
│   ├── src/
│   ├── package.json
│   └── vite.config.ts
├── backend/                # 后端项目
│   ├── main.go
│   ├── go.mod
│   └── ...
└── dist/                   # 前端构建输出目录
```

## 可用的 Mage 命令

### 查看所有可用命令
```bash
mage -l
```

### 基本构建命令

#### 构建整个项目
```bash
mage build
```

#### 仅构建前端
```bash
mage buildFrontend
```

#### 仅构建后端
```bash
mage buildBackend
```

#### 清理构建产物
```bash
mage clean
```

### 开发命令

#### 启动完整开发环境
```bash
mage dev
```

#### 仅启动前端开发服务器
```bash
mage devFrontend
```

#### 仅启动后端开发服务器
```bash
mage devBackend
```

### 依赖管理

#### 安装所有依赖
```bash
mage install
```

### 测试和代码质量

#### 运行所有测试
```bash
mage test
```

#### 运行代码检查
```bash
mage lint
```

### Docker 相关

#### 构建 Docker 镜像
```bash
mage dockerBuild
```

#### 构建并运行 Docker 容器
```bash
mage docker
```

## 部署流程

### 方式一：使用 Mage 直接部署

```bash
# 1. 清理环境
mage clean

# 2. 安装依赖
mage install

# 3. 构建项目
mage build

# 4. 构建 Docker 镜像
mage dockerBuild

# 5. 使用 Docker Compose 启动服务
docker-compose -f docker-compose.unified.yml up -d
```

### 方式二：使用部署脚本

```bash
# 给脚本添加执行权限
chmod +x scripts/mage-deploy.sh

# 完整部署
./scripts/mage-deploy.sh deploy

# 仅构建
./scripts/mage-deploy.sh build

# 查看服务状态
./scripts/mage-deploy.sh status

# 查看日志
./scripts/mage-deploy.sh logs

# 停止服务
./scripts/mage-deploy.sh stop

# 重启服务
./scripts/mage-deploy.sh restart

# 清理环境
./scripts/mage-deploy.sh clean
```

## 构建流程说明

### 前端构建流程
1. 进入 `frontend` 目录
2. 安装 npm 依赖 (`npm install`)
3. 执行构建 (`npm run build`)
4. 将构建结果复制到根目录的 `dist` 文件夹

### 后端构建流程
1. 进入 `backend` 目录
2. 整理 Go 模块依赖 (`go mod tidy`)
3. 编译 Go 应用程序到根目录 (`iptables-backend`)

### Docker 构建流程
1. 使用多阶段构建
2. 第一阶段：构建前端应用
3. 第二阶段：构建后端应用
4. 第三阶段：创建运行时镜像，包含前后端构建结果

## 统一容器架构

在统一容器架构中：

- **前端**：构建为静态文件，放在 `dist` 目录
- **后端**：Go 应用程序，提供 API 服务和静态文件服务
- **静态文件服务**：后端通过 Gin 框架提供静态文件访问
- **路由处理**：支持前端 SPA 路由，API 请求和静态资源请求分离

### 访问地址
- 应用首页：`http://localhost:8080/`
- API 接口：`http://localhost:8080/api/*`
- 静态资源：`http://localhost:8080/static/*`
- 健康检查：`http://localhost:8080/health`

## 开发工作流

### 日常开发
```bash
# 启动开发环境（前后端同时启动）
mage dev
```

### 前端开发
```bash
# 仅启动前端开发服务器
mage devFrontend
```

### 后端开发
```bash
# 仅启动后端开发服务器
mage devBackend
```

### 测试构建
```bash
# 测试完整构建流程
mage clean && mage build
```

## 环境变量配置

在 `backend/.env` 文件中配置环境变量：

```env
PORT=8080
DB_HOST=localhost
DB_PORT=3306
DB_USER=iptables_user
DB_PASSWORD=iptables_pass
DB_NAME=iptables_management
JWT_SECRET=your-super-secret-jwt-key
GIN_MODE=release
```

## 故障排除

### 常见问题

1. **Mage 命令未找到**
   ```bash
   # 确保 GOPATH/bin 在 PATH 中
   export PATH=$PATH:$(go env GOPATH)/bin
   ```

2. **前端构建失败**
   ```bash
   # 清理 node_modules 并重新安装
   cd frontend
   rm -rf node_modules package-lock.json
   npm install
   ```

3. **后端构建失败**
   ```bash
   # 清理 Go 模块缓存
   cd backend
   go clean -modcache
   go mod tidy
   ```

4. **Docker 构建失败**
   ```bash
   # 清理 Docker 缓存
   docker system prune -f
   docker builder prune -f
   ```

### 查看日志

```bash
# 查看构建日志
mage -v build

# 查看容器日志
docker-compose -f docker-compose.unified.yml logs -f

# 查看特定服务日志
docker-compose -f docker-compose.unified.yml logs -f app
```

## 性能优化

### 构建优化
- 使用 Docker 多阶段构建减少镜像大小
- 前端资源压缩和优化
- Go 应用程序静态编译

### 运行时优化
- 使用非 root 用户运行容器
- 配置健康检查
- 资源限制和监控

## 与传统 Makefile 的对比

| 特性 | Makefile | Mage |
|------|----------|------|
| 语言 | Shell/Make | Go |
| 跨平台 | 有限 | 优秀 |
| 类型安全 | 无 | 有 |
| 依赖管理 | 基础 | 强大 |
| 错误处理 | 有限 | 完善 |
| IDE 支持 | 有限 | 优秀 |

## 总结

使用 Mage 构建系统的优势：

1. **类型安全**：使用 Go 语言编写，提供编译时类型检查
2. **跨平台**：在 Windows、macOS、Linux 上行为一致
3. **强大的依赖管理**：支持复杂的构建依赖关系
4. **易于维护**：Go 代码比 Shell 脚本更易读和维护
5. **集成开发**：与 Go 生态系统完美集成

通过 Mage，我们实现了前后端的统一构建和部署，简化了开发和运维流程。