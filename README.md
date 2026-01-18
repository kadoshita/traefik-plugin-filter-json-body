# traefik-plugin-filter-json-body
JSONのbodyの内容を元にフィルタリングを行うプラグイン

## 概要
リクエストのJSONボディの内容を検査し、指定した条件に一致する場合にリクエストを拒否する。

## 設定

### パラメータ
- `rules`: フィルタリングルールの配列
  - `path`: リクエストパス（完全一致）
  - `method`: HTTPメソッド（完全一致）
  - `bodyPath`: JSONボディ内の検査対象パス（XPath形式）
  - `bodyValueCondition`: 値の一致条件（正規表現）

### 動作
- すべての条件に一致するリクエストは403 Forbiddenを返す
- 条件に一致しないリクエストは次のハンドラに渡される
- Content-Typeがapplication/jsonまたはapplication/*+json以外の場合は検査をスキップする
- ボディサイズの上限は10MB

## 設定例

### 例1: 特定の文字列値を拒否
```yaml
rules:
  - path: /api/test
    method: POST
    bodyPath: key
    bodyValueCondition: ^value$
```

### 例2: ネストされたオブジェクトの値を検査
```yaml
rules:
  - path: /api/test
    method: POST
    bodyPath: //nestedObject/innerString
    bodyValueCondition: ^inner$
```

### 例3: 配列内のオブジェクトの値を検査
```yaml
rules:
  - path: /api/test
    method: POST
    bodyPath: //arrayOfObjects/*/objString[text()='obj2']
    bodyValueCondition: ^obj2$
```
