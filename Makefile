.PHONY: help dev up down logs build clean shell version ps restart test test-verbose test-cover test-bench test-all-versions generate-version-errs install-wire go-mod-download go-mod-tidy fmt vet lint quickstart

# デフォルトターゲット
.DEFAULT_GOAL := help

# カラー出力
CYAN := \033[36m
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
RESET := \033[0m

##@ ヘルプ

help: ## このヘルプメッセージを表示
	@echo "$(CYAN)Wire Docker環境 - Makefile コマンド$(RESET)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "使用方法:\n  make $(CYAN)<target>$(RESET)\n\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(CYAN)%-20s$(RESET) %s\n", $$1, $$2 } /^##@/ { printf "\n$(GREEN)%s$(RESET)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ 開発環境

dev: ## 開発環境を起動
	@echo "$(GREEN)開発環境を起動中...$(RESET)"
	docker compose up -d wire-dev
	@echo "$(GREEN)開発環境が起動しました$(RESET)"
	@echo "シェルに接続: make shell"

shell: ## 開発環境のシェルに接続
	@echo "$(CYAN)開発環境シェルに接続中...$(RESET)"
	docker compose exec wire-dev bash

##@ テスト

test: ## テストを実行
	@echo "$(GREEN)テストを実行中...$(RESET)"
	docker compose exec wire-dev go test -mod=readonly -race ./...

test-verbose: ## 詳細モードでテストを実行
	@echo "$(GREEN)テスト（詳細モード）を実行中...$(RESET)"
	docker compose exec wire-dev go test -v -mod=readonly -race ./...

test-cover: ## カバレッジ付きでテストを実行
	@echo "$(GREEN)カバレッジ付きテストを実行中...$(RESET)"
	docker compose exec wire-dev go test -cover -mod=readonly -race ./...

test-bench: ## ベンチマークを実行
	@echo "$(GREEN)ベンチマークを実行中...$(RESET)"
	docker compose exec wire-dev go test -bench=. ./...

test-all-versions: ## 全Goバージョンでテストを実行
	@echo "$(GREEN)全Goバージョンでテストを実行中...$(RESET)"
	cd internal/wire && ./test_all_versions.sh

##@ バージョン別エラーファイル管理

generate-version-errs: ## 全Goバージョン用のwire_errs.txtを生成
	@echo "$(GREEN)全Goバージョン用のエラーファイルを生成中...$(RESET)"
	cd internal/wire && ./generate_version_errs.sh
	@echo "$(GREEN)生成完了$(RESET)"

##@ Docker Compose基本操作

up: ## 開発環境を起動
	@echo "$(GREEN)開発環境を起動中...$(RESET)"
	docker compose up -d
	@echo "$(GREEN)開発環境が起動しました$(RESET)"
	@$(MAKE) ps

down: ## 開発環境を停止・削除
	@echo "$(RED)開発環境を停止中...$(RESET)"
	docker compose down
	@echo "$(RED)開発環境が停止しました$(RESET)"

restart: ## 開発環境を再起動
	@echo "$(YELLOW)開発環境を再起動中...$(RESET)"
	docker compose restart
	@echo "$(GREEN)開発環境が再起動しました$(RESET)"

ps: ## 実行中のコンテナを表示
	@echo "$(CYAN)実行中のコンテナ:$(RESET)"
	@docker compose ps

logs: ## ログを表示
	@echo "$(CYAN)ログを表示中...$(RESET)"
	docker compose logs -f wire-dev

##@ ビルド操作

build: ## 開発環境イメージをビルド
	@echo "$(GREEN)開発環境イメージをビルド中...$(RESET)"
	docker compose build

build-no-cache: ## キャッシュなしでビルド
	@echo "$(GREEN)キャッシュなしでビルド中...$(RESET)"
	docker compose build --no-cache

build-wire: ## wire.goをビルド
	@echo "$(GREEN)wire.goをビルド中...$(RESET)"
	docker compose exec wire-dev go build -o bin/wire ./wire.go
	@echo "$(GREEN)ビルド完了: bin/wire$(RESET)"

##@ クリーンアップ

clean: ## コンテナを停止・削除
	@echo "$(RED)クリーンアップ中...$(RESET)"
	docker compose down
	@echo "$(GREEN)クリーンアップ完了$(RESET)"

clean-volumes: ## コンテナとボリュームを削除
	@echo "$(RED)コンテナとボリュームを削除中...$(RESET)"
	docker compose down -v
	@echo "$(GREEN)削除完了$(RESET)"

clean-all: ## すべて削除（イメージ含む）
	@echo "$(RED)すべてを削除中（イメージ含む）...$(RESET)"
	docker compose down --rmi all -v
	@echo "$(GREEN)完全削除完了$(RESET)"

prune: ## 未使用のDockerリソースを削除
	@echo "$(RED)未使用のDockerリソースを削除中...$(RESET)"
	docker system prune -f
	@echo "$(GREEN)削除完了$(RESET)"

##@ 情報表示

version: ## Docker ComposeとDockerのバージョンを表示
	@echo "$(CYAN)Docker Compose バージョン:$(RESET)"
	@docker compose version
	@echo ""
	@echo "$(CYAN)Docker バージョン:$(RESET)"
	@docker version --format '{{.Server.Version}}'

info: ## システム情報を表示
	@echo "$(CYAN)Docker システム情報:$(RESET)"
	@docker info

volumes: ## ボリューム一覧を表示
	@echo "$(CYAN)Wireプロジェクトのボリューム:$(RESET)"
	@docker volume ls | grep wire || echo "ボリュームが見つかりません"

images: ## イメージ一覧を表示
	@echo "$(CYAN)Wireプロジェクトのイメージ:$(RESET)"
	@docker images | grep wire || echo "イメージが見つかりません"

##@ 便利なコマンド

install-wire: ## コンテナ内でWireツールをインストール
	@echo "$(GREEN)Wireツールをインストール中...$(RESET)"
	docker compose exec wire-dev go install github.com/almondoo/wire/cmd/wire@latest

go-mod-download: ## Go依存関係をダウンロード
	@echo "$(GREEN)Go依存関係をダウンロード中...$(RESET)"
	docker compose exec wire-dev go mod download

go-mod-tidy: ## go.modをクリーンアップ
	@echo "$(GREEN)go.modをクリーンアップ中...$(RESET)"
	docker compose exec wire-dev go mod tidy

fmt: ## コードをフォーマット
	@echo "$(GREEN)コードをフォーマット中...$(RESET)"
	docker compose exec wire-dev go fmt ./...

vet: ## コードを静的解析
	@echo "$(GREEN)コードを静的解析中...$(RESET)"
	docker compose exec wire-dev go vet ./...

lint: fmt vet ## コードをフォーマット＆静的解析
