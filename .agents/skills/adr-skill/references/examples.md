# 記入例

hub の実際の判断で書いた例。**そのままコピーせず、書き方の水準を見る。**

どちらも共通しているのは、判断が「何を選んだか」ではなく
**「なぜ素直な方を選ばなかったか」**を中心に書かれている点である。

---

## 例 1: simple テンプレート

```markdown
---
status: accepted
date: 2026-09-10
decision-makers: linzhengen
---

# Record a delegation as a database fact, not a token

## 文脈と課題

Agent がユーザーの代理で動くとき、「誰の代理か」をどこに持つかを決める必要がある。

素直なのはトークンに載せることである。RFC 8693 の `act` クレームがあり、
Keycloak も token exchange を持っている。A2A の on-behalf-of とも語彙が揃う。

しかし hub には既に決まっている線がある。`service_accounts`（`000014`）が
「機械も `users` 行を持ち、**認可経路は 1 本しかない**」と決めた。
自前で第二の署名トークンを発行すれば、その線を越える。

もう一つ、失効の問題がある。トークンは発行された瞬間から、有効期限まで生きる。
委譲は「今日の午後だけ」「この作業のあいだだけ」という短い単位で渡され、
**途中で取り消せることが前提**になる。

## 判断

委譲を `delegations` テーブルの行として持つ。トークンには載せない。

- Agent は自分自身として Keycloak に認証する（service account と同じ client credentials）
- hub を呼ぶときに `hub-on-behalf-of` ヘッダで代理を名乗る
- interceptor が `delegations` を引き、チェーンを `contextx` に載せる
- 行を作れるのは委譲元ユーザー本人か、アクセス申請の承認だけ

### 非目標

- Keycloak Token Exchange は導入しない。組織をまたぐ外部 A2A が現実の要求に
  なった時点で改めて検討する
- 委譲の UI はこの判断に含まない（別タスク）

## 帰結

- 良い点: 認証は Keycloak のまま。認証経路も認可経路も増えない
- 良い点: **即時失効できる。**`rbac_revisions` トリガを付ければ 1 秒以内に効く。
  トークンの残存期間という概念が存在しない
- 良い点: Agent が任意の相手の代理を名乗れないのは「その行が無いから」で、
  署名検証の正しさに依存しない
- 悪い点: 呼び出しごとに 1 テーブル引く。ただし `Enforce` と同じキャッシュに乗る
- 悪い点: 組織をまたぐ外部 A2A では、この方式は使えない。そのとき判断し直しになる
- どちらでもない点: A2A の on-behalf-of と語彙は揃うが、実装は揃わない

## 実装計画

- **影響範囲**:
  - `server/db/migrations/postgres/` — `delegations` テーブル
  - `server/internal/domain/auth/` — `Request.OnBehalfOf`
  - `server/internal/interface/grpc/interceptor/interceptor.go` — ヘッダの解決
  - `server/proto/system/delegation/v1/`
- **依存**: なし
- **倣う型**:
  - ヘッダの解決は `requestedOrg`（`hub-org` を読む関数）と同じ形
  - 判定は期限付き付与と同じく**判定時**に落とす（`auth.Service.Enforce`）
- **避けること**:
  - 付与時に権限を交差させない。委譲元の権限が後で失われても Agent が持ち続ける
  - `delegations` に `rbac_revisions` トリガを付け忘れないこと。
    認可判定がこの表を読むため、無いと失効が最大 TTL 分効かない
- **移行**: 不要（新規テーブル）

### 検証

- [ ] `make test` が通る
- [ ] `golangci-lint run ./...` が 0 issues
- [ ] 委譲を失効させると 1 秒以内に効く（`HUB_TEST_DSN` を設定した結合テスト）
- [ ] Agent が委譲元より広い権限を行使できない（委譲元の付与を落とすと即座に狭まる）
- [ ] チェーンが深くなっても権限が広がらない（単調減少のテスト）
- [ ] `escalation` の rpc は、どれだけ強い委譲があっても通らない

## 却下した案

- **Keycloak Token Exchange**: 語彙は正しいが、失効が発行済みトークンの寿命に縛られる。
  委譲は短命で取り消せることが前提なので、そこが合わない。
- **hub が自前で署名トークンを発行する**: 認可経路が 2 本になる。
  `service_accounts` が避けたのと同じ失敗で、片方に空いた穴がもう片方から見えなくなる。

## 補足

- 方向性メモ: `docs/design/agent-platform-direction.md` §4.4
- タスク: `docs/tasks/03a-delegation-model.md`, `docs/tasks/03b-delegation-enforcement.md`
- 再考の引き金: 組織をまたぐ外部 A2A が必要になったとき
```

---

## 例 2: MADR テンプレート（抜粋）

選択肢が 3 つ以上あるときの形。文脈・実装計画・検証は例 1 と同じ水準で書く。

```markdown
---
status: accepted
date: 2026-09-04
decision-makers: linzhengen
---

# Scope a permission to an organization

## 文脈と課題

`Enforce` は `(resource, action)` のワイルドカード一致しか見ない。
「グループ X の管理者は X のメンバーだけ操作できる」が表現できず、
管理者は全能かゼロかの二択になる（`docs/design/iga-direction.md` の「穴 1」）。

ToB / ToC 両対応と Agent の権限委譲は、どちらもこの穴の上に建てられない。

## 判断の軸

- 既存の付与の意味を変えないこと（単一テナント運用が壊れない）
- キャッシュの不変条件を壊さないこと
- 規則が 1 か所にだけ書かれること

## 検討した選択肢

- A: 既存の group ツリーをテナント境界に流用する
- B: `organizations` を第一級で導入し、group が属する
- C: 権限文字列にスコープを埋め込む（`org:X/api.*`）

## 結論

**B** を採る。境界を持つのは group である —— group が user と permission を
つなぐ唯一の辺だからで、そこに組織を置けば、通る経路すべてが組織を通る。

### 非目標

- Agent ごとのインスタンススコープは作らない。構成の許可リストと組織スコープで足りる
- 一覧の絞り込みはこの判断に含むが、行レベルの所有権は含まない

### 帰結

- 良い点: 穴 1 と「マルチテナント」が同じ変更で埋まる
- 良い点: PLATFORM の付与が全組織に届くので、移行が既存の判定を 1 つも変えない
- 悪い点: `Enforce` の署名が変わる。呼び出し側すべてに影響する
- どちらでもない点: 組織を指定しないリクエストは従来どおりの意味になる。
  互換のためだが、絞り込みは呼び出し側が明示したときだけ効くということでもある

## 実装計画

- **影響範囲**: `server/internal/domain/auth/`, `.../interceptor/interceptor.go`,
  `.../postgres/query/auth.sql`, `server/proto/system/organization/v1/`
- **倣う型**: 期限（`ExpiresAt`）とまったく同じ。SQL でフィルタせず、
  `grant.reaches` が判定時に落とす
- **避けること**: ポリシークエリに `org_id = $n` を足さないこと。
  キャッシュが「判定結果」を持つことになり、`rbac_revisions` が防いでいる失敗に戻る
- **移行**: 既存 group はすべて PLATFORM 組織に入る

### 検証

- [ ] `make test` / `golangci-lint run ./...`
- [ ] 移行前後で既存ユーザーの `Explain` 結果が一致する
- [ ] `organizations` への書き込みが `rbac_revisions` を動かす（結合テスト）
- [ ] 「全組織に届く」規則が `Kind.AppliesEverywhere` 1 か所にしか書かれていない

## 各選択肢の長所と短所

### A: group ツリーの流用

- 良い点: スキーマ変更が最小
- 悪い点: group が「所属」と「テナント」の二重の意味を持ち、
  `Enforce` のスコープ判定が曖昧になる

### C: 権限文字列へのスコープ埋め込み

- 良い点: `Enforce` の署名が変わらない
- 悪い点: 文字列のパースが増え、ワイルドカード規則と干渉する。
  「この人はどの組織にいるか」に答えられない

## 補足

- 方向性メモ: `docs/design/agent-platform-direction.md`
- 先送りされていた論点: `docs/design/iga-direction.md` §5
```

---

## この 2 例から読み取ること

- **文脈が、解決策ではなく問題を書いている。**「なぜ今」が具体的
- **却下した案に、藁人形が無い。**どちらも本当に成立しうる案で、
  落とした理由がこのリポジトリの既存の判断に紐づいている
- **実装計画が実在のパスと実在の型を指している**
- **検証が走るコマンドで書かれ、構造の検証（規則が 1 か所か）も入っている**
- **悪い点が書かれている。**良い点しか無い ADR は、まだ考え切れていない
