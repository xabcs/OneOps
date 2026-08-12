<script setup lang="ts">
  import { computed } from 'vue';

  defineOptions({ name: 'SecretDetailDialog' });

  const props = defineProps<{ detail: K8s.Secret | null }>();

  const visible = defineModel<boolean>('visible', { default: false });

  const title = computed(() => `Secret: ${props.detail?.name ?? ''}`);
</script>

<template>
  <ElDialog v-model="visible" :title="title" width="800px">
    <div v-if="detail" class="space-y-4">
      <!-- 安全警告 -->
      <div class="flex items-start gap-2 border border-amber-200 rounded bg-amber-50 p-3">
        <svg class="mt-0.5 h-5 w-5 text-amber-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
          />
        </svg>
        <div class="text-sm text-amber-800">
          <strong>敏感数据保护：</strong>
          为保护敏感信息，Secret 的实际数据值已被隐藏（显示为 ******）。如需查看或编辑，请直接使用 kubectl 命令。
        </div>
      </div>

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

      <!-- Data Keys (隐藏实际值) -->
      <div v-if="detail.data && Object.keys(detail.data).length > 0">
        <h4 class="mb-2 text-sm text-gray-700 font-medium">数据键:</h4>
        <div class="space-y-2">
          <div v-for="(_value, key) in detail.data" :key="key" class="border rounded p-2">
            <div class="mb-1 text-sm text-gray-700 font-medium">{{ key }}</div>
            <div class="rounded bg-gray-50 p-2 text-sm text-amber-600 font-mono">****** (敏感数据已隐藏)</div>
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
