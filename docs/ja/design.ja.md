# 詳細設計: gem-image

> Generated: 2026-04-12
> Status: Draft

## 概要

gem-image は Vertex AI Gemini 2.5 Flash の画像生成・編集機能を
CLI から利用するためのツールである。

---

## CLI インターフェース

### 基本構文

```
gem-image -p <prompt> [-i <image>...] -o <output> [--format png|jpeg] [--debug]
```

### フラグ詳細

| フラグ | 短縮 | 型 | 必須 | デフォルト | 説明 |
|--------|------|----|------|-----------|------|
| `--prompt` | `-p` | string | ※ | — | 画像生成プロンプト |
| `--input` | `-i` | []string | No | — | 入力画像パス（リピート可） |
| `--output` | `-o` | string | Yes | — | 出力ファイルパス |
| `--format` | — | string | No | `png` | 出力形式（`png` \| `jpeg`） |
| `--config` | `-c` | string | No | `~/.config/gem-image/config.toml` | 設定ファイルパス |
| `--model` | `-m` | string | No | (設定ファイル参照) | モデル名オーバーライド |
| `--debug` | — | bool | No | false | デバッグ出力有効化 |
| `--version` | `-v` | bool | No | — | バージョン表示 |

※ `-p` 未指定時は stdin からプロンプトを読み取り。`-p` と stdin の両方がある場合、`-p` を優先。

### 出力形式の決定ロジック

```
if -o の拡張子が .png or .jpeg/.jpg:
    → 拡張子に従う
elif --format が指定:
    → --format に従う
else:
    → png（デフォルト）
```

### 終了コード

| コード | 意味 |
|--------|------|
| 0 | 正常終了 |
| 1 | 一般エラー（設定不備、ファイル I/O エラー等） |
| 2 | 入力検証エラー（不正なファイル形式、パストラバーサル等） |
| 3 | API エラー（認証失敗、クォータ超過、ネットワークエラー等） |
| 4 | セーフティフィルタによるブロック |

---

## パッケージ設計

### cmd/root.go

責務：CLI フラグ解析、入力の組み立て、各パッケージの呼び出しオーケストレーション

```go
func Execute(version string)
func runGenerate(cmd *cobra.Command, args []string) error
```

**処理フロー：**

1. 設定読み込み（config.Load）
2. プロンプト取得（-p / stdin）
3. 入力画像検証（security.ValidateImagePath）
4. プロンプト保護（security.WrapPrompt）
5. API 呼び出し（client.Generate）
6. 画像書き込み（image.WriteFile）
7. トークン情報表示（stderr）

### internal/config

gem-cli と同一パターン。TOML 設定ファイル + 環境変数オーバーライド。

```go
type Config struct {
    GCP   GCPConfig
    Model ModelConfig
}

type GCPConfig struct {
    Project  string `toml:"project"`
    Location string `toml:"location"`
}

type ModelConfig struct {
    Name string `toml:"name"`
}

func Load(path string) (*Config, error)
func (c *Config) ApplyFlags(model string)
```

**設定ファイル例（config.example.toml）：**

```toml
[gcp]
project  = "your-project-id"
location = "global"

[model]
name = "gemini-3.1-flash-image"
```

### internal/client

Gemini API クライアント。`GenerateContent` を画像生成用に設定して呼び出す。

```go
type Client struct {
    inner *genai.Client
    model string
}

type GenerateResult struct {
    ImageData []byte         // 生成画像のバイナリ
    MIMEType  string         // "image/png" or "image/jpeg"
    Text      string         // テキスト応答（あれば）
    Usage     *UsageInfo
}

type UsageInfo struct {
    InputTokens  int64
    OutputTokens int64
    TotalTokens  int64
}

func New(ctx context.Context, cfg *config.Config) (*Client, error)
func (c *Client) Generate(ctx context.Context, opts *GenerateOpts) (*GenerateResult, error)
func (c *Client) Close() error
```

**GenerateOpts：**

```go
type GenerateOpts struct {
    SystemPrompt string
    UserPrompt   string       // ラップ済みプロンプト
    Images       []*ImageInput
    OutputFormat string       // "image/png" or "image/jpeg"
}

type ImageInput struct {
    Data     []byte
    MIMEType string
}
```

**API 設定：**

```go
config := &genai.GenerateContentConfig{
    ResponseModalities: []string{
        string(genai.ModalityImage),
    },
}
```

### internal/image

画像ファイルの入出力を担当。

**input.go：**

```go
// ReadImageFile は画像ファイルを読み込み、検証済みの ImageInput を返す
func ReadImageFile(path string) (*client.ImageInput, error)

// detectMIME はマジックバイトからMIMEタイプを判定する
func detectMIME(data []byte) (string, error)
```

**output.go：**

```go
// WriteFile は画像データをファイルに書き込む（パーミッション 0644）
func WriteFile(path string, data []byte) error

// ResolveFormat は -o の拡張子と --format から出力MIMEタイプを決定する
func ResolveFormat(outputPath string, formatFlag string) string
```

### internal/security

セキュリティ関連の処理を集約。

**guard.go：**

```go
// WrapPrompt はユーザープロンプトを nlk/guard でラッピングする
func WrapPrompt(userPrompt string) (systemPrompt string, wrappedUser string, err error)
```

**validate.go：**

```go
// ValidateImagePath はファイルパスの安全性を検証する
func ValidateImagePath(path string) (string, error)

// ValidateOutputPath は出力パスの安全性を検証する
func ValidateOutputPath(path string) error

// ValidateImageData はマジックバイトとサイズを検証する
func ValidateImageData(data []byte) error
```

**検証項目：**

| 検証 | 実装 |
|------|------|
| パストラバーサル | `filepath.Abs()` + `filepath.EvalSymlinks()` |
| マジックバイト | PNG: `\x89PNG\r\n\x1a\n`, JPEG: `\xFF\xD8\xFF` |
| ファイルサイズ | 上限チェック（モデル入力制限に準拠） |
| 出力先ディレクトリ | 存在確認・書き込み権限確認 |

---

## エラーハンドリング

### API エラー

| エラー種別 | 対応 |
|-----------|------|
| 認証エラー (401/403) | ADC 設定を促すメッセージを表示 |
| クォータ超過 (429) | レートリミット超過を通知、リトライはシェル側で制御 |
| セーフティフィルタ (FinishReasonSafety) | 終了コード 4 で明示的に通知 |
| モデル未対応 | モデル名の確認を促すメッセージ |
| 画像未生成 | レスポンスに InlineData がない場合のエラー |

### 入力エラー

| エラー種別 | 対応 |
|-----------|------|
| `-o` 未指定 | 必須フラグエラー（Cobra が処理） |
| プロンプト未入力 | stdin が端末かつ `-p` なしでエラー |
| 画像ファイル不正 | マジックバイト不一致でエラー（終了コード 2） |
| パストラバーサル | 検出時にエラー（終了コード 2） |

---

## テスト設計

### ユニットテスト

| パッケージ | テスト内容 |
|-----------|-----------|
| config | TOML 読み込み、環境変数オーバーライド、必須項目チェック |
| client | GenerateOpts の構築、レスポンスからの画像抽出、UsageMetadata 取得 |
| image/input | マジックバイト判定、MIME 検出、不正ファイル拒否 |
| image/output | ファイル書き込み、フォーマット解決ロジック |
| security/guard | nlk/guard ラッピング、システムプロンプト展開 |
| security/validate | パス正規化、トラバーサル検出、サイズ制限 |

### E2E テスト

- テキストから画像生成 → ファイル保存 → マジックバイト検証
- 入力画像 + プロンプト → 編集画像生成 → ファイル保存
- 複数入力画像での動作確認
- セーフティフィルタブロック時の終了コード確認
- stderr へのトークン情報出力確認
- 不正入力（パストラバーサル、偽装ファイル）の拒否確認

---

## 依存関係

### 直接依存

| パッケージ | バージョン | 用途 |
|-----------|-----------|------|
| `github.com/spf13/cobra` | latest | CLI フレームワーク |
| `github.com/BurntSushi/toml` | latest | 設定ファイル解析 |
| `google.golang.org/genai` | latest | Gemini API SDK |
| `github.com/nlink-jp/nlk` | latest | guard（プロンプトインジェクション対策） |

### 間接依存

genai SDK が Google Cloud 関連パッケージを pull する（gem-cli と同様）。
