<template>
  <div class="chain-table-view">
    <!-- 页面头部 -->
    <div class="page-header">
      <h1>IPTables 五链四表可视化</h1>
      <p class="description">展示PREROUTING、INPUT、FORWARD、OUTPUT、POSTROUTING五链与raw、mangle、nat、filter四表的关系</p>
    </div>

    <!-- 控制面板 -->
    <div class="control-panel">
      <el-card>
        <!-- 主要控制区 -->
        <div class="main-controls">
          <div class="view-tabs">
            <el-tabs v-model="viewMode" @tab-change="handleViewModeChange" type="card">
              <el-tab-pane label="链视图" name="chain">
                <template #label>
                  <el-icon><Share /></el-icon>
                  链视图
                </template>
              </el-tab-pane>
              <el-tab-pane label="表视图" name="table">
                <template #label>
                  <el-icon><Grid /></el-icon>
                  表视图
                </template>
              </el-tab-pane>
              <el-tab-pane label="接口视图" name="interface">
                <template #label>
                  <el-icon><Connection /></el-icon>
                  接口视图
                </template>
              </el-tab-pane>
            </el-tabs>
          </div>
          
          <div class="action-buttons">
            <el-button @click="refreshData" :loading="loading" type="primary">
              <el-icon><Refresh /></el-icon>
              刷新数据
            </el-button>
            <el-button v-if="viewMode === 'chain'" @click="showTopoSettingsDialog = true" type="success">
              <el-icon><Setting /></el-icon>
              拓扑设置
            </el-button>
            <el-button v-if="viewMode === 'chain'" @click="standardizeConnectionPaths" type="info">
              <el-icon><Position /></el-icon>
              标准化路径
            </el-button>
          </div>
        </div>

        <!-- 筛选工具栏 -->
        <el-collapse v-model="activeFilterPanels" class="filter-panel">
          <el-collapse-item title="筛选条件" name="filters">
            <template #title>
              <div class="filter-title">
                <el-icon><Filter /></el-icon>
                <span>筛选条件</span>
                <el-badge :value="activeFiltersCount" :hidden="activeFiltersCount === 0" type="primary" />
              </div>
            </template>
            
            <div class="filter-content">
              <!-- 快捷筛选标签 -->
              <div class="quick-filters">
                <div class="filter-group">
                  <label class="filter-label">网络接口:</label>
                  <div class="filter-tags">
                    <el-tag
                      v-for="iface in interfaces"
                      :key="iface.name"
                      :type="selectedInterfaces.includes(iface.name) ? 'primary' : 'info'"
                      :effect="selectedInterfaces.includes(iface.name) ? 'dark' : 'plain'"
                      @click="toggleInterface(iface.name)"
                      class="filter-tag"
                    >
                      {{ iface.name }}
                    </el-tag>
                  </div>
                </div>

                <div class="filter-group">
                  <label class="filter-label">协议类型:</label>
                  <div class="filter-tags">
                    <el-tag
                      v-for="protocol in availableProtocols"
                      :key="protocol"
                      :type="selectedProtocols.includes(protocol) ? 'success' : 'info'"
                      :effect="selectedProtocols.includes(protocol) ? 'dark' : 'plain'"
                      @click="toggleProtocol(protocol)"
                      class="filter-tag"
                    >
                      {{ protocol.toUpperCase() }}
                    </el-tag>
                  </div>
                </div>

                <div class="filter-group">
                  <label class="filter-label">目标动作:</label>
                  <div class="filter-tags">
                    <el-tag
                      v-for="target in availableTargets"
                      :key="target"
                      :type="getTargetTagType(target)"
                      :effect="selectedTargets.includes(target) ? 'dark' : 'plain'"
                      @click="toggleTarget(target)"
                      class="filter-tag"
                    >
                      {{ target }}
                    </el-tag>
                  </div>
                </div>
              </div>

              <!-- 高级筛选 -->
              <div class="advanced-filters">
                <el-row :gutter="16">
                  <el-col :span="8">
                    <el-input
                      v-model="ipRangeFilter"
                      placeholder="IP地址范围 (如: 192.168.1.0/24)"
                      clearable
                      size="small"
                    >
                      <template #prefix>
                        <el-icon><Location /></el-icon>
                      </template>
                    </el-input>
                  </el-col>
                  <el-col :span="8">
                    <el-input
                      v-model="portRangeFilter"
                      placeholder="端口范围 (如: 80,443,8000-9000)"
                      clearable
                      size="small"
                    >
                      <template #prefix>
                        <el-icon><Connection /></el-icon>
                      </template>
                    </el-input>
                  </el-col>
                  <el-col :span="8">
                    <div class="filter-actions">
                      <el-button size="small" @click="clearAllFilters">
                        <el-icon><Delete /></el-icon>
                        清空筛选
                      </el-button>
                      <el-button size="small" type="primary" @click="applyFilters">
                        <el-icon><Search /></el-icon>
                        应用筛选
                      </el-button>
                    </div>
                  </el-col>
                </el-row>
              </div>
            </div>
          </el-collapse-item>
        </el-collapse>
      </el-card>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="8" animated />
    </div>

    <!-- 主要内容区域 -->
    <div class="main-content">
      <!-- Vue Flow 数据流图视图 -->
      <div v-if="viewMode === 'chain'" class="dataflow-view">
        <div class="vue-flow-wrapper">
          <VueFlow
            v-model="flowElements"
            class="dataflow-diagram"
            :class="{ 'dark': topoSettings.darkMode }"
            :default-viewport="{ zoom: 0.8 }"
            :min-zoom="0.5"
            :max-zoom="2"
            :snap-to-grid="topoSettings.snapToGrid"
            :snap-grid="[20, 20]"
            :node-draggable="topoSettings.enableDrag"
            :auto-connect="false"
            :connection-mode="ConnectionMode.Strict"
            :fit-view-on-init="true"
            :elevate-edges-on-select="true"
            :default-edge-options="{ animated: topoSettings.animateEdges }"
            @node-click="onNodeClick"
            @edge-click="onEdgeClick"
          >
            <!-- 背景 -->
            <Background 
              :pattern-color="topoSettings.darkMode ? '#2d3748' : '#e2e8f0'" 
              :gap="20" 
              :variant="topoSettings.darkMode ? 'dots' : 'lines'"
            />
            
            <!-- 控制面板 -->
            <Controls />
            
            <!-- 小地图 -->
            <MiniMap v-if="topoSettings.showMinimap" height="100" width="150" />
            
            <!-- 自定义节点模板 -->
            <template #node-chain="{ data }">
              <div class="chain-node" :class="[data.chainType, topoSettings.nodeStyle]">
                <div class="chain-header">
                  <h3 class="chain-title">{{ data.label }}</h3>
                </div>
                <div class="chain-tables">
                  <span 
                    v-for="table in data.tables" 
                    :key="table"
                    class="table-tag"
                    :class="table"
                    @click.stop="selectChainTable(data.chainName, table)"
                  >
                    {{ table }}
                  </span>
                </div>
                <div class="chain-stats">
                  <span>{{ data.ruleCount }} 规则</span>
                </div>
              </div>
            </template>
            
            <template #node-decision="{ data }">
              <div class="decision-node" :class="topoSettings.nodeStyle">
                <div class="decision-content">
                  <div class="decision-icon">
                    <el-icon class="router-icon"><Connection /></el-icon>
                  </div>
                  <div class="decision-label">{{ data.label }}</div>
                </div>
              </div>
            </template>
            
            <template #node-endpoint="{ data }">
              <div class="endpoint-node" :class="[data.type, topoSettings.nodeStyle]">
                <div class="endpoint-content">
                  <div class="endpoint-icon">
                    <el-icon v-if="data.type === 'entry'" class="server-icon"><Monitor /></el-icon>
                    <el-icon v-else-if="data.type === 'exit'" class="server-icon"><House /></el-icon>
                    <el-icon v-else><Connection /></el-icon>
                  </div>
                  <div class="endpoint-label">{{ data.label }}</div>
                </div>
              </div>
            </template>
            
            <template #node-process="{ data }">
              <div class="process-node" :class="topoSettings.nodeStyle">
                <div class="process-content">
                  <div class="process-icon">
                    <el-icon class="gear-icon"><Setting /></el-icon>
                  </div>
                  <div class="process-label">{{ data.label }}</div>
                </div>
              </div>
            </template>

            <template #node-protocol="{ data }">
              <div class="protocol-node" :class="topoSettings.nodeStyle">
                <div class="protocol-content">
                  <div class="protocol-icon">
                    <el-icon><Grid /></el-icon>
                  </div>
                  <div class="protocol-label">{{ data.label }}</div>
                </div>
              </div>
            </template>
          </VueFlow>
        </div>
      </div>

      <!-- 表视图 - 卡片式布局 -->
      <div v-else-if="viewMode === 'table'" class="table-view">
        <div class="rules-cards-container">
          <div class="rules-grid">
            <div
              v-for="rule in filteredTableRules"
              :key="`${rule.table}-${rule.chain_name}-${rule.line_number}`"
              class="rule-card"
            >
              <el-card shadow="hover" class="rule-card-content">
                <!-- 卡片头部 -->
                <template #header>
                  <div class="rule-card-header">
                    <div class="rule-info">
                      <div class="rule-number">
                        <el-icon><Document /></el-icon>
                        #{{ rule.line_number }}
                      </div>
                      <div class="rule-stats">
                        <el-tag size="small" type="info">{{ rule.packets || 0 }} 包</el-tag>
                        <el-tag size="small" type="warning">{{ rule.bytes || '0B' }}</el-tag>
                      </div>
                    </div>
                    <div class="rule-actions">
                      <el-button type="primary" size="small" circle @click="editRuleFromDetail(rule)">
                        <el-icon><Edit /></el-icon>
                      </el-button>
                      <el-button type="danger" size="small" circle @click="deleteRuleFromDetail(rule)">
                        <el-icon><Delete /></el-icon>
                      </el-button>
                    </div>
                  </div>
                </template>

                <!-- 卡片主要内容 -->
                <div class="rule-card-body">
                  <!-- 链和表信息 -->
                  <div class="rule-chain-table">
                    <el-tag :type="getChainTagType(rule.chain_name)" size="small">
                      {{ rule.chain_name }}
                    </el-tag>
                    <el-icon><ArrowRight /></el-icon>
                    <el-tag :type="getTableTagType(rule.table)" size="small">
                      {{ rule.table?.toUpperCase() }}
                    </el-tag>
                  </div>

                  <!-- 目标动作 -->
                  <div class="rule-target">
                    <label>目标:</label>
                    <el-tag :type="getTargetTagType(rule.target)" size="medium">
                      {{ rule.target || '-' }}
                    </el-tag>
                  </div>

                  <!-- 网络信息 -->
                  <div class="rule-network">
                    <div class="network-item">
                      <label>协议:</label>
                      <span class="network-value">{{ rule.protocol || 'all' }}</span>
                    </div>
                    <div class="network-item">
                      <label>源地址:</label>
                      <span class="network-value" :title="rule.source">
                        {{ rule.source || '0.0.0.0/0' }}
                      </span>
                    </div>
                    <div class="network-item">
                      <label>目标地址:</label>
                      <span class="network-value" :title="rule.destination">
                        {{ rule.destination || '0.0.0.0/0' }}
                      </span>
                    </div>
                  </div>

                  <!-- 接口信息 -->
                  <div class="rule-interfaces" v-if="rule.in_interface || rule.out_interface">
                    <div class="interface-item" v-if="rule.in_interface && rule.in_interface !== '-'">
                      <el-icon><Download /></el-icon>
                      <el-tag type="info" size="small">{{ rule.in_interface }}</el-tag>
                    </div>
                    <div class="interface-arrow" v-if="rule.in_interface && rule.out_interface && rule.in_interface !== '-' && rule.out_interface !== '-'">
                      <el-icon><ArrowRight /></el-icon>
                    </div>
                    <div class="interface-item" v-if="rule.out_interface && rule.out_interface !== '-'">
                      <el-icon><Upload /></el-icon>
                      <el-tag type="warning" size="small">{{ rule.out_interface }}</el-tag>
                    </div>
                  </div>

                  <!-- 其他选项 -->
                  <div class="rule-options" v-if="rule.options">
                    <label>选项:</label>
                    <span class="options-text">{{ rule.options }}</span>
                  </div>
                </div>
              </el-card>
            </div>
          </div>

          <!-- 空状态 -->
          <div v-if="filteredTableRules.length === 0" class="empty-state">
            <el-empty description="没有找到匹配的规则">
              <el-button type="primary" @click="clearAllFilters">清空筛选条件</el-button>
            </el-empty>
          </div>
        </div>
      </div>

      <!-- 接口视图 -->
      <div v-else-if="viewMode === 'interface'" class="interface-view">
        <!-- 统计信息面板 -->
        <div class="stats-panel">
          <el-row :gutter="16">
            <el-col :span="6">
              <el-card class="stats-card">
                <div class="stats-content">
                  <div class="stats-icon">
                    <el-icon><Connection /></el-icon>
                  </div>
                  <div class="stats-info">
                    <div class="stats-number">{{ filteredInterfaceData.length }}</div>
                    <div class="stats-label">网络接口</div>
                  </div>
                </div>
              </el-card>
            </el-col>
            <el-col :span="6">
              <el-card class="stats-card">
                <div class="stats-content">
                  <div class="stats-icon active">
                    <el-icon><Check /></el-icon>
                  </div>
                  <div class="stats-info">
                    <div class="stats-number">{{ activeInterfacesCount }}</div>
                    <div class="stats-label">活跃接口</div>
                  </div>
                </div>
              </el-card>
            </el-col>
            <el-col :span="6">
              <el-card class="stats-card">
                <div class="stats-content">
                  <div class="stats-icon docker">
                    <el-icon><Box /></el-icon>
                  </div>
                  <div class="stats-info">
                    <div class="stats-number">{{ dockerInterfacesCount }}</div>
                    <div class="stats-label">Docker接口</div>
                  </div>
                </div>
              </el-card>
            </el-col>
            <el-col :span="6">
              <el-card class="stats-card">
                <div class="stats-content">
                  <div class="stats-icon rules">
                    <el-icon><List /></el-icon>
                  </div>
                  <div class="stats-info">
                    <div class="stats-number">{{ totalInterfaceRules }}</div>
                    <div class="stats-label">关联规则</div>
                  </div>
                </div>
              </el-card>
            </el-col>
          </el-row>
        </div>

        <!-- 接口筛选面板 -->
        <div class="interface-filters">
          <el-card>
            <div class="filter-content">
              <div class="filter-group">
                <label class="filter-label">接口类型:</label>
                <div class="filter-tags">
                  <el-tag
                    v-for="type in availableInterfaceTypes"
                    :key="type"
                    :type="selectedInterfaceTypes.includes(type) ? 'primary' : 'info'"
                    :effect="selectedInterfaceTypes.includes(type) ? 'dark' : 'plain'"
                    @click="toggleInterfaceType(type)"
                    class="filter-tag"
                  >
                    {{ type.toUpperCase() }}
                  </el-tag>
                </div>
              </div>

              <div class="filter-group">
                <label class="filter-label">接口状态:</label>
                <div class="filter-tags">
                  <el-tag
                    :type="interfaceStatusFilter === 'up' ? 'success' : 'info'"
                    :effect="interfaceStatusFilter === 'up' ? 'dark' : 'plain'"
                    @click="toggleInterfaceStatus('up')"
                    class="filter-tag"
                  >
                    启用
                  </el-tag>
                  <el-tag
                    :type="interfaceStatusFilter === 'down' ? 'danger' : 'info'"
                    :effect="interfaceStatusFilter === 'down' ? 'dark' : 'plain'"
                    @click="toggleInterfaceStatus('down')"
                    class="filter-tag"
                  >
                    禁用
                  </el-tag>
                  <el-tag
                    :type="interfaceStatusFilter === 'docker' ? 'warning' : 'info'"
                    :effect="interfaceStatusFilter === 'docker' ? 'dark' : 'plain'"
                    @click="toggleInterfaceStatus('docker')"
                    class="filter-tag"
                  >
                    Docker
                  </el-tag>
                </div>
              </div>

              <div class="filter-actions">
                <el-button size="small" @click="clearInterfaceFilters">
                  <el-icon><Delete /></el-icon>
                  清空筛选
                </el-button>
              </div>
            </div>
          </el-card>
        </div>

        <!-- 接口卡片列表 -->
        <div class="interfaces-container">
          <div class="interfaces-grid">
            <div
              v-for="iface in filteredInterfaceData"
              :key="iface.name"
              class="interface-card"
            >
              <el-card class="interface-card-content" :class="{ 'docker-interface': iface.is_docker }">
                <template #header>
                  <div class="interface-header">
                    <div class="interface-title">
                      <el-icon class="interface-icon">
                        <Connection v-if="iface.type === 'ethernet'" />
                        <Box v-else-if="iface.is_docker" />
                        <Monitor v-else />
                      </el-icon>
                      <h3>{{ iface.name }}</h3>
                    </div>
                    <div class="interface-badges">
                      <el-tag :type="iface.is_up ? 'success' : 'danger'" size="small">
                        {{ iface.is_up ? '启用' : '禁用' }}
                      </el-tag>
                      <el-tag v-if="iface.is_docker" type="warning" size="small">
                        {{ iface.docker_type || 'Docker' }}
                      </el-tag>
                    </div>
                  </div>
                </template>
                
                <div class="interface-content">
                  <!-- 基本信息 -->
                  <div class="interface-basic-info">
                    <div class="info-row">
                      <span class="info-label">类型:</span>
                      <span class="info-value">{{ iface.type }}</span>
                    </div>
                    <div class="info-row">
                      <span class="info-label">状态:</span>
                      <span class="info-value">{{ iface.state }}</span>
                    </div>
                    <div class="info-row">
                      <span class="info-label">MTU:</span>
                      <span class="info-value">{{ iface.mtu }}</span>
                    </div>
                  </div>

                  <!-- 网络信息 -->
                  <div class="interface-network-info">
                    <div class="network-section">
                      <h4>网络地址</h4>
                      <div class="address-list">
                        <el-tag
                          v-for="ip in iface.ip_addresses"
                          :key="ip"
                          type="info"
                          size="small"
                          class="address-tag"
                        >
                          {{ ip }}
                        </el-tag>
                        <span v-if="!iface.ip_addresses || iface.ip_addresses.length === 0" class="no-address">
                          无IP地址
                        </span>
                      </div>
                    </div>
                    
                    <div class="network-section" v-if="iface.mac_address">
                      <h4>MAC地址</h4>
                      <code class="mac-address">{{ iface.mac_address }}</code>
                    </div>
                  </div>

                  <!-- 规则统计 -->
                  <div class="interface-rules-stats">
                    <div class="stats-header">
                      <h4>规则统计</h4>
                      <el-tag v-if="hasActiveFilters" type="info" size="small">已筛选</el-tag>
                    </div>
                    <div class="rules-grid">
                      <div class="rule-stat-item">
                        <div class="rule-stat-icon input">
                          <el-icon><Download /></el-icon>
                        </div>
                        <div class="rule-stat-info">
                          <div class="rule-stat-number">{{ getInterfaceRuleCount(iface.name, 'in') }}</div>
                          <div class="rule-stat-label">输入规则</div>
                        </div>
                      </div>
                      <div class="rule-stat-item">
                        <div class="rule-stat-icon output">
                          <el-icon><Upload /></el-icon>
                        </div>
                        <div class="rule-stat-info">
                          <div class="rule-stat-number">{{ getInterfaceRuleCount(iface.name, 'out') }}</div>
                          <div class="rule-stat-label">输出规则</div>
                        </div>
                      </div>
                      <div class="rule-stat-item">
                        <div class="rule-stat-icon forward">
                          <el-icon><Share /></el-icon>
                        </div>
                        <div class="rule-stat-info">
                          <div class="rule-stat-number">{{ getInterfaceRuleCount(iface.name, 'forward') }}</div>
                          <div class="rule-stat-label">转发规则</div>
                        </div>
                      </div>
                      <div class="rule-stat-item total">
                        <div class="rule-stat-icon total">
                          <el-icon><List /></el-icon>
                        </div>
                        <div class="rule-stat-info">
                          <div class="rule-stat-number">{{ 
                            getInterfaceRuleCount(iface.name, 'in') + 
                            getInterfaceRuleCount(iface.name, 'out') + 
                            getInterfaceRuleCount(iface.name, 'forward') 
                          }}</div>
                          <div class="rule-stat-label">总计</div>
                        </div>
                      </div>
                    </div>
                  </div>

                  <!-- 流量统计 -->
                  <div class="interface-traffic-stats" v-if="iface.statistics">
                    <h4>流量统计</h4>
                    <div class="traffic-grid">
                      <div class="traffic-item">
                        <span class="traffic-label">接收:</span>
                        <span class="traffic-value">
                          {{ formatBytes(iface.statistics.rx_bytes) }}
                          ({{ iface.statistics.rx_packets }} 包)
                        </span>
                      </div>
                      <div class="traffic-item">
                        <span class="traffic-label">发送:</span>
                        <span class="traffic-value">
                          {{ formatBytes(iface.statistics.tx_bytes) }}
                          ({{ iface.statistics.tx_packets }} 包)
                        </span>
                      </div>
                    </div>
                  </div>

                  <!-- 操作按钮 -->
                  <div class="interface-actions">
                    <el-button size="small" @click="viewInterfaceRules(iface.name)">
                      <el-icon><View /></el-icon>
                      查看规则
                    </el-button>
                  </div>
                </div>
              </el-card>
            </div>
          </div>

          <!-- 空状态 -->
          <div v-if="filteredInterfaceData.length === 0" class="empty-state">
            <el-empty description="没有找到匹配的网络接口">
              <el-button type="primary" @click="clearInterfaceFilters">清空筛选条件</el-button>
            </el-empty>
          </div>
        </div>
      </div>
    </div>

    <!-- 链详情对话框 -->
    <el-dialog
      v-model="showChainDialog"
      :title="`${selectedChain} 链详细规则`"
      width="90%"
      top="5vh"
      :close-on-click-modal="false"
    >
      <div class="chain-detail-content">
        <!-- 控制面板 -->
        <div class="chain-detail-controls">
          <div class="chain-detail-left-controls">
            <el-switch
              v-model="groupByChain"
              active-text="按链名分组"
              inactive-text="列表展示"
              @change="handleGroupModeChange"
            />
            <el-select 
              v-model="tableFilter" 
              placeholder="筛选表" 
              size="small" 
              clearable 
              style="width: 120px; margin-left: 12px"
            >
              <el-option label="全部" value="" />
              <el-option label="RAW" value="raw" />
              <el-option label="MANGLE" value="mangle" />
              <el-option label="NAT" value="nat" />
              <el-option label="FILTER" value="filter" />
            </el-select>
            <el-select 
              v-model="targetFilter" 
              placeholder="筛选目标" 
              size="small" 
              clearable 
              style="width: 120px; margin-left: 8px"
            >
              <el-option label="全部" value="" />
              <el-option label="ACCEPT" value="ACCEPT" />
              <el-option label="DROP" value="DROP" />
              <el-option label="REJECT" value="REJECT" />
              <el-option label="RETURN" value="RETURN" />
            </el-select>
          </div>
          <div class="chain-detail-right-controls">
            <el-button type="primary" size="small" @click="showAddRuleDialog = true">
              <el-icon><Plus /></el-icon>
              添加规则
            </el-button>
            <el-input
              v-model="ruleSearchText"
              placeholder="搜索规则..."
              style="width: 200px; margin-left: 12px"
              size="small"
              clearable
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
          </div>
        </div>
        
        <!-- 分组展示 -->
        <div v-if="groupByChain" class="grouped-rules">
          <div
            v-for="(group, chainName) in groupedRules"
            :key="chainName"
            class="rule-group"
          >
            <div class="group-header">
              <div class="group-title">
                <h4>{{ chainName || '未指定链' }}</h4>
                <el-tag :type="getChainTagType(chainName)" size="small">{{ group.length }} 条规则</el-tag>
              </div>
              <div class="group-stats">
                <span class="stat-item">表: {{ getTablesInGroup(group).join(', ') }}</span>
              </div>
            </div>
            <el-table 
              :data="group" 
              stripe 
              size="small" 
              class="chain-rules-table"
              :default-sort="{prop: 'line_number', order: 'ascending'}"
              style="width: 100%"
            >
              <el-table-column prop="line_number" label="行号" width="70" sortable align="center" />
              <el-table-column prop="table" label="表" width="70" align="center" sortable>
                <template #default="{ row }">
                  <el-tag :type="getTableTagType(row.table)" size="small">{{ row.table }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="target" label="目标" width="100" align="center" sortable>
                <template #default="{ row }">
                  <el-tag v-if="row.target" :type="getTargetTagType(row.target)" size="small">{{ row.target }}</el-tag>
                  <span v-else class="no-target">-</span>
                </template>
              </el-table-column>
              <el-table-column prop="protocol" label="协议" width="70" align="center" sortable />
              <el-table-column prop="source" label="源地址" min-width="120" show-overflow-tooltip sortable />
              <el-table-column prop="destination" label="目标地址" min-width="120" show-overflow-tooltip sortable />
              <el-table-column prop="in_interface" label="入接口" width="90" align="center" sortable>
                <template #default="{ row }">
                  <el-tag v-if="row.in_interface" type="info" size="small">{{ row.in_interface }}</el-tag>
                  <span v-else class="no-interface">-</span>
                </template>
              </el-table-column>
              <el-table-column prop="out_interface" label="出接口" width="90" align="center" sortable>
                <template #default="{ row }">
                  <el-tag v-if="row.out_interface" type="warning" size="small">{{ row.out_interface }}</el-tag>
                  <span v-else class="no-interface">-</span>
                </template>
              </el-table-column>
              <el-table-column prop="options" label="选项" min-width="150" show-overflow-tooltip />
              <el-table-column label="操作" width="100" fixed="right" align="center">
                <template #default="{ row }">
                  <el-button type="primary" size="small" @click="editRuleFromDetail(row)" circle>
                    <el-icon><Edit /></el-icon>
                  </el-button>
                  <el-button type="danger" size="small" @click="deleteRuleFromDetail(row)" circle>
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>
        
        <!-- 列表展示 -->
        <div v-else class="list-rules">
          <el-table 
            :data="filteredDetailRules" 
            stripe 
            :default-sort="{prop: 'line_number', order: 'ascending'}"
            style="width: 100%"
            max-height="600px"
          >
            <el-table-column prop="line_number" label="行号" width="70" sortable align="center" />
            <el-table-column prop="chain_name" label="链" width="100" align="center" sortable>
              <template #default="{ row }">
                <el-tag :type="getChainTagType(row.chain_name)" size="small">{{ row.chain_name }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="table" label="表" width="70" align="center" sortable>
              <template #default="{ row }">
                <el-tag :type="getTableTagType(row.table)" size="small">{{ row.table }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="target" label="目标" width="100" align="center" sortable>
              <template #default="{ row }">
                <el-tag v-if="row.target" :type="getTargetTagType(row.target)" size="small">{{ row.target }}</el-tag>
                <span v-else class="no-target">-</span>
              </template>
            </el-table-column>
            <el-table-column prop="protocol" label="协议" width="70" align="center" sortable />
            <el-table-column prop="source" label="源地址" min-width="120" show-overflow-tooltip sortable />
            <el-table-column prop="destination" label="目标地址" min-width="120" show-overflow-tooltip sortable />
            <el-table-column prop="in_interface" label="入接口" width="90" align="center" sortable>
              <template #default="{ row }">
                <el-tag v-if="row.in_interface" type="info" size="small">{{ row.in_interface }}</el-tag>
                <span v-else class="no-interface">-</span>
              </template>
            </el-table-column>
            <el-table-column prop="out_interface" label="出接口" width="90" align="center" sortable>
              <template #default="{ row }">
                <el-tag v-if="row.out_interface" type="warning" size="small">{{ row.out_interface }}</el-tag>
                <span v-else class="no-interface">-</span>
              </template>
            </el-table-column>
            <el-table-column prop="options" label="选项" min-width="150" show-overflow-tooltip />
            <el-table-column label="操作" width="100" fixed="right" align="center">
              <template #default="{ row }">
                <el-button type="primary" size="small" @click="editRuleFromDetail(row)" circle>
                  <el-icon><Edit /></el-icon>
                </el-button>
                <el-button type="danger" size="small" @click="deleteRuleFromDetail(row)" circle>
                  <el-icon><Delete /></el-icon>
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </div>
    </el-dialog>

    <!-- 添加规则对话框 -->
    <el-dialog
      v-model="showAddRuleDialog"
      title="添加规则"
      width="800px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="ruleFormRef"
        :model="ruleForm"
        :rules="ruleFormRules"
        label-width="120px"
        class="rule-form"
      >
        <!-- 基础信息 -->
        <el-divider content-position="left">基础信息</el-divider>
        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="链名" prop="chain_name">
              <el-select v-model="ruleForm.chain_name" placeholder="选择链" @change="handleChainChange">
                <el-option label="PREROUTING" value="PREROUTING" />
                <el-option label="INPUT" value="INPUT" />
                <el-option label="FORWARD" value="FORWARD" />
                <el-option label="OUTPUT" value="OUTPUT" />
                <el-option label="POSTROUTING" value="POSTROUTING" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="表" prop="table">
              <el-select v-model="ruleForm.table" placeholder="选择表" :disabled="!ruleForm.chain_name">
                <el-option 
                  v-for="table in availableTables" 
                  :key="table" 
                  :label="table.toUpperCase()" 
                  :value="table" 
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="目标" prop="target">
              <el-select v-model="ruleForm.target" placeholder="选择目标">
                <el-option label="ACCEPT" value="ACCEPT" />
                <el-option label="DROP" value="DROP" />
                <el-option label="REJECT" value="REJECT" />
                <el-option label="RETURN" value="RETURN" />
                <el-option label="LOG" value="LOG" />
                <el-option label="DNAT" value="DNAT" />
                <el-option label="SNAT" value="SNAT" />
                <el-option label="MASQUERADE" value="MASQUERADE" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        
        <!-- 协议和端口 -->
        <el-divider content-position="left">协议和端口</el-divider>
        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="协议" prop="protocol">
              <el-select v-model="ruleForm.protocol" placeholder="选择协议">
                <el-option label="ALL" value="all" />
                <el-option label="TCP" value="tcp" />
                <el-option label="UDP" value="udp" />
                <el-option label="ICMP" value="icmp" />
                <el-option label="AH" value="ah" />
                <el-option label="ESP" value="esp" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="源端口" prop="source_port">
              <el-input v-model="ruleForm.source_port" placeholder="如: 80, 443, 1024:65535" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="目标端口" prop="destination_port">
              <el-input v-model="ruleForm.destination_port" placeholder="如: 80, 443, 22" />
            </el-form-item>
          </el-col>
        </el-row>
        
        <!-- 地址信息 -->
        <el-divider content-position="left">地址信息</el-divider>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="源地址" prop="source_ip">
              <el-input v-model="ruleForm.source_ip" placeholder="如: 192.168.1.0/24, 0.0.0.0/0" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="目标地址" prop="destination_ip">
              <el-input v-model="ruleForm.destination_ip" placeholder="如: 192.168.1.100, 0.0.0.0/0" />
            </el-form-item>
          </el-col>
        </el-row>
        
        <!-- 接口信息 -->
        <el-divider content-position="left">接口信息</el-divider>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="入接口" prop="in_interface">
              <el-select v-model="ruleForm.in_interface" placeholder="选择入接口" clearable>
                <el-option 
                  v-for="iface in interfaces" 
                  :key="iface.name" 
                  :label="iface.name" 
                  :value="iface.name" 
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="出接口" prop="out_interface">
              <el-select v-model="ruleForm.out_interface" placeholder="选择出接口" clearable>
                <el-option 
                  v-for="iface in interfaces" 
                  :key="iface.name" 
                  :label="iface.name" 
                  :value="iface.name" 
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        
        <!-- 其他选项 -->
        <el-divider content-position="left">其他选项</el-divider>
        <el-row :gutter="20">
          <el-col :span="24">
            <el-form-item label="其他选项" prop="options">
              <el-input 
                v-model="ruleForm.options" 
                type="textarea" 
                :rows="2" 
                placeholder="如: --state NEW,ESTABLISHED, --limit 1/sec" 
              />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showAddRuleDialog = false">取消</el-button>
          <el-button type="primary" @click="submitRuleForm" :loading="rulesLoading">确定</el-button>
        </div>
      </template>
    </el-dialog>
    
    <!-- 拓扑设置对话框 -->
    <el-dialog
      v-model="showTopoSettingsDialog"
      title="拓扑图设置"
      width="500px"
      destroy-on-close
    >
      <el-form :model="topoSettings" label-position="top">
        <el-form-item label="节点样式">
          <el-radio-group v-model="topoSettings.nodeStyle">
            <el-radio-button label="flat">扁平</el-radio-button>
            <el-radio-button label="gradient">渐变</el-radio-button>
            <el-radio-button label="glass">玻璃</el-radio-button>
          </el-radio-group>
        </el-form-item>
        
        <el-divider>交互设置</el-divider>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="启用拖拽">
              <el-switch v-model="topoSettings.enableDrag" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="吸附网格">
              <el-switch v-model="topoSettings.snapToGrid" />
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="显示小地图">
              <el-switch v-model="topoSettings.showMinimap" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="边缘动画">
              <el-switch v-model="topoSettings.animateEdges" />
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="暗色模式">
              <el-switch v-model="topoSettings.darkMode" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="显示标签">
              <el-switch v-model="topoSettings.showLabels" />
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-form-item label="节点间距">
          <el-slider v-model="topoSettings.nodeDistance" :min="50" :max="200" :step="10" show-stops />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="resetTopoSettings">重置默认</el-button>
          <el-button type="primary" @click="applyTopoSettings">应用设置</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed, watch, nextTick } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import { 
  Refresh, Plus, Edit, Delete, Search, Filter, Share, Grid, Connection, 
  Location, ArrowRight, Document, Download, Upload, Check, Box, List, Monitor, View, Setting
} from '@element-plus/icons-vue'
import { VueFlow, useVueFlow } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { Controls } from '@vue-flow/controls'
import { MiniMap } from '@vue-flow/minimap'
import { MarkerType, ConnectionMode, Position } from '@vue-flow/core'
import type { Node, Edge, Elements, NodeTypes } from '@vue-flow/core'
import '@vue-flow/core/dist/style.css'
import '@vue-flow/core/dist/theme-default.css'
import { apiService, networkAPI, tablesAPI } from '@/api'

// 响应式数据
const loading = ref(false)
const selectedInterface = ref('')
const viewMode = ref('chain')
const selectedChain = ref('')
const showChainDialog = ref(false)
const showAddRuleDialog = ref(false)
const showTopoSettingsDialog = ref(false)
const detailTitle = ref('')
const detailRules = ref([])
const groupByChain = ref(true)

// 拓扑图设置
const topoSettings = reactive({
  nodeStyle: 'gradient',
  enableDrag: true,
  snapToGrid: true,
  showMinimap: true,
  animateEdges: true,
  darkMode: false,
  showLabels: true,
  nodeDistance: 100
})

// 筛选相关数据
const activeFilterPanels = ref(['filters'])
const selectedInterfaces = ref<string[]>([])
const selectedProtocols = ref<string[]>([])
const selectedTargets = ref<string[]>([])
const ipRangeFilter = ref('')
const portRangeFilter = ref('')

// 接口视图筛选数据
const selectedInterfaceTypes = ref<string[]>([])
const interfaceStatusFilter = ref('')

// 可用选项
const availableProtocols = ref(['tcp', 'udp', 'icmp', 'all'])
const availableTargets = ref(['ACCEPT', 'DROP', 'REJECT', 'RETURN', 'MASQUERADE', 'SNAT', 'DNAT'])
const availableInterfaceTypes = ref(['ethernet', 'bridge', 'loopback', 'tunnel'])

// 规则管理相关状态
const rulesLoading = ref(false)
const ruleDialogVisible = ref(false)
const isEditRule = ref(false)
const ruleSearchText = ref('')
const rules = ref<any[]>([])
const tableFilter = ref('')
const targetFilter = ref('')
const ruleFormRef = ref<FormInstance>()

// 规则表单数据
const ruleForm = reactive({
  id: undefined as number | undefined,
  chain_name: '',
  table: '',
  target: '',
  protocol: 'all',
  source_ip: '',
  destination_ip: '',
  source_port: '',
  destination_port: '',
  in_interface: '',
  out_interface: '',
  options: ''
})

// 规则表单验证规则
const ruleFormRules = {
  chain_name: [
    { required: true, message: '请选择链名', trigger: 'change' }
  ],
  table: [
    { required: true, message: '请选择表', trigger: 'change' }
  ],
  target: [
    { required: true, message: '请选择目标', trigger: 'change' }
  ],
  protocol: [
    { required: true, message: '请选择协议', trigger: 'change' }
  ]
}

// 可用表的计算属性
const availableTables = computed(() => {
  if (!ruleForm.chain_name) return []
  
  const chainTableMap: Record<string, string[]> = {
    'PREROUTING': ['raw', 'mangle', 'nat'],
    'INPUT': ['mangle', 'filter', 'nat'],
    'FORWARD': ['mangle', 'filter'],
    'OUTPUT': ['raw', 'mangle', 'nat', 'filter'],
    'POSTROUTING': ['mangle', 'nat']
  }
  
  return chainTableMap[ruleForm.chain_name] || []
})

// 数据
const interfaces = ref([])
const flowElements = ref<Elements>([])

// 获取链的规则数量
const getChainRuleCount = (chainName: string) => {
  if (!chainTableData.value || !chainTableData.value.chains) {
    console.log(`获取${chainName}规则数量失败: 数据未加载`)
    return 0
  }
  
  const chain = chains.value.find((c: any) => c.name === chainName)
  const count = chain ? (chain.rules || []).length : 0
  
  console.log(`${chainName}链规则数量:`, count, '链数据:', chain)
  return count
}

// 获取筛选后的链规则数量
const getFilteredChainRuleCount = (chainName: string): number => {
  // 从筛选后的表规则中统计指定链的规则数量
  const filteredCount = filteredTableRules.value.filter((rule: any) => rule.chain_name === chainName).length
  console.log(`${chainName}链筛选后规则数量:`, filteredCount)
  return filteredCount
}

// 选择链和表
const selectChainTable = (chainName: string, tableName: string) => {
  console.log('选择链和表:', { chainName, tableName })
  selectedChain.value = chainName
  showChainDialog.value = true
  detailTitle.value = `${chainName} - ${tableName.toUpperCase()} 表详细规则`
  
  // 获取特定链和表的规则
  const chain = chains.value.find(c => c.name === chainName)
  console.log('找到的链:', chain)
  
  if (chain) {
    // 从链的rules数组中筛选出指定表的规则
    const chainRules = chain.rules || []
    const filteredRules = chainRules.filter(rule => rule.table === tableName)
    
    console.log('链的所有规则:', chainRules)
    console.log('筛选后的规则:', filteredRules)
    
    // 为规则添加更多详细信息
    detailRules.value = filteredRules.map((rule, index) => ({
      ...rule,
      line_number: rule.line_number || (index + 1).toString(),
      chain_name: rule.chain_name || chainName,
      table: rule.table || tableName,
      target: rule.target || extractTarget(rule.rule_text || ''),
      source: rule.source || '-',
      destination: rule.destination || '-',
      protocol: rule.protocol || extractProtocol(rule.rule_text || ''),
      in_interface: rule.in_interface || '-',
      out_interface: rule.out_interface || '-',
      options: rule.options || rule.rule_text || ''
    }))
  } else {
    console.log('未找到指定链')
    detailRules.value = []
  }
  
  console.log('最终设置的规则数据:', detailRules.value)
}

// 处理链变化
const handleChainChange = () => {
  // 重置表选择
  ruleForm.table = ''
  // 如果当前选择的表不在可用表中，清空
  if (ruleForm.table && !availableTables.value.includes(ruleForm.table)) {
    ruleForm.table = ''
  }
}

// 辅助函数：从规则文本中提取目标
const extractTarget = (ruleText: string): string => {
  const targetMatch = ruleText.match(/-j\s+(\w+)/)
  if (targetMatch) return targetMatch[1]
  
  // 检查常见的目标关键词
  if (ruleText.includes('ACCEPT')) return 'ACCEPT'
  if (ruleText.includes('DROP')) return 'DROP'
  if (ruleText.includes('REJECT')) return 'REJECT'
  if (ruleText.includes('RETURN')) return 'RETURN'
  if (ruleText.includes('MASQUERADE')) return 'MASQUERADE'
  if (ruleText.includes('SNAT')) return 'SNAT'
  if (ruleText.includes('DNAT')) return 'DNAT'
  
  return '-'
}

// 辅助函数：从规则文本中提取协议
const extractProtocol = (ruleText: string): string => {
  const protocolMatch = ruleText.match(/-p\s+(\w+)/)
  return protocolMatch ? protocolMatch[1] : 'all'
}
const chainTableData = ref({
  chains: [],
  tables: [],
  interfaceRules: {}
})

// 计算属性
const chains = computed(() => chainTableData.value.chains || [])
const tables = computed(() => chainTableData.value.tables || [])
const interfaceData = computed(() => {
  return interfaces.value.map((iface: any) => ({
    ...iface,
    inRules: getInterfaceRuleCount(iface.name, 'in'),
    outRules: getInterfaceRuleCount(iface.name, 'out'),
    forwardRules: getInterfaceRuleCount(iface.name, 'forward')
  }))
})

// 按行号排序的详细规则
const sortedDetailRules = computed(() => {
  return [...detailRules.value].sort((a: any, b: any) => {
    const lineA = parseInt(a.line_number || '0', 10)
    const lineB = parseInt(b.line_number || '0', 10)
    return lineA - lineB
  })
})

// 按链名分组的规则
const groupedRules = computed(() => {
  const groups: Record<string, any[]> = {}
  sortedDetailRules.value.forEach((rule: any) => {
    const chainName = rule.chain_name || '未指定链'
    if (!groups[chainName]) {
      groups[chainName] = []
    }
    groups[chainName].push(rule)
  })
  
// 对每个分组内的规则按行号排序
      Object.keys(groups).forEach((chainName: string) => {
        groups[chainName].sort((a: any, b: any) => {
          const lineA = parseInt(a.line_number || '0', 10)
          const lineB = parseInt(b.line_number || '0', 10)
          return lineA - lineB
        })
      })
  
  return groups
})

// 筛选相关计算属性
const activeFiltersCount = computed(() => {
  let count = 0
  if (selectedInterfaces.value.length > 0) count++
  if (selectedProtocols.value.length > 0) count++
  if (selectedTargets.value.length > 0) count++
  if (ipRangeFilter.value) count++
  if (portRangeFilter.value) count++
  return count
})

const hasActiveFilters = computed(() => {
  return activeFiltersCount.value > 0
})

// 表视图筛选规则
const filteredTableRules = computed(() => {
  let allRules: any[] = []
  
  // 收集所有表中的所有规则
  tables.value.forEach((table: any) => {
    if (table.chains && Array.isArray(table.chains)) {
      table.chains.forEach((chain: any) => {
        if (chain.rules && Array.isArray(chain.rules)) {
          chain.rules.forEach((rule: any) => {
            allRules.push({
              ...rule,
              table: table.name,
              chain_name: chain.name,
              line_number: rule.line_number || allRules.length + 1
            })
          })
        }
      })
    }
  })
  
  // 应用筛选条件
  let filtered = allRules
  
  // 按接口筛选
  if (selectedInterfaces.value.length > 0) {
    filtered = filtered.filter((rule: any) => 
      selectedInterfaces.value.includes(rule.in_interface) ||
      selectedInterfaces.value.includes(rule.out_interface)
    )
  }
  
  // 按协议筛选
  if (selectedProtocols.value.length > 0) {
    filtered = filtered.filter((rule: any) => 
      selectedProtocols.value.includes(rule.protocol?.toLowerCase())
    )
  }
  
  // 按目标动作筛选
  if (selectedTargets.value.length > 0) {
    filtered = filtered.filter((rule: any) => 
      selectedTargets.value.includes(rule.target)
    )
  }
  
  // 按IP范围筛选
  if (ipRangeFilter.value) {
    const ipPattern = ipRangeFilter.value.toLowerCase()
    filtered = filtered.filter((rule: any) => 
      rule.source?.toLowerCase().includes(ipPattern) ||
      rule.destination?.toLowerCase().includes(ipPattern)
    )
  }
  
  // 按端口范围筛选
  if (portRangeFilter.value) {
    const portPattern = portRangeFilter.value.toLowerCase()
    filtered = filtered.filter((rule: any) => 
      rule.options?.toLowerCase().includes(portPattern)
    )
  }
  
  return filtered.sort((a: any, b: any) => {
    const lineA = parseInt(a.line_number || '0', 10)
    const lineB = parseInt(b.line_number || '0', 10)
    return lineA - lineB
  })
})

// 接口视图相关计算属性
const filteredInterfaceData = computed(() => {
  let filtered = [...interfaceData.value]
  
  // 按接口类型筛选
  if (selectedInterfaceTypes.value.length > 0) {
    filtered = filtered.filter((iface: any) => 
      selectedInterfaceTypes.value.includes(iface.type)
    )
  }
  
  // 按接口状态筛选
  if (interfaceStatusFilter.value) {
    switch (interfaceStatusFilter.value) {
      case 'up':
        filtered = filtered.filter((iface: any) => iface.is_up)
        break
      case 'down':
        filtered = filtered.filter((iface: any) => !iface.is_up)
        break
      case 'docker':
        filtered = filtered.filter((iface: any) => iface.is_docker)
        break
    }
  }
  
  // 应用全局筛选条件：如果设置了接口筛选，只显示被选中的接口
  if (selectedInterfaces.value.length > 0) {
    filtered = filtered.filter((iface: any) => 
      selectedInterfaces.value.includes(iface.name)
    )
  }
  
  return filtered
})

const activeInterfacesCount = computed(() => {
  return filteredInterfaceData.value.filter((iface: any) => iface.is_up).length
})

const dockerInterfacesCount = computed(() => {
  return filteredInterfaceData.value.filter((iface: any) => iface.is_docker).length
})

const totalInterfaceRules = computed(() => {
  return filteredInterfaceData.value.reduce((total: number, iface: any) => {
    return total + getInterfaceRuleCount(iface.name, 'in') + 
           getInterfaceRuleCount(iface.name, 'out') + 
           getInterfaceRuleCount(iface.name, 'forward')
  }, 0)
})

// 规则管理相关计算属性
const filteredDetailRules = computed(() => {
  let filtered = [...sortedDetailRules.value]
  
  // 按表筛选
  if (tableFilter.value) {
    filtered = filtered.filter((rule: any) => rule.table === tableFilter.value)
  }
  
  // 按目标筛选
  if (targetFilter.value) {
    filtered = filtered.filter((rule: any) => rule.target === targetFilter.value)
  }
  
  // 按接口筛选
  if (selectedInterfaces.value.length > 0) {
    filtered = filtered.filter((rule: any) => 
      selectedInterfaces.value.includes(rule.in_interface) ||
      selectedInterfaces.value.includes(rule.out_interface)
    )
  }
  
  // 按协议筛选
  if (selectedProtocols.value.length > 0) {
    filtered = filtered.filter((rule: any) => 
      selectedProtocols.value.includes(rule.protocol?.toLowerCase())
    )
  }
  
  // 按目标动作筛选
  if (selectedTargets.value.length > 0) {
    filtered = filtered.filter((rule: any) => 
      selectedTargets.value.includes(rule.target)
    )
  }
  
  // 按IP范围筛选
  if (ipRangeFilter.value) {
    const ipPattern = ipRangeFilter.value.toLowerCase()
    filtered = filtered.filter((rule: any) => 
      rule.source?.toLowerCase().includes(ipPattern) ||
      rule.destination?.toLowerCase().includes(ipPattern)
    )
  }
  
  // 按端口范围筛选
  if (portRangeFilter.value) {
    const portPattern = portRangeFilter.value.toLowerCase()
    filtered = filtered.filter((rule: any) => 
      rule.options?.toLowerCase().includes(portPattern) ||
      rule.source_port?.toLowerCase().includes(portPattern) ||
      rule.destination_port?.toLowerCase().includes(portPattern)
    )
  }
  
  // 按搜索文本筛选
  if (ruleSearchText.value) {
    const searchText = ruleSearchText.value.toLowerCase()
    filtered = filtered.filter((rule: any) => 
      rule.chain_name?.toLowerCase().includes(searchText) ||
      rule.target?.toLowerCase().includes(searchText) ||
      rule.source?.toLowerCase().includes(searchText) ||
      rule.destination?.toLowerCase().includes(searchText) ||
      rule.protocol?.toLowerCase().includes(searchText) ||
      rule.in_interface?.toLowerCase().includes(searchText) ||
      rule.out_interface?.toLowerCase().includes(searchText) ||
      rule.options?.toLowerCase().includes(searchText)
    )
  }
  
  return filtered
})

// 计算节点边缘位置
const getNodeEdgePosition = (node: any, targetNode: any, isSource: boolean = true) => {
  if (!node?.position || !targetNode?.position) {
    return { x: 0, y: 0 }
  }

  // 节点尺寸（根据实际节点大小调整）
  const nodeWidth = 120
  const nodeHeight = 80

  // 节点中心位置
  const centerX = node.position.x + nodeWidth / 2
  const centerY = node.position.y + nodeHeight / 2

  // 目标节点中心位置
  const targetCenterX = targetNode.position.x + nodeWidth / 2
  const targetCenterY = targetNode.position.y + nodeHeight / 2

  // 计算方向向量
  const dx = targetCenterX - centerX
  const dy = targetCenterY - centerY
  const distance = Math.sqrt(dx * dx + dy * dy)

  if (distance === 0) {
    return { x: centerX, y: centerY }
  }

  // 标准化方向向量
  const unitX = dx / distance
  const unitY = dy / distance

  // 计算边缘交点
  let edgeX, edgeY

  // 计算与节点边界的交点
  const absUnitX = Math.abs(unitX)
  const absUnitY = Math.abs(unitY)

  if (absUnitX > absUnitY) {
    // 主要是水平方向，与左右边界相交
    const halfWidth = nodeWidth / 2
    edgeX = centerX + (unitX > 0 ? halfWidth : -halfWidth)
    edgeY = centerY + (unitY * halfWidth / absUnitX)
  } else {
    // 主要是垂直方向，与上下边界相交
    const halfHeight = nodeHeight / 2
    edgeX = centerX + (unitX * halfHeight / absUnitY)
    edgeY = centerY + (unitY > 0 ? halfHeight : -halfHeight)
  }

  return { x: edgeX, y: edgeY }
}

// 初始化 Vue Flow 节点和边
const initializeFlowElements = () => {
  console.log('开始初始化流程图元素...')
  console.log('当前数据状态:', {
    chainTableData: chainTableData.value,
    hasChains: chainTableData.value?.chains?.length > 0,
    chainsLength: chainTableData.value?.chains?.length || 0,
    activeFilters: activeFiltersCount.value,
    topoSettings: topoSettings
  })
  
  // 确保数据已加载
  if (!chainTableData.value || !chainTableData.value.chains || !Array.isArray(chainTableData.value.chains)) {
    console.log('数据尚未加载或格式异常，跳过流程图初始化')
    console.log('chainTableData.value:', chainTableData.value)
    return
  }
  
  // 获取Vue Flow实例
  const { fitView } = useVueFlow()
  
  // 根据布局模式确定节点位置
  let nodePositions: Record<string, {x: number, y: number}> = {}
  
  // 根据图片中的布局设置节点位置
  nodePositions = {
    'input': { x: 620, y: 150 },
    'local-process': { x: 880, y: 150 },
    'output': { x: 1140, y: 150 },
    'external-entry': { x: 140, y: 310 },
    'prerouting': { x: 400, y: 310 },
    'routing-decision': { x: 620, y: 310 },
    'forward': { x: 880, y: 310 },
    'postrouting': { x: 1140, y: 310 },
    'external-exit': { x: 1400, y: 310 }
  }

  const nodes: Node[] = [
    // INPUT 链
    {
      id: 'input',
      type: 'chain',
      position: nodePositions['input'],
      data: {
        label: 'INPUT',
        chainName: 'INPUT',
        chainType: 'input',
        tables: ['mangle', 'filter', 'nat'],
        ruleCount: getFilteredChainRuleCount('INPUT')
      }
    },
    // 本地处理
    {
      id: 'local-process',
      type: 'process',
      position: nodePositions['local-process'],
      data: { label: '本地处理' }
    },
    // OUTPUT 链
    {
      id: 'output',
      type: 'chain',
      position: nodePositions['output'],
      data: {
        label: 'OUTPUT',
        chainName: 'OUTPUT',
        chainType: 'output',
        tables: ['raw', 'mangle', 'nat', 'filter'],
        ruleCount: getFilteredChainRuleCount('OUTPUT')
      }
    },
    // 外部网络入口
    {
      id: 'external-entry',
      type: 'endpoint',
      position: nodePositions['external-entry'],
      data: { label: '外部网络', type: 'entry' }
    },
    // PREROUTING 链
    {
      id: 'prerouting',
      type: 'chain',
      position: nodePositions['prerouting'],
      data: {
        label: 'PREROUTING',
        chainName: 'PREROUTING',
        chainType: 'prerouting',
        tables: ['raw', 'mangle', 'nat'],
        ruleCount: getFilteredChainRuleCount('PREROUTING')
      }
    },
    // 路由决策
    {
      id: 'routing-decision',
      type: 'decision',
      position: nodePositions['routing-decision'],
      data: { label: '路由决策' }
    },
    // FORWARD 链
    {
      id: 'forward',
      type: 'chain',
      position: nodePositions['forward'],
      data: {
        label: 'FORWARD',
        chainName: 'FORWARD',
        chainType: 'forward',
        tables: ['mangle', 'filter'],
        ruleCount: getFilteredChainRuleCount('FORWARD')
      }
    },
    // POSTROUTING 链
    {
      id: 'postrouting',
      type: 'chain',
      position: nodePositions['postrouting'],
      data: {
        label: 'POSTROUTING',
        chainName: 'POSTROUTING',
        chainType: 'postrouting',
        tables: ['mangle', 'nat'],
        ruleCount: getFilteredChainRuleCount('POSTROUTING')
      }
    },
    // 外部网络出口
    {
      id: 'external-exit',
      type: 'endpoint',
      position: nodePositions['external-exit'],
      data: { label: '内部网络', type: 'exit' }
    }
  ]

  // 创建优化的边连接
  const createOptimizedEdge = (id: string, source: string, target: string, options: any = {}) => {
    // 查找源节点和目标节点
    const sourceNode = nodes.find(n => n.id === source)
    const targetNode = nodes.find(n => n.id === target)
    
    if (!sourceNode || !targetNode) {
      return null
    }
    
    // 计算边缘连接点
    const sourceEdge = getNodeEdgePosition(sourceNode, targetNode, true)
    const targetEdge = getNodeEdgePosition(targetNode, sourceNode, false)
    
    // 默认边样式
    const defaultStyle = { 
      stroke: '#4A90E2', 
      strokeWidth: 2
    }
    
    // 默认标记样式
    const defaultMarker = { 
      type: MarkerType.ArrowClosed, 
      color: '#4A90E2',
      width: 20,
      height: 20
    }
    
    // 合并选项
    const style = { ...defaultStyle, ...options.style }
    const markerEnd = { ...defaultMarker, ...options.markerEnd }
    
    // 创建边
    return {
      id,
      source,
      target,
      type: 'straight',
      style,
      markerEnd,
      label: options.label || '',
      labelStyle: options.labelStyle || { fill: '#2d3748', fontWeight: 'bold' },
      labelBgStyle: options.labelBgStyle || { fill: 'rgba(255, 255, 255, 0.8)', fillOpacity: 0.8 },
      animated: topoSettings.animateEdges,
      zIndex: options.zIndex || 1,
      // 使用边缘连接点
      sourceX: sourceEdge.x,
      sourceY: sourceEdge.y,
      targetX: targetEdge.x,
      targetY: targetEdge.y
    }
  }

  const edges: Edge[] = [
    // 入站数据包流向
    createOptimizedEdge('e1', 'external-entry', 'prerouting', {
      style: { stroke: '#4A90E2', strokeWidth: 2, strokeDasharray: '5,5' },
      label: topoSettings.showLabels ? '入站数据包' : '',
      zIndex: 1
    }),
    createOptimizedEdge('e2', 'prerouting', 'routing-decision', {
      zIndex: 1
    }),
    
    // 本地处理路径
    createOptimizedEdge('e3', 'routing-decision', 'input', {
      label: topoSettings.showLabels ? '本地处理' : '',
      zIndex: 2
    }),
    createOptimizedEdge('e4', 'input', 'local-process', {
      zIndex: 2
    }),
    createOptimizedEdge('e5', 'local-process', 'output', {
      zIndex: 2
    }),
    createOptimizedEdge('e6', 'output', 'postrouting', {
      label: topoSettings.showLabels ? '出站数据包' : '',
      zIndex: 2
    }),
    
    // 转发路径
    createOptimizedEdge('e7', 'routing-decision', 'forward', {
      style: { stroke: '#4A90E2', strokeWidth: 2, strokeDasharray: '5,5' },
      label: topoSettings.showLabels ? '转发处理' : '',
      zIndex: 2
    }),
    createOptimizedEdge('e8', 'forward', 'postrouting', {
      style: { stroke: '#4A90E2', strokeWidth: 2, strokeDasharray: '5,5' },
      zIndex: 2
    }),
    
    // 最终出口
    createOptimizedEdge('e9', 'postrouting', 'external-exit', {
      style: { stroke: '#4A90E2', strokeWidth: 2, strokeDasharray: '5,5' },
      label: topoSettings.showLabels ? '出站数据包' : '',
      zIndex: 1
    })
  ].filter(edge => edge !== null) as Edge[]

  flowElements.value = [...nodes, ...edges]
}

// 节点点击事件
const onNodeClick = (event: any) => {
  const node = event.node
  if (node.type === 'chain') {
    selectChain(node.data.chainName)
  }
}

// 边点击事件
const onEdgeClick = (event: any) => {
  console.log('Edge clicked:', event.edge)
}

// 应用拓扑布局
const applyTopoLayout = () => {
  const nodes = flowElements.value.filter(el => el.type !== 'edge') as Node[]
  const edges = flowElements.value.filter(el => el.type === 'edge') as Edge[]
  
  // 根据图片中的布局设置节点位置
  const positions = {
    'input': { x: 620, y: 150 },
    'local-process': { x: 880, y: 150 },
    'output': { x: 1140, y: 150 },
    'external-entry': { x: 140, y: 310 },
    'prerouting': { x: 400, y: 310 },
    'routing-decision': { x: 620, y: 310 },
    'forward': { x: 880, y: 310 },
    'postrouting': { x: 1140, y: 310 },
    'external-exit': { x: 1400, y: 310 }
  }
  
  nodes.forEach(node => {
    if (positions[node.id]) {
      node.position = positions[node.id]
    }
  })
  
  // 更新流程图
  flowElements.value = [...nodes, ...edges]
}

// 应用拓扑设置
const applyTopoSettings = () => {
  // 应用布局
  applyTopoLayout()
  
  // 更新边的动画状态
  const nodes = flowElements.value.filter(el => el.type !== 'edge') as Node[]
  const edges = flowElements.value.filter(el => el.type === 'edge') as Edge[]
  
  edges.forEach(edge => {
    edge.animated = topoSettings.animateEdges
    
    // 更新标签显示
    if (edge.id === 'e4' || edge.id === 'e8') {
      if (topoSettings.showLabels) {
        edge.label = edge.id === 'e4' ? '本机设备' : '非本机设备'
      } else {
        edge.label = ''
      }
    }
  })
  
  // 更新节点样式
  nodes.forEach(node => {
    if (node.type === 'chain' && node.data) {
      node.data.nodeStyle = topoSettings.nodeStyle
    }
  })
  
  // 更新流程图
  flowElements.value = [...nodes, ...edges]
  
  // 获取Vue Flow实例并适应视图
  nextTick(() => {
    const { fitView } = useVueFlow()
    fitView({ padding: 0.2 })
  })
  
  // 关闭对话框
  showTopoSettingsDialog.value = false
  
  ElMessage.success('拓扑设置已应用')
}

// 重置拓扑设置
const resetTopoSettings = () => {
  Object.assign(topoSettings, {
    layoutMode: 'horizontal',
    nodeStyle: 'gradient',
    enableDrag: true,
    snapToGrid: true,
    showMinimap: true,
    animateEdges: true,
    darkMode: false,
    showLabels: true,
    nodeDistance: 100
  })
  
  // 重新初始化流程图
  initializeFlowElements()
  
  ElMessage.success('拓扑设置已重置为默认值')
}

// 选择链
const selectChain = (chainName: string) => {
  console.log('选择链:', chainName)
  selectedChain.value = chainName
  const chain = chains.value.find((c: any) => c.name === chainName)
  console.log('找到的链数据:', chain)
  
  if (chain) {
    // 检查是否有筛选条件
    const hasFilters = selectedInterfaces.value.length > 0 || 
                      selectedProtocols.value.length > 0 || 
                      selectedTargets.value.length > 0 || 
                      ipRangeFilter.value.trim() !== '' || 
                      portRangeFilter.value.trim() !== ''
    
    const filterStatus = hasFilters ? ' (已筛选)' : ''
    detailTitle.value = `${chainName} 链详细规则${filterStatus}`
    
    // 使用筛选后的规则数据，而不是原始的链规则数据
    const filteredChainRules = filteredTableRules.value.filter((rule: any) => rule.chain_name === chainName)
    console.log(`${chainName}链筛选后的规则数据:`, filteredChainRules)
    
    // 处理规则数据，确保格式正确
    detailRules.value = filteredChainRules.map((rule: any, index: number) => ({
      ...rule,
      line_number: rule.line_number || (index + 1).toString(),
      chain_name: rule.chain_name || chainName,
      table: rule.table || 'filter',
      target: rule.target || extractTarget(rule.rule_text || ''),
      source: rule.source || '-',
      destination: rule.destination || '-',
      protocol: rule.protocol || extractProtocol(rule.rule_text || ''),
      in_interface: rule.in_interface || '-',
      out_interface: rule.out_interface || '-',
      options: rule.options || rule.rule_text || ''
    }))
    
    console.log('处理后的规则数据:', detailRules.value)
    showChainDialog.value = true
  } else {
    console.log('未找到指定链')
  }
}

// 选择链在表中
const selectChainInTable = (tableName: string, chainName: string) => {
  console.log('选择表中的链:', { tableName, chainName })
  const table = tables.value.find((t: any) => t.name === tableName)
  console.log('找到的表:', table)
  
  if (table) {
    const chain = table.chains.find((c: any) => c.name === chainName)
    console.log('找到的链:', chain)
    
    if (chain) {
      detailTitle.value = `${tableName.toUpperCase()}.${chainName} 详细规则`
      
      // 处理规则数据，确保格式正确
      const chainRules = chain.rules || []
      detailRules.value = chainRules.map((rule: any, index: number) => ({
        ...rule,
        line_number: rule.line_number || (index + 1).toString(),
        chain_name: rule.chain_name || chainName,
        table: rule.table || tableName,
        target: rule.target || extractTarget(rule.rule_text || ''),
        source: rule.source || '-',
        destination: rule.destination || '-',
        protocol: rule.protocol || extractProtocol(rule.rule_text || ''),
        in_interface: rule.in_interface || '-',
        out_interface: rule.out_interface || '-',
        options: rule.options || rule.rule_text || ''
      }))
      
      console.log('处理后的规则数据:', detailRules.value)
      showChainDialog.value = true
    } else {
      console.log('在表中未找到指定链')
    }
  } else {
    console.log('未找到指定表')
  }
}

// 关闭链详情对话框
const closeChainDialog = () => {
  showChainDialog.value = false
  detailRules.value = []
  groupByChain.value = true
}

// 调试函数：检查数据状态
const debugDataState = () => {
  console.log('=== 数据状态调试 ===')
  console.log('chainTableData:', chainTableData.value)
  console.log('chains:', chains.value)
  console.log('tables:', tables.value)
  console.log('detailRules:', detailRules.value)
  console.log('showChainDialog:', showChainDialog.value)
  console.log('selectedChain:', selectedChain.value)
  console.log('==================')
}

// 处理分组模式变化
const handleGroupModeChange = () => {
  // 分组模式切换时的处理逻辑
  console.log('分组模式切换:', groupByChain.value)
}

// 获取分组中的表名列表
const getTablesInGroup = (rules: any[]) => {
  const tables = new Set(rules.map(rule => rule.table).filter(Boolean))
  return Array.from(tables)
}

// 获取表的标签类型
const getTableTagType = (tableName: string) => {
  const types: Record<string, string> = {
    'raw': 'info',
    'mangle': 'warning', 
    'nat': 'success',
    'filter': 'danger'
  }
  return types[tableName] || 'default'
}

// 获取目标的标签类型
const getTargetTagType = (target: string) => {
  const types: Record<string, string> = {
    'ACCEPT': 'success',
    'DROP': 'danger',
    'REJECT': 'warning',
    'RETURN': 'info',
    'MASQUERADE': 'primary',
    'SNAT': 'primary',
    'DNAT': 'primary'
  }
  return types[target] || 'default'
}

const getChainTagType = (chainName: string) => {
  const types: Record<string, string> = {
    'PREROUTING': 'primary',
    'INPUT': 'success',
    'FORWARD': 'warning',
    'OUTPUT': 'info',
    'POSTROUTING': 'danger'
  }
  return types[chainName] || 'default'
}

// 格式化日期
const formatDate = (dateString: string) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleString('zh-CN')
}

const getInterfaceRuleCount = (interfaceName: string, direction: string) => {
  // 使用筛选后的规则数据进行统计
  const allRules = filteredTableRules.value
  
  return allRules.filter((rule: any) => {
    if (direction === 'in') {
      return rule.InInterface === interfaceName || rule.in_interface === interfaceName
    } else if (direction === 'out') {
      return rule.OutInterface === interfaceName || rule.out_interface === interfaceName
    } else if (direction === 'forward') {
      return rule.chain_name === 'FORWARD' && 
             (rule.InInterface === interfaceName || rule.OutInterface === interfaceName ||
              rule.in_interface === interfaceName || rule.out_interface === interfaceName)
    }
    return false
  }).length
}

// 自动比对并同步系统规则
const autoSyncSystemRules = async () => {
  console.log('[DEBUG] autoSyncSystemRules called')
  try {
    // 先比对系统规则和数据库规则
    const compareResult = await apiService.compareSystemAndDatabaseRules()
    console.log('[DEBUG] Compare result:', compareResult.data)
    
    if (!compareResult.data.consistent) {
      console.log('[DEBUG] Rules are inconsistent, auto syncing...')
      // 如果不一致，自动同步
      await apiService.syncSystemRules()
      console.log('[DEBUG] Auto sync completed')
      return true // 返回true表示进行了同步
    } else {
      console.log('[DEBUG] Rules are consistent, no sync needed')
      return false // 返回false表示无需同步
    }
  } catch (error) {
    console.error('[ERROR] Failed to auto sync system rules:', error)
    throw error
  }
}

// 同步系统规则 - 复用Rules.vue中的实现（保留原方法以备手动调用）
const syncSystemRules = async () => {
  console.log('[DEBUG] syncSystemRules called')
  try {
    loading.value = true
    await apiService.syncSystemRules()
    ElMessage.success('系统规则同步成功')
    // 同步成功后重新加载数据
    await loadChainTableData()
  } catch (error) {
    console.error('[ERROR] Failed to sync system rules:', error)
    ElMessage.error('同步系统规则失败')
  } finally {
    loading.value = false
  }
}

// 规则管理相关方法
const loadRules = async () => {
  rulesLoading.value = true
  try {
    const response = await apiService.getRules()
    rules.value = response.data
  } catch (error) {
    console.error('Failed to load rules:', error)
    ElMessage.error('加载规则失败')
  } finally {
    rulesLoading.value = false
  }
}

const openAddRuleDialog = (chainName?: string) => {
  isEditRule.value = false
  showAddRuleDialog.value = true
  resetRuleForm()
  
  // 如果指定了链名，自动设置
  if (chainName) {
    ruleForm.chain_name = chainName
    // 触发链变化处理
    handleChainChange()
  }
}

const editRule = (rule: any) => {
  isEditRule.value = true
  ruleDialogVisible.value = true
  Object.assign(ruleForm, rule)
}

const deleteRule = async (rule: any) => {
  try {
    await ElMessageBox.confirm('确定要删除这条规则吗？', '确认删除', {
      type: 'warning'
    })
    
    await apiService.deleteRule(rule.id!)
    ElMessage.success('删除成功')
    await loadRules()
    await refreshData() // 刷新五链四表数据
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const submitRuleForm = async () => {
  if (!ruleFormRef.value) return
  
  await ruleFormRef.value.validate(async (valid) => {
    if (valid) {
      try {
        if (isEditRule.value) {
          await apiService.updateRule(ruleForm.id!, ruleForm)
          ElMessage.success('更新成功')
        } else {
          await apiService.addRule(ruleForm)
          ElMessage.success('添加成功')
        }
        showAddRuleDialog.value = false
        await loadRules()
        await refreshData() // 刷新五链四表数据
      } catch (error) {
        ElMessage.error(isEditRule.value ? '更新失败' : '添加失败')
      }
    }
  })
}

const resetRuleForm = () => {
  Object.assign(ruleForm, {
    id: undefined,
    chain_name: '',
    target: '',
    protocol: '',
    source_ip: '',
    destination_ip: '',
    destination_port: ''
  })
  ruleFormRef.value?.clearValidate()
}

// 从详细规则页面编辑规则
const editRuleFromDetail = (rule: any) => {
  // 将详细规则数据转换为规则表单数据
  isEditRule.value = true
  showAddRuleDialog.value = true
  Object.assign(ruleForm, {
    id: rule.id,
    chain_name: rule.chain_name || '',
    target: rule.target || '',
    protocol: rule.protocol || '',
    source_ip: rule.source || '',
    destination_ip: rule.destination || '',
    destination_port: rule.destination_port || ''
  })
}

// 从详细规则页面删除规则
const deleteRuleFromDetail = async (rule: any) => {
  try {
    await ElMessageBox.confirm('确定要删除这条规则吗？', '确认删除', {
      type: 'warning'
    })
    
    if (rule.id) {
      await apiService.deleteRule(rule.id)
      ElMessage.success('删除成功')
      await refreshData() // 刷新所有数据
      // 重新加载详细规则数据
      const chain = chains.value.find((c: any) => c.name === selectedChain.value)
      if (chain) {
        detailRules.value = chain.rules || []
      }
    } else {
      ElMessage.warning('无法删除：规则ID不存在')
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 刷新数据
const refreshData = async () => {
  loading.value = true
  try {
    console.log('开始刷新数据...')
    
    // 自动比对并同步系统规则
    const synced = await autoSyncSystemRules()
    
    // 加载数据
    await Promise.all([
      loadChainTableData(),
      loadInterfaces()
    ])

    console.log('数据刷新完成:', {
      chainTableData: chainTableData.value,
      chains: chainTableData.value?.chains?.length || 0,
      tables: chainTableData.value?.tables?.length || 0,
      interfaces: interfaces.value?.length || 0
    })
    
    // 验证数据结构
    if (chainTableData.value && chainTableData.value.chains) {
      console.log('链数据详情:', chainTableData.value.chains.map(chain => ({
        name: chain.name,
        rulesCount: chain.rules?.length || 0,
        tablesCount: chain.tables?.length || 0
      })))
    } else {
      console.warn('链数据结构异常:', chainTableData.value)
    }
    
    // 根据是否进行了同步显示不同的消息
    if (synced) {
      ElMessage.success('检测到数据不一致，已自动同步并刷新数据')
    } else {
      ElMessage.success('数据刷新成功')
    }
  } catch (error) {
    console.error('刷新数据失败:', error)
    ElMessage.error('数据刷新失败: ' + (error.message || '未知错误'))
  } finally {
    loading.value = false
  }
}

const loadChainTableData = async () => {
  try {
    console.log('开始加载链表数据...')
    const response = await tablesAPI.getAllTables()
    console.log('API响应:', response)
    
    if (response && response.data) {
      // 检查数据结构：API返回的是数组格式还是对象格式
      if (Array.isArray(response.data)) {
        console.log('检测到数组格式数据，进行转换...')
        // API返回的是数组格式：[{table_name: 'raw', chains: [...]}, ...]
        const tableDataArray = response.data
        
        // 转换为目标格式：{chains: [...], tables: [...]}
        const convertedData = {
          chains: [],
          tables: []
        }
        
        // 用于去重链名
        const chainMap = new Map()
        
        // 处理每个表的数据
tableDataArray.forEach((tableItem: any) => {
          if (tableItem && tableItem.table_name && Array.isArray(tableItem.chains)) {
            // 添加表信息
            convertedData.tables.push({
              name: tableItem.table_name,
              total_rules: tableItem.chains.reduce((total: number, chain: any) => total + (chain.rules?.length || 0), 0),
              chains: tableItem.chains.map((chain: any) => {
                return {
                  name: chain.chain_name,
                  policy: chain.policy || 'ACCEPT',
                  rules: chain.rules || []
                };
              })
            })
            
            // 处理链数据，合并相同链名的规则
            tableItem.chains.forEach((chain: any) => {
              if (chain && chain.chain_name) {
                const chainKey = chain.chain_name
                if (!chainMap.has(chainKey)) {
                  chainMap.set(chainKey, {
                    name: chain.chain_name,
                    policy: chain.policy || 'ACCEPT',
                    rules: [],
                    tables: []
                  })
                }
                
                const existingChain = chainMap.get(chainKey)
                // 添加规则
                if (chain.rules && Array.isArray(chain.rules)) {
                  existingChain.rules.push(...chain.rules)
                }
                // 添加表名
                if (!existingChain.tables.includes(tableItem.table_name)) {
                  existingChain.tables.push(tableItem.table_name)
                }
              }
            })
          }
        })
        
// 将Map转换为数组
        convertedData.chains = Array.from(chainMap.values())
        
        chainTableData.value = convertedData
        console.log('数据转换成功:', {
          chains: convertedData.chains.length,
          tables: convertedData.tables.length,
          chainNames: convertedData.chains.map((c: any) => c.name),
          tableNames: convertedData.tables.map((t: any) => t.name)
        })
      } else if (response.data.chains && Array.isArray(response.data.chains)) {
        // 已经是目标格式
        chainTableData.value = response.data
        console.log('链表数据加载成功:', {
          chains: response.data.chains.length,
          tables: response.data.tables?.length || 0
        })
      } else {
        console.warn('API返回数据格式异常，使用模拟数据')
        chainTableData.value = getMockChainTableData()
      }
    } else {
      console.warn('API返回数据为空，使用模拟数据')
      chainTableData.value = getMockChainTableData()
    }
} catch (error: any) {
    console.error('加载链表数据失败:', error)
    console.error('错误详情:', {
      message: error.message,
      status: error.response?.status,
      data: error.response?.data
    })
    
    // 如果API调用失败，使用模拟数据
    chainTableData.value = getMockChainTableData()
    console.log('使用模拟数据:', chainTableData.value)
    
    // 显示用户友好的错误信息
    ElMessage.warning('无法连接到后端服务，正在使用模拟数据')
  }
}

// 模拟数据函数
const getMockChainTableData = () => {
  return {
    chains: [
      {
        name: 'PREROUTING',
        tables: [
          { name: 'raw', rules: [] },
          { name: 'mangle', rules: [{ id: 1, rule_text: 'MARK --set-mark 1' }] },
          { name: 'nat', rules: [{ id: 2, rule_text: 'DNAT --to-destination 192.168.1.100' }] },
          { name: 'filter', rules: [] }
        ],
        rules: [
          { 
            id: 1, 
            chain_name: 'PREROUTING', 
            table: 'mangle', 
            rule_text: 'MARK --set-mark 1',
            line_number: '1',
            target: 'MARK',
            source: '0.0.0.0/0',
            destination: '0.0.0.0/0',
            protocol: 'all',
            in_interface: 'any',
            out_interface: 'any',
            options: '--set-mark 1'
          },
          { 
            id: 2, 
            chain_name: 'PREROUTING', 
            table: 'nat', 
            rule_text: 'DNAT --to-destination 192.168.1.100',
            line_number: '2',
            target: 'DNAT',
            source: '0.0.0.0/0',
            destination: '0.0.0.0/0',
            protocol: 'tcp',
            in_interface: 'eth0',
            out_interface: 'any',
            options: '--to-destination 192.168.1.100'
          }
        ]
      },
      {
        name: 'INPUT',
        tables: [
          { name: 'raw', rules: [] },
          { name: 'mangle', rules: [] },
          { name: 'nat', rules: [] },
          { name: 'filter', rules: [
            { id: 3, rule_text: 'ACCEPT -p tcp --dport 22' },
            { id: 4, rule_text: 'ACCEPT -p tcp --dport 80' },
            { id: 5, rule_text: 'DROP -p tcp --dport 23' }
          ] }
        ],
        rules: [
          { 
            id: 3, 
            chain_name: 'INPUT', 
            table: 'filter', 
            rule_text: 'ACCEPT -p tcp --dport 22',
            line_number: '3',
            target: 'ACCEPT',
            source: '0.0.0.0/0',
            destination: '0.0.0.0/0',
            protocol: 'tcp',
            in_interface: 'any',
            out_interface: 'any',
            options: '--dport 22'
          },
          { 
            id: 4, 
            chain_name: 'INPUT', 
            table: 'filter', 
            rule_text: 'ACCEPT -p tcp --dport 80',
            line_number: '4',
            target: 'ACCEPT',
            source: '0.0.0.0/0',
            destination: '0.0.0.0/0',
            protocol: 'tcp',
            in_interface: 'any',
            out_interface: 'any',
            options: '--dport 80'
          },
          { 
            id: 5, 
            chain_name: 'INPUT', 
            table: 'filter', 
            rule_text: 'DROP -p tcp --dport 23',
            line_number: '5',
            target: 'DROP',
            source: '0.0.0.0/0',
            destination: '0.0.0.0/0',
            protocol: 'tcp',
            in_interface: 'any',
            out_interface: 'any',
            options: '--dport 23'
          }
        ]
      },
      {
        name: 'FORWARD',
        tables: [
          { name: 'raw', rules: [] },
          { name: 'mangle', rules: [] },
          { name: 'nat', rules: [] },
          { name: 'filter', rules: [
            { id: 6, rule_text: 'ACCEPT -i eth0 -o eth1' },
            { id: 7, rule_text: 'DOCKER-ISOLATION-STAGE-1' }
          ] }
        ],
        rules: [
          { id: 6, chain_name: 'FORWARD', table: 'filter', rule_text: 'ACCEPT -i eth0 -o eth1' },
          { id: 7, chain_name: 'FORWARD', table: 'filter', rule_text: 'DOCKER-ISOLATION-STAGE-1' }
        ]
      },
      {
        name: 'OUTPUT',
        tables: [
          { name: 'raw', rules: [] },
          { name: 'mangle', rules: [] },
          { name: 'nat', rules: [] },
          { name: 'filter', rules: [
            { id: 8, rule_text: 'ACCEPT -p tcp --dport 443' }
          ] }
        ],
        rules: [
          { id: 8, chain_name: 'OUTPUT', table: 'filter', rule_text: 'ACCEPT -p tcp --dport 443' }
        ]
      },
      {
        name: 'POSTROUTING',
        tables: [
          { name: 'raw', rules: [] },
          { name: 'mangle', rules: [] },
          { name: 'nat', rules: [
            { id: 9, rule_text: 'MASQUERADE -o eth0' },
            { id: 10, rule_text: 'SNAT --to-source 192.168.1.1' }
          ] },
          { name: 'filter', rules: [] }
        ],
        rules: [
          { id: 9, chain_name: 'POSTROUTING', table: 'nat', rule_text: 'MASQUERADE -o eth0' },
          { id: 10, chain_name: 'POSTROUTING', table: 'nat', rule_text: 'SNAT --to-source 192.168.1.1' }
        ]
      }
    ],
    tables: [
      {
        name: 'raw',
        total_rules: 0,
        chains: []
      },
      {
        name: 'mangle',
        total_rules: 1,
        chains: [
          {
            name: 'PREROUTING',
            policy: 'ACCEPT',
            rules: [{ id: 1, rule_text: 'MARK --set-mark 1' }]
          }
        ]
      },
      {
        name: 'nat',
        total_rules: 3,
        chains: [
          {
            name: 'PREROUTING',
            policy: 'ACCEPT',
            rules: [{ id: 2, rule_text: 'DNAT --to-destination 192.168.1.100' }]
          },
          {
            name: 'POSTROUTING',
            policy: 'ACCEPT',
            rules: [
              { id: 9, rule_text: 'MASQUERADE -o eth0' },
              { id: 10, rule_text: 'SNAT --to-source 192.168.1.1' }
            ]
          }
        ]
      },
      {
        name: 'filter',
        total_rules: 6,
        chains: [
          {
            name: 'INPUT',
            policy: 'ACCEPT',
            rules: [
              { id: 3, rule_text: 'ACCEPT -p tcp --dport 22' },
              { id: 4, rule_text: 'ACCEPT -p tcp --dport 80' },
              { id: 5, rule_text: 'DROP -p tcp --dport 23' }
            ]
          },
          {
            name: 'FORWARD',
            policy: 'ACCEPT',
            rules: [
              { id: 6, rule_text: 'ACCEPT -i eth0 -o eth1' },
              { id: 7, rule_text: 'DOCKER-ISOLATION-STAGE-1' }
            ]
          },
          {
            name: 'OUTPUT',
            policy: 'ACCEPT',
            rules: [
              { id: 8, rule_text: 'ACCEPT -p tcp --dport 443' }
            ]
          }
        ]
      }
    ],
    interface_rules: {
      'eth0': [
        { id: 6, in_interface: 'eth0', rule_text: 'ACCEPT -i eth0 -o eth1' },
        { id: 9, out_interface: 'eth0', rule_text: 'MASQUERADE -o eth0' }
      ],
      'eth1': [
        { id: 6, out_interface: 'eth1', rule_text: 'ACCEPT -i eth0 -o eth1' }
      ]
    }
  }
}

const loadInterfaces = async () => {
  try {
    const response = await networkAPI.getInterfaces()
    interfaces.value = response.data
  } catch (error) {
    console.error('加载网络接口失败:', error)
    // 如果API调用失败，使用模拟数据
    interfaces.value = getMockInterfaces()
  }
}

// 模拟网络接口数据
const getMockInterfaces = () => {
  return [
    {
      name: 'eth0',
      type: 'ethernet',
      state: 'UP',
      ip_addresses: ['192.168.1.100', '10.0.0.1'],
      mac_address: '00:1B:44:11:3A:B7',
      mtu: 1500,
      is_up: true,
      is_docker: false,
      statistics: {
        rx_bytes: 1024000,
        tx_bytes: 512000,
        rx_packets: 1000,
        tx_packets: 500
      }
    },
    {
      name: 'eth1',
      type: 'ethernet',
      state: 'UP',
      ip_addresses: ['172.16.0.1'],
      mac_address: '00:1B:44:11:3A:B8',
      mtu: 1500,
      is_up: true,
      is_docker: false,
      statistics: {
        rx_bytes: 2048000,
        tx_bytes: 1024000,
        rx_packets: 2000,
        tx_packets: 1000
      }
    },
    {
      name: 'docker0',
      type: 'bridge',
      state: 'UP',
      ip_addresses: ['172.17.0.1'],
      mac_address: '02:42:C0:A8:01:01',
      mtu: 1500,
      is_up: true,
      is_docker: true,
      docker_type: 'bridge',
      statistics: {
        rx_bytes: 512000,
        tx_bytes: 256000,
        rx_packets: 500,
        tx_packets: 250
      }
    },
    {
      name: 'lo',
      type: 'loopback',
      state: 'UP',
      ip_addresses: ['127.0.0.1', '::1'],
      mac_address: '',
      mtu: 65536,
      is_up: true,
      is_docker: false,
      statistics: {
        rx_bytes: 1000000,
        tx_bytes: 1000000,
        rx_packets: 10000,
        tx_packets: 10000
      }
    }
  ]
}

// 筛选相关方法
const toggleInterface = (interfaceName: string) => {
  const index = selectedInterfaces.value.indexOf(interfaceName)
  if (index > -1) {
    selectedInterfaces.value.splice(index, 1)
  } else {
    selectedInterfaces.value.push(interfaceName)
  }
}

const toggleProtocol = (protocol: string) => {
  const index = selectedProtocols.value.indexOf(protocol)
  if (index > -1) {
    selectedProtocols.value.splice(index, 1)
  } else {
    selectedProtocols.value.push(protocol)
  }
}

const toggleTarget = (target: string) => {
  const index = selectedTargets.value.indexOf(target)
  if (index > -1) {
    selectedTargets.value.splice(index, 1)
  } else {
    selectedTargets.value.push(target)
  }
}

const clearAllFilters = () => {
  selectedInterfaces.value = []
  selectedProtocols.value = []
  selectedTargets.value = []
  ipRangeFilter.value = ''
  portRangeFilter.value = ''
}

const applyFilters = () => {
  // 筛选条件已通过计算属性自动应用
  ElMessage.success(`已应用 ${activeFiltersCount.value} 个筛选条件`)
}

// 接口视图相关方法
const toggleInterfaceType = (type: string) => {
  const index = selectedInterfaceTypes.value.indexOf(type)
  if (index > -1) {
    selectedInterfaceTypes.value.splice(index, 1)
  } else {
    selectedInterfaceTypes.value.push(type)
  }
}

const toggleInterfaceStatus = (status: string) => {
  if (interfaceStatusFilter.value === status) {
    interfaceStatusFilter.value = ''
  } else {
    interfaceStatusFilter.value = status
  }
}

const clearInterfaceFilters = () => {
  selectedInterfaceTypes.value = []
  interfaceStatusFilter.value = ''
}

const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const viewInterfaceRules = (interfaceName: string) => {
  // 获取该接口相关的所有规则
  const interfaceRules = filteredTableRules.value.filter((rule: any) => 
    rule.InInterface === interfaceName || rule.OutInterface === interfaceName ||
    rule.in_interface === interfaceName || rule.out_interface === interfaceName
  )
  
  // 设置弹窗数据
  selectedChain.value = `接口 ${interfaceName}`
  detailRules.value = interfaceRules
  detailTitle.value = `接口 ${interfaceName} 相关规则${hasActiveFilters.value ? ' (已筛选)' : ''}`
  showChainDialog.value = true
  
  console.log(`显示接口 ${interfaceName} 的规则:`, interfaceRules.length, '条规则')
}

const handleInterfaceChange = () => {
  loadChainTableData()
}

// 标准化连接路径处理
const standardizeConnectionPaths = () => {
  const edges = flowElements.value.filter((el: any) => 'source' in el)
  const nodes = flowElements.value.filter((el: any) => 'position' in el)

  let standardizedCount = 0

  edges.forEach((edge: any) => {
    const sourceNode = nodes.find(n => n.id === edge.source)
    const targetNode = nodes.find(n => n.id === edge.target)

    if (sourceNode && targetNode) {
      // 使用改进的路径计算，支持边缘对齐
      const optimalPath = calculateOptimalPath(sourceNode, targetNode, nodes, edges)

      // 检查是否为关键连接（外部网络→PREROUTING，PREROUTING→路由决策）
      const isKeyConnection = (
          (edge.source === 'interface-external' && edge.target === 'prerouting') ||
          (edge.source === 'prerouting' && edge.target === 'routing-decision')
      )

      if (isKeyConnection) {
        // 关键连接强制使用直线，但保持边缘对齐
        edge.type = 'straight'
        edge.pathOptions = optimalPath.pathOptions // 保留边缘位置信息

        // 确保箭头指向正确，使用边缘对齐
        if (edge.markerEnd) {
          edge.markerEnd.orient = 'auto-start-reverse'
          edge.markerEnd.markerUnits = 'userSpaceOnUse'
          edge.markerEnd.refX = 0 // 箭头尖端对齐到连线终点
          edge.markerEnd.refY = 0
          edge.markerEnd.width = optimalPath.arrowSize
          edge.markerEnd.height = optimalPath.arrowSize
        }

        // 重置样式确保直线显示
        edge.style = {
          ...edge.style,
          strokeLinecap: 'round',
          strokeLinejoin: 'round',
          zIndex: optimalPath.zIndex
        }

        standardizedCount++
      } else {
        // 非关键连接也应用边缘对齐优化
        edge.type = optimalPath.connectionType
        edge.pathOptions = optimalPath.pathOptions
        edge.style = {
          ...edge.style,
          zIndex: optimalPath.zIndex,
          strokeLinecap: 'round',
          strokeLinejoin: 'round'
        }

        if (edge.markerEnd) {
          edge.markerEnd.width = optimalPath.arrowSize
          edge.markerEnd.height = optimalPath.arrowSize
          edge.markerEnd.orient = 'auto-start-reverse'
          edge.markerEnd.markerUnits = 'userSpaceOnUse'
          edge.markerEnd.refX = 0
          edge.markerEnd.refY = 0
        }

        standardizedCount++
      }
    }
  })

  ElMessage.success(`连接路径已标准化，处理了 ${standardizedCount} 个连接，应用边缘对齐显示`)
}

// 计算最优连接路径（增强版）
const calculateOptimalPath = (sourceNode: any, targetNode: any, allNodes: any[], allEdges: any[]) => {
  // 添加安全检查，防止position属性为undefined
  if (!sourceNode?.position || !targetNode?.position) {
    console.warn('Node position is undefined:', { sourceNode, targetNode })
    return {
      needsOptimization: false,
      connectionType: 'straight',
      pathOptions: {},
      arrowSize: 20,
      zIndex: 1000,
      edgePositions: null
    }
  }

  // 计算节点边缘位置
  const sourceEdge = getNodeEdgePosition(sourceNode, targetNode, true)
  const targetEdge = getNodeEdgePosition(targetNode, sourceNode, false)

  const dx = targetEdge.x - sourceEdge.x
  const dy = targetEdge.y - sourceEdge.y
  const distance = Math.sqrt(dx * dx + dy * dy)

  // 优先使用直线连接，确保一致性
  let connectionType = 'straight'
  let pathOptions: any = {}
  let needsAvoidance = false

  // 检查是否为关键连接（外部网络→PREROUTING，PREROUTING→路由决策）
  const isKeyConnection = (
      (sourceNode.id === 'interface-external' && targetNode.id === 'prerouting') ||
      (sourceNode.id === 'prerouting' && targetNode.id === 'routing-decision')
  )

  // 关键连接始终保持直线，不进行避让优化
  if (isKeyConnection) {
    connectionType = 'straight'
    pathOptions = {
      // 使用边缘位置进行连接
      sourceX: sourceEdge.x,
      sourceY: sourceEdge.y,
      targetX: targetEdge.x,
      targetY: targetEdge.y
    }
  } else {
    // 检测是否需要避让其他节点
    needsAvoidance = checkNodeAvoidance(sourceNode, targetNode, allNodes)

    // 只有在必须避让时才使用曲线连接
    if (needsAvoidance) {
      connectionType = 'smoothstep'

      // 水平连接（左右节点）
      if (Math.abs(dy) < 50 && Math.abs(dx) > 100) {
        pathOptions = {
          borderRadius: 8,
          offset: 30,
          centerX: 0.5,
          centerY: 0.5,
          sourceX: sourceEdge.x,
          sourceY: sourceEdge.y,
          targetX: targetEdge.x,
          targetY: targetEdge.y
        }
      }
      // 垂直连接（上下节点）
      else if (Math.abs(dx) < 50 && Math.abs(dy) > 80) {
        pathOptions = {
          borderRadius: 12,
          offset: 35,
          centerX: 0.5,
          centerY: 0.5,
          sourceX: sourceEdge.x,
          sourceY: sourceEdge.y,
          targetX: targetEdge.x,
          targetY: targetEdge.y
        }
      }
      // 对角线连接
      else {
        pathOptions = {
          borderRadius: 15,
          offset: Math.max(30, distance / 6),
          centerX: dx > 0 ? 0.3 : 0.7,
          centerY: dy > 0 ? 0.3 : 0.7,
          sourceX: sourceEdge.x,
          sourceY: sourceEdge.y,
          targetX: targetEdge.x,
          targetY: targetEdge.y
        }
      }
    } else {
      // 直线连接也使用边缘位置
      pathOptions = {
        sourceX: sourceEdge.x,
        sourceY: sourceEdge.y,
        targetX: targetEdge.x,
        targetY: targetEdge.y
      }
    }
  }

// 计算箭头大小 - 增大50%以提升可见性
  let arrowSize = 27 // 默认大小（18 * 1.5）
  if (distance < 150) {
    arrowSize = 21 // 14 * 1.5
  } else if (distance > 300) {
    arrowSize = 33 // 22 * 1.5
  }

  // 计算Z-index层级
  const zIndex = calculateEdgeZIndex(sourceNode, targetNode, allEdges)

  return {
    needsOptimization: needsAvoidance,
    connectionType,
    pathOptions,
    arrowSize,
    zIndex,
    edgePositions: {
      source: sourceEdge,
      target: targetEdge
    }
  }
}

// 检查是否需要避让其他节点
const checkNodeAvoidance = (sourceNode: any, targetNode: any, allNodes: any[]) => {
  // 添加安全检查
  if (!sourceNode?.position || !targetNode?.position) {
    return false
  }

  const path = {
    x1: sourceNode.position.x,
    y1: sourceNode.position.y,
    x2: targetNode.position.x,
    y2: targetNode.position.y
  }

  // 检查路径是否经过其他节点
  return allNodes.some(node => {
    if (node.id === sourceNode.id || node.id === targetNode.id) return false

    // 添加安全检查，防止node.position为undefined
    if (!node?.position) return false

    const nodeCenter = {
      x: node.position.x + 50, // 假设节点宽度100px
      y: node.position.y + 40   // 假设节点高度80px
    }

    // 计算点到线段的距离
    const distance = pointToLineDistance(nodeCenter, path)
    return distance < 60 // 如果距离小于60px，需要避让
  })
}

// 点到线段距离计算
const pointToLineDistance = (point: any, line: any) => {
  const A = point.x - line.x1
  const B = point.y - line.y1
  const C = line.x2 - line.x1
  const D = line.y2 - line.y1

  const dot = A * C + B * D
  const lenSq = C * C + D * D

  if (lenSq === 0) return Math.sqrt(A * A + B * B)

  let param = dot / lenSq
  param = Math.max(0, Math.min(1, param))

  const xx = line.x1 + param * C
  const yy = line.y1 + param * D

  const dx = point.x - xx
  const dy = point.y - yy

  return Math.sqrt(dx * dx + dy * dy)
}

// 计算边的Z-index层级
const calculateEdgeZIndex = (sourceNode: any, targetNode: any, allEdges: any[]) => {
  // 基础层级
  let baseZIndex = 1000

  // 关键路径获得更高层级
  if (sourceNode.data?.chainType === 'forward' || targetNode.data?.chainType === 'forward') {
    baseZIndex += 100
  }

  // 根据连接重要性调整
  if (sourceNode.type === 'interface' || targetNode.type === 'interface') {
    baseZIndex += 50
  }

  return baseZIndex
}


const handleViewModeChange = () => {
  selectedChain.value = ''
}

// 初始化
onMounted(async () => {
  console.log('组件挂载，开始初始化...')
  await refreshData()
  console.log('数据刷新完成')
  
  // 调试数据状态
  debugDataState()
  
  // 确保数据加载完成后再初始化流程图
  nextTick(() => {
    initializeFlowElements()
    
    // 适应视图
    setTimeout(() => {
      const { fitView } = useVueFlow()
      fitView({ padding: 0.2 })
    }, 100)
  })
})

// 数据验证和调试函数
const validateAndDebugData = () => {
  console.log('=== 数据验证和调试 ===')
  console.log('chainTableData.value:', chainTableData.value)
  console.log('chains.value:', chains.value)
  console.log('tables.value:', tables.value)
  console.log('interfaces.value:', interfaces.value)
  console.log('flowElements.value:', flowElements.value)
  
  if (chainTableData.value) {
    console.log('链数据详情:')
    if (chainTableData.value.chains) {
      chainTableData.value.chains.forEach((chain: any, index: number) => {
        console.log(`  链${index + 1}: ${chain.name}`, {
          rules: chain.rules?.length || 0,
          tables: chain.tables?.length || 0
        })
      })
    }
    
    if (chainTableData.value.tables) {
      console.log('表数据详情:')
      chainTableData.value.tables.forEach((table: any, index: number) => {
        console.log(`  表${index + 1}: ${table.name}`, {
          chains: table.chains?.length || 0
        })
      })
    }
  }
  console.log('=== 数据验证结束 ===')
}

// 监听数据变化，更新流程图
watch([chainTableData, rules], () => {
  console.log('数据变化监听器触发')
  validateAndDebugData()
  
  // 确保数据存在后再更新流程图
  if (chainTableData.value && chainTableData.value.chains) {
    console.log('数据验证通过，更新流程图')
    nextTick(() => {
      initializeFlowElements()
    })
  } else {
    console.log('数据验证失败，跳过流程图更新')
  }
}, { deep: true })

// 监听筛选条件变化，更新拓扑图
watch([selectedInterfaces, selectedProtocols, selectedTargets, ipRangeFilter, portRangeFilter], () => {
  console.log('筛选条件变化，更新拓扑图')
  if (chainTableData.value && chainTableData.value.chains) {
    nextTick(() => {
      initializeFlowElements()
    })
  }
}, { deep: true })

// 监听拓扑设置变化，更新拓扑图
watch(topoSettings, () => {
  console.log('拓扑设置变化，更新拓扑图')
  if (viewMode.value === 'chain' && chainTableData.value && chainTableData.value.chains) {
    nextTick(() => {
      // 如果是布局模式变化，重新初始化流程图
      if (topoSettings.layoutMode) {
        initializeFlowElements()
      } else {
        // 否则只应用拓扑设置
        applyTopoSettings()
      }
      
      // 适应视图
      setTimeout(() => {
        const { fitView } = useVueFlow()
        fitView({ padding: 0.2 })
      }, 100)
    })
  }
}, { deep: true })
</script>

<style scoped>
.chain-table-view {
  padding: 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  min-height: 100vh;
}

.page-header {
  text-align: center;
  margin-bottom: 30px;
  color: white;
}

.page-header h1 {
  font-size: 2.5rem;
  margin-bottom: 10px;
  text-shadow: 2px 2px 4px rgba(0,0,0,0.3);
}

.description {
  font-size: 1.1rem;
  opacity: 0.9;
}

.control-panel {
  margin-bottom: 20px;
}

.controls {
  display: flex;
  gap: 15px;
  align-items: center;
  flex-wrap: wrap;
}

.main-content {
  background: white;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0,0,0,0.1);
  overflow: hidden;
}

.dataflow-view {
  height: 800px;
  position: relative;
}

.vue-flow-wrapper {
  height: 100%;
  width: 100%;
}

.dataflow-diagram {
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
}

/* 自定义节点样式 */
.chain-node {
  border-radius: 8px;
  padding: 12px;
  min-width: 140px;
  transition: all 0.3s ease;
  cursor: pointer;
  box-shadow: 0 4px 8px rgba(0,0,0,0.1);
}

/* 扁平风格 */
.chain-node.flat {
  background: white;
  border: 1px solid #e1e5e9;
}

/* 渐变风格 */
.chain-node.gradient {
  background: linear-gradient(135deg, #ffffff 0%, #f5f7fa 100%);
  border: 1px solid #e1e5e9;
}

/* 玻璃风格 */
.chain-node.glass {
  background: rgba(255, 255, 255, 0.7);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.3);
}

.chain-node:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 12px rgba(0,0,0,0.15);
}

/* 链节点颜色 - 扁平风格 */
.chain-node.flat.prerouting {
  border: 2px solid #ff9800;
  background: #fff8e1;
}

.chain-node.flat.input {
  border: 2px solid #8bc34a;
  background: #f1f8e9;
}

.chain-node.flat.forward {
  border: 2px solid #ff5722;
  background: #fbe9e7;
}

.chain-node.flat.output {
  border: 2px solid #03a9f4;
  background: #e1f5fe;
}

.chain-node.flat.postrouting {
  border: 2px solid #9c27b0;
  background: #f3e5f5;
}

/* 链节点颜色 - 渐变风格 */
.chain-node.gradient.prerouting {
  border: 2px solid #ff9800;
  background: linear-gradient(135deg, #fff8e1 0%, #ffe0b2 100%);
}

.chain-node.gradient.input {
  border: 2px solid #8bc34a;
  background: linear-gradient(135deg, #f1f8e9 0%, #dcedc8 100%);
}

.chain-node.gradient.forward {
  border: 2px solid #ff5722;
  background: linear-gradient(135deg, #fbe9e7 0%, #ffccbc 100%);
}

.chain-node.gradient.output {
  border: 2px solid #03a9f4;
  background: linear-gradient(135deg, #e1f5fe 0%, #b3e5fc 100%);
}

.chain-node.gradient.postrouting {
  border: 2px solid #9c27b0;
  background: linear-gradient(135deg, #f3e5f5 0%, #e1bee7 100%);
}

/* 链节点颜色 - 玻璃风格 */
.chain-node.glass.prerouting {
  border: 2px solid rgba(255, 152, 0, 0.5);
  background: rgba(255, 248, 225, 0.7);
}

.chain-node.glass.input {
  border: 2px solid rgba(139, 195, 74, 0.5);
  background: rgba(241, 248, 233, 0.7);
}

.chain-node.glass.forward {
  border: 2px solid rgba(255, 87, 34, 0.5);
  background: rgba(251, 233, 231, 0.7);
}

.chain-node.glass.output {
  border: 2px solid rgba(3, 169, 244, 0.5);
  background: rgba(225, 245, 254, 0.7);
}

.chain-node.glass.postrouting {
  border: 2px solid rgba(156, 39, 176, 0.5);
  background: rgba(243, 229, 245, 0.7);
}

.chain-title {
  font-size: 1.2rem;
  font-weight: bold;
  margin-bottom: 10px;
  text-align: center;
  color: #2d3748;
}

.chain-tables {
  display: flex;
  gap: 5px;
  justify-content: center;
  margin-bottom: 10px;
  flex-wrap: wrap;
}

.table-tag {
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  display: inline-block;
  margin: 2px;
}

.table-tag:hover {
  transform: scale(1.05);
}

.table-tag.raw {
  background: #ff5722;
  color: white;
}

.table-tag.mangle {
  background: #ff9800;
  color: white;
}

.table-tag.nat {
  background: #03a9f4;
  color: white;
}

.table-tag.filter {
  background: #8bc34a;
  color: white;
}

.chain-stats {
  text-align: center;
  font-size: 0.9rem;
  color: #666;
  font-weight: 500;
}

.decision-node {
  background: #ff9800;
  border: none;
  border-radius: 8px;
  width: 100px;
  height: 100px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 8px rgba(0,0,0,0.15);
  transition: all 0.3s ease;
}

.decision-node:hover {
  transform: scale(1.05);
  box-shadow: 0 6px 12px rgba(0,0,0,0.2);
}

.decision-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
}

.decision-icon {
  font-size: 20px;
  margin-bottom: 6px;
  color: #ffffff;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 50%;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.router-icon {
  font-size: 24px;
  color: white;
}

.decision-label {
  font-weight: bold;
  color: #ffffff;
  font-size: 0.85rem;
  line-height: 1.2;
  max-width: 90px;
}

.endpoint-node {
  background: #f5f7fa;
  color: #333;
  border-radius: 8px;
  padding: 10px 15px;
  font-weight: bold;
  box-shadow: 0 4px 8px rgba(0,0,0,0.1);
  transition: all 0.3s ease;
  min-width: 120px;
  border: 1px solid #e1e5e9;
}

.endpoint-node:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 12px rgba(0,0,0,0.15);
}

.endpoint-node.entry {
  background: #f1f8e9;
  border: 1px solid #8bc34a;
}

.endpoint-node.exit {
  background: #f1f8e9;
  border: 1px solid #8bc34a;
}

.endpoint-node.protocol {
  background: #f5f7fa;
  border: 1px solid #e1e5e9;
}

.endpoint-content {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  flex-direction: column;
}

.endpoint-icon {
  background: rgba(0, 0, 0, 0.05);
  border-radius: 50%;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.server-icon {
  font-size: 24px;
  color: #4A90E2;
}

.endpoint-label {
  font-weight: bold;
  font-size: 0.85rem;
  text-align: center;
}

.process-node {
  background: #9c27b0;
  color: white;
  border-radius: 8px;
  padding: 12px 15px;
  font-weight: bold;
  box-shadow: 0 4px 8px rgba(0,0,0,0.15);
  transition: all 0.3s ease;
  min-width: 100px;
}

.process-node:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 12px rgba(0,0,0,0.2);
}

.process-content {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.process-icon {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 50%;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.gear-icon {
  font-size: 24px;
  color: white;
}

.process-label {
  font-weight: bold;
  font-size: 0.85rem;
}

.protocol-node {
  background: linear-gradient(135deg, #a8edea 0%, #fed6e3 100%);
  border: 2px solid #48bb78;
  border-radius: 12px;
  padding: 15px 25px;
  font-weight: bold;
  color: #2d3748;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
  transition: all 0.3s ease;
  min-width: 160px;
}

.protocol-node:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0,0,0,0.15);
}

.protocol-content {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
}

.protocol-icon {
  background: rgba(255, 255, 255, 0.5);
  border-radius: 50%;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #2d3748;
}

.protocol-label {
  font-weight: bold;
  font-size: 0.9rem;
}

.loading-container {
  padding: 40px;
  background: white;
  border-radius: 12px;
  margin-top: 20px;
}

/* Vue Flow 自定义样式 */
:deep(.vue-flow__edge-path) {
  stroke-width: 2px;
}

:deep(.vue-flow__edge-label) {
  background: rgba(255, 255, 255, 0.9);
  border-radius: 6px;
  padding: 4px 8px;
  font-size: 0.8rem;
  font-weight: 500;
  color: #2d3748;
}

:deep(.vue-flow__edge.animated .vue-flow__edge-path) {
  stroke-dasharray: 5, 5;
  animation: flowEdgeAnimation 1s linear infinite;
}

@keyframes flowEdgeAnimation {
  from {
    stroke-dashoffset: 10;
  }
  to {
    stroke-dashoffset: 0;
  }
}

:deep(.vue-flow__controls) {
  background: rgba(255, 255, 255, 0.9);
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
}

:deep(.vue-flow__controls-button) {
  border: none;
  background: transparent;
  color: #4a5568;
  transition: all 0.2s ease;
}

:deep(.vue-flow__controls-button:hover) {
  background: #e2e8f0;
  color: #2d3748;
}

/* 控制面板样式 */
.main-controls {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.view-tabs {
  flex: 1;
}

.action-buttons {
  display: flex;
  gap: 12px;
}

/* 筛选面板样式 */
.filter-panel {
  margin-top: 16px;
}

.filter-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
}

.filter-content {
  padding: 16px 0;
}

.quick-filters {
  margin-bottom: 16px;
}

.filter-group {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
  gap: 12px;
}

.filter-label {
  min-width: 80px;
  font-weight: 500;
  color: #606266;
}

.filter-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.filter-tag {
  cursor: pointer;
  transition: all 0.3s ease;
}

.filter-tag:hover {
  transform: scale(1.05);
}

.advanced-filters {
  border-top: 1px solid #e4e7ed;
  padding-top: 16px;
}

.filter-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

/* 卡片式表视图样式 */
.rules-cards-container {
  padding: 20px;
}

.rules-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 20px;
  margin-bottom: 20px;
}

.rule-card {
  transition: all 0.3s ease;
}

.rule-card:hover {
  transform: translateY(-2px);
}

.rule-card-content {
  height: 100%;
}

.rule-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.rule-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.rule-number {
  display: flex;
  align-items: center;
  gap: 4px;
  font-weight: bold;
  color: #409eff;
}

.rule-stats {
  display: flex;
  gap: 6px;
}

.rule-actions {
  display: flex;
  gap: 6px;
}

.rule-card-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.rule-chain-table {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px;
  background: #f8f9fa;
  border-radius: 6px;
}

.rule-target {
  display: flex;
  align-items: center;
  gap: 8px;
}

.rule-target label {
  font-weight: 500;
  color: #606266;
  min-width: 40px;
}

.rule-network {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 8px;
  background: #f0f9ff;
  border-radius: 6px;
  border-left: 3px solid #409eff;
}

.network-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.network-item label {
  font-weight: 500;
  color: #606266;
  min-width: 60px;
  font-size: 12px;
}

.network-value {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
  color: #2d3748;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rule-interfaces {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px;
  background: #f0fdf4;
  border-radius: 6px;
  border-left: 3px solid #10b981;
}

.interface-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.interface-arrow {
  color: #9ca3af;
}

.rule-options {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 8px;
  background: #fffbeb;
  border-radius: 6px;
  border-left: 3px solid #f59e0b;
}

.rule-options label {
  font-weight: 500;
  color: #606266;
  min-width: 40px;
  font-size: 12px;
}

.options-text {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 11px;
  color: #2d3748;
  line-height: 1.4;
  word-break: break-all;
}

.empty-state {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 300px;
}

/* 接口视图新增样式 */
.stats-panel {
  margin-bottom: 20px;
}

.stats-card {
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.stats-content {
  display: flex;
  align-items: center;
  padding: 16px;
}

.stats-icon {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 16px;
  background: #e3f2fd;
  color: #1976d2;
}

.stats-icon.active {
  background: #e8f5e8;
  color: #4caf50;
}

.stats-icon.docker {
  background: #fff3e0;
  color: #ff9800;
}

.stats-icon.rules {
  background: #f3e5f5;
  color: #9c27b0;
}

.stats-info {
  flex: 1;
}

.stats-number {
  font-size: 24px;
  font-weight: bold;
  color: #333;
  line-height: 1;
}

.stats-label {
  font-size: 14px;
  color: #666;
  margin-top: 4px;
}

.interface-filters {
  margin-bottom: 20px;
}

.interface-filters .filter-content {
  padding: 16px;
}

.interface-filters .filter-group {
  margin-bottom: 16px;
}

.interface-filters .filter-group:last-child {
  margin-bottom: 0;
}

.interface-filters .filter-actions {
  text-align: right;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #eee;
}

.interfaces-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(450px, 1fr));
  gap: 20px;
}

.interface-card-content {
  height: 100%;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transition: all 0.3s ease;
}

.interface-card-content:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}

.interface-card-content.docker-interface {
  border-left: 4px solid #ff9800;
}

.interface-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.interface-icon {
  font-size: 20px;
  color: #1976d2;
}

.interface-badges {
  display: flex;
  gap: 8px;
}

.interface-basic-info {
  margin-bottom: 20px;
  padding: 16px;
  background: #f8f9fa;
  border-radius: 6px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
}

.info-row:last-child {
  margin-bottom: 0;
}

.info-label {
  font-weight: 500;
  color: #666;
}

.info-value {
  color: #333;
}

.interface-network-info {
  margin-bottom: 20px;
}

.network-section {
  margin-bottom: 16px;
}

.network-section h4 {
  margin: 0 0 8px 0;
  color: #333;
  font-size: 14px;
  font-weight: 600;
}

.address-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.address-tag {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
}

.no-address {
  color: #999;
  font-style: italic;
}

.mac-address {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  background: #f5f5f5;
  padding: 4px 8px;
  border-radius: 4px;
  color: #666;
}

.interface-rules-stats {
  margin-bottom: 20px;
}

.stats-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.stats-header h4 {
  margin: 0;
  color: #333;
  font-size: 14px;
  font-weight: 600;
}

.rules-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.rule-stat-item {
  display: flex;
  align-items: center;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 6px;
  border: 1px solid #e9ecef;
}

.rule-stat-icon {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 12px;
  font-size: 16px;
}

.rule-stat-icon.input {
  background: #e3f2fd;
  color: #1976d2;
}

.rule-stat-icon.output {
  background: #fff3e0;
  color: #f57c00;
}

.rule-stat-icon.forward {
  background: #e8f5e8;
  color: #388e3c;
}

.rule-stat-icon.total {
  background: #f3e5f5;
  color: #7b1fa2;
}

.rule-stat-item.total {
  grid-column: span 2;
  background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
  border: 2px solid #dee2e6;
}

.rule-stat-info {
  flex: 1;
}

.rule-stat-number {
  font-size: 18px;
  font-weight: bold;
  color: #333;
  line-height: 1;
}

.rule-stat-label {
  font-size: 12px;
  color: #666;
  margin-top: 2px;
}

.interface-traffic-stats {
  margin-bottom: 20px;
}

.interface-traffic-stats h4 {
  margin: 0 0 12px 0;
  color: #333;
  font-size: 14px;
  font-weight: 600;
}

.traffic-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.traffic-item {
  display: flex;
  justify-content: space-between;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 6px;
  border: 1px solid #e9ecef;
}

.traffic-label {
  font-weight: 500;
  color: #666;
}

.traffic-value {
  color: #333;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
}

.interface-actions {
  text-align: center;
  padding-top: 16px;
  border-top: 1px solid #eee;
}

/* 拓扑图节点样式 - 新增 */
/* 扁平风格 */
.chain-node.flat {
  background: white;
  border: 2px solid #e1e5e9;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
}

/* 渐变风格 */
.chain-node.gradient {
  background: linear-gradient(135deg, #ffffff 0%, #f5f7fa 100%);
  border: 2px solid #e1e5e9;
  box-shadow: 0 6px 16px rgba(0,0,0,0.15);
}

/* 玻璃风格 */
.chain-node.glass {
  background: rgba(255, 255, 255, 0.7);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.3);
  box-shadow: 0 8px 32px rgba(31, 38, 135, 0.15);
}

/* 链节点颜色 - 扁平风格 */
.chain-node.flat.prerouting {
  border-color: #ff6b6b;
  background: #fff5f5;
}

.chain-node.flat.input {
  border-color: #4ecdc4;
  background: #f0fdfc;
}

.chain-node.flat.forward {
  border-color: #45b7d1;
  background: #f0f9ff;
}

.chain-node.flat.output {
  border-color: #96ceb4;
  background: #f0fdf4;
}

.chain-node.flat.postrouting {
  border-color: #feca57;
  background: #fffbeb;
}

/* 链节点颜色 - 渐变风格 */
.chain-node.gradient.prerouting {
  border-color: #ff6b6b;
  background: linear-gradient(135deg, #fff5f5 0%, #ffe0e0 100%);
}

.chain-node.gradient.input {
  border-color: #4ecdc4;
  background: linear-gradient(135deg, #f0fdfc 0%, #ccfbf1 100%);
}

.chain-node.gradient.forward {
  border-color: #45b7d1;
  background: linear-gradient(135deg, #f0f9ff 0%, #dbeafe 100%);
}

.chain-node.gradient.output {
  border-color: #96ceb4;
  background: linear-gradient(135deg, #f0fdf4 0%, #dcfce7 100%);
}

.chain-node.gradient.postrouting {
  border-color: #feca57;
  background: linear-gradient(135deg, #fffbeb 0%, #fef3c7 100%);
}

/* 链节点颜色 - 玻璃风格 */
.chain-node.glass.prerouting {
  border-color: rgba(255, 107, 107, 0.3);
  background: rgba(255, 245, 245, 0.7);
}

.chain-node.glass.input {
  border-color: rgba(78, 205, 196, 0.3);
  background: rgba(240, 253, 252, 0.7);
}

.chain-node.glass.forward {
  border-color: rgba(69, 183, 209, 0.3);
  background: rgba(240, 249, 255, 0.7);
}

.chain-node.glass.output {
  border-color: rgba(150, 206, 180, 0.3);
  background: rgba(240, 253, 244, 0.7);
}

.chain-node.glass.postrouting {
  border-color: rgba(254, 202, 87, 0.3);
  background: rgba(255, 251, 235, 0.7);
}

/* 决策节点样式 */
.decision-node.flat {
  background: #feca57;
  border: 2px solid #f39c12;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
}

.decision-node.gradient {
  background: linear-gradient(135deg, #f6d365 0%, #fda085 100%);
  border: 2px solid #f39c12;
  box-shadow: 0 6px 16px rgba(0,0,0,0.15);
}

.decision-node.glass {
  background: rgba(254, 202, 87, 0.7);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(243, 156, 18, 0.3);
  box-shadow: 0 8px 32px rgba(31, 38, 135, 0.15);
}

/* 端点节点样式 */
.endpoint-node.flat {
  background: white;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
  border: 2px solid #e1e5e9;
}

.endpoint-node.gradient {
  background: linear-gradient(135deg, #ffffff 0%, #f5f7fa 100%);
  box-shadow: 0 6px 16px rgba(0,0,0,0.15);
  border: 2px solid #e1e5e9;
}

.endpoint-node.glass {
  background: rgba(255, 255, 255, 0.7);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.3);
  box-shadow: 0 8px 32px rgba(31, 38, 135, 0.15);
}

/* 进程节点样式 */
.process-node.flat {
  background: #e0e7ff;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
  border: 2px solid #a5b4fc;
}

.process-node.gradient {
  background: linear-gradient(135deg, #e0e7ff 0%, #c7d2fe 100%);
  box-shadow: 0 6px 16px rgba(0,0,0,0.15);
  border: 2px solid #a5b4fc;
}

.process-node.glass {
  background: rgba(224, 231, 255, 0.7);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(165, 180, 252, 0.3);
  box-shadow: 0 8px 32px rgba(31, 38, 135, 0.15);
}

/* 协议节点样式 */
.protocol-node.flat {
  background: #dbeafe;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
  border: 2px solid #93c5fd;
}

.protocol-node.gradient {
  background: linear-gradient(135deg, #dbeafe 0%, #bfdbfe 100%);
  box-shadow: 0 6px 16px rgba(0,0,0,0.15);
  border: 2px solid #93c5fd;
}

.protocol-node.glass {
  background: rgba(219, 234, 254, 0.7);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(147, 197, 253, 0.3);
  box-shadow: 0 8px 32px rgba(31, 38, 135, 0.15);
}

/* 暗色模式支持 */
.dataflow-diagram.dark {
  background: linear-gradient(135deg, #1a202c 0%, #2d3748 100%);
}

.dataflow-diagram.dark .vue-flow__edge-path {
  stroke: rgba(255, 255, 255, 0.7);
}

.dataflow-diagram.dark .vue-flow__edge-text {
  fill: white;
}

/* 响应式设计 */
@media (max-width: 1200px) {
  .rules-grid {
    grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  }
}

@media (max-width: 768px) {
  .chain-table-view {
    padding: 10px;
  }
  
  .page-header h1 {
    font-size: 2rem;
  }
  
  .main-controls {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }
  
  .filter-group {
    flex-direction: column;
    align-items: flex-start;
  }
  
  .filter-label {
    min-width: auto;
  }
  
  .rules-grid {
    grid-template-columns: 1fr;
  }
  
  .rule-card-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
  
  .rule-network {
    font-size: 12px;
  }
  
  .network-value {
    max-width: 150px;
  }
  
  .dataflow-view {
    height: 600px;
  }
  
  .chain-node {
    min-width: 120px;
    padding: 10px;
  }
  
  .decision-node {
    width: 80px;
    height: 80px;
  }
}
</style>