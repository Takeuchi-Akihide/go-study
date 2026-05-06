# go-study

go-study は、Go 言語でバックエンド開発の基本を 1 週間で学ぶためのサンプルプロジェクトです。

## 目的

このプロジェクトでは、Go の標準ライブラリを使って HTTP サーバーを立て、
エンドポイントの定義、JSON の入出力、リクエストメソッドごとの処理、
簡単な CRUD 操作を理解することを目標としています。

## 現在の実装内容

- `main.go` に単純な HTTP サーバーを実装
- `/health` でサーバーの稼働確認
- `/users` でユーザー一覧取得とユーザー作成
- `/users/{id}` でユーザーの取得、更新、削除
- メモリ上の `map[int]User` を使った簡易データストア
- JSON のパースとレスポンスを標準ライブラリのみで処理

## 起動方法

```bash
cd go-study
go run main.go
```

起動後、サーバーは `http://localhost:8080` で待ち受けます。

## エンドポイント

### GET /health

サーバーの状態確認用エンドポイント。

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
- レスポンス:
  ```json
  [
    {"id":1,"name":"Alice"},
    {"id":2,"name":"Bob"}
  ]
  ```

### POST /users

新しいユーザーを作成します。

- リクエスト例:
  ```bash
  curl -X POST http://localhost:8080/users \
    -H "Content-Type: application/json" \
    -d '{"name":"Charlie"}'
  ```
- レスポンス:
  ```json
  {"id":3,"name":"Charlie"}
  ```

### GET /users/{id}

指定した ID のユーザーを取得します。

- リクエスト例:
  ```bash
  curl http://localhost:8080/users/1
  ```
- 成功レスポンス:
  ```json
  {"id":1,"name":"Alice"}
  ```

### PUT /users/{id}

指定した ID のユーザー情報を更新します。

- リクエスト例:
  ```bash
  curl -X PUT http://localhost:8080/users/1 \
    -H "Content-Type: application/json" \
    -d '{"name":"Alice Updated"}'
  ```
- 成功レスポンス:
  ```json
  {"id":1,"name":"Alice Updated"}
  ```

### DELETE /users/{id}

指定した ID のユーザーを削除します。

- リクエスト例:
  ```bash
  curl -X DELETE http://localhost:8080/users/1
  ```
- 成功レスポンス:
  ```json
  {"status":"deleted"}
  ```

## 現時点の注意点

- データはメモリ上に保持され、サーバー再起動で消えます。
- 実践的なアプリケーションでは、永続ストレージやルーティングライブラリを追加する予定です。
