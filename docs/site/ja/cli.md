# CLI ガイド

`hub` は hub API のコマンドラインクライアントです。コマンドツリーはサーバーが
ルーティングに使うのと同じ Protobuf 定義から生成されるため、**すべての rpc が
コマンドになります**。rpc を追加して `make gen` を実行すれば、そのコマンドは
存在します。エンドポイントの一覧を手で維持する必要はありません。

```text
hub <service> <method> [flags]     例: hub user list-user --limit 10
```

## インストール

```bash
make cli          # go install ./server/cmd/hub
hub version
```

CLI は素の HTTP クライアントです。エンドポイントとトークンさえあれば動き、
データベースもサーバーのソースも必要ありません。

## 設定

設定は 3 層で解決され、後のものが前を上書きします。

1. `~/.hub/config.yaml` のプロファイル
2. 環境変数 `HUB_ENDPOINT` / `HUB_TOKEN` / `HUB_OIDC_ISSUER` / `HUB_OIDC_CLIENT_ID` / `HUB_OIDC_CLIENT_SECRET` / `HUB_PASSWORD`
3. フラグ `--endpoint` / `--token` / `--profile`

```bash
hub config set --endpoint http://localhost:9090 \
  --oidc-issuer http://localhost:8080/realms/hub \
  --oidc-client-id hub-web

hub config show     # 実際に解決された設定。トークンは伏字
hub config path     # 設定ファイルの場所
```

```json
{
  "endpoint": "http://localhost:9090",
  "oidcClientId": "hub-web",
  "oidcClientSecret": "",
  "oidcIssuer": "http://localhost:8080/realms/hub",
  "token": "***"
}
```

`--profile <name>` で、同じファイルに複数の環境を持てます。

```bash
hub config set --profile staging --endpoint https://hub.example.com
hub --profile staging user list
```

## ログイン

`hub auth login` は Keycloak からトークンを取得し、リフレッシュトークンと共に
プロファイルへ保存します。アクセストークンの期限切れは自動で更新されます。

### ブラウザ (デバイスフロー, RFC 8628)

ユーザー名もクライアントシークレットも指定しない場合の既定の動作です。
クライアントシークレットが不要なので、端末の前に人がいる場合はこれを使います。

```bash
HUB_OIDC_ISSUER=http://localhost:8080/realms/hub \
HUB_OIDC_CLIENT_ID=hub-web \
hub auth login
```

ワンタイムコードが表示され、ブラウザが自動で開き、サインインすると CLI が
`✓ Authentication successful` を表示します。クライアントシークレットが設定
されていてもこのフローを使いたい場合は `--web` を付けます。

### サービスアカウント (クライアントクレデンシャル)

CI・エージェント・cron など、無人実行のための経路です。

```bash
HUB_OIDC_ISSUER=http://localhost:8080/realms/hub \
HUB_OIDC_CLIENT_ID=hub-api \
HUB_OIDC_CLIENT_SECRET=… \
hub auth login
```

クライアントシークレットがあり `--username` が無いとき、クライアント
クレデンシャルグラントが使われます。

### パスワードグラント

```bash
hub auth login --username admin        # パスワードは対話入力
HUB_PASSWORD=… hub auth login -u admin # 非対話環境
```

> `--password` はシェル履歴とプロセス一覧に残ります。対話入力か
> `HUB_PASSWORD` を使ってください。

### 確認

```bash
hub auth whoami   # 呼び出し元のプロフィールと所属グループ
hub auth token    # 生のアクセストークン。API エクスプローラーへの貼り付けなどに
```

プロファイルを書き換えずにトークンだけ表示したい場合は
`hub auth login --save=false` を使います。

## API を調べる

エンドポイントやフラグ名を推測せず、カタログに聞いてください。

```bash
hub api list                          # 全操作を JSON で
hub api list --service UserService    # サービスで絞り込み
hub api describe ListUser             # 1 つの rpc とその全リクエストフィールド
```

`hub api list` は操作ごとに、REST のマッピング、必要な権限、それを呼ぶ
コマンドそのものを返します。

```json
{
  "service": "user.v1.UserService",
  "method": "ListUser",
  "command": "hub user list-user",
  "httpMethod": "GET",
  "path": "/api/v1/users",
  "summary": "List users, optionally filtered by id, email, name, status or group.",
  "public": false,
  "resource": "api.user.v1.UserService",
  "action": "ListUser"
}
```

`hub api describe` はさらに各フィールドについて、フラグ名、送信先
(`path` / `query` / `body`)、型、enum の値、サーバーが強制する制約を返します。

```json
{
  "name": "limit",
  "flag": "--limit",
  "in": "query",
  "kind": "uint32",
  "constraints": ["<= 200"]
}
```

`constraints` はサーバーが実際に弾く条件です。**送る前に読んでください。**
同じ情報は [API リファレンス](../api-reference.html)に、ブラウズしやすい形では
[API エクスプローラー](../api.html)にあります。

## 操作を実行する

コマンド名は rpc 名の kebab-case です (`ListUser` → `hub user list-user`)。
素の CRUD には短縮形 (`list` / `get` / `create` / `update` / `delete`) も
用意されています。

```bash
hub user list --limit 20 --status STATUS_ACTIVE
hub user get --id 6f1cf7e2-0a1d-4c0e-9d9f-a1b2c3d4e5f6
hub group create --name platform --description "Platform team"
hub group add-users-to-group --group-id <uuid> --user-ids <uuid> --user-ids <uuid>
hub role add-permissions-to-role --role-id <uuid> --permission-ids <uuid>
```

生成コマンドに `--help` を付けると、フラグの上に rpc・エンドポイント・RBAC
ルールが表示され、各フラグには送信先とサーバーが受け付ける条件が添えられます。

```text
$ hub group create --help
Create a group.

rpc:      system.group.v1.GroupService.CreateGroup
endpoint: POST /api/v1/groups
rbac:     CreateGroup on api.system.group.v1.GroupService
```

サーバーに叱られる前に知っておきたい規則です。

- **id は UUID。** `--id abc` は `InvalidArgument` で返ってきます。
- **enum は完全名。** `ACTIVE` ではなく `STATUS_ACTIVE`。
- **繰り返しフィールドはフラグを繰り返す。** `--user-ids a --user-ids b`。`a,b` は不可。
- **`--limit` の上限は 200**、未指定は 50 件。全件は `--all`。
- **message / map 型のフィールドは JSON 文字列**:

```bash
hub resource create --name "Users API" --type TYPE_API \
  --identifier '{"api":"api.user.v1.UserService","category":"user"}' \
  --metadata '{"icon":"users"}'
```

### 壊す前に確認する

`--dry-run` は送信せずにリクエスト内容を表示します。破壊的な操作の前には
必ず使ってください。

```bash
hub user delete --id 6f1cf7e2-0a1d-4c0e-9d9f-a1b2c3d4e5f6 --dry-run
```

```json
{
  "method": "DELETE",
  "url": "http://localhost:9090/api/v1/users/6f1cf7e2-0a1d-4c0e-9d9f-a1b2c3d4e5f6"
}
```

## 出力

既定は整形済み JSON です。そのまま `jq` に流せます。

```bash
hub user list -o table          # レスポンス中の配列を表に
hub user list -o yaml
hub user list | jq '.users[] | {id, email}'
```

`-o table` はレスポンス内の唯一の配列フィールドを表にします。表にできない形は
JSON にフォールバックします。

### ページング

一覧系は既定で 50 件、最大でも 200 件しか返しません。`--all` はページを追って
1 つのレスポンスにまとめます。

```bash
hub user list --all | jq '.users | length'
hub permission list --all
```

`--all` はページングを持つ操作専用で、それ以外に付けるとエラーになります。

## 逃げ道

生成コマンドで届かないものは `hub api call` で叩きます。パスはそのまま使われる
ため、`/api/v1` を含めて渡します。

```bash
hub api call GET /api/v1/users --query limit=10 --query status=STATUS_ACTIVE
hub api call POST /api/v1/groups --data '{"name":"platform"}'
hub api call PUT /api/v1/users/<uuid> --data @user.json
```

## 権限 (RBAC)

権限は「リソース × 動詞」です。既定ではリソースが
`api.<proto パッケージ>.<Service>`、動詞が rpc 名で、これは `hub api describe`
の `resource` と `action` にあたります。ユーザーはグループに属し、グループは
ロールを持ち、ロールは権限を持ちます。パターンには `*` が使えます
(`api.*`、`api.system.*.v1.*Service`)。`public: true` の rpc (現状 `GetMe`
のみ) は、認証さえ通っていれば権限は不要です。

`401` はトークンが無いか期限切れです。`hub auth login` をやり直してください。
`403` は認証は通っていて**権限が足りない**状態で、原因は次のように調べます。

```bash
hub api describe DeleteUser   # 必要な resource と action
hub auth whoami               # 自分の所属グループ
hub group get --id <groupId>  # そのグループのロール
hub role get --id <roleId>    # そのロールの権限
hub permission list           # 権限と動詞の一覧
```

足りない権限を付与する順序は、権限 → ロール → グループです。

```bash
hub permission create --resource-id <uuid> --verb DeleteUser
hub role add-permissions-to-role --role-id <uuid> --permission-ids <uuid>
hub group add-roles-to-group --id <groupId> --role-ids <uuid>
```

## シェル補完

```bash
hub completion bash > /etc/bash_completion.d/hub
hub completion zsh  > "${fpath[1]}/_hub"
hub completion fish > ~/.config/fish/completions/hub.fish
```

## AI エージェント向け

リポジトリには `.agents/skills/hub-api` にエージェントスキルが同梱されており、
全フラグ・エンドポイント・RBAC ルールを含む生成済みの
[API リファレンス](../api-reference.html)と併せて、この操作面を AI
エージェント向けに説明しています。特に重要なのは 2 点です。推測せず
`hub api list` / `hub api describe` で調べること。破壊的な呼び出しは必ず
`--dry-run` で内容を提示し、承認を得てから実行すること。トークンや
クライアントシークレットをコミット・PR・ログに残さないでください
(`hub config show` が伏字にするのはそのためです)。
