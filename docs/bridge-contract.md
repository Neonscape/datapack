# Wails 与 UI 的契约边界

`frontend/src/bridge` 是组件访问生成绑定与模型的单一入口。生成文件由 Wails 维护，不手改。组件使用自己的 UI 类型，DTO 的 kind/mode、集合和空结果转换集中在 bridge；请求参数仍由生成绑定签名检查。

有效 DTO 保留身份、标题、源表达式、默认值、false/0、导出选择与诊断时点。取消选择项目的零 View 保留空 path sentinel；Schema 命令未使用的 table/schema 半部允许为空；nil 集合归一化为空集合。未知 kind/mode 在边界报错。

工作区和表带 projectEpoch/revision。`UpdateCellDelta`、`SetCellTextDelta` 仅在独立字面量事务已写盘并提交历史后返回单元格 delta；复杂表保留完整 hydrated DTO。`UndoDelta`/`RedoDelta` 对安全单元格历史返回 delta，其他历史由客户端读取完整权威文档。旧完整 API 保留；旧测试桥可继续使用原响应。

`ExportResult` 包含本次成功导出对应的权威 workspace，客户端安装其编译状态并刷新必要文档。目录、选择和 targets 都在编译刷新后的同一配置上解析；导出提交前仍核验完整输入。`GetOperationState` 在独立锁下返回真实阶段和计数，供重操作期间轮询。

写调用按顺序派发；只读文档调用可并发，并在会话安装时做 generation/epoch/revision fencing。契约映射不修改冻结输入。真实 App 组合测试覆盖队列、发布与刷新，不把纯队列/纯 patch 单测称作完整交互验证。
