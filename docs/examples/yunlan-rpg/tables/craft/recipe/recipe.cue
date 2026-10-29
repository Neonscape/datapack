package recipe

import "example.com/datapack/yunlan-rpg/schema:common"

#PrimaryKey: common.#resource_id

#Row: {
    name_text_id: common.#resource_id
    output_item_id: common.#resource_id
    output_quantity: int & common.#positive
    workshop_level: int & >=1 & <=12
    gold_cost: int & common.#non_negative
    unlock_quest_id: string | null
}

rows: {
  "recipe_01": {
    "name_text_id": "recipe_01_name",
    "output_item_id": "item_weapon_01",
    "output_quantity": 1,
    "workshop_level": 1,
    "gold_cost": 90,
    "unlock_quest_id": "quest_02"
  },
  "recipe_02": {
    "name_text_id": "recipe_02_name",
    "output_item_id": "item_weapon_02",
    "output_quantity": 1,
    "workshop_level": 2,
    "gold_cost": 100,
    "unlock_quest_id": "quest_06"
  },
  "recipe_03": {
    "name_text_id": "recipe_03_name",
    "output_item_id": "item_weapon_03",
    "output_quantity": 1,
    "workshop_level": 3,
    "gold_cost": 110,
    "unlock_quest_id": "quest_10"
  },
  "recipe_04": {
    "name_text_id": "recipe_04_name",
    "output_item_id": "item_weapon_04",
    "output_quantity": 1,
    "workshop_level": 4,
    "gold_cost": 120,
    "unlock_quest_id": "quest_14"
  },
  "recipe_05": {
    "name_text_id": "recipe_05_name",
    "output_item_id": "item_weapon_05",
    "output_quantity": 1,
    "workshop_level": 5,
    "gold_cost": 130,
    "unlock_quest_id": "quest_18"
  },
  "recipe_06": {
    "name_text_id": "recipe_06_name",
    "output_item_id": "item_weapon_06",
    "output_quantity": 1,
    "workshop_level": 6,
    "gold_cost": 140,
    "unlock_quest_id": "quest_22"
  },
  "recipe_07": {
    "name_text_id": "recipe_07_name",
    "output_item_id": "item_weapon_07",
    "output_quantity": 1,
    "workshop_level": 7,
    "gold_cost": 150,
    "unlock_quest_id": "quest_26"
  },
  "recipe_08": {
    "name_text_id": "recipe_08_name",
    "output_item_id": "item_weapon_08",
    "output_quantity": 1,
    "workshop_level": 8,
    "gold_cost": 160,
    "unlock_quest_id": "quest_30"
  },
  "recipe_09": {
    "name_text_id": "recipe_09_name",
    "output_item_id": "item_weapon_09",
    "output_quantity": 1,
    "workshop_level": 9,
    "gold_cost": 170,
    "unlock_quest_id": "quest_34"
  },
  "recipe_10": {
    "name_text_id": "recipe_10_name",
    "output_item_id": "item_weapon_10",
    "output_quantity": 1,
    "workshop_level": 10,
    "gold_cost": 180,
    "unlock_quest_id": "quest_38"
  },
  "recipe_11": {
    "name_text_id": "recipe_11_name",
    "output_item_id": "item_weapon_11",
    "output_quantity": 1,
    "workshop_level": 11,
    "gold_cost": 190,
    "unlock_quest_id": "quest_42"
  },
  "recipe_12": {
    "name_text_id": "recipe_12_name",
    "output_item_id": "item_weapon_12",
    "output_quantity": 1,
    "workshop_level": 12,
    "gold_cost": 200,
    "unlock_quest_id": "quest_46"
  },
  "recipe_13": {
    "name_text_id": "recipe_13_name",
    "output_item_id": "item_armor_01",
    "output_quantity": 1,
    "workshop_level": 1,
    "gold_cost": 210,
    "unlock_quest_id": "quest_02"
  },
  "recipe_14": {
    "name_text_id": "recipe_14_name",
    "output_item_id": "item_armor_02",
    "output_quantity": 1,
    "workshop_level": 2,
    "gold_cost": 220,
    "unlock_quest_id": "quest_06"
  },
  "recipe_15": {
    "name_text_id": "recipe_15_name",
    "output_item_id": "item_armor_03",
    "output_quantity": 1,
    "workshop_level": 3,
    "gold_cost": 230,
    "unlock_quest_id": "quest_10"
  },
  "recipe_16": {
    "name_text_id": "recipe_16_name",
    "output_item_id": "item_armor_04",
    "output_quantity": 1,
    "workshop_level": 4,
    "gold_cost": 240,
    "unlock_quest_id": "quest_14"
  },
  "recipe_17": {
    "name_text_id": "recipe_17_name",
    "output_item_id": "item_armor_05",
    "output_quantity": 1,
    "workshop_level": 5,
    "gold_cost": 250,
    "unlock_quest_id": "quest_18"
  },
  "recipe_18": {
    "name_text_id": "recipe_18_name",
    "output_item_id": "item_armor_06",
    "output_quantity": 1,
    "workshop_level": 6,
    "gold_cost": 260,
    "unlock_quest_id": "quest_22"
  },
  "recipe_19": {
    "name_text_id": "recipe_19_name",
    "output_item_id": "item_armor_07",
    "output_quantity": 1,
    "workshop_level": 7,
    "gold_cost": 270,
    "unlock_quest_id": "quest_26"
  },
  "recipe_20": {
    "name_text_id": "recipe_20_name",
    "output_item_id": "item_armor_08",
    "output_quantity": 1,
    "workshop_level": 8,
    "gold_cost": 280,
    "unlock_quest_id": "quest_30"
  },
  "recipe_21": {
    "name_text_id": "recipe_21_name",
    "output_item_id": "item_armor_09",
    "output_quantity": 1,
    "workshop_level": 9,
    "gold_cost": 290,
    "unlock_quest_id": "quest_34"
  },
  "recipe_22": {
    "name_text_id": "recipe_22_name",
    "output_item_id": "item_armor_10",
    "output_quantity": 1,
    "workshop_level": 10,
    "gold_cost": 300,
    "unlock_quest_id": "quest_38"
  },
  "recipe_23": {
    "name_text_id": "recipe_23_name",
    "output_item_id": "item_armor_11",
    "output_quantity": 1,
    "workshop_level": 11,
    "gold_cost": 310,
    "unlock_quest_id": "quest_42"
  },
  "recipe_24": {
    "name_text_id": "recipe_24_name",
    "output_item_id": "item_armor_12",
    "output_quantity": 1,
    "workshop_level": 12,
    "gold_cost": 320,
    "unlock_quest_id": "quest_46"
  },
  "recipe_25": {
    "name_text_id": "recipe_25_name",
    "output_item_id": "item_accessory_01",
    "output_quantity": 1,
    "workshop_level": 1,
    "gold_cost": 330,
    "unlock_quest_id": "quest_02"
  },
  "recipe_26": {
    "name_text_id": "recipe_26_name",
    "output_item_id": "item_accessory_02",
    "output_quantity": 1,
    "workshop_level": 2,
    "gold_cost": 340,
    "unlock_quest_id": "quest_06"
  },
  "recipe_27": {
    "name_text_id": "recipe_27_name",
    "output_item_id": "item_accessory_03",
    "output_quantity": 1,
    "workshop_level": 3,
    "gold_cost": 350,
    "unlock_quest_id": "quest_10"
  },
  "recipe_28": {
    "name_text_id": "recipe_28_name",
    "output_item_id": "item_accessory_04",
    "output_quantity": 1,
    "workshop_level": 4,
    "gold_cost": 360,
    "unlock_quest_id": "quest_14"
  },
  "recipe_29": {
    "name_text_id": "recipe_29_name",
    "output_item_id": "item_accessory_05",
    "output_quantity": 1,
    "workshop_level": 5,
    "gold_cost": 370,
    "unlock_quest_id": "quest_18"
  },
  "recipe_30": {
    "name_text_id": "recipe_30_name",
    "output_item_id": "item_accessory_06",
    "output_quantity": 1,
    "workshop_level": 6,
    "gold_cost": 380,
    "unlock_quest_id": "quest_22"
  },
  "recipe_31": {
    "name_text_id": "recipe_31_name",
    "output_item_id": "item_accessory_07",
    "output_quantity": 1,
    "workshop_level": 7,
    "gold_cost": 390,
    "unlock_quest_id": "quest_26"
  },
  "recipe_32": {
    "name_text_id": "recipe_32_name",
    "output_item_id": "item_accessory_08",
    "output_quantity": 1,
    "workshop_level": 8,
    "gold_cost": 400,
    "unlock_quest_id": "quest_30"
  },
  "recipe_33": {
    "name_text_id": "recipe_33_name",
    "output_item_id": "item_accessory_09",
    "output_quantity": 1,
    "workshop_level": 9,
    "gold_cost": 410,
    "unlock_quest_id": "quest_34"
  },
  "recipe_34": {
    "name_text_id": "recipe_34_name",
    "output_item_id": "item_accessory_10",
    "output_quantity": 1,
    "workshop_level": 10,
    "gold_cost": 420,
    "unlock_quest_id": "quest_38"
  },
  "recipe_35": {
    "name_text_id": "recipe_35_name",
    "output_item_id": "item_accessory_11",
    "output_quantity": 1,
    "workshop_level": 11,
    "gold_cost": 430,
    "unlock_quest_id": "quest_42"
  },
  "recipe_36": {
    "name_text_id": "recipe_36_name",
    "output_item_id": "item_accessory_12",
    "output_quantity": 1,
    "workshop_level": 12,
    "gold_cost": 440,
    "unlock_quest_id": "quest_46"
  }
}

_tool: {
  "version": 1,
  "table": {
    "namespace": "craft",
    "name": "recipe"
  },
  "foreignKeys": {
    "name_text_id": {
      "table": "localization.text"
    },
    "output_item_id": {
      "table": "content.item"
    },
    "unlock_quest_id": {
      "table": "quest.quest"
    }
  },
  "formulas": {}
}

_editor: {
  "table": {
    "title": "合成配方：产物与解锁"
  },
  "columns": {
    "name_text_id": {
      "title": "名称文本ID"
    },
    "output_item_id": {
      "title": "产物ID"
    },
    "output_quantity": {
      "title": "产出数量"
    },
    "workshop_level": {
      "title": "工坊等级"
    },
    "gold_cost": {
      "title": "金币消耗"
    },
    "unlock_quest_id": {
      "title": "解锁任务ID"
    }
  },
  "cellComments": {
    "recipe_01": {
      "unlock_quest_id": "任务→图样→材料→装备构成同一条玩法链。"
    }
  }
}

_export: {
  "csv": {
    "enabled": true,
    "columns": ["id", "name_text_id", "output_item_id", "output_quantity", "workshop_level", "gold_cost", "unlock_quest_id"],
    "encoding": {}
  }
}
