# Workspace 发布职责

Service 持有工程互斥锁、当前 view、不可变字面量快照与历史栈。`viewLifecycle` 只拥有 source revision、上次编译 inputRevision 和编译状态，不读文件或安排调度。每次成功持久化后发布新 revision；manual 可以同时携带旧诊断/时间戳与 pending。

新默认模式为 manual，保留用户已有的显式 auto 设置。manual 首次打开清空历史并使用新 projectEpoch，标记尚未校验、inputRevision=-1；auto 打开和成功编辑执行完整校验。Ctrl+Shift+B 触发明确编译，输入框/组合输入保留快捷键 ownership。编译和导出验证当前输入，失败不会伪装为干净成功。

`compile_pipeline` 共享文档读取，完整输入指纹包括同包祖先、模块/配置和文件集合；导入依赖发现后仍执行相同外部输入限制并检查各依赖包。可证明完整的无导入输入才复用成功结果；导入回退仍做编译前后及导出提交前一致性检查。导出目录在编译可能重载 .dtproj 之后解析，避免混合新配置与旧输出目录。

快路径保持 Write/Sync/Close、同文件系统 rename 和历史提交先于成功响应；普通单元格历史使用可逆补丁，不截断 Undo。回滚失败时发布 sourceUncertain，阻止旧快照继续修改或导出，只有成功重新打开才能恢复权威状态。文件同步不等于多文件或掉电事务保证。

重操作状态使用独立锁，报告实际阶段和行数。前端锁住冲突操作并保留到必要文档刷新完成；服务也在获得工程锁前后检查冲突。原始测量、计时边界、未测场景及回归映射见 [100k 产品化报告](../research/100k-product/REPORT_zh.md)。
