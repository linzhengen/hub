import { useCallback, useState } from 'react';
import type { TablePaginationConfig } from 'antd';

/**
 * The largest page the server will serve, mirroring
 * `pagination.MaxLimit` in `server/internal/usecase/pagination`.
 *
 * Asking for more is not an error — the server silently caps it — so a screen
 * that needs "everything" has to know where everything stops.
 */
export const MAX_PAGE_SIZE = 200;

const DEFAULT_PAGE_SIZE = 20;

export interface ServerPagination {
  /** Query parameters to send to a list endpoint. */
  params: { limit: number; offset: number };
  /** Props for antd's `Table pagination`, driving the server rather than the client. */
  tableProps: (total: number, noun: string) => TablePaginationConfig;
}

/**
 * Drives a table from the server's paging rather than the browser's.
 *
 * antd's Table paginates whatever array it is handed, which only looked right
 * while that array was the entire table. Now that list endpoints return a
 * bounded page, the component has to ask for each page instead.
 *
 * `filterKey` is whatever the caller is filtering by. When it changes the page
 * returns to the first, because page 4 of the old result is meaningless — and
 * may not exist — under the new one.
 */
export function useServerPagination(filterKey: unknown = null, pageSize = DEFAULT_PAGE_SIZE): ServerPagination {
  const [page, setPage] = useState(1);
  const [size, setSize] = useState(pageSize);
  const [appliedFilter, setAppliedFilter] = useState(filterKey);

  // Adjusted while rendering rather than in an effect. An effect would let one
  // render — and so one request — go out with the new filter but the old
  // offset before the reset landed.
  let currentPage = page;
  if (appliedFilter !== filterKey) {
    setAppliedFilter(filterKey);
    setPage(1);
    currentPage = 1;
  }

  const tableProps = useCallback(
    (total: number, noun: string): TablePaginationConfig => ({
      current: currentPage,
      pageSize: size,
      total,
      showSizeChanger: true,
      pageSizeOptions: [10, 20, 50, 100],
      showQuickJumper: true,
      showTotal: (count, range) => `${range[0]}-${range[1]} of ${count} ${noun}`,
      onChange: (nextPage, nextSize) => {
        setPage(nextPage);
        setSize(nextSize);
      },
    }),
    [currentPage, size],
  );

  return {
    params: { limit: size, offset: (currentPage - 1) * size },
    tableProps,
  };
}
