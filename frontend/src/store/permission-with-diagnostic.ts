// frontend/src/store/permission.ts
// 权限管理store - 添加诊断相关权限

import { computed, ref } from 'vue';
import { defineStore } from 'pinia';

export const usePermissionStore = defineStore('permission', () => {
  // 用户权限列表
  const permissions = ref<string[]>([]);

  // 加载用户权限
  const loadPermissions = async () => {
    try {
      // 从后端API获取用户权限
      const response = await fetch('/api/v1/user/permissions');
      const data = await response.json();

      if (data.success) {
        permissions.value = data.data.permissions || [];
      }
    } catch (error) {
      console.error('加载权限失败:', error);
      // 开发环境默认权限
      permissions.value = [
        'k8s:clusters:view',
        'k8s:pods:view',
        'k8s:diagnostic:execute', // 添加诊断权限
        'system:users:view',
        'audit:login:view'
      ];
    }
  };

  // 检查是否有指定权限
  const hasPermission = (permission: string): boolean => {
    // 超级管理员拥有所有权限
    if (permissions.value.includes('*:*:*')) {
      return true;
    }

    return permissions.value.includes(permission);
  };

  // 检查是否有任意一个权限
  const hasAnyPermission = (permissionList: string[]): boolean => {
    return permissionList.some(permission => hasPermission(permission));
  };

  // 获取所有权限
  const getAllPermissions = () => {
    return permissions.value;
  };

  // 权限列表定义（用于权限管理界面）
  const permissionDefinitions = [
    // K8s基础权限
    {
      id: 'k8s:clusters:view',
      name: '查看集群',
      category: 'K8s集群',
      description: '查看K8s集群列表和详情'
    },
    {
      id: 'k8s:nodes:view',
      name: '查看节点',
      category: 'K8s集群',
      description: '查看K8s节点信息'
    },
    {
      id: 'k8s:namespaces:view',
      name: '查看命名空间',
      category: 'K8s集群',
      description: '查看K8s命名空间'
    },
    {
      id: 'k8s:pods:view',
      name: '查看Pod',
      category: 'K8s集群',
      description: '查看Pod列表和详情'
    },
    {
      id: 'k8s:services:view',
      name: '查看服务',
      category: 'K8s集群',
      description: '查看Service服务'
    },
    {
      id: 'k8s:deployments:view',
      name: '查看部署',
      category: 'K8s集群',
      description: '查看Deployment部署'
    },
    // 新增：诊断相关权限
    {
      id: 'k8s:diagnostic:execute',
      name: '执行诊断',
      category: 'K8s诊断',
      description: '对Java应用执行Arthas诊断',
      riskLevel: 'high' // 高风险操作
    },
    {
      id: 'k8s:diagnostic:view',
      name: '查看诊断结果',
      category: 'K8s诊断',
      description: '查看诊断结果和历史记录'
    },
    // 监控权限
    {
      id: 'k8s:monitoring:view',
      name: '查看监控',
      category: 'K8s监控',
      description: '查看Pod和节点监控数据'
    },
    // 工具权限
    {
      id: 'k8s:tools:logs',
      name: '日志查询',
      category: 'K8s工具',
      description: '查询Pod日志'
    },
    {
      id: 'k8s:tools:events',
      name: '事件查询',
      category: 'K8s工具',
      description: '查询K8s事件'
    },
    {
      id: 'k8s:tools:yaml',
      name: 'YAML编辑',
      category: 'K8s工具',
      description: '使用YAML编辑器'
    },
    // 系统管理权限
    {
      id: 'system:users:view',
      name: '查看用户',
      category: '系统管理',
      description: '查看用户列表'
    },
    {
      id: 'system:users:edit',
      name: '编辑用户',
      category: '系统管理',
      description: '创建、编辑、删除用户'
    },
    {
      id: 'system:roles:view',
      name: '查看角色',
      category: '系统管理',
      description: '查看角色列表'
    },
    {
      id: 'system:roles:edit',
      name: '编辑角色',
      category: '系统管理',
      description: '创建、编辑、删除角色'
    },
    {
      id: 'system:menus:view',
      name: '查看菜单',
      category: '系统管理',
      description: '查看菜单配置'
    },
    {
      id: 'system:menus:edit',
      name: '编辑菜单',
      category: '系统管理',
      description: '编辑菜单配置'
    },
    // 审计权限
    {
      id: 'audit:login:view',
      name: '查看登录日志',
      category: '审计日志',
      description: '查看用户登录日志'
    },
    {
      id: 'audit:operation:view',
      name: '查看操作日志',
      category: '审计日志',
      description: '查看用户操作日志'
    },
    // CMDB权限
    {
      id: 'cmdb:servers:view',
      name: '查看服务器',
      category: 'CMDB',
      description: '查看CMDB服务器信息'
    },
    {
      id: 'cmdb:servers:edit',
      name: '编辑服务器',
      category: 'CMDB',
      description: '编辑CMDB服务器信息'
    }
  ];

  // 根据类别获取权限
  const getPermissionsByCategory = (category: string) => {
    return permissionDefinitions.filter(p => p.category === category);
  };

  // 获取所有权限类别
  const getCategories = () => {
    const categories = new Set(permissionDefinitions.map(p => p.category));
    return Array.from(categories);
  };

  return {
    permissions,
    loadPermissions,
    hasPermission,
    hasAnyPermission,
    getAllPermissions,
    permissionDefinitions,
    getPermissionsByCategory,
    getCategories
  };
});
