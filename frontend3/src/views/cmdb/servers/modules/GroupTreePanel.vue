<script setup lang="ts">
  /**
   * 左侧分组树面板
   * 从 index.vue 拆分：包含分组搜索、树展示、右键菜单、分组创建对话框
   */

  import { computed, ref } from 'vue';
  import type { FormInstance, FormRules } from 'element-plus';
  import type { TreeNode } from '../types/server.types';

  const props = defineProps<{
    groupLoading: boolean;
    filteredGroupTree: TreeNode[];
    groupSearchKeyword: string;
    selectedGroupId: number | undefined;
    selectedUngrouped: boolean;
    editingNodeId: number | null;
    editingNodeName: string;
    contextMenuVisible: boolean;
    contextMenuPosition: { x: number; y: number };
    groupDialogVisible: boolean;
    groupFormData: { name: string; parentId: number };
    groupTree: TreeNode[];
    contextMenuNodeId: number | null;
    getNodeClass: (node: TreeNode) => string;
  }>();

  const emit = defineEmits<{
    (e: 'update:groupSearchKeyword', val: string): void;
    (e: 'update:editingNodeName', val: string): void;
    (e: 'node-click', data: TreeNode): void;
    (e: 'node-contextmenu', event: MouseEvent, data: TreeNode): void;
    (e: 'context-menu-close'): void;
    (e: 'add-root-group'): void;
    (e: 'refresh-groups'): void;
    (e: 'edit-keydown', event: KeyboardEvent): void;
    (e: 'save-edit-group'): void;
    (e: 'add-group'): void;
    (e: 'add-server'): void;
    (e: 'edit-group'): void;
    (e: 'delete-group'): void;
    (e: 'update:groupDialogVisible', val: boolean): void;
    (e: 'update:groupFormData', val: { name: string; parentId: number }): void;
    (e: 'save-group'): void;
  }>();

  // @ts-expect-error vue-tsc noUnusedLocals: template ref
  const treeRef = ref();
  // @ts-expect-error vue-tsc noUnusedLocals: template ref
  const editInputRef = ref<HTMLInputElement | null>(null);
  // @ts-expect-error vue-tsc noUnusedLocals: template ref
  const groupFormRef = ref<FormInstance>();

  // 当前右键节点是否为虚拟节点（根节点/未分组节点），不允增删改
  const isVirtualNode = computed(() => {
    return props.contextMenuNodeId === 0 || props.contextMenuNodeId === -1;
  });

  const groupFormRules: FormRules = {
    name: [
      { required: true, message: '请输入分组名称', trigger: 'blur' },
      { min: 2, max: 50, message: '分组名称长度在 2 到 50 个字符', trigger: 'blur' }
    ],
    parentId: [{ required: true, message: '请选择父分组', trigger: 'change' }]
  };
</script>

<template>
  <div style="width: 220px; height: 100%; display: flex; flex-direction: column; flex-shrink: 0; overflow: hidden;">
    <ElCard
      shadow="never"
      body-style="padding: 12px; border-radius: 0;"
      style="border-radius: 0; flex: 1; display: flex; flex-direction: column;"
    >
      <div class="mb-8px flex items-center justify-between">
        <span class="text-14px text-gray-700 font-bold">资产分组</span>
        <div class="flex items-center gap-4px">
          <ElButton link size="small" @click="emit('add-root-group')"><icon-mdi-plus class="text-16px" /></ElButton>
          <ElButton link size="small" @click="emit('refresh-groups')"><icon-mdi-refresh class="text-16px" /></ElButton>
        </div>
      </div>
      <ElInput
        :model-value="groupSearchKeyword"
        placeholder="搜索分组"
        clearable
        size="small"
        class="mb-12px"
        @update:model-value="emit('update:groupSearchKeyword', $event)"
      >
        <template #prefix><icon-mdi-magnify class="text-16px text-gray-400" /></template>
      </ElInput>

      <div class="flex-1 overflow-auto" @click="emit('context-menu-close')">
        <ElTree
          ref="treeRef"
          v-loading="groupLoading"
          :data="filteredGroupTree"
          node-key="id"
          :props="{ label: 'name', children: 'children' }"
          :highlight-current="true"
          :current-node-key="selectedUngrouped ? -1 : selectedGroupId"
          :expand-on-click-node="false"
          :default-expand-all="true"
          @node-click="(data: TreeNode) => emit('node-click', data)"
          @node-contextmenu="(event: MouseEvent, data: TreeNode) => emit('node-contextmenu', event, data)"
        >
          <template #default="{ node, data }">
            <div :class="getNodeClass(data)" class="w-full">
              <div class="w-full flex items-center justify-between pr-8px">
                <div v-if="editingNodeId === data.id" class="min-w-0 flex flex-1 items-center gap-6px">
                  <component
                    :is="'icon-' + data.icon.replace(':', '-')"
                    class="flex-shrink-0 text-16px"
                    :style="{ color: data.color }"
                  />
                  <ElInput
                    ref="editInputRef"
                    :model-value="editingNodeName"
                    size="small"
                    style="min-width: 0; flex: 1;"
                    @update:model-value="emit('update:editingNodeName', $event)"
                    @keydown="emit('edit-keydown', $event as KeyboardEvent)"
                    @blur="emit('save-edit-group')"
                    @click.stop
                  />
                </div>
                <div v-else class="min-w-0 flex flex-1 items-center gap-6px">
                  <component
                    :is="'icon-' + data.icon.replace(':', '-')"
                    class="flex-shrink-0 text-16px"
                    :style="{ color: data.color }"
                  />
                  <span class="truncate text-14px">{{ node.label }}</span>
                </div>
                <ElTag
                  v-if="data.serverCount !== undefined && editingNodeId !== data.id"
                  size="small"
                  type="info"
                  class="flex-shrink-0"
                >
                  {{ data.serverCount }}
                </ElTag>
              </div>
            </div>
          </template>
        </ElTree>
      </div>
    </ElCard>

    <!-- 右键菜单 -->
    <Teleport to="body">
      <Transition name="fade">
        <ElMenu
          v-if="contextMenuVisible"
          :style="{
            position: 'fixed',
            left: contextMenuPosition.x + 'px',
            top: contextMenuPosition.y + 'px',
            zIndex: 9999,
            minWidth: '150px',
            border: '1px solid #e4e7ed',
            borderRadius: '4px',
            boxShadow: '0 2px 12px 0 rgba(0, 0, 0, 0.1)',
            borderRight: 'none',
          }"
          @click.stop="emit('context-menu-close')"
        >
          <template v-if="isVirtualNode">
            <ElMenuItem
              index="add-server"
              :style="{ height: '36px', lineHeight: '36px' }"
              @click.stop="emit('add-server')"
            >
              <icon-mdi-server class="mr-8px" />
              <span>添加主机</span>
            </ElMenuItem>
          </template>
          <template v-else>
            <ElMenuItem
              index="add-group"
              :style="{ height: '36px', lineHeight: '36px' }"
              @click.stop="emit('add-group')"
            >
              <icon-mdi-plus class="mr-8px" />
              <span>添加分组</span>
            </ElMenuItem>
            <ElMenuItem
              index="add-server"
              :style="{ height: '36px', lineHeight: '36px' }"
              @click.stop="emit('add-server')"
            >
              <icon-mdi-server class="mr-8px" />
              <span>添加主机</span>
            </ElMenuItem>
            <ElMenuItem
              index="edit"
              :style="{ height: '36px', lineHeight: '36px' }"
              @click.stop="emit('edit-group')"
            >
              <icon-mdi-pencil class="mr-8px" />
              <span>重命名</span>
            </ElMenuItem>
            <ElMenuItem
              index="delete"
              :style="{ height: '36px', lineHeight: '36px', color: '#f56c6c' }"
              @click.stop="emit('delete-group')"
            >
              <icon-mdi-delete class="mr-8px" />
              <span>删除分组</span>
            </ElMenuItem>
          </template>
        </ElMenu>
      </Transition>
    </Teleport>

    <!-- 分组创建对话框 -->
    <ElDialog
      :model-value="groupDialogVisible"
      title="创建分组"
      width="500px"
      @update:model-value="emit('update:groupDialogVisible', $event)"
    >
      <ElForm ref="groupFormRef" :model="groupFormData" :rules="groupFormRules" label-width="100px">
        <ElFormItem label="分组名称" prop="name">
          <ElInput
            :model-value="groupFormData.name"
            placeholder="请输入分组名称（2-50个字符）"
            @update:model-value="
              groupFormData.name = $event;
              emit('update:groupFormData', groupFormData);
            "
          />
        </ElFormItem>
        <ElFormItem label="父分组" prop="parentId">
          <ElTreeSelect
            :model-value="groupFormData.parentId"
            :data="groupTree"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            placeholder="请选择父分组"
            check-strictly
            :render-after-expand="false"
            style="width: 100%"
            @update:model-value="
              groupFormData.parentId = $event;
              emit('update:groupFormData', groupFormData);
            "
          />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="emit('update:groupDialogVisible', false)">取消</ElButton>
        <ElButton type="primary" @click="emit('save-group')">确定</ElButton>
      </template>
    </ElDialog>
  </div>
</template>
