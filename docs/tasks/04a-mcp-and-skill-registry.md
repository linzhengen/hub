# 4a. MCP サーバーと Skill の登録簿

## 目的

Agent に持たせられる道具を、組織ごとに登録・管理できるようにする。

## なぜ

「何を持っているか」と「何をしてよいか」は別のゲートである（方向性メモ §4.5）。
これはその前者。`toolbox.go` の `exposed` と同じ思想 ——
増えた道具が黙って Agent の手に渡らないよう、明示的な登録を要求する。

## やること

- マイグレーション `000017_create_table_mcp_servers` と
  `000018_create_table_agent_skills`。どちらも `org_id` を持つ。
- MCP サーバー: `name`, `transport`（stdio / streamable-http）, `endpoint`,
  `auth_kind`, 資格情報の参照。**シークレットは hub に保存しない** ——
  `service_accounts` が secret を保存しないのと同じ理由。
- Skill: A2A の `AgentSkill` に寄せる（`id`, `name`, `description`, `tags`,
  `input_modes`, `output_modes`, `examples`）。6a でそのまま Agent Card に載る。
- proto `ai/mcpserver/v1`, `ai/skill/v1`。
- どちらも認可判定は読まないので `rbac_revisions` トリガは不要。

## 完了条件

- 組織ごとに MCP サーバーと Skill を登録・一覧・削除できる。
- 他組織のものが一覧に混ざらない。
- 登録した MCP サーバーへ疎通確認（tools/list）ができる。
