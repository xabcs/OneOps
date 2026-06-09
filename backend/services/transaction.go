package services

import (
	"database/sql"

	"gorm.io/gorm"
)

// TransactionManager 事务管理器
// 提供简洁的事务API，自动处理提交和回滚
type TransactionManager struct {
	db *gorm.DB
}

// NewTransactionManager 创建事务管理器
func NewTransactionManager(db *gorm.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

// Transaction 执行无返回值的事务
// 使用示例：
//
//	err := tm.Transaction(func(tx *gorm.DB) error {
//	    if err := tx.Create(&user).Error; err != nil {
//	        return err  // 返回错误会自动回滚
//	    }
//	    if err := tx.Create(&profile).Error; err != nil {
//	        return err  // 返回错误会自动回滚
//	    }
//	    return nil  // 返回nil会自动提交
//	})
func (tm *TransactionManager) Transaction(fn func(tx *gorm.DB) error) error {
	return tm.db.Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// TransactionWithResult 执行有返回值的事务
// 使用示例：
//
//	result, err := tm.TransactionWithResult(func(tx *gorm.DB) (interface{}, error) {
//	    var user User
//	    if err := tx.First(&user, id).Error; err != nil {
//	        return nil, err
//	    }
//	    if err := tx.Delete(&user).Error; err != nil {
//	        return nil, err
//	    }
//	    return user.Name, nil
//	})
func (tm *TransactionManager) TransactionWithResult(fn func(tx *gorm.DB) (interface{}, error)) (interface{}, error) {
	var result interface{}
	err := tm.db.Transaction(func(tx *gorm.DB) error {
		r, err := fn(tx)
		if err != nil {
			return err
		}
		result = r
		return nil
	})
	return result, err
}

// TransactionWithError 执行事务，只关心错误（无返回值）
// 这是Transaction的别名，提供更清晰的语义
func (tm *TransactionManager) TransactionWithError(fn func(tx *gorm.DB) error) error {
	return tm.Transaction(fn)
}

// TransactionWithTypedResult 执行有类型化返回值的事务（使用泛型）
// 使用示例：
//
//	userName, err := TransactionWithTypedResult(func(tx *gorm.DB) (string, error) {
//	    var user User
//	    if err := tx.First(&user, id).Error; err != nil {
//	        return "", err
//	    }
//	    return user.Name, nil
//	})
func TransactionWithTypedResult[T any](db *gorm.DB, fn func(tx *gorm.DB) (T, error)) (T, error) {
	var result T
	err := db.Transaction(func(tx *gorm.DB) error {
		r, err := fn(tx)
		if err != nil {
			return err
		}
		result = r
		return nil
	})
	return result, err
}

// TransactionOptions 事务选项
type TransactionOptions struct {
	// TxOptions SQL事务选项（隔离级别、只读等）
	TxOptions *sql.TxOptions
	// ReadOnly 是否只读事务
	ReadOnly bool
	// Timeout 超时时间（秒）
	Timeout int
}

// TransactionWithOptions 使用指定选项执行事务
func (tm *TransactionManager) TransactionWithOptions(opts TransactionOptions, fn func(tx *gorm.DB) error) error {
	// 开始事务
	tx := tm.db.Begin(opts.TxOptions)

	// 执行事务函数
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // 重新抛出panic
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

// BatchIn 批量执行多个事务操作（失败时继续执行，收集所有错误）
func (tm *TransactionManager) BatchIn(operations []func(tx *gorm.DB) error) []error {
	errors := make([]error, 0, len(operations))

	for _, op := range operations {
		err := tm.Transaction(func(tx *gorm.DB) error {
			return op(tx)
		})
		if err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// BatchInAtomic 批量执行多个操作在同一事务中（任何一个失败则全部回滚）
func (tm *TransactionManager) BatchInAtomic(operations []func(tx *gorm.DB) error) error {
	return tm.Transaction(func(tx *gorm.DB) error {
		for _, op := range operations {
			if err := op(tx); err != nil {
				return err // 任何一个失败都会回滚整个事务
			}
		}
		return nil
	})
}

// RetryOnDeadlock 重试事务（当遇到死锁时自动重试）
func (tm *TransactionManager) RetryOnDeadlock(fn func(tx *gorm.DB) error, maxRetries int) error {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		err := tm.Transaction(func(tx *gorm.DB) error {
			return fn(tx)
		})
		if err == nil {
			return nil
		}
		lastErr = err

		// 检查是否是死锁错误（MySQL 1213）
		// TODO: 根据不同的数据库返回不同的错误码
	}

	return lastErr
}
