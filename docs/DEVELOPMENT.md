# Vespa Knowledge Hub - Development Guide

## プロジェクト概要

VespaとGoを使用して、GitHubリポジトリとNotion（将来追加予定）を横断検索できる個人用ナレッジベースシステムです。

### 目的
- 自分のコードとドキュメントを一元検索
- ハイブリッド検索（キーワード + セマンティック）による高精度な検索
- 段階的開発：Phase 1 (GitHub) → Phase 2 (Notion統合)

### 技術スタック
- **検索エンジン**: Vespa (Docker)
- **バックエンド**: Go 1.21+
- **フロントエンド**: React + Vite
- **データソース**: GitHub API (Phase 1), Notion API (Phase 2)

## アーキテクチャ

```
┌─────────────┐
│  Frontend   │ React + Vite
│  (Port 5173)│
└──────┬──────┘
       │ HTTP
       ▼
┌─────────────┐
│  Backend    │ Go API Server
│  (Port 3000)│
└──────┬──────┘
       │ HTTP
       ▼
┌─────────────┐
│   Vespa     │ Search Engine
│  (Port 8080)│
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Indexer    │ Go CLI (GitHub → Vespa)
└─────────────┘
```

## ディレクトリ構成

```
vespa-knowledge-hub/
├── CLAUDE.md              # このファイル
├── README.md              # ユーザー向けドキュメント
├── .skills/
│   ├── github-indexer.md  # GitHub データ取得のベストプラクティス
│   └── vespa-schema.md    # Vespaスキーマ設計ガイド
├── vespa/
│   ├── schemas/
│   │   └── knowledge_item.sd
│   ├── services.xml
│   └── docker-compose.yml
├── backend/
│   ├── cmd/
│   │   ├── indexer/       # データ投入CLI
│   │   │   └── main.go
│   │   └── api/           # 検索API
│   │       └── main.go
│   ├── internal/
│   │   ├── github/        # GitHub API client
│   │   │   ├── client.go
│   │   │   └── models.go
│   │   ├── vespa/         # Vespa client
│   │   │   ├── client.go
│   │   │   ├── search.go
│   │   │   └── models.go
│   │   └── models/        # 共通データモデル
│   │       └── document.go
│   ├── go.mod
│   └── go.sum
└── frontend/
    ├── src/
    │   ├── App.jsx
    │   ├── components/
    │   │   ├── SearchBar.jsx
    │   │   └── ResultList.jsx
    │   └── main.jsx
    ├── package.json
    └── vite.config.js
```

## Phase 1: GitHub検索システム

### データフロー

1. **インデックス作成**
   ```
   GitHub API → Indexer (Go) → Vespa
   ```
   - GitHub APIでリポジトリのコードを取得
   - ファイル内容を解析してVespaドキュメントに変換
   - Vespa Document APIで投入

2. **検索**
   ```
   User → Frontend → Backend API → Vespa → Backend → Frontend → User
   ```
   - ユーザーがクエリ入力
   - Vespa Query APIで検索
   - 結果をフォーマットして返却

### Vespaスキーマ設計

**knowledge_item** ドキュメントタイプ：

| フィールド名 | 型 | 用途 | インデックス |
|------------|-----|------|-------------|
| id | string | ユニークID | attribute, summary |
| title | string | ファイル名/タイトル | index, summary |
| content | string | コード/テキスト本文 | index, summary |
| source_type | string | ソース種別 (github_code, notion_page等) | attribute |
| source_url | string | 元URLへのリンク | summary |
| repo_name | string | GitHubリポジトリ名 | attribute |
| file_path | string | ファイルパス | attribute |
| language | string | プログラミング言語 | attribute |
| created_at | long | 作成日時 (Unix timestamp) | attribute |
| updated_at | long | 更新日時 (Unix timestamp) | attribute |

**Phase 2でNotionを追加する際の拡張フィールド:**
- notion_database: string
- notion_workspace: string
- tags: array<string>

### ランキング戦略

Phase 1では **BM25** ベースのテキストマッチング：

```
rank-profile default {
    first-phase {
        expression: bm25(title) + bm25(content)
    }
}
```

Phase 2で追加予定：
- セマンティック検索（embedding フィールド追加）
- ハイブリッドランキング（BM25 + ベクトル類似度）

## 開発の進め方

### ステップ1: Vespaセットアップ

```bash
cd vespa
docker-compose up -d

# ヘルスチェック
curl http://localhost:8080/state/v1/health

# スキーマデプロイ
# 初回はvespa-cliをインストール: brew install vespa-cli
vespa config set target local
vespa deploy --wait 300
```

### ステップ2: Goバックエンド開発

```bash
cd backend

# 依存関係インストール
go mod init github.com/yourusername/vespa-knowledge-hub
go get github.com/google/go-github/v57/github
go get golang.org/x/oauth2

# 開発中のテスト実行
go test ./...
```

### ステップ3: GitHub Indexer実装

必要な環境変数：
- `GITHUB_TOKEN`: Personal Access Token (repo 権限)
- `VESPA_URL`: Vespaエンドポイント (通常 http://localhost:8080)
- `TARGET_REPOS`: インデックス対象リポジトリ (カンマ区切り)

実行例：
```bash
export GITHUB_TOKEN=ghp_xxxxxxxxxxxx
export VESPA_URL=http://localhost:8080
export TARGET_REPOS=owner/repo1,owner/repo2

go run cmd/indexer/main.go
```

### ステップ4: API開発

```bash
cd backend
go run cmd/api/main.go
# Port 3000 で起動

# テスト
curl "http://localhost:3000/api/search?q=authentication"
```

### ステップ5: フロントエンド開発

```bash
cd frontend
npm install
npm run dev
# Port 5173 で起動
```

## API仕様

### 検索エンドポイント

**GET /api/search**

クエリパラメータ：
- `q` (required): 検索クエリ
- `source_type` (optional): フィルタ (github_code, github_issue 等)
- `repo` (optional): リポジトリ名でフィルタ
- `language` (optional): プログラミング言語でフィルタ
- `limit` (optional): 結果数 (デフォルト: 20)

レスポンス例：
```json
{
  "total_count": 42,
  "hits": [
    {
      "id": "gh_code_12345",
      "fields": {
        "title": "auth/middleware.go",
        "content": "package auth\n\nfunc AuthMiddleware...",
        "source_type": "github_code",
        "source_url": "https://github.com/owner/repo/blob/main/auth/middleware.go",
        "repo_name": "owner/repo",
        "file_path": "auth/middleware.go",
        "language": "go"
      },
      "relevance": 0.85
    }
  ]
}
```

## トラブルシューティング

### Vespaが起動しない
- Dockerのメモリ設定を確認（推奨: 4GB以上）
- ポート競合を確認 (8080, 19071)

### インデックスが遅い
- GitHub API Rate Limitに注意（5000 requests/hour）
- 大きいファイルはスキップする実装を追加
- バッチサイズを調整

### 検索結果が出ない
- Vespaのクエリログを確認: `docker logs vespa`
- スキーマが正しくデプロイされているか確認
- ドキュメントが投入されているか確認: 
  ```bash
  curl http://localhost:8080/document/v1/default/knowledge_item/docid/test_id
  ```

## Phase 2への移行（Notion統合）

Phase 1完成後、以下を追加：

1. **Notion APIクライアント** (`internal/notion/`)
2. **スキーマ拡張** (notion関連フィールド追加)
3. **Notion Indexer** (`cmd/notion-indexer/`)
4. **マルチソース検索UI** (ソース種別のフィルタリング)

詳細は `.skills/notion-integration.md` (Phase 2で作成予定) を参照。

## コントリビューション

個人プロジェクトですが、改善提案は歓迎します。

## ライセンス

MIT