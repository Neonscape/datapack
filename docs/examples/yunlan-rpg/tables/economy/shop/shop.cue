package shop

import "example.com/datapack/yunlan-rpg/schema:common"

#PrimaryKey: common.#resource_id

#Row: {
    name_text_id: common.#resource_id
    chapter: common.#chapter
    currency: "金币"
    unlock_quest_id: string | null
    is_open: bool
}

rows: {
  "shop_01": {
    "name_text_id": "shop_01_name",
    "chapter": "青岚村",
    "currency": "金币",
    "unlock_quest_id": null,
    "is_open": true
  },
  "shop_02": {
    "name_text_id": "shop_02_name",
    "chapter": "青岚村",
    "currency": "金币",
    "unlock_quest_id": "quest_05",
    "is_open": true
  },
  "shop_03": {
    "name_text_id": "shop_03_name",
    "chapter": "青岚村",
    "currency": "金币",
    "unlock_quest_id": "quest_09",
    "is_open": true
  },
  "shop_04": {
    "name_text_id": "shop_04_name",
    "chapter": "青岚村",
    "currency": "金币",
    "unlock_quest_id": "quest_13",
    "is_open": true
  },
  "shop_05": {
    "name_text_id": "shop_05_name",
    "chapter": "赤砂遗迹",
    "currency": "金币",
    "unlock_quest_id": "quest_17",
    "is_open": true
  },
  "shop_06": {
    "name_text_id": "shop_06_name",
    "chapter": "赤砂遗迹",
    "currency": "金币",
    "unlock_quest_id": "quest_21",
    "is_open": true
  },
  "shop_07": {
    "name_text_id": "shop_07_name",
    "chapter": "赤砂遗迹",
    "currency": "金币",
    "unlock_quest_id": "quest_25",
    "is_open": true
  },
  "shop_08": {
    "name_text_id": "shop_08_name",
    "chapter": "赤砂遗迹",
    "currency": "金币",
    "unlock_quest_id": "quest_29",
    "is_open": true
  },
  "shop_09": {
    "name_text_id": "shop_09_name",
    "chapter": "霜月城",
    "currency": "金币",
    "unlock_quest_id": "quest_33",
    "is_open": true
  },
  "shop_10": {
    "name_text_id": "shop_10_name",
    "chapter": "霜月城",
    "currency": "金币",
    "unlock_quest_id": "quest_37",
    "is_open": true
  },
  "shop_11": {
    "name_text_id": "shop_11_name",
    "chapter": "霜月城",
    "currency": "金币",
    "unlock_quest_id": "quest_41",
    "is_open": true
  },
  "shop_12": {
    "name_text_id": "shop_12_name",
    "chapter": "霜月城",
    "currency": "金币",
    "unlock_quest_id": "quest_45",
    "is_open": true
  }
}

_tool: {
  "version": 1,
  "table": {
    "namespace": "economy",
    "name": "shop"
  },
  "foreignKeys": {
    "name_text_id": {
      "table": "localization.text"
    },
    "unlock_quest_id": {
      "table": "quest.quest"
    }
  },
  "formulas": {}
}

_editor: {
  "table": {
    "title": "商店：区域经营"
  },
  "columns": {
    "name_text_id": {
      "title": "名称文本ID"
    },
    "chapter": {
      "title": "章节"
    },
    "currency": {
      "title": "货币"
    },
    "unlock_quest_id": {
      "title": "解锁任务ID"
    },
    "is_open": {
      "title": "营业中"
    }
  },
  "cellComments": {
    "shop_01": {
      "unlock_quest_id": "新手药铺默认开放，其余商店在对应任务后解锁。"
    }
  }
}

_export: {
  "csv": {
    "enabled": true,
    "columns": ["id", "name_text_id", "chapter", "currency", "unlock_quest_id", "is_open"],
    "encoding": {}
  }
}
