<script setup lang="tsx">
  import { computed, onMounted, ref } from 'vue';
  import { ElMessageBox } from 'element-plus';
  import { Setting, User } from '@element-plus/icons-vue';
  import {
    fetchDeleteSSHCredential,
    fetchGetSSHCredentials,
    fetchTestSSHCredential
  } from '@/service/api/cmdb';
  import SshCredentialOperateDrawer from '../ssh/modules/ssh-credential-operate-drawer.vue';

  defineOptions({ name: 'CmdbAccessCredentials' });

  // 类型 Tab
  const activeTab = ref<CMDB.CredentialType | 'all'>('all');

  // 表格数据
  const allData = ref<CMDB.SSHCredential[]>([]);
  const loading = ref(false);

  const tableData = computed(() => {
    if (activeTab.value === 'all') return allData.value;
    return allData.value.filter(c => c.credentialType === activeTab.value);
  });

  // Drawer 状态
  const drawerVisible = ref(false);
  const operateType = ref<UI.TableOperateType>('add');
  const editingData = ref<CMDB.SSHCredential | null>(null);
  const defaultCredentialType = ref<CMDB.CredentialType>('user');

  // 测试连接状态
  const testLoading = ref(false);

  async function getData() {
    loading.value = true;
    try {
      const { data, error } = await fetchGetSSHCredentials();
      if (!error) {
        allData.value = data || [];
      }
    } finally {
      loading.value = false;
    }
  }

  function handleAdd() {
    operateType.value = 'add';
    editingData.value = null;
    defaultCredentialType.value = activeTab.value === 'all' ? 'user' : activeTab.value;
    drawerVisible.value = true;
  }

  function handleEdit(row: CMDB.SSHCredential) {
    operateType.value = 'edit';
    editingData.value = row;
    defaultCredentialType.value = row.credentialType || 'user';
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

  function getCredentialTypeTag(ct: CMDB.CredentialType) {
    return ct === 'system'
      ? { text: '系统运维', type: 'warning' as const }
      : { text: '用户连接', type: 'primary' as const };
  }

  onMounted(() => {
    getData();
  });
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <ElCard class="card-wrapper">
      <!-- 类型 Tab + 新增按钮 -->
      <div class="mb-16px flex items-center justify-between">
        <ElTabs v-model="activeTab" class="credential-tabs">
          <ElTabPane label="全部凭证" name="all" />
          <ElTabPane name="user">
            <template #label>
              <span class="flex items-center gap-4px">
                <ElIcon><User /></ElIcon>
                用户连接凭证
              </span>
            </template>
          </ElTabPane>
          <ElTabPane name="system">
            <template #label>
              <span class="flex items-center gap-4px">
                <ElIcon><Setting /></ElIcon>
                系统运维凭证
              </span>
            </template>
          </ElTabPane>
        </ElTabs>
        <ElButton type="primary" @click="handleAdd">新增凭证</ElButton>
      </div>

      <!-- 凭证用途说明 -->
      <ElAlert
        v-if="activeTab === 'system'"
        type="warning"
        :closable="false"
        class="mb-12px"
        description="系统运维凭证仅供 OneOps 后端使用（Agent 部署、重启、卸载、SSH 指标采集），不出现在用户连接弹窗中。通常需要 root 或具备 sudo 权限的账号。"
      />
      <ElAlert
        v-else-if="activeTab === 'user'"
        type="info"
        :closable="false"
        class="mb-12px"
        description="用户连接凭证用于用户通过堡垒机建立 SSH 会话，受访问策略约束，全程记录会话和命令审计。"
      />

      <!-- 数据表格 -->
      <ElTable v-loading="loading" :data="tableData" border stripe>
        <ElTableColumn prop="id" label="ID" width="70" align="center" />
        <ElTableColumn prop="name" label="凭证名称" min-width="150" />
        <ElTableColumn label="凭证用途" width="120" align="center">
          <template #default="{ row }">
            <ElTag :type="getCredentialTypeTag(row.credentialType).type" size="small">
              {{ getCredentialTypeTag(row.credentialType).text }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="username" label="用户名" width="120" align="center" />
        <ElTableColumn label="认证方式" width="100" align="center">
          <template #default="{ row }">
            <ElTag :type="getAuthTypeTag(row.authType).type" size="small">
              {{ getAuthTypeTag(row.authType).text }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="port" label="端口" width="80" align="center" />
        <ElTableColumn prop="description" label="描述" min-width="180" show-overflow-tooltip />
        <ElTableColumn label="状态" width="80" align="center">
          <template #default="{ row }">
            <ElTag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="操作" width="230" align="center" fixed="right">
          <template #default="{ row }">
            <ElButton type="success" size="small" :loading="testLoading" @click="handleTest(row)">测试</ElButton>
            <ElButton type="primary" size="small" @click="handleEdit(row)">编辑</ElButton>
            <ElButton type="danger" size="small" @click="handleDelete(row)">删除</ElButton>
          </template>
        </ElTableColumn>
      </ElTable>

      <SshCredentialOperateDrawer
        v-model:visible="drawerVisible"
        :operate-type="operateType"
        :row-data="editingData"
        :default-credential-type="defaultCredentialType"
        @submitted="getData"
      />
    </ElCard>
  </div>
</template>

<style scoped lang="scss">
  .card-wrapper {
    @apply flex-col-stretch;
  }
  .credential-tabs {
    :deep(.el-tabs__header) {
      margin-bottom: 0;
    }
  }
</style>
