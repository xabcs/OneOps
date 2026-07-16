# OneOps 事件通知系统设计方案

## 1. 事件分类体系

### 1.1 事件类型
```
系统事件 (system)
├── 用户登录 (user_login)
├── 用户登出 (user_logout)
├── 权限变更 (permission_change)
└── 系统配置修改 (config_change)

资源事件 (resource)
├── 服务器创建 (server_create)
├── 服务器删除 (server_delete)
├── K8s集群变更 (k8s_cluster_change)
├── 容器状态变更 (container_status_change)
└── 存储扩容 (storage_expand)

操作事件 (operation)
├── 任务执行完成 (task_complete)
├── 备份完成 (backup_complete)
├── 部署成功 (deploy_success)
└── 批量操作完成 (batch_operation_complete)

维护事件 (maintenance)
├── 维护计划通知 (maintenance_schedule)
├── 系统升级 (system_upgrade)
├── 功能变更 (feature_change)
└── 服务重启 (service_restart)

安全事件 (security)
├── 异常登录 (abnormal_login)
├── 权限异常 (permission_anomaly)
├── 安全漏洞扫描 (security_scan)
└── 合规检查 (compliance_check)
```

### 1.2 事件优先级
```go
const (
    PriorityCritical = "critical"  // 系统级、影响全局
    PriorityHigh     = "high"     // 业务中断、安全相关
    PriorityMedium   = "medium"   // 重要变更、需关注
    PriorityLow      = "low"      // 信息记录、可选查看
    PriorityInfo     = "info"     // 一般信息、日常操作
)
```

## 2. 数据库设计

### 2.1 系统事件表
```sql
CREATE TABLE system_events (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL COMMENT '事件类型',
    event_category VARCHAR(30) NOT NULL COMMENT '事件分类: system/resource/operation/maintenance/security',
    priority VARCHAR(20) NOT NULL DEFAULT 'medium' COMMENT '优先级: critical/high/medium/low/info',
    title VARCHAR(200) NOT NULL COMMENT '事件标题',
    description TEXT COMMENT '事件描述',
    metadata JSON COMMENT '事件元数据(JSON格式)',
    operator VARCHAR(100) COMMENT '操作人',
    operator_type VARCHAR(20) COMMENT '操作人类型: user/system/agent',
    target_type VARCHAR(50) COMMENT '目标类型: server/k8s_cluster/database',
    target_id VARCHAR(100) COMMENT '目标ID',
    source_ip VARCHAR(50) COMMENT '来源IP',
    status VARCHAR(20) DEFAULT 'unread' COMMENT '状态: unread/read/archived',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    read_at DATETIME COMMENT '阅读时间',
    expire_at DATETIME COMMENT '过期时间',

    INDEX idx_category_status (event_category, status),
    INDEX idx_priority_created (priority, created_at DESC),
    INDEX idx_operator (operator, created_at DESC),
    INDEX idx_target (target_type, target_id),
    INDEX idx_user_status (operator, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统事件表';
```

### 2.2 用户事件订阅表
```sql
CREATE TABLE user_event_subscriptions (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id INT UNSIGNED NOT NULL COMMENT '用户ID',
    event_type VARCHAR(50) COMMENT '事件类型(为空表示订阅分类下所有类型)',
    event_category VARCHAR(30) NOT NULL COMMENT '事件分类',
    notification_methods JSON COMMENT '通知方式: ["web","email","wechat","sms"]',
    priority_filter VARCHAR(50) COMMENT '优先级过滤: high+ 或 medium+',
    enabled BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    UNIQUE KEY uk_user_event (user_id, event_type, event_category),
    INDEX idx_user_enabled (user_id, enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户事件订阅表';
```

### 2.3 事件通知模板表
```sql
CREATE TABLE event_notification_templates (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL COMMENT '事件类型',
    template_name VARCHAR(100) NOT NULL COMMENT '模板名称',
    title_template VARCHAR(500) COMMENT '标题模板: 支持变量{{.Title}}',
    content_template TEXT COMMENT '内容模板: 支持Markdown',
    variables JSON COMMENT '可用变量列表',
    channel_type VARCHAR(20) NOT NULL COMMENT '渠道类型: web/email/wechat/sms',
    enabled BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    UNIQUE KEY uk_event_channel (event_type, channel_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='事件通知模板表';
```

## 3. 事件生命周期管理

### 3.1 事件状态流转
```
创建 (unread)
    ↓ 用户阅读
已读 (read)
    ↓ 用户操作
已处理 (processed)
    ↓ 7天后自动归档
归档 (archived)
```

### 3.2 事件过期策略
```go
// 事件过期时间配置
var EventExpiryRules = map[string]time.Duration{
    "critical": 30 * 24 * time.Hour,  // 30天
    "high":     14 * 24 * time.Hour,  // 14天
    "medium":   7  * 24 * time.Hour,  // 7天
    "low":      3  * 24 * time.Hour,  // 3天
    "info":     1  * 24 * time.Hour,  // 1天
}
```

## 4. API 设计

### 4.1 事件查询接口
```go
// GET /api/events/query
type EventQueryRequest struct {
    Category     string `json:"category"`      // 事件分类
    Type         string `json:"type"`          // 事件类型
    Priority     string `json:"priority"`      // 优先级
    Status       string `json:"status"`        // 状态
    Operator     string `json:"operator"`      // 操作人
    TargetType   string `json:"targetType"`    // 目标类型
    TargetID     string `json:"targetId"`      // 目标ID
    StartTime    string `json:"startTime"`     // 开始时间
    EndTime      string `json:"endTime"`       // 结束时间
    Page         int    `json:"page"`          // 页码
    PageSize     int    `json:"pageSize"`      // 每页数量
}

type EventQueryResponse struct {
    Total       int64          `json:"total"`
    UnreadCount int           `json:"unreadCount"`
    Events      []SystemEvent `json:"events"`
}
```

### 4.2 事件操作接口
```go
// POST /api/events/:id/read - 标记为已读
// POST /api/events/batch/read - 批量标记已读
// DELETE /api/events/:id - 删除事件
// POST /api/events/:id/archive - 归档事件
// GET /api/events/stats - 事件统计
// GET /api/events/unread-count - 未读事件数量
```

### 4.3 用户订阅接口
```go
// GET /api/events/subscriptions - 获取用户订阅
// POST /api/events/subscriptions - 创建订阅
// PUT /api/events/subscriptions/:id - 更新订阅
// DELETE /api/events/subscriptions/:id - 删除订阅
```

## 5. 前端组件设计

### 5.1 通知中心增强
```typescript
// types/event.ts
export interface SystemEvent {
  id: number;
  eventType: string;
  eventCategory: string;
  priority: 'critical' | 'high' | 'medium' | 'low' | 'info';
  title: string;
  description: string;
  metadata: Record<string, any>;
  operator: string;
  operatorType: 'user' | 'system' | 'agent';
  targetType?: string;
  targetId?: string;
  status: 'unread' | 'read' | 'archived';
  createdAt: string;
  readAt?: string;
}

export interface EventStats {
  total: number;
  unreadCount: number;
  byCategory: Record<string, number>;
  byPriority: Record<string, number>;
}
```

### 5.2 事件筛选器
```vue
<EventCenter>
  <EventFilter>
    <CategorySelector v-model="filter.category" />
    <PrioritySelector v-model="filter.priority" />
    <TimeRangePicker v-model="filter.timeRange" />
    <StatusSelector v-model="filter.status" />
  </EventFilter>

  <EventList :events="events" :loading="loading">
    <EventItem
      v-for="event in events"
      :key="event.id"
      :event="event"
      @read="handleReadEvent"
      @delete="handleDeleteEvent"
    />
  </EventList>

  <EventPagination
    v-model:page="pagination.page"
    v-model:pageSize="pagination.pageSize"
    :total="total"
  />
</EventCenter>
```

## 6. 事件发布机制

### 6.1 事件发布接口
```go
// service/event_publisher.go
type EventPublisher interface {
    // 发布事件
    PublishEvent(event *SystemEvent) error

    // 批量发布事件
    PublishEvents(events []*SystemEvent) error

    // 发布带目标的特定事件
    PublishTargetEvent(targetType, targetID, eventType string, metadata map[string]interface{}) error
}

// 使用示例
eventPublisher.PublishEvent(&SystemEvent{
    EventType:     "server_create",
    EventCategory: "resource",
    Priority:      "medium",
    Title:         "服务器创建成功",
    Description:   fmt.Sprintf("服务器 %s (%s) 创建成功", server.Name, server.IP),
    Metadata:      json.Marshal(server),
    Operator:      currentUser.Username,
    OperatorType:  "user",
    TargetType:    "server",
    TargetID:      fmt.Sprintf("%d", server.ID),
})
```

### 6.2 事件监听和处理
```go
// service/event_handler.go
type EventHandler interface {
    // 处理事件，决定是否通知用户
    HandleEvent(event *SystemEvent) error

    // 获取应该通知的用户列表
    GetSubscribers(event *SystemEvent) ([]int, error)

    // 发送通知
    SendNotification(userID int, event *SystemEvent, methods []string) error
}
```

## 7. 实时推送集成

### 7.1 WebSocket 事件推送
```go
// 通过现有 WebSocket Hub 推送新事件
func (h *EventService) PublishEvent(event *SystemEvent) error {
    // 1. 保存到数据库
    if err := h.db.Create(event).Error; err != nil {
        return err
    }

    // 2. 获取订阅用户
    subscribers, _ := h.handler.GetSubscribers(event)

    // 3. 通过 WebSocket 推送
    wsMessage := &WebSocketMessage{
        Type: "system_event",
        Data: event,
        Timestamp: time.Now().Format(time.RFC3339),
    }

    for _, userID := range subscribers {
        hub.SendToClient(fmt.Sprintf("user_%d", userID), wsMessage)
    }

    return nil
}
```

### 7.2 前端 WebSocket 监听
```typescript
// service/websocket.ts
websocket.addEventListener('message', (event) => {
  const message = JSON.parse(event.data);

  if (message.type === 'system_event') {
    // 显示新事件通知
    showEventNotification(message.data);

    // 更新未读数量
    updateUnreadCount();

    // 刷新事件列表（如果在事件页面）
    if (isEventPage()) {
      refreshEventList();
    }
  }
});
```

## 8. 与告警系统集成

### 8.1 告警转事件
```go
// 当告警确认时自动生成事件
func (s *AlertService) AcknowledgeAlert(alertID uint, comment string) error {
    alert, err := s.GetAlert(alertID)
    if err != nil {
        return err
    }

    // 更新告警状态
    alert.Acknowledged = true
    alert.AcknowledgedAt = time.Now()
    alert.AcknowledgedBy = currentUser

    // 生成对应的系统事件
    eventPublisher.PublishEvent(&SystemEvent{
        EventType:     "alert_acknowledged",
        EventCategory: "operation",
        Priority:      "medium",
        Title:         "告警已确认",
        Description:   fmt.Sprintf("告警 '%s' 已被 %s 确认", alert.Message, currentUser.Username),
        Metadata:      map[string]interface{}{"alert_id": alertID, "comment": comment},
        Operator:      currentUser.Username,
        OperatorType:  "user",
    })

    return s.db.Save(alert).Error
}
```

## 9. 实施步骤

### 阶段1: 数据库和后端API (Week 1-2)
1. 创建数据库表结构
2. 实现事件发布服务
3. 实现事件查询API
4. 实现用户订阅管理

### 阶段2: 前端界面 (Week 3)
1. 增强通知中心组件
2. 实现事件列表页面
3. 实现事件筛选和搜索
4. 实现用户订阅配置

### 阶段3: 实时推送 (Week 4)
1. 集成WebSocket事件推送
2. 实现前端实时通知
3. 优化推送性能和用户体验

### 阶段4: 集成和测试 (Week 5)
1. 与现有系统集成
2. 事件发布点覆盖
3. 性能测试和优化
4. 用户验收测试

## 10. 关键设计决策

### 10.1 为什么分离事件和告警？
- **职责分离**: 告警关注异常，事件关注变更
- **用户场景不同**: 告警需要快速响应，事件用于审计和追溯
- **存储策略不同**: 告警短期存储，事件长期归档

### 10.2 事件粒度控制
- 避免过度细粒度导致的"事件疲劳"
- 批量操作生成汇总事件
- 提供事件聚合配置

### 10.3 性能考虑
- 事件表按时间分区
- 定期归档过期事件
- 使用Redis缓存未读数量
- WebSocket推送限流保护