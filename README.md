# md-viewer

Go標準ライブラリだけで実装した、コンソール用Markdownビューアーです。MarkdownをANSIカラーと罫線で読みやすく整形し、対話端末では `less` を使って閲覧できます。

## 特徴

- ファイルと標準入力に対応
- 見出し、リスト、引用、コード、テーブルをANSI装飾付きで表示
- 日本語などの全角文字を考慮した折り返しと列揃え
- テーブルの左寄せ、中央寄せ、右寄せ
- Unicode罫線とASCII罫線を選択可能
- `NO_COLOR` に対応
- Goモジュールとしての外部依存なし
- 対話端末では `less` と同じ操作でスクロール・検索
- Markdown内の端末制御文字を可視化し、ANSI/OSC注入を防止

## 必要環境

- Go 1.22以降
- `less`（任意）

`less` が見つからない場合は、組み込みの簡易ページャーへフォールバックします。

## ビルド

```bash
go build -o mdv .
```

任意のディレクトリへインストールする場合:

```bash
go install .
```

## 使い方

Markdownファイルを開く:

```bash
./mdv README.md
```

標準入力から読む:

```bash
cat README.md | ./mdv
```

サンプル文書を表示する:

```bash
go run . sample.md
```

## オプション

```text
Usage: mdv [options] [file]

  -ascii
        use ASCII table borders
  -no-color
        disable ANSI colors
  -no-pager
        print all output without paging
  -width int
        display width (default: COLUMNS or 80)
  -watch
        watch file and refresh when it changes
```

使用例:

```bash
# カラーとページャーを無効化
./mdv -no-color -no-pager README.md

# 表示幅を60桁に固定
./mdv -width 60 sample.md

# ASCII文字だけで罫線を描画
./mdv -ascii sample.md

# ファイルの変更を監視して自動更新
./mdv -watch README.md
```

`NO_COLOR` 環境変数が設定されている場合と、出力先が端末でない場合は、ANSIカラーを自動的に無効化します。表示幅は `-width`、`COLUMNS`、既定値80の順で決定します。

## ページャー操作

対話端末ではレンダリング結果を `less -R` へ渡します。代表的な操作は次のとおりです。

| キー | 操作 |
|---|---|
| `j` / `k`、矢印キー | 1行移動 |
| Space / `b` | 1画面進む／戻る |
| `g` / `G` | 先頭／末尾へ移動 |
| `/文字列` | 前方検索 |
| `n` / `N` | 次／前の検索結果 |
| `q` | 終了 |

ページャーを使わず、そのまま出力する場合は `-no-pager` を指定してください。

## ファイル監視

`-watch` を指定すると、ファイルの内容を250ms間隔で監視し、変更されたときだけ画面を再描画します。

```bash
./mdv -watch README.md
```

監視中も `j` / `k` と矢印キーで1行、Space / `b` で1画面、`g` / `G` で先頭／末尾へ移動できます。`q` または `Ctrl+C` で終了します。ファイルが変更されても現在位置を可能な限り維持します。表示領域は実際の端末高から3行差し引き、ステータス行と下部余白を確保します。

`-watch` はファイル指定時のみ利用でき、標準入力やリダイレクト出力とは併用できません。監視中は `less` を使用せず、LinuxまたはmacOSのraw terminal制御を使用します。

## 対応しているMarkdown

- ATX見出し（`#`〜`######`）
- 段落
- 箇条書き、番号付きリスト
- 引用
- 区切り線
- バッククォートまたはチルダによるfenced code block
- インラインコード
- 太字、斜体、取り消し線
- リンクと画像の代替テキスト
- GFM風パイプテーブル
  - 左寄せ、中央寄せ、右寄せ
  - セル内の `\|` エスケープ
  - 端末幅に応じた列幅の縮小と折り返し

動作確認用の構文は [sample.md](sample.md) にまとめています。

## 制約

CommonMarkやGFMへの完全準拠を目的としたパーサーではありません。現在、次の機能は対象外です。

- ネストしたリストや引用
- HTMLブロック
- 脚注、タスクリスト
- 複雑なインライン構文の入れ子
- raw terminal制御を使った内蔵フルスクリーンUI

Unicodeの表示幅は日本語と一般的な全角文字を優先して独自に判定しています。複雑な絵文字シーケンスは、端末によって幅がずれる場合があります。

## 開発

テスト:

```bash
go test ./...
```

静的検査:

```bash
go vet ./...
```

## リリース

`v` で始まるタグをGitHubへpushすると、GitHub Actionsがテストとクロスビルドを実行し、自動生成したGitHub Releaseへバイナリを添付します。

```bash
git tag v0.1.0
git push origin v0.1.0
```

公開されるファイル:

- `mdv-macos-arm64`: Apple Silicon搭載Mac向け
- `mdv-linux-x86_64`: Linux x86_64向け
- `checksums.txt`: SHA-256チェックサム

Releaseはリポジトリの「Releases」ページからダウンロードできます。ワークフローは [`.github/workflows/release.yml`](.github/workflows/release.yml) にあります。

実装計画と設計上の判断は [plan.md](plan.md) を参照してください。

## ライセンス

ライセンスはまだ指定されていません。
