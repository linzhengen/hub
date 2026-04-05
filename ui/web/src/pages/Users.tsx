import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { userService, User } from '@/services/user';
import { groupService, Group, ListGroupsResponse } from '@/services/group';
import { Button, Modal, Input, Table, Form, Select, Tag, Space, Card } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, UserOutlined } from '@ant-design/icons';
import { FolderKanban } from 'lucide-react';

import { toast } from 'sonner';
import { cn } from '@/lib/utils';

const { Option } = Select;

export function Users() {
  const queryClient = useQueryClient();
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [managingGroupsUser, setManagingGroupsUser] = useState<User | null>(null);
  const [createForm] = Form.useForm();
  const [editForm] = Form.useForm();
  const [searchText, setSearchText] = useState('');

  const { data, isLoading, error } = useQuery({
    queryKey: ['users'],
    queryFn: () => userService.listUsers(),
  });

  const { data: groupsData } = useQuery<ListGroupsResponse>({
    queryKey: ['groups'],
    queryFn: () => groupService.listGroups(),
  });

  // Helper function to get group names from group IDs
  const getGroupNames = (groupIds: string[]): string[] => {
    if (!groupsData?.groups) return [];
    return groupIds
      .map(id => groupsData.groups.find(g => g.id === id)?.name)
      .filter((name): name is string => name !== undefined);
  };

  const createMutation = useMutation({
    mutationFn: userService.createUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      setIsCreateOpen(false);
      toast.success('User created successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<User> }) => userService.updateUser(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      setEditingUser(null);
      toast.success('User updated successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const deleteMutation = useMutation({
    mutationFn: userService.deleteUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      toast.success('User deleted successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const assignGroupMutation = useMutation({
    mutationFn: ({ userId, groupId }: { userId: string; groupId: string }) =>
      userService.assignGroup(userId, groupId),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      // Update the managingGroupsUser state to reflect the change immediately
      if (managingGroupsUser && managingGroupsUser.id === variables.userId) {
        setManagingGroupsUser({
          ...managingGroupsUser,
          groupIds: [...managingGroupsUser.groupIds, variables.groupId]
        });
      }
      toast.success('Group assigned successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const unassignGroupMutation = useMutation({
    mutationFn: ({ userId, groupId }: { userId: string; groupId: string }) =>
      userService.unassignGroup(userId, groupId),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      // Update the managingGroupsUser state to reflect the change immediately
      if (managingGroupsUser && managingGroupsUser.id === variables.userId) {
        setManagingGroupsUser({
          ...managingGroupsUser,
          groupIds: managingGroupsUser.groupIds.filter(id => id !== variables.groupId)
        });
      }
      toast.success('Group unassigned successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });


  const columns = [
    {
      title: 'Username',
      dataIndex: 'username',
      key: 'username',
    },
    {
      title: 'Email',
      dataIndex: 'email',
      key: 'email',
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
      title: 'Groups',
      key: 'groups',
      render: (_: any, record: User) => {
        const groupNames = getGroupNames(record.groupIds);
        return (
          <div className="flex flex-wrap gap-1">
            {groupNames.length > 0 ? (
              groupNames.map((name, index) => (
                <span
                  key={index}
                  style={{
                    display: 'inline-flex',
                    alignItems: 'center',
                    padding: '2px 8px',
                    borderRadius: '6px',
                    fontSize: '12px',
                    fontWeight: '400',
                    color: '#3b82f6',
                    backgroundColor: '#eff6ff',
                    border: '1px solid #dbeafe'
                  }}
                >
                  {name}
                </span>
              ))
            ) : (
              <span style={{ color: '#94a3b8', fontSize: '14px' }}>None</span>
            )}
          </div>
        );
      },
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: any, record: User) => (
        <Space>
          <Button
            type="text"
            icon={<UserOutlined />}
            onClick={() => setManagingGroupsUser(record)}
            style={{
              padding: '6px',
              borderRadius: '6px',
              color: '#64748b'
            }}
            className="hover:bg-gray-100"
          />
          <Button
            type="text"
            icon={<EditOutlined />}
            onClick={() => {
              setEditingUser(record);
              editForm.setFieldsValue({
                username: record.username,
                email: record.email,
                status: record.status,
                groupIds: record.groupIds,
              });
            }}
            style={{
              padding: '6px',
              borderRadius: '6px',
              color: '#3b82f6'
            }}
            className="hover:bg-blue-50"
          />
          <Button
            type="text"
            icon={<DeleteOutlined />}
            onClick={() => {
              if (confirm('Are you sure you want to delete this user?')) {
                deleteMutation.mutate(record.id);
              }
            }}
            style={{
              padding: '6px',
              borderRadius: '6px',
              color: '#dc2626'
            }}
            className="hover:bg-red-50"
          />
        </Space>
      ),
    },
  ];

  const handleCreateSubmit = (values: any) => {
    createMutation.mutate({
      username: values.username,
      email: values.email,
      password: values.password,
      groupIds: values.groupIds || [],
    });
  };

  const handleEditSubmit = (values: any) => {
    if (!editingUser) return;

    const data: any = {
      username: values.username,
      email: values.email,
    };

    if (values.password) data.password = values.password;
    if (values.status) data.status = values.status;
    if (values.groupIds) data.groupIds = values.groupIds;

    updateMutation.mutate({
      id: editingUser.id,
      data
    });
  };

  // 検索フィルター
  const filteredUsers = data?.users?.filter(user =>
    !searchText ||
    user.username.toLowerCase().includes(searchText.toLowerCase()) ||
    user.email.toLowerCase().includes(searchText.toLowerCase())
  );

  return (
    <div className="space-y-6">
      {/* ヘッダーセクション */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold tracking-tight" style={{ color: '#1e293b' }}>Users</h2>
          <p className="text-sm" style={{ color: '#64748b' }}>Manage user accounts and permissions</p>
        </div>
        <div className="flex items-center gap-3">
          <Input
            placeholder="Search users..."
            prefix={<UserOutlined style={{ color: '#94a3b8' }} />}
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
            Add User
          </Button>
        </div>
      </div>

      {/* 統計カード */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card className="shadow-sm">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm font-medium" style={{ color: '#64748b' }}>Total Users</div>
              <div className="text-2xl font-bold mt-1" style={{ color: '#1e293b' }}>{data?.users?.length || 0}</div>
            </div>
            <div className="p-2 rounded-lg bg-blue-50">
              <UserOutlined style={{ fontSize: '20px', color: '#3b82f6' }} />
            </div>
          </div>
        </Card>
        <Card className="shadow-sm">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm font-medium" style={{ color: '#64748b' }}>Active Users</div>
              <div className="text-2xl font-bold mt-1" style={{ color: '#1e293b' }}>
                {data?.users?.filter(u => u.status === 'STATUS_ACTIVE').length || 0}
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
              <div className="text-sm font-medium" style={{ color: '#64748b' }}>Inactive Users</div>
              <div className="text-2xl font-bold mt-1" style={{ color: '#1e293b' }}>
                {data?.users?.filter(u => u.status === 'STATUS_INACTIVE').length || 0}
              </div>
            </div>
            <div className="p-2 rounded-lg bg-gray-50">
              <div className="h-5 w-5 rounded-full bg-gray-400"></div>
            </div>
          </div>
        </Card>
        <Card className="shadow-sm">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm font-medium" style={{ color: '#64748b' }}>Avg. Groups/User</div>
              <div className="text-2xl font-bold mt-1" style={{ color: '#1e293b' }}>
                {data?.users?.length ?
                  (data.users.reduce((acc, user) => acc + (user.groupIds?.length || 0), 0) / data.users.length).toFixed(1)
                  : '0.0'
                }
              </div>
            </div>
            <div className="p-2 rounded-lg bg-purple-50">
              <FolderKanban className="h-5 w-5 text-purple-600" />
            </div>
          </div>
        </Card>
      </div>

      {/* ユーザーテーブル */}
      <Card className="shadow-sm">
        <Table
          columns={columns}
          dataSource={filteredUsers}
          loading={isLoading}
          rowKey="id"
          locale={{
            emptyText: error ? 'Failed to load users' : 'No users found'
          }}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => `${range[0]}-${range[1]} of ${total} users`,
          }}
        />
      </Card>

      <Modal
        title="Create New User"
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
            name="username"
            label="Username"
            rules={[{ required: true, message: 'Please input username!' }]}
          >
            <Input />
          </Form.Item>
          <Form.Item
            name="email"
            label="Email"
            rules={[
              { required: true, message: 'Please input email!' },
              { type: 'email', message: 'Please input valid email!' }
            ]}
          >
            <Input type="email" />
          </Form.Item>
          <Form.Item
            name="password"
            label="Password"
            rules={[{ required: true, message: 'Please input password!' }]}
          >
            <Input.Password />
          </Form.Item>
          <Form.Item
            name="groupIds"
            label="Groups (optional)"
          >
            <Select
              mode="multiple"
              placeholder="Select groups"
              allowClear
            >
              {groupsData?.groups?.map((group) => (
                <Option key={group.id} value={group.id}>
                  {group.name}
                </Option>
              ))}
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
        title="Edit User"
        open={!!editingUser}
        onCancel={() => setEditingUser(null)}
        footer={null}
      >
        {editingUser && (
          <Form
            form={editForm}
            layout="vertical"
            onFinish={handleEditSubmit}
          >
            <Form.Item
              name="username"
              label="Username"
              rules={[{ required: true, message: 'Please input username!' }]}
            >
              <Input />
            </Form.Item>
            <Form.Item
              name="email"
              label="Email"
              rules={[
                { required: true, message: 'Please input email!' },
                { type: 'email', message: 'Please input valid email!' }
              ]}
            >
              <Input type="email" />
            </Form.Item>
            <Form.Item
              name="password"
              label="Password (leave empty to keep unchanged)"
            >
              <Input.Password placeholder="••••••••" />
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
            <Form.Item
              name="groupIds"
              label="Groups (optional)"
            >
              <Select
                mode="multiple"
                placeholder="Select groups"
                allowClear
              >
                {groupsData?.groups?.map((group) => (
                  <Option key={group.id} value={group.id}>
                    {group.name}
                  </Option>
                ))}
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
        title={`Manage Groups for ${managingGroupsUser?.username}`}
        open={!!managingGroupsUser}
        onCancel={() => setManagingGroupsUser(null)}
        width={800}
        footer={null}
      >
        {managingGroupsUser && (
          <div className="space-y-6">
            <div>
              <h3 className="text-lg font-medium">Current Groups</h3>
              {managingGroupsUser.groupIds.length === 0 ? (
                <p className="text-sm text-gray-500">No groups assigned</p>
              ) : (
                <div className="mt-2 space-y-2">
                  {managingGroupsUser.groupIds.map((groupId) => {
                    const group = groupsData?.groups?.find(g => g.id === groupId);
                    return group ? (
                      <div key={groupId} className="flex items-center justify-between rounded-md border px-3 py-2">
                        <span>{group.name}</span>
                        <Button
                          type="text"
                          danger
                          onClick={() => {
                            if (confirm(`Are you sure you want to remove ${group.name} from this user?`)) {
                              unassignGroupMutation.mutate({
                                userId: managingGroupsUser.id,
                                groupId: group.id
                              });
                            }
                          }}
                          loading={unassignGroupMutation.isPending && unassignGroupMutation.variables?.groupId === group.id}
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
              <h3 className="text-lg font-medium">Available Groups</h3>
              {groupsData?.groups?.filter(group => !managingGroupsUser.groupIds.includes(group.id)).length === 0 ? (
                <p className="text-sm text-gray-500">No available groups</p>
              ) : (
                <div className="mt-2 space-y-2">
                  {groupsData?.groups
                    ?.filter(group => !managingGroupsUser.groupIds.includes(group.id))
                    .map((group) => (
                      <div key={group.id} className="flex items-center justify-between rounded-md border px-3 py-2">
                        <span>{group.name}</span>
                        <Button
                          type="text"
                          onClick={() => {
                            assignGroupMutation.mutate({
                              userId: managingGroupsUser.id,
                              groupId: group.id
                            });
                          }}
                          loading={assignGroupMutation.isPending && assignGroupMutation.variables?.groupId === group.id}
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
