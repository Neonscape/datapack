# 100k 单表优化产品化候选（2026-11-02）

本轮候选基于 alpha `5f8156f829f77dda2cd1636b76af4c79bec62c8d`，保留该版本的结构操作 hydration、enum 类型与重复提交修复。工作分支 `candidate/100k-local-20261102`；精确最终 head、补丁及校验和见交付包 `candidate-metadata.json`。这是待独立审查的候选，不表示已合入、发布或部署。

> 测量版本说明：下列 100k 宏测与原生 PNG 的实测代码为 `eaf172bf91faaa8d054a8a44bf80f83c111b515e`。独立审查后增加了 `df5fd63` 的后端配置/完整闭包修正和 `927a01b` 的前端顺序/freshness 修正；未重跑耗时基准。新候选已通过必要的正确性回归，原始日志、截图像素与计时记录保持不变。5.75 秒等导出数字仍是原测量版本的 Service+JSON，不包含新增的必要 UI 文档刷新成本。

## 本轮结论

在严格单表 100000 行、10 列的既有 fixture 上，手动模式保持即时持久化，并达到已测原生编辑的 ≤2 秒目标。Linux Wails/Xvfb 的三次首行数值编辑，磁盘可观察到更新为 0.130–0.132 秒；在 1.165–1.178 秒完成的整屏截图中，画布已显示权威新值。这是固定等待后截图得到的**上界**，不是精确 paint latency。撤销与重做各一轮也在约 1.17 秒的整屏截图中显示正确结果；末行字符串及中部布尔操作另有记录，详见 native 摘要。权威提交、原子替换及历史提交先于 RPC 成功响应；最终画布显示采用真实 Wails 路径，不以输入框文字或 spinner 作为完成。

**该结论仅适用于本次手动、独立字面量样例。自动模式编辑仍约 12.55 秒，未达到 2 秒。** 大规模公式、FK、导入或结构操作继续采用保守慢路径。未证明所有 100k 工程都能达到目标。Windows 完成交叉编译，未执行 Windows 原生体验验收。

## 四部分实现与设计

1. **消除重复工作。** 一个操作内共享解析后的文档供索引使用；没有公式时不做全表公式求值；无 FK 时不搜集全工程候选；公式显示用 row ID 索引。安全的标准标量才走快解码，CUE 表达式、精度边界及复杂语法保留 CUE fallback。
2. **即时保存的精确补丁与历史。** 仅对已证明相互独立的字面量表建立不可变快照、行索引和源字节范围。编辑用精确源补丁，保留无关文本与注释，文件 staging→Write/Sync/Close→同文件系统 rename 成功后才提交权威快照、revision 与历史。单元格历史为可逆值补丁及摘要，不静默截断 Undo。一般结构操作仍走完整事务。文件 Sync 不宣称目录同步或完整断电事务保证。
3. **真实依赖闭包与编译复用。** 指纹涵盖配置、表/Schema 根、同包祖先和兄弟 .cue、模块元数据及文件集合变化。直接、传递 import 通过真实 CUE 加载发现依赖，并校验依赖包自身，避免被根投影隐藏的 bottom。导入路径不复用成功编译结果，但仍做编译前后输入检查与导出提交前检查。无导入的完整可证明输入才复用成功结果；祖先约束使 SchemaCapsule 回退后也受完整指纹约束。无法建立闭包或含 @embed/@extern 时，当前候选明确拒绝 compile/export，而非导出无法核验的结果；这是需要审查的兼容性边界。
4. **产品生命周期与小响应。** 新默认 manual，尊重已有显式用户级 auto 设置。首次 manual 打开不做完整编译，标明未校验；编辑即时保存，旧结果带 inputRevision/pending，编译与导出校验当前输入。Ctrl+Shift+B 支持明确编译，跳过输入框、组合输入与重复按键。重操作提供真实阶段及实际行计数并锁住冲突操作，失败解锁。写 RPC 串行派发；表和工作区带 projectEpoch/revision。安全字面量编辑及其 Undo/Redo 返回单格 delta，版本或身份不匹配则读取权威完整表；其他操作保留 alpha 已补齐的完整 DTO。可见画布显式刷新；活动网格使用捕获阶段的剪贴板监听，继续保留原生输入框与非活动标签的 ownership。

完整表的初始 DTO 仍约 70.24 MB。增量编辑将本次实测响应降至约 1.6 KB。初始打开仍需完整解析、发送表，不是分页或部分 CUE 编译。多线程不是本轮收益来源：主要瓶颈来自重复解析、无效求值和大响应；直接并发共享 CUE context 没有安全性证据。

## 同环境实际结果

CPU、内存、Go 版本及保护参数见 `hardware.json`。Go 1.26.5；GOMAXPROCS=2，GOMEMLIMIT=2048MiB，宏测进程树 RSS 保护 4608 MiB。未清 OS 页缓存。“修改后编译”是应用输入缓存 miss，并非 OS 冷启动。部分正确性与原生检查和基线宏测有重叠；未做 CPU 绑核或系统负载隔离，数字不代表用户设备 30/7/12 秒的对应耗时。

下表为真实 Service+JSON，除注明外各 3 次；打开值不含初次 GetTable/IPC/画布。

| 操作 | alpha 基线中位数 / 最慢 (s) | 候选中位数 / 最慢 (s) | 计时边界 |
|---|---:|---:|---|
| manual OpenProject | 34.463 / 34.468 | 1.474 / 1.499 | 基线强制完整校验；候选标记尚未校验 |
| GetTable | 16.063 / 16.084 | 0.235 / 2.243 | 候选第一次 2.243s，其余为同 revision 快照；仍完整 DTO |
| manual 编辑，完整 DTO | 19.229 / 19.546 | 0.343 / 0.347 | 包含即时写盘、历史和 JSON，不含最终 UI |
| 修改后完整编译 | 33.233 / 34.488 | 5.630 / 5.659 | 候选另跑 3 次修改后的 cache miss；非 OS 冷缓存 |
| 修改后导出 | 34.622 / 34.966 | 5.748 / 5.814 | 候选另跑 3 次修改后导出，包含完整校验 |
| 未变化导出 | 34.622 / 34.966 | 0.236 / 0.248 | 复用相同完整输入的成功编译结果，提交前再次核验 |
| SaveProjectAs | 34.028 / 34.280 | 1.461 / 1.481 | 实际复制、目标打开；候选 manual 目标仍未完整校验 |
| auto 编辑，完整 DTO | 74.011 / 74.176 | 12.552 / 12.669 | 未达到 2 秒，保留即时完整校验 |

增量 API 的独立 12 次首、中、末行混合 string/number/bool/enum 测量：edit 中位数 0.147s、最慢 0.159s；Undo 0.134/0.138s；Redo 0.133/0.139s。该序列是 Service+JSON，**不是 12 次原生 UI 撤销**。响应最大约 1.7 KB，保留 12 条历史，无整表源快照，表达式 262 字节加摘要 1536 字节，不含 Go 对象开销。

相同 full 宏工作流（3 Open、3 Get、3 manual edit、3 Undo+Get、3 Redo+Get、3 Compile、3 Export、切 auto、3 auto edit）基线总墙钟 747.809s、候选 58.159s。此工作流包含新的 freshness 与正确缓存语义，不应将其全部差值解读为同一项编译算法加速。另有修改后 cache miss 测试，证明没有仅把编译等待转移后冒充全面优化。宏测采样进程树峰值 RSS 基线约 2064 MiB、候选约 1895 MiB；增量序列约 1505 MiB。原生进程仅记录点时 RSS，**未测峰值**。

## 正确性与负面测试映射

- 祖先 .cue 修改/新增/删除/改名：`TestAncestorConstraintChangesCannotReuseSuccess`。旧原型 4 项失败日志仍保留，新实现通过。新兄弟文件、第二张表、导入文件集合：`TestNewSiblingAndImportedFileSetChangesInvalidate`。
- 直接/传递 import 与模块变化：`TestTransitiveImportAndModuleChangesInvalidate`。新增导入文件曾实测失败；通过依赖包自身校验修复，现为通过，不把失败归为缓存命中。
- 编译中外部修改：`TestCompileInputChangeRejectedWithAndWithoutCache`；缓存命中后的修改：`TestExportChangeAfterCacheHitCannotPublish`；提交前修改含导入回退：`TestExportCommitFenceWorksForImportFallback`。拒绝时保留上一份有效导出。
- Write/Sync/Close：`TestStageFailuresLeaveOriginalAndNoTemporarySource`；rename 失败：既有 literal failed-writes 测试；提交后刷新失败及回滚失败：`TestLiteralRefreshFailureRollsBackWithoutPublishing`、`TestLiteralRollbackFailureLocksAuthorityUntilSuccessfulReopen`；Undo/Redo 刷新失败：`TestHistoryRefreshFailureDoesNotConsumeUndoOrRedo`。回滚失败明确 sourceUncertain，阻止旧快照继续写，失败重新打开不能解锁。
- manual 首次打开、成功后非法值、失败后修复但未重编译：`TestManualOpenAndStaleDiagnosticsHaveExplicitRevision`；冲突锁与错误解锁：`TestOperationProgressReadableAndConflictingEditsRejected`。
- delta 与实际保存的完整文档一致、text 转义、公式回退：`TestCommittedCellDeltaAndHistoryMatchFullDocument`；前端 `operation-order.spec.ts` 检查串行派发、失败重试、跨工程/乱序/旧 revision 拒绝及真实进度锁。
- alpha enum 删除行、公式结构操作、值/公式切换、Schema/列 DTO、CUE 默认值/边界数值、FK、注释、重命名与粘贴：既有 Go/functional 和前端 suite 全量回归。`structural-responses.spec.ts` 仍从真实 Service 生成 DTO；历史采样的 width fixture 在模拟新写请求时显式重标请求 revision，避免将真实旧版本伪装成新响应。
- 最终验证：Go `go test -p 1 ./...`、内部与 functional `-race`、`go vet -p 1 ./...`、前端构建及测试类型检查、**155 项前端测试通过**。Linux Wails 原生构建与实际窗口操作、Windows amd64 交叉构建成功；Windows 原生未执行。
- 四份真实 Service CSV 各独立解析 100000 行、10 列、100000 唯一 ID，逐字段与固定 seed 原始数据叠加实际编辑记录比对，共 400 万字段；包含逗号、引号、换行、制表符、首中末值。六份 SaveAs 源文件与原始源 hash 一致。见 `verify_exports.py` 与 `export-verification.json`。

## 仍需独立审查的边界与远程状态

进程 mutex 不是文件系统事务。当前读前/读后及导出提交前的内容核验可拒绝确定性注入的外部变化；无法与不合作的外部程序原子锁住所有输入及整个输出目录。未测试恶意在最终核验之后、rename 之间反复切换文件，或字节 ABA 改回原值；也未宣称掉电恢复或多文件原子提交。快路径在外部内容改变后拒绝旧快照写入；用户需要重新加载。复杂表、大规模跨表 FK 和公式的 100k 性能仍未量化；语义回归采用小工程。

本地所有变更在隔离 worktree，未碰原有 12 项用户修改。用户已授权在测试与独立审查后接入 canonical Gitea、再核验 GitHub 镜像；未授权发新版。只读查询 canonical master 返回 `could not read Username for https://dev.projectmiraland.top:50005`，现有运行环境没有可用认证。没有读取或修改凭据、权限、同步或 runner 设置。远程推送/PR/合并/镜像 CI 状态以 `candidate-metadata.json` 的实际结果为准；本报告不将本地完成描述为远程接入完成。

建议采用即时保存 + 默认 manual + 严格 freshness + revisioned delta 的路线，先独立审查完整依赖和失败恢复，再通过现有授权认证接入。保留已有显式 auto 用户选择，但应清楚提示其 100k 编辑仍有完整编译成本。导出始终校验当前数据，不依赖过期的“成功”标记。

## 独立审查修正（2026-11-02）

- `TestExportReloadsExternalDestinationBeforePublishing` 在旧候选上复现 .dtproj 导出目录 A→B 后写旧 A。现在先完成编译/外部刷新，再从同一权威项目配置解析和安全验证目录。A 保持不变、B 得到新数据；B 非法或指向受保护源目录时拒绝，不能退回 A。
- `TestTransitiveInputsRejectUnsupportedExternalAttributes` 在旧候选上证明新发现的传递依赖 @embed/@extern 未拒绝。当前对发现文件、实际加载字节及闭包稳定后的全集执行相同限制，在依赖包求值前递归检查。
- `review-composed-lifecycle.spec.ts` 用真实 App/session/ordered 组合复现延迟 edit 后已提交 Undo 的结果被 enqueue-time publication token 丢弃，以及 Redo 失败被静默隐藏。现在用工程/请求归属与权威 epoch/revision 协调；delta 和完整读取均覆盖，旧工程及更新 revision 仍受保护。
- 成功 Export 返回权威 workspace。App 安装当前校验状态，必要时读取公式/变更文档；同 revision 无公式表只清理过期诊断，避免重传 100k 行。首次 manual 打开→导出、旧错误修复→导出、排队 edit→导出均有组合回归，未增加重复 Compile。
- 目标回归、Go/race/vet、构建与类型检查通过；最终完整前端 155 项通过。一次全量中原有画布双击夹具尚未建立编辑器，发生在粘贴前；独立 7 项通过，并补充先确认单元格选中状态的步骤后取得完整 155 项通过。原失败日志保留，不将其删去或当作通过。
- 原生 PNG 已作为独立像素证据包交付，含逐操作 JSON 和 SHA256 对应表，可供独立像素复核；producer 的 visually_verified 标记本身不是独立审查结论。截图仍来自原测量，没有为审查重跑或修改像素。

## 连续 history 审查修正

fdba 上真实 App 加桥 stub 故障注入回归，在同一事件循环触发双 Undo，第一条成功已把 stub 的权威值从2撤为1，第二条失败后画面仍为2；`history-flight-before.log` 保留该失败。本次只改前端 history 入口，不再逐例调整 request token：Undo/Redo 共用 single-flight，覆盖派发、delta/完整文档读取、权威结果的 React 提交。重复输入既不发第二条 RPC，也不产生淘汰首条结果的新 serial；两种菜单均禁用。锁在 layout commit 后释放，而不是仅在 RPC 返回或 finally 赋值时释放。

真实 App 回归覆盖双 Undo/双 Redo（delta/full）、完整读取仍在途时的重复触发、成功后 exhausted error 保持已提交画面、失败后重试成功，以及切工程后的旧读取不能覆盖新工程。测试保留服务端过期 availability 的故障注入，以覆盖成功后返回错误的序列。原 Go/backend 与认证配置不变，原性能/PNG测量仍归属 eaf；没有重跑昂贵基准或改动像素。

本次目标组合 23 项通过，完整前端 160 项通过；应用/测试 TypeScript 检查及 production build 通过。完整日志为 `history-flight-targeted-after.log`、`history-flight-all.log`、`history-flight-types.log`、`history-flight-test-types.log`、`history-flight-build.log`。Go/backend 未变，沿用 fdba 的全量/race/vet 和平台构建证据，不将前端 stub 回归称作原生磁盘或 100k 性能实测。

最新认证只读检查仍未从现有 store helper 找到 canonical `https://dev.projectmiraland.top:50005` 的可用用户名/密码。遵从原有 `useHttpPath=false` 匹配规则，未修改配置、凭据或旧 API 脚本，也未因此发起网络请求。该轮推送/PR 未执行；远程合并仍待独立复审和认证恢复。
