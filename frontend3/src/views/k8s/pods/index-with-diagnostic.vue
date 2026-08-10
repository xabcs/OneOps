<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import {
  CircleCheck,
  CircleClose,
  Delete,
  Document,
  Monitor,
  MoreFilled,
  Operation,
  Refresh,
  Search,
  View,
  Warning
} from '@element-plus/icons-vue';

const router = useRouter();

// 数据状态
const loading = ref(false);
const clusters = ref([]);
const namespaces = ref([]);
const pods = ref([]);

// 过滤条件
const filters = ref({
  clusterId: '',
  namespace: '',
  status: '',
  search: ''
});

// 分页
const pagination = ref({
  page: 1,
  pageSize: 20,
  total: 0
});

// 计算属性
const filteredPods = computed(() => {
  let result = pods.value;

  if (filters.value.status) {
    result = result.filter(pod => pod.status === filters.value.status);
  }

  if (filters.value.search) {
    const search = filters.value.search.toLowerCase();
    result = result.filter(
      pod => pod.name.toLowerCase().includes(search) || pod.namespace.toLowerCase().includes(search)
    );
  }

  return result;
});

// 生命周期
onMounted(() => {
  initializeData();
});

// 方法
const initializeData = async () => {
  await Promise.all([loadClusters(), loadNamespaces()]);

  // 自动选择第一个集群和命名空间
  if (clusters.value.length > 0) {
    filters.value.clusterId = clusters.value[0].id;
    await loadNamespaces();
  }

  if (namespaces.value.length > 0) {
    filters.value.namespace = namespaces.value[0].name;
    await loadPods();
  }
};

const loadClusters = async () => {
  try {
    const response = await fetch('/api/v1/kubernetes/clusters');
    const data = await response.json();
    clusters.value = data.data || [];
  } catch (error) {
    console.error('加载集群失败:', error);
  }
};

const loadNamespaces = async () => {
  if (!filters.value.clusterId) return;

  try {
    const response = await fetch(`/api/v1/kubernetes/clusters/${filters.value.clusterId}/namespaces`);
    const data = await response.json();
    namespaces.value = data.data || [];
  } catch (error) {
    console.error('加载命名空间失败:', error);
  }
};

const loadPods = async () => {
  if (!filters.value.clusterId || !filters.value.namespace) return;

  loading.value = true;
  try {
    const response = await fetch(`/api/v1/kubernetes/pods/${filters.value.clusterId}/${filters.value.namespace}`);
    const data = await response.json();

    pods.value = (data.data || []).map(pod => ({
      ...pod,
      // 检查诊断能力
      diagnostic: checkDiagnosticCapability(pod)
    }));

    pagination.value.total = pods.value.length;
  } catch (error) {
    console.error('加载Pods失败:', error);
  } finally {
    loading.value = false;
  }
};

// 检查Pod的诊断能力
const checkDiagnosticCapability = (pod: any) => {
  // 检查是否有Java容器
  const hasJavaContainer = pod.containers?.some((container: any) => {
    const image = container.image.toLowerCase();
    return image.includes('java') || image.includes('jdk') || image.includes('tomcat') || image.includes('spring');
  });

  // 检查是否有Arthas sidecar
  const hasArthasSidecar = pod.containers?.some((container: any) => {
    return container.name === 'arthas-agent' || container.image.includes('arthas');
  });

  if (hasArthasSidecar) {
    return { enabled: true, method: 'sidecar' };
  }

  if (hasJavaContainer) {
    // 假设有DaemonSet agent
    return { enabled: true, method: 'daemonset' };
  }

  return { enabled: false };
};

const onClusterChange = () => {
  filters.value.namespace = '';
  loadNamespaces();
  loadPods();
};

const onSearch = () => {
  // 搜索由计算属性自动处理
};

const refreshPods = () => {
  loadPods();
};

// 新增：快速诊断
const quickDiagnose = (pod: any) => {
  router.push({
    name: 'K8sDiagnostic',
    query: {
      cluster: filters.value.clusterId,
      namespace: pod.namespace,
      pod: pod.name
    }
  });
};

// 跳转到诊断中心
const goToDiagnostic = () => {
  router.push({
    name: 'K8sDiagnostic',
    query: {
      cluster: filters.value.clusterId,
      namespace: filters.value.namespace
    }
  });
};

const viewLogs = (pod: any) => {
  // 跳转到日志页面
  router.push({
    name: 'K8sLogs',
    query: {
      cluster: filters.value.clusterId,
      namespace: pod.namespace,
      pod: pod.name
    }
  });
};

const viewDetails = (pod: any) => {
  // 显示Pod详情对话框
  console.log('查看详情:', pod);
};

const handleCommand = (pod: any, command: string) => {
  switch (command) {
    case 'exec':
      console.log('进入容器:', pod);
      break;
    case 'delete':
      console.log('删除Pod:', pod);
      break;
  }
};

// 辅助方法
const getPodStatusColor = (status: string) => {
  const colors = {
    Running: 'success',
    Pending: 'warning',
    Failed: 'danger',
    Succeeded: 'info'
  };
  return colors[status] || 'info';
};

const getPodStatusIcon = (status: string) => {
  const icons = {
    Running: CircleCheck,
    Pending: Warning,
    Failed: CircleClose,
    Succeeded: CircleCheck
  };
  return icons[status] || Warning;
};

const getPodStatusType = (status: string) => {
  const types = {
    Running: 'success',
    Pending: 'warning',
    Failed: 'danger',
    Succeeded: 'info'
  };
  return types[status] || 'info';
};

const formatTime = (timestamp: string) => {
  return new Date(timestamp).toLocaleString('zh-CN');
};
</script>

<template>
  <div class="k8s-pods-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <h2>Pod管理</h2>
      <div class="header-actions">
        <ElButton :icon="Refresh" :loading="loading" @click="refreshPods">刷新</ElButton>
        <ElButton :icon="Operation" type="primary" @click="goToDiagnostic">诊断中心</ElButton>
      </div>
    </div>

    <!-- 过滤条件 -->
    <div class="filter-section">
      <ElForm :inline="true" :model="filters" class="filter-form">
        <ElFormItem label="集群">
          <ElSelect v-model="filters.clusterId" placeholder="选择集群" style="width: 180px" @change="onClusterChange">
            <ElOption v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" />
          </ElSelect>
        </ElFormItem>

        <ElFormItem label="命名空间">
          <ElSelect v-model="filters.namespace" placeholder="选择命名空间" style="width: 150px" @change="loadPods">
            <ElOption v-for="ns in namespaces" :key="ns.name" :label="ns.name" :value="ns.name" />
          </ElSelect>
        </ElFormItem>

        <ElFormItem label="状态">
          <ElSelect v-model="filters.status" placeholder="全部状态" clearable style="width: 120px" @change="loadPods">
            <ElOption label="运行中" value="Running" />
            <ElOption label="等待中" value="Pending" />
            <ElOption label="失败" value="Failed" />
            <ElOption label="成功" value="Succeeded" />
          </ElSelect>
        </ElFormItem>

        <ElFormItem>
          <ElInput v-model="filters.search" placeholder="搜索Pod名称" style="width: 200px" clearable @input="onSearch">
            <template #prefix>
              <ElIcon><Search /></ElIcon>
            </template>
          </ElInput>
        </ElFormItem>
      </ElForm>
    </div>

    <!-- Pod表格 -->
    <ElTable v-loading="loading" :data="filteredPods" stripe style="width: 100%">
      <ElTableColumn prop="name" label="Pod名称" width="200">
        <template #default="{ row }">
          <div class="pod-name">
            <ElIcon :color="getPodStatusColor(row.status)">
              <component :is="getPodStatusIcon(row.status)" />
            </ElIcon>
            {{ row.name }}
          </div>
        </template>
      </ElTableColumn>

      <ElTableColumn prop="namespace" label="命名空间" width="120" />

      <ElTableColumn prop="podIp" label="Pod IP" width="120" />

      <ElTableColumn prop="nodeName" label="节点" width="150" />

      <ElTableColumn prop="status" label="状态" width="100">
        <template #default="{ row }">
          <ElTag :type="getPodStatusType(row.status)" size="small">
            {{ row.status }}
          </ElTag>
        </template>
      </ElTableColumn>

      <ElTableColumn prop="createTime" label="创建时间" width="160">
        <template #default="{ row }">
          {{ formatTime(row.createTime) }}
        </template>
      </ElTableColumn>

      <!-- 新增：诊断状态列 -->
      <ElTableColumn label="诊断" width="100">
        <template #default="{ row }">
          <ElTag
            v-if="row.diagnostic?.enabled"
            :type="row.diagnostic.method === 'sidecar' ? 'success' : 'primary'"
            size="small"
            effect="plain"
          >
            {{ row.diagnostic.method === 'sidecar' ? 'Sidecar' : 'DaemonSet' }}
          </ElTag>
          <ElTag v-else type="info" size="small" effect="plain">不可诊断</ElTag>
        </template>
      </ElTableColumn>

      <ElTableColumn label="操作" width="300" fixed="right">
        <template #default="{ row }">
          <!-- 新增：诊断按钮 -->
          <ElButton
            v-if="row.diagnostic?.enabled"
            type="primary"
            size="small"
            :icon="Operation"
            @click="quickDiagnose(row)"
          >
            诊断
          </ElButton>

          <ElButton size="small" :icon="Document" @click="viewLogs(row)">日志</ElButton>

          <ElButton size="small" :icon="View" @click="viewDetails(row)">详情</ElButton>

          <ElDropdown @command="handleCommand(row, $event)">
            <ElButton size="small" :icon="MoreFilled">更多</ElButton>
            <template #dropdown>
              <ElDropdownMenu>
                <ElDropdownItem command="exec">
                  <ElIcon><Monitor /></ElIcon>
                  进入容器
                </ElDropdownItem>
                <ElDropdownItem command="delete">
                  <ElIcon><Delete /></ElIcon>
                  删除Pod
                </ElDropdownItem>
              </ElDropdownMenu>
            </template>
          </ElDropdown>
        </template>
      </ElTableColumn>
    </ElTable>

    <!-- 分页 -->
    <div class="pagination">
      <ElPagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="loadPods"
        @current-change="loadPods"
      />
    </div>
  </div>
</template>

<style scoped lang="scss">
.k8s-pods-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;

    h2 {
      margin: 0;
      font-size: 20px;
    }

    .header-actions {
      display: flex;
      gap: 12px;
    }
  }

  .filter-section {
    background: #fff;
    padding: 16px;
    border-radius: 4px;
    margin-bottom: 20px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  .pod-name {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .pagination {
    display: flex;
    justify-content: center;
    margin-top: 20px;
  }
}
</style>
