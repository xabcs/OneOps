# 操作审计功能设计方案

## 一、当前问题分析

### 1.1 存在的问题

1. **Module 字段都是"未知模块"**
   - 原因：尝试将 API 路径映射到 Menu 表，但路径不匹配
   - API 路径：`/api/monitoring/stats`
   - Menu 路径：`/monitoring`
   - 无法对应

2. **设计缺陷**
   - 依赖 Menu 表来确定模块，但 Menu 存储的是前端路由
   - 新增 API 需要手动配置 Menu 映射
   - 学习机制在首次访问时无效

## 二、最佳实践设计

### 2.1 核心原则

1. **独立性**：审计日志不应依赖 Menu 表
2. **自动化**：自动识别模块和操作，无需手动配置
3. **完整性**：记录所有关键信息（5W1H）
4. **性能**：异步记录，不影响业务性能

### 2.2 应该记录的信息

| 字段 | 说明 | 示例 |
|------|------|------|
| 用户信息 | 谁 | admin (ID: 1) |
| 操作时间 | 何时 | 2024-04-23 10:30:00 |
| 模块 | 哪个功能 | 监控中心 |
| 操作 | 做了什么 | 访问、查询、刷新、处理告警 |
| 描述 | 详细说明 | 处理了来自 Web-Server-01 的告警 |
| 方法 | HTTP方法 | POST /api/monitoring/alert/handle |
| IP地址 | 从哪里操作 | 192.168.1.100 |
| 耗时 | 性能指标 | 125ms |
| 状态 | 成功/失败 | success |
| 错误信息 | 失败原因 | 权限不足 |

### 2.3 Module（模块）识别规则

```
基于 API 路径前缀自动识别：

/api/login*          → 认证管理
/api/user*           → 用户管理
/api/system/menus*   → 菜单管理
/api/system/roles*   → 角色管理
/api/system/users*   → 用户管理
/api/audit*          → 审计管理
/api/monitoring*     → 监控中心
/api/tasks*          → 任务管理
/api/servers*        → 服务器管理
/api/containers*     → 容器管理
/api/certificates*   → 证书管理
/api/*               → 其他
```

### 2.4 Action（操作）识别规则

```
基于 HTTP 方法和路径：

GET    /api/monitoring/stats  → 查询统计数据
POST   /api/monitoring/refresh → 刷新数据
POST   /api/monitoring/alert/handle → 处理告警
DELETE /api/system/users/1 → 删除用户
PUT    /api/system/roles/2 → 更新角色
POST   /api/system/menus → 创建菜单
```

## 三、修复方案

### 3.1 重写 getModuleFromPath 方法

```go
// 从路径智能识别模块
func (m *AuditMiddleware) getModuleFromPath(path string) string {
    if !strings.HasPrefix(path, "/api/") {
        return "其他"
    }

    apiPath := strings.TrimPrefix(path, "/api")

    // 基于 API 路径前缀识别
    moduleMap := map[string]string{
        "/login":         "认证管理",
        "/logout":        "认证管理",
        "/user":          "用户管理",
        "/system/menus":  "菜单管理",
        "/system/roles":  "角色管理",
        "/system/users":  "用户管理",
        "/audit":         "审计管理",
        "/monitoring":    "监控中心",
        "/tasks":         "任务管理",
        "/servers":       "服务器管理",
        "/containers":    "容器管理",
        "/certificates":  "证书管理",
    }

    for prefix, module := range moduleMap {
        if strings.HasPrefix(apiPath, prefix) {
            return module
        }
    }

    // 从路径提取第一级作为模块名
    parts := strings.Split(strings.Trim(apiPath, "/"), "/")
    if len(parts) > 0 {
        return parts[0]
    }

    return "其他"
}
```

### 3.2 改进 getActionFromMethod 方法

```go
// 从方法和路径识别操作
func (m *AuditMiddleware) getActionFromMethodAndPath(method, path string) string {
    // 特殊路径优先处理
    if strings.Contains(path, "/refresh") {
        return "刷新"
    }
    if strings.Contains(path, "/handle") {
        return "处理"
    }
    if strings.Contains(path, "/ignore") {
        return "忽略"
    }
    if strings.Contains(path, "/export") {
        return "导出"
    }
    if strings.Contains(path, "/visit") {
        return "访问"
    }

    // 基于 HTTP 方法
    switch method {
    case "GET":
        return "查询"
    case "POST":
        // 判断是创建还是其他操作
        if strings.HasSuffix(path, "/login") {
            return "登录"
        }
        if strings.HasSuffix(path, "/logout") {
            return "登出"
        }
        return "新增"
    case "PUT":
        return "更新"
    case "DELETE":
        return "删除"
    case "PATCH":
        return "修改"
    default:
        return method
    }
}
```

### 3.3 生成更友好的描述

```go
// 生成操作描述
func (m *AuditMiddleware) generateDescription(module, action, path string) string {
    // 特殊处理
    if action == "刷新" {
        return fmt.Sprintf("刷新%s数据", module)
    }
    if action == "访问" {
        return fmt.Sprintf("访问%s", module)
    }
    if action == "处理" && strings.Contains(path, "alert") {
        return "处理告警"
    }

    // 通用格式
    return fmt.Sprintf("%s%s", action, module)
}
```

## 四、实施步骤

### 4.1 修改审计中间件

1. 重写 `getModuleFromPath` - 使用路径前缀映射
2. 改进 `getActionFromMethod` - 考虑路径语义
3. 添加 `generateDescription` - 生成友好描述

### 4.2 清理历史数据

```sql
-- 更新历史数据中的 module 字段
UPDATE operation_logs
SET module = CASE
    WHEN path LIKE '/api/monitoring%' THEN '监控中心'
    WHEN path LIKE '/api/audit%' THEN '审计管理'
    WHEN path LIKE '/api/system/menus%' THEN '菜单管理'
    WHEN path LIKE '/api/system/roles%' THEN '角色管理'
    WHEN path LIKE '/api/system/users%' THEN '用户管理'
    WHEN path LIKE '/api/user%' THEN '用户管理'
    WHEN path LIKE '/api/login%' THEN '认证管理'
    WHEN path LIKE '/api/logout%' THEN '认证管理'
    ELSE '其他'
END
WHERE module = '未知模块';
```

### 4.3 验证和测试

1. 访问监控中心 → 应记录为"监控中心-访问"
2. 刷新监控数据 → 应记录为"监控中心-刷新"
3. 处理告警 → 应记录为"监控中心-处理告警"
4. 查询操作日志 → 应记录为"审计管理-查询"

## 五、未来优化建议

### 5.1 性能优化

- 使用异步队列记录审计日志
- 批量写入，减少数据库压力

### 5.2 数据保留

- 定期归档旧数据
- 设置保留期限（如 6 个月）

### 5.3 数据分析

- 统计用户活跃度
- 分析高频操作
- 发现异常行为

### 5.4 告警集成

- 敏感操作告警
- 异常访问告警
- 批量操作告警

## 六、参考示例

### 6.1 完整的审计记录示例

```json
{
  "id": 1234,
  "userId": 1,
  "username": "admin",
  "nickname": "系统管理员",
  "module": "监控中心",
  "action": "处理",
  "description": "处理告警",
  "method": "POST",
  "path": "/api/monitoring/alert/handle",
  "params": {
    "alertId": 1,
    "action": "handle",
    "reason": "手动处理"
  },
  "response": {
    "code": 200,
    "message": "告警处理成功"
  },
  "statusCode": 200,
  "ip": "192.168.1.100",
  "userAgent": "Mozilla/5.0...",
  "duration": 125,
  "status": "success",
  "errorMsg": "",
  "operateTime": "2024-04-23 10:30:00"
}
```

## 七、常见问题

### Q1: 为什么不使用 Menu 表？
A: Menu 表存储的是前端路由，而审计记录的是 API 调用，两者路径不一致。使用 API 路径直接映射更准确。

### Q2: 如何处理新的 API？
A: 新的 API 会自动根据路径前缀识别模块，无需额外配置。

### Q3: Module 是中文还是英文？
A: 建议使用中文，方便用户阅读和理解。

### Q4: 如何区分同名操作？
A: 通过组合 Module + Action + Path，例如"监控中心-刷新"和"审计管理-刷新"。
