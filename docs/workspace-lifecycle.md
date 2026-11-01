# Workspace 发布职责

Service 继续持有互斥锁、当前工程、编辑命令和历史栈。`viewLifecycle` 只拥有 source revision、上次编译 revision 和编译状态，负责纯状态转换；它不读文件、不触发编译、不安排计时器。当前模式与历史可用性由调用方明确传入，因此 manual 的旧诊断/时间戳与新 pending 可以并存。

`compile_pipeline` 聚合源收集、索引构建、完整编译及视图构造。它调用生命周期对象记录实际编译、再发布视图；领域命令仍通过成功刷新后记历史的已有协议运行。没有统一宽容树解析与严格编辑，没有缓存源文件，没有改变 auto/manual 的调用时点。

先加入的特征测试锁定：manual 修复不覆盖旧诊断和时间戳，只有显式 Compile 更新；auto 成功编辑会编译；失败刷新保持 revision/history/compile；重开即使 manual 也执行初始编译并清历史。纯状态单测验证发布不会改写缓存模式或输入 View，nil diagnostics 的UI表示保持空数组。

云岚基线（02fe4f1，Linux/Go1.26.5，3次×3迭代）扫描约4.9ms、解析11.5–12.5ms、索引0.043–0.062ms、完整编译191–198ms。主要成本是CUE编译，未引入跨命令缓存、增量编译或debounce改变可观察行为。完整命令及输出在工作任务证据中，基准可用 `go test ./internal/workspace -run '^$' -bench '^BenchmarkYunlan$' -benchtime=3x -count=3 -benchmem` 重现。
