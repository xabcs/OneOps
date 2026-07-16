#!/bin/bash

# 找出所有重复的函数声明
echo "检查重复的函数声明..."

# 获取所有函数名
functions=$(grep -o "export function [a-zA-Z]*" src/service/api/k8s.ts | awk '{print $3}')

# 找出重复的函数名
duplicates=$(echo "$functions" | sort | uniq -d)

if [ -z "$duplicates" ]; then
  echo "没有发现重复的函数声明"
  exit 0
fi

echo "发现重复的函数:"
for func in $duplicates; do
  echo "  - $func"
  # 显示所有重复函数的行号
  grep -n "export function $func" src/service/api/k8s.ts
done