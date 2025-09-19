<template>
  <div class="tables-page">
    <!-- 页面标题 -->
    <el-card class="header-card">
      <div class="header-content">
        <h2>IPTables 表管理</h2>
        <p>查看和管理 iptables 的各个表和链的详细信息</p>
        <div class="action-buttons">
          <el-button type="primary" @click="refreshAllTables" :loading="loading">
            <el-icon><Refresh /></el-icon>
            刷新数据
          </el-button>
          <el-button type="info" @click="loadSpecialChains">
            <el-icon><View /></el-icon>
            查看特殊链
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- 表格选择器 -->
    <el-card class="selector-card">
      <el-tabs v-model="activeTable" @tab-click="handleTableChange">
        <el-tab-pane label="所有表" name="all">
          <div class="table-overview">
            <el-row :gutter="20">
              <el-col :span="6" v-for="table in tables" :key="table.table_name">
                <el-card class="table-card" @click="selectTable(table.table_name)">
                  <div class="table-info">
                    <h3>{{ table.table_name.toUpperCase() }} 表</h3>
                    <p>{{ table.chains.length }} 个链</p>
                    <el-tag :type="getTableTagType(table.table_name)">
                      {{ getTableDescription(table.table_name) }}
                    </el-tag>
                  </div>
                </el-card>
              </el-col>
            </el-row>
          </div>
        </el-tab-pane>
        <el-tab-pane label="RAW 表" name="raw"></el-tab-pane>
        <el-tab-pane label="MANGLE 表" name="mangle"></el-tab-pane>
        <el-tab-pane label="NAT 表" name="nat"></el-tab-pane>
        <el-tab-pane label="FILTER 表" name="filter"></el-tab-pane>
        <el-tab-pane label="特殊链" name="special"></el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 表详情 -->
    <el-card v-if="selectedTableInfo && activeTable !== 'all' && activeTable !== 'special'" class="table-detail-card">
      <template #header>
        <div class="card-header">
          <span>{{ selectedTableInfo.table_name.toUpperCase() }} 表详情</span>
          <el-button type="text" @click="refreshTableInfo">
            <el-icon><Refresh /></el-icon>
          </el-button>
        </div>
      </template>
      
      <div v-for="chain in selectedTableInfo.chains" :key="chain.chain_name" class="chain-section">
        <el-collapse v-model="activeChains">
          <el-collapse-item :name="chain.chain_name">
            <template #title>
              <div class="chain-header">
                <el-tag :type="getChainTagType(chain.chain_name)" size="large">
                  {{ chain.chain_name }}
                </el-tag>
                <span class="chain-stats">
                  策略: <el-tag size="small" :type="chain.policy === 'ACCEPT' ? 'success' : 'danger'">{{ chain.policy }}</el-tag>
                  包数: {{ chain.packets }} 字节数: {{ chain.bytes }}
                </span>
                <el-button 
                  type="text" 
                  size="small" 
                  @click.stop="loadChainVerbose(selectedTableInfo.table_name, chain.chain_name)"
                >
                  查看详细
                </el-button>
              </div>
            </template>
            
            <div class="rules-table">
              <el-table :data="chain.rules" stripe style="width: 100%">
                <el-table-column prop="line_number" label="行号" width="80" />
                <el-table-column prop="target" label="目标" width="120">
                  <template #default="scope">
                    <el-tag :type="getTargetTagType(scope.row.target)" size="small">
                      {{ scope.row.target }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="protocol" label="协议" width="100" />
                <el-table-column prop="source" label="源地址" width="150" />
                <el-table-column prop="destination" label="目标地址" width="150" />
                <el-table-column prop="options" label="选项" />
                <el-table-column label="操作" width="120">
                  <template #default="scope">
                    <el-button type="text" size="small" @click="showRuleDetail(scope.row)">
                      详情
                    </el-button>
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </el-collapse-item>
        </el-collapse>
      </div>
    </el-card>

    <!-- 特殊链信息 -->
    <el-card v-if="activeTable === 'special'" class="special-chains-card">
      <template #header>
        <div class="card-header">
          <span>特殊链详情</span>
          <el-button type="text" @click="loadSpecialChains">
            <el-icon><Refresh /></el-icon>
          </el-button>
        </div>
      </template>
      
      <div v-for="specialChain in specialChains" :key="specialChain.name" class="special-chain-section">
        <el-collapse>
          <el-collapse-item>
            <template #title>
              <div class="chain-header">
                <el-tag type="warning" size="large">{{ specialChain.name }}</el-tag>
                <span class="chain-stats">
                  表: {{ specialChain.table }} | 链: {{ specialChain.chain }}
                </span>
              </div>
            </template>
            
            <div class="verbose-info" v-if="specialChain.info">
              <el-descriptions :column="3" border>
                <el-descriptions-item label="链名">{{ specialChain.info.chain_name }}</el-descriptions-item>
                <el-descriptions-item label="策略">
                  <el-tag :type="specialChain.info.policy === 'ACCEPT' ? 'success' : 'danger'">
                    {{ specialChain.info.policy }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="包数/字节数">
                  {{ specialChain.info.packets }} / {{ specialChain.info.bytes }}
                </el-descriptions-item>
              </el-descriptions>
              
              <div class="rules-table" style="margin-top: 20px;">
                <el-table :data="specialChain.info.rules" stripe style="width: 100%">
                  <el-table-column prop="packets" label="包数" width="100" />
                  <el-table-column prop="bytes" label="字节数" width="120" />
                  <el-table-column prop="target" label="目标" width="120">
                    <template #default="scope">
                      <el-tag :type="getTargetTagType(scope.row.target)" size="small">
                        {{ scope.row.target }}
                      </el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="protocol" label="协议" width="100" />
                  <el-table-column prop="source" label="源地址" width="150" />
                  <el-table-column prop="destination" label="目标地址" width="150" />
                  <el-table-column prop="options" label="选项" />
                </el-table>
              </div>
            </div>
          </el-collapse-item>
        </el-collapse>
      </div>
    </el-card>

    <!-- 规则详情对话框 -->
    <el-dialog v-model="ruleDetailVisible" title="规则详情" width="60%">
      <div v-if="selectedRule">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="行号">{{ selectedRule.line_number || 'N/A' }}</el-descriptions-item>
          <el-descriptions-item label="目标">
            <el-tag :type="getTargetTagType(selectedRule.target)">{{ selectedRule.target }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="协议">{{ selectedRule.protocol || 'all' }}</el-descriptions-item>
          <el-descriptions-item label="源地址">{{ selectedRule.source || 'anywhere' }}</el-descriptions-item>
          <el-descriptions-item label="目标地址">{{ selectedRule.destination || 'anywhere' }}</el-descriptions-item>
          <el-descriptions-item label="选项">{{ selectedRule.options || 'none' }}</el-descriptions-item>
          <el-descriptions-item label="包数" v-if="selectedRule.packets">{{ selectedRule.packets }}</el-descriptions-item>
          <el-descriptions-item label="字节数" v-if="selectedRule.bytes">{{ selectedRule.bytes }}</el-descriptions-item>
          <el-descriptions-item label="完整规则" :span="2">
            <el-input
              :value="selectedRule.rule_text"
              type="textarea"
              :rows="3"
              readonly
            />
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, View } from '@element-plus/icons-vue'
import { apiService, type TableInfo, type ChainInfo, type RuleInfo } from '../api'

const loading = ref(false)
const activeTable = ref('all')
const activeChains = ref<string[]>([])
const tables = ref<TableInfo[]>([])
const selectedTableInfo = ref<TableInfo | null>(null)
const specialChains = ref<any[]>([])
const ruleDetailVisible = ref(false)
const selectedRule = ref<RuleInfo | null>(null)

const getTableTagType = (tableName: string) => {
  const types: Record<string, string> = {
    'raw': 'info',
    'mangle': 'warning', 
    'nat': 'success',
    'filter': 'primary'
  }
  return types[tableName] || 'default'
}

const getTableDescription = (tableName: string) => {
  const descriptions: Record<string, string> = {
    'raw': '连接跟踪',
    'mangle': '包修改',
    'nat': '地址转换',
    'filter': '包过滤'
  }
  return descriptions[tableName] || '未知'
}

const getChainTagType = (chainName: string) => {
  const types: Record<string, string> = {
    'INPUT': 'success',
    'OUTPUT': 'warning',
    'FORWARD': 'info',
    'PREROUTING': 'primary',
    'POSTROUTING': 'success'
  }
  return types[chainName] || 'default'
}

const getTargetTagType = (target: string) => {
  const types: Record<string, string> = {
    'ACCEPT': 'success',
    'DROP': 'danger',
    'REJECT': 'warning',
    'RETURN': 'info',
    'SNAT': 'primary',
    'DNAT': 'primary',
    'MASQUERADE': 'success'
  }
  return types[target] || 'default'
}

const refreshAllTables = async () => {
  loading.value = true
  try {
    console.log('[DEBUG] Loading all tables...')
    const response = await apiService.getAllTables()
    tables.value = response.data
    console.log('[DEBUG] Loaded tables:', tables.value)
    ElMessage.success('表信息刷新成功')
  } catch (error) {
    console.error('[ERROR] Failed to load tables:', error)
    ElMessage.error('加载表信息失败')
  } finally {
    loading.value = false
  }
}

const selectTable = (tableName: string) => {
  activeTable.value = tableName
  loadTableInfo(tableName)
}

const loadTableInfo = async (tableName: string) => {
  loading.value = true
  try {
    console.log(`[DEBUG] Loading table info for: ${tableName}`)
    const response = await apiService.getTableInfo(tableName)
    selectedTableInfo.value = response.data
    console.log(`[DEBUG] Loaded table info:`, selectedTableInfo.value)
  } catch (error) {
    console.error(`[ERROR] Failed to load table info for ${tableName}:`, error)
    ElMessage.error(`加载${tableName}表信息失败`)
  } finally {
    loading.value = false
  }
}

const loadChainVerbose = async (tableName: string, chainName: string) => {
  try {
    console.log(`[DEBUG] Loading verbose info for chain: ${tableName}.${chainName}`)
    const response = await apiService.getChainVerbose(tableName, chainName)
    console.log(`[DEBUG] Loaded chain verbose info:`, response.data)
    
    // 更新当前表信息中的链数据
    if (selectedTableInfo.value) {
      const chainIndex = selectedTableInfo.value.chains.findIndex(c => c.chain_name === chainName)
      if (chainIndex !== -1) {
        selectedTableInfo.value.chains[chainIndex] = response.data
      }
    }
    
    ElMessage.success(`${chainName}链详细信息加载成功`)
  } catch (error) {
    console.error(`[ERROR] Failed to load chain verbose info:`, error)
    ElMessage.error(`加载${chainName}链详细信息失败`)
  }
}

const loadSpecialChains = async () => {
  loading.value = true
  try {
    console.log('[DEBUG] Loading special chains...')
    const response = await apiService.getSpecialChains()
    specialChains.value = response.data
    console.log('[DEBUG] Loaded special chains:', specialChains.value)
    ElMessage.success('特殊链信息加载成功')
  } catch (error) {
    console.error('[ERROR] Failed to load special chains:', error)
    ElMessage.error('加载特殊链信息失败')
  } finally {
    loading.value = false
  }
}

const handleTableChange = (tab: any) => {
  const tableName = tab.props.name
  if (tableName === 'all') {
    selectedTableInfo.value = null
  } else if (tableName === 'special') {
    selectedTableInfo.value = null
    loadSpecialChains()
  } else {
    loadTableInfo(tableName)
  }
}

const refreshTableInfo = () => {
  if (selectedTableInfo.value) {
    loadTableInfo(selectedTableInfo.value.table_name)
  }
}

const showRuleDetail = (rule: RuleInfo) => {
  selectedRule.value = rule
  ruleDetailVisible.value = true
}

onMounted(() => {
  refreshAllTables()
})
</script>

<style scoped>
.tables-page {
  padding: 20px;
}

.header-card {
  margin-bottom: 20px;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-content h2 {
  margin: 0;
  color: #303133;
}

.header-content p {
  margin: 5px 0 0 0;
  color: #606266;
  font-size: 14px;
}

.action-buttons {
  display: flex;
  gap: 10px;
}

.selector-card {
  margin-bottom: 20px;
}

.table-overview {
  padding: 20px 0;
}

.table-card {
  cursor: pointer;
  transition: all 0.3s;
  height: 120px;
}

.table-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.table-info {
  text-align: center;
}

.table-info h3 {
  margin: 0 0 10px 0;
  color: #303133;
}

.table-info p {
  margin: 0 0 10px 0;
  color: #606266;
  font-size: 14px;
}

.table-detail-card, .special-chains-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.chain-section, .special-chain-section {
  margin-bottom: 20px;
}

.chain-header {
  display: flex;
  align-items: center;
  gap: 15px;
  width: 100%;
}

.chain-stats {
  color: #606266;
  font-size: 14px;
}

.rules-table {
  margin-top: 15px;
}

.verbose-info {
  padding: 15px;
  background-color: #f8f9fa;
  border-radius: 4px;
}

:deep(.el-collapse-item__header) {
  padding-left: 0;
}

:deep(.el-collapse-item__content) {
  padding-bottom: 15px;
}
</style>