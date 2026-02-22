# CLAUDE.md

このファイルは、このリポジトリで作業する Claude Code (claude.ai/code) にガイダンスを提供します。

## Wire とは

Wire は Go 向けのコンパイル時依存性注入コードジェネレータです。[google/wire](https://github.com/google/wire) のメンテナンスフォークです。プロジェクトは**機能完成済み** — バグ修正のみ受け付けています。

Wire は AST 解析によりプロバイダ関数を分析し（ランタイムリフレクションなし）、依存関係の DAG を構築し、`wire_gen.go` ファイルに初期化コードを生成します。

## よく使うコマンド

**重要: テストやビルドコマンドは必ず Docker コンテナ内で実行すること。** ローカル環境ではなく `docker compose exec wire-dev` 経由で実行する。

```bash
# Docker ベースの開発（推奨）
make test          # コンテナ内でテスト実行
make lint          # gofmt + go vet
make shell         # 開発コンテナに入る

# コンテナ内で直接実行する場合
docker compose exec wire-dev go test -mod=readonly -race ./...
docker compose exec wire-dev go test -v -run TestXxx ./internal/wire
docker compose exec wire-dev gofmt -s -w .

# 依存関係の一致確認（CI が Linux で強制）
docker compose exec wire-dev sh -c './internal/listdeps.sh | diff ./internal/alldeps -'

# import 変更後の依存関係リスト更新
docker compose exec wire-dev sh -c './internal/listdeps.sh > ./internal/alldeps'
```

## アーキテクチャ

### コード生成パイプライン

```
ユーザーの wire.go（//+build wireinject 付き）
  → parse.go: AST 解析、wire.Build() 呼び出しの検出、プロバイダセットの抽出
  → analyze.go: 型→プロバイダマップの構築、依存グラフの解決、循環検出、トポロジカルソート
  → wire.go (internal): 初期化コードの生成、出力のフォーマット
  → wire_gen.go の書き出し（//+build !wireinject 付き）
```

### 主要ソースファイル

- **`wire.go`（ルート）** — 公開 API。マーカー型のみ（`NewSet`, `Build`, `Bind`, `Value`, `Struct`, `FieldsOf`）。実行時には呼ばれず、CLI が AST 経由で読み取る。
- **`internal/wire/parse.go`** — AST 解析。インジェクタ関数（本体に `wire.Build()` 呼び出しを1つだけ持つ関数）の検出、プロバイダセットの抽出、バインディングの解決。
- **`internal/wire/analyze.go`** — 依存関係の解決。インジェクタの戻り値型から型→プロバイダマップを構築、グラフを逆方向に走査、循環検出、トポロジカルソート済みプロバイダリストを生成。
- **`internal/wire/wire.go`** — コード生成。解析済みグラフからフォーマット済み Go ソースを生成。
- **`cmd/wire/main.go`** — CLI。サブコマンド: `gen`（デフォルト）、`check`、`diff`、`show`。

### テスト構造

ユニットテストがパイプラインの各段階をカバー:
- **`parse_test.go`** — AST 解析、プロバイダセット抽出のテスト
- **`analyze_test.go`** — 依存関係解決、循環検出のテスト
- **`errors_test.go`** — エラーメッセージのテスト
- **`generate_test.go`** — コード生成ヘルパー（unexport, disambiguate, zeroValue 等）のテスト
- **`copyast_test.go`** — AST コピーのテスト
- **`benchmark_test.go`** — パフォーマンスベンチマーク

## プロジェクト固有の規約

- **フォーマット**: `gofmt -s`（簡略化フラグ付き）が必須。これなしでは CI が失敗する。
- **フィーチャーブランチ**: AI 生成ブランチには `claude/` プレフィックスを使用。
- **PR はスカッシュマージ**。PR 作成後に `git rebase` や `git push --force` は使わない — 代わりに `git merge` を使う。
- **依存関係の追跡**: import を追加・変更したら `./internal/listdeps.sh > ./internal/alldeps` で依存関係リストを更新する。CI が Linux でこれをチェックする。
- **CI マトリックス**: Go 1.19〜1.25 で Linux、macOS、Windows の全環境でテスト実行。
- **日本語ドキュメント**: `docs/jp/`、`JP_README.md`、`_tutorial/JP_README.md`、Makefile のコメントは日本語。
