package stage

import "example.com/datapack/yunlan-rpg/schema:common"

#PrimaryKey: common.#resource_id
_local: {
    // 坐标：仅当前表使用。
    #position: {x:number,y:number}
}
_schema: {
  "localDefinitions": {
    "position": {
      "kind": "type",
      "type": "struct"
    }
  }
}

#Row: {
    name_text_id: common.#resource_id
    chapter: common.#chapter
    recommended_level: int & >=1 & <=60
    stamina_cost: int & common.#positive
    enemy_ids: [...common.#resource_id]
    first_clear_reward_id: string | null
    next_stage_id: string | null
    entrance_position: _local.#position
}

rows: {
  "stage_01": {
    "name_text_id": "stage_01_name",
    "chapter": "青岚村",
    "recommended_level": 1,
    "stamina_cost": 5,
    "enemy_ids": ["enemy_01"],
    "first_clear_reward_id": null,
    "next_stage_id": "stage_02",
    "entrance_position": {
      "x": 0,
      "y": 0
    }
  },
  "stage_02": {
    "name_text_id": "stage_02_name",
    "chapter": "青岚村",
    "recommended_level": 3,
    "stamina_cost": 5,
    "enemy_ids": ["enemy_01", "enemy_02"],
    "first_clear_reward_id": "item_quest_01",
    "next_stage_id": "stage_03",
    "entrance_position": {
      "x": 10,
      "y": 5
    }
  },
  "stage_03": {
    "name_text_id": "stage_03_name",
    "chapter": "青岚村",
    "recommended_level": 5,
    "stamina_cost": 5,
    "enemy_ids": ["enemy_02", "enemy_03"],
    "first_clear_reward_id": null,
    "next_stage_id": "stage_04",
    "entrance_position": {
      "x": 20,
      "y": 10
    }
  },
  "stage_04": {
    "name_text_id": "stage_04_name",
    "chapter": "青岚村",
    "recommended_level": 7,
    "stamina_cost": 5,
    "enemy_ids": ["enemy_03", "enemy_04"],
    "first_clear_reward_id": "item_quest_02",
    "next_stage_id": "stage_05",
    "entrance_position": {
      "x": 30,
      "y": 15
    }
  },
  "stage_05": {
    "name_text_id": "stage_05_name",
    "chapter": "青岚村",
    "recommended_level": 9,
    "stamina_cost": 5,
    "enemy_ids": ["enemy_04", "enemy_05"],
    "first_clear_reward_id": null,
    "next_stage_id": "stage_06",
    "entrance_position": {
      "x": 40,
      "y": 20
    }
  },
  "stage_06": {
    "name_text_id": "stage_06_name",
    "chapter": "青岚村",
    "recommended_level": 11,
    "stamina_cost": 5,
    "enemy_ids": ["enemy_05", "enemy_06"],
    "first_clear_reward_id": "item_quest_03",
    "next_stage_id": "stage_07",
    "entrance_position": {
      "x": 50,
      "y": 25
    }
  },
  "stage_07": {
    "name_text_id": "stage_07_name",
    "chapter": "青岚村",
    "recommended_level": 13,
    "stamina_cost": 5,
    "enemy_ids": ["enemy_06", "enemy_07"],
    "first_clear_reward_id": null,
    "next_stage_id": "stage_08",
    "entrance_position": {
      "x": 60,
      "y": 30
    }
  },
  "stage_08": {
    "name_text_id": "stage_08_name",
    "chapter": "青岚村",
    "recommended_level": 15,
    "stamina_cost": 5,
    "enemy_ids": ["enemy_07", "enemy_08"],
    "first_clear_reward_id": "item_quest_04",
    "next_stage_id": "stage_09",
    "entrance_position": {
      "x": 70,
      "y": 35
    }
  },
  "stage_09": {
    "name_text_id": "stage_09_name",
    "chapter": "赤砂遗迹",
    "recommended_level": 16,
    "stamina_cost": 7,
    "enemy_ids": ["enemy_09"],
    "first_clear_reward_id": null,
    "next_stage_id": "stage_10",
    "entrance_position": {
      "x": 100,
      "y": 0
    }
  },
  "stage_10": {
    "name_text_id": "stage_10_name",
    "chapter": "赤砂遗迹",
    "recommended_level": 18,
    "stamina_cost": 7,
    "enemy_ids": ["enemy_09", "enemy_10"],
    "first_clear_reward_id": "item_quest_05",
    "next_stage_id": "stage_11",
    "entrance_position": {
      "x": 110,
      "y": 5
    }
  },
  "stage_11": {
    "name_text_id": "stage_11_name",
    "chapter": "赤砂遗迹",
    "recommended_level": 20,
    "stamina_cost": 7,
    "enemy_ids": ["enemy_10", "enemy_11"],
    "first_clear_reward_id": null,
    "next_stage_id": "stage_12",
    "entrance_position": {
      "x": 120,
      "y": 10
    }
  },
  "stage_12": {
    "name_text_id": "stage_12_name",
    "chapter": "赤砂遗迹",
    "recommended_level": 22,
    "stamina_cost": 7,
    "enemy_ids": ["enemy_11", "enemy_12"],
    "first_clear_reward_id": "item_quest_06",
    "next_stage_id": "stage_13",
    "entrance_position": {
      "x": 130,
      "y": 15
    }
  },
  "stage_13": {
    "name_text_id": "stage_13_name",
    "chapter": "赤砂遗迹",
    "recommended_level": 24,
    "stamina_cost": 7,
    "enemy_ids": ["enemy_12", "enemy_13"],
    "first_clear_reward_id": null,
    "next_stage_id": "stage_14",
    "entrance_position": {
      "x": 140,
      "y": 20
    }
  },
  "stage_14": {
    "name_text_id": "stage_14_name",
    "chapter": "赤砂遗迹",
    "recommended_level": 26,
    "stamina_cost": 7,
    "enemy_ids": ["enemy_13", "enemy_14"],
    "first_clear_reward_id": "item_quest_07",
    "next_stage_id": "stage_15",
    "entrance_position": {
      "x": 150,
      "y": 25
    }
  },
  "stage_15": {
    "name_text_id": "stage_15_name",
    "chapter": "赤砂遗迹",
    "recommended_level": 28,
    "stamina_cost": 7,
    "enemy_ids": ["enemy_14", "enemy_15"],
    "first_clear_reward_id": null,
    "next_stage_id": "stage_16",
    "entrance_position": {
      "x": 160,
      "y": 30
    }
  },
  "stage_16": {
    "name_text_id": "stage_16_name",
    "chapter": "赤砂遗迹",
    "recommended_level": 30,
    "stamina_cost": 7,
    "enemy_ids": ["enemy_15", "enemy_16"],
    "first_clear_reward_id": "item_quest_08",
    "next_stage_id": "stage_17",
    "entrance_position": {
      "x": 170,
      "y": 35
    }
  },
  "stage_17": {
    "name_text_id": "stage_17_name",
    "chapter": "霜月城",
    "recommended_level": 31,
    "stamina_cost": 9,
    "enemy_ids": ["enemy_17"],
    "first_clear_reward_id": null,
    "next_stage_id": "stage_18",
    "entrance_position": {
      "x": 200,
      "y": 0
    }
  },
  "stage_18": {
    "name_text_id": "stage_18_name",
    "chapter": "霜月城",
    "recommended_level": 33,
    "stamina_cost": 9,
    "enemy_ids": ["enemy_17", "enemy_18"],
    "first_clear_reward_id": "item_quest_09",
    "next_stage_id": "stage_19",
    "entrance_position": {
      "x": 210,
      "y": 5
    }
  },
  "stage_19": {
    "name_text_id": "stage_19_name",
    "chapter": "霜月城",
    "recommended_level": 35,
    "stamina_cost": 9,
    "enemy_ids": ["enemy_18", "enemy_19"],
    "first_clear_reward_id": null,
    "next_stage_id": "stage_20",
    "entrance_position": {
      "x": 220,
      "y": 10
    }
  },
  "stage_20": {
    "name_text_id": "stage_20_name",
    "chapter": "霜月城",
    "recommended_level": 37,
    "stamina_cost": 9,
    "enemy_ids": ["enemy_19", "enemy_20"],
    "first_clear_reward_id": "item_quest_10",
    "next_stage_id": "stage_21",
    "entrance_position": {
      "x": 230,
      "y": 15
    }
  },
  "stage_21": {
    "name_text_id": "stage_21_name",
    "chapter": "霜月城",
    "recommended_level": 39,
    "stamina_cost": 9,
    "enemy_ids": ["enemy_20", "enemy_21"],
    "first_clear_reward_id": null,
    "next_stage_id": "stage_22",
    "entrance_position": {
      "x": 240,
      "y": 20
    }
  },
  "stage_22": {
    "name_text_id": "stage_22_name",
    "chapter": "霜月城",
    "recommended_level": 41,
    "stamina_cost": 9,
    "enemy_ids": ["enemy_21", "enemy_22"],
    "first_clear_reward_id": "item_quest_11",
    "next_stage_id": "stage_23",
    "entrance_position": {
      "x": 250,
      "y": 25
    }
  },
  "stage_23": {
    "name_text_id": "stage_23_name",
    "chapter": "霜月城",
    "recommended_level": 43,
    "stamina_cost": 9,
    "enemy_ids": ["enemy_22", "enemy_23"],
    "first_clear_reward_id": null,
    "next_stage_id": "stage_24",
    "entrance_position": {
      "x": 260,
      "y": 30
    }
  },
  "stage_24": {
    "name_text_id": "stage_24_name",
    "chapter": "霜月城",
    "recommended_level": 45,
    "stamina_cost": 9,
    "enemy_ids": ["enemy_23", "enemy_24"],
    "first_clear_reward_id": "item_quest_12",
    "next_stage_id": null,
    "entrance_position": {
      "x": 270,
      "y": 35
    }
  }
}

_tool: {
  "version": 1,
  "table": {
    "namespace": "world",
    "name": "stage"
  },
  "foreignKeys": {
    "name_text_id": {
      "table": "localization.text"
    },
    "enemy_ids": {
      "table": "world.enemy"
    },
    "first_clear_reward_id": {
      "table": "content.item"
    },
    "next_stage_id": {
      "table": "world.stage"
    }
  },
  "formulas": {}
}

_editor: {
  "table": {
    "title": "关卡：章节地图"
  },
  "columns": {
    "name_text_id": {
      "title": "名称文本ID"
    },
    "chapter": {
      "title": "章节"
    },
    "recommended_level": {
      "title": "建议等级"
    },
    "stamina_cost": {
      "title": "体力消耗"
    },
    "enemy_ids": {
      "title": "敌人列表"
    },
    "first_clear_reward_id": {
      "title": "首通奖励ID"
    },
    "next_stage_id": {
      "title": "后续关卡ID"
    },
    "entrance_position": {
      "title": "入口坐标"
    }
  },
  "cellComments": {
    "stage_01": {
      "next_stage_id": "同表可空外键：最后一关明确写null，而不是空字符串。",
      "entrance_position": "当前只支持扁平结构；这里没有结构体数组。"
    }
  }
}

_export: {
  "csv": {
    "enabled": true,
    "columns": ["id", "name_text_id", "chapter", "recommended_level", "stamina_cost", "enemy_ids", "first_clear_reward_id", "next_stage_id", "entrance_position"],
    "encoding": {
      "enemy_ids": "json",
      "entrance_position": "json"
    }
  }
}
