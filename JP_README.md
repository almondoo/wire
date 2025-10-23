# Wire: Goにおける自動初期化

[![Build Status](https://github.com/google/wire/actions/workflows/tests.yml/badge.svg?branch=main)](https://github.com/google/wire/actions)
[![godoc](https://godoc.org/github.com/google/wire?status.svg)][godoc]

> [!WARNING]
> このプロジェクトはメンテナンスされなくなりました。
>
> Wireを更新または拡張したい場合は、フォークして実施してください。

Wireは、[依存性注入][]を使用してコンポーネントを接続する作業を自動化するコード生成ツールです。コンポーネント間の依存関係はWireでは関数パラメータとして表現され、グローバル変数ではなく明示的な初期化が推奨されます。Wireは実行時の状態やリフレクションなしで動作するため、Wireで使用するために書かれたコードは、手書きの初期化にも有用です。

概要については、[紹介ブログ記事][]をご覧ください。

[依存性注入]: https://en.wikipedia.org/wiki/Dependency_injection
[紹介ブログ記事]: https://blog.golang.org/wire
[godoc]: https://godoc.org/github.com/google/wire
[travis]: https://travis-ci.com/google/wire

## インストール

次のコマンドを実行してWireをインストールします：

```shell
go install github.com/google/wire/cmd/wire@latest
```

そして、`$GOPATH/bin`が`$PATH`に追加されていることを確認してください。

## ドキュメント

- [チュートリアル][]
- [ユーザーガイド][]
- [ベストプラクティス][]
- [FAQ][]

[チュートリアル]: ./_tutorial/README.md
[ベストプラクティス]: ./docs/best-practices.md
[FAQ]: ./docs/faq.md
[ユーザーガイド]: ./docs/guide.md

## プロジェクトの状況

バージョンv0.3.0の時点で、Wireは*ベータ版*であり、機能は完成していると考えられています。設計された目的のタスクについてはうまく機能しており、可能な限りシンプルに保つことを優先しています。

現時点では新機能の受け入れは行っていませんが、バグレポートと修正については喜んで受け付けます。

## コミュニティ

ご質問については、[GitHub Discussions](https://github.com/google/wire/discussions)をご利用ください。

このプロジェクトはGoの[行動規範][]の対象です。

[行動規範]: ./CODE_OF_CONDUCT.md
[go-cloud mailing list]: https://groups.google.com/forum/#!forum/go-cloud
