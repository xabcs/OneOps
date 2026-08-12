<script setup lang="ts">
  import { nextTick, onBeforeUnmount, ref, watch } from 'vue';
  import { ElMessage } from 'element-plus';
  import { Check, Edit, MagicStick, WarningFilled } from '@element-plus/icons-vue';
  import * as Monaco from 'monaco-editor/esm/vs/editor/editor.api.js';

  // 简单配置 - 不需要 worker，使用基本的编辑器功能
  (self as Record<string, unknown>).MonacoEnvironment = {
    getWorkerUrl(_moduleId: string, _label: string) {
      // 不使用 worker，直接返回空
      return '';
    }
  };

  interface Props {
    modelValue: boolean;
    title: string;
    yaml: string;
    canEdit?: boolean;
    onApply?: (yaml: string) => Promise<void>;
  }

  interface Emits {
    (e: 'update:modelValue', value: boolean): void;
  }

  const props = withDefaults(defineProps<Props>(), {
    canEdit: true,
    onApply: undefined
  });

  const emit = defineEmits<Emits>();

  const visible = ref(false);
  const yamlContent = ref('');
  const originalYaml = ref('');
  const isEditMode = ref(false);
  const errorMessage = ref('');
  const applying = ref(false);
  const editorContainer = ref<HTMLElement | null>(null);
  const useMonaco = ref(true); // 是否使用 Monaco Editor
  let editor: Monaco.editor.IStandaloneCodeEditor | null = null;

  // Monaco Editor 主题定义
  const DARK_THEME = 'k8s-yaml-dark';

  // 初始化 Monaco Editor
  function initEditor() {
    if (!editorContainer.value) return;

    try {
      // 定义自定义暗色主题（类似VS Code）
      Monaco.editor.defineTheme(DARK_THEME, {
        base: 'vs-dark',
        inherit: true,
        rules: [
          { token: 'key', foreground: '9CDCFE' },
          { token: 'string', foreground: 'CE9178' },
          { token: 'number', foreground: 'B5CEA8' },
          { token: 'comment', foreground: '6A9955' },
          { token: 'type', foreground: '4EC9B0' }
        ],
        colors: {
          'editor.background': '#1e1e1e',
          'editor.foreground': '#d4d4d4',
          'editor.lineHighlightBackground': '#2a2d2e',
          'editorLineNumber.foreground': '#858585',
          'editorCursor.foreground': '#aeafad',
          'editor.selectionBackground': '#264f78'
        }
      });

      editor = Monaco.editor.create(editorContainer.value, {
        value: yamlContent.value,
        language: 'yaml',
        theme: DARK_THEME,
        readOnly: true,
        automaticLayout: true,
        minimap: { enabled: false },
        fontSize: 14,
        fontFamily: "'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', 'source-code-pro', monospace",
        lineNumbers: 'on',
        scrollBeyondLastLine: false,
        wordWrap: 'on',
        tabSize: 2,
        insertSpaces: true,
        renderWhitespace: 'selection',
        bracketPairColorization: { enabled: true },
        scrollbar: {
          verticalScrollbarSize: 14,
          horizontalScrollbarSize: 14
        },
        padding: { top: 16, bottom: 16 }
      });

      // 监听内容变化
      editor.onDidChangeModelContent(() => {
        errorMessage.value = '';
        if (editor) {
          yamlContent.value = editor.getValue();
        }
      });

      useMonaco.value = true;
    } catch (error) {
      console.error('Monaco Editor 初始化失败，切换到简单编辑器:', error);
      useMonaco.value = false;
      // 不显示错误消息，因为已经有备用方案
    }
  }

  // 更新编辑器内容
  function updateEditorContent(content: string) {
    if (editor) {
      const model = editor.getModel();
      if (model) {
        model.setValue(content);
      }
    }
  }

  // 监听 modelValue 变化
  watch(
    () => props.modelValue,
    async val => {
      visible.value = val;
      if (val) {
        yamlContent.value = props.yaml || '';
        originalYaml.value = props.yaml || '';
        isEditMode.value = false;
        errorMessage.value = '';

        await nextTick();
        if (!editor) {
          initEditor();
        }
        if (editor) {
          updateEditorContent(yamlContent.value);
          editor.updateOptions({ readOnly: true });
        }
      }
    }
  );

  // 监听 visible 变化
  watch(visible, val => {
    if (!val) {
      emit('update:modelValue', false);
    }
  });

  // 开始编辑
  function startEdit() {
    isEditMode.value = true;
    errorMessage.value = '';
    if (useMonaco.value && editor) {
      editor.updateOptions({ readOnly: false });
      editor.focus();
    }
  }

  // 取消编辑
  function handleCancel() {
    yamlContent.value = originalYaml.value;
    isEditMode.value = false;
    errorMessage.value = '';
    if (useMonaco.value && editor) {
      updateEditorContent(originalYaml.value);
      editor.updateOptions({ readOnly: true });
    }
  }

  // 应用更改
  async function handleApply() {
    if (!props.onApply) {
      ElMessage.warning('未设置应用回调函数');
      return;
    }

    // 验证 YAML 格式
    if (!validateYaml()) {
      return;
    }

    applying.value = true;
    try {
      await props.onApply(yamlContent.value);
      ElMessage.success('YAML 应用成功');
      originalYaml.value = yamlContent.value;
      isEditMode.value = false;
      if (useMonaco.value && editor) {
        editor.updateOptions({ readOnly: true });
      }
    } catch (error: unknown) {
      ElMessage.error((error as Error).message || 'YAML 应用失败');
    } finally {
      applying.value = false;
    }
  }

  // 关闭对话框
  function handleClose() {
    if (isEditMode.value && yamlContent.value !== originalYaml.value) {
      ElMessage.warning('有未保存的更改，请先应用或取消');
      return;
    }
    visible.value = false;
  }

  // 格式化 YAML
  function handleFormat() {
    if (!useMonaco.value) {
      // 简单编辑器模式：直接格式化 yamlContent
      try {
        const content = yamlContent.value;
        const lines = content.split('\n');
        const formatted: string[] = [];

        for (const line of lines) {
          const trimmed = line.trim();
          if (!trimmed) {
            formatted.push('');
            continue;
          }
          // 计算当前行前面有多少空格
          const match = line.match(/^(\s*)/);
          const indent = match ? match[1].length : 0;
          // 确保缩进是2的倍数
          const newIndent = Math.floor(indent / 2) * 2;
          formatted.push(' '.repeat(newIndent) + trimmed);
        }

        yamlContent.value = formatted.join('\n');
        ElMessage.success('格式化成功');
      } catch (error) {
        ElMessage.error('格式化失败');
      }
      return;
    }

    if (!editor) return;

    try {
      // 简单格式化：调整缩进
      const content = editor.getValue();
      const lines = content.split('\n');
      const formatted: string[] = [];

      for (const line of lines) {
        const trimmed = line.trim();
        if (!trimmed) {
          formatted.push('');
          continue;
        }
        // 计算当前行前面有多少空格
        const match = line.match(/^(\s*)/);
        const indent = match ? match[1].length : 0;
        // 确保缩进是2的倍数
        const newIndent = Math.floor(indent / 2) * 2;
        formatted.push(' '.repeat(newIndent) + trimmed);
      }

      const formattedContent = formatted.join('\n');
      updateEditorContent(formattedContent);
      yamlContent.value = formattedContent;
      ElMessage.success('格式化成功');
    } catch (error) {
      ElMessage.error('格式化失败');
    }
  }

  // 简单的 YAML 验证
  function validateYaml(): boolean {
    const yaml = yamlContent.value.trim();

    if (!yaml) {
      errorMessage.value = 'YAML 内容不能为空';
      return false;
    }

    const lines = yaml.split('\n');
    for (let i = 0; i < lines.length; i++) {
      const line = lines[i];

      // 检查是否包含非法的Tab字符
      if (line.includes('\t')) {
        errorMessage.value = `第 ${i + 1} 行: YAML 不允许使用 Tab 字符，请使用空格缩进`;
        return false;
      }
    }

    return true;
  }

  // 组件卸载前销毁编辑器
  onBeforeUnmount(() => {
    if (editor) {
      editor.dispose();
      editor = null;
    }
  });
</script>

<template>
  <ElDialog
    v-model="visible"
    :title="title"
    width="900px"
    :close-on-click-modal="false"
    class="yaml-editor-dialog"
    @close="handleClose"
  >
    <div class="yaml-editor-container">
      <!-- 工具栏 -->
      <div class="yaml-toolbar">
        <div class="toolbar-left">
          <ElTag v-if="!isEditMode" type="info">只读模式</ElTag>
          <ElTag v-else type="warning">编辑模式</ElTag>
          <span class="yaml-type-label">YAML</span>
        </div>
        <div class="toolbar-right">
          <ElButton v-if="!isEditMode && canEdit" type="primary" size="small" @click="startEdit">
            <ElIcon>
              <Edit />
            </ElIcon>
            编辑
          </ElButton>
          <template v-if="isEditMode">
            <ElButton type="success" size="small" :loading="applying" @click="handleApply">
              <ElIcon>
                <Check />
              </ElIcon>
              应用
            </ElButton>
            <ElButton size="small" @click="handleCancel">取消</ElButton>
          </template>
          <ElButton size="small" @click="handleFormat">
            <ElIcon>
              <MagicStick />
            </ElIcon>
            格式化
          </ElButton>
        </div>
      </div>

      <!-- Monaco Editor 容器 -->
      <div v-if="useMonaco" ref="editorContainer" class="monaco-editor-wrapper" />

      <!-- 备用简单编辑器 -->
      <textarea
        v-else
        v-model="yamlContent"
        class="simple-editor"
        :readonly="!isEditMode"
        spellcheck="false"
      ></textarea>

      <!-- 错误信息 -->
      <div v-if="errorMessage" class="error-message">
        <ElIcon class="error-icon">
          <WarningFilled />
        </ElIcon>
        <span>{{ errorMessage }}</span>
      </div>
    </div>
  </ElDialog>
</template>

<style scoped>
  .yaml-editor-dialog :deep(.el-dialog__body) {
    padding: 0;
    max-height: 70vh;
  }

  .yaml-editor-container {
    display: flex;
    flex-direction: column;
  }

  .yaml-toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 20px;
    background-color: #252526;
    border-bottom: 1px solid #3e3e3e;
  }

  .toolbar-left {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .yaml-type-label {
    color: #858585;
    font-size: 12px;
    font-family: monospace;
  }

  .toolbar-right {
    display: flex;
    gap: 8px;
  }

  .monaco-editor-wrapper {
    width: 100%;
    min-height: 500px;
    max-height: 60vh;
    border: none;
  }

  .simple-editor {
    width: 100%;
    min-height: 500px;
    max-height: 60vh;
    padding: 16px;
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', monospace;
    font-size: 14px;
    line-height: 1.6;
    background: #1e1e1e;
    color: #d4d4d4;
    border: none;
    resize: none;
    outline: none;
  }

  .simple-editor:focus {
    outline: none;
  }

  .error-message {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 12px 20px;
    background-color: #4a2a2a;
    border-top: 1px solid #8b3a3a;
    color: #ff6b6b;
    font-size: 13px;
  }

  .error-icon {
    font-size: 18px;
    flex-shrink: 0;
  }
</style>
