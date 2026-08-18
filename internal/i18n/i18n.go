// Package i18n translates user-facing backend messages (error toasts) using a
// small built-in catalog. Frontend UI text is localized separately on the
// client, so this package only covers messages the Go backend returns as
// errors. Messages without a translation fall back to English, then to the key.
package i18n

import "fmt"

const (
	English  = "en"
	Chinese  = "zh-CN"
	Fallback = English
)

// catalog maps a message key to its per-language template. Templates use Go
// fmt verbs (%s, %q, %d, ...) and must keep the same argument order across
// languages.
var catalog = map[string]map[string]string{
	"workspace.noProject": {
		English: "no project is open",
		Chinese: "未打开项目",
	},
	"workspace.tableNotInProject": {
		English: "table %q is not part of the open project",
		Chinese: "表 %q 不属于当前打开的项目",
	},
	"workspace.tableExists": {
		English: "table %q already exists",
		Chinese: "表 %q 已存在",
	},
	"workspace.tableNameInvalid": {
		English: "table name and namespace must use lowercase letters, digits, and underscores",
		Chinese: "表名和命名空间只能使用小写字母、数字和下划线",
	},
	"workspace.tableReferenced": {
		English: "table %q is referenced by %s",
		Chinese: "表 %q 被 %s 引用",
	},
	"workspace.rowReferenced": {
		English: "row %q is referenced by %s.%s.%s",
		Chinese: "行 %q 被 %s.%s.%s 引用",
	},
	"workspace.schemaNamespaceExists": {
		English: "schema namespace %q already exists",
		Chinese: "Schema 命名空间 %q 已存在",
	},
	"workspace.schemaNamespaceInvalid": {
		English: "schema namespace must use lowercase letters, digits, and underscores",
		Chinese: "Schema 命名空间只能使用小写字母、数字和下划线",
	},
	"workspace.schemaPackageNotInProject": {
		English: "schema package %q is not part of the open project",
		Chinese: "Schema 包 %q 不属于当前打开的项目",
	},
	"workspace.schemaPackageOutside": {
		English: "schema package %q is outside the project directory",
		Chinese: "Schema 包 %q 位于项目目录之外",
	},
	"workspace.schemaFileNotInProject": {
		English: "schema file %q is not part of the open project",
		Chinese: "Schema 文件 %q 不属于当前打开的项目",
	},
	"workspace.pathNotInProject": {
		English: "path %q is not part of the open project",
		Chinese: "路径 %q 不属于当前打开的项目",
	},
	"workspace.nothingToUndoRedo": {
		English: "nothing to %s",
		Chinese: "没有可%s的内容",
	},
	"workspace.nothingToUndo": {
		English: "nothing to undo",
		Chinese: "没有可撤销的内容",
	},
	"workspace.nothingToRedo": {
		English: "nothing to redo",
		Chinese: "没有可重做的内容",
	},
	"workspace.exportBlocked": {
		English: "export blocked by %d diagnostics: %s",
		Chinese: "导出被 %d 条诊断信息阻止：%s",
	},
	"workspace.noSelectedTables": {
		English: "no selected tables are available for export",
		Chinese: "没有可导出的已选表",
	},
	"workspace.noExportTargets": {
		English: "no configured export formats are available",
		Chinese: "没有可用的已配置导出格式",
	},
	"workspace.unsupportedLanguage": {
		English: "unsupported language %q",
		Chinese: "不支持的语言 %q",
	},
	"workspace.unsupportedTheme": {
		English: "unsupported theme %q",
		Chinese: "不支持的主题 %q",
	},
	"workspace.unsupportedBuildMode": {
		English: "unsupported build mode %q",
		Chinese: "不支持的编译模式 %q",
	},
	"workspace.unsupportedNavigationKind": {
		English: "unsupported navigation kind %q",
		Chinese: "不支持的导航类型 %q",
	},
	"schema.referenceNotFound": {
		English: "Schema reference %s not found",
		Chinese: "找不到 Schema 引用 %s",
	},
	"schema.referenceNotFoundHint": {
		English: "Schema reference %s not found; local references use _local.%s and public references use package.%s",
		Chinese: "找不到 Schema 引用 %s；本地引用使用 _local.%s，公共引用使用 package.%s",
	},
	"schema.referenceAmbiguous": {
		English: "Schema reference %s is ambiguous; qualify it with a package name",
		Chinese: "Schema 引用 %s 存在歧义；请使用包名限定",
	},
	"schema.referencedBy": {
		English: "Schema %s is referenced by %s field %s",
		Chinese: "Schema %s 被 %s 的字段 %s 引用",
	},
	"navigation.noReferences": {
		English: "no references to %s.%s",
		Chinese: "没有指向 %s.%s 的引用",
	},
	"navigation.noFormula": {
		English: "no formula at %s.%s",
		Chinese: "%s.%s 没有公式",
	},
	"navigation.formulaNoReferences": {
		English: "formula has no resolvable references",
		Chinese: "公式没有可解析的引用",
	},
	"navigation.fkTargetNotFound": {
		English: "foreign key target table %q not found",
		Chinese: "找不到外键目标表 %q",
	},
	"navigation.notForeignKey": {
		English: "column %q is not a foreign key",
		Chinese: "列 %q 不是外键",
	},
	"config.exportEscapes": {
		English: "export directory %q escapes the project directory",
		Chinese: "导出目录 %q 超出了项目目录",
	},
}

// Message returns the message for key in the requested language, formatted
// with args. It falls back to English and then to the raw key.
func Message(language, key string, args []any) string {
	messages, ok := catalog[key]
	if !ok {
		if len(args) == 0 {
			return key
		}
		return fmt.Sprintf(key, args...)
	}
	template, ok := messages[language]
	if !ok {
		template = messages[Fallback]
	}
	if template == "" {
		if len(args) == 0 {
			return key
		}
		return fmt.Sprintf(key, args...)
	}
	if len(args) == 0 {
		return template
	}
	return fmt.Sprintf(template, args...)
}
