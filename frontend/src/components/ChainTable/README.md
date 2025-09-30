# MainTable 组件重构文档

## 重构概述

本次重构将原本复杂的 MainTable 组件按照 Vue 项目规范进行了模块化拆分，采用了组合式 API、工厂模式和状态管理模式，提高了代码的可维护性、可测试性和可复用性。

## 架构设计

### 1. 组件拆分架构

```
MainTable (主组件)
├── views/ (视图组件)
│   ├── ChainFlowView.vue (数据流图视图)
│   ├── TableView.vue (表格视图)
│   ├── InterfaceView.vue (接口视图)
│   ├── InterfaceStatsPanel.vue (接口统计面板)
│   ├── InterfaceCard.vue (接口卡片)
│   ├── InterfaceBasicInfo.vue (接口基本信息)
│   ├── InterfaceNetworkInfo.vue (接口网络信息)
│   ├── InterfaceRulesStats.vue (接口规则统计)
│   ├── InterfaceTrafficStats.vue (接口流量统计)
│   └── nodes/ (节点组件)
│       ├── ChainNode.vue (链节点)
│       ├── DecisionNode.vue (决策节点)
│       ├── EndpointNode.vue (端点节点)
│       └── ProcessNode.vue (处理节点)
├── dialogs/ (对话框组件)
│   ├── ChainDetailDialog.vue (链详情对话框)
│   ├── ChainDetailFilters.vue (链详情筛选)
│   └── ChainGroupedView.vue (链分组视图)
└── common/ (通用组件)
    └── RulesTable.vue (规则表格)
```

### 2. Composables 架构

```
composables/
├── useTagTypes.ts (标签类型管理)
├── useFormatters.ts (格式化工具)
├── useMainTableEvents.ts (事件处理)
└── useMainTableStore.ts (状态管理)
```

### 3. 工具类架构

```
utils/
└── tableFactory.ts (表格工厂)
```

## 设计模式应用

### 1. 组合式 API 模式

所有组件都采用了 `<script setup>` 语法，使用组合式 API 实现逻辑复用：

```typescript
// 使用 composables 实现逻辑复用
const { getTableTagType, getTargetTagType } = useTagTypes()
const { formatBytes } = useFormatters()
const events = useMainTableEvents(emit)
```

### 2. 工厂模式

使用工厂模式创建不同类型的表格配置：

```typescript
// 创建基础规则表格
const basicConfig = TableItemFactory.createRulesTableConfig()

// 创建详细规则表格（包含链和表列）
const detailConfig = TableItemFactory.createDetailRulesTableConfig()

// 使用构建器模式创建自定义表格
const customConfig = new TableConfigBuilder()
  .addColumn({ prop: 'name', label: '名称' })
  .setOptions({ stripe: true })
  .build()
```

### 3. 状态管理模式

采用类似 Pinia 的状态管理模式：

```typescript
const store = useMainTableStore()

// 状态访问
const { state, filteredDetailRules, groupedRules } = store

// 状态修改
store.setViewMode('table')
store.showChainDetail('INPUT', '输入链详情', rules)
```

## 组件职责划分

### 主组件 (MainTable.vue)
- 作为容器组件，负责组合各个子组件
- 处理组件间的数据传递和事件通信
- 维护组件的整体状态

### 视图组件
- **ChainFlowView**: 负责数据流图的渲染和交互
- **TableView**: 负责表格视图的展示
- **InterfaceView**: 负责接口视图的展示

### 对话框组件
- **ChainDetailDialog**: 负责链详情的展示和交互
- **ChainDetailFilters**: 负责筛选条件的管理
- **ChainGroupedView**: 负责分组视图的展示

### 通用组件
- **RulesTable**: 可复用的规则表格组件

## Composables 功能

### useTagTypes
- 统一管理各种标签的类型映射
- 提供标签类型获取方法
- 支持表格、目标、链等不同类型的标签

### useFormatters
- 提供各种数据格式化功能
- 字节数格式化、数字格式化等
- 统一的格式化标准

### useMainTableEvents
- 统一管理组件事件处理逻辑
- 提供类型安全的事件处理方法
- 简化事件传递流程

### useMainTableStore
- 集中式状态管理
- 提供响应式状态和计算属性
- 包含完整的状态操作方法

## 类型安全

所有组件都使用 TypeScript 进行了完整的类型定义：

```typescript
interface Props {
  viewMode: ViewMode
  tables: TableInfo[]
  // ...
}

interface Emits {
  (e: 'edit-rule', rule: IPTablesRule): void
  // ...
}
```

## 性能优化

### 1. 组件懒加载
- 大型组件按需加载
- 减少初始包体积

### 2. 计算属性缓存
- 使用 computed 进行数据计算缓存
- 避免不必要的重复计算

### 3. 事件处理优化
- 统一的事件处理机制
- 避免重复的事件监听器

## 测试策略

### 1. 单元测试
- 每个 composable 都可以独立测试
- 工厂类的测试覆盖
- 组件逻辑的单元测试

### 2. 组件测试
- 子组件的独立测试
- Props 和 Events 的测试
- 用户交互的测试

### 3. 集成测试
- 组件间通信的测试
- 状态管理的集成测试
- 端到端的功能测试

## 使用示例

### 基本使用

```vue
<template>
  <MainTable
    :view-mode="viewMode"
    :tables="tables"
    :interfaces="interfaces"
    @edit-rule="handleEditRule"
    @delete-rule="handleDeleteRule"
  />
</template>

<script setup lang="ts">
import MainTable from './components/ChainTable/MainTable.vue'

const viewMode = ref('table')
const tables = ref([])
const interfaces = ref([])

const handleEditRule = (rule) => {
  // 处理编辑规则
}

const handleDeleteRule = (rule) => {
  // 处理删除规则
}
</script>
```

### 使用状态管理

```typescript
const store = useMainTableStore()

// 更新数据
store.updateTables(newTables)
store.updateInterfaces(newInterfaces)

// 显示详情
store.showChainDetail('INPUT', '输入链', rules)

// 设置筛选
store.setRuleSearchText('tcp')
store.setTableFilter('filter')
```

### 使用工厂模式

```typescript
// 创建表格配置
const tableConfig = TableItemFactory.createRulesTableConfig()

// 添加操作列
const configWithActions = TableItemFactory.addActionColumn(tableConfig, [
  { label: '编辑', type: 'primary', handler: 'edit' },
  { label: '删除', type: 'danger', handler: 'delete' }
])
```

## 迁移指南

### 从旧版本迁移

1. **组件引用更新**: 更新组件的导入路径
2. **Props 调整**: 检查并更新组件的 props 定义
3. **事件处理**: 更新事件处理器的命名和参数
4. **样式调整**: 检查样式是否需要调整

### 兼容性说明

- 保持了原有的 API 接口兼容性
- 事件名称和参数保持不变
- 主要的 props 结构保持一致

## 最佳实践

### 1. 组件设计
- 遵循单一职责原则
- 保持组件的纯净性
- 合理使用 props 和 events

### 2. 状态管理
- 使用集中式状态管理
- 避免 prop drilling
- 保持状态的不可变性

### 3. 类型安全
- 完整的 TypeScript 类型定义
- 使用泛型提高代码复用性
- 严格的类型检查

### 4. 性能考虑
- 合理使用 computed 和 watch
- 避免不必要的响应式数据
- 优化大列表的渲染

## 总结

本次重构显著提升了代码的：
- **可维护性**: 清晰的组件结构和职责划分
- **可测试性**: 独立的组件和逻辑单元
- **可复用性**: 通用的 composables 和工具类
- **类型安全**: 完整的 TypeScript 支持
- **性能**: 优化的渲染和状态管理

重构后的架构更符合 Vue 3 的最佳实践，为后续的功能扩展和维护提供了良好的基础。