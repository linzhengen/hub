import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { roleService, Role, CreateRoleRequest, UpdateRoleRequest } from '@/services/role';
import { permissionService, Permission } from '@/services/permission';
import { Button, Modal, Input, Table, Form, Space } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, KeyOutlined } from '@ant-design/icons';
import { toast } from 'sonner';

export function Roles() {
  const queryClient = useQueryClient();
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [editingRole, setEditingRole] = useState<Role | null>(null);
  const [managingPermissionsRole, setManagingPermissionsRole] = useState<Role | null>(null);
  const [createForm] = Form.useForm();
  const [editForm] = Form.useForm();

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
      render: (_: any, record: Role) => (
        <span>{record.permissionIds?.length || 0} permission(s)</span>
      ),
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
          />
          <Button
            type="text"
            icon={<DeleteOutlined />}
            danger
            onClick={() => {
              if (confirm('Are you sure you want to delete this role?')) {
                deleteMutation.mutate(record.id);
              }
            }}
          />
        </Space>
      ),
    },
  ];

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold tracking-tight">Roles</h2>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => setIsCreateOpen(true)}
        >
          Add Role
        </Button>
      </div>

      <Table
        columns={columns}
        dataSource={data?.roles}
        loading={isLoading}
        rowKey="id"
        locale={{
          emptyText: error ? 'Failed to load roles' : 'No roles found'
        }}
      />

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
