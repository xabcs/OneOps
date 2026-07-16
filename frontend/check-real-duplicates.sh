#!/bin/bash

echo "检查真正重复的函数声明..."

# 获取所有完整的函数名
all_funcs=$(grep -o "export function [a-zA-Z0-9]*" src/service/api/k8s.ts | awk '{print $3}')

# 找出真正重复的完整函数名
duplicates=$(echo "$all_funcs" | sort | uniq -d)

if [ -z "$duplicates" ]; then
  echo "没有发现重复的函数声明"
  exit 0
fi

echo "发现真正重复的函数:"
for func in $duplicates; do
  echo "  重复函数: $func"
  grep -n "export function $func" src/service/api/k8s.ts
  echo "---"
done