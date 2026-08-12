<script setup lang="ts">
  import { computed } from 'vue';

  defineOptions({ name: 'ServiceDetailDialog' });

  const props = defineProps<{ detail: K8s.Service | null }>();

  const visible = defineModel<boolean>('visible', { default: false });

  const title = computed(() => `Service: ${props.detail?.name ?? ''}`);
</script>

<template>
  <ElDialog v-model="visible" :title="title" width="800px">
    <div v-if="detail" class="space-y-4">
      <!-- 基本信息 -->
      <div class="grid grid-cols-2 gap-4">
        <div>
          <span class="text-sm text-gray-700 font-medium">名称:</span>
          <span class="ml-2">{{ detail.name }}</span>
        </div>
        <div>
          <span class="text-sm text-gray-700 font-medium">命名空间:</span>
          <span class="ml-2">{{ detail.namespace }}</span>
        </div>
        <div>
          <span class="text-sm text-gray-700 font-medium">类型:</span>
          <span class="ml-2">{{ detail.type }}</span>
        </div>
        <div>
          <span class="text-sm text-gray-700 font-medium">Cluster IP:</span>
          <span class="ml-2">{{ detail.clusterIP || '-' }}</span>
        </div>
        <div>
          <span class="text-sm text-gray-700 font-medium">外部 IP:</span>
          <span class="ml-2">{{ detail.externalIP?.join(', ') || '-' }}</span>
        </div>
        <div>
          <span class="text-sm text-gray-700 font-medium">年龄:</span>
          <span class="ml-2">{{ detail.age }}</span>
        </div>
      </div>

      <!-- 端口信息 -->
      <div v-if="detail.ports && detail.ports.length > 0">
        <h4 class="mb-2 text-sm text-gray-700 font-medium">端口:</h4>
        <ElTable :data="detail.ports" size="small">
          <ElTableColumn prop="name" label="名称" width="120" />
          <ElTableColumn prop="protocol" label="协议" width="80" />
          <ElTableColumn prop="port" label="端口" width="80" />
          <ElTableColumn prop="targetPort" label="目标端口" width="100" />
          <ElTableColumn prop="nodePort" label="NodePort" width="100" />
        </ElTable>
      </div>

      <!-- Selector -->
      <div v-if="detail.selector">
        <h4 class="mb-2 text-sm text-gray-700 font-medium">选择器:</h4>
        <div class="flex flex-wrap gap-2">
          <ElTag v-for="(value, key) in detail.selector" :key="key" type="info">{{ key }}: {{ value }}</ElTag>
        </div>
      </div>

      <!-- Endpoints -->
      <div v-if="detail.endpoints && detail.endpoints.length > 0">
        <h4 class="mb-2 text-sm text-gray-700 font-medium">后端端点:</h4>
        <ElTable :data="detail.endpoints" size="small">
          <ElTableColumn prop="ip" label="IP" width="140" />
          <ElTableColumn prop="hostname" label="主机名" width="180" />
          <ElTableColumn label="端口" min-width="200">
            <template #default="{ row }">
              <span v-if="row.ports">
                {{ row.ports.map((p: K8s.ServicePort) => `${p.port}/${p.protocol}`).join(', ') }}
              </span>
              <span v-else>-</span>
            </template>
          </ElTableColumn>
        </ElTable>
      </div>

      <!-- YAML Manifest -->
      <div>
        <h4 class="mb-2 text-sm text-gray-700 font-medium">YAML 配置:</h4>
        <div class="max-h-400 overflow-auto whitespace-pre rounded bg-gray-50 p-3 text-sm font-mono">
          {{ detail.manifest }}
        </div>
      </div>
    </div>
  </ElDialog>
</template>
