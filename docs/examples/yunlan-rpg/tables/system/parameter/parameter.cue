package parameter

import "example.com/datapack/yunlan-rpg/schema:common"

#PrimaryKey: common.#资源标识

#Row: {
    数值: number & common.#非负
    单位: "次" | "级" | "比例" | "格"
    说明: common.#必填文本
}

rows: {
  "party_size": {
    "数值": 4,
    "单位": "格",
    "说明": "队伍人数上限。"
  },
  "level_cap": {
    "数值": 60,
    "单位": "级",
    "说明": "与成长表最后一级一致。"
  },
  "daily_reset": {
    "数值": 1,
    "单位": "次",
    "说明": "每日商店重置次数。"
  },
  "reward_scale": {
    "数值": 1,
    "单位": "比例",
    "说明": "设计参考倍率，不被其他表公式读取。"
  },
  "workshop_cap": {
    "数值": 12,
    "单位": "级",
    "说明": "工坊等级上限。"
  },
  "loot_draws": {
    "数值": 1,
    "单位": "次",
    "说明": "各掉落组默认抽取一次。"
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
  "cellComments": {
    "level_cap": {
      "数值": "即使项目设置排除此表的CSV，非法数值仍会阻止整个工程导出。"
    }
  }
}

_export: {
  "csv": {
    "enabled": true,
    "columns": ["id", "数值", "单位", "说明"],
    "encoding": {}
  }
}
