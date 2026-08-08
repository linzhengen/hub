import React, { useState, useMemo } from 'react';
import { useQuery, useMutation, useQueryClient, keepPreviousData } from '@tanstack/react-query';
import { groupService, Group, CreateGroupRequest, UpdateGroupRequest } from '@/services/group.ts';
import { roleService, Role } from '@/services/role.ts';
import { userService, User } from '@/services/user.ts';
import { Button, Modal, Input, Table, Form, Select, Space, Card, Drawer, Checkbox, Tooltip } from 'antd';

const { Option } = Select;
import { PlusOutlined, EditOutlined, DeleteOutlined, UserOutlined, SearchOutlined, FolderOutlined, ArrowRightOutlined, ArrowLeftOutlined } from '@ant-design/icons';
import { useNotify } from '@/hooks/useNotify';
import { MAX_PAGE_SIZE, useServerPagination } from '@/hooks/useServerPagination';
import { useDebouncedValue } from '@/hooks/useDebouncedValue';
import { useDangerConfirm } from '@/hooks/useDangerConfirm';
import { Users, Shield, Key } from 'lucide-react';

type GroupFormValues = CreateGroupRequest;

export function Groups() {
  const queryClient = useQueryClient();
  const confirmDanger = useDangerConfirm();
  const notify = useNotify();
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [editingGroup, setEditingGroup] = useState<Group | null>(null);
  const [managingRolesGroup, setManagingRolesGroup] = useState<Group | null>(null);
  const [managingUsersGroup, setManagingUsersGroup] = useState<Group | null>(null);
  const [createForm] = Form.useForm();
  const [editForm] = Form.useForm();
  const [searchText, setSearchText] = useState('');

  const search = useDebouncedValue(searchText);
  const pagination = useServerPagination(search);


  const { data, isLoading, error } = useQuery({
    queryKey: ['groups', pagination.params, search],
    queryFn: () => groupService.listGroups({ ...pagination.params, groupName: search || undefined }),
    placeholderData: keepPreviousData,
  });

  // The role and user pickers need the whole list, not a page of it.
  const { data: rolesData } = useQuery({
    queryKey: ['roles', 'all'],
    queryFn: () => roleService.listRoles({ limit: MAX_PAGE_SIZE }),
  });

  const { data: usersData } = useQuery({
    queryKey: ['users', 'all'],
    queryFn: () => userService.listUsers({ limit: MAX_PAGE_SIZE }),
  });

  const createMutation = useMutation({
    mutationFn: groupService.createGroup,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['groups'] });
      setIsCreateOpen(false);
      notify.success('Group created successfully');
    },
    onError: (error: Error) => notify.error(error.message),
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateGroupRequest }) => groupService.updateGroup(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['groups'] });
      setEditingGroup(null);
      notify.success('Group updated successfully');
    },
    onError: (error: Error) => notify.error(error.message),
  });

  const deleteMutation = useMutation({
    mutationFn: groupService.deleteGroup,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['groups'] });
      notify.success('Group deleted successfully');
    },
    onError: (error: Error) => notify.error(error.message),
  });

  const assignRoleMutation = useMutation({
    mutationFn: ({ id, roleIds }: { id: string; roleIds: string[] }) => groupService.addRoles(id, { roleIds }),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['groups'] });
      if (managingRolesGroup && managingRolesGroup.id === variables.id) {
        setManagingRolesGroup({
          ...managingRolesGroup,
          roleIds: [...(managingRolesGroup.roleIds || []), ...variables.roleIds],
        });
      }
      notify.success('Role(s) assigned successfully');
    },
    onError: (error: Error) => notify.error(error.message),
  });

  const unassignRoleMutation = useMutation({
    mutationFn: ({ id, roleIds }: { id: string; roleIds: string[] }) =>
      groupService.removeRoles(id, { roleIds }),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['groups'] });
      if (managingRolesGroup && managingRolesGroup.id === variables.id) {
        setManagingRolesGroup({
          ...managingRolesGroup,
          roleIds: (managingRolesGroup.roleIds || []).filter(id => !variables.roleIds.includes(id)),
        });
      }
      notify.success('Role(s) unassigned successfully');
    },
    onError: (error: Error) => notify.error(error.message),
  });

  const addUsersToGroupMutation = useMutation({
    mutationFn: ({ id, userIds }: { id: string; userIds: string[] }) => groupService.addUsers(id, { userIds }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['groups'] });
      queryClient.invalidateQueries({ queryKey: ['users', 'all'] });
      notify.success('User added successfully');
    },
    onError: (error: Error) => notify.error(error.message),
  });

  const removeUsersFromGroupMutation = useMutation({
    mutationFn: ({ id, userIds }: { id: string; userIds: string[] }) => groupService.removeUsers(id, { userIds }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['groups'] });
      queryClient.invalidateQueries({ queryKey: ['users', 'all'] });
      notify.success('User removed successfully');
    },
    onError: (error: Error) => notify.error(error.message),
  });

  const handleCreateSubmit = (values: GroupFormValues) => {
    createMutation.mutate({
      name: values.name,
      description: values.description,
      status: values.status || 'STATUS_UNSPECIFIED',
    });
  };

  const handleEditSubmit = (values: GroupFormValues) => {
    if (!editingGroup) return;

    const data: UpdateGroupRequest = {
      name: values.name,
      description: values.description,
    };

    if (values.status) {
      data.status = values.status;
    }

    updateMutation.mutate({
      id: editingGroup.id || '',
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
        let color: string;
        let bgColor: string;
        let text: string;

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
      render: (_: unknown, record: Group) => {
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
                {rolesData.roles.filter(role => record.roleIds?.includes(role.id ?? '')).slice(0, 2).map(r => r.name).join(', ')}
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
      render: (_: unknown, record: Group) => (
        <Space>
          <Button
            type="text"
            icon={<Key size={15} />}
            onClick={() => setManagingRolesGroup(record)}
            className="p-1.5 rounded-md text-purple-600 dark:text-purple-400 hover:bg-purple-50 dark:hover:bg-purple-900/20 transition-colors"
            title="Manage Roles"
          />
          <Button
            type="text"
            icon={<UserOutlined />}
            onClick={() => setManagingUsersGroup(record)}
            className="p-1.5 rounded-md text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/20 transition-colors"
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
            className="p-1.5 rounded-md text-emerald-600 dark:text-emerald-400 hover:bg-emerald-50 dark:hover:bg-emerald-900/20 transition-colors"
            title="Edit Group"
          />
          <Button
            type="text"
            icon={<DeleteOutlined />}
            onClick={() => {
              const id = record.id;
              if (!id) return;
              confirmDanger(
                {
                  title: 'Delete group',
                  content: 'Are you sure you want to delete this group?',
                },
                () => deleteMutation.mutate(id),
              );
            }}
            className="p-1.5 rounded-md text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors"
            title="Delete Group"
          />
        </Space>
      ),
    },
  ];

  // 検索フィルター

  // 統計データの計算
  const totalGroups = data?.groups?.length || 0;
  const activeGroups = data?.groups?.filter(g => g.status === 'STATUS_ACTIVE').length || 0;
  const averageRolesPerGroup = data?.groups?.length ?
    (data.groups.reduce((acc, group) => acc + (group.roleIds?.length || 0), 0) / data.groups.length).toFixed(1)
    : '0.0';
  const averageUsersPerGroup = data?.groups?.length ?
    (data.groups.reduce((acc, group) => {
      const groupUsers = usersData?.users?.filter(user => user.groupIds?.includes(group.id ?? '')).length || 0;
      return acc + groupUsers;
    }, 0) / data.groups.length).toFixed(1)
    : '0.0';

  return (
    <div className="space-y-6">
      {/* ヘッダーセクション */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold tracking-tight text-gray-900 dark:text-white">Groups</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400">Manage user groups and assign roles/permissions</p>
        </div>
        <div className="flex items-center gap-3">
          <Input
            placeholder="Search by name..."
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
            Add Group
          </Button>
        </div>
      </div>

      {/* 統計カード */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card className="shadow-sm dark:bg-gray-800 dark:border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm font-medium text-gray-500 dark:text-gray-400">Total Groups</div>
              <div className="text-2xl font-bold mt-1 text-gray-900 dark:text-white">{totalGroups}</div>
            </div>
            <div className="p-2 rounded-lg bg-blue-50 dark:bg-blue-900/20">
              <FolderOutlined style={{ fontSize: '20px' }} className="text-blue-600 dark:text-blue-400" />
            </div>
          </div>
        </Card>
        <Card className="shadow-sm dark:bg-gray-800 dark:border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm font-medium text-gray-500 dark:text-gray-400">Active Groups</div>
              <div className="text-2xl font-bold mt-1 text-gray-900 dark:text-white">{activeGroups}</div>
              <div className="text-sm mt-2 text-gray-500 dark:text-gray-400">
                {totalGroups > 0 ? `${Math.round((activeGroups / totalGroups) * 100)}% active` : 'No groups'}
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
              <div className="text-sm font-medium text-gray-500 dark:text-gray-400">Avg. Roles/Group</div>
              <div className="text-2xl font-bold mt-1 text-gray-900 dark:text-white">{averageRolesPerGroup}</div>
              <div className="text-sm mt-2 text-gray-500 dark:text-gray-400">
                {rolesData?.roles?.length || 0} total roles
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
              <div className="text-sm font-medium text-gray-500 dark:text-gray-400">Avg. Users/Group</div>
              <div className="text-2xl font-bold mt-1 text-gray-900 dark:text-white">{averageUsersPerGroup}</div>
              <div className="text-sm mt-2 text-gray-500 dark:text-gray-400">
                {usersData?.users?.length || 0} total users
              </div>
            </div>
            <div className="p-2 rounded-lg bg-orange-50 dark:bg-orange-900/20">
              <Users className="h-5 w-5 text-orange-600 dark:text-orange-400" />
            </div>
          </div>
        </Card>
      </div>

      {/* グループテーブル */}
      <Card className="shadow-sm dark:bg-gray-800 dark:border-gray-700">
        <Table
          columns={columns}
          dataSource={data?.groups}
          loading={isLoading}
          rowKey="id"
          locale={{
            emptyText: error ? 'Failed to load groups' : 'No groups found'
          }}
          pagination={pagination.tableProps(Number(data?.total ?? 0), 'groups')}
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

      <ManageRolesDrawer
        group={managingRolesGroup}
        allRoles={rolesData?.roles ?? []}
        onClose={() => setManagingRolesGroup(null)}
        onAssign={(groupId, roleIds) => assignRoleMutation.mutate({ id: groupId, roleIds })}
        onUnassign={(groupId, roleIds) => unassignRoleMutation.mutate({ id: groupId, roleIds })}
        isAssigning={assignRoleMutation.isPending}
        isUnassigning={unassignRoleMutation.isPending}
      />

      <ManageUsersDrawer
        group={managingUsersGroup}
        allUsers={usersData?.users ?? []}
        onClose={() => setManagingUsersGroup(null)}
        onAdd={(groupId, userIds) => addUsersToGroupMutation.mutate({ id: groupId, userIds })}
        onRemove={(groupId, userIds) => {
          confirmDanger(
            {
              title: 'Remove users',
              content: `Remove ${userIds.length} user(s) from this group?`,
              okText: 'Remove',
            },
            () => removeUsersFromGroupMutation.mutate({ id: groupId, userIds }),
          );
        }}
        isAdding={addUsersToGroupMutation.isPending}
        isRemoving={removeUsersFromGroupMutation.isPending}
      />
    </div>
  );
}

// ── Manage Roles Drawer ────────────────────────────────────────────────────────

interface ManageRolesDrawerProps {
  group: Group | null;
  allRoles: Role[];
  onClose: () => void;
  onAssign: (groupId: string, roleIds: string[]) => void;
  onUnassign: (groupId: string, roleIds: string[]) => void;
  isAssigning: boolean;
  isUnassigning: boolean;
}

function ManageRolesDrawer({ group, allRoles, onClose, onAssign, onUnassign, isAssigning, isUnassigning }: ManageRolesDrawerProps) {
  const [assignedSearch, setAssignedSearch] = useState('');
  const [availableSearch, setAvailableSearch] = useState('');
  const [selectedAssigned, setSelectedAssigned] = useState<string[]>([]);
  const [selectedAvailable, setSelectedAvailable] = useState<string[]>([]);

  const groupId = group?.id ?? '';
  const assignedIds = group?.roleIds ?? [];

  const assignedRoles = useMemo(
    () => allRoles.filter(r => assignedIds.includes(r.id ?? '')),
    [allRoles, assignedIds],
  );

  const availableRoles = useMemo(
    () => allRoles.filter(r => !assignedIds.includes(r.id ?? '')),
    [allRoles, assignedIds],
  );

  const filteredAssigned = useMemo(() => {
    const q = assignedSearch.toLowerCase();
    return q ? assignedRoles.filter(r => r.name?.toLowerCase().includes(q)) : assignedRoles;
  }, [assignedRoles, assignedSearch]);

  const filteredAvailable = useMemo(() => {
    const q = availableSearch.toLowerCase();
    return q ? availableRoles.filter(r => r.name?.toLowerCase().includes(q)) : availableRoles;
  }, [availableRoles, availableSearch]);

  const handleAssign = () => {
    if (!groupId || selectedAvailable.length === 0) return;
    onAssign(groupId, selectedAvailable);
    setSelectedAvailable([]);
  };

  const handleUnassign = () => {
    if (!groupId || selectedAssigned.length === 0) return;
    onUnassign(groupId, selectedAssigned);
    setSelectedAssigned([]);
  };

  const toggleAssigned = (id: string) =>
    setSelectedAssigned(prev => prev.includes(id) ? prev.filter(x => x !== id) : [...prev, id]);

  const toggleAvailable = (id: string) =>
    setSelectedAvailable(prev => prev.includes(id) ? prev.filter(x => x !== id) : [...prev, id]);

  const toggleAllAssigned = (checked: boolean) =>
    setSelectedAssigned(checked ? filteredAssigned.map(r => r.id ?? '') : []);

  const toggleAllAvailable = (checked: boolean) =>
    setSelectedAvailable(checked ? filteredAvailable.map(r => r.id ?? '') : []);

  return (
    <Drawer
      title={
        <div className="flex items-center gap-2">
          <Key size={15} />
          <span>Manage Roles — {group?.name}</span>
        </div>
      }
      open={!!group}
      onClose={onClose}
      width={760}
      destroyOnHidden
      footer={null}
    >
      {group && (
        <div className="flex gap-3 h-full" style={{ minHeight: 0 }}>
          {/* Left: assigned roles */}
          <div className="flex flex-col flex-1 min-w-0">
            <div className="flex items-center mb-2">
              <span className="text-sm font-semibold text-gray-700 dark:text-gray-300 uppercase tracking-wide">
                Assigned
                <span className="ml-2 inline-flex items-center justify-center rounded-full bg-purple-100 dark:bg-purple-900/40 text-purple-700 dark:text-purple-300 text-xs px-2 py-0.5">
                  {assignedRoles.length}
                </span>
              </span>
            </div>
            <Input
              placeholder="Search assigned..."
              prefix={<SearchOutlined className="text-gray-400" />}
              value={assignedSearch}
              onChange={e => setAssignedSearch(e.target.value)}
              allowClear
              size="small"
              className="mb-2"
            />
            <div className="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden flex flex-col flex-1">
              {filteredAssigned.length > 0 && (
                <div className="flex items-center gap-2 px-3 py-1.5 bg-gray-50 dark:bg-gray-800/60 border-b border-gray-200 dark:border-gray-700">
                  <Checkbox
                    checked={selectedAssigned.length === filteredAssigned.length && filteredAssigned.length > 0}
                    indeterminate={selectedAssigned.length > 0 && selectedAssigned.length < filteredAssigned.length}
                    onChange={e => toggleAllAssigned(e.target.checked)}
                  />
                  <span className="text-xs text-gray-500 dark:text-gray-400">
                    {selectedAssigned.length > 0 ? `${selectedAssigned.length} selected` : 'Select all'}
                  </span>
                </div>
              )}
              <div className="overflow-y-auto flex-1" style={{ maxHeight: 'calc(100vh - 280px)' }}>
                {filteredAssigned.length === 0 ? (
                  <p className="text-sm text-gray-400 py-8 text-center">
                    {assignedSearch ? 'No matches' : 'No roles assigned'}
                  </p>
                ) : (
                  filteredAssigned.map(role => (
                    <RoleRow
                      key={role.id}
                      role={role}
                      selected={selectedAssigned.includes(role.id ?? '')}
                      onToggle={() => toggleAssigned(role.id ?? '')}
                      variant="assigned"
                    />
                  ))
                )}
              </div>
            </div>
          </div>

          {/* Center: transfer buttons */}
          <div className="flex flex-col items-center justify-center gap-2 pt-10">
            <Tooltip title="Assign selected">
              <Button
                shape="circle"
                icon={<ArrowLeftOutlined />}
                disabled={selectedAvailable.length === 0}
                loading={isAssigning}
                onClick={handleAssign}
              />
            </Tooltip>
            <Tooltip title="Unassign selected">
              <Button
                shape="circle"
                danger
                icon={<ArrowRightOutlined />}
                disabled={selectedAssigned.length === 0}
                loading={isUnassigning}
                onClick={handleUnassign}
              />
            </Tooltip>
          </div>

          {/* Right: available roles */}
          <div className="flex flex-col flex-1 min-w-0">
            <div className="flex items-center mb-2">
              <span className="text-sm font-semibold text-gray-700 dark:text-gray-300 uppercase tracking-wide">
                Available
                <span className="ml-2 inline-flex items-center justify-center rounded-full bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 text-xs px-2 py-0.5">
                  {availableRoles.length}
                </span>
              </span>
            </div>
            <Input
              placeholder="Search available..."
              prefix={<SearchOutlined className="text-gray-400" />}
              value={availableSearch}
              onChange={e => setAvailableSearch(e.target.value)}
              allowClear
              size="small"
              className="mb-2"
            />
            <div className="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden flex flex-col flex-1">
              {filteredAvailable.length > 0 && (
                <div className="flex items-center gap-2 px-3 py-1.5 bg-gray-50 dark:bg-gray-800/60 border-b border-gray-200 dark:border-gray-700">
                  <Checkbox
                    checked={selectedAvailable.length === filteredAvailable.length && filteredAvailable.length > 0}
                    indeterminate={selectedAvailable.length > 0 && selectedAvailable.length < filteredAvailable.length}
                    onChange={e => toggleAllAvailable(e.target.checked)}
                  />
                  <span className="text-xs text-gray-500 dark:text-gray-400">
                    {selectedAvailable.length > 0 ? `${selectedAvailable.length} selected` : 'Select all'}
                  </span>
                </div>
              )}
              <div className="overflow-y-auto flex-1" style={{ maxHeight: 'calc(100vh - 280px)' }}>
                {filteredAvailable.length === 0 ? (
                  <p className="text-sm text-gray-400 py-8 text-center">
                    {availableSearch ? 'No matches' : 'All roles are assigned'}
                  </p>
                ) : (
                  filteredAvailable.map(role => (
                    <RoleRow
                      key={role.id}
                      role={role}
                      selected={selectedAvailable.includes(role.id ?? '')}
                      onToggle={() => toggleAvailable(role.id ?? '')}
                      variant="available"
                    />
                  ))
                )}
              </div>
            </div>
          </div>
        </div>
      )}
    </Drawer>
  );
}

interface RoleRowProps {
  role: Role;
  selected: boolean;
  onToggle: () => void;
  variant: 'assigned' | 'available';
}

function RoleRow({ role, selected, onToggle, variant }: RoleRowProps) {
  const iconBg = variant === 'assigned'
    ? 'bg-purple-100 dark:bg-purple-900/40'
    : 'bg-gray-100 dark:bg-gray-700';
  const iconColor = variant === 'assigned'
    ? 'text-purple-600 dark:text-purple-400'
    : 'text-gray-400';

  return (
    <div
      className={`flex items-center gap-2 px-3 py-2 cursor-pointer border-b border-gray-100 dark:border-gray-700/50 last:border-0 transition-colors ${
        selected
          ? 'bg-purple-50 dark:bg-purple-900/20'
          : 'hover:bg-gray-50 dark:hover:bg-gray-800/40'
      }`}
      onClick={onToggle}
    >
      <Checkbox checked={selected} onChange={onToggle} onClick={e => e.stopPropagation()} />
      <div className={`h-7 w-7 rounded-full ${iconBg} flex items-center justify-center flex-shrink-0`}>
        <Key size={13} className={iconColor} />
      </div>
      <div className="min-w-0 flex-1">
        <div className="text-sm font-medium text-gray-900 dark:text-white truncate">{role.name}</div>
        {role.description && (
          <div className="text-xs text-gray-500 dark:text-gray-400 truncate">{role.description}</div>
        )}
      </div>
    </div>
  );
}

// ── Manage Users Drawer ────────────────────────────────────────────────────────

interface ManageUsersDrawerProps {
  group: Group | null;
  allUsers: User[];
  onClose: () => void;
  onAdd: (groupId: string, userIds: string[]) => void;
  onRemove: (groupId: string, userIds: string[]) => void;
  isAdding: boolean;
  isRemoving: boolean;
}

function ManageUsersDrawer({ group, allUsers, onClose, onAdd, onRemove, isAdding, isRemoving }: ManageUsersDrawerProps) {
  const [memberSearch, setMemberSearch] = useState('');
  const [nonMemberSearch, setNonMemberSearch] = useState('');
  const [selectedMembers, setSelectedMembers] = useState<string[]>([]);
  const [selectedNonMembers, setSelectedNonMembers] = useState<string[]>([]);

  const groupId = group?.id ?? '';

  const members = useMemo(
    () => allUsers.filter(u => u.groupIds?.includes(groupId)),
    [allUsers, groupId],
  );

  const nonMembers = useMemo(
    () => allUsers.filter(u => !u.groupIds?.includes(groupId)),
    [allUsers, groupId],
  );

  const filteredMembers = useMemo(() => {
    const q = memberSearch.toLowerCase();
    return q ? members.filter(u => u.username?.toLowerCase().includes(q) || u.email?.toLowerCase().includes(q)) : members;
  }, [members, memberSearch]);

  const filteredNonMembers = useMemo(() => {
    const q = nonMemberSearch.toLowerCase();
    return q ? nonMembers.filter(u => u.username?.toLowerCase().includes(q) || u.email?.toLowerCase().includes(q)) : nonMembers;
  }, [nonMembers, nonMemberSearch]);

  const handleAdd = () => {
    if (!groupId || selectedNonMembers.length === 0) return;
    onAdd(groupId, selectedNonMembers);
    setSelectedNonMembers([]);
  };

  const handleRemove = () => {
    if (!groupId || selectedMembers.length === 0) return;
    onRemove(groupId, selectedMembers);
    setSelectedMembers([]);
  };

  const toggleMember = (id: string) =>
    setSelectedMembers(prev => prev.includes(id) ? prev.filter(x => x !== id) : [...prev, id]);

  const toggleNonMember = (id: string) =>
    setSelectedNonMembers(prev => prev.includes(id) ? prev.filter(x => x !== id) : [...prev, id]);

  const toggleAllMembers = (checked: boolean) =>
    setSelectedMembers(checked ? filteredMembers.map(u => u.id ?? '') : []);

  const toggleAllNonMembers = (checked: boolean) =>
    setSelectedNonMembers(checked ? filteredNonMembers.map(u => u.id ?? '') : []);

  return (
    <Drawer
      title={
        <div className="flex items-center gap-2">
          <UserOutlined />
          <span>Manage Users — {group?.name}</span>
        </div>
      }
      open={!!group}
      onClose={onClose}
      width={760}
      destroyOnHidden
      footer={null}
    >
      {group && (
        <div className="flex gap-3 h-full" style={{ minHeight: 0 }}>
          {/* Left panel: current members */}
          <div className="flex flex-col flex-1 min-w-0">
            <div className="flex items-center mb-2">
              <span className="text-sm font-semibold text-gray-700 dark:text-gray-300 uppercase tracking-wide">
                Members
                <span className="ml-2 inline-flex items-center justify-center rounded-full bg-blue-100 dark:bg-blue-900/40 text-blue-700 dark:text-blue-300 text-xs px-2 py-0.5">
                  {members.length}
                </span>
              </span>
            </div>
            <Input
              placeholder="Search members..."
              prefix={<SearchOutlined className="text-gray-400" />}
              value={memberSearch}
              onChange={e => setMemberSearch(e.target.value)}
              allowClear
              size="small"
              className="mb-2"
            />
            <div className="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden flex flex-col flex-1">
              {filteredMembers.length > 0 && (
                <div className="flex items-center gap-2 px-3 py-1.5 bg-gray-50 dark:bg-gray-800/60 border-b border-gray-200 dark:border-gray-700">
                  <Checkbox
                    checked={selectedMembers.length === filteredMembers.length && filteredMembers.length > 0}
                    indeterminate={selectedMembers.length > 0 && selectedMembers.length < filteredMembers.length}
                    onChange={e => toggleAllMembers(e.target.checked)}
                  />
                  <span className="text-xs text-gray-500 dark:text-gray-400">
                    {selectedMembers.length > 0 ? `${selectedMembers.length} selected` : 'Select all'}
                  </span>
                </div>
              )}
              <div className="overflow-y-auto flex-1" style={{ maxHeight: 'calc(100vh - 280px)' }}>
                {filteredMembers.length === 0 ? (
                  <p className="text-sm text-gray-400 py-8 text-center">
                    {memberSearch ? 'No matches' : 'No members yet'}
                  </p>
                ) : (
                  filteredMembers.map(user => (
                    <UserRow
                      key={user.id}
                      user={user}
                      selected={selectedMembers.includes(user.id ?? '')}
                      onToggle={() => toggleMember(user.id ?? '')}
                      variant="member"
                    />
                  ))
                )}
              </div>
            </div>
          </div>

          {/* Center: transfer buttons */}
          <div className="flex flex-col items-center justify-center gap-2 pt-10">
            <Tooltip title="Add selected">
              <Button
                shape="circle"
                icon={<ArrowLeftOutlined />}
                disabled={selectedNonMembers.length === 0}
                loading={isAdding}
                onClick={handleAdd}
              />
            </Tooltip>
            <Tooltip title="Remove selected">
              <Button
                shape="circle"
                danger
                icon={<ArrowRightOutlined />}
                disabled={selectedMembers.length === 0}
                loading={isRemoving}
                onClick={handleRemove}
              />
            </Tooltip>
          </div>

          {/* Right panel: non-members */}
          <div className="flex flex-col flex-1 min-w-0">
            <div className="flex items-center mb-2">
              <span className="text-sm font-semibold text-gray-700 dark:text-gray-300 uppercase tracking-wide">
                Available
                <span className="ml-2 inline-flex items-center justify-center rounded-full bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 text-xs px-2 py-0.5">
                  {nonMembers.length}
                </span>
              </span>
            </div>
            <Input
              placeholder="Search users..."
              prefix={<SearchOutlined className="text-gray-400" />}
              value={nonMemberSearch}
              onChange={e => setNonMemberSearch(e.target.value)}
              allowClear
              size="small"
              className="mb-2"
            />
            <div className="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden flex flex-col flex-1">
              {filteredNonMembers.length > 0 && (
                <div className="flex items-center gap-2 px-3 py-1.5 bg-gray-50 dark:bg-gray-800/60 border-b border-gray-200 dark:border-gray-700">
                  <Checkbox
                    checked={selectedNonMembers.length === filteredNonMembers.length && filteredNonMembers.length > 0}
                    indeterminate={selectedNonMembers.length > 0 && selectedNonMembers.length < filteredNonMembers.length}
                    onChange={e => toggleAllNonMembers(e.target.checked)}
                  />
                  <span className="text-xs text-gray-500 dark:text-gray-400">
                    {selectedNonMembers.length > 0 ? `${selectedNonMembers.length} selected` : 'Select all'}
                  </span>
                </div>
              )}
              <div className="overflow-y-auto flex-1" style={{ maxHeight: 'calc(100vh - 280px)' }}>
                {filteredNonMembers.length === 0 ? (
                  <p className="text-sm text-gray-400 py-8 text-center">
                    {nonMemberSearch ? 'No matches' : 'All users are members'}
                  </p>
                ) : (
                  filteredNonMembers.map(user => (
                    <UserRow
                      key={user.id}
                      user={user}
                      selected={selectedNonMembers.includes(user.id ?? '')}
                      onToggle={() => toggleNonMember(user.id ?? '')}
                      variant="available"
                    />
                  ))
                )}
              </div>
            </div>
          </div>
        </div>
      )}
    </Drawer>
  );
}

interface UserRowProps {
  user: User;
  selected: boolean;
  onToggle: () => void;
  variant: 'member' | 'available';
}

function UserRow({ user, selected, onToggle, variant }: UserRowProps) {
  const avatarBg = variant === 'member'
    ? 'bg-blue-100 dark:bg-blue-900/40'
    : 'bg-gray-100 dark:bg-gray-700';
  const avatarColor = variant === 'member'
    ? 'text-blue-600 dark:text-blue-400'
    : 'text-gray-400';

  return (
    <div
      className={`flex items-center gap-2 px-3 py-2 cursor-pointer border-b border-gray-100 dark:border-gray-700/50 last:border-0 transition-colors ${
        selected
          ? 'bg-blue-50 dark:bg-blue-900/20'
          : 'hover:bg-gray-50 dark:hover:bg-gray-800/40'
      }`}
      onClick={onToggle}
    >
      <Checkbox checked={selected} onChange={onToggle} onClick={e => e.stopPropagation()} />
      <div className={`h-7 w-7 rounded-full ${avatarBg} flex items-center justify-center flex-shrink-0`}>
        <UserOutlined className={`${avatarColor} text-xs`} />
      </div>
      <div className="min-w-0 flex-1">
        <div className="text-sm font-medium text-gray-900 dark:text-white truncate">{user.username}</div>
        <div className="text-xs text-gray-500 dark:text-gray-400 truncate">{user.email}</div>
      </div>
    </div>
  );
}
