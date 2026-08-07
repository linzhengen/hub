import { useEffect, useState } from 'react';

/**
 * Returns `value` once it has stopped changing for `delay` milliseconds.
 *
 * Search boxes now filter on the server, so without this every keystroke would
 * be a request.
 */
export function useDebouncedValue<T>(value: T, delay = 300): T {
  const [debounced, setDebounced] = useState(value);

  useEffect(() => {
    const timer = setTimeout(() => setDebounced(value), delay);
    return () => clearTimeout(timer);
  }, [value, delay]);

  return debounced;
}
