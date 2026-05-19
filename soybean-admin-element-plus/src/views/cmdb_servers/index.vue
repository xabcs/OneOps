<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue';
import {
  fetchGetServers,
  fetchCreateServer,
  fetchUpdateServer,
  fetchDeleteServer,
  fetchGetServerGroups,
  fetchCreateServerGroup,
  fetchUpdateServerGroup,
  fetchDeleteServerGroup,
  fetchAssignServerToGroup,
  fetchGetSSHCredentials,
  fetchGetBusinessUnits,
  fetchCreateSSHCredential,
  fetchUpdateSSHCredential,
  fetchDeleteSSHCredential
} from '@/service/api';
import { ElNotification, ElMessageBox, FormInstance, FormRules } from 'element-plus';

defineOptions({ name: 'CmdbServers' });

// 查询表单
const searchForm = reactive<CMDB.ServerQuery>({
  hostname: '',
  ip: '',
  env: undefined,
  status: undefined,
  businessId: undefined,
  provider: undefined
});

// 表格数据
const tableData = ref<CMDB.Server[]>([]);
const loading = ref(false);
const total = ref(0);

// 分页信息
const pagination = reactive({
  page: 1,
  pageSize: 20
});

// 业务系统列表
const businessUnits = ref<CMDB.BusinessUnit[]>([]);

// 主机分组列表
const serverGroups = ref<CMDB.ServerGroup[]>([]);

// 环境选项
const envOptions = [
  { label: '生产', value: 'prod' },
  { label: '测试', value: 'test' },
  { label: '开发', value: 'dev' }
] as const;

const statusOptions = [
  { label: '在线', value: 'online' },
  { label: '离线', value: 'offline' },
  { label: '未知', value: 'unknown' }
] as const;

const providerOptions = [
  { label: '阿里云', value: 'aliyun' },
  { label: '腾讯云', value: 'tencent' },
  { label: 'AWS', value: 'aws' },
  { label: '华为云', value: 'huawei' },
  { label: '自建', value: 'self' },
  { label: '其他', value: 'other' }
] as const;

const treeSelectProps = { label: 'name', value: 'id', children: 'children' } as any;

// 对话框
const dialogVisible = ref(false);
const dialogTitle = ref('');
const serverFormRef = ref<FormInstance>();
const serverType = ref<'normal' | 'cloud'>('normal'); // 主机类型
const serverForm = reactive<CMDB.ServerForm & { groupIds?: number[] }>({
  hostname: '',
  ip: '',
  innerIp: '',
  credentialId: undefined as unknown as number,
  serverType: 'vm',
  groupIds: [],
  sshPort: 22,
  remarks: ''
});

// 云主机表单
const cloudForm = reactive<CMDB.CloudServerForm>({
  provider: 'aliyun',
  instanceName: '',
  instanceType: '',
  region: '',
  zone: '',
  chargeType: 'postpay'
});

// SSH凭证列表
const sshCredentials = ref<CMDB.SSHCredential[]>([]);

// 表单验证规则
const serverFormRules: FormRules = {
  hostname: [
    { required: true, message: '请输入主机名', trigger: 'blur' },
    { min: 2, max: 100, message: '主机名长度在 2 到 100 个字符', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9.-]+$/, message: '主机名只能包含字母、数字、点和连字符', trigger: 'blur' }
  ],
  ip: [
    { required: true, message: '请输入连接IP', trigger: 'blur' },
    { pattern: /^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$/, message: '请输入有效的IP地址', trigger: 'blur' }
  ],
  innerIp: [
    { pattern: /^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$/, message: '请输入有效的IP地址', trigger: 'blur' }
  ],
  credentialId: [
    { required: true, message: '请选择SSH凭证', trigger: 'change' }
  ],
  groupIds: [
    { required: true, message: '请选择所属分组', trigger: 'change' }
  ]
};

// 获取SSH凭证列表
async function getSSHCredentials() {
  try {
    const { data } = await fetchGetSSHCredentials();
    sshCredentials.value = data || [];
  } catch (err) {
    console.error('获取SSH凭证失败:', err);
    ElNotification.error('获取SSH凭证失败');
  }
}

// 获取配置相关

async function getBusinessUnits() {
  try {
    const { data } = await fetchGetBusinessUnits();
    businessUnits.value = data || [];
  } catch (err) {
    console.error('获取业务系统失败:', err);
  }
}

// 扁平化的分组列表（用于下拉选择）
const flatServerGroups = computed(() => {
  const flat: CMDB.ServerGroup[] = [];

  function flatten(groups: CMDB.ServerGroup[]) {
    groups.forEach(group => {
      flat.push(group);
      if (group.children && group.children.length > 0) {
        flatten(group.children);
      }
    });
  }

  flatten(serverGroups.value);
  return flat;
});

const flatBusinessUnits = computed(() => {
  const flat: CMDB.BusinessUnit[] = [];

  function flatten(units: CMDB.BusinessUnit[]) {
    units.forEach(unit => {
      flat.push(unit);
      if (unit.children && unit.children.length > 0) {
        flatten(unit.children);
      }
    });
  }

  flatten(businessUnits.value);
  return flat;
});


// 获取主机分组列表
async function getServerGroups() {
  try {
    const { data } = await fetchGetServerGroups();
    if (data && Array.isArray(data)) {
      // 构建树形结构
      serverGroups.value = buildGroupTree(data, 0);
    }
  } catch (err) {
    console.error('获取主机分组失败:', err);
  }
}

// 构建分组树
function buildGroupTree(groups: CMDB.ServerGroup[], parentId: number): CMDB.ServerGroup[] {
  const result: CMDB.ServerGroup[] = [];
  groups.forEach(group => {
    if (group.parentId === parentId) {
      const children = buildGroupTree(groups, group.id);
      result.push({
        ...group,
        children: children.length > 0 ? children : undefined
      });
    }
  });
  return result;
}

function getBusinessName(businessId?: number) {
  if (!businessId) return '-';
  const unit = flatBusinessUnits.value.find(item => item.id === businessId);
  return unit?.name || '-';
}

function getStatusTag(status: string) {
  const statusMap: Record<string, { text: string; type: 'success' | 'warning' | 'danger' | 'info' }> = {
    online: { text: '在线', type: 'success' },
    offline: { text: '离线', type: 'danger' },
    unknown: { text: '未知', type: 'info' }
  };
  return statusMap[status] || { text: status, type: 'info' };
}

// 搜索
function handleSearch() {
  pagination.page = 1;
  getServers();
}

// 重置
function handleReset() {
  Object.assign(searchForm, {
    hostname: '',
    ip: '',
    env: undefined,
    status: undefined,
    businessId: undefined,
    provider: undefined
  });
  pagination.page = 1;
  getServers();
}

// 分页变化
function handlePageChange(page: number) {
  pagination.page = page;
  getServers();
}

// 页面大小变化
function handlePageSizeChange(pageSize: number) {
  pagination.pageSize = pageSize;
  pagination.page = 1;
  getServers();
}

// 打开新增对话框
function handleAdd() {
  dialogTitle.value = '新增服务器';
  serverType.value = 'normal';
  Object.assign(serverForm, {
    id: undefined,
    hostname: '',
    ip: '',
    innerIp: '',
    credentialId: undefined as unknown as number,
    serverType: 'vm',
    groupIds: [],
    sshPort: 22,
    remarks: ''
  });
  Object.assign(cloudForm, {
    provider: 'aliyun',
    instanceId: '',
    instanceName: '',
    instanceType: '',
    region: '',
    zone: '',
    chargeType: 'postpay'
  });
  dialogVisible.value = true;
}

// 打开编辑对话框
function handleEdit(row: CMDB.Server) {
  dialogTitle.value = '编辑服务器';
  serverType.value = row.cloudInfo ? 'cloud' : 'normal';
  Object.assign(serverForm, {
    id: row.id,
    hostname: row.hostname,
    ip: row.ip,
    innerIp: row.innerIp,
    credentialId: row.credentialId || 0,
    groupIds: row.groups?.map(g => g.id) || [],
    sshPort: row.sshPort,
    remarks: row.remarks
  });
  if (row.cloudInfo) {
    Object.assign(cloudForm, {
      provider: row.provider as any,
      instanceId: row.cloudInfo.instanceId,
      instanceName: row.cloudInfo.instanceName,
      instanceType: row.cloudInfo.instanceType,
      region: row.cloudInfo.region,
      zone: row.cloudInfo.zone,
      chargeType: row.cloudInfo.chargeType
    });
  } else {
    Object.assign(cloudForm, {
      provider: 'aliyun',
      instanceId: '',
      instanceName: '',
      instanceType: '',
      region: '',
      zone: '',
      chargeType: 'postpay'
    });
  }
  dialogVisible.value = true;
}

// 凭证管理对话框状态
const credentialDialogVisible = ref(false);
const credentialFormRef = ref<FormInstance>();
const credentialForm = reactive<CMDB.SSHCredentialForm>({
  name: '',
  description: '',
  username: 'root',
  authType: 'password',
  password: '',
  privateKey: '',
  passphrase: '',
  port: 22,
  sortOrder: 0,
  status: 1
});

// 凭证表单验证规则
const credentialFormRules: FormRules = {
  name: [
    { required: true, message: '请输入凭证名称', trigger: 'blur' },
    { min: 2, max: 100, message: '凭证名称长度在 2 到 100 个字符', trigger: 'blur' }
  ],
  username: [
    { required: true, message: '请输入SSH用户名', trigger: 'blur' }
  ],
  password: [
    {
      validator: (rule, value, callback) => {
        if (credentialForm.authType === 'password' && !value) {
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
      validator: (rule, value, callback) => {
        if (credentialForm.authType === 'key' && !value) {
          callback(new Error('请输入私钥'));
        } else {
          callback();
        }
      },
      trigger: 'blur'
    }
  ]
};

// 打开凭证管理
async function handleManageCredentials() {
  await getSSHCredentials();
  credentialDialogVisible.value = true;
}

// 新增凭证
function handleAddCredential() {
  Object.assign(credentialForm, {
    id: undefined,
    name: '',
    description: '',
    username: 'root',
    authType: 'password',
    password: '',
    privateKey: '',
    passphrase: '',
    port: 22,
    sortOrder: 0,
    status: 1
  });
}

// 编辑凭证
function handleEditCredential(cred: CMDB.SSHCredential) {
  Object.assign(credentialForm, {
    id: cred.id,
    name: cred.name,
    description: cred.description,
    username: cred.username,
    authType: cred.authType,
    port: cred.port,
    sortOrder: cred.sortOrder,
    status: cred.status
  });
}

// 保存凭证
async function handleSaveCredential() {
  if (!credentialFormRef.value) return;

  try {
    await credentialFormRef.value.validate();

    if (credentialForm.id) {
      await fetchUpdateSSHCredential(credentialForm.id, credentialForm);
      ElNotification.success('凭证更新成功');
    } else {
      await fetchCreateSSHCredential(credentialForm);
      ElNotification.success('凭证创建成功');
    }
    Object.assign(credentialForm, { name: '' });
    await getSSHCredentials();
  } catch (error) {
    if (error && typeof error === 'object' && 'message' in error) {
      ElNotification.error(typeof error.message === 'string' ? error.message : '保存失败');
    }
  }
}

// 删除凭证
function handleDeleteCredential(cred: CMDB.SSHCredential) {
  ElMessageBox.confirm(`确定要删除凭证 "${cred.name}" 吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  })
    .then(async () => {
      await fetchDeleteSSHCredential(cred.id);
      ElNotification.success('删除成功');
      await getSSHCredentials();
    })
    .catch(() => {});
}

// 保存
async function handleSave() {
  if (!serverFormRef.value) return;

  try {
    await serverFormRef.value.validate();

    const formData: CMDB.ServerForm = {
      ...serverForm,
      serverType: serverType.value === 'cloud' ? 'vm' : serverForm.serverType || 'vm',
      cloudInfo:
        serverType.value === 'cloud'
          ? {
              ...cloudForm,
              provider: cloudForm.provider,
              publicIp: serverForm.ip,
              privateIp: serverForm.innerIp
            }
          : null
    };

    if (serverForm.id) {
      await fetchUpdateServer(serverForm.id, formData);
      ElNotification.success('服务器更新成功');
    } else {
      await fetchCreateServer(formData);
      ElNotification.success('服务器创建成功');
    }
    dialogVisible.value = false;
    getServers();
  } catch (error) {
    if (error && typeof error === 'object' && 'message' in error) {
      ElNotification.error(typeof error.message === 'string' ? error.message : '保存失败');
    } else {
      ElNotification.error('保存失败');
    }
  }
}

// 删除
function handleDelete(row: CMDB.Server) {
  ElMessageBox.confirm(`确定要删除服务器 "${row.hostname}" 吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  })
    .then(async () => {
      try {
        await fetchDeleteServer(row.id);
        ElNotification.success('删除成功');
        getServers();
      } catch (error) {
        console.error('删除失败:', error);
        ElNotification.error('删除失败');
      }
    })
    .catch(() => {});
}

// ========== 主机分组管理 ==========

// 打开新增分组对话框
function handleAddGroup() {
  dialogVisible.value = false;
  Object.assign(groupForm, {
    id: undefined,
    name: '',
    code: '',
    parentId: 0,
    description: '',
    color: '#409EFF',
    icon: 'mdi:folder',
    sortOrder: 0,
    status: 1
  });
  if (!assignForm.groupId) {
    ElNotification.warning('请选择分组');
    return;
  }

  try {
    await fetchAssignServerToGroup(assignForm.serverId, assignForm.groupId);
    ElNotification.success('分配成功');
    assignDialogVisible.value = false;
    getServers();
  } catch (error) {
    console.error('分配失败:', error);
    ElNotification.error('分配失败');
  }
}

// 获取服务器所属分组名称

// 环境标签
function getEnvTag(env: string) {
  const envMap: Record<string, { text: string; type: any }> = {
    prod: { text: '生产', type: 'danger' },
    test: { text: '测试', type: 'warning' },
    dev: { text: '开发', type: 'info' }
  };
  return envMap[env] || { text: env, type: 'info' };
}

// 获取服务器类型文本
function getServerTypeText(type: string) {
  const typeMap: Record<string, string> = {
    physical: '物理机',
    vm: '虚拟机',
    container: '容器'
  };
  return typeMap[type] || type;
}

// 获取服务器所属分组名称
function getGroupNames(groups?: CMDB.ServerGroup[]) {
  if (!groups || groups.length === 0) return '-';
  return groups.map(g => g.name).join(', ');
}

// 初始化
onMounted(() => {
  getBusinessUnits();
  getSSHCredentials();
  getServerGroups();
  getServers();
});
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <div class="flex gap-16px h-full">
      <!-- 左侧分组树 -->
      <div class="w-280px flex-shrink-0">
        <ElCard class="h-full" shadow="never">
          <template #header>
            <div class="flex justify-between items-center">
              <span class="font-bold">主机分组</span>
            </div>
          </template>
          <div class="h-full overflow-auto">
            <ElTree
              :data="serverGroups"
              node-key="id"
              :highlight-current="true"
              :current-node-key="selectedGroupId"
              default-expand-all
              :props="{ label: 'name', children: 'children' }"
              @node-click="handleGroupClick"
            >
              <template #default="{ node, data }">
                <div class="flex items-center justify-between w-full pr-8px">
                  <div class="flex items-center gap-8px">
                    <span class="w-8px h-8px rounded-full" :style="{ backgroundColor: data.color }" />
                    <span>{{ node.label }}</span>
                    <ElTag v-if="data.servers && data.servers.length > 0" size="small" type="info">
                      {{ data.servers.length }}
                    </ElTag>
                  </div>
                </div>
              </template>
            </ElTree>
            <div v-if="selectedGroupId" class="mt-12px">
              <ElButton size="small" plain @click="handleClearGroup">显示全部主机</ElButton>
            </div>
          </div>
        </ElCard>
      </div>

      <!-- 右侧主机列表 -->
      <div class="flex-1 min-w-0">
        <ElCard class="card-wrapper h-full">
          <!-- 搜索表单 -->
          <ElForm :model="searchForm" inline class="search-form">
            <ElFormItem label="主机名">
              <ElInput v-model="searchForm.hostname" placeholder="请输入主机名" clearable style="width: 200px" />
            </ElFormItem>
            <ElFormItem label="IP地址">
              <ElInput v-model="searchForm.ip" placeholder="请输入IP地址" clearable style="width: 200px" />
            </ElFormItem>
            <ElFormItem label="环境">
              <ElSelect v-model="searchForm.env" placeholder="请选择环境" clearable style="width: 150px">
                <ElOption v-for="option in envOptions" :key="option.value" :label="option.label" :value="option.value" />
              </ElSelect>
            </ElFormItem>
            <ElFormItem label="状态">
              <ElSelect v-model="searchForm.status" placeholder="请选择状态" clearable style="width: 150px">
                <ElOption v-for="option in statusOptions" :key="option.value" :label="option.label" :value="option.value" />
              </ElSelect>
            </ElFormItem>
            <ElFormItem label="服务商">
              <ElSelect v-model="searchForm.provider" placeholder="请选择服务商" clearable style="width: 150px">
                <ElOption v-for="option in providerOptions" :key="option.value" :label="option.label" :value="option.value" />
              </ElSelect>
            </ElFormItem>
            <ElFormItem>
              <ElButton type="primary" @click="handleSearch">搜索</ElButton>
              <ElButton @click="handleReset">重置</ElButton>
              <ElButton type="success" @click="handleAdd">新增</ElButton>
            </ElFormItem>
          </ElForm>

      <!-- 数据表格 -->
      <ElTable v-loading="loading" :data="tableData" border stripe class="h-full" height="calc(100vh - 400px)">
        <ElTableColumn prop="id" label="ID" width="70" align="center" />
        <ElTableColumn prop="hostname" label="主机名" min-width="150" align="center" show-overflow-tooltip />
        <ElTableColumn prop="ip" label="外网IP" width="140" align="center" />
        <ElTableColumn prop="innerIp" label="内网IP" width="140" align="center" />
        <ElTableColumn label="配置" width="140" align="center">
          <template #default="{ row }">
            <span>{{ row.cpu }}C / {{ row.memory }}G / {{ row.disk }}G</span>
          </template>
        </ElTableColumn>
        <ElTableColumn label="环境" width="80" align="center">
          <template #default="{ row }">
            <ElTag :type="getEnvTag(row.env).type" size="small">
              {{ getEnvTag(row.env).text }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="状态" width="80" align="center">
          <template #default="{ row }">
            <ElTag :type="getStatusTag(row.status).type" size="small">
              {{ getStatusTag(row.status).text }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="类型" width="90" align="center">
          <template #default="{ row }">
            <ElTag size="small">{{ getServerTypeText(row.serverType) }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="os" label="操作系统" min-width="120" align="center" show-overflow-tooltip />
        <ElTableColumn label="所属业务" min-width="120" align="center" show-overflow-tooltip>
          <template #default="{ row }">{{ getBusinessName(row.businessId) }}</template>
        </ElTableColumn>
        <ElTableColumn label="所属分组" min-width="120" align="center" show-overflow-tooltip>
          <template #default="{ row }">{{ getGroupNames(row.groups) }}</template>
        </ElTableColumn>
        <ElTableColumn prop="provider" label="服务商" width="100" align="center" />
        <ElTableColumn label="操作" width="260" align="center" fixed="right">
          <template #default="{ row }">
            <ElButton type="primary" size="small" @click="handleEdit(row)">编辑</ElButton>
            <ElButton type="info" size="small" @click="handleAssignGroup(row)">分配分组</ElButton>
            <ElButton type="danger" size="small" @click="handleDelete(row)">删除</ElButton>
          </template>
        </ElTableColumn>
      </ElTable>

      <!-- 分页 -->
      <div class="flex justify-end mt-16px">
        <ElPagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="handlePageChange"
          @size-change="handlePageSizeChange"
        />
      </div>
    </ElCard>
  </div>
</div>

    <!-- 新增/编辑对话框 -->
    <ElDialog v-model="dialogVisible" :title="dialogTitle" width="700px">
      <ElForm ref="serverFormRef" :model="serverForm" :rules="serverFormRules" label-width="120px">
        <!-- 主机类型选择 -->
        <ElFormItem label="主机类型">
          <ElRadioGroup v-model="serverType">
            <ElRadio value="normal">普通主机</ElRadio>
            <ElRadio value="cloud">云主机</ElRadio>
          </ElRadioGroup>
        </ElFormItem>

        <ElFormItem label="主机名" prop="hostname">
          <ElInput v-model="serverForm.hostname" placeholder="请输入主机名" />
        </ElFormItem>
        <ElFormItem label="连接IP" prop="ip">
          <ElInput v-model="serverForm.ip" placeholder="请输入连接IP" />
        </ElFormItem>
        <ElFormItem label="内网IP">
          <ElInput v-model="serverForm.innerIp" placeholder="请输入内网IP（可选）" />
        </ElFormItem>

        <!-- SSH凭证选择 -->
        <template v-if="serverType === 'cloud'">
          <ElFormItem label="云服务商">
            <ElSelect v-model="cloudForm.provider" style="width: 100%">
              <ElOption label="阿里云" value="aliyun" />
              <ElOption label="腾讯云" value="tencent" />
              <ElOption label="华为云" value="huawei" />
              <ElOption label="AWS" value="aws" />
              <ElOption label="其他" value="other" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem label="实例名称">
            <ElInput v-model="cloudForm.instanceName" placeholder="请输入实例名称" />
          </ElFormItem>
          <ElFormItem label="实例规格">
            <ElInput v-model="cloudForm.instanceType" placeholder="如: ecs.t6-c1m2.large" />
          </ElFormItem>
          <ElFormItem label="地域">
            <ElInput v-model="cloudForm.region" placeholder="如: cn-hangzhou" />
          </ElFormItem>
          <ElFormItem label="可用区">
            <ElInput v-model="cloudForm.zone" placeholder="如: cn-hangzhou-i" />
          </ElFormItem>
          <ElFormItem label="计费类型">
            <ElSelect v-model="cloudForm.chargeType" style="width: 100%">
              <ElOption label="按量付费" value="postpay" />
              <ElOption label="包年包月" value="prepay" />
            </ElSelect>
          </ElFormItem>
        </template>

        <!-- SSH凭证选择 -->
        <ElFormItem label="SSH凭证" prop="credentialId">
          <div class="flex gap-8px w-full">
            <ElSelect v-model="serverForm.credentialId" placeholder="请选择SSH凭证" class="flex-1">
              <ElOption
                v-for="cred in sshCredentials"
                :key="cred.id"
                :label="`${cred.name} (${cred.username})`"
                :value="cred.id"
              />
            </ElSelect>
            <ElButton type="primary" @click="handleManageCredentials">管理凭证</ElButton>
          </div>
        </ElFormItem>

        <ElFormItem label="SSH端口">
          <ElInputNumber v-model="serverForm.sshPort" :min="1" :max="65535" style="width: 100%" />
        </ElFormItem>

        <!-- 分组选择 -->
        <ElFormItem label="所属分组" prop="groupIds">
          <ElTreeSelect
            v-model="serverForm.groupIds"
            :data="serverGroups"
            :props="treeSelectProps"
            multiple
            show-checkbox check-strictly
            placeholder="请选择分组"
            style="width: 100%"
          />
        </ElFormItem>

        <ElFormItem label="备注">
          <ElInput v-model="serverForm.remarks" type="textarea" :rows="3" placeholder="请输入备注信息" />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="dialogVisible = false">取消</ElButton>
        <ElButton type="primary" @click="handleSave">保存</ElButton>
      </template>
    </ElDialog>


    <!-- 分配分组对话框 -->
    <ElDialog v-model="assignDialogVisible" title="分配主机分组" width="500px">
      <ElForm ref="assignFormRef" :model="assignForm" label-width="120px">
        <ElFormItem label="服务器">
          <span>{{ assignForm.serverName }}</span>
        </ElFormItem>
        <ElFormItem label="目标分组" prop="groupId">
          <ElTreeSelect
            v-model="assignForm.groupId"
            :data="serverGroups"
            :props="treeSelectProps"
            placeholder="请选择分组"
            clearable
            check-strictly
          />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="assignDialogVisible = false">取消</ElButton>
        <ElButton type="primary" @click="handleSaveAssign">确定</ElButton>
      </template>
    </ElDialog>

    <!-- SSH凭证管理对话框 -->
    <ElDialog v-model="credentialDialogVisible" title="SSH凭证管理" width="900px">
      <div class="flex gap-16px">
        <!-- 凭证列表 -->
        <div class="flex-1 border rounded p-12px">
          <div class="flex justify-between items-center mb-12px">
            <h3 class="text-16px font-bold">凭证列表</h3>
            <ElButton type="primary" size="small" @click="handleAddCredential">新增凭证</ElButton>
          </div>
          <ElTable :data="sshCredentials" border stripe size="small" max-height="400">
            <ElTableColumn prop="name" label="凭证名称" min-width="120" />
            <ElTableColumn prop="username" label="用户名" width="100" />
            <ElTableColumn label="认证类型" width="80" align="center">
              <template #default="{ row }">
                <ElTag :type="row.authType === 'password' ? 'primary' : 'success'" size="small">
                  {{ row.authType === 'password' ? '密码' : '密钥' }}
                </ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="port" label="端口" width="60" align="center" />
            <ElTableColumn label="操作" width="120" align="center">
              <template #default="{ row }">
                <ElButton type="primary" size="small" link @click="handleEditCredential(row)">编辑</ElButton>
                <ElButton type="danger" size="small" link @click="handleDeleteCredential(row)">删除</ElButton>
              </template>
            </ElTableColumn>
          </ElTable>
        </div>

        <!-- 凭证表单 -->
        <div v-if="credentialForm.name || sshCredentials.length === 0" class="w-360px border rounded p-12px">
          <h3 class="text-16px font-bold mb-12px">{{ credentialForm.id ? '编辑凭证' : '新增凭证' }}</h3>
          <ElForm ref="credentialFormRef" :model="credentialForm" :rules="credentialFormRules" label-width="100px">
            <ElFormItem label="凭证名称" prop="name">
              <ElInput v-model="credentialForm.name" placeholder="请输入凭证名称" />
            </ElFormItem>
            <ElFormItem label="用户名" prop="username">
              <ElInput v-model="credentialForm.username" placeholder="请输入SSH用户名" />
            </ElFormItem>
            <ElFormItem label="认证类型" prop="authType">
              <ElRadioGroup v-model="credentialForm.authType">
                <ElRadio value="password">密码认证</ElRadio>
                <ElRadio value="key">密钥认证</ElRadio>
              </ElRadioGroup>
            </ElFormItem>
            <ElFormItem v-if="credentialForm.authType === 'password'" label="密码" prop="password">
              <ElInput v-model="credentialForm.password" type="password" placeholder="请输入密码" show-password />
            </ElFormItem>
            <ElFormItem v-if="credentialForm.authType === 'key'" label="私钥" prop="privateKey">
              <ElInput v-model="credentialForm.privateKey" type="textarea" :rows="6" placeholder="请输入私钥内容" />
            </ElFormItem>
            <ElFormItem v-if="credentialForm.authType === 'key'" label="私钥密码">
              <ElInput v-model="credentialForm.passphrase" type="password" placeholder="如果私钥有密码请输入" show-password />
            </ElFormItem>
            <ElFormItem label="SSH端口">
              <ElInputNumber v-model="credentialForm.port" :min="1" :max="65535" style="width: 100%" />
            </ElFormItem>
            <ElFormItem label="描述">
              <ElInput v-model="credentialForm.description" type="textarea" :rows="2" placeholder="请输入描述" />
            </ElFormItem>
            <ElFormItem>
              <ElButton type="primary" @click="handleSaveCredential">保存</ElButton>
              <ElButton @click="credentialForm.name = ''">取消</ElButton>
            </ElFormItem>
          </ElForm>
        </div>
      </div>
    </ElDialog>
  </div>
</template>

<style scoped lang="scss">
.search-form {
  :deep(.el-form-item) {
    margin-bottom: 12px;
  }
}
</style>
