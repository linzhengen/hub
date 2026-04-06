import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { resourceService, Resource, CreateResourceRequest, UpdateResourceRequest, ResourceType } from '@/services/resource';
import { Button, Modal, Input, Table, Form, Select, Tag, Space, Card } from 'antd';

import { cn } from '@/lib/utils';
const { Option } = Select;
import { PlusOutlined, EditOutlined, DeleteOutlined, SearchOutlined, DatabaseOutlined, FileOutlined, ApiOutlined } from '@ant-design/icons';
import { toast } from 'sonner';
import { Database, Folder, TrendingUp, Server } from 'lucide-react';

export function Resources() {
  const queryClient = useQueryClient();
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [editingResource, setEditingResource] = useState<Resource | null>(null);
  const [createForm] = Form.useForm();
  const [editForm] = Form.useForm();
  const [searchText, setSearchText] = useState('');

  const { data, isLoading, error } = useQuery({
    queryKey: ['resources'],
    queryFn: () => resourceService.listResources(),
  });

  const createMutation = useMutation({
    mutationFn: resourceService.createResource,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['resources'] });
      setIsCreateOpen(false);
      toast.success('Resource created successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateResourceRequest }) => resourceService.updateResource(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['resources'] });
      setEditingResource(null);
      toast.success('Resource updated successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const deleteMutation = useMutation({
    mutationFn: resourceService.deleteResource,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['resources'] });
      toast.success('Resource deleted successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const handleCreateSubmit = (values: any) => {
    createMutation.mutate({
      name: values.name,
      type: values.type,
      description: values.description,
    });
  };

  const handleEditSubmit = (values: any) => {
    if (!editingResource) return;

    updateMutation.mutate({
      id: editingResource.id,
      data: {
        name: values.name,
        type: values.type,
        description: values.description,
      }
    });
  };

  const columns = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      render: (name: string) => (
        <div className="flex items-center gap-2">
          <span className="text-gray-900 dark:text-white font-medium">{name}</span>
        </div>
      ),
    },
    {
      title: 'Type',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => {
        const style = getResourceTypeStyle(type);
        const typeLower = type.toLowerCase();
        return (
          <div className="flex items-center gap-2">
            <div className="text-sm" style={{ color: style.color }}>
              {style.icon}
            </div>
            <span
              className={cn(
                "inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border",
                typeLower.includes('api') ? "bg-blue-100 text-blue-800 border-blue-200 dark:bg-blue-500/10 dark:text-blue-400 dark:border-blue-500/20" :
                typeLower.includes('menu') ? "bg-emerald-100 text-emerald-800 border-emerald-200 dark:bg-emerald-500/10 dark:text-emerald-400 dark:border-emerald-500/20" :
                "bg-gray-100 text-gray-800 border-gray-200 dark:bg-gray-500/10 dark:text-gray-400 dark:border-gray-500/20"
              )}
            >
              {type.replace('TYPE_', '').replace('_', ' ')}
            </span>
          </div>
        );
      },
    },
    {
      title: 'Description',
      dataIndex: 'description',
      key: 'description',
      render: (description: string) => (
        <span style={{ color: '#64748b' }}>{description || '-'}</span>
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: any, record: Resource) => (
        <Space>
          <Button
            type="text"
            icon={<EditOutlined />}
            onClick={() => {
              setEditingResource(record);
              editForm.setFieldsValue({
                name: record.name,
                type: record.type,
                description: record.description,
              });
            }}
            className="p-1.5 rounded-md text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/20 transition-colors"
            title="Edit Resource"
          />
          <Button
            type="text"
            icon={<DeleteOutlined />}
            onClick={() => {
              if (confirm('Are you sure you want to delete this resource?')) {
                deleteMutation.mutate(record.id);
              }
            }}
            className="p-1.5 rounded-md text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors"
            title="Delete Resource"
          />
        </Space>
      ),
    },
  ];

  // 検索フィルター
  const filteredResources = data?.resources?.filter(resource =>
    !searchText ||
    resource.name.toLowerCase().includes(searchText.toLowerCase()) ||
    (resource.identifier?.api && resource.identifier.api.toLowerCase().includes(searchText.toLowerCase())) ||
    (resource.identifier?.category && resource.identifier.category.toLowerCase().includes(searchText.toLowerCase())) ||
    resource.type.toLowerCase().includes(searchText.toLowerCase())
  );

  // 統計データの計算
  const totalResources = data?.resources?.length || 0;
  const apiResources = data?.resources?.filter(r => r.type === 'TYPE_API' || r.type.toLowerCase().includes('api')).length || 0;
  const menuResources = data?.resources?.filter(r => r.type === 'TYPE_MENU' || r.type.toLowerCase().includes('menu')).length || 0;
  const unspecifiedResources = data?.resources?.filter(r => r.type === 'TYPE_UNSPECIFIED' || !r.type).length || 0;

  // リソースタイプに基づくスタイル
  const getResourceTypeStyle = (type: string) => {
    const typeLower = type.toLowerCase();
    if (typeLower.includes('api')) {
      return { color: '#3b82f6', bgColor: '#dbeafe', borderColor: '#93c5fd', icon: <ApiOutlined /> };
    } else if (typeLower.includes('menu')) {
      return { color: '#059669', bgColor: '#d1fae5', borderColor: '#a7f3d0', icon: <FileOutlined /> };
    } else {
      return { color: '#6b7280', bgColor: '#f3f4f6', borderColor: '#d1d5db', icon: <DatabaseOutlined /> };
    }
  };

  return (
    <div className="space-y-6">
      {/* ヘッダーセクション */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold tracking-tight text-gray-900 dark:text-white">Resources</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400">Manage system resources and access controls</p>
        </div>
        <div className="flex items-center gap-3">
          <Input
            placeholder="Search resources..."
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
            Add Resource
          </Button>
        </div>
      </div>

      {/* 統計カード */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card className="shadow-sm dark:bg-gray-800 dark:border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm font-medium text-gray-500 dark:text-gray-400">Total Resources</div>
              <div className="text-2xl font-bold mt-1 text-gray-900 dark:text-white">{totalResources}</div>
              <div className="flex items-center gap-1 mt-2">
                <TrendingUp className="h-4 w-4 text-green-500 dark:text-green-400" />
                <span className="text-sm text-green-600 dark:text-green-400">+15.7%</span>
                <span className="text-sm text-gray-500 dark:text-gray-400">from last month</span>
              </div>
            </div>
            <div className="p-2 rounded-lg bg-blue-50 dark:bg-blue-900/20">
              <Database className="h-5 w-5 text-blue-600 dark:text-blue-400" />
            </div>
          </div>
        </Card>
        <Card className="shadow-sm dark:bg-gray-800 dark:border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm font-medium text-gray-500 dark:text-gray-400">API Resources</div>
              <div className="text-2xl font-bold mt-1 text-gray-900 dark:text-white">{apiResources}</div>
              <div className="text-sm mt-2 text-gray-500 dark:text-gray-400">
                {totalResources > 0 ? `${Math.round((apiResources / totalResources) * 100)}% of total` : 'No resources'}
              </div>
            </div>
            <div className="p-2 rounded-lg bg-purple-50 dark:bg-purple-900/20">
              <Server className="h-5 w-5 text-purple-600 dark:text-purple-400" />
            </div>
          </div>
        </Card>
        <Card className="shadow-sm dark:bg-gray-800 dark:border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm font-medium text-gray-500 dark:text-gray-400">Menu Resources</div>
              <div className="text-2xl font-bold mt-1 text-gray-900 dark:text-white">{menuResources}</div>
              <div className="text-sm mt-2 text-gray-500 dark:text-gray-400">
                {totalResources > 0 ? `${Math.round((menuResources / totalResources) * 100)}% of total` : 'No resources'}
              </div>
            </div>
            <div className="p-2 rounded-lg bg-green-50 dark:bg-green-900/20">
              <FileOutlined style={{ fontSize: '20px' }} className="text-green-600 dark:text-green-400" />
            </div>
          </div>
        </Card>
        <Card className="shadow-sm dark:bg-gray-800 dark:border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm font-medium text-gray-500 dark:text-gray-400">Unspecified Resources</div>
              <div className="text-2xl font-bold mt-1 text-gray-900 dark:text-white">{unspecifiedResources}</div>
              <div className="text-sm mt-2 text-gray-500 dark:text-gray-400">
                {totalResources > 0 ? `${Math.round((unspecifiedResources / totalResources) * 100)}% of total` : 'No resources'}
              </div>
            </div>
            <div className="p-2 rounded-lg bg-orange-50 dark:bg-orange-900/20">
              <DatabaseOutlined style={{ fontSize: '20px' }} className="text-orange-600 dark:text-orange-400" />
            </div>
          </div>
        </Card>
      </div>

      {/* リソーステーブル */}
      <Card className="shadow-sm dark:bg-gray-800 dark:border-gray-700">
        <Table
          columns={columns}
          dataSource={filteredResources}
          loading={isLoading}
          rowKey="id"
          locale={{
            emptyText: error ? 'Failed to load resources' : 'No resources found'
          }}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => `${range[0]}-${range[1]} of ${total} resources`,
          }}
        />
      </Card>

      <Modal
        title="Create New Resource"
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
              name="name"
              label="Name"
              rules={[{ required: true, message: 'Please input resource name!' }]}
            >
              <Input />
            </Form.Item>
            <Form.Item
              name="type"
              label="Type"
              rules={[{ required: true, message: 'Please select resource type!' }]}
              initialValue="TYPE_UNSPECIFIED"
            >
              <Select>
                <Option value="TYPE_UNSPECIFIED">Unspecified</Option>
                <Option value="TYPE_MENU">Menu</Option>
                <Option value="TYPE_API">API</Option>
              </Select>
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
        title="Edit Resource"
        open={!!editingResource}
        onCancel={() => setEditingResource(null)}
        footer={null}
      >
        {editingResource && (
          <Form
            form={editForm}
            layout="vertical"
            onFinish={handleEditSubmit}
          >
            <div className="grid grid-cols-2 gap-4">
              <Form.Item
                name="name"
                label="Name"
                rules={[{ required: true, message: 'Please input resource name!' }]}
              >
                <Input />
              </Form.Item>
              <Form.Item
                name="type"
                label="Type"
                rules={[{ required: true, message: 'Please select resource type!' }]}
              >
                <Select>
                  <Option value="TYPE_UNSPECIFIED">Unspecified</Option>
                  <Option value="TYPE_MENU">Menu</Option>
                  <Option value="TYPE_API">API</Option>
                </Select>
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
