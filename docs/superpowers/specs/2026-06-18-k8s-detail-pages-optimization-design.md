# K8s 资源详情页样式优化设计文档

**日期**: 2026-06-18
**状态**: 待实施
**优先级**: P1

## 概述

统一优化所有 Kubernetes 资源详情页的显示效果，通过提取通用组件保证页面布局一致性，提升用户体验和代码可维护性。

## 问题分析

### 当前问题

1. **基本信息布局不统一**
   - 标签（Labels）使用独立的 ElTag 组件展示
   - 状态条件使用独立的 ElTable 展示
   - 与基本信息（名称、命名空间、年龄）的 desc-grid 布局不一致

2. **文案不统一**
   - "年龄"应该改为"创建时间"
   - "容器"应该改为"容器组"

3. **镜像信息位置不当**
   - 镜像信息单独显示在基本信息区域
   - 应该显示在容器组表格中，方便查看每个 Pod 的镜像

4. **代码重复严重**
   - 每个详情页都重复相同的布局代码
   - 修改样式需要同步多个文件

### 受影响的页面

- Deployments（无状态）
- StatefulSets（有状态）
- DaemonSets（守护进程集）
- Pods（容器组）
- Jobs（任务）
- CronJobs（定时任务）

## 设计方案

### 1. 基本信息区域重构

**目标**: 统一使用 desc-grid 网格布局展示所有基本信息

**当前结构**:
```
基本信息区域
├── 基本信息部分（desc-grid）: 名称、命名空间、年龄、副本数等
├── 镜像信息部分（独立 section）
├── 标签部分（独立 section，使用 ElTag）
└── 状态条件部分（独立 section，使用 ElTable）
```

**优化后结构**:
```
基本信息区域（basic-info-section）
└── desc-grid
    ├── desc-row: 名称、命名空间、创建时间
    ├── desc-row: 副本数、可用副本、最新副本（仅 workloads）
    └── desc-row: 标签列表、状态条件
```

**数据格式化规则**:

| 类型 | 输入 | 输出 |
|------|------|------|
| 标签 | `{app: nginx, env: prod}` | `app:nginx, env:prod` |
| 状态条件 | `[{type: "Available", status: "True"}]` | `Available:True` |

### 2. 文案统一

| 位置 | 原文案 | 新文案 |
|------|--------|--------|
| 基本信息区域 | 年龄 | 创建时间 |
| 所有详情页 | 容器 | 容器组 |
| 容器组表格 | 年龄 | 创建时间 |

### 3. 容器组表格优化

**新增镜像列**:

| 列名 | 宽度 | 说明 |
|------|------|------|
| Pod 名称 | 200px | show-overflow-tooltip |
| 状态 | 120px | ElTag 显示 |
| 镜像 | auto | 多个镜像换行显示 |
| IP 地址 | 140px | - |
| 节点 | 150px | show-overflow-tooltip |
| 重启次数 | 100px | align-center |
| 创建时间 | 140px | 原"年龄"列 |
| 操作 | 150px | 终端、日志按钮 |

**镜像列格式**:
```
nginx:1.21-alpine
redis:6.2.6
```

### 4. 通用组件提取

#### 组件结构

```
soybean-admin-element-plus/src/components/k8s/
├── detail/
│   ├── K8sDetailPage.vue       # 详情页主框架
│   ├── K8sBasicInfoGrid.vue     # 基本信息网格
│   ├── K8sPodsTable.vue         # 容器组表格
│   └── K8sEventsTable.vue       # 事件表格
└── common/
    └── K8sStatusTag.vue         # 状态标签组件（复用现有）
```

#### 组件职责

**K8sDetailPage.vue**
- 顶部操作栏（返回箭头、资源名称、操作按钮）
- 标签页框架（基本信息、容器组、事件、YAML）
- 数据加载和刷新逻辑

**K8sBasicInfoGrid.vue**
- 接收字段配置和数据
- 统一的 desc-grid 布局渲染
- 支持文本、标签、状态条件等格式

**K8sPodsTable.vue**
- 容器组列表表格
- 镜像列显示
- 操作按钮（终端、日志）

**K8sEventsTable.vue**
- 事件列表表格
- 类型标签显示

#### Props 设计

```typescript
// K8sDetailPage.vue
interface K8sDetailPageProps {
  resourceType: ResourceType;
  resource: any;
  clusterId: number;
  actions?: ActionButton[];
  showPodsTab?: boolean;      // 是否显示容器组标签页
  showEventsTab?: boolean;    // 是否显示事件标签页
}

// K8sBasicInfoGrid.vue
interface BasicInfoField {
  label: string;
  key: string;
  type?: 'text' | 'tags' | 'conditions';
  format?: (value: any) => string;
}

interface K8sBasicInfoGridProps {
  fields: BasicInfoField[];
  data: any;
}

// K8sPodsTable.vue
interface K8sPodsTableProps {
  pods: any[];
  loading?: boolean;
  onTerminal?: (pod: any) => void;
  onLogs?: (pod: any) => void;
}
```

### 5. 迁移策略

#### 阶段 1: Deployment 详情页优化（参考模板）

1. 重构基本信息区域
2. 添加镜像列到容器组表格
3. 修改文案（年龄 → 创建时间）
4. 验证效果

#### 阶段 2: 应用到其他详情页

按优先级顺序：
1. StatefulSet（结构与 Deployment 相似）
2. DaemonSet（结构与 Deployment 相似）
3. Pod（容器组本身，略有不同）
4. Job / CronJob

#### 阶段 3: 提取通用组件

1. 创建通用组件文件
2. 迁移 Deployment 到通用组件
3. 逐个迁移其他详情页
4. 删除重复代码

## 技术实现要点

### 1. 数据格式化函数

```typescript
// 标签格式化
function formatLabels(labels: Record<string, string>): string {
  if (!labels) return '-';
  return Object.entries(labels)
    .map(([k, v]) => `${k}:${v}`)
    .join(', ');
}

// 状态条件格式化
function formatConditions(conditions: Array<{type: string, status: string}>): string {
  if (!conditions?.length) return '-';
  return conditions
    .map(c => `${c.type}:${c.status}`)
    .join(', ');
}

// 镜像列表格式化
function formatImages(pod: any): string {
  if (!pod?.containers?.length) return '-';
  return pod.containers
    .map(c => c.image)
    .join('\n');
}
```

### 2. 字段配置示例

```typescript
// Deployment 基本信息
const deploymentBasicFields: BasicInfoField[] = [
  { label: '名称', key: 'name', type: 'text' },
  { label: '命名空间', key: 'namespace', type: 'text' },
  { label: '创建时间', key: 'age', type: 'text' },
  { label: '副本数', key: 'replicas', type: 'text', format: (v, d) => `${d.ready || 0}/${v || 0}` },
  { label: '标签', key: 'labels', type: 'tags' },
  { label: '状态条件', key: 'conditions', type: 'conditions' }
];
```

### 3. 样式复用

通用组件复用现有的 scoped 样式：
- `.detail-page`
- `.instance-bar`
- `.basic-info-section`
- `.desc-grid`
- `.desc-row`
- `.desc-item`

## 验收标准

### 功能验收

- [ ] 所有详情页基本信息使用统一网格布局
- [ ] 文案"年龄"全部改为"创建时间"
- [ ] 文案"容器"全部改为"容器组"
- [ ] 容器组表格显示镜像列
- [ ] 多镜像正确换行显示
- [ ] 所有资源详情页效果一致

### 代码质量验收

- [ ] 通用组件通过 TypeScript 类型检查
- [ ] 无 ESLint 警告
- [ ] 代码重复率降低
- [ ] 组件 Props 有完整的类型定义

### 兼容性验收

- [ ] 不影响现有的 YAML 编辑功能
- [ ] 不影响现有的终端和日志功能
- [ ] 不影响现有的缩放、重启等操作

## 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 样式冲突 | 中 | 通用组件使用 scoped 样式，逐个验证 |
| 数据格式不兼容 | 低 | 添加默认值和容错处理 |
| 功能遗漏 | 低 | 对照现有功能清单逐项验证 |

## 后续优化

此设计完成后，可以考虑：

1. 提取列表页的通用组件（表格、筛选器、分页）
2. 统一 YAML 编辑器的实现
3. 添加详情页的面包屑导航
