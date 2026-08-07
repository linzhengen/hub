import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router';
import { render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import UserDropdown from '@/components/header/UserDropdown';
import { userService } from '@/services/user';
import type { GetMeResponse } from '@/services/user';

vi.mock('@/services/user', () => ({
  userService: { getMe: vi.fn() },
}));

const mockUseAuth = vi.hoisted(() => vi.fn());
vi.mock('@/providers/auth', () => ({ useAuth: mockUseAuth }));

/** Keycloak トークン由来の値。プロフィール更新後も古いまま残る */
const staleTokenUser = {
  isAuthenticated: true,
  isLoading: false,
  user: { id: 'user-1', name: 'old-name', email: 'old@example.com', emailVerified: true },
  login: vi.fn(),
  logout: vi.fn(),
};

const meResponse: GetMeResponse = {
  user: { id: 'user-1', username: 'new-name', email: 'new@example.com', status: 'STATUS_ACTIVE' },
  groups: [],
};

const renderDropdown = () => {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });

  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>
        <UserDropdown />
      </MemoryRouter>
    </QueryClientProvider>,
  );
};

beforeEach(() => {
  mockUseAuth.mockReturnValue(staleTokenUser);
  vi.mocked(userService.getMe).mockResolvedValue(meResponse);
});

afterEach(() => {
  document.body.innerHTML = '';
});

describe('UserDropdown', () => {
  it('トークンではなく GetMe の表示名を出す', async () => {
    renderDropdown();

    expect(await screen.findByText('new-name')).toBeTruthy();
    expect(screen.queryByText('old-name')).toBeNull();
  });

  it('ドロップダウンを開くと GetMe のメールアドレスを出す', async () => {
    renderDropdown();
    (await screen.findByRole('button')).click();

    expect(await screen.findByText('new@example.com')).toBeTruthy();
    expect(screen.queryByText('old@example.com')).toBeNull();
  });

  it('GetMe が未取得のうちはトークンの値にフォールバックする', async () => {
    vi.mocked(userService.getMe).mockRejectedValue(new Error('boom'));

    renderDropdown();

    expect(await screen.findByText('old-name')).toBeTruthy();
  });

  it('emailVerified はトークン由来の値を表示する', async () => {
    renderDropdown();
    (await screen.findByRole('button')).click();

    expect(await screen.findByText('✓ Email verified')).toBeTruthy();
  });
});
