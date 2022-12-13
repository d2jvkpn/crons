package internal

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"go.uber.org/zap"
)

func Serve(addr string, parameters map[string]any) (err error) {
	Logger.Info(
		"Server is starting",
		zap.Any("parameters", parameters),
		zap.String("address", addr),
		zap.Int("pid", os.Getpid()),
		zap.String("os", fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)),
	)
	_Server.Addr = addr

	if err = _Server.ListenAndServe(); err != http.ErrServerClosed {
		Shutdown()
	} else {
		err = nil
	}

	return err
}

func Shutdown() {
	var err error

	Logger.Warn("Server is shutting down")

	if _Server != nil {
		ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
		if err = _Server.Shutdown(ctx); err != nil {
			Logger.Error(fmt.Sprintf("server shutdown: %v", err))
		}
		cancel()
	}

	// close other goroutines or services
}
