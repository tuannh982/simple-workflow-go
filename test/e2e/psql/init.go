package psql

import (
	"github.com/tuannh982/simple-workflows-go/pkg/backend"
	"github.com/tuannh982/simple-workflows-go/pkg/backend/psql"
	"github.com/tuannh982/simple-workflows-go/pkg/dataconverter"
	"github.com/tuannh982/simple-workflows-go/pkg/registry"
	"github.com/tuannh982/simple-workflows-go/pkg/worker"
	"go.uber.org/zap"
	"os"
)

func InitBackend() (backend.Backend, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	err = psql.PrepareDB(db)
	if err != nil {
		return nil, err
	}
	err = psql.TruncateDB(db)
	if err != nil {
		return nil, err
	}
	dataConverter := dataconverter.NewJsonDataConverter()
	be := psql.NewPSQLBackend(hostname, dataConverter, db, logger)
	return be, nil
}

func InitWorkers(be backend.Backend, logger *zap.Logger) (*worker.ActivityWorker, *worker.WorkflowWorker, error) {
	ar := registry.NewActivityRegistry()
	wr := registry.NewWorkflowRegistry()
	err := ar.RegisterActivities(
		InterBankTransferActivity,
		CrossBankTransferActivity,
	)
	if err != nil {
		return nil, nil, err
	}
	err = wr.RegisterWorkflows(
		PaymentWorkflow,
	)
	if err != nil {
		return nil, nil, err
	}
	aw := worker.NewActivityWorker("activity_worker_0", be, ar, be.DataConverter(), logger)
	ww := worker.NewWorkflowWorker("workflow_worker_0", be, wr, be.DataConverter(), logger)
	return aw, ww, nil
}
