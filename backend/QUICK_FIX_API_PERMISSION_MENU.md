# 快速修复：删除废弃的API权限管理菜单

## 🎯 直接执行 SQL

连接到数据库后，执行以下 SQL：

```sql
-- 1. 删除菜单项
DELETE FROM menus WHERE id = 73 AND name = 'API权限管理';

-- 2. 删除角色菜单关联
DELETE FROM role_menus WHERE menu_id = 73;

-- 3. 验证删除结果
SELECT COUNT(*) as remaining FROM menus WHERE name = 'API权限管理';
-- 预期结果：remaining = 0
```

## 🔧 连接数据库命令

```bash
# MySQL 客户端连接
mysql -h 60.191.116.75 -P 38089 -u root -p ops

# 或使用配置文件中的连接信息
# 数据库：ops
# 主机：60.191.116.75
# 端口：38089
```

## ✅ 执行后的验证

1. **刷新前端页面**：清除浏览器缓存后重新登录
2. **检查菜单**：确认不再出现 "API权限管理" 菜单
3. **验证路由**：确认无路由错误

---

**预期结果**：错误消失，菜单正常显示。
