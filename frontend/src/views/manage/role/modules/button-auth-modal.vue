<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { $t } from '@/locales';
import { fetchGetPermissionList, fetchGetRolePermissions, fetchAssignRolePermissions } from '@/service/api';
import { ElMessage, ElNotification } from 'element-plus';

defineOptions({ name: 'ButtonAuthModal' });

interface Props {
  /** the roleId */
  roleId: number;
  /** the role data */
  roleData?: Api.SystemManage.Role | null;
}

const props = defineProps<Props>();

const visible = defineModel<boolean>('visible', {
  default: false
});

function closeModal() {
  visible.value = false;
}

const title = computed(() => `为角色 "${props.roleData?.name || ''}" 分配权限`);

interface PermissionNode {
  id: number;
  label: string;
  code: string;
  module: string;
  children?: PermissionNode[];
}

const permissionTree = ref<PermissionNode[]>([]);
const selectedPermissionIds = ref<number[]>([]);
const loading = ref(false);

// 获取所有权限并构建树
async function getAllPermissions() {
  loading.value = true;
  const { error, data } = await fetchGetPermissionList({ current: 1, size: 1000 });

  if (!error && data) {
    const permissions = data.records || data || [];

    // 构建权限树
    const tree: PermissionNode[] = [];
    const moduleMap = new Map<string, PermissionNode>();

    permissions.forEach((perm: Api.SystemManage.Permission) => {
      // 按模块分组
      if (!moduleMap.has(perm.module)) {
        const moduleNode: PermissionNode = {
          id: -1, // 模块节点的ID设为负数，避免与实际权限ID冲突
          label: getModuleLabel(perm.module),
          code: perm.module,
          module: perm.module,
          children: []
        };
        moduleMap.set(perm.module, moduleNode);
        tree.push(moduleNode);
      }

      // 添加权限到对应模块
      const moduleNode = moduleMap.get(perm.module)!;
      moduleNode.children!.push({
        id: perm.id,
        label: perm.name,
        code: perm.code,
        module: perm.module
      });
    });

    permissionTree.value = tree;
  }

  loading.value = false;
}

// 获取角色已拥有的权限
async function getRolePermissions() {
  loading.value = true;
  const { error, data } = await fetchGetRolePermissions(props.roleId);

  if (!error && data) {
    selectedPermissionIds.value = data;
  }

  loading.value = false;
}

// 模块名称映射
function getModuleLabel(module: string): string {
  const moduleMap: Record<string, string> = {
    'system': '系统管理',
    'auth': '授权中心',
    'cmdb': '资产管理',
    'monitoring': '监控中心',
    'k8s': 'K8s管理'
  };
  return moduleMap[module] || module;
}

// 提交权限分配
async function handleSubmit() {
  loading.value = true;

  try {
    const { error } = await fetchAssignRolePermissions(props.roleId, {
      permissionIds: selectedPermissionIds.value
    });

    loading.value = false;

    if (!error) {
      ElNotification({
        title: '分配成功',
        message: `成功为角色分配 ${selectedPermissionIds.value.length} 个权限`,
        type: 'success',
        duration: 3000,
        position: 'top-right'
      });
      closeModal();
    } else {
      ElMessage.error(error.msg || '分配失败');
    }
  } catch (err) {
    loading.value = false;
    ElMessage.error('分配失败');
  }
}

// 监听模态框显示
watch(
  () => visible.value,
  async (val) => {
    if (val && props.roleId) {
      await getAllPermissions();
      await getRolePermissions();
    }
  }
);

// 监听模态框显示
watch(
  () => visible.value,
  async (val) => {
    if (val && props.roleId) {
      await getAllPermissions();
      await getRolePermissions();
    }
  }
);
</script>

<template>
  <ElDialog
    v-model="visible"
    :title="title"
    width="600px"
    :close-on-click-modal="false"
  >
    <div v-loading="loading" class="permission-tree-container">
      <div class="tip-text">
        <span>勾选权限节点，为角色分配相应的操作权限</span>
      </div>

      <ElTree
        :data="permissionTree"
        v-model:checked-keys="selectedPermissionIds"
        node-key="id"
        show-checkbox
        :props="{
          label: 'label',
          children: 'children'
        }"
        class="permission-tree"
        :default-expand-all="true"
      />
    </div>

    <template #footer>
      <div class="dialog-footer">
        <ElButton @click="closeModal">取消</ElButton>
        <ElButton type="primary" :loading="loading" @click="handleSubmit">
          确定
        </ElButton>
      </div>
    </template>
  </ElDialog>
</template>

<style scoped lang="scss">
.permission-tree-container {
  min-height: 400px;
  max-height: 500px;
  overflow-y: auto;
  padding: 16px;
}

.tip-text {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px;
  background: var(--el-fill-color-light);
  border-radius: 4px;
  margin-bottom: 16px;
  font-size: 14px;
  color: var(--el-text-color-regular);
}

.permission-tree {
  :deep(.el-tree-node__content) {
    height: 36px;
  }

  :deep(.el-checkbox__input.is-checked .el-checkbox__inner) {
    background-color: var(--el-color-primary);
    border-color: var(--el-color-primary);
  }
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>
