<script setup lang="ts">
import { computed } from 'vue';
import PermissionMatrix from './PermissionMatrix.vue';

interface Column {
  id: number | string;
  name: string;
  type?: string;
  [key: string]: any;
}

interface User {
  id: number;
  username: string;
  nickname?: string;
  external_username?: string;
  external_user_id?: string;
}

interface Props {
  data: {
    columns: Column[];
    users: User[];
    matrix: Record<string, Record<string, boolean>>;
    permissions_detail: Record<string, any>;
    message?: string;
  };
  itemType?: 'all' | string;
  appType?: string;
  // 配置不同应用类型的显示
  config?: {
    columnTypeName?: string; // "角色" | "授权规则"
    typeOptions?: Array<{ value: string; label: string; description?: string }>;
    descriptions?: {
      all?: string;
      [key: string]: string;
    };
    syncAction?: boolean; // 是否显示同步按钮
  };
}

const props = withDefaults(defineProps<Props>(), {
  itemType: 'all',
  appType: 'jenkins',
  config: () => ({})
});

const emit = defineEmits<{
  permissionChange: [userId: number, columnId: number | string, action: 'grant' | 'revoke'];
  refresh: [];
}>();

// 默认配置
const defaultConfig = computed(() => {
  const configs: Record<string, any> = {
    jenkins: {
      columnTypeName: '角色',
      typeOptions: [
        { value: 'global', label: 'Global 角色', description: '全局角色，适用于整个 Jenkins 实例的管理权限' },
        { value: 'project', label: 'Project 角色', description: '项目角色，用于控制特定项目的构建和查看权限' }
      ],
      descriptions: {
        all: '角色类型说明：\n• Global 角色：全局角色，适用于整个 Jenkins 实例（如 admin、manager）\n• Project 角色：项目角色，用于控制特定 Job/Folder 的访问权限',
        global: 'Global 角色：适用于整个 Jenkins 实例的管理权限',
        project: 'Project 角色：用于控制特定项目的构建和查看权限'
      },
      syncAction: false
    },
    gitlab: {
      columnTypeName: '角色',
      typeOptions: [
        { value: 'global', label: 'Global 角色', description: '全局角色，适用于整个 GitLab 实例的管理权限' },
        { value: 'project', label: 'Project 角色', description: '项目角色，用于控制特定项目的访问权限' }
      ],
      descriptions: {
        all: '角色类型说明：\n• Global 角色：全局角色，适用于整个 GitLab 实例（如 admin、maintainer）\n• Project 角色：项目角色，用于控制特定项目的访问权限',
        global: 'Global 角色：适用于整个 GitLab 实例的管理权限',
        project: 'Project 角色：用于控制特定项目的访问权限'
      },
      syncAction: false
    },
    jumpserver: {
      columnTypeName: '授权规则',
      typeOptions: [
        { value: 'user', label: '用户规则', description: '直接为特定用户授权的规则' },
        { value: 'group', label: '用户组规则', description: '为用户组授权，组成员自动继承权限' }
      ],
      descriptions: {
        all: '授权规则类型说明：\n• 用户规则：直接为特定用户授权的规则\n• 用户组规则：为用户组授权，组成员自动继承权限\n\n授权对象可以是：资产（单个服务器）、节点（资产分组）、或全部资产',
        user: '用户规则：直接为特定用户授权访问资产',
        group: '用户组规则：为用户组授权，组成员自动继承权限'
      },
      syncAction: true
    }
  };

  return { ...configs[props.appType], ...props.config };
});

// 安全访问数据
const safeData = computed(() => {
  return (
    props.data || {
      columns: [],
      users: [],
      matrix: {},
      permissions_detail: {},
      message: ''
    }
  );
});

// 根据 itemType 过滤列
const filteredColumns = computed(() => {
  if (!safeData.value.columns || safeData.value.columns.length === 0) {
    return [];
  }

  if (props.itemType === 'all') {
    return safeData.value.columns;
  }

  return safeData.value.columns.filter((col: Column) => {
    // 支持 type 或 roleType 字段
    const colType = col.type || col.roleType || '';
    return colType === props.itemType;
  });
});

// 为过滤后的列创建子矩阵
const filteredMatrix = computed(() => {
  const matrix: Record<string, Record<string, boolean>> = {};

  if (!safeData.value.matrix || filteredColumns.value.length === 0) {
    return matrix;
  }

  for (const col of filteredColumns.value) {
    const colId = String(col.id);
    if (safeData.value.matrix[colId]) {
      matrix[colId] = safeData.value.matrix[colId];
    }
  }

  return matrix;
});

// 统计各类型的列数量
const typeCounts = computed(() => {
  const counts: Record<string, number> = {};

  if (!safeData.value.columns) {
    return counts;
  }

  for (const typeOption of defaultConfig.value.typeOptions || []) {
    const typeValue = typeOption.value;
    counts[typeValue] = safeData.value.columns.filter((col: Column) => {
      const colType = col.type || col.roleType || '';
      return colType === typeValue;
    }).length;
  }

  return counts;
});

function handleCellClick(user: User, column: Column, hasPermission: boolean) {
  if (!user || !column) return;

  const userId = user.id;
  const columnId = column.id;
  const action = hasPermission ? 'revoke' : 'grant';

  emit('permissionChange', userId, columnId, action);
}

// 获取当前类型的描述信息
const currentDescription = computed(() => {
  const descriptions = defaultConfig.value.descriptions || {};
  return descriptions[props.itemType] || '';
});

// 获取背景颜色样式
const getBgClass = (type: string) => {
  const bgMap: Record<string, string> = {
    global: 'bg-green-50',
    project: 'bg-orange-50',
    user: 'bg-green-50',
    group: 'bg-orange-50',
    all: 'bg-blue-50'
  };
  return bgMap[type] || 'bg-blue-50';
};

// 获取文本颜色样式
const getTextClass = (type: string) => {
  const textMap: Record<string, string> = {
    global: 'text-green-900',
    project: 'text-orange-900',
    user: 'text-green-900',
    group: 'text-orange-900',
    all: 'text-blue-900'
  };
  return textMap[type] || 'text-blue-900';
};

const getListTextClass = (type: string) => {
  const textMap: Record<string, string> = {
    global: 'text-green-800',
    project: 'text-orange-800',
    user: 'text-green-800',
    group: 'text-orange-800',
    all: 'text-blue-800'
  };
  return textMap[type] || 'text-blue-800';
};

// 获取当前类型的中文名称
const getDefaultConfigTypeLabel = (type: string) => {
  const typeOption = defaultConfig.value.typeOptions?.find((opt: any) => opt.value === type);
  return typeOption?.label || type;
};
</script>

<template>
  <div class="app-permission-matrix space-y-4">
    <!-- 统计信息 -->
    <div class="mb-4 flex items-center justify-between">
      <div class="flex items-center gap-2">
        <ElTag type="primary">{{ appType.toUpperCase() }} 权限矩阵</ElTag>

        <template v-if="itemType === 'all'">
          <ElTag
            v-for="typeOption in defaultConfig.typeOptions"
            :key="typeOption.value"
            :type="typeOption.value === 'global' || typeOption.value === 'user' ? 'success' : 'warning'"
          >
            {{ typeOption.label }}: {{ typeCounts[typeOption.value] || 0 }}
          </ElTag>
        </template>

        <ElTag>用户: {{ safeData.users?.length || 0 }}</ElTag>
      </div>
      <div class="flex items-center gap-2 text-sm text-gray-500">
        <span>点击单元格可查看详情</span>
      </div>
    </div>

    <!-- 类型说明 -->
    <div v-if="currentDescription" :class="`rounded ${getBgClass(itemType)} p-3 text-sm`">
      <p :class="`font-medium ${getTextClass(itemType)}`">
        {{ itemType === 'all' ? `${defaultConfig.columnTypeName}类型说明：` : '' }}
      </p>
      <ul v-if="itemType === 'all'" :class="`ml-4 mt-2 space-y-1 ${getListTextClass('all')}`">
        <li v-for="typeOption in defaultConfig.typeOptions" :key="typeOption.value">
          <strong>{{ typeOption.label }}</strong>
          ：{{ typeOption.description }}
        </li>
      </ul>
      <p v-else :class="`mt-2 ${getListTextClass(itemType)}`">
        {{ currentDescription }}
      </p>
    </div>

    <!-- 消息提示 -->
    <ElAlert v-if="safeData.message" type="info" :closable="false" show-icon class="mb-4">
      {{ safeData.message }}
    </ElAlert>

    <!-- 权限矩阵 -->
    <PermissionMatrix
      v-if="safeData.columns?.length > 0 && safeData.users?.length > 0 && filteredColumns.length > 0"
      :rows="safeData.users"
      :columns="filteredColumns"
      :matrix="filteredMatrix"
      row-key="id"
      col-key="id"
      row-label="username"
      col-label="name"
      :permissions-detail="safeData.permissions_detail || {}"
      @cell-click="handleCellClick"
    />

    <!-- 空状态：过滤后的列为空 -->
    <ElEmpty
      v-else-if="safeData.columns?.length > 0 && filteredColumns.length === 0"
      :description="`暂无${itemType === 'all' ? '' : getDefaultConfigTypeLabel(itemType)}${defaultConfig.columnTypeName}数据`"
    >
      <ElButton v-if="defaultConfig.syncAction" type="primary" @click="$emit('refresh')">
        同步{{ defaultConfig.columnTypeName }}
      </ElButton>
    </ElEmpty>

    <!-- 空状态：完全没有列数据 -->
    <ElEmpty
      v-else-if="!safeData.columns || safeData.columns.length === 0"
      :description="`暂无${defaultConfig.columnTypeName}数据`"
    >
      <ElButton v-if="defaultConfig.syncAction" type="primary" @click="$emit('refresh')">
        同步{{ defaultConfig.columnTypeName }}
      </ElButton>
    </ElEmpty>

    <ElEmpty v-else-if="!safeData.users || safeData.users.length === 0" description="暂无用户数据" />

    <!-- 提示信息 -->
    <div v-if="safeData.columns?.length > 0" class="mt-4 text-sm text-gray-500">
      <p>提示：绿色 ✓ 表示有权限，灰色 ✗ 表示无权限</p>
      <p v-if="appType === 'jumpserver'" class="mt-1">授权规则来自 Jumpserver，通过用户组绑定自动管理用户权限</p>
    </div>
  </div>
</template>

<style scoped>
.app-permission-matrix {
  @apply flex flex-col gap-4;
}
</style>
