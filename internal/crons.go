package internal

import (
	// "fmt"

	"crons/crons"
)

func LoadCron(fp, field string) (num int, err error) {
	Manager = crons.NewManager(Logger.Named("manager"))
	if num, err = Manager.LoadTasksFronConfig(fp, field); err != nil {
		return 0, err
	}

	return num, nil
}
