/**
 * 角色表格列定义
 */
import { ElButton, ElLink, ElSwitch, ElTag } from 'element-plus';
import { $t } from '@/locales';
import { getRoleTagType } from '../../user/modules/user-columns';

export function createRoleColumns(handlers: {
  edit: (id: number) => void;
  handleViewDetail: (row: Api.SystemManage.Role) => void;
  handleAssignPermissions: (id: number) => void;
  handleDelete: (id: number) => void;
  handleStatusChange: (row: Api.SystemManage.Role, val: number) => void;
}) {
  return () => [
    { prop: 'selection', type: 'selection', width: 48 },
    { prop: 'index', type: 'index', label: $t('common.index'), width: 64 },
    {
      prop: 'name',
      label: $t('page.manage.role.roleName'),
      minWidth: 120,
      formatter: (row: Api.SystemManage.Role) => (
        <ElLink type="primary" underline onClick={() => handlers.handleViewDetail(row)}>
          {row.name}
        </ElLink>
      )
    },
    { prop: 'code', label: $t('page.manage.role.roleCode'), minWidth: 120 },
    { prop: 'description', label: $t('page.manage.role.roleDesc'), minWidth: 120 },
    {
      prop: 'users',
      label: '绑定用户',
      minWidth: 200,
      formatter: (row: Api.SystemManage.Role) => {
        if (!row.users || row.users.length === 0) {
          return <span class="pl-12px text-gray">-</span>;
        }
        const userTags = row.users.map(u => (
          <ElTag key={u.id} size="small" type={getRoleTagType(u.username)}>
            {u.username}
          </ElTag>
        ));
        return <div class="flex flex-wrap items-center gap-4px pl-12px">{userTags}</div>;
      }
    },
    {
      prop: 'status',
      label: $t('page.manage.role.roleStatus'),
      align: 'center',
      width: 100,
      formatter: (row: Api.SystemManage.Role) => {
        if (row.status === undefined) {
          return '';
        }

        return (
          <ElSwitch
            v-model={row.status}
            activeValue={1}
            inactiveValue={0}
            disabled={row.code === 'admin'}
            onChange={(val: string | number | boolean) => handlers.handleStatusChange(row, Number(val))}
          />
        );
      }
    },
    {
      prop: 'operate',
      label: $t('common.operate'),
      className: 'msre-table-actions',
      align: 'center',
      width: 260,
      formatter: (row: Api.SystemManage.Role) => (
        <div>
          <ElButton link type="primary" size="small" onClick={() => handlers.edit(row.id)}>
            {$t('common.edit')}
          </ElButton>
          <ElButton
            link
            type="primary"
            size="small"
            disabled={row.code === 'admin'}
            title={row.code === 'admin' ? '超级管理员默认拥有全部权限，无需分配' : undefined}
            onClick={() => handlers.handleAssignPermissions(row.id)}
          >
            分配权限
          </ElButton>
          <ElButton
            link
            type="danger"
            size="small"
            disabled={row.code === 'admin'}
            onClick={() => handlers.handleDelete(row.id)}
          >
            {$t('common.delete')}
          </ElButton>
        </div>
      )
    }
  ];
}
