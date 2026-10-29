package loot_group

import "example.com/datapack/yunlan-rpg/schema:common"

#PrimaryKey: common.#resource_id

#Row: {
    chapter: common.#chapter
    description: common.#non_empty_text
    draw_count: int & >=1 & <=10
}

rows: {
  "loot_01": {
    "chapter": "青岚村",
    "description": "青岚村第1段敌人掉落池；独立抽一次，明细权重合计1。",
    "draw_count": 1
  },
  "loot_02": {
    "chapter": "青岚村",
    "description": "青岚村第2段敌人掉落池；独立抽一次，明细权重合计1。",
    "draw_count": 1
  },
  "loot_03": {
    "chapter": "青岚村",
    "description": "青岚村第3段敌人掉落池；独立抽一次，明细权重合计1。",
    "draw_count": 1
  },
  "loot_04": {
    "chapter": "青岚村",
    "description": "青岚村第4段敌人掉落池；独立抽一次，明细权重合计1。",
    "draw_count": 1
  },
  "loot_05": {
    "chapter": "赤砂遗迹",
    "description": "赤砂遗迹第1段敌人掉落池；独立抽一次，明细权重合计1。",
    "draw_count": 1
  },
  "loot_06": {
    "chapter": "赤砂遗迹",
    "description": "赤砂遗迹第2段敌人掉落池；独立抽一次，明细权重合计1。",
    "draw_count": 1
  },
  "loot_07": {
    "chapter": "赤砂遗迹",
    "description": "赤砂遗迹第3段敌人掉落池；独立抽一次，明细权重合计1。",
    "draw_count": 1
  },
  "loot_08": {
    "chapter": "赤砂遗迹",
    "description": "赤砂遗迹第4段敌人掉落池；独立抽一次，明细权重合计1。",
    "draw_count": 1
  },
  "loot_09": {
    "chapter": "霜月城",
    "description": "霜月城第1段敌人掉落池；独立抽一次，明细权重合计1。",
    "draw_count": 1
  },
  "loot_10": {
    "chapter": "霜月城",
    "description": "霜月城第2段敌人掉落池；独立抽一次，明细权重合计1。",
    "draw_count": 1
  },
  "loot_11": {
    "chapter": "霜月城",
    "description": "霜月城第3段敌人掉落池；独立抽一次，明细权重合计1。",
    "draw_count": 1
  },
  "loot_12": {
    "chapter": "霜月城",
    "description": "霜月城第4段敌人掉落池；独立抽一次，明细权重合计1。",
    "draw_count": 1
  }
}

_tool: {
  "version": 1,
  "table": {
    "namespace": "world",
    "name": "loot_group"
  },
  "foreignKeys": {},
  "formulas": {}
}

_editor: {
  "table": {
    "title": "掉落组：抽取规则"
  },
  "columns": {
    "chapter": {
      "title": "章节"
    },
    "description": {
      "title": "说明"
    },
    "draw_count": {
      "title": "抽取次数"
    }
  },
  "cellComments": {
    "loot_01": {
      "draw_count": "明细权重属于自定义游戏规则，由样例回归额外验证合计1。"
    }
  }
}

_export: {
  "csv": {
    "enabled": true,
    "columns": ["id", "chapter", "description", "draw_count"],
    "encoding": {}
  }
}
