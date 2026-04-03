import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { resourceService, Resource, CreateResourceRequest, UpdateResourceRequest, ResourceType } from '@/services/resource';
import { Button, Modal, Input, Table, Form, Select, Tag, Space } from 'antd';

const { Option } = Select;
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import { toast } from 'sonner';

export function Resources() {
  const queryClient = useQueryClient();
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [editingResource, setEditingResource] = useState<Resource | null>(null);
  const [createForm] = Form.useForm();
  const [editForm] = Form.useForm();

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
    },
    {
      title: 'Type',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => {
        let color = 'default';
        let text = 'Unspecified';

        if (type === 'TYPE_MENU') {
          color = 'blue';
          text = 'Menu';
        } else if (type === 'TYPE_API') {
          color = 'green';
          text = 'API';
        }

        return <Tag color={color}>{text}</Tag>;
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
          />
          <Button
            type="text"
            icon={<DeleteOutlined />}
            danger
            onClick={() => {
              if (confirm('Are you sure you want to delete this resource?')) {
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
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold tracking-tight">Resources</h2>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => setIsCreateOpen(true)}
        >
          Add Resource
        </Button>
      </div>

      <Table
        columns={columns}
        dataSource={data?.resources}
        loading={isLoading}
        rowKey="id"
        locale={{
          emptyText: error ? 'Failed to load resources' : 'No resources found'
        }}
      />

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
