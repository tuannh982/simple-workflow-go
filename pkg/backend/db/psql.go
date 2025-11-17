package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tuannh982/simple-workflow-go/pkg/backend/persistent"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const PostgresDBType = "postgres"

type PostgresDB struct {
	Database *gorm.DB
}

func (pg *PostgresDB) Type() DatabaseType {
	return PostgresDBType
}

// Prepare only use for testing, don't use this function in production!. You should manually create tables instead
func (pg *PostgresDB) Prepare() error {
	err := pg.Database.AutoMigrate(
		&persistent.Event{},
		&persistent.HistoryEvent{},
		&persistent.Task{},
		&persistent.Workflow{},
	)
	return err
}

// Truncate only use for testing, don't use this function in production!
func (pg *PostgresDB) Truncate(db *gorm.DB) error {
	return pg.Database.Transaction(func(tx *gorm.DB) error {
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

func (pg *PostgresDB) Connect(c ConnectionDetails) error {
	connStr := fmt.Sprintf(
		"Host=%s user=%s Password=%s dbname=%s Port=%d",
		c.Host, c.Username, c.Password, c.DatabaseName, c.Port,
	)
	if c.Config == nil {
		c.Config = DefaultConnectConfig
	}
	d, err := gorm.Open(postgres.Open(connStr), c.Config)
	pg.Database = d
	if err != nil {
		return err
	}
	return nil
}
