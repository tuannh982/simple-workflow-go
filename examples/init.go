package examples

import (
	"os"
	"path"
	"time"

	"github.com/tuannh982/simple-workflow-go/pkg/backend"
	"github.com/tuannh982/simple-workflow-go/pkg/backend/db"
	"github.com/tuannh982/simple-workflow-go/pkg/codec"
	"go.uber.org/zap"
)

const (
	DbHost     = "localhost"
	DbPort     = 5432
	DbName     = "postgres"
	DbUser     = "user"
	DbPassword = "123456"
	SQLiteDB   = "test.db"
)

func InitPSQLBackend(psql *db.PostgresDB, logger *zap.Logger) (backend.Backend, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	err = psql.Connect(db.ConnectionDetails{Host: DbHost, Port: DbPort, Username: DbUser, Password: DbPassword, DatabaseName: DbName})
	if err != nil {
		return nil, err
	}
	err = psql.Prepare() // auto-create table if not exists
	if err != nil {
		return nil, err
	}
	dataConverter := codec.NewJSONCodec()
	be := backend.NewPSQLBackend(hostname, 5*time.Minute, dataConverter, *psql, logger)
	return be, nil
}

func InitSQLiteBackend(sqlite *db.SQLiteDB, logger *zap.Logger) (backend.Backend, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	err = sqlite.Connect(db.ConnectionDetails{DatabaseName: path.Join(wd, "..", "test.db"), Config: nil})
	if err != nil {
		return nil, err
	}
	err = sqlite.Prepare()
	if err != nil {
		return nil, err
	}
	dataConverter := codec.NewJSONCodec()
	be := backend.NewSQLiteBackend(hostname, 5*time.Minute, dataConverter, *sqlite, logger)
	return be, nil
}
