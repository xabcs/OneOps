# 完整清理 API权限管理菜单方案

## 🎯 问题原因

错误出现是因为：
1. ❌ 数据库中仍存在 `id=73` 的菜单
2. ❌ 用户角色关联了这个菜单
3. ❌ 前端加载菜单时包含了这个菜单项
4. ❌ 点击菜单时找不到对应路由

## ✅ 完整清理步骤

### 步骤1：数据库清理

**连接数据库**：
```bash
mysql -h 60.191.116.75 -P 38089 -u root -p ops
```

**执行清理 SQL**：
```sql
-- 1. 查看当前状态
SELECT id, name, path, status FROM menus WHERE id = 73;

-- 2. 删除角色菜单关联（关键！）
DELETE FROM role_menus WHERE menu_id = 73;

-- 3. 删除菜单项
DELETE FROM menus WHERE id = 73;

-- 4. 验证删除
SELECT COUNT(*) as menus_count FROM menus WHERE id = 73;
SELECT COUNT(*) as relations_count FROM role_menus WHERE menu_id = 73;
-- 预期结果：两个 COUNT 都应该是 0

-- 5. 确认系统管理下的菜单
SELECT id, name, path, status FROM menus WHERE parent_id = 6 ORDER BY sort;
-- 应该只有：用户管理、角色管理、菜单管理、权限管理
```

### 步骤2：清理前端缓存

**方法1：清除浏览器存储**
```javascript
// 在浏览器控制台执行
localStorage.clear();
sessionStorage.clear();
location.reload();
```

**方法2：手动清除**
1. 打开浏览器开发者工具（F12）
2. Application → Storage → Clear site data
3. 或直接清除 localStorage 中的菜单缓存

### 步骤3：重新登录

1. 退出登录
2. 清除浏览器缓存
3. 重新登录系统
4. 系统会从数据库重新加载菜单

## 📋 快速执行脚本

**完整的数据库清理脚本**：

```sql
-- ========================================
-- 清理废弃的 API权限管理菜单
-- ========================================

USE ops;

-- 开始事务
START TRANSACTION;

-- 1. 删除角色菜单关联
DELETE FROM role_menus WHERE menu_id = 73;
SELECT ROW_COUNT() as deleted_relations;

-- 2. 删除菜单项
DELETE FROM menus WHERE id = 73;
SELECT ROW_COUNT() as deleted_menus;

-- 提交事务
COMMIT;

-- 验证结果
SELECT '=== 清理结果 ===' as title;
SELECT '菜单记录' as item, COUNT(*) as count FROM menus WHERE id = 73
UNION ALL
SELECT '角色关联' as item, COUNT(*) as count FROM role_menus WHERE menu_id = 73;

-- 查看系统管理菜单
SELECT '=== 系统管理菜单 ===' as title;
SELECT id, name, path, sort, status FROM menus WHERE parent_id = 6 ORDER BY sort;
```

## 🔍 排查步骤

如果清理后仍有问题：

### 1. 检查数据库
```sql
-- 是否还有残留菜单
SELECT * FROM menus WHERE name LIKE '%API%';

-- 是否还有角色关联
SELECT rm.*, m.name as menu_name
FROM role_menus rm
LEFT JOIN menus m ON rm.menu_id = m.id
WHERE rm.menu_id = 73;
```

### 2. 检查前端缓存

**在浏览器控制台检查**：
```javascript
// 查看菜单数据
console.log(JSON.parse(localStorage.getItem('menuList') || '[]'));

// 查看用户权限
console.log(JSON.parse(localStorage.getItem('userInfo') || '{}'));
```

### 3. 检查网络请求

**查看菜单 API 返回**：
```
GET /api/v1/menus/tree
```

确认响应中不包含 `api-permission` 相关菜单。

## ✅ 验证成功的标志

1. ✅ 数据库查询返回 0 条记录
2. ✅ 前端菜单不显示 "API权限管理"
3. ✅ 点击其他菜单无路由错误
4. ✅ 控制台无报错信息

## 🎯 一键清理命令

```bash
# 连接数据库并执行清理
mysql -h 60.191.116.75 -P 38089 -u root -p ops << 'EOF'
DELETE FROM role_menus WHERE menu_id = 73;
DELETE FROM menus WHERE id = 73;
SELECT '菜单清理完成' as status;
EOF

# 然后在浏览器控制台执行
localStorage.clear();
sessionStorage.clear();
location.reload();
```

---

**关键点**：
- ✅ 必须删除 `role_menus` 关联
- ✅ 必须清除前端缓存
- ✅ 必须重新登录
