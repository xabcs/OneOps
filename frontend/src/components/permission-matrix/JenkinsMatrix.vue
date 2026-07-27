<script setup lang="ts">
import PermissionMatrix from './PermissionMatrix.vue';
import { computed } from 'vue';

interface Props {
  data: {
    roles: any[];
    users: any[];
    matrix: Record<number, Record<number, boolean>>;
    permissions_detail: Record<string, any>;
  };
  roleType?: 'all' | 'global' | 'project';
}

const props = withDefaults(defineProps<Props>(), {
  roleType: 'all'
});

const emit = defineEmits<{
  permissionChange: [userId: number, roleId: number, action: 'grant' | 'revoke']
}>();

// 安全访问数据的辅助函数
const safeData = computed(() => {
  return props.data || {
    roles: [],
    users: [],
    matrix: {},
    permissions_detail: {}
  };
});

// 根据 roleType 过滤角色
const filteredRoles = computed(() => {
  if (!safeData.value.roles || safeData.value.roles.length === 0) {
    return [];
  }

  if (props.roleType === 'all') {
    return safeData.value.roles;
  }
  return safeData.value.roles.filter(role => {
    if (props.roleType === 'global') {
      return role.roleType === 'global';
    } else if (props.roleType === 'project') {
      return role.roleType === 'project';
    }
    return true;
  });
});

// 为过滤后的角色创建子矩阵
const filteredMatrix = computed(() => {
  const matrix: Record<number, Record<number, boolean>> = {};
  if (!safeData.value.matrix || filteredRoles.value.length === 0) {
    return matrix;
  }

  for (const role of filteredRoles.value) {
    const roleId = Number(role.id);
    if (safeData.value.matrix[roleId]) {
      matrix[roleId] = safeData.value.matrix[roleId];
    }
  }
  console.log('filteredMatrix:', {
    roleType: props.roleType,
    filteredRolesCount: filteredRoles.value.length,
    matrixKeys: Object.keys(matrix),
    sampleData: matrix
  });
  return matrix;
});

// 统计信息
const globalRolesCount = computed(() => {
  if (!safeData.value.roles) return 0;
  return safeData.value.roles.filter(role => role.roleType === 'global').length;
});

const projectRolesCount = computed(() => {
  if (!safeData.value.roles) return 0;
  return safeData.value.roles.filter(role => role.roleType === 'project').length;
});

function handleCellClick(user: any, role: any, hasPermission: boolean) {
  if (!user || !role) return;

  const userId = user.id;
  const roleId = role.id;
  const action = hasPermission ? 'revoke' : 'grant';

  emit('permissionChange', userId, roleId, action);
}

function getPermissionDetail(userId: number, roleId: number) {
  if (!safeData.value.permissions_detail) return null;

  const key = `${roleId}_${userId}`;
  return safeData.value.permissions_detail[key] || null;
}
</script>

<template>
  <div class="jenkins-matrix space-y-4">
    <!-- 统计信息 -->
    <div class="mb-4 flex items-center justify-between">
      <div class="flex items-center gap-2">
        <ElTag type="primary">Jenkins 权限矩阵</ElTag>
        <ElTag v-if="roleType === 'all' || roleType === 'global'" type="success">
          Global 角色: {{ globalRolesCount }}
        </ElTag>
        <ElTag v-if="roleType === 'all' || roleType === 'project'" type="warning">
          Project 角色: {{ projectRolesCount }}
        </ElTag>
        <ElTag>用户: {{ safeData.users?.length || 0 }}</ElTag>
      </div>
      <div class="flex items-center gap-2 text-sm text-gray-500">
        <span>点击单元格可查看详情</span>
      </div>
    </div>

    <!-- 角色类型说明 -->
    <div v-if="roleType === 'all'" class="rounded bg-blue-50 p-3 text-sm">
      <p class="font-medium text-blue-900">角色类型说明：</p>
      <ul class="ml-4 mt-2 space-y-1 text-blue-800">
        <li>• <strong>Global 角色</strong>：全局角色，适用于整个 Jenkins 实例（如 admin、manager）</li>
        <li>• <strong>Project 角色</strong>：项目角色，用于控制特定 Job/Folder 的访问权限</li>
      </ul>
    </div>
    <div v-else-if="roleType === 'global'" class="rounded bg-green-50 p-3 text-sm">
      <p class="font-medium text-green-900">Global 角色：适用于整个 Jenkins 实例的管理权限</p>
    </div>
    <div v-else-if="roleType === 'project'" class="rounded bg-orange-50 p-3 text-sm">
      <p class="font-medium text-orange-900">Project 角色：用于控制特定项目的构建和查看权限</p>
    </div>

    <!-- 权限矩阵 -->
    <PermissionMatrix
      v-if="safeData.roles?.length > 0 && safeData.users?.length > 0 && filteredRoles.length > 0"
      :rows="safeData.users"
      :columns="filteredRoles"
      :matrix="filteredMatrix"
      row-key="id"
      col-key="id"
      row-label="username"
      col-label="roleName"
      :permissions-detail="safeData.permissions_detail || {}"
      @cell-click="handleCellClick"
    />

    <!-- 空状态 -->
    <ElEmpty
      v-else-if="!safeData.roles || safeData.roles.length === 0"
      description="暂无角色数据"
    />

    <ElEmpty
      v-else-if="!safeData.users || safeData.users.length === 0"
      description="暂无用户数据"
    />

    <!-- 提示信息 -->
    <div v-if="safeData.roles?.length > 0" class="mt-4 text-sm text-gray-500">
      <p>提示：绿色 ✓ 表示有权限，灰色 ✗ 表示无权限</p>
    </div>
  </div>
</template>

<style scoped>
.jenkins-matrix {
  @apply flex flex-col gap-4;
}
</style>
