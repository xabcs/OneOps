<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import { createAuthGroup, updateAuthGroup } from '@/service/api/application-permission';
  import { useForm, useFormRules } from '@/hooks/common/form';

  defineOptions({ name: 'AuthRoleOperateDrawer' });

  interface Props {
    /** operation type: add | edit */
    operateType: UI.TableOperateType;
    /** editing row data */
    rowData?: Api.ApplicationPermission.AuthGroup | null;
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
  const { defaultRequiredRule } = useFormRules();

  const title = computed(() => (props.operateType === 'add' ? '添加用户组' : '编辑用户组'));

  type Model = Pick<Api.ApplicationPermission.AuthGroup, 'name' | 'code' | 'description' | 'status'>;

  function createDefaultModel(): Model {
    return { name: '', code: '', description: '', status: 1 };
  }

  const model = ref<Model>(createDefaultModel());

  type RuleKey = Extract<keyof Model, 'name' | 'code'>;

  const rules: Record<RuleKey, App.Global.FormRule> = {
    name: defaultRequiredRule,
    code: defaultRequiredRule
  };

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
      const { error } = await updateAuthGroup(props.rowData!.id, model.value);
      if (!error) {
        ElMessage.success('更新成功');
        closeDrawer();
        emit('submitted');
      }
    } else {
      const { error } = await createAuthGroup(model.value);
      if (!error) {
        ElMessage.success('添加成功');
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
  <ElDrawer v-model="visible" :title="title" :width="500">
    <ElForm ref="formRef" :model="model" :rules="rules" label-width="100px">
      <ElFormItem label="用户组名称" prop="name">
        <ElInput v-model="model.name" placeholder="请输入用户组名称" />
      </ElFormItem>
      <ElFormItem label="用户组代码" prop="code">
        <ElInput v-model="model.code" placeholder="请输入用户组代码" :disabled="operateType === 'edit'" />
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
  </ElDrawer>
</template>
