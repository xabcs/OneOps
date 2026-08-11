/**
 * 角色表格列定义
 */
import { $t } from '@/locales';
import { canDeleteRole } from './role-helper';

export function createRoleColumns(handlers: {
  roleUsersMap: () => Map<number, string[]>;
  edit: (id: number) => void;
  handleViewPermissions: (id: number) => void;
  handleAssignPermissions: (id: number) => void;
  handleDelete: (id: number) => void;
  handleStatusChange: (row: Api.SystemManage.Role, val: number) => void;
}) {
  return () => [
    { prop: 'selection', type: 'selection', width: 48 },
    { prop: 'index', type: 'index', label: $t('common.index'), width: 64 },
    { prop: 'name', label: $t('page.manage.role.roleName'), minWidth: 120 },
    { prop: 'code', label: $t('page.manage.role.roleCode'), minWidth: 120 },
    { prop: 'description', label: $t('page.manage.role.roleDesc'), minWidth: 120 },
    {
      prop: 'users',
      label: '关联用户',
      minWidth: 150,
      formatter: (row: Api.SystemManage.Role) => {
        const users = handlers.roleUsersMap().get(row.id);
        if (!users || users.length === 0) {
          return <span class="pl-12px text-gray">-</span>;
        }

        const displayUsers = users.slice(0, 3);
        const userCount = users.length;
        const moreText = userCount > 3 ? `等${userCount}人` : '';

        return (
          <div class="flex flex-wrap items-center gap-4px pl-12px">
            {displayUsers.map(username => (
              <ElTag key={username} size="small" type="info">
                {username}
              </ElTag>
            ))}
            {moreText && <span class="text-12px text-gray">{moreText}</span>}
          </div>
        );
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
      prop: 'canDelete',
      label: '状态',
      align: 'center',
      width: 120,
      formatter: (row: Api.SystemManage.Role) => {
        const { canDelete, reason } = canDeleteRole(row, handlers.roleUsersMap());

        if (!canDelete) {
          let tagType: UI.ThemeColor = 'danger';
          let statusText = '不可删除';

          if (reason?.includes('系统内置')) {
            statusText = '系统内置';
            tagType = 'warning';
          } else if (reason?.includes('已关联')) {
            statusText = '已关联用户';
            tagType = 'info';
          }

          return (
            <ElTag size="small" type={tagType}>
              {statusText}
            </ElTag>
          );
        }

        return (
          <ElTag size="small" type="success">
            可删除
          </ElTag>
        );
      }
    },
    {
      prop: 'operate',
      label: $t('common.operate'),
      align: 'center',
      width: 260,
      formatter: (row: Api.SystemManage.Role) => (
        <div class="flex-center gap-8px">
          <ElButton type="primary" plain size="small" onClick={() => handlers.edit(row.id)}>
            {$t('common.edit')}
          </ElButton>
          <ElButton type="info" plain size="small" onClick={() => handlers.handleViewPermissions(row.id)}>
            查看权限
          </ElButton>
          <ElButton type="success" plain size="small" onClick={() => handlers.handleAssignPermissions(row.id)}>
            分配权限
          </ElButton>
          <ElPopconfirm title={$t('common.confirmDelete')} onConfirm={() => handlers.handleDelete(row.id)}>
            {{
              reference: () => (
                <ElButton type="danger" plain size="small">
                  {$t('common.delete')}
                </ElButton>
              )
            }}
          </ElPopconfirm>
        </div>
      )
    }
  ];
}
