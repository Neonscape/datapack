# 结构编辑响应与公式编辑所有权

Add/Delete/RenameRow 与 Add/Delete/Rename/MoveColumn 的成功响应，现在与 GetTable 一样，在公有返回边界对原始 document.Table 执行一次 withTableMetadata。此前只返回源 AST 模型，公式保存在 Formulas 而非 Row.Values，替换前端表格后会同时丢失显示值与编辑表达式；effective enum、FK options 和诊断也缺失。事务/历史/源写入顺序不变，不在通用 edit helper 盲目再次水合，已有装饰路径不变。

公式栏以原行 ID、列 ID 和是否草稿捕获编辑起点。只将实际输入视为保存；外部刷新填入此前空白的预览不会因失焦触发清空。原行/列消失时不将旧输入写到同坐标的新对象。显式清空仍走原 ClearCell；没有增加保存队列或改变源公式语义。

回归包含 manual/auto 的9种结构变体，比较全部 Table DTO 与 fresh GetTable，检查幸存公式表达式/显示/flag、effective enum、FK选项及旧manual诊断，Undo/Redo和预览零源写入。浏览器 fixture 由实际 Service 经 JSON 生成，覆盖删首/中/末行、语言切换/历史和无输入刷新误清。测试导出器仅在 Playwright 指定的私有临时文件下运行，正常Go测试跳过导出；没有新增运行时API或依赖。

成功响应夹具给公式目标保留合法默认值，以到达已有 Add/MoveColumn 写前校验的成功分支。无默认值公式目标被这两类写前校验拒绝的问题另记，不将本次响应修复宣称为覆盖该独立边界。
