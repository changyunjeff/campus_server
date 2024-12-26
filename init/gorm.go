package init

import (
	"campus2/pkg/global"
	"fmt"
)

func RegisterTables() error {
	db := global.GVA_DB
	err := db.AutoMigrate()
	if err != nil {
		return fmt.Errorf("注册表格时出错: %w", err)
	}
	return nil
}
