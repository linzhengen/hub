# 設計メモ: Agent Platform と Agent Identity

2026-09 / 対象: `server`, `ui/web`, `infra`

このメモは公開サイトには含まれない（`docs/build.mjs` は `docs/site/` だけを読む）。
`iga-direction.md` の続きであり、そこで先送りした 2 つの穴を埋める判断を残す。

## 1. 何をやろうとしているのか

hub の上に **Agent Platform** を載せる。A2A と ADK でエージェントを作り、
hub から deploy・公開・管理する。実行基盤は Google Cloud Agent Platform
（Vertex AI Agent Engine）を第一実装とする。ToB / ToC 両対応。

hub が管理するのは Agent そのものだけではない。**Agent・MCP サーバー・
Agent Skill・ユーザー**が一つの権限グラフの上に載っている状態を作る。

### なぜ hub がやるのか

エージェントを作る道具は世に多い。足りないのは
**「そのエージェントが誰の代わりに、何をしてよいか」を管理する基盤**である。

hub はそれを既に持っている。`service_accounts`（`000014`）の設計思想 ——

> A Keycloak client with service accounts enabled has a user of its own, and hub
> stores that user in `users` like any other. So a service account joins groups,
> holds roles and appears in the audit log through exactly the machinery a person
> does - no second authorization path, and no second kind of actor.

—— をそのまま Agent に延長すれば、Agent Identity は**新しい認可機構を 1 つも
足さずに**成立する。これがこの計画全体の背骨であり、差別化でもある。

## 2. 足りないもの

`iga-direction.md` §5「今回は扱わないもの」が挙げた 2 つが、そのまま前提条件になった。

### 穴 1: 権限モデルが平坦（再掲）

`internal/domain/auth/service.go` の `Enforce` は `(resource, action)` の
ワイルドカード一致しか見ない。「この Agent は、このユーザーの代理で、この範囲だけ」
が表現できない。**これは権限委譲の本命であり、サブ Agent の権限制御そのもの**である。

### 穴 2: 主体はいるが、境界がない

`service_accounts` で「人以外の主体」は埋まった（`iga-direction.md` の穴 3）。
残っているのは境界の方である。単一テナント前提のままでは、ToB（顧客ごとの分離）と
ToC（個人）を同じ経路に載せられない。

この 2 つは**同じ変更で埋まる**。`Enforce` に「どの範囲での問いか」を持ち込むことが、
組織分離であり、スコープ付き権限である。だから最初の Phase はそれになる。

## 3. 方向性

**Agent は「新しい種類の主体」ではなく、「`users` 行を持つ主体」である。**

一行で言えば「Agent が人の代わりに動くが、その権限は必ず委譲元の部分集合であり、
権限の変更だけは人間しかできない IAM」。

`iga-direction.md` が確立した非対称性 —— AI は起票までしかできず、承認は
人間が会話の外で行う —— は**そのまま維持し、Agent にも同じく適用する**。

## 4. 設計

### 4.1 Agent Identity

`agents` は登録簿である。認証は Keycloak のまま、認可経路は 1 本のまま。

```
agents
  id
  org_id             どの組織のものか
  user_id            agent が「として振る舞う」hub ユーザー（UNIQUE）
  name
  client_id          Keycloak client（hub-agent-<name>）
  keycloak_id
  parent_agent_id    サブ Agent。NULL はルート
  created_by_user_id
```

これで Agent はグループに入り、ロールを持ち、期限付き付与を受け、アクセス申請の
subject になり、監査ログに actor として現れる。`Enforce` も `audit_logs` も無改造で効く。

> **新しい actor 種別を作ってはいけない。**作った瞬間に「人の認可」と「機械の認可」が
> 分岐し、片方に空いた穴がもう片方から見えなくなる。`service_accounts` が
> `users` 行を持つと決めたのと同じ理由である。

### 4.2 委譲 = 判定時の交差

「ユーザーから Agent への権限委任」を、Agent の権限を**増やす**操作として実装しては
ならない。それは委譲ではなく昇格である。委譲の定義は 1 行で書ける。

> **Agent の実効権限 = Agent 自身の権限 ∩ 委譲元ユーザーの権限**

交差は**付与時ではなく判定時**に取る。

> **これは期限付き付与と同じ罠である。**付与時にだけ交差を取ると、委譲元ユーザーの
> 権限が後で失われても Agent は持ち続ける。「期限は SQL でフィルタせず `Enforce` が
> 時刻で落とす」という既存の判断とまったく同じ理由で、**交差は判定の瞬間に取る**。
> キャッシュが保持するのは「ポリシーの集合」であって「判定結果」ではない、という
> 不変条件はここでも保たれる。

したがって `auth.Request` に代理チェーンが載る。

```go
type Request struct {
    Subject    string   // 実行主体（agent の user_id）
    OnBehalfOf []string // 代理チェーン: [親agent, ..., 大元のユーザー]
    OrgId      string
    Object     string
    Action     string
}
```

判定は「チェーン上の**全員**が許可されていること」。チェーンは単調減少しかしない。

### 4.3 親 Agent → サブ Agent も、同じチェーンが伸びるだけ

親が子を呼ぶ = `OnBehalfOf` の先頭に親が積まれる。深さ制限と有効期限は
`delegations` 行が持つ。A2A の on-behalf-of セマンティクス（RFC 8693 の `act` 相当）
と一致するので、対外的にもそのまま説明できる。

サブ Agent 専用の権限機構は作らない。作れば、委譲とサブ Agent で 2 つの
「権限を狭める仕組み」が並び、どちらが効いているか誰にも分からなくなる。

### 4.4 委譲はトークンではなく、データベースの事実である

Agent は自分自身として Keycloak に認証する（service account と同じ
`client_credentials`）。hub を呼ぶときに代理を名乗り、interceptor が
`delegations` 行を引いてチェーンを `contextx` に載せる。

- **認証は Keycloak のまま。**認証経路は増えない。
- **委譲は DB 行なので即時失効できる。**`rbac_revisions` トリガを付ければ
  キャッシュも 1 秒以内に落ちる。トークンの残存期間という概念が存在しない。
- **Agent が任意の相手の代理を名乗れないのは、その行が無いからである。**
  行を作れるのは委譲元ユーザー本人か、アクセス申請の承認だけ。

自前で第二の署名トークンを発行する案は採らない。認可経路を 2 本にすることであり、
`service_accounts` が避けたのと同じ失敗になる。Keycloak Token Exchange は、
組織をまたぐ外部 A2A が必要になった段階で改めて検討する。

### 4.5 「何を持っているか」と「何をしてよいか」は別のゲート

- **Agent の構成**（どの MCP サーバー・どの Skill・どのサブ Agent を持つか）
  = 明示的な許可リスト。`toolbox.go` の `exposed` と同じ思想 ——
  増えた API が黙ってモデルの手に渡らない。
- **その道具で何をしてよいか** = RBAC + 委譲チェーン。

両方を通らないと実行されない。これは既存の `ToolBox.Tools()` / `Call()` の
二重チェック（一覧は提案を防ぎ、`Call` の中の検査が実際に効く）の Agent 版である。

この分離があるので、「Agent ごとの RBAC インスタンススコープ」は要らない。
組織スコープだけで足りる。

### 4.6 Runtime は抽象化し、Agent Engine を第一実装にする

```go
// internal/domain/ai/agent/runtime.go
type Runtime interface {
    Deploy(ctx context.Context, v *Version) (*Deployment, error)
    Status(ctx context.Context, d *Deployment) (Status, error)
    Endpoint(ctx context.Context, d *Deployment) (string, error) // A2A エンドポイント
    Delete(ctx context.Context, d *Deployment) error
}
```

> **これは軽い制約ではない。**現在 `go.mod` に GCP SDK は 1 つも無く、`infra/tf` には
> クラウドプロバイダが 1 つも無い（`infra/tf/AGENTS.md` が明言、backend は `local {}`）。
> Agent Engine を入れることは、このリポジトリに**初めてクラウドプロバイダを持ち込む**
> ことであり、Terraform state をローカルに置き続けられなくなる。
> だから Runtime の実装を差し替え可能にし、hub 本体は GCP に依存させない。
> クラウドを持ち込む判断は Phase 5 で単独の論点として扱う。

### 4.7 Authenticated Extended Agent Card

A2A 仕様は、公開 Agent Card と、認証済みクライアントにだけ返す
**Extended Agent Card**（追加スキルを含む）を定義している。

hub はここに**呼び出し元の RBAC で絞ったスキル一覧**を返せる。同じ Agent が、
相手によって違う能力を名乗る。これは `ToolBox.Tools(ctx)` が「そのユーザーが
呼べるものだけ返す」のを A2A に持ち上げたものである。Agent Card の
`securitySchemes` には Keycloak の OIDC discovery URL をそのまま宣言できる。

**これが hub にしか書けない機能である。**権限グラフを持っていないと書けない。

### 4.8 escalation は主体の種類に関係なく効く

Agent がどれだけ強い委譲を受けていても、`AddPermissionsToRole` /
`AddRolesToGroup` / `AddGroupsToUser` / `DecideAccessRequest` には触れない。
**委譲チェーンは `escalation` を通過する理由にならない。**

その正当性は `toolbox.go` に書かれたものと同じで、Agent でも変わらない ——
ツール結果に含まれる他人が書いた文字列が、承認する当人を説得しにくる。
むしろ Agent の方が、人の目を通らない回数が多いぶん強く効かせる必要がある。

`audit_logs.channel` の CHECK に `'agent'` を追加し、代理チェーンを記録する。
承認者が「これは誰の代理で動いた Agent の行為か」を分かった上で判断できるようにする。

## 5. 実装順

| # | Phase | 内容 | 依存 | 状態 |
|:--|:--|:--|:--|:--|
| 1 | 組織 + スコープ付き Enforce | 穴 1・穴 2 を同時に埋める。全体の土台 | — | **完了** |
| 2 | Agent Identity | `agents` 登録簿 + Keycloak client + `users` 行 | 1 | 未着手 |
| 3 | 委譲 | `delegations` + 代理チェーン付き `Enforce` + 監査 | 2 | 未着手 |
| 4 | Agent 構成 | MCP サーバー / Skill / サブ Agent の登録と紐付け | 2 | 未着手 |
| 5 | Runtime 抽象 + Agent Engine デプロイ | `Runtime` interface + agentengine 実装 | 4 | 未着手 |
| 6 | A2A 公開 | Agent Card / Extended Card / A2A エンドポイント | 3, 5 | 未着手 |
| 7 | Web UI + CLI | Agent 一覧・構成・委譲承認画面 | 2–6 | 未着手 |
| 8 | ToC | 個人 org、セルフサービス公開、Agent カタログ | 1, 6 | 未着手 |

Phase 1 は地味で、Agent は 1 体も動かない。それでも先にやった理由は、
`Enforce` の署名を変える作業だからである。Phase 3 の委譲は `Enforce` の上に積むので、
順番を逆にすると委譲を作り直すことになる。

残りは PR 1 本ぶんの単位に割って `docs/tasks/` にある。
このメモは判断を、そちらは手順を持つ。判断が変わったらこのメモを直し、
タスクをそれに従わせること —— 逆をやると、なぜそうしているのかが失われる。

## 6. 今回は扱わないもの

- **Agent ごとの RBAC インスタンススコープ**: §4.5 の二重ゲートで足りる。
  構成の許可リストと組織スコープの両方が既にあるのに三つ目を足すと、
  どれが効いているのか説明できなくなる。
- **Keycloak Token Exchange**: §4.4 のとおり、組織をまたぐ外部 A2A が
  現実の要求になってから。
- **Agent の課金・レート制限**: `ANTHROPIC_SESSION_TOKEN_BUDGET` と同じ問題を
  Agent 単位で解く必要が出るが、Agent が動いてからでよい。
- **マーケットプレイス / 公開カタログ**: Phase 8。ToC の形が決まってから。
