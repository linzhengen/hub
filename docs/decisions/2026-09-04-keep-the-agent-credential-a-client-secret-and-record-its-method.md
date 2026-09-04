---
status: accepted
date: 2026-09-04
decision-makers: 'linzhengen'
---

# Keep the agent credential a client secret and record its method

## 何が既定だったか

`service_accounts`（`000014`）をそのまま延長し、Agent にも Keycloak の
client secret を 1 回だけ返して終わり ——
`docs/tasks/02-agent-identity.md` はそう書いていた。

一方、機械の身元について今すすめられている作法はそこから離れている。
Google Cloud の Agent Identity は Agent を service account とは別の一級の
principal として扱い、**共有しない・なりすまし不可・長命な鍵を発行させない**を掲げ、
24 時間で自動更新される X.509 にアクセストークンを暗号的にバインドする。
A2A の企業向け指針も「Agent Card に平文の秘密を置かない」「資格情報は帯域外で得る」を
求めている。hub が発行しようとしているのは、まさにその長命な共有シークレットである。

## 何をどう決めたか

**当面は client secret のまま出す。ただし認証方式を `agents` の列にする。**

SPIFFE も mTLS も hub には無く、持ち込めば `service_accounts` が引いた
「認証経路を 2 本にしない」という線を最初に越えることになる。
Keycloak 26.5.6 は `client-jwt`（private_key_jwt）を既に持っているので、
**共有シークレットをやめる道は、認証経路を増やさずに同じ Keycloak の中にある。**
足りないのは移行の受け皿だけである。

そこで `agents.auth_method`（`CLIENT_SECRET` / `PRIVATE_KEY_JWT`、CHECK 制約付き）を
先に作る。列があれば 2 つの方式は共存でき、Agent を 1 体ずつ移せる。
無ければ移行はフラグデーになり、フラグデーは実行されない。

あわせて `agents.secret_rotated_at` を持つ。hub はシークレットを保存しないので、
**いつ発行されたかが、その古さについて hub が言える唯一のこと**である。
見えない古さは誰も直さない。

`last_authenticated_at` は**作らない**。認証済みリクエストのたびに書き込みが要り、
「この Agent が最後に何をしたか」は `audit_logs` が actor として既に答えている。

## 何が変わるか

- `server/db/migrations/postgres/000015_create_table_agents.up.sql` の
  `auth_method` 列・CHECK 制約・`secret_rotated_at` 列
- `server/internal/domain/ai/agent/agent.go` の `AuthMethod` 型。
  `AuthMethodPrivateKeyJWT` は**今は発行されない値**として存在する
- `proto/ai/agent/v1/model.proto` の `AuthMethod` enum。
  CLI と web と Agent リファレンスには `make gen` 経由で届く
- `RotateAgentSecret` は Keycloak で発行し直したあと `secret_rotated_at` を打つ。
  打刻に失敗しても回転自体は成功として返す —— 古い資格情報は既に死んでいる

## いつ見直すか

次のどれかが起きたら `PRIVATE_KEY_JWT` の実装に進む。

- Agent が hub の外（組織をまたぐ A2A）から呼ばれるようになったとき。
  帯域外に置かれた共有シークレットの数だけ、漏洩の面が増える
- Agent の数が「秘密を配って回れる」規模を超えたとき
- `docs/tasks/05b-agent-engine-adapter.md` でクラウド上に Agent を置くとき。
  実行基盤に秘密を渡すのは、鍵を渡すより明確に悪い
