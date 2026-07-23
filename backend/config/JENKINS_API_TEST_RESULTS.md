# Jenkins API 实际测试结果与集成方案

## 测试结果

### ✅ 可以使用的 API 端点

1. **根 API**
   ```
   GET /api/json
   返回：Jenkins 系统信息、任务列表等
   ```

2. **用户信息**
   ```
   GET /user/{username}/api/json
   返回：指定用户的详细信息
   ```

### ❌ 不可用的 API 端点

1. **角色列表** - `/roleStrategy/roles` - 404 Not Found
2. **用户列表** - `/securityRealm/api/json` - 404 Not Found
3. **用户搜索** - `/userSearch/api/json` - 404 Not Found

## 问题分析

### 1. Jenkins 没有 REST API 来获取用户列表

Jenkins 的设计理念是通过 Web UI 管理用户，而不是通过 REST API。要获取用户列表，需要：

- 使用 Jenkins Script Console (Groovy 脚本)
- 使用 Jenkins CLI
- 直接访问数据库（不推荐）

### 2. 角色管理依赖插件

Role Strategy 插件提供角色管理功能，但其 REST API 端点路径不是标准的 `/roleStrategy/roles`。

根据插件文档，需要访问：
- `/role-strategy/` (注意连字符)
- 或者通过 Script Console 执行 Groovy 脚本

## 推荐方案

### 方案一：使用 Script Console API（推荐）

Jenkins 提供 Script Console API，可以通过 POST 请求执行 Groovy 脚本。

#### 获取用户列表

```bash
POST /scriptText
Content-Type: application/x-www-form-urlencoded

script=import jenkins.model.*
def users = Jenkins.instance.securityRealm.getAllUsers()
users.each { user ->
  println "${user.id},${user.fullName},${user.getProperty(hudson.tasks.Mailer$UserProperty)?.address}"
}
```

#### 获取角色列表（需要 Role Strategy 插件）

```groovy
POST /scriptText

script=import com.michelin.cio.hudson.plugins.rolestrategy.*
import com.synopsys.arc.jenkins.plugins.rolestrategy.*

def rbas = Jenkins.instance.getAuthorizationStrategy()
if (rbas instanceof RoleBasedAuthorizationStrategy) {
    def globalRoles = rbas.getRoleMap(RoleType.Global).getRoles()
    globalRoles.each { role ->
        println "${role.name},${role.permissions}"
    }
}
```

#### 创建用户

```groovy
POST /scriptText

script=import jenkins.model.*
import hudson.security.*

def instance = Jenkins.getInstance()
def realm = instance.getSecurityRealm()
if (realm instanceof HudsonPrivateSecurityRealm) {
    def user = realm.createAccount("username", "password")
    user.setFullName("Full Name")
    user.save()
    println "User created: ${user.id}"
}
```

#### 分配角色（需要 Role Strategy 插件）

```groovy
POST /scriptText

script=import com.michelin.cio.hudson.plugins.rolestrategy.*
import com.synopsys.arc.jenkins.plugins.rolestrategy.*
import com.michelin.cio.hudson.plugins.rolestrategy.AuthorizationType

def rbas = Jenkins.instance.getAuthorizationStrategy()
if (rbas instanceof RoleBasedAuthorizationStrategy) {
    def roleMap = rbas.getRoleMap(RoleType.Global)
    def role = roleMap.getRoles().find { it.name == "rolename" }
    if (role) {
        roleMap.assignRole(role, new PermissionEntry(AuthorizationType.USER, "username"))
        println "Role assigned"
    }
}
```

### 方案二：使用 Jenkins CLI

Jenkins CLI 提供部分用户管理功能：

```bash
# 下载 jenkins-cli.jar
wget http://jenkins.hzmeipingmi.com/jnlpJars/jenkins-cli.jar

# 创建用户
java -jar jenkins-cli.jar -s http://jenkins.hzmeipingmi.com/ \
  -auth admin:ZRfX6x%dv@4pfKE&K8h \
  create-user username password

# 列出用户（需要 groovy 命令）
java -jar jenkins-cli.jar -s http://jenkins.hzmeipingmi.com/ \
  -auth admin:ZRfX6x%dv@4pfKE&K8h \
  groovy = < list-users.groovy
```

### 方案三：跳过同步，手动配置（最简单）

由于 Jenkins API 的限制，最简单的方案是：

1. **不使用"同步角色"和"同步用户"功能**
2. 在"角色绑定"页面手动配置 Jenkins 角色映射
3. 授权时使用 Groovy 脚本创建用户

## 实际配置步骤

### 步骤 1：更新应用配置

编辑 Jenkins 应用，使用正确的端点配置：

```json
{
  "name": "Jenkins",
  "code": "jenkins",
  "type": "jenkins",
  "baseUrl": "http://jenkins.hzmeipingmi.com",  // 注意：末尾不要斜杠
  "authConfig": {
    "type": "basic",
    "username": "admin",
    "password": "ZRfX6x%dv@4pfKE&K8h"
  },
  "endpoints": {
    "createUser": "/scriptText",           // 使用 Script Console
    "listUsers": "",                        // 留空，不使用
    "getUser": "/user/{username}/api/json", // 可用
    "getRoles": "",                         // 留空，不使用
    "assignRole": "/scriptText",            // 使用 Script Console
    "revokeRole": "",
    "getUserRoles": ""
  }
}
```

### 步骤 2：修改后端实现

需要修改 `CreateExternalUser` 和 `GrantRoleToUser` 方法，使其支持通过 Script Console API 执行 Groovy 脚本。

### 步骤 3：手动配置角色绑定

1. 进入"授权中心 > 角色绑定"
2. 手动添加 Jenkins 角色映射
3. 例如：
   - 授权中心角色 "开发人员" -> Jenkins 角色 "developer"
   - 授权中心角色 "管理员" -> Jenkins 角色 "admin"

## 安全注意事项

### Script Console 权限

使用 Script Console API 需要管理员权限。建议：

1. 创建专用的 API 用户
2. 只授予必要的权限
3. 记录所有脚本执行日志
4. 定期审计脚本内容

### CSRF 保护

Jenkins 默认启用 CSRF 保护，POST 请求需要包含 CSRF Token：

```bash
# 1. 获取 CSRF Token
CRUMB=$(curl -s -u admin:password 'http://jenkins.example.com/crumbIssuer/api/xml?xpath=concat(//crumbRequestField,":",//crumb)')

# 2. 使用 Token 发送请求
curl -u admin:password -H "$CRUMB" -X POST -d "script=..." http://jenkins.example.com/scriptText
```

## 总结

### Jenkins API 现状

- ✅ Basic Auth 认证可用
- ✅ 获取单个用户信息可用
- ❌ 没有标准的用户列表 API
- ❌ 没有标准的角色管理 API
- ✅ Script Console API 可用（需要管理员权限）

### 推荐集成方案

1. **短期方案**：跳过同步功能，手动配置角色绑定
2. **长期方案**：实现 Script Console API 调用，通过 Groovy 脚本管理用户和角色

### 下一步行动

1. ✅ 已修复 URL 双斜杠问题
2. ✅ 已支持 Basic Auth 认证
3. 🔄 需要实现 Script Console API 调用逻辑
4. 🔄 需要编写 Groovy 脚本模板
5. 🔄 需要处理 CSRF Token

## 参考资料

- [Jenkins Remote Access API](https://www.jenkins.io/doc/book/using/remote-access-api/)
- [Role Strategy Plugin](https://plugins.jenkins.io/role-strategy/)
- [Jenkins Script Console](https://www.jenkins.io/doc/book/managing/script-console/)
