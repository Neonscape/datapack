# Wails 与 UI 的契约边界

`frontend/src/bridge` 是组件访问生成绑定与模型的单一入口。生成文件由 Wails 维护，不手改。组件保留自己的 UI 类型，DTO 到 UI 的 enum/集合/空结果转换集中在 `bridge/models`，请求参数仍由生成绑定签名检查；需要构造生成模型的输入也留在 bridge 中。

有效 DTO 的身份、显示名、源码表达式、默认值、false/0、导出选择、manual/pending/诊断时点保持不变。取消选择项目的零 View 保留空 path sentinel；Schema 命令没有使用的 table/schema 半部可为空；空集合进入 UI 时归一化，不能因为 Go nil 列表而崩溃。

已知 tagged kind/mode 通过明确检查收窄，未知值在边界报错，不再用类型断言让错误协议进入组件状态。这只影响不受当前生成协议支持的响应，不改变有效项目数据语义。没有增删服务 API、修改生成模型或新增运行时 schema 框架。

契约测试以真实生成构造器创建 DTO，冻结输入验证映射无写入，检查源表达式/ID/标题/历史/手动编译状态、取消/零值、导出选择和未知 tag；另检查 PrintSchema 请求的生成模型转换保持词法输入。完整 UI 测试仍执行 Chrome 交互，契约测试本身是纯 Node 检查，不能称作原生 Wails smoke。
