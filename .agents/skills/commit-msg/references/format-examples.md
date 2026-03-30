# コミットメッセージの形式と具体例 (hub)

`hub` プロジェクトでは、Conventional Commits 規約に基づいた英語のコミットメッセージを使用します。

## 基本フォーマット

```text
<type>(<scope>): <subject>
```

- **type**: 変更の種類（必須）
- **scope**: 変更の影響範囲（任意だが推奨）
- **subject**: 変更内容の簡潔な説明（必須）

---

## タイプの選択ガイド

| Type | 説明 | 具体例 |
| :--- | :--- | :--- |
| **feat** | 新機能の追加 | `feat(server): add user authentication` |
| **fix** | バグ修正 | `fix(ui): resolve button alignment issue` |
| **docs** | ドキュメントの変更 | `docs: update setup instructions` |
| **style** | フォーマット等のコードの意味に影響しない変更 | `style(ui): fix linting errors` |
| **refactor** | リファクタリング（機能変更・バグ修正なし） | `refactor(server): simplify data fetching` |
| **perf** | パフォーマンス改善 | `perf(db): optimize user search query` |
| **test** | テストの追加・修正 | `test(server): add integration tests for auth` |
| **build** | ビルドシステム、依存関係の変更 | `build: upgrade go version to 1.26` |
| **ci** | CIの設定変更 | `ci: add github actions for linting` |
| **chore** | その他、ソースコード以外の変更 | `chore: update .gitignore` |

---

## スコープ (Scope) の選択ガイド

`hub` のディレクトリ構造に基づいて選択してください。

- **server**: `server/` 配下の Go バックエンドコード
- **ui**: `ui/` 配下のフロントエンドコード
- **infra**: `infra/` 配下の Terraform, Kubernetes マニフェスト
- **proto**: `proto/` 配下の Protocol Buffers 定義
- **db**: `db/` 配下のマイグレーションファイル、SQL定義
- **pkg**: `pkg/` 配下の共通ライブラリ

---

## 良い具体例 (Good Examples)

### Server 関連
- `feat(server): implement password hashing with argon2`
- `fix(server): handle null pointer in user repository`
- `refactor(server): extract jwt logic to pkg/auth`

### UI 関連
- `feat(ui): add dashboard page for metrics`
- `fix(ui): correct mobile responsive navigation`
- `style(ui): apply new color theme to sidebar`

### Infrastructure / CI 関連
- `infra: update kubernetes resource limits`
- `ci: add sonarcloud analysis to pull requests`
- `build: add air for hot reloading in server`

### Database / Proto 関連
- `db: add groups table for rbac`
- `proto: define UserGroup message and service`

---

## 悪い例と改善案

- ❌ `update files`
  - ✅ `feat(server): add user list api`
- ❌ `fix bug`
  - ✅ `fix(ui): prevent double submission on login`
- ❌ `WIP: refactor`
  - ✅ `refactor(server): split large handler into smaller functions`
- ❌ `Added new feature` (祈使形でない、首文字が大文字)
  - ✅ `feat(server): add search functionality`

---

## 生成時のヒント

- **複数の変更がある場合**: 最も重要な変更を type/scope に選び、subject で全体を簡潔にまとめます。
  - 例: server のバグ修正とリファクタリングを同時にコミットする場合
  - `fix(server): correct auth check and refactor interceptor`
- **迷った場合**: `chore` または scope なしの `fix`/`feat` を検討してください。
- **最近の傾向を確認**: `git log --oneline -10` を実行して、他の開発者がどのような scope を使っているか参考にしてください。
