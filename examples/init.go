package examples

import (
	"github.com/tuannh982/simple-workflow-go/pkg/backend"
	"github.com/tuannh982/simple-workflow-go/pkg/backend/db"
	"github.com/tuannh982/simple-workflow-go/pkg/dataconverter"
	"go.uber.org/zap"
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

func InitPSQLBackend(psql db.PostgresDB, logger *zap.Logger) (backend.Backend, error) {
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

func InitSQLiteBackend(sqlite db.SQLiteDB, logger *zap.Logger) (backend.Backend, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	database, err := sqlite.Connect(db.ConnectionDetails{Database: "test.db", Config: nil})
	if err != nil {
		return nil, err
	}
	err = sqlite.Prepare(database)
	dataConverter := dataconverter.NewJsonDataConverter()
	be := db.NewSQLiteBackend(hostname, 5*time.Minute, dataConverter, database, logger)
	return be, nil
}
