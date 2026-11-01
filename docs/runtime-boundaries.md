# 设置和日志的局部边界

manual 只是不使用自动编译 debounce，不豁免 language/theme 校验。有效旧版本缺失字段仍在 Load 时补默认值；manual 的零 debounce 仍有效。无效设置在 Save 写文件前拒绝，Load 也不再静默接受。

logger 的内部锁仅保护 writer 的完整 JSON 行及 listener 注册快照；外部 Sink 不在该锁内调用，允许注册后续 listener 或有界重入日志。新注册 listener 不追溯当前 entry。并发 Log 的 writer 仍串行，Sink 不承诺跨并发调用的全局顺序，必须线程安全且不能阻塞应用；没有新增队列或后台线程。应用唯一 Wails event Sink 未增添有状态 listener。回归覆盖注册重入、嵌套日志、快照与并发完整 JSON 行；不声称慢 Sink 已异步化。
