<template>
  <div class="tunnel-analysis-container">
    <el-card class="header-card">
      <template #header>
        <div class="card-header">
          <h2>隧道接口与Docker网桥通信分析</h2>
          <el-button type="primary" @click="refreshData" :loading="loading">
            <el-icon><Refresh /></el-icon>
            刷新数据
          </el-button>
        </div>
      </template>
      
      <div class="analysis-controls">
        <el-row :gutter="20">
          <el-col :span="8">
            <el-select
              v-model="selectedTunnelInterface"
              placeholder="选择隧道接口"
              @change="onTunnelInterfaceChange"
              style="width: 100%"
            >
              <el-option
                v-for="tunnel in tunnelInterfaces"
                :key="tunnel.name"
                :label="`${tunnel.name} (${tunnel.type})`"
                :value="tunnel.name"
              />
            </el-select>
          </el-col>
          <el-col :span="8">
            <el-select
              v-model="selectedDockerBridge"
              placeholder="选择Docker网桥"
              @change="onDockerBridgeChange"
              filterable
              style="width: 100%"
              clearable
            >
              <el-option
                v-for="bridge in dockerBridges"
                :key="bridge.name"
                :label="`${bridge.name} - ${bridge.ip_address || 'N/A'}`"
                :value="bridge.name"
              >
                <span style="float: left">{{ bridge.name }}</span>
                <span style="float: right; color: #8492a6; font-size: 13px">
                  {{ bridge.ip_address || 'N/A' }}
                </span>
              </el-option>
            </el-select>
          </el-col>
          <el-col :span="8">
            <el-button
              type="success"
              @click="analyzeConnection"
              :disabled="!selectedTunnelInterface || !selectedDockerBridge"
              :loading="analyzing"
            >
              <el-icon><Connection /></el-icon>
              分析通信路径
            </el-button>
          </el-col>
        </el-row>
      </div>
    </el-card>

    <!-- 隧道接口信息卡片 -->
    <el-row :gutter="20" v-if="selectedTunnelInterface">
      <el-col :span="12">
        <el-card class="info-card">
          <template #header>
            <h3>隧道接口信息</h3>
          </template>
          <div v-if="tunnelInfo">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="接口名称">{{ tunnelInfo.name }}</el-descriptions-item>
              <el-descriptions-item label="隧道类型">{{ tunnelInfo.tunnel_type }}</el-descriptions-item>
              <el-descriptions-item label="状态">
                <el-tag :type="tunnelInfo.is_up ? 'success' : 'danger'">
                  {{ tunnelInfo.is_up ? 'UP' : 'DOWN' }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="MTU">{{ tunnelInfo.mtu }}</el-descriptions-item>
              <el-descriptions-item label="本地地址">{{ tunnelInfo.local_address || 'N/A' }}</el-descriptions-item>
              <el-descriptions-item label="对端地址">{{ tunnelInfo.peer_address || 'N/A' }}</el-descriptions-item>
              <el-descriptions-item label="IP地址" :span="2">
                <el-tag v-for="ip in tunnelInfo.ip_addresses" :key="ip" style="margin-right: 5px;">
                  {{ ip }}
                </el-tag>
              </el-descriptions-item>
            </el-descriptions>
            
            <div class="statistics-section" style="margin-top: 20px;">
              <h4>流量统计</h4>
              <el-row :gutter="10">
                <el-col :span="6">
                  <el-statistic title="接收字节" :value="tunnelInfo.statistics.rx_bytes" suffix="B" />
                </el-col>
                <el-col :span="6">
                  <el-statistic title="发送字节" :value="tunnelInfo.statistics.tx_bytes" suffix="B" />
                </el-col>
                <el-col :span="6">
                  <el-statistic title="接收包数" :value="tunnelInfo.statistics.rx_packets" />
                </el-col>
                <el-col :span="6">
                  <el-statistic title="发送包数" :value="tunnelInfo.statistics.tx_packets" />
                </el-col>
              </el-row>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="12">
        <el-card class="info-card">
          <template #header>
            <h3>相关iptables规则</h3>
          </template>
          <div v-if="tunnelRules.length > 0">
            <el-table :data="tunnelRules" size="small" max-height="300">
              <el-table-column prop="table" label="表" width="80" />
              <el-table-column prop="chain_name" label="链" width="120" />
              <el-table-column prop="target" label="目标" width="100" />
              <el-table-column prop="protocol" label="协议" width="80" />
              <el-table-column prop="in_interface" label="入接口" width="100" />
              <el-table-column prop="out_interface" label="出接口" width="100" />
              <el-table-column prop="packets" label="包数" width="80" />
            </el-table>
          </div>
          <el-empty v-else description="暂无相关规则" />
        </el-card>
      </el-col>
    </el-row>

    <!-- 通信分析结果 -->
    <el-card v-if="analysisResult" class="analysis-result-card">
      <template #header>
        <div class="analysis-result-header">
          <h3>通信路径分析结果</h3>
          <el-button
            v-if="hasConnectivityIssues"
            type="warning"
            @click="fixConnectivityIssues"
            :loading="fixing"
          >
            <el-icon><Tools /></el-icon>
            一键修复
          </el-button>
        </div>
      </template>
      
      <el-tabs v-model="activeTab">
        <el-tab-pane label="通信路径" name="path">
          <div class="communication-path">
            <el-steps :active="analysisResult.communication_path.length" direction="vertical">
              <el-step
                v-for="step in analysisResult.communication_path"
                :key="step.step"
                :title="step.description"
                :description="`表: ${step.table} | 链: ${step.chain} | 动作: ${step.action}`"
              />
            </el-steps>
          </div>
        </el-tab-pane>
        
        <el-tab-pane label="FORWARD规则" name="forward">
          <el-table :data="analysisResult.forward_rules" size="small">
            <el-table-column prop="line_number" label="行号" width="80" />
            <el-table-column prop="target" label="目标" width="100" />
            <el-table-column prop="protocol" label="协议" width="80" />
            <el-table-column prop="in_interface" label="入接口" width="120" />
            <el-table-column prop="out_interface" label="出接口" width="120" />
            <el-table-column prop="source" label="源地址" width="150" />
            <el-table-column prop="destination" label="目标地址" width="150" />
            <el-table-column prop="packets" label="包数" width="80" />
            <el-table-column prop="bytes" label="字节数" width="100" />
          </el-table>
        </el-tab-pane>
        
        <el-tab-pane label="NAT规则" name="nat">
          <el-table :data="analysisResult.nat_rules" size="small">
            <el-table-column prop="chain_name" label="链" width="120" />
            <el-table-column prop="line_number" label="行号" width="80" />
            <el-table-column prop="target" label="目标" width="100" />
            <el-table-column prop="protocol" label="协议" width="80" />
            <el-table-column prop="in_interface" label="入接口" width="120" />
            <el-table-column prop="out_interface" label="出接口" width="120" />
            <el-table-column prop="source" label="源地址" width="150" />
            <el-table-column prop="destination" label="目标地址" width="150" />
            <el-table-column prop="packets" label="包数" width="80" />
          </el-table>
        </el-tab-pane>
        
        <el-tab-pane label="隔离规则" name="isolation">
          <div v-if="analysisResult.isolation_rules && analysisResult.isolation_rules.length > 0">
            <el-alert
              v-if="hasIsolationDropRules"
              title="检测到有效的Docker隔离规则正在阻断通信"
              type="warning"
              style="margin-bottom: 15px;"
              show-icon
            >
              <template #default>
                DOCKER-ISOLATION-STAGE-2链中的DROP规则正在影响隧道接口与Docker网桥的通信。
                这些规则是Docker网络隔离机制的一部分，需要添加RETURN规则来绕过隔离限制。
              </template>
            </el-alert>
            
            <el-alert
              v-if="!hasIsolationDropRules && analysisResult.isolation_rules && analysisResult.isolation_rules.length > 0"
              title="Docker隔离规则配置正常"
              type="success"
              style="margin-bottom: 15px;"
              show-icon
            >
              <template #default>
                检测到隔离规则，但已通过RETURN规则正确配置，不会影响当前通信路径。
              </template>
            </el-alert>
            
            <el-table :data="analysisResult.isolation_rules" size="small">
              <el-table-column prop="line_number" label="行号" width="80" />
              <el-table-column prop="target" label="目标" width="100">
                <template #default="scope">
                  <el-tag 
                    :type="scope.row.target === 'DROP' ? 'danger' : scope.row.target === 'RETURN' ? 'warning' : 'success'">
                    {{ scope.row.target }}
                  </el-tag>
                  <el-tooltip 
                    v-if="scope.row.target === 'RETURN'" 
                    content="RETURN规则用于绕过后续的DROP规则" 
                    placement="top"
                  >
                    <el-icon style="margin-left: 5px; color: #E6A23C;"><InfoFilled /></el-icon>
                  </el-tooltip>
                </template>
              </el-table-column>
              <el-table-column prop="protocol" label="协议" width="80" />
              <el-table-column prop="in_interface" label="入接口" width="120" />
              <el-table-column prop="out_interface" label="出接口" width="120" />
              <el-table-column prop="source" label="源地址" width="150" />
              <el-table-column prop="destination" label="目标地址" width="150" />
              <el-table-column prop="packets" label="包数" width="80" />
              <el-table-column prop="bytes" label="字节数" width="100" />
            </el-table>
          </div>
          <el-empty v-else description="未检测到相关的Docker隔离规则" />
        </el-tab-pane>
        
        <el-tab-pane label="统计信息" name="statistics">
          <el-row :gutter="20">
            <el-col :span="6">
              <el-statistic
                title="隧道→Docker包数"
                :value="analysisResult.statistics.tunnel_to_docker_packets"
              />
            </el-col>
            <el-col :span="6">
              <el-statistic
                title="Docker→隧道包数"
                :value="analysisResult.statistics.docker_to_tunnel_packets"
              />
            </el-col>
            <el-col :span="6">
              <el-statistic
                title="转发包数"
                :value="analysisResult.statistics.forwarded_packets"
              />
            </el-col>
            <el-col :span="6">
              <el-statistic
                title="丢弃包数"
                :value="analysisResult.statistics.dropped_packets"
                :value-style="{ color: analysisResult.statistics.dropped_packets > 0 ? '#f56c6c' : '#67c23a' }"
              />
            </el-col>
          </el-row>
          
          <div style="margin-top: 20px;">
            <h4>字节统计</h4>
            <el-row :gutter="20">
              <el-col :span="12">
                <el-statistic
                  title="隧道→Docker字节数"
                  :value="analysisResult.statistics.tunnel_to_docker_bytes"
                  suffix="B"
                />
              </el-col>
              <el-col :span="12">
                <el-statistic
                  title="Docker→隧道字节数"
                  :value="analysisResult.statistics.docker_to_tunnel_bytes"
                  suffix="B"
                />
              </el-col>
            </el-row>
          </div>
        </el-tab-pane>
        
        <el-tab-pane label="优化建议" name="recommendations">
          <div v-if="analysisResult.recommendations.length > 0">
            <el-alert
              v-for="(recommendation, index) in analysisResult.recommendations"
              :key="index"
              :title="recommendation"
              type="info"
              style="margin-bottom: 10px;"
              show-icon
            />
          </div>
          <el-empty v-else description="暂无优化建议" />
        </el-tab-pane>
      </el-tabs>
    </el-card>


  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Connection, Tools, InfoFilled } from '@element-plus/icons-vue'
import api from '@/api'

// 响应式数据
const loading = ref(false)
const analyzing = ref(false)
const fixing = ref(false)
const activeTab = ref('path')

const tunnelInterfaces = ref([])
const dockerBridges = ref([])
const selectedTunnelInterface = ref('')
const selectedDockerBridge = ref('')

const tunnelInfo = ref(null)
const tunnelRules = ref([])
const analysisResult = ref(null)

// 计算属性
const hasConnectivityIssues = computed(() => {
  if (!analysisResult.value) return false
  
  // 检查是否存在连通性问题
  const stats = analysisResult.value.statistics
  return stats.dropped_packets > 0 || 
         stats.forwarded_packets === 0 ||
         analysisResult.value.recommendations.some((rec: string) => 
           rec.includes('连通性') || rec.includes('阻塞') || rec.includes('失败')
         )
})

const hasIsolationDropRules = computed(() => {
  if (!analysisResult.value || !analysisResult.value.isolation_rules) return false
  
  // 按行号排序规则
  const sortedRules = [...analysisResult.value.isolation_rules].sort((a, b) => a.line_number - b.line_number)
  
  // 检查是否存在早期的RETURN规则
  let hasEarlyReturn = false
  for (const rule of sortedRules) {
    if (rule.target === 'RETURN') {
      hasEarlyReturn = true
      break
    }
    if (rule.target === 'DROP') {
      break // 如果先遇到DROP规则，说明没有早期RETURN
    }
  }
  
  // 如果有早期RETURN规则，则DROP规则无效
  if (hasEarlyReturn) {
    return false
  }
  
  // 否则检查是否有DROP规则
  return sortedRules.some((rule: any) => rule.target === 'DROP')
})

// 方法
const refreshData = async () => {
  loading.value = true
  try {
    await Promise.all([
      loadTunnelInterfaces(),
      loadDockerBridges()
    ])
    ElMessage.success('数据刷新成功')
  } catch (error: any) {
    ElMessage.error('数据刷新失败: ' + error.message)
  } finally {
    loading.value = false
  }
}

const loadTunnelInterfaces = async () => {
  try {
    const response = await api.get('/tunnel/interfaces')
    tunnelInterfaces.value = response.data.tunnel_interfaces || []
  } catch (error) {
    console.error('加载隧道接口失败:', error)
    throw error
  }
}

const loadDockerBridges = async () => {
  try {
    const response = await api.get('/tunnel/docker-bridges')
    // 仅保留类型为bridge的数据项
    const allBridges = response.data.docker_bridges || []
    dockerBridges.value = allBridges.filter((bridge: any) => bridge.driver === 'bridge')
    console.log('[DEBUG] Filtered Docker bridges:', dockerBridges.value)
  } catch (error) {
    console.error('加载Docker网桥失败:', error)
    throw error
  }
}

const onTunnelInterfaceChange = async () => {
  if (!selectedTunnelInterface.value) return
  
  try {
    // 获取隧道接口信息
    const infoResponse = await api.get(`/tunnel/${selectedTunnelInterface.value}/info`)
    tunnelInfo.value = infoResponse.data.tunnel_info
    
    // 获取相关规则
    const rulesResponse = await api.get(`/tunnel/${selectedTunnelInterface.value}/rules`)
    tunnelRules.value = rulesResponse.data.rules || []
  } catch (error: any) {
    ElMessage.error('获取隧道接口信息失败: ' + error.message)
  }
}

const onDockerBridgeChange = () => {
  // Docker网桥变化时的处理逻辑
}

const fixConnectivityIssues = async () => {
  if (!selectedTunnelInterface.value || !selectedDockerBridge.value) {
    ElMessage.warning('请先选择隧道接口和Docker网桥')
    return
  }
  
  fixing.value = true
  try {
    console.log('[修复开始] 隧道接口:', selectedTunnelInterface.value, '网桥:', selectedDockerBridge.value)
    
    const response = await api.post('/tunnel/fix-connectivity', {
      tunnel_interface: selectedTunnelInterface.value,
      docker_bridge: selectedDockerBridge.value
    })
    
    const fixResult = response.data.fix_result
    if (fixResult && fixResult.success) {
      // 显示详细的修复结果
      const fixedIssues = fixResult.fixed_issues || []
      const appliedRules = fixResult.applied_rules || []
      
      let message = `🎉 修复成功！`
      if (fixedIssues.length > 0) {
        message += `共处理 ${fixedIssues.length} 个问题`
      }
      if (appliedRules.length > 0) {
        message += `，应用了 ${appliedRules.length} 条iptables规则`
      }
      
      ElMessage({
        message: message,
        type: 'success',
        duration: 8000,
        showClose: true
      })
      
      // 在控制台显示详细信息
      console.log('[修复成功] 修复详情:')
      console.log('  已修复问题:', fixedIssues)
      console.log('  应用规则:', appliedRules)
      console.log('  修复配置:', {
        tunnel: selectedTunnelInterface.value,
        bridge: selectedDockerBridge.value
      })
      
      // 显示修复详情的通知
      if (fixedIssues.length > 0) {
        const issuesList = fixedIssues.map((issue, index) => `${index + 1}. ${issue}`).join('\n')
        ElMessage({
          message: `修复详情:\n${issuesList}`,
          type: 'info',
          duration: 10000,
          showClose: true
        })
      }
    } else {
      ElMessage.warning('修复完成，但可能存在部分问题。请检查日志获取详细信息。')
      console.log('[修复警告] 修复结果:', fixResult)
    }
    
    // 修复完成后重新分析
    console.log('[修复完成] 开始重新分析连通性...')
    await analyzeConnection()
  } catch (error: any) {
    console.error('[修复失败] 错误详情:', error)
    ElMessage.error('修复失败: ' + error.message)
  } finally {
    fixing.value = false
  }
}

const analyzeConnection = async () => {
  if (!selectedTunnelInterface.value || !selectedDockerBridge.value) {
    ElMessage.warning('请先选择隧道接口和Docker网桥')
    return
  }
  
  analyzing.value = true
  try {
    const response = await api.get('/tunnel/analyze-communication', {
      params: {
        tunnel_interface: selectedTunnelInterface.value,
        docker_bridge: selectedDockerBridge.value
      }
    })
    analysisResult.value = response.data.analysis
    ElMessage.success('通信路径分析完成')
  } catch (error: any) {
    ElMessage.error('分析失败: ' + error.message)
  } finally {
    analyzing.value = false
  }
}



// 生命周期
onMounted(() => {
  refreshData()
})
</script>

<style scoped>
.tunnel-analysis-container {
  padding: 20px;
}

.header-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.analysis-controls {
  margin-top: 20px;
}

.info-card {
  margin-bottom: 20px;
}

.analysis-result-card {
  margin-bottom: 20px;
}

.analysis-result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.communication-path {
  padding: 20px;
}

.statistics-section {
  border-top: 1px solid #ebeef5;
  padding-top: 20px;
}
</style>