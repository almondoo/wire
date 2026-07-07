# CLAUDE.md

このファイルは、このリポジトリで作業する Claude Code (claude.ai/code) にガイダンスを提供します。

## 方針

機能完成済みのメンテナンスフォーク。バグ修正のみ受け付け、新機能の提案・実装は行わない。

## コマンド(Docker 必須)

テスト・ビルド・lint は必ず Docker コンテナ内で実行する(ローカルの Go は使わない)。コンテナは `make up` で起動。

```bash
make test          # コンテナ内で go test -mod=readonly -race ./...(コンテナ起動済みが前提)
make test-full     # Docker 起動→全テスト→停止まで一発実行(常駐コンテナも停止する点に注意)
make lint          # コンテナ内で go vet + gofmt -s チェック
make shell         # 開発コンテナに入る

# 単一テストの実行
docker compose exec wire-dev go test -mod=readonly -race -run TestXxx ./internal/wire

# フォーマット修正
docker compose exec wire-dev gofmt -s -w .
```

## CI が強制するチェック

- `gofmt -s`(簡略化フラグ付き)と `go vet`。違反すると CI が失敗する。
- 依存関係リストの一致(Linux で強制)。import を追加・変更したら必ず更新する:
  ```bash
  docker compose exec wire-dev sh -c './internal/listdeps.sh > ./internal/alldeps'
  ```
  一致確認: `docker compose exec wire-dev sh -c './internal/listdeps.sh | diff ./internal/alldeps -'`
- CI マトリックスは Go 1.19 / 1.24 / 1.25 / 1.26 × Linux / macOS / Windows。go.mod は `go 1.19` のため、Go 1.20 以降でしか使えない構文・標準ライブラリ API を使わない。

## 非自明な注意点

- `_tutorial/` は `_` 始まりディレクトリのため `go build/test/vet ./...` の対象外だが、`gofmt -s` チェックの対象には含まれる。
- 生成コード(wire_gen.go)のビルドタグは `format.Source` 経由で新旧両形式(`//go:build` + `// +build`)が出力される。
- `internal/wire/integration_test.go` は `GOPROXY=off` で動作する(テストにネットワークは不要)。
- `.github/workflows/go-compat.yml` が週次 cron で Go の stable/oldstable に対する互換性チェックを実行し、失敗時は `internal/reportgocompat.sh` が `go-version-check` ラベルで重複防止しつつ Issue を自動起票する。public リポジトリの scheduled workflow は 60 日間リポジトリに活動がないと自動停止される点に注意。

## ドキュメント規約

英日ドキュメントは同期必須: `README.md` ↔ `JP_README.md`、`docs/*.md` ↔ `docs/jp/*.md`、`_tutorial/README.md` ↔ `_tutorial/JP_README.md`。英語版を変更したら対応する日本語版も必ず更新する。
