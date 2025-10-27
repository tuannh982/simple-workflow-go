package db

import (
	"fmt"

	"github.com/tuannh982/simple-workflow-go/pkg/backend/persistent"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	SQLiteDBType DatabaseType = "sqlite"
)

type SQLiteDB struct {
	Database *gorm.DB
}

func (s *SQLiteDB) Type() DatabaseType {
	return SQLiteDBType
}

func (s *SQLiteDB) Prepare() error {
	err := s.Database.AutoMigrate(
		&persistent.Event{},
		&persistent.HistoryEvent{},
		&persistent.Task{},
		&persistent.Workflow{},
	)
	return err
}

func (s *SQLiteDB) Truncate() error {
	return s.Database.Transaction(func(tx *gorm.DB) error {
		tx.Exec("DELETE FROM events")
		tx.Exec("DELETE FROM history_events")
		tx.Exec("DELETE FROM tasks")
		tx.Exec("DELETE FROM workflows")
		return nil
	})
}

func (s *SQLiteDB) Connect(c ConnectionDetails) error {
	connStr := fmt.Sprintf("%s", c.DatabaseName) // is of the form "test.db"
	if c.Config == nil {
		c.Config = nil
	}
	if c.Config == nil {
		c.Config = DefaultConnectConfig
	}
	d, err := gorm.Open(sqlite.Open(connStr), c.Config)
	s.Database = d
	return err
}
