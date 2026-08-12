/**
 * 角色表格列定义
 */
import { $t } from '@/locales';

export function createRoleColumns(handlers: {
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
          <ElButton type="danger" plain size="small" onClick={() => handlers.handleDelete(row.id)}>
            {$t('common.delete')}
          </ElButton>
        </div>
      )
    }
  ];
}
