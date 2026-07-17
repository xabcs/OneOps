/**
 * 组件统一导出
 * 按原子设计组织：atoms -> molecules -> organisms
 */

// 原子组件
export { default as AButton } from './atoms/AButton.vue';
export { default as AInput } from './atoms/AInput.vue';

// 分子组件
export { default as ConfirmDialog } from './molecules/ConfirmDialog.vue';
export { default as DataTable } from './molecules/DataTable.vue';
export { default as FilterBar } from './molecules/FilterBar.vue';
export { default as SearchBar } from './molecules/SearchBar.vue';
export { default as StatusBadge } from './molecules/StatusBadge.vue';

// 生物组件
export { default as InfoCard } from './organisms/InfoCard.vue';
export { default as ListView } from './organisms/ListView.vue';
export { default as PageHeader } from './organisms/PageHeader.vue';

// 类型导出
export type { AButtonProps } from './atoms/AButton.vue';
export type { AInputProps } from './atoms/AInput.vue';
export type { ConfirmDialogProps } from './molecules/ConfirmDialog.vue';
export type { DataTableProps } from './molecules/DataTable.vue';
export type { FilterBarProps } from './molecules/FilterBar.vue';
export type { SearchBarProps } from './molecules/SearchBar.vue';
export type { StatusBadgeProps } from './molecules/StatusBadge.vue';
export type { InfoCardProps } from './organisms/InfoCard.vue';
export type { ListViewProps } from './organisms/ListView.vue';
export type { PageHeaderProps } from './organisms/PageHeader.vue';
