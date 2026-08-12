<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import {
    createGroupBinding,
    fetchApplicationAuthorizationRules,
    fetchApplicationRoles,
    fetchApplications
  } from '@/service/api/application-permission';
  import { useForm, useFormRules } from '@/hooks/common/form';

  defineOptions({ name: 'AuthBindingOperateDrawer' });

  interface Props {
    /** group id (must be selected before opening drawer) */
    groupId: number | null;
  }

  const props = defineProps<Props>();

  interface Emits {
    (e: 'submitted'): void;
  }

  const emit = defineEmits<Emits>();

  // P1a: defineModel('visible')
  const visible = defineModel<boolean>('visible', { default: false });

  // P0a: drawer 内部 useForm()
  // @ts-expect-error vue-tsc noUnusedLocals: template ref
  const { formRef, validate, restoreValidation } = useForm();
  const { defaultRequiredRule } = useFormRules();

  type Model = {
    appId: number | null;
    applicationRoleId: number | null;
    authorizationRuleId: number | null;
  };

  function createDefaultModel(): Model {
    return { appId: null, applicationRoleId: null, authorizationRuleId: null };
  }

  const model = ref<Model>(createDefaultModel());

  const applications = ref<Api.ApplicationPermission.Application[]>([]);
  const applicationRoles = ref<Api.ApplicationPermission.ApplicationRole[]>([]);
  const authorizationRules = ref<Api.ApplicationPermission.AuthorizationRule[]>([]);
  const selectedApplication = ref<Api.ApplicationPermission.Application | null>(null);

  const title = computed(() => '添加权限映射');

  async function getApplications() {
    const { data, error } = await fetchApplicationOptions();
    if (!error && data) {
      applications.value = data || [];
    }
  }

  async function handleAppChange(appId: number | null) {
    applicationRoles.value = [];
    authorizationRules.value = [];

    if (!appId) {
      selectedApplication.value = null;
      return;
    }

    const app = applications.value.find(a => a.id === appId);
    selectedApplication.value = app || null;

    if (app?.type === 'jumpserver') {
      // Jumpserver 应用：加载授权规则
      const { data, error } = await fetchApplicationAuthorizationRules(appId);
      if (!error && data) {
        authorizationRules.value = data;
      }
    } else {
      // 其他应用：加载角色
      const { data, error } = await fetchApplicationRoles(appId);
      if (!error && data) {
        applicationRoles.value = data;
      }
    }
  }

  function handleInitModel() {
    model.value = createDefaultModel();
    selectedApplication.value = null;
    applicationRoles.value = [];
    authorizationRules.value = [];
  }

  function closeDrawer() {
    visible.value = false;
  }

  async function handleSubmit() {
    await validate();

    if (!props.groupId || !model.value.appId) {
      ElMessage.warning('请填写完整信息');
      return;
    }

    // 根据应用类型检查必填项
    if (selectedApplication.value?.type === 'jumpserver') {
      if (!model.value.authorizationRuleId) {
        ElMessage.warning('请选择授权规则');
        return;
      }
    } else if (!model.value.applicationRoleId) {
      ElMessage.warning('请选择角色');
      return;
    }

    // 对于 Jumpserver，使用 authorizationRuleId 作为 applicationRoleId
    const bindingData = {
      groupId: props.groupId,
      appId: model.value.appId,
      applicationRoleId:
        selectedApplication.value?.type === 'jumpserver'
          ? model.value.authorizationRuleId!
          : model.value.applicationRoleId!
    };

    const { data, error } = await createGroupBinding(bindingData);

    if (!error) {
      closeDrawer();
      emit('submitted');

      // 显示权限分配结果（drawer 内部副作用）
      if (data && data.result) {
        showAssignmentResult(data.result);
      } else {
        ElMessage.success('添加映射成功');
      }
    }
  }

  function showAssignmentResult(result: Api.ApplicationPermission.GroupBindingResult) {
    const messages: string[] = [];

    if (result.successCount > 0) {
      messages.push(`成功分配权限: ${result.successCount} 个成员`);
    }

    if (typeof result.createdIdentities === 'number' && result.createdIdentities > 0) {
      messages.push(`创建外部账号: ${result.createdIdentities} 个`);
    }

    if (result.failedCount > 0) {
      messages.push(`失败: ${result.failedCount} 个成员`);
    }

    if (result.pendingMembers && result.pendingMembers.length > 0) {
      messages.push(`待处理: ${result.pendingMembers.length} 个成员`);
    }

    const createdIdentitiesList = Array.isArray(result.createdIdentities) ? result.createdIdentities : [];

    ElMessageBox.alert(
      `
    <div style="max-height: 400px; overflow-y: auto;">
      <h4 style="margin-bottom: 10px;">权限分配详细结果</h4>
      <div style="margin-bottom: 15px;">
        ${messages.map(m => `<p style="margin: 5px 0;">${m}</p>`).join('')}
      </div>

      ${
        createdIdentitiesList.length > 0
          ? `
        <div style="margin-bottom: 15px;">
          <h5 style="margin-bottom: 5px; color: #67C23A;">✓ 已创建的外部账号</h5>
          <ul style="margin: 0; padding-left: 20px;">
            ${createdIdentitiesList
              .map(
                (item: { username: string; appName: string; status: string }) => `
              <li style="margin: 3px 0;">${item.username} - ${item.appName} - ${item.status}</li>
            `
              )
              .join('')}
          </ul>
        </div>
      `
          : ''
      }

      ${
        result.pendingMembers && result.pendingMembers.length > 0
          ? `
        <div style="margin-bottom: 15px;">
          <h5 style="margin-bottom: 5px; color: #E6A23C;">⚠ 待处理成员</h5>
          <ul style="margin: 0; padding-left: 20px;">
            ${result.pendingMembers
              .map(
                (item: { username: string; reason: string }) => `
              <li style="margin: 3px 0;">${item.username} - ${item.reason}</li>
            `
              )
              .join('')}
          </ul>
          <p style="margin-top: 10px; color: #909399; font-size: 12px;">
            建议：可以为这些成员手动创建外部账号后重新分配
          </p>
        </div>
      `
          : ''
      }

      ${
        result.failedMembers && result.failedMembers.length > 0
          ? `
        <div>
          <h5 style="margin-bottom: 5px; color: #F56C6C;">✗ 失败的成员</h5>
          <ul style="margin: 0; padding-left: 20px;">
            ${result.failedMembers
              .map(
                (item: { username: string; error: string }) => `
              <li style="margin: 3px 0;">${item.username} - ${item.error}</li>
            `
              )
              .join('')}
          </ul>
        </div>
      `
          : ''
      }
    </div>
    `,
      '权限分配结果',
      {
        dangerouslyUseHTMLString: true,
        confirmButtonText: '确定'
      }
    );
  }

  watch(visible, val => {
    if (val) {
      handleInitModel();
      restoreValidation();
      if (applications.value.length === 0) {
        getApplications();
      }
    }
  });
</script>

<template>
  <ElDrawer v-model="visible" :title="title" :width="400">
    <ElForm ref="formRef" :model="model" label-width="80px">
      <ElFormItem label="应用" prop="appId" :rules="[defaultRequiredRule]">
        <ElSelect v-model="model.appId" placeholder="请选择应用" class="w-full" @change="handleAppChange">
          <ElOption v-for="app in applications" :key="app.id" :label="`${app.name} (${app.type})`" :value="app.id" />
        </ElSelect>
      </ElFormItem>

      <!-- Jumpserver 应用：选择授权规则 -->
      <ElFormItem v-if="selectedApplication?.type === 'jumpserver'" label="授权规则" prop="authorizationRuleId">
        <ElSelect v-model="model.authorizationRuleId" placeholder="请选择授权规则" class="w-full">
          <ElOption
            v-for="rule in authorizationRules"
            :key="rule.id"
            :label="`${rule.ruleName} - ${rule.subjectType}`"
            :value="rule.id"
          />
        </ElSelect>
      </ElFormItem>

      <!-- 其他应用：选择角色 -->
      <ElFormItem v-else label="外部角色" prop="applicationRoleId">
        <ElSelect v-model="model.applicationRoleId" placeholder="请选择外部角色" class="w-full">
          <ElOption v-for="role in applicationRoles" :key="role.id" :label="role.roleName" :value="role.id" />
        </ElSelect>
      </ElFormItem>
    </ElForm>

    <template #footer>
      <ElButton @click="closeDrawer">取消</ElButton>
      <ElButton type="primary" @click="handleSubmit">确定</ElButton>
    </template>
  </ElDrawer>
</template>
