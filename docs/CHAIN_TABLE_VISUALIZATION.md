# IPTables 五链四表可视化功能

## 功能概述

本功能提供了一个直观的可视化界面，用于展示iptables的五链（PREROUTING、INPUT、FORWARD、OUTPUT、POSTROUTING）与四表（raw、mangle、nat、filter）之间的关系，并支持按网络接口维度进行数据聚合和分析。

## 主要特性

### 1. 五链顺序展示
- 按照数据包在iptables中的处理顺序展示五个链：
  - **PREROUTING**: 数据包进入系统后的第一个处理点
  - **INPUT**: 目标为本机的数据包处理
  - **FORWARD**: 需要转发的数据包处理
  - **OUTPUT**: 本机发出的数据包处理
  - **POSTROUTING**: 数据包离开系统前的最后处理点

### 2. 四表关系展示
- 清晰展示四个表在每个链中的作用：
  - **raw表**: 连接跟踪处理
  - **mangle表**: 数据包修改
  - **nat表**: 网络地址转换
  - **filter表**: 数据包过滤

### 3. 网络接口维度聚合
- 支持按网络接口过滤和查看规则
- 展示每个网络接口相关的iptables规则统计
- 清晰显示FORWARD链中数据包的转发路径

### 4. 多种视图模式
- **链视图**: 以链为主线，展示每个链中各表的规则分布
- **表视图**: 以表为主线，展示每个表中各链的规则分布
- **接口视图**: 以网络接口为主线，展示接口相关的规则统计

## 使用方法

### 访问页面
1. 登录系统后，在左侧导航栏中点击"五链四表"
2. 页面将自动加载当前系统的iptables规则数据

### 功能操作
1. **选择网络接口**: 使用顶部的接口选择器过滤特定接口的规则
2. **切换视图模式**: 使用视图模式选择器在不同视图间切换
3. **查看详细规则**: 点击链或表卡片查看详细的规则列表
4. **刷新数据**: 点击刷新按钮获取最新的iptables规则

### 视图说明

#### 链视图
- 水平排列显示五个链的处理顺序
- 每个链卡片显示该链中各表的规则数量
- 不同表使用不同颜色标识：
  - raw表: 蓝色
  - mangle表: 黄色
  - nat表: 绿色
  - filter表: 红色

#### 表视图
- 网格布局显示四个表
- 每个表卡片显示该表中各链的规则分布
- 显示链的策略（ACCEPT/DROP等）

#### 接口视图
- 显示所有网络接口及其基本信息
- 统计每个接口相关的规则数量：
  - 输入规则：以该接口为入接口的规则
  - 输出规则：以该接口为出接口的规则
  - 转发规则：在FORWARD链中涉及该接口的规则

## 技术实现

### 前端组件
- **文件位置**: `frontend/src/views/ChainTableView.vue`
- **API接口**: `frontend/src/api/chainTable.ts`
- **主要技术**: Vue 3 + TypeScript + Element Plus

### 后端API
- **处理器**: `backend/handlers/chain_table_handler.go`
- **主要接口**:
  - `GET /api/chain-table-data`: 获取五链四表聚合数据
  - `GET /api/network/interfaces/:name/rules`: 获取指定接口的规则统计

### 数据结构

#### 链数据结构
```go
type ChainData struct {
    Name   string                   `json:"name"`
    Tables []ChainTableData         `json:"tables"`
    Rules  []models.IPTablesRule    `json:"rules"`
}
```

#### 表数据结构
```go
type TableData struct {
    Name       string      `json:"name"`
    TotalRules int         `json:"total_rules"`
    Chains     []ChainInfo `json:"chains"`
}
```

## 数据流向说明

### 数据包处理流程
1. **PREROUTING**: 数据包进入系统，进行路由决策前的处理
   - raw表：连接跟踪标记
   - mangle表：数据包标记和修改
   - nat表：目标地址转换(DNAT)

2. **路由决策**: 决定数据包是发往本机还是转发

3. **INPUT** (发往本机的数据包):
   - mangle表：数据包修改
   - filter表：数据包过滤

4. **FORWARD** (需要转发的数据包):
   - mangle表：数据包修改
   - filter表：转发规则检查

5. **OUTPUT** (本机发出的数据包):
   - raw表：连接跟踪处理
   - mangle表：数据包修改
   - nat表：源地址转换(SNAT)
   - filter表：数据包过滤

6. **POSTROUTING**: 数据包离开系统前的最后处理
   - mangle表：最后的数据包修改
   - nat表：源地址转换(SNAT/MASQUERADE)

## 注意事项

1. **权限要求**: 需要系统管理员权限才能读取iptables规则
2. **实时性**: 数据不是实时更新的，需要手动刷新获取最新状态
3. **兼容性**: 支持标准的iptables规则格式，对于自定义链可能显示有限
4. **性能**: 在规则数量较多时，建议使用接口过滤功能减少数据量

## 故障排除

### 常见问题
1. **数据加载失败**: 检查后端服务是否正常运行，iptables命令是否可用
2. **规则显示不完整**: 确认用户权限，某些规则可能需要root权限才能查看
3. **接口信息不准确**: 网络接口状态可能发生变化，建议刷新数据

### 调试模式
- 开启浏览器开发者工具查看网络请求
- 检查后端日志获取详细错误信息
- 使用模拟数据模式进行功能测试