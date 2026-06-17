-- 更新 K8s 菜单名称为中文
-- 根据 PRD 需求，使用标准的 Kubernetes 资源中文翻译

-- 更新 Deployments → 工作负载
UPDATE menus SET name = '工作负载' WHERE name = 'Deployments' AND path = '/k8s/resources/deployments';

-- 更新 Pods → 容器组
UPDATE menus SET name = '容器组' WHERE name = 'Pods' AND path = '/k8s/resources/pods';

-- 更新 Services → 服务
UPDATE menus SET name = '服务' WHERE name = 'Services' AND path = '/k8s/resources/services';

-- 更新 ConfigMaps → 配置项
UPDATE menus SET name = '配置项' WHERE name = 'ConfigMaps' AND path = '/k8s/resources/configmaps';

-- 更新 Secrets → 保密字典
UPDATE menus SET name = '保密字典' WHERE name = 'Secrets' AND path = '/k8s/resources/secrets';

-- 验证更新结果
SELECT id, name, path, parent_id, sort FROM menus WHERE path LIKE '/k8s%' ORDER BY sort;
