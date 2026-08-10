/**
 * 服务器管理相关类型定义
 */

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

export interface ServerFormData {
  hostname?: string;
  ip?: string;
  innerIp?: string;
  sshPort?: number;
  env?: string;
  provider?: string;
  agentStatus?: string;
  status?: number;
  groupId?: number[];
  credentialId?: number;
  systemCredentialId?: number;
  cabinetId?: number;
  remarks?: string;
}

export interface GroupFormData {
  name?: string;
  code?: string;
  parentId?: number;
  owner?: string;
  description?: string;
}

export interface ServerFilters {
  keyword?: string;
  env?: string;
  status?: number;
  agentStatus?: string;
  groupId?: number;
}

export interface ServerStats {
  total: number;
  online: number;
  offline: number;
  warning: number;
  healthy: number;
}

export interface ConnectInfo {
  serverId: number;
  hostname: string;
  ip: string;
  sshPort: number;
  username?: string;
  password?: string;
  privateKey?: string;
}

export type EnvType = 'prod' | 'test' | 'dev';
export type AgentStatus = 'running' | 'offline' | 'uninstalled';
export type ServerStatus = 1 | 0;
