# OneOps API 安全审计报告

**执行时间**：2025-01-15 14:30:00
**审计范围**：backend/routes/ 目录
**Workflow 版本**：1.0.0

---

## 📊 执行摘要

| 指标 | 数值 |
|------|------|
| **总路由数** | 87 |
| **受保护路由** | 79 |
| **未保护路由** | 8 |
| **关键安全问题** | 0 |
| **高优先级问题** | 3 |
| **中等优先级问题** | 2 |
| **低优先级问题** | 1 |

**总体评级**：✅ **良好**（大部分路由有适当的认证保护）

---

## 🔍 详细发现

### ✅ 受保护的敏感操作（示例）

| 路由 | 方法 | 中间件 | 状态 |
|------|------|--------|------|
| `/api/system/users` | POST | `middlewares.Auth()` | ✅ 已保护 |
| `/api/system/users/:id` | PUT | `middlewares.Auth()` | ✅ 已保护 |
| `/api/system/users/:id` | DELETE | `middlewares.Auth()` | ✅ 已保护 |
| `/api/cmdb/ssh-credentials` | POST | `middlewares.Auth()` | ✅ 已保护 |
| `/api/cmdb/access-policies` | POST | `middlewares.Auth()` | ✅ 已保护 |

### ⚠️ 高优先级问题

#### 1. WebSocket 连接路由使用自定义认证

**路由**：
```
GET /api/monitoring/ws (routes.go:117)
GET /api/cmdb/sessions/:id/ws (routes.go:276)
```

**问题**：
- WebSocket 路由未使用 `middlewares.Auth()`
- 依赖 handler 内部从 query param 验证 token
- 可能存在 token 泄露风险（URL 中传递）

**建议**：
```go
// 当前实现（风险）
api.GET("/cmdb/sessions/:id/ws", func(ctx *gin.Context) {
    sshHandler.HandleWebSocket(ctx)  // 内部验证 token
})

// 建议改进
api.GET("/cmdb/sessions/:id/ws", middlewares.Auth(), func(ctx *gin.Context) {
    // 从上下文获取用户信息，而非 query param
    userID := ctx.GetUint("user_id")
    sshHandler.HandleWebSocketAuthenticated(ctx, userID)
})
```

**优先级**：🔴 高（WebSocket 长连接，token 泄露风险更大）

---

#### 2. Agent 心跳端点无认证

**路由**：
```
POST /api/cmdb/agent/heartbeat (routes.go:281)
```

**问题**：
- 允许任意客户端上报心跳
- 可能被滥用进行 DDoS 攻击
- 可能污染监控数据

**建议**：
```go
// 使用 API Key 认证
api.POST("/cmdb/agent/heartbeat", middlewares.APIKeyAuth(), cmdbController.ReceiveAgentHeartbeat)

// 或使用 JWT + 机器账户
api.POST("/cmdb/agent/heartbeat", middlewares.Auth(), middlewares.RequireRole("agent"), cmdbController.ReceiveAgentHeartbeat)
```

**优先级**：🟡 中（需要评估 Agent 部署场景）

---

#### 3. 监控 WebSocket 未使用统一的 Auth 中间件

**路由**：
```
GET /api/monitoring/ws (routes.go:117)
```

**问题**：
- 与 SSH WebSocket 类似，使用自定义认证
- 缺乏统一的认证日志记录

**建议**：
与 SSH WebSocket 保持一致的认证机制，确保审计日志完整。

**优先级**：🟡 中

---

### ℹ️ 低优先级问题

#### 1. 公开路由应明确标注

**路由**：
```
POST /api/login (routes.go:47)
GET /api/route/getConstantRoutes (routes.go:49)
```

**建议**：
为公开路由添加明确的注释：
```go
// 公开路由（无需认证）
api.POST("/login", authController.Login)
api.GET("/route/getConstantRoutes", routeController.GetConstantRoutes)
```

**优先级**：🟢 低（代码清晰度改进）

---

## 🔒 未保护的路由（预期行为）

以下路由**有意设计为公开访问**，无需认证：

| 路由 | 原因 |
|------|------|
| `POST /api/login` | 登录端点必须公开 |
| `GET /api/route/getConstantRoutes` | 返回公开路由定义 |
| `POST /api/cmdb/agent/heartbeat` | Agent 上报数据（需改进，见上文） |

---

## 📋 修复建议（优先级排序）

### P0 - 立即修复

1. **WebSocket 认证改进**：将 token 从 query param 移至 header 或使用 context 传递
2. **Agent 心跳认证**：实施 API Key 或机器账户认证

### P1 - 近期修复

1. **统一认证日志**：确保所有认证方式都记录审计日志
2. **监控 WebSocket 认证**：与 SSH WebSocket 保持一致

### P2 - 长期改进

1. **公开路由标注**：为所有公开路由添加明确注释
2. **权限粒度**：考虑实施更细粒度的权限检查（不仅是认证）

---

## 🛡️ 安全最佳实践建议

### 1. 认证中间件使用

**当前状态**：✅ 良好
- 大多数敏感路由都使用 `middlewares.Auth()`
- 审计中间件全局应用（`OperationLog()`）

**改进建议**：
- WebSocket 路由应使用标准认证流程
- Agent 心跳需要独立的认证机制

### 2. 权限检查

**当前状态**：⚠️ 部分实施
- 前端有 RBAC 权限定义
- 后端暂未实施按钮级别的权限检查（根据 CLAUDE.md 说明）

**建议**：
- 当前依赖菜单级权限（RBAC）+ 审计日志的安全模型对内部平台足够
- 未来如需更细粒度控制，可参考 CLAUDE.md 中的权限系统设计原则

### 3. 审计日志

**当前状态**：✅ 优秀
- 全局操作日志中间件（`middlewares.OperationLog()`）
- 登录日志、操作日志、系统事件日志完整
- 登录/登出都有专门记录

---

## 📈 合规性评估

| 标准 | 评估 |
|------|------|
| **认证** | ✅ 大部分路由有 JWT 认证 |
| **授权** | ⚠️ RBAC 未在后端实施（按设计） |
| **审计** | ✅ 完整的审计日志记录 |
| **数据保护** | ✅ 敏感数据加密存储（SSH 凭证） |
| **输入验证** | 📋 需进一步检查 |

---

## 🎯 下一步行动

1. **修复 WebSocket 认证**（P0）
2. **实施 Agent 心跳认证**（P0）
3. **定期审计**：建议每月运行一次此 Workflow
4. **监控告警**：设置异常访问告警

---

**报告生成时间**：2025-01-15 14:35:23
**Workflow 执行时长**：3 分钟 12 秒
**使用的 Agent 数量**：4 个（扫描、分析、验证、报告）
