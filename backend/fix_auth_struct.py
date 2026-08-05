#!/usr/bin/env python3
"""修复 auth.go 中的 PermissionInfo 结构体定义位置问题"""

file_path = "services/auth.go"

with open(file_path, 'r', encoding='utf-8') as f:
    lines = f.readlines()

# 查找需要修改的位置
user_info_line = None
nested_permission_info_start = None
nested_permission_info_end = None

for i, line in enumerate(lines):
    if 'type UserInfo struct' in line:
        user_info_line = i
        # 查找后面几行是否有嵌套的 type PermissionInfo struct
        for j in range(i, min(i + 10, len(lines))):
            if 'type PermissionInfo struct' in lines[j]:
                nested_permission_info_start = j
                # 查找结束的 }
                for k in range(j + 1, len(lines)):
                    if lines[k].strip() == '}':
                        nested_permission_info_end = k
                        break
                break
        break

if nested_permission_info_start and nested_permission_info_end:
    print(f"找到嵌套的 PermissionInfo 定义: 行 {nested_permission_info_start + 1} 到 {nested_permission_info_end + 1}")

    # 删除嵌套的 PermissionInfo 定义（包括前面的注释）
    # 查找注释开始
    comment_start = nested_permission_info_start
    for j in range(nested_permission_info_start - 1, user_info_line, -1):
        if 'PermissionInfo' in lines[j] or '权限' in lines[j]:
            comment_start = j
        else:
            break

    print(f"删除行 {comment_start + 1} 到 {nested_permission_info_end + 1}")
    del lines[comment_start:nested_permission_info_end + 1]

    # 在 UserInfo 结构体定义之前添加独立的 PermissionInfo 类型
    permission_info_code = '''
// PermissionInfo 权限详细信息（用于前端显示权限名称）
type PermissionInfo struct {
\tCode string `json:"code"`
\tName string `json:"name"`
}
'''

    lines.insert(user_info_line, permission_info_code)

    with open(file_path, 'w', encoding='utf-8') as f:
        f.writelines(lines)

    print("修复完成！")
else:
    print("未找到嵌套的 PermissionInfo 定义，可能已经修复过了")
