# 授权中心数据库表结构说明

## 数据存储概览

同步的角色和用户数据存储在以下表中：

## 一、角色数据存储

### 表名：`application_roles`

**用途：** 存储从外部应用同步过来的角色信息

**表结构：**

```sql
CREATE TABLE application_roles (
  id          INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  app_id      INT UNSIGNED NOT NULL COMMENT '应用ID',
  role_code   VARCHAR(100) NOT NULL COMMENT '角色代码',
  role_name   VARCHAR(100) NOT NULL COMMENT '角色名称',
  role_type   VARCHAR(50) NOT NULL COMMENT '角色类型：global/project/agent',
  description VARCHAR(200) COMMENT '角色描述',
  sync_time   DATETIME COMMENT '同步时间',
  created_at  DATETIME,
  updated_at  DATETIME,
  INDEX idx_app_id (app_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='外部应用角色表';
```

**字段说明：**

| 字段 | 类型 | 说明 | 示例 |
|------|------|------|------|
| id | INT | 主键ID | 1 |
| app_id | INT | 外部应用ID | 1 (Jenkins) |
| role_code | VARCHAR(100) | 角色代码 | admin, developer |
| role_name | VARCHAR(100) | 角色名称 | Administrator, Developer |
| role_type | VARCHAR(50) | 角色类型 | global, project |
| description | VARCHAR(200) | 角色描述 | Global Role, Project Role |
| sync_time | DATETIME | 同步时间 | 2026-07-22 15:30:00 |

**数据示例：**

```sql
INSERT INTO application_roles VALUES
(1, 1, 'admin', 'admin', 'global', 'Global Role', '2026-07-22 15:30:00', NOW(), NOW()),
(2, 1, 'developer', 'developer', 'global', 'Global Role', '2026-07-22 15:30:00', NOW(), NOW()),
(3, 1, 'front-saas-prod', 'front-saas-prod', 'project', 'Project Role', '2026-07-22 15:30:00', NOW(), NOW());
```

**查询示例：**

```sql
-- 查询某个应用的所有角色
SELECT * FROM application_roles WHERE app_id = 1;

-- 查询全局角色
SELECT * FROM application_roles WHERE app_id = 1 AND role_type = 'global';

-- 查询项目角色
SELECT * FROM application_roles WHERE app_id = 1 AND role_type = 'project';

-- 统计角色数量
SELECT 
  role_type,
  COUNT(*) as count
FROM application_roles
WHERE app_id = 1
GROUP BY role_type;
```

## 二、用户数据存储

### 表名：`application_users`

**用途：** 存储从外部应用同步过来的用户信息

**表结构：**

```sql
CREATE TABLE application_users (
  id           INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  app_id       INT UNSIGNED NOT NULL COMMENT '应用ID',
  username     VARCHAR(100) NOT NULL COMMENT '用户名',
  display_name VARCHAR(100) COMMENT '显示名称',
  email        VARCHAR(100) COMMENT '邮箱',
  status       VARCHAR(20) DEFAULT 'active' COMMENT '状态',
  sync_time    DATETIME COMMENT '同步时间',
  created_at   DATETIME,
  updated_at   DATETIME,
  INDEX idx_app_id (app_id),
  INDEX idx_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='外部应用用户表';
```

**字段说明：**

| 字段 | 类型 | 说明 | 示例 |
|------|------|------|------|
| id | INT | 主键ID | 1 |
| app_id | INT | 外部应用ID | 1 (Jenkins) |
| username | VARCHAR(100) | 用户名 | zhangsan |
| display_name | VARCHAR(100) | 显示名称 | 张三 |
| email | VARCHAR(100) | 邮箱 | zhangsan@example.com |
| status | VARCHAR(20) | 状态 | active, inactive |
| sync_time | DATETIME | 同步时间 | 2026-07-22 15:30:00 |

**数据示例：**

```sql
INSERT INTO application_users VALUES
(1, 1, 'admin', '管理员', 'admin@example.com', 'active', '2026-07-22 15:30:00', NOW(), NOW()),
(2, 1, 'zhangsan', '张三', 'zhangsan@example.com', 'active', '2026-07-22 15:30:00', NOW(), NOW());
```

**查询示例：**

```sql
-- 查询某个应用的所有用户
SELECT * FROM application_users WHERE app_id = 1;

-- 查询活跃用户
SELECT * FROM application_users WHERE app_id = 1 AND status = 'active';

-- 按用户名搜索
SELECT * FROM application_users WHERE app_id = 1 AND username LIKE '%zhang%';
```

## 三、角色绑定数据存储

### 表名：`role_bindings`

**用途：** 存储授权中心角色与外部应用角色的绑定关系

**表结构：**

```sql
CREATE TABLE role_bindings (
  id                  INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  role_id             INT UNSIGNED NOT NULL COMMENT '授权中心角色ID',
  app_id              INT UNSIGNED NOT NULL COMMENT '外部应用ID',
  application_role_id INT UNSIGNED NOT NULL COMMENT '外部应用角色ID',
  created_at          DATETIME,
  updated_at          DATETIME,
  INDEX idx_role_id (role_id),
  INDEX idx_app_id (app_id),
  INDEX idx_application_role_id (application_role_id),
  FOREIGN KEY (role_id) REFERENCES auth_roles(id),
  FOREIGN KEY (app_id) REFERENCES applications(id),
  FOREIGN KEY (application_role_id) REFERENCES application_roles(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色绑定表';
```

**字段说明：**

| 字段 | 类型 | 说明 | 示例 |
|------|------|------|------|
| id | INT | 主键ID | 1 |
| role_id | INT | 授权中心角色ID | 1 (开发人员) |
| app_id | INT | 外部应用ID | 1 (Jenkins) |
| application_role_id | INT | 外部角色ID | 2 (developer) |

**数据示例：**

```sql
-- 授权中心角色"开发人员"绑定到 Jenkins 的 developer 角色
INSERT INTO role_bindings VALUES
(1, 1, 1, 2, NOW(), NOW());

-- 授权中心角色"开发人员"绑定到 Nacos 的 developer 角色
INSERT INTO role_bindings VALUES
(2, 1, 2, 5, NOW(), NOW());
```

**查询示例：**

```sql
-- 查询某个角色的所有绑定
SELECT 
  rb.id,
  ar.name as auth_role_name,
  app.name as app_name,
  appr.role_name as external_role_name,
  appr.role_type
FROM role_bindings rb
JOIN auth_roles ar ON rb.role_id = ar.id
JOIN applications app ON rb.app_id = app.id
JOIN application_roles appr ON rb.application_role_id = appr.id
WHERE rb.role_id = 1;

-- 查询某个应用的角色绑定情况
SELECT 
  ar.name as auth_role_name,
  appr.role_name as external_role_name
FROM role_bindings rb
JOIN auth_roles ar ON rb.role_id = ar.id
JOIN application_roles appr ON rb.application_role_id = appr.id
WHERE rb.app_id = 1;
```

## 四、用户角色分配数据存储

### 表名：`auth_user_roles`

**用途：** 存储授权中心用户与角色的分配关系

**表结构：**

```sql
CREATE TABLE auth_user_roles (
  id         INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id    INT UNSIGNED NOT NULL COMMENT '用户ID',
  role_id    INT UNSIGNED NOT NULL COMMENT '角色ID',
  granted_by VARCHAR(50) COMMENT '授权人',
  granted_at DATETIME COMMENT '授权时间',
  created_at DATETIME,
  updated_at DATETIME,
  UNIQUE KEY uk_user_role (user_id, role_id),
  INDEX idx_user_id (user_id),
  INDEX idx_role_id (role_id),
  FOREIGN KEY (user_id) REFERENCES auth_users(id),
  FOREIGN KEY (role_id) REFERENCES auth_roles(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户角色分配表';
```

**字段说明：**

| 字段 | 类型 | 说明 | 示例 |
|------|------|------|------|
| id | INT | 主键ID | 1 |
| user_id | INT | 用户ID | 1 (张三) |
| role_id | INT | 角色ID | 1 (开发人员) |
| granted_by | VARCHAR(50) | 授权人 | admin |
| granted_at | DATETIME | 授权时间 | 2026-07-22 15:30:00 |

**数据示例：**

```sql
-- 为张三分配"开发人员"角色
INSERT INTO auth_user_roles VALUES
(1, 1, 1, 'admin', '2026-07-22 15:30:00', NOW(), NOW());
```

**查询示例：**

```sql
-- 查询某个用户的所有角色
SELECT 
  au.username,
  au.nickname,
  ar.name as role_name,
  aur.granted_by,
  aur.granted_at
FROM auth_user_roles aur
JOIN auth_users au ON aur.user_id = au.id
JOIN auth_roles ar ON aur.role_id = ar.id
WHERE aur.user_id = 1;

-- 查询某个角色被分配给了哪些用户
SELECT 
  au.username,
  au.nickname,
  ar.name as role_name
FROM auth_user_roles aur
JOIN auth_users au ON aur.user_id = au.id
JOIN auth_roles ar ON aur.role_id = ar.id
WHERE aur.role_id = 1;
```

## 五、应用数据存储

### 表名：`applications`

**用途：** 存储外部应用的注册信息

**表结构：**

```sql
CREATE TABLE applications (
  id            INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name          VARCHAR(50) NOT NULL COMMENT '应用名称',
  code          VARCHAR(50) NOT NULL COMMENT '应用代码',
  type          VARCHAR(20) NOT NULL COMMENT '应用类型：jenkins/nacos/xxl-job',
  base_url      VARCHAR(255) NOT NULL COMMENT '应用访问地址',
  endpoints     TEXT COMMENT 'API端点配置(JSON)',
  auth_config   TEXT COMMENT '认证配置(JSON)',
  sync_interval INT DEFAULT 300 COMMENT '同步间隔(秒)',
  last_sync_time DATETIME COMMENT '最后同步时间',
  description   VARCHAR(200) COMMENT '描述',
  status        INT DEFAULT 1 COMMENT '状态：1启用 0禁用',
  created_at    DATETIME,
  updated_at    DATETIME,
  UNIQUE KEY uk_name (name),
  UNIQUE KEY uk_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='外部应用表';
```

**数据示例：**

```sql
INSERT INTO applications VALUES
(1, 'Jenkins', 'jenkins', 'jenkins', 'http://jenkins.hzmeipingmi.com', 
 '{"createUser":"/api/users","listUsers":"/api/users","getRoles":"/api/roles"}',
 '{"type":"basic","username":"admin","password":"xxx"}',
 300, '2026-07-22 15:30:00', 'CI/CD工具', 1, NOW(), NOW());
```

## 六、授权中心用户表

### 表名：`auth_users`

**用途：** 存储授权中心的用户信息

**重要字段：**
- `password`: 存储用户的统一初始密码（明文存储）
- 该密码用于所有外部系统的登录

**查询示例：**

```sql
-- 查看用户的统一密码
SELECT id, username, nickname, password FROM auth_users WHERE id = 1;
```

## 七、完整数据流程示例

### 示例：为张三分配"开发人员"角色

**步骤 1：同步外部角色**
```
Jenkins → application_roles 表
├── admin (global)
├── developer (global)
├── front-saas-prod (project)
└── ...
```

**步骤 2：创建角色绑定**
```
授权中心角色 "开发人员" (id=1, auth_roles)
  ↓
role_bindings 表
  ├── Jenkins → developer (id=1, role_bindings)
  └── Nacos → developer (id=2, role_bindings)
```

**步骤 3：分配角色给用户**
```
用户 "张三" (id=1, auth_users)
  ↓
auth_user_roles 表
  └── 分配角色 "开发人员" (id=1, auth_user_roles)
```

**步骤 4：自动在外部系统创建用户**
```
系统自动读取：
├── auth_users.password (统一密码)
├── role_bindings (角色绑定关系)
└── application_roles (外部角色)

系统自动执行：
├── 在 Jenkins 创建用户 zhangsan (使用统一密码)
├── 在 Jenkins 分配 developer 角色
├── 在 Nacos 创建用户 zhangsan (使用统一密码)
└── 在 Nacos 分配 developer 角色
```

## 八、数据关系图

```
auth_users (授权中心用户)
    ↓ 1:N
auth_user_roles (用户角色分配)
    ↓ N:1
auth_roles (授权中心角色)
    ↓ 1:N
role_bindings (角色绑定)
    ↓ N:1
application_roles (外部应用角色) ← 同步自 Jenkins
    ↓ N:1
applications (外部应用)
```

## 九、常用查询SQL

### 查看完整的授权链路

```sql
-- 从用户到外部角色的完整链路
SELECT 
  au.username as '用户名',
  au.nickname as '昵称',
  ar.name as '授权中心角色',
  app.name as '外部应用',
  appr.role_name as '外部角色',
  appr.role_type as '角色类型'
FROM auth_users au
JOIN auth_user_roles aur ON au.id = aur.user_id
JOIN auth_roles ar ON aur.role_id = ar.id
JOIN role_bindings rb ON ar.id = rb.role_id
JOIN applications app ON rb.app_id = app.id
JOIN application_roles appr ON rb.application_role_id = appr.id
WHERE au.username = 'zhangsan';
```

### 统计数据

```sql
-- 统计各应用的角色和用户数量
SELECT 
  app.name as '应用名称',
  COUNT(DISTINCT appr.id) as '角色数量',
  COUNT(DISTINCT appu.id) as '用户数量'
FROM applications app
LEFT JOIN application_roles appr ON app.id = appr.app_id
LEFT JOIN application_users appu ON app.id = appu.app_id
GROUP BY app.id;
```

## 十、数据清理

### 清理同步数据

```sql
-- 清理某个应用的角色数据
DELETE FROM application_roles WHERE app_id = 1;

-- 清理某个应用的用户数据
DELETE FROM application_users WHERE app_id = 1;
```

### 清理绑定关系

```sql
-- 清理某个角色的所有绑定
DELETE FROM role_bindings WHERE role_id = 1;

-- 清理某个应用的所有绑定
DELETE FROM role_bindings WHERE app_id = 1;
```

**注意：** 清理数据前请做好备份！

---

**总结：**
- **角色数据**：`application_roles` 表
- **用户数据**：`application_users` 表
- **绑定关系**：`role_bindings` 表
- **用户角色**：`auth_user_roles` 表
- **应用信息**：`applications` 表
- **统一密码**：`auth_users.password` 字段
