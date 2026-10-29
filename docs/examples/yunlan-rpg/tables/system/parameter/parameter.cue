package parameter

import "example.com/datapack/yunlan-rpg/schema:common"

#PrimaryKey: common.#resource_id

#Row: {
    value: number & common.#non_negative
    unit: "次" | "级" | "比例" | "格"
    description: common.#non_empty_text
}

rows: {
  "party_size": {
    "value": 4,
    "unit": "格",
    "description": "队伍人数上限。"
  },
  "level_cap": {
    "value": 60,
    "unit": "级",
    "description": "与成长表最后一级一致。"
  },
  "daily_reset": {
    "value": 1,
    "unit": "次",
    "description": "每日商店重置次数。"
  },
  "reward_scale": {
    "value": 1,
    "unit": "比例",
    "description": "设计参考倍率，不被其他表公式读取。"
  },
  "workshop_cap": {
    "value": 12,
    "unit": "级",
    "description": "工坊等级上限。"
  },
  "loot_draws": {
    "value": 1,
    "unit": "次",
    "description": "各掉落组默认抽取一次。"
  }
}

_tool: {
  "version": 1,
  "table": {
    "namespace": "system",
    "name": "parameter"
  },
  "foreignKeys": {},
  "formulas": {}
}

_editor: {
  "table": {
    "title": "系统参数：参与校验但不导出"
  },
  "columns": {
    "value": {
      "title": "数值"
    },
    "unit": {
      "title": "单位"
    },
    "description": {
      "title": "说明"
    }
  },
  "cellComments": {
    "level_cap": {
      "value": "即使项目设置排除此表的CSV，非法数值仍会阻止整个工程导出。"
    }
  }
}

_export: {
  "csv": {
    "enabled": true,
    "columns": ["id", "value", "unit", "description"],
    "encoding": {}
  }
}
