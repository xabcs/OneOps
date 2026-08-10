/**
 * 批量删除操作的通用 Hook
 *
 * 用于统一处理各种实体的批量删除操作，减少代码重复
 *
 * @example
 * const { handleBatchDelete } = useBatchDelete({
 *   deleteApi: fetchDeleteUser,
 *   confirmMessage: (count) => `确定要删除选中的 ${count} 个用户吗？`,
 *   successMessage: `成功删除用户`,
 *   onDeleted: async () => {
 *     await getAllRoles();
 *   }
 * });
 */

import { ElMessageBox, ElNotification } from 'element-plus';

export interface BatchDeleteOptions {
  /**
   * 删除API函数
   * @param id 要删除的项目ID
   * @returns Promise<any>
   */
  deleteApi: (id: number) => Promise<any>;

  /**
   * 成功消息模板
   * @default '成功删除 {count} 个项目'
   */
  successMessage?: string;

  /**
   * 错误消息模板
   * @default '{count} 个项目删除失败'
   */
  errorMessage?: string;

  /**
   * 确认消息生成函数
   * @param count 选中的项目数量
   * @returns 确认消息
   */
  confirmMessage?: (count: number) => string;

  /**
   * 删除完成后的回调
   */
  onDeleted?: () => void | Promise<void>;

  /**
   * 是否显示确认对话框
   * @default true
   */
  showConfirm?: boolean;
}

export function useBatchDelete(options: BatchDeleteOptions) {
  const {
    deleteApi,
    successMessage = '成功删除项目',
    errorMessage = '个项目删除失败',
    confirmMessage = (count: number) => `确定要删除选中的 ${count} 个项目吗？`,
    onDeleted,
    showConfirm = true
  } = options;

  /**
   * 处理批量删除操作
   * @param selectedIds 选中的项目ID数组
   */
  const handleBatchDelete = async (selectedIds: number[]) => {
    // 检查是否有选中项目
    if (selectedIds.length === 0) {
      ElNotification.warning({
        title: '警告',
        message: '请先选择要删除的项目'
      });
      return;
    }

    // 显示确认对话框
    if (showConfirm) {
      try {
        await ElMessageBox.confirm(confirmMessage(selectedIds.length), '批量删除', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        });
      } catch {
        // 用户取消操作
        return;
      }
    }

    // 执行批量删除
    let successCount = 0;
    let failCount = 0;

    for (const id of selectedIds) {
      try {
        await deleteApi(id);
        successCount++;
      } catch (error) {
        console.error(`删除项目 ${id} 失败:`, error);
        failCount++;
      }
    }

    // 显示结果通知
    if (successCount > 0) {
      ElNotification.success({
        title: '删除成功',
        message: `${successMessage} ${successCount} 个项目`
      });
    }

    if (failCount > 0) {
      ElNotification.error({
        title: '删除失败',
        message: `${failCount} ${errorMessage}`
      });
    }

    // 执行删除完成回调
    if (successCount > 0 && onDeleted) {
      try {
        await onDeleted();
      } catch (error) {
        console.error('删除完成回调执行失败:', error);
      }
    }
  };

  return {
    handleBatchDelete
  };
}
