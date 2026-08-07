import { App } from 'antd';
import { render, screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { useDangerConfirm } from '@/hooks/useDangerConfirm';

const TriggerButton = ({ onConfirm }: { onConfirm: () => void }) => {
  const confirmDanger = useDangerConfirm();

  return (
    <button
      onClick={() =>
        confirmDanger({ title: 'Delete user', content: 'Are you sure you want to delete this user?' }, onConfirm)
      }
    >
      delete
    </button>
  );
};

const RemoveTrigger = ({ onConfirm }: { onConfirm: () => void }) => {
  const confirmDanger = useDangerConfirm();

  return (
    <button
      onClick={() =>
        confirmDanger(
          { title: 'Remove group', content: 'Are you sure you want to remove this group?', okText: 'Remove' },
          onConfirm,
        )
      }
    >
      remove
    </button>
  );
};

const renderTrigger = (onConfirm: () => void) =>
  render(
    <App>
      <TriggerButton onConfirm={onConfirm} />
    </App>,
  );

const openConfirm = async (onConfirm: () => void) => {
  renderTrigger(onConfirm);
  screen.getByRole('button', { name: 'delete' }).click();
  await screen.findByText('Are you sure you want to delete this user?');
};

describe('useDangerConfirm', () => {
  it('ネイティブの confirm ではなくモーダルを表示する', async () => {
    const nativeConfirm = vi.spyOn(window, 'confirm');

    await openConfirm(vi.fn());

    expect(screen.getByRole('dialog', { name: 'Delete user' })).toBeTruthy();
    expect(nativeConfirm).not.toHaveBeenCalled();
  });

  it('確認前はコールバックを呼ばない', async () => {
    const onConfirm = vi.fn();

    await openConfirm(onConfirm);

    expect(onConfirm).not.toHaveBeenCalled();
  });

  it('OK を押すとコールバックを実行する', async () => {
    const onConfirm = vi.fn();
    await openConfirm(onConfirm);

    screen.getByRole('button', { name: 'Delete' }).click();

    await waitFor(() => expect(onConfirm).toHaveBeenCalledTimes(1));
  });

  it('Cancel を押してもコールバックを実行しない', async () => {
    const onConfirm = vi.fn();
    await openConfirm(onConfirm);

    screen.getByRole('button', { name: 'Cancel' }).click();

    // onOk は非同期に解決されるため、1 tick 待ってから呼ばれていないことを確認する
    await new Promise((resolve) => setTimeout(resolve, 0));
    expect(onConfirm).not.toHaveBeenCalled();
  });

  it('okText を指定するとボタンのラベルに反映される', async () => {
    render(
      <App>
        <RemoveTrigger onConfirm={vi.fn()} />
      </App>,
    );
    screen.getByRole('button', { name: 'remove' }).click();

    expect(await screen.findByRole('button', { name: 'Remove' })).toBeTruthy();
  });
});
