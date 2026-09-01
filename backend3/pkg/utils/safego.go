package utils

import (
	"runtime/debug"

	"oneops/backend3/pkg/logger"

	"go.uber.org/zap"
)

// SafeGo 安全启动后台 goroutine：捕获 panic 并记录堆栈。
// Go 中子 goroutine 的 panic 会直接终止整个进程（无法被上层 recover），
// 所有长期运行的后台任务必须经由本函数启动，防止单点故障打崩服务。
func SafeGo(name string, fn func()) {
	go func() {
		defer recoverPanic(name)
		fn()
	}()
}

// SafeRun 在当前 goroutine 内安全执行一次性任务（捕获 panic 并记录堆栈），
// 适用于 ticker 循环体内逐轮防护：单轮 panic 不会中断后续轮次
func SafeRun(name string, fn func()) {
	defer recoverPanic(name)
	fn()
}

func recoverPanic(name string) {
	if r := recover(); r != nil {
		logger.Error("后台任务 panic 已恢复",
			zap.String("task", name),
			zap.Any("panic", r),
			zap.ByteString("stack", debug.Stack()))
	}
}
