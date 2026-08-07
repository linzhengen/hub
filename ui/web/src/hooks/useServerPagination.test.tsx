import { act, renderHook } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { useServerPagination } from '@/hooks/useServerPagination';

describe('useServerPagination', () => {
  it('最初のページを既定のページサイズで要求する', () => {
    const { result } = renderHook(() => useServerPagination());

    expect(result.current.params).toEqual({ limit: 20, offset: 0 });
  });

  it('ページ送りで offset を進める', () => {
    const { result } = renderHook(() => useServerPagination());

    act(() => result.current.tableProps(100, 'users').onChange?.(3, 20));

    expect(result.current.params).toEqual({ limit: 20, offset: 40 });
  });

  it('ページサイズの変更を offset に反映する', () => {
    const { result } = renderHook(() => useServerPagination());

    act(() => result.current.tableProps(100, 'users').onChange?.(2, 50));

    expect(result.current.params).toEqual({ limit: 50, offset: 50 });
  });

  it('絞り込みが変わると先頭ページに戻す', () => {
    // 絞り込み後の結果に4ページ目が存在するとは限らないため。
    const { result, rerender } = renderHook(({ filter }) => useServerPagination(filter), {
      initialProps: { filter: '' },
    });

    act(() => result.current.tableProps(100, 'users').onChange?.(4, 20));
    expect(result.current.params.offset).toBe(60);

    rerender({ filter: 'taro' });

    // 効果ではなくレンダリング中に調整するので、絞り込み変更後の最初の
    // リクエストが古い offset で飛ぶことはない。
    expect(result.current.params).toEqual({ limit: 20, offset: 0 });
  });

  it('絞り込みが変わらなければページを保つ', () => {
    const { result, rerender } = renderHook(({ filter }) => useServerPagination(filter), {
      initialProps: { filter: 'taro' },
    });

    act(() => result.current.tableProps(100, 'users').onChange?.(2, 20));
    rerender({ filter: 'taro' });

    expect(result.current.params.offset).toBe(20);
  });

  it('総件数を表示に渡す', () => {
    const { result } = renderHook(() => useServerPagination());
    const props = result.current.tableProps(137, 'users');

    expect(props.total).toBe(137);
    expect(props.showTotal?.(137, [1, 20])).toBe('1-20 of 137 users');
  });
});
