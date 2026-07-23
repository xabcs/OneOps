# 角色绑定和用户授权 - 完整使用指南

## 功能概述

授权中心提供两个核心功能：

1. **角色绑定**：将授权中心角色映射到外部应用角色
2. **用户授权**：为用户分配角色，自动在外部系统创建用户并授权

## 一、角色绑定

### 1.1 功能说明

**作用：** 建立授权中心角色与外部应用角色的映射关系

**示例：**
```
授权中心角色：开发人员
    ↓ 映射
Jenkins 角色：developer
Nacos 角色：developer
XXL-Job 角色：developer
```

### 1.2 操作步骤

#### 步骤 1：准备外部应用角色

1. 进入"授权中心 > 应用管理"
2. 找到目标应用（如 Jenkins）
3. 点击"同步角色"按钮
4. 确认角色同步成功
5. 点击"查看角色"确认角色列表

**预期结果：**
```
✅ 同步成功！同步了 3 个角色：
  - admin (管理员)
  - developer (开发者)
  - reader (只读用户)
```

#### 步骤 2：创建授权中心角色

1. 进入"授权中心 > 角色管理"
2. 点击"添加角色"
3. 填写角色信息：
   - 角色名称：开发人员
   - 角色代码：developer
   - 描述：开发团队成员角色
4. 保存

#### 步骤 3：创建角色绑定

1. 进入"授权中心 > 角色绑定"
2. 选择授权中心角色：开发人员
3. 点击"添加绑定"按钮
4. 在弹出对话框中：
   - 选择应用：Jenkins
   - 选择外部角色：developer
5. 点击"确定"

**预期结果：**
```
✅ 添加绑定成功

角色绑定列表：
| 应用名称 | 外部角色名称 | 外部角色代码 | 角色类型 |
|---------|-------------|-------------|---------|
| Jenkins | Developer   | developer   | global  |
```

#### 步骤 4：继续添加其他应用的绑定

重复步骤 3，为其他应用创建绑定：

```
开发人员角色绑定：
├── Jenkins → developer
├── Nacos → developer
└── XXL-Job → developer
```

### 1.3 数据流程

```
授权中心角色（开发人员）
     ↓
角色绑定表（role_bindings）
     ↓
┌─────────────────────────────┐
│ id: 1                       │
│ roleId: 1 (开发人员ID)      │
│ appId: 1 (Jenkins应用ID)    │
│ applicationRoleId: 2 (developer角色ID) │
└─────────────────────────────┘
     ↓
应用角色表（application_roles）
     ↓
外部应用角色：developer
```

### 1.4 界面示例

**角色绑定页面：**

```
┌──────────────────────────────────────────────┐
│ 请选择角色: [开发人员 ▼]  [添加绑定]        │
└──────────────────────────────────────────────┘

角色绑定列表：

| 序号 | 应用名称 | 外部角色名称 | 外部角色代码 | 角色类型 | 绑定时间 | 操作 |
|------|---------|-------------|-------------|---------|---------|------|
|  1   | Jenkins | Developer   | developer   | global  | 2026-07-22 | [删除] |
|  2   | Nacos   | Developer   | developer   | global  | 2026-07-22 | [删除] |
|  3   | XXL-Job | Developer   | developer   | global  | 2026-07-22 | [删除] |
```

## 二、用户授权

### 2.1 功能说明

**作用：**
- 为授权中心用户分配角色
- 自动在外部系统创建用户（如果不存在）
- 自动为外部用户分配对应的角色

**流程：**
```
授权中心用户：张三
     ↓ 分配角色
授权中心角色：开发人员
     ↓ 通过角色绑定
Jenkins 角色：developer
     ↓ 自动创建和授权
Jenkins 用户：张三 (角色: developer)
```

### 2.2 前提条件

1. ✅ 已创建授权中心用户
2. ✅ 已创建授权中心角色
3. ✅ 已同步外部应用角色
4. ✅ 已创建角色绑定

### 2.3 操作步骤

#### 步骤 1：创建授权中心用户

1. 进入"授权中心 > 用户管理"
2. 点击"添加用户"
3. 填写用户信息：
   - 用户名：zhangsan
   - 昵称：张三
   - 邮箱：zhangsan@example.com
4. 保存

**预期结果：**
```
✅ 用户创建成功！

显示初始密码对话框：
┌────────────────────────────────┐
│ 重要提示：初始密码仅显示一次   │
│                                │
│ 用户名: zhangsan                │
│ 初始密码: Abc123!@#             │
│ [复制密码]                      │
└────────────────────────────────┘
```

#### 步骤 2：为用户分配角色

1. 进入"授权中心 > 用户授权"
2. 选择用户：张三
3. 点击"分配角色"按钮
4. 选择角色：开发人员
5. 点击"确定"

#### 步骤 3：查看授权结果

**情况 A：全部成功**
```
✅ 授权成功！已在 3 个外部系统中授权

授权详情：
├── Jenkins: 已创建用户并分配 developer 角色
├── Nacos: 已创建用户并分配 developer 角色
└── XXL-Job: 已创建用户并分配 developer 角色
```

**情况 B：部分失败**
```
弹出授权结果对话框：

| 应用名称 | 用户名 | 分配角色 | 状态 | 备注 |
|---------|-------|---------|------|------|
| Jenkins | zhangsan | developer | ✅ 成功 | 已授权 |
| Nacos   | zhangsan | developer | ✅ 成功 | 已授权 |
| XXL-Job | zhangsan | developer | ❌ 失败 | 连接超时 |
```

### 2.4 授权过程详解

```
用户点击"分配角色"
     ↓
1. 检查用户是否已分配该角色
     ↓
2. 获取角色绑定的所有外部角色
   SELECT * FROM role_bindings WHERE role_id = ?
   结果：
   - Jenkins → developer
   - Nacos → developer
   - XXL-Job → developer
     ↓
3. 获取用户的统一密码
   用户密码在创建时已生成并存储
     ↓
4. 遍历每个应用，执行授权
   for each binding:
     a. 在外部系统创建用户
        POST /api/users
        {
          "username": "zhangsan",
          "password": "Abc123!@#",
          "email": "zhangsan@example.com"
        }

     b. 为用户分配角色
        POST /api/users/zhangsan/roles
        {
          "roleCode": "developer"
        }
     ↓
5. 保存授权记录
   INSERT INTO auth_user_roles (user_id, role_id, granted_by, granted_at)
   VALUES (1, 1, 'admin', NOW())
     ↓
6. 返回授权结果
```

### 2.5 界面示例

**用户授权页面：**

```
┌──────────────────────────────────────────────┐
│ 请选择用户: [张三 (zhangsan) ▼]  [分配角色] │
└──────────────────────────────────────────────┘

用户角色列表：

| 序号 | 角色名称 | 角色代码 | 授权人 | 授权时间 | 操作 |
|------|---------|---------|-------|---------|------|
|  1   | 开发人员 | developer | admin | 2026-07-22 14:30 | [删除] |
```

**授权结果对话框：**

```
┌──────────────────────────────────────────────┐
│ 授权结果                                     │
│                                              │
│ ℹ️ 用户已使用统一初始密码（创建用户时已显示）│
│    本次授权在外部系统中使用了该密码          │
├──────────────────────────────────────────────┤

| 应用名称 | 用户名 | 分配角色 | 状态 | 备注 |
|---------|-------|---------|------|------|
| Jenkins | zhangsan | developer | ✅ 成功 | 已授权 |
| Nacos   | zhangsan | developer | ✅ 成功 | 已授权 |

                                  [关闭]
└──────────────────────────────────────────────┘
```

## 三、完整示例：新员工入职授权流程

### 场景

新员工张三入职，需要在所有外部系统中创建账号并分配开发人员权限。

### 传统方式（手动操作）

```
1. 在 Jenkins 创建用户
2. 在 Jenkins 分配 developer 角色
3. 在 Nacos 创建用户
4. 在 Nacos 分配 developer 角色
5. 在 XXL-Job 创建用户
6. 在 XXL-Job 分配 developer 角色
...

问题：
❌ 重复操作多
❌ 容易遗漏
❌ 密码不统一
❌ 难以管理
```

### 使用授权中心（自动化）

```
步骤 1：创建用户（1次）
授权中心 > 用户管理 > 添加用户
   ↓ 自动生成统一密码

步骤 2：分配角色（1次）
授权中心 > 用户授权 > 选择用户 > 分配角色
   ↓ 自动在所有外部系统创建用户并授权

完成！
✅ 2步完成，自动同步到所有系统
✅ 统一密码管理
✅ 集中权限控制
```

### 详细操作

#### 1. 创建授权中心用户

```
用户名：zhangsan
昵称：张三
邮箱：zhangsan@example.com
电话：13800138000
描述：开发部员工

→ 系统自动生成密码：Abc123!@#
→ 显示初始密码对话框
→ 管理员记录密码并通知张三
```

#### 2. 查看角色绑定

```
授权中心角色：开发人员
已绑定外部角色：
├── Jenkins → developer
├── Nacos → developer
└── XXL-Job → developer
```

#### 3. 分配角色

```
选择用户：张三
选择角色：开发人员
点击"确定"

→ 系统自动：
  1. 在 Jenkins 创建用户 zhangsan (密码: Abc123!@#)
  2. 在 Jenkins 分配 developer 角色
  3. 在 Nacos 创建用户 zhangsan (密码: Abc123!@#)
  4. 在 Nacos 分配 developer 角色
  5. 在 XXL-Job 创建用户 zhangsan (密码: Abc123!@#)
  6. 在 XXL-Job 分配 developer 角色

→ 显示授权结果：
  ✅ 授权成功！已在 3 个外部系统中授权
```

#### 4. 用户登录

张三使用统一密码在所有外部系统登录：
- Jenkins：zhangsan / Abc123!@#
- Nacos：zhangsan / Abc123!@#
- XXL-Job：zhangsan / Abc123!@#

## 四、常见场景

### 场景 1：员工离职

```
操作：
1. 进入用户授权页面
2. 选择用户：张三
3. 删除所有角色
4. （可选）在用户管理页面禁用用户

结果：
✅ 自动从所有外部系统撤销权限
```

### 场景 2：员工晋升

```
当前状态：张三 - 开发人员

操作：
1. 进入用户授权页面
2. 选择用户：张三
3. 分配角色：技术主管

结果：
✅ 自动在所有外部系统添加技术主管权限
✅ 保留原有的开发人员权限
```

### 场景 3：新应用接入

```
新接入：RocketMQ 控制台

操作：
1. 应用管理 > 添加应用 > RocketMQ
2. 同步 RocketMQ 角色
3. 角色绑定 > 为"开发人员"角色添加 RocketMQ 绑定

结果：
✅ 所有已有"开发人员"角色的用户自动获得 RocketMQ 权限
```

## 五、数据库设计

### 5.1 角色绑定表（role_bindings）

```sql
CREATE TABLE role_bindings (
  id INT PRIMARY KEY AUTO_INCREMENT,
  role_id INT NOT NULL COMMENT '授权中心角色ID',
  app_id INT NOT NULL COMMENT '外部应用ID',
  application_role_id INT NOT NULL COMMENT '外部应用角色ID',
  created_at DATETIME,
  updated_at DATETIME,
  INDEX idx_role_id (role_id),
  INDEX idx_app_id (app_id)
);
```

### 5.2 用户角色表（auth_user_roles）

```sql
CREATE TABLE auth_user_roles (
  id INT PRIMARY KEY AUTO_INCREMENT,
  user_id INT NOT NULL COMMENT '用户ID',
  role_id INT NOT NULL COMMENT '角色ID',
  granted_by VARCHAR(50) COMMENT '授权人',
  granted_at DATETIME COMMENT '授权时间',
  created_at DATETIME,
  updated_at DATETIME,
  UNIQUE KEY uk_user_role (user_id, role_id)
);
```

## 六、API 接口

### 6.1 角色绑定相关

#### 创建角色绑定

```http
POST /api/system/roles/{roleId}/bindings
Content-Type: application/json

{
  "roleId": 1,
  "appId": 1,
  "applicationRoleId": 2
}

Response:
{
  "code": 200,
  "success": true,
  "data": {
    "id": 1,
    "roleId": 1,
    "appId": 1,
    "applicationRoleId": 2
  }
}
```

#### 获取角色绑定列表

```http
GET /api/system/roles/{roleId}/bindings

Response:
{
  "code": 200,
  "success": true,
  "data": [
    {
      "id": 1,
      "roleId": 1,
      "appId": 1,
      "appIDField": {
        "id": 1,
        "name": "Jenkins"
      },
      "applicationRoleId": 2,
      "applicationRole": {
        "id": 2,
        "roleCode": "developer",
        "roleName": "Developer"
      }
    }
  ]
}
```

### 6.2 用户授权相关

#### 为用户分配角色

```http
POST /api/system/users/assign-role
Content-Type: application/json

{
  "userId": 1,
  "roleId": 1
}

Response:
{
  "code": 200,
  "success": true,
  "data": {
    "message": "授权成功",
    "results": [
      {
        "appName": "Jenkins",
        "username": "zhangsan",
        "roleCode": "developer",
        "roleName": "Developer",
        "success": true
      }
    ]
  }
}
```

## 七、注意事项

### 7.1 密码管理

- ✅ 用户创建时自动生成统一密码
- ✅ 密码仅在创建时显示一次
- ✅ 可通过"查看密码"功能查看初始密码
- ⚠️ 外部系统使用相同的统一密码

### 7.2 权限控制

- ⚠️ 角色绑定需要先同步外部角色
- ⚠️ 用户授权需要先创建角色绑定
- ⚠️ 删除角色绑定不影响已授权的用户
- ✅ 删除用户角色会自动撤销外部权限

### 7.3 错误处理

**常见错误：**

1. **同步角色失败**
   - 检查外部应用连接
   - 检查认证配置
   - 查看操作日志

2. **授权失败**
   - 检查角色绑定是否正确
   - 检查外部应用 API
   - 查看授权结果对话框

## 八、最佳实践

### 8.1 角色设计

```
建议角色层次：

1. 管理员（admin）
   - 绑定：Jenkins → admin
   - 绑定：Nacos → admin
   - 绑定：XXL-Job → admin

2. 开发人员（developer）
   - 绑定：Jenkins → developer
   - 绑定：Nacos → developer
   - 绑定：XXL-Job → developer

3. 测试人员（tester）
   - 绑定：Jenkins → tester
   - 绑定：Nacos → viewer
   - 绑定：XXL-Job → viewer

4. 只读用户（viewer）
   - 绑定：Jenkins → reader
   - 绑定：Nacos → viewer
   - 绑定：XXL-Job → viewer
```

### 8.2 操作流程

```
标准流程：

1. 创建应用并同步角色
   └─> 应用管理 > 添加应用 > 同步角色

2. 创建授权中心角色
   └─> 角色管理 > 添加角色

3. 创建角色绑定
   └─> 角色绑定 > 添加绑定

4. 创建用户
   └─> 用户管理 > 添加用户

5. 分配角色
   └─> 用户授权 > 分配角色

6. 验证权限
   └─> 用户登录外部系统验证
```

## 九、故障排查

### 问题 1：同步角色失败

**检查步骤：**
1. 检查应用类型是否选择正确
2. 检查认证配置（Basic Auth）
3. 检查用户名密码是否正确
4. 检查网络连接
5. 查看操作日志

### 问题 2：授权失败

**检查步骤：**
1. 检查角色绑定是否创建
2. 检查外部应用 API 端点配置
3. 检查用户是否有密码
4. 查看授权结果对话框
5. 查看操作日志

### 问题 3：用户无法登录外部系统

**检查步骤：**
1. 确认用户已创建
2. 确认已分配角色
3. 确认授权结果是否成功
4. 检查外部系统用户是否存在
5. 尝试重置密码

## 十、总结

### 核心价值

1. **统一管理**：集中管理所有外部应用的用户和权限
2. **自动化**：一次操作，自动同步到所有系统
3. **规范化**：统一的角色体系和权限管理流程
4. **可追溯**：完整的操作日志和授权记录

### 关键概念

- **角色绑定**：授权中心角色 ↔ 外部应用角色
- **用户授权**：用户 + 角色 → 外部系统自动创建和授权
- **统一密码**：一个密码访问所有系统

### 使用要点

1. 先同步外部角色，再创建角色绑定
2. 先创建角色绑定，再为用户授权
3. 用户密码在创建时生成，统一使用
4. 删除用户角色会自动撤销外部权限
