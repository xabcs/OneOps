<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { ElNotification, ElMessage } from 'element-plus';
import { fetchGetPermissionList, fetchAssignRolePermissions } from '@/service/api';

defineOptions({
  name: 'RoleAssignModal'
});

interface Props {
  visible: boolean;
  roleId: number;
  roleData?: Api.SystemManage.Role | null;
}

interface Emits {
  (e: 'update:visible', visible: boolean): void;
  (e: 'submitted'): void;
}

const props = defineProps<Props>();
const emit = defineEmits<Emits>();

const visible = computed({
  get: () => props.visible,
  set: (val: boolean) => emit('update:visible', val)
});

const title = computed(() => {
  return props.roleData ? `为角色 "${props.roleData.name}" 分配权限` : '分配权限';
});

// 所有权限
const allPermissions = ref<Api.SystemManage.Permission[]>([]);
// 选中的权限ID列表
const selectedPermissionIds = ref<number[]>([]);
// 加载状态
const loading = ref(false);

// 获取所有权限
async function getAllPermissions() {
  loading.value = true;
  const { error, data } = await fetchGetPermissionList({ current: 1, size: 1000 });
  loading.value = false;

  if (!error && data) {
    allPermissions.value = (data.records || data || []).filter((perm: Api.SystemManage.Permission) => perm.status === 1);
  }
}

// 监听模态框显示状态
watch(
  () => props.visible,
  async (val) => {
    if (val) {
      await getAllPermissions();
      // 如果有角色数据，获取已分配的权限
      if (props.roleData) {
        // TODO: 调用API获取该角色已分配的权限
        // 暂时使用空数组
        selectedPermissionIds.value = [];
      }
    }
  }
);

// 提交分配
async function handleSubmit() {
  if (selectedPermissionIds.value.length === 0) {
    ElMessage.warning('请选择至少一个权限');
    return;
  }

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
      visible.value = false;
      emit('submitted');
    } else {
      ElMessage.error(error.msg || '分配失败');
    }
  } catch (err) {
    loading.value = false;
    ElMessage.error('分配失败');
  }
}

// 权限状态标签
function getPermissionStatusType(status: number): UI.ThemeColor {
  if (status === 1) return 'success';
  if (status === 0) return 'danger';
  return 'info';
}

function getPermissionStatusText(status: number): string {
  if (status === 1) return '启用';
  if (status === 0) return '禁用';
  return '未知';
}
</script>

<template>
  <ElDialog
    v-model="visible"
    :title="title"
    width="600px"
    :close-on-click-modal="false"
    @closed="selectedPermissionIds = []"
  >
    <div class="permission-assign-container">
      <div class="selected-info">
        <ElText type="info">
          已选择 <ElText type="primary">{{ selectedPermissionIds.length }}</ElText> 个权限
        </ElText>
      </div>

      <ElTable
        v-loading="loading"
        :data="allPermissions"
        border
        stripe
        class="permission-table"
        row-key="id"
        max-height="400px"
        @selection-change="selectedPermissionIds = $event.map((row: any) => row.id)"
      >
        <ElTableColumn
          type="selection"
          width="48"
          :selectable="(row: Api.SystemManage.Permission) => row.status === 1"
        />
        <ElTableColumn prop="name" label="权限名称" minWidth="120" />
        <ElTableColumn prop="code" label="权限编码" minWidth="150" />
        <ElTableColumn prop="module" label="模块" minWidth="100" />
        <ElTableColumn prop="resource" label="资源" minWidth="100" />
        <ElTableColumn prop="action" label="操作" minWidth="100" />
        <ElTableColumn
          prop="status"
          label="状态"
          align="center"
          width="100"
        >
          <template #default="{ row }">
            <ElTag :type="getPermissionStatusType(row.status)" size="small">
              {{ getPermissionStatusText(row.status) }}
            </ElTag>
          </template>
        </ElTableColumn>
      </ElTable>
    </div>

    <template #footer>
      <div class="modal-footer">
        <ElButton @click="visible.value = false">取消</ElButton>
        <ElButton
          type="primary"
          :loading="loading"
          :disabled="selectedPermissionIds.length === 0"
          @click="handleSubmit"
        >
          确定
        </ElButton>
      </div>
    </template>
  </ElDialog>
</template>

<style scoped lang="scss">
.permission-assign-container {
  padding: 16px 0;
}

.selected-info {
  margin-bottom: 16px;
  padding: 12px;
  background: var(--el-fill-color-light);
  border-radius: 4px;
}

.permission-table {
  width: 100%;
  border-radius: var(--sx-table-radius);
  overflow: hidden;
  border: 1px solid var(--sx-table-border);

  :deep(th.el-table__cell) {
    background-color: var(--sx-table-header-bg);
    color: var(--sx-table-header-text);
    font-weight: 600;
  }

  :deep(.el-checkbox__inner),
  :deep(.el-checkbox__inner::before),
  :deep(.el-checkbox__inner::after) {
    border-radius: 0;
  }
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>
