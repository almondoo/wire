# Wire Docker環境

このディレクトリには、Wireプロジェクト用のDocker環境が含まれています。

## 必要要件

- Docker Engine 20.10+
- Docker Compose v2.0+
- GNU Make（オプション、推奨）

Docker Composeのバージョン確認:
```bash
docker compose version
```

## クイックスタート

### Makeを使用する場合（推奨）

```bash
# ヘルプを表示
make help

# クイックスタート（ビルド＆起動）
make quickstart

# 開発環境を起動
make dev

# シェルに接続
make shell

# チュートリアルデモを実行
make demo
```

### Docker Composeを直接使用する場合

```bash
# 開発環境を起動
docker compose up -d wire-dev

# シェルに接続
docker compose exec wire-dev bash
```

## ディレクトリ構造

```
wire/
├── dockerfiles/
│   ├── Dockerfile.dev    # 開発用Dockerfile
│   ├── Dockerfile.prod   # 本番用Dockerfile（マルチステージビルド）
│   └── README.md         # このファイル
├── docker-compose.yml    # Docker Compose設定
├── Makefile              # Make設定（便利なコマンド集）
└── .dockerignore         # Docker無視ファイル
```

## サービス一覧

### 1. wire-dev（開発環境）
開発用のインタラクティブな環境です。

**特徴:**
- Go 1.19ベース
- 開発ツール（git, make, bash, gcc）含む
- ホットリロード対応（Docker Compose Watch）
- Goモジュールキャッシュ対応

**使用方法:**

Makeを使用:
```bash
# 起動
make dev

# シェルに接続
make shell

# Watchモードで起動（ファイル変更を自動同期）
make watch-dev
```

Docker Composeを直接使用:
```bash
# 起動
docker compose up -d wire-dev

# シェルに接続
docker compose exec wire-dev bash

# Watchモードで起動
docker compose watch wire-dev
```

### 2. wire-prod（本番環境）
Wireツールの本番ビルド・実行環境です。

**特徴:**
- マルチステージビルドで最適化
- 最小限のAlpineベースイメージ
- セキュリティ強化（read-only, no-new-privileges）
- 非rootユーザーで実行

**使用方法:**

Makeを使用:
```bash
# 本番環境をビルド
make prod

# Wireコマンドの実行
make wire ARGS="--help"

# Wireコード生成
make wire-gen
```

Docker Composeを直接使用:
```bash
# Wireコマンドの実行
docker compose run --rm wire-prod

# カスタムコマンドの実行
docker compose run --rm wire-prod wire --help
```

### 3. wire-tutorial（チュートリアル環境）
チュートリアルコードの実行・学習用環境です。

**特徴:**
- _tutorialディレクトリにフォーカス
- ファイル変更時の自動再起動
- インタラクティブシェル

**使用方法:**

Makeを使用:
```bash
# チュートリアル環境を起動
make tutorial

# シェルに接続
make shell-tutorial

# チュートリアル実行
make run-tutorial

# Watchモード
make watch-tutorial

# デモを実行
make demo
```

Docker Composeを直接使用:
```bash
# 起動
docker compose up -d wire-tutorial

# シェルに接続
docker compose exec wire-tutorial bash

# チュートリアル実行（コンテナ内）
go run main.go wire.go

# Watchモード
docker compose watch wire-tutorial
```

### 4. wire-test（テスト環境）
自動テスト実行環境です。

**使用方法:**

Makeを使用:
```bash
# テスト実行
make test

# 詳細モードでテスト
make test-verbose

# カバレッジ付きテスト
make test-cover

# ベンチマーク
make test-bench
```

Docker Composeを直接使用:
```bash
# テスト実行
docker compose run --rm wire-test

# 特定のパッケージをテスト
docker compose run --rm wire-test go test -v ./internal/wire
```

## Docker Compose Watch機能

Docker Compose Watchは、ファイル変更を自動的に検出してコンテナに同期する機能です。

### 基本的な使い方

Makeを使用:
```bash
# 開発環境のWatch
make watch-dev

# チュートリアル環境のWatch
make watch-tutorial

# すべてのサービスのWatch
make watch

# ログを分離してWatch
make watch-logs
```

Docker Composeを直接使用:
```bash
# 単一サービスのWatch
docker compose watch wire-dev

# 複数サービスのWatch
docker compose watch

# ログを分離してWatch
docker compose up -d
docker compose watch
```

### Watchアクション

本プロジェクトでは以下のアクションを使用しています:

1. **sync**: ファイルをコンテナに同期（再起動なし）
   - Goソースコード（*.go）
   - 高速な開発サイクル

2. **sync+restart**: ファイル同期後にコンテナを再起動
   - チュートリアルファイル
   - 変更の即座反映

3. **rebuild**: イメージを再ビルド
   - go.mod / go.sum
   - 依存関係の変更時

### 無視されるファイル

以下のファイル/ディレクトリは監視対象外です:
- `.git/` - Gitディレクトリ
- `.github/` - GitHub設定
- `dockerfiles/` - Dockerファイル
- `docs/` - ドキュメント
- `internal/wire/testdata/` - テストデータ
- `wire_gen.go` - 生成ファイル

## 一般的な使用例

### 開発ワークフロー

Makeを使用:
```bash
# 1. 開発環境を起動（Watchモード）
make watch-dev

# 2. 別ターミナルでシェルに接続
make shell

# 3. コンテナ内で開発
cd /workspace
go build ./cmd/wire
./wire --help

# 4. ホスト側でファイルを編集
# → 自動的にコンテナに同期される
```

Docker Composeを直接使用:
```bash
# 1. 開発環境を起動（Watchモード）
docker compose watch wire-dev

# 2. 別ターミナルでシェルに接続
docker compose exec wire-dev bash

# 3. コンテナ内で開発
cd /workspace
go build ./cmd/wire
./wire --help

# 4. ホスト側でファイルを編集
# → 自動的にコンテナに同期される
```

### チュートリアルの実行

Makeを使用:
```bash
# デモを実行（最も簡単）
make demo

# または手動で
make watch-tutorial  # Watchモードで起動
make shell-tutorial  # 別ターミナルで接続
make run-tutorial    # チュートリアル実行
```

Docker Composeを直接使用:
```bash
# Watchモードでチュートリアル環境を起動
docker compose watch wire-tutorial

# 別ターミナルで接続
docker compose exec wire-tutorial bash

# チュートリアルを実行
go run main.go wire.go

# ホスト側でmain.goを編集
# → 自動的に再起動される
```

### Wireコード生成

Makeを使用:
```bash
# 現在のディレクトリでWireコード生成
make wire-gen

# Wireヘルプを表示
make wire-help
```

Docker Composeを直接使用:
```bash
# プロジェクトディレクトリでwireコマンドを実行
docker compose run --rm -v $(pwd):/app wire-prod wire

# 特定のディレクトリで実行
docker compose run --rm -v $(pwd)/examples:/app -w /app wire-prod wire
```

### テストの実行

Makeを使用:
```bash
# 全テスト実行
make test

# 詳細モード
make test-verbose

# カバレッジ付き
make test-cover

# ベンチマーク
make test-bench
```

Docker Composeを直接使用:
```bash
# 全テスト実行
docker compose run --rm wire-test

# カバレッジ付きテスト
docker compose run --rm wire-test go test -cover ./...

# ベンチマーク
docker compose run --rm wire-test go test -bench=. ./...
```

## パフォーマンス最適化

### キャッシュボリューム

以下のキャッシュボリュームを使用してビルド速度を向上させています:

- `go-mod-cache`: Goモジュールキャッシュ（`/go/pkg/mod`）
- `go-build-cache`: Goビルドキャッシュ（`/root/.cache/go-build`）

### キャッシュの管理

```bash
# キャッシュボリュームの確認
docker volume ls | grep wire

# キャッシュのクリア（必要な場合）
docker compose down -v
```

## トラブルシューティング

### Watchモードが動作しない

```bash
# Docker Composeのバージョン確認
docker compose version  # v2.22.0以降が必要

# ログで詳細を確認
docker compose watch --verbose
```

### パーミッションエラー

```bash
# コンテナのユーザーIDを確認
docker compose exec wire-dev id

# 必要に応じてホスト側のファイルパーミッションを調整
chmod -R 755 .
```

### ボリュームマウントの問題

```bash
# ボリュームを再作成
docker compose down -v
docker compose up -d
```

## セキュリティ考慮事項

### 本番環境（wire-prod）

- ✅ 非rootユーザーで実行（UID/GID: 1000）
- ✅ Read-onlyファイルシステム
- ✅ no-new-privileges有効
- ✅ 最小限のベースイメージ（Alpine）

### 開発環境（wire-dev）

開発環境は利便性を優先していますが、以下に注意してください:

- ⚠️ rootユーザーで実行
- ⚠️ ホストファイルシステムをマウント
- 💡 本番環境では使用しないでください

## 環境のクリーンアップ

```bash
# コンテナの停止・削除
docker compose down

# ボリュームも含めて削除
docker compose down -v

# イメージも削除
docker compose down --rmi all -v
```

## 参考リソース

- [Docker Compose公式ドキュメント](https://docs.docker.com/compose/)
- [Docker Compose Watch](https://docs.docker.com/compose/how-tos/file-watch/)
- [Compose Specification](https://docs.docker.com/reference/compose-file/)
- [Wireプロジェクト](https://github.com/almondoo/wire)
