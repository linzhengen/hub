import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { userService, User } from '@/services/user';
import { groupService, Group, ListGroupsResponse } from '@/services/group';
import { Button, Modal, Input, Table, Form, Select, Tag, Space } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, UserOutlined } from '@ant-design/icons';

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
      title: 'Groups',
      key: 'groups',
      render: (_: any, record: User) => (
        <span>{getGroupNames(record.groupIds).join(', ') || 'None'}</span>
      ),
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
          />
          <Button
            type="text"
            icon={<DeleteOutlined />}
            danger
            onClick={() => {
              if (confirm('Are you sure you want to delete this user?')) {
                deleteMutation.mutate(record.id);
              }
            }}
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

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold tracking-tight">Users</h2>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => setIsCreateOpen(true)}
        >
          Add User
        </Button>
      </div>

      <Table
        columns={columns}
        dataSource={data?.users}
        loading={isLoading}
        rowKey="id"
        locale={{
          emptyText: error ? 'Failed to load users' : 'No users found'
        }}
      />

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
