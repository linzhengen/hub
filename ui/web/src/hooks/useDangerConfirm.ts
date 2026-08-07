import { App } from "antd";
import { useCallback } from "react";

export interface DangerConfirmOptions {
  title: string;
  content: string;
  okText?: string;
}

/**
 * 破壊的な操作（削除・割り当て解除など）の確認ダイアログ。
 *
 * ネイティブの `window.confirm` はフォーカストラップやキーボード操作、
 * ダークテーマへの追従がアプリ内の他モーダルと揃わないため使用しない。
 * `App.useApp()` 経由の `modal` は静的メソッドと異なり ConfigProvider の
 * コンテキスト（テーマ）を引き継ぐ。
 */
export const useDangerConfirm = () => {
  const { modal } = App.useApp();

  return useCallback(
    ({ title, content, okText = "Delete" }: DangerConfirmOptions, onConfirm: () => void) => {
      modal.confirm({
        title,
        content,
        okText,
        okButtonProps: { danger: true },
        cancelText: "Cancel",
        onOk: onConfirm,
      });
    },
    [modal],
  );
};
