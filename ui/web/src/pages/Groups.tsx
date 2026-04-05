import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { groupService, Group, CreateGroupRequest, UpdateGroupRequest, GroupStatus } from '@/services/group';
import { roleService, Role } from '@/services/role';
import { userService, User } from '@/services/user';
import { Button, Modal, Input, Table, Form, Select, Tag, Space, Card } from 'antd';

const { Option } = Select;
import { PlusOutlined, EditOutlined, DeleteOutlined, KeyOutlined, UserOutlined, SearchOutlined, FolderOutlined } from '@ant-design/icons';
import { toast } from 'sonner';
import { Users, Shield, TrendingUp } from 'lucide-react';

export function Groups() {
  const queryClient = useQueryClient();
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [editingGroup, setEditingGroup] = useState<Group | null>(null);
  const [managingRolesGroup, setManagingRolesGroup] = useState<Group | null>(null);
  const [managingUsersGroup, setManagingUsersGroup] = useState<Group | null>(null);
  const [createForm] = Form.useForm();
  const [editForm] = Form.useForm();
  const [searchText, setSearchText] = useState('');

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
        let color = '';
        let bgColor = '';
        let text = 'Unspecified';

        if (status === 'STATUS_ACTIVE') {
          color = '#059669';
          bgColor = '#d1fae5';
          text = 'Active';
        } else if (status === 'STATUS_INACTIVE') {
          color = '#dc2626';
          bgColor = '#fee2e2';
          text = 'Inactive';
        } else {
          color = '#6b7280';
          bgColor = '#f3f4f6';
          text = 'Unspecified';
        }

        return (
          <span style={{
            display: 'inline-flex',
            alignItems: 'center',
            padding: '4px 10px',
            borderRadius: '9999px',
            fontSize: '12px',
            fontWeight: '500',
            color: color,
            backgroundColor: bgColor
          }}>
            {text}
          </span>
        );
      },
    },
    {
      title: 'Roles',
      key: 'roles',
      render: (_: any, record: Group) => {
        const roleCount = record.roleIds?.length || 0;
        return (
          <div className="flex items-center gap-2">
            <span style={{
              display: 'inline-flex',
              alignItems: 'center',
              padding: '4px 10px',
              borderRadius: '6px',
              fontSize: '12px',
              fontWeight: '500',
              color: '#7c3aed',
              backgroundColor: '#f3e8ff',
              border: '1px solid #e9d5ff'
            }}>
              {roleCount} {roleCount === 1 ? 'role' : 'roles'}
            </span>
            {roleCount > 0 && rolesData?.roles && (
              <span className="text-sm" style={{ color: '#64748b' }}>
                {rolesData.roles.filter(role => record.roleIds?.includes(role.id)).slice(0, 2).map(r => r.name).join(', ')}
                {roleCount > 2 ? '...' : ''}
              </span>
            )}
          </div>
        );
      },
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
            style={{
              padding: '6px',
              borderRadius: '6px',
              color: '#8b5cf6'
            }}
            className="hover:bg-purple-50"
            title="Manage Roles"
          />
          <Button
            type="text"
            icon={<UserOutlined />}
            onClick={() => setManagingUsersGroup(record)}
            style={{
              padding: '6px',
              borderRadius: '6px',
              color: '#3b82f6'
            }}
            className="hover:bg-blue-50"
            title="Manage Users"
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
            style={{
              padding: '6px',
              borderRadius: '6px',
              color: '#059669'
            }}
            className="hover:bg-green-50"
            title="Edit Group"
          />
          <Button
            type="text"
            icon={<DeleteOutlined />}
            onClick={() => {
              if (confirm('Are you sure you want to delete this group?')) {
                deleteMutation.mutate(record.id);
              }
            }}
            style={{
              padding: '6px',
              borderRadius: '6px',
              color: '#dc2626'
            }}
            className="hover:bg-red-50"
            title="Delete Group"
          />
        </Space>
      ),
    },
  ];

  // 検索フィルター
  const filteredGroups = data?.groups?.filter(group =>
    !searchText ||
    group.name.toLowerCase().includes(searchText.toLowerCase()) ||
    (group.description && group.description.toLowerCase().includes(searchText.toLowerCase()))
  );

  // 統計データの計算
  const totalGroups = data?.groups?.length || 0;
  const activeGroups = data?.groups?.filter(g => g.status === 'STATUS_ACTIVE').length || 0;
  const averageRolesPerGroup = data?.groups?.length ?
    (data.groups.reduce((acc, group) => acc + (group.roleIds?.length || 0), 0) / data.groups.length).toFixed(1)
    : '0.0';
  const averageUsersPerGroup = data?.groups?.length ?
    (data.groups.reduce((acc, group) => {
      const groupUsers = usersData?.users?.filter(user => user.groupIds?.includes(group.id)).length || 0;
      return acc + groupUsers;
    }, 0) / data.groups.length).toFixed(1)
    : '0.0';

  return (
    <div className="space-y-6">
      {/* ヘッダーセクション */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold tracking-tight" style={{ color: '#1e293b' }}>Groups</h2>
          <p className="text-sm" style={{ color: '#64748b' }}>Manage user groups and assign roles/permissions</p>
        </div>
        <div className="flex items-center gap-3">
          <Input
            placeholder="Search groups..."
            prefix={<SearchOutlined style={{ color: '#94a3b8' }} />}
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
            Add Group
          </Button>
        </div>
      </div>

      {/* 統計カード */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card className="shadow-sm">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm font-medium" style={{ color: '#64748b' }}>Total Groups</div>
              <div className="text-2xl font-bold mt-1" style={{ color: '#1e293b' }}>{totalGroups}</div>
              <div className="flex items-center gap-1 mt-2">
                <TrendingUp className="h-4 w-4 text-green-500" />
                <span className="text-sm text-green-600">+8.2%</span>
                <span className="text-sm text-gray-500">from last month</span>
              </div>
            </div>
            <div className="p-2 rounded-lg bg-blue-50">
              <FolderOutlined style={{ fontSize: '20px', color: '#3b82f6' }} />
            </div>
          </div>
        </Card>
        <Card className="shadow-sm">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm font-medium" style={{ color: '#64748b' }}>Active Groups</div>
              <div className="text-2xl font-bold mt-1" style={{ color: '#1e293b' }}>{activeGroups}</div>
              <div className="text-sm mt-2" style={{ color: '#64748b' }}>
                {totalGroups > 0 ? `${Math.round((activeGroups / totalGroups) * 100)}% active` : 'No groups'}
              </div>
            </div>
            <div className="p-2 rounded-lg bg-green-50">
              <div className="h-5 w-5 rounded-full bg-green-500"></div>
            </div>
          </div>
        </Card>
        <Card className="shadow-sm">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm font-medium" style={{ color: '#64748b' }}>Avg. Roles/Group</div>
              <div className="text-2xl font-bold mt-1" style={{ color: '#1e293b' }}>{averageRolesPerGroup}</div>
              <div className="text-sm mt-2" style={{ color: '#64748b' }}>
                {rolesData?.roles?.length || 0} total roles
              </div>
            </div>
            <div className="p-2 rounded-lg bg-purple-50">
              <Shield className="h-5 w-5 text-purple-600" />
            </div>
          </div>
        </Card>
        <Card className="shadow-sm">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm font-medium" style={{ color: '#64748b' }}>Avg. Users/Group</div>
              <div className="text-2xl font-bold mt-1" style={{ color: '#1e293b' }}>{averageUsersPerGroup}</div>
              <div className="text-sm mt-2" style={{ color: '#64748b' }}>
                {usersData?.users?.length || 0} total users
              </div>
            </div>
            <div className="p-2 rounded-lg bg-orange-50">
              <Users className="h-5 w-5 text-orange-600" />
            </div>
          </div>
        </Card>
      </div>

      {/* グループテーブル */}
      <Card className="shadow-sm">
        <Table
          columns={columns}
          dataSource={filteredGroups}
          loading={isLoading}
          rowKey="id"
          locale={{
            emptyText: error ? 'Failed to load groups' : 'No groups found'
          }}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => `${range[0]}-${range[1]} of ${total} groups`,
          }}
        />
      </Card>

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
