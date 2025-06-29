package dbTasks

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB(withConnString string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(withConnString), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	return db, nil
}
