# K8s 资源详情页样式优化实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**目标:** 统一优化所有 Kubernetes 资源详情页的显示效果，通过提取通用组件保证一致性

**架构:** 先优化 Deployment 详情页作为参考模板，然后将相同模式应用到其他详情页，最后提取通用组件

**技术栈:** Vue 3 + TypeScript + Element Plus + Vite

## 全局约束

- 所有详情页必须保持视觉一致性
- 文案必须统一：年龄→创建时间，容器→容器组
- 标签显示格式：`key:value, key:value`
- 状态条件显示格式：`Type:Status`
- 镜像列多镜像用换行符显示
- 必须保持现有功能不受影响（YAML 编辑、终端、日志等）

---

## 阶段 1：Deployment 详情页优化（参考模板）

### Task 1: 添加数据格式化工具函数

**文件:**
- 创建: `soybean-admin-element-plus/src/utils/k8s-formatters.ts`

**接口:**
- 无依赖（纯工具函数）
- 提供: `formatLabels`, `formatConditions`, `formatImages` 函数

- [ ] **Step 1: 创建格式化工具文件**

```typescript
// soybean-admin-element-plus/src/utils/k8s-formatters.ts

/**
 * 格式化标签对象为字符串
 * 输入: {app: "nginx", env: "prod"}
 * 输出: "app:nginx, env:prod"
 */
export function formatLabels(labels: Record<string, string> | undefined): string {
  if (!labels || Object.keys(labels).length === 0) return '-';
  return Object.entries(labels)
    .map(([key, value]) => `${key}:${value}`)
    .join(', ');
}

/**
 * 格式化状态条件数组为字符串
 * 输入: [{type: "Available", status: "True"}, {type: "Progressing", status: "True"}]
 * 输出: "Available:True, Progressing:True"
 */
export function formatConditions(conditions: Array<{type: string, status: string}> | undefined): string {
  if (!conditions || conditions.length === 0) return '-';
  return conditions
    .map(c => `${c.type}:${c.status}`)
    .join(', ');
}

/**
 * 从 Pod 对象提取镜像列表（换行分隔）
 */
export function formatImages(pod: any): string {
  if (!pod?.containers || pod.containers.length === 0) return '-';
  return pod.containers
    .map((c: any) => c.image)
    .join('\n');
}

/**
 * 格式化副本数显示
 */
export function formatReplicas(ready: number | undefined, total: number | undefined): string {
  const r = ready ?? 0;
  const t = total ?? 0;
  return `${r}/${t}`;
}
```

- [ ] **Step 2: 运行类型检查验证**

Run: `cd soybean-admin-element-plus && pnpm typecheck`
Expected: PASS（无类型错误）

- [ ] **Step 3: 提交**

```bash
git add soybean-admin-element-plus/src/utils/k8s-formatters.ts
git commit -m "feat: add K8s data formatting utilities"
```

---

### Task 2: 优化 Deployment 详情页 - 基本信息

**文件:**
- 修改: `soybean-admin-element-plus/src/views/k8s/resources/deployments/detail.vue`

**接口:**
- 消费: `formatLabels`, `formatConditions`（来自 Task 1）
- 产生: 优化后的基本信息区域

- [ ] **Step 1: 在 script 部分导入格式化函数**

在第 19 行之后（`yaml` 导入后）添加：

```typescript
import { formatLabels, formatConditions } from '@/utils/k8s-formatters';
```

- [ ] **Step 2: 修改基本信息区域模板（第 384-468 行）**

将整个 `basic-info-section` 替换为：

```vue
    <!-- 基本信息区域 -->
    <div class="basic-info-section">
      <!-- 基本信息 -->
      <div class="info-section">
        <h3 class="section-title">基本信息</h3>
        <div class="desc-grid">
          <div class="desc-row">
            <div class="desc-item">
              <div class="item-label">名称</div>
              <div class="item-content">
                <span class="value-text">{{ deployment?.name || '-' }}</span>
              </div>
            </div>
            <div class="desc-item">
              <div class="item-label">命名空间</div>
              <div class="item-content">
                <span class="value-text">{{ deployment?.namespace || '-' }}</span>
              </div>
            </div>
            <div class="desc-item">
              <div class="item-label">创建时间</div>
              <div class="item-content">
                <span class="value-text">{{ deployment?.age || '-' }}</span>
              </div>
            </div>
          </div>
          <div class="desc-row">
            <div class="desc-item">
              <div class="item-label">副本数</div>
              <div class="item-content">
                <span class="value-text">{{ deployment?.ready || 0 }} / {{ deployment?.replicas || 0 }}</span>
              </div>
            </div>
            <div class="desc-item">
              <div class="item-label">可用副本</div>
              <div class="item-content">
                <span class="value-text">{{ deployment?.available || 0 }}</span>
              </div>
            </div>
            <div class="desc-item">
              <div class="item-label">最新副本</div>
              <div class="item-content">
                <span class="value-text">{{ deployment?.upToDate || 0 }}</span>
              </div>
            </div>
          </div>
          <div v-if="deployment?.labels" class="desc-row">
            <div class="desc-item desc-item-full">
              <div class="item-label">标签</div>
              <div class="item-content">
                <span class="value-text">{{ formatLabels(deployment.labels) }}</span>
              </div>
            </div>
          </div>
          <div v-if="deployment?.conditions && deployment.conditions.length > 0" class="desc-row">
            <div class="desc-item desc-item-full">
              <div class="item-label">状态条件</div>
              <div class="item-content">
                <span class="value-text">{{ formatConditions(deployment.conditions) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
```

- [ ] **Step 3: 运行 lint 检查**

Run: `cd soybean-admin-element-plus && pnpm lint`
Expected: PASS 或自动修复

- [ ] **Step 4: 运行类型检查**

Run: `cd soybean-admin-element-plus && pnpm typecheck`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add soybean-admin-element-plus/src/views/k8s/resources/deployments/detail.vue
git commit -m "feat(deployment): unify basic info layout, add formatted labels and conditions"
```

---

### Task 3: 优化 Deployment 详情页 - 容器组表格

**文件:**
- 修改: `soybean-admin-element-plus/src/views/k8s/resources/deployments/detail.vue`

**接口:**
- 消费: `formatImages`（来自 Task 1）
- 产生: 带镜像列的容器组表格

- [ ] **Step 1: 导入 formatImages 函数**

在第 19 行之后添加：

```typescript
import { formatLabels, formatConditions, formatImages } from '@/utils/k8s-formatters';
```

- [ ] **Step 2: 修改容器组表格（第 479-502 行）**

将表格定义替换为：

```vue
          <ElTable v-loading="podsLoading" :data="pods" stripe size="small">
            <ElTableColumn prop="name" label="Pod 名称" min-width="200" show-overflow-tooltip />
            <ElTableColumn label="状态" width="120">
              <template #default="{ row }">
                <ElTag :type="getPodStatusTag(row).type" size="small">
                  {{ getPodStatusTag(row).text }}
                </ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn label="镜像" min-width="200" show-overflow-tooltip>
              <template #default="{ row }">
                <span class="whitespace-pre-line">{{ formatImages(row) }}</span>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="ip" label="IP 地址" width="140" />
            <ElTableColumn prop="node" label="节点" width="150" show-overflow-tooltip />
            <ElTableColumn label="重启次数" width="100" align="center">
              <template #default="{ row }">
                {{ row.restarts || 0 }}
              </template>
            </ElTableColumn>
            <ElTableColumn label="创建时间" width="140">
              <template #default="{ row }">
                {{ row.age || '-' }}
              </template>
            </ElTableColumn>
            <ElTableColumn label="操作" width="150" fixed="right">
              <template #default="{ row }">
                <ElButton size="small" type="primary" link @click="handlePodTerminal(row)">终端</ElButton>
                <ElButton size="small" link @click="handlePodLogs(row)">日志</ElButton>
              </template>
            </ElTableColumn>
          </ElTable>
```

- [ ] **Step 3: 添加 CSS 样式支持镜像换行**

在 `<style scoped>` 部分（第 583 行后）添加：

```css
/* 镜像列换行显示 */
.whitespace-pre-line {
  white-space: pre-line;
  word-break: break-all;
}
```

- [ ] **Step 4: 运行 lint 检查**

Run: `cd soybean-admin-element-plus && pnpm lint`
Expected: PASS

- [ ] **Step 5: 运行类型检查**

Run: `cd soybean-admin-element-plus && pnpm typecheck`
Expected: PASS

- [ ] **Step 6: 提交**

```bash
git add soybean-admin-element-plus/src/views/k8s/resources/deployments/detail.vue
git commit -m "feat(deployment): add images column to pods table"
```

---

### Task 4: 移除 Deployment 详情页的独立镜像信息区域

**文件:**
- 修改: `soybean-admin-element-plus/src/views/k8s/resources/deployments/detail.vue`

**接口:**
- 无

- [ ] **Step 1: 删除镜像信息区域（第 433-446 行）**

删除以下代码块：

```vue
      <!-- 镜像信息 -->
      <div v-if="deployment?.images && deployment.images.length > 0" class="info-section">
        <h3 class="section-title">镜像信息</h3>
        <div class="desc-grid">
          <div class="desc-row">
            <div v-for="(image, index) in deployment.images" :key="index" class="desc-item desc-item-full">
              <div class="item-label">镜像 {{ index + 1 }}</div>
              <div class="item-content">
                <span class="value-text font-mono">{{ image }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
```

- [ ] **Step 2: 运行 lint 检查**

Run: `cd soybean-admin-element-plus && pnpm lint`
Expected: PASS

- [ ] **Step 3: 提交**

```bash
git add soybean-admin-element-plus/src/views/k8s/resources/deployments/detail.vue
git commit -m "refactor(deployment): remove standalone images section"
```

---

## 阶段 2：应用到其他详情页

### Task 5: 优化 StatefulSet 详情页

**文件:**
- 修改: `soybean-admin-element-plus/src/views/k8s/resources/statefulsets/detail.vue`

**接口:**
- 消费: `formatLabels`, `formatConditions`, `formatImages`

- [ ] **Step 1: 导入格式化函数**

在 script 部分添加导入（约第 5 行后）：

```typescript
import { formatLabels, formatConditions, formatImages } from '@/utils/k8s-formatters';
```

- [ ] **Step 2: 查看基本信息区域结构**

查看文件第 188-230 行左右，找到 `basic-info-section` 区域

- [ ] **Step 3: 重构基本信息区域**

参考 Task 2 的模式，将标签和状态条件整合到 desc-grid 布局中。查找类似以下结构的代码：
- 使用 ElTag 展示标签的部分
- 使用 ElTable 展示状态条件的部分

替换为统一的 desc-grid 布局

- [ ] **Step 4: 修改容器组表格添加镜像列**

查找 pods 表格定义（约在 240 行左右），参考 Task 3 的模式添加镜像列

- [ ] **Step 5: 修改文案"年龄"为"创建时间"**

将模板中所有 `年龄` 替换为 `创建时间`

- [ ] **Step 6: 运行 lint 和类型检查**

Run: `cd soybean-admin-element-plus && pnpm lint && pnpm typecheck`
Expected: PASS

- [ ] **Step 7: 提交**

```bash
git add soybean-admin-element-plus/src/views/k8s/resources/statefulsets/detail.vue
git commit -m "feat(statefulset): apply unified layout and add images column"
```

---

### Task 6: 优化 DaemonSet 详情页

**文件:**
- 修改: `soybean-admin-element-plus/src/views/k8s/resources/daemonsets/detail.vue`

**接口:**
- 消费: `formatLabels`, `formatConditions`, `formatImages`

- [ ] **Step 1: 导入格式化函数**

```typescript
import { formatLabels, formatConditions, formatImages } from '@/utils/k8s-formatters';
```

- [ ] **Step 2: 重构基本信息区域**

参考 Task 2 和 Task 5 的模式，统一基本信息布局

- [ ] **Step 3: 修改容器组表格添加镜像列**

参考 Task 3 的模式

- [ ] **Step 4: 修改文案**

将所有 `年龄` 替换为 `创建时间`，`容器` 替换为 `容器组`

- [ ] **Step 5: 运行检查**

Run: `cd soybean-admin-element-plus && pnpm lint && pnpm typecheck`
Expected: PASS

- [ ] **Step 6: 提交**

```bash
git add soybean-admin-element-plus/src/views/k8s/resources/daemonsets/detail.vue
git commit -m "feat(daemonset): apply unified layout and add images column"
```

---

### Task 7: 优化 Pod 详情页

**文件:**
- 修改: `soybean-admin-element-plus/src/views/k8s/resources/pods/detail.vue`

**接口:**
- 消费: `formatLabels`, `formatConditions`

- [ ] **Step 1: 导入格式化函数**

```typescript
import { formatLabels, formatConditions } from '@/utils/k8s-formatters';
```

- [ ] **Step 2: 查看详情页结构**

Pod 详情页略有不同，它直接显示 Pod 的容器信息

- [ ] **Step 3: 重构基本信息区域**

参考之前的模式，统一布局。注意 Pod 没有副本数概念

- [ ] **Step 4: 修改文案**

将 `年龄` 替换为 `创建时间`

- [ ] **Step 5: 运行检查**

Run: `cd soybean-admin-element-plus && pnpm lint && pnpm typecheck`
Expected: PASS

- [ ] **Step 6: 提交**

```bash
git add soybean-admin-element-plus/src/views/k8s/resources/pods/detail.vue
git commit -m "feat(pod): apply unified layout"
```

---

### Task 8: 优化 Job 详情页

**文件:**
- 修改: `soybean-admin-element-plus/src/views/k8s/resources/jobs/detail.vue`

**接口:**
- 消费: `formatLabels`, `formatConditions`

- [ ] **Step 1: 导入格式化函数**

```typescript
import { formatLabels, formatConditions } from '@/utils/k8s-formatters';
```

- [ ] **Step 2: 重构基本信息区域**

参考之前的模式

- [ ] **Step 3: 修改文案**

将 `年龄` 替换为 `创建时间`，`容器` 替换为 `容器组`

- [ ] **Step 4: 运行检查**

Run: `cd soybean-admin-element-plus && pnpm lint && pnpm typecheck`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add soybean-admin-element-plus/src/views/k8s/resources/jobs/detail.vue
git commit -m "feat(job): apply unified layout"
```

---

### Task 9: 优化 CronJob 详情页

**文件:**
- 修改: `soybean-admin-element-plus/src/views/k8s/resources/cronjobs/detail.vue`

**接口:**
- 消费: `formatLabels`, `formatConditions`

- [ ] **Step 1: 导入格式化函数**

```typescript
import { formatLabels, formatConditions } from '@/utils/k8s-formatters';
```

- [ ] **Step 2: 重构基本信息区域**

参考之前的模式

- [ ] **Step 3: 修改文案**

将 `年龄` 替换为 `创建时间`，`容器` 替换为 `容器组`

- [ ] **Step 4: 运行检查**

Run: `cd soybean-admin-element-plus && pnpm lint && pnpm typecheck`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add soybean-admin-element-plus/src/views/k8s/resources/cronjobs/detail.vue
git commit -m "feat(cronjob): apply unified layout"
```

---

## 阶段 3：提取通用组件

### Task 10: 创建基本信息网格组件

**文件:**
- 创建: `soybean-admin-element-plus/src/components/k8s/K8sBasicInfoGrid.vue`

**接口:**
- 消费: 格式化工具函数（可选）
- 提供: 可复用的基本信息网格组件

- [ ] **Step 1: 创建组件文件**

```vue
<!-- soybean-admin-element-plus/src/components/k8s/K8sBasicInfoGrid.vue -->
<script setup lang="ts">
interface Field {
  label: string;
  value: string | number;
  fullRow?: boolean;  // 是否占满一行
}

defineProps<{
  fields: Field[][];
}>();
</script>

<template>
  <div class="desc-grid">
    <div v-for="(row, rowIndex) in fields" :key="rowIndex" class="desc-row">
      <div
        v-for="(field, fieldIndex) in row"
        :key="fieldIndex"
        :class="['desc-item', { 'desc-item-full': field.fullRow }]"
      >
        <div class="item-label">{{ field.label }}</div>
        <div class="item-content">
          <span class="value-text">{{ field.value }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
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
</style>
```

- [ ] **Step 2: 运行类型检查**

Run: `cd soybean-admin-element-plus && pnpm typecheck`
Expected: PASS

- [ ] **Step 3: 提交**

```bash
git add soybean-admin-element-plus/src/components/k8s/K8sBasicInfoGrid.vue
git commit -m "feat: create K8sBasicInfoGrid component"
```

---

### Task 11: 创建容器组表格组件

**文件:**
- 创建: `soybean-admin-element-plus/src/components/k8s/K8sPodsTable.vue`

**接口:**
- 提供: 可复用的容器组表格组件

- [ ] **Step 1: 创建组件文件**

```vue
<!-- soybean-admin-element-plus/src/components/k8s/K8sPodsTable.vue -->
<script setup lang="ts">
import { ElButton, ElTag, ElTable, ElTableColumn } from 'element-plus';
import { formatImages } from '@/utils/k8s-formatters';

interface Pod {
  name: string;
  phase?: string;
  containers?: Array<{ name: string; image: string }>;
  ip?: string;
  node?: string;
  restarts?: number;
  age?: string;
}

defineProps<{
  pods: Pod[];
  loading?: boolean;
}>();

const emit = defineEmits<{
  terminal: [pod: Pod];
  logs: [pod: Pod];
}>();

const getPodStatusTag = (pod: Pod) => {
  const phase = pod.phase || 'Unknown';
  switch (phase) {
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
</script>

<template>
  <ElTable :data="pods" :loading="loading" stripe size="small">
    <ElTableColumn prop="name" label="Pod 名称" min-width="200" show-overflow-tooltip />
    <ElTableColumn label="状态" width="120">
      <template #default="{ row }">
        <ElTag :type="getPodStatusTag(row).type" size="small">
          {{ getPodStatusTag(row).text }}
        </ElTag>
      </template>
    </ElTableColumn>
    <ElTableColumn label="镜像" min-width="200" show-overflow-tooltip>
      <template #default="{ row }">
        <span class="whitespace-pre-line">{{ formatImages(row) }}</span>
      </template>
    </ElTableColumn>
    <ElTableColumn prop="ip" label="IP 地址" width="140" />
    <ElTableColumn prop="node" label="节点" width="150" show-overflow-tooltip />
    <ElTableColumn label="重启次数" width="100" align="center">
      <template #default="{ row }">
        {{ row.restarts || 0 }}
      </template>
    </ElTableColumn>
    <ElTableColumn label="创建时间" width="140">
      <template #default="{ row }">
        {{ row.age || '-' }}
      </template>
    </ElTableColumn>
    <ElTableColumn label="操作" width="150" fixed="right">
      <template #default="{ row }">
        <ElButton size="small" type="primary" link @click="emit('terminal', row)">终端</ElButton>
        <ElButton size="small" link @click="emit('logs', row)">日志</ElButton>
      </template>
    </ElTableColumn>
  </ElTable>
</template>

<style scoped>
.whitespace-pre-line {
  white-space: pre-line;
  word-break: break-all;
}
</style>
```

- [ ] **Step 2: 运行类型检查**

Run: `cd soybean-admin-element-plus && pnpm typecheck`
Expected: PASS

- [ ] **Step 3: 提交**

```bash
git add soybean-admin-element-plus/src/components/k8s/K8sPodsTable.vue
git commit -m "feat: create K8sPodsTable component"
```

---

### Task 12: 创建事件表格组件

**文件:**
- 创建: `soybean-admin-element-plus/src/components/k8s/K8sEventsTable.vue`

**接口:**
- 提供: 可复用的事件表格组件

- [ ] **Step 1: 创建组件文件**

```vue
<!-- soybean-admin-element-plus/src/components/k8s/K8sEventsTable.vue -->
<script setup lang="ts">
import { ElTag, ElTable, ElTableColumn } from 'element-plus';

interface Event {
  type: string;
  reason: string;
  message: string;
  source?: string;
  count?: number;
  lastTimestamp?: string;
}

defineProps<{
  events: Event[];
  loading?: boolean;
}>();
</script>

<template>
  <ElTable :data="events" :loading="loading" stripe size="small">
    <ElTableColumn prop="type" label="类型" width="120">
      <template #default="{ row }">
        <ElTag :type="row.type === 'Normal' ? 'success' : 'warning'" size="small">
          {{ row.type }}
        </ElTag>
      </template>
    </ElTableColumn>
    <ElTableColumn prop="reason" label="原因" width="150" show-overflow-tooltip />
    <ElTableColumn prop="message" label="消息" min-width="300" show-overflow-tooltip />
    <ElTableColumn prop="source" label="来源" width="150" show-overflow-tooltip />
    <ElTableColumn prop="count" label="次数" width="80" align="center" />
    <ElTableColumn prop="lastTimestamp" label="最后时间" width="160" />
  </ElTable>
</template>
```

- [ ] **Step 2: 运行类型检查**

Run: `cd soybean-admin-element-plus && pnpm typecheck`
Expected: PASS

- [ ] **Step 3: 提交**

```bash
git add soybean-admin-element-plus/src/components/k8s/K8sEventsTable.vue
git commit -m "feat: create K8sEventsTable component"
```

---

### Task 13: 重构 Deployment 详情页使用通用组件

**文件:**
- 修改: `soybean-admin-element-plus/src/views/k8s/resources/deployments/detail.vue`

**接口:**
- 消费: `K8sBasicInfoGrid`, `K8sPodsTable`, `K8sEventsTable`

- [ ] **Step 1: 导入通用组件**

在 script 部分添加：

```typescript
import K8sBasicInfoGrid from '@/components/k8s/K8sBasicInfoGrid.vue';
import K8sPodsTable from '@/components/k8s/K8sPodsTable.vue';
import K8sEventsTable from '@/components/k8s/K8sEventsTable.vue';
```

- [ ] **Step 2: 添加计算属性生成基本信息字段**

在 script 部分添加：

```typescript
const basicInfoFields = computed(() => {
  if (!deployment.value) return [[]];

  return [
    [
      { label: '名称', value: deployment.value.name || '-' },
      { label: '命名空间', value: deployment.value.namespace || '-' },
      { label: '创建时间', value: deployment.value.age || '-' }
    ],
    [
      { label: '副本数', value: `${deployment.value.ready || 0} / ${deployment.value.replicas || 0}` },
      { label: '可用副本', value: deployment.value.available?.toString() || '0' },
      { label: '最新副本', value: deployment.value.upToDate?.toString() || '0' }
    ],
    deployment.value.labels ? [
      { label: '标签', value: formatLabels(deployment.value.labels), fullRow: true }
    ] : [],
    deployment.value.conditions?.length ? [
      { label: '状态条件', value: formatConditions(deployment.value.conditions), fullRow: true }
    ] : []
  ].filter(row => row.length > 0);
});
```

- [ ] **Step 3: 替换基本信息模板**

将 `basic-info-section` 替换为：

```vue
    <!-- 基本信息区域 -->
    <div class="basic-info-section">
      <div class="info-section">
        <h3 class="section-title">基本信息</h3>
        <K8sBasicInfoGrid :fields="basicInfoFields" />
      </div>
    </div>
```

- [ ] **Step 4: 替换容器组表格**

将 pods 表格替换为：

```vue
          <div class="section-header">
            <span class="section-title">共 {{ pods.length }} 个容器组</span>
            <ElButton size="small" @click="handleRefreshPods">刷新</ElButton>
          </div>
          <K8sPodsTable
            :pods="pods"
            :loading="podsLoading"
            @terminal="handlePodTerminal"
            @logs="handlePodLogs"
          />
```

- [ ] **Step 5: 替换事件表格**

将 events 表格替换为：

```vue
          <div class="section-header">
            <span class="section-title">共 {{ events.length }} 个事件</span>
            <ElButton size="small" @click="loadEvents">刷新</ElButton>
          </div>
          <K8sEventsTable :events="events" :loading="eventsLoading" />
```

- [ ] **Step 6: 运行检查**

Run: `cd soybean-admin-element-plus && pnpm lint && pnpm typecheck`
Expected: PASS

- [ ] **Step 7: 提交**

```bash
git add soybean-admin-element-plus/src/views/k8s/resources/deployments/detail.vue
git commit -m "refactor(deployment): use common components"
```

---

### Task 14: 重构其他详情页使用通用组件

**文件:**
- 修改: `soybean-admin-element-plus/src/views/k8s/resources/statefulsets/detail.vue`
- 修改: `soybean-admin-element-plus/src/views/k8s/resources/daemonsets/detail.vue`
- 修改: `soybean-admin-element-plus/src/views/k8s/resources/pods/detail.vue`

**接口:**
- 消费: `K8sBasicInfoGrid`, `K8sPodsTable`, `K8sEventsTable`

- [ ] **Step 1: 重构 StatefulSet 详情页**

参考 Task 13 的模式：
1. 导入通用组件
2. 添加 basicInfoFields 计算属性
3. 替换模板

- [ ] **Step 2: 重构 DaemonSet 详情页**

参考 Task 13 的模式

- [ ] **Step 3: 重构 Pod 详情页**

Pod 详情页不使用 K8sPodsTable（本身就是 Pod）

- [ ] **Step 4: 运行检查**

Run: `cd soybean-admin-element-plus && pnpm lint && pnpm typecheck`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add soybean-admin-element-plus/src/views/k8s/resources/statefulsets/detail.vue
git add soybean-admin-element-plus/src/views/k8s/resources/daemonsets/detail.vue
git add soybean-admin-element-plus/src/views/k8s/resources/pods/detail.vue
git commit -m "refactor: use common components for all detail pages"
```

---

### Task 15: 最终验证和清理

**文件:**
- 所有详情页文件

- [ ] **Step 1: 运行完整 lint 检查**

Run: `cd soybean-admin-element-plus && pnpm lint`
Expected: PASS，无警告

- [ ] **Step 2: 运行完整类型检查**

Run: `cd soybean-admin-element-plus && pnpm typecheck`
Expected: PASS

- [ ] **Step 3: 验证前端构建**

Run: `cd soybean-admin-element-plus && pnpm build`
Expected: 构建成功

- [ ] **Step 4: 检查是否有遗留的旧代码**

grep 搜索以下模式，确保已全部更新：
- `年龄`（应该已全部改为"创建时间"）
- `ElTag.*label.*标签`（基本信息中的标签应该已改为网格显示）

- [ ] **Step 5: 提交最终版本**

```bash
git add -A
git commit -m "chore: final cleanup and validation"
```

---

## 验收标准

完成所有任务后，验证以下内容：

- [ ] 所有详情页基本信息使用统一网格布局
- [ ] 标签显示为 `key:value, key:value` 格式
- [ ] 状态条件显示为 `Type:Status` 格式
- [ ] 所有"年龄"文案已改为"创建时间"
- [ ] 所有"容器"文案已改为"容器组"
- [ ] 容器组表格显示镜像列
- [ ] 多镜像正确换行显示
- [ ] YAML 编辑功能正常
- [ ] 终端功能正常
- [ ] 日志功能正常
- [ ] 无 lint 警告
- [ ] 无类型错误
- [ ] 前端构建成功
