<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue';
import { ElNotification, ElMessageBox, FormInstance, FormRules } from 'element-plus';
import {
  fetchGetSSHCredentials,
  fetchCreateSSHCredential,
  fetchUpdateSSHCredential,
  fetchDeleteSSHCredential,
  fetchTestSSHCredential
} from '@/service/api/cmdb';

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

// 对话框
const dialogVisible = ref(false);
const dialogTitle = ref('');
const formRef = ref<FormInstance>();
const form = reactive<CMDB.SSHCredentialForm>({
  name: '',
  description: '',
  username: 'root',
  authType: 'password',
  password: '',
  privateKey: '',
  passphrase: '',
  credentialType: 'user',
  status: 1
});

// 测试连接状态
const testLoading = ref(false);

// 表单验证规则
const rules: FormRules = {
  name: [
    { required: true, message: '请输入凭证名称', trigger: 'blur' },
    { min: 2, max: 100, message: '凭证名称长度在 2 到 100 个字符', trigger: 'blur' }
  ],
  username: [{ required: true, message: '请输入SSH用户名', trigger: 'blur' }],
  credentialType: [{ required: true, message: '请选择凭证用途', trigger: 'change' }],
  password: [
    {
      validator: (_rule, value, callback) => {
        if (form.authType === 'password' && !value) {
          callback(new Error('请输入密码'));
        } else {
          callback();
        }
      },
      trigger: 'blur'
    }
  ],
  privateKey: [
    {
      validator: (_rule, value, callback) => {
        if (form.authType === 'key' && !value) {
          callback(new Error('请输入私钥'));
        } else {
          callback();
        }
      },
      trigger: 'blur'
    }
  ]
};

// 获取凭证列表
async function getData() {
  loading.value = true;
  try {
    const { data } = await fetchGetSSHCredentials();
    allData.value = data || [];
  } catch (err) {
    console.error('获取凭证列表失败:', err);
    ElNotification.error('获取凭证列表失败');
  } finally {
    loading.value = false;
  }
}

// 新增
function handleAdd() {
  dialogTitle.value = '新增SSH凭证';
  Object.assign(form, {
    id: undefined,
    name: '',
    description: '',
    username: 'root',
    authType: 'password',
    password: '',
    privateKey: '',
    passphrase: '',
    credentialType: activeTab.value === 'all' ? 'user' : activeTab.value,
    status: 1
  });
  dialogVisible.value = true;
}

// 编辑
function handleEdit(row: CMDB.SSHCredential) {
  dialogTitle.value = '编辑SSH凭证';
  Object.assign(form, {
    id: row.id,
    name: row.name,
    description: row.description,
    username: row.username,
    authType: row.authType,
    credentialType: row.credentialType || 'user',
    password: '',
    privateKey: '',
    passphrase: ''
  });
  dialogVisible.value = true;
}

// 保存
async function handleSave() {
  if (!formRef.value) return;
  try {
    await formRef.value.validate();
    const formData = { ...form };
    if (form.id) {
      await fetchUpdateSSHCredential(form.id, formData);
      ElNotification.success('保存成功');
    } else {
      await fetchCreateSSHCredential(formData);
      ElNotification.success('创建成功');
    }
    dialogVisible.value = false;
    getData();
  } catch (error) {
    if (error && typeof error === 'object' && 'message' in error) {
      ElNotification.error(typeof error.message === 'string' ? error.message : '保存失败');
    } else if (error !== false) {
      ElNotification.error('保存失败');
    }
  }
}

// 删除
function handleDelete(row: CMDB.SSHCredential) {
  ElMessageBox.confirm(`确定要删除凭证 "${row.name}" 吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  })
    .then(async () => {
      try {
        await fetchDeleteSSHCredential(row.id);
        ElNotification.success('删除成功');
        getData();
      } catch (error) {
        console.error('删除失败:', error);
        ElNotification.error('删除失败');
      }
    })
    .catch(() => {});
}

// 测试连接
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
    const { data } = await fetchTestSSHCredential(row.id, testIp.value, parseInt(testPort.value));
    if (data.success) {
      ElNotification.success(`连接测试成功：${data.message || '可以连接'}`);
    } else {
      ElNotification.warning(`连接测试：${data.message || '无法连接'}`);
    }
  } catch (error) {
    ElNotification.error('连接测试失败');
  } finally {
    testLoading.value = false;
  }
}

// 辅助：认证类型标签
function getAuthTypeTag(type: string) {
  const map: Record<string, { text: string; type: any }> = {
    password: { text: '密码', type: 'primary' },
    key: { text: '密钥', type: 'success' }
  };
  return map[type] || { text: type, type: 'info' };
}

// 辅助：凭证类型标签
function getCredentialTypeTag(ct: CMDB.CredentialType) {
  return ct === 'system'
    ? { text: '系统运维', type: 'warning' as const }
    : { text: '用户连接', type: 'primary' as const };
}

// 初始化
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
                <el-icon><User /></el-icon>
                用户连接凭证
              </span>
            </template>
          </ElTabPane>
          <ElTabPane name="system">
            <template #label>
              <span class="flex items-center gap-4px">
                <el-icon><Setting /></el-icon>
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
    </ElCard>

    <!-- 新增/编辑对话框 -->
    <ElDialog v-model="dialogVisible" :title="dialogTitle" width="620px">
      <ElForm ref="formRef" :model="form" :rules="rules" label-width="120px">
        <ElFormItem label="凭证用途" prop="credentialType">
          <ElRadioGroup v-model="form.credentialType">
            <ElRadio value="user">
              <span class="font-medium">用户连接</span>
              <span class="ml-6px text-12px text-gray-400">用于用户堡垒 SSH，受访问策略约束</span>
            </ElRadio>
          </ElRadioGroup>
          <ElRadioGroup v-model="form.credentialType" class="mt-8px">
            <ElRadio value="system">
              <span class="font-medium">系统运维</span>
              <span class="ml-6px text-12px text-gray-400">供 OneOps 后端 Agent 部署、采集使用，需 root / sudo 权限</span>
            </ElRadio>
          </ElRadioGroup>
        </ElFormItem>
        <ElFormItem label="凭证名称" prop="name">
          <ElInput v-model="form.name" placeholder="请输入凭证名称" />
        </ElFormItem>
        <ElFormItem label="用户名" prop="username">
          <ElInput v-model="form.username" placeholder="请输入SSH用户名" />
        </ElFormItem>
        <ElFormItem label="认证类型" prop="authType">
          <ElRadioGroup v-model="form.authType">
            <ElRadio value="password">密码认证</ElRadio>
            <ElRadio value="key">密钥认证</ElRadio>
          </ElRadioGroup>
        </ElFormItem>
        <template v-if="form.authType === 'password'">
          <ElFormItem label="密码" prop="password">
            <ElInput v-model="form.password" type="password" placeholder="请输入密码" show-password />
          </ElFormItem>
        </template>
        <template v-if="form.authType === 'key'">
          <ElFormItem label="私钥" prop="privateKey">
            <ElInput v-model="form.privateKey" type="textarea" :rows="6" placeholder="请输入私钥内容" />
          </ElFormItem>
          <ElFormItem label="私钥密码">
            <ElInput v-model="form.passphrase" type="password" placeholder="如果私钥有密码请输入" show-password />
          </ElFormItem>
        </template>
        <ElFormItem label="描述">
          <ElInput v-model="form.description" type="textarea" :rows="3" placeholder="请输入描述" />
        </ElFormItem>
        <ElFormItem label="状态">
          <ElRadioGroup v-model="form.status">
            <ElRadio :value="1">启用</ElRadio>
            <ElRadio :value="0">禁用</ElRadio>
          </ElRadioGroup>
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="dialogVisible = false">取消</ElButton>
        <ElButton type="primary" @click="handleSave">保存</ElButton>
      </template>
    </ElDialog>
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
