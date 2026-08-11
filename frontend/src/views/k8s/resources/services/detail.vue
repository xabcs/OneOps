<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import {
  ElButton,
  ElDescriptions,
  ElDescriptionsItem,
  ElMessage,
  ElMessageBox,
  ElTabPane,
  ElTable,
  ElTableColumn,
  ElTabs,
  ElTag
} from 'element-plus';
import yaml from 'js-yaml';
import { deleteK8sService, fetchK8sEvents, getK8sService, updateK8sService } from '@/service/api/k8s';
import YamlEditor from '@/components/YamlEditor.vue';
import K8sResourceActionBar from '@/components/k8s/K8sResourceActionBar.vue';

defineOptions({ name: 'K8sServiceDetail' });

const route = useRoute();
const router = useRouter();
const message = ElMessage;

const clusterId = computed(() => Number(route.query.clusterId));
const namespace = computed(() => route.query.namespace as string);
const resourceName = computed(() => route.query.name as string);

const resource = ref<any>(null);
const events = ref<any[]>([]);
const ports = ref<any[]>([]);
const loading = ref(true);
const eventsLoading = ref(false);
const activeTab = ref('basic');
const showYamlEditor = ref(false);
const yamlContent = ref('');
const yamlSaving = ref(false);

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

// 返回列表页的路径 - Services 返回到网络页面
const backPath = '/k8s/network';

async function loadData() {
  loading.value = true;
  try {
    const response = await getK8sService(clusterId.value, namespace.value, resourceName.value);
    const data = response.data || response;
    resource.value = data;
    yamlContent.value = parseManifest(data.manifest || '');
    ports.value = data.ports || [];
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

async function handleDelete() {
  try {
    await ElMessageBox.confirm(`确定要删除 Service "${resourceName.value}" 吗？`, '确认删除', { type: 'warning' });

    await deleteK8sService(clusterId.value, {
      namespace: namespace.value,
      name: resourceName.value
    });

    message.success('删除成功');
    router.push(backPath);
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
    await updateK8sService(clusterId.value, {
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

onMounted(() => {
  console.log('[Service详情] onMounted');
  console.log('[Service详情] 当前路由参数:', {
    clusterId: route.query.clusterId,
    namespace: route.query.namespace,
    name: route.query.name
  });

  loadData();
  loadEvents();
});
</script>

<template>
  <div class="detail-page bg-layout">
    <!-- 顶部操作栏 - 使用新组件 -->
    <K8sResourceActionBar
      :name="resourceName || '-'"
      :namespace="namespace || '-'"
      :status-tag="{ type: 'info', text: 'Service' }"
      :back-path="backPath"
      :actions="[
        { label: '编辑YAML', type: 'primary', handler: () => (showYamlEditor = true), tooltip: '编辑 YAML 配置' },
        { label: '删除', type: 'danger', handler: handleDelete, tooltip: '删除 Service（危险操作）' }
      ]"
    />

    <ElTabs v-model="activeTab" class="detail-tabs">
      <ElTabPane label="基本信息" name="basic">
        <ElDescriptions :column="2" border>
          <ElDescriptionsItem label="名称">{{ resource?.name }}</ElDescriptionsItem>
          <ElDescriptionsItem label="命名空间">{{ resource?.namespace }}</ElDescriptionsItem>
          <ElDescriptionsItem label="类型">{{ resource?.type }}</ElDescriptionsItem>
          <ElDescriptionsItem label="集群IP">{{ resource?.clusterIP }}</ElDescriptionsItem>
          <ElDescriptionsItem label="外部IP">
            <span v-if="resource?.externalIP && resource.externalIP.length">
              <ElTag v-for="ip in resource.externalIP" :key="ip" size="small" style="margin-right: 4px">{{ ip }}</ElTag>
            </span>
            <span v-else>-</span>
          </ElDescriptionsItem>
          <ElDescriptionsItem label="创建时间">{{ resource?.age }}</ElDescriptionsItem>
          <ElDescriptionsItem label="标签">
            <ElTag v-for="(value, key) in resource?.labels" :key="key" size="small" style="margin-right: 4px">
              {{ key }}: {{ value }}
            </ElTag>
          </ElDescriptionsItem>
          <ElDescriptionsItem label="选择器">
            <span v-if="resource?.selector">
              <ElTag v-for="(value, key) in resource.selector" :key="key" size="small" style="margin-right: 4px">
                {{ key }}: {{ value }}
              </ElTag>
            </span>
            <span v-else>-</span>
          </ElDescriptionsItem>
        </ElDescriptions>

        <h4 style="margin-top: 20px">端口</h4>
        <ElTable :data="ports" border>
          <ElTableColumn type="index" label="序号" width="60" />
          <ElTableColumn prop="name" label="名称" />
          <ElTableColumn prop="protocol" label="协议" width="80" />
          <ElTableColumn prop="port" label="端口" width="80" />
          <ElTableColumn prop="targetPort" label="目标端口" width="100" />
          <ElTableColumn prop="nodePort" label="NodePort" width="100">
            <template #default="{ row }">
              {{ row.nodePort || '-' }}
            </template>
          </ElTableColumn>
        </ElTable>
      </ElTabPane>

      <ElTabPane label="YAML" name="yaml">
        <div class="tab-toolbar">
          <ElButton type="primary" size="small" @click="showYamlEditor = true">编辑 YAML</ElButton>
        </div>
        <div class="yaml-viewer">
          <pre>{{ yamlContent }}</pre>
        </div>
      </ElTabPane>

      <ElTabPane :label="`事件 (${events.length})`" name="events">
        <div class="tab-toolbar">
          <ElButton size="small" @click="loadEvents">刷新</ElButton>
        </div>
        <ElTable v-loading="eventsLoading" :data="events">
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
  min-height: 100vh;
  padding: 24px;
}

.tab-toolbar {
  display: flex;
  justify-content: flex-end;
  padding: 0 16px 8px 16px;
}

.detail-tabs {
  margin-top: 16px;
}

.yaml-viewer {
  background: #1e1e1e;
  padding: 16px;
  border-radius: 4px;
  max-height: 600px;
  overflow: auto;
}

.yaml-viewer pre {
  margin: 0;
  color: #d4d4d4;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
