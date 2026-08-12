<script setup lang="ts">
  import { ref, watch } from 'vue';
  import { ElNotification } from 'element-plus';
  import { createK8sCluster, updateK8sCluster } from '@/service/api/k8s';

  defineOptions({ name: 'ClusterFormDialog' });

  const props = defineProps<{
    mode: 'create' | 'edit';
    formData?: K8s.Cluster | null;
  }>();

  const emit = defineEmits<{
    (e: 'submitted'): void;
  }>();

  const visible = defineModel<boolean>('visible', { default: false });

  interface ClusterFormState {
    id: number | null;
    name: string;
    description: string;
    endpoint: string;
    kubeconfig: string;
    clusterType: string;
    region: string;
    nodeCount: number;
  }

  function getDefaultForm(): ClusterFormState {
    return {
      id: null,
      name: '',
      description: '',
      endpoint: '',
      kubeconfig: '',
      clusterType: 'standard',
      region: '',
      nodeCount: 0
    };
  }

  const form = ref<ClusterFormState>(getDefaultForm());
  const submitting = ref(false);

  // 同步 props.formData 到内部表单
  watch(visible, val => {
    if (val) {
      if (props.mode === 'edit' && props.formData) {
        form.value = {
          id: props.formData.id,
          name: props.formData.name,
          description: props.formData.description,
          endpoint: props.formData.endpoint,
          kubeconfig: '',
          clusterType: props.formData.clusterType,
          region: props.formData.region,
          nodeCount: props.formData.nodeCount
        };
      } else {
        form.value = getDefaultForm();
      }
    }
  });

  async function handleSubmit() {
    submitting.value = true;
    try {
      if (props.mode === 'edit' && form.value.id) {
        await updateK8sCluster(form.value.id, form.value);
        ElNotification({
          title: '操作成功',
          message: '更新集群成功',
          type: 'success',
          duration: 3000
        });
      } else {
        await createK8sCluster(form.value);
        ElNotification({
          title: '操作成功',
          message: '创建集群成功',
          type: 'success',
          duration: 3000
        });
      }
      visible.value = false;
      emit('submitted');
    } catch (error: unknown) {
      const err = error as Error;
      ElNotification({
        title: '操作失败',
        message: err.message || '操作失败',
        type: 'error',
        duration: 3000
      });
    } finally {
      submitting.value = false;
    }
  }
</script>

<template>
  <ElDialog
    v-model="visible"
    :title="mode === 'edit' ? '编辑集群' : '添加集群'"
    width="600px"
    :close-on-click-modal="false"
  >
    <ElForm :model="form" label-width="100px">
      <ElFormItem label="集群名称" required>
        <ElInput v-model="form.name" placeholder="请输入集群名称" />
      </ElFormItem>
      <ElFormItem label="集群描述">
        <ElInput v-model="form.description" type="textarea" placeholder="请输入集群描述" :rows="3" />
      </ElFormItem>
      <ElFormItem label="API 地址" required>
        <ElInput v-model="form.endpoint" placeholder="https://k8s-api.example.com:6443" />
      </ElFormItem>
      <ElFormItem label="Kubeconfig" required>
        <ElInput v-model="form.kubeconfig" type="textarea" placeholder="粘贴 kubeconfig 内容" :rows="10" />
      </ElFormItem>
      <ElFormItem label="集群类型" required>
        <ElSelect v-model="form.clusterType" placeholder="选择集群类型">
          <ElOption label="标准集群" value="standard" />
          <ElOption label="托管集群" value="managed" />
          <ElOption label="边缘集群" value="edge" />
        </ElSelect>
      </ElFormItem>
      <ElFormItem label="区域">
        <ElInput v-model="form.region" placeholder="如：us-west-2" />
      </ElFormItem>
      <ElFormItem label="节点数">
        <ElInputNumber v-model="form.nodeCount" :min="0" placeholder="自动获取" />
      </ElFormItem>
    </ElForm>

    <template #footer>
      <ElButton @click="visible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitting" @click="handleSubmit">确定</ElButton>
    </template>
  </ElDialog>
</template>
