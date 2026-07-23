# Jenkins 集成配置说明

## 问题分析

从错误信息可以看出：
```
认证失败(403)，请检查 Token 是否正确或是否有权限访问 API: http://jenkins.hzmeipingmi.com//api/roles
```

问题：
1. **URL 双斜杠**：已修复 ✅
2. **认证方式错误**：Jenkins 不支持 Bearer Token，需要使用 Basic Auth ✅
3. **API 端点不正确**：Jenkins 没有标准的 `/api/roles` 端点

## Jenkins 特殊配置

### 1. 认证方式

Jenkins 使用 **Basic Auth** 认证，不是 Bearer Token。

**方式一：使用用户名密码（推荐用于测试）**
```json
{
  "type": "basic",
  "username": "admin",
  "password": "ZRfX6x%dv@4pfKE&K8h"
}
```

**方式二：使用 API Token（推荐用于生产）**

1. 登录 Jenkins：http://jenkins.hzmeipingmi.com/
2. 点击右上角用户名 > 设置
3. API Token 区域 > 添加新 Token
4. 生成并复制 Token
5. 配置：
```json
{
  "type": "basic",
  "username": "admin",
  "password": "你的API_TOKEN"  // 注意：这里填 API Token，不是密码
}
```

### 2. Jenkins API 端点

Jenkins 没有标准的 REST API，需要通过插件或特殊端点访问。

#### 推荐方案：使用 Role-Based Strategy 插件

**前提条件：**
- Jenkins 已安装 "Role-based Authorization Strategy" 插件
- 用户有管理员权限

**获取角色列表：**
```
端点: /roleStrategy/roles
方法: GET
```

**注意：** Jenkins 默认没有这个端点，需要安装插件。

#### 临时方案：跳过角色同步

由于 Jenkins 的角色管理比较复杂，建议：

1. **暂时不使用"同步角色"功能**
2. 直接在"角色绑定"页面手动配置 Jenkins 角色映射

### 3. 实际配置示例

#### 步骤 1：编辑应用配置

进入"授权中心 > 应用管理"，编辑 Jenkins 应用：

```
基本信息:
  应用名称: Jenkins
  应用代码: jenkins
  应用类型: Jenkins
  Base URL: http://jenkins.hzmeipingmi.com  // 注意：末尾不要加斜杠

认证配置:
  认证方式: Basic Auth
  用户名: admin
  密码: ZRfX6x%dv@4pfKE&K8h  // 或使用 API Token

API 端点配置:
  创建用户: /securityRealm/createAccountByAdmin
  获取用户列表: /user/search   // 可留空
  获取单个用户: /user/{username}/api/json
  获取角色列表: /roleStrategy/roles  // 如果没有此端点，可留空
  分配角色: /roleStrategy/assignRoles  // 如果没有此端点，可留空
  撤销角色: /roleStrategy/unassignRoles
  获取用户角色: /roleStrategy/getRolesForUser
```

#### 步骤 2：手动配置角色绑定

1. 进入"授权中心 > 角色绑定"
2. 选择授权中心角色（如"开发人员"）
3. 手动绑定到 Jenkins 角色（如"developer"）

### 4. Jenkins 用户管理 API

#### 创建用户
```bash
# 方法 1: 使用 Jenkins CLI
java -jar jenkins-cli.jar -s http://jenkins.hzmeipingmi.com/ create-user username --username admin --password ZRfX6x%dv@4pfKE&K8h

# 方法 2: 通过 Groovy 脚本
# 需要 Script Console 权限
POST /scriptText
```

#### 获取用户列表
```bash
# 使用 Search API
GET /user/search?q={query}
```

## 完整配置流程

### 步骤 1：配置 Jenkins 应用

在"应用管理"页面：

1. **基本信息**
   - 应用名称：`Jenkins`
   - 应用代码：`jenkins`
   - 应用类型：`Jenkins`
   - Base URL：`http://jenkins.hzmeipingmi.com`（末尾不要 `/`）

2. **认证配置**
   - 认证方式：选择 `Basic Auth`
   - 用户名：`admin`
   - 密码：`ZRfX6x%dv@4pfKE&K8h`（或 API Token）

3. **API 端点**（暂时可留空或使用示例值）
   - 创建用户：`/securityRealm/createAccountByAdmin`
   - 获取用户列表：`/user/search`
   - 获取角色列表：留空（Jenkins 没有标准端点）

4. 保存配置

### 步骤 2：手动添加 Jenkins 角色

由于 Jenkins 没有 API 端点获取角色列表：

1. 进入"授权中心 > 角色绑定"
2. 点击"添加绑定"
3. 选择授权中心角色
4. 选择应用：Jenkins
5. 手动输入 Jenkins 角色代码（如 `admin`, `developer`, `reader`）

### 步骤 3：为用户授权

1. 进入"授权中心 > 用户授权"
2. 选择用户
3. 分配角色
4. 系统会在 Jenkins 中创建用户并分配权限

## 测试认证

使用 curl 测试 Jenkins API：

```bash
# 测试 Basic Auth 认证
curl -u admin:ZRfX6x%dv@4pfKE&K8h \
  http://jenkins.hzmeipingmi.com/user/admin/api/json

# 测试创建用户（需要 CSRF Token）
curl -u admin:ZRfX6x%dv@4pfKE&K8h \
  -X POST \
  http://jenkins.hzmeipingmi.com/securityRealm/createAccountByAdmin \
  -d "username=testuser&password1=testpass&password2=testpass&fullname=Test User"
```

## 常见问题

### Q: 为什么同步角色失败？
A: Jenkins 没有标准的角色管理 API 端点。建议手动配置角色绑定。

### Q: 为什么认证失败？
A:
1. 检查用户名密码是否正确
2. 检查用户是否有管理员权限
3. 尝试使用 API Token 而不是密码
4. 检查 Base URL 是否正确（不要末尾斜杠）

### Q: 如何生成 API Token？
A:
1. 登录 Jenkins
2. 点击右上角用户名 > 设置
3. API Token > 添加新 Token
4. 复制生成的 Token

### Q: 如何在 Jenkins 中管理角色？
A: 需要安装 "Role-based Authorization Strategy" 插件：
1. 系统管理 > 插件管理
2. 搜索并安装 "Role-based Authorization Strategy"
3. 系统管理 > 全局安全配置
4. 选择 "Role-Based Strategy"
5. 在 "Manage Roles" 中配置角色

## 推荐配置

由于 Jenkins API 的限制，推荐以下配置：

1. **不使用同步角色功能**：手动配置角色绑定
2. **不使用同步用户功能**：直接在用户授权时创建用户
3. **使用 Basic Auth**：使用 API Token 作为密码
4. **先测试 API**：使用 curl 测试所有端点是否可用

## 下一步

1. ✅ 修复了 URL 双斜杠问题
2. ✅ 支持 Basic Auth 认证
3. 🔄 测试 Jenkins 用户创建和授权功能
4. 🔄 根据实际情况调整 API 端点

请尝试：
1. 重新配置 Jenkins 应用（使用 Basic Auth）
2. 手动添加角色绑定
3. 测试用户授权功能
