<script setup lang="ts">
import { ref } from 'vue';
import { ElTag, ElDialog, ElInput, ElButton, ElMessage } from 'element-plus';
import type { AnnotationItem } from '@/utils/k8s-formatters';

interface Field {
  label: string;
  value: string | number | Array<{ key: string; value: string; type?: string }> | AnnotationItem[] | StatusSummaryItem[];
  fullRow?: boolean;
  isTags?: boolean;
  isConditions?: boolean;
  isAnnotations?: boolean;
  isStatusSummary?: boolean;
}

interface StatusSummaryItem {
  type: string;
  value: string;
  status: 'success' | 'warning' | 'danger' | 'info';
}

const props = defineProps<{
  fields: Field[][];
}>();

// 注解弹窗
const showAnnotationDialog = ref(false);
const editingAnnotationKey = ref('');
const editingAnnotationValue = ref('');

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

// 获取注解标签颜色
const getAnnotationTagType = (key: string) => {
  const keyLower = key.toLowerCase();
  if (keyLower.includes('last-applied')) return 'warning';
  if (keyLower.includes('revision')) return 'success';
  if (keyLower.includes('project')) return 'primary';
  if (keyLower.includes('version')) return 'info';
  if (keyLower.includes('kubernetes.io/') || keyLower.includes('k8s.io/')) return 'info';
  return 'default';
};

// 查看注解完整内容
const handleViewAnnotation = (item: AnnotationItem) => {
  editingAnnotationKey.value = item.key;
  editingAnnotationValue.value = item.value;
  showAnnotationDialog.value = true;
};

// 关闭注解对话框
const handleCloseAnnotationDialog = () => {
  showAnnotationDialog.value = false;
  editingAnnotationKey.value = '';
  editingAnnotationValue.value = '';
};

// 复制注解内容
const handleCopyAnnotation = (text: string) => {
  navigator.clipboard.writeText(text).then(() => {
    ElMessage.success('已复制到剪贴板');
  }).catch(() => {
    ElMessage.error('复制失败');
  });
};

// 获取截断的键名
const getShortKey = (key: string) => {
  if (key.length > 25) {
    // 保留开头和结尾，中间用...代替
    if (key.startsWith('kubectl.kubernetes.io/')) {
      return 'kubectl.../' + key.split('/').pop();
    }
    if (key.startsWith('deployment.kubernetes.io/')) {
      return 'dep.../' + key.split('/').pop();
    }
    return key.substring(0, 10) + '...' + key.substring(key.length - 10);
  }
  return key;
};

// 获取截断的值
const getShortValue = (value: string) => {
  if (value.length > 20) {
    return value.substring(0, 20) + '...';
  }
  return value;
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
          <span class="item-label">{{ field.label }}</span>
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
                :type="cond.isStatusInfo ? 'info' : getConditionType(cond.status)"
                size="small"
                class="condition-item"
              >
                <template v-if="cond.isStatusInfo">
                  {{ cond.type }}: {{ cond.status }}
                </template>
                <template v-else>
                  {{ cond.type }}: {{ cond.status }}
                  <span v-if="cond.reason" class="condition-reason">({{ cond.reason }})</span>
                </template>
              </ElTag>
            </span>
          </template>
          <!-- 注解列表 -->
          <template v-else-if="field.isAnnotations && Array.isArray(field.value)">
            <span class="annotations-list">
              <ElTag
                v-for="(item, idx) in field.value"
                :key="idx"
                :type="getAnnotationTagType(item.key)"
                size="small"
                class="annotation-tag"
                @click="handleViewAnnotation(item)"
              >
                <span v-if="item.isLong" class="annotation-key-short">{{ getShortKey(item.key) }}</span>
                <span v-else class="annotation-key-full">{{ item.key }}</span>
                <span v-if="!item.isLong" class="annotation-sep">:</span>
                <span v-if="item.isLong" class="expand-hint">展开</span>
                <span v-else class="annotation-value-short">{{ getShortValue(item.value) }}</span>
              </ElTag>
            </span>
          </template>
          <!-- 状态摘要列表 -->
          <template v-else-if="field.isStatusSummary && Array.isArray(field.value)">
            <span class="status-summary-list">
              <ElTag
                v-for="(item, idx) in field.value"
                :key="idx"
                :type="item.status"
                size="small"
                class="status-summary-item"
              >
                {{ item.type }}: {{ item.value }}
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

    <!-- 注解详情对话框 -->
    <ElDialog
      v-model="showAnnotationDialog"
      :title="`注解: ${editingAnnotationKey}`"
      width="800px"
      top="5vh"
      @close="handleCloseAnnotationDialog"
    >
      <div class="annotation-detail">
        <div class="detail-label">键名:</div>
        <div class="detail-content">{{ editingAnnotationKey }}</div>

        <div class="detail-label">值:</div>
        <ElInput
          v-model="editingAnnotationValue"
          type="textarea"
          :rows="15"
          readonly
          class="detail-textarea"
        />

        <div class="detail-actions">
          <ElButton @click="handleCloseAnnotationDialog">关闭</ElButton>
          <ElButton type="primary" @click="handleCopyAnnotation(editingAnnotationValue)">复制内容</ElButton>
        </div>
      </div>
    </ElDialog>
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
  grid-template-columns: repeat(2, 1fr);
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
  align-items: flex-start;
  gap: 12px;
  flex-wrap: nowrap;
}

.item-label {
  font-size: 12px;
  color: #333;
  white-space: nowrap;
  width: 80px;
  flex-shrink: 0;
  padding-top: 2px;
}

.item-value {
  font-size: 12px;
  color: #333;
}

.tags-list,
.conditions-list,
.annotations-list,
.status-summary-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  align-items: flex-start;
  width: 100%;
}

/* 统一所有标签的基础样式 */
.tag-item,
.condition-item,
.annotation-tag,
.status-summary-item {
  display: flex;
  align-items: center;
  width: 100%;
  cursor: pointer;
  font-size: 12px;
  font-family: "Courier New", Courier, monospace;
  margin: 0;
  padding: 4px 8px;
  border: none !important;
}

/* 标签和状态条件特殊样式 */
.tag-item,
.condition-item {
  /* 垂直排列，不需要横向间距 */
}

/* 注解标签特殊样式 */
.annotation-tag {
  /* 保持默认样式 */
}

/* 标签hover效果 */
.tag-item:hover,
.condition-item:hover,
.annotation-tag:hover {
  opacity: 0.85;
}

/* 注解标签内部元素 */
.annotation-key-full,
.annotation-key-short,
.annotation-sep,
.annotation-value-short {
  color: inherit;
}

.annotation-key-full {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 180px;
}

.annotation-key-short {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 120px;
  font-weight: 500;
}

.annotation-sep {
  flex-shrink: 0;
}

.annotation-value-short {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 80px;
  flex-shrink: 1;
}

.expand-hint {
  color: inherit;
  font-weight: 600;
  font-size: 11px;
  flex-shrink: 0;
}


/* 状态条件原因文字 */
.condition-reason {
  font-size: 11px;
  color: inherit;
  opacity: 0.8;
  font-style: italic;
}

/* 注解详情对话框样式 */
.annotation-detail {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.detail-label {
  font-size: 13px;
  font-weight: 500;
  color: #606266;
}

.detail-content {
  font-size: 14px;
  color: #303133;
  word-break: break-all;
  padding: 8px 12px;
  background: #f5f7fa;
  border-radius: 4px;
}

.detail-textarea {
  font-family: "Courier New", Courier, monospace;
}

.detail-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding-top: 8px;
}
</style>
