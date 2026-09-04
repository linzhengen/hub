# 3a. 委譲のモデルと API

## 目的

「このユーザーの代理で動いてよい」という事実を、失効できる行として持つ。

## なぜ

委譲をトークンにしない。DB の行なら即時失効でき、認証経路も増えない
（Agent は自分自身として Keycloak に認証する）。Agent が任意の相手の代理を
名乗れないのは、**その行が無いから**である。

## やること

- マイグレーション `000016_create_table_delegations`。
  `agent_id`, `principal_user_id`（誰の代理か）, `granted_by_user_id`,
  `expires_at`, `max_depth`, `revoked_at`, `reason`。
- **`rbac_revisions` トリガを付ける。** 認可判定がこの表を読む。
  忘れると失効が最大 TTL 分効かない。
- proto `system/delegation/v1`。`CreateDelegation` / `ListDelegations` /
  `RevokeDelegation`。
- 行を作れるのは委譲元ユーザー本人か、アクセス申請の承認だけ。
  第三者が他人の代理権を配れるようにしない。
- `expires_at` は `user_groups` / `group_roles` と同じ意味（NULL = 無期限）。
  ただし委譲は無期限を既定にしない —— 期限のある委譲が普通で、無期限が例外。

## 完了条件

- 委譲を作り、一覧し、失効できる。
- 失効が 1 秒以内に効く（`rbac_revisions` の結合テストで確認）。
- 自分以外を principal にした委譲を、権限のない者が作れない。
- `DecideAccessRequest` と同じく、**AI からは作らせない**（`escalation` に入れる）。
