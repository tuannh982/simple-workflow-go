package db

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/tuannh982/simple-workflow-go/pkg/backend"
	"github.com/tuannh982/simple-workflow-go/pkg/backend/persistent"
	"github.com/tuannh982/simple-workflow-go/pkg/dataconverter"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewPSQLBackend(
	lockedBy string,
	lockExpirationDuration time.Duration,
	codec dataconverter.Codec,
	db *gorm.DB,
	logger *zap.Logger,
) backend.Backend {
	workflowRepo := persistent.NewWorkflowRepository(db)
	historyEventRepo := persistent.NewHistoryEventRepository(db)
	taskRepo := persistent.NewTaskRepository(db)
	eventRepo := persistent.NewEventRepository(db)
	return &backend.SimpleWorkflowGoBackend{
		LockedBy:               lockedBy,
		LockExpirationDuration: lockExpirationDuration,
		Codec:                  codec,
		DB:                     db,
		WorkflowRepo:           workflowRepo,
		HistoryEventRepo:       historyEventRepo,
		TaskRepo:               taskRepo,
		EventRepo:              eventRepo,
		Logger:                 logger,
		WorkflowTaskMu:         &sync.Mutex{},
		ActivityTaskMu:         &sync.Mutex{},
	}
}

type PostgresDB struct{}

// Prepare only use for testing, don't use this function in production!. You should manually create tables instead
func (pg *PostgresDB) Prepare(db *gorm.DB) error {
	err := db.AutoMigrate(
		&persistent.Event{},
		&persistent.HistoryEvent{},
		&persistent.Task{},
		&persistent.Workflow{},
	)
	return err
}

// Truncate only use for testing, don't use this function in production!
func (pg *PostgresDB) Truncate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		tx.Exec("TRUNCATE TABLE events")
		tx.Exec("TRUNCATE TABLE history_events")
		tx.Exec("TRUNCATE TABLE tasks")
		tx.Exec("TRUNCATE TABLE workflows")
		return nil
	})
}

var DefaultConnectConfig = &gorm.Config{
	DisableForeignKeyConstraintWhenMigrating: true,
	Logger: logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Silent, // Log level
			IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      true,          // Don't include params in the SQL log
			Colorful:                  false,         // Disable color
		},
	),
}

func (pg *PostgresDB) Connect(c ConnectionDetails) (*gorm.DB, error) {
	connStr := fmt.Sprintf(
		"Host=%s user=%s Password=%s dbname=%s Port=%d",
		c.Host, c.Username, c.Password, c.Database, c.Port,
	)
	if c.Config == nil {
		c.Config = DefaultConnectConfig
	}
	return gorm.Open(postgres.Open(connStr), c.Config)
}
