# 调试指南

## 问题描述
页面显示无数据，但执行iptables命令有数据。

## 调试步骤

### 1. 运行调试脚本
```bash
./debug.sh
```

### 2. 检查后端日志
查看详细的调试日志：
```bash
docker-compose logs -f backend
```

### 3. 手动测试API
```bash
# 登录获取token
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# 使用token获取规则 (替换YOUR_TOKEN)
curl -X GET http://localhost:8080/api/rules \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 4. 导入示例规则
在前端页面点击"导入规则"按钮，或者使用API：
```bash
curl -X POST http://localhost:8080/api/rules/import \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 5. 检查数据库
```bash
# 进入后端容器
docker-compose exec backend sh

# 查看数据库文件
ls -la /app/data/

# 如果有sqlite3命令，可以直接查询
sqlite3 /app/data/iptables.db "SELECT * FROM iptables_rules;"
```

## 调试日志说明

### 后端日志关键信息
- `[DEBUG] GetRules API called` - API被调用
- `[DEBUG] RuleService.GetAllRules called` - 服务层被调用
- `[DEBUG] Database query successful, found X rules` - 数据库查询结果
- `[DEBUG] Returning X rules to client` - 返回给客户端的数据

### 前端日志关键信息
- `[DEBUG] loadRules called` - 前端开始加载规则
- `[DEBUG] API Request: GET /api/rules` - API请求发送
- `[DEBUG] API Response: GET /api/rules Status: 200 Data length: X` - API响应
- `[DEBUG] Rules assigned to reactive variable` - 数据赋值给响应式变量

## 常见问题解决

### 1. 数据库为空
- 点击前端"导入规则"按钮
- 或者重启服务让默认数据初始化

### 2. API认证失败
- 检查token是否正确
- 重新登录获取新token

### 3. 前端显示问题
- 检查浏览器控制台错误
- 检查网络请求是否成功

### 4. 后端连接问题
- 检查Docker容器是否正常运行
- 检查端口映射是否正确