import { auditServiceOperations as ops } from '@/api/operations';
import { apiRequest } from '@/api/request';
import type { paths, components } from '@/api/schema/system-audit-v1-service';
import type { RequestParameters } from '@/api/helper';

export type AuditLog = components['schemas']['v1AuditLog'];
export type AuditChannel = components['schemas']['v1Channel'];
export type ListAuditLogResponse = components['schemas']['v1ListAuditLogResponse'];

export type ListAuditLogParams = RequestParameters<paths, '/api/v1/audit-logs', 'get'>;

/**
 * The audit log is read-only by design: records are written by the server as
 * changes are made, and one that could be edited afterwards would prove
 * nothing. There is no write rpc to wrap.
 */
export const auditService = {
  list: (params?: ListAuditLogParams) =>
    apiRequest<ListAuditLogResponse>(ops.listAuditLog, { query: params }),
};
