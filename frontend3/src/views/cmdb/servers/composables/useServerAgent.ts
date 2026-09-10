/**
 * Agent 管理逻辑
 * 从 index.vue 拆分：Agent 部署/重启/卸载/状态轮询、批量操作
 */

import { onScopeDispose } from 'vue';
import { ElMessageBox, ElNotification } from 'element-plus';
import {
  fetchBatchDeployAgent,
  fetchBatchUninstallAgent,
  fetchDeployAgent,
  fetchGetAgentStatus,
  fetchRestartAgent,
  fetchTestSSHConnection,
  fetchUninstallAgent
} from '@/service/api';

function resolveErrorMsg(error: unknown): string {
  if (error && typeof error === 'object') {
    const errObj = error as Record<string, unknown>;
    const response = errObj.response as { data?: { message?: string } } | undefined;
    if (response?.data?.message) return response.data.message;
    if (typeof errObj.message === 'string') return errObj.message;
  }
  return '';
}

export function useServerAgent() {
  // 收集进行中的轮询定时器，作用域销毁（组件卸载）时统一清理，避免切页后仍持续请求
  const pollTimers = new Set<ReturnType<typeof setInterval>>();

  // ===== 单台 Agent 操作 =====
  async function handleAgentDeploy(row: CMDB.Server, tableData: CMDB.Server[]) {
    try {
      const testResult = await fetchTestSSHConnection(row.id);
      if (!testResult.data || !testResult.data.success) {
        await ElMessageBox.alert(`${testResult.data?.message || '连接失败'}`, '连接失败', {
          type: 'error',
          confirmButtonText: '我知道了'
        });
        return;
      }
      await fetchDeployAgent(row.id);
      ElNotification.success({
        message: `Agent 部署任务已提交，正在后台执行...\nSSH连接测试成功（延迟: ${testResult.data.latency}）`,
        duration: 3000
      });
      pollAgentStatus(row.id, 'running', tableData);
    } catch (error: unknown) {
      console.error('Agent 部署失败:', error);
      const errorMsg = resolveErrorMsg(error) || '部署失败';
      if (errorMsg.includes('系统运维凭证') || errorMsg.includes('systemCredential')) {
        ElNotification.error({
          message: '部署失败：主机未配置系统运维凭证，请先在主机编辑页配置',
          duration: 0,
          customClass: 'agent-error-notification'
        });
      } else if (errorMsg.includes('SSH 连接失败')) {
        ElNotification.error({ message: `部署失败：无法连接到主机 ${row.hostname || row.ip}`, duration: 0 });
      } else {
        ElNotification.error(`Agent 部署失败：${errorMsg}`);
      }
    }
  }

  async function handleAgentRestart(row: CMDB.Server, tableData: CMDB.Server[]) {
    try {
      await fetchRestartAgent(row.id);
      ElNotification.info('Agent 重启任务已提交...');
      pollAgentStatus(row.id, 'running', tableData);
    } catch (error: unknown) {
      console.error('Agent 重启失败:', error);
      const errorMsg = resolveErrorMsg(error) || '重启失败';
      if (errorMsg.includes('SSH 连接失败')) {
        ElNotification.error({ message: `重启失败：无法连接到主机 ${row.hostname || row.ip}`, duration: 0 });
      } else {
        ElNotification.error(`Agent 重启失败：${errorMsg}`);
      }
    }
  }

  async function handleAgentUninstall(row: CMDB.Server, tableData: CMDB.Server[]) {
    try {
      await ElMessageBox.confirm(`确认卸载 ${row.hostname} 上的 Agent？卸载后主机将不再上报监控数据。`, '卸载确认', {
        type: 'warning',
        confirmButtonText: '确认卸载',
        cancelButtonText: '取消'
      });
      const testResult = await fetchTestSSHConnection(row.id);
      if (!testResult.data || !testResult.data.success) {
        await ElMessageBox.alert(`${testResult.data?.message || '连接失败'}`, '连接失败', {
          type: 'error',
          confirmButtonText: '我知道了'
        });
        return;
      }
      await fetchUninstallAgent(row.id);
      ElNotification.success({
        message: `Agent 卸载任务已提交，正在后台执行...\nSSH连接测试成功（延迟: ${testResult.data.latency}）`,
        duration: 3000
      });
      pollAgentStatus(row.id, 'uninstalled', tableData);
    } catch (error: unknown) {
      if (error === 'cancel') return;
      console.error('Agent 卸载失败:', error);
      const errorMsg = resolveErrorMsg(error) || '卸载失败';
      if (errorMsg.includes('系统运维凭证') || errorMsg.includes('systemCredential')) {
        ElNotification.error({
          message: '卸载失败：主机未配置系统运维凭证，请先在主机编辑页配置',
          duration: 0,
          customClass: 'agent-error-notification'
        });
      } else if (errorMsg.includes('SSH 连接失败')) {
        ElNotification.error({ message: `卸载失败：无法连接到主机 ${row.hostname || row.ip}`, duration: 0 });
      } else {
        ElNotification.error(`Agent 卸载失败：${errorMsg}`);
      }
    }
  }

  // ===== 轮询 Agent 状态 =====
  function pollAgentStatus(
    serverId: number,
    expectedStatus: CMDB.Server['agentStatus'] = 'running',
    tableData: CMDB.Server[],
    maxTimes = 20
  ) {
    let count = 0;
    let previousStatus: string | null = null;
    const timer = setInterval(async () => {
      count++;
      try {
        const res = await fetchGetAgentStatus(serverId);
        // 异常场景后端可能返回 failed，类型断言将其纳入，避免比较被判定为无重叠
        const status = res.data?.agentStatus as CMDB.Server['agentStatus'] | 'failed' | undefined;

        const idx = tableData.findIndex(s => s.id === serverId);
        if (idx !== -1 && res.data) {
          tableData[idx] = {
            ...tableData[idx],
            agentStatus: status as CMDB.Server['agentStatus'],
            agentPort: res.data.agentPort,
            agentVersion: res.data.agentVersion,
            lastHeartbeatAt: res.data.lastHeartbeatAt
          };

          if (previousStatus !== null && previousStatus !== status) {
            if (status === 'running')
              ElNotification.success(`主机 ${tableData[idx].hostname} 的 Agent 已成功部署并运行`);
            else if (status === 'uninstalled')
              ElNotification.success(`主机 ${tableData[idx].hostname} 的 Agent 已成功卸载`);
            else if (status === 'failed')
              ElNotification.warning(`主机 ${tableData[idx].hostname} 的 Agent 状态异常，请检查日志`);
          }
          previousStatus = status ?? null;
        }

        if (status === expectedStatus) {
          clearInterval(timer);
          pollTimers.delete(timer);
          return;
        }
        if (count >= maxTimes) {
          clearInterval(timer);
          pollTimers.delete(timer);
          const serverName = tableData.find(s => s.id === serverId)?.hostname || serverId;
          if (status !== expectedStatus)
            ElNotification.warning(
              `主机 ${serverName} 的 Agent 状态轮询超时（当前状态: ${status}），请手动刷新页面查看最新状态`
            );
        }
      } catch (error: unknown) {
        console.error('轮询 Agent 状态失败:', error);
        if (count >= maxTimes) {
          clearInterval(timer);
          pollTimers.delete(timer);
          ElNotification.error('轮询 Agent 状态失败，请刷新页面查看最新状态');
        }
      }
    }, 3000);
    pollTimers.add(timer);
  }

  // ===== 批量操作 =====
  async function handleBatchTestConnection(selectedIds: number[], tableData: CMDB.Server[]) {
    if (selectedIds.length === 0) {
      ElNotification.warning('请先选择要测试连接的主机');
      return;
    }

    const results = {
      success: 0,
      failed: 0,
      details: [] as Array<{ id: number; hostname: string; success: boolean; message: string }>
    };
    ElNotification.info(`正在测试 ${selectedIds.length} 台主机的SSH连接...`);

    for (const serverId of selectedIds) {
      const server = tableData.find(s => s.id === serverId);
      if (!server) continue;
      try {
        const testResult = await fetchTestSSHConnection(serverId);
        // flat 封装的 data 可能为 null，先收窄再使用（为空视为测试失败，走 catch 计数）
        const result = testResult.data;
        if (!result) throw new Error('连接测试无返回');
        results.details.push({
          id: serverId,
          hostname: server.hostname || server.ip,
          success: result.success,
          message: result.message
        });
        if (result.success) results.success++;
        else results.failed++;
      } catch (error: unknown) {
        results.failed++;
        results.details.push({
          id: serverId,
          hostname: server.hostname || server.ip,
          success: false,
          message: resolveErrorMsg(error) || '连接测试失败'
        });
      }
    }

    const summary = `SSH连接测试完成：\n成功: ${results.success} 台\n失败: ${results.failed} 台`;
    if (results.failed > 0) {
      const failedServers = results.details
        .filter(d => !d.success)
        .map(d => `- ${d.hostname}: ${d.message}`)
        .join('\n');
      await ElMessageBox.alert(
        `${summary}\n\n以下主机连接失败：\n${failedServers}\n\n请检查主机配置和网络连接。`,
        '连接测试结果',
        { type: 'warning', confirmButtonText: '我知道了' }
      );
    } else {
      ElNotification.success(`${summary}\n所有主机连接测试成功！`);
    }
  }

  async function handleBatchDeploy(
    selectedIds: number[],
    tableData: CMDB.Server[],
    pollFn: (id: number, status: CMDB.Server['agentStatus']) => void
  ) {
    if (selectedIds.length === 0) {
      ElNotification.warning('请先选择要部署 Agent 的主机');
      return;
    }

    const serversWithoutCred = tableData
      .filter(s => selectedIds.includes(s.id))
      .filter(s => !s.systemCredentialId || s.systemCredentialId === 0);
    if (serversWithoutCred.length > 0) {
      const serverNames = serversWithoutCred.map(s => s.hostname || s.ip).join('、');
      await ElMessageBox.alert(
        `以下主机未配置系统运维凭证，无法部署：\n\n${serverNames}\n\n请先编辑主机配置系统运维凭证`,
        '凭证未配置',
        { type: 'warning', confirmButtonText: '我知道了' }
      );
      return;
    }

    try {
      await ElMessageBox.confirm(
        `确定要为选中的 ${selectedIds.length} 台主机部署 Agent 吗？部署过程可能需要几分钟，请耐心等待。`,
        '批量部署确认',
        { type: 'info', confirmButtonText: '确认部署', cancelButtonText: '取消' }
      );
      await fetchBatchDeployAgent(selectedIds);
      ElNotification.success(`批量部署任务已提交，正在后台执行 ${selectedIds.length} 台主机的 Agent 部署`);
      selectedIds.forEach(serverId => pollFn(serverId, 'running'));
    } catch (error: unknown) {
      if (error === 'cancel') return;
      console.error('批量部署失败:', error);
      ElNotification.error(`批量部署失败：${resolveErrorMsg(error) || '批量部署失败'}`);
    }
  }

  async function handleBatchUninstall(
    selectedIds: number[],
    tableData: CMDB.Server[],
    pollFn: (id: number, status: CMDB.Server['agentStatus']) => void
  ) {
    if (selectedIds.length === 0) {
      ElNotification.warning('请先选择要卸载 Agent 的主机');
      return;
    }

    const serversWithoutCred = tableData
      .filter(s => selectedIds.includes(s.id))
      .filter(s => !s.systemCredentialId || s.systemCredentialId === 0);
    if (serversWithoutCred.length > 0) {
      const serverNames = serversWithoutCred.map(s => s.hostname || s.ip).join('、');
      await ElMessageBox.alert(
        `以下主机未配置系统运维凭证，无法卸载：\n\n${serverNames}\n\n请先编辑主机配置系统运维凭证`,
        '凭证未配置',
        { type: 'warning', confirmButtonText: '我知道了' }
      );
      return;
    }

    try {
      await ElMessageBox.confirm(
        `确定要从选中的 ${selectedIds.length} 台主机卸载 Agent 吗？卸载后主机将不再上报监控数据。`,
        '批量卸载确认',
        { type: 'warning', confirmButtonText: '确认卸载', cancelButtonText: '取消' }
      );
      await fetchBatchUninstallAgent(selectedIds);
      ElNotification.success(`批量卸载任务已提交，正在后台执行 ${selectedIds.length} 台主机的 Agent 卸载`);
      selectedIds.forEach(serverId => pollFn(serverId, 'uninstalled'));
    } catch (error: unknown) {
      if (error === 'cancel') return;
      console.error('批量卸载失败:', error);
      ElNotification.error(`批量卸载失败：${resolveErrorMsg(error) || '批量卸载失败'}`);
    }
  }

  // 组件卸载（作用域销毁）时清理所有仍在进行的轮询定时器
  onScopeDispose(() => {
    pollTimers.forEach(timer => clearInterval(timer));
    pollTimers.clear();
  });

  return {
    handleAgentDeploy,
    handleAgentRestart,
    handleAgentUninstall,
    pollAgentStatus,
    handleBatchTestConnection,
    handleBatchDeploy,
    handleBatchUninstall
  };
}
