import { App } from "antd";

/**
 * API の成功・エラー通知。
 *
 * `App.useApp()` 経由の `message` を使うことで ConfigProvider の
 * テーマ（light/dark）を引き継ぐ。
 * `sonner` の静的 `toast` は Toaster が未 mount だと表示されず、
 * ネイティブ `alert` はテーマに追従しないため使用しない。
 */
export const useNotify = () => {
  const { message } = App.useApp();

  return {
    success: (content: string) => message.success(content),
    error: (content: string) => message.error(content),
  };
};
