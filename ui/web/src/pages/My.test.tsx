import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { App } from 'antd';
import { render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { My } from '@/pages/My';
import { userService } from '@/services/user';
import type { GetMeResponse, UpdateMeResponse } from '@/services/user';

vi.mock('@/services/user', () => ({
  userService: {
    getMe: vi.fn(),
    updateMe: vi.fn(),
    sendMeVerifyEmail: vi.fn(),
  },
}));

const mockUseAuth = vi.hoisted(() => vi.fn());
vi.mock('@/providers/auth', () => ({ useAuth: mockUseAuth }));

const toastSuccess = vi.hoisted(() => vi.fn());
const toastError = vi.hoisted(() => vi.fn());
vi.mock('@/hooks/useNotify', () => ({
  useNotify: () => ({ success: toastSuccess, error: toastError }),
}));

const meResponse: GetMeResponse = {
  user: { id: 'user-1', username: 'linzhengen', email: 'user@example.com', status: 'STATUS_ACTIVE' },
  groups: [],
};

const renderMy = () => {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });

  return render(
    <QueryClientProvider client={queryClient}>
      <App>
        <My />
      </App>
    </QueryClientProvider>,
  );
};

/** 編集モーダルを開き、フォームの値を差し替えて Save する */
const submitProfileEdit = async (values: { name: string; email: string }) => {
  renderMy();
  await screen.findByText('linzhengen');

  screen.getByRole('button', { name: /Edit My/ }).click();

  const nameInput = await screen.findByLabelText('Display Name');
  const emailInput = screen.getByLabelText('Email Address');
  setInputValue(nameInput as HTMLInputElement, values.name);
  setInputValue(emailInput as HTMLInputElement, values.email);

  screen.getByRole('button', { name: /Save Changes/ }).click();
};

/** React 管理下の input に値を入れ、change イベントを発火させる */
const setInputValue = (input: HTMLInputElement, value: string) => {
  const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, 'value')?.set;
  setter?.call(input, value);
  input.dispatchEvent(new Event('input', { bubbles: true }));
};

const authUser = (emailVerified: boolean) => ({
  isAuthenticated: true,
  isLoading: false,
  user: { id: 'user-1', name: 'linzhengen', email: 'user@example.com', emailVerified },
  login: vi.fn(),
  logout: vi.fn(),
});

/** Resend Verification ボタンはメール未確認のときだけ描画される */
const mockUnverifiedUser = () => mockUseAuth.mockReturnValue(authUser(false));

beforeEach(() => {
  mockUseAuth.mockReturnValue(authUser(true));
  vi.mocked(userService.getMe).mockResolvedValue(meResponse);
});

afterEach(() => {
  document.body.innerHTML = '';
});

describe('My', () => {
  it('GetMe のレスポンスを表示する', async () => {
    renderMy();

    expect(await screen.findByText('linzhengen')).toBeTruthy();
    expect(screen.getAllByText('user@example.com').length).toBeGreaterThan(0);
  });

  it('保存時に UpdateMe を呼び出す', async () => {
    vi.mocked(userService.updateMe).mockResolvedValue(meResponse as UpdateMeResponse);

    await submitProfileEdit({ name: 'renamed', email: 'renamed@example.com' });

    await waitFor(() =>
      expect(userService.updateMe).toHaveBeenCalledWith({ username: 'renamed', email: 'renamed@example.com' }),
    );
  });

  it('更新成功時はレスポンスを表示に反映し、再取得しない', async () => {
    const updated: UpdateMeResponse = {
      user: { ...meResponse.user, username: 'renamed', email: 'renamed@example.com' },
      groups: [],
    };
    vi.mocked(userService.updateMe).mockResolvedValue(updated);

    await submitProfileEdit({ name: 'renamed', email: 'renamed@example.com' });

    expect(await screen.findByText('renamed')).toBeTruthy();
    expect(toastSuccess).toHaveBeenCalledWith('Profile updated successfully');
    expect(userService.getMe).toHaveBeenCalledTimes(1);
  });

  it('更新失敗時はエラーメッセージを通知し、成功トーストを出さない', async () => {
    vi.mocked(userService.updateMe).mockRejectedValue(new Error('username already taken'));

    await submitProfileEdit({ name: 'renamed', email: 'renamed@example.com' });

    await waitFor(() => expect(toastError).toHaveBeenCalledWith('username already taken'));
    expect(toastSuccess).not.toHaveBeenCalled();
  });

  it('メール未確認のとき Resend Verification で SendMeVerifyEmail を呼ぶ', async () => {
    mockUnverifiedUser();
    vi.mocked(userService.sendMeVerifyEmail).mockResolvedValue({});

    renderMy();
    (await screen.findByRole('button', { name: /Resend Verification/ })).click();

    await waitFor(() => expect(userService.sendMeVerifyEmail).toHaveBeenCalledTimes(1));
    await waitFor(() => expect(toastSuccess).toHaveBeenCalledWith('Verification email sent'));
  });

  it('メール再送に失敗したらエラーを通知し、成功トーストを出さない', async () => {
    mockUnverifiedUser();
    vi.mocked(userService.sendMeVerifyEmail).mockRejectedValue(new Error('smtp not configured'));

    renderMy();
    (await screen.findByRole('button', { name: /Resend Verification/ })).click();

    await waitFor(() => expect(toastError).toHaveBeenCalledWith('smtp not configured'));
    expect(toastSuccess).not.toHaveBeenCalled();
  });

  it('メール確認済みなら Resend Verification を表示しない', async () => {
    renderMy();
    await screen.findByText('linzhengen');

    expect(screen.queryByRole('button', { name: /Resend Verification/ })).toBeNull();
  });

  it('GetMe が失敗しても認証情報にフォールバックして表示する', async () => {
    vi.mocked(userService.getMe).mockRejectedValue(new Error('boom'));
    vi.spyOn(console, 'error').mockImplementation(() => {});

    renderMy();

    expect(await screen.findByText('linzhengen')).toBeTruthy();
    await waitFor(() => expect(toastError).toHaveBeenCalledWith('Failed to load profile data from server'));
  });
});
