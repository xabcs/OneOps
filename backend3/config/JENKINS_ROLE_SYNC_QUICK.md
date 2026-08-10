# Jenkins 角色同步 - 快速理解

## 一句话说明

**通过执行 Groovy 脚本访问 Jenkins 内部 API 来获取角色列表**

## 工作原理（简化版）

```
OneOps 系统
    ↓
调用 Jenkins Script Console API
    ↓
执行 Groovy 脚本
    ↓
脚本访问 Jenkins 内部对象
    ↓
获取角色信息并返回
```

## 核心代码片段

### 1. Groovy 脚本（获取角色）

```groovy
import com.michelin.cio.hudson.plugins.rolestrategy.*

def rbas = Jenkins.instance.getAuthorizationStrategy()

if (rbas instanceof RoleBasedAuthorizationStrategy) {
    def globalRoles = rbas.getRoleMap(RoleType.Global).getRoles()
    globalRoles.each { role ->
        println "${role.name}|${role.name}|global|Global Role"
    }
}
```

### 2. Go 代码（执行脚本）

```go
// 1. 获取 CSRF Token
crumb := getJenkinsCrumb(baseURL, username, password)

// 2. 执行 Groovy 脚本
result := executeJenkinsScript(baseURL, username, password, groovyScript)

// 3. 解析结果
roles := parseRoleList(result)
```

## 实际执行示例

### HTTP 请求流程

```http
### 步骤 1: 获取 CSRF Token
GET /crumbIssuer/api/json HTTP/1.1
Host: jenkins.hzmeipingmi.com
Authorization: Basic YWRtaW46WlJmWDZ4JWR2QDRwZktFJks4aA==

Response:
{
  "crumb": "83ebb80f7dd5419461a8b83ebd8b0d122d0f1f6e...",
  "crumbRequestField": "Jenkins-Crumb"
}

### 步骤 2: 执行脚本
POST /scriptText HTTP/1.1
Host: jenkins.hzmeipingmi.com
Authorization: Basic YWRtaW46WlJmWDZ4JWR2QDRwZktFJks4aA==
Jenkins-Crumb: 83ebb80f7dd5419461a8b83ebd8b0d122d0f1f6e...
Content-Type: application/x-www-form-urlencoded

script=import jenkins.model.*
def rbas = Jenkins.instance.getAuthorizationStrategy()
...

Response:
admin|admin|global|Global Role
developer|developer|global|Global Role
reader|reader|global|Global Role
```

## 为什么这样实现？

### 问题
- ❌ Jenkins 没有 `/api/roles` 这样的 REST API
- ❌ 无法通过标准 HTTP 请求获取角色列表
- ❌ 角色信息存储在 Jenkins 内部对象中

### 解决
- ✅ Script Console API 提供执行任意 Groovy 代码的能力
- ✅ Groovy 脚本可以访问 Jenkins 所有内部对象
- ✅ 通过脚本输出获取角色信息

## 必需条件

| 条件 | 说明 | 是否必需 |
|------|------|---------|
| 管理员权限 | 执行 Script Console 需要 | ✅ 必需 |
| Basic Auth 认证 | Jenkins 认证方式 | ✅ 必需 |
| Role Strategy 插件 | 管理角色的插件 | 🔄 可选（但强烈建议） |
| 应用类型 = jenkins | 系统自动选择 Script Console | ✅ 必需 |

## 测试方法

```bash
# 快速测试 Script Console
curl -u admin:'password' \
  -X POST \
  -H "Jenkins-Crumb: $(curl -s -u admin:'password' http://jenkins.example.com/crumbIssuer/api/json | jq -r .crumb)" \
  -d "script=println 'Hello'" \
  http://jenkins.example.com/scriptText
```

## 数据流程图

```
┌─────────────┐
│  用户操作   │ 点击"同步角色"
└──────┬──────┘
       │
       ↓
┌─────────────────────────────┐
│  检测应用类型                │
│  type == "jenkins" ? ✓      │
└──────┬──────────────────────┘
       │
       ↓
┌─────────────────────────────┐
│  获取 CSRF Token             │
│  GET /crumbIssuer/api/json  │
│  返回: {crumb: "abc123"}    │
└──────┬──────────────────────┘
       │
       ↓
┌─────────────────────────────┐
│  执行 Groovy 脚本            │
│  POST /scriptText           │
│  Body: script=Groovy代码    │
│  Header: Jenkins-Crumb      │
└──────┬──────────────────────┘
       │
       ↓
┌─────────────────────────────┐
│  脚本输出                    │
│  admin|admin|global|...     │
│  developer|developer|...    │
└──────┬──────────────────────┘
       │
       ↓
┌─────────────────────────────┐
│  解析输出                    │
│  按行分割，按|分割字段       │
└──────┬──────────────────────┘
       │
       ↓
┌─────────────────────────────┐
│  保存到数据库                │
│  application_roles 表       │
└─────────────────────────────┘
```

## 关键技术点

### 1. CSRF 保护
```go
// 每次请求前获取新的 Crumb
crumb := getJenkinsCrumb()
req.Header.Set("Jenkins-Crumb", crumb)
```

### 2. Basic Auth 认证
```go
// 设置 Basic Auth
req.SetBasicAuth(username, password)
```

### 3. 脚本执行
```go
// 通过 POST 请求执行脚本
data := "script=" + groovyScript
req.Body = strings.NewReader(data)
```

### 4. 结果解析
```go
// 解析格式：code|name|type|description
lines := strings.Split(result, "\n")
for _, line := range lines {
    parts := strings.Split(line, "|")
    role := ApplicationRole{
        RoleCode: parts[0],
        RoleName: parts[1],
        ...
    }
}
```

## 总结

**Jenkins 角色同步 = Script Console API + Groovy 脚本**

1. **为什么？** Jenkins 没有标准 REST API
2. **怎么做？** 执行 Groovy 脚本访问内部对象
3. **需要什么？** 管理员权限 + Basic Auth + Role Strategy 插件
4. **数据格式？** 管道符分隔的文本
5. **安全吗？** 需要正确处理 CSRF Token

---

**详细文档：** `JENKINS_ROLE_SYNC_TECHNICAL.md`
