# 設計メモ: アクセスガバナンス基盤と AI 運用

2026-08 / 対象: `server`, `ui/web`, `ai chat`

このメモは公開サイトには含まれない（`docs/build.mjs` は `docs/site/` だけを読む）。
方向性の決定と、そこから導かれる設計判断を残すためのもの。

> **続き**: このメモの穴 2（期限）と穴 3（人以外の主体）は埋まった。
> その後の「どこまでを一元化するのか」というコンセプトの決定は
> [`entitlement-catalog.md`](./entitlement-catalog.md) にある。
> 穴 1（スコープ付き権限）の保留はそこでも続いているが、
> スコープの単位が「グループ」ではなく「システム」に決まっている。

## 1. 現状と、何が足りないのか

hub は RBAC グラフ（user → group → role → permission → resource）と、
その変更をすべて記録する監査ログ、そして自然言語からその API を叩く AI chat を持つ。
土台は良い。足りないのは機能の数ではなく、モデルに空いた 3 つの穴である。

### 穴 1: 権限モデルが平坦

`internal/domain/auth/service.go` の `Enforce` は `(resource, action)` の
ワイルドカード一致だけを見る。「グループ X の管理者は X のメンバーだけ操作できる」
が表現できないため、**権限の委譲ができない**。管理者は全能かゼロかの二択になる。

### 穴 2: 付与に時間の概念がない

`user_groups` / `group_roles` に有効期限がない。付与は永久である。
そのため JIT アクセス・一時昇格・棚卸しのどれも実装できない。

### 穴 3: 主体が人しかいない

`AuditLog` は "The actor is always a person" を前提にしている。
CLI 向けの device flow / client credentials はあるのに、サービスアカウントや
API トークンを hub 側で発行・失効・監査できない。

### そして、AI chat の構造的な矛盾

`internal/infrastructure/ai/tool/toolbox.go` の `escalation` リストは、
hub の本業（権限の付け外し）を AI から完全に締め出している。その理由付け
—— ツール結果に含まれる他人が書いた文字列が、承認する当人を説得しにくる ——
は正しく、緩めるべきではない。

しかし結果として、chat は「読める検索窓 + ユーザー作成」で止まっている。
これが「機能不足」の体感の中身である。

## 2. 方向性

**IGA（アクセスガバナンス）を土台にし、その上に AI 運用を差別化として載せる。**

一行で言えば「AI がアクセス運用を回すが、権限変更は必ず人間の承認と監査を通る IAM」。

`escalation` リストは維持したまま、AI が本業に触れる道を作る。鍵は
**申請オブジェクトを間に挟むこと**である。

```
今:  AI ──[禁止]──> AddRolesToGroup
案:  AI ──> AccessRequest 起票 ──> 人が承認 ──> 期限付き付与 ──> 監査
```

AI は「権限を変える」のではなく「申請を起票する」。承認は人間が、AI の文脈の
外にあるコンソールで行う。`escalation` の正当性を一切崩さずに、
自然言語の入口（「田中さんに来週だけ本番の閲覧権限を」）が本業に届く。

## 3. 設計

### 3.1 期限付き付与

`user_groups` と `group_roles` に `expires_at TIMESTAMPTZ NULL` を足す。
`NULL` は無期限、すなわち現在の全行の意味を変えない。

`SelectUserAuthorizedPolices` に期限の条件を足す……のではなく、
**期限を `auth.Policy` に載せて `Enforce` で評価する**。理由はキャッシュにある。

> **キャッシュの罠（この設計で一番重要な点）**
>
> `infrastructure/auth/cache.go` の `cachingRepository` は `rbac_revisions` の
> 差分でのみキャッシュを捨てる。ところが**期限切れは書き込みを伴わない**ので
> トリガが発火せず revision が動かない。SQL 側で `expires_at > now()` を評価すると、
> 期限が切れた権限を TTL のあいだキャッシュが配り続ける。
> これは `rbac_revisions` が防ぐために作られた失敗そのものである。

したがって:

- `auth.Policy` に `ExpiresAt *time.Time` を追加する。
- SQL は期限切れ行も返す（フィルタしない）。
- `Enforce` が `policy.ExpiresAt` と現在時刻を比較して落とす。

こうするとキャッシュ層に手を入れずに済み、判定は常に評価時点の時刻で行われる。
キャッシュが保持するのは「期限つきポリシーの集合」であって「判定結果」ではない、
という不変条件で説明できる。

期限切れ行の物理削除は判定には不要。掃除は表示と行数の問題なので、
別途 sweeper を置くかどうかは後で決める（そのとき、失効を監査ログに
記録するかも併せて決める。現状、期限切れは誰の操作でもないのでイベントが無い）。

API は既存の `AddGroupsToUser` / `AddRolesToGroup` に `expires_at` を足す形にする。
新しい rpc は作らない。

### 3.2 アクセス申請・承認

新しい proto パッケージ `system/access/v1`。

```
AccessRequest
  id
  requester_user_id    申請した人
  subject_user_id      権限を受け取る人（自分への申請なら requester と同じ）
  group_id             入りたいグループ
  reason               申請理由（必須）
  requested_until      いつまで欲しいか（NULL = 無期限申請）
  status               PENDING | APPROVED | REJECTED | CANCELLED
  origin               CONSOLE | CLI | AI_CHAT
  session_id           origin = AI_CHAT のときの chat セッション
  decided_by_user_id
  decided_at
  decision_comment
```

> **実装時の訂正**: このメモは当初 `kind = GROUP | ROLE` と `target_id` を持たせる
> 設計にしていたが、hub のモデルでは**ロールはユーザーではなくグループに付く**ため、
> `subject_user_id` を持つ申請で ROLE を表現できない。申請は「ユーザーをグループに
> 入れる」ことに一本化した。JIT アクセスの用途はこれで足り、`AddGroupsToUser` という
> 単一の実行経路を保てる。グループへのロール付与は構造的な変更で、
> `escalation` リストの対象でもあるので、コンソールで人が行う。
> 種別が実際に 2 つになった時点で `kind` を足せばよい。

rpc は `CreateAccessRequest` / `ListAccessRequests` / `CancelAccessRequest` /
`DecideAccessRequest`。

承認時の挙動:

- `DecideAccessRequest(approved=true)` が内部で付与を実行し、
  `expires_at = requested_until` を設定する。
- 監査ログの当該行に申請 ID を残す。「なぜこの付与が行われたか」が
  監査ログから申請と理由まで辿れる状態にする。
- **自己承認は拒否する**（`decided_by == requester` を弾く）。
  申請と承認が同一人物なら、承認という仕組みは何も担保していない。

承認者は誰か: 第一段階では「その付与を自分で実行できる権限を持つ人」、
つまり既存 RBAC で `AddGroupsToUser` / `AddRolesToGroup` を許可された人とする。
グループごとの owner による承認は、穴 1（スコープ付き権限）を埋めてからでよい。
ここで承認者モデルを作り込むと、後でスコープ機構と二重になる。

### 3.3 AI との接続

`toolbox.go` に対する変更は 2 行で表現できる。

- `exposed` に `"/system.access.v1.AccessRequestService/CreateAccessRequest": true`
  を追加（write なので提案 → 承認フローを通る）。
- `escalation` に `DecideAccessRequest` を追加。**AI は永久に承認できない。**

この非対称性が設計の核心である。AI は起票までしかできず、
承認は必ず別の人間が別の画面で行う。プロンプトインジェクションが
成功したとして、得られるのは「人間が読んで却下できる申請」までになる。

`origin = AI_CHAT` と `session_id` を残すので、承認者は
「この申請は AI が会話から起こしたものだ」と分かった上で判断できる。

### 3.4 権限の逆引き（explain）

データは既に揃っていて、既存の 7-way join を逆向きに辿るだけで実装できる。
コストに対する価値が最も高い。

- `ExplainUserAccess(user_id, resource, action)`
  → 許可されている経路（group → role → permission）を全部列挙し、
    それぞれの有効期限を添える。「なぜこの人はこれができるのか」に答える。
- `ListPrincipalsForOperation(resource, action)`
  → 「この操作を今どれだけの人が実行できるか」に答える。

どちらも read only なので、`toolbox.go` の `exposed` に `false`（read）で
追加してよい。chat が「検索窓」から「調査の道具」に変わるのはこの 2 本による。

## 4. 実装順

| # | Issue | 内容 | 依存 |
|:--|:--|:--|:--|
| 1 | [#132](https://github.com/linzhengen/hub/issues/132) | 期限付き付与（schema + `Policy.ExpiresAt` + `Enforce` + API） | — |
| 2 | [#133](https://github.com/linzhengen/hub/issues/133) | 権限の逆引き / explain API | — |
| 3 | [#134](https://github.com/linzhengen/hub/issues/134) | アクセス申請・承認ドメイン（proto + domain + usecase） | #132 |
| 4 | [#135](https://github.com/linzhengen/hub/issues/135) | 申請・承認の Web UI | #134 |
| 5 | [#136](https://github.com/linzhengen/hub/issues/136) | AI 連携（`CreateAccessRequest` を提案可能に、`Decide` を遮断、explain を read ツールに） | #133, #134 |
| 6 | [#137](https://github.com/linzhengen/hub/issues/137) | サービスアカウント + API トークン | — |

#132 と #133 は互いに独立で、どちらも既存の面を壊さずに入る。
#134 以降は #132 の上に積む。

## 5. 今回は扱わないもの

- **スコープ付き権限（穴 1）**: 委譲の本命だが、`Enforce` の署名と
  ポリシーモデルの変更を伴うので独立した設計が要る。承認者モデルを
  ここに寄せない理由でもある。
- **マルチテナント / 組織**: 単一テナント前提を崩す判断はまだ早い。
- **アクセスレビュー（棚卸しキャンペーン）**: 1 と 3 が入ってからでよい。
  期限付き付与が普及すれば、棚卸しの対象そのものが減る。
- **webhook / outbox**: hub が下流の source of truth になる段階で。
