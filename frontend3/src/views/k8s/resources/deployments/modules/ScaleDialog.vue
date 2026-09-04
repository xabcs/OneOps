<script setup lang="ts">
  import { ref, watch } from 'vue';
  import type { FormInstance, FormRules } from 'element-plus';
  import { scaleK8sDeployment } from '@/service/api/k8s';

  defineOptions({ name: 'ScaleDialog' });

  interface Props {
    clusterId: number | null;
    namespace: string;
    deployment: K8s.Deployment | null;
  }

  const props = defineProps<Props>();

  interface Emits {
    (e: 'submitted'): void;
  }

  const emit = defineEmits<Emits>();

  const visible = defineModel<boolean>('visible', { default: false });

  const formRef = ref<FormInstance | null>(null);
  const submitting = ref(false);
  const replicas = ref(1);

  const rules: FormRules = {
    replicas: [
      { required: true, message: '请输入副本数', trigger: 'blur' },
      {
        validator: (_rule, value, callback) => {
          if (value < 0 || value > 100) {
            callback(new Error('副本数必须在 0-100 之间'));
          } else {
            callback();
          }
        },
        trigger: 'blur'
      }
    ]
  };

  watch(visible, val => {
    if (val && props.deployment) {
      replicas.value = props.deployment.replicas || 1;
      formRef.value?.clearValidate();
    }
  });

  async function handleSubmit() {
    if (!formRef.value) return;
    await formRef.value.validate();

    submitting.value = true;
    try {
      await scaleK8sDeployment(props.clusterId!, {
        namespace: props.namespace,
        name: props.deployment!.name,
        replicas: replicas.value
      });
      ElMessage.success('缩放成功');
      visible.value = false;
      emit('submitted');
    } catch (error: unknown) {
      const err = error as Error;
      ElMessage.error(err.message || '缩放失败');
    } finally {
      submitting.value = false;
    }
  }
</script>

<template>
  <ElDialog v-model="visible" title="缩放 Deployment" width="500px">
    <ElForm ref="formRef" :model="{ replicas }" :rules="rules" label-width="100px">
      <ElFormItem label="Deployment">
        <ElInput :value="deployment?.name" disabled />
      </ElFormItem>
      <ElFormItem label="副本数" prop="replicas">
        <ElInputNumber v-model="replicas" :min="0" :max="100" :step="1" />
      </ElFormItem>
    </ElForm>
    <template #footer>
      <ElButton @click="visible = false">取消</ElButton>
      <PermissionButton code="k8s.resource.update" type="primary" :loading="submitting" @click="handleSubmit">
        确定
      </PermissionButton>
    </template>
  </ElDialog>
</template>
