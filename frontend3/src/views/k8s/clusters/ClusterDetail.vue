<script setup lang="ts">
  import { onMounted, ref } from 'vue';
  import { ElDescriptions, ElDescriptionsItem, ElMessage, ElTabPane, ElTabs } from 'element-plus';
  import { getK8sClusterDetail } from '@/service/api/k8s';
  import ClusterNodes from './ClusterNodes.vue';
  import ClusterNamespaces from './ClusterNamespaces.vue';
  import ClusterUsers from './ClusterUsers.vue';

  interface Props {
    clusterId: number;
  }

  const props = defineProps<Props>();

  const loading = ref(false);
  const cluster = ref<K8s.Cluster | null>(null);

  const loadClusterDetail = async () => {
    loading.value = true;
    try {
      const { data, error } = await getK8sClusterDetail(props.clusterId);
      if (!error && data) {
        cluster.value = data;
      }
    } catch (error: unknown) {
      const err = error as Error;
      ElMessage.error(err.message || '加载集群详情失败');
    } finally {
      loading.value = false;
    }
  };

  onMounted(() => {
    loadClusterDetail();
  });
</script>

<template>
  <div class="p-4">
    <ElDescriptions v-if="cluster" v-loading="loading" bordered :column="1">
      <ElDescriptionsItem label="集群ID">{{ cluster.id }}</ElDescriptionsItem>
      <ElDescriptionsItem label="集群名称">{{ cluster.name }}</ElDescriptionsItem>
      <ElDescriptionsItem label="描述">{{ cluster.description || '-' }}</ElDescriptionsItem>
      <ElDescriptionsItem label="API 地址">{{ cluster.endpoint }}</ElDescriptionsItem>
      <ElDescriptionsItem label="集群类型">{{ cluster.clusterType }}</ElDescriptionsItem>
      <ElDescriptionsItem label="区域">{{ cluster.region || '-' }}</ElDescriptionsItem>
      <ElDescriptionsItem label="版本">{{ cluster.version || '-' }}</ElDescriptionsItem>
      <ElDescriptionsItem label="节点数">{{ cluster.nodeCount }}</ElDescriptionsItem>
      <ElDescriptionsItem label="状态">
        <span :class="cluster.status === 1 ? 'text-green-500' : 'text-red-500'">
          {{ cluster.status === 1 ? '正常' : '禁用' }}
        </span>
      </ElDescriptionsItem>
      <ElDescriptionsItem label="创建时间">{{ cluster.createdAt }}</ElDescriptionsItem>
      <ElDescriptionsItem label="更新时间">{{ cluster.updatedAt }}</ElDescriptionsItem>
    </ElDescriptions>

    <ElTabs v-if="cluster" class="mt-6">
      <ElTabPane label="节点" name="nodes">
        <ClusterNodes :cluster-id="clusterId" />
      </ElTabPane>
      <ElTabPane label="命名空间" name="namespaces">
        <ClusterNamespaces :cluster-id="clusterId" />
      </ElTabPane>
      <ElTabPane label="用户权限" name="users">
        <ClusterUsers :cluster-id="clusterId" />
      </ElTabPane>
    </ElTabs>
  </div>
</template>
