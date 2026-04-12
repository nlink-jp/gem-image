# 開発計画書: gem-image

> Generated: 2026-04-12
> Status: Draft

## フェーズ概要

| フェーズ | 内容 | 成果物 |
|---------|------|--------|
| Phase 1 | 設計 | RFP、アーキテクチャ、詳細設計、開発計画 |
| Phase 2 | Core 実装 | 動作するバイナリ + テスト |
| Phase 3 | リリース | README、CHANGELOG、リリースバイナリ |

---

## Phase 1: 設計（現フェーズ）

### タスク

| # | タスク | 成果物 | 状態 |
|---|--------|--------|------|
| 1.1 | RFP 作成 | `docs/ja/gem-image-rfp.ja.md`, `docs/en/gem-image-rfp.md` | 完了 |
| 1.2 | アーキテクチャドキュメント | `docs/ja/architecture.ja.md`, `docs/en/architecture.md` | 完了 |
| 1.3 | 詳細設計 | `docs/ja/design.ja.md`, `docs/en/design.md` | 完了 |
| 1.4 | 開発計画書 | `docs/ja/development-plan.ja.md`, `docs/en/development-plan.md` | 完了 |
| 1.5 | レビュー・承認 | — | 未着手 |

### 完了条件

- 全設計ドキュメントの日英版が揃っている
- レビューで指摘された修正が反映されている

---

## Phase 2: Core 実装

### 前提条件

- Phase 1 のドキュメントが承認済み
- Vertex AI API が有効化済み
- ADC 認証が設定済み

### タスク

実装は以下の順序で進める。各タスクは 1 コミット単位を目安とする。

| # | タスク | 依存 | 説明 |
|---|--------|------|------|
| 2.1 | プロジェクトスキャフォールド | — | go mod init、Makefile、ディレクトリ構造、.gitignore |
| 2.2 | config パッケージ | 2.1 | TOML 読み込み + env override + テスト |
| 2.3 | security パッケージ | 2.1 | パス検証、マジックバイト検証、nlk/guard ラッパー + テスト |
| 2.4 | image パッケージ | 2.3 | 画像読み込み・書き込み・フォーマット解決 + テスト |
| 2.5 | client パッケージ | 2.2 | Gemini API クライアント + テスト |
| 2.6 | CLI 統合 | 2.3, 2.4, 2.5 | cmd/root.go でフラグ解析・オーケストレーション |
| 2.7 | E2E テスト | 2.6 | 実際の API を使った統合テスト |
| 2.8 | AGENTS.md / CLAUDE.md | 2.7 | プロジェクトメタ情報 |

### 依存関係図

```
2.1 (scaffold)
 ├── 2.2 (config)
 │    └── 2.5 (client) ──┐
 └── 2.3 (security)      │
      └── 2.4 (image) ───┤
                          ▼
                     2.6 (CLI統合)
                          │
                          ▼
                     2.7 (E2E)
                          │
                          ▼
                     2.8 (meta docs)
```

### 完了条件

- `make build` でバイナリがビルドできる
- `go test ./...` が全て通る
- 実データで以下の E2E テストが通る：
  - テキスト → 画像生成
  - 画像 + テキスト → 画像編集
  - 複数入力画像 → 画像生成
  - 不正入力の拒否
  - stderr にトークン情報が表示される

---

## Phase 3: リリース

### 前提条件

- Phase 2 の E2E テストが通過済み
- ビルドバイナリでの実機シミュレーション完了

### タスク

| # | タスク | 説明 |
|---|--------|------|
| 3.1 | README.md | 英語版 README（プロジェクトルート） |
| 3.2 | README.ja.md | 日本語版 README（プロジェクトルート） |
| 3.3 | CHANGELOG.md | v0.1.0 エントリ |
| 3.4 | ビルドバイナリ実機テスト | `make build` したバイナリで最終確認 |
| 3.5 | Git リポジトリ作成 | nlink-jp/gem-image |
| 3.6 | リリース | tag v0.1.0、`gh release create`、バイナリアップロード |
| 3.7 | util-series サブモジュール登録 | サブモジュール追加 + ポインタ更新 |
| 3.8 | org プロファイル更新 | `.github/profile/README.md` に gem-image 追加 |
| 3.9 | check-org.sh 実行 | 最終検証 |

### 完了条件

- GitHub リリースが公開されている
- util-series のサブモジュールポインタが更新されている
- org プロファイルに gem-image が記載されている
- `check-org.sh` が通る

---

## リスクと対策

| リスク | 影響 | 対策 |
|--------|------|------|
| Gemini 2.5 Flash Image の API 仕様変更 | 実装の手戻り | GA 版リリースまではプレビューモデル名を設定可能にしておく |
| セーフティフィルタの過剰ブロック | ユーザビリティ低下 | 明確なエラーメッセージで原因を通知 |
| nlk の guard パッケージ API 変更 | ビルドエラー | go.mod でバージョン固定 |
| genai SDK のメジャーバージョンアップ | 互換性問題 | gem-cli の更新と同期して対応 |
