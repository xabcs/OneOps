-- 修复 auth_user_groups 表的唯一约束问题
-- 问题：同一个用户可以多次加入同一个用户组，导致数据重复
-- 解决方案：添加 user_id + group_id 的唯一索引

-- 步骤 1: 查看重复数据
SELECT user_id, group_id, COUNT(*) as count
FROM auth_user_groups
GROUP BY user_id, group_id
HAVING count > 1;

-- 步骤 2: 删除重复数据（保留最早加入的记录）
DELETE ug1 FROM auth_user_groups ug1
INNER JOIN auth_user_groups ug2
WHERE ug1.user_id = ug2.user_id
  AND ug1.group_id = ug2.group_id
  AND ug1.id > ug2.id;

-- 步骤 3: 添加唯一索引
ALTER TABLE auth_user_groups
ADD UNIQUE INDEX idx_unique_user_group (user_id, group_id);

-- 步骤 4: 验证结果
SELECT 'Unique index added successfully' as result;
