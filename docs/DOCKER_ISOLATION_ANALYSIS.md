# Docker隔离规则分析功能

## 功能概述

本次更新为隧道接口分析页面增加了对DOCKER-ISOLATION-STAGE-2链路的检测和分析功能，用于诊断和解决Docker网络隔离规则对隧道通信的影响。

## 新增功能

### 1. 后端增强

#### 1.1 模型扩展
- 在`TunnelDockerAnalysis`模型中添加了`IsolationRules`字段
- 用于存储DOCKER-ISOLATION-STAGE-2链中的相关规则

#### 1.2 规则获取与解析
- 新增`getDockerIsolationRules()`方法：获取DOCKER-ISOLATION-STAGE-2链规则
- 新增`parseDockerIsolationRules()`方法：解析隔离规则，重点识别DROP规则
- 支持检测以下类型的阻断规则：
  - 针对特定Docker网桥的DROP规则
  - 针对隧道接口的DROP规则  
  - 通用隔离规则（如 any -> br-xxx）

#### 1.3 通信路径分析增强
- 新增`generateCommunicationPathWithIsolation()`方法
- 在通信路径中添加Docker隔离规则检查步骤
- 显示隔离规则对通信的影响状态

#### 1.4 智能建议系统
- 在`generateRecommendationsWithTest()`中添加隔离规则分析
- 当检测到DROP规则时，提供具体的修复建议
- 包含具体的iptables命令建议

#### 1.5 一键修复功能增强
- 在`FixConnectivity()`中添加`fixDockerIsolationRules()`处理
- 自动在DOCKER-ISOLATION-STAGE-2链开头添加允许规则
- 支持双向通信规则的自动配置

### 2. 前端增强

#### 2.1 界面扩展
- 在分析结果中新增"隔离规则"标签页
- 显示DOCKER-ISOLATION-STAGE-2链中的相关规则
- 对DROP规则使用红色标签突出显示

#### 2.2 智能提示
- 添加`hasIsolationDropRules`计算属性
- 当检测到DROP规则时显示警告提示
- 提供隔离规则影响的详细说明

## 使用场景

### 典型问题场景
```bash
# 常见的Docker隔离规则示例
iptables -L DOCKER-ISOLATION-STAGE-2 -v
Chain DOCKER-ISOLATION-STAGE-2 (1 references)
 pkts bytes target prot opt in     out     source        destination
    0     0 DROP   all  --  any    br-xxx   anywhere      anywhere
```

### 分析流程
1. 用户选择隧道接口和Docker网桥
2. 系统自动获取DOCKER-ISOLATION-STAGE-2链规则
3. 分析DROP规则对通信的影响
4. 在通信路径中显示隔离检查步骤
5. 提供具体的修复建议

### 修复建议示例
当检测到阻断规则时，系统会提供如下建议：
```bash
# 添加允许规则到隔离链开头
iptables -I DOCKER-ISOLATION-STAGE-2 1 -i tun0 -o br-xxx -j ACCEPT
iptables -I DOCKER-ISOLATION-STAGE-2 2 -i br-xxx -o tun0 -j ACCEPT
```

## 技术实现

### 关键方法
- `getDockerIsolationRules()`: 获取隔离规则
- `parseDockerIsolationRules()`: 解析规则内容
- `fixDockerIsolationRules()`: 自动修复隔离问题
- `generateCommunicationPathWithIsolation()`: 生成包含隔离检查的通信路径

### 规则匹配逻辑
- 精确匹配：接口名称完全匹配
- 模糊匹配：规则文本包含接口名称
- 通配符匹配：支持"any"和"br-*"模式

## 优势特点

1. **全面诊断**：覆盖Docker网络隔离的完整检查
2. **智能识别**：自动识别可能影响通信的DROP规则
3. **精确定位**：提供具体的规则行号和内容
4. **一键修复**：自动生成和应用修复规则
5. **可视化展示**：在通信路径中清晰显示隔离检查步骤

## 注意事项

1. 修复操作需要root权限
2. 修复规则会添加到链的开头，优先级最高
3. 建议在测试环境中验证修复效果
4. 某些Docker版本可能没有DOCKER-ISOLATION-STAGE-2链

## 兼容性

- 支持Docker 17.06+版本
- 兼容标准的iptables规则格式
- 适用于bridge网络模式的Docker容器