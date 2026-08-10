-- ============================================================
-- OneOps 表名重构迁移脚本
-- 功能:将旧表名统一改为模块前缀命名(sys_ / cmdb_ / auth_ / audit_ / k8s_ / mon_)
-- 日期: 2026-08-07
--
-- 使用方式:
--   方式A(推荐): 在新数据库 oneops 上直接启动后端, AutoMigrate 会自动创建新表名的表。
--                 然后用 INSERT ... SELECT 从 msre 导入数据(见脚本底部)。
--
--   方式B: 在旧数据库 msre 上执行本脚本的 RENAME TABLE 部分,
--           然后 mysqldump 导出并恢复到 oneops。
--
-- 注意: 部分表可能尚未创建(裸表),脚本使用存储过程安全跳过。
-- ============================================================

-- 安全 RENAME: 仅当源表存在且目标表不存在时执行
DELIMITER //
DROP PROCEDURE IF EXISTS safe_rename //
CREATE PROCEDURE safe_rename(IN old_name VARCHAR(200), IN new_name VARCHAR(200))
BEGIN
    DECLARE old_exists INT DEFAULT 0;
    DECLARE new_exists INT DEFAULT 0;
    SELECT COUNT(*) INTO old_exists FROM information_schema.tables
        WHERE table_schema = DATABASE() AND table_name = old_name;
    SELECT COUNT(*) INTO new_exists FROM information_schema.tables
        WHERE table_schema = DATABASE() AND table_name = new_name;
    IF old_exists = 1 AND new_exists = 0 THEN
        SET @sql = CONCAT('RENAME TABLE `', old_name, '` TO `', new_name, '`');
        PREPARE stmt FROM @sql;
        EXECUTE stmt;
        DEALLOCATE PREPARE stmt;
        SELECT CONCAT('  OK: ', old_name, ' -> ', new_name) AS result;
    ELSEIF old_exists = 0 THEN
        SELECT CONCAT('  SKIP: ', old_name, ' (不存在)') AS result;
    ELSE
        SELECT CONCAT('  SKIP: ', new_name, ' (已存在)') AS result;
    END IF;
END //
DELIMITER ;

-- ============================================================
-- 1. 系统管理 sys_ (10张)
-- ============================================================
CALL safe_rename('users',                    'sys_users');
CALL safe_rename('roles',                    'sys_roles');
CALL safe_rename('menus',                    'sys_menus');
CALL safe_rename('permissions',              'sys_permissions');
CALL safe_rename('role_permissions',         'sys_role_permissions');
CALL safe_rename('user_permissions',         'sys_user_permissions');
CALL safe_rename('permission_logs',          'sys_permission_logs');
CALL safe_rename('attribute_definitions',    'sys_attribute_definitions');
CALL safe_rename('api_resources',            'sys_api_resources');
CALL safe_rename('casbin_rule',              'sys_casbin_rule');

-- ============================================================
-- 2. 资产管理 cmdb_ (20张)
-- ============================================================
-- 基础资产
CALL safe_rename('business_units',           'cmdb_business_units');
CALL safe_rename('server_rooms',             'cmdb_server_rooms');
CALL safe_rename('cabinets',                 'cmdb_cabinets');
CALL safe_rename('servers',                  'cmdb_servers');
CALL safe_rename('server_tags',              'cmdb_server_tags');
CALL safe_rename('server_tag_relations',     'cmdb_server_tag_relations');
CALL safe_rename('asset_changes',            'cmdb_asset_changes');
CALL safe_rename('server_groups',            'cmdb_server_groups');
CALL safe_rename('server_group_relations',   'cmdb_server_group_relations');
CALL safe_rename('ssh_credentials',          'cmdb_ssh_credentials');
CALL safe_rename('cloud_servers',            'cmdb_cloud_servers');
CALL safe_rename('server_credentials',       'cmdb_server_credentials');
CALL safe_rename('server_attributes',        'cmdb_server_attributes');
-- Agent
CALL safe_rename('agent_versions',           'cmdb_agent_versions');
CALL safe_rename('agent_upgrade_tasks',      'cmdb_agent_upgrade_tasks');
-- 堡垒机
CALL safe_rename('asset_access_policies',    'cmdb_access_policies');
CALL safe_rename('bastion_sessions',         'cmdb_bastion_sessions');
CALL safe_rename('bastion_commands',         'cmdb_bastion_commands');
CALL safe_rename('bastion_file_transfers',   'cmdb_bastion_file_transfers');
CALL safe_rename('bastion_approvals',        'cmdb_bastion_approvals');

-- ============================================================
-- 3. 授权中心 auth_ (10张改名, 4张已合规)
-- ============================================================
-- 以下保持不变: auth_users, auth_groups, auth_user_groups, auth_group_permission_mappings
CALL safe_rename('applications',                     'auth_applications');
CALL safe_rename('application_roles',                'auth_application_roles');
CALL safe_rename('application_users',                'auth_application_users');
CALL safe_rename('application_groups',               'auth_application_groups');
CALL safe_rename('application_authorization_rules',  'auth_authorization_rules');
CALL safe_rename('application_operation_logs',       'auth_operation_logs');
CALL safe_rename('group_bindings',                   'auth_group_bindings');
CALL safe_rename('group_binding_executions',         'auth_group_binding_executions');
CALL safe_rename('user_identity_mappings',           'auth_user_identity_mappings');
CALL safe_rename('permission_assignment_statuses',   'auth_permission_assignment_statuses');

-- ============================================================
-- 4. 审计中心 audit_ (3张)
-- ============================================================
CALL safe_rename('login_logs',              'audit_login_logs');
CALL safe_rename('operation_logs',          'audit_operation_logs');
CALL safe_rename('system_event_logs',       'audit_system_event_logs');

-- ============================================================
-- 5. K8s管理 k8s_ (4张改名, 3张已合规)
-- ============================================================
-- 以下保持不变: k8s_clusters, k8s_sessions, k8s_commands
CALL safe_rename('cluster_role_bindings',   'k8s_cluster_role_bindings');
CALL safe_rename('diagnostic_history',      'k8s_diagnostic_history');
CALL safe_rename('diagnostic_config',       'k8s_diagnostic_config');
CALL safe_rename('diagnostic_permissions',  'k8s_diagnostic_permissions');

-- ============================================================
-- 6. 监控中心 mon_ (6张)
-- ============================================================
CALL safe_rename('agent_metrics',           'mon_agent_metrics');
CALL safe_rename('agent_metrics_archive',   'mon_agent_metrics_archive');
CALL safe_rename('agent_alerts',            'mon_agent_alerts');
CALL safe_rename('agent_alert_rules',       'mon_agent_alert_rules');
CALL safe_rename('notification_channels',   'mon_notification_channels');
CALL safe_rename('inspection_reports',      'mon_inspection_reports');

-- 清理存储过程
DROP PROCEDURE IF EXISTS safe_rename;

-- ============================================================
-- 附:从 msre 导入数据到 oneops (方式A)
-- 在 oneops 数据库中执行,前提: 后端已启动并 AutoMigrate 完成建表
-- ============================================================
-- INSERT INTO sys_users SELECT * FROM msre.sys_users;
-- INSERT INTO sys_roles SELECT * FROM msre.sys_roles;
-- INSERT INTO sys_menus SELECT * FROM msre.sys_menus;
-- INSERT INTO sys_permissions SELECT * FROM msre.sys_permissions;
-- INSERT INTO sys_role_permissions SELECT * FROM msre.sys_role_permissions;
-- INSERT INTO sys_casbin_rule SELECT * FROM msre.sys_casbin_rule;
-- INSERT INTO cmdb_servers SELECT * FROM msre.cmdb_servers;
-- INSERT INTO cmdb_business_units SELECT * FROM msre.cmdb_business_units;
-- INSERT INTO cmdb_ssh_credentials SELECT * FROM msre.cmdb_ssh_credentials;
-- INSERT INTO auth_applications SELECT * FROM msre.auth_applications;
-- INSERT INTO auth_users SELECT * FROM msre.auth_users;
-- INSERT INTO auth_groups SELECT * FROM msre.auth_groups;
-- ... 按需补充其他表
