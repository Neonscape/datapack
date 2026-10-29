package hero

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
    biography_text_id: common.#resource_id
    role: common.#hero_role
    starting_level: int & >=1 & <=60
    base_hp: int & common.#positive
    base_attack: int & common.#non_negative
    starting_equipment_id: string | null
    skill_ids: [...common.#resource_id]
    spawn_position: _local.#position
    developer_notes: string | null
}

rows: {
  "hero_01": {
    "name_text_id": "hero_01_name",
    "biography_text_id": "hero_01_bio",
    "role": "战士",
    "starting_level": 1,
    "base_hp": 300,
    "base_attack": 25,
    "starting_equipment_id": "item_weapon_01",
    "skill_ids": ["skill_01", "skill_02", "skill_03", "skill_04"],
    "spawn_position": {
      "x": 0,
      "y": 0
    },
    "developer_notes": "初始技能数组顺序对应技能栏位置。"
  },
  "hero_02": {
    "name_text_id": "hero_02_name",
    "biography_text_id": "hero_02_bio",
    "role": "游侠",
    "starting_level": 1,
    "base_hp": 320,
    "base_attack": 28,
    "starting_equipment_id": "item_weapon_01",
    "skill_ids": ["skill_05", "skill_06", "skill_07", "skill_08"],
    "spawn_position": {
      "x": 1,
      "y": 3
    },
    "developer_notes": null
  },
  "hero_03": {
    "name_text_id": "hero_03_name",
    "biography_text_id": "hero_03_bio",
    "role": "医师",
    "starting_level": 1,
    "base_hp": 340,
    "base_attack": 31,
    "starting_equipment_id": "item_weapon_01",
    "skill_ids": ["skill_09", "skill_10", "skill_11", "skill_12"],
    "spawn_position": {
      "x": 2,
      "y": 6
    },
    "developer_notes": "初始技能数组顺序对应技能栏位置。"
  },
  "hero_04": {
    "name_text_id": "hero_04_name",
    "biography_text_id": "hero_04_bio",
    "role": "术士",
    "starting_level": 1,
    "base_hp": 360,
    "base_attack": 34,
    "starting_equipment_id": "item_weapon_01",
    "skill_ids": ["skill_13", "skill_14", "skill_15", "skill_16"],
    "spawn_position": {
      "x": 3,
      "y": 9
    },
    "developer_notes": null
  },
  "hero_05": {
    "name_text_id": "hero_05_name",
    "biography_text_id": "hero_05_bio",
    "role": "战士",
    "starting_level": 16,
    "base_hp": 400,
    "base_attack": 37,
    "starting_equipment_id": "item_weapon_05",
    "skill_ids": ["skill_17", "skill_18", "skill_19", "skill_20"],
    "spawn_position": {
      "x": 104,
      "y": 12
    },
    "developer_notes": "初始技能数组顺序对应技能栏位置。"
  },
  "hero_06": {
    "name_text_id": "hero_06_name",
    "biography_text_id": "hero_06_bio",
    "role": "游侠",
    "starting_level": 16,
    "base_hp": 420,
    "base_attack": 40,
    "starting_equipment_id": "item_weapon_05",
    "skill_ids": ["skill_21", "skill_22", "skill_23", "skill_24"],
    "spawn_position": {
      "x": 105,
      "y": 15
    },
    "developer_notes": null
  },
  "hero_07": {
    "name_text_id": "hero_07_name",
    "biography_text_id": "hero_07_bio",
    "role": "医师",
    "starting_level": 16,
    "base_hp": 440,
    "base_attack": 43,
    "starting_equipment_id": "item_weapon_05",
    "skill_ids": ["skill_25", "skill_26", "skill_27", "skill_28"],
    "spawn_position": {
      "x": 106,
      "y": 18
    },
    "developer_notes": "初始技能数组顺序对应技能栏位置。"
  },
  "hero_08": {
    "name_text_id": "hero_08_name",
    "biography_text_id": "hero_08_bio",
    "role": "术士",
    "starting_level": 16,
    "base_hp": 460,
    "base_attack": 46,
    "starting_equipment_id": "item_weapon_05",
    "skill_ids": ["skill_29", "skill_30", "skill_31", "skill_32"],
    "spawn_position": {
      "x": 107,
      "y": 21
    },
    "developer_notes": null
  },
  "hero_09": {
    "name_text_id": "hero_09_name",
    "biography_text_id": "hero_09_bio",
    "role": "战士",
    "starting_level": 31,
    "base_hp": 500,
    "base_attack": 49,
    "starting_equipment_id": "item_weapon_09",
    "skill_ids": ["skill_33", "skill_34", "skill_35", "skill_36"],
    "spawn_position": {
      "x": 208,
      "y": 24
    },
    "developer_notes": "初始技能数组顺序对应技能栏位置。"
  },
  "hero_10": {
    "name_text_id": "hero_10_name",
    "biography_text_id": "hero_10_bio",
    "role": "游侠",
    "starting_level": 31,
    "base_hp": 520,
    "base_attack": 52,
    "starting_equipment_id": "item_weapon_09",
    "skill_ids": ["skill_37", "skill_38", "skill_39", "skill_40"],
    "spawn_position": {
      "x": 209,
      "y": 27
    },
    "developer_notes": null
  },
  "hero_11": {
    "name_text_id": "hero_11_name",
    "biography_text_id": "hero_11_bio",
    "role": "医师",
    "starting_level": 31,
    "base_hp": 540,
    "base_attack": 55,
    "starting_equipment_id": "item_weapon_09",
    "skill_ids": ["skill_41", "skill_42", "skill_43", "skill_44"],
    "spawn_position": {
      "x": 210,
      "y": 30
    },
    "developer_notes": "初始技能数组顺序对应技能栏位置。"
  },
  "hero_12": {
    "name_text_id": "hero_12_name",
    "biography_text_id": "hero_12_bio",
    "role": "术士",
    "starting_level": 31,
    "base_hp": 560,
    "base_attack": 58,
    "starting_equipment_id": "item_weapon_09",
    "skill_ids": ["skill_45", "skill_46", "skill_47", "skill_48"],
    "spawn_position": {
      "x": 211,
      "y": 33
    },
    "developer_notes": null
  }
}

_tool: {
  "version": 1,
  "table": {
    "namespace": "content",
    "name": "hero"
  },
  "foreignKeys": {
    "name_text_id": {
      "table": "localization.text"
    },
    "biography_text_id": {
      "table": "localization.text"
    },
    "starting_equipment_id": {
      "table": "content.item"
    },
    "skill_ids": {
      "table": "content.skill"
    }
  },
  "formulas": {}
}

_editor: {
  "table": {
    "title": "英雄：队伍配置"
  },
  "columns": {
    "name_text_id": {
      "title": "名称文本ID"
    },
    "biography_text_id": {
      "title": "简介文本ID"
    },
    "role": {
      "title": "定位"
    },
    "starting_level": {
      "title": "初始等级"
    },
    "base_hp": {
      "title": "基础生命"
    },
    "base_attack": {
      "title": "基础攻击"
    },
    "starting_equipment_id": {
      "title": "初始装备ID"
    },
    "skill_ids": {
      "title": "技能列表"
    },
    "spawn_position": {
      "title": "出生坐标"
    },
    "developer_notes": {
      "title": "开发备注"
    }
  },
  "cellComments": {
    "hero_01": {
      "skill_ids": "从这里沿数组外键读取四个技能，再读取技能的主效果；编辑器不做自动联表。",
      "spawn_position": "扁平结构示例，横纵是结构成员，不是独立列。"
    },
    "hero_09": {
      "starting_level": "第三章招募角色从31级开始，仍受1至60约束。"
    }
  }
}

_export: {
  "csv": {
    "enabled": true,
    "columns": ["id", "name_text_id", "biography_text_id", "role", "starting_level", "base_hp", "base_attack", "starting_equipment_id", "skill_ids", "spawn_position"],
    "encoding": {
      "skill_ids": "json",
      "spawn_position": "json"
    }
  }
}
