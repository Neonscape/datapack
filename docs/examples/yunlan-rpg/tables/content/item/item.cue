package item

import "example.com/datapack/yunlan-rpg/schema:common"

#PrimaryKey: common.#resource_id

#Row: {
    name_text_id: common.#resource_id
    description_text_id: common.#resource_id
    category: common.#item_category
    rarity: common.#rarity
    list_price: int & common.#non_negative & <=999999
    tradable: *true | bool
    max_stack: *1 | int & >=1 & <=999
    equipment_slot: common.#equipment_slot | null
    effect_id: string | null
    bonus_effect_ids: [...common.#resource_id]
    designer_notes: string | null
}

rows: {
  "item_potion_01": {
    "name_text_id": "item_potion_01_name",
    "description_text_id": "item_potion_01_desc",
    "category": "消耗品",
    "rarity": "普通",
    "list_price": 40,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": *"effect_04" | string,
    "bonus_effect_ids": [],
    "designer_notes": "第1章消耗品，阶级1；导出时排除此策划说明列。"
  },
  "item_potion_02": {
    "name_text_id": "item_potion_02_name",
    "description_text_id": "item_potion_02_desc",
    "category": "消耗品",
    "rarity": "精良",
    "list_price": 60,
    "tradable": true,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": "effect_05",
    "bonus_effect_ids": [],
    "designer_notes": "第1章消耗品，阶级2；导出时排除此策划说明列。"
  },
  "item_potion_03": {
    "name_text_id": "item_potion_03_name",
    "description_text_id": "item_potion_03_desc",
    "category": "消耗品",
    "rarity": "稀有",
    "list_price": 80,
    "tradable": true,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": "effect_06",
    "bonus_effect_ids": [],
    "designer_notes": "第1章消耗品，阶级3；导出时排除此策划说明列。"
  },
  "item_potion_04": {
    "name_text_id": "item_potion_04_name",
    "description_text_id": "item_potion_04_desc",
    "category": "消耗品",
    "rarity": "史诗",
    "list_price": 100,
    "tradable": true,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": "effect_06",
    "bonus_effect_ids": [],
    "designer_notes": "第1章消耗品，阶级4；导出时排除此策划说明列。"
  },
  "item_potion_05": {
    "name_text_id": "item_potion_05_name",
    "description_text_id": "item_potion_05_desc",
    "category": "消耗品",
    "rarity": "普通",
    "list_price": 100,
    "tradable": true,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": "effect_16",
    "bonus_effect_ids": [],
    "designer_notes": "第2章消耗品，阶级1；导出时排除此策划说明列。"
  },
  "item_potion_06": {
    "name_text_id": "item_potion_06_name",
    "description_text_id": "item_potion_06_desc",
    "category": "消耗品",
    "rarity": "精良",
    "list_price": 120,
    "tradable": true,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": "effect_17",
    "bonus_effect_ids": [],
    "designer_notes": "第2章消耗品，阶级2；导出时排除此策划说明列。"
  },
  "item_potion_07": {
    "name_text_id": "item_potion_07_name",
    "description_text_id": "item_potion_07_desc",
    "category": "消耗品",
    "rarity": "稀有",
    "list_price": 140,
    "tradable": true,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": "effect_18",
    "bonus_effect_ids": [],
    "designer_notes": "第2章消耗品，阶级3；导出时排除此策划说明列。"
  },
  "item_potion_08": {
    "name_text_id": "item_potion_08_name",
    "description_text_id": "item_potion_08_desc",
    "category": "消耗品",
    "rarity": "史诗",
    "list_price": 160,
    "tradable": true,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": "effect_18",
    "bonus_effect_ids": [],
    "designer_notes": "第2章消耗品，阶级4；导出时排除此策划说明列。"
  },
  "item_potion_09": {
    "name_text_id": "item_potion_09_name",
    "description_text_id": "item_potion_09_desc",
    "category": "消耗品",
    "rarity": "普通",
    "list_price": 160,
    "tradable": true,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": "effect_28",
    "bonus_effect_ids": [],
    "designer_notes": "第3章消耗品，阶级1；导出时排除此策划说明列。"
  },
  "item_potion_10": {
    "name_text_id": "item_potion_10_name",
    "description_text_id": "item_potion_10_desc",
    "category": "消耗品",
    "rarity": "精良",
    "list_price": 180,
    "tradable": true,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": "effect_29",
    "bonus_effect_ids": [],
    "designer_notes": "第3章消耗品，阶级2；导出时排除此策划说明列。"
  },
  "item_potion_11": {
    "name_text_id": "item_potion_11_name",
    "description_text_id": "item_potion_11_desc",
    "category": "消耗品",
    "rarity": "稀有",
    "list_price": 200,
    "tradable": true,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": "effect_30",
    "bonus_effect_ids": [],
    "designer_notes": "第3章消耗品，阶级3；导出时排除此策划说明列。"
  },
  "item_potion_12": {
    "name_text_id": "item_potion_12_name",
    "description_text_id": "item_potion_12_desc",
    "category": "消耗品",
    "rarity": "史诗",
    "list_price": 220,
    "tradable": true,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": "effect_30",
    "bonus_effect_ids": [],
    "designer_notes": "第3章消耗品，阶级4；导出时排除此策划说明列。"
  },
  "item_material_01": {
    "name_text_id": "item_material_01_name",
    "description_text_id": "item_material_01_desc",
    "category": "材料",
    "rarity": "普通",
    "list_price": 40,
    "tradable": true,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": null,
    "bonus_effect_ids": [],
    "designer_notes": "第1章材料，阶级1；材料用两条关系记录配比。"
  },
  "item_material_02": {
    "name_text_id": "item_material_02_name",
    "description_text_id": "item_material_02_desc",
    "category": "材料",
    "rarity": "精良",
    "list_price": 60,
    "tradable": true,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": null,
    "bonus_effect_ids": [],
    "designer_notes": "第1章材料，阶级2；材料用两条关系记录配比。"
  },
  "item_material_03": {
    "name_text_id": "item_material_03_name",
    "description_text_id": "item_material_03_desc",
    "category": "材料",
    "rarity": "稀有",
    "list_price": 80,
    "tradable": true,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": null,
    "bonus_effect_ids": [],
    "designer_notes": "第1章材料，阶级3；材料用两条关系记录配比。"
  },
  "item_material_04": {
    "name_text_id": "item_material_04_name",
    "description_text_id": "item_material_04_desc",
    "category": "材料",
    "rarity": "史诗",
    "list_price": 100,
    "tradable": true,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": null,
    "bonus_effect_ids": [],
    "designer_notes": "第1章材料，阶级4；材料用两条关系记录配比。"
  },
  "item_material_05": {
    "name_text_id": "item_material_05_name",
    "description_text_id": "item_material_05_desc",
    "category": "材料",
    "rarity": "普通",
    "list_price": 100,
    "tradable": true,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": null,
    "bonus_effect_ids": [],
    "designer_notes": "第2章材料，阶级1；材料用两条关系记录配比。"
  },
  "item_material_06": {
    "name_text_id": "item_material_06_name",
    "description_text_id": "item_material_06_desc",
    "category": "材料",
    "rarity": "精良",
    "list_price": 120,
    "tradable": true,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": null,
    "bonus_effect_ids": [],
    "designer_notes": "第2章材料，阶级2；材料用两条关系记录配比。"
  },
  "item_material_07": {
    "name_text_id": "item_material_07_name",
    "description_text_id": "item_material_07_desc",
    "category": "材料",
    "rarity": "稀有",
    "list_price": 140,
    "tradable": true,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": null,
    "bonus_effect_ids": [],
    "designer_notes": "第2章材料，阶级3；材料用两条关系记录配比。"
  },
  "item_material_08": {
    "name_text_id": "item_material_08_name",
    "description_text_id": "item_material_08_desc",
    "category": "材料",
    "rarity": "史诗",
    "list_price": 160,
    "tradable": true,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": null,
    "bonus_effect_ids": [],
    "designer_notes": "第2章材料，阶级4；材料用两条关系记录配比。"
  },
  "item_material_09": {
    "name_text_id": "item_material_09_name",
    "description_text_id": "item_material_09_desc",
    "category": "材料",
    "rarity": "普通",
    "list_price": 160,
    "tradable": true,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": null,
    "bonus_effect_ids": [],
    "designer_notes": "第3章材料，阶级1；材料用两条关系记录配比。"
  },
  "item_material_10": {
    "name_text_id": "item_material_10_name",
    "description_text_id": "item_material_10_desc",
    "category": "材料",
    "rarity": "精良",
    "list_price": 180,
    "tradable": true,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": null,
    "bonus_effect_ids": [],
    "designer_notes": "第3章材料，阶级2；材料用两条关系记录配比。"
  },
  "item_material_11": {
    "name_text_id": "item_material_11_name",
    "description_text_id": "item_material_11_desc",
    "category": "材料",
    "rarity": "稀有",
    "list_price": 200,
    "tradable": true,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": null,
    "bonus_effect_ids": [],
    "designer_notes": "第3章材料，阶级3；材料用两条关系记录配比。"
  },
  "item_material_12": {
    "name_text_id": "item_material_12_name",
    "description_text_id": "item_material_12_desc",
    "category": "材料",
    "rarity": "史诗",
    "list_price": 220,
    "tradable": true,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": null,
    "bonus_effect_ids": [],
    "designer_notes": "第3章材料，阶级4；材料用两条关系记录配比。"
  },
  "item_weapon_01": {
    "name_text_id": "item_weapon_01_name",
    "description_text_id": "item_weapon_01_desc",
    "category": "武器",
    "rarity": "普通",
    "list_price": 140,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "主手",
    "effect_id": "effect_01",
    "bonus_effect_ids": ["effect_10"],
    "designer_notes": "第1章武器，阶级1；导出时排除此策划说明列。"
  },
  "item_weapon_02": {
    "name_text_id": "item_weapon_02_name",
    "description_text_id": "item_weapon_02_desc",
    "category": "武器",
    "rarity": "精良",
    "list_price": 160,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "主手",
    "effect_id": "effect_01",
    "bonus_effect_ids": ["effect_10"],
    "designer_notes": "第1章武器，阶级2；导出时排除此策划说明列。"
  },
  "item_weapon_03": {
    "name_text_id": "item_weapon_03_name",
    "description_text_id": "item_weapon_03_desc",
    "category": "武器",
    "rarity": "稀有",
    "list_price": 180,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "主手",
    "effect_id": "effect_01",
    "bonus_effect_ids": ["effect_10"],
    "designer_notes": "第1章武器，阶级3；导出时排除此策划说明列。"
  },
  "item_weapon_04": {
    "name_text_id": "item_weapon_04_name",
    "description_text_id": "item_weapon_04_desc",
    "category": "武器",
    "rarity": "史诗",
    "list_price": 200,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "主手",
    "effect_id": "effect_01",
    "bonus_effect_ids": ["effect_10"],
    "designer_notes": "第1章武器，阶级4；导出时排除此策划说明列。"
  },
  "item_weapon_05": {
    "name_text_id": "item_weapon_05_name",
    "description_text_id": "item_weapon_05_desc",
    "category": "武器",
    "rarity": "普通",
    "list_price": 200,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "主手",
    "effect_id": "effect_13",
    "bonus_effect_ids": ["effect_22"],
    "designer_notes": "第2章武器，阶级1；导出时排除此策划说明列。"
  },
  "item_weapon_06": {
    "name_text_id": "item_weapon_06_name",
    "description_text_id": "item_weapon_06_desc",
    "category": "武器",
    "rarity": "精良",
    "list_price": 220,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "主手",
    "effect_id": "effect_13",
    "bonus_effect_ids": ["effect_22"],
    "designer_notes": "第2章武器，阶级2；导出时排除此策划说明列。"
  },
  "item_weapon_07": {
    "name_text_id": "item_weapon_07_name",
    "description_text_id": "item_weapon_07_desc",
    "category": "武器",
    "rarity": "稀有",
    "list_price": 240,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "主手",
    "effect_id": "effect_13",
    "bonus_effect_ids": ["effect_22"],
    "designer_notes": "第2章武器，阶级3；导出时排除此策划说明列。"
  },
  "item_weapon_08": {
    "name_text_id": "item_weapon_08_name",
    "description_text_id": "item_weapon_08_desc",
    "category": "武器",
    "rarity": "史诗",
    "list_price": 260,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "主手",
    "effect_id": "effect_13",
    "bonus_effect_ids": ["effect_22"],
    "designer_notes": "第2章武器，阶级4；导出时排除此策划说明列。"
  },
  "item_weapon_09": {
    "name_text_id": "item_weapon_09_name",
    "description_text_id": "item_weapon_09_desc",
    "category": "武器",
    "rarity": "普通",
    "list_price": 260,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "主手",
    "effect_id": "effect_25",
    "bonus_effect_ids": ["effect_34"],
    "designer_notes": "第3章武器，阶级1；导出时排除此策划说明列。"
  },
  "item_weapon_10": {
    "name_text_id": "item_weapon_10_name",
    "description_text_id": "item_weapon_10_desc",
    "category": "武器",
    "rarity": "精良",
    "list_price": 280,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "主手",
    "effect_id": "effect_25",
    "bonus_effect_ids": ["effect_34"],
    "designer_notes": "第3章武器，阶级2；导出时排除此策划说明列。"
  },
  "item_weapon_11": {
    "name_text_id": "item_weapon_11_name",
    "description_text_id": "item_weapon_11_desc",
    "category": "武器",
    "rarity": "稀有",
    "list_price": 300,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "主手",
    "effect_id": "effect_25",
    "bonus_effect_ids": ["effect_34"],
    "designer_notes": "第3章武器，阶级3；导出时排除此策划说明列。"
  },
  "item_weapon_12": {
    "name_text_id": "item_weapon_12_name",
    "description_text_id": "item_weapon_12_desc",
    "category": "武器",
    "rarity": "史诗",
    "list_price": 320,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "主手",
    "effect_id": "effect_25",
    "bonus_effect_ids": ["effect_34"],
    "designer_notes": "第3章武器，阶级4；导出时排除此策划说明列。"
  },
  "item_armor_01": {
    "name_text_id": "item_armor_01_name",
    "description_text_id": "item_armor_01_desc",
    "category": "防具",
    "rarity": "普通",
    "list_price": 140,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "躯干",
    "effect_id": "effect_10",
    "bonus_effect_ids": [],
    "designer_notes": "第1章防具，阶级1；导出时排除此策划说明列。"
  },
  "item_armor_02": {
    "name_text_id": "item_armor_02_name",
    "description_text_id": "item_armor_02_desc",
    "category": "防具",
    "rarity": "精良",
    "list_price": 160,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "躯干",
    "effect_id": "effect_11",
    "bonus_effect_ids": [],
    "designer_notes": "第1章防具，阶级2；导出时排除此策划说明列。"
  },
  "item_armor_03": {
    "name_text_id": "item_armor_03_name",
    "description_text_id": "item_armor_03_desc",
    "category": "防具",
    "rarity": "稀有",
    "list_price": 180,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "躯干",
    "effect_id": "effect_12",
    "bonus_effect_ids": [],
    "designer_notes": "第1章防具，阶级3；导出时排除此策划说明列。"
  },
  "item_armor_04": {
    "name_text_id": "item_armor_04_name",
    "description_text_id": "item_armor_04_desc",
    "category": "防具",
    "rarity": "史诗",
    "list_price": 200,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "躯干",
    "effect_id": "effect_12",
    "bonus_effect_ids": [],
    "designer_notes": "第1章防具，阶级4；导出时排除此策划说明列。"
  },
  "item_armor_05": {
    "name_text_id": "item_armor_05_name",
    "description_text_id": "item_armor_05_desc",
    "category": "防具",
    "rarity": "普通",
    "list_price": 200,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "躯干",
    "effect_id": "effect_22",
    "bonus_effect_ids": [],
    "designer_notes": "第2章防具，阶级1；导出时排除此策划说明列。"
  },
  "item_armor_06": {
    "name_text_id": "item_armor_06_name",
    "description_text_id": "item_armor_06_desc",
    "category": "防具",
    "rarity": "精良",
    "list_price": 220,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "躯干",
    "effect_id": "effect_23",
    "bonus_effect_ids": [],
    "designer_notes": "第2章防具，阶级2；导出时排除此策划说明列。"
  },
  "item_armor_07": {
    "name_text_id": "item_armor_07_name",
    "description_text_id": "item_armor_07_desc",
    "category": "防具",
    "rarity": "稀有",
    "list_price": 240,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "躯干",
    "effect_id": "effect_24",
    "bonus_effect_ids": [],
    "designer_notes": "第2章防具，阶级3；导出时排除此策划说明列。"
  },
  "item_armor_08": {
    "name_text_id": "item_armor_08_name",
    "description_text_id": "item_armor_08_desc",
    "category": "防具",
    "rarity": "史诗",
    "list_price": 260,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "躯干",
    "effect_id": "effect_24",
    "bonus_effect_ids": [],
    "designer_notes": "第2章防具，阶级4；导出时排除此策划说明列。"
  },
  "item_armor_09": {
    "name_text_id": "item_armor_09_name",
    "description_text_id": "item_armor_09_desc",
    "category": "防具",
    "rarity": "普通",
    "list_price": 260,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "躯干",
    "effect_id": "effect_34",
    "bonus_effect_ids": [],
    "designer_notes": "第3章防具，阶级1；导出时排除此策划说明列。"
  },
  "item_armor_10": {
    "name_text_id": "item_armor_10_name",
    "description_text_id": "item_armor_10_desc",
    "category": "防具",
    "rarity": "精良",
    "list_price": 280,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "躯干",
    "effect_id": "effect_35",
    "bonus_effect_ids": [],
    "designer_notes": "第3章防具，阶级2；导出时排除此策划说明列。"
  },
  "item_armor_11": {
    "name_text_id": "item_armor_11_name",
    "description_text_id": "item_armor_11_desc",
    "category": "防具",
    "rarity": "稀有",
    "list_price": 300,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "躯干",
    "effect_id": "effect_36",
    "bonus_effect_ids": [],
    "designer_notes": "第3章防具，阶级3；导出时排除此策划说明列。"
  },
  "item_armor_12": {
    "name_text_id": "item_armor_12_name",
    "description_text_id": "item_armor_12_desc",
    "category": "防具",
    "rarity": "史诗",
    "list_price": 320,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "躯干",
    "effect_id": "effect_36",
    "bonus_effect_ids": [],
    "designer_notes": "第3章防具，阶级4；导出时排除此策划说明列。"
  },
  "item_accessory_01": {
    "name_text_id": "item_accessory_01_name",
    "description_text_id": "item_accessory_01_desc",
    "category": "饰品",
    "rarity": "普通",
    "list_price": 140,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "饰品",
    "effect_id": "effect_10",
    "bonus_effect_ids": [],
    "designer_notes": "第1章饰品，阶级1；导出时排除此策划说明列。"
  },
  "item_accessory_02": {
    "name_text_id": "item_accessory_02_name",
    "description_text_id": "item_accessory_02_desc",
    "category": "饰品",
    "rarity": "精良",
    "list_price": 160,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "饰品",
    "effect_id": "effect_11",
    "bonus_effect_ids": [],
    "designer_notes": "第1章饰品，阶级2；导出时排除此策划说明列。"
  },
  "item_accessory_03": {
    "name_text_id": "item_accessory_03_name",
    "description_text_id": "item_accessory_03_desc",
    "category": "饰品",
    "rarity": "稀有",
    "list_price": 180,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "饰品",
    "effect_id": "effect_12",
    "bonus_effect_ids": [],
    "designer_notes": "第1章饰品，阶级3；导出时排除此策划说明列。"
  },
  "item_accessory_04": {
    "name_text_id": "item_accessory_04_name",
    "description_text_id": "item_accessory_04_desc",
    "category": "饰品",
    "rarity": "史诗",
    "list_price": 200,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "饰品",
    "effect_id": "effect_12",
    "bonus_effect_ids": [],
    "designer_notes": "第1章饰品，阶级4；导出时排除此策划说明列。"
  },
  "item_accessory_05": {
    "name_text_id": "item_accessory_05_name",
    "description_text_id": "item_accessory_05_desc",
    "category": "饰品",
    "rarity": "普通",
    "list_price": 200,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "饰品",
    "effect_id": "effect_22",
    "bonus_effect_ids": [],
    "designer_notes": "第2章饰品，阶级1；导出时排除此策划说明列。"
  },
  "item_accessory_06": {
    "name_text_id": "item_accessory_06_name",
    "description_text_id": "item_accessory_06_desc",
    "category": "饰品",
    "rarity": "精良",
    "list_price": 220,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "饰品",
    "effect_id": "effect_23",
    "bonus_effect_ids": [],
    "designer_notes": "第2章饰品，阶级2；导出时排除此策划说明列。"
  },
  "item_accessory_07": {
    "name_text_id": "item_accessory_07_name",
    "description_text_id": "item_accessory_07_desc",
    "category": "饰品",
    "rarity": "稀有",
    "list_price": 240,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "饰品",
    "effect_id": "effect_24",
    "bonus_effect_ids": [],
    "designer_notes": "第2章饰品，阶级3；导出时排除此策划说明列。"
  },
  "item_accessory_08": {
    "name_text_id": "item_accessory_08_name",
    "description_text_id": "item_accessory_08_desc",
    "category": "饰品",
    "rarity": "史诗",
    "list_price": 260,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "饰品",
    "effect_id": "effect_24",
    "bonus_effect_ids": [],
    "designer_notes": "第2章饰品，阶级4；导出时排除此策划说明列。"
  },
  "item_accessory_09": {
    "name_text_id": "item_accessory_09_name",
    "description_text_id": "item_accessory_09_desc",
    "category": "饰品",
    "rarity": "普通",
    "list_price": 260,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "饰品",
    "effect_id": "effect_34",
    "bonus_effect_ids": [],
    "designer_notes": "第3章饰品，阶级1；导出时排除此策划说明列。"
  },
  "item_accessory_10": {
    "name_text_id": "item_accessory_10_name",
    "description_text_id": "item_accessory_10_desc",
    "category": "饰品",
    "rarity": "精良",
    "list_price": 280,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "饰品",
    "effect_id": "effect_35",
    "bonus_effect_ids": [],
    "designer_notes": "第3章饰品，阶级2；导出时排除此策划说明列。"
  },
  "item_accessory_11": {
    "name_text_id": "item_accessory_11_name",
    "description_text_id": "item_accessory_11_desc",
    "category": "饰品",
    "rarity": "稀有",
    "list_price": 300,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "饰品",
    "effect_id": "effect_36",
    "bonus_effect_ids": [],
    "designer_notes": "第3章饰品，阶级3；导出时排除此策划说明列。"
  },
  "item_accessory_12": {
    "name_text_id": "item_accessory_12_name",
    "description_text_id": "item_accessory_12_desc",
    "category": "饰品",
    "rarity": "史诗",
    "list_price": 320,
    "tradable": true,
    "max_stack": 1,
    "equipment_slot": "饰品",
    "effect_id": "effect_36",
    "bonus_effect_ids": [],
    "designer_notes": "第3章饰品，阶级4；导出时排除此策划说明列。"
  },
  "item_quest_01": {
    "name_text_id": "item_quest_01_name",
    "description_text_id": "item_quest_01_desc",
    "category": "任务物品",
    "rarity": "普通",
    "list_price": 40,
    "tradable": false,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": null,
    "bonus_effect_ids": [],
    "designer_notes": "第1章任务物品，阶级1；导出时排除此策划说明列。"
  },
  "item_quest_02": {
    "name_text_id": "item_quest_02_name",
    "description_text_id": "item_quest_02_desc",
    "category": "任务物品",
    "rarity": "精良",
    "list_price": 60,
    "tradable": false,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": null,
    "bonus_effect_ids": [],
    "designer_notes": "第1章任务物品，阶级2；导出时排除此策划说明列。"
  },
  "item_quest_03": {
    "name_text_id": "item_quest_03_name",
    "description_text_id": "item_quest_03_desc",
    "category": "任务物品",
    "rarity": "稀有",
    "list_price": 80,
    "tradable": false,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": null,
    "bonus_effect_ids": [],
    "designer_notes": "第1章任务物品，阶级3；导出时排除此策划说明列。"
  },
  "item_quest_04": {
    "name_text_id": "item_quest_04_name",
    "description_text_id": "item_quest_04_desc",
    "category": "任务物品",
    "rarity": "史诗",
    "list_price": 100,
    "tradable": false,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": null,
    "bonus_effect_ids": [],
    "designer_notes": "第1章任务物品，阶级4；导出时排除此策划说明列。"
  },
  "item_quest_05": {
    "name_text_id": "item_quest_05_name",
    "description_text_id": "item_quest_05_desc",
    "category": "任务物品",
    "rarity": "普通",
    "list_price": 100,
    "tradable": false,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": null,
    "bonus_effect_ids": [],
    "designer_notes": "第2章任务物品，阶级1；导出时排除此策划说明列。"
  },
  "item_quest_06": {
    "name_text_id": "item_quest_06_name",
    "description_text_id": "item_quest_06_desc",
    "category": "任务物品",
    "rarity": "精良",
    "list_price": 120,
    "tradable": false,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": null,
    "bonus_effect_ids": [],
    "designer_notes": "第2章任务物品，阶级2；导出时排除此策划说明列。"
  },
  "item_quest_07": {
    "name_text_id": "item_quest_07_name",
    "description_text_id": "item_quest_07_desc",
    "category": "任务物品",
    "rarity": "稀有",
    "list_price": 140,
    "tradable": false,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": null,
    "bonus_effect_ids": [],
    "designer_notes": "第2章任务物品，阶级3；导出时排除此策划说明列。"
  },
  "item_quest_08": {
    "name_text_id": "item_quest_08_name",
    "description_text_id": "item_quest_08_desc",
    "category": "任务物品",
    "rarity": "史诗",
    "list_price": 160,
    "tradable": false,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": null,
    "bonus_effect_ids": [],
    "designer_notes": "第2章任务物品，阶级4；导出时排除此策划说明列。"
  },
  "item_quest_09": {
    "name_text_id": "item_quest_09_name",
    "description_text_id": "item_quest_09_desc",
    "category": "任务物品",
    "rarity": "普通",
    "list_price": 160,
    "tradable": false,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": null,
    "bonus_effect_ids": [],
    "designer_notes": "第3章任务物品，阶级1；导出时排除此策划说明列。"
  },
  "item_quest_10": {
    "name_text_id": "item_quest_10_name",
    "description_text_id": "item_quest_10_desc",
    "category": "任务物品",
    "rarity": "精良",
    "list_price": 180,
    "tradable": false,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": null,
    "bonus_effect_ids": [],
    "designer_notes": "第3章任务物品，阶级2；导出时排除此策划说明列。"
  },
  "item_quest_11": {
    "name_text_id": "item_quest_11_name",
    "description_text_id": "item_quest_11_desc",
    "category": "任务物品",
    "rarity": "稀有",
    "list_price": 200,
    "tradable": false,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": null,
    "bonus_effect_ids": [],
    "designer_notes": "第3章任务物品，阶级3；导出时排除此策划说明列。"
  },
  "item_quest_12": {
    "name_text_id": "item_quest_12_name",
    "description_text_id": "item_quest_12_desc",
    "category": "任务物品",
    "rarity": "史诗",
    "list_price": 220,
    "tradable": false,
    "max_stack": 99,
    "equipment_slot": null,
    "effect_id": null,
    "bonus_effect_ids": [],
    "designer_notes": "第3章任务物品，阶级4；导出时排除此策划说明列。"
  }
}

_tool: {
  "version": 1,
  "table": {
    "namespace": "content",
    "name": "item"
  },
  "foreignKeys": {
    "name_text_id": {
      "table": "localization.text"
    },
    "description_text_id": {
      "table": "localization.text"
    },
    "effect_id": {
      "table": "content.effect"
    },
    "bonus_effect_ids": {
      "table": "content.effect"
    }
  },
  "formulas": {}
}

_editor: {
  "table": {
    "title": "道具与装备"
  },
  "columns": {
    "name_text_id": {
      "title": "名称文本ID"
    },
    "description_text_id": {
      "title": "描述文本ID"
    },
    "category": {
      "title": "类别"
    },
    "rarity": {
      "title": "稀有度"
    },
    "list_price": {
      "title": "标价"
    },
    "tradable": {
      "title": "可交易"
    },
    "max_stack": {
      "title": "堆叠上限"
    },
    "equipment_slot": {
      "title": "装备槽位"
    },
    "effect_id": {
      "title": "效果ID"
    },
    "bonus_effect_ids": {
      "title": "附加效果列表"
    },
    "designer_notes": {
      "title": "策划备注"
    }
  },
  "cellComments": {
    "item_potion_01": {
      "tradable": "普通输入省略时使用Schema默认true。",
      "effect_id": "这瓶新手药剂使用行内默认表达式选回春effect_04；这是非字面量来源，目标改名会安全拒绝。"
    },
    "item_material_01": {
      "effect_id": "普通材料明确无效果：本列没有Schema默认值，null保持为空。"
    },
    "item_weapon_01": {
      "category": "装备也是道具：武器、防具、饰品用类别和槽位区分。",
      "bonus_effect_ids": "数组外键只保存效果ID；倍率不放平行数组。"
    },
    "item_quest_01": {
      "tradable": "任务信物禁止交易，客户端应尊重这条配置。"
    }
  }
}

_export: {
  "csv": {
    "enabled": true,
    "columns": ["id", "name_text_id", "description_text_id", "category", "rarity", "list_price", "tradable", "max_stack", "equipment_slot", "effect_id", "bonus_effect_ids"],
    "encoding": {
      "bonus_effect_ids": "json"
    }
  }
}
