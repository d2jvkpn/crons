package internal

import (
	"embed"
	"net/http"
	"time"

	"crons/crons"

	"github.com/d2jvkpn/go-web/pkg/wrap"
)

const (
	HTTP_MaxHeaderBytes     = 2 << 11 // 4K
	HTTP_ReadHeaderTimeout  = 2 * time.Second
	HTTP_ReadTimeout        = 10 * time.Second
	HTTP_WriteTimeout       = 10 * time.Second
	HTTP_IdleTimeout        = 60
	HTTP_MaxMultipartMemory = 8 << 20 // 8M

	LOG_SizeMB = 512
)

var (
	//go:embed static
	_Static embed.FS
	//go:embed templates
	_Templates embed.FS

	_Release bool
	_Server  *http.Server
	_Logger  *wrap.Logger
	Manager  *crons.Manager
)
