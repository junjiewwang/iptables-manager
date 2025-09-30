<template>
  <div class="table-view">
    <el-tabs type="border-card">
      <el-tab-pane v-for="table in tables" :key="table.name" :label="table.name.toUpperCase()">
        <template #label>
          <span :class="['table-label', table.name]">
            {{ table.name.toUpperCase() }}
            <el-badge :value="table.total_rules" :hidden="!table.total_rules" />
          </span>
        </template>
        
        <div class="table-chains">
          <el-collapse>
            <el-collapse-item 
              v-for="chain in table.chains" 
              :key="chain.name" 
              :title="chain.name"
            >
              <template #title>
                <div class="chain-header">
                  <span class="chain-name">{{ chain.name }}</span>
                  <span class="chain-policy">策略: {{ chain.policy }}</span>
                  <el-badge :value="chain.rules?.length || 0" :hidden="!chain.rules?.length" />
                </div>
              </template>
              
              <div class="chain-rules">
                <el-empty v-if="!chain.rules?.length" description="无规则" />
                <RulesTable 
                  v-else
                  :rules="chain.rules"
                  @edit-rule="onEditRule"
                  @delete-rule="onDeleteRule"
                />
              </div>
              
              <div class="chain-actions">
                <el-button 
                  type="primary" 
                  size="small" 
                  @click="onAddRule(chain.name, table.name)"
                >
                  添加规则
                </el-button>
              </div>
            </el-collapse-item>
          </el-collapse>
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import RulesTable from '@/components/common/RulesTable.vue'
import type { TableInfo, IPTablesRule } from '@/types/ChainTable/types'

interface Props {
  tables: TableInfo[]
}

interface Emits {
  (e: 'edit-rule', rule: IPTablesRule): void
  (e: 'delete-rule', rule: IPTablesRule): void
  (e: 'add-rule', chainName: string, tableName: string): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const onEditRule = (rule: IPTablesRule) => {
  emit('edit-rule', rule)
}

const onDeleteRule = (rule: IPTablesRule) => {
  emit('delete-rule', rule)
}

const onAddRule = (chainName: string, tableName: string) => {
  emit('add-rule', chainName, tableName)
}
</script>

<style scoped>
.table-view {
  min-height: 600px;
}

.table-label {
  display: flex;
  align-items: center;
  gap: 5px;
}

.table-label.raw {
  color: #909399;
}

.table-label.mangle {
  color: #e6a23c;
}

.table-label.nat {
  color: #67c23a;
}

.table-label.filter {
  color: #f56c6c;
}

.chain-header {
  display: flex;
  align-items: center;
  gap: 10px;
}

.chain-name {
  font-weight: bold;
}

.chain-policy {
  color: #909399;
  font-size: 0.9em;
}

.chain-rules {
  margin-bottom: 15px;
}

.chain-actions {
  display: flex;
  justify-content: flex-end;
}
</style>