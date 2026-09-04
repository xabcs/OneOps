import { alova } from './request';

/**
 * alova 演示页 API（最小恢复版本）
 *
 * 该模块此前随 service-alova 目录被误删导致构建失败，此处仅保证
 * src/views/alova 下的演示页面可编译运行，接口指向不存在的后端路径时按请求失败处理。
 */

/** 用户表单模型 */
export interface UserModel {
  userName: string;
  userGender: null | number;
  nickName: string;
  userPhone: string;
  userEmail: string;
  userRoles: string[];
  status: null | number;
}

/** 分页用户列表响应 */
interface UserListResponse {
  records: Api.SystemManage.User[];
  current: number;
  size: number;
  total: number;
}

export function fetchGetUserList(params: Partial<UserModel> & { current: number; size: number }) {
  return alova.Get<UserListResponse>('/user/list', { params });
}

export function fetchGetAllRoles() {
  return alova.Get<Array<{ roleCode: string; roleName: string }>>('/user/getAllRoles');
}

export function addUser(data: UserModel) {
  return alova.Post<null>('/user/add', data);
}

export function updateUser(data: UserModel) {
  return alova.Put<null>('/user/update', data);
}

export function deleteUser(id: number) {
  return alova.Delete<null>(`/user/delete/${id}`);
}

export function batchDeleteUser(ids: number[]) {
  return alova.Post<null>('/user/batch-delete', { ids });
}

export function sendCaptcha(phone: string) {
  return alova.Get<null>('/captcha/send', { params: { phone } });
}

export function verifyCaptcha(phone: string, code: string) {
  return alova.Post<null>('/captcha/verify', { phone, code });
}

/** 演示自定义后端错误（请求拦截层消息演示用） */
export function fetchCustomBackendError(code: string, msg: string) {
  return alova.Get<null>(`/error/${code}`, { params: { msg } });
}
