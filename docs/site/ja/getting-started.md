# はじめかた

クローンした直後の状態からスタックを起動し、API・Web アプリ・インフラを日々
変更していくループまでを説明します。

## 前提ツール

| ツール | 用途 |
|:---|:---|
| Docker + Docker Compose | PostgreSQL、Keycloak、MailHog、API サーバー |
| Go 1.26+ | バックエンド、コードジェネレーター、`hub` CLI |
| Node.js + pnpm | React の Web アプリと本ドキュメントサイト |
| Terraform | Keycloak のレルムとクライアントの作成 |
| `kubectl` + `minikube` | 任意: Kubernetes へのデプロイ |

## 1. ツールチェーンの導入

```bash
git clone https://github.com/linzhengen/hub.git
cd hub

# バックエンド側: buf プラグイン、sqlc、migrate、pre-commit
make init

# フロントエンド側
cd ui/web && pnpm install && cd ../..
```

## 2. ローカルスタックの起動

```bash
make dev
```

`make dev` は Docker Compose を起動し、Keycloak が healthy になるのを待ち、
ローカル用に master レルムの TLS 要求を緩め、`infra/tf/dev` の Terraform を
適用してレルムとクライアントを作成し、DB にシードを投入してからログを追跡します。

| サービス | URL | 備考 |
|:---|:---|:---|
| API (gRPC-Gateway) | <http://localhost:9090> | REST は `/api/v1` 配下 |
| Keycloak | <http://localhost:8080> | 管理コンソール、`admin` / `admin` |
| PostgreSQL | `localhost:5432` | データベース `hub` |
| MailHog | <http://localhost:8025> | 確認メールの受信箱 |
| Web UI | <http://localhost:3000> | 下記のとおり別途起動 |

Web アプリは Compose の外、Vite の開発サーバーで動かします。

```bash
cd ui/web && pnpm dev
```

## 3. シードデータ

`make dev` の中でも実行されますが、いつでも再実行できます。

```bash
make dev-seed
```

初期のユーザー・グループ・ロール・権限を作る `seed` と、UI が描画するメニュー
(リソース) ツリーを取り込む `resource-import` を、API コンテナ内で実行します。

## 4. API を叩いてみる

一通り繋がったかを確認する一番早い方法は CLI です。

```bash
make cli   # `hub` バイナリをインストール

HUB_ENDPOINT=http://localhost:9090 \
HUB_OIDC_ISSUER=http://localhost:8080/realms/hub \
HUB_OIDC_CLIENT_ID=hub-web \
hub auth login

hub auth whoami
hub user list -o table
```

プロファイル、他の認可フロー、コマンド体系の全体は
[CLI ガイド](cli.html)にまとめてあります。

## ディレクトリ構成

```text
.
├── server/                 # Go バックエンド
│   ├── cmd/                # エントリポイント: サーバー、hub CLI、ジェネレーター
│   ├── internal/           # domain / usecase / interface / infrastructure
│   ├── proto/              # API 定義 — 単一の情報源
│   ├── openapi/            # 生成された OpenAPI ドキュメント
│   └── pkg/hubcli/         # API カタログから組み立てられる CLI
├── ui/
│   ├── web/                # React SPA
│   └── keycloak-theme/     # Keycloak のカスタムログインテーマ
├── infra/
│   ├── tf/                 # Terraform (Keycloak レルム、クライアント)
│   └── k8s/                # Kubernetes マニフェスト (Kustomize base + overlays)
├── docs/                   # このドキュメントサイト
└── Makefile                # 共通タスク
```

トップレベルの各ディレクトリには、その領域の規約をまとめた `AGENTS.md` が
あります。コードを変更する前に必ず読んでください。

## 生成のループ

`.proto` が単一の情報源です。proto や sqlc のクエリを変更したら、次を実行します。

```bash
make gen
```

sqlc のコード、DI アダプタ、gRPC と gRPC-Gateway のスタブ、`server/openapi/` の
OpenAPI ドキュメント、Web クライアントの操作テーブル
(`ui/web/src/api/operations.ts`)、エージェントスキルが同梱する API リファレンス、
Web アプリの TypeScript 型が、この順で再生成されます。

**生成ファイルを手で編集しないこと。REST パスや権限を二度書かないこと。**
CI は Go 由来の生成物を再生成し、コミットされた内容が古ければ失敗します。

バックエンドの変更は
[`server/AGENTS.md`](https://github.com/linzhengen/hub/blob/main/server/AGENTS.md)
の手順に従います。ドメイン → proto → SQL → `make gen` → リポジトリ →
ユースケース → gRPC ハンドラ → DI → lint → テスト。

## チェック

```bash
make lint                     # サーバーへの golangci-lint
make test                     # go test ./...
cd ui/web && pnpm lint        # tsc --noEmit
cd ui/web && pnpm lint:eslint
cd ui/web && pnpm test        # vitest
cd ui/web && pnpm test:e2e    # Playwright
```

コミットは [Conventional Commits](https://www.conventionalcommits.org/) に
従います。`make init` が入れる `commit-msg` フックが検査します。

## マイグレーション

マイグレーションは `server/db/migrations/postgres` にあり、`golang-migrate`
で適用します。

```bash
make migrate
```

## Kubernetes (minikube) へのデプロイ

```bash
# minikube の Docker デーモン上でイメージをビルド
eval $(minikube docker-env) && docker build -t hub .

# プレビューしてから適用
kubectl kustomize infra/k8s/overlays/minikube
kubectl apply -k infra/k8s/overlays/minikube
```

`infra/k8s/base` は使い捨ての dev / minikube オーバーレイ向けに書かれており、
Keycloak を平文 HTTP で動かし、PostgreSQL を永続化なしで動かします。他の環境で
再利用する前に、オーバーレイが何を上書きしなければならないかを列挙した
[`infra/k8s/AGENTS.md`](https://github.com/linzhengen/hub/blob/main/infra/k8s/AGENTS.md)
を必ず読んでください。

## ドキュメントサイト

このサイトは `docs/` から生成され、`main` への push ごとに
`.github/workflows/pages.yml` が GitHub Pages へ公開します。ローカルで編集する
場合は次のとおりです。

```bash
cd docs
pnpm install
pnpm serve      # docs/_site をビルドし http://localhost:4000 で配信
```

ガイド本文は `docs/site/` 配下の Markdown です。API エクスプローラーと API
リファレンスは手書きではなく、`server/openapi/` と `make gen` が生成する
リファレンスからビルド時に組み立てられるため、proto と乖離しません。
