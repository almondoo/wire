# Wire Docker環境

このディレクトリには、Wireプロジェクトの開発用Docker環境が含まれています。

## 必要要件

- Docker Engine
- Docker Compose v2

Docker Composeのバージョン確認:
```bash
docker compose version
```

## 構成

```
wire/
├── dockerfiles/
│   ├── Dockerfile.dev    # 開発用Dockerfile
│   └── README.md         # このファイル
├── docker-compose.yml    # Docker Compose設定
├── Makefile              # Make設定（便利なコマンド集）
└── .dockerignore         # Docker無視ファイル
```

`docker-compose.yml` は `wire-dev` サービスを1つだけ定義しています。

**特徴:**
- `golang:1.26-alpine` ベース（`dockerfiles/Dockerfile.dev`）
- 開発ツール（git, make, bash, gcc, musl-dev）を含む
- ホストのプロジェクトディレクトリを `/workspace` に直接マウント
- Goモジュール・ビルドキャッシュ用のボリューム（`go-mod-cache` / `go-build-cache`）を使用

## クイックスタート

### Makeを使用する場合（推奨）

```bash
# ヘルプを表示
make help

# コンテナの状態を確認
make ps

# 開発環境イメージをビルド
make build

# 開発環境を起動
make up

# シェルに接続
make shell

# 全テスト実行（CIと同等）
make test

# Docker 起動→全テスト実行→停止までを一発実行
make test-full

# gofmt + go vet
make lint

# 開発環境を停止・削除
make down
```

### Docker Composeを直接使用する場合

```bash
# イメージをビルド
docker compose build

# 開発環境を起動
docker compose up -d

# コンテナの状態を確認
docker compose ps

# シェルに接続
docker compose exec wire-dev /bin/bash

# テスト実行
docker compose exec wire-dev go test -mod=readonly -race ./...

# 開発環境を停止・削除
docker compose down
```

## キャッシュボリューム

以下のキャッシュボリュームを使用してビルド速度を向上させています:

- `go-mod-cache`: Goモジュールキャッシュ（`/go/pkg/mod`）
- `go-build-cache`: Goビルドキャッシュ（`/root/.cache/go-build`）

```bash
# キャッシュボリュームの確認
docker volume ls | grep wire

# キャッシュのクリア（必要な場合）
docker compose down -v
```

## 参考リソース

- [Docker Compose公式ドキュメント](https://docs.docker.com/compose/)
- [Compose Specification](https://docs.docker.com/reference/compose-file/)
- [Wireプロジェクト](https://github.com/almondoo/wire)
