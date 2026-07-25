# Schema Composition

此工程展示字段的类型 Schema 和约束 Schema 组合。

```text
price
  类型 Schema：rules.#Price

stackLimit
  内置类型：int
  本地约束 Schema：_local.#WeaponCount
```

`rules.#Price` 是类型 Schema，提供完整整数类型和其内置限制。`rules.#NonNegative`、`rules.#AtLeastOne`、`rules.#NonEmpty` 是约束 Schema，只能附加到适用类型上。
