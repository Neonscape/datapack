package ingredient

import "example.com/datapack/yunlan-rpg/schema:common"

#PrimaryKey: common.#资源标识

#Row: {
    配方ID: common.#资源标识
    材料ID: common.#资源标识
    数量: int & common.#正数
}

rows: {
  "ingredient_01_1": {
    "配方ID": "recipe_01",
    "材料ID": "item_material_01",
    "数量": 3
  },
  "ingredient_01_2": {
    "配方ID": "recipe_01",
    "材料ID": "item_material_02",
    "数量": 4
  },
  "ingredient_02_1": {
    "配方ID": "recipe_02",
    "材料ID": "item_material_02",
    "数量": 3
  },
  "ingredient_02_2": {
    "配方ID": "recipe_02",
    "材料ID": "item_material_03",
    "数量": 4
  },
  "ingredient_03_1": {
    "配方ID": "recipe_03",
    "材料ID": "item_material_03",
    "数量": 3
  },
  "ingredient_03_2": {
    "配方ID": "recipe_03",
    "材料ID": "item_material_04",
    "数量": 4
  },
  "ingredient_04_1": {
    "配方ID": "recipe_04",
    "材料ID": "item_material_04",
    "数量": 3
  },
  "ingredient_04_2": {
    "配方ID": "recipe_04",
    "材料ID": "item_material_01",
    "数量": 4
  },
  "ingredient_05_1": {
    "配方ID": "recipe_05",
    "材料ID": "item_material_05",
    "数量": 3
  },
  "ingredient_05_2": {
    "配方ID": "recipe_05",
    "材料ID": "item_material_06",
    "数量": 4
  },
  "ingredient_06_1": {
    "配方ID": "recipe_06",
    "材料ID": "item_material_06",
    "数量": 3
  },
  "ingredient_06_2": {
    "配方ID": "recipe_06",
    "材料ID": "item_material_07",
    "数量": 4
  },
  "ingredient_07_1": {
    "配方ID": "recipe_07",
    "材料ID": "item_material_07",
    "数量": 3
  },
  "ingredient_07_2": {
    "配方ID": "recipe_07",
    "材料ID": "item_material_08",
    "数量": 4
  },
  "ingredient_08_1": {
    "配方ID": "recipe_08",
    "材料ID": "item_material_08",
    "数量": 3
  },
  "ingredient_08_2": {
    "配方ID": "recipe_08",
    "材料ID": "item_material_05",
    "数量": 4
  },
  "ingredient_09_1": {
    "配方ID": "recipe_09",
    "材料ID": "item_material_09",
    "数量": 3
  },
  "ingredient_09_2": {
    "配方ID": "recipe_09",
    "材料ID": "item_material_10",
    "数量": 4
  },
  "ingredient_10_1": {
    "配方ID": "recipe_10",
    "材料ID": "item_material_10",
    "数量": 3
  },
  "ingredient_10_2": {
    "配方ID": "recipe_10",
    "材料ID": "item_material_11",
    "数量": 4
  },
  "ingredient_11_1": {
    "配方ID": "recipe_11",
    "材料ID": "item_material_11",
    "数量": 3
  },
  "ingredient_11_2": {
    "配方ID": "recipe_11",
    "材料ID": "item_material_12",
    "数量": 4
  },
  "ingredient_12_1": {
    "配方ID": "recipe_12",
    "材料ID": "item_material_12",
    "数量": 3
  },
  "ingredient_12_2": {
    "配方ID": "recipe_12",
    "材料ID": "item_material_09",
    "数量": 4
  },
  "ingredient_13_1": {
    "配方ID": "recipe_13",
    "材料ID": "item_material_01",
    "数量": 4
  },
  "ingredient_13_2": {
    "配方ID": "recipe_13",
    "材料ID": "item_material_02",
    "数量": 5
  },
  "ingredient_14_1": {
    "配方ID": "recipe_14",
    "材料ID": "item_material_02",
    "数量": 4
  },
  "ingredient_14_2": {
    "配方ID": "recipe_14",
    "材料ID": "item_material_03",
    "数量": 5
  },
  "ingredient_15_1": {
    "配方ID": "recipe_15",
    "材料ID": "item_material_03",
    "数量": 4
  },
  "ingredient_15_2": {
    "配方ID": "recipe_15",
    "材料ID": "item_material_04",
    "数量": 5
  },
  "ingredient_16_1": {
    "配方ID": "recipe_16",
    "材料ID": "item_material_04",
    "数量": 4
  },
  "ingredient_16_2": {
    "配方ID": "recipe_16",
    "材料ID": "item_material_01",
    "数量": 5
  },
  "ingredient_17_1": {
    "配方ID": "recipe_17",
    "材料ID": "item_material_05",
    "数量": 4
  },
  "ingredient_17_2": {
    "配方ID": "recipe_17",
    "材料ID": "item_material_06",
    "数量": 5
  },
  "ingredient_18_1": {
    "配方ID": "recipe_18",
    "材料ID": "item_material_06",
    "数量": 4
  },
  "ingredient_18_2": {
    "配方ID": "recipe_18",
    "材料ID": "item_material_07",
    "数量": 5
  },
  "ingredient_19_1": {
    "配方ID": "recipe_19",
    "材料ID": "item_material_07",
    "数量": 4
  },
  "ingredient_19_2": {
    "配方ID": "recipe_19",
    "材料ID": "item_material_08",
    "数量": 5
  },
  "ingredient_20_1": {
    "配方ID": "recipe_20",
    "材料ID": "item_material_08",
    "数量": 4
  },
  "ingredient_20_2": {
    "配方ID": "recipe_20",
    "材料ID": "item_material_05",
    "数量": 5
  },
  "ingredient_21_1": {
    "配方ID": "recipe_21",
    "材料ID": "item_material_09",
    "数量": 4
  },
  "ingredient_21_2": {
    "配方ID": "recipe_21",
    "材料ID": "item_material_10",
    "数量": 5
  },
  "ingredient_22_1": {
    "配方ID": "recipe_22",
    "材料ID": "item_material_10",
    "数量": 4
  },
  "ingredient_22_2": {
    "配方ID": "recipe_22",
    "材料ID": "item_material_11",
    "数量": 5
  },
  "ingredient_23_1": {
    "配方ID": "recipe_23",
    "材料ID": "item_material_11",
    "数量": 4
  },
  "ingredient_23_2": {
    "配方ID": "recipe_23",
    "材料ID": "item_material_12",
    "数量": 5
  },
  "ingredient_24_1": {
    "配方ID": "recipe_24",
    "材料ID": "item_material_12",
    "数量": 4
  },
  "ingredient_24_2": {
    "配方ID": "recipe_24",
    "材料ID": "item_material_09",
    "数量": 5
  },
  "ingredient_25_1": {
    "配方ID": "recipe_25",
    "材料ID": "item_material_01",
    "数量": 5
  },
  "ingredient_25_2": {
    "配方ID": "recipe_25",
    "材料ID": "item_material_02",
    "数量": 6
  },
  "ingredient_26_1": {
    "配方ID": "recipe_26",
    "材料ID": "item_material_02",
    "数量": 5
  },
  "ingredient_26_2": {
    "配方ID": "recipe_26",
    "材料ID": "item_material_03",
    "数量": 6
  },
  "ingredient_27_1": {
    "配方ID": "recipe_27",
    "材料ID": "item_material_03",
    "数量": 5
  },
  "ingredient_27_2": {
    "配方ID": "recipe_27",
    "材料ID": "item_material_04",
    "数量": 6
  },
  "ingredient_28_1": {
    "配方ID": "recipe_28",
    "材料ID": "item_material_04",
    "数量": 5
  },
  "ingredient_28_2": {
    "配方ID": "recipe_28",
    "材料ID": "item_material_01",
    "数量": 6
  },
  "ingredient_29_1": {
    "配方ID": "recipe_29",
    "材料ID": "item_material_05",
    "数量": 5
  },
  "ingredient_29_2": {
    "配方ID": "recipe_29",
    "材料ID": "item_material_06",
    "数量": 6
  },
  "ingredient_30_1": {
    "配方ID": "recipe_30",
    "材料ID": "item_material_06",
    "数量": 5
  },
  "ingredient_30_2": {
    "配方ID": "recipe_30",
    "材料ID": "item_material_07",
    "数量": 6
  },
  "ingredient_31_1": {
    "配方ID": "recipe_31",
    "材料ID": "item_material_07",
    "数量": 5
  },
  "ingredient_31_2": {
    "配方ID": "recipe_31",
    "材料ID": "item_material_08",
    "数量": 6
  },
  "ingredient_32_1": {
    "配方ID": "recipe_32",
    "材料ID": "item_material_08",
    "数量": 5
  },
  "ingredient_32_2": {
    "配方ID": "recipe_32",
    "材料ID": "item_material_05",
    "数量": 6
  },
  "ingredient_33_1": {
    "配方ID": "recipe_33",
    "材料ID": "item_material_09",
    "数量": 5
  },
  "ingredient_33_2": {
    "配方ID": "recipe_33",
    "材料ID": "item_material_10",
    "数量": 6
  },
  "ingredient_34_1": {
    "配方ID": "recipe_34",
    "材料ID": "item_material_10",
    "数量": 5
  },
  "ingredient_34_2": {
    "配方ID": "recipe_34",
    "材料ID": "item_material_11",
    "数量": 6
  },
  "ingredient_35_1": {
    "配方ID": "recipe_35",
    "材料ID": "item_material_11",
    "数量": 5
  },
  "ingredient_35_2": {
    "配方ID": "recipe_35",
    "材料ID": "item_material_12",
    "数量": 6
  },
  "ingredient_36_1": {
    "配方ID": "recipe_36",
    "材料ID": "item_material_12",
    "数量": 5
  },
  "ingredient_36_2": {
    "配方ID": "recipe_36",
    "材料ID": "item_material_09",
    "数量": 6
  }
}

_tool: {
  "version": 1,
  "table": {
    "namespace": "craft",
    "name": "ingredient"
  },
  "foreignKeys": {
    "配方ID": {
      "table": "craft.recipe"
    },
    "材料ID": {
      "table": "content.item"
    }
  },
  "formulas": {}
}

_editor: {
  "table": {
    "title": "配方材料：多对多关系"
  },
  "cellComments": {
    "ingredient_01_1": {
      "数量": "一个配方有两条材料明细；修改此数量不会改变另一种材料的数量。"
    }
  }
}

_export: {
  "csv": {
    "enabled": true,
    "columns": ["id", "配方ID", "材料ID", "数量"],
    "encoding": {}
  }
}
