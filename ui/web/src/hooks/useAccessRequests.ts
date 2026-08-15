import { useQuery } from '@tanstack/react-query';
import { useMe } from '@/hooks/useMe';
import { accessRequestService } from '@/services/access';
import type { AccessRequest } from '@/services/access';

export const ACCESS_REQUESTS_QUERY_KEY = ['accessRequests'];

/**
 * How often the header re-reads the queue.
 *
 * Approvals keep somebody else waiting, so noticing one late has a cost - but a
 * minute of it is not the cost. Polling harder would spend requests on a
 * question whose answer changes a few times a day.
 */
const REFETCH_INTERVAL_MS = 60_000;

/**
 * pendingForApprover returns the requests userId is able to decide.
 *
 * Your own requests are not among them: the server refuses a decision from the
 * person who raised it, so counting them would promise an action that leads to
 * a refusal.
 *
 * The rule lives here rather than in each screen because it is used twice - the
 * queue on the page and the count in the header - and two copies would let the
 * badge and the list disagree about what is waiting on you.
 */
export const pendingForApprover = (
  requests: AccessRequest[],
  userId: string | undefined,
): AccessRequest[] =>
  requests.filter(
    (r) => r.status === 'REQUEST_STATUS_PENDING' && r.requesterUserId !== userId,
  );

/** useAccessRequests reads every request the caller may see. */
export const useAccessRequests = (options?: { poll?: boolean }) =>
  useQuery({
    queryKey: ACCESS_REQUESTS_QUERY_KEY,
    queryFn: () => accessRequestService.list({ limit: 200 }),
    refetchInterval: options?.poll ? REFETCH_INTERVAL_MS : false,
  });

/**
 * usePendingApprovals is what the header shows: the requests waiting on the
 * signed-in user, and nothing else.
 */
export const usePendingApprovals = () => {
  const { data: me } = useMe();
  const query = useAccessRequests({ poll: true });

  return {
    ...query,
    pending: pendingForApprover(query.data?.accessRequests ?? [], me?.user?.id),
  };
};
