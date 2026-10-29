package stage

import "example.com/datapack/yunlan-rpg/schema:common"

#PrimaryKey: common.#资源标识
_local: {
    #坐标: {横:number,纵:number}
}
_schema: {
  "localDefinitions": {
    "坐标": {
      "kind": "type",
      "type": "struct"
    }
  }
}

#Row: {
    名称文本ID: common.#资源标识
    章节: common.#章节
    建议等级: int & >=1 & <=60
    体力消耗: int & common.#正数
    敌人ID: [...common.#资源标识]
    首通奖励ID: string | null
    后续关卡ID: string | null
    入口坐标: _local.#坐标
}

rows: {
  "stage_01": {
    "名称文本ID": "stage_01_name",
    "章节": "青岚村",
    "建议等级": 1,
    "体力消耗": 5,
    "敌人ID": ["enemy_01"],
    "首通奖励ID": null,
    "后续关卡ID": "stage_02",
    "入口坐标": {
      "横": 0,
      "纵": 0
    }
  },
  "stage_02": {
    "名称文本ID": "stage_02_name",
    "章节": "青岚村",
    "建议等级": 3,
    "体力消耗": 5,
    "敌人ID": ["enemy_01", "enemy_02"],
    "首通奖励ID": "item_quest_01",
    "后续关卡ID": "stage_03",
    "入口坐标": {
      "横": 10,
      "纵": 5
    }
  },
  "stage_03": {
    "名称文本ID": "stage_03_name",
    "章节": "青岚村",
    "建议等级": 5,
    "体力消耗": 5,
    "敌人ID": ["enemy_02", "enemy_03"],
    "首通奖励ID": null,
    "后续关卡ID": "stage_04",
    "入口坐标": {
      "横": 20,
      "纵": 10
    }
  },
  "stage_04": {
    "名称文本ID": "stage_04_name",
    "章节": "青岚村",
    "建议等级": 7,
    "体力消耗": 5,
    "敌人ID": ["enemy_03", "enemy_04"],
    "首通奖励ID": "item_quest_02",
    "后续关卡ID": "stage_05",
    "入口坐标": {
      "横": 30,
      "纵": 15
    }
  },
  "stage_05": {
    "名称文本ID": "stage_05_name",
    "章节": "青岚村",
    "建议等级": 9,
    "体力消耗": 5,
    "敌人ID": ["enemy_04", "enemy_05"],
    "首通奖励ID": null,
    "后续关卡ID": "stage_06",
    "入口坐标": {
      "横": 40,
      "纵": 20
    }
  },
  "stage_06": {
    "名称文本ID": "stage_06_name",
    "章节": "青岚村",
    "建议等级": 11,
    "体力消耗": 5,
    "敌人ID": ["enemy_05", "enemy_06"],
    "首通奖励ID": "item_quest_03",
    "后续关卡ID": "stage_07",
    "入口坐标": {
      "横": 50,
      "纵": 25
    }
  },
  "stage_07": {
    "名称文本ID": "stage_07_name",
    "章节": "青岚村",
    "建议等级": 13,
    "体力消耗": 5,
    "敌人ID": ["enemy_06", "enemy_07"],
    "首通奖励ID": null,
    "后续关卡ID": "stage_08",
    "入口坐标": {
      "横": 60,
      "纵": 30
    }
  },
  "stage_08": {
    "名称文本ID": "stage_08_name",
    "章节": "青岚村",
    "建议等级": 15,
    "体力消耗": 5,
    "敌人ID": ["enemy_07", "enemy_08"],
    "首通奖励ID": "item_quest_04",
    "后续关卡ID": "stage_09",
    "入口坐标": {
      "横": 70,
      "纵": 35
    }
  },
  "stage_09": {
    "名称文本ID": "stage_09_name",
    "章节": "赤砂遗迹",
    "建议等级": 16,
    "体力消耗": 7,
    "敌人ID": ["enemy_09"],
    "首通奖励ID": null,
    "后续关卡ID": "stage_10",
    "入口坐标": {
      "横": 100,
      "纵": 0
    }
  },
  "stage_10": {
    "名称文本ID": "stage_10_name",
    "章节": "赤砂遗迹",
    "建议等级": 18,
    "体力消耗": 7,
    "敌人ID": ["enemy_09", "enemy_10"],
    "首通奖励ID": "item_quest_05",
    "后续关卡ID": "stage_11",
    "入口坐标": {
      "横": 110,
      "纵": 5
    }
  },
  "stage_11": {
    "名称文本ID": "stage_11_name",
    "章节": "赤砂遗迹",
    "建议等级": 20,
    "体力消耗": 7,
    "敌人ID": ["enemy_10", "enemy_11"],
    "首通奖励ID": null,
    "后续关卡ID": "stage_12",
    "入口坐标": {
      "横": 120,
      "纵": 10
    }
  },
  "stage_12": {
    "名称文本ID": "stage_12_name",
    "章节": "赤砂遗迹",
    "建议等级": 22,
    "体力消耗": 7,
    "敌人ID": ["enemy_11", "enemy_12"],
    "首通奖励ID": "item_quest_06",
    "后续关卡ID": "stage_13",
    "入口坐标": {
      "横": 130,
      "纵": 15
    }
  },
  "stage_13": {
    "名称文本ID": "stage_13_name",
    "章节": "赤砂遗迹",
    "建议等级": 24,
    "体力消耗": 7,
    "敌人ID": ["enemy_12", "enemy_13"],
    "首通奖励ID": null,
    "后续关卡ID": "stage_14",
    "入口坐标": {
      "横": 140,
      "纵": 20
    }
  },
  "stage_14": {
    "名称文本ID": "stage_14_name",
    "章节": "赤砂遗迹",
    "建议等级": 26,
    "体力消耗": 7,
    "敌人ID": ["enemy_13", "enemy_14"],
    "首通奖励ID": "item_quest_07",
    "后续关卡ID": "stage_15",
    "入口坐标": {
      "横": 150,
      "纵": 25
    }
  },
  "stage_15": {
    "名称文本ID": "stage_15_name",
    "章节": "赤砂遗迹",
    "建议等级": 28,
    "体力消耗": 7,
    "敌人ID": ["enemy_14", "enemy_15"],
    "首通奖励ID": null,
    "后续关卡ID": "stage_16",
    "入口坐标": {
      "横": 160,
      "纵": 30
    }
  },
  "stage_16": {
    "名称文本ID": "stage_16_name",
    "章节": "赤砂遗迹",
    "建议等级": 30,
    "体力消耗": 7,
    "敌人ID": ["enemy_15", "enemy_16"],
    "首通奖励ID": "item_quest_08",
    "后续关卡ID": "stage_17",
    "入口坐标": {
      "横": 170,
      "纵": 35
    }
  },
  "stage_17": {
    "名称文本ID": "stage_17_name",
    "章节": "霜月城",
    "建议等级": 31,
    "体力消耗": 9,
    "敌人ID": ["enemy_17"],
    "首通奖励ID": null,
    "后续关卡ID": "stage_18",
    "入口坐标": {
      "横": 200,
      "纵": 0
    }
  },
  "stage_18": {
    "名称文本ID": "stage_18_name",
    "章节": "霜月城",
    "建议等级": 33,
    "体力消耗": 9,
    "敌人ID": ["enemy_17", "enemy_18"],
    "首通奖励ID": "item_quest_09",
    "后续关卡ID": "stage_19",
    "入口坐标": {
      "横": 210,
      "纵": 5
    }
  },
  "stage_19": {
    "名称文本ID": "stage_19_name",
    "章节": "霜月城",
    "建议等级": 35,
    "体力消耗": 9,
    "敌人ID": ["enemy_18", "enemy_19"],
    "首通奖励ID": null,
    "后续关卡ID": "stage_20",
    "入口坐标": {
      "横": 220,
      "纵": 10
    }
  },
  "stage_20": {
    "名称文本ID": "stage_20_name",
    "章节": "霜月城",
    "建议等级": 37,
    "体力消耗": 9,
    "敌人ID": ["enemy_19", "enemy_20"],
    "首通奖励ID": "item_quest_10",
    "后续关卡ID": "stage_21",
    "入口坐标": {
      "横": 230,
      "纵": 15
    }
  },
  "stage_21": {
    "名称文本ID": "stage_21_name",
    "章节": "霜月城",
    "建议等级": 39,
    "体力消耗": 9,
    "敌人ID": ["enemy_20", "enemy_21"],
    "首通奖励ID": null,
    "后续关卡ID": "stage_22",
    "入口坐标": {
      "横": 240,
      "纵": 20
    }
  },
  "stage_22": {
    "名称文本ID": "stage_22_name",
    "章节": "霜月城",
    "建议等级": 41,
    "体力消耗": 9,
    "敌人ID": ["enemy_21", "enemy_22"],
    "首通奖励ID": "item_quest_11",
    "后续关卡ID": "stage_23",
    "入口坐标": {
      "横": 250,
      "纵": 25
    }
  },
  "stage_23": {
    "名称文本ID": "stage_23_name",
    "章节": "霜月城",
    "建议等级": 43,
    "体力消耗": 9,
    "敌人ID": ["enemy_22", "enemy_23"],
    "首通奖励ID": null,
    "后续关卡ID": "stage_24",
    "入口坐标": {
      "横": 260,
      "纵": 30
    }
  },
  "stage_24": {
    "名称文本ID": "stage_24_name",
    "章节": "霜月城",
    "建议等级": 45,
    "体力消耗": 9,
    "敌人ID": ["enemy_23", "enemy_24"],
    "首通奖励ID": "item_quest_12",
    "后续关卡ID": null,
    "入口坐标": {
      "横": 270,
      "纵": 35
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
    "名称文本ID": {
      "table": "localization.text"
    },
    "敌人ID": {
      "table": "world.enemy"
    },
    "首通奖励ID": {
      "table": "content.item"
    },
    "后续关卡ID": {
      "table": "world.stage"
    }
  },
  "formulas": {}
}

_editor: {
  "table": {
    "title": "关卡：章节地图"
  },
  "cellComments": {
    "stage_01": {
      "后续关卡ID": "同表可空外键：最后一关明确写null，而不是空字符串。",
      "入口坐标": "当前只支持扁平结构；这里没有结构体数组。"
    }
  }
}

_export: {
  "csv": {
    "enabled": true,
    "columns": ["id", "名称文本ID", "章节", "建议等级", "体力消耗", "敌人ID", "首通奖励ID", "后续关卡ID", "入口坐标"],
    "encoding": {
      "敌人ID": "json",
      "入口坐标": "json"
    }
  }
}
