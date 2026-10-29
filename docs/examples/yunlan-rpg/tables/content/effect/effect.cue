package effect

import "example.com/datapack/yunlan-rpg/schema:common"

#PrimaryKey: common.#resource_id

#Row: {
    name_text_id: common.#resource_id
    category: common.#effect_category
    magnitude: number & common.#non_negative
    duration_turns: int & common.#non_negative
    proc_chance: common.#probability
    designer_notes: common.#non_empty_text
}

rows: {
  "effect_01": {
    "name_text_id": "effect_01_name",
    "category": "伤害",
    "magnitude": 10,
    "duration_turns": 1,
    "proc_chance": 1,
    "designer_notes": "青岚村流派的伤害效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_02": {
    "name_text_id": "effect_02_name",
    "category": "伤害",
    "magnitude": 16,
    "duration_turns": 2,
    "proc_chance": 1,
    "designer_notes": "青岚村流派的伤害效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_03": {
    "name_text_id": "effect_03_name",
    "category": "伤害",
    "magnitude": 22,
    "duration_turns": 3,
    "proc_chance": 1,
    "designer_notes": "青岚村流派的伤害效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_04": {
    "name_text_id": "effect_04_name",
    "category": "治疗",
    "magnitude": 10,
    "duration_turns": 1,
    "proc_chance": 1,
    "designer_notes": "青岚村流派的治疗效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_05": {
    "name_text_id": "effect_05_name",
    "category": "治疗",
    "magnitude": 16,
    "duration_turns": 2,
    "proc_chance": 1,
    "designer_notes": "青岚村流派的治疗效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_06": {
    "name_text_id": "effect_06_name",
    "category": "治疗",
    "magnitude": 22,
    "duration_turns": 3,
    "proc_chance": 1,
    "designer_notes": "青岚村流派的治疗效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_07": {
    "name_text_id": "effect_07_name",
    "category": "控制",
    "magnitude": 10,
    "duration_turns": 1,
    "proc_chance": 0.6,
    "designer_notes": "青岚村流派的控制效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_08": {
    "name_text_id": "effect_08_name",
    "category": "控制",
    "magnitude": 16,
    "duration_turns": 2,
    "proc_chance": 0.7,
    "designer_notes": "青岚村流派的控制效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_09": {
    "name_text_id": "effect_09_name",
    "category": "控制",
    "magnitude": 22,
    "duration_turns": 3,
    "proc_chance": 0.8,
    "designer_notes": "青岚村流派的控制效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_10": {
    "name_text_id": "effect_10_name",
    "category": "增益",
    "magnitude": 10,
    "duration_turns": 1,
    "proc_chance": 1,
    "designer_notes": "青岚村流派的增益效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_11": {
    "name_text_id": "effect_11_name",
    "category": "增益",
    "magnitude": 16,
    "duration_turns": 2,
    "proc_chance": 1,
    "designer_notes": "青岚村流派的增益效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_12": {
    "name_text_id": "effect_12_name",
    "category": "增益",
    "magnitude": 22,
    "duration_turns": 3,
    "proc_chance": 1,
    "designer_notes": "青岚村流派的增益效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_13": {
    "name_text_id": "effect_13_name",
    "category": "伤害",
    "magnitude": 18,
    "duration_turns": 1,
    "proc_chance": 1,
    "designer_notes": "赤砂遗迹流派的伤害效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_14": {
    "name_text_id": "effect_14_name",
    "category": "伤害",
    "magnitude": 24,
    "duration_turns": 2,
    "proc_chance": 1,
    "designer_notes": "赤砂遗迹流派的伤害效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_15": {
    "name_text_id": "effect_15_name",
    "category": "伤害",
    "magnitude": 30,
    "duration_turns": 3,
    "proc_chance": 1,
    "designer_notes": "赤砂遗迹流派的伤害效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_16": {
    "name_text_id": "effect_16_name",
    "category": "治疗",
    "magnitude": 18,
    "duration_turns": 1,
    "proc_chance": 1,
    "designer_notes": "赤砂遗迹流派的治疗效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_17": {
    "name_text_id": "effect_17_name",
    "category": "治疗",
    "magnitude": 24,
    "duration_turns": 2,
    "proc_chance": 1,
    "designer_notes": "赤砂遗迹流派的治疗效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_18": {
    "name_text_id": "effect_18_name",
    "category": "治疗",
    "magnitude": 30,
    "duration_turns": 3,
    "proc_chance": 1,
    "designer_notes": "赤砂遗迹流派的治疗效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_19": {
    "name_text_id": "effect_19_name",
    "category": "控制",
    "magnitude": 18,
    "duration_turns": 1,
    "proc_chance": 0.6,
    "designer_notes": "赤砂遗迹流派的控制效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_20": {
    "name_text_id": "effect_20_name",
    "category": "控制",
    "magnitude": 24,
    "duration_turns": 2,
    "proc_chance": 0.7,
    "designer_notes": "赤砂遗迹流派的控制效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_21": {
    "name_text_id": "effect_21_name",
    "category": "控制",
    "magnitude": 30,
    "duration_turns": 3,
    "proc_chance": 0.8,
    "designer_notes": "赤砂遗迹流派的控制效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_22": {
    "name_text_id": "effect_22_name",
    "category": "增益",
    "magnitude": 18,
    "duration_turns": 1,
    "proc_chance": 1,
    "designer_notes": "赤砂遗迹流派的增益效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_23": {
    "name_text_id": "effect_23_name",
    "category": "增益",
    "magnitude": 24,
    "duration_turns": 2,
    "proc_chance": 1,
    "designer_notes": "赤砂遗迹流派的增益效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_24": {
    "name_text_id": "effect_24_name",
    "category": "增益",
    "magnitude": 30,
    "duration_turns": 3,
    "proc_chance": 1,
    "designer_notes": "赤砂遗迹流派的增益效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_25": {
    "name_text_id": "effect_25_name",
    "category": "伤害",
    "magnitude": 26,
    "duration_turns": 1,
    "proc_chance": 1,
    "designer_notes": "霜月城流派的伤害效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_26": {
    "name_text_id": "effect_26_name",
    "category": "伤害",
    "magnitude": 32,
    "duration_turns": 2,
    "proc_chance": 1,
    "designer_notes": "霜月城流派的伤害效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_27": {
    "name_text_id": "effect_27_name",
    "category": "伤害",
    "magnitude": 38,
    "duration_turns": 3,
    "proc_chance": 1,
    "designer_notes": "霜月城流派的伤害效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_28": {
    "name_text_id": "effect_28_name",
    "category": "治疗",
    "magnitude": 26,
    "duration_turns": 1,
    "proc_chance": 1,
    "designer_notes": "霜月城流派的治疗效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_29": {
    "name_text_id": "effect_29_name",
    "category": "治疗",
    "magnitude": 32,
    "duration_turns": 2,
    "proc_chance": 1,
    "designer_notes": "霜月城流派的治疗效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_30": {
    "name_text_id": "effect_30_name",
    "category": "治疗",
    "magnitude": 38,
    "duration_turns": 3,
    "proc_chance": 1,
    "designer_notes": "霜月城流派的治疗效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_31": {
    "name_text_id": "effect_31_name",
    "category": "控制",
    "magnitude": 26,
    "duration_turns": 1,
    "proc_chance": 0.6,
    "designer_notes": "霜月城流派的控制效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_32": {
    "name_text_id": "effect_32_name",
    "category": "控制",
    "magnitude": 32,
    "duration_turns": 2,
    "proc_chance": 0.7,
    "designer_notes": "霜月城流派的控制效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_33": {
    "name_text_id": "effect_33_name",
    "category": "控制",
    "magnitude": 38,
    "duration_turns": 3,
    "proc_chance": 0.8,
    "designer_notes": "霜月城流派的控制效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_34": {
    "name_text_id": "effect_34_name",
    "category": "增益",
    "magnitude": 26,
    "duration_turns": 1,
    "proc_chance": 1,
    "designer_notes": "霜月城流派的增益效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_35": {
    "name_text_id": "effect_35_name",
    "category": "增益",
    "magnitude": 32,
    "duration_turns": 2,
    "proc_chance": 1,
    "designer_notes": "霜月城流派的增益效果；强度由技能倍率使用，不在表间自动计算。"
  },
  "effect_36": {
    "name_text_id": "effect_36_name",
    "category": "增益",
    "magnitude": 38,
    "duration_turns": 3,
    "proc_chance": 1,
    "designer_notes": "霜月城流派的增益效果；强度由技能倍率使用，不在表间自动计算。"
  }
}

_tool: {
  "version": 1,
  "table": {
    "namespace": "content",
    "name": "effect"
  },
  "foreignKeys": {
    "name_text_id": {
      "table": "localization.text"
    }
  },
  "formulas": {}
}

_editor: {
  "table": {
    "title": "效果：战斗机制"
  },
  "columns": {
    "name_text_id": {
      "title": "名称文本ID"
    },
    "category": {
      "title": "类别"
    },
    "magnitude": {
      "title": "强度"
    },
    "duration_turns": {
      "title": "持续回合"
    },
    "proc_chance": {
      "title": "触发概率"
    },
    "designer_notes": {
      "title": "策划备注"
    }
  },
  "cellComments": {
    "effect_01": {
      "magnitude": "这是效果的基础量；技能倍率由游戏运行时组合，不是跨表公式。"
    },
    "effect_04": {
      "category": "治疗与伤害用不同分支，强度仍使用同一非负列。"
    },
    "effect_07": {
      "proc_chance": "0.6代表60%，取值范围由公共概率Schema约束。"
    }
  }
}

_export: {
  "csv": {
    "enabled": true,
    "columns": ["id", "name_text_id", "category", "magnitude", "duration_turns", "proc_chance"],
    "encoding": {}
  }
}
