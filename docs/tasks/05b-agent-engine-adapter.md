# 5b. Agent Engine アダプタ

## 目的

Vertex AI Agent Engine を `Runtime` の実装として足す。

## なぜ / 注意

**3b（代理チェーンつきの認可）が先に要る。**デプロイした Agent は人に呼ばれる。
3b が無いと、Agent は自分の付与だけで動き、監査ログには Agent しか残らない ——
呼んだ人は自分が持っていない権限を Agent 越しに行使でき、痕跡も無い。
`agents` の登録簿だけでは代理は成立しない（`02-agent-identity.md`
「この時点で成立していないこと」）。

**このタスクはこのリポジトリに初めてクラウドプロバイダを持ち込む。**
現状 `go.mod` に GCP SDK は 1 つも無く（`cloud.google.com/*` は
すべて `/go.mod` のみのプルーニング済みハッシュ）、`infra/tf` にも
プロバイダが 1 つも無い（`infra/tf/AGENTS.md` が明言、backend は `local {}`）。

したがって Terraform state をローカルに置き続けられなくなる。
**その判断を先に済ませること。**コードより先に決める論点である。

## やること

- 判断: Terraform backend をどこに置くか。GCP を必須にするのか、
  任意の追加インフラとして切り離せるか。`infra/tf/AGENTS.md` を更新する。
- `internal/infrastructure/ai/agent/runtime/agentengine/`。
  Agent Engine の Session / Memory Bank を使うかどうかもここで決める
  （hub 側に既に chat セッションがあるので、二重に持たない）。
- ADK は Go を主とし、Python も受け入れる。ユーザー持ち込みの Agent は
  A2A の Agent Card と MCP で疎結合にする。
- 資格情報は環境変数で、hub のリポジトリには置かない。

## 完了条件

- Agent Engine にデプロイした Agent の A2A エンドポイントが `Endpoint` から返る。
- Agent Engine を使わない構成（ローカル runtime のみ）でも
  `make dev` が今までどおり動く。
