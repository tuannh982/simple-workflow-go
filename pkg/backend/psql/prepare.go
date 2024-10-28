package psql

import (
	"github.com/tuannh982/simple-workflows-go/pkg/backend/psql/persistent"
	"gorm.io/gorm"
)

// PrepareDB only use for testing, don't use this function in production!. You should manually create tables instead
func PrepareDB(db *gorm.DB) error {
	err := db.AutoMigrate(
		&persistent.Event{},
		&persistent.HistoryEvent{},
		&persistent.Task{},
		&persistent.Workflow{},
	)
	return err
}

// TruncateDB only use for testing, don't use this function in production!
func TruncateDB(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		tx.Exec("TRUNCATE TABLE events")
		tx.Exec("TRUNCATE TABLE history_events")
		tx.Exec("TRUNCATE TABLE tasks")
		tx.Exec("TRUNCATE TABLE workflows")
		return nil
	})
}
