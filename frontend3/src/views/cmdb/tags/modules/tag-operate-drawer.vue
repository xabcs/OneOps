<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import { fetchCreateServerTag, fetchUpdateServerTag } from '@/service/api';
  import { useForm, useFormRules } from '@/hooks/common/form';

  defineOptions({ name: 'CmdbTagOperateDrawer' });

  interface Props {
    operateType: UI.TableOperateType;
    rowData?: CMDB.ServerTag | null;
  }

  const props = defineProps<Props>();

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

  const title = computed(() => (props.operateType === 'add' ? '新增标签' : '编辑标签'));

  type Model = CMDB.ServerTagForm;

  function createDefaultModel(): Model {
    return { id: undefined, name: '', color: '#409EFF', description: '', sortOrder: 0, status: 1 };
  }

  const model = ref<Model>(createDefaultModel());

  const rules = {
    name: defaultRequiredRule,
    color: defaultRequiredRule
  };

  function handleInitModel() {
    model.value = createDefaultModel();
    if (props.operateType === 'edit' && props.rowData) {
      model.value = {
        id: props.rowData.id,
        name: props.rowData.name,
        color: props.rowData.color,
        description: props.rowData.description || '',
        sortOrder: props.rowData.sortOrder,
        status: props.rowData.status
      };
    }
  }

  function closeDrawer() {
    visible.value = false;
  }

  async function handleSubmit() {
    await validate();

    if (model.value.id) {
      const { error } = await fetchUpdateServerTag(model.value.id, model.value);
      if (!error) {
        ElMessage.success('标签更新成功');
        closeDrawer();
        emit('submitted');
      }
    } else {
      const { error } = await fetchCreateServerTag(model.value);
      if (!error) {
        ElMessage.success('标签创建成功');
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
      <ElFormItem label="标签名称" prop="name">
        <ElInput v-model="model.name" placeholder="请输入标签名称" />
      </ElFormItem>
      <ElFormItem label="颜色" prop="color">
        <ElColorPicker v-model="model.color" />
      </ElFormItem>
      <ElFormItem label="排序">
        <ElInputNumber v-model="model.sortOrder" :min="0" class="w-full" />
      </ElFormItem>
      <ElFormItem label="状态">
        <ElRadioGroup v-model="model.status">
          <ElRadio :value="1">启用</ElRadio>
          <ElRadio :value="0">禁用</ElRadio>
        </ElRadioGroup>
      </ElFormItem>
      <ElFormItem label="描述">
        <ElInput v-model="model.description" type="textarea" :rows="3" placeholder="请输入标签描述" />
      </ElFormItem>
    </ElForm>

    <template #footer>
      <ElButton @click="closeDrawer">取消</ElButton>
      <ElButton type="primary" @click="handleSubmit">保存</ElButton>
    </template>
  </ElDrawer>
</template>
