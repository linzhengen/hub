import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { permissionService, Permission, CreatePermissionRequest, UpdatePermissionRequest } from '@/services/permission';
import { Button, Modal, Input, Table, Form, Space } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import { toast } from 'sonner';

export function Permissions() {
  const queryClient = useQueryClient();
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [editingPermission, setEditingPermission] = useState<Permission | null>(null);
  const [createForm] = Form.useForm();
  const [editForm] = Form.useForm();

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
      title: 'Resource ID',
      dataIndex: 'resourceId',
      key: 'resourceId',
    },
    {
      title: 'Verb',
      dataIndex: 'verb',
      key: 'verb',
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
          />
          <Button
            type="text"
            icon={<DeleteOutlined />}
            danger
            onClick={() => {
              if (confirm('Are you sure you want to delete this permission?')) {
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
        <h2 className="text-2xl font-bold tracking-tight">Permissions</h2>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => setIsCreateOpen(true)}
        >
          Add Permission
        </Button>
      </div>

      <Table
        columns={columns}
        dataSource={data?.permissions}
        loading={isLoading}
        rowKey="id"
        locale={{
          emptyText: error ? 'Failed to load permissions' : 'No permissions found'
        }}
      />

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
