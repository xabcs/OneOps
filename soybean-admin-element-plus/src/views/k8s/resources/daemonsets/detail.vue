<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElButton, ElDialog, ElMessage, ElTabPane, ElTable, ElTableColumn, ElTabs, ElTag } from 'element-plus';
import yaml from 'js-yaml';
import { fetchK8sEvents, getK8sDaemonSet, getK8sDaemonSetPods, updateK8sDaemonSet } from '@/service/api/k8s';
import { formatAnnotations, formatConditions, formatImages, formatLabels, formatSelectors, formatStrategy } from '@/utils/k8s-formatters';
import K8sBasicInfoGrid from '@/components/k8s/K8sBasicInfoGrid.vue';
import K8sPodsTable from '@/components/k8s/K8sPodsTable.vue';
import K8sEventsTable from '@/components/k8s/K8sEventsTable.vue';

defineOptions({ name: 'K8sDaemonSetDetail' });

const route = useRoute();
const router = useRouter();

const resource = ref<any>(null);
const pods = ref<any[]>([]);
const events = ref<any[]>([]);
const loading = ref(true);
const podsLoading = ref(false);
const eventsLoading = ref(false);
const activeTab = ref('pods');

const clusterId = ref<number>(0);
const namespace = ref('');
const resourceName = ref('');

// YAML 编辑相关
const showYamlEditor = ref(false);
const yamlContent = ref('');
const yamlEditingContent = ref('');
const yamlSaving = ref(false);

const getStatusTag = computed(() => {
  if (!resource.value) return { type: 'info', text: '未知' };
  const ready = resource.value.ready || 0;
  const desired = resource.value.desired || 0;
  if (desired === 0) return { type: 'info', text: '未就绪' };
  if (ready === desired) return { type: 'success', text: `运行中 (${ready}/${desired})` };
  if (ready > 0) return { type: 'warning', text: `部分就绪 (${ready}/${desired})` };
  return { type: 'danger', text: '未就绪' };
});

// 基本信息
const basicInfoFields = computed(() => {
  if (!resource.value) return [[]];

  const fields = [
    [
      { label: '名称', value: resource.value.name || '-' },
      { label: '命名空间', value: resource.value.namespace || '-' },
      { label: '创建时间', value: resource.value.age || '-' }
    ],
    [
      { label: '期望节点数', value: resource.value.desired?.toString() || '0' },
      { label: '当前节点数', value: resource.value.current?.toString() || '0' },
      { label: '就绪数量', value: resource.value.ready?.toString() || '0' }
    ],
    [
      { label: '已更新数量', value: resource.value.updated?.toString() || '0' },
      { label: '可用数量', value: resource.value.available?.toString() || '0' }
    ]
  ];

  if (resource.value.labels) {
    fields.push([
      { label: '标签', value: formatLabels(resource.value.labels), fullRow: true, isTags: true }
    ]);
  }

  if (resource.value.selector && Object.keys(resource.value.selector).length > 0) {
    fields.push([
      { label: '选择器', value: formatSelectors(resource.value.selector), fullRow: true }
    ]);
  }

  if (resource.value.strategy) {
    fields.push([
      { label: '更新策略', value: formatStrategy(resource.value.strategy), fullRow: false },
      { label: '策略类型', value: resource.value.strategy.type || '-', fullRow: false }
    ]);
  }

  if (resource.value.annotations && Object.keys(resource.value.annotations).length > 0) {
    const annotationStr = formatAnnotations(resource.value.annotations);
    if (annotationStr !== '-') {
      fields.push([
        { label: '注解', value: annotationStr, fullRow: true }
      ]);
    }
  }

  if (resource.value.conditions?.length) {
    fields.push([
      { label: '状态条件', value: formatConditions(resource.value.conditions), fullRow: true, isConditions: true }
    ]);
  }

  return fields;
});

const getPodStatusTag = (pod: any) => {
  switch (pod.phase) {
    case 'Running':
      return { type: 'success', text: '运行中' };
    case 'Succeeded':
      return { type: 'info', text: '已完成' };
    case 'Failed':
      return { type: 'danger', text: '失败' };
    case 'Pending':
      return { type: 'warning', text: '等待中' };
    default:
      return { type: 'info', text: '未知' };
  }
};

async function loadDetail() {
  const qClusterId = route.query.clusterId as string;
  const qNamespace = route.query.namespace as string;
  const qName = route.query.name as string;

  if (!qClusterId || !qNamespace || !qName) {
    ElMessage.error('参数不完整');
    router.push('/k8s/workloads');
    return;
  }

  clusterId.value = Number(qClusterId);
  namespace.value = qNamespace;
  resourceName.value = qName;

  loading.value = true;
  try {
    const res = await getK8sDaemonSet(clusterId.value, namespace.value, resourceName.value);
    resource.value = res.data || res;
    if (activeTab.value === 'pods') await loadPods();
  } catch (error: any) {
    ElMessage.error(error.message || '获取详情失败');
  } finally {
    loading.value = false;
  }
}

async function loadPods() {
  if (!clusterId.value) return;
  podsLoading.value = true;
  try {
    const res = await getK8sDaemonSetPods(clusterId.value, namespace.value, resourceName.value);
    const podData = res.data || res;
    pods.value = Array.isArray(podData) ? podData : [];
  } catch (error: any) {
    ElMessage.error(error.message || '获取 Pods 失败');
    pods.value = [];
  } finally {
    podsLoading.value = false;
  }
}

async function loadEvents() {
  if (!clusterId.value) return;
  eventsLoading.value = true;
  try {
    const fieldSelector = `involvedObject.name=${resourceName.value},involvedObject.kind=DaemonSet`;
    const res = await fetchK8sEvents(clusterId.value, namespace.value, fieldSelector);
    const eventData = res.data || res;
    events.value = Array.isArray(eventData) ? eventData : [];
  } catch (error: any) {
    ElMessage.error(error.message || '获取 Events 失败');
    events.value = [];
  } finally {
    eventsLoading.value = false;
  }
}

async function handleTabChange(tab: string) {
  activeTab.value = tab;
  if (tab === 'pods' && pods.value.length === 0) await loadPods();
  if (tab === 'events' && events.value.length === 0) await loadEvents();
}

// 打开 YAML 编辑器
function handleEditYaml() {
  try {
    // 后端返回的是 JSON 字符串，需要先解析为对象
    const manifestObj = resource.value?.manifest ? JSON.parse(resource.value.manifest) : {};
    // 转换为格式化的 YAML
    yamlEditingContent.value = yaml.dump(manifestObj, {
      indent: 2,
      lineWidth: -1,
      noRefs: true
    });
    yamlContent.value = yamlEditingContent.value;
    showYamlEditor.value = true;
  } catch (error) {
    console.error('解析 manifest 失败:', error);
    ElMessage.error('解析 YAML 失败');
  }
}

// 关闭 YAML 编辑器
function closeYamlEditor() {
  showYamlEditor.value = false;
  yamlEditingContent.value = '';
}

// 保存 YAML
async function handleSaveYaml() {
  yamlSaving.value = true;
  try {
    // 将 YAML 转换回 JSON 对象
    const manifestObj = yaml.load(yamlEditingContent.value);
    // 调用更新 API，将 manifestObj 传递给后端
    await updateK8sDaemonSet(clusterId.value, {
      namespace: namespace.value,
      manifest: manifestObj
    });
    ElMessage.success('保存成功');
    closeYamlEditor();
    // 刷新数据
    await loadDetail();
  } catch (error: any) {
    console.error('保存 YAML 失败:', error);
    ElMessage.error(`保存失败: ${error.message || 'YAML 格式错误'}`);
  } finally {
    yamlSaving.value = false;
  }
}

onMounted(() => loadDetail());
</script>

<template>
  <div v-loading="loading" class="detail-page">
    <!-- 顶部操作栏 -->
    <div class="instance-bar">
      <div class="instance-left">
        <span class="back-arrow" @click="router.push('/k8s/workloads')">←</span>
        <span class="instance-name">{{ resource?.name || '-' }}</span>
        <span class="instance-meta">命名空间: {{ resource?.namespace || '-' }}</span>
        <span class="instance-meta">节点: {{ resource?.ready || 0 }} / {{ resource?.desired || 0 }}</span>
        <ElTag v-if="resource" :type="getStatusTag.type" size="small">{{ getStatusTag.text }}</ElTag>
      </div>
      <div class="instance-actions">
        <ElButton size="small" @click="handleEditYaml">YAML 编辑</ElButton>
        <ElButton size="small" @click="router.push('/k8s/workloads')">返回列表</ElButton>
      </div>
    </div>

    <!-- 基本信息 -->
    <div class="basic-info-section">
      <div class="info-section">
        <h3 class="section-title">基本信息</h3>
        <K8sBasicInfoGrid :fields="basicInfoFields" />
      </div>
    </div>

    <!-- 标签切换 -->
    <ElTabs v-model="activeTab" class="detail-tabs" @tab-change="handleTabChange">
      <ElTabPane label="容器组" name="pods">
        <div class="tab-content">
          <div class="section-header">
            <span class="section-title">共 {{ pods.length }} 个容器组</span>
            <ElButton size="small" @click="loadPods">刷新</ElButton>
          </div>
          <K8sPodsTable :pods="pods" :loading="podsLoading" />
        </div>
      </ElTabPane>

      <ElTabPane label="事件" name="events">
        <div class="tab-content">
          <div class="section-header">
            <span class="section-title">共 {{ events.length }} 个事件</span>
            <ElButton size="small" @click="loadEvents">刷新</ElButton>
          </div>
          <K8sEventsTable :events="events" :loading="eventsLoading" />
        </div>
      </ElTabPane>
    </ElTabs>

    <!-- YAML 编辑对话框 -->
    <ElDialog v-model="showYamlEditor" title="YAML 编辑" width="900px" top="5vh" @close="closeYamlEditor">
      <div class="yaml-editor-wrapper">
        <textarea v-model="yamlEditingContent" class="yaml-editor" spellcheck="false"></textarea>
      </div>
      <template #footer>
        <ElButton @click="closeYamlEditor">取消</ElButton>
        <ElButton type="primary" :loading="yamlSaving" @click="handleSaveYaml">保存</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<style scoped>
.detail-page {
  min-height: 100vh;
  padding: 24px;
}
.instance-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.instance-left {
  display: flex;
  align-items: center;
  gap: 16px;
}
.back-arrow {
  font-size: 20px;
  color: #0052d9;
  cursor: pointer;
  transition: opacity 0.2s;
}
.back-arrow:hover {
  opacity: 0.8;
}
.instance-name {
  font-size: 16px;
  font-weight: 500;
  color: #303133;
}
.instance-meta {
  font-size: 14px;
  color: #909399;
}
.instance-actions {
  display: flex;
  gap: 8px;
}
.basic-info-section {
  background: transparent;
  margin-bottom: 24px;
}
.info-section {
  background: transparent;
  margin-bottom: 16px;
}
.section-title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 12px;
}
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}
.tab-content {
  padding: 16px;
}
.detail-tabs {
  margin-top: 16px;
}
.desc-grid {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.desc-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 24px;
}
.desc-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.desc-item-full {
  grid-column: 1 / -1;
}
.item-label {
  font-size: 12px;
  color: #909399;
}
.item-content {
  font-size: 14px;
  color: #303133;
}
.value-text {
  color: #303133;
}

/* 镜像列换行显示 */
.whitespace-pre-line {
  white-space: pre-line;
  word-break: break-all;
}

.font-mono {
  font-family: 'Courier New', Courier, monospace;
  word-break: break-all;
}
.tag-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.yaml-editor-wrapper {
  width: 100%;
  height: 500px;
}
.yaml-editor {
  width: 100%;
  height: 100%;
  padding: 12px;
  font-family: 'Courier New', Courier, monospace;
  font-size: 14px;
  line-height: 1.5;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  resize: none;
  background-color: #f5f7fa;
  color: #303133;
}
.yaml-editor:focus {
  outline: none;
  border-color: #409eff;
}
</style>
