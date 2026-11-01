# 代码质量与职责边界

2026-11-01 的维护覆盖后端、前端、桥接、测试、文档与构建。以下记录完成的有限边界及保留理由，不以文件行数或重写数量衡量质量。源文件仍是权威，内部 ID、文件格式、CUE 语义、默认值时点、导出筛选和历史协议保持兼容；已确认的正确性问题以独立纠错提交处理。

|模块|本轮处理 / 保留理由|验证与边界|
|---|---|---|
|workspace 源命令与历史|统一成功刷新后记历史，快照包含未入树目标；读取错误不伪装不存在，失败保留状态并报告补偿错误。|故障注入覆盖第二文件提交/恢复失败、特殊 mode、目录/附件、history/revision/compile；拒绝 symlink/非普通源文件。见 source-transactions.md。|
|document / sourcefile|共享裸字节 stage/restore 机制，AST/模型校验仍由领域命令持有。|非法语法不落盘，合法业务错误仍能保存并诊断。只有原字节完整暂存后最终 rename 失败才保证保留恢复材料；不宣称 crash、外部并发、ACL/xattr/inode 事务。|
|Schema 改名|使用 CUE 词法绑定、import/package 身份，拒绝名称捕获等不安全迁移。|公共/本地/嵌套/同名/alias/comprehension 编译前后、无关源零写入与历史回归。见 schema-renames.md。|
|Schema import 身份|内部解析上下文统一索引/反向使用/导航/有效类型，不改原表达式或 DTO。|真实 module 默认 major，implicit/explicit v1、qualifier、错误路径/版本、循环、修复坏 module；不硬编码省略 major 为 v0。见 schema-imports.md。|
|有效 Schema 模型|递归展开前复制 Fields slice，避免改写共享 document/registry。|源引用不变性及并行展开 Go race 回归；是局部所有权修复，不是通用深拷贝框架。见 schema-preview-isolation.md。|
|workspace 生命周期|viewLifecycle 拥有 revision/上次编译/纯发布，compile_pipeline 负责源收集、索引、完整编译及 view 构造；Service 保留锁、源命令和历史。|先特征测试：manual 旧诊断/时间与 pending、auto 编辑、失败刷新不发布、重开后确实重编译外部变化；无新计时器或缓存。见 workspace-lifecycle.md。|
|preview / compiler|共享 cuerows 的纯 CUE 主键/字段/整行检查，保留调用方 package loader、诊断策略和 final Decode。|默认整数/空数组/null 与源预览不水合、诊断次数/位置/key、多 package 选择差异先锁定再提取；manual/pending 未统一。见 cue-row-validation.md。|
|formula / foreignkey / registry 领域 / exporter / export_path / cueformat|领域边界与安全回归有效，没有重写依据；只改实证的局部兼容问题。|维持最终值、FK 来源、compiled 消费、危险路径和格式兼容保护；不因“质量维护”重新设计成熟模块。|
|cuesource / document|宽容树扫描与严格编辑服务不同场景。|保持分离，损坏文件仍可进树并诊断，严格编辑不会以容错解析掩盖非法源。|
|App / session|session 原子拥有 workspace/tabs/active/peek，封装初始化、tab 身份转换和 command/result 应用，减少裸 setter。|project generation、request sequence、tab 实例、focus intent 与 publication counter 保护 open/history；延迟 Promise 真浏览器回归。并未为所有持久化命令增加通用并发队列。见 workspace-session.md。|
|Wails bridge / types|单一 adapter 处理生成 DTO 与 UI 类型；组件不散落生成绑定/断言，生成文件未手改。|raw JSON/null、false/0/map、diagnostics/timestamp、输入参数及 native rejection identity 契约。取消选择的 zero view 保留现有归一化。见 bridge-contract.md。|
|TableGrid / valueCodecs|保留既有 codecs；useRowDraft 拥有未存行，纯 selection 负责坐标转换，eventOwnership 明确活动焦点。|是适度提取，复杂 array/struct editor UI 仍在 Grid；全量复制粘贴/F12/草稿/选区/数组输入回归。见 grid-editing.md。|
|在途草稿|单个 CreateRow 在途时继续保本地输入，禁止重复提交；迟到结果只清原样草稿。|成功/失败/同 ID 新字段的真实延迟回归；已保存 ID 的新草稿提示重复并保输入，显式编辑才重试，不自动保存或静默更新旧行。|
|projecttree / navigation / display names|滚动受左侧容器限制；模态滚轮/焦点/快捷键隔离；首次只展开默认文档祖先，刷新保手工状态；显示名不参与身份。|长树/小窗口/长文案截图、默认与无默认工程、语言/项目切换、同名、显示名保存/Undo/草稿、同 package Schema 标题兼容。|
|i18n / theme|固定树/Schema 文案进入现有中英 catalog，检查键集合一致性。|显示内容与稳定 key/ID 分离；没有风格性重写 catalog 或主题系统。|
|settings / logging|manual 仍校验语言/主题；日志 writer/注册快照锁与外部 callback 分离。|legacy 默认和零 manual debounce 保留，无效保存写前拒绝。**并发 Log 的 Sink 可并发、无全局顺序**，这是一项明确的合同变化；Sink 要线程安全、不阻塞，无新增队列。生产唯一 Sink 是无局部状态的 Wails 事件转发。见 runtime-boundaries.md。|
|build / CI / scripts|前端构建在 Go embed 前，可移植浏览器配置；打包用临时 staging、无覆盖发布和精确版本附件白名单。|干净构建、链接/旧产物/中途失败保护、shell syntax 和 mock release contract；同版本不同字节拒绝，应使用新版本。本轮没有实际 release/deploy。|
|测试 / 文档 / 样例|保留 Go 功能/golden 和 mock IPC 的真实 Chromium 两类边界，产品及测试源码均严格类型检查；更新过时 API/职责说明。|云岚数据/golden不变；没有删除断言来通过检查。原生 smoke 实测窗口、设置 IPC 和打开工程对话框取消，不等于桌面编辑/系统剪贴板/保存对话框全测。|

## 测量与验证

`internal/workspace/quality_benchmark_test.go::BenchmarkYunlan` 使用 `docs/examples/yunlan-rpg/yunlan-rpg.dtproj`，不改源文件。复跑：

```sh
go test ./internal/workspace -run '^$' -bench '^BenchmarkYunlan$' -benchtime=3x -count=3 -benchmem
```

本机 Linux / Go 1.26.5，维护前后各 3 次 × 3 迭代：基线 scan 约4.9ms、parse 11.5–12.5ms、index 0.043–0.062ms、compile 191–198ms；组合版约4.2–4.8ms、10.6–11.1ms、0.049–0.058ms、189–192ms。这是小样本本机观测，不能推断性能提升。完整 CUE 编译主导，本轮不引入跨命令缓存、增量编译或 debounce 改变诊断时机。

验证包括 Go 全量/race/vet、产品及全部测试源码 strict TS、Vite 与 Linux Wails 构建，以及串行真实 Chromium 回归。组合版本110项中96项实际使用浏览器、14项为纯catalog/DTO/坐标契约；UI仍mock Wails IPC，真实持久化/编译/导出由Go测试覆盖。额外使用隔离 Xvfb/临时 XDG 设置运行原生窗口、GetSettings 和文件选择取消 smoke；不读取或改写真实工程/用户配置。

## 后续维护边界

本轮已识别的实质职责项已处理；保留上述验证与复杂安全模块，不继续无边界审计。更完整的持久化命令迟到结果协议、大型 editor 受控组件拆分、真实桌面编辑与系统剪贴板自动化，可作为后续独立任务。缓存和调度变化需测量与兼容设计，不能以本轮行为保持重构为名加入新语义。
