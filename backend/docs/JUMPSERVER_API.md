# JumpServer API 端点说明

## 官方 API 端点

根据 JumpServer 源码 (`apps/users/urls/api_urls.py`) 分析：

### 用户管理 API

| 端点 | URL | 说明 |
|------|-----|------|
| 用户列表 | `/api/v1/users/users/` | GET - 获取用户列表 |
| 用户详情 | `/api/v1/users/users/{id}/` | GET - 获取单个用户详情 |
| 创建用户 | `/api/v1/users/users/` | POST - 创建新用户 |
| 更新用户 | `/api/v1/users/users/{id}/` | PUT/PATCH - 更新用户信息 |
| 删除用户 | `/api/v1/users/users/{id}/` | DELETE - 删除用户 |

### 用户组管理 API

| 端点 | URL | 说明 |
|------|-----|------|
| 用户组列表 | `/api/v1/users/groups/` | GET - 获取用户组列表 |
| 用户组详情 | `/api/v1/users/groups/{id}/` | GET - 获取单个用户组详情 |
| 创建用户组 | `/api/v1/users/groups/` | POST - 创建新用户组 |
| 更新用户组 | `/api/v1/users/groups/{id}/` | PUT/PATCH - 更新用户组信息 |
| 删除用户组 | `/api/v1/users/groups/{id}/` | DELETE - 删除用户组 |

### 其他重要 API 命名空间

根据 JumpServer 架构：

| 命名空间 | 说明 |
|---------|------|
| `/api/v1/users/` | 用户与用户组管理 |
| `/api/v1/assets/` | 资产清单与节点管理 |
| `/api/v1/accounts/` | 资产账号凭证管理 |
| `/api/v1/perms/` | 资产访问权限管理 |
| `/api/v1/authentication/` | 认证、MFA、令牌管理 |
| `/api/v1/orgs/` | 组织管理 |
| `/api/v1/settings/` | 系统与组织级设置 |

## 认证方式

JumpServer 支持 **三种 API 认证方式**，推荐使用 AccessKey 签名认证：

### 1. AccessKey + Secret 签名认证（推荐）⭐

**最安全可靠的方式**，使用 HTTP Signature 签名：

- **特点**：长期有效、签名认证、安全可靠
- **适用场景**：自动化集成、服务账号
- **认证方式**：HTTP Signature (HMAC-SHA256)

**获取方式**：
1. 登录 JumpServer 管理界面
2. 进入 **系统设置** → **账号管理** → **服务账号**
3. 创建服务账号或查看现有账号
4. 获取 **Access Key ID** 和 **Access Key Secret**

**配置示例**：
```json
{
  "authType": "accessKey",
  "accessKey": "your-access-key-id",
  "secret": "your-access-key-secret",
  "orgId": "00000000-00000000-00000000-000000000002"
}
```

**签名实现**：
```go
// 自动实现 HTTP Signature 签名
// 签名算法: HMAC-SHA256
// 签名头: (request-target), date
// 自动添加: Date, Accept, X-JMS-ORG headers
```

### 2. Private Token（简单易用）

**最简单的方式**，适合个人使用：

- **特点**：长期有效、配置简单
- **适用场景**：个人 API 调用
- **认证方式**：Bearer Token

**获取方式**：
1. 登录 JumpServer Web 界面
2. 点击 **右上角用户头像** → **个人中心**
3. 选择 **「API Token」** 标签页
4. 点击 **「创建 Token」** 或复制现有 Token

**配置示例**：
```json
{
  "authType": "token",
  "token": "your-private-token"
}
```

### 3. 用户名 + 密码（基础认证）

**传统方式**，不推荐：

- **特点**：简单但安全性较低
- **适用场景**：测试环境
- **认证方式**：HTTP Basic Auth

**配置示例**：
```json
{
  "authType": "basic",
  "username": "admin",
  "password": "your-password"
}
```

### 认证优先级

系统按以下顺序检查认证配置：

```
1. AccessKey + Secret（优先）
2. 用户名 + 密码
3. Private Token
```

### 请求示例

#### AccessKey 签名认证

```bash
# JumpServer 会自动处理签名，无需手动构造
# 系统会自动添加以下 Headers:
# - Date: Mon, 02 Jan 2006 15:04:05 GMT
# - Accept: application/json
# - X-JMS-ORG: 00000000-00000000-00000000-000000000002
# - Authorization: Signature keyId="...",algorithm="hmac-sha256",...
```

#### Private Token 认证

```bash
curl -X GET "https://jumpserver.example.com/api/v1/users/users/" \
  -H "Authorization: Bearer YOUR_PRIVATE_TOKEN" \
  -H "Content-Type: application/json"
```

#### Basic Auth 认证

```bash
curl -X GET "https://jumpserver.example.com/api/v1/users/users/" \
  -u "username:password" \
  -H "Content-Type: application/json"
```

## 响应格式

标准响应格式：

```json
{
  "id": "uuid",
  "name": "用户组名称",
  "comment": "描述信息",
  ...
}
```

## 注意事项

1. **角色管理**：JumpServer 不提供传统的角色列表 API，授权通过授权规则（AssetPermission）管理
2. **组织边界**：API 会自动过滤当前用户所属组织的数据
3. **用户组概念**：JumpServer 的用户组用于组织用户，权限通过授权规则分配
4. **Token 权限**：确保 API Token 有访问用户和用户组的权限

## 在 OneOps 中配置

### 应用配置

在 OneOps 应用管理中添加 JumpServer 应用：

- **应用名称**：JumpServer
- **应用类型**：jumpserver
- **Base URL**：`https://jumpserver.example.com`（不要包含 `/api/v1`）
- **API Token**：在认证配置中填写从 JumpServer 获取的 Private Token

### 端点配置

在端点配置中设置：

- **获取用户列表**: `/api/v1/users/users/`
- **获取用户组列表**: `/api/v1/users/groups/`
- **创建用户**: `/api/v1/users/users/`

### 完整配置示例

```json
{
  "name": "JumpServer",
  "code": "jumpserver",
  "type": "jumpserver",
  "baseUrl": "https://jumpserver.example.com",
  "endpoints": {
    "getUsers": "/api/v1/users/users/",
    "getGroups": "/api/v1/users/groups/",
    "createUser": "/api/v1/users/users/"
  },
  "authConfig": {
    "type": "token",
    "token": "your-private-token-here"
  },
  "syncInterval": 300
}
```

## 常见问题

### 1. 404 错误

**原因**：
- Base URL 配置错误
- 端点配置错误
- API Token 没有权限

**解决**：
- 确认 Base URL 格式正确（不包含 `/api/v1`）
- 检查端点配置是否正确
- 确认 Token 有访问权限

### 2. 401 认证失败

**原因**：
- API Token 无效或过期
- Token 格式错误

**解决**：
- 重新获取 Private Token
- 确认 Token 没有过期

### 3. 用户组列表为空

**原因**：
- JumpServer 版本不支持
- Token 所属用户不在任何用户组

**解决**：
- 确认 JumpServer 版本 ≥ 3.0
- 检查用户组成员关系

## 参考资料

- [JumpServer GitHub 仓库](https://github.com/jumpserver/jumpserver)
- [JumpServer API 路由源码](https://github.com/jumpserver/jumpserver/blob/master/apps/users/urls/api_urls.py)
- [JumpServer 认证机制](https://github.com/jumpserver/jumpserver/blob/master/apps/users/models/user/__init__.py)
