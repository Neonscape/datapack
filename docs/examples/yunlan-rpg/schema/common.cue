package common

#资源标识: string & =~"^[a-z][a-z0-9_]*$"
#必填文本: string & != ""
#非负: >=0
#正数: >0
#概率范围: >=0 & <=1
#概率: number & #概率范围
#稀有度: "普通" | "精良" | "稀有" | "史诗"
#道具类别: "消耗品" | "材料" | "武器" | "防具" | "饰品" | "任务物品"
#效果类别: "伤害" | "治疗" | "控制" | "增益"
#章节: "青岚村" | "赤砂遗迹" | "霜月城"
#角色定位: "战士" | "游侠" | "医师" | "术士"
#装备槽位: "主手" | "躯干" | "饰品"

_schema: {
  "version": 1,
  "definitions": {
    "资源标识": {
      "kind": "type",
      "type": "string"
    },
    "必填文本": {
      "kind": "type",
      "type": "string"
    },
    "非负": {
      "kind": "constraint",
      "appliesTo": ["integer", "number"]
    },
    "正数": {
      "kind": "constraint",
      "appliesTo": ["integer", "number"]
    },
    "概率范围": {
      "kind": "constraint",
      "appliesTo": ["integer", "number"]
    },
    "概率": {
      "kind": "type",
      "type": "number"
    },
    "稀有度": {
      "kind": "type",
      "type": "enum"
    },
    "道具类别": {
      "kind": "type",
      "type": "enum"
    },
    "效果类别": {
      "kind": "type",
      "type": "enum"
    },
    "章节": {
      "kind": "type",
      "type": "enum"
    },
    "角色定位": {
      "kind": "type",
      "type": "enum"
    },
    "装备槽位": {
      "kind": "type",
      "type": "enum"
    }
  }
}
