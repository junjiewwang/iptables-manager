# Linux 网络数据包处理流程拓扑图

## 概述

本项目实现了一个可视化的Linux网络数据包处理流程拓扑图，展示了iptables五链四表的完整处理流程。

## 功能特性

### 🎯 核心功能
- **完整流程展示**: 展示从外部网络到内部网络的完整数据包处理流程
- **五链四表可视化**: PREROUTING、INPUT、FORWARD、OUTPUT、POSTROUTING五链与raw、mangle、nat、filter四表
- **交互式节点**: 点击链节点可查看详细规则信息
- **实时规则统计**: 显示每个链的实时规则数量
- **流向标注**: 清晰的箭头和标签显示数据包流向

### 🎨 视觉设计
- **节点颜色编码**:
  - 🔴 PREROUTING: 红色 - 预路由处理
  - 🟢 INPUT/OUTPUT: 绿色 - 本机输入输出
  - 🔵 FORWARD: 蓝色 - 转发处理
  - 🟡 POSTROUTING: 黄色 - 路由后处理
  - 🟠 路由决策: 橙色 - 路由判断
  - 🟣 本地处理: 紫色 - 应用程序处理
  - ⚪ 网络节点: 白色 - 外部/内部网络

- **表标签颜色**:
  - `raw`: 红色标签
  - `mangle`: 黄色标签  
  - `nat`: 蓝色标签
  - `filter`: 绿色标签

## 数据包处理流程

```
外部网络 
    ↓ (入站数据包)
PREROUTING (raw, mangle, nat - 3条规则)
    ↓ (预路由处理)
路由决策
    ├─→ INPUT (mangle, filter, nat - 0条规则)
    │       ↓ (输入过滤)
    │   本地处理
    │       ↓ (本地响应)
    │   OUTPUT (raw, mangle, nat, filter - 2条规则)
    │       ↓ (出站数据包)
    └─→ FORWARD (mangle, filter - 41条规则)
            ↓ (转发数据包)
        POSTROUTING (mangle, nat - 14条规则)
            ↓ (路由后处理)
        内部网络
```

## 组件架构

### 主要组件

1. **ChainFlowView.vue** - 主流程图组件
   - 使用Vue Flow库实现拓扑图
   - 定义节点和边的数据结构
   - 处理用户交互事件

2. **节点组件**:
   - `NetworkNode.vue` - 网络节点（外部/内部网络）
   - `ChainProcessNode.vue` - 链处理节点（五链）
   - `DecisionNode.vue` - 决策节点（路由决策）
   - `ProcessNode.vue` - 处理节点（本地处理）

3. **边组件**:
   - `FlowEdge.vue` - 自定义流程边，支持标签显示

### 技术栈

- **Vue 3** - 前端框架
- **TypeScript** - 类型安全
- **Vue Flow** - 流程图库
- **Element Plus** - UI组件库
- **Vite** - 构建工具

## 使用方法

### 基本使用

```vue
<template>
  <ChainFlowView
    :flowElements="flowElements"
    :topoSettings="topoSettings"
    @select-chain-table="onSelectChainTable"
    @node-click="onNodeClick"
  />
</template>

<script setup>
import ChainFlowView from './views/ChainFlowView.vue'

const onSelectChainTable = (tableName, chainName) => {
  console.log(`选择了表: ${tableName}, 链: ${chainName}`)
}

const onNodeClick = (event) => {
  console.log('节点点击:', event)
}
</script>
```

### 节点数据结构

```typescript
interface ChainNode {
  id: string
  type: 'chain'
  position: { x: number, y: number }
  data: {
    label: string           // 显示名称
    chainName: string       // 链名称
    tables: string[]        // 关联的表
    ruleCount: number       // 规则数量
    color: string          // 背景色
    borderColor: string    // 边框色
  }
}
```

### 边数据结构

```typescript
interface FlowEdge {
  id: string
  source: string          // 源节点ID
  target: string          // 目标节点ID
  type: 'flow'
  data: {
    label: string         // 边标签
  }
  style: {
    stroke: string        // 线条颜色
    strokeDasharray: string // 虚线样式
  }
  animated?: boolean      // 是否动画
}
```

## 交互功能

### 节点交互
- **点击链节点**: 查看该链的详细规则
- **点击表标签**: 查看特定表的规则
- **悬停效果**: 节点高亮和阴影效果

### 控制功能
- **缩放**: 支持鼠标滚轮缩放
- **拖拽**: 支持节点拖拽调整位置
- **小地图**: 显示整体布局导航
- **控制面板**: 缩放、居中等操作

## 自定义配置

### 拓扑设置

```typescript
interface TopoSettings {
  darkMode: boolean        // 暗色模式
  snapToGrid: boolean      // 网格对齐
  enableDrag: boolean      // 启用拖拽
  animateEdges: boolean    // 边动画
  showMinimap: boolean     // 显示小地图
}
```

### 样式自定义

可以通过CSS变量自定义节点样式：

```css
.chain-process-node {
  --node-bg-color: #f0f0f0;
  --node-border-color: #d0d0d0;
  --node-text-color: #333;
}
```

## 开发指南

### 添加新节点类型

1. 创建节点组件文件
2. 在ChainFlowView中注册节点模板
3. 更新节点数据结构

### 添加新的交互功能

1. 在节点组件中添加事件处理
2. 通过emit向父组件传递事件
3. 在ChainFlowView中处理事件

## 性能优化

- **懒加载**: 大型拓扑图支持节点懒加载
- **虚拟化**: 大量节点时使用虚拟滚动
- **缓存**: 节点渲染结果缓存
- **防抖**: 交互事件防抖处理

## 浏览器兼容性

- Chrome 88+
- Firefox 85+
- Safari 14+
- Edge 88+

## 许可证

MIT License