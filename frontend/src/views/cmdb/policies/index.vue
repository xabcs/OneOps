<script setup lang="ts">
import { computed, h, onMounted, ref } from 'vue';
import { ElMessageBox } from 'element-plus';
import {
  fetchCreateAccessPolicy,
  fetchDeleteAccessPolicy,
  fetchGetAccessPolicies,
  fetchUpdateAccessPolicy
} from '@/service/api/cmdb';
import {
  fetchGetAllRoles,
  fetchGetBusinessUnits,
  fetchGetServerGroups,
  fetchGetServerTags,
  fetchGetUserList
} from '@/service/api';

defineOptions({
  name: 'CMDBAccessPolicies'
});

const loading = ref(false);
const policies = ref<Bastion.AccessPolicy[]>([]);
const total = ref(0);
const dialogVisible = ref(false);
const dialogMode = ref<'create' | 'edit'>('create');
const currentPolicy = ref<Partial<Bastion.AccessPolicyForm>>({});

// 分页
const pagination = ref({
  page: 1,
  pageSize: 20
});

// 表单引用
const formRef = ref();

// 表单验证规则
const rules = {
  name: [{ required: true, message: '请输入策略名称', trigger: 'blur' }],
  subjectType: [{ required: true, message: '请选择授权对象类型', trigger: 'change' }],
  subjectId: [{ required: true, message: '请选择授权对象', trigger: 'change' }],
  assetScopeType: [{ required: true, message: '请选择资产范围类型', trigger: 'change' }],
  assetScopeId: [{ required: true, message: '请选择资产范围', trigger: 'change' }]
};

// 授权对象类型选项
const subjectTypeOptions = [
  { label: '用户', value: 'user' },
  { label: '角色', value: 'role' },
  { label: '用户组', value: 'user_group' }
];

// 资产范围类型选项
const assetScopeTypeOptions = [
  { label: '全部资产', value: 'all' },
  { label: '单台服务器', value: 'server' },
  { label: '主机分组', value: 'group' },
  { label: '业务系统', value: 'business' },
  { label: '标签', value: 'tag' }
];

// 协议选项
const protocolOptions = [
  { label: 'SSH', value: 'ssh' },
  { label: 'SFTP', value: 'sftp' }
];

// ========== 资产数据 ==========
const users = ref<Api.SystemManage.User[]>([]);
const roles = ref<Api.SystemManage.AllRole[]>([]);
const serverGroups = ref<CMDB.ServerGroup[]>([]);
const businessUnits = ref<CMDB.BusinessUnit[]>([]);
const serverTags = ref<CMDB.ServerTag[]>([]);

// 获取策略列表
async function getPolicies() {
  loading.value = true;
  try {
    const { data } = await fetchGetAccessPolicies({
      page: pagination.value.page,
      pageSize: pagination.value.pageSize
    });
    policies.value = data?.list || [];
    total.value = data?.total || 0;
  } catch (error) {
    window.$message?.error('获取策略列表失败');
  } finally {
    loading.value = false;
  }
}

// 新增策略
function handleCreate() {
  dialogMode.value = 'create';
  currentPolicy.value = {
    name: '',
    subjectType: 'role',
    subjectId: [],
    assetScopeType: 'all',
    assetScopeId: 0,
    loginAccounts: ['root'],
    protocols: ['ssh', 'sftp'],
    allowFileTransfer: true,
    allowSudo: false,
    requireApproval: false,
    timeWindow: undefined,
    highRiskCommands: [],
    status: 1
  };
  dialogVisible.value = true;
}

// 编辑策略
function handleEdit(policy: Bastion.AccessPolicy) {
  dialogMode.value = 'edit';
  currentPolicy.value = { ...policy };
  dialogVisible.value = true;
}

// 删除策略
async function handleDelete(policy: Bastion.AccessPolicy) {
  try {
    await ElMessageBox.confirm(`确定要删除策略"${policy.name}"吗？`, '确认删除', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    });
  } catch {
    return;
  }

  try {
    await fetchDeleteAccessPolicy(policy.id);
    window.$message?.success('删除成功');
    getPolicies();
  } catch (error: any) {
    window.$message?.error(error.message || '删除失败');
  }
}

// 切换策略状态
async function handleToggleStatus(policy: Bastion.AccessPolicy) {
  const newStatus = policy.status === 1 ? 0 : 1;
  try {
    await fetchUpdateAccessPolicy(policy.id, { status: newStatus });
    window.$message?.success('状态更新成功');
    getPolicies();
  } catch (error: any) {
    window.$message?.error(error.message || '状态更新失败');
  }
}

// 提交表单
async function handleSubmit() {
  try {
    await formRef.value?.validate();

    const data = { ...currentPolicy.value };

    if (dialogMode.value === 'create') {
      await fetchCreateAccessPolicy(data as Bastion.AccessPolicyForm);
      window.$message?.success('创建成功');
    } else {
      await fetchUpdateAccessPolicy(currentPolicy.value.id!, data);
      window.$message?.success('更新成功');
    }

    dialogVisible.value = false;
    getPolicies();
  } catch (error: any) {
    if (error !== false) {
      // 表单验证失败时会返回 false
      window.$message?.error(error.message || '操作失败');
    }
  }
}

// 关闭对话框
function handleCloseDialog() {
  dialogVisible.value = false;
  formRef.value?.resetFields();
}

// 移除登录账号
function handleRemoveLoginAccount(index: number) {
  currentPolicy.value.loginAccounts?.splice(index, 1);
}

// 添加高危命令
function handleAddHighRiskCommand() {
  if (!currentPolicy.value.highRiskCommands) {
    currentPolicy.value.highRiskCommands = [];
  }
  currentPolicy.value.highRiskCommands.push('');
}

// 移除高危命令
function handleRemoveHighRiskCommand(index: number) {
  currentPolicy.value.highRiskCommands?.splice(index, 1);
}

// 格式化时间窗口
function formatTimeWindow(timeWindow?: Bastion.AccessPolicy['timeWindow']): string {
  if (!timeWindow) return '-';
  const days = ['周一', '周二', '周三', '周四', '周五', '周六', '周日'];
  const dayNames = (timeWindow.days || []).map(d => days[d - 1]).join('、');
  return `${timeWindow.start}-${timeWindow.end} (${dayNames})`;
}

// 获取状态标签
function getStatusTag(policy: Bastion.AccessPolicy) {
  return h(
    'el-tag',
    {
      type: policy.status === 1 ? 'success' : 'info',
      size: 'small'
    },
    policy.status === 1 ? '启用' : '禁用'
  );
}

// ========== 获取资产数据 ==========
async function getAssetData() {
  try {
    const [usersRes, rolesRes, groupsRes, businessRes, tagsRes] = await Promise.allSettled([
      fetchGetUserList({ current: 1, size: 1000 }),
      fetchGetAllRoles(),
      fetchGetServerGroups(),
      fetchGetBusinessUnits(),
      fetchGetServerTags()
    ]);

    if (usersRes.status === 'fulfilled') {
      users.value = usersRes.value.data?.records || [];
    }
    if (rolesRes.status === 'fulfilled') {
      roles.value = rolesRes.value.data || [];
    }
    if (groupsRes.status === 'fulfilled') {
      serverGroups.value = groupsRes.value.data || [];
    }
    if (businessRes.status === 'fulfilled') {
      businessUnits.value = businessRes.value.data || [];
    }
    if (tagsRes.status === 'fulfilled') {
      serverTags.value = tagsRes.value.data || [];
    }
  } catch (error) {
    console.error('获取资产数据失败:', error);
  }
}

// ========== 授权对象相关 ==========
const userOptions = computed(() => {
  return users.value.map(u => ({ label: `${u.username} (${u.email || '无邮箱'})`, value: u.id }));
});

const roleOptions = computed(() => {
  return roles.value.map(r => ({ label: r.name, value: r.id }));
});

// ========== 资产范围相关 ==========
const groupOptions = computed(() => {
  return serverGroups.value;
});

const businessOptions = computed(() => {
  return businessUnits.value;
});

const tagOptions = computed(() => {
  return serverTags.value.map(t => ({ label: t.name, value: t.id }));
});

// ========== 辅助函数 ==========
// 获取授权对象名称
function getSubjectName(type: string, id: number): string {
  if (type === 'user') {
    const user = users.value.find(u => u.id === id);
    return user ? user.username : `ID: ${id}`;
  }
  if (type === 'role') {
    const role = roles.value.find(r => r.id === id);
    return role ? role.name : `ID: ${id}`;
  }
  return `ID: ${id}`;
}

// 获取资产范围类型名称
function getAssetScopeTypeName(type: string): string {
  const typeMap: Record<string, string> = {
    all: '全部资产',
    server: '单台服务器',
    group: '主机分组',
    business: '业务系统',
    tag: '标签'
  };
  return typeMap[type] || type;
}

// 获取资产范围名称
function getAssetScopeName(type: string, id: number): string {
  if (type === 'group') {
    const findGroup = (groups: CMDB.ServerGroup[], targetId: number): string => {
      for (const group of groups) {
        if (group.id === targetId) return group.name;
        if (group.children) {
          const found = findGroup(group.children, targetId);
          if (found) return found;
        }
      }
      return '';
    };
    const name = findGroup(serverGroups.value, id);
    return name || `ID: ${id}`;
  }
  if (type === 'business') {
    const findBusiness = (units: CMDB.BusinessUnit[], targetId: number): string => {
      for (const unit of units) {
        if (unit.id === targetId) return unit.name;
        if (unit.children) {
          const found = findBusiness(unit.children, targetId);
          if (found) return found;
        }
      }
      return '';
    };
    const name = findBusiness(businessUnits.value, id);
    return name || `ID: ${id}`;
  }
  if (type === 'tag') {
    const tag = serverTags.value.find(t => t.id === id);
    return tag ? tag.name : `ID: ${id}`;
  }
  return `ID: ${id}`;
}

onMounted(() => {
  getPolicies();
  getAssetData();
});
</script>

<template>
  <div class="access-policies-page">
    <ElCard shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">访问策略管理</span>
          <ElButton type="primary" @click="handleCreate">新增策略</ElButton>
        </div>
      </template>

      <!-- 策略列表 -->
      <ElTable v-loading="loading" :data="policies" stripe style="width: 100%">
        <ElTableColumn prop="id" label="ID" width="60" />

        <ElTableColumn prop="name" label="策略名称" min-width="150" />

        <ElTableColumn label="授权对象" width="200">
          <template #default="{ row }">
            <ElTag size="small" type="primary">
              {{ row.subjectType === 'user' ? '用户' : row.subjectType === 'role' ? '角色' : '用户组' }}
            </ElTag>
            <span style="margin-left: 8px">
              {{ getSubjectName(row.subjectType, row.subjectId) }}
            </span>
          </template>
        </ElTableColumn>

        <ElTableColumn label="资产范围" width="200">
          <template #default="{ row }">
            <ElTag size="small" type="success">
              {{ getAssetScopeTypeName(row.assetScopeType) }}
            </ElTag>
            <span v-if="row.assetScopeType !== 'all'" style="margin-left: 8px">
              {{ getAssetScopeName(row.assetScopeType, row.assetScopeId) }}
            </span>
          </template>
        </ElTableColumn>

        <ElTableColumn label="允许账号" width="200">
          <template #default="{ row }">
            <ElTag
              v-for="(account, idx) in (row.loginAccounts || []).slice(0, 2)"
              :key="idx"
              size="small"
              style="margin-right: 4px"
            >
              {{ account }}
            </ElTag>
            <span v-if="(row.loginAccounts || []).length > 2" style="font-size: 12px; color: #909399">
              +{{ (row.loginAccounts || []).length - 2 }}
            </span>
          </template>
        </ElTableColumn>

        <ElTableColumn label="允许协议" width="120">
          <template #default="{ row }">
            <ElTag
              v-for="(protocol, idx) in row.protocols || []"
              :key="idx"
              size="small"
              :type="protocol === 'ssh' ? 'primary' : 'success'"
              style="margin-right: 4px"
            >
              {{ protocol.toUpperCase() }}
            </ElTag>
          </template>
        </ElTableColumn>

        <ElTableColumn label="时间窗口" width="180">
          <template #default="{ row }">
            {{ formatTimeWindow(row.timeWindow) }}
          </template>
        </ElTableColumn>

        <ElTableColumn label="需要审批" width="100">
          <template #default="{ row }">
            <ElTag :type="row.requireApproval ? 'warning' : 'info'" size="small">
              {{ row.requireApproval ? '是' : '否' }}
            </ElTag>
          </template>
        </ElTableColumn>

        <ElTableColumn label="状态" width="80">
          <template #default="{ row }">
            <component :is="() => getStatusTag(row)" />
          </template>
        </ElTableColumn>

        <ElTableColumn label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <ElButton type="primary" size="small" @click="handleEdit(row)">编辑</ElButton>
            <ElButton :type="row.status === 1 ? 'warning' : 'success'" size="small" @click="handleToggleStatus(row)">
              {{ row.status === 1 ? '禁用' : '启用' }}
            </ElButton>
            <ElButton type="danger" size="small" @click="handleDelete(row)">删除</ElButton>
          </template>
        </ElTableColumn>
      </ElTable>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <ElPagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="getPolicies"
          @current-change="getPolicies"
        />
      </div>
    </ElCard>

    <!-- 策略表单对话框 -->
    <ElDialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '新增策略' : '编辑策略'"
      width="700px"
      :close-on-click-modal="false"
      @close="handleCloseDialog"
    >
      <ElForm ref="formRef" :model="currentPolicy" :rules="rules" label-width="120px">
        <ElFormItem label="策略名称" prop="name">
          <ElInput v-model="currentPolicy.name" placeholder="请输入策略名称" maxlength="100" show-word-limit />
        </ElFormItem>

        <ElFormItem label="授权对象类型" prop="subjectType">
          <ElSelect v-model="currentPolicy.subjectType" placeholder="选择授权对象类型" style="width: 100%">
            <ElOption
              v-for="option in subjectTypeOptions"
              :key="option.value"
              :value="option.value"
              :label="option.label"
            />
          </ElSelect>
        </ElFormItem>

        <!-- 授权对象为用户时 -->
        <ElFormItem v-if="currentPolicy.subjectType === 'user'" label="授权对象" prop="subjectId">
          <ElSelect v-model="currentPolicy.subjectId" placeholder="请选择用户" filterable style="width: 100%">
            <ElOption v-for="option in userOptions" :key="option.value" :label="option.label" :value="option.value" />
          </ElSelect>
        </ElFormItem>

        <!-- 授权对象为角色时 -->
        <ElFormItem v-if="currentPolicy.subjectType === 'role'" label="授权对象" prop="subjectId">
          <ElSelect v-model="currentPolicy.subjectId" placeholder="请选择角色" style="width: 100%">
            <ElOption v-for="option in roleOptions" :key="option.value" :label="option.label" :value="option.value" />
          </ElSelect>
        </ElFormItem>

        <!-- 授权对象为用户组时（暂时不支持，显示提示） -->
        <ElFormItem v-if="currentPolicy.subjectType === 'user_group'" label="授权对象" prop="subjectId">
          <ElInput placeholder="用户组功能暂未开放" disabled />
        </ElFormItem>

        <ElFormItem label="资产范围类型" prop="assetScopeType">
          <ElSelect v-model="currentPolicy.assetScopeType" placeholder="选择资产范围类型" style="width: 100%">
            <ElOption
              v-for="option in assetScopeTypeOptions"
              :key="option.value"
              :value="option.value"
              :label="option.label"
            />
          </ElSelect>
        </ElFormItem>

        <!-- 全部资产时不需要选择 -->
        <ElFormItem v-if="currentPolicy.assetScopeType === 'all'" label="资产范围">
          <ElInput value="全部资产" disabled />
        </ElFormItem>

        <!-- 资产范围为主机分组时 -->
        <ElFormItem v-if="currentPolicy.assetScopeType === 'group'" label="资产范围" prop="assetScopeId">
          <ElTreeSelect
            v-model="currentPolicy.assetScopeId"
            :data="groupOptions"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            placeholder="请选择主机分组"
            check-strictly
            style="width: 100%"
          />
        </ElFormItem>

        <!-- 资产范围为业务系统时 -->
        <ElFormItem v-if="currentPolicy.assetScopeType === 'business'" label="资产范围" prop="assetScopeId">
          <ElTreeSelect
            v-model="currentPolicy.assetScopeId"
            :data="businessOptions"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            placeholder="请选择业务系统"
            check-strictly
            style="width: 100%"
          />
        </ElFormItem>

        <!-- 资产范围为标签时 -->
        <ElFormItem v-if="currentPolicy.assetScopeType === 'tag'" label="资产范围" prop="assetScopeId">
          <ElSelect v-model="currentPolicy.assetScopeId" placeholder="请选择标签" style="width: 100%">
            <ElOption v-for="option in tagOptions" :key="option.value" :label="option.label" :value="option.value" />
          </ElSelect>
        </ElFormItem>

        <!-- 资产范围为单台服务器时（暂时不支持，显示提示） -->
        <ElFormItem v-if="currentPolicy.assetScopeType === 'server'" label="资产范围" prop="assetScopeId">
          <ElInput placeholder="单台服务器选择功能即将开放" disabled />
        </ElFormItem>

        <ElFormItem label="允许的登录账号">
          <div class="tags-input-wrapper">
            <ElTag
              v-for="(account, idx) in currentPolicy.loginAccounts"
              :key="idx"
              closable
              style="margin-right: 8px; margin-bottom: 8px"
              @close="handleRemoveLoginAccount(idx)"
            >
              {{ account }}
            </ElTag>
            <ElInput
              v-if="!currentPolicy.loginAccounts || currentPolicy.loginAccounts.length === 0"
              placeholder="输入账号后按回车"
              size="small"
              style="width: 150px"
              @change="
                (val: string) => {
                  if (val) {
                    currentPolicy.loginAccounts = [val];
                  }
                }
              "
            />
          </div>
        </ElFormItem>

        <ElFormItem label="允许的协议">
          <ElCheckboxGroup v-model="currentPolicy.protocols">
            <ElCheckbox
              v-for="option in protocolOptions"
              :key="option.value"
              :value="option.value"
              :label="option.label"
            />
          </ElCheckboxGroup>
        </ElFormItem>

        <ElFormItem label="文件传输">
          <ElSwitch v-model="currentPolicy.allowFileTransfer" />
          <span style="margin-left: 8px">允许文件传输</span>
        </ElFormItem>

        <ElFormItem label="Sudo 权限">
          <ElSwitch v-model="currentPolicy.allowSudo" />
          <span style="margin-left: 8px">允许 sudo</span>
        </ElFormItem>

        <ElFormItem label="需要审批">
          <ElSwitch v-model="currentPolicy.requireApproval" />
          <span style="margin-left: 8px">连接前需要审批</span>
        </ElFormItem>

        <ElFormItem label="高危命令">
          <div class="tags-input-wrapper">
            <ElTag
              v-for="(cmd, idx) in currentPolicy.highRiskCommands"
              :key="idx"
              closable
              type="danger"
              style="margin-right: 8px; margin-bottom: 8px"
              @close="handleRemoveHighRiskCommand(idx)"
            >
              {{ cmd }}
            </ElTag>
            <ElButton size="small" @click="handleAddHighRiskCommand">添加</ElButton>
          </div>
        </ElFormItem>

        <ElFormItem label="状态">
          <ElRadioGroup v-model="currentPolicy.status">
            <ElRadio :value="1">启用</ElRadio>
            <ElRadio :value="0">禁用</ElRadio>
          </ElRadioGroup>
        </ElFormItem>
      </ElForm>

      <template #footer>
        <ElButton @click="handleCloseDialog">取消</ElButton>
        <ElButton type="primary" @click="handleSubmit">确定</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<style scoped>
.access-policies-page {
  padding: 16px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.title {
  font-size: 16px;
  font-weight: 500;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 16px;
}

.tags-input-wrapper {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
}
</style>
