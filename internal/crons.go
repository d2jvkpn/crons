package internal

import (
	// "fmt"

	"crons/crons"

	"github.com/d2jvkpn/go-web/pkg/wrap"
)

func LoadCron(fp, field string) (num int, err error) {
	var logger *wrap.Logger

	logger = wrap.NewLogger("logs/crons.log", wrap.LogLevelFromStr("info"), 256, nil)
	logger.Logger = logger.Logger.Named("manager")
	_Manager = crons.NewManager(logger)
	if num, err = _Manager.LoadTasksFronConfig(fp, field); err != nil {
		return 0, err
	}

	return num, nil
}
