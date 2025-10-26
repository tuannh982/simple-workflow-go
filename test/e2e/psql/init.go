//go:build e2e
// +build e2e

package psql

import (
	"log"
	"os"
	"time"

	"github.com/tuannh982/simple-workflow-go/pkg/backend"
	"github.com/tuannh982/simple-workflow-go/pkg/backend/db"
	"github.com/tuannh982/simple-workflow-go/pkg/dataconverter"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	DbHost     = "localhost"
	DbPort     = 5432
	DbName     = "postgres"
	DbUser     = "user"
	DbPassword = "123456"
)

var (
	gormConfig = &gorm.Config{
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
)

func InitBackend(psql db.PostgresDB, logger *zap.Logger) (backend.Backend, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	database, err := psql.Connect(db.ConnectionDetails{Host: DbHost, Port: DbPort, Username: DbUser, Password: DbPassword, Database: DbName})
	if err != nil {
		return nil, err
	}
	err = psql.Prepare(database) // auto-create table if not exists
	if err != nil {
		return nil, err
	}
	dataConverter := dataconverter.NewJsonDataConverter()
	be := db.NewPSQLBackend(hostname, 5*time.Minute, dataConverter, database, logger)
	return be, nil
}
