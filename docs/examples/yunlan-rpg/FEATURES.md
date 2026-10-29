# 功能到表/字段的索引

路径规则：逻辑表 `content.item` 位于 `tables/content/item/item.cue`；主键是 `rows` 的键，不是额外的中文“ID”列。表头ID导出为 `id`，其他字段都使用真实中文名称。

| 逻辑表 | 行数 | 主要用途、关系与可操作字段 |
|---|---:|---|
| localization.text | 480 | `简体中文/英语` 两种译文，文本ID被名称/描述列引用，无自动join |
| content.item | 72 | 六类道具；`稀有度`枚举、`可交易`默认、`装备槽位`nullable、`效果ID/额外效果ID`外键；新手药剂单独演示默认表达式，普通材料null保持无效果 |
| content.effect | 36 | 四类效果；`触发概率`公共约束、`持续回合`整数、隐藏`策划说明` |
| content.skill | 48 | `主效果ID`标量FK、`额外效果ID`数组FK、`标签`字符串数组、可空描述文本 |
| content.hero | 12 | 四种定位；`技能ID`数组、`初始装备ID`nullable FK、`出生坐标`本地扁平结构 |
| balance.growth | 60 | `等级`本地约束、`参考战力`同行公式、`累计经验`同表跨行依赖 |
| world.enemy | 24 | 章节、等级、生命、攻击、布尔`精英`；技能与掉落组外键 |
| world.loot_group | 12 | `抽取次数`与说明，明细权重和由样例回归验证 |
| world.loot_entry | 36 | 组-道具关系；同一行保存权重和数量范围 |
| world.stage | 24 | `敌人ID`数组、`后续关卡ID`同表nullable FK、`入口坐标`结构 |
| quest.quest | 48 | 关卡/前置任务/目标材料/奖励道具关系；中文任务目标文本 |
| craft.recipe | 36 | 装备产物、工坊等级、金币与解锁任务 |
| craft.ingredient | 72 | 每配方两行材料；数量与材料同一行，避免平行数组 |
| economy.shop | 12 | 章节商店、nullable开放任务、布尔营业状态 |
| economy.product | 24 | 商品关系、独立标价、折扣、`实售价`公式、隐藏`策划备注` |
| system.parameter | 6 | 设计参考参数；项目明确排除CSV，但仍参与编译/外键校验 |

## 跨表链索引

- 角色链：`hero_01 → skill_01 → effect_01`；英雄名称 `hero_01_name` → 本地化“陆青禾”。
- 内容链：`quest_02 → stage_01 → enemy_01 → loot_01 → drop_01_1 → item_material_01`。
- 工坊链：`recipe_01 → ingredient_01_1/ingredient_01_2 → 材料`；产物 `item_weapon_01` → `hero_01.初始装备ID` / `product_01_2.道具ID`。
- 数值链：`level_02.累计经验 → level_01.累计经验 + level_01.升级经验`。所有地址仍在 `balance.growth` 内，不是跨表公式。

## Schema索引

- 公共type：资源标识、必填文本、概率、稀有度、道具类别、效果类别、章节、角色定位、装备槽位。
- 公共constraint：非负、正数、概率范围。公共 `#概率` requirements 引用 `#概率范围`；列中 `int & common.#非负` 展示type与constraint组合。
- 本地type：英雄/关卡各有 `#坐标`，只供本表使用。
- 本地constraint：成长 `#等级`、商品 `#折扣范围`。
- 约束错误示例：`effect_07.触发概率=1.2`、`system.parameter.level_cap.数值=-1`，仅在副本实验。

## 工程级能力与边界

通过已有工程菜单练习另存为、自动/手动编译、项目导出表选择、CSV列与JSON编码、评论、引用导航、Schema编辑、批量粘贴、行/列/表改名和全局Undo/Redo。新建/删除表请在副本添加实验表，避免破坏核心关系。

不演示尚未实现的行筛选/排序/搜索、跨表公式、译文自动联表、结构体数组、嵌套容器、动态公式下标；这些没有被伪装成示例功能。导出会校验整个项目，输出选择不是绕过校验的方式。

语义验收逐一检查12位英雄的48项技能名/定位/效果和被动消耗，24个收集任务指定关卡的材料可获得性、配方解锁任务与原料关系；这些游戏规则由样例回归补充，不冒称编辑器通用业务规则。
