<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import {
    fetchCreateAccessPolicy,
    fetchUpdateAccessPolicy
  } from '@/service/api/cmdb';
  import {
    fetchGetAllRoles,
    fetchGetBusinessUnits,
    fetchGetServerGroups,
    fetchGetServerTags,
    fetchUserOptions
  } from '@/service/api';
  import { useForm, useFormRules } from '@/hooks/common/form';

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

  // @ts-expect-error vue-tsc noUnusedLocals: template ref
  const { formRef, validate, restoreValidation } = useForm();
  const { defaultRequiredRule } = useFormRules();

  const title = computed(() => (props.operateType === 'add' ? '新增策略' : '编辑策略'));

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

  const userOptions = computed(() => {
    return users.value.map(u => ({ label: `${u.username} (${u.email || '无邮箱'})`, value: u.id }));
  });

  const roleOptions = computed(() => {
    return roles.value.map(r => ({ label: r.name, value: r.id }));
  });

  const groupOptions = computed(() => serverGroups.value);
  const businessOptions = computed(() => businessUnits.value);
  const tagOptions = computed(() => serverTags.value.map(t => ({ label: t.name, value: t.id })));

  type PolicyModel = Partial<Bastion.AccessPolicyForm>;

  function createDefaultModel(): PolicyModel {
    return {
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
  }

  const model = ref<PolicyModel>(createDefaultModel());

  const rules = {
    name: defaultRequiredRule,
    subjectType: defaultRequiredRule,
    subjectId: defaultRequiredRule,
    assetScopeType: defaultRequiredRule,
    assetScopeId: defaultRequiredRule
  };

  function handleInitModel() {
    model.value = createDefaultModel();
    if (props.operateType === 'edit' && props.rowData) {
      model.value = { ...props.rowData };
    }
  }

  async function loadAssetData() {
    const [usersRes, rolesRes, groupsRes, businessRes, tagsRes] = await Promise.allSettled([
      fetchUserOptions(),
      fetchGetAllRoles(),
      fetchGetServerGroups(),
      fetchGetBusinessUnits(),
      fetchGetServerTags()
    ]);

    if (usersRes.status === 'fulfilled') users.value = usersRes.value.data || [];
    if (rolesRes.status === 'fulfilled') roles.value = rolesRes.value.data || [];
    if (groupsRes.status === 'fulfilled') serverGroups.value = groupsRes.value.data || [];
    if (businessRes.status === 'fulfilled') businessUnits.value = businessRes.value.data || [];
    if (tagsRes.status === 'fulfilled') serverTags.value = tagsRes.value.data || [];
  }

  function closeDrawer() {
    visible.value = false;
  }

  function handleRemoveLoginAccount(index: number) {
    model.value.loginAccounts?.splice(index, 1);
  }

  function handleAddHighRiskCommand() {
    if (!model.value.highRiskCommands) model.value.highRiskCommands = [];
    model.value.highRiskCommands.push('');
  }

  function handleRemoveHighRiskCommand(index: number) {
    model.value.highRiskCommands?.splice(index, 1);
  }

  async function handleSubmit() {
    await validate();

    try {
      if (props.operateType === 'edit' && model.value.id) {
        await fetchUpdateAccessPolicy(model.value.id, model.value);
        window.$message?.success('更新成功');
      } else {
        await fetchCreateAccessPolicy(model.value as Bastion.AccessPolicyForm);
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
  <ElDrawer v-model="visible" :title="title" :width="700">
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
        <ElSelect v-model="model.subjectId" placeholder="请选择用户" filterable style="width: 100%">
          <ElOption v-for="option in userOptions" :key="option.value" :label="option.label" :value="option.value" />
        </ElSelect>
      </ElFormItem>

      <!-- 授权对象为角色时 -->
      <ElFormItem v-if="model.subjectType === 'role'" label="授权对象" prop="subjectId">
        <ElSelect v-model="model.subjectId" placeholder="请选择角色" style="width: 100%">
          <ElOption v-for="option in roleOptions" :key="option.value" :label="option.label" :value="option.value" />
        </ElSelect>
      </ElFormItem>

      <!-- 授权对象为用户组时 -->
      <ElFormItem v-if="model.subjectType === 'user_group'" label="授权对象" prop="subjectId">
        <ElInput placeholder="用户组功能暂未开放" disabled />
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

      <!-- 资产范围为主机分组时 -->
      <ElFormItem v-if="model.assetScopeType === 'group'" label="资产范围" prop="assetScopeId">
        <ElTreeSelect
          v-model="model.assetScopeId"
          :data="groupOptions"
          :props="{ label: 'name', value: 'id', children: 'children' }"
          placeholder="请选择主机分组"
          check-strictly
          style="width: 100%"
        />
      </ElFormItem>

      <!-- 资产范围为业务系统时 -->
      <ElFormItem v-if="model.assetScopeType === 'business'" label="资产范围" prop="assetScopeId">
        <ElTreeSelect
          v-model="model.assetScopeId"
          :data="businessOptions"
          :props="{ label: 'name', value: 'id', children: 'children' }"
          placeholder="请选择业务系统"
          check-strictly
          style="width: 100%"
        />
      </ElFormItem>

      <!-- 资产范围为标签时 -->
      <ElFormItem v-if="model.assetScopeType === 'tag'" label="资产范围" prop="assetScopeId">
        <ElSelect v-model="model.assetScopeId" placeholder="请选择标签" style="width: 100%">
          <ElOption v-for="option in tagOptions" :key="option.value" :label="option.label" :value="option.value" />
        </ElSelect>
      </ElFormItem>

      <!-- 资产范围为单台服务器时 -->
      <ElFormItem v-if="model.assetScopeType === 'server'" label="资产范围" prop="assetScopeId">
        <ElInput placeholder="单台服务器选择功能即将开放" disabled />
      </ElFormItem>

      <ElFormItem label="允许的登录账号">
        <div class="tags-input-wrapper">
          <ElTag
            v-for="(account, idx) in model.loginAccounts"
            :key="idx"
            closable
            style="margin-right: 8px; margin-bottom: 8px"
            @close="handleRemoveLoginAccount(idx)"
          >
            {{ account }}
          </ElTag>
          <ElInput
            v-if="!model.loginAccounts || model.loginAccounts.length === 0"
            placeholder="输入账号后按回车"
            size="small"
            style="width: 150px"
            @change="
              (val: string) => {
                if (val) {
                  model.loginAccounts = [val];
                }
              }
            "
          />
        </div>
      </ElFormItem>

      <ElFormItem label="允许的协议">
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
        <span style="margin-left: 8px">允许文件传输</span>
      </ElFormItem>

      <ElFormItem label="Sudo 权限">
        <ElSwitch v-model="model.allowSudo" />
        <span style="margin-left: 8px">允许 sudo</span>
      </ElFormItem>

      <ElFormItem label="需要审批">
        <ElSwitch v-model="model.requireApproval" />
        <span style="margin-left: 8px">连接前需要审批</span>
      </ElFormItem>

      <ElFormItem label="高危命令">
        <div class="tags-input-wrapper">
          <ElTag
            v-for="(cmd, idx) in model.highRiskCommands"
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
  }
</style>
