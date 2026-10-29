package growth

import "example.com/datapack/yunlan-rpg/schema:common"

#PrimaryKey: common.#resource_id
_local: {
    // 等级：仅当前表使用。
    #level_range: >=1 & <=60
}
_schema: {
  "localDefinitions": {
    "level_range": {
      "kind": "constraint",
      "appliesTo": ["integer"]
    }
  }
}

#Row: {
    level: int & _local.#level_range
    level_up_xp: int & common.#positive
    total_xp: int & common.#non_negative
    base_attack: int & common.#non_negative
    base_defense: int & common.#non_negative
    base_hp: int & common.#positive
    power_rating: int & common.#non_negative
}

rows: {
  "level_01": {
    "level": 1,
    "level_up_xp": 103,
    "base_attack": 23,
    "base_defense": 12,
    "base_hp": 225
  },
  "level_02": {
    "level": 2,
    "level_up_xp": 132,
    "base_attack": 26,
    "base_defense": 14,
    "base_hp": 250
  },
  "level_03": {
    "level": 3,
    "level_up_xp": 167,
    "base_attack": 29,
    "base_defense": 16,
    "base_hp": 275
  },
  "level_04": {
    "level": 4,
    "level_up_xp": 208,
    "base_attack": 32,
    "base_defense": 18,
    "base_hp": 300
  },
  "level_05": {
    "level": 5,
    "level_up_xp": 255,
    "base_attack": 35,
    "base_defense": 20,
    "base_hp": 325
  },
  "level_06": {
    "level": 6,
    "level_up_xp": 308,
    "base_attack": 38,
    "base_defense": 22,
    "base_hp": 350
  },
  "level_07": {
    "level": 7,
    "level_up_xp": 367,
    "base_attack": 41,
    "base_defense": 24,
    "base_hp": 375
  },
  "level_08": {
    "level": 8,
    "level_up_xp": 432,
    "base_attack": 44,
    "base_defense": 26,
    "base_hp": 400
  },
  "level_09": {
    "level": 9,
    "level_up_xp": 503,
    "base_attack": 47,
    "base_defense": 28,
    "base_hp": 425
  },
  "level_10": {
    "level": 10,
    "level_up_xp": 580,
    "base_attack": 50,
    "base_defense": 30,
    "base_hp": 450
  },
  "level_11": {
    "level": 11,
    "level_up_xp": 663,
    "base_attack": 53,
    "base_defense": 32,
    "base_hp": 475
  },
  "level_12": {
    "level": 12,
    "level_up_xp": 752,
    "base_attack": 56,
    "base_defense": 34,
    "base_hp": 500
  },
  "level_13": {
    "level": 13,
    "level_up_xp": 847,
    "base_attack": 59,
    "base_defense": 36,
    "base_hp": 525
  },
  "level_14": {
    "level": 14,
    "level_up_xp": 948,
    "base_attack": 62,
    "base_defense": 38,
    "base_hp": 550
  },
  "level_15": {
    "level": 15,
    "level_up_xp": 1055,
    "base_attack": 65,
    "base_defense": 40,
    "base_hp": 575
  },
  "level_16": {
    "level": 16,
    "level_up_xp": 1168,
    "base_attack": 68,
    "base_defense": 42,
    "base_hp": 600
  },
  "level_17": {
    "level": 17,
    "level_up_xp": 1287,
    "base_attack": 71,
    "base_defense": 44,
    "base_hp": 625
  },
  "level_18": {
    "level": 18,
    "level_up_xp": 1412,
    "base_attack": 74,
    "base_defense": 46,
    "base_hp": 650
  },
  "level_19": {
    "level": 19,
    "level_up_xp": 1543,
    "base_attack": 77,
    "base_defense": 48,
    "base_hp": 675
  },
  "level_20": {
    "level": 20,
    "level_up_xp": 1680,
    "base_attack": 80,
    "base_defense": 50,
    "base_hp": 700
  },
  "level_21": {
    "level": 21,
    "level_up_xp": 1823,
    "base_attack": 83,
    "base_defense": 52,
    "base_hp": 725
  },
  "level_22": {
    "level": 22,
    "level_up_xp": 1972,
    "base_attack": 86,
    "base_defense": 54,
    "base_hp": 750
  },
  "level_23": {
    "level": 23,
    "level_up_xp": 2127,
    "base_attack": 89,
    "base_defense": 56,
    "base_hp": 775
  },
  "level_24": {
    "level": 24,
    "level_up_xp": 2288,
    "base_attack": 92,
    "base_defense": 58,
    "base_hp": 800
  },
  "level_25": {
    "level": 25,
    "level_up_xp": 2455,
    "base_attack": 95,
    "base_defense": 60,
    "base_hp": 825
  },
  "level_26": {
    "level": 26,
    "level_up_xp": 2628,
    "base_attack": 98,
    "base_defense": 62,
    "base_hp": 850
  },
  "level_27": {
    "level": 27,
    "level_up_xp": 2807,
    "base_attack": 101,
    "base_defense": 64,
    "base_hp": 875
  },
  "level_28": {
    "level": 28,
    "level_up_xp": 2992,
    "base_attack": 104,
    "base_defense": 66,
    "base_hp": 900
  },
  "level_29": {
    "level": 29,
    "level_up_xp": 3183,
    "base_attack": 107,
    "base_defense": 68,
    "base_hp": 925
  },
  "level_30": {
    "level": 30,
    "level_up_xp": 3380,
    "base_attack": 110,
    "base_defense": 70,
    "base_hp": 950
  },
  "level_31": {
    "level": 31,
    "level_up_xp": 3583,
    "base_attack": 113,
    "base_defense": 72,
    "base_hp": 975
  },
  "level_32": {
    "level": 32,
    "level_up_xp": 3792,
    "base_attack": 116,
    "base_defense": 74,
    "base_hp": 1000
  },
  "level_33": {
    "level": 33,
    "level_up_xp": 4007,
    "base_attack": 119,
    "base_defense": 76,
    "base_hp": 1025
  },
  "level_34": {
    "level": 34,
    "level_up_xp": 4228,
    "base_attack": 122,
    "base_defense": 78,
    "base_hp": 1050
  },
  "level_35": {
    "level": 35,
    "level_up_xp": 4455,
    "base_attack": 125,
    "base_defense": 80,
    "base_hp": 1075
  },
  "level_36": {
    "level": 36,
    "level_up_xp": 4688,
    "base_attack": 128,
    "base_defense": 82,
    "base_hp": 1100
  },
  "level_37": {
    "level": 37,
    "level_up_xp": 4927,
    "base_attack": 131,
    "base_defense": 84,
    "base_hp": 1125
  },
  "level_38": {
    "level": 38,
    "level_up_xp": 5172,
    "base_attack": 134,
    "base_defense": 86,
    "base_hp": 1150
  },
  "level_39": {
    "level": 39,
    "level_up_xp": 5423,
    "base_attack": 137,
    "base_defense": 88,
    "base_hp": 1175
  },
  "level_40": {
    "level": 40,
    "level_up_xp": 5680,
    "base_attack": 140,
    "base_defense": 90,
    "base_hp": 1200
  },
  "level_41": {
    "level": 41,
    "level_up_xp": 5943,
    "base_attack": 143,
    "base_defense": 92,
    "base_hp": 1225
  },
  "level_42": {
    "level": 42,
    "level_up_xp": 6212,
    "base_attack": 146,
    "base_defense": 94,
    "base_hp": 1250
  },
  "level_43": {
    "level": 43,
    "level_up_xp": 6487,
    "base_attack": 149,
    "base_defense": 96,
    "base_hp": 1275
  },
  "level_44": {
    "level": 44,
    "level_up_xp": 6768,
    "base_attack": 152,
    "base_defense": 98,
    "base_hp": 1300
  },
  "level_45": {
    "level": 45,
    "level_up_xp": 7055,
    "base_attack": 155,
    "base_defense": 100,
    "base_hp": 1325
  },
  "level_46": {
    "level": 46,
    "level_up_xp": 7348,
    "base_attack": 158,
    "base_defense": 102,
    "base_hp": 1350
  },
  "level_47": {
    "level": 47,
    "level_up_xp": 7647,
    "base_attack": 161,
    "base_defense": 104,
    "base_hp": 1375
  },
  "level_48": {
    "level": 48,
    "level_up_xp": 7952,
    "base_attack": 164,
    "base_defense": 106,
    "base_hp": 1400
  },
  "level_49": {
    "level": 49,
    "level_up_xp": 8263,
    "base_attack": 167,
    "base_defense": 108,
    "base_hp": 1425
  },
  "level_50": {
    "level": 50,
    "level_up_xp": 8580,
    "base_attack": 170,
    "base_defense": 110,
    "base_hp": 1450
  },
  "level_51": {
    "level": 51,
    "level_up_xp": 8903,
    "base_attack": 173,
    "base_defense": 112,
    "base_hp": 1475
  },
  "level_52": {
    "level": 52,
    "level_up_xp": 9232,
    "base_attack": 176,
    "base_defense": 114,
    "base_hp": 1500
  },
  "level_53": {
    "level": 53,
    "level_up_xp": 9567,
    "base_attack": 179,
    "base_defense": 116,
    "base_hp": 1525
  },
  "level_54": {
    "level": 54,
    "level_up_xp": 9908,
    "base_attack": 182,
    "base_defense": 118,
    "base_hp": 1550
  },
  "level_55": {
    "level": 55,
    "level_up_xp": 10255,
    "base_attack": 185,
    "base_defense": 120,
    "base_hp": 1575
  },
  "level_56": {
    "level": 56,
    "level_up_xp": 10608,
    "base_attack": 188,
    "base_defense": 122,
    "base_hp": 1600
  },
  "level_57": {
    "level": 57,
    "level_up_xp": 10967,
    "base_attack": 191,
    "base_defense": 124,
    "base_hp": 1625
  },
  "level_58": {
    "level": 58,
    "level_up_xp": 11332,
    "base_attack": 194,
    "base_defense": 126,
    "base_hp": 1650
  },
  "level_59": {
    "level": 59,
    "level_up_xp": 11703,
    "base_attack": 197,
    "base_defense": 128,
    "base_hp": 1675
  },
  "level_60": {
    "level": 60,
    "level_up_xp": 12080,
    "base_attack": 200,
    "base_defense": 130,
    "base_hp": 1700
  }
}

_tool: {
  "version": 1,
  "table": {
    "namespace": "balance",
    "name": "growth"
  },
  "foreignKeys": {},
  "formulas": {
    "level_01": {
      "total_xp": "0",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_02": {
      "total_xp": "balance.growth.level_01.total_xp + balance.growth.level_01.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_03": {
      "total_xp": "balance.growth.level_02.total_xp + balance.growth.level_02.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_04": {
      "total_xp": "balance.growth.level_03.total_xp + balance.growth.level_03.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_05": {
      "total_xp": "balance.growth.level_04.total_xp + balance.growth.level_04.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_06": {
      "total_xp": "balance.growth.level_05.total_xp + balance.growth.level_05.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_07": {
      "total_xp": "balance.growth.level_06.total_xp + balance.growth.level_06.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_08": {
      "total_xp": "balance.growth.level_07.total_xp + balance.growth.level_07.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_09": {
      "total_xp": "balance.growth.level_08.total_xp + balance.growth.level_08.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_10": {
      "total_xp": "balance.growth.level_09.total_xp + balance.growth.level_09.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_11": {
      "total_xp": "balance.growth.level_10.total_xp + balance.growth.level_10.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_12": {
      "total_xp": "balance.growth.level_11.total_xp + balance.growth.level_11.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_13": {
      "total_xp": "balance.growth.level_12.total_xp + balance.growth.level_12.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_14": {
      "total_xp": "balance.growth.level_13.total_xp + balance.growth.level_13.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_15": {
      "total_xp": "balance.growth.level_14.total_xp + balance.growth.level_14.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_16": {
      "total_xp": "balance.growth.level_15.total_xp + balance.growth.level_15.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_17": {
      "total_xp": "balance.growth.level_16.total_xp + balance.growth.level_16.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_18": {
      "total_xp": "balance.growth.level_17.total_xp + balance.growth.level_17.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_19": {
      "total_xp": "balance.growth.level_18.total_xp + balance.growth.level_18.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_20": {
      "total_xp": "balance.growth.level_19.total_xp + balance.growth.level_19.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_21": {
      "total_xp": "balance.growth.level_20.total_xp + balance.growth.level_20.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_22": {
      "total_xp": "balance.growth.level_21.total_xp + balance.growth.level_21.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_23": {
      "total_xp": "balance.growth.level_22.total_xp + balance.growth.level_22.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_24": {
      "total_xp": "balance.growth.level_23.total_xp + balance.growth.level_23.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_25": {
      "total_xp": "balance.growth.level_24.total_xp + balance.growth.level_24.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_26": {
      "total_xp": "balance.growth.level_25.total_xp + balance.growth.level_25.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_27": {
      "total_xp": "balance.growth.level_26.total_xp + balance.growth.level_26.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_28": {
      "total_xp": "balance.growth.level_27.total_xp + balance.growth.level_27.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_29": {
      "total_xp": "balance.growth.level_28.total_xp + balance.growth.level_28.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_30": {
      "total_xp": "balance.growth.level_29.total_xp + balance.growth.level_29.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_31": {
      "total_xp": "balance.growth.level_30.total_xp + balance.growth.level_30.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_32": {
      "total_xp": "balance.growth.level_31.total_xp + balance.growth.level_31.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_33": {
      "total_xp": "balance.growth.level_32.total_xp + balance.growth.level_32.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_34": {
      "total_xp": "balance.growth.level_33.total_xp + balance.growth.level_33.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_35": {
      "total_xp": "balance.growth.level_34.total_xp + balance.growth.level_34.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_36": {
      "total_xp": "balance.growth.level_35.total_xp + balance.growth.level_35.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_37": {
      "total_xp": "balance.growth.level_36.total_xp + balance.growth.level_36.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_38": {
      "total_xp": "balance.growth.level_37.total_xp + balance.growth.level_37.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_39": {
      "total_xp": "balance.growth.level_38.total_xp + balance.growth.level_38.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_40": {
      "total_xp": "balance.growth.level_39.total_xp + balance.growth.level_39.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_41": {
      "total_xp": "balance.growth.level_40.total_xp + balance.growth.level_40.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_42": {
      "total_xp": "balance.growth.level_41.total_xp + balance.growth.level_41.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_43": {
      "total_xp": "balance.growth.level_42.total_xp + balance.growth.level_42.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_44": {
      "total_xp": "balance.growth.level_43.total_xp + balance.growth.level_43.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_45": {
      "total_xp": "balance.growth.level_44.total_xp + balance.growth.level_44.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_46": {
      "total_xp": "balance.growth.level_45.total_xp + balance.growth.level_45.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_47": {
      "total_xp": "balance.growth.level_46.total_xp + balance.growth.level_46.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_48": {
      "total_xp": "balance.growth.level_47.total_xp + balance.growth.level_47.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_49": {
      "total_xp": "balance.growth.level_48.total_xp + balance.growth.level_48.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_50": {
      "total_xp": "balance.growth.level_49.total_xp + balance.growth.level_49.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_51": {
      "total_xp": "balance.growth.level_50.total_xp + balance.growth.level_50.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_52": {
      "total_xp": "balance.growth.level_51.total_xp + balance.growth.level_51.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_53": {
      "total_xp": "balance.growth.level_52.total_xp + balance.growth.level_52.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_54": {
      "total_xp": "balance.growth.level_53.total_xp + balance.growth.level_53.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_55": {
      "total_xp": "balance.growth.level_54.total_xp + balance.growth.level_54.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_56": {
      "total_xp": "balance.growth.level_55.total_xp + balance.growth.level_55.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_57": {
      "total_xp": "balance.growth.level_56.total_xp + balance.growth.level_56.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_58": {
      "total_xp": "balance.growth.level_57.total_xp + balance.growth.level_57.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_59": {
      "total_xp": "balance.growth.level_58.total_xp + balance.growth.level_58.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    },
    "level_60": {
      "total_xp": "balance.growth.level_59.total_xp + balance.growth.level_59.level_up_xp",
      "power_rating": "base_attack * 3 + base_defense * 2 + floor(base_hp / 10)"
    }
  }
}

_editor: {
  "table": {
    "title": "等级成长：数值与依赖"
  },
  "columns": {
    "level": {
      "title": "等级"
    },
    "level_up_xp": {
      "title": "升级经验"
    },
    "total_xp": {
      "title": "累计经验"
    },
    "base_attack": {
      "title": "基础攻击"
    },
    "base_defense": {
      "title": "基础防御"
    },
    "base_hp": {
      "title": "基础生命"
    },
    "power_rating": {
      "title": "参考战力"
    }
  },
  "cellComments": {
    "level_01": {
      "total_xp": "达到1级无需经验，因此起点是0；升至2级累计103。",
      "power_rating": "参考战力：23×3 + 12×2 + floor(225/10) = 115。"
    },
    "level_02": {
      "total_xp": "明确引用同表上一行，累计经验为103。"
    },
    "level_60": {
      "level": "最后一级保留升级经验供曲线观察，客户端满级后不得继续升级。"
    }
  }
}

_export: {
  "csv": {
    "enabled": true,
    "columns": ["id", "level", "level_up_xp", "total_xp", "base_attack", "base_defense", "base_hp", "power_rating"],
    "encoding": {}
  }
}
