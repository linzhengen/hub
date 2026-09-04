import React, { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Alert, Button, Empty, Form, Input, Modal, Select, Spin, Table, Tag, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { Link } from 'react-router';
import { useNotify } from '@/hooks/useNotify';
import { useDangerConfirm } from '@/hooks/useDangerConfirm';
import PageMeta from '@/components/common/PageMeta';
import PageBreadcrumb from '@/components/common/PageBreadCrumb';
import { organizationService } from '@/services/organization';
import type { Organization, OrganizationKind } from '@/services/organization';

const ORGANIZATIONS_QUERY_KEY = ['organizations'];

const formatDate = (value?: string) => (value ? new Date(value).toLocaleString() : '—');

/**
 * KindTag is the one piece of the page that has to be read carefully.
 *
 * A PLATFORM organization's grants reach every other organization, so telling
 * it apart from a tenant at a glance is the difference between "these people
 * administer Acme" and "these people administer everybody".
 */
const KindTag: React.FC<{ kind?: OrganizationKind }> = ({ kind }) => {
  switch (kind) {
    case 'ORGANIZATION_KIND_PLATFORM':
      return <Tag color="volcano">Platform</Tag>;
    case 'ORGANIZATION_KIND_BUSINESS':
      return <Tag color="blue">Business</Tag>;
    case 'ORGANIZATION_KIND_PERSONAL':
      return <Tag color="green">Personal</Tag>;
    default:
      return <Tag>Unknown</Tag>;
  }
};

const isPlatform = (organization: Organization) =>
  organization.kind === 'ORGANIZATION_KIND_PLATFORM';

export const Organizations: React.FC = () => {
  const notify = useNotify();
  const confirmDanger = useDangerConfirm();
  const queryClient = useQueryClient();

  const [creating, setCreating] = useState(false);

  const organizations = useQuery({
    queryKey: ORGANIZATIONS_QUERY_KEY,
    queryFn: () => organizationService.list({ limit: 200 }),
  });

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ORGANIZATIONS_QUERY_KEY });

  const create = useMutation({
    mutationFn: organizationService.create,
    onSuccess: () => {
      setCreating(false);
      notify.success('Organization created');
      void invalidate();
    },
    onError: (error: Error) => notify.error(error.message),
  });

  const remove = useMutation({
    mutationFn: (id: string) => organizationService.remove(id),
    onSuccess: () => {
      notify.success('Organization deleted');
      void invalidate();
    },
    onError: (error: Error) => notify.error(error.message),
  });

  const columns: ColumnsType<Organization> = [
    {
      title: 'Name',
      key: 'name',
      render: (_, organization) => (
        <div>
          <div className="font-medium">{organization.name}</div>
          <div className="text-xs text-gray-500">{organization.description || '—'}</div>
        </div>
      ),
    },
    {
      title: 'Slug',
      key: 'slug',
      render: (_, organization) => <Typography.Text copyable>{organization.slug}</Typography.Text>,
    },
    {
      title: 'Kind',
      key: 'kind',
      render: (_, organization) => <KindTag kind={organization.kind} />,
    },
    {
      title: 'Created',
      key: 'created',
      render: (_, organization) => (
        <span className="text-xs text-gray-500">{formatDate(organization.createdAt)}</span>
      ),
    },
    {
      title: '',
      key: 'actions',
      render: (_, organization) =>
        isPlatform(organization) ? (
          // Not a disabled button: there is nothing to enable. Deleting it
          // would take every group that predates organizations with it.
          <span className="text-xs text-gray-400">Cannot be deleted</span>
        ) : (
          <Button
            size="small"
            danger
            loading={remove.isPending}
            onClick={() =>
              confirmDanger(
                {
                  title: `Delete ${organization.name}?`,
                  content:
                    'Every group in this organization is deleted with it, so everyone who held ' +
                    'access through one loses it. This cannot be undone.',
                },
                () => organization.id && remove.mutate(organization.id),
              )
            }
          >
            Delete
          </Button>
        ),
    },
  ];

  if (organizations.isLoading) {
    return (
      <div className="flex justify-center p-12">
        <Spin />
      </div>
    );
  }

  return (
    <>
      <PageMeta
        title="Organizations"
        description="The tenants of this installation: the boundary a permission is held within."
      />
      <PageBreadcrumb pageTitle="Organizations" />

      {organizations.isError ? (
        <Alert type="error" className="mb-4" message={(organizations.error as Error).message} />
      ) : null}

      <div className="mb-4 flex items-center justify-between gap-4">
        {/* Nothing is stored "inside" an organization: a group says which one
            it belongs to, and a user joins by joining one of its groups. Saying
            so here beats building a second way to manage membership. */}
        <p className="m-0 text-sm text-gray-500 dark:text-gray-400">
          A group belongs to an organization, and a user joins one by joining a group in it — on the{' '}
          <Link className="text-brand-500 hover:underline" to="/system/groups">
            Groups
          </Link>{' '}
          page.
        </p>
        <Button type="primary" onClick={() => setCreating(true)}>
          Create an organization
        </Button>
      </div>

      <Table
        rowKey="id"
        dataSource={organizations.data?.organizations ?? []}
        columns={columns}
        pagination={false}
        locale={{ emptyText: <Empty description="No organizations" /> }}
      />

      <CreateModal
        open={creating}
        submitting={create.isPending}
        onCancel={() => setCreating(false)}
        onSubmit={(values) => create.mutate(values)}
      />
    </>
  );
};

const CreateModal: React.FC<{
  open: boolean;
  submitting: boolean;
  onCancel: () => void;
  onSubmit: (values: {
    name: string;
    slug: string;
    kind: OrganizationKind;
    description: string;
  }) => void;
}> = ({ open, submitting, onCancel, onSubmit }) => {
  const [form] = Form.useForm();

  return (
    <Modal
      open={open}
      title="Create an organization"
      okText="Create"
      confirmLoading={submitting}
      onCancel={onCancel}
      onOk={() => {
        void form
          .validateFields()
          .then((values) => {
            onSubmit({
              name: values.name,
              slug: values.slug,
              kind: values.kind,
              description: values.description ?? '',
            });
            form.resetFields();
          })
          // validateFields rejects when a field is invalid. Antd has already
          // put the message under the field, so there is nothing to report -
          // but the rejection still has to be taken, or it surfaces as an
          // unhandled one every time somebody mistypes.
          .catch(() => {});
      }}
    >
      <Form
        form={form}
        layout="vertical"
        initialValues={{ kind: 'ORGANIZATION_KIND_BUSINESS' as OrganizationKind }}
      >
        <Form.Item
          name="name"
          label="Name"
          rules={[{ required: true, message: 'A name is required' }]}
        >
          <Input placeholder="Acme Inc." />
        </Form.Item>
        <Form.Item
          name="slug"
          label="Slug"
          extra="Lowercase letters, digits and hyphens. It travels in URLs, so it cannot be changed casually."
          rules={[
            { required: true, message: 'A slug is required' },
            {
              pattern: /^[a-z0-9][a-z0-9-]{1,62}[a-z0-9]$/,
              message: '3-64 characters of lowercase letters, digits and hyphens',
            },
          ]}
        >
          <Input placeholder="acme" />
        </Form.Item>
        <Form.Item
          name="kind"
          label="Kind"
          extra="Personal is a one-person organization. There is one platform organization and it already exists."
          rules={[{ required: true }]}
        >
          <Select
            options={[
              { value: 'ORGANIZATION_KIND_BUSINESS', label: 'Business' },
              { value: 'ORGANIZATION_KIND_PERSONAL', label: 'Personal' },
            ]}
          />
        </Form.Item>
        <Form.Item name="description" label="Description">
          <Input.TextArea rows={3} />
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default Organizations;
