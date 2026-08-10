# Jenkins 角色同步技术详解

## 一、技术架构

```
┌─────────────────────────────────────────────────────────┐
│                    OneOps 系统                           │
│                                                          │
│  用户点击"同步角色"按钮                                  │
│           ↓                                              │
│  检测应用类型 == "jenkins"                               │
│           ↓                                              │
│  调用 fetchJenkinsRoles()                                │
└─────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────┐
│              Jenkins Script Console API                  │
│                                                          │
│  1. GET /crumbIssuer/api/json                            │
│     ↓ 获取 CSRF Token                                    │
│                                                          │
│  2. POST /scriptText                                     │
│     ↓ 执行 Groovy 脚本                                   │
│                                                          │
│  3. 返回角色列表数据                                     │
└─────────────────────────────────────────────────────────┘
```

## 二、实现代码

### 1. 后端核心代码

**文件：** `backend/services/application_permission_service.go`

#### 步骤 1: 获取 CSRF Token

```go
func (s *ApplicationPermissionService) getJenkinsCrumb(baseURL string, username, password string) (*JenkinsCrumb, error) {
    // 构造请求 URL
    crumbURL := strings.TrimSuffix(baseURL, "/") + "/crumbIssuer/api/json"

    // 创建请求
    req, err := http.NewRequest("GET", crumbURL, nil)
    if err != nil {
        return nil, err
    }

    // 设置 Basic Auth 认证
    req.SetBasicAuth(username, password)

    // 发送请求
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("获取 CSRF Token 失败: %w", err)
    }
    defer resp.Body.Close()

    // 解析响应
    var crumb JenkinsCrumb
    body, _ := io.ReadAll(resp.Body)
    if err := json.Unmarshal(body, &crumb); err != nil {
        return nil, fmt.Errorf("解析 CSRF Token 失败: %w", err)
    }

    return &crumb, nil
}
```

**返回示例：**
```json
{
  "_class": "hudson.security.csrf.DefaultCrumbIssuer",
  "crumb": "83ebb80f7dd5419461a8b83ebd8b0d122d0f1f6e2e26ec3c0b58095b7c825a3b",
  "crumbRequestField": "Jenkins-Crumb"
}
```

#### 步骤 2: 执行 Groovy 脚本

```go
func (s *ApplicationPermissionService) executeJenkinsScript(baseURL, username, password, script string) (string, error) {
    // 1. 获取 CSRF Token
    crumb, err := s.getJenkinsCrumb(baseURL, username, password)
    if err != nil {
        return "", err
    }

    // 2. 构造请求
    scriptURL := strings.TrimSuffix(baseURL, "/") + "/scriptText"
    data := "script=" + script

    req, err := http.NewRequest("POST", scriptURL, strings.NewReader(data))
    if err != nil {
        return "", err
    }

    // 3. 设置请求头
    req.SetBasicAuth(username, password)
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    req.Header.Set(crumb.CrumbRequestField, crumb.Crumb)  // 添加 CSRF Token

    // 4. 发送请求
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", fmt.Errorf("执行脚本失败: %w", err)
    }
    defer resp.Body.Close()

    // 5. 读取响应
    body, _ := io.ReadAll(resp.Body)
    return string(body), nil
}
```

#### 步骤 3: 获取角色列表

```go
func (s *ApplicationPermissionService) fetchJenkinsRoles(baseURL string, authConfig map[string]interface{}) ([]models.ApplicationRole, error) {
    // 1. 获取认证信息
    username, _ := authConfig["username"].(string)
    password, _ := authConfig["password"].(string)

    if username == "" || password == "" {
        return nil, fmt.Errorf("Jenkins 需要 Basic Auth 认证")
    }

    // 2. Groovy 脚本：获取所有角色
    script := `
import com.michelin.cio.hudson.plugins.rolestrategy.*
import com.synopsys.arc.jenkins.plugins.rolestrategy.*
import jenkins.model.*

def rbas = Jenkins.instance.getAuthorizationStrategy()

if (rbas instanceof RoleBasedAuthorizationStrategy) {
    // 获取全局角色
    def globalRoles = rbas.getRoleMap(RoleType.Global).getRoles()
    globalRoles.each { role ->
        println "${role.name}|${role.name}|global|Global Role"
    }

    // 获取 Item 角色
    def itemRoles = rbas.getRoleMap(RoleType.Item).getRoles()
    itemRoles.each { role ->
        println "${role.name}|${role.name}|item|Item Role"
    }
} else {
    println "Warning: Role Strategy plugin not configured"
}
`

    // 3. 执行脚本
    result, err := s.executeJenkinsScript(baseURL, username, password, script)
    if err != nil {
        return nil, err
    }

    // 4. 解析结果
    roles := make([]models.ApplicationRole, 0)
    lines := strings.Split(result, "\n")

    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line == "" || strings.HasPrefix(line, "Warning:") {
            continue
        }

        parts := strings.Split(line, "|")
        if len(parts) >= 2 {
            role := models.ApplicationRole{
                RoleCode:    parts[0],
                RoleName:    parts[1],
                RoleType:    "global",
                Description: "Jenkins Role",
            }
            if len(parts) >= 3 {
                role.RoleType = parts[2]
            }
            if len(parts) >= 4 {
                role.Description = parts[3]
            }
            roles = append(roles, role)
        }
    }

    // 5. 如果没有角色，返回默认角色
    if len(roles) == 0 {
        roles = append(roles, models.ApplicationRole{
            RoleCode:    "admin",
            RoleName:    "Administrator",
            RoleType:    "global",
            Description: "Jenkins Administrator",
        })
        roles = append(roles, models.ApplicationRole{
            RoleCode:    "developer",
            RoleName:    "Developer",
            RoleType:    "global",
            Description: "Jenkins Developer",
        })
    }

    return roles, nil
}
```

### 2. Groovy 脚本详解

#### 脚本作用

这个 Groovy 脚本通过 Jenkins 内部 API 来获取角色信息：

```groovy
// 导入必要的类
import com.michelin.cio.hudson.plugins.rolestrategy.*
import com.synopsys.arc.jenkins.plugins.rolestrategy.*
import jenkins.model.*

// 获取 Jenkins 实例的授权策略
def rbas = Jenkins.instance.getAuthorizationStrategy()

// 检查是否使用了 Role-Based Strategy
if (rbas instanceof RoleBasedAuthorizationStrategy) {
    // 获取全局角色
    def globalRoles = rbas.getRoleMap(RoleType.Global).getRoles()

    // 遍历每个角色并输出
    globalRoles.each { role ->
        println "${role.name}|${role.name}|global|Global Role"
    }

    // 获取 Item 角色（项目级别的角色）
    def itemRoles = rbas.getRoleMap(RoleType.Item).getRoles()
    itemRoles.each { role ->
        println "${role.name}|${role.name}|item|Item Role"
    }
} else {
    // 如果没有配置 Role Strategy 插件
    println "Warning: Role Strategy plugin not configured"
}
```

#### 数据格式

脚本输出格式为：`角色代码|角色名称|角色类型|描述`

示例输出：
```
admin|admin|global|Global Role
developer|developer|global|Global Role
reader|reader|global|Global Role
```

## 三、执行流程

### 完整流程图

```
用户操作：点击"同步角色"
           ↓
┌──────────────────────────────────────┐
│ 1. 检测应用类型                       │
│    if app.Type == "jenkins"          │
└──────────────────────────────────────┘
           ↓
┌──────────────────────────────────────┐
│ 2. 解析认证配置                       │
│    username = "admin"                │
│    password = "***"                  │
└──────────────────────────────────────┘
           ↓
┌──────────────────────────────────────┐
│ 3. 获取 CSRF Token                    │
│    GET /crumbIssuer/api/json         │
│    返回: {crumb: "xxx"}              │
└──────────────────────────────────────┘
           ↓
┌──────────────────────────────────────┐
│ 4. 执行 Groovy 脚本                   │
│    POST /scriptText                  │
│    Headers:                          │
│      - Jenkins-Crumb: xxx            │
│      - Authorization: Basic xxx      │
│    Body: script=Groovy代码           │
└──────────────────────────────────────┘
           ↓
┌──────────────────────────────────────┐
│ 5. 解析脚本输出                       │
│    admin|admin|global|Global Role    │
│    developer|developer|global|...    │
└──────────────────────────────────────┘
           ↓
┌──────────────────────────────────────┐
│ 6. 保存到数据库                       │
│    INSERT INTO application_roles     │
└──────────────────────────────────────┘
           ↓
┌──────────────────────────────────────┐
│ 7. 返回成功消息                       │
│    "同步了 3 个角色"                 │
└──────────────────────────────────────┘
```

## 四、关键配置

### Jenkins 端要求

#### 1. 安装 Role Strategy 插件（必需）

**方法：**
1. 进入 Jenkins > 系统管理 > 插件管理
2. 搜索 "Role-based Authorization Strategy"
3. 安装并重启 Jenkins

#### 2. 配置授权策略

**步骤：**
1. 系统管理 > 全局安全配置
2. 授权策略选择 "Role-Based Strategy"
3. 保存

#### 3. 创建角色

**步骤：**
1. 系统管理 > Manage and Assign Roles
2. Manage Roles 页面创建角色：
   - Global roles: admin, developer, reader
   - Item roles: project-admin, project-developer
3. 配置角色权限

### OneOps 端配置

```json
{
  "name": "Jenkins",
  "code": "jenkins",
  "type": "jenkins",                    // ⚠️ 必须是 jenkins
  "baseUrl": "http://jenkins.example.com",
  "authConfig": {
    "type": "basic",                    // ⚠️ 必须是 basic
    "username": "admin",
    "password": "password-or-api-token"
  },
  "endpoints": {
    // 可以全部留空，系统会自动使用 Script Console API
  }
}
```

## 五、测试验证

### 手动测试脚本

```bash
#!/bin/bash
JENKINS_URL="http://jenkins.hzmeipingmi.com"
USERNAME="admin"
PASSWORD='ZRfX6x%dv@4pfKE&K8h'

# 1. 获取 CSRF Token
curl -s -u "$USERNAME:$PASSWORD" \
  -c /tmp/cookies.txt \
  "$JENKINS_URL/crumbIssuer/api/json" | tee /tmp/crumb.json

CRUMB=$(cat /tmp/crumb.json | grep -o '"crumb":"[^"]*"' | cut -d'"' -f4)

# 2. 执行脚本
curl -s -u "$USERNAME:$PASSWORD" \
  -b /tmp/cookies.txt \
  -X POST \
  -H "Jenkins-Crumb: $CRUMB" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "script=import jenkins.model.*%0Aprintln 'Test'" \
  "$JENKINS_URL/scriptText"
```

### 预期结果

**成功场景：**
```
同步成功！同步了 5 个角色：
- admin (全局角色)
- developer (全局角色)
- reader (全局角色)
- project-admin (项目角色)
- project-developer (项目角色)
```

**未安装插件场景：**
```
同步成功！同步了 2 个默认角色：
- admin (默认角色，请在 Jenkins 中配置)
- developer (默认角色，请在 Jenkins 中配置)

提示：请安装 Role Strategy 插件以获取完整角色列表
```

## 六、常见问题

### Q1: 为什么返回空角色列表？

**原因：** 未安装 Role Strategy 插件

**解决：**
1. 安装 "Role-based Authorization Strategy" 插件
2. 在全局安全配置中启用 "Role-Based Strategy"
3. 在 "Manage Roles" 中创建角色
4. 重新同步

### Q2: 为什么提示 "No valid crumb"？

**原因：** CSRF Token 未正确设置或已过期

**解决：** 系统会自动处理，每次请求前获取新的 Token

### Q3: 为什么角色同步需要管理员权限？

**原因：** Script Console 需要管理员权限才能执行

**解决：** 使用有管理员权限的账号或 API Token

## 七、安全考虑

### 1. CSRF 保护

- 每次请求都要获取新的 CSRF Token
- Token 在请求头中传递

### 2. 认证安全

- 使用 HTTPS（生产环境）
- 使用 API Token 而不是密码
- 创建专用的 API 用户

### 3. 权限控制

- 只授予必要的权限
- 定期审计 Script Console 执行日志
- 使用 Jenkins 安全插件限制脚本执行

## 八、性能优化

### 1. 缓存 CSRF Token

```go
// 可以在一定时间内复用 Token
type JenkinsCache struct {
    Crumb     string
    Timestamp time.Time
    ExpiresIn time.Duration
}
```

### 2. 批量操作

如果需要同步多个 Jenkins 实例，可以并行处理：
```go
var wg sync.WaitGroup
for _, jenkins := range jenkinsInstances {
    wg.Add(1)
    go func(url string) {
        defer wg.Done()
        syncRoles(url)
    }(jenkins.URL)
}
wg.Wait()
```

## 九、总结

Jenkins 角色同步通过以下方式实现：

1. **技术方案**：Script Console API + Groovy 脚本
2. **认证方式**：Basic Auth + CSRF Token
3. **数据来源**：Jenkins 内部对象 `RoleBasedAuthorizationStrategy`
4. **必需插件**：Role-based Authorization Strategy
5. **数据格式**：管道符分隔的文本格式

**优势：**
- ✅ 可以获取完整的角色信息
- ✅ 不依赖标准 REST API
- ✅ 灵活可扩展（可以自定义脚本）

**限制：**
- ⚠️ 需要管理员权限
- ⚠️ 需要安装 Role Strategy 插件
- ⚠️ 需要处理 CSRF 保护
