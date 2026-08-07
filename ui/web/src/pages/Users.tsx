import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient, keepPreviousData } from '@tanstack/react-query';
import { userService, User, UpdateUserRequest } from '@/services/user';
import { groupService, ListGroupsResponse } from '@/services/group';
import { Button, Modal, Input, Table, Form, Select, Space, Card } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, UserOutlined } from '@ant-design/icons';
import { FolderKanban } from 'lucide-react';

import { toast } from 'sonner';
import { MAX_PAGE_SIZE, useServerPagination } from '@/hooks/useServerPagination';
import { useDebouncedValue } from '@/hooks/useDebouncedValue';
import { cn } from '@/lib/utils';
import { useDangerConfirm } from '@/hooks/useDangerConfirm';

const { Option } = Select;

interface UserFormValues {
  username: string;
  email: string;
  password?: string;
  status?: User['status'];
  groupIds?: string[];
}

export function Users() {
  const queryClient = useQueryClient();
  const confirmDanger = useDangerConfirm();
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [managingGroupsUser, setManagingGroupsUser] = useState<User | null>(null);
  const [createForm] = Form.useForm();
  const [editForm] = Form.useForm();
  const [searchText, setSearchText] = useState('');

  const search = useDebouncedValue(searchText);
  const pagination = useServerPagination(search);


  const { data, isLoading, error } = useQuery({
    queryKey: ['users', pagination.params, search],
    queryFn: () => userService.listUsers({ ...pagination.params, userName: search || undefined }),
    placeholderData: keepPreviousData,
  });

  // The group names shown against each user, and the group pickers, need the
  // whole list rather than a page of it.
  const { data: groupsData } = useQuery<ListGroupsResponse>({
    queryKey: ['groups', 'all'],
    queryFn: () => groupService.listGroups({ limit: MAX_PAGE_SIZE }),
  });

  // Helper function to get group names from group IDs
  const getGroupNames = (groupIds: string[] | undefined): string[] => {
    if (!groupsData?.groups || !groupIds) return [];
    return groupIds
      .map(id => groupsData.groups?.find(g => g.id === id)?.name)
      .filter((name): name is string => name !== undefined);
  };

  const createMutation = useMutation({
    mutationFn: userService.createUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      setIsCreateOpen(false);
      toast.success('User created successfully');
    },
    onError: (error: Error) => toast.error(error.message),
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateUserRequest }) => userService.updateUser(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      setEditingUser(null);
      toast.success('User updated successfully');
    },
    onError: (error: Error) => toast.error(error.message),
  });

  const deleteMutation = useMutation({
    mutationFn: userService.deleteUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      toast.success('User deleted successfully');
    },
    onError: (error: Error) => toast.error(error.message),
  });

  const assignGroupMutation = useMutation({
    mutationFn: ({ userId, groupId }: { userId: string; groupId: string }) =>
      userService.addGroups(userId, [groupId]),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      // Update the managingGroupsUser state to reflect the change immediately
      if (managingGroupsUser && managingGroupsUser.id === variables.userId) {
        setManagingGroupsUser({
          ...managingGroupsUser,
          groupIds: [...(managingGroupsUser.groupIds ?? []), variables.groupId]
        });
      }
      toast.success('Group assigned successfully');
    },
    onError: (error: Error) => toast.error(error.message),
  });

  const unassignGroupMutation = useMutation({
    mutationFn: ({ userId, groupId }: { userId: string; groupId: string }) =>
      userService.removeGroups(userId, [groupId]),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      // Update the managingGroupsUser state to reflect the change immediately
      if (managingGroupsUser && managingGroupsUser.id === variables.userId) {
        setManagingGroupsUser({
          ...managingGroupsUser,
          groupIds: (managingGroupsUser.groupIds ?? []).filter(id => id !== variables.groupId)
        });
      }
      toast.success('Group unassigned successfully');
    },
    onError: (error: Error) => toast.error(error.message),
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
        let text = 'Unspecified';

        if (status === 'STATUS_ACTIVE') {
          text = 'Active';
        } else if (status === 'STATUS_INACTIVE') {
          text = 'Inactive';
        }

        return (
          <span
            className={cn(
              "inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium",
              status === 'STATUS_ACTIVE' ? "bg-emerald-100 text-emerald-800 dark:bg-emerald-500/10 dark:text-emerald-400" :
              status === 'STATUS_INACTIVE' ? "bg-red-100 text-red-800 dark:bg-red-500/10 dark:text-red-400" :
              "bg-gray-100 text-gray-800 dark:bg-gray-500/10 dark:text-gray-400"
            )}
          >
            {text}
          </span>
        );
      },
    },
    {
      title: 'Groups',
      key: 'groups',
      render: (_: unknown, record: User) => {
        const groupNames = getGroupNames(record.groupIds);
        return (
          <div className="flex flex-wrap gap-1">
            {groupNames.length > 0 ? (
              groupNames.map((name, index) => (
                <span
                  key={index}
                  className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-500/10 dark:text-blue-400 border border-blue-200 dark:border-blue-500/20"
                >
                  {name}
                </span>
              ))
            ) : (
              <span className="text-sm text-gray-400 dark:text-gray-500">None</span>
            )}
          </div>
        );
      },
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: unknown, record: User) => (
        <Space>
          <Button
            type="text"
            icon={<UserOutlined />}
            onClick={() => setManagingGroupsUser(record)}
            className="p-1.5 rounded-md text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
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
            className="p-1.5 rounded-md text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/20 transition-colors"
          />
          <Button
            type="text"
            icon={<DeleteOutlined />}
            onClick={() => {
              const id = record.id;
              if (!id) return;
              confirmDanger(
                {
                  title: 'Delete user',
                  content: 'Are you sure you want to delete this user?',
                },
                () => deleteMutation.mutate(id),
              );
            }}
            className="p-1.5 rounded-md text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors"
          />
        </Space>
      ),
    },
  ];

  const handleCreateSubmit = (values: UserFormValues) => {
    createMutation.mutate({
      username: values.username,
      email: values.email,
      password: values.password,
      groupIds: values.groupIds || [],
    });
  };

  const handleEditSubmit = (values: UserFormValues) => {
    if (!editingUser?.id) return;

    const data: UpdateUserRequest = {
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

  return (
    <div className="space-y-6">
      {/* ヘッダーセクション */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold tracking-tight text-gray-900 dark:text-white">Users</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400">Manage user accounts and permissions</p>
        </div>
        <div className="flex items-center gap-3">
          <Input
            placeholder="Search by username..."
            prefix={<UserOutlined className="text-gray-400 dark:text-gray-500" />}
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
        <Card className="shadow-sm dark:bg-gray-800 dark:border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm font-medium text-gray-500 dark:text-gray-400">Total Users</div>
              <div className="text-2xl font-bold mt-1 text-gray-900 dark:text-white">{data?.users?.length || 0}</div>
            </div>
            <div className="p-2 rounded-lg bg-blue-50 dark:bg-blue-900/20">
              <UserOutlined style={{ fontSize: '20px' }} className="text-blue-600 dark:text-blue-400" />
            </div>
          </div>
        </Card>
        <Card className="shadow-sm dark:bg-gray-800 dark:border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm font-medium text-gray-500 dark:text-gray-400">Active Users</div>
              <div className="text-2xl font-bold mt-1 text-gray-900 dark:text-white">
                {data?.users?.filter(u => u.status === 'STATUS_ACTIVE').length || 0}
              </div>
            </div>
            <div className="p-2 rounded-lg bg-green-50 dark:bg-green-900/20">
              <div className="h-5 w-5 rounded-full bg-green-500"></div>
            </div>
          </div>
        </Card>
        <Card className="shadow-sm dark:bg-gray-800 dark:border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm font-medium text-gray-500 dark:text-gray-400">Inactive Users</div>
              <div className="text-2xl font-bold mt-1 text-gray-900 dark:text-white">
                {data?.users?.filter(u => u.status === 'STATUS_INACTIVE').length || 0}
              </div>
            </div>
            <div className="p-2 rounded-lg bg-gray-50 dark:bg-gray-700/50">
              <div className="h-5 w-5 rounded-full bg-gray-400"></div>
            </div>
          </div>
        </Card>
        <Card className="shadow-sm dark:bg-gray-800 dark:border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm font-medium text-gray-500 dark:text-gray-400">Avg. Groups/User</div>
              <div className="text-2xl font-bold mt-1 text-gray-900 dark:text-white">
                {data?.users?.length ?
                  (data.users.reduce((acc, user) => acc + (user.groupIds?.length || 0), 0) / data.users.length).toFixed(1)
                  : '0.0'
                }
              </div>
            </div>
            <div className="p-2 rounded-lg bg-purple-50 dark:bg-purple-900/20">
              <FolderKanban className="h-5 w-5 text-purple-600 dark:text-purple-400" />
            </div>
          </div>
        </Card>
      </div>

      {/* ユーザーテーブル */}
      <Card className="shadow-sm dark:bg-gray-800 dark:border-gray-700">
        <Table
          columns={columns}
          dataSource={data?.users}
          loading={isLoading}
          rowKey="id"
          locale={{
            emptyText: error ? 'Failed to load users' : 'No users found'
          }}
          pagination={pagination.tableProps(Number(data?.total ?? 0), 'users')}
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
              {(managingGroupsUser.groupIds ?? []).length === 0 ? (
                <p className="text-sm text-gray-500">No groups assigned</p>
              ) : (
                <div className="mt-2 space-y-2">
                  {(managingGroupsUser.groupIds ?? []).map((groupId) => {
                    const group = groupsData?.groups?.find(g => g.id === groupId);
                    return group?.id ? (
                      <div key={groupId} className="flex items-center justify-between rounded-md border px-3 py-2">
                        <span>{group.name}</span>
                        <Button
                          type="text"
                          danger
                          onClick={() => {
                            const userId = managingGroupsUser.id;
                            const targetGroupId = group.id;
                            if (!userId || !targetGroupId) return;
                            confirmDanger(
                              {
                                title: 'Remove group',
                                content: `Are you sure you want to remove ${group.name} from this user?`,
                                okText: 'Remove',
                              },
                              () => unassignGroupMutation.mutate({ userId, groupId: targetGroupId }),
                            );
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
              {groupsData?.groups?.filter(group => !(managingGroupsUser.groupIds ?? []).includes(group.id ?? '')).length === 0 ? (
                <p className="text-sm text-gray-500">No available groups</p>
              ) : (
                <div className="mt-2 space-y-2">
                  {groupsData?.groups
                    ?.filter(group => !(managingGroupsUser.groupIds ?? []).includes(group.id ?? ''))
                    .map((group) => (
                      <div key={group.id} className="flex items-center justify-between rounded-md border px-3 py-2">
                        <span>{group.name}</span>
                        <Button
                          type="text"
                          onClick={() => {
                            if (managingGroupsUser.id && group.id) {
                              assignGroupMutation.mutate({
                                userId: managingGroupsUser.id,
                                groupId: group.id
                              });
                            }
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
