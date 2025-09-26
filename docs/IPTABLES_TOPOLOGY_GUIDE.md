# IPTables链路拓扑图使用指南

## 功能概述

IPTables链路拓扑图是一个基于Vue3+Vue Flow实现的可视化工具，用于直观展示网络数据包在IPTables防火墙中的流转路径和规则处理流程。

## 核心功能特性

### 1. 完整的IPTables五链架构展示

- **PREROUTING链**: 包含raw、mangle、nat表的执行顺序
- **INPUT链**: 处理发往本机的数据包
- **FORWARD链**: 处理转发的数据包  
- **OUTPUT链**: 处理本机发出的数据包
- **POSTROUTING链**: 包含mangle、nat表的最终处理

### 2. 交互式可视化操作

#### 视图控制
- ✅ **画布缩放**: 鼠标滚轮或控制面板按钮
- ✅ **自由拖拽**: 空格+拖动进行画布平移
- ✅ **节点拖拽**: 直接拖拽节点调整位置
- ✅ **一键复位**: 重置视图到初始状态
- ✅ **适应画布**: 自动调整视图以显示所有内容

#### 数据流高亮
- **转发流量**: 外部网络 → PREROUTING → 路由决策 → FORWARD → POSTROUTING → 内部网络
- **本地入站**: 外部网络 → PREROUTING → 路由决策 → INPUT → 本地进程
- **本地出站**: 本地进程 → OUTPUT → POSTROUTING → 内部网络

### 3. 节点类型与样式

#### 网络接口节点 (圆形)
- **外部网络**: 红色边框，表示外部网络接口
- **内部网络**: 绿色边框，表示内部网络接口
- 支持悬停显示详细信息

#### IPTables链节点 (矩形)
- **PREROUTING**: 橙色主题，显示raw→mangle→nat处理顺序
- **INPUT**: 绿色主题，显示mangle→filter处理顺序
- **FORWARD**: 红色主题，显示mangle→filter处理顺序
- **OUTPUT**: 蓝色主题，显示raw→mangle→nat→filter处理顺序
- **POSTROUTING**: 紫色主题，显示mangle→nat处理顺序

#### 特殊节点
- **路由决策**: 菱形节点，表示内核路由判断点
- **本地进程**: 紫色圆角矩形，表示用户空间进程

### 4. 连接边与动画

#### 边的类型
- **实线箭头**: 常规数据流路径
- **动画效果**: 选中特定数据流时的动态展示
- **颜色编码**: 
  - 蓝色: 通用数据流
  - 绿色: INPUT路径
  - 红色: FORWARD路径
  - 蓝色: OUTPUT路径

#### 边标签
- 显示流量类型描述
- 半透明背景，便于阅读
- 自动定位在连接线中点

## 使用方法

### 1. 基本操作

```typescript
// 选择数据流类型
selectedFlow.value = 'forward'  // 转发流量
selectedFlow.value = 'input'    // 本地入站
selectedFlow.value = 'output'   // 本地出站
```

### 2. 视图控制

- **缩放**: 使用鼠标滚轮或右侧控制面板的+/-按钮
- **平移**: 按住空格键并拖拽画布
- **重置**: 点击"重置视图"按钮
- **适应**: 点击"适应画布"按钮自动调整视图

### 3. 节点交互

- **点击节点**: 显示详细信息面板
- **拖拽节点**: 重新排列节点位置
- **悬停节点**: 显示快速信息提示

### 4. 过滤功能

- **协议过滤**: 按TCP/UDP/ICMP等协议类型过滤
- **端口过滤**: 按特定端口号过滤显示

## 技术实现

### 核心技术栈
- **Vue 3**: 响应式框架
- **Vue Flow**: 流程图可视化库
- **TypeScript**: 类型安全
- **Element Plus**: UI组件库

### 关键组件
```vue
<VueFlow
  v-model="flowElements"
  :default-viewport="{ zoom: 0.8 }"
  :min-zoom="0.3"
  :max-zoom="2"
  :snap-to-grid="true"
  :snap-grid="[20, 20]"
  :fit-view-on-init="true"
>
  <Background pattern-color="#e2e8f0" :gap="20" />
  <Controls />
  <MiniMap />
</VueFlow>
```

### 自定义节点模板
- `node-chain`: IPTables链节点
- `node-interface`: 网络接口节点  
- `node-decision`: 路由决策节点
- `node-process`: 本地进程节点

## 最佳实践

### 1. 性能优化
- 使用计算属性缓存节点和边的计算结果
- 合理使用Vue Flow的内置优化功能
- 避免频繁的DOM操作

### 2. 用户体验
- 提供清晰的视觉反馈
- 保持一致的交互模式
- 合理的动画时长和缓动效果

### 3. 可访问性
- 支持键盘导航
- 提供替代文本描述
- 合理的颜色对比度

## 扩展功能

### 1. 数据导出
```typescript
const exportTopology = () => {
  const data = {
    nodes: flowElements.value.filter(el => 'type' in el),
    edges: flowElements.value.filter(el => 'source' in el),
    timestamp: new Date().toISOString()
  }
  // 导出为JSON格式
}
```

### 2. 自定义主题
- 支持深色/浅色主题切换
- 可配置的颜色方案
- 自定义节点样式

### 3. 实时数据更新
- WebSocket连接实时数据
- 自动刷新机制
- 增量更新优化

## 故障排除

### 常见问题

1. **节点重叠**: 使用自动布局功能或手动调整节点位置
2. **性能问题**: 减少同时显示的节点数量，使用分页或过滤
3. **样式异常**: 检查CSS样式优先级和Vue Flow主题配置

### 调试技巧

```typescript
// 开启调试模式
console.log('Flow elements:', flowElements.value)
console.log('Selected flow:', selectedFlow.value)
```

## 更新日志

### v1.0.0 (2024-01-XX)
- ✅ 基础IPTables五链架构展示
- ✅ 交互式节点拖拽和缩放
- ✅ 数据流高亮动画
- ✅ 响应式设计支持
- ✅ 导出功能

### 计划功能
- 🔄 实时规则监控
- 🔄 性能指标展示
- 🔄 多语言支持
- 🔄 自定义布局算法

---

*本文档持续更新中，如有问题请提交Issue或Pull Request。*