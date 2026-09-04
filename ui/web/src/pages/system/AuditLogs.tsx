import React, { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Alert, Button, Empty, Input, Select, Space, Spin, Table, Tag, Tooltip } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { RobotOutlined } from '@ant-design/icons';
import PageMeta from '@/components/common/PageMeta';
import PageBreadcrumb from '@/components/common/PageBreadCrumb';
import { auditService } from '@/services/audit';
import type { AuditChannel, AuditLog, ListAuditLogParams } from '@/services/audit';

import { userService } from '@/services/user';
import { organizationService } from '@/services/organization';

// The generated query type is optional as a whole, since every filter is. The
// page always holds one, so it is narrowed here rather than guarded at each of
// the dozen places that read it.
type Filters = NonNullable<ListAuditLogParams>;

const PAGE_SIZE = 50;

const formatDate = (value?: string) => (value ? new Date(value).toLocaleString() : '—');

/**
 * Marks a change the assistant made.
 *
 * The assistant is a channel, never an actor: the record names the person it
 * was answering. But "they did this at the console" and "they approved this in
 * a conversation" are different accounts of the same change, and an
 * investigation needs to know which - so this is the one column that must not
 * be left to be inferred.
 */
const ChannelTag: React.FC<{ channel?: string; sessionId?: string }> = ({ channel, sessionId }) => {
  if (channel !== 'CHANNEL_AI_CHAT') {
    return <Tag className="m-0">Direct</Tag>;
  }
  return (
    <Tooltip title={sessionId ? `Chat session ${sessionId}` : 'Made through the assistant'}>
      <Tag icon={<RobotOutlined />} color="warning" className="m-0">
        AI chat
      </Tag>
    </Tooltip>
  );
};

/**
 * Marks whether the attempt got through.
 *
 * Refused attempts are the events this log exists for - somebody trying to
 * grant themselves admin and being told no is exactly what an investigation is
 * looking for - so a failure is coloured rather than filtered out.
 */
const OutcomeTag: React.FC<{ succeeded?: boolean; error?: string }> = ({ succeeded, error }) => {
  if (succeeded) return <Tag color="success" className="m-0">Allowed</Tag>;
  return (
    <Tooltip title={error || 'The attempt did not go through'}>
      <Tag color="error" className="m-0">Refused</Tag>
    </Tooltip>
  );
};

/** Pretty-prints the recorded request, which the server has already stripped of secrets. */
const formatArguments = (raw?: string) => {
  if (!raw) return '';
  try {
    return JSON.stringify(JSON.parse(raw), null, 2);
  } catch {
    // Not JSON after all - showing it as it was recorded beats showing nothing.
    return raw;
  }
};

export const AuditLogs: React.FC = () => {
  const [filters, setFilters] = useState<Filters>({ limit: PAGE_SIZE, offset: 0 });

  const logs = useQuery({
    queryKey: ['auditLogs', filters],
    queryFn: () => auditService.list(filters),
    // Keeping the previous page while the next one loads. Without it every
    // keystroke in a filter replaces the whole screen - filters included - with
    // a spinner, and the reader loses what they were half way through typing.
    placeholderData: (previous) => previous,
  });

  const users = useQuery({
    queryKey: ['users', 'forAuditLogs'],
    queryFn: () => userService.listUsers({ limit: 200 }),
  });

  const organizations = useQuery({
    queryKey: ['organizations', 'forAuditLogs'],
    queryFn: () => organizationService.list({ limit: 200 }),
  });

  const userName = useMemo(() => {
    const byId = new Map((users.data?.users ?? []).map((u) => [u.id, u.username]));
    // Falling back to the id rather than to "unknown": an id a reader can look
    // up beats a word that tells them nothing. It also matters here because a
    // deleted user's changes stay in the log after the row is gone.
    return (id?: string) => (id ? (byId.get(id) ?? id) : '—');
  }, [users.data]);

  const narrow = (next: Partial<Filters>) =>
    setFilters((current) => ({ ...current, ...next, offset: 0 }));

  const organizationName = useMemo(() => {
    const byId = new Map((organizations.data?.organizations ?? []).map((o) => [o.id, o.name]));
    return (id?: string) => (id ? (byId.get(id) ?? id) : '—');
  }, [organizations.data]);

  const columns: ColumnsType<AuditLog> = [
    {
      title: 'When',
      key: 'when',
      width: 180,
      render: (_, log) => <span className="text-sm">{formatDate(log.createdAt)}</span>,
    },
    {
      title: 'Who',
      key: 'who',
      render: (_, log) => (
        <div>
          <div>{userName(log.actorUserId)}</div>
          <div className="mt-1">
            <ChannelTag channel={log.channel} sessionId={log.sessionId} />
          </div>
        </div>
      ),
    },
    {
      // The listing is already narrowed to what the reader may see; this is for
      // somebody who may see several tenants and has to tell them apart.
      title: 'Where',
      key: 'where',
      render: (_, log) => (
        <span className="text-sm">{organizationName(log.orgId)}</span>
      ),
    },
    {
      title: 'What',
      key: 'what',
      render: (_, log) => (
        <div>
          <div className="font-medium">{log.action}</div>
          <div className="text-xs text-gray-500 break-all">{log.resource}</div>
          {log.targetId ? (
            <div className="text-xs text-gray-500 break-all">on {log.targetId}</div>
          ) : null}
        </div>
      ),
    },
    {
      title: 'Outcome',
      key: 'outcome',
      width: 110,
      render: (_, log) => <OutcomeTag succeeded={log.succeeded} error={log.error} />,
    },
  ];

  if (logs.isLoading) {
    return (
      <div className="flex justify-center p-12">
        <Spin />
      </div>
    );
  }

  const rows = logs.data?.auditLogs ?? [];
  const total = Number(logs.data?.total ?? 0);
  const offset = filters.offset ?? 0;

  return (
    <>
      <PageMeta
        title="Audit log"
        description="Every recorded attempt to change who can do what."
      />
      <PageBreadcrumb pageTitle="Audit log" />

      {logs.isError ? (
        <Alert type="error" className="mb-4" message={(logs.error as Error).message} />
      ) : null}

      <Space wrap className="mb-4">
        <Select
          allowClear
          aria-label="Actor"
          placeholder="Anyone"
          style={{ minWidth: 200 }}
          showSearch
          optionFilterProp="label"
          options={(users.data?.users ?? []).map((u) => ({
            value: u.id ?? '',
            label: u.username ?? '',
          }))}
          onChange={(actorUserId?: string) => narrow({ actorUserId })}
        />
        <Select
          allowClear
          aria-label="Route"
          placeholder="Any route"
          style={{ minWidth: 160 }}
          options={[
            { value: 'CHANNEL_API', label: 'Direct' },
            { value: 'CHANNEL_AI_CHAT', label: 'AI chat' },
          ]}
          onChange={(channel?: AuditChannel) => narrow({ channel })}
        />
        <Input
          allowClear
          aria-label="Action"
          placeholder="Action, e.g. AddRolesToGroup"
          style={{ width: 260 }}
          onChange={(e) => narrow({ action: e.target.value || undefined })}
        />
        {/* The input yields local wall-clock text; the API takes an instant, so
            it is resolved in the reader's own zone - the one they typed it in. */}
        <Input
          type="datetime-local"
          aria-label="From"
          style={{ width: 220 }}
          onChange={(e) =>
            narrow({ since: e.target.value ? new Date(e.target.value).toISOString() : undefined })
          }
        />
        <Input
          type="datetime-local"
          aria-label="Until"
          style={{ width: 220 }}
          onChange={(e) =>
            narrow({ until: e.target.value ? new Date(e.target.value).toISOString() : undefined })
          }
        />
      </Space>

      <Table
        rowKey="id"
        dataSource={rows}
        columns={columns}
        pagination={false}
        locale={{ emptyText: <Empty description="Nothing recorded for this filter" /> }}
        expandable={{
          // The recorded request, not a summary of it. "Somebody changed a
          // group" is not something an investigation can conclude anything
          // from; the arguments are what it is reading the log for. Secrets are
          // already stripped server-side.
          expandedRowRender: (log) => (
            <pre className="m-0 overflow-x-auto text-xs whitespace-pre-wrap">
              {formatArguments(log.arguments) || 'No arguments recorded'}
            </pre>
          ),
          rowExpandable: (log) => Boolean(log.arguments),
        }}
      />

      <div className="mt-4 flex items-center justify-between">
        <span className="text-sm text-gray-500">
          {total > 0 ? `${offset + 1}–${offset + rows.length} of ${total}` : 'No records'}
        </span>
        <Space>
          <Button
            disabled={offset === 0}
            onClick={() =>
              setFilters((f) => ({ ...f, offset: Math.max(0, (f.offset ?? 0) - PAGE_SIZE) }))
            }
          >
            Newer
          </Button>
          <Button
            disabled={offset + rows.length >= total}
            onClick={() => setFilters((f) => ({ ...f, offset: (f.offset ?? 0) + PAGE_SIZE }))}
          >
            Older
          </Button>
        </Space>
      </div>
    </>
  );
};

export default AuditLogs;
