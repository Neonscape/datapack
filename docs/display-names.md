# 表与校验规则显示名

内部名称是稳定身份；显示名只帮助编辑器阅读。右键项目树中的表、公共校验规则文件、公共或本地定义，选择“编辑显示名”。界面显示“显示名 (内部名)”，空白输入会清除标题并回退内部名。显示名可以重复，标签页、展开状态和编辑请求仍使用文件路径与定义 ID。

元数据兼容已有表标题；表与定义标题保存在源 CUE 的 `_editor` 中：

```cue
_editor: {
    table: {title: "道具表"}       // 表文件
    schemas: {ID: {title: "稳定标识"}} // 当前文件的公共或本地 #ID
}
```

公共 Schema 文件的文件标题使用文件级 attribute（放在 package 声明前），避免多个同 package 文件的不同标题参与 CUE 字段统一：

```cue
@datapackTitle("公共规则")
package common
```

旧 `_editor.schema.title` 仍可读取；编辑文件标题时迁移到 attribute。定义标题仍按各自内部 ID 存在 `_editor.schemas`。旧数据无标题时行为不变；标题不参与 package、namespace、表/列 ID、文件路径、CUE 类型、公式、外键、CSV 头或 JSON 键。

编辑标题支持 Undo/Redo，保存后重新打开仍可读取。表改名保留标题；定义改名同时迁移 `_editor.schemas` 中的对应项；删除定义移除该项，Undo 恢复。若改名目标已有孤立显示元数据，操作拒绝，避免无提示覆盖。未知元数据会保留，清除显示名只移除 `title`，不会删除其他属性。
