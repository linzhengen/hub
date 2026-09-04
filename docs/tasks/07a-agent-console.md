# 7a. Agent コンソール

## 目的

Agent の一覧・構成・デプロイ状態を Web から扱えるようにする。

## やること

- `ui/web/src/pages/system/Agents.tsx`。
  `AuditLogs.tsx` / `ServiceAccounts.tsx` と同じ形:
  Ant Design 6 の `Table`、`useQuery` に
  `placeholderData: (previous) => previous`、`PageMeta` / `PageBreadcrumb`、
  co-located なテスト。
- `menus.yaml` に項目を足す（**サイドバーはルートではなく `resources` 行から作られる**）。
- `src/services/agent.ts` は `src/api/operations.ts` 経由。
  REST パスもクエリ文字列も手で書かない。
- 構成画面で MCP サーバー / Skill / サブ Agent を付け外しできる。
- **資格情報は作成時に 1 回だけ表示**し、読み戻す導線を作らない
  （`ServiceAccounts.tsx` の `CredentialsModal` が雛形）。

## 完了条件

- `pnpm lint && pnpm lint:eslint && pnpm test` が通る。
  React Compiler のルールは緩めない。
- 組織スイッチャで切り替えると Agent 一覧が切り替わる。
