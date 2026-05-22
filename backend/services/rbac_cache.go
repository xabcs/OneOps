package services

import (
	"oneops/backend/models"
	"sync"
	"time"
)

// rbacCacheEntry 缓存条目
type rbacCacheEntry struct {
	menuTree    []*models.Menu
	permissions []string
	roles       []*models.Role
	expireAt    time.Time
}

// rbacCache 全局 RBAC 结果缓存（per user）
var rbacCache = struct {
	mu      sync.RWMutex
	entries map[uint]*rbacCacheEntry
}{
	entries: make(map[uint]*rbacCacheEntry),
}

const rbacCacheTTL = 5 * time.Minute

// InvalidateRBACCache 使指定用户（或所有用户）的 RBAC 缓存失效
// userID=0 时清除所有用户的缓存（菜单/角色变更后调用）
func InvalidateRBACCache(userID uint) {
	rbacCache.mu.Lock()
	defer rbacCache.mu.Unlock()
	if userID == 0 {
		rbacCache.entries = make(map[uint]*rbacCacheEntry)
	} else {
		delete(rbacCache.entries, userID)
	}
}
