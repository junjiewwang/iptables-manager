<template>
  <el-table 
    :data="rules" 
    style="width: 100%" 
    border 
    stripe
    size="small"
  >
    <el-table-column prop="line_number" label="#" width="60" />
    <el-table-column prop="chain_name" label="链" width="120" v-if="showChainColumn" />
    <el-table-column prop="table" label="表" width="80" v-if="showTableColumn">
      <template #default="scope">
        <el-tag :type="getTableTagType(scope.row.table)" size="small">
          {{ scope.row.table }}
        </el-tag>
      </template>
    </el-table-column>
    <el-table-column prop="target" label="目标" width="120">
      <template #default="scope">
        <el-tag :type="getTargetTagType(scope.row.target)">
          {{ scope.row.target }}
        </el-tag>
      </template>
    </el-table-column>
    <el-table-column prop="protocol" label="协议" width="80" />
    <el-table-column prop="source" label="源地址" width="140" />
    <el-table-column prop="destination" label="目标地址" width="140" />
    <el-table-column prop="in_interface" label="入接口" width="100" />
    <el-table-column prop="out_interface" label="出接口" width="100" />
    <el-table-column prop="options" label="选项" min-width="200" show-overflow-tooltip />
    <el-table-column label="操作" width="120" fixed="right">
      <template #default="scope">
        <el-button 
          size="small" 
          type="primary" 
          @click="onEditRule(scope.row)"
          text
        >
          编辑
        </el-button>
        <el-button 
          size="small" 
          type="danger" 
          @click="onDeleteRule(scope.row)"
          text
        >
          删除
        </el-button>
      </template>
    </el-table-column>
  </el-table>
</template>

<script setup lang="ts">
import { useTagTypes } from '@/composables/core/useTagTypes'
import type { IPTablesRule } from '../types'

interface Props {
  rules: IPTablesRule[]
  showChainColumn?: boolean
  showTableColumn?: boolean
}

interface Emits {
  (e: 'edit-rule', rule: IPTablesRule): void
  (e: 'delete-rule', rule: IPTablesRule): void
}

const props = withDefaults(defineProps<Props>(), {
  showChainColumn: false,
  showTableColumn: false
})

const emit = defineEmits<Emits>()

// 使用标签类型 composable
const { getTableTagType, getTargetTagType } = useTagTypes()

const onEditRule = (rule: IPTablesRule) => {
  emit('edit-rule', rule)
}

const onDeleteRule = (rule: IPTablesRule) => {
  emit('delete-rule', rule)
}
</script>

<style scoped>
/* 表格样式可以在这里自定义 */
</style>