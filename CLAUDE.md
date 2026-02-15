# Vespa Knowledge Hub

GitHubリポジトリとNotionを横断検索できる個人用ナレッジベースシステム

## 特徴

- 🔍 **統合検索**: 自分のコードとドキュメントを一箇所から検索
- ⚡ **高速**: Vespa検索エンジンによる高速な全文検索
- 🎯 **高精度**: BM25ランキングアルゴリズムによる関連性の高い結果
- 🔄 **段階的構築**: まずGitHub、その後Notionを統合

## 現在の状態

✅ **Phase 1 (進行中)**: GitHub検索システム
- GitHubリポジトリのコード検索
- ファイルパス、言語でのフィルタリング

🚧 **Phase 2 (予定)**: Notion統合
- Notionページの検索
- GitHubとNotionの横断検索

## 必要要件

- Docker & Docker Compose
- Go 1.25 以上
- Node.js 18 以上
- [Task](https://taskfile.dev/) - タスクランナー
- [vespa-cli](https://docs.vespa.ai/en/vespa-cli.html) - Vespaコマンドラインツール
- GitHub Personal Access Token

## クイックスタート

### 1. 必要なツールをインストール

```bash
# Task（タスクランナー）
# macOS
brew install go-task/tap/go-task

# Linux
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin

# vespa-cli
brew install vespa-cli  # macOS
# その他: https://docs.vespa.ai/en/vespa-cli.html
```

### 2. リポジトリをクローン

```bash
git clone https://github.com/yourusername/vespa-knowledge-hub.git
cd vespa-knowledge-hub
```

### 3. 環境変数を設定

```bash
cp .env.example .env
# .envを編集してGITHUB_TOKENとTARGET_REPOSを設定
```

### 4. 開発環境をセットアップ

```bash
task dev:setup
```

### 5. データをインデックス

```bash
export GITHUB_TOKEN=ghp_your_token_here
export TARGET_REPOS=owner/repo1,owner/repo2
task index:run
```

### 6. 開発サーバーを起動

```bash
# Terminal 1: APIサーバー
task backend:run

# Terminal 2: フロントエンド
task frontend:dev
```

ブラウザで http://localhost:5173 を開く

### すべてのタスクを見る

```bash
task --list
```

## 使い方

### 基本検索

検索バーにキーワードを入力するだけ：
```
authentication
```

### フィルタ付き検索（API経由）

```bash
# 特定リポジトリのみ
curl "http://localhost:3000/api/search?q=auth&repo=owner/repo"

# 特定言語のみ
curl "http://localhost:3000/api/search?q=function&language=go"
```

### よく使うTaskコマンド

```bash
# タスク一覧
task

# Vespaの状態確認
task vespa:health

# ログ確認
task vespa:logs

# バックエンドテスト
task backend:test

# クリーンアップ
task clean
```

## 設定

### 環境変数

| 変数名 | 説明 | デフォルト |
|--------|------|-----------|
| GITHUB_TOKEN | GitHub Personal Access Token | (必須) |
| VESPA_URL | VespaエンドポイントURL | http://localhost:8080 |
| TARGET_REPOS | 対象リポジトリ（カンマ区切り） | (必須) |
| API_PORT | APIサーバーのポート | 3000 |

### GitHub Token の作成

1. GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
2. "Generate new token" をクリック
3. 権限: `repo` (Full control of private repositories)
4. トークンをコピーして環境変数に設定

## アーキテクチャ

```
Frontend (React) → Backend API (Go) → Vespa Search Engine
                         ↓
                   GitHub API
```

**詳しい開発の進め方は [docs/GETTING_STARTED.md](docs/GETTING_STARTED.md) を参照してください。**

詳細なアーキテクチャは [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) を参照

## トラブルシューティング

### Vespaに接続できない

```bash
# ヘルスチェック
task vespa:health

# コンテナ状態確認
docker ps
task vespa:logs
```

### 検索結果が出ない

```bash
# ドキュメント数確認
curl 'http://localhost:8080/search/?yql=select%20*%20from%20knowledge_item%20where%20true'
```

ドキュメントが0件の場合、インデックスを再実行

### GitHub API Rate Limit

```bash
# 残りリクエスト数確認
curl -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/rate_limit
```

## 開発

詳細な開発ガイドは [docs/GETTING_STARTED.md](docs/GETTING_STARTED.md) を参照

```bash
# バックエンドテスト
task backend:test

# フロントエンドビルド
task frontend:build

# 開発環境のリセット
task clean:all
```

## ロードマップ

- [] Phase 1: GitHub検索システム
  - [] Vespaセットアップ
  - [] GitHub Indexer
  - [ ] 検索API
  - [ ] フロントエンドUI
- [ ] Phase 2: Notion統合
  - [ ] Notion API連携
  - [ ] セマンティック検索（埋め込みベクトル）
  - [ ] 横断検索UI

## ライセンス

MIT License

## 参考資料

- [Vespa Documentation](https://docs.vespa.ai/)
- [GitHub REST API](https://docs.github.com/en/rest)
- [Notion API](https://developers.notion.com/)