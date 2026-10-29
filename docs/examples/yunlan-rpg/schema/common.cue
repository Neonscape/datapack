package common

// 资源标识：英文标识供程序与Schema引用，中文说明供策划阅读。
#resource_id: string & =~"^[a-z][a-z0-9_]*$"
// 必填文本：英文标识供程序与Schema引用，中文说明供策划阅读。
#non_empty_text: string & != ""
// 非负：英文标识供程序与Schema引用，中文说明供策划阅读。
#non_negative: >=0
// 正数：英文标识供程序与Schema引用，中文说明供策划阅读。
#positive: >0
// 概率范围：英文标识供程序与Schema引用，中文说明供策划阅读。
#probability_range: >=0 & <=1
// 概率：英文标识供程序与Schema引用，中文说明供策划阅读。
#probability: number & #probability_range
// 稀有度：英文标识供程序与Schema引用，中文说明供策划阅读。
#rarity: "普通" | "精良" | "稀有" | "史诗"
// 道具类别：英文标识供程序与Schema引用，中文说明供策划阅读。
#item_category: "消耗品" | "材料" | "武器" | "防具" | "饰品" | "任务物品"
// 效果类别：英文标识供程序与Schema引用，中文说明供策划阅读。
#effect_category: "伤害" | "治疗" | "控制" | "增益"
// 章节：英文标识供程序与Schema引用，中文说明供策划阅读。
#chapter: "青岚村" | "赤砂遗迹" | "霜月城"
// 角色定位：英文标识供程序与Schema引用，中文说明供策划阅读。
#hero_role: "战士" | "游侠" | "医师" | "术士"
// 装备槽位：英文标识供程序与Schema引用，中文说明供策划阅读。
#equipment_slot: "主手" | "躯干" | "饰品"

_schema: {
  "version": 1,
  "definitions": {
    "resource_id": {
      "kind": "type",
      "type": "string"
    },
    "non_empty_text": {
      "kind": "type",
      "type": "string"
    },
    "non_negative": {
      "kind": "constraint",
      "appliesTo": ["integer", "number"]
    },
    "positive": {
      "kind": "constraint",
      "appliesTo": ["integer", "number"]
    },
    "probability_range": {
      "kind": "constraint",
      "appliesTo": ["integer", "number"]
    },
    "probability": {
      "kind": "type",
      "type": "number"
    },
    "rarity": {
      "kind": "type",
      "type": "enum"
    },
    "item_category": {
      "kind": "type",
      "type": "enum"
    },
    "effect_category": {
      "kind": "type",
      "type": "enum"
    },
    "chapter": {
      "kind": "type",
      "type": "enum"
    },
    "hero_role": {
      "kind": "type",
      "type": "enum"
    },
    "equipment_slot": {
      "kind": "type",
      "type": "enum"
    }
  }
}
