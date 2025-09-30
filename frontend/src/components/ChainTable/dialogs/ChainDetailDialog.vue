<template>
  <el-dialog
    :value="showDialog"
    @input="$emit('update:showDialog', $event)"
    :title="title"
    width="90%"
    destroy-on-close
    @closed="onClosed"
  >
    <div class="chain-detail-controls">
      <el-radio-group 
        :value="groupByChain"
        @input="$emit('update:groupByChain', $event)" 
        size="small"
      >
        <el-radio-button :label="true">按链分组</el-radio-button>
        <el-radio-button :label="false">列表视图</el-radio-button>
      </el-radio-group>
      
      <ChainDetailFilters
        :rule-search-text="ruleSearchText"
        :table-filter="tableFilter"
        :target-filter="targetFilter"
        @update:rule-search-text="$emit('update:ruleSearchText', $event)"
        @update:table-filter="$emit('update:tableFilter', $event)"
        @update:target-filter="$emit('update:targetFilter', $event)"
      />
      
      <el-button type="primary" size="small" @click="onAddRule(selectedChain)">
        添加规则
      </el-button>
    </div>
    
    <!-- 分组视图 -->
    <ChainGroupedView
      v-if="groupByChain && Object.keys(groupedRules).length > 0"
      :grouped-rules="groupedRules"
      @edit-rule="onEditRule"
      @delete-rule="onDeleteRule"
    />
    
    <!-- 列表视图 -->
    <template v-else>
      <el-empty v-if="filteredDetailRules.length === 0" description="无匹配规则" />
      <RulesTable
        v-else
        :rules="filteredDetailRules"
        :show-chain-column="true"
        :show-table-column="true"
        @edit-rule="onEditRule"
        @delete-rule="onDeleteRule"
      />
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import ChainDetailFilters from './ChainDetailFilters.vue'
import ChainGroupedView from './ChainGroupedView.vue'
import RulesTable from '@/components/common/RulesTable.vue'
import type { IPTablesRule } from '@/types/ChainTable/types'

interface Props {
  showDialog: boolean
  title: string
  selectedChain: string
  groupByChain: boolean
  ruleSearchText: string
  tableFilter: string
  targetFilter: string
  groupedRules: Record<string, IPTablesRule[]>
  filteredDetailRules: IPTablesRule[]
}

interface Emits {
  (e: 'update:showDialog', value: boolean): void
  (e: 'update:groupByChain', value: boolean): void
  (e: 'update:ruleSearchText', value: string): void
  (e: 'update:tableFilter', value: string): void
  (e: 'update:targetFilter', value: string): void
  (e: 'closed'): void
  (e: 'edit-rule', rule: IPTablesRule): void
  (e: 'delete-rule', rule: IPTablesRule): void
  (e: 'add-rule', chainName: string): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const onClosed = () => {
  emit('closed')
}

const onEditRule = (rule: IPTablesRule) => {
  emit('edit-rule', rule)
}

const onDeleteRule = (rule: IPTablesRule) => {
  emit('delete-rule', rule)
}

const onAddRule = (chainName: string) => {
  emit('add-rule', chainName)
}
</script>

<style scoped>
.chain-detail-controls {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

@media (max-width: 768px) {
  .chain-detail-controls {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }
}
</style>