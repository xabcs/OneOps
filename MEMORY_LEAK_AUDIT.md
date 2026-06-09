# 内存泄漏审查报告

## 审查日期
2026年6月9日

## 审查范围
- Goroutine 泄漏
- 连接泄漏（数据库、Redis、HTTP）
- 文件句柄泄漏
- WebSocket 连接泄漏

## 审查方法
1. 静态代码审查
2. 检查 goroutine 生命周期管理
3. 检查资源清理逻辑

## 审查结果

### 🔴 发现的问题

#### 问题1: Agent 指标采集调度器 Goroutine 泄漏（P1 严重）

**位置**: `services/agent.go:359-382`

**问题代码**:
```go
// 启动心跳超时检测（每1分钟）
heartbeatTicker := time.NewTicker(1 * time.Minute)
go func() {
    for range heartbeatTicker.C {
        checkHeartbeatTimeout()
    }
}()

// 快速采集：性能指标（每30秒）
fastMetricsTicker := time.NewTicker(30 * time.Second)
go func() {
    for range fastMetricsTicker.C {
        collectFastMetrics(monitoringService)
    }
}()

// 常规采集：基础指标 + 系统信息（每5分钟）
regularMetricsTicker := time.NewTicker(5 * time.Minute)
go func() {
    for range regularMetricsTicker.C {
        collectRegularMetrics(monitoringService)
    }
}()
```

**问题分析**:
1. ❌ Goroutine 没有退出机制
2. ❌ Ticker 没有调用 Stop()
3. ❌ 如果函数被多次调用，会创建多个 goroutine
4. ❌ 没有使用 context.Context 进行取消控制

**影响**:
- Goroutine 泄漏
- 内存持续增长
- CPU 资源浪费

**修复方案**:

```go
// 添加全局 context 控制
var agentSchedulerCtx, agentSchedulerCancel = context.WithCancel(context.Background())

// StartAgentMetricsScheduler 启动 Agent 指标采集调度器
func StartAgentMetricsScheduler() {
    monitoringService := NewMonitoringService()

    // 启动心跳超时检测（每1分钟）
    heartbeatTicker := time.NewTicker(1 * time.Minute)
    go func() {
        defer heartbeatTicker.Stop()
        for {
            select {
            case <-heartbeatTicker.C:
                checkHeartbeatTimeout()
            case <-agentSchedulerCtx.Done():
                logger.Info("心跳超时检测 goroutine 退出")
                return
            }
        }
    }()

    // 快速采集：性能指标（每30秒）
    fastMetricsTicker := time.NewTicker(30 * time.Second)
    go func() {
        defer fastMetricsTicker.Stop()
        for {
            select {
            case <-fastMetricsTicker.C:
                collectFastMetrics(monitoringService)
            case <-agentSchedulerCtx.Done():
                logger.Info("快速指标采集 goroutine 退出")
                return
            }
        }
    }()

    // 常规采集：基础指标 + 系统信息（每5分钟）
    regularMetricsTicker := time.NewTicker(5 * time.Minute)
    go func() {
        defer regularMetricsTicker.Stop()
        for {
            select {
            case <-regularMetricsTicker.C:
                collectRegularMetrics(monitoringService)
            case <-agentSchedulerCtx.Done():
                logger.Info("常规指标采集 goroutine 退出")
                return
            }
        }
    }()

    // 立即执行一次采集
    go collectFastMetrics(monitoringService)
    go collectRegularMetrics(monitoringService)
}

// StopAgentMetricsScheduler 停止调度器（在服务关闭时调用）
func StopAgentMetricsScheduler() {
    if agentSchedulerCancel != nil {
        agentSchedulerCancel()
    }
}
```

**优先级**: 🔴 P1（严重，必须修复）

#### 问题2: 并发采集中的 Goroutine 泄漏风险（P2 中等）

**位置**: `services/agent.go:412-540`

**问题代码**:
```go
for _, sid := range serverIDs {
    sem <- struct{}{}
    wg.Add(1)
    go func(sid uint) {
        defer wg.Done()
        defer func() { <-sem }()

        // 采集逻辑...
    }(sid)
}
wg.Wait()
```

**问题分析**:
1. ✅ 使用 WaitGroup 等待所有 goroutine 完成
2. ✅ 使用 semaphore 限制并发数
3. ⚠️ 如果采集逻辑 panic，可能会导致 sem 泄漏
4. ⚠️ 没有超时控制

**建议优化**:

```go
for _, sid := range serverIDs {
    sem <- struct{}{}
    wg.Add(1)

    go func(sid uint) {
        defer wg.Done()
        defer func() { <-sem }()

        // 添加 panic 恢复
        defer func() {
            if r := recover(); r != nil {
                logger.Error("Agent指标采集panic",
                    zap.Uint("server_id", sid),
                    zap.Any("panic", r))
            }
        }()

        // 添加超时控制
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        if err := collectServerMetrics(ctx, sid); err != nil {
            logger.Warn("采集失败",
                zap.Uint("server_id", sid),
                zap.Error(err))
        }
    }(sid)
}
wg.Wait()
```

**优先级**: 🟡 P2（建议优化）

### ✅ 良好的实现

#### 1. WebSocket 连接管理
**状态**: ✅ 正确实现

**位置**: `handlers/websocket.go:249-276`

**优点**:
- ✅ 使用 context.Context 控制 goroutine 退出
- ✅ 使用 WaitGroup 等待 goroutine 完成
- ✅ 正确清理资源（sshSession.Close()）
- ✅ 双向转发都有退出机制

**示例代码**:
```go
// WebSocket -> SSH
go func() {
    defer wg.Done()
    h.forwardWebSocketToSSH(conn, stdinPipe, session, sshCtx)

    // 退出后清理资源
    if sshSession != nil {
        sshSession.Close()
    }
    cancel()
}()

// SSH -> WebSocket
go func() {
    defer wg.Done()
    h.forwardSSHToWebSocket(stdoutPipe, conn, session, sshCtx)
}()

// 等待完成
wg.Wait()
```

#### 2. 数据库连接管理
**状态**: ✅ 正确实现

**说明**:
- GORM 自动管理连接池
- 使用 `defer sqlDB.Close()` 确保关闭
- 连接池配置合理（MaxIdleConns: 10, MaxOpenConns: 100）

**位置**: `services/db.go:44-52`

```go
sqlDB2, err := db.DB()
if err != nil {
    return err
}

// 设置连接池
sqlDB2.SetMaxIdleConns(10)
sqlDB2.SetMaxOpenConns(100)
sqlDB2.SetConnMaxLifetime(time.Hour)
```

#### 3. Redis 连接管理
**状态**: ✅ 正确实现

**说明**:
- go-redis 客户端是并发安全的
- 服务关闭时调用 `CloseRedis()`
- 使用连接池管理连接

**位置**: `main.go:103`

```go
defer services.CloseRedis()
```

#### 4. HTTP 客户端管理
**状态**: ✅ 正确实现

**说明**:
- 使用 http.Client 的连接池
- 设置超时避免连接泄漏

**位置**: `services/notification/wechat.go:31-39`

```go
httpClient := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: true,
        },
    },
}
```

### 检查清单

#### Goroutine 泄漏检查
- [x] 所有 goroutine 都有退出机制
- [ ] ❌ Agent 调度器 goroutine 缺少退出机制
- [x] 使用 context.Context 控制生命周期
- [x] 使用 defer 确保资源清理

#### 连接泄漏检查
- [x] 数据库连接使用连接池
- [x] Redis 连接正确关闭
- [x] HTTP 客户端设置超时
- [x] WebSocket 连接正确管理

#### 资源清理检查
- [x] 文件使用 defer 关闭
- [x] Ticker 使用 defer Stop()
- [ ] ❌ 部分 Ticker 未调用 Stop()

## 修复优先级

### P1（立即修复）
1. **修复 Agent 调度器 goroutine 泄漏**
   - 添加 context 控制
   - 确保所有 ticker 调用 Stop()
   - 添加 StopAgentMetricsScheduler() 函数

### P2（建议优化）
1. **优化并发采集的 panic 恢复**
   - 添加 recover() 机制
   - 添加超时控制

2. **添加资源监控**
   - 监控 goroutine 数量
   - 监控内存使用

### P3（长期改进）
1. **添加 pprof 支持**
   - HTTP pprof 端点
   - 性能分析工具

2. **添加健康检查**
   - Goroutine 数量检查
   - 内存使用检查

## 测试建议

### 1. Goroutine 泄漏测试
```go
func TestNoGoroutineLeak(t *testing.T) {
    initialGoroutines := runtime.NumGoroutine()

    // 执行可能泄漏 goroutine 的操作
    StartAgentMetricsScheduler()
    time.Sleep(2 * time.Second)
    StopAgentMetricsScheduler()
    time.Sleep(1 * time.Second)

    finalGoroutines := runtime.NumGoroutine()

    // 允许 10% 的误差
    diff := finalGoroutines - initialGoroutines
    assert.Less(t, diff, initialGoroutines/10)
}
```

### 2. 内存泄漏测试
```bash
# 使用 pprof 监控内存
curl http://localhost:8082/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# 查找内存泄漏
go tool pprof -http=:8080 heap.prof
```

### 3. 运行时监控
```go
// 添加监控中间件
func GoroutineMonitor() gin.HandlerFunc {
    return func(c *gin.Context) {
        goroutines := runtime.NumGoroutine()
        var m runtime.MemStats
        runtime.ReadMemStats(&m)

        logger.Info("运行时状态",
            zap.Int("goroutines", goroutines),
            zap.Uint64("memory_alloc", m.Alloc),
            zap.Uint64("memory_total_alloc", m.TotalAlloc),
            zap.Uint64("memory_sys", m.Sys))

        c.Next()
    }
}
```

## 监控指标

### 需要监控的指标
1. **Goroutine 数量**:
   - 当前 goroutine 数
   - 峰值 goroutine 数
   - 增长趋势

2. **内存使用**:
   - 堆内存使用
   - 栈内存使用
   - GC 频率

3. **资源使用**:
   - 文件句柄数
   - 网络连接数
   - 数据库连接数

## 审查结论

### 总体评估：🟡 需要改进

当前项目在内存泄漏管理方面存在一些问题：

1. **Goroutine 泄漏**: 🔴 P1 严重问题
2. **连接管理**: ✅ 良好
3. **资源清理**: 🟡 部分需要改进

### 风险评估

- **风险等级**: 🟡 中等
- **主要风险**: Agent 调度器 goroutine 泄漏
- **影响**: 长时间运行后内存和 CPU 持续增长

### 下一步行动

1. **立即修复**（P1）:
   - [ ] 修复 Agent 调度器 goroutine 泄漏
   - [ ] 添加所有 ticker 的 Stop() 调用
   - [ ] 添加服务关闭时的清理逻辑

2. **短期优化**（P2）:
   - [ ] 添加并发采集的 panic 恢复
   - [ ] 添加运行时监控

3. **长期改进**（P3）:
   - [ ] 添加 pprof 支持
   - [ ] 添加健康检查端点

---

**审查人**: Claude Code AI
**审查状态**: 🟡 需要改进
**优先级**: P1（立即修复 goroutine 泄漏）
**预期收益**: 防止内存持续增长，提高系统稳定性
