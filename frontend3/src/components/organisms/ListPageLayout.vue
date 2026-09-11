<script setup lang="ts">
  import { computed, ref, useSlots } from 'vue';
  import type { PaginationProps } from 'element-plus';
  import { Refresh, Search } from '@element-plus/icons-vue';

  /**
   * 列表页统一布局骨架
   *
   * 结构：#hero（可选静态卡片）→ 搜索区（独立卡片，或 embedded-search 内嵌表格卡片顶部）→ 表格卡片（header：标题行 + 可选 #tabs 行 + #toolbar；body：内嵌筛选行 + 表格区 + 内置统一分页）
   *
   * 高度模型：复用 global.css 的 .table-page 体系 —— 整页不滚、表格区撑满剩余高度内部滚动、分页常驻底部。
   *
   * 用法（标准模式）：
   * ```vue
   * <ListPageLayout
   *   title="用户列表"
   *   description="..."
   *   :pagination="mobilePagination"
   *   @search="handleSearch"
   *   @reset="handleReset"
   * >
   *   <template #hero>...</template>
   *   <template #search>筛选控件...</template>
   *   <template #toolbar>新增/批量操作按钮...</template>
   *   <ElTable height="100%" ...>列...</ElTable>
   * </ListPageLayout>
   * ```
   *
   * tab 页签变体（如工单中心）：#tabs 渲染在表格卡片 header 的标题行下方，与内容同卡片联动；
   * 配合 embedded-search 将筛选行内嵌表格卡片顶部，整页一张卡：
   * ```vue
   * <ListPageLayout title="工单中心" description="..." embedded-search :pagination="mobilePagination">
   *   <template #tabs><ElTabs>...</ElTabs></template>
   *   <template #search>筛选控件...</template>
   *   <template #toolbar>主操作按钮（渲染在标题行右侧）...</template>
   *   <ElTable height="100%" ...>列...</ElTable>
   * </ListPageLayout>
   * ```
   */
  defineOptions({ name: 'ListPageLayout' });

  /** 分页对象：useUIPaginatedTable 返回的 pagination / mobilePagination（内含 current-change、size-change 处理器） */
  type LayoutPagination = Partial<PaginationProps> &
    Partial<{
      'current-change': (page: number) => void;
      'size-change': (size: number) => void;
    }>;

  interface Props {
    /** 表格卡片标题 */
    title?: string;
    /** 表格卡片描述（副标题） */
    description?: string;
    /** 分页配置；不传则不渲染分页（本地表格场景） */
    pagination?: LayoutPagination | null;
    /** 分页 layout，全站统一约定 */
    paginationLayout?: string;
    /** 搜索区是否可折叠（收起后仅保留标题与展开按钮） */
    collapsible?: boolean;
    /** 搜索区内嵌到表格卡片顶部（与 tabs/表格同卡片联动，不再独立成卡） */
    embeddedSearch?: boolean;
  }

  const props = withDefaults(defineProps<Props>(), {
    title: '',
    description: '',
    pagination: null,
    paginationLayout: 'total, sizes, prev, pager, next, jumper',
    collapsible: false,
    embeddedSearch: false
  });

  interface Emits {
    (e: 'search'): void;
    (e: 'reset'): void;
  }

  const emit = defineEmits<Emits>();

  const slots = useSlots();
  const hasSearchSlot = computed(() => Boolean(slots.search));
  const searchCollapsed = ref(false);
</script>

<template>
  <div class="table-page">
    <!-- Hero / 统计等静态卡片区（可选） -->
    <slot name="hero" />

    <!-- 搜索区（独立卡片模式）：重置/查询按钮固定在右侧（普通卡片样式，与表格卡片视觉一致） -->
    <ElCard v-if="hasSearchSlot && !embeddedSearch" shadow="hover" class="card-static">
      <ElSpace wrap class="w-full" align="center">
        <span class="whitespace-nowrap text-14px font-semibold">搜索筛选</span>
        <template v-if="!searchCollapsed">
          <slot name="search" />
        </template>
        <div class="ml-auto flex items-center gap-8px">
          <template v-if="!searchCollapsed">
            <ElButton @click="emit('reset')">
              <template #icon>
                <ElIcon><Refresh /></ElIcon>
              </template>
              重置
            </ElButton>
            <ElButton type="primary" @click="emit('search')">
              <template #icon>
                <ElIcon><Search /></ElIcon>
              </template>
              查询
            </ElButton>
          </template>
          <ElButton v-if="collapsible" link type="primary" @click="searchCollapsed = !searchCollapsed">
            {{ searchCollapsed ? '展开' : '收起' }}
          </ElButton>
        </div>
      </ElSpace>
    </ElCard>

    <!-- 表格卡片：撑满剩余高度 -->
    <ElCard
      shadow="hover"
      :class="slots.tabs ? ['card-tabs', { 'card-tabs-only': !title && !description && !slots.toolbar }] : []"
    >
      <template v-if="title || description || slots.toolbar || slots.tabs" #header>
        <div class="flex flex-col gap-8px">
          <!-- 标题行：与其他列表页同构（标题/描述 + 右侧主操作按钮） -->
          <div v-if="title || description || slots.toolbar" class="flex items-center justify-between">
            <div class="flex flex-col gap-4px">
              <span v-if="title" class="text-16px font-bold">{{ title }}</span>
              <span v-if="description" class="text-13px color-[var(--el-text-color-secondary)]">{{ description }}</span>
            </div>
            <ElSpace v-if="slots.toolbar" class="ml-16px shrink-0">
              <slot name="toolbar" />
            </ElSpace>
          </div>
          <!-- tabs 行：页签贴 header 底边，与表格内容同卡片联动 -->
          <div v-if="slots.tabs" class="min-w-0">
            <slot name="tabs" />
          </div>
        </div>
      </template>

      <!-- 搜索区（内嵌模式）：渲染在表格卡片顶部，与 tabs/表格/分页同容器联动 -->
      <ElSpace v-if="hasSearchSlot && embeddedSearch" wrap class="mb-12px w-full" align="center">
        <template v-if="!searchCollapsed">
          <slot name="search" />
        </template>
        <div class="ml-auto flex items-center gap-8px">
          <template v-if="!searchCollapsed">
            <ElButton @click="emit('reset')">
              <template #icon>
                <ElIcon><Refresh /></ElIcon>
              </template>
              重置
            </ElButton>
            <ElButton type="primary" @click="emit('search')">
              <template #icon>
                <ElIcon><Search /></ElIcon>
              </template>
              查询
            </ElButton>
          </template>
          <ElButton v-if="collapsible" link type="primary" @click="searchCollapsed = !searchCollapsed">
            {{ searchCollapsed ? '展开' : '收起' }}
          </ElButton>
        </div>
      </ElSpace>

      <!-- 表格区：内部放 <ElTable height="100%">，表体滚动、表头固定 -->
      <div class="table-scroll-wrap">
        <slot />
      </div>

      <!-- 分页：统一 layout，常驻底部 -->
      <div v-if="pagination" class="mt-16px flex justify-end">
        <ElPagination
          v-if="pagination.total"
          :layout="paginationLayout"
          v-bind="pagination"
          @current-change="pagination['current-change']?.($event)"
          @size-change="pagination['size-change']?.($event)"
        />
      </div>
    </ElCard>
  </div>
</template>

<style lang="scss" scoped>
  /* tabs 变体：header 底边去 padding 与边框，页签导航线充当卡片分隔线，页签与表格内容贴合联动；顶部保留正常标题行 padding */
  .card-tabs :deep(.el-card__header) {
    padding: calc(var(--el-card-padding) - 2px) var(--el-card-padding) 0;
    border-bottom: none;
  }

  /* tabs 独占 header（无标题行）时贴顶 */
  .card-tabs-only :deep(.el-card__header) {
    padding-top: 0;
  }

  .card-tabs :deep(.el-tabs__header) {
    margin: 0;
  }
</style>
