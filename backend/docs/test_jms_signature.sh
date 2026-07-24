#!/bin/bash

# JumpServer API 签名测试工具
# 用于验证 AccessKey 和 Secret 配置是否正确

echo "===================================="
echo "JumpServer API 签名测试工具"
echo "===================================="
echo ""

# 配置参数
JMS_URL="${1:-http://jumpserver.hzmeipingmi.com}"
ACCESS_KEY="${2}"
SECRET="${3}"

if [ -z "$ACCESS_KEY" ] || [ -z "$SECRET" ]; then
    echo "用法: $0 <JumpServer_URL> <Access_Key_ID> <Access_Key_Secret>"
    echo ""
    echo "示例:"
    echo "  $0 http://jumpserver.example.com AK-xxx SK-xxx"
    exit 1
fi

# 获取当前 GMT 时间（强制使用英文格式）
DATE=$(LC_ALL=C date -u +"%a, %d %b %Y %H:%M:%S GMT")

# API 端点
ENDPOINT="/api/v1/users/users/"

# 构造签名字符串
METHOD="get"
PATH_QUERY="$ENDPOINT"
SIGNING_STRING="(request-target): $METHOD $PATH_QUERY
date: $DATE"

echo "签名字符串:"
echo "---"
echo "$SIGNING_STRING"
echo "---"
echo ""

# 计算 HMAC-SHA256 签名
SIGNATURE=$(echo -n "$SIGNING_STRING" | openssl dgst -sha256 -hmac "$SECRET" -binary | base64)

echo "计算出的签名: $SIGNATURE"
echo ""

# 构造 Authorization header
AUTH_HEADER="Signature keyId=\"$ACCESS_KEY\",algorithm=\"hmac-sha256\",headers=\"(request-target) date\",signature=\"$SIGNATURE\""

echo "Authorization Header:"
echo "$AUTH_HEADER"
echo ""

# 发送请求
echo "发送请求到: $JMS_URL$ENDPOINT"
echo "---"

curl -v -X GET "$JMS_URL$ENDPOINT" \
  -H "Date: $DATE" \
  -H "Accept: application/json" \
  -H "X-JMS-ORG: 00000000-00000000-00000000-000000000002" \
  -H "Authorization: $AUTH_HEADER"

echo ""
echo "---"
echo "测试完成"
