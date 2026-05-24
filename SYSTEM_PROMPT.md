# Alterminal MCP Tools 使用指南

本指南用於指導 AI Agent 如何使用 Alterminal MCP 的工具來操作偽終端（PTY）。

---

## 目錄

1. [工具總覽](#1-工具總覽)
2. [核心工具詳解](#2-核心工具詳解)
3. [使用流程與最佳實踐](#3-使用流程與最佳實踐)
4. [VT100 控制序列](#4-vt100-控制序列)
5. [常見使用場景](#5-常見使用場景)
6. [故障排除](#6-故障排除)

---

## 1. 工具總覽

| 工具名稱 | 描述 | 返回值 |
|----------|------|--------|
| `screenshot` | 獲取當前終端狀態的文本截圖 | 終端內容（ASCII 格式） |
| `get_size` | 獲取終端尺寸 | `size=<cols>x<rows>` |
| `write` | 向 PTY 寫入文本 | 截圖 |
| `get_cursor` | 獲取當前光標位置 | `cursor=<col>,<row>` |
| `reset` | 重置終端到初始狀態 | 截圖 |
| `control_code` | 發送控制碼到終端 | 截圖 |
| `write_byte` | 向 PTY 寫入單個字節 | 截圖 |

---

## 2. 核心工具詳解

### 2.1 screenshot

**用途**：獲取當前終端的文本快照。

**使用時機**：
- 執行命令後查看輸出
- 確認提示符是否出現（準備接收下一個命令）
- 檢查交互式程序的輸出
- 確認操作結果

**參數**：無

**返回**：終端內容的 ASCII 文字表示

**範例**：
```json
// 調用
{
  "name": "screenshot",
  "arguments": {}
}

// 返回（示例）
"dan@mac ~ % ls -la\ntotal  128 drwx------   8 dan  staff   256 Jan 15 10:30 .\ndrwxr-xr-x   3 dan  staff    96 Jan 15 10:30 ..\n-rw-r--r--   1 dan  staff  1024 Jan 15 10:30 file.txt"
```

---

### 2.2 get_size

**用途**：獲取終端的大小（行數和列數）。

**參數**：無

**返回格式**：`size=<cols>x<rows>`
- `cols`：列數（橫向字符數）
- `rows`：行數（縱向字符數）

**範例**：
```json
// 返回
"size=80x24"
```

---

### 2.3 write

**用途**：向 PTY 寫入文本。

**參數**：
| 參數 | 類型 | 必填 | 說明 |
|------|------|------|------|
| `text` | string | ✅ | 要寫入的文本內容 |
| `with_enter` | boolean | ❌ | 是否自動按 Enter，預設 `false` |

**使用時機**：
- 執行 shell 命令
- 在交互式程序中輸入文字
- 填寫表單或提示

**注意事項**：
- 當 `with_enter: false` 時，只寫入文字不會執行
- 當 `with_enter: true` 時，會自動附加一個回車符（CR, 0x0D）

**範例**：
```json
// 執行命令
{
  "name": "write",
  "arguments": {
    "text": "ls -la",
    "with_enter": true
  }
}

// 輸入文字（不執行）
{
  "name": "write",
  "arguments": {
    "text": "Hello World",
    "with_enter": false
  }
}
```

---

### 2.4 get_cursor

**用途**：獲取當前光標位置。

**返回格式**：`cursor=<col>,<row>`
- `col`：列位置（從 1 開始）
- `row`：行位置（從 1 開始）

**參數**：無

**範例**：
```json
// 返回
"cursor=10,5"
// 表示光標在第 5 行，第 10 列
```

---

### 2.5 reset

**用途**：將終端重置為初始狀態。

**效果**：
- 清除所有輸出
- 重置光標到左上角
- 清除所有滾動區域

**參數**：無

**使用時機**：
- 終端顯示混亂需要清理
- 開始新的任務前重置環境

---

### 2.6 control_code

**用途**：發送控制碼（Control Codes）到終端。

**參數**：
| 參數 | 類型 | 必填 | 說明 |
|------|------|------|------|
| `code` | string | ✅ | 控制碼名稱 |

**可用控制碼**：
| 控制碼 | ASCII 值 | 說明 | 常用場景 |
|--------|----------|------|----------|
| `NUL` | 0x00 | 空字符 | |
| `ETX` | 0x03 | 中斷（Ctrl+C） | 終止當前程序 |
| `ENQ` | 0x05 | 詢問 | |
| `BEL` | 0x07 | 響鈴 | 發出提示音 |
| `BS` | 0x08 | 退格（Backspace） | 刪除前一個字符 |
| `HT` | 0x09 | 水平製表（Tab） | 補全或跳轉 |
| `LF` | 0x0A | 換行（Line Feed） | 換行 |
| `VT` | 0x0B | 垂直製表 | |
| `FF` | 0x0C | 換頁 | |
| `CR` | 0x0D | 回車（Carriage Return） | Enter 鍵 |
| `SO` | 0x0E | 移出 | |
| `SI` | 0x0F | 移入 | |
| `DC1` | 0x11 | 設備控制 1 | |
| `DC3` | 0x13 | 設備控制 3 | |
| `CAN` | 0x18 | 取消 | |
| `SUB` | 0x1A | 替換 | |
| `ESC` | 0x1B | 轉義（Escape） | 退出當前模式、進入命令模式 |
| `DEL` | 0x7F | 刪除字符 | 刪除當前字符 |

**範例**：
```json
// 發送 Escape 鍵（退出當前模式）
{
  "name": "control_code",
  "arguments": {
    "code": "ESC"
  }
}

// 發送 Enter 鍵
{
  "name": "control_code",
  "arguments": {
    "code": "CR"
  }
}

// 發送 Ctrl+C（中斷程序）
{
  "name": "control_code",
  "arguments": {
    "code": "ETX"
  }
}
```

---

### 2.7 write_byte

**用途**：向 PTY 寫入原始字節序列。

**參數**：
| 參數 | 類型 | 必填 | 說明 |
|------|------|------|------|
| `byte` | integer | ✅ | 單個字節的整數值（0-255） |

**使用時機**：
- 發送方向鍵、功能鍵等特殊鍵序列
- 處理特殊的控制字符
- 發送 ESC 序列的各個部分

**範例**：
```json
// 發送上箭頭（ESC [ A）
{
  "name": "write_byte",
  "arguments": {
    "byte": 27
  }
}
// 然後
{
  "name": "write_byte",
  "arguments": {
    "byte": 91
  }
}
// 然後
{
  "name": "write_byte",
  "arguments": {
    "byte": 65
  }
}
```

---

## 3. 使用流程與最佳實踐

### 3.1 基本工作流程

```
┌─────────────────────────────────────────────────────────┐
│                     開始任務                             │
└─────────────────────────┬───────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  1. screenshot ──▶ 查看當前終端狀態                       │
└─────────────────────────┬───────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  2. write ──▶ 輸入命令或文字                             │
│              (with_enter: true 執行，false 只輸入)        │
└─────────────────────────┬───────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  3. screenshot ──▶ 確認命令執行結果                       │
└─────────────────────────┬───────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  4. 重複步驟 2-3 直到任務完成                            │
└─────────────────────────────────────────────────────────┘
```

### 3.2 等待提示符

**重要原則**：在執行下一個命令前，應等待 shell 出現提示符（例如 `dan@mac ~ %` 或 `$`）。

如果沒有出現提示符，可能是：
- 命令正在執行中
- 程序在等待輸入

**判斷方法**：使用 `screenshot` 查看輸出。

### 3.3 確認操作結果

在以下情況**必須使用** `screenshot` 確認狀態：
- ✅ 執行命令後
- ✅ 輸入信息後
- ✅ 不確定當前狀態時
- ✅ 操作交互式程序時

### 3.4 控制輸入節奏

命令執行後，建議等待一段時間再截圖確認結果：

```python
# 偽代碼示例
write(text="ls -la", with_enter=True)
sleep(1)  # 等待 1 秒
screenshot()  # 確認結果
```

---

## 4. VT100 控制序列

### 4.1 方向鍵 ESC 序列

| 方向鍵 | 字節序列 |
|--------|----------|
| ↑ 上 | `[27, 91, 65]` |
| ↓ 下 | `[27, 91, 66]` |
| → 右 | `[27, 91, 67]` |
| ← 左 | `[27, 91, 68]` |

**發送方式**：使用 `write_byte` 工具依次發送每個字節。

### 4.2 功能鍵 ESC 序列

| 按鍵 | 字節序列 |
|------|----------|
| F1 | `[27, 79, 80]` |
| F2 | `[27, 79, 81]` |
| F3 | `[27, 79, 82]` |
| F4 | `[27, 79, 83]` |
| Home | `[27, 91, 72]` |
| End | `[27, 91, 70]` |
| Page Up | `[27, 91, 53, 126]` |
| Page Down | `[27, 91, 54, 126]` |
| Insert | `[27, 91, 50, 126]` |
| Delete | `[27, 91, 51, 126]` |

### 4.3 常用的 ESC 序列組合

| 功能 | ESC 序列 | 說明 |
|------|----------|------|
| 清除屏幕 | `[2J` | 清屏 |
| 清除行 | `[2K` | 清除當前行 |
| 移動到左上 | `[H` | 光標移到 (1,1) |
| 上移 n 行 | `[nA` | 光標上移 n 行 |
| 下移 n 行 | `[nB` | 光標下移 n 行 |
| 右移 n 列 | `[nC` | 光標右移 n 列 |
| 左移 n 列 | `[nD` | 光標左移 n 列 |

---

## 5. 常見使用場景

### 5.1 執行 Shell 命令

**場景**：執行 `ls -la` 查看目錄內容。

**步驟**：
1. 確認終端處於等待輸入狀態
2. 使用 `write` 輸入命令，`with_enter: true`
3. 使用 `screenshot` 確認執行結果

```json
// 步驟 1: 確認狀態
{"name": "screenshot", "arguments": {}}

// 步驟 2: 執行命令
{"name": "write", "arguments": {"text": "ls -la", "with_enter": true}}

// 步驟 3: 查看結果
{"name": "screenshot", "arguments": {}}
```

### 5.2 交互式程序輸入

**場景**：使用 `nano` 編輯文件。

**步驟**：
1. 輸入 `nano filename`
2. 在編輯器中輸入文字
3. 按 `Ctrl+O` 保存，`Ctrl+X` 退出

```json
// 啟動 nano
{"name": "write", "arguments": {"text": "nano test.txt", "with_enter": true}}

// 等待編輯器開啟後截圖確認
{"name": "screenshot", "arguments": {}}

// 輸入文字（使用 write 每次輸入一行）
{"name": "write", "arguments": {"text": "Hello World", "with_enter": true}}

// 保存 (Ctrl+O = ETX)
{"name": "control_code", "arguments": {"code": "ETX"}}

// 退出 (Ctrl+X)
{"name": "control_code", "arguments": {"code": "ETX"}}
```

### 5.3 Vim 操作

**場景**：使用 `vim` 編輯文件。

**Vim 模式切换**：
- **Normal 模式**：按 `ESC` 進入，用於移動和命令
- **Insert 模式**：按 `i` 進入，用於輸入文字
- **Command 模式**：在 Normal 模式下按 `:` 進入

**常用操作**：
```json
// 啟動 vim
{"name": "write", "arguments": {"text": "vim test.txt", "with_enter": true}}

// 進入插入模式
{"name": "write", "arguments": {"text": "i", "with_enter": false}}

// 輸入內容
{"name": "write", "arguments": {"text": "Hello Vim!", "with_enter": true}}

// 退出插入模式，回到 Normal
{"name": "control_code", "arguments": {"code": "ESC"}}

// 保存並退出
{"name": "write", "arguments": {"text": ":wq", "with_enter": true}}

// 或不保存退出
{"name": "write", "arguments": {"text": ":q!", "with_enter": true}}
```

詳細的 Vim 操作指南請參考 [VIM.md](./VIM.md)。

### 5.4 處理長文本輸出

**場景**：輸出很長，需要滾動查看。

**方法**：使用 `less` 分頁查看，或使用方向鍵滾動。

```json
// 使用 less 分頁
{"name": "write", "arguments": {"text": "cat long_output.txt | less", "with_enter": true}}

// 在 less 中向下滾動（按空格）
{"name": "write", "arguments": {"text": " ", "with_enter": false}}

// 退出 less
{"name": "write", "arguments": {"text": "q", "with_enter": false}}
```

### 5.5 使用 Tab 補全

**場景**：命令或文件名的自動補全。

```json
// 輸入部分命令
{"name": "write", "arguments": {"text": "nano tes", "with_enter": false}}

// 按 Tab 補全
{"name": "control_code", "arguments": {"code": "HT"}}

// 截圖確認補全結果
{"name": "screenshot", "arguments": {}}
```

---

## 6. 故障排除

### 6.1 問題：截圖顯示空白或不正確

**原因**：
- 命令執行需要時間
- 終端緩衝區尚未更新

**解決**：
1. 等待 1-2 秒後重新截圖
2. 檢查命令是否執行成功

### 6.2 問題：命令沒有執行

**原因**：
- 忘記設置 `with_enter: true`
- 終端忙於執行上一個命令

**解決**：
1. 檢查 `with_enter` 參數
2. 先用 `screenshot` 確認終端狀態
3. 如果終端忙，按 `Ctrl+C` (ETX) 中斷

### 6.3 問題：輸入被截斷

**原因**：
- 輸入的文本太長
- 終端宽度限制

**解決**：
拆分長命令為多個較短的 `write` 操作：

```json
// 錯誤：長命令可能被截斷
{"name": "write", "arguments": {"text": "git commit -m 'This is a very long commit message that might be truncated'", "with_enter": true}}

// 正確：拆分成多次輸入
{"name": "write", "arguments": {"text": "git commit -m 'This is a very long commit message'", "with_enter": true}}
```

### 6.4 問題：Vim 無響應

**原因**：
- 卡在某個模式中
- 等待輸入

**解決**：
1. 嘗試按 `ESC` 幾次，確保進入 Normal 模式
2. 使用 `:q!` 強制退出
3. 如果還是卡住，按 `Ctrl+C` 中斷

### 6.5 問題：特殊字符無法輸入

**原因**：
- 需要使用字節序列

**解決**：
使用 `write_byte` 發送原始字節：

```json
// 例如：輸入管道符 |
{"name": "write_byte", "arguments": {"byte": 124}}

// 或使用 write
{"name": "write", "arguments": {"text": "|", "with_enter": false}}
```

---

## 附錄：快速參考卡

### 工具調用模板

```json
// 基本截圖
{"name": "screenshot", "arguments": {}}

// 執行命令
{"name": "write", "arguments": {"text": "<命令>", "with_enter": true}}

// 輸入文字
{"name": "write", "arguments": {"text": "<文字>", "with_enter": false}}

// 發送控制鍵
{"name": "control_code", "arguments": {"code": "<控制碼>"}}

// 發送字節
{"name": "write_byte", "arguments": {"byte": <整數>}}
```

### 常用控制碼速查

| 功能 | 控制碼 | 快捷鍵 |
|------|--------|--------|
| 退出/返回 | `ESC` | Escape |
| 回車 | `CR` | Enter |
| 退格 | `BS` | Backspace |
| Tab | `HT` | Tab |
| 刪除字符 | `DEL` | Delete |
| 中斷 | `ETX` | Ctrl+C |

### 方向鍵字節序列

| 按鍵 | 字節序列 |
|------|----------|
| ↑ | `[27, 91, 65]` |
| ↓ | `[27, 91, 66]` |
| → | `[27, 91, 67]` |
| ← | `[27, 91, 68]` |

---

**最後更新**：2024年

**相關文檔**：[VIM.md](./VIM.md) - 詳細的 Vim 使用指南