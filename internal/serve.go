package internal

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

func Serve(addr string, meta map[string]any) (err error) {
	Logger.Info("Server is starting", zap.String("address", addr), zap.Any("meta", meta))
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
