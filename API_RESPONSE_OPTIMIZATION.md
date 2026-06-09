# API 响应优化计划

## 优化日期
2026年6月9日

## 优化范围
- 响应压缩
- 请求限流
- 响应大小优化

## 当前状态

### ✅ 已实现的功能

#### 1. 分页功能
**状态**: ✅ 已实现

**实现位置**:
- `services/audit.go` - 审计日志分页
- `services/agent.go` - Agent升级任务分页
- `services/cmdb.go` - 服务器列表分页

**实现方式**:
```go
offset := (page - 1) * pageSize
query.Limit(pageSize).Offset(offset)
```

**优点**:
- 统一的分页参数（page, pageSize）
- 返回总数（total）用于前端显示

#### 2. 统一响应格式
**状态**: ✅ 已实现

**实现位置**: `middlewares/response.go`

**响应格式**:
```go
{
    "code": 200,
    "message": "success",
    "data": {...},
    "total": 100  // 分页查询时返回
}
```

### ❌ 未实现的功能

#### 1. 响应压缩
**状态**: ❌ 未实现

**影响**:
- API 响应体积大
- 网络传输时间长
- 带宽消耗高

#### 2. 请求限流
**状态**: ❌ 未实现

**风险**:
- 容易遭受 DDoS 攻击
- 恶意用户可能耗尽服务器资源
- 无保护机制

## 优化方案

### 1. 响应压缩（P1 优先级）

#### 实施方案
使用 Gin 的 gzip 中间件

#### 实施步骤

**步骤1**: 安装依赖
```bash
go get github.com/gin-contrib/gzip
```

**步骤2**: 修改 `routes/routes.go`
```go
import (
    "github.com/gin-contrib/gzip"
    // ... 其他导入
)

// SetupRoutes 设置路由
func SetupRoutes(r *gin.Engine) {
    // 添加 gzip 压缩中间件（必须在其他中间件之前）
    r.Use(gzip.Gzip(gzip.DefaultCompression))

    // 应用其他中间件
    r.Use(gin.Recovery())
    r.Use(middlewares.Response())
    // ...
}
```

**步骤3**: 配置压缩级别
```go
// 压缩级别选项：
// - gzip.NoCompression: 不压缩
// - gzip.BestSpeed: 最快压缩
// - gzip.BestCompression: 最高压缩率
// - gzip.DefaultCompression: 默认（推荐）
// - gzip.NoCompression: 不压缩

// 排除特定路径（如 WebSocket）
r.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedPaths([]string{"/api/ws"})))
```

#### 预期收益
- JSON 响应体积减少 70-90%
- 网络传输时间减少 50-70%
- 带宽消耗降低 60-80%

#### 测试验证
```bash
# 测试压缩是否生效
curl -H "Accept-Encoding: gzip" http://localhost:8082/api/cmdb/servers --output - | gunzip

# 检查响应头
curl -I http://localhost:8082/api/cmdb/servers
# 应该看到: Content-Encoding: gzip
```

### 2. 请求限流（P1 优先级）

#### 实施方案
使用基于 IP 的限流中间件

#### 实施步骤

**步骤1**: 安装依赖
```bash
go get github.com/ulule/limiter/v3
go get github.com/ulule/limiter/v3/drivers/gin
go get golang.org/x/time/rate
```

**步骤2**: 创建限流中间件 `middlewares/rate_limit.go`
```go
package middlewares

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "golang.org/x/time/rate"
)

// RateLimiter 限流器
type RateLimiter struct {
    limiter *rate.Limiter
}

// NewRateLimiter 创建限流器
// r: 每秒允许的请求数
// b: 突发请求数（允许的瞬时最大请求数）
func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
    return &RateLimiter{
        limiter: rate.NewLimiter(r, b),
    }
}

// Middleware 返回 Gin 中间件
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        if !rl.limiter.Allow() {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "code":    429,
                "message": "请求过于频繁，请稍后再试",
            })
            c.Abort()
            return
        }
        c.Next()
    }
}
```

**步骤3**: 修改 `routes/routes.go`
```go
// 创建限流器（每秒 100 个请求，突发 200 个）
rateLimiter := middlewares.NewRateLimiter(100, 200)

// SetupRoutes 设置路由
func SetupRoutes(r *gin.Engine) {
    // 应用限流中间件（全局限流）
    r.Use(rateLimiter.Middleware())

    // 或针对特定路由组限流
    api := r.Group("/api")
    api.Use(rateLimiter.Middleware())
    // ...
}
```

**步骤4**: 高级限流配置（按 IP 或用户）
```go
// 创建基于 IP 的限流器
func IPRateLimiter(rps int) gin.HandlerFunc {
    // 使用 map 存储每个 IP 的限流器
    limiters := make(map[string]*rate.Limiter)
    var mu sync.Mutex

    return func(c *gin.Context) {
        ip := c.ClientIP()

        mu.Lock()
        limiter, exists := limiters[ip]
        if !exists {
            limiter = rate.NewLimiter(rate.Limit(rps), rps*2)
            limiters[ip] = limiter
        }
        mu.Unlock()

        if !limiter.Allow() {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "code":    429,
                "message": "请求过于频繁，请稍后再试",
            })
            c.Abort()
            return
        }

        c.Next()
    }
}
```

#### 限流策略建议

**不同路由的限流配置**:

| 路由类型 | 限流配置 | 原因 |
|---------|---------|------|
| 认证路由（/api/login） | 10 req/s | 防止暴力破解 |
| 查询路由（/api/cmdb/servers） | 100 req/s | 允许较高并发 |
| 写操作（POST/PUT/DELETE） | 50 req/s | 保护数据库 |
| WebSocket（/api/ws） | 不限流 | 长连接特殊处理 |
| 静态资源 | 不限流 | 通过 CDN 处理 |

#### 预期收益
- 防止 DDoS 攻击
- 保护服务器资源
- 提供公平的服务访问

#### 测试验证
```bash
# 使用 Apache Bench 进行压力测试
ab -n 1000 -c 50 http://localhost:8082/api/cmdb/servers

# 应该看到部分请求返回 429 状态码
```

### 3. 响应大小优化（P2 优先级）

#### 优化方案

**方案1**: 字段过滤（按需返回）
```go
// 只返回客户端需要的字段
type ServerListItem struct {
    ID        uint   `json:"id"`
    Hostname  string `json:"hostname"`
    IP        string `json:"ip"`
    Status    string `json:"status"`
    // 不返回大字段（如 Attributes、Tags）
}

func (s *CMDBService) GetServersLight(query map[string]interface{}, page, pageSize int) ([]ServerListItem, int64, error) {
    // ...
}
```

**方案2**: 列表项和详情分离
```go
// 列表接口只返回核心字段
GET /api/cmdb/servers
// 返回: id, hostname, ip, status

// 详情接口返回完整数据
GET /api/cmdb/servers/:id
// 返回: 所有字段
```

**方案3**: 数据压缩存储
```go
// 对于大的 JSON 字段，使用压缩存储
type Server struct {
    // ...
    AttributesJSON string `json:"-"` // 压缩存储
    Attributes     map[string]interface{} `json:"attributes" gorm:"-"`
}
```

#### 预期收益
- 列表 API 响应体积减少 30-50%
- 查询性能提升 20-30%

### 4. 响应时间优化（P2 优先级）

#### 优化方案

**方案1**: 并发查询
```go
// 使用 goroutine 并发查询多个数据源
func (s *CMDBService) GetServerDetail(id uint) (*ServerDetail, error) {
    var wg sync.WaitGroup
    var server models.Server
    var credentials []models.SSHCredential
    var tags []models.Tag
    var errServer, errCred, errTags error

    wg.Add(3)

    // 并发查询服务器信息
    go func() {
        defer wg.Done()
        errServer = db.First(&server, id).Error
    }()

    // 并发查询凭证
    go func() {
        defer wg.Done()
        errCred = db.Where("server_id = ?", id).Find(&credentials).Error
    }()

    // 并发查询标签
    go func() {
        defer wg.Done()
        errTags = db.Joins("JOIN server_tags ON server_tags.tag_id = tags.id").
            Where("server_tags.server_id = ?", id).
            Find(&tags).Error
    }()

    wg.Wait()

    // 检查错误
    if errServer != nil {
        return nil, errServer
    }

    return &ServerDetail{
        Server:      server,
        Credentials: credentials,
        Tags:        tags,
    }, nil
}
```

**方案2**: 缓存热点数据
```go
// 缓存频繁访问但不常变化的数据
func (s *CMDBService) GetServerByID(id uint) (*models.Server, error) {
    // 先从 Redis 获取
    cacheKey := fmt.Sprintf("server:%d", id)
    var server models.Server
    if err := redisCache.Get(cacheKey, &server); err == nil {
        return &server, nil
    }

    // 从数据库查询
    err := db.Preload("Credentials").First(&server, id).Error
    if err != nil {
        return nil, err
    }

    // 存入缓存（5 分钟）
    redisCache.SetWithTTL(cacheKey, server, 5*time.Minute)
    return &server, nil
}
```

## 实施优先级

### P1（立即实施）
1. **响应压缩** - 高优先级，收益明显
2. **请求限流** - 安全必需，防止滥用

### P2（短期实施）
1. **响应大小优化** - 字段过滤
2. **响应时间优化** - 并发查询

### P3（长期优化）
1. **数据压缩存储** - 针对大字段
2. **CDN 集成** - 静态资源加速

## 测试计划

### 1. 压缩测试
```bash
# 测试压缩效果
curl http://localhost:8082/api/cmdb/servers -H "Accept-Encoding: gzip" --output - | wc -c
curl http://localhost:8082/api/cmdb/servers --output - | wc -c

# 计算压缩率
# 压缩率 = (1 - 压缩后大小 / 原始大小) * 100%
```

### 2. 限流测试
```bash
# 测试限流是否生效
for i in {1..200}; do
    curl http://localhost:8082/api/cmdb/servers &
done
wait

# 应该看到部分请求返回 429
```

### 3. 性能测试
```bash
# 对比优化前后的响应时间
ab -n 1000 -c 10 http://localhost:8082/api/cmdb/servers

# 关注指标：
# - Requests per second (RPS)
# - Time per request (平均响应时间)
# - Transfer rate (传输速率)
```

## 监控指标

### 需要监控的指标
1. **响应大小**:
   - 平均响应体积
   - 90/95/99 分位数
   - 压缩率

2. **响应时间**:
   - P50/P90/P99 延迟
   - 慢请求数量

3. **限流效果**:
   - 被限流的请求数
   - 限流触发频率
   - IP 分布

4. **错误率**:
   - 4xx 错误率
   - 5xx 错误率

## 审查结论

### 总体评估：🟡 需要改进

当前项目在 API 响应优化方面还有改进空间：

1. **响应压缩**: ❌ 未实现（P1 优先级）
2. **请求限流**: ❌ 未实现（P1 优先级，安全必需）
3. **分页功能**: ✅ 已正确实现
4. **响应格式**: ✅ 统一格式

### 风险评估

- **风险等级**: 🟡 中等
- **主要风险**: 无限流保护，易受攻击
- **建议**: 立即实施 P1 优化项

### 下一步行动

1. **立即执行**（P1）:
   - [ ] 实施响应压缩
   - [ ] 实施请求限流
   - [ ] 进行压力测试

2. **短期优化**（P2）:
   - [ ] 优化响应大小
   - [ ] 实施并发查询

3. **长期优化**（P3）:
   - [ ] 数据压缩存储
   - [ ] CDN 集成

---

**审查人**: Claude Code AI
**审查状态**: 🟡 需要改进
**优先级**: P1（立即实施）
**预期收益**: 响应体积减少 70%，安全防护提升
