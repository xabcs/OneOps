#!/bin/bash

echo "=== Web终端功能修复脚本 ==="
echo ""

# 1. 重启后端服务
echo "1. 重启后端服务..."
cd /Users/mpm/Desktop/doc/go/OneOpsV2/OneOps/backend
pkill -f "go run main.go run" 2>/dev/null || true
sleep 2
go run main.go run > /tmp/backend.log 2>&1 &
echo "   后端服务已启动，PID: $!"
sleep 3

# 2. 等待后端启动
echo "2. 等待后端服务启动..."
for i in {1..10}; do
  if curl -s http://localhost:8082/api/health > /dev/null 2>&1; then
    echo "   后端服务已就绪"
    break
  fi
  if [ $i -eq 10 ]; then
    echo "   ❌ 后端服务启动失败"
    exit 1
  fi
  sleep 1
done

# 3. 清除缓存并重新同步菜单
echo "3. 清除RBAC缓存并重新同步菜单..."
curl -s -X POST http://localhost:8082/api/route/invalidateCache | jq .

# 4. 测试调试端点
echo "4. 测试调试端点（需要token）..."
echo "   请在浏览器中打开："
echo "   http://localhost:8082/api/route/debugCache"
echo ""

echo "=== 修复完成 ==="
echo ""
echo "接下来的操作："
echo "1. 清除浏览器缓存 (Ctrl+Shift+Delete)"
echo "2. 在浏览器控制台执行: localStorage.clear(); sessionStorage.clear(); location.reload()"
echo "3. 重新登录并测试 web终端 功能"
echo ""
echo "如果还有问题，请查看后端日志: tail -f /tmp/backend.log"
