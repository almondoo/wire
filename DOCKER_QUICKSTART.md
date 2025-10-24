# Docker環境 クイックスタートガイド

WireプロジェクトをDockerで簡単に始めるためのガイドです。

## 必要要件

- Docker Engine 20.10+
- Docker Compose v2.0+
- GNU Make（オプション、推奨）

## 最速スタート

### 1. クイックスタート（推奨）

```bash
# すべてをビルドして開発環境を起動
make quickstart

# シェルに接続
make shell
```

### 2. チュートリアルを試す

```bash
# チュートリアルデモを実行
make demo
```

## よく使うコマンド

### Makeコマンド一覧

```bash
# ヘルプを表示
make help

# 開発環境
make dev              # 開発環境を起動
make shell            # シェルに接続
make watch-dev        # Watchモードで起動

# チュートリアル
make tutorial         # チュートリアル環境を起動
make shell-tutorial   # シェルに接続
make run-tutorial     # チュートリアルを実行
make demo             # デモを実行

# テスト
make test             # テストを実行
make test-cover       # カバレッジ付きテスト
make test-bench       # ベンチマーク

# Wireコマンド
make wire-gen         # Wireコード生成
make wire-help        # Wireヘルプ

# ビルド
make build            # すべてのイメージをビルド
make build-dev        # 開発環境イメージをビルド
make build-prod       # 本番環境イメージをビルド

# クリーンアップ
make clean            # コンテナを停止・削除
make clean-volumes    # ボリュームも削除
make clean-all        # すべて削除（イメージ含む）

# 情報表示
make ps               # 実行中のコンテナ表示
make logs-dev         # 開発環境のログ表示
make version          # バージョン表示
make volumes          # ボリューム一覧
```

## 典型的なワークフロー

### 開発フロー

```bash
# 1. 環境を起動
make quickstart

# 2. 開発環境に接続
make shell

# 3. コンテナ内で作業
cd /workspace
go build ./cmd/wire
./wire --help

# 4. 別ターミナルでWatchモードを有効化（オプション）
make watch-dev

# 5. ホスト側でコードを編集
# → 自動的にコンテナに同期される
```

### チュートリアルフロー

```bash
# 最も簡単な方法
make demo

# または段階的に
make tutorial         # 起動
make shell-tutorial   # 接続
make run-tutorial     # 実行
```

### テストフロー

```bash
# 基本的なテスト
make test

# より詳細に
make test-verbose     # 詳細出力
make test-cover       # カバレッジ
make test-bench       # ベンチマーク
```

## Docker Compose Watchについて

Docker Compose Watchは、ファイル変更を自動検出してコンテナに同期する最新機能です。

### 使い方

```bash
# 開発環境でWatch
make watch-dev

# チュートリアル環境でWatch
make watch-tutorial

# すべてのサービスでWatch
make watch
```

### 何が起こるか

- **Goソースコード（.go）の変更**: 即座にコンテナに同期
- **go.mod/go.sumの変更**: イメージを自動再ビルド
- **チュートリアルファイルの変更**: コンテナを自動再起動

## トラブルシューティング

### コマンドが見つからない

```bash
# Docker Composeのバージョン確認
docker compose version

# v2.0以降が必要です
```

### Makeが使えない

Makeがインストールされていない場合は、Docker Composeコマンドを直接使用できます:

```bash
# 開発環境を起動
docker compose up -d wire-dev

# シェルに接続
docker compose exec wire-dev bash

# Watchモード
docker compose watch wire-dev
```

### ポートが使用中

```bash
# 実行中のコンテナを確認
make ps

# すべて停止
make down
```

### ディスク容量不足

```bash
# 未使用リソースをクリーンアップ
make prune

# すべて削除（注意）
make clean-all
```

### キャッシュの問題

```bash
# キャッシュなしでビルド
make build-no-cache
```

## より詳しい情報

詳細なドキュメントは以下を参照してください:

- [dockerfiles/README.md](dockerfiles/README.md) - 完全なドキュメント
- [docker-compose.yml](docker-compose.yml) - サービス設定
- [Makefile](Makefile) - すべてのMakeコマンド

## ヘルプ

```bash
# Makeコマンドの完全なヘルプ
make help

# Wireツールのヘルプ
make wire-help
```

## 次のステップ

1. `make quickstart` でスタート
2. `make shell` でコンテナに接続
3. プロジェクトを探索
4. コードを編集して `make watch-dev` で自動同期
5. `make test` でテストを実行

楽しい開発を！
