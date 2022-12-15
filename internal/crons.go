package internal

import (
	// "fmt"

	"crons/crons"

	"go.uber.org/zap"
)

func LoadCron(fp, field string, meta map[string]any) (num int, err error) {
	Manager = crons.NewManager(Logger.Named("manager"))
	Logger.Info("program", zap.Any("meta", meta))

	if num, err = Manager.LoadTasksFronConfig(fp, field); err != nil {
		return 0, err
	}

	return num, nil
}
