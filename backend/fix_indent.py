#!/usr/bin/env python3
"""修复 UserInfo 结构体字段的缩进问题"""

file_path = "services/auth.go"

with open(file_path, 'r', encoding='utf-8') as f:
    lines = f.readlines()

# 查找并修复 UserInfo 结构体中的字段缩进
in_user_info = False
user_info_start = None
user_info_end = None

for i, line in enumerate(lines):
    if 'type UserInfo struct' in line:
        in_user_info = True
        user_info_start = i
    elif in_user_info:
        if line.strip().startswith('}') and 'type ' not in line:
            user_info_end = i
            break

if user_info_start and user_info_end:
    print(f"找到 UserInfo 结构体: 行 {user_info_start + 1} 到 {user_info_end + 1}")

    # 修复字段缩进
    for i in range(user_info_start + 1, user_info_end):
        original_line = lines[i]
        stripped = original_line.lstrip()

        # 如果是字段定义
        if stripped and not stripped.startswith('//') and not stripped.startswith('}'):
            # 使用 tab 缩进
            lines[i] = '\t' + stripped + ('\n' if original_line.endswith('\n') else '')
            if lines[i] != original_line:
                print(f"行 {i + 1}: 修复缩进")

    with open(file_path, 'w', encoding='utf-8') as f:
        f.writelines(lines)

    print("修复完成！")
else:
    print("未找到 UserInfo 结构体")
