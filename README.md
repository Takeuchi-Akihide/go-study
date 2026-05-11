# go-study

go-study は、Go 言語でバックエンド開発の基本を 1 週間で学ぶためのサンプルプロジェクトです。

## 目的

このプロジェクトでは、Go の HTTP サーバー構築、
PostgreSQL との接続、サービス層・リポジトリ層の分離、
JSON 出力の実装を通じて、バックエンド実装の流れを学びます。

## 現在の構成

- `main.go` : HTTP サーバー、PostgreSQL 接続、マイグレーション、ハンドラー実装
- `main_test.go` : サービスとハンドラーの単体テスト
- `Dockerfile` : Go バイナリをビルドしてコンテナ化する設定
- `docker-compose.yml` : アプリと PostgreSQL の開発用構成
- `go.mod` / `go.sum` : Go モジュール依存管理

## 現在の実装内容

- `database/sql` と `pgx` ドライバで PostgreSQL に接続
- 起動時に `users` テーブルを自動作成するマイグレーション
- `/health` でサーバー稼働確認
- `/users` でユーザー一覧取得
- サービス層 (`UserService`) とリポジトリ層 (`UserRepository`) の分離
- テストでは Postgres 接続を使わず、置き換え用のフェイクストアを利用

## 起動方法

```bash
cd go-study
go run main.go
```

環境変数が指定されていない場合、デフォルトでは以下の接続先が使われます。

- `PORT`: `8080`
- `DATABASE_URL`: `postgres://dev:password@localhost:5436/app_db?sslmode=disable`

起動後、サーバーは `http://localhost:8080` で待ち受けます。

## テスト実行方法

このプロジェクトでは `go test` による単体テストを用意しています。

```bash
cd go-study
go test ./...
```

`main_test.go` では `UserStore` インターフェースを実装したフェイクストアを使い、
Postgres を使わずに `UserService` と HTTP ハンドラーの振る舞いを検証します。

## Docker ビルド / 実行方法

### Docker イメージをビルド

```bash
cd go-study
docker build -t go-study .
```

### コンテナを実行

```bash
docker run --rm -p 8080:8080 go-study
```

### docker-compose を使う

```bash
cd go-study
docker compose up --build
```

`docker-compose.yml` はアプリと Postgres の開発用構成を定義しています。
アプリは `DATABASE_URL` を使って接続します。

## エンドポイント

### GET /health

サーバーの稼働確認用エンドポイント。

- リクエスト例:
  ```bash
  curl http://localhost:8080/health
  ```
- レスポンス:
  ```json
  {
    "status": "ok"
  }
  ```

### GET /users

登録済みユーザーの一覧を取得します。

- リクエスト例:
  ```bash
  curl http://localhost:8080/users
  ```
- レスポンス例:
  ```json
  [
    {"id":1,"name":"Alice"},
    {"id":2,"name":"Bob"}
  ]
  ```

## 現時点の注意点

- `main.go` は PostgreSQL を使ってユーザーデータを取得します。
- `main_test.go` では Postgres 依存を排除し、フェイクストアで代替しています。
- `docker-compose.yml` は Postgres を含む構成のため、`docker compose up --build` でアプリと DB を同時に起動できます。
