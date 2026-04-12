# gem-image

Vertex AI Gemini 2.5 Flash（ネイティブ画像生成）を使った画像生成・編集CLI。

テキストプロンプトからの画像生成や、既存画像の編集をコマンドラインから実行する。
シェルスクリプトやパイプラインを通じたバッチ処理に対応。

## 前提条件

- **Google Cloudプロジェクト** — Vertex AI APIが有効であること
- **Application Default Credentials** — `gcloud auth application-default login` を実行

## インストール

```bash
git clone https://github.com/nlink-jp/gem-image.git
cd gem-image
make build
# バイナリ: dist/gem-image
```

## 設定

| 変数 | 必須 | デフォルト | 説明 |
|------|------|-----------|------|
| `GEMIMAGE_PROJECT` | はい | — | GCPプロジェクトID |
| `GEMIMAGE_LOCATION` | いいえ | `us-central1` | Vertex AIリージョン |
| `GEMIMAGE_MODEL` | いいえ | `gemini-2.5-flash-image` | Geminiモデル名 |

ツール固有の変数が未設定の場合、`GOOGLE_CLOUD_PROJECT` / `GOOGLE_CLOUD_LOCATION` にフォールバックする。

設定ファイルも利用可能（`~/.config/gem-image/config.toml`）:

```toml
[gcp]
project  = "your-project-id"
location = "us-central1"

[model]
name = "gemini-2.5-flash-image"
```

優先順位: CLIフラグ > 環境変数 > 設定ファイル > デフォルト値

## 使い方

```bash
# テキストから画像生成
gem-image -p "窓辺に座る猫" -o cat.png

# 既存画像を編集
gem-image -p "空に虹を追加して" -i photo.png -o edited.png

# 複数の入力画像
gem-image -p "これらをコラージュにして" -i a.png -i b.png -o collage.png

# JPEG出力（拡張子から自動判定）
gem-image -p "海に沈む夕日" -o sunset.jpg

# 明示的なフォーマット指定
gem-image -p "山の風景" -o landscape.bin --format jpeg

# 標準入力からプロンプト（パイプライン）
echo "コーヒーショップのミニマルなロゴ" | gem-image -o logo.png

# モデルの上書き
gem-image -p "水彩画" -o art.png -m gemini-2.5-flash-image
```

### フラグ

| フラグ | 短縮 | デフォルト | 説明 |
|--------|------|-----------|------|
| `--prompt` | `-p` | — | 画像生成プロンプト（未指定時はstdinから読み取り） |
| `--input` | `-i` | — | 入力画像パス（複数指定可） |
| `--output` | `-o` | — | 出力ファイルパス（必須） |
| `--format` | — | `png` | 出力形式: `png` または `jpeg` |
| `--config` | `-c` | — | 設定ファイルパス |
| `--model` | `-m` | — | モデル名の上書き |
| `--debug` | — | `false` | デバッグ出力を有効化 |

### 出力形式の決定ロジック

1. `-o` の拡張子が `.png`/`.jpg`/`.jpeg` → その形式を使用
2. それ以外は `--format` フラグの値を使用
3. デフォルト: `png`

### 終了コード

| コード | 意味 |
|--------|------|
| 0 | 正常終了 |
| 1 | 一般エラー |
| 2 | 入力検証エラー |
| 3 | APIエラー |
| 4 | セーフティフィルタによるブロック |

### トークン使用量

リクエスト後にstderrにトークン消費量を表示する:

```
tokens: input=218 output=1290 total=1508
```

## セキュリティ

- **プロンプトインジェクション対策** — ユーザープロンプトは [nlk/guard](https://github.com/nlink-jp/nlk) のノンスタグXMLでラッピングしてからAPIに送信
- **入力検証** — 画像ファイルはマジックバイトで形式を検証（拡張子だけに依存しない）
- **パストラバーサル防止** — すべてのファイルパスを正規化・検証
- **シークレット非出力** — プロジェクトIDやトークンをログに出力しない

## ライセンス

MIT
