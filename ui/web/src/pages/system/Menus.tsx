import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { resourceService, Resource, UpdateMenuResourceRequest, CreateMenuResourceRequest } from '@/services/resource.ts';
import { Button, Modal, Input, Table, Form, Select, Tag, Space, Card, InputNumber, TreeSelect } from 'antd';

const { Option } = Select;
import { PlusOutlined, EditOutlined, DeleteOutlined, SearchOutlined, FileOutlined } from '@ant-design/icons';
import { toast } from 'sonner';

export function Menus() {
  const queryClient = useQueryClient();
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [editingResource, setEditingResource] = useState<Resource | null>(null);
  const [createForm] = Form.useForm();
  const [editForm] = Form.useForm();
  const [searchText, setSearchText] = useState('');

  const { data, isLoading, error } = useQuery({
    queryKey: ['menu-resources'],
    queryFn: () => resourceService.listMenuResources(),
  });

  const createMutation = useMutation({
    mutationFn: resourceService.createMenuResource,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['menu-resources'] });
      setIsCreateOpen(false);
      createForm.resetFields();
      toast.success('Menu resource created successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateMenuResourceRequest }) => resourceService.updateMenuResource(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['menu-resources'] });
      setEditingResource(null);
      toast.success('Menu resource updated successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const deleteMutation = useMutation({
    mutationFn: resourceService.deleteResource,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['menu-resources'] });
      toast.success('Resource deleted successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const handleCreateSubmit = (values: any) => {
    createMutation.mutate({
      name: values.name,
      description: values.description,
      status: values.status || 'STATUS_ACTIVE',
      path: values.path,
      component: values.component,
      displayOrder: values.displayOrder,
      parentId: values.parentId,
    });
  };

  const handleEditSubmit = (values: any) => {
    if (!editingResource) return;

    updateMutation.mutate({
      id: editingResource.id,
      data: {
        name: values.name,
        description: values.description,
        status: values.status,
        path: values.path,
        component: values.component,
        displayOrder: values.displayOrder,
        parentId: values.parentId,
      }
    });
  };

  const columns = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      width: 200,
      render: (name: string, record: Resource) => (
        <div className="flex items-center gap-2">
          <span className="text-gray-900 dark:text-white font-medium">{name}</span>
        </div>
      ),
    },
    {
      title: 'Path',
      dataIndex: 'path',
      key: 'path',
      width: 150,
      render: (path: string) => (
        <code className="text-xs bg-gray-100 dark:bg-gray-800 px-1 py-0.5 rounded text-gray-600 dark:text-gray-400">
          {path || '-'}
        </code>
      ),
    },
    {
      title: 'Order',
      dataIndex: 'displayOrder',
      key: 'displayOrder',
      width: 80,
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => (
        <Tag color={status === 'STATUS_ACTIVE' ? 'green' : 'red'}>
          {status?.replace('STATUS_', '') || 'UNKNOWN'}
        </Tag>
      ),
    },
    {
      title: 'Description',
      dataIndex: 'description',
      key: 'description',
      width: 200,
      render: (description: string) => (
        <span className="text-gray-500 text-sm">{description || '-'}</span>
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 100,
      fixed: 'right' as const,
      render: (_: any, record: Resource) => (
        <Space>
          <Button
            type="text"
            icon={<EditOutlined />}
            onClick={() => {
              setEditingResource(record);
              editForm.setFieldsValue({
                name: record.name,
                description: record.description,
                status: record.status,
                path: record.path,
                component: record.component,
                displayOrder: record.displayOrder,
                parentId: record.parentId,
              });
            }}
            className="text-blue-600"
          />
          <Button
            type="text"
            icon={<DeleteOutlined />}
            onClick={() => {
              if (confirm('Are you sure you want to delete this resource?')) {
                deleteMutation.mutate(record.id);
              }
            }}
            className="text-red-600"
          />
        </Space>
      ),
    },
  ];

  const buildResourceTree = (resources: Resource[]): any[] => {
    const map = new Map();
    const roots: any[] = [];

    resources.forEach(res => {
      map.set(res.id, { ...res, children: [] });
    });

    resources.forEach(res => {
      const node = map.get(res.id);
      if (res.parentId && map.has(res.parentId)) {
        map.get(res.parentId).children.push(node);
      } else {
        roots.push(node);
      }
    });

    // Remove empty children arrays for Ant Design Table tree compatibility
    const cleanTree = (nodes: any[]) => {
      nodes.forEach(node => {
        if (node.children.length === 0) {
          delete node.children;
        } else {
          node.children.sort((a: any, b: any) => (a.displayOrder || 0) - (b.displayOrder || 0));
          cleanTree(node.children);
        }
      });
      return nodes.sort((a, b) => (a.displayOrder || 0) - (b.displayOrder || 0));
    };

    return cleanTree(roots);
  };

  const filteredResources = data?.resources?.filter(resource =>
    !searchText ||
    resource.name.toLowerCase().includes(searchText.toLowerCase()) ||
    (resource.path && resource.path.toLowerCase().includes(searchText.toLowerCase()))
  ) || [];

  const treeData = buildResourceTree(filteredResources);

  const treeSelectOptions = buildResourceTree(data?.resources || []);

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold tracking-tight text-gray-900 dark:text-white">Menu Resources</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400">Manage menu structure and UI components</p>
        </div>
        <div className="flex items-center gap-3">
          <Input
            placeholder="Search menu resources..."
            prefix={<SearchOutlined className="text-gray-400" />}
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
            Add Menu
          </Button>
        </div>
      </div>

      <Card className="shadow-sm overflow-hidden">
        <Table
          columns={columns}
          dataSource={treeData}
          loading={isLoading}
          rowKey="id"
          scroll={{ x: 800 }}
          expandable={{ defaultExpandAllRows: true }}
          locale={{
            emptyText: error ? 'Failed to load resources' : 'No menu resources found'
          }}
          pagination={false}
        />
      </Card>

      {/* Create Modal */}
      <Modal
        title="Create New Menu"
        open={isCreateOpen}
        onCancel={() => setIsCreateOpen(false)}
        footer={null}
        width={600}
      >
        <Form form={createForm} layout="vertical" onFinish={handleCreateSubmit}>
          <div className="grid grid-cols-2 gap-4">
            <Form.Item name="name" label="Name" rules={[{ required: true }]}>
              <Input />
            </Form.Item>
            <Form.Item name="parentId" label="Parent Menu">
              <TreeSelect
                showSearch
                style={{ width: '100%' }}
                popupStyle={{ maxHeight: 800, overflow: 'auto' }}
                placeholder="Select parent menu"
                allowClear
                treeData={treeSelectOptions}
                fieldNames={{ label: 'name', value: 'id', children: 'children' }}
              />
            </Form.Item>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <Form.Item name="path" label="Path">
              <Input placeholder="/example" />
            </Form.Item>
            <Form.Item name="component" label="Component">
              <Input placeholder="ExamplePage" />
            </Form.Item>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <Form.Item name="displayOrder" label="Display Order" initialValue={0}>
              <InputNumber className="w-full" />
            </Form.Item>
            <Form.Item name="status" label="Status" initialValue="STATUS_ACTIVE">
              <Select>
                <Option value="STATUS_ACTIVE">Active</Option>
                <Option value="STATUS_INACTIVE">Inactive</Option>
              </Select>
            </Form.Item>
          </div>
          <Form.Item name="description" label="Description">
            <Input.TextArea rows={2} />
          </Form.Item>
          <div className="flex justify-end gap-2">
            <Button onClick={() => setIsCreateOpen(false)}>Cancel</Button>
            <Button type="primary" htmlType="submit" loading={createMutation.isPending}>
              Create
            </Button>
          </div>
        </Form>
      </Modal>

      {/* Edit Modal */}
      <Modal
        title="Edit Menu"
        open={!!editingResource}
        onCancel={() => setEditingResource(null)}
        footer={null}
        width={600}
      >
        <Form form={editForm} layout="vertical" onFinish={handleEditSubmit}>
          <div className="grid grid-cols-2 gap-4">
            <Form.Item name="name" label="Name" rules={[{ required: true }]}>
              <Input />
            </Form.Item>
            <Form.Item name="parentId" label="Parent Menu">
              <TreeSelect
                showSearch
                style={{ width: '100%' }}
                dropdownStyle={{ maxHeight: 400, overflow: 'auto' }}
                placeholder="Select parent menu"
                allowClear
                treeData={treeSelectOptions}
                fieldNames={{ label: 'name', value: 'id', children: 'children' }}
              />
            </Form.Item>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <Form.Item name="path" label="Path">
              <Input />
            </Form.Item>
            <Form.Item name="component" label="Component">
              <Input />
            </Form.Item>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <Form.Item name="displayOrder" label="Display Order">
              <InputNumber className="w-full" />
            </Form.Item>
            <Form.Item name="status" label="Status">
              <Select>
                <Option value="STATUS_ACTIVE">Active</Option>
                <Option value="STATUS_INACTIVE">Inactive</Option>
              </Select>
            </Form.Item>
          </div>
          <Form.Item name="description" label="Description">
            <Input.TextArea rows={2} />
          </Form.Item>
          <div className="flex justify-end gap-2">
            <Button onClick={() => setEditingResource(null)}>Cancel</Button>
            <Button type="primary" htmlType="submit" loading={updateMutation.isPending}>
              Save Changes
            </Button>
          </div>
        </Form>
      </Modal>
    </div>
  );
}
