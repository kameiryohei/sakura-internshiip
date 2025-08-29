---
marp: true
title: NetHygiene - ネットワークセキュリティ可視化ツール
paginate: true
size: 16:9
theme: sakura
style: |
  table {
    font-size: 70%;
  }

  .small {
    font-size: 70%;
  }

  h2 {
    border-left: 6px solid #ff5577;
    padding-left: 0.7em;
    background: none;
    color: inherit;
    margin-top: 1.5em;
    margin-bottom: 0.7em;
  }

  .label {
    background: #7799dd;
    color: #fff;
    border-radius: 4px;
    padding: 2px 8px;
    font-size: 80%;
  }

  .label2 {
    background: #ff5577;
    color: #fff;
    border-radius: 4px;
    padding: 2px 8px;
    font-size: 60%;
  }

  .highlight {
    background: #ffeb3b;
    padding: 2px 6px;
    border-radius: 3px;
    font-weight: bold;
  }

  .safe {
    color: #4caf50;
    font-weight: bold;
  }

  .warning {
    color: #ff9800;
    font-weight: bold;
  }

  .danger {
    color: #f44336;
    font-weight: bold;
  }
---

<!-- _class: pink lead -->

# NetHygiene

## 誰でも使えるネットワークセキュリティ可視化ツール

MVP Development Team

---

## なぜ NetHygiene が必要か？

### 🔍 現状の課題

- デジタル化の進展で**ネットワーク機器が急増**
- サイバー犯罪の**増加と巧妙化**
- IT 専門知識がない人でもネットワークを管理する必要性
- 既存システムを変更せずにリスクを把握したい

### 💡 私たちの解決策

- <span class="highlight">専門知識不要</span>で使える監視ツール
- 色分けによる<span class="highlight">直感的な危険度表示</span>
- 既存ネットワークへの<span class="highlight">影響ゼロ</span>
- お店・オフィス・家庭など幅広い場面で活用可能

---

## 主要機能

### 🖥️ ネットワーク機器の自動検出

- LAN 内の全機器を**自動スキャン** <span class="label">10 分ごと</span>
- IP アドレス・MAC アドレス・メーカー情報を取得

### 🎨 視覚的な危険度表示

- <span class="safe">● 安全な機器（緑）</span>
- <span class="warning">● 不明な機器（黄）</span>
- <span class="danger">● 不審な機器（赤）</span>

### ⚠️ リアルタイム警告機能

- 不正な DHCP サーバーの検出
- 異常なブロードキャスト/マルチキャストトラフィックの監視

---

## システム構成

### アーキテクチャ

```
[NetHygiene エージェント] ---(HTTPS/Basic認証)---> [AppRun ダッシュボード]
     ↓                                                    ↓
  ARPスキャン                                         データ表示
  データ送信                                          DB更新
```

### 技術スタック

| コンポーネント | 技術                                             | 用途                      |
| :------------- | :----------------------------------------------- | :------------------------ |
| **NetHygiene** | ShellScript + arp-scan                           | LAN スキャン・データ送信  |
| **AppRun**     | Go + SQLite3                                     | HTTP サーバー・データ管理 |
| **インフラ**   | Ubuntu 24.04 + さくらのクラウド                  | VM 環境・仮想 SW          |
| **CI/CD**      | GitHub Actions                                   | 自動デプロイ              |
| **その他**     | Terraform <span class="label2">チャレンジ</span> | インフラプロビジョニング  |

---

## 今後の展望とターゲットユーザー

### 🚀 展開予定

- **Raspberry Pi 対応** - 小型デバイスでの運用
- **自動設定機能** - cloud-config.yaml / PXE ブート対応
- **セキュリティ強化** - トークン管理・HTTPS 暗号化の改善

### 🎯 ターゲットユーザー

- 小規模店舗の経営者
- オフィスの事務担当者
- 家庭内ネットワークが心配な方
- IT 専門知識がない管理者

<br>

<div style="text-align: center; background: #e3f2fd; padding: 20px; border-radius: 10px;">
<strong>NetHygieneで、誰もが安心できるネットワーク環境を実現</strong>
</div>
