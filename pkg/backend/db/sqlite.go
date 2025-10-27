package db

import (
	"fmt"
	"sync"
	"time"

	"github.com/tuannh982/simple-workflow-go/pkg/backend"
	"github.com/tuannh982/simple-workflow-go/pkg/backend/persistent"
	"github.com/tuannh982/simple-workflow-go/pkg/codec"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TODO:
func NewSQLiteBackend(lockedBy string, lockExpirationDuration time.Duration, codec codec.Codec, sqliteDB SQLiteDB, logger *zap.Logger) backend.Backend {
	workflowRepo := persistent.NewWorkflowRepository(sqliteDB.database)
	historyEventRepo := persistent.NewHistoryEventRepository(sqliteDB.database)
	taskRepo := persistent.NewTaskRepository(sqliteDB.database)
	eventRepo := persistent.NewEventRepository(sqliteDB.database)
	return &backend.SimpleWorkflowGoBackend{
		LockedBy:               lockedBy,
		LockExpirationDuration: lockExpirationDuration,
		Codec:                  codec,
		DB:                     sqliteDB.database,
		DBType:                 backend.DBTypeSQLite,
		WorkflowRepo:           workflowRepo,
		HistoryEventRepo:       historyEventRepo,
		TaskRepo:               taskRepo,
		EventRepo:              eventRepo,
		Logger:                 logger,
		WorkflowTaskMu:         &sync.Mutex{},
		ActivityTaskMu:         &sync.Mutex{},
	}
}

type SQLiteDB struct {
	database *gorm.DB
}

func (s *SQLiteDB) Type() DatabaseType {
	return SQLiteDBType
}

func (s *SQLiteDB) Prepare() error {
	err := s.database.AutoMigrate(
		&persistent.Event{},
		&persistent.HistoryEvent{},
		&persistent.Task{},
		&persistent.Workflow{},
	)
	return err
}

func (s *SQLiteDB) Truncate() error {
	return s.database.Transaction(func(tx *gorm.DB) error {
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
	s.database = d
	return err
}
