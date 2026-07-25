# 编辑器设置格式 v1

编辑器设置是用户级 JSON 文件，不属于任何数据工程，也不进入版本控制。

程序通过 Go 的 `os.UserConfigDir()` 取得配置根目录，并使用：

```text
<用户配置目录>/datapack/settings.json
```

Linux 的默认路径为：

```text
~/.config/datapack/settings.json
```

首次启动时，文件不存在则使用默认值；用户修改设置后由编辑器创建或更新该文件。

## 完整示例

```json
{
  "version": 1,
  "build": {
    "mode": "auto",
    "autoCompileDebounceSeconds": 10
  }
}
```

## 字段

| 字段 | 必填 | 说明 |
|---|---:|---|
| `version` | 是 | 编辑器设置格式版本。v1 固定为 `1`。 |
| `build.mode` | 是 | `auto` 或 `manual`。 |
| `build.autoCompileDebounceSeconds` | `auto` 时是 | 自动编译等待秒数；默认 `10`。 |

## 编译模式

- `auto`：编辑保存后等待设定秒数，再对受影响内容增量编译。
- `manual`：编辑保存后只标记工程待编译；用户点击编译时执行完整编译。
- 无论当前模式为何，导出前始终执行完整编译。

当不同工程同时打开时，它们使用同一份用户级编辑器设置。修改设置后，编辑器对后续编辑事务立即采用新值。
