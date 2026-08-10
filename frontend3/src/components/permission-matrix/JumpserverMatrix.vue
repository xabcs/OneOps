<script setup lang="ts">
import { computed } from 'vue';
import PermissionMatrix from './PermissionMatrix.vue';

interface Rule {
  id: number;
  rule_id: string;
  rule_name: string;
  rule_type?: string;
  subject_type?: string;
  subject_name?: string;
  object_type?: string;
  object_name?: string;
  user_count?: number;
  is_enabled?: boolean;
  is_expired?: boolean;
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
    rules: Rule[];
    users: User[];
    matrix: Record<string, Record<string, boolean>>;
    permissions_detail: Record<string, any>;
    message?: string;
  };
  ruleType?: 'all' | 'user' | 'group';
}

const props = withDefaults(defineProps<Props>(), {
  ruleType: 'all'
});

const emit = defineEmits<{
  permissionChange: [userId: number, ruleId: string, action: 'grant' | 'revoke'];
}>();

// 安全访问数据的辅助函数
const safeData = computed(() => {
  return (
    props.data || {
      rules: [],
      users: [],
      matrix: {},
      permissions_detail: {},
      message: ''
    }
  );
});

// 根据 ruleType 过滤规则
const filteredRules = computed(() => {
  if (!safeData.value.rules || safeData.value.rules.length === 0) {
    return [];
  }

  if (props.ruleType === 'all') {
    return safeData.value.rules;
  }
  return safeData.value.rules.filter(rule => {
    if (props.ruleType === 'user') {
      return rule.subject_type === 'user';
    } else if (props.ruleType === 'group') {
      return rule.subject_type === 'group';
    }
    return true;
  });
});

// 为过滤后的规则创建子矩阵
const filteredMatrix = computed(() => {
  const matrix: Record<string, Record<string, boolean>> = {};
  if (!safeData.value.matrix || filteredRules.value.length === 0) {
    return matrix;
  }

  for (const rule of filteredRules.value) {
    const ruleId = rule.rule_id;
    if (safeData.value.matrix[ruleId]) {
      matrix[ruleId] = safeData.value.matrix[ruleId];
    }
  }
  return matrix;
});

// 统计信息
const userRulesCount = computed(() => {
  if (!safeData.value.rules) return 0;
  return safeData.value.rules.filter(rule => rule.subject_type === 'user').length;
});

const groupRulesCount = computed(() => {
  if (!safeData.value.rules) return 0;
  return safeData.value.rules.filter(rule => rule.subject_type === 'group').length;
});

const enabledRulesCount = computed(() => {
  if (!safeData.value.rules) return 0;
  return safeData.value.rules.filter(rule => rule.is_enabled).length;
});

function handleCellClick(user: User, rule: Rule, hasPermission: boolean) {
  if (!user || !rule) return;

  const userId = user.id;
  const ruleId = rule.rule_id;
  const action = hasPermission ? 'revoke' : 'grant';

  emit('permissionChange', userId, ruleId, action);
}

function getPermissionDetail(userId: number, ruleId: string) {
  if (!safeData.value.permissions_detail) return null;

  const key = `${ruleId}_${userId}`;
  return safeData.value.permissions_detail[key] || null;
}

// 格式化规则显示名称
function formatRuleLabel(rule: Rule): string {
  if (!rule) return '-';
  return rule.rule_name || rule.rule_id || '-';
}

// 获取规则的额外信息
function getRuleMeta(rule: Rule): string {
  if (!rule) return '';

  const parts: string[] = [];

  if (rule.user_count !== undefined) {
    parts.push(`${rule.user_count} 用户`);
  }

  if (rule.object_name) {
    parts.push(rule.object_name);
  }

  return parts.length > 0 ? `(${parts.join(' | ')})` : '';
}
</script>

<template>
  <div class="jumpserver-matrix space-y-4">
    <!-- 统计信息 -->
    <div class="mb-4 flex items-center justify-between">
      <div class="flex items-center gap-2">
        <ElTag type="primary">Jumpserver 权限矩阵</ElTag>
        <ElTag v-if="ruleType === 'all' || ruleType === 'user'" type="success">用户规则: {{ userRulesCount }}</ElTag>
        <ElTag v-if="ruleType === 'all' || ruleType === 'group'" type="warning">
          用户组规则: {{ groupRulesCount }}
        </ElTag>
        <ElTag>用户: {{ safeData.users?.length || 0 }}</ElTag>
      </div>
      <div class="flex items-center gap-2 text-sm text-gray-500">
        <span>点击单元格可查看详情</span>
      </div>
    </div>

    <!-- 授权规则说明 -->
    <div v-if="ruleType === 'all'" class="rounded bg-blue-50 p-3 text-sm">
      <p class="text-blue-900 font-medium">授权规则类型说明：</p>
      <ul class="ml-4 mt-2 text-blue-800 space-y-1">
        <li>
          •
          <strong>用户规则</strong>
          ：直接为特定用户授权的规则
        </li>
        <li>
          •
          <strong>用户组规则</strong>
          ：为用户组授权，组成员自动继承权限
        </li>
      </ul>
      <p class="mt-2 text-blue-700">授权对象可以是：资产（单个服务器）、节点（资产分组）、或全部资产</p>
    </div>
    <div v-else-if="ruleType === 'user'" class="rounded bg-green-50 p-3 text-sm">
      <p class="text-green-900 font-medium">用户规则：直接为特定用户授权访问资产</p>
    </div>
    <div v-else-if="ruleType === 'group'" class="rounded bg-orange-50 p-3 text-sm">
      <p class="text-orange-900 font-medium">用户组规则：为用户组授权，组成员自动继承权限</p>
    </div>

    <!-- 消息提示（如果有） -->
    <ElAlert v-if="safeData.message" type="info" :closable="false" show-icon class="mb-4">
      {{ safeData.message }}
    </ElAlert>

    <!-- 权限矩阵 -->
    <PermissionMatrix
      v-if="safeData.rules?.length > 0 && safeData.users?.length > 0 && filteredRules.length > 0"
      :rows="safeData.users"
      :columns="filteredRules"
      :matrix="filteredMatrix"
      row-key="id"
      col-key="rule_id"
      row-label="username"
      col-label="rule_name"
      :permissions-detail="safeData.permissions_detail || {}"
      @cell-click="handleCellClick"
    />

    <!-- 空状态 -->
    <ElEmpty v-else-if="!safeData.rules || safeData.rules.length === 0" description="暂无授权规则数据">
      <ElButton type="primary" @click="$emit('refresh')">同步授权规则</ElButton>
    </ElEmpty>

    <ElEmpty v-else-if="!safeData.users || safeData.users.length === 0" description="暂无用户数据" />

    <!-- 提示信息 -->
    <div v-if="safeData.rules?.length > 0" class="mt-4 text-sm text-gray-500">
      <p>提示：绿色 ✓ 表示有权限，灰色 ✗ 表示无权限</p>
      <p class="mt-1">授权规则来自 Jumpserver，通过用户组绑定自动管理用户权限</p>
    </div>
  </div>
</template>

<style scoped>
.jumpserver-matrix {
  @apply flex flex-col gap-4;
}
</style>
