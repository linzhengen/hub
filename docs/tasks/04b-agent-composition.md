# 4b. Agent の構成

## 目的

Agent に「どの MCP サーバー・どの Skill・どのサブ Agent を持つか」を紐付ける。

## なぜ

ここが二重ゲートの前半で、RBAC + 委譲チェーンが後半。両方を通らないと実行されない。
これは既存の `ToolBox.Tools()` / `Call()` の二重チェック
（一覧は提案を防ぎ、`Call` の中の検査が実際に効く）の Agent 版である。

この分離があるので、**Agent ごとの RBAC インスタンススコープは要らない**。
組織スコープと構成の許可リストで足りる。

## やること

- 結合テーブル: `agent_mcp_servers`, `agent_skills`, および
  `agents.parent_agent_id` を使ったサブ Agent の紐付け。
- rpc は §17 の規約どおり `Add<Children>To<Parent>` /
  `Remove<Children>From<Parent>` の 2 本ずつ。「まるごと置換」は作らない。
- 構成にバージョンを持たせる（`agent_versions`）。5a のデプロイ単位になる。
- サブ Agent の循環を拒否する。親子は木であって、グラフではない。

## 完了条件

- Agent に道具を付け外しでき、構成がバージョンとして固まる。
- 他組織の MCP サーバー / Skill を紐付けられない。
- 循環するサブ Agent 構成が作れない。
