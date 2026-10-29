# 云岚纪：中文 RPG 配表实战

这是一个原创单机队伍 RPG 的配表工程：青岚村的星脉失衡，旅人沿赤砂遗迹寻找线索，最后在霜月城修复星井。三章各有八个关卡、八种敌人和十六个任务，人物和地点均为虚构。

打开本目录的 `yunlan-rpg.dtproj`。工程独立于 skyharbor，16表、1002行：480行中英本地化、522行玩法数据。它面向中文策划和程序共同维护：**程序使用稳定英文列 ID；编辑器使用中文显示名与中文备注。CSV 列头仍是英文 ID。** 行 ID、路径、package 和 Schema 标识也为英文；中文内容、枚举和译文保持中文。

例如列 ID 为 `list_price`，中文显示名为“标价”，编辑器表头显示“标价 (list_price)”。修改显示名不会改程序字段、公式或 CSV；真正重命名列 ID 才会迁移这些引用。结构体坐标成员使用 `x/y`，导出 JSON 的键也为英文。完整对照见 [字段与 Schema 字典](FIELD_IDS.md)，功能落点见 [功能索引](FEATURES.md)。Schema 面板以英文标识列出定义，中文含义写在字典与源文件注释中，未使用尚不支持的独立 Schema 显示名元数据。

编辑实验前先“文件→另存为”创建副本。下文以原始数据为准；可以在设置中选择简体中文界面。手动编译模式下，改值后先点“编译”再观察公式与诊断。

## 五分钟：改价并导出

1. 打开 `economy.product`（商品），定位 `product_01_1`。看中文表头“标价 (list_price)”“折扣 (discount)”“实售价 (sale_price)”：原值40、0.8、32，公式是 `ceil(list_price * discount)`。
2. 将“标价”改100，编译后“实售价”应变80。向上取整避免金币出现小数。这是商品表的独立定价，不是跨表读取道具价格，策划需要主动维护两表价格关系。
3. 右键“标价”单元格查看中文评论。也可通过列菜单把显示名改成“商品定价”；此时真实 ID 仍是 `list_price`，公式与 CSV 不变。
4. 工程导出设置排除 `system.parameter`，商品表的导出列排除“策划备注 (designer_notes)”。执行 CSV 导出，应得到15个文件；`economy.product.csv` 含 `list_price,discount,sale_price`，不含中文列头或 `designer_notes`，也不生成参数表 CSV。
5. Undo/Redo 对比数值变化。公式始终保存在 `_tool.formulas`，计算结果不会被复制成行内固定值。

程序在导出后按英文 ID 读取，示例（从当前工程目录运行）：

```python
import csv

with open("exports/economy.product.csv", encoding="utf-8", newline="") as source:
    for row in csv.DictReader(source):
        print(row["item_id"], int(row["sale_price"]))
```

显示名可调整为团队习惯的中文，以上代码使用的字段 ID 保持稳定。CSV 文本和枚举内容仍可包含中文。


## 核心五表：文本、英雄、技能、效果、道具

先看 `localization.text.hero_01_name`：“简体中文 (zh_cn)”是“陆青禾”，“英语 (en_us)”是“Lu Qinghe”。名称引用列 `name_text_id` 保存文本 ID，不自动联表显示译文；游戏运行时选择语言列。

接着打开 `content.hero.hero_01`：

- “名称文本ID (name_text_id)”和“简介文本ID (biography_text_id)”指向文本表。
- “技能列表 (skill_ids)”含四个技能 ID，第一项 `skill_01`。双击打开数组编辑器；数组是一个完整单元格，不是四列，也不是结构体数组。
- `content.skill.skill_01.primary_effect_id=effect_01`。在效果表看“类别 (category)”为“伤害”、“强度 (magnitude)”10、“持续回合 (duration_turns)”1。技能倍率由游戏运行时与效果强度组合，不声称有跨表公式。
- 四种定位的动作一致：战士破阵斩为伤害，游侠疾矢射为伤害/缚足箭为控制，医师回春术为治疗，术士引星术为伤害。第三槽是开战触发一次的被动，`mana_cost/cooldown_turns=0`。`skill_ids` 列出完整四槽招式，`unlock_level` 是英雄学习门槛，不表示初始全部可用；敌人直接按遭遇表装配既有主动招式。
- `starting_equipment_id=item_weapon_01` 引向道具。武器、防具、饰品都是道具类别，用 `category/equipment_slot` 区分，背包、合成和商店共用一套道具 ID。
- “出生坐标 (spawn_position)”为 `{x:0,y:0}`，使用本表 `_local.#position`。`x/y` 表示横/纵坐标，双击结构编辑器可以修改；CSV 中该单元格是英文键的 JSON。

F12 或 Ctrl/Command 点击引用可导航。数组使用专用编辑器/引用菜单，不把整个 JSON 当单个目标 ID。公式、外键和导出配置始终引用英文 ID，而不是中文显示名。

## 任务、关卡、掉落与合成

`quest_01` 是青岚村溪桥讨伐任务，`stage_id=stage_01`、`required_quantity=3`；`quest_02` 要收集两份 `item_material_01`。

1. `world.stage.stage_01.enemy_ids=["enemy_01"]`，`next_stage_id=stage_02`；最后一关 `stage_24.next_stage_id=null`。
2. `world.enemy.enemy_01.loot_group_id=loot_01`。明细 `drop_01_1` 至 `drop_01_3` 是两种材料和一种药剂，`weight` 为0.4/0.3/0.3，合计1，组内 `draw_count=1`。跨行权重和由样例回归额外验证，工具没有通用业务规则引擎。
3. `craft.recipe.recipe_01.output_item_id=item_weapon_01`，`unlock_quest_id=quest_02`。材料明细 `ingredient_01_1/ingredient_01_2` 各自在同一行保存 `material_id/quantity`，避免平行数组错位。
4. 任务“奖励道具列表 (reward_item_ids)”约定每个 ID 发一件，金币在 `reward_gold`；不使用另一条数量数组。
5. `economy.shop.shop_01` 是默认开放的药铺，其他商店通过 `unlock_quest_id` 开放。商品行同时保存 `shop_id/item_id`、价格和限购数量。

每个地区有四阶材料装备和八处遭遇，差异体现在等级、强度、价格、配方与解锁。每个收集任务的指定关卡确实能掉落目标材料，目标文字写出材料名；概率掉落可能需要重复挑战。敌人列表表示种类，不规定生成数量；收集任务检查背包拥有数量，不额外消耗，合成才按明细扣除。这些是游戏运行约定，编辑器不会执行战斗或任务。

## Schema 与数值

公共定义位于 `schema/common.cue`：`#resource_id`（资源标识）、`#non_empty_text`（非空文本）、`#rarity`（稀有度）等是 type；`#non_negative/#positive/#probability_range` 是 constraint。`#probability` 将 number 与概率范围组合，列使用 `int & common.#non_negative` 展示 requirements。这里的 requirements 是类型约束引用，不是任务解锁条件。

本地定义有英雄/关卡的 `#position` type、成长表 `#level_range` constraint、商品表 `#discount_range` constraint。它们只供本表使用；公共定义可跨表导入。英文标识和中文释义见字段字典。

- `balance.growth.level_01.power_rating` 使用 `base_attack * 3 + base_defense * 2 + floor(base_hp / 10)`，原值 `23*3+12*2+floor(225/10)=115`。
- `level_01.total_xp=0`；`level_02.total_xp` 引用同表上一行的 `total_xp/level_up_xp`，结果103。60级形成无环依赖，使用稳定行 ID 而非 A10 地址。
- 公式支持同行和同表完整地址，不跨表读取，不经外键联表，也不读取数组元素或动态下标。
- `item_potion_01` 省略 `tradable`，按 Schema 默认 true；其 `effect_id` 用行内 `*"effect_04" | string` 演示默认表达式外键。其他药剂有明确治疗效果；材料和任务物品 `effect_id=null`，列无 Schema 默认，因此真正无效果。
- 若列本身有 Schema 默认，本工具按既有规则把省略值或 null 交给默认处理；null 不能绕过默认。空字符串、null、`[]` 不同，见 [Schema 格式](../../restricted-cue-schema-v1.md)。
- 未导出的 `system.parameter.level_cap.value=60` 是设计参考，与成长上限一致，不被其他表公式读取。

## 在副本里练习

- 在商品的“标价 (list_price)”从 `product_01_1` 起向下粘贴两行100、200；`sale_price` 应为80、200，Undo 一次撤销粘贴。
- 修改显示名“标价→商品定价”，验证 CSV 的 `list_price` 不变；另试真正重命名列 ID `list_price→base_price`，检查公式、备注键和 CSV 列头迁移，中文显示名保留。
- 将公共 Schema 标识 `non_empty_text→required_text`，检查文本表和参数表引用同步，再 Undo；仍被引用的定义不允许直接删除。
- `item_weapon_01→item_qinglan_blade` 应更新英雄、配方、商品的字面量外键。删除仍被引用的装备应被拒绝。
- 改名 `effect_04` 会因新手药剂的非字面量默认表达式引用而安全拒绝，文件和历史不变，不能只手工改目标 ID。
- 双击“技能列表 (skill_ids)”重排 `skill_01/skill_02`，保存后关闭重开、编译，再 Undo/Redo；双击技能“标签 (tags)”新增字符串“教学验证”，同样验证保存和历史。
- 新行 ID 如 `item_trial_blade`，新增列 ID 也用英文 snake_case，再单独设置中文显示名。新行字符串按 CUE 语法加引号，数组/结构用专用编辑器；已有普通字符串直接输入文字，以 `=` 开头表示公式。
- 删除带附件的实验表前确认版本包含 PR #10；进程内 Undo 可恢复附件，但历史不是崩溃备份。

当前没有行排序/筛选/全文搜索和自动联表显示。已有稳定 ID 排序、项目导出表选择和表级列选择，不把输出选择冒称行筛选。

## 故障排查与验证

仅在副本中把 `effect_07.proc_chance` 改1.2，或把 `primary_effect_id` 改成不存在的 ID，应出现可定位诊断。把未导出的 `system.parameter.level_cap.value` 改-1，也必须阻止整个工程导出并保留原有效 CSV，Undo 后恢复。

在仓库根目录、已有 Go 依赖可用时运行：

```sh
go test ./test/functional -run 'TestYunlan|TestChineseColumns' -count=1
```

样例回归检查英文列/Schema/结构成员 ID、中文显示名与备注、英文 CSV 头及 JSON 键、16表/1002行、15份 golden、实际公式/玩法链、显示名与 ID 独立性、改名/粘贴/删除保护和数组保存/重开/编译/UndoRedo。多行数组由格式兼容回归保护，模块仍为 v0.9.0。

`frontend/tests/yunlan-arrays.spec.ts` 读取本例实际行和显示名，操作真实网格与数组对话框；Wails 边界仍 mock，实际 CUE 文件链由 Go 回归验证。另有独立的 `chinese_columns_test.go` 与 `chinese-columns.spec.ts` 检查工具对中文 ID 的兼容性，它们不是本样例的字段设计方式。
