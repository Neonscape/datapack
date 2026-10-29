package hero

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
    简介文本ID: common.#资源标识
    定位: common.#角色定位
    初始等级: int & >=1 & <=60
    基础生命: int & common.#正数
    基础攻击: int & common.#非负
    初始装备ID: string | null
    技能ID: [...common.#资源标识]
    出生坐标: _local.#坐标
    开发备注: string | null
}

rows: {
  "hero_01": {
    "名称文本ID": "hero_01_name",
    "简介文本ID": "hero_01_bio",
    "定位": "战士",
    "初始等级": 1,
    "基础生命": 300,
    "基础攻击": 25,
    "初始装备ID": "item_weapon_01",
    "技能ID": ["skill_01", "skill_02", "skill_03", "skill_04"],
    "出生坐标": {
      "横": 0,
      "纵": 0
    },
    "开发备注": "初始技能数组顺序对应技能栏位置。"
  },
  "hero_02": {
    "名称文本ID": "hero_02_name",
    "简介文本ID": "hero_02_bio",
    "定位": "游侠",
    "初始等级": 1,
    "基础生命": 320,
    "基础攻击": 28,
    "初始装备ID": "item_weapon_01",
    "技能ID": ["skill_05", "skill_06", "skill_07", "skill_08"],
    "出生坐标": {
      "横": 1,
      "纵": 3
    },
    "开发备注": null
  },
  "hero_03": {
    "名称文本ID": "hero_03_name",
    "简介文本ID": "hero_03_bio",
    "定位": "医师",
    "初始等级": 1,
    "基础生命": 340,
    "基础攻击": 31,
    "初始装备ID": "item_weapon_01",
    "技能ID": ["skill_09", "skill_10", "skill_11", "skill_12"],
    "出生坐标": {
      "横": 2,
      "纵": 6
    },
    "开发备注": "初始技能数组顺序对应技能栏位置。"
  },
  "hero_04": {
    "名称文本ID": "hero_04_name",
    "简介文本ID": "hero_04_bio",
    "定位": "术士",
    "初始等级": 1,
    "基础生命": 360,
    "基础攻击": 34,
    "初始装备ID": "item_weapon_01",
    "技能ID": ["skill_13", "skill_14", "skill_15", "skill_16"],
    "出生坐标": {
      "横": 3,
      "纵": 9
    },
    "开发备注": null
  },
  "hero_05": {
    "名称文本ID": "hero_05_name",
    "简介文本ID": "hero_05_bio",
    "定位": "战士",
    "初始等级": 16,
    "基础生命": 400,
    "基础攻击": 37,
    "初始装备ID": "item_weapon_05",
    "技能ID": ["skill_17", "skill_18", "skill_19", "skill_20"],
    "出生坐标": {
      "横": 104,
      "纵": 12
    },
    "开发备注": "初始技能数组顺序对应技能栏位置。"
  },
  "hero_06": {
    "名称文本ID": "hero_06_name",
    "简介文本ID": "hero_06_bio",
    "定位": "游侠",
    "初始等级": 16,
    "基础生命": 420,
    "基础攻击": 40,
    "初始装备ID": "item_weapon_05",
    "技能ID": ["skill_21", "skill_22", "skill_23", "skill_24"],
    "出生坐标": {
      "横": 105,
      "纵": 15
    },
    "开发备注": null
  },
  "hero_07": {
    "名称文本ID": "hero_07_name",
    "简介文本ID": "hero_07_bio",
    "定位": "医师",
    "初始等级": 16,
    "基础生命": 440,
    "基础攻击": 43,
    "初始装备ID": "item_weapon_05",
    "技能ID": ["skill_25", "skill_26", "skill_27", "skill_28"],
    "出生坐标": {
      "横": 106,
      "纵": 18
    },
    "开发备注": "初始技能数组顺序对应技能栏位置。"
  },
  "hero_08": {
    "名称文本ID": "hero_08_name",
    "简介文本ID": "hero_08_bio",
    "定位": "术士",
    "初始等级": 16,
    "基础生命": 460,
    "基础攻击": 46,
    "初始装备ID": "item_weapon_05",
    "技能ID": ["skill_29", "skill_30", "skill_31", "skill_32"],
    "出生坐标": {
      "横": 107,
      "纵": 21
    },
    "开发备注": null
  },
  "hero_09": {
    "名称文本ID": "hero_09_name",
    "简介文本ID": "hero_09_bio",
    "定位": "战士",
    "初始等级": 31,
    "基础生命": 500,
    "基础攻击": 49,
    "初始装备ID": "item_weapon_09",
    "技能ID": ["skill_33", "skill_34", "skill_35", "skill_36"],
    "出生坐标": {
      "横": 208,
      "纵": 24
    },
    "开发备注": "初始技能数组顺序对应技能栏位置。"
  },
  "hero_10": {
    "名称文本ID": "hero_10_name",
    "简介文本ID": "hero_10_bio",
    "定位": "游侠",
    "初始等级": 31,
    "基础生命": 520,
    "基础攻击": 52,
    "初始装备ID": "item_weapon_09",
    "技能ID": ["skill_37", "skill_38", "skill_39", "skill_40"],
    "出生坐标": {
      "横": 209,
      "纵": 27
    },
    "开发备注": null
  },
  "hero_11": {
    "名称文本ID": "hero_11_name",
    "简介文本ID": "hero_11_bio",
    "定位": "医师",
    "初始等级": 31,
    "基础生命": 540,
    "基础攻击": 55,
    "初始装备ID": "item_weapon_09",
    "技能ID": ["skill_41", "skill_42", "skill_43", "skill_44"],
    "出生坐标": {
      "横": 210,
      "纵": 30
    },
    "开发备注": "初始技能数组顺序对应技能栏位置。"
  },
  "hero_12": {
    "名称文本ID": "hero_12_name",
    "简介文本ID": "hero_12_bio",
    "定位": "术士",
    "初始等级": 31,
    "基础生命": 560,
    "基础攻击": 58,
    "初始装备ID": "item_weapon_09",
    "技能ID": ["skill_45", "skill_46", "skill_47", "skill_48"],
    "出生坐标": {
      "横": 211,
      "纵": 33
    },
    "开发备注": null
  }
}

_tool: {
  "version": 1,
  "table": {
    "namespace": "content",
    "name": "hero"
  },
  "foreignKeys": {
    "名称文本ID": {
      "table": "localization.text"
    },
    "简介文本ID": {
      "table": "localization.text"
    },
    "初始装备ID": {
      "table": "content.item"
    },
    "技能ID": {
      "table": "content.skill"
    }
  },
  "formulas": {}
}

_editor: {
  "table": {
    "title": "英雄：队伍配置"
  },
  "cellComments": {
    "hero_01": {
      "技能ID": "从这里沿数组外键读取四个技能，再读取技能的主效果；编辑器不做自动联表。",
      "出生坐标": "扁平结构示例，横纵是结构成员，不是独立列。"
    },
    "hero_09": {
      "初始等级": "第三章招募角色从31级开始，仍受1至60约束。"
    }
  }
}

_export: {
  "csv": {
    "enabled": true,
    "columns": ["id", "名称文本ID", "简介文本ID", "定位", "初始等级", "基础生命", "基础攻击", "初始装备ID", "技能ID", "出生坐标"],
    "encoding": {
      "技能ID": "json",
      "出生坐标": "json"
    }
  }
}
