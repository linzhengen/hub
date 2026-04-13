## タイプ・ラベルの選択ガイド

PR本文のタイプ選択（チェックボックス等）では、変更の主目的に最も近いものを選択してください。

### 選択例:

- **feat**: 新機能の追加、公開APIの追加
  - 例: `feat(server): add user rbac support`
- **fix**: バグ修正、不具合の改修
  - 例: `fix(ui): modal background overlay issue`
- **docs**: ドキュメントの修正、READMEの更新、コメントの追加
  - 例: `docs: update deployment guide`
- **refactor**: リファクタリング（機能変更を伴わないコード整理）
  - 例: `refactor(server): split di and add error package`
- **perf**: パフォーマンス改善
  - 例: `perf(server): optimize database query for user list`
- **test**: テストコードの追加・修正
  - 例: `test(server): add unit tests for rbac usecase`
- **infra / ci / chore**: インフラ、CI/CD、ビルド設定、依存関係の更新など
  - 例: `infra: update kubernetes manifests for dev`
  - 例: `ci: add golangci-lint workflow`
  - 例: `chore: upgrade dependencies`

## Background and Solution の書き方

問題の背景、解決策、および外部への影響（API変更、UI変更など）を2〜5行程度で記述します。

### 英語例 (English Example):
```markdown
### 💡 Background and Solution

The user RBAC system was missing a check for group-level permissions in gRPC interceptors.
This PR adds the necessary logic to validate permissions against the user's groups.
No breaking changes to the existing API.
```

### 日本語例 (Japanese Example):
```markdown
### 💡 変更の背景と解決策

gRPCインターセプターにおいて、ユーザーのグループ権限チェックが漏れていました。
このPRでは、ユーザーが所属するグループに基づいて権限を検証するロジックを追加しました。
既存のAPIに対する破壊的変更はありません。
```

## Change Log の書き方

### 1) 更新ログが必要なケース
ユーザーや他の開発者が感知できる変更（APIの追加/変更、UIの変更、動作仕様の変更など）がある場合。

```markdown
### 📝 Change Log

- Add support for group-based RBAC in gRPC interceptors.
```

### 2) 更新ログが不要なケース
内部的なリファクタリング、ドキュメントのみ、CI設定のみ、テストのみなど、ユーザーに直接影響がない場合。

```markdown
### 📝 Change Log

N/A
```

または

```markdown
### 📝 Change Log

No changelog required.
```

## ベースブランチ判断のヒント

デフォルトの `main` 以外を検討すべきケース：
- `feature/*` ブランチで作業しており、その親となる別の `feature` ブランチへマージする場合。
- 緊急のホットフィックスで、特定のリリースブランチへマージする場合。

判断に迷う場合は `git reflog` や `git branch -vv` を確認してください。

## PR タイトルの例 (Conventional Commits)

`lin-hub` では Conventional Commits に従った英語タイトルを使用します。

- `feat(server): implement password hashing`
- `fix(ui): correct button alignment in login page`
- `docs: add setup instructions for local development`
- `refactor(infra): simplify terraform modules`
- `ci: update pre-commit hooks version`
- `chore: update go-zero to v1.7.0`
- `test(server): add integration test for user registration`

## ユーザーへの確認メッセージ例

```markdown
PRの準備ができました。以下の内容で作成してもよろしいでしょうか？

- **Base branch**: `main`
- **PR title**: `feat(server): add user rbac support`
- **PR type**: `feat`
- **Checklist**: `make test`, `make gen` の実行を想定

内容に問題がなければ、`gh pr create` を実行します。
修正が必要な箇所があれば教えてください。
```
