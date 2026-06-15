<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { ElButton, ElMessage, ElMessageBox, ElTable, ElTableColumn } from 'element-plus';
import { assignK8sClusterRole, fetchK8sClusterUsers, revokeK8sClusterRole } from '@/service/api/k8s';

interface Props {
  clusterId: number;
}

const props = defineProps<Props>();

const loading = ref(false);
const dataSource = ref<any[]>([]);

const loadUsers = async () => {
  loading.value = true;
  try {
    const res = await fetchK8sClusterUsers(props.clusterId);
    if (res.code === 200) {
      dataSource.value = res.data || [];
    }
  } catch (error: any) {
    ElMessage.error(error.message || '加载用户列表失败');
  } finally {
    loading.value = false;
  }
};

const handleAddUser = () => {
  // TODO: 实现添加用户对话框
  ElMessage.info('添加用户功能待实现');
};

const handleRevoke = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定要撤销用户 "${row.username}" 的权限吗？`, '确认', {
      type: 'warning'
    });
    await revokeK8sClusterRole(props.clusterId, row.user_id);
    ElMessage.success('撤销权限成功');
    loadUsers();
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '撤销权限失败');
    }
  }
};

onMounted(() => {
  loadUsers();
});
</script>

<template>
  <div class="p-4">
    <div class="mb-4">
      <ElButton type="primary" @click="handleAddUser">添加用户</ElButton>
    </div>
    <ElTable v-loading="loading" :data="dataSource" stripe>
      <ElTableColumn prop="user_id" label="用户ID" />
      <ElTableColumn prop="username" label="用户名" />
      <ElTableColumn prop="nickname" label="昵称" />
      <ElTableColumn prop="role_id" label="角色ID" />
      <ElTableColumn prop="role_name" label="角色名称" />
      <ElTableColumn prop="created_at" label="授权时间" />
      <ElTableColumn label="操作" width="120">
        <template #default="{ row }">
          <ElButton type="danger" size="small" @click="handleRevoke(row)">撤销权限</ElButton>
        </template>
      </ElTableColumn>
    </ElTable>
  </div>
</template>
