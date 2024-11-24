package base

import (
	"context"
	"errors"
	"github.com/tuannh982/simple-workflow-go/pkg/backend/psql/persistent/uow"
	"gorm.io/gorm"
)

const (
	IsolationLevelSerializable = "SERIALIZABLE"
)

type BaseRepository struct {
	DB *gorm.DB
}

func (b *BaseRepository) UnitOfWork(ctx context.Context) *uow.UnitOfWork {
	unitOfWork, err := uow.ExtractUnitOfWork(ctx)
	if err != nil {
		if errors.Is(err, uow.ErrUnitOfWorkNotExists) {
			return uow.NewUnitOfWork(b.DB)
		} else {
			panic(err)
		}
	}
	return unitOfWork
}
