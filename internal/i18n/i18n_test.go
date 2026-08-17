package i18n

import "testing"

func TestMessageTranslatesAndFallsBack(t *testing.T) {
	if got := Message(English, "workspace.noProject", nil); got != "no project is open" {
		t.Fatalf("English message = %q", got)
	}
	if got := Message(Chinese, "workspace.noProject", nil); got != "未打开项目" {
		t.Fatalf("Chinese message = %q", got)
	}
	// Unknown language falls back to English.
	if got := Message("fr", "workspace.noProject", nil); got != "no project is open" {
		t.Fatalf("fallback message = %q", got)
	}
	// Unknown key returns the key itself when there are no args.
	if got := Message(English, "does.not.exist", nil); got != "does.not.exist" {
		t.Fatalf("unknown key = %q", got)
	}
}

func TestMessageFormatsArgs(t *testing.T) {
	if got := Message(English, "workspace.tableNotInProject", []any{"/tables/item.cue"}); got != "table \"/tables/item.cue\" is not part of the open project" {
		t.Fatalf("formatted message = %q", got)
	}
	if got := Message(Chinese, "workspace.exportBlocked", []any{2, "x"}); got != "导出被 2 条诊断信息阻止：x" {
		t.Fatalf("formatted Chinese message = %q", got)
	}
}
