//go:build e2e
// +build e2e

package psql

import (
	"github.com/tuannh982/simple-workflows-go/pkg/backend"
	"github.com/tuannh982/simple-workflows-go/pkg/backend/psql"
	"github.com/tuannh982/simple-workflows-go/pkg/dataconverter"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
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

func InitBackend(logger *zap.Logger) (backend.Backend, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	db, err := psql.Connect(DbHost, DbPort, DbUser, DbPassword, DbName, gormConfig)
	if err != nil {
		return nil, err
	}
	err = psql.PrepareDB(db) // auto-create table if not exists
	if err != nil {
		return nil, err
	}
	dataConverter := dataconverter.NewJsonDataConverter()
	be := psql.NewPSQLBackend(hostname, 5*time.Minute, dataConverter, db, logger)
	return be, nil
}
