#!/bin/bash
# Grafana匿名访问配置脚本

GRAFANA_CONFIG="/etc/grafana/grafana.ini"
BACKUP_CONFIG="/etc/grafana/grafana.ini.backup"

echo "=== Grafana 匿名访问配置工具 ==="
echo ""

# 检查是否为root用户
if [ "$EUID" -ne 0 ]; then
  echo "请使用sudo运行此脚本"
  exit 1
fi

# 备份原配置
echo "1. 备份原配置文件..."
cp $GRAFANA_CONFIG $BACKUP_CONFIG
echo "   配置已备份到: $BACKUP_CONFIG"
echo ""

# 添加匿名访问配置
echo "2. 配置匿名访问..."
grep -q "^\[auth.anonymous\]" $GRAFANA_CONFIG || echo "" >> $GRAFANA_CONFIG
grep -q "^enabled = true" $GRAFANA_CONFIG || sed -i '/^\[auth.anonymous\]/a enabled = true' $GRAFANA_CONFIG
grep -q "^org_role = Viewer" $GRAFANA_CONFIG || sed -i '/^\[auth.anonymous\]/a org_role = Viewer' $GRAFANA_CONFIG
echo "   匿名访问已启用"
echo ""

# 配置允许iframe嵌入
echo "3. 配置允许iframe嵌入..."
grep -q "^allow_embedding = true" $GRAFANA_CONFIG || sed -i '/^\[security\]/a allow_embedding = true' $GRAFANA_CONFIG
echo "   iframe嵌入已允许"
echo ""

# 配置隐藏版本信息
echo "4. 配置隐藏版本信息..."
grep -q "^hide_version = true" $GRAFANA_CONFIG || sed -i '/^\[security\]/a hide_version = true' $GRAFANA_CONFIG
echo "   版本信息已隐藏"
echo ""

# 重启Grafana
echo "5. 重启Grafana服务..."
systemctl restart grafana-server
echo "   Grafana服务已重启"
echo ""

# 等待服务启动
echo "6. 等待服务启动..."
sleep 5
echo ""

# 测试访问
echo "7. 测试匿名访问..."
TEST_URL="http://192.168.4.168:3000/d/fdwecevaqo7wge/cloud-dns-record-info?orgId=1&kiosk"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$TEST_URL")

if [ "$HTTP_CODE" -eq 200 ]; then
  echo "   ✓ 匿名访问配置成功！"
  echo "   测试URL: $TEST_URL"
else
  echo "   ✗ 访问失败，HTTP状态码: $HTTP_CODE"
  echo "   请检查配置文件: $GRAFANA_CONFIG"
fi

echo ""
echo "=== 配置完成 ==="
echo ""
echo "如果还有问题，请手动检查配置："
echo "  sudo cat $GRAFANA_CONFIG | grep -A 3 '\\[auth.anonymous\\]'"
echo "  sudo cat $GRAFANA_CONFIG | grep -A 2 'allow_embedding'"
