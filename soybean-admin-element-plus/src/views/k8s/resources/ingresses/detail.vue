<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElMessage, ElMessageBox } from 'element-plus';
import { ArrowLeft } from '@element-plus/icons-vue';
import yaml from 'js-yaml';
import { deleteK8sIngress, fetchK8sEvents, getK8sIngress, updateK8sIngress } from '@/service/api/k8s';
import YamlEditor from '@/components/YamlEditor.vue';

defineOptions({ name: 'K8sIngressDetail' });

const route = useRoute();
const router = useRouter();
const message = ElMessage;

const clusterId = computed(() => Number(route.query.clusterId));
const namespace = computed(() => route.query.namespace as string);
const resourceName = computed(() => route.query.name as string);

const resource = ref<any>(null);
const events = ref<any[]>([]);
const rules = ref<any[]>([]);
const tls = ref<any[]>([]);
const loading = ref(true);
const eventsLoading = ref(false);
const activeTab = ref('basic');
const showYamlEditor = ref(false);
const yamlContent = ref('');
const yamlSaving = ref(false);

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
    const response = await getK8sIngress(clusterId.value, namespace.value, resourceName.value);
    const data = response.data || response;
    resource.value = data;
    yamlContent.value = parseManifest(data.manifest || '');
    rules.value = data.rules || [];
    tls.value = data.tls || [];
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

function goBack() {
  router.back();
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(`确定要删除 Ingress "${resourceName.value}" 吗？`, '确认删除', { type: 'warning' });

    await deleteK8sIngress(clusterId.value, {
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
    await updateK8sIngress(clusterId.value, {
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
  loadData();
  loadEvents();
});
</script>

<template>
  <div class="detail-page">
    <div class="page-header">
      <div class="header-left">
        <ElButton :icon="ArrowLeft" @click="goBack">返回</ElButton>
        <h2 class="title">{{ resourceName }}</h2>
        <ElTag type="info">Ingress</ElTag>
      </div>
      <div class="header-actions">
        <ElButton type="primary" @click="showYamlEditor = true">编辑YAML</ElButton>
        <ElButton type="danger" @click="handleDelete">删除</ElButton>
      </div>
    </div>

    <ElTabs v-model="activeTab" class="detail-tabs">
      <ElTabPane label="基本信息" name="basic">
        <ElDescriptions :column="2" border>
          <ElDescriptionsItem label="名称">{{ resource?.name }}</ElDescriptionsItem>
          <ElDescriptionsItem label="命名空间">{{ resource?.namespace }}</ElDescriptionsItem>
          <ElDescriptionsItem label="Ingress类">{{ resource?.ingressClassName || '-' }}</ElDescriptionsItem>
          <ElDescriptionsItem label="创建时间">{{ resource?.age }}</ElDescriptionsItem>
          <ElDescriptionsItem label="地址">
            <span v-if="resource?.addresses && resource.addresses.length">
              <ElTag v-for="ip in resource.addresses" :key="ip" size="small" style="margin-right: 4px">{{ ip }}</ElTag>
            </span>
            <span v-else>-</span>
          </ElDescriptionsItem>
          <ElDescriptionsItem label="注解">
            <span v-if="resource?.annotations">
              <ElTag v-for="(value, key) in resource.annotations" :key="key" size="small" style="margin-right: 4px">
                {{ key }}: {{ value }}
              </ElTag>
            </span>
            <span v-else>-</span>
          </ElDescriptionsItem>
        </ElDescriptions>

        <h4 style="margin-top: 20px">规则</h4>
        <ElTable v-loading="!rules.length" :data="rules" border>
          <ElTableColumn type="index" label="序号" width="60" />
          <ElTableColumn prop="host" label="主机" />
          <ElTableColumn label="路径" min-width="300">
            <template #default="{ row }">
              <div v-if="row.http && row.http.paths">
                <div v-for="(path, idx) in row.http.paths" :key="idx" class="path-item">
                  <ElTag size="small">{{ path.path }}</ElTag>
                  <span class="path-backend">{{ path.serviceName }}:{{ path.servicePort }}</span>
                </div>
              </div>
              <span v-else>-</span>
            </template>
          </ElTableColumn>
        </ElTable>

        <h4 v-if="tls && tls.length" style="margin-top: 20px">TLS</h4>
        <ElTable v-if="tls && tls.length" :data="tls" border>
          <ElTableColumn type="index" label="序号" width="60" />
          <ElTableColumn prop="hosts" label="主机">
            <template #default="{ row }">
              <ElTag v-for="host in row.hosts" :key="host" size="small" style="margin-right: 4px">{{ host }}</ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="secretName" label="密钥" />
        </ElTable>
      </ElTabPane>

      <ElTabPane label="YAML" name="yaml">
        <div class="yaml-viewer">
          <pre>{{ yamlContent }}</pre>
        </div>
      </ElTabPane>

      <ElTabPane label="事件" name="events">
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
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid #ebeef5;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.detail-tabs {
  background: #fff;
  padding: 16px;
  border-radius: 4px;
}

.yaml-viewer {
  background: #1e1e1e;
  padding: 16px;
  border-radius: 4px;
  max-height: 500px;
  overflow: auto;
}

.yaml-viewer pre {
  margin: 0;
  color: #d4d4d4;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  white-space: pre-wrap;
  word-break: break-all;
}

.path-item {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.path-backend {
  color: #409eff;
  font-size: 12px;
}
</style>
