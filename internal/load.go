package internal

import (
	// "fmt"
	"html/template"
	"net/http"

	"github.com/d2jvkpn/go-web/pkg/resp"
	"github.com/d2jvkpn/go-web/pkg/wrap"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Load(release bool) (err error) {
	var (
		engi *gin.Engine
		rg   *gin.RouterGroup
		tmpl *template.Template
	)

	if !release {
		_Logger = wrap.NewLogger("logs/api.log", zap.DebugLevel, LOG_SizeMB, nil)
	} else {
		_Logger = wrap.NewLogger("logs/api.log", zap.InfoLevel, LOG_SizeMB, nil)
	}
	_Release = release

	if release {
		gin.SetMode(gin.ReleaseMode)
		engi = gin.New()
		engi.Use(gin.Recovery())
	} else {
		engi = gin.Default()
	}
	rg = &engi.RouterGroup

	// engi.LoadHTMLGlob("templates/*.tmpl")
	tmpl, err = template.ParseFS(_Templates, "templates/*.html", "templates/*/*.html")
	if err != nil {
		return err
	}
	engi.SetHTMLTemplate(tmpl)
	engi.Use(wrap.Cors("*"))

	apiLogger := resp.NewLogHandler[any](_Logger, "api")

	rg.GET("/healthz", wrap.Healthz)
	LoadAPI(rg, apiLogger)

	_Server = &http.Server{ // TODO: set consts in base.go
		ReadTimeout:       HTTP_ReadTimeout,
		WriteTimeout:      HTTP_WriteTimeout,
		ReadHeaderTimeout: HTTP_ReadHeaderTimeout,
		MaxHeaderBytes:    HTTP_MaxHeaderBytes,
		// Addr:              addr,
		Handler: engi,
	}

	return nil
}
