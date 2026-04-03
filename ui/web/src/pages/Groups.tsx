import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { groupService, Group, CreateGroupRequest, UpdateGroupRequest, GroupStatus } from '@/services/group';
import { roleService, Role } from '@/services/role';
import { userService, User } from '@/services/user';
import { Button, Modal, Input, Table, Form, Select, Tag, Space } from 'antd';

const { Option } = Select;
import { PlusOutlined, EditOutlined, DeleteOutlined, KeyOutlined, UserOutlined } from '@ant-design/icons';
import { toast } from 'sonner';

export function Groups() {
  const queryClient = useQueryClient();
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [editingGroup, setEditingGroup] = useState<Group | null>(null);
  const [managingRolesGroup, setManagingRolesGroup] = useState<Group | null>(null);
  const [managingUsersGroup, setManagingUsersGroup] = useState<Group | null>(null);
  const [createForm] = Form.useForm();
  const [editForm] = Form.useForm();

  const { data, isLoading, error } = useQuery({
    queryKey: ['groups'],
    queryFn: () => groupService.listGroups(),
  });

  const { data: rolesData } = useQuery({
    queryKey: ['roles'],
    queryFn: () => roleService.listRoles(),
  });

  const { data: usersData } = useQuery({
    queryKey: ['users'],
    queryFn: () => userService.listUsers(),
  });

  const createMutation = useMutation({
    mutationFn: groupService.createGroup,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['groups'] });
      setIsCreateOpen(false);
      toast.success('Group created successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateGroupRequest }) => groupService.updateGroup(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['groups'] });
      setEditingGroup(null);
      toast.success('Group updated successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const deleteMutation = useMutation({
    mutationFn: groupService.deleteGroup,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['groups'] });
      toast.success('Group deleted successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const assignRoleMutation = useMutation({
    mutationFn: ({ id, roleId }: { id: string; roleId: string }) => groupService.assignRole(id, { roleId }),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['groups'] });
      // Update managingRolesGroup state immediately
      if (managingRolesGroup && managingRolesGroup.id === variables.id) {
        setManagingRolesGroup({
          ...managingRolesGroup,
          roleIds: [...managingRolesGroup.roleIds, variables.roleId]
        });
      }
      toast.success('Role assigned successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const unassignRoleMutation = useMutation({
    mutationFn: ({ id, roleId, currentRoleIds }: { id: string; roleId: string; currentRoleIds: string[] }) =>
      groupService.assignRolesToGroup(id, { roleIds: currentRoleIds.filter(rId => rId !== roleId) }),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['groups'] });
      // Update managingRolesGroup state immediately
      if (managingRolesGroup && managingRolesGroup.id === variables.id) {
        setManagingRolesGroup({
          ...managingRolesGroup,
          roleIds: managingRolesGroup.roleIds.filter(id => id !== variables.roleId)
        });
      }
      toast.success('Role unassigned successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const addUsersToGroupMutation = useMutation({
    mutationFn: ({ id, userIds }: { id: string; userIds: string[] }) => groupService.addUsersToGroup(id, { userIds }),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['groups'] });
      // Update managingUsersGroup state immediately
      if (managingUsersGroup && managingUsersGroup.id === variables.id) {
        // Note: group doesn't have userIds field in response, we'll just invalidate
      }
      toast.success('Users added successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const removeUsersFromGroupMutation = useMutation({
    mutationFn: ({ id, userIds }: { id: string; userIds: string[] }) => groupService.removeUsersFromGroup(id, { userIds }),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['groups'] });
      // Update managingUsersGroup state immediately
      if (managingUsersGroup && managingUsersGroup.id === variables.id) {
        // Note: group doesn't have userIds field in response, we'll just invalidate
      }
      toast.success('Users removed successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const handleCreateSubmit = (values: any) => {
    createMutation.mutate({
      name: values.name,
      description: values.description,
      status: values.status || 'STATUS_UNSPECIFIED',
    });
  };

  const handleEditSubmit = (values: any) => {
    if (!editingGroup) return;

    const data: UpdateGroupRequest = {
      name: values.name,
      description: values.description,
    };

    if (values.status) {
      data.status = values.status;
    }

    updateMutation.mutate({
      id: editingGroup.id,
      data
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
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        let color = 'default';
        let text = 'Unspecified';

        if (status === 'STATUS_ACTIVE') {
          color = 'success';
          text = 'Active';
        } else if (status === 'STATUS_INACTIVE') {
          color = 'error';
          text = 'Inactive';
        }

        return <Tag color={color}>{text}</Tag>;
      },
    },
    {
      title: 'Roles',
      key: 'roles',
      render: (_: any, record: Group) => (
        <span>{record.roleIds?.length || 0} role(s)</span>
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: any, record: Group) => (
        <Space>
          <Button
            type="text"
            icon={<KeyOutlined />}
            onClick={() => setManagingRolesGroup(record)}
          />
          <Button
            type="text"
            icon={<UserOutlined />}
            onClick={() => setManagingUsersGroup(record)}
          />
          <Button
            type="text"
            icon={<EditOutlined />}
            onClick={() => {
              setEditingGroup(record);
              editForm.setFieldsValue({
                name: record.name,
                description: record.description,
                status: record.status,
              });
            }}
          />
          <Button
            type="text"
            icon={<DeleteOutlined />}
            danger
            onClick={() => {
              if (confirm('Are you sure you want to delete this group?')) {
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
      <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between" }}>
        <h2 style={{ fontSize: "24px", fontWeight: "bold", letterSpacing: "-0.025em" }}>Groups</h2>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => setIsCreateOpen(true)}
        >
          Add Group
        </Button>
      </div>

      <Table
        columns={columns}
        dataSource={data?.groups}
        loading={isLoading}
        rowKey="id"
        locale={{
          emptyText: error ? 'Failed to load groups' : 'No groups found'
        }}
      />

      <Modal
        title="Create New Group"
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
            rules={[{ required: true, message: 'Please input group name!' }]}
          >
            <Input />
          </Form.Item>
          <Form.Item
            name="description"
            label="Description"
          >
            <Input />
          </Form.Item>
          <Form.Item
            name="status"
            label="Status"
            initialValue="STATUS_UNSPECIFIED"
          >
            <Select>
              <Option value="STATUS_UNSPECIFIED">Unspecified</Option>
              <Option value="STATUS_ACTIVE">Active</Option>
              <Option value="STATUS_INACTIVE">Inactive</Option>
            </Select>
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
        title="Edit Group"
        open={!!editingGroup}
        onCancel={() => setEditingGroup(null)}
        footer={null}
      >
        {editingGroup && (
          <Form
            form={editForm}
            layout="vertical"
            onFinish={handleEditSubmit}
          >
            <Form.Item
              name="name"
              label="Name"
              rules={[{ required: true, message: 'Please input group name!' }]}
            >
              <Input />
            </Form.Item>
            <Form.Item
              name="description"
              label="Description"
            >
              <Input />
            </Form.Item>
            <Form.Item
              name="status"
              label="Status"
            >
              <Select>
                <Option value="STATUS_UNSPECIFIED">Unspecified</Option>
                <Option value="STATUS_ACTIVE">Active</Option>
                <Option value="STATUS_INACTIVE">Inactive</Option>
              </Select>
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
        title={`Manage Roles for ${managingRolesGroup?.name}`}
        open={!!managingRolesGroup}
        onCancel={() => setManagingRolesGroup(null)}
        width={800}
        footer={null}
      >
        {managingRolesGroup && (
          <div className="space-y-6">
            <div>
              <h3 className="text-lg font-medium">Current Roles</h3>
              {managingRolesGroup.roleIds.length === 0 ? (
                <p className="text-sm text-gray-500">No roles assigned</p>
              ) : (
                <div className="mt-2 space-y-2">
                  {managingRolesGroup.roleIds.map((roleId) => {
                    const role = rolesData?.roles?.find(r => r.id === roleId);
                    return role ? (
                      <div key={roleId} className="flex items-center justify-between rounded-md border px-3 py-2">
                        <span>{role.name}</span>
                        <Button
                          type="text"
                          danger
                          onClick={() => {
                            if (confirm(`Are you sure you want to remove ${role.name} from this group?`)) {
                              unassignRoleMutation.mutate({
                                id: managingRolesGroup.id,
                                roleId: role.id,
                                currentRoleIds: managingRolesGroup.roleIds
                              });
                            }
                          }}
                          loading={unassignRoleMutation.isPending && unassignRoleMutation.variables?.id === managingRolesGroup.id && unassignRoleMutation.variables?.roleId === role.id}
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
              <h3 className="text-lg font-medium">Available Roles</h3>
              {rolesData?.roles?.filter(role => !managingRolesGroup.roleIds.includes(role.id)).length === 0 ? (
                <p className="text-sm text-gray-500">No available roles</p>
              ) : (
                <div className="mt-2 space-y-2">
                  {rolesData?.roles
                    ?.filter(role => !managingRolesGroup.roleIds.includes(role.id))
                    .map((role) => (
                      <div key={role.id} className="flex items-center justify-between rounded-md border px-3 py-2">
                        <span>{role.name}</span>
                        <Button
                          type="text"
                          onClick={() => {
                            assignRoleMutation.mutate({
                              id: managingRolesGroup.id,
                              roleId: role.id
                            });
                          }}
                          loading={assignRoleMutation.isPending && assignRoleMutation.variables?.id === managingRolesGroup.id && assignRoleMutation.variables?.roleId === role.id}
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

      <Modal
        title={`Manage Users for ${managingUsersGroup?.name}`}
        open={!!managingUsersGroup}
        onCancel={() => setManagingUsersGroup(null)}
        width={800}
        footer={null}
      >
        {managingUsersGroup && (
          <div className="space-y-6">
            <div>
              <h3 className="text-lg font-medium">Current Users</h3>
              <p className="text-sm text-gray-500">User list not available in group response</p>
            </div>

            <div>
              <h3 className="text-lg font-medium">Add Users</h3>
              <div className="mt-2 space-y-2">
                {usersData?.users?.map((user) => (
                  <div key={user.id} className="flex items-center justify-between rounded-md border px-3 py-2">
                    <span>{user.username} ({user.email})</span>
                    <Button
                      type="text"
                      onClick={() => {
                        addUsersToGroupMutation.mutate({
                          id: managingUsersGroup.id,
                          userIds: [user.id]
                        });
                      }}
                      loading={addUsersToGroupMutation.isPending && addUsersToGroupMutation.variables?.id === managingUsersGroup.id && addUsersToGroupMutation.variables?.userIds?.includes(user.id)}
                    >
                      Add
                    </Button>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
}
