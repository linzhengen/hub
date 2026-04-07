import React, { useEffect, useRef } from 'react';
import { Modal } from 'antd';
import { useAuth } from '@/providers/AuthProvider';

/**
 * 401エラー（セッション切れ）を検知してモーダルを表示するコンポーネント
 */
export const AuthErrorHandler: React.FC = () => {
  const { logout } = useAuth();
  const isModalShown = useRef(false);

  useEffect(() => {
    const handleUnauthorized = () => {
      if (isModalShown.current) return;

      isModalShown.current = true;
      Modal.error({
        title: 'セッションが切れました',
        content: 'もう一度ログインしてください。',
        okText: 'ログイン画面へ',
        onOk: () => {
          isModalShown.current = false;
          logout();
        },
        // モーダルを閉じたり枠外をクリックした際もログアウト（ログイン画面へ遷移）させる
        onCancel: () => {
          isModalShown.current = false;
          logout();
        },
      });
    };

    window.addEventListener('api-unauthorized', handleUnauthorized);

    return () => {
      window.removeEventListener('api-unauthorized', handleUnauthorized);
    };
  }, [logout]);

  return null;
};
