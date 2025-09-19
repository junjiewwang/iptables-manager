# 拓扑图API文档

## 概述

本文档描述了iptables管理系统的拓扑图API接口规范，支持网络拓扑数据获取、过滤、统计和导出功能。

## 接口规范

### 1. 获取拓扑数据

**接口地址**: `GET /api/topology`

**描述**: 获取完整的网络拓扑图数据，包括节点、连接和数据流信息。

**查询参数**:

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| protocol | string | 否 | 协议过滤（tcp/udp/icmp等） |
| chain | string | 否 | 链类型过滤（INPUT/OUTPUT/FORWARD） |
| interface | string | 否 | 接口名称过滤 |
| rule_type | string | 否 | 规则类型过滤 |
| page | int | 否 | 页码（默认1） |
| page_size | int | 否 | 每页大小（默认50） |
| include_stats | boolean | 否 | 是否包含统计信息 |
| include_metadata | boolean | 否 | 是否包含元数据 |

**响应格式**:
```json
{
  "success": true,
  "data": {
    "nodes": [
      {
        "id": "interface_eth0",
        "label": "eth0",
        "type": "interface",
        "interface_name": "eth0",
        "interface_type": "ethernet",
        "layer": 1,
        "position": { "x": 100, "y": 50 },
        "properties": {
          "type": "ethernet",
          "state": "up",
          "mac_address": "00:11:22:33:44:55",
          "mtu": "1500",
          "is_up": "true",
          "is_docker": "false",
          "rx_bytes": "1234567",
          "tx_bytes": "7654321"
        }
      }
    ],
    "links": [
      {
        "id": "link_interface_eth0_to_rule_1",
        "source": "interface_eth0",
        "target": "rule_filter_INPUT_1",
        "type": "input",
        "chain_type": "INPUT",
        "action": "ACCEPT",
        "protocol": "tcp",
        "rule_number": 1,
        "properties": {
          "direction": "input",
          "table": "filter",
          "chain": "INPUT"
        }
      }
    ],
    "flow": [
      {
        "id": "flow_input",
        "name": "入站数据流",
        "description": "外部网络接口通过INPUT规则进入系统的数据流",
        "path": ["interface_eth0", "rule_filter_INPUT_1"],
        "color": "#4CAF50"
      }
    ]
  },
  "meta": {
    "timestamp": 1640995200,
    "version": "1.0"
  }
}
```

**错误响应**:
```json
{
  "success": false,
  "error": {
    "message": "Failed to get topology data",
    "code": "TOPOLOGY_FETCH_ERROR",
    "details": "详细错误信息"
  }
}
```

### 2. 获取拓扑统计信息

**接口地址**: `GET /api/topology/stats`

**描述**: 获取拓扑图的统计信息，包括节点数量、连接数量等。

**响应格式**:
```json
{
  "success": true,
  "data": {
    "total_nodes": 15,
    "total_links": 25,
    "total_flows": 4,
    "node_types": {
      "interface": 5,
      "rule": 10
    },
    "chain_types": {
      "INPUT": 4,
      "OUTPUT": 3,
      "FORWARD": 3
    },
    "interface_types": {
      "ethernet": 2,
      "docker": 3
    },
    "generated_at": 1640995200
  },
  "meta": {
    "timestamp": 1640995200
  }
}
```

### 3. 强制刷新拓扑缓存

**接口地址**: `POST /api/topology/refresh`

**描述**: 强制刷新拓扑数据缓存，获取最新的网络状态。

**响应格式**:
```json
{
  "success": true,
  "data": {
    // 同获取拓扑数据的响应格式
  },
  "meta": {
    "timestamp": 1640995200,
    "refreshed": true
  }
}
```

### 4. 导出拓扑数据

**接口地址**: `GET /api/topology/export`

**描述**: 导出拓扑数据为指定格式。

**查询参数**:

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| format | string | 否 | 导出格式（json/csv，默认json） |

**响应格式**:
- JSON格式: 直接返回JSON数据
- CSV格式: 返回CSV文本内容

**响应头**:
```
Content-Type: application/json
Content-Disposition: attachment; filename=topology.json
```

### 5. 获取拓扑服务健康状态

**接口地址**: `GET /api/topology/health`

**描述**: 检查拓扑服务的健康状态。

**响应格式**:
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "response_time": 150,
    "timestamp": 1640995200
  }
}
```

## 数据模型

### 节点类型 (TopologyNode)

| 字段名 | 类型 | 描述 |
|--------|------|------|
| id | string | 节点唯一标识 |
| label | string | 节点显示标签 |
| type | string | 节点类型（interface/table/chain/rule） |
| interface_name | string | 接口名称（接口节点） |
| interface_type | string | 接口类型（接口节点） |
| table_name | string | 表名称（规则节点） |
| chain_name | string | 链名称（规则节点） |
| policy | string | 策略（链节点） |
| rule_count | int | 规则数量（链节点） |
| rule_number | int | 规则编号（规则节点） |
| packets | string | 数据包计数 |
| bytes | string | 字节计数 |
| properties | object | 扩展属性 |
| position | object | 位置坐标 {x, y} |
| layer | int | 层级（用于布局） |

### 连接类型 (TopologyLink)

| 字段名 | 类型 | 描述 |
|--------|------|------|
| id | string | 连接唯一标识 |
| source | string | 源节点ID |
| target | string | 目标节点ID |
| type | string | 连接类型（input/output/forward） |
| label | string | 连接标签 |
| rule_text | string | 规则文本 |
| rule_number | int | 规则编号 |
| chain_type | string | 链类型（INPUT/OUTPUT/FORWARD） |
| action | string | 动作（ACCEPT/DROP/REJECT） |
| protocol | string | 协议类型 |
| port | string | 端口号 |
| properties | object | 扩展属性 |

### 数据流类型 (FlowPath)

| 字段名 | 类型 | 描述 |
|--------|------|------|
| id | string | 数据流唯一标识 |
| name | string | 数据流名称 |
| description | string | 数据流描述 |
| path | array | 节点ID序列 |
| color | string | 显示颜色 |

## 性能优化

### 缓存机制
- 拓扑数据默认缓存30秒
- 可通过刷新接口强制更新缓存

### 分页支持
- 支持分页查询，默认每页50条记录
- 大数据集建议使用分页参数

### 过滤优化
- 支持多种过滤条件组合
- 过滤操作在后端执行，减少数据传输

## 错误处理

### 错误码说明

| 错误码 | 描述 |
|--------|------|
| TOPOLOGY_FETCH_ERROR | 拓扑数据获取失败 |
| STATS_FETCH_ERROR | 统计信息获取失败 |
| REFRESH_ERROR | 缓存刷新失败 |
| EXPORT_ERROR | 数据导出失败 |
| INVALID_FORMAT | 不支持的导出格式 |

### 错误响应格式
```json
{
  "success": false,
  "error": {
    "message": "错误消息",
    "code": "ERROR_CODE",
    "details": "详细错误信息"
  }
}
```

## 使用示例

### 获取带过滤条件的拓扑数据
```bash
curl -X GET "http://localhost:8080/api/topology?protocol=tcp&chain=INPUT&include_stats=true" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 导出拓扑数据为JSON
```bash
curl -X GET "http://localhost:8080/api/topology/export?format=json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -o topology.json
```

### 强制刷新缓存
```bash
curl -X POST "http://localhost:8080/api/topology/refresh" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 版本历史

### v1.0 (当前版本)
- 基础拓扑数据获取
- 支持过滤和分页
- 统计信息接口
- 数据导出功能
- 健康检查接口
- 缓存机制

## 注意事项

1. **认证要求**: 所有API接口都需要有效的JWT令牌
2. **性能考虑**: 大数据集建议使用分页和过滤参数
3. **缓存策略**: 拓扑数据有30秒缓存，可通过刷新接口更新
4. **错误处理**: 建议实现重试机制和优雅的错误提示
5. **数据安全**: 导出的数据可能包含敏感信息，请妥善保管