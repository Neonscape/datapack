# Basic RPG Auto

此工程展示公共类型/约束 Schema、本地结构体和数组外键。自动或手动编译由用户级编辑器设置决定，不保存在该工程中。

- `rpg.effect`：效果表。
- `rpg.item`：道具表，使用数组外键引用效果表。
- `common.#GameID` 和 `common.#Rarity`：公共类型 Schema。
- `common.#NonNegative`：公共约束 Schema。
- `_local.#Position`：仅供道具表使用的本地类型 Schema。

`effectIds` 中的每个值必须存在于 `rpg.effect` 的主键中。它只是外键约束，不能用于公式读取效果表数据。
