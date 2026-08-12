<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import { fetchCreateServerRoom, fetchUpdateServerRoom } from '@/service/api';
  import { useForm, useFormRules } from '@/hooks/common/form';

  defineOptions({ name: 'CmdbRoomOperateDrawer' });

  interface Props {
    operateType: UI.TableOperateType;
    rowData?: CMDB.ServerRoom | null;
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

  const title = computed(() => (props.operateType === 'add' ? '新增机房' : '编辑机房'));

  type Model = {
    id?: number;
    name: string;
    code: string;
    location: string;
    address: string;
    provider: string;
    contact: string;
    phone: string;
    status: number;
    remarks: string;
  };

  function createDefaultModel(): Model {
    return {
      name: '',
      code: '',
      location: '',
      address: '',
      provider: '',
      contact: '',
      phone: '',
      status: 1,
      remarks: ''
    };
  }

  const model = ref<Model>(createDefaultModel());

  const rules = {
    name: defaultRequiredRule,
    code: defaultRequiredRule
  };

  const providerOptions = [
    { label: '阿里云', value: 'aliyun' },
    { label: '腾讯云', value: 'tencent' },
    { label: 'AWS', value: 'aws' },
    { label: '华为云', value: 'huawei' },
    { label: '自建', value: 'self' },
    { label: '其他', value: 'other' }
  ];

  function handleInitModel() {
    model.value = createDefaultModel();
    if (props.operateType === 'edit' && props.rowData) {
      model.value = {
        id: props.rowData.id,
        name: props.rowData.name,
        code: props.rowData.code,
        location: props.rowData.location || '',
        address: props.rowData.address || '',
        provider: props.rowData.provider || '',
        contact: props.rowData.contact || '',
        phone: props.rowData.phone || '',
        status: props.rowData.status,
        remarks: props.rowData.remarks || ''
      };
    }
  }

  function closeDrawer() {
    visible.value = false;
  }

  async function handleSubmit() {
    await validate();

    if (model.value.id) {
      const { error } = await fetchUpdateServerRoom(model.value.id, model.value);
      if (!error) {
        ElMessage.success('机房更新成功');
        closeDrawer();
        emit('submitted');
      }
    } else {
      const { error } = await fetchCreateServerRoom(model.value);
      if (!error) {
        ElMessage.success('机房创建成功');
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
    <ElForm ref="formRef" :model="model" :rules="rules" label-width="100px">
      <ElFormItem label="机房名称" prop="name">
        <ElInput v-model="model.name" placeholder="请输入机房名称" />
      </ElFormItem>
      <ElFormItem label="机房代码" prop="code">
        <ElInput v-model="model.code" placeholder="请输入机房代码（如：ALIYUN-HUADONG）" />
      </ElFormItem>
      <ElFormItem label="位置">
        <ElInput v-model="model.location" placeholder="请输入位置（如：杭州）" />
      </ElFormItem>
      <ElFormItem label="详细地址">
        <ElInput v-model="model.address" placeholder="请输入详细地址" />
      </ElFormItem>
      <ElFormItem label="服务商">
        <ElSelect v-model="model.provider" placeholder="请选择服务商" class="w-full">
          <ElOption v-for="option in providerOptions" :key="option.value" :label="option.label" :value="option.value" />
        </ElSelect>
      </ElFormItem>
      <ElFormItem label="联系人">
        <ElInput v-model="model.contact" placeholder="请输入联系人" />
      </ElFormItem>
      <ElFormItem label="联系电话">
        <ElInput v-model="model.phone" placeholder="请输入联系电话" />
      </ElFormItem>
      <ElFormItem label="状态">
        <ElRadioGroup v-model="model.status">
          <ElRadio :value="1">启用</ElRadio>
          <ElRadio :value="0">禁用</ElRadio>
        </ElRadioGroup>
      </ElFormItem>
      <ElFormItem label="备注">
        <ElInput v-model="model.remarks" type="textarea" :rows="3" placeholder="请输入备注信息" />
      </ElFormItem>
    </ElForm>

    <template #footer>
      <ElButton @click="closeDrawer">取消</ElButton>
      <ElButton type="primary" @click="handleSubmit">保存</ElButton>
    </template>
  </ElDrawer>
</template>
