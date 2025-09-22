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
        <div class="controls">
          <el-select v-model="selectedInterface" placeholder="选择网络接口" @change="handleInterfaceChange" clearable>
            <el-option label="全部接口" value="" />
            <el-option
              v-for="iface in interfaces"
              :key="iface.name"
              :label="`${iface.name} (${iface.ip_addresses.join(', ')})`"
              :value="iface.name"
            />
          </el-select>
          
          <el-select v-model="viewMode" @change="handleViewModeChange">
            <el-option label="链视图" value="chain" />
            <el-option label="表视图" value="table" />
            <el-option label="接口视图" value="interface" />
          </el-select>
          
          <el-button @click="refreshData" :loading="loading" type="primary">
            <el-icon><Refresh /></el-icon>
            刷新数据
          </el-button>
          

        </div>
      </el-card>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="8" animated />
    </div>



    <!-- 主要内容区域 -->
    <div class="main-content">
      <!-- 数据流图视图 -->
      <div v-if="viewMode === 'chain'" class="dataflow-view">
        <!-- 上层协议栈 -->
        <div class="protocol-stack">
          <div class="protocol-box">
            <span>上层协议栈</span>
          </div>
        </div>
        
        <!-- 虚线分隔 -->
        <div class="dashed-line"></div>
        
        <!-- 主要数据流 -->
        <div class="main-flow">
          <!-- 数据包入口 -->
          <div class="flow-start">
            <div class="flow-node start-node">
              <span>数据包入口</span>
            </div>
          </div>
          
          <!-- PREROUTING 链 -->
          <div class="chain-section prerouting">
            <div class="chain-box" @click="selectChain('PREROUTING')">
              <div class="chain-title">PREROUTING</div>
              <div class="chain-tables">
                <div class="table-tag raw" @click.stop="selectChainTable('PREROUTING', 'raw')">raw</div>
                <div class="table-tag mangle" @click.stop="selectChainTable('PREROUTING', 'mangle')">mangle</div>
                <div class="table-tag nat" @click.stop="selectChainTable('PREROUTING', 'nat')">nat</div>
              </div>
              <div class="rule-count">{{ getChainRuleCount('PREROUTING') }} 规则</div>
            </div>
          </div>
          
          <!-- 路由决策点 -->
          <div class="routing-decision">
            <div class="decision-diamond">
              <span>路由决策</span>
            </div>
          </div>
          
          <!-- 分流：本机设备 vs 非本机设备 -->
          <div class="flow-split">
            <!-- 本机设备路径 -->
            <div class="local-path">
              <div class="path-label">本机设备</div>
              
              <!-- INPUT 链 -->
              <div class="chain-section input">
                <div class="chain-box" @click="selectChain('INPUT')">
                  <div class="chain-title">INPUT</div>
                  <div class="chain-tables">
                    <div class="table-tag mangle" @click.stop="selectChainTable('INPUT', 'mangle')">mangle</div>
                    <div class="table-tag nat" @click.stop="selectChainTable('INPUT', 'nat')">nat</div>
                    <div class="table-tag filter" @click.stop="selectChainTable('INPUT', 'filter')">filter</div>
                  </div>
                  <div class="rule-count">{{ getChainRuleCount('INPUT') }} 规则</div>
                </div>
              </div>
              
              <!-- 本地进程 -->
              <div class="local-process">
                <div class="process-box">
                  <span>本地进程</span>
                </div>
              </div>
              
              <!-- OUTPUT 链 -->
              <div class="chain-section output">
                <div class="chain-box" @click="selectChain('OUTPUT')">
                  <div class="chain-title">OUTPUT</div>
                  <div class="chain-tables">
                    <div class="table-tag raw" @click.stop="selectChainTable('OUTPUT', 'raw')">raw</div>
                    <div class="table-tag mangle" @click.stop="selectChainTable('OUTPUT', 'mangle')">mangle</div>
                    <div class="table-tag nat" @click.stop="selectChainTable('OUTPUT', 'nat')">nat</div>
                    <div class="table-tag filter" @click.stop="selectChainTable('OUTPUT', 'filter')">filter</div>
                  </div>
                  <div class="rule-count">{{ getChainRuleCount('OUTPUT') }} 规则</div>
                </div>
              </div>
            </div>
            
            <!-- 非本机设备路径 -->
            <div class="forward-path">
              <div class="path-label">非本机设备</div>
              <div class="path-sublabel">ip_forward=1</div>
              
              <!-- FORWARD 链 -->
              <div class="chain-section forward">
                <div class="chain-box" @click="selectChain('FORWARD')">
                  <div class="chain-title">FORWARD</div>
                  <div class="chain-tables">
                    <div class="table-tag mangle" @click.stop="selectChainTable('FORWARD', 'mangle')">mangle</div>
                    <div class="table-tag filter" @click.stop="selectChainTable('FORWARD', 'filter')">filter</div>
                  </div>
                  <div class="rule-count">{{ getChainRuleCount('FORWARD') }} 规则</div>
                </div>
              </div>
              
              <!-- 路由决策点2 -->
              <div class="routing-decision">
                <div class="decision-diamond">
                  <span>输出路由选择</span>
                  <div class="decision-sublabel">根据路由表选择</div>
                </div>
              </div>
            </div>
          </div>
          
          <!-- POSTROUTING 链 -->
          <div class="chain-section postrouting">
            <div class="chain-box" @click="selectChain('POSTROUTING')">
              <div class="chain-title">POSTROUTING</div>
              <div class="chain-tables">
                <div class="table-tag mangle" @click.stop="selectChainTable('POSTROUTING', 'mangle')">mangle</div>
                <div class="table-tag nat" @click.stop="selectChainTable('POSTROUTING', 'nat')">nat</div>
              </div>
              <div class="rule-count">{{ getChainRuleCount('POSTROUTING') }} 规则</div>
            </div>
          </div>
          
          <!-- 数据包出口 -->
          <div class="flow-end">
            <div class="flow-node end-node">
              <span>数据包出口</span>
            </div>
          </div>
        </div>
        
        <!-- 流向箭头 -->
        <div class="flow-arrows">
          <!-- 这里会通过CSS添加箭头 -->
        </div>
      </div>

      <!-- 表视图 -->
      <div v-if="viewMode === 'table'" class="table-view">
        <div class="tables-grid">
          <div
            v-for="table in tables"
            :key="table.name"
            class="table-card"
            :class="table.name"
          >
            <el-card>
              <template #header>
                <div class="table-header">
                  <h3>{{ table.name.toUpperCase() }} 表</h3>
                  <el-tag>{{ table.totalRules }} 规则</el-tag>
                </div>
              </template>
              
              <div class="table-chains">
                <div
                  v-for="chain in table.chains"
                  :key="chain.name"
                  class="chain-in-table"
                  @click="selectChainInTable(table.name, chain.name)"
                >
                  <div class="chain-name">{{ chain.name }}</div>
                  <div class="chain-rules-count">{{ (chain.rules || []).length }}</div>
                  <div class="chain-policy" v-if="chain.policy">
                    策略: {{ chain.policy }}
                  </div>
                </div>
              </div>
            </el-card>
          </div>
        </div>
      </div>

      <!-- 接口视图 -->
      <div v-if="viewMode === 'interface'" class="interface-view">
        <div class="interfaces-container">
          <div
            v-for="iface in interfaceData"
            :key="iface.name"
            class="interface-card"
          >
            <el-card>
              <template #header>
                <div class="interface-header">
                  <h3>{{ iface.name }}</h3>
                  <div class="interface-info">
                    <el-tag :type="iface.is_up ? 'success' : 'danger'">
                      {{ iface.is_up ? '启用' : '禁用' }}
                    </el-tag>
                    <el-tag v-if="iface.is_docker" type="info">Docker</el-tag>
                  </div>
                </div>
              </template>
              
              <div class="interface-content">
                <div class="interface-details">
                  <p><strong>IP地址:</strong> {{ iface.ip_addresses.join(', ') || '无' }}</p>
                  <p><strong>MAC地址:</strong> {{ iface.mac_address || '无' }}</p>
                  <p><strong>类型:</strong> {{ iface.type }}</p>
                </div>
                
                <div class="interface-rules">
                  <h4>相关规则统计</h4>
                  <div class="rules-stats">
                    <div class="stat-item">
                      <span class="stat-label">输入规则:</span>
                      <span class="stat-value">{{ iface.inRules }}</span>
                    </div>
                    <div class="stat-item">
                      <span class="stat-label">输出规则:</span>
                      <span class="stat-value">{{ iface.outRules }}</span>
                    </div>
                    <div class="stat-item">
                      <span class="stat-label">转发规则:</span>
                      <span class="stat-value">{{ iface.forwardRules }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </el-card>
          </div>
        </div>
      </div>
    </div>

    <!-- 详细信息弹窗 -->
    <el-dialog
      v-model="showDetailDialog"
      :title="detailTitle"
      width="95%"
      :before-close="closeDetailDialog"
    >
      <div class="detail-content">
        <!-- 控制面板 -->
        <div class="detail-controls">
          <div class="detail-left-controls">
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
          <div class="detail-right-controls">
            <el-button type="primary" size="small" @click="showAddRuleDialog">
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

    <!-- 添加/编辑规则对话框 -->
    <el-dialog
      v-model="ruleDialogVisible"
      :title="isEditRule ? '编辑规则' : '添加规则'"
      width="700px"
      @close="resetRuleForm"
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
          <el-button @click="ruleDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="submitRuleForm" :loading="rulesLoading">确定</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, reactive } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import { Refresh, Right, Setting, Plus, Monitor, Search, Edit, Delete } from '@element-plus/icons-vue'
import { getChainTableData, getNetworkInterfaces } from '../api/chainTable'
import { apiService } from '@/api'

// 响应式数据
const loading = ref(false)
const selectedInterface = ref('')
const viewMode = ref('chain')
const selectedChain = ref('')
const showDetailDialog = ref(false)
const detailTitle = ref('')
const detailRules = ref([])
const groupByChain = ref(true)

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

// 获取链的规则数量
const getChainRuleCount = (chainName: string) => {
  const chain = chains.value.find(c => c.name === chainName)
  return chain ? (chain.rules || []).length : 0
}

// 选择链和表
const selectChainTable = (chainName: string, tableName: string) => {
  selectedChain.value = chainName
  showDetailDialog.value = true
  detailTitle.value = `${chainName} - ${tableName.toUpperCase()} 表详细规则`
  
  // 获取特定链和表的规则
  const chain = chains.value.find(c => c.name === chainName)
  if (chain) {
    const table = chain.tables?.find(t => t.name === tableName)
    detailRules.value = table ? (table.rules || []) : []
  } else {
    detailRules.value = []
  }
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
  Object.keys(groups).forEach(chainName => {
    groups[chainName].sort((a: any, b: any) => {
      const lineA = parseInt(a.line_number || '0', 10)
      const lineB = parseInt(b.line_number || '0', 10)
      return lineA - lineB
    })
  })
  
  return groups
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

// 方法
const refreshData = async () => {
  loading.value = true
  try {
    // 自动比对并同步系统规则
    const synced = await autoSyncSystemRules()
    
    // 加载数据
    await Promise.all([
      loadChainTableData(),
      loadInterfaces()
    ])
    

    
    // 根据是否进行了同步显示不同的消息
    if (synced) {
      ElMessage.success('检测到数据不一致，已自动同步并刷新数据')
    } else {
      ElMessage.success('数据刷新成功')
    }
  } catch (error) {
    console.error('刷新数据失败:', error)
    ElMessage.error('数据刷新失败')
  } finally {
    loading.value = false
  }
}

const loadChainTableData = async () => {
  try {
    const data = await getChainTableData(selectedInterface.value)
    chainTableData.value = data
  } catch (error) {
    console.error('加载链表数据失败:', error)
    // 如果API调用失败，使用模拟数据
    chainTableData.value = getMockChainTableData()
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
          { id: 1, chain_name: 'PREROUTING', table: 'mangle', rule_text: 'MARK --set-mark 1' },
          { id: 2, chain_name: 'PREROUTING', table: 'nat', rule_text: 'DNAT --to-destination 192.168.1.100' }
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
          { id: 3, chain_name: 'INPUT', table: 'filter', rule_text: 'ACCEPT -p tcp --dport 22' },
          { id: 4, chain_name: 'INPUT', table: 'filter', rule_text: 'ACCEPT -p tcp --dport 80' },
          { id: 5, chain_name: 'INPUT', table: 'filter', rule_text: 'DROP -p tcp --dport 23' }
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
    const data = await getNetworkInterfaces()
    interfaces.value = data
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

const handleInterfaceChange = () => {
  loadChainTableData()
}

const handleViewModeChange = () => {
  selectedChain.value = ''
}

const selectChain = (chainName: string) => {
  selectedChain.value = chainName
  const chain = chains.value.find((c: any) => c.name === chainName)
  if (chain) {
    detailTitle.value = `${chainName} 链详细规则`
    detailRules.value = chain.rules || []
    showDetailDialog.value = true
  }
}

const selectChainInTable = (tableName: string, chainName: string) => {
  const table = tables.value.find((t: any) => t.name === tableName)
  if (table) {
    const chain = table.chains.find((c: any) => c.name === chainName)
    if (chain) {
      detailTitle.value = `${tableName}.${chainName} 详细规则`
      detailRules.value = chain.rules || []
      showDetailDialog.value = true
    }
  }
}

const closeDetailDialog = () => {
  showDetailDialog.value = false
  detailRules.value = []
  groupByChain.value = true
}

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
  const rules = chainTableData.value.interfaceRules?.[interfaceName] || []
  return rules.filter((rule: any) => {
    if (direction === 'in') {
      return rule.in_interface === interfaceName || rule.interface_in === interfaceName
    } else if (direction === 'out') {
      return rule.out_interface === interfaceName || rule.interface_out === interfaceName
    } else if (direction === 'forward') {
      return rule.chain_name === 'FORWARD' && 
             (rule.in_interface === interfaceName || rule.out_interface === interfaceName)
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

const showAddRuleDialog = () => {
  isEditRule.value = false
  ruleDialogVisible.value = true
  resetRuleForm()
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
        ruleDialogVisible.value = false
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
  ruleDialogVisible.value = true
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

// 生命周期
onMounted(async () => {
  loading.value = true
  try {
    // 页面加载时自动比对并同步
    await autoSyncSystemRules()
    
    // 加载数据
    await Promise.all([
      loadChainTableData(),
      loadInterfaces()
    ])
  } catch (error) {
    console.error('初始化数据失败:', error)
    ElMessage.error('初始化数据失败')
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.chain-table-view {
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;
}

.page-header h1 {
  color: #303133;
  margin-bottom: 8px;
}

.description {
  color: #606266;
  font-size: 14px;
}

.control-panel {
  margin-bottom: 20px;
}

.controls {
  display: flex;
  gap: 16px;
  align-items: center;
}

.loading-container {
  margin-top: 20px;
}

.main-content {
  margin-top: 20px;
}

/* 链视图样式 */
.chains-container {
  display: flex;
  gap: 16px;
  overflow-x: auto;
  padding: 16px 0;
}

.chain-card {
  min-width: 280px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.chain-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.chain-card.active {
  border: 2px solid #409EFF;
}

.chain-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.chain-header h3 {
  margin: 0;
  color: #303133;
}

.chain-content {
  position: relative;
}

.tables-in-chain {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.table-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  border-radius: 4px;
  font-size: 12px;
}

.table-item.raw {
  background-color: #f0f9ff;
  border-left: 3px solid #0ea5e9;
}

.table-item.mangle {
  background-color: #fefce8;
  border-left: 3px solid #eab308;
}

.table-item.nat {
  background-color: #f0fdf4;
  border-left: 3px solid #22c55e;
}

.table-item.filter {
  background-color: #fef2f2;
  border-left: 3px solid #ef4444;
}

.table-name {
  font-weight: 600;
  text-transform: uppercase;
}

.table-rules {
  color: #6b7280;
}

.flow-indicator {
  position: absolute;
  right: -20px;
  top: 50%;
  transform: translateY(-50%);
  color: #409EFF;
  font-size: 20px;
}

/* 表视图样式 */
.tables-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 20px;
}

.table-card {
  transition: all 0.3s ease;
}

.table-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.table-card.raw :deep(.el-card__header) {
  background-color: #f0f9ff;
  border-bottom: 2px solid #0ea5e9;
}

.table-card.mangle :deep(.el-card__header) {
  background-color: #fefce8;
  border-bottom: 2px solid #eab308;
}

.table-card.nat :deep(.el-card__header) {
  background-color: #f0fdf4;
  border-bottom: 2px solid #22c55e;
}

.table-card.filter :deep(.el-card__header) {
  background-color: #fef2f2;
  border-bottom: 2px solid #ef4444;
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.table-header h3 {
  margin: 0;
  color: #303133;
}

.table-chains {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.chain-in-table {
  padding: 12px;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.chain-in-table:hover {
  border-color: #409EFF;
  background-color: #f5f7fa;
}

.chain-name {
  font-weight: 600;
  color: #303133;
  margin-bottom: 4px;
}

.chain-rules-count {
  font-size: 12px;
  color: #909399;
}

.chain-policy {
  font-size: 12px;
  color: #606266;
  margin-top: 4px;
}

/* 接口视图样式 */
.interfaces-container {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: 20px;
}

.interface-card {
  transition: all 0.3s ease;
}

.interface-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.interface-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.interface-header h3 {
  margin: 0;
  color: #303133;
}

.interface-info {
  display: flex;
  gap: 8px;
}

.interface-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.interface-details p {
  margin: 8px 0;
  font-size: 14px;
  color: #606266;
}

.interface-rules h4 {
  margin: 0 0 12px 0;
  color: #303133;
  font-size: 16px;
}

.rules-stats {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.stat-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.stat-label {
  color: #606266;
  font-size: 14px;
}

.stat-value {
  color: #303133;
  font-weight: 600;
}

/* 详细信息弹窗样式 */
.detail-content {
  max-height: 75vh;
  overflow-y: auto;
}

.detail-controls {
  margin-bottom: 20px;
  padding: 16px;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  border-radius: 8px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.detail-left-controls {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
}

.detail-right-controls {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.grouped-rules {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.rule-group {
  border: 1px solid #e4e7ed;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  transition: all 0.3s ease;
}

.rule-group:hover {
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.12);
  transform: translateY(-2px);
}

.group-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border-bottom: none;
}

.group-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.group-title h4 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}

.group-stats {
  display: flex;
  align-items: center;
  gap: 16px;
  font-size: 14px;
  opacity: 0.9;
}

.stat-item {
  background: rgba(255, 255, 255, 0.2);
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
}

.chain-rules-table {
  border: none;
}

.chain-rules-table :deep(.el-table__header) {
  background-color: #fafbfc;
}

.chain-rules-table :deep(.el-table__header th) {
  background-color: #fafbfc;
  color: #606266;
  font-weight: 600;
  border-bottom: 2px solid #e4e7ed;
}

.chain-rules-table :deep(.el-table__row:hover) {
  background-color: #f0f9ff;
}

.chain-rules-table :deep(.el-table__row:hover td) {
  background-color: #f0f9ff;
}

.no-target {
  color: #c0c4cc;
  font-style: italic;
}

.list-rules {
  margin-top: 16px;
}



/* 响应式设计 */
@media (max-width: 768px) {
  .chains-container {
    flex-direction: column;
  }
  
  .chain-card {
    min-width: auto;
  }
  
  .controls {
    flex-direction: column;
    align-items: stretch;
  }
  
  .tables-grid {
    grid-template-columns: 1fr;
  }
  
  .interfaces-container {
    grid-template-columns: 1fr;
  }
  
  .detail-controls {
    flex-direction: column;
    gap: 12px;
    align-items: stretch;
  }
  
  .detail-right-controls {
    flex-wrap: wrap;
    justify-content: center;
  }
  
  .detail-right-controls .el-input {
    width: 100% !important;
    margin-left: 0 !important;
    margin-top: 8px;
  }
}

/* 表格样式优化 */
.chain-rules-table .el-table__header-wrapper {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.chain-rules-table .el-table__header-wrapper th {
  background: transparent !important;
  color: white !important;
  font-weight: 600;
}

.no-target, .no-interface {
  color: #909399;
  font-style: italic;
}

/* 规则表单样式 */
.rule-form .el-divider {
  margin: 20px 0 16px 0;
}

.rule-form .el-divider__text {
  background: #f5f7fa;
  color: #606266;
  font-weight: 600;
  padding: 0 16px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 16px;
  border-top: 1px solid #e4e7ed;
}

/* 标签样式优化 */
.el-tag {
  border-radius: 4px;
  font-weight: 500;
}

/* 数据流图样式 */
.dataflow-view {
  padding: 20px;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  border-radius: 12px;
  min-height: 600px;
  position: relative;
}

.protocol-stack {
  text-align: center;
  margin-bottom: 20px;
}

.protocol-box {
  display: inline-block;
  padding: 12px 24px;
  background: #e8f5e8;
  border: 2px solid #4caf50;
  border-radius: 8px;
  font-weight: bold;
  color: #2e7d32;
}

.dashed-line {
  height: 2px;
  background: repeating-linear-gradient(
    to right,
    #666 0px,
    #666 10px,
    transparent 10px,
    transparent 20px
  );
  margin: 20px 0;
}

.main-flow {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 30px;
}

.flow-start, .flow-end {
  display: flex;
  justify-content: center;
}

.flow-node {
  padding: 12px 24px;
  border-radius: 50px;
  font-weight: bold;
  color: white;
}

.start-node {
  background: linear-gradient(135deg, #42a5f5, #1976d2);
}

.end-node {
  background: linear-gradient(135deg, #42a5f5, #1976d2);
}

.chain-section {
  display: flex;
  justify-content: center;
  margin: 10px 0;
}

.chain-box {
  background: white;
  border-radius: 12px;
  padding: 16px 20px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  cursor: pointer;
  transition: all 0.3s ease;
  min-width: 200px;
  text-align: center;
}

.chain-box:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.2);
}

.chain-title {
  font-size: 16px;
  font-weight: bold;
  margin-bottom: 8px;
  color: #333;
}

.chain-tables {
  display: flex;
  justify-content: center;
  gap: 6px;
  margin-bottom: 8px;
  flex-wrap: wrap;
}

.table-tag {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: bold;
  color: white;
  cursor: pointer;
  transition: all 0.2s ease;
}

.table-tag:hover {
  transform: scale(1.05);
}

.table-tag.raw {
  background: #ff9800;
}

.table-tag.mangle {
  background: #9c27b0;
}

.table-tag.nat {
  background: #4caf50;
}

.table-tag.filter {
  background: #f44336;
}

.rule-count {
  font-size: 12px;
  color: #666;
}

.routing-decision {
  display: flex;
  justify-content: center;
  margin: 20px 0;
}

.decision-diamond {
  background: #fff3e0;
  border: 2px solid #ff9800;
  padding: 16px;
  border-radius: 8px;
  text-align: center;
  font-weight: bold;
  color: #e65100;
  position: relative;
}

.decision-sublabel {
  font-size: 12px;
  font-weight: normal;
  color: #666;
  margin-top: 4px;
}

.flow-split {
  display: flex;
  justify-content: space-around;
  width: 100%;
  gap: 40px;
  margin: 30px 0;
}

.local-path, .forward-path {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
}

.path-label {
  font-weight: bold;
  color: #333;
  font-size: 14px;
}

.path-sublabel {
  font-size: 12px;
  color: #666;
  margin-top: -10px;
}

.local-process {
  display: flex;
  justify-content: center;
}

.process-box {
  padding: 12px 20px;
  background: #e3f2fd;
  border: 2px solid #2196f3;
  border-radius: 8px;
  font-weight: bold;
  color: #1565c0;
}

/* 特定链的颜色 */
.prerouting .chain-box {
  border-left: 4px solid #ff9800;
}

.input .chain-box {
  border-left: 4px solid #4caf50;
}

.forward .chain-box {
  border-left: 4px solid #2196f3;
}

.output .chain-box {
  border-left: 4px solid #9c27b0;
}

.postrouting .chain-box {
  border-left: 4px solid #f44336;
}

/* 数据流图响应式设计 */
@media (max-width: 768px) {
  .dataflow-view {
    padding: 10px;
  }
  
  .flow-split {
    flex-direction: column;
    gap: 20px;
  }
  
  .chain-box {
    min-width: 150px;
  }
  
  .chain-tables {
    gap: 4px;
  }
  
  .table-tag {
    font-size: 10px;
    padding: 2px 6px;
  }
}
</style>