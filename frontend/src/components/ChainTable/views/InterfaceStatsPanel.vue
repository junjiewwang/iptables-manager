<template>
  <div class="stats-panel">
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card class="stats-card">
          <div class="stats-content">
            <div class="stats-icon">
              <el-icon><Connection /></el-icon>
            </div>
            <div class="stats-info">
              <div class="stats-number">{{ interfacesCount }}</div>
              <div class="stats-label">总接口数</div>
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
              <el-icon><Ship /></el-icon>
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
              <div class="stats-label">相关规则</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { Connection, Check, Ship, List } from '@element-plus/icons-vue'

interface Props {
  interfacesCount: number
  activeInterfacesCount: number
  dockerInterfacesCount: number
  totalInterfaceRules: number
}

const props = defineProps<Props>()
</script>

<style scoped>
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
</style>