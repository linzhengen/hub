# 6a. 公開 Agent Card

## 目的

登録された Agent を A2A 仕様の Agent Card として公開する。

## やること

- `GET /.well-known/agent-card.json` 相当を Agent ごとに出す。
- `skills` は 4a で登録した Skill から組み立てる。手で二重に持たない。
- `securitySchemes` に Keycloak の OIDC discovery URL を宣言し、
  `capabilities.extendedAgentCard` を `true` にする（中身は 6b）。
- `supportedInterfaces` の URL は 5a の `Runtime.Endpoint` から取る。
- 公開カードには**認証なしで読める情報しか載せない**。
  組織名や内部の識別子を漏らさない。

## 完了条件

- A2A クライアントがカードを取得し、スキルを列挙できる。
- カードの内容が Skill 登録簿と一致し、手書きの重複が無い。
