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
  import { deleteK8sIngress, fetchK8sEvents, getK8sIngress, updateK8sIngress } from '@/service/api/k8s';
  import YamlEditor from '@/components/k8s/YamlEditor.vue';
  import K8sResourceActionBar from '@/components/k8s/K8sResourceActionBar.vue';

  defineOptions({ name: 'K8sIngressDetail' });

  const route = useRoute();
  const router = useRouter();
  const message = ElMessage;

  const clusterId = computed(() => Number(route.query.clusterId));
  const namespace = computed(() => route.query.namespace as string);
  const resourceName = computed(() => route.query.name as string);

  const resource = ref<K8s.Ingress | null>(null);
  const events = ref<K8s.Event[]>([]);
  const rules = ref<K8s.IngressRule[]>([]);
  const tls = ref<K8s.IngressTLS[]>([]);
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

  // 返回列表页的路径 - Ingress 返回到网络页面
  const backPath = '/k8s/network';

  async function loadData() {
    loading.value = true;
    try {
      const response = await getK8sIngress(clusterId.value, namespace.value, resourceName.value);
      const data = response.data || response;
      resource.value = data;
      yamlContent.value = parseManifest(data.manifest || '');
      rules.value = data.rules || [];
      tls.value = data.tls || [];
    } catch (error: unknown) {
      const err = error as Error;
      message.error(err.message || '加载失败');
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
    } catch (error: unknown) {
      const err = error as Error;
      message.error(err.message || '加载事件失败');
    } finally {
      eventsLoading.value = false;
    }
  }

  async function handleDelete() {
    try {
      await ElMessageBox.confirm(`确定要删除 Ingress "${resourceName.value}" 吗？`, '确认删除', { type: 'warning' });

      await deleteK8sIngress(clusterId.value, {
        namespace: namespace.value,
        name: resourceName.value
      });

      message.success('删除成功');
      router.push(backPath);
    } catch (error: unknown) {
      if (error !== 'cancel') {
        const err = error as Error;
        message.error(err.message || '删除失败');
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
    } catch (error: unknown) {
      const err = error as Error;
      throw new Error(err.message || '更新失败');
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
  <div class="detail-page bg-layout">
    <!-- 顶部操作栏 - 使用新组件 -->
    <K8sResourceActionBar
      :name="resourceName || '-'"
      :namespace="namespace || '-'"
      :status-tag="{ type: 'info', text: 'Ingress' }"
      :back-path="backPath"
      :actions="[
        { label: '编辑YAML', type: 'primary', handler: () => (showYamlEditor = true), tooltip: '编辑 YAML 配置' },
        { label: '删除', type: 'danger', handler: handleDelete, tooltip: '删除 Ingress（危险操作）' }
      ]"
    />

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
