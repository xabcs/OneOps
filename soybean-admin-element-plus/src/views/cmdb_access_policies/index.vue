<script setup lang="ts">
import { computed, onMounted, ref, h } from 'vue';
import {
  fetchGetAccessPolicies,
  fetchCreateAccessPolicy,
  fetchUpdateAccessPolicy,
  fetchDeleteAccessPolicy
} from '@/service/api/cmdb';
import { $t } from '@/locales';

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

// 星期选项
const weekdayOptions = [
  { label: '周一', value: 1 },
  { label: '周二', value: 2 },
  { label: '周三', value: 3 },
  { label: '周四', value: 4 },
  { label: '周五', value: 5 },
  { label: '周六', value: 6 },
  { label: '周日', value: 7 }
];

// 获取策略列表
async function getPolicies() {
  loading.value = true;
  try {
    const result = await fetchGetAccessPolicies({
      page: pagination.value.page,
      pageSize: pagination.value.pageSize
    });
    policies.value = result.list;
    total.value = result.total;
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
  const confirmed = await window.$confirm?.(
    `确定要删除策略"${policy.name}"吗？`,
    '确认删除',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  );

  if (!confirmed) return;

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
    if (error !== false) { // 表单验证失败时会返回 false
      window.$message?.error(error.message || '操作失败');
    }
  }
}

// 关闭对话框
function handleCloseDialog() {
  dialogVisible.value = false;
  formRef.value?.resetFields();
}

// 添加登录账号
function handleAddLoginAccount() {
  if (!currentPolicy.value.loginAccounts) {
    currentPolicy.value.loginAccounts = [];
  }
  currentPolicy.value.loginAccounts.push('');
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
  return h('el-tag', {
    type: policy.status === 1 ? 'success' : 'info',
    size: 'small'
  }, policy.status === 1 ? '启用' : '禁用');
}

onMounted(() => {
  getPolicies();
});
</script>

<template>
  <div class="access-policies-page">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">访问策略管理</span>
          <el-button type="primary" :icon="ICON_REGISTRY.plus" @click="handleCreate">
            新增策略
          </el-button>
        </div>
      </template>

      <!-- 策略列表 -->
      <el-table
        v-loading="loading"
        :data="policies"
        stripe
        style="width: 100%"
      >
        <el-table-column prop="id" label="ID" width="60" />

        <el-table-column prop="name" label="策略名称" min-width="150" />

        <el-table-column label="授权对象" width="150">
          <template #default="{ row }">
            <el-tag size="small" type="primary">
              {{ row.subjectType === 'user' ? '用户' : row.subjectType === 'role' ? '角色' : '用户组' }}
            </el-tag>
            <span style="margin-left: 8px">ID: {{ row.subjectId }}</span>
          </template>
        </el-table-column>

        <el-table-column label="资产范围" width="150">
          <template #default="{ row }">
            <el-tag size="small" type="success">
              {{ row.assetScopeType === 'all' ? '全部' : row.assetScopeType }}
            </el-tag>
            <span v-if="row.assetScopeType !== 'all'" style="margin-left: 8px">
              ID: {{ row.assetScopeId }}
            </span>
          </template>
        </el-table-column>

        <el-table-column label="允许账号" width="200">
          <template #default="{ row }">
            <el-tag
              v-for="(account, idx) in (row.loginAccounts || []).slice(0, 2)"
              :key="idx"
              size="small"
              style="margin-right: 4px"
            >
              {{ account }}
            </el-tag>
            <span v-if="(row.loginAccounts || []).length > 2" style="font-size: 12px; color: #909399">
              +{{ (row.loginAccounts || []).length - 2 }}
            </span>
          </template>
        </el-table-column>

        <el-table-column label="允许协议" width="120">
          <template #default="{ row }">
            <el-tag
              v-for="(protocol, idx) in (row.protocols || [])"
              :key="idx"
              size="small"
              :type="protocol === 'ssh' ? 'primary' : 'success'"
              style="margin-right: 4px"
            >
              {{ protocol.toUpperCase() }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="时间窗口" width="180">
          <template #default="{ row }">
            {{ formatTimeWindow(row.timeWindow) }}
          </template>
        </el-table-column>

        <el-table-column label="需要审批" width="100">
          <template #default="{ row }">
            <el-tag :type="row.requireApproval ? 'warning' : 'info'" size="small">
              {{ row.requireApproval ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <component :is="() => getStatusTag(row)" />
          </template>
        </el-table-column>

        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button
              type="primary"
              size="small"
              @click="handleEdit(row)"
            >
              编辑
            </el-button>
            <el-button
              :type="row.status === 1 ? 'warning' : 'success'"
              size="small"
              @click="handleToggleStatus(row)"
            >
              {{ row.status === 1 ? '禁用' : '启用' }}
            </el-button>
            <el-button
              type="danger"
              size="small"
              @click="handleDelete(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="getPolicies"
          @current-change="getPolicies"
        />
      </div>
    </el-card>

    <!-- 策略表单对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '新增策略' : '编辑策略'"
      width="700px"
      :close-on-click-modal="false"
      @close="handleCloseDialog"
    >
      <el-form
        ref="formRef"
        :model="currentPolicy"
        :rules="rules"
        label-width="120px"
      >
        <el-form-item label="策略名称" prop="name">
          <el-input
            v-model="currentPolicy.name"
            placeholder="请输入策略名称"
            maxlength="100"
            show-word-limit
          />
        </el-form-item>

        <el-form-item label="授权对象类型" prop="subjectType">
          <el-select
            v-model="currentPolicy.subjectType"
            placeholder="选择授权对象类型"
            style="width: 100%"
          >
            <el-option
              v-for="option in subjectTypeOptions"
              :key="option.value"
              :value="option.value"
              :label="option.label"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="授权对象ID" prop="subjectId">
          <el-input
            v-model.number="currentPolicy.subjectId"
            type="number"
            placeholder="请输入授权对象ID"
          />
        </el-form-item>

        <el-form-item label="资产范围类型" prop="assetScopeType">
          <el-select
            v-model="currentPolicy.assetScopeType"
            placeholder="选择资产范围类型"
            style="width: 100%"
          >
            <el-option
              v-for="option in assetScopeTypeOptions"
              :key="option.value"
              :value="option.value"
              :label="option.label"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="资产范围ID" prop="assetScopeId">
          <el-input
            v-model.number="currentPolicy.assetScopeId"
            type="number"
            placeholder="选择全部资产时可填0"
            :disabled="currentPolicy.assetScopeType === 'all'"
          />
        </el-form-item>

        <el-form-item label="允许的登录账号">
          <div class="tags-input-wrapper">
            <el-tag
              v-for="(account, idx) in currentPolicy.loginAccounts"
              :key="idx"
              closable
              @close="handleRemoveLoginAccount(idx)"
              style="margin-right: 8px; margin-bottom: 8px"
            >
              {{ account }}
            </el-tag>
            <el-input
              v-if="!currentPolicy.loginAccounts || currentPolicy.loginAccounts.length === 0"
              v-model="currentPolicy.loginAccounts[0]"
              placeholder="输入账号后按回车"
              size="small"
              style="width: 150px"
            />
          </div>
        </el-form-item>

        <el-form-item label="允许的协议">
          <el-checkbox-group v-model="currentPolicy.protocols">
            <el-checkbox
              v-for="option in protocolOptions"
              :key="option.value"
              :value="option.value"
              :label="option.label"
            />
          </el-checkbox-group>
        </el-form-item>

        <el-form-item label="文件传输">
          <el-switch v-model="currentPolicy.allowFileTransfer" />
          <span style="margin-left: 8px">允许文件传输</span>
        </el-form-item>

        <el-form-item label="Sudo 权限">
          <el-switch v-model="currentPolicy.allowSudo" />
          <span style="margin-left: 8px">允许 sudo</span>
        </el-form-item>

        <el-form-item label="需要审批">
          <el-switch v-model="currentPolicy.requireApproval" />
          <span style="margin-left: 8px">连接前需要审批</span>
        </el-form-item>

        <el-form-item label="高危命令">
          <div class="tags-input-wrapper">
            <el-tag
              v-for="(cmd, idx) in currentPolicy.highRiskCommands"
              :key="idx"
              closable
              type="danger"
              @close="handleRemoveHighRiskCommand(idx)"
              style="margin-right: 8px; margin-bottom: 8px"
            >
              {{ cmd }}
            </el-tag>
            <el-button
              size="small"
              :icon="ICON_REGISTRY.plus"
              @click="handleAddHighRiskCommand"
            >
              添加
            </el-button>
          </div>
        </el-form-item>

        <el-form-item label="状态">
          <el-radio-group v-model="currentPolicy.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="handleCloseDialog">取消</el-button>
        <el-button type="primary" @click="handleSubmit">
          确定
        </el-button>
      </template>
    </el-dialog>
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
