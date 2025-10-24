# Wire チュートリアル

例を通じてWireの使い方を学びましょう。[Wireガイド][guide]はツールの使用方法について詳細なドキュメントを提供しています。より大規模なサーバーにWireが適用されているのを見たい読者には、[Go Cloudのゲストブックサンプル][guestbook]がコンポーネントの初期化にWireを使用しています。ここでは、Wireの使い方を理解するために小さなgreeterプログラムを構築します。完成品はこのREADMEと同じディレクトリにあります。

[guestbook]: https://github.com/google/go-cloud/tree/master/samples/guestbook
[guide]:     https://github.com/google/wire/blob/master/docs/guide.md

## Greeterプログラムを構築する最初のパス

特定のメッセージでゲストに挨拶するgreeterを使ったイベントをシミュレートする小さなプログラムを作成しましょう。

まず、3つの型を作成します: 1) greeter用のメッセージ、2) そのメッセージを伝えるgreeter、3) greeterがゲストに挨拶することから始まるイベント。この設計では、3つの`struct`型があります:

``` go
type Message string

type Greeter struct {
    // ... 未定
}

type Event struct {
    // ... 未定
}
```

`Message`型は単に文字列をラップします。今のところ、常にハードコードされたメッセージを返すシンプルな初期化関数を作成します:

``` go
func NewMessage() Message {
    return Message("Hi there!")
}
```

`Greeter`は`Message`への参照が必要です。そこで、`Greeter`用の初期化関数も作成しましょう。

``` go
func NewGreeter(m Message) Greeter {
    return Greeter{Message: m}
}

type Greeter struct {
    Message Message // <- Messageフィールドを追加
}
```

初期化関数で`Greeter`に`Message`フィールドを割り当てます。これで、`Greeter`に`Greet`メソッドを作成する際に`Message`を使用できます:

``` go
func (g Greeter) Greet() Message {
    return g.Message
}
```

次に、`Event`が`Greeter`を持つ必要があるため、それ用の初期化関数も作成します。

``` go
func NewEvent(g Greeter) Event {
    return Event{Greeter: g}
}

type Event struct {
    Greeter Greeter // <- Greeterフィールドを追加
}
```

次に、`Event`を開始するメソッドを追加します:

``` go
func (e Event) Start() {
    msg := e.Greeter.Greet()
    fmt.Println(msg)
}
```

`Start`メソッドは、小さなアプリケーションのコアを保持しています: greeterに挨拶を発行するように指示し、そのメッセージを画面に出力します。

これでアプリケーションのすべてのコンポーネントの準備ができたので、Wireを使用せずにすべてのコンポーネントを初期化するのに何が必要か見てみましょう。main関数は次のようになります:

``` go
func main() {
    message := NewMessage()
    greeter := NewGreeter(message)
    event := NewEvent(greeter)

    event.Start()
}
```

まずメッセージを作成し、次にそのメッセージでgreeterを作成し、最後にそのgreeterでイベントを作成します。すべての初期化が完了したら、イベントを開始する準備が整います。

私たちは[依存性注入][di]設計原則を使用しています。実際には、各コンポーネントが必要とするものを渡すことを意味します。このスタイルの設計は、テストしやすいコードを書くのに適しており、ある依存関係を別のものと簡単に入れ替えることができます。

[di]: https://stackoverflow.com/questions/130794/what-is-dependency-injection

## Wireを使用したコード生成

依存性注入の欠点の1つは、多くの初期化ステップが必要なことです。コンポーネントの初期化プロセスをよりスムーズにするためにWireを使用する方法を見てみましょう。

まず、`main`関数を次のように変更しましょう:

``` go
func main() {
    e := InitializeEvent()

    e.Start()
}
```

次に、`wire.go`という別のファイルで`InitializeEvent`を定義します。ここからが面白くなります:

``` go
// wire.go

func InitializeEvent() Event {
    wire.Build(NewEvent, NewGreeter, NewMessage)
    return Event{}
}
```

各コンポーネントを順番に初期化して次のコンポーネントに渡すという手間をかける代わりに、使用したい初期化関数を渡す`wire.Build`への単一の呼び出しがあります。Wireでは、初期化関数は「プロバイダ」として知られ、特定の型を提供する関数です。コンパイラを満足させるために、`Event`のゼロ値を戻り値として追加します。`Event`に値を追加してもWireはそれらを無視することに注意してください。実際、インジェクタの目的は、`Event`を構築するためにどのプロバイダを使用するかについての情報を提供することであるため、ファイルの先頭にビルド制約を使用して最終バイナリから除外します:

``` go
//+build wireinject

```

注意: [ビルド制約][constraint]には空白の末尾行が必要です。

Wire用語では、`InitializeEvent`は「インジェクタ」です。インジェクタが完成したので、`wire`コマンドラインツールを使用する準備が整いました。

次のコマンドでツールをインストールします:

``` shell
go install github.com/google/wire/cmd/wire@latest
```

次に、上記のコードと同じディレクトリで単に`wire`を実行します。Wireは`InitializeEvent`インジェクタを見つけ、必要なすべての初期化ステップで本体が埋められた関数を生成します。結果は`wire_gen.go`という名前のファイルに書き込まれます。

Wireが私たちのために何をしてくれたか見てみましょう:

``` go
// wire_gen.go

func InitializeEvent() Event {
    message := NewMessage()
    greeter := NewGreeter(message)
    event := NewEvent(greeter)
    return event
}
```

上で書いたものとまったく同じです！これは3つのコンポーネントだけのシンプルな例なので、手動で初期化関数を書くことはそれほど苦痛ではありません。はるかに複雑なコンポーネントに対してWireがどれほど便利かを想像してください。Wireを使用する場合、`wire.go`と`wire_gen.go`の両方をソース管理にコミットします。

[constraint]: https://godoc.org/go/build#hdr-Build_Constraints

## Wireで変更を加える

Wireがより複雑なセットアップをどのように処理するかのごく一部を示すために、`Event`の初期化関数をエラーを返すようにリファクタリングして、何が起こるか見てみましょう。

``` go
func NewEvent(g Greeter) (Event, error) {
    if g.Grumpy {
        return Event{}, errors.New("could not create event: event greeter is grumpy")
    }
    return Event{Greeter: g}, nil
}
```

時々`Greeter`が不機嫌で、`Event`を作成できない場合があるとしましょう。`NewGreeter`初期化関数は次のようになります:

``` go
func NewGreeter(m Message) Greeter {
    var grumpy bool
    if time.Now().Unix()%2 == 0 {
        grumpy = true
    }
    return Greeter{Message: m, Grumpy: grumpy}
}
```

`Greeter`構造体に`Grumpy`フィールドを追加し、初期化関数の呼び出し時刻がUnixエポックからの秒数が偶数の場合、友好的なgreeterの代わりに不機嫌なgreeterを作成します。

次に、`Greet`メソッドは次のようになります:

``` go
func (g Greeter) Greet() Message {
    if g.Grumpy {
        return Message("Go away!")
    }
    return g.Message
}
```

これで、不機嫌な`Greeter`が`Event`には適していないことがわかります。そのため、`NewEvent`は失敗する可能性があります。`main`は、`InitializeEvent`が実際に失敗する可能性があることを考慮する必要があります:

``` go
func main() {
    e, err := InitializeEvent()
    if err != nil {
        fmt.Printf("failed to create event: %s\n", err)
        os.Exit(2)
    }
    e.Start()
}
```

また、`InitializeEvent`を更新して戻り値に`error`型を追加する必要があります:

``` go
// wire.go

func InitializeEvent() (Event, error) {
    wire.Build(NewEvent, NewGreeter, NewMessage)
    return Event{}, nil
}
```

セットアップが完了したら、再び`wire`コマンドを実行する準備が整いました。注意: `wire_gen.go`ファイルを生成するために一度`wire`を実行した後、`go generate`も使用できます。コマンドを実行すると、`wire_gen.go`ファイルは次のようになります:

``` go
// wire_gen.go

func InitializeEvent() (Event, error) {
    message := NewMessage()
    greeter := NewGreeter(message)
    event, err := NewEvent(greeter)
    if err != nil {
        return Event{}, err
    }
    return event, nil
}
```

Wireは`NewEvent`プロバイダが失敗する可能性があることを検出し、生成されたコード内で正しいことを行いました: エラーをチェックし、エラーがある場合は早期にリターンします。

## インジェクタのシグネチャを変更する

もう1つの改善として、Wireがインジェクタのシグネチャに基づいてコードを生成する方法を見てみましょう。現在、`NewMessage`内でメッセージをハードコードしています。実際には、呼び出し側がそのメッセージを好きなように変更できるようにする方がはるかに優れています。そこで、`InitializeEvent`を次のように変更しましょう:

``` go
func InitializeEvent(phrase string) (Event, error) {
    wire.Build(NewEvent, NewGreeter, NewMessage)
    return Event{}, nil
}
```

これで`InitializeEvent`により、呼び出し側が`Greeter`が使用する`phrase`を渡すことができます。また、`NewMessage`に`phrase`引数を追加します:

``` go
func NewMessage(phrase string) Message {
    return Message(phrase)
}
```

再び`wire`を実行すると、ツールが`phrase`値を`Message`として`Greeter`に渡す初期化関数を生成したことがわかります。素晴らしい！

``` go
// wire_gen.go

func InitializeEvent(phrase string) (Event, error) {
    message := NewMessage(phrase)
    greeter := NewGreeter(message)
    event, err := NewEvent(greeter)
    if err != nil {
        return Event{}, err
    }
    return event, nil
}
```

Wireはインジェクタの引数を検査し、引数のリストに文字列(例: `phrase`)を追加したことを確認し、同様にすべてのプロバイダの中で`NewMessage`が文字列を受け取ることを確認し、`phrase`を`NewMessage`に渡します。

## 役立つエラーでミスをキャッチする

Wireがコード内のミスを検出したときに何が起こるか、またWireのエラーメッセージが問題を修正するのにどのように役立つかを見てみましょう。

例えば、インジェクタ`InitializeEvent`を書く際に、`Greeter`のプロバイダを追加するのを忘れたとしましょう。何が起こるか見てみましょう:

``` go
func InitializeEvent(phrase string) (Event, error) {
    wire.Build(NewEvent, NewMessage) // おっと！ Greeterのプロバイダを追加するのを忘れました
    return Event{}, nil
}
```

`wire`を実行すると、次のように表示されます:

``` shell
# 読みやすさのためにエラーを複数行に分割
$GOPATH/src/github.com/google/wire/_tutorial/wire.go:24:1:
inject InitializeEvent: no provider found for github.com/google/wire/_tutorial.Greeter
(required by provider of github.com/google/wire/_tutorial.Event)
wire: generate failed
```

Wireは有用な情報を教えてくれています: `Greeter`のプロバイダが見つかりません。エラーメッセージは`Greeter`型への完全なパスを出力していることに注意してください。また、問題が発生した行番号とインジェクタ名も教えてくれます: `InitializeEvent`内の24行目。さらに、エラーメッセージはどのプロバイダが`Greeter`を必要としているかを教えてくれます。それは`Event`型です。`Greeter`のプロバイダを渡せば、問題は解決します。

あるいは、`wire.Build`に1つ余分なプロバイダを提供した場合はどうなるでしょうか?

``` go
func NewEventNumber() int  {
    return 1
}

func InitializeEvent(phrase string) (Event, error) {
     // おっと！ NewEventNumberは使用されていません。
    wire.Build(NewEvent, NewGreeter, NewMessage, NewEventNumber)
    return Event{}, nil
}
```

Wireは、使用されていないプロバイダがあることを親切に教えてくれます:

``` shell
$GOPATH/src/github.com/google/wire/_tutorial/wire.go:24:1:
inject InitializeEvent: unused provider "NewEventNumber"
wire: generate failed
```

`wire.Build`の呼び出しから使用されていないプロバイダを削除すると、エラーは解決されます。

## まとめ

ここで行ったことをまとめましょう。まず、対応する初期化関数、つまりプロバイダを持つ多数のコンポーネントを作成しました。次に、インジェクタ関数を作成し、それが受け取る引数と返す型を指定しました。次に、すべての必要なプロバイダを提供する`wire.Build`への呼び出しでインジェクタ関数を埋めました。最後に、`wire`コマンドを実行して、すべての異なる初期化関数を接続するコードを生成しました。インジェクタに引数とエラー戻り値を追加したとき、再び`wire`を実行すると、生成されたコードに必要なすべての更新が行われました。

ここでの例は小さいですが、Wireの力の一部を示しており、依存性注入を使用したコードの初期化から多くの苦痛を取り除く方法を示しています。さらに、Wireを使用して生成されたコードは、私たちが書くものとよく似ています。ユーザーをWireにコミットさせるような特殊な型はありません。代わりに、単に生成されたコードです。私たちはそれを好きなように使用できます。最後に、考慮すべきもう1つの点は、コンポーネントの初期化に新しい依存関係を追加することがいかに簡単かということです。Wireにコンポーネントの提供方法(つまり、初期化)を教える限り、依存関係グラフのどこにでもそのコンポーネントを追加でき、Wireが残りを処理します。

最後に、Wireはここで説明されていない多数の追加機能をサポートしていることを述べる価値があります。プロバイダは[プロバイダセット][sets]にグループ化できます。[インターフェースのバインディング][interfaces]、[値のバインディング][values]、および[クリーンアップ関数][cleanup]のサポートがあります。詳細については、[高度な機能][advanced]セクションを参照してください。

[advanced]:   https://github.com/google/wire/blob/master/docs/guide.md#advanced-features
[cleanup]:    https://github.com/google/wire/blob/master/docs/guide.md#cleanup-functions
[interfaces]: https://github.com/google/wire/blob/master/docs/guide.md#binding-interfaces
[sets]:       https://github.com/google/wire/blob/master/docs/guide.md#defining-providers
[values]:     https://github.com/google/wire/blob/master/docs/guide.md#binding-values
