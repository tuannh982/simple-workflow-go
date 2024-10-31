package psql

import (
	"github.com/tuannh982/simple-workflows-go/pkg/utils/worker"
	"strings"
)

func HandleSQLError(err error) error {
	if err != nil {
		if strings.Contains(err.Error(), "could not serialize access due to concurrent update") {
			return worker.SuppressedError
		}
	}
	return err
}
