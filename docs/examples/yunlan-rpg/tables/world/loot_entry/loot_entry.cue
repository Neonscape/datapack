package loot_entry

import "example.com/datapack/yunlan-rpg/schema:common"

#PrimaryKey: common.#resource_id

#Row: {
    loot_group_id: common.#resource_id
    item_id: common.#resource_id
    weight: common.#probability
    min_quantity: int & >=1
    max_quantity: int & >=1
}

rows: {
  "drop_01_1": {
    "loot_group_id": "loot_01",
    "item_id": "item_material_01",
    "weight": 0.4,
    "min_quantity": 1,
    "max_quantity": 1
  },
  "drop_01_2": {
    "loot_group_id": "loot_01",
    "item_id": "item_material_02",
    "weight": 0.3,
    "min_quantity": 1,
    "max_quantity": 2
  },
  "drop_01_3": {
    "loot_group_id": "loot_01",
    "item_id": "item_potion_01",
    "weight": 0.3,
    "min_quantity": 1,
    "max_quantity": 3
  },
  "drop_02_1": {
    "loot_group_id": "loot_02",
    "item_id": "item_material_02",
    "weight": 0.4,
    "min_quantity": 1,
    "max_quantity": 1
  },
  "drop_02_2": {
    "loot_group_id": "loot_02",
    "item_id": "item_material_03",
    "weight": 0.3,
    "min_quantity": 1,
    "max_quantity": 2
  },
  "drop_02_3": {
    "loot_group_id": "loot_02",
    "item_id": "item_potion_02",
    "weight": 0.3,
    "min_quantity": 1,
    "max_quantity": 3
  },
  "drop_03_1": {
    "loot_group_id": "loot_03",
    "item_id": "item_material_03",
    "weight": 0.4,
    "min_quantity": 1,
    "max_quantity": 1
  },
  "drop_03_2": {
    "loot_group_id": "loot_03",
    "item_id": "item_material_04",
    "weight": 0.3,
    "min_quantity": 1,
    "max_quantity": 2
  },
  "drop_03_3": {
    "loot_group_id": "loot_03",
    "item_id": "item_potion_03",
    "weight": 0.3,
    "min_quantity": 1,
    "max_quantity": 3
  },
  "drop_04_1": {
    "loot_group_id": "loot_04",
    "item_id": "item_material_04",
    "weight": 0.4,
    "min_quantity": 1,
    "max_quantity": 1
  },
  "drop_04_2": {
    "loot_group_id": "loot_04",
    "item_id": "item_material_01",
    "weight": 0.3,
    "min_quantity": 1,
    "max_quantity": 2
  },
  "drop_04_3": {
    "loot_group_id": "loot_04",
    "item_id": "item_potion_04",
    "weight": 0.3,
    "min_quantity": 1,
    "max_quantity": 3
  },
  "drop_05_1": {
    "loot_group_id": "loot_05",
    "item_id": "item_material_05",
    "weight": 0.4,
    "min_quantity": 1,
    "max_quantity": 1
  },
  "drop_05_2": {
    "loot_group_id": "loot_05",
    "item_id": "item_material_06",
    "weight": 0.3,
    "min_quantity": 1,
    "max_quantity": 2
  },
  "drop_05_3": {
    "loot_group_id": "loot_05",
    "item_id": "item_potion_05",
    "weight": 0.3,
    "min_quantity": 1,
    "max_quantity": 3
  },
  "drop_06_1": {
    "loot_group_id": "loot_06",
    "item_id": "item_material_06",
    "weight": 0.4,
    "min_quantity": 1,
    "max_quantity": 1
  },
  "drop_06_2": {
    "loot_group_id": "loot_06",
    "item_id": "item_material_07",
    "weight": 0.3,
    "min_quantity": 1,
    "max_quantity": 2
  },
  "drop_06_3": {
    "loot_group_id": "loot_06",
    "item_id": "item_potion_06",
    "weight": 0.3,
    "min_quantity": 1,
    "max_quantity": 3
  },
  "drop_07_1": {
    "loot_group_id": "loot_07",
    "item_id": "item_material_07",
    "weight": 0.4,
    "min_quantity": 1,
    "max_quantity": 1
  },
  "drop_07_2": {
    "loot_group_id": "loot_07",
    "item_id": "item_material_08",
    "weight": 0.3,
    "min_quantity": 1,
    "max_quantity": 2
  },
  "drop_07_3": {
    "loot_group_id": "loot_07",
    "item_id": "item_potion_07",
    "weight": 0.3,
    "min_quantity": 1,
    "max_quantity": 3
  },
  "drop_08_1": {
    "loot_group_id": "loot_08",
    "item_id": "item_material_08",
    "weight": 0.4,
    "min_quantity": 1,
    "max_quantity": 1
  },
  "drop_08_2": {
    "loot_group_id": "loot_08",
    "item_id": "item_material_05",
    "weight": 0.3,
    "min_quantity": 1,
    "max_quantity": 2
  },
  "drop_08_3": {
    "loot_group_id": "loot_08",
    "item_id": "item_potion_08",
    "weight": 0.3,
    "min_quantity": 1,
    "max_quantity": 3
  },
  "drop_09_1": {
    "loot_group_id": "loot_09",
    "item_id": "item_material_09",
    "weight": 0.4,
    "min_quantity": 1,
    "max_quantity": 1
  },
  "drop_09_2": {
    "loot_group_id": "loot_09",
    "item_id": "item_material_10",
    "weight": 0.3,
    "min_quantity": 1,
    "max_quantity": 2
  },
  "drop_09_3": {
    "loot_group_id": "loot_09",
    "item_id": "item_potion_09",
    "weight": 0.3,
    "min_quantity": 1,
    "max_quantity": 3
  },
  "drop_10_1": {
    "loot_group_id": "loot_10",
    "item_id": "item_material_10",
    "weight": 0.4,
    "min_quantity": 1,
    "max_quantity": 1
  },
  "drop_10_2": {
    "loot_group_id": "loot_10",
    "item_id": "item_material_11",
    "weight": 0.3,
    "min_quantity": 1,
    "max_quantity": 2
  },
  "drop_10_3": {
    "loot_group_id": "loot_10",
    "item_id": "item_potion_10",
    "weight": 0.3,
    "min_quantity": 1,
    "max_quantity": 3
  },
  "drop_11_1": {
    "loot_group_id": "loot_11",
    "item_id": "item_material_11",
    "weight": 0.4,
    "min_quantity": 1,
    "max_quantity": 1
  },
  "drop_11_2": {
    "loot_group_id": "loot_11",
    "item_id": "item_material_12",
    "weight": 0.3,
    "min_quantity": 1,
    "max_quantity": 2
  },
  "drop_11_3": {
    "loot_group_id": "loot_11",
    "item_id": "item_potion_11",
    "weight": 0.3,
    "min_quantity": 1,
    "max_quantity": 3
  },
  "drop_12_1": {
    "loot_group_id": "loot_12",
    "item_id": "item_material_12",
    "weight": 0.4,
    "min_quantity": 1,
    "max_quantity": 1
  },
  "drop_12_2": {
    "loot_group_id": "loot_12",
    "item_id": "item_material_09",
    "weight": 0.3,
    "min_quantity": 1,
    "max_quantity": 2
  },
  "drop_12_3": {
    "loot_group_id": "loot_12",
    "item_id": "item_potion_12",
    "weight": 0.3,
    "min_quantity": 1,
    "max_quantity": 3
  }
}

_tool: {
  "version": 1,
  "table": {
    "namespace": "world",
    "name": "loot_entry"
  },
  "foreignKeys": {
    "loot_group_id": {
      "table": "world.loot_group"
    },
    "item_id": {
      "table": "content.item"
    }
  },
  "formulas": {}
}

_editor: {
  "table": {
    "title": "掉落明细：关系表"
  },
  "columns": {
    "loot_group_id": {
      "title": "掉落组ID"
    },
    "item_id": {
      "title": "道具ID"
    },
    "weight": {
      "title": "权重"
    },
    "min_quantity": {
      "title": "最小数量"
    },
    "max_quantity": {
      "title": "最大数量"
    }
  },
  "cellComments": {
    "drop_01_1": {
      "item_id": "一行表示一种掉落，数量和权重在同一行，避免平行数组错位。"
    }
  }
}

_export: {
  "csv": {
    "enabled": true,
    "columns": ["id", "loot_group_id", "item_id", "weight", "min_quantity", "max_quantity"],
    "encoding": {}
  }
}
