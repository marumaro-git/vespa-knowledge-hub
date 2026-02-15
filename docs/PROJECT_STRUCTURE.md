# Project Structure

```
vespa-knowledge-hub/
│
├── README.md                     # ユーザー向けドキュメント
├── CLAUDE.md                     # 開発ガイド（AI用）
├── Makefile                      # よく使うコマンドのショートカット
├── .gitignore                    # Git除外設定
├── .env.example                  # 環境変数のサンプル
│
├── .skills/                      # 開発ベストプラクティス
│   ├── github-indexer.md        # GitHub API連携のスキル
│   └── vespa-schema.md          # Vespaスキーマ設計のスキル
│
├── vespa/                        # Vespa検索エンジン設定
│   ├── docker-compose.yml       # Docker設定
│   ├── services.xml             # Vespaサービス定義
│   ├── deploy.sh                # デプロイスクリプト
│   └── schemas/
│       └── knowledge_item.sd    # 検索スキーマ定義
│
├── backend/                      # Goバックエンド
│   ├── README.md                # バックエンド実装ガイド
│   ├── go.mod                   # (作成予定)
│   ├── go.sum                   # (作成予定)
│   ├── cmd/
│   │   ├── indexer/             # GitHubデータ投入CLI
│   │   │   └── main.go          # (作成予定)
│   │   └── api/                 # 検索APIサーバー
│   │       └── main.go          # (作成予定)
│   └── internal/
│       ├── github/              # GitHub APIクライアント
│       │   ├── client.go        # (作成予定)
│       │   └── models.go        # (作成予定)
│       ├── vespa/               # Vespaクライアント
│       │   ├── client.go        # (作成予定)
│       │   ├── search.go        # (作成予定)
│       │   └── models.go        # (作成予定)
│       └── models/              # 共通データモデル
│           └── document.go      # (作成予定)
│
└── frontend/                     # React フロントエンド
    ├── README.md                # フロントエンド実装ガイド
    ├── package.json             # (作成予定)
    ├── vite.config.js           # (作成予定)
    ├── index.html               # (作成予定)
    └── src/
        ├── App.jsx              # (作成予定)
        ├── main.jsx             # (作成予定)
        ├── components/          # (作成予定)
        ├── hooks/               # (作成予定)
        └── utils/               # (作成予定)
```

## 現在の状態

✅ **完了**
- プロジェクト構成ドキュメント
- Vespa設定ファイル（スキーマ、services.xml、Docker設定）
- 開発スキルドキュメント（GitHub Indexer、Vespaスキーマ）
- Makefile（開発タスク自動化）

🚧 **次のステップ**
- Goバックエンドの実装
- React フロントエンドの実装
- インテグレーションテスト

## Quick Start

```bash
# 1. 環境変数を設定
cp .env.example .env
# .env を編集してGITHUB_TOKENなどを設定

# 2. 開発環境をセットアップ
make dev-setup

# 3. データをインデックス
make index

# 4. APIサーバーとフロントエンドを起動（別々のターミナルで）
make backend-run
make frontend-run
```

詳細は各ディレクトリのREADME.mdを参照してください。