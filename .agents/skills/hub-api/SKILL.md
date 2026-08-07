---
name: hub-api
description: hub の API を CLI (`hub` コマンド) 経由で呼び出します。ユーザー・グループ・ロール・権限・リソース (メニュー) の照会や作成・更新・削除、RBAC の権限調査、「誰がこの API を呼べるか」の確認、API エンドポイントやパーミッションの一覧取得を求められた際に使用します。「hub のユーザーを一覧して」「グループにロールを付与して」「この API に必要な権限は」「hub api」「hubctl」などが該当します。
---

# hub API を CLI から呼ぶ

## 目的

`hub` CLI は hub の REST ゲートウェイのクライアントです。コマンドツリーは
protobuf 定義から生成されるため、rpc が増えればコマンドも自動的に増えます。
**エンドポイントを推測せず、必ずカタログを引いてください。**

## 前提: 接続とトークン

CLI は 3 層で設定を解決します（後のものが勝ちます）。

1. `~/.hub/config.yaml` のプロファイル
2. 環境変数 `HUB_ENDPOINT` / `HUB_TOKEN` / `HUB_OIDC_ISSUER` / `HUB_OIDC_CLIENT_ID` / `HUB_OIDC_CLIENT_SECRET`
3. `--endpoint` / `--token` などのフラグ

```bash
# 現在の解決結果を確認（トークンは伏字）
hub config show

# 未設定ならプロファイルを作る
hub config set --endpoint http://localhost:9090 \
  --oidc-issuer http://localhost:8080/realms/hub --oidc-client-id hub

# トークンを取得（--username 無しならクライアントクレデンシャル）
hub auth login --username admin        # パスワードは対話入力
HUB_PASSWORD=... hub auth login -u admin   # 非対話環境
hub auth whoami
```

**`--password` はシェル履歴とプロセス一覧に残ります。** 端末があれば省略して
対話入力、無ければ `HUB_PASSWORD` を使ってください。

アクセストークンは有効期限が来ると**自動で更新されます**（`hub auth login` が
リフレッシュトークンをプロファイルに保存するため）。それでも 401 が続く場合は
リフレッシュトークンも切れているので、再度 `hub auth login` してください。
403 の場合は認証は通っていて **権限が足りません** — 下の「RBAC」を読んでください。

## 手順

### 1. カタログを引く

```bash
hub api list                       # 全 API を JSON で（service / path / rbac / command）
hub api list --service UserService # サービスで絞り込み
hub api describe ListUser          # 1 つの rpc の全リクエストフィールド
```

`hub api describe` は各フィールドについて `flag`（CLI のフラグ名）、`name`
（JSON 名）、`in`（`path` / `query` / `body`）、`kind`、`enumValues`、
`constraints` を返します。

enum は **必ず** `enumValues` にある文字列を渡してください（`STATUS_ACTIVE`
のような完全な名前で、`ACTIVE` ではありません）。

`constraints` はサーバーが実際に弾く条件です。**送る前にここを読んでください。**

```
--username   ["length 1..64"]
--email      ["email"]
--password   ["length 8..128"]
--group-ids  ["each uuid"]
```

主な規則:

- **id は必ず UUID** です。`--id abc` のような値は `InvalidArgument` で弾かれます
- 一覧の `--limit` は **200 が上限**です。超えると弾かれるので、全件は `--all` を使ってください
- 追加・削除系（`add-*` / `remove-*`）は **1件以上**の id が必要です

違反すると gRPC の `InvalidArgument` が返り、**どのフィールドがなぜ駄目か**が
メッセージに入ります。

### 2. 呼ぶ

コマンド名は rpc 名の kebab-case です。`ListUser` → `hub user list-user`。
素の CRUD には短縮形があります（`list` / `get` / `create` / `update` / `delete`）。

```bash
hub user list --limit 20 --status STATUS_ACTIVE
hub user get --id 6f1c...
hub group create --name platform --description "Platform team"
hub group add-users-to-group --group-id g1 --user-ids u1 --user-ids u2
hub role add-permissions-to-role --role-id r1 --permission-ids p1 --permission-ids p2
```

繰り返しフィールドは**フラグを繰り返します**（`--user-ids a --user-ids b`）。
カンマ区切りの 1 個の値としては解釈されません。

message / map 型のフィールドは JSON 文字列で渡します。

```bash
hub resource create --name "Users API" --type TYPE_API \
  --identifier '{"api":"api.user.v1.UserService","category":"user"}' \
  --metadata '{"icon":"users"}'
```

### 3. 確認してから壊す

**削除・権限変更などの破壊的な操作の前には必ず `--dry-run` で送信内容を確認し、
ユーザーに提示して承認を得てください。**

```bash
hub user delete --id u1 --dry-run
# => {"method":"DELETE","url":"http://localhost:9090/api/v1/users/u1"}
```

## 出力

既定は整形済み JSON です。パースするならそのまま、目視するなら切り替えます。

```bash
hub user list -o table   # 一覧レスポンスを表に
hub user list -o yaml
hub user list | jq '.users[] | {id, email}'
```

### 一覧は既定で 50 件までです

API は limit 未指定なら 50 件、最大 200 件しか返しません。全件必要なら
`--all` を付けてください。CLI が最後のページまで追って1つのレスポンスに
まとめます。

```bash
hub user list --all | jq '.users | length'
hub permission list --all
```

`--all` はページングを持つ一覧操作にのみ使えます。持たない操作に付けると
エラーになります。

`-o table` はレスポンス内の唯一の配列フィールドを表にします。表にできない形
（単一オブジェクトなど）は JSON にフォールバックします。

## 生成コマンドで足りないとき

`hub api call` が逃げ道です。パスは `/api/v1` を含めて丸ごと渡します。

```bash
hub api call GET /api/v1/users --query limit=10 --query userIds=a --query userIds=b
hub api call POST /api/v1/groups --data '{"name":"platform"}'
hub api call PUT /api/v1/users/u1 --data @user.json
```

## RBAC

権限は「リソース識別子 × 動詞」です。既定ではリソースが
`api.<protoパッケージ>.<Service>`、動詞が rpc 名になります
（`hub api describe` の `resource` / `action`）。

- ユーザーは複数のグループに属し、グループはロールを持ち、ロールは権限を持ちます。
- パターンには `*` が使えます（`api.*`、`api.system.*.v1.*Service` など）。
- `public: true` の rpc（現状 `GetMe` のみ）は認証さえ通れば誰でも呼べます。

403 を受けたときの調べ方:

```bash
hub api describe DeleteUser        # 必要な resource / action を確認
hub auth whoami                    # 自分の所属グループを確認
hub group get --id <groupId>       # そのグループのロール
hub role get --id <roleId>         # そのロールの権限
hub permission list                # 権限と verb の一覧
```

権限を足す場合は、`hub permission create` →
`hub role add-permissions-to-role` → `hub group assign-roles-to-group` の順です。

## 詳細リファレンス

全 rpc のフラグ・エンドポイント・必要権限は
[references/api-reference.md](references/api-reference.md) にあります。
`hub api docs` で生成され `make gen` で更新されるため、proto と乖離しません。

## やってはいけないこと

- URL やフラグ名を推測しない。`hub api describe` を引く。
- enum に短縮名を渡さない（`STATUS_ACTIVE` であって `ACTIVE` ではない）。
- 繰り返しフィールドをカンマ区切りで渡さない。
- トークンやクライアントシークレットをログ・PR・コミットに書かない。
  `hub config show` は既定で伏字にします。
