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
  `expires_at`, `max_depth`, `revoked_at`, `reason`、および**スコープ**。
- **スコープ列を最初から入れる。**委譲の実効権限は
  `Agent 自身 ∩ 委譲元 ∩ このスコープ` の三項で、`NULL` が委譲元の全権を意味する。
  「自分の権限を全部渡すか、何も渡さないか」しか選べないのは委譲ではない。
  列を後から足すのは、既にある行を書き直すことになる。
  ADR: [Narrow a delegation with its own scope](../decisions/2026-09-04-narrow-a-delegation-with-its-own-scope.md)
- **`rbac_revisions` トリガを付ける。** 認可判定がこの表を読む。
  忘れると失効が最大 TTL 分効かない。
- proto `system/delegation/v1`。`CreateDelegation` / `ListDelegations` /
  `RevokeDelegation`。
- 行を作れるのは委譲元ユーザー本人か、アクセス申請の承認だけ。
  第三者が他人の代理権を配れるようにしない。
- **委譲元ユーザーの削除は委譲行を消す**（`principal_user_id` は `ON DELETE CASCADE`）。
  `agents` の外部キーが `RESTRICT` なのと逆なのは、残ったときに何が起きるかが逆だから ——
  記録の無い資格情報は危険だが、主体のいない代理権は単に効かないより悪い。
- `expires_at` は `user_groups` / `group_roles` と同じ意味（NULL = 無期限）。
  ただし委譲は無期限を既定にしない —— 期限のある委譲が普通で、無期限が例外。

## 完了条件

- 委譲を作り、一覧し、失効できる。
- スコープを狭めた委譲が、その範囲の外を拒む。
- 失効が 1 秒以内に効く（`rbac_revisions` の結合テストで確認）。
- 自分以外を principal にした委譲を、権限のない者が作れない。
- `DecideAccessRequest` と同じく、**AI からは作らせない**（`escalation` に入れる）。
