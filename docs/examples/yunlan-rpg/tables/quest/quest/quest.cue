package quest

import "example.com/datapack/yunlan-rpg/schema:common"

#PrimaryKey: common.#resource_id

#Row: {
    name_text_id: common.#resource_id
    objective_text_id: common.#resource_id
    stage_id: common.#resource_id
    prerequisite_quest_id: string | null
    quest_type: "讨伐" | "收集"
    target_item_id: string | null
    required_quantity: int & common.#positive
    reward_gold: int & common.#non_negative
    reward_item_ids: [...common.#resource_id]
}

rows: {
  "quest_01": {
    "name_text_id": "quest_01_name",
    "objective_text_id": "quest_01_goal",
    "stage_id": "stage_01",
    "prerequisite_quest_id": null,
    "quest_type": "讨伐",
    "target_item_id": null,
    "required_quantity": 3,
    "reward_gold": 60,
    "reward_item_ids": ["item_potion_01"]
  },
  "quest_02": {
    "name_text_id": "quest_02_name",
    "objective_text_id": "quest_02_goal",
    "stage_id": "stage_01",
    "prerequisite_quest_id": "quest_01",
    "quest_type": "收集",
    "target_item_id": "item_material_01",
    "required_quantity": 2,
    "reward_gold": 60,
    "reward_item_ids": ["item_potion_01"]
  },
  "quest_03": {
    "name_text_id": "quest_03_name",
    "objective_text_id": "quest_03_goal",
    "stage_id": "stage_02",
    "prerequisite_quest_id": "quest_02",
    "quest_type": "讨伐",
    "target_item_id": null,
    "required_quantity": 3,
    "reward_gold": 70,
    "reward_item_ids": ["item_potion_01"]
  },
  "quest_04": {
    "name_text_id": "quest_04_name",
    "objective_text_id": "quest_04_goal",
    "stage_id": "stage_02",
    "prerequisite_quest_id": "quest_03",
    "quest_type": "收集",
    "target_item_id": "item_material_01",
    "required_quantity": 2,
    "reward_gold": 70,
    "reward_item_ids": ["item_potion_01"]
  },
  "quest_05": {
    "name_text_id": "quest_05_name",
    "objective_text_id": "quest_05_goal",
    "stage_id": "stage_03",
    "prerequisite_quest_id": "quest_04",
    "quest_type": "讨伐",
    "target_item_id": null,
    "required_quantity": 3,
    "reward_gold": 80,
    "reward_item_ids": ["item_potion_01"]
  },
  "quest_06": {
    "name_text_id": "quest_06_name",
    "objective_text_id": "quest_06_goal",
    "stage_id": "stage_03",
    "prerequisite_quest_id": "quest_05",
    "quest_type": "收集",
    "target_item_id": "item_material_02",
    "required_quantity": 2,
    "reward_gold": 80,
    "reward_item_ids": ["item_potion_01"]
  },
  "quest_07": {
    "name_text_id": "quest_07_name",
    "objective_text_id": "quest_07_goal",
    "stage_id": "stage_04",
    "prerequisite_quest_id": "quest_06",
    "quest_type": "讨伐",
    "target_item_id": null,
    "required_quantity": 3,
    "reward_gold": 90,
    "reward_item_ids": ["item_potion_01"]
  },
  "quest_08": {
    "name_text_id": "quest_08_name",
    "objective_text_id": "quest_08_goal",
    "stage_id": "stage_04",
    "prerequisite_quest_id": "quest_07",
    "quest_type": "收集",
    "target_item_id": "item_material_02",
    "required_quantity": 2,
    "reward_gold": 90,
    "reward_item_ids": ["item_potion_01"]
  },
  "quest_09": {
    "name_text_id": "quest_09_name",
    "objective_text_id": "quest_09_goal",
    "stage_id": "stage_05",
    "prerequisite_quest_id": "quest_08",
    "quest_type": "讨伐",
    "target_item_id": null,
    "required_quantity": 3,
    "reward_gold": 100,
    "reward_item_ids": ["item_potion_01"]
  },
  "quest_10": {
    "name_text_id": "quest_10_name",
    "objective_text_id": "quest_10_goal",
    "stage_id": "stage_05",
    "prerequisite_quest_id": "quest_09",
    "quest_type": "收集",
    "target_item_id": "item_material_03",
    "required_quantity": 2,
    "reward_gold": 100,
    "reward_item_ids": ["item_potion_01"]
  },
  "quest_11": {
    "name_text_id": "quest_11_name",
    "objective_text_id": "quest_11_goal",
    "stage_id": "stage_06",
    "prerequisite_quest_id": "quest_10",
    "quest_type": "讨伐",
    "target_item_id": null,
    "required_quantity": 3,
    "reward_gold": 110,
    "reward_item_ids": ["item_potion_01"]
  },
  "quest_12": {
    "name_text_id": "quest_12_name",
    "objective_text_id": "quest_12_goal",
    "stage_id": "stage_06",
    "prerequisite_quest_id": "quest_11",
    "quest_type": "收集",
    "target_item_id": "item_material_03",
    "required_quantity": 2,
    "reward_gold": 110,
    "reward_item_ids": ["item_potion_01"]
  },
  "quest_13": {
    "name_text_id": "quest_13_name",
    "objective_text_id": "quest_13_goal",
    "stage_id": "stage_07",
    "prerequisite_quest_id": "quest_12",
    "quest_type": "讨伐",
    "target_item_id": null,
    "required_quantity": 3,
    "reward_gold": 120,
    "reward_item_ids": ["item_potion_01"]
  },
  "quest_14": {
    "name_text_id": "quest_14_name",
    "objective_text_id": "quest_14_goal",
    "stage_id": "stage_07",
    "prerequisite_quest_id": "quest_13",
    "quest_type": "收集",
    "target_item_id": "item_material_04",
    "required_quantity": 2,
    "reward_gold": 120,
    "reward_item_ids": ["item_potion_01"]
  },
  "quest_15": {
    "name_text_id": "quest_15_name",
    "objective_text_id": "quest_15_goal",
    "stage_id": "stage_08",
    "prerequisite_quest_id": "quest_14",
    "quest_type": "讨伐",
    "target_item_id": null,
    "required_quantity": 3,
    "reward_gold": 130,
    "reward_item_ids": ["item_potion_01"]
  },
  "quest_16": {
    "name_text_id": "quest_16_name",
    "objective_text_id": "quest_16_goal",
    "stage_id": "stage_08",
    "prerequisite_quest_id": "quest_15",
    "quest_type": "收集",
    "target_item_id": "item_material_04",
    "required_quantity": 2,
    "reward_gold": 130,
    "reward_item_ids": ["item_potion_01"]
  },
  "quest_17": {
    "name_text_id": "quest_17_name",
    "objective_text_id": "quest_17_goal",
    "stage_id": "stage_09",
    "prerequisite_quest_id": "quest_16",
    "quest_type": "讨伐",
    "target_item_id": null,
    "required_quantity": 3,
    "reward_gold": 140,
    "reward_item_ids": ["item_potion_05"]
  },
  "quest_18": {
    "name_text_id": "quest_18_name",
    "objective_text_id": "quest_18_goal",
    "stage_id": "stage_09",
    "prerequisite_quest_id": "quest_17",
    "quest_type": "收集",
    "target_item_id": "item_material_05",
    "required_quantity": 2,
    "reward_gold": 140,
    "reward_item_ids": ["item_potion_05"]
  },
  "quest_19": {
    "name_text_id": "quest_19_name",
    "objective_text_id": "quest_19_goal",
    "stage_id": "stage_10",
    "prerequisite_quest_id": "quest_18",
    "quest_type": "讨伐",
    "target_item_id": null,
    "required_quantity": 3,
    "reward_gold": 150,
    "reward_item_ids": ["item_potion_05"]
  },
  "quest_20": {
    "name_text_id": "quest_20_name",
    "objective_text_id": "quest_20_goal",
    "stage_id": "stage_10",
    "prerequisite_quest_id": "quest_19",
    "quest_type": "收集",
    "target_item_id": "item_material_05",
    "required_quantity": 2,
    "reward_gold": 150,
    "reward_item_ids": ["item_potion_05"]
  },
  "quest_21": {
    "name_text_id": "quest_21_name",
    "objective_text_id": "quest_21_goal",
    "stage_id": "stage_11",
    "prerequisite_quest_id": "quest_20",
    "quest_type": "讨伐",
    "target_item_id": null,
    "required_quantity": 3,
    "reward_gold": 160,
    "reward_item_ids": ["item_potion_05"]
  },
  "quest_22": {
    "name_text_id": "quest_22_name",
    "objective_text_id": "quest_22_goal",
    "stage_id": "stage_11",
    "prerequisite_quest_id": "quest_21",
    "quest_type": "收集",
    "target_item_id": "item_material_06",
    "required_quantity": 2,
    "reward_gold": 160,
    "reward_item_ids": ["item_potion_05"]
  },
  "quest_23": {
    "name_text_id": "quest_23_name",
    "objective_text_id": "quest_23_goal",
    "stage_id": "stage_12",
    "prerequisite_quest_id": "quest_22",
    "quest_type": "讨伐",
    "target_item_id": null,
    "required_quantity": 3,
    "reward_gold": 170,
    "reward_item_ids": ["item_potion_05"]
  },
  "quest_24": {
    "name_text_id": "quest_24_name",
    "objective_text_id": "quest_24_goal",
    "stage_id": "stage_12",
    "prerequisite_quest_id": "quest_23",
    "quest_type": "收集",
    "target_item_id": "item_material_06",
    "required_quantity": 2,
    "reward_gold": 170,
    "reward_item_ids": ["item_potion_05"]
  },
  "quest_25": {
    "name_text_id": "quest_25_name",
    "objective_text_id": "quest_25_goal",
    "stage_id": "stage_13",
    "prerequisite_quest_id": "quest_24",
    "quest_type": "讨伐",
    "target_item_id": null,
    "required_quantity": 3,
    "reward_gold": 180,
    "reward_item_ids": ["item_potion_05"]
  },
  "quest_26": {
    "name_text_id": "quest_26_name",
    "objective_text_id": "quest_26_goal",
    "stage_id": "stage_13",
    "prerequisite_quest_id": "quest_25",
    "quest_type": "收集",
    "target_item_id": "item_material_07",
    "required_quantity": 2,
    "reward_gold": 180,
    "reward_item_ids": ["item_potion_05"]
  },
  "quest_27": {
    "name_text_id": "quest_27_name",
    "objective_text_id": "quest_27_goal",
    "stage_id": "stage_14",
    "prerequisite_quest_id": "quest_26",
    "quest_type": "讨伐",
    "target_item_id": null,
    "required_quantity": 3,
    "reward_gold": 190,
    "reward_item_ids": ["item_potion_05"]
  },
  "quest_28": {
    "name_text_id": "quest_28_name",
    "objective_text_id": "quest_28_goal",
    "stage_id": "stage_14",
    "prerequisite_quest_id": "quest_27",
    "quest_type": "收集",
    "target_item_id": "item_material_07",
    "required_quantity": 2,
    "reward_gold": 190,
    "reward_item_ids": ["item_potion_05"]
  },
  "quest_29": {
    "name_text_id": "quest_29_name",
    "objective_text_id": "quest_29_goal",
    "stage_id": "stage_15",
    "prerequisite_quest_id": "quest_28",
    "quest_type": "讨伐",
    "target_item_id": null,
    "required_quantity": 3,
    "reward_gold": 200,
    "reward_item_ids": ["item_potion_05"]
  },
  "quest_30": {
    "name_text_id": "quest_30_name",
    "objective_text_id": "quest_30_goal",
    "stage_id": "stage_15",
    "prerequisite_quest_id": "quest_29",
    "quest_type": "收集",
    "target_item_id": "item_material_08",
    "required_quantity": 2,
    "reward_gold": 200,
    "reward_item_ids": ["item_potion_05"]
  },
  "quest_31": {
    "name_text_id": "quest_31_name",
    "objective_text_id": "quest_31_goal",
    "stage_id": "stage_16",
    "prerequisite_quest_id": "quest_30",
    "quest_type": "讨伐",
    "target_item_id": null,
    "required_quantity": 3,
    "reward_gold": 210,
    "reward_item_ids": ["item_potion_05"]
  },
  "quest_32": {
    "name_text_id": "quest_32_name",
    "objective_text_id": "quest_32_goal",
    "stage_id": "stage_16",
    "prerequisite_quest_id": "quest_31",
    "quest_type": "收集",
    "target_item_id": "item_material_08",
    "required_quantity": 2,
    "reward_gold": 210,
    "reward_item_ids": ["item_potion_05"]
  },
  "quest_33": {
    "name_text_id": "quest_33_name",
    "objective_text_id": "quest_33_goal",
    "stage_id": "stage_17",
    "prerequisite_quest_id": "quest_32",
    "quest_type": "讨伐",
    "target_item_id": null,
    "required_quantity": 3,
    "reward_gold": 220,
    "reward_item_ids": ["item_potion_09"]
  },
  "quest_34": {
    "name_text_id": "quest_34_name",
    "objective_text_id": "quest_34_goal",
    "stage_id": "stage_17",
    "prerequisite_quest_id": "quest_33",
    "quest_type": "收集",
    "target_item_id": "item_material_09",
    "required_quantity": 2,
    "reward_gold": 220,
    "reward_item_ids": ["item_potion_09"]
  },
  "quest_35": {
    "name_text_id": "quest_35_name",
    "objective_text_id": "quest_35_goal",
    "stage_id": "stage_18",
    "prerequisite_quest_id": "quest_34",
    "quest_type": "讨伐",
    "target_item_id": null,
    "required_quantity": 3,
    "reward_gold": 230,
    "reward_item_ids": ["item_potion_09"]
  },
  "quest_36": {
    "name_text_id": "quest_36_name",
    "objective_text_id": "quest_36_goal",
    "stage_id": "stage_18",
    "prerequisite_quest_id": "quest_35",
    "quest_type": "收集",
    "target_item_id": "item_material_09",
    "required_quantity": 2,
    "reward_gold": 230,
    "reward_item_ids": ["item_potion_09"]
  },
  "quest_37": {
    "name_text_id": "quest_37_name",
    "objective_text_id": "quest_37_goal",
    "stage_id": "stage_19",
    "prerequisite_quest_id": "quest_36",
    "quest_type": "讨伐",
    "target_item_id": null,
    "required_quantity": 3,
    "reward_gold": 240,
    "reward_item_ids": ["item_potion_09"]
  },
  "quest_38": {
    "name_text_id": "quest_38_name",
    "objective_text_id": "quest_38_goal",
    "stage_id": "stage_19",
    "prerequisite_quest_id": "quest_37",
    "quest_type": "收集",
    "target_item_id": "item_material_10",
    "required_quantity": 2,
    "reward_gold": 240,
    "reward_item_ids": ["item_potion_09"]
  },
  "quest_39": {
    "name_text_id": "quest_39_name",
    "objective_text_id": "quest_39_goal",
    "stage_id": "stage_20",
    "prerequisite_quest_id": "quest_38",
    "quest_type": "讨伐",
    "target_item_id": null,
    "required_quantity": 3,
    "reward_gold": 250,
    "reward_item_ids": ["item_potion_09"]
  },
  "quest_40": {
    "name_text_id": "quest_40_name",
    "objective_text_id": "quest_40_goal",
    "stage_id": "stage_20",
    "prerequisite_quest_id": "quest_39",
    "quest_type": "收集",
    "target_item_id": "item_material_10",
    "required_quantity": 2,
    "reward_gold": 250,
    "reward_item_ids": ["item_potion_09"]
  },
  "quest_41": {
    "name_text_id": "quest_41_name",
    "objective_text_id": "quest_41_goal",
    "stage_id": "stage_21",
    "prerequisite_quest_id": "quest_40",
    "quest_type": "讨伐",
    "target_item_id": null,
    "required_quantity": 3,
    "reward_gold": 260,
    "reward_item_ids": ["item_potion_09"]
  },
  "quest_42": {
    "name_text_id": "quest_42_name",
    "objective_text_id": "quest_42_goal",
    "stage_id": "stage_21",
    "prerequisite_quest_id": "quest_41",
    "quest_type": "收集",
    "target_item_id": "item_material_11",
    "required_quantity": 2,
    "reward_gold": 260,
    "reward_item_ids": ["item_potion_09"]
  },
  "quest_43": {
    "name_text_id": "quest_43_name",
    "objective_text_id": "quest_43_goal",
    "stage_id": "stage_22",
    "prerequisite_quest_id": "quest_42",
    "quest_type": "讨伐",
    "target_item_id": null,
    "required_quantity": 3,
    "reward_gold": 270,
    "reward_item_ids": ["item_potion_09"]
  },
  "quest_44": {
    "name_text_id": "quest_44_name",
    "objective_text_id": "quest_44_goal",
    "stage_id": "stage_22",
    "prerequisite_quest_id": "quest_43",
    "quest_type": "收集",
    "target_item_id": "item_material_11",
    "required_quantity": 2,
    "reward_gold": 270,
    "reward_item_ids": ["item_potion_09"]
  },
  "quest_45": {
    "name_text_id": "quest_45_name",
    "objective_text_id": "quest_45_goal",
    "stage_id": "stage_23",
    "prerequisite_quest_id": "quest_44",
    "quest_type": "讨伐",
    "target_item_id": null,
    "required_quantity": 3,
    "reward_gold": 280,
    "reward_item_ids": ["item_potion_09"]
  },
  "quest_46": {
    "name_text_id": "quest_46_name",
    "objective_text_id": "quest_46_goal",
    "stage_id": "stage_23",
    "prerequisite_quest_id": "quest_45",
    "quest_type": "收集",
    "target_item_id": "item_material_12",
    "required_quantity": 2,
    "reward_gold": 280,
    "reward_item_ids": ["item_potion_09"]
  },
  "quest_47": {
    "name_text_id": "quest_47_name",
    "objective_text_id": "quest_47_goal",
    "stage_id": "stage_24",
    "prerequisite_quest_id": "quest_46",
    "quest_type": "讨伐",
    "target_item_id": null,
    "required_quantity": 3,
    "reward_gold": 290,
    "reward_item_ids": ["item_potion_09"]
  },
  "quest_48": {
    "name_text_id": "quest_48_name",
    "objective_text_id": "quest_48_goal",
    "stage_id": "stage_24",
    "prerequisite_quest_id": "quest_47",
    "quest_type": "收集",
    "target_item_id": "item_material_12",
    "required_quantity": 2,
    "reward_gold": 290,
    "reward_item_ids": ["item_potion_09"]
  }
}

_tool: {
  "version": 1,
  "table": {
    "namespace": "quest",
    "name": "quest"
  },
  "foreignKeys": {
    "name_text_id": {
      "table": "localization.text"
    },
    "objective_text_id": {
      "table": "localization.text"
    },
    "stage_id": {
      "table": "world.stage"
    },
    "prerequisite_quest_id": {
      "table": "quest.quest"
    },
    "target_item_id": {
      "table": "content.item"
    },
    "reward_item_ids": {
      "table": "content.item"
    }
  },
  "formulas": {}
}

_editor: {
  "table": {
    "title": "任务：推进旅程"
  },
  "columns": {
    "name_text_id": {
      "title": "名称文本ID"
    },
    "objective_text_id": {
      "title": "目标文本ID"
    },
    "stage_id": {
      "title": "关卡ID"
    },
    "prerequisite_quest_id": {
      "title": "前置任务ID"
    },
    "quest_type": {
      "title": "类型"
    },
    "target_item_id": {
      "title": "目标道具ID"
    },
    "required_quantity": {
      "title": "需求数量"
    },
    "reward_gold": {
      "title": "奖励金币"
    },
    "reward_item_ids": {
      "title": "奖励道具列表"
    }
  },
  "cellComments": {
    "quest_01": {
      "prerequisite_quest_id": "第一条主线没有前置，明确为null。",
      "reward_item_ids": "此学习工程约定每个奖励ID发一件，不存在另一条平行数量数组。"
    },
    "quest_02": {
      "target_item_id": "收集任务关联材料；讨伐任务此列为null。"
    }
  }
}

_export: {
  "csv": {
    "enabled": true,
    "columns": ["id", "name_text_id", "objective_text_id", "stage_id", "prerequisite_quest_id", "quest_type", "target_item_id", "required_quantity", "reward_gold", "reward_item_ids"],
    "encoding": {
      "reward_item_ids": "json"
    }
  }
}
