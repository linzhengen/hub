# CLI guide

`hub` is the command line client for the hub API. Its command tree is generated
from the same Protobuf definitions the server routes on, so **every rpc is a
command** — add an rpc, run `make gen`, and the command exists. There is no
table of endpoints kept in step by hand.

```text
hub <service> <method> [flags]     e.g. hub user list-user --limit 10
```

## Install

```bash
make cli          # go install ./server/cmd/hub
hub version
```

The CLI is a plain HTTP client: it needs an endpoint and a token, but no
database and no access to the server's source.

## Configure

Settings resolve in three layers, each overriding the one before it:

1. the profile in `~/.hub/config.yaml`
2. the environment: `HUB_ENDPOINT`, `HUB_TOKEN`, `HUB_OIDC_ISSUER`, `HUB_OIDC_CLIENT_ID`, `HUB_OIDC_CLIENT_SECRET`, `HUB_PASSWORD`
3. the flags: `--endpoint`, `--token`, `--profile`

```bash
hub config set --endpoint http://localhost:9090 \
  --oidc-issuer http://localhost:8080/realms/hub \
  --oidc-client-id hub-web

hub config show     # what the CLI actually resolved; the token is masked
hub config path     # where the file lives
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

Use `--profile <name>` to keep more than one environment in the same file:

```bash
hub config set --profile staging --endpoint https://hub.example.com
hub --profile staging user list
```

## Sign in

`hub auth login` obtains a token from Keycloak and stores it — with its refresh
token — in the profile, so an expired access token is renewed automatically.

### Browser (device flow, RFC 8628)

The default when no username and no client secret are given. No client secret is
required, which makes it the right choice for a human at a terminal.

```bash
HUB_OIDC_ISSUER=http://localhost:8080/realms/hub \
HUB_OIDC_CLIENT_ID=hub-web \
hub auth login
```

A short code is printed, the browser opens, you sign in, and the CLI reports
`✓ Authentication successful`. Pass `--web` to force this flow even when a client
secret is configured.

### Service account (client credentials)

The unattended path — CI jobs, agents, cron:

```bash
HUB_OIDC_ISSUER=http://localhost:8080/realms/hub \
HUB_OIDC_CLIENT_ID=hub-api \
HUB_OIDC_CLIENT_SECRET=… \
hub auth login
```

With a client secret and no `--username`, the client credentials grant is used.

### Password grant

```bash
hub auth login --username admin        # prompts for the password
HUB_PASSWORD=… hub auth login -u admin # non-interactive
```

> `--password` is visible in your shell history and in the process list. Prefer
> the prompt, or `HUB_PASSWORD`.

### Check

```bash
hub auth whoami   # the caller's profile and groups
hub auth token    # the raw access token, e.g. to paste into the API explorer
```

Add `--save=false` to `hub auth login` to print a token without touching the
profile.

## Explore the API

Do not guess endpoints or flag names — ask the catalog.

```bash
hub api list                          # every operation as JSON
hub api list --service UserService    # narrowed to one service
hub api describe ListUser             # one rpc and all of its request fields
```

`hub api list` names, for each operation, the REST mapping, the permission it
needs and the exact command that calls it:

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

`hub api describe` adds every field with its flag, where it goes (`path`,
`query` or `body`), its type, its enum values and the constraints the server
enforces:

```json
{
  "name": "limit",
  "flag": "--limit",
  "in": "query",
  "kind": "uint32",
  "constraints": ["<= 200"]
}
```

Read `constraints` before sending: they are what the server actually rejects.
The same information is on the [API reference](api-reference.html) page and, in
a browsable form, in the [API explorer](api.html).

## Call an operation

The command name is the rpc name in kebab-case — `ListUser` becomes
`hub user list-user` — and plain CRUD rpcs also get a short alias (`list`,
`get`, `create`, `update`, `delete`).

```bash
hub user list --limit 20 --status STATUS_ACTIVE
hub user get --id 6f1cf7e2-0a1d-4c0e-9d9f-a1b2c3d4e5f6
hub group create --name platform --description "Platform team"
hub group add-users-to-group --group-id <uuid> --user-ids <uuid> --user-ids <uuid>
hub role add-permissions-to-role --role-id <uuid> --permission-ids <uuid>
```

`--help` on any generated command prints the rpc, the endpoint and the RBAC rule
above the flags, each flag annotated with where it goes and what the server
accepts:

```text
$ hub group create --help
Create a group.

rpc:      system.group.v1.GroupService.CreateGroup
endpoint: POST /api/v1/groups
rbac:     CreateGroup on api.system.group.v1.GroupService
```

Rules worth knowing before the server tells you off:

- **Ids are UUIDs.** `--id abc` comes back as `InvalidArgument`.
- **Enums take their full name.** `STATUS_ACTIVE`, not `ACTIVE`.
- **Repeated fields repeat the flag.** `--user-ids a --user-ids b`, never `a,b`.
- **`--limit` caps at 200**, and defaults to 50. Use `--all` for everything.
- **Message and map fields take JSON**:

```bash
hub resource create --name "Users API" --type TYPE_API \
  --identifier '{"api":"api.user.v1.UserService","category":"user"}' \
  --metadata '{"icon":"users"}'
```

### Check before you break something

`--dry-run` prints the request instead of sending it. Use it before anything
destructive:

```bash
hub user delete --id 6f1cf7e2-0a1d-4c0e-9d9f-a1b2c3d4e5f6 --dry-run
```

```json
{
  "method": "DELETE",
  "url": "http://localhost:9090/api/v1/users/6f1cf7e2-0a1d-4c0e-9d9f-a1b2c3d4e5f6"
}
```

## Output

The default is pretty-printed JSON — pipe it straight into `jq`.

```bash
hub user list -o table          # the list field as a table
hub user list -o yaml
hub user list | jq '.users[] | {id, email}'
```

`-o table` renders the single array field in the response; anything that is not
a list falls back to JSON.

### Paging

List endpoints return 50 rows by default and 200 at most. `--all` follows the
pages and merges them into one response:

```bash
hub user list --all | jq '.users | length'
hub permission list --all
```

`--all` only applies to operations that page; elsewhere it is an error.

## The escape hatch

Anything the generated commands do not cover goes through `hub api call`. The
path is taken verbatim, so it includes the `/api/v1` prefix.

```bash
hub api call GET /api/v1/users --query limit=10 --query status=STATUS_ACTIVE
hub api call POST /api/v1/groups --data '{"name":"platform"}'
hub api call PUT /api/v1/users/<uuid> --data @user.json
```

## Permissions (RBAC)

A permission is a **verb on a resource**: by default the resource is
`api.<proto package>.<Service>` and the verb is the rpc name — the `resource`
and `action` fields of `hub api describe`. Users belong to groups, groups hold
roles, roles hold permissions, and patterns may use `*` (`api.*`,
`api.system.*.v1.*Service`). An rpc marked `public: true` — today only `GetMe` —
needs authentication but no permission.

A `401` means the token is gone or expired: run `hub auth login` again. A `403`
means you are authenticated but not allowed, and this is how you find out why:

```bash
hub api describe DeleteUser   # which resource and action are required
hub auth whoami               # which groups you are in
hub group get --id <groupId>  # that group's roles
hub role get --id <roleId>    # that role's permissions
hub permission list           # every permission and its verb
```

To grant the missing one — permission, then role, then group:

```bash
hub permission create --resource-id <uuid> --verb DeleteUser
hub role add-permissions-to-role --role-id <uuid> --permission-ids <uuid>
hub group add-roles-to-group --id <groupId> --role-ids <uuid>
```

## Shell completion

```bash
hub completion bash > /etc/bash_completion.d/hub
hub completion zsh  > "${fpath[1]}/_hub"
hub completion fish > ~/.config/fish/completions/hub.fish
```

## For AI agents

The repository ships an agent skill at `.agents/skills/hub-api` that documents
this surface for coding agents, including the generated
[API reference](api-reference.html) with every flag, endpoint and RBAC rule. Two
rules matter most: look operations up with `hub api list` / `hub api describe`
instead of guessing, and always `--dry-run` a destructive call and get it
approved before sending it. Never paste a token or a client secret into a
commit, a PR or a log — `hub config show` masks it for a reason.
