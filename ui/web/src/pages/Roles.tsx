import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { roleService, Role, CreateRoleRequest, UpdateRoleRequest } from '@/services/role';
import { permissionService, Permission } from '@/services/permission';
import { Button, Modal, Input, Table, Form, Space, Card } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, KeyOutlined, SearchOutlined, SafetyOutlined } from '@ant-design/icons';
import { toast } from 'sonner';
import { Shield, TrendingUp, Lock } from 'lucide-react';

export function Roles() {
  const queryClient = useQueryClient();
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [editingRole, setEditingRole] = useState<Role | null>(null);
  const [managingPermissionsRole, setManagingPermissionsRole] = useState<Role | null>(null);
  const [createForm] = Form.useForm();
  const [editForm] = Form.useForm();
  const [searchText, setSearchText] = useState('');

  const { data, isLoading, error } = useQuery({
    queryKey: ['roles'],
    queryFn: () => roleService.listRoles(),
  });

  const { data: permissionsData } = useQuery({
    queryKey: ['permissions'],
    queryFn: () => permissionService.listPermissions(),
  });

  const createMutation = useMutation({
    mutationFn: roleService.createRole,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] });
      setIsCreateOpen(false);
      toast.success('Role created successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateRoleRequest }) => roleService.updateRole(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] });
      setEditingRole(null);
      toast.success('Role updated successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const deleteMutation = useMutation({
    mutationFn: roleService.deleteRole,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] });
      toast.success('Role deleted successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const assignPermissionMutation = useMutation({
    mutationFn: ({ id, permissionId }: { id: string; permissionId: string }) => roleService.assignPermission(id, { permissionId }),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['roles'] });
      // Update managingPermissionsRole state immediately
      if (managingPermissionsRole && managingPermissionsRole.id === variables.id) {
        setManagingPermissionsRole({
          ...managingPermissionsRole,
          permissionIds: [...managingPermissionsRole.permissionIds, variables.permissionId]
        });
      }
      toast.success('Permission assigned successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const unassignPermissionMutation = useMutation({
    mutationFn: ({ id, permissionId, currentPermissionIds }: { id: string; permissionId: string; currentPermissionIds: string[] }) =>
      roleService.removePermissionsFromRole(id, { permissionIds: currentPermissionIds.filter(pId => pId !== permissionId) }),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['roles'] });
      // Update managingPermissionsRole state immediately
      if (managingPermissionsRole && managingPermissionsRole.id === variables.id) {
        setManagingPermissionsRole({
          ...managingPermissionsRole,
          permissionIds: managingPermissionsRole.permissionIds.filter(id => id !== variables.permissionId)
        });
      }
      toast.success('Permission unassigned successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const handleCreateSubmit = (values: any) => {
    createMutation.mutate({
      name: values.name,
      description: values.description,
    });
  };

  const handleEditSubmit = (values: any) => {
    if (!editingRole) return;

    updateMutation.mutate({
      id: editingRole.id,
      data: {
        name: values.name,
        description: values.description,
      }
    });
  };

  const columns = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: 'Description',
      dataIndex: 'description',
      key: 'description',
    },
    {
      title: 'Permissions',
      key: 'permissions',
      render: (_: any, record: Role) => {
        const permissionCount = record.permissionIds?.length || 0;
        return (
          <div className="flex items-center gap-2">
            <span className="inline-flex items-center px-2.5 py-0.5 rounded-md text-xs font-medium bg-emerald-100 text-emerald-800 dark:bg-emerald-500/10 dark:text-emerald-400 border border-emerald-200 dark:border-emerald-500/20">
              {permissionCount} {permissionCount === 1 ? 'permission' : 'permissions'}
            </span>
            {permissionCount > 0 && permissionsData?.permissions && (
              <span className="text-sm text-gray-500 dark:text-gray-400">
                {permissionsData.permissions.filter(perm => record.permissionIds?.includes(perm.id)).slice(0, 2).map(p => p.name).join(', ')}
                {permissionCount > 2 ? '...' : ''}
              </span>
            )}
          </div>
        );
      },
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: any, record: Role) => (
        <Space>
          <Button
            type="text"
            icon={<KeyOutlined />}
            onClick={() => setManagingPermissionsRole(record)}
            className="p-1.5 rounded-md text-purple-600 dark:text-purple-400 hover:bg-purple-50 dark:hover:bg-purple-900/20 transition-colors"
            title="Manage Permissions"
          />
          <Button
            type="text"
            icon={<EditOutlined />}
            onClick={() => {
              setEditingRole(record);
              editForm.setFieldsValue({
                name: record.name,
                description: record.description,
              });
            }}
            className="p-1.5 rounded-md text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/20 transition-colors"
            title="Edit Role"
          />
          <Button
            type="text"
            icon={<DeleteOutlined />}
            onClick={() => {
              if (confirm('Are you sure you want to delete this role?')) {
                deleteMutation.mutate(record.id);
              }
            }}
            className="p-1.5 rounded-md text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors"
            title="Delete Role"
          />
        </Space>
      ),
    },
  ];

  // 検索フィルター
  const filteredRoles = data?.roles?.filter(role =>
    !searchText ||
    role.name.toLowerCase().includes(searchText.toLowerCase()) ||
    (role.description && role.description.toLowerCase().includes(searchText.toLowerCase()))
  );

  // 統計データの計算
  const totalRoles = data?.roles?.length || 0;
  const averagePermissionsPerRole = data?.roles?.length ?
    (data.roles.reduce((acc, role) => acc + (role.permissionIds?.length || 0), 0) / data.roles.length).toFixed(1)
    : '0.0';
  const totalPermissions = permissionsData?.permissions?.length || 0;

  return (
    <div className="space-y-6">
      {/* ヘッダーセクション */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold tracking-tight text-gray-900 dark:text-white">Roles</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400">Define and manage role-based access permissions</p>
        </div>
        <div className="flex items-center gap-3">
          <Input
            placeholder="Search roles..."
            prefix={<SearchOutlined className="text-gray-400 dark:text-gray-500" />}
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            style={{ width: 250, borderRadius: '8px' }}
            allowClear
          />
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => setIsCreateOpen(true)}
            style={{ borderRadius: '8px' }}
          >
            Add Role
          </Button>
        </div>
      </div>

      {/* 統計カード */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card className="shadow-sm dark:bg-gray-800 dark:border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm font-medium text-gray-500 dark:text-gray-400">Total Roles</div>
              <div className="text-2xl font-bold mt-1 text-gray-900 dark:text-white">{totalRoles}</div>
              <div className="flex items-center gap-1 mt-2">
                <TrendingUp className="h-4 w-4 text-green-500 dark:text-green-400" />
                <span className="text-sm text-green-600 dark:text-green-400">+5.4%</span>
                <span className="text-sm text-gray-500 dark:text-gray-400">from last month</span>
              </div>
            </div>
            <div className="p-2 rounded-lg bg-purple-50 dark:bg-purple-900/20">
              <Shield className="h-5 w-5 text-purple-600 dark:text-purple-400" />
            </div>
          </div>
        </Card>
        <Card className="shadow-sm dark:bg-gray-800 dark:border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm font-medium text-gray-500 dark:text-gray-400">Avg. Permissions/Role</div>
              <div className="text-2xl font-bold mt-1 text-gray-900 dark:text-white">{averagePermissionsPerRole}</div>
              <div className="text-sm mt-2 text-gray-500 dark:text-gray-400">
                {totalPermissions} total permissions
              </div>
            </div>
            <div className="p-2 rounded-lg bg-blue-50 dark:bg-blue-900/20">
              <KeyOutlined style={{ fontSize: '20px' }} className="text-blue-600 dark:text-blue-400" />
            </div>
          </div>
        </Card>
        <Card className="shadow-sm dark:bg-gray-800 dark:border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm font-medium text-gray-500 dark:text-gray-400">System Roles</div>
              <div className="text-2xl font-bold mt-1 text-gray-900 dark:text-white">
                {data?.roles?.filter(r => r.name.toLowerCase().includes('admin') || r.name.toLowerCase().includes('system')).length || 0}
              </div>
              <div className="text-sm mt-2 text-gray-500 dark:text-gray-400">
                Administrative roles
              </div>
            </div>
            <div className="p-2 rounded-lg bg-red-50 dark:bg-red-900/20">
              <SafetyOutlined style={{ fontSize: '20px' }} className="text-red-600 dark:text-red-400" />
            </div>
          </div>
        </Card>
        <Card className="shadow-sm dark:bg-gray-800 dark:border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm font-medium text-gray-500 dark:text-gray-400">Custom Roles</div>
              <div className="text-2xl font-bold mt-1 text-gray-900 dark:text-white">
                {data?.roles?.filter(r => !r.name.toLowerCase().includes('admin') && !r.name.toLowerCase().includes('system')).length || 0}
              </div>
              <div className="text-sm mt-2 text-gray-500 dark:text-gray-400">
                User-defined roles
              </div>
            </div>
            <div className="p-2 rounded-lg bg-green-50 dark:bg-green-900/20">
              <Lock className="h-5 w-5 text-green-600 dark:text-green-400" />
            </div>
          </div>
        </Card>
      </div>

      {/* ロールテーブル */}
      <Card className="shadow-sm dark:bg-gray-800 dark:border-gray-700">
        <Table
          columns={columns}
          dataSource={filteredRoles}
          loading={isLoading}
          rowKey="id"
          locale={{
            emptyText: error ? 'Failed to load roles' : 'No roles found'
          }}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => `${range[0]}-${range[1]} of ${total} roles`,
          }}
        />
      </Card>

      <Modal
        title="Create New Role"
        open={isCreateOpen}
        onCancel={() => setIsCreateOpen(false)}
        footer={null}
      >
        <Form
          form={createForm}
          layout="vertical"
          onFinish={handleCreateSubmit}
        >
          <Form.Item
            name="name"
            label="Name"
            rules={[{ required: true, message: 'Please input role name!' }]}
          >
            <Input />
          </Form.Item>
          <Form.Item
            name="description"
            label="Description"
          >
            <Input />
          </Form.Item>
          <Form.Item className="mb-0">
            <div className="flex justify-end">
              <Button
                type="primary"
                htmlType="submit"
                loading={createMutation.isPending}
              >
                Create
              </Button>
            </div>
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="Edit Role"
        open={!!editingRole}
        onCancel={() => setEditingRole(null)}
        footer={null}
      >
        {editingRole && (
          <Form
            form={editForm}
            layout="vertical"
            onFinish={handleEditSubmit}
          >
            <Form.Item
              name="name"
              label="Name"
              rules={[{ required: true, message: 'Please input role name!' }]}
            >
              <Input />
            </Form.Item>
            <Form.Item
              name="description"
              label="Description"
            >
              <Input />
            </Form.Item>
            <Form.Item className="mb-0">
              <div className="flex justify-end">
                <Button
                  type="primary"
                  htmlType="submit"
                  loading={updateMutation.isPending}
                >
                  Save Changes
                </Button>
              </div>
            </Form.Item>
          </Form>
        )}
      </Modal>

      <Modal
        title={`Manage Permissions for ${managingPermissionsRole?.name}`}
        open={!!managingPermissionsRole}
        onCancel={() => setManagingPermissionsRole(null)}
        width={800}
        footer={null}
      >
        {managingPermissionsRole && (
          <div className="space-y-6">
            <div>
              <h3 className="text-lg font-medium">Current Permissions</h3>
              {managingPermissionsRole.permissionIds.length === 0 ? (
                <p className="text-sm text-gray-500">No permissions assigned</p>
              ) : (
                <div className="mt-2 space-y-2">
                  {managingPermissionsRole.permissionIds.map((permissionId) => {
                    const permission = permissionsData?.permissions?.find(p => p.id === permissionId);
                    return permission ? (
                      <div key={permissionId} className="flex items-center justify-between rounded-md border px-3 py-2">
                        <span>{permission.name} ({permission.verb} on {permission.resourceId})</span>
                        <Button
                          type="text"
                          danger
                          onClick={() => {
                            if (confirm(`Are you sure you want to remove ${permission.name} from this role?`)) {
                              unassignPermissionMutation.mutate({
                                id: managingPermissionsRole.id,
                                permissionId: permission.id,
                                currentPermissionIds: managingPermissionsRole.permissionIds
                              });
                            }
                          }}
                          loading={unassignPermissionMutation.isPending && unassignPermissionMutation.variables?.id === managingPermissionsRole.id && unassignPermissionMutation.variables?.permissionId === permission.id}
                        >
                          Remove
                        </Button>
                      </div>
                    ) : null;
                  })}
                </div>
              )}
            </div>

            <div>
              <h3 className="text-lg font-medium">Available Permissions</h3>
              {permissionsData?.permissions?.filter(permission => !managingPermissionsRole.permissionIds.includes(permission.id)).length === 0 ? (
                <p className="text-sm text-gray-500">No available permissions</p>
              ) : (
                <div className="mt-2 space-y-2">
                  {permissionsData?.permissions
                    ?.filter(permission => !managingPermissionsRole.permissionIds.includes(permission.id))
                    .map((permission) => (
                      <div key={permission.id} className="flex items-center justify-between rounded-md border px-3 py-2">
                        <span>{permission.name} ({permission.verb} on {permission.resourceId})</span>
                        <Button
                          type="text"
                          onClick={() => {
                            assignPermissionMutation.mutate({
                              id: managingPermissionsRole.id,
                              permissionId: permission.id
                            });
                          }}
                          loading={assignPermissionMutation.isPending && assignPermissionMutation.variables?.id === managingPermissionsRole.id && assignPermissionMutation.variables?.permissionId === permission.id}
                        >
                          Assign
                        </Button>
                      </div>
                    ))}
                </div>
              )}
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
}
