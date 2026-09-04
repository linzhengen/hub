# 5a. Runtime の抽象化

## 目的

デプロイ先を差し替え可能にしたうえで、hub 側のデプロイ管理を作る。

## なぜ

hub 本体を GCP に依存させないため。実装が 1 つしかない段階で
インターフェースを切っておかないと、Agent Engine の都合がドメインに漏れる。

## やること

- `internal/domain/ai/agent/runtime.go`:

  ```go
  type Runtime interface {
      Deploy(ctx context.Context, v *Version) (*Deployment, error)
      Status(ctx context.Context, d *Deployment) (Status, error)
      Endpoint(ctx context.Context, d *Deployment) (string, error)
      Delete(ctx context.Context, d *Deployment) error
  }
  ```

- マイグレーション `000019_create_table_agent_deployments`。
  `agent_version_id`, `runtime`, `external_id`, `endpoint`, `status`。
- `DeployAgent` / `GetDeployment` / `DeleteDeployment` の rpc。
- **このタスクではクラウドを持ち込まない。** 最初の実装は
  「何もしないローカル runtime」でよく、それでデプロイ管理そのものをテストできる。

## 完了条件

- 構成のバージョンをデプロイ・状態確認・削除でき、監査ログに残る。
- `go.mod` にクラウド SDK が 1 つも増えていない。
