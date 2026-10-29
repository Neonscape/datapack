package quest

import "example.com/datapack/yunlan-rpg/schema:common"

#PrimaryKey: common.#资源标识

#Row: {
    名称文本ID: common.#资源标识
    目标文本ID: common.#资源标识
    关卡ID: common.#资源标识
    前置任务ID: string | null
    类型: "讨伐" | "收集"
    目标道具ID: string | null
    需求数量: int & common.#正数
    奖励金币: int & common.#非负
    奖励道具ID: [...common.#资源标识]
}

rows: {
  "quest_01": {
    "名称文本ID": "quest_01_name",
    "目标文本ID": "quest_01_goal",
    "关卡ID": "stage_01",
    "前置任务ID": null,
    "类型": "讨伐",
    "目标道具ID": null,
    "需求数量": 3,
    "奖励金币": 60,
    "奖励道具ID": ["item_potion_01"]
  },
  "quest_02": {
    "名称文本ID": "quest_02_name",
    "目标文本ID": "quest_02_goal",
    "关卡ID": "stage_01",
    "前置任务ID": "quest_01",
    "类型": "收集",
    "目标道具ID": "item_material_01",
    "需求数量": 2,
    "奖励金币": 60,
    "奖励道具ID": ["item_potion_01"]
  },
  "quest_03": {
    "名称文本ID": "quest_03_name",
    "目标文本ID": "quest_03_goal",
    "关卡ID": "stage_02",
    "前置任务ID": "quest_02",
    "类型": "讨伐",
    "目标道具ID": null,
    "需求数量": 3,
    "奖励金币": 70,
    "奖励道具ID": ["item_potion_01"]
  },
  "quest_04": {
    "名称文本ID": "quest_04_name",
    "目标文本ID": "quest_04_goal",
    "关卡ID": "stage_02",
    "前置任务ID": "quest_03",
    "类型": "收集",
    "目标道具ID": "item_material_01",
    "需求数量": 2,
    "奖励金币": 70,
    "奖励道具ID": ["item_potion_01"]
  },
  "quest_05": {
    "名称文本ID": "quest_05_name",
    "目标文本ID": "quest_05_goal",
    "关卡ID": "stage_03",
    "前置任务ID": "quest_04",
    "类型": "讨伐",
    "目标道具ID": null,
    "需求数量": 3,
    "奖励金币": 80,
    "奖励道具ID": ["item_potion_01"]
  },
  "quest_06": {
    "名称文本ID": "quest_06_name",
    "目标文本ID": "quest_06_goal",
    "关卡ID": "stage_03",
    "前置任务ID": "quest_05",
    "类型": "收集",
    "目标道具ID": "item_material_02",
    "需求数量": 2,
    "奖励金币": 80,
    "奖励道具ID": ["item_potion_01"]
  },
  "quest_07": {
    "名称文本ID": "quest_07_name",
    "目标文本ID": "quest_07_goal",
    "关卡ID": "stage_04",
    "前置任务ID": "quest_06",
    "类型": "讨伐",
    "目标道具ID": null,
    "需求数量": 3,
    "奖励金币": 90,
    "奖励道具ID": ["item_potion_01"]
  },
  "quest_08": {
    "名称文本ID": "quest_08_name",
    "目标文本ID": "quest_08_goal",
    "关卡ID": "stage_04",
    "前置任务ID": "quest_07",
    "类型": "收集",
    "目标道具ID": "item_material_02",
    "需求数量": 2,
    "奖励金币": 90,
    "奖励道具ID": ["item_potion_01"]
  },
  "quest_09": {
    "名称文本ID": "quest_09_name",
    "目标文本ID": "quest_09_goal",
    "关卡ID": "stage_05",
    "前置任务ID": "quest_08",
    "类型": "讨伐",
    "目标道具ID": null,
    "需求数量": 3,
    "奖励金币": 100,
    "奖励道具ID": ["item_potion_01"]
  },
  "quest_10": {
    "名称文本ID": "quest_10_name",
    "目标文本ID": "quest_10_goal",
    "关卡ID": "stage_05",
    "前置任务ID": "quest_09",
    "类型": "收集",
    "目标道具ID": "item_material_03",
    "需求数量": 2,
    "奖励金币": 100,
    "奖励道具ID": ["item_potion_01"]
  },
  "quest_11": {
    "名称文本ID": "quest_11_name",
    "目标文本ID": "quest_11_goal",
    "关卡ID": "stage_06",
    "前置任务ID": "quest_10",
    "类型": "讨伐",
    "目标道具ID": null,
    "需求数量": 3,
    "奖励金币": 110,
    "奖励道具ID": ["item_potion_01"]
  },
  "quest_12": {
    "名称文本ID": "quest_12_name",
    "目标文本ID": "quest_12_goal",
    "关卡ID": "stage_06",
    "前置任务ID": "quest_11",
    "类型": "收集",
    "目标道具ID": "item_material_03",
    "需求数量": 2,
    "奖励金币": 110,
    "奖励道具ID": ["item_potion_01"]
  },
  "quest_13": {
    "名称文本ID": "quest_13_name",
    "目标文本ID": "quest_13_goal",
    "关卡ID": "stage_07",
    "前置任务ID": "quest_12",
    "类型": "讨伐",
    "目标道具ID": null,
    "需求数量": 3,
    "奖励金币": 120,
    "奖励道具ID": ["item_potion_01"]
  },
  "quest_14": {
    "名称文本ID": "quest_14_name",
    "目标文本ID": "quest_14_goal",
    "关卡ID": "stage_07",
    "前置任务ID": "quest_13",
    "类型": "收集",
    "目标道具ID": "item_material_04",
    "需求数量": 2,
    "奖励金币": 120,
    "奖励道具ID": ["item_potion_01"]
  },
  "quest_15": {
    "名称文本ID": "quest_15_name",
    "目标文本ID": "quest_15_goal",
    "关卡ID": "stage_08",
    "前置任务ID": "quest_14",
    "类型": "讨伐",
    "目标道具ID": null,
    "需求数量": 3,
    "奖励金币": 130,
    "奖励道具ID": ["item_potion_01"]
  },
  "quest_16": {
    "名称文本ID": "quest_16_name",
    "目标文本ID": "quest_16_goal",
    "关卡ID": "stage_08",
    "前置任务ID": "quest_15",
    "类型": "收集",
    "目标道具ID": "item_material_04",
    "需求数量": 2,
    "奖励金币": 130,
    "奖励道具ID": ["item_potion_01"]
  },
  "quest_17": {
    "名称文本ID": "quest_17_name",
    "目标文本ID": "quest_17_goal",
    "关卡ID": "stage_09",
    "前置任务ID": "quest_16",
    "类型": "讨伐",
    "目标道具ID": null,
    "需求数量": 3,
    "奖励金币": 140,
    "奖励道具ID": ["item_potion_05"]
  },
  "quest_18": {
    "名称文本ID": "quest_18_name",
    "目标文本ID": "quest_18_goal",
    "关卡ID": "stage_09",
    "前置任务ID": "quest_17",
    "类型": "收集",
    "目标道具ID": "item_material_05",
    "需求数量": 2,
    "奖励金币": 140,
    "奖励道具ID": ["item_potion_05"]
  },
  "quest_19": {
    "名称文本ID": "quest_19_name",
    "目标文本ID": "quest_19_goal",
    "关卡ID": "stage_10",
    "前置任务ID": "quest_18",
    "类型": "讨伐",
    "目标道具ID": null,
    "需求数量": 3,
    "奖励金币": 150,
    "奖励道具ID": ["item_potion_05"]
  },
  "quest_20": {
    "名称文本ID": "quest_20_name",
    "目标文本ID": "quest_20_goal",
    "关卡ID": "stage_10",
    "前置任务ID": "quest_19",
    "类型": "收集",
    "目标道具ID": "item_material_05",
    "需求数量": 2,
    "奖励金币": 150,
    "奖励道具ID": ["item_potion_05"]
  },
  "quest_21": {
    "名称文本ID": "quest_21_name",
    "目标文本ID": "quest_21_goal",
    "关卡ID": "stage_11",
    "前置任务ID": "quest_20",
    "类型": "讨伐",
    "目标道具ID": null,
    "需求数量": 3,
    "奖励金币": 160,
    "奖励道具ID": ["item_potion_05"]
  },
  "quest_22": {
    "名称文本ID": "quest_22_name",
    "目标文本ID": "quest_22_goal",
    "关卡ID": "stage_11",
    "前置任务ID": "quest_21",
    "类型": "收集",
    "目标道具ID": "item_material_06",
    "需求数量": 2,
    "奖励金币": 160,
    "奖励道具ID": ["item_potion_05"]
  },
  "quest_23": {
    "名称文本ID": "quest_23_name",
    "目标文本ID": "quest_23_goal",
    "关卡ID": "stage_12",
    "前置任务ID": "quest_22",
    "类型": "讨伐",
    "目标道具ID": null,
    "需求数量": 3,
    "奖励金币": 170,
    "奖励道具ID": ["item_potion_05"]
  },
  "quest_24": {
    "名称文本ID": "quest_24_name",
    "目标文本ID": "quest_24_goal",
    "关卡ID": "stage_12",
    "前置任务ID": "quest_23",
    "类型": "收集",
    "目标道具ID": "item_material_06",
    "需求数量": 2,
    "奖励金币": 170,
    "奖励道具ID": ["item_potion_05"]
  },
  "quest_25": {
    "名称文本ID": "quest_25_name",
    "目标文本ID": "quest_25_goal",
    "关卡ID": "stage_13",
    "前置任务ID": "quest_24",
    "类型": "讨伐",
    "目标道具ID": null,
    "需求数量": 3,
    "奖励金币": 180,
    "奖励道具ID": ["item_potion_05"]
  },
  "quest_26": {
    "名称文本ID": "quest_26_name",
    "目标文本ID": "quest_26_goal",
    "关卡ID": "stage_13",
    "前置任务ID": "quest_25",
    "类型": "收集",
    "目标道具ID": "item_material_07",
    "需求数量": 2,
    "奖励金币": 180,
    "奖励道具ID": ["item_potion_05"]
  },
  "quest_27": {
    "名称文本ID": "quest_27_name",
    "目标文本ID": "quest_27_goal",
    "关卡ID": "stage_14",
    "前置任务ID": "quest_26",
    "类型": "讨伐",
    "目标道具ID": null,
    "需求数量": 3,
    "奖励金币": 190,
    "奖励道具ID": ["item_potion_05"]
  },
  "quest_28": {
    "名称文本ID": "quest_28_name",
    "目标文本ID": "quest_28_goal",
    "关卡ID": "stage_14",
    "前置任务ID": "quest_27",
    "类型": "收集",
    "目标道具ID": "item_material_07",
    "需求数量": 2,
    "奖励金币": 190,
    "奖励道具ID": ["item_potion_05"]
  },
  "quest_29": {
    "名称文本ID": "quest_29_name",
    "目标文本ID": "quest_29_goal",
    "关卡ID": "stage_15",
    "前置任务ID": "quest_28",
    "类型": "讨伐",
    "目标道具ID": null,
    "需求数量": 3,
    "奖励金币": 200,
    "奖励道具ID": ["item_potion_05"]
  },
  "quest_30": {
    "名称文本ID": "quest_30_name",
    "目标文本ID": "quest_30_goal",
    "关卡ID": "stage_15",
    "前置任务ID": "quest_29",
    "类型": "收集",
    "目标道具ID": "item_material_08",
    "需求数量": 2,
    "奖励金币": 200,
    "奖励道具ID": ["item_potion_05"]
  },
  "quest_31": {
    "名称文本ID": "quest_31_name",
    "目标文本ID": "quest_31_goal",
    "关卡ID": "stage_16",
    "前置任务ID": "quest_30",
    "类型": "讨伐",
    "目标道具ID": null,
    "需求数量": 3,
    "奖励金币": 210,
    "奖励道具ID": ["item_potion_05"]
  },
  "quest_32": {
    "名称文本ID": "quest_32_name",
    "目标文本ID": "quest_32_goal",
    "关卡ID": "stage_16",
    "前置任务ID": "quest_31",
    "类型": "收集",
    "目标道具ID": "item_material_08",
    "需求数量": 2,
    "奖励金币": 210,
    "奖励道具ID": ["item_potion_05"]
  },
  "quest_33": {
    "名称文本ID": "quest_33_name",
    "目标文本ID": "quest_33_goal",
    "关卡ID": "stage_17",
    "前置任务ID": "quest_32",
    "类型": "讨伐",
    "目标道具ID": null,
    "需求数量": 3,
    "奖励金币": 220,
    "奖励道具ID": ["item_potion_09"]
  },
  "quest_34": {
    "名称文本ID": "quest_34_name",
    "目标文本ID": "quest_34_goal",
    "关卡ID": "stage_17",
    "前置任务ID": "quest_33",
    "类型": "收集",
    "目标道具ID": "item_material_09",
    "需求数量": 2,
    "奖励金币": 220,
    "奖励道具ID": ["item_potion_09"]
  },
  "quest_35": {
    "名称文本ID": "quest_35_name",
    "目标文本ID": "quest_35_goal",
    "关卡ID": "stage_18",
    "前置任务ID": "quest_34",
    "类型": "讨伐",
    "目标道具ID": null,
    "需求数量": 3,
    "奖励金币": 230,
    "奖励道具ID": ["item_potion_09"]
  },
  "quest_36": {
    "名称文本ID": "quest_36_name",
    "目标文本ID": "quest_36_goal",
    "关卡ID": "stage_18",
    "前置任务ID": "quest_35",
    "类型": "收集",
    "目标道具ID": "item_material_09",
    "需求数量": 2,
    "奖励金币": 230,
    "奖励道具ID": ["item_potion_09"]
  },
  "quest_37": {
    "名称文本ID": "quest_37_name",
    "目标文本ID": "quest_37_goal",
    "关卡ID": "stage_19",
    "前置任务ID": "quest_36",
    "类型": "讨伐",
    "目标道具ID": null,
    "需求数量": 3,
    "奖励金币": 240,
    "奖励道具ID": ["item_potion_09"]
  },
  "quest_38": {
    "名称文本ID": "quest_38_name",
    "目标文本ID": "quest_38_goal",
    "关卡ID": "stage_19",
    "前置任务ID": "quest_37",
    "类型": "收集",
    "目标道具ID": "item_material_10",
    "需求数量": 2,
    "奖励金币": 240,
    "奖励道具ID": ["item_potion_09"]
  },
  "quest_39": {
    "名称文本ID": "quest_39_name",
    "目标文本ID": "quest_39_goal",
    "关卡ID": "stage_20",
    "前置任务ID": "quest_38",
    "类型": "讨伐",
    "目标道具ID": null,
    "需求数量": 3,
    "奖励金币": 250,
    "奖励道具ID": ["item_potion_09"]
  },
  "quest_40": {
    "名称文本ID": "quest_40_name",
    "目标文本ID": "quest_40_goal",
    "关卡ID": "stage_20",
    "前置任务ID": "quest_39",
    "类型": "收集",
    "目标道具ID": "item_material_10",
    "需求数量": 2,
    "奖励金币": 250,
    "奖励道具ID": ["item_potion_09"]
  },
  "quest_41": {
    "名称文本ID": "quest_41_name",
    "目标文本ID": "quest_41_goal",
    "关卡ID": "stage_21",
    "前置任务ID": "quest_40",
    "类型": "讨伐",
    "目标道具ID": null,
    "需求数量": 3,
    "奖励金币": 260,
    "奖励道具ID": ["item_potion_09"]
  },
  "quest_42": {
    "名称文本ID": "quest_42_name",
    "目标文本ID": "quest_42_goal",
    "关卡ID": "stage_21",
    "前置任务ID": "quest_41",
    "类型": "收集",
    "目标道具ID": "item_material_11",
    "需求数量": 2,
    "奖励金币": 260,
    "奖励道具ID": ["item_potion_09"]
  },
  "quest_43": {
    "名称文本ID": "quest_43_name",
    "目标文本ID": "quest_43_goal",
    "关卡ID": "stage_22",
    "前置任务ID": "quest_42",
    "类型": "讨伐",
    "目标道具ID": null,
    "需求数量": 3,
    "奖励金币": 270,
    "奖励道具ID": ["item_potion_09"]
  },
  "quest_44": {
    "名称文本ID": "quest_44_name",
    "目标文本ID": "quest_44_goal",
    "关卡ID": "stage_22",
    "前置任务ID": "quest_43",
    "类型": "收集",
    "目标道具ID": "item_material_11",
    "需求数量": 2,
    "奖励金币": 270,
    "奖励道具ID": ["item_potion_09"]
  },
  "quest_45": {
    "名称文本ID": "quest_45_name",
    "目标文本ID": "quest_45_goal",
    "关卡ID": "stage_23",
    "前置任务ID": "quest_44",
    "类型": "讨伐",
    "目标道具ID": null,
    "需求数量": 3,
    "奖励金币": 280,
    "奖励道具ID": ["item_potion_09"]
  },
  "quest_46": {
    "名称文本ID": "quest_46_name",
    "目标文本ID": "quest_46_goal",
    "关卡ID": "stage_23",
    "前置任务ID": "quest_45",
    "类型": "收集",
    "目标道具ID": "item_material_12",
    "需求数量": 2,
    "奖励金币": 280,
    "奖励道具ID": ["item_potion_09"]
  },
  "quest_47": {
    "名称文本ID": "quest_47_name",
    "目标文本ID": "quest_47_goal",
    "关卡ID": "stage_24",
    "前置任务ID": "quest_46",
    "类型": "讨伐",
    "目标道具ID": null,
    "需求数量": 3,
    "奖励金币": 290,
    "奖励道具ID": ["item_potion_09"]
  },
  "quest_48": {
    "名称文本ID": "quest_48_name",
    "目标文本ID": "quest_48_goal",
    "关卡ID": "stage_24",
    "前置任务ID": "quest_47",
    "类型": "收集",
    "目标道具ID": "item_material_12",
    "需求数量": 2,
    "奖励金币": 290,
    "奖励道具ID": ["item_potion_09"]
  }
}

_tool: {
  "version": 1,
  "table": {
    "namespace": "quest",
    "name": "quest"
  },
  "foreignKeys": {
    "名称文本ID": {
      "table": "localization.text"
    },
    "目标文本ID": {
      "table": "localization.text"
    },
    "关卡ID": {
      "table": "world.stage"
    },
    "前置任务ID": {
      "table": "quest.quest"
    },
    "目标道具ID": {
      "table": "content.item"
    },
    "奖励道具ID": {
      "table": "content.item"
    }
  },
  "formulas": {}
}

_editor: {
  "table": {
    "title": "任务：推进旅程"
  },
  "cellComments": {
    "quest_01": {
      "前置任务ID": "第一条主线没有前置，明确为null。",
      "奖励道具ID": "此学习工程约定每个奖励ID发一件，不存在另一条平行数量数组。"
    },
    "quest_02": {
      "目标道具ID": "收集任务关联材料；讨伐任务此列为null。"
    }
  }
}

_export: {
  "csv": {
    "enabled": true,
    "columns": ["id", "名称文本ID", "目标文本ID", "关卡ID", "前置任务ID", "类型", "目标道具ID", "需求数量", "奖励金币", "奖励道具ID"],
    "encoding": {
      "奖励道具ID": "json"
    }
  }
}
