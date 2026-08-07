/**
 * Vitest のグローバルセットアップ。
 *
 * jsdom は matchMedia を実装していないが、Ant Design のレスポンシブ対応
 * コンポーネント（Descriptions / Grid など）は初期描画で参照するため、
 * 常に「マッチしない」を返すスタブを用意する。
 */
if (!window.matchMedia) {
  window.matchMedia = (query: string): MediaQueryList =>
    ({
      matches: false,
      media: query,
      onchange: null,
      addListener: () => {},
      removeListener: () => {},
      addEventListener: () => {},
      removeEventListener: () => {},
      dispatchEvent: () => false,
    }) as MediaQueryList;
}
