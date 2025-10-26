package db

import "gorm.io/gorm"

type ConnectionDetails struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
	*gorm.Config
}

type Database interface {
	Prepare(db *gorm.DB) error
	Truncate(db *gorm.DB) error
	Connect(conn ConnectionDetails) (*gorm.DB, error)
}
