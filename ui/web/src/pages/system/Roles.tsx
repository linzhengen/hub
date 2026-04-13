import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { roleService, Role, CreateRoleRequest, UpdateRoleRequest } from '@/services/role.ts';
import { permissionService, Permission } from '@/services/permission.ts';
import { resourceService } from '@/services/resource.ts';
import { Button, Modal, Input, Table, Form, Space, Card, TreeSelect, Tag, Tooltip } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, KeyOutlined, SearchOutlined, SafetyOutlined } from '@ant-design/icons';
import { toast } from 'sonner';
import { Shield } from 'lucide-react';

export function Roles() {
  const queryClient = useQueryClient();
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [editingRole, setEditingRole] = useState<Role | null>(null);
  const [createForm] = Form.useForm();
  const [editForm] = Form.useForm();
  const [searchText, setSearchText] = useState('');

  const { data, isLoading, error } = useQuery({
    queryKey: ['roles'],
    queryFn: () => roleService.listRoles(),
  });

  const { data: permissionsData } = useQuery({
    queryKey: ['permissions'],
    queryFn: () => permissionService.listPermissions(),
  });

  const { data: resourcesData } = useQuery({
    queryKey: ['resources'],
    queryFn: () => resourceService.listResources(),
  });

  const resourceMap = React.useMemo(() => {
    const map = new Map<string, string>();
    resourcesData?.resources?.forEach(r => {
      if (r.id) {
        // identifier が存在すれば、category + api の形式などを検討
        // ここでは、ユーザーに分かりやすい identifier を優先的に表示
        const iden = r.identifier;
        const displayName = iden ? (iden.category && iden.api ? `${iden.category}:${iden.api}` : (iden.api || iden.category || r.name)) : r.name;
        map.set(r.id, displayName || r.id);
      }
    });
    return map;
  }, [resourcesData]);

  const createMutation = useMutation({
    mutationFn: async (values: { name: string; description?: string; permissionIds: string[] }) => {
      const { role } = await roleService.createRole({
        name: values.name,
        description: values.description,
      });

      if (role?.id && values.permissionIds.length > 0) {
        await roleService.addPermissionsToRole(role.id, {
          permissionIds: values.permissionIds,
        });
      }
      return role;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] });
      setIsCreateOpen(false);
      createForm.resetFields();
      toast.success('Role created successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const updateMutation = useMutation({
    mutationFn: async ({ id, values, currentPermissionIds }: { id: string; values: { name: string; description?: string; permissionIds: string[] }; currentPermissionIds: string[] }) => {
      await roleService.updateRole(id, {
        name: values.name,
        description: values.description,
      });

      const selectedIds = values.permissionIds || [];
      const toAdd = selectedIds.filter(id => !currentPermissionIds.includes(id));
      const toRemove = currentPermissionIds.filter(id => !selectedIds.includes(id));

      if (toAdd.length > 0) {
        await roleService.addPermissionsToRole(id, { permissionIds: toAdd });
      }
      if (toRemove.length > 0) {
        await roleService.removePermissionsFromRole(id, { permissionIds: toRemove });
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] });
      setEditingRole(null);
      editForm.resetFields();
      toast.success('Role updated successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const deleteMutation = useMutation({
    mutationFn: roleService.deleteRole,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] });
      toast.success('Role deleted successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const handleCreateSubmit = (values: any) => {
    createMutation.mutate({
      name: values.name,
      description: values.description,
      permissionIds: values.permissionIds || [],
    });
  };

  const handleEditSubmit = (values: any) => {
    if (!editingRole) return;

    updateMutation.mutate({
      id: editingRole.id,
      values: {
        name: values.name,
        description: values.description,
        permissionIds: values.permissionIds || [],
      },
      currentPermissionIds: editingRole.permissionIds || [],
    });
  };

  const permissionTreeData = React.useMemo(() => {
    if (!permissionsData?.permissions) return [];

    const grouped = permissionsData.permissions.reduce((acc, p) => {
      const resourceId = p.resourceId || 'Other';
      if (!acc[resourceId]) acc[resourceId] = [];
      acc[resourceId].push(p);
      return acc;
    }, {} as Record<string, Permission[]>);

    return Object.entries(grouped).map(([resourceId, perms]) => {
      const resourceDisplayName = resourceMap.get(resourceId) || resourceId;
      return {
        title: resourceDisplayName,
        value: `resource:${resourceId}`,
        key: `resource:${resourceId}`,
        children: perms.map(p => ({
          title: `${p.verb || p.id} (${resourceDisplayName})`,
          value: p.id,
          key: p.id,
        })),
      };
    });
  }, [permissionsData, resourceMap]);

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
      title: 'Permissions',
      key: 'permissions',
      render: (_: any, record: Role) => {
        const permissionCount = record.permissionIds?.length || 0;
        if (permissionCount === 0) return '-';

        const rolePermissions = permissionsData?.permissions?.filter(perm =>
          record.permissionIds?.includes(perm.id)
        ) || [];

        return (
          <div className="flex flex-wrap gap-1 max-w-md">
            {rolePermissions.map(p => {
              const resourceDisplayName = p.resourceId ? (resourceMap.get(p.resourceId) || p.resourceId) : '-';
              return (
                <Tag key={p.id} color="blue" className="mr-0">
                  {`${p.verb} (${resourceDisplayName})`}
                </Tag>
              );
            })}
          </div>
        );
      },
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: any, record: Role) => (
        <Space>
          <Button
            type="text"
            icon={<EditOutlined />}
            onClick={() => {
              setEditingRole(record);
              editForm.setFieldsValue({
                name: record.name,
                description: record.description,
                permissionIds: record.permissionIds || [],
              });
            }}
            className="p-1.5 rounded-md text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/20 transition-colors"
            title="Edit Role"
          />
          <Button
            type="text"
            icon={<DeleteOutlined />}
            onClick={() => {
              if (confirm('Are you sure you want to delete this role?')) {
                deleteMutation.mutate(record.id);
              }
            }}
            className="p-1.5 rounded-md text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors"
            title="Delete Role"
          />
        </Space>
      ),
    },
  ];

  // 検索フィルター
  const filteredRoles = data?.roles?.filter(role =>
    !searchText ||
    role.name.toLowerCase().includes(searchText.toLowerCase()) ||
    (role.description && role.description.toLowerCase().includes(searchText.toLowerCase()))
  );

  return (
    <div className="space-y-6">
      {/* ヘッダーセクション */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold tracking-tight text-gray-900 dark:text-white">Roles</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400">Define and manage role-based access permissions</p>
        </div>
        <div className="flex items-center gap-3">
          <Input
            placeholder="Search roles..."
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
            Add Role
          </Button>
        </div>
      </div>

      {/* ロールテーブル */}
      <Card className="shadow-sm dark:bg-gray-800 dark:border-gray-700">
        <Table
          columns={columns}
          dataSource={filteredRoles}
          loading={isLoading}
          rowKey="id"
          locale={{
            emptyText: error ? 'Failed to load roles' : 'No roles found'
          }}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => `${range[0]}-${range[1]} of ${total} roles`,
          }}
        />
      </Card>

      <Modal
        title="Create New Role"
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
            rules={[{ required: true, message: 'Please input role name!' }]}
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
            name="permissionIds"
            label="Permissions"
          >
            <TreeSelect
              treeData={permissionTreeData}
              placeholder="Select permissions"
              treeCheckable={true}
              showCheckedStrategy={TreeSelect.SHOW_CHILD}
              style={{ width: '100%' }}
              allowClear
              treeDefaultExpandAll
              showSearch
              filterTreeNode={(input, node) =>
                String(node?.title).toLowerCase().includes(input.toLowerCase())
              }
            />
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
        title="Edit Role"
        open={!!editingRole}
        onCancel={() => setEditingRole(null)}
        footer={null}
      >
        {editingRole && (
          <Form
            form={editForm}
            layout="vertical"
            onFinish={handleEditSubmit}
          >
            <Form.Item
              name="name"
              label="Name"
              rules={[{ required: true, message: 'Please input role name!' }]}
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
              name="permissionIds"
              label="Permissions"
            >
              <TreeSelect
                treeData={permissionTreeData}
                placeholder="Select permissions"
                treeCheckable={true}
                showCheckedStrategy={TreeSelect.SHOW_CHILD}
                style={{ width: '100%' }}
                allowClear
                treeDefaultExpandAll
                showSearch
                filterTreeNode={(input, node) =>
                  String(node?.title).toLowerCase().includes(input.toLowerCase())
                }
              />
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
