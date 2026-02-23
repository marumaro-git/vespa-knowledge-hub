# 日本語検索の問題分析

## 問題1: 「はじめに」が検索で引っかからない

### 現象
- README.mdに「# はじめに」というヘッダーが実際に含まれている
- 「はじめに」で検索しても結果が0件
- 一方、「並行」で検索すると2件ヒットする

### 原因
VespaのLuceneLinguisticsデフォルト設定では、**ひらがなと漢字が異なるトークナイゼーション処理**を受けている。

#### トークナイゼーションの違い

**漢字「並行」の場合:**
```
Query parsed to: weakAnd(default contains "並行")
→ 単一の単語として検索
```

**ひらがな「はじめに」の場合:**
```
Query parsed to: weakAnd(default contains phrase("はじ", "じめ", "めに"))
→ 2文字ずつのバイグラム + フレーズ検索
```

#### 詳細な処理の流れ

1. **クエリ解析時:**
   - 「はじめに」（4文字）が入力される
   - LuceneLinguisticsのデフォルト設定でCJK Bigramトークナイザーが適用される
   - 2文字ずつに分解: "はじ", "じめ", "めに"

2. **検索処理:**
   - これらが**フレーズ検索**として処理される
   - つまり、"はじ" AND "じめ" AND "めに" が**この順序で連続して**出現する必要がある
   - `andSegmenting: true` により、さらに厳密な一致が求められる

3. **なぜヒットしないか:**
   - インデックス時とクエリ時のトークナイゼーションが微妙に異なる
   - または、フレーズ検索の条件が厳しすぎる
   - ストップワード処理や正規化の影響で、完全一致しない

### なぜ漢字は動作するのか

漢字の「並行」は：
- 単一の単語として認識される
- バイグラム化されない
- シンプルな単語検索として処理される

---

## 問題2: standardトークナイザー設定で「並行」が引っかからなくなった理由

### 変更内容

**変更前（デフォルト設定）:**
```xml
<component id="linguistics" class="com.yahoo.language.lucene.LuceneLinguistics" bundle="lucene-linguistics" />
```

**変更後（カスタム設定）:**
```xml
<component id="linguistics" class="com.yahoo.language.lucene.LuceneLinguistics" bundle="lucene-linguistics">
    <config name="com.yahoo.language.lucene.lucene-analysis">
        <analysis>
            <item key="ja">
                <tokenizer>
                    <name>standard</name>
                </tokenizer>
                <tokenFilters>
                    <item><name>cjkWidth</name></item>
                    <item><name>lowercase</name></item>
                </tokenFilters>
            </item>
        </analysis>
    </config>
</component>
```

### 原因

#### standardトークナイザーの特性
- **空白とピリオドで単語を区切る**設計
- 英語などの**空白区切り言語**を想定
- 日本語のような連続した文字列を適切に処理できない

#### 実際の動作

**「並行」のトークナイゼーション（standardトークナイザー）:**
```
Query parsed to: phrase("並", "行")
→ 1文字ずつに分解 + フレーズ検索
```

1. standardトークナイザーが日本語を1文字ずつに分解
2. フレーズ検索として処理される
3. インデックスとの一致条件が厳しすぎてヒットしない

#### デフォルト設定との違い

| 設定 | 「並行」の処理 | 結果 |
|------|--------------|------|
| デフォルト | 単一の単語として認識 | ✅ ヒットする |
| standard | "並" + "行" のフレーズ | ❌ ヒットしない |

### なぜ悪化したか

デフォルト設定では：
- 漢字は適切に単語として認識される
- 日本語特有の処理が組み込まれている

standardトークナイザーでは：
- すべての日本語文字を1文字ずつに分解
- 英語向けの処理が適用される
- 日本語検索が機能しなくなる

---

## 解決策の選択肢

### 1. デフォルト設定を維持（現状）
**メリット:**
- 漢字の検索は動作する
- 特別な設定不要

**デメリット:**
- ひらがなの検索が困難

### 2. nGramトークナイザーを使用
```xml
<item key="ja">
    <tokenizer>
        <name>nGram</name>
        <conf>
            <item key="minGramSize">2</item>
            <item key="maxGramSize">3</item>
        </conf>
    </tokenizer>
    <tokenFilters>
        <item><name>lowercase</name></item>
    </tokenFilters>
</item>
```

**メリット:**
- ひらがな・漢字・カタカナすべてで部分一致検索が可能
- 「はじめに」も「並行」も検索可能

**デメリット:**
- インデックスサイズが増加
- 完全一致の精度が下がる可能性

### 3. Kuromojiを導入（理想的だが複雑）
```xml
<item key="ja">
    <tokenizer>
        <name>japanese</name>  <!-- Kuromoji -->
    </tokenizer>
    <tokenFilters>
        <item><name>japaneseBaseForm</name></item>
        <item><name>japanesePartOfSpeechStop</name></item>
        <item><name>lowercase</name></item>
    </tokenFilters>
</item>
```

**問題点:**
- セルフホストVespaにはKuromojiが含まれていない
- カスタムMavenビルドが必要
- 複雑な設定とデプロイプロセス

---

## 推奨アプローチ

現時点では**nGramトークナイザー**が最も実用的：
- 追加の依存関係不要
- ひらがな・漢字の両方で検索可能
- 設定がシンプル

ただし、本格的な日本語検索システムを構築する場合は、将来的にKuromojiの導入を検討すべき。

---

## nGramトークナイザー導入の試み

### 実施した設定

**services.xml:**
```xml
<component id="linguistics" class="com.yahoo.language.lucene.LuceneLinguistics" bundle="lucene-linguistics">
    <config name="com.yahoo.language.lucene.lucene-analysis">
        <analysis>
            <item key="unknown">
                <tokenizer>
                    <name>nGram</name>
                    <conf>
                        <item key="minGramSize">2</item>
                        <item key="maxGramSize">3</item>
                    </conf>
                </tokenizer>
                <tokenFilters>
                    <item><name>lowercase</name></item>
                </tokenFilters>
            </item>
            <item key="ja">
                <!-- 同様のnGram設定 -->
            </item>
            <item key="en">
                <!-- 同様のnGram設定 -->
            </item>
        </analysis>
    </config>
</component>
```

### 結果

**クエリ時のトークナイゼーション: ✅ 成功**
```
「はじめに」 → phrase("はじ", "はじめ", "じめ", "じめに", "めに")
```
2-3文字のn-gramに正しく分解されている。

**インデックス時のトークナイゼーション: ❌ 失敗**
- スキーマで言語を指定していないため、デフォルトのトークナイザーが使用される
- nGram設定が適用されない

**検索結果: ❌ ヒットしない**
- インデックス時とクエリ時のトークナイゼーションが不一致
- 結果として検索できない

### 問題の根本原因

**Vespaのトークナイゼーション処理は2段階:**

1. **インデックス時**: データを登録する際のトークナイゼーション
   - スキーマ（.sd）のフィールド設定で制御される
   - `language`を明示的に指定しないと、デフォルトの言語検出 + デフォルトアナライザーが使用される

2. **クエリ時**: 検索時のトークナイゼーション
   - services.xmlのLuceneLinguistics設定で制御される
   - 正しく設定されている（nGramが適用されている）

**両方が一致しないと検索が機能しない！**

### なぜ「並行」は検索できるのか

漢字「並行」の場合：
- インデックス時: デフォルトアナライザーで単一の単語として認識
- クエリ時: 同様に単一の単語として認識
- 一致するため、検索成功 ✅

ひらがな「はじめに」の場合：
- インデックス時: デフォルトアナライザー（CJK Bigram）で処理
- クエリ時: nGramで処理（設定済み）
- 不一致のため、検索失敗 ❌

### 混合言語（英語+日本語）の課題

README.mdのような混合言語ドキュメント：
```markdown
# はじめに
Goの並行処理を学ぶためのサンプルコード
```

- Vespaの言語検出は**フィールド単位**
- フィールド全体が一つの言語として判定される
- `language: ja`を指定すると、英語部分が適切に処理されない可能性

### 解決策

#### オプション1: スキーマで言語を指定（最も簡単）

**vespa/schemas/knowledge_item.sd:**
```sd
field title type string {
    indexing: summary | index
    index: enable-bm25
    match {
        language: unknown  # または ja
    }
}

field content type string {
    indexing: summary | index
    index: enable-bm25
    match {
        language: unknown  # または ja
    }
}
```

**メリット:**
- services.xmlのnGram設定が適用される
- 「はじめに」も検索可能になる

**デメリット:**
- 混合言語ドキュメントで、一方の言語が適切に処理されない可能性
- `language: ja`の場合、英語検索が劣化する可能性

#### オプション2: Kuromojiを導入（理想的だが複雑）

完全な日本語形態素解析を実現。詳細は次のセクションを参照。

---

## Kuromoji導入方法

Kuromojiは日本語形態素解析ライブラリで、Vespaで本格的な日本語検索を実現するための最適な選択肢です。

### Kuromojiの利点

- **高精度な形態素解析**: 「はじめに」→「はじめ」+「に」と適切に分解
- **品詞フィルタリング**: 助詞や助動詞を除外可能
- **基本形変換**: 「走っている」→「走る」
- **日本語ストップワード**: 「です」「ます」などを除外
- **英語との混在対応**: 日本語部分のみ形態素解析、英語は通常処理

### セルフホストVespaでの課題

**問題点:**
- `vespaengine/vespa:8`イメージには**Kuromojiが含まれていない**
- `lucene-analysis-kuromoji`のJARがクラスパスに存在しない

エラーメッセージ:
```
A SPI class of type org.apache.lucene.analysis.TokenizerFactory with name 'japanese'
does not exist. You need to add the corresponding JAR file supporting this SPI to your classpath.
```

### 導入手順

Kuromojiをセルフホスト環境で使用するには、カスタムDockerイメージを作成する必要があります。

#### ステップ1: カスタムDockerイメージの作成

**Dockerfile:**
```dockerfile
FROM vespaengine/vespa:8

# Kuromoji JARをダウンロード
USER root
RUN curl -L https://repo1.maven.org/maven2/org/apache/lucene/lucene-analysis-kuromoji/9.11.1/lucene-analysis-kuromoji-9.11.1.jar \
    -o /opt/vespa/lib/jars/lucene-analysis-kuromoji-9.11.1.jar

USER vespa
```

**ビルドとタグ付け:**
```bash
docker build -t vespa-kuromoji:8 .
```

#### ステップ2: docker-compose.ymlの更新

```yaml
services:
  vespa:
    image: vespa-kuromoji:8  # カスタムイメージを使用
    container_name: vespa
    hostname: localhost
    privileged: true
    volumes:
      - vespa-data:/opt/vespa/var
    ports:
      - "8080:8080"
      - "19071:19071"
      - "19092:19092"
      - "19050:19050"

volumes:
  vespa-data:
```

#### ステップ3: services.xmlの設定

```xml
<component id="linguistics" class="com.yahoo.language.lucene.LuceneLinguistics" bundle="lucene-linguistics">
    <config name="com.yahoo.language.lucene.lucene-analysis">
        <analysis>
            <item key="ja">
                <tokenizer>
                    <name>japanese</name>  <!-- Kuromojiトークナイザー -->
                </tokenizer>
                <tokenFilters>
                    <item><name>japaneseBaseForm</name></item>          <!-- 基本形変換 -->
                    <item><name>japanesePartOfSpeechStop</name></item>  <!-- 品詞フィルタ -->
                    <item><name>japaneseStop</name></item>              <!-- 日本語ストップワード -->
                    <item><name>lowercase</name></item>                 <!-- 小文字化 -->
                </tokenFilters>
            </item>
            <item key="en">
                <tokenizer>
                    <name>standard</name>
                </tokenizer>
                <tokenFilters>
                    <item><name>lowercase</name></item>
                    <item><name>stop</name></item>
                </tokenFilters>
            </item>
        </analysis>
    </config>
</component>
```

#### ステップ4: スキーマの設定

**混合言語対応の場合:**
```sd
field title type string {
    indexing: summary | index
    index: enable-bm25
    match {
        language: ja  # 日本語として処理
    }
}

field content type string {
    indexing: summary | index
    index: enable-bm25
    match {
        language: ja  # 日本語として処理
    }
}
```

**または、言語を指定せず自動検出に任せる:**
```sd
field title type string {
    indexing: summary | index
    index: enable-bm25
    # language指定なし = 自動検出
}
```

#### ステップ5: デプロイと再インデックス

```bash
# Vespaを再起動
docker-compose down
docker-compose up -d

# アプリケーションをデプロイ
task vespa:deploy

# データを再インデックス
task index:run
```

### Kuromoji導入後の動作

**「はじめに」の処理:**
```
インデックス時: 「はじめ」（動詞基本形） + 「に」（助詞、除外される可能性）
クエリ時: 同様に処理
結果: ✅ ヒットする
```

**「並行処理」の処理:**
```
インデックス時: 「並行」（名詞） + 「処理」（名詞）
クエリ時: 同様に処理
結果: ✅ 「並行」でも「処理」でもヒットする
```

**英語との混在:**
- 日本語部分: Kuromojiで形態素解析
- 英語部分: スペースで区切られた単語として処理
- 両方とも適切に検索可能

### Vespa Cloudを使用する場合

Vespa Cloudでは、Kuromojiが標準で利用可能です。カスタムDockerイメージは不要で、services.xmlの設定のみで使用できます。

**参考リンク:**
- [Vespa Cloud](https://cloud.vespa.ai/)
- [Lucene Kuromoji Documentation](https://lucene.apache.org/core/9_11_1/analyzers-kuromoji/overview-summary.html)

---

## まとめ

### 現状の選択肢

| アプローチ | 難易度 | 日本語検索 | 英語検索 | 混合言語 |
|-----------|--------|-----------|---------|---------|
| デフォルト設定 | ⭐ | △（漢字のみ） | ⭕ | ⭕ |
| nGram + スキーマ修正 | ⭐⭐ | ⭕ | △ | △ |
| Kuromoji（カスタムイメージ） | ⭐⭐⭐⭐ | ⭕⭕ | ⭕ | ⭕ |
| Vespa Cloud + Kuromoji | ⭐⭐ | ⭕⭕ | ⭕ | ⭕ |

### 推奨される進め方

**短期的（すぐに使いたい）:**
1. ~~スキーマに`language: unknown`または`language: ja`を追加~~ ❌ 失敗
   - スキーマの`match`ブロック内では`language`キーワードは使用できない
   - `linguistics`ブロックを使う方法もあるが、複雑
2. **現状のデフォルト設定を受け入れる**
   - 漢字（「並行」「処理」など）は検索可能 ✅
   - ひらがな（「はじめに」など）は検索困難 ❌
   - 英語は正常に検索可能 ✅

**長期的（本格的な日本語検索）:**
1. **カスタムDockerイメージでKuromojiを導入**（推奨） ⭐
2. または、**Vespa Cloudへの移行を検討**
3. 高精度な日本語形態素解析を実現

---

## nGram導入の最終結果

### 試行錯誤の記録

**試みた方法:**
1. ✅ services.xmlにnGram設定を追加（unknown, ja, en）
2. ❌ スキーマで`language: unknown`を指定 → 構文エラー
3. ❌ `linguistics`ブロックの使用 → 複雑で断念

**結論:**
- services.xmlの設定だけでは不十分
- スキーマでの言語指定が必要だが、シンプルな方法が見つからない
- **Kuromojiの導入が最も確実な解決策**

### 現在の検索動作（デフォルト設定）

| キーワード | 結果 | 理由 |
|-----------|------|------|
| 並行 | ✅ ヒットする（2件） | 漢字は単語として認識される |
| はじめに | ❌ ヒットしない | バイグラム+フレーズ検索で厳しすぎる |
| はじ | ❌ ヒットしない | 部分一致が機能しない |
| Go | ✅ ヒットする | 英語は正常に動作 |
| function | ✅ ヒットする | 英語は正常に動作 |

### 実用的な回避策

**現時点で日本語検索を改善する方法:**

1. **漢字を使用する**
   - 「はじめに」→「初め」や「開始」と言い換えて検索
   - 文書内に漢字表記がある場合は検索可能

2. **英語キーワードを使用**
   - 技術文書の場合、英語の技術用語で検索
   - 例: 「並行処理」→「concurrency」

3. **複数キーワードで検索**
   - 「並行 処理」のように複数の漢字キーワードを組み合わせる

4. **Kuromoji導入を計画**
   - カスタムDockerイメージの作成（上記参照）
   - Vespa Cloudへの移行検討

---

## 学んだこと

### Vespaの言語処理の仕組み

1. **2段階のトークナイゼーション**
   - インデックス時: スキーマで制御
   - クエリ時: services.xmlで制御
   - **両方が一致しないと検索が機能しない**

2. **言語検出の動作**
   - フィールド単位で自動検出される
   - 混合言語ドキュメントでは一つの言語として判定される
   - 明示的に言語を指定するにはスキーマ設定が必要（が複雑）

3. **デフォルト動作の特性**
   - 漢字: 単語として認識され、検索可能
   - ひらがな: CJK Bigramで処理され、フレーズ検索となり厳しい
   - 英語: 空白区切りで正常に動作

4. **Kuromojiの必要性**
   - セルフホスト環境では追加セットアップが必要
   - しかし、日本語検索を実現する最も確実な方法
   - 混合言語ドキュメントにも対応可能

**Sources:**
- [Vespa Newsletter, February 2026](https://blog.vespa.ai/vespa-newsletter-february-2026/)
- [Schema reference](https://docs.vespa.ai/en/reference/schemas/schemas.html)
- [Lucene Linguistics](https://docs.vespa.ai/en/linguistics/lucene-linguistics.html)
