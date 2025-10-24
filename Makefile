.PHONY: help dev prod tutorial test watch watch-dev watch-tutorial up down logs build clean shell shell-tutorial shell-test wire version ps restart

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

watch-dev: ## 開発環境をWatchモードで起動
	@echo "$(GREEN)開発環境をWatchモードで起動中...$(RESET)"
	docker compose watch wire-dev

shell: ## 開発環境のシェルに接続
	@echo "$(CYAN)開発環境シェルに接続中...$(RESET)"
	docker compose exec wire-dev bash

##@ チュートリアル環境

tutorial: ## チュートリアル環境を起動
	@echo "$(GREEN)チュートリアル環境を起動中...$(RESET)"
	docker compose up -d wire-tutorial
	@echo "$(GREEN)チュートリアル環境が起動しました$(RESET)"
	@echo "シェルに接続: make shell-tutorial"

watch-tutorial: ## チュートリアル環境をWatchモードで起動
	@echo "$(GREEN)チュートリアル環境をWatchモードで起動中...$(RESET)"
	docker compose watch wire-tutorial

shell-tutorial: ## チュートリアル環境のシェルに接続
	@echo "$(CYAN)チュートリアル環境シェルに接続中...$(RESET)"
	docker compose exec wire-tutorial bash

run-tutorial: ## チュートリアルを実行
	@echo "$(GREEN)チュートリアルを実行中...$(RESET)"
	docker compose exec wire-tutorial go run main.go wire.go

##@ 本番環境

prod: ## 本番環境をビルド
	@echo "$(GREEN)本番環境をビルド中...$(RESET)"
	docker compose build wire-prod

wire: ## Wireコマンドを実行（引数: ARGS="..."）
	@echo "$(GREEN)Wireコマンドを実行中...$(RESET)"
	docker compose run --rm wire-prod $(ARGS)

wire-help: ## Wireのヘルプを表示
	@echo "$(CYAN)Wireヘルプ:$(RESET)"
	docker compose run --rm wire-prod --help

wire-gen: ## 現在のディレクトリでWireコード生成
	@echo "$(GREEN)Wireコード生成中...$(RESET)"
	docker compose run --rm wire-prod wire
	@echo "$(GREEN)コード生成が完了しました$(RESET)"

##@ テスト環境

test: ## テストを実行
	@echo "$(GREEN)テストを実行中...$(RESET)"
	docker compose run --rm wire-test

test-verbose: ## 詳細モードでテストを実行
	@echo "$(GREEN)テスト（詳細モード）を実行中...$(RESET)"
	docker compose run --rm wire-test go test -v ./...

test-cover: ## カバレッジ付きでテストを実行
	@echo "$(GREEN)カバレッジ付きテストを実行中...$(RESET)"
	docker compose run --rm wire-test go test -cover ./...

test-bench: ## ベンチマークを実行
	@echo "$(GREEN)ベンチマークを実行中...$(RESET)"
	docker compose run --rm wire-test go test -bench=. ./...

shell-test: ## テスト環境のシェルに接続
	@echo "$(CYAN)テスト環境シェルに接続中...$(RESET)"
	docker compose run --rm wire-test bash

##@ Watch機能

watch: ## すべてのサービスをWatchモードで起動
	@echo "$(GREEN)すべてのサービスをWatchモードで起動中...$(RESET)"
	docker compose watch

watch-logs: ## Watchモード（ログ分離）
	@echo "$(GREEN)Watchモードを起動中（ログ分離）...$(RESET)"
	docker compose up -d
	docker compose watch

##@ Docker Compose基本操作

up: ## すべてのサービスを起動
	@echo "$(GREEN)すべてのサービスを起動中...$(RESET)"
	docker compose up -d
	@echo "$(GREEN)サービスが起動しました$(RESET)"
	@$(MAKE) ps

down: ## すべてのサービスを停止・削除
	@echo "$(RED)すべてのサービスを停止中...$(RESET)"
	docker compose down
	@echo "$(RED)サービスが停止しました$(RESET)"

restart: ## すべてのサービスを再起動
	@echo "$(YELLOW)すべてのサービスを再起動中...$(RESET)"
	docker compose restart
	@echo "$(GREEN)サービスが再起動しました$(RESET)"

ps: ## 実行中のコンテナを表示
	@echo "$(CYAN)実行中のコンテナ:$(RESET)"
	@docker compose ps

logs: ## ログを表示（引数: SERVICE=wire-dev）
	@echo "$(CYAN)ログを表示中...$(RESET)"
	docker compose logs -f $(SERVICE)

logs-dev: ## 開発環境のログを表示
	@$(MAKE) logs SERVICE=wire-dev

logs-tutorial: ## チュートリアル環境のログを表示
	@$(MAKE) logs SERVICE=wire-tutorial

logs-test: ## テスト環境のログを表示
	@$(MAKE) logs SERVICE=wire-test

##@ ビルド操作

build: ## すべてのイメージをビルド
	@echo "$(GREEN)すべてのイメージをビルド中...$(RESET)"
	docker compose build

build-dev: ## 開発環境イメージをビルド
	@echo "$(GREEN)開発環境イメージをビルド中...$(RESET)"
	docker compose build wire-dev

build-prod: ## 本番環境イメージをビルド
	@echo "$(GREEN)本番環境イメージをビルド中...$(RESET)"
	docker compose build wire-prod

build-no-cache: ## キャッシュなしでビルド
	@echo "$(GREEN)キャッシュなしでビルド中...$(RESET)"
	docker compose build --no-cache

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

##@ クイックスタート

quickstart: build dev ## クイックスタート（ビルド＆起動）
	@echo ""
	@echo "$(GREEN)クイックスタート完了！$(RESET)"
	@echo ""
	@echo "次のステップ:"
	@echo "  1. $(CYAN)make shell$(RESET)        - 開発環境に接続"
	@echo "  2. $(CYAN)make tutorial$(RESET)     - チュートリアル環境を起動"
	@echo "  3. $(CYAN)make test$(RESET)         - テストを実行"
	@echo "  4. $(CYAN)make watch-dev$(RESET)    - Watchモードで開発"
	@echo ""

demo: ## デモ（チュートリアル実行）
	@echo "$(GREEN)========================================$(RESET)"
	@echo "$(GREEN)   Wireチュートリアルデモ$(RESET)"
	@echo "$(GREEN)========================================$(RESET)"
	@$(MAKE) tutorial
	@sleep 2
	@echo ""
	@echo "$(CYAN)チュートリアルを実行中...$(RESET)"
	@$(MAKE) run-tutorial
	@echo ""
	@echo "$(GREEN)デモ完了！$(RESET)"
