# 网络数据获取重构文档

## 问题描述

原系统依赖于Docker CLI命令获取网络信息，在容器化环境中由于未安装Docker CLI导致功能无法正常工作。

## 解决方案

### 1. 重构网络服务

**原实现问题：**
- 依赖 `docker network ls` 和 `docker network inspect` 命令
- 容器内无Docker CLI导致命令执行失败
- 无法获取Docker网桥和容器信息

**新实现方案：**
- 使用系统原生 `ip` 命令获取网络接口信息
- 使用 `netstat` 命令获取网络连接状态
- 通过解析系统网络配置文件获取详细信息

### 2. 主要改进内容

#### 2.1 Docker网桥信息获取
```go
// 原方法：依赖docker命令
cmd := exec.Command("docker", "network", "ls", "--format", "{{.ID}}\t{{.Name}}\t{{.Driver}}\t{{.Scope}}")

// 新方法：使用系统原生命令
interfaces, err := s.GetAllInterfaces()
// 筛选Docker相关的网桥接口
for _, iface := range interfaces {
    if s.isDockerBridge(iface.Name) {
        bridge, err := s.createBridgeFromInterface(iface)
        // ...
    }
}
```

#### 2.2 网桥配置信息获取
```go
// 使用ip命令获取网桥的IP配置
cmd := exec.Command("ip", "addr", "show", bridgeName)
// 解析输出获取CIDR、网关等信息
```

#### 2.3 容器信息获取
```go
// 使用ip命令获取连接到网桥的veth接口
cmd := exec.Command("ip", "link", "show", "master", bridgeName)
// 解析veth接口获取容器相关信息
```

### 3. 新增功能

#### 3.1 网络连接信息
- 使用 `netstat -tuln` 获取网络连接状态
- 解析TCP/UDP连接信息
- 提供本地地址、远程地址、连接状态等信息

#### 3.2 路由表信息
- 使用 `ip route show` 获取路由表
- 解析目标网络、网关、接口等信息
- 支持默认路由和特定路由的识别

### 4. 容器环境优化

#### 4.1 Dockerfile更新
```dockerfile
# 添加必要的网络工具包
RUN apt-get update && apt-get install -y ca-certificates iptables sqlite3 iproute2 net-tools && rm -rf /var/lib/apt/lists/*
```

#### 4.2 新增工具包
- `iproute2`: 提供 `ip` 命令
- `net-tools`: 提供 `netstat` 命令

### 5. API接口扩展

#### 5.1 新增接口
- `GET /api/network/connections` - 获取网络连接信息
- `GET /api/network/routes` - 获取路由表信息

#### 5.2 增强现有接口
- `GET /api/docker/bridges` - 使用系统原生命令获取网桥信息
- `GET /api/interfaces` - 增强网络接口信息获取

### 6. 数据模型扩展

#### 6.1 新增模型
```go
// NetworkConnection 网络连接信息
type NetworkConnection struct {
    Protocol       string `json:"protocol"`
    LocalAddress   string `json:"local_address"`
    ForeignAddress string `json:"foreign_address"`
    State          string `json:"state"`
}

// RouteEntry 路由表条目
type RouteEntry struct {
    Destination string `json:"destination"`
    Gateway     string `json:"gateway"`
    Interface   string `json:"interface"`
    Source      string `json:"source"`
    Metric      int    `json:"metric"`
}
```

### 7. 兼容性和可靠性

#### 7.1 错误处理
- 命令执行失败时返回空结果而非错误
- 增强日志记录便于问题排查
- 优雅处理解析失败的情况

#### 7.2 性能优化
- 减少外部命令调用次数
- 缓存网络接口信息
- 异步处理大量数据

### 8. 测试验证

#### 8.1 功能测试
- 创建测试脚本验证所有API接口
- 测试系统原生命令的执行
- 验证数据解析的准确性

#### 8.2 容器环境测试
- 在Docker容器中测试所有功能
- 验证网络工具包的正确安装
- 确保权限配置正确

### 9. 预期效果

#### 9.1 解决的问题
- ✅ 消除对Docker CLI的依赖
- ✅ 提高容器化环境的兼容性
- ✅ 增强系统的可靠性和稳定性

#### 9.2 新增价值
- ✅ 提供更丰富的网络信息
- ✅ 支持实时网络状态监控
- ✅ 增强网络故障排查能力

### 10. 使用说明

#### 10.1 构建和部署
```bash
# 构建Docker镜像
make docker-build

# 运行容器
docker run -d --name iptables-manager \
  --cap-add=NET_ADMIN \
  --net=host \
  -p 8080:8080 \
  iptables-manager:latest
```

#### 10.2 功能验证
```bash
# 运行测试脚本
./test_native_network.sh
```

## 总结

通过重构网络数据获取逻辑，系统现在完全依赖系统原生网络命令，不再需要Docker CLI。这大大提高了系统在容器化环境中的兼容性和可靠性，同时还增加了更多有用的网络监控功能。