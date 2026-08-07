import { MemoryRouter, Route, Routes } from 'react-router';
import { render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import AppSidebar from '@/layout/AppSidebar';

const mockUseSidebar = vi.hoisted(() => vi.fn());
vi.mock('@/context/sidebar', () => ({ useSidebar: mockUseSidebar }));

mockUseSidebar.mockReturnValue({
  isExpanded: true,
  isMobileOpen: false,
  isHovered: false,
  setIsHovered: vi.fn(),
});

/**
 * "Systems" のサブメニュー（Groups / Roles / Menus）が開いているかは、
 * トグルボタンに menu-item-active が付くかで判定する。
 */
const systemsToggle = () => screen.getByRole('button', { name: /Systems/ });
const isSystemsOpen = () => systemsToggle().className.includes('menu-item-active');

const renderAt = (pathname: string) =>
  render(
    <MemoryRouter initialEntries={[pathname]}>
      <Routes>
        <Route path="*" element={<AppSidebar />} />
      </Routes>
    </MemoryRouter>,
  );

afterEach(() => {
  document.body.innerHTML = '';
});

describe('AppSidebar のサブメニュー開閉', () => {
  it('サブ項目のルートにいるとき、そのサブメニューは開いている', () => {
    renderAt('/system/roles');

    expect(isSystemsOpen()).toBe(true);
  });

  it('どのサブ項目にも該当しないルートでは閉じている', () => {
    renderAt('/users');

    expect(isSystemsOpen()).toBe(false);
  });

  it('手動で開ける', async () => {
    renderAt('/users');
    expect(isSystemsOpen()).toBe(false);

    systemsToggle().click();

    await waitFor(() => expect(isSystemsOpen()).toBe(true));
  });

  it('該当ルートにいても手動で閉じられる', async () => {
    renderAt('/system/roles');
    expect(isSystemsOpen()).toBe(true);

    systemsToggle().click();

    await waitFor(() => expect(isSystemsOpen()).toBe(false));
  });

  it('別のサブ項目へ遷移するとルート由来の状態に戻る', async () => {
    renderAt('/system/roles');
    // 手動で閉じたあとに遷移すると、その操作は破棄されて再び開く
    systemsToggle().click();
    await waitFor(() => expect(isSystemsOpen()).toBe(false));

    screen.getByRole('link', { name: 'Groups' }).click();

    await waitFor(() => expect(isSystemsOpen()).toBe(true));
  });
});
