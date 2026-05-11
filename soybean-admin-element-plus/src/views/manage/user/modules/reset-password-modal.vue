<script setup lang="ts">
import { ref, watch } from 'vue';
import { useForm, useFormRules } from '@/hooks/common/form';
import { fetchResetUserPassword } from '@/service/api';
import { $t } from '@/locales';

defineOptions({ name: 'ResetPasswordModal' });

interface Props {
  userId: number;
  username?: string;
}

const props = defineProps<Props>();

interface Emits {
  (e: 'submitted'): void;
}

const emit = defineEmits<Emits>();

const visible = defineModel<boolean>('visible', {
  default: false
});

const { formRef, validate, restoreValidation } = useForm();
const { defaultRequiredRule } = useFormRules();

const title = '重置密码';

type Model = {
  password: string;
  confirmPassword: string;
};

const model = ref(createDefaultModel());

function createDefaultModel(): Model {
  return {
    password: '',
    confirmPassword: ''
  };
}

type RuleKey = keyof Model;

const rules: Record<RuleKey, App.Global.FormRule> = {
  password: defaultRequiredRule,
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    {
      validator: (rule: any, value: string, callback: any) => {
        if (value !== model.value.password) {
          callback(new Error('两次输入的密码不一致'));
        } else {
          callback();
        }
      },
      trigger: 'blur'
    }
  ]
};

function closeModal() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();

  const { error } = await fetchResetUserPassword(props.userId, model.value.password);

  if (!error) {
    window.$message?.success('密码重置成功');
    closeModal();
    emit('submitted');
  } else {
    window.$message?.error('密码重置失败');
  }
}

watch(visible, () => {
  if (visible.value) {
    model.value = createDefaultModel();
    restoreValidation();
  }
});
</script>

<template>
  <ElDialog v-model="visible" :title="title" :size="400">
    <ElForm ref="formRef" :model="model" :rules="rules" label-position="top">
      <ElFormItem label="用户名">
        <ElInput :value="username" disabled />
      </ElFormItem>
      <ElFormItem label="新密码" prop="password">
        <ElInput
          v-model="model.password"
          type="password"
          placeholder="请输入新密码"
          show-password
        />
      </ElFormItem>
      <ElFormItem label="确认密码" prop="confirmPassword">
        <ElInput
          v-model="model.confirmPassword"
          type="password"
          placeholder="请再次输入新密码"
          show-password
        />
      </ElFormItem>
    </ElForm>
    <template #footer>
      <ElSpace :size="16">
        <ElButton @click="closeModal">{{ $t('common.cancel') }}</ElButton>
        <ElButton type="primary" @click="handleSubmit">{{ $t('common.confirm') }}</ElButton>
      </ElSpace>
    </template>
  </ElDialog>
</template>

<style scoped></style>
