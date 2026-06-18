<script setup lang="ts">
import { ElTag } from 'element-plus';

interface Field {
  label: string;
  value: string | number | Array<{ key: string; value: string; type?: string }>;
  fullRow?: boolean;
  isTags?: boolean;
  isConditions?: boolean;
}

defineProps<{
  fields: Field[][];
}>();

// 获取标签颜色
const getTagType = (key: string) => {
  const keyLower = key.toLowerCase();
  if (keyLower.includes('app') || keyLower.includes('name')) return 'primary';
  if (keyLower.includes('env') || keyLower.includes('environment')) return 'success';
  if (keyLower.includes('version') || keyLower.includes('ver')) return 'warning';
  if (keyLower.includes('component')) return 'info';
  return 'default';
};

// 获取状态条件颜色
const getConditionType = (status: string) => {
  if (status === 'True') return 'success';
  if (status === 'False') return 'danger';
  if (status === 'Unknown') return 'info';
  return 'warning';
};
</script>

<template>
  <div class="desc-grid">
    <div v-for="(row, rowIndex) in fields" :key="rowIndex" class="desc-row">
      <div
        v-for="(field, fieldIndex) in row"
        :key="fieldIndex"
        :class="['desc-item', { 'desc-item-full': field.fullRow }]"
      >
        <div class="item-content-inline">
          <span class="item-label">{{ field.label }}:</span>
          <!-- 标签列表 -->
          <template v-if="field.isTags && Array.isArray(field.value)">
            <span class="tags-list">
              <ElTag
                v-for="(tag, idx) in field.value"
                :key="idx"
                :type="getTagType(tag.key)"
                size="small"
                class="tag-item"
              >
                {{ tag.key }}: {{ tag.value }}
              </ElTag>
            </span>
          </template>
          <!-- 状态条件列表 -->
          <template v-else-if="field.isConditions && Array.isArray(field.value)">
            <span class="conditions-list">
              <ElTag
                v-for="(cond, idx) in field.value"
                :key="idx"
                :type="getConditionType(cond.status)"
                size="small"
                class="condition-item"
              >
                {{ cond.type }}: {{ cond.status }}
                <span v-if="cond.reason" class="condition-reason">({{ cond.reason }})</span>
              </ElTag>
            </span>
          </template>
          <!-- 普通值 -->
          <template v-else>
            <span class="item-value">{{ field.value }}</span>
          </template>
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
  align-items: flex-start;
  gap: 8px;
}

.desc-item-full {
  grid-column: 1 / -1;
}

.item-content-inline {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.item-label {
  font-size: 12px;
  color: #909399;
  white-space: nowrap;
}

.item-value {
  font-size: 14px;
  color: #303133;
}

.tags-list,
.conditions-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tag-item,
.condition-item {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.condition-reason {
  font-size: 11px;
  opacity: 0.8;
  font-style: italic;
}
</style>
