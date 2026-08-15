import React, { useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Alert, Button, Empty, Form, Input, Modal, Spin, Table, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { Link } from 'react-router';
import { useNotify } from '@/hooks/useNotify';
import { useDangerConfirm } from '@/hooks/useDangerConfirm';
import PageMeta from '@/components/common/PageMeta';
import PageBreadcrumb from '@/components/common/PageBreadCrumb';
import { serviceAccountService } from '@/services/serviceAccount';
import type { Credentials, ServiceAccount } from '@/services/serviceAccount';
import { userService } from '@/services/user';

const SERVICE_ACCOUNTS_QUERY_KEY = ['serviceAccounts'];

const formatDate = (value?: string) => (value ? new Date(value).toLocaleString() : '—');

export const ServiceAccounts: React.FC = () => {
  const notify = useNotify();
  const confirmDanger = useDangerConfirm();
  const queryClient = useQueryClient();

  const [registering, setRegistering] = useState(false);
  // issued holds the credentials just handed out. It is the only place a secret
  // exists in the browser, and it is cleared when the dialog closes.
  const [issued, setIssued] = useState<Credentials | null>(null);

  const accounts = useQuery({
    queryKey: SERVICE_ACCOUNTS_QUERY_KEY,
    queryFn: () => serviceAccountService.list({ limit: 200 }),
  });

  const users = useQuery({
    queryKey: ['users', 'forServiceAccounts'],
    queryFn: () => userService.listUsers({ limit: 200 }),
  });

  const userName = useMemo(() => {
    const byId = new Map((users.data?.users ?? []).map((u) => [u.id, u.username]));
    return (id?: string) => (id ? (byId.get(id) ?? id) : '—');
  }, [users.data]);

  const invalidate = () =>
    queryClient.invalidateQueries({ queryKey: SERVICE_ACCOUNTS_QUERY_KEY });

  const register = useMutation({
    mutationFn: serviceAccountService.create,
    onSuccess: (response) => {
      setRegistering(false);
      setIssued(response.credentials ?? null);
      void invalidate();
    },
    onError: (error: Error) => notify.error(error.message),
  });

  const rotate = useMutation({
    mutationFn: (id: string) => serviceAccountService.rotateSecret(id),
    onSuccess: (response) => setIssued(response.credentials ?? null),
    onError: (error: Error) => notify.error(error.message),
  });

  const remove = useMutation({
    mutationFn: (id: string) => serviceAccountService.remove(id),
    onSuccess: () => {
      notify.success('Service account removed');
      void invalidate();
    },
    onError: (error: Error) => notify.error(error.message),
  });

  const columns: ColumnsType<ServiceAccount> = [
    {
      title: 'Name',
      key: 'name',
      render: (_, account) => (
        <div>
          <div className="font-medium">{account.name}</div>
          <div className="text-xs text-gray-500">{account.description || '—'}</div>
        </div>
      ),
    },
    {
      title: 'Client ID',
      key: 'clientId',
      // Derived from the name, so it is shown rather than edited: an operator
      // copies it into HUB_OIDC_CLIENT_ID and never types it themselves.
      render: (_, account) => <Typography.Text copyable>{account.clientId}</Typography.Text>,
    },
    {
      title: 'Registered',
      key: 'registered',
      render: (_, account) => (
        <div>
          <div>{userName(account.createdByUserId)}</div>
          <div className="text-xs text-gray-500">{formatDate(account.createdAt)}</div>
        </div>
      ),
    },
    {
      title: '',
      key: 'actions',
      render: (_, account) => (
        <div className="flex gap-2">
          <Button
            size="small"
            loading={rotate.isPending}
            onClick={() =>
              confirmDanger(
                {
                  title: `Rotate the secret for ${account.name}?`,
                  content:
                    'The current secret stops working immediately. Anything still using it will ' +
                    'fail to authenticate until it is given the new one.',
                  okText: 'Rotate',
                },
                () => account.id && rotate.mutate(account.id),
              )
            }
          >
            Rotate secret
          </Button>
          <Button
            size="small"
            danger
            loading={remove.isPending}
            onClick={() =>
              confirmDanger(
                {
                  title: `Remove ${account.name}?`,
                  content:
                    'This deletes the Keycloak client, so the machine can no longer authenticate ' +
                    'at all, and drops the group memberships it held. It cannot be undone.',
                },
                () => account.id && remove.mutate(account.id),
              )
            }
          >
            Remove
          </Button>
        </div>
      ),
    },
  ];

  if (accounts.isLoading) {
    return (
      <div className="flex justify-center p-12">
        <Spin />
      </div>
    );
  }

  return (
    <>
      <PageMeta
        title="Service accounts"
        description="The machines that call the API: CI, jobs, other services."
      />
      <PageBreadcrumb pageTitle="Service accounts" />

      {accounts.isError ? (
        <Alert type="error" className="mb-4" message={(accounts.error as Error).message} />
      ) : null}

      <div className="mb-4 flex items-center justify-between gap-4">
        {/* A service account is a row in `users`, so its groups are managed
            where everybody else's are. Saying so once beats a link on every
            row, and beats building a second way to do the same thing. */}
        <p className="m-0 text-sm text-gray-500 dark:text-gray-400">
          A service account is a user, so its groups and roles are managed on the{' '}
          <Link className="text-brand-500 hover:underline" to="/users">
            Users
          </Link>{' '}
          page.
        </p>
        <Button type="primary" onClick={() => setRegistering(true)}>
          Register a machine
        </Button>
      </div>

      <Table
        rowKey="id"
        dataSource={accounts.data?.serviceAccounts ?? []}
        columns={columns}
        pagination={false}
        locale={{ emptyText: <Empty description="No machines registered" /> }}
      />

      <RegisterModal
        open={registering}
        submitting={register.isPending}
        onCancel={() => setRegistering(false)}
        onSubmit={(values) => register.mutate(values)}
      />

      <CredentialsModal credentials={issued} onClose={() => setIssued(null)} />
    </>
  );
};

const RegisterModal: React.FC<{
  open: boolean;
  submitting: boolean;
  onCancel: () => void;
  onSubmit: (values: { name: string; description: string }) => void;
}> = ({ open, submitting, onCancel, onSubmit }) => {
  const [form] = Form.useForm();

  return (
    <Modal
      open={open}
      title="Register a machine"
      okText="Register"
      confirmLoading={submitting}
      onCancel={onCancel}
      onOk={() => {
        void form.validateFields().then((values) => {
          onSubmit({ name: values.name, description: values.description ?? '' });
          form.resetFields();
        })
        // validateFields rejects when a field is invalid. Antd has already put
        // the message under the field, so there is nothing to report here - but
        // the rejection still has to be taken, or it surfaces as an unhandled
        // one every time somebody mistypes.
        .catch(() => {});
      }}
    >
      <Form form={form} layout="vertical">
        <Form.Item
          name="name"
          label="Name"
          extra="Lowercase letters, digits and hyphens. The client id is derived from it and cannot be changed afterwards."
          rules={[
            { required: true, message: 'Give the machine a name' },
            {
              pattern: /^[a-z0-9][a-z0-9-]{1,62}[a-z0-9]$/,
              message: 'Lowercase letters, digits and hyphens; 3-64 characters',
            },
          ]}
        >
          <Input placeholder="ci-deploy" />
        </Form.Item>
        <Form.Item
          name="description"
          label="What it is for"
          extra="A machine nobody can describe is a machine nobody dares turn off."
        >
          <Input.TextArea rows={3} maxLength={1024} />
        </Form.Item>
      </Form>
    </Modal>
  );
};

/**
 * The one and only sight of a secret.
 *
 * hub does not store it and Keycloak cannot always show it again, so there is
 * no screen that reads one back and this dialog must not pretend otherwise: it
 * says so plainly, and closing it loses the value for good. Offering a "show it
 * again" link would be a lie the first person to trust it discovers the hard
 * way.
 */
const CredentialsModal: React.FC<{
  credentials: Credentials | null;
  onClose: () => void;
}> = ({ credentials, onClose }) => {
  if (!credentials) return null;

  return (
    <Modal
      open
      title="Credentials"
      onCancel={onClose}
      footer={[
        <Button key="done" type="primary" onClick={onClose}>
          I have copied the secret
        </Button>,
      ]}
    >
      <Alert
        type="warning"
        showIcon
        className="mb-4"
        message="This secret is shown once"
        description="It is not stored here and cannot be shown again. If it is lost, rotate it."
      />

      <div className="mb-3">
        <div className="mb-1 text-xs text-gray-500">HUB_OIDC_CLIENT_ID</div>
        <Typography.Text copyable code>
          {credentials.clientId}
        </Typography.Text>
      </div>

      <div>
        <div className="mb-1 text-xs text-gray-500">HUB_OIDC_CLIENT_SECRET</div>
        <Typography.Text copyable code>
          {credentials.secret}
        </Typography.Text>
      </div>
    </Modal>
  );
};

export default ServiceAccounts;
