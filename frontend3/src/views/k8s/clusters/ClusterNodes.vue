<script setup lang="ts">
  import { onMounted, ref } from 'vue';
  import { ElAlert, ElTable, ElTableColumn } from 'element-plus';
  import { fetchK8sClusterNodes } from '@/service/api/k8s';

  interface Props {
    clusterId: number;
  }

  const props = defineProps<Props>();

  const loading = ref(false);
  const dataSource = ref<K8s.Node[]>([]);
  /** 集群内 RBAC 终判拒绝（如原生绑定 view 不含 nodes）时的内联提示，替代原始报错 */
  const forbiddenMsg = ref('');

  const loadNodes = async () => {
    loading.value = true;
    forbiddenMsg.value = '';
    try {
      const { data, error } = await fetchK8sClusterNodes(props.clusterId);
      if (!error && data) {
        dataSource.value = data || [];
      } else if (error) {
        forbiddenMsg.value = (error as Error).message || '加载节点列表失败';
      }
    } finally {
      loading.value = false;
    }
  };

  onMounted(() => {
    loadNodes();
  });
</script>

<template>
  <div class="p-4">
    <ElAlert v-if="forbiddenMsg" :title="forbiddenMsg" type="warning" :closable="false" class="mb-3" />
    <ElTable v-loading="loading" :data="dataSource" stripe>
      <ElTableColumn prop="name" label="名称" />
      <ElTableColumn prop="status" label="状态" />
      <ElTableColumn prop="roles" label="角色" />
      <ElTableColumn prop="version" label="版本" />
      <ElTableColumn prop="created" label="创建时间" />
    </ElTable>
  </div>
</template>
