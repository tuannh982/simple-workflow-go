package db

import "gorm.io/gorm"

type DatabaseType string

const (
	PostgresDBType DatabaseType = "postgres"
	SQLiteDBType   DatabaseType = "sqlite"
)

type ConnectionDetails struct {
	Host         string
	Port         int
	Username     string
	Password     string
	DatabaseName string
	*gorm.Config
}

type Database interface {
	Prepare(db *gorm.DB) error
	Truncate(db *gorm.DB) error
	Connect(conn ConnectionDetails) (*gorm.DB, error)
	Type() DatabaseType
}
