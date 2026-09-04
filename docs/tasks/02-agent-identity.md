# 2. Agent Identity

## 目的

Agent を hub が知っている主体にする。`agents` は登録簿であって、認証機構ではない。

## なぜ

`service_accounts`（`000014`）が確立した形をそのまま延長する。Agent が `users` 行を
持てば、グループに入り、ロールを持ち、期限付き付与を受け、アクセス申請の subject に
なり、監査ログに actor として現れる ——「認可経路は 1 本」のまま。

**新しい actor 種別を作らないこと。**作った瞬間に人の認可と機械の認可が分岐し、
片方に空いた穴がもう片方から見えなくなる。

## やること

- マイグレーション `000015_create_table_agents`。`org_id`, `user_id`(UNIQUE),
  `name`, `client_id`(UNIQUE), `keycloak_id`, `parent_agent_id`,
  `created_by_user_id`。`service_accounts` の DDL とコメントが雛形。
- ドメイン `internal/domain/ai/agent/`。命名は
  `serviceaccount.ClientIdFor` / `UsernameFor` / `EmailFor` に倣い
  `hub-agent-<name>` / `…@service-account.invalid`。
- proto `proto/ai/agent/v1/{model,service}.proto`。`ai/chat/v1` の隣。
  この名前空間は既に想定されている（`buf.yaml` の `RPC_NO_SERVER_STREAMING` 除外リストに
  まだ存在しない `ai/app/v1` と `ai/workflow/v1` が書かれている）。
- Keycloak client は既存経路を再利用する:
  `internal/infrastructure/oidc/admin/client.go` の `CreateServiceAccountClient`
  （`PublicClient=false, ServiceAccountsEnabled=true`、他のフローは全部 false）。
  失敗時に client を best-effort で消す `removeClient` の作法も同じ ——
  記録の無い client は、誰にも見えない生きた資格情報。
- `parent_agent_id` は列を作るだけでよい。親子の権限制御は 3b で入る。
  設定できるのは Create 時のみで、`UpdateAgent` rpc は作らない ——
  **循環を防いでいるのはこの「書き換え経路が無い」ことだけ**である
  （付け替えを足す 4b が検査も持つ）。

## この時点で成立していないこと

**Agent は「自分自身として」API を叩く。**「誰かの代理で」はまだ成立しない。

`agents` 行は身元であって、代理権ではない。Agent は自分の Keycloak client として
認証し、自分の付与だけで認可される。「私はこの人の代わりに動いている」と名乗る手段は
無く、それを検証する `delegations` 行も無い。したがって:

- 監査ログの actor は Agent である。`audit_logs.actor_user_id` を `agents.user_id` に
  つなげば、**どの Agent か・誰が登録したか**（`created_by_user_id`）までは辿れる。
- 辿れないのは**その 1 回を誰が起こさせたか**である。同じ Agent への呼び出しは、
  誰が起こしても同じ actor 1 つに潰れる。

これは 3a / 3b が埋める穴であって、ここで塞ぐものではない。ただし**危険になる時点は
はっきりしている** —— 何かが「人の代わりに」Agent を起動した瞬間である。その時点で、
呼び出した人は自分が持っていない権限を Agent 越しに行使でき、しかも痕跡が残らない
（confused deputy）。だから **Agent を人の代わりに起動する機能は 3b の後に置く**。

それまでの間、Agent に与える付与は「その Agent を呼べる人が自分で持っていても
おかしくない範囲」に留める。これは仕組みではなく運用上の約束なので、
仕組みに変わるのが 3b である。

## 完了条件

- `hub agent create-agent --name … --org-id …` が資格情報を 1 回だけ返す。
- 作られた Agent が `users` に現れ、`AddUsersToGroup` でグループに入れられる。
- Agent の client_credentials トークンで API を呼べ、監査ログに actor として残る。
- `ai/agent/v1` は AI の `exposed` マップに入れない —— 機械の身元を作れることは、
  人の付いていない権限を作れることであり、承認フローが自動化を防いでいる当のもの。
