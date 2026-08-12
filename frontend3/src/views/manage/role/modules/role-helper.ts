/**
 * 角色辅助函数：删除检查、用户映射
 */
import { computed, ref } from 'vue';
import { fetchGetUserList } from '@/service/api';

/** 内置角色列表 */
const BUILTIN_ROLES: Record<string, string> = {
  admin: '超级管理员',
  ops: '运维工程师',
  auditor: '审计员',
  user: '普通用户',
  test: '测试角色'
};

/** 检查角色是否可以删除 */
export function canDeleteRole(
  role: Api.SystemManage.Role,
  roleUsersMap: Map<number, string[]>
): { canDelete: boolean; reason?: string } {
  if (BUILTIN_ROLES[role.code]) {
    return {
      canDelete: false,
      reason: `"${BUILTIN_ROLES[role.code]}" 是系统内置角色，不能删除`
    };
  }

  const users = roleUsersMap.get(role.id);
  if (users && users.length > 0) {
    return {
      canDelete: false,
      reason: `该角色已关联 ${users.length} 个用户：${users.slice(0, 3).join('、')}${users.length > 3 ? '...' : ''}。请先解除关联后再删除。`
    };
  }

  return { canDelete: true };
}

/** 创建角色-用户映射的 composable */
export function useRoleUsersMap() {
  const allUsers = ref<Api.SystemManage.User[]>([]);

  const roleUsersMap = computed(() => {
    const map = new Map<number, string[]>();
    if (allUsers.value && Array.isArray(allUsers.value)) {
      allUsers.value.forEach(user => {
        if (user && user.roleIds && user.roleIds.length > 0) {
          user.roleIds.forEach(roleId => {
            if (!map.has(roleId)) {
              map.set(roleId, []);
            }
            map.get(roleId)?.push(user.username);
          });
        }
      });
    }
    return map;
  });

  async function getAllUsers() {
    const { error, data } = await fetchGetUserList({ page: 1, pageSize: 1000 });
    if (!error && data) {
      allUsers.value = data.list || [];
    }
  }

  return {
    allUsers,
    roleUsersMap,
    getAllUsers
  };
}
