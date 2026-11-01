# CUE 行校验机制

`cuerows.Validate` 是无 I/O 的机制：在调用者已有 CUE context 中依次验证主键、实际提供的字段、整行，返回候选 CUE value 和带位置的原始错误。不解码、不格式化诊断、不加载 package，也不改变输入 map。

compiler 仍显式选择表的 package，保留主键类型与外键 Schema 检查、诊断 key/位置/次数，并且仅无错误时解码最终值和默认值。workspace 预览仍使用原来的未限定 package loader、原诊断策略，不将 CUE 默认值写回 source cell。manual 的旧诊断、pending 及显式编译时点也没有合并。多 package 同目录的预览/编译差别是保留下来的既有行为，本批没有借重构改变它。

提取前先加入特征测试：默认整数/空数组/null 的最终解码与预览源表示、字段和整行重复诊断、主键诊断位置与 compiler key、多 package 的选择差异。纯机制测试验证无输入改写及三类原始 issue。云岚基准显示完整 CUE 编译约195ms，索引约0.05ms；没有引入缓存或 debounce，也不宣称本次是性能优化。

可重跑基准由同轮独立生命周期批次提供：`internal/workspace/quality_benchmark_test.go::BenchmarkYunlan`，fixture 是 `docs/examples/yunlan-rpg/yunlan-rpg.dtproj`。运行命令及前后小样本范围见 [workspace 生命周期](workspace-lifecycle.md) 和 [质量台账](code-quality.md)；这些数字不是性能提升结论。
