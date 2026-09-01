package cmdb

import "testing"

// TestNormalizeCommandForMatch 验证命令归一化能对抗常见混淆绕过（黑名单匹配前置步骤）
func TestNormalizeCommandForMatch(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"普通命令", "rm -rf /", "rm -rf /"},
		{"Tab分隔", "rm\t-rf\t/", "rm -rf /"},
		{"多空格", "rm  -rf   /", "rm -rf /"},
		{"Tab补全残留", "rm -rf /us\t", "rm -rf /us"},
		{"反斜杠续行", "rm -rf \\\n/", "rm -rf /"},
		{"反斜杠续行CR", "rm -rf \\\r/", "rm -rf /"},
		{"引号混淆", "'r'm -rf /", "rm -rf /"},
		{"双引号拼接", "r\"m\" -rf /", "rm -rf /"},
		{"转义混淆", "r\\m -rf /", "rm -rf /"},
		{"IFS替换", "rm${IFS}-rf${IFS}/", "rm -rf /"},
		{"IFS裸变量", "rm$IFS-rf$IFS/", "rm -rf /"},
		{"bracketed paste标记", "\x1b[200~rm -rf /\x1b[201~", "rm -rf /"},
		{"ANSI颜色", "\x1b[31mrm -rf /\x1b[0m", "rm -rf /"},
		{"大小写", "RM -RF /", "rm -rf /"},
		{"首尾空白", "  rm -rf /  ", "rm -rf /"},
		{"中文保留", "echo 你好", "echo 你好"},
		{"空串", "", ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := normalizeCommandForMatch(c.in); got != c.want {
				t.Errorf("normalizeCommandForMatch(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}
