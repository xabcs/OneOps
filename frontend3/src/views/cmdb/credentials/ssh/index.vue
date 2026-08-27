<script setup lang="tsx">
  import { onMounted, ref } from 'vue';
  import { Plus, Refresh } from '@element-plus/icons-vue';
  import {
    fetchDeleteSSHCredential,
    fetchGetSSHCredentials,
    fetchTestSSHCredential
  } from '@/service/api/cmdb';
  import SshCredentialOperateDrawer from './modules/ssh-credential-operate-drawer.vue';

  defineOptions({ name: 'CmdbSshCredentials' });

  const loading = ref(false);
  const tableData = ref<CMDB.SSHCredential[]>([]);
  const testLoading = ref(false);

  const drawerVisible = ref(false);
  const operateType = ref<UI.TableOperateType>('add');
  const editingData = ref<CMDB.SSHCredential | null>(null);

  async function getData() {
    loading.value = true;
    try {
      const { data, error } = await fetchGetSSHCredentials();
      if (!error) {
        tableData.value = data || [];
      }
    } finally {
      loading.value = false;
    }
  }

  function handleAdd() {
    operateType.value = 'add';
    editingData.value = null;
    drawerVisible.value = true;
  }

  function handleEdit(row: CMDB.SSHCredential) {
    operateType.value = 'edit';
    editingData.value = row;
    drawerVisible.value = true;
  }

  async function handleDelete(row: CMDB.SSHCredential) {
    try {
      await ElMessageBox.confirm(`确定要删除凭证 "${row.name}" 吗？`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      });
      const { error } = await fetchDeleteSSHCredential(row.id);
      if (!error) {
        ElMessage.success('删除成功');
        getData();
      }
    } catch {
      // 用户取消
    }
  }

  async function handleTest(row: CMDB.SSHCredential) {
    const testIp = await ElMessageBox.prompt('请输入要测试连接的IP地址', '测试连接 - IP地址', {
      confirmButtonText: '下一步',
      cancelButtonText: '取消',
      inputPattern: /^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$/,
      inputErrorMessage: '请输入有效的IP地址'
    }).catch(() => null);

    if (!testIp) return;

    const testPort = await ElMessageBox.prompt('请输入SSH端口号', '测试连接 - SSH端口', {
      confirmButtonText: '测试',
      cancelButtonText: '取消',
      inputValue: '22',
      inputPattern: /^[1-9][0-9]{0,4}$/,
      inputErrorMessage: '请输入有效的端口号（1-65535）'
    }).catch(() => null);

    if (!testPort) return;

    testLoading.value = true;
    try {
      const { data } = await fetchTestSSHCredential(row.id, testIp.value, Number.parseInt(testPort.value));
      if (data?.success) {
        ElMessage.success(`连接测试成功：${data.message || '可以连接'}`);
      } else {
        ElMessage.warning(`连接测试：${data?.message || '无法连接'}`);
      }
    } catch {
      ElMessage.error('连接测试失败');
    } finally {
      testLoading.value = false;
    }
  }

  function getAuthTypeTag(
    type: string
  ): { text: string; type: 'primary' | 'success' | 'warning' | 'info' | 'danger' } {
    const typeMap: Record<string, { text: string; type: 'primary' | 'success' | 'warning' | 'info' | 'danger' }> = {
      password: { text: '密码', type: 'primary' },
      key: { text: '密钥', type: 'success' }
    };
    return typeMap[type] || { text: type, type: 'info' };
  }

  onMounted(() => {
    getData();
  });
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <ElCard class="card-wrapper">
      <div class="mb-16px flex justify-between">
        <ElButton :icon="Refresh" @click="getData">刷新</ElButton>
        <PermissionButton code="cmdb.credential.create" type="primary" :icon="Plus" @click="handleAdd">新增凭证</PermissionButton>
      </div>

      <ElTable v-loading="loading" :data="tableData" border stripe>
        <ElTableColumn prop="id" label="ID" width="70" align="center" />
        <ElTableColumn prop="name" label="凭证名称" min-width="150" align="center" />
        <ElTableColumn prop="username" label="用户名" width="120" align="center" />
        <ElTableColumn label="认证类型" width="100" align="center">
          <template #default="{ row }">
            <ElTag :type="getAuthTypeTag(row.authType).type" size="small">
              {{ getAuthTypeTag(row.authType).text }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="port" label="端口" width="80" align="center" />
        <ElTableColumn prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <ElTableColumn label="状态" width="80" align="center">
          <template #default="{ row }">
            <ElTag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="操作" width="230" align="center" fixed="right">
          <template #default="{ row }">
            <PermissionButton code="cmdb.credential.test" type="success" size="small" :loading="testLoading" @click="handleTest(row)">测试</PermissionButton>
            <PermissionButton code="cmdb.credential.update" type="primary" size="small" @click="handleEdit(row)">编辑</PermissionButton>
            <PermissionButton code="cmdb.credential.delete" type="danger" size="small" @click="handleDelete(row)">删除</PermissionButton>
          </template>
        </ElTableColumn>
      </ElTable>

      <SshCredentialOperateDrawer
        v-model:visible="drawerVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="getData"
      />
    </ElCard>
  </div>
</template>

<style scoped lang="scss">
  .card-wrapper {
    @apply flex-col-stretch;
  }
</style>
