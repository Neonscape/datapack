package product

import "example.com/datapack/yunlan-rpg/schema:common"

#PrimaryKey: common.#resource_id
_local: {
    // 折扣范围：仅当前表使用。
    #discount_range: >=0 & <=1
}
_schema: {
  "localDefinitions": {
    "discount_range": {
      "kind": "constraint",
      "appliesTo": ["number"]
    }
  }
}

#Row: {
    shop_id: common.#resource_id
    item_id: common.#resource_id
    list_price: int & common.#non_negative
    discount: number & _local.#discount_range
    sale_price: int & common.#non_negative
    daily_limit: int & common.#positive
    listed: *true | bool
    designer_notes: string | null
}

rows: {
  "product_01_1": {
    "shop_id": "shop_01",
    "item_id": "item_potion_01",
    "list_price": 40,
    "discount": 0.8,
    "daily_limit": 5,
    "listed": true,
    "designer_notes": "标价为人工同步的商店定价，不是跨表公式。"
  },
  "product_01_2": {
    "shop_id": "shop_01",
    "item_id": "item_weapon_01",
    "list_price": 140,
    "discount": 1,
    "daily_limit": 1,
    "listed": true,
    "designer_notes": "标价为人工同步的商店定价，不是跨表公式。"
  },
  "product_02_1": {
    "shop_id": "shop_02",
    "item_id": "item_potion_02",
    "list_price": 60,
    "discount": 0.8,
    "daily_limit": 5,
    "listed": true,
    "designer_notes": "标价为人工同步的商店定价，不是跨表公式。"
  },
  "product_02_2": {
    "shop_id": "shop_02",
    "item_id": "item_weapon_02",
    "list_price": 160,
    "discount": 1,
    "daily_limit": 1,
    "listed": true,
    "designer_notes": "标价为人工同步的商店定价，不是跨表公式。"
  },
  "product_03_1": {
    "shop_id": "shop_03",
    "item_id": "item_potion_03",
    "list_price": 80,
    "discount": 0.8,
    "daily_limit": 5,
    "listed": true,
    "designer_notes": "标价为人工同步的商店定价，不是跨表公式。"
  },
  "product_03_2": {
    "shop_id": "shop_03",
    "item_id": "item_weapon_03",
    "list_price": 180,
    "discount": 1,
    "daily_limit": 1,
    "listed": true,
    "designer_notes": "标价为人工同步的商店定价，不是跨表公式。"
  },
  "product_04_1": {
    "shop_id": "shop_04",
    "item_id": "item_potion_04",
    "list_price": 100,
    "discount": 0.8,
    "daily_limit": 5,
    "listed": true,
    "designer_notes": "标价为人工同步的商店定价，不是跨表公式。"
  },
  "product_04_2": {
    "shop_id": "shop_04",
    "item_id": "item_weapon_04",
    "list_price": 200,
    "discount": 1,
    "daily_limit": 1,
    "listed": true,
    "designer_notes": "标价为人工同步的商店定价，不是跨表公式。"
  },
  "product_05_1": {
    "shop_id": "shop_05",
    "item_id": "item_potion_05",
    "list_price": 100,
    "discount": 0.8,
    "daily_limit": 5,
    "listed": true,
    "designer_notes": "标价为人工同步的商店定价，不是跨表公式。"
  },
  "product_05_2": {
    "shop_id": "shop_05",
    "item_id": "item_weapon_05",
    "list_price": 200,
    "discount": 1,
    "daily_limit": 1,
    "listed": true,
    "designer_notes": "标价为人工同步的商店定价，不是跨表公式。"
  },
  "product_06_1": {
    "shop_id": "shop_06",
    "item_id": "item_potion_06",
    "list_price": 120,
    "discount": 0.8,
    "daily_limit": 5,
    "listed": true,
    "designer_notes": "标价为人工同步的商店定价，不是跨表公式。"
  },
  "product_06_2": {
    "shop_id": "shop_06",
    "item_id": "item_weapon_06",
    "list_price": 220,
    "discount": 1,
    "daily_limit": 1,
    "listed": true,
    "designer_notes": "标价为人工同步的商店定价，不是跨表公式。"
  },
  "product_07_1": {
    "shop_id": "shop_07",
    "item_id": "item_potion_07",
    "list_price": 140,
    "discount": 0.8,
    "daily_limit": 5,
    "listed": true,
    "designer_notes": "标价为人工同步的商店定价，不是跨表公式。"
  },
  "product_07_2": {
    "shop_id": "shop_07",
    "item_id": "item_weapon_07",
    "list_price": 240,
    "discount": 1,
    "daily_limit": 1,
    "listed": true,
    "designer_notes": "标价为人工同步的商店定价，不是跨表公式。"
  },
  "product_08_1": {
    "shop_id": "shop_08",
    "item_id": "item_potion_08",
    "list_price": 160,
    "discount": 0.8,
    "daily_limit": 5,
    "listed": true,
    "designer_notes": "标价为人工同步的商店定价，不是跨表公式。"
  },
  "product_08_2": {
    "shop_id": "shop_08",
    "item_id": "item_weapon_08",
    "list_price": 260,
    "discount": 1,
    "daily_limit": 1,
    "listed": true,
    "designer_notes": "标价为人工同步的商店定价，不是跨表公式。"
  },
  "product_09_1": {
    "shop_id": "shop_09",
    "item_id": "item_potion_09",
    "list_price": 160,
    "discount": 0.8,
    "daily_limit": 5,
    "listed": true,
    "designer_notes": "标价为人工同步的商店定价，不是跨表公式。"
  },
  "product_09_2": {
    "shop_id": "shop_09",
    "item_id": "item_weapon_09",
    "list_price": 260,
    "discount": 1,
    "daily_limit": 1,
    "listed": true,
    "designer_notes": "标价为人工同步的商店定价，不是跨表公式。"
  },
  "product_10_1": {
    "shop_id": "shop_10",
    "item_id": "item_potion_10",
    "list_price": 180,
    "discount": 0.8,
    "daily_limit": 5,
    "listed": true,
    "designer_notes": "标价为人工同步的商店定价，不是跨表公式。"
  },
  "product_10_2": {
    "shop_id": "shop_10",
    "item_id": "item_weapon_10",
    "list_price": 280,
    "discount": 1,
    "daily_limit": 1,
    "listed": true,
    "designer_notes": "标价为人工同步的商店定价，不是跨表公式。"
  },
  "product_11_1": {
    "shop_id": "shop_11",
    "item_id": "item_potion_11",
    "list_price": 200,
    "discount": 0.8,
    "daily_limit": 5,
    "listed": true,
    "designer_notes": "标价为人工同步的商店定价，不是跨表公式。"
  },
  "product_11_2": {
    "shop_id": "shop_11",
    "item_id": "item_weapon_11",
    "list_price": 300,
    "discount": 1,
    "daily_limit": 1,
    "listed": true,
    "designer_notes": "标价为人工同步的商店定价，不是跨表公式。"
  },
  "product_12_1": {
    "shop_id": "shop_12",
    "item_id": "item_potion_12",
    "list_price": 220,
    "discount": 0.8,
    "daily_limit": 5,
    "listed": true,
    "designer_notes": "标价为人工同步的商店定价，不是跨表公式。"
  },
  "product_12_2": {
    "shop_id": "shop_12",
    "item_id": "item_weapon_12",
    "list_price": 320,
    "discount": 1,
    "daily_limit": 1,
    "listed": true,
    "designer_notes": "标价为人工同步的商店定价，不是跨表公式。"
  }
}

_tool: {
  "version": 1,
  "table": {
    "namespace": "economy",
    "name": "product"
  },
  "foreignKeys": {
    "shop_id": {
      "table": "economy.shop"
    },
    "item_id": {
      "table": "content.item"
    }
  },
  "formulas": {
    "product_01_1": {
      "sale_price": "ceil(list_price * discount)"
    },
    "product_01_2": {
      "sale_price": "ceil(list_price * discount)"
    },
    "product_02_1": {
      "sale_price": "ceil(list_price * discount)"
    },
    "product_02_2": {
      "sale_price": "ceil(list_price * discount)"
    },
    "product_03_1": {
      "sale_price": "ceil(list_price * discount)"
    },
    "product_03_2": {
      "sale_price": "ceil(list_price * discount)"
    },
    "product_04_1": {
      "sale_price": "ceil(list_price * discount)"
    },
    "product_04_2": {
      "sale_price": "ceil(list_price * discount)"
    },
    "product_05_1": {
      "sale_price": "ceil(list_price * discount)"
    },
    "product_05_2": {
      "sale_price": "ceil(list_price * discount)"
    },
    "product_06_1": {
      "sale_price": "ceil(list_price * discount)"
    },
    "product_06_2": {
      "sale_price": "ceil(list_price * discount)"
    },
    "product_07_1": {
      "sale_price": "ceil(list_price * discount)"
    },
    "product_07_2": {
      "sale_price": "ceil(list_price * discount)"
    },
    "product_08_1": {
      "sale_price": "ceil(list_price * discount)"
    },
    "product_08_2": {
      "sale_price": "ceil(list_price * discount)"
    },
    "product_09_1": {
      "sale_price": "ceil(list_price * discount)"
    },
    "product_09_2": {
      "sale_price": "ceil(list_price * discount)"
    },
    "product_10_1": {
      "sale_price": "ceil(list_price * discount)"
    },
    "product_10_2": {
      "sale_price": "ceil(list_price * discount)"
    },
    "product_11_1": {
      "sale_price": "ceil(list_price * discount)"
    },
    "product_11_2": {
      "sale_price": "ceil(list_price * discount)"
    },
    "product_12_1": {
      "sale_price": "ceil(list_price * discount)"
    },
    "product_12_2": {
      "sale_price": "ceil(list_price * discount)"
    }
  }
}

_editor: {
  "table": {
    "title": "商品：折扣与导出练习"
  },
  "columns": {
    "shop_id": {
      "title": "商店ID"
    },
    "item_id": {
      "title": "道具ID"
    },
    "list_price": {
      "title": "标价"
    },
    "discount": {
      "title": "折扣"
    },
    "sale_price": {
      "title": "实售价"
    },
    "daily_limit": {
      "title": "每日限购"
    },
    "listed": {
      "title": "上架"
    },
    "designer_notes": {
      "title": "策划备注"
    }
  },
  "cellComments": {
    "product_01_1": {
      "sale_price": "ceil(40×0.8)=32；把标价改100，导出应为80。",
      "list_price": "商店独立定价，需要策划主动与道具标价保持关系，工具不自动join。"
    },
    "product_01_2": {
      "discount": "1表示原价；0表示免费，范围由本地约束控制。"
    }
  }
}

_export: {
  "csv": {
    "enabled": true,
    "columns": ["id", "shop_id", "item_id", "list_price", "discount", "sale_price", "daily_limit", "listed"],
    "encoding": {}
  }
}
