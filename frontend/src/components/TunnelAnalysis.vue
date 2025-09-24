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
              style="width: 100%"
            >
              <el-option
                v-for="bridge in dockerBridges"
                :key="bridge.name"
                :label="`${bridge.name} (${bridge.docker_type})`"
                :value="bridge.name"
              />
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
        <h3>通信路径分析结果</h3>
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

    <!-- 规则生成器 -->
    <el-card class="rule-generator-card">
      <template #header>
        <h3>规则生成器</h3>
      </template>
      
      <el-form :model="ruleForm" label-width="120px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="通信方向">
              <el-select v-model="ruleForm.direction" style="width: 100%">
                <el-option label="双向通信" value="bidirectional" />
                <el-option label="入站" value="inbound" />
                <el-option label="出站" value="outbound" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="协议">
              <el-select v-model="ruleForm.protocol" style="width: 100%">
                <el-option label="全部" value="all" />
                <el-option label="TCP" value="tcp" />
                <el-option label="UDP" value="udp" />
                <el-option label="ICMP" value="icmp" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="目标端口">
              <el-input v-model="ruleForm.dest_port" placeholder="如: 80,443,8080-8090" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="动作">
              <el-select v-model="ruleForm.action" style="width: 100%">
                <el-option label="接受" value="ACCEPT" />
                <el-option label="丢弃" value="DROP" />
                <el-option label="拒绝" value="REJECT" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item>
              <el-checkbox v-model="ruleForm.enable_nat">启用NAT</el-checkbox>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item>
              <el-checkbox v-model="ruleForm.enable_logging">启用日志</el-checkbox>
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-form-item>
          <el-button
            type="primary"
            @click="generateRules"
            :disabled="!selectedTunnelInterface || !selectedDockerBridge"
            :loading="generating"
          >
            生成规则
          </el-button>
        </el-form-item>
      </el-form>
      
      <div v-if="generatedRules.length > 0" class="generated-rules">
        <h4>生成的规则:</h4>
        <el-input
          v-for="(rule, index) in generatedRules"
          :key="index"
          :value="rule"
          readonly
          style="margin-bottom: 5px;"
        >
          <template #append>
            <el-button @click="copyRule(rule)">复制</el-button>
          </template>
        </el-input>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Connection } from '@element-plus/icons-vue'
import api from '@/api'

// 响应式数据
const loading = ref(false)
const analyzing = ref(false)
const generating = ref(false)
const activeTab = ref('path')

const tunnelInterfaces = ref([])
const dockerBridges = ref([])
const selectedTunnelInterface = ref('')
const selectedDockerBridge = ref('')

const tunnelInfo = ref(null)
const tunnelRules = ref([])
const analysisResult = ref(null)
const generatedRules = ref([])

const ruleForm = ref({
  direction: 'bidirectional',
  protocol: 'all',
  dest_port: '',
  action: 'ACCEPT',
  enable_nat: true,
  enable_logging: false
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
  } catch (error) {
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
    dockerBridges.value = response.data.docker_bridges || []
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
  } catch (error) {
    ElMessage.error('获取隧道接口信息失败: ' + error.message)
  }
}

const onDockerBridgeChange = () => {
  // Docker网桥变化时的处理逻辑
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
  } catch (error) {
    ElMessage.error('分析失败: ' + error.message)
  } finally {
    analyzing.value = false
  }
}

const generateRules = async () => {
  if (!selectedTunnelInterface.value || !selectedDockerBridge.value) {
    ElMessage.warning('请先选择隧道接口和Docker网桥')
    return
  }
  
  generating.value = true
  try {
    const response = await api.post('/tunnel/generate-rules', {
      tunnel_interface: selectedTunnelInterface.value,
      docker_bridge: selectedDockerBridge.value,
      ...ruleForm.value
    })
    generatedRules.value = response.data.generated_rules || []
    ElMessage.success('规则生成成功')
  } catch (error) {
    ElMessage.error('规则生成失败: ' + error.message)
  } finally {
    generating.value = false
  }
}

const copyRule = async (rule: string) => {
  try {
    await navigator.clipboard.writeText(rule)
    ElMessage.success('规则已复制到剪贴板')
  } catch (error) {
    ElMessage.error('复制失败')
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

.rule-generator-card {
  margin-bottom: 20px;
}

.communication-path {
  padding: 20px;
}

.statistics-section {
  border-top: 1px solid #ebeef5;
  padding-top: 20px;
}

.generated-rules {
  margin-top: 20px;
  padding: 20px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.generated-rules h4 {
  margin-bottom: 15px;
  color: #303133;
}
</style>