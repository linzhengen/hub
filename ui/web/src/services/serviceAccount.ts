import { serviceAccountServiceOperations as ops } from '@/api/operations';
import { apiRequest } from '@/api/request';
import type { paths, components } from '@/api/schema/system-serviceaccount-v1-service';
import type { RequestParameters } from '@/api/helper';

export type ServiceAccount = components['schemas']['v1ServiceAccount'];
export type Credentials = components['schemas']['v1Credentials'];
export type CreateServiceAccountRequest = components['schemas']['v1CreateServiceAccountRequest'];
export type CreateServiceAccountResponse = components['schemas']['v1CreateServiceAccountResponse'];
export type ListServiceAccountsResponse = components['schemas']['v1ListServiceAccountsResponse'];
export type RotateServiceAccountSecretResponse =
  components['schemas']['v1RotateServiceAccountSecretResponse'];
export type DeleteServiceAccountResponse = components['schemas']['v1DeleteServiceAccountResponse'];

export type ListServiceAccountsParams = RequestParameters<paths, '/api/v1/service-accounts', 'get'>;

/**
 * The secret comes back from `create` and `rotate`, and from nowhere else.
 *
 * hub does not store it and Keycloak cannot always show it again, so there is
 * no rpc that reads one back - and this module must not grow one. Losing a
 * secret means rotating it, which is the right cost for a credential.
 */
export const serviceAccountService = {
  list: (params?: ListServiceAccountsParams) =>
    apiRequest<ListServiceAccountsResponse>(ops.listServiceAccounts, { query: params }),

  create: (data: CreateServiceAccountRequest) =>
    apiRequest<CreateServiceAccountResponse>(ops.createServiceAccount, { body: data }),

  rotateSecret: (id: string) =>
    apiRequest<RotateServiceAccountSecretResponse>(ops.rotateServiceAccountSecret, {
      path: { id },
      body: {},
    }),

  remove: (id: string) =>
    apiRequest<DeleteServiceAccountResponse>(ops.deleteServiceAccount, { path: { id } }),
};
