<script setup lang="ts">
  import { computed } from 'vue';

  defineOptions({ name: 'ConfigMapDetailDialog' });

  const props = defineProps<{ detail: K8s.ConfigMap | null }>();

  const visible = defineModel<boolean>('visible', { default: false });

  const title = computed(() => `ConfigMap: ${props.detail?.name ?? ''}`);
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
          <span class="text-sm text-gray-700 font-medium">年龄:</span>
          <span class="ml-2">{{ detail.age }}</span>
        </div>
      </div>

      <!-- Labels -->
      <div v-if="detail.labels && Object.keys(detail.labels).length > 0">
        <h4 class="mb-2 text-sm text-gray-700 font-medium">标签:</h4>
        <div class="flex flex-wrap gap-2">
          <ElTag v-for="(value, key) in detail.labels" :key="key" type="info">{{ key }}: {{ value }}</ElTag>
        </div>
      </div>

      <!-- Data -->
      <div v-if="detail.data && Object.keys(detail.data).length > 0">
        <h4 class="mb-2 text-sm text-gray-700 font-medium">数据:</h4>
        <div class="space-y-2">
          <div v-for="(value, key) in detail.data" :key="key" class="border rounded p-2">
            <div class="mb-1 text-sm text-gray-700 font-medium">{{ key }}</div>
            <div class="whitespace-pre-wrap break-all rounded bg-gray-50 p-2 text-sm font-mono">
              {{ value }}
            </div>
          </div>
        </div>
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
