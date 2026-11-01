# Code Quality And Maintenance

本轮评估覆盖全项目。保留源文件权威、受限 CUE、显式编译结果和稳定内部 ID；重构不能改变默认值、诊断时点、允许保存业务错误、导出筛选和历史语义。以下按源码证据和回归区分已处理、无需变更与后续工作，不能把待验证风险称为缺陷。

| 模块 | 当前证据 / 处理 | 验证与后续优先级 |
|---|---|---|
| compiler / formula / foreignkey / schemaregistry | 领域职责明确：最终值、依赖状态、FK 来源与导航索引各有边界。显示名仅随 DTO 传播，不参与 ID。 | 维持包/functional/golden/race。领域流程没有重写依据；compiler/preview 仅共享无 I/O 的 cuerows 校验机制，package选择、诊断和默认解码保留各自策略。详见 cue-row-validation.md。 |
| cuesource / document / schemamodel / cueformat | 宽容树解析与严格编辑有不同用途，不能合并；CUE 格式兼容层已有测试。 | 保留边界。裸字节 staging 提到 sourcefile，语法/模型校验仍由命令负责。 |
| sourcefile | 统一临时文件创建、权限、写入、同步、关闭及恢复副本；document 三类 writer、多文件替换和 workspace 历史写入复用。 | 测试 create/chmod/write/sync/close 故障，原字节与 mode、空文件、无效 CUE 的原样 staging。机制层不做语义校验。 |
| workspace | 已有 Service RWMutex、历史补偿、目录快照及 FK 来源保护。旧文档“没有锁”的结论过时。新事务边界在成功刷新后发布历史，读取失败不会伪装成不存在。 | 新增刷新失败的行/Schema 命令、目录、附件、历史/revision/compile 状态回归。仍需 P1 会话生命周期与发布协议特征测试，P2 测量重复扫描后再考虑缓存。 |
| projecttree / navigation | 固定分组与标题由 UI 呈现；内部名、路径和节点 kind 保持身份。 | 导航回归覆盖大树/小窗口、初始默认祖先、刷新保留手动展开、切换项目和语言。无需领域导航框架。 |
| exporters / export_path | 导出只消费 compiled 包，危险路径保护与 CSV 编码已有回归。 | 维持现有安全边界，无重构依据。显示名修改前后 CSV 字节一致。 |
| App / workspace session | 已实证三项迟到读取/刷新丢标签问题并独立纠错。session 原子维护 view/tabs/active/peek，显式命令替代 React setter；项目 generation、请求序号、标签实例和焦点意图约束 open/history 的结果应用。 | 纠错前真实浏览器3项准确失败，修复后11项延迟 Promise 回归通过，含删除/改名请求失效及peek/命令交错。最终全量94项通过；其余编辑命令的异步结果协议仍需后续验证与适配。 |
| TableGrid / valueCodecs | 保留已有codecs。行草稿工作流归useRowDraft，输入字段身份/创建回调解耦坐标和IPC；选区边界是独立纯规则，clipboard/F12共用事件归属判断。 | 保留全部真实浏览器输入/失败/焦点/数组回归，新增纯边界契约。专用数组/结构编辑器继续由Grid协调，本批不重写成熟输入或引入框架。 |
| types / Wails bridge | DTO/UI模型保留各自职责；bridge为组件唯一生成绑定入口，集中tag/集合/零结果转换和请求模型构造。App/Settings/SchemaEditor的DTO断言与generated imports已移除。 | 生成模型契约验证身份/源码/false/0/manual/pending、取消/空半部/非法tag及输入不变；有效输入维持原行为。生成文件未改，真实Wails交互仍另列缺口。 |
| i18n / theme / logging / settings | Catalog 与主题边界清楚。固定导航文案已入中英资源；logger 调外部 sink 时持锁、settings manual 分支早返回为待验证风险。 | 不为风格重写 catalog。低优先级复现 sink 重入/错误校验边界后决定修复，未将静态疑点当确认缺陷。 |
| tests / examples | Go 包、functional/golden 与真实 Chromium UI 各有价值；UI 只 mock Wails IPC，不证明原生窗口/序列化。云岚保留业务链与英文 ID/中文显示名。 | 产品和测试源码都做严格 TS 检查；浏览器串行并检查布局截图。真实桌面 smoke 尚未执行。 |
| build / CI / scripts | 已复现并修复干净 checkout 缺 embed 资源、打包清空输出目录及预存链接/附件范围风险。PR14已正常合并且精确/post CI成功。 | 前端 build 前置、可移植浏览器、无覆盖 staging 发布和当前版本附件白名单均有安全回归；未实际 release/deploy。 |
| docs | 原质量文档已与锁、session、codecs 和损坏文档诊断实现不符。 | 本文纠正现状；test/README 的 API 边界说明也已更新。 |

## 独立的纠错与兼容边界

文件改写成功但视图刷新失败时，必须恢复源与权限，保留历史双栈、revision 和 compile 状态。补偿失败必须报告；只有原字节已成功暂存、最后 rename 失败时保证保留该副本，暂存阶段失败不保证存在恢复文件。这里是内存历史与已报告 I/O 失败的补偿，不能宣称崩溃原子性或阻止外部并发修改。

校验所有权仍属于领域命令。非法 CUE 不落盘；合法语法但违反业务约束的数据仍可保存并产生诊断。sourcefile 只能原样写字节，不得悄悄升级为“必须编译全项目成功才保存”。

Schema 改名的旧遍历器会按同名 `#Name` 修改无关作用域，已在真实可编译的公共/本地同名定义下复现 panic。独立纠错批次改用 CUE 词法绑定与模块/import 身份识别，覆盖名称捕获拒绝、无关源零写入和编译/导出/历史回归；这属于正确性修复，不算纯 writer 重构。范围与保守拒绝条件见 `schema-renames.md`。独立 alias 纠错使用私有词法/import 身份上下文，统一索引、循环、反向引用、有效类型与导航；原词法表达式/DTO不变，参见 `schema-imports.md`。

## 持续维护次序

1. 先完成已复现的数据/作用域与会话竞态纠错，每项保留旧代码失败证据。
2. 用特征测试锁定会话初始化、命令成功/失败发布及标签身份，再减少裸 setter 和重复刷新协议。
3. 测量云岚样例 scan/load/index/compile 时间，再决定性能维护；不先引入缓存/增量编译。
4. 整理桥接契约与编辑器状态归属。成熟小模块没有实质问题则明确不改。

每批提交可单独审阅。Go 全测、race、vet、产品/测试类型、前端构建、真实浏览器 E2E 与适用的原生构建通过后，仍须独立复核、精确 CI 和正常合并流程。不安装依赖、不修改同步/安全设置、不部署发布来完成本轮维护。
