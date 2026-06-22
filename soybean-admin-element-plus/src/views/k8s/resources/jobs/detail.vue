<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElMessage, ElMessageBox } from 'element-plus';
import { ArrowLeft } from '@element-plus/icons-vue';
import yaml from 'js-yaml';
import { deleteK8sJob, fetchK8sEvents, getK8sJob, getK8sJobPods, updateK8sJob } from '@/service/api/k8s';
import { formatLabels, formatConditions } from '@/utils/k8s-formatters';
import YamlEditor from '@/components/YamlEditor.vue';
import K8sPodsTable from '@/components/k8s/K8sPodsTable.vue';

defineOptions({ name: 'K8sJobDetail' });

const route = useRoute();
const router = useRouter();
const message = ElMessage;

const clusterId = computed(() => Number(route.query.clusterId));
const namespace = computed(() => route.query.namespace as string);
const resourceName = computed(() => route.query.name as string);

const resource = ref<any>(null);
const pods = ref<any[]>([]);
const events = ref<any[]>([]);
const loading = ref(true);
const podsLoading = ref(false);
const eventsLoading = ref(false);
const activeTab = ref('basic');
const showYamlEditor = ref(false);
const yamlContent = ref('');
const yamlSaving = ref(false);

const statusType = computed(() => {
  if (!resource.value) return 'info';
  return 'success';
});

const statusText = computed(() => {
  return 'Job';
});

// 解析manifest为YAML
function parseManifest(manifestStr: string): string {
  if (!manifestStr) return '';
  try {
    const obj = JSON.parse(manifestStr);
    delete obj.managedFields;
    return yaml.dump(obj, { indent: 2, lineWidth: 120, noRefs: true });
  } catch (e) {
    return manifestStr;
  }
}

async function loadData() {
  loading.value = true;
  try {
    const response = await getK8sJob(clusterId.value, namespace.value, resourceName.value);
    const data = response.data || response;
    resource.value = data;
    yamlContent.value = parseManifest(data.manifest || '');
  } catch (error: any) {
    message.error(error.message || '加载失败');
  } finally {
    loading.value = false;
  }
}

async function loadEvents() {
  eventsLoading.value = true;
  try {
    const response = await fetchK8sEvents(
      clusterId.value,
      namespace.value,
      `involvedObject.name=${resourceName.value}`
    );
    events.value = response.data || response || [];
  } catch (error: any) {
    message.error(error.message || '加载事件失败');
  } finally {
    eventsLoading.value = false;
  }
}

async function loadPods() {
  if (!clusterId.value) return;
  podsLoading.value = true;
  try {
    const response = await getK8sJobPods(clusterId.value, namespace.value, resourceName.value);
    pods.value = response.data || response || [];
  } catch (error: any) {
    message.error(error.message || '加载容器组失败');
  } finally {
    podsLoading.value = false;
  }
}

function goBack() {
  router.back();
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(`确定要删除 Job "${resourceName.value}" 吗？此操作不可恢复！`, '确认删除', {
      type: 'warning'
    });

    await deleteK8sJob(clusterId.value, {
      namespace: namespace.value,
      name: resourceName.value
    });

    message.success('删除成功');
    router.back();
  } catch (error: any) {
    if (error !== 'cancel') {
      message.error(error.message || '删除失败');
    }
  }
}

async function handleYamlApply(yamlStr: string) {
  yamlSaving.value = true;
  try {
    const manifest = yaml.load(yamlStr);
    await updateK8sJob(clusterId.value, {
      namespace: namespace.value,
      manifest
    });
    message.success('更新成功');
    await loadData();
  } catch (error: any) {
    throw new Error(error.message || '更新失败');
  } finally {
    yamlSaving.value = false;
  }
}

async function handleTabChange(tab: string) {
  activeTab.value = tab;
  if (tab === 'pods' && pods.value.length === 0) {
    await loadPods();
  } else if (tab === 'events' && events.value.length === 0) {
    await loadEvents();
  }
}

onMounted(() => {
  loadData();
  loadPods();
  loadEvents();
});
</script>

<template>
  <div class="detail-page">
    <div class="page-header">
      <div class="header-left">
        <ElButton :icon="ArrowLeft" @click="goBack">返回</ElButton>
        <h2 class="title">{{ resourceName }}</h2>
        <ElTag :type="statusType">{{ statusText }}</ElTag>
      </div>
      <div class="header-actions">
        <ElButton type="primary" @click="showYamlEditor = true">编辑YAML</ElButton>
        <ElButton type="danger" @click="handleDelete">删除</ElButton>
      </div>
    </div>

    <ElTabs v-model="activeTab" class="detail-tabs" @tab-change="handleTabChange">
      <ElTabPane label="基本信息" name="basic">
        <ElDescriptions :column="2" border>
          <ElDescriptionsItem label="名称">{{ resource?.name }}</ElDescriptionsItem>
          <ElDescriptionsItem label="命名空间">{{ resource?.namespace }}</ElDescriptionsItem>
          <ElDescriptionsItem label="创建时间">{{ resource?.age }}</ElDescriptionsItem>
          <ElDescriptionsItem label="标签">
            <span style="font-size: 13px; color: #606266">{{ formatLabels(resource?.labels) }}</span>
          </ElDescriptionsItem>
        </ElDescriptions>
      </ElTabPane>

      <ElTabPane label="容器组" name="pods">
        <K8sPodsTable :pods="pods" :loading="podsLoading" />
      </ElTabPane>

      <ElTabPane label="事件" name="events">
        <ElTable
          v-loading="eventsLoading"
          :data="events"
          class="jobs-events-table"
          :header-cell-style="{ background: '#f5f7fa', color: '#303133', fontWeight: '600', paddingLeft: '16px', paddingRight: '16px' }"
          :row-style="{ backgroundColor: 'transparent' }"
          :cell-style="{ backgroundColor: 'transparent', padding: '8px 16px' }"
        >
          <ElTableColumn type="index" label="序号" width="60" />
          <ElTableColumn prop="type" label="类型" width="100" />
          <ElTableColumn prop="reason" label="原因" width="150" />
          <ElTableColumn prop="message" label="消息" />
          <ElTableColumn prop="count" label="次数" width="80" />
          <ElTableColumn prop="lastTimestamp" label="最后时间" width="180" />
        </ElTable>
      </ElTabPane>
    </ElTabs>

    <!-- YAML编辑器 -->
    <YamlEditor
      v-model="showYamlEditor"
      :title="`编辑 ${resourceName}`"
      :yaml="yamlContent"
      :can-edit="true"
      :on-apply="handleYamlApply"
    />
  </div>
</template>

<style scoped>
.detail-page {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.detail-tabs {
  background: var(--el-bg-color);
  border-radius: 8px;
  padding: 16px;
}

.yaml-viewer {
  background: #1e1e1e;
  border-radius: 4px;
  padding: 16px;
  max-height: 600px;
  overflow: auto;
}

.yaml-viewer pre {
  margin: 0;
  color: #d4d4d4;
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 13px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
}

/* 表格透明样式 - 完全覆盖 Element Plus 默认样式 */
.jobs-events-table,
.jobs-events-table.el-table,
:deep(.el-table),
:deep(.el-table__body),
:deep(.el-table__body-wrapper),
:deep(.el-table__inner-wrapper),
:deep(.el-table__header) {
  background-color: transparent !important;
}

/* 表头样式 */
:deep(.el-table__header-wrapper) {
  background-color: transparent !important;

  th.el-table__cell {
    background-color: #f5f7fa !important;
    color: #303133;
    font-weight: 600;
    text-align: left;
  }
}

/* 移除所有行的背景色 */
:deep(.el-table__body-wrapper) {
  background-color: transparent !important;
}

:deep(.el-table__body) {
  background-color: transparent !important;
}

:deep(.el-table__body tr) {
  background-color: transparent !important;
}

:deep(.el-table__body td.el-table__cell) {
  background-color: transparent !important;
}

/* 去掉斑马纹 */
:deep(.el-table--striped .el-table__body tr.el-table__row--striped) {
  background-color: transparent !important;

  td.el-table__cell {
    background-color: transparent !important;
  }
}

/* 移除 hover 效果的背景色 */
:deep(.el-table__body tr:hover > td.el-table__cell) {
  background-color: transparent !important;
}

/* 移除固定列的背景色 */
:deep(.el-table__fixed),
:deep(.el-table__fixed-body-wrapper) {
  background-color: transparent !important;

  .el-table__body tr {
    background-color: transparent !important;
  }

  .el-table__body td.el-table__cell {
    background-color: transparent !important;
  }
}

/* 移除表格容器的背景色 */
:deep(.el-table__inner-wrapper) {
  background-color: transparent !important;
}
</style>
