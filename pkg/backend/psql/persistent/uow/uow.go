package uow

import (
	"context"
	"errors"
	"gorm.io/gorm"
)

const UnitOfWorkKey = "unitOfWork"

var ErrUnitOfWorkNotExists = errors.New("unit of work does not exists")
var ErrUnitOfWorkFoundButWrongType = errors.New("unit of work found but wrong type")

type UnitOfWork struct {
	Tx *gorm.DB
}

func NewUnitOfWork(tx *gorm.DB) *UnitOfWork {
	return &UnitOfWork{
		Tx: tx,
	}
}

func (u *UnitOfWork) InjectCtx(ctx context.Context) context.Context {
	return context.WithValue(ctx, UnitOfWorkKey, u)
}

func ExtractUnitOfWork(ctx context.Context) (*UnitOfWork, error) {
	if v := ctx.Value(UnitOfWorkKey); v != nil {
		if casted, ok := v.(*UnitOfWork); ok {
			return casted, nil
		} else {
			return nil, ErrUnitOfWorkFoundButWrongType
		}
	} else {
		return nil, ErrUnitOfWorkNotExists
	}
}
