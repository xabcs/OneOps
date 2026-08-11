<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { ElMessage, ElTable, ElTableColumn } from 'element-plus';
import { fetchK8sClusterNodes } from '@/service/api/k8s';

interface Props {
  clusterId: number;
}

const props = defineProps<Props>();

const loading = ref(false);
const dataSource = ref<K8s.Node[]>([]);

const loadNodes = async () => {
  loading.value = true;
  try {
    const res = await fetchK8sClusterNodes(props.clusterId);
    if (res.code === 200) {
      dataSource.value = res.data || [];
    }
  } catch (error: unknown) {
    const err = error as Error;
    ElMessage.error(err.message || '加载节点列表失败');
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
    <ElTable v-loading="loading" :data="dataSource" stripe>
      <ElTableColumn prop="name" label="名称" />
      <ElTableColumn prop="status" label="状态" />
      <ElTableColumn prop="roles" label="角色" />
      <ElTableColumn prop="version" label="版本" />
      <ElTableColumn prop="created" label="创建时间" />
    </ElTable>
  </div>
</template>
