# Jenkins 同步功能配置指南

## 功能说明

**已实现 Jenkins Script Console API 支持，现在可以正常使用"同步角色"和"同步用户"功能！**

## 配置步骤

### 1. 前提条件

确保 Jenkins 满足以下条件：

1. ✅ **用户有管理员权限**：执行 Groovy 脚本需要管理员权限
2. ✅ **启用 Script Console**：Jenkins 默认启用
3. 🔄 **安装 Role Strategy 插件**（可选）：如果要同步角色，需要安装此插件

### 2. 配置 Jenkins 应用

在"授权中心 > 应用管理"中配置 Jenkins：

```json
{
  "应用名称": "Jenkins",
  "应用代码": "jenkins",
  "应用类型": "jenkins",  // 重要：必须选择 jenkins
  "Base URL": "http://jenkins.hzmeipingmi.com",  // 注意：末尾不要斜杠

  "认证配置": {
    "认证方式": "Basic Auth",  // 重要：必须使用 Basic Auth
    "用户名": "admin",
    "密码": "ZRfX6x%dv@4pfKE&K8h"  // 或使用 API Token
  },

  "API 端点配置": {
    // 这些端点可以留空，系统会自动使用 Script Console API
    "创建用户": "",
    "获取用户列表": "",
    "获取角色列表": "",
    "分配角色": "",
    "其他端点": ""
  }
}
```

### 3. 测试同步功能

#### 3.1 同步用户

1. 点击"同步用户"按钮
2. 系统会通过 Script Console API 执行 Groovy 脚本获取用户列表
3. 成功后会显示同步的用户数量

**预期结果：**
```
✅ 同步 Jenkins 用户成功
   - 用户数量：43
   - 包含：admin, changyan, chenguangji, duyeting...
```

#### 3.2 同步角色

1. 点击"同步角色"按钮
2. 系统会尝试获取 Jenkins 中的角色列表

**预期结果：**

**情况 A：已安装 Role Strategy 插件**
```
✅ 同步 Jenkins 角色成功
   - 角色数量：5
   - 包含：admin, developer, reader...
```

**情况 B：未安装 Role Strategy 插件**
```
✅ 同步 Jenkins 角色成功
   - 角色数量：2（默认角色）
   - 包含：admin, developer
   - 说明：请在 Jenkins 中配置角色
```

## 技术实现

### 工作原理

1. **检测应用类型**：当应用类型为 `jenkins` 时，自动使用 Script Console API
2. **获取 CSRF Token**：每次请求前先获取 Jenkins Crumb
3. **执行 Groovy 脚本**：通过 `/scriptText` 端点执行脚本
4. **解析结果**：将脚本输出解析为用户/角色列表

### Groovy 脚本示例

#### 获取用户列表

```groovy
import jenkins.model.*
import hudson.tasks.Mailer

def users = Jenkins.instance.securityRealm.getAllUsers()
users.each { user ->
  def email = user.getProperty(Mailer.UserProperty)?.address ?: ""
  println "${user.id}|${user.fullName}|${email}|active"
}
```

#### 获取角色列表（需要 Role Strategy 插件）

```groovy
import com.michelin.cio.hudson.plugins.rolestrategy.*
import com.synopsys.arc.jenkins.plugins.rolestrategy.*
import jenkins.model.*

def rbas = Jenkins.instance.getAuthorizationStrategy()
if (rbas instanceof RoleBasedAuthorizationStrategy) {
    def globalRoles = rbas.getRoleMap(RoleType.Global).getRoles()
    globalRoles.each { role ->
        println "${role.name}|${role.name}|global|Global Role"
    }
}
```

## 常见问题

### Q1: 同步用户失败，提示"获取 CSRF Token 失败"？

**原因：** 用户名或密码错误，或用户没有管理员权限

**解决：**
1. 检查用户名密码是否正确
2. 确认用户有管理员权限
3. 尝试使用 API Token 而不是密码

### Q2: 同步角色失败，或只显示默认角色？

**原因：** 未安装 Role Strategy 插件

**解决：**
1. 进入 Jenkins > 系统管理 > 插件管理
2. 搜索并安装 "Role-based Authorization Strategy"
3. 在系统管理 > 全局安全配置中启用 "Role-Based Strategy"
4. 重新同步角色

### Q3: 如何生成 Jenkins API Token？

1. 登录 Jenkins
2. 点击右上角用户名 > 设置
3. API Token 区域 > 添加新 Token
4. 复制生成的 Token
5. 在认证配置中：
   - 用户名：admin
   - 密码：粘贴 API Token

### Q4: 为什么必须选择应用类型为 "jenkins"？

系统会根据应用类型自动选择 API 调用方式：
- **jenkins**：使用 Script Console API + Groovy 脚本
- **其他类型**：使用标准 REST API

### Q5: Script Console 有安全风险吗？

是的，Script Console 权限很高。建议：
1. 创建专用的 API 用户
2. 只授予必要的权限
3. 定期审计脚本执行日志
4. 使用 API Token 而不是密码

## 测试命令

### 手动测试 Script Console API

```bash
# 1. 获取 CSRF Token
curl -u admin:'ZRfX6x%dv@4pfKE&K8h' \
  'http://jenkins.hzmeipingmi.com/crumbIssuer/api/json'

# 2. 执行脚本（替换 CRUMB_VALUE）
curl -u admin:'ZRfX6x%dv@4pfKE&K8h' \
  -X POST \
  -H "Jenkins-Crumb: CRUMB_VALUE" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "script=Jenkins.instance.securityRealm.getAllUsers().each { println it.id }" \
  http://jenkins.hzmeipingmi.com/scriptText
```

## 下一步

现在您可以：

1. ✅ **配置 Jenkins 应用**：按照上述步骤配置
2. ✅ **测试同步功能**：点击"同步用户"和"同步角色"
3. ✅ **配置角色绑定**：在"角色绑定"页面映射角色
4. ✅ **授权用户**：在"用户授权"页面为用户分配角色

## 支持的 Jenkins 版本

- Jenkins 2.x 及以上
- 需要 Script Console 功能（默认启用）
- 角色同步需要 Role Strategy 插件 2.x 及以上

## 更新日志

### v1.0 (2026-07-22)
- ✅ 实现 Script Console API 支持
- ✅ 实现用户同步功能
- ✅ 实现角色同步功能
- ✅ 自动处理 CSRF Token
- ✅ 支持 Basic Auth 认证
- ✅ 详细的错误提示
