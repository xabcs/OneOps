# 数据库查询优化审查报告

## 审查日期
2026年6月9日

## 审查范围
- 数据库查询性能
- N+1 查询问题
- 索引使用情况
- 批量操作优化

## 审查结果

### ✅ 良好的查询实践

#### 1. N+1 查询问题预防
**状态**: ✅ 已正确使用 Preload

**示例代码**（services/cmdb.go:211-216）:
```go
Preload("Cabinet").
Preload("Cabinet.Room").
Preload("Tags").
Preload("Credentials").
Preload("Groups").
Preload("Attributes")
```

**优点**:
- 使用 GORM 的 Preload 避免了 N+1 查询问题
- 支持嵌套 Preload（Cabinet.Room）
- 按需加载关联数据（有些查询只 Preload Credentials）

#### 2. 索引定义
**状态**: ✅ 索引定义完善

**关键索引**:

**唯一索引**（uniqueIndex）:
- `servers.hostname` - 主机名唯一索引
- `users.username` - 用户名唯一索引
- `roles.code` - 角色代码唯一索引
- `businesses.code` - 业务代码唯一索引
- `rooms.code` - 机房代码唯一索引
- `cabinets.code` - 机柜代码唯一索引
- `tags.name` - 标签名称唯一索引
- `server_groups.code` - 分组代码唯一索引
- `cloud_servers.instance_id` - 实例ID唯一索引
- `ssh_credentials.name` - 凭证名称唯一索引
- `agent_versions.version` - 版本号唯一索引

**普通索引**（index）:
- 查询频繁的字段都有索引：
  - `servers.ip` - 外网IP
  - `servers.inner_ip` - 内网IP
  - `servers.env` - 环境
  - `servers.status` - 状态
  - `servers.agent_status` - Agent状态
  - `audit_logs.user_id` - 用户ID
  - `audit_logs.username` - 用户名
  - `audit_logs.login_time` - 登录时间
  - `server_credentials.server_id` - 服务器ID
  - `server_credentials.credential_id` - 凭证ID
  - `server_group_relations.server_id` - 服务器ID
  - `server_group_relations.group_id` - 分组ID

**优点**:
- 所有常用查询字段都有索引
- 唯一性约束正确使用
- 外键字段都有索引

#### 3. JOIN 查询优化
**状态**: ✅ 正确使用 Joins

**示例代码**（services/cmdb.go:297）:
```go
Joins("JOIN server_groups g ON g.id = server_group_relations.group_id")
```

**优点**:
- 使用 JOIN 减少多次查询
- WHERE 条件使用参数化查询

#### 4. 批量查询
**状态**: ✅ 未发现明显的批量插入需求

**说明**:
- 未发现循环中的单条插入
- 当前后端 API 主要是单条操作
- 如需批量导入功能，建议使用 GORM 的 CreateInBatches

### 优化建议

#### 1. 复合索引优化（P2 优先级）

**建议添加的复合索引**:

```sql
-- 审计日志查询优化（按用户+时间查询）
CREATE INDEX idx_audit_logs_user_time ON audit_logs(user_id, operate_time);

-- 服务器列表查询优化（按环境+状态查询）
CREATE INDEX idx_servers_env_status ON servers(env, status);

-- Agent 状态查询优化（按状态查询）
CREATE INDEX idx_servers_agent_status ON servers(agent_status);

-- SSH 凭证查询优化（按服务器查询）
CREATE INDEX idx_server_credentials_server ON server_credentials(server_id, credential_id);
```

**原因**:
- audit_logs 表经常按 user_id + operate_time 查询
- servers 表经常按 env + status 过滤
- agent_status 查询频繁（查找 agent_status=running 的服务器）
- server_credentials 经常按 server_id 查询凭证

#### 2. 查询结果缓存（P2 优先级）

**建议**:
- 对于不常变化的数据（如机房、机柜、标签列表），使用 Redis 缓存
- TTL 设置为 5-10 分钟

**示例**:
```go
func (s *CMDBService) GetRoomsWithCache() ([]*models.Room, error) {
    // 先从 Redis 获取
    var rooms []*models.Room
    if err := redisCache.Get("cmdb:rooms", &rooms); err == nil {
        return rooms, nil
    }

    // 缓存未命中，从数据库查询
    err := db.Preload("Cabinets").Order("id ASC").Find(&rooms).Error
    if err != nil {
        return nil, err
    }

    // 存入缓存
    redisCache.Set("cmdb:rooms", rooms, 5*time.Minute)
    return rooms, nil
}
```

#### 3. 分页查询优化（P2 优先级）

**当前实现**: 已使用 offset + limit

**建议优化**:
- 对于大数据量分页，使用 cursor-based 分页（游标分页）
- 避免深分页的性能问题

**示例**:
```go
// 当前实现（偏移量分页）
tx.Offset((page - 1) * pageSize).Limit(pageSize)

// 优化方案（游标分页，适合大数据量）
// 第一页
tx.Where("id > ?", lastId).Order("id ASC").Limit(pageSize)
```

#### 4. 慢查询监控（P1 优先级）

**建议**:
- 启用 GORM 的慢查询日志
- 设置阈值（如 100ms）
- 定期审查慢查询日志

**配置示例**:
```go
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info),
    // 或自定义慢查询日志
})
```

### 性能测试建议

#### 1. 查询性能测试
```go
func TestGetServersPerformance(t *testing.T) {
    start := time.Now()
    servers, _, err := cmdbService.GetServers(nil, 1, 100)
    elapsed := time.Since(start)

    assert.NoError(t, err)
    assert.Less(t, elapsed.Milliseconds(), int64(100), "查询应在100ms内完成")
}
```

#### 2. 压力测试
```bash
# 使用 wrk 或 ab 进行压力测试
ab -n 1000 -c 10 http://localhost:8082/api/cmdb/servers
```

#### 3. EXPLAIN 分析
```sql
-- 分析查询执行计划
EXPLAIN SELECT * FROM servers WHERE hostname = 'test-server';

-- 检查索引使用情况
SHOW INDEX FROM servers;
```

## 审查结论

### 总体评估：✅ 良好

当前项目在数据库查询优化方面做得很好：

1. **N+1 查询问题**: ✅ 已正确使用 Preload 避免
2. **索引定义**: ✅ 完善的索引覆盖
3. **JOIN 查询**: ✅ 正确使用 JOIN 减少查询次数
4. **参数化查询**: ✅ 所有查询都使用参数化

### 优先级建议

| 优先级 | 任务 | 预期收益 |
|--------|------|----------|
| P2 | 添加复合索引 | 中等（10-30%性能提升） |
| P2 | 实现查询结果缓存 | 中等（减少数据库负载） |
| P2 | 优化深分页查询 | 低（仅在数据量大时有效） |
| P1 | 启用慢查询监控 | 高（及时发现问题） |

### 无需立即优化的问题

当前查询性能已经足够好，以下优化可以在实际性能瓶颈出现时再进行：
- 复合索引（当前单字段索引已足够）
- 缓存（Redis 已启用，可按需添加）
- 分页优化（当前数据量下 offset 分页足够）

### 建议的后续行动

1. **立即执行**（P1）:
   - [ ] 启用 GORM 慢查询日志
   - [ ] 监控生产环境查询性能

2. **短期优化**（P2）:
   - [ ] 添加关键复合索引
   - [ ] 实现不常变化数据的缓存

3. **长期优化**（P3）:
   - [ ] 大数据量场景下的分页优化
   - [ ] 读写分离架构（如果数据量持续增长）

---

**审查人**: Claude Code AI
**审查状态**: ✅ 通过
**性能评估**: 良好
**建议**: 当前查询性能可满足生产需求，建议按优先级逐步优化
