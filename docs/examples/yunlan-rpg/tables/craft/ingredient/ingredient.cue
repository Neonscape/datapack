package ingredient

import "example.com/datapack/yunlan-rpg/schema:common"

#PrimaryKey: common.#resource_id

#Row: {
    recipe_id: common.#resource_id
    material_id: common.#resource_id
    quantity: int & common.#positive
}

rows: {
  "ingredient_01_1": {
    "recipe_id": "recipe_01",
    "material_id": "item_material_01",
    "quantity": 3
  },
  "ingredient_01_2": {
    "recipe_id": "recipe_01",
    "material_id": "item_material_02",
    "quantity": 4
  },
  "ingredient_02_1": {
    "recipe_id": "recipe_02",
    "material_id": "item_material_02",
    "quantity": 3
  },
  "ingredient_02_2": {
    "recipe_id": "recipe_02",
    "material_id": "item_material_03",
    "quantity": 4
  },
  "ingredient_03_1": {
    "recipe_id": "recipe_03",
    "material_id": "item_material_03",
    "quantity": 3
  },
  "ingredient_03_2": {
    "recipe_id": "recipe_03",
    "material_id": "item_material_04",
    "quantity": 4
  },
  "ingredient_04_1": {
    "recipe_id": "recipe_04",
    "material_id": "item_material_04",
    "quantity": 3
  },
  "ingredient_04_2": {
    "recipe_id": "recipe_04",
    "material_id": "item_material_01",
    "quantity": 4
  },
  "ingredient_05_1": {
    "recipe_id": "recipe_05",
    "material_id": "item_material_05",
    "quantity": 3
  },
  "ingredient_05_2": {
    "recipe_id": "recipe_05",
    "material_id": "item_material_06",
    "quantity": 4
  },
  "ingredient_06_1": {
    "recipe_id": "recipe_06",
    "material_id": "item_material_06",
    "quantity": 3
  },
  "ingredient_06_2": {
    "recipe_id": "recipe_06",
    "material_id": "item_material_07",
    "quantity": 4
  },
  "ingredient_07_1": {
    "recipe_id": "recipe_07",
    "material_id": "item_material_07",
    "quantity": 3
  },
  "ingredient_07_2": {
    "recipe_id": "recipe_07",
    "material_id": "item_material_08",
    "quantity": 4
  },
  "ingredient_08_1": {
    "recipe_id": "recipe_08",
    "material_id": "item_material_08",
    "quantity": 3
  },
  "ingredient_08_2": {
    "recipe_id": "recipe_08",
    "material_id": "item_material_05",
    "quantity": 4
  },
  "ingredient_09_1": {
    "recipe_id": "recipe_09",
    "material_id": "item_material_09",
    "quantity": 3
  },
  "ingredient_09_2": {
    "recipe_id": "recipe_09",
    "material_id": "item_material_10",
    "quantity": 4
  },
  "ingredient_10_1": {
    "recipe_id": "recipe_10",
    "material_id": "item_material_10",
    "quantity": 3
  },
  "ingredient_10_2": {
    "recipe_id": "recipe_10",
    "material_id": "item_material_11",
    "quantity": 4
  },
  "ingredient_11_1": {
    "recipe_id": "recipe_11",
    "material_id": "item_material_11",
    "quantity": 3
  },
  "ingredient_11_2": {
    "recipe_id": "recipe_11",
    "material_id": "item_material_12",
    "quantity": 4
  },
  "ingredient_12_1": {
    "recipe_id": "recipe_12",
    "material_id": "item_material_12",
    "quantity": 3
  },
  "ingredient_12_2": {
    "recipe_id": "recipe_12",
    "material_id": "item_material_09",
    "quantity": 4
  },
  "ingredient_13_1": {
    "recipe_id": "recipe_13",
    "material_id": "item_material_01",
    "quantity": 4
  },
  "ingredient_13_2": {
    "recipe_id": "recipe_13",
    "material_id": "item_material_02",
    "quantity": 5
  },
  "ingredient_14_1": {
    "recipe_id": "recipe_14",
    "material_id": "item_material_02",
    "quantity": 4
  },
  "ingredient_14_2": {
    "recipe_id": "recipe_14",
    "material_id": "item_material_03",
    "quantity": 5
  },
  "ingredient_15_1": {
    "recipe_id": "recipe_15",
    "material_id": "item_material_03",
    "quantity": 4
  },
  "ingredient_15_2": {
    "recipe_id": "recipe_15",
    "material_id": "item_material_04",
    "quantity": 5
  },
  "ingredient_16_1": {
    "recipe_id": "recipe_16",
    "material_id": "item_material_04",
    "quantity": 4
  },
  "ingredient_16_2": {
    "recipe_id": "recipe_16",
    "material_id": "item_material_01",
    "quantity": 5
  },
  "ingredient_17_1": {
    "recipe_id": "recipe_17",
    "material_id": "item_material_05",
    "quantity": 4
  },
  "ingredient_17_2": {
    "recipe_id": "recipe_17",
    "material_id": "item_material_06",
    "quantity": 5
  },
  "ingredient_18_1": {
    "recipe_id": "recipe_18",
    "material_id": "item_material_06",
    "quantity": 4
  },
  "ingredient_18_2": {
    "recipe_id": "recipe_18",
    "material_id": "item_material_07",
    "quantity": 5
  },
  "ingredient_19_1": {
    "recipe_id": "recipe_19",
    "material_id": "item_material_07",
    "quantity": 4
  },
  "ingredient_19_2": {
    "recipe_id": "recipe_19",
    "material_id": "item_material_08",
    "quantity": 5
  },
  "ingredient_20_1": {
    "recipe_id": "recipe_20",
    "material_id": "item_material_08",
    "quantity": 4
  },
  "ingredient_20_2": {
    "recipe_id": "recipe_20",
    "material_id": "item_material_05",
    "quantity": 5
  },
  "ingredient_21_1": {
    "recipe_id": "recipe_21",
    "material_id": "item_material_09",
    "quantity": 4
  },
  "ingredient_21_2": {
    "recipe_id": "recipe_21",
    "material_id": "item_material_10",
    "quantity": 5
  },
  "ingredient_22_1": {
    "recipe_id": "recipe_22",
    "material_id": "item_material_10",
    "quantity": 4
  },
  "ingredient_22_2": {
    "recipe_id": "recipe_22",
    "material_id": "item_material_11",
    "quantity": 5
  },
  "ingredient_23_1": {
    "recipe_id": "recipe_23",
    "material_id": "item_material_11",
    "quantity": 4
  },
  "ingredient_23_2": {
    "recipe_id": "recipe_23",
    "material_id": "item_material_12",
    "quantity": 5
  },
  "ingredient_24_1": {
    "recipe_id": "recipe_24",
    "material_id": "item_material_12",
    "quantity": 4
  },
  "ingredient_24_2": {
    "recipe_id": "recipe_24",
    "material_id": "item_material_09",
    "quantity": 5
  },
  "ingredient_25_1": {
    "recipe_id": "recipe_25",
    "material_id": "item_material_01",
    "quantity": 5
  },
  "ingredient_25_2": {
    "recipe_id": "recipe_25",
    "material_id": "item_material_02",
    "quantity": 6
  },
  "ingredient_26_1": {
    "recipe_id": "recipe_26",
    "material_id": "item_material_02",
    "quantity": 5
  },
  "ingredient_26_2": {
    "recipe_id": "recipe_26",
    "material_id": "item_material_03",
    "quantity": 6
  },
  "ingredient_27_1": {
    "recipe_id": "recipe_27",
    "material_id": "item_material_03",
    "quantity": 5
  },
  "ingredient_27_2": {
    "recipe_id": "recipe_27",
    "material_id": "item_material_04",
    "quantity": 6
  },
  "ingredient_28_1": {
    "recipe_id": "recipe_28",
    "material_id": "item_material_04",
    "quantity": 5
  },
  "ingredient_28_2": {
    "recipe_id": "recipe_28",
    "material_id": "item_material_01",
    "quantity": 6
  },
  "ingredient_29_1": {
    "recipe_id": "recipe_29",
    "material_id": "item_material_05",
    "quantity": 5
  },
  "ingredient_29_2": {
    "recipe_id": "recipe_29",
    "material_id": "item_material_06",
    "quantity": 6
  },
  "ingredient_30_1": {
    "recipe_id": "recipe_30",
    "material_id": "item_material_06",
    "quantity": 5
  },
  "ingredient_30_2": {
    "recipe_id": "recipe_30",
    "material_id": "item_material_07",
    "quantity": 6
  },
  "ingredient_31_1": {
    "recipe_id": "recipe_31",
    "material_id": "item_material_07",
    "quantity": 5
  },
  "ingredient_31_2": {
    "recipe_id": "recipe_31",
    "material_id": "item_material_08",
    "quantity": 6
  },
  "ingredient_32_1": {
    "recipe_id": "recipe_32",
    "material_id": "item_material_08",
    "quantity": 5
  },
  "ingredient_32_2": {
    "recipe_id": "recipe_32",
    "material_id": "item_material_05",
    "quantity": 6
  },
  "ingredient_33_1": {
    "recipe_id": "recipe_33",
    "material_id": "item_material_09",
    "quantity": 5
  },
  "ingredient_33_2": {
    "recipe_id": "recipe_33",
    "material_id": "item_material_10",
    "quantity": 6
  },
  "ingredient_34_1": {
    "recipe_id": "recipe_34",
    "material_id": "item_material_10",
    "quantity": 5
  },
  "ingredient_34_2": {
    "recipe_id": "recipe_34",
    "material_id": "item_material_11",
    "quantity": 6
  },
  "ingredient_35_1": {
    "recipe_id": "recipe_35",
    "material_id": "item_material_11",
    "quantity": 5
  },
  "ingredient_35_2": {
    "recipe_id": "recipe_35",
    "material_id": "item_material_12",
    "quantity": 6
  },
  "ingredient_36_1": {
    "recipe_id": "recipe_36",
    "material_id": "item_material_12",
    "quantity": 5
  },
  "ingredient_36_2": {
    "recipe_id": "recipe_36",
    "material_id": "item_material_09",
    "quantity": 6
  }
}

_tool: {
  "version": 1,
  "table": {
    "namespace": "craft",
    "name": "ingredient"
  },
  "foreignKeys": {
    "recipe_id": {
      "table": "craft.recipe"
    },
    "material_id": {
      "table": "content.item"
    }
  },
  "formulas": {}
}

_editor: {
  "table": {
    "title": "配方材料：多对多关系"
  },
  "columns": {
    "recipe_id": {
      "title": "配方ID"
    },
    "material_id": {
      "title": "材料ID"
    },
    "quantity": {
      "title": "数量"
    }
  },
  "cellComments": {
    "ingredient_01_1": {
      "quantity": "一个配方有两条材料明细；修改此数量不会改变另一种材料的数量。"
    }
  }
}

_export: {
  "csv": {
    "enabled": true,
    "columns": ["id", "recipe_id", "material_id", "quantity"],
    "encoding": {}
  }
}
