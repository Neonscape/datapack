# Character Growth Manual

此工程展示本表跨行公式。自动或手动编译由用户级编辑器设置决定；手动模式下，保存修改后底部状态栏显示上一次编译结果已过期，用户执行编译时工具对整个工程调用 `CompileAll`。

`combat.growth` 的 `attack` 使用本表跨行公式。引用保存完整表名：

```text
combat.growth.level_01.attack
```

第一版仍要求公式中的表名等于当前表，因此这些不是跨表公式。
