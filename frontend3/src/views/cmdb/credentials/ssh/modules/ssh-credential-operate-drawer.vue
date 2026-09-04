<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import { fetchCreateSSHCredential, fetchUpdateSSHCredential } from '@/service/api/cmdb';
  import { useForm, useFormRules } from '@/hooks/common/form';

  defineOptions({ name: 'CmdbSshCredentialOperateDrawer' });

  interface Props {
    operateType: UI.TableOperateType;
    rowData?: CMDB.SSHCredential | null;
    defaultCredentialType?: CMDB.CredentialType;
  }

  const props = withDefaults(defineProps<Props>(), {
    defaultCredentialType: 'user'
  });

  interface Emits {
    (e: 'submitted'): void;
  }

  const emit = defineEmits<Emits>();

  // P1a
  const visible = defineModel<boolean>('visible', { default: false });

  // P0a
  // @ts-expect-error vue-tsc noUnusedLocals: template ref
  const { formRef, validate, restoreValidation } = useForm();
  const { defaultRequiredRule } = useFormRules();

  const title = computed(() => (props.operateType === 'add' ? '新增SSH凭证' : '编辑SSH凭证'));

  type Model = Omit<CMDB.SSHCredentialForm, 'credentialType'> & { credentialType?: CMDB.CredentialType };

  function createDefaultModel(): Model {
    return {
      name: '',
      description: '',
      username: 'root',
      authType: 'password',
      password: '',
      privateKey: '',
      passphrase: '',
      credentialType: props.defaultCredentialType,
      status: 1
    };
  }

  const model = ref<Model>(createDefaultModel());

  const rules = {
    name: defaultRequiredRule,
    username: defaultRequiredRule
  };

  function handleInitModel() {
    model.value = createDefaultModel();
    if (props.operateType === 'edit' && props.rowData) {
      model.value = {
        id: props.rowData.id,
        name: props.rowData.name,
        description: props.rowData.description || '',
        username: props.rowData.username,
        authType: props.rowData.authType,
        credentialType: props.rowData.credentialType || props.defaultCredentialType,
        status: 1
      };
    }
  }

  function closeDrawer() {
    visible.value = false;
  }

  async function handleSubmit() {
    await validate();

    const formData = { ...model.value, credentialType: model.value.credentialType || props.defaultCredentialType };

    if (model.value.id) {
      const { error } = await fetchUpdateSSHCredential(model.value.id, formData);
      if (!error) {
        ElMessage.success('保存成功');
        closeDrawer();
        emit('submitted');
      }
    } else {
      const { error } = await fetchCreateSSHCredential(formData);
      if (!error) {
        ElMessage.success('创建成功');
        closeDrawer();
        emit('submitted');
      }
    }
  }

  watch(visible, val => {
    if (val) {
      handleInitModel();
      restoreValidation();
    }
  });
</script>

<template>
  <ElDrawer v-model="visible" :title="title" :width="600">
    <ElForm ref="formRef" :model="model" :rules="rules" label-width="120px">
      <ElFormItem label="凭证用途" prop="credentialType">
        <ElRadioGroup v-model="model.credentialType">
          <ElRadio value="user">
            <span class="font-medium">用户连接</span>
            <span class="ml-6px text-12px text-gray-400">用于用户堡垒 SSH，受访问策略约束</span>
          </ElRadio>
        </ElRadioGroup>
        <ElRadioGroup v-model="model.credentialType" class="mt-8px">
          <ElRadio value="system">
            <span class="font-medium">系统运维</span>
            <span class="ml-6px text-12px text-gray-400">供 OneOps 后端 Agent 部署、采集使用，需 root / sudo 权限</span>
          </ElRadio>
        </ElRadioGroup>
      </ElFormItem>
      <ElFormItem label="凭证名称" prop="name">
        <ElInput v-model="model.name" placeholder="请输入凭证名称" />
      </ElFormItem>
      <ElFormItem label="用户名" prop="username">
        <ElInput v-model="model.username" placeholder="请输入SSH用户名" />
      </ElFormItem>
      <ElFormItem label="认证类型" prop="authType">
        <ElRadioGroup v-model="model.authType">
          <ElRadio value="password">密码认证</ElRadio>
          <ElRadio value="key">密钥认证</ElRadio>
        </ElRadioGroup>
      </ElFormItem>
      <template v-if="model.authType === 'password'">
        <ElFormItem label="密码" prop="password">
          <ElInput
            v-model="model.password"
            type="password"
            :placeholder="isEdit ? '不修改请留空' : '请输入密码'"
            show-password
          />
        </ElFormItem>
      </template>
      <template v-if="model.authType === 'key'">
        <ElFormItem label="私钥" prop="privateKey">
          <ElInput
            v-model="model.privateKey"
            type="textarea"
            :rows="6"
            :placeholder="isEdit ? '不修改请留空' : '请输入私钥内容'"
          />
        </ElFormItem>
        <ElFormItem label="私钥密码">
          <ElInput
            v-model="model.passphrase"
            type="password"
            :placeholder="isEdit ? '不修改请留空' : '如果私钥有密码请输入'"
            show-password
          />
        </ElFormItem>
      </template>
      <ElFormItem label="描述">
        <ElInput v-model="model.description" type="textarea" :rows="3" placeholder="请输入描述" />
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
      <ElButton type="primary" @click="handleSubmit">保存</ElButton>
    </template>
  </ElDrawer>
</template>
