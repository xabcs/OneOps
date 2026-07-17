<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { ElButton, ElDialog, ElForm, ElFormItem, ElIcon, ElInput, ElMessage, ElTag } from 'element-plus';
import { Delete, Edit, Folder, FolderOpened, Search } from '@element-plus/icons-vue';
import type { TreeNode } from '../types/server.types';

interface Props {
  treeData: TreeNode[];
  loading?: boolean;
  selectedGroupId?: number;
}

interface Emits {
  (e: 'node-click', node: TreeNode): void;
  (e: 'create-group', parentData: TreeNode): void;
  (e: 'update-group', nodeId: number, newName: string): void;
  (e: 'delete-group', nodeId: number): void;
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  selectedGroupId: undefined
});

const emit = defineEmits<Emits>();

const treeRef = ref();
const searchKeyword = ref('');

// 编辑状态
const editDialogVisible = ref(false);
const editForm = ref({
  id: 0,
  name: ''
});
const editingNode = ref<TreeNode | null>(null);

// 树形组件配置
const treeProps = {
  children: 'children',
  label: 'name'
};

// 过滤后的树数据
const filteredTreeData = computed(() => {
  if (!searchKeyword.value) return props.treeData;
  return filterTree(props.treeData, searchKeyword.value);
});

// 过滤树节点
const filterTree = (nodes: TreeNode[], keyword: string): TreeNode[] => {
  if (!keyword) return nodes;

  const result: TreeNode[] = [];
  const lowerKeyword = keyword.toLowerCase();

  nodes.forEach(node => {
    const matches = node.name.toLowerCase().includes(lowerKeyword);
    const filteredChildren = node.children ? filterTree(node.children, keyword) : [];

    if (matches || filteredChildren.length > 0) {
      result.push({
        ...node,
        children: filteredChildren.length > 0 ? filteredChildren : node.children
      });
    }
  });

  return result;
};

// 过滤节点方法（供el-tree使用）
const filterNode = (value: string, data: TreeNode) => {
  if (!value) return true;
  return data.name.toLowerCase().includes(value.toLowerCase());
};

// 处理节点点击
const handleNodeClick = (data: TreeNode) => {
  emit('node-click', data);
};

// 处理编辑节点
const handleEditNode = (data: TreeNode) => {
  editingNode.value = data;
  editForm.value = {
    id: data.id,
    name: data.name
  };
  editDialogVisible.value = true;
};

// 确认编辑
const confirmEdit = () => {
  if (!editForm.value.name.trim()) {
    ElMessage.warning('分组名称不能为空');
    return;
  }

  emit('update-group', editForm.value.id, editForm.value.name);
  editDialogVisible.value = false;
};

// 处理删除节点
const handleDeleteNode = (data: TreeNode) => {
  emit('delete-group', data.id);
};

// 监听搜索关键词变化
watch(searchKeyword, val => {
  treeRef.value?.filter(val);
});
</script>

<template>
  <div class="group-tree-container">
    <!-- 搜索框 -->
    <ElInput v-model="searchKeyword" placeholder="搜索分组..." clearable size="small" class="search-input">
      <template #prefix>
        <ElIcon><Search /></ElIcon>
      </template>
    </ElInput>

    <!-- 分组树 -->
    <ElTree
      ref="treeRef"
      v-loading="loading"
      :data="filteredTreeData"
      :props="treeProps"
      :expand-on-click-node="false"
      :highlight-current="true"
      :filter-node-method="filterNode"
      node-key="id"
      class="group-tree"
      @node-click="handleNodeClick"
    >
      <template #default="{ node, data }">
        <div class="custom-node">
          <div class="node-content">
            <div class="node-icon">
              <ElIcon v-if="data.id === 0" color="#409EFF">
                <FolderOpened />
              </ElIcon>
              <ElIcon v-else color="#67C23A">
                <Folder />
              </ElIcon>
            </div>

            <span class="node-label">{{ node.label }}</span>

            <div v-if="data.serverCount !== undefined" class="node-count">
              <ElTag size="small" type="info">{{ data.serverCount }}</ElTag>
            </div>
          </div>

          <!-- 操作按钮 -->
          <div v-if="data.id !== 0" class="node-actions">
            <ElButton size="small" text @click.stop="handleEditNode(data)">
              <ElIcon><Edit /></ElIcon>
            </ElButton>
            <ElButton size="small" text type="danger" @click.stop="handleDeleteNode(data)">
              <ElIcon><Delete /></ElIcon>
            </ElButton>
          </div>
        </div>
      </template>
    </ElTree>

    <!-- 编辑节点对话框 -->
    <ElDialog v-model="editDialogVisible" title="编辑分组" width="400px">
      <ElForm :model="editForm" label-width="80px">
        <ElFormItem label="分组名称">
          <ElInput v-model="editForm.name" placeholder="请输入分组名称" />
        </ElFormItem>
      </ElForm>

      <template #footer>
        <ElButton @click="editDialogVisible = false">取消</ElButton>
        <ElButton type="primary" @click="confirmEdit">确定</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<style scoped>
.group-tree-container {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.search-input {
  margin-bottom: 12px;
}

.group-tree {
  flex: 1;
  overflow-y: auto;
}

.custom-node {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 4px 0;
}

.node-content {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
}

.node-icon {
  display: flex;
  align-items: center;
}

.node-label {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.node-count {
  margin-left: auto;
}

.node-actions {
  display: flex;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.2s;
}

.custom-node:hover .node-actions {
  opacity: 1;
}
</style>
