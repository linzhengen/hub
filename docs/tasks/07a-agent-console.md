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
- **親の付け替えは 4b の rpc を通す。**画面から `parent_agent_id` を直接編集する
  導線を足さないこと。循環の検査は 4b が持っており、画面が独自の更新経路を作ると
  検査を迂回できてしまう。
- **Agent が自分自身として持っている権限を、一覧の目立つ位置に出す。**
  委譲（3b）が入るまで、Agent の実効権限は自分の付与そのものである。
  過剰な権限を持った Agent は見えていないと減らされない。

## 完了条件

- `pnpm lint && pnpm lint:eslint && pnpm test` が通る。
  React Compiler のルールは緩めない。
- 組織スイッチャで切り替えると Agent 一覧が切り替わる。
