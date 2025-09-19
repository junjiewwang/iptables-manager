<template>
  <div class="interfaces-container">
    <div class="page-header">
      <h1>网络接口管理</h1>
      <p>查看系统网络接口信息，包括Docker网桥设备</p>
    </div>

    <div class="content-tabs">
      <div class="tab-buttons">
        <button 
          :class="{ active: activeTab === 'interfaces' }"
          @click="activeTab = 'interfaces'"
        >
          网络接口
        </button>
        <button 
          :class="{ active: activeTab === 'docker' }"
          @click="activeTab = 'docker'"
        >
          Docker网桥
        </button>
      </div>

      <!-- 网络接口标签页 -->
      <div v-if="activeTab === 'interfaces'" class="tab-content">
        <div class="toolbar">
          <button @click="loadInterfaces" class="btn btn-primary">
            <i class="fas fa-sync-alt"></i>
            刷新接口
          </button>
          <div class="filter-group">
            <select v-model="interfaceFilter" @change="filterInterfaces">
              <option value="all">所有接口</option>
              <option value="up">活动接口</option>
              <option value="docker">Docker接口</option>
              <option value="ethernet">以太网接口</option>
              <option value="bridge">网桥接口</option>
            </select>
          </div>
        </div>

        <div class="interfaces-grid">
          <div 
            v-for="iface in filteredInterfaces" 
            :key="iface.name"
            class="interface-card"
            :class="{ 
              'docker-interface': iface.is_docker,
              'interface-up': iface.is_up,
              'interface-down': !iface.is_up
            }"
          >
            <div class="interface-header">
              <h3>{{ iface.name }}</h3>
              <div class="interface-badges">
                <span class="badge" :class="iface.is_up ? 'badge-success' : 'badge-danger'">
                  {{ iface.state }}
                </span>
                <span v-if="iface.is_docker" class="badge badge-info">
                  {{ iface.docker_type || 'Docker' }}
                </span>
                <span class="badge badge-secondary">{{ iface.type }}</span>
              </div>
            </div>

            <div class="interface-details">
              <div class="detail-row">
                <span class="label">MAC地址:</span>
                <span class="value">{{ iface.mac_address || 'N/A' }}</span>
              </div>
              <div class="detail-row">
                <span class="label">MTU:</span>
                <span class="value">{{ iface.mtu }}</span>
              </div>
              <div class="detail-row">
                <span class="label">IP地址:</span>
                <div class="ip-addresses">
                  <span 
                    v-for="ip in (iface.ip_addresses || [])" 
                    :key="ip"
                    class="ip-address"
                  >
                    {{ ip }}
                  </span>
                  <span v-if="!iface.ip_addresses || iface.ip_addresses.length === 0" class="no-ip">
                    无IP地址
                  </span>
                </div>
              </div>
            </div>

              <div class="interface-stats" v-if="iface.statistics">
                <h4>流量统计</h4>
                <div class="stats-grid">
                  <div class="stat-item">
                    <span class="stat-label">接收字节</span>
                    <span class="stat-value">{{ formatBytes(iface.statistics.rx_bytes || 0) }}</span>
                  </div>
                  <div class="stat-item">
                    <span class="stat-label">发送字节</span>
                    <span class="stat-value">{{ formatBytes(iface.statistics.tx_bytes || 0) }}</span>
                  </div>
                  <div class="stat-item">
                    <span class="stat-label">接收包数</span>
                    <span class="stat-value">{{ (iface.statistics.rx_packets || 0).toLocaleString() }}</span>
                  </div>
                  <div class="stat-item">
                    <span class="stat-label">发送包数</span>
                    <span class="stat-value">{{ (iface.statistics.tx_packets || 0).toLocaleString() }}</span>
                  </div>
                  <div class="stat-item">
                    <span class="stat-label">接收错误</span>
                    <span class="stat-value error">{{ iface.statistics.rx_errors || 0 }}</span>
                  </div>
                  <div class="stat-item">
                    <span class="stat-label">发送错误</span>
                    <span class="stat-value error">{{ iface.statistics.tx_errors || 0 }}</span>
                  </div>
                </div>
              </div>
          </div>
        </div>
      </div>

      <!-- Docker网桥标签页 -->
      <div v-if="activeTab === 'docker'" class="tab-content">
        <div class="toolbar">
          <button @click="loadDockerBridges" class="btn btn-primary">
            <i class="fas fa-sync-alt"></i>
            刷新网桥
          </button>
        </div>

        <div class="bridges-grid">
          <div 
            v-for="bridge in dockerBridges" 
            :key="bridge.network_id"
            class="bridge-card"
          >
            <div class="bridge-header">
              <h3>{{ bridge.name }}</h3>
              <div class="bridge-badges">
                <span class="badge badge-primary">{{ bridge.driver }}</span>
                <span class="badge badge-secondary">{{ bridge.scope }}</span>
              </div>
            </div>

            <div class="bridge-details">
              <div class="detail-section">
                <h4>网络信息</h4>
                <div class="detail-row">
                  <span class="label">网络ID:</span>
                  <span class="value">{{ bridge.network_id.substring(0, 12) }}...</span>
                </div>
                <div class="detail-row">
                  <span class="label">IPAM驱动:</span>
                  <span class="value">{{ bridge.ipam_config?.driver || 'N/A' }}</span>
                </div>
                <div v-if="bridge.ipam_config && bridge.ipam_config.config && bridge.ipam_config.config.length > 0" class="detail-row">
                  <span class="label">子网:</span>
                  <div class="subnets">
                    <div 
                      v-for="subnet in bridge.ipam_config.config" 
                      :key="subnet.subnet"
                      class="subnet-info"
                    >
                      <span class="subnet">{{ subnet.subnet }}</span>
                      <span class="gateway">网关: {{ subnet.gateway }}</span>
                    </div>
                  </div>
                </div>
              </div>

              <div class="detail-section">
                <h4>接口信息</h4>
                <div class="detail-row">
                  <span class="label">接口名称:</span>
                  <span class="value">{{ bridge.interface?.name || 'N/A' }}</span>
                </div>
                <div class="detail-row">
                  <span class="label">状态:</span>
                  <span class="badge" :class="bridge.interface?.is_up ? 'badge-success' : 'badge-danger'">
                    {{ bridge.interface?.state || 'UNKNOWN' }}
                  </span>
                </div>
                <div class="detail-row">
                  <span class="label">IP地址:</span>
                  <div class="ip-addresses">
                    <span 
                      v-for="ip in (bridge.interface?.ip_addresses || [])" 
                      :key="ip"
                      class="ip-address"
                    >
                      {{ ip }}
                    </span>
                    <span v-if="!bridge.interface?.ip_addresses || bridge.interface.ip_addresses.length === 0" class="no-ip">
                      无IP地址
                    </span>
                  </div>
                </div>
              </div>

              <div v-if="bridge.containers && bridge.containers.length > 0" class="detail-section">
                <h4>连接的容器 ({{ bridge.containers.length }})</h4>
                <div class="containers-list">
                  <div 
                    v-for="container in bridge.containers" 
                    :key="container.id"
                    class="container-item"
                  >
                    <div class="container-name">{{ container.name }}</div>
                    <div class="container-ip">{{ container.ip_address }}</div>
                    <div class="container-id">{{ container.id.substring(0, 12) }}</div>
                  </div>
                </div>
              </div>

              <div class="bridge-actions">
                <button 
                  @click="viewBridgeRules(bridge.name)"
                  class="btn btn-outline"
                >
                  <i class="fas fa-list"></i>
                  查看规则
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 规则查看模态框 -->
    <div v-if="showRulesModal" class="modal-overlay" @click="closeRulesModal">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>{{ selectedBridge }} 网桥规则</h3>
          <button @click="closeRulesModal" class="close-btn">&times;</button>
        </div>
        <div class="modal-body">
          <div v-if="bridgeRules.length === 0" class="no-rules">
            该网桥暂无相关规则
          </div>
          <div v-else class="rules-list">
            <div 
              v-for="rule in bridgeRules" 
              :key="rule.id"
              class="rule-item"
            >
              <div class="rule-header">
                <span class="rule-table">{{ rule.table }}</span>
                <span class="rule-chain">{{ rule.chain_name }}</span>
                <span class="rule-target">{{ rule.target }}</span>
              </div>
              <div class="rule-details">
                <span v-if="rule.protocol">协议: {{ rule.protocol }}</span>
                <span v-if="rule.source">源: {{ rule.source }}</span>
                <span v-if="rule.destination">目标: {{ rule.destination }}</span>
                <span v-if="rule.in_interface">入接口: {{ rule.in_interface }}</span>
                <span v-if="rule.out_interface">出接口: {{ rule.out_interface }}</span>
              </div>
              <div v-if="rule.extra" class="rule-extra">{{ rule.extra }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-overlay">
      <div class="loading-spinner">
        <i class="fas fa-spinner fa-spin"></i>
        <span>加载中...</span>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted, computed } from 'vue'
import { apiService } from '../api'

export default {
  name: 'Interfaces',
  setup() {
    const activeTab = ref('interfaces')
    const loading = ref(false)
    const interfaces = ref([])
    const dockerBridges = ref([])
    const interfaceFilter = ref('all')
    const showRulesModal = ref(false)
    const selectedBridge = ref('')
    const bridgeRules = ref([])

    // 过滤后的接口列表
    const filteredInterfaces = computed(() => {
      if (interfaceFilter.value === 'all') {
        return interfaces.value
      }
      
      return interfaces.value.filter(iface => {
        switch (interfaceFilter.value) {
          case 'up':
            return iface.is_up
          case 'docker':
            return iface.is_docker
          case 'ethernet':
            return iface.type === 'ethernet'
          case 'bridge':
            return iface.type === 'bridge'
          default:
            return true
        }
      })
    })

    // 加载网络接口
    const loadInterfaces = async () => {
      loading.value = true
      try {
        console.log('[DEBUG] Loading network interfaces...')
        const response = await apiService.getInterfaces()
        // 确保数据结构正确，防止null值
        interfaces.value = (response.data || []).map(iface => ({
          ...iface,
          ip_addresses: iface.ip_addresses || [],
          statistics: iface.statistics || {
            rx_bytes: 0,
            tx_bytes: 0,
            rx_packets: 0,
            tx_packets: 0,
            rx_errors: 0,
            tx_errors: 0
          }
        }))
        console.log('[DEBUG] Loaded interfaces:', interfaces.value)
      } catch (error) {
        console.error('[ERROR] Failed to load interfaces:', error)
        interfaces.value = []
        // 这里可以添加错误提示
      } finally {
        loading.value = false
      }
    }

    // 加载Docker网桥
    const loadDockerBridges = async () => {
      loading.value = true
      try {
        console.log('[DEBUG] Loading Docker bridges...')
        const response = await apiService.getDockerBridges()
        // 确保数据结构正确，防止null值
        dockerBridges.value = (response.data || []).map(bridge => ({
          ...bridge,
          containers: bridge.containers || [],
          ipam_config: bridge.ipam_config || {
            driver: 'default',
            config: [],
            options: {}
          },
          interface: bridge.interface || {
            name: 'N/A',
            state: 'UNKNOWN',
            is_up: false,
            ip_addresses: []
          }
        }))
        console.log('[DEBUG] Loaded Docker bridges:', dockerBridges.value)
      } catch (error) {
        console.error('[ERROR] Failed to load Docker bridges:', error)
        dockerBridges.value = []
        // 这里可以添加错误提示
      } finally {
        loading.value = false
      }
    }

    // 查看网桥规则
    const viewBridgeRules = async (bridgeName) => {
      selectedBridge.value = bridgeName
      showRulesModal.value = true
      
      try {
        console.log('[DEBUG] Loading rules for bridge:', bridgeName)
        const response = await apiService.getBridgeRules(bridgeName)
        bridgeRules.value = response.data
        console.log('[DEBUG] Loaded bridge rules:', bridgeRules.value)
      } catch (error) {
        console.error('[ERROR] Failed to load bridge rules:', error)
        bridgeRules.value = []
      }
    }

    // 关闭规则模态框
    const closeRulesModal = () => {
      showRulesModal.value = false
      selectedBridge.value = ''
      bridgeRules.value = []
    }

    // 格式化字节数
    const formatBytes = (bytes) => {
      if (bytes === 0) return '0 B'
      const k = 1024
      const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
      const i = Math.floor(Math.log(bytes) / Math.log(k))
      return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
    }

    // 过滤接口
    const filterInterfaces = () => {
      // 过滤逻辑已在computed中实现
    }

    onMounted(() => {
      loadInterfaces()
      loadDockerBridges()
    })

    return {
      activeTab,
      loading,
      interfaces,
      dockerBridges,
      interfaceFilter,
      filteredInterfaces,
      showRulesModal,
      selectedBridge,
      bridgeRules,
      loadInterfaces,
      loadDockerBridges,
      viewBridgeRules,
      closeRulesModal,
      formatBytes,
      filterInterfaces
    }
  }
}
</script>

<style scoped>
.interfaces-container {
  padding: 20px;
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 30px;
}

.page-header h1 {
  color: #2c3e50;
  margin-bottom: 10px;
}

.page-header p {
  color: #7f8c8d;
  font-size: 16px;
}

.content-tabs {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.tab-buttons {
  display: flex;
  background: #f8f9fa;
  border-bottom: 1px solid #dee2e6;
}

.tab-buttons button {
  flex: 1;
  padding: 15px 20px;
  border: none;
  background: transparent;
  cursor: pointer;
  font-size: 16px;
  font-weight: 500;
  color: #6c757d;
  transition: all 0.3s ease;
}

.tab-buttons button.active {
  background: white;
  color: #007bff;
  border-bottom: 2px solid #007bff;
}

.tab-buttons button:hover {
  background: #e9ecef;
}

.tab-content {
  padding: 20px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 15px;
  background: #f8f9fa;
  border-radius: 6px;
}

.filter-group select {
  padding: 8px 12px;
  border: 1px solid #ced4da;
  border-radius: 4px;
  background: white;
}

.interfaces-grid, .bridges-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 20px;
}

.interface-card, .bridge-card {
  background: white;
  border: 1px solid #dee2e6;
  border-radius: 8px;
  padding: 20px;
  transition: all 0.3s ease;
}

.interface-card:hover, .bridge-card:hover {
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.interface-card.docker-interface {
  border-left: 4px solid #007bff;
}

.interface-card.interface-up {
  border-left: 4px solid #28a745;
}

.interface-card.interface-down {
  border-left: 4px solid #dc3545;
}

.interface-header, .bridge-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  padding-bottom: 10px;
  border-bottom: 1px solid #eee;
}

.interface-header h3, .bridge-header h3 {
  margin: 0;
  color: #2c3e50;
  font-size: 18px;
}

.interface-badges, .bridge-badges {
  display: flex;
  gap: 8px;
}

.badge {
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
  text-transform: uppercase;
}

.badge-success {
  background: #d4edda;
  color: #155724;
}

.badge-danger {
  background: #f8d7da;
  color: #721c24;
}

.badge-info {
  background: #d1ecf1;
  color: #0c5460;
}

.badge-primary {
  background: #cce5ff;
  color: #004085;
}

.badge-secondary {
  background: #e2e3e5;
  color: #383d41;
}

.interface-details, .bridge-details {
  margin-bottom: 15px;
}

.detail-section {
  margin-bottom: 20px;
}

.detail-section h4 {
  margin: 0 0 10px 0;
  color: #495057;
  font-size: 14px;
  font-weight: 600;
  text-transform: uppercase;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  padding: 5px 0;
}

.detail-row .label {
  font-weight: 500;
  color: #6c757d;
  min-width: 80px;
}

.detail-row .value {
  color: #2c3e50;
  font-family: 'Courier New', monospace;
}

.ip-addresses {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.ip-address {
  background: #f8f9fa;
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'Courier New', monospace;
  font-size: 12px;
}

.no-ip {
  color: #6c757d;
  font-style: italic;
}

.interface-stats {
  border-top: 1px solid #eee;
  padding-top: 15px;
}

.interface-stats h4 {
  margin: 0 0 10px 0;
  color: #495057;
  font-size: 14px;
  font-weight: 600;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  padding: 8px;
  background: #f8f9fa;
  border-radius: 4px;
}

.stat-label {
  font-size: 12px;
  color: #6c757d;
  margin-bottom: 4px;
}

.stat-value {
  font-weight: 600;
  color: #2c3e50;
  font-family: 'Courier New', monospace;
}

.stat-value.error {
  color: #dc3545;
}

.subnets {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.subnet-info {
  display: flex;
  flex-direction: column;
  background: #f8f9fa;
  padding: 6px;
  border-radius: 4px;
}

.subnet {
  font-family: 'Courier New', monospace;
  font-weight: 600;
}

.gateway {
  font-size: 12px;
  color: #6c757d;
}

.containers-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.container-item {
  background: #f8f9fa;
  padding: 10px;
  border-radius: 4px;
  border-left: 3px solid #007bff;
}

.container-name {
  font-weight: 600;
  color: #2c3e50;
}

.container-ip {
  font-family: 'Courier New', monospace;
  color: #495057;
  font-size: 14px;
}

.container-id {
  font-family: 'Courier New', monospace;
  color: #6c757d;
  font-size: 12px;
}

.bridge-actions {
  margin-top: 15px;
  padding-top: 15px;
  border-top: 1px solid #eee;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  transition: all 0.3s ease;
}

.btn-primary {
  background: #007bff;
  color: white;
}

.btn-primary:hover {
  background: #0056b3;
}

.btn-outline {
  background: transparent;
  color: #007bff;
  border: 1px solid #007bff;
}

.btn-outline:hover {
  background: #007bff;
  color: white;
}

/* 模态框样式 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  border-radius: 8px;
  max-width: 800px;
  max-height: 80vh;
  width: 90%;
  overflow: hidden;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #dee2e6;
  background: #f8f9fa;
}

.modal-header h3 {
  margin: 0;
  color: #2c3e50;
}

.close-btn {
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
  color: #6c757d;
  padding: 0;
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.close-btn:hover {
  color: #dc3545;
}

.modal-body {
  padding: 20px;
  max-height: 60vh;
  overflow-y: auto;
}

.no-rules {
  text-align: center;
  color: #6c757d;
  font-style: italic;
  padding: 40px;
}

.rules-list {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.rule-item {
  background: #f8f9fa;
  border: 1px solid #dee2e6;
  border-radius: 6px;
  padding: 15px;
}

.rule-header {
  display: flex;
  gap: 10px;
  margin-bottom: 10px;
}

.rule-table, .rule-chain, .rule-target {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  font-family: 'Courier New', monospace;
}

.rule-table {
  background: #e3f2fd;
  color: #1565c0;
}

.rule-chain {
  background: #f3e5f5;
  color: #7b1fa2;
}

.rule-target {
  background: #e8f5e8;
  color: #2e7d32;
}

.rule-details {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 8px;
}

.rule-details span {
  background: white;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  border: 1px solid #dee2e6;
}

.rule-extra {
  background: white;
  padding: 8px;
  border-radius: 4px;
  font-family: 'Courier New', monospace;
  font-size: 12px;
  border: 1px solid #dee2e6;
  margin-top: 8px;
}

/* 加载状态 */
.loading-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 999;
}

.loading-spinner {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  color: #007bff;
}

.loading-spinner i {
  font-size: 24px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .interfaces-grid, .bridges-grid {
    grid-template-columns: 1fr;
  }
  
  .toolbar {
    flex-direction: column;
    gap: 10px;
  }
  
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .modal-content {
    width: 95%;
    margin: 10px;
  }
}
</style>