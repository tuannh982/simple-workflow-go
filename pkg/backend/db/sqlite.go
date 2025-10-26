package db

import (
	"fmt"
	"time"

	"github.com/tuannh982/simple-workflow-go/pkg/backend"
	"github.com/tuannh982/simple-workflow-go/pkg/dataconverter"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewSQLiteBackend(lockedBy string, lockExpirationDuration time.Duration, codec dataconverter.Codec, db *gorm.DB, config *zap.Logger) backend.Backend {
	return &backend.SimpleWorkflowGoBackend{}
}

type SQLiteDB struct{}

func (s *SQLiteDB) Prepare(db *gorm.DB) error {

	return nil
}

func (s *SQLiteDB) Truncate(db *gorm.DB) error {
	return nil
}

func (s *SQLiteDB) Connect(c ConnectionDetails) (*gorm.DB, error) {
	connStr := fmt.Sprintf(c.Database) // is of the form "test.db"
	if c.Config == nil {
		c.Config = nil
	}
	return gorm.Open(sqlite.Open(connStr), c.Config)
}
