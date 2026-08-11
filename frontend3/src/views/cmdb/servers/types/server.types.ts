/**
 * 服务器模块公共类型定义
 */

/** 树节点 */
export interface TreeNode {
  id: number;
  name: string;
  code: string;
  parentId: number;
  level: number;
  description?: string;
  color: string;
  icon: string;
  sortOrder: number;
  status: number;
  children?: TreeNode[];
  serverCount?: number;
}

/** 搜索类型 */
export type SearchType = 'hostname' | 'ip' | 'group';

/** 服务器筛选参数 */
export interface ServerFilters {
  keyword?: string;
  env?: string;
  status?: number;
  agentStatus?: string;
  groupId?: number;
}

/** 服务器表单数据 */
export interface ServerFormData {
  hostname?: string;
  ip?: string;
  innerIp?: string;
  sshPort?: number;
  env?: string;
  status?: number;
  provider?: string;
  groupId?: number[];
  credentialId?: number;
  systemCredentialId?: number;
  cabinetId?: number;
  remarks?: string;
}

/** Agent 状态类型 */
export type AgentStatus = 'running' | 'offline' | 'uninstalled' | 'failed' | 'unknown' | '';
