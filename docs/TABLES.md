# IPTables 表管理功能

## 功能概述

新增的表管理功能允许用户查看和管理 iptables 的各个表和链的详细信息，包括：

- **RAW 表**: 连接跟踪相关规则
- **MANGLE 表**: 包修改相关规则  
- **NAT 表**: 网络地址转换规则
- **FILTER 表**: 包过滤规则

## 支持的命令

系统会执行以下 iptables 命令来获取信息：

```bash
# 获取各个表的基本信息
sudo iptables -t raw -L -n --line-numbers
sudo iptables -t mangle -L -n --line-numbers
sudo iptables -t nat -L -n --line-numbers
sudo iptables -t filter -L -n --line-numbers

# 获取特定链的详细信息
iptables -L FORWARD -v
iptables -t nat -L POSTROUTING -v
iptables -L DOCKER-ISOLATION-STAGE-2 -v  # 如果存在
```

## API 端点

### 后端 API

- `GET /api/tables` - 获取所有表信息
- `GET /api/tables/:table` - 获取指定表信息
- `GET /api/tables/:table/chains/:chain` - 获取指定链的详细信息
- `GET /api/special-chains` - 获取特殊链信息（FORWARD、POSTROUTING、DOCKER相关）

### 前端页面

访问 `/tables` 路径可以查看表管理页面，包含：

1. **表概览**: 显示所有表的基本信息
2. **表详情**: 显示选定表的所有链和规则
3. **特殊链**: 显示重要的特殊链详细信息
4. **规则详情**: 点击规则可查看完整的规则信息

## 使用方法

### 1. 启动系统

```bash
# 使用 Docker
docker-compose up -d

# 或者手动启动
cd backend && go run main.go
cd frontend && npm run dev
```

### 2. 访问表管理

1. 登录系统
2. 点击左侧菜单的"表管理"
3. 选择要查看的表或特殊链
4. 查看详细的规则信息

### 3. 测试功能

运行测试脚本验证功能：

```bash
chmod +x scripts/test_tables.sh
./scripts/test_tables.sh
```

## 功能特点

### 实时数据
- 每次访问都会执行实际的 iptables 命令获取最新数据
- 支持手动刷新功能

### 详细信息
- 显示规则的行号、包数、字节数
- 显示完整的规则文本
- 支持查看链的策略和统计信息

### 用户友好
- 直观的表格和卡片布局
- 颜色编码的标签（ACCEPT=绿色，DROP=红色等）
- 可折叠的链和规则显示

### 调试支持
- 详细的日志记录
- 错误处理和用户提示
- 支持查看原始命令输出

## 权限要求

- 后端需要 sudo 权限来执行 iptables 命令
- 确保运行后端的用户在 sudoers 中配置了无密码执行 iptables 的权限

```bash
# 在 /etc/sudoers 中添加（替换 username 为实际用户名）
username ALL=(ALL) NOPASSWD: /sbin/iptables
```

## 故障排除

### 常见问题

1. **权限错误**: 确保后端有执行 sudo iptables 的权限
2. **命令不存在**: 确保系统已安装 iptables
3. **链不存在**: 某些链（如 DOCKER 相关）只在特定环境下存在
4. **数据为空**: 检查 iptables 规则是否确实存在

### 调试步骤

1. 运行测试脚本检查基本功能
2. 查看后端日志了解详细错误信息
3. 手动执行 iptables 命令验证权限
4. 检查前端控制台的网络请求

## 安全注意事项

- 该功能只读取 iptables 信息，不会修改规则
- 所有操作都会记录在操作日志中
- 建议在生产环境中限制访问权限