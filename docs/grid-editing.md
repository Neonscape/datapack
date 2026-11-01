# 表格编辑的局部边界

TableGrid 保留坐标、Glide 集成与当前打开的单元格编辑器。以下边界不再混在组件事件处理中：

- `useRowDraft` 拥有未保存行、ID校验、值/公式分拆、创建成功清空和失败保留/单次局部错误。输入是字段身份，不是列坐标；创建回调不包含IPC、JSON或语言资源依赖，Grid负责把它接到既有 `__create__` 协议。
- `selection` 是不依赖 React/Glide 的纯坐标规则：有效单元格包括 ID 和草稿行；范围 clipboard 排除两者并拒绝失效边界。结构编辑后的选区协调和复制粘贴使用同一规则，仍由 Grid 把数字集合适配成 Glide CompactSelection。
- `eventOwnership` 是 clipboard/F12 共用的归属判断：活动标签、网格实际焦点、未被消费事件；原生 input/textarea/select/contenteditable 保留自己的行为。Grid保留事件注册/清理，没有另加全局监听器。

此批保持既有编辑语义，不重写已有 valueCodecs，不引入状态框架或并发保存队列。数组/结构体的专用编辑器、公式栏和选择仍由当前组件协调；不是把复杂输入改成另一套模型。保留所有草稿失败、ID兼容、复制粘贴、隐藏标签、F12、数组/结构和Undo/Redo浏览器断言。
