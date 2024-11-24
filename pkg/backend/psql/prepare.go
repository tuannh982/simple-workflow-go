package psql

import (
	"fmt"
	"github.com/tuannh982/simple-workflow-go/pkg/backend/psql/persistent"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
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

func Connect(
	host string,
	port int,
	username string,
	password string,
	database string,
	config *gorm.Config,
) (*gorm.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d",
		host, username, password, database, port,
	)
	if config == nil {
		config = DefaultConnectConfig
	}
	return gorm.Open(postgres.Open(connStr), config)
}
