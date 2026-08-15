import React, { useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Alert,
  Badge,
  Button,
  Descriptions,
  Empty,
  Form,
  Input,
  Modal,
  Select,
  Spin,
  Table,
  Tabs,
  Tag,
  Tooltip,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { RobotOutlined } from '@ant-design/icons';
import { useNotify } from '@/hooks/useNotify';
import { useMe } from '@/hooks/useMe';
import PageMeta from '@/components/common/PageMeta';
import PageBreadcrumb from '@/components/common/PageBreadCrumb';
import { accessRequestService } from '@/services/access';
import {
  ACCESS_REQUESTS_QUERY_KEY,
  pendingForApprover,
  useAccessRequests,
} from '@/hooks/useAccessRequests';
import type { AccessRequest, RequestStatus } from '@/services/access';
import { groupService } from '@/services/group';
import { userService } from '@/services/user';

const STATUS_TAG: Record<string, { color: string; label: string }> = {
  REQUEST_STATUS_PENDING: { color: 'processing', label: 'Pending' },
  REQUEST_STATUS_APPROVED: { color: 'success', label: 'Approved' },
  REQUEST_STATUS_REJECTED: { color: 'error', label: 'Rejected' },
  REQUEST_STATUS_CANCELLED: { color: 'default', label: 'Cancelled' },
};

const StatusTag: React.FC<{ status?: RequestStatus }> = ({ status }) => {
  const tag = STATUS_TAG[status ?? ''] ?? { color: 'default', label: 'Unknown' };
  return <Tag color={tag.color}>{tag.label}</Tag>;
};

/**
 * Marks a request the assistant raised.
 *
 * This is the one piece of provenance an approver has to see. The assistant
 * cannot exceed anyone's permissions and cannot decide anything, but it
 * composes a request out of text other people wrote - a group description, a
 * message in a conversation - so "a colleague typed this" and "a model was
 * talked into typing this" are different things to be asked to approve, and the
 * screen has to say which one this is.
 */
const OriginTag: React.FC<{ origin?: string }> = ({ origin }) => {
  if (origin !== 'REQUEST_ORIGIN_AI_CHAT') return null;
  return (
    <Tooltip title="Raised by the assistant while answering a conversation. Read what it asks for before approving.">
      <Tag icon={<RobotOutlined />} color="warning">
        AI chat
      </Tag>
    </Tooltip>
  );
};

const formatDate = (value?: string) => (value ? new Date(value).toLocaleString() : '—');

/** "Permanently" is a bigger ask than a term, and reads as one here. */
const formatTerm = (value?: string) =>
  value ? `until ${new Date(value).toLocaleString()}` : 'permanently';

export const AccessRequests: React.FC = () => {
  const notify = useNotify();
  const queryClient = useQueryClient();
  const { data: me } = useMe();
  const myUserId = me?.user?.id;

  const [composing, setComposing] = useState(false);
  const [deciding, setDeciding] = useState<AccessRequest | null>(null);

  const requests = useAccessRequests();

  const groups = useQuery({
    queryKey: ['groups', 'forAccessRequests'],
    queryFn: () => groupService.listGroups({ limit: 200 }),
  });

  const users = useQuery({
    queryKey: ['users', 'forAccessRequests'],
    queryFn: () => userService.listUsers({ limit: 200 }),
  });

  const groupName = useMemo(() => {
    const byId = new Map((groups.data?.groups ?? []).map((g) => [g.id, g.name]));
    // Falling back to the id rather than to "unknown": an id a reader can look
    // up beats a word that tells them nothing.
    return (id?: string) => (id ? (byId.get(id) ?? id) : '—');
  }, [groups.data]);

  const userName = useMemo(() => {
    const byId = new Map((users.data?.users ?? []).map((u) => [u.id, u.username]));
    return (id?: string) => (id ? (byId.get(id) ?? id) : '—');
  }, [users.data]);

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ACCESS_REQUESTS_QUERY_KEY });

  const create = useMutation({
    mutationFn: accessRequestService.create,
    onSuccess: () => {
      notify.success('Request raised');
      setComposing(false);
      void invalidate();
    },
    onError: (error: Error) => notify.error(error.message),
  });

  const cancel = useMutation({
    mutationFn: (id: string) => accessRequestService.cancel(id),
    onSuccess: () => {
      notify.success('Request withdrawn');
      void invalidate();
    },
    onError: (error: Error) => notify.error(error.message),
  });

  const decide = useMutation({
    mutationFn: ({ id, approved, comment }: { id: string; approved: boolean; comment: string }) =>
      accessRequestService.decide(id, approved, comment),
    onSuccess: (_, variables) => {
      notify.success(variables.approved ? 'Approved, and the access granted' : 'Rejected');
      setDeciding(null);
      void invalidate();
    },
    onError: (error: Error) => notify.error(error.message),
  });

  const all = requests.data?.accessRequests ?? [];
  const mine = all.filter((r) => r.requesterUserId === myUserId);
  // The queue excludes your own asks, because you may not decide them. Showing
  // them here with the buttons disabled would only invite the click. The rule
  // is shared with the header's badge, so the two cannot disagree about what is
  // waiting on you.
  const queue = pendingForApprover(all, myUserId);

  const sharedColumns: ColumnsType<AccessRequest> = [
    {
      title: 'Access',
      key: 'access',
      render: (_, r) => (
        <div>
          <div>
            <strong>{userName(r.subjectUserId)}</strong> in <strong>{groupName(r.groupId)}</strong>
          </div>
          <div className="text-xs text-gray-500">{formatTerm(r.requestedUntil)}</div>
        </div>
      ),
    },
    { title: 'Reason', dataIndex: 'reason', key: 'reason' },
    {
      title: 'Raised',
      key: 'raised',
      render: (_, r) => (
        <div>
          <div>
            {userName(r.requesterUserId)} <OriginTag origin={r.origin} />
          </div>
          <div className="text-xs text-gray-500">{formatDate(r.createdAt)}</div>
        </div>
      ),
    },
  ];

  const myColumns: ColumnsType<AccessRequest> = [
    ...sharedColumns,
    {
      title: 'Status',
      key: 'status',
      render: (_, r) => (
        <div>
          <StatusTag status={r.status} />
          {r.decidedByUserId ? (
            <div className="text-xs text-gray-500">
              by {userName(r.decidedByUserId)} · {formatDate(r.decidedAt)}
            </div>
          ) : null}
          {r.decisionComment ? (
            <div className="text-xs text-gray-500">“{r.decisionComment}”</div>
          ) : null}
        </div>
      ),
    },
    {
      title: '',
      key: 'actions',
      render: (_, r) =>
        r.status === 'REQUEST_STATUS_PENDING' ? (
          <Button size="small" loading={cancel.isPending} onClick={() => cancel.mutate(r.id!)}>
            Withdraw
          </Button>
        ) : null,
    },
  ];

  const queueColumns: ColumnsType<AccessRequest> = [
    ...sharedColumns,
    {
      title: '',
      key: 'actions',
      render: (_, r) => (
        <Button type="primary" size="small" onClick={() => setDeciding(r)}>
          Review
        </Button>
      ),
    },
  ];

  if (requests.isLoading) {
    return (
      <div className="flex justify-center p-12">
        <Spin />
      </div>
    );
  }

  return (
    <>
      <PageMeta title="Access requests" description="Ask to be put in a group, and decide requests." />
      <PageBreadcrumb pageTitle="Access requests" />

      {requests.isError ? (
        <Alert type="error" className="mb-4" message={(requests.error as Error).message} />
      ) : null}

      <div className="mb-4 flex justify-end">
        <Button type="primary" onClick={() => setComposing(true)}>
          Ask for access
        </Button>
      </div>

      <Tabs
        items={[
          {
            key: 'queue',
            label: <Badge count={queue.length} offset={[12, 0]}>To decide</Badge>,
            children: (
              <Table
                rowKey="id"
                dataSource={queue}
                columns={queueColumns}
                pagination={false}
                locale={{ emptyText: <Empty description="Nothing waiting on you" /> }}
              />
            ),
          },
          {
            key: 'mine',
            label: 'My requests',
            children: (
              <Table
                rowKey="id"
                dataSource={mine}
                columns={myColumns}
                pagination={false}
                locale={{ emptyText: <Empty description="You have not asked for anything" /> }}
              />
            ),
          },
        ]}
      />

      <ComposeModal
        open={composing}
        submitting={create.isPending}
        groups={(groups.data?.groups ?? []).map((g) => ({ value: g.id!, label: g.name! }))}
        users={(users.data?.users ?? []).map((u) => ({ value: u.id!, label: u.username! }))}
        onCancel={() => setComposing(false)}
        onSubmit={(values) => create.mutate(values)}
      />

      <DecideModal
        request={deciding}
        submitting={decide.isPending}
        groupName={groupName}
        userName={userName}
        onCancel={() => setDeciding(null)}
        onDecide={(approved, comment) =>
          deciding?.id && decide.mutate({ id: deciding.id, approved, comment })
        }
      />
    </>
  );
};

interface ComposeValues {
  subjectUserId?: string;
  groupId: string;
  reason: string;
  requestedUntil?: string;
}

const ComposeModal: React.FC<{
  open: boolean;
  submitting: boolean;
  groups: { value: string; label: string }[];
  users: { value: string; label: string }[];
  onCancel: () => void;
  onSubmit: (values: ComposeValues) => void;
}> = ({ open, submitting, groups, users, onCancel, onSubmit }) => {
  const [form] = Form.useForm();

  return (
    <Modal
      open={open}
      title="Ask for access"
      okText="Raise request"
      confirmLoading={submitting}
      onCancel={onCancel}
      onOk={() => {
        void form.validateFields().then((values) => {
          // The input yields local wall-clock text ("2026-08-20T17:00"); the
          // API takes an instant. Parsing it in the browser resolves it in the
          // reader's own zone, which is the one they typed it in.
          const until = values.requestedUntil as string | undefined;
          onSubmit({
            subjectUserId: values.subjectUserId || undefined,
            groupId: values.groupId,
            reason: values.reason,
            requestedUntil: until ? new Date(until).toISOString() : undefined,
          });
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
          name="subjectUserId"
          label="For"
          extra="Leave empty to ask for yourself."
        >
          <Select allowClear showSearch optionFilterProp="label" options={users} placeholder="Yourself" />
        </Form.Item>
        <Form.Item
          name="groupId"
          label="Group"
          rules={[{ required: true, message: 'Choose the group being asked for' }]}
        >
          <Select showSearch optionFilterProp="label" options={groups} />
        </Form.Item>
        <Form.Item
          name="reason"
          label="Reason"
          rules={[{ required: true, message: 'A request nobody explained is a request nobody can judge' }]}
        >
          <Input.TextArea rows={3} maxLength={1024} />
        </Form.Item>
        <Form.Item
          name="requestedUntil"
          label="Until"
          extra="Leave empty to ask for it permanently, which is a bigger ask."
        >
          <Input type="datetime-local" />
        </Form.Item>
      </Form>
    </Modal>
  );
};

/**
 * The approval screen.
 *
 * It shows the real operation and its real arguments - this user, this group,
 * this long - rather than a summary of them, for the same reason the
 * assistant's tool proposals do: "grant some access" is not something anybody
 * can meaningfully agree to.
 */
const DecideModal: React.FC<{
  request: AccessRequest | null;
  submitting: boolean;
  groupName: (id?: string) => string;
  userName: (id?: string) => string;
  onCancel: () => void;
  onDecide: (approved: boolean, comment: string) => void;
}> = ({ request, submitting, groupName, userName, onCancel, onDecide }) => {
  const [comment, setComment] = useState('');

  if (!request) return null;

  return (
    <Modal
      open
      title="Review request"
      onCancel={onCancel}
      afterClose={() => setComment('')}
      footer={[
        <Button key="cancel" onClick={onCancel}>
          Close
        </Button>,
        <Button key="reject" danger loading={submitting} onClick={() => onDecide(false, comment)}>
          Reject
        </Button>,
        <Button
          key="approve"
          type="primary"
          loading={submitting}
          onClick={() => onDecide(true, comment)}
        >
          Approve and grant
        </Button>,
      ]}
    >
      {request.origin === 'REQUEST_ORIGIN_AI_CHAT' ? (
        <Alert
          type="warning"
          showIcon
          className="mb-4"
          message="Raised by the assistant"
          description="This was composed while answering a conversation. Read what it asks for, not what it says about itself."
        />
      ) : null}

      <Descriptions column={1} size="small" bordered>
        <Descriptions.Item label="Put">{userName(request.subjectUserId)}</Descriptions.Item>
        <Descriptions.Item label="In group">{groupName(request.groupId)}</Descriptions.Item>
        <Descriptions.Item label="For">{formatTerm(request.requestedUntil)}</Descriptions.Item>
        <Descriptions.Item label="Asked by">{userName(request.requesterUserId)}</Descriptions.Item>
        <Descriptions.Item label="Reason">{request.reason}</Descriptions.Item>
      </Descriptions>

      <Input.TextArea
        className="mt-4"
        rows={2}
        maxLength={1024}
        placeholder="Comment (optional)"
        value={comment}
        onChange={(e) => setComment(e.target.value)}
      />
    </Modal>
  );
};

export default AccessRequests;
