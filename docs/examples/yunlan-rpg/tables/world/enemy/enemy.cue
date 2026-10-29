package enemy

import "example.com/datapack/yunlan-rpg/schema:common"

#PrimaryKey: common.#资源标识

#Row: {
    名称文本ID: common.#资源标识
    章节: common.#章节
    等级: int & >=1 & <=60
    生命: int & common.#正数
    攻击: int & common.#非负
    技能ID: [...common.#资源标识]
    掉落组ID: common.#资源标识
    精英: bool
}

rows: {
  "enemy_01": {
    "名称文本ID": "enemy_01_name",
    "章节": "青岚村",
    "等级": 1,
    "生命": 225,
    "攻击": 17,
    "技能ID": ["skill_01"],
    "掉落组ID": "loot_01",
    "精英": false
  },
  "enemy_02": {
    "名称文本ID": "enemy_02_name",
    "章节": "青岚村",
    "等级": 3,
    "生命": 270,
    "攻击": 22,
    "技能ID": ["skill_06"],
    "掉落组ID": "loot_01",
    "精英": false
  },
  "enemy_03": {
    "名称文本ID": "enemy_03_name",
    "章节": "青岚村",
    "等级": 5,
    "生命": 315,
    "攻击": 27,
    "技能ID": ["skill_02"],
    "掉落组ID": "loot_02",
    "精英": false
  },
  "enemy_04": {
    "名称文本ID": "enemy_04_name",
    "章节": "青岚村",
    "等级": 7,
    "生命": 360,
    "攻击": 32,
    "技能ID": ["skill_05"],
    "掉落组ID": "loot_02",
    "精英": false
  },
  "enemy_05": {
    "名称文本ID": "enemy_05_name",
    "章节": "青岚村",
    "等级": 9,
    "生命": 405,
    "攻击": 37,
    "技能ID": ["skill_09"],
    "掉落组ID": "loot_03",
    "精英": false
  },
  "enemy_06": {
    "名称文本ID": "enemy_06_name",
    "章节": "青岚村",
    "等级": 11,
    "生命": 450,
    "攻击": 42,
    "技能ID": ["skill_04"],
    "掉落组ID": "loot_03",
    "精英": false
  },
  "enemy_07": {
    "名称文本ID": "enemy_07_name",
    "章节": "青岚村",
    "等级": 13,
    "生命": 495,
    "攻击": 47,
    "技能ID": ["skill_08"],
    "掉落组ID": "loot_04",
    "精英": false
  },
  "enemy_08": {
    "名称文本ID": "enemy_08_name",
    "章节": "青岚村",
    "等级": 15,
    "生命": 540,
    "攻击": 52,
    "技能ID": ["skill_16"],
    "掉落组ID": "loot_04",
    "精英": true
  },
  "enemy_09": {
    "名称文本ID": "enemy_09_name",
    "章节": "赤砂遗迹",
    "等级": 16,
    "生命": 585,
    "攻击": 57,
    "技能ID": ["skill_17"],
    "掉落组ID": "loot_05",
    "精英": false
  },
  "enemy_10": {
    "名称文本ID": "enemy_10_name",
    "章节": "赤砂遗迹",
    "等级": 18,
    "生命": 630,
    "攻击": 62,
    "技能ID": ["skill_22"],
    "掉落组ID": "loot_05",
    "精英": false
  },
  "enemy_11": {
    "名称文本ID": "enemy_11_name",
    "章节": "赤砂遗迹",
    "等级": 20,
    "生命": 675,
    "攻击": 67,
    "技能ID": ["skill_18"],
    "掉落组ID": "loot_06",
    "精英": false
  },
  "enemy_12": {
    "名称文本ID": "enemy_12_name",
    "章节": "赤砂遗迹",
    "等级": 22,
    "生命": 720,
    "攻击": 72,
    "技能ID": ["skill_21"],
    "掉落组ID": "loot_06",
    "精英": false
  },
  "enemy_13": {
    "名称文本ID": "enemy_13_name",
    "章节": "赤砂遗迹",
    "等级": 24,
    "生命": 765,
    "攻击": 77,
    "技能ID": ["skill_25"],
    "掉落组ID": "loot_07",
    "精英": false
  },
  "enemy_14": {
    "名称文本ID": "enemy_14_name",
    "章节": "赤砂遗迹",
    "等级": 26,
    "生命": 810,
    "攻击": 82,
    "技能ID": ["skill_20"],
    "掉落组ID": "loot_07",
    "精英": false
  },
  "enemy_15": {
    "名称文本ID": "enemy_15_name",
    "章节": "赤砂遗迹",
    "等级": 28,
    "生命": 855,
    "攻击": 87,
    "技能ID": ["skill_24"],
    "掉落组ID": "loot_08",
    "精英": false
  },
  "enemy_16": {
    "名称文本ID": "enemy_16_name",
    "章节": "赤砂遗迹",
    "等级": 30,
    "生命": 900,
    "攻击": 92,
    "技能ID": ["skill_32"],
    "掉落组ID": "loot_08",
    "精英": true
  },
  "enemy_17": {
    "名称文本ID": "enemy_17_name",
    "章节": "霜月城",
    "等级": 31,
    "生命": 945,
    "攻击": 97,
    "技能ID": ["skill_33"],
    "掉落组ID": "loot_09",
    "精英": false
  },
  "enemy_18": {
    "名称文本ID": "enemy_18_name",
    "章节": "霜月城",
    "等级": 33,
    "生命": 990,
    "攻击": 102,
    "技能ID": ["skill_38"],
    "掉落组ID": "loot_09",
    "精英": false
  },
  "enemy_19": {
    "名称文本ID": "enemy_19_name",
    "章节": "霜月城",
    "等级": 35,
    "生命": 1035,
    "攻击": 107,
    "技能ID": ["skill_34"],
    "掉落组ID": "loot_10",
    "精英": false
  },
  "enemy_20": {
    "名称文本ID": "enemy_20_name",
    "章节": "霜月城",
    "等级": 37,
    "生命": 1080,
    "攻击": 112,
    "技能ID": ["skill_37"],
    "掉落组ID": "loot_10",
    "精英": false
  },
  "enemy_21": {
    "名称文本ID": "enemy_21_name",
    "章节": "霜月城",
    "等级": 39,
    "生命": 1125,
    "攻击": 117,
    "技能ID": ["skill_41"],
    "掉落组ID": "loot_11",
    "精英": false
  },
  "enemy_22": {
    "名称文本ID": "enemy_22_name",
    "章节": "霜月城",
    "等级": 41,
    "生命": 1170,
    "攻击": 122,
    "技能ID": ["skill_36"],
    "掉落组ID": "loot_11",
    "精英": false
  },
  "enemy_23": {
    "名称文本ID": "enemy_23_name",
    "章节": "霜月城",
    "等级": 43,
    "生命": 1215,
    "攻击": 127,
    "技能ID": ["skill_40"],
    "掉落组ID": "loot_12",
    "精英": false
  },
  "enemy_24": {
    "名称文本ID": "enemy_24_name",
    "章节": "霜月城",
    "等级": 45,
    "生命": 1260,
    "攻击": 132,
    "技能ID": ["skill_48"],
    "掉落组ID": "loot_12",
    "精英": true
  }
}

_tool: {
  "version": 1,
  "table": {
    "namespace": "world",
    "name": "enemy"
  },
  "foreignKeys": {
    "名称文本ID": {
      "table": "localization.text"
    },
    "技能ID": {
      "table": "content.skill"
    },
    "掉落组ID": {
      "table": "world.loot_group"
    }
  },
  "formulas": {}
}

_editor: {
  "table": {
    "title": "敌人：三章遭遇"
  },
  "cellComments": {
    "enemy_08": {
      "精英": "各章第八个敌人为章末首领，仍复用技能和掉落结构。"
    }
  }
}

_export: {
  "csv": {
    "enabled": true,
    "columns": ["id", "名称文本ID", "章节", "等级", "生命", "攻击", "技能ID", "掉落组ID", "精英"],
    "encoding": {
      "技能ID": "json"
    }
  }
}
