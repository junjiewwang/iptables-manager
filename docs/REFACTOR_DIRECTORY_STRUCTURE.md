# 项目目录结构重构文档

## 📋 重构概述

本次重构旨在解决项目中存在的目录结构混乱问题，建立清晰的模块化目录结构，提升代码的可维护性和可读性。

## 🔍 重构前的问题

### 1. 重复的目录结构
- `src/components/` 和 `src/components/ChainTable/components/`
- `src/composables/` 和 `src/components/ChainTable/composables/`

### 2. 混乱的导入路径
- 相对路径过深：`../../../composables/useTagTypes`
- 不一致的导入方式：有些用 `@/composables`，有些用相对路径

### 3. 职责不清晰
- ChainTable 模块内部有自己的 composables，但也依赖外部 composables

## 🎯 重构方案

采用**按技术类型划分**的方案，符合 Vue 3 项目的最佳实践：

```
src/
├── components/           # 所有Vue组件
│   ├── common/          # 公共组件
│   └── ChainTable/      # ChainTable业务组件
│       ├── dialogs/     # 对话框组件
│       ├── views/       # 视图组件
│       └── nodes/       # 节点组件
├── composables/         # 所有组合式函数
│   ├── core/           # 核心逻辑
│   │   ├── useFormatters.ts
│   │   ├── useTagTypes.ts
│   │   ├── __tests__/
│   │   └── index.ts    # 导出索引
│   └── ChainTable/     # ChainTable业务逻辑
│       ├── useChainTable.ts
│       ├── useTableFilters.ts
│       ├── useTableActions.ts
│       ├── useLayout.ts
│       ├── __tests__/
│       └── index.ts    # 导出索引
├── types/              # 类型定义
│   └── ChainTable/     # ChainTable相关类型
│       ├── types.ts
│       └── index.ts    # 导出索引
└── utils/              # 工具函数
```

## 🔧 重构执行步骤

### 1. 创建新目录结构
```bash
mkdir -p src/composables/core src/composables/ChainTable src/types/ChainTable
```

### 2. 移动文件
- 移动 `ChainTable/composables/useLayout.ts` → `composables/ChainTable/`
- 移动 `ChainTable/types.ts` → `types/ChainTable/`
- 移动核心 composables 到 `composables/core/`
- 移动 ChainTable 业务 composables 到 `composables/ChainTable/`

### 3. 更新配置文件
- 更新 `vite.config.ts` 添加路径别名
- 更新 `tsconfig.json` 添加路径映射

### 4. 更新导入路径
- 统一使用 `@/` 别名导入
- 更新所有组件中的导入路径
- 更新测试文件中的导入路径

### 5. 创建索引文件
- 为每个模块创建 `index.ts` 文件
- 简化导入语句

## 📊 重构效果

### ✅ 解决的问题

1. **消除重复目录结构**
   - 移除了 `components/ChainTable/composables/`
   - 统一了目录结构

2. **简化导入路径**
   - 统一使用 `@/` 别名
   - 消除了深层相对路径

3. **清晰的职责划分**
   - `composables/core/` - 核心通用逻辑
   - `composables/ChainTable/` - ChainTable 业务逻辑
   - `types/ChainTable/` - ChainTable 类型定义

### 📈 改进效果

1. **可维护性提升**
   - 清晰的模块边界
   - 一致的导入方式
   - 更好的代码组织

2. **开发体验改善**
   - IDE 自动补全更准确
   - 更容易找到相关文件
   - 减少导入错误

3. **测试覆盖**
   - 所有测试文件正常运行
   - 构建过程无错误

## 🔄 路径映射配置

### vite.config.ts
```typescript
resolve: {
  alias: {
    '@': resolve(__dirname, 'src'),
    '@/components': resolve(__dirname, 'src/components'),
    '@/composables': resolve(__dirname, 'src/composables'),
    '@/types': resolve(__dirname, 'src/types'),
    '@/utils': resolve(__dirname, 'src/utils')
  }
}
```

### tsconfig.json
```json
"paths": {
  "@/*": ["src/*"],
  "@/components/*": ["src/components/*"],
  "@/composables/*": ["src/composables/*"],
  "@/types/*": ["src/types/*"],
  "@/utils/*": ["src/utils/*"]
}
```

## 📝 导入示例

### 重构前
```typescript
// 混乱的导入路径
import { useTagTypes } from '../../../composables/useTagTypes'
import type { IPTablesRule } from '../components/ChainTable/types'
import { useLayout } from '../composables/useLayout'
```

### 重构后
```typescript
// 清晰的导入路径
import { useTagTypes } from '@/composables/core'
import type { IPTablesRule } from '@/types/ChainTable'
import { useLayout } from '@/composables/ChainTable'
```

## ✅ 验证结果

- ✅ 构建成功：`npm run build`
- ✅ 测试通过：22/22 测试用例通过
- ✅ 类型检查：无 TypeScript 错误
- ✅ 功能完整：所有原有功能保持不变

## 🎉 总结

本次重构成功解决了项目目录结构混乱的问题，建立了清晰的模块化架构。通过统一的路径别名和合理的目录划分，显著提升了代码的可维护性和开发体验。

重构过程中严格遵循了以下原则：
- 不破坏现有功能
- 保持向后兼容
- 提升代码质量
- 改善开发体验

所有变更都经过了充分的测试验证，确保项目的稳定性和可靠性。