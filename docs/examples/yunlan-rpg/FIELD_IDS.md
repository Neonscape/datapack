# 字段 ID、中文显示名与 Schema 字典

**英文 ID 是程序接口；中文显示名只供编辑器阅读。CSV 列头使用下列 ID。** 修改显示名不改公式、FK、导出列或行数据。坐标结构的 JSON 成员为 `x/y`，分别表示横/纵。

## 效果：战斗机制 · `content.effect`

| 英文列 ID / CSV 列头 | 中文显示名 | Schema / 导出 |
|---|---|---|
| `name_text_id` | 名称文本ID | `common.#resource_id` |
| `category` | 类别 | `common.#effect_category` |
| `magnitude` | 强度 | `number & common.#non_negative` |
| `duration_turns` | 持续回合 | `int & common.#non_negative` |
| `proc_chance` | 触发概率 | `common.#probability` |
| `designer_notes` | 策划备注 | `common.#non_empty_text`；不导出 |

## 道具与装备 · `content.item`

| 英文列 ID / CSV 列头 | 中文显示名 | Schema / 导出 |
|---|---|---|
| `name_text_id` | 名称文本ID | `common.#resource_id` |
| `description_text_id` | 描述文本ID | `common.#resource_id` |
| `category` | 类别 | `common.#item_category` |
| `rarity` | 稀有度 | `common.#rarity` |
| `list_price` | 标价 | `int & common.#non_negative & <=999999` |
| `tradable` | 可交易 | `*true \| bool` |
| `max_stack` | 堆叠上限 | `*1 \| int & >=1 & <=999` |
| `equipment_slot` | 装备槽位 | `common.#equipment_slot \| null` |
| `effect_id` | 效果ID | `string \| null` |
| `bonus_effect_ids` | 附加效果列表 | `[...common.#resource_id]` |
| `designer_notes` | 策划备注 | `string \| null`；不导出 |

## 技能：角色招式 · `content.skill`

| 英文列 ID / CSV 列头 | 中文显示名 | Schema / 导出 |
|---|---|---|
| `name_text_id` | 名称文本ID | `common.#resource_id` |
| `description_text_id` | 描述文本ID | `string \| null` |
| `unlock_level` | 解锁等级 | `int & >=1 & <=60` |
| `mana_cost` | 法力消耗 | `int & common.#non_negative` |
| `cooldown_turns` | 冷却回合 | `int & common.#non_negative` |
| `primary_effect_id` | 主效果ID | `common.#resource_id` |
| `bonus_effect_ids` | 附加效果列表 | `[...common.#resource_id]` |
| `multiplier` | 倍率 | `number & common.#positive` |
| `tags` | 标签 | `[...string]` |

## 英雄：队伍配置 · `content.hero`

| 英文列 ID / CSV 列头 | 中文显示名 | Schema / 导出 |
|---|---|---|
| `name_text_id` | 名称文本ID | `common.#resource_id` |
| `biography_text_id` | 简介文本ID | `common.#resource_id` |
| `role` | 定位 | `common.#hero_role` |
| `starting_level` | 初始等级 | `int & >=1 & <=60` |
| `base_hp` | 基础生命 | `int & common.#positive` |
| `base_attack` | 基础攻击 | `int & common.#non_negative` |
| `starting_equipment_id` | 初始装备ID | `string \| null` |
| `skill_ids` | 技能列表 | `[...common.#resource_id]` |
| `spawn_position` | 出生坐标 | `_local.#position` |
| `developer_notes` | 开发备注 | `string \| null`；不导出 |

## 等级成长：数值与依赖 · `balance.growth`

| 英文列 ID / CSV 列头 | 中文显示名 | Schema / 导出 |
|---|---|---|
| `level` | 等级 | `int & _local.#level_range` |
| `level_up_xp` | 升级经验 | `int & common.#positive` |
| `total_xp` | 累计经验 | `int & common.#non_negative` |
| `base_attack` | 基础攻击 | `int & common.#non_negative` |
| `base_defense` | 基础防御 | `int & common.#non_negative` |
| `base_hp` | 基础生命 | `int & common.#positive` |
| `power_rating` | 参考战力 | `int & common.#non_negative` |

## 掉落组：抽取规则 · `world.loot_group`

| 英文列 ID / CSV 列头 | 中文显示名 | Schema / 导出 |
|---|---|---|
| `chapter` | 章节 | `common.#chapter` |
| `description` | 说明 | `common.#non_empty_text` |
| `draw_count` | 抽取次数 | `int & >=1 & <=10` |

## 掉落明细：关系表 · `world.loot_entry`

| 英文列 ID / CSV 列头 | 中文显示名 | Schema / 导出 |
|---|---|---|
| `loot_group_id` | 掉落组ID | `common.#resource_id` |
| `item_id` | 道具ID | `common.#resource_id` |
| `weight` | 权重 | `common.#probability` |
| `min_quantity` | 最小数量 | `int & >=1` |
| `max_quantity` | 最大数量 | `int & >=1` |

## 敌人：三章遭遇 · `world.enemy`

| 英文列 ID / CSV 列头 | 中文显示名 | Schema / 导出 |
|---|---|---|
| `name_text_id` | 名称文本ID | `common.#resource_id` |
| `chapter` | 章节 | `common.#chapter` |
| `level` | 等级 | `int & >=1 & <=60` |
| `hp` | 生命 | `int & common.#positive` |
| `attack` | 攻击 | `int & common.#non_negative` |
| `skill_ids` | 技能列表 | `[...common.#resource_id]` |
| `loot_group_id` | 掉落组ID | `common.#resource_id` |
| `elite` | 精英 | `bool` |

## 关卡：章节地图 · `world.stage`

| 英文列 ID / CSV 列头 | 中文显示名 | Schema / 导出 |
|---|---|---|
| `name_text_id` | 名称文本ID | `common.#resource_id` |
| `chapter` | 章节 | `common.#chapter` |
| `recommended_level` | 建议等级 | `int & >=1 & <=60` |
| `stamina_cost` | 体力消耗 | `int & common.#positive` |
| `enemy_ids` | 敌人列表 | `[...common.#resource_id]` |
| `first_clear_reward_id` | 首通奖励ID | `string \| null` |
| `next_stage_id` | 后续关卡ID | `string \| null` |
| `entrance_position` | 入口坐标 | `_local.#position` |

## 任务：推进旅程 · `quest.quest`

| 英文列 ID / CSV 列头 | 中文显示名 | Schema / 导出 |
|---|---|---|
| `name_text_id` | 名称文本ID | `common.#resource_id` |
| `objective_text_id` | 目标文本ID | `common.#resource_id` |
| `stage_id` | 关卡ID | `common.#resource_id` |
| `prerequisite_quest_id` | 前置任务ID | `string \| null` |
| `quest_type` | 类型 | `"讨伐" \| "收集"` |
| `target_item_id` | 目标道具ID | `string \| null` |
| `required_quantity` | 需求数量 | `int & common.#positive` |
| `reward_gold` | 奖励金币 | `int & common.#non_negative` |
| `reward_item_ids` | 奖励道具列表 | `[...common.#resource_id]` |

## 合成配方：产物与解锁 · `craft.recipe`

| 英文列 ID / CSV 列头 | 中文显示名 | Schema / 导出 |
|---|---|---|
| `name_text_id` | 名称文本ID | `common.#resource_id` |
| `output_item_id` | 产物ID | `common.#resource_id` |
| `output_quantity` | 产出数量 | `int & common.#positive` |
| `workshop_level` | 工坊等级 | `int & >=1 & <=12` |
| `gold_cost` | 金币消耗 | `int & common.#non_negative` |
| `unlock_quest_id` | 解锁任务ID | `string \| null` |

## 配方材料：多对多关系 · `craft.ingredient`

| 英文列 ID / CSV 列头 | 中文显示名 | Schema / 导出 |
|---|---|---|
| `recipe_id` | 配方ID | `common.#resource_id` |
| `material_id` | 材料ID | `common.#resource_id` |
| `quantity` | 数量 | `int & common.#positive` |

## 商店：区域经营 · `economy.shop`

| 英文列 ID / CSV 列头 | 中文显示名 | Schema / 导出 |
|---|---|---|
| `name_text_id` | 名称文本ID | `common.#resource_id` |
| `chapter` | 章节 | `common.#chapter` |
| `currency` | 货币 | `"金币"` |
| `unlock_quest_id` | 解锁任务ID | `string \| null` |
| `is_open` | 营业中 | `bool` |

## 商品：折扣与导出练习 · `economy.product`

| 英文列 ID / CSV 列头 | 中文显示名 | Schema / 导出 |
|---|---|---|
| `shop_id` | 商店ID | `common.#resource_id` |
| `item_id` | 道具ID | `common.#resource_id` |
| `list_price` | 标价 | `int & common.#non_negative` |
| `discount` | 折扣 | `number & _local.#discount_range` |
| `sale_price` | 实售价 | `int & common.#non_negative` |
| `daily_limit` | 每日限购 | `int & common.#positive` |
| `listed` | 上架 | `*true \| bool` |
| `designer_notes` | 策划备注 | `string \| null`；不导出 |

## 系统参数：参与校验但不导出 · `system.parameter`

| 英文列 ID / CSV 列头 | 中文显示名 | Schema / 导出 |
|---|---|---|
| `value` | 数值 | `number & common.#non_negative` |
| `unit` | 单位 | `"次" \| "级" \| "比例" \| "格"` |
| `description` | 说明 | `common.#non_empty_text` |

## 本地化：中文与英语 · `localization.text`

| 英文列 ID / CSV 列头 | 中文显示名 | Schema / 导出 |
|---|---|---|
| `zh_cn` | 简体中文 | `common.#non_empty_text` |
| `en_us` | 英语 | `common.#non_empty_text` |
| `usage_notes` | 使用说明 | `string \| null` |

## Schema 英文标识与中文释义

Schema 目前没有独立显示名字段；面板显示英文标识，这里和源文件注释提供中文说明。定义的 enum 值仍为中文。

| 标识 | 中文释义 | 作用域 |
|---|---|---|
| `#resource_id` | 资源标识 | 公共 `common` |
| `#non_empty_text` | 必填文本 | 公共 `common` |
| `#non_negative` | 非负 | 公共 `common` |
| `#positive` | 正数 | 公共 `common` |
| `#probability_range` | 概率范围 | 公共 `common` |
| `#probability` | 概率 | 公共 `common` |
| `#rarity` | 稀有度 | 公共 `common` |
| `#item_category` | 道具类别 | 公共 `common` |
| `#effect_category` | 效果类别 | 公共 `common` |
| `#chapter` | 章节 | 公共 `common` |
| `#hero_role` | 角色定位 | 公共 `common` |
| `#equipment_slot` | 装备槽位 | 公共 `common` |
| `#position` | 坐标 | 本表 `_local` |
| `#level_range` | 等级 | 本表 `_local` |
| `#discount_range` | 折扣范围 | 本表 `_local` |
