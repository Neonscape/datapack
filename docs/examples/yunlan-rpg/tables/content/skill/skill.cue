package skill

import "example.com/datapack/yunlan-rpg/schema:common"

#PrimaryKey: common.#resource_id

#Row: {
    name_text_id: common.#resource_id
    description_text_id: string | null
    unlock_level: int & >=1 & <=60
    mana_cost: int & common.#non_negative
    cooldown_turns: int & common.#non_negative
    primary_effect_id: common.#resource_id
    bonus_effect_ids: [...common.#resource_id]
    multiplier: number & common.#positive
    tags: [...string]
}

rows: {
  "skill_01": {
    "name_text_id": "skill_01_name",
    "description_text_id": "skill_01_desc",
    "unlock_level": 1,
    "mana_cost": 5,
    "cooldown_turns": 0,
    "primary_effect_id": "effect_01",
    "bonus_effect_ids": [],
    "multiplier": 1.0,
    "tags": ["主动", "战士"]
  },
  "skill_02": {
    "name_text_id": "skill_02_name",
    "description_text_id": "skill_02_desc",
    "unlock_level": 6,
    "mana_cost": 13,
    "cooldown_turns": 1,
    "primary_effect_id": "effect_11",
    "bonus_effect_ids": [],
    "multiplier": 1.25,
    "tags": ["主动", "战士"]
  },
  "skill_03": {
    "name_text_id": "skill_03_name",
    "description_text_id": "skill_03_desc",
    "unlock_level": 11,
    "mana_cost": 0,
    "cooldown_turns": 0,
    "primary_effect_id": "effect_12",
    "bonus_effect_ids": [],
    "multiplier": 1.5,
    "tags": ["被动", "战士"]
  },
  "skill_04": {
    "name_text_id": "skill_04_name",
    "description_text_id": "skill_04_desc",
    "unlock_level": 16,
    "mana_cost": 29,
    "cooldown_turns": 3,
    "primary_effect_id": "effect_09",
    "bonus_effect_ids": ["effect_10"],
    "multiplier": 1.75,
    "tags": ["主动", "战士"]
  },
  "skill_05": {
    "name_text_id": "skill_05_name",
    "description_text_id": "skill_05_desc",
    "unlock_level": 1,
    "mana_cost": 5,
    "cooldown_turns": 0,
    "primary_effect_id": "effect_01",
    "bonus_effect_ids": [],
    "multiplier": 1.0,
    "tags": ["主动", "游侠"]
  },
  "skill_06": {
    "name_text_id": "skill_06_name",
    "description_text_id": "skill_06_desc",
    "unlock_level": 6,
    "mana_cost": 13,
    "cooldown_turns": 1,
    "primary_effect_id": "effect_08",
    "bonus_effect_ids": [],
    "multiplier": 1.25,
    "tags": ["主动", "游侠"]
  },
  "skill_07": {
    "name_text_id": "skill_07_name",
    "description_text_id": "skill_07_desc",
    "unlock_level": 11,
    "mana_cost": 0,
    "cooldown_turns": 0,
    "primary_effect_id": "effect_12",
    "bonus_effect_ids": [],
    "multiplier": 1.5,
    "tags": ["被动", "游侠"]
  },
  "skill_08": {
    "name_text_id": "skill_08_name",
    "description_text_id": "skill_08_desc",
    "unlock_level": 16,
    "mana_cost": 29,
    "cooldown_turns": 3,
    "primary_effect_id": "effect_03",
    "bonus_effect_ids": ["effect_10"],
    "multiplier": 1.75,
    "tags": ["主动", "游侠"]
  },
  "skill_09": {
    "name_text_id": "skill_09_name",
    "description_text_id": "skill_09_desc",
    "unlock_level": 1,
    "mana_cost": 5,
    "cooldown_turns": 0,
    "primary_effect_id": "effect_04",
    "bonus_effect_ids": [],
    "multiplier": 1.0,
    "tags": ["主动", "医师"]
  },
  "skill_10": {
    "name_text_id": "skill_10_name",
    "description_text_id": "skill_10_desc",
    "unlock_level": 6,
    "mana_cost": 13,
    "cooldown_turns": 1,
    "primary_effect_id": "effect_05",
    "bonus_effect_ids": [],
    "multiplier": 1.25,
    "tags": ["主动", "医师"]
  },
  "skill_11": {
    "name_text_id": "skill_11_name",
    "description_text_id": "skill_11_desc",
    "unlock_level": 11,
    "mana_cost": 0,
    "cooldown_turns": 0,
    "primary_effect_id": "effect_12",
    "bonus_effect_ids": [],
    "multiplier": 1.5,
    "tags": ["被动", "医师"]
  },
  "skill_12": {
    "name_text_id": "skill_12_name",
    "description_text_id": "skill_12_desc",
    "unlock_level": 16,
    "mana_cost": 29,
    "cooldown_turns": 3,
    "primary_effect_id": "effect_06",
    "bonus_effect_ids": ["effect_10"],
    "multiplier": 1.75,
    "tags": ["主动", "医师"]
  },
  "skill_13": {
    "name_text_id": "skill_13_name",
    "description_text_id": "skill_13_desc",
    "unlock_level": 1,
    "mana_cost": 5,
    "cooldown_turns": 0,
    "primary_effect_id": "effect_01",
    "bonus_effect_ids": [],
    "multiplier": 1.0,
    "tags": ["主动", "术士"]
  },
  "skill_14": {
    "name_text_id": "skill_14_name",
    "description_text_id": "skill_14_desc",
    "unlock_level": 6,
    "mana_cost": 13,
    "cooldown_turns": 1,
    "primary_effect_id": "effect_08",
    "bonus_effect_ids": [],
    "multiplier": 1.25,
    "tags": ["主动", "术士"]
  },
  "skill_15": {
    "name_text_id": "skill_15_name",
    "description_text_id": "skill_15_desc",
    "unlock_level": 11,
    "mana_cost": 0,
    "cooldown_turns": 0,
    "primary_effect_id": "effect_12",
    "bonus_effect_ids": [],
    "multiplier": 1.5,
    "tags": ["被动", "术士"]
  },
  "skill_16": {
    "name_text_id": "skill_16_name",
    "description_text_id": "skill_16_desc",
    "unlock_level": 16,
    "mana_cost": 29,
    "cooldown_turns": 3,
    "primary_effect_id": "effect_03",
    "bonus_effect_ids": ["effect_10"],
    "multiplier": 1.75,
    "tags": ["主动", "术士"]
  },
  "skill_17": {
    "name_text_id": "skill_17_name",
    "description_text_id": "skill_17_desc",
    "unlock_level": 1,
    "mana_cost": 5,
    "cooldown_turns": 0,
    "primary_effect_id": "effect_13",
    "bonus_effect_ids": [],
    "multiplier": 1.0,
    "tags": ["主动", "战士"]
  },
  "skill_18": {
    "name_text_id": "skill_18_name",
    "description_text_id": "skill_18_desc",
    "unlock_level": 6,
    "mana_cost": 13,
    "cooldown_turns": 1,
    "primary_effect_id": "effect_23",
    "bonus_effect_ids": [],
    "multiplier": 1.25,
    "tags": ["主动", "战士"]
  },
  "skill_19": {
    "name_text_id": "skill_19_name",
    "description_text_id": "skill_19_desc",
    "unlock_level": 11,
    "mana_cost": 0,
    "cooldown_turns": 0,
    "primary_effect_id": "effect_24",
    "bonus_effect_ids": [],
    "multiplier": 1.5,
    "tags": ["被动", "战士"]
  },
  "skill_20": {
    "name_text_id": "skill_20_name",
    "description_text_id": "skill_20_desc",
    "unlock_level": 16,
    "mana_cost": 29,
    "cooldown_turns": 3,
    "primary_effect_id": "effect_21",
    "bonus_effect_ids": ["effect_22"],
    "multiplier": 1.75,
    "tags": ["主动", "战士"]
  },
  "skill_21": {
    "name_text_id": "skill_21_name",
    "description_text_id": "skill_21_desc",
    "unlock_level": 1,
    "mana_cost": 5,
    "cooldown_turns": 0,
    "primary_effect_id": "effect_13",
    "bonus_effect_ids": [],
    "multiplier": 1.0,
    "tags": ["主动", "游侠"]
  },
  "skill_22": {
    "name_text_id": "skill_22_name",
    "description_text_id": "skill_22_desc",
    "unlock_level": 6,
    "mana_cost": 13,
    "cooldown_turns": 1,
    "primary_effect_id": "effect_20",
    "bonus_effect_ids": [],
    "multiplier": 1.25,
    "tags": ["主动", "游侠"]
  },
  "skill_23": {
    "name_text_id": "skill_23_name",
    "description_text_id": "skill_23_desc",
    "unlock_level": 11,
    "mana_cost": 0,
    "cooldown_turns": 0,
    "primary_effect_id": "effect_24",
    "bonus_effect_ids": [],
    "multiplier": 1.5,
    "tags": ["被动", "游侠"]
  },
  "skill_24": {
    "name_text_id": "skill_24_name",
    "description_text_id": "skill_24_desc",
    "unlock_level": 16,
    "mana_cost": 29,
    "cooldown_turns": 3,
    "primary_effect_id": "effect_15",
    "bonus_effect_ids": ["effect_22"],
    "multiplier": 1.75,
    "tags": ["主动", "游侠"]
  },
  "skill_25": {
    "name_text_id": "skill_25_name",
    "description_text_id": "skill_25_desc",
    "unlock_level": 1,
    "mana_cost": 5,
    "cooldown_turns": 0,
    "primary_effect_id": "effect_16",
    "bonus_effect_ids": [],
    "multiplier": 1.0,
    "tags": ["主动", "医师"]
  },
  "skill_26": {
    "name_text_id": "skill_26_name",
    "description_text_id": "skill_26_desc",
    "unlock_level": 6,
    "mana_cost": 13,
    "cooldown_turns": 1,
    "primary_effect_id": "effect_17",
    "bonus_effect_ids": [],
    "multiplier": 1.25,
    "tags": ["主动", "医师"]
  },
  "skill_27": {
    "name_text_id": "skill_27_name",
    "description_text_id": "skill_27_desc",
    "unlock_level": 11,
    "mana_cost": 0,
    "cooldown_turns": 0,
    "primary_effect_id": "effect_24",
    "bonus_effect_ids": [],
    "multiplier": 1.5,
    "tags": ["被动", "医师"]
  },
  "skill_28": {
    "name_text_id": "skill_28_name",
    "description_text_id": "skill_28_desc",
    "unlock_level": 16,
    "mana_cost": 29,
    "cooldown_turns": 3,
    "primary_effect_id": "effect_18",
    "bonus_effect_ids": ["effect_22"],
    "multiplier": 1.75,
    "tags": ["主动", "医师"]
  },
  "skill_29": {
    "name_text_id": "skill_29_name",
    "description_text_id": "skill_29_desc",
    "unlock_level": 1,
    "mana_cost": 5,
    "cooldown_turns": 0,
    "primary_effect_id": "effect_13",
    "bonus_effect_ids": [],
    "multiplier": 1.0,
    "tags": ["主动", "术士"]
  },
  "skill_30": {
    "name_text_id": "skill_30_name",
    "description_text_id": "skill_30_desc",
    "unlock_level": 6,
    "mana_cost": 13,
    "cooldown_turns": 1,
    "primary_effect_id": "effect_20",
    "bonus_effect_ids": [],
    "multiplier": 1.25,
    "tags": ["主动", "术士"]
  },
  "skill_31": {
    "name_text_id": "skill_31_name",
    "description_text_id": "skill_31_desc",
    "unlock_level": 11,
    "mana_cost": 0,
    "cooldown_turns": 0,
    "primary_effect_id": "effect_24",
    "bonus_effect_ids": [],
    "multiplier": 1.5,
    "tags": ["被动", "术士"]
  },
  "skill_32": {
    "name_text_id": "skill_32_name",
    "description_text_id": "skill_32_desc",
    "unlock_level": 16,
    "mana_cost": 29,
    "cooldown_turns": 3,
    "primary_effect_id": "effect_15",
    "bonus_effect_ids": ["effect_22"],
    "multiplier": 1.75,
    "tags": ["主动", "术士"]
  },
  "skill_33": {
    "name_text_id": "skill_33_name",
    "description_text_id": "skill_33_desc",
    "unlock_level": 1,
    "mana_cost": 5,
    "cooldown_turns": 0,
    "primary_effect_id": "effect_25",
    "bonus_effect_ids": [],
    "multiplier": 1.0,
    "tags": ["主动", "战士"]
  },
  "skill_34": {
    "name_text_id": "skill_34_name",
    "description_text_id": "skill_34_desc",
    "unlock_level": 6,
    "mana_cost": 13,
    "cooldown_turns": 1,
    "primary_effect_id": "effect_35",
    "bonus_effect_ids": [],
    "multiplier": 1.25,
    "tags": ["主动", "战士"]
  },
  "skill_35": {
    "name_text_id": "skill_35_name",
    "description_text_id": "skill_35_desc",
    "unlock_level": 11,
    "mana_cost": 0,
    "cooldown_turns": 0,
    "primary_effect_id": "effect_36",
    "bonus_effect_ids": [],
    "multiplier": 1.5,
    "tags": ["被动", "战士"]
  },
  "skill_36": {
    "name_text_id": "skill_36_name",
    "description_text_id": "skill_36_desc",
    "unlock_level": 16,
    "mana_cost": 29,
    "cooldown_turns": 3,
    "primary_effect_id": "effect_33",
    "bonus_effect_ids": ["effect_34"],
    "multiplier": 1.75,
    "tags": ["主动", "战士"]
  },
  "skill_37": {
    "name_text_id": "skill_37_name",
    "description_text_id": null,
    "unlock_level": 1,
    "mana_cost": 5,
    "cooldown_turns": 0,
    "primary_effect_id": "effect_25",
    "bonus_effect_ids": [],
    "multiplier": 1.0,
    "tags": ["主动", "游侠"]
  },
  "skill_38": {
    "name_text_id": "skill_38_name",
    "description_text_id": null,
    "unlock_level": 6,
    "mana_cost": 13,
    "cooldown_turns": 1,
    "primary_effect_id": "effect_32",
    "bonus_effect_ids": [],
    "multiplier": 1.25,
    "tags": ["主动", "游侠"]
  },
  "skill_39": {
    "name_text_id": "skill_39_name",
    "description_text_id": null,
    "unlock_level": 11,
    "mana_cost": 0,
    "cooldown_turns": 0,
    "primary_effect_id": "effect_36",
    "bonus_effect_ids": [],
    "multiplier": 1.5,
    "tags": ["被动", "游侠"]
  },
  "skill_40": {
    "name_text_id": "skill_40_name",
    "description_text_id": null,
    "unlock_level": 16,
    "mana_cost": 29,
    "cooldown_turns": 3,
    "primary_effect_id": "effect_27",
    "bonus_effect_ids": ["effect_34"],
    "multiplier": 1.75,
    "tags": ["主动", "游侠"]
  },
  "skill_41": {
    "name_text_id": "skill_41_name",
    "description_text_id": null,
    "unlock_level": 1,
    "mana_cost": 5,
    "cooldown_turns": 0,
    "primary_effect_id": "effect_28",
    "bonus_effect_ids": [],
    "multiplier": 1.0,
    "tags": ["主动", "医师"]
  },
  "skill_42": {
    "name_text_id": "skill_42_name",
    "description_text_id": null,
    "unlock_level": 6,
    "mana_cost": 13,
    "cooldown_turns": 1,
    "primary_effect_id": "effect_29",
    "bonus_effect_ids": [],
    "multiplier": 1.25,
    "tags": ["主动", "医师"]
  },
  "skill_43": {
    "name_text_id": "skill_43_name",
    "description_text_id": null,
    "unlock_level": 11,
    "mana_cost": 0,
    "cooldown_turns": 0,
    "primary_effect_id": "effect_36",
    "bonus_effect_ids": [],
    "multiplier": 1.5,
    "tags": ["被动", "医师"]
  },
  "skill_44": {
    "name_text_id": "skill_44_name",
    "description_text_id": null,
    "unlock_level": 16,
    "mana_cost": 29,
    "cooldown_turns": 3,
    "primary_effect_id": "effect_30",
    "bonus_effect_ids": ["effect_34"],
    "multiplier": 1.75,
    "tags": ["主动", "医师"]
  },
  "skill_45": {
    "name_text_id": "skill_45_name",
    "description_text_id": null,
    "unlock_level": 1,
    "mana_cost": 5,
    "cooldown_turns": 0,
    "primary_effect_id": "effect_25",
    "bonus_effect_ids": [],
    "multiplier": 1.0,
    "tags": ["主动", "术士"]
  },
  "skill_46": {
    "name_text_id": "skill_46_name",
    "description_text_id": null,
    "unlock_level": 6,
    "mana_cost": 13,
    "cooldown_turns": 1,
    "primary_effect_id": "effect_32",
    "bonus_effect_ids": [],
    "multiplier": 1.25,
    "tags": ["主动", "术士"]
  },
  "skill_47": {
    "name_text_id": "skill_47_name",
    "description_text_id": null,
    "unlock_level": 11,
    "mana_cost": 0,
    "cooldown_turns": 0,
    "primary_effect_id": "effect_36",
    "bonus_effect_ids": [],
    "multiplier": 1.5,
    "tags": ["被动", "术士"]
  },
  "skill_48": {
    "name_text_id": "skill_48_name",
    "description_text_id": null,
    "unlock_level": 16,
    "mana_cost": 29,
    "cooldown_turns": 3,
    "primary_effect_id": "effect_27",
    "bonus_effect_ids": ["effect_34"],
    "multiplier": 1.75,
    "tags": ["主动", "术士"]
  }
}

_tool: {
  "version": 1,
  "table": {
    "namespace": "content",
    "name": "skill"
  },
  "foreignKeys": {
    "name_text_id": {
      "table": "localization.text"
    },
    "description_text_id": {
      "table": "localization.text"
    },
    "primary_effect_id": {
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
    "title": "技能：角色招式"
  },
  "columns": {
    "name_text_id": {
      "title": "名称文本ID"
    },
    "description_text_id": {
      "title": "描述文本ID"
    },
    "unlock_level": {
      "title": "解锁等级"
    },
    "mana_cost": {
      "title": "法力消耗"
    },
    "cooldown_turns": {
      "title": "冷却回合"
    },
    "primary_effect_id": {
      "title": "主效果ID"
    },
    "bonus_effect_ids": {
      "title": "附加效果列表"
    },
    "multiplier": {
      "title": "倍率"
    },
    "tags": {
      "title": "标签"
    }
  },
  "cellComments": {
    "skill_01": {
      "multiplier": "运行时将倍率与效果强度组合，此处没有跨表公式。"
    },
    "skill_04": {
      "bonus_effect_ids": "终章招式附带增益；列表仍是一个强类型单元格。"
    }
  }
}

_export: {
  "csv": {
    "enabled": true,
    "columns": ["id", "name_text_id", "description_text_id", "unlock_level", "mana_cost", "cooldown_turns", "primary_effect_id", "bonus_effect_ids", "multiplier", "tags"],
    "encoding": {
      "bonus_effect_ids": "json",
      "tags": "json"
    }
  }
}
