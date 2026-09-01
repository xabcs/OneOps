package utils

import (
	"bytes"
	"testing"
)

// TestExtractSubmittedLines 验证行编辑感知的命令行提取：
// 帧内回车、多行粘贴、行编辑（Ctrl-C/Ctrl-U/退格）、ANSI 序列剥离、UTF-8 回退
func TestExtractSubmittedLines(t *testing.T) {
	cases := []struct {
		name      string
		in        string
		wantLines []string
		wantPend  string
		wantSub   bool
	}{
		{
			name:      "帧内回车单帧双命令",
			in:        "ls\r rm -rf /\r",
			wantLines: []string{"ls", " rm -rf /"},
			wantPend:  "",
			wantSub:   true,
		},
		{
			name:      "bracketed paste 整块粘贴",
			in:        "\x1b[200~echo hi\r\nls -l\r\n\x1b[201~",
			wantLines: []string{"echo hi", "ls -l"},
			wantPend:  "",
			wantSub:   true,
		},
		{
			name:      "普通单行提交",
			in:        "uptime\r",
			wantLines: []string{"uptime"},
			wantPend:  "",
			wantSub:   true,
		},
		{
			name:      "无回车未提交",
			in:        "ech",
			wantLines: nil,
			wantPend:  "ech",
			wantSub:   false,
		},
		{
			name:      "多行帧尾带未提交片段",
			in:        "a\rb\rpar",
			wantLines: []string{"a", "b"},
			wantPend:  "par",
			wantSub:   true,
		},
		{
			name:      "退格修正命令",
			in:        "lss\x7f -l\r",
			wantLines: []string{"ls -l"},
			wantSub:   true,
		},
		{
			name:      "Ctrl-U 作废重输",
			in:        "rm -rf \x15ls\r",
			wantLines: []string{"ls"},
			wantSub:   true,
		},
		{
			name:      "Ctrl-C 作废整行后新命令",
			in:        "rm -rf /\x03ls\r",
			wantLines: []string{"ls"},
			wantSub:   true,
		},
		{
			name:      "空行提交",
			in:        "\r",
			wantLines: []string{""},
			wantSub:   true,
		},
		{
			name:      "ANSI 颜色序列不入行",
			in:        "\x1b[31mrm -rf /\x1b[0m\r",
			wantLines: []string{"rm -rf /"},
			wantSub:   true,
		},
		{
			name:      "UTF-8 中文整体回退",
			in:        "echo 你好\x7f\x7f世界\r", // 删掉"好"和"你"再输入"世界"
			wantLines: []string{"echo 世界"},
			wantSub:   true,
		},
		{
			name:      "Tab 转空格保持单词边界",
			in:        "rm\t-rf\t/\r",
			wantLines: []string{"rm -rf /"},
			wantSub:   true,
		},
		{
			name:      "回车后 Ctrl-U 作用于新行",
			in:        "ls\rrm -rf \x15pwd\r",
			wantLines: []string{"ls", "pwd"},
			wantSub:   true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			lines, pending, hasSubmit := ExtractSubmittedLines([]byte(c.in))
			if hasSubmit != c.wantSub {
				t.Errorf("hasSubmit = %v, want %v", hasSubmit, c.wantSub)
			}
			if len(lines) != len(c.wantLines) {
				t.Fatalf("lines = %q, want %q", lines, c.wantLines)
			}
			for i := range lines {
				if string(lines[i]) != c.wantLines[i] {
					t.Errorf("lines[%d] = %q, want %q", i, lines[i], c.wantLines[i])
				}
			}
			if string(pending) != c.wantPend {
				t.Errorf("pending = %q, want %q", pending, c.wantPend)
			}
		})
	}
}

// TestExtractSubmittedLinesNoAliasing 验证 lines/pending 不共享底层数组（防后续缓冲写坏已提取行）
func TestExtractSubmittedLinesNoAliasing(t *testing.T) {
	data := []byte("abc\rdef")
	lines, pending, _ := ExtractSubmittedLines(data)
	// 修改 pending 底层数组不应影响 lines[0]
	for i := range pending {
		pending[i] = 'X'
	}
	if !bytes.Equal(lines[0], []byte("abc")) {
		t.Errorf("lines[0] corrupted: %q", lines[0])
	}
}
