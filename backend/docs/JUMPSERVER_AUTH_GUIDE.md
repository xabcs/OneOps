# JumpServer 认证配置指南

## 三种认证方式对比

| 认证方式 | 安全性 | 有效期 | 推荐度 | 获取难度 |
|---------|--------|--------|--------|----------|
| **AccessKey + Secret** | ⭐⭐⭐⭐⭐ | 永久 | ⭐⭐⭐⭐⭐ | 中等 |
| **Private Token** | ⭐⭐⭐⭐ | 永久 | ⭐⭐⭐⭐ | 简单 |
| **用户名密码** | ⭐⭐ | 永久 | ⭐⭐ | 简单 |

---

## 方式1：AccessKey + Secret 签名认证（推荐）⭐

### 为什么推荐？

- ✅ **最安全**：使用 HTTP Signature 签名算法
- ✅ **长期有效**：密钥不会过期
- ✅ **服务账号**：专为 API 集成设计
- ✅ **审计追踪**：可追踪到具体服务账号
- ✅ **权限控制**：可精确控制服务账号权限

### 如何获取

#### 步骤1：登录 JumpServer 管理界面

使用管理员账号登录 JumpServer

#### 步骤2：创建或查看服务账号

**路径**：系统设置 → 账号管理 → 服务账号

**如果没有服务账号**：

1. 点击「创建服务账号」
2. 填写基本信息：
   - 名称：`OneOps API集成`
   - 用户名：`oneops-api`
   - 组织：选择需要的组织
3. 点击「保存」

#### 步骤3：获取密钥

创建服务账号后，系统会显示：

```
Access Key ID:      AK-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
Access Key Secret:  SK-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
```

**⚠️ 重要提示**：
- **Access Key Secret 只显示一次**，请立即复制保存
- 如果忘记 Secret，需要重新生成新的密钥对

#### 步骤4：在 OneOps 中配置

在应用配置界面：

- **认证方式**：选择 `AccessKey + Secret（推荐）`
- **Access Key ID**：粘贴 Access Key ID
- **Access Key Secret**：粘贴 Access Key Secret
- **组织ID**（可选）：如果不填，使用默认组织

### 配置示例

```json
{
  "authConfig": {
    "authType": "accessKey",
    "accessKey": "AK-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    "secret": "SK-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    "orgId": "00000000-00000000-00000000-000000000002"
  }
}
```

### 技术细节

**签名过程**（自动完成）：

1. **设置 Headers**：
   ```
   Date: Mon, 02 Jan 2006 15:04:05 GMT
   Accept: application/json
   X-JMS-ORG: <组织ID>
   ```

2. **构造签名字符串**：
   ```
   (request-target): get /api/v1/users/users/
   date: Mon, 02 Jan 2006 15:04:05 GMT
   ```

3. **HMAC-SHA256 签名**：
   ```
   signature = Base64(HMAC-SHA256(secret, signingString))
   ```

4. **设置 Authorization**：
   ```
   Authorization: Signature keyId="<accessKey>",algorithm="hmac-sha256",headers="(request-target) date",signature="<base64-signature>"
   ```

---

## 方式2：Private Token（简单）

### 适合场景

- 个人使用
- 快速测试
- 临时集成

### 如何获取

#### 步骤1：登录 JumpServer Web 界面

使用您的个人账号登录

#### 步骤2：进入个人中心

点击 **右上角用户头像** → **个人中心**

#### 步骤3：获取 API Token

1. 选择 **「API Token」** 标签页
2. 点击 **「创建 Token」** 或复制现有 Token

### 配置示例

```json
{
  "authConfig": {
    "authType": "token",
    "token": "your-private-token-here"
  }
}
```

### 注意事项

- ✅ Token 长期有效
- ⚠️ 如果 Token 泄露，立即在个人中心删除并重新生成
- ⚠️ 使用个人账号 Token 时，API 权限等同于账号权限

---

## 方式3：用户名密码（不推荐）

### 为什么不推荐？

- ⚠️ **安全性较低**：密码明文传输
- ⚠️ **权限过大**：使用账号的所有权限
- ⚠️ **难以审计**：无法区分不同 API 调用来源

### 配置示例

```json
{
  "authConfig": {
    "authType": "basic",
    "username": "admin",
    "password": "your-password"
  }
}
```

### 适用场景

- 测试环境
- 内网环境
- 临时测试

---

## 常见问题

### Q1: Access Key Secret 忘记了怎么办？

**A**: Secret 只显示一次。如果忘记，需要：
1. 删除旧的 AccessKey
2. 重新创建新的 AccessKey
3. 使用新密钥重新配置

### Q2: 401 认证失败怎么办？

**排查步骤**：

1. **检查 AccessKey 和 Secret 是否正确**
   - 确认复制时没有多余空格
   - 确认 Access Key ID 和 Secret 对应正确

2. **检查服务账号权限**
   - 确认服务账号有所需权限
   - 确认服务账号状态为「启用」

3. **检查组织ID**
   - 确认服务账号属于正确的组织
   - 或在配置中指定正确的组织ID

4. **查看后端日志**
   ```
   INFO  使用 AccessKey 签名认证
   INFO  准备请求 JumpServer 用户列表 API
   ERROR JumpServer API 请求失败
   ```

### Q3: Private Token 过期了怎么办？

**A**: Private Token 不会自动过期。如果提示过期：
- 在个人中心重新生成 Token
- 更新 OneOps 配置

### Q4: 如何测试认证是否成功？

**方法1**：使用 curl 测试

```bash
# 测试 AccessKey 认证（需要手动签名，建议直接在 OneOps 中测试）
curl -X GET "https://jumpserver.example.com/api/v1/users/users/" \
  -H "Accept: application/json"

# 测试 Private Token
curl -X GET "https://jumpserver.example.com/api/v1/users/users/" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**方法2**：在 OneOps 中测试

1. 配置应用
2. 点击「同步用户」
3. 查看是否成功

### Q5: 哪种认证方式最安全？

**推荐顺序**：

1. **AccessKey + Secret**（最推荐）
   - 使用签名认证
   - 可追溯、可控制权限
   - 使用服务账号而非个人账号

2. **Private Token**
   - 简单易用
   - 长期有效
   - 使用个人账号权限

3. **用户名密码**（不推荐）
   - 安全性最低
   - 仅用于测试

---

## 最佳实践

### 1. 为不同用途创建不同的服务账号

```
oneops-sync-users    - 用于用户同步
oneops-sync-groups   - 用于用户组同步
oneops-create-users  - 用于创建用户
```

### 2. 定期轮换密钥

建议每 90 天轮换一次 AccessKey

### 3. 最小权限原则

为服务账号只授予必要的权限：
- 用户同步：只读权限
- 用户创建：用户管理权限
- 不需要管理员权限

### 4. 记录密钥用途

记录每个密钥的用途、创建时间、负责人

### 5. 监控 API 调用

通过 JumpServer 审计日志监控 API 调用情况

---

## 参考链接

- [JumpServer 官方文档](https://jumpserver.readthedocs.io/)
- [HTTP Signature 规范](https://tools.ietf.org/html/draft-cavage-http-signatures)
- [JumpServer GitHub](https://github.com/jumpserver/jumpserver)
