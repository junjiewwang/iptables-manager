<template>
  <div class="topology-container">
    <div class="topology-header">
      <h2>网络拓扑图</h2>
      <div class="topology-controls">
        <el-button-group>
          <el-button 
            v-for="flow in topologyData?.flow || []" 
            :key="flow.id"
            :type="selectedFlow === flow.id ? 'primary' : 'default'"
            size="small"
            @click="highlightFlow(flow.id)"
          >
            {{ flow.name }}
          </el-button>
        </el-button-group>
        
        <el-divider direction="vertical" />
        
        <!-- 新增：iptables链路视图切换 -->
        <el-button 
          :type="viewMode === 'iptables' ? 'primary' : 'default'"
          size="small"
          @click="toggleViewMode"
        >
          IPTables链路
        </el-button>
        
        <el-divider direction="vertical" />
        
        <!-- 过滤控件 -->
        <el-select 
          v-model="protocolFilter" 
          placeholder="协议过滤" 
          size="small" 
          style="width: 120px"
          clearable
          @change="applyFilters"
        >
          <el-option label="全部" value="" />
          <el-option label="TCP" value="tcp" />
          <el-option label="UDP" value="udp" />
          <el-option label="ICMP" value="icmp" />
          <el-option label="HTTP" value="http" />
          <el-option label="HTTPS" value="https" />
        </el-select>
        
        <el-input 
          v-model="portFilter" 
          placeholder="端口过滤" 
          size="small" 
          style="width: 100px"
          clearable
          @input="applyFilters"
        />
        
        <el-select 
          v-model="chainFilter" 
          placeholder="链类型" 
          size="small" 
          style="width: 120px"
          clearable
          @change="applyFilters"
        >
          <el-option label="全部" value="" />
          <el-option label="INPUT" value="INPUT" />
          <el-option label="OUTPUT" value="OUTPUT" />
          <el-option label="FORWARD" value="FORWARD" />
          <el-option label="PREROUTING" value="PREROUTING" />
          <el-option label="POSTROUTING" value="POSTROUTING" />
        </el-select>
        
        <el-divider direction="vertical" />
        
        <el-button @click="resetView" size="small">重置视图</el-button>
        <el-button @click="refreshTopology" size="small" :loading="loading">刷新</el-button>
        <el-button @click="exportTopology" size="small">导出</el-button>
        <el-button @click="toggleAutoRefresh" size="small" :type="autoRefresh ? 'success' : 'default'">
          {{ autoRefresh ? '停止自动刷新' : '自动刷新' }}
        </el-button>
      </div>
    </div>

    <div class="topology-content">
      <div class="topology-sidebar">
        <!-- IPTables链路图例 -->
        <el-card class="legend-card" v-if="viewMode === 'iptables'">
          <template #header>
            <span>IPTables链路图例</span>
          </template>
          <div class="legend-items">
            <div class="legend-section">
              <h4>数据包流向</h4>
              <div class="legend-item">
                <div class="legend-line packet-flow"></div>
                <span>数据包流</span>
              </div>
              <div class="legend-item">
                <div class="legend-line nat-flow"></div>
                <span>NAT流</span>
              </div>
              <div class="legend-item">
                <div class="legend-line filter-flow"></div>
                <span>过滤流</span>
              </div>
            </div>
            
            <div class="legend-section">
              <h4>处理阶段</h4>
              <div class="legend-item">
                <div class="legend-icon prerouting-icon"></div>
                <span>PREROUTING</span>
              </div>
              <div class="legend-item">
                <div class="legend-icon input-icon"></div>
                <span>INPUT</span>
              </div>
              <div class="legend-item">
                <div class="legend-icon forward-icon"></div>
                <span>FORWARD</span>
              </div>
              <div class="legend-item">
                <div class="legend-icon output-icon"></div>
                <span>OUTPUT</span>
              </div>
              <div class="legend-item">
                <div class="legend-icon postrouting-icon"></div>
                <span>POSTROUTING</span>
              </div>
            </div>
            
            <div class="legend-section">
              <h4>表类型</h4>
              <div class="legend-item">
                <div class="legend-icon raw-table"></div>
                <span>raw表</span>
              </div>
              <div class="legend-item">
                <div class="legend-icon mangle-table"></div>
                <span>mangle表</span>
              </div>
              <div class="legend-item">
                <div class="legend-icon nat-table"></div>
                <span>nat表</span>
              </div>
              <div class="legend-item">
                <div class="legend-icon filter-table"></div>
                <span>filter表</span>
              </div>
            </div>
          </div>
        </el-card>

        <!-- 原始图例 -->
        <el-card class="legend-card" v-else>
          <template #header>
            <span>图例</span>
          </template>
          <div class="legend-items">
            <!-- 节点类型图例 -->
            <div class="legend-section">
              <h4>节点类型</h4>
              <div class="legend-item">
                <div class="legend-icon interface-external-icon"></div>
                <span>外部网络接口</span>
              </div>
              <div class="legend-item">
                <div class="legend-icon interface-internal-icon"></div>
                <span>内部网络接口</span>
              </div>
              <div class="legend-item">
                <div class="legend-icon interface-docker-icon"></div>
                <span>Docker网络接口</span>
              </div>
              <div class="legend-item">
                <div class="legend-icon rule-icon"></div>
                <span>IPTables规则</span>
              </div>
            </div>
            
            <!-- 连接类型图例 -->
            <div class="legend-section">
              <h4>连接类型</h4>
              <div class="legend-item">
                <div class="legend-line link-input"></div>
                <span>入站连接</span>
              </div>
              <div class="legend-item">
                <div class="legend-line link-output"></div>
                <span>出站连接</span>
              </div>
              <div class="legend-item">
                <div class="legend-line link-forward"></div>
                <span>转发连接</span>
              </div>
            </div>
            
            <!-- 动作类型图例 -->
            <div class="legend-section">
              <h4>动作类型</h4>
              <div class="legend-item">
                <el-tag type="success" size="small">ACCEPT</el-tag>
                <span>允许</span>
              </div>
              <div class="legend-item">
                <el-tag type="danger" size="small">DROP</el-tag>
                <span>丢弃</span>
              </div>
              <div class="legend-item">
                <el-tag type="warning" size="small">REJECT</el-tag>
                <span>拒绝</span>
              </div>
            </div>
            
            <!-- 处理阶段图例 -->
            <div class="legend-section">
              <h4>处理阶段</h4>
              <div class="legend-item">
                <div class="legend-icon stage-prerouting-icon"></div>
                <span>PREROUTING</span>
              </div>
              <div class="legend-item">
                <div class="legend-icon stage-input-icon"></div>
                <span>INPUT</span>
              </div>
              <div class="legend-item">
                <div class="legend-icon stage-forward-icon"></div>
                <span>FORWARD</span>
              </div>
              <div class="legend-item">
                <div class="legend-icon stage-output-icon"></div>
                <span>OUTPUT</span>
              </div>
              <div class="legend-item">
                <div class="legend-icon stage-postrouting-icon"></div>
                <span>POSTROUTING</span>
              </div>
            </div>
          </div>
        </el-card>

        <!-- 统计信息卡片 -->
        <el-card class="stats-card" v-if="topologyStats">
          <template #header>
            <span>统计信息</span>
          </template>
          <div class="stats-content">
            <div class="stat-item">
              <span class="stat-label">总节点数:</span>
              <span class="stat-value">{{ topologyStats.total_nodes }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">总连接数:</span>
              <span class="stat-value">{{ topologyStats.total_links }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">数据流数:</span>
              <span class="stat-value">{{ topologyStats.total_flows }}</span>
            </div>
            <div class="stat-item" v-for="(count, type) in topologyStats.node_types" :key="type">
              <span class="stat-label">{{ type }}:</span>
              <span class="stat-value">{{ count }}</span>
            </div>
          </div>
        </el-card>

        <!-- 悬停信息卡片 -->
        <el-card class="hover-info-card" v-if="hoveredNode || hoveredLink">
          <template #header>
            <span>{{ hoveredNode ? '节点信息' : '连接信息' }}</span>
          </template>
          <div class="hover-info-content">
            <!-- 节点悬停信息 -->
            <div v-if="hoveredNode">
              <el-descriptions :column="1" size="small">
                <el-descriptions-item label="ID">{{ hoveredNode.id }}</el-descriptions-item>
                <el-descriptions-item label="类型">{{ hoveredNode.type }}</el-descriptions-item>
                <el-descriptions-item label="标签">{{ hoveredNode.label }}</el-descriptions-item>
                <el-descriptions-item v-if="hoveredNode.interface_name" label="接口">
                  {{ hoveredNode.interface_name }}
                </el-descriptions-item>
                <el-descriptions-item v-if="hoveredNode.chain_name" label="链">
                  {{ hoveredNode.chain_name }}
                </el-descriptions-item>
                <el-descriptions-item v-if="hoveredNode.packets" label="数据包">
                  {{ hoveredNode.packets }}
                </el-descriptions-item>
                <el-descriptions-item v-if="hoveredNode.bytes" label="字节">
                  {{ hoveredNode.bytes }}
                </el-descriptions-item>
              </el-descriptions>
            </div>
            
            <!-- 连接悬停信息 -->
            <div v-if="hoveredLink">
              <el-descriptions :column="1" size="small">
                <el-descriptions-item label="ID">{{ hoveredLink.id }}</el-descriptions-item>
                <el-descriptions-item label="类型">{{ hoveredLink.type }}</el-descriptions-item>
                <el-descriptions-item label="源节点">{{ hoveredLink.source }}</el-descriptions-item>
                <el-descriptions-item label="目标节点">{{ hoveredLink.target }}</el-descriptions-item>
                <el-descriptions-item v-if="hoveredLink.chain_type" label="链类型">
                  {{ hoveredLink.chain_type }}
                </el-descriptions-item>
                <el-descriptions-item v-if="hoveredLink.action" label="动作">
                  <el-tag :type="hoveredLink.action === 'ACCEPT' ? 'success' : 'danger'" size="small">
                    {{ hoveredLink.action }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item v-if="hoveredLink.protocol" label="协议">
                  {{ hoveredLink.protocol }}
                </el-descriptions-item>
                <el-descriptions-item v-if="hoveredLink.port" label="端口">
                  {{ hoveredLink.port }}
                </el-descriptions-item>
              </el-descriptions>
              <div v-if="hoveredLink.rule_text" class="rule-text">
                <h5>完整规则:</h5>
                <code>{{ hoveredLink.rule_text }}</code>
              </div>
            </div>
          </div>
        </el-card>
      </div>

      <div class="topology-main">
        <div 
          ref="topologyCanvas" 
          class="topology-canvas"
          v-loading="loading"
          element-loading-text="加载拓扑图数据..."
        >
          <!-- IPTables链路视图 -->
          <svg 
            v-if="viewMode === 'iptables'"
            ref="svgElement" 
            class="topology-svg iptables-topology"
            @click="onCanvasClick"
          >
            <!-- 定义箭头标记和渐变 -->
            <defs>
              <!-- 数据包流箭头 -->
              <marker id="packet-arrow" markerWidth="12" markerHeight="8" refX="11" refY="4" orient="auto">
                <polygon points="0 0, 12 4, 0 8" fill="#409EFF" />
              </marker>
              
              <!-- NAT流箭头 -->
              <marker id="nat-arrow" markerWidth="12" markerHeight="8" refX="11" refY="4" orient="auto">
                <polygon points="0 0, 12 4, 0 8" fill="#FF5722" />
              </marker>
              
              <!-- 过滤流箭头 -->
              <marker id="filter-arrow" markerWidth="12" markerHeight="8" refX="11" refY="4" orient="auto">
                <polygon points="0 0, 12 4, 0 8" fill="#4CAF50" />
              </marker>
              
              <!-- 处理阶段渐变 -->
              <linearGradient id="preroutingGradient" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" style="stop-color:#FF9800;stop-opacity:1" />
                <stop offset="100%" style="stop-color:#F57C00;stop-opacity:1" />
              </linearGradient>
              
              <linearGradient id="inputGradient" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" style="stop-color:#4CAF50;stop-opacity:1" />
                <stop offset="100%" style="stop-color:#388E3C;stop-opacity:1" />
              </linearGradient>
              
              <linearGradient id="forwardGradient" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" style="stop-color:#FF5722;stop-opacity:1" />
                <stop offset="100%" style="stop-color:#D84315;stop-opacity:1" />
              </linearGradient>
              
              <linearGradient id="outputGradient" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" style="stop-color:#2196F3;stop-opacity:1" />
                <stop offset="100%" style="stop-color:#1976D2;stop-opacity:1" />
              </linearGradient>
              
              <linearGradient id="postroutingGradient" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" style="stop-color:#9C27B0;stop-opacity:1" />
                <stop offset="100%" style="stop-color:#7B1FA2;stop-opacity:1" />
              </linearGradient>
            </defs>

            <!-- 绘制IPTables链路流程 -->
            <g class="iptables-flow">
              <!-- 数据包入口 -->
              <g class="packet-entry" transform="translate(50, 200)">
                <circle r="25" fill="#607D8B" stroke="#455A64" stroke-width="2"/>
                <text y="5" text-anchor="middle" dominant-baseline="middle" fill="white" font-size="12" font-weight="bold">
                  数据包入口
                </text>
              </g>

              <!-- PREROUTING阶段 -->
              <g class="prerouting-stage" transform="translate(200, 200)">
                <rect x="-40" y="-20" width="80" height="40" rx="5" fill="url(#preroutingGradient)" stroke="#F57C00" stroke-width="2"/>
                <text y="5" text-anchor="middle" dominant-baseline="middle" fill="white" font-size="11" font-weight="bold">
                  PREROUTING
                </text>
                <text y="20" text-anchor="middle" dominant-baseline="middle" fill="#666" font-size="9">
                  raw → mangle → nat
                </text>
              </g>

              <!-- 路由判断 -->
              <g class="routing-decision" transform="translate(350, 200)">
                <polygon points="0,-20 25,0 0,20 -25,0" fill="#FFC107" stroke="#F57C00" stroke-width="2"/>
                <text y="5" text-anchor="middle" dominant-baseline="middle" fill="#333" font-size="10" font-weight="bold">
                  路由判断
                </text>
              </g>

              <!-- INPUT路径 -->
              <g class="input-path" transform="translate(500, 100)">
                <rect x="-35" y="-20" width="70" height="40" rx="5" fill="url(#inputGradient)" stroke="#388E3C" stroke-width="2"/>
                <text y="5" text-anchor="middle" dominant-baseline="middle" fill="white" font-size="11" font-weight="bold">
                  INPUT
                </text>
                <text y="20" text-anchor="middle" dominant-baseline="middle" fill="#666" font-size="9">
                  mangle → filter
                </text>
              </g>

              <!-- FORWARD路径 -->
              <g class="forward-path" transform="translate(500, 200)">
                <rect x="-35" y="-20" width="70" height="40" rx="5" fill="url(#forwardGradient)" stroke="#D84315" stroke-width="2"/>
                <text y="5" text-anchor="middle" dominant-baseline="middle" fill="white" font-size="11" font-weight="bold">
                  FORWARD
                </text>
                <text y="20" text-anchor="middle" dominant-baseline="middle" fill="#666" font-size="9">
                  mangle → filter
                </text>
              </g>

              <!-- OUTPUT路径 -->
              <g class="output-path" transform="translate(500, 300)">
                <rect x="-35" y="-20" width="70" height="40" rx="5" fill="url(#outputGradient)" stroke="#1976D2" stroke-width="2"/>
                <text y="5" text-anchor="middle" dominant-baseline="middle" fill="white" font-size="11" font-weight="bold">
                  OUTPUT
                </text>
                <text y="20" text-anchor="middle" dominant-baseline="middle" fill="#666" font-size="9">
                  raw → mangle → nat → filter
                </text>
              </g>

              <!-- 本地进程 -->
              <g class="local-process" transform="translate(650, 100)">
                <circle r="20" fill="#4CAF50" stroke="#388E3C" stroke-width="2"/>
                <text y="5" text-anchor="middle" dominant-baseline="middle" fill="white" font-size="10" font-weight="bold">
                  本地进程
                </text>
              </g>

              <!-- POSTROUTING阶段 -->
              <g class="postrouting-stage" transform="translate(650, 250)">
                <rect x="-40" y="-20" width="80" height="40" rx="5" fill="url(#postroutingGradient)" stroke="#7B1FA2" stroke-width="2"/>
                <text y="5" text-anchor="middle" dominant-baseline="middle" fill="white" font-size="11" font-weight="bold">
                  POSTROUTING
                </text>
                <text y="20" text-anchor="middle" dominant-baseline="middle" fill="#666" font-size="9">
                  mangle → nat
                </text>
              </g>

              <!-- 数据包出口 -->
              <g class="packet-exit" transform="translate(800, 250)">
                <circle r="25" fill="#607D8B" stroke="#455A64" stroke-width="2"/>
                <text y="5" text-anchor="middle" dominant-baseline="middle" fill="white" font-size="12" font-weight="bold">
                  数据包出口
                </text>
              </g>

              <!-- 连接线 -->
              <!-- 入口到PREROUTING -->
              <path d="M 75 200 L 160 200" class="packet-flow-line" marker-end="url(#packet-arrow)"/>
              
              <!-- PREROUTING到路由判断 -->
              <path d="M 240 200 L 325 200" class="packet-flow-line" marker-end="url(#packet-arrow)"/>
              
              <!-- 路由判断到INPUT -->
              <path d="M 350 180 L 465 120" class="packet-flow-line" marker-end="url(#packet-arrow)"/>
              
              <!-- 路由判断到FORWARD -->
              <path d="M 350 200 L 465 200" class="packet-flow-line" marker-end="url(#packet-arrow)"/>
              
              <!-- 路由判断到OUTPUT -->
              <path d="M 350 220 L 465 280" class="packet-flow-line" marker-end="url(#packet-arrow)"/>
              
              <!-- INPUT到本地进程 -->
              <path d="M 535 100 L 630 100" class="packet-flow-line" marker-end="url(#packet-arrow)"/>
              
              <!-- FORWARD到POSTROUTING -->
              <path d="M 535 200 L 610 230" class="packet-flow-line" marker-end="url(#packet-arrow)"/>
              
              <!-- OUTPUT到POSTROUTING -->
              <path d="M 535 300 L 610 270" class="packet-flow-line" marker-end="url(#packet-arrow)"/>
              
              <!-- POSTROUTING到出口 -->
              <path d="M 690 250 L 775 250" class="packet-flow-line" marker-end="url(#packet-arrow)"/>
              
              <!-- 本地进程返回路径 -->
              <path d="M 650 120 L 650 230" class="return-flow-line" marker-end="url(#nat-arrow)"/>
            </g>

            <!-- 表处理阶段标注 -->
            <g class="table-labels">
              <text x="200" y="280" text-anchor="middle" class="table-label">raw表</text>
              <text x="200" y="295" text-anchor="middle" class="table-label">mangle表</text>
              <text x="200" y="310" text-anchor="middle" class="table-label">nat表</text>
              
              <text x="500" y="60" text-anchor="middle" class="table-label">mangle表</text>
              <text x="500" y="75" text-anchor="middle" class="table-label">filter表</text>
              
              <text x="500" y="160" text-anchor="middle" class="table-label">mangle表</text>
              <text x="500" y="175" text-anchor="middle" class="table-label">filter表</text>
              
              <text x="500" y="360" text-anchor="middle" class="table-label">raw表</text>
              <text x="500" y="375" text-anchor="middle" class="table-label">mangle表</text>
              <text x="500" y="390" text-anchor="middle" class="table-label">nat表</text>
              <text x="500" y="405" text-anchor="middle" class="table-label">filter表</text>
              
              <text x="650" y="310" text-anchor="middle" class="table-label">mangle表</text>
              <text x="650" y="325" text-anchor="middle" class="table-label">nat表</text>
            </g>

            <!-- 特殊说明 -->
            <g class="special-notes">
              <rect x="50" y="350" width="200" height="60" fill="#FFF3E0" stroke="#FF9800" stroke-width="1" rx="3"/>
              <text x="60" y="370" class="note-text">CentOS6: INPUT链规则不能存在于Nat类中</text>
              <text x="60" y="385" class="note-text">CentOS7: INPUT链规则可以存在于Nat类中</text>
              <text x="60" y="400" class="note-text">IP_FORWARD: /proc/sys/net/ipv4/ip_forward</text>
            </g>
          </svg>

          <!-- 原始拓扑图视图 -->
          <svg 
            v-else
            ref="svgElement" 
            class="topology-svg"
            @click="onCanvasClick"
          >
            <!-- 原有的SVG内容保持不变 -->
            <defs>
              <!-- 普通箭头 -->
              <marker id="arrowhead" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
                <polygon points="0 0, 10 3.5, 0 7" fill="#666" />
              </marker>
              
              <!-- 高亮箭头 -->
              <marker id="arrowhead-highlight" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
                <polygon points="0 0, 10 3.5, 0 7" fill="#409EFF" />
              </marker>
              
              <!-- PREROUTING阶段箭头 -->
              <marker id="arrowhead-prerouting" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
                <polygon points="0 0, 10 3.5, 0 7" fill="#FF5722" />
              </marker>
              
              <!-- INPUT阶段箭头 -->
              <marker id="arrowhead-input" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
                <polygon points="0 0, 10 3.5, 0 7" fill="#4CAF50" />
              </marker>
              
              <!-- FORWARD阶段箭头 -->
              <marker id="arrowhead-forward" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
                <polygon points="0 0, 10 3.5, 0 7" fill="#FF9800" />
              </marker>
              
              <!-- OUTPUT阶段箭头 -->
              <marker id="arrowhead-output" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
                <polygon points="0 0, 10 3.5, 0 7" fill="#2196F3" />
              </marker>
              
              <!-- POSTROUTING阶段箭头 -->
              <marker id="arrowhead-postrouting" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
                <polygon points="0 0, 10 3.5, 0 7" fill="#9C27B0" />
              </marker>
              
              <!-- 渐变定义 -->
              <linearGradient id="interfaceGradient" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" style="stop-color:#409EFF;stop-opacity:1" />
                <stop offset="100%" style="stop-color:#337ecc;stop-opacity:1" />
              </linearGradient>
              
              <linearGradient id="ruleGradient" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" style="stop-color:#E6A23C;stop-opacity:1" />
                <stop offset="100%" style="stop-color:#b88230;stop-opacity:1" />
              </linearGradient>
              
              <!-- 处理阶段渐变 -->
              <linearGradient id="preroutingGradient" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" style="stop-color:#FF5722;stop-opacity:1" />
                <stop offset="100%" style="stop-color:#D84315;stop-opacity:1" />
              </linearGradient>
              
              <linearGradient id="inputGradient" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" style="stop-color:#4CAF50;stop-opacity:1" />
                <stop offset="100%" style="stop-color:#388E3C;stop-opacity:1" />
              </linearGradient>
              
              <linearGradient id="forwardGradient" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" style="stop-color:#FF9800;stop-opacity:1" />
                <stop offset="100%" style="stop-color:#F57C00;stop-opacity:1" />
              </linearGradient>
              
              <linearGradient id="outputGradient" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" style="stop-color:#2196F3;stop-opacity:1" />
                <stop offset="100%" style="stop-color:#1976D2;stop-opacity:1" />
              </linearGradient>
              
              <linearGradient id="postroutingGradient" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" style="stop-color:#9C27B0;stop-opacity:1" />
                <stop offset="100%" style="stop-color:#7B1FA2;stop-opacity:1" />
              </linearGradient>
            </defs>

            <!-- 连接线 -->
            <g class="links">
              <path
                v-for="link in filteredLinks"
                :key="link.id"
                :d="getLinkPath(link)"
                :class="['link', `link-${link.type}`, { 'link-highlighted': isLinkHighlighted(link.id) }]"
                :marker-end="getLinkMarker(link)"
                @mouseenter="onLinkHover(link, true)"
                @mouseleave="onLinkHover(link, false)"
              />
              
              <!-- 规则编号标签 -->
              <text
                v-for="link in filteredLinks"
                :key="`label-${link.id}`"
                :x="getLinkLabelPosition(link).x"
                :y="getLinkLabelPosition(link).y"
                class="link-label"
                text-anchor="middle"
                dominant-baseline="middle"
                v-show="link.rule_number"
              >
                #{{ link.rule_number }}
              </text>
            </g>

            <!-- 节点 -->
            <g class="nodes">
              <g
                v-for="node in filteredNodes"
                :key="node.id"
                :class="['node', `node-${node.type}`, { 'node-highlighted': isNodeHighlighted(node.id) }]"
                :transform="`translate(${node.position.x}, ${node.position.y})`"
                @click="onNodeClick(node)"
                @mouseenter="onNodeHover(node, true)"
                @mouseleave="onNodeHover(node, false)"
              >
                <!-- 接口节点 -->
                <circle
                  v-if="node.type === 'interface'"
                  r="20"
                  :class="['node-circle', `node-interface-${getInterfaceType(node)}`]"
                  fill="url(#interfaceGradient)"
                />
                
                <!-- 规则节点 -->
                <rect
                  v-if="node.type === 'rule'"
                  x="-15"
                  y="-10"
                  width="30"
                  height="20"
                  rx="3"
                  class="node-rect"
                  fill="url(#ruleGradient)"
                />
                
                <!-- 处理阶段节点 -->
                <circle
                  v-if="node.type === 'stage'"
                  r="25"
                  :class="['node-stage', `node-stage-${getStageType(node)}`]"
                  :fill="getStageGradient(node)"
                />
                
                <!-- 路由判断节点 -->
                <polygon
                  v-if="node.type === 'decision'"
                  points="-15,-15 15,-15 0,15"
                  class="node-decision"
                  fill="url(#decisionGradient)"
                />
                
                <!-- 数据包节点 -->
                <circle
                  v-if="node.type === 'packet'"
                  r="12"
                  class="node-packet"
                  fill="#FF4444"
                />
                
                <!-- 节点标签 -->
                <text
                  y="30"
                  class="node-label"
                  text-anchor="middle"
                  dominant-baseline="middle"
                >
                  {{ node.label }}
                </text>
                
                <!-- 接口类型图标 -->
                <text
                  v-if="node.type === 'interface'"
                  y="3"
                  class="node-icon"
                  text-anchor="middle"
                  dominant-baseline="middle"
                >
                  {{ getInterfaceIcon(node) }}
                </text>
                
                <!-- 规则编号 -->
                <text
                  v-if="node.type === 'rule' && node.rule_number"
                  y="3"
                  class="node-rule-number"
                  text-anchor="middle"
                  dominant-baseline="middle"
                >
                  {{ node.rule_number }}
                </text>
                
                <!-- 处理阶段图标 -->
                <text
                  v-if="node.type === 'stage'"
                  y="3"
                  class="node-stage-icon"
                  text-anchor="middle"
                  dominant-baseline="middle"
                >
                  {{ getStageIcon(node) }}
                </text>
                
                <!-- 路由判断图标 -->
                <text
                  v-if="node.type === 'decision'"
                  y="3"
                  class="node-decision-icon"
                  text-anchor="middle"
                  dominant-baseline="middle"
                >
                  ?
                </text>
                
                <!-- 数据包图标 -->
                <text
                  v-if="node.type === 'packet'"
                  y="3"
                  class="node-packet-icon"
                  text-anchor="middle"
                  dominant-baseline="middle"
                >
                  📦
                </text>
              </g>
            </g>
          </svg>
        </div>
      </div>
    </div>

    <!-- 节点详情对话框 -->
    <el-dialog
      v-model="nodeDetailVisible"
      :title="selectedNode ? `${selectedNode.label} - 详细信息` : '节点详情'"
      width="600px"
      :close-on-click-modal="false"
    >
      <div v-if="selectedNode" class="node-detail-content">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="ID">{{ selectedNode.id }}</el-descriptions-item>
          <el-descriptions-item label="类型">{{ selectedNode.type }}</el-descriptions-item>
          <el-descriptions-item label="标签">{{ selectedNode.label }}</el-descriptions-item>
          <el-descriptions-item label="层级">{{ selectedNode.layer }}</el-descriptions-item>
          <el-descriptions-item v-if="selectedNode.interface_name" label="接口名称">
            {{ selectedNode.interface_name }}
          </el-descriptions-item>
          <el-descriptions-item v-if="selectedNode.interface_type" label="接口类型">
            {{ selectedNode.interface_type }}
          </el-descriptions-item>
          <el-descriptions-item v-if="selectedNode.table_name" label="表名称">
            {{ selectedNode.table_name }}
          </el-descriptions-item>
          <el-descriptions-item v-if="selectedNode.chain_name" label="链名称">
            {{ selectedNode.chain_name }}
          </el-descriptions-item>
          <el-descriptions-item v-if="selectedNode.policy" label="策略">
            <el-tag :type="selectedNode.policy === 'ACCEPT' ? 'success' : 'danger'">
              {{ selectedNode.policy }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item v-if="selectedNode.rule_count" label="规则数量">
            {{ selectedNode.rule_count }}
          </el-descriptions-item>
          <el-descriptions-item v-if="selectedNode.rule_number" label="规则编号">
            {{ selectedNode.rule_number }}
          </el-descriptions-item>
          <el-descriptions-item v-if="selectedNode.packets" label="数据包">
            {{ selectedNode.packets }}
          </el-descriptions-item>
          <el-descriptions-item v-if="selectedNode.bytes" label="字节数">
            {{ selectedNode.bytes }}
          </el-descriptions-item>
        </el-descriptions>

        <div v-if="selectedNode.properties" class="node-properties">
          <h4>属性信息</h4>
          <el-descriptions :column="1" border>
            <el-descriptions-item 
              v-for="(value, key) in selectedNode.properties" 
              :key="key"
              :label="key"
            >
              {{ value }}
            </el-descriptions-item>
          </el-descriptions>
        </div>
      </div>
    </el-dialog>

    <!-- 错误提示对话框 -->
    <el-dialog
      v-model="errorDialogVisible"
      title="数据加载错误"
      width="400px"
      :close-on-click-modal="false"
    >
      <div class="error-content">
        <el-alert
          :title="errorMessage"
          type="error"
          :description="errorDetails"
          show-icon
          :closable="false"
        />
        
        <div class="error-actions">
          <el-button @click="retryLoadData" type="primary">重试</el-button>
          <el-button @click="goToDashboard">返回首页</el-button>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowRight } from '@element-plus/icons-vue'
import { topologyAPI, type TopologyData, type TopologyNode, type TopologyLink, type FlowPath, type TopologyStats, type TopologyOptions } from '@/api'

// 响应式数据
const loading = ref(false)
const topologyData = ref<TopologyData | null>(null)
const topologyStats = ref<TopologyStats | null>(null)
const selectedFlow = ref<string>('')
const selectedNode = ref<TopologyNode | null>(null)
const hoveredNode = ref<TopologyNode | null>(null)
const hoveredLink = ref<TopologyLink | null>(null)
const nodeDetailVisible = ref(false)
const errorDialogVisible = ref(false)
const errorMessage = ref('')
const errorDetails = ref('')

// 过滤控件
const protocolFilter = ref<string>('')
const portFilter = ref<string>('')
const chainFilter = ref<string>('')

// 新增：处理阶段过滤
const stageFilter = ref<string>('')

// 自动刷新
const autoRefresh = ref(false)
const refreshInterval = ref<number | null>(null)

// 分页和加载状态
const currentPage = ref(1)
const pageSize = ref(50)
const totalItems = ref(0)

// DOM 引用
const topologyCanvas = ref<HTMLElement>()
const svgElement = ref<SVGElement>()

// 计算属性
const selectedFlowInfo = computed(() => {
  if (!selectedFlow.value || !topologyData.value) return null
  return topologyData.value.flow.find((f: FlowPath) => f.id === selectedFlow.value) || null
})

// 过滤后的节点
const filteredNodes = computed(() => {
  if (!topologyData.value) return []
  
  return topologyData.value.nodes.filter((node: TopologyNode) => {
    // 如果是接口节点，始终显示
    if (node.type === 'interface') return true
    
    // 如果是规则节点，根据过滤条件过滤
    if (node.type === 'rule') {
      if (chainFilter.value && node.chain_name !== chainFilter.value) return false
      if (protocolFilter.value && !node.properties?.protocol?.toLowerCase().includes(protocolFilter.value.toLowerCase())) return false
      if (portFilter.value && 
          !node.properties?.source_port?.includes(portFilter.value) && 
          !node.properties?.dest_port?.includes(portFilter.value)) return false
    }
    
    return true
  })
})

// 过滤后的连接
const filteredLinks = computed(() => {
  if (!topologyData.value) return []
  
  const visibleNodeIds = new Set(filteredNodes.value.map((n: TopologyNode) => n.id))
  
  return topologyData.value.links.filter((link: TopologyLink) => {
    // 只显示两端节点都可见的连接
    if (!visibleNodeIds.has(link.source) || !visibleNodeIds.has(link.target)) return false
    
    // 根据过滤条件过滤
    if (chainFilter.value && link.chain_type !== chainFilter.value) return false
    if (protocolFilter.value && !link.protocol?.toLowerCase().includes(protocolFilter.value.toLowerCase())) return false
    if (portFilter.value && !link.port?.includes(portFilter.value)) return false
    
    return true
  })
})

// 生命周期
onMounted(() => {
  loadTopologyData()
  loadTopologyStats()
})

onUnmounted(() => {
  if (refreshInterval.value) {
    clearInterval(refreshInterval.value)
  }
})

// 方法
const loadTopologyData = async (showLoading = true) => {
  try {
    if (showLoading) {
      loading.value = true
    }
    
    console.log('[DEBUG] Loading topology data...')
    
    const options: TopologyOptions = {
      protocol_filter: protocolFilter.value || undefined,
      chain_filter: chainFilter.value || undefined,
      interface_filter: undefined,
      page: currentPage.value,
      page_size: pageSize.value,
      include_stats: true
    }
    
    const response = await topologyAPI.getTopology(options)
    console.log('[DEBUG] Topology data loaded:', response.data)
    
    if (response.data.success) {
      topologyData.value = response.data.data
      if (response.data.stats) {
        topologyStats.value = response.data.stats
      }
      
      console.log('[DEBUG] Topology nodes:', topologyData.value?.nodes.length)
      console.log('[DEBUG] Topology links:', topologyData.value?.links.length)
      console.log('[DEBUG] Topology flows:', topologyData.value?.flow.length)
    } else {
      throw new Error(response.data.error?.message || 'Failed to load topology data')
    }
  } catch (error: any) {
    console.error('[ERROR] Failed to load topology data:', error)
    errorMessage.value = '加载拓扑图数据失败'
    errorDetails.value = error.response?.data?.error?.details || error.message
    errorDialogVisible.value = true
    
    if (!showLoading) {
      ElMessage.error('刷新失败: ' + (error.response?.data?.error?.message || error.message))
    }
  } finally {
    if (showLoading) {
      loading.value = false
    }
  }
}

const loadTopologyStats = async () => {
  try {
    const response = await topologyAPI.getTopologyStats()
    if (response.data.success) {
      topologyStats.value = response.data.data
    }
  } catch (error) {
    console.error('[ERROR] Failed to load topology stats:', error)
  }
}

const refreshTopology = async () => {
  await loadTopologyData(false)
  ElMessage.success('拓扑图已刷新')
}

const highlightFlow = (flowId: string) => {
  if (selectedFlow.value === flowId) {
    selectedFlow.value = ''
  } else {
    selectedFlow.value = flowId
  }
}

const resetView = () => {
  selectedFlow.value = ''
  selectedNode.value = null
  protocolFilter.value = ''
  portFilter.value = ''
  chainFilter.value = ''
  loadTopologyData()
}

const toggleAutoRefresh = () => {
  if (autoRefresh.value) {
    if (refreshInterval.value) {
      clearInterval(refreshInterval.value)
      refreshInterval.value = null
    }
    autoRefresh.value = false
    ElMessage.success('已停止自动刷新')
  } else {
    autoRefresh.value = true
    refreshInterval.value = window.setInterval(() => {
      loadTopologyData(false)
    }, 30000) // 30秒自动刷新
    ElMessage.success('已开启自动刷新（30秒）')
  }
}

const exportTopology = async () => {
  try {
    const response = await topologyAPI.exportTopology('json')
    const blob = new Blob([JSON.stringify(response.data, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `topology-${new Date().toISOString().split('T')[0]}.json`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
    ElMessage.success('拓扑图已导出')
  } catch (error: any) {
    console.error('[ERROR] Failed to export topology:', error)
    ElMessage.error('导出失败: ' + (error.response?.data?.error?.message || error.message))
  }
}

const onNodeClick = (node: TopologyNode) => {
  selectedNode.value = node
  nodeDetailVisible.value = true
}

const onNodeHover = (node: TopologyNode, isEnter: boolean) => {
  if (isEnter) {
    hoveredNode.value = node
  } else {
    hoveredNode.value = null
  }
}

const onLinkHover = (link: TopologyLink, isEnter: boolean) => {
  if (isEnter) {
    hoveredLink.value = link
  } else {
    hoveredLink.value = null
  }
}

const onCanvasClick = (event: MouseEvent) => {
  // 点击空白区域时的处理
  if (event.target === svgElement.value) {
    selectedFlow.value = ''
    hoveredNode.value = null
    hoveredLink.value = null
  }
}

// 过滤功能
const applyFilters = () => {
  currentPage.value = 1
  loadTopologyData()
}

// 高亮功能
const isNodeHighlighted = (nodeId: string): boolean => {
  if (!selectedFlow.value || !selectedFlowInfo.value) return false
  
  const path = selectedFlowInfo.value.path
  return path.includes(nodeId)
}

const isLinkHighlighted = (linkId: string): boolean => {
  if (!selectedFlow.value || !selectedFlowInfo.value) return false
  
  const link = topologyData.value?.links.find((l: TopologyLink) => l.id === linkId)
  if (!link) return false
  
  const path = selectedFlowInfo.value.path
  for (let i = 0; i < path.length - 1; i++) {
    if (link.source === path[i] && link.target === path[i + 1]) {
      return true
    }
  }
  return false
}

// 工具方法
const getInterfaceType = (node: TopologyNode): string => {
  if (node.interface_type?.includes('ethernet')) return 'ethernet'
  if (node.interface_type?.includes('wifi')) return 'wifi'
  if (node.interface_name?.includes('docker')) return 'docker'
  return 'default'
}

const getInterfaceIcon = (node: TopologyNode): string => {
  if (node.interface_type?.includes('ethernet')) return '🌐'
  if (node.interface_type?.includes('wifi')) return '📡'
  if (node.interface_name?.includes('docker')) return '🐳'
  return '🖧'
}

const getLinkPath = (link: TopologyLink): string => {
  const sourceNode = topologyData.value?.nodes.find((n: TopologyNode) => n.id === link.source)
  const targetNode = topologyData.value?.nodes.find((n: TopologyNode) => n.id === link.target)
  
  if (!sourceNode || !targetNode) return ''
  
  const x1 = sourceNode.position.x
  const y1 = sourceNode.position.y
  const x2 = targetNode.position.x
  const y2 = targetNode.position.y
  
  return `M ${x1} ${y1} L ${x2} ${y2}`
}

const getLinkLabelPosition = (link: TopologyLink) => {
  const sourceNode = topologyData.value?.nodes.find((n: TopologyNode) => n.id === link.source)
  const targetNode = topologyData.value?.nodes.find((n: TopologyNode) => n.id === link.target)
  
  if (!sourceNode || !targetNode) return { x: 0, y: 0 }
  
  return {
    x: (sourceNode.position.x + targetNode.position.x) / 2,
    y: (sourceNode.position.y + targetNode.position.y) / 2
  }
}

const getLinkMarker = (link: TopologyLink): string => {
  switch (link.type) {
    case 'input': return 'url(#arrowhead-input)'
    case 'output': return 'url(#arrowhead-output)'
    case 'forward': return 'url(#arrowhead-forward)'
    default: return 'url(#arrowhead)'
  }
}

// 错误处理
const retryLoadData = () => {
  errorDialogVisible.value = false
  loadTopologyData()
}

const goToDashboard = () => {
  window.location.href = '/'
}
</script>

<style scoped>
.topology-container {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #f5f5f5;
}

.topology-header {
  padding: 20px;
  background: white;
  border-bottom: 1px solid #e4e7ed;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.topology-header h2 {
  margin: 0;
  color: #303133;
}

.topology-controls {
  display: flex;
  gap: 10px;
  align-items: center;
}

.topology-content {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.topology-sidebar {
  width: 300px;
  padding: 20px;
  background: white;
  border-right: 1px solid #e4e7ed;
  overflow-y: auto;
}

.legend-card, .flow-info-card {
  margin-bottom: 20px;
}

.legend-items {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.legend-section {
  border-bottom: 1px solid #e4e7ed;
  padding-bottom: 10px;
}

.legend-section:last-child {
  border-bottom: none;
}

.legend-section h4 {
  margin: 0 0 8px 0;
  font-size: 12px;
  color: #909399;
  font-weight: bold;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 5px;
}

.legend-icon {
  width: 16px;
  height: 16px;
  border-radius: 50%;
}

.legend-line {
  width: 20px;
  height: 2px;
  border-radius: 1px;
}

.legend-action {
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 10px;
  font-weight: bold;
  color: white;
}

/* 接口图标 */
.interface-external-icon {
  background: linear-gradient(45deg, #409EFF, #337ecc);
}

.interface-internal-icon {
  background: linear-gradient(45deg, #67C23A, #529b2e);
}

/* 规则节点样式 */
.rule-bg {
  fill: url(#ruleGradient);
  stroke: #b88230;
  stroke-width: 2;
  transition: all 0.3s ease;
}

.rule-input {
  fill: #4CAF50;
  stroke: #388E3C;
}

.rule-output {
  fill: #2196F3;
  stroke: #1976D2;
}

.rule-forward {
  fill: #FF9800;
  stroke: #F57C00;
}

.rule-label {
  fill: white;
  font-size: 11px;
  font-weight: bold;
  pointer-events: none;
}

.rule-number {
  fill: white;
  font-size: 9px;
  pointer-events: none;
  opacity: 0.9;
}

/* 兼容旧节点样式 */
.node-bg-table {
  fill: #409EFF;
  stroke: #337ecc;
  stroke-width: 2;
}

.node-bg-chain {
  fill: #67C23A;
  stroke: #529b2e;
  stroke-width: 2;
}

.node-bg-rule {
  fill: #E6A23C;
  stroke: #b88230;
  stroke-width: 1;
}

.node-text {
  fill: white;
  font-size: 12px;
  font-weight: bold;
  pointer-events: none;
}

.node-stats {
  fill: white;
  font-size: 10px;
  pointer-events: none;
}

/* 连接线样式 */
.link {
  stroke: #666;
  stroke-width: 2;
  fill: none;
  transition: all 0.3s ease;
  cursor: pointer;
}

.link:hover {
  stroke-width: 3;
  filter: drop-shadow(0 0 3px currentColor);
}

/* 不同类型的连接线 */
.link-input {
  stroke: #4CAF50;
  stroke-width: 2;
}

.link-output {
  stroke: #2196F3;
  stroke-width: 2;
}

.link-forward {
  stroke: #FF9800;
  stroke-width: 2;
}

.link-interface_rule {
  stroke: #67C23A;
  stroke-width: 2;
}

.link-rule_interface {
  stroke: #E6A23C;
  stroke-width: 2;
}

/* 兼容旧连接线样式 */
.link-table_chain {
  stroke: #909399;
  stroke-width: 2;
}

.link-chain_rule {
  stroke: #C0C4CC;
  stroke-width: 1;
}

.link-jump {
  stroke: #F56C6C;
  stroke-width: 2;
  stroke-dasharray: 5,5;
}

.link-highlighted {
  stroke: #409EFF !important;
  stroke-width: 4 !important;
  filter: drop-shadow(0 0 6px #409EFF);
}

/* 连接线标签 */
.link-label {
  fill: #606266;
  font-size: 10px;
  font-weight: bold;
  pointer-events: none;
  background: white;
  padding: 2px 4px;
  border-radius: 3px;
}

.node-detail {
  max-height: 500px;
  overflow-y: auto;
}

.node-properties {
  margin-top: 20px;
}

.node-properties h4 {
  margin: 0 0 10px 0;
  color: #303133;
}

.node-info-card, .link-info-card {
  margin-bottom: 20px;
}

.node-info, .link-info {
  font-size: 13px;
}

.rule-text {
  margin-top: 10px;
  padding: 8px;
  background: #f8f9fa;
  border-radius: 4px;
  border-left: 3px solid #409EFF;
}

.rule-text h5 {
  margin: 0 0 5px 0;
  font-size: 12px;
  color: #606266;
}

.rule-text code {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 11px;
  color: #303133;
  word-break: break-all;
  white-space: pre-wrap;
}

/* 新增样式：统计卡片 */
.stats-card {
  margin-bottom: 20px;
}

.stats-content {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.stat-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 0;
  border-bottom: 1px solid #f0f0f0;
}

.stat-item:last-child {
  border-bottom: none;
}

.stat-label {
  font-size: 12px;
  color: #606266;
}

.stat-value {
  font-size: 14px;
  font-weight: bold;
  color: #303133;
}

/* 新增样式：悬停信息卡片 */
.hover-info-card {
  margin-bottom: 20px;
}

.hover-info-content {
  font-size: 12px;
}

/* 新增样式：错误对话框 */
.error-content {
  text-align: center;
}

.error-actions {
  margin-top: 20px;
  display: flex;
  gap: 10px;
  justify-content: center;
}

/* 响应式设计增强 */
@media (max-width: 768px) {
  .topology-sidebar {
    width: 280px;
  }
  
  .topology-controls {
    flex-wrap: wrap;
  }
}

@media (max-width: 480px) {
  .topology-sidebar {
    width: 220px;
  }
  
  .topology-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }
  
  .topology-controls {
    width: 100%;
    justify-content: flex-start;
  }
}
</style>