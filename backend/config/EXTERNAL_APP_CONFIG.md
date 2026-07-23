# 外部应用集成配置说明

## 概述

授权中心支持与外部应用（Jenkins、Nacos、XXL-Job、RocketMQ 等）集成，实现统一用户管理和权限控制。

## 配置步骤

### 1. 添加外部应用

在"授权中心 > 应用管理"页面添加外部应用，需要配置以下信息：

#### 基本信息
- **应用名称**：应用的显示名称（如：生产环境 Jenkins）
- **应用代码**：唯一标识符（如：jenkins-prod）
- **应用类型**：选择应用类型
- **Base URL**：应用的访问地址（如：https://jenkins.example.com）

#### 认证配置

##### API Token 认证（推荐）
```json
{
  "type": "token",
  "token": "your-api-token-here"
}
```

**Jenkins 配置示例：**
1. 登录 Jenkins
2. 进入"用户" > 选择用户 > "设置" > "API Token"
3. 生成新的 API Token
4. 将 Token 填入认证配置

##### Basic Auth 认证
```json
{
  "type": "basic",
  "username": "admin",
  "password": "your-password"
}
```

#### API 端点配置

根据外部应用的 API 文档配置相应的端点：

| 端点名称 | 说明 | 示例路径 | HTTP 方法 |
|---------|------|---------|----------|
| createUser | 创建用户 | /api/users | POST |
| listUsers | 获取用户列表 | /api/users | GET |
| getUser | 获取单个用户 | /api/users/{username} | GET |
| getRoles | 获取角色列表 | /api/roles | GET |
| assignRole | 为用户分配角色 | /api/users/{username}/roles | POST |
| revokeRole | 撤销用户角色 | /api/users/{username}/roles/{roleCode} | DELETE |
| getUserRoles | 获取用户角色 | /api/users/{username}/roles | GET |

**注意：** `listUsers` 和 `getRoles` 是同步用户和角色功能的必需端点。

### 2. 同步角色

添加应用后，点击"同步角色"按钮，系统会自动从外部应用获取所有角色信息。

**可能遇到的错误：**

#### 错误：缺少 getRoles 配置
```
同步角色失败: 缺少 getRoles 配置，请在应用配置中添加 getRoles 端点
```
**解决方法：** 在应用配置中添加 `getRoles` 端点路径。

#### 错误：认证失败(403)
```
同步角色失败: 认证失败(403)，请检查 Token 是否正确或是否有权限访问 API
```
**解决方法：**
1. 检查 API Token 是否正确
2. 检查 Token 是否有权限访问对应的 API
3. 确认 API 端点路径是否正确

#### 错误：未授权(401)
```
同步角色失败: 未授权(401)，请检查认证配置是否正确
```
**解决方法：**
1. 检查认证配置是否正确
2. 确认 Token 或用户名密码是否有效

### 3. 同步用户

点击"同步用户"按钮，系统会从外部应用获取所有用户信息。

**可能遇到的错误：**

#### 错误：缺少 listUsers 配置
```
同步用户失败: 缺少 listUsers 配置，请在应用配置中添加 listUsers 端点
```
**解决方法：** 在应用配置中添加 `listUsers` 端点路径。

## 应用配置示例

### Jenkins 配置示例

```
应用名称: 生产环境 Jenkins
应用代码: jenkins-prod
应用类型: Jenkins
Base URL: https://jenkins.example.com

认证配置:
  类型: API Token
  Token: 11a2b3c4d5e6f7g8h9i0j

API 端点:
  创建用户: /user/createUser
  获取用户列表: /user/search
  获取单个用户: /user/{username}
  获取角色列表: /roleStrategy/roles
  分配角色: /roleStrategy/assignRole
  撤销角色: /roleStrategy/unassignRole
  获取用户角色: /user/{username}/roles
```

### Nacos 配置示例

```
应用名称: Nacos 配置中心
应用代码: nacos-config
应用类型: Nacos
Base URL: https://nacos.example.com:8848/nacos

认证配置:
  类型: Basic Auth
  用户名: nacos
  密码: nacos

API 端点:
  创建用户: /v1/auth/users
  获取用户列表: /v1/auth/users
  获取单个用户: /v1/auth/users/{username}
  获取角色列表: /v1/auth/roles
  分配角色: /v1/auth/roles
  撤销角色: /v1/auth/roles/{role}
  获取用户角色: /v1/auth/roles?username={username}
```

## 使用流程

1. **添加应用**：在应用管理页面添加外部应用
2. **配置认证**：根据外部应用的认证方式配置 Token 或账号密码
3. **配置端点**：根据外部应用 API 文档配置各个端点
4. **同步角色**：点击"同步角色"获取外部应用的角色列表
5. **角色绑定**：在"角色绑定"中将授权中心角色映射到外部应用角色
6. **用户授权**：为用户分配角色，系统自动在外部应用创建用户并授权

## 常见问题

### Q: 同步角色失败，提示 403 错误？
A: 这通常是因为 Token 权限不足或 API 端点不正确。请检查：
1. Token 是否有管理员权限
2. API 端点路径是否正确
3. Base URL 是否正确

### Q: 同步用户失败，提示缺少 listUsers 配置？
A: 请在应用配置的 API 端点中添加 `listUsers` 端点。

### Q: 如何获取 Jenkins 的 API Token？
A:
1. 登录 Jenkins
2. 点击右上角用户名 > 设置
3. API Token 区域点击"添加新 Token"
4. 生成并复制 Token

### Q: 如何确认 API 端点是否正确？
A: 可以使用 curl 或 Postman 测试：
```bash
# 测试 Jenkins 角色列表 API
curl -H "Authorization: Bearer YOUR_TOKEN" \
  https://jenkins.example.com/roleStrategy/roles
```

## 技术支持

如遇到问题，请检查：
1. 应用配置是否正确
2. 网络是否可达（Base URL 是否可访问）
3. 认证信息是否有效
4. API 端点路径是否正确
5. 查看"操作日志"了解详细错误信息
