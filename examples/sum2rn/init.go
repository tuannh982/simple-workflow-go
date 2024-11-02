package main

import (
	"github.com/tuannh982/simple-workflows-go/pkg/backend"
	"github.com/tuannh982/simple-workflows-go/pkg/backend/psql"
	"github.com/tuannh982/simple-workflows-go/pkg/dataconverter"
	"go.uber.org/zap"
	"os"
)

const (
	DbHost     = "localhost"
	DbPort     = 5432
	DbName     = "postgres"
	DbUser     = "user"
	DbPassword = "123456"
)

func InitBackend(logger *zap.Logger) (backend.Backend, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	db, err := psql.Connect(DbHost, DbPort, DbUser, DbPassword, DbName, nil)
	if err != nil {
		return nil, err
	}
	err = psql.PrepareDB(db) // auto-create table if not exists
	if err != nil {
		return nil, err
	}
	err = psql.TruncateDB(db) // truncate DB data
	if err != nil {
		return nil, err
	}
	dataConverter := dataconverter.NewJsonDataConverter()
	be := psql.NewPSQLBackend(hostname, dataConverter, db, logger)
	return be, nil
}
