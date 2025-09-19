# IPTables 拓扑图功能

## 概述

IPTables 拓扑图功能提供了一个直观的可视化界面，用于展示 iptables 规则的表、链和规则之间的关系，以及数据包在系统中的流向路径。

## 功能特性

### 🎯 核心功能

1. **表结构可视化**
   - 显示所有 iptables 表（raw、mangle、nat、filter）
   - 展示每个表下的链结构
   - 显示重要规则的详细信息

2. **数据流路径**
   - 入站数据流：外部数据包进入系统的处理路径
   - 出站数据流：系统内部数据包向外发送的处理路径
   - 转发数据流：数据包在系统中转发的处理路径

3. **交互式界面**
   - 点击节点查看详细信息
   - 高亮显示数据流路径
   - 实时刷新拓扑数据

### 🔍 支持的 IPTables 表和链

#### Raw 表
- **PREROUTING**: 数据包进入系统时的第一个处理点
- **OUTPUT**: 本地生成数据包的处理

#### Mangle 表
- **PREROUTING**: 路由前的数据包修改
- **INPUT**: 进入本地的数据包修改
- **FORWARD**: 转发数据包的修改
- **OUTPUT**: 本地生成数据包的修改
- **POSTROUTING**: 路由后的数据包修改

#### NAT 表
- **PREROUTING**: 目标地址转换（DNAT）
- **INPUT**: 进入本地的NAT处理
- **OUTPUT**: 本地生成数据包的NAT处理
- **POSTROUTING**: 源地址转换（SNAT/MASQUERADE）

#### Filter 表
- **INPUT**: 进入本地的数据包过滤
- **FORWARD**: 转发数据包的过滤
- **OUTPUT**: 本地生成数据包的过滤

### 🎨 可视化元素

#### 节点类型
- **表节点**（蓝色）：表示 iptables 表
- **链节点**（绿色）：表示 iptables 链
- **规则节点**（橙色）：表示重要的 iptables 规则

#### 连接类型
- **实线**：表到链的包含关系
- **细线**：链到规则的包含关系
- **虚线**：跳转关系（JUMP/GOTO）

#### 数据流高亮
- **绿色路径**：入站数据流
- **蓝色路径**：出站数据流
- **橙色路径**：转发数据流

## 使用方法

### 1. 访问拓扑图

1. 登录系统后，点击左侧菜单的"拓扑图"
2. 系统将自动加载当前的 iptables 配置
3. 拓扑图将显示所有表、链和重要规则的关系

### 2. 查看数据流

1. 点击顶部的数据流按钮（入站数据流、出站数据流、转发数据流）
2. 相关的节点和连接将被高亮显示
3. 左侧面板将显示数据流的详细路径信息

### 3. 查看节点详情

1. 点击任意节点（表、链或规则）
2. 弹出详情对话框，显示节点的详细信息
3. 包括统计数据、策略信息和属性详情

### 4. 刷新数据

1. 点击"刷新"按钮重新加载 iptables 数据
2. 系统将重新分析当前的规则配置
3. 拓扑图将更新为最新状态

## API 接口

### 获取拓扑数据

```http
GET /api/topology
Authorization: Bearer <token>
```

**响应格式：**

```json
{
  "success": true,
  "data": {
    "nodes": [
      {
        "id": "table_filter",
        "label": "FILTER",
        "type": "table",
        "table_name": "filter",
        "position": {"x": 1000, "y": 100},
        "properties": {"chains": "3"}
      }
    ],
    "links": [
      {
        "id": "link_table_filter_to_chain_filter_INPUT",
        "source": "table_filter",
        "target": "chain_filter_INPUT",
        "type": "table_chain",
        "label": "contains"
      }
    ],
    "flow": [
      {
        "id": "flow_incoming",
        "name": "入站数据流",
        "description": "外部数据包进入系统的处理路径",
        "path": ["table_raw", "chain_raw_PREROUTING", "table_filter", "chain_filter_INPUT"],
        "color": "#4CAF50"
      }
    ]
  }
}
```

## 技术实现

### 后端实现

1. **TopologyService**: 分析 iptables 数据并生成拓扑结构
2. **TopologyHandler**: 提供 REST API 接口
3. **数据解析**: 解析 iptables 命令输出，提取表、链和规则信息

### 前端实现

1. **Vue 3 + TypeScript**: 现代化的前端框架
2. **SVG 渲染**: 使用 SVG 绘制拓扑图，支持缩放和交互
3. **Element Plus**: 提供丰富的 UI 组件
4. **响应式设计**: 适配不同屏幕尺寸

### 数据流分析

系统自动分析以下数据流路径：

1. **入站流量**: `RAW(PREROUTING) → MANGLE(PREROUTING) → NAT(PREROUTING) → FILTER(INPUT)`
2. **出站流量**: `RAW(OUTPUT) → MANGLE(OUTPUT) → NAT(OUTPUT) → FILTER(OUTPUT)`
3. **转发流量**: `RAW(PREROUTING) → MANGLE(FORWARD) → FILTER(FORWARD) → MANGLE(POSTROUTING) → NAT(POSTROUTING)`

## 故障排除

### 常见问题

1. **拓扑图显示为空**
   - 检查后端服务是否有权限执行 iptables 命令
   - 确保在 Docker 容器中运行时有足够的权限
   - 查看后端日志中的错误信息

2. **某些表或链不显示**
   - 检查系统是否支持相应的 iptables 表
   - 确认 iptables 模块已正确加载
   - 验证用户权限是否足够

3. **数据流高亮不正确**
   - 检查链名称是否与系统中的实际链名称匹配
   - 确认表和链的关系是否正确配置

### 调试方法

1. **使用测试脚本**
   ```bash
   chmod +x scripts/test_topology.sh
   ./scripts/test_topology.sh
   ```

2. **检查 API 响应**
   ```bash
   curl -H "Authorization: Bearer <token>" http://localhost:8080/api/topology
   ```

3. **查看后端日志**
   ```bash
   docker logs <container-name>
   ```

## 扩展功能

### 计划中的功能

1. **规则编辑**: 直接在拓扑图中编辑规则
2. **性能监控**: 显示每个链的流量统计
3. **告警系统**: 监控异常流量和规则变化
4. **导出功能**: 导出拓扑图为图片或PDF
5. **历史对比**: 比较不同时间点的拓扑结构

### 自定义扩展

1. **添加新的数据流路径**
   - 修改 `TopologyService.generateFlowPaths()` 方法
   - 定义新的流路径和颜色

2. **自定义节点样式**
   - 修改 CSS 样式文件
   - 调整节点大小、颜色和形状

3. **添加新的节点类型**
   - 扩展 `TopologyNode` 接口
   - 实现相应的渲染逻辑

## 性能优化

1. **数据缓存**: 缓存 iptables 数据，减少重复查询
2. **增量更新**: 只更新变化的部分，而不是重新渲染整个拓扑图
3. **虚拟化**: 对于大型拓扑图，使用虚拟化技术提高渲染性能
4. **懒加载**: 按需加载详细的规则信息

## 安全考虑

1. **权限控制**: 确保只有授权用户可以访问拓扑图
2. **数据脱敏**: 隐藏敏感的IP地址和端口信息
3. **审计日志**: 记录所有拓扑图访问和操作
4. **输入验证**: 验证所有用户输入，防止注入攻击

---

**版本**: 1.0.0  
**更新时间**: 2025-09-19  
**维护者**: IPTables Manager Team