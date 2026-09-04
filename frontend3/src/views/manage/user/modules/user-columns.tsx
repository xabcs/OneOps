/**
 * 用户表格列定义 & 角色颜色辅助函数
 */
// tsx 文件不受 unplugin-vue-components 自动注册覆盖（仅 .vue），必须显式 import 组件，
// 否则 JSX 编译降级为 resolveComponent("ElTag")，运行时解析不到会渲染成无样式的原生元素
import { ElButton, ElTag } from 'element-plus';

// 预定义颜色列表
const COLOR_PALETTE: Array<{ type: UI.ThemeColor; customClass?: string }> = [
  { type: 'primary' },
  { type: 'success' },
  { type: 'warning' },
  { type: 'danger' },
  { type: 'info' }
];

/** 字符串哈希函数 */
function stringHash(str: string): number {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i);
    hash = (hash << 5) - hash + char;
    hash &= hash;
  }
  return Math.abs(hash);
}

/** 根据角色代码获取标签颜色 */
export function getRoleTagType(roleCode: string): UI.ThemeColor {
  const hash = stringHash(roleCode);
  const colorIndex = hash % COLOR_PALETTE.length;
  return COLOR_PALETTE[colorIndex].type;
}

/** 获取角色颜色的自定义样式类 */
export function getRoleTagClass(roleCode: string): string {
  const hash = stringHash(roleCode);
  return `role-tag-${hash % 10}`;
}

/** 创建用户表格列 */
export function createUserColumns(handlers: {
  roleMap: () => Map<number, Api.SystemManage.AllRole> | undefined;
  edit: (id: number) => void;
  openResetPassword: (row: Api.SystemManage.User) => void;
  handleDelete: (id: number) => void;
}) {
  return () => [
    { prop: 'selection', type: 'selection', width: 48 },
    { prop: 'index', type: 'index', label: '序号', width: 64 },
    { prop: 'username', label: '用户名', minWidth: 100 },
    { prop: 'nickname', label: '昵称', minWidth: 100 },
    { prop: 'email', label: '邮箱', minWidth: 200 },
    { prop: 'phone', label: '手机号', width: 130 },
    {
      prop: 'roleIds',
      label: '分配角色',
      minWidth: 150,
      formatter: (row: Api.SystemManage.User) => {
        if (!row.roleIds || row.roleIds.length === 0) {
          return <span class="pl-12px text-gray">-</span>;
        }

        const roleTags = row.roleIds
          .map((id: number) => {
            const role = handlers.roleMap()?.get(id);
            if (!role) return null;

            const tagType = getRoleTagType(role.code);

            return (
              <ElTag key={id} size="small" type={tagType}>
                {role.name}
              </ElTag>
            );
          })
          .filter(Boolean);

        if (roleTags.length === 0) {
          return <span class="pl-12px text-gray">-</span>;
        }

        return <div class="flex flex-wrap items-center gap-4px pl-12px">{roleTags}</div>;
      }
    },
    {
      prop: 'status',
      label: '状态',
      align: 'center',
      width: 100,
      formatter: (row: Api.SystemManage.User) => {
        if (row.status === undefined) {
          return '';
        }

        // 状态值约定：'1' 表示启用；'2' 表示禁用
        const statusMap: Record<string, UI.ThemeColor> = {
          '1': 'success',
          '2': 'warning'
        };

        const label = row.status === '1' ? '启用' : '禁用';

        return <ElTag type={statusMap[row.status] || 'info'}>{label}</ElTag>;
      }
    },
    {
      prop: 'homePath',
      label: '家目录',
      align: 'center',
      width: 150,
      formatter: (row: Api.SystemManage.User) => {
        const path = row.homePath || '/';
        const pathMap: Record<string, string> = {
          '/': '首页',
          '/servers': '资产管理',
          '/tasks': '自动化任务',
          '/monitoring': '监控中心',
          '/system': '系统管理'
        };
        return <span class="text-primary">{pathMap[path] || path}</span>;
      }
    },
    {
      prop: 'operate',
      label: '操作',
      align: 'center',
      width: 260,
      formatter: (row: Api.SystemManage.User) => (
        <div class="flex-center gap-8px">
          <ElButton type="primary" plain size="small" onClick={() => handlers.edit(row.id)}>
            编辑
          </ElButton>
          <ElButton type="warning" plain size="small" onClick={() => handlers.openResetPassword(row)}>
            重置密码
          </ElButton>
          <ElButton
            type="danger"
            plain
            size="small"
            disabled={row.username === 'admin'}
            onClick={() => handlers.handleDelete(row.id)}
          >
            删除
          </ElButton>
        </div>
      )
    }
  ];
}
