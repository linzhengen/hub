import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { permissionService, Permission, CreatePermissionRequest, UpdatePermissionRequest } from '@/services/permission.ts';
import { Button, Modal, Input, Table, Form, Space, Card, Tag } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, SearchOutlined, KeyOutlined, LockOutlined } from '@ant-design/icons';
import { toast } from 'sonner';
import { cn } from '@/lib/utils.ts';
import { Key, TrendingUp, Shield } from 'lucide-react';

export function Permissions() {
  const queryClient = useQueryClient();
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [editingPermission, setEditingPermission] = useState<Permission | null>(null);
  const [createForm] = Form.useForm();
  const [editForm] = Form.useForm();
  const [searchText, setSearchText] = useState('');

  const { data, isLoading, error } = useQuery({
    queryKey: ['permissions'],
    queryFn: () => permissionService.listPermissions(),
  });

  const createMutation = useMutation({
    mutationFn: permissionService.createPermission,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['permissions'] });
      setIsCreateOpen(false);
      toast.success('Permission created successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdatePermissionRequest }) => permissionService.updatePermission(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['permissions'] });
      setEditingPermission(null);
      toast.success('Permission updated successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const deleteMutation = useMutation({
    mutationFn: permissionService.deletePermission,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['permissions'] });
      toast.success('Permission deleted successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const handleCreateSubmit = (values: any) => {
    createMutation.mutate({
      resourceId: values.resourceId,
      verb: values.verb,
      description: values.description,
    });
  };

  const handleEditSubmit = (values: any) => {
    if (!editingPermission) return;

    updateMutation.mutate({
      id: editingPermission.id,
      data: {
        verb: values.verb,
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
      title: 'Resource',
      dataIndex: 'resourceId',
      key: 'resourceId',
      render: (resourceId: string) => (
        <div className="flex items-center gap-2">
          <LockOutlined className="text-gray-500 dark:text-gray-400" style={{ fontSize: '14px' }} />
          <span className="text-gray-900 dark:text-white font-medium">{resourceId}</span>
        </div>
      ),
    },
    {
      title: 'Action',
      dataIndex: 'verb',
      key: 'verb',
      render: (verb: string) => {
        const style = getActionStyle(verb);
        const actionLower = verb.toLowerCase();
        return (
          <span
            className={cn(
              "inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border",
              actionLower.includes('read') || actionLower === 'view' ? "bg-emerald-100 text-emerald-800 border-emerald-200 dark:bg-emerald-500/10 dark:text-emerald-400 dark:border-emerald-500/20" :
              actionLower.includes('write') || actionLower.includes('create') || actionLower.includes('update') ? "bg-blue-100 text-blue-800 border-blue-200 dark:bg-blue-500/10 dark:text-blue-400 dark:border-blue-500/20" :
              actionLower.includes('delete') || actionLower.includes('remove') ? "bg-red-100 text-red-800 border-red-200 dark:bg-red-500/10 dark:text-red-400 dark:border-red-500/20" :
              "bg-gray-100 text-gray-800 border-gray-200 dark:bg-gray-500/10 dark:text-gray-400 dark:border-gray-500/20"
            )}
          >
            {verb.toUpperCase()}
          </span>
        );
      },
    },
    {
      title: 'Description',
      dataIndex: 'description',
      key: 'description',
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: any, record: Permission) => (
        <Space>
          <Button
            type="text"
            icon={<EditOutlined />}
            onClick={() => {
              setEditingPermission(record);
              editForm.setFieldsValue({
                verb: record.verb,
                description: record.description,
              });
            }}
            className="p-1.5 rounded-md text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/20 transition-colors"
            title="Edit Permission"
          />
          <Button
            type="text"
            icon={<DeleteOutlined />}
            onClick={() => {
              if (confirm('Are you sure you want to delete this permission?')) {
                deleteMutation.mutate(record.id);
              }
            }}
            className="p-1.5 rounded-md text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors"
            title="Delete Permission"
          />
        </Space>
      ),
    },
  ];

  // 検索フィルター
  const filteredPermissions = data?.permissions?.filter(permission =>
    !searchText ||
    permission.name.toLowerCase().includes(searchText.toLowerCase()) ||
    permission.resourceId.toLowerCase().includes(searchText.toLowerCase()) ||
    permission.verb.toLowerCase().includes(searchText.toLowerCase())
  );

  // 統計データの計算
  const totalPermissions = data?.permissions?.length || 0;
  const readPermissions = data?.permissions?.filter(p => p.verb === 'read' || p.verb === 'READ').length || 0;
  const writePermissions = data?.permissions?.filter(p => p.verb === 'write' || p.verb === 'WRITE').length || 0;
  const deletePermissions = data?.permissions?.filter(p => p.verb === 'delete' || p.verb === 'DELETE').length || 0;

  // アクションタイプに基づくスタイル
  const getActionStyle = (verb: string) => {
    const actionLower = verb.toLowerCase();
    if (actionLower.includes('read') || actionLower === 'view') {
      return { color: '#059669', bgColor: '#d1fae5', borderColor: '#a7f3d0' };
    } else if (actionLower.includes('write') || actionLower.includes('create') || actionLower.includes('update')) {
      return { color: '#3b82f6', bgColor: '#dbeafe', borderColor: '#93c5fd' };
    } else if (actionLower.includes('delete') || actionLower.includes('remove')) {
      return { color: '#dc2626', bgColor: '#fee2e2', borderColor: '#fca5a5' };
    } else {
      return { color: '#6b7280', bgColor: '#f3f4f6', borderColor: '#d1d5db' };
    }
  };

  return (
    <div className="space-y-6">
      {/* ヘッダーセクション */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold tracking-tight text-gray-900 dark:text-white">Permissions</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400">Define fine-grained access control permissions</p>
        </div>
        <div className="flex items-center gap-3">
          <Input
            placeholder="Search permissions..."
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
            Add Permission
          </Button>
        </div>
      </div>

      {/* 統計カード */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card className="shadow-sm dark:bg-gray-800 dark:border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm font-medium text-gray-500 dark:text-gray-400">Total Permissions</div>
              <div className="text-2xl font-bold mt-1 text-gray-900 dark:text-white">{totalPermissions}</div>
              <div className="flex items-center gap-1 mt-2">
                <TrendingUp className="h-4 w-4 text-green-500 dark:text-green-400" />
                <span className="text-sm text-green-600 dark:text-green-400">+12.3%</span>
                <span className="text-sm text-gray-500 dark:text-gray-400">from last month</span>
              </div>
            </div>
            <div className="p-2 rounded-lg bg-blue-50 dark:bg-blue-900/20">
              <Key className="h-5 w-5 text-blue-600 dark:text-blue-400" />
            </div>
          </div>
        </Card>
        <Card className="shadow-sm dark:bg-gray-800 dark:border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm font-medium text-gray-500 dark:text-gray-400">Read Permissions</div>
              <div className="text-2xl font-bold mt-1 text-gray-900 dark:text-white">{readPermissions}</div>
              <div className="text-sm mt-2 text-gray-500 dark:text-gray-400">
                {totalPermissions > 0 ? `${Math.round((readPermissions / totalPermissions) * 100)}% of total` : 'No permissions'}
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
              <div className="text-sm font-medium text-gray-500 dark:text-gray-400">Write Permissions</div>
              <div className="text-2xl font-bold mt-1 text-gray-900 dark:text-white">{writePermissions}</div>
              <div className="text-sm mt-2 text-gray-500 dark:text-gray-400">
                {totalPermissions > 0 ? `${Math.round((writePermissions / totalPermissions) * 100)}% of total` : 'No permissions'}
              </div>
            </div>
            <div className="p-2 rounded-lg bg-blue-50 dark:bg-blue-900/20">
              <div className="h-5 w-5 rounded-full bg-blue-500"></div>
            </div>
          </div>
        </Card>
        <Card className="shadow-sm dark:bg-gray-800 dark:border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm font-medium text-gray-500 dark:text-gray-400">Delete Permissions</div>
              <div className="text-2xl font-bold mt-1 text-gray-900 dark:text-white">{deletePermissions}</div>
              <div className="text-sm mt-2 text-gray-500 dark:text-gray-400">
                {totalPermissions > 0 ? `${Math.round((deletePermissions / totalPermissions) * 100)}% of total` : 'No permissions'}
              </div>
            </div>
            <div className="p-2 rounded-lg bg-red-50 dark:bg-red-900/20">
              <div className="h-5 w-5 rounded-full bg-red-500"></div>
            </div>
          </div>
        </Card>
      </div>

      {/* 権限テーブル */}
      <Card className="shadow-sm dark:bg-gray-800 dark:border-gray-700">
        <Table
          columns={columns}
          dataSource={filteredPermissions}
          loading={isLoading}
          rowKey="id"
          locale={{
            emptyText: error ? 'Failed to load permissions' : 'No permissions found'
          }}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => `${range[0]}-${range[1]} of ${total} permissions`,
          }}
        />
      </Card>

      <Modal
        title="Create New Permission"
        open={isCreateOpen}
        onCancel={() => setIsCreateOpen(false)}
        footer={null}
      >
        <Form
          form={createForm}
          layout="vertical"
          onFinish={handleCreateSubmit}
        >
          <div className="grid grid-cols-2 gap-4">
            <Form.Item
              name="resourceId"
              label="Resource ID"
              rules={[{ required: true, message: 'Please input resource ID!' }]}
            >
              <Input />
            </Form.Item>
            <Form.Item
              name="verb"
              label="Verb"
              rules={[{ required: true, message: 'Please input verb!' }]}
            >
              <Input />
            </Form.Item>
          </div>
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
        title="Edit Permission"
        open={!!editingPermission}
        onCancel={() => setEditingPermission(null)}
        footer={null}
      >
        {editingPermission && (
          <Form
            form={editForm}
            layout="vertical"
            onFinish={handleEditSubmit}
          >
            <Form.Item
              label="Name"
            >
              <Input value={editingPermission.name} disabled />
            </Form.Item>
            <div className="grid grid-cols-2 gap-4">
              <Form.Item
                label="Resource ID"
              >
                <Input value={editingPermission.resourceId} disabled />
              </Form.Item>
              <Form.Item
                name="verb"
                label="Verb"
                rules={[{ required: true, message: 'Please input verb!' }]}
              >
                <Input />
              </Form.Item>
            </div>
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
    </div>
  );
}
