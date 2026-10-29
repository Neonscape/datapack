# 功能与英文 ID 索引

逻辑表 `content.item` 位于 `tables/content/item/item.cue`。主键来自 `rows` 的英文键，CSV 首列为 `id`。所有业务列使用英文 snake_case ID；中文标题来自 `_editor.columns.<id>.title`，仅供编辑器阅读。CSV 输出 ID，不输出标题。完整逐列对照见 [FIELD_IDS.md](FIELD_IDS.md)。

| 逻辑表 | 行数 | 英文列 ID 与功能 |
|---|---:|---|
| localization.text | 480 | `zh_cn/en_us/usage_notes`：中文、英语及备注，文本 ID 外键，无自动 join |
| content.item | 72 | `category/rarity/tradable/equipment_slot/effect_id/bonus_effect_ids`：道具子类、枚举、默认、nullable 与数组 FK |
| content.effect | 36 | `category/magnitude/duration_turns/proc_chance`：四类效果与公共约束，隐藏 `designer_notes` |
| content.skill | 48 | `primary_effect_id/bonus_effect_ids/tags`：标量/数组 FK 与中文标签 |
| content.hero | 12 | `role/skill_ids/starting_equipment_id/spawn_position`：定位、技能列表、nullable FK、`x/y` 扁平结构 |
| balance.growth | 60 | `level/power_rating/total_xp`：本地约束、同行公式、同表跨行依赖 |
| world.enemy | 24 | `chapter/level/hp/attack/skill_ids/loot_group_id/elite`：遭遇、主动招式、掉落和布尔 |
| world.loot_group | 12 | `draw_count/description`：抽取规则，权重和由业务回归验证 |
| world.loot_entry | 36 | `loot_group_id/item_id/weight/min_quantity/max_quantity`：组-道具关系 |
| world.stage | 24 | `enemy_ids/next_stage_id/first_clear_reward_id/entrance_position`：数组/nullable FK 与坐标 |
| quest.quest | 48 | `stage_id/prerequisite_quest_id/target_item_id/required_quantity/reward_item_ids`：任务链和奖励 |
| craft.recipe | 36 | `output_item_id/output_quantity/workshop_level/unlock_quest_id`：产物与解锁 |
| craft.ingredient | 72 | `recipe_id/material_id/quantity`：每配方两行材料，不用平行数组 |
| economy.shop | 12 | `chapter/currency/unlock_quest_id/is_open`：章节商店与 nullable 解锁 |
| economy.product | 24 | `shop_id/item_id/list_price/discount/sale_price/daily_limit`：商品、折扣公式和隐藏 `designer_notes` |
| system.parameter | 6 | `value/unit/description`：设计参考，排除 CSV 但仍参与完整校验 |

核心链：`hero_01 → skill_01 → effect_01`，名称经 `name_text_id → localization.text.zh_cn`；`quest_02 → stage_01 → enemy_01 → loot_01 → drop_01_1 → item_material_01`；`recipe_01 → ingredient_01_1/ingredient_01_2 → 材料`，产物 `item_weapon_01` 被英雄和商品引用。

数值地址：`balance.growth.level_02.total_xp` 读取同表 `level_01.total_xp + level_01.level_up_xp`。显示名“累计经验”不参与表达式解析。中文内容与枚举值保留中文，但所有引用字段/Schema 标识为英文。

公共 type：`resource_id/non_empty_text/probability/rarity/item_category/effect_category/chapter/hero_role/equipment_slot`。公共 constraint：`non_negative/positive/probability_range`。本地 type：`position`；本地 constraint：`level_range/discount_range`。中文释义与源注释见字段字典。

语义回归检查12英雄/48技能的定位与效果、被动消耗、敌人招式、全部收集任务指定关卡的材料可获得性和配方原料链。这些是样例游戏规则，不冒称编辑器有通用业务规则引擎。

工程菜单可练习另存为、自动/手动编译、导出表/列选择、JSON 编码、中文显示名和评论、引用导航、Schema 编辑、粘贴、改名与 Undo/Redo。没有行筛选/搜索、跨表公式、译文自动 join、结构体数组、嵌套容器或动态下标。
