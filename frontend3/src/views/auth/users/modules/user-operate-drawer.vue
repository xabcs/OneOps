<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import { createAuthUser, updateAuthUser } from '@/service/api/application-permission';
  import type { CreatedAuthUser } from '@/service/api/application-permission';
  import { useForm, useFormRules } from '@/hooks/common/form';

  defineOptions({ name: 'AuthUserOperateDrawer' });

  interface Props {
    /** operation type: add | edit */
    operateType: UI.TableOperateType;
    /** editing row data */
    rowData?: Api.ApplicationPermission.AuthUser | null;
  }

  const props = defineProps<Props>();

  interface Emits {
    (e: 'submitted'): void;
  }

  const emit = defineEmits<Emits>();

  // P1a: defineModel('visible') 替代 :visible prop + @update:visible emit
  const visible = defineModel<boolean>('visible', { default: false });

  // P0a: drawer 内部 useForm()，表单 ref 不再上抛
  // @ts-expect-error vue-tsc noUnusedLocals: template ref
  const { formRef, validate, restoreValidation } = useForm();
  const { defaultRequiredRule, patternRules } = useFormRules();

  const title = computed(() => (props.operateType === 'add' ? '添加用户' : '编辑用户'));

  type Model = Pick<
    Api.ApplicationPermission.AuthUser,
    'username' | 'nickname' | 'email' | 'phone' | 'description' | 'status'
  >;

  function createDefaultModel(): Model {
    return {
      username: '',
      nickname: '',
      email: '',
      phone: '',
      description: '',
      status: 1
    };
  }

  const model = ref<Model>(createDefaultModel());

  type RuleKey = Extract<keyof Model, 'username' | 'email'>;

  const rules: Record<RuleKey, App.Global.FormRule> = {
    username: defaultRequiredRule,
    email: patternRules.email
  };

  // 创建成功后展示初始密码（drawer 内部副作用，与父组件解耦）
  const passwordDialogVisible = ref(false);
  const createdPassword = ref<CreatedAuthUser>({ username: '', password: '' });

  function handleInitModel() {
    model.value = createDefaultModel();
    if (props.operateType === 'edit' && props.rowData) {
      Object.assign(model.value, props.rowData);
    }
  }

  function closeDrawer() {
    visible.value = false;
  }

  async function handleSubmit() {
    await validate();

    if (props.operateType === 'edit') {
      const { error } = await updateAuthUser(props.rowData!.id, model.value);
      if (!error) {
        ElMessage.success('更新成功');
        closeDrawer();
        emit('submitted');
      }
    } else {
      const { data, error } = await createAuthUser(model.value);
      if (!error && data) {
        closeDrawer();
        emit('submitted');
        // 创建成功后展示初始密码
        if (data.password) {
          createdPassword.value = {
            username: data.username,
            password: data.password
          };
          passwordDialogVisible.value = true;
        }
      }
    }
  }

  watch(visible, val => {
    if (val) {
      handleInitModel();
      restoreValidation();
    }
  });

  function copyPassword() {
    navigator.clipboard.writeText(createdPassword.value.password);
    ElMessage.success('密码已复制到剪贴板');
  }
</script>

<template>
  <ElDrawer v-model="visible" :title="title" :width="500">
    <ElForm ref="formRef" :model="model" :rules="rules" label-width="100px">
      <ElFormItem label="用户名" prop="username">
        <ElInput v-model="model.username" placeholder="请输入用户名" :disabled="operateType === 'edit'" />
      </ElFormItem>
      <ElFormItem label="昵称" prop="nickname">
        <ElInput v-model="model.nickname" placeholder="请输入昵称" />
      </ElFormItem>
      <ElFormItem label="邮箱" prop="email">
        <ElInput v-model="model.email" placeholder="请输入邮箱" />
      </ElFormItem>
      <ElFormItem label="电话" prop="phone">
        <ElInput v-model="model.phone" placeholder="请输入电话" />
      </ElFormItem>
      <ElFormItem label="描述" prop="description">
        <ElInput v-model="model.description" type="textarea" :rows="3" placeholder="请输入描述" />
      </ElFormItem>
      <ElFormItem label="状态" prop="status">
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

    <!-- 创建成功后初始密码展示 -->
    <ElDialog v-model="passwordDialogVisible" title="用户创建成功" width="500px" append-to-body>
      <ElAlert type="warning" :closable="false" class="mb-4">
        <template #title>
          <strong>重要提示：</strong>
          以下初始密码仅显示一次，请妥善保管或立即通知用户修改密码
        </template>
      </ElAlert>

      <ElDescriptions :column="1" border>
        <ElDescriptionsItem label="用户名">{{ createdPassword.username }}</ElDescriptionsItem>
        <ElDescriptionsItem label="初始密码">
          <div class="flex items-center gap-2">
            <code class="text-lg text-primary font-bold">{{ createdPassword.password }}</code>
            <ElButton size="small" type="primary" @click="copyPassword">复制密码</ElButton>
          </div>
        </ElDescriptionsItem>
      </ElDescriptions>

      <div class="mt-4 text-sm text-gray-500">
        <p>• 该密码将用于所有外部系统的初始登录</p>
        <p>• 请通知用户在首次登录后立即修改密码</p>
        <p>• 此密码仅显示一次，关闭后将无法再次查看</p>
      </div>

      <template #footer>
        <ElButton type="primary" @click="passwordDialogVisible = false">我已保存，关闭</ElButton>
      </template>
    </ElDialog>
  </ElDrawer>
</template>
