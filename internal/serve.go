package internal

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.uber.org/zap"
)

func Serve(addr string, parameters map[string]any) (err error) {
	_Logger.Info("startup", zap.Any("parameters", parameters), zap.String("address", addr))
	_Server.Addr = addr

	_Manager.Start()
	if err = _Server.ListenAndServe(); err != http.ErrServerClosed {
		Shutdown()
	} else {
		err = nil
	}

	return err
}

func Shutdown() {
	var err error

	_Logger.Warn("server down")

	if _Server != nil {
		ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
		if err = _Server.Shutdown(ctx); err != nil {
			_Logger.Error(fmt.Sprintf("server shutdown: %v", err))
		}
		cancel()
	}

	// close other goroutines or services
	_Manager.Shutdown()
	_Logger.Down()
	log.Println("<<< Exit")
}
