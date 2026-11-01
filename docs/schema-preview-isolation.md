# 有效 Schema 的只读展开

有效 Schema 是编辑器视图，不是源 Schema。旧实现浅拷贝 Node 后直接替换共享 Fields slice，展开引用会修改原 document/registry 的字段，且并发 GetTable 的只读展开可能竞争。最小模型回归在旧实现失败。

现在为每层 Fields 分配独立 slice 再递归展开，Element 继续独立复制。nil slice 语义不变，源表达式、引用身份、默认值和文件均不参与写入。本批是模型所有权纠错，不是重构；单测验证原模型/registry 不变，race 回归并行读取。
