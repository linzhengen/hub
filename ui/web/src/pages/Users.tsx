import React, { useMemo, useState } from 'react';
import { expiresSoon, formatTerm, groupIdsOf } from '@/lib/grants';
import { useQuery, useMutation, useQueryClient, keepPreviousData } from '@tanstack/react-query';
import { userService, User, UpdateUserRequest } from '@/services/user';
import { groupService, ListGroupsResponse } from '@/services/group';
import {
  Button, Modal, Input, Table, Form, Select, Space, Card, Tag, Drawer,
  Checkbox, Tooltip, Dropdown,
} from 'antd';
import type { MenuProps } from 'antd';
import type { TableRowSelection } from 'antd/es/table/interface';
import {
  PlusOutlined, EditOutlined, DeleteOutlined, UserOutlined,
  SearchOutlined, DownloadOutlined, FilterOutlined, MoreOutlined,
} from '@ant-design/icons';
import { Users as UsersIcon } from 'lucide-react';

import { MAX_PAGE_SIZE, useServerPagination } from '@/hooks/useServerPagination';
import { useDebouncedValue } from '@/hooks/useDebouncedValue';
import { cn } from '@/lib/utils';
import { useDangerConfirm } from '@/hooks/useDangerConfirm';
import { useNotify } from '@/hooks/useNotify';

const { Option } = Select;

type StatusFilter = 'STATUS_ACTIVE' | 'STATUS_INACTIVE' | '';

interface UserFormValues {
  username: string;
  email: string;
  password?: string;
  status?: User['status'];
  groupIds?: string[];
}

// ── CSV download helper ──────────────────────────────────────────────────────

function downloadCsv(users: User[], groupNameFn: (ids: string[] | undefined) => string[]) {
  const header = ['ID', 'Username', 'Email', 'Status', 'Groups', 'Created At'];
  const rows = users.map((u) => [
    u.id ?? '',
    u.username ?? '',
    u.email ?? '',
    u.status?.replace('STATUS_', '') ?? '',
    groupNameFn(groupIdsOf(u.groups)).join('; '),
    u.createdAt ? new Date(u.createdAt).toISOString() : '',
  ]);
  const csv = [header, ...rows]
    .map((r) => r.map((v) => `"${v.replace(/"/g, '""')}"`).join(','))
    .join('\n');
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `users-${new Date().toISOString().slice(0, 10)}.csv`;
  a.click();
  URL.revokeObjectURL(url);
}

// ── Status badge ─────────────────────────────────────────────────────────────

function StatusBadge({ status }: { status: string | undefined }) {
  if (status === 'STATUS_ACTIVE') return <Tag color="success">Active</Tag>;
  if (status === 'STATUS_INACTIVE') return <Tag color="error">Inactive</Tag>;
  return <Tag>Unknown</Tag>;
}

// ── Main component ────────────────────────────────────────────────────────────

export function Users() {
  const queryClient = useQueryClient();
  const confirmDanger = useDangerConfirm();
  const notify = useNotify();

  // ── modal / drawer state ──
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [managingGroupsUser, setManagingGroupsUser] = useState<User | null>(null);
  const [createForm] = Form.useForm();
  const [editForm] = Form.useForm();

  // ── filter state ──
  const [searchText, setSearchText] = useState('');
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('');
  const [groupFilter, setGroupFilter] = useState<string[]>([]);
  const [showFilters, setShowFilters] = useState(false);

  // ── bulk selection ──
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);

  const search = useDebouncedValue(searchText);
  const filterKey = useMemo(
    () => ({ search, statusFilter, groupFilter }),
    [search, statusFilter, groupFilter],
  );
  const pagination = useServerPagination(filterKey);

  const { data, isLoading, error } = useQuery({
    queryKey: ['users', pagination.params, filterKey],
    queryFn: () =>
      userService.listUsers({
        ...pagination.params,
        userName: search || undefined,
        status: (statusFilter as User['status']) || undefined,
        groupIds: groupFilter.length ? groupFilter : undefined,
      }),
    placeholderData: keepPreviousData,
  });

  const { data: groupsData } = useQuery<ListGroupsResponse>({
    queryKey: ['groups', 'all'],
    queryFn: () => groupService.listGroups({ limit: MAX_PAGE_SIZE }),
  });

  const getGroupNames = (groupIds: string[] | undefined): string[] => {
    if (!groupsData?.groups || !groupIds) return [];
    return groupIds
      .map((id) => groupsData.groups?.find((g) => g.id === id)?.name)
      .filter((name): name is string => name !== undefined);
  };

  // ── mutations ──

  const createMutation = useMutation({
    mutationFn: userService.createUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      setIsCreateOpen(false);
      createForm.resetFields();
      notify.success('User created successfully');
    },
    onError: (error: Error) => notify.error(error.message),
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateUserRequest }) =>
      userService.updateUser(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      setEditingUser(null);
      notify.success('User updated successfully');
    },
    onError: (error: Error) => notify.error(error.message),
  });

  const deleteMutation = useMutation({
    mutationFn: userService.deleteUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      notify.success('User deleted successfully');
    },
    onError: (error: Error) => notify.error(error.message),
  });

  const assignGroupMutation = useMutation({
    mutationFn: ({ userId, groupIds }: { userId: string; groupIds: string[] }) =>
      userService.addGroups(userId, groupIds),
    onSuccess: (res, variables) => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      if (managingGroupsUser?.id === variables.userId) {
        setManagingGroupsUser(res.user ?? managingGroupsUser);
      }
      notify.success('Groups assigned');
    },
    onError: (error: Error) => notify.error(error.message),
  });

  const unassignGroupMutation = useMutation({
    mutationFn: ({ userId, groupIds }: { userId: string; groupIds: string[] }) =>
      userService.removeGroups(userId, groupIds),
    onSuccess: (res, variables) => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      if (managingGroupsUser?.id === variables.userId) {
        setManagingGroupsUser(res.user ?? managingGroupsUser);
      }
      notify.success('Groups removed');
    },
    onError: (error: Error) => notify.error(error.message),
  });

  // ── bulk actions ──

  const selectedUsers = useMemo(
    () => (data?.users ?? []).filter((u) => selectedRowKeys.includes(u.id ?? '')),
    [data?.users, selectedRowKeys],
  );

  const handleBulkDelete = () => {
    confirmDanger(
      {
        title: `Delete ${selectedRowKeys.length} users`,
        content: 'This cannot be undone. Are you sure?',
      },
      async () => {
        await Promise.all(selectedRowKeys.map((key) => deleteMutation.mutateAsync(String(key))));
        setSelectedRowKeys([]);
      },
    );
  };

  const handleBulkStatusChange = (status: 'STATUS_ACTIVE' | 'STATUS_INACTIVE') => {
    confirmDanger(
      {
        title: `Set ${selectedRowKeys.length} users to ${status === 'STATUS_ACTIVE' ? 'Active' : 'Inactive'}`,
        content: 'This will update all selected users.',
        okText: 'Confirm',
      },
      async () => {
        await Promise.all(
          selectedUsers.map((u) =>
            updateMutation.mutateAsync({
              id: u.id!,
              data: { username: u.username!, email: u.email!, status },
            }),
          ),
        );
        setSelectedRowKeys([]);
      },
    );
  };

  const handleBulkDownload = () => {
    const target = selectedUsers.length > 0 ? selectedUsers : (data?.users ?? []);
    downloadCsv(target, getGroupNames);
    notify.success(`Downloaded ${target.length} users as CSV`);
  };

  // ── manage-groups drawer state ──
  const [drawerMembers, setDrawerMembers] = useState<string[]>([]);
  const [drawerAvailable, setDrawerAvailable] = useState<string[]>([]);
  const [memberSearch, setMemberSearch] = useState('');
  const [availableSearch, setAvailableSearch] = useState('');

  const openManageGroups = (user: User) => {
    setManagingGroupsUser(user);
    setDrawerMembers([]);
    setDrawerAvailable([]);
    setMemberSearch('');
    setAvailableSearch('');
  };

  const drawerMemberGroups = useMemo(
    () =>
      (groupsData?.groups ?? []).filter((g) =>
        groupIdsOf(managingGroupsUser?.groups).includes(g.id ?? ''),
      ),
    [groupsData?.groups, managingGroupsUser?.groups],
  );

  const drawerAvailableGroups = useMemo(
    () =>
      (groupsData?.groups ?? []).filter(
        (g) => !groupIdsOf(managingGroupsUser?.groups).includes(g.id ?? ''),
      ),
    [groupsData?.groups, managingGroupsUser?.groups],
  );

  const filteredMemberGroups = useMemo(
    () =>
      drawerMemberGroups.filter((g) =>
        g.name?.toLowerCase().includes(memberSearch.toLowerCase()),
      ),
    [drawerMemberGroups, memberSearch],
  );

  const filteredAvailableGroups = useMemo(
    () =>
      drawerAvailableGroups.filter((g) =>
        g.name?.toLowerCase().includes(availableSearch.toLowerCase()),
      ),
    [drawerAvailableGroups, availableSearch],
  );

  const handleRemoveGroups = () => {
    if (!managingGroupsUser?.id || drawerMembers.length === 0) return;
    unassignGroupMutation.mutate({ userId: managingGroupsUser.id, groupIds: drawerMembers });
    setDrawerMembers([]);
  };

  const handleAddGroups = () => {
    if (!managingGroupsUser?.id || drawerAvailable.length === 0) return;
    assignGroupMutation.mutate({ userId: managingGroupsUser.id, groupIds: drawerAvailable });
    setDrawerAvailable([]);
  };

  // ── table ──

  const rowSelection: TableRowSelection<User> = {
    selectedRowKeys,
    onChange: setSelectedRowKeys,
  };

  const bulkMenuItems: MenuProps['items'] = [
    { key: 'activate', label: 'Set Active', onClick: () => handleBulkStatusChange('STATUS_ACTIVE') },
    { key: 'deactivate', label: 'Set Inactive', onClick: () => handleBulkStatusChange('STATUS_INACTIVE') },
    { type: 'divider' },
    { key: 'delete', label: <span className="text-red-600">Delete selected</span>, onClick: handleBulkDelete },
  ];

  const columns = [
    {
      title: 'Username',
      dataIndex: 'username',
      key: 'username',
      width: 220,
      render: (name: string, record: User) => (
        <div className="flex items-center gap-2">
          <div className="h-8 w-8 rounded-full bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center flex-shrink-0">
            <span className="text-blue-600 dark:text-blue-400 font-semibold text-sm">
              {name?.charAt(0)?.toUpperCase()}
            </span>
          </div>
          <div>
            <div className="font-medium text-gray-900 dark:text-white">{name}</div>
            <div className="text-xs text-gray-400 dark:text-gray-500">{record.email}</div>
          </div>
        </div>
      ),
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => <StatusBadge status={status} />,
    },
    {
      title: 'Groups',
      key: 'groups',
      width: 200,
      render: (_: unknown, record: User) => {
        const memberships = record.groups ?? [];
        return (
          <div className="flex flex-wrap gap-1">
            {memberships.length > 0 ? (
              memberships.map((membership) => {
                const [name] = getGroupNames([membership.groupId ?? '']);
                // A membership that ends is not the same fact as one that does
                // not, so it does not get the same tag. The tooltip carries the
                // date; the colour carries "somebody has to act".
                return membership.expiresAt ? (
                  <Tooltip key={membership.groupId} title={formatTerm(membership.expiresAt)}>
                    <Tag
                      color={expiresSoon(membership.expiresAt) ? 'orange' : 'cyan'}
                      className="m-0"
                    >
                      {name ?? membership.groupId} ⏳
                    </Tag>
                  </Tooltip>
                ) : (
                  <Tag key={membership.groupId} color="blue" className="m-0">
                    {name ?? membership.groupId}
                  </Tag>
                );
              })
            ) : (
              <span className="text-xs text-gray-400">—</span>
            )}
          </div>
        );
      },
    },
    {
      title: 'Created',
      dataIndex: 'createdAt',
      key: 'createdAt',
      width: 110,
      render: (v: string) => (v ? new Date(v).toLocaleDateString() : '—'),
    },
    {
      title: '',
      key: 'actions',
      width: 80,
      fixed: 'right' as const,
      render: (_: unknown, record: User) => (
        <Space size={4}>
          <Tooltip title="Manage Groups">
            <Button type="text" size="small" icon={<UserOutlined />} onClick={() => openManageGroups(record)} />
          </Tooltip>
          <Tooltip title="Edit">
            <Button
              type="text"
              size="small"
              icon={<EditOutlined />}
              onClick={() => {
                setEditingUser(record);
                editForm.setFieldsValue({
                  username: record.username,
                  email: record.email,
                  status: record.status,
                  groupIds: groupIdsOf(record.groups),
                });
              }}
              className="text-blue-600 dark:text-blue-400"
            />
          </Tooltip>
          <Tooltip title="Delete">
            <Button
              type="text"
              size="small"
              icon={<DeleteOutlined />}
              onClick={() => {
                const id = record.id;
                if (!id) return;
                confirmDanger(
                  { title: 'Delete user', content: 'Are you sure you want to delete this user?' },
                  () => deleteMutation.mutate(id),
                );
              }}
              className="text-red-600 dark:text-red-400"
            />
          </Tooltip>
        </Space>
      ),
    },
  ];

  // ── stats ──
  const totalUsers = Number(data?.total ?? 0);
  const activeCount = data?.users?.filter((u) => u.status === 'STATUS_ACTIVE').length ?? 0;
  const inactiveCount = data?.users?.filter((u) => u.status === 'STATUS_INACTIVE').length ?? 0;
  const avgGroups = data?.users?.length
    ? ((data.users.reduce((s, u) => s + (u.groups?.length ?? 0), 0)) / data.users.length).toFixed(1)
    : '—';

  const activeFilter = statusFilter !== '' || groupFilter.length > 0;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold tracking-tight text-gray-900 dark:text-white">Users</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400">Manage user accounts and permissions</p>
        </div>
        <div className="flex items-center gap-2">
          <Tooltip title="Export visible/selected users as CSV">
            <Button icon={<DownloadOutlined />} onClick={handleBulkDownload}>Export</Button>
          </Tooltip>
          <Button
            icon={<FilterOutlined />}
            onClick={() => setShowFilters((v) => !v)}
            type={activeFilter ? 'primary' : 'default'}
            ghost={activeFilter}
          >
            Filters{activeFilter ? ' ●' : ''}
          </Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setIsCreateOpen(true)} style={{ borderRadius: '8px' }}>
            Add User
          </Button>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        {[
          { label: 'Total Users', value: totalUsers, icon: <UsersIcon className="h-5 w-5 text-blue-600 dark:text-blue-400" />, bg: 'bg-blue-50 dark:bg-blue-900/20' },
          { label: 'Active', value: activeCount, icon: <div className="h-3 w-3 rounded-full bg-green-500" />, bg: 'bg-green-50 dark:bg-green-900/20' },
          { label: 'Inactive', value: inactiveCount, icon: <div className="h-3 w-3 rounded-full bg-gray-400" />, bg: 'bg-gray-50 dark:bg-gray-700/50' },
          { label: 'Groups / User', value: avgGroups, icon: null, bg: 'bg-purple-50 dark:bg-purple-900/20' },
        ].map(({ label, value, icon, bg }) => (
          <Card key={label} className="shadow-sm dark:bg-gray-800 dark:border-gray-700" size="small">
            <div className="flex items-center justify-between">
              <div>
                <div className="text-xs font-medium text-gray-500 dark:text-gray-400">{label}</div>
                <div className="text-2xl font-bold mt-1 text-gray-900 dark:text-white">{value}</div>
              </div>
              {icon && <div className={cn('p-2 rounded-lg', bg)}>{icon}</div>}
            </div>
          </Card>
        ))}
      </div>

      {/* Filter panel (expanded) */}
      {showFilters && (
        <Card className="shadow-sm dark:bg-gray-800 dark:border-gray-700" size="small">
          <div className="flex flex-wrap items-end gap-4">
            <div className="flex-1 min-w-48">
              <div className="text-xs text-gray-500 mb-1">Search by name</div>
              <Input
                placeholder="Username..."
                prefix={<SearchOutlined className="text-gray-400" />}
                value={searchText}
                onChange={(e) => setSearchText(e.target.value)}
                allowClear
              />
            </div>
            <div className="w-40">
              <div className="text-xs text-gray-500 mb-1">Status</div>
              <Select
                value={statusFilter}
                onChange={setStatusFilter}
                style={{ width: '100%' }}
                allowClear
                placeholder="All statuses"
              >
                <Option value="STATUS_ACTIVE">Active</Option>
                <Option value="STATUS_INACTIVE">Inactive</Option>
              </Select>
            </div>
            <div className="flex-1 min-w-48">
              <div className="text-xs text-gray-500 mb-1">Group</div>
              <Select
                mode="multiple"
                value={groupFilter}
                onChange={setGroupFilter}
                style={{ width: '100%' }}
                allowClear
                placeholder="All groups"
                maxTagCount="responsive"
              >
                {groupsData?.groups?.map((g) => (
                  <Option key={g.id} value={g.id}>{g.name}</Option>
                ))}
              </Select>
            </div>
            <Button
              onClick={() => { setSearchText(''); setStatusFilter(''); setGroupFilter([]); }}
              disabled={!activeFilter && !searchText}
            >
              Clear
            </Button>
          </div>
        </Card>
      )}

      {/* Compact search (collapsed filter view) */}
      {!showFilters && (
        <Input
          placeholder="Search by username..."
          prefix={<SearchOutlined className="text-gray-400" />}
          value={searchText}
          onChange={(e) => setSearchText(e.target.value)}
          style={{ maxWidth: 320, borderRadius: '8px' }}
          allowClear
        />
      )}

      {/* Bulk action bar */}
      {selectedRowKeys.length > 0 && (
        <div className="flex items-center gap-3 px-4 py-2 rounded-lg bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800">
          <span className="text-sm font-medium text-blue-700 dark:text-blue-300">
            {selectedRowKeys.length} selected
          </span>
          <Button size="small" icon={<DownloadOutlined />} onClick={handleBulkDownload}>Export</Button>
          <Dropdown menu={{ items: bulkMenuItems }} trigger={['click']}>
            <Button size="small" icon={<MoreOutlined />}>Actions</Button>
          </Dropdown>
          <Button size="small" type="text" onClick={() => setSelectedRowKeys([])}>Clear</Button>
        </div>
      )}

      {/* Table */}
      <Card className="shadow-sm dark:bg-gray-800 dark:border-gray-700">
        <Table
          rowSelection={rowSelection}
          columns={columns}
          dataSource={data?.users}
          loading={isLoading}
          rowKey="id"
          scroll={{ x: 700 }}
          locale={{ emptyText: error ? 'Failed to load users' : 'No users found' }}
          pagination={pagination.tableProps(Number(data?.total ?? 0), 'users')}
        />
      </Card>

      {/* Create Modal */}
      <Modal
        title="Create New User"
        open={isCreateOpen}
        onCancel={() => { setIsCreateOpen(false); createForm.resetFields(); }}
        footer={null}
        width={480}
      >
        <Form
          form={createForm}
          layout="vertical"
          onFinish={(v: UserFormValues) => {
            createMutation.mutate({ username: v.username, email: v.email, password: v.password, groupIds: v.groupIds ?? [] });
          }}
        >
          <Form.Item name="username" label="Username" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="email" label="Email" rules={[{ required: true }, { type: 'email' }]}>
            <Input type="email" />
          </Form.Item>
          <Form.Item name="password" label="Password" rules={[{ required: true }, { min: 8, message: 'At least 8 characters' }]}>
            <Input.Password />
          </Form.Item>
          <Form.Item name="groupIds" label="Groups (optional)">
            <Select mode="multiple" placeholder="Select groups" allowClear>
              {groupsData?.groups?.map((g) => <Option key={g.id} value={g.id}>{g.name}</Option>)}
            </Select>
          </Form.Item>
          <Form.Item className="mb-0">
            <div className="flex justify-end gap-2">
              <Button onClick={() => { setIsCreateOpen(false); createForm.resetFields(); }}>Cancel</Button>
              <Button type="primary" htmlType="submit" loading={createMutation.isPending}>Create</Button>
            </div>
          </Form.Item>
        </Form>
      </Modal>

      {/* Edit Modal */}
      <Modal
        title="Edit User"
        open={!!editingUser}
        onCancel={() => setEditingUser(null)}
        footer={null}
        width={480}
      >
        {editingUser && (
          <Form
            form={editForm}
            layout="vertical"
            onFinish={(v: UserFormValues) => {
              const d: UpdateUserRequest = { username: v.username, email: v.email };
              if (v.password) d.password = v.password;
              if (v.status) d.status = v.status;
              if (v.groupIds) d.groupIds = v.groupIds;
              updateMutation.mutate({ id: editingUser.id!, data: d });
            }}
          >
            <Form.Item name="username" label="Username" rules={[{ required: true }]}>
              <Input />
            </Form.Item>
            <Form.Item name="email" label="Email" rules={[{ required: true }, { type: 'email' }]}>
              <Input type="email" />
            </Form.Item>
            <Form.Item name="password" label="Password (leave empty to keep unchanged)" rules={[{ min: 8, message: 'At least 8 characters' }]}>
              <Input.Password placeholder="••••••••" />
            </Form.Item>
            <Form.Item name="status" label="Status">
              <Select>
                <Option value="STATUS_UNSPECIFIED">Unspecified</Option>
                <Option value="STATUS_ACTIVE">Active</Option>
                <Option value="STATUS_INACTIVE">Inactive</Option>
              </Select>
            </Form.Item>
            <Form.Item name="groupIds" label="Groups (optional)">
              <Select mode="multiple" placeholder="Select groups" allowClear>
                {groupsData?.groups?.map((g) => <Option key={g.id} value={g.id}>{g.name}</Option>)}
              </Select>
            </Form.Item>
            <Form.Item className="mb-0">
              <div className="flex justify-end gap-2">
                <Button onClick={() => setEditingUser(null)}>Cancel</Button>
                <Button type="primary" htmlType="submit" loading={updateMutation.isPending}>Save Changes</Button>
              </div>
            </Form.Item>
          </Form>
        )}
      </Modal>

      {/* Manage Groups Drawer */}
      <Drawer
        title={
          <div className="flex items-center gap-2">
            <UserOutlined />
            <span>Manage Groups — {managingGroupsUser?.username}</span>
          </div>
        }
        open={!!managingGroupsUser}
        onClose={() => setManagingGroupsUser(null)}
        width={720}
        footer={null}
      >
        {managingGroupsUser && (
          <div className="flex gap-4" style={{ height: 'calc(100vh - 120px)' }}>
            {/* Members panel */}
            <div className="flex-1 flex flex-col min-w-0">
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm font-semibold text-gray-700 dark:text-gray-300">
                  Members ({drawerMemberGroups.length})
                </span>
                <Checkbox
                  indeterminate={drawerMembers.length > 0 && drawerMembers.length < filteredMemberGroups.length}
                  checked={filteredMemberGroups.length > 0 && drawerMembers.length === filteredMemberGroups.length}
                  onChange={(e) =>
                    setDrawerMembers(e.target.checked ? filteredMemberGroups.map((g) => g.id ?? '') : [])
                  }
                >
                  All
                </Checkbox>
              </div>
              <Input
                placeholder="Search..."
                prefix={<SearchOutlined className="text-gray-400" />}
                value={memberSearch}
                onChange={(e) => setMemberSearch(e.target.value)}
                allowClear
                size="small"
                className="mb-2"
              />
              <div className="flex-1 overflow-y-auto border rounded-lg divide-y dark:divide-gray-700 dark:border-gray-700">
                {filteredMemberGroups.length === 0 ? (
                  <div className="py-8 text-center text-sm text-gray-400">No groups assigned</div>
                ) : (
                  filteredMemberGroups.map((g) => (
                    <label key={g.id} className="flex items-center gap-3 px-3 py-2 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700/50">
                      <Checkbox
                        checked={drawerMembers.includes(g.id ?? '')}
                        onChange={(e) =>
                          setDrawerMembers(
                            e.target.checked
                              ? [...drawerMembers, g.id ?? '']
                              : drawerMembers.filter((id) => id !== g.id),
                          )
                        }
                      />
                      <div className="h-7 w-7 rounded-full bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center flex-shrink-0 text-blue-600 dark:text-blue-400 text-xs font-bold">
                        {g.name?.charAt(0)?.toUpperCase()}
                      </div>
                      <span className="text-sm text-gray-800 dark:text-gray-200">{g.name}</span>
                    </label>
                  ))
                )}
              </div>
            </div>

            {/* Center buttons */}
            <div className="flex flex-col items-center justify-center gap-3 flex-shrink-0">
              <Tooltip title="Remove selected from user">
                <Button
                  shape="circle"
                  icon={<span className="font-bold">←</span>}
                  disabled={drawerMembers.length === 0}
                  loading={unassignGroupMutation.isPending}
                  onClick={handleRemoveGroups}
                />
              </Tooltip>
              <Tooltip title="Add selected to user">
                <Button
                  shape="circle"
                  icon={<span className="font-bold">→</span>}
                  disabled={drawerAvailable.length === 0}
                  loading={assignGroupMutation.isPending}
                  onClick={handleAddGroups}
                />
              </Tooltip>
            </div>

            {/* Available panel */}
            <div className="flex-1 flex flex-col min-w-0">
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm font-semibold text-gray-700 dark:text-gray-300">
                  Available ({drawerAvailableGroups.length})
                </span>
                <Checkbox
                  indeterminate={drawerAvailable.length > 0 && drawerAvailable.length < filteredAvailableGroups.length}
                  checked={filteredAvailableGroups.length > 0 && drawerAvailable.length === filteredAvailableGroups.length}
                  onChange={(e) =>
                    setDrawerAvailable(e.target.checked ? filteredAvailableGroups.map((g) => g.id ?? '') : [])
                  }
                >
                  All
                </Checkbox>
              </div>
              <Input
                placeholder="Search..."
                prefix={<SearchOutlined className="text-gray-400" />}
                value={availableSearch}
                onChange={(e) => setAvailableSearch(e.target.value)}
                allowClear
                size="small"
                className="mb-2"
              />
              <div className="flex-1 overflow-y-auto border rounded-lg divide-y dark:divide-gray-700 dark:border-gray-700">
                {filteredAvailableGroups.length === 0 ? (
                  <div className="py-8 text-center text-sm text-gray-400">No groups available</div>
                ) : (
                  filteredAvailableGroups.map((g) => (
                    <label key={g.id} className="flex items-center gap-3 px-3 py-2 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700/50">
                      <Checkbox
                        checked={drawerAvailable.includes(g.id ?? '')}
                        onChange={(e) =>
                          setDrawerAvailable(
                            e.target.checked
                              ? [...drawerAvailable, g.id ?? '']
                              : drawerAvailable.filter((id) => id !== g.id),
                          )
                        }
                      />
                      <div className="h-7 w-7 rounded-full bg-gray-100 dark:bg-gray-700 flex items-center justify-center flex-shrink-0 text-gray-600 dark:text-gray-400 text-xs font-bold">
                        {g.name?.charAt(0)?.toUpperCase()}
                      </div>
                      <span className="text-sm text-gray-800 dark:text-gray-200">{g.name}</span>
                    </label>
                  ))
                )}
              </div>
            </div>
          </div>
        )}
      </Drawer>
    </div>
  );
}
