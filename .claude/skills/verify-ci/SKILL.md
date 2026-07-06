---
name: verify-ci
description: CI と同等の検証一式(go test -race / go vet / gofmt -s / 依存関係リスト差分)を Docker コンテナ内で一括実行し、結果を報告する。コード変更後や PR 作成前の検証に使う。
---

# verify-ci

CI(`internal/runtests.sh` 相当)のチェックを Docker コンテナ内で実行し、全結果を報告する。

## 手順

1. `docker compose ps` でコンテナ状態を確認。`wire-dev` が起動していなければ `docker compose up -d wire-dev` で起動する。
2. 以下の4チェックを順に実行し、それぞれの生の出力を記録する。途中で失敗しても止めず、可能な限り全チェックを実行してから報告する:

   ```bash
   docker compose exec -T wire-dev go test -mod=readonly -race ./...
   docker compose exec -T wire-dev go vet ./...
   docker compose exec -T wire-dev gofmt -s -l .    # 出力ゼロが合格
   docker compose exec -T wire-dev sh -c './internal/listdeps.sh | diff ./internal/alldeps -'
   ```

3. チェックごとに合格 / 不合格を報告する。不合格があれば該当出力を引用し、以下の対処を実施して再チェックする:
   - gofmt 不合格 → `docker compose exec wire-dev gofmt -s -w .` を実行
   - alldeps 差分 → import 変更が意図どおりであることを確認したうえで `docker compose exec wire-dev sh -c './internal/listdeps.sh > ./internal/alldeps'` で更新
   - test / vet 不合格 → 出力をもとに原因を調査して修正(勝手にテストを skip・削除しない)
4. 4チェックすべての合格を確認して完了とする。実行していないチェックを「合格」と報告しない。
