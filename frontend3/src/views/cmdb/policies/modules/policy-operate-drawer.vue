<script setup lang="ts">
  import { computed, nextTick, ref, watch } from 'vue';
  import type { FormInstance, FormRules } from 'element-plus';
  import { fetchCreateAccessPolicy, fetchServerOptions, fetchUpdateAccessPolicy } from '@/service/api/cmdb';
  import {
    fetchGetBusinessUnits,
    fetchGetServerGroups,
    fetchGetServerTags,
    fetchRoleOptions,
    fetchUserOptions
  } from '@/service/api';

  defineOptions({ name: 'PolicyOperateDrawer' });

  interface Props {
    operateType: UI.TableOperateType;
    rowData?: Bastion.AccessPolicy | null;
  }

  const props = defineProps<Props>();

  interface Emits {
    (e: 'submitted'): void;
  }

  const emit = defineEmits<Emits>();

  const visible = defineModel<boolean>('visible', { default: false });

  const title = computed(() => (props.operateType === 'add' ? '新增策略' : '编辑策略'));

  // 授权对象类型选项（user_group 暂未实现，不提供创建入口；编辑历史数据时只读展示）
  const subjectTypeOptions = [
    { label: '用户', value: 'user' },
    { label: '角色', value: 'role' }
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

  // 星期选项（1=周一，与后端 TimeWindow.Days 一致）
  const dayOptions = [
    { label: '周一', value: 1 },
    { label: '周二', value: 2 },
    { label: '周三', value: 3 },
    { label: '周四', value: 4 },
    { label: '周五', value: 5 },
    { label: '周六', value: 6 },
    { label: '周日', value: 7 }
  ];

  // ========== 资产数据 ==========
  const users = ref<{ id: number; username: string; nickname: string }[]>([]);
  const roles = ref<{ id: number; name: string; code: string }[]>([]);
  const serverGroups = ref<CMDB.ServerGroup[]>([]);
  const businessUnits = ref<CMDB.BusinessUnit[]>([]);
  const serverTags = ref<CMDB.ServerTag[]>([]);
  const servers = ref<{ id: number; hostname: string; ip: string }[]>([]);

  const userOptions = computed(() => {
    return users.value.map(u => ({
      label: `${u.username}${u.nickname ? ` (${u.nickname})` : ''}`,
      value: u.id
    }));
  });

  const roleOptions = computed(() => {
    return roles.value.map(r => ({ label: r.name, value: r.id }));
  });

  const serverOptions = computed(() => {
    return servers.value.map(s => ({ label: `${s.hostname} (${s.ip})`, value: s.id }));
  });

  const tagOptions = computed(() => serverTags.value.map(t => ({ label: t.name, value: t.id })));

  type PolicyModel = Partial<Bastion.AccessPolicyForm>;

  function createDefaultModel(): PolicyModel {
    return {
      name: '',
      subjectType: 'role',
      subjectId: undefined,
      assetScopeType: 'all',
      assetScopeId: undefined,
      loginAccounts: ['root'],
      protocols: ['ssh', 'sftp'],
      allowFileTransfer: true,
      allowSudo: false,
      highRiskCommands: [],
      status: 1
    };
  }

  const model = ref<PolicyModel>(createDefaultModel());

  // 时间窗口（内部始终持有完整结构，提交时按开关决定是否携带）
  const timeEnabled = ref(false);
  const timeWindow = ref<{ start: string; end: string; days: number[] }>({ start: '', end: '', days: [] });

  function handleInitModel() {
    model.value = createDefaultModel();
    timeEnabled.value = false;
    timeWindow.value = { start: '', end: '', days: [] };

    if (props.operateType === 'edit' && props.rowData) {
      const row = props.rowData;
      model.value = {
        ...row,
        subjectId: row.subjectId || undefined,
        assetScopeId: row.assetScopeType === 'all' ? undefined : row.assetScopeId || undefined
      };
      if (row.timeWindow?.start || row.timeWindow?.end || row.timeWindow?.days?.length) {
        timeEnabled.value = true;
        timeWindow.value = {
          start: row.timeWindow?.start || '',
          end: row.timeWindow?.end || '',
          days: row.timeWindow?.days || []
        };
      }
    }
  }

  async function loadAssetData() {
    const [usersRes, rolesRes, groupsRes, businessRes, tagsRes, serversRes] = await Promise.allSettled([
      fetchUserOptions(),
      fetchRoleOptions(),
      fetchGetServerGroups(),
      fetchGetBusinessUnits(),
      fetchGetServerTags(),
      fetchServerOptions()
    ]);

    if (usersRes.status === 'fulfilled') users.value = usersRes.value.data || [];
    if (rolesRes.status === 'fulfilled') roles.value = rolesRes.value.data || [];
    if (groupsRes.status === 'fulfilled') serverGroups.value = groupsRes.value.data?.groups || [];
    if (businessRes.status === 'fulfilled') businessUnits.value = businessRes.value.data || [];
    if (tagsRes.status === 'fulfilled') serverTags.value = tagsRes.value.data || [];
    if (serversRes.status === 'fulfilled') servers.value = serversRes.value.data || [];
  }

  // ========== 表单校验 ==========
  const formRef = ref<FormInstance | null>(null);

  const rules: FormRules = {
    name: [
      { required: true, message: '请输入策略名称', trigger: 'blur' },
      { max: 100, message: '策略名称不能超过 100 个字符', trigger: 'blur' }
    ],
    subjectType: [{ required: true, message: '请选择授权对象类型' }],
    subjectId: [
      {
        required: true,
        validator: (_rule, value: number | undefined, callback) => {
          if (!value || value <= 0) callback(new Error('请选择授权对象'));
          else callback();
        },
        trigger: 'change'
      }
    ],
    assetScopeType: [{ required: true, message: '请选择资产范围类型' }],
    assetScopeId: [
      {
        validator: (_rule, value: number | undefined, callback) => {
          if (model.value.assetScopeType === 'all') {
            callback();
            return;
          }
          if (!value || value <= 0) callback(new Error('请选择资产范围'));
          else callback();
        },
        trigger: 'change'
      }
    ],
    protocols: [
      {
        validator: (_rule, value: string[] | undefined, callback) => {
          if (!value || value.length === 0) callback(new Error('请至少选择一种协议'));
          else callback();
        },
        trigger: 'change'
      }
    ],
    loginAccounts: [
      {
        validator: (_rule, value: string[] | undefined, callback) => {
          if (!value || value.filter(Boolean).length === 0) {
            callback(new Error('请至少添加一个登录账号'));
          } else callback();
        },
        trigger: 'change'
      }
    ],
    timeWindow: [
      {
        validator: (_rule, _value, callback) => {
          if (!timeEnabled.value) {
            callback();
            return;
          }
          if (!timeWindow.value.start || !timeWindow.value.end) {
            callback(new Error('请选择完整的起止时间'));
            return;
          }
          if (timeWindow.value.start >= timeWindow.value.end) {
            callback(new Error('开始时间必须早于结束时间'));
            return;
          }
          callback();
        },
        trigger: 'change'
      }
    ]
  };

  function closeDrawer() {
    visible.value = false;
  }

  function restoreValidation() {
    formRef.value?.clearValidate();
  }

  // ========== 标签输入（登录账号 / 高危命令）==========
  const accountInputVisible = ref(false);
  const accountInputValue = ref('');
  const accountInputRef = ref<{ focus: () => void } | null>(null);

  const commandInputVisible = ref(false);
  const commandInputValue = ref('');
  const commandInputRef = ref<{ focus: () => void } | null>(null);

  function showAccountInput() {
    accountInputVisible.value = true;
    nextTick(() => {
      accountInputRef.value?.focus();
    });
  }

  function handleAccountInputConfirm() {
    const value = accountInputValue.value.trim();
    if (value) {
      const accounts = model.value.loginAccounts || [];
      if (accounts.includes(value)) {
        window.$message?.warning(`账号 "${value}" 已存在`);
      } else {
        accounts.push(value);
        model.value.loginAccounts = accounts;
      }
    }
    accountInputVisible.value = false;
    accountInputValue.value = '';
  }

  function handleRemoveLoginAccount(index: number) {
    model.value.loginAccounts?.splice(index, 1);
  }

  function showCommandInput() {
    commandInputVisible.value = true;
    nextTick(() => {
      commandInputRef.value?.focus();
    });
  }

  function handleCommandInputConfirm() {
    const value = commandInputValue.value.trim();
    if (value) {
      const commands = model.value.highRiskCommands || [];
      if (commands.includes(value)) {
        window.$message?.warning(`命令 "${value}" 已存在`);
      } else {
        commands.push(value);
        model.value.highRiskCommands = commands;
      }
    }
    commandInputVisible.value = false;
    commandInputValue.value = '';
  }

  function handleRemoveHighRiskCommand(index: number) {
    model.value.highRiskCommands?.splice(index, 1);
  }

  // ========== 提交 ==========
  function buildPayload(): Bastion.AccessPolicyForm {
    return {
      name: (model.value.name || '').trim(),
      subjectType: model.value.subjectType as Bastion.AccessPolicyForm['subjectType'],
      subjectId: Number(model.value.subjectId),
      assetScopeType: model.value.assetScopeType as Bastion.AccessPolicyForm['assetScopeType'],
      assetScopeId: model.value.assetScopeType === 'all' ? 0 : Number(model.value.assetScopeId),
      loginAccounts: (model.value.loginAccounts || []).map(a => a.trim()).filter(Boolean),
      protocols: (model.value.protocols || []) as Bastion.AccessPolicyForm['protocols'],
      allowFileTransfer: Boolean(model.value.allowFileTransfer),
      allowSudo: Boolean(model.value.allowSudo),
      timeWindow: timeEnabled.value ? { ...timeWindow.value } : undefined,
      highRiskCommands: (model.value.highRiskCommands || []).map(c => c.trim()).filter(Boolean),
      status: model.value.status ?? 1
    };
  }

  async function handleSubmit() {
    const valid = await formRef.value?.validate().catch(() => false);
    if (!valid) return;

    try {
      if (props.operateType === 'edit' && props.rowData?.id) {
        await fetchUpdateAccessPolicy(props.rowData.id, buildPayload());
        window.$message?.success('更新成功');
      } else {
        await fetchCreateAccessPolicy(buildPayload());
        window.$message?.success('创建成功');
      }
      closeDrawer();
      emit('submitted');
    } catch (error: unknown) {
      window.$message?.error(error instanceof Error ? error.message : '操作失败');
    }
  }

  watch(visible, val => {
    if (val) {
      handleInitModel();
      restoreValidation();
      loadAssetData();
    }
  });
</script>

<template>
  <ElDrawer v-model="visible" :title="title" :width="700" destroy-on-close>
    <ElForm ref="formRef" :model="model" :rules="rules" label-width="120px">
      <ElFormItem label="策略名称" prop="name">
        <ElInput v-model="model.name" placeholder="请输入策略名称" maxlength="100" show-word-limit />
      </ElFormItem>

      <ElFormItem label="授权对象类型" prop="subjectType">
        <ElSelect v-model="model.subjectType" placeholder="选择授权对象类型" style="width: 100%">
          <ElOption
            v-for="option in subjectTypeOptions"
            :key="option.value"
            :value="option.value"
            :label="option.label"
          />
        </ElSelect>
      </ElFormItem>

      <!-- 授权对象为用户时 -->
      <ElFormItem v-if="model.subjectType === 'user'" label="授权对象" prop="subjectId">
        <ElSelect v-model="model.subjectId" placeholder="请选择用户" filterable clearable style="width: 100%">
          <ElOption v-for="option in userOptions" :key="option.value" :label="option.label" :value="option.value" />
        </ElSelect>
      </ElFormItem>

      <!-- 授权对象为角色时 -->
      <ElFormItem v-else-if="model.subjectType === 'role'" label="授权对象" prop="subjectId">
        <ElSelect v-model="model.subjectId" placeholder="请选择角色" style="width: 100%">
          <ElOption v-for="option in roleOptions" :key="option.value" :label="option.label" :value="option.value" />
        </ElSelect>
      </ElFormItem>

      <!-- 历史数据中的用户组类型：无管理入口，只读展示 -->
      <ElFormItem v-else-if="model.subjectType === 'user_group'" label="授权对象" prop="subjectId">
        <ElInput :value="`用户组 ID: ${model.subjectId}`" disabled />
      </ElFormItem>

      <ElFormItem label="资产范围类型" prop="assetScopeType">
        <ElSelect v-model="model.assetScopeType" placeholder="选择资产范围类型" style="width: 100%">
          <ElOption
            v-for="option in assetScopeTypeOptions"
            :key="option.value"
            :value="option.value"
            :label="option.label"
          />
        </ElSelect>
      </ElFormItem>

      <!-- 全部资产时不需要选择 -->
      <ElFormItem v-if="model.assetScopeType === 'all'" label="资产范围">
        <ElInput value="全部资产" disabled />
      </ElFormItem>

      <!-- 资产范围为单台服务器时 -->
      <ElFormItem v-else-if="model.assetScopeType === 'server'" label="资产范围" prop="assetScopeId">
        <ElSelect v-model="model.assetScopeId" placeholder="请选择服务器" filterable style="width: 100%">
          <ElOption v-for="option in serverOptions" :key="option.value" :label="option.label" :value="option.value" />
        </ElSelect>
      </ElFormItem>

      <!-- 资产范围为主机分组时 -->
      <ElFormItem v-else-if="model.assetScopeType === 'group'" label="资产范围" prop="assetScopeId">
        <ElTreeSelect
          v-model="model.assetScopeId"
          :data="serverGroups"
          :props="{ label: 'name', value: 'id', children: 'children' }"
          placeholder="请选择主机分组"
          check-strictly
          style="width: 100%"
        />
      </ElFormItem>

      <!-- 资产范围为业务系统时 -->
      <ElFormItem v-else-if="model.assetScopeType === 'business'" label="资产范围" prop="assetScopeId">
        <ElTreeSelect
          v-model="model.assetScopeId"
          :data="businessUnits"
          :props="{ label: 'name', value: 'id', children: 'children' }"
          placeholder="请选择业务系统"
          check-strictly
          style="width: 100%"
        />
      </ElFormItem>

      <!-- 资产范围为标签时 -->
      <ElFormItem v-else-if="model.assetScopeType === 'tag'" label="资产范围" prop="assetScopeId">
        <ElSelect v-model="model.assetScopeId" placeholder="请选择标签" style="width: 100%">
          <ElOption v-for="option in tagOptions" :key="option.value" :label="option.label" :value="option.value" />
        </ElSelect>
      </ElFormItem>

      <ElFormItem label="登录账号" prop="loginAccounts">
        <div class="tags-input-wrapper">
          <ElTag
            v-for="(account, idx) in model.loginAccounts"
            :key="account"
            closable
            style="margin-right: 8px; margin-bottom: 8px"
            @close="handleRemoveLoginAccount(idx)"
          >
            {{ account }}
          </ElTag>
          <ElInput
            v-if="accountInputVisible"
            ref="accountInputRef"
            v-model="accountInputValue"
            placeholder="输入账号后回车"
            size="small"
            style="width: 180px"
            @keyup.enter="handleAccountInputConfirm"
            @blur="handleAccountInputConfirm"
          />
          <ElButton v-else size="small" @click="showAccountInput">+ 添加账号</ElButton>
        </div>
      </ElFormItem>

      <ElFormItem label="允许的协议" prop="protocols">
        <ElCheckboxGroup v-model="model.protocols">
          <ElCheckbox
            v-for="option in protocolOptions"
            :key="option.value"
            :value="option.value"
            :label="option.label"
          />
        </ElCheckboxGroup>
      </ElFormItem>

      <ElFormItem label="文件传输">
        <ElSwitch v-model="model.allowFileTransfer" />
        <span class="form-item-tip">允许文件传输</span>
      </ElFormItem>

      <ElFormItem label="Sudo 权限">
        <ElSwitch v-model="model.allowSudo" />
        <span class="form-item-tip">允许 sudo</span>
      </ElFormItem>

      <ElFormItem label="高危命令">
        <div class="tags-input-wrapper">
          <ElTag
            v-for="(cmd, idx) in model.highRiskCommands"
            :key="cmd"
            closable
            type="danger"
            style="margin-right: 8px; margin-bottom: 8px"
            @close="handleRemoveHighRiskCommand(idx)"
          >
            {{ cmd }}
          </ElTag>
          <ElInput
            v-if="commandInputVisible"
            ref="commandInputRef"
            v-model="commandInputValue"
            placeholder="输入命令后回车"
            size="small"
            style="width: 180px"
            @keyup.enter="handleCommandInputConfirm"
            @blur="handleCommandInputConfirm"
          />
          <ElButton v-else size="small" @click="showCommandInput">+ 添加命令</ElButton>
        </div>
      </ElFormItem>

      <ElFormItem label="时间窗口" prop="timeWindow">
        <div class="time-window-wrapper">
          <div>
            <ElSwitch v-model="timeEnabled" />
            <span class="form-item-tip">限制访问时段</span>
          </div>
          <template v-if="timeEnabled">
            <div class="time-window-row">
              <ElTimeSelect v-model="timeWindow.start" start="00:00" end="23:30" step="00:30" placeholder="开始时间" />
              <span class="time-separator">至</span>
              <ElTimeSelect
                v-model="timeWindow.end"
                :start="timeWindow.start || '00:00'"
                end="23:30"
                step="00:30"
                placeholder="结束时间"
              />
            </div>
            <div class="time-window-row">
              <ElCheckboxGroup v-model="timeWindow.days">
                <ElCheckbox v-for="day in dayOptions" :key="day.value" :value="day.value" :label="day.label" />
              </ElCheckboxGroup>
              <span class="form-item-tip">未选择星期时默认每天生效</span>
            </div>
          </template>
        </div>
      </ElFormItem>

      <ElFormItem label="状态">
        <ElRadioGroup v-model="model.status">
          <ElRadio :value="1">启用</ElRadio>
          <ElRadio :value="0">禁用</ElRadio>
        </ElRadioGroup>
      </ElFormItem>
    </ElForm>

    <template #footer>
      <ElButton @click="closeDrawer">取消</ElButton>
      <ElButton type="primary" @click="handleSubmit">确定</ElButton>
    </template>
  </ElDrawer>
</template>

<style scoped>
  .tags-input-wrapper {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    row-gap: 8px;
    width: 100%;
  }

  .time-window-wrapper {
    display: flex;
    flex-direction: column;
    gap: 12px;
    width: 100%;
  }

  .time-window-row {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 4px;
  }

  .time-separator {
    margin: 0 4px;
    color: var(--el-text-color-secondary);
  }

  .form-item-tip {
    margin-left: 8px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }
</style>
