import { accessRequestServiceOperations as ops } from '@/api/operations';
import { apiRequest } from '@/api/request';
import type { paths, components } from '@/api/schema/system-access-v1-service';
import type { RequestParameters } from '@/api/helper';

export type AccessRequest = components['schemas']['v1AccessRequest'];
export type RequestStatus = components['schemas']['v1RequestStatus'];
export type RequestOrigin = components['schemas']['v1RequestOrigin'];
export type CreateAccessRequestRequest = components['schemas']['v1CreateAccessRequestRequest'];
export type CreateAccessRequestResponse = components['schemas']['v1CreateAccessRequestResponse'];
export type ListAccessRequestsResponse = components['schemas']['v1ListAccessRequestsResponse'];
export type CancelAccessRequestResponse = components['schemas']['v1CancelAccessRequestResponse'];
export type DecideAccessRequestResponse = components['schemas']['v1DecideAccessRequestResponse'];

export type ListAccessRequestsParams = RequestParameters<paths, '/api/v1/access-requests', 'get'>;

export const accessRequestService = {
  list: (params?: ListAccessRequestsParams) =>
    apiRequest<ListAccessRequestsResponse>(ops.listAccessRequests, { query: params }),

  create: (data: CreateAccessRequestRequest) =>
    apiRequest<CreateAccessRequestResponse>(ops.createAccessRequest, { body: data }),

  cancel: (id: string) =>
    apiRequest<CancelAccessRequestResponse>(ops.cancelAccessRequest, { path: { id }, body: {} }),

  /**
   * Approves or rejects a request. Approving performs the grant, for the term
   * that was asked for.
   *
   * The server refuses a decision from the person who raised the request. The
   * UI hides the buttons in that case, but the refusal is the server's: a
   * hidden button is a courtesy, not a control.
   */
  decide: (id: string, approved: boolean, comment: string) =>
    apiRequest<DecideAccessRequestResponse>(ops.decideAccessRequest, {
      path: { id },
      body: { approved, comment },
    }),
};
