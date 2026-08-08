import { MemoryRouter, Route, Routes } from 'react-router';
import { render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import AppSidebar from '@/layout/AppSidebar';
import type { Menu } from '@/services/user';

const mockUseSidebar = vi.hoisted(() => vi.fn());
vi.mock('@/context/sidebar', () => ({ useSidebar: mockUseSidebar }));

const mockUseMenuResources = vi.hoisted(() => vi.fn());
vi.mock('@/hooks/useMenuResources', () => ({ useMenuResources: mockUseMenuResources }));

/** テスト用の動的メニューデータ（GetMeMenus レスポンスのツリー構造） */
const mockMenuResources: Menu[] = [
  {
    name: 'Dashboard',
    path: '/dashboard',
    meta: { title: 'Dashboard', icon: 'grid', order: 1 },
  },
  {
    name: 'Users',
    path: '/users',
    meta: { title: 'Users', icon: 'UserIcon', order: 2 },
  },
  {
    name: 'Systems',
    path: '/system',
    meta: { title: 'Systems', icon: 'carbon:settings', order: 3 },
    children: [
      {
        name: 'Groups',
        path: '/system/groups',
        meta: { title: 'Groups', icon: 'GroupIcon', order: 1 },
      },
      {
        name: 'Roles',
        path: '/system/roles',
        meta: { title: 'Roles', icon: 'Key', order: 2 },
      },
      {
        name: 'Menus',
        path: '/system/menus',
        meta: { title: 'Menus', icon: 'FolderIcon', order: 3 },
      },
    ],
  },
];

mockUseSidebar.mockReturnValue({
  isExpanded: true,
  isMobileOpen: false,
  isHovered: false,
  setIsHovered: vi.fn(),
});

mockUseMenuResources.mockReturnValue({
  data: mockMenuResources,
  isLoading: false,
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
