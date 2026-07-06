.PHONY: help ps build up down test test-full lint shell

.DEFAULT_GOAL := help

##@ ヘルプ

help: ## このヘルプメッセージを表示
	@awk 'BEGIN {FS = ":.*##"; printf "使用方法:\n  make \033[36m<target>\033[0m\n\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[32m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Docker

ps:
	docker compose ps

build: ## 開発環境イメージをビルド
	docker compose build

up: ## 開発環境を起動
	docker compose up -d

down: ## 開発環境を停止・削除
	docker compose down

##@ テスト

test: ## 全テスト実行（CI と同等）
	docker compose exec wire-dev go test -mod=readonly -race ./...

test-full: ## Docker 起動→全テスト実行→停止までを一発実行
	docker compose up -d --build
	docker compose exec -T wire-dev go test -mod=readonly -race ./...; \
	status=$$?; \
	docker compose down; \
	exit $$status

lint: ## gofmt + go vet
	docker compose exec wire-dev go vet ./...
	docker compose exec wire-dev sh -c 'test -z "$$(gofmt -s -l . | grep -v testdata)"' || { echo "run: gofmt -s -w ."; exit 1; }

shell: ## 開発コンテナに入る
	docker compose exec wire-dev /bin/bash
