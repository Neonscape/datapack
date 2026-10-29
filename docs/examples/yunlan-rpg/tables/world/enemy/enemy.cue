package enemy

import "example.com/datapack/yunlan-rpg/schema:common"

#PrimaryKey: common.#resource_id

#Row: {
    name_text_id: common.#resource_id
    chapter: common.#chapter
    level: int & >=1 & <=60
    hp: int & common.#positive
    attack: int & common.#non_negative
    skill_ids: [...common.#resource_id]
    loot_group_id: common.#resource_id
    elite: bool
}

rows: {
  "enemy_01": {
    "name_text_id": "enemy_01_name",
    "chapter": "青岚村",
    "level": 1,
    "hp": 225,
    "attack": 17,
    "skill_ids": ["skill_01"],
    "loot_group_id": "loot_01",
    "elite": false
  },
  "enemy_02": {
    "name_text_id": "enemy_02_name",
    "chapter": "青岚村",
    "level": 3,
    "hp": 270,
    "attack": 22,
    "skill_ids": ["skill_06"],
    "loot_group_id": "loot_01",
    "elite": false
  },
  "enemy_03": {
    "name_text_id": "enemy_03_name",
    "chapter": "青岚村",
    "level": 5,
    "hp": 315,
    "attack": 27,
    "skill_ids": ["skill_02"],
    "loot_group_id": "loot_02",
    "elite": false
  },
  "enemy_04": {
    "name_text_id": "enemy_04_name",
    "chapter": "青岚村",
    "level": 7,
    "hp": 360,
    "attack": 32,
    "skill_ids": ["skill_05"],
    "loot_group_id": "loot_02",
    "elite": false
  },
  "enemy_05": {
    "name_text_id": "enemy_05_name",
    "chapter": "青岚村",
    "level": 9,
    "hp": 405,
    "attack": 37,
    "skill_ids": ["skill_09"],
    "loot_group_id": "loot_03",
    "elite": false
  },
  "enemy_06": {
    "name_text_id": "enemy_06_name",
    "chapter": "青岚村",
    "level": 11,
    "hp": 450,
    "attack": 42,
    "skill_ids": ["skill_04"],
    "loot_group_id": "loot_03",
    "elite": false
  },
  "enemy_07": {
    "name_text_id": "enemy_07_name",
    "chapter": "青岚村",
    "level": 13,
    "hp": 495,
    "attack": 47,
    "skill_ids": ["skill_08"],
    "loot_group_id": "loot_04",
    "elite": false
  },
  "enemy_08": {
    "name_text_id": "enemy_08_name",
    "chapter": "青岚村",
    "level": 15,
    "hp": 540,
    "attack": 52,
    "skill_ids": ["skill_16"],
    "loot_group_id": "loot_04",
    "elite": true
  },
  "enemy_09": {
    "name_text_id": "enemy_09_name",
    "chapter": "赤砂遗迹",
    "level": 16,
    "hp": 585,
    "attack": 57,
    "skill_ids": ["skill_17"],
    "loot_group_id": "loot_05",
    "elite": false
  },
  "enemy_10": {
    "name_text_id": "enemy_10_name",
    "chapter": "赤砂遗迹",
    "level": 18,
    "hp": 630,
    "attack": 62,
    "skill_ids": ["skill_22"],
    "loot_group_id": "loot_05",
    "elite": false
  },
  "enemy_11": {
    "name_text_id": "enemy_11_name",
    "chapter": "赤砂遗迹",
    "level": 20,
    "hp": 675,
    "attack": 67,
    "skill_ids": ["skill_18"],
    "loot_group_id": "loot_06",
    "elite": false
  },
  "enemy_12": {
    "name_text_id": "enemy_12_name",
    "chapter": "赤砂遗迹",
    "level": 22,
    "hp": 720,
    "attack": 72,
    "skill_ids": ["skill_21"],
    "loot_group_id": "loot_06",
    "elite": false
  },
  "enemy_13": {
    "name_text_id": "enemy_13_name",
    "chapter": "赤砂遗迹",
    "level": 24,
    "hp": 765,
    "attack": 77,
    "skill_ids": ["skill_25"],
    "loot_group_id": "loot_07",
    "elite": false
  },
  "enemy_14": {
    "name_text_id": "enemy_14_name",
    "chapter": "赤砂遗迹",
    "level": 26,
    "hp": 810,
    "attack": 82,
    "skill_ids": ["skill_20"],
    "loot_group_id": "loot_07",
    "elite": false
  },
  "enemy_15": {
    "name_text_id": "enemy_15_name",
    "chapter": "赤砂遗迹",
    "level": 28,
    "hp": 855,
    "attack": 87,
    "skill_ids": ["skill_24"],
    "loot_group_id": "loot_08",
    "elite": false
  },
  "enemy_16": {
    "name_text_id": "enemy_16_name",
    "chapter": "赤砂遗迹",
    "level": 30,
    "hp": 900,
    "attack": 92,
    "skill_ids": ["skill_32"],
    "loot_group_id": "loot_08",
    "elite": true
  },
  "enemy_17": {
    "name_text_id": "enemy_17_name",
    "chapter": "霜月城",
    "level": 31,
    "hp": 945,
    "attack": 97,
    "skill_ids": ["skill_33"],
    "loot_group_id": "loot_09",
    "elite": false
  },
  "enemy_18": {
    "name_text_id": "enemy_18_name",
    "chapter": "霜月城",
    "level": 33,
    "hp": 990,
    "attack": 102,
    "skill_ids": ["skill_38"],
    "loot_group_id": "loot_09",
    "elite": false
  },
  "enemy_19": {
    "name_text_id": "enemy_19_name",
    "chapter": "霜月城",
    "level": 35,
    "hp": 1035,
    "attack": 107,
    "skill_ids": ["skill_34"],
    "loot_group_id": "loot_10",
    "elite": false
  },
  "enemy_20": {
    "name_text_id": "enemy_20_name",
    "chapter": "霜月城",
    "level": 37,
    "hp": 1080,
    "attack": 112,
    "skill_ids": ["skill_37"],
    "loot_group_id": "loot_10",
    "elite": false
  },
  "enemy_21": {
    "name_text_id": "enemy_21_name",
    "chapter": "霜月城",
    "level": 39,
    "hp": 1125,
    "attack": 117,
    "skill_ids": ["skill_41"],
    "loot_group_id": "loot_11",
    "elite": false
  },
  "enemy_22": {
    "name_text_id": "enemy_22_name",
    "chapter": "霜月城",
    "level": 41,
    "hp": 1170,
    "attack": 122,
    "skill_ids": ["skill_36"],
    "loot_group_id": "loot_11",
    "elite": false
  },
  "enemy_23": {
    "name_text_id": "enemy_23_name",
    "chapter": "霜月城",
    "level": 43,
    "hp": 1215,
    "attack": 127,
    "skill_ids": ["skill_40"],
    "loot_group_id": "loot_12",
    "elite": false
  },
  "enemy_24": {
    "name_text_id": "enemy_24_name",
    "chapter": "霜月城",
    "level": 45,
    "hp": 1260,
    "attack": 132,
    "skill_ids": ["skill_48"],
    "loot_group_id": "loot_12",
    "elite": true
  }
}

_tool: {
  "version": 1,
  "table": {
    "namespace": "world",
    "name": "enemy"
  },
  "foreignKeys": {
    "name_text_id": {
      "table": "localization.text"
    },
    "skill_ids": {
      "table": "content.skill"
    },
    "loot_group_id": {
      "table": "world.loot_group"
    }
  },
  "formulas": {}
}

_editor: {
  "table": {
    "title": "敌人：三章遭遇"
  },
  "columns": {
    "name_text_id": {
      "title": "名称文本ID"
    },
    "chapter": {
      "title": "章节"
    },
    "level": {
      "title": "等级"
    },
    "hp": {
      "title": "生命"
    },
    "attack": {
      "title": "攻击"
    },
    "skill_ids": {
      "title": "技能列表"
    },
    "loot_group_id": {
      "title": "掉落组ID"
    },
    "elite": {
      "title": "精英"
    }
  },
  "cellComments": {
    "enemy_08": {
      "elite": "各章第八个敌人为章末首领，仍复用技能和掉落结构。"
    }
  }
}

_export: {
  "csv": {
    "enabled": true,
    "columns": ["id", "name_text_id", "chapter", "level", "hp", "attack", "skill_ids", "loot_group_id", "elite"],
    "encoding": {
      "skill_ids": "json"
    }
  }
}
