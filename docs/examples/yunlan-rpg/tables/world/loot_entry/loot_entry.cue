package loot_entry

import "example.com/datapack/yunlan-rpg/schema:common"

#PrimaryKey: common.#资源标识

#Row: {
    掉落组ID: common.#资源标识
    道具ID: common.#资源标识
    权重: common.#概率
    最小数量: int & >=1
    最大数量: int & >=1
}

rows: {
  "drop_01_1": {
    "掉落组ID": "loot_01",
    "道具ID": "item_material_01",
    "权重": 0.4,
    "最小数量": 1,
    "最大数量": 1
  },
  "drop_01_2": {
    "掉落组ID": "loot_01",
    "道具ID": "item_material_02",
    "权重": 0.3,
    "最小数量": 1,
    "最大数量": 2
  },
  "drop_01_3": {
    "掉落组ID": "loot_01",
    "道具ID": "item_potion_01",
    "权重": 0.3,
    "最小数量": 1,
    "最大数量": 3
  },
  "drop_02_1": {
    "掉落组ID": "loot_02",
    "道具ID": "item_material_02",
    "权重": 0.4,
    "最小数量": 1,
    "最大数量": 1
  },
  "drop_02_2": {
    "掉落组ID": "loot_02",
    "道具ID": "item_material_03",
    "权重": 0.3,
    "最小数量": 1,
    "最大数量": 2
  },
  "drop_02_3": {
    "掉落组ID": "loot_02",
    "道具ID": "item_potion_02",
    "权重": 0.3,
    "最小数量": 1,
    "最大数量": 3
  },
  "drop_03_1": {
    "掉落组ID": "loot_03",
    "道具ID": "item_material_03",
    "权重": 0.4,
    "最小数量": 1,
    "最大数量": 1
  },
  "drop_03_2": {
    "掉落组ID": "loot_03",
    "道具ID": "item_material_04",
    "权重": 0.3,
    "最小数量": 1,
    "最大数量": 2
  },
  "drop_03_3": {
    "掉落组ID": "loot_03",
    "道具ID": "item_potion_03",
    "权重": 0.3,
    "最小数量": 1,
    "最大数量": 3
  },
  "drop_04_1": {
    "掉落组ID": "loot_04",
    "道具ID": "item_material_04",
    "权重": 0.4,
    "最小数量": 1,
    "最大数量": 1
  },
  "drop_04_2": {
    "掉落组ID": "loot_04",
    "道具ID": "item_material_01",
    "权重": 0.3,
    "最小数量": 1,
    "最大数量": 2
  },
  "drop_04_3": {
    "掉落组ID": "loot_04",
    "道具ID": "item_potion_04",
    "权重": 0.3,
    "最小数量": 1,
    "最大数量": 3
  },
  "drop_05_1": {
    "掉落组ID": "loot_05",
    "道具ID": "item_material_05",
    "权重": 0.4,
    "最小数量": 1,
    "最大数量": 1
  },
  "drop_05_2": {
    "掉落组ID": "loot_05",
    "道具ID": "item_material_06",
    "权重": 0.3,
    "最小数量": 1,
    "最大数量": 2
  },
  "drop_05_3": {
    "掉落组ID": "loot_05",
    "道具ID": "item_potion_05",
    "权重": 0.3,
    "最小数量": 1,
    "最大数量": 3
  },
  "drop_06_1": {
    "掉落组ID": "loot_06",
    "道具ID": "item_material_06",
    "权重": 0.4,
    "最小数量": 1,
    "最大数量": 1
  },
  "drop_06_2": {
    "掉落组ID": "loot_06",
    "道具ID": "item_material_07",
    "权重": 0.3,
    "最小数量": 1,
    "最大数量": 2
  },
  "drop_06_3": {
    "掉落组ID": "loot_06",
    "道具ID": "item_potion_06",
    "权重": 0.3,
    "最小数量": 1,
    "最大数量": 3
  },
  "drop_07_1": {
    "掉落组ID": "loot_07",
    "道具ID": "item_material_07",
    "权重": 0.4,
    "最小数量": 1,
    "最大数量": 1
  },
  "drop_07_2": {
    "掉落组ID": "loot_07",
    "道具ID": "item_material_08",
    "权重": 0.3,
    "最小数量": 1,
    "最大数量": 2
  },
  "drop_07_3": {
    "掉落组ID": "loot_07",
    "道具ID": "item_potion_07",
    "权重": 0.3,
    "最小数量": 1,
    "最大数量": 3
  },
  "drop_08_1": {
    "掉落组ID": "loot_08",
    "道具ID": "item_material_08",
    "权重": 0.4,
    "最小数量": 1,
    "最大数量": 1
  },
  "drop_08_2": {
    "掉落组ID": "loot_08",
    "道具ID": "item_material_05",
    "权重": 0.3,
    "最小数量": 1,
    "最大数量": 2
  },
  "drop_08_3": {
    "掉落组ID": "loot_08",
    "道具ID": "item_potion_08",
    "权重": 0.3,
    "最小数量": 1,
    "最大数量": 3
  },
  "drop_09_1": {
    "掉落组ID": "loot_09",
    "道具ID": "item_material_09",
    "权重": 0.4,
    "最小数量": 1,
    "最大数量": 1
  },
  "drop_09_2": {
    "掉落组ID": "loot_09",
    "道具ID": "item_material_10",
    "权重": 0.3,
    "最小数量": 1,
    "最大数量": 2
  },
  "drop_09_3": {
    "掉落组ID": "loot_09",
    "道具ID": "item_potion_09",
    "权重": 0.3,
    "最小数量": 1,
    "最大数量": 3
  },
  "drop_10_1": {
    "掉落组ID": "loot_10",
    "道具ID": "item_material_10",
    "权重": 0.4,
    "最小数量": 1,
    "最大数量": 1
  },
  "drop_10_2": {
    "掉落组ID": "loot_10",
    "道具ID": "item_material_11",
    "权重": 0.3,
    "最小数量": 1,
    "最大数量": 2
  },
  "drop_10_3": {
    "掉落组ID": "loot_10",
    "道具ID": "item_potion_10",
    "权重": 0.3,
    "最小数量": 1,
    "最大数量": 3
  },
  "drop_11_1": {
    "掉落组ID": "loot_11",
    "道具ID": "item_material_11",
    "权重": 0.4,
    "最小数量": 1,
    "最大数量": 1
  },
  "drop_11_2": {
    "掉落组ID": "loot_11",
    "道具ID": "item_material_12",
    "权重": 0.3,
    "最小数量": 1,
    "最大数量": 2
  },
  "drop_11_3": {
    "掉落组ID": "loot_11",
    "道具ID": "item_potion_11",
    "权重": 0.3,
    "最小数量": 1,
    "最大数量": 3
  },
  "drop_12_1": {
    "掉落组ID": "loot_12",
    "道具ID": "item_material_12",
    "权重": 0.4,
    "最小数量": 1,
    "最大数量": 1
  },
  "drop_12_2": {
    "掉落组ID": "loot_12",
    "道具ID": "item_material_09",
    "权重": 0.3,
    "最小数量": 1,
    "最大数量": 2
  },
  "drop_12_3": {
    "掉落组ID": "loot_12",
    "道具ID": "item_potion_12",
    "权重": 0.3,
    "最小数量": 1,
    "最大数量": 3
  }
}

_tool: {
  "version": 1,
  "table": {
    "namespace": "world",
    "name": "loot_entry"
  },
  "foreignKeys": {
    "掉落组ID": {
      "table": "world.loot_group"
    },
    "道具ID": {
      "table": "content.item"
    }
  },
  "formulas": {}
}

_editor: {
  "table": {
    "title": "掉落明细：关系表"
  },
  "cellComments": {
    "drop_01_1": {
      "道具ID": "一行表示一种掉落，数量和权重在同一行，避免平行数组错位。"
    }
  }
}

_export: {
  "csv": {
    "enabled": true,
    "columns": ["id", "掉落组ID", "道具ID", "权重", "最小数量", "最大数量"],
    "encoding": {}
  }
}
