# 快速开始指南 - Mage 构建系统

## 🚀 一分钟快速部署

### 前提条件
确保已安装以下工具：
- Go 1.21+
- Node.js 18+
- Docker & Docker Compose

### 快速部署
```bash
# 1. 克隆项目（如果需要）
# git clone <repository-url>
# cd iptables-management

# 2. 一键部署
chmod +x scripts/mage-deploy.sh
./scripts/mage-deploy.sh deploy

# 3. 访问应用
# 打开浏览器访问: http://localhost:8080
```

## 📋 详细步骤

### 步骤 1: 安装 Mage
```bash
go install github.com/magefile/mage@latest
```

### 步骤 2: 查看可用命令
```bash
# 查看 Mage 命令
mage -l

# 或使用 Make 命令
make help
```

### 步骤 3: 构建项目
```bash
# 方式一：使用 Mage
mage clean
mage install
mage build

# 方式二：使用 Make
make clean
make install
make build
```

### 步骤 4: 部署应用
```bash
# 方式一：使用部署脚本（推荐）
./scripts/mage-deploy.sh deploy

# 方式二：手动部署
mage dockerBuild
docker-compose -f docker-compose.unified.yml up -d
```

## 🛠️ 开发模式

### 启动开发环境
```bash
# 启动完整开发环境（前端 + 后端）
mage dev

# 或分别启动
mage devFrontend  # 前端开发服务器
mage devBackend   # 后端开发服务器
```

### 开发地址
- 前端开发服务器: http://localhost:5173
- 后端开发服务器: http://localhost:8080
- 数据库: localhost:3306

## 🐳 生产部署

### 使用统一容器部署
```bash
# 构建并启动
./scripts/mage-deploy.sh deploy

# 查看状态
./scripts/mage-deploy.sh status

# 查看日志
./scripts/mage-deploy.sh logs
```

### 生产环境地址
- 应用首页: http://localhost:8080
- API 接口: http://localhost:8080/api/*
- 健康检查: http://localhost:8080/health

## 🔧 常用命令

### Mage 命令
```bash
mage build          # 构建整个项目
mage buildFrontend  # 仅构建前端
mage buildBackend   # 仅构建后端
mage clean          # 清理构建产物
mage dev            # 启动开发环境
mage test           # 运行测试
mage lint           # 代码检查
mage dockerBuild    # 构建 Docker 镜像
```

### Make 命令（兼容性）
```bash
make build          # 构建项目
make dev            # 开发环境
make deploy         # 部署应用
make clean          # 清理
make status         # 服务状态
make logs           # 查看日志
```

### 部署脚本命令
```bash
./scripts/mage-deploy.sh deploy    # 完整部署
./scripts/mage-deploy.sh build     # 仅构建
./scripts/mage-deploy.sh status    # 服务状态
./scripts/mage-deploy.sh logs      # 查看日志
./scripts/mage-deploy.sh stop      # 停止服务
./scripts/mage-deploy.sh restart   # 重启服务
./scripts/mage-deploy.sh clean     # 清理环境
```

## 📁 项目结构

```
iptables-management/
├── magefile.go                    # Mage 构建脚本
├── Dockerfile.unified             # 统一容器构建文件
├── docker-compose.unified.yml     # 统一容器编排文件
├── scripts/mage-deploy.sh         # 部署脚本
├── frontend/                      # 前端项目
│   ├── src/
│   ├── package.json
│   └── dist/                      # 构建输出
├── backend/                       # 后端项目
│   ├── main.go
│   └── go.mod
└── dist/                          # 前端构建产物（复制到根目录）
```

## 🔍 故障排除

### 常见问题

1. **Mage 未找到**
   ```bash
   export PATH=$PATH:$(go env GOPATH)/bin
   ```

2. **端口被占用**
   ```bash
   # 查看端口占用
   lsof -i :8080
   lsof -i :3306
   
   # 停止服务
   ./scripts/mage-deploy.sh stop
   ```

3. **构建失败**
   ```bash
   # 清理并重新构建
   mage clean
   mage install
   mage build
   ```

4. **容器启动失败**
   ```bash
   # 查看日志
   docker-compose -f docker-compose.unified.yml logs
   
   # 重置环境
   make reset
   ```

## 📊 监控和维护

### 健康检查
```bash
curl http://localhost:8080/health
```

### 查看服务状态
```bash
docker-compose -f docker-compose.unified.yml ps
```

### 备份数据库
```bash
make backup-db
```

### 查看资源使用
```bash
docker stats
```

## 🎯 下一步

1. 阅读完整的 [README-MAGE.md](./README-MAGE.md) 了解详细信息
2. 查看 [项目文档](./README.md) 了解功能特性
3. 参考 [部署指南](./DEPLOYMENT_GUIDE.md) 进行生产部署

## 💡 提示

- 使用 `mage -l` 查看所有可用的 Mage 命令
- 使用 `make help` 查看所有可用的 Make 命令
- 开发时推荐使用 `mage dev` 启动开发环境
- 生产部署推荐使用 `./scripts/mage-deploy.sh deploy`
- 遇到问题时先尝试 `mage clean` 清理构建产物

---

🎉 **恭喜！** 您已经成功使用 Mage 构建系统部署了 IPTables 管理系统！