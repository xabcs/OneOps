import { ref } from 'vue';
import { ElMessage } from 'element-plus';
import yaml from 'js-yaml';

export interface YamlEditOptions {
  /** 当前集群 ID（由调用方闭包提供，通常解析自 route.query） */
  getClusterId: () => number;
  /** 当前命名空间（由调用方闭包提供） */
  getNamespace: () => string;
  /** 资源更新 API（flat 请求，恒返回 {data, error} 不 reject） */
  updateFn: (
    clusterId: number,
    params: { namespace: string; manifest: Record<string, unknown> }
  ) => Promise<{ error?: unknown }>;
  /** 保存成功后的刷新回调（重载详情等） */
  reload: () => void | Promise<void>;
  /** 成功提示文案，默认'保存成功' */
  successMessage?: string;
}

/**
 * K8s 资源 YAML 编辑保存流程 composable
 * 统一各 detail 页 YamlEditor 的 on-apply 处理：YAML 语法校验 → 更新 API → 成功后刷新
 */
export function useYamlEdit(options: YamlEditOptions) {
  const { getClusterId, getNamespace, updateFn, reload, successMessage = '保存成功' } = options;

  const yamlSaving = ref(false);

  async function handleYamlApply(yamlStr: string) {
    yamlSaving.value = true;
    try {
      // 先在前端校验 YAML 语法，避免无效内容提交到后端
      let manifest: Record<string, unknown>;
      try {
        manifest = yaml.load(yamlStr) as Record<string, unknown>;
      } catch (error: unknown) {
        ElMessage.error(`YAML 格式错误: ${(error as Error).message}`);
        return;
      }

      // flat 请求恒返回 {data, error}，通过 error 字段判断结果
      const { error } = await updateFn(getClusterId(), { namespace: getNamespace(), manifest });
      if (error) {
        ElMessage.error((error as Error).message || '保存失败');
        return;
      }

      ElMessage.success(successMessage);
      await reload();
    } catch (error: unknown) {
      // reload 等环节的异常包装后上抛，交给调用链全局处理
      const err = error as Error;
      throw new Error(err.message || '保存失败');
    } finally {
      yamlSaving.value = false;
    }
  }

  return { yamlSaving, handleYamlApply };
}
