# skype-history-viewer-cli

一個用於查看和搜尋 Skype 聊天記錄的命令列工具，支援從匯出的 JSON 檔案讀取。

[![Test Status](https://github.com/beckxie/skype-history-viewer-cli/actions/workflows/test.yml/badge.svg)](https://github.com/beckxie/skype-history-viewer-cli/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/beckxie/skype-history-viewer-cli/graph/badge.svg?token=...)](https://codecov.io/gh/beckxie/skype-history-viewer-cli)
[![Go Report Card](https://goreportcard.com/badge/github.com/beckxie/skype-history-viewer-cli)](https://goreportcard.com/report/github.com/beckxie/skype-history-viewer-cli)
[![Go Version](https://img.shields.io/github/go-mod/go-version/beckxie/skype-history-viewer-cli)](https://github.com/beckxie/skype-history-viewer-cli/blob/main/go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

[English](README.md) | **繁體中文**

## 功能特色

- 📱 **命令列介面**：易於使用的 CLI 工具，支援多個子命令
- 🔍 **進階搜尋**：透過各種過濾條件搜尋訊息
- 📊 **統計資訊**：查看聊天記錄的詳細統計數據
- 💬 **對話檢視器**：使用分頁功能瀏覽對話內容
- 📎 **匯出功能**：將單個對話匯出為 JSON 格式
- 🎨 **彩色輸出**：美觀的終端機彩色輸出
- ⚡ **效能優化**：針對大型聊天記錄進行優化，並提供進度指示器

## 安裝

### 從原始碼安裝

```bash
git clone https://github.com/beckxie/skype-history-viewer-cli.git
cd skype-history-viewer-cli
go build -o skype-viewer
```

### 使用 Go Install

```bash
go install github.com/beckxie/skype-history-viewer-cli@latest
```

## 使用方法

### 基本用法

```bash
# 顯示說明
skype-viewer --help

# 列出所有對話
skype-viewer list -f /path/to/messages.json

# 查看特定對話
skype-viewer view 1 -f /path/to/messages.json

# 搜尋訊息
skype-viewer search -q "搜尋關鍵字" -f /path/to/messages.json

# 顯示統計資訊
skype-viewer stats -f /path/to/messages.json

# 匯出對話
skype-viewer export 1 -f /path/to/messages.json -o output.json
```

### 命令說明

#### `list` - 列出所有對話

```bash
skype-viewer list -f messages.json [flags]

Flags:
  --show-system    包含系統訊息在計數中
```

#### `view` - 查看對話訊息

```bash
skype-viewer view [對話編號] -f messages.json [flags]

Flags:
  --page-size int       每頁顯示的訊息數量 (預設 20)
  --newest-first        依最新訊息優先排序
  --show-system         顯示系統訊息
  --date-from string    篩選此日期之後的訊息 (YYYY-MM-DD)
  --date-to string      篩選此日期之前的訊息 (YYYY-MM-DD)
```

#### `search` - 搜尋訊息

```bash
skype-viewer search -q "查詢內容" -f messages.json [flags]

Flags:
  -q, --query string         搜尋查詢文字 (必要)
  --content                  在訊息內容中搜尋 (預設 true)
  --sender                   在發送者名稱中搜尋 (預設 true)
  --case-sensitive           區分大小寫搜尋
  --conversation string      依對話名稱篩選
  --limit int                最大結果數量 (預設 50)
  --date-from string         搜尋此日期之後的訊息 (YYYY-MM-DD)
  --date-to string           搜尋此日期之前的訊息 (YYYY-MM-DD)
```

#### `export` - 匯出對話

```bash
skype-viewer export [對話編號] -f messages.json [flags]

Flags:
  -o, --output string    輸出檔案路徑 (預設: 自動產生)
```

#### `stats` - 顯示統計資訊

```bash
skype-viewer stats -f messages.json
```

#### `convert` - 轉換舊版匯出格式

```bash
skype-viewer convert [舊版匯出檔案]
```

將舊格式的 JSON 檔案轉換為可被所有命令讀取的新格式。

### 全域選項

```bash
-f, --file string    Skype 匯出 JSON 檔案或目錄的路徑
-v, --verbose        啟用詳細輸出
```

## 匯出 Skype 資料

要匯出您的 Skype 聊天記錄：

1. 造訪 [Skype 匯出支援頁面](https://support.microsoft.com/zh-tw/skype/how-do-i-export-or-delete-my-skype-data-84546e00-2fef-4c45-8ef6-3a27f83242cc)
2. 使用您的 Microsoft 帳號登入
3. 請求匯出您的資料
4. 準備好時下載 `messages.json` 檔案

## 使用範例

### 搜尋特定人員的訊息

```bash
skype-viewer search -q "John" --sender -f messages.json
```

### 使用日期篩選查看對話

```bash
skype-viewer view 1 -f messages.json --date-from 2024-01-01 --date-to 2024-12-31
```

### 使用自訂輸出匯出對話

```bash
skype-viewer export 5 -f messages.json -o "john_doe_chat.json"
```

### 互動式對話檢視

```bash
# 不指定對話編號時，進入互動模式
skype-viewer view -f messages.json
```

## 詳細功能

- **分頁功能**：大型對話會分頁顯示，方便瀏覽
- **日期篩選**：可依日期範圍篩選訊息
- **系統訊息**：可選擇顯示或隱藏系統訊息
- **進度指示器**：載入檔案和搜尋時會顯示進度
- **快取機制**：搜尋結果會被快取以加快重複搜尋
- **Unicode 支援**：正確處理表情符號和特殊字元

## 系統需求

- Go 1.25.5 或更新版本
- 支援彩色輸出的終端機（以獲得最佳體驗）

## 授權條款

MIT License
