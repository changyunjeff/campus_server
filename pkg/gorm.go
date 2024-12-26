package pkg

import (
	"campus2/pkg/config"
	"fmt"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	db     *gorm.DB
	dbOnce sync.Once
)

// GetDB 获取数据库连接
func GetDB(cfg config.Mysql) *gorm.DB {
	var err error
	dbOnce.Do(func() {
		db, err = newDB(cfg)
	})
	if err != nil {
		panic(err)
	}
	return db
}

// newDB 创建新的数据库连接
func newDB(cfg config.Mysql) (*gorm.DB, error) {
	if cfg.DbName == "" {
		return nil, fmt.Errorf("数据库名不能为空")
	}

	mysqlConfig := mysql.Config{
		DSN:                       cfg.Dsn(),
		DefaultStringSize:         191,
		SkipInitializeWithVersion: false,
	}

	gormConfig := &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	db, err := gorm.Open(mysql.New(mysqlConfig), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取底层 *sql.DB 失败: %v", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}
