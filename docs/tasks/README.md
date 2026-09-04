# 残タスク: Agent Platform

`docs/design/agent-platform-direction.md` のロードマップを、PR 1 本ぶんの単位に割ったもの。
このディレクトリは公開サイトに含まれない（`docs/build.mjs` は `docs/site/` だけを読む）。

設計判断はここに書かない。**なぜそうするのか**は方向性メモ側にあり、
ここにあるのは「何を、どの順で、何が満たされたら終わりか」だけである。
判断が変わったらメモを直し、タスクはそれに従わせる。

## 一覧

| # | タスク | 依存 | 状態 |
|:--|:--|:--|:--|
| 1 | 組織 + スコープ付き Enforce | — | **完了** |
| 2 | [Agent Identity](02-agent-identity.md) | 1 | 未着手 |
| 3a | [委譲のモデルと API](03a-delegation-model.md) | 2 | 未着手 |
| 3b | [代理チェーンつきの認可](03b-delegation-enforcement.md) | 3a | 未着手 |
| 4a | [MCP サーバーと Skill の登録簿](04a-mcp-and-skill-registry.md) | 2 | 未着手 |
| 4b | [Agent の構成](04b-agent-composition.md) | 4a | 未着手 |
| 5a | [Runtime の抽象化](05a-runtime-abstraction.md) | 4b | 未着手 |
| 5b | [Agent Engine アダプタ](05b-agent-engine-adapter.md) | 5a | 未着手 |
| 6a | [公開 Agent Card](06a-agent-card.md) | 4b | 未着手 |
| 6b | [Extended Card と A2A エンドポイント](06b-a2a-endpoint.md) | 3b, 5b, 6a | 未着手 |
| 7a | [Agent コンソール](07a-agent-console.md) | 2, 4b | 未着手 |
| 7b | [委譲の承認画面](07b-delegation-approval-ui.md) | 3a | 未着手 |
| 8a | [個人組織](08a-personal-organizations.md) | 1 | 未着手 |
| 8b | [Agent カタログ](08b-agent-catalog.md) | 6b, 8a | 未着手 |

3a / 4a は互いに独立で、2 が入れば並行して進められる。
5b はこのリポジトリに初めてクラウドプロバイダを持ち込むので、
それまでの全タスクはクラウド非依存のまま完結させる。

## どのタスクにも共通する落とし穴

新しいサービスや変更系 rpc を足すとき、忘れると黙って壊れる 3 か所。
いずれもテストが拾うが、先に知っておく方が早い。

- `server/pkg/apicatalog/catalog.go` の blank import —
  忘れると RBAC ルール・CLI コマンド・web operations・agent reference が黙って消える。
  `TestDefaultCoversEveryService` が落ちる。
- `server/internal/interface/grpc/interceptor/audit.go` の `audited` マップ —
  変更系 rpc はここか `notAudited`（理由付き）のどちらかに分類する。
  `TestEveryMutationIsClassified` が落ちる。
- `server/internal/infrastructure/persistence/yaml/menu/menus.yaml` —
  画面を足すときのサイドバー項目。サイドバーはルートではなく `resources` 行から作られる。

加えて:

- **認可判定が読む新しいテーブルには `rbac_revisions` トリガが要る**（`server/AGENTS.md` §11）。
- **enum のメンバにコメントを書かない。** protoc-gen-openapiv2 がそれらを 1 つの
  description にまとめ、継続行のインデントが足りないブロックスカラーを出力するため、
  数ステップ後の `openapi-typescript` で初めて落ちる。説明は enum 本体に書く。
- **`SELECT *` を位置指定でスキャンしている箇所**（`internal/usecase/system/{group,role,permission,resource}.go`）は、
  列を足すとスキャン側も直す必要がある。
