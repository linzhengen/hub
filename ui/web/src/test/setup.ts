/**
 * Vitest のグローバルセットアップ。
 *
 * jsdom は matchMedia を実装していないが、Ant Design のレスポンシブ対応
 * コンポーネント（Descriptions / Grid など）は初期描画で参照するため、
 * 常に「マッチしない」を返すスタブを用意する。
 */
import { cleanup } from '@testing-library/react';
import { afterEach } from 'vitest';

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

/**
 * jsdom は ResizeObserver も実装していない。Ant Design が要素の寸法を追う
 * コンポーネント（autoSize な Input.TextArea など）はマウント時に監視を始めるため、
 * 何も通知しないスタブを用意する。無いと描画時に ReferenceError で落ちる。
 */
if (!globalThis.ResizeObserver) {
  globalThis.ResizeObserver = class {
    observe() {}
    unobserve() {}
    disconnect() {}
  } as unknown as typeof ResizeObserver;
}

/**
 * Node 26 は `localStorage` をグローバルに持つが、`--localstorage-file` が無いと
 * 参照した時点で例外になる。これが jsdom の実装を覆い隠すため、
 * ブラウザでは動くコードがテストでだけ「保存できない」状態になる。
 *
 * 保存された値を読む側（組織スイッチャなど）は例外を握り潰して既定値に倒すので
 * テストは落ちずに素通りしてしまい、壊れていることに気づけない。
 * そのため、メモリ上に持つ最小の実装へ差し替える。
 */
const localStorageAvailable = (() => {
  try {
    return typeof globalThis.localStorage?.getItem === 'function';
  } catch {
    return false;
  }
})();

if (!localStorageAvailable) {
  const store = new Map<string, string>();
  Object.defineProperty(globalThis, 'localStorage', {
    configurable: true,
    writable: true,
    value: {
      get length() {
        return store.size;
      },
      key: (index: number) => [...store.keys()][index] ?? null,
      getItem: (key: string) => store.get(key) ?? null,
      setItem: (key: string, value: string) => {
        store.set(key, String(value));
      },
      removeItem: (key: string) => {
        store.delete(key);
      },
      clear: () => {
        store.clear();
      },
    } satisfies Storage,
  });
}

/**
 * jsdom は scrollIntoView を実装していない。会話が伸びたときに末尾へ追従する
 * 実装がこれを呼ぶため、何もしないスタブを用意する。
 */
if (!Element.prototype.scrollIntoView) {
  Element.prototype.scrollIntoView = () => {};
}

/**
 * Testing Library の自動クリーンアップは `globals: true` のときだけ登録される。
 * このプロジェクトは vitest の API を明示 import しているため自前で呼ぶ必要があり、
 * 呼ばないと React のルートがマウントされたままテストファイルが終了する。
 * 残ったスケジュール済みの処理が jsdom の破棄後に走り、
 * "ReferenceError: window is not defined" として散発的に落ちる。
 */
afterEach(cleanup);
