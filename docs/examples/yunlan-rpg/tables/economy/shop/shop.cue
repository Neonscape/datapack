package shop

import "example.com/datapack/yunlan-rpg/schema:common"

#PrimaryKey: common.#资源标识

#Row: {
    名称文本ID: common.#资源标识
    章节: common.#章节
    货币: "金币"
    开放任务ID: string | null
    营业中: bool
}

rows: {
  "shop_01": {
    "名称文本ID": "shop_01_name",
    "章节": "青岚村",
    "货币": "金币",
    "开放任务ID": null,
    "营业中": true
  },
  "shop_02": {
    "名称文本ID": "shop_02_name",
    "章节": "青岚村",
    "货币": "金币",
    "开放任务ID": "quest_05",
    "营业中": true
  },
  "shop_03": {
    "名称文本ID": "shop_03_name",
    "章节": "青岚村",
    "货币": "金币",
    "开放任务ID": "quest_09",
    "营业中": true
  },
  "shop_04": {
    "名称文本ID": "shop_04_name",
    "章节": "青岚村",
    "货币": "金币",
    "开放任务ID": "quest_13",
    "营业中": true
  },
  "shop_05": {
    "名称文本ID": "shop_05_name",
    "章节": "赤砂遗迹",
    "货币": "金币",
    "开放任务ID": "quest_17",
    "营业中": true
  },
  "shop_06": {
    "名称文本ID": "shop_06_name",
    "章节": "赤砂遗迹",
    "货币": "金币",
    "开放任务ID": "quest_21",
    "营业中": true
  },
  "shop_07": {
    "名称文本ID": "shop_07_name",
    "章节": "赤砂遗迹",
    "货币": "金币",
    "开放任务ID": "quest_25",
    "营业中": true
  },
  "shop_08": {
    "名称文本ID": "shop_08_name",
    "章节": "赤砂遗迹",
    "货币": "金币",
    "开放任务ID": "quest_29",
    "营业中": true
  },
  "shop_09": {
    "名称文本ID": "shop_09_name",
    "章节": "霜月城",
    "货币": "金币",
    "开放任务ID": "quest_33",
    "营业中": true
  },
  "shop_10": {
    "名称文本ID": "shop_10_name",
    "章节": "霜月城",
    "货币": "金币",
    "开放任务ID": "quest_37",
    "营业中": true
  },
  "shop_11": {
    "名称文本ID": "shop_11_name",
    "章节": "霜月城",
    "货币": "金币",
    "开放任务ID": "quest_41",
    "营业中": true
  },
  "shop_12": {
    "名称文本ID": "shop_12_name",
    "章节": "霜月城",
    "货币": "金币",
    "开放任务ID": "quest_45",
    "营业中": true
  }
}

_tool: {
  "version": 1,
  "table": {
    "namespace": "economy",
    "name": "shop"
  },
  "foreignKeys": {
    "名称文本ID": {
      "table": "localization.text"
    },
    "开放任务ID": {
      "table": "quest.quest"
    }
  },
  "formulas": {}
}

_editor: {
  "table": {
    "title": "商店：区域经营"
  },
  "cellComments": {
    "shop_01": {
      "开放任务ID": "新手药铺默认开放，其余商店在对应任务后解锁。"
    }
  }
}

_export: {
  "csv": {
    "enabled": true,
    "columns": ["id", "名称文本ID", "章节", "货币", "开放任务ID", "营业中"],
    "encoding": {}
  }
}
