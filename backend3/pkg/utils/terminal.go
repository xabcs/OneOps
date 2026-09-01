package utils

// ExtractSubmittedLines 行编辑感知地从终端输入流提取已提交命令行：
// 逐字节回放远端行编辑语义——ANSI 转义序列不入行；Ctrl-C/Ctrl-U 清空当前行；
// 退格/删除回退一个字符（UTF-8 多字节字符整体回退）；CRLF 计为一次提交；回车/换行提交当前行；
// Tab 转为空格入行（保持单词边界，防止 "rm\trf" 类分隔符变化漏拦）。
// 返回的 lines/pending 均为回放后的"远端实际行内容"（纯可显示文本），
// 保证黑名单送检文本与 shell 实际执行的命令行一致。
// \r/\n/0x03/0x15/0x7f/0x1b 不会出现在 UTF-8 续字节（0x80-0xBF）中，按字节扫描安全。
// SSH 堡垒机与 K8s 容器终端共用本函数，保证两类终端拦截/审计口径一致。
func ExtractSubmittedLines(data []byte) (lines [][]byte, pending []byte, hasSubmit bool) {
	var line []byte
	inEscape := false

	for i := 0; i < len(data); i++ {
		b := data[i]
		if b == 0x1b { // ESC：ANSI 转义序列开始，整段不入行
			inEscape = true
			continue
		}
		if inEscape {
			// CSI 序列终止符：字母或 '~'（如 bracketed paste 标记 \x1b[200~）
			if b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b == '~' {
				inEscape = false
			}
			continue
		}
		switch b {
		case '\r', '\n':
			lines = append(lines, append([]byte(nil), line...))
			hasSubmit = true
			line = line[:0]
			// CRLF / LFCR 计为一次提交，跳过紧随的配对换行符
			if i+1 < len(data) && (data[i+1] == '\r' || data[i+1] == '\n') && data[i+1] != b {
				i++
			}
		case 0x03, 0x15: // Ctrl-C / Ctrl-U：作废当前行
			line = line[:0]
		case 0x08, 0x7f: // Backspace / Ctrl-H：回退一个字符
			// UTF-8：先弹掉被删字符的全部续字节（0x80-0xBF），再弹首字节
			for len(line) > 0 && line[len(line)-1] >= 0x80 && line[len(line)-1] < 0xC0 {
				line = line[:len(line)-1]
			}
			if len(line) > 0 {
				line = line[:len(line)-1]
			}
		case '\t': // Tab 转空格：保持单词边界可匹配
			line = append(line, ' ')
		default:
			if b >= 32 && b <= 126 || b >= 0x80 { // 可打印 ASCII + UTF-8 多字节
				line = append(line, b)
			}
		}
	}
	pending = append([]byte(nil), line...) // 拷贝，避免与 lines 共享底层数组
	return
}
