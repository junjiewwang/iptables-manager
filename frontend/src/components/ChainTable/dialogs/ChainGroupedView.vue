<template>
  <el-collapse accordion>
    <el-collapse-item
      v-for="(rules, chainName) in groupedRules"
      :key="chainName"
      :title="chainName"
    >
      <template #title>
        <div class="chain-group-header">
          <el-tag :type="getChainTagType(chainName)" effect="dark">
            {{ chainName }}
          </el-tag>
          <span class="chain-group-count">
            {{ rules.length }} 条规则
          </span>
          <div class="chain-group-tables">
            <el-tag
              v-for="table in getTablesInGroup(rules)"
              :key="table"
              :type="getTableTagType(table)"
              size="small"
              effect="plain"
            >
              {{ table }}
            </el-tag>
          </div>
        </div>
      </template>
      
      <RulesTable
        :rules="rules"
        :show-table-column="true"
        @edit-rule="onEditRule"
        @delete-rule="onDeleteRule"
      />
    </el-collapse-item>
  </el-collapse>
</template>

<script setup lang="ts">
import { useTagTypes } from '@/composables/core/useTagTypes'
import RulesTable from '@/components/common/RulesTable.vue'
import type { IPTablesRule } from '@/types/ChainTable/types'

interface Props {
  groupedRules: Record<string, IPTablesRule[]>
}

interface Emits {
  (e: 'edit-rule', rule: IPTablesRule): void
  (e: 'delete-rule', rule: IPTablesRule): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// 使用标签类型 composable
const { getChainTagType, getTableTagType, getTablesInGroup } = useTagTypes()

const onEditRule = (rule: IPTablesRule) => {
  emit('edit-rule', rule)
}

const onDeleteRule = (rule: IPTablesRule) => {
  emit('delete-rule', rule)
}
</script>

<style scoped>
.chain-group-header {
  display: flex;
  align-items: center;
  gap: 10px;
}

.chain-group-count {
  color: #909399;
  font-size: 0.9em;
}

.chain-group-tables {
  display: flex;
  gap: 5px;
}
</style>