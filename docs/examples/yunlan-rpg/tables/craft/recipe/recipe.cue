package recipe

import "example.com/datapack/yunlan-rpg/schema:common"

#PrimaryKey: common.#资源标识

#Row: {
    名称文本ID: common.#资源标识
    产物ID: common.#资源标识
    产出数量: int & common.#正数
    工坊等级: int & >=1 & <=12
    金币消耗: int & common.#非负
    解锁任务ID: string | null
}

rows: {
  "recipe_01": {
    "名称文本ID": "recipe_01_name",
    "产物ID": "item_weapon_01",
    "产出数量": 1,
    "工坊等级": 1,
    "金币消耗": 90,
    "解锁任务ID": "quest_02"
  },
  "recipe_02": {
    "名称文本ID": "recipe_02_name",
    "产物ID": "item_weapon_02",
    "产出数量": 1,
    "工坊等级": 2,
    "金币消耗": 100,
    "解锁任务ID": "quest_06"
  },
  "recipe_03": {
    "名称文本ID": "recipe_03_name",
    "产物ID": "item_weapon_03",
    "产出数量": 1,
    "工坊等级": 3,
    "金币消耗": 110,
    "解锁任务ID": "quest_10"
  },
  "recipe_04": {
    "名称文本ID": "recipe_04_name",
    "产物ID": "item_weapon_04",
    "产出数量": 1,
    "工坊等级": 4,
    "金币消耗": 120,
    "解锁任务ID": "quest_14"
  },
  "recipe_05": {
    "名称文本ID": "recipe_05_name",
    "产物ID": "item_weapon_05",
    "产出数量": 1,
    "工坊等级": 5,
    "金币消耗": 130,
    "解锁任务ID": "quest_18"
  },
  "recipe_06": {
    "名称文本ID": "recipe_06_name",
    "产物ID": "item_weapon_06",
    "产出数量": 1,
    "工坊等级": 6,
    "金币消耗": 140,
    "解锁任务ID": "quest_22"
  },
  "recipe_07": {
    "名称文本ID": "recipe_07_name",
    "产物ID": "item_weapon_07",
    "产出数量": 1,
    "工坊等级": 7,
    "金币消耗": 150,
    "解锁任务ID": "quest_26"
  },
  "recipe_08": {
    "名称文本ID": "recipe_08_name",
    "产物ID": "item_weapon_08",
    "产出数量": 1,
    "工坊等级": 8,
    "金币消耗": 160,
    "解锁任务ID": "quest_30"
  },
  "recipe_09": {
    "名称文本ID": "recipe_09_name",
    "产物ID": "item_weapon_09",
    "产出数量": 1,
    "工坊等级": 9,
    "金币消耗": 170,
    "解锁任务ID": "quest_34"
  },
  "recipe_10": {
    "名称文本ID": "recipe_10_name",
    "产物ID": "item_weapon_10",
    "产出数量": 1,
    "工坊等级": 10,
    "金币消耗": 180,
    "解锁任务ID": "quest_38"
  },
  "recipe_11": {
    "名称文本ID": "recipe_11_name",
    "产物ID": "item_weapon_11",
    "产出数量": 1,
    "工坊等级": 11,
    "金币消耗": 190,
    "解锁任务ID": "quest_42"
  },
  "recipe_12": {
    "名称文本ID": "recipe_12_name",
    "产物ID": "item_weapon_12",
    "产出数量": 1,
    "工坊等级": 12,
    "金币消耗": 200,
    "解锁任务ID": "quest_46"
  },
  "recipe_13": {
    "名称文本ID": "recipe_13_name",
    "产物ID": "item_armor_01",
    "产出数量": 1,
    "工坊等级": 1,
    "金币消耗": 210,
    "解锁任务ID": "quest_02"
  },
  "recipe_14": {
    "名称文本ID": "recipe_14_name",
    "产物ID": "item_armor_02",
    "产出数量": 1,
    "工坊等级": 2,
    "金币消耗": 220,
    "解锁任务ID": "quest_06"
  },
  "recipe_15": {
    "名称文本ID": "recipe_15_name",
    "产物ID": "item_armor_03",
    "产出数量": 1,
    "工坊等级": 3,
    "金币消耗": 230,
    "解锁任务ID": "quest_10"
  },
  "recipe_16": {
    "名称文本ID": "recipe_16_name",
    "产物ID": "item_armor_04",
    "产出数量": 1,
    "工坊等级": 4,
    "金币消耗": 240,
    "解锁任务ID": "quest_14"
  },
  "recipe_17": {
    "名称文本ID": "recipe_17_name",
    "产物ID": "item_armor_05",
    "产出数量": 1,
    "工坊等级": 5,
    "金币消耗": 250,
    "解锁任务ID": "quest_18"
  },
  "recipe_18": {
    "名称文本ID": "recipe_18_name",
    "产物ID": "item_armor_06",
    "产出数量": 1,
    "工坊等级": 6,
    "金币消耗": 260,
    "解锁任务ID": "quest_22"
  },
  "recipe_19": {
    "名称文本ID": "recipe_19_name",
    "产物ID": "item_armor_07",
    "产出数量": 1,
    "工坊等级": 7,
    "金币消耗": 270,
    "解锁任务ID": "quest_26"
  },
  "recipe_20": {
    "名称文本ID": "recipe_20_name",
    "产物ID": "item_armor_08",
    "产出数量": 1,
    "工坊等级": 8,
    "金币消耗": 280,
    "解锁任务ID": "quest_30"
  },
  "recipe_21": {
    "名称文本ID": "recipe_21_name",
    "产物ID": "item_armor_09",
    "产出数量": 1,
    "工坊等级": 9,
    "金币消耗": 290,
    "解锁任务ID": "quest_34"
  },
  "recipe_22": {
    "名称文本ID": "recipe_22_name",
    "产物ID": "item_armor_10",
    "产出数量": 1,
    "工坊等级": 10,
    "金币消耗": 300,
    "解锁任务ID": "quest_38"
  },
  "recipe_23": {
    "名称文本ID": "recipe_23_name",
    "产物ID": "item_armor_11",
    "产出数量": 1,
    "工坊等级": 11,
    "金币消耗": 310,
    "解锁任务ID": "quest_42"
  },
  "recipe_24": {
    "名称文本ID": "recipe_24_name",
    "产物ID": "item_armor_12",
    "产出数量": 1,
    "工坊等级": 12,
    "金币消耗": 320,
    "解锁任务ID": "quest_46"
  },
  "recipe_25": {
    "名称文本ID": "recipe_25_name",
    "产物ID": "item_accessory_01",
    "产出数量": 1,
    "工坊等级": 1,
    "金币消耗": 330,
    "解锁任务ID": "quest_02"
  },
  "recipe_26": {
    "名称文本ID": "recipe_26_name",
    "产物ID": "item_accessory_02",
    "产出数量": 1,
    "工坊等级": 2,
    "金币消耗": 340,
    "解锁任务ID": "quest_06"
  },
  "recipe_27": {
    "名称文本ID": "recipe_27_name",
    "产物ID": "item_accessory_03",
    "产出数量": 1,
    "工坊等级": 3,
    "金币消耗": 350,
    "解锁任务ID": "quest_10"
  },
  "recipe_28": {
    "名称文本ID": "recipe_28_name",
    "产物ID": "item_accessory_04",
    "产出数量": 1,
    "工坊等级": 4,
    "金币消耗": 360,
    "解锁任务ID": "quest_14"
  },
  "recipe_29": {
    "名称文本ID": "recipe_29_name",
    "产物ID": "item_accessory_05",
    "产出数量": 1,
    "工坊等级": 5,
    "金币消耗": 370,
    "解锁任务ID": "quest_18"
  },
  "recipe_30": {
    "名称文本ID": "recipe_30_name",
    "产物ID": "item_accessory_06",
    "产出数量": 1,
    "工坊等级": 6,
    "金币消耗": 380,
    "解锁任务ID": "quest_22"
  },
  "recipe_31": {
    "名称文本ID": "recipe_31_name",
    "产物ID": "item_accessory_07",
    "产出数量": 1,
    "工坊等级": 7,
    "金币消耗": 390,
    "解锁任务ID": "quest_26"
  },
  "recipe_32": {
    "名称文本ID": "recipe_32_name",
    "产物ID": "item_accessory_08",
    "产出数量": 1,
    "工坊等级": 8,
    "金币消耗": 400,
    "解锁任务ID": "quest_30"
  },
  "recipe_33": {
    "名称文本ID": "recipe_33_name",
    "产物ID": "item_accessory_09",
    "产出数量": 1,
    "工坊等级": 9,
    "金币消耗": 410,
    "解锁任务ID": "quest_34"
  },
  "recipe_34": {
    "名称文本ID": "recipe_34_name",
    "产物ID": "item_accessory_10",
    "产出数量": 1,
    "工坊等级": 10,
    "金币消耗": 420,
    "解锁任务ID": "quest_38"
  },
  "recipe_35": {
    "名称文本ID": "recipe_35_name",
    "产物ID": "item_accessory_11",
    "产出数量": 1,
    "工坊等级": 11,
    "金币消耗": 430,
    "解锁任务ID": "quest_42"
  },
  "recipe_36": {
    "名称文本ID": "recipe_36_name",
    "产物ID": "item_accessory_12",
    "产出数量": 1,
    "工坊等级": 12,
    "金币消耗": 440,
    "解锁任务ID": "quest_46"
  }
}

_tool: {
  "version": 1,
  "table": {
    "namespace": "craft",
    "name": "recipe"
  },
  "foreignKeys": {
    "名称文本ID": {
      "table": "localization.text"
    },
    "产物ID": {
      "table": "content.item"
    },
    "解锁任务ID": {
      "table": "quest.quest"
    }
  },
  "formulas": {}
}

_editor: {
  "table": {
    "title": "合成配方：产物与解锁"
  },
  "cellComments": {
    "recipe_01": {
      "解锁任务ID": "任务→图样→材料→装备构成同一条玩法链。"
    }
  }
}

_export: {
  "csv": {
    "enabled": true,
    "columns": ["id", "名称文本ID", "产物ID", "产出数量", "工坊等级", "金币消耗", "解锁任务ID"],
    "encoding": {}
  }
}
