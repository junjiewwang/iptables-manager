# ChainFlowView 自动布局功能

## 🎯 功能概述

ChainFlowView 组件现在集成了强大的自动布局功能，基于 dagre 算法实现智能的节点排列和优化的视觉效果。

## ✨ 新增功能

### 1. 自动布局算法
- **垂直布局 (TB)**: 从上到下的流程布局，适合展示数据包处理流程
- **水平布局 (LR)**: 从左到右的流程布局，适合宽屏显示
- **智能间距**: 自动计算最优的节点和层级间距
- **动画过渡**: 平滑的布局切换动画效果

### 2. 布局控制面板
- **布局方向切换**: 一键切换垂直/水平布局
- **视图适配**: 自动调整缩放以适应屏幕
- **布局重置**: 恢复到初始布局状态
- **间距调整**: 实时调整节点间距和层级间距
- **动画开关**: 控制布局切换时的动画效果

### 3. 响应式设计
- **移动端适配**: 在小屏幕设备上优化显示
- **触摸友好**: 支持触摸操作和手势
- **自适应缩放**: 根据屏幕尺寸自动调整

## 🛠️ 技术实现

### 核心组件

#### useLayout Composable
```typescript
// 使用示例
const { layout, relayout } = useLayout()

// 应用布局
const layoutedNodes = layout(nodes, edges, 'TB')

// 重新布局
const newNodes = relayout(nodes, edges, 'LR')
```

#### LayoutPanel 组件
```vue
<LayoutPanel
  :current-direction="layoutDirection"
  :node-spacing="80"
  :rank-spacing="120"
  :animate-layout="true"
  @layout-change="onLayoutChange"
  @fit-view="onFitView"
  @reset-layout="onResetLayout"
/>
```

### 布局算法参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `rankdir` | 'TB' | 布局方向 (TB/LR/BT/RL) |
| `nodesep` | 80 | 同层节点间距 |
| `ranksep` | 120 | 不同层级间距 |
| `marginx` | 50 | 水平边距 |
| `marginy` | 50 | 垂直边距 |

### 节点尺寸配置

```typescript
// 根据节点类型设置不同尺寸
const nodeSizes = {
  chain: { width: 200, height: 120 },
  decision: { width: 160, height: 100 },
  network: { width: 140, height: 80 },
  process: { width: 180, height: 80 }
}
```

## 🎨 视觉效果

### 动画效果
- **节点过渡**: 0.3s 缓动动画
- **悬停效果**: 节点上浮和阴影增强
- **边动画**: 悬停时边线加粗
- **布局切换**: 渐进式节点位置更新

### 样式特性
- **现代UI**: 毛玻璃效果的控制面板
- **响应式**: 适配不同屏幕尺寸
- **主题一致**: 与整体设计风格保持一致

## 📱 使用方法

### 基本使用
```vue
<template>
  <ChainFlowView
    :flow-elements="flowElements"
    :topo-settings="topoSettings"
    @node-click="handleNodeClick"
    @select-chain-table="handleChainSelect"
  />
</template>
```

### 事件处理
```typescript
// 节点点击事件
const handleNodeClick = (event) => {
  console.log('节点点击:', event.node)
}

// 链表选择事件
const handleChainSelect = (tableName, chainName) => {
  console.log(`选择了表: ${tableName}, 链: ${chainName}`)
}
```

## 🔧 自定义配置

### 布局参数调整
```typescript
// 在 useLayout 中自定义参数
const layoutConfig = {
  rankdir: 'TB',
  nodesep: 100,    // 增加节点间距
  ranksep: 150,    // 增加层级间距
  marginx: 60,     // 增加水平边距
  marginy: 60      // 增加垂直边距
}
```

### 动画配置
```css
/* 自定义动画时长 */
.vue-flow__node {
  transition: all 0.5s cubic-bezier(0.4, 0, 0.2, 1);
}
```

## 🚀 性能优化

### 渲染优化
- **按需更新**: 只更新位置变化的节点
- **动画分批**: 避免同时更新所有节点
- **内存管理**: 及时清理不需要的引用

### 最佳实践
1. **合理设置间距**: 避免节点重叠
2. **选择合适布局**: 根据内容选择垂直或水平布局
3. **启用动画**: 提升用户体验
4. **响应式适配**: 考虑移动端使用场景

## 🐛 故障排除

### 常见问题

#### 1. 布局不生效
```typescript
// 确保在节点初始化后应用布局
onMounted(() => {
  nextTick(() => {
    applyLayout('TB')
  })
})
```

#### 2. 动画卡顿
```typescript
// 减少动画复杂度
layoutSettings.animateLayout = false
```

#### 3. 节点重叠
```typescript
// 增加节点间距
layoutSettings.nodeSpacing = 120
layoutSettings.rankSpacing = 180
```

## 📈 未来规划

### 计划功能
- [ ] 更多布局算法支持 (力导向、圆形等)
- [ ] 布局模板保存和加载
- [ ] 自定义布局约束
- [ ] 性能监控和优化工具
- [ ] 布局历史记录和撤销功能

### 改进方向
- 更智能的自动布局
- 更丰富的动画效果
- 更好的移动端体验
- 更强的自定义能力

---

## 📞 技术支持

如有问题或建议，请联系开发团队或提交 Issue。